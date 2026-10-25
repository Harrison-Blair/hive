package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// CurrentSchemaVersion is the record schema version that this build writes.
const CurrentSchemaVersion = 3

const (
	schemaName       = "schema.json"
	agentFactsStep   = "agent facts"
	ownerClaimChange = "owner claim"
	ownerInboxChange = "owner inbox"
	taskMailChange   = "task mail"
	inboxSourceField = "migrated_from"
	// ownerInboxPlanName is the work list of the version 3 inbox change. The
	// step keeps it until it writes schema version 3.
	ownerInboxPlanName = "migrate-v3.json"
)

type schemaFile struct {
	Version int `json:"version"`
}

type storedAgent struct {
	id     string
	record records.Agent
}

type agentFactsChange struct {
	path   string
	id     string
	record records.Agent
}

// SchemaVersion returns the version of the state store of home. A missing
// schema file is version 1.
func SchemaVersion(home string) (int, error) {
	path := filepath.Join(home, ".hive", "state", schemaName)
	data, err := readRecordFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	var schema schemaFile
	if err := json.Unmarshal(data, &schema); err != nil || schema.Version < 1 {
		return 0, fmt.Errorf("schema invalid: %s", path)
	}
	return schema.Version, nil
}

// migrationContended records a home whose migration lock wait timed out in
// this process. A later store open of the same home tries the lock once and
// does not wait again, so one command does not add a two-second wait for each
// of its store opens. The first store open of the next command waits again.
var migrationContended sync.Map

// migrate brings the state store to CurrentSchemaVersion under the store
// lock. An open of a store at the current version takes no lock. Each step is
// idempotent, so a second process that read an old version before the first
// process finished does no further change.
func (s *Store) migrate() error {
	version, err := SchemaVersion(s.home)
	if err != nil {
		return err
	}
	if version > CurrentSchemaVersion {
		return fmt.Errorf("state schema version %d is newer than %d", version, CurrentSchemaVersion)
	}
	if version == CurrentSchemaVersion {
		if err := s.removeOwnerInboxPlan(); err != nil {
			return err
		}
		migrationContended.Delete(s.home)
		return nil
	}
	if _, contended := migrationContended.Load(s.home); contended {
		err = s.withLockOnce(func(locked *LockedStore) error {
			return s.migrateLocked(locked)
		})
	} else {
		err = s.withLock(context.Background(), func(locked *LockedStore) error {
			return s.migrateLocked(locked)
		})
	}
	if errors.Is(err, ErrLockTimeout) {
		migrationContended.Store(s.home, struct{}{})
		return err
	}
	migrationContended.Delete(s.home)
	return err
}

// withLockOnce runs fn with the store lock only if the lock is free now. It
// returns ErrLockTimeout without a wait when another process holds the lock.
// It is the one-try form of withLock for a migration that already waited.
func (s *Store) withLockOnce(fn func(*LockedStore) error) error {
	if err := s.checkHive(); err != nil {
		return err
	}
	if err := s.checkRoot(); err != nil {
		return err
	}
	if err := s.checkLock(); err != nil {
		return err
	}
	file, err := os.OpenFile(s.lockPath(), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return ErrLockTimeout
		}
		return err
	}
	locked := &LockedStore{store: s}
	defer func() {
		locked.released.Store(true)
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		_ = file.Close()
	}()
	return fn(locked)
}

