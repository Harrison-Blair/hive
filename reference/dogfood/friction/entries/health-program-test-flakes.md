---
title: "TestHealthProgramUsesStringReader fails at times on TempDir cleanup"
status: "Open"
opened: 2026-10-21
decision: ""
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
