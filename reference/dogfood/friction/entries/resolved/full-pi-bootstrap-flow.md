---
title: "Full Pi worktree, agent, and delivery bootstrap flow is missing"
status: "Resolved 2026-10-19"
opened: 2026-10-19
decision: "Stale: `hive worktree` and `hive agent spawn/send/stop` exist."
---

**Summary:**
The full Pi worktree, agent, and delivery bootstrap flow still
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
