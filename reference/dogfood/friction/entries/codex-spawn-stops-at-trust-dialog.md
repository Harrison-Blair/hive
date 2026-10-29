---
title: "A Codex spawn in an untrusted Git directory stops at the trust dialog"
description: "Codex shows its folder trust dialog in a Git repository with no saved trust, and hive has no per-process trust method"
status: "Open"
opened: 2026-10-25
decision: ""
---

**Issue:** Codex shows its folder trust dialog in a Git repository that has no
saved trust in `~/.codex/config.toml`, also with
`--dangerously-bypass-approvals-and-sandbox`. Hive has no per-process trust
method for Codex (`internal/lib/harness/codex.go:51-54`). Spawn starts Codex,
sees `blocked`, returns outcome `partial`, and sends no prompt
(`internal/agent/spawn/spawn.go:170-171`).

**Summary:** During the verification of task 181, about 2026-10-26 UTC, a
verifier tried to start three optional Codex helpers in its task worktree. The
old hive refused each spawn with the saved-trust message. The verifier changed
no trust setting and used another runner. The current hive does not refuse. It
starts Codex and reports the dialog. This newer behavior comes from the code. No
live check tested it.

**Reproduction steps:**
1. Use a Git working directory with no saved Codex trust.
2. In a herdr pane in that directory, run `hive agent spawn --name N --harness codex --model M --prompt TEXT --json`.
3. Observe outcome `partial` and the error "agent N is blocked by a dialog".
