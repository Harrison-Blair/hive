# V-gate Diff scope and gate

dev was merged into the task branch (merge 9565e349, dev = origin/dev =
77021a80). Run from the worktree root.

## `git diff dev...HEAD --stat`

```text
 .hive/spec/tui-v2/README.md                      |   4 +-
 .hive/spec/tui-v2/tasks/23-components-2/brief.md | 177 +++++++++++++++++++++++
 go.mod                                           |  11 ++
 go.sum                                           |  29 ++++
 internal/tui/component/card.go                   |  52 +++++++
 internal/tui/component/card_test.go              |  33 +++++
 internal/tui/component/chat.go                   | 134 +++++++++++++++++
 internal/tui/component/chat_test.go              | 123 ++++++++++++++++
 internal/tui/component/details.go                |  31 ++++
 internal/tui/component/details_test.go           |  36 +++++
 internal/tui/component/doc.go                    |  10 ++
 internal/tui/component/filterbar.go              |  54 +++++++
 internal/tui/component/filterbar_test.go         |  47 ++++++
 internal/tui/component/fit_test.go               |  41 ++++++
 internal/tui/component/golden_test.go            |  22 +++
 internal/tui/component/input.go                  |  34 +++++
 internal/tui/component/input_test.go             |  43 ++++++
 internal/tui/component/list.go                   | 123 ++++++++++++++++
 internal/tui/component/list_test.go              |  75 ++++++++++
 internal/tui/component/markdown.go               | 104 +++++++++++++
 internal/tui/component/markdown_style.go         |  55 +++++++
 internal/tui/component/markdown_table_test.go    |  94 ++++++++++++
 internal/tui/component/markdown_test.go          | 106 ++++++++++++++
 internal/tui/component/notice.go                 |  15 ++
 internal/tui/component/notice_test.go            |  24 +++
 internal/tui/component/popup.go                  |  64 ++++++++
 internal/tui/component/popup_test.go             |  58 ++++++++
 testdata/ste-baseline.json                       |   1 +
 28 files changed, 1598 insertions(+), 2 deletions(-)
```

Each file is in the source table of the brief: `internal/tui/component/` and
its tests, `go.mod` and `go.sum` (Glamour and the modules it needs; goldmark is
a direct import), `.hive/spec/tui-v2/README.md` (decisions 32 and 33 only),
`testdata/ste-baseline.json` (the brief with 0), and the brief copy. The
evidence files are added by the commit of this file.

## Full gate

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
ok  	github.com/Harrison-Blair/hive	17.847s
ok  	github.com/Harrison-Blair/hive/cmd	6.005s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.140s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.332s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.276s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.122s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.805s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.815s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.290s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.074s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.142s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.037s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	12.707s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.507s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.394s
ok  	github.com/Harrison-Blair/hive/internal/doctor	5.149s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.297s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.012s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.296s
ok  	github.com/Harrison-Blair/hive/internal/land	3.715s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.020s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.232s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.018s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	7.699s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.040s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.280s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	4.009s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.180s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.045s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.013s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.012s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.249s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.162s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	27.089s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.011s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.211s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.418s
ok  	github.com/Harrison-Blair/hive/internal/project	1.109s
ok  	github.com/Harrison-Blair/hive/internal/supervise	3.280s
ok  	github.com/Harrison-Blair/hive/internal/task	10.796s
ok  	github.com/Harrison-Blair/hive/internal/tui	52.772s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	17.236s
ok  	github.com/Harrison-Blair/hive/internal/worktree	5.864s
exit 0
```