func (s *Store) migrateLocked(locked *LockedStore) error {
	from, err := SchemaVersion(s.home)
	if err != nil {
		return err
	}
	to := CurrentSchemaVersion
	if from > to {
		return fmt.Errorf("state schema version %d is newer than %d", from, to)
	}
	if from >= to {
		return nil
	}

	// The migration is a store lifecycle event, not the action of the
	// command. It carries no request ID, so a request filter shows only the
	// command events.
	restoreRequestID := hivelog.SetRequestID("")
	defer restoreRequestID()
	writer, err := hivelog.NewWriter(s.home, "store", time.Now)
	if err != nil {
		return fmt.Errorf("state migration log: %w", err)
	}
	defer func() { _ = writer.Close() }()
	logger := slog.New(writer)
	for version := from; version < to; version++ {
		stepFrom, stepTo := version, version+1
		started := time.Now()
		logger.Debug("state migrate start", "from", stepFrom, "to", stepTo)
		count, failedPath, stepErr := s.migrationStep(locked, stepTo, logger)
		if stepErr != nil {
			logger.Debug("state migrate failed", "from", stepFrom, "to", stepTo, "path", failedPath, "error", stepErr.Error())
			return stepErr
		}
		if err := s.writeSchema(stepTo); err != nil {
			logger.Debug("state migrate failed", "from", stepFrom, "to", stepTo, "path", s.schemaRelative(), "error", err.Error())
			return err
		}
		if stepTo == 3 {
			// The version 3 plan stays until its schema version is written,
			// so a failed write continues from the same plan.
			if err := s.removeOwnerInboxPlan(); err != nil {
				logger.Debug("state migrate failed", "from", stepFrom, "to", stepTo, "path", s.stateRelative(ownerInboxPlanName), "error", err.Error())
				return err
			}
		}
		logger.Debug("state migrate done", "from", stepFrom, "to", stepTo, "count", count, "duration_ms", time.Since(started).Milliseconds())
	}
	return nil
}

// migrationStep runs one version step and returns the number of changed
// records and the path of a failure.
func (s *Store) migrationStep(locked *LockedStore, version int, logger *slog.Logger) (int, string, error) {
	switch version {
	case 2:
		return s.agentFactsStep(locked, logger)
	case 3:
		return s.ownerStep(locked, logger)
	default:
		return 0, s.schemaRelative(), fmt.Errorf("state migrate has no step for version %d", version)
	}
}

// agentFactsStep writes the version 2 facts of each agent record.
func (s *Store) agentFactsStep(locked *LockedStore, logger *slog.Logger) (int, string, error) {
	changes, err := s.agentFactsChanges(locked)
	if err != nil {
		return 0, s.stateRelative(), err
	}
	for _, change := range changes {
		if err := locked.WriteIn([]string{"agents"}, change.id, change.record); err != nil {
			return 0, change.path, fmt.Errorf("state migrate %s: %w", change.path, err)
		}
		logger.Debug("state migrate record", "path", change.path, "change", agentFactsStep)
	}
	return len(changes), "", nil
}

// ownerStep writes the version 3 change: the owner claim, the unread records
// of each old manager inbox, and the task records without controlling_role.
func (s *Store) ownerStep(locked *LockedStore, logger *slog.Logger) (int, string, error) {
	agents, err := s.storedAgentRecords(locked)
	if err != nil {
		return 0, s.stateRelative("agents"), err
	}
	count := 0
	claims, err := s.renameOwnerClaims(locked, logger)
	if err != nil {
		return count, s.stateRelative("claims"), err
	}
	count += claims
	inboxes, failedPath, err := s.moveOwnerInboxes(locked, agents, logger)
	if err != nil {
		return count, failedPath, err
	}
	count += inboxes
	cleaned, failedPath, err := s.cleanInboxMarkers(locked, logger)
	if err != nil {
		return count, failedPath, err
	}
	count += cleaned
	tasks, failedPath, err := s.removeTaskMail(locked, logger)
	if err != nil {
		return count, failedPath, err
	}
	count += tasks
	return count, "", nil
}

