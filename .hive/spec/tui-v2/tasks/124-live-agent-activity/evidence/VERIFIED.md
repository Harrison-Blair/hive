# Verification — 124-live-agent-activity

Date: 2026-10-22. Agent: `t124-verifier-1`.
Implementation: `e934afdf`. Base: `26ff49cd` (`dev`).
Verdict: **fail**. V3 has an intermittent arrow-direction failure.

All paths below are relative to this evidence folder. The verifier did not
change an implementation file, a worker evidence file, or the tracker.

## Items

| Item | Result | Command and observation |
| --- | --- | --- |
| V1 Boards | pass | `cmp` of all 4 Live boards and the 11 approved replacement boards returns 0. Both sources of each board were captured with Firefox. The staged source used the spec copy of `support.js` in a scratch folder. No PNG shows `{{`. See `verifier/board-copies.log` and `verifier/board-pixel-compare.log`. |
| V2 Grids | pass | `go run ./internal/boardexport/cmd/boardexport -out .hive/tmp/verify124/exported .hive/spec/tui-v2/design`. All 35 exported grids match the committed grids with `cellgrid.Compare`, with no masks. Each of the 11 replacement grids also matches `new-grids/` by `cmp`. The other 20 grids are unchanged from `dev`. See `verifier/grid-compare.log` and `verifier/unchanged-grids.log`. |
| V3 Tests red and green | fail | All 13 `TestLive*` tests fail on the base. The sender test fails because `from` is absent. The exporter test rejects `┰`. Each passes in the first task-branch run. But `TestLiveSendEventGivesTheReceiverByKey` fails in 13 of 100 repeats, and in 9 of 100 repeats with `-race`. See the finding below and `verifier/*-red.log`, `verifier/*-green.log`, and `verifier/receiver-key*-repeat.log`. |
| V4 Live check | pass | The executable ran all 15 fixtures with `COLORTERM=truecolor`, `--demo FILE`, and `--size WxH`. Each capture matches its golden grid. The requested keys move selection as specified. A scratch inbox record lights its line after the next refresh and the line is dark at 6.5 seconds. A scratch `agent send` event has `from: worker-alpha`, and its line lights and goes dark. See the live details below. The additional live reproduction of the V3 failure is recorded separately. |
| V5 Visual alignment | pass | The four Live captures match their grids at 120×40 and 80×24. Each required point passes in these board states. The 11 replacement screen captures also match. The verifier viewed the board and terminal PNGs side by side. See the tables below and `verifier/BOARD.png`, `.ansi`, `.diff.txt`, and `.meta.json`. |
| V-gate Scope and gate | pass | `git merge dev` reports already up to date. `git diff dev...HEAD --stat` was checked. Changes are within the brief and the later board approval. The five gate commands exit 0. `gofmt -l` and `go mod tidy -diff` print nothing; lint reports `0 issues.`. Full output: `verifier/gate.log`. This single gate pass does not remove the repeated test failure in V3. |

## Finding: an older message can reverse the newest message arrow

**AC7, AC10, and decision 67.** The newest message on a line must give the
arrow. A record key and an agent name can name the same card.

Reproduction on the unmodified task branch:

```sh
go test ./internal/tui -run '^TestLiveSendEventGivesTheReceiverByKey$' -count=100
go test -race ./internal/tui -run '^TestLiveSendEventGivesTheReceiverByKey$' -count=100
```

Both commands exit 1. This is the worker's existing test.

- The fixture has an older message from `manager-hive` to `t6-worker-1`.
- The newer send event is from `t6-worker-1` to `manager-hive-g2`.
- Both destinations name the same manager card.
- Expected: `▲` in `info` at column 60, row 16; `┃` in `info` at column 75,
  row 18. Coordinates are zero-based.
- Intermittent result: `┃` at (60,16) and `▼` at (75,18). The arrow points
  to the worker, which was the receiver of the older message.

Cause: `internal/tui/agents_screen.go:1161` (`litMessages`) groups messages
by their raw sender and receiver strings. It does not first resolve these
strings to card keys. Thus these two messages remain in separate map
entries. `lightLines` at line 1285 draws both paths in map iteration order.
`graphKey` resolves their names later, after the newest-message choice.
The older path can be drawn last.

The verifier also reproduced the fault in the actual executable. Fixture:
`verifier/reply-key.fixture.json`. It adds the newer event at
`14:21:59.5Z` to LivePod. The fixed clock is `14:22:00Z`.

```sh
COLORTERM=truecolor .hive/tmp/verify124/hive --demo .hive/spec/tui-v2/tasks/124-live-agent-activity/evidence/verifier/reply-key.fixture.json --size 120x40
```

The seventh fresh run showed the wrong arrow. Evidence:
`verifier/reply-key-live.ansi`, `.png`, `.meta.json`, and `.diff.txt`.
The expected grid is `verifier/reply-key.expected.grid`: LivePod with only
the two literal arrow cells above changed. The live comparison reports
exactly those two differing cells. No implementation fix was made.

## Red runs and comparison checks

The verifier extracted `git archive dev` into a scratch folder. The new
live test file and the updated sender and exporter test files were copied
there. The four Live fixtures and grids were also copied there. A test-only
`type agentsLitMsg struct{}` declaration let the base compile. No feature
code was added to the base. All 13 live tests fail for missing behavior.
The no-record and outside-herdr sender subtests already pass on the base;
the matching-pane subtest fails, as expected.

