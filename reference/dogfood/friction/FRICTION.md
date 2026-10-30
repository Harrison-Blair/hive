# Friction log

This folder records each bug, missing capability, and workaround that an agent
finds when it uses hive, herdr, or a harness. The section "Friction log" in
`AGENTS.md` gives the procedure. `_template.md` gives the fields of an entry.

Each entry has one line in this file, under `## Open` or `## Resolved`:
`- [TITLE](entries/SLUG.md) — DESCRIPTION`. The title and the description are
the same as in the frontmatter of the entry. `friction_test.go` fails when this
file and the entry files do not agree.

The entries from before the teardown are at the tag `pre-teardown`, below
`reference/dogfood/friction/`.

## Open

- [An agent can send a message with user authority](entries/send-gives-user-authority.md) — hive agent send labels the sender user when the caller pane has no record or HERDR_PANE_ID is empty
- [Hive has no command to stop all agents](entries/no-stop-all-command.md) — hive agent stop takes one NAME, so the end of a session means one stop for each agent
- [A restart can silently run a stale hive binary](entries/restart-runs-stale-binary.md) — Bare hive and agents run the hive on PATH, and nothing compares it with the home checkout
- [A Codex spawn in an untrusted Git directory stops at the trust dialog](entries/codex-spawn-stops-at-trust-dialog.md) — Codex shows its folder trust dialog in a Git repository with no saved trust, and hive has no per-process trust method
- [Claude Code in bypass mode still asks before a dangerous rm](entries/bypass-still-prompts-dangerous-rm.md) — A Claude agent that runs rm on a variable path stops at a dialog despite --dangerously-skip-permissions
- [A harness can export NO_COLOR=1 and fail the colour tests of a project](entries/no-color-fails-colour-tests.md) — Hive neither sets nor clears NO_COLOR, so the test run of a spawned agent can fail on colour assertions
- [TestWithLockWaiterIsNotStarved fails at times in the full gate run](entries/lock-waiter-test-flakes.md) — The 100 ms waiter deadline can expire while the whole test suite runs in parallel
- [herdr agent.start can cut a long command line at 4095 bytes](entries/agent-start-line-cut-at-4095-bytes.md) — A start command line above 4095 bytes can reach the new pane shell cut short, so the harness never starts
- [A spawn worktree name cannot contain a slash](entries/worktree-name-no-slash.md) — hive agent spawn --worktree rejects a name with /, so a branch such as friction/SLUG needs a push with HEAD:friction/SLUG

## Resolved

- [hive agent send can split a message that the user types](entries/resolved/send-splits-typed-message.md) — agent.prompt pastes text and presses Enter into a prompt that already holds the partial text of the user
- [A report sent to a pane with an open dialog is lost](entries/resolved/send-blocked-by-dialog.md) — hive agent send refuses a target that a dialog blocks, and a worker that ends its turn never sends the report again
- [A test TMPDIR below the home lets tests reach the real home](entries/resolved/test-tmpdir-reaches-real-home.md) — Home discovery picks the outermost marker, so a t.TempDir() below the home resolves to the real home
