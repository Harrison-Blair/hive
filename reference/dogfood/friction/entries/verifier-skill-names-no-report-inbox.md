---
title: "The verifier skill names no inbox for a report"
status: "Open"
opened: 2026-10-24
decision: "Goal g13 replaces the skills with profiles. The verifier profile must name the inbox of the manager."
---

**Summary:**
On 2026-10-24 verifier t140-verifier-3 of task 140-spec-local-state sent its
verdict text and a question for the manager to the coordinator inbox (record
503). The manager got only the `verified` record of `hive task verify`. The
coordinator relayed the message. `.agents/skills/hive-verifier/SKILL.md` names
no inbox for a report or a question. Each verifier selects an inbox by itself.

**Reproduction steps:**
1. Read `.agents/skills/hive-verifier/SKILL.md`.
2. Search it for `inbox send` or the manager inbox. No step names one.
3. Spawn a verifier. Its report can go to the coordinator in place of the manager.
