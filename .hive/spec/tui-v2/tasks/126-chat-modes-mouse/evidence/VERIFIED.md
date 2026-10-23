# Verification: task 126-chat-modes-mouse

Date: 2026-10-23. Verifier: `t126-verifier-2`.

Checked commit: `9aa295ab2573ba6e17b0132de09eed3346c1c6a9`.
The branch includes `dev` at `bdd2ee46`. `git merge --no-edit dev` reported
`Already up to date.` The verifier changed only evidence.

**Verdict: pass.** This check replaces the failed verdict at `b7875fff`.
The Logs filter now permits wheel events and tab clicks. The current brief
permits the existing-behavior regression guard to pass on the base.

All paths below are relative to this evidence folder.

## Results

| Item | Result | Observation and evidence |
| --- | --- | --- |
| V1 Boards | pass | All 10 latest approved source copies have `cmp` exit 0. Each staged/design PNG pair is pixel-equal. All 42 design boards were captured and inspected. No capture shows `{{`. See `verifier2/boards.log`. |
| V2 Grids | pass | The exporter exits 0. `cellgrid.Compare`, without masks, finds zero differences for all 42 exported grids against design. All 41 approved `new-grids` files equal design. The base comparison shows only the permitted rows. See `verifier2/grids.log`. |
| V3 Tests red and green | pass | All 19 new tests ran 50 times normally and 50 times with race detection on both versions. On the base, 18 fail for their behavior. The Agents guard passes, as the brief permits. A deliberate break makes the guard fail. On the task branch, all 19 pass in both runs. See `verifier2/test-results.log`, `tests.log`, and `guard-red-green.log`. |
| V4 Live check | pass | All 40 screen captures have zero differing cells. The required keys and mouse events pass. Wheel events and tab clicks also work with the Logs filter open. See `verifier2/live-summary.log`, `keys.log`, `live/`, and `keys/`. |
| V5 Visual alignment | pass | Both modes at both sizes match their grids. The four captures were repeated in an owned scratch session. Borders, count, earlier count, thumb, placeholder, and cursor pass. See the points table below. |
| V-gate Scope and gate | pass | The diff contains the named source files, approved AC6 board changes, the required brief copy, and evidence. All five gate commands exit 0. Formatting and tidy print nothing. See `verifier2/diff-stat.txt` and the complete `verifier2/gate.log`. |

## V1: approved sources

The verifier read the plan, brief, spec, rule cases, and worker evidence.
Decisions 89 and 90 record the later user approval. That approval replaces
the original v88 boards in `/home/penguin/source/hive/.hive/tmp/g6/boards/`.
The original three copies therefore differ. This is an approved replacement.

The current approved sources are:

- `/home/penguin/source/hive/.hive/tmp/g6/approved/`: the eight Chat boards.
- `/home/penguin/source/hive/.hive/tmp/g6/approved-ok/`: Help and Main.

Each current approved copy equals its design file. Each pair was captured
with the Firefox command below. A scratch staged copy had the design
`support.js` beside it. Each capture used a new scratch profile folder.
PIL found no pixel difference in any pair.

```sh
cmp STAGED.dc.html .hive/spec/tui-v2/design/BOARD.dc.html
firefox --headless --no-remote --profile SCRATCH \
  --window-size=1220,780 --screenshot OUT.png file://BOARD.dc.html
```

All 42 design boards were also captured and inspected. The component sheet
capture includes its changed header sample. No rendered capture shows `{{`.
The board PNGs and profiles were removed after inspection, as V1 requires.

## V2: grids and scope

```sh
mkdir -p .hive/tmp/t126-v2/export
go run ./internal/boardexport/cmd/boardexport \
  -out .hive/tmp/t126-v2/export .hive/spec/tui-v2/design
```

The exporter printed nothing and exited 0. A scratch Go program used
`cellgrid.ReadGrid` and `cellgrid.Compare`, without masks, for these checks:

1. All 42 exported grids equal their committed design grids.
2. All 41 approved grids in `new-grids/` equal their design grids.
3. Each existing screen outside Chat and Help changes only in row 0.
4. Components changes only in row 7, its header sample.
5. TooSmall has no changed cell.

