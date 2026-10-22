# Independent verification: 123-agents-live-default

- Verifier: `t123-verifier-1` (Pi, `openai-codex/gpt-6.1-sol:xhigh`).
- Date: 2026-10-22.
- Task head checked: `a7fabef151f666b7571edce2bcbb5db1063e8306`.
- Implementation: `6ddf6659`.
- Base: `10a7e9922a9f23285271643ecce08030f1e5ddd0` (`dev` and `origin/dev`).
- `git fetch origin dev` succeeded. `git merge dev --no-edit` reported
  `Already up to date.`
- Verdict: **pass**. No source file, worker evidence file, or tracker was changed
  by the verifier. Paths below are relative to this evidence folder.

## Verification items

| Item | Result | Independent result and evidence |
| --- | --- | --- |
| V1 Boards | pass | All 11 `cmp` commands exit 0. Fresh Firefox captures of each staged board and its design copy have equal SHA256 values and zero differing pixels. No capture shows `{{`. See `verifier/boards.txt`. |
| V2 Grid diff | pass | `cellgrid.Compare` compared all 31 grids with `git show dev:PATH`, with no mask. Exactly the eight expected grids change. All 31 grids also equal a fresh exporter run. See `verifier/grids.txt` and `verifier/grid-scope.txt`. |
| V3 Tests red and green | pass | All three new tests fail on the base for the expected reason and pass on the task code. Changed behavior tests and golden checks also fail on the base and pass on the task code. Health and demo compatibility checks pass on both. See `verifier/red.txt` and `verifier/green.txt`. |
| V4 Live check | pass | All 11 board-size captures equal their grids, including glyph, token, and bold. First `h` shows the stopped card and `h hides 1 stopped`. Second `h` restores all 4,800 cells of AgentsHiddenParent. See the capture files and `verifier/live-metadata.json`. |
| V5 Visual alignment | pass | The compiled TUI ran with each of the five specified fixtures at both sizes. All board-size comparisons pass. The visual checks and the visible border, joint, external rail, and plain-text state checks pass. See `verifier/alignment.txt`, the PNG captures, and the tables below. |
| V-gate Diff scope and gate | pass | Changes stay within the source table and task evidence. Decisions 58–61 have the brief's meaning. All five gate commands exit 0. `gofmt -l` and `go mod tidy -diff` print nothing. Lint reports `0 issues.` See `verifier/gate.txt` for the full output. |

## V1 commands and observations

```sh
S=/home/penguin/source/hive/.hive/tmp/g4-agents/boards
for f in "$S"/*.dc.html; do
  cmp "$f" ".hive/spec/tui-v2/design/${f##*/}"
done
firefox --headless --no-remote --profile SCRATCH --window-size=1220,780 --screenshot OUT.png file://BOARD.dc.html
```

The staged folder has no `support.js`. For capture only, the verifier copied
its unchanged board bytes and the unchanged design `support.js` to a scratch
folder. Each Firefox invocation used a new profile folder. Each staged capture
was compared with its design-copy capture with `compare -metric AE`; each
result was `0 (0)`. The verifier inspected a montage of all 11 rendered boards.
The profiles and board captures were temporary and were removed after review.

## V2 changed rows

Rows are zero-based. Each non-text change preserves the glyph and bold flag
and changes `border` to `working` on a working card border or its joint.
The footer changes are confined to row 39. The Help change is on row 13.

| Grid | Changed cells | Changed rows |
| --- | ---: | --- |
| Main | 20 | 39 |
| AgentsMany | 228 | 7–10, 12–15, 22–25, 39 |
| AgentsHome | 280 | 7–10, 12–15, 17–20, 22–25, 39 |
| AgentsNested | 332 | 7–10, 14–20, 22–25, 32–35, 39 |
| AgentsHiddenParent | 280 | 7–10, 12–15, 17–20, 39 |
| AgentsCoordinatorChild | 72 | 7–10, 39 |
| CompactAgentsHome | 104 | 7–10, 12–15 |
| Help | 24 | 13 |

CompactAgents, AgentDetails, AgentDetailsGeneral, and all other grids have no
changed cell. The exporter command was:

```sh
go run ./internal/boardexport/cmd/boardexport -out .hive/tmp/verify-123/export .hive/spec/tui-v2/design
```

## V3 red and green commands

The verifier exported `dev` to a scratch folder with `git archive dev`. It
copied only the task's four changed test files and design files into that
base tree. Production code in that tree remained the base code. This isolated
the red run from the task implementation. Both runs used this command from
the corresponding root:

