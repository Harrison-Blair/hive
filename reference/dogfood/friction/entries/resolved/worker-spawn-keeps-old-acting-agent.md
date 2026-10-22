---
title: "A worker spawn into an assigned task keeps the old acting agent"
status: "Resolved 2026-10-19"
opened: 2026-10-19
decision: "Fixed in 19-worker-respawn (78381ba1): the spawn sets the acting agent and refuses while the old worker is live."
---

**Summary:**
On 2026-10-19, the manager spawned `t11-worker-2` with
`hive agent spawn --role worker --pane` to repair 11-ci-workflow after a
verify fail. The task still named `t11-worker-1` as the acting agent.
`hive task complete` from `t11-worker-2` failed with `not the acting agent`.
The workaround was `hive task assign --task 11-ci-workflow --agent
t11-worker-2`, then `hive task complete`.

Resolution: Live check 2026-10-19: after a verify fail on 18-task-reopen, `hive agent spawn --role worker --task 18-task-reopen` with no `--name` and no `--pane` picked `t18-worker-2`, set it as the acting agent, and reopened the task workspace `wDJ` on the same worktree.

**Reproduction steps:**
1. Let a worker finish a task. Let the verifier record a fail.
2. Spawn a second worker for the same task with `--pane`.
3. In the second worker, run `hive task complete`.
4. Check that hive returns `not the acting agent`.
