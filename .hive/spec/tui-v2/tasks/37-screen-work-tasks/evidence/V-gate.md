# V-gate Diff scope and gate

- Branch: `task/37-screen-work-tasks` at `57553121`, after the merge of `dev` at
  `13210c3f` (tasks 36, 42, and 43).
- Each changed file is in the source table of the brief: `internal/tui/` and
  its tests, the components `list.go` and `subtabs.go` with their tests (each
  change recorded in the README), the README section "Manager decisions",
  `testdata/ste-baseline.json`, the brief copy, and the evidence.
- `work_queue.go` and `work_queue_test.go` changed only in the merge of task
  36: they referred to the removed Tasks status tab.
- `gofmt -l` and `go mod tidy -diff` printed nothing. Each other command
  exited 0.

## Full output

```text
$ git diff dev...HEAD --stat
 .hive/spec/tui-v2/README.md                        |  12 +
 .../tui-v2/tasks/37-screen-work-tasks/brief.md     | 144 ++++
 .../tasks/37-screen-work-tasks/evidence/V1.md      | 177 +++++
 .../tasks/37-screen-work-tasks/evidence/V2.md      |  90 +++
 .../evidence/screen-compact.png                    | Bin 0 -> 85304 bytes
 .../37-screen-work-tasks/evidence/screen-work.png  | Bin 0 -> 128666 bytes
 internal/tui/agents.go                             |   1 -
 internal/tui/agents_test.go                        |   4 +-
 internal/tui/component/list.go                     |  11 +-
 internal/tui/component/list_test.go                |  11 +
 internal/tui/component/subtabs.go                  |   8 +-
 internal/tui/component/subtabs_test.go             |  11 +
 internal/tui/demo.go                               |   7 +-
 internal/tui/demo_test.go                          |   8 +-
 internal/tui/doc.go                                |   2 +-
 internal/tui/frame.go                              |  18 +-
 internal/tui/goals_queue.go                        |   3 -
 internal/tui/goals_queue_test.go                   |  25 +-
 internal/tui/refresh.go                            |   3 +-
 internal/tui/selection_marker_test.go              |   5 +-
 internal/tui/shell.go                              |   2 +-
 internal/tui/shell_test.go                         |   4 +-
 internal/tui/task_screen.go                        | 757 ---------------------
 internal/tui/task_screen_test.go                   |  96 ---
 internal/tui/tasks_friction_test.go                | 188 +++--
 internal/tui/work_queue.go                         |   2 -
 internal/tui/work_queue_test.go                    |   5 +-
 internal/tui/work_tasks.go                         | 592 ++++++++++++++++
 internal/tui/work_tasks_test.go                    | 231 +++++++
 testdata/ste-baseline.json                         |   1 +
 30 files changed, 1407 insertions(+), 1011 deletions(-)
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
ok  	github.com/Harrison-Blair/hive	19.777s
ok  	github.com/Harrison-Blair/hive/cmd	6.035s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.328s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.442s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.486s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.518s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.302s
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
ok  	github.com/Harrison-Blair/hive/cmd/task	1.475s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.288s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.381s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.259s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	(cached)
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.551s
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
ok  	github.com/Harrison-Blair/hive/internal/tui	59.924s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	20.681s
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
exit 0
```