Commands: `go test ./internal/tui -run '^TestLive' -count=1 -v`,
`go test ./internal/agent -run '^TestSendLogsTheSender$' -count=1 -v`, and
`go test ./internal/boardexport -run '^TestCheckAcceptsTheLiveBoards$' -count=1 -v`.
The same commands were run on the task branch. Component joint tests also
pass (`verifier/component-green.log`). The grid comparator was checked
against a deliberately wrong grid; it reports differences
(`verifier/comparator-red.log`).

## Live details

The first owned pane in the normal session had only 119 columns. Its reads
were rejected as evidence. The verifier then made the isolated herdr
session `verify124-scratch`, with an owned 180×60 client PTY. The demo pane
had enough space for each board. Server: herdr 0.9.5, protocol 22.

Each demo read uses `pane.read`, `source: visible`, `format: ansi`. Two reads
500 ms apart were equal. Each read has `truncated: false` and scroll offset
0. The converter crops the top-left board cells and compares with
`cellgrid.Compare`. Only invisible styles of space cells are ignored, as
the spec permits. All non-space cells retain glyph, token, and bold checks.
No mask was added. Metadata is beside each capture.

The four new boards, all 10 AC11 boards, and Help match their grids.
`verifier/capture-results.json` lists each result. Each demo quit with two
`ctrl+c` presses.

| Key in LiveOverview | Observed result |
| --- | --- |
| `→` | example-api pod has the heavy border |
| `↑` | coordinator card has the heavy border |
| `↓` | example-api pod is selected again |
| `enter` | one-pod view of example-api; footer has `esc pods` |
| `esc` | overview; example-api is still selected |

Captures: `verifier/key1-right.ansi` through `verifier/key5-esc.ansi`, with
plain-text copies and metadata.

Normal mode used `HIVE_HOME=.hive/tmp/verify124/home` and
`HIVE_NO_STARTUP=1`. No inbox record or test prompt was sent in the real
home. A new inbox record was not yet visible in a read before the next
2-second refresh. It was lit at 2.5 and 3.5 seconds, and dark at 6.5 seconds.
The send event was lit at 1.5 seconds and dark after another 6 seconds.
`verifier/normal-points.log` checks 20 literal glyph/token cells in these
four captures.

The scratch send command returned `agent_not_ready`: the receiver was an
owned input-loop process, with no live agent harness. No prompt was
delivered. The log event and the lit line still prove the specified AC9
field and event-loader behavior. Successful harness delivery is not
claimed. See `verifier/normal-send-command.log` and
`verifier/normal-send-event.json`.

## Visual points

These results apply to the approved board states. The reply-key regression
above fails the arrow rule in an additional message state.

| Point | 120×40 | 80×24 |
| --- | --- | --- |
| Panel, pod, card borders, and corners line up | pass | pass |
| Coordinator line meets the pod and manager joints without a gap | pass | pass |
| Lit path is heavy, `info`, continuous, and has one receiver arrow | pass | pass |
| Selected pod or card has a heavy `accent` border | pass | pass |
| Legend and footer agree with the board | pass | pass |

| Board element | State without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header | coordinator `● working`; supervisor `● live`; needs-you count | matches | fixed fixture data |
| Tab bar and active rule | `1 Agents`, underline `━` | matches | frame cells match |
| Graph panel | title, rounded border | matches | overview or pod title |
| Coordinator card | name and `●`; heavy border when selected | matches | decisions 60 and 65 |
| Coordinator bus and pod drops | `┬`, `┴`, `┼`; mixed joints when lit | matches | decision 65 |
| Pod selection | heavy `┏━┓┃┗┛` border | matches | decision 62 |
| Agent cards | `● working`, `◐ idle`, `○ stopped` | matches | glyph and state word remain |
| Tree and rail | connecting box glyphs | matches | all 15 grid comparisons pass |
| Lit path | heavy `━┃` and `◀` or `▼` | matches | four Live board states |
| Message legend | `━ ▶ message, lit 5s` | matches | present at 120; absent at 80 |
| Stopped count | `h shows` or `h hides` and count | matches | plain text |
| Page and scroll marks | edge text; `━` and `┃` | matches | compact and many-pod cases |
| Footer | key words and arrow glyphs | matches | decision 70 |
| Details popup | title, field names, ACTIONS, RECENT | matches | both details boards |
| Help overlay | key words and descriptions | overridden | manager inbox 9 approves the replacement Help board |

The verifier checked plain-text captures with colour removed. The four
Live PNGs also have `-no-colour.png` copies. Each state retains its glyph,
word, or border weight. Colour is not its only signal.

## Scope and cleanup

Manager inbox 9 (`verifier/board-approval.json`) approves the 11 replacement
boards, including Help, and both layout exceptions. This decision explains
the extra Help files and README decision 73. The other 20 grids do not
change. `tui.go` schedules line-expiry redraws. `refresh.go` watches send log
events. These edits are within the conditional source list.

The verifier removed the scratch home, base copy, exporter output, tools,
board PNGs, and Firefox profiles after recording the result. All owned demo
processes, the input loop, the scratch client and server, and the temporary
normal-session tab were closed. Cleanup was checked. Final captures and
logs remain in `verifier/`.
