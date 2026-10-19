# 19 evidence fix

| Item | Value |
| --- | --- |
| Wave | wave 4 of 4, task D |
| Tier | small |
| Depends on | none. Base: `dev` at `5d879894` or later |
| Grant | `grants.md 2026-10-19 19 evidence fix task` |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

The evidence files V1 to V3 of 19-worker-respawn on `dev` name the base commit
`39a35e6c` for the green runs. The verifier `t19-verifier-1` wrote corrected
copies but did not commit them in
`/home/penguin/source/hive/projects/hive/worktrees/19-worker-respawn/spec/hive-v1/evidence/19-worker-respawn/`.
This task commits the corrected copies. Do not change Go code. When the files
do not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `spec/hive-v1/evidence/19-worker-respawn/V1.md`, `V2.md`, `V3.md` | R1. |
| `spec/hive-v1/evidence/19-worker-respawn/V4.md` | R1. Only the commit name. |
| `testdata/ste-baseline.json` | Only to add the brief copy with 0. |
| `spec/hive-v1/tasks/<task-id>.md` | R2. New. A copy of this brief. |

Do not edit other files. Do not change the old worktree. The manager removes
it after the landing.

## Requirements

### R1 Evidence

1. Copy the three corrected files from the old worktree into your worktree.
   Read the old worktree only.
2. Compare each copy with `VERIFIED.md` of 19-worker-respawn on `dev`. The
   commit names must agree with it. The task commit is `2c69b4ad`.
3. In `V4.md`, change each green-run commit name from `39a35e6c` to the task
   commit that `VERIFIED.md` names. Change nothing else.

### R2 Brief copy

Copy this brief to `spec/hive-v1/tasks/<task-id>.md` with a baseline of 0.
When the copy causes a violation, change the words of that sentence and keep
its meaning. Record each change in the evidence.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | `diff` each committed V1-V3 file with its copy in the old worktree | No difference. |
| V2 | `grep -n 39a35e6c spec/hive-v1/evidence/19-worker-respawn/V*.md` | Each match names the base of a red run, not a green run. |
| V3 | `git diff dev --stat` | Only the files of the source table changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `docs(evidence): correct the
19-worker-respawn commit names`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V4 results to the manager
with `hive inbox send`.
