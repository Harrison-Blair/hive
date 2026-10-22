# Copy of board slip batch 1

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, wave 6 (Screens) |
| Tier | small |
| Depends on | 10-board-export (landed). Base: `dev` |
| Blocks | 37-screen-work-tasks, 38-screen-friction |
| Grant | `grants.md 2026-10-21 tui-v2 delivery on dev and PR 47 update` |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The coordinator fixed two board slips (artifact version 39). On the boards
`Work` and `FrictionResolved`, each row that is not selected is cut to the
column width less 1 and keeps a gap of 1 cell, as `component.List` draws it.
The coordinator also changed the code of the boards `Friction` and
`CompactFriction` with the same rule. Their cells do not change. This task
copies the four boards and exports the golden grids again. Tasks
`37-screen-work-tasks` and `38-screen-friction` then merge it.

When the board or the grid does not agree with this brief, stop and report it
with `hive inbox send --to manager-hive TEXT --agent NAME`.

## Design

- `/home/penguin/source/hive/plans/tui-v2/design/Work.dc.html`
- `/home/penguin/source/hive/plans/tui-v2/design/FrictionResolved.dc.html`
- `/home/penguin/source/hive/plans/tui-v2/design/Friction.dc.html`
- `/home/penguin/source/hive/plans/tui-v2/design/CompactFriction.dc.html`

This task changes no screen. The visual check captures the board copy from
the spec folder and compares it with a capture of the approved board.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/tui-v2/design/Work.dc.html`, `FrictionResolved.dc.html`, `Friction.dc.html`, `CompactFriction.dc.html` | AC1. Replace them. |
| `.hive/spec/tui-v2/design/Work.grid`, `FrictionResolved.grid` | AC2. Export them again. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/tui-v2/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files. Do not edit the exporter. Do not add a mask.

## Requirements

### AC1 The board

- Copy the four boards from `/home/penguin/source/hive/plans/tui-v2/design/`
  over the copies in the worktree. `cmp` gives no difference.

### AC2 The grid

- Run the exporter of task `10-board-export` on the design folder. Commit
  only the changed grids. Each other grid stays the same.
- `Work.grid`: only row 8 changes, near columns 28-29.
- `FrictionResolved.grid`: only rows 16, 18, 21, 24, and 25 change, near
  columns 36-37.
- The gate stays green on the task branch. A test that reads one of these
  grids on `dev` can fail. Then stop and report it. Do not change a test.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Each board copy equals the fixed board | V1 |
| AC2 | Only the cut cells change | V2 |

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

None.
