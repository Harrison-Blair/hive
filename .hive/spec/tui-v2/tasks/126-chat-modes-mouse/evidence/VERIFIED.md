# Verification: task 126-chat-modes-mouse

Date: 2026-10-23. Verifier: `t126-verifier-1`.

Worker commit: `fb01e40b`. Verification commit before this report:
`e0a7b834`, which merges `dev` at `fa3ebf34`. The merge adds only the existing
dogfood record from `dev`. The verifier changed no task source file.

**Verdict: fail.** The Logs filter blocks mouse input. V3 also has one
literal expected-output difference for its regression guard.

## Results

| Item | Result | Observation and evidence |
| --- | --- | --- |
| V1 Boards | pass | The 10 latest approved canvas copies have `cmp` exit 0. Their staged and design PNGs have no pixel difference. The original three v88 copies are superseded by the approval in decisions 89 and 90. See `verifier/boards.log`. |
| V2 Grids | pass | The exporter exits 0. `cellgrid.ReadGrid` and `cellgrid.Compare`, with no mask, find 0 differences for all 42 exported grids against design. All 41 `new-grids` files equal design. The base comparison confirms the permitted changed rows. See `verifier/grids.log`. |
| V3 Tests red and green | fail | On base `f7990406`, 16 of 17 new tests fail for their behavior. The Agents-wheel guard passes on the base, which ignores all mouse input. V3 expects each new test to fail on the base. The guard does fail on a deliberate mutation. All 17 tests pass on the task branch with `-count=50` and with `-race -count=50`. See `verifier/red-base.log`, `red-repeat.log`, `guard-red-green.log`, and `green.log`. |
| V4 Live check | fail | All 40 screen captures equal their grids. The listed keys and mouse events work in their initial demo states. Extra checks find that a wheel event and a tab click do nothing while the Logs filter is open. See the failure below and `verifier/keys/`. |
| V5 Visual alignment | pass | The four mode/size captures have 0 differing cells. Borders, header count, feed note, thumb, placeholder, and cursor match. PNGs were inspected. See the points table and `verifier/live/`. |
| V-gate Scope and gate | pass | After the merge of `dev`, the diff is within the brief, AC6 board updates, the required brief copy, and evidence. All five gate commands exit 0. Formatting and tidy print nothing. See `verifier/diff-stat.txt` and the full output in `verifier/gate.log`. |

## Failure: mouse input in the Logs filter

Source: `internal/tui/mouse.go:20`.

`updateMouse` returns before it handles an event when `m.mode == editing`.
The Logs filter uses this mode. AC3 requires the wheel to scroll the region
under the pointer and a click on a tab name to open that tab. The brief and
decisions 79–81 give no filter exception.

Reproduction in the task binary, in the verifier's own pane `wW2:p2`:

```sh
TERM=vt100 COLORTERM=truecolor HIVE_BINARY --no-startup \
  --demo .hive/spec/tui-v2/design/Logs.fixture.json --size 120x40
herdr pane send-keys wW2:p2 esc /
herdr pane read wW2:p2 --source visible --format ansi
herdr pane send-text wW2:p2 $'\e[<65;5;20M'
herdr pane read wW2:p2 --source visible --format ansi
herdr pane send-text wW2:p2 $'\e[<0;22;2M\e[<0;22;2m'
herdr pane read wW2:p2 --source visible --format ansi
```

| Event | Expected | Actual | Result |
| --- | --- | --- | --- |
| Wheel down over the event list, one-based cell 5,20 | Selection moves from screen row 30 to row 31. Filter focus stays. | Selection stays on row 30. The full ANSI capture is byte-equal to the pre-event capture. | fail |
| Click on `3 Work`, one-based cell 22,2 | Work opens. | Logs stays open. The footer stays `[filter]`. The full ANSI capture is byte-equal to the pre-event capture. | fail |

Evidence: `verifier/keys/filter-00-editing`, `filter-01-wheel-events`, and
`filter-02-click-work`, each with `.ansi`, `.txt`, and `.png` files. The
comparison and expected results are in `verifier/keys/filter-mouse.diff.txt`.

## V1 and V2 method

The plan `/home/penguin/source/hive/plans/hive-ask.md`, the brief, the spec,
and the worker evidence were read. The latest approved Chat boards are in
`/home/penguin/source/hive/.hive/tmp/g6/approved/`; Help and Main are in
`/home/penguin/source/hive/.hive/tmp/g6/approved-ok/`. These sources are
recorded in decision 90 and the worker's V1 evidence.

