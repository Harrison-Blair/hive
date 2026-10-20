# 88-tui-terminal-io verification

| Item | Value |
| --- | --- |
| Task | 88-tui-terminal-io |
| Task commit | c3f15af1 |
| Base | 46c363e9 (`origin/dev`, merge base) |
| Date | 2026-10-20 |
| Agent | t88-verifier-1, Claude Opus 5.5 |

## Results

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | `TestFramesMatchTerminalSizesOnEveryScreen` fails on the base (56 of 56 subtests: 8 screens x 7 sizes, short frames and Chat border 3 columns short). It passes on the task branch (56 of 56). |
| V2 | pass | Raw mode is on for both binaries (`-isig -icanon -echo`). Single keys change the screen at once on both. The raw-mode hypothesis is false, as the worker found. R1 is recorded from this raw-mode evidence only, per the manager. |
| V3 | pass | Task binary fills the pane at 206x57 at start and at 91x55 after a herdr split, on all 8 screens. Base fills only 50 of 55 lines. |
| V4 | pass | Only the source-table files and the evidence files changed. |
| V5 | pass | `gofmt -l` and `go mod tidy -diff` print nothing. `go vet`, `golangci-lint` (0 issues), and `go test -race ./...` exit 0. |

## V1 Model frame test

Red: base tree from `git archive 46c363e9`, with the task `internal/tui/tui_test.go` copied in.

```sh
go test ./internal/tui -run '^TestFramesMatchTerminalSizesOnEveryScreen$' -count=1 | grep -c -- '--- FAIL'
```

```text
57
```

That is the parent test and 56 subtests. Sample lines:

```text
--- FAIL: TestFramesMatchTerminalSizesOnEveryScreen/Logs/57x20 (0.00s)
    tui_test.go:276: Logs frame has 17 lines, want 20
--- FAIL: TestFramesMatchTerminalSizesOnEveryScreen/Logs/300x80 (0.00s)
    tui_test.go:276: Logs frame has 75 lines, want 80
```

Green: task worktree.

```text
57 lines with --- PASS (parent and 56 subtests)
ok  	github.com/Harrison-Blair/hive/internal/tui	0.068s
```

The test measures each line with `ansi.StringWidth` and counts lines with `strings.Split`. It does not use the layout code to compute the expected size. It covers all five sizes of R3 and also 207x57 and 103x57.

### Changed existing test

The task changes `TestRunHomeLogsKeysOnStringReader`. It no longer looks for `level: warn` in the program output. It reads the TUI action log for the `level` action. I checked why:

- The original test passes 10 of 10 on the base, and fails 9 of 10 on the task code (`-race -count=10`).
- The failure output shows `…reques…warnno events todaywarn`. The level did change. The renderer wrote only the changed cells (`warn`), so the full text `level: warn` is not in the byte stream.
- With the input `"8"` (no `l`), the new test fails with `program did not apply the level key`. With `"8l"`, it passes.

The new assertion still fails when the key has no effect. I do not consider it a weaker test. The worker evidence does not explain this change.

R1 test: the worker gives no R1 unit test, because a string reader does not model a TTY. V2 is the R1 check.

## V2 Raw mode and keys

Binaries built from the task worktree and the base tree. Each run used `HIVE_HOME` set to a scratch home (only `.hive/README.md`) and a scratch `XDG_CONFIG_HOME`. The runs were in a new herdr tab, PTY `/dev/pts/13`. Keys were digits and `ctrl+c` sent with `herdr pane send-keys`. No `i` and no Enter on Chat.

| Binary | `stty -a < /dev/pts/13` | Keys |
| --- | --- | --- |
| task | `-isig -icanon -echo` | `2`..`8`, `1` each moved to the named screen within 1.5 s |
| base | `-isig -icanon -echo` | `8` then `1` each moved to the named screen within 1.5 s |

Per the manager, the extra typed-key path check (tmux or `script`) was skipped. The user closed the keyboard item. A brief tmux attempt started and was stopped before any key check.

## V3 Frame size live

Measure: `herdr pane read --source visible`, then a Python script counts lines, finds lines wider than the TTY, and checks that each line that ends in `╮`, `╯`, or `│` ends at the last column. The TTY size came from `stty size`.

Task binary, 206x57 at start (no resize before the first frame):

```text
lines=57 want=57 over_width=[] border_at_last_col=53  ▶ 1..8 (all 8 screens)
```

Task binary, after `herdr pane split --direction right --ratio 0.45`, TTY 91x55:

```text
lines=55 want=55 over_width=[] border_at_last_col=50  ▶ 1..8 (all 8 screens)
```

The script also flagged one empty line on 7 screens. That is a script fault: an empty string passed the end-character check. The line has no border.

Base binary at 91x55:

```text
Logs: lines=50 want=55
Chat: lines=50 want=55, 46 border lines end before the last column
```

## V4 Changed files

```sh
git diff origin/dev --stat
```

```text
 .../hive-tui/tasks/88-tui-terminal-io/brief.md     | 109 +++++++++++++
 .../tasks/88-tui-terminal-io/evidence/V1.md        | 181 +++++++++++++++++++++
 .../tasks/88-tui-terminal-io/evidence/V2.md        | 126 ++++++++++++++
 .../tasks/88-tui-terminal-io/evidence/V3.md        |  70 ++++++++
 .../tasks/88-tui-terminal-io/evidence/V4.md        |  32 ++++
 .../tasks/88-tui-terminal-io/evidence/V5.md        | 105 ++++++++++++
 internal/tui/chat_screen.go                        |   4 +-
 internal/tui/logs_screen_test.go                   |  18 +-
 internal/tui/shell.go                              |   2 +-
 internal/tui/tui.go                                |   2 +-
 internal/tui/tui_test.go                           |  52 ++++++
 testdata/ste-baseline.json                         |   1 +
 12 files changed, 696 insertions(+), 6 deletions(-)
```

`testdata/ste-baseline.json` only adds the brief copy with 0.

## V5 Gate

```text
--- gofmt
--- tidy
exit 0
--- vet
exit 0
--- lint
0 issues.
exit 0
--- test
ok  	github.com/Harrison-Blair/hive/internal/tui	14.360s
(all packages ok or no test files)
exit 0
```

## Notes

- `Run` still sets `tea.WithWindowSize(80, 24)`. The live start at 206x57 shows that the first frame uses the real size, so R2 holds.
- Bare `hive` on the scratch home opened `coordinator` and `supervisor` tabs in this workspace on each run. The coordinator Claude stopped at the folder-trust prompt and did nothing. I closed each of these tabs.

## Cleanup

I closed the tabs `wPW:t2` to `wPW:t7`, which my runs opened. No hive process from the scratch binaries is running. The scratch files are in the session scratchpad.
