# V-gate Diff scope and gate

Proves each criterion.

## Diff scope

`dev` merged into the task branch at `43e88c58` (`origin/dev` at `b145a41f`, with task 41).

```text
$ git diff dev...HEAD --stat   (with this evidence staged)
 .../tui-v2/tasks/35-screen-work-goals/brief.md     | 140 +++++++++
 .../tasks/35-screen-work-goals/evidence/V1.md      | 100 ++++++
 .../tasks/35-screen-work-goals/evidence/V2.md      |  82 +++++
 .../35-screen-work-goals/evidence/base-run.txt     |  13 +
 .../35-screen-work-goals/evidence/live/G.ansi      |  40 +++
 .../35-screen-work-goals/evidence/live/c79.ansi    |  24 ++
 .../evidence/live/c80-Down.ansi                    |  24 ++
 .../evidence/live/c80-Enter.ansi                   |  24 ++
 .../evidence/live/c80-Escape.ansi                  |  24 ++
 .../35-screen-work-goals/evidence/live/c80.ansi    |  24 ++
 .../35-screen-work-goals/evidence/live/compare.txt |  10 +
 .../35-screen-work-goals/evidence/live/down.ansi   |  40 +++
 .../35-screen-work-goals/evidence/live/g.ansi      |  40 +++
 .../35-screen-work-goals/evidence/live/j.ansi      |  40 +++
 .../35-screen-work-goals/evidence/live/k.ansi      |  40 +++
 .../35-screen-work-goals/evidence/live/o.ansi      |  40 +++
 .../35-screen-work-goals/evidence/live/start.ansi  |  40 +++
 .../35-screen-work-goals/evidence/live/t.ansi      |  40 +++
 .../35-screen-work-goals/evidence/live/tab.ansi    |  40 +++
 .../35-screen-work-goals/evidence/live/tab2.ansi   |  40 +++
 .../35-screen-work-goals/evidence/mutations.txt    |  13 +
 .../evidence/screens/workgoals-c79.png             | Bin 0 -> 19745 bytes
 .../evidence/screens/workgoals-c80-Enter.png       | Bin 0 -> 72070 bytes
 .../evidence/screens/workgoals-c80.png             | Bin 0 -> 35030 bytes
 .../evidence/screens/workgoals-o.png               | Bin 0 -> 90761 bytes
 .../evidence/screens/workgoals-start.png           | Bin 0 -> 93782 bytes
 .../evidence/screens/workgoals-tab.png             | Bin 0 -> 92831 bytes
 .../evidence/task-run-before-41.txt                |  13 +
 .../35-screen-work-goals/evidence/task-run.txt     |  11 +
 internal/tui/frame.go                              |  35 ++-
 internal/tui/goals_queue.go                        | 349 ++-------------------
 internal/tui/goals_queue_test.go                   |  96 +-----
 internal/tui/selection_marker_test.go              |   9 -
 internal/tui/shell_test.go                         |   2 +-
 internal/tui/work_goals.go                         | 261 +++++++++++++++
 internal/tui/work_goals_test.go                    | 270 ++++++++++++++++
 testdata/ste-baseline.json                         |   1 +
 37 files changed, 1491 insertions(+), 434 deletions(-)
```

Each changed file is in the source table of the brief, or is evidence:

- `internal/tui/` and its tests: the screen (`work_goals.go`), its tests, the registry entry and the panels hook in `frame.go`, and the removal of the Goals part of `goals_queue.go` and its tests.
- No file in `internal/tui/component/` changed.
- `testdata/ste-baseline.json`: the brief copy with 0.
- The brief copy. One sentence of the Manager decisions is in active voice, so that its STE count is 0: "You can make a component fix that the screen needs."
- `.hive/spec/tui-v2/README.md` did not change.

## Gate

Run from the worktree root. `gofmt -l` and `go mod tidy -diff` print nothing. Each other command exits 0.

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
$ go mod tidy -diff
exit 0
$ go vet ./...
exit 0
$ golangci-lint run ./...
0 issues.
exit 0
$ go test -race ./...
ok  	github.com/Harrison-Blair/hive	23.311s
ok  	github.com/Harrison-Blair/hive/cmd	7.089s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.545s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.721s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.805s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.803s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.657s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.409s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.513s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.738s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.371s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.465s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.490s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	15.752s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.891s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.434s
ok  	github.com/Harrison-Blair/hive/internal/doctor	5.267s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.356s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.013s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.303s
ok  	github.com/Harrison-Blair/hive/internal/land	4.109s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.016s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.186s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.013s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	8.058s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.043s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.323s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	4.170s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.204s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.045s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.016s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.014s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.255s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.176s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	29.437s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.015s
ok  	github.com/Harrison-Blair/hive/internal/logs	5.749s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.441s
ok  	github.com/Harrison-Blair/hive/internal/project	1.118s
ok  	github.com/Harrison-Blair/hive/internal/supervise	3.831s
ok  	github.com/Harrison-Blair/hive/internal/task	10.406s
ok  	github.com/Harrison-Blair/hive/internal/tui	67.682s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	22.294s
ok  	github.com/Harrison-Blair/hive/internal/worktree	5.873s
exit 0
```
