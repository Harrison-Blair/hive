# T19 cursor profile

| Item | Value |
| --- | --- |
| Task ID | T19-cursor-profile |
| Wave | 4 |
| Tier | small |
| Depends on | T10, CP2 |
| Blocks | CP5, unless the user skips the task |
| Branch | `task/T19-cursor-profile` |
| Worktree | `projects/hive/worktrees/T19-cursor-profile/` |
| Evidence | `evidence/T19-cursor-profile/` |

## Overview

This task adds the harness profile for Cursor agent, the herdr agent kind `cursor`.
A profile gives the kind, the native arguments for a model, the two start
timeouts, the exit keys, and the stop timeout. It takes one line in the
registry. The spawn core of T10 does the rest. The installed `cursor-agent --help`
shows the model flag `--model <model>`, with bracket parameters in some model names. The worker checks the flag against the installed
version and records the choice in the evidence.

This task is optional. The user can skip it when the Cursor agent has no quota.
The merger then sets the status to skipped with the reason. CP5 waits for this
task until the merger lands it or the user skips it.

## User flows

1. A manager runs `hive agent spawn --pane w3:p1 --role worker --project hive --task 12-spawn-worker --harness cursor --model <model>`. Cursor agent starts in the pane with that model, and hive waits until it is idle.
2. The manager runs `hive agent send --agent worker-hive-1 --wait --timeout 120 "Reply with the word pong."`. The reply shows in the pane.
3. The manager runs `hive agent stop --agent worker-hive-1`. Hive sends the exit keys of the profile, and the pane shows a shell.
4. `hive doctor` shows `ok harness cursor-agent` when the binary is on `PATH`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/lib/harness/cursor.go` | New. The profile. |
| `internal/lib/harness/cursor_test.go` | New. |
| `internal/lib/harness/registry.go` | Add one line. |
| `internal/lib/harness/doc.go` | Add the map line. |

## Requirements

### AC1 Profile

- The kind is `cursor`. With a model, the arguments hold `--model` and the model. Without a model, the arguments are empty. The extra arguments come last.
- The model value stays one unchanged argument. A value with bracket parameters keeps them.
- The measured start is the time from the `agent.start` request to the first settled status. The settled statuses are `idle`, `blocked`, and `done`. The worker takes the time from the `duration_ms` of the `agent.wait` log event of the spawn in V2. The first live run uses the default of 120000 ms.
- The start timeout and the ready timeout are two times the measured start. The floor is 120000 ms and the ceiling is 300000 ms. herdr rejects a start timeout outside `(3000, 300000]` with `invalid_agent_timeout`. A start timeout below the real start makes herdr drop the name with no error, so the floor protects the spawn. When two times the measured start exceeds 300000 ms, the profile uses 300000 ms and the worker reports it to the merger.
- The profile passes `harness.Check` of T10. It gives three stop values. `ExitKeys()` holds the keys that end Cursor agent from its prompt. `ForceExitKeys()` holds the keys for `--force`. `StopTimeout()` bounds the wait for the exit. The worker takes the exit command from the documentation of Cursor agent. It writes each key as a logical key name for `agent.send_keys`: one key for each typed character, then `enter`. V2 measures the exit. The evidence records the keys and the measured stop time.
- The profile file has a comment for each dialog that blocks the first start, such as a trust or a login dialog. Hive reports `blocked` and does not answer a dialog.

### AC2 Live spawn

- A spawn with `--harness cursor` through the core of T10 reaches `idle` with the outcome `ok`.
- A prompt with `--wait` gets a reply.
- The stop with the exit keys of the profile ends the agent within `StopTimeout()`. The record shows `stopped`, and the pane shows a shell.

### AC3 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The profile gives the kind, the arguments, the timeouts, and the exit keys, and the registry finds it. | V1 |
| AC2 | A live spawn, send, and stop succeed with this harness. | V2 |
| AC3 | The gate passes. | V3 |

## Verification items

### V1 Profile tests

- Proves: AC1.
- Command: `go test ./internal/lib/harness/ -run 'TestCursor|TestLookup' -v`, red and green.
- Expected: the arguments with and without a model, the extra arguments, and a model value with bracket parameters as one unchanged argument. Also a start timeout above 3000 ms and at most 300000 ms, exit keys that are not empty, and the lookup by kind.

### V2 Live transcript

- Proves: AC2.
- Command, from a pane of the hive session, with a new tab in the own workspace of the worker. The name gets a suffix from `date +%s`. herdr names are unique for the session, and T16 to T19 run in parallel.

```sh
go build .
NAME=t19-check-$(date +%s)
herdr tab create --workspace "$HERDR_WORKSPACE_ID"
./hive --home "$PWD" agent spawn --pane <pane of the new tab> --role coordinator --name "$NAME" --harness cursor --model <model>
echo "exit $?"
./hive --home "$PWD" agent send --agent "$NAME" --wait --timeout 120 'Reply with the word pong.'
echo "exit $?"
herdr agent read "$NAME" --source visible
./hive --home "$PWD" agent stop --agent "$NAME"
echo "exit $?"
./hive --home "$PWD" agent show "$NAME"
herdr pane read <pane of the new tab> --lines 5
herdr tab close <the new tab>
```

- Expected: the spawn prints the name and the outcome `ok` with exit 0. The send prints the reply with the word pong. The read shows the model name. The stop prints `ok` with exit 0. The show prints the status `stopped`. The pane read shows a shell prompt. The tab is gone at the end. The evidence records the measured start, the exit keys, and the measured stop time.
- A spawn that ends `partial` with `agent blocked` goes into the evidence as a separate run. The worker answers the dialog in the pane by hand, stops the agent, and runs the check again from the spawn. Only a run with the outcome `ok` proves AC2.
- When the Cursor agent has no quota, the worker records the failed run and reports it. The user then decides on the skip. A partial run does not pass AC2.

### V3 Gate

- Proves: AC3.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- Changes to the spawn core. A quirk that needs one goes to the merger as a brief change request.

## Changes

None.
