// Package list lists the agent records with their herdr status.
package list

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/harness"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// Options holds the values of one list.
type Options struct {
	Logger *slog.Logger
	// Now returns the current time. Nil means time.Now.
	Now func() time.Time
}

// Entry is one record and the status of its pane in herdr.
type Entry struct {
	agent.Record
	// Status is the herdr agent status, or gone when herdr has no such pane.
	Status string `json:"status"`
	// HarnessSessionID is the native session ID of the harness that herdr
	// reports, or empty when herdr reports none.
	HarnessSessionID string `json:"harness_session_id"`
}

// minAge is the age below which a record stays though its pane shows no
// harness process. Bare hive saves the coordinator record before it execs the
// harness, so for a moment the foreground process of its pane is hive.
const minAge = 60 * time.Second

// Run returns an Entry for each record, sorted by name. It deletes the record
// of a dead agent: its pane is gone, or the record is at least minAge old and
// its pane runs no process of its harness.
func Run(ctx context.Context, home string, options Options) output.Envelope {
	logger := agent.Logger(options.Logger)
	records, err := agent.List(home)
	if err != nil {
		logger.Debug("agent list", "command", "agent list", "outcome", output.OutcomeFailed, "error", err.Error())
		return output.Failed("records", err.Error())
	}
	now := options.Now
	if now == nil {
		now = time.Now
	}
	entries := make([]Entry, 0, len(records))
	effects := []output.Effect{}
	for _, record := range records {
		status, session, reason := "gone", "", ""
		pane, err := herdr.PaneGet(ctx, record.Pane, herdr.WithLogger(logger))
		switch {
		case err == nil:
			status, session = pane.AgentStatus, agent.SessionID(pane.AgentSession)
			reason = deadReason(ctx, record, now(), logger)
		case agent.IsCode(err, herdr.CodePaneNotFound):
			reason = reasonPaneGone
		default:
			message := fmt.Sprintf("agent %s: pane %s: %s", record.Name, record.Pane, err)
			logger.Debug("agent list", "command", "agent list", "outcome", output.OutcomeFailed, "error", message)
			failed := output.Failed("herdr", message)
			failed.Effects = effects
			return failed
		}
		if reason != "" && remove(home, record, reason, logger) {
			effects = append(effects, output.Effect{Resource: "record", Path: "agents/" + record.Name + ".json", Action: "removed"})
			continue
		}
		entries = append(entries, Entry{Record: record, Status: status, HarnessSessionID: session})
	}
	logger.Debug("agent list", "command", "agent list", "outcome", output.OutcomeOK, "count", len(entries), "removed", len(effects))
	return output.Envelope{Outcome: output.OutcomeOK, Phase: "complete", Effects: effects, Error: nil, Result: entries}
}

const (
	reasonPaneGone  = "pane gone"
	reasonNoHarness = "no harness process"
)

// deadReason returns why the agent of record, whose pane exists, is dead, or
// empty when it is alive or hive cannot tell. It never uses the herdr agent
// status, which can be unknown for a live harness.
func deadReason(ctx context.Context, record agent.Record, now time.Time, logger *slog.Logger) string {
	if now.Sub(record.StartedAt) < minAge {
		return ""
	}
	profile, err := harness.Lookup(record.Harness)
	if err != nil {
		return ""
	}
	running, err := harness.Running(ctx, profile, record.Pane, logger)
	switch {
	case agent.IsCode(err, herdr.CodePaneNotFound):
		return reasonPaneGone
	case err != nil:
		logger.Warn("agent record kept", "command", "agent list", "target", record.Name, "pane", record.Pane, "error", err.Error())
		return ""
	case running:
		return ""
	}
	return reasonNoHarness
}

// remove deletes record when its file still names the same pane, and logs
// the result. It reports whether it deleted the record.
func remove(home string, record agent.Record, reason string, logger *slog.Logger) bool {
	removed, err := agent.RemoveIfSame(home, record)
	switch {
	case err != nil:
		logger.Warn("agent record kept", "command", "agent list", "target", record.Name, "pane", record.Pane, "reason", reason, "error", err.Error())
	case removed:
		logger.Info("agent record removed", "command", "agent list", "target", record.Name, "pane", record.Pane, "reason", reason)
	default:
		logger.Info("agent record kept", "command", "agent list", "target", record.Name, "pane", record.Pane, "reason", "record replaced")
	}
	return removed
}

// Text renders the entries as a table.
func Text(envelope output.Envelope) string {
	entries, ok := envelope.Result.([]Entry)
	if !ok {
		return ""
	}
	if len(entries) == 0 {
		return "no agents\n"
	}
	var out strings.Builder
	table := tabwriter.NewWriter(&out, 0, 4, 2, ' ', 0)
	_, _ = fmt.Fprintln(table, "NAME\tSTATUS\tPANE\tPARENT\tHARNESS\tMODEL\tWORKTREE")
	for _, entry := range entries {
		_, _ = fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", entry.Name, entry.Status, entry.Pane, dash(entry.Parent), entry.Harness, entry.Model, dash(entry.Worktree))
	}
	_ = table.Flush()
	return out.String()
}

func dash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
