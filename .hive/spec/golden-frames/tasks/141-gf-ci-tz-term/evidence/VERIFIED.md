# Task 141 independent verification

- Verifier: `t141-verifier-1`.
- Date: 2026-10-24 UTC.
- Task branch: `task/141-gf-ci-tz-term` at `23653c3a`.
- Base: `e3f107dc`.
- Verdict: pass.

## Items

| Item | Result | Observation |
| --- | --- | --- |
| V1 Time zones | pass | All six base tests fail in UTC and Kolkata. They pass in New York. All six task tests pass in each zone. |
| V2 TERM | pass | The base fails with each TERM setting in a new clone. The task passes in each setting. A broken quit fails in each setting. Each restored run passes. |
| V3 CI | pass | Both full race suites pass in a new clone of the pushed task branch. One run uses `TERM=dumb`. The other run has no `TERM`. |
| V4 Gate | pass | All five commands exit 0. `gofmt -l` and `go mod tidy -diff` print nothing. The linter reports `0 issues.` |

## Scope and setup

I read the task brief, the task 132 brief, and the worker evidence. I ran the
checks again. The source diff changes only the named tests and the STE baseline.
The six time checks retain their expected text. The shared test helper sets the
model clock to a fixed UTC-04:00 zone and rebuilds its cached transcript.
No product source changes. No screen changes require a visual comparison.

The bare-hive test still checks status 0 and the absence of an `outcome`
envelope. It checks the alternate-screen entry and exit sequences. It requires
the new Chat status row between those sequences.

I created a new clone from `git@github.com:Harrison-Blair/hive.git` with
`--single-branch --branch task/141-gf-ci-tz-term`. Its HEAD was `23653c3a`.
I created separate base and mutation clones from that clone. The base clone
used detached commit `e3f107dc`.

The scratch clones were below
`/home/penguin/source/hive/.hive/tmp/t141-verifier-5p0B9b/`, outside `projects/`.
The worktree used for V1 green runs and V4 was
`/home/penguin/source/hive/projects/hive/worktrees/141-gf-ci-tz-term`.

Tools: `go1.27.2-X:nodwarf5 linux/amd64`, `golangci-lint 2.15.0`.

## V1 Time zones, red and green

For each row, I ran this command in the specified checkout:

```sh
TZ=ZONE go test ./internal/tui/ -count=1 -run '^(TestAskReviewOneQuestion|TestAskCardSitsInTheComposerFrame|TestAskJAndKMoveBetweenCards|TestAskHistoryShowsTheAnswerAndTheDismissal|TestChatCompactStatusRowKeepsTheClock|TestChatCompactAskStatusRowCutsKeys)$'
```

| Checkout | ZONE | Output | Exit |
| --- | --- | --- | --- |
| Base clone | UTC | All six named tests fail, package `FAIL` | 1 |
| Base clone | America/New_York | `ok github.com/Harrison-Blair/hive/internal/tui 0.025s` | 0 |
| Base clone | Asia/Kolkata | All six named tests fail, package `FAIL` | 1 |
| Task worktree | UTC | `ok github.com/Harrison-Blair/hive/internal/tui 0.075s` | 0 |
| Task worktree | America/New_York | `ok github.com/Harrison-Blair/hive/internal/tui 0.054s` | 0 |
| Task worktree | Asia/Kolkata | `ok github.com/Harrison-Blair/hive/internal/tui 0.072s` | 0 |

Base UTC failure excerpts:

```text
--- FAIL: TestAskReviewOneQuestion (0.01s)
    ask_review_test.go:50: view does not contain "? manager-hive 07:30  asked · Model · in the composer, ^r to answer":
--- FAIL: TestAskCardSitsInTheComposerFrame (0.00s)
    ask_test.go:123: view does not contain "? manager-hive 07:30  asked · Model · in the composer, ^r to answer":
--- FAIL: TestAskJAndKMoveBetweenCards (0.00s)
    ask_test.go:312: view does not contain "? coordinator 07:30  asked · Model · in the composer, ^r to answer":
--- FAIL: TestAskHistoryShowsTheAnswerAndTheDismissal (0.00s)
    ask_test.go:350: view does not contain "? manager-hive 07:30  asked · Model":
--- FAIL: TestChatCompactStatusRowKeepsTheClock (0.00s)
    chat_modes_test.go:69: compact status row = " ○ none · ○ supervisor · ? 0 asks · needs you 0                           12:00 "
--- FAIL: TestChatCompactAskStatusRowCutsKeys (0.00s)
    chat_test.go:225: compact ask status row = " ○ none · ask · ↑↓ move · 1..4 pick · enter review · n note · ? 2 asks    12:00 "
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui	0.065s
FAIL
```

Kolkata fails in the same six tests. Its compact rows show `17:30`.
The task tests retain the expected `07:30` ask text and `08:00` status clocks.

## V2 TERM, red and green

Commands:

```sh
TERM=dumb go test ./cmd/ -count=1 -run TestBareHiveRunsTUIUntilQuit
TERM=xterm-256color go test ./cmd/ -count=1 -run TestBareHiveRunsTUIUntilQuit
env -u TERM go test ./cmd/ -count=1 -run TestBareHiveRunsTUIUntilQuit
```

| Checkout or state | TERM=dumb | TERM=xterm-256color | No TERM |
| --- | --- | --- | --- |
| Base clone | fail, exit 1 | fail, exit 1 | fail, exit 1 |
| Task clone, before break | pass, exit 0 | pass, exit 0 | pass, exit 0 |
| Task clone, broken quit | fail, exit 1 | fail, exit 1 | fail, exit 1 |
| Task clone, restored | pass, exit 0 | pass, exit 0 | pass, exit 0 |
| Task worktree | pass, exit 0 | pass, exit 0 | pass, exit 0 |

