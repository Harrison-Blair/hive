# V-gate Diff scope and gate

Repair run after the verify fail of `t17-verifier-1`. I merged `origin/dev` (`0fe88b95`) into the task branch. This worktree has no local `dev` branch, so the diff uses `origin/dev`.

## Diff scope

```text
$ git diff origin/dev...HEAD --stat
 .agents/skills/hive-coordinator/SKILL.md           |  32 +-
 .hive/spec/general-agent/README.md                 |   2 +-
 .../tasks/17-coordinator-start/brief.md            | 155 ++++++
 .../tasks/17-coordinator-start/evidence/V-gate.md  | 138 +++++
 .../tasks/17-coordinator-start/evidence/V-red.md   | 175 +++++++
 .../tasks/17-coordinator-start/evidence/V1.md      |  99 ++++
 .../tasks/17-coordinator-start/evidence/V2.md      |  74 +++
 .../17-coordinator-start/evidence/VERIFIED.md      | 228 ++++++++
 AGENTS.md                                          |  12 +-
 cmd/agent/register/register_test.go                |  41 +-
 cmd/agent/spawn/cli_test.go                        |  28 +-
 cmd/agent/spawn/spawn.go                           |   2 +-
 cmd/agent_refuse_root_test.go                      | 134 +++++
 cmd/agent_root_test.go                             |  10 +-
 cmd/task_lifecycle_test.go                         |   6 +-
 internal/agent/coordinator_claim.go                | 416 +++++++++++++++
 internal/agent/coordinator_claim_test.go           | 576 +++++++++++++++++++++
 internal/agent/doc.go                              |   4 +-
 internal/agent/register.go                         |  29 +-
 internal/agent/restart.go                          | 120 ++++-
 internal/agent/restart_test.go                     |  11 +-
 internal/agent/spawn.go                            |  10 +
 internal/agent/start_coordinator.go                |  92 +++-
 internal/agent/start_coordinator_test.go           |  30 +-
 internal/lib/records/claim.go                      |  11 +
 internal/lib/store/lock.go                         |  16 +
 internal/lib/store/remove_test.go                  |  66 +++
 internal/tui/startup.go                            |  37 +-
 internal/tui/startup_claim_test.go                 |  95 ++++
 internal/tui/startup_test.go                       |  14 +-
 testdata/ste-baseline.json                         |   1 +
 31 files changed, 2549 insertions(+), 115 deletions(-)
```

Each file is in the source table of the brief, or is evidence. `internal/lib/store/lock.go` and `cmd/task_lifecycle_test.go` are the two exceptions that the manager approved in the first run. `.hive/spec/general-agent/README.md` changes only decision 34, as the changed brief permits.

## Gate

Run from the worktree root. Full output:

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
ok  	github.com/Harrison-Blair/hive	14.666s
ok  	github.com/Harrison-Blair/hive/cmd	5.678s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.144s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.276s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.236s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.115s
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
ok  	github.com/Harrison-Blair/hive/cmd/task	1.241s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.061s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.131s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.026s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	10.310s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.358s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.726s
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
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.217s
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
ok  	github.com/Harrison-Blair/hive/internal/tui	56.801s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	(cached)
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
exit=0
```
