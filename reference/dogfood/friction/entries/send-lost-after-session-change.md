---
title: "A message to a Claude coordinator can be lost with outcome ok"
description: "hive agent send writes to the inbox socket and reports ok, but after the Claude session changed in the same process no message appeared"
status: "Open"
opened: 2026-11-01
decision: ""
---

**Issue:** `hive agent send` reports `outcome ok` when it writes the line to
the inbox socket of the Claude Code process (`internal/lib/harness/claude.go:54`,
`Deliver`). Hive does not check that the message appears in the session. Four
reports from agents to a Claude coordinator never appeared, and hive logged
each send as ok. The workers ended their turns, and the coordinator waited
with no signal. `hive logs` cuts long bodies (`truncated=true`), so the log
did not hold the full lost text either.

**Summary:** On 2026-11-01 the coordinator `ror-desktop-coordinator` (pane
`wBZ:p7`, Claude Code 2.1.316, PID 165427, socket
`$XDG_RUNTIME_DIR/cc-socks/165427.sock`) orchestrated agents in ror-desktop.
- The last message that appeared was from `review-d6` at 17:30:34.
- At about 17:31 the Claude session in that same process changed. The session
  scratchpad moved from the directory of session `8de644eb-…` to that of
  `f7dfdfea-…`, and Claude Code said the old scratchpad was no longer
  available. Two new sockets and a `claude daemon run --origin transient`
  process appeared at 17:31. The herdr pane record still showed
  `agent_session` `8de644eb-…`.
- After that, these sends logged `agent messaged` and `outcome=ok`, and none
  appeared in the coordinator: `canvas-sync` at 17:32:47, `sheet-star-preview`
  at 17:50:36 and 17:54:32 (the second with `--wait`), `review-f1` at
  17:56:18, and `canvas-sync` at 17:56:36.
- The user saw the work stop twice and had to ask. The coordinator found the
  reports with `hive logs --agent NAME` and in report files the agents wrote.

The cause is not verified. The session change in the same PID is the only
known event at the start of the loss. Claude Code 2.1.316 is newer than the
2.1.314 that the delivery code was checked against.

Workaround: tell each agent to write its full report to a file and send only
the path. The coordinator watches for the file with a background loop, and
does not depend on the message.

Possible fixes, not decided:
- After the write, hive checks that the session took the message, and
  reports `partial` or `failed` when it can't tell.
- `hive agent send` keeps a copy of each message, so a lost message can be
  read again in full.
- `hive logs` keeps the full body of `agent messaged` events.

**Reproduction steps:**
1. Run bare `hive` with the claude harness to start the coordinator, and
   spawn one agent.
2. Send a message from the agent to the coordinator. It appears.
3. Make the coordinator's Claude session change inside the same process.
   The 2026-11-01 run did this at a context compaction and session
   continuation, but the exact trigger is not confirmed.
4. Send another message from the agent with `hive agent send`. The send gives
   `outcome ok`, and the message does not appear in the coordinator.
