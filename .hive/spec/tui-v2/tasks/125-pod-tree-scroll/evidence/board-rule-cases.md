# Board rule cases

Each case is a board cell range that breaks a spec rule that exists and adds
no new design. The code follows the rule. Rows and columns are 0-based, as in
the grid files. The manager accepted this plan on 2026-10-22 (inbox record
from manager-hive). The code grids of the 3 boards are in `new-grids/`.

| Board | Rows | Columns | Board shows | Rule that the code follows |
| --- | --- | --- | --- | --- |
| `CompactTreeWide` | 7–8 | 1–78 | The pod top border is on the row below the coordinator card. No row is between them. | Decision 65: one bus row is below the coordinator card, and in the one-pod view the line turns on that row. `CompactLivePod` (row 8) and `CompactTreeNested` have this row. `CompactTreeNested` agrees with its scroll position (top 5) and its `┃` thumb only with this row. With the row, the tree is 21 rows high in 17 rows. To keep the selected card on the screen, the view scrolls down 1 row (decision 76), so rows 4–7 and the `┃` thumb (14 rows from row 5) change. |
| `TreeWide` | 36–37 | 2–117 | The scroll line is on row 36 and the legend on the last body row 37. | Task 33 decisions (legend and scroll line) and AC4 of the brief: the scroll line is on the last body row, and the legend moves up one row. `AgentsMany` rows 36–37 show this order. |
| `TreeNested` | 30–31 | 27–92 | One blank row is between `h shows 1 stopped` (row 29) and the pod bottom border (row 31). | Task 33 decisions and `podLines`: the stopped text is on the second row below the last child, and the pod border is the next row. `LivePod` and `TreeWide` (rows 24–25) show no blank row after the text. The pod bottom border is on row 30. |
