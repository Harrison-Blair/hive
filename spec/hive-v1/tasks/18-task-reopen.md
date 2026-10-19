# Task reopen

| Item | Value |
| --- | --- |
| Wave | wave 3 of 4, task A |
| Tier | decent |
| Depends on | none. Base: `dev` at `39a35e6c` or later |
| Grant | `grants.md 2026-10-19 task reopen task` |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

Today a verified task has no transition. When `hive land` fails with a merge
conflict, hive has no path to repair the task. This task adds
`hive task reopen`. It moves a verified task back to `assigned`, and the task
must get a new verify before `hive land` accepts it. Do not change a decision
in this brief. When the code does not agree with this brief, stop and report
it with `hive inbox send --to manager-hive TEXT --agent NAME`.

The friction entry is
`reference/dogfood/friction/entries/no-hive-path-repairs-verified-task.md`.
Read it. Do not edit a file in `reference/dogfood/friction/`. The manager
resolves the entry after the landing.

## Known facts

- `internal/task/transition.go` has the table `transitions`. `verified` has no
  transition. `applyTransition` changes the record for each command.
- `internal/lib/records/task.go`: `Task` has `ControllingRole`, `ActingAgent`,
  `Status`, and `Verification *Verification` (the last check only).
- `internal/land/land.go` refuses a task whose status is not `verified`.
- The commands of `hive task` are in `cmd/task/<name>/`. Each one calls into
  `internal/task/`.
- `hive task show` prints the task record.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `cmd/task/reopen/` | R1. New command. |
| `cmd/task/task.go` | R1. Register the command. |
| `internal/task/` | R1, R2. New file for reopen, and edits to `transition.go`. Update `doc.go`. |
| `internal/lib/records/task.go` | R2. The history field. |
| `internal/land/land.go` | R3. Only when the fix needs it. |
| Tests beside each changed file, `cmd/task_lifecycle_test.go` | R1 to R3 tests. |
| `.agents/skills/hive-manager/SKILL.md` | R4. |
| `internal/agent/templates/project-agents.md` | R4. |
| `internal/agent/manager_prompt_test.go` | R4 test. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `spec/hive-v1/tasks/<task-id>.md` | R6. New. A copy of this brief. |

Do not edit other files.

## Requirements

### R1 Command

Add `hive task reopen --task ID --reason TEXT`.

- `--reason` must have a non-empty value.
- It moves a `verified` task to `assigned`. Each other status fails with
  `invalid transition: STATUS -> reopen` and changes nothing.
- Only the controlling role of the task can run it. Hive resolves the caller
  from `--agent` or `--pane`, as the other task commands do. The caller's
  inbox key (`manager-<project>` or `coordinator`) must equal
  `ControllingRole`. Each other caller fails with
  `not the controlling role: NAME` and changes nothing.
- The acting agent stays the same.
- It obeys the command output rules of `AGENTS.md`: text by default, one JSON
  object with `--json`. It logs the transition with the reason.

### R2 History

- Reopen clears `Verification` to `null`.
- Before it clears it, reopen appends the old verification, the reason, the
  caller, and the time to a new list field `verification_history` in the task
  record. Hive never deletes an item of this list.
- `hive task show` prints the history.

### R3 Land after reopen

`hive land` accepts a reopened task only after a new pass verify. A reopened
task that is `assigned`, or `completed` with no new verify, fails the land.
Also make `hive land` refuse a `verified` task whose `Verification` is `null`
or not a pass.

### R4 Instructions

In both instruction files, add this rule: when `hive land` fails with a merge
conflict, reopen the task with `hive task reopen`, spawn a repair worker, and
then spawn a fresh verifier with no `--pane`. Add a check of this text to
`internal/agent/manager_prompt_test.go`. Run it, and make sure that it fails
before you change the template. Write the text in ASD-STE100 Strict. The counts
in `testdata/ste-baseline.json` must not rise.

### R5 Tests

Work test-first. Write each test, run it, and make sure that it fails for the
correct reason before you change the code. The tests cover these cases:
- A controlling role reopens a verified task.
- A worker cannot reopen a task.
- Reopening an `assigned` task fails.
- Reopening a task without a reason fails.
- A second reopen keeps the earlier verification.
- Land fails before a new verify and passes after a new verify.
Show in the evidence that each new test fails against the old code, then passes.

### R6 Brief copy

Copy this brief to `spec/hive-v1/tasks/<task-id>.md` with a baseline of 0.
When the copy causes a violation, change the words of that sentence and keep
its meaning. Record each change in the evidence.

## Not in scope

The worker spawn into an assigned task (wave 3 task B), agent names, the
friction folder, `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each R5 and R4 test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Live check in a scratch home (`hive --home DIR`) with a fake verified task: reopen it as the manager, then `hive task show`, then `hive land` | Status `assigned`, `verification` is `null`, the history holds the old check, and land fails with `task is assigned`. |
| V3 | Read the R4 text in both instruction files | Present in both. |
| V4 | `git diff dev --stat` | Only the files of the source table changed. |
| V5 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `feat(task): add hive task
reopen`. Push the task branch. Write the evidence. Run `hive task complete`.
Send the commit SHA and the V1-V5 results to the manager with
`hive inbox send`. Read only your own inbox.
