---
title: "Shell-ready race in worker and verifier spawns"
status: "Resolved 2026-10-19"
opened: 2026-10-19
decision: ""
---

**Summary:**
A worker or verifier spawn in a fresh worktree workspace could hit
`agent_pane_busy`. Commit `aa29b8d9` adds the shell-ready wait before the shared
`agent.start` call. This resolves the race for both roles.

**Reproduction steps:**
1. Before commit `aa29b8d9`, create a fresh worktree workspace.
2. Spawn a worker or verifier in its root pane.
3. Check the log for `agent.start` before the shell is ready.
4. Check for `agent_pane_busy`.
