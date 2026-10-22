# Chat working status, growing input, and event toggle

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, goal `g3-chat-status`, wave A |
| Tier | strong |
| Depends on | none. Base: `dev` |
| Grant | goal `g3-chat-status`: user chat c13 to c26, 2026-10-22. Land on `dev`. |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The user approved three new Chat boards on 2026-10-22. This task copies the
boards into the spec, exports their golden grids, adds the decisions to the
tui-v2 spec, and builds the Chat screen to match the boards:

1. The message input border shows the coordinator status and the queued
   mark.
2. The message input wraps and grows to a cap, then scrolls inside.
3. `ctrl+o` opens and folds all events in the feed.

When a board and this brief do not agree, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`. Do not choose for the
user.

## Design

The approved boards, copied from the canvas "Hive TUI Redesign"
(https://claude.ai/artifact/VrhcA98g9G2G9cFQao5Z9h, version
1792696057-b27d):

- `/home/penguin/source/hive/.hive/tmp/g3-chat-status/boards/ChatStatus.dc.html` (120×40)
- `/home/penguin/source/hive/.hive/tmp/g3-chat-status/boards/ChatEvents.dc.html` (120×40)
- `/home/penguin/source/hive/.hive/tmp/g3-chat-status/boards/CompactChatStatus.dc.html` (80×24)

Follow the visual check rule in the worker skill. The user is colourblind.
Each state has a glyph and a word. Colour only adds to a state.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/tui-v2/design/ChatStatus.dc.html`, `ChatEvents.dc.html`, `CompactChatStatus.dc.html` | AC1. New copies. |
| `.hive/spec/tui-v2/design/<Board>.grid` and `<Board>.fixture.json` for the three boards | AC1. New. |
| `.hive/spec/tui-v2/design/Help.dc.html`, `Help.grid` | AC5. One row only. |
| `.hive/spec/tui-v2/design/Chat.dc.html`, `CompactChat.dc.html`, their `.grid` and `.fixture.json` | AC7. Repainted boards. |
| `internal/boardexport/check.go` | AC1. Add the three new boards to `Sizes`. |
| `.hive/spec/tui-v2/README.md` | AC2. The board table, decisions 55 to 57, and the symbol list of decision 13. |
| `internal/boardexport/check.go` and its tests | AC2. Only the symbol list. |
| `internal/tui/component/symbols_test.go` | AC2. Only the symbol list. |
| `internal/tui/chat.go`, `internal/tui/component/chat.go`, `internal/tui/component/input.go`, `internal/tui/component/header.go`, `internal/tui/header.go`, `internal/tui/frame.go`, `internal/tui/demo.go`, and their tests | AC3 to AC6. |
| `internal/tui/shell.go`, `internal/tui/tui.go`, `internal/tui/component/keyhints.go`, `internal/tui/component/scrollmark.go` | Only when AC3 to AC6 need a change there. Record the cause in the evidence. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/tui-v2/tasks/<task-id>/brief.md` | The brief copy, when `hive task create` did not write it. |

Do not edit other files. Do not add a mask. Do not change the exporter except
its symbol list. Add no new data source: the TUI reads only the agent record
and the chat log that it reads now.

## Requirements

### AC1 Boards, fixtures, and grids

- Copy the three boards into `.hive/spec/tui-v2/design/`. `cmp` against the
  staged copies gives no difference.
- Write one fixture file for each board with the data and the UI state that
  the board shows: the coordinator `agent_status` and `status_changed_at`,
  the chat entries with their delivery, the input text, the focus, the feed
  scroll, and the event toggle state.
- Run the exporter `internal/boardexport/cmd/boardexport` on the design
  folder. Commit the three new grids and the changed `Help.grid` only. Each
  other grid stays the same.

### AC2 Spec

- Add the three boards to the board table of the tui-v2 README.
- Add decisions 55, 56, and 57 below. Write them in ASD-STE100 Strict.
  Keep the meaning of this brief.
- Add `▸`, `▾`, and `▲` to decision 13, to the list in `check.go`, and to
  `symbols_test.go`. Add each other symbol that the three boards use and
  that the list does not have. The three lists are the same. The width test
  gives 1 for each new symbol.

### AC3 Decision 55: working status in the input border

- The input border always shows the coordinator status after the title
  ` message `: `· ` in `dim`, then the glyph, then the word and the time.
  - `working`: `●` in token `working`, then ` working Nm`.
  - `idle`: `○` in `dim`, then ` idle Nm`.
  - `blocked`: `▲` in `accent`, then ` blocked Nm`.
- `N` is the whole minutes from `status_changed_at` of the coordinator record
  to the TUI clock, rounded down. Below one minute it is `0m`.
- The data comes from the live coordinator agent record, `agent_status` and
  `status_changed_at`. These fields exist on `dev`.
- For another `agent_status` value, show `○` in `dim` and the value as the
  word. With no live coordinator record, show `○` in `dim` and the word
  `none`. With no `status_changed_at`, show no time. Record this choice in
  the evidence. No board shows these states.
- While the delivery of the last user chat entry is pending, add `· queued`
  in `dim` after the status.
- The compact header shows `▲` in `accent` for a blocked coordinator, as
  `CompactChatStatus` shows. The full header shows the glyph and the word
  for each status.

### AC4 Decision 56: the input grows

- The input wraps its text at a space. A word wider than the line breaks at
  the line edge. It grows one row for each text row.
- The cap of the input panel, with its borders, is 1/3 of the content rows,
  rounded down: 12 rows at 120×40, 6 rows at 80×24. The content rows are
  the rows between the tab rule and the footer.
- Past the cap, the input scrolls inside and keeps the cursor row visible.
  It draws a scroll mark on its right border, as `CompactChatStatus` shows.
- The input keeps its height when the focus goes to the feed. The feed
  shrinks by the same rows. The empty input has one text row.

### AC5 Decision 57: event toggle

- A folded event keeps one dim line, with `▸` in place of the `·` before
  the name.
- `ctrl+o` opens each event, from the feed and from the input. An open
  event shows `▾`, then its full text, wrapped, in `dim`. A line break in
  the text starts a new line. A second `ctrl+o` folds each event.
- The feed starts folded at each TUI start. The TUI does not keep the state.
  Add no noise key.
- The key hints of the feed and the input show `ctrl+o`, as the boards show.
- The Help board is the key map (decision 21). Copy
  `/home/penguin/source/hive/.hive/tmp/g3-chat-status/boards/Help.dc.html`
  (canvas version 1792696315-31ec) over
  `.hive/spec/tui-v2/design/Help.dc.html`. It adds one line in the CHAT
  section: `ctrl+o opens or folds all chat events` in `dim` at row 31.
  Export `Help.grid` again. Only row 31 changes. The Help overlay matches
  the new grid.

### AC7 Repainted Chat and CompactChat

- The user repainted the boards `Chat` and `CompactChat` to decisions 55
  and 57 (canvas version 1792697876-d233). Copy both from
  `/home/penguin/source/hive/.hive/tmp/g3-chat-status/boards/` over the
  copies in the design folder. Export `Chat.grid` and `CompactChat.grid`
  again. Update their fixtures: the coordinator is `working` for 1 minute.
- The boards show status `· ● working 1m` after the input title, `▸` before
  the folded event name on `Chat`, and `ctrl+o` in the input footer.
- Use the boards from the staging folder for all six boards. The staged
  `ChatStatus`, `ChatEvents`, and `CompactChatStatus` are also from version
  1792697876-d233. Copy them again if your copies differ.
- The 80×24 input footer is
  `enter send · ctrl+o events · esc leave input · pgup/pgdn feed`.
- The 120-column feed footer with folded events has no board. Use the text
  of `ChatEvents` with `ctrl+o open events` in place of
  `ctrl+o fold events`. The user approved this.
- Scroll marks follow the boards (`dim`), also in the focused region. The
  user approved this.

### AC6 Tests

- Write each test first and see it fail for the correct reason.
- The golden check passes for `ChatStatus`, `ChatEvents`,
  `CompactChatStatus`, `Help`, and each other board that passes on `dev`.
- Model tests cover: each status word with its glyph, the time at 0, 1, and
  59 minutes, the queued mark on and off, the cap at 120×40 and 80×24, the
  scroll mark and the cursor row past the cap, the height after `esc`, and
  `ctrl+o` from the feed and from the input. Use literal expected text.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Board copies equal the staged boards. Only the new grids, `Help.grid`, `Chat.grid`, and `CompactChat.grid` change. | V1, V2 |
| AC2 | The spec has the boards, decisions 55 to 57, and the symbols. The three symbol lists agree. | V-gate |
| AC3 | The input border shows the status, the time, and the queued mark | V3, V4, V5 |
| AC4 | The input grows, caps, and scrolls | V3, V4, V5 |
| AC5 | `ctrl+o` opens and folds events. Help shows the key. | V3, V4, V5 |
| AC6 | The tests fail on the base and pass on the task branch | V3 |

## Verification items

### V1 Boards

- Proves: AC1.
- Command: `cmp` of each staged board, `Help.dc.html` included, against its copy. Capture both with
  the Firefox command of the verifier skill. Put the profile folder and the
  captures in `.hive/tmp/` of the worktree. Remove them after the check.
- Expected: no difference. No capture shows `{{`.

### V2 Grid diff

- Proves: AC1, AC5.
- Command: compare each grid on `dev` with the task grid, cell by cell, with
  `internal/lib/cellgrid`.
- Expected: three new grids. `Help.grid` changes only in row 31.
  `Chat.grid` and `CompactChat.grid` change only in the cells of AC7. Each
  other grid has no difference.

### V3 Tests red and green

- Proves: AC3 to AC6.
- Command: run the new tests on the base code, then on the task branch.
- Expected: each fails on the base and passes on the task branch.

### V4 Live check

- Proves: AC3, AC4, AC5.
- Command: the live check of decision 43 for `ChatStatus` at 120×40,
  `ChatEvents` at 120×40, `CompactChatStatus` at 80×24, `Chat` at 120×40,
  `CompactChat` at 80×24, and `Help` at 120×40, with `COLORTERM=truecolor hive --demo FILE --size WxH`. Then, in
  the demo of `ChatStatus`, press `ctrl+o` and `esc`. Read the pane after
  each key.
- Expected: each capture equals its grid. After `ctrl+o` the events show
  `▾` and full text. After `esc` the input keeps its height.

### V5 Visual alignment

- Proves: AC3, AC4, AC5.
- Command: run the real TUI with `COLORTERM=truecolor hive --demo FILE --size WxH`
  in a herdr pane for each board state: working, queued, and tall input
  (`ChatStatus`), open events and idle (`ChatEvents`), and capped input and
  blocked (`CompactChatStatus`). Capture each state at 120×40
  and at 80×24. Compare the capture at the board size with the board grid,
  cell by cell. Look at each capture and check each point:
  - Each panel border and corner lines up.
  - The status text sits in the input top border after the title, with no
    broken border.
  - The input right border and the scroll mark stay in their column.
  - The feed bottom border meets the input top border with no gap and no
    overlap, at each input height from one text row to the cap.
  - The text of an open event starts in the same column as the text of a
    message.
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
- Close only the panes and the processes that you started.

## Out of scope

Each other screen. The canvas. A persisted toggle state. A noise key.

## Finish

Commit on the task branch with a subject such as `feat(tui): show the
coordinator status in the chat input`. Push the task branch. Write the
evidence. Run `hive task complete`. Send the commit SHA and the V1 to V5
results to the manager with `hive inbox send --to manager-hive`.

## Changes

- 2026-10-22: The user requires a cell-for-cell match with the three boards
  and a visual check of the alignment. AC6 golden tests cover each board
  state. New item V5: live captures at 120×40 and 80×24 for each board
  state, and the alignment points. A mismatch is a fail.
- 2026-10-22: The coordinator painted the Help line on the canvas (version
  1792696315-31ec). AC5 now copies the staged `Help.dc.html`. The user
  approved the status assumptions of AC3.
- 2026-10-22: The user repainted `Chat` and `CompactChat` (AC7, canvas
  version 1792697876-d233). The three new boards go in `Sizes` of
  `check.go`. The user approved the folded feed footer and dim scroll marks.
