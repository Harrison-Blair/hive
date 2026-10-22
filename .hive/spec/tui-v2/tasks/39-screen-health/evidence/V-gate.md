# V-gate Diff scope and gate

Proves each criterion. Task branch `31dde07e`, after the merge of `dev` at
`13210c3f` (tasks 36 and 43).

## Merge repair

The merge of `dev` at `13210c3f` had conflicts in four files. Each
resolution keeps the code of `dev` and adds the Health parts:

- `internal/tui/frame.go`: `dev` (task 36) adds the field `panels`, keyed by
  view name. The Health fields `panel` and `details` and the wrapper
  `listPanel` are removed. The Health entry gives `panels` for the views Home
  and Agents, with `healthPanel` and `healthDetails`.
- `internal/tui/selection_marker_test.go`: `dev` renames the field `bold` to
  `honey`. The Health case uses `honey`.
- `.hive/spec/tui-v2/README.md`: both sections of `Manager decisions` stay.
  The Health `Registry` bullet now refers to `panels` of task 36.
- `testdata/ste-baseline.json`: both entries stay, in sorted order.

## Diff scope

```text
$ git diff origin/dev...HEAD --stat (without the evidence folder)
 .hive/spec/tui-v2/README.md                       |   8 +
 .hive/spec/tui-v2/tasks/39-screen-health/brief.md | 148 +++++
 internal/tui/agents.go                            |  34 --
 internal/tui/demo.go                              |   8 +-
 internal/tui/demo_test.go                         |   4 +-
 internal/tui/frame.go                             |  25 +-
 internal/tui/health.go                            | 659 ++++++++++++++++++++++
 internal/tui/health_screen.go                     | 190 -------
 internal/tui/health_screen_test.go                | 176 ------
 internal/tui/health_test.go                       | 320 +++++++++++
 internal/tui/restart_test.go                      |  30 +-
 internal/tui/selection_marker_test.go             |   4 +-
 internal/tui/shell_test.go                        |   4 +-
 internal/tui/tui.go                               |   8 +-
 testdata/ste-baseline.json                        |   1 +
 15 files changed, 1179 insertions(+), 440 deletions(-)
```

Each file is in the source table of the brief. The other changes are the
evidence files of this folder.

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
ok  	github.com/Harrison-Blair/hive	21.937s
ok  	github.com/Harrison-Blair/hive/cmd	6.139s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.354s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.463s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.485s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.537s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.331s
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
ok  	github.com/Harrison-Blair/hive/cmd/task	1.481s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.274s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.346s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.295s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	13.456s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.554s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.721s
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
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.383s
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
ok  	github.com/Harrison-Blair/hive/internal/tui	62.209s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	22.183s
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
exit 0
```
