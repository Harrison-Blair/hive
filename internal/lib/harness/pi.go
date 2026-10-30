package harness

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/Harrison-Blair/hive/internal/guide"
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

// inboxExtension is the Pi extension of the native message path. It listens
// on $XDG_RUNTIME_DIR/hive-pi/<pid>.sock and passes each line to
// pi.sendUserMessage: as a new turn when Pi is idle, else as a follow-up.
//
//go:embed hive-inbox.ts
var inboxExtension []byte

// InboxArgs writes the hive-inbox extension to <home>/extensions when it is
// missing or different, and returns the -e argument that loads it.
func (piProfile) InboxArgs(home string) ([]string, error) {
	path := filepath.Join(home, "extensions", "hive-inbox.ts")
	if current, err := os.ReadFile(path); err != nil || !bytes.Equal(current, inboxExtension) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, inboxExtension, 0o644); err != nil {
			return nil, err
		}
	}
	return []string{"-e", path}, nil
}

// Deliver writes text as one JSON line {"text": TEXT} to the socket of the
// hive-inbox extension of the Pi process in pane.
func (p piProfile) Deliver(ctx context.Context, paneID, text string, logger *slog.Logger) error {
	process, err := harnessProcess(ctx, p, paneID, logger)
	if err != nil {
		return err
	}
	path, err := socketPath("hive-pi", process.PID, filepath.Join(os.TempDir(), "hive-pi"))
	if err != nil {
		return fmt.Errorf("%w: restart the agent, so pi loads the hive-inbox extension", err)
	}
	line, err := json.Marshal(struct {
		Text string `json:"text"`
	}{text})
	if err != nil {
		return err
	}
	if err := writeLine(ctx, path, line); err != nil {
		return fmt.Errorf("write to the message socket %s: %w", path, err)
	}
	return nil
}
