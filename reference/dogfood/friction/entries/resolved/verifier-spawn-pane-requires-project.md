---
title: "A verifier spawn with `--pane` requires `--project`"
status: "Resolved 2026-10-19"
opened: 2026-10-19
decision: "Fixed in 17-verifier-project (e7031497): hive derives the project from the task, and the docs are fixed."
---

**Summary:**
The manager skill shows a verifier spawn without `--project`. A
verifier spawn with `--pane` requires `--project`.

**Reproduction steps:**
1. Read the verifier spawn command in `hive-manager/SKILL.md`.
2. Add `--pane PANE`. Do not add `--project`.
3. Run the command.
4. Check that hive requires `--project`.
