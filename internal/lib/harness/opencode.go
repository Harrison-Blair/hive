package harness

import "time"

// opencodeProfile is the native profile of OpenCode.
//
// OpenCode takes the model as one value of the form "provider/model". The slash
// stays inside the one argument, so the profile passes "--model" and the model.
//
// One dialog can block the first start in a new directory:
//   - The login dialog for a provider with no credential. Hive reports
//     "blocked" and does not answer the dialog. The user logs in in the pane.
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
