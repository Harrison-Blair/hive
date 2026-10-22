# V-gate Diff scope and gate

Proves each criterion.

## Merge

`git merge origin/dev` at `b145a41f` gave the merge commit `9525b700`. The one conflict was in `testdata/ste-baseline.json`: both sides added a brief line. The merge keeps both lines.

## Diff scope

```
$ git diff origin/dev...HEAD --stat
 .hive/spec/tui-v2/README.md                        |  10 +
 .../tui-v2/tasks/36-screen-work-queue/brief.md     | 139 ++++++
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
 internal/tui/work_queue.go                         | 547 +++++++++++++++++++++
 internal/tui/work_queue_test.go                    | 230 +++++++++
 testdata/ste-baseline.json                         |   1 +
 24 files changed, 1372 insertions(+), 433 deletions(-)
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
ok  	github.com/Harrison-Blair/hive	22.891s
ok  	github.com/Harrison-Blair/hive/cmd	6.774s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.333s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.480s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.543s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.549s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.317s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.133s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.192s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.518s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.373s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.399s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.321s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	15.899s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.646s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.461s
ok  	github.com/Harrison-Blair/hive/internal/doctor	5.029s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.429s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.013s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.365s
ok  	github.com/Harrison-Blair/hive/internal/land	3.887s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.034s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.262s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.025s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	7.983s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.039s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.318s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	4.222s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.288s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.049s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.035s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.021s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.260s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.177s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	29.739s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.010s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.295s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.419s
ok  	github.com/Harrison-Blair/hive/internal/project	1.101s
ok  	github.com/Harrison-Blair/hive/internal/supervise	3.741s
ok  	github.com/Harrison-Blair/hive/internal/task	10.931s
ok  	github.com/Harrison-Blair/hive/internal/tui	62.423s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	22.505s
ok  	github.com/Harrison-Blair/hive/internal/worktree	5.711s
exit 0
```
