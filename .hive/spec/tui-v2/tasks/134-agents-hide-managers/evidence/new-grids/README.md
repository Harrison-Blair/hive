# New grids of task 134

Each grid in this folder is the code grid of one board at its board size.
The scratch test `TestZZDumpGrids` (not committed) built the model from the
fixture of each board with the demo loader (`newTestDemoModel`), rendered
`View()` with `renderGrid`, and wrote it with `cellgrid.WriteGrid`. The same
path is in `checkBoardGolden`. Each `.txt` file is the plain text of the same
view. The terminal captures are in `../captures/`: one `.ansi` pane read, one
`.png`, and one `.diff.txt` for each grid. Each capture is equal to its grid
here cell for cell (0 cells differ).

## Changes

The user approved these changes in ask `a5` (coordinator record 536). The
grids here are the approved code grids.

1. A current manager that is stopped, exited, or stale hides with `h`, as
   each other stopped card does. The manager exemption of task 129 is gone.
2. A pod with no visible card is not on the Agents screen. `h` shows it
   again.
3. A hidden manager with a live child keeps the pod border and title. The
   children are drawn as in the home pod. In the overview, the top cards
   are on content column 0 with no rail joint. The coordinator drop on the
   pod top border is the anchor. In the one-pod view, the coordinator line
   goes down through the pod top border (`┼`) to the top joint of each top
   card (rule A of decision 74). A nested child keeps its parent. A direct
   child of the hidden manager is a top card of the pod, of each role. This
   includes a general agent (`api-review-1` in the HideManager fixtures). A
   child of a hidden worker stays in `home` (general-agent decision 39).
4. The pod note counts the hidden manager (`h shows N stopped`) and a shown
   stopped manager (`h hides N stopped`).
5. The hint: while `h` hides one or more cards, the box title ends with
   `· h shows N stopped`. N is the count of all hidden cards, also in the
   hidden pods and in home. It shows in the overview and the one-pod view,
   at each width, and also when every project pod is gone. The title is the
   one place that stays on the screen at 80 columns.
6. When each project pod is hidden, the body shows the coordinator card,
   the `home` pod with its visible cards, and the dim line
   `no project agents` above the legend. The notice alone shows only when
   no card is visible.

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
| `HideManagerOverview` | 120×40 | `manager-api` stopped, `t3-worker-1` live, `t3-verifier-1` exited, `api-review-1` (role `agent`, child of `manager-api`) idle: the `example-api` pod keeps its border and title, `t3-worker-1` and `api-review-1` are top cards, note `h shows 2 stopped`. |
| `CompactHideManagerOverview` | 80×24 | The same at 80 columns. |
| `HideManagerPod` | 120×40 | The one-pod view of `example-api`: the line goes through the pod top border (`┼`) to a bus and the top joints of `t3-worker-1` (`┷`, selected) and `api-review-1` (`┴`). |
| `CompactHideManagerPod` | 80×24 | The same at 80 columns. |
| `HideAllPods` | 120×40 | Each project agent is stopped, exited, or stale: no project pod, title `project agents · 0 projects · h shows 8 stopped`. The coordinator card and the `home` pod with `plan-review-1` show, and `no project agents` is on the row above the legend. |
| `CompactHideAllPods` | 80×24 | The same at 80 columns. `no project agents` is on the last body row. |

## Boards that do not change

`Main`, `AgentsMany`, `AgentDetails`, `AgentDetailsGeneral`, `CompactAgents`,
`AgentsHome`, `CompactAgentsHome`, `AgentsNested`,
`AgentsCoordinatorChild`, `Help`, `TooSmall`, and each board of the other
screens (Chat, Work, Friction, Health, Logs). `Components` is a sheet of
components; its grid differs from the code at base and at this task in the
same way, and the task does not touch it.

## Decisions on the points to check

- A project pod with no agent records is not on the screen (decision 106).
- With each project pod hidden, the coordinator and the `home` pod stay on
  the screen with the notice line (ask `a5`, decision 110).
- Selection: when the selected pod hides, the first pod is selected. The
  TUI keeps the id of the pod, and the pod is selected again when `h`
  shows it (decision 111).
