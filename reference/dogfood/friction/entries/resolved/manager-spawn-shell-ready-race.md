---
title: "Manager spawn fails with `agent_pane_busy` right after `workspace.create`"
status: "Resolved 2026-10-19"
opened: 2026-10-18
decision: ""
---

**Summary:**
On 2026-10-18 the first `hive agent spawn --role manager --project
hive --harness claude --model claude-opus-5-5` in a home with no manager
workspace failed with `first tab busy: stop the agent or pass --pane`. The log
shows `workspace.create` at 22:34:09.993 and `agent.start` on the new root pane
at 22:34:10.003, ten milliseconds later. herdr answered `agent_pane_busy`:
`agent target pane wBD:p1 is not an available shell`. The shell of the new
pane was not ready. The spawn kept the workspace `wBD` and its project record,
and it marked the agent record `exited`. A second identical spawn one minute
later adopted `wBD` and started the manager. The error text names a wrong cause:
no agent held the pane. It occurred on each first spawn of a project manager:
twice in two attempts on 2026-10-18, for `hive` and for `stenographer`. The
retry succeeded each time. During rc-file startup, the shell runs commands in
its own process group. Thus, a group ID equal to the shell PID is not sufficient.
Commit `aa29b8d9` also requires one foreground process whose PID is the shell
PID.

**Reproduction steps:**
1. Register a project with no manager workspace record.
2. Run `hive agent spawn --role manager --project ID --harness claude --model claude-opus-5-5`.
3. Read `.hive/logs/<date>.jsonl`: `agent.start` follows `workspace.create` within a few milliseconds and herdr returns `agent_pane_busy`.
4. Run the same spawn again: it adopts the workspace and succeeds.
