---
title: "TestStartupFailureShowsInFooterAndDoesNotStopTUI fails at times in full package runs"
status: "Resolved 2026-10-25"
opened: 2026-10-21
decision: "Tasks 165 and 179 settle header work before return and preserve the startup error and log events."
---

**Summary:**
On 2026-10-21 the worker of task `33-screen-agents` saw
`TestStartupFailureShowsInFooterAndDoesNotStopTUI` in `internal/tui` fail
twice in full package runs without `-race`. The test passed 40 times alone,
and the gate passed. The worker thinks that the `ctrl+c` input of the test can
come before the asynchronous startup error. The test is not in the files of
the task.

**Reproduction steps:**
1. Run `go test ./internal/tui/` several times while other gates run on the host.
2. See `TestStartupFailureShowsInFooterAndDoesNotStopTUI` fail at times.
3. Run the test alone. It passes.

**Recurrence 2026-10-24:**
Verifier t140-verifier-3 of task 140-spec-local-state saw the test fail once in
a full package sweep with `TempDir RemoveAll cleanup: directory not empty`. Its
single rerun passed. On `dev` at `9ba502c5` the test passed 300 runs alone and
200 runs with `-race`. The fault is the same class as
`health-program-test-flakes`: a log write after the TUI returns, under load.


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
