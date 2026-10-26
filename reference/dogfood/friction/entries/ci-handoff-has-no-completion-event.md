---
title: "A pending CI handoff has no completion event and stalls the next task"
status: "Open"
opened: 2026-10-26
decision: "Use a bounded one-shot gh run watch with a native project inbox outcome before ending with unfinished CI."
---

**Issue:** A manager ended after checking running CI without arranging a completion notification. The approved dependent task then waited for a user nudge.

**Status:** Open. The one-shot notification mechanism works on the completed checkpoint. Its next use must cover an unfinished landing checkpoint.

**Summary:**
Task 172 landed and pushed to dev at abe363e66b6378753f4bddbc64edd6b070174c30. The manager saw CI runs 40361195688 and 40361176408 in progress, then ended with task 173 waiting. No CI completion event would resume the chain.
Coordinator record 307 gives completion times 16:41:51-52 UTC and the stale pending report at 16:42:29 UTC on 2026-10-26. The user c241 nudge at about 16:49 resumed the chain. The manager then checked actual success, granted task 173 its slot, and received acknowledgment 308 at 16:51:07 UTC.
This is an orchestration handoff failure. It does not authorize a product source change, another agent, or a runtime restart.

**Reproduction steps:**
1. Land an independently verified task and push dev.
2. Check the exact-head CI runs while they are in progress.
3. Leave the next approved task waiting and end without a completion notification.
4. Let CI finish. The chain has no event to resume it until another nudge arrives.

**Workaround:**
Before ending with unfinished CI, start a detached, bounded gh run watch --exit-status for each exact-head run. Record the final head, status, conclusion, and command exit. Send the actual success, failure, timeout, or status-read error to project:hive with the accepted native CLI. Save launch PID, arguments, logs, outcome, and inbox receipt in the task evidence. Use the result to advance the next slot or report an actual failure.
The trusted one-shot helper is .hive/tmp/ci-notify/watch.py. It uses a 1800-second limit per run and no observer agent. An actual run on the already-completed task 172 checkpoint produced native inbox outcome 309, with both watch exits 0 and notification exit 0. This proves the notification path for that completed checkpoint; it does not claim a pending run was watched then.

**Evidence:**
Coordinator records 306-307, builder acknowledgment 308, and CI notification 309. Task 172 evidence/ci-watch-abe363e6/ preserves actual launcher arguments, watch logs, status responses, outcome.json, and notification receipts. Task 173 evidence/manager-slot-working-agent.json records the resumed builder.
