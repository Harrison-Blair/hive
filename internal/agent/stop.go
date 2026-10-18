package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// StopResult is the post-close topology and the known effects of one stop.
type StopResult struct {
	Agent     records.Agent        `json:"agent"`
	Topology  string               `json:"topology"`
	Workspace *herdr.WorkspaceInfo `json:"workspace,omitempty"`
}

// StopByName resolves one agent name and stops its record.
func StopByName(ctx context.Context, home, name string) (StopResult, output.Envelope) {
	shown := Show(home, name)
	if shown.Outcome != "ok" {
		return StopResult{}, shown
	}
	record, ok := shown.Result.(records.Agent)
	if !ok {
		return StopResult{}, failed("agent", "agent record missing")
	}
	return Stop(ctx, home, record)
}

// Stop closes the pane of one spawned agent. T13 and T15 call it with a record.
func Stop(ctx context.Context, home string, record records.Agent) (StopResult, output.Envelope) {
	if ctx == nil {
		ctx = context.Background()
	}
	if record.Status != records.AgentLive {
		return StopResult{Agent: record}, stopEnvelope("failed", "agent", nil, "agent is "+record.Status, record)
	}
	if !record.Spawned {
		return StopResult{Agent: record}, stopEnvelope("failed", "validate", nil, "agent "+record.Name+" was not spawned by hive: exit it by hand", record)
	}
	writer, err := hivelog.NewWriter(home, "agent", time.Now)
	if err != nil {
		return StopResult{Agent: record}, stopEnvelope("failed", "log", nil, err.Error(), record)
	}
	defer func() { _ = writer.Close() }()
	logger := slog.New(writer)
	logger.Debug("agent stop start", "command", "agent stop", "agent_id", record.Key, "pane_id", record.PaneID)

	effects := []output.Effect{}
	pane, paneErr := herdr.PaneGet(ctx, record.PaneID, herdr.WithLogger(logger))
	if paneErr != nil {
		message := "pane get: " + paneErr.Error()
		var serverError *herdr.Error
		if errors.As(paneErr, &serverError) && serverError.Code == herdr.CodePaneNotFound {
			message = "pane " + record.PaneID + " not found: record unresolved"
		}
		updated, effect, updateErr := setRecordStatus(home, record, records.AgentLive, records.AgentUnresolved, message, "unresolved")
		effects = appendEffect(effects, effect)
		if updateErr != nil {
			message += "; " + updateErr.Error()
		}
		envelope := stopEnvelope("failed", "pane", effects, message, chooseRecord(updated, record))
		logStop(logger, envelope, record.Key)
		return StopResult{Agent: envelope.Result.(StopResult).Agent}, envelope
	}
	if record.TerminalID != "" && pane.TerminalID != record.TerminalID {
		message := "pane " + record.PaneID + " holds another terminal: record unresolved"
		updated, effect, updateErr := setRecordStatus(home, record, records.AgentLive, records.AgentUnresolved, message, "unresolved")
		effects = appendEffect(effects, effect)
		if updateErr != nil {
			message += "; " + updateErr.Error()
		}
		envelope := stopEnvelope("failed", "identity", effects, message, chooseRecord(updated, record))
		logStop(logger, envelope, record.Key)
		return StopResult{Agent: envelope.Result.(StopResult).Agent}, envelope
	}
	workspaceID := pane.WorkspaceID
	if workspaceID == "" {
		workspaceID = record.WorkspaceID
	}

	info, processErr := herdr.PaneProcessInfoGet(ctx, record.PaneID, herdr.WithLogger(logger))
	if processErr != nil {
		return stopUnresolved(home, logger, record, effects, "process info: "+processErr.Error())
	}
	if reason := checkProcessIdentity(record, info, pane.AgentSession); reason != "" {
		return stopUnresolved(home, logger, record, effects, "agent identity unresolved: "+reason)
	}
	if _, workspaceErr := herdr.WorkspaceGet(ctx, workspaceID, herdr.WithLogger(logger)); workspaceErr != nil {
		message := "workspace get: " + workspaceErr.Error()
		var serverError *herdr.Error
		if errors.As(workspaceErr, &serverError) && serverError.Code == herdr.CodeWorkspaceNotFound {
			message = "workspace " + workspaceID + " not found: record unresolved"
		}
		return stopUnresolved(home, logger, record, effects, message)
	}

	stopped, effect, updateErr := setRecordStatus(home, record, records.AgentLive, records.AgentStopped, "", "stopped")
	if updateErr != nil {
		envelope := stopEnvelope("failed", "agent", effects, updateErr.Error(), record)
		logStop(logger, envelope, record.Key)
		return StopResult{Agent: record}, envelope
	}
	effects = appendEffect(effects, effect)
	record = stopped
	logger.Debug("agent stopped", "command", "agent stop", "agent_id", record.Key, "pane_id", record.PaneID)

	closeErr := herdr.PaneClose(ctx, record.PaneID, herdr.WithLogger(logger))
	if closeErr != nil {
		var serverError *herdr.Error
		if errors.As(closeErr, &serverError) && serverError.Code == herdr.CodePaneNotFound {
			message := "pane " + record.PaneID + " not found: record unresolved"
			updated, unresolvedEffect, updateFail := setRecordStatus(home, record, records.AgentStopped, records.AgentUnresolved, message, "unresolved")
			effects = appendEffect(effects, unresolvedEffect)
			if updateFail != nil {
				message += "; " + updateFail.Error()
			}
			envelope := stopEnvelope("failed", "close", effects, message, chooseRecord(updated, record))
			logStop(logger, envelope, record.Key)
			return StopResult{Agent: envelope.Result.(StopResult).Agent}, envelope
		}
		message := "pane close: " + closeErr.Error()
		restored, liveEffect, restoreErr := setRecordStatus(home, record, records.AgentStopped, records.AgentLive, message, "live")
		effects = appendEffect(effects, liveEffect)
		if restoreErr != nil {
			message += "; " + restoreErr.Error()
		}
		envelope := stopEnvelope("failed", "close", effects, message, chooseRecord(restored, record))
		logStop(logger, envelope, record.Key)
		return StopResult{Agent: envelope.Result.(StopResult).Agent}, envelope
	}
	effects = append(effects, output.Effect{Resource: "pane", Path: record.PaneID, Action: "closed"})
	logger.Debug("pane closed", "command", "agent stop", "agent_id", record.Key, "pane_id", record.PaneID)

	result := StopResult{Agent: record, Topology: "unknown"}
	workspace, workspaceErr := herdr.WorkspaceGet(ctx, workspaceID, herdr.WithLogger(logger), herdr.WithDeadline(5*time.Second))
	switch {
	case workspaceErr == nil:
		result.Topology = "surviving"
		result.Workspace = &workspace
	default:
		var serverError *herdr.Error
		if errors.As(workspaceErr, &serverError) && serverError.Code == herdr.CodeWorkspaceNotFound {
			result.Topology = "closed"
		}
	}
	logger.Debug("agent stop topology", "command", "agent stop", "agent_id", record.Key, "topology", result.Topology)
	envelope := stopEnvelope("ok", "complete", effects, nil, result.Agent)
	envelope.Result = result
	logStop(logger, envelope, record.Key)
	return result, envelope
}

