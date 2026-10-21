# Independent verification

- Task: `17-coordinator-start`
- Verifier: `t17-verifier-2`
- Date: 2026-10-21
- Tested commit: `54f12551`
- Verdict: **pass** under the revised decision 34 (user option A).

This report replaces the previous verdict. Git keeps the previous report.

## Items

| Item | Result | Observation |
| --- | --- | --- |
| V1 live check | pass | Bare hive started one pi coordinator. Restart closed its pane and process. The claim names the new live record. The old record is stale. |
| V2 instruction text | pass | `go test ./ -run 'STE\|Skill'` passed. I read the changed instruction lines. Only a hive prompt gives the coordinator role. The skill has no registration step. |
| V-red tests | pass | CLI tests fail against dev and pass against task code. Other new tests cannot compile against dev because their API is absent. Task tests pass. Deliberate mutations fail the relevant tests. |
| V-gate scope and gate | pass | Dev is an ancestor of HEAD. Changes match the source table and evidence. Each gate command exits 0. Format and tidy print nothing. |

The task changes startup and restart behavior. It changes no screen layout or rendered state. No design-board comparison applies.

## V1 live check

Installed herdr: 0.9.5, protocol 22, schema version 1. These values match the reference.

I built the task binary with `go build -o .hive/tmp/t17-verify2/hive .`.
I started the scratch server `t17verify2` with inherited `HERDR_*` variables removed.
The server used private XDG directories. A short symbolic link avoided the socket path limit.
The scratch workspace was `w1`, with the TUI in `w1:p1`.
The temp hive config selected pi and `opencode-go/deepseek-v4.1-flash:medium`.
Pi used a private agent directory and a temporary authentication copy.
The scratch home prohibited tool use. This check did not run the coordinator's operational startup steps.

Commands in the scratch server:

```sh
herdr workspace create --cwd SCRATCH_HOME --label t17verify2 --no-focus
herdr pane process-info --pane w1:p1
herdr pane run w1:p1 hive
herdr pane list --workspace w1
herdr pane send-keys w1:p1 2
herdr pane send-keys w1:p1 r
herdr pane send-keys w1:p1 y
herdr pane list --workspace w1
```

Before the run, process-info showed an available bash shell.
The task binary ran without `--no-startup`.

| Step | Claim | Record and process |
| --- | --- | --- |
| Start | `coordinator-g1`, generation 1, live, PID 3071853 | Live pi record in `w1:p2`. One pi coordinator pane. |
| Restart | `coordinator-g2`, generation 2, live, PID 3076092 | g1 stale. g2 live in `w1:p4`, replaces g1. Old pane absent. Old PID absent. New PID present. One pi coordinator pane. |

Logs show this order:

```text
19:32:02.521 coordinator starting claim: bare hive, generation 1
19:32:02.525 coordinator harness start: bare hive, w1:p2
19:32:06.455 coordinator record written: bare hive, coordinator-g1, PID 3071853
19:32:28.602 coordinator starting claim: tui agent restart, generation 2
19:32:28.633 coordinator harness start: tui agent restart, w1:p4
19:32:32.553 coordinator record written: tui agent restart, coordinator-g2, PID 3076092
```

The TUI reported `restarted coordinator`.
The pi response acknowledged the hive prompt and existing record. It ran no operational startup commands.
The scratch skill lacked a description field, so pi displayed a skill warning.
This warning did not block harness start, prompt delivery, or restart.
The known chat-send restriction for a directory below `projects/` also appeared. It did not block this check.

Cleanup commands:

```sh
herdr workspace close w1
herdr server stop
```

After cleanup, server PID 3069793 and coordinator PIDs 3071853 and 3076092 were absent.
No socket listener for `t17verify2` remained.
I removed the scratch files, private authentication copy, and symbolic link after recording the results.
No command targeted the real coordinator.

## V-red independent runs

Base: local `dev`, archived into a scratch directory.
`git merge-base --is-ancestor dev HEAD` exited 0. No merge was necessary.
I copied these new task tests into the base archive:

