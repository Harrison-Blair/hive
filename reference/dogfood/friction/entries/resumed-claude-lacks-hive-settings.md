---
title: "A Claude agent that herdr resumes can lack the hive start arguments"
description: "After a herdr restart, herdr resumes Claude as claude --resume ID, probably without --settings crossSessionInbound, so hive agent send to it can fail"
status: "Open"
opened: 2026-11-02
decision: ""
---

**Issue:** Hive starts each Claude agent with `--settings
'{"crossSessionInbound":"accept"}'`, the guide, and the trust env. `hive
agent send` checks the argv of the claude process for that setting. After a
server restart, herdr resumes a Claude session with the default
`resume_agents_on_restore = true`. `reference/herdr/RESYNC.md` records the
command `claude --resume <id>` for a session with no `resume_argv`. If the
resumed process lacks the hive arguments, the hive record still matches the
pane, but `send` fails. The agent can also lack the hive guide and the model
of the record.

**Summary:** Not observed with a hive agent. This entry comes from the herdr
reference on 2026-11-02, during a review of the entry
`gone-pane-records-stay`. Unverified: whether a hive agent reports a
`resume_argv` to herdr, and what argv the resumed process has.

**Reproduction steps:**
1. On a scratch herdr server, run `hive agent spawn --harness claude` with a
   scratch `HIVE_HOME`.
2. Stop the server and start the same session again.
3. Run `herdr pane process-info --pane PANE` and read the argv of `claude`.
4. Run `hive agent send NAME TEXT`, and check the outcome.
