# Verification: 19-kept-first-prompt

Verdict: **fail**. V-red differs from its expected output. No runtime defect was observed in these checks.

Verifier: independent pi session in pane `w8D:p1`.
Task code: `7288d2fc`. Verified merge: `499d53c4` (includes current local `dev`).
Date: 2026-10-22 UTC.

| Item | Result | Observation |
| --- | --- | --- |
| V1 | pass | A new codex directory showed a trust dialog. The live blocked record kept the prompt. The parent inbox received one blocked record. After manual dialog answers, the supervisor sent one prompt. Both times were present. |
| V-red | fail | All 11 new tests pass on task code. Nine fail on the base with the required compile scaffolding. `TestKeptFirstPromptWaitsWhileBlocked` and `TestKeptFirstPromptFailedAttemptWriteDoesNotSend` pass on that base. The brief requires each new test to fail on the base. Both guard tests fail with deliberate guard breaks. |
| V-gate | pass | Merged `dev` before the scope check. Only allowed source files, the brief, the STE baseline, and task evidence differ. All five gate commands exit 0. Formatting and module checks print nothing. |

## V-red

Base: `ae14fe29`, in a detached scratch worktree. Copied the three changed test files from the task.

```sh
go test -count=1 -v ./internal/agent/ ./internal/lib/records/ -run 'GeneralSpawnBlocked|FirstPromptTimes'
```

Both new tests fail on assertions. For supervise, added only the task record fields and a no-op `firstPromptRule` to compile the tests against the base. Then ran:

```sh
go test -count=1 -v ./internal/supervise/ -run KeptFirstPrompt
```

Seven supervisor tests fail on assertions; two guard tests pass. This agrees with the detailed worker output, but differs from the brief and the final claim in `V-red.md` that each new test fails on the base.

Green command on the task merge:

```sh
go test -count=1 -v ./internal/supervise/ ./internal/agent/ ./internal/lib/records/ -run 'KeptFirstPrompt|GeneralSpawnBlocked|FirstPromptTimes'
```

All 11 tests pass.

Deliberate breaks ran only in the scratch checkout with task source copied in. Removed the idle/done status guard. Also let a failed attempt write continue to the send. Ran the supervisor command again. The blocked-status test fails with `prompts = 1; want none while blocked`. The failed-attempt-write test fails with `prompts = 1; want none after a failed attempt write`. These results establish that both guard tests detect broken behavior. They do not establish a base failure.

Full test and gate output is in `verifier-runs.md`.

## V1: independent live run

Versions: herdr 0.9.5, protocol 22; codex-cli 0.160.1.

Built the task binary with `go build -o .hive/tmp/verify-hive .`.
Created a temp home at `.hive/tmp/verify-home`, with its own `.hive/README.md` marker. Created an empty Git repository at `/tmp/hive-t19-verifier-GnxRe2`. The real home was not used for the live commands.

Created workspace `w8F`, label `t19-verifier-live`, without focus. Its shell pane `w8F:p1` ran these commands with `HIVE_HOME` set to the temp home and the absolute path to the built binary:

```sh
hive agent register --role agent --name verify-boss
hive agent spawn --role agent --name verify-child --harness codex --model gpt-6.1-sol --allow-trust-dialog --cwd /tmp/hive-t19-verifier-GnxRe2 --prompt 'Reply with the single word READY. Do not run a tool.'
```

Spawn exits 1 with a blocked message. The record has `status: live`, `agent_status: blocked`, the complete first prompt, and no attempt or sent time. The parent inbox has one record with `kind: blocked`, `agent: verify-child`.

In the created child pane `w8F:p2`, manually selected Skip at the update notice, Trust and continue at the folder dialog, and Continue without trusting at the hook dialog. Hive did not answer the dialogs. Herdr then reports `done` for the idle child.

Created supervisor tab `w8F:t3`, pane `w8F:p3`. Ran the built binary as `hive supervise` with the temp `HIVE_HOME`. The supervisor prints:

```text
supervisor live, generation 1
exited record for coordinator
first prompt verify-child
```

The parent is a shell with no harness, so reconcile marks it exited. The child pane shows one complete first prompt and one `READY` response.

The record remains live and has:

```text
first_prompt_attempt_at: 2026-10-22T00:12:04.746970477Z
first_prompt_sent_at:    2026-10-22T00:12:05.053338275Z
```

After more supervisor passes, the temp log has exactly one `agent.prompt` request and one matching reply (`request_id: r-c7a8b2f8`). The supervisor still shows one first-prompt line. The captured pane has one first-prompt role line.

## Cleanup

Stopped `verify-child` through `hive agent stop --agent verify-child` in the temp home. It exits 0 and closes its pane and tab. Sent Ctrl-C to the owned supervisor pane and confirmed the shell prompt. Closed workspace `w8F`; the workspace list has no matching ID or label.

Removed only the trust entry for the scratch repository from the Codex config. The config then matches the pre-check backup byte for byte. Removed the scratch Git worktree, temp repository, temp home, binary, captures, and config backup after recording the results. No task source was changed for mutation tests.

## Required follow-up

Reconcile V-red with the observed guard-test results. The brief requires base failure for every new test. The two guard tests instead prove their sensitivity through deliberate breaks. This verdict does not waive that requirement.
