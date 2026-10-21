# 22-components-board-fix V-gate Diff scope and gate

| Item | Value |
| --- | --- |
| Task | 22-components-board-fix |
| Item | V-gate |
| Proves | AC1, AC2, AC3 |
| Commit | 3308639a |
| Date | 2026-10-21 |
| Agent | worker, Claude Code, claude-opus-5-5 |

## Diff scope

`git merge --no-edit dev` merged `9d4def4d` without a conflict. Then:

```sh
git diff dev...HEAD --stat
```

```text
 .hive/spec/tui-v2/design/Components.dc.html        |   4 +-
 .hive/spec/tui-v2/design/Components.grid           |   4 +-
 .hive/spec/tui-v2/design/support.js                |  38 +++++++
 .../tui-v2/tasks/22-components-board-fix/brief.md  | 126 +++++++++++++++++++++
 .../22-components-board-fix/evidence/V-gate.md     | 120 ++++++++++++++++++++
 .../tasks/22-components-board-fix/evidence/V1.md   |  26 +++++
 .../tasks/22-components-board-fix/evidence/V2.md   | 119 +++++++++++++++++++
 .../22-components-board-fix/evidence/VERIFIED.md   |  60 ++++++++++
 .../evidence/verifier-approved.png                 | Bin 0 -> 9879 bytes
 .../evidence/verifier-copy.png                     | Bin 0 -> 9879 bytes
 .../evidence/verifier-gate.txt                     |  94 +++++++++++++++
 .../evidence/verifier-grid.txt                     |  50 ++++++++
 testdata/ste-baseline.json                         |   1 +
 13 files changed, 638 insertions(+), 4 deletions(-)
```

Only the files of the source table and the task evidence changed.
`.gitignore` did not change, because Git does not ignore `support.js`.
The `verifier-*` files and `VERIFIED.md` came from the verifier in commit
`81b41902`.

## Gate

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go'); echo "exit $?"
go mod tidy -diff; echo "exit $?"
go vet ./...; echo "exit $?"
golangci-lint run ./...; echo "exit $?"
go test -race ./...; echo "exit $?"
```

```text
commit 3308639a
$ gofmt -l
exit 0
$ go mod tidy -diff
exit 0
$ go vet ./...
exit 0
$ golangci-lint run ./...
0 issues.
exit 0
$ go test -race ./...
ok  	github.com/Harrison-Blair/hive	13.756s
ok  	github.com/Harrison-Blair/hive/cmd	3.072s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.113s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.252s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.211s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.094s
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
ok  	github.com/Harrison-Blair/hive/cmd/task	1.229s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.052s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.102s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.019s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	10.148s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.334s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.636s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	(cached)
ok  	github.com/Harrison-Blair/hive/internal/homeinit	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/land	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.017s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.260s
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
ok  	github.com/Harrison-Blair/hive/internal/supervise	2.646s
ok  	github.com/Harrison-Blair/hive/internal/task	(cached)
ok  	github.com/Harrison-Blair/hive/internal/tui	48.890s
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
exit 0
```

## Result

`gofmt -l` and `go mod tidy -diff` print nothing. Each other command exits 0.
