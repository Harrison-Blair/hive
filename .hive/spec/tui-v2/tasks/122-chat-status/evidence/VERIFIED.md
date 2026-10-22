# Verification: 122-chat-status

**Verdict: pass.** Each verification item passes after the cursor repair.

- Verifier: `t122-verifier-2`.
- Date: 2026-10-22.
- Worker evidence commit: `55605fa1` (source repair: `60969662`).
- Checked task commit after the required merge: `333f56319c17adb0af519037313d8928e4cd6a94`.
- Merged base: `6f741a6b3a3f61741896919679bc18f144a0abeb` (`dev`).
- Red-run base: `dd4ec26e`. The later base commits change dogfood documents only.
- I changed no project source file, worker evidence file, or tracker.
- Paths below are relative to this evidence folder. New results are in `verifier2/`.
- The first failed verdict remains in commit `885b4863` and in `verifier/`.

## Verification items

| Item | Result | Observation and evidence |
| --- | --- | --- |
| V1 Boards | pass | All six staged boards equal their copies with `cmp`. Fresh Firefox captures of both sets have equal PNG hashes. I inspected the captures. No capture shows `{{`. See `verifier2/boards.log`. |
| V2 Grid diff | pass | Three new grids. Help changes only in row 31. Chat changes only in rows 8, 36, and 39. CompactChat changes only in rows 20 and 23. These are the AC7 cells. Each other grid is unchanged. All 31 exported grids equal their committed files. See `verifier2/grids.diff.txt` and `export.log`. |
| V3 Tests red and green | pass | The new behavior tests fail on base code and pass on task code. The repair model tests fail on the code before the repair. All five cursor-component cases fail with an end-only cursor stub and pass with the real helper. The first verifier's independent regression test now passes. See the red and green logs below. |
| V4 Live check | pass | All six specified board-size captures have zero cell differences. `ctrl+o` opens the full event text from the input and the feed. A second press folds it. `esc` keeps the input height. See `verifier2/live.log` and `captures/`. |
| V5 Visual alignment | pass | Fresh captures cover each of the three states at both sizes. Each board-size capture matches its grid. The five alignment points pass at both sizes. Each input height through the cap, and one height past it, passes. See the tables below and `verifier2/captures/`. |
| V-gate Diff scope and gate | pass | After `git merge dev --no-edit`, only the named source files, their tests, and task evidence differ from dev. All five gate commands exit 0. The first two print nothing. See `verifier2/scope.log` and `gate.log`. |

## V1 and V2 procedure

For each of ChatStatus, ChatEvents, CompactChatStatus, Chat, CompactChat, and Help:

```text
cmp /home/penguin/source/hive/.hive/tmp/g3-chat-status/boards/BOARD.dc.html .hive/spec/tui-v2/design/BOARD.dc.html
firefox --headless --no-remote --profile PROFILE --window-size=1220,780 --screenshot OUT.png file://BOARD.dc.html
```

I put byte-identical board copies and the unchanged `support.js` in scratch folders so both sets render. Each capture used a new profile. Both sets, profiles, and board PNGs were under `.hive/tmp/verifier2/`. I removed them after inspection.

I ran the exporter into a separate scratch output folder:

```sh
go run ./internal/boardexport/cmd/boardexport \
  -out .hive/tmp/verifier2/exported .hive/spec/tui-v2/design
```

All 31 exports equal the committed grids. I compared each committed grid with `git show dev:PATH`, using `internal/lib/cellgrid.ReadGrid` and `Compare`, without masks. Chat has 85 changed cells, CompactChat has 62, and Help has 37. The three new grids are ChatStatus, ChatEvents, and CompactChatStatus.

The three symbol lists agree: 42 symbols. The width test passes for each symbol. The README has all three board rows and decisions 55 to 57. The existing Chat and CompactChat fixtures already give working for one minute.

## V3 red and green runs

I used `git archive` copies below `.hive/tmp/verifier2/` for the red runs. No worktree source changed. Each run uses `-count=1 -v`.

