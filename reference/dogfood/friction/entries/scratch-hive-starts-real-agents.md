---
title: "Bare hive in a scratch home starts real coordinator and supervisor tabs"
status: "Open"
opened: 2026-10-20
decision: ""
---

**Summary:**
Bare `hive` runs the startup of G9: in herdr, it starts a coordinator and a
supervisor in their own tabs when they are missing. A worker or verifier that
runs bare `hive` with `HIVE_HOME` set to a scratch home gets no live
coordinator and a stale supervisor claim, so each run starts a new
coordinator tab and a new `hive supervise` process in the workspace of the
caller. On 2026-10-20 the worker of 101-store-list-long-keys left 4
`hive supervise` processes and 8 tabs in workspace `wQ8`. One tab held a
Claude session at its folder-trust prompt. The scratch homes were deleted
after the runs, and the supervisors kept running.

**Reproduction steps:**
1. In a herdr pane, make a scratch home with a `.hive/README.md` marker.
2. Run `HIVE_HOME=<scratch> hive` and quit the TUI.
3. Run `herdr tab list` in the workspace: a coordinator tab and a supervisor
   tab are new. `pgrep -a 'hive supervise'` shows a new process.
