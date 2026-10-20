---
title: "A completed task cannot be reopened or edited"
status: "Open"
opened: 2026-10-20
decision: ""
---

**Summary:**
`hive task reopen` accepts a verified task, and `hive task edit` accepts a
created or assigned task. A completed task accepts neither. On 2026-10-20 the
repair worker of 85-logs-screen completed Repair 2. Before a verifier ran,
task 83 landed on `dev`, and the branch got a new conflict in
`internal/tui/tui.go`. The manager could not add Repair 3 to the brief or
reopen the task. It had to spawn a verifier only to record a fail, so that the
task went back to `assigned`.

**Reproduction steps:**
1. Run `hive task complete` on a task.
2. Run `hive task reopen --task ID --reason TEXT`: `invalid transition:
   completed -> reopen`.
3. Run `hive task edit --task ID --brief FILE`: `edit needs created or
   assigned`.

**Recurrence 2026-10-20 (stenographer 63-hub-sessions):** the worker
completed while tasks 61 and 72 landed on `dev`, and the branch got content
conflicts in `cli/hub/app.py` and `cli/hub/services.py`. `hive agent spawn
--role worker` refused (`task is completed`) and `hive task reopen` refused
(`completed -> reopen`). The manager again spawned a verifier only to record
the conflict as a failed check. Parallel tasks that touch one shared file make
this common.

**Related:** `hive worktree remove` on a task whose worktree is already
removed fails with `worktree check: start git: fork/exec /usr/bin/git: no such
file or directory`. Git exists. The working directory of the Git call is the
removed worktree. The message does not say that the worktree is gone.

**Recurrence 2026-10-20 (stenographer hub board tasks 109-111):** four tasks
ran in parallel and all edited `cli/hub/app.py`. Each landing (112, 115, 109,
116, 110) gave the others new conflicts. Tasks 110 and 111 each needed two
extra verifier runs only to record "dev merge needed", then a merge-only
worker; 111 also conflicted in `site/index.html`. That is six verifier and
worker cycles spent on conflicts. A verifier-free path for a completed task
that only needs a dev merge would remove most of them.
