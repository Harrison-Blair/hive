---
title: "Each brief copy row conflicts in the STE baseline"
status: "Open"
opened: 2026-10-20
decision: ""
---

**Summary:**
Each spec task adds one row with 0 for its brief copy to
`testdata/ste-baseline.json`. Two tasks that run in parallel add rows at the
same place, so the second land fails with a merge conflict. Each conflict
needs a reopen, a repair worker, and a new verifier. On 2026-10-20 this
occurred for 85-logs-screen (repair 2) and 89-screen-board-audit (repair 1
after 94-chat-user-only landed). Wave D has ten parallel tasks.

**Reproduction steps:**
1. Create two spec tasks from the same `dev` commit.
2. Let each worker add its brief copy row to `testdata/ste-baseline.json`.
3. Land the first task.
4. Run `git merge-tree --write-tree origin/dev origin/task/<second>`: a
   conflict in `testdata/ste-baseline.json`.
