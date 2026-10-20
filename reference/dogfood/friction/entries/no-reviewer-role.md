---
title: "Hive cannot spawn a reviewer or another one-off agent"
status: "Open"
opened: 2026-10-19
decision: ""
---

**Summary:**
On 2026-10-19 the coordinator needed a plan review of the hive TUI goal by
GPT 6.1 Sol before the goal went to the manager. `hive agent spawn` accepts
only the roles coordinator, supervisor, manager, worker, and verifier
(`internal/agent/register.go:162`). A worker or a verifier needs a task. A
plan review has no task. `hive agent spawn --role reviewer` fails with
"invalid role: reviewer".

The coordinator ran `pi -p` in a background shell instead. That run waited
for 39 minutes with no output, because pi showed its "Trust project
folder?" prompt and no person could see it. The coordinator then started pi
by hand in a new herdr tab with `herdr pane run`. Hive has no record of that
agent, so `hive agent list` does not show it, and the supervisor does not
watch it. The agent also stopped at the same trust prompt, which the user
must answer.

**Reproduction steps:**
1. Run `hive agent spawn --role reviewer --project hive --pane PANE --harness pi --model MODEL`.
2. The command fails with "invalid role: reviewer".
3. Run `pi -p --model MODEL "PROMPT"` in the home with no TTY.
4. The process waits with no output and no CPU use, at the trust prompt.
