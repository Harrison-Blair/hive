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

**Update 2026-11-02:** Two coordinators lost their record and sent as `user`
for an evening. At that time bare `hive` gave each coordinator the one name
`coordinator`. On 2026-10-30 at 22:37:56, bare `hive` in the hearth pane
`w167:p1` saved the record `coordinator` for its pane, so the ror-desktop
coordinator pane had no record. At 23:20:14, bare `hive` in the hive pane
`w16A:p1` saved it again, so the hearth coordinator pane had no record. From
22:38 to 2026-10-31 00:18, the logs show 33 events with `agent_id` `user` from
these two coordinators: 17 `agent prompted` and 16 `agent messaged`. Their
workers and reviewers got the messages as `⬢ user: …`. 12 of the messages
start `⬢ user: Coordinator: …`, so the text names a coordinator while the
marker gives user authority.

A second form: at 23:41:00 and 23:44:02 the hive coordinator relayed words of
the user to `claude-artifact-check` as `⬢ coordinator: ⬢ user: …`. The inner
`⬢ user:` marker is text that the coordinator typed, inside an agent message.
A reader that looks for the marker without its position can take it as user
authority.

Since then, each coordinator gets a name per project (`<dir>-coordinator`),
and bare `hive` takes the next free name when a live pane holds the record.
So one coordinator no longer takes the record of another. A pane with no
record, or an agent with `HERDR_PANE_ID` unset, still sends as `user`.
