---
title: "The user must ask an agent to change a hive setting"
status: "Open"
opened: 2026-10-22
decision: ""
---

**Summary:**
On 2026-10-22 the user reported that they must talk to an agent to change a
hive setting. The settings are in files: model presets and preferences in
`overrides/` (for example `overrides/presets.md`), and project records in
`.hive/state/projects/`. The TUI has no screen that shows or changes these
settings. Thus the user asks the coordinator, and the coordinator edits the
file. The user wants a settings tab in the TUI.

**Reproduction steps:**
1. Run bare `hive`.
2. Try to change the model preset of a role, or a project setting.
3. No screen shows the setting. The user must edit the file or ask the
   coordinator.
