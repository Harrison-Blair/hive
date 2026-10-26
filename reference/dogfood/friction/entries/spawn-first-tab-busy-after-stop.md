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


**Occurrence: task 170 on 2026-10-26**
The ordinary builder had completed and its native record was stopped. The
actual Herdr actor, old pane `w13A:p1`, and recorded PID were absent.
A fresh native Codex verifier spawn with no `--pane` failed before its turn
with `first tab busy: stop the agent or pass --pane`. Hive recorded
`t170-verifier-1` as exited with pane `w13B:p1`. Actual pane get and process
reads then returned `pane_not_found`. No occupied user pane needed removal.

Coordinator record 265 authorized one retry of the same native spawn.
The retry used the accepted schema-4 CLI pin from source `72a338da`, exact
Codex `gpt-6.1-sol` with high reasoning, and no pane override. It succeeded
as `t170-verifier-2` in `w13C:p1`. No placeholder tab, manual registration,
metadata change, model substitution, or source fix was used. The source
candidate remained frozen. That verifier later reported an inventory error;
its source verdict is separate from the startup refusal. A corrected
inventory then passed a fresh independent verifier.

Task 170 evidence preserves `manager-verifier-spawn.json`, stderr and exit,
`manager-verifier-retry265/`, pane-absence receipts, and the correction
history. This occurrence uses the same documented workaround. The underlying
startup issue remains open. No speculative product repair is requested.
