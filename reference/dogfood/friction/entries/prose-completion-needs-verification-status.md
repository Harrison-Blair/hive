---
title: "A prose-only result has no accepted completion status without verification"
status: "Open"
opened: 2026-10-26
decision: "Record accepted external completion. Cancel only the waived workflow. Keep code verification unchanged."
---

**Issue:** Native delivery and dependency filtering require verification even when a task changes only external prose.

**Status:** Open.

**Summary:**
Task 182 updates PR 49 title, body, and image attachments. It changes no
repository source. Its branch is already on `dev`. A fresh independent
check failed on two documentation details. Builder 2 fixed both details
and completed the task. The earlier failed verdict remains historical.

The user then requested no separate verification for prose-only tasks.
The user permits a real check if required for progress, but prefers to skip it.
`hive land` requires status `verified` and a recorded pass. It has no waiver
flag. The ready list also filters dependencies through verification status.

Coordinator record 196 gives the supported workaround for tasks 182 and 184:

1. Preserve completed history, the external PR result, the user instruction,
   and any historical verdict or correction in evidence.
2. Record the prose result as `COMPLETE/ACCEPTED`.
3. Cancel only the waived verification and landing workflow with an explicit
   acceptance reason. The PR work is complete. It is not abandoned.
4. Preserve evidence before native worktree removal.
5. Close the PR wave on its actual external result.
6. Start an approved dependent task by explicit ID after its written prerequisites.

Native worktree creation and builder spawn support the last step without
requiring a verified prose dependency. Keep the task IDs and dependencies.
Do not fabricate a pass or delivery. Do not edit state JSON.
Code tasks still require their real independent checks and native delivery.
The cancelled label and ready-list omission do not express accepted completion
clearly. This entry records that status limitation. This entry requests no product fix.

**Reproduction steps:**
1. Complete a task that changes only an existing PR summary and attachments.
2. Waive separate verification under the user's explicit instruction.
3. Inspect `hive land` and the ready-list rules.
4. Observe the required verification status despite no source to merge.
5. Use the documented accepted-result closure and explicit dependent task ID.

**Evidence and limit:**
Task 182 evidence contains `manager-prose-only-waiver.json`,
`manager-closure/accepted-external-completion.json`, the corrected builder
reports, and the preserved historical failed verdict.
`internal/land/land.go` requires status `verified` and a pass.
The manager recorded no fresh verifier, false pass, or native delivery for this closure.
