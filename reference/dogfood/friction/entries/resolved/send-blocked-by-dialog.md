---
title: "A report sent to a pane with an open dialog is lost"
description: "hive agent send refuses a target that a dialog blocks, and a worker that ends its turn never sends the report again"
status: "Resolved 2026-10-30"
opened: 2026-10-30
decision: "hive agent send writes the message to the claude inbox socket, and claude queues it while a dialog is open and delivers it after the dialog closes (live check with claude 2.1.314, an AskUserQuestion dialog). Send no longer refuses a blocked target. With --wait it returns partial with status blocked and the message waits. The second point, the worktree name with a slash, moved to the entry worktree-name-no-slash"
---

**Issue:** `hive agent send` fails in phase `prompt` when herdr reports the
target agent as blocked by a dialog (`internal/agent/send/send.go:83-85`). The
text is not sent, and hive keeps no copy. With `--wait` the send fails the same
way, because the refusal comes before the wait starts. A worker that reports
to the coordinator while the coordinator shows a dialog gets only the error.
When the worker then ends its turn, nothing retries, and the report is lost.

Second point: a `--worktree` name of `hive agent spawn` cannot contain `/`. So
a worktree branch cannot have a name such as `friction/SLUG`, and the agent
must push with `git push origin HEAD:friction/SLUG` to get that remote branch
name.

**Summary:** On 2026-10-30 the coordinator pane `w15P:p1` (Claude Code) had an
AskUserQuestion dialog open. The worker `gitignore-builder` finished and ran
`hive agent send coordinator REPORT`. Hive refused the send because of the
dialog. The worker said that it would retry after the dialog closed, then
ended its turn. The report never arrived. The coordinator found the report
only when it read the worker pane.

Workaround: when a worker shows `done` and no report arrived, the orchestrator
reads the worker pane with `herdr pane read`.

Possible fixes, not decided:
- Hive queues the message and delivers it when the target is idle.
- `hive agent send` waits, with a timeout, until the dialog closes, then sends.
- The worker brief tells agents to retry the send in a loop until it is delivered.

**Reproduction steps:**
1. Run bare `hive` to start the coordinator, and spawn one agent.
2. In the coordinator, open a dialog, for example an AskUserQuestion dialog.
3. Make the agent run `hive agent send coordinator TEXT`.
4. The send fails with "agent coordinator is blocked by a dialog". The text does not reach the coordinator.
5. Let the agent end its turn. Close the dialog. The text never arrives.
