---
title: "Attach on the supervisor fails with agent_not_found, and the error stays in the footer"
status: "Open"
opened: 2026-10-22
decision: ""
---

**Summary:**
On 2026-10-22 at 00:22:18 local time, the TUI logged a `tui error` with the
context `attach` and the error `agent_not_found: agent target w58:p5 not
found`. The user had pressed `a` on the supervisor. Pane `w58:p5` runs
`hive supervise`. herdr tracks only a pane that runs a known agent program,
for example claude, codex, or pi. Thus herdr has no agent in that pane, and
attach fails. The supervisor was live and correct.

The error stayed in the footer until 00:27 and later, with no time and no
context. The user read it as a current fault. The supervisor log also has
the same `agent_not_found` reply 13 times on the same day, from `agent.get`
on its own pane.

Expected behavior:
1. Attach focuses the pane when the pane has no herdr agent.
2. A footer error goes away after a short time or after the next key, or it
   shows its time and its context.

**Reproduction steps:**
1. Run bare `hive`.
2. Open 5 Health › Agents and select the live supervisor.
3. Press `a`.
4. The footer shows `error: agent_not_found: agent target PANE not found`,
   and the error stays there.
