---
title: "A verify fail leaves no workspace and no verifier for a re-verify"
status: "Resolved 2026-10-19"
opened: 2026-10-19
decision: "By design. The manager skill fix landed in 17-verifier-project (e7031497): after a verify fail, spawn a fresh verifier with no --pane."
---

**Summary:**
On 2026-10-19, `t11-verifier-1` recorded a fail for
11-ci-workflow. Then the verifier stopped, and the supervisor closed the
worker workspace `wCK` when its last agent finished. After the repair,
`hive agent send` to the verifier returned `agent is stopped`, and
`herdr tab create --workspace wCK` returned `workspace_not_found`. The
workaround was a fresh verifier spawn with no `--pane`, which made the new
workspace `wCN` on the same worktree.

**Reproduction steps:**
1. Spawn a worker and a verifier for a task in the same workspace.
2. Let the verifier record a fail and finish.
3. Repair the task and complete it again.
4. Run `hive agent send` to the verifier. Hive returns `agent is stopped`.
5. Run `herdr tab create --workspace ID` for the task workspace. herdr
   returns `workspace_not_found`.
