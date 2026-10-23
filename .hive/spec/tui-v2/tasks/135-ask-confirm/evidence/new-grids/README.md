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
- Only `↑↓`, `enter`, `x`, and `J/K` act on the card in the review. Digits,
  space, tab, backspace, text, and a click do nothing. `pgup/pgdn`, `i`,
  and `ctrl+o` keep their feed and input meanings.
- The bottom border: `enter send · ↑↓ question to change · x drop` on
  Send, `enter change · …` on a question, and `enter sends to PARENT · …`
  on Send for a stopped asker.
- The footer label is `[review]`: `↑↓ question · enter send · x dismiss ·
  J/K next/prev card · wheel/pgup scroll feed · i type` (120 columns), and
  `↑↓ question · enter send · x dismiss · J/K card` (80 columns). On a
  question, `enter change` replaces `enter send`.
- The card keeps the dock height cap. It scrolls to keep the highlighted
  row in view.

### Navigation decision for approval

After an edit, the review opens with `Send` selected (as each entry to
the review does), so one `enter` sends. The edited row shows its new
answer. At 80x24 with long content, the card scrolls to keep `Send` in
view. Run `flow-*/7-return-review` shows it.

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
