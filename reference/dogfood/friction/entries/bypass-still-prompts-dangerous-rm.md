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

**Recurrence 2026-10-20 (stenographer 108, 110):** t108-verifier-3 hit the
dialog on `rm -rf $S/*` in its final commit step, and t110-verifier-2 on
`rm -rf $S/$rev` before the manager's prompt arrived. The manager declined
with `herdr pane send-keys <pane> Down`, then `Return` (`Enter` did not
register). `hive agent send` refuses while the pane is blocked
(`agent_blocked`).

**Possible recurrence 2026-10-21 (hive 19):** the manager got a `blocked`
record with no reason for t19-worker-1 during its cleanup step. When the
manager read the pane, it showed no dialog. The worker was in
`hive task complete`, and the `completed` record came a short time later.
The cause is not known. The record can be the same dialog with an
automatic deny, or a short false `blocked` state. The blocked record still
gives no reason.

**Recurrence 2026-10-21 (hive 29):** t29-worker-1 chained the evidence
commit and push with `rm -rf $S/*` in one command. The dialog held the commit
and the push also. The manager declined it with `herdr pane send-keys PANE
esc`, which moved the choice to `No`, then `herdr pane send-keys PANE enter`.
The key name `Return` did not register. The manager told the worker to commit
first and to clean up with `"${S:?}"/*`.

2026-10-21, task 40-screen-logs: the dialog came again, although the brief
Rules now forbid `rm` on a bare variable. The worker ran `rm -f $T/*.ansi` in
one command with a commit and a live capture. The task went to `blocked`. The
manager declined it the same way. A rule in a brief does not stop the error.
A rule in the worker and verifier skills, or a harness hook, is still open.

**Recurrence 2026-10-22 and 2026-10-23 (hive 125, 126, 131):** t125-worker-1,
t126-worker-1 (`rm -f $C/*.ansi` and `rm -f $H/*.png`), and t131-worker-1
(`rm -f $B/*.png`) each hit the dialog in a capture cleanup. Each gave a
`blocked` record with no reason, and each prompt denied itself after about two
minutes. The t131 brief had the `"${VAR:?}"` rule, and the worker did not
obey it. The fix in the worker and verifier skills, or a harness hook, is
still open.

**Recurrence 2026-10-24 (hive g12 wave C):** workers t146-worker-1
(`rm $S/*.events.js`) and t145-worker-1 (`rm -f $B/*.png`) hit the same
dialog during board and capture steps. The supervisor wrote a `blocked`
record with no reason, and the manager read each pane with
`herdr pane read` to find the cause. The manager sent each worker an inbox
record to rewrite the command with `"${VAR:?}"` and let the dialog deny
itself. Workers spawned by hive still repeat the pattern. A third hit came the same day: t144-worker-1 on `rm -rf $SP/*` in its scratchpad cleanup.
