# Friction detail panel

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
and the board `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Friction.dc.html`. The feature README is
`.hive/spec/hive-tui/README.md`. Its decisions and G1-G14 win over the
board. G12: the TUI has no action except chat, attach, peek, and quit, and
read-only views. When the code does not agree with this brief, stop and report
it with `hive inbox send --to manager-hive TEXT --agent NAME`.

## Gaps to fix

- F2: The details of the selected entry show in a panel on the right of the list, as on the board.
- Do not fix F1. Keep the `decision` column. Friction entries have no goal field.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/friction_screen.go` and their tests | The gaps. |
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
| V2 | Build the task binary. Make a temporary copy of a marked home with fixture data. Set `HIVE_HOME` to this home. Set `XDG_CONFIG_HOME` to a temporary directory. Run the binary in a herdr tab at about 200 x 57. Capture the screen for each gap with `herdr pane read`. | The screen matches the board for each gap. Save each capture in the evidence directory. |
| V3 | `git diff origin/dev --stat` | Only the files of the source table changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0. |

Never open the Chat input on the real home. Close each herdr tab and
workspace that you open. Do not close a pane that you did not open.

## Finish

Commit on the task branch with a subject such as `fix(tui): match the Friction board`. Push the task
branch. Write the evidence. Run `hive task complete`. Send the commit SHA and
the V1-V4 results to the manager with `hive inbox send`. Read only your own
inbox.
## Repair 1

`hive land` failed: the branch has a merge conflict with `origin/dev` in
`testdata/ste-baseline.json`. Other tasks added rows for their brief copies.
Do only this repair:

1. Merge `origin/dev` into the task branch.
2. In `testdata/ste-baseline.json`, keep each row of both sides. Keep the
   order and format of the file.
3. Run the gate of `AGENTS.md`. It must pass.
4. Commit the merge. Push the task branch.
5. Make sure that `git merge-tree --write-tree origin/dev HEAD` exits 0.
6. Run `hive task complete`. Send the commit SHA to the manager with
   `hive inbox send`.

Do not change other files. Do not run the TUI.
