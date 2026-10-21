# Independent verification

- Task: `17-coordinator-start`
- Verifier: `t17-verifier-1`
- Date: 2026-10-21
- Tested commit: `e4eea11a`
- Verdict: **fail** (AC4, decision 34).

## Items

| Item | Result | Observation |
| --- | --- | --- |
| V1 live start and restart | pass | Isolated pi coordinator started once. Restart closed the old pane and process. The new claim names the live record. |
| V2 instruction text | pass | `go test ./ -run 'STE\|Skill'` passed. Changed instructions require a hive prompt for the coordinator role. The skill has no registration step. |
| V-red base, task, and mutation | pass | CLI refusal tests fail against dev and pass against task code. Other new tests cannot compile against dev because the new API is absent. All new tests pass on task code. A deliberate RemoveIn mutation fails both store tests and the coordinator cleanup test. |
| V-gate scope and gate | pass | Dev is an ancestor of HEAD. Diff scope matches the brief and the manager's two exceptions. All five gate commands pass. |
| AC4 claim guard | fail | With no coordinator claim, restart creates a starting claim and marks the old record stale. Decision 34 requires the claim to name the old record before a change. |

The task changes startup and restart behavior. It does not change a screen layout or a rendered state. No design-board comparison applies.

## Failure: restart accepts a missing claim

Location: `internal/agent/coordinator_claim.go`, `claimCoordinator`.

The claim-to-record check is inside `if hasClaim`. When the claim is absent, the code skips that check. It marks the requested record stale and writes a new starting claim.

Decision 34 requires the restart to check that the claim names the requested record under the store lock. A missing claim cannot name that record. The restart must refuse without changes.

I reproduced the failure in an archived task checkout. No delivered source file was changed. This regression calls the actual claim function and store. It has no mocked dependency:

```go
func TestVerifierRestartRequiresClaimThatNamesRecord(t *testing.T) {
    home := testHome(t)
    old := records.Agent{
        Key: "coordinator-g1", Name: "coordinator", Role: "coordinator",
        PaneID: "w1:p1", Status: records.AgentLive,
        Generation: 1, StartedByHive: true,
    }
    writeRecord(t, home, old)
    launch, err := claimCoordinator(context.Background(), home, &old, nil)
    if err == nil {
        claim, _ := readCoordinatorClaim(t, home)
        after := readAgent(t, home, old.Key)
        t.Fatalf("restart accepted missing claim: launch=%v claim=%#v old_status=%s; decision 34 requires refusal with unchanged state", launch != nil, claim, after.Status)
    }
}
```

Command: `go test -count=1 ./internal/agent -run TestVerifierRestartRequiresClaimThatNamesRecord -v`

```text
=== RUN   TestVerifierRestartRequiresClaimThatNamesRecord
    verifier_missing_claim_test.go:17: restart accepted missing claim: launch=true claim=records.Claim{Key:"coordinator", Agent:"", Generation:0x2, PID:2962313, PIDStart:0x1b7c4d5, ClaimedAt:"2026-10-21T23:22:17.711737698Z", Status:"starting", PaneID:""} old_status=stale; decision 34 requires refusal with unchanged state
--- FAIL: TestVerifierRestartRequiresClaimThatNamesRecord (0.01s)
FAIL
FAIL github.com/Harrison-Blair/hive/internal/agent 0.015s
FAIL
```

Repair: refuse a restart when the claim is absent. Add a regression that checks the error and unchanged claim/record state.

## V1: independent live check

Installed herdr: 0.9.5, protocol 22, schema version 1. These values match the reference.

I built the task binary. I started a private herdr server with private XDG directories and session `t17verify`. Its workspace was `w1`. Its root pane was `w1:p1`. A short symbolic link avoided the Unix socket path-length limit. No operation targeted the real coordinator.

The temp config selected pi and `opencode-go/deepseek-v4.1-flash:medium`. The temp home had a local coordinator skill that required only a READY reply and prohibited tool use. This kept the harness from operating the real home. The inherited pi harness was used. It created two task-specific session files, which were removed after the check.

I ran the task binary without `--no-startup` in `w1:p1`.

| Step | Claim | Record | Harness observation |
| --- | --- | --- | --- |
| Start | `coordinator-g1`, generation 1, live, PID 2956694 | live, pi, pane `w1:p2`, started_by_hive | One coordinator pi pane. |
| Restart with `2`, `r`, `y` | `coordinator-g2`, generation 2, live, PID 2958866 | g1 stale; g2 live in `w1:p4`, replaces g1 | Old PID 2956694 absent. New PID 2958866 present. Old pane absent. One coordinator pi pane. |

