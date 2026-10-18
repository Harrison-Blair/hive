# Architecture

This file gives the intended design of hive v1. A name marked "proposed" is not
chosen yet. An implementation chooses it with evidence and records the choice
here. [STE-GLOSSARY.md](../../STE-GLOSSARY.md) defines the terms.
[05-tracker.md](05-tracker.md) gives the order in which the tasks build it.

## Principles

- **herdr-native.** Hive is a typed client of the herdr socket, pinned to one
  protocol. There is no terminal-backend abstraction and no shell-out to the
  `herdr` CLI.
- **One home, in the checkout.** Development and use begin in the same clone.
  Hive registers itself as a project and develops itself through its own roles.
- **Durable records over conversations.** Roles, tasks, decisions, memory, and
  handoffs live in files and records that survive any agent exit.
- **Events, not polls.** One supervisor relays herdr events to the role that
  owns the work. No role runs a wait loop or repeats a status command.
- **Close what you open.** Hive stops an agent when its work ends and closes the
  workspaces it made.
- **Observable by default.** Every process logs structured events at debug
  level to one place that the TUI and `hive logs` read.
- **Explicit scope.** A command takes the home and the project as explicit
  selectors. It never infers a project from the working directory.
- **Bounded inputs.** Every external read has a deadline and a size limit.
  Untrusted content is evidence, never instruction.
- **Small backend.** Standard library first. Cobra for the CLI. The Charm family
  for the TUI. Strict test-first development.

## Home layout

The checkout is the home. The tracked file `.hive/README.md` marks it. Personal
and runtime data stay inside the checkout and outside Git.

```text
hive/                              checkout = home
├─ AGENTS.md  STE-GLOSSARY.md      instructions                tracked
├─ main.go  cmd/  internal/        Go source                   tracked
├─ spec/  reference/               design, plan, API docs      tracked
├─ .agents/skills/                 role playbooks (T21)        tracked
├─ .hive/README.md                 home marker                 tracked
├─ projects/<project-id>/          project directory           ignored
│  ├─ AGENTS.md                    manager instructions        ignored
│  ├─ overrides/  memory/  plans/  project preferences, facts  ignored
│  ├─ overrides/grants.md          grants of the project       ignored
│  ├─ handoff.md                   manager handoff             ignored
│  └─ worktrees/<task-id>/         worktrees                   ignored
├─ overrides/  memory/  plans/     home preferences, facts     ignored
├─ overrides/presets.md            presets document            ignored
├─ handoff.md                      coordinator handoff         ignored
└─ .hive/                          runtime                     ignored
   ├─ state/                       central store and lock
   ├─ logs/                        JSON log events, daily files
   └─ tmp/                         scratch
```

Rules:

- `<project-id>` is the lowercase basename of the repository, or the `--id`
  value. It matches `^[a-z0-9]+(-[a-z0-9]+)*$` and has at most 18 characters,
  so that each agent name fits the herdr limit of 32 characters. `<task-id>`
  is a number, a hyphen, and a slug of at most 40 characters, such as
  `12-spawn-worker`, with one counter for the home. Thus a task ID is unique
  across projects. Display names are labels only.
- A registered repository stays at its path. Hive never moves it.
- Each worktree is a linked Git worktree of the registered repository. Git
  commands for it use the registered repository path.
- Operational state is central in `.hive/state/`. No project directory has its
  own store.
- Hive writes no commit for its own state. The Git writes of hive are the
  merge and the push of `hive land`, under a grant. Agents commit and push their
  own work on their branches without a request from the user.

## Home resolution

Hive never creates a home. A clone of this repository is a home because Git
tracks `.hive/README.md`. Each command resolves one home in this order: the
`--home` argument, then `HIVE_HOME`, then the outermost ancestor directory of
the working directory that contains the marker. The outermost rule matters
because a worktree of this repository also contains the marker, and every
worktree sits inside the real home. A named path without the marker fails with
one error line. A `--home` value wins without a conflict check. A `HIVE_HOME`
value that conflicts with an ancestor marker fails. A worker brief carries the
absolute home path, and the worker passes `--home` to each command.

