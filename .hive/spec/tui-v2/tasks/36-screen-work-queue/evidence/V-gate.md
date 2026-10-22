# V-gate Diff scope and gate

Proves each criterion.

## Merge

The task branch merged dev two times. The first merge (dev `b145a41f`) had one
conflict in `testdata/ste-baseline.json`: both sides added a brief line, and
the merge keeps both lines. The second merge (dev `866bde2a`, after task 42) had no
conflict. The gate below ran at `6238268a`.

## Diff scope

```
$ git diff origin/dev...HEAD --stat
 .hive/spec/tui-v2/README.md                        |  10 +
 .../tui-v2/tasks/36-screen-work-queue/brief.md     | 139 ++++++
 .../tasks/36-screen-work-queue/evidence/V-gate.md  | 140 ++++++
 .../tasks/36-screen-work-queue/evidence/V1.md      | 144 ++++++
 .../tasks/36-screen-work-queue/evidence/V2.md      |  53 ++
 .../36-screen-work-queue/evidence/live/enter.ansi  |  40 ++
 .../36-screen-work-queue/evidence/live/enter.png   | Bin 0 -> 93701 bytes
 .../36-screen-work-queue/evidence/live/k.ansi      |  40 ++
 .../tasks/36-screen-work-queue/evidence/live/k.png | Bin 0 -> 89829 bytes
 .../36-screen-work-queue/evidence/live/s.ansi      |  40 ++
 .../tasks/36-screen-work-queue/evidence/live/s.png | Bin 0 -> 80522 bytes
 .../evidence/live/start-diff.txt                   |   1 +
 .../36-screen-work-queue/evidence/live/start.ansi  |  40 ++
 .../36-screen-work-queue/evidence/live/start.png   | Bin 0 -> 95140 bytes
 .../36-screen-work-queue/evidence/live/t.ansi      |  40 ++
 .../tasks/36-screen-work-queue/evidence/live/t.png | Bin 0 -> 99023 bytes
 internal/tui/demo.go                               |   2 +-
 internal/tui/frame.go                              |  30 +-
 internal/tui/goals_queue.go                        | 309 +-----------
 internal/tui/goals_queue_test.go                   | 126 +----
 internal/tui/selection_marker_test.go              |  11 +-
 internal/tui/shell_test.go                         |   2 +-
 internal/tui/work_queue.go                         | 546 +++++++++++++++++++++
 internal/tui/work_queue_test.go                    | 230 +++++++++
 testdata/ste-baseline.json                         |   1 +
 25 files changed, 1511 insertions(+), 433 deletions(-)
```

Each file is in the source table of the brief, or is evidence:
`internal/tui/` and its tests, `.hive/spec/tui-v2/README.md` (only the
section `Manager decisions`), `testdata/ste-baseline.json` (the brief copy
with 0), the brief copy, and the evidence. The brief copy changed one
sentence to active voice for the STE check. No file of
`internal/tui/component/` changed.

## Gate

From the worktree root:

```
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
ok  	github.com/Harrison-Blair/hive	22.844s
ok  	github.com/Harrison-Blair/hive/cmd	6.838s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.417s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.553s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.622s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.630s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.395s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	(cached)
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	(cached)
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.593s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.282s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.400s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.279s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	(cached)
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.696s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	(cached)
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	(cached)
ok  	github.com/Harrison-Blair/hive/internal/homeinit	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/land	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/log	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/process	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/records	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/run	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/store	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/logs	(cached)
ok  	github.com/Harrison-Blair/hive/internal/memory	(cached)
ok  	github.com/Harrison-Blair/hive/internal/project	(cached)
ok  	github.com/Harrison-Blair/hive/internal/supervise	(cached)
ok  	github.com/Harrison-Blair/hive/internal/task	(cached)
ok  	github.com/Harrison-Blair/hive/internal/tui	65.878s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	23.798s
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
exit 0
```