// renameOwnerClaims renames each manager-PROJECT claim to owner-PROJECT.
func (s *Store) renameOwnerClaims(locked *LockedStore, logger *slog.Logger) (int, error) {
	names, raw, err := namedRecords(locked, []string{"claims"})
	if err != nil {
		return 0, err
	}
	count := 0
	for index, name := range names {
		oldKey := strings.TrimSuffix(name, ".json")
		project := strings.TrimPrefix(oldKey, "manager-")
		if project == oldKey || !records.ValidID(project) {
			continue
		}
		var claim records.Claim
		if err := json.Unmarshal(raw[index], &claim); err != nil {
			return count, fmt.Errorf("record invalid: %s", s.stateRelative("claims", name))
		}
		newKey := "owner-" + project
		newPath := s.stateRelative("claims", newKey+".json")
		oldPath := s.stateRelative("claims", oldKey+".json")
		var current records.Claim
		currentErr := locked.ReadIn([]string{"claims"}, newKey, &current)
		same := currentErr == nil && current.Agent == claim.Agent && current.Generation == claim.Generation
		if !same {
			claim.Key = newKey
			if err := locked.WriteIn([]string{"claims"}, newKey, claim); err != nil {
				return count, fmt.Errorf("state migrate %s: %w", newPath, err)
			}
			logger.Debug("state migrate record", "path", newPath, "change", ownerClaimChange)
			count++
		}
		if err := locked.RemoveIn([]string{"claims"}, oldKey); err != nil {
			return count, fmt.Errorf("state migrate %s: %w", oldPath, err)
		}
		logger.Debug("state migrate record", "path", oldPath, "change", ownerClaimChange)
		count++
	}
	return count, nil
}

// managerInbox is one old manager inbox in the version 2 set. Its ids are the
// records to move, and its target is the inbox of the live owner of its
// project, or coordinator when the project has no live owner.
type managerInbox struct {
	key    string
	target string
	ids    []string
}

// moveOwnerInboxes appends each unread record of an old manager inbox to the
// inbox of the live owner, or to coordinator. It deletes each source record,
// then the old folder and counter. When the owner inbox has the old key, the
// records already have their place, so the step keeps the folder and counter.
//
// An old inbox can also be the owner inbox of another project. The step reads
// the source records of every old inbox before it copies one, and it deletes
// only a source record of the inbox. A record that another source appended to
// the folder stays.
func (s *Store) moveOwnerInboxes(locked *LockedStore, agents []records.Agent, logger *slog.Logger) (int, string, error) {
	sources, failedPath, err := s.managerInboxes(locked, agents)
	if err != nil {
		return 0, failedPath, err
	}
	count := 0
	for _, source := range sources {
		copied, failedPath, err := s.copyUnreadInbox(locked, source, logger)
		count += copied
		if err != nil {
			return count, failedPath, err
		}
	}
	for _, source := range sources {
		removed, err := s.removeMovedInbox(locked, source, logger)
		count += removed
		if err != nil {
			return count, s.stateRelative("inbox", source.key), err
		}
	}
	return count, "", nil
}

// managerInboxes returns each old manager inbox with its destination and the
// IDs of its source records. The step writes its plan before it copies a
// record and keeps the plan until the schema version is written, so a retry
// after a failure reads the same sources and destinations. Without a plan,
// the step reads the old inbox folders.
func (s *Store) managerInboxes(locked *LockedStore, agents []records.Agent) ([]managerInbox, string, error) {
	plan, found, err := s.ownerInboxPlan()
	if err != nil {
		return nil, s.stateRelative(ownerInboxPlanName), err
	}
	if !found {
		failedPath := ""
		plan, failedPath, err = s.discoverManagerInboxes(locked, agents)
		if err != nil {
			return nil, failedPath, err
		}
		if err := s.writeOwnerInboxPlan(plan); err != nil {
			return nil, s.stateRelative(ownerInboxPlanName), err
		}
	}
	sources := make([]managerInbox, 0, len(plan.Sources))
	for _, source := range plan.Sources {
		sources = append(sources, managerInbox{key: source.Key, target: source.Target, ids: source.IDs})
	}
	return sources, "", nil
}