| Tests | Red result and log | Green result and log |
| --- | --- | --- |
| Status, all eight status/age cases, queued mark, caps at both sizes, focus height, event toggle from both regions | fail, `base-model.log` | pass, `green-all.log` |
| Input growth, long-word wrapping, explicit newlines, blocked header, open-event text and message column | fail, `base-component.log` | pass, `green-all.log` |
| New board sizes and symbols | fail, `base-boardexport.log` | pass, `green-all.log` |
| Three new board golden checks | fail, `base-golden.log` | pass, `green-all.log` |
| Long-word cursor movement and insertion at both caps; space-wrapped cursor movement and insertion | fail on `885b4863`, `before-fix-model.log` | pass, `green-all.log` |
| Five component cursor positions | fail with the end-only helper stub, `mutated-cursor.log` | pass, `green-all.log` |
| First verifier's unchanged independent cursor test | previous fail is in `verifier/cursor-red.log` | pass, `independent-cursor.log` |
| Minute-boundary refresh and queued-to-sent/failed transitions | fail, `transitions-red.log` | pass, `transitions-green.log` |

These logs are in `verifier2/`. The transition test source is `verifier2/transitions-test.txt`. It uses a scratch home and literal expected text. It checks 59 seconds as `0m`, the next tick as `1m`, a pending user entry followed by a coordinator reply, and sent and failed delivery with no queued mark.

The main green command is:

```sh
go test ./internal/tui ./internal/tui/component ./internal/boardexport -count=1 -v
```

All 317 top-level tests pass, including all golden checks and responsive checks in these packages.

Two base-only adaptations permit behavioral red checks with the old API:

- Omit the absent `Open: true` initializer in the open-event component test. The test fails its full-text expectation.
- Omit the unsupported `chat_events_open` fixture field. The new golden checks fail on the old rendering.

The green run uses the original tests and fixtures. The cursor component test has new API fields, so its red run uses a Go overlay that returns the end position from `inputCursor`. All five cases fail. The overlay changes no source file.

## V4 live checks

I built this branch with `go build -o .hive/tmp/verifier2/hive .`. I used only owned Herdr pane `wVK:p2`. Its size was 207 columns and at least 46 rows. Each demo command has this form:

```text
COLORTERM=truecolor .hive/tmp/verifier2/hive --no-startup --demo FILE --size WxH
```

I did not change `TERM`. I started no coordinator or supervisor and sent no chat. Each capture has two equal visible ANSI reads at least 500 ms apart. A separate API read confirms `truncated: false`. The pane scroll offset is 0. Each `.read.json` records this evidence.

I cropped the top-left W×H cells and converted them with `internal/lib/cellgrid`. The approved live-check rule excludes token and bold differences only when both cells are spaces. Each other cell compares fully. No mask is used.

| Capture | Cells | Differences |
| --- | ---: | ---: |
| ChatStatus-120x40 | 4,800 | 0 |
| ChatEvents-120x40 | 4,800 | 0 |
| CompactChatStatus-80x24 | 1,920 | 0 |
| Chat-120x40 | 4,800 | 0 |
| CompactChat-80x24 | 1,920 | 0 |
| Help-120x40 | 4,800 | 0 |

Each capture has ANSI, plain text, a converted grid, read metadata, a diff, and a PNG in `verifier2/captures/`. I rendered the captured SGR colours and bold flags to HTML, then captured the HTML with Firefox.

`ChatStatus-open` shows `▾`, the full event ending `landed on dev at 3807a25.`, and `ctrl+o fold events`. `ChatStatus-esc` keeps the input top at row 33 and its four text rows. `ChatStatus-feed-fold` and `ChatStatus-feed-open` check both toggle directions from the feed.

### Cursor repair: user-visible result

I repeated the first verifier's live reproduction at both caps. At 80×24, the input is 445 `x` characters plus `END`. At 120×40, it is 1,141 `x` characters plus `END`.

| Step | Visible last row at both sizes | Result |
| --- | --- | --- |
| Initial position | `xEND█` | pass |
| Press Left | `xEN█` (the cursor covers D) | pass |
| Type Z | `xENZ█` (the cursor still covers D) | pass |

The captures are `Cursor-cap-{before,left,insert}-{80x24,120x40}`. The independent model test also checks that the stored text ends with `xENZD`.

## V5 alignment

I inspected fresh PNGs of all three states at both sizes. Each board-size capture matches its approved grid. The additional-size captures have no approved grid; their diff files state this limit.

| Alignment point | 120×40 | 80×24 | Evidence |
| --- | --- | --- | --- |
| Each panel border and corner lines up | pass | pass | All six state captures |
| Status follows the input title with no broken border | pass | pass | All six state captures |
| Input right border and scroll mark keep their column | pass | pass | CompactChatStatus and all Input captures |
| Feed bottom and input top have no gap or overlap | pass | pass | Each Input capture |
| Open-event text starts in the message text column | pass | pass | Both ChatEvents captures |

