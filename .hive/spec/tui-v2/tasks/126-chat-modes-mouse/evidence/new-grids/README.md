# New grids of task 126

Each grid in this folder is the code grid of one board at its board size.
The scratch test `TestZZCodeGrids` (removed after the run) built the model
from the fixture of each board with the demo loader, rendered `View()` with
the TrueColor profile, converted it with `cellgrid.Convert`, and wrote it
with `cellgrid.WriteGrid`. `Components.grid` is a component sheet with no
screen. Its new grid is the committed grid with row 7 (the Header sample)
drawn by `component.Header{...}.View(120)`, as `TestHeaderGolden` checks.
`TooSmall` does not change.

## Rows that change

The scratch program `.hive/tmp/gridrows` (removed after the run) compared
each new grid with the committed grid in `.hive/spec/tui-v2/design/`, cell by
cell, with `cellgrid.ReadGrid` and `cellgrid.Compare` and no mask:

```text
$ go run ./.hive/tmp/gridrows .hive/spec/tui-v2/design .hive/spec/tui-v2/tasks/126-chat-modes-mouse/evidence/new-grids
AgentDetails.grid              48 cells differ, rows [0] <nil>
AgentDetailsGeneral.grid       48 cells differ, rows [0] <nil>
AgentsCoordinatorChild.grid    48 cells differ, rows [0] <nil>
AgentsHiddenParent.grid        48 cells differ, rows [0] <nil>
AgentsHome.grid                48 cells differ, rows [0] <nil>
AgentsMany.grid                48 cells differ, rows [0] <nil>
AgentsNested.grid              48 cells differ, rows [0] <nil>
Chat.grid                     108 cells differ, rows [0 39] <nil>
ChatEvents.grid               164 cells differ, rows [0 37 39] <nil>
ChatInput.grid                297 cells differ, rows [3 5 7 10 12 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 36 37 39] <nil>
ChatNormal.grid               220 cells differ, rows [5 7 10 12 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 36 39] <nil>
ChatStatus.grid               108 cells differ, rows [0 39] <nil>
Compact.grid                   18 cells differ, rows [0] <nil>
CompactAgents.grid             18 cells differ, rows [0] <nil>
CompactAgentsHome.grid         18 cells differ, rows [0] <nil>
CompactChat.grid               72 cells differ, rows [0 3 23] <nil>
CompactChatNormal.grid        405 cells differ, rows [1 3 4 5 6 8 9 10 11 12 13 14 15 16 17 18 20 23] <nil>
CompactChatStatus.grid         72 cells differ, rows [0 3 23] <nil>
CompactFriction.grid           18 cells differ, rows [0] <nil>
CompactHealth.grid             18 cells differ, rows [0] <nil>
CompactLiveOverview.grid       18 cells differ, rows [0] <nil>
CompactLivePod.grid            18 cells differ, rows [0] <nil>
CompactLogs.grid               18 cells differ, rows [0] <nil>
CompactTreeNested.grid         18 cells differ, rows [0] <nil>
CompactTreeWide.grid           18 cells differ, rows [0] <nil>
Components.grid                48 cells differ, rows [7] <nil>
Friction.grid                  48 cells differ, rows [0] <nil>
FrictionResolved.grid          48 cells differ, rows [0] <nil>
Health.grid                    48 cells differ, rows [0] <nil>
HealthAgents.grid              48 cells differ, rows [0] <nil>
HealthAgentsGeneral.grid       48 cells differ, rows [0] <nil>
Help.grid                     111 cells differ, rows [0 30] <nil>
LiveOverview.grid              48 cells differ, rows [0] <nil>
LivePod.grid                   48 cells differ, rows [0] <nil>
Logs.grid                      48 cells differ, rows [0] <nil>
Main.grid                      48 cells differ, rows [0] <nil>
TreeNested.grid                48 cells differ, rows [0] <nil>
TreeWide.grid                  48 cells differ, rows [0] <nil>
Work.grid                      48 cells differ, rows [0] <nil>
WorkGoals.grid                 48 cells differ, rows [0] <nil>
WorkQueue.grid                 48 cells differ, rows [0] <nil>
```

- Each board outside Chat and Help changes only in row 0, the header: the
  count `? 0 asks` (120 columns, 48 cells) or `? 0` (80 columns, 18 cells)
  before `needs you`. `Components` changes only in row 7, its Header sample.
- `Help` changes in row 0 and in row 30, the CHAT line of the key map:
  `i or a click on the input types · ctrl+] or a click on the feed leaves`.
- `Chat`, `ChatStatus` (input mode): row 0 and row 39, the footer
  `[input] enter send · shift+enter newline · ctrl+] leave input · ctrl+o events · wheel or pgup/pgdn scroll`.
- `ChatEvents` (normal mode): row 0, row 37 (the dim placeholder
  `i or click here to type` in place of `>`), and row 39 (the `[normal]`
  footer).
- `CompactChat`, `CompactChatStatus` (input mode): row 0, row 3 (the
  ` ↑ 42 earlier ` note from column 40), and row 23 (the compact `[input]`
  footer).
- `ChatNormal`, `ChatInput`, `CompactChatNormal`: the new boards of this
  task. Their grids are the exported board grids. The code follows the
  spec rules that `board-rule-cases.md` lists, so their code grids differ in
  the rows above.

## Approval

On 2026-10-23 the manager sent the approval of the user (chat c133–c143):
the code follows the rule for each of the 14 cases, and the 41 code grids of
commit 6bf9c8d4 are approved. The approval also changed the input footer:
`click feed leave` is removed and `ctrl+o events` is added (decision 57). A
click on the feed still stops input mode. So `Chat`, `ChatStatus`, and
`ChatInput` were rendered again. Only row 39 of each changed.


The golden files do not change until the manager sends the approval. Then
each approved board replaces its `.dc.html` file, the exporter runs, and
each exported grid must equal its grid in this folder.
