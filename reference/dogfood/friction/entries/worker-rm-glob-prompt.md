---
title: "A worker rm with a variable glob stops on a Claude Code prompt"
status: "Open"
opened: 2026-10-23
decision: ""
---

**Summary:**
Claude Code workers often run `rm -f $VAR/*.ext` to clear scratch captures.
Claude Code flags each one as a dangerous rm on a possibly empty variable and
shows a permission prompt. In an unattended pane the prompt auto-denies after
about two minutes, and hive sends a `blocked` record to the manager. It
occurred for t125-worker-1, t126-worker-1 (two times), and t131-worker-1 on
2026-10-22 and 2026-10-23. The t131 brief had a rule to write
`rm -f "${VAR:?}"/...`, and the worker did not obey it. The worker skill has
no such rule.

**Reproduction steps:**
1. As a Claude Code worker, run `B=/some/dir; rm -f $B/*.png`.
2. See the prompt "Dangerous rm operation on possibly-empty variable path".
3. Do not answer. The prompt denies the command, and hive records `blocked`.