## Roles

| Role | Count | Duties | Working directory |
| --- | --- | --- | --- |
| User | 1 | Approves goals, constraints, grants, and material decisions | Any |
| Coordinator | 1 for each home | Talks to the user. Spawns managers. Starts the supervisor. Keeps preferences and shared memory. Reports open decisions and status. | The home root |
| Manager | 1 for each active project | Owns research, planning, implementation, documentation, independent checks, and reports for approved goals | `projects/<project-id>/` |
| Code worker | Chosen by the manager | Changes project source for one task | `projects/<project-id>/worktrees/<task-id>/` |
| Verifier | 1 for each completed task | Re-runs the checks of one task and records the result | The worktree of the task |
| Read-only worker | Chosen by the manager | Researches or reviews without a change to source | `projects/<project-id>/` |

Rules:

- The user talks to the coordinator only.
- A manager delegates every change to project source to a worker. It dispatches
  routine tasks without approval for each wave. It escalates each material
  decision.
- An active manager continues approved work when the coordinator stops.
- The worker and the verifier of one task take turns in the same worktree.
- An agent takes its role from its assigned role and its durable registration.
  A manager or worker never runs coordinator startup, also after a context reset.
- The bootstrap clause in `AGENTS.md` adds a temporary developer role for a root
  session. It ends when the hive manager spawns its first worker.

### Role claims

A singleton role (the coordinator, the manager of one project, the supervisor)
has a claim record with a generation number. A new session claims the role by
incrementing the generation under the store lock. A claim with an older
generation is stale. Each claim generation has its own agent record, with the
key `<name>-g<generation>`, so a stale record and the live record both exist.
The claim functions make no herdr request. Tasks and agent
records name the controlling role and the acting agent separately. Inbox
records go to the controlling role, so a replacement agent inherits them.

## herdr layout

- The **hive session** is the one herdr session that hosts the home. The
  coordinator and the supervisor run in panes of that session.
- Each managed project has one workspace with the label `hive:<project>`. Hive
  makes it when it spawns the manager, with the project directory as the
  working directory. The manager runs in the root pane of the first tab. Hive
  reuses the workspace that the project record names when it still exists.
- Hive makes each worktree with `worktree.create`. The source is the registered
  repository path in `cwd`, never the `hive:<project>` workspace, whose
  directory belongs to the hive repository. The request gives an explicit
  `path` under `projects/<project-id>/worktrees/`, the branch
  `task/<task-id>`, the base branch of the project, and the task ID as the
  label. herdr opens the worktree as a workspace beside the repository
  workspace and draws that grouping itself.
- The first tab of a hive workspace is never an empty shell. Hive starts the
  agent in the root pane that herdr returns.
- herdr has no stop method. `agent stop` marks the record `stopped` and closes
  the pane of the agent with `pane.close`. herdr ends the process, and it
  closes the workspace when that pane was its last pane. Hive closes only a
  pane in which it started the agent. An agent that the user registered by
  hand exits by hand.
- Hive closes what it made. `worktree remove` closes the worktree workspace
  when it is still open, then removes the worktree with Git. The stop of a
  manager closes its pane, and herdr closes `hive:<project>` when no other
  pane is in it.
- Adoption of a workspace that already exists for a repository needs user
  approval, through `--adopt <workspace-id>` on `hive agent spawn`.

## Agent lifecycle

```text
t0  agent spawn: workspace or worktree, agent in the first tab, record, claim
t1  the agent reads its brief through hive task show --brief
t2  task complete: record, inbox record to the manager, agent marked finished
t3  the supervisor stops the finished agent when idle, its pane and workspace close
t4  verifier spawn reopens the worktree, task verify, verifier finished and stopped
t5  land: merge or pull request under the grant
t6  worktree remove: workspace closed when open, directory gone, branch kept
```

