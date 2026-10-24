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

// moveOwnerInboxes appends each unread record of an old manager inbox to the
// inbox of the live owner, or to coordinator. It deletes the old folder and
// counter. When the owner inbox has the old key, the records already have
// their place, so the step keeps the folder and the counter.
func (s *Store) moveOwnerInboxes(locked *LockedStore, agents []records.Agent, logger *slog.Logger) (int, string, error) {
	root := filepath.Join(s.stateDir(), "inbox")
	if err := s.checkList(root, []string{"inbox"}); err != nil {
		return 0, s.stateRelative("inbox"), err
	}
	entries, err := os.ReadDir(root)
	if errors.Is(err, os.ErrNotExist) {
		return 0, "", nil
	}
	if err != nil {
		return 0, s.stateRelative("inbox"), err
	}
	count := 0
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
		copied, failedPath, err := s.copyUnreadInbox(locked, source, target, logger)
		count += copied
		if err != nil {
			return count, failedPath, err
		}
		if err := s.removeInbox(source, logger); err != nil {
			return count, s.stateRelative("inbox", source), err
		}
		count++
	}
	return count, "", nil
}

// copyUnreadInbox appends each unread record of source to target. A record
// that target already has is not copied again.
func (s *Store) copyUnreadInbox(locked *LockedStore, source, target string, logger *slog.Logger) (int, string, error) {
	_, raw, err := namedRecords(locked, []string{"inbox", source})
	if err != nil {
		return 0, s.stateRelative("inbox", source), err
	}
	copied, err := copiedRecords(locked, target)
	if err != nil {
		return 0, s.stateRelative("inbox", target), err
	}
	count := 0
	for _, data := range raw {
		var record records.Inbox
		if err := json.Unmarshal(data, &record); err != nil {
			return count, s.stateRelative("inbox", source), fmt.Errorf("record invalid: %s", s.stateRelative("inbox", source))
		}
		if record.ReadAt != "" {
			continue
		}
		signature := inboxSignature(record)
		if copied[signature] {
			continue
		}
		seq, err := locked.Next("inbox:" + target)
		if err != nil {
			return count, s.stateRelative("counters", "inbox:"+target+".json"), err
		}
		moved := record
		moved.Seq = seq
		moved.Role = target
		moved.NudgedAt = ""
		moved.NudgedGeneration = 0
		id := strconv.FormatUint(seq, 10)
		if err := locked.WriteIn([]string{"inbox", target}, id, moved); err != nil {
			return count, s.stateRelative("inbox", target, id+".json"), err
		}
		logger.Debug("state migrate record", "path", s.stateRelative("inbox", target, id+".json"), "change", ownerInboxChange)
		copied[signature] = true
		count++
	}
	return count, "", nil
}

// copiedRecords returns the signature of each record in one inbox.
func copiedRecords(locked *LockedStore, key string) (map[string]bool, error) {
	_, raw, err := namedRecords(locked, []string{"inbox", key})
	if err != nil {
		return nil, err
	}
	found := make(map[string]bool, len(raw))
	for _, data := range raw {
		var record records.Inbox
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, fmt.Errorf("record invalid: %s", ".hive/state/inbox/"+key)
		}
		found[inboxSignature(record)] = true
	}
	return found, nil
}

// inboxSignature returns the fields that identify a copied inbox record.
func inboxSignature(record records.Inbox) string {
	return strings.Join([]string{
		record.Kind, record.Task, record.Agent, record.Text, record.Path, record.CreatedAt,
	}, "\x00")
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
