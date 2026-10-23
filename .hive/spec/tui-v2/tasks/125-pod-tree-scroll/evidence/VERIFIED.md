# Independent verification: task 125

Result: **pass**. Agent: `t125-verifier-1`.

Checked on 2026-10-23 UTC. Task commit: `27df61da`.
Base: `dev`, `022c71971d95b859ed1218ebd57a7fff7bc84f5f`.
`git merge dev` reported `Already up to date.`

I ran each item again. I did not change project source files or worker evidence.
The files in `verifier/` contain this run's output and captures.

## Results

| Item | Result | Observation | Evidence |
| --- | --- | --- | --- |
| V1 Boards | pass | All 6 staged files equal their copies. Each staged PNG equals its copy PNG. No capture shows `{{`. | `verifier/boards.log` |
| V2 Grids | pass | All 39 exported grids equal the committed grids, cell by cell, with no mask. Only the 6 permitted grids changed. | `verifier/grids.log`, `verifier/approved-code-grids.log`, `verifier/scope.log` |
| V3 Tests | pass | The new behavior tests fail on base code and pass on the task branch. Both runs include `-count=50` and `-race`. The two changed regression guards also pass on the base. | `verifier/tests-*.log`, `verifier/mutation-red.log` |
| V4 Live check | pass | All 6 initial captures equal their grids. Arrows keep the selected card on the screen. Both page keys leave the tree unchanged. | `verifier/*.ansi`, `*.grid`, `*.diff.txt`, `*.meta.json`, `extra-live.log` |
| V5 Visual alignment | pass | Borders, joints, buses, clipping, scroll text, and thumbs pass at both widths. | `verifier/*.png`, `*.txt`, and the tables below |
| V-gate | pass | Scope agrees with the brief. All five gate commands exit 0. The first two commands print nothing. | `verifier/scope.log`, `verifier/gate.log` |

## V1 and V2: approved boards

The later approved boards replace the first boards under decision 78.
Each staged path below is relative to `/home/penguin/source/hive/.hive/tmp/g5b/`.

| Board | Staged file | Size | Live comparison |
| --- | --- | --- | --- |
| TreeWide | `approved/TreeWide.dc.html` | 120×40 | 4800 cells match |
| TreeNested | `approved/TreeNested.dc.html` | 120×40 | 4800 cells match |
| CompactTreeWide | `approved/CompactTreeWide.dc.html` | 80×24 | 1920 cells match |
| CompactTreeNested | `boards/CompactTreeNested.dc.html` | 80×24 | 1920 cells match |
| AgentsNested | `approved/AgentsNested.dc.html` | 120×40 | 4800 cells match |
| AgentsCoordinatorChild | `approved/AgentsCoordinatorChild.dc.html` | 120×40 | 4800 cells match |

For each file, I ran `cmp` against its design copy.
I captured both files with this command and a new profile folder:

```sh
firefox --headless --no-remote --profile SCRATCH --window-size=1220,780 --screenshot OUT.png file://BOARD.dc.html
```

The scratch staged folders used `design/support.js`.
I looked at all six rendered boards. The templates rendered without `{{`.
The staged and copied PNG files were pixel-equal for each board.
I removed these scratch captures and profiles after the check.

I ran the exporter from the worktree root:

```sh
mkdir -p .hive/tmp/verify125/export
go run ./internal/boardexport/cmd/boardexport -out .hive/tmp/verify125/export .hive/spec/tui-v2/design
```

A scratch Go tool used `cellgrid.ReadGrid` and `cellgrid.Compare` with no mask.
It compared all 39 exported grids with the design grids.
It also compared the 5 approved replacement grids with `evidence/new-grids/`.
Every comparison passed.
Only `TreeWide`, `TreeNested`, `CompactTreeWide`, `CompactTreeNested`, `AgentsNested`, and `AgentsCoordinatorChild` changed against `dev`.

As a negative check, I compared the TreeWide live capture with TreeNested.
The comparator rejected it. See `verifier/comparator-negative.log`.

### Rule cases

I checked the three cases in `board-rule-cases.md` against their rules.
The later approved boards include these changes. No new exception is necessary.

| First board | Rows | Columns | Rule and approved result |
| --- | --- | --- | --- |
| CompactTreeWide | 7–8 | 1–78 | Decision 65 requires the coordinator bus row. Decision 76 then moves the view down one row. The vertical thumb follows that position. |
| TreeWide | 36–37 | 2–117 | AC4 puts the scroll line on the last body row. The legend moves up one row. |
| TreeNested | 30–31 | 27–92 | The pod border follows the stopped text with no extra blank row. |

