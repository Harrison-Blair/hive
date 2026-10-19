---
title: "Any agent can read and consume the inbox of another role"
status: "Open"
opened: 2026-10-19
decision: "Only the owner's read marks records read. The coordinator can read a manager inbox and marks nothing. Any other caller gets an error. Workers get no inbox (wave 2)."
---

**Summary:**
On 2026-10-19, verifier `t12-verifier-1` ran
`hive inbox --role manager-hive` at 12:46:30. That command marked record 66
(`verified` for 12-friction-log) as read. The supervisor then did not nudge
the manager, and the landing waited until the coordinator saw the record.
`hive inbox --role` accepts any role from any agent. The workaround is a
sentence in each brief that tells the agent to read only its own inbox.

**Reproduction steps:**
1. Spawn a verifier for a task of project `hive`.
2. In the verifier pane, run `hive inbox --role manager-hive`.
3. Check that the command prints and marks the unread manager records.
4. Check that the supervisor sends no nudge to the manager for them.
