# T17 opencode profile

| Item | Value |
| --- | --- |
| Task ID | T17-opencode-profile |
| Wave | 4 |
| Tier | small |
| Depends on | T10, CP2 |
| Blocks | CP5 |
| Branch | `task/T17-opencode-profile` |
| Worktree | `projects/hive/worktrees/T17-opencode-profile/` |
| Evidence | `evidence/T17-opencode-profile/` |

## Overview

This task adds the harness profile for OpenCode, the herdr agent kind `opencode`.
A profile gives the kind, the native arguments for a model, and the two
start timeouts. It takes one line in the registry. The spawn core of T10 does the rest. The installed `opencode --help`
shows the model flag `-m, --model` with a value of the form `provider/model`. The worker checks the flag against the installed
version and records the choice in the evidence.

## User flows

1. A manager runs `hive agent spawn --pane w3:p1 --role worker --project hive --task 12-spawn-worker --harness opencode --model <model>`. OpenCode starts in the pane with that model, and hive waits until it is idle.
2. The manager runs `hive agent send --agent worker-hive-1 --wait --timeout 120 "Reply with the word pong."`. The reply shows in the pane.
3. The manager runs `hive agent stop --agent worker-hive-1`. Hive closes the pane. herdr closes the workspace when that pane was its last one.
4. `hive doctor` shows `ok harness opencode` when the binary is on `PATH`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/lib/harness/opencode.go` | New. The profile. |
| `internal/lib/harness/opencode_test.go` | New. |
| `internal/lib/harness/registry.go` | Add one line. |
| `internal/lib/harness/doc.go` | Add the map line. |

## Requirements

### AC1 Profile

- The kind is `opencode`. With a model, the arguments hold `--model` and the model. Without a model, the arguments are empty. The extra arguments come last.
- The profile gives `ProcessNames()` with the actual executable name `opencode`. It does not return `node`, `python`, or a shell name.
- The model value stays one unchanged argument. A value such as `anthropic/claude-sonnet-5-5` keeps its slash.
- The measured start is the time from the `agent.start` request to the first settled status. The settled statuses are `idle`, `blocked`, and `done`. The worker takes the time from the `duration_ms` of the `agent.wait` log event of the spawn in V2. The first live run uses the default of 120000 ms.
- The start timeout and the ready timeout are two times the measured start. The floor is 120000 ms and the ceiling is 300000 ms. herdr rejects a start timeout outside `(3000, 300000]` with `invalid_agent_timeout`. A start timeout below the real start makes herdr drop the name with no error, so the floor protects the spawn. When two times the measured start exceeds 300000 ms, the profile uses 300000 ms and the worker reports it to the merger.
- The profile passes `harness.Check` of T10. That check covers the timeout rules above.
- The profile file has a comment for each dialog that blocks the first start, such as a trust or a login dialog. Hive reports `blocked` and does not answer a dialog.

### AC2 Live spawn

- A spawn with `--harness opencode` through the core of T10 reaches `idle` with the outcome `ok`.
- A prompt with `--wait` gets a reply.
- The stop closes the pane of the agent. The record shows `stopped`, and the workspace of the check is gone.

### AC3 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The profile gives the kind, the arguments, and the timeouts, and the registry finds it. | V1 |
| AC2 | A live spawn, send, and stop succeed with this harness. | V2 |
| AC3 | The gate passes. | V3 |

## Verification items

### V1 Profile tests

- Proves: AC1.
- Command: `go test ./internal/lib/harness/ -run 'TestOpenCode|TestLookup' -v`, red and green.
- Expected: the arguments with and without a model, the extra arguments, and a `provider/model` value as one unchanged argument. Also a start timeout above 3000 ms and at most 300000 ms, and the lookup by kind. The process-name test accepts `opencode` and rejects `node`, `python`, `sh`, and `bash`.

### V2 Live transcript

- Proves: AC2.
- Use one unique fixture home with the marker `.hive/README.md` and one absolute
  tested binary. Record the binary SHA256. Discover the exact installed OpenCode
  model first. Record the requested model and the actual model.
- Command, from a pane of the hive session, with a new workspace for the check. The name gets a suffix from `date +%s`. herdr names are unique for the session, and T16 to T19 run in parallel.

```sh
FIX=$(mktemp -d)
mkdir -p "$FIX/.hive"
: > "$FIX/.hive/README.md"
BIN="$FIX/hive-$(date +%s)"
go build -o "$BIN" .
sha256sum "$BIN"
NAME=t17-check-$(date +%s)
herdr workspace create --cwd "$FIX" --label "$NAME" --no-focus
"$BIN" --home "$FIX" agent spawn --pane <root pane of the new workspace> --role coordinator --name "$NAME" --harness opencode --model {{T17_MODEL}}
echo "exit $?"
"$BIN" --home "$FIX" agent send --agent "$NAME" --wait --timeout 120 'Reply with the word pong.'
echo "exit $?"
herdr agent read "$NAME" --source visible
"$BIN" --home "$FIX" agent stop --agent "$NAME"
echo "exit $?"
"$BIN" --home "$FIX" agent show "$NAME"
herdr workspace list
rm -rf "$FIX"
```

- Expected: the spawn prints the name and the outcome `ok` with exit 0. The send prints the reply with the word pong. The read shows the model name. The stop prints `ok` with exit 0. The show prints the status `stopped`. The workspace list no longer shows the workspace of the check, because its pane was the last one. The evidence records the measured start, the requested model, and the actual model. A wrong or absent model leaves AC2 `BLOCKED`.
- A spawn that ends `partial` with `agent blocked` goes into the evidence as a separate run. The worker answers the dialog in the pane by hand, stops the agent, and runs the check again from the spawn. Only a run with the outcome `ok` proves AC2.
- When a run fails before the stop, the worker closes the workspace of the check with `herdr workspace close <workspace>`.

### V3 Gate

- Proves: AC3.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- Changes to the spawn core. A quirk that needs one goes to the merger as a brief change request.

## Changes

2026-10-18. Add the profile process-name method with `opencode`. Replace the
live transcript with a unique fixture home, an absolute tested binary, and exact
model discovery.
