package harness

import "time"

// cursorProfile is the native profile of Cursor agent.
//
// Cursor agent takes the model as one value with the flag "--model". Some
// model names carry a bracket parameter block, for example
// "claude-opus-4-8[context=1m,effort=high]". The block stays inside the one
// argument, so the profile passes "--model" and the model unchanged.
//
// Two dialogs can block the first start in a new directory:
//   - The trust dialog for the current workspace. Hive reports "blocked" and
//     does not answer the dialog. The user accepts the dialog in the pane or
//     starts the agent with the native "--trust" flag. Both save the trust for
//     all time.
//   - The login dialog for an account with no credential. Hive reports
//     "blocked" and does not answer the dialog. The user logs in in the pane.
//
// Both timeouts are two times the measured start of the live check. The
// measured start is below the 120000 ms floor, so the profile gives the floor
// for each timeout.
type cursorProfile struct{}

func (cursorProfile) Kind() string { return "cursor" }

func (cursorProfile) Args(model string, extra []string) []string {
	args := []string{}
	if model != "" {
		args = append(args, "--model", model)
	}
	return append(args, extra...)
}

func (cursorProfile) StartTimeout() time.Duration { return 120 * time.Second }
func (cursorProfile) ReadyTimeout() time.Duration { return 120 * time.Second }
func (cursorProfile) ProcessNames() []string      { return []string{"cursor-agent"} }

// Trust gives the trust result of Cursor agent 2026.10.19-6bbf2c4. A new
// directory shows the workspace trust dialog. The --trust flag skips it, but
// writes ~/.cursor/projects/<slug>/.workspace-trusted, so the directory stays
// trusted for all time.
func (cursorProfile) Trust() Trust {
	return Trust{
		Version:  "2026.10.19-6bbf2c4",
		NoMethod: "--trust skips the workspace trust dialog but saves the trust for all time in ~/.cursor/projects",
	}
}
