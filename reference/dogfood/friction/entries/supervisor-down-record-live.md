---
title: "Supervisor stopped, its record stayed live, and no nudge came"
status: "Open"
opened: 2026-10-28
decision: ""
---

**Issue:** The supervisor stopped, but `hive agent list` still showed its
record as `live`. No process sent inbox nudges. A planner did not learn that a
builder completed its task.

**Status:** Open.

**Summary:**
At 2026-10-28T19:35:08-04:00 the supervisor logged `supervisor stop
generation=7 reason=signal`. Its tab and its pane `w14M:p3` then were absent.
The agent record `supervisor` still had status `live`. At 21:28 the builder
`t206-builder-1` ran `hive task complete`, which appended records 374 and 375 to
the inbox of `manager-hive`. No nudge reached the planner, and the user saw that
the planner did not get the result. `hive doctor` reported `supervisor down`.
The coordinator started a new supervisor in a new `supervisor` tab at about
21:31. The source of the signal is not known.

**Reproduction steps:**
1. Start hive with a live supervisor.
2. Stop the supervisor process with a signal, or close its tab.
3. Run `hive agent list`. Observe that the supervisor record is `live`.
4. Append an inbox record for an idle agent. Observe that no nudge comes.
