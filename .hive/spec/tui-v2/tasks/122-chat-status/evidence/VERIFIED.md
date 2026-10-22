# Verification: 122-chat-status

**Verdict: fail.** The wrapped input cursor does not show the insertion position after Left, at both input caps.

- Verifier: `t122-verifier-1`.
- Date: 2026-10-22.
- Task commit: `e52a04aa`.
- Base: `dd4ec26ec10f1ff2552c50d0f616524a6a9bbb0f` (`dev`).
- `git merge dev --no-edit` printed `Already up to date.`
- I changed no project source file, worker evidence file, or tracker.
- Paths below are relative to this evidence folder.

## Verification items

| Item | Result | Observation and evidence |
| --- | --- | --- |
| V1 Boards | pass | All six `cmp` commands exit 0. I captured each staged board and its copy with Firefox. Each pair has the same PNG hash. No image shows `{{`. See `verifier/boards.log`. |
| V2 Grid diff | pass | Three new grids. `Help.grid` changes in row 31 only. `Chat.grid` changes in rows 8, 36, and 39. `CompactChat.grid` changes in rows 20 and 23. These are the AC7 cells. Each other grid is unchanged. See `verifier/grids.diff.txt`. |
| V3 Tests red and green | fail | The worker tests fail on base code and pass on task code. An additional model check fails on task code at both caps. It detects the cursor defect below. See `verifier/red.log`, `green.log`, `green-all.log`, `red-green-extra.log`, and `cursor-red.log`. |
| V4 Live check | pass | The six specified captures match their grids. `ctrl+o` opens the full event text. `esc` keeps the six-row input panel. See `verifier/live.log` and `verifier/captures/`. |
| V5 Visual alignment | pass | All specified alignment points pass at both sizes. All board-size captures match their grids. Each input height through the cap, and one height past it, keeps the panel join and border columns. See the tables below. The additional cursor check is the V3 failure. |
| V-gate Diff scope and gate | pass | The source changes stay within the brief. All five gate commands exit 0. `gofmt -l` and `go mod tidy -diff` print nothing. See `verifier/scope.log` and `verifier/gate.log`. |

## Failure: cursor and insertion position disagree

The failure occurs in demo mode and in the model, without a live coordinator or chat send.

1. Use the `ChatStatus` fixture with input focus.
2. At 80×24, set the input to 445 `x` characters followed by `END`.
3. At 120×40, use 1,141 `x` characters followed by `END`.
4. Press Left once.
5. Type `Z`.

Both inputs exceed their panel caps. The final wrapped text row starts with `xEND`.

| Step | Expected final row | Actual final row |
| --- | --- | --- |
| Initial position | `xEND█` | `xEND█` |
| After Left | `xEN█` (cursor before `D`) | `xEND█` (cursor stays after `D`) |
| After `Z` | `xENZ█` (cursor remains before `D`) | `xENZD█` (cursor stays after `D`) |

The stored text ends with `xENZD`. Thus insertion occurs before `D`, although the cursor shows a position after `D`.

Evidence:

- `verifier/cursor-test.txt`: independent model check with literal expected text.
- `verifier/cursor-red.log`: both size cases fail on the task code.
- `verifier/cursor-live.log`: live commands and captures.
- `verifier/captures/Cursor-cap-before-{80x24,120x40}.png`.
- `verifier/captures/Cursor-cap-left-{80x24,120x40}.png`.
- `verifier/captures/Cursor-cap-insert-{80x24,120x40}.png`.
- A shorter 75-character prefix also fails at 80×24. The `Cursor-*80x24` captures record that check.

The width calculations disagree:

- `internal/tui/chat.go:289` gives `width-6` to `textarea.SetWidth`.
- The textarea reserves its two-cell prompt inside that width.
- `internal/tui/component/input.go:30` wraps the displayed text at `width-6`, without that additional deduction.
- The renderer then uses the textarea cursor coordinates. Its end-column clamp hides the discrepancy until the user moves the cursor.

I reported the defect to `manager-hive` in inbox record 429. I made no source fix.

### Re-run the failing check

From the worktree root:

