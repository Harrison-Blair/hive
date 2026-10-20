---
title: "Claude Code in bypass mode still asks before a dangerous rm"
status: "Open"
opened: 2026-10-19
decision: ""
---

**Summary:**
On 2026-10-19, verifier t44-verifier-1 (Claude Code, `--dangerously-skip-permissions`)
ran a command with `rm -rf $S/*`. Claude Code showed a dialog "Dangerous rm
operation on possibly-empty variable path" and waited, with an automatic deny
after about 2 minutes. The supervisor wrote a `blocked` record to the manager.
The manager denied the dialog. The verifier ran the step again without the rm
and recorded the verdict. The bypass flag does not stop this safety check.

**Reproduction steps:**
1. Spawn a Claude Code verifier with hive (bypass flag on).
2. The agent runs `rm -rf $VAR/*` with a shell variable.
3. Claude Code shows the dangerous rm dialog. herdr reports `blocked`.

**Recurrence 2026-10-19 (qmeter):** verifier t52-verifier-1 hit the same
dialog on `rm -rf $S/*` in its scratchpad cleanup, chained after
`hive task verify` in one command. herdr reported `blocked`; the blocked
inbox record had no reason, so the manager had to read the pane with
`herdr pane read` to find it. The manager denied it and the verifier reran
`hive task verify` alone.

**Occurrence 2026-10-20:**
Verifier t81-verifier-2 ran `rm -f -- "$S"/*.txt` on its scratchpad in a
cleanup step, although its note said to avoid `rm` on an unquoted variable.
The same dialog showed. The manager denied it with `herdr pane send-keys
<pane> Escape`. A note in the verifier prompt does not stop the pattern.
