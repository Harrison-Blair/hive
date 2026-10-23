# Dogfooding friction log

This folder records bugs, missing capabilities, and workarounds from hive use.
The coordinator and manager can edit files in this folder directly. They can add or extend an entry.
To resolve an entry, set its status, move its file to `entries/resolved/`, and update its row.
Commit on `dev` with the subject `docs(dogfood): SUMMARY`. Push `dev`.
Do not use a worker or verifier for this edit. Edit only this folder in this way.
For all other files, the coordinator and manager never change a project source file.

| Issue | Status | Decision |
| --- | --- | --- |
| [An agent can send a user chat message through the TUI](entries/agent-can-send-user-chat.md) | Open |  |
| [A verified task that never landed cannot be cancelled](entries/resolved/verified-task-cannot-cancel.md) | Resolved 2026-10-20 | Task 105 adds cancel from verified and the already-merged land |
| [A brief tells a worker to read its inbox, but a worker has no inbox](entries/worker-has-no-inbox.md) | Open |  |
| [Bare hive in a scratch home starts real coordinator and supervisor tabs](entries/scratch-hive-starts-real-agents.md) | Open | Task 8-no-startup adds the opt-in switch --no-startup and HIVE_NO_STARTUP=1 |
| [A stale coordinator record blocks bare hive startup](entries/resolved/stale-coordinator-blocks-startup.md) | Resolved 2026-10-21 | Task 2 makes bare hive check the coordinator pane at startup, and task 3 adds the r restart on the Agents screen |
| [The --agent flag of agent send names the target, but names the sender in inbox send](entries/agent-send-flag-means-target.md) | Open |  |
| [A spawn after a stop fails with first tab busy, and a retry works](entries/spawn-first-tab-busy-after-stop.md) | Open |  |
| [TestHealthProgramUsesStringReader fails at times on TempDir cleanup](entries/health-program-test-flakes.md) | Open |  |
| [TestUpdateCallbackRunsUnderLock fails at times under heavy load](entries/store-lock-test-flakes-under-load.md) | Open |  |
| [TestStartupFailureShowsInFooterAndDoesNotStopTUI fails at times in full package runs](entries/startup-failure-footer-test-flakes.md) | Open |  |
| [herdr pane send-keys has no page up or page down key name](entries/herdr-send-keys-no-page-keys.md) | Open |  |
| [A scratch coordinator fails with agent_name_taken on the live herdr server](entries/coordinator-name-clash-in-scratch-home.md) | Open |  |
| [Each brief copy row conflicts in the STE baseline](entries/ste-baseline-row-conflicts.md) | Open |  |
| [The local gate does not run the module tidy check of CI](entries/local-gate-misses-ci-tidy.md) | Open |  |
| [A completed task cannot be reopened or edited](entries/completed-task-cannot-reopen.md) | Open |  |
| [A failed verification is lost when the next verification passes](entries/verify-fail-lost-on-pass.md) | Open |  |
| [store.List skips an agent record with a key longer than 18 characters](entries/resolved/store-list-skips-long-keys.md) | Resolved 2026-10-20 | Task 101 keeps long keys |
| [Hive cannot spawn a reviewer or another one-off agent](entries/resolved/no-reviewer-role.md) | Resolved 2026-10-21 | Goal g3-general-agent adds the role agent, hive agent report, and the hive-agent skill |
| [A new task can go only in a feature that has a README](entries/old-features-lack-readme.md) | Open |  |
| [A supervisor nudge splits a message that the user types](entries/nudge-interrupts-user-typing.md) | Open |  |
| [No dashboard shows the agents, their lineage, their activity, and their tasks](entries/no-agent-dashboard.md) | Open |  |
| [Claude Code in bypass mode still asks before a dangerous rm](entries/bypass-still-prompts-dangerous-rm.md) | Open |  |
| [The verifier skill does not give the path of VERIFIED.md](entries/verified-md-path-unspecified.md) | Open |  |
| [The supervisor keeps running an old hive binary after a rebuild](entries/resolved/supervisor-runs-stale-binary.md) | Resolved 2026-10-21 | Task 3 adds the r restart for the supervisor and an old-binary mark in the TUI header |
| [The worker skill gate adds golangci-lint that the project gate does not have](entries/worker-gate-adds-golangci-lint.md) | Open |  |
| [A feature spec task cannot start in a repository that ignores .hive/spec/](entries/resolved/spec-task-needs-allowlist-first.md) | Resolved 2026-10-20 | Fixed in 70-spec-allowlist (e1276caa): a spec task worktree stages the brief with git add -f and reports allowlist_needed; the spec task commits the .gitignore rules. |
| [A manager stops on a harness permission block and the coordinator does not know](entries/resolved/manager-blocked-by-harness-permission.md) | Resolved 2026-10-19 | Fixed in 29-message-marker (5c8f6846): Claude Code and Codex spawns launch with the bypass flag, each hive message starts with the marker, and a manager sends a blocked record before it stops. |
| [An agent spawn can skip the first prompt](entries/resolved/agent-spawn-skips-first-prompt.md) | Resolved 2026-10-19 | Fixed in 28-first-prompt-fix (642fe78d): the first prompt uses the herdr prompt wait until working or blocked, with one retry on a stall. The user accepts the risk of a duplicate prompt. |
| [No `hive task edit --brief` command](entries/resolved/task-edit-brief-command.md) | Resolved 2026-10-19 | Fixed in 22-task-edit-brief (77cdf44d): hive task edit --task ID --brief FILE, only for created or assigned, only by the controlling role. |
| [A verifier can record a verdict without a pushed VERIFIED.md](entries/resolved/verifier-verdict-without-commit.md) | Resolved 2026-10-19 | Fixed in 23-verify-clean-branch (27a883e9): hive task verify refuses a dirty worktree or an unpushed branch, and the verifier commits and pushes VERIFIED.md first. |
| [No hive path repairs a verified task](entries/resolved/no-hive-path-repairs-verified-task.md) | Resolved 2026-10-19 | Fixed in 18-task-reopen (4f2bd6a5): hive task reopen moves a verified task to assigned, keeps the old check in verification_history, and land needs a new pass verify. |
| [An exited agent name blocks reuse](entries/resolved/exited-agent-name-blocks-reuse.md) | Resolved 2026-10-19 | Fixed in 19-worker-respawn (78381ba1): with no --name, hive picks the next free name tNN-ROLE-N. |
| [A verify fail leaves no workspace for the repair worker](entries/resolved/verify-fail-repair-worker-no-workspace.md) | Resolved 2026-10-19 | Fixed in 19-worker-respawn (78381ba1): the spawn reopens the worktree workspace with worktree.open and updates the record. |
| [A worker spawn into an assigned task keeps the old acting agent](entries/resolved/worker-spawn-keeps-old-acting-agent.md) | Resolved 2026-10-19 | Fixed in 19-worker-respawn (78381ba1): the spawn sets the acting agent and refuses while the old worker is live. |
| [Any agent can read and consume the inbox of another role](entries/resolved/any-agent-read-another-role-inbox.md) | Resolved 2026-10-19 | Fixed in 15-inbox-ownership (1d97f21f): only the owner's read marks records read. The coordinator can read a manager inbox and marks nothing. Any other caller gets an error that names the owner. |
| [An idle session burns tokens on supervisor nudges](entries/resolved/idle-session-supervisor-nudges.md) | Resolved 2026-10-19 | Fixed in 16-quiet-no-nudge (848ef0b7): an unread quiet record never causes a nudge. |
| [A verifier spawn with `--pane` requires `--project`](entries/resolved/verifier-spawn-pane-requires-project.md) | Resolved 2026-10-19 | Fixed in 17-verifier-project (e7031497): hive derives the project from the task, and the docs are fixed. |
| [Full Pi worktree, agent, and delivery bootstrap flow is missing](entries/resolved/full-pi-bootstrap-flow.md) | Resolved 2026-10-19 | Stale: `hive worktree` and `hive agent spawn/send/stop` exist. |
| [A verify fail leaves no workspace and no verifier for a re-verify](entries/resolved/verify-fail-no-reverify-workspace.md) | Resolved 2026-10-19 | By design. The manager skill fix landed in 17-verifier-project (e7031497): after a verify fail, spawn a fresh verifier with no --pane. |
| [Shell-ready race in worker and verifier spawns](entries/resolved/shell-ready-spawn-race.md) | Resolved 2026-10-19 |  |
| [Manager spawn fails with `agent_pane_busy` right after `workspace.create`](entries/resolved/manager-spawn-shell-ready-race.md) | Resolved 2026-10-19 |  |
| [The user cannot answer a structured question in the Chat screen](entries/chat-has-no-question-prompt.md) | Open |  |
| [The Chat message box does not grow or scroll while the user types](entries/chat-input-does-not-grow.md) | Open |  |
| [A supervisor restart does not stop the old supervisor or close its tab](entries/supervisor-restart-keeps-old-supervisor.md) | Open |  |
| [The Chat screen does not show that the coordinator is working](entries/chat-shows-no-coordinator-activity.md) | Open |  |
| [Attach on the supervisor fails with agent_not_found, and the error stays in the footer](entries/attach-fails-on-supervisor.md) | Open |  |
| [The user must ask an agent to change a hive setting](entries/config-needs-an-agent.md) | Open |  |
| [Hive has no command to shut down a session](entries/no-shutdown-command.md) | Open |  |
| [Up and Down in a wrapped chat input can land one word off](entries/chat-input-updown-wrap-offset.md) | Open | Later: the user deferred a fix on 2026-10-22 |
| [A manager cannot send a completed task back to a worker before the verify](entries/completed-task-no-repair-before-verify.md) | Open |  |
| [Agents overview draws a card from a stale record instead of the live one](entries/agents-card-shows-stale-record.md) | Open |  |
