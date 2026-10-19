---
title: "A verifier spawn with `--pane` requires `--project`"
status: "Open"
opened: 2026-10-19
decision: "Hive derives the project from the task, and the docs are fixed (wave 2)."
---

**Summary:**
The manager skill shows a verifier spawn without `--project`. A
verifier spawn with `--pane` requires `--project`.

**Reproduction steps:**
1. Read the verifier spawn command in `hive-manager/SKILL.md`.
2. Add `--pane PANE`. Do not add `--project`.
3. Run the command.
4. Check that hive requires `--project`.
