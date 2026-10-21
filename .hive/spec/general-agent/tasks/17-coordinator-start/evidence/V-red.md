# V-red Tests red and green

Base: `fdb4eedf` (the `dev` commit at the task start). Task: `task/17-coordinator-start`.

## New tests and the criterion of each

| Test | File | Criterion |
| --- | --- | --- |
| `TestAgentCLIRefusesHiveOwnedRoles` (register coordinator, supervisor, manager; spawn coordinator, supervisor; each from an empty pane, a registered pane, the coordinator pane, and with `--pane`). The first subtest is the first failing test of the brief. | `cmd/agent_refuse_root_test.go` | AC1 |
| `TestCoordinatorStartClaimsThenRecordsThenPrompts` | `internal/agent/coordinator_claim_test.go` | AC2 |
| `TestCoordinatorStartTwoStartsAtTheSameTime` | same | AC2 |
| `TestCoordinatorStartRefusesStartingClaimWithLiveLauncher` | same | AC2 |
| `TestCoordinatorStartRefusesLiveClaim` | same | AC2 |
| `TestCoordinatorStartHarnessFailureRemovesClaim` | same | AC2 |
| `TestCoordinatorStartPromptFailureAfterRecordWrite` | same | AC2 |
| `TestCoordinatorStartLauncherCrashBeforeHarnessStart` | same | AC3 |
| `TestCoordinatorLauncherCrashAfterHarnessStartThenRestart` | same | AC3, AC4 |
| `TestCoordinatorRestartNormal` (also checks the `command` field of the start log events) | same | AC4 |
| `TestCoordinatorRestartUnresolved` | same | AC4 |
| `TestCoordinatorRestartRefusesRecordThatTheClaimDoesNotName` | same | AC4 |
| `TestStartupTakesStartingClaimBeforeCoordinatorTab` | `internal/tui/startup_claim_test.go` | AC2 (bare `hive`) |
| `TestStartupRefusesWhileAnotherLauncherStarts` | same | AC2 (bare `hive`) |
| `TestRemoveInDeletesOneRecordAndKeepsOthers`, `TestRemoveInMissingRecordIsNotExist`, `TestRemoveInAfterCallbackFails` | `internal/lib/store/remove_test.go` | AC2, AC3 (claim removal under the store lock) |

## Changed tests

- `TestStartCoordinatorSendsStartupPromptWithoutRegistering` is now `TestStartCoordinatorSendsStartupPromptAfterHiveRegisters`. The old test expected no record. Decision 32 makes hive write the record before the prompt.
- `internal/agent/restart_test.go`: the old coordinator record is now stale after a restart (decision 34). The fake herdr gives the replies for the new pane and identity checks.
- `internal/tui/startup_test.go`: the starter seams use the new launch signature.
- `cmd/agent_root_test.go`, `cmd/agent/register/register_test.go`, `cmd/agent/spawn/cli_test.go`, `cmd/task_lifecycle_test.go`: these tests used `hive agent register` or `hive agent spawn` with a coordinator or manager role only as a sample role. AC1 refuses these roles in the CLI. The tests now use the worker role, or the internal `agent.Register` for the manager claim key. Each test keeps its assertion.

## Red on the base

AC1 test on the base (excerpt). The base registers the hive-owned roles:

```text
--- FAIL: TestAgentCLIRefusesHiveOwnedRoles (80.50s)
    --- FAIL: TestAgentCLIRefusesHiveOwnedRoles/register_coordinator/empty_pane (0.03s)
        agent_refuse_root_test.go:98: status = 0, stdout = "stale agent .hive/state/agents/coordinator-g1.json\ncreated agent .hive/state/agents/coordinator-g2.json\nupdated claim .hive/state/claims/coordinator.json\n"; want a refusal
    --- FAIL: TestAgentCLIRefusesHiveOwnedRoles/register_coordinator/registered_agent_pane (0.01s)
        agent_refuse_root_test.go:102: stderr = "hive: pane busy: w1:general holds worker-demo-1\n"; want the hive-owned role refusal
    --- FAIL: TestAgentCLIRefusesHiveOwnedRoles/register_coordinator/coordinator_pane (0.02s)
        agent_refuse_root_test.go:98: status = 0, stdout = "stale agent .hive/state/agents/coordinator-g1.json\ncreated agent .hive/state/agents/coordinator-g2.json\nupdated claim .hive/state/claims/coordinator.json\n"; want a refusal
    --- FAIL: TestAgentCLIRefusesHiveOwnedRoles/register_coordinator/pane_override (0.02s)
        agent_refuse_root_test.go:98: status = 0, stdout = "stale agent .hive/state/agents/coordinator-g1.json\ncreated agent .hive/state/agents/coordinator-g2.json\nupdated claim .hive/state/claims/coordinator.json\n"; want a refusal
    --- FAIL: TestAgentCLIRefusesHiveOwnedRoles/register_supervisor/empty_pane (0.02s)
        agent_refuse_root_test.go:98: status = 0, stdout = "created agent .hive/state/agents/supervisor-g1.json\nupdated claim .hive/state/claims/supervisor.json\n"; want a refusal
    --- FAIL: TestAgentCLIRefusesHiveOwnedRoles/register_supervisor/registered_agent_pane (0.01s)
```