Each of these 10 boards was captured from both copies with a new Firefox
profile and the skill command, at 1220×780. A scratch copy of each staged
board had the design `support.js` beside it. PIL compared the two PNGs.
Each pair was pixel-equal. All 42 design boards were also captured as PNGs.
The captures were inspected, including a contact sheet. No rendered board
showed `{{`. Board PNGs and profiles were removed after the check.

Exporter command:

```sh
mkdir -p .hive/tmp/t126-verifier/export
go run ./internal/boardexport/cmd/boardexport \
  -out .hive/tmp/t126-verifier/export .hive/spec/tui-v2/design
```

It printed nothing and exited 0. A first attempt omitted the output folder
and failed to open an output file. The folder was then created and the
exact exporter command was run again.

The scratch comparison used `internal/lib/cellgrid` with no mask. Outside
Chat and Help, each existing screen changes only in row 0. The component
sheet changes only in row 7, its header sample. `TooSmall` has 0 changes.
The three new grids and the 38 updated grids are tracked. No other grid
changes. The detailed row results are in `verifier/grids.log`.

## V3 method and limit

A detached scratch worktree at `f7990406` received the new test files and
three new board sets. Two inert fields, `Header.Asks` and `headerState.asks`,
were added there so the tests could compile. No feature code was added.
Each new test ran on that base with `-count=50` and with
`-race -count=50`. The failing assertions are recorded in `red-base.log`;
the repeated results are in `red-repeat.log`.

The same tests ran on the task branch in these packages:

```sh
go test ./internal/tui ./internal/tui/component ./internal/boardexport \
  -run NEW_TESTS -count=50
go test -race ./internal/tui ./internal/tui/component ./internal/boardexport \
  -run NEW_TESTS -count=50
```

`NEW_TESTS` is the exact 17-test regular expression in `red-repeat.log`.
Both task runs exit 0 (`green.log`).

`TestMouseWheelOverTheAgentsGraphDoesNothing` passes on the base. To check
its strength, the scratch worktree was switched to the task commit. Its
Agents-graph wheel path was changed to send `KeyDown`. The guard failed
with `the wheel changed the Agents graph`. After restoration it passed
50 times. No code fault is inferred from this guard. Its base result does
not meet V3's literal expected output. The guard must be retained.

## V4 live observations

A fresh task binary ran in the verifier's own herdr 0.9.5 pane `wW2:p2`.
The pane was 207×57, which is larger than each board. Each run used:

```sh
clear; TERM=vt100 COLORTERM=truecolor HIVE_BINARY --no-startup \
  --demo .hive/spec/tui-v2/design/BOARD.fixture.json --size WxH
```

Each accepted read used `source: visible`, `format: ansi`, had
`truncated: false`, and had scroll offset 0. Two reads 500 ms apart were
equal. The comparison cropped the top-left W×H cells, converted them with
`cellgrid.Convert`, and called `cellgrid.Compare` with no mask. Space-cell
style was ignored only for live captures, as the spec permits. Each other
cell was compared fully.

All 40 screen boards have their own `.ansi`, `.diff.txt`, and `.png` files
in `verifier/live/`. Each diff reports 0 cells. `Components` has no screen;
V2 and the model gate check it. `TooSmall` is unchanged. As a negative
control, the ChatNormal capture was compared against ChatInput: 602 cells
differ (`verifier/mismatch-control.diff.txt`).

