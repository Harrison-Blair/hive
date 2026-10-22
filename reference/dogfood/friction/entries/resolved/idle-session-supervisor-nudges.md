---
title: "An idle session burns tokens on supervisor nudges"
status: "Resolved 2026-10-19"
opened: 2026-10-19
decision: "Fixed in 16-quiet-no-nudge (848ef0b7): an unread quiet record never causes a nudge."
---

**Summary:**
On 2026-10-19, with all tasks landed and both managers on hold,
the supervisor kept the coordinator awake. Each idle manager writes a `quiet`
record to the coordinator inbox every ten minutes. The supervisor types a
nudge into the coordinator pane for each unread record. Each nudge costs one
coordinator turn, which costs tokens, and a user who is away pays for it. The
coordinator's own file watch doubled the cost before it was stopped. The
workaround was to stop both managers and the supervisor by hand. Hive has no
hush mode: a state in which the supervisor holds routine kinds (`quiet`,
`completed`, `verified`) and nudges only for `blocked`, `exited`,
`unresolved`, and urgent messages. The firstmate `/quiet` mode is the model.

**Reproduction steps:**
1. Spawn two managers with no goals. Leave the coordinator idle.
2. Wait more than ten minutes.
3. Each manager writes a `quiet` record. The supervisor types
   `hive: 1 new inbox record, run: hive --home ... inbox` into the coordinator
   pane for each one. The coordinator spends a turn on each nudge.
