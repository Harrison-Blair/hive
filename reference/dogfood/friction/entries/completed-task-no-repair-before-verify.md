---
title: "A manager cannot send a completed task back to a worker before the verify"
status: "Open"
opened: 2026-10-23
decision: ""
---

**Summary:**
On 2026-10-23 task 126-chat-modes-mouse completed with a small defect: the
brief copy had an STE baseline of 12, not 0. The worker stopped when it ran
`hive task complete`. The manager wanted a short repair before the verify.
`hive agent spawn --role worker` refused with `task is completed`, and
`hive task reopen` acts only on a verified task. The manager had to spawn the
verifier and let a verify fail start the repair.

**Reproduction steps:**
1. Complete a task with `hive task complete`.
2. Run `hive agent spawn --role worker --task ID`: `task is completed`.
3. Run `hive task reopen --task ID`: the command acts only on a verified task.