The temp home again showed the known chat-send restriction for a path below `projects/`. Start and restart both succeeded.

I closed workspace `w1` and stopped the scratch server through its own socket. The recorded server and coordinator PIDs were absent after cleanup. I removed the scratch files and symbolic link.

## V-red: independent reruns

Base: archived local `dev`. `git merge-base --is-ancestor dev HEAD` exited 0. I copied these task test files into that archive:

- `cmd/agent_refuse_root_test.go`
- `internal/agent/coordinator_claim_test.go`
- `internal/tui/startup_claim_test.go`
- `internal/lib/store/remove_test.go`

For each affected package I ran `go test -count=1 PACKAGE -run 'TestAgentCLIRefusesHiveOwnedRoles|TestCoordinator|TestStartupTakesStartingClaim|TestStartupRefusesWhileAnother|TestRemoveIn'`.

Each package exited 1. The CLI test failed all refused-role cases: register accepted hive-owned roles or returned pane-busy errors, and spawn reached shell waits. The first coordinator register case exited 0 and changed the claim instead of refusing. Agent and TUI tests failed to compile because dev has no starting-claim API or claim fields. Store tests failed to compile because dev has no `RemoveIn`. These compile failures establish API absence; they do not independently prove each runtime assertion against dev.

Task green command:

```sh
go test -count=1 ./cmd/ ./internal/agent/ ./internal/tui/ ./internal/lib/store/ -run 'TestAgentCLIRefusesHiveOwnedRoles|TestCoordinator|TestStartCoordinator|TestStartupTakesStartingClaim|TestStartupRefusesWhileAnother|TestRemoveIn' -v
```

All selected tests passed. Package results:

```text
ok github.com/Harrison-Blair/hive/cmd 0.526s
ok github.com/Harrison-Blair/hive/internal/agent 1.054s
ok github.com/Harrison-Blair/hive/internal/tui 0.034s
ok github.com/Harrison-Blair/hive/internal/lib/store 0.018s
```

In a separate archived task checkout, I replaced `return os.Remove(path)` in `RemoveIn` with `return nil`. Commands:

```sh
go test -count=1 ./internal/lib/store -run TestRemoveIn -v
go test -count=1 ./internal/agent -run TestCoordinatorStartHarnessFailureRemovesClaim -v
```

Both exited 1:

```text
--- FAIL: TestRemoveInDeletesOneRecordAndKeepsOthers
    remove_test.go:29: read removed claim = <nil>, map[string]string{"key":"coordinator"}; want not exist
--- FAIL: TestRemoveInMissingRecordIsNotExist
    remove_test.go:46: RemoveIn() = <nil>; want not exist
--- FAIL: TestCoordinatorStartHarnessFailureRemovesClaim
    coordinator_claim_test.go:241: claim after the failed start = records.Claim{Key:"coordinator", Agent:"", Generation:0x1, PID:2960861, PIDStart:0x1b7ba18, ClaimedAt:"2026-10-21T23:21:50.216765925Z", Status:"starting", PaneID:"w1:p1"}; want none
```

I restored the archive's implementation. The original task checkout was unchanged. The green run above confirms the correct implementation passes these tests.

## V-gate

`git diff dev...HEAD --stat`: 29 files changed, 2206 insertions, 114 deletions. Files match the source table plus evidence and the manager-approved exceptions `internal/lib/store/lock.go` and `cmd/task_lifecycle_test.go`.

Exact commands:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go mod tidy -diff
go vet ./...
golangci-lint run ./...
go test -race ./...
```

Full output follows. Exit markers were added by the verification shell.

```text
gofmt exit=0
tidy exit=0
vet exit=0
0 issues.
lint exit=0
ok  	github.com/Harrison-Blair/hive	18.282s
ok  	github.com/Harrison-Blair/hive/cmd	6.335s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.201s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.412s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.323s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.164s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.041s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.038s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.325s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.109s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.287s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.026s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	13.081s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.455s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	5.200s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	(cached)
ok  	github.com/Harrison-Blair/hive/internal/homeinit	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/land	3.772s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.195s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.267s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/log	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.017s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.240s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/store	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/logs	(cached)
ok  	github.com/Harrison-Blair/hive/internal/memory	(cached)
ok  	github.com/Harrison-Blair/hive/internal/project	(cached)
ok  	github.com/Harrison-Blair/hive/internal/supervise	(cached)
ok  	github.com/Harrison-Blair/hive/internal/task	9.885s
ok  	github.com/Harrison-Blair/hive/internal/tui	60.945s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	(cached)
ok  	github.com/Harrison-Blair/hive/internal/worktree	5.195s
race exit=0
```
