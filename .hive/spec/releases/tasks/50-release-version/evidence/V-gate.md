# 50-release-version V-gate Gate

| Item | Value |
| --- | --- |
| Task | 50-release-version |
| Item | V-gate |
| Proves | AC1 |
| Commit | 2b0762b3 |
| Date | 2026-10-22 |
| Agent | worker, Claude Code, claude-opus-5-5 |

## Run

```sh
git fetch origin dev && git merge --no-edit origin/dev
```

```text
Already up to date.
```

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go'); echo "exit $?"
go mod tidy -diff; echo "exit $?"
go vet ./...; echo "exit $?"
golangci-lint run ./...; echo "exit $?"
go test -race ./...; echo "exit $?"
```

```text
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
ok  	github.com/Harrison-Blair/hive	23.215s
ok  	github.com/Harrison-Blair/hive/cmd	6.815s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.397s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.508s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.576s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.588s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.410s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.288s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.338s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.598s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.340s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.443s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.317s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	15.080s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.694s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.401s
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.990s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.322s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.012s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.301s
ok  	github.com/Harrison-Blair/hive/internal/land	3.727s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.019s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.194s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.016s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	7.735s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.038s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.282s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	4.049s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.159s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.042s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.014s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.014s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.265s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.171s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	28.798s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.016s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.337s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.423s
ok  	github.com/Harrison-Blair/hive/internal/project	1.100s
ok  	github.com/Harrison-Blair/hive/internal/supervise	3.670s
ok  	github.com/Harrison-Blair/hive/internal/task	10.204s
ok  	github.com/Harrison-Blair/hive/internal/tui	42.820s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	21.449s
ok  	github.com/Harrison-Blair/hive/internal/worktree	5.762s
exit 0
```

## Result

`gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0.

## Repair 2026-10-22

`hive land` had a merge conflict in `testdata/ste-baseline.json` after
`52-readme` landed. The worker merged `origin/dev` into the task branch in
merge commit `7ea79096`. The resolution keeps each `dev` entry and adds the
entry `.hive/spec/releases/tasks/50-release-version/brief.md` with 0 next to
the entry `.hive/spec/releases/tasks/52-readme/brief.md` with 0. No other
file changed. The gate ran again after the merge:

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
$ go mod tidy -diff
$ go vet ./...          exit 0
$ golangci-lint run ./...
0 issues.
$ go test -race ./...   exit 0, no failed package
```
