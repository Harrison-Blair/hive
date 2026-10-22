# Verification — 124-live-agent-activity

Date: 2026-10-22. Verifier: `t124-verifier-2`.
Implementation: `ff154fd0`. Base: `26ff49cd` (`dev` and `origin/dev`).
Verdict: **pass**.

The verifier repeated each item after the fix. The previous failure at
`ab1b7fab` is resolved. Its captures and logs remain in `verifier/`.
The new evidence is in `verifier2/`. Paths below refer to that folder.
The verifier changed no implementation file, worker evidence file, or tracker.

## Items

| Item | Result | Command and observation |
| --- | --- | --- |
| V1 Boards | pass | `cmp` gives no difference for the 4 staged Live boards and the 11 approved replacement boards. Firefox captured each staged board and each spec copy with a new profile. Each pair is pixel-equal. No capture shows `{{`. See `board-copies.log` and `board-pixel-compare.log`. |
| V2 Grids | pass | `go run ./internal/boardexport/cmd/boardexport -out .hive/tmp/verify124b/exported .hive/spec/tui-v2/design`. All 35 exported grids match the committed grids through `cellgrid.Compare`, without masks. All 11 replacement grids match `new-grids/` through `cmp`. The other 20 grids remain equal to `dev`. See `export.log`, `grid-compare.log`, and `unchanged-grids.log`. |
| V3 Tests red and green | pass | The 13 initial `TestLive*` tests fail on `dev` for the missing behavior. The sender test fails because `from` is absent. The exporter test rejects `┰`. Each passes on the task branch. The new grouping test fails on `ab1b7fab` and passes on `ff154fd0`. Repeated runs also pass. See the test table below. |
| V4 Live check | pass | All 15 demo screens match their grids. The requested keys move selection as specified. A scratch inbox record lights its line, which is dark after 6.2 seconds. A scratch `agent send` event has `from: worker-alpha`. Its line lights and is dark after 6.5 seconds. See the live details below. |
| V5 Visual alignment | pass | The four Live screens match their grids at 120×40 and 80×24. The 11 replacement screens also match. The verifier viewed the board and screen PNGs. All 15 pairs are pixel-equal with the same HTML font and cell styling. See `screen-pixel-compare.log`, each `BOARD.png`, and each `BOARD.diff.txt`. |
| V-gate Diff scope and gate | pass | `git merge dev` reports already up to date. The diff stays within the brief and the later board approval. All five gate commands exit 0. `gofmt -l` and `go mod tidy -diff` print nothing. Lint reports `0 issues.`. See `merge-dev.log`, `scope-stat.txt`, and the full output in `gate.log`. |

## V3: repeated tests

The verifier extracted `git archive dev` into a scratch folder.
It copied the new Live tests, sender tests, exporter tests, and Live fixtures there.
A test-only declaration supplied the absent `agentsLitMsg` type.
No feature implementation entered that base run.
The grouping test ran separately against `ab1b7fab` with the current test file.

| Command | Base or broken result | Task branch result | Evidence |
| --- | --- | --- | --- |
| `go test ./internal/tui -run '^TestLive' -count=1 -v` | All 13 initial tests fail on `dev`. | All 14 tests pass, including the grouping test. | `live-red.log`, `live-green.log` |
| `go test ./internal/agent -run '^TestSendLogsTheSender$' -count=1 -v` | The matching-pane case fails. The two absence cases pass. | All three cases pass. | `sender-red.log`, `sender-green.log` |
| `go test ./internal/boardexport -run '^TestCheckAcceptsTheLiveBoards$' -count=1 -v` | The symbol check fails on `dev`. | pass | `export-red.log`, `export-green.log` |
| `go test ./internal/tui -run '^TestLiveMessagesGroupByCard$' -count=20 -v` | All 20 runs fail on `ab1b7fab`. Both cases return two messages. | All 20 runs pass. | `group-red.log`, `group-green.log` |
| `go test ./internal/tui -run '^TestLiveSendEventGivesTheReceiverByKey$' -count=200` | 31 of 200 runs fail on `ab1b7fab`. | All 500 runs pass with `-count=500`. | `receiver-red.log`, `receiver-green.log` |
| `go test -race ./internal/tui -run '^TestLiveSendEventGivesTheReceiverByKey$' -count=200` | The red run above proves the wrong arrow. | All 200 runs pass. | `receiver-race-green.log` |
| `go test -race ./internal/tui -run '^TestLive' -count=50` | The base run above proves the missing behavior. | All 50 runs pass. | `live-race-green.log` |
| `go test ./internal/tui/component -count=1 -v` | Both new joint tests fail with test-only functions that return no glyph or arms. | All component tests pass. | `component-stub-red.log`, `component-green.log` |

