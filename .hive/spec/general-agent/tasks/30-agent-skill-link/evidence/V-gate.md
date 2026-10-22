# V-gate Diff scope and gate

Proves: each criterion.

## Diff scope

```text
$ git merge --no-edit dev
Already up to date.
$ git diff dev...HEAD --stat
 .claude/skills/hive-agent                          |  1 +
 .gitignore                                         |  1 +
 .../tasks/30-agent-skill-link/brief.md             | 83 ++++++++++++++++++++++
 testdata/ste-baseline.json                         |  1 +
 4 files changed, 86 insertions(+)
```

Each changed file is in the source table. This evidence folder is added in the
next commit. No test file changed, because no test lists the skill links (see
V1).

## Gate, from the worktree root

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
ok  	github.com/Harrison-Blair/hive	29.058s
ok  	github.com/Harrison-Blair/hive/cmd	7.926s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.205s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.369s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.458s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.485s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.156s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.322s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.331s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.459s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.099s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.211s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.055s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	17.822s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.682s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.535s
ok  	github.com/Harrison-Blair/hive/internal/doctor	5.869s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.498s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.017s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.388s
ok  	github.com/Harrison-Blair/hive/internal/land	5.269s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.039s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.373s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.021s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	8.712s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.062s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.414s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	4.556s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.292s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.056s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.021s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.033s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.392s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.194s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	32.625s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.021s
ok  	github.com/Harrison-Blair/hive/internal/logs	5.925s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.522s
ok  	github.com/Harrison-Blair/hive/internal/project	1.146s
ok  	github.com/Harrison-Blair/hive/internal/supervise	3.946s
ok  	github.com/Harrison-Blair/hive/internal/task	11.786s
ok  	github.com/Harrison-Blair/hive/internal/tui	77.740s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	29.539s
ok  	github.com/Harrison-Blair/hive/internal/worktree	7.607s
exit 0
```

`gofmt -l` and `go mod tidy -diff` printed nothing. Each other command exited 0.
