package harness

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/Harrison-Blair/hive/internal/guide"
)

// claudeProfile is the native profile of Claude Code.
type claudeProfile struct{}

func (claudeProfile) Kind() string { return "claude" }

func (claudeProfile) Args(model string, extra []string) []string {
	args := []string{"--dangerously-skip-permissions", "--settings", claudeInbound}
	if model != "" {
		args = append(args, "--model", model)
	}
	return append(args, extra...)
}

func (claudeProfile) StartTimeout() time.Duration { return 120 * time.Second }
func (claudeProfile) ReadyTimeout() time.Duration { return 120 * time.Second }
func (claudeProfile) ProcessNames() []string      { return []string{"claude"} }

// GuideArgs appends the guide to the system prompt of Claude Code. Each
// newline of the guide becomes the two characters \n.
func (claudeProfile) GuideArgs(text string) []string {
	if text == "" {
		return nil
	}
	return []string{"--append-system-prompt", guide.OneLine(text)}
}

// Trust gives the trust method of Claude Code 2.1.307. CLAUDE_CODE_SANDBOXED=1
// makes the workspace trust check pass for this process only. Claude Code
// writes no trust record to ~/.claude.json, so a later start without the
// value shows the dialog again.
func (claudeProfile) Trust() Trust {
	return Trust{Version: "2.1.307", Env: []string{"CLAUDE_CODE_SANDBOXED=1"}}
}

// Deliver writes text as one user message line to the inbox socket of the
// Claude Code process in pane. Claude Code shows it as a message from another
// session. It starts a turn when the session is idle, and it queues the
// message while a turn runs or a dialog is open.
func (p claudeProfile) Deliver(ctx context.Context, paneID, text string, logger *slog.Logger) error {
	process, err := harnessProcess(ctx, p, paneID, logger)
	if err != nil {
		return err
	}
	if !hasArgPair(process.Argv, "--settings", claudeInbound) {
		return fmt.Errorf("claude process %d was started without --settings %s: restart the agent", process.PID, claudeInbound)
	}
	path, err := claudeSocket(process.PID)
	if err != nil {
		return fmt.Errorf("%w: restart the agent", err)
	}
	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	line, err := json.Marshal(struct {
		Type    string  `json:"type"`
		Message message `json:"message"`
	}{"user", message{"user", text}})
	if err != nil {
		return err
	}
	if err := writeLine(ctx, path, line); err != nil {
		return fmt.Errorf("write to the message socket %s: %w", path, err)
	}
	return nil
}
