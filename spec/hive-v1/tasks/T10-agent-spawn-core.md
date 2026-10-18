# T10 agent spawn core

| Item | Value |
| --- | --- |
| Task ID | T10-agent-spawn-core |
| Wave | 3 |
| Tier | strong |
| Depends on | T04, T05 |
| Blocks | CP2, T13, T14, T15, T16, T17, T18, T19 |
| Branch | `task/T10-agent-spawn-core` |
| Worktree | `projects/hive/worktrees/T10-agent-spawn-core/` |
| Evidence | `evidence/T10-agent-spawn-core/` |

## Overview

This task gives the core of agent control: the harness profile interface with
the Claude Code profile, the herdr agent methods, the `pane.get` wrapper, and
the three commands `agent spawn --pane`, `agent send`, and `agent stop`. The
spawn takes an explicit pane. T13 and T14 choose the pane for a manager and a
worker. The profile tasks T16 to T19 add one profile each. The registry takes
one line for each profile.

herdr 0.9.5 has no stop method. Hive stops an agent with the exit keys of its
harness profile through `agent.send_keys`, then waits until the agent leaves
the pane. The pane keeps its shell. Hive targets each agent request at the pane
ID of the record, never at the name, because herdr can drop a name while the
process continues.

## User flows

1. The user runs `hive agent spawn --pane w3:p1 --role worker --project hive --task 12-spawn-worker --harness claude --model <small model>`. Claude Code starts in the pane. Hive writes the record, waits until the agent is idle, and prints the name `worker-hive-1`.
2. The user runs `hive agent send --agent worker-hive-1 "Reply with the word pong."`. The text arrives in the pane and the agent answers.
3. Claude Code shows its folder trust dialog. The spawn ends with the outcome `partial` and the error `agent blocked: accept the dialog in w3:p1`. The record exists with the status `live`.
4. The user runs `hive agent stop --agent worker-hive-1`. Hive types `/exit` and Enter into the pane. The agent process ends, the pane shows a shell, and the record has the status `stopped`.
5. The agent ignores the exit keys. After 30 seconds hive ends with `partial` and the error `agent did not exit within 30 s`. The user runs `hive agent stop --agent worker-hive-1 --force`. Hive sends Ctrl-C twice, and the agent ends.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `cmd/agent/spawn/spawn.go`, `send/send.go`, `stop/stop.go` | New. |
| `cmd/agent/agent.go` | Register the three subcommands. |
| `internal/agent/spawn.go` | New. The spawn sequence. T13 and T14 call it with a pane. |
| `internal/agent/send.go` | New. |
| `internal/agent/stop.go` | New. The stop function. T13 and T15 call it. |
| `internal/agent/inspect.go` | New. The pane inspection after `agent_not_found`: `exited` or `unresolved`. |
| `internal/agent/doc.go` | Add the map lines. |
| `internal/lib/harness/doc.go` | New. Package summary and file map. |
| `internal/lib/harness/harness.go` | New. The `Profile` interface, the lookup by kind, and the bounds check. |
| `internal/lib/harness/registry.go` | New. One line for each profile. |
| `internal/lib/harness/claude.go` | New. The Claude Code profile. |
| `internal/lib/herdr/agent.go` | New. `AgentStart`, `AgentWait`, `AgentPrompt`, `AgentSendKeys`, `AgentGet`, `AgentList`. |
| `internal/lib/herdr/pane.go` | New. `PaneGet`. T13 adds `PaneList`. |
| `internal/lib/herdr/doc.go` | Add the map lines. |
| `internal/lib/records/agent.go` | Add `terminal_id`, `cwd`, and `last_error`. |

## Requirements

### AC1 Harness profile

- A profile gives `Kind()`, the herdr agent kind such as `claude`. It gives `Args(model string, extra []string)`, the native arguments for `agent.start`. It gives `StartTimeout()`, the value for `timeout_ms`, and `ReadyTimeout()`, the bound of the wait for a settled status.
- A profile gives `ExitKeys()`, the logical keys that end the agent from an idle prompt. It gives `ForceExitKeys()`, the keys for `--force`, and `StopTimeout()`, the bound of the wait for the exit.
- `harness.Check(profile)` fails when `StartTimeout()` is not above 3 seconds and at most 300 seconds, when `StartTimeout()` is below `ReadyTimeout()`, or when `ExitKeys()` is empty. A start timeout below the ready timeout lets herdr drop the name while hive still waits.
- The Claude Code profile has the kind `claude`. With a model it adds `--model <model>`. It appends the extra arguments. Its start timeout and its ready timeout are 120 seconds. Its exit keys are `/`, `e`, `x`, `i`, `t`, and `enter`. Its force exit keys are `ctrl+c` and `ctrl+c`. Its stop timeout is 30 seconds.
- `harness.Lookup(kind)` returns the profile or fails with `unknown harness: <kind>` and the list of known kinds.

