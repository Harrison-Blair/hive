// Package send sends one prompt to one agent by name.
package send

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// waitTimeout bounds the wait of --wait for the end of the turn.
const waitTimeout = 60 * time.Minute

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

// Run sends "⬢ SENDER: TEXT" to the agent of the record Name.
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
	if _, err := herdr.PaneGet(ctx, record.Pane, herdr.WithLogger(logger)); err != nil {
		if agent.IsCode(err, herdr.CodePaneNotFound) {
			return output.Failed("agent", fmt.Sprintf("agent %s is gone: herdr has no pane %s", record.Name, record.Pane))
		}
		return output.Failed("agent", fmt.Sprintf("agent %s: %s", record.Name, err))
	}

	params := herdr.AgentPromptParams{Target: record.Pane, Text: "⬢ " + sender + ": " + options.Text}
	opts := []herdr.Option{herdr.WithLogger(logger)}
	if options.Wait {
		params.Wait = &herdr.AgentPromptWaitOptions{Until: []string{"idle", "done", "blocked"}, TimeoutMS: uint64(waitTimeout / time.Millisecond)}
		opts = append(opts, herdr.WithDeadline(waitTimeout+10*time.Second))
	}
	info, err := herdr.AgentPrompt(ctx, params, opts...)
	if agent.IsCode(err, herdr.CodeAgentBlocked) {
		return output.Failed("prompt", fmt.Sprintf("agent %s is blocked by a dialog in pane %s: answer it in the pane; the prompt was not sent", record.Name, record.Pane))
	}
	if err != nil {
		return output.Failed("prompt", fmt.Sprintf("prompt to agent %s in pane %s: %s", record.Name, record.Pane, err))
	}
	logger.Debug("agent prompted", "command", "agent send", "target", record.Name, "pane_id", record.Pane, hivelog.Body(params.Text))
	result := Result{Name: record.Name, Pane: record.Pane, Sender: sender, Status: info.AgentStatus}
	effects := []output.Effect{{Resource: "agent", Path: record.Pane, Action: "prompted"}}
	if options.Wait && info.AgentStatus == "blocked" {
		return output.Envelope{Outcome: output.OutcomePartial, Phase: "wait", Effects: effects, Error: fmt.Sprintf("agent %s is blocked by a dialog in pane %s: answer it in the pane", record.Name, record.Pane), Result: result}
	}
	return output.Envelope{Outcome: output.OutcomeOK, Phase: "complete", Effects: effects, Error: nil, Result: result}
}

// Text renders the result of one send.
func Text(envelope output.Envelope) string {
	result, ok := envelope.Result.(Result)
	if !ok || envelope.Outcome != output.OutcomeOK {
		return ""
	}
	return fmt.Sprintf("sent to %s (%s): %s\n", result.Name, result.Pane, result.Status)
}
