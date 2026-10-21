# Components, part 1

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, wave 3 (Components, part 1) |
| Tier | strong |
| Depends on | 9-cellgrid, 10-board-export (landed). Base: `dev` |
| Grant | `grants.md 2026-10-21 tui-v2 delivery on dev and PR 47 update` |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Wave 3 of the spec makes the first reusable components: Header, TabBar,
SubTabs, Panel, Viewport with ScrollMark, and KeyHints, and the symbol width
test of decision 13. Each later screen uses them. The board
`Components.dc.html` shows each component, and its golden grid
`Components.grid` is the expected output.

## Manager decisions

- Package: `internal/tui/component/`. One file for each component. The
  package has a `doc.go`. Wave 4 adds the other components to the same
  package.
- Each component is a plain value with a `View` (or `Render`) method that
  takes a width and returns a string. It does no I/O. Choose the exact API
  and record it.

## Design

- `/home/penguin/source/hive/.hive/spec/tui-v2/design/Components.dc.html`,
  sections Palette, Header, TabBar, SubTabs, Panel, Viewport + ScrollMark,
  and KeyHints.
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/Components.grid`.
- The 80×24 boards for the short header and the tab bar below 120 columns
  (decisions 2, 3), for example `CompactAgents.grid`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/component/` and its tests | AC1-AC4. A new package. |
| `internal/tui/component/testdata/` | Only test inputs, if a test needs a file. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/tui-v2/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files. Do not wire a component into a screen. Wave 5 and
wave 6 do that.

## Requirements

### AC1 Components (decisions 1 to 6, 26, 27)

- Header: the long form from 120 columns and the short form below 120
  (decision 2).
- TabBar: the six labels, the active tab in `accent` and bold with the `━`
  underline, and the spacing of decision 3 at each width.
- SubTabs: the views of a screen, as on the board.
- Panel: `lipgloss.RoundedBorder()`, `Padding(0, 1)`, the title in the top
  border, `accent` border when focused and `border` when not (decisions 6,
  27).
- Viewport with ScrollMark: a region that scrolls, with `┃` on its right
  border and `━` on its bottom border when it scrolls sideways (decision 5).
  The mark is `accent` when focused and `dim` when not.
- KeyHints: `[region]` in `accent`, then the hints in `dim` (decision 4). It
  cuts the hints with `…` when they do not fit.
- Each component measures text with `lipgloss.Width`, never `len`
  (decision 12). Each uses only the palette tokens of decision 26.

### AC2 Golden checks of the components

- For each component, a test renders it with the sample input of its board
  section and compares the result with the same cell rectangle of
  `Components.grid`, with `internal/lib/cellgrid`. Use the TrueColor profile
  of `lipgloss`. The test prints each differing cell.
- Take each sample input from the board script. Record the rectangle and the
  source line of each sample in the evidence.
- First failing test: the Panel golden check, before Panel exists.

### AC3 Symbol width (decision 13)

- A test checks that `lipgloss.Width` gives 1 for each symbol of the list in
  decision 13.

### AC4 Fit

- Each component, at each width from its minimum to 200 columns, gives no
  line wider than the width and no broken border. Test it as a property.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Each component follows its decisions | V1, V2 |
| AC2 | Each component matches its rectangle of `Components.grid` | V1, V2 |
| AC3 | Each symbol of decision 13 has width 1 | V1 |
| AC4 | No component breaks a border or exceeds its width | V1 |

## Verification items

### V1 Tests

- Proves: AC1, AC2, AC3, AC4.
- Command: run each new test against the base code, then against the task
  code. Then change one token in one component on purpose and run the tests
  again.
- Expected: each new test fails on the base and passes on the task branch.
  The changed token fails a golden check that prints the cell.

### V2 Rectangles

- Proves: AC1, AC2.
- Command: for three components that the verifier picks, compare the
  recorded rectangle with the board script section.
- Expected: each rectangle covers the whole component and only it.

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
  `HIVE_NO_STARTUP=1`. Refer to the friction entry
  `scratch-hive-starts-real-agents`.
- A live check never changes the real home. Close each workspace, tab, and
  pane that the check makes.

## Out of scope

The wave 4 components. The frame. Each screen.

## Changes

None.
