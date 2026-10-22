package harness

import "time"

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
