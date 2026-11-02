---
title: "A message to a busy Codex agent waits until its turn ends"
description: "codex queue holds a hive agent send until the turn of the target ends, so a Codex agent that polls gets its reports in one late burst"
status: "Open"
opened: 2026-11-02
decision: ""
---

**Issue:** `hive agent send` delivers to Codex with `codex queue --thread
UUID --message TEXT` (`internal/lib/harness/codex.go`). Codex holds a queued
message until the current turn ends. Claude delivers a cross-session message
after the next tool call. Pi does the same with `deliverAs: "steer"`. A Codex
agent that stays in one long turn, for example because it polls
`hive agent list`, gets none of its messages until that turn ends. Then all
of the messages arrive at once and most of them are out of date.

**Summary:** On 2026-11-01 the Codex coordinator `ror-desktop-coordinator` ran
`hive agent list` 73 times and `hive logs` 40 times. From 09:15 to 09:22 it ran
`agent list` every 15 to 50 s in one turn. Its workers sent it more than 60
messages, and each send reported `ok`. On 2026-11-02 a test with Codex 0.166.0
confirmed the hold. A message sent during `sleep 40` appeared only after the
turn ended.

A test of a possible path also worked on 2026-11-02:
1. Run `codex app-server --listen unix://SOCK`.
2. Start the Codex TUI with `--remote unix://SOCK`.
3. Connect a second client to SOCK. The protocol is a WebSocket over the Unix
   socket, with JSON-RPC messages.
4. Send `initialize` and then `initialized`.
5. Send `thread/loaded/list` and then `thread/read {threadId, includeTurns:
   true}` to get the turn that has the status `inProgress`.
6. Send `turn/steer {threadId, expectedTurnId, input}`.

The message appeared during the running command, before its result. The app
server protocol is marked experimental. To use this path, hive must manage an
app server process for each Codex agent. Hive must also have a WebSocket
client and a fallback for a turn that ends before the steer arrives.

**Reproduction steps:**
1. Spawn a Codex agent with a brief that runs `sleep 40` and then `date`.
2. During the sleep, run `hive agent send NAME PING`.
3. See that the ping appears only after the agent ends its turn.