Hive stops an agent only on its own record transitions, never on a herdr status
alone. A herdr status of idle or done never means success. The stop checks the
foreground process identity from the spawn. A pane and terminal match is not an
ownership check.

## Supervision

`hive supervise` is one relay process for each home. The coordinator starts it
when its claim is stale, in a pane of the hive session. It holds the
`supervisor` claim and one herdr event subscription. herdr cannot change a
subscription, so the supervisor opens a new one when the set of watched panes
changes. It watches the inbox and agent directories of the store with inotify.
It never polls.

| Input | Action |
| --- | --- |
| Agent status `blocked` for a registered agent | Inbox record `blocked` to the controlling role |
| A registered agent leaves its pane and the pane shows a shell (`pane_agent_detected` with `released`) | Inbox record `exited`, agent record `exited` |
| The pane, tab, or workspace of a registered agent closes | Inbox record `unresolved`, agent record `unresolved` |
| Idle for the quiet period with no completion record | Inbox record `quiet`, once for each idle period. Default 10 minutes, set at spawn |
| Idle agent with the mark finished | The T10 stop function: identity check, agent record `stopped`, `pane.close`, then a bounded post-close topology read |
| Unread inbox records for a role whose agent is idle | One nudge: one line typed into that pane that names `hive inbox` |
| Unread inbox records for a role whose agent is working | Wait. Nudge when the role turns idle |
| Start | Reconcile: compare live herdr agents with the records. Mark a record with no pane `unresolved` |

A long `working` status is never an event. A notification group increments the
inbox counter, writes its inbox record, then writes its stored status. A working
transition stores the status only. The group stores every relevant transition.
It does not store `working` only when a quiet timer runs. A nudge uses one external prompt, a fresh claim-holder and generation
check, and one mark write for each record. The supervisor reports the known
prefix after a failure. A failed nudge, such as
`agent_prompt_stalled`, is a log event and a retry at the next idle transition.
When the supervisor exits, the TUI and `hive doctor` show it, and the
coordinator starts it again at its next startup. Nothing restarts it in a loop.

## State store and tasks

### Dated decision: store operations (2026-10-18)

The approved trust boundary is a local user and a cooperating writer. A hostile
file replacement between `Lstat` and `Open` stays deferred.

The store uses one method matrix:

- `Store.Read` and `Store.ReadIn` do not acquire the lock.
- `Store.List` and `Store.ListIn` do not acquire the lock. A standalone list sees
  one directory snapshot.
- A read of related records uses one `WithLock` callback and its handle.
- Each standalone mutation uses the bounded acquisition path. This rule covers
  the legacy `Write`, `Update`, and `Next`.
- A `LockedStore` method uses the held lock and never acquires another lock.

Acquisition defaults to two seconds. A context deadline can shorten or extend the
wait. A cancellation returns the context error. An acquisition failure changes no
record. No caller holds the lock for a model wait, a Git wait, a network wait, or
a stdin wait.

Before a read, list, or create, the store checks the state root, each collection
ancestor, the lock, and the selected record path with `Lstat`. A symbolic link or
a wrong type fails before the open. `WalkRecords` rejects a linked collection or
record with its path and does not descend the link. The atomic temporary name
starts with a dot and contains `.tmp-`. `StoredRecord.Name` includes the `.json`
suffix. A colon name such as `counters/inbox:coordinator.json` remains visible.

Dated correction (2026-10-18): the record writer rejects an encoded record above
1048576 bytes before `atomicWrite`. The encoded record is the JSON object and one
final newline. The error is `record too large: <path>`. An encoded record of
exactly 1048576 bytes is valid. A rejected write keeps the old record bytes and
makes no new record. The legacy, generic, and locked write and update paths use
this check. A multi-write caller reports its earlier effects. The reader and lock
contracts do not change. The store adds no title limit, no journal, no rollback,
no public API, and no dependency.

