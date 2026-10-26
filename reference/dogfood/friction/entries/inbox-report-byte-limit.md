---
title: "An inbox report over 4096 bytes is refused"
status: "Open"
opened: 2026-10-26
decision: "Send a concise report within the limit and link the full evidence. Preserve the refused attempt."
---

**Issue:** A detailed verifier report can exceed the native inbox message limit.

**Status:** Open.

**Summary:**
Task 190's verifier attempted to send a 4,690-byte report to `project:hive`.
The native command refused it because the limit is 4,096 bytes. The report
did not reach the owner on that attempt. The verifier preserved the refusal
and sent a shorter report, which arrived as owner record 360. Its full
`VERIFIED.md`, raw checks, and qualification archives remained in task evidence.

**Reproduction steps:**
1. Prepare a verifier report longer than 4,096 bytes.
2. Send it with `hive inbox send --to project:hive --agent AGENT TEXT`.
3. Observe the byte-limit refusal and preserve its output and exit status.
4. Send a report within the limit with the full evidence path. Check delivery.

**Evidence and limit:**
Task `190-inbox-domain` evidence in `verifier-1/` retains the refused first
report and successful shorter report. Owner record 360 states both actual
outcomes. No limit bypass, state edit, or product change was used. This entry
records reporting friction and its workaround; it requests no source fix.
