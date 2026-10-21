# 22-components-board-fix V-gate Diff scope and gate

| Item | Value |
| --- | --- |
| Task | 22-components-board-fix |
| Item | V-gate |
| Proves | AC1, AC2 |
| Commit | ef897341 |
| Date | 2026-10-21 |
| Agent | worker, Claude Code, claude-opus-5-5 |

## Run

```sh
echo "commit $(git rev-parse --short HEAD)"
git merge --no-edit dev; echo "exit $?"
git diff dev...HEAD --stat; echo "exit $?"
gofmt -l $(git ls-files -co --exclude-standard '*.go'); echo "exit $?"
go mod tidy -diff; echo "exit $?"
go vet ./...; echo "exit $?"
golangci-lint run ./...; echo "exit $?"
go test -race ./...; echo "exit $?"
```

```text
commit ef897341
Already up to date.
exit 0
 .hive/spec/tui-v2/design/Components.dc.html        |  4 +-
 .hive/spec/tui-v2/design/Components.grid           |  4 +-
 .../tui-v2/tasks/22-components-board-fix/brief.md  | 93 ++++++++++++++++++++++
 testdata/ste-baseline.json                         |  1 +
 4 files changed, 98 insertions(+), 4 deletions(-)
exit 0
exit 0
exit 0
exit 0
0 issues.
exit 0
ok  	github.com/Harrison-Blair/hive	16.875s
ok  	github.com/Harrison-Blair/hive/cmd	4.188s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.229s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.360s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.450s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.161s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.963s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.001s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.460s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.073s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.211s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.027s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	12.303s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.510s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.535s
ok  	github.com/Harrison-Blair/hive/internal/doctor	5.143s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.473s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.018s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.290s
ok  	github.com/Harrison-Blair/hive/internal/land	4.358s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.026s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.204s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.018s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	8.098s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.052s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.380s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	4.264s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.317s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.060s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.039s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.059s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.264s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.229s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	33.545s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.026s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.176s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.425s
ok  	github.com/Harrison-Blair/hive/internal/project	1.099s
ok  	github.com/Harrison-Blair/hive/internal/supervise	2.984s
ok  	github.com/Harrison-Blair/hive/internal/task	9.329s
ok  	github.com/Harrison-Blair/hive/internal/tui	63.631s
ok  	github.com/Harrison-Blair/hive/internal/worktree	4.930s
exit 0
```

## Result

The task branch already contains `dev`. The diff touches only the files of the source table: the board, the grid, the STE baseline, and the brief copy. The evidence files come in the next commit, which adds only files in the evidence folder. `gofmt -l` and `go mod tidy -diff` print nothing. `go vet`, `golangci-lint` (0 issues), and `go test -race` exit 0.
