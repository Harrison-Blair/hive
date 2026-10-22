# V-gate Diff scope and gate

Proves each criterion.

## Diff scope

`dev` merged into the task branch at `b2dd38f1` (`origin/dev` at `866bde2a`, with tasks 41 and 42).

```text
$ git diff dev...HEAD --stat   (with this evidence staged)
 .../tui-v2/tasks/35-screen-work-goals/brief.md     | 140 +++++++++
 .../tasks/35-screen-work-goals/evidence/V-gate.md  | 151 +++++++++
 .../tasks/35-screen-work-goals/evidence/V1.md      | 126 ++++++++
 .../tasks/35-screen-work-goals/evidence/V2.md      |  87 +++++
 .../35-screen-work-goals/evidence/base-run.txt     |  13 +
 .../35-screen-work-goals/evidence/live/G.ansi      |  40 +++
 .../35-screen-work-goals/evidence/live/c79.ansi    |  24 ++
 .../evidence/live/c80-Down.ansi                    |  24 ++
 .../evidence/live/c80-Enter.ansi                   |  24 ++
 .../evidence/live/c80-Escape.ansi                  |  24 ++
 .../35-screen-work-goals/evidence/live/c80.ansi    |  24 ++
 .../35-screen-work-goals/evidence/live/compare.txt |   7 +
 .../35-screen-work-goals/evidence/live/down.ansi   |  40 +++
 .../35-screen-work-goals/evidence/live/k.ansi      |  40 +++
 .../35-screen-work-goals/evidence/live/o.ansi      |  40 +++
 .../35-screen-work-goals/evidence/live/start.ansi  |  40 +++
 .../35-screen-work-goals/evidence/live/tab.ansi    |  40 +++
 .../35-screen-work-goals/evidence/live/tab2.ansi   |  40 +++
 .../35-screen-work-goals/evidence/mutations.txt    |  13 +
 .../evidence/screens/workgoals-c79.png             | Bin 0 -> 19745 bytes
 .../evidence/screens/workgoals-c80-Enter.png       | Bin 0 -> 68919 bytes
 .../evidence/screens/workgoals-c80.png             | Bin 0 -> 34557 bytes
 .../evidence/screens/workgoals-o.png               | Bin 0 -> 90761 bytes
 .../evidence/screens/workgoals-start.png           | Bin 0 -> 93372 bytes
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
 35 files changed, 1550 insertions(+), 434 deletions(-)
```

Each changed file is in the source table of the brief, or is evidence:

- `internal/tui/` and its tests: the screen (`work_goals.go`), its tests, the registry entry and the panels hook in `frame.go`, and the removal of the Goals part of `goals_queue.go` and its tests.
- No file in `internal/tui/component/` changed in this task. The shared List and Details rules come from task 42 on `dev`.
- `testdata/ste-baseline.json`: the brief copy with 0.
- The brief copy. One sentence of the Manager decisions is in active voice, so that its STE count is 0: "You can make a component fix that the screen needs."
- `.hive/spec/tui-v2/README.md` did not change.

## Gate

Run from the worktree root at `b2dd38f1`. `gofmt -l` and `go mod tidy -diff` print nothing. Each other command exits 0.

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
ok  	github.com/Harrison-Blair/hive	20.986s
ok  	github.com/Harrison-Blair/hive/cmd	6.631s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.364s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.480s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.538s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.576s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.338s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.096s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.159s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.515s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.290s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.362s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.298s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	14.674s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.617s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.418s
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.972s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.359s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.015s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.302s
ok  	github.com/Harrison-Blair/hive/internal/land	3.648s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.019s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.210s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.013s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	7.741s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.042s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.293s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	4.019s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.195s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.043s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.017s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.013s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.249s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.170s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	29.409s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.009s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.201s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.422s
ok  	github.com/Harrison-Blair/hive/internal/project	1.101s
ok  	github.com/Harrison-Blair/hive/internal/supervise	3.674s
ok  	github.com/Harrison-Blair/hive/internal/task	10.154s
ok  	github.com/Harrison-Blair/hive/internal/tui	64.395s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	21.054s
ok  	github.com/Harrison-Blair/hive/internal/worktree	5.498s
exit 0
```

## Repair after verification 1 (2026-10-21)

`git fetch origin` and `git merge origin/dev` reported `Already up to date`.
The repair changes only the brief copy, `work_goals_test.go`, and evidence.
The earlier verifier evidence stays in `VERIFIED.md` and `verifier/`.
Full gate output: `gate-after-repair.txt`. `gofmt -l` and `go mod tidy -diff`
print nothing. Each of the five commands exits 0.

## Repair after verification 3: merge of dev with task 36 (2026-10-21)

`git merge origin/dev` (`e7cdbd09`, with task 36) at `f6c961e5`. `V1.md`
gives the resolution of each conflict.

```text
$ git diff origin/dev...HEAD --stat -- . ':!.hive/spec/tui-v2/tasks/35-screen-work-goals/evidence'
 .../tui-v2/tasks/35-screen-work-goals/brief.md     | 141 ++++++++
 internal/tui/frame.go                              |  13 +-
 internal/tui/goals_queue.go                        | 357 ++-------------------
 internal/tui/goals_queue_test.go                   |  97 +-----
 internal/tui/selection_marker_test.go              |   9 -
 internal/tui/shell_test.go                         |   2 +-
 internal/tui/work_goals.go                         | 261 +++++++++++++++
 internal/tui/work_goals_test.go                    | 286 +++++++++++++++++
 testdata/ste-baseline.json                         |   1 +
 9 files changed, 726 insertions(+), 441 deletions(-)
```

Each other changed file is evidence of this task. Each file is in the source
table of the brief. `.hive/spec/tui-v2/README.md` has no change from this
task.

Full gate output at `f6c961e5`: `gate-after-merge.txt`. `gofmt -l` and
`go mod tidy -diff` print nothing. `golangci-lint` gives `0 issues.` Each
package of `go test -race ./...` gives `ok` or no test files. Each of the five
commands exits 0. The verifier evidence in `VERIFIED.md`, `verifier/`, and
`verifier3/` stays.
