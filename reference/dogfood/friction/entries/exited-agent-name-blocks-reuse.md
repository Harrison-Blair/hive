---
title: "An exited agent name blocks reuse"
status: "Open"
opened: 2026-10-19
decision: "Hive picks the next free name when `--name` is absent (wave 3)."
---

**Summary:**
An exited agent name blocks a later spawn with `agent exists`.
Each retry needs a new agent name.

**Reproduction steps:**
1. Spawn an agent with the name `NAME`.
2. Let the agent exit.
3. Spawn another agent with the name `NAME`.
4. Check that hive returns `agent exists`.
