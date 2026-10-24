---
title: "Up and Down in a wrapped chat input can land one word off"
status: "Open"
opened: 2026-10-22
decision: "Later: the user chose to defer a fix on 2026-10-22"
---

**Summary:**
Task `122-chat-status` (goal `g3-chat-status`) makes the Chat input wrap its
text at spaces. The repair at commit `55605fa1` maps the textarea line and
column onto the wrapped rows that the input shows, so the cursor always
shows the true insertion point. Up and Down still move through the wrap
points of the textarea. These points can differ by one word from the rows
on the screen. Thus Up or Down can put the cursor some columns away from
the cell directly above or below it. Typing, Left, and Right are correct.
The manager of hive reported the gap. It is not in the scope of task 122.

**Reproduction steps:**
1. Open the Chat screen and give the focus to the input.
2. Type text that wraps to three or more rows at a space, for example
   "Keep the changelog short and group the lines by screen".
3. Put the cursor in the last row, after a word.
4. Press Up.
5. The cursor can land one word to the right or left of the cell directly
   above it on the screen.
