---
title: "A paused supervisor leaves incident inbox messages without nudges"
status: "Open"
opened: 2026-10-24
decision: "Preserve the durable inbox record and directly notify the live recipient during incident coordination."
---

**Issue:** Durable incident messages did not cause timely inbox reads while the supervisor was paused.

**Summary:**
The coordinator paused the runtime during the schema incident on 2026-10-24.
Inbox records remained durable, but the paused supervisor could not deliver
its usual nudges. Records 612–618 in the coordinator inbox and child freeze
messages therefore needed a separate notification path. A stored message alone did
not establish that its recipient read the freeze instruction.

The coordinator explicitly permitted installed `hive agent send` for incident
notifications alongside the durable inbox. After each report, the sender
also tells `coordinator-g15` to read it. Child freeze acknowledgments remain
explicit, with process inventories and safe-stop limits.

**Reproduction steps:**
1. Inspect the preserved incident timeline and queued inbox records.
2. Compare durable message timestamps with the later recipient acknowledgment.
3. Observe that no supervisor nudge ran during the pause.
4. In any future executable check, use a synthetic isolated supervisor and inbox. Do not pause production for reproduction.

**Workaround:** Write the durable inbox record first. Use installed
`hive agent send` to notify the known live recipient under the explicit
incident/progress exception. Check the recipient acknowledgment separately.
A stopped recipient or an unanswered prompt remains a delivery limit.

Do not restart a supervisor merely because its claim is stale. The incident
preserves uncertainty about the old process: PID 2075927 was stopped, then
later absent in OS checks, without a coordinator resume or termination.
The user TUI PID 1318049 uses installed commit 3187a9fb. These observations
permit no automatic startup, repair, installation, or production process action.

**Evidence:** Coordinator containment and direct-notification instructions,
manager-hive incident inbox/handoff records, and child acknowledgment 924
(reconfirmed in 927). The incident memory and backups retain the timeline.

This entry records the workaround. It requests no new capability or feature scope.
