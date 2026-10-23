# Board rule cases of task 126

Each case is a cell range where a new board breaks a spec rule that exists
and adds no new design. The code follows the rule. The code grids are in
`new-grids/`. Rows and columns are zero-based.

| # | Board | Row | Columns | Board shows | Rule that the code follows |
| ---: | --- | --- | --- | --- | --- |
| 1 | CompactChatNormal | 1 | 28–79 | label `4 Frict` at 28–34, and the next labels 3 cells to the left | Decision 3: the tab bar shows `4 Friction`. At 80 columns the labels touch, and all six labels fit. |
| 2 | CompactChatNormal | 3 | 3–21 | title `chat · noise normal` | Task 34 decision: below 120 columns, the feed title is `chat`. The ` ↑ 42 earlier ` note of the brief shows from column 40, as the board shows. |
| 3 | CompactChatNormal | 4 | 1–78 | blank first body row | Task 34 decision: the first body row of the feed is blank only from 120 columns. |
| 4 | CompactChatNormal | 4–18 | 1–78 | time, name, and text in a gutter, name cut to 9 cells, text wrapped near column 59 | Decision 30: below 120 columns the feed has no gutter. The time and the name use their own line, and the text starts at the panel padding and wraps at the panel width. |
| 5 | ChatNormal, ChatInput | 5, 7, 10 | 9–118 | `you` in `accent` bold, `coordinator` in `fg` bold, user text in `fg` | Decision 15: user text uses `user`. Decision 26: `user` for user chat text, `info` for agent replies. The code draws the gutter as the approved `Chat` board does: the name and its padding in `user` or `info` bold, user text in `user`. |
| 6 | CompactChatNormal | 5–12 | 1–78 | the same name and text tokens as case 5 | Decisions 15 and 26, as case 5. |
| 7 | ChatNormal, ChatInput | 12 | 9–118 | event name `▸ manager-hi…` in `fg` bold | Decision 15: an event folds to one dim line. Decision 57: a folded event has one dim line. |
| 8 | ChatNormal, ChatInput | 20–34 | 119 | `┃` in `accent` on rows 20–30 | Decision 5: on Chat, scroll marks use `dim` in each focus region. Task 34 decision: `component.Thumb` gives the mark, and its total is the count of feed lines. With 42 earlier messages the mark is on rows 25–34. |
| 9 | CompactChatNormal | 12–18 | 79 | `┃` in `accent` on rows 12–18 | Decision 5 and task 34, as case 8. The code mark is on row 18. |
| 10 | ChatInput | 3 | 2–22 | unfocused feed title in `border`, not bold | Decision 26: `accent` is for titles. `component.Panel` draws each title in `accent` bold (decision 27). |
| 11 | ChatNormal, CompactChatNormal | 36, 20 | 2–25 | unfocused input title ` message · ● working 1m ` in `border`, not bold | Decision 55: `· ` in `dim`, then the glyph in its token, the word, and the age. Decision 27: the title ` message ` in `accent` bold. |
| 12 | ChatInput | 36 | 2–25 | focused input title and status all in `accent` bold | Decision 55, as case 11. |
| 13 | ChatInput | 37 | 2–32 | `> ` and the draft in `fg` | Decisions 15 and 26: user chat text uses `user`, as the approved `Chat` board shows. |
| 14 | ChatNormal, ChatInput, CompactChatNormal | 39, 39, 23 | after the hints | padding cells in `fg` | Decision 4: the key hints are `dim`. `component.KeyHints` pads the row in `dim`, as each approved board shows. |

## Open question for the manager

Decision 57 says: "Show `ctrl+o` in the key hints of the feed and input."
The `ChatInput` board footer has no `ctrl+o` hint. The new input hints of
the board are 103 cells. With ` · ctrl+o events` they are 119 cells, and the
footer has room for 111 cells, so the hints would be cut with `…`. The code
follows the board footer for now. `ctrl+o` works in both modes. Tell me if the
input footer must show `ctrl+o` and which hint to remove.
