# Tasks evidence view

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, repair after wave C |
| Tier | decent |
| Depends on | 88-tui-terminal-io. Base: `dev` after it lands |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Task 89-screen-board-audit compared each screen with the board that the user
accepted. This task fixes the gaps below. Read the gap list in
`.hive/spec/hive-tui/tasks/89-screen-board-audit/evidence/GAPS.md` (on
`dev` after 89 lands, else on the branch `origin/task/89-screen-board-audit`)
and the board `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Tasks.dc.html`. The feature README is
`.hive/spec/hive-tui/README.md`. Its decisions and G1-G14 win over the
board. G12: the TUI has no action except chat, attach, peek, and quit, and
read-only views. When the code does not agree with this brief, stop and report
it with `hive inbox send --to manager-hive TEXT --agent NAME`.

## Gaps to fix

- T1: Separate task detail and evidence panels, as on the board. The evidence panel shows the brief and the verification history.
- T2: `v` opens `VERIFIED.md` and `d` shows the dependency graph of the selected task, as read-only views.
- T3: Add short definitions for `landed` and `verified`. Use G8 for `landed`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/task_screen.go` and their tests | The gaps. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy, when `hive task create` did not write it. |

Do not edit other files. When a gap needs a change in another file, stop and
report it.

## Requirements

- Write each test first and see it fail for the correct reason. Test at the
  model with `Update` and `View`. Write each expected text as a literal.
- Each frame stays exactly the size of the terminal at 120 x 40 and 200 x 57.
  Use the frame size test of 88-tui-terminal-io.
- Keep each test that passes today. Do not weaken an assertion.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each new test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Build the task binary. Copy a marked home to a temporary path. Add fixture data. Set `HIVE_HOME` to the copy. Set `XDG_CONFIG_HOME` to a temporary path. Run the binary in a herdr tab at about 200 x 57. Use `herdr pane read` to capture each view. | Each view matches the board. Put each capture in the evidence. |
| V3 | `git diff origin/dev --stat` | Only the files of the source table changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0. |

Never open the Chat input on the real home. Close each herdr tab and
workspace that you open. Do not close a pane that you did not open.

## Finish

Commit on the task branch with a subject such as `fix(tui): match the Tasks board`. Push the task
branch. Write the evidence. Run `hive task complete`. Send the commit SHA and
the V1-V4 results to the manager with `hive inbox send`. Read only your own
inbox.

## Repair 1

The verifier failed V2: the `v` view (`VERIFIED.md`) and the `d` view (the
dependency graph) break the frame. The panels collapse to about 8 rows, and
`Model.View()` gives 41 lines at 120 x 40 and 58 lines at 200 x 57. The `b`
view (the brief) has the same fault on `dev`. The shared file viewer is
`fileView` in `internal/tui/friction_screen.go`. See `VERIFIED.md`.

Scope change: you can also edit `internal/tui/friction_screen.go`, only for
the file viewer, and its tests.

Do these steps:

1. Merge `origin/dev` into the task branch. In `testdata/ste-baseline.json`,
   keep each row of both sides.
2. This edit copies the manager brief over your brief copy again. Apply your
   STE rewording to the copy again, and keep its meaning.
3. Add a frame test first: open the `b`, `v`, and `d` views on Tasks, and
   the file view on Friction, at 120 x 40 and 200 x 57. Each frame has
   exactly the height in lines and the width in cells. See it fail.
4. Fix the file viewer and the views so that the test passes and the panels
   use the full height.
5. Run V2 again on a scratch home in a tab of your own workspace, started
   with `env -u HERDR_ENV -u HERDR_PANE_ID -u HERDR_WORKSPACE_ID -u HERDR_SOCKET_PATH`.
   Stop and close only the pid and pane that you started. Never work in
   workspace `wBC`.
6. Run the gate. Commit and push. Make sure that `git merge-tree
   --write-tree origin/dev HEAD` exits 0. Run `hive task complete`. Send the
   commit SHA to the manager with `hive inbox send`.
