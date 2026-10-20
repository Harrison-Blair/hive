---
title: "A verified task that never landed cannot be cancelled"
status: "Open"
opened: 2026-10-20
decision: ""
---

**Summary:**
`hive task cancel` refuses a task with status `verified`: `invalid transition:
verified -> cancel`. A verified task that will not land stays open, gives
ghosts on the Agents screen, and keeps a worktree that doctor reports. On
2026-10-20 the user cancelled `10-pr47-final`. The manager ran `hive task
reopen` (verified to assigned) and then `hive task cancel` as a workaround.
The worktree held only untracked evidence, so `hive worktree remove` refused
it as dirty. The manager copied the evidence to
`.hive/tmp/archive/10-pr47-final/` and used `--force`.

**Reproduction steps:**
1. Take a task with status `verified` that has no delivery.
2. Run `hive task cancel --task ID --reason R`: `invalid transition:
   verified -> cancel`.
3. Run `hive task reopen --task ID --reason R`, then `hive task cancel`: the
   task is cancelled.

**Second case (2026-10-20):** stenographer `45-ci-fix` has status
`verified` and a delivery with commit `4a94e73` and `pushed: false`. The
commit is on `origin/dev` and `origin/main`. The record does not read landed.
`hive task cancel` gives the same refusal. A `hive land --mode merge` for a
task with a delivery commit skips the merge and only pushes the base branch
(`internal/land/merge.go`, the `retry` path). The local `dev` of stenographer
is equal to `origin/dev`, so that push has nothing to send and sets `pushed`
to true. This is a workaround, not a clear command: no command records a land
for a commit that is already on the base, and no command cancels a verified
task. Same family as `completed-task-cannot-reopen`.
