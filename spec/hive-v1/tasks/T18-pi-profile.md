# T18 pi profile

| Item | Value |
| --- | --- |
| Task ID | T18-pi-profile |
| Wave | 4 |
| Tier | small |
| Depends on | T10, CP2 |
| Blocks | CP3, CP5 |
| Branch | `task/T18-pi-profile` |
| Worktree | `projects/hive/worktrees/T18-pi-profile/` |
| Evidence | `evidence/T18-pi-profile/` |

## Overview

This task adds the harness profile for Pi, the herdr agent kind `pi`.
A profile gives the kind, the native arguments for a model, and the two
start timeouts. It takes one line in the registry. The spawn core of T10 does the rest. The installed `pi --help`
shows the model flag `--model <pattern>`, with a value of the form `provider/id` and an optional `:<thinking>` suffix. The worker checks the flag against the installed
version and records the choice in the evidence.

## User flows

1. A manager runs `hive agent spawn --pane w3:p1 --role worker --project hive --task 12-spawn-worker --harness pi --model <model>`. Pi starts in the pane with that model, and hive waits until it is idle.
2. The manager runs `hive agent send --agent worker-hive-1 --wait --timeout 120 "Reply with the word pong."`. The reply appears in the pane.
3. The manager runs `hive agent stop --agent worker-hive-1`. Hive closes the pane. herdr closes the workspace when that pane was its last one.
4. `hive doctor` shows `ok harness pi` when the binary is on `PATH`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/lib/harness/pi.go` | New. The profile. |
| `internal/lib/harness/pi_test.go` | New. |
| `internal/lib/harness/registry.go` | Add one line. |
| `internal/lib/harness/doc.go` | Add the map line. |

## Requirements

### AC1 Profile

- The kind is `pi`. With a model, the arguments hold `--model` and the model. Without a model, the arguments are empty. The profile does not pass `--provider`, because the model pattern carries the provider. The extra arguments come last.
- The profile gives `ProcessNames()` with the actual executable name `pi`. It does not return `node`, `python`, or a shell name.
- The model value stays one unchanged argument. A value such as `anthropic/claude-sonnet-5-5:high` keeps its slash and its suffix.
- Pi 0.87.1 maps native DeepSeek `:medium` to null. Supported-level clamping selects the effective `high`. Record the requested `:medium` and the effective `high` separately. Keep the requested argument. This is native normalization of the same provider and model. It does not change a global catalog or config. PiSol `:high` stays effective `high`. A wrong or absent model blocks.
- The measured start is the time from the `agent.start` request to the first settled status. The settled statuses are `idle`, `blocked`, and `done`. The worker takes the time from the `duration_ms` of the `agent.wait` log event of the spawn in V2. The first live run uses the default of 120000 ms.
- The start timeout and the ready timeout are two times the measured start. The floor is 120000 ms and the ceiling is 300000 ms. herdr rejects a start timeout outside `(3000, 300000]` with `invalid_agent_timeout`. A start timeout below the real start makes herdr drop the name with no error, so the floor protects the spawn. When two times the measured start exceeds 300000 ms, the profile uses 300000 ms and the worker reports it to the merger.
- The profile passes `harness.Check` of T10. That check covers the timeout rules above.
- The profile file has a comment for each dialog that blocks the first start, such as a trust or a login dialog. Hive reports `blocked` and does not answer a dialog.

### AC2 Live spawn

- A spawn with `--harness pi` through the core of T10 reaches `idle` with the outcome `ok`.
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
- Command: `go test ./internal/lib/harness/ -run 'TestPi|TestLookup' -v`, red and green.
- Expected: the arguments with and without a model, the extra arguments, and a `provider/id:thinking` value as one unchanged argument. Also a start timeout above 3000 ms and at most 300000 ms, and the lookup by kind. The process-name test accepts `pi` and rejects `node`, `python`, `sh`, and `bash`. The evidence records the requested `:medium` and the effective `high` separately.

### V2 Live transcript

- Proves: AC2.
- Use one unique fixture home with the marker `.hive/README.md` and one absolute
  tested binary. Record the binary SHA256. Discover the exact installed Pi model
  first. Record the requested model and the actual effective model.
- Set the state directory of the child harness to a path inside `$FIX`, with the harness's own variable. For Pi use `PI_CODING_AGENT_SESSION_DIR=$FIX/pi-sessions`. Record the SHA256 of the real state directory of the harness before and after the run. The fixture cleanup removes only `$FIX`.
- Command, from a pane of the hive session, with a new workspace for the check. The name gets a suffix from `date +%s`. herdr names are unique for the session, and T16 to T19 run in parallel.

```sh
FIX=$(mktemp -d)
mkdir -p "$FIX/.hive"
: > "$FIX/.hive/README.md"
BIN="$FIX/hive-$(date +%s)"
go build -o "$BIN" .
sha256sum "$BIN"
NAME=t18-check-$(date +%s)
herdr workspace create --cwd "$FIX" --label "$NAME" --no-focus
"$BIN" --home "$FIX" agent spawn --pane <root pane of the new workspace> --role coordinator --name "$NAME" --harness pi --model {{T18_MODEL}}
echo "exit $?"
"$BIN" --home "$FIX" agent send --agent "$NAME" --wait --timeout 120 'Reply with the word pong.'
echo "exit $?"
herdr pane read <root pane of the new workspace> --source visible
"$BIN" --home "$FIX" agent stop --agent "$NAME"
echo "exit $?"
"$BIN" --home "$FIX" agent show "$NAME"
herdr workspace list
rm -rf "$FIX"
```

- Expected: the spawn prints the name and the outcome `ok` with exit 0. The send reports the settled status and exits 0. A separate pane read shows the reply with the word pong. The read shows the model name. The stop prints `ok` with exit 0. The show prints the status `stopped`. The workspace list no longer shows the workspace of the check, because its pane was the last one. The evidence records the measured start, the requested model, and the actual effective model. A `:medium` request can show the effective `high` after native normalization. A wrong or absent model leaves AC2 `BLOCKED`. The real state directory of the harness has the same SHA256 before and after the run.
- A spawn that ends `partial` with `agent blocked` goes into the evidence as a separate run. The worker answers the dialog in the pane by hand, stops the agent, and runs the check again from the spawn. Only a run with the outcome `ok` proves AC2.
- When a run fails before the stop, the worker closes the workspace of the check with `herdr workspace close <workspace>`.

### V3 Gate

- Proves: AC3.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- Changes to the spawn core. A quirk that needs one goes to the merger as a brief change request.

## Changes

2026-10-18. Add CP3 as a dependent. Add the profile process-name method with
`pi`, the native `:medium` normalization record, and the exact model discovery.
Replace the live transcript with a unique fixture home and an absolute tested
binary.
2026-10-18. Add the D10 settled-status wording. The send reports the settled
status and exits 0. A separate pane read shows the reply. The send does not
print the reply text.
2026-10-18. Correct the V2 live transcript from the T18 verification report
`t18-verify-01.md`, blocker B01 and observation N03. Read the output from the
pane. Set the state directory of the child harness inside `$FIX` and record the
SHA256 of that directory before and after the run.
