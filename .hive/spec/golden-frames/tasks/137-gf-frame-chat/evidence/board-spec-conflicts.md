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
- Row 8, the 80-column example: the `chords` segment is cut inside the
  segment (`^g agents  ^t work` then the clock). Decision 10 removes whole
  segments from the right.
- Bold: on row 7 the state glyph and the `●` of `supervisor` are bold. On
  rows 8 and 9 the `●` of `supervisor` is not bold. On the Chat boards
  (row 39) both glyphs are bold. Decision 11 gives no bold for a glyph.
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

## C10 The transcript follows its end, but the boards leave empty rows below the last message

- Found while fixing C6 under ask a18.
- Rule 7: the transcript follows its end. When the view shows `N earlier`
  with N more than 0, the content is taller than the transcript rows, so the
  last message is on the last transcript row.
- `GfChat`: the last message ends on row 29, rows 30 to 32 are empty.
  `GfChatAsk`, `GfChatAskInput`, `GfChatAskMany`: the ask line is on row 13,
  rows 14 to 27 are empty. `GfChatEvents`: rows 21 to 35 are empty.
  `GfChatNotice`: rows 10 to 34 are empty. The other boards with
  `N earlier` are the same. Each board draws the first message on row 2,
  below the `earlier` row.
- No scroll position gives `N earlier` above the view and empty rows below
  the last message. A thumb that ends on the last row (the C6 fix) makes this
  contradiction visible.
- Option A (smallest change): on each board, remove the `· · · N earlier · · ·`
  row and the transcript thumb, because the shown content fits in the rows.
  Rule 7 and the `earlier` row of the layout then have no board, and tests
  check them in words.
- Option B: move the content of each board down so that the last message is
  on the last transcript row, and fill the rows above with earlier messages.
  This changes most transcript rows of 13 boards.
- The worker recommends option A.

## C11 After the C10 fix, rows 0 and 1 are empty above the first message

- Found while applying the C10 ruling (option A).
- Each Chat board with a transcript draws the first message on row 2. With
  the `earlier` row removed, rows 0 and 1 are empty.
- Layout table: row 0 is the `earlier` row when N is more than 0, "else the
  first transcript row". Rule 2 gives a blank row only between two messages.
  No rule gives 2 empty rows above the first message. Rule 8 puts the welcome
  block from row 1, so `GfChatEmpty` has 1 empty row on row 0.
- Option A: the code keeps row 0 for the `earlier` row (empty when N is 0)
  and row 1 empty, and the transcript starts on row 2. The boards stay as they
  are. The layout table needs a change.
- Option B: move the transcript of 12 boards up by 2 rows, so the first
  message is on row 0. The layout table stays. The welcome block of rule 8
  starts on row 1 and stays as it is.
- The worker recommends option B, because it follows the layout table, and
  the model tests of C10 then check row 0 and row 1 for overflowing content.

## C12 GfChatMarkdown: the table is not the Glamour table

- Board rows 6 to 12: each column has the width of its widest cell plus 3
  or 4 cells, the table is 82 cells wide, and a `─┼─` rule row separates each
  two body rows.
- Spec: rule 4 and the state table give "the Glamour table at 118 cells".
  `tui-v2` decision 32 (a user decision of 2026-10-21) says that Glamour
  v2.0.1 draws each table, and hive does not change the result. Glamour
  widens the narrowest column until the table is 118 cells wide, and it
  draws one rule row, below the header only.
- Option A (worker recommends): change the board rows 6 to 12 to the Glamour
  output at 118 cells. The rows below the table move up by 2 rows.
- Option B: hive draws its own table as the board shows. This reverses
  decision 32.

## C13 GfChat and GfChatMarkdown: the token of the code block margin

- Board: the 2 margin cells to the left of each code line (GfChat rows 14 to
  16, GfChatMarkdown rows 16 to 20, columns 1 and 2) are spaces in `code`.
- Glamour v2.0.1 writes the margin of a code block with no style, so the
  cells are spaces in `fg`. A space shows no colour, so the screen looks the
  same. `tui-v2` decision 31 keeps the Glamour output.
- Option A (worker recommends): change the token of these spaces on the
  boards to `fg`.
- Option B: hive styles the margin. This changes the Glamour output.

## C14 GfChatAskMulti: `n close note` while the note takes the keys

- Board row 39: `note · type the note · enter keep note · n close note · …`.
- Spec key table: `n` opens the `note` line on the ask card. While the note
  takes the keys, a printable key types. Then `n` types the letter n, and
  it cannot close the note.
- Option A (worker recommends): `n` types in the note. The hint on the board
  stays as it is, and a separate change corrects the hint text.
- Option B: `n` closes the note, and the user cannot type the letter n in a
  note.

## C15 GfComponents: the open event row is not padded

- Board row 44: the text of the open event ends after its last word.
- GfChatEvents rows 3 to 6: the text of each open event is padded with `dim`
  spaces to column 117. A space shows no colour.
- Option A (worker recommends): pad the GfComponents row to column 117 in
  `dim`, as GfChatEvents shows.
- Option B: change GfChatEvents to no padding.

## Rulings C12 to C15

The manager ruled on 2026-10-23, under the ask a18 answer: the spec wins, and
the worker fixes the copies. C12: the board shows the Glamour table. C13: the
board token is what Glamour gives. C14: in the note focus, `n` types n; the
`n close note` hint goes from GfChatAskMulti and decision 33. C15: the
GfComponents open event row is padded as GfChatEvents shows.
`board-changes.md` records each change.

## C16 GfChatNoCoordinator: the token of `· not delivered`

- Board row 8: `· not delivered` is `dim`. The state table says `no board`
  for this state, but the board shows it.
- Rule 5: `· not delivered` is `accent`.
- The worker applied the a18 ruling (the spec wins) and changed the board
  token to `accent`.