// discoverManagerInboxes returns the work list of the version 3 inbox change
// from the old inbox folders of the store.
func (s *Store) discoverManagerInboxes(locked *LockedStore, agents []records.Agent) (ownerInboxPlan, string, error) {
	root := filepath.Join(s.stateDir(), "inbox")
	if err := s.checkList(root, []string{"inbox"}); err != nil {
		return ownerInboxPlan{}, s.stateRelative("inbox"), err
	}
	entries, err := os.ReadDir(root)
	if errors.Is(err, os.ErrNotExist) {
		return ownerInboxPlan{}, "", nil
	}
	if err != nil {
		return ownerInboxPlan{}, s.stateRelative("inbox"), err
	}
	plan := ownerInboxPlan{}
	for _, entry := range entries {
		source := entry.Name()
		project := strings.TrimPrefix(source, "manager-")
		if project == source || !records.ValidID(project) {
			continue
		}
		target := ownerInboxKey(agents, project)
		if target == source {
			continue
		}
		ids, err := s.sourceInboxIDs(locked, source)
		if err != nil {
			return ownerInboxPlan{}, s.stateRelative("inbox", source), err
		}
		plan.Sources = append(plan.Sources, ownerInboxPlanSource{Key: source, Target: target, IDs: ids})
	}
	return plan, "", nil
}

// ownerInboxPlan is the work list of the version 3 inbox change. A source is
// an old inbox folder. Its target is the inbox that receives its unread
// records, and its IDs are the records to move.
type ownerInboxPlan struct {
	Sources []ownerInboxPlanSource `json:"sources"`
}

type ownerInboxPlanSource struct {
	Key    string   `json:"key"`
	Target string   `json:"target"`
	IDs    []string `json:"ids"`
}

// ownerInboxPlan reads the work list of the version 3 inbox change. A missing
// file means that the step has not read the old inboxes yet.
func (s *Store) ownerInboxPlan() (ownerInboxPlan, bool, error) {
	path := filepath.Join(s.stateDir(), ownerInboxPlanName)
	data, err := readRecordFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return ownerInboxPlan{}, false, nil
	}
	if err != nil {
		return ownerInboxPlan{}, false, err
	}
	var plan ownerInboxPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return ownerInboxPlan{}, false, fmt.Errorf("record invalid: %s", s.stateRelative(ownerInboxPlanName))
	}
	for _, source := range plan.Sources {
		project := strings.TrimPrefix(source.Key, "manager-")
		valid := project != source.Key && records.ValidID(project) &&
			source.Target != source.Key && validateComponent("inbox", source.Target) == nil
		for _, id := range source.IDs {
			if validateComponent("id", id) != nil {
				valid = false
			}
		}
		if !valid {
			return ownerInboxPlan{}, false, fmt.Errorf("record invalid: %s", s.stateRelative(ownerInboxPlanName))
		}
	}
	return plan, true, nil
}

// writeOwnerInboxPlan writes the work list of the version 3 inbox change.
func (s *Store) writeOwnerInboxPlan(plan ownerInboxPlan) error {
	data, err := marshalRecord(plan)
	if err != nil {
		return err
	}
	return writeRecordFile(filepath.Join(s.stateDir(), ownerInboxPlanName), data)
}

// removeOwnerInboxPlan deletes the work list of the version 3 inbox change.
func (s *Store) removeOwnerInboxPlan() error {
	err := os.Remove(filepath.Join(s.stateDir(), ownerInboxPlanName))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// sourceInboxIDs returns the IDs of the records of one old inbox that came
// from version 2. It skips each record with a migration marker. It reads only
// the marker, so a record with another invalid field does not stop the step
// before it copies the records that come first.
func (s *Store) sourceInboxIDs(locked *LockedStore, source string) ([]string, error) {
	names, raw, err := namedRecords(locked, []string{"inbox", source})
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(names))
	for index, data := range raw {
		var marker struct {
			MigratedFrom string `json:"migrated_from"`
		}
		if err := json.Unmarshal(data, &marker); err != nil {
			return nil, fmt.Errorf("record invalid: %s", s.stateRelative("inbox", source, names[index]))
		}
		if marker.MigratedFrom != "" {
			continue
		}
		ids = append(ids, strings.TrimSuffix(names[index], ".json"))
	}
	return ids, nil
}

