# V-red Tests red and green

## New tests and the criterion of each

| Test | File | Criterion |
| --- | --- | --- |
| `TestInboxSendAgentOverrideInPaneFailsWithoutRecord` (first failing test of the brief) | `cmd/inbox_root_test.go` | AC1 |
| `TestInboxOverridesInPaneFailAndChangeNothing` (read `--agent`, read `--pane`, send `--pane`) | `cmd/inbox_root_test.go` | AC1 |
| `TestInboxAgentInUnregisteredPaneFails` | `cmd/inbox_root_test.go` | AC1 |
| `TestCallerAgentOverrideInPaneFailsForCallerFlags` (task complete, verify, reopen, edit; goal create, edit; chat send, reply) | `cmd/caller_root_test.go` | AC1 |
| `TestTaskCompleteAgentOverrideFromOtherPaneLeavesTask` | `cmd/caller_root_test.go` | AC1 |
| `TestTaskCompleteOwnAgentInPaneSucceeds` (guard: own name still works) | `cmd/caller_root_test.go` | AC1 |
| `TestInboxSendBetweenAgentInboxesInPanes` | `cmd/inbox_root_test.go` | AC2 |
| `TestInboxCoordinatorReadsAgentInboxWithoutConsume` | `cmd/inbox_root_test.go` | AC2 |
| `TestInboxAgentCannotReadOtherAgentInbox` | `cmd/inbox_root_test.go` | AC2 |
| `TestReadDefaultsWorkerToNameInbox` (replaces `TestReadRejectsWorkerWithoutInbox`, which tested the old rule) | `internal/inbox/read_send_test.go` | AC2 |
| `TestNudgeIdleWorkerGetsPromptForOwnInbox` | `internal/supervise/routing_test.go` | AC2, AC4 |
| `TestRoutingUsualParentsMatchCurrentRule` (worker, verifier, manager) | `internal/supervise/routing_test.go` | AC4 (decision 29) |
| `TestRoutingWorkerOfCoordinatorGoesToCoordinator` | `internal/supervise/routing_test.go` | AC4 (decision 29) |
| `TestRoutingWorkerReachesRestartedManager` | `internal/supervise/routing_test.go` | AC4 (replaces chain) |
| `TestRoutingParentWithoutLiveSuccessorGoesToCoordinator` (stopped, missing, no parent) | `internal/supervise/routing_test.go` | AC4 |
| `TestDoctorListsUnreadRecordsWithoutLiveOwner` | `internal/doctor/orphan_inbox_test.go` | AC4 |

## Changed tests

- `TestReadRefusalsLogOneErrorEvent`: the second refusal was "role has no inbox: worker". AC2 gives each worker an inbox, so the test now uses the worker read of `manager-one`, which stays refused.
- 13 supervisor tests in `events_test.go`, `reconcile_test.go`, and `fix_test.go` use a worker fixture with no `parent`. Decision 28 sends the nudge of such a record to `coordinator`, so these tests read `coordinator` in place of `manager-demo`. Each check stays the same. The "want zero" checks also read `coordinator`, so they are not vacuous.
- `TestRunChecksWithFakeHerdr`: 18 checks, and the new check is an optional ok.

## Red on the base

The new tests ran in a scratch worktree of `origin/dev` (`.hive/tmp/base`, now removed).
`read_send_test.go` was copied with the old 3-argument `Send` call.

