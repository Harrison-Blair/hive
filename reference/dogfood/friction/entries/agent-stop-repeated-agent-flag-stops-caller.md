---
title: "Agent stop with a repeated --agent flag stops the caller"
status: "Open"
opened: 2026-10-28
decision: ""
---

**Issue:** In `hive agent stop`, `--agent` names the agent to stop. Other
commands use `--agent` for the caller. The manager ran
`hive agent stop --agent t205-builder-1 --agent manager-stenographer` to stop
its builder and name itself as the caller. The last flag value won, and hive
stopped the manager. Hive gave no error for the repeated flag.

**Status:** Open.

**Summary:**
On 2026-10-28 at 22:47:56 UTC, `manager-stenographer` (generation 2)
stopped itself while it ran the stenographer v0.17.0 merge and release.
The command closed pane `w14N:p1` and workspace `hive:stenographer`.
The log of request `r-2b746cac` gives no caller, so the coordinator had to
read the manager transcript to find the cause. The user did not stop the
manager.

**Reproduction steps:**
1. Spawn an owner planner and a builder.
2. From the planner, run
   `hive agent stop --agent BUILDER --agent PLANNER`.
3. Observe that hive stops the planner, not the builder, with no error.
