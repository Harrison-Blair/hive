---
title: "A completed task cannot be reopened or edited"
status: "Open"
opened: 2026-10-20
decision: ""
---

**Summary:**
`hive task reopen` accepts a verified task, and `hive task edit` accepts a
created or assigned task. A completed task accepts neither. On 2026-10-20 the
repair worker of 85-logs-screen completed Repair 2. Before a verifier ran,
task 83 landed on `dev`, and the branch got a new conflict in
`internal/tui/tui.go`. The manager could not add Repair 3 to the brief or
reopen the task. It had to spawn a verifier only to record a fail, so that the
task went back to `assigned`.

**Reproduction steps:**
1. Run `hive task complete` on a task.
2. Run `hive task reopen --task ID --reason TEXT`: `invalid transition:
   completed -> reopen`.
3. Run `hive task edit --task ID --brief FILE`: `edit needs created or
   assigned`.
