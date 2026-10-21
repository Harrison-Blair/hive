# 10-board-export V-gate Diff scope and gate

| Item | Value |
| --- | --- |
| Task | 10-board-export |
| Item | V-gate |
| Proves | AC1, AC2, AC3 |
| Commit | f74f2323 |
| Date | 2026-10-21 |
| Agent | worker, Claude Code, claude-opus-5-5 |

## Run

The task branch has `origin/dev` at `0b3296b2` merged. The diff uses
`origin/dev`, because the local `dev` ref of this worktree is older. The
stat does not show this file, because this file comes after the run.

```sh
echo "commit $(git rev-parse --short HEAD)"
git diff origin/dev...HEAD --stat; echo "exit $?"
gofmt -l $(git ls-files -co --exclude-standard '*.go'); echo "exit $?"
go mod tidy -diff; echo "exit $?"
go vet ./...; echo "exit $?"
golangci-lint run ./...; echo "exit $?"
go test -race ./...; echo "exit $?"
```

```text
commit f74f2323
 .hive/spec/tui-v2/design/AgentDetails.grid         | 121 +++
 .hive/spec/tui-v2/design/AgentsMany.grid           | 121 +++
 .hive/spec/tui-v2/design/Chat.grid                 | 121 +++
 .hive/spec/tui-v2/design/Compact.grid              |  73 ++
 .hive/spec/tui-v2/design/CompactAgents.grid        |  73 ++
 .hive/spec/tui-v2/design/CompactChat.grid          |  73 ++
 .hive/spec/tui-v2/design/CompactFriction.grid      |  73 ++
 .hive/spec/tui-v2/design/CompactHealth.grid        |  73 ++
 .hive/spec/tui-v2/design/CompactLogs.grid          |  73 ++
 .hive/spec/tui-v2/design/Components.grid           | 202 +++++
 .hive/spec/tui-v2/design/Friction.grid             | 121 +++
 .hive/spec/tui-v2/design/FrictionResolved.grid     | 121 +++
 .hive/spec/tui-v2/design/Health.grid               | 121 +++
 .hive/spec/tui-v2/design/HealthAgents.grid         | 121 +++
 .hive/spec/tui-v2/design/Help.grid                 | 121 +++
 .hive/spec/tui-v2/design/Logs.grid                 | 121 +++
 .hive/spec/tui-v2/design/Main.grid                 | 121 +++
 .hive/spec/tui-v2/design/TooSmall.grid             |  49 ++
 .hive/spec/tui-v2/design/Work.grid                 | 121 +++
 .hive/spec/tui-v2/design/WorkGoals.grid            | 121 +++
 .hive/spec/tui-v2/design/WorkQueue.grid            | 121 +++
 .hive/spec/tui-v2/tasks/10-board-export/brief.md   | 127 +++
 .../tui-v2/tasks/10-board-export/evidence/V1.md    | 303 +++++++
 .../tui-v2/tasks/10-board-export/evidence/V2.md    |  79 ++
 .../tasks/10-board-export/evidence/grids-review.md | 897 +++++++++++++++++++++
 internal/boardexport/check.go                      | 141 ++++
 internal/boardexport/check_test.go                 | 140 ++++
 internal/boardexport/cmd/boardexport/main.go       |  30 +
 internal/boardexport/doc.go                        |   7 +
 internal/boardexport/export.go                     | 117 +++
 internal/boardexport/export_test.go                | 149 ++++
 internal/boardexport/golden_test.go                |  43 +
 testdata/ste-baseline.json                         |   1 +
 33 files changed, 4296 insertions(+)
exit 0
exit 0
exit 0
exit 0
0 issues.
exit 0
ok  	github.com/Harrison-Blair/hive	23.102s
ok  	github.com/Harrison-Blair/hive/cmd	3.395s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.118s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.289s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.240s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.108s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.717s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.702s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.259s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.061s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.127s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.029s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	15.141s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.403s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.363s
ok  	github.com/Harrison-Blair/hive/internal/doctor	6.921s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.446s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.011s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.300s
ok  	github.com/Harrison-Blair/hive/internal/land	5.311s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.016s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.230s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.013s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	9.579s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.032s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.305s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	5.041s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.197s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.055s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.025s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.021s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.644s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.210s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	29.356s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.035s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.248s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.781s
ok  	github.com/Harrison-Blair/hive/internal/project	1.188s
ok  	github.com/Harrison-Blair/hive/internal/supervise	3.981s
ok  	github.com/Harrison-Blair/hive/internal/task	13.479s
ok  	github.com/Harrison-Blair/hive/internal/tui	55.929s
ok  	github.com/Harrison-Blair/hive/internal/worktree	8.147s
exit 0
```

## Result

Only the files of the source table and the evidence changed. `ste_test.go`
needs no change, because its glob `.hive/spec/*/tasks/*/brief.md` covers the
brief copy. `gofmt -l` and `go mod tidy -diff` print nothing, and `go vet`,
`golangci-lint`, and `go test -race` exit 0.