```sh
go test ./internal/tui/... -count=1 -v -run 'Test(AgentsHideStoppedAtStart|AgentsWorkingCardBorder|CardWorkingBorder|AgentsBoardsMatchTheirGrids|GeneralAgentBoardsMatchTheirGrids|FrameRowsMatchMain|FrameRowsMatchEachScreenBoard|HelpMatchesItsBoard|AgentsBoardKeyHintsMatchTheBoard|AgentsNormalModeLoadsAndRefreshes|AgentsGraphShowsTheCurrentGenerationOnly|AgentsKeysLogTheirActions|HealthAgentsHTogglesStoppedAgents|AgentsHHidesStoppedAgents|CardSelectedHeavyBorder|KeyHintsGolden)$'
```

| New test | Base result | Task result |
| --- | --- | --- |
| TestAgentsHideStoppedAtStart | fail: missing `h shows 2 stopped`; stopped cards are shown | pass |
| TestAgentsWorkingCardBorder | fail: card corners and joints use `border`, expected `working` | pass |
| TestCardWorkingBorder | fail: working border cells use `border`, expected `working` | pass |

The red command exits 1; the green command exits 0. The Health stopped-row
check, fixture-mode stopped visibility check, and selected heavy border check
pass on both versions, as expected for preserved behavior. The full race gate
also runs the other screen and board tests.

## V4 live commands and results

The verifier built the task code with:

```sh
go build -o .hive/tmp/verify-123/hive .
```

It created only its own sibling pane, `wVN:p2`, terminal
`term_65e74be17c6841a`, in workspace `wVN`. The pane was 145×57. Each demo used:

```sh
TERM=vt100 COLORTERM=truecolor .hive/tmp/verify-123/hive --no-startup --demo .hive/spec/tui-v2/design/BOARD.fixture.json --size WxH
```

`TERM=vt100` disables the terminal erase-character optimisation for these
captures. Reads used `pane.read`, `source: visible`, `format: ansi`, with ANSI
preserved. The socket read was needed to retain the `truncated` field, which
the CLI text output omits. Server and schema protocol were both 22, herdr
0.9.5. Two reads 500 ms apart had equal text. Every accepted read had
`truncated: false` and scroll offset 0. Metadata is retained in
`verifier/live-metadata.json`.

A scratch Go tool used `cellgrid.Convert`, cropped the top-left W×H cells,
and used `cellgrid.Compare` without masks. PNGs were captured from an HTML
`pre` made from the captured ANSI cells. `.grid` and `.txt` files contain the
cropped screen; `.ansi` files retain the pane styling on the first H rows.

| Board | Size | Result |
| --- | --- | --- |
| Main | 120×40 | 4,800 cells match |
| CompactAgents | 80×24 | 1,920 cells match |
| AgentsMany | 120×40 | 4,800 cells match |
| AgentsHome | 120×40 | 4,800 cells match |
| AgentsNested | 120×40 | 4,800 cells match |
| AgentsHiddenParent | 120×40 | 4,800 cells match |
| AgentsCoordinatorChild | 120×40 | 4,800 cells match |
| CompactAgentsHome | 80×24 | 1,920 cells match |
| AgentDetails | 120×40 | 4,800 cells match |
| AgentDetailsGeneral | 120×40 | 4,800 cells match |
| Help | 120×40 | 4,800 cells match |

Each row has `verifier/BOARD-WxH.ansi`, `.grid`, `.txt`, `.png`, and `.diff.txt`.
The first `h` capture is `AgentsHiddenParent-h1-120x40`: it shows
`○ t4-verifier-1`, its child, and `h hides 1 stopped`. The second is
`AgentsHiddenParent-h2-120x40`: all 4,800 cells match the original grid.
Each demo quit with two `ctrl+c` presses. No Chat input was opened.

Negative control: comparing the Main live capture with AgentsHome reports
1,337 differing cells. See `verifier/live-negative-control.txt`.

## V5 visual alignment and state without colour

These checks used the real compiled TUI with the approved demo fixtures, not
only model output. Normal-mode start behavior was checked separately by the
V3 model test. The verifier inspected the board montage, the screen montage,
and full-size default, rail, nested-tree, and compact screenshots.

