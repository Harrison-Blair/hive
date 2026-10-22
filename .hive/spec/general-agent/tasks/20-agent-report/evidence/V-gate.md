# V-gate Diff scope and gate

- Proves: AC1, AC2, AC3, AC4.
- `origin/dev` is `ae14fe29`. The task branch already contains it, so the merge changes nothing.

```
$ git diff origin/dev...HEAD --stat
 .agents/skills/hive-agent/SKILL.md                 |  69 +++++
 .../general-agent/tasks/20-agent-report/brief.md   | 115 ++++++++
 cmd/agent/agent.go                                 |   3 +-
 cmd/agent/report/report.go                         |  41 +++
 cmd/agent/report/report_test.go                    | 150 +++++++++++
 internal/agent/doc.go                              |   1 +
 internal/agent/report.go                           | 184 +++++++++++++
 internal/agent/report_test.go                      | 294 +++++++++++++++++++++
 internal/inbox/render.go                           |   6 +-
 internal/inbox/render_test.go                      |  16 ++
 internal/lib/inbox/inbox.go                        |   5 +-
 internal/lib/inbox/inbox_test.go                   |  12 +
 internal/lib/records/agent.go                      |   2 +
 internal/lib/records/inbox.go                      |  15 +-
 skills_test.go                                     |  15 ++
 ste_test.go                                        |   1 +
 testdata/ste-baseline.json                         |   2 +
 17 files changed, 922 insertions(+), 9 deletions(-)
```

Each file is in the source table of the brief, with the change of 2026-10-21
for `internal/lib/records/inbox.go` and `internal/lib/inbox/`. The evidence
files are added in the next commit.

## Gate

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
ok  	github.com/Harrison-Blair/hive	33.921s
ok  	github.com/Harrison-Blair/hive/cmd	7.950s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.236s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.363s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.519s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.536s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.186s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.425s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.428s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.454s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.116s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.224s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.073s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	20.549s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.664s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.552s
ok  	github.com/Harrison-Blair/hive/internal/doctor	6.018s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.591s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.031s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.417s
ok  	github.com/Harrison-Blair/hive/internal/land	5.048s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.034s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.377s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.028s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	8.610s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.062s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.408s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	4.675s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.298s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.051s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.030s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.031s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.356s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.210s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	34.326s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.022s
ok  	github.com/Harrison-Blair/hive/internal/logs	5.946s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.573s
ok  	github.com/Harrison-Blair/hive/internal/project	1.155s
ok  	github.com/Harrison-Blair/hive/internal/supervise	3.809s
ok  	github.com/Harrison-Blair/hive/internal/task	13.113s
ok  	github.com/Harrison-Blair/hive/internal/tui	79.185s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	34.740s
ok  	github.com/Harrison-Blair/hive/internal/worktree	8.440s
exit 0
```