// copyUnreadInbox appends each unread source record of one old inbox to its
// target. A moved record keeps a marker of its source until the whole step
// succeeds, so a retry copies each source record once. A target record without
// a marker is a separate record, so it stays even when its content is equal.
func (s *Store) copyUnreadInbox(locked *LockedStore, source managerInbox, logger *slog.Logger) (int, string, error) {
	state, err := targetInboxState(locked, source.target)
	if err != nil {
		return 0, s.stateRelative("inbox", source.target), err
	}
	count := 0
	for _, id := range source.ids {
		var record records.Inbox
		err := locked.ReadIn([]string{"inbox", source.key}, id, &record)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return count, s.stateRelative("inbox", source.key, id+".json"), err
		}
		if record.ReadAt != "" {
			continue
		}
		sourceID := source.key + "/" + id
		if state.markers[sourceID] {
			continue
		}
		seq, err := locked.Next("inbox:" + source.target)
		if err != nil {
			return count, s.stateRelative("counters", "inbox:"+source.target+".json"), err
		}
		moved := movedInbox{Inbox: record, MigratedFrom: sourceID}
		moved.Seq = seq
		moved.Role = source.target
		moved.NudgedAt = ""
		moved.NudgedGeneration = 0
		next := strconv.FormatUint(seq, 10)
		if err := locked.WriteIn([]string{"inbox", source.target}, next, moved); err != nil {
			return count, s.stateRelative("inbox", source.target, next+".json"), err
		}
		logger.Debug("state migrate record", "path", s.stateRelative("inbox", source.target, next+".json"), "change", ownerInboxChange)
		count++
	}
	return count, "", nil
}

// removeMovedInbox deletes each moved source record from one old inbox. When
// the inbox has no record left, it deletes the folder and its counter. A
// record that another source appended to the folder stays, because the folder
// is also the target of that source.
func (s *Store) removeMovedInbox(locked *LockedStore, source managerInbox, logger *slog.Logger) (int, error) {
	count := 0
	for _, id := range source.ids {
		err := locked.RemoveIn([]string{"inbox", source.key}, id)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return count, err
		}
		logger.Debug("state migrate record", "path", s.stateRelative("inbox", source.key, id+".json"), "change", ownerInboxChange)
		count++
	}
	names, _, err := namedRecords(locked, []string{"inbox", source.key})
	if err != nil {
		return count, err
	}
	if len(names) > 0 {
		return count, nil
	}
	if err := s.removeInbox(source.key, logger); err != nil {
		return count, err
	}
	return count + 1, nil
}

// movedInbox is one copied inbox record. The marker names the source record
// of a run that did not finish, and it is removed when the step succeeds.
// It is not a field of records.Inbox, because no reader needs it.
type movedInbox struct {
	records.Inbox
	MigratedFrom string `json:"migrated_from,omitempty"`
}

// inboxCopyState holds the source IDs of the records that target already has
// from an earlier run of the step.
type inboxCopyState struct {
	markers map[string]bool
}

// targetInboxState reads the migration markers in one target inbox.
func targetInboxState(locked *LockedStore, key string) (inboxCopyState, error) {
	_, raw, err := namedRecords(locked, []string{"inbox", key})
	if err != nil {
		return inboxCopyState{}, err
	}
	state := inboxCopyState{markers: map[string]bool{}}
	for _, data := range raw {
		var record movedInbox
		if err := json.Unmarshal(data, &record); err != nil {
			return inboxCopyState{}, fmt.Errorf("record invalid: %s", ".hive/state/inbox/"+key)
		}
		if record.MigratedFrom != "" {
			state.markers[record.MigratedFrom] = true
		}
	}
	return state, nil
}

