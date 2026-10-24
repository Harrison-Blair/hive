package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// CurrentSchemaVersion is the record schema version that this build writes.
const CurrentSchemaVersion = 2

const (
	schemaName      = "schema.json"
	migrationChange = "agent facts"
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

// migrate brings the state store to CurrentSchemaVersion under the store
// lock. An open of a store at the current version takes no lock. The step is
// idempotent, so a second process that read version 1 before the first
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
		return nil
	}
	return s.withLock(context.Background(), func(locked *LockedStore) error {
		return s.migrateLocked(locked)
	})
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
	changes, err := s.agentFactsChanges(locked)
	if err != nil {
		s.logMigrationFailure(from, to, s.stateRelative(), err)
		return err
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
	started := time.Now()
	logger.Debug("state migrate start", "from", from, "to", to)
	for _, change := range changes {
		if err := locked.WriteIn([]string{"agents"}, change.id, change.record); err != nil {
			logger.Debug("state migrate failed", "from", from, "to", to, "path", change.path, "error", err.Error())
			return fmt.Errorf("state migrate %s: %w", change.path, err)
		}
		logger.Debug("state migrate record", "path", change.path, "change", migrationChange)
	}
	if err := s.writeSchema(to); err != nil {
		logger.Debug("state migrate failed", "from", from, "to", to, "path", s.schemaRelative(), "error", err.Error())
		return err
	}
	logger.Debug("state migrate done", "from", from, "to", to, "count", len(changes), "duration_ms", time.Since(started).Milliseconds())
	return nil
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

func (s *Store) logMigrationFailure(from, to int, path string, failure error) {
	restoreRequestID := hivelog.SetRequestID("")
	defer restoreRequestID()
	writer, err := hivelog.NewWriter(s.home, "store", time.Now)
	if err != nil {
		return
	}
	defer func() { _ = writer.Close() }()
	slog.New(writer).Debug("state migrate failed", "from", from, "to", to, "path", path, "error", failure.Error())
}
