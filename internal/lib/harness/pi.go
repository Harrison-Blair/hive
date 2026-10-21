package harness

import "time"

// piProfile is the native profile of Pi.
//
// Pi takes the model as one pattern: "provider/id" with an optional
// ":<thinking>" suffix. The pattern carries the provider, so the profile never
// passes --provider. Pi normalizes the requested level at start. For example,
// native DeepSeek maps ":medium" to null, and the supported-level clamp then
// selects ":high". The profile keeps the requested argument unchanged.
//
// Two dialogs can block the first start in a new directory:
//   - The trust dialog for project-local files. Hive reports "blocked" and does
//     not answer the dialog. The user accepts the dialog in the pane.
//   - The login dialog for a provider with no credential. Hive reports
//     "blocked" and does not answer the dialog. The user logs in in the pane.
type piProfile struct{}

func (piProfile) Kind() string { return "pi" }

func (piProfile) Args(model string, extra []string) []string {
	args := []string{}
	if model != "" {
		args = append(args, "--model", model)
	}
	return append(args, extra...)
}

func (piProfile) StartTimeout() time.Duration { return 120 * time.Second }
func (piProfile) ReadyTimeout() time.Duration { return 120 * time.Second }
func (piProfile) ProcessNames() []string      { return []string{"pi"} }

// Trust gives the trust method of Pi 1.3.0. Pi shows the project trust dialog
// only for project-local resources, for example .pi/settings.json. --approve
// trusts them for this run only and writes nothing to ~/.pi/agent/trust.json.
func (piProfile) Trust() Trust {
	return Trust{Version: "1.3.0", Args: []string{"--approve"}}
}