The base was `f7990406`. Chat and Help change only in the rows that
`new-grids/README.md` names. The three new grids and the 38 updated grids
are tracked. No mask or exporter implementation changed.

The AC6 approval permits the corresponding `.dc.html` changes. The source
table and AC6 together define the scope. Decisions 79–90 record the modes,
mouse behavior, count, approved rule cases, and replacement grids.

## V3: independent red and green runs

The exact 19-test expression is in `verifier2/test-regex.txt`.
The commands were:

```sh
go test ./internal/tui ./internal/tui/component ./internal/boardexport \
  -run "$NEW_TESTS" -count=50 -json
go test ./internal/tui ./internal/tui/component ./internal/boardexport \
  -run "$NEW_TESTS" -race -count=50 -json
```

A detached scratch worktree at `f7990406` received the three changed test
files and the new board sets. Two inert fields permitted compilation: `Header.Asks` and
`headerState.asks`. No feature behavior was added to the base.

Each run has 950 top-level test results. The base gives 900 expected
failures and 50 guard passes. The task branch gives 950 passes. The new
board test reports 167, 108, and 118 differing cells on the base. These
failures concern the expected rendering, rather than compilation or loading.

The guard is `TestMouseWheelOverTheAgentsGraphDoesNothing`. The base
already ignores mouse input, so this guard passes there. In a scratch copy
of the task commit, the Agents wheel path deliberately sent `KeyDown`.
The guard then failed with `the wheel changed the Agents graph`.
After restoration, it passed 50 times with race detection.

The two filter tests also fail on the base in every run:

- `TestMouseWheelWorksWhileTheLogsFilterIsOpen`: selection stays at row 30
  instead of moving to row 31.
- `TestMouseClickOnATabNameWorksWhileTheLogsFilterIsOpen`: the active tab
  stays Logs instead of becoming Work.

Both pass on the task branch. `verifier2/test-results.log` gives the result
count and first failure for each test. The four compressed JSONL files
contain the full test output. `guard-red-green.log` contains the mutation
and restored results.

## V4: live behavior

A fresh task binary ran with this command form:

```sh
TERM=vt100 COLORTERM=truecolor HIVE_BINARY --no-startup \
  --demo .hive/spec/tui-v2/design/BOARD.fixture.json --size WxH
```

Each accepted capture has two equal reads, 500 ms apart. Reads use
`source: visible` and `format: ansi`. Each has `truncated: false` and
scroll offset 0. The CLI supplies the capture. A structured read supplies
the truncated flag that the CLI omits.

The comparison crops the top-left W×H cells and uses `cellgrid.Convert`
and `cellgrid.Compare`, without masks. Only the style of space cells is
ignored in live captures, as the spec permits. Every other cell is compared
fully. All 40 screen boards match. Components has no screen and is checked
by V2 and the gate. TooSmall is unchanged.

As a negative control, ChatNormal was compared against ChatInput.
It reports 602 differing cells in `verifier2/mismatch-control.diff.txt`.

The first board run used owned pane `wW7:p2` at 207×57. During the later
key checks, the user window became 90×61. The verifier then used an owned
scratch session, `t126-verifier2-3995380`, with pane `w1:p1` at 206×69.
All final key checks and repeated V5 captures used that pane. The log
`verifier2/live-harness.log` records the size checks and cleanup.

