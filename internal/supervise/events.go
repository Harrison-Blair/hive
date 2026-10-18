package supervise

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// eventData holds the fields the supervisor reads from each pushed event.
type eventData struct {
	PaneID              string       `json:"pane_id"`
	WorkspaceID         string       `json:"workspace_id"`
	TabID               string       `json:"tab_id"`
	AgentStatus         string       `json:"agent_status"`
	PreviousPaneID      string       `json:"previous_pane_id"`
	PreviousWorkspaceID string       `json:"previous_workspace_id"`
	PreviousTabID       string       `json:"previous_tab_id"`
	Released            bool         `json:"released"`
	FinalStatus         string       `json:"final_status"`
	Pane                *panePayload `json:"pane"`
}

type panePayload struct {
	PaneID      string `json:"pane_id"`
	WorkspaceID string `json:"workspace_id"`
	TabID       string `json:"tab_id"`
}

// handleEvent applies one row of the AC2 event table.
func (s *Supervisor) handleEvent(ctx context.Context, event herdr.Event) error {
	if err := s.snapshot(ctx); err != nil {
		return err
	}
	var data eventData
	if len(event.Data) > 0 {
		if err := json.Unmarshal(event.Data, &data); err != nil {
			s.logEvent(event, "invalid data: "+err.Error())
			return nil
		}
	}
	s.logEvent(event, data.PaneID)
	switch event.Event {
	case "pane.agent_status_changed":
		return s.onStatusChanged(ctx, event, data)
	case "pane_agent_detected":
		return s.onAgentDetected(ctx, event, data)
	case "pane_exited", "pane_closed":
		return s.onTopologyClosed(ctx, event, "pane_id", data.PaneID)
	case "tab_closed":
		return s.onTopologyClosed(ctx, event, "tab_id", data.TabID)
	case "workspace_closed":
		return s.onTopologyClosed(ctx, event, "workspace_id", data.WorkspaceID)
	case "pane_moved":
		return s.onPaneMoved(ctx, event, data)
	default:
		return nil
	}
}

func (s *Supervisor) onStatusChanged(ctx context.Context, event herdr.Event, data eventData) error {
	record, found := s.liveByPane(data.PaneID)
	if !found {
		s.logEvent(event, "no live record for pane")
		return nil
	}
	if data.AgentStatus == record.AgentStatus {
		s.logEvent(event, "status unchanged")
		return nil
	}
	switch data.AgentStatus {
	case "blocked":
		role := controllingRole(record)
		if role == "" {
			s.logEvent(event, "no controlling role")
			return nil
		}
		return s.notify(ctx, event, notification{
			role:        role,
			kind:        records.InboxBlocked,
			task:        record.Task,
			agent:       record.Name,
			text:        "blocked",
			agentKey:    record.Key,
			agentStatus: "blocked",
		})
	case "idle", "done":
		delete(s.quietDone, record.Key)
		if record.Finished {
			return s.stopAgent(ctx, event, record)
		}
		if !s.hasCompletedRecord(ctx, record) {
			s.quiet.start(record.Key, quietAfter(record))
		}
		return s.storeAgentStatus(ctx, record, data.AgentStatus)
	case "working":
		delete(s.quietDone, record.Key)
		s.quiet.cancel(record.Key)
		return s.storeAgentStatus(ctx, record, data.AgentStatus)
	default:
		s.logEvent(event, "ignored status "+data.AgentStatus)
		return nil
	}
}

func (s *Supervisor) onAgentDetected(ctx context.Context, event herdr.Event, data eventData) error {
	if !data.Released || data.FinalStatus != "unknown" {
		return nil
	}
	record, found := s.liveByPane(data.PaneID)
	if !found {
		s.logEvent(event, "no live record for pane")
		return nil
	}
	role := controllingRole(record)
	if role == "" {
		s.logEvent(event, "no controlling role")
		return nil
	}
	return s.notify(ctx, event, notification{
		role:      role,
		kind:      records.InboxExited,
		task:      record.Task,
		agent:     record.Name,
		text:      "exited",
		agentKey:  record.Key,
		lifecycle: records.AgentExited,
	})
}

func (s *Supervisor) onTopologyClosed(ctx context.Context, event herdr.Event, field, value string) error {
	if value == "" {
		return nil
	}
	record, found := s.liveByField(field, value)
	if !found {
		s.logEvent(event, "no live record for closed "+field)
		return nil
	}
	role := controllingRole(record)
	if role == "" {
		s.logEvent(event, "no controlling role")
		return nil
	}
	return s.notify(ctx, event, notification{
		role:      role,
		kind:      records.InboxUnresolved,
		task:      record.Task,
		agent:     record.Name,
		text:      "unresolved",
		agentKey:  record.Key,
		lifecycle: records.AgentUnresolved,
	})
}

