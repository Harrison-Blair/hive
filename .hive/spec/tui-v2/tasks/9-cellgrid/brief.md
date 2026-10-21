# The cell grid and the ANSI converter

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, wave 2 (Harness) |
| Tier | decent |
| Depends on | 7-tui-v2-spec-copy. Base: `dev` after they land |
| Grant | `grants.md 2026-10-21 tui-v2 delivery on dev and PR 47 update` |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Decisions 39, 40, 42, and 45 need one model of a terminal cell grid. The
golden check, the live check, and the exporter all use it. This task makes
the grid, its file format, the ANSI converter, and the compare with masks.
It also records the manager decisions on the layout and the formats in the
spec.

## Manager decisions

- Package: `internal/lib/cellgrid/`. The golden check and the exporter call it.
- Grid file: `<Board>.grid` beside each board in `.hive/spec/tui-v2/design/`.
  It is UTF-8 text. Line 1 is `hive-grid 1 W H`. Then each row has three lines:
  `t|` and the W characters of the row, `k|` and W token letters, `b|` and W
  digits `0` or `1` for bold. Token letters: `f` fg, `r` border, `d` dim,
  `a` accent, `w` working, `o` done, `i` info, `u` user, `c` code.
- Mask file: `<Board>.mask` beside the grid. Each line is `ROW COL WIDTH`,
  zero-based, and an optional `# reason`. A blank line or a `#` line is a
  comment. A board with no mask has no mask file.
- Fixture file: `<Board>.fixture.json` beside the grid. Task `demo-mode`
  defines its schema.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/lib/cellgrid/` and its tests | AC1-AC4. A new package. |
| `.hive/spec/tui-v2/README.md` | AC5. Only the section `Manager decisions`. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/tui-v2/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### AC1 Grid and file format

- A grid has a width, a height, and one cell for each position: a rune, a
  token, and a bold flag. The palette has the nine tokens of decision 26.
- The package reads and writes the grid file format above. A write and a
  read give the same grid. A bad header, a wrong row length, an unknown
  token letter, or a wide rune fails with an error that names the line.

### AC2 Converter (decision 40)

- The converter turns ANSI text and a size into a grid. It maps each 24-bit
  SGR foreground to the token with the same hex. It maps the default
  foreground to `fg`. It reads bold on and off, and reset.
- It fails on an indexed color, a 16-color code, a background color, or a hex
  value outside the palette. The error names the row, the column, and the
  code.
- It measures each rune with `lipgloss.Width` semantics. A cell with no
  character is a space with token `fg`.
- First failing test: hand-written ANSI input for a two-row box with an
  accent title gives a hand-written expected grid. Do not build an expected
  grid with the converter.

### AC3 Compare with masks (decisions 42, 45)

- Compare gives each cell that differs: row, column, the expected and the
  actual rune, token, and bold. Its text output prints each one.
- Inside a mask, compare ignores the rune but still compares the token and
  the bold flag.
- A mask that touches a border rune, or a cell with token `accent` in the
  expected grid, fails the compare with an error. Record in the evidence how
  the check finds a title, a label, or a key hint, and each limit.

### AC4 Palette

- One table gives the hex of each token, as in decision 26. The fg value for
  a board render is `#e6e1d6`. The converter treats only the default
  foreground as `fg`.

### AC5 The spec

- Write the manager decisions above in the section `Manager decisions` of the
  spec, in ASD-STE100 Strict. Do not change a numbered decision.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The grid file reads and writes with clear errors | V1 |
| AC2 | The converter maps palette colors and fails on others | V1 |
| AC3 | Compare reports each differing cell and applies masks | V1 |
| AC4 | One palette table matches decision 26 | V1 |
| AC5 | The spec records the manager decisions | V2 |

## Verification items

### V1 Tests

- Proves: AC1, AC2, AC3, AC4.
- Command: run each new test against the base code, then against the task
  code. Then break the converter on purpose (map one hex to a wrong token)
  and run the tests again.
- Expected: each new test fails on the base and passes on the task branch.
  The broken converter fails a test.

### V2 Spec

- Proves: AC5.
- Command: `git diff dev...HEAD -- .hive/spec/tui-v2/README.md` and
  `go test ./ -run STE`.
- Expected: only the section `Manager decisions` changes. The test passes.

### V-gate Diff scope and gate

- Proves: each criterion.
- Command: `git diff dev...HEAD --stat`, after you merge `dev` into the task
  branch. Then the full gate of `AGENTS.md` from the worktree root.
- Expected: only the files of the source table and the evidence changed.
  `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0.
  Put the full gate output in the evidence.

## Rules

- Read `.hive/spec/tui-v2/README.md` first. When the code, a board, or the
  spec does not agree with this brief, stop and report it with
  `hive inbox send --to manager-hive TEXT --agent NAME`.
- Each Go change is test-first. The brief names the first failing test. See
  it fail for the correct reason.
- Do not run bare `hive` in herdr without `--no-startup` or
  `HIVE_NO_STARTUP=1`. Before task `no-startup` lands, do not run bare `hive`
  in herdr. Refer to the friction entry `scratch-hive-starts-real-agents`.
- A live check never changes the real home. Close each workspace, tab, and
  pane that the check makes.

## Out of scope

The exporter, the golden grids, and demo mode.

## Changes

None.
