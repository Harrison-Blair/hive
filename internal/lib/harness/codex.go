package harness

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"github.com/Harrison-Blair/hive/internal/guide"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
)

const (
	// sessionTimeout bounds the wait for the session ID of a new Codex.
	sessionTimeout = 3 * time.Second
	sessionPause   = 250 * time.Millisecond
)

// queueTimeout bounds one codex queue process. A test can shorten it.
var queueTimeout = 30 * time.Second

// codexProfile is the native profile of Codex.
//
// Codex starts its tool commands through one shared app-server daemon by
// default. That daemon keeps the HERDR_* values of the pane that started it
// first, so a tool command in a later pane sees the wrong pane ID. The profile
// always passes --no-daemon, so the Codex process of the pane runs the tool
// commands with the values of that pane. See reference/herdr/environment.md.
//
// Three dialogs can block the first start in a new directory:
//   - The trust dialog for the directory contents. Hive reports "blocked" and
//     does not answer the dialog. The user accepts the dialog in the pane.
//   - The hook trust dialog for new or changed hooks. Hive reports "blocked"
//     and does not answer the dialog. The user reviews or trusts the hooks in
//     the pane.
//   - The login dialog for a provider with no credential. Hive reports
//     "blocked" and does not answer the dialog. The user logs in in the pane.
//
// Both timeouts are two times the measured start of the live check. The
// measured start is below the 120000 ms floor, so the profile gives the floor
// for each timeout.
type codexProfile struct{}

func (codexProfile) Kind() string { return "codex" }

func (codexProfile) Args(model string, extra []string) []string {
	args := []string{"--no-daemon", "--dangerously-bypass-approvals-and-sandbox"}
	if model != "" {
		args = append(args, "--model", model)
	}
	return append(args, extra...)
}

func (codexProfile) StartTimeout() time.Duration { return 120 * time.Second }
func (codexProfile) ReadyTimeout() time.Duration { return 120 * time.Second }
func (codexProfile) ProcessNames() []string      { return []string{"codex"} }

// GuideArgs gives the guide to Codex as developer instructions: a config
// override with the guide as one TOML basic string.
func (codexProfile) GuideArgs(text string) []string {
	if text == "" {
		return nil
	}
	return []string{"-c", "developer_instructions=" + guide.TOMLString(text)}
}

// Trust gives the trust result of Codex 0.160.1. A directory outside a Git
// repository shows no trust dialog. A Git repository with no saved trust shows
// the folder trust dialog, also with --dangerously-bypass-approvals-and-sandbox.
// Codex reads folder trust only from the user config: a "-c projects.<dir>"
// override does not skip the dialog. The other methods save trust in
// ~/.codex/config.toml for all time, or start Codex with a copy of CODEX_HOME.
func (codexProfile) Trust() Trust {
	return Trust{
		Version:  "0.160.1",
		NoMethod: "a Git repository with no saved trust shows the folder trust dialog; only a saved trust in ~/.codex/config.toml skips it",
	}
}

// Deliver queues text on the thread of the Codex session in pane with
// codex queue (codex-cli 0.166.0). Codex starts a turn with it when it is
// idle, and it never types into the pane. The thread UUID is the herdr
// agent_session of the pane. Codex makes the thread only at its first turn,
// so right after a spawn the session can be missing: Deliver waits up to
// sessionTimeout for it.
func (codexProfile) Deliver(ctx context.Context, paneID, text string, logger *slog.Logger) error {
	thread, err := codexThread(ctx, paneID, logger)
	if err != nil {
		return err
	}
	started := time.Now()
	queueCtx, cancel := context.WithTimeout(ctx, queueTimeout)
	defer cancel()
	cmd := exec.CommandContext(queueCtx, "codex", "queue", "--thread", thread, "--message", text)
	// A child of codex can keep the output pipe open after the kill.
	cmd.WaitDelay = time.Second
	out, err := cmd.CombinedOutput()
	attrs := []any{"binary", "codex", "args", []string{"queue", "--thread", thread, "--message", "<message>"}, "pane_id", paneID, "duration_ms", time.Since(started).Milliseconds()}
	if errors.Is(queueCtx.Err(), context.DeadlineExceeded) {
		err = fmt.Errorf("did not finish within %s: check codex, then send again", queueTimeout)
	}
	if err != nil {
		message := fmt.Sprintf("codex queue --thread %s: %s: %s", thread, err, strings.TrimSpace(string(out)))
		if logger != nil {
			logger.Error("external process", append(attrs, "error", message)...)
		}
		return errors.New(message)
	}
	if logger != nil {
		logger.Debug("external process", attrs...)
	}
	return nil
}

// codexThread returns the thread UUID of the Codex session in pane. It polls
// pane.get until herdr reports a herdr:codex session ID, at most for
// sessionTimeout.
func codexThread(ctx context.Context, paneID string, logger *slog.Logger) (string, error) {
	deadline := time.Now().Add(sessionTimeout)
	for {
		pane, err := herdr.PaneGet(ctx, paneID, herdr.WithLogger(logger))
		if err != nil {
			return "", fmt.Errorf("pane %s: %w", paneID, err)
		}
		if session := pane.AgentSession; session != nil && session.Source == "herdr:codex" && session.Kind == "id" && session.Value != "" {
			return session.Value, nil
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("codex in pane %s has no session ID after %s: codex makes its session at its first turn; send again after that turn, or restart the agent with a prompt", paneID, sessionTimeout)
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(sessionPause):
		}
	}
}

// InboxArgs returns nil: the native message path of Codex needs no file.
func (codexProfile) InboxArgs(string) ([]string, error) { return nil, nil }
