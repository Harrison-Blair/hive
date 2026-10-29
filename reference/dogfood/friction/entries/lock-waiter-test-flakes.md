---
title: "TestWithLockWaiterIsNotStarved fails at times in the full gate run"
description: "The 100 ms waiter deadline can expire while the whole test suite runs in parallel"
status: "Open"
opened: 2026-10-29
decision: ""
---

**Issue:** `TestWithLockWaiterIsNotStarved` gives the waiter a 100 ms deadline
(`internal/lib/store/lock_test.go:396`). Under the load of the full
`go test -race ./...` run, the deadline can expire before the waiter gets the
lock. The failure is in the test timing, not in the lock.

**Summary:** On 2026-10-29, on branch `docs/friction-log`, the gate test step
failed once with `waiter WithLock() error = context deadline exceeded after
101.080738ms`. The branch changes nothing below `internal/`. A background agent
ran git commands at the same time. The test then passed 10 times of 10 alone
(`-count=10`), and the next full gate run passed. Before the teardown, the same
package had a similar flake in `TestUpdateCallbackRunsUnderLock`. Commit
93237c5b changed the lock wait to block in the kernel.

**Reproduction steps:**
1. Put the machine under load, for example with another build or test run.
2. Run `env -u TERM TZ=UTC go test -race ./...` from the root.
3. At times, `TestWithLockWaiterIsNotStarved` fails with a context deadline after about 100 ms.
