---
title: "Unread inbox records of a stopped agent warn forever, with no way to clear them"
status: "Open"
opened: 2026-10-23
decision: ""
---

**Summary:**
On 2026-10-23 `hive doctor` and the Health tab of the TUI warn:
`ownerless inbox records: 2: t126-worker-1/4, t126-worker-1/5`. The check
follows decision 28: it lists each unread record in an inbox whose owner has
no live agent record (`internal/doctor/checks.go`, `checkOwnerlessInboxes`).
The two records are messages that manager-hive sent to t126-worker-1 near the
end of task 126. The worker stopped before it read them, and the task landed
through the verify-repair path, so the records carry nothing that anyone must
still act on. `hive inbox` has only the read path and `hive inbox send`. No
command marks a record read, and no command removes a record. The warning
therefore stays until the records disappear, and nothing can make them
disappear. Every finished task that ends this way adds a permanent warning.

**Reproduction steps:**
1. A manager sends a message to a worker near the end of its task.
2. The worker stops before it reads the message.
3. Run `hive doctor`. It prints the `ownerless inbox records` warning.
4. Look for a command that marks the records read or removes them. None exists.
