# VERIFIED — 93-ci-test-fixes

Verifier: `t93-verifier-1` (Claude Opus 5.5). Date: 2026-10-20.
Task branch: `task/93-ci-test-fixes` at `8d442b6c`.
Base: `46c363e9` (merge base with `origin/dev`), in a detached scratch worktree.
Each run used a temp `HIVE_HOME` and a temp `XDG_CONFIG_HOME`. No TUI run used the real home.

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | R1: with BusyBox v1.36.1 (`/usr/lib/initcpio/busybox`, multi-call) linked as `sleep` first on `PATH`, both supervisor tests fail on the base with `read supervisor executable: readlink /proc/PID/exe: no such file or directory`. R2: `go test -race -count=50 -run TestRunHomeLogsKeysOnStringReader ./internal/tui/` on the base: 3 of 50 fail. Under CPU load (32 busy loops on 16 CPUs, load average 27.6), `-count=20` on the base: 13 of 20 fail. |
| V2 | pass | R1: the same BusyBox run on the task branch: `ok internal/doctor`. R2: `-count=50` on the task branch: 0 failures, `ok ... 201.580s`. Under load from a parallel `go test -race -count=3 ./...` run: 20 of 20 pass. Under the 32 busy loops: 20 of 20 pass. |
| V3 | pass | `origin/task/93-ci-test-fixes` is at `8d442b6c`. No pull request. CI not run: the workflow has no `workflow_dispatch`. |
| V4 | pass | `git diff origin/dev...HEAD --stat`: 9 files. Each is in the source table: the brief copy and evidence, `internal/doctor/health_test.go`, `internal/tui/logs_screen_test.go`, `testdata/ste-baseline.json`. No non-test Go file changed. `ci.yml` did not change. |
| V5 | pass | `gofmt -l` printed nothing. `go mod tidy -diff` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` `0 issues.` exit 0. `go test -race ./...` exit 0. |

## Assertion check

- R1: each assertion in the two supervisor tests did not change. Only the fake process changed: a copy of the test binary named `supervisor` that runs `TestFakeSupervisorProcess` and sleeps. It stays live under any `sleep`.
- R2: the base asserted `▶ 8 Logs` and `level: warn` in the final output. The task test waits for `▶ 8 Logs` and the fixture event, sends `l`, then asserts `▶ 8 Logs` and that `warnno events today` follows the event in the output (the renderer writes only the changed cells, so the warn label and the filtered empty list show as one run). The new model test `TestLogsScreenCyclesInfoLevelToWarn` asserts the exact `level: warn` text. The end-to-end check also proves the filter removed the INFO event. No assertion is weaker than on the base.

## Note for the land

`origin/dev` moved to `ef2dde2e` (task 88) during verification. Task 88 also changed `TestRunHomeLogsKeysOnStringReader`: it removed the `level: warn` output check and checks a `tui action` log event instead. `git merge-tree origin/dev HEAD` reports a conflict in `internal/tui/logs_screen_test.go`. The manager must resolve it at the land and run R2 again on the result.
