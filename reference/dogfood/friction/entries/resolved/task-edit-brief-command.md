---
title: "No `hive task edit --brief` command"
status: "Resolved 2026-10-19"
opened: 2026-10-18
decision: "Fixed in 22-task-edit-brief (77cdf44d): hive task edit --task ID --brief FILE, only for created or assigned, only by the controlling role."
---

**Summary:**
The manager changed a task brief by hand in
`.hive/state/briefs/` on 2026-10-18. hive has no `hive task edit --brief`
command. The manager must edit the stored brief by hand.

**Reproduction steps:**
1. Run `hive task --help`.
2. Check that the command list has no `edit` command.
3. Edit a task brief in `.hive/state/briefs/` by hand.
