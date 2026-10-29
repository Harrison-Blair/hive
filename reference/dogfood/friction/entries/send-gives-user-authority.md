---
title: "An agent can send a message with user authority"
description: "hive agent send labels the sender user when the caller pane has no record or HERDR_PANE_ID is empty"
status: "Open"
opened: 2026-10-20
decision: ""
---

**Issue:** `hive agent send` delivers `⬢ SENDER: TEXT`. SENDER is the record
name of the caller pane. When `HERDR_PANE_ID` is empty, or no record matches the
pane, SENDER is `user` (`internal/agent/send/send.go:62-68`,
`internal/agent/caller.go:12-24`). AGENTS.md gives a `⬢ user:` message the
authority of the user. So an agent that runs `hive agent send` from a pane with
no record, or with `HERDR_PANE_ID` unset, sends a message with user authority.

**Summary:** On 2026-10-20, worker t89-worker-1 ran the old TUI in its own pane
and typed keys on the Chat screen. The coordinator got `⬢ user: 278` and treated
it as a message from the user. Only the report of the manager showed the true
source. Chat was later blocked and then removed. The current `hive agent send`
has the same class of gap.

**Reproduction steps:**
1. In a herdr pane, start the coordinator with bare `hive`.
2. From another pane that has no hive record, run `hive agent send coordinator TEXT`.
3. Or from an agent pane, run `env -u HERDR_PANE_ID hive agent send coordinator TEXT`.
4. The coordinator pane gets `⬢ user: TEXT`.
