---
title: "No hive path repairs a verified task"
status: "Open"
opened: 2026-10-19
decision: "Add `hive task reopen` (verified -> assigned, re-verify required) (wave 3)."
---

**Summary:**
On 2026-10-19, `hive land --task 12-friction-log --mode merge`
failed with a merge conflict in `testdata/ste-baseline.json`. Task 11 landed
first, and each task added one baseline line. The task had the status
`verified`. `hive agent spawn --role worker` refused it with `task is
verified`. `hive task assign` refused it with `invalid transition: verified ->
assign`. The workaround was a Pi agent that the manager started by hand with
`herdr agent start` in a new tab of the task workspace. That agent merged
`origin/dev` into the task branch, kept both lines, ran the gate, and pushed.
Then `hive land` passed. The repair has no hive agent record and no second
verification record.

**Reproduction steps:**
1. Make two tasks that each add one line at the same place in a file.
2. Verify both tasks. Land the first task.
3. Run `hive land` for the second task. The merge fails with a conflict.
4. Run `hive agent spawn --role worker` for the second task. Hive returns
   `task is verified`.
5. Run `hive task assign` for the second task. Hive returns
   `invalid transition: verified -> assign`.
