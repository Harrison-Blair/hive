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
