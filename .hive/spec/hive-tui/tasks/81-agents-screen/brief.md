# Agents screen

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, wave B |
| Tier | strong |
| Depends on | 79-tui-shell. Base: `dev` after it lands |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Fill screen 2, Agents: a topology of the agents with the tabs tree, graph, and
table. The feature README is `.hive/spec/hive-tui/README.md`. Read it first,
mainly goal decision 1 and G4, G5, G12, G14. Use its approved decisions. Do not change one. When the code does not agree with this brief or the
README, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Sources

- The design in the home: `.hive/tmp/coordinator/designs/hive-tui/Main.dc.html`.
  The decisions win over the design. Do not make `t open task` or the nudge
  edges in this task.
- The shell: `internal/tui/shell.go` (the `screen` interface and attach).
- The agent record: `internal/lib/records/agent.go`, with `parent`,
  `replaces`, and `status_changed_at`. The goal record:
  `internal/lib/records/goal.go`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/` (new files for the screen, and the shell wiring for it) and its tests | R1-R6. |
| `testdata/ste-baseline.json` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### R1 Groups and edges

- Groups: `home` (the coordinator and the supervisor), then one group for each
  project, in the order of the project records. The group title has the
  project ID and the title of its active goal, when it has one.
- An edge goes from `parent` to the agent. A `replaces` link is a separate
  edge kind.
- G4: an agent with no `parent` gets an `inferred` edge: a worker or a
  verifier from the manager of its project, a manager from the coordinator.
  An inferred edge has its own style and the label `inferred` in the detail
  panel.

### R2 Status

- A mark for each agent: working, idle, blocked, stopped, or unknown. G5:
  `done` shows as idle. When the supervisor claim is stale, each agent shows
  unknown. The age is the time since `status_changed_at`, else `?`.
- Stopped and exited agents are dim. `h` hides them and shows them again. The
  default hides them.

### R3 Tabs

- `tab` cycles tree, graph, and table. The tab bar shows the current tab.
- Tree: the groups, with each agent below its parent, indented.
- Graph: one box for each group. In each box, nodes with the name, the
  harness, and the task or the age. Lines connect a parent to its children
  with box-drawing characters. When the graph does not fit, it scrolls with
  the arrow keys and keeps the selected node in view.
- Table: one row for each agent: name, role, project, task, harness, model,
  status, age, parent.

### R4 Selection and detail

- The arrow keys (and `j`, `k`) move the selection. A detail panel shows the
  selected agent: parent (or `inferred`), replaces, model, task and its
  status, pane ID, and age.
- `enter` shows the last 20 lines of the pane of the selected agent, read with
  `herdr pane read` in a `tea.Cmd`. `esc` closes it.
- `a` attaches to the pane of the selected agent with the shell attach.

### R5 Loading

All reads run in a `tea.Cmd` through the shell `DataRequest`. The screen keeps
the selection by agent key across a refresh.

### R6 Tests

- Test at the model with a temp home that holds agent, project, goal, and
  supervisor records, with and without lineage fields.
- Cover each R1-R4 rule. Assert each rendered graph line of one small fixture
  as a literal string.
- A test with 40 agents: the graph scrolls and the selected node stays in
  view.
- A program test on a string reader.

## Rules

- Each Go change is test-first. See each new test fail for the correct
  reason.
- When `ste_test.go` fails on the brief copy, change the words of each
  sentence in the copy that causes a violation. Keep its meaning. Record each
  change in the evidence.
- Live checks use a temp `HIVE_HOME` and a temp `XDG_CONFIG_HOME`, and clean
  up each herdr workspace that they make. Do not run `rm -rf` on an unquoted
  variable.


## Not in scope

Nudge edges. `t open task`. Any action other than attach and peek. `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each new test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Live, in a new herdr workspace: run the task binary against the real home with `HIVE_HOME` set to it (read only). Press `2`, then `tab` two times, `h`, the arrow keys, `enter`, `esc`. Read the pane with `herdr pane read` after each step. Do not press `a` on a real agent. | Each tab shows the live agents with their parents. The peek shows pane lines. Clean up the workspace. |
| V3 | Live, with a temp `HIVE_HOME` and one agent record in a scratch pane: select it and press `a` | herdr focuses the scratch pane. Clean up the workspace. |
| V4 | `git diff dev --stat` | Only the files of the source table and the evidence changed. |
| V5 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `feat(tui): add the Agents
screen`. Push the task branch. Write the evidence. Run `hive task complete`.
Send the commit SHA and the V1-V5 results to the manager with
`hive inbox send`.
