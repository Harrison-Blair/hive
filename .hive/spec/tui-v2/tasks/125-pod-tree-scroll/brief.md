# Agents: the one-pod view is always a tree

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, goal `g5-live-agent-activity`, wave B |
| Tier | strong |
| Depends on | none. Base: `dev` at 022c7197 or later |
| Grant | goal `g5-live-agent-activity`: the user approved wave B on 2026-10-22. Build and land on `dev`. |
| Evidence | `.hive/spec/tui-v2/tasks/125-pod-tree-scroll/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The user approved four new boards on 2026-10-22 (canvas version 78, id
1792709136-2bec). In the one-pod view, the pod always draws a tree. The rail
is not used in the one-pod view. When the tree is wider or higher than the
region, the view scrolls to keep the selected card on screen.

This task copies the boards into the spec, exports their grids, adds the
decisions to the tui-v2 spec, and builds the one-pod view to match the
boards. The overview does not change.

The TUI must match each board cell for cell: the glyph, the colour token, and
the bold flag.

Rule slips: when a board breaks a spec rule that exists and adds no new
design (for example a thumb size or a text position), follow the rule, do
not stop, and list each case (board, row, columns, rule) in
`evidence/board-rule-cases.md` and in your report. The manager found one
case before the task: `CompactTreeWide` has no `│` row between the
coordinator card and the pod top border (rows 7 and 8). `CompactLivePod`
has this row. Check it against the rules of decision 65. Stop and report
only when no rule applies.

## Design

The approved boards, copied from the canvas "Hive TUI Redesign"
(https://claude.ai/artifact/VrhcA98g9G2G9cFQao5Z9h, version
1792709136-2bec), in `/home/penguin/source/hive/.hive/tmp/g5b/boards/`:

- `TreeWide.dc.html` (120×40): 5 children, wider than the region, first child
  selected, scrolled to the left edge.
- `TreeNested.dc.html` (120×40): a nested child under `t5-worker-1`,
  selected.
- `CompactTreeWide.dc.html` (80×24): 5 children, the middle child selected and
  centred, scrolled sideways and vertically.
- `CompactTreeNested.dc.html` (80×24): the nested child selected, scrolled
  vertically.

The staged folder has no `support.js`. Use the copy in
`.hive/spec/tui-v2/design/`.

Follow the visual check rule in the worker skill. The user is colourblind.
Each state has a glyph and a word. Colour only adds to a state.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/tui-v2/design/<Board>.dc.html`, `<Board>.grid`, `<Board>.fixture.json` for the 4 boards | AC1. New copies, grids, and fixtures. |
| `internal/boardexport/check.go` and its tests | AC1. Add the 4 boards to `Sizes`. |
| `.hive/spec/tui-v2/design/AgentsNested.*`, `AgentsCoordinatorChild.*` | AC6 only. |
| `.hive/spec/tui-v2/README.md` | AC2. Decisions 74 to 78. |
| `internal/tui/agents_screen.go`, `internal/tui/agents.go`, `internal/tui/frame.go`, `internal/tui/demo.go`, `internal/tui/component/*.go`, and their tests | AC3 to AC6. |
| `internal/tui/tui.go` | Only when AC3 to AC6 need a change there. Record the cause in the evidence. |
| `testdata/ste-baseline.json` | To add the brief copy with 0, and to lower a count. |

Do not edit other files. Do not add a mask. Do not change the exporter
except `check.go`. Except for the boards of AC6, each grid and golden check
that exists stays the same.

## Requirements

### AC1 Boards and grids

- Copy the 4 boards into `.hive/spec/tui-v2/design/`. `cmp` against the
  staged copies gives no difference.
- Add each board to `Sizes`: 120×40 for `TreeWide` and `TreeNested`, 80×24
  for the two `Compact` boards. Add each new symbol to decision 13 and the
  symbol list, if a board uses one.
- Run the exporter. Commit the 4 new grids. Write one demo fixture for each
  board.

### AC2 Spec

Add decisions 74 to 78 to the tui-v2 README in ASD-STE100 Strict. Keep the
meaning of this brief. Record that decision 74 replaces the rail in the
one-pod view (decision 63).

### AC3 Decision 74: always a tree

- The one-pod view always draws the tree. It does not use the rail.
- A nested child is one more tree level under its parent: the parent bottom
  joint `┬`, one `│` row, then the child card with `┴` (or `┷` on a selected
  card). Several children of one parent use a bus, as at level 1.
- The coordinator card and its line stay above the manager card (decision
  65). They move with the tree when the view scrolls sideways.

### AC4 Decision 75 and 76: scroll

- When the tree is wider than the region, the view scrolls sideways and keeps
  the selected card on screen. At 80 columns, the selected card is at the
  centre when possible. The boards show where it is at 120 columns.
