---
title: "No hive path repairs a verified task"
status: "Resolved 2026-10-19"
opened: 2026-10-19
decision: "Fixed in 18-task-reopen (4f2bd6a5): hive task reopen moves a verified task to assigned, keeps the old check in verification_history, and land needs a new pass verify."
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

On 2026-10-19 the same conflict occurred for 15-inbox-ownership, after
17-verifier-project and 16-quiet-no-nudge landed. The worker and verifier had
exited, and herdr had closed the task workspace. The manager made a workspace
with `herdr workspace create`, started `t15-merge-1` with `herdr agent start`,
and that agent merged `origin/dev` twice and pushed. That agent has no hive
record, so it sent its report with `--agent manager-hive`.

Resolution: wave 3 did not use the reopen path. Task 19 landed first with no conflict. Task 18 then failed its verify, so it was assigned, not verified, and the respawned worker `t18-worker-2` merged `origin/dev` during the repair. The first live use of `hive task reopen` is still to come.

**Reproduction steps:**
1. Make two tasks that each add one line at the same place in a file.
2. Verify both tasks. Land the first task.
3. Run `hive land` for the second task. The merge fails with a conflict.
4. Run `hive agent spawn --role worker` for the second task. Hive returns
   `task is verified`.
5. Run `hive task assign` for the second task. Hive returns
   `invalid transition: verified -> assign`.