The approved AgentsCoordinatorChild board also includes rule A of decision 74.
Its coordinator line reaches the top home card through the pod border.

## V3: red and green runs

I made a scratch copy of the base with `git archive`.
I copied only changed test files and design data into it.
The implementation stayed at `022c7197`.
The task branch stayed unchanged.

```sh
RUN='TestCheckAcceptsTheTreeBoards|TestPanelBleedFillsThePadding|TestTree|TestHomeTree|TestAgentsNestGeneralAgentsUnderTheirParentCard|TestAgentsNestingUsesTheRailFormInTheOverview|TestAgentsSelectedCardJoints|TestAgentsUpDownMoveInThePod|TestAgentsWorkingCardBorder|TestAgentsFitEachSize|TestGeneralAgentBoardsMatchTheirGrids|TestEachBoardFixtureLoadsAtItsSpecSize'
go test ./internal/tui/... ./internal/boardexport/... -run "$RUN" -count=1 -v
go test ./internal/tui/... ./internal/boardexport/... -run "$RUN" -count=50
go test ./internal/tui/... ./internal/boardexport/... -run "$RUN" -race -count=1
```

| Test or group | Base result | Task result |
| --- | --- | --- |
| `TestCheckAcceptsTheTreeBoards` | fail: four boards are absent from `Sizes` | pass |
| `TestPanelBleedFillsThePadding` | build fails: base `Panel` has no `Bleed` field | pass |
| `TestTreeBoardsMatchTheirGrids` | all 4 subtests fail: base draws the rail | pass |
| `TestTreeNestedChild` | fail: nested joint and card positions differ | pass |
| `TestTreeNestedChildrenUseABus` | fail: pod width and level-2 bus differ | pass |
| `TestTreeScrollsSidewaysAt80`, `TestTreeScrollsSidewaysAt120` | fail: selected card, scroll line, and thumb differ | pass |
| `TestTreeScrollsVertically` | fail: selected card and vertical thumb differ | pass |
| `TestTreePageKeysDoNotScroll` | fail: initial thumbs differ | pass |
| `TestHomeTreeCoordinatorLineGoesToTheTopCard`, `TestHomeTreeCoordinatorLineUsesABus` | fail: base stops the line at the pod border | pass |
| Changed nesting, joint, navigation, and working-border tests | fail: base uses rail positions or rail navigation | pass |
| `TestGeneralAgentBoardsMatchTheirGrids` | fail: the 2 replaced grids differ | pass |
| `TestEachBoardFixtureLoadsAtItsSpecSize` | fail: base has 35 board sizes, not 39 | pass |
| `TestAgentsNestingUsesTheRailFormInTheOverview`, `TestAgentsFitEachSize` | pass: these guards preserve existing behavior | pass |

The 50-run red log records failure counts instead of repeated screen dumps.
The single-run red log retains the full failures.
The full gate also runs every other board test.

I made two deliberate breaks in a separate scratch copy of the task code:

- I disabled `Panel.Bleed`. Its test failed against the literal expected panel.
- I added page-key scrolling. All 3 page-key subtests failed with `view changed after a page key`.

Both tests passed again on the unchanged task branch.
These checks distinguish page-key behavior from the initial thumb check.
I removed both scratch code copies.

## V4 and V5: live method

I built the task binary with `go build -o .hive/tmp/verify125/hive .`.
The first live pane changed to 90 columns during the check.
A fresh pane had only 119 columns.
I discarded those undersized captures and closed both verifier-owned panes.

I then used a named scratch herdr session, `t125-independent-verifier`.
A verifier-owned terminal process supplied a 160×60 terminal.
The session used herdr 0.9.5, protocol 22.
No live-home pane or Chat input received input.

```sh
clear; TERM=vt100 COLORTERM=truecolor .hive/tmp/verify125/hive --no-startup --demo FILE --size WxH
herdr pane read PANE --source visible --format ansi
herdr pane get PANE
```

For each capture, two CLI reads 500 ms apart were equal.
A socket `pane.read` supplied the `truncated` flag that the CLI omits.
Every saved read has `truncated: false` and scroll offset 0.
The `*.meta.json` files record these checks.