func (s *Supervisor) onPaneMoved(ctx context.Context, event herdr.Event, data eventData) error {
	if data.PreviousPaneID == "" || data.Pane == nil {
		return nil
	}
	record, found := s.liveByPane(data.PreviousPaneID)
	if !found {
		s.logEvent(event, "no live record for the previous pane")
		return nil
	}
	if err := s.moveRecord(ctx, record, data); err != nil {
		return err
	}
	if err := s.snapshot(ctx); err != nil {
		return err
	}
	return s.resubscribe(ctx)
}

// notification is one inbox record group.
type notification struct {
	role  string
	kind  string
	task  string
	agent string
	text  string

	agentKey    string
	lifecycle   string
	agentStatus string
}

// notify writes one notification group under one held handle: the inbox
// counter, the inbox record, then the final status.
func (s *Supervisor) notify(ctx context.Context, event herdr.Event, n notification) error {
	envelope := s.applyNotification(ctx, n)
	s.logNotification(n, envelope)
	if envelope.Outcome != "ok" {
		return newStorageError(envelope)
	}
	s.writeLine(n.kind + " record for " + n.role)
	return s.snapshot(ctx)
}

func (s *Supervisor) applyNotification(ctx context.Context, n notification) output.Envelope {
	return s.appendAndStatus(ctx, n)
}

// handleQuiet writes one quiet record when the agent is still settled.
func (s *Supervisor) handleQuiet(ctx context.Context, key string) error {
	s.quiet.cancel(key)
	if s.quietDone[key] {
		return nil
	}
	if err := s.snapshot(ctx); err != nil {
		return err
	}
	record, ok := s.agents[key]
	if !ok || record.Status != records.AgentLive {
		return nil
	}
	if record.AgentStatus != "idle" && record.AgentStatus != "done" {
		return nil
	}
	role := controllingRole(record)
	if role == "" {
		return nil
	}
	s.quietDone[key] = true
	return s.notify(ctx, herdr.Event{Event: "quiet"}, notification{
		role:     role,
		kind:     records.InboxQuiet,
		task:     record.Task,
		agent:    record.Name,
		text:     "quiet",
		agentKey: record.Key,
	})
}

func (s *Supervisor) stopAgent(ctx context.Context, event herdr.Event, record records.Agent) error {
	s.writeLine("stop " + record.Key)
	result, envelope := s.ops.Stop(ctx, s.options.Home, record)
	if envelope.Outcome != "ok" {
		s.logger.Debug("supervisor stop failed", "command", "supervise", "agent_id", record.Key, "outcome", envelope.Outcome, "error", envelope.Error)
	} else {
		s.logger.Debug("supervisor stopped agent", "command", "supervise", "agent_id", result.Agent.Key)
	}
	return s.snapshot(ctx)
}

func (s *Supervisor) storeAgentStatus(ctx context.Context, record records.Agent, status string) error {
	envelope := s.statusOnly(ctx, record, status)
	if envelope.Outcome != "ok" {
		return newStorageError(envelope)
	}
	return s.snapshot(ctx)
}

// moveRecord updates the pane of one moved record and appends the old pane.
func (s *Supervisor) moveRecord(ctx context.Context, record records.Agent, data eventData) error {
	registry, err := store.Open(s.options.Home)
	if err != nil {
		return err
	}
	var updated records.Agent
	err = registry.WithLock(ctx, func(locked *store.LockedStore) error {
		var current records.Agent
		if err := locked.ReadIn([]string{"agents"}, record.Key, &current); err != nil {
			return err
		}
		if current.Status != records.AgentLive || current.Generation != record.Generation {
			return nil
		}
		current.PaneHistory = append(current.PaneHistory, current.PaneID)
		current.PaneID = data.Pane.PaneID
		if data.Pane.WorkspaceID != "" {
			current.WorkspaceID = data.Pane.WorkspaceID
		}
		if data.Pane.TabID != "" {
			current.TabID = data.Pane.TabID
		}
		updated = current
		return locked.WriteIn([]string{"agents"}, current.Key, current)
	})
	if err != nil {
		return err
	}
	_ = updated
	return nil
}

func isAgentNotFound(err error) bool {
	var serverError *herdr.Error
	return errors.As(err, &serverError) && serverError.Code == herdr.CodeAgentNotFound
}
