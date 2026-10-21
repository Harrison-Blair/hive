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

**Takeover:**
The pi reviewer in the home root ran coordinator startup two times. The
first prompt was "Read FILE and do what it says". The second prompt also
said "you are a plan reviewer, not the coordinator". Each time it ran
`hive agent register --role coordinator`, which marked the live coordinator
stale, took the coordinator claim, and renamed the review tab to
`coordinator`. Nudges then went to the reviewer pane. It also read the
coordinator inbox and ran `hive doctor`. The real coordinator ran
`hive agent register --role coordinator` again to take back the claim. A
pi agent in a scratch directory outside the home, with the tools read,
grep, find, ls, and write, did not load `AGENTS.md` or the hive skills and
did the review only.

5. Start pi in the home root with a prompt that names another role.
6. pi runs `hive agent register --role coordinator` and takes the claim.

**Second occurrence (2026-10-21):** the user asked for a review of the
`tui-v2` spec draft by GPT 6.1 Sol in a pane that the user can see. Hive
still cannot spawn it: `hive agent spawn --role reviewer --pane w58:p7`
fails with "invalid role: reviewer" (`internal/agent/register.go:179`).
`hive agent spawn` has no free-text prompt flag. The coordinator made tab
`w58:t7` with `herdr tab create` and started pi there with
`herdr pane run`, with the prompt in a file. The tools were read, grep,
find, and ls only. With no bash tool, pi cannot run
`hive agent register`, so the coordinator claim stayed safe. Hive has no
record of this agent.

The reviewer finished, but it had no way to report back. Hive had no record
of it, so the supervisor sent no nudge, and the agent had no inbox and no
`hive` command. The coordinator did not know that the review was complete.
The user had to tell the coordinator. The coordinator then read the result
from the pane with `herdr pane read`.
