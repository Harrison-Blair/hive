---
title: "hive agent send can split a message that the user types"
description: "agent.prompt pastes text and presses Enter into a prompt that already holds the partial text of the user"
status: "Resolved 2026-10-30"
opened: 2026-10-19
decision: "hive agent send hands the message to the native message path of the harness and never types into the pane. claude gets it as one JSON line on its cross-session inbox socket, and a draft in the input box stays unchanged (live check with claude 2.1.314). codex and pi get their native paths in the next layers of the stack; until then send fails for them with no paste fallback"
---

**Issue:** `hive agent send` delivers its text with herdr `agent.prompt`
(`internal/agent/send/send.go:76-82`). That call pastes the text and then sends
Enter (`reference/herdr/api/agent.md:371`). Hive does not check whether the
target prompt already holds typed text. A message sent to the coordinator while
the user types joins the partial text and submits it.

**Summary:** On 2026-10-19, from 02:40 to 02:45, the user pasted three messages
into the coordinator by speech-to-text. Each time, a supervisor nudge went in
after the partial text and submitted the prompt, for example "…a tree/lineage
type for⬢ hive: 3 new inbox records, run: hive inbox". The user had to send the
rest again. The supervisor is gone, but its nudge used the same `agent.prompt`
call. Nobody has seen this yet with `hive agent send`.

**Reproduction steps:**
1. Run bare `hive` to start the coordinator, and spawn one agent.
2. In the coordinator pane, type a partial message. Do not submit it.
3. Make the agent run `hive agent send coordinator TEXT`.
4. The text `⬢ NAME: TEXT` goes after the partial text, and herdr submits the prompt.