A multi-record operation writes in a safe order and names the last write as its
commit marker. A completed write stays after a later error. The command reports
`partial`, the actual failed phase, and each completed effect. No journal and no
rollback exist. A later reader can see the completed prefix of a failed group.

- One store under `.hive/state/`, one `flock` lock, one JSON file for each
  record: `projects/`, `tasks/`, `briefs/`, `agents/`, `claims/`,
  `inbox/<role>/`, `counters/`. Each read of a record has a size limit.
  Partitions wait for measured evidence of contention.
- A task has an ID, a title, a brief, a project, a controlling role, an acting
  agent, a status, dependencies, verifications, a worktree, a delivery, and
  timestamps. Status moves created, assigned, completed, verified, with cancel
  as the early exit and a failed verification as the step back to assigned.
  Other transitions fail.
- Dependencies between tasks form a graph. The store rejects a cycle. A ready
  list shows tasks with every prerequisite verified.
- `verify` refuses the owner of the task and an unregistered caller. A repeated
  verification replaces the earlier one.
- A caller identity comes from `--agent`, else from `HERDR_PANE_ID` and the
  agent records.
- Writes to a record or a Markdown file are atomic: write a temporary file,
  then rename.

### Dated decision: process, finish, and relay contracts (2026-10-18)

A terminal ID identifies a terminal. It does not identify the agent process. A
new process of the same kind can keep the name. A suspended agent can leave a
settled status while a shell holds the foreground.

T10 owns `internal/lib/process/`. The spawn captures the typed
`pane.process_info` result after bounded evidence that the harness holds the
foreground. It records `process_pid`, `process_name`, `process_start`, and an
optional `agent_session_ref`. The profile gives the process names that can
identify its harness. The capture requires exactly one foreground entry with a
profile name. The entry PID must belong to the foreground process set. The
protocol does not require the entry PID to equal the foreground process group
ID. A multiple, absent, shell, unknown, wrapper-only, or unstable entry gives
`unresolved`. The capture permits an absent session reference. A recorded
session that is missing or different fails. A wait timeout with independently proven
ownership can retain the live partial-readiness result. Explicit stop and
automatic stop use the same identity check. A manual registration refusal sends
no close.

After a checked `pane.close`, a bounded `workspace.get` outside the store lock
returns `closed`, `surviving`, or `unknown` topology plus the known effects. A
`closed` result clears the project workspace. A `surviving` result keeps it. An
`unknown` result keeps the metadata and reports the pane-close effect. Do not
infer closure from an old pane count.

A finish group checks both lookup routes under one held handle. It checks the
caller identity, the role, and the allowed transition before any write. It also
checks the project, the task, the claim key, the holder record key, and the
generation. It then writes the task transition, the inbox counter, the inbox
record, and the finished mark in that order. A first task-write failure has no
effects. A later failure is `partial` with the exact completed prefix, the
actual phase, the actual new effects, and a nil partial Go error. A successful
durable group with a failed stdout exits 1. It writes a working stderr with the
known effects and the inspection commands. It does not replay. A retry uses
normal validation only. After `verify --fail` moves a task to `assigned`, the
same verify command refuses with no writes, also when the reason changes. A
fresh worker assignment and its `complete` are a separate operation. A known
finished agent cannot start a new finish attempt. T11 owns that narrow assignment
refusal. T03 stays independent of the agent record. Fresh-agent and
unknown-agent assignment behavior does not change. No `finish_seq`,
`finish_key`, `relay_seq`, `event_key`, journal, automatic rollback, or automatic
completion exists.

A supervisor notification group increments the inbox counter, appends the inbox
record, then writes the stored status, under one held handle. It uses this order
for each notification kind, including `exited` and `unresolved`. A working
transition stores the status only. A first counter failure has no effects. A record failure is
`partial` with the counter gap. A status failure is `partial` with the counter
and the record. Every storage failure closes the subscription, exits 1, and
stops queued processing. An explicit restart reconciles the current state only.
A gap can lose transition history. A committed inbox prefix can duplicate a
notification after restart. A restart starts a new quiet window. Normal
uninterrupted delivery stays once for each observed period.

