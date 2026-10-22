---
title: "A scratch coordinator fails with agent_name_taken on the live herdr server"
status: "Open"
opened: 2026-10-21
decision: ""
---

**Summary:**
On 2026-10-21, worker `t17-worker-1` ran a coordinator start in a scratch home
for task `17-coordinator-start`. Herdr refused it with `agent_name_taken`.
Herdr keeps each agent name unique across one herdr server, and the real
coordinator of the home already had the name `coordinator`. The base code on
`dev` gives the same clash. A live check of the coordinator start must use a
scratch herdr server, which the worker did with session `t17-live`. A scratch
home that shares the live server cannot test the coordinator start.

**Reproduction steps:**
1. Run the real hive coordinator in the live herdr server.
2. In a scratch home on the same server, start the coordinator, for example
   with bare `hive` in a new pane.
3. Herdr refuses the agent name with `agent_name_taken`.
