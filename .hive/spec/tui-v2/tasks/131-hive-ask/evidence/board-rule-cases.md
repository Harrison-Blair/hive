# Board rule cases of task 131

Each case is a cell range where a new ask board breaks a spec rule that
exists and adds no new design. The code follows the rule. The code grids are
in `new-grids/`. Rows and columns are zero-based. Each case is the same as a
case of task `126-chat-modes-mouse` that the user approved on 2026-10-23.
The ask card, the ask lines in the feed, the footers, and the header count
match each board cell for cell.

| # | Board | Row | Columns | Board shows | Rule that the code follows |
| ---: | --- | --- | --- | --- | --- |
| 1 | ChatAsk, ChatAskMulti, ChatAskPreview, ChatAskStopped | 3 | 2–22 | unfocused feed title ` chat · noise normal ` in `border`, not bold | Decision 26: `accent` is for titles. `component.Panel` draws each title in `accent` bold (decision 27). Task 126 case 10. |
| 2 | ChatAsk, ChatAskPreview | 5, 10 | 9–105 | `you` in `accent` bold, the name pad and the user text in `fg` | Decisions 15 and 26: user chat text uses `user`. The gutter name and its pad are `user` bold. Task 126 case 5. |
| 3 | ChatAskMulti, ChatAskStopped | 5, 6 | 9–27 | `you` and the user text, as case 2 | Decisions 15 and 26, as case 2. |
| 4 | ChatAsk, ChatAskMulti, ChatAskPreview, ChatAskStopped | 5, 7, 8, 10, 14 | 9–21 | agent names `coordinator`, `manager-hive`, `t126-worker-1` in `fg` bold | Decision 26: `info` for agent replies. Task 126 case 5. |
| 5 | ChatAsk, ChatAskPreview | 12 | 9–116 | event `▸ manager-hi…` in `fg` bold | Decisions 15 and 57: a folded event has one `dim` line. Task 126 case 7. |
| 6 | ChatAsk, ChatAskMulti, ChatAskPreview, ChatAskStopped | 17–27 | 119 | scroll mark `│` in `border` | Decision 5: on Chat, scroll marks use `dim`. `component.Thumb` gives the mark `┃`. Task 126 case 8. |
| 7 | CompactChatAsk | 10 | 79 | scroll mark `│` in `border` | Decision 5, as case 6. Task 126 case 9. |
| 8 | ChatAsk, ChatAskMulti, ChatAskPreview, ChatAskStopped, CompactChatAsk | 36, 36, 36, 36, 20 | 2–25 | unfocused input title ` message · ● working 1m ` in `border`, not bold | Decision 55 and decision 27: ` message ` in `accent` bold, the status word and age in `fg`. Task 126 case 11. |
| 9 | ChatAsk, ChatAskMulti, ChatAskPreview, ChatAskStopped, CompactChatAsk | 39, 39, 39, 39, 23 | after the hints | padding cells in `fg` | Decision 4: the key hints are `dim`. `component.KeyHints` pads the row in `dim`. Task 126 case 14. |
| 10 | CompactChatAsk | 1 | 35–56 | tab labels `4 Frict`, then `5 Health` and `6 Logs` 3 cells to the left | Decision 3: the tab bar shows `4 Friction`. Task 126 case 1. |
| 11 | CompactChatAsk | 3 | 2–22 | feed title `chat · noise normal` | Task 34 decision: below 120 columns, the feed title is `chat`. Task 126 case 2. |
| 12 | CompactChatAsk | 4–5 | 1–78 | blank first body row, so the ask line is on row 5 | Task 34 decision: the first body row of the feed is blank only from 120 columns. The ask line is on row 4. Task 126 case 3. |
| 13 | CompactChatAsk | 5 | 9–17 | name `manager-…` in `fg` bold | Decision 26, as case 4. Task 126 case 6. |

## Help

`Help` gets one new row for the ask keys (AC8). This row is a new design, not
a rule case:

```text
row 33, columns 21–95, dim:
ask card: ↑↓ move · 1–5 pick · space mark · tab note · x dismiss · J/K card
```

## Approval

On 2026-10-23 the user approved the rule for each of the 13 cases and the
new `Help` row (decision 104).
