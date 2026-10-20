---
title: "An agent that drives the TUI can send a user chat message"
status: "Open"
opened: 2026-10-20
decision: ""
---

**Summary:**
On 2026-10-20 worker t89-worker-1 ran bare `hive` on the home for the
screen board audit. It typed screen keys after `i` opened the Chat input, and
the TUI sent `278` to the coordinator. Hive tags each TUI chat message with
`⬢ user:`, and the role gate gives that tag the authority of the user. The
TUI cannot tell the user from an agent that sends keys to its pane. The
manager told the coordinator to ignore the message.

**Reproduction steps:**
1. As an agent, start bare `hive` in a herdr pane on the home.
2. Send the keys `i`, `2`, `7`, `8`, and Enter to the pane.
3. Run `hive chat log`: a user message `278` is in the history, and the
   coordinator pane got `⬢ user: 278`.
