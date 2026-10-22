# V-gate Diff scope and gate

Proves: AC1, AC2.

Run after `git merge origin/dev` into the task branch (merge commit 99ee6918).

## Diff scope

```sh
$ git diff origin/dev...HEAD --stat
 .hive/spec/tui-v2/design/CompactFriction.dc.html   |   2 +-
 .hive/spec/tui-v2/design/Friction.dc.html          |   2 +-
 .hive/spec/tui-v2/design/FrictionResolved.dc.html  |   2 +-
 .hive/spec/tui-v2/design/FrictionResolved.grid     |  10 +-
 .hive/spec/tui-v2/design/Work.dc.html              |   2 +-
 .hive/spec/tui-v2/design/Work.grid                 |   2 +-
 .../tui-v2/tasks/43-board-slip-batch-1/brief.md    | 109 +++++++++++++++++++++
 .../tasks/43-board-slip-batch-1/evidence/V1.md     |  39 ++++++++
 .../tasks/43-board-slip-batch-1/evidence/V2.md     |  45 +++++++++
 testdata/ste-baseline.json                         |   1 +
 10 files changed, 204 insertions(+), 10 deletions(-)
```

Only the files of the source table and the evidence changed. The evidence file V-gate.md is added in the next commit. `ste_test.go` needs no change: its glob `.hive/spec/*/tasks/*/brief.md` already covers the brief copy, and the baseline entry is 0.

## Gate

```sh
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
exit=0
$ go mod tidy -diff
exit=0
$ go vet ./...
exit=0
$ golangci-lint run ./...
0 issues.
exit=0
$ go test -race -count=1 ./...
ok  	github.com/Harrison-Blair/hive	25.127s
ok  	github.com/Harrison-Blair/hive/cmd	8.219s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.348s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.462s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.575s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.637s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.322s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.123s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.143s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.499s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.278s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.344s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.204s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	16.753s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.443s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.372s
ok  	github.com/Harrison-Blair/hive/internal/doctor	6.314s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.380s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.018s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.298s
ok  	github.com/Harrison-Blair/hive/internal/land	4.841s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.016s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.194s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.018s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	8.697s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.034s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.303s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	4.763s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.243s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.044s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.015s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.012s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.440s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.196s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	29.874s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.025s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.283s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.633s
ok  	github.com/Harrison-Blair/hive/internal/project	1.110s
ok  	github.com/Harrison-Blair/hive/internal/supervise	4.838s
ok  	github.com/Harrison-Blair/hive/internal/task	11.868s
ok  	github.com/Harrison-Blair/hive/internal/tui	60.080s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	23.398s
ok  	github.com/Harrison-Blair/hive/internal/worktree	6.639s
exit=0
```

Result: pass. `gofmt -l` and `go mod tidy -diff` print nothing. Each other command exits 0.
