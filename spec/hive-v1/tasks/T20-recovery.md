# T20 recovery

| Item | Value |
| --- | --- |
| Task ID | T20-recovery |
| Wave | 5 |
| Tier | small |
| Depends on | T13, T15, CP3 |
| Blocks | CP5 |
| Branch | `task/T20-recovery` |
| Worktree | `projects/hive/worktrees/T20-recovery/` |
| Evidence | `evidence/T20-recovery/` |

## Overview

The supervisor marks a record `unresolved` when its live state is uncertain.
This task gives the way to settle such a record and the guards that hold its
work until then. Recovery is conservative: hive holds and reports, and a
person decides. A herdr name does not identify a process, because a new
process of the same kind can keep the name. The resolve therefore compares the
identity of the live agent with the record before it trusts the name.

## User flows

1. After a herdr restart the coordinator runs `hive agent list`. Two records show `unresolved` at the top.
2. The coordinator runs `hive agent resolve --agent worker-hive-3 --as exited`. The record becomes `exited`.
3. The coordinator runs `hive agent resolve --agent manager-hive --as live`. Hive finds the live agent, compares its kind, its directory, and its terminal with the record, and updates the pane IDs from it.
4. A manager runs `hive worktree remove --task 12-spawn-worker` while its worker is `unresolved`. Hive refuses with `agent unresolved: worker-hive-3`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `cmd/agent/resolve/resolve.go` | New. |
| `cmd/agent/agent.go` | Register the subcommand. |
| `internal/agent/resolve.go` | New. |
| `internal/agent/list.go` | Sort `unresolved` records first. |
| `internal/agent/stop.go` | Refuse an `unresolved` record. |
| `internal/worktree/remove.go` | Refuse when an agent record or a live herdr agent occupies the worktree. |
| `internal/task/show.go`, `render.go` | Show the status of the acting agent. |
| `internal/doctor/checks.go` | Add one check: the count of `unresolved` records, `warn` above zero. |

## Requirements

### AC1 Resolve

- `hive agent resolve --agent NAME --as exited|live` needs a record with the status `unresolved`. Another status fails with `agent is <status>`.
- `--as exited` sets the status `exited`.
- `--as live` calls `agent.get` with the recorded herdr name as the target. When the record has no herdr name, the target is the recorded pane ID.
- Hive compares the fields of the live agent with the record. The agent kind must equal the recorded harness. The canonical `cwd` of the pane must equal the recorded `cwd`. The `terminal_id` must equal the recorded `terminal_id`.
- When the record holds `process_pid`, compare it exactly with the live process. A difference fails with `agent identity differs: process_pid`. When the record holds `process_start`, compare it exactly. A difference fails with `agent identity differs: process_start`. Skip a field that the record does not hold. A difference leaves the record unchanged.
- A match sets the status `live` and updates `pane_id`, `workspace_id`, `tab_id`, and `terminal_id` from the live agent.
- The herdr error `agent_not_found` fails with `no live agent named <name>`. Each other failure, such as a timeout or a transport error, fails with the herdr code and message. In each failure the record stays unchanged.
- The resolve writes the agent record. The supervisor of T15 watches the agent directory and rebuilds its subscription for the new pane ID.

### AC2 Guards

- `agent stop` on an `unresolved` record fails with `agent unresolved: run hive agent resolve`.
- `worktree remove` fails with `agent <status>: <name>` when a record of the task has a status other than `stopped` or `exited`. The same rule applies to a record with the `workspace_id` or the path of the worktree. A live herdr agent in the workspace, from `agent.list`, fails with `agent in workspace: <pane>`. `--force` does not override these guards.
- `agent list` prints `unresolved` records first. `task show` prints the status of the acting agent: the record status, or `no record` when the record is missing.
- `hive doctor` prints `warn unresolved agents: <n>` when the count is above zero, else `ok`.

### AC3 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | `resolve` settles an unresolved record in the two ways, checks the identity, and refuses other records and other failures. | V1, V3 |
| AC2 | The stop, the remove, the list, the show, and the doctor act on unresolved records as the requirements give. | V2, V3 |
| AC3 | The gate passes. | V4 |

## Verification items

### V1 Resolve tests

- Proves: AC1.
- Command: `go test ./internal/agent/ -run TestResolve -v`, red and green.
- Expected: a fake herdr server. Cases: `exited`. `live` with a matching agent, with the four updated IDs. A matching name with a different kind, a different `cwd`, or a different `terminal_id` fails with the message and an unchanged record. A different `process_pid` or `process_start` fails the same way. A record without `process_pid` or `process_start` skips that field. `agent_not_found` with its message. A timeout with the herdr code and an unchanged record. A record that is not unresolved.

### V2 Guard tests

- Proves: AC2.
- Command: `go test ./internal/agent/ ./internal/worktree/ ./internal/task/ ./internal/doctor/ -run 'TestStopUnresolved|TestRemoveAgentGuard|TestListOrder|TestShowAgentStatus|TestDoctorUnresolved' -v`, red and green.
- Expected: the stop message. The remove guard for four occupants. The occupants: a record of the task, a record with the workspace, a record with the path, and a live herdr agent in the workspace. Each case also with `--force`. The list with a `live` record written before an `unresolved` record, and the `unresolved` record first. The show for `live`, `unresolved`, `exited`, `stopped`, and a missing record. The doctor line.

### V3 Command test

- Proves: AC1, AC2.
- Command: `go test ./cmd/agent/resolve/ -run TestResolveCommand -v`, red and green.
- Expected: the test makes a temporary home with the marker and a fake herdr server. It seeds one `unresolved` record through the store library. It runs the commands through `ExecuteWithArgs` with `--home`: `agent list`, `agent stop`, `doctor`, `agent resolve --as live` with a matching agent, and `agent resolve --as exited` on a second seeded record. It asserts the exact messages, the exit status of each command, and the updated IDs.

### V4 Gate

- Proves: AC3.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- Automatic resolution. A person decides.
- The `cwd` and `terminal_id` fields of a record. T04 and T10 write them at registration and at spawn.

## Changes

2026-10-18. Compare `process_pid` and `process_start` exactly when the record
holds them. Skip an absent field. Keep the manual resolve rule.
