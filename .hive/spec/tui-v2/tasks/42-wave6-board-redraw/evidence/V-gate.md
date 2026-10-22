# 42-wave6-board-redraw V-gate Diff scope and gate

| Item | Value |
| --- | --- |
| Task | 42-wave6-board-redraw |
| Item | V-gate |
| Proves | AC1, AC2, AC3, AC4 |
| Commit | 5bc092e8 |
| Date | 2026-10-21 |
| Agent | worker, claude, claude-opus-5-5 |

## Run

commit 5bc092e8

```sh
git fetch -q origin; git merge --no-edit origin/dev; echo "exit $?"; git diff origin/dev...HEAD --stat; echo "exit $?"
```

```text
Already up to date.
exit 0
 .hive/spec/tui-v2/design/AgentDetails.dc.html      |   3 +
 .hive/spec/tui-v2/design/AgentDetails.grid         |  22 +--
 .hive/spec/tui-v2/design/AgentsMany.dc.html        |  38 +----
 .hive/spec/tui-v2/design/AgentsMany.grid           |  40 +++---
 .hive/spec/tui-v2/design/Chat.dc.html              |   4 +-
 .hive/spec/tui-v2/design/Chat.grid                 |  30 ++--
 .hive/spec/tui-v2/design/Compact.dc.html           |   8 +-
 .hive/spec/tui-v2/design/Compact.grid              |  16 +--
 .hive/spec/tui-v2/design/CompactAgents.dc.html     |  12 +-
 .hive/spec/tui-v2/design/CompactAgents.grid        |  54 +++----
 .hive/spec/tui-v2/design/CompactChat.dc.html       |   6 +-
 .hive/spec/tui-v2/design/CompactChat.grid          |  32 ++---
 .hive/spec/tui-v2/design/CompactFriction.dc.html   |   4 +-
 .hive/spec/tui-v2/design/CompactFriction.grid      |  58 ++++----
 .hive/spec/tui-v2/design/CompactHealth.dc.html     |   4 +-
 .hive/spec/tui-v2/design/CompactHealth.grid        |  20 +--
 .hive/spec/tui-v2/design/CompactLogs.dc.html       |   1 -
 .hive/spec/tui-v2/design/CompactLogs.grid          |  20 +--
 .hive/spec/tui-v2/design/Components.dc.html        |   2 +-
 .hive/spec/tui-v2/design/Components.grid           |   4 +-
 .hive/spec/tui-v2/design/Friction.dc.html          |  11 +-
 .hive/spec/tui-v2/design/Friction.grid             | 112 +++++++--------
 .hive/spec/tui-v2/design/FrictionResolved.dc.html  |  11 +-
 .hive/spec/tui-v2/design/FrictionResolved.grid     | 130 ++++++++---------
 .hive/spec/tui-v2/design/Health.dc.html            |   2 +-
 .hive/spec/tui-v2/design/Health.grid               |  64 ++++-----
 .hive/spec/tui-v2/design/HealthAgents.dc.html      |   4 +-
 .hive/spec/tui-v2/design/HealthAgents.grid         |  16 +--
 .hive/spec/tui-v2/design/Help.dc.html              |   3 +
 .hive/spec/tui-v2/design/Help.grid                 |   4 +-
 .hive/spec/tui-v2/design/Logs.dc.html              |  14 +-
 .hive/spec/tui-v2/design/Logs.grid                 | 154 ++++++++++----------
 .hive/spec/tui-v2/design/Main.dc.html              |  20 +--
 .hive/spec/tui-v2/design/Main.grid                 |  54 +++----
 .hive/spec/tui-v2/design/Work.dc.html              |   8 +-
 .hive/spec/tui-v2/design/Work.grid                 |  24 ++--
 .hive/spec/tui-v2/design/WorkGoals.dc.html         |   4 +-
 .hive/spec/tui-v2/design/WorkGoals.grid            |   6 +-
 .hive/spec/tui-v2/design/WorkQueue.dc.html         |   4 +-
 .hive/spec/tui-v2/design/WorkQueue.grid            |   4 +-
 .../tui-v2/tasks/42-wave6-board-redraw/brief.md    | 155 +++++++++++++++++++++
 internal/tui/component/chat_test.go                |  14 +-
 internal/tui/component/details.go                  |  18 ++-
 internal/tui/component/details_test.go             |  31 +++++
 internal/tui/component/list.go                     |  13 +-
 internal/tui/component/list_test.go                |  35 +++++
 internal/tui/component/popup_test.go               |   9 +-
 internal/tui/frame.go                              |  12 +-
 internal/tui/frame_test.go                         |  21 +++
 testdata/ste-baseline.json                         |   1 +
 50 files changed, 779 insertions(+), 557 deletions(-)
exit 0
```

Each path is in the source table of the brief: the boards and grids of
`.hive/spec/tui-v2/design/`, the brief copy, `list.go`, `details.go`, and
their tests, `popup_test.go` and `chat_test.go` (sample input only),
`frame.go` and `frame_test.go`, and `testdata/ste-baseline.json`. No
fixture, `canvas.json`, `support.js`, exporter, or screen file changes. The
evidence files come in the next commit.

Gate, from the worktree root:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go'); echo "exit $?"
go mod tidy -diff; echo "exit $?"
go vet ./...; echo "exit $?"
golangci-lint run ./...; echo "exit $?"
go test -race ./... 2>&1; echo "exit $?"
```

```text
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
ok  	github.com/Harrison-Blair/hive	19.689s
ok  	github.com/Harrison-Blair/hive/cmd	6.511s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.368s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.474s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.466s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.518s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.301s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.976s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.049s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.463s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.265s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.344s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.228s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	14.641s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.539s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.419s
ok  	github.com/Harrison-Blair/hive/internal/doctor	5.001s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.337s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.010s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.302s
ok  	github.com/Harrison-Blair/hive/internal/land	3.443s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.014s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.187s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.015s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	7.791s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.039s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.277s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	3.979s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.146s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.042s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.013s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.012s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.258s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.168s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	27.319s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.009s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.713s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.398s
ok  	github.com/Harrison-Blair/hive/internal/project	1.094s
ok  	github.com/Harrison-Blair/hive/internal/supervise	3.668s
ok  	github.com/Harrison-Blair/hive/internal/task	10.014s
ok  	github.com/Harrison-Blair/hive/internal/tui	53.913s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	19.011s
ok  	github.com/Harrison-Blair/hive/internal/worktree	5.433s
exit 0
```

## Result

The diff holds only the files of the source table, `gofmt -l` and
`go mod tidy -diff` print nothing, and `go vet`, `golangci-lint`, and
`go test -race` exit 0.
