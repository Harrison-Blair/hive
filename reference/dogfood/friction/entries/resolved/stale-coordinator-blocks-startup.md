---
title: "A stale coordinator record blocks bare hive startup"
status: "Resolved 2026-10-21"
opened: 2026-10-20
decision: "Task 2 makes bare hive check the coordinator pane at startup, and task 3 adds the r restart on the Agents screen"
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

**Plan (approved by the user 2026-10-21):**
The user asked for a TUI action that restarts the coordinator or the
supervisor, and for stale records to stop blocking startup. This plan also
removes the manual workaround in `supervisor-runs-stale-binary.md`.

1. Bare `hive` counts a `live` coordinator record as dead when its Herdr pane
   is gone. It starts a new coordinator. Register marks the old record
   `stale` with the existing claim rule.
2. The coordinator and supervisor records show that bare `hive` started them.
   Restart can then close their old panes under the AGENTS.md rule.
3. On the Agents screen, the user selects the coordinator or supervisor node
   and presses `r`. Hive asks for confirmation only when the target is live.
   A dead or stale target restarts at once.
4. A coordinator restart closes the old pane, opens a new `coordinator` tab,
   and starts the configured harness with the startup prompt, as bare `hive`
   does.
5. A supervisor restart closes its pane through Herdr, opens a new
   `supervisor` tab, and runs `hive supervise`. The abrupt SIGHUP exit is
   accepted.
6. Old records get the correct status (`stale` or `stopped`). Hive keeps the
   files. `h` hides them. Hive deletes no record.
7. The TUI header marks the supervisor `old` when it runs an older binary than
   the installed `hive`, with the hint `r to restart`.

Delivery: task 1 is steps 1 and 2. Task 2 is steps 3 to 7 and starts after
task 1 lands. The manager lands each verified task on `dev` under the
2026-10-18 grant.
