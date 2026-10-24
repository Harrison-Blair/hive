---
title: "The user cannot answer a structured question in the Chat screen"
status: "Resolved 2026-10-23"
opened: 2026-10-22
decision: "Goal g6 task 131 adds ask cards in the Chat dock: picks, notes, previews, dismiss, and answers routed to the asker"
---

**Summary:**
On 2026-10-22 the user reported that the Chat screen has no way to answer a
question from the coordinator. The coordinator asks each design question
through its harness question tool, in the coordinator pane. The user must go to
that pane to answer. Hive has no command that the coordinator can use to send a
question with options to the Chat screen, and the Chat screen has no control to
select an option. The user can only type free text in the message box.

**Reproduction steps:**
1. The coordinator has a question with options for the user.
2. The user is on screen 2 Chat in the TUI.
3. The coordinator can only send text with `hive chat reply`. The Chat screen
   shows no options to select, so the user must answer in the coordinator pane.
