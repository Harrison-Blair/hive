# V-gate Diff scope and gate

Proves each criterion.

## Merge of dev

Task 42 landed on dev at `866bde2a`. The merge commit `fb8004ef` merged `origin/dev` into `task/33-screen-agents`. The only conflict was in `testdata/ste-baseline.json`: both sides added brief entries. The merge keeps all three entries (33, 41, 42). `internal/tui/component/` has no change from dev, so the shared `List`, `Details`, and frame hint lines are the dev version. `git merge-base --is-ancestor origin/dev HEAD` passes.

## Diff scope

```text
$ git diff origin/dev...HEAD --stat
 .hive/spec/tui-v2/README.md                        |  15 +
 .hive/spec/tui-v2/tasks/33-screen-agents/brief.md  | 150 +++++
 internal/tui/agents.go                             |  34 +-
 internal/tui/agents_screen.go                      | 726 +++++++++++++++++++++
 internal/tui/agents_screen_test.go                 | 358 ++++++++++
 internal/tui/agents_test.go                        | 113 ++--
 internal/tui/demo.go                               |  26 +-
 internal/tui/fix_test.go                           |   2 +-
 internal/tui/frame.go                              |  59 +-
 internal/tui/frame_test.go                         |  24 +-
 internal/tui/selection_marker_test.go              |   5 +-
 internal/tui/shell_test.go                         |   6 +-
 .../tui/testdata/agents-board/graph-120x40.golden  |  40 --
 .../tui/testdata/agents-board/graph-200x57.golden  |  57 --
 .../tui/testdata/agents-board/table-120x40.golden  |  40 --
 .../tui/testdata/agents-board/table-200x57.golden  |  57 --
 .../tui/testdata/agents-board/tree-120x40.golden   |  40 --
 .../tui/testdata/agents-board/tree-200x57.golden   |  57 --
 internal/tui/tui.go                                |   1 +
 testdata/ste-baseline.json                         |   1 +
 20 files changed, 1361 insertions(+), 450 deletions(-)
```

Each file is in the source table of the brief:

- `internal/tui/`: the new screen (`agents_screen.go`), its registry entry and keys (`frame.go`), the graph state (`tui.go`), the loader fields (`agents.go`, `demo.go`), and the removal of the old Agents tab view and its goldens. Health › Agents still uses the old screen.
- `internal/tui/component/`: no change.
- `.hive/spec/tui-v2/README.md`: only the "Manager decisions" section, with the rule of task 33.
- `testdata/ste-baseline.json`: the brief copy with a count of 0. `ste_test.go` needed no change: the pattern `.hive/spec/*/tasks/*/brief.md` covers the brief.
- The brief copy: one passive phrase changed for the STE check, "A component fix that the screen needs is allowed" to "You can make a component fix that the screen needs". The meaning is the same. The manager was told (inbox 219).
- The evidence in this folder.

## Gate

The first full run had one failure in a package that this task does not change:

```text
--- FAIL: TestUpdateCallbackRunsUnderLock (2.46s)
    store_test.go:218: store lock timeout
FAIL	github.com/Harrison-Blair/hive/internal/lib/store	42.799s
```

The load average was 21 at that time: other workers ran their gates on the same machine. `go test -race -count=5 -run TestUpdateCallbackRunsUnderLock ./internal/lib/store` then passed. The full gate below is the second run, at load average 12. It passed. The flake was reported to the manager.

Full output of the second run, from the worktree root:

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
$ go test -race -count=1 ./...
ok  	github.com/Harrison-Blair/hive	25.071s
ok  	github.com/Harrison-Blair/hive/cmd	8.000s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.639s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.632s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.899s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.830s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.377s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.616s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.784s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.645s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.332s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.421s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.250s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	16.209s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.477s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.472s
ok  	github.com/Harrison-Blair/hive/internal/doctor	5.912s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.592s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.018s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.462s
ok  	github.com/Harrison-Blair/hive/internal/land	5.278s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.042s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.525s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.032s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	8.572s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.066s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.403s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	4.616s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.370s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.062s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.019s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.036s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.318s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.202s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	29.497s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.023s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.214s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.504s
ok  	github.com/Harrison-Blair/hive/internal/project	1.244s
ok  	github.com/Harrison-Blair/hive/internal/supervise	4.121s
ok  	github.com/Harrison-Blair/hive/internal/task	10.500s
ok  	github.com/Harrison-Blair/hive/internal/tui	60.361s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	22.555s
ok  	github.com/Harrison-Blair/hive/internal/worktree	5.702s
exit 0
```
