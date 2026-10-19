package agent

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// ResolveOptions holds the values of one agent resolve.
type ResolveOptions struct {
	Agent  string
	As     string
	Logger *slog.Logger
}

// ResolveResult is the settled record of one resolve.
type ResolveResult struct {
	Agent records.Agent `json:"agent"`
}

// Resolve settles an unresolved agent record by hand. The record must hold the
// status unresolved. The exited choice ends it. The live choice compares the
// identity of the named live agent with the record before it trusts the name.
func Resolve(ctx context.Context, home string, options ResolveOptions) output.Envelope {
	if ctx == nil {
		ctx = context.Background()
	}
	if options.As != records.AgentExited && options.As != records.AgentLive {
		return resolveFailure(home, options, "validate", "resolve requires --as exited or --as live")
	}
	shown := Show(home, options.Agent)
	if shown.Outcome != "ok" {
		return shown
	}
	record, ok := shown.Result.(records.Agent)
	if !ok {
		return resolveFailure(home, options, "agent", "agent record missing")
	}
	if record.Status != records.AgentUnresolved {
		return resolveFailure(home, options, "agent", "agent is "+record.Status)
	}
	if options.As == records.AgentExited {
		return resolveWrite(home, options, record, records.AgentExited, nil)
	}
	target := resolveTarget(record)
	live, err := herdr.AgentGet(ctx, target, herdr.WithLogger(options.Logger))
	if err != nil {
		return resolveHerdrFailure(home, options, record, err)
	}
	reason, checkErr := resolveIdentityCheck(ctx, record, live, options.Logger)
	if checkErr != nil {
		return resolveHerdrFailure(home, options, record, checkErr)
	}
	if reason != "" {
		return resolveFailure(home, options, "identity", reason)
	}
	return resolveWrite(home, options, record, records.AgentLive, &live)
}

// resolveTarget returns the herdr name of the record, or its pane ID when it
// has no name.
func resolveTarget(record records.Agent) string {
	if record.Name != "" {
		return record.Name
	}
	return record.PaneID
}

// resolveIdentityCheck compares the live agent with the record. An empty
// reason and a nil error mean a match. A non-nil error is a herdr failure.
func resolveIdentityCheck(ctx context.Context, record records.Agent, live herdr.AgentInfo, logger *slog.Logger) (string, error) {
	if live.Agent != record.Harness {
		return "agent identity differs: kind", nil
	}
	liveCwd := live.Cwd
	if liveCwd == "" {
		liveCwd = live.ForegroundCwd
	}
	if filepath.Clean(liveCwd) != filepath.Clean(record.Cwd) {
		return "agent identity differs: cwd", nil
	}
	if live.TerminalID != record.TerminalID {
		return "agent identity differs: terminal_id", nil
	}
	if record.ProcessPID == 0 {
		return "", nil
	}
	info, err := herdr.PaneProcessInfoGet(ctx, live.PaneID, herdr.WithLogger(logger))
	if err != nil {
		return "", err
	}
	found := false
	for _, entry := range info.ForegroundProcesses {
		if int64(entry.PID) == record.ProcessPID {
			found = true
			break
		}
	}
	if !found {
		return "agent identity differs: process_pid", nil
	}
	if record.ProcessStart != 0 {
		start, err := process.Start(record.ProcessPID)
		if err != nil || start != record.ProcessStart {
			return "agent identity differs: process_start", nil
		}
	}
	return "", nil
}

// resolveWrite settles the record under one held store handle.
func resolveWrite(home string, options ResolveOptions, record records.Agent, status string, live *herdr.AgentInfo) output.Envelope {
	updated, effect, err := mutateAgentRecord(home, record, records.AgentUnresolved, func(current *records.Agent) {
		current.Status = status
		current.LastError = ""
		if live != nil {
			current.PaneID = live.PaneID
			current.WorkspaceID = live.WorkspaceID
			current.TabID = live.TabID
			current.TerminalID = live.TerminalID
		}
	}, "resolved")
	if err != nil {
		envelope := resolveFailure(home, options, "store", "agent resolve: "+err.Error())
		envelope.Effects = appendEffect([]output.Effect{}, effect)
		return envelope
	}
	logAgentEvent(home, "agent resolve", "name", displayValue(options.Agent), "status", status)
	return output.Envelope{
		Outcome: "ok",
		Phase:   "complete",
		Effects: appendEffect([]output.Effect{}, effect),
		Error:   nil,
		Result:  ResolveResult{Agent: updated},
	}
}

// resolveHerdrFailure renders one herdr failure. The agent_not_found code gets
// its own message.
func resolveHerdrFailure(home string, options ResolveOptions, record records.Agent, err error) output.Envelope {
	message := err.Error()
	var serverError *herdr.Error
	if errors.As(err, &serverError) && serverError.Code == herdr.CodeAgentNotFound {
		message = "no live agent named " + resolveTarget(record)
	}
	return resolveFailure(home, options, "herdr", message)
}

func resolveFailure(home string, options ResolveOptions, phase, message string) output.Envelope {
	logAgentEvent(home, "agent resolve failed", "name", displayValue(options.Agent), "phase", phase, "error", message)
	return failed(phase, message)
}

// ResolveText renders the settled record of one resolve.
func ResolveText(envelope output.Envelope) string {
	text := effectsText(envelope)
	if envelope.Error != nil {
		return text
	}
	result, ok := envelope.Result.(ResolveResult)
	if !ok {
		return text
	}
	label := result.Agent.Name
	if label == "" {
		label = result.Agent.Key
	}
	return text + "agent " + label + " " + result.Agent.Status + "\n"
}
