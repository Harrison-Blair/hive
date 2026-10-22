# Copy of the fixed WorkGoals board

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, wave 6 (Screens) |
| Tier | small |
| Depends on | 10-board-export (landed). Base: `dev` |
| Blocks | 35-screen-work-goals |
| Grant | `grants.md 2026-10-21 tui-v2 delivery on dev and PR 47 update` |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

On 2026-10-21 the user chose to remove the label text `g2-tui-v2 is example data`
from the board `WorkGoals` (artifact version 33). This task copies the board
and exports its golden grid again. Task `35-screen-work-goals` then merges it.

When the board or the grid does not agree with this brief, stop and report it
with `hive inbox send --to manager-hive TEXT --agent NAME`.

## Design

- `/home/penguin/source/hive/plans/tui-v2/design/WorkGoals.dc.html`

This task changes no screen. The visual check captures the board copy from
the spec folder and compares it with a capture of the approved board.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/tui-v2/design/WorkGoals.dc.html` | AC1. Replace it. |
| `.hive/spec/tui-v2/design/WorkGoals.grid` | AC2. Export it again. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/tui-v2/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files. Do not edit the exporter. Do not add a mask.

## Requirements

### AC1 The board

- Copy the board from `/home/penguin/source/hive/plans/tui-v2/design/` over
  the copy in the worktree. `cmp` gives no difference.

### AC2 The grid

- Run the exporter of task `10-board-export` on the design folder. Commit
  only the changed `WorkGoals.grid`. Each other grid stays the same.
- Only row 9 changes: the cells of the label text become spaces with the token
  `fg`, not bold. No other cell changes.
- The fixture `WorkGoals.fixture.json` still loads. Change it only if it
  holds the label text.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The board copy equals the fixed board | V1 |
| AC2 | Only the cells of the label text change | V2 |

## Verification items

### V1 Board

- Proves: AC1.
- Command: `cmp` of the plans board against its copy. Capture both with the
  Firefox command of the verifier skill. Firefox is a snap. Put the profile
  folder and the captures in `.hive/tmp/` of the worktree. Remove them after
  the check.
- Expected: no difference. No capture shows `{{`.

### V2 Grid diff

- Proves: AC2.
- Command: compare `WorkGoals.grid` on `dev` with the task grid, cell by
  cell, with `internal/lib/cellgrid`. List each changed cell.
- Expected: only the label text cells of row 9 change. Each other grid file has no
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
