# V-gate Diff scope and gate

Proves AC1, AC2. `dev` and `origin/dev` were merged into the task branch first
(`git merge dev`: Already up to date).

## Run

```text
commit 8892954d
$ git diff dev...HEAD --stat
 .hive/spec/tui-v2/design/WorkGoals.dc.html         |   1 -
 .hive/spec/tui-v2/design/WorkGoals.grid            |   4 +-
 .../tui-v2/tasks/41-workgoals-board-fix/brief.md   | 101 +++++++++++++++++
 .../tasks/41-workgoals-board-fix/evidence/V1.md    |  64 +++++++++++
 .../tasks/41-workgoals-board-fix/evidence/V2.md    | 123 +++++++++++++++++++++
 testdata/ste-baseline.json                         |   1 +
 6 files changed, 291 insertions(+), 3 deletions(-)
exit 0
$ gofmt -l $(git ls-files -co --exclude-standard "*.go")
exit 0
$ go mod tidy -diff
exit 0
$ go vet ./...
exit 0
$ golangci-lint run ./...
0 issues.
exit 0
$ go test -race ./...
ok  	github.com/Harrison-Blair/hive	21.745s
ok  	github.com/Harrison-Blair/hive/cmd	7.486s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.446s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.573s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.816s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.792s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.440s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.651s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.620s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.733s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.329s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.421s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.319s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	15.435s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.860s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.449s
ok  	github.com/Harrison-Blair/hive/internal/doctor	5.586s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.341s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.023s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.322s
ok  	github.com/Harrison-Blair/hive/internal/land	3.918s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.019s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.228s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.015s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	7.989s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.035s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.321s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	4.073s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.191s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.045s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.016s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.014s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.290s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.176s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	27.922s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.012s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.696s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.440s
ok  	github.com/Harrison-Blair/hive/internal/project	1.116s
ok  	github.com/Harrison-Blair/hive/internal/supervise	3.960s
ok  	github.com/Harrison-Blair/hive/internal/task	10.193s
ok  	github.com/Harrison-Blair/hive/internal/tui	55.938s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	19.868s
ok  	github.com/Harrison-Blair/hive/internal/worktree	5.386s
exit 0
```

## Result

Only the files of the source table and the evidence changed (`ste_test.go`
needs no change: its glob `.hive/spec/*/tasks/*/brief.md` covers the brief).
`gofmt -l` and `go mod tidy -diff` print nothing. `go vet`, `golangci-lint`,
and `go test -race` exit 0.
