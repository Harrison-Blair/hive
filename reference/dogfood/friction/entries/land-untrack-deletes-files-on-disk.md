---
title: "hive land of a git rm --cached commit deletes the files from the project checkout"
status: "Open"
opened: 2026-10-24
decision: "qmeter g15: back up with cp -a before hive land, restore after"
---

**Summary:**
`hive land --mode merge` merges in the primary checkout of the project. That
checkout still tracks the files, so a merge of a commit made with
`git rm --cached` deletes them from its disk. qmeter task 138 (g15) untracked
`.hive/spec/`. The merge `adebfdf` left 0 of the 55 files on disk. The manager
restored them from a `cp -a` backup made before the landing. Each other clone
loses the files at its next pull. hive g14 untracks `.hive/spec/` in hive and
meets the same case. It happened once.

**Reproduction steps:**
1. Make a task whose commit runs `git rm -r --cached DIR` and ignores `DIR`.
2. Verify the task, then run `hive land --task ID --mode merge --grant TEXT`.
3. Run `find DIR -type f | wc -l` in the project checkout. It prints `0`.
