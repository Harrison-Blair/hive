---
title: "A manager spawn marks the live manager stale but leaves its pane running, and agent stop refuses it"
status: "Open"
opened: 2026-10-24
decision: ""
---

**Summary:**
On 2026-10-24 the coordinator swapped the qmeter and stenographer managers to a new model. It ran `hive agent stop manager-qmeter`, which failed because the name must go in `--agent`. It then ran `hive agent spawn --role manager --project qmeter`. Hive started the new manager in a new pane of the same workspace (wX8:p2) and marked the old live record (manager-qmeter-5, wX8:p1) as stale. The old Claude session still ran in wX8:p1. `hive agent stop --agent manager-qmeter-g5` failed with `agent is stale`. The coordinator closed wX8:p1 and wX9:p1 with `herdr pane close` after it checked the foreground argv. Until then, two manager sessions could act on the same project.

**Reproduction steps:**
1. With a live manager for project P, run `hive agent spawn --role manager --project P --harness H --model M`.
2. `hive agent list --all` shows the old record as stale. Its pane still runs the old harness.
3. Run `hive agent stop --agent OLD-KEY`. Hive refuses with `agent is stale`.
