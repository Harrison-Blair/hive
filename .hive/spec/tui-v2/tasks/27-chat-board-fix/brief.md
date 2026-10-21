# Copy of the redrawn Chat boards

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, wave 4 (Components, part 2) |
| Tier | small |
| Depends on | 10-board-export (landed). Base: `dev` |
| Blocks | 23-components-2 |
| Grant | `grants.md 2026-10-21 tui-v2 delivery on dev and PR 47 update` |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

On 2026-10-21 the user chose to draw tables as Glamour v2.0.1 draws them. The
coordinator redrew the boards `Chat` and `CompactChat` (artifact version 29).
This task copies the two boards and exports their golden grids again. Task
`23-components-2` then merges it.

When the board or the grid does not agree with this brief, stop and report it
with `hive inbox send --to manager-hive TEXT --agent NAME`.

## Design

- `/home/penguin/source/hive/plans/tui-v2/design/Chat.dc.html`
- `/home/penguin/source/hive/plans/tui-v2/design/CompactChat.dc.html`
- The Glamour table report:
  `/home/penguin/source/hive/.hive/tmp/reports/tui-v2-glamour-tables.md`.

This task changes no screen. The visual check captures each board copy from
the spec folder and compares it with a capture of the approved board.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/tui-v2/design/Chat.dc.html`, `CompactChat.dc.html` | AC1. Replace them. |
| `.hive/spec/tui-v2/design/Chat.grid`, `CompactChat.grid` | AC2. Export them again. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/tui-v2/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files. Do not edit the exporter. No mask file exists for
these boards. Do not add one.

## Requirements

### AC1 The boards

- Copy the two boards from `/home/penguin/source/hive/plans/tui-v2/design/`
  over the copies in the worktree. `cmp` gives no difference.

### AC2 The grids

- Run the exporter of task `10-board-export` on the design folder. Commit
  only the changed `Chat.grid` and `CompactChat.grid`. Each other grid stays
  the same.
- In `Chat.grid`, only the cells of the table change. The table starts at
  panel column 23 with the column widths 19, 35, 18, and 19. The header row
  uses `fg` and is not bold. Row 21, column 37, the space after
  `Recommended:`, uses `fg` and is not bold.
- In `CompactChat.grid`, only the cells of the table and of rows 17 and 18
  change. The table has the column widths 9, 34, 17, and 13, and the full
  text of row A. Row 17 keeps `14:24 you`. The user text `1A, 2A, 3B` is on
  row 18 at column 2.
- Each table cell agrees with the Glamour report: section 1 for `Chat`,
  section 3 for `CompactChat`.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Each board copy equals the redrawn board | V1 |
| AC2 | Only the named cells of the two grids change, and they agree with the report | V2 |

## Verification items

### V1 Boards

- Proves: AC1.
- Command: `cmp` of each plans board against its copy. Capture each copy and
  each plans board with the Firefox command of the verifier skill. Firefox is
  a snap. Put the profile folder and the captures in `.hive/tmp/` of the
  worktree. Remove them after the check.
- Expected: no difference. No capture shows `{{`.

### V2 Grid diff

- Proves: AC2.
- Command: compare each grid on `dev` with the task grid, cell by cell, with
  `internal/lib/cellgrid`. List each changed cell. Compare each table row
  with the report.
- Expected: only the cells of AC2 change, and each table cell agrees with the
  report. Each other grid file has no difference.

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

- 2026-10-21: the `Chat` board is artifact version 29. One more cell
  changes: row 21, column 37.
