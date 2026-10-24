---
title: "A restart can silently run a stale hive binary"
status: "Open"
opened: 2026-10-23
decision: ""
---

**Summary:**
On 2026-10-23 the user restarted hive to load the new binary after goal g6
landed. The restart ran the old binary. No step installs the binary: the
coordinator's instruction named `scripts/reinstall.sh`, a script of the
stenographer project that does not exist in hive. The TUI, the supervisor, and
the coordinator started on commit f7990406, which predates all of g6 (tasks
126, 129, 131). No check reported the gap. `hive doctor` compares the binary
with releases, not with dev, and the TUI old-binary mark covers only the
supervisor against the CLI binary. The coordinator installed the new binary
with `go install .` and told the user to restart once more.

**Reproduction steps:**
1. Land new work on dev. Do not run `go install .`.
2. Quit the TUI and run `hive` again.
3. The new TUI, supervisor, and coordinator run the old binary. No check
   reports it.
