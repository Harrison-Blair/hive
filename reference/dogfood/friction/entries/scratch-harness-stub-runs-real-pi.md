---
title: "A scratch harness stub did not override the installed Pi executable"
status: "Open"
opened: 2026-10-23
decision: "Use the explicit stub executable for scratch delivery checks"
---

**Summary:**
During task `135-ask-confirm`, worker `t135-worker-1` intended to use a
scratch `pi` stub for answer delivery. The scratch panes ran the installed
`~/.local/bin/pi` instead of `/tmp/t135v4/bin/pi`.

The first delivered answer started a real Pi turn. The worker aborted it
with Escape before any tool call. Its working directory was `/tmp/t135v4`.
The real home and real herdr server got no test message.

The worker used the explicit stub for the parent-routing and dismiss cases.
It removed the scratch home, herdr server, and tmux server after the checks.
The cause of the executable choice remains unknown.

Evidence: manager inbox record 552 and commit `259aa9d7` on
`task/135-ask-confirm`, file
`.hive/spec/tui-v2/tasks/135-ask-confirm/evidence/V4.md`.

**Reproduction steps:**
1. Create an isolated scratch home and herdr server for a delivery check.
2. Put a `pi` stub in a scratch `bin` directory.
3. Start the scratch asker and parent panes with the Pi harness.
4. Check which executable runs before sending the test answer.
5. Use the explicit stub executable when the installed harness runs instead.

The worker observed this once. The manager did not independently repeat it.
