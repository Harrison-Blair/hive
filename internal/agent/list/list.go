// Package list lists the agent records with their herdr status.
package list

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"text/tabwriter"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// Options holds the values of one list.
type Options struct {
	Logger *slog.Logger
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

// Run returns an Entry for each record, sorted by name.
func Run(ctx context.Context, home string, options Options) output.Envelope {
	logger := agent.Logger(options.Logger)
	records, err := agent.List(home)
	if err != nil {
		logger.Debug("agent list", "command", "agent list", "outcome", "failed", "error", err.Error())
		return agent.Failed("records", err.Error())
	}
	entries := make([]Entry, 0, len(records))
	for _, record := range records {
		status, session := "gone", ""
		pane, err := herdr.PaneGet(ctx, record.Pane, herdr.WithLogger(logger))
		switch {
		case err == nil:
			status, session = pane.AgentStatus, agent.SessionID(pane.AgentSession)
		case !agent.IsCode(err, herdr.CodePaneNotFound):
			message := fmt.Sprintf("agent %s: pane %s: %s", record.Name, record.Pane, err)
			logger.Debug("agent list", "command", "agent list", "outcome", "failed", "error", message)
			return agent.Failed("herdr", message)
		}
		entries = append(entries, Entry{Record: record, Status: status, HarnessSessionID: session})
	}
	logger.Debug("agent list", "command", "agent list", "outcome", "ok", "count", len(entries))
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}, Error: nil, Result: entries}
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