`Input-120x40-1` through `Input-120x40-10` cover each text height through the 12-row panel cap. `Input-120x40-11` checks overflow. `Input-80x24-1` through `Input-80x24-4` cover each text height through the six-row panel cap. `Input-80x24-5` checks overflow. Border-cell checks and PNG inspection pass at every height. Overflow keeps the cursor row and the dim right-border scroll mark visible.

### Board elements and states without colour

I removed ANSI colour codes and checked each state label in the plain text.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| ChatStatus header | `coordinator ● working`, `supervisor ● live` | matches | Approved board |
| ChatStatus tabs and feed frame | `2 Chat`, `chat · noise normal`, `↑ 42 earlier` | matches | Grid comparison |
| ChatStatus folded events | `▸` and event name | matches | Decision 57 |
| ChatStatus pending message | `you · pending` | matches | Approved board |
| ChatStatus input border | `message · ● working 2m · queued` | matches | Decision 55 |
| ChatStatus growing input and cursor | Four text rows and `█` | matches | Approved board |
| ChatStatus scroll and footer | `┃`, `[input]`, `ctrl+o open events` | matches | AC7 |
| ChatEvents header | `coordinator ○ idle`, `supervisor ● live` | matches | Approved board |
| ChatEvents tabs and feed frame | `2 Chat`, `chat · noise normal`, `↑ 44 earlier` | matches | Grid comparison |
| ChatEvents open events | `▾` and full event text | matches | Decision 57 |
| ChatEvents input border | `message · ○ idle 5m`, `>` | matches | No queued mark |
| ChatEvents scroll and footer | `┃`, `[feed]`, `ctrl+o fold events` | matches | Approved board |
| CompactChatStatus header | `coord ▲`, `sup ●` | matches | Decision 55 |
| CompactChatStatus tabs and feed | `2 Chat`, `chat`, `▸` and event names | matches | Approved board |
| CompactChatStatus input border | `message · ▲ blocked 1m` | matches | Decision 55 |
| CompactChatStatus capped input | Four visible text rows, `█`, `┃` | matches | Six-row panel cap |
| CompactChatStatus footer | `[input]`, `ctrl+o events` | matches | AC7 |
| Chat header, tabs, and feed | `● working`, `2 Chat`, `▸` and event name | matches | AC7 |
| Chat input and footer | `message · ● working 1m`, `█`, `ctrl+o open events` | matches | AC7 |
| CompactChat header, tabs, and feed | `coord ●`, `2 Chat`, `chat` | matches | Approved board |
| CompactChat input and footer | `message · ● working 1m`, `█`, `ctrl+o events` | matches | AC7 |
| Help overlay and key map | `keys`, `CHAT`, literal key names | matches | Grid comparison |
| Help event-toggle row | `ctrl+o opens or folds all chat events` | matches | Row 31 |
| Help footer | `[help]`, `? or esc close` | matches | Approved board |

No checked state depends on colour alone. The approved fallback choices also pass the model tests: another status uses `○` and its word, no live coordinator uses `○ none`, and no timestamp omits the age.

## Scope, gate, known limit, and cleanup

The conditional `internal/tui/tui.go` changes store the event-toggle state and refresh the status age. The worker records this cause. The cursor repair removes the earlier cursor-coordinate cache. Source changes stay within the brief.

The full gate output after the dev merge is in `verifier2/gate.log`:

```text
gofmt -l $(git ls-files -co --exclude-standard '*.go')  # no output; exit 0
go mod tidy -diff                                    # no output; exit 0
go vet ./...                                         # no output; exit 0
golangci-lint run ./...                              # 0 issues.; exit 0
go test -race ./...                                  # all packages pass; exit 0
```

Known limit: Up and Down can use different word-wrap points from the displayed rows. The user deferred this fix on 2026-10-22. The base records the decision in `reference/dogfood/friction/entries/chat-input-updown-wrap-offset.md` and states that it is outside task 122. This pass does not claim vertical-column alignment for Up and Down. Typing, Left, and the shown insertion point are correct in the checked cases.

I stopped each owned demo process. I confirmed that `wVK:p2` had only its shell before I closed it. The final pane list has only the verifier pane. I removed the scratch archive copies, binaries, scripts, HTML files, board PNGs, browser profiles, contact sheets, and schema capture. I preserved all final evidence and pre-existing files.
