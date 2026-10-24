# Board changes, task 137

The user answered ask a18 on 2026-10-23, with the manager rulings for C7 and C10. The manager ruled C12 to C15 under the same answer: the spec wins, and the worker fixes the copies. C16 follows the same ruling. The worker changed the board copies in `design/` and exported the grids again. `board-review.html` shows each board before and after, side by side, with `▶` on each changed row. Row and column numbers are 0-based.

`canvas.json` did not change. Its pinned hash stays.

## Cases

- C2: GfChatTall rows 31 to 37, column 119: the composer thumb `┃` becomes the border `│`. The 10 text rows fit in the 12-row frame, so the text does not scroll (rule 11).
- C3: GfChatEmpty row 2: `● coordinator working · started 09:41` becomes `● working 2m` (rule 8, decision 11).
- C4: GfChatNoCoordinator row 35: `error: no live coordinator · ^y health, r restarts it` becomes `error: no live coordinator` (rule 12). The placeholder of rule 10 stays.
- C5: GfChatMarkdown old row 24: the line wraps at 118 cells. After the C11 move, new row 22 ends with `Details are in`, and new row 23 is `tasks/12-board-fixtures/evidence/V4.md.` with the path in `code`.
- C7: GfComponents. Row 7: `? 1 asks` becomes `? 1 ask`. Row 8: `● idle 5m` becomes `○ idle 5m` with `○` in `dim` bold; the row at 80 columns removes the whole `chords` and `? help` segments (decision 10); the clock moves from column 76 to column 74, the right edge of an 80-column row less the pad; the `●` of `supervisor` is bold. Row 9: the `●` of `supervisor` is bold. Rows 43 and 44: the open event shows `▾ manager-store 14:05` and the full text from column 1 on the next row (rule 3). Rows 45 and 46: the answered line of rule 6 is a message from `you`: `› you 14:23`, then `answered manager-hive · Model → Yes, restart now`. Row 41 does not change: no rule gives other cells. Rows 47 to 65 are wave C and do not change.
- C9: GfChatAskPreview rows 25 to 38: the card loses 2 empty rows, and the frame starts on row 27, not row 25. The card has the question, a blank row, the 4 options, and the `note` row. One empty row stays below the `note` row, because the preview box (6 rows from the first option row) needs it.
- C10: Row 0: the `· · · N earlier · · ·` row goes. Column 119: the transcript thumb `┃` goes. The shown content fits in the transcript rows (ruling option A). Model tests prove rule 7 and the `earlier` row with more content than rows.
- C11: The transcript moves up by 2 rows, so the first message is on row 0, as the layout table gives (ruling option B). Each transcript row is the old row 2 rows lower, cell for cell, except the C5 rows of GfChatMarkdown. The ruling names 12 boards; 13 boards have a transcript, and the worker moved all 13. GfChatEmpty does not change: its welcome block stays on rows 1 to 5 (rule 8).
- C12: GfChatMarkdown rows 6 to 21: the table is the Glamour v2.0.1 table at 118 cells (rule 4, `tui-v2` decision 32). Each column is 29 cells wide, the last 28, with the rules at columns 30, 60, and 90. One rule row is below the header, and the body rows have no rule rows. The rows below the table move up by 2 rows. Glamour breaks the path of the last paragraph after the hyphen of `12-board-`, so row 20 ends with `tasks/12-board-` in `code`, and row 21 is `fixtures/evidence/V4.md` in `code` and `.` in `fg`. This replaces the C5 text of the path.
- C13: GfChat rows 14 to 16 and GfChatMarkdown rows 14 to 18, columns 1 and 2: the 2 margin spaces of each code line are `fg`, not `code`, as Glamour gives them.
- C14: GfChatAskMulti row 39: the `n close note` segment goes. In the note focus, `n` types the letter n. No other key of the row changes. README decision 33 loses `n close note` too.
- C15: GfComponents row 44: the text of the open event is padded with `dim` spaces to column 117, as GfChatEvents shows.
- C16: GfChatNoCoordinator row 8: `· not delivered` is `accent`, not `dim` (rule 5).

## Boards

| Board | Cases | Changed rows | Pinned hash (old) | New hash |
| --- | --- | --- | --- | --- |
| GfChat | C10, C11, C13 | 0, 1, 2, 3, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 24, 25, 26, 27, 28, 29, 30, 31, 32 | `b3583f8edf9bbe61` | `37626c85ecc0ed17` |
| GfChatAsk | C10, C11 | 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 13, 20, 21, 22, 23, 24, 25, 26, 27 | `fdaeef457739bf74` | `92f8f5fa1bf6eea6` |
| GfChatAskInput | C10, C11 | 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 13, 20, 21, 22, 23, 24, 25, 26, 27 | `6e5b4c55fc455a3a` | `b9522ca10df70235` |
| GfChatAskMany | C10, C11 | 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12, 13, 14, 15, 20, 21, 22, 23, 24, 25, 26, 27 | `de5a1f67a46b193a` | `83b3a28236f244af` |
| GfChatAskMulti | C10, C11, C14 | 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 13, 16, 17, 18, 19, 39 | `2e309915e90bf049` | `9b4cbdd31e97fa8c` |
| GfChatAskPreview | C9, C10, C11 | 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 13, 18, 19, 20, 21, 22, 23, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35 | `d1bf1d4ff2c164b3` | `67ea27113e8efe16` |
| GfChatAskReview | C10, C11 | 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 13, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30 | `e48d9294b44407ae` | `77445f46bf75f047` |
| GfChatAskStopped | C10, C11 | 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 13, 21, 22, 23, 24, 25, 26, 27, 28 | `10ad921968b99701` | `5f5ff41ccf00354e` |
| GfChatEmpty | C3 | 2 | `3ebd2a2c48fb2e63` | `55fcde6ae75941d4` |
| GfChatEvents | C10, C11 | 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33 | `879f0cfa49619cc0` | `1eb19efa5808360f` |
| GfChatMarkdown | C5, C10, C11, C12, C13 | 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33 | `c718c501afc8e541` | `15eb94876859b1db` |
| GfChatNoCoordinator | C4, C10, C11, C16 | 0, 1, 2, 3, 4, 5, 6, 8, 9, 10, 11, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 35 | `cd5d6f94067dca54` | `7662c324eeebc56d` |
| GfChatNotice | C10, C11 | 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33 | `19379bb3d6ecb484` | `e569788ec63c6ae2` |
| GfChatTall | C2, C10, C11 | 0, 1, 2, 3, 4, 5, 6, 8, 20, 21, 22, 23, 24, 25, 31, 32, 33, 34, 35, 36, 37 | `5e43a477544db214` | `2c8991e887975ab3` |
| GfComponents | C7, C15 | 7, 8, 9, 43, 44, 45, 46 | `223b755a06fb2bb9` | `97c1ea858d5dd933` |

The hashes are the first 16 hex digits of the SHA-256 of each `.dc.html` copy. `V6.md` gives the full hashes.
