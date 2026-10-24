---
title: "Memory write help omits the required frontmatter"
status: "Open"
opened: 2026-10-24
decision: ""
---

**Issue:** `hive memory write --help` gives no input format.

**Summary:**
`hive memory write` reads Markdown from standard input. It requires YAML
frontmatter with `name` and `description` fields. The help gives only the
command and flags. The coordinator and manager skills give no input template.
On 2026-10-24, both agents supplied plain Markdown. Each command failed with
`frontmatter: no block`. Each write succeeded after the agent added the two
fields. The failure gives no example of the required format.

**Reproduction steps:**
1. Run `hive memory write --help`. It gives no format for standard input.
2. In a scratch home, supply plain Markdown to `hive memory write --scope home example`.
3. The command fails with `frontmatter: no block`.
4. Add YAML frontmatter with `name: example` and `description: Example memory`. The write succeeds.

**Workaround:** supply this format on standard input:

```markdown
---
name: example
description: Example memory
---
The memory text.
```

No source fix or task is requested in this entry.
