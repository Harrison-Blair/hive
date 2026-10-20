# Close a verified task that will not land

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, repair after wave C |
| Tier | decent |
| Depends on | 102-selection-marker (order only). Base: `dev` |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

A task with status `verified` that does not land stays open. It gives ghosts
on the Agents screen, and doctor reports its worktree. Two gaps cause this.
`hive task cancel` refuses a verified task. `hive land --mode merge` cannot
record a task whose branch is already on the base branch: it records the base
tip as the merge commit. The friction entry is
`reference/dogfood/friction/entries/verified-task-cannot-cancel.md`. Do not
edit it. When the code does not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- The status machine is the `transitions` map in
  `internal/task/transition.go`. `verified` has only `reopen`.
- Hive considers a task landed when its status is `verified` and its delivery
  has `pushed` true. Hive has no `landed` status.
- `landMerge` in `internal/land/merge.go` runs `git merge --no-ff` of the
  task branch when the task has no delivery commit. When the branch is
  already on the base, Git makes no commit and prints `Already up to date`.
  `landMerge` then records the base tip as the delivery commit.
- `gitx.IsMerged` in `internal/lib/gitx/gitx.go` reports whether a branch is
  merged into the base.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/task/transition.go`, `internal/task/cancel.go`, and their tests | R1. |
| `internal/land/merge.go`, `internal/land/render.go`, and their tests | R2. |
| `internal/lib/gitx/` and its tests | R2, only when a helper is necessary. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy, when `hive task create` did not write it. |

Do not edit other files.

## Requirements

### R1 Cancel a verified task

- `hive task cancel` moves a `verified` task to `cancelled` when the task has
  no delivery, or a delivery with `pushed` false.
- It refuses a landed task (a delivery with `pushed` true) with the error
  `task is landed: cancel refused`. The record does not change.
- Hive stores the reason as it does for each other cancel.

### R2 Land a branch that is already on the base

- Before the merge, `landMerge` checks whether the task branch is already
  merged into the base. When it is, it runs no merge. It records the tip of
  the task branch as the delivery commit, and it pushes the base as it does
  after a merge.
- The command output has the effect `merge` with the action
  `already-merged` and the path of the branch tip. The text output says that
  the branch was already on the base.
- A branch that is not on the base lands as before.

### R3 Tests

Write each test first. Run it on the base and check that it fails. Use a
temp Git repository with a fixture task record. Cover verified-task
cancellation with no delivery, with `pushed` false, and with `pushed` true.
Cover landing a branch that is already on the base. Check that its delivery
commit is the branch tip and that the base has no new commit. Cover landing a
branch that is not on the base. Check that landing keeps the current
behavior. Use literal expected values.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run the R3 tests against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Build the task binary. Use a scratch home for `HIVE_HOME` and a temp directory for `XDG_CONFIG_HOME`. Use a fixture project with a temp Git repository and a bare temp origin. Create three verified tasks: one without a delivery, one whose branch is already on `dev`, and one landed task. Run `cancel` on task one, `land --mode merge` on task two, and `cancel` on task three | Task one has status `cancelled`. Task two meets Hive's landed definition: status `verified` and delivery `pushed` true. Its delivery commit matches its branch tip. The `dev` tip stays the same. Task three's cancel command refuses and leaves its record unchanged. |
| V3 | `git diff origin/dev --stat` | Only the files of the source table changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0. |

Never run a check on the real home or on a registered project repository.
Never work in workspace `wBC`. Stop and close only the pid and pane that you
started.

## Finish

Commit on the task branch with a subject such as `fix(task): close a
verified task that will not land`. Push the task branch. Write the evidence.
Run `hive task complete`. Send the commit SHA and the V1-V4 results to the
manager with `hive inbox send --to manager-hive`.
