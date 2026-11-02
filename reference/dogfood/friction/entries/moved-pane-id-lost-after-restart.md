---
title: "The old ID of a moved pane stops working after a herdr restart"
description: "herdr resolves the ID of a pane before a pane.move only until the server restarts, so a hive record of a moved agent then shows the agent as gone"
status: "Open"
opened: 2026-11-02
decision: ""
---

**Issue:** A hive record keeps the pane ID from the spawn. A `pane.move` to
another tab or workspace gives the pane a new ID. herdr 0.9.5 continues to
resolve the old ID for the `pane.*` methods, against its documented contract
(see `reference/herdr/addressing.md`). That alias does not survive a server
restart. After a restart, `pane.get` with the old ID returns `pane_not_found`
while the pane is alive with its new ID. `hive agent list` then shows the
agent as `gone`, and `hive agent send` and `hive agent stop` cannot reach it.

**Summary:** On 2026-11-02, a probe on a scratch herdr 0.9.5 server moved
pane `w1:p4` into tab `w3:t1`. The move returned the new ID `w3:p2`. Before
the restart, `pane get w1:p4` returned the pane `w3:p2`. After `herdr server
stop` and a new start of the same session, `pane list` held `w3:p2`, and
`pane get w1:p4` returned `pane_not_found`. The same probe found that herdr
does not reuse a closed pane or workspace ID, also after a restart. No real
agent was moved, so the effect on a hive agent is inferred.

**Reproduction steps:**
1. Start a scratch herdr server. Make two workspaces.
2. Make a pane `P` in the first workspace. Move it with `herdr pane move P
   --tab TAB --split right` into the second workspace.
3. Run `herdr pane get P`. It returns the moved pane.
4. Stop the server and start the same session again.
5. Run `herdr pane get P`. It returns `pane_not_found`, and `herdr pane list`
   still holds the moved pane with its new ID.
