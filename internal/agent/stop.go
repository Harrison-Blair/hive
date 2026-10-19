package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/harness"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
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
	if record.Status == records.AgentStarting {
		settled, waitErr := waitForStartingRecord(ctx, home, record, startSettleTimeout(record))
		if waitErr != nil {
			message := waitErr.Error()
			if errors.Is(waitErr, errStartSettleTimeout) {
				updated, effect, updateErr := setRecordStatus(home, record, records.AgentStarting, records.AgentUnresolved, message, "unresolved")
				effects := appendEffect(nil, effect)
				if updateErr != nil {
					message += "; " + updateErr.Error()
				}
				updated = chooseRecord(updated, record)
				return StopResult{Agent: updated}, stopEnvelope("failed", "start", effects, message, updated)
			}
			return StopResult{Agent: record}, stopEnvelope("failed", "start", nil, message, record)
		}
		record = settled
	}
	if record.Status == records.AgentUnresolved {
		return StopResult{Agent: record}, stopEnvelope("failed", "agent", nil, "agent unresolved: run hive agent resolve", record)
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
	workspace, workspaceErr := herdr.WorkspaceGet(ctx, workspaceID, herdr.WithLogger(logger))
	if workspaceErr != nil {
		message := "workspace get: " + workspaceErr.Error()
		var serverError *herdr.Error
		if errors.As(workspaceErr, &serverError) && serverError.Code == herdr.CodeWorkspaceNotFound {
			message = "workspace " + workspaceID + " not found: record unresolved"
		}
		return stopUnresolved(home, logger, record, effects, message)
	}

	stopped, effect, updateErr := stopRecord(home, record)
	if updateErr != nil {
		if errors.Is(updateErr, errRecordChanged) || errors.Is(updateErr, errStopIdentityChanged) {
			message := "agent record changed: no close"
			envelope := stopEnvelope("failed", "identity", effects, message, chooseRecord(stopped, record))
			logStop(logger, envelope, record.Key)
			return StopResult{Agent: envelope.Result.(StopResult).Agent}, envelope
		}
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
		if isHerdrCode(closeErr, "confirmation_required") {
			message := "pane kept: " + closeErr.Error()
			effects = append(effects, output.Effect{Resource: "pane", Path: record.PaneID, Action: message})
			result := StopResult{Agent: record, Topology: "surviving", Workspace: &workspace}
			envelope := stopEnvelope("partial", "close", effects, message, record)
			envelope.Result = result
			logStop(logger, envelope, record.Key)
			return result, envelope
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
	workspace, workspaceErr = herdr.WorkspaceGet(ctx, workspaceID, herdr.WithLogger(logger), herdr.WithDeadline(5*time.Second))
	if workspaceErr == nil {
		result.Topology = "surviving"
		result.Workspace = &workspace
	} else {
		var serverError *herdr.Error
		if errors.As(workspaceErr, &serverError) && serverError.Code == herdr.CodeWorkspaceNotFound {
			result.Topology = "closed"
		}
	}
	logger.Debug("agent stop topology", "command", "agent stop", "agent_id", record.Key, "topology", result.Topology)
	if record.Role == "manager" && record.Project != "" {
		workspaceEffect, workspaceErr := recordStoppedWorkspace(home, record, result)
		effects = appendEffect(effects, workspaceEffect)
		if workspaceErr != nil {
			envelope := stopEnvelope("partial", "workspace", effects, workspaceErr.Error(), result.Agent)
			envelope.Result = result
			logStop(logger, envelope, record.Key)
			return result, envelope
		}
	}
	envelope := stopEnvelope("ok", "complete", effects, nil, result.Agent)
	envelope.Result = result
	logStop(logger, envelope, record.Key)
	return result, envelope
}

// recordStoppedWorkspace updates the project record from the post-close
// topology of one manager stop.
func recordStoppedWorkspace(home string, record records.Agent, result StopResult) (output.Effect, error) {
	switch result.Topology {
	case "closed":
		if err := clearProjectWorkspace(home, record.Project); err != nil {
			return output.Effect{}, err
		}
		return output.Effect{Resource: "workspace", Path: record.WorkspaceID, Action: "closed"}, nil
	case "surviving":
		other := 0
		if result.Workspace != nil {
			other = int(result.Workspace.PaneCount)
		}
		project, message := readProjectRecord(home, record.Project)
		if message != "" {
			return output.Effect{}, errors.New(message)
		}
		if project.Workspace != nil && project.Workspace.ID == record.WorkspaceID && project.Workspace.TabCreatedByHive {
			if err := clearProjectTab(home, record.Project, *project.Workspace); err != nil {
				return output.Effect{}, err
			}
		}
		return output.Effect{Resource: "workspace", Path: record.WorkspaceID, Action: fmt.Sprintf("kept: %d other panes", other)}, nil
	default:
		return output.Effect{}, nil
	}
}

var errStopIdentityChanged = errors.New("agent stop identity changed")

var errStartSettleTimeout = errors.New("agent start did not settle")

const startSettlePollInterval = 25 * time.Millisecond

func startSettleTimeout(record records.Agent) time.Duration {
	profile, err := harness.Lookup(record.Harness)
	if err == nil && profile.StartTimeout() > 0 {
		return profile.StartTimeout()
	}
	return defaultShellReadyTimeout
}

func waitForStartingRecord(ctx context.Context, home string, expected records.Agent, timeout time.Duration) (records.Agent, error) {
	registry, err := store.Open(home)
	if err != nil {
		return expected, err
	}
	deadline := time.Now().Add(timeout)
	if started, parseErr := time.Parse(time.RFC3339Nano, expected.StartedAt); parseErr == nil {
		startedDeadline := started.Add(timeout)
		if startedDeadline.Before(deadline) {
			deadline = startedDeadline
		}
	}
	for {
		var current records.Agent
		if err := registry.ReadIn([]string{"agents"}, expected.Key, &current); err != nil {
			return expected, err
		}
		if current.Key != expected.Key || current.Name != expected.Name || current.Role != expected.Role ||
			current.Project != expected.Project || current.Task != expected.Task ||
			current.PaneID != expected.PaneID || current.Generation != expected.Generation {
			return current, errRecordChanged
		}
		if current.Status != records.AgentStarting {
			return current, nil
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return current, fmt.Errorf("%w within %s", errStartSettleTimeout, timeout)
		}
		pause := startSettlePollInterval
		if remaining < pause {
			pause = remaining
		}
		timer := time.NewTimer(pause)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return current, ctx.Err()
		case <-timer.C:
		}
	}
}

// stopRecord makes the final stopped transition under one held store handle.
// It refuses when the checked ownership tuple changed since the external reads.
func stopRecord(home string, expected records.Agent) (records.Agent, output.Effect, error) {
	effect := output.Effect{Resource: "agent", Path: statePath("agents", expected.Key), Action: "stopped"}
	registry, err := store.Open(home)
	if err != nil {
		return expected, output.Effect{}, err
	}
	var updated records.Agent
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		var current records.Agent
		if err := locked.ReadIn([]string{"agents"}, expected.Key, &current); err != nil {
			return err
		}
		if err := checkCurrentRecord(locked, current, expected, records.AgentLive); err != nil {
			updated = current
			return err
		}
		if !sameStopIdentity(current, expected) {
			updated = current
			return errStopIdentityChanged
		}
		current.Status = records.AgentStopped
		current.LastError = ""
		current.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		if err := locked.WriteIn([]string{"agents"}, current.Key, current); err != nil {
			updated = current
			return err
		}
		updated = current
		return nil
	})
	if err != nil {
		return updated, output.Effect{}, err
	}
	return updated, effect, nil
}

func sameStopIdentity(current, expected records.Agent) bool {
	if current.Harness != expected.Harness {
		return false
	}
	if expected.TerminalID != "" && current.TerminalID != expected.TerminalID {
		return false
	}
	if expected.WorkspaceID != "" && current.WorkspaceID != expected.WorkspaceID {
		return false
	}
	if expected.TabID != "" && current.TabID != expected.TabID {
		return false
	}
	if current.ProcessPID != expected.ProcessPID || current.ProcessName != expected.ProcessName || current.ProcessStart != expected.ProcessStart {
		return false
	}
	return current.AgentSessionRef == expected.AgentSessionRef
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
