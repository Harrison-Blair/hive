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