### AC2 Spawn

- `hive agent spawn --role ROLE --harness KIND [--pane ID] [--model M] [--project ID] [--task ID] [--name N] [--quiet-after MIN] [--arg X]...`. `--arg` repeats. The command line does not require `--pane` or `--project`. The spawn function checks them by role.
- The pane choice is one `switch` on the role in `spawn.go`, for a spawn without `--pane`. Each case is one line that calls a function in its own file. This task has only the default case, which fails with `pane required: pass --pane`. T13 and T14 add one case each.
- The role rules of T04 apply after the pane choice. The name comes from the same rules, or from `--name`. The name is the record key and the `name` of `agent.start`.
- Hive writes the record before the start: the register of T04 with the pane, the harness, and the model. The record is the durable intent. A crash after this point leaves a `live` record that the supervisor of T15 reconciles.
- Then `agent.start` with `name`, `kind`, `pane_id`, `args`, and `timeout_ms` from the start timeout. An error envelope means that no agent started. Hive sets the record `exited` with `last_error` from the code and the message, and the outcome is `failed`.
- A client error during the start, such as a deadline or a closed connection, leaves the result unknown. Hive sets the record `unresolved` and the outcome `partial` with the phase `start`.
- After the start reply, hive calls `pane.get` on the pane. It records `terminal_id`, `workspace_id`, `tab_id`, and the canonical `cwd` at once.
- Then `agent.wait` with `target` equal to the pane ID, `until` equal to `idle`, `blocked`, and `done`, and `timeout_ms` from the ready timeout. The client deadline of the call is the ready timeout plus 5 seconds, set with `herdr.WithDeadline` of T05.
- The result `agent_info` with the status `idle` or `done` gives the outcome `ok`. The status `blocked` gives `partial` with the error `agent blocked: accept the dialog in <pane>`. The error `timeout` gives `partial` with `agent not ready after <seconds> s`. In each `partial` case the record stays `live`.
- The error `agent_not_running` or `agent_not_found` during the wait means that the agent left the pane. Hive runs the inspection of AC4 and gives `failed` with `agent exited during start` or the `unresolved` error.
- After the wait hive calls `agent.get` with the pane ID and logs the `name` of the reply. A missing name is a log event, not an error, because later requests use the pane ID.
- The effects: record created, agent started, claim taken when the role is a singleton.
- Hive logs the `argv` of the start reply through `log.Body`.

### AC3 Send

- `hive agent send --agent NAME TEXT [--wait] [--timeout SECONDS]` sends `agent.prompt` with `target` equal to the pane ID of the record and `text`. Without `--wait` the request has no `wait` field. With `--wait` the request has `wait` with `until` equal to `idle`, `blocked`, and `done`, and `timeout_ms` from `--timeout`, default 600 seconds.
- The exact shape with `--wait`: `{"target":"<pane>","text":"<text>","wait":{"until":["idle","blocked","done"],"timeout_ms":600000}}`.
- The client deadline of the call is the wait timeout plus 5 seconds, set with `herdr.WithDeadline`. Without `--wait` the default deadline of the client applies.
- A record with a status other than `live` fails with `agent is <status>`.
- `agent_blocked` gives `failed`. The text did not reach the agent. `agent_prompt_stalled` gives `failed` with the remark that the text can be in the input of the agent. `agent_not_ready` gives `failed`. The error `timeout` gives `partial`: the text reached the agent and no settled status came within the bound.
- Hive never sends the text a second time by itself. `--wait` observes a settled status of the agent. It does not prove that the agent completed that prompt.
- `agent_not_found` runs the inspection of AC4 and gives `failed`.
- The text goes to the log through `log.Body`.

### AC4 Stop

- `hive agent stop --agent NAME [--force]` needs a record with the status `live`. Another status fails with `agent is <status>`.
- Step 1: `agent.get` with the pane ID of the record. On `agent_not_found` hive runs the inspection. The inspection calls `pane.get`. A pane with `agent_status` `unknown` and no `agent` kind shows a shell, so the record becomes `exited` and the command prints `agent was not live`. The outcome is `ok`. A pane that does not exist makes the record `unresolved`, and the outcome is `failed` with `pane <id> not found: record unresolved`.
- Step 2: `agent.send_keys` with `target` equal to the pane ID and `keys` from `ExitKeys()`, or from `ForceExitKeys()` with `--force`. The error `agent_not_ready` gives `failed` with the herdr message, and the record stays `live`. A pending start and a suspended agent give this error.
- Step 3: `agent.wait` with `target` equal to the pane ID, `until` equal to `unknown`, and `timeout_ms` from `StopTimeout()`. The client deadline is the stop timeout plus 5 seconds through `herdr.WithDeadline`. The errors `agent_not_running` and `agent_not_found`, and a result with the status `unknown`, end the wait. Then `pane.get` must show `agent_status` `unknown`. The record becomes `stopped`, and the outcome is `ok`.
- The error `timeout` gives `partial` with the phase `wait`, the effect `keys sent`, and the error `agent did not exit within <n> s`. The record stays `live`. The user runs the command again, with `--force` when needed.
- `internal/agent/stop.go` exports the stop function with the record and the force flag as arguments, so T13 and T15 call it without the command.

