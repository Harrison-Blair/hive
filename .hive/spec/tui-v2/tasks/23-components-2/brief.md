# Components, part 2

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, wave 4 (Components, part 2) |
| Tier | strong |
| Depends on | 21-components-1, 22-components-board-fix (landed). Base: `dev` |
| Grant | `grants.md 2026-10-21 tui-v2 delivery on dev and PR 47 update` |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Wave 4 of the spec makes the other components: List, Details, FilterBar,
Input with the Chat text, Card, Popup, Notice, and Markdown with the Glamour
style. They go in the package `internal/tui/component/` that wave 3 made. Use
the patterns of wave 3: one file for each component, a `View` that takes a
width, the palette of `palette.go`, the golden helper of `golden_test.go`, and
the fit property of `fit_test.go`.

## Design

- `/home/penguin/source/hive/.hive/spec/tui-v2/design/Components.dc.html`,
  sections List, FilterBar, Input + Chat text, Card, Popup, and Notice.
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/Components.grid`.
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/Chat.dc.html` and
  `Chat.grid`, for the Chat text and the Markdown style.
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/AgentDetails.dc.html`,
  for the Popup and the Card.
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/Work.dc.html`, for the
  List and the Details.

This task changes no screen. The visual check compares the rendered
components with the Components board.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/component/` and its tests | AC1-AC5. |
| `internal/tui/component/testdata/` | Only test inputs, if a test needs a file. |
| `go.mod`, `go.sum` | AC4. Only Glamour and the modules that it needs. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/tui-v2/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files. Do not wire a component into a screen. Wave 6 does
that.

## Requirements

### AC1 Components (decisions 5, 6, 8, 9, 10, 12, 13, 26, 27)

- List: rows with the selection marker, the selected row visible after each
  move, and the scroll mark of wave 3 when the rows do not fit.
- Details: the details panel of a list, as on the Work board.
- FilterBar: the typed fields of decision 25 and the free text, with the
  applied filter shown as on the board.
- Input: the fixed chat input of decision 15, as on the board.
- Card: the agent card of decision 10, 24×4 cells, with the idle dot `◐` in
  `accent`.
- Popup: a centered box over a dimmed screen (decision 5). Each cell of the
  screen behind it, including the header and the tab bar, uses `border`, as
  on the boards `AgentDetails` and `Help`. The Popup uses the Panel padding of
  1. The Details inside it adds one space. The golden check uses
  `AgentDetails.grid` rows 7-33, columns 30-89.
- Notice: a one-line notice. No board draws a sample. Its text uses `dim`.
  Test it with a hand-written expected grid.
- Each component measures text with `lipgloss.Width`, never `len`. Each uses
  only the palette tokens and the symbols of decision 13.

### AC2 Chat text (decisions 15, 30, 31)

- From 120 columns the feed has the gutter for the time and the name. Message
  text starts 23 cells from the left edge of the feed panel. The wrap width is
  the panel width minus 26 cells. Below 120 columns the time and the name use
  their own line.
- User text uses the token `user`. An event folds to one `dim` line.

### AC3 Golden checks

- For each component, a test renders it with the sample input of its board
  section and compares the result with its rectangle of `Components.grid`, or
  of `Chat.grid` for the Chat text. Record each rectangle and the source line
  of each sample in the evidence.
- First failing test: the List golden check, before List exists.

### AC4 Markdown (decisions 28, 29, 31, 32, 33)

- Add `charm.land/glamour/v2` at v2.0.1. The commit that adds it states this
  reason: "markdown rendering of agent chat replies, chosen by the owner
  2026-10-21".
- One Go file holds the complete Glamour style of decision 29.
- The component removes only the blank first line and the trailing spaces
  of each line that Glamour adds. It keeps leading spaces.
- Tables follow decision 32. Tests cover uneven content, columns of equal
  width, and a table with too many columns. The Chat golden check includes
  the table rows of `Chat.grid`.
- A test renders hand-written markdown with a heading, strong text, inline
  code, a code block, a list, and a link. It compares the tokens with a
  hand-written expected grid or with the Chat board.

### AC5 Fit

- Each component, at each width from its minimum to 200 columns, gives no
  line wider than the width and no broken border.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Each component follows its decisions | V1, V2 |
| AC2 | The Chat text follows decisions 15, 30, and 31 | V1, V2 |
| AC3 | Each component matches its board rectangle | V1, V2 |
| AC4 | Markdown uses Glamour with the palette style | V1 |
| AC5 | No component breaks a border or exceeds its width | V1 |

## Verification items

### V1 Tests

- Proves: AC1, AC2, AC3, AC4, AC5.
- Command: run each new test against the base code, then against the task
  code. Then change one token in one component on purpose and run the tests
  again.
- Expected: each new test fails on the base and passes on the task branch.
  The changed token fails a golden check that prints the cell.

### V2 Rectangles

- Proves: AC1, AC2, AC3.
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

The frame. The key routes. Each screen.

## Changes

- 2026-10-21: after the questions of t23-worker-1, the Popup uses `border`
  behind it and the AgentDetails rectangle, the Notice uses a hand-written
  grid, and AC4 includes decisions 32 and 33.
