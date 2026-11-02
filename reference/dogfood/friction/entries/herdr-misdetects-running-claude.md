---
title: "herdr can misdetect a running Claude agent as another agent"
description: "herdr reported agent opencode with status unknown for a pane whose foreground process was claude, so hive agent list showed a live agent as unknown"
status: "Open"
opened: 2026-11-02
decision: ""
---

**Issue:** herdr can name the wrong agent for a pane and give it the status
`unknown`, while the harness runs. `hive agent list` and the graph show the
status of `pane.get`, so a live agent shows as `unknown`. Hive cannot use the
herdr status to decide that an agent exited. `pane.process_info` gave the
correct process.

**Summary:** On 2026-11-02, the record `ror-desktop-coordinator` in `~/.hive`
had pane `w177:p1`. `hive agent list` showed it as `unknown`. `herdr pane get
w177:p1` gave `"agent":"opencode"`, `"agent_status":"unknown"`, and
`"agent_session":null`. `herdr pane process-info --pane w177:p1` gave one
foreground process, `claude`. The cause of the wrong detection is not known.
It was seen one time, on herdr 0.9.5.

**Reproduction steps:**
1. Not known. Run `hive agent list` and look for a record with status
   `unknown`.
2. For that pane, run `herdr pane get PANE` and `herdr pane process-info
   --pane PANE`.
3. Compare the `agent` field with the foreground process.
