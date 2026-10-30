package harness

import (
	"time"

	"github.com/Harrison-Blair/hive/internal/guide"
)

// claudeProfile is the native profile of Claude Code.
type claudeProfile struct{}

func (claudeProfile) Kind() string { return "claude" }

func (claudeProfile) Args(model string, extra []string) []string {
	args := []string{"--dangerously-skip-permissions"}
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
