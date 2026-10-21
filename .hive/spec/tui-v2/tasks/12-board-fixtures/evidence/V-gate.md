# V-gate Diff scope and gate

Proves each criterion. `origin/dev` was merged into the task branch first (merge commit `998e019b`).

## Diff scope

```text
$ git diff origin/dev...HEAD --stat
 .hive/spec/tui-v2/design/AgentDetails.fixture.json |  58 ++++
 .hive/spec/tui-v2/design/AgentsMany.fixture.json   |  71 ++++
 .hive/spec/tui-v2/design/Chat.fixture.json         |  93 ++++++
 .hive/spec/tui-v2/design/Compact.fixture.json      |  60 ++++
 .../spec/tui-v2/design/CompactAgents.fixture.json  |  53 +++
 .hive/spec/tui-v2/design/CompactChat.fixture.json  |  93 ++++++
 .../tui-v2/design/CompactFriction.fixture.json     |  82 +++++
 .../spec/tui-v2/design/CompactHealth.fixture.json  |  66 ++++
 .hive/spec/tui-v2/design/CompactLogs.fixture.json  | 356 +++++++++++++++++++++
 .hive/spec/tui-v2/design/Components.fixture.json   |  57 ++++
 .hive/spec/tui-v2/design/Friction.fixture.json     |  82 +++++
 .../tui-v2/design/FrictionResolved.fixture.json    |  82 +++++
 .hive/spec/tui-v2/design/Health.fixture.json       |  66 ++++
 .hive/spec/tui-v2/design/HealthAgents.fixture.json |  66 ++++
 .hive/spec/tui-v2/design/Help.fixture.json         |  45 +++
 .hive/spec/tui-v2/design/Logs.fixture.json         | 356 +++++++++++++++++++++
 .hive/spec/tui-v2/design/Work.fixture.json         |  60 ++++
 .hive/spec/tui-v2/design/WorkGoals.fixture.json    |  61 ++++
 .hive/spec/tui-v2/design/WorkQueue.fixture.json    |  60 ++++
 .hive/spec/tui-v2/tasks/12-board-fixtures/brief.md | 110 +++++++
 internal/tui/demo_fixtures_test.go                 |  50 +++
 testdata/ste-baseline.json                         |   1 +
 22 files changed, 2028 insertions(+)
```

The diff holds only the files of the source table: the 19 fixtures, the load test, the STE baseline row, and the brief copy. The evidence files are added in the next commit. `ste_test.go` needs no change, because its glob `.hive/spec/*/tasks/*/brief.md` already covers the brief copy. Only the baseline row is new.

## Gate

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
ok  	github.com/Harrison-Blair/hive	15.270s
ok  	github.com/Harrison-Blair/hive/cmd	5.782s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.118s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.260s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.268s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.107s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.814s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.831s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.291s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.064s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.124s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.028s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	11.326s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.479s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.433s
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.856s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.326s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.010s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.287s
ok  	github.com/Harrison-Blair/hive/internal/land	3.353s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.016s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.182s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.015s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	7.467s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.033s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.262s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	3.925s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.147s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.042s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.013s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.011s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.240s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.160s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	27.139s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.010s
ok  	github.com/Harrison-Blair/hive/internal/logs	5.649s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.386s
ok  	github.com/Harrison-Blair/hive/internal/project	1.089s
ok  	github.com/Harrison-Blair/hive/internal/supervise	3.138s
ok  	github.com/Harrison-Blair/hive/internal/task	9.016s
ok  	github.com/Harrison-Blair/hive/internal/tui	51.251s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	6.510s
ok  	github.com/Harrison-Blair/hive/internal/worktree	4.751s
exit 0
```

`gofmt -l` and `go mod tidy -diff` print nothing. Each other command exits 0.