- The last body row shows `←→ at the edge scrolls · children A–B of N` on
  the left in `dim`, and on the right the name of the next hidden child with
  `→`. When no child is hidden on the right and one is hidden on the left,
  the right text is `← NAME`, as the overview scroll line. A child that is
  only partly on screen counts as hidden, as `TreeWide` shows.
- The bottom border shows the `━` thumb (`component.Thumb`).
- When the tree is higher than the region, the view scrolls vertically and
  keeps the selected card on screen. The right border shows the `┃` thumb.

### AC5 Decision 77: scroll follows the selection only

The view scrolls only to keep the selection on screen. There is no free
scroll of the view: no `pgup`, `pgdn`, or wheel scroll. The user chose this.

### AC6 Decision 78: AgentsNested and AgentsCoordinatorChild

- These boards show the one-pod view with the rail. The new code draws them
  as trees.
- Render the 2 new grids from the code at their board size. Write them to
  `.hive/spec/tui-v2/tasks/125-pod-tree-scroll/evidence/new-grids/` in the grid format of
  the exporter. Commit and push them. Send the path to the manager. Do not
  change their golden files yet.
- Hold until the manager sends the approval. Then the manager stages the 2
  approved boards. Copy them over the `.dc.html` files. Run the exporter.
  Each exported grid must equal its grid from the step above. Commit the
  grids. Then the golden checks use them. In the spec, record that the new
  grids replace the old ones.
- Do the other work while you hold.

### AC7 Tests

- Write each test first and see it fail for the correct reason.
- The golden check passes for each of the 4 new boards, for the 2 boards of
  AC6, and for each other board that passes on `dev`.
- Model tests with literal expected glyphs, tokens, and text:
  - a nested child.
  - several nested children with a bus.
  - sideways scroll that follows `←` and `→` across all children.
  - the scroll line text at the left edge, in the middle, and at the right
    edge.
  - vertical scroll that follows `↓` and `↑`.
  - `pgup` and `pgdn` do not move the view.
  - the thumbs at both widths.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Board copies, grids, fixtures | V1, V2 |
| AC2 | The spec has decisions 74 to 78 | V-gate |
| AC3 to AC5 | The one-pod view matches the boards and the rules | V3, V4, V5 |
| AC6 | The 2 boards use approved new grids | V2, V4, V5 |
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
- Expected: a commit has the 4 new grids and the 2 approved grids. No other
  grid changes.

### V3 Tests red and green

- Proves: AC3 to AC7.
- Command: run the new tests on the base code, then on the task branch. Run
  each new test with `-count=50` and with `-race`.
- Expected: each fails on the base and passes on the task branch in each run.

### V4 Live check

- Proves: AC3 to AC6.
- Command: the live check of decision 43 for each of the 4 boards and the 2
  boards of AC6 at its board size, with
  `COLORTERM=truecolor hive --demo FILE --size WxH`. Then, in the demo of
  `CompactTreeWide`, press `←` and `→` across all children, and `↓` and `↑`.
  Read the pane after each key. Press `pgup` and `pgdn` and read the pane.
- Expected: each capture equals its grid. The selected card stays on screen
  after each key. The scroll line and the thumbs follow AC4. `pgup` and
  `pgdn` do not move the view.

### V5 Visual alignment

- Proves: AC3 to AC6.
- Command: run the real TUI in a herdr pane for the wide tree and the nested
  tree, each at 120×40 and at 80×24, also after a sideways and a vertical
  scroll. Compare the capture at the board size with the board grid, cell by
  cell. Look at each capture and check each point:
  - Each panel border, pod border, card border, and corner lines up.
  - Each tree line meets each card joint and each bus with no gap, at each
    level.
  - The coordinator line meets the pod border and the manager card.
  - A clipped card or pod border at the region edge stops cleanly.
  - The scroll line text and the thumbs agree with the scroll position.
- Expected: each capture at the board size equals its grid. Each point
  passes at both sizes. Put each capture, each diff, and a table of the
  points with pass or fail in the evidence. A mismatch is a fail.
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
- Never open the Chat input on the real home. Never send chat.
- Close only the panes and the processes that you started. A demo quits
  with two `ctrl+c` presses.
- Remove each scratch file that you make in `.hive/tmp/` after the check.

## Out of scope

The overview. Health › Agents. Each other screen. The canvas.

## Finish

Commit on the task branch with a subject such as `feat(tui): draw the
one-pod view as a scrolled tree`. Push the task branch. Write the evidence.
Run `hive task complete`. Send the commit SHA and the V1 to V5 results to the
manager with `hive inbox send --to manager-hive`.