```text
--- FAIL: TestCallerAgentOverrideInPaneFailsForCallerFlags (0.10s)
    --- FAIL: TestCallerAgentOverrideInPaneFailsForCallerFlags/task_complete (0.01s)
        caller_root_test.go:52: status = 1, stdout = "", stderr = "hive: task not found: 2-b\n"; want the --agent refusal
    --- FAIL: TestCallerAgentOverrideInPaneFailsForCallerFlags/task_verify (0.01s)
        caller_root_test.go:52: status = 1, stdout = "", stderr = "hive: task not found: 2-b\n"; want the --agent refusal
    --- FAIL: TestCallerAgentOverrideInPaneFailsForCallerFlags/task_reopen (0.01s)
        caller_root_test.go:52: status = 1, stdout = "", stderr = "hive: task not found: 2-b\n"; want the --agent refusal
    --- FAIL: TestCallerAgentOverrideInPaneFailsForCallerFlags/task_edit (0.01s)
        caller_root_test.go:52: status = 1, stdout = "", stderr = "hive: task not found: 2-b\n"; want the --agent refusal
    --- FAIL: TestCallerAgentOverrideInPaneFailsForCallerFlags/goal_create (0.01s)
        caller_root_test.go:52: status = 1, stdout = "", stderr = "hive: project not found: hive\n"; want the --agent refusal
    --- FAIL: TestCallerAgentOverrideInPaneFailsForCallerFlags/goal_edit (0.01s)
        caller_root_test.go:52: status = 1, stdout = "", stderr = "hive: --add-wave needs a title and a valid --done-when value\n"; want the --agent refusal
    --- FAIL: TestCallerAgentOverrideInPaneFailsForCallerFlags/chat_send (0.01s)
        caller_root_test.go:52: status = 1, stdout = "", stderr = "hive: chat send is for the user only: only the user can send chat; rule 1 (caller used --agent)\n"; want the --agent refusal
    --- FAIL: TestCallerAgentOverrideInPaneFailsForCallerFlags/chat_reply (0.01s)
        caller_root_test.go:52: status = 1, stdout = "", stderr = "hive: chat reply text is empty\n"; want the --agent refusal
--- FAIL: TestTaskCompleteAgentOverrideFromOtherPaneLeavesTask (0.04s)
    caller_root_test.go:74: status = 0, stdout = "id: 1-step\nstatus: completed\n", stderr = ""; want the --agent refusal
--- FAIL: TestInboxSendAgentOverrideInPaneFailsWithoutRecord (0.02s)
    inbox_root_test.go:265: status = 0, stdout = "updated counter .hive/state/counters/inbox:manager-hive.json\ncreated inbox .hive/state/inbox/manager-hive/1.json\n"; want a refusal
--- FAIL: TestInboxOverridesInPaneFailAndChangeNothing (0.09s)
    --- FAIL: TestInboxOverridesInPaneFailAndChangeNothing/read_agent (0.03s)
        inbox_root_test.go:297: status = 1, stdout = "", stderr = "hive: role has no inbox: worker\n"; want a refusal with "--agent agent-b is not the caller agent-a"
    --- FAIL: TestInboxOverridesInPaneFailAndChangeNothing/read_pane (0.03s)
        inbox_root_test.go:297: status = 1, stdout = "", stderr = "hive: role has no inbox: worker\n"; want a refusal with "--pane w2:p1 is not the caller pane w1:p1"
    --- FAIL: TestInboxOverridesInPaneFailAndChangeNothing/send_pane (0.04s)
        inbox_root_test.go:297: status = 0, stdout = "updated counter .hive/state/counters/inbox:agent-b.json\ncreated inbox .hive/state/inbox/agent-b/2.json\n", stderr = ""; want a refusal with "--pane w2:p1 is not the caller pane w1:p1"
--- FAIL: TestInboxAgentInUnregisteredPaneFails (0.03s)
    inbox_root_test.go:314: status = 0, stderr = ""; want the no-agent refusal
--- FAIL: TestInboxSendBetweenAgentInboxesInPanes (0.03s)
    inbox_root_test.go:333: read status = 1, stdout = "", stderr = "hive: role has no inbox: worker\n"; want the record
--- FAIL: TestInboxCoordinatorReadsAgentInboxWithoutConsume (0.04s)
    inbox_root_test.go:351: status = 1, stdout = "", stderr = "hive: inbox agent-b belongs to the owner\n"; want the record
--- FAIL: TestInboxAgentCannotReadOtherAgentInbox (0.03s)
    inbox_root_test.go:369: [inbox --role agent-b]: status = 1, stdout = "", stderr = "hive: inbox agent-b belongs to the owner\n"; want the ownership refusal
FAIL
FAIL	github.com/Harrison-Blair/hive/cmd	0.425s
FAIL
--- FAIL: TestReadDefaultsWorkerToNameInbox (0.02s)
    read_send_test.go:61: Read() = output.Envelope{Outcome:"failed", Phase:"validate", Effects:[]output.Effect{}, Error:"role has no inbox: worker", Result:interface {}(nil)}; want one consumed record of inbox worker-1
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/inbox	0.027s
FAIL
--- FAIL: TestRoutingWorkerOfCoordinatorGoesToCoordinator (0.04s)
    routing_test.go:75: blocked records = map[manager-demo:1]; want one in coordinator
--- FAIL: TestRoutingParentWithoutLiveSuccessorGoesToCoordinator (0.11s)
    --- FAIL: TestRoutingParentWithoutLiveSuccessorGoesToCoordinator/stopped_parent (0.04s)
        routing_test.go:118: blocked records = map[manager-other:1]; want one in coordinator
    --- FAIL: TestRoutingParentWithoutLiveSuccessorGoesToCoordinator/missing_parent (0.04s)
        routing_test.go:118: blocked records = map[manager-other:1]; want one in coordinator
    --- FAIL: TestRoutingParentWithoutLiveSuccessorGoesToCoordinator/no_parent (0.03s)
        routing_test.go:118: blocked records = map[manager-other:1]; want one in coordinator
--- FAIL: TestNudgeIdleWorkerGetsPromptForOwnInbox (0.02s)
    routing_test.go:139: prompts = 0; want one nudge to the worker
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/supervise	0.348s
FAIL
# github.com/Harrison-Blair/hive/internal/doctor [github.com/Harrison-Blair/hive/internal/doctor.test]
internal/doctor/orphan_inbox_test.go:19:11: undefined: checkOwnerlessInboxes
internal/doctor/orphan_inbox_test.go:36:10: undefined: checkOwnerlessInboxes
FAIL	github.com/Harrison-Blair/hive/internal/doctor [build failed]
FAIL
```