- `cmd/agent_refuse_root_test.go`
- `internal/agent/coordinator_claim_test.go`
- `internal/tui/startup_claim_test.go`
- `internal/lib/store/remove_test.go`

For each package, I ran:

```sh
go test -count=1 PACKAGE -run 'TestAgentCLIRefusesHiveOwnedRoles|TestCoordinator|TestStartupTakesStartingClaim|TestStartupRefusesWhileAnother|TestRemoveIn'
```

| Base package | Result |
| --- | --- |
| `./cmd` | Exit 1. Each refused-role case fails. The first coordinator registration exits 0 and changes records and claim. Spawn reaches shell waits. |
| `./internal/agent` | Exit 1. No `ClaimCoordinatorStart`, claim status, or starting-claim fields. |
| `./internal/tui` | Exit 1. No `CoordinatorLaunch` or starting-claim fields. |
| `./internal/lib/store` | Exit 1. No `RemoveIn`. |

Compile failures establish API absence. They do not prove each runtime assertion against base code.

Task green command:

```sh
go test -count=1 ./cmd/ ./internal/agent/ ./internal/tui/ ./internal/lib/store/ -run 'TestAgentCLIRefusesHiveOwnedRoles|TestCoordinator|TestStartCoordinator|TestStartupTakesStartingClaim|TestStartupRefusesWhileAnother|TestRemoveIn' -v
```

All selected tests passed, including each new decision 34 test.
The tests cover simultaneous starts, prompt failure, both launcher crash points, normal restart, and unresolved restart.

I made four deliberate mutations in a separate task archive. I restored each file after its run.

| Mutation | Command suffix after `go test -count=1` | Observed failure |
| --- | --- | --- |
| Refuse an absent claim | `./internal/agent -run TestCoordinatorRestartWithNoClaim -v` | `Restart() of a live record, claimed=false: MUTATION: absent claim refused` |
| Accept only live or unresolved records | `./internal/agent -run 'TestCoordinatorRestart(Stale\|Stopped\|Exited)$' -v` | All three tests fail with `cannot restart coordinator with status ...`. |
| Ignore the record named by the claim | `./internal/agent -run TestCoordinatorRestartRefuses -v` | Both refusal tests fail. Unexpected pane/tab requests replace the required refusal. |
| Make `RemoveIn` return nil without deletion | `./internal/lib/store -run TestRemoveIn -v` | Deletion and missing-record tests fail. The record remains present. |

Each mutation command exited 1.
After restoration, this command passed in the task archive:

```sh
go test -count=1 ./internal/agent ./internal/lib/store -run 'TestCoordinatorRestart|TestRemoveIn' -v
```

The delivered source files were unchanged throughout verification.

## V-gate

`git diff dev...HEAD --stat`: 31 files changed, 2651 insertions, 115 deletions.
The diff contains the named source areas, the revised decision 34, the brief, and evidence.
`internal/lib/store/lock.go` and `cmd/task_lifecycle_test.go` retain the exceptions recorded by the previous verifier.

Exact commands:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go mod tidy -diff
go vet ./...
golangci-lint run ./...
go test -race ./...
```

Full output follows. The shell added command labels and exit markers.

```text
$ gofmt -l
EXIT=0
$ go mod tidy -diff
EXIT=0
$ go vet ./...
EXIT=0
$ golangci-lint run ./...
0 issues.
EXIT=0
$ go test -race ./...
ok  	github.com/Harrison-Blair/hive	14.936s
ok  	github.com/Harrison-Blair/hive/cmd	5.953s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.155s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.333s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.297s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.125s
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
ok  	github.com/Harrison-Blair/hive/cmd/task	1.324s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.084s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.146s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.041s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	10.654s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.397s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.861s
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
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.276s
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
ok  	github.com/Harrison-Blair/hive/internal/tui	51.438s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	(cached)
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
EXIT=0
```
