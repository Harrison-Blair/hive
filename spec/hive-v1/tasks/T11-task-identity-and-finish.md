# T11 task identity and finish

| Item | Value |
| --- | --- |
| Task ID | T11-task-identity-and-finish |
| Wave | 3 |
| Tier | small |
| Depends on | T03, T04 |
| Blocks | T14, T15 |
| Branch | `task/T11-task-identity-and-finish` |
| Worktree | `projects/hive/worktrees/T11-task-identity-and-finish/` |
| Evidence | `evidence/T11-task-identity-and-finish/` |

## Overview

This task joins the task records of T03 with the agent records of T04. A
caller no longer needs `--agent` inside herdr: hive finds its record from
`HERDR_PANE_ID`. `task complete` and `task verify` mark the caller finished and
write an inbox record to the controlling role. The supervisor of T15 acts on
the mark and the record.

## User flows

1. A worker runs `hive task complete --task 12-spawn-worker` in its pane. Hive finds the record `worker-hive-3` by the pane, moves the task, marks the worker finished, and writes `completed` to the inbox `manager-hive`.
2. A verifier runs `hive task verify --task 12-spawn-worker`. Hive writes `verified` to the inbox and marks the verifier finished.
3. The manager runs `hive inbox` without `--role` in its pane. Hive reads the role from the record of the pane.
4. The user runs `hive task complete --task 12-spawn-worker` from a shell without an agent record. Hive fails with `caller unknown: pass --agent`.
5. A worker of another task runs `hive task verify --task 12-spawn-worker`. Hive fails with `not a verifier of 12-spawn-worker: worker-hive-4` and changes nothing.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/lib/identity/identity.go` | New. Resolve the caller from `--agent`, else from the pane ID and the agent records. |
| `internal/task/complete.go` | Use the identity. Mark the agent finished. Write the inbox record. |
| `internal/task/verify.go` | Use the identity. Check the verifier record. Mark the verifier finished. Write the inbox record. |
| `internal/task/show.go`, `render.go` | Show the acting agent and its status. |
| `internal/inbox/read.go`, `send.go` | Default the role and the sender from the identity. |
| `cmd/task/task.go`, `cmd/inbox/inbox.go` | Pass the pane ID from the environment. |
| `internal/agent/register.go` | Check that `--task` names a task record of the project. T04 defers this check to this task. |
| `internal/lib/records/agent.go` | Add `pane_history`, the earlier pane IDs of a moved pane. T15 fills it. |

## Requirements

### AC1 Identity

- `identity.Resolve(agentFlag, paneID, store)` returns one agent record. With `agentFlag` it returns that record or fails with `agent not found: <name>`.
- Without the flag, it keeps each record with the status `live` whose `pane_id` equals the pane ID, or whose `pane_history` holds the pane ID. One match is the caller. No match fails with `caller unknown: pass --agent`. Two or more matches fail with `caller ambiguous: <names>`.
- A record of a singleton role is a match only when its `generation` equals the generation of the claim of its role.
- The command reads `HERDR_PANE_ID` and passes it. The library does not read the environment.
- `hive agent register` with `--task` for a worker or a verifier needs a task record with the same project as `--project`. Otherwise it fails with `task not found in <project>: <id>` and writes no record.
- A process in a moved pane keeps the old value in `HERDR_PANE_ID`. The supervisor of T15 writes the old pane ID into `pane_history` when it sees the move. The lookup through `pane_history` keeps the identity of that process.

### AC2 Complete and verify

- `task complete` resolves the caller. It must be the acting agent. On success it sets `finished` true on the agent record. It appends an inbox record of the kind `completed` to the controlling role, with the task and the agent.
- `task verify` resolves the caller. The caller must be a record with the role `verifier`, the project of the task, and the task. Another caller fails with `not a verifier of <task>: <name>`. The rule of T03 that refuses the acting agent stays.
- On success `task verify` sets `finished` true on the verifier record and appends `verified` with the result. On `--fail` the text starts with `fail: <reason>`.
- The task transition, the agent record, and the inbox record change in one store mutation. A refused transition changes none of the three records.

### AC3 Inbox defaults

- `hive inbox` without `--role` resolves the caller and uses the key of its role: `coordinator` for a coordinator, `manager-<project>` for a manager. A worker, a verifier, or a supervisor without `--role` fails with `role has no inbox: <role>`.
- `hive inbox send` without `--agent` fills `agent` from the caller.

### AC4 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The identity resolves from the flag, then from one live record by pane, and fails as the requirements give. A registration with `--task` needs a task of the project. | V1 |
| AC2 | `complete` and `verify` check the caller, mark it finished, and write the inbox record in one mutation. | V2, V3 |
| AC3 | `inbox` and `inbox send` default the role and the sender from the caller. | V3 |
| AC4 | The gate passes. | V4 |

## Verification items

### V1 Identity tests

- Proves: AC1.
- Command: `go test ./internal/lib/identity/ ./internal/agent/ -run 'TestResolve|TestRegisterTask' -v`, red and green.
- Expected: the flag case, the pane case, the `pane_history` case, and the unknown case. A stale record with the same pane does not match. Two live records with the same pane fail with `caller ambiguous`. A singleton record with an old generation does not match. A registration with a task of another project, or with an unknown task, fails and writes no record. A registration with a task of the project succeeds.

### V2 Complete and verify tests

- Proves: AC2.
- Command: `go test ./internal/task/ -run 'TestCompleteFinish|TestVerifyFinish' -v`, red and green.
- Expected: after `complete`, the agent record has `finished` true and the inbox of `manager-<project>` has one `completed` record. After `verify`, the verifier record has `finished` true and the inbox has a `verified` record. After `verify --fail`, the text starts with `fail:`. A worker of another task, a manager, and a verifier of another task each fail. The task, the agent records, and the inbox do not change. A refused transition, such as `complete` on a `created` task, changes none of the three records.

### V3 Command transcript

- Proves: AC2, AC3.
- Command:

```sh
go build .
rm -rf .hive/state .hive/tmp/repo-one
git init -q .hive/tmp/repo-one
./hive --home "$PWD" project register .hive/tmp/repo-one
./hive --home "$PWD" task create --project repo-one --title "step"
./hive --home "$PWD" agent register --role manager --project repo-one --pane w9:p1
./hive --home "$PWD" agent register --role worker --project repo-one --task 1-step --pane w9:p2
./hive --home "$PWD" agent register --role verifier --project repo-one --task 1-step --pane w9:p3
./hive --home "$PWD" task assign --task 1-step --agent worker-repo-one-1
HERDR_PANE_ID=w9:p2 ./hive --home "$PWD" inbox send --to manager-repo-one "hello"
HERDR_PANE_ID=w9:p2 ./hive --home "$PWD" inbox
echo "exit $?"
HERDR_PANE_ID=w9:p2 ./hive --home "$PWD" task complete --task 1-step
HERDR_PANE_ID=w9:p3 ./hive --home "$PWD" task verify --task 1-step
HERDR_PANE_ID=w9:p1 ./hive --home "$PWD" inbox
./hive --home "$PWD" agent show worker-repo-one-1
./hive --home "$PWD" agent show verifier-repo-one-2
rm -rf .hive/tmp/repo-one
```

- Expected: the `message` record has the agent `worker-repo-one-1`. The inbox call of the worker fails with `role has no inbox: worker` and exit status 1. The complete and the verify succeed without `--agent`. The inbox of the manager prints the `message`, the `completed`, and the `verified` records. Both agent records show `finished: true`.

### V4 Gate

- Proves: AC4.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- The stop of a finished agent. T15 does it.
- The update of `pane_id` and `pane_history` on a move. T15 does it.

## Changes

None.
