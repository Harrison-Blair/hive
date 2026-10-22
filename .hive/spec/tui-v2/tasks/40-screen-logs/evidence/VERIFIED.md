# Independent verification: 40-screen-logs

Verifier: `t40-verifier-1`.
Date: 2026-10-22.
Verdict: **pass**.

## Revision and scope

Merged `dev` (`13210c3f64f03f3228a5140688ac9f4c2c0106a9`) into the task branch as the brief requires. Verified revision: `e658faeea50086dbb6e1931008cc27178c40f420`.
The verifier changed no product source. Deliberate changes ran in a scratch copy.

`git diff dev...HEAD --stat` is in `verify/scope.txt`. Each changed source file is in `internal/tui/` or `internal/tui/component/`. The README changes are in Manager decisions. The remaining changes are the brief, evidence, and its STE baseline entry. The old Logs code was removed. The changes to shared model, demo, refresh, and shell files wire that removal and the new Logs screen. The removed Escape test has a replacement in `TestLogsEscapeClearsAppliedFilter`. No change concerns another screen.

## Verification items

| Item | Result | Observation |
| --- | --- | --- |
| V1 Tests | pass | New tests reject the base. Both golden checks fail on the old renderer. All task tests pass. A token change, a bold change, and a key change each fail tests. Restored code passes. |
| V2 Live check | pass | Both captures have zero differing cells under the live space-cell rule. All 23 wide key steps and 13 compact key steps pass. |
| V-gate Diff scope and gate | pass | Scope is allowed. All five gate commands exit 0. Formatting and tidy print nothing. Lint reports `0 issues.` |

## V1 runs

1. Exported `dev` to a new scratch folder. Copied `logs_golden_test.go`, `logs_screen_test.go`, and `component/filterbar_test.go` from the task. Ran `go test -count=1 ./internal/tui/...`.
   - Expected build failure: the base lacks `refreshLogSelection`, `parseLogFilter`, and the component fields `Status` and `Focused`. See `verify/tests-base.txt`. These are compile-time red results; they do not claim that each assertion ran on the base.
2. In a clean base export, copied only `logs_golden_test.go`. Ran `go test -count=1 -run 'Test(LogsMatchesItsBoard|CompactLogsMatchesItsBoard)$' -v ./internal/tui`.
   - Logs: 2448 differing cells. CompactLogs: 651 differing cells. Both tests fail on the old renderer. See `verify/golden-base.txt`.
3. On the task revision, ran `go test -count=1 -v ./internal/tui/...`.
   - Both packages pass. This includes the goldens, each responsive size and data case, JSON overflow and paging, typed filters and paste, follow and selection, day navigation, refresh, normal/demo content equality, and the normal-mode program with pipe input. See `verify/tests-green.txt`.
4. In a clean task export, changed the selected message style from `titleStyle` to `accentTextStyle`.
   - `TestLogsMatchesItsBoard` fails on 33 cells. The Logs selection style test also fails. See `verify/mutation-token.txt`.
5. Changed the selected message style to `dimStyle`.
   - The golden and selection tests fail. This changes the foreground token as well as weight. See `verify/mutation-token-dim.txt`.
6. With the style restored, changed the follow key from `f` to `F`.
   - `TestLogsDayAndFollowKeys` and `TestFollowControlsSelection` fail. See `verify/mutation-key.txt`.
7. Restored the scratch source from the verified revision and reran the affected tests.
   - Both goldens, the follow key tests, and the selection tests pass. See `verify/mutations-restored.txt`.

The normal-mode tests use scratch homes. No live coordinator or supervisor was used.

## V2 setup and captures

Built the binary from the task revision with `go build -o .hive/tmp/verify40/hive .`.
Created scratch tab `w9W:t2`, pane `w9W:p2`, in the verifier workspace. Its viewport has 57 rows and fits both boards. Used a scratch `HIVE_HOME` with a home marker.

For each fixture, ran:

```sh
HIVE_HOME=SCRATCH COLORTERM=truecolor .hive/tmp/verify40/hive --demo .hive/spec/tui-v2/design/Logs.fixture.json --size 120x40
HIVE_HOME=SCRATCH COLORTERM=truecolor .hive/tmp/verify40/hive --demo .hive/spec/tui-v2/design/CompactLogs.fixture.json --size 80x24
```

