# New grids of task 131

Each grid in this folder is the code grid of one board at its board size.
The scratch test `TestZZDumpGrids` (removed after the run) built the model
from the fixture of each board with the demo loader (`newTestDemoModel`),
rendered `View()` with `renderGrid`, and wrote it with `cellgrid.WriteGrid`.
`TestAskBoardsMatch` and `TestHelpMatchesItsBoard` use the same render path.

The 5 ask boards are new. `Help` changes for the ask keys (AC8).

## Rows that change

A scratch script compared each new grid with the committed grid in
`.hive/spec/tui-v2/design/`, cell by cell (rune, token, bold), with no mask:

```text
ChatAsk.grid                  230 cells differ, rows [3, 5, 7, 10, 12, 14, 22, 23, 24, 25, 26, 36, 39]
ChatAskMulti.grid             119 cells differ, rows [3, 5, 7, 8, 17, 18, 36, 39]
ChatAskPreview.grid           229 cells differ, rows [3, 5, 7, 10, 12, 14, 20, 21, 22, 23, 36, 39]
ChatAskStopped.grid           129 cells differ, rows [3, 5, 6, 8, 10, 22, 23, 24, 25, 26, 27, 36, 39]
CompactChatAsk.grid           158 cells differ, rows [1, 3, 4, 5, 10, 20, 23]
Help.grid                      75 cells differ, rows [33]
```

Each changed cell of the 5 ask boards is a rule case in
`../board-rule-cases.md` (13 cases, each the same as an approved case of task
126). No cell of an ask card, an ask line, a footer, or the header count
differs. `Help` row 33 is the new ask key line.

## Approval

Waiting for the approval of the user.
