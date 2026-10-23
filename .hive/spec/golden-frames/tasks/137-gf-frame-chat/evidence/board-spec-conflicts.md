# Board and spec conflicts, task 137

The worker exported the 15 grids from the pinned boards (version
`1792795470-106f`). The brief requires a cell-for-cell match with no mask. The
brief also requires each rule of `screens/chat.md` and README decisions 8 to
22. The cases below are cells where the board and the spec do not agree, or
where the board does not agree with itself. The brief says: stop, report, do
not choose. Row and column numbers are 0-based.

## C1 Status row on the ask boards: segments that the spec does not have

- Boards: `GfChatAsk`, `GfChatAskMulti`, `GfChatAskPreview`,
  `GfChatAskStopped`, `GfChatAskReview`, `GfChatAskMany` (row 39).
- Board: `● working 1m · ask · ↑↓ move · 1..4 pick · enter review · n note · x dismiss · J/K card · ^r input · esc input   14:22`.
  `ask` is `accent` bold. `supervisor`, `asks`, `needs_you`, `chords`, and
  `hints` are not on the row. `GfChatAskMulti` shows `note · type the note · enter keep note · n close note · ↑↓ option · space mark · ^r input`.
  `GfChatAskReview` shows `review · ↑↓ question · enter send · …`.
  `GfChatAskStopped` shows `enter review for parent`. `GfChatAskMany` ends
  with `· ? 3 asks` after the keys.
- Spec: README decision 9 gives the default row `state, supervisor, asks,
  needs_you, chords, hints, clock`. No decision replaces the row while the
  card has the focus. The `screen` segment of decision 11 is not in the
  default row, and it shows the screen name, not the focus name.
- Also: `GfChatAskInput` row 39 ends with `^r answer the ask` in place of the
  chords and `? help`. No decision gives this segment.

## C2 GfChatTall: the composer thumb shows a scroll that does not exist

- Board rows 27 to 38: the frame has 12 rows, 10 text rows. All 10 text rows
  show, from `> Next, the release notes` to the cursor at the end. Rows 31 to
  37 still have the `┃` thumb on the right border.
- Spec: rule 11. The frame grows up to one third of the content rows,
  rounded down: 39 / 3 = 13 rows. The text scrolls only beyond that. 10 text
  rows fit in 12 rows, so no scroll and no thumb. When the text scrolls, the
  first row is not visible.

## C3 GfChatEmpty: the state line of the welcome block

- Board row 2: `● coordinator working · started 09:41`.
- Spec: rule 8. The line is "the coordinator state in the form of README
  decision 11", which is `● working 2m`.

## C4 GfChatNoCoordinator: the error notice text

- Board row 35: `error: no live coordinator · ^y health, r restarts it`.
- Spec: rule 12 and the state table: `error: no live coordinator`.

## C5 GfChatMarkdown: a transcript row runs into the thumb column

- Board row 24: `… Details are intasks/12-board-fixtures/evidence/V4.m.┃`.
  The space between `in` and `tasks` is missing, the text reaches column 118,
  and `V4.md` is cut by the thumb. A Glamour wrap at 118 cells (rule 4) does
  not give this row.

## C6 The transcript thumb does not reach the end while the view follows the end

- Rule 7: the transcript follows its end. The thumb then ends on the last
  transcript row.
- `GfChat`, `GfChatAsk`, `GfChatAskStopped`, `GfChatAskReview`,
  `GfChatAskMany`: the thumb ends on the last transcript row.
- `GfChatNoCoordinator`, `GfChatNotice`: the last transcript row is 34, the
  thumb ends on row 33.
- `GfChatEvents`, `GfChatMarkdown`: the last transcript row is 35, the thumb
  ends on row 33.
- `GfChatTall`: last row 26, thumb ends on row 25. `GfChatAskPreview`: last
  row 24, thumb ends on row 23.
- One thumb rule cannot give all of these boards.

## C7 GfComponents: the sheet contradicts the spec

- Row 7: `? 1 asks`. Decision 11 gives `? 1 ask`.
- Row 8: `● idle 5m`. Decision 11 gives `○ idle 5m` with `○` in `dim`.
- Rows 43 and 44: an open event shows the text after the time on the `▾` row,
  and the next row is indented by 3 cells. `GfChatEvents` and rule 3 show the
  `▾` row with no text and the text from column 1.
- Row 45: `? manager-hive 14:21  asked · Model · answered 14:23 · Yes, restart now`.
  Rule 6 gives a line from `you`: `answered ASKER · HEADER → PICKS`.
- Row 41: the text `[tui.status].` touches the legend `?  ask` with no gap.
- Rows 47 to 65 (TaskCard, WaveColumn, Rail, Popup, CheckTile,
  DensityStrip, SeenBar) are components of the wave C screen bodies, which the
  brief puts out of scope for this task. The brief (AC8) asks for a golden
  check of each component on the sheet.

## C8 Brief: board count

- The brief overview says "Chat matches its 13 boards". The Design section
  and AC8 name 14 Chat screen boards.

## C9 GfChatAskPreview: card height

- Board rows 25 to 38: the card has 10 rows: the question, a blank row, 4
  option rows, the `note` row, and 3 blank rows. The preview box has 6 rows.
  No rule gives the 3 blank rows below the `note` row. Rule 16 gives a frame
  of at most half the content rows, and "the card gets the rest" only when it
  has more rows.

## Question

For each case, which one is the authority: the board (then the spec needs a
change), or the spec (then the grid needs a change, which is a mask or a new
board, and needs the user)?
