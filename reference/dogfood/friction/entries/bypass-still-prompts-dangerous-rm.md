---
title: "Claude Code in bypass mode still asks before a dangerous rm"
description: "A Claude agent that runs rm on a variable path stops at a dialog despite --dangerously-skip-permissions"
status: "Open"
opened: 2026-10-19
decision: ""
---

**Issue:** Hive starts Claude Code with `--dangerously-skip-permissions`
(`internal/lib/harness/claude.go:11`). Claude Code still shows the dialog
"Dangerous rm operation on possibly-empty variable path" for commands such as
`rm -rf $S/*`. The dialog denies the command by itself after about two minutes.
While the dialog is open, herdr reports the agent `blocked`, and
`hive agent send` refuses it (`internal/agent/send/send.go:83-84`).

**Summary:** From 2026-10-19 to 2026-10-27, Claude agents hit this dialog more
than 15 times, in hive and stenographer tasks. The commands were `rm -rf $S/*`,
`rm -f $S/*.txt`, `rm -f -- "$S"/*.txt`, and `rm -rf $S/$r-wt`. The dialog also
held commits and pushes that agents chained in the same command. Rules in briefs
did not stop the pattern. A manager declined the dialog with
`herdr pane send-keys PANE esc` and then `enter`. The key name `Return` did not
register. Commands with literal paths or `"${VAR:?}"` did not show the dialog.

**Reproduction steps:**
1. Run `hive agent spawn --name N --harness claude --model M --prompt TEXT`.
2. Let the agent run `rm -rf $VAR/*` with a shell variable.
3. Claude Code shows the dangerous rm dialog, and `hive agent list` shows the agent `blocked`.
4. `hive agent send N TEXT` fails with "agent N is blocked by a dialog".
