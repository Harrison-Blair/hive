---
title: "A supervisor restart does not stop the old supervisor or close its tab"
status: "Open"
opened: 2026-10-22
decision: ""
---

**Summary:**
On 2026-10-22 at 00:22 local time, the user pressed `r` on supervisor-g5 in
Health › Agents and accepted the prompt. The log (request r-2547cda9) shows the outcome
`started`, a new tab `w58:tD` (pane `w58:pD`), and the warning `old pane kept
because Hive did not start it`. The old supervisor in tab `supervisor` (pane
`w58:p5`) was not stopped, and its tab stayed open. Bare `hive` made that tab
at startup, but the supervisor record does not mark it as started by hive.
The user expects the restart to stop the old supervisor and close its pane and
tab. Later the same morning, one supervisor (generation 6) was live in `w58:p5`.

**Reproduction steps:**
1. Run bare `hive`, so that startup makes the `supervisor` tab.
2. Open 5 Health › Agents, select the live supervisor, press `r`, and accept the prompt.
3. A new supervisor tab opens. The old `supervisor` tab stays, and the result
   says `old pane kept because Hive did not start it`.
