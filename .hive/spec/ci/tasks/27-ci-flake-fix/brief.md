# CI flake fix

| Item | Value |
| --- | --- |
| Wave | CI flake fix |
| Tier | strong |
| Depends on | none. Base: `dev` at `99ed618d` or later |
| Grant | `grants.md 2026-10-19 ci flake fix task` |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

Four tests fail the GitHub Actions gate (`.github/workflows/ci.yml`, Ubuntu)
from time to time. Find the cause of each failure and fix it. Do not change a
decision in this brief. When the code does not agree with this brief, stop and
report it with `hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

Failures in the last 40 runs. Read each log with
`gh run view ID --log-failed`.

| Package | Test | Failure | Runs |
| --- | --- | --- | --- |
| `internal/tui` | `TestRunHomeReportsInputFailure` | `race detected during execution of test` | 39313104438, 39301292389, 39300420861, 39305650266 |
| `internal/lib/inbox` | `TestParallelAppendsLoseNoRecord` | `inbox_test.go:200: store lock timeout` | 39311856837 |
| `internal/supervise` | `TestWatchFinishedMarkReachesStop` | see the log | 39311829750 |
| `internal/lib/store` | `TestGroupedReadSeesEqualVersions` | see the log | 39309318840 |

On 2026-10-19, a local parallel race run also failed one time on
`TestExitSignalSIGTERM` with an empty process-start file. Check it with the
same method. When it fails, fix it in this task. When it does not fail in the
R2 runs, record that.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| The test files and the code files of the four packages above, and of the package of `TestExitSignalSIGTERM` | R1. Only the files that each fix needs. |
| `testdata/ste-baseline.json` | Only to add the brief copy with 0. |
| `spec/hive-v1/tasks/<task-id>.md` | R4. New. A copy of this brief. |

Do not edit other files. Do not change `.github/workflows/ci.yml`. Do not add
Windows.

## Requirements

### R1 Cause and fix

- For each failure, find the cause. Give it in the evidence with the log
  lines and the code lines.
- The `internal/tui` failure is a real data race. Find it. Fix the code or the
  test. Do not hide it.
- A longer timeout is acceptable only when the test checks correctness and not
  speed. The evidence must say why.
- Never weaken, skip, or delete a test.

### R2 Reproduction

For each package, run before and after the fix:

```sh
go test -race -count=50 -cpu=1,2 ./PACKAGE/ -run TEST
taskset -c 0 go test -race -count=50 -cpu=1,2 ./PACKAGE/ -run TEST
```

Show a failure before the fix and none after it in the evidence. When a
failure does not reproduce before the fix, record the runs and explain the
cause from the CI log and the code.

### R3 CI

Push the task branch. Then open one pull request with `gh pr create --base dev
--head task/<task-id>`. Cite the grant `grants.md 2026-10-19 ci flake fix pull
request` in its body. Wait for the CI run of the pull request with
`gh pr checks --watch`. Record the run ID and its result in the evidence. When
the run fails, fix the cause and push again. Do not merge the pull request.

### R4 Brief copy

Copy this brief to `spec/hive-v1/tasks/<task-id>.md` with a baseline of 0.
When the copy causes a violation, change the words of that sentence and keep
its meaning. Record each change in the evidence.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Read the cause of each failure in the evidence | Each cause has log and code evidence. The tests remain enabled and intact. Each longer timeout has a reason. |
| V2 | The R2 runs before and after the fix | A failure before the fix, or a recorded reason. No failure after the fix. |
| V3 | `git diff dev --stat` | Only the files of the source table changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |
| V5 | `gh pr checks` of the pull request | The CI run passes on the last pushed commit. |

## Finish

Commit on the task branch with a subject such as `fix(test): remove the CI
flakes`. Push the task branch. Write the evidence. Run `hive task complete`.
Send the commit SHA, the pull request URL, and the V1-V5 results to the manager with
`hive inbox send`.
