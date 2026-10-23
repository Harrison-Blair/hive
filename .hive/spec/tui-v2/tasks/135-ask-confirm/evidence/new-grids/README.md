# New grids of task 135

Each grid in this folder is the code grid of one screen at its board size.
A scratch test (`TestZZDumpGrids`, not committed) loaded each fixture with
the demo loader, rendered `View()` with `renderGrid`, and wrote it with
`cellgrid.WriteGrid`. This is the path of `checkBoard`. Each `.txt` file is
the plain text of the same view. The fixtures of the review grids are in
`fixtures/`. The fixtures of the 6 existing boards are the fixtures in
`design/`, with no change.

The terminal captures are in `../captures/`: one `.ansi` pane read, one
`.png`, and one `.diff.txt` for each grid. Each capture equals its grid here
cell for cell (0 cells differ). The key runs are in `../v3/`.

## The review (user chats c185 and c187)

- `enter` on the last question opens the review in the same card. It
  sends nothing. This applies to each completion path: `enter`, a final
  digit, a click, `enter` on a `note`, and `enter` on Other text.
- The card title is `? ASKER asks · confirm`. The body has one row for each
  question: `HEADER → LABEL, LABEL · note: TEXT`. The Other text replaces
  the Other label, as in the answer prompt. Long text wraps to the card
  width. Nothing is cut.
- A blank row and the `Send` row follow. For a stopped asker the row is
  `Send to PARENT`.
- `›` marks the highlighted row. `Send` is the first selection, so a second
  `enter` sends.
- `↑↓` moves the highlight over the questions and `Send`. `enter` on a
  question opens its options, picks, and `note` again. Its completion
  returns to the review with `Send` selected. Other questions keep their
  draft.
- A highlighted row that is taller than the dock pages with the same
  arrows: `↓` scrolls the row by one dock height less one line until its
  last line shows, and then moves to the next row. `↑` pages back while
  the row is scrolled, and then moves to the row above, at its first line.
  The first visible line of the highlighted row has the `›` mark. So each
  part of a long pick list or `note` can show, with the existing height
  cap and no new key (manager review, inbox after record 550).
- Only `↑↓`, `enter`, `x`, and `J/K` act in the review. Each other key does
  nothing: digits, space, tab, backspace, text, `i`, `ctrl+o`, `pgup`,
  `pgdn`, `?`, `esc`, `ctrl+]`, `←→`, and `g/G`. The global quit
  (`ctrl+c` twice) does not change. A click on the card does nothing. The
  mouse wheel is not a key and keeps its feed scroll.
- The bottom border: `enter send · ↑↓ question to change · x drop` on
  Send, `enter change · …` on a question, and `enter sends to PARENT · …`
  on Send for a stopped asker.
- The footer label is `[review]`: `↑↓ question · enter send · x dismiss ·
  J/K next/prev card` (120 columns), and `↑↓ question · enter send · x
  dismiss · J/K card` (80 columns). On a question, `enter change` replaces
  `enter send`.
- The card keeps the dock height cap. It scrolls to keep the highlighted
  row in view.

### Return after an edit

After an edit, the review opens with `Send` selected (as each entry to
the review does), so one `enter` sends. The edited row shows its new
answer. At 80x24 with long content, the card scrolls to keep `Send` in
view. Run `flow-*/7-return-review` shows it. The manager confirmed that
this follows the brief.

### Long rows (reproduction before the fix)

With commit `ff977f6d`, the dock showed a highlighted row from its first
line, and the arrows moved whole rows. A row taller than the dock lost its
middle. `../v1/long-row-repro-before.txt` records one question with a
`note` of 900 words and `MIDDLE-SENTINEL` after word 450: at 120x40, words
`w365` to `w590` never showed; at 80x24, words `w117` to `w817` never
showed. `MIDDLE-SENTINEL` never showed at either width. The paging rule
above fixes it. `../v3/tall-*/summary.txt` shows each word, the middle
sentinel, and the final sentinel in the live TUI.

## Review grids

| Grid | Size | Case | Planned use |
| --- | --- | --- | --- |
| `ChatAskReview` | 120x40 | one question, a `note`, Send selected | permanent board |
| `ChatAskReviewMulti` | 120x40 | multi-select with Other text, a `note` on question 2 | permanent board |
| `CompactChatAskReview` | 80x24 | one question with a `note` | permanent board |
| `ChatAskReviewHighlight` | 120x40 | question 1 highlighted, `enter change` | evidence only |
| `ChatAskReviewStopped` | 120x40 | stopped asker, `Send to manager-hive` | evidence only |
| `ChatAskReviewLong` | 120x40 | 5 picks, long Other text, two long notes, wrapped | evidence only |
| `CompactChatAskReviewLong` | 80x24 | the same at 80 columns, question 1 highlighted, scrolled | evidence only |
| `CompactChatAskReviewMulti` | 80x24 | `ChatAskReviewMulti` at 80 columns | evidence only |
| `ChatAskReviewTall` | 120x40 | a `note` of 900 words with a middle and a final sentinel, question highlighted | evidence only |
| `CompactChatAskReviewTall` | 80x24 | a `note` of 300 words with the sentinels, question highlighted | evidence only |
| `ChatAskFlow`, `CompactChatAskFlow` | both | fixtures of the key runs only, no grid | evidence only |

## Existing boards that change

`diff-rows.txt` has the cell counts. `diff-text.txt` has each old and new
row. Rows of the card hint and of the footer change, because `enter` on
the last question now opens the review:

| Board | Rows | Change |
| --- | --- | --- |
| `ChatAsk` | 35, 39 | `enter send` → `enter review` |
| `ChatAskMulti` | 35, 39 | `enter on last sends` → `enter on last reviews`; note footer `enter send with note` → `enter keep note` |
| `ChatAskPreview` | 35, 39 | `enter send` → `enter review` |
| `ChatAskStopped` | 35, 39 | `enter sends to manager-hive · x dismiss` → `enter review · sends to manager-hive · x dismiss`; footer `enter send to parent` → `enter review for parent` |
| `CompactChatAsk` | 19, 23 | `enter send` → `enter review` |
| `Help` | 24 to 33 | new line `ask review: enter on the last question · ↑↓ question · enter send or change`. One of the two blank rows above `FILTER FIELDS` is removed so that the new line fits at 120x40 with no scroll. The `ask card:` line of decision 102 does not change. |

The Other footer (`[other] type the answer · enter keep answer · ctrl+]
leave`, was `enter send`) is on no board.

## Boards that do not change

Each other board in `design/`. A scratch run rendered all 47 fixtures of
`design/` and compared each with its grid: only the 6 boards above differ,
and `Components`. `Components` is a sheet of components. It differs from
the code at the base commit in the same way (as task 134 records), and this
task does not touch it.
