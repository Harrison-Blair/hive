---
title: "An accepted multiline label breaks ask-option editor borders"
status: "Open"
opened: 2026-10-23
decision: ""
---

**Summary:**
Verifier `t135-verifier-2` reproduced this defect on `dev` at `981e21de`
and task `135-ask-confirm` at `935bc91b`, at 120x40 and 80x24.
`hive ask` accepts an option label with a line feed.
Before review, the option editor draws that label as multiple physical rows.
The first row lacks its right border. The continuation lacks its left border.

The review wrapper fix handles the same label correctly after the pick.
The option editor defect predates that fix and remains outside task 135.
No source fix for this entry has approval.

The independent report records the request, accepted record, commands, and captures for both versions.
Its commit is `235892c0c932bea3b782844657e924c364f20be7`.
Its path is `.hive/spec/tui-v2/tasks/135-ask-confirm/evidence/VERIFIED.md`.

**Reproduction steps:**
1. In an isolated scratch home, create an ask with option label `First line\nSecond line` in its JSON.
2. Open the ask in the TUI at 120x40 or 80x24.
3. Inspect the option editor before selecting the option.
4. Observe the missing side borders on the two physical label rows.
