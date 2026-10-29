// Package stop stops one agent that has a hive record.
package stop

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// Options holds the values of one stop.
type Options struct {
	Name string
	// RemoveWorktree also removes the clean worktree checkout of the agent.
	RemoveWorktree bool
	Logger         *slog.Logger
}

// Run closes the pane of the agent, or its worktree workspace, and removes
// its record. It keeps the checkout and the branch unless RemoveWorktree
// removes a clean checkout. The branch stays always.
func Run(ctx context.Context, home string, options Options) output.Envelope {
	logger := agent.Logger(options.Logger)
	envelope, session := run(ctx, home, options, logger)
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
		return agent.Failed("agent", "unknown agent: "+options.Name), ""
	}
	if err != nil {
		return agent.Failed("agent", err.Error()), ""
	}
	if options.RemoveWorktree {
		if record.Worktree == "" {
			return agent.Failed("validate", "agent "+record.Name+" has no worktree"), ""
		}
		if message := checkClean(ctx, record.Worktree); message != "" {
			return agent.Failed("worktree", message), ""
		}
	}

	effects := []output.Effect{}
	pane, err := herdr.PaneGet(ctx, record.Pane, herdr.WithLogger(logger))
	session := agent.SessionID(pane.AgentSession)
	switch {
	case agent.IsCode(err, herdr.CodePaneNotFound):
		effects = append(effects, output.Effect{Resource: "pane", Path: record.Pane, Action: "gone"})
	case err != nil:
		return agent.Failed("pane", fmt.Sprintf("agent %s: %s", record.Name, err)), session
	case record.Worktree != "":
		if err := herdr.WorkspaceClose(ctx, pane.WorkspaceID, false, herdr.WithLogger(logger)); err != nil {
			return agent.Failed("close", fmt.Sprintf("close workspace %s: %s", pane.WorkspaceID, err)), session
		}
		effects = append(effects, output.Effect{Resource: "workspace", Path: pane.WorkspaceID, Action: "closed"})
	default:
		if err := herdr.PaneClose(ctx, record.Pane, herdr.WithLogger(logger)); err != nil {
			return agent.Failed("close", fmt.Sprintf("close pane %s: %s", record.Pane, err)), session
		}
		effects = append(effects, output.Effect{Resource: "pane", Path: record.Pane, Action: "closed"})
	}

	if options.RemoveWorktree {
		out, err := exec.CommandContext(ctx, "git", "-C", record.Worktree, "worktree", "remove", record.Worktree).CombinedOutput()
		if err != nil {
			message := fmt.Sprintf("git worktree remove %s: %s: %s", record.Worktree, err, strings.TrimSpace(string(out)))
			return output.Envelope{Outcome: "partial", Phase: "worktree", Effects: effects, Error: message, Result: record}, session
		}
		logger.Debug("worktree removed", "command", "agent stop", "target", record.Name, "path", record.Worktree)
		effects = append(effects, output.Effect{Resource: "worktree", Path: record.Worktree, Action: "removed"})
	}
	if err := agent.Remove(home, record.Name); err != nil {
		return output.Envelope{Outcome: "partial", Phase: "record", Effects: effects, Error: "remove record: " + err.Error(), Result: record}, session
	}
	effects = append(effects, output.Effect{Resource: "record", Path: ".hive/agents/" + record.Name + ".json", Action: "removed"})
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: effects, Error: nil, Result: record}, session
}

// checkClean returns a message when the checkout at path has changes or
// untracked files, or when Git cannot read it.
func checkClean(ctx context.Context, path string) string {
	out, err := exec.CommandContext(ctx, "git", "-C", path, "status", "--porcelain").CombinedOutput()
	if err != nil {
		return fmt.Sprintf("git status %s: %s: %s", path, err, strings.TrimSpace(string(out)))
	}
	if len(strings.TrimSpace(string(out))) > 0 {
		return fmt.Sprintf("worktree %s is dirty: commit or remove the changes first; nothing was stopped or removed", path)
	}
	return ""
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
