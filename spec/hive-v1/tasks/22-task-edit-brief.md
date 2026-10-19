# Task edit brief

| Item | Value |
| --- | --- |
| Wave | wave 4 of 4, task B |
| Tier | decent |
| Depends on | none. Base: `dev` at `5d879894` or later |
| Grant | `grants.md 2026-10-19 task edit brief task` |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

Today a manager must edit a stored brief by hand in `.hive/state/briefs/`.
This task adds `hive task edit --task ID --brief FILE`. Do not change a
decision in this brief. When the code does not agree with this brief, stop and
report it with `hive inbox send --to manager-hive TEXT --agent NAME`.

The friction entry is
`reference/dogfood/friction/entries/task-edit-brief-command.md`. Do not edit a
file in `reference/dogfood/friction/`. The manager resolves the entry after the
landing.

## Known facts

- `internal/task/create.go` reads a brief with `readBrief` (limit, deadline,
  `-` for standard input) and stores it in `.hive/state/briefs/ID.md`.
- `hive task reopen` (`cmd/task/reopen/`, `internal/task/`) already checks
  that the caller's inbox key equals `ControllingRole`. Reuse that check.
- The commands of `hive task` are in `cmd/task/<name>/` and register in
  `cmd/task/task.go`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `cmd/task/edit/` | R1. New command. |
| `cmd/task/task.go` | R1. Register the command. |
| `internal/task/` | R1. New file for edit, and the shared helpers it needs. Update `doc.go`. |
| Tests beside each changed file, `cmd/task_lifecycle_test.go` | R2. |
| `.agents/skills/hive-manager/SKILL.md` | R3. |
| `internal/agent/templates/project-agents.md` | R3. |
| `internal/agent/manager_prompt_test.go` | R3 test. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `spec/hive-v1/tasks/<task-id>.md` | R4. New. A copy of this brief. |

Do not edit other files.

## Requirements

### R1 Command

Add `hive task edit --task ID --brief FILE`.

- It reads the brief as `hive task create` does, with `-` for standard input.
- It replaces the stored brief under the store lock. A failure leaves the old
  brief unchanged.
- Allow edits only when the task status is `created` or `assigned`. Each other
  status fails with `task is STATUS: edit needs created or assigned`.
- Only the controlling role of the task can run it. Each other caller fails
  with `not the controlling role: NAME`.
- It logs the change with the task ID, the caller, and the old and new sizes.
  It does not log the brief text.
- It obeys the command output rules of `AGENTS.md`.

### R2 Tests

Work test-first. Write each test before you change code. Run it. Check that it
fails for the expected reason.

Test these cases:

- A controlling role edits a brief. `hive task show --brief` prints the new text.
- An edit of a `completed` task fails. The old brief stays unchanged.
- An edit by a worker fails.
- An edit from standard input works.
- An edit with a brief that is too large fails. The old brief stays unchanged.

Record each new test's failing run against the base code and its passing run on
the task branch.

### R3 Instructions

In both instruction files, add `hive task edit --task ID --brief FILE` to the
commands, with one sentence: use it to change a brief before the work is
complete. Add a check of this text to `internal/agent/manager_prompt_test.go`.
Run it, and make sure that it fails before you change the template. Write the
text in ASD-STE100 Strict. The counts in `testdata/ste-baseline.json` must not
rise.

### R4 Brief copy

Copy this brief to `spec/hive-v1/tasks/<task-id>.md` with a baseline of 0.
When the copy causes a violation, change the words of that sentence and keep
its meaning. Record each change in the evidence.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each R2 and R3 test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Live check in a scratch home (`hive --home DIR`): edit an assigned task as the manager, then a completed task | The first edit changes `hive task show --brief`. The second fails and changes nothing. |
| V3 | `git diff dev --stat` | Only the files of the source table changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `feat(task): add hive task
edit --brief`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V4 results to the manager
with `hive inbox send`.
