---
title: "TestUpdateCallbackRunsUnderLock fails at times under heavy load"
status: "Open"
opened: 2026-10-21
decision: ""
---

**Summary:**
On 2026-10-21 the worker of task `33-screen-agents` ran the full gate while
eight screen tasks ran at the same time. The load average was 21. The test
`TestUpdateCallbackRunsUnderLock` in `internal/lib/store` failed with
`store lock timeout`. It passed five times alone, and the next full gate
passed. The test is not in the package of the task. The lock timeout of the
test is possibly too short for a host under heavy load.

The same day, the worker of task `34-screen-chat` saw the same fail in the
full gate, with `TestConcurrentUpdateProcesses` also. Both passed three times
alone, and the next full gate passed.

**Reproduction steps:**
1. Run several `go test -race ./...` gates at the same time on one host.
2. See `TestUpdateCallbackRunsUnderLock` fail with `store lock timeout`.
3. Run the test alone. It passes.
