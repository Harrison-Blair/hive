---
title: "A report to a parent in a dialog is lost"
description: "hive agent send refuses delivery while the target pane shows a dialog, and the sender ends its turn without a retry"
status: "Open"
opened: 2026-10-30
decision: ""
---

**Issue:** `hive agent send` refuses delivery when herdr reports the target
agent as blocked by a dialog. It exits with status 1 and sends nothing
(`internal/agent/send/send.go:83-84`). Hive does not wait, retry, or queue the
message. A worker that reports to its parent then often ends its turn, and the
report is lost. A coordinator asks the user questions often, so its pane often
shows a dialog when a worker reports. The work then stalls with no signal.

**Summary:** On 2026-10-30 the reviewer agent `settings-review-1` (codex)
finished its review and ran `hive agent send coordinator '…'`. The coordinator
pane (a Claude Code session) showed an AskUserQuestion dialog at that moment.
Hive printed `hive: agent coordinator is blocked by a dialog in pane w14V:pK:
answer it in the pane; the prompt was not sent`. The reviewer did not retry and
ended its turn. The coordinator found the report only when the user asked
"hello?" and the coordinator read the reviewer pane. Workaround: the parent
reads the pane of the worker with `herdr pane read`.

Options, not decided:
- `send` waits or retries until the target leaves the dialog.
- Hive queues the message and delivers it when the target is ready.
- The error tells the sender to retry, and the hive guide says so.

**Reproduction steps:**
1. Start an agent A and a Claude Code agent B with hive.
2. In B, open an interactive dialog, for example AskUserQuestion.
3. From A, run `hive agent send B TEXT`.
4. Observe exit status 1 and the error "agent B is blocked by a dialog in pane P: answer it in the pane; the prompt was not sent". B never receives TEXT.
