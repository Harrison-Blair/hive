---
title: "No `hive task edit --brief` command"
status: "Open"
opened: 2026-10-18
decision: "Add `hive task edit --brief`, allowed only before `completed` (wave 4)."
---

**Summary:**
The manager changed a task brief by hand in
`.hive/state/briefs/` on 2026-10-18. hive has no `hive task edit --brief`
command. The manager must edit the stored brief by hand.

**Reproduction steps:**
1. Run `hive task --help`.
2. Check that the command list has no `edit` command.
3. Edit a task brief in `.hive/state/briefs/` by hand.
