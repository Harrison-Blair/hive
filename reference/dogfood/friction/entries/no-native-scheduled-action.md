---
title: "Hive has no native command for a scheduled action"
status: "Open"
opened: 2026-10-26
decision: "Use the authorized one-shot user timer and effective-time override. This entry requests no product fix."
---

**Issue:** Hive has no native command to run an action at a future time.

**Status:** Open.

**Summary:**
User preference c202 changes new builder models at 06:00 America/New_York
on 2026-10-26, which is 10:00 UTC. Before that instant, all three builder
presets stay on Pi with `opencode-go/deepseek-v4.1-flash:xhigh`.
At that instant, new builders use `claude` with `claude-opus-5-5`.
Running agents keep their models. Other agent choices do not change.

The coordinator uses a one-shot systemd user timer,
`hive-worker-model-20261026-0600.timer`, with the matching `.service`.
The timer has `OnCalendar=2026-10-26 10:00:00 UTC` and `AccuracySec=1s`.
The service runs `/usr/bin/python3` with the pending action
`.hive/tmp/worker-model-20261026-0600/apply.py` in the home.
It changes only the three builder choices in `overrides/presets.md`
and sends Hive notices. It records `result.json` beside the action.

This transient timer requires the local systemd user manager to remain
available at the effective time. The record does not claim execution
while the machine or user manager is unavailable.
`overrides/worker-model-schedule.md` and the effective-time section of
`overrides/presets.md` govern each new builder spawn.
The pending timer and script must remain until the action runs.

**Reproduction steps:**
1. Ask Hive to change a model preference at a future instant.
2. Inspect the available Hive commands. No scheduled-action command exists.
3. Record the effective-time override and use the authorized local user timer.
4. Read the clock and effective policy before each new builder spawn.

**Evidence and limit:**
Coordinator project inbox records 191 to 193 name the schedule and workaround.
`systemctl --user cat hive-worker-model-20261026-0600.timer`
shows the exact calendar time. The matching service names the action above.
This entry records a missing capability. It requests no product or runtime change.
