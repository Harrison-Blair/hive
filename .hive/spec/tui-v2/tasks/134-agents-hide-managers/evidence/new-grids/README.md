# New grids of task 134

Each grid in this folder is the code grid of one board at its board size.
The scratch test `TestZZDumpGrids` (not committed) built the model from the
fixture of each board with the demo loader (`newTestDemoModel`), rendered
`View()` with `renderGrid`, and wrote it with `cellgrid.WriteGrid`. The same
path is in `checkBoardGolden`. Each `.txt` file is the plain text of the same
view. The terminal captures are in `../captures/`: one `.ansi` pane read, one
`.png`, and one `.diff.txt` for each grid. Each capture is equal to its grid
here cell for cell (0 cells differ).

## Changes for approval

1. A current manager that is stopped, exited, or stale hides with `h`, as
   each other stopped card does. The manager exemption of task 129 is gone.
2. A pod with no visible card is not on the Agents screen. `h` shows it
   again.
3. A hidden manager with a live child keeps the pod border and title. The
   children are drawn as in the home pod. In the overview, the top cards
   are on content column 0 with no rail joint. The coordinator drop on the
   pod top border is the anchor. In the one-pod view, the coordinator line
   goes down through the pod top border (`┼`) to the top joint of each top
   card (rule A of decision 74). A nested child keeps its parent.
4. The pod note counts the hidden manager (`h shows N stopped`) and a shown
   stopped manager (`h hides N stopped`).
5. The hint: while `h` hides one or more cards, the box title ends with
   `· h shows N stopped`. N is the count of all hidden cards, also in the
   hidden pods and in home. It shows in the overview and the one-pod view,
   at each width, and also when every project pod is gone. The title is the
   one place that stays on the screen at 80 columns.

## Changed boards (title row only)

A scratch script compared each code grid with the committed grid in
`.hive/spec/tui-v2/design/`, cell by cell (rune, token, bold), with no mask:

```text
AgentsHiddenParent         20 cells, rows [3]
CompactLiveOverview        20 cells, rows [3]
CompactLivePod             20 cells, rows [3]
CompactTreeNested          20 cells, rows [3]
CompactTreeWide            20 cells, rows [3]
LiveOverview               20 cells, rows [3]
LivePod                    20 cells, rows [3]
TreeNested                 20 cells, rows [3]
TreeWide                   20 cells, rows [3]
```

The cause in each board: the fixture has `hide_stopped: true` and hidden
stopped cards, so the title gets the hint of change 5. For example
`project agents · 2 projects · h shows 2 stopped`. No board has a stopped
manager, so no other row changes. The 20 cells are the 20 cells of
` · h shows N stopped`, which replace `─` cells of the top border.

## New fixtures (no board yet)

The fixtures are in `fixtures/`. Each is `LiveOverview.fixture.json` with a
few agent states changed.

| Grid | Size | Case |
| --- | --- | --- |
| `HideManagerOverview` | 120×40 | `manager-api` stopped, `t3-worker-1` live, `t3-verifier-1` exited: the `example-api` pod keeps its border and title, `t3-worker-1` is a top card, note `h shows 2 stopped`. |
| `CompactHideManagerOverview` | 80×24 | The same at 80 columns. |
| `HideManagerPod` | 120×40 | The one-pod view of `example-api`: the line goes through the pod top border (`┼`) to `t3-worker-1` (`┷`, selected). |
| `CompactHideManagerPod` | 80×24 | The same at 80 columns. |
| `HideAllPods` | 120×40 | Each project agent is stopped, exited, or stale: no project pod, `no project agents`, title `project agents · 0 projects · h shows 8 stopped`. |
| `CompactHideAllPods` | 80×24 | The same at 80 columns. |

## Boards that do not change

`Main`, `AgentsMany`, `AgentDetails`, `AgentDetailsGeneral`, `CompactAgents`,
`AgentsHome`, `CompactAgentsHome`, `AgentsNested`,
`AgentsCoordinatorChild`, `Help`, `TooSmall`, and each board of the other
screens (Chat, Work, Friction, Health, Logs). `Components` is a sheet of
components; its grid differs from the code at base and at this task in the
same way, and the task does not touch it.

## Points to check

- Change 2 also removes a project pod that has no agent records at all. The
  old screen drew such a pod with `no manager`. The brief says: remove a
  pod when it has no visible card.
- When every project pod is gone, the screen shows `no project agents`, as
  it did before with no project. A live home agent is then not on the
  screen (rule that exists). The hint stays on the title.
- Selection: when the selected pod hides, the first pod is selected. The
  TUI keeps the id of the pod, and the pod is selected again when `h`
  shows it.
