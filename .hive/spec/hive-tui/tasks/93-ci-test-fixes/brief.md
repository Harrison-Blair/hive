# CI test failures after the tidy fix

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, repair after wave C |
| Tier | decent |
| Depends on | none. Base: `dev` at `9f848584` or later |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Task 87-tidy-gate made the module tidy. CI then ran the tests for the first
time since 7204d699. The step "Test with the race detector" fails on `dev` at
`9f848584` (run 39442344823). Three tests fail on the CI runner and pass on
this host. Make each test pass on the runner without a weaker assertion. When
the code does not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `internal/doctor`: `TestSupervisorBinaryCheckWarnsWhenOlderThanInstalledHive`
  and `TestSupervisorBinaryCheckWarnsWhenExecutableIsDeleted` fail with
  `read supervisor executable: readlink /proc/PID/exe: no such file or
  directory`. `startFakeSupervisor` in `internal/doctor/health_test.go`
  copies the `sleep` binary to a file named `supervisor` and starts it.
- Hypothesis, not proved: on the runner, `sleep` is a multi-call binary (for
  example uutils coreutils) that selects its tool from the name of the
  program. A copy named `supervisor` exits at once, and `/proc/PID/exe` is
  then gone. Prove or disprove this. For example, run the copy on an Ubuntu
  container or read the runner image documentation. Record the result.
- `internal/tui`: `TestRunHomeLogsKeysOnStringReader` in
  `internal/tui/logs_screen_test.go` fails on the runner: the output does not
  show `▶ 8 Logs` and `level: warn`. The worker of 87-tidy-gate saw it fail
  once on this host in a full race run, then pass. It depends on timing.
- Task 88-tui-terminal-io changes `Run` in `internal/tui/tui.go` at the same
  time. Do not change `internal/tui/tui.go` unless the fix needs it. When it
  does, keep the change small and name it in the evidence.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/doctor/health_test.go` | R1. |
| `internal/doctor/` non-test files | R1, only when the cause is in the check. |
| `internal/tui/logs_screen_test.go` and test helpers in `internal/tui/` | R2. |
| `internal/tui/` non-test files | R2, only when the cause is in the program. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy, when `hive task create` did not write it. |

Do not edit other files. Do not edit `.github/workflows/ci.yml`.

## Requirements

### R1 Doctor tests

Find the cause. Make the fake supervisor a process that stays live with a
known executable path on any Linux host, for example a small test helper
binary that the test builds, or the test binary itself started with a helper
flag. Keep each assertion. When the cause is in the check, fix the check and
add a test.

### R2 Logs key test

Find the race. Make the test wait for a defined state, not a time. Keep the
assertion that the program shows Logs and the cycled level. When the cause is
in the program (for example a key that arrives before the screen data and is
lost), fix the program and add a model test for it.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Reproduce each failure before the fix. For R1, use a multi-call `sleep` or the proven cause. For R2, run `go test -race -count=50 -run TestRunHomeLogsKeysOnStringReader ./internal/tui/` or run under CPU load. | Each failure shows on the base. Record the output. |
| V2 | The same runs on the task branch | Each passes. R2 passes 50 of 50 runs. |
| V3 | Push the task branch and open no pull request. Run `gh workflow run` only when the workflow has `workflow_dispatch`. The workflow has none. Report "not run" for CI here. The manager checks CI after the land. | Recorded. |
| V4 | `git diff origin/dev --stat` | Only the files of the source table changed. |
| V5 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0. |

## Finish

Commit on the task branch with a subject such as `fix(test): make the doctor
and logs tests pass on CI`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V5 results to the
manager with `hive inbox send`. Read only your own inbox.

## Scope change: chat safety

A TUI check uses a scratch home (a temp copy of a marked home, set with
`HIVE_HOME`, and a temp `XDG_CONFIG_HOME`), or it does not open the Chat
input. On the real home, never press `i` on the Chat screen and never press
Enter there. A chat message from an agent has the authority of the user.

## Repair 1

Task 88-tui-terminal-io landed on `dev` (merge `ef2dde2e`). It also changed
`internal/tui/logs_screen_test.go`, so this branch has a merge conflict with
`origin/dev`. Do only this repair:

1. Merge `origin/dev` into the task branch.
2. Resolve the conflict in `internal/tui/logs_screen_test.go`. Keep the change
   of 88 and the R2 fix of this task. Keep each assertion of both sides.
   Resolve a conflict in `testdata/ste-baseline.json` by keeping each row.
3. Run `go test -race -count=50 -run TestRunHomeLogsKeysOnStringReader ./internal/tui/`.
   It must pass 50 of 50.
4. Run the gate of `AGENTS.md`.
5. Commit with subject `merge(dev): resolve logs test conflict`. Push the
   task branch. Add a `Repair 1` section to `evidence/V2.md` with the
   results of steps 3 and 4.
6. Make sure that `git merge-tree --write-tree origin/dev HEAD` exits 0.
7. Run `hive task complete`. Send the commit SHA to the manager with
   `hive inbox send`.

Do not change other files. Do not run the TUI on the real home.
