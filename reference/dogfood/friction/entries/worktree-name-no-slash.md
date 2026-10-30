---
title: "A spawn worktree name cannot contain a slash"
description: "hive agent spawn --worktree rejects a name with /, so a branch such as friction/SLUG needs a push with HEAD:friction/SLUG"
status: "Open"
opened: 2026-10-30
decision: ""
---

**Issue:** `hive agent spawn --worktree NAME` checks NAME with the agent name
rule (`internal/agent/spawn/spawn.go`, `agent.ValidateName`). That rule
rejects `/`. The worktree name is also the branch name, so a worktree branch
cannot have a name such as `friction/SLUG` or `feat/x`. The agent must push
with `git push origin HEAD:friction/SLUG` to get that remote branch name.

**Summary:** On 2026-10-30 this point was the second point of the entry "A
report sent to a pane with an open dialog is lost". It moved to this entry
when the pull request that changes `hive agent send` resolved that entry.

**Reproduction steps:**
1. Run `hive agent spawn --project hive --worktree friction/x --harness claude --model M --name n`.
2. Spawn fails in phase `validate`: the `--worktree` name is not valid.
