# Copy of the wave 6 board redraw

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, wave 6 (Screens) |
| Tier | small |
| Depends on | 10-board-export (landed). Base: `dev` |
| Blocks | 33-screen-agents to 40-screen-logs |
| Grant | `grants.md 2026-10-21 tui-v2 delivery on dev and PR 47 update` |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

On 2026-10-21 the user gave the board decisions of wave 6. The coordinator
redrew the boards from the target cells of the screen tasks. This task copies
the boards and exports their golden grids again. The screen tasks then merge it.

When a board, a fixture, or a grid does not agree with this brief, stop and
report it with `hive inbox send --to manager-hive TEXT --agent NAME`.

## Design

- Each `.dc.html` board in `/home/penguin/source/hive/plans/tui-v2/design/`.
- `/home/penguin/source/hive/.hive/tmp/reports/tui-v2-wave-f-redraw.md` gives
  the decisions and the target cells.

This task changes no screen file. The visual check captures each board copy from
the spec folder and compares it with a capture of the plans board.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/tui-v2/design/*.dc.html` | AC1. Replace each board that differs. |
| `.hive/spec/tui-v2/design/*.grid` | AC2. Export them again. |
| `internal/tui/component/list.go`, `details.go`, and their tests | AC4. Only the shared rules of the redraw. |
| `internal/tui/frame.go` and `frame_test.go` | AC4. Only the `/ filter` hints of Work and Friction. |
| `internal/tui/component/popup_test.go`, `chat_test.go` | AC4. Only the sample input that the redrawn boards change. |
| `internal/tui/component/chat.go` | AC4. Only when the Chat sample input cannot give the redrawn rows. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/tui-v2/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files. Do not edit the exporter. Do not add a mask. Do not
copy `canvas.json` or `support.js` unless they differ only by the redraw.
Report each other difference to the manager.

## Requirements

### AC1 The boards

- Copy each `.dc.html` board that differs from the plans folder over the copy
  in the worktree. For each board, `cmp` gives no difference.

### AC2 The grids

- Run the exporter of task `10-board-export` on the design folder. Commit
  each changed grid.
- A grid changes only when its board changes. Each changed cell agrees with
  the target cells of the report.
- After the export, a golden test can fail. AC4 gives the repair.

### AC3 The fixtures

- Change no fixture. The `Logs` row text `verify fail · V4 diff scope` comes
  from the `reason` field of the event. Task `40-screen-logs` makes that
  rule.

### AC4 The gate stays green

The grids change before the screens, so this task makes the shared code
follow the shared rules of the redraw. Change only these rules:

- `component.List`: the selected name is `accent` bold over the full column.
  The space after `▶` is `fg`, not bold. Column 2 of each other row is `fg`,
  not bold.
- `component.Details`: it shows each line that fits. It shows
  `↓ N more lines` only when the lines overflow.
- `frame.go`: remove `/ filter` from the hints of Work and Friction. Logs
  keeps it.
- The sample input of the component golden tests follows the redrawn boards.
  `TestPopupGolden` has one blank line before `RECENT`. `TestCompactChatGolden`
  has no hard line break in the reply that the board wraps. For
  `TestChatGolden`, first change the sample input. Change `chat.go` only
  when the input cannot give the rows, and record the reason.

Write a failing test for each rule first, or show that a current golden test
fails for the correct reason. Task `38-screen-friction` has commits for the
List and Details rules on its branch. You can read them. Do not merge that
branch.

When a test fails for a cause that only a screen task can repair, stop and
report the test and the cells to the manager. Do not change a screen file.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Each board copy equals the plans board | V1 |
| AC2 | Only the grids of changed boards change, as the report gives | V2 |
| AC3 | No fixture changes | V-gate |
| AC4 | The shared rules hold and the gate is green | V3, V-gate |

## Verification items

### V1 Boards

- Proves: AC1.
- Command: `cmp` of each plans board against its copy. Capture three changed
  boards and their plans copies with the Firefox command of the verifier
  skill: `Main`, `Logs`, and `Components`. Firefox is a snap. Put the profile
  folder and the captures in `.hive/tmp/` of the worktree. Remove them after
  the check.
- Expected: no difference. No capture shows `{{`.

### V2 Grid diff

- Proves: AC2.
- Command: compare each grid on `dev` with the task grid, cell by cell, with
  `internal/lib/cellgrid`. List the changed cells of each grid. Compare
  samples of them with the target cells of the report. Run the export again
  and compare.
- Expected: only the grids of changed boards change. The samples agree with
  the report. A second export gives no difference.

### V3 Shared rules

- Proves: AC4.
- Command: run each new test against `dev`, then against the task branch.
  Then break each of the three rules on purpose and run the tests again.
- Expected: each new test fails on `dev` and passes on the task branch. Each
  deliberate break fails a test.

### V-gate Diff scope and gate

- Proves: each criterion.
- Command: `git diff dev...HEAD --stat`, after you merge `dev` into the task
  branch. Then the full gate of `AGENTS.md` from the worktree root.
- Expected: only the files of the source table and the evidence changed.
  `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0.
  Put the full gate output in the evidence.

## Rules

- Do not run bare `hive` in herdr without `--no-startup` or
  `HIVE_NO_STARTUP=1`.

## Out of scope

Each screen file. Each screen golden test of wave 6.

## Changes

- 2026-10-21: AC4 adds the sample input of three component golden tests.
- 2026-10-21: AC3 changes no fixture. Task `40-screen-logs` makes the row
  text from the `reason` field.
