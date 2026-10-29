---
title: "A restart can silently run a stale hive binary"
description: "Bare hive and agents run the hive on PATH, and nothing compares it with the home checkout"
status: "Open"
opened: 2026-10-23
decision: ""
---

**Issue:** Bare `hive` and each agent run the `hive` binary on PATH. No check
compares that binary with the commit of the home checkout. Only `hive version`
prints the commit, and nothing else reads it. After new work merges, a restart
without a reinstall runs the old code and reports nothing.

**Summary:** On 2026-10-23 the user restarted hive to load the new binary after
goal g6 landed. The restart ran the old binary, commit f7990406. No step had
installed the new one. No check reported the gap. The coordinator installed the
binary with `go install .` and told the user to restart again.

**Reproduction steps:**
1. Merge new work to `main` in the home. Do not run `go install .`.
2. Run `hive version` and compare its commit with `git -C HOME rev-parse --short=12 HEAD`. They differ.
3. Run bare `hive`. It runs the old binary and gives no warning.