The grid comparator also rejects LivePod as an expected LiveOverview grid.
Its differing cells are in `comparator-red.log`.

The fixed-clock tests check t, t+4s, t+5s, and a new message at t+3s.
They check literal glyphs and tokens, including tree paths and lit rails.
They also check selection, the coordinator, the single-project case, merged drops,
unknown senders, and footers at both widths.
The full gate includes the golden checks for all 35 boards.

## V4: executable and keys

Build command: `go build -o .hive/tmp/verify124b/hive .`.
Server: herdr 0.9.5, protocol 22 (`herdr-ping.json`).
Owned demo pane: `wVS:p2`, with 57 visible rows and sufficient width.
Each fixture ran with this command:

```sh
TERM=vt100 COLORTERM=truecolor HIVE_NO_STARTUP=1 .hive/tmp/verify124b/hive --demo FILE --size WxH
```

Each capture used `pane.read`, `source: visible`, `format: ansi`.
Two reads 500 ms apart gave equal text.
Every read has `truncated: false` and scroll offset 0.
The converter crops the top-left board cells and uses `cellgrid.Compare`.
The live comparison ignores only invisible styles where both cells are spaces,
as the spec permits. Model golden checks compare each cell fully.
No mask was added. Metadata is beside each capture.

| Screens | Size | Result |
| --- | --- | --- |
| LiveOverview, LivePod | 120×40 | 0 differing cells each |
| CompactLiveOverview, CompactLivePod | 80×24 | 0 differing cells each |
| Main, AgentsMany, AgentDetails, AgentsHome, AgentsNested, AgentsHiddenParent, AgentsCoordinatorChild, AgentDetailsGeneral, Help | 120×40 | 0 differing cells each |
| CompactAgents, CompactAgentsHome | 80×24 | 0 differing cells each |

The previous live regression fixture also ran 30 times in fresh processes.
All 30 captures match `verifier/reply-key.expected.grid`.
The newest reply gives `▲` at (60,16) and `┃` at (75,18).
See `live-runs.log`, the 30 `reply-key-NN.diff.txt` files,
and the final capture `reply-key-30.ansi` and `.png`.

| Key in LiveOverview | Observation | Result |
| --- | --- | --- |
| `→` | The example-api pod has the heavy border. | pass |
| `↑` | The coordinator card has the heavy border. Its lit joint remains `info`. | pass |
| `↓` | The example-api pod is selected again. | pass |
| `enter` | The example-api one-pod view opens. The footer has `esc pods`. | pass |
| `esc` | The overview opens with example-api still selected. | pass |

Captures: `key1-right.ansi` through `key5-esc.ansi`, with plain text and metadata.
Each demo quit with two `ctrl+c` presses.

## V4: normal mode in a scratch home

Home: `.hive/tmp/verify124b/home`.
Normal TUI pane: `wVS:p3`. Sender pane: `wVS:p4`.
The normal TUI used `HIVE_HOME` for this home and `HIVE_NO_STARTUP=1`.
It ran without a demo fixture. No Chat input was opened.
No test inbox record or test prompt entered the real home.

1. The verifier wrote an inbox record from coordinator to manager-beta.
   The immediate read preceded the next refresh and still showed the dark line.
   At 2.503 seconds, the path is heavy `info` with `▶` toward beta.
   At 6.205 seconds, the path is light `border` again.
