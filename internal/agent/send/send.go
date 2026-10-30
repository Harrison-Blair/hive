// Package send sends one message to one agent by name.
package send

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/harness"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

const (
	// waitTimeout bounds the wait of --wait for the end of the turn.
	waitTimeout = 60 * time.Minute
	// startTimeout bounds the wait of --wait for the start of the turn.
	startTimeout = 15 * time.Second
)

// Options holds the values of one send.
type Options struct {
	Name string
	Text string
	Wait bool
	// CallerPane is the HERDR_PANE_ID of the caller.
	CallerPane string
	Logger     *slog.Logger
}

// Result is the target and its status after the send.
type Result struct {
	Name   string `json:"name"`
	Pane   string `json:"pane"`
	Sender string `json:"sender"`
	Status string `json:"status"`
}

// Run delivers "⬢ SENDER: TEXT" to the agent of the record Name through the
// native message path of its harness. It never types into the pane.
func Run(ctx context.Context, home string, options Options) output.Envelope {
	logger := agent.Logger(options.Logger)
	envelope := run(ctx, home, options, logger)
	attrs := []any{"command", "agent send", "target", options.Name, "outcome", envelope.Outcome}
	if envelope.Error != nil {
		attrs = append(attrs, "error", fmt.Sprint(envelope.Error))
	}
	logger.Debug("agent send", attrs...)
	return envelope
}

func run(ctx context.Context, home string, options Options, logger *slog.Logger) output.Envelope {
	if strings.TrimSpace(options.Text) == "" {
		return output.Failed("validate", "TEXT is empty")
	}
	record, err := agent.Load(home, options.Name)
	if errors.Is(err, agent.ErrNotFound) {
		return output.Failed("agent", "unknown agent: "+options.Name)
	}
	if err != nil {
		return output.Failed("agent", err.Error())
	}
	sender, err := agent.CallerName(home, options.CallerPane)
	if err != nil {
		return output.Failed("caller", err.Error())
	}
	if sender == "" {
		sender = "user"
	}
	pane, err := herdr.PaneGet(ctx, record.Pane, herdr.WithLogger(logger))
	if err != nil {
		if agent.IsCode(err, herdr.CodePaneNotFound) {
			return output.Failed("agent", fmt.Sprintf("agent %s is gone: herdr has no pane %s", record.Name, record.Pane))
		}
		return output.Failed("agent", fmt.Sprintf("agent %s: %s", record.Name, err))
	}
	profile, err := harness.Lookup(record.Harness)
	if err != nil {
		return output.Failed("agent", fmt.Sprintf("agent %s: %s", record.Name, err))
	}

	text := "⬢ " + sender + ": " + options.Text
	if err := profile.Deliver(ctx, record.Pane, text, herdr.WithLogger(logger)); err != nil {
		return output.Failed("deliver", fmt.Sprintf("agent %s in pane %s: %s", record.Name, record.Pane, err))
	}
	logger.Debug("agent messaged", "command", "agent send", "target", record.Name, "pane_id", record.Pane, hivelog.Body(text))
	result := Result{Name: record.Name, Pane: record.Pane, Sender: sender, Status: pane.AgentStatus}
	effects := []output.Effect{{Resource: "agent", Path: record.Pane, Action: "messaged"}}
	if !options.Wait {
		return output.Envelope{Outcome: output.OutcomeOK, Phase: "complete", Effects: effects, Error: nil, Result: result}
	}
	status, err := waitTurn(ctx, record.Pane, pane.AgentStatus, logger)
	if err != nil {
		return output.Envelope{Outcome: output.OutcomePartial, Phase: "wait", Effects: effects, Error: fmt.Sprintf("wait for agent %s in pane %s: %s", record.Name, record.Pane, err), Result: result}
	}
	result.Status = status
	if status == "blocked" {
		return output.Envelope{Outcome: output.OutcomePartial, Phase: "wait", Effects: effects, Error: fmt.Sprintf("agent %s is blocked by a dialog in pane %s: the message waits until the dialog closes; answer it in the pane", record.Name, record.Pane), Result: result}
	}
	return output.Envelope{Outcome: output.OutcomeOK, Phase: "complete", Effects: effects, Error: nil, Result: result}
}

// waitTurn waits for the turn of the message to end and returns the status.
// The harness starts the turn some time after the delivery, and agent.wait
// matches a status that the agent already has. So when the agent was idle or
// done, waitTurn first waits up to startTimeout for the status working.
func waitTurn(ctx context.Context, paneID, before string, logger *slog.Logger) (string, error) {
	if before != "working" && before != "blocked" {
		_, err := herdr.AgentWait(ctx, herdr.AgentWaitParams{Target: paneID, Until: []string{"working"}, TimeoutMS: uint64(startTimeout / time.Millisecond)}, herdr.WithLogger(logger), herdr.WithDeadline(startTimeout+10*time.Second))
		if err != nil && !agent.IsCode(err, "timeout") {
			return "", err
		}
	}
	info, err := herdr.AgentWait(ctx, herdr.AgentWaitParams{Target: paneID, Until: []string{"idle", "done", "blocked"}, TimeoutMS: uint64(waitTimeout / time.Millisecond)}, herdr.WithLogger(logger), herdr.WithDeadline(waitTimeout+10*time.Second))
	if err != nil {
		return "", err
	}
	return info.AgentStatus, nil
}

// Text renders the result of one send.
func Text(envelope output.Envelope) string {
	result, ok := envelope.Result.(Result)
	if !ok || envelope.Outcome != output.OutcomeOK {
		return ""
	}
	return fmt.Sprintf("sent to %s (%s): %s\n", result.Name, result.Pane, result.Status)
}
