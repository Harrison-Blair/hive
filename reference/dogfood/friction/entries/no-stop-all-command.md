---
title: "Hive has no command to stop all agents"
description: "hive agent stop takes one NAME, so the end of a session means one stop for each agent"
status: "Open"
opened: 2026-10-22
decision: ""
---

**Issue:** `hive agent stop` takes exactly one NAME
(`cmd/agent/stop/stop.go:17`). No command stops every agent of a session.

**Summary:** On 2026-10-22 the user asked how to shut down hive. No command did
it, and the user did several steps by hand. The handoff, TUI, and supervisor
steps of that time no longer exist. The stop step remains: the user or the
coordinator stops each agent, and then the user exits the coordinator.

**Reproduction steps:**
1. Run bare `hive` and spawn two agents with `hive agent spawn`.
2. Try to stop both with one command.
3. No command exists. Run `hive agent stop NAME` once for each agent.
