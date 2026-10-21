# V-gate Diff scope and gate

Proves: AC1, AC2. `origin/dev` (cf5d94ef) was merged into the task branch
before the gate.

## Diff scope

```text
$ git diff origin/dev...HEAD --stat
 .../general-agent/tasks/15-harness-trust/brief.md  | 104 +++++++++++++++++++++
 internal/lib/harness/claude.go                     |   8 ++
 internal/lib/harness/codex.go                      |  13 +++
 internal/lib/harness/cursor.go                     |  14 ++-
 internal/lib/harness/harness.go                    |  19 ++++
 internal/lib/harness/harness_test.go               |   1 +
 internal/lib/harness/opencode.go                   |   6 ++
 internal/lib/harness/pi.go                         |   7 ++
 internal/lib/harness/trust_test.go                 |  48 ++++++++++
 testdata/ste-baseline.json                         |   1 +
 10 files changed, 220 insertions(+), 1 deletion(-)
```

Each changed file is in the source table of the brief: the harness package and
its tests, `testdata/ste-baseline.json` (the brief copy with 0), and the brief
copy. `ste_test.go` did not change: its glob `.hive/spec/*/tasks/*/brief.md`
already covers the brief copy. The evidence files are added in the next commit.

## Gate

Run from the worktree root. The output of `go test -race ./...` is complete.

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
ok  	github.com/Harrison-Blair/hive	16.975s
ok  	github.com/Harrison-Blair/hive/cmd	3.574s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.132s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.351s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.286s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.122s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.867s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.879s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.312s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.083s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.142s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.040s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	11.793s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.565s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.500s
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.839s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.356s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.014s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.291s
ok  	github.com/Harrison-Blair/hive/internal/land	3.596s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.016s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.177s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.014s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	7.581s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.041s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.290s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	3.981s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.162s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.042s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.015s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.017s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.234s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.179s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	27.572s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.010s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.216s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.417s
ok  	github.com/Harrison-Blair/hive/internal/project	1.099s
ok  	github.com/Harrison-Blair/hive/internal/supervise	2.847s
ok  	github.com/Harrison-Blair/hive/internal/task	9.308s
ok  	github.com/Harrison-Blair/hive/internal/tui	59.590s
ok  	github.com/Harrison-Blair/hive/internal/worktree	4.921s
exit 0
```
