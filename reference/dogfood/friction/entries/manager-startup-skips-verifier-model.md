---
title: "The manager startup list does not name overrides/verifier-model.md"
status: "Open"
opened: 2026-10-24
decision: ""
---

**Summary:**
The Startup section of the hive-manager skill names three files: the project
`AGENTS.md`, `overrides/presets.md`, and the project `grants.md`. It does not
name `overrides/verifier-model.md`. The presets name only worker choices. On
2026-10-23 the qmeter manager read the presets for the verifier too and spawned
the task 138 verifier with `claude` and `claude-opus-5-5`. The verifier model
file requires `pi` with `openai-codex/gpt-6.1-sol:xhigh`. The coordinator
accepted the result without a re-verify. It happened once.

**Reproduction steps:**
1. Start a manager and follow the Startup section of
   `.agents/skills/hive-manager/SKILL.md`.
2. Spawn a verifier with the first choice of the task tier in
   `overrides/presets.md`.
3. Compare the model with `overrides/verifier-model.md`. They differ.
