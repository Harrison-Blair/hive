# Dogfooding friction log

This folder records bugs, missing capabilities, and workarounds from hive use.
The coordinator and manager can edit files in this folder directly. They can add or extend an entry.
To resolve an entry, set its status, move its file to `entries/resolved/`, and update its row.
Commit on `dev` with the subject `docs(dogfood): SUMMARY`. Push `dev`.
Do not use a worker or verifier for this edit. Edit only this folder in this way.
For all other files, the coordinator and manager never change a project source file.

| Issue | Status | Decision |
| --- | --- | --- |
| [No hive path repairs a verified task](entries/no-hive-path-repairs-verified-task.md) | Open | Add `hive task reopen` (verified -> assigned, re-verify required) (wave 3). |
| [An agent spawn can skip the first prompt](entries/agent-spawn-skips-first-prompt.md) | Open | A report task investigates first. The fix is decided after the report. |
| [No `hive task edit --brief` command](entries/task-edit-brief-command.md) | Open | Add `hive task edit --brief`, allowed only before `completed` (wave 4). |
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
