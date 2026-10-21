# Copy of the fixed Components board

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, wave 3 (Components, part 1) |
| Tier | small |
| Depends on | 10-board-export (landed). Base: `dev` |
| Blocks | 21-components-1 |
| Grant | `grants.md 2026-10-21 tui-v2 delivery on dev and PR 47 update` |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

On 2026-10-21 the user fixed two places in the Components board (artifact
version 25). The sideways scroll mark at row 30 uses `dim`. The SubTabs gaps
at row 14, columns 8, 9, 17, and 18, use `fg`. This task copies the fixed
board and exports its golden grid again. Task `21-components-1` then merges
it.

When the board or the grid does not agree with this brief, stop and report it
with `hive inbox send --to manager-hive TEXT --agent NAME`.

## Design

- `/home/penguin/source/hive/plans/tui-v2/design/Components.dc.html`, the
  fixed board, artifact version 25.
- `/home/penguin/source/hive/plans/tui-v2/design/support.js`, the script that
  each board loads with `./support.js`. The coordinator wrote it as a small
  shim. It runs the board script and fills the row template. Do not copy the
  runtime of the artifact type.

This task changes no screen. The visual check captures the board copy from the
spec folder and compares it with a capture of the approved board.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/tui-v2/design/Components.dc.html` | AC1. Replace it. |
| `.hive/spec/tui-v2/design/Components.grid` | AC2. Export it again. |
| `.hive/spec/tui-v2/design/support.js` | AC3. A new file. |
| `.gitignore` | AC3. Only if Git ignores the new file. Use the narrowest rule. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/tui-v2/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files. Do not edit the exporter.

## Requirements

### AC1 The board

- Copy `/home/penguin/source/hive/plans/tui-v2/design/Components.dc.html`
  over the copy in the worktree. `cmp` gives no difference.

### AC2 The grid

- Run the exporter of task `10-board-export` on the design folder. Commit
  only the changed `Components.grid`. Each other grid stays the same.
- Only these cells change: row 30, the cells of the sideways scroll mark,
  token `accent` to `dim`. Row 14, columns 8, 9, 17, and 18: token `dim` to
  `fg`. No rune and no bold flag changes.

### AC3 The board script

- Copy `/home/penguin/source/hive/plans/tui-v2/design/support.js` into
  `.hive/spec/tui-v2/design/`. `cmp` gives no difference.
- Each board of the spec folder renders in Firefox from a `file://` URL. No
  capture shows a literal `{{`.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The board copy equals the fixed board | V1 |
| AC2 | Only the named cells of `Components.grid` change | V2 |
| AC3 | Each board in the spec folder renders | V3 |

## Verification items

### V1 Board

- Proves: AC1.
- Command: `cmp` of the plans board against the copy.
- Expected: no difference.

### V2 Grid diff

- Proves: AC2.
- Command: compare `Components.grid` on `dev` with the task grid, cell by
  cell, with `internal/lib/cellgrid`. List each changed cell.
- Expected: only the cells of AC2 change. Each other grid file has no
  difference.

### V3 Board render

- Proves: AC3.
- Command: capture `Components.dc.html` and two other boards from the spec
  folder with the Firefox command of the verifier skill. Capture the approved
  `Components.dc.html` from the plans folder.
- Expected: no capture shows `{{`. The two captures of `Components.dc.html`
  agree.
- Firefox is a snap. It cannot read `/tmp`. Put the profile folder and the
  captures in `.hive/tmp/` of the worktree. Remove them after the check.

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

- 2026-10-21: after a verify fail, added the Design section, AC3, and V3.
  The boards load `./support.js`, which the spec folder did not have.