| Input or event | Observed result | Result |
| --- | --- | --- |
| `i`, then `verify`, at both widths | `[input]` and `> verify█`. | pass |
| `enter` | Demo reports `demo mode: no action`. Mode stays input. | pass |
| `esc`, then `tab` | Mode, draft, and cursor stay. | pass |
| `ctrl+]` | `[normal]`. Draft stays and cursor leaves. | pass |
| Normal-mode `x` | Draft stays `verify`. No character is inserted. | pass |
| `j`, then `k` | Feed moves. Draft stays. | pass |
| Wheel up/down over feed in input mode | Feed moves. Mode and draft stay. | pass |
| Click input, then feed | Input mode starts, then normal mode starts. | pass |
| Click Logs at both widths | Logs opens. | pass |
| Click Logs while input mode is active, then click Chat | Logs opens. Chat returns to input mode with its draft. | pass |
| Logs list wheel down/up, with JSON focus | Selected row moves 30 → 31 → 30. JSON keeps focus. | pass |
| Wheel over JSON that fits | No change. | pass |
| Click Logs list row 14 | Row 14 is selected. Events gets focus. | pass |
| Logs filter open, draft ends in `hivex`, wheel down/up | Selected row moves 14 → 15 → 14. Filter keeps focus and draft. | pass |
| Click Work with filter open, then click Logs | Work opens. Logs returns to the filter with its draft. | pass |
| Agents graph wheel and click | No change. Decision 77 applies. | pass |
| Friction list click | Row 7 selects `agent-can-send-user-chat`. | pass |
| Click Friction details, then wheel over list | Details keeps focus. List selection moves to row 8. | pass |
| Friction details wheel down/up | Details move three lines and return. Selection and focus stay. | pass |
| Work and Health list clicks and wheel | The clicked row is selected. The wheel moves the selection. | pass |

There are 60 final key/mouse captures in `verifier2/keys/`. Each has ANSI
and plain text. Raw SGR wheel bytes are `ESC[<64;X;YM` and
`ESC[<65;X;YM`. Click bytes are `ESC[<0;X;YM ESC[<0;X;Ym`.
Coordinates count from 1. Logical keys use `pane send-keys`. Raw bytes use
`pane send-text`. Each input was followed by a stable read.

The fixture starts its feed beyond its end. The first scroll clamps the
view to the end. Later steps move by lines. Logs wheel-up also disables
follow, as its existing key behavior specifies. Neither behavior changes
mode or draft. `verifier2/mouse-log-sample.jsonl` shows mouse log events
with the pointer cell, focus, and `demo: true`.

A physical shift-and-drag selection was not exercised. The model test
checks cell-motion reports. Terminal copy behavior remains a manual check.
The safe demo blocks delivery. The model tests check the send command.
No chat was sent and no Chat input was opened on the real home.

## V5: visual points

The actual TUI binary used demo fixtures and the normal screen renderers.
The four final captures below ran in the owned scratch pane. Each has an
ANSI file, plain-text file, PNG, and zero-cell diff in `verifier2/live/`.
The PNGs and board captures were inspected. Coordinates count from 0.

| Point | ChatNormal 120×40 | ChatInput 120×40 | CompactChatNormal 80×24 | CompactChat 80×24 | Result |
| --- | --- | --- | --- | --- | --- |
| Panel borders and corners | aligned | aligned | aligned | aligned | pass |
| Focused border | feed, accent | input, accent | feed, accent | input, accent | pass |
| Other border | input, border | feed, border | input, border | feed, border | pass |
| Footer mode | `[normal]` | `[input]` | `[normal]` | `[input]` | pass |
| Header count | `? 0 asks`, col 89 | `? 0 asks`, col 89 | `? 0`, col 54 | `? 0`, col 54 | pass |
| Earlier count | `↑ 42 earlier`, col 61 | `↑ 42 earlier`, col 61 | `↑ 42 earlier`, col 41 | `↑ 44 earlier`, col 41 | pass |
| Thumb | col 119, rows 25–34 | col 119, rows 25–34 | col 79, row 18 | col 79, row 18 | pass |
| Placeholder | row 37, col 2 | absent | row 21, col 2 | absent | pass |
| Cursor | absent | row 37, col 33 | absent | row 21, col 4 | pass |

## Board elements and state without colour

