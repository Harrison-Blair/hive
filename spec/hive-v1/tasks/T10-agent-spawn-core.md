# T10 agent spawn core

| Item | Value |
| --- | --- |
| Task ID | T10-agent-spawn-core |
| Wave | 3 |
| Tier | strong |
| Depends on | T04, T05, T09 |
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

herdr 0.9.5 has no stop method. Hive stops an agent with `pane.close` on the
pane of the record. herdr ends the process and removes the pane. When the pane
was the last pane of its workspace, herdr closes the workspace too. Hive closes
only a pane in which hive started the agent. Hive targets each agent request at
the pane ID of the record, never at the name, because herdr can drop a name
while the process continues.

## User flows

1. The user runs `hive agent spawn --pane w3:p1 --role worker --project hive --task 12-spawn-worker --harness claude --model <small model>`. Claude Code starts in the pane. Hive writes the record, waits until the agent is idle, and prints the name `worker-hive-1`.
2. The user runs `hive agent send --agent worker-hive-1 "Reply with the word pong."`. The text arrives in the pane and the agent answers.
3. Claude Code shows its folder trust dialog. The spawn ends with the outcome `partial` and the error `agent blocked: accept the dialog in w3:p1`. The record exists with the status `live`.
4. The user runs `hive agent stop --agent worker-hive-1`. Hive closes the pane. The agent process ends, the pane is gone, and the record has the status `stopped`. When the pane was the last one of its workspace, the workspace is gone too.
5. The user registered the coordinator in their own pane with `hive agent register`. `hive agent stop --agent coordinator` fails with `agent coordinator was not spawned by hive: exit it by hand`. The pane stays.

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
| `internal/lib/harness/harness.go` | New. The `Profile` interface, `ProcessNames()`, the lookup by kind, and the bounds check. |
| `internal/lib/harness/registry.go` | New. One line for each profile. |
| `internal/lib/harness/claude.go` | New. The Claude Code profile. |
| `internal/lib/herdr/agent.go` | New. `AgentStart`, `AgentWait`, `AgentPrompt`, `AgentGet`, `AgentList`, and the typed `AgentSessionInfo`. |
| `internal/lib/herdr/pane.go` | New. `PaneGet`, `PaneClose`, and the typed `PaneProcessInfo`. T13 adds `PaneList`. |
| `internal/lib/herdr/types.go` | Shared with T05 and T09. Add only the optional `agent_session` field to `AgentInfo` and `PaneInfo`. Change no other field. |
| `internal/lib/herdr/doc.go` | Add the map lines. |
| `internal/lib/process/process.go` | New. `Start(pid)` and `Alive(pid, start)` with a bounded `/proc/<pid>/stat` read. |
| `internal/lib/process/process_test.go` | New. |
| `internal/lib/records/agent.go` | Add `last_error`, `spawned`, `process_pid`, `process_name`, `process_start`, and `agent_session_ref`. Reuse the existing T04 fields `terminal_id` and `cwd`. Do not duplicate them. |

## Requirements

### AC1 Harness profile

- A profile gives `Kind()`, the herdr agent kind such as `claude`. It gives `Args(model string, extra []string)`, the native arguments for `agent.start`. It gives `StartTimeout()`, the value for `timeout_ms`, and `ReadyTimeout()`, the bound of the wait for a settled status.
- A profile gives `ProcessNames()`, the executable names that can identify its harness. The Claude Code profile returns `claude`. A generic interpreter or shell name does not identify a harness.
- `harness.Check(profile)` fails when `StartTimeout()` is not above 3 seconds and at most 300 seconds, or when `StartTimeout()` is below `ReadyTimeout()`. A start timeout below the ready timeout lets herdr drop the name while hive still waits.
- The Claude Code profile has the kind `claude`. With a model it adds `--model <model>`. It appends the extra arguments. Its start timeout and its ready timeout are 120 seconds.
- A profile process-name test accepts each actual executable name. It rejects `node`, `python`, `sh`, and `bash` when they are not the harness name.
- `harness.Lookup(kind)` returns the profile or fails with `unknown harness: <kind>` and the list of known kinds.

### AC2 Spawn

