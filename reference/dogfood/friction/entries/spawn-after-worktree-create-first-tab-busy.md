---
title: "Builder spawn right after worktree create can fail with first tab busy"
status: "Open"
opened: 2026-10-28
decision: ""
---

**Issue:** `hive agent spawn --profile builder --task ID` run at once after
`hive worktree create --task ID` printed "hive: first tab busy: stop the
agent or pass --pane". The first tab held only the new idle shell.

**Status:** Open.

**Summary:**
On project stenographer, the manager ran `hive worktree create` and then
`hive agent spawn` for tasks 198 and 202 in one loop. The 198 spawn worked.
The 202 spawn failed with "first tab busy" and the task kept no acting agent.
`herdr pane list` showed pane `w142:p1` as an idle shell with
`agent_status` "unknown". The same spawn command worked when it was run
again a few seconds later; hive named the agent `t202-builder-2`. The
failed spawn probably raced the shell start in the new workspace.

**Reproduction steps:**
1. Run `hive worktree create --task ID` and the builder spawn in one shell
   command, for two tasks in a row.
2. Observe "first tab busy" on one spawn.
3. Run the spawn again and observe success with suffix `-2`.