Read `pane.read` with `source: visible` and `format: ansi`. Two reads 500 ms apart agree for each board. Each response has `truncated: false`. `pane.get` gives scroll offset 0. The JSON metadata is in `verify/Logs.json` and `verify/CompactLogs.json`.

Cropped the top-left board area, converted it with `cellgrid.Convert`, and compared each cell with the board grid. Ignored token and bold differences only when both cells are spaces.

| Board | Size | Result | Evidence |
| --- | --- | --- | --- |
| Logs | 120×40 | 0 differences; 673 space-only differences ignored | `verify/Logs.ansi`, `verify/Logs-diff.txt` |
| CompactLogs | 80×24 | 0 differences; 146 space-only differences ignored | `verify/CompactLogs.ansi`, `verify/CompactLogs-diff.txt` |

Captured both approved HTML boards again with Firefox headless, a new scratch profile, and `--window-size=1220,780`. Converted each ANSI capture to an HTML `pre` and captured it with Firefox. The four PNGs are `verify/*-board.png` and `verify/*-live.png`. The HTML wrappers have different padding and font layout; cell positions, text, palette tokens, and weights are established by the grid comparison.

The live key transcripts are `verify/keys-wide.txt` and `verify/keys-compact.txt`. They cover JSON paging, focus, selection, `j`/`k`, top/end, follow, day changes, filter cancel/apply/reset, Help, compact overlay open/close, and the footer hints. Page keys use `send-text` with ESC `[5~` and ESC `[6~`. Each run exits with two `ctrl+c` keys.

Two initial key-check expectations were corrected from fixture and board data: the first fixture event is 09:19:00; the compact board omits the follow status mark. The final transcripts contain only the rerun results. No product change was needed.

## Visual and plain-text review

Read the four PNGs. Removed SGR colour from both initial screens; see `verify/*-plain.txt`. The key transcripts are also plain text.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Wide header | `coordinator ● working`, `supervisor ● live`, `needs you 0` | matches | Logs board |
| Compact header | `coord ●`, `sup ●`, `needs you 0` | matches | CompactLogs board |
| Active tab | `6 Logs`, `━` underline | matches | Decision 3 |
| Fixed filter | `/ level:debug proj:hive` | matches | Both boards, decision 25 |
| Editing filter | cursor `█`, footer `[filter]` | matches | Enter applies; Escape keeps the applied filter |
| Follow | `●` on, `○` off; compact selection moves to newest on follow | matches | Wide board shows the status; compact board omits it |
| Day | `< 2026-10-21 >` | matches | Both boards; keys change the day |
| Event count | `events · 312` | matches | Wide board and compact list after Escape |
| Selection | `▶` | matches | Selection stays visible through movement |
| Level | `DEBUG`, `WARN` | matches | Level is readable without colour |
| Event reason | `verify fail · V4 diff scope` | matches | Manager decision: reason before first semicolon |
| Event scroll mark | `┃` | matches | Wide board and compact list |
| JSON title and focus | `event 13:38:02`, footer `[json]` | matches | Both boards; focus moves independently |
| JSON content and truncation | JSON keys and values, `…` | matches | Full reason remains in the JSON data |
| JSON overflow | `↓ N more lines` | matches | Exercised by the model overflow and page tests |
| Compact overlay | `event 13:38:02 · esc back` | matches | Enter opens; Escape returns to events |
| Empty list | `no events on DATE` or `no events today` | matches | Live day/filter checks and empty-data tests |
| Footer | `[events]`, `[json]`, `[filter]`, key words | matches | Hints follow the focus state |

## Gate and cleanup

Full command output is in `verify/gate.txt`:

```text
gofmt -l $(git ls-files -co --exclude-standard '*.go')   exit 0, no output
go mod tidy -diff                                      exit 0, no output
go vet ./...                                           exit 0, no output
golangci-lint run ./...                                 exit 0, 0 issues.
go test -race ./...                                     exit 0
```

The known flakes did not occur. `git diff dev...HEAD --check` exits 0 before the verifier evidence is added. The raw captures and test output retain terminal trailing spaces and CRLF; Git reports these as whitespace in the evidence. They are capture data.

The scratch program exited. The scratch tab was closed. `verify/cleanup.json` shows only the verifier tab in workspace `w9W`. Removed both scratch source exports, binary, scratch home, HTML, scripts, and Firefox profiles after the evidence was recorded. The final deliverables are only the evidence in this task folder.