- `hive agent spawn --role ROLE --harness KIND [--pane ID] [--model M] [--project ID] [--task ID] [--name N] [--quiet-after MIN] [--arg X]...`. `--arg` repeats. The command line does not require `--pane` or `--project`. The spawn function checks them by role.
- The pane choice is one `switch` on the role in `spawn.go`, for a spawn without `--pane`. Each case is one line that calls a function in its own file. This task has only the default case, which fails with `pane required: pass --pane`. T13 and T14 add one case each.
- The role rules of T04 apply after the pane choice. The name comes from the same rules, or from `--name`. The name is the record key and the `name` of `agent.start`.
- Hive writes the record before the start: the register of T04 with the pane, the harness, and the model, and `spawned` true. A record from `hive agent register` has `spawned` false. The record is the durable intent. A crash after this point leaves a `live` record that the supervisor of T15 reconciles.
- Then `agent.start` with `name`, `kind`, `pane_id`, `args`, and `timeout_ms` from the start timeout. An error envelope means that no agent started. Hive sets the record `exited` with `last_error` from the code and the message, and the outcome is `failed`.
- A client error during the start, such as a deadline or a closed connection, leaves the result unknown. Hive sets the record `unresolved` and the outcome `partial` with the phase `start`.
- After the start reply, hive calls `pane.get` on the pane. It records `terminal_id`, `workspace_id`, `tab_id`, and the canonical `cwd` at once.
- Then `agent.wait` with `target` equal to the pane ID, `until` equal to `idle`, `blocked`, and `done`, and `timeout_ms` from the ready timeout. The client deadline of the call is the ready timeout plus 5 seconds, set with `herdr.WithDeadline` of T05.
- The result `agent_info` with the status `idle` or `done` gives the outcome `ok`. The status `blocked` gives `partial` with the error `agent blocked: accept the dialog in <pane>`. The error `timeout` gives `partial` with `agent not ready after <seconds> s`. In each `partial` case the record stays `live`.
- After the wait result, hive calls `pane.process_info`. It records the process identity only when exactly one foreground entry names the harness and its PID belongs to the foreground process set. A multiple, absent, shell, unknown, wrapper-only, or unstable entry gives `unresolved`. Keep the resource and report the started-agent effect and the uncertain phase.
- The optional `agent_session` fields of `AgentInfo` and `PaneInfo` use the typed `AgentSessionInfo`. It has `source`, `agent`, `kind`, and `value`. Record it when present. An absent session alone passes. A recorded session that is missing or differs gives `unresolved`.
- A wait timeout with independently proven ownership can keep the live partial-readiness result. The evidence uses one bounded `agent.get` reply and the stable typed process identity.
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

- `hive agent stop --agent NAME` needs a record with the status `live`. Another status fails with `agent is <status>`.
- A record with `spawned` false fails with `agent <name> was not spawned by hive: exit it by hand`. Hive closes only a pane in which it started the agent.
- Step 1: `pane.get` with the pane ID of the record. The error `pane_not_found` makes the record `unresolved`, and the outcome is `failed` with `pane <id> not found: record unresolved`. A reply whose `terminal_id` differs from the record makes the record `unresolved` too, with `pane <id> holds another terminal: record unresolved`.
- Step 2: `pane.process_info` and the recorded process identity. The recorded `process_pid` must be in the foreground process set. The recorded `process_name` must match. The recorded `process_start` must match. An absent recorded session passes. A recorded session reference that is missing or different gives `unresolved`. A mismatch, a missing process, a shell in the foreground, or an unreadable `/proc` gives `unresolved` and closes nothing.
- Step 3: `workspace.get` with the `workspace_id` of the reply. Hive keeps the workspace ID for the post-close read. The error `workspace_not_found` also makes the record `unresolved`.
- Step 4: hive sets the record `stopped` under the store lock, before the close. The close events of this pane then find no `live` record, so the supervisor of T15 logs them only.
- Step 5: `pane.close` with the pane ID. The result `ok` gives the effect `pane closed`. The error `pane_not_found` here makes the record `unresolved`. Another error sets the record back to `live` with `last_error`, and the outcome is `failed`.
- Step 6: a bounded `workspace.get` outside the store lock. The result `closed` means `workspace_not_found`. The result `surviving` means a workspace reply with its pane count. The result `unknown` means a timeout, a transport error, or another failure. Do not infer closure from the pre-close pane count. An `unknown` result keeps the project metadata.
- A manual registration has `spawned` false. Its stop refusal sends no close.
- The stop closes a working agent at once. The command has no graceful mode and no `--force`.
- `internal/agent/stop.go` exports the stop function with the record as its argument, so T13 and T15 call it without the command. It returns the post-close result and the known effects, so T13 can update the project record.

### AC5 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The profile interface and the Claude Code profile give the kind, the arguments, and the timeouts, and the bounds check holds. | V1 |
| AC2 | The spawn writes the record first and sends start and wait with the pane as the target. It maps each result to an outcome and a record status. | V2, V4 |
| AC3 | `send` prompts the agent with the exact request shape and maps each herdr error. | V3, V4 |
| AC4 | `stop` refuses a registered agent, checks the process identity, closes the pane of a spawned agent, returns the post-close topology, and updates the record. | V3, V4 |
| AC5 | The gate passes. | V5 |

## Verification items

### V1 Profile tests

- Proves: AC1.
- Command: `go test ./internal/lib/harness/ -v`, red and green.
- Expected: the Claude Code arguments with and without a model, the extra arguments, the timeouts, and the unknown kind error. The bounds check fails a profile with a start timeout of 3 seconds, of 301 seconds, or below its ready timeout. It passes each registered profile. The process-name test accepts `claude` and rejects `node`, `python`, `sh`, and `bash`.