A scratch tool cropped the top-left W×H cells.
It used `cellgrid.Convert` and `cellgrid.Compare`, with no mask.
Each initial board capture has a diff file with zero differing cells.
Each key-state capture has its ANSI, plain text, grid, PNG, and comparison result.
No approved board defines the intermediate key states.
I checked those states against the literal positions and rules below.

A scratch script converted the SGR cells to HTML for Firefox.
Each cell had a fixed width. This keeps glyph advances from moving a border.
The PNG files show the saved terminal cells.

Each demo received two `ctrl+c` presses before the next demo started.
I stopped and removed the scratch session after the checks.
The first scratch process waited during cleanup after the checks finished.
Its outer command timed out. I then checked the stopped server and removed its session folder.
The extra live run used bounded cleanup and exited 0.

### Capture groups

| Capture prefix in `verifier/` | Check |
| --- | --- |
| `TreeWide-120x40-*` | Initial board and `→` across all five children |
| `TreeWide-80x24-*` | Initial view and `→` across all five children |
| `TreeNested-120x40-*`, `NestedKeys-120x40-*` | Initial board, `↑` to manager, and `↓` to nested child |
| `TreeNested-80x24-*` | Initial view, `↑` to manager, and `↓` to nested child |
| `CompactTreeWide-80x24-*` | Initial board, arrows across all children, parent navigation, and page keys |
| `CompactTreeNested-80x24-*` | Initial compact nested board |
| `AgentsNested-120x40-*`, `AgentsCoordinatorChild-120x40-*` | Both approved replacement boards |
| `TallTree-80x24-*` | Three nested levels, vertical scroll in both directions, and boundary keys |
| `PageKeys-child-*`, `PageDelivery-health-*` | Page keys keep child selection unchanged. The same sequence moves Health selection. |

TallTree adds one idle child under `t5-research-1` to the CompactTreeNested fixture.
Its initial selection is `manager-hive-g2`.
The approved nested fixtures fit the height at 120×40, so they require no vertical scroll there.

### Sideways scroll observations

Coordinates and thumb endpoints are 0-based and inclusive.

| Width and selection | Selected card x | Fully visible children | Hidden-child text | Horizontal thumb |
| --- | ---: | --- | --- | --- |
| 120, t5-worker-1 | 5 | 1–4 | `t7-review-1 →` | 1–103 |
| 120, next 3 children | 31, 57, 83 | 1–4 | `t7-review-1 →` | 1–103 |
| 120, t7-review-1 | 95 | 2–5 | `← t5-worker-1` | 13–115 |
| 80, t5-worker-1 | 5 | 1–2 | `t6-worker-1 →` | 1–45 |
| 80, t5-verifier-1 | 28 | 1–3 | `t6-verifier-1 →` | 2–46 |
| 80, t6-worker-1 | 28 | 2–4 | `t7-review-1 →` | 17–61 |
| 80, t6-verifier-1 | 28 | 3–5 | `← t5-verifier-1` | 32–76 |
| 80, t7-review-1 | 52 | 4–5 | `← t6-worker-1` | 34–78 |

The selected card stays complete on the screen after every key.
At 80 columns, the view centers it when the tree bounds permit it.
At 120 columns, the view moves only when the last child needs more room.
The coordinator card and line move sideways with the tree.
A partly visible child counts as hidden.

`←` at the first child and `→` at the last child do not move selection.
`↑` selects the manager. `↓` selects the first child.
`↓` on a leaf does nothing.

### Vertical scroll observations

| TallTree selection | Card x,y | Vertical thumb rows |
| --- | --- | --- |
| manager-hive | 28,11 | 4–13 |
| t5-worker-1 | 10,18 | 4–13 |
| t5-research-1 | 10,18 | 7–16 |
| t5-research-2 | 10,18 | 10–19 |
| t5-research-1, after `↑` | 10,13 | 10–19 |
| t5-worker-1, after `↑` | 10,8 | 10–19 |
| manager-hive, after `↑` | 28,4 | 8–17 |

The selected card stays complete on the screen in every state.
`↓` at the leaf and `↑` at the manager do nothing.
The view does not move back when the selection is already on the screen.

### Page keys and colour-free state check

I sent `ESC[6~` and `ESC[5~` with `herdr pane send-text`.
Each sequence left the tree's ANSI read byte-equal to the prior read.
This held with the manager selected and with a child selected.
On Health › Agents, the same `ESC[6~` moved the selected row.
Thus the live test delivered the page key to hive.

