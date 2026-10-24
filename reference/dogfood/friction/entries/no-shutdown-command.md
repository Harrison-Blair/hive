---
title: "Hive has no command to shut down a session"
status: "Open"
opened: 2026-10-22
decision: ""
---

**Summary:**
On 2026-10-22 the user asked how to shut down hive. Hive has no command for
it. A clean shutdown needs these manual steps:

1. The coordinator writes its handoff with `hive handoff write --scope home`.
2. Each manager writes its handoff with `hive handoff write --scope
   project:ID`.
3. The coordinator stops each manager with `hive agent stop`.
4. The user stops the TUI (ctrl+c two times), the supervisor (ctrl+c in its
   tab), and the coordinator.

When the user closes the panes without these steps, the handoffs are not
written, and the agent records stay `live` until the next bare `hive` startup
marks them stale.

**Reproduction steps:**
1. Run bare `hive` and spawn one manager or more.
2. Try to end the session with one command or one TUI action.
3. No command or TUI action exists. The user must do the manual steps above.