func checkProcessIdentity(record records.Agent, info herdr.PaneProcessInfo, session *herdr.AgentSessionInfo) string {
	if record.ProcessPID <= 0 {
		return "process_pid is empty"
	}
	found := false
	for _, entry := range info.ForegroundProcesses {
		if int64(entry.PID) != record.ProcessPID {
			continue
		}
		found = true
		if entry.Name != record.ProcessName {
			return "process_name differs"
		}
		break
	}
	if !found {
		return "process_pid is not in the foreground process set"
	}
	if record.ProcessName == "" {
		return "process_name is empty"
	}
	if record.ProcessStart == 0 {
		return "process_start is empty"
	}
	start, err := process.Start(record.ProcessPID)
	if err != nil {
		return err.Error()
	}
	if start != record.ProcessStart {
		return "process_start differs"
	}
	alive, err := process.Alive(record.ProcessPID, record.ProcessStart)
	if err != nil {
		return err.Error()
	}
	if !alive {
		return "process_start differs"
	}
	after, err := process.Start(record.ProcessPID)
	if err != nil {
		return err.Error()
	}
	if after != start {
		return "process_start is unstable"
	}
	if record.AgentSessionRef != "" {
		if session == nil {
			return "the recorded agent session is missing"
		}
		if sessionRef(session) != record.AgentSessionRef {
			return "the recorded agent session differs"
		}
	}
	return ""
}

func stopUnresolved(home string, logger *slog.Logger, record records.Agent, effects []output.Effect, message string) (StopResult, output.Envelope) {
	updated, effect, updateErr := setRecordStatus(home, record, records.AgentLive, records.AgentUnresolved, message, "unresolved")
	effects = appendEffect(effects, effect)
	if updateErr != nil {
		message += "; " + updateErr.Error()
	}
	envelope := stopEnvelope("failed", "identity", effects, message, chooseRecord(updated, record))
	logStop(logger, envelope, record.Key)
	return StopResult{Agent: envelope.Result.(StopResult).Agent}, envelope
}

func stopEnvelope(outcome, phase string, effects []output.Effect, err any, record records.Agent) output.Envelope {
	if effects == nil {
		effects = []output.Effect{}
	}
	return output.Envelope{Outcome: outcome, Phase: phase, Effects: effects, Error: err, Result: StopResult{Agent: record}}
}

func logStop(logger *slog.Logger, envelope output.Envelope, key string) {
	if logger == nil {
		return
	}
	attrs := []any{"command", "agent stop", "agent_id", key, "outcome", envelope.Outcome, "phase", envelope.Phase}
	if envelope.Error != nil {
		attrs = append(attrs, "error", fmt.Sprint(envelope.Error))
	}
	logger.Debug("agent stop end", attrs...)
}

// StopText renders the effects and the post-close topology of one stop.
func StopText(envelope output.Envelope) string {
	text := effectsText(envelope)
	if envelope.Error != nil {
		return text
	}
	result, ok := envelope.Result.(StopResult)
	if !ok {
		return text
	}
	return text + "workspace " + result.Topology + "\n"
}
