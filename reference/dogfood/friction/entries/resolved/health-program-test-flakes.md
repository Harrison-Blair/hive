---
title: "TestHealthProgramUsesStringReader fails at times on TempDir cleanup"
status: "Resolved 2026-10-25"
opened: 2026-10-21
decision: "Tasks 165 and 179 settle home-access loaders and close the TUI writer before return."
---

**Summary:**
On 2026-10-21, worker `t11-worker-2` ran `TestHealthProgramUsesStringReader`
in `internal/tui/` 40 times on `origin/dev` and 40 times on the task branch
`task/11-demo-mode`. It failed 4 times on `origin/dev` and 3 times on the task
branch. Each fail is in the `t.TempDir` cleanup: `directory not empty` for
`.hive/logs`. A log write can occur after the program ends and before the
cleanup. The fail can make a gate run red with no change in the code. A
verifier can then record a false fail.

**Reproduction steps:**
1. On `dev`, run
   `go test -race -run TestHealthProgramUsesStringReader -count=40 ./internal/tui/`.
2. Some runs fail in the `TempDir` cleanup with `directory not empty` for
   `.hive/logs`.

**Recurrence 2026-10-24:**
Worker t140-worker-3 saw the same fault during the repair of task
140-spec-local-state. The test failed about 2 or 3 times in 300 runs, with and
without `.hive/spec/`. The probable cause is a log write of `RunHome` after the
function returns. Evidence: task 140 `evidence/repair-2/health-flake-full-spec.log`.


**Resolution 2026-10-25:**
Tasks 165 and 179 landed on `dev` in merge `1599859c`. The verified source
candidate is `22c783f9`. The TUI settles admitted home-access loaders before
return and closes its writer after accepted events finish. Header and Health
commands receive the shutdown context. The separate CLI writer keeps its
command-end event.

The worker and independent verifier passed both exact 2000-run race checks
and the exact 50-run race check under continuous second load. All 75 load
iterations in the final independent check exited 0. The full gate passed.
Evidence: task 165 `evidence/VERIFIED.md` and task 179 `evidence/VERIFIED.md`
in the home feature folders.
