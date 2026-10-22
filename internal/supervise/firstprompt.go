package supervise

import (
	"context"
	"errors"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/panelock"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// keptPromptPending reports a kept first prompt with no first_prompt_sent_at.
// The supervisor sends no other prompt or nudge to that agent (decision 17).
func keptPromptPending(record records.Agent) bool {
	return record.FirstPrompt != "" && record.FirstPromptSentAt == ""
}

// firstPromptRule sends each kept first prompt to its idle agent one time
// (decisions 15 to 17). herdr gives done for an idle agent in a tab that
// nobody saw, so done counts as idle. A record with an attempt and no sent
// time becomes unresolved.
func (s *Supervisor) firstPromptRule(ctx context.Context) error {
	for _, key := range sortedKeys(s.agents) {
		record := s.agents[key]
		if record.Key == s.ownKey || record.Status != records.AgentLive || !keptPromptPending(record) {
			continue
		}
		if record.FirstPromptAttemptAt != "" {
			if err := s.notifyLifecycle(ctx, record, records.InboxUnresolved, records.AgentUnresolved, "first prompt delivery is not certain"); err != nil {
				return err
			}
			continue
		}
		if record.AgentStatus != "idle" && record.AgentStatus != "done" {
			continue
		}
		if err := s.sendFirstPrompt(ctx, record); err != nil {
			return err
		}
	}
	return nil
}

// sendFirstPrompt runs the four steps of decision 16 under the pane lock.
func (s *Supervisor) sendFirstPrompt(ctx context.Context, record records.Agent) error {
	lock, err := panelock.Lock(ctx, s.options.Home, record.PaneID)
	if err != nil {
		s.logger.Debug("supervisor first prompt lock failed", "command", "supervise", "agent_id", record.Key, "error", err.Error())
		return nil
	}
	defer func() { _ = lock.Unlock() }()

	attempted, err := s.markFirstPrompt(ctx, record, func(current *records.Agent, now string) bool {
		if current.FirstPromptAttemptAt != "" {
			return false
		}
		current.FirstPromptAttemptAt = now
		return true
	})
	if err != nil {
		s.logger.Debug("supervisor first prompt attempt write failed", "command", "supervise", "agent_id", record.Key, "error", err.Error())
		return nil
	}
	if !attempted {
		return nil
	}

	s.logger.Debug("supervisor first prompt", "command", "supervise", "agent_id", record.Key, "pane_id", record.PaneID)
	if _, err := s.ops.Prompt(ctx, herdr.AgentPromptParams{Target: record.PaneID, Text: record.FirstPrompt}); err != nil {
		s.logger.Debug("supervisor first prompt not certain", "command", "supervise", "agent_id", record.Key, "error", err.Error())
		return s.notifyLifecycle(ctx, record, records.InboxUnresolved, records.AgentUnresolved, "first prompt delivery is not certain: "+err.Error())
	}
	s.writeLine("first prompt " + record.Key)
	promptEffect := output.Effect{Resource: "agent", Path: statePath("agents", record.Key), Action: "prompted"}
	if _, err := s.markFirstPrompt(ctx, record, func(current *records.Agent, now string) bool {
		current.FirstPromptSentAt = now
		return true
	}); err != nil {
		return &storageError{envelope: output.Envelope{
			Outcome: "partial",
			Phase:   "first prompt",
			Effects: []output.Effect{promptEffect},
			Error:   "write first_prompt_sent_at: " + err.Error(),
		}}
	}
	return nil
}

// markFirstPrompt reads the record again under the store lock and writes the
// change of apply. It writes nothing when apply returns false.
func (s *Supervisor) markFirstPrompt(ctx context.Context, record records.Agent, apply func(*records.Agent, string) bool) (bool, error) {
	registry, err := store.Open(s.options.Home)
	if err != nil {
		return false, err
	}
	now := s.clock.Now().UTC().Format(time.RFC3339Nano)
	written := false
	err = registry.WithLock(ctx, func(locked *store.LockedStore) error {
		var current records.Agent
		if err := locked.ReadIn([]string{"agents"}, record.Key, &current); err != nil {
			return err
		}
		if current.Status != records.AgentLive || current.Generation != record.Generation || current.PaneID != record.PaneID {
			return errors.New("agent record changed")
		}
		if !apply(&current, now) {
			return nil
		}
		current.UpdatedAt = now
		if err := locked.WriteIn([]string{"agents"}, current.Key, current); err != nil {
			return err
		}
		written = true
		return nil
	})
	return written, err
}
