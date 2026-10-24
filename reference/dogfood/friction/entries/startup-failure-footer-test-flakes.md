---
title: "TestStartupFailureShowsInFooterAndDoesNotStopTUI fails at times in full package runs"
status: "Open"
opened: 2026-10-21
decision: ""
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