I checked the plain text of all 57 saved tree states after ANSI removal.
Each selected card has `┏`, `┓`, `┗`, and `┛`.
Each visible agent state has a glyph or a word.
Working, idle, and stopped states also have state words in their card text.
No state depends only on colour.

### V5 alignment checklist

| Point | 120×40 | 80×24 | Observation |
| --- | --- | --- | --- |
| Panel, pod, and card borders and corners line up | pass | pass | All 6 board-sized captures equal the grids. Key-state borders stay in their cell columns. |
| Tree lines meet joints and buses without gaps | pass | pass | Manager bus and nested `┬`, `│`, `┴` or `┷` meet. TallTree adds a third level. |
| Coordinator line meets pod border and manager card | pass | pass | `┬`, `│`, `┼`, goal-row `│`, and top-card `┴` meet. Home follows rule A. |
| Clipped borders stop cleanly at the region edge | pass | pass | Wide-tree edge cards and pod borders stop at the panel border. Vertical clipping stops at the top or bottom content edge. |
| Scroll line and thumbs agree with scroll position | pass | pass | Last-row text and inclusive thumb ranges agree with the observations above. |

### Board elements and state without colour

Each row covers that element on all boards where it occurs.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Home title | `hive`, home path | matches | Wide and compact header forms |
| Coordinator and supervisor status | `● working`, `● live` | matches | State words remain in plain text |
| Decision count and clock | `needs you 0`, `14:22` | matches | The fixture uses a fixed clock |
| Screen tabs | `1 Agents` through `6 Logs` | matches | Active tab has an underline |
| Panel title and border | `project agents · hive` or `home`, `╭─╮` | matches | Same cells as each grid |
| Coordinator card | `●`, `working`, card outline | matches | It moves with the tree |
| Coordinator connector | `┬`, `│`, `┼`, `┴` | matches | Decision 65 and home rule A |
| Pod title and outline | `hive` or `home`, box glyphs | matches | Clipping stops at the region edge |
| Goal row | `g2-tui-v2 · wave A` | matches | Text stops before the coordinator line |
| Manager card | `●`, `working`, name | matches | Heavy border when selected |
| First-level bus | `│`, `╭─┴─╮`, `┼` | matches | Every drop meets its child |
| t5-worker-1 card | `● working`, name | matches | Selected or light outline follows selection |
| t5-verifier-1 card | `◐ idle`, name | matches | Sideways centering and state words |
| t6-worker-1 card | `● working`, name | matches | Selected middle child on compact wide board |
| t6-verifier-1 card | `◐ idle`, name | matches | Right-edge hidden-child text |
| t7-review-1 card | `● working`, name | matches | Starts partly hidden on TreeWide |
| t4-verifier-1 card | `○ stopped`, `pass` | matches | Approved AgentsNested board |
| t5-research-1 card | `● working`, name | matches | One tree level below its parent |
| t5-research-2 card | `◐ idle`, name | matches | Third level on AgentsNested and TallTree |
| plan-review-1 card | `● working`, name | matches | Home line reaches its top joint |
| plan-review-2 card | `◐ idle`, name | matches | Heavy selected border below its parent |
| Nested connectors | `┬`, `│`, `┴` or `┷` | matches | No rail in the one-pod view |
| Hidden-stopped text | `h shows 1 stopped` or `h hides 1 stopped` | matches | Literal state words |
| Legend | `spawn`, `working`, `idle`, `stopped`, `━ ▶ message` | matches | Wide form only |
| Scroll row | `←→ at the edge scrolls · children A–B of N` | matches | Last body row |
| Hidden-child name | `NAME →` or `← NAME` | matches | Direction remains visible without colour |
| Horizontal thumb | `━` on `─` | matches | Track and selected position agree |
| Vertical thumb | `┃` on `│` | matches | Height and selected position agree |
| Footer | `[graph]`, arrow hints, action words | matches | Wide or compact form |

## V-gate and cleanup

Decisions 74 to 78 are present in the spec.
Decision 74 replaces the one-pod rail of decision 63.
Decision 78 records all approved grid replacements.
The diff contains only named source files, the brief copy, baseline data, and evidence.
`internal/tui/tui.go` did not change. No mask changed.

The complete five-command gate output is in `verifier/gate.log`.
It records empty formatter and tidy output, zero lint issues, and passing race tests.

I removed the scratch tools, binaries, base and broken-code copies, HTML files, profiles, and board captures.
I closed both panes that I created in the task workspace.
I stopped and removed the named scratch session.
Only final verification evidence remains.
