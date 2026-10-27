---
title: "A repair builder spawned while the old builder is live does not become the acting agent"
status: "Open"
opened: 2026-10-27
decision: ""
---

**Issue:** `hive agent spawn --profile builder --task ID` skips the
assignment when the current acting agent of the task is live. The new builder
then cannot run `hive task complete`.

**Status:** Open.

**Summary:**
Task `193-vocabulary-config` (stenographer) failed verify. Its first builder
`t193-builder-1` had completed the task but stayed live at a Claude Code
permission prompt (an `rm` in its report command). The manager spawned
`t193-builder-2` for the repair. `assignTask` in
`internal/agent/spawn_worker.go` saw a live acting agent and kept
`t193-builder-1`. After the fix, `hive task complete` refused:
"not the acting agent of 193-vocabulary-config: t193-builder-2". The spawn
printed no warning. Workaround: after the old builder stopped, the new
builder ran `hive task assign` under its own name.

**Reproduction steps:**
1. Let a builder complete a task and stay live (for example at a prompt).
2. Record a verify fail.
3. Spawn a new builder for the task.
4. Run `hive task complete` as the new builder and observe the refusal.
