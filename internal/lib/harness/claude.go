package harness

import "time"

// claudeProfile is the native profile of Claude Code.
type claudeProfile struct{}

func (claudeProfile) Kind() string { return "claude" }

func (claudeProfile) Args(model string, extra []string) []string {
	args := []string{}
	if model != "" {
		args = append(args, "--model", model)
	}
	return append(args, extra...)
}

func (claudeProfile) StartTimeout() time.Duration { return 120 * time.Second }
func (claudeProfile) ReadyTimeout() time.Duration { return 120 * time.Second }
func (claudeProfile) ProcessNames() []string      { return []string{"claude"} }
