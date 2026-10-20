# Goals and Queue screens

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, wave C |
| Tier | decent |
| Depends on | 80-chat-screen, 81-agents-screen. Base: `dev` after both land |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Fill screen 3, Goals, and screen 4, Queue. The feature README is `.hive/spec/hive-tui/README.md`. Read it first,
mainly goal decisions 2, 8, 9, 13, 17, 24 and G7, G8. Its decisions have approval. Do not change a decision. When the
code does not agree with this brief or the README, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Sources

- The design in the home: `.hive/tmp/coordinator/designs/hive-tui/Goals.dc.html` and `Queue.dc.html`. The
  decisions win over the design. G12: the TUI has no action except chat,
  attach, peek, and quit. Where a design shows another action, show the hint
  `ask in Chat` in the footer, or leave the action out.
- The shell: `internal/tui/shell.go` (the `screen` interface, the data
  requests, attach). The Chat and Agents screens are examples.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/` (new files for the screens, and the shell wiring) and its tests | The R items. |
| `testdata/ste-baseline.json` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### R1 Goals

- One panel for each goal record, active goals first, then done goals dim.
  The title line: the project, the goal title, and `done when: <point>` of
  the current wave (the first wave that is not done).
- For each wave: a stage line `dispatch ─▶ tasks N/M ─▶ <done-when point>`
  and the stages after it up to `release`, as in the design. `◎` marks the
  done-when point. The stages after `◎` are dim: they need the user. `✋` marks
  a wave with `needs_user`, with its text.
- Below the stage line: each task ID of the wave with a mark: `✓` landed,
  `▶` in work (assigned, completed, or verified), `·` queued (created), `✕`
  cancelled, and `✋` when the task has `needs_user`.
- "Landed" uses the README rule: the task delivery has a commit and `pushed`
  is true.
- A last panel `unassigned` lists the open tasks that no goal names, grouped
  by project.
- `↑` and `↓` select a goal.

### R2 Queue

- One panel for each project with open tasks, with its active goal and grant.
- Rows in run order: first `now` (assigned, completed, verified), then the
  created tasks in dependency order, with the column `starts when` (`N lands`,
  or `you` for a stop) and a state mark: `◆ auto` (no stop), `✋ you`
  (`needs_user` on the task), `┄ waits` (a task before it has `needs_user`).
- G8: a stop shows on the task with `needs_user` and on each task that
  depends on it, also through other tasks.
- `s` shows only the rows with a stop. `p` cycles the project filter.
- The footer says `order comes from the dependencies`.

## Rules

- Each Go change is test-first. See each new test fail for the correct
  reason. Test at the model: call `Update` and `View` with messages. Assert
  the rendered lines of one small fixture as literal strings. Add one program
  test on a string reader.
- All reads run in a `tea.Cmd` through the shell data request. A screen keeps
  its selection across a refresh.
- When `ste_test.go` fails on the brief copy, change the words of each
  sentence in the copy that causes a violation. Keep its meaning. Record each
  change in the evidence. Add the brief copy to `testdata/ste-baseline.json`
  with 0.
- Live checks use a temp `XDG_CONFIG_HOME`, and clean up each herdr workspace
  that they make. A live check against the real home only reads. Do not run
  `rm -rf` on an unquoted variable.
- Other wave C tasks run at the same time and edit the shell wiring in
  `internal/tui/`. Keep the change to shared files small.

## Not in scope

`enter decisions`, `o open design`, `enter open task` (show `ask in Chat` for
an action). `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each new test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | A test fixture with a chain A → B → C where B has `needs_user` | Queue shows `✋ you` on B and `┄ waits` on C. Goals shows `✋` on B. |
| V3 | Live, in a new herdr workspace: run the task binary with `HIVE_HOME` set to the real home. Press `3`, `↓`, `4`, `s`, `p`. Read the pane after each step. | Goal `g1-hive-tui` shows its three waves with the task marks. Queue shows the open tasks. Clean up the workspace. |
| V4 | `git diff dev --stat` | Only the files of the source table and the evidence changed. |
| V5 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `feat(tui): add the Goals and Queue screens`. Push the task
branch. Write the evidence. Run `hive task complete`. Send the commit SHA
and the V1-V5 results to the manager with `hive inbox send`.
