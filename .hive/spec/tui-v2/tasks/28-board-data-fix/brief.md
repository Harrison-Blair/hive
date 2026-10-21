# Board data fixes and re-export

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, wave 4 (Components, part 2) |
| Tier | small |
| Depends on | 12-board-fixtures, 23-components-2, 27-chat-board-fix (landed). Base: `dev` |
| Blocks | wave 5 (Frame and layout) |
| Grant | `grants.md 2026-10-21 tui-v2 delivery on dev and PR 47 update` |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

On 2026-10-21 the user chose to fix five data inconsistencies in the boards,
which task `12-board-fixtures` recorded. The coordinator fixed the boards
(artifact version 30). This task copies the seven changed boards and makes
their grids and fixtures agree.

When a board, a grid, or a fixture does not agree with this brief, stop and
report it with `hive inbox send --to manager-hive TEXT --agent NAME`.

## Design

- `/home/penguin/source/hive/plans/tui-v2/design/`, the boards
  `AgentDetails`, `AgentsMany`, `Chat`, `CompactChat`, `CompactLogs`,
  `HealthAgents`, and `Logs`.

This task changes no screen. The visual check captures each changed board copy
from the spec folder and compares it with a capture of the approved board.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/tui-v2/design/<Board>.dc.html` | AC1. Only the seven changed boards. |
| `.hive/spec/tui-v2/design/<Board>.grid` | AC2. Only the seven changed boards. |
| `.hive/spec/tui-v2/design/<Board>.fixture.json` | AC3. |
| `internal/tui/component/chat_test.go`, `popup_test.go` | AC2. Only the sample time literals that the fixed boards changed. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/tui-v2/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files. Do not edit the exporter. No mask file exists. Do not
add one. When you find that a mask is necessary, stop and report it, because
the user reviews each mask change.

## Requirements

### AC1 The boards

- Copy the seven boards from `/home/penguin/source/hive/plans/tui-v2/design/`
  over the copies in the worktree. After the copy, each of the 21 boards in
  the spec folder is equal to its plans copy. `cmp` gives no difference.

### AC2 The grids

- Run the exporter of task `10-board-export` on the design folder. Commit
  only the changed grids of the seven boards. Each other grid stays the same.
- Only these cells change:
  1. Times. `Chat` and `CompactChat`: the user message time `14:24` becomes
     `14:22`. `Logs` and `CompactLogs`: each event time is one hour earlier,
     and the event popup and the JSON time also. `AgentDetails`: the RECENT
     times become `14:08`, `14:07`, `13:58`, and `13:44`.
  2. `HealthAgents`: the title says `h hides 1 stopped`.
  3. `Logs` and `CompactLogs`: the filter shows `proj: hive` in place of
     `agent: manager-hive`. `proj:` uses `dim`, and `hive` uses `fg`.
  4. `AgentsMany`: the goals are `g4-rate-limits · wave B` for
     `example-api` and `g2-guides · wave A` for `example-docs`.

### AC3 The fixtures

- Change each fixture of the seven boards to hold the data of its fixed
  board: the times, the stopped count, the filter, and the goals of AC2.
- In each fixture of the 21 boards, the coordinator pane is `w58:p4`, as each
  board shows. No fixture holds `w58:p3`.
- Record the source line of each changed value in the evidence.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Each board copy equals its plans copy | V1 |
| AC2 | Only the named cells of the seven grids change | V2 |
| AC3 | Each fixture agrees with its board | V3 |

## Verification items

### V1 Boards

- Proves: AC1.
- Command: `cmp` of each plans board against its copy. Capture each of the
  seven copies and its plans board with the Firefox command of the verifier
  skill. Firefox is a snap. Put the profile folder and the captures in
  `.hive/tmp/` of the worktree. Remove them after the check.
- Expected: no difference. No capture shows `{{`.

### V2 Grid diff

- Proves: AC2.
- Command: compare each grid on `dev` with the task grid, cell by cell, with
  `internal/lib/cellgrid`. List each changed cell.
- Expected: only the cells of AC2 change. The 14 other grid files have no
  difference.

### V3 Fixtures

- Proves: AC3.
- Command: run the fixture load test of task `12-board-fixtures`. Search each
  fixture for `w58:p3`. For three of the seven boards that the verifier picks,
  compare each changed fixture value with the board script.
- Expected: each fixture loads. No fixture holds `w58:p3`. Each compared value
  agrees with its board.

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

The components. The frame. Each screen.

## Changes

- 2026-10-21: the source table allows the sample times in two component
  tests, because their golden checks read the fixed grids.
