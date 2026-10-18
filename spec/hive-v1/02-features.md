# Features

"v1" tells a manager what it can build toward without a question. "Later" tells
it where to leave a seam. "Out" stops speculative work.

## v1

| Feature | Meaning |
| --- | --- |
| Project registry | Register one Git repository by path. It gets a stable project ID and a project directory `projects/<project-id>/`. The repository stays at its path. The record keeps the base branch. |
| Roles | Coordinator, one manager for each project, and workers. Each role has a durable record, a claim with a generation number, and an inbox that survives an agent exit. |
| Tasks | Create, assign, complete, verify, and cancel. Dependencies between tasks. One central store and one lock. A task ID is a number and a short name, such as `12-spawn-worker`. |
| Worktrees | One linked Git worktree for each task at `projects/<project-id>/worktrees/<task-id>/`, made and shown through herdr. |
| herdr control | Spawn, send to, and stop agents through the socket. Make the workspace of a manager and the workspace of a worktree. An agent starts in the root pane of the first tab. A stop sends the exit keys of the harness, because herdr has no stop method. Hive closes what it made when the agent stops. |
| Supervise | One relay process for each home. It turns herdr events into inbox records and nudges an idle role. No agent runs a wait loop. |
| Memory and handoffs | Scoped Markdown files at home scope and project scope. A backend command writes each file atomically. |
| Delivery | Land a verified branch by local merge or by pull request, under a grant. |
| Presets | A freeform Markdown document. Each preset has a prose purpose and an ordered list of harness and model choices. The manager reads it and chooses. |
| Logging | JSON log events in daily files, debug always on. `hive logs` follows and filters them. The TUI shows them live, with the supervisor state, the live agent count, and the unread inbox count. |
| Doctor | One command that checks the environment: herdr, home, state store, git, supervisor, harnesses, and `gh`. |

## Later

| Feature | Meaning |
| --- | --- |
| TUI dashboard | Views of agents, tasks, and attention beside the live log. |
| TUI hosts the coordinator | The coordinator conversation runs inside the TUI. |
| qmeter in preset choice | A preset can name a qmeter provider. A manager reads remaining quota before it picks a preset. Unknown stays unknown. |
| Attention summary | One view of what needs the user across all roles. |
| Native trust and integration | Hive writes trust settings and harness integration files for the user, with approval. |
| Self-update and releases | Tagged releases and an updater, if the tool stops being personal. |

## Out

| Capability | Status |
| --- | --- |
| Remote or multi-machine execution | Out |
| Automatic task scheduler | Out. No process dispatches queued work. The supervisor relays events only. |
| A wait command for agents | Out. The supervisor nudges an idle role instead. |
| A project activate state | Out. A live manager claim means active. |
| A `hive init` command | Out. The tracked marker `.hive/README.md` makes every clone a home. |
| Budget engine | Out |
| Hard quota gates and automatic worker counts | Out. Managers choose counts. |
| Background updates | Out |
| Isolation from hostile agents that run as the same user | Out |
| Fledge compatibility or import of Fledge data | Out |
