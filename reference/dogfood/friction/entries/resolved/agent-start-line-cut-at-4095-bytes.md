---
title: "herdr agent.start can cut a long command line at 4095 bytes"
description: "A start command line above 4095 bytes can reach the new pane shell cut short, so the harness never starts"
status: "Resolved 2026-10-29"
opened: 2026-10-29
decision: "Hive refuses a spawn whose start command line is above 4000 bytes (5e4489ab), so the line never reaches the 4095-byte limit."
---

**Issue:** herdr `agent.start` types the whole harness command line into the
pane shell, with each argument in single quotes. When the line is longer than
4095 bytes, the shell can get only the first 4095 bytes. 4095 bytes is the line
limit of a Linux tty in canonical mode, so the probable cause is that herdr
types the line before the new shell has switched its line editor on. The cut
line has no closing quote, so the shell waits at a `>` continuation prompt.
The harness does not start, and `hive agent spawn` returns `partial` with
`agent_not_found` in phase `wait`. reference/herdr does not give a length
limit for `agent.start`.

**Summary:** On 2026-10-29 the agent `builder-guide` checked the branch
`builder-guide` (PR #61). That branch adds the hive guide (about 4.6 KB) to
`--append-system-prompt`, which makes a claude command line of about 4.7 KB.
1 of 4 claude spawns in a new tab failed as above: the pane showed the typed
line up to "…send returns when herdr has deliver" and then `>`. The cut was at
byte 4095 of the line. The 3 other claude spawns, 1 codex spawn (`-c
developer_instructions=…`, also above 4095 bytes), and 5 direct `herdr agent
start` runs (6.4 KB and 4.7 KB lines, new tabs, started at once or 3 s after
`tab create`) all started. The failure is intermittent.

**Reproduction steps:**
1. Build hive from the branch `builder-guide` and make a scratch home: a
   directory with `.hive/README.md`.
2. From a herdr pane in that directory, run `HIVE_HOME=$PWD hive agent spawn
   --name t1 --harness claude --model claude-haiku-5-5 --json`. Repeat with new
   names.
3. When the outcome is `partial` with `agent_not_found`, run `herdr pane read
   PANE` on the pane of the result: the line ends at byte 4095 and the shell
   shows `>`.

**Mitigation:** PR #61 keeps the full start line of each harness below a
4000-byte budget. `cmd/guide_budget_test.go` fails when the line of claude,
pi, or codex with the guide and a long model name reaches 4000 bytes. On
2026-10-29, with the dense guide (about 3.56 KB per line), 3 of 3 new claude
spawns and 1 of 1 codex spawn started.

**Next steps:**
1. Follow-up in hive: before `agent.start`, wait until the shell of the new
   pane is at its interactive prompt with its line editor on. Then a line above
   4095 bytes is safe, and the budget is only a guard.
2. Report the cut to herdr upstream: `agent.start` can type a line longer than
   the tty canonical-mode limit before the shell is ready, and herdr reports no
   error.