| Input or event | Observation | Result |
| --- | --- | --- |
| `i`, text `verify`, at 120×40 | `[input]`; input shows `> verify█`. | pass |
| `enter` | Demo reports `demo mode: no action`; mode stays `[input]`. No chat is sent. | pass |
| `esc`, then `tab` | Mode stays `[input]`; draft and cursor stay. | pass |
| `ctrl+]` | `[normal]`; draft stays and cursor leaves. | pass |
| `j`, then `k` | Feed moves; input draft stays `verify`. | pass |
| Wheel up, then down over the feed | Earlier note moves to 29, then 31; draft and mode stay. | pass |
| Click input, then feed | Mode changes to `[input]`, then `[normal]`. | pass |
| Click `6 Logs` | Logs opens, with `[events]` in the footer. | pass |
| Compact `i`, text `ok`, `esc`, `tab`, `ctrl+]` | Input starts, draft stays, esc and tab keep input, ctrl+] leaves. | pass |
| Compact feed wheel and input/feed clicks | Feed note moves 42 to 41; clicks change the mode. | pass |
| Compact click `4 Friction` | Friction opens with `[list]`. | pass |
| Logs list wheel down/up, with JSON focus | Selected event moves row 30 to 31 to 30. Focus stays `[json]`. | pass |
| Logs wheel over short JSON | No change; the content fits. | pass |
| Logs click on list row 14 | Row 14 is selected; focus becomes `[events]`. | pass |
| Agents graph wheel and click | Both reads equal the initial read. Decision 77 applies. | pass |
| Friction click on row 7 | `agent-can-send-user-chat` is selected. | pass |
| Friction click details, then wheel over list | Focus becomes `[details]`; wheel moves list selection to row 8 and keeps that focus. | pass |
| Friction details wheel | Details scroll; selection and focus stay. | pass |
| Logs filter wheel and tab click | Both are ignored, as detailed above. | fail |

The ChatNormal fixture starts past the end of its feed. Its first scroll
clamps the view to the end, so the note changes from 42 to 31. Later
scrolls move by lines. This is the existing `scrollChat` behavior.

Logical keys and raw SGR bytes used `pane.send_keys` and `pane.send_text`.
An initial harness attempt used `pane.send_input.text`, which herdr treats
as a paste. Those events were ignored. The full event sequence was then
run again with the correct methods, and the final captures replace that
attempt. Raw wheel bytes were `ESC[<64;X;YM` and `ESC[<65;X;YM`; left-click
bytes were `ESC[<0;X;YM ESC[<0;X;Ym`. Coordinates count from 1.

A physical shift-and-drag terminal selection was not exercised. The model
check confirms cell-motion mouse reports. The live send check uses the
safe demo block; the model tests check that enter starts the send command.

## V5 points

Each named capture is in `verifier/live/`, with its PNG and cell diff.
Coordinates count from 0. Each result below was checked against the board
and the capture. Each complete capture has 0 differing cells.

| Point | ChatNormal 120×40 | ChatInput 120×40 | CompactChatNormal 80×24 | CompactChat 80×24 | Result |
| --- | --- | --- | --- | --- | --- |
| Panel borders and corners | aligned | aligned | aligned | aligned | pass |
| Focused border | feed, accent | input, accent | feed, accent | input, accent | pass |
| Other border | input, border | feed, border | input, border | feed, border | pass |
| Footer mode | `[normal]` | `[input]` | `[normal]` | `[input]` | pass |
| Header count | `? 0 asks`, col 89 | `? 0 asks`, col 89 | `? 0`, col 54 | `? 0`, col 54 | pass |
| Feed note text | `↑ 42 earlier`, col 61 | `↑ 42 earlier`, col 61 | `↑ 42 earlier`, col 41 | `↑ 44 earlier`, col 41 | pass |
| Thumb `┃` | col 119, rows 25–34 | col 119, rows 25–34 | col 79, row 18 | col 79, row 18 | pass |
| Placeholder | row 37, col 2 | absent | row 21, col 2 | absent | pass |
| Cursor | absent | row 37, col 33 | absent | row 21, col 4 | pass |

## Board elements and state without colour

