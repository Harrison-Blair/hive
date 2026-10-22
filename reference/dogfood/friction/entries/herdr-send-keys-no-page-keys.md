---
title: "herdr pane send-keys has no page up or page down key name"
status: "Open"
opened: 2026-10-21
decision: ""
---

**Summary:**
On 2026-10-21 the worker of task `34-screen-chat` needed `pgup` and `pgdn`
for a live check of the Chat feed. herdr 0.9.5 `pane send-keys` gives
`invalid_key` for `pgup`, `pageup`, and `PageUp`. The worker sent the escape
sequences `ESC [5~` and `ESC [6~` with `pane send-text` instead. Each live
check of a TUI that scrolls by page needs this workaround.

**Reproduction steps:**
1. Run `herdr pane send-keys PANE pgup`.
2. See the error `invalid_key`.
3. Send `ESC [5~` with `herdr pane send-text`. The page moves up.
