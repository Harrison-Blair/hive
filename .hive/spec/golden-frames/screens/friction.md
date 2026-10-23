# Friction

Friction shows the dogfood friction entries of `reference/dogfood/friction/` as one list. The two views Open and Resolved of `tui-v2` decision 17 go away. The resolved entries sit behind one fold line.

## Layout at 120×40

| Rows | Region |
| --- | --- |
| 0 | title row: `friction` in `accent` bold, `· N open · M resolved` in `dim`, right text `/ filter` |
| 2 | column headers `ENTRY  AGE  SEEN  LAST  DECISION` in `dim` bold |
| 3 to the fold line | the open entries, one row each |
| the fold line | `▸ resolved · M` in `dim` bold, then `enter opens` in `dim` |
| 21 to 37 | the details box |
| the last rows | the status rows |

## Rules

1. An entry row shows the entry name, the age since `opened` in days, the seen bar, the last-seen date, and the decision text. The seen bar is one `▮` for each occurrence, up to 10. The bar is `accent` from 3 occurrences, else `dim`. The decision text is the `decision` field of the entry, cut with `…` at the row end.
2. The selected row starts with `▶` in `accent` bold and shows its name in `accent` bold. The first selection is the first open entry.
3. `enter` on the fold line expands the resolved entries below it. The fold line then reads `▾ resolved · M` and `enter folds`. A resolved row shows its resolved date in the LAST column. `↑` and `↓` walk into the resolved rows. `enter` on the fold line again folds them. The list scrolls with the `┃` thumb on column 119 when its rows do not fit.
4. The details box shows the selected entry. Its title is `NAME · STATUS · opened DATE · seen N · PATH`. Its body shows the sections SUMMARY, REPRODUCTION, OCCURRENCES, and DECISION in `accent` bold, each with its text. For a resolved entry, the title shows `Resolved DATE` as the status.
5. The list and the details box are the two boxes of README decision 6. `ctrl+]` moves the focus. The focused box takes `↑`, `↓`, `pgup`, `pgdn`, `g`, and `G`. The details box scrolls with the `┃` thumb on its right border.
6. `/` opens the filter line on row 2 in place of the column headers. The filter matches the entry name and the title, case-insensitive. `enter` applies the filter and closes the line. `esc` closes the line and keeps the applied filter. While a filter applies, the title row shows `· / TEXT` in `accent`. With no match, the list shows `no entry matches TEXT` in `dim`.
7. With no open entry, the list shows `no open entries` in `dim` on row 3. The fold line and the details box stay. With no entry at all, the details box shows `no entry` in `dim`.
8. There is no key that opens the entry file. README decision 25 applies.

## States

| State | Trigger | What shows | Board |
| --- | --- | --- | --- |
| open list | open entries exist | rules 1, 2, and 4 | `GfFriction` |
| resolved expanded | `enter` on the fold line | rule 3 | `GfFrictionResolved` |
| no open entries | no entry has the status Open | rule 7 | `GfFrictionEmpty` |
| filter | `/` with a typed text | rule 6 | `GfFrictionFilter` |
| filter, no match | a filter with no match | rule 6 | no board |
| details focus | `ctrl+]` | the details box has the `accent` border and scrolls | no board |
| long entry | an entry with more lines than the box | the `┃` thumb on the details box | no board |

## Keys

| Focus | Key | Action |
| --- | --- | --- |
| list | `↑`, `↓`, `g`, `G` | select |
| list | `enter` | expand or fold the resolved entries, on the fold line |
| list | `/` | filter |
| list, details | `ctrl+]` | move the focus to the other box |
| details | `↑`, `↓`, `pgup`, `pgdn`, `g`, `G` | scroll |
| filter | printable, `backspace` | edit |
| filter | `enter` | apply |
| filter | `esc` | close, keep the applied filter |
| any | `esc` | Chat, when no filter line is open |
