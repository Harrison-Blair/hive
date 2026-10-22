# V-gate Diff scope and gate

Proves each criterion.

## Merges of dev

The repair round merged `origin/dev` four times. Each conflict kept both sides.

| Merge commit | dev change | Conflict and resolution |
| --- | --- | --- |
| `ea832f5d` | task 36 | `testdata/ste-baseline.json`: kept the 33 and 36 brief entries. The README merged without a conflict. |
| `ddc64a61` | task 40 (`fbca94d5`) | `internal/tui/frame.go`, `contentView`: kept the Agents case and the Logs case. |
| `3f1a876f` | task 45 (`5a4b29e6`): the re-exported Main, AgentsMany, and CompactAgents grids and the decision 13 symbols | `testdata/ste-baseline.json`: kept the 33 and 38 entries. The decision 13 symbol list, `internal/boardexport/check.go`, and `internal/tui/component/symbols_test.go` are the dev version; this task does not change them. |
| `853a05e1` | task 35 (`7ea7497d`) | `testdata/ste-baseline.json`: the file of dev plus the one entry of task 33, so the order of dev stays. |

`git merge-base --is-ancestor origin/dev HEAD` passes at `853a05e1`. The merge of task 35 changed only the Work tab code of `frame.go`. V2 run 2 repeated the live cell compare on `853a05e1`.

## Diff scope

Each file is in the source table of the brief, or is evidence. `internal/tui/component/card.go` and `panel.go` are the component fix of the repair (the `Heavy` field of `Panel`). The README change is in the section `Manager decisions` only.

```text
$ git diff origin/dev...HEAD --stat
 .hive/spec/tui-v2/README.md                        |  16 +
 .hive/spec/tui-v2/tasks/33-screen-agents/brief.md  | 150 +++++
 .../tasks/33-screen-agents/evidence/V-gate.md      | 148 +++++
 .../tui-v2/tasks/33-screen-agents/evidence/V1.md   | 237 +++++++
 .../tui-v2/tasks/33-screen-agents/evidence/V2.md   |  90 +++
 .../tasks/33-screen-agents/evidence/VERIFIED.md    |  71 ++
 .../evidence/screens/AgentDetails.ansi             |  40 ++
 .../evidence/screens/AgentDetails.png              | Bin 0 -> 89786 bytes
 .../evidence/screens/AgentsMany.ansi               |  40 ++
 .../evidence/screens/AgentsMany.png                | Bin 0 -> 105077 bytes
 .../evidence/screens/CompactAgents.ansi            |  24 +
 .../evidence/screens/CompactAgents.png             | Bin 0 -> 43710 bytes
 .../33-screen-agents/evidence/screens/Help.ansi    |  40 ++
 .../33-screen-agents/evidence/screens/Help.png     | Bin 0 -> 104143 bytes
 .../33-screen-agents/evidence/screens/Main.ansi    |  40 ++
 .../33-screen-agents/evidence/screens/Main.png     | Bin 0 -> 55846 bytes
 .../evidence/screens/ca-enter.ansi                 |  24 +
 .../33-screen-agents/evidence/screens/ca-enter.png | Bin 0 -> 57518 bytes
 .../evidence/screens/compare-program.go.txt        |  57 ++
 .../33-screen-agents/evidence/screens/compare.txt  |   5 +
 .../33-screen-agents/evidence/screens/m-down.ansi  |  40 ++
 .../33-screen-agents/evidence/screens/m-down.png   | Bin 0 -> 56574 bytes
 .../evidence/screens/m2-enter.ansi                 |  40 ++
 .../33-screen-agents/evidence/screens/m2-enter.png | Bin 0 -> 65495 bytes
 .../33-screen-agents/evidence/screens/m2-t.ansi    |  40 ++
 .../33-screen-agents/evidence/screens/m2-t.png     | Bin 0 -> 104103 bytes
 .../33-screen-agents/evidence/screens/main-h.ansi  |  40 ++
 .../33-screen-agents/evidence/screens/main-h.png   | Bin 0 -> 48444 bytes
 .../33-screen-agents/evidence/screens/many-k.ansi  |  40 ++
 .../33-screen-agents/evidence/screens/many-k.png   | Bin 0 -> 105344 bytes
 .../33-screen-agents/evidence/screens/many-r4.ansi |  40 ++
 .../33-screen-agents/evidence/screens/many-r4.png  | Bin 0 -> 90408 bytes
 internal/tui/agents.go                             |  34 +-
 internal/tui/agents_screen.go                      | 729 +++++++++++++++++++++
 internal/tui/agents_screen_test.go                 | 393 +++++++++++
 internal/tui/agents_test.go                        | 113 ++--
 internal/tui/component/card.go                     |   6 +-
 internal/tui/component/card_test.go                |  20 +-
 internal/tui/component/fit_test.go                 |  17 +-
 internal/tui/component/panel.go                    |   6 +-
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
 54 files changed, 2494 insertions(+), 461 deletions(-)
```

The stat includes the evidence of this commit.

## Gate (`853a05e1`, from the worktree root)

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
ok  	github.com/Harrison-Blair/hive	19.050s
ok  	github.com/Harrison-Blair/hive/cmd	6.002s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.324s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.436s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.477s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.515s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.308s
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
ok  	github.com/Harrison-Blair/hive/cmd/task	1.465s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.266s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.338s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.255s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	12.747s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.542s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.885s
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
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.331s
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
ok  	github.com/Harrison-Blair/hive/internal/tui	49.000s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	18.917s
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
exit 0
```

`gofmt -l` and `go mod tidy -diff` printed nothing. Each other command exited 0.

## Note: a timing flake outside this task

`TestStartupFailureShowsInFooterAndDoesNotStopTUI` (`internal/tui/startup_test.go`) failed twice in a full package run without `-race`, before the gate. It passed in the gate, and it passed 30 times alone and 10 times alone with `-race -cpu 1`. The test sends `\x03\x03` at once, and the startup error comes in async, so the two can race under load. This task does not change the startup code. The diff of this task touches only one line of the startup files: the `graph` field of `Model` in `tui.go`.
