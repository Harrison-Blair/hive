// Package stop stops one agent that has a hive record.
package stop

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// Options holds the values of one stop.
type Options struct {
	Name   string
	Logger *slog.Logger
}

// Run closes the pane of the agent and removes its record. It never closes a
// workspace and never removes a worktree, so the other agents of the
// workspace keep running.
func Run(ctx context.Context, home string, options Options) output.Envelope {
	logger := agent.Logger(options.Logger)
	// The layout lock keeps a spawn of the name from starting between the
	// close of the pane and the removal of the record.
	var envelope output.Envelope
	var session string
	if err := agent.Hold(home, func() error {
		envelope, session = run(ctx, home, options, logger)
		return nil
	}); err != nil {
		envelope = output.Failed("lock", err.Error())
	}
	attrs := []any{"command", "agent stop", "target", options.Name, "outcome", envelope.Outcome}
	if session != "" {
		attrs = append(attrs, "harness_session_id", session)
	}
	if envelope.Error != nil {
		attrs = append(attrs, "error", fmt.Sprint(envelope.Error))
	}
	logger.Debug("agent stop", attrs...)
	return envelope
}

// run returns the envelope and the harness session ID that pane.get reports,
// or an empty ID.
func run(ctx context.Context, home string, options Options, logger *slog.Logger) (output.Envelope, string) {
	record, err := agent.Load(home, options.Name)
	if errors.Is(err, agent.ErrNotFound) {
		return output.Failed("agent", "unknown agent: "+options.Name), ""
	}
	if err != nil {
		return output.Failed("agent", err.Error()), ""
	}
	effects := []output.Effect{}
	pane, err := herdr.PaneGet(ctx, record.Pane, herdr.WithLogger(logger))
	session := agent.SessionID(pane.AgentSession)
	switch {
	case agent.IsCode(err, herdr.CodePaneNotFound):
		effects = append(effects, output.Effect{Resource: "pane", Path: record.Pane, Action: "gone"})
	case err != nil:
		return output.Failed("pane", fmt.Sprintf("agent %s: %s", record.Name, err)), session
	default:
		if err := herdr.PaneClose(ctx, record.Pane, herdr.WithLogger(logger)); err != nil {
			return output.Failed("close", fmt.Sprintf("close pane %s: %s", record.Pane, err)), session
		}
		effects = append(effects, output.Effect{Resource: "pane", Path: record.Pane, Action: "closed"})
	}

	if err := agent.Remove(home, record.Name); err != nil {
		return output.Envelope{Outcome: output.OutcomePartial, Phase: "record", Effects: effects, Error: "remove record: " + err.Error(), Result: record}, session
	}
	effects = append(effects, output.Effect{Resource: "record", Path: "agents/" + record.Name + ".json", Action: "removed"})
	return output.Envelope{Outcome: output.OutcomeOK, Phase: "complete", Effects: effects, Error: nil, Result: record}, session
}

// Text renders the effects of one stop.
func Text(envelope output.Envelope) string {
	var lines []string
	for _, effect := range envelope.Effects {
		lines = append(lines, fmt.Sprintf("%s %s %s", effect.Action, effect.Resource, effect.Path))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
