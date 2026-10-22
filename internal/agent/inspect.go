package agent

import (
	"context"
	"log/slog"

	"github.com/Harrison-Blair/hive/internal/lib/harness"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

type inspection struct {
	status string
	detail string
	record records.Agent
	effect output.Effect
}

func (i inspection) messageForStart() string {
	if i.status == records.AgentExited {
		return "agent exited during start"
	}
	return "agent start unresolved: " + i.detail
}

func (i inspection) messageForSend() string {
	if i.status == records.AgentExited {
		return "agent exited"
	}
	return "agent unresolved: " + i.detail
}

// inspectAgent decides whether an agent that herdr no longer names left the
// pane or still runs there. Only positive shell evidence in the recorded
// terminal permits the exited result.
func inspectAgent(ctx context.Context, home string, record records.Agent, logger *slog.Logger) inspection {
	pane, err := herdr.PaneGet(ctx, record.PaneID, herdr.WithLogger(logger))
	if err != nil {
		var serverError *herdr.Error
		if isServerError(err, &serverError) && serverError.Code == herdr.CodePaneNotFound {
			return markInspection(home, record, records.AgentUnresolved, "pane "+record.PaneID+" not found: record unresolved", logger)
		}
		return markInspection(home, record, records.AgentUnresolved, "pane get: "+err.Error(), logger)
	}
	if record.TerminalID != "" && pane.TerminalID != record.TerminalID {
		return markInspection(home, record, records.AgentUnresolved, "pane "+record.PaneID+" holds another terminal: record unresolved", logger)
	}
	if pane.AgentStatus != "" && pane.AgentStatus != "unknown" {
		return markInspection(home, record, records.AgentUnresolved, "pane "+record.PaneID+" holds agent status "+pane.AgentStatus, logger)
	}
	info, err := herdr.PaneProcessInfoGet(ctx, record.PaneID, herdr.WithLogger(logger))
	if err != nil {
		return markInspection(home, record, records.AgentUnresolved, "process info: "+err.Error(), logger)
	}
	if record.ProcessPID > 0 {
		for _, entry := range info.ForegroundProcesses {
			if int64(entry.PID) != record.ProcessPID || entry.Name != record.ProcessName {
				continue
			}
			alive, err := process.Alive(record.ProcessPID, record.ProcessStart)
			if err == nil && alive {
				return markInspection(home, record, records.AgentUnresolved, "the recorded agent process still runs in pane "+record.PaneID, logger)
			}
		}
	}
	profile, err := harness.Lookup(record.Harness)
	if err == nil {
		names := map[string]bool{}
		for _, name := range profile.ProcessNames() {
			names[name] = true
		}
		for _, entry := range info.ForegroundProcesses {
			if names[entry.Name] {
				return markInspection(home, record, records.AgentUnresolved, "a harness process still runs in pane "+record.PaneID, logger)
			}
		}
	}
	if shellProof(info) {
		return markInspection(home, record, records.AgentExited, "pane "+record.PaneID+" holds a proved shell", logger)
	}
	return markInspection(home, record, records.AgentUnresolved, "pane "+record.PaneID+" holds no proved shell", logger)
}

// shellProof accepts exactly one foreground entry that is the actual shell of
// the pane and is alive. A process name alone is not proof.
func shellProof(info herdr.PaneProcessInfo) bool {
	if info.ShellPID == nil || len(info.ForegroundProcesses) != 1 {
		return false
	}
	entry := info.ForegroundProcesses[0]
	if entry.PID != *info.ShellPID {
		return false
	}
	start, err := process.Start(int64(entry.PID))
	if err != nil {
		return false
	}
	alive, err := process.Alive(int64(entry.PID), start)
	return err == nil && alive
}

func markInspection(home string, record records.Agent, status, detail string, logger *slog.Logger) inspection {
	updated, effect, err := setRecordStatus(home, record, record.Status, status, detail, status)
	if err != nil {
		if logger != nil {
			logger.Debug("agent inspection update failed", "command", "agent inspect", "agent_id", record.Key, "status", status, "error", err.Error())
		}
		return inspection{status: records.AgentUnresolved, detail: detail + "; " + err.Error(), record: record}
	}
	if logger != nil {
		logger.Debug("agent inspection", "command", "agent inspect", "agent_id", record.Key, "status", status, "detail", detail)
	}
	return inspection{status: status, detail: detail, record: updated, effect: effect}
}

func isServerError(err error, target **herdr.Error) bool {
	if err == nil {
		return false
	}
	for {
		if serverError, ok := err.(*herdr.Error); ok {
			*target = serverError
			return true
		}
		type unwrapper interface{ Unwrap() error }
		wrapped, ok := err.(unwrapper)
		if !ok {
			return false
		}
		err = wrapped.Unwrap()
		if err == nil {
			return false
		}
	}
}