A nudge prompt, a stop request, and a subscription change run outside the store
lock. After the prompt succeeds, recheck the claim holder and the generation.
Then write the nudge marks in one atomic write for each record. A successful
prompt followed by a failed mark can repeat the nudge on an explicit retry.
State this limit.

## Worktrees

- One task, one worktree, one branch. The path is
  `projects/<project-id>/worktrees/<task-id>/`.
- A manager workspace can survive a stop with no recorded root pane. Hive keeps
the workspace ownership metadata and refuses the next automatic manager spawn
with an explicit `--pane` instruction. Hive does not select a shell by itself
and does not close a user tab.
- Removal keeps a dirty or unmerged worktree and a worktree with an active or
  unresolved agent. `--force` overrides the Git guards only.
- A removal compares the recorded path and branch with the Git state of the
  worktree, and with the worktree data of herdr when the workspace is still
  open. A mismatch refuses, also with `--force`.

## Memory and handoffs

- Memory is one fact for each Markdown file, at home scope or at project scope,
  with frontmatter and a generated `MEMORY.md` index beside the files.
- Each role has one handoff: `handoff.md` in the home for the coordinator,
  `projects/<project-id>/handoff.md` for a manager. A handoff summarizes and
  refers to records. Records and the repository state win over a stale handoff.
- Agents compose the Markdown. A scoped backend command writes it atomically.
  The command requires an explicit scope argument and does not authenticate the
  role of the caller.
- A worker gets its brief and selected context, not the whole memory index.

## Delivery and grants

- A grant is a standing user approval for named actions (push, pull request,
  merge) on one project and named target branches. Grants live in
  `projects/<project-id>/overrides/grants.md` as text the manager reads. The
  backend records the grant that a landing cites and enforces nothing.
- Each action outside a grant needs explicit approval for that action.
  Deployments, releases, spending, and external messages always need explicit
  approval.
- Delivery modes: a local merge into the base branch in the registered
  repository, then a push of the base branch to `origin`, or a pull request
  through `gh`. A landing needs a verified task.

## Presets

`overrides/presets.md` is a freeform Markdown document. Each preset is a heading,
a prose paragraph that says what kind of task it serves, and a list of harness
and model choices in order of preference. Hive ships a default document with
three tiers: a strong model for planning and verification, a decent model for
coding, a small model for small tasks. A project can add
`projects/<project-id>/overrides/presets.md`. The backend does not parse these
files. A manager reads them and passes `--harness` and `--model` to
`hive agent spawn`. Independent checks prefer a different model family from the
implementer. Quota data from qmeter comes later.

## Logging

- Every hive process (a CLI command, the TUI, the supervisor) writes log events
  with `log/slog` as JSON lines to `.hive/logs/YYYY-MM-DD.jsonl`. Debug level is
  always on. Each event is one `write` call, so lines from several processes do
  not interleave.
- Logged: each socket request and reply, each state transition, each external
  process start and exit, each user action in the TUI, and each error.
- Fields when they apply: `component`, `command`, `request_id`, `project_id`,
  `task_id`, `agent_id`, `duration_ms`, `error`. A body (a prompt, a reply) is
  capped at 4 KiB with `truncated: true`. No secret is logged.
- `hive logs` follows and filters the files. Daily files rotate by name. Hive
  never deletes a log file.

## TUI

Bare `hive` opens the TUI inside a herdr pane. Its first increment is a live
log view over `.hive/logs/` with a header that shows the home, the supervisor
state, the live agent count, and the unread inbox count. Next it shows agents,
tasks, and attention beside the coordinator agent. Later it hosts the
coordinator conversation itself. The TUI is a client of the same store, logs,
and socket as the CLI, with no logic of its own.

