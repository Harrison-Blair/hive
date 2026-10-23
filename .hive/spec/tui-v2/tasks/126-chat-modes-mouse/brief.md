# Mouse on all tabs, and normal and input mode in Chat

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, goal `g6-g6-hive-ask`, wave A, task 1 of 2 |
| Tier | strong |
| Depends on | none. Base: `dev` at f7990406 or later |
| Grant | goal `g6-g6-hive-ask`: the user approved `plans/hive-ask.md` and the boards (canvas version 88). Build and land on `dev`. |
| Evidence | `.hive/spec/tui-v2/tasks/126-chat-modes-mouse/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The user approved the plan `/home/penguin/source/hive/plans/hive-ask.md` on
2026-10-22. Read it first. This task does decisions 8 and 9 of the plan and
the header count of the board details. Task 2 does `hive ask` and the question
card. Do not do task 2 work.

1. Mouse on all tabs. The wheel scrolls the pane under the pointer. A click
   selects or gives the focus. Copy needs shift and drag.
2. Chat has a normal mode and an input mode. `i`, or a click on the input,
   starts input mode. `ctrl+]`, or a click on the feed, stops input mode.
   `esc` and `tab` do not change the mode. `enter` sends and keeps input mode.
   In normal mode, letters stay shortcuts.
3. The header shows `? N asks` in `accent` before `needs you`, on all tabs,
   also when N is 0. The compact header shows `? N`. In this task N is always
   0. Task 2 gives the count. Use the singular `ask` when N is 1, as the
   `ChatAsk` board shows.

The TUI must match each board cell for cell: the glyph, the colour token, and
the bold flag.

Rule slips: when a board breaks a spec rule that exists and adds no new
design, follow the rule, do not stop, and list each case (board, row,
columns, rule) in `evidence/board-rule-cases.md` and in your report. Stop
and report only when no rule applies.

## Design

The approved boards (canvas "Hive TUI Redesign",
https://claude.ai/artifact/VrhcA98g9G2G9cFQao5Z9h), in
`/home/penguin/source/hive/.hive/tmp/g6/boards/`:

- `ChatNormal.dc.html` (120×40): normal mode. The feed has the focus.
- `ChatInput.dc.html` (120×40): input mode. The input has the focus.
- `CompactChatNormal.dc.html` (80×24): normal mode at 80 columns.

Each board has a `NAME.txt` text copy beside it. The `.dc.html` file is the
authority. The staged folder has no `support.js`. Use the copy in
`.hive/spec/tui-v2/design/`. The other boards in the folder are for task 2.

Follow the visual check rule in the worker skill. The user is colourblind.
Each state has a glyph and a word. Colour only adds to a state. The mode tag
`[normal]` or `[input]` and the focused border give the mode.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/tui-v2/design/<Board>.dc.html`, `.grid`, `.fixture.json` for the 3 boards | AC1. New copies, grids, and fixtures. |
| `internal/boardexport/check.go` and its tests | AC1. Add the 3 boards to `Sizes`. |
| `.hive/spec/tui-v2/design/*.grid` of the boards that AC6 changes | AC6 only. |
| `.hive/spec/tui-v2/README.md` | AC2. Decisions 79 and later. |
| `internal/tui/*.go`, `internal/tui/component/*.go`, and their tests | AC3 to AC6. |
| `testdata/ste-baseline.json` | To add the brief copy with 0, and to lower a count. |

Do not edit other files. Do not add a mask. Do not change the exporter
except `check.go`.

## Requirements

### AC1 Boards and grids

- Copy the 3 boards into `.hive/spec/tui-v2/design/`. `cmp` against the
  staged copies gives no difference.
- Add each board to `Sizes`. Add each new symbol to decision 13 and the
  symbol list, if a board uses one.
- Run the exporter. Commit the 3 new grids. Write one demo fixture for each
  board.

### AC2 Spec

Add the decisions of this task to the tui-v2 README in ASD-STE100 Strict,
from 79. Keep the meaning of the plan and of this brief.

### AC3 Mouse

- Turn on mouse reports in the Bubble Tea program. Use the mode that gives
  wheel and click events. Shift and drag must still select text in the
  terminal for copy.
- The wheel scrolls the pane or the list under the pointer, on each tab that
  has a scroll region. A wheel event over a region that cannot scroll does
  nothing.
- A click on a row of a list selects that row. A click on a pane gives it the
  focus. A click on a tab name opens that tab.
- Log each mouse action that changes the state, as each other user action.

### AC4 Chat modes

- Normal mode is the default. The footer starts with `[normal]`. The feed
  border is `accent`. The input border is `border` and shows
  `i or click here to type` in `dim`.
- In normal mode: `i` starts input mode. `↑` `↓` `j` `k` and the wheel scroll
  the feed. `pgup` `pgdn` page. `g` and `G` go to the top and the end.
  `ctrl+o` and `?` keep their functions. Each other letter keeps its shortcut.
- Input mode: the footer starts with `[input]`. The input border is
  `accent`, with a bold title. The feed border is `border`. The input shows
  `> ` and the cursor.
