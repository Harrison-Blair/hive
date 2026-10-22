# Copy of the selected card boards

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, wave 6 (Screens) |
| Tier | small |
| Depends on | 10-board-export (landed). Base: `dev` |
| Blocks | 33-screen-agents |
| Grant | `grants.md 2026-10-21 tui-v2 delivery on dev and PR 47 update` |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

On 2026-10-21 the user chose a heavy `accent` border for the selected agent
card. The coordinator painted it on the boards `Main`, `AgentsMany`, and
`CompactAgents` (artifact version 40). This task copies the three boards and
exports their golden grids again. Task `33-screen-agents` then merges it.

When the board or the grid does not agree with this brief, stop and report it
with `hive inbox send --to manager-hive TEXT --agent NAME`.

## Design

- `/home/penguin/source/hive/plans/tui-v2/design/Main.dc.html`
- `/home/penguin/source/hive/plans/tui-v2/design/AgentsMany.dc.html`
- `/home/penguin/source/hive/plans/tui-v2/design/CompactAgents.dc.html`

This task changes no screen. The visual check captures the board copy from
the spec folder and compares it with a capture of the approved board.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/tui-v2/design/Main.dc.html`, `AgentsMany.dc.html`, `CompactAgents.dc.html` | AC1. Replace them. |
| `.hive/spec/tui-v2/design/Main.grid`, `AgentsMany.grid`, `CompactAgents.grid` | AC2. Export them again. |
| `.hive/spec/tui-v2/README.md` | AC3. Only the symbol list of decision 13. |
| `internal/boardexport/check.go` and its tests | AC3. Only the symbol list. |
| `internal/tui/component/symbols_test.go` | AC3. Only the symbol list. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/tui-v2/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files. Do not edit the exporter except its symbol list.
Do not add a mask.

## Requirements

### AC1 The board

- Copy the three boards from `/home/penguin/source/hive/plans/tui-v2/design/`
  over the copies in the worktree. `cmp` gives no difference.

### AC2 The grid

- Run the exporter of task `10-board-export` on the design folder. Commit
  only the changed grids. Each other grid stays the same.
- Each of the three grids changes in exactly 52 cells: the border of the
  selected card and one joint. The report
  `/home/penguin/source/hive/.hive/tmp/reports/tui-v2-selected-card.md` gives
  the cells.
- The gate stays green on the task branch. A test that reads one of these
  grids on `dev` can fail. Then stop and report it. Do not change a test.

### AC3 The symbol list (decision 13)

- The heavy border of the user decision needs new symbols. Add to decision
  13, to the list in `check.go`, and to `symbols_test.go` each symbol that the
  three boards use and that the list does not have. Also add each heavy
  corner and each mixed joint that `component.Card` of task
  `33-screen-agents` can draw. Read that code on the branch
  `task/33-screen-agents`. Do not merge that branch.
- The list in the spec, in `check.go`, and in the test is the same. The
  width test gives 1 for each new symbol.
- First failing test: the width test, or a check test, with a new symbol.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Each board copy equals the fixed board | V1 |
| AC2 | Only the 52 cells of each grid change | V2 |
| AC3 | The three symbol lists agree and each new symbol takes one cell | V-gate |

## Verification items

### V1 Board

- Proves: AC1.
- Command: `cmp` of each plans board against its copy. Capture both with the
  Firefox command of the verifier skill. Firefox is a snap. Put the profile
  folder and the captures in `.hive/tmp/` of the worktree. Remove them after
  the check.
- Expected: no difference. No capture shows `{{`.

### V2 Grid diff

- Proves: AC2.
- Command: compare each grid on `dev` with the task grid, cell by cell, with
  `internal/lib/cellgrid`. List each changed cell.
- Expected: only the rows of AC2 change. Each other grid file has no
  difference.

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

The components. Each other board.

## Changes

- 2026-10-21: AC3 adds the heavy border symbols to decision 13.
