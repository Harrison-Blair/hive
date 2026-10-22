# V-gate Diff scope and gate

Proves AC1, AC2, and AC3. Run on 2026-10-21 from the worktree root, after
`git merge dev` (dev at fbca94d5; the merge said `Already up to date`).

## AC3 test-first

The first failing test is `TestCheckAcceptsTheHeavyBorder` in
`internal/boardexport/check_test.go`. It checks a board row with
`┏━┯━┓┃┗┷┛┡┨`.

Red run, before the change of the symbol list:

```text
$ go test ./internal/boardexport/ -run Heavy -v
=== RUN   TestCheckAcceptsTheHeavyBorder
    check_test.go:52: board TooSmall: row 0 col 0: symbol "┏" is not in the list of decision 13
--- FAIL: TestCheckAcceptsTheHeavyBorder (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/boardexport	0.002s
FAIL
```

Green run, after the change:

```text
$ go test ./internal/boardexport/ ./internal/tui/component/ -v -run 'Heavy|SymbolWidth'
=== RUN   TestCheckAcceptsTheHeavyBorder
--- PASS: TestCheckAcceptsTheHeavyBorder (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/boardexport	0.004s
=== RUN   TestSymbolWidth
--- PASS: TestSymbolWidth (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/tui/component	0.020s
```

`TestSymbolWidth` now lists 39 symbols. It gives width 1 for each new
symbol. It also passed with the new symbols before the list change, because
it checks only the width; its count check fails when a symbol is removed
from the test list without a change to the count.

New symbols: `┏ ┓ ┗ ┛` (the heavy corners of `lipgloss.ThickBorder`, which
`component.Panel` with `Heavy` draws for `component.Card` on
`task/33-screen-agents`), and `┯ ┷ ┡ ┨` (the mixed joints of
`internal/tui/agents_screen.go` lines 295, 325, 333, 347 on that branch).
`━` and `┃` were already in the list. The boards use `┏ ┓ ┗ ┛ ┯ ┨`.

The three lists agree:

```text
$ (README decision 13, check.go symbols, symbols_test.go Fields; spaces removed)
lists agree: ─│╭╮╰╯├┤┼┴┬└━┃┏┓┗┛┯┷┡┨●◐○▶✓•·–—…↑↓←→›█×
```

## Diff scope and gate

Only the files of the source table and the brief changed. The evidence is
added after this run.

```text
$ git diff dev...HEAD --stat
 .hive/spec/tui-v2/README.md                        |   2 +-
 .hive/spec/tui-v2/design/AgentsMany.dc.html        |   2 +-
 .hive/spec/tui-v2/design/AgentsMany.grid           |   8 +-
 .hive/spec/tui-v2/design/CompactAgents.dc.html     |   2 +-
 .hive/spec/tui-v2/design/CompactAgents.grid        |   8 +-
 .hive/spec/tui-v2/design/Main.dc.html              |   2 +-
 .hive/spec/tui-v2/design/Main.grid                 |   8 +-
 .../tui-v2/tasks/45-selected-card-boards/brief.md  | 123 +++++++++++++++++++++
 internal/boardexport/check.go                      |   2 +-
 internal/boardexport/check_test.go                 |   7 ++
 internal/tui/component/symbols_test.go             |   6 +-
 testdata/ste-baseline.json                         |   1 +
 12 files changed, 151 insertions(+), 20 deletions(-)

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
ok  	github.com/Harrison-Blair/hive	24.017s
ok  	github.com/Harrison-Blair/hive/cmd	7.224s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.398s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.547s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.584s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.593s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.427s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.115s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.182s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.530s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.326s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.405s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.312s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	15.566s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.668s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.413s
ok  	github.com/Harrison-Blair/hive/internal/doctor	5.332s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.451s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.013s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.330s
ok  	github.com/Harrison-Blair/hive/internal/land	3.945s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.023s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.241s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.014s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	8.016s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.049s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.353s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	4.328s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.213s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.049s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.016s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.014s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.271s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.179s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	33.400s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.012s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.741s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.516s
ok  	github.com/Harrison-Blair/hive/internal/project	1.092s
ok  	github.com/Harrison-Blair/hive/internal/supervise	3.955s
ok  	github.com/Harrison-Blair/hive/internal/task	10.227s
ok  	github.com/Harrison-Blair/hive/internal/tui	55.559s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	22.474s
ok  	github.com/Harrison-Blair/hive/internal/worktree	5.361s
exit 0
```