- In input mode: `enter` sends and keeps input mode. `shift+enter` adds a
  line. `ctrl+]` stops input mode. The wheel and `pgup` `pgdn` scroll the
  feed. `esc` and `tab` do not change the mode.
- A click on the input starts input mode. A click on the feed stops it.
- The feed title has the `↑ N earlier` note when lines are above the view, as
  the boards show. The right border shows the `┃` thumb.

### AC5 Header count

- The header shows `? N asks` before `needs you` on each tab, as the boards
  show. In this task N is 0.
- Put the count in the header state, so that task 2 can set it.

### AC6 Boards that change

The header count changes row 0 of each board that exists. The modes change
the boards of Chat that exist (for example `Chat`, `CompactChat`,
`ChatEvents`, `ChatStatus`, `CompactChatStatus`) and can change `Help`.

- Render the new grid of each board that changes from the code at its board
  size. Write them to `evidence/new-grids/` in the grid format of the
  exporter. Commit and push them.
- Write `evidence/new-grids/README.md`. For each board, give the rows that
  change. Prove with `internal/lib/cellgrid` that a board outside Chat and
  Help changes only in row 0.
- Send the path to the manager. Do not change the golden files yet. Hold until
  the manager sends the approval. Then copy each approved board over its
  `.dc.html` file, run the exporter, and check that each grid equals its
  grid from `new-grids`. Commit the grids. In the spec, record which grids
  the approved grids replace.
- Do the other work while you hold.

### AC7 Tests

- Write each test first and see it fail for the correct reason.
- The golden check passes for each of the 3 new boards and for each board of
  AC6 after the approval.
- Model tests with literal expected text and tokens: the footer tag and the
  focused border in each mode; `i`, `ctrl+]`, a click on the input, and a
  click on the feed change the mode; `esc` and `tab` do not change it;
  `enter` sends and keeps input mode; a letter in normal mode runs its
  shortcut and does not go into the input; the wheel over the feed and over a
  list on another tab scrolls that region; a click on a list row selects it;
  a click on a tab name opens the tab; the header count at 0 and 1 at both
  widths.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Board copies, grids, fixtures | V1, V2 |
| AC2 | The spec has the new decisions | V-gate |
| AC3 to AC5 | Mouse, modes, and header match the boards and the plan | V3, V4, V5 |
| AC6 | The changed boards have approved new grids | V2, V4 |
| AC7 | The tests fail on the base and pass on the task branch | V3 |

## Verification items

### V1 Boards

- Proves: AC1.
- Command: `cmp` of each staged board against its copy. Capture both with
  the Firefox command of the verifier skill. Put the profile folder and the
  captures in `.hive/tmp/` of the worktree. Remove them after the check.
- Expected: no difference. No capture shows `{{`.

### V2 Grids

- Proves: AC1, AC6.
- Command: run the exporter on the design folder. Compare each grid with the
  committed grid, cell by cell, with `internal/lib/cellgrid`.
- Expected: the 3 new grids and the approved grids are committed. No other
  grid changes.

### V3 Tests red and green

- Proves: AC3 to AC7.
- Command: run the new tests on the base code, then on the task branch. Run
  each new test with `-count=50` and with `-race`.
- Expected: each fails on the base and passes on the task branch in each run.

### V4 Live check

- Proves: AC3 to AC6.
- Command: the live check of decision 43 for each new board and each board of
  AC6 at its board size, with `COLORTERM=truecolor hive --demo FILE --size
  WxH`. Then in the Chat demo, press `i`, type text, press `enter`, press
  `esc` and `tab`, press `ctrl+]`, and press `j` and `k`. Read the pane after
  each key. Send wheel and click events with the raw SGR mouse sequences
  (herdr `send-keys` has no mouse names): the wheel over the feed and over a
  list on Agents and Logs, a click on the input, on the feed, on a list row,
  and on a tab name. Read the pane after each event.
- Expected: each capture equals its grid. Each key and each event gives the
  result of AC3 and AC4.

### V5 Visual alignment

- Proves: AC3 to AC5.
- Command: run the real TUI in a herdr pane at 120×40 and at 80×24, in each
  mode. Compare the capture at the board size with the board grid, cell by
  cell. Look at each capture and check each point:
  - Each panel border and corner lines up. The focused border is on the
    correct pane.
  - The header count, the feed note, and the thumb are at the board columns.
  - The input placeholder and the cursor are in the input.
- Expected: each capture at the board size equals its grid. Each point
  passes at both sizes. Put each capture, each diff, and a table of the points
  with pass or fail in the evidence. A mismatch is a fail.
- The worker and the verifier each do this item.

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
- Never open the Chat input on the real home. Never send chat. Use a demo or
  a scratch home for each Chat check.
- Close only the panes and the processes that you started. A demo quits
  with two `ctrl+c` presses.
- When a pane is smaller than the board size, use a scratch herdr session of
  your own at a larger size. Stop it after the check.
- Remove each scratch file that you make in `.hive/tmp/` after the check.

## Out of scope

`hive ask`, the question card and dock, and the skill edits (task 2). The
canvas.

## Finish

Commit on the task branch with a subject such as `feat(tui): add mouse
support and the chat modes`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1 to V5 results to the
manager with `hive inbox send --to manager-hive`.
