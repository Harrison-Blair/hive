---
title: "A feature spec task cannot start in a repository that ignores .hive/spec/"
status: "Open"
opened: 2026-10-19
decision: ""
---

**Summary:**
Since the spec migration (task 48), `hive task create` needs `--feature`,
and `hive worktree create` writes the brief to
`.hive/spec/<feature>/tasks/<id>/brief.md` in the project worktree. When the
project `.gitignore` does not allow `.hive/spec/`, the worktree create fails:
"path is ignored ... The .gitignore of the repo must allow .hive/spec/". The
only way to allow the path is a source change, which needs a task and a
worktree. So the first spec task of such a project cannot start. A manager
never edits a source file, so the manager is blocked. It occurred once, in
project `stenographer`, whose `.gitignore` ignores every path that it does not
allow.

**Reproduction steps:**
1. Use a project whose `.gitignore` starts with `*` and does not allow `.hive/`.
2. Run `hive task create --project ID --feature F --spec --title T --slug S --brief FILE`. It succeeds.
3. Run `hive worktree create --task ID-S`. It fails with "path is ignored".
4. Run `hive task create` without `--feature`. It fails with "feature requires a value".
