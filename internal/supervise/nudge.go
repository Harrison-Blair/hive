package supervise

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/marker"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/panelock"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// nudgeLine is the one prompt line of one nudge.
func nudgeLine(count int) string {
	noun := "record"
	if count != 1 {
		noun = "records"
	}
	return marker.Prefix + fmt.Sprintf("hive: %d new inbox %s, run: hive inbox", count, noun)
}

// nudgeRule sends one prompt to each settled role with a record that needs a
// nudge.
func (s *Supervisor) nudgeRule(ctx context.Context) error {
	roles, err := s.inboxRoles()
	if err != nil {
		return err
	}
	for _, role := range roles {
		record, ok := s.liveAgentForRole(role)
		if !ok {
			continue
		}
		if record.AgentStatus != "idle" && record.AgentStatus != "done" {
			continue
		}
		if keptPromptPending(record) {
			continue
		}
		todos, err := s.nudgeTodos(ctx, role, record.Generation)
		if err != nil {
			return err
		}
		if len(todos) == 0 {
			continue
		}
		if err := s.nudgeRole(ctx, role, record, todos); err != nil {
			return err
		}
	}
	return nil
}

func (s *Supervisor) nudgeRole(ctx context.Context, role string, record records.Agent, todos []records.Inbox) error {
	line := nudgeLine(len(todos))
	if err := s.promptNudge(ctx, record.PaneID, line); err != nil {
		s.logger.Debug("supervisor nudge failed", "command", "supervise", "role", role, "agent_id", record.Key, "error", err.Error())
		return nil
	}
	s.writeLine("nudged " + role)
	promptEffect := output.Effect{Resource: "agent", Path: statePath("agents", record.Key), Action: "nudged"}
	if err := s.markNudged(ctx, role, record, todos); err != nil {
		return &storageError{envelope: output.Envelope{
			Outcome: "partial",
			Phase:   "nudge",
			Effects: []output.Effect{promptEffect},
			Error:   err.Error(),
		}}
	}
	return s.snapshot(ctx)
}

func (s *Supervisor) promptNudge(ctx context.Context, paneID, text string) error {
	if strings.HasPrefix(strings.TrimPrefix(text, marker.Prefix), "user:") {
		return fmt.Errorf("supervisor nudge cannot use the user: tag")
	}
	lock, err := panelock.Lock(ctx, s.options.Home, paneID)
	if err != nil {
		return err
	}
	defer func() { _ = lock.Unlock() }()
	_, err = s.ops.Prompt(ctx, herdr.AgentPromptParams{Target: paneID, Text: text})
	return err
}

// inboxRoles returns the sorted role directories of the inbox.
func (s *Supervisor) inboxRoles() ([]string, error) {
	root := filepath.Join(s.options.Home, ".hive", "state", "inbox")
	entries, err := os.ReadDir(root)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	roles := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			roles = append(roles, entry.Name())
		}
	}
	sort.Strings(roles)
	return roles, nil
}

// liveAgentForRole returns the live agent of one inbox role.
func (s *Supervisor) liveAgentForRole(role string) (records.Agent, bool) {
	for _, key := range sortedKeys(s.agents) {
		record := s.agents[key]
		if record.Status == records.AgentLive && roleForAgent(record) == role {
			return record, true
		}
	}
	return records.Agent{}, false
}

// nudgeTodos returns the unread records that need a nudge for one generation.
func (s *Supervisor) nudgeTodos(ctx context.Context, role string, generation uint64) ([]records.Inbox, error) {
	var todos []records.Inbox
	registry, err := store.Open(s.options.Home)
	if err != nil {
		return nil, err
	}
	err = registry.WithLock(ctx, func(locked *store.LockedStore) error {
		all, err := inbox.AllHeld(locked, role)
		if err != nil {
			return err
		}
		for _, record := range all {
			if record.ReadAt != "" || record.Kind == records.InboxQuiet {
				continue
			}
			if record.NudgedAt == "" || record.NudgedGeneration != generation {
				todos = append(todos, record)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return todos, nil
}

// markNudged writes nudged_at and nudged_generation for each record after a
// successful prompt.
func (s *Supervisor) markNudged(ctx context.Context, role string, expected records.Agent, todos []records.Inbox) error {
	registry, err := store.Open(s.options.Home)
	if err != nil {
		return err
	}
	now := s.clock.Now().UTC().Format("2006-01-02T15:04:05.999999999Z07:00")
	var markErr error
	err = registry.WithLock(ctx, func(locked *store.LockedStore) error {
		var current records.Agent
		if err := locked.ReadIn([]string{"agents"}, expected.Key, &current); err != nil {
			markErr = err
			return nil
		}
		if current.Status != records.AgentLive || current.Generation != expected.Generation || roleForAgent(current) != role {
			markErr = errors.New("agent record changed")
			return nil
		}
		for _, record := range todos {
			id := strconv.FormatUint(record.Seq, 10)
			if err := locked.UpdateIn([]string{"inbox", role}, id, func(raw json.RawMessage) (json.RawMessage, error) {
				var item records.Inbox
				if err := json.Unmarshal(raw, &item); err != nil {
					return nil, err
				}
				item.NudgedAt = now
				item.NudgedGeneration = current.Generation
				return json.Marshal(item)
			}); err != nil {
				markErr = err
				return nil
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return markErr
}

// hasCompletedRecord reports whether the task of one record has a completed
// inbox record. A store read failure is returned, not treated as no record.
func (s *Supervisor) hasCompletedRecord(ctx context.Context, record records.Agent) (bool, error) {
	role := controllingRole(record)
	if role == "" || record.Task == "" {
		return false, nil
	}
	registry, err := store.Open(s.options.Home)
	if err != nil {
		return false, err
	}
	found := false
	err = registry.WithLock(ctx, func(locked *store.LockedStore) error {
		all, err := inbox.AllHeld(locked, role)
		if err != nil {
			return err
		}
		for _, item := range all {
			if item.Task == record.Task && (item.Kind == records.InboxCompleted || item.Kind == records.InboxVerified) {
				found = true
				return nil
			}
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return found, nil
}
