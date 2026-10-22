---
title: "A supervisor nudge splits a message that the user types"
status: "Open"
opened: 2026-10-19
decision: ""
---

**Summary:**
On 2026-10-19, between 02:40 and 02:45, the user typed (by speech-to-text
paste) 3 messages to the coordinator. The supervisor sent a nudge to the
coordinator pane while each message was in the prompt. The nudge text
"⬢ hive: N new inbox record(s), run: hive inbox" went into the prompt after
the partial text, and the nudge submitted the prompt. The coordinator got a
cut message, for example "For agents working, I am interested in a
tree/lineage type for⬢ hive: 3 new inbox records, run: hive inbox". The user
had to send the rest again. This occurred 3 times in about 5 minutes, while
the managers sent many records.

The supervisor does not check if the prompt of the pane has text before it
sends the nudge. A nudge for the coordinator is the most likely to collide,
because the user types in that pane.

**Reproduction steps:**
1. Start `hive supervise`.
2. In the coordinator pane, type a partial message. Do not submit it.
3. Make a manager send a record to the coordinator inbox.
4. The nudge text goes after the partial text, and the prompt is submitted.