| State | 120×40 capture | 80×24 capture |
| --- | --- | --- |
| Default | AgentsHiddenParent-120x40 | AgentsHiddenParent-80x24 |
| Stopped shown | Main-120x40 | Main-80x24 |
| Rail | AgentsHome-120x40 | AgentsHome-80x24 |
| Tree and nested rails | AgentsNested-120x40 | AgentsNested-80x24 |
| Home page | CompactAgentsHome-120x40 | CompactAgentsHome-80x24 |

At its approved board size, each capture equals its board grid. Main at
80×24 also equals CompactAgents.grid. The other four alternate-size states
have no approved grid; their `.diff.txt` files state this limit. Their visual
and cell-property checks pass. Browser page margins and raster glyph rendering
are capture presentation differences; the terminal glyph, token, and bold
comparisons use the grids.

The independent cell check covers 56 cards with a visible top across the ten
V5 screens. It checks every visible perimeter cell and joint against literal
`working`, `accent`, or `border` expectations from the card's state and heavy
selection glyphs. Adjacent external rails and tree lines keep `border`.
Negative control: changing working tokens to border produces 260 failed
border checks in AgentsHome.

| Point | 120×40 | 80×24 |
| --- | --- | --- |
| Panel borders, card borders, and corners align | pass | pass |
| Working card perimeter and joints use working | pass | pass |
| Selected cards keep heavy accent borders | pass | pass |
| External rails and tree lines keep border and meet joints | pass | pass |
| Visible pod note and footer match the board form | pass | pass |

The compact home board clips the idle card's bottom row at the viewport edge.
The same clipping occurs in the alternate compact home-page states. This
agrees with the approved compact grid. Nested compact mode scrolls to keep
its selected child visible. A pod note can be outside that viewport. The
compact footer ends with `?`, as on CompactAgents. These are the existing
responsive board forms, not new golden grids or masks.

Plain-text captures remove colour. The following elements retain a glyph or
word. State does not depend on colour alone.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header | `coordinator`, `working`, `supervisor`, `live`, `needs you`; compact `coord`, `sup`, `●` | matches | Existing full and compact frame |
| Active tab | `1 Agents` and heavy underline | matches | Existing frame |
| Graph panel | `project agents` and project count | matches | Existing focused panel |
| Project and home pods | Project name, `home`, goal and wave text | matches | Existing graph groups |
| Working card | `●` and `working` | matches | Decision 60: rounded working border |
| Selected working card | `●`, `working`, and heavy `┏━┓┃┗┛` border | matches | Selection keeps accent |
| Idle card | `◐` and `idle` | matches | Border stays border |
| Stopped card | `○` and `stopped` | matches | Border stays border |
| Selected idle card | `◐`, `idle`, and heavy border | matches | Selection keeps accent |
| Working border joints | `├`, `┤`, `┬`, `┴`, `╰`; card state word | matches | Decision 60: card token at joint |
| External rails and tree lines | `│`, `├`, `╰`, `─` | matches | Border token outside card |
| Hidden stopped note | `h shows 1 stopped` | matches | Decision 58 |
| Shown stopped note | `h hides N stopped` | matches | Decision 58 |
| Legend | `spawn`, `● working`, `◐ idle`, `○ stopped` | matches | Existing wide graph legend |
| Scroll state | `←→ at the edge scrolls`, project range, `┃`, `━` | matches | Existing responsive viewport |
| Footer | `[graph]`, `h stopped`; compact action words | matches | Decision 61 and existing compact footer |
| Details popups | Agent name, role, status word, `ACTIONS`, `RECENT`, `esc closes` | matches | Existing popup boards preserved |
| Help | `h show or hide stopped agents` | matches | Decision 61 |

## Diff scope, gate, and cleanup

The diff contains only the approved boards, expected grids, decisions, brief,
STE baseline, named TUI source and tests, and task evidence. `tui.go` must change
to set the normal-mode initial graph default. `component/panel.go` must change
to accept the working edge colour while keeping focused accent precedence.
Health source, fixtures, the exporter, and masks do not change.

The five exact gate commands were:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go mod tidy -diff
go vet ./...
golangci-lint run ./...
go test -race ./...
```

Full output and exit statuses are in `verifier/gate.txt`. All exit statuses
are 0. The first two commands have no output.

The verifier closed only its own pane after it returned to its original bash
process. The scratch base tree, binaries, conversion scripts, HTML files,
Firefox profiles, and temporary board PNGs were removed. Final live captures,
diffs, logs, and this report stay in the task evidence folder.
