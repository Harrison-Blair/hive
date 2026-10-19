# Dogfooding friction log

This file records the friction that agents meet when they use hive itself in
this repository: bugs, missing capabilities, and workarounds found in real work.
Append each new entry at the end with the template below. An entry has the
status `Open` until a fix lands, then `Resolved YYYY-MM-DD` with the fix in its
summary. A resolved entry keeps its full text as a record.

## Entry template

**Issue:** Short title

**Status:** Open

**Summary:** What happened, and how often it occurred.

**Reproduction steps:**
1. ...
2. ...

## Entries

**Issue:** Full Pi worktree, agent, and delivery bootstrap flow is missing

**Status:** Open

**Summary:** The full Pi worktree, agent, and delivery bootstrap flow still
lacks T09, T10, T12, and T18. The coordinator dispatches work by hand with
`herdr worktree` and `herdr agent` commands. This manual dispatch is the current
workaround. No runtime task or agent record simulates a missing capability.

**Reproduction steps:**
1. Run `herdr worktree --help`. It lists `list`, `create`, `open`, and `remove`.
   The hive `worktree` command does not exist.
2. Run `herdr agent --help`. It lists `start`, `prompt`, `wait`, and related
   commands. The hive `agent spawn`, `agent send`, and `agent stop` commands do
   not exist.
3. The safe probe output is in
   `.hive/tmp/coordinator/reports/later-fix-land-01-captures/herdr-help-probes.txt`.

**Issue:** Manager spawn fails with `agent_pane_busy` right after `workspace.create`

**Status:** Open

**Summary:** On 2026-10-18 the first `hive agent spawn --role manager --project
hive --harness claude --model claude-opus-5-5` in a home with no manager
workspace failed with `first tab busy: stop the agent or pass --pane`. The log
shows `workspace.create` at 22:34:09.993 and `agent.start` on the new root pane
at 22:34:10.003, ten milliseconds later. herdr answered `agent_pane_busy`:
`agent target pane wBD:p1 is not an available shell`. The shell of the new
pane was not ready. The spawn kept the workspace `wBD` and its project record,
and it marked the agent record `exited`. A second identical spawn one minute
later adopted `wBD` and started the manager. The error text names a wrong cause:
no agent held the pane. It occurred once in one attempt.

**Reproduction steps:**
1. Register a project with no manager workspace record.
2. Run `hive agent spawn --role manager --project ID --harness claude --model claude-opus-5-5`.
3. Read `.hive/logs/<date>.jsonl`: `agent.start` follows `workspace.create` within a few milliseconds and herdr returns `agent_pane_busy`.
4. Run the same spawn again: it adopts the workspace and succeeds.