The AC2 to AC4 tests on the base. The base has no starting claim, no claim status, no launch type, and no `RemoveIn`. The tests do not compile:

```text
# github.com/Harrison-Blair/hive/internal/agent [github.com/Harrison-Blair/hive/internal/agent.test]
internal/agent/coordinator_claim_test.go:31:17: undefined: ClaimCoordinatorStart
internal/agent/coordinator_claim_test.go:134:13: atStart.Status undefined (type records.Claim has no field or method Status)
internal/agent/coordinator_claim_test.go:134:31: undefined: records.ClaimStarting
internal/agent/coordinator_claim_test.go:134:122: atStart.PaneID undefined (type records.Claim has no field or method PaneID)
internal/agent/coordinator_claim_test.go:147:14: atPrompt.Status undefined (type records.Claim has no field or method Status)
internal/agent/coordinator_claim_test.go:147:32: undefined: records.ClaimStarting
internal/agent/coordinator_claim_test.go:200:122: unknown field Status in struct literal of type records.Claim
===
# github.com/Harrison-Blair/hive/internal/tui [github.com/Harrison-Blair/hive/internal/tui.test]
internal/tui/startup_claim_test.go:37:51: undefined: agent.CoordinatorLaunch
internal/tui/startup_claim_test.go:45:30: claim.Status undefined (type records.Claim has no field or method Status)
internal/tui/startup_claim_test.go:45:48: undefined: records.ClaimStarting
internal/tui/startup_claim_test.go:72:110: unknown field Status in struct literal of type records.Claim
internal/tui/startup_claim_test.go:72:126: undefined: records.ClaimStarting
===
# github.com/Harrison-Blair/hive/internal/lib/store [github.com/Harrison-Blair/hive/internal/lib/store.test]
internal/lib/store/remove_test.go:22:17: locked.RemoveIn undefined (type *LockedStore has no field or method RemoveIn)
internal/lib/store/remove_test.go:43:17: locked.RemoveIn undefined (type *LockedStore has no field or method RemoveIn)
internal/lib/store/remove_test.go:63:17: held.RemoveIn undefined (type *LockedStore has no field or method RemoveIn)
```

## Green on the task branch

```text
$ go test -count=1 ./cmd/ ./internal/agent/ ./internal/tui/ ./internal/lib/store/ -run 'TestAgentCLIRefusesHiveOwnedRoles|TestCoordinator|TestStartCoordinator|TestStartupTakesStartingClaim|TestStartupRefusesWhileAnother|TestRemoveIn' -v
--- PASS: TestAgentCLIRefusesHiveOwnedRoles (0.52s)
ok  	github.com/Harrison-Blair/hive/cmd	0.526s
--- PASS: TestCoordinatorStartClaimsThenRecordsThenPrompts (0.03s)
--- PASS: TestCoordinatorStartTwoStartsAtTheSameTime (0.03s)
--- PASS: TestCoordinatorStartRefusesStartingClaimWithLiveLauncher (0.01s)
--- PASS: TestCoordinatorStartRefusesLiveClaim (0.01s)
--- PASS: TestCoordinatorStartHarnessFailureRemovesClaim (0.01s)
--- PASS: TestCoordinatorStartPromptFailureAfterRecordWrite (0.03s)
--- PASS: TestCoordinatorStartLauncherCrashBeforeHarnessStart (0.23s)
--- PASS: TestCoordinatorLauncherCrashAfterHarnessStartThenRestart (0.25s)
--- PASS: TestCoordinatorRestartNormal (0.05s)
--- PASS: TestCoordinatorRestartUnresolved (0.05s)
--- PASS: TestCoordinatorRestartRefusesRecordThatTheClaimDoesNotName (0.02s)
--- PASS: TestCoordinatorInstructionsRequireChatReplies (0.00s)
--- PASS: TestStartCoordinatorSendsStartupPromptAfterHiveRegisters (0.03s)
--- PASS: TestStartCoordinatorWaitsForHerdrReadiness (0.28s)
--- PASS: TestStartCoordinatorReadinessCancellationDoesNotPromptAgain (0.03s)
ok  	github.com/Harrison-Blair/hive/internal/agent	1.078s
--- PASS: TestStartupTakesStartingClaimBeforeCoordinatorTab (0.01s)
--- PASS: TestStartupRefusesWhileAnotherLauncherStarts (0.01s)
ok  	github.com/Harrison-Blair/hive/internal/tui	0.032s
--- PASS: TestRemoveInDeletesOneRecordAndKeepsOthers (0.01s)
--- PASS: TestRemoveInMissingRecordIsNotExist (0.00s)
--- PASS: TestRemoveInAfterCallbackFails (0.00s)
ok  	github.com/Harrison-Blair/hive/internal/lib/store	0.019s
```

