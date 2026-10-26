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


**Later decision and required fallback:**
Coordinator records 197 and 198 preserve the actual accepted closure of task
182. The next prose task must not use cancellation merely to clear guards.
The user permits real verification if needed to continue, while preferring
to skip it. The native normal closure still requires a pass and has no waiver.
For task 184, the manager plans one real summary and evidence check if that
requirement remains. This check runs no code gate or product test.
After a real pass, normal already-merged delivery can close the task honestly.
This is process friction: a prose-only update needs a separate verifier solely
for native completion bookkeeping. The plan does not claim that the future
check already ran or passed. The manager reports the exact disposition before
finalization. This entry requests no source feature or runtime change.


**Actual task 184 fallback:**
Task 184 used the permitted real-check fallback after the user asked to skip
separate prose verification. The final fresh verifier checked only PR 49,
its evidence, and all 105 authenticated remote image bytes. It ran no product
test, code gate, build, or new capture. V1, V2, and V3 passed.
The manager then used normal already-merged merge-mode delivery at
`f683e0c304e65400bd7f0cfd871c8a551dc7fcad`, with the existing PR as evidence.
Native delivery reported that the branch was already on `dev` and pushed it.
Task 182's earlier accepted closure and task 184's historical failed verdict
remain unchanged. This actual result replaces no earlier history.
The separate verifier was required for native closure. The status limitation
and the user's preference remain open process friction.

**Actual task 177 audit closure on 2026-10-26:**
Task 177 updated the ignored command inventory and feature README. Its only
tracked change was its own STE baseline entry at zero. It changed no Go
source or test. The accepted audit brief and delivery grant still required
a real independent pass before native landing. The manager used that
fallback and did not ask the user for another decision.

The fresh verifier checked all 42 inventory rows, the actual predecessor
deliveries, retained APIs, CLI behavior, and the required gate. It added no
regression test or semantic-break campaign for this documentation audit.
The executed CLI check had 264 passes and no skips. The passing race gate
had 3,792 passes and 17 actual skips; those skipped paths remain unverified.
Its first reporting-setting failure and unchanged-base comparison remain
preserved separately from the passing gate.

Native pass 366 permitted merge/push to `dev` at
`8fbb6029130071e8779289b6f570befa2b216d6a`. Native landing also copied the
ignored final inventory and README to the home; their bytes were checked.
This records the required fallback. It does not resolve the user's request
to skip separate verification for prose-only work. Direct friction-log
updates continue to need no builder or verifier under the project rules.
