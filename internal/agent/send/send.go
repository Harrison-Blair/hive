// Package send prompts one agent for the agent send command.
package send

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"strings"
	"time"

	agentlib "github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/marker"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/panelock"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

const defaultPromptTimeout = 600

// Options holds explicit values for one prompt.
type Options struct {
	Agent      string
	Text       string
	Wait       bool
	Timeout    int
	TimeoutSet bool
}

// Result is the record and the reply of one prompt.
type Result struct {
	Agent  records.Agent `json:"agent"`
	Status string        `json:"status"`
}

// Run prompts one agent. It never sends the text a second time.
func Run(ctx context.Context, home string, options Options) output.Envelope {
	if ctx == nil {
		ctx = context.Background()
	}
	if options.Text == "" {
		return agentlib.Failed("validate", "prompt text is empty")
	}
	if agentlib.IsUserTagged(options.Text) {
		return agentlib.Failed("validate", "agent send cannot send a user-tagged message")
	}
	if len(options.Text) > agentlib.MaxPromptBytes {
		return agentlib.Failed("validate", fmt.Sprintf("prompt text is longer than %d bytes", agentlib.MaxPromptBytes))
	}
	timeout := defaultPromptTimeout
	if options.TimeoutSet {
		timeout = options.Timeout
	}
	if options.Wait && !promptTimeoutFits(timeout) {
		return agentlib.Failed("validate", fmt.Sprintf("invalid --timeout: %d", timeout))
	}
	writer, err := hivelog.NewWriter(home, "agent", time.Now)
	if err != nil {
		return agentlib.Failed("log", err.Error())
	}
	defer func() { _ = writer.Close() }()
	logger := slog.New(writer)

	shown := agentlib.Show(home, options.Agent)
	if shown.Outcome != "ok" {
		return shown
	}
	record, ok := shown.Result.(records.Agent)
	if !ok {
		return agentlib.Failed("agent", "agent record missing")
	}
	if record.Status != records.AgentLive {
		return agentlib.Failed("agent", "agent is "+record.Status)
	}
	attrs := []any{"command", "agent send", "agent_id", record.Key, "pane_id", record.PaneID}
	if sender := paneAgent(home, os.Getenv("HERDR_PANE_ID")); sender != "" {
		attrs = append(attrs, "from", sender)
	}
	logger.Debug("agent prompt", append(attrs, hivelog.Body(options.Text))...)

	text := options.Text
	if !strings.HasPrefix(text, marker.Prefix) {
		text = marker.Prefix + text
	}
	params := herdr.AgentPromptParams{Target: record.PaneID, Text: text}
	opts := []herdr.Option{herdr.WithLogger(logger)}
	if options.Wait {
		params.Wait = &herdr.AgentPromptWaitOptions{
			Until:     agentlib.SettledStatuses,
			TimeoutMS: uint64(timeout) * 1000,
		}
		opts = append(opts, herdr.WithDeadline(time.Duration(timeout)*time.Second+5*time.Second))
	}
	paneLock, lockErr := panelock.Lock(ctx, home, record.PaneID)
	if lockErr != nil {
		return sendEnvelope("failed", "pane lock: "+lockErr.Error(), record, logger)
	}
	defer func() { _ = paneLock.Unlock() }()
	reply, promptErr := herdr.AgentPrompt(ctx, params, opts...)
	if promptErr != nil {
		return sendFailure(ctx, home, logger, record, promptErr)
	}
	envelope := output.Envelope{
		Outcome: "ok",
		Phase:   "complete",
		Effects: []output.Effect{},
		Error:   nil,
		Result:  Result{Agent: record, Status: reply.AgentStatus},
	}
	logger.Debug("agent prompt end", "command", "agent send", "agent_id", record.Key, "outcome", envelope.Outcome, "status", reply.AgentStatus)
	return envelope
}

// paneAgent gives the name of the live agent record of pane (decision 72),
// or "" when no record has that pane.
func paneAgent(home, pane string) string {
	if pane == "" {
		return ""
	}
	listed, ok := agentlib.List(home, agentlib.ListOptions{}).Result.([]records.Agent)
	if !ok {
		return ""
	}
	for _, record := range listed {
		if record.PaneID == pane && record.Status == records.AgentLive {
			return record.Name
		}
	}
	return ""
}

func promptTimeoutFits(seconds int) bool {
	if seconds <= 0 {
		return false
	}
	if uint64(seconds) > math.MaxUint64/1000 {
		return false
	}
	const maxDeadlineSeconds = int64((math.MaxInt64 - 5*time.Second) / time.Second)
	return int64(seconds) <= maxDeadlineSeconds
}

func sendFailure(ctx context.Context, home string, logger *slog.Logger, record records.Agent, promptErr error) output.Envelope {
	var serverError *herdr.Error
	if errors.As(promptErr, &serverError) {
		switch serverError.Code {
		case herdr.CodeAgentBlocked:
			return sendEnvelope("failed", "agent_blocked: "+serverError.Message, record, logger)
		case herdr.CodeAgentPromptStalled:
			return sendEnvelope("failed", "agent prompt stalled: the text can be in the input of the agent: "+serverError.Error(), record, logger)
		case herdr.CodeAgentNotReady:
			return sendEnvelope("failed", serverError.Error(), record, logger)
		case herdr.CodeTimeout:
			return sendEnvelope("partial", "agent prompt timeout: the text reached the agent and no settled status came within the bound", record, logger)
		case herdr.CodeAgentNotFound:
			inspection := agentlib.InspectAgent(ctx, home, record, logger)
			envelope := output.Envelope{
				Outcome: "failed",
				Phase:   "prompt",
				Effects: []output.Effect{},
				Error:   inspection.MessageForSend(),
				Result:  Result{Agent: agentlib.ChooseRecord(inspection.Record, record)},
			}
			if inspection.Effect.Resource != "" {
				envelope.Effects = append(envelope.Effects, inspection.Effect)
			}
			logger.Debug("agent prompt end", "command", "agent send", "agent_id", record.Key, "outcome", envelope.Outcome, "error", fmt.Sprint(envelope.Error))
			return envelope
		default:
			return sendEnvelope("failed", serverError.Error(), record, logger)
		}
	}
	return sendEnvelope("partial", "agent prompt: "+promptErr.Error(), record, logger)
}

func sendEnvelope(outcome, message string, record records.Agent, logger *slog.Logger) output.Envelope {
	envelope := output.Envelope{
		Outcome: outcome,
		Phase:   "prompt",
		Effects: []output.Effect{},
		Error:   message,
		Result:  Result{Agent: record},
	}
	if logger != nil {
		logger.Debug("agent prompt end", "command", "agent send", "agent_id", record.Key, "outcome", outcome, "error", message)
	}
	return envelope
}

// Text renders the agent status of one prompt.
func Text(envelope output.Envelope) string {
	if envelope.Error != nil {
		return ""
	}
	result, ok := envelope.Result.(Result)
	if !ok {
		return ""
	}
	if result.Status == "" {
		return fmt.Sprintf("agent %s prompt sent\n", result.Agent.Name)
	}
	return fmt.Sprintf("agent %s %s\n", result.Agent.Name, result.Status)
}