## Behavior broken on purpose

Each break changed one line in the code, ran the tests, and restored the file with `git checkout`. Each break failed a test.

```text
### break: a live launcher does not block a start (decision 31)
--- FAIL: TestCoordinatorStartRefusesStartingClaimWithLiveLauncher (0.00s)
    fakeherdr_test.go:147: unexpected method "pane.process_info"
    coordinator_claim_test.go:205: claim = check pane w1:p1 of the dead launcher: bad reply: empty reply; want a refusal for the live launcher
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.043s
FAIL
### break: a harness start failure keeps the claim (decision 32)
--- FAIL: TestCoordinatorStartHarnessFailureRemovesClaim (0.02s)
    coordinator_claim_test.go:239: claim after the failed start = records.Claim{Key:"coordinator", Agent:"", Generation:0x1, PID:2878861, PIDStart:0x1b70bf6, ClaimedAt:"2026-10-21T23:14:24.4363454Z", Status:"starting", PaneID:"w1:p1"}; want none
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.019s
FAIL
### break: a prompt failure keeps the record live (decision 32)
--- FAIL: TestCoordinatorStartPromptFailureAfterRecordWrite (0.02s)
    coordinator_claim_test.go:259: record = records.Agent{Key:"coordinator-g1", Name:"coordinator", Role:"coordinator", Project:"", Task:"", Harness:"claude", Model:"m", PaneID:"w1:p1", PaneHistory:[]string(nil), WorkspaceID:"w1", TabID:"w1:t1", TerminalID:"term_w1:p1", Cwd:"/tmp/TestCoordinatorStartPromptFailureAfterRecordWrite320099462/001", Status:"live", Parent:"", Replaces:"", StatusChangedAt:"", AgentStatus:"", Finished:false, QuietAfterMin:10, Generation:0x1, StartedAt:"2026-10-21T23:14:25.417718096Z", UpdatedAt:"2026-10-21T23:14:25.417718096Z", LastError:"", Spawned:false, StartedByHive:true, ProcessPID:2879050, ProcessName:"claude", ProcessStart:0x1b70c57, AgentSessionRef:""}; want unresolved with the prompt error
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.026s
FAIL
### break: a dead launcher with a harness removes the claim (decision 33)
--- FAIL: TestCoordinatorLauncherCrashAfterHarnessStartThenRestart (0.22s)
    coordinator_claim_test.go:317: start after the crash = <nil>; want a stop that names the restart
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.220s
FAIL
### break: restart ignores the record that the claim names (decision 34)
--- FAIL: TestCoordinatorRestartRefusesRecordThatTheClaimDoesNotName (0.03s)
    fakeherdr_test.go:147: unexpected method "pane.get"
    coordinator_claim_test.go:437: Restart() = close old coordinator pane: bad reply: empty reply; want a refusal that names the claim holder
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.034s
FAIL
### break: restart keeps the old record live (decision 34)
--- FAIL: TestCoordinatorRestartNormal (0.03s)
    coordinator_claim_test.go:360: at close: claim records.Claim{Key:"coordinator", Agent:"", Generation:0x2, PID:2879634, PIDStart:0x1b70d91, ClaimedAt:"2026-10-21T23:14:28.564375467Z", Status:"starting", PaneID:""}, old record records.Agent{Key:"coordinator-g1", Name:"coordinator", Role:"coordinator", Project:"", Task:"", Harness:"", Model:"", PaneID:"w1:p1", PaneHistory:[]string(nil), WorkspaceID:"w1", TabID:"w1:t1", TerminalID:"term_w1:p1", Cwd:"", Status:"live", Parent:"", Replaces:"", StatusChangedAt:"", AgentStatus:"", Finished:false, QuietAfterMin:0, Generation:0x1, StartedAt:"", UpdatedAt:"2026-10-21T23:14:28.561323093Z", LastError:"", Spawned:false, StartedByHive:true, ProcessPID:2879641, ProcessName:"claude", ProcessStart:0x1b70d91, AgentSessionRef:""}; want a starting claim and a stale old record
--- FAIL: TestCoordinatorRestartUnresolved (0.02s)
    coordinator_claim_test.go:365: at close: claim records.Claim{Key:"coordinator", Agent:"", Generation:0x2, PID:2879634, PIDStart:0x1b70d91, ClaimedAt:"2026-10-21T23:14:28.590416618Z", Status:"starting", PaneID:""}, old record records.Agent{Key:"coordinator-g1", Name:"coordinator", Role:"coordinator", Project:"", Task:"", Harness:"", Model:"", PaneID:"w1:p1", PaneHistory:[]string(nil), WorkspaceID:"w1", TabID:"w1:t1", TerminalID:"term_w1:p1", Cwd:"", Status:"unresolved", Parent:"", Replaces:"", StatusChangedAt:"", AgentStatus:"", Finished:false, QuietAfterMin:0, Generation:0x1, StartedAt:"", UpdatedAt:"2026-10-21T23:14:28.587479793Z", LastError:"", Spawned:false, StartedByHive:true, ProcessPID:2879644, ProcessName:"claude", ProcessStart:0x1b70d95, AgentSessionRef:""}; want a starting claim and a stale old record
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.064s
FAIL
### break: register accepts hive-owned roles (decision 30)
--- FAIL: TestAgentCLIRefusesHiveOwnedRoles (0.23s)
        agent_refuse_root_test.go:113: status = 0, stdout = "stale agent .hive/state/agents/coordinator-g1.json\ncreated agent .hive/state/agents/coordinator-g2.json\nupdated claim .hive/state/claims/coordinator.json\n"; want a refusal
        agent_refuse_root_test.go:117: stderr = "hive: pane busy: w1:general holds worker-demo-1\n"; want the hive-owned role refusal
        agent_refuse_root_test.go:113: status = 0, stdout = "stale agent .hive/state/agents/coordinator-g1.json\ncreated agent .hive/state/agents/coordinator-g2.json\nupdated claim .hive/state/claims/coordinator.json\n"; want a refusal
        agent_refuse_root_test.go:113: status = 0, stdout = "stale agent .hive/state/agents/coordinator-g1.json\ncreated agent .hive/state/agents/coordinator-g2.json\nupdated claim .hive/state/claims/coordinator.json\n"; want a refusal
        agent_refuse_root_test.go:113: status = 0, stdout = "created agent .hive/state/agents/supervisor-g1.json\nupdated claim .hive/state/claims/supervisor.json\n"; want a refusal
        agent_refuse_root_test.go:117: stderr = "hive: pane busy: w1:general holds worker-demo-1\n"; want the hive-owned role refusal
        agent_refuse_root_test.go:117: stderr = "hive: pane busy: w1:coord holds coordinator\n"; want the hive-owned role refusal
        agent_refuse_root_test.go:113: status = 0, stdout = "created agent .hive/state/agents/supervisor-g1.json\nupdated claim .hive/state/claims/supervisor.json\n"; want a refusal
        agent_refuse_root_test.go:113: status = 0, stdout = "created agent .hive/state/agents/manager-demo-g1.json\nupdated claim .hive/state/claims/manager-demo.json\n"; want a refusal
        agent_refuse_root_test.go:117: stderr = "hive: pane busy: w1:general holds worker-demo-1\n"; want the hive-owned role refusal
        agent_refuse_root_test.go:117: stderr = "hive: pane busy: w1:coord holds coordinator\n"; want the hive-owned role refusal
### break: bare hive skips the starting claim (decision 31)
--- FAIL: TestStartupTakesStartingClaimBeforeCoordinatorTab (0.00s)
    startup_claim_test.go:46: claim at coordinator start = records.Claim{Key:"", Agent:"", Generation:0x0, PID:0, PIDStart:0x0, ClaimedAt:"", Status:"", PaneID:""}, <nil>; want the starting claim of this launcher
--- FAIL: TestStartupRefusesWhileAnotherLauncherStarts (0.02s)
    startup_claim_test.go:87: setupSessionWith() = <nil>; want the starting claim refusal
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui	0.022s
FAIL
```

The log `command` fix (commit `9dd2954c`) also went red first: `TestCoordinatorRestartNormal` failed with `"coordinator harness start" command = bare hive; want tui agent restart`, then passed after the fix.
