package supervise

import (
	"context"
	"sort"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// refresh reconciles the store, subscribes to the watched panes, and evaluates
// the stop rule and the nudge rule.
func (s *Supervisor) refresh(ctx context.Context) error {
	if err := s.reconcile(ctx); err != nil {
		return err
	}
	if err := s.snapshot(ctx); err != nil {
		return err
	}
	if err := s.resubscribe(ctx); err != nil {
		return err
	}
	if err := s.snapshot(ctx); err != nil {
		return err
	}
	return s.evaluate(ctx)
}

// evaluate starts the quiet timers, then applies the stop rule and the nudge
// rule over the current records.
func (s *Supervisor) evaluate(ctx context.Context) error {
	if err := s.settleQuiet(ctx); err != nil {
		return err
	}
	if err := s.stopRule(ctx); err != nil {
		return err
	}
	return s.nudgeRule(ctx)
}

// settleQuiet starts the quiet timer of each settled live agent that has no
// completed record and no running timer.
func (s *Supervisor) settleQuiet(ctx context.Context) error {
	for _, key := range sortedKeys(s.agents) {
		record := s.agents[key]
		if record.Key == s.ownKey || record.Status != records.AgentLive {
			continue
		}
		if record.Finished {
			continue
		}
		if record.AgentStatus != "idle" && record.AgentStatus != "done" {
			continue
		}
		if s.quietDone[key] || s.quiet.running(key) {
			continue
		}
		completed, err := s.hasCompletedRecord(ctx, record)
		if err != nil {
			return err
		}
		if completed {
			continue
		}
		s.quiet.start(record.Key, quietAfter(record))
	}
	return nil
}

// stopRule stops each finished agent that is settled.
func (s *Supervisor) stopRule(ctx context.Context) error {
	for _, key := range sortedKeys(s.agents) {
		record := s.agents[key]
		if record.Key == s.ownKey || record.Status != records.AgentLive {
			continue
		}
		if !record.Finished {
			continue
		}
		if record.AgentStatus != "idle" && record.AgentStatus != "done" {
			continue
		}
		if err := s.stopAgent(ctx, herdr.Event{Event: "finished"}, record); err != nil {
			return err
		}
	}
	return nil
}

// reconcile marks each live record by the state of its pane and the own claim.
func (s *Supervisor) reconcile(ctx context.Context) error {
	if err := s.snapshot(ctx); err != nil {
		return err
	}
	for _, key := range sortedKeys(s.agents) {
		record := s.agents[key]
		if record.Key == s.ownKey || record.Status != records.AgentLive {
			continue
		}
		if record.Role == "supervisor" && record.Generation != s.generation {
			if err := s.markLifecycle(ctx, record, records.AgentStopped); err != nil {
				return err
			}
			continue
		}
		if err := s.reconcileRecord(ctx, record); err != nil {
			return err
		}
	}
	s.reconcileUnknownAgents(ctx)
	return s.snapshot(ctx)
}

func (s *Supervisor) reconcileRecord(ctx context.Context, record records.Agent) error {
	if record.PaneID == "" {
		return nil
	}
	info, err := s.ops.GetAgent(ctx, record.PaneID)
	if err == nil {
		if record.TerminalID != "" && info.TerminalID != record.TerminalID {
			return s.notifyLifecycle(ctx, record, records.InboxUnresolved, records.AgentUnresolved, "pane holds another terminal")
		}
		if record.TerminalID == "" && info.TerminalID != "" {
			if err := s.adoptAgent(ctx, record, info.TerminalID); err != nil {
				return err
			}
		}
		if info.AgentStatus != "" && info.AgentStatus != record.AgentStatus {
			// A status the supervisor learns from a read produces the same
			// record as a pushed event, so a queued push is not lost.
			return s.applyStatus(ctx, herdr.Event{Event: "reconcile"}, record, info.AgentStatus)
		}
		return nil
	}
	if !isAgentNotFound(err) {
		s.logger.Debug("supervisor reconcile skip", "command", "supervise", "agent_id", record.Key, "error", err.Error())
		return nil
	}
	pane, paneErr := s.ops.GetPane(ctx, record.PaneID)
	if paneErr != nil {
		return s.notifyLifecycle(ctx, record, records.InboxUnresolved, records.AgentUnresolved, "pane not found")
	}
	if pane.RestoreError != nil {
		return s.notifyLifecycle(ctx, record, records.InboxUnresolved, records.AgentUnresolved, "pane holds a restore error")
	}
	if pane.AgentStatus == "unknown" && pane.Agent == "" {
		return s.notifyLifecycle(ctx, record, records.InboxExited, records.AgentExited, "pane holds a shell")
	}
	return s.notifyLifecycle(ctx, record, records.InboxUnresolved, records.AgentUnresolved, "pane holds no live record")
}

// reconcileUnknownAgents logs each live herdr agent without a record.
func (s *Supervisor) reconcileUnknownAgents(ctx context.Context) {
	agents, err := s.ops.ListAgents(ctx)
	if err != nil {
		return
	}
	known := map[string]bool{}
	for _, record := range s.agents {
		if record.PaneID != "" {
			known[record.PaneID] = true
		}
	}
	for _, agent := range agents {
		if !known[agent.PaneID] {
			s.logger.Debug("supervisor unrecorded agent", "command", "supervise", "pane_id", agent.PaneID)
		}
	}
}

func (s *Supervisor) notifyLifecycle(ctx context.Context, record records.Agent, kind, lifecycle, detail string) error {
	role := controllingRole(record)
	if role == "" {
		s.logger.Debug("supervisor lifecycle record", "command", "supervise", "agent_id", record.Key, "status", lifecycle, "detail", detail)
		return nil
	}
	return s.notify(ctx, herdr.Event{Event: "reconcile"}, notification{
		role:      role,
		kind:      kind,
		task:      record.Task,
		agent:     record.Name,
		text:      lifecycle,
		agentKey:  record.Key,
		lifecycle: lifecycle,
	})
}

// adoptAgent stores the terminal ID of a matching agent reply.
func (s *Supervisor) adoptAgent(ctx context.Context, record records.Agent, terminalID string) error {
	registry, err := store.Open(s.options.Home)
	if err != nil {
		return err
	}
	envelope := output.Envelope{}
	now := s.clock.Now().UTC().Format(time.RFC3339Nano)
	err = registry.WithLock(ctx, func(locked *store.LockedStore) error {
		var current records.Agent
		if err := locked.ReadIn([]string{"agents"}, record.Key, &current); err != nil {
			envelope = failedEnvelope("status", "read agent: "+err.Error())
			return nil
		}
		if current.Status != records.AgentLive || current.Generation != record.Generation {
			envelope = failedEnvelope("status", "agent record changed")
			return nil
		}
		if terminalID != "" {
			current.TerminalID = terminalID
		}
		current.UpdatedAt = now
		if err := locked.WriteIn([]string{"agents"}, current.Key, current); err != nil {
			envelope = failedEnvelope("status", "write agent: "+err.Error())
			return nil
		}
		envelope = output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}}
		return nil
	})
	if err != nil {
		return err
	}
	return newStorageError(envelope)
}

// markLifecycle changes the lifecycle status of one record without an inbox
// record.
func (s *Supervisor) markLifecycle(ctx context.Context, record records.Agent, status string) error {
	registry, err := store.Open(s.options.Home)
	if err != nil {
		return err
	}
	envelope := output.Envelope{}
	now := s.clock.Now().UTC().Format(time.RFC3339Nano)
	err = registry.WithLock(ctx, func(locked *store.LockedStore) error {
		var current records.Agent
		if err := locked.ReadIn([]string{"agents"}, record.Key, &current); err != nil {
			envelope = failedEnvelope("status", "read agent: "+err.Error())
			return nil
		}
		if current.Status != records.AgentLive || current.Generation != record.Generation {
			envelope = failedEnvelope("status", "agent record changed")
			return nil
		}
		current.Status = status
		current.UpdatedAt = now
		if err := locked.WriteIn([]string{"agents"}, current.Key, current); err != nil {
			envelope = failedEnvelope("status", "write agent: "+err.Error())
			return nil
		}
		envelope = output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}}
		return nil
	})
	if err != nil {
		return err
	}
	return newStorageError(envelope)
}

func sortedKeys(agents map[string]records.Agent) []string {
	keys := make([]string, 0, len(agents))
	for key := range agents {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
