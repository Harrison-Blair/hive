# V-gate Diff scope and gate

I merged `origin/dev` (`46aaff47`) into the task branch. This worktree has no local `dev` branch, so the diff uses `origin/dev`.

## Diff scope

```text
$ git diff origin/dev...HEAD --stat
 .agents/skills/hive-coordinator/SKILL.md           |  32 +-
 .../tasks/17-coordinator-start/brief.md            | 145 +++++++
 AGENTS.md                                          |  12 +-
 cmd/agent/register/register_test.go                |  41 +-
 cmd/agent/spawn/cli_test.go                        |  28 +-
 cmd/agent/spawn/spawn.go                           |   2 +-
 cmd/agent_refuse_root_test.go                      | 134 ++++++
 cmd/agent_root_test.go                             |  10 +-
 cmd/task_lifecycle_test.go                         |   6 +-
 internal/agent/coordinator_claim.go                | 416 ++++++++++++++++++
 internal/agent/coordinator_claim_test.go           | 472 +++++++++++++++++++++
 internal/agent/doc.go                              |   4 +-
 internal/agent/register.go                         |  29 +-
 internal/agent/restart.go                          | 120 +++++-
 internal/agent/restart_test.go                     |  11 +-
 internal/agent/spawn.go                            |  10 +
 internal/agent/start_coordinator.go                |  92 +++-
 internal/agent/start_coordinator_test.go           |  30 +-
 internal/lib/records/claim.go                      |  11 +
 internal/lib/store/lock.go                         |  16 +
 internal/lib/store/remove_test.go                  |  66 +++
 internal/tui/startup.go                            |  37 +-
 internal/tui/startup_claim_test.go                 |  95 +++++
 internal/tui/startup_test.go                       |  14 +-
 testdata/ste-baseline.json                         |   1 +
 25 files changed, 1720 insertions(+), 114 deletions(-)
```

The evidence files are added in the same commit as this file.

Each file is in the source table of the brief, with two notes:

- `cmd/task_lifecycle_test.go` is outside the list. `TestTaskCLIReopenRegisteredManagerUsesClaimKey` registered a manager through the CLI as setup. AC1 refuses that registration, so the test now calls the internal `agent.Register`. The assertion is the same.
- `internal/lib/store/lock.go` and `remove_test.go` add `LockedStore.RemoveIn`. The starting claim needs a removal under the store lock (decisions 32 and 33). The brief lists `internal/lib/` claim and record code for the starting claim only. This method is used only by the starting claim.

## Gate

Full output, from the worktree root:

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
exit=0
$ go mod tidy -diff
exit=0
$ go vet ./...
exit=0
$ golangci-lint run ./...
0 issues.
exit=0
$ go test -race ./...
ok  	github.com/Harrison-Blair/hive	16.138s
ok  	github.com/Harrison-Blair/hive/cmd	6.588s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.146s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.282s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.273s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.114s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.681s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.716s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.282s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.069s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.140s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.030s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	12.515s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.452s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.423s
ok  	github.com/Harrison-Blair/hive/internal/doctor	5.053s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.332s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.010s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.298s
ok  	github.com/Harrison-Blair/hive/internal/land	3.662s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.015s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.165s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.012s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	7.751s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.031s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.314s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	4.045s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.175s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.042s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.015s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.012s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.269s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.183s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	27.215s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.020s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.158s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.398s
ok  	github.com/Harrison-Blair/hive/internal/project	1.088s
ok  	github.com/Harrison-Blair/hive/internal/supervise	3.385s
ok  	github.com/Harrison-Blair/hive/internal/task	9.334s
ok  	github.com/Harrison-Blair/hive/internal/tui	53.003s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	6.890s
ok  	github.com/Harrison-Blair/hive/internal/worktree	4.968s
exit=0
```

`gofmt -l` and `go mod tidy -diff` printed nothing. Each command exited 0.