The rows cover each changed element of the three new boards. The existing
Chat boards use the same elements; each full grid comparison passes.
Other screen boards change only their header, except Help's CHAT line.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| ChatNormal header | `? 0 asks`, `● working`, `● live`, `needs you 0` | matches | Latest approved board; decision 87. |
| ChatNormal tabs | `2 Chat` and `━━━━━━━━` | matches | Active-tab underline. |
| ChatNormal feed title and border | `chat · noise normal`, `↑ 42 earlier`, box corners | matches | Feed has focus; footer gives the mode word. |
| ChatNormal feed body and thumb | actor names, event `▸`, right-side `┃` | matches | Exact glyphs and tokens. |
| ChatNormal input title and body | `message`, `● working 1m`, `i or click here to type` | matches | Unfocused input has a text instruction. |
| ChatNormal footer | `[normal]` and key words | matches | Explicit mode word. |
| ChatInput header | `? 0 asks`, `● working`, `● live`, `needs you 0` | matches | Latest approved board. |
| ChatInput tabs | `2 Chat` and `━━━━━━━━` | matches | Active-tab underline. |
| ChatInput feed title and border | `chat · noise normal`, `↑ 42 earlier`, box corners | matches | Unfocused feed. |
| ChatInput feed body and thumb | actor names, event `▸`, right-side `┃` | matches | Exact glyphs and tokens. |
| ChatInput input title and body | `message`, `● working 1m`, `> Sounds good, go ahead with g6█` | matches | Prompt and cursor show input focus. |
| ChatInput footer | `[input]` and key words | matches | Explicit mode word. |
| CompactChatNormal header | `? 0`, `needs you 0`, labelled status dots | matches | Compact header. |
| CompactChatNormal tabs | full `4 Friction`, `2 Chat` underline | matches | Existing tab-label rule. |
| CompactChatNormal feed title and border | `chat`, `↑ 42 earlier`, box corners | matches | Compact title rule. |
| CompactChatNormal feed body and thumb | actor names on separate lines, `┃` | matches | Compact feed rules. |
| CompactChatNormal input title and body | `message`, `● working 1m`, placeholder words | matches | Explicit input instruction. |
| CompactChatNormal footer | `[normal]` and key words | matches | Explicit mode word. |
| Other screen headers | `? 0 asks` or `? 0`, with status words or labelled glyphs | matches | Each header has an exact cell comparison. |
| Help CHAT line | `i`, `ctrl+]`, click instructions | matches | Decision 85. |
| Original board cells in the 14 rule cases | listed below | overridden | User approval, decision 89, replaces the original cells. |

All changed mode states remain readable with colour removed. The footer
mode words, prompt/cursor, placeholder, feed note, thumb, and header count
remain. Colour adds to these signals.

## Approved rule cases

These are the cases in `board-rule-cases.md`. The latest approved boards
and grids contain the corrected cells. Rows and columns count from 0.

| Case | Board | Row | Columns | Rule or approved decision |
| --- | --- | --- | --- | --- |
| 1 | CompactChatNormal | 1 | 28–79 | Decision 3: full tab name `4 Friction`. |
| 2 | CompactChatNormal | 3 | 3–21 | Task 34: compact title `chat`. |
| 3 | CompactChatNormal | 4 | 1–78 | Task 34: no blank first compact body row. |
| 4 | CompactChatNormal | 4–18 | 1–78 | Decision 30: compact feed has no gutter. |
| 5 | ChatNormal, ChatInput | 5, 7, 10 | 9–118 | Decisions 15 and 26: user and agent tokens. |
| 6 | CompactChatNormal | 5–12 | 1–78 | Same token rules as case 5. |
| 7 | ChatNormal, ChatInput | 12 | 9–118 | Decisions 15 and 57: folded event is dim. |
| 8 | ChatNormal, ChatInput | 20–34 | 119 | Decision 5 and task 34: dim thumb from feed lines. |
| 9 | CompactChatNormal | 12–18 | 79 | Same thumb rules as case 8. |
| 10 | ChatInput | 3 | 2–22 | Decisions 26 and 27: title is accent and bold. |
| 11 | ChatNormal, CompactChatNormal | 36, 20 | 2–25 | Decisions 55 and 27: title and status tokens. |
| 12 | ChatInput | 36 | 2–25 | Decision 55: status tokens stay separate from title. |
| 13 | ChatInput | 37 | 2–32 | Decisions 15 and 26: user input text uses `user`. |
| 14 | Three new boards | 39, 39, 23 | after hints | Decision 4: dim footer padding. |

Decision 89 also replaces `click feed leave` in the input footer with
`ctrl+o events`. Feed clicks still leave input mode. The approved header at
0 uses `fg` bold; decision 87 records this palette choice.

## Gate and cleanup

`verifier/gate.log` contains the full five-command output. Results:

- `gofmt -l $(git ls-files -co --exclude-standard '*.go')`: no output, exit 0.
- `go mod tidy -diff`: no output, exit 0.
- `go vet ./...`: exit 0.
- `golangci-lint run ./...`: `0 issues.`, exit 0.
- `go test -race ./...`: all packages pass, exit 0.

Each demo quit with two ctrl+c presses. Before closure, the verifier's pane
had only its shell in the foreground. Only the verifier's tab `wW2:t2` was
closed. The scratch detached worktree was removed with Git. The entire
`.hive/tmp/t126-verifier/` folder, including tools, binary, exported grids,
board PNGs, HTML, and Firefox profiles, was removed. Absence was checked.
The remaining workspace has only the original verifier tab `wW2:t1`.
No chat was sent and no input was opened on the real home.
