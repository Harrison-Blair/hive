# V-gate Diff scope and gate

Proves each criterion.

## Merge

`git merge dev` at `909645df` gave merge commit `7aa1516d`. The gate ran at `5b905199`.

## Diff scope

Each path is in the source table of the brief or in the evidence folder:
`internal/tui/` and its tests (with `testdata/agents-board/`), `internal/tui/component/header.go` and its test (component fix, see V1),
`cmd/root_test.go` and `cmd/root_demo_test.go` (header text and demo keys), `.hive/spec/tui-v2/README.md` (section `Manager decisions` only),
`testdata/ste-baseline.json` (brief copy with 0), the brief copy, and the evidence.

```text
$ git diff dev...HEAD --stat
 .hive/spec/tui-v2/README.md                        |  17 +
 .hive/spec/tui-v2/tasks/29-frame-layout/brief.md   | 175 +++++
 .../tui-v2/tasks/29-frame-layout/evidence/V1.md    | 168 +++++
 .../tui-v2/tasks/29-frame-layout/evidence/V2.md    |  72 ++
 .../evidence/screens/compact-enter.png             | Bin 0 -> 46005 bytes
 .../evidence/screens/compact-esc.png               | Bin 0 -> 63646 bytes
 .../evidence/screens/compact-tab.png               | Bin 0 -> 46005 bytes
 .../29-frame-layout/evidence/screens/compact.png   | Bin 0 -> 46005 bytes
 .../evidence/screens/compactagents-enter.png       | Bin 0 -> 40417 bytes
 .../evidence/screens/compactagents-esc.png         | Bin 0 -> 52359 bytes
 .../evidence/screens/compactagents-help.png        | Bin 0 -> 68420 bytes
 .../evidence/screens/compactagents-quit.png        | Bin 0 -> 51196 bytes
 .../evidence/screens/compactagents-tab.png         | Bin 0 -> 52359 bytes
 .../evidence/screens/compactagents.png             | Bin 0 -> 52359 bytes
 .../29-frame-layout/evidence/screens/help-esc.png  | Bin 0 -> 56136 bytes
 .../29-frame-layout/evidence/screens/help.png      | Bin 0 -> 105495 bytes
 .../evidence/screens/main-help-closed.png          | Bin 0 -> 57891 bytes
 .../29-frame-layout/evidence/screens/main-help.png | Bin 0 -> 105465 bytes
 .../evidence/screens/main-quit-notice.png          | Bin 0 -> 54928 bytes
 .../29-frame-layout/evidence/screens/main-tab.png  | Bin 0 -> 57891 bytes
 .../evidence/screens/main-work-tab.png             | Bin 0 -> 96074 bytes
 .../evidence/screens/main-work-tab2.png            | Bin 0 -> 97117 bytes
 .../29-frame-layout/evidence/screens/main-work.png | Bin 0 -> 97117 bytes
 .../29-frame-layout/evidence/screens/main.png      | Bin 0 -> 57891 bytes
 .../29-frame-layout/evidence/screens/toosmall.png  | Bin 0 -> 19458 bytes
 cmd/root_demo_test.go                              |  10 +-
 cmd/root_test.go                                   |   4 +-
 internal/tui/agents.go                             |   4 +-
 internal/tui/agents_test.go                        | 112 +--
 internal/tui/chat_screen_test.go                   |  57 +-
 internal/tui/component/header.go                   |  11 +-
 internal/tui/component/header_test.go              |   8 +
 internal/tui/demo.go                               |  27 +
 internal/tui/demo_test.go                          |  19 +-
 internal/tui/doc.go                                |   3 +-
 internal/tui/escape_clear_test.go                  |   9 +-
 internal/tui/fix_test.go                           |  18 +-
 internal/tui/frame.go                              | 765 +++++++++++++++++++++
 internal/tui/frame_test.go                         | 359 ++++++++++
 internal/tui/goals_queue_test.go                   |  57 +-
 internal/tui/header.go                             | 110 +--
 internal/tui/health_screen_test.go                 |  27 +-
 internal/tui/logs_screen_test.go                   |  35 +-
 internal/tui/restart_test.go                       |  51 +-
 internal/tui/selection_marker_test.go              |  14 +-
 internal/tui/shell.go                              | 185 +----
 internal/tui/shell_test.go                         | 105 ++-
 internal/tui/task_screen_test.go                   |   9 +-
 internal/tui/tasks_friction_test.go                |  53 +-
 .../tui/testdata/agents-board/graph-120x40.golden  |  80 +--
 .../tui/testdata/agents-board/graph-200x57.golden  | 114 +--
 .../tui/testdata/agents-board/table-120x40.golden  |  80 +--
 .../tui/testdata/agents-board/table-200x57.golden  | 114 +--
 .../tui/testdata/agents-board/tree-120x40.golden   |  80 +--
 .../tui/testdata/agents-board/tree-200x57.golden   | 114 +--
 internal/tui/tui.go                                |  97 +--
 internal/tui/tui_test.go                           |  40 +-
 testdata/ste-baseline.json                         |   1 +
 58 files changed, 2327 insertions(+), 877 deletions(-)
```

## Gate

Run from the worktree root. `gofmt -l` and `go mod tidy -diff` print nothing. Each other command exits 0.

```text
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
ok  	github.com/Harrison-Blair/hive	(cached)
ok  	github.com/Harrison-Blair/hive/cmd	(cached)
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	(cached)
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	(cached)
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	(cached)
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	(cached)
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	(cached)
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
ok  	github.com/Harrison-Blair/hive/cmd/task	(cached)
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	(cached)
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	(cached)
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	(cached)
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
ok  	github.com/Harrison-Blair/hive/internal/tui	49.379s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	(cached)
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
exit 0
```
