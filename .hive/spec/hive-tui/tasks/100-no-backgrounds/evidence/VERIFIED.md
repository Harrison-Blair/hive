# VERIFIED: 100-no-backgrounds

Verifier: `t100-verifier-1` (Claude Opus 5.5). Task commit: `0a9c3d3d`. Base: `ef2dde2e`.

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | Red: the task `shell_test.go` in a detached worktree at `ef2dde2e`. `go test ./internal/tui -run '^TestEightScreensDoNotPaintTerminalBackground$' -count=1` fails on all 8 subtests with `SGR sets a terminal background: "\x1b[1;38;2;232;228;218;48;2;14;15;19m"`. Green: the same command on the task branch passes all 8 subtests. |
| V2 | pass | Task binary in a herdr tab with a scratch `HIVE_HOME`, a temp `XDG_CONFIG_HOME`, and `HERDR_ENV` and `HERDR_PANE_ID` unset. Screens 1-8 selected with the digit keys. Each screen captured with `herdr pane read --format ansi --lines 60`. |
| V3 | pass | `git diff origin/dev...HEAD --stat` lists only the brief copy, the evidence files, `internal/tui/shell.go`, `internal/tui/shell_test.go`, and `testdata/ste-baseline.json`. |
| V4 | pass | `gofmt -l` and `go mod tidy -diff` print nothing. `go vet ./...` exits 0. `golangci-lint run ./...` prints `0 issues.`. `go test -race ./...` exits 0. |

## V1 test review

The test parses each `ESC [ ... m` sequence itself with `strconv`. It does not
use the style code to find the codes. It rejects `48`, `40`-`47`, `100`-`107`,
and `7`. It skips the RGB and the 256-color arguments of `38`. The honey check
looks for `38;2;245;185;66` in the full screen output. The brief asks for the
sidebar. The live captures in V2 show the honey code on the `▶ N <screen>`
sidebar row.

## V2 captures

An independent Python SGR checker read each capture. It also accepts `:`
separators. A negative control with `48;2;14;15;19` and `41` gave 2 hits.

| Screen | SGR sequences | Background codes | Honey sidebar selection |
| --- | --- | --- | --- |
| 1 Chat | 260 | 0 | `▶ 1 Chat` |
| 2 Agents | 386 | 0 | `▶ 2 Agents` |
| 3 Goals | 393 | 0 | `▶ 3 Goals` |
| 4 Queue | 381 | 0 | `▶ 4 Queue` |
| 5 Tasks | 396 | 0 | `▶ 5 Tasks` |
| 6 Friction | 396 | 0 | `▶ 6 Friction` |
| 7 Health | 438 | 0 | `▶ 7 Health` |
| 8 Logs | 265 | 0 | `▶ 8 Logs` |

The Chat input was not opened. No `i` and no Enter were sent. The TUI quit
with two `ctrl+c` keys. The verifier closed its tab `wQD:t2`.

## V3 note

`origin/dev` moved to `cd2e672e` (tasks 93 and 98) during the verification.
The two-dot `git diff origin/dev --stat` now also shows those dev changes in
reverse. The merge base is still `ef2dde2e`. `git merge-tree` against
`cd2e672e` gives one conflict in `testdata/ste-baseline.json`: adjacent added
lines. The lander must resolve it.

## Repair 1

Verifier: claude-opus-5-5. Scope: the merge of `origin/dev` and the
`testdata/ste-baseline.json` rows only. The TUI was not run.

Merge commit `ac28f3a6`, parents `4c1ef298` (last `VERIFIED.md` commit) and
`13366406` (`origin/dev`).

| Check | Result | Observation |
| --- | --- | --- |
| Rows of both sides kept | pass | Each line of the baseline in `4c1ef298` and in `13366406` is in `HEAD`. `HEAD` has no line from neither side. 88 keys. Against `origin/dev`, the only change is the added row `100-no-backgrounds/brief.md: 0`, in sorted order. |
| Only the merge changed | pass | `git diff --name-only 4c1ef298 HEAD` gives 20 files. This set is equal to the set that `origin/dev` changed since the merge base. Each of the 19 files other than the baseline is identical to `origin/dev`. |
| `git diff origin/dev HEAD --stat` | pass | Only the files of the source table: the brief copy, the evidence, `internal/tui/shell.go`, `internal/tui/shell_test.go`, `testdata/ste-baseline.json`. |
| `git merge-tree --write-tree origin/dev HEAD` | pass | Exit 0, tree `89a191f6`. `origin/dev` is an ancestor of `HEAD`. |
| Gate | pass | `gofmt -l` and `go mod tidy -diff` print nothing. `go vet` exit 0. `golangci-lint` 0 issues. `go test -race ./...` exit 0, no FAIL line. |

The brief copy change (the Repair 1 section) is committed with this file.
