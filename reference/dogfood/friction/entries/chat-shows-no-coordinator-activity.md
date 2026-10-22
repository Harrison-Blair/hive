---
title: "The Chat screen does not show that the coordinator is working"
status: "Open"
opened: 2026-10-22
decision: ""
---

**Summary:**
On 2026-10-22 the user reported that the Chat screen does not show when the
coordinator is working on a reply. The header shows `coordinator ● working`,
but the chat feed shows nothing until the coordinator saves a reply with
`hive chat reply`. The user sent `hello?` (chat record c29) because the feed
showed no activity. The user wants an activity row in the feed, at the
position of the next coordinator reply, while the coordinator works. The row
is similar to a "typing" indicator in a chat program.

**Reproduction steps:**
1. Run bare `hive` and open screen 2 Chat.
2. Send a message that the coordinator must do work to answer.
3. While the coordinator works, the feed shows no row for the coordinator.
   Only the header status changes.
