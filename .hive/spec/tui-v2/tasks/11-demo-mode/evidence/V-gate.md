# V-gate

## Merge of dev

`git fetch origin` and `git merge origin/dev` at `fdb4eedf`. The merge commit is `f93fea4d`.

One conflict was in `testdata/ste-baseline.json`. This branch added `.hive/spec/tui-v2/tasks/11-demo-mode/brief.md: 0`, and dev added `.hive/spec/tui-v2/tasks/22-components-board-fix/brief.md: 0` on the same line. Both entries were kept, in sorted order. The STE and gitignore tests passed after the merge.

## Diff against dev

```text
$ git diff origin/dev...HEAD --stat
 .hive/spec/tui-v2/README.md                        |  12 +
 .hive/spec/tui-v2/design/Main.fixture.json         |  58 +++
 .hive/spec/tui-v2/design/TooSmall.fixture.json     |  36 ++
 .hive/spec/tui-v2/tasks/11-demo-mode/brief.md      | 151 ++++++
 .../spec/tui-v2/tasks/11-demo-mode/evidence/V1.md  | 141 ++++++
 .../11-demo-mode/evidence/V2-pane-120x40.ansi      |  40 ++
 .../tasks/11-demo-mode/evidence/V2-pane-120x40.txt |  40 ++
 .../tasks/11-demo-mode/evidence/V2-pane-small.ansi |  55 +++
 .../tasks/11-demo-mode/evidence/V2-pane-small.txt  |  55 +++
 .../11-demo-mode/evidence/V2-too-small-screen.png  | Bin 0 -> 18599 bytes
 .../spec/tui-v2/tasks/11-demo-mode/evidence/V2.md  | 108 +++++
 .../spec/tui-v2/tasks/11-demo-mode/evidence/V3.md  |  97 ++++
 cmd/root.go                                        |  42 ++
 cmd/root_demo_test.go                              | 191 ++++++++
 internal/tui/agents.go                             |  75 ++-
 internal/tui/chat_screen.go                        |  28 +-
 internal/tui/demo.go                               | 521 +++++++++++++++++++++
 internal/tui/demo_test.go                          | 239 ++++++++++
 internal/tui/doc.go                                |   1 +
 internal/tui/friction_screen.go                    |  14 +-
 internal/tui/goals_queue.go                        |   6 +-
 internal/tui/header.go                             |  30 +-
 internal/tui/refresh.go                            |   2 +-
 internal/tui/shell.go                              |   2 +
 internal/tui/task_screen.go                        |  31 +-
 internal/tui/tui.go                                |  60 ++-
 testdata/ste-baseline.json                         |   1 +
 27 files changed, 1976 insertions(+), 60 deletions(-)
```

The V-gate file is not in this stat, because it was committed after the gate run.

Each source file is in the brief source table: `cmd/root.go` and its test, `internal/tui/` and its tests, the two fixtures, the README section `Manager decisions`, `testdata/ste-baseline.json`, and the brief. The other files are evidence.

## Gate

Run from the worktree root after the merge.

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
(exit 0)
$ go mod tidy -diff
(exit 0)
$ go vet ./...
(exit 0)
$ golangci-lint run ./...
0 issues.
(exit 0)
$ go test -race ./...
ok  	github.com/Harrison-Blair/hive	13.729s
ok  	github.com/Harrison-Blair/hive/cmd	5.275s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.103s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.219s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.205s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.091s
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
ok  	github.com/Harrison-Blair/hive/cmd/task	1.220s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.052s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.116s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.024s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	9.983s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.360s
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
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.014s
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
ok  	github.com/Harrison-Blair/hive/internal/supervise	2.597s
ok  	github.com/Harrison-Blair/hive/internal/task	(cached)
ok  	github.com/Harrison-Blair/hive/internal/tui	50.944s
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
(exit 0)
```
