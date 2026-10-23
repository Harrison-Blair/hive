---
title: "An accepted newline header breaks the ask feed and card title"
status: "Open"
opened: 2026-10-23
decision: ""
---

**Summary:**
Verifier `t135-verifier-2` reproduced this defect on `dev` at `981e21de`
and task `135-ask-confirm` at `935bc91b`, at 120x40 and 80x24.
`hive ask` accepts the header `Head\nTail` in JSON.
The feed entry and option-editor title split into unframed physical rows.

The new review body handles that multiline header.
The feed and editor-title defects predate the review and remain outside task 135.
No source fix for this entry has approval.

The independent report records the request, accepted record, commands, and captures for both versions.
Its commit is `235892c0c932bea3b782844657e924c364f20be7`.
Its path is `.hive/spec/tui-v2/tasks/135-ask-confirm/evidence/VERIFIED.md`.

**Reproduction steps:**
1. In an isolated scratch home, create an ask with header `Head\nTail` in its JSON.
2. Open the TUI at 120x40 or 80x24.
3. Read the feed entry and the option-editor title.
4. Observe the split text and missing frame around its continuation rows.
