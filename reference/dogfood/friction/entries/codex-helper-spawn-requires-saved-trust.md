---
title: "Optional Codex helper spawn stops at the saved-trust check"
status: "Open"
opened: 2026-10-25
decision: "Preserve the refusal. Do not change trust for optional parallel work. The parent verifier owns the verdict."
---

**Issue:** A Hive Codex helper spawn requires saved trust for its working directory.

**Status:** Open.

**Summary:**
During task 181 fresh verification on 2026-10-26 UTC, the verifier tried to
start three optional general Codex helpers in its existing task worktree.
Hive refused each spawn in validation, with no effects. The error says the
Git repository has no saved trust in `~/.codex/config.toml` and can show a
folder trust dialog. The check used the accepted schema-3 control binary
from `a813921f`.

The verifier changed no trust setting and used no trust bypass. No optional
Hive helper pane was created. It used the available harness agent runner for
private source archives at the frozen candidate `5044a687`. Each helper used
isolated home, config, and socket paths. The parent verifier kept the gate,
captures, evidence checks, cleanup, and final verdict. The refusal blocked no
required verification item.

**Reproduction steps:**
1. Use a Git working directory with no saved Codex trust.
2. Run `hive agent spawn --profile general --harness codex --cwd PATH --prompt-file FILE --json`.
3. Observe exit 1, phase `validate`, no effects, and the saved-trust message.

**Evidence and limit:**
Task 181 evidence contains `verifier-2/optional-spawn-refusal.json`, the raw
stdout/stderr, command metadata, and helper prompts. Exact copies of the
refusal are preserved in `manager-preserved-optional-spawn/`.

This records the workflow friction and the fallback. It does not establish
that the trust check is wrong. No trust-policy or product change is requested.
