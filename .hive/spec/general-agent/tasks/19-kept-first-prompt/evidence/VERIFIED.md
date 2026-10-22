# Verification: 19-kept-first-prompt

Verdict: **pass**. This verdict replaces the first verdict after the brief added the guard-test exception.

Verifier: `t19-verifier-2`, independent pi session, pane `w8K:p1`.
Code: `7ae66e6c`. Date: 2026-10-22 UTC.

| Item | Result | Observation |
| --- | --- | --- |
| V1 | pass | A new Codex directory showed a trust dialog. Spawn kept a live blocked record and sent a blocked record to the parent inbox. After manual dialog answers, the supervisor sent the first prompt once. Both times were present. |
| V-red | pass | All 11 new tests pass on task code. Nine fail on base code with the required compile scaffolding. The two guard tests pass on the base, as the revised brief permits. Each guard test fails when its guard is removed. Restoring the guards makes all supervisor tests pass. |
| V-gate | pass | `git merge dev --no-edit` reports Already up to date. Only allowed source files, the brief, the STE baseline, and evidence differ. All five gate commands pass. Formatting and module checks print nothing. |

## V-red

Used a detached scratch worktree at base `ae14fe29`. Copied the three changed test files. Ran:

```sh
go test -count=1 -v ./internal/agent/ ./internal/lib/records/ -run 'GeneralSpawnBlocked|FirstPromptTimes'
```

Both new tests fail on assertions. Added only the two record fields and a no-op `firstPromptRule` to compile the supervisor tests. Ran:

```sh
go test -count=1 -v ./internal/supervise/ -run KeptFirstPrompt
```

Seven tests fail on assertions. `TestKeptFirstPromptWaitsWhileBlocked` and `TestKeptFirstPromptFailedAttemptWriteDoesNotSend` pass, under the guard-test exception.

Ran the green command in the task worktree:

```sh
go test -count=1 -v ./internal/supervise/ ./internal/agent/ ./internal/lib/records/ -run 'KeptFirstPrompt|GeneralSpawnBlocked|FirstPromptTimes'
```

All 11 tests pass. In the scratch checkout, copied the task supervisor source and record type. Removed the idle/done status guard. Also changed the failed attempt write path to continue with `attempted = true`. Ran the supervisor command again. The blocked guard test fails with `prompts = 1; want none while blocked`. The write guard test fails with `prompts = 1; want none after a failed attempt write`. Restored the task source and ran again: all nine supervisor tests pass. No task source was changed.

## V1

Versions: herdr 0.9.5, protocol 22; Codex 0.160.1.
Built with `go build -o .hive/tmp/t19-reverify/hive .`.
Created a temp home at `.hive/tmp/t19-reverify/home` and a new Git repository at `/tmp/hive-t19-reverify-NXrCyc`. Created workspace `w8M` without focus.

Pane `w8M:p1` set `HIVE_HOME` to the temp home, registered `reverify-boss`, then ran the built binary:

```sh
hive agent spawn --role agent --name reverify-child --harness codex --model gpt-6.1-sol --allow-trust-dialog --cwd /tmp/hive-t19-reverify-NXrCyc --prompt 'Reply with the single word READY. Do not run a tool.'
```

Spawn exits 1 with a blocked message. The child record is live, has agent status blocked, and keeps the complete prompt with neither timestamp. The parent inbox holds one blocked record for the child.

In child pane `w8M:p2`, manually selected Skip for the update notice, Trust and continue for the empty repository, and Continue without trusting for hooks. Hive sent no dialog answer. Herdr then reported `done`.

Started `hive supervise` in owned pane `w8M:p3`, with the temp home. It printed one `first prompt reverify-child` line. The child received one complete prompt and replied READY. The record has:

```text
first_prompt_attempt_at: 2026-10-22T00:20:15.459999791Z
first_prompt_sent_at:    2026-10-22T00:20:15.766145049Z
```

After 24 seconds, the temp logs contain exactly one `agent.prompt` request and its reply, with request ID `r-0f91f778`. The child capture contains one role line. The supervisor also marks the shell-only parent exited during reconcile; this behavior is outside this task.

## Gate and scope

Ran these commands from the task worktree root:

```sh
git diff dev...HEAD --stat
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go mod tidy -diff
go vet ./...
golangci-lint run ./...
go test -race ./...
```

Formatting and module checks print nothing. Vet prints nothing. Lint prints `0 issues.`. Race tests pass for all packages. Full gate and test output is in `reverify-runs.md`.

## Cleanup

Stopped the child with `hive agent stop --agent reverify-child` in the temp home; it exits 0 and closes the child pane and tab. Sent Ctrl-C to the owned supervisor and confirmed the shell prompt. Closed workspace `w8M` and confirmed its absence. Removed only the new repository trust entry; the Codex config matches the pre-check backup byte for byte. Removed the scratch worktree, temp repository, temp home, binary, captures, and backup after recording the evidence. The task source remains unchanged.
