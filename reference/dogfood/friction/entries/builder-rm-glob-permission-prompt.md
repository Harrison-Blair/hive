---
title: "Claude Code builders stop at a permission prompt for rm on a variable glob"
status: "Open"
opened: 2026-10-27
decision: ""
---

**Issue:** a builder that cleans scratch files with `rm -f $S/*.txt` hits the
Claude Code check "Dangerous rm operation on possibly-empty variable path".
The pane waits about two minutes and then denies the command. Hive reports the
builder as `blocked`.

**Status:** Open.

**Summary:**
On project stenographer, goal g3-hub-redesign, builders of tasks 193, 196
and 197 each stopped at this prompt during evidence cleanup. Each time the
prompt timed out, the command was denied, and the builder went on. The
planner receives a `blocked` record and must read the pane to see that no
action is needed. The builder profile could tell builders to use
`"${S:?}"/*.txt` or literal paths.

**Reproduction steps:**
1. Spawn a builder with harness `claude`.
2. Let it run a cleanup command `rm -f $S/*.txt` with `S` set in the same
   command.
3. Observe the permission prompt and the `blocked` inbox record.
