---
title: "An agent can send a user chat message through the TUI"
status: "Open"
opened: 2026-10-20
decision: ""
---

**Summary:**
On 2026-10-20 at 14:4x UTC, worker t89-worker-1 (task 89-screen-board-audit)
ran the TUI in its own pane to audit the screens. It pressed `i` on the Chat
screen and typed screen keys. The TUI sent "278" with `hive chat send`, and the
coordinator got `⬢ user: 278` (chat record c1). A message with `⬢ user:`
carries the authority of the user. The coordinator treated it as a message from
the user and replied (c2).

`hive chat send` refuses a caller that passes `--agent`. The TUI does not pass
it, and nothing checks where the TUI runs. Thus each agent that can run `hive`
can make a message with the authority of the user. Only the report from the
manager showed the true source.

**Reproduction steps:**
1. In the pane of a worker, run bare `hive`.
2. On screen 1, press `i`, type text, and press Enter.
3. The coordinator pane gets `⬢ user: TEXT`, and `hive chat log` shows a user
   record.

**Fix and remaining gap:**
Task 94-chat-user-only makes `chat.Send` refuse a caller with `--agent`, a
caller in the pane of a live agent, a caller with a working directory below
`projects/`, and a caller in the workspace of a live manager, worker, or
verifier. The TUI passes its pane and working directory. A gap stays open: an
agent can open a tab in the hive session workspace with a working directory
outside `projects/`, and that tab can still send a user chat message.
