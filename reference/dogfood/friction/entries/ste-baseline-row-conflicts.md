---
title: "Each brief copy row conflicts in the STE baseline"
status: "Open"
opened: 2026-10-20
decision: ""
---

**Summary:**
Each spec task adds one row with 0 for its brief copy to
`testdata/ste-baseline.json`. Two tasks that run in parallel add rows at the
same place, so the second land fails with a merge conflict. Each conflict
needs a reopen, a repair worker, and a new verifier. On 2026-10-20 this
occurred for 85-logs-screen (repair 2) and 89-screen-board-audit (repair 1
after 94-chat-user-only landed). Wave D has ten parallel tasks.

**Reproduction steps:**
1. Create two spec tasks from the same `dev` commit.
2. Let each worker add its brief copy row to `testdata/ste-baseline.json`.
3. Land the first task.
4. Run `git merge-tree --write-tree origin/dev origin/task/<second>`: a
   conflict in `testdata/ste-baseline.json`.

**More occurrences (2026-10-20):** 93-ci-test-fixes (with 88), and
100-no-backgrounds and 97-friction-detail-panel after 93 and 98 landed. The
manager repairs one task at a time, because a parallel repair conflicts again
after the first land. Possible fixes: one baseline file for each task brief
copy, or a test that gives each brief copy a baseline of 0 when no row exists.

**Related (2026-10-20):** a worker rewords its brief copy when the copy fails
`TestSTERatchet`. A later `hive task edit` (for a repair) copies the manager
brief into the worktree again. That undoes the rewording, and the repair
worker gets the same STE failure. This occurred for 93 and 97.

**Again (2026-10-21):** goal `g2-tui-v2` ran `8-no-startup` and `9-cellgrid`
at the same time. Task 8 landed first. `hive land` for task 9 then failed with
a merge conflict in `testdata/ste-baseline.json`. The manager reopened task 9,
a worker merged `dev`, and a second verifier ran. The conflict added one
worker and one verifier. CP5 task `6-cp5-record` had the same conflict after
`13-remove-bootstrap` landed.
