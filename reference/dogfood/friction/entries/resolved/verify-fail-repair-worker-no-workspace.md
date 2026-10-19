---
title: "A verify fail leaves no workspace for the repair worker"
status: "Resolved 2026-10-19"
opened: 2026-10-19
decision: "Fixed in 19-worker-respawn (78381ba1): the spawn reopens the worktree workspace with worktree.open and updates the record."
---

**Summary:**
On 2026-10-19, `t16-verifier-1` recorded a fail for 16-quiet-no-nudge. The
verifier ran in `wD6:p1`, the root pane of the task workspace, after the worker
`t16-worker-1` exited. `hive agent stop --agent t16-verifier-1` closed `wD6`.
The task record still named `wD6:p1` as the worktree pane. Then
`hive agent spawn --role worker --task 16-quiet-no-nudge` failed with
`worktree pane mismatch: pane_id`, and `hive worktree create --task
16-quiet-no-nudge` failed with `task has a worktree`. The workaround was
`herdr workspace create --cwd WORKTREE --label 16-quiet-no-nudge`, then a
worker spawn with `--pane` and `--project`. The task record kept
`t16-worker-1` as the acting agent, so the brief told the repair worker to run
`hive task complete --agent t16-worker-1`. The fresh-verifier rule of wave 2
covers the re-verify. It does not cover the repair worker.

`hive task complete --agent t16-worker-1` then failed with `agent not found:
t16-worker-1`, because that agent was stopped. The fix was
`hive task assign --task 16-quiet-no-nudge --agent t16-worker-2` by the
manager. Then `hive task complete --agent t16-worker-2` passed.

Resolution: Live check 2026-10-19: after a verify fail on 18-task-reopen, `hive agent spawn --role worker --task 18-task-reopen` with no `--name` and no `--pane` picked `t18-worker-2`, set it as the acting agent, and reopened the task workspace `wDJ` on the same worktree.

**Reproduction steps:**
1. Spawn a worker for a task. Let it complete and exit.
2. Spawn a verifier with no `--pane`. Hive puts it in the task workspace.
3. Let the verifier record a fail. Run `hive agent stop` for the verifier.
4. Check that herdr closed the task workspace.
5. Run `hive agent spawn --role worker` for the task with no `--pane`. Hive
   returns `worktree pane mismatch: pane_id`.
6. Run `hive worktree create` for the task. Hive returns `task has a worktree`.
