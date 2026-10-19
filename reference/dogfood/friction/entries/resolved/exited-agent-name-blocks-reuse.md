---
title: "An exited agent name blocks reuse"
status: "Resolved 2026-10-19"
opened: 2026-10-19
decision: "Fixed in 19-worker-respawn (78381ba1): with no --name, hive picks the next free name tNN-ROLE-N."
---

**Summary:**
An exited agent name blocks a later spawn with `agent exists`.
Each retry needs a new agent name.

Resolution: Live check 2026-10-19: after a verify fail on 18-task-reopen, `hive agent spawn --role worker --task 18-task-reopen` with no `--name` and no `--pane` picked `t18-worker-2`, set it as the acting agent, and reopened the task workspace `wDJ` on the same worktree.

**Reproduction steps:**
1. Spawn an agent with the name `NAME`.
2. Let the agent exit.
3. Spawn another agent with the name `NAME`.
4. Check that hive returns `agent exists`.
