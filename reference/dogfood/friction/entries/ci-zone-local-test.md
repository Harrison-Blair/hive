---
title: "A test that passes in the local zone fails the UTC CI gate"
status: "Open"
opened: 2026-10-23
decision: "Task 132 fixes the two tests. The gate still runs in the local zone."
---

**Summary:**
On 2026-10-23 six GitHub Actions runs on `dev` failed in the step "Test with
the race detector". Two TUI tests failed: `TestHeaderShowsTheOpenAskRecords`
(`internal/tui/ask_test.go`) and `TestHeaderShowsTheAskCountOfTheState`
(`internal/tui/chat_modes_test.go`). The tests use the fixed instant
`2026-10-20T12:00:00Z`. The test clock shows it in `time.Local`, and the tests
expect the text `08:00`, which is the US Eastern text. The owner's machine
uses US Eastern, so the local gate passed. The CI runner uses UTC, where the
same instant shows as `12:00`. A worker and a verifier can pass the gate on a
tree that fails CI. The first failing landing was task 126 (`0c12e20e`).
Tasks 129 and 131 landed on the red tree.

**Reproduction steps:**
1. Write a test with a fixed instant and an expected clock text of the local zone.
2. Run the gate in a zone where the text agrees with the test. The gate passes.
3. Push the commit. The CI runner uses UTC and the test fails.
