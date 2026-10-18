package harness

import "time"

// opencodeProfile is the native profile of OpenCode.
//
// OpenCode takes the model as one value of the form "provider/model". The slash
// stays inside the one argument, so the profile passes "--model" and the model.
//
// No dialog blocks the first start of the installed OpenCode. A new directory
// has no trust dialog. A provider with no credential does not show a login
// dialog: OpenCode starts with a default free model in place of the requested
// one. Hive reports "blocked" and does not answer a dialog when one appears.
// The user answers in the pane.
//
// Both timeouts are two times the measured start of the live check. The
// measured start is below the 120000 ms floor, so the profile gives the floor
// for each timeout.
type opencodeProfile struct{}

func (opencodeProfile) Kind() string { return "opencode" }

func (opencodeProfile) Args(model string, extra []string) []string {
	args := []string{}
	if model != "" {
		args = append(args, "--model", model)
	}
	return append(args, extra...)
}

func (opencodeProfile) StartTimeout() time.Duration { return 120 * time.Second }
func (opencodeProfile) ReadyTimeout() time.Duration { return 120 * time.Second }
func (opencodeProfile) ProcessNames() []string      { return []string{"opencode"} }
