---
title: "A memory write refuses an unrelated file with no frontmatter"
status: "Open"
opened: 2026-10-26
decision: "Preserve the refusal and use the saved handoff. Do not change the existing memory file or product code."
---

**Issue:** A native memory write stops on an existing unrelated file with no frontmatter block.

**Status:** Open.

**Summary:**
The manager tried to save the accepted prose-task completion procedure as
project memory. The new input had YAML `name` and `description` fields.
The command refused an existing memory file instead of writing the requested
record. The manager did not change that file or add a product fix.

The exact command used standard input for the Markdown text:

```sh
HIVE_HOME=/home/penguin/source/hive /home/penguin/go/bin/hive memory write --scope project:hive prose-only-accepted-completion
```

The command reported:

```text
hive: memory file invalid: /home/penguin/source/hive/projects/hive/memory/handoff-before-live-schema4-control.md: frontmatter: no block
```

This records the malformed-frontmatter condition that the command observed.
It makes no claim about who created the existing file or another memory
operation. The manager retained the exact refusal and used the successful
native `hive handoff write --scope project:hive` command for durable state.
Task 183 continued with its branch and verification base unchanged.

**Reproduction steps:**
1. Keep the existing memory file named in the error without a frontmatter block.
2. Run the memory write command above with valid YAML frontmatter on standard input.
3. Observe the refusal that names the existing unrelated file.
4. Use the saved project handoff without changing that memory file.

**Evidence and limit:**
Task 183 evidence contains `manager-memory-write-refusal.json`, including
the exact command, environment, error, and handoff fallback.
Coordinator record 199 defers this entry until the next friction-only
window and requires the documentation commit before task 184 head selection.
Task 183 now has an actual independent pass and native source delivery.
This entry changes no memory file, state record, or product code.