## Command contract

A command prints text for people by default. With `--json` it prints one object
with `outcome`, `phase`, `effects`, `error`, and `result`. `outcome` is `ok`,
`failed`, or `partial`. `phase` names the step where the command stopped.
`effects` lists what changed, even on failure, so a caller can inspect
resources after a partial outcome before it retries. `result` carries the data
of the command. Exit status 0 means `ok` only.

## Command tree

Nouns before verbs. `--home PATH` and `--json` are flags of every command. The
task commands take `--agent NAME` to name the caller. The last column names the
tasks in [05-tracker.md](05-tracker.md) that build the command.

| Command | Purpose | Tasks |
| --- | --- | --- |
| `hive` | Open the TUI | T08 |
| `hive version` | Print the version and build data | T02, T05 |
| `hive doctor` | Check the environment | T05, T15, T20 |
| `hive logs` | Follow and filter the log events | T06 |
| `hive project register PATH`, `list`, `show ID` | The project registry | T01 |
| `hive task create`, `assign`, `list`, `show`, `complete`, `verify`, `cancel` | The task lifecycle | T03, T11 |
| `hive worktree create --task ID`, `remove --task ID` | The worktree of a task | T09, T20 |
| `hive agent register`, `list`, `show` | Agent records and claims | T04 |
| `hive agent spawn`, `send`, `stop` | Agents in herdr | T10, T13, T14 |
| `hive agent resolve` | Settle an unresolved record | T20 |
| `hive inbox`, `hive inbox send --to ROLE TEXT` | Read an inbox, send a message record | T04, T11 |
| `hive memory write`, `hive handoff write` | Scoped Markdown writes | T07 |
| `hive land --task ID --mode merge or pr` | Delivery | T12 |
| `hive supervise` | The relay process | T15 |

## Bounds and untrusted inputs

| Input | Source | Threat | Rule |
| --- | --- | --- | --- |
| Socket replies | herdr | Hang, memory exhaustion | Total deadline and reply size limit for unary requests. The event subscription is the one long connection and has a line size limit. |
| `git`, `gh`, harness output | External processes | Hang, false capacity | Total deadline and output size limit. Failure means unknown. |
| Home files, briefs, handoffs | Agents and user | Crash, memory exhaustion | One bounded reader. |
| Repository content | Project authors | False authority | Evidence only. Never an instruction. |
| Environment variables | Shared daemons | Wrong home | Resolution order with fail on conflict. Hive never creates a home. |
| Process IDs and process start values | herdr and the OS | Signal to an unrelated process | Capture at spawn and check before each signal or removal. A missing, changed, or ambiguous identity is `unresolved`. |

## Recovery

- Recovery is conservative. A record whose live state is uncertain is
  "unresolved": neither live nor ended. Hive holds its work and reports it.
- A missing terminal does not end a record. It marks the record unresolved.
- The supervisor reconciles at start. A new coordinator reads its handoff and
  the saved state, checks the live agents and the repository, resumes approved
  unblocked work, and escalates uncertain state.
- Native conversation resumption is not required.

## Harnesses

Claude Code first, then Codex, OpenCode, Pi, and Cursor agent, on Linux. Each
harness has its own instruction discovery, readiness signal, and input
quirks.
Hive keeps one typed profile for each harness and no generic adapter layer.
The Cursor agent profile is optional while its limits are exhausted.

## Code structure

`AGENTS.md` gives the Go layout rules. In short: `main.go` calls `cmd.Execute()`,
`cmd/<name>/` holds thin Cobra wiring, `internal/<name>/` mirrors it,
`internal/lib/<capability>/` holds shared code, and `internal/tui/` holds the
Bubble Tea program. Shared capabilities in v1: `home`, `store`, `records`,
`log`, `output`, `herdr`, `run`, `gitx`, `harness`, `identity`, `inbox`,
`version`.