// cleanInboxMarkers removes the source marker from each copied inbox record.
// A run that stopped after it deleted each source inbox leaves markers with
// no source left, so a later run removes them too.
func (s *Store) cleanInboxMarkers(locked *LockedStore, logger *slog.Logger) (int, string, error) {
	root := filepath.Join(s.stateDir(), "inbox")
	entries, err := os.ReadDir(root)
	if errors.Is(err, os.ErrNotExist) {
		return 0, "", nil
	}
	if err != nil {
		return 0, s.stateRelative("inbox"), err
	}
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		key := entry.Name()
		names, raw, err := namedRecords(locked, []string{"inbox", key})
		if err != nil {
			return count, s.stateRelative("inbox", key), err
		}
		for index, data := range raw {
			var object map[string]json.RawMessage
			if err := json.Unmarshal(data, &object); err != nil {
				return count, s.stateRelative("inbox", key, names[index]), fmt.Errorf("record invalid: %s", s.stateRelative("inbox", key, names[index]))
			}
			if _, ok := object[inboxSourceField]; !ok {
				continue
			}
			delete(object, inboxSourceField)
			encoded, err := marshalRecord(object)
			if err != nil {
				return count, s.stateRelative("inbox", key, names[index]), err
			}
			id := strings.TrimSuffix(names[index], ".json")
			if err := locked.WriteIn([]string{"inbox", key}, id, json.RawMessage(encoded)); err != nil {
				return count, s.stateRelative("inbox", key, names[index]), err
			}
			logger.Debug("state migrate record", "path", s.stateRelative("inbox", key, names[index]), "change", ownerInboxChange)
			count++
		}
	}
	return count, "", nil
}

// removeInbox deletes one old inbox folder and its counter.
func (s *Store) removeInbox(key string, logger *slog.Logger) error {
	path := filepath.Join(s.stateDir(), "inbox", key)
	info, err := os.Lstat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("store path invalid: %s", path)
		}
		if err := os.RemoveAll(path); err != nil {
			return err
		}
		logger.Debug("state migrate record", "path", s.stateRelative("inbox", key), "change", ownerInboxChange)
	}
	counter := filepath.Join(s.stateDir(), "counters", "inbox:"+key+".json")
	if err := os.Remove(counter); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// removeTaskMail removes controlling_role from each task record.
func (s *Store) removeTaskMail(locked *LockedStore, logger *slog.Logger) (int, string, error) {
	names, raw, err := namedRecords(locked, []string{"tasks"})
	if err != nil {
		return 0, s.stateRelative("tasks"), err
	}
	count := 0
	for index, name := range names {
		var object map[string]json.RawMessage
		if err := json.Unmarshal(raw[index], &object); err != nil {
			return count, s.stateRelative("tasks", name), fmt.Errorf("record invalid: %s", s.stateRelative("tasks", name))
		}
		if _, ok := object["controlling_role"]; !ok {
			continue
		}
		delete(object, "controlling_role")
		data, err := marshalRecord(object)
		if err != nil {
			return count, s.stateRelative("tasks", name), err
		}
		path := s.stateRelative("tasks", name)
		if err := writeRecordFile(filepath.Join(s.stateDir(), "tasks", name), data); err != nil {
			return count, path, fmt.Errorf("state migrate %s: %w", path, err)
		}
		logger.Debug("state migrate record", "path", path, "change", taskMailChange)
		count++
	}
	return count, "", nil
}

// ownerInboxKey returns the inbox key of the live owner of one project, or
// coordinator when the project has no live owner.
func ownerInboxKey(agents []records.Agent, project string) string {
	owner := ""
	for _, record := range agents {
		if record.Status != records.AgentLive || record.Owns != project {
			continue
		}
		key := record.Name
		if record.Kind == records.AgentKindCoordinator {
			key = records.AgentKindCoordinator
		}
		if owner == "" || key < owner {
			owner = key
		}
	}
	if owner != "" {
		return owner
	}
	return records.AgentKindCoordinator
}

