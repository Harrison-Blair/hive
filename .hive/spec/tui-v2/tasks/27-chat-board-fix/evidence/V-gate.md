# V-gate Diff scope and gate

## Diff scope

`dev` merged into the task branch (already up to date at 0fe88b95).

```text
$ git diff dev...HEAD --stat
 .hive/spec/tui-v2/design/Chat.dc.html              |   6 +-
 .hive/spec/tui-v2/design/Chat.grid                 |  24 ++---
 .hive/spec/tui-v2/design/CompactChat.dc.html       |   8 +-
 .hive/spec/tui-v2/design/CompactChat.grid          |  26 ++---
 .hive/spec/tui-v2/tasks/27-chat-board-fix/brief.md | 115 +++++++++++++++++++++
 testdata/ste-baseline.json                         |   1 +
 6 files changed, 148 insertions(+), 32 deletions(-)
```

Only the files of the source table changed. The evidence is added in the next commit.
`ste_test.go` needed no edit: its glob `.hive/spec/*/tasks/*/brief.md`
already covers the brief copy.

STE baseline, red run without the entry:

```text
--- FAIL: TestSTERatchet
    ste_test.go:571: .hive/spec/tui-v2/tasks/27-chat-board-fix/brief.md: no baseline entry (now 0); add it to testdata/ste-baseline.json
FAIL	github.com/Harrison-Blair/hive
```

Green run with the entry: `ok  github.com/Harrison-Blair/hive`.

## Gate (full output, from the worktree root, after the final export)

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
ok  	github.com/Harrison-Blair/hive	17.280s
ok  	github.com/Harrison-Blair/hive/cmd	5.958s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.118s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.381s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.267s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.116s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.808s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.836s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.260s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.066s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.124s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.025s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	11.791s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.478s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.387s
ok  	github.com/Harrison-Blair/hive/internal/doctor	5.213s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.317s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.017s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.311s
ok  	github.com/Harrison-Blair/hive/internal/land	3.591s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.019s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.210s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.016s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	7.779s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.038s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.311s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	3.926s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.183s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.044s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.014s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.013s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.256s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.164s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	28.094s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.011s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.710s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.420s
ok  	github.com/Harrison-Blair/hive/internal/project	1.099s
ok  	github.com/Harrison-Blair/hive/internal/supervise	3.283s
ok  	github.com/Harrison-Blair/hive/internal/task	9.453s
ok  	github.com/Harrison-Blair/hive/internal/tui	58.251s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	7.264s
ok  	github.com/Harrison-Blair/hive/internal/worktree	5.131s
exit 0
```