```sh
mkdir -p .hive/tmp/verifier-cursor
python3 - <<'PY'
import json
from pathlib import Path
r = Path.cwd()
p = r / '.hive/tmp/verifier-cursor/overlay.json'
p.write_text(json.dumps({'Replace': {
    str(r / 'internal/tui/verifier_cursor_test.go'):
    str(r / '.hive/spec/tui-v2/tasks/122-chat-status/evidence/verifier/cursor-test.txt')
}}))
PY
go test -overlay .hive/tmp/verifier-cursor/overlay.json ./internal/tui \
  -run '^TestVerifierWrappedInputCursorTracksInsertion$' -count=1 -v
```

Expected correct behavior: exit 0. Observed task behavior: exit 1 for both size cases.

## V1 and V2 details

I supplied the unchanged design `support.js` beside temporary, byte-identical HTML copies so Firefox could render both sets.
Each capture used a separate new profile under `.hive/tmp/verifier-122/`:

```text
firefox --headless --no-remote --profile PROFILE --window-size=1220,780 --screenshot OUT.png file://BOARD.dc.html
```

I ran the exporter without changing the committed grids:

```sh
go run ./internal/boardexport/cmd/boardexport \
  -out .hive/tmp/verifier-122/exported .hive/spec/tui-v2/design
```

It exits 0. Each exported grid equals its committed grid.
The comparison with `dev` uses `internal/lib/cellgrid.ReadGrid` and `Compare`, without masks.
The changed-cell counts are 85 for Chat, 62 for CompactChat, and 37 for Help.

## V3 details

The base run uses a `git archive dev` copy under `.hive/tmp/verifier-122/base/`.
I copied the new tests into that copy. The base and task runs use `-count=1 -v`.
Each named behavior test fails on the base and passes on the task:

- `TestChatInputBorderShowsCoordinatorStatusAndQueuedMark`.
- `TestChatInputShowsStatusGlyphAndWholeMinuteAge`, including all eight status cases.
- `TestChatInputCapsAndKeepsCursorVisibleAfterFocusLeaves`, at both sizes.
- `TestChatCtrlOOpensAndFoldsEventsFromBothRegions`.
- `TestInputGrowsToFitWrappedText`.
- `TestInputShowsWrappedTextAndCursor`.
- `TestHeaderBlockedStateUsesWarningGlyph`.
- `TestCheckAcceptsChatStatusSymbols`.
- `TestCheckAcceptsNewChatBoardSizes`.

Two base-only adaptations permit behavior checks with the old API:

- The open-event component test omits the absent `Open: true` initializer. The base fails the full-text expectation.
- The three new fixture copies omit the unsupported `chat_events_open` field. Their golden checks fail on the base render.

The task runs use the original tests and fixtures. The open-event component test and all three new golden checks pass.
`go test ./internal/tui ./internal/tui/component ./internal/boardexport -count=1 -v` also passes.
These checks do not detect cursor movement after wrapping. The independent cursor check does detect that defect.

## V4 live captures

I built the task binary and ran only demo fixtures in owned Herdr pane `wVH:p2`.
Each run uses `COLORTERM=truecolor`, `--no-startup`, `--demo FILE`, and `--size WxH`.
I did not change `TERM` for these final captures.

Each capture has two equal CLI reads at least 500 ms apart.
The API read reports `truncated: false`. The pane scroll offset is 0.
I cropped the top-left grid size and converted it with `internal/lib/cellgrid`.
Decision 54 excludes token and bold differences only when both cells are spaces. Each other cell compares fully. No mask is used.

| Capture | Cells | Differences |
| --- | ---: | ---: |
| `ChatStatus-120x40` | 4,800 | 0 |
| `ChatEvents-120x40` | 4,800 | 0 |
| `CompactChatStatus-80x24` | 1,920 | 0 |
| `Chat-120x40` | 4,800 | 0 |
| `CompactChat-80x24` | 1,920 | 0 |
| `Help-120x40` | 4,800 | 0 |

The live `ChatStatus-open` capture shows `▾`, the complete event text, and the `ctrl+o fold events` hint.
The following `ChatStatus-esc` capture keeps the input top at row 33 and its four text rows.

## V5 alignment

The captures include all three board states at 120×40 and 80×24.
The additional-size captures have no approved grid. Their diff files say this explicitly.