### AC5 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The profile interface and the Claude Code profile give the kind, the arguments, the timeouts, and the exit keys, and the bounds check holds. | V1 |
| AC2 | The spawn writes the record first and sends start and wait with the pane as the target. It maps each result to an outcome and a record status. | V2, V4 |
| AC3 | `send` prompts the agent with the exact request shape and maps each herdr error. | V3, V4 |
| AC4 | `stop` sends the exit keys, waits for the exit, and updates the record, also when the agent is already gone. | V3, V4 |
| AC5 | The gate passes. | V5 |

## Verification items

### V1 Profile tests

- Proves: AC1.
- Command: `go test ./internal/lib/harness/ -v`, red and green.
- Expected: the Claude Code arguments with and without a model, the extra arguments, the timeouts, the exit keys, and the unknown kind error. The bounds check fails a profile with a start timeout of 3 seconds, of 301 seconds, or below its ready timeout. It passes each registered profile.

### V2 Spawn tests

- Proves: AC2.
- Command: `go test -race ./internal/agent/ -run TestSpawn -v`, red and green.
- Expected: a fake herdr server scripts the replies with the shapes of `reference/herdr/api/agent.md`: `agent_info` for the wait, and the error codes `timeout`, `agent_not_running`, and `agent_not_found`. Cases: idle gives `ok`, done gives `ok`, blocked gives `partial`, a wait timeout gives `partial`, and a start error gives `failed` with the record `exited`. A closed connection during the start gives `partial` with the record `unresolved`. `agent_not_running` during the wait gives `failed` with the record `exited`. A reply without a name after the wait gives `ok`. The fake server checks that the record file exists when the start request arrives. The test checks the start parameters, the wait target, and the record fields with `harness`, `cwd`, and `terminal_id`.

### V3 Send and stop tests

- Proves: AC3, AC4.
- Command: `go test ./internal/agent/ -run 'TestSend|TestStop' -v`, red and green.
- Expected: the prompt request without `wait`, the exact nested `wait` shape, and the outcomes for `agent_blocked`, `agent_prompt_stalled`, `agent_not_ready`, and `timeout`. The stop cases: the exit keys and the `stopped` record after `agent_not_running`, the force keys with `--force`, and `partial` after `timeout` with the record still `live`. More stop cases: `agent_not_ready` from the keys, `exited` after `agent_not_found` with a shell in the pane, and `unresolved` after `agent_not_found` with a missing pane.

### V4 Live transcript

- Proves: AC2, AC3, AC4.
- Command, from a pane of the hive session, with a new tab in the own workspace of the worker:

```sh
go build .
NAME="t10-check-$(date +%s)"
herdr tab create --workspace "$HERDR_WORKSPACE_ID"
PANE=<root pane ID from the output>
./hive --home "$PWD" agent spawn --pane "$PANE" --role coordinator --name "$NAME" --harness claude --model <small model>
./hive --home "$PWD" agent send --agent "$NAME" --wait --timeout 120 "Reply with the word pong."
herdr pane read "$PANE" --lines 20
./hive --home "$PWD" agent list
./hive --home "$PWD" agent stop --agent "$NAME"
echo "exit $?"
herdr pane read "$PANE" --lines 5
herdr agent get "$PANE"
./hive --home "$PWD" agent list --all
herdr tab close <the new tab>
```

- Expected: the spawn prints the name and the outcome `ok`. The pane read shows the word pong. The list shows the agent `live`. The stop exits with 0. The second pane read shows a shell prompt, and `agent get` answers `agent_not_found`. The last list shows the record `stopped`. The tab is gone at the end.

### V5 Gate

- Proves: AC5.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- The pane choice for a manager and the workspace of a project. T13 adds them.
- The pane choice for a worker, the brief prompt, and the task assignment. T14 adds them.
- The close of a workspace on stop. T13 adds it.
- `hive agent resolve` for an `unresolved` record. T20 adds it.

## Changes

None.
