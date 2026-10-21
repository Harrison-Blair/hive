---
title: "A spawn after a stop fails with first tab busy, and a retry works"
status: "Open"
opened: 2026-10-21
decision: ""
---

**Summary:**
On 2026-10-21, the manager of hive spawned an agent for task `6-cp5-record`
two times after `hive agent stop` closed the task workspace. Each time the
first `hive agent spawn` failed with `hive: first tab busy: stop the agent or
pass --pane`. The failed spawn left an exited agent record (`t6-worker-2` with
pane `w6A:p1`, then `t6-verifier-2` with pane `w6H:p1`) and no live pane. The
same spawn command, run again at once, made a live agent (`t6-worker-3`,
`t6-verifier-3`). The workaround is to run the spawn again. Other tasks of
the same session did not show this.

**Reproduction steps:**
1. Run a task to `completed` or to a verify fail. Run `hive agent stop` for
   its agent. The output says `workspace closed`.
2. Run `hive agent spawn --role verifier --task ID ...` or `--role worker`.
3. The spawn fails with `first tab busy`. `hive agent show` for the new name
   gives status `exited`.
4. Run the same spawn again. It makes a live agent.
