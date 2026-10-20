# Agents live generation

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, repair after wave C |
| Tier | decent |
| Depends on | 88-tui-terminal-io. Base: `dev` after T1 lands |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The Agents screen shows each generation of each agent in TREE, GRAPH, and
TABLE. On the live home it shows 6 coordinator rows, 13 supervisor rows, and
9 manager-hive rows. Most of them show `unknown · ?`. The user wants only
the current agents by default. The feature README is
`.hive/spec/hive-tui/README.md`. Read decision 1, G4, and G5. The board is
`/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Main.dc.html`. It shows one ghost, `t51-worker-2`. When the code does
not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/agents.go`, the Agents load in `internal/tui/`, and their tests | R1, R2, R3. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy, when `hive task create` did not write it. |

Do not edit other files. Do not change the screen layout. T4 does that.

## Requirements

### R1 Default set

By default, TREE, GRAPH, and TABLE show these agents only:

- For each agent identity (role, project, and task, or the name when the
  record has no lineage), the newest generation. Use `replaces` and
  `parent` when the record has them (G4). Else use the newest
  `status_changed_at`, then the newest record time.
- A generation that is not live and that a newer generation replaces is
  hidden.
- A stopped worker or verifier of a task that is not `landed` or `done`
  shows as a dim ghost, as `t51-worker-2` on the board.
- A stopped agent with no active task is hidden.

Write the exact rules that you use in the evidence.

### R2 History toggle

`h` toggles the history. With history on, the screen shows each record, as
today, with old generations dim. The footer shows the state of the toggle.
Decision 1 says that `h` hides stopped agents. This brief keeps the key and
changes the default: history is off at start.

### R3 Tests

- Write each test first and see it fail for the correct reason.
- Use fixture records with several generations of one coordinator, one
  supervisor, and one manager, one live worker, one stopped worker of an
  active task, and one stopped worker of a landed task. Write the expected
  rows as literal names.

## Not in scope

The layout of the board (T4). Agent records and `hive agent` commands.
`main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each R3 test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Build the task binary. Run it on this home in a herdr pane. Open Agents. Capture TREE, GRAPH, and TABLE with `herdr pane read`. Press `h` and capture again | With history off, one row for each current agent and ghosts only for stopped agents of active tasks. With history on, all records. |
| V3 | `git diff origin/dev --stat` | Only the files of the source table changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other commands exit 0. |

Close each herdr tab and workspace that you open. Do not close a pane that
you did not open.

## Finish

Commit on the task branch with a subject such as `fix(tui): show the live
agent generation`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V4 results to the
manager with `hive inbox send`. Read only your own inbox.

## Task names

This brief uses these names: T1 is 88-tui-terminal-io, T2 is 90-tui-idle-cpu,
T3 is 91-agents-live-generation, T4 is 92-agents-board-layout, and T5 is
89-screen-board-audit.

## Scope change: chat safety

A TUI check uses a scratch home (a temp copy of a marked home, set with
`HIVE_HOME`, and a temp `XDG_CONFIG_HOME`), or it does not open the Chat
input. On the real home, never press `i` on the Chat screen and never press
Enter there. A chat message from an agent has the authority of the user.
