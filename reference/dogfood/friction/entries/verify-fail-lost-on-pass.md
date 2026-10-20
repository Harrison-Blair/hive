---
title: "A failed verification is lost when the next verification passes"
status: "Open"
opened: 2026-10-20
decision: ""
---

**Summary:**
`verification_history` gets an entry only from `hive task reopen`. After a
verify fail, the manager edits the brief and spawns a repair worker without a
reopen, because the task goes back to `assigned`. The next `hive task verify`
replaces `verification` with the pass, and the fail is gone from the record.
On 2026-10-20 task 79-tui-shell failed V1 with t79-verifier-1 and passed with
t79-verifier-2. Its record shows only the pass, and the new `VERIFIED.md`
replaced the old one. The Tasks screen (task 83) cannot show the repair
history that its design shows.

**Reproduction steps:**
1. Run `hive task verify` with a fail on a completed task.
2. Complete the task again and run `hive task verify` with a pass.
3. Run `hive task show --json`: `verification_history` is empty.