Each base failure is at `root_test.go:79`, with `shows no TUI ran`.
The output contains the Chat frame and the alternate-screen entry and exit.
It has no `\x1b[?25l` hide-cursor sequence.

For the break, I changed only the second-ctrl+c return in the mutation clone:

```diff
 m.quitNotice = false
-return tea.Quit
+return tea.Interrupt
```

Broken-quit output with `TERM=dumb`:

```text
--- FAIL: TestBareHiveRunsTUIUntilQuit (0.00s)
    root_test.go:82: output "\x1b[?2026$p\x1b[?2027$p" shows no TUI ran and quit
FAIL
FAIL	github.com/Harrison-Blair/hive/cmd	0.009s
FAIL
exit 1
```

The other two TERM settings fail at the same assertion. I restored
`internal/tui/tui.go` with `git restore --source=HEAD -- internal/tui/tui.go`.
`git status --short` then printed nothing. The three restored runs reported:

```text
TERM=dumb:           ok github.com/Harrison-Blair/hive/cmd 0.010s, exit 0
TERM=xterm-256color: ok github.com/Harrison-Blair/hive/cmd 0.009s, exit 0
No TERM:            ok github.com/Harrison-Blair/hive/cmd 0.009s, exit 0
```

## V3 CI-like full suites

Both runs used the new remote clone at `23653c3a`. I removed these variables
from the test shell: `HERDR_ENV`, `HERDR_PANE_ID`, `HERDR_SOCKET_PATH`,
`HERDR_TAB_ID`, `HERDR_WORKSPACE_ID`, and `HERDR_BIN_PATH`.
I exported `GOFLAGS=-count=1` so both commands ran tests without cached results.
The two commands ran in sequence. Each run reports 45 `ok` packages and
34 packages with no test files. No race report occurs.

### TERM=dumb, complete output

```text
$ TZ=UTC TERM=dumb go test -race ./...
ok  	github.com/Harrison-Blair/hive	24.227s
ok  	github.com/Harrison-Blair/hive/cmd	4.080s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.248s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.189s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.486s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.349s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/ask	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.213s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	3.903s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	3.992s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.346s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.299s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.310s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.165s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	6.014s
ok  	github.com/Harrison-Blair/hive/internal/ask	1.173s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.667s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.068s
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.655s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	1.177s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.020s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.107s
ok  	github.com/Harrison-Blair/hive/internal/land	2.870s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.043s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.245s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.012s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	7.027s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.022s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.051s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	1.349s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.107s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.041s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.012s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.008s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.191s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.131s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	17.908s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.007s
ok  	github.com/Harrison-Blair/hive/internal/logs	5.988s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.169s
ok  	github.com/Harrison-Blair/hive/internal/project	1.017s
ok  	github.com/Harrison-Blair/hive/internal/supervise	1.698s
ok  	github.com/Harrison-Blair/hive/internal/task	1.916s
ok  	github.com/Harrison-Blair/hive/internal/tui	35.779s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	11.274s
ok  	github.com/Harrison-Blair/hive/internal/worktree	2.257s
exit 0
```

### No TERM, complete output

```text
$ env -u TERM TZ=UTC go test -race ./...
ok  	github.com/Harrison-Blair/hive	18.897s
ok  	github.com/Harrison-Blair/hive/cmd	3.669s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.122s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.144s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.244s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.155s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/ask	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.112s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.446s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.451s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.140s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.083s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.090s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.077s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	4.307s
ok  	github.com/Harrison-Blair/hive/internal/ask	1.071s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.199s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.034s
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.481s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	1.119s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.004s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.023s
ok  	github.com/Harrison-Blair/hive/internal/land	2.027s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.006s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.060s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.005s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	6.577s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.010s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.023s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	1.202s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.052s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.034s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.007s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.005s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.179s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.120s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	17.597s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.005s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.017s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.154s
ok  	github.com/Harrison-Blair/hive/internal/project	1.010s
ok  	github.com/Harrison-Blair/hive/internal/supervise	1.677s
ok  	github.com/Harrison-Blair/hive/internal/task	1.750s
ok  	github.com/Harrison-Blair/hive/internal/tui	30.021s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	7.521s
ok  	github.com/Harrison-Blair/hive/internal/worktree	2.009s
exit 0
```

`git status --short` in the CI clone printed nothing after both runs.
No pull request was opened. The CI check on `dev` remains the manager's check
after landing.

## V4 Gate

I ran all five commands from the task worktree root at `23653c3a`.
The race suite used `GOFLAGS=-count=1`.

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
exit 0
$ go mod tidy -diff
exit 0
$ go vet ./...
exit 0
$ golangci-lint run ./...
0 issues.
exit 0
$ go test -race ./...
45 packages report ok, 34 packages report [no test files]
exit 0
```

The last command's package output is summarized above. It reports no failed
package and no race. Its root package reports `25.308s`, `cmd` reports `4.126s`,
`internal/tui` reports `37.007s`, and `internal/tui/component` reports `12.363s`.
The first three commands each produce zero output bytes.
`git status --short` printed nothing before this evidence file was created.

## Cleanup

I compared both complete V3 output blocks with their captured logs. Each block
matches its log exactly. All 28 captured exit statuses match the results above.
Each scratch clone was clean after the quit restore. I removed the three
scratch clones and their logs. I checked that the scratch folder was absent.
The only new worktree file is this evidence file. No root `VERIFIED.md` exists.
