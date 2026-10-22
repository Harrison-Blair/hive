# V-gate

At `37712901` on `task/40-screen-logs`, after the merge of `origin/dev` (`e7cdbd09`, with tasks 42 and 36).

## git diff dev...HEAD --stat

```text
 .hive/spec/tui-v2/README.md                     |  14 +
 .hive/spec/tui-v2/tasks/40-screen-logs/brief.md | 141 ++++++
 internal/tui/component/filterbar.go             |  23 +-
 internal/tui/component/filterbar_test.go        |  29 ++
 internal/tui/demo.go                            |  10 +-
 internal/tui/doc.go                             |   2 +-
 internal/tui/escape_clear_test.go               |  89 ----
 internal/tui/fix_test.go                        |   8 +-
 internal/tui/frame.go                           |   9 +-
 internal/tui/logs_golden_test.go                |  32 ++
 internal/tui/logs_screen.go                     | 639 ++++++++++++++++++++++++
 internal/tui/logs_screen_test.go                | 581 ++++++++++++++-------
 internal/tui/logview.go                         | 529 --------------------
 internal/tui/refresh.go                         |   2 +-
 internal/tui/selection_marker_test.go           |   4 +-
 internal/tui/shell.go                           |  56 ---
 internal/tui/shell_test.go                      |   2 +-
 internal/tui/tui.go                             |  67 +--
 internal/tui/tui_test.go                        |  75 +--
 testdata/ste-baseline.json                      |   1 +
 20 files changed, 1352 insertions(+), 961 deletions(-)
```

The stat leaves out this evidence folder, which is in the next commit.

## Gate

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
$ go mod tidy -diff
$ go vet ./...
$ golangci-lint run ./...
0 issues.
$ go test -race ./...
ok  	github.com/Harrison-Blair/hive	19.466s
ok  	github.com/Harrison-Blair/hive/cmd	5.906s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.364s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.407s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.576s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.619s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.310s
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
ok  	github.com/Harrison-Blair/hive/cmd/task	1.499s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.281s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.331s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.220s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	(cached)
ok  	github.com/Harrison-Blair/hive/internal/boardexport	(cached)
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
ok  	github.com/Harrison-Blair/hive/internal/tui	50.598s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	(cached)
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
exit 0
```
