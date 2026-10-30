package harness

import (
	"context"
	"fmt"
	"time"

	"github.com/Harrison-Blair/hive/internal/guide"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
)

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

// GuideArgs appends the guide to the system prompt of Pi. Each newline of the
// guide becomes the two characters \n.
func (piProfile) GuideArgs(text string) []string {
	if text == "" {
		return nil
	}
	return []string{"--append-system-prompt", guide.OneLine(text)}
}

// Trust gives the trust method of Pi 1.3.0. Pi shows the project trust dialog
// only for project-local resources, for example .pi/settings.json. --approve
// trusts them for this run only and writes nothing to ~/.pi/agent/trust.json.
func (piProfile) Trust() Trust {
	return Trust{Version: "1.3.0", Args: []string{"--approve"}}
}

// Deliver fails: hive has no native message path to pi yet.
func (piProfile) Deliver(context.Context, string, string, ...herdr.Option) error {
	return fmt.Errorf("hive has no native message path to pi yet")
}