// namedRecords returns the file names and data of one nested collection.
func namedRecords(locked *LockedStore, collection []string) ([]string, []json.RawMessage, error) {
	names := []string{}
	raw, err := locked.ListIn(collection, func(filename string) bool {
		names = append(names, filename)
		return true
	})
	if err != nil {
		return nil, nil, err
	}
	return names, raw, nil
}

// agentFactsChanges returns the version 2 change of each agent record that
// does not have its facts yet.
func (s *Store) agentFactsChanges(locked *LockedStore) ([]agentFactsChange, error) {
	agents, err := s.storedAgents(locked)
	if err != nil {
		return nil, err
	}
	needsWorktree := false
	for _, entry := range agents {
		if entry.record.Role == "worker" || entry.record.Role == "verifier" {
			needsWorktree = true
			break
		}
	}
	var worktrees map[string]string
	if needsWorktree {
		worktrees, err = taskWorktrees(locked)
		if err != nil {
			return nil, err
		}
	}
	changes := []agentFactsChange{}
	for _, entry := range agents {
		updated := entry.record
		if !updated.ApplyRoleFacts(worktrees[entry.record.Task]) {
			continue
		}
		if !agentFactsDiffer(entry.record, updated) {
			continue
		}
		changes = append(changes, agentFactsChange{
			path:   s.stateRelative("agents", entry.id+".json"),
			id:     entry.id,
			record: updated,
		})
	}
	return changes, nil
}

func agentFactsDiffer(before, after records.Agent) bool {
	return before.Kind != after.Kind ||
		before.Profile != after.Profile ||
		before.Owns != after.Owns ||
		before.Worktree != after.Worktree
}

func taskWorktrees(locked *LockedStore) (map[string]string, error) {
	raw, err := locked.ListIn([]string{"tasks"}, nil)
	if err != nil {
		return nil, err
	}
	found := make(map[string]string, len(raw))
	for _, data := range raw {
		var task records.Task
		if err := json.Unmarshal(data, &task); err != nil {
			return nil, err
		}
		if task.ID == "" || task.Worktree == nil {
			continue
		}
		found[task.ID] = task.Worktree.Path
	}
	return found, nil
}

func (s *Store) storedAgents(locked *LockedStore) ([]storedAgent, error) {
	dir := filepath.Join(s.stateDir(), "agents")
	if err := locked.store.checkList(dir, []string{"agents"}); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	found := []storedAgent{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".json") {
			continue
		}
		path := filepath.Join(dir, name)
		data, err := readRecordFile(path)
		if err != nil {
			return nil, err
		}
		id := strings.TrimSuffix(name, ".json")
		if err := validateComponent("id", id); err != nil {
			return nil, err
		}
		var record records.Agent
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, fmt.Errorf("record invalid: %s", path)
		}
		found = append(found, storedAgent{id: id, record: record})
	}
	return found, nil
}

// storedAgentRecords returns each agent record of the store.
func (s *Store) storedAgentRecords(locked *LockedStore) ([]records.Agent, error) {
	agents, err := s.storedAgents(locked)
	if err != nil {
		return nil, err
	}
	found := make([]records.Agent, 0, len(agents))
	for _, entry := range agents {
		found = append(found, entry.record)
	}
	return found, nil
}

func (s *Store) writeSchema(version int) error {
	data, err := marshalRecord(schemaFile{Version: version})
	if err != nil {
		return err
	}
	return writeRecordFile(filepath.Join(s.stateDir(), schemaName), data)
}

func (s *Store) stateRelative(parts ...string) string {
	all := append([]string{".hive", "state"}, parts...)
	return filepath.ToSlash(filepath.Join(all...))
}

func (s *Store) schemaRelative() string {
	return s.stateRelative(schemaName)
}
