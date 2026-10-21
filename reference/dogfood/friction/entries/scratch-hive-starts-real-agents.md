---
title: "Bare hive in a scratch home starts real coordinator and supervisor tabs"
status: "Open"
opened: 2026-10-20
decision: "Task 8-no-startup adds the opt-in switch --no-startup and HIVE_NO_STARTUP=1"
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

**Second occurrence (2026-10-20, 12:03 local):** worker t91-worker-1 ran its
task binary on the real home with the herdr variables set. It then sent
SIGINT to PID 1518955 and closed tab `wBC:t11` as its own cleanup. That PID
and tab were the user's TUI, which the coordinator had started in pane
`wBC:p1A`. The coordinator started the TUI again in tab `wBC:t14`. The worker
picked the process and tab by name and place, not by the pid and pane ID that
herdr gave it at start. Rule now in each brief: never signal a process or
close a tab, pane, or workspace that the agent did not start, and never work
in workspace `wBC`.

**Third occurrence (2026-10-21):** worker t4-worker-1 of task
4-agents-equal-columns ran a live check (V3) with a temp home. Bare `hive`
left two idle shell tabs, `coordinator` and `supervisor`, in workspace `w5T`.
The temp home was deleted after the run. No process ran in the tabs. The
manager closed both panes. The real supervisor (supervisor-g4) was not
touched.

**Mitigation (2026-10-21):** task 8-no-startup of feature `tui-v2` adds
the switch `hive --no-startup` and the variable `HIVE_NO_STARTUP=1`, from
decision 34 of the spec. With either one, bare `hive` opens the TUI. It makes
no tab, starts no coordinator and no supervisor, and does not rename the tab.
The worker skill and the verifier skill tell each agent to run bare `hive`
only with the switch or the variable. The entry stays open: the switch is
opt-in, and automatic detection of a scratch home is out of scope.