2. The verifier ran `agent send` from the pane of the worker-alpha record.
   The event has `from: worker-alpha` and `agent_id: manager-alpha-g1`.
   At 2.493 seconds, the path is heavy `info` with `▲` toward manager-alpha.
   At 6.495 seconds, the path is light again.

`normal-points.log` checks 22 literal glyph/token cells in these captures.
The inbox record and send event are saved as JSON.
The receiver pane held an owned shell, with no harness ready for a prompt.
The send command therefore failed and marked its scratch record unresolved.
The event and line checks pass. Successful harness delivery is not claimed.
See `normal-send-command.log` for the command result.

## V5: visual points

The verifier captured all 15 boards and screens again.
Board PNGs used new Firefox profiles under `.hive/tmp/verify124b/`.
The staged copies used the spec `support.js` in a scratch folder.
Screen PNGs came from the ANSI captures, with the board font and cell styling.
`screen-pixel-compare.log` records pixel equality for all 15 pairs.
The screen PNGs remain in `verifier2/`. The board PNGs were removed after the check.

| Point | 120×40 | 80×24 |
| --- | --- | --- |
| Panel, pod, and card borders and corners line up. | pass | pass |
| Coordinator lines meet the pod and manager joints without gaps. | pass | pass |
| Lit paths are heavy, continuous, and `info`, with one receiver arrow. | pass | pass |
| Selected pods or cards have heavy `accent` borders. | pass | pass |
| Legend and footer text agree with the boards. | pass | pass |

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header | coordinator `● working`, supervisor `● live`, needs-you count | matches | Fixed fixture data. |
| Tab bar | `1 Agents` and `━` underline | matches | Frame cells agree. |
| Graph panel | Rounded border and project title | matches | Overview or one-pod title. |
| Coordinator card | Name, `●`, and state word | matches | Decisions 60 and 65. |
| Coordinator selection | Heavy card border | matches | Key capture `key2-up`. |
| Coordinator bus and pod drops | `┬`, `┴`, `┼`, and mixed lit joints | matches | Decision 65. |
| Selected pod | Heavy `┏━┓┃┗┛` border | matches | Decision 62. |
| Manager and child cards | `● working`, `◐ idle`, `○ stopped` | matches | Glyphs and words remain without colour. |
| Tree and rail | Connected box glyphs | matches | All 15 grid checks pass. |
| Lit path | Heavy line and `◀`, `▼`, or `▲` receiver arrow | matches | Live boards and the repeated reply case. |
| Selected agent | Heavy card border and bold name | matches | One-pod views. |
| Goal and wave | Goal ID and wave word | matches | Fixed fixture data. |
| Stopped count | `h shows` or `h hides` and count | matches | Plain text. |
| Page and scroll marks | Edge text and line glyphs | matches | Compact and many-pod cases. |
| Legend | `━ ▶ message, lit 5s` | matches | Present at 120 columns. |
| Footer | Key glyphs and words | matches | Decision 70. |
| Details popup | Named fields, ACTIONS, and RECENT | matches | Both details boards. |
| Help overlay | Key glyphs and descriptions | overridden | Manager inbox 9 approves the replacement Help board. It matches the new grid. |

The four Live screens also have `-no-colour.png` captures.
The verifier checked their plain text with colour removed.
Each state retains a word, glyph, or border weight.

## Scope and cleanup

Manager inbox 9 approves all 11 replacement boards and both layout exceptions.
Its record remains in `verifier/board-approval.json`.
This approval explains Help and spec decision 73.
Decisions 62 to 72 are present. Decision 14 says that decision 65 replaces it.
The other 20 grids do not change.
`tui.go` schedules line-expiry redraws. `refresh.go` watches send log events.
Both changes fall within the conditional source list.

The verifier closed only its panes `wVS:p2`, `wVS:p3`, and `wVS:p4`.
All three held shells before closure. Only the verifier pane remains in that workspace.
The verifier removed its scratch home, base copies, tools, exporter output,
Firefox profiles, board PNGs, and temporary HTML files.
`cleanup.log` records the check. Final screen captures and logs remain in `verifier2/`.
