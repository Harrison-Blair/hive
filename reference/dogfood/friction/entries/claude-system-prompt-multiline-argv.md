---
title: "Native Claude spawn rejects the multiline system prompt"
status: "Open"
opened: 2026-10-26
decision: "Fix the existing D4 instruction delivery with a private file. Keep model and trust behavior unchanged."
---

**Issue:** Native Claude spawn puts multiline instructions in an argument that herdr refuses.

**Status:** Open.

**Summary:**
On 2026-10-26, the task 184 upload fix needed the new approved Claude Code
model, `claude-opus-5-5`. Native `hive agent spawn --profile builder` failed
with `invalid_agent_argument: agent arguments cannot be encoded safely for
the target shell`. No builder started in that attempt.

`internal/agent/brief.go` combines the profile, guide, and reset reminder.
`systemPromptArgs` passes that multiline text after `--append-system-prompt`.
Both ordinary spawn and coordinator startup use this path. The installed
Claude Code 2.1.310 supports `--append-system-prompt-file`.

Coordinator records 241, 242, and 247 authorized a bounded manual workaround
for task 184. The manager wrote the exact unchanged instruction bytes to an
ignored file and opened the existing worktree without changing its source.
Herdr started Claude with the approved model and permission arguments, using
the supported file flag. No trust setting or environment override was added.
The file held 23,138 bytes, with SHA-256
`def71bc2ae02366bc47479224a9f54cb6287c49c91edd6678169547618750c3c`.

The actual agent registered itself from its own pane. Its native record kept
profile `general`, `spawned: false`, and no parent. The manager checked its
actual binding before the one-time authorized task assignment. These limits
were recorded without state JSON changes. The agent completed task 184.
After late evidence preservation, the manager checked the owned process and
pane identity, closed only that pane, and used native exited settlement.
A delayed unresolved record preceded that settlement and required no new fix.

**Reproduction steps:**
1. Use the accepted control binary and a profile with multiline text.
2. Run `hive agent spawn --profile builder --task ID --harness claude --model claude-opus-5-5`.
3. Observe the argument-encoding refusal before a native Claude start.
4. Inspect `systemPromptArgs` and the Claude system-prompt flag.

**Evidence and limit:**
Task 184 evidence contains `manual-claude241/`, including exact instruction
hashes, start receipts, self-registration, binding, and assignment receipts.
`manager-manual-close247/` preserves the late-evidence hash manifest, actual
identity checks, pane close, process absence, and native exited settlement.
The manual file route succeeded. It did not fix ordinary native spawn.
Coordinator records 245 and 247 require a bounded D4 source fix before G2.
No generic framework, schema change, encoder bypass, or UI change is requested.