| Alignment point | 120×40 | 80×24 | Evidence |
| --- | --- | --- | --- |
| Panel borders and corners align | pass | pass | All six state captures |
| Status follows the input title without a broken border | pass | pass | All six state captures |
| Input right border and scroll mark keep their column | pass | pass | CompactChatStatus and Input captures |
| Feed bottom and input top have no gap or overlap | pass | pass | Every Input capture |
| Open event text uses the message text column | pass | pass | ChatEvents captures |

`Input-120x40-1` through `Input-120x40-10` cover each text height through the 12-row panel cap.
`Input-120x40-11` checks overflow.
`Input-80x24-1` through `Input-80x24-4` cover each text height through the six-row panel cap.
`Input-80x24-5` checks overflow.
I checked the PNGs and the border cells at every height.

### Board elements and states without colour

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| ChatStatus header | `coordinator ● working`, `supervisor ● live` | matches | Approved board |
| ChatStatus tab and feed frame | `2 Chat`, `chat · noise normal`, `↑ 42 earlier` | matches | Grid comparison |
| ChatStatus folded events | `▸` and event name | matches | Decision 57 |
| ChatStatus pending message | `you · pending` | matches | Approved board |
| ChatStatus input border | `message · ● working 2m · queued` | matches | Decision 55 |
| ChatStatus growing input and cursor | Four text rows and `█` at text end | matches | Initial board state |
| ChatStatus scroll and footer | `┃`, `[input]`, `ctrl+o open events` | matches | AC7 and decision 57 |
| ChatEvents header | `coordinator ○ idle`, `supervisor ● live` | matches | Approved board |
| ChatEvents tab and feed frame | `2 Chat`, `chat · noise normal`, `↑ 44 earlier` | matches | Grid comparison |
| ChatEvents open events | `▾` and full event text | matches | Decision 57 |
| ChatEvents empty input border | `message · ○ idle 5m`, `>` | matches | No queued mark |
| ChatEvents scroll and footer | `┃`, `[feed]`, `ctrl+o fold events` | matches | Approved board |
| CompactChatStatus header | `coord ▲`, `sup ●` | matches | Decision 55 |
| CompactChatStatus tab and feed | `2 Chat`, `chat`, `▸` before event names | matches | Approved board |
| CompactChatStatus input border | `message · ▲ blocked 1m` | matches | Decision 55 |
| CompactChatStatus capped input | Four visible text rows, `█`, and `┃` | matches | Six-row panel cap |
| CompactChatStatus footer | `[input]`, `ctrl+o events` | matches | AC7 |
| Chat header, tabs, and feed | `● working`, `2 Chat`, `▸` before event name | matches | AC7 |
| Chat input and footer | `message · ● working 1m`, `█`, `ctrl+o open events` | matches | AC7 |
| CompactChat header, tabs, and feed | `coord ●`, `2 Chat`, `chat` | matches | Approved board |
| CompactChat input and footer | `message · ● working 1m`, `█`, `ctrl+o events` | matches | AC7 |
| Help overlay and key map | `keys`, `CHAT`, literal key names | matches | Grid comparison |
| Help event-toggle row | `ctrl+o opens or folds all chat events` | matches | Row 31 |
| Help footer | `[help]`, `? or esc close` | matches | Approved board |

I also checked stripped text files for the state glyphs and words. No state depends on colour alone.
The approved fallback states pass the model tests: another status shows `○` and its word, no coordinator shows `○ none`, and no timestamp omits the age.

## Scope, gate, and cleanup

The three symbol lists agree. All 42 symbols have width 1.
The README has the three board rows and decisions 55–57.
The existing Chat and CompactChat fixtures already specify working for one minute, so their data needs no change.
The conditional `internal/tui/tui.go` edit holds toggle state, cursor state, and age refresh, as the worker evidence records.

The full gate output is in `verifier/gate.log`:

```text
gofmt -l $(git ls-files -co --exclude-standard '*.go')  # no output, exit 0
go mod tidy -diff                                    # no output, exit 0
go vet ./...                                         # exit 0
golangci-lint run ./...                              # 0 issues., exit 0
go test -race ./...                                  # all packages pass, exit 0
```

The owned demo processes stopped. I checked that `wVH:p2` contained only its shell before I closed it.
The final pane list contains only the verifier pane in that workspace.
I removed the temporary base copy, binary, scripts, browser profiles, board PNGs, and schema capture after recording the results.
I preserved the final evidence and the pre-existing worker files.
