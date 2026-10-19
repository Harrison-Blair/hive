---
title: "The verifier skill does not give the path of VERIFIED.md"
status: "Open"
opened: 2026-10-19
decision: ""
---

**Summary:**
On 2026-10-19, verifier t36-verifier-1 (Claude Opus 5.5) of
36-nudge-no-home wrote `VERIFIED.md` at the root of the checkout. All earlier
tasks put it in `evidence/<task-id>/VERIFIED.md`. The skill
`.agents/skills/hive-verifier/SKILL.md` names `VERIFIED.md` but gives no path.
`hive task verify` accepted the root file. A root `VERIFIED.md` on `dev` causes
a conflict with the next verifier that does the same. The workaround was
`hive task reopen`, a repair worker to move the file, and a fresh verifier.

**Reproduction steps:**
1. Spawn a verifier for a completed task.
2. The verifier reads the skill and writes `VERIFIED.md` at the checkout root.
3. Run `hive task verify`. It records a pass with no path check.
