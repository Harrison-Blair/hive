# Dogfooding friction log

This folder records bugs, missing capabilities, and workarounds from hive use.
The coordinator and manager can edit files in this folder directly. They can add or extend an entry.
To resolve an entry, set its status, move its file to `entries/resolved/`, and update its row.
Commit on `dev` with the subject `docs(dogfood): SUMMARY`. Push `dev`.
Do not use a worker or verifier for this edit. Edit only this folder in this way.
For all other files, the coordinator and manager never change a project source file.

| Issue | Status | Decision |
| --- | --- | --- |
| [Claude Code in bypass mode still asks before a dangerous rm](entries/bypass-still-prompts-dangerous-rm.md) | Open |  |
| [The verifier skill does not give the path of VERIFIED.md](entries/verified-md-path-unspecified.md) | Open |  |
| [The supervisor keeps running an old hive binary after a rebuild](entries/supervisor-runs-stale-binary.md) | Open |  |
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
