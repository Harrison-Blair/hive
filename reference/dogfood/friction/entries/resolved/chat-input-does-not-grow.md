---
title: "The Chat message box does not grow or scroll while the user types"
status: "Resolved 2026-10-23"
opened: 2026-10-22
decision: "Goal g3 task 122 lands the growing input: the box grows with the typed text and the feed shrinks"
---

**Summary:**
On 2026-10-22 the user reported that the Chat message box does not get bigger
and does not scroll while they type a long message. The box stays one line
high, so the user cannot see all of the text that they typed.

**Reproduction steps:**
1. Run bare `hive` (dev 6f14e3bd) and open screen 2 Chat.
2. Press `i` and type a message that is longer than the width of the box.
3. The box stays one line high, and the text does not scroll into view.
