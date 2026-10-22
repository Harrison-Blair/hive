# V-gate Diff scope and gate

Proves AC1, AC2, AC3, AC4, AC5.

Run on 2026-10-22 at `7d130f54`, from the worktree root. `origin/dev` is at `2adaf0ab`. It has no commit that is not on the task branch, so the merge changes nothing.

## Diff scope

Each file is in the source table of the brief, or is evidence. With the approval of the manager on 2026-10-22 (AC5), the table also has `internal/boardexport/check.go` (`Sizes` only), the board table of `.hive/spec/tui-v2/README.md`, and the board count in `internal/tui/demo_fixtures_test.go`. Each of these edits is data only. `AgentsMany.dc.html`, `AgentsMany.grid`, and `canvas.json` change by the approval of the user of 2026-10-22 (AC4, AC5). No current fixture changes.

```text
$ git merge origin/dev
Already up to date.
$ git diff dev...HEAD --stat
 .hive/spec/general-agent/README.md                 |  12 +
 .../tasks/46-agents-home-graph/brief.md            | 189 ++++++++++
 .../tasks/46-agents-home-graph/evidence/V1.md      | 201 ++++++++++
 .../tasks/46-agents-home-graph/evidence/V2.md      |  96 +++++
 .../screens/AgentDetailsGeneral-1-esc.ansi         |  40 ++
 .../evidence/screens/AgentDetailsGeneral.ansi      |  40 ++
 .../evidence/screens/AgentDetailsGeneral.png       | Bin 0 -> 75756 bytes
 .../screens/AgentsCoordinatorChild-1-up.ansi       |  40 ++
 .../screens/AgentsCoordinatorChild-2-left.ansi     |  40 ++
 .../screens/AgentsCoordinatorChild-3-r.ansi        |  40 ++
 .../evidence/screens/AgentsCoordinatorChild.ansi   |  40 ++
 .../evidence/screens/AgentsCoordinatorChild.png    | Bin 0 -> 64710 bytes
 .../evidence/screens/AgentsHiddenParent-1-h.ansi   |  40 ++
 .../evidence/screens/AgentsHiddenParent-1-h.png    | Bin 0 -> 79697 bytes
 .../evidence/screens/AgentsHiddenParent-2-h.ansi   |  40 ++
 .../evidence/screens/AgentsHiddenParent.ansi       |  40 ++
 .../evidence/screens/AgentsHiddenParent.png        | Bin 0 -> 74590 bytes
 .../evidence/screens/AgentsHome-1-enter.ansi       |  40 ++
 .../evidence/screens/AgentsHome-2-esc-left.ansi    |  40 ++
 .../evidence/screens/AgentsHome-3-right.ansi       |  40 ++
 .../evidence/screens/AgentsHome.ansi               |  40 ++
 .../evidence/screens/AgentsHome.png                | Bin 0 -> 86412 bytes
 .../evidence/screens/AgentsMany.ansi               |  40 ++
 .../evidence/screens/AgentsMany.png                | Bin 0 -> 102270 bytes
 .../evidence/screens/AgentsNested-1-up.ansi        |  40 ++
 .../evidence/screens/AgentsNested-2-up.ansi        |  40 ++
 .../screens/AgentsNested-3-down-down-down.ansi     |  40 ++
 .../evidence/screens/AgentsNested-4-enter.ansi     |  40 ++
 .../evidence/screens/AgentsNested.ansi             |  40 ++
 .../evidence/screens/AgentsNested.png              | Bin 0 -> 81930 bytes
 .../evidence/screens/CompactAgentsHome-1-left.ansi |  24 ++
 .../evidence/screens/CompactAgentsHome-2-left.ansi |  24 ++
 .../evidence/screens/CompactAgentsHome-2-left.png  | Bin 0 -> 60243 bytes
 .../screens/CompactAgentsHome-3-right-right.ansi   |  24 ++
 .../evidence/screens/CompactAgentsHome.ansi        |  24 ++
 .../evidence/screens/CompactAgentsHome.png         | Bin 0 -> 51868 bytes
 .../evidence/screens/HealthAgentsGeneral-1-r.ansi  |  40 ++
 .../evidence/screens/HealthAgentsGeneral-1-r.png   | Bin 0 -> 71732 bytes
 .../HealthAgentsGeneral-2-up-up-up-up-up-r.ansi    |  40 ++
 .../evidence/screens/HealthAgentsGeneral.ansi      |  40 ++
 .../evidence/screens/HealthAgentsGeneral.png       | Bin 0 -> 73037 bytes
 .../evidence/screens/compare-program.go.txt        |  57 +++
 .../evidence/screens/compare.txt                   |   8 +
 .hive/spec/tui-v2/README.md                        |   7 +
 .../spec/tui-v2/design/AgentDetailsGeneral.dc.html |  37 ++
 .../tui-v2/design/AgentDetailsGeneral.fixture.json |  53 +++
 .hive/spec/tui-v2/design/AgentDetailsGeneral.grid  | 121 ++++++
 .../tui-v2/design/AgentsCoordinatorChild.dc.html   |  37 ++
 .../design/AgentsCoordinatorChild.fixture.json     |  51 +++
 .../spec/tui-v2/design/AgentsCoordinatorChild.grid | 121 ++++++
 .../spec/tui-v2/design/AgentsHiddenParent.dc.html  |  37 ++
 .../tui-v2/design/AgentsHiddenParent.fixture.json  |  59 +++
 .hive/spec/tui-v2/design/AgentsHiddenParent.grid   | 121 ++++++
 .hive/spec/tui-v2/design/AgentsHome.dc.html        |  37 ++
 .hive/spec/tui-v2/design/AgentsHome.fixture.json   |  60 +++
 .hive/spec/tui-v2/design/AgentsHome.grid           | 121 ++++++
 .hive/spec/tui-v2/design/AgentsMany.dc.html        |   2 +-
 .hive/spec/tui-v2/design/AgentsMany.grid           |  44 +--
 .hive/spec/tui-v2/design/AgentsNested.dc.html      |  37 ++
 .hive/spec/tui-v2/design/AgentsNested.fixture.json |  60 +++
 .hive/spec/tui-v2/design/AgentsNested.grid         | 121 ++++++
 .hive/spec/tui-v2/design/CompactAgentsHome.dc.html |  37 ++
 .../tui-v2/design/CompactAgentsHome.fixture.json   |  59 +++
 .hive/spec/tui-v2/design/CompactAgentsHome.grid    |  73 ++++
 .../spec/tui-v2/design/HealthAgentsGeneral.dc.html |  37 ++
 .../tui-v2/design/HealthAgentsGeneral.fixture.json |  67 ++++
 .hive/spec/tui-v2/design/HealthAgentsGeneral.grid  | 121 ++++++
 .hive/spec/tui-v2/design/canvas.json               | 404 ++++++++++++++++++++-
 internal/boardexport/check.go                      |   8 +
 internal/tui/agents.go                             |   4 +
 internal/tui/agents_home_test.go                   | 374 +++++++++++++++++++
 internal/tui/agents_screen.go                      | 245 +++++++++++--
 internal/tui/agents_screen_test.go                 |  15 +-
 internal/tui/demo_fixtures_test.go                 |   4 +-
 internal/tui/health.go                             |   7 +-
 testdata/ste-baseline.json                         |   1 +
 76 files changed, 4048 insertions(+), 73 deletions(-)
```