Plain-text captures remove colour. Each changed state keeps a word or
glyph. Mode has `[normal]` or `[input]`. Input has a placeholder or prompt
and cursor. The count and scroll state have numbers and glyphs.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| ChatNormal header | `? 0 asks`, status words, `needs you 0` | matches | Latest approved board, decision 87. |
| ChatNormal tabs | `2 Chat`, underline `━━━━━━━━` | matches | Active tab has a text and glyph signal. |
| ChatNormal feed title and border | `chat · noise normal`, `↑ 42 earlier`, box corners | matches | Footer mode names feed focus. |
| ChatNormal feed body and thumb | actor names, folded event `▸`, thumb `┃` | matches | Full grid equality. |
| ChatNormal input | `message`, `● working 1m`, `i or click here to type` | matches | Placeholder names the action. |
| ChatNormal footer | `[normal]`, key words | matches | Explicit mode word. |
| ChatInput header | `? 0 asks`, status words, `needs you 0` | matches | Latest approved board. |
| ChatInput tabs | `2 Chat`, underline `━━━━━━━━` | matches | Active tab has a glyph signal. |
| ChatInput feed title and border | title, earlier count, box corners | matches | Footer mode names input focus. |
| ChatInput feed body and thumb | actor names, `▸`, `┃` | matches | Full grid equality. |
| ChatInput input | `message`, status word, `> Sounds good, go ahead with g6█` | matches | Prompt and cursor show input focus. |
| ChatInput footer | `[input]`, key words | matches | Explicit mode word. |
| CompactChatNormal header | `? 0`, labelled status dots, `needs you 0` | matches | Compact header. |
| CompactChatNormal tabs | full tab names, active underline | matches | Existing tab-label rule. |
| CompactChatNormal feed title and border | `chat`, `↑ 42 earlier`, box corners | matches | Compact title rule. |
| CompactChatNormal feed body and thumb | actor names on separate lines, `┃` | matches | Compact feed rules. |
| CompactChatNormal input | `message`, status word, placeholder | matches | Explicit input instruction. |
| CompactChatNormal footer | `[normal]`, key words | matches | Explicit mode word. |
| Other Chat boards | mode word, prompt or placeholder, earlier count, thumb | matches | Each full screen equals its grid. |
| Other screen headers | `? 0 asks` or `? 0`, status words or labelled glyphs | matches | Each changed header equals its grid. |
| Components header sample | `? 0 asks`, status words | matches | Exporter and golden component check. |
| Help CHAT line | `i`, `ctrl+]`, click instructions | matches | Decision 85. |
| Original rule-case cells | the words and glyphs listed below | overridden | User approval in decision 89. |

## Approved rule cases

The latest approved boards contain the corrections in
`board-rule-cases.md`. This check found no additional rule case.
Rows and columns count from 0.

| Case | Board | Row | Columns | Rule |
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
| 10 | ChatInput | 3 | 2–22 | Decisions 26 and 27: accent bold title. |
| 11 | ChatNormal, CompactChatNormal | 36, 20 | 2–25 | Decisions 55 and 27: title and status tokens. |
| 12 | ChatInput | 36 | 2–25 | Decision 55: separate status and title tokens. |
| 13 | ChatInput | 37 | 2–32 | Decisions 15 and 26: input text uses `user`. |
| 14 | Three new boards | 39, 39, 23 | after hints | Decision 4: dim footer padding. |

Decision 89 also replaces `click feed leave` with `ctrl+o events` in the
input footer. Feed clicks still stop input mode. The approved count at 0
uses `fg` bold. Decision 87 records that palette choice.

## Gate and cleanup

`verifier2/gate.log` contains the complete output of these commands:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go mod tidy -diff
go vet ./...
golangci-lint run ./...
go test -race ./...
```

The first two commands print nothing. Lint reports `0 issues.` All commands
exit 0. All test packages pass. `git diff --check` also passes.

Each demo quit with two ctrl+c presses. The owned panes had only their
shells in the foreground before cleanup. Only owned tab `wW7:t2` was closed.
The owned herdr TUI exited. The verifier stopped the scratch server and
deleted its session. The detached Git worktree was removed. All scratch files,
profiles, tools, binaries, exports, board PNGs, and temporary HTML were
removed from `.hive/tmp/t126-v2/`.

The scratch folder and session folder are absent. The live workspace keeps
only the original verifier tab `wW7:t1`. Final captures and logs remain in
`verifier2/`. No worker evidence or plan tracker changed.
