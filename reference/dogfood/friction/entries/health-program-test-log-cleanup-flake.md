---
title: "TestHealthProgramUsesStringReader fails at random on the log cleanup"
status: "Open"
opened: 2026-10-24
decision: ""
---

**Summary:**
Worker t140-worker-3 found this flake during the repair of task
140-spec-local-state on 2026-10-24. `TestHealthProgramUsesStringReader`
(`internal/tui/health_test.go:318`) fails about 2 or 3 times in 300 runs. The
error is `TempDir RemoveAll cleanup: ... .hive/logs: directory not empty`. The
fault occurs with and without `.hive/spec/`. The probable cause is a log write
of `RunHome` after the function returns. The flake can make the gate fail at
random. Evidence: task 140 `evidence/repair-2/health-flake-full-spec.log`.

**Reproduction steps:**
1. Check out `dev`.
2. Run `go test ./internal/tui/ -run TestHealthProgramUsesStringReader -count=300`.
3. About 2 or 3 runs fail with the cleanup error.
