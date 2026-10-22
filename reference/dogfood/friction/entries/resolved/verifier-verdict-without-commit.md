---
title: "A verifier can record a verdict without a pushed VERIFIED.md"
status: "Resolved 2026-10-19"
opened: 2026-10-19
decision: "Fixed in 23-verify-clean-branch (27a883e9): hive task verify refuses a dirty worktree or an unpushed branch, and the verifier commits and pushes VERIFIED.md first."
---

**Summary:**
On 2026-10-19, `t18-verifier-2` ran `hive task verify` with a pass for
18-task-reopen, and the task became `verified`. The new pass `VERIFIED.md`
stayed uncommitted in the worktree. The branch still held the old fail
`VERIFIED.md` of `t18-verifier-1`. A land at that point would put a fail
record on `dev`. The manager saw it in `git status` of the worktree and told
the verifier to commit and push (75f8c6e2). The same day, `t19-verifier-1`
left corrected V1-V3 evidence uncommitted in the 19-worker-respawn worktree.
After the land, `hive worktree remove` refused with `worktree is dirty`.
`hive task verify` and `hive land` do not check that the branch holds a
committed `VERIFIED.md` with the same verdict.

**Reproduction steps:**
1. Let a verifier write `VERIFIED.md` and run `hive task verify` with no commit.
2. Check that the task status is `verified`.
3. Run `git status` in the worktree. `VERIFIED.md` is modified.
4. Run `hive land`. Hive does not refuse it.