## Gate

`gofmt -l` and `go mod tidy -diff` print nothing. Each command exits 0. Each golden check passes, with `TestAgentsBoardsMatchTheirGrids/AgentsMany` and the seven subtests of `TestGeneralAgentBoardsMatchTheirGrids`.

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
ok  	github.com/Harrison-Blair/hive	18.033s
ok  	github.com/Harrison-Blair/hive/cmd	6.353s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.334s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.438s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.490s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.478s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.302s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.091s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.187s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.457s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.301s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.357s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.228s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	13.677s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.553s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.409s
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.887s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.308s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.011s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.305s
ok  	github.com/Harrison-Blair/hive/internal/land	3.364s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.017s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.179s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.012s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	7.558s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.031s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.276s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	3.907s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.134s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.042s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.014s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.012s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.233s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.169s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	26.946s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.010s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.650s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.389s
ok  	github.com/Harrison-Blair/hive/internal/project	1.085s
ok  	github.com/Harrison-Blair/hive/internal/supervise	3.501s
ok  	github.com/Harrison-Blair/hive/internal/task	9.356s
ok  	github.com/Harrison-Blair/hive/internal/tui	37.039s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	16.625s
ok  	github.com/Harrison-Blair/hive/internal/worktree	4.670s
exit 0
```
