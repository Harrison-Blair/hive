---
title: "The --agent flag of agent send names the target, but names the sender in inbox send"
status: "Open"
opened: 2026-10-21
decision: ""
---

**Summary:**
On 2026-10-21, the manager of hive sent an instruction to worker
`t2-worker-1` with `hive agent send --agent t2-worker-1 --agent manager-hive
--wait TEXT`. The manager added the second `--agent` to name itself as the
sender, as `hive inbox send` requires. In `hive agent send`, `--agent` names
the target. The second value replaced the first with no error. Hive sent the
instruction to the manager pane, and the worker did not get it. The `--wait`
call then waited on the busy manager until the tool timeout. The workaround
was to send again with one `--agent`. The manager instructions list both
commands with `--agent NAME`, with two different meanings.

**Reproduction steps:**
1. As a manager, run `hive agent send --agent WORKER --agent MANAGER TEXT`.
2. Hive sends TEXT to the manager pane, not to the worker. Hive shows no
   error for the repeated flag.
3. Compare `hive inbox send --to coordinator TEXT --agent NAME`, where
   `--agent` names the sender.