Each failure is for the correct reason:

- `cmd`: the base accepts the override. For example, `TestInboxSendAgentOverrideInPaneFailsWithoutRecord` gets status 0 and a created record. `TestTaskCompleteAgentOverrideFromOtherPaneLeavesTask` gets `status: completed`: a verifier pane completed the task of the worker.
  In `TestCallerAgentOverrideInPaneFailsForCallerFlags`, the base fails some cases on another validation, for example "task not found". Each case lacks the `--agent` refusal. The case in `TestTaskCompleteAgentOverrideFromOtherPaneLeavesTask` is fully valid, so the base completes the task.
- `internal/inbox`: the base refuses the worker read with "role has no inbox: worker".
- `internal/supervise`: the base sends the record to `manager-demo` or `manager-other` by project, and it gives no worker an inbox to nudge.
- `internal/doctor`: the check does not exist (build failure on `checkOwnerlessInboxes`).

These tests pass on the base by design, because the new rule gives the same result there: `TestRoutingUsualParentsMatchCurrentRule` (decision 29), `TestRoutingWorkerReachesRestartedManager` (the old rule also used `manager-PROJECT`), `TestTaskCompleteOwnAgentInPaneSucceeds`, and `TestReadRefusalsLogOneErrorEvent`. Break 2 below proves that the restart test needs the replaces chain.

## Green on the task branch

The full `go test -race ./...` run in `V-gate.md` passes each test.

## Deliberate breaks

Each break was a temporary edit. `git checkout` restored the file after each run.

```text
### Break 1: Caller ignores an --agent mismatch
40:	if false {
--- FAIL: TestCallerAgentOverrideInPaneFailsForCallerFlags (0.07s)
    --- FAIL: TestCallerAgentOverrideInPaneFailsForCallerFlags/task_complete (0.01s)
    --- FAIL: TestCallerAgentOverrideInPaneFailsForCallerFlags/task_verify (0.01s)
    --- FAIL: TestCallerAgentOverrideInPaneFailsForCallerFlags/task_reopen (0.01s)
    --- FAIL: TestCallerAgentOverrideInPaneFailsForCallerFlags/task_edit (0.01s)
    --- FAIL: TestCallerAgentOverrideInPaneFailsForCallerFlags/goal_create (0.01s)
    --- FAIL: TestCallerAgentOverrideInPaneFailsForCallerFlags/goal_edit (0.01s)
    --- FAIL: TestCallerAgentOverrideInPaneFailsForCallerFlags/chat_send (0.01s)
    --- FAIL: TestCallerAgentOverrideInPaneFailsForCallerFlags/chat_reply (0.01s)
--- FAIL: TestTaskCompleteAgentOverrideFromOtherPaneLeavesTask (0.04s)
--- FAIL: TestInboxSendAgentOverrideInPaneFailsWithoutRecord (0.01s)
--- FAIL: TestInboxOverridesInPaneFailAndChangeNothing (0.04s)
    --- FAIL: TestInboxOverridesInPaneFailAndChangeNothing/read_agent (0.01s)
FAIL
FAIL	github.com/Harrison-Blair/hive/cmd	0.169s
FAIL
### Break 2: ParentInbox does not follow replaces
31:		ok = false
--- FAIL: TestRoutingWorkerReachesRestartedManager (0.04s)
    routing_test.go:96: blocked records = map[coordinator:1]; want one in manager-demo
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/supervise	0.229s
FAIL
### Break 3: InboxKey of a worker is not its name
15:		return ""
--- FAIL: TestNudgeIdleWorkerGetsPromptForOwnInbox (0.02s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/supervise	0.024s
--- FAIL: TestReadDefaultsWorkerToNameInbox (0.02s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/inbox	0.020s
ok  	github.com/Harrison-Blair/hive/internal/doctor	0.087s
--- FAIL: TestInboxSendBetweenAgentInboxesInPanes (0.01s)
FAIL
FAIL	github.com/Harrison-Blair/hive/cmd	0.018s
FAIL
### Restored
### Break 4: doctor counts each agent record as a live owner
388:		if true {
--- FAIL: TestDoctorListsUnreadRecordsWithoutLiveOwner (0.04s)
    orphan_inbox_test.go:38: check = {Name:ownerless inbox records Status:warn Required:false Detail:1: coordinator/1}; want warn with the two unread records
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/doctor	0.047s
FAIL
### Restored
```
