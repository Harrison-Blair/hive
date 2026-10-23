---
title: "Agents overview draws a card from a stale record instead of the live one"
status: "Open"
opened: 2026-10-22
decision: ""
---

**Summary:**
The Agents overview shows some cards from old stale records. The user saw it on 2026-10-22 (dev f7990406 and later):

- The coordinator card shows "○ unknown · 2d". The live record coordinator-g8 is claude, live, working.
- The manager-qmeter card shows "claude · unknown". The live record manager-qmeter-g4 is pi, live, working. Only the stale records g1 to g3 are claude.
- The manager-stenographer card shows "claude · unknown" and is not hidden. Record g8 has exited, and g1 to g7 are stale.

Probable cause, from internal/tui/agents.go (not verified with a test):

- `agentIdentity` gives a record with no `parent` and no `replaces` the identity "name". A record with `replaces` gets the identity "lineage". Old records (for example manager-qmeter-g1 and g2, coordinator-g1 to g6) have neither field. Newer records (manager-qmeter-g4, coordinator-g8) have `replaces`. So one agent name falls into two identity groups.
- `currentAgentKeys` then marks one record as current in each group: a stale one and the live one.
- The card takes the stale current record. `agentStatus` returns "unknown" for a stale record.
- A stale record does not count as stopped, so `h` does not hide it.

**Reproduction steps:**
1. Have several records with one agent name, where the old ones have no `parent` or `replaces` and the newest live one has `replaces` (as in .hive/state/agents/manager-qmeter-g1..g4.json).
2. Run bare `hive` and open Agents.
3. See that the card shows the harness and the status of a stale record ("claude · unknown"), not the live record.
