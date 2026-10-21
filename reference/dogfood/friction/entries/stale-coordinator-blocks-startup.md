---
title: "A stale coordinator record blocks bare hive startup"
status: "Open"
opened: 2026-10-20
decision: ""
---

**Summary:**
On 2026-10-20, the coordinator record had status `live` and named pane
`w2T:p1`. Herdr reported `pane_not_found` for that pane. Bare `hive` checked
only the record status. It did not start a new coordinator. Chat delivery
failed with `agent_not_found`. The supervisor subscribed to the missing pane
and exited with `pane_not_found`. The workaround was to confirm that the pane
was gone, mark the stale coordinator record `exited`, and start bare `hive`
again. Startup must check the recorded Herdr pane before it treats a
coordinator as live.

**Reproduction steps:**
1. Leave a coordinator agent record with status `live` after its Herdr pane
   closes.
2. Run bare `hive` in a Herdr pane of the home. Hive does not start a new
   coordinator.
3. Send a chat message in the TUI. Delivery reports `agent_not_found`.
4. Run `hive doctor`. It reports that the supervisor is down. The supervisor
   log reports `pane_not_found` for the stale coordinator pane.
