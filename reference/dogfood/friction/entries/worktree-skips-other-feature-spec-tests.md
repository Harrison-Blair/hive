---
title: "A worktree and CI skip the tests that read the spec of another feature"
status: "Open"
opened: 2026-10-24
decision: ""
---

**Summary:**
Since task 140 (g14), `hive worktree create` copies only the feature folder
of the task into the worktree. Many `internal/tui` tests read fixtures and
grids from `.hive/spec/tui-v2/design/`. In a `golden-frames` worktree these
tests skip with `local spec file missing: ../../.hive/spec/tui-v2/design/...:
Git does not track .hive/spec/`. On 2026-10-24, `go test ./internal/tui/ -v`
in worktree 146-gf-logs gave 78 SKIP lines. CI has no `.hive/spec/` at all, so
CI skips the same tests. A task can break such a test, and the gate, the
verifier, and CI all stay green. Worker t143-worker-1 found it and copied
`tui-v2/` from the home by hand to run the tests.

**Reproduction steps:**
1. `hive worktree create --task ID` for a task of feature `golden-frames`.
2. In the worktree, run `env -u TERM TZ=UTC go test ./internal/tui/ -count=1 -v | grep -c '^--- SKIP'`.
3. The count is 78. In the home the same tests run.

**Workaround:** the manager tells each wave C worker and verifier to copy
`.hive/spec/tui-v2/` from the home into the worktree (and into the fresh
clone) before the race run, and to report the SKIP count.
