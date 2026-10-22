# V-gate Diff scope and gate

Proves each criterion.

## Diff scope

Task branch at `d1e919d2`, after the merge of `dev` at `13210c3f` (tasks 36, 42, and 43).

```text
$ git diff dev...HEAD --stat
 .hive/spec/tui-v2/README.md                        |   13 +
 .../spec/tui-v2/tasks/38-screen-friction/brief.md  |  140 +
 .../tui-v2/tasks/38-screen-friction/evidence/V1.md |   71 +
 .../tui-v2/tasks/38-screen-friction/evidence/V2.md |   82 +
 .../38-screen-friction/evidence/live/ansi2html.py  |   30 +
 .../evidence/live/c-bracket-pane.json              |    1 +
 .../evidence/live/c-bracket.ansi                   |   24 +
 .../evidence/live/c-enter-1-pane.json              |    1 +
 .../evidence/live/c-enter-1.ansi                   |   24 +
 .../evidence/live/c-enter-esc-pane.json            |    1 +
 .../evidence/live/c-enter-esc.ansi                 |   24 +
 .../evidence/live/c-enter-pane.json                |    1 +
 .../38-screen-friction/evidence/live/c-enter.ansi  |   24 +
 .../tasks/38-screen-friction/evidence/live/cap.sh  |   11 +
 .../evidence/live/compactfriction-compare.txt      |    1 +
 .../evidence/live/compactfriction-pane.json        |    1 +
 .../evidence/live/compactfriction.ansi             |   24 +
 .../evidence/live/compare.go.txt                   |   57 +
 .../evidence/live/f-G-g-pane.json                  |    1 +
 .../38-screen-friction/evidence/live/f-G-g.ansi    |   40 +
 .../38-screen-friction/evidence/live/f-G-pane.json |    1 +
 .../38-screen-friction/evidence/live/f-G.ansi      |   40 +
 .../evidence/live/f-a-slash-pane.json              |    1 +
 .../evidence/live/f-a-slash.ansi                   |   40 +
 .../evidence/live/f-bracket-pane.json              |    1 +
 .../evidence/live/f-bracket.ansi                   |   40 +
 .../evidence/live/f-enter-esc-pane.json            |    1 +
 .../evidence/live/f-enter-esc.ansi                 |   40 +
 .../evidence/live/f-enter-j-pane.json              |    1 +
 .../evidence/live/f-enter-j.ansi                   |   40 +
 .../evidence/live/f-enter-pane.json                |    1 +
 .../38-screen-friction/evidence/live/f-enter.ansi  |   40 +
 .../evidence/live/f-help-pane.json                 |    1 +
 .../38-screen-friction/evidence/live/f-help.ansi   |   40 +
 .../38-screen-friction/evidence/live/f-j-pane.json |    1 +
 .../38-screen-friction/evidence/live/f-j.ansi      |   40 +
 .../evidence/live/f-tab-G-pane.json                |    1 +
 .../38-screen-friction/evidence/live/f-tab-G.ansi  |   40 +
 .../evidence/live/f-tab-esc-pane.json              |    1 +
 .../evidence/live/f-tab-esc.ansi                   |   40 +
 .../evidence/live/f-tab-pane.json                  |    1 +
 .../38-screen-friction/evidence/live/f-tab.ansi    |   40 +
 .../evidence/live/friction-compare.txt             |    1 +
 .../evidence/live/friction-pane.json               |    1 +
 .../38-screen-friction/evidence/live/friction.ansi |   40 +
 .../evidence/live/frictionresolved-compare.txt     |    1 +
 .../evidence/live/frictionresolved-pane.json       |    1 +
 .../evidence/live/frictionresolved.ansi            |   40 +
 .../tasks/38-screen-friction/evidence/live/run.sh  |   14 +
 .../evidence/runs/base-details.txt                 |   14 +
 .../evidence/runs/base-friction-golden.txt         | 5828 ++++++++++++++++++++
 .../evidence/runs/base-friction.txt                |   34 +
 .../evidence/runs/base-list-hint.txt               |    4 +
 .../evidence/runs/base-list-selected.txt           |    6 +
 .../evidence/runs/break-file-esc.txt               |    1 +
 .../evidence/runs/break-file-j.txt                 |    1 +
 .../38-screen-friction/evidence/runs/break-key.txt |   10 +
 .../evidence/runs/break-token.txt                  |    6 +
 .../evidence/runs/task-component.txt               |    4 +
 .../evidence/runs/task-friction.txt                |   32 +
 .../evidence/screens/c-enter.png                   |  Bin 0 -> 77255 bytes
 .../evidence/screens/compactfriction.png           |  Bin 0 -> 90372 bytes
 .../evidence/screens/f-bracket.png                 |  Bin 0 -> 187607 bytes
 .../evidence/screens/f-enter-j.png                 |  Bin 0 -> 184467 bytes
 .../evidence/screens/f-enter.png                   |  Bin 0 -> 179901 bytes
 .../38-screen-friction/evidence/screens/f-tab.png  |  Bin 0 -> 182624 bytes
 .../evidence/screens/friction.png                  |  Bin 0 -> 183082 bytes
 .../evidence/screens/frictionresolved.png          |  Bin 0 -> 176757 bytes
 internal/tui/component/list.go                     |   21 +-
 internal/tui/component/list_test.go                |   18 +
 internal/tui/doc.go                                |    2 +-
 internal/tui/frame.go                              |   10 +-
 internal/tui/friction.go                           |  553 ++
 internal/tui/friction_screen.go                    |  537 --
 internal/tui/friction_test.go                      |  254 +
 internal/tui/selection_marker_test.go              |    2 +-
 internal/tui/shell_test.go                         |    2 +-
 internal/tui/tasks_friction_test.go                |   83 +-
 testdata/ste-baseline.json                         |    1 +
 79 files changed, 7950 insertions(+), 593 deletions(-)
```

Only the files of the source table and the evidence changed: `internal/tui/` (the Friction screen, its registry entry, and the removal of `friction_screen.go`), `internal/tui/component/list.go` (the `Hint` field, recorded in the README), the README section `Manager decisions`, `testdata/ste-baseline.json`, the brief, and `evidence/`.

## Gate

From the worktree root, on the code of `d1e919d2` (the commits after the run change only `evidence/`).

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
ok  	github.com/Harrison-Blair/hive	18.530s
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
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.393s
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
ok  	github.com/Harrison-Blair/hive/internal/tui	55.411s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	19.109s
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
exit 0
```

`gofmt -l` and `go mod tidy -diff` print nothing. `go vet`, `golangci-lint`, and `go test -race` exit 0.