### V2 Spawn tests

- Proves: AC2.
- Command: `go test -race ./internal/agent/ -run TestSpawn -v`, red and green.
- Expected: a fake herdr server scripts the replies with the shapes of `reference/herdr/api/agent.md`: `agent_info` for the wait, and the error codes `timeout`, `agent_not_running`, and `agent_not_found`. Cases: idle gives `ok`, done gives `ok`, blocked gives `partial`, a wait timeout gives `partial`, and a start error gives `failed` with the record `exited`. A closed connection during the start gives `partial` with the record `unresolved`. `agent_not_running` during the wait gives `failed` with the record `exited`. A reply without a name after the wait gives `ok`. The fake server checks that the record file exists when the start request arrives. The test checks the start parameters, the wait target, and the record fields with `harness`, `cwd`, and `terminal_id`. The fake replies use the typed `AgentSessionInfo` when present. Process cases: one named foreground entry records the PID, name, start, and optional session. A multiple entry, absent entry, shell entry, unknown name, wrapper-only entry, or unstable start gives `unresolved`. A wait timeout with the proven ownership keeps `live`. An absent session alone passes. A recorded session that is missing or different fails. A reused PID and an unreadable `/proc` fail.

### V3 Send and stop tests

- Proves: AC3, AC4.
- Command: `go test ./internal/agent/ -run 'TestSend|TestStop' -v`, red and green.
- Expected: the prompt request without `wait`, the exact nested `wait` shape, and the outcomes for `agent_blocked`, `agent_prompt_stalled`, `agent_not_ready`, and `timeout`. The stop cases: `spawned` false refuses with no request. `pane_not_found` from `pane.get` gives `unresolved`. A `terminal_id` difference gives `unresolved`. A process PID, name, start, or session difference gives `unresolved` and no close. The success case shows the record `stopped` when the close request arrives. The post-close read gives `closed`, `surviving`, or `unknown`, and each result keeps or clears the project metadata as AC4 gives. A close error gives `failed` with the record `live` and `last_error`.

### V4 Live transcript

- Proves: AC2, AC3, AC4.
- Use one unique fixture home with the marker `.hive/README.md` and one absolute
  tested binary. Record the binary SHA256. Keep each earlier report.
- Command, from a pane of the hive session, with a new tab in the own workspace of the worker:

```sh
FIX=$(mktemp -d)
mkdir -p "$FIX/.hive"
: > "$FIX/.hive/README.md"
BIN="$FIX/hive-$(date +%s)"
go build -o "$BIN" .
sha256sum "$BIN"
NAME="t10-check-$(date +%s)"
herdr tab create --workspace "$HERDR_WORKSPACE_ID" --cwd "$FIX"
PANE=<root pane ID from the output>
"$BIN" --home "$FIX" agent spawn --pane "$PANE" --role coordinator --name "$NAME" --harness claude --model {{T10_SMOKE_MODEL}}
"$BIN" --home "$FIX" agent send --agent "$NAME" --wait --timeout 120 "Reply with the word pong."
herdr pane read "$PANE" --lines 20
"$BIN" --home "$FIX" agent list
"$BIN" --home "$FIX" agent stop --agent "$NAME"
echo "exit $?"
herdr pane get "$PANE"
"$BIN" --home "$FIX" agent list --all
herdr tab list --workspace "$HERDR_WORKSPACE_ID"
rm -rf "$FIX"
```

- Expected: the spawn prints the name and the outcome `ok`. The pane read shows the word pong. The list shows the agent `live`. The stop exits with 0 and lists the effect `pane closed`. The post-close read returns one of `closed`, `surviving`, or `unknown`, and the project metadata follows that result. `pane get` answers `pane_not_found`. The last list shows the record `stopped`. The tab list no longer shows the new tab, because its only pane is gone.

### V5 Gate

- Proves: AC5.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- The pane choice for a manager and the workspace of a project. T13 adds them.
- The pane choice for a worker, the brief prompt, and the task assignment. T14 adds them.
- The update of the project record after a manager stop. T13 adds it.
- `hive agent resolve` for an `unresolved` record. T20 adds it.

## Changes

2026-10-18. Add T09 as a direct prerequisite. Add the typed process identity,
the profile process-name method, the process reader, the exact stop check, and
the post-close topology result. Reuse the existing T04 record fields and add
only the new identity fields. Replace the live transcript with a unique fixture
home and an absolute tested binary.
2026-10-18. Add the narrow `internal/lib/herdr/types.go` line. Add only the
optional `agent_session` field to `AgentInfo` and `PaneInfo`. Add the typed
`AgentSessionInfo` in `agent.go`. Keep optional absence allowed. Keep the
recorded-session mismatch refusal. Do not change the planned agent record.
