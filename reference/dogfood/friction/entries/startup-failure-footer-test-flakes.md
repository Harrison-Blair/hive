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
