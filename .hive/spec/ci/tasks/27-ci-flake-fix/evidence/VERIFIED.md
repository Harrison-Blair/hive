# 27-ci-flake-fix VERIFIED

| Item | Value |
| --- | --- |
| Task | 27-ci-flake-fix |
| Commit | 1b4a76e5 |
| Base | dev 3717fb22 |
| Date | 2026-10-19 |
| Agent | verifier, claude-code, claude-opus-5-5 |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | R1, R4 | pass | V1.md gives the CI log lines and the code lines for each of the four failures. The TUI fix replaces `program.Kill()` with `program.Quit()` in `RunHome`. The pre-fix run below shows `WARNING: DATA RACE`. The post-fix runs show no race. The supervisor fix makes `Close` wait for the reader goroutine before it closes the inotify descriptor. The read is nonblocking and polls `done`. The test keeps its one-second check. The one test-file change (`internal/lib/store/grouped_test.go`) gives each lock operation a 10 s deadline in place of the 2 s default (`lock.go:21,74-78`). The test keeps 200 rounds and the equal-version assertion. The inbox `Append` lock wait is 10 s in production code. The test that checks it does not check speed. V1.md gives the reason for each longer wait. No test is skipped or deleted. The brief copy changed one sentence, recorded in V1.md. |
| V2 | R2 | pass | Same commands, detached worktree at `99ed618d` for the runs before the fix. Before the fix: TUI fails in both runs (data race). Supervisor watcher fails in both runs (4 and 60 failures). Inbox, store, and SIGTERM pass in both runs. V1.md explains the inbox and store causes from the CI logs (`store lock timeout` at the 2 s default wait) and the code. SIGTERM did not reproduce, which V1.md records. After the fix: all ten runs exit 0. |
| V3 | scope | pass | `git diff dev --stat` shows `internal/lib/inbox/inbox.go`, `internal/lib/store/grouped_test.go`, `internal/supervise/watch.go`, `internal/tui/tui.go`, `spec/hive-v1/tasks/27-ci-flake-fix.md`, `testdata/ste-baseline.json` (one new entry), and the evidence folder `spec/hive-v1/evidence/27-ci-flake-fix/`. `.github/workflows/ci.yml` is unchanged. |
| V4 | gate | pass | `gofmt -l` prints nothing. `go vet ./...` exits 0. `golangci-lint run ./...` prints `0 issues.` and exits 0. `go test -race ./...` exits 0. |
| V5 | R3 | pass | PR #48 is open with base `dev` and head `1b4a76e5`, the same as `origin/task/27-ci-flake-fix`. `gh pr checks 48` shows `gate pass` for run `39320462539`. `gh run view` shows `pull_request`, `success`, headSha `1b4a76e5ee9ec31550870de7ac2a2ac103404de8`. |

## Runs

### V2 before the fix (99ed618d)

Output is filtered to the `ok`, `FAIL`, `---`, and `WARNING` lines and counted with `uniq -c`.

```text
===  go test -race -count=50 -cpu=1,2 ./internal/tui/ -run TestRunHomeReportsInputFailure
      1 --- FAIL: TestRunHomeReportsInputFailure (0.00s)
      2 FAIL
      1 FAIL	github.com/Harrison-Blair/hive/internal/tui	0.149s
      1 WARNING: DATA RACE
exit 1
=== taskset -c 0 go test -race -count=50 -cpu=1,2 ./internal/tui/ -run TestRunHomeReportsInputFailure
      1 FAIL
      1 FAIL	github.com/Harrison-Blair/hive/internal/tui	0.105s
      1 WARNING: DATA RACE
exit 1
===  go test -race -count=50 -cpu=1,2 ./internal/lib/inbox/ -run TestParallelAppendsLoseNoRecord
      1 ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	7.059s
exit 0
=== taskset -c 0 go test -race -count=50 -cpu=1,2 ./internal/lib/inbox/ -run TestParallelAppendsLoseNoRecord
      1 ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	5.050s
exit 0
===  go test -race -count=50 -cpu=1,2 ./internal/supervise/ -run TestWatchFinishedMarkReachesStop
      4 --- FAIL: TestWatchFinishedMarkReachesStop (1.01s)
      2 FAIL
      1 FAIL	github.com/Harrison-Blair/hive/internal/supervise	5.939s
exit 1
=== taskset -c 0 go test -race -count=50 -cpu=1,2 ./internal/supervise/ -run TestWatchFinishedMarkReachesStop
     11 --- FAIL: TestWatchFinishedMarkReachesStop (1.01s)
     34 --- FAIL: TestWatchFinishedMarkReachesStop (1.02s)
     15 --- FAIL: TestWatchFinishedMarkReachesStop (1.03s)
      2 FAIL
      1 FAIL	github.com/Harrison-Blair/hive/internal/supervise	62.247s
exit 1
===  go test -race -count=50 -cpu=1,2 ./internal/lib/store/ -run TestGroupedReadSeesEqualVersions
      1 ok  	github.com/Harrison-Blair/hive/internal/lib/store	5.625s
exit 0
=== taskset -c 0 go test -race -count=50 -cpu=1,2 ./internal/lib/store/ -run TestGroupedReadSeesEqualVersions
      1 ok  	github.com/Harrison-Blair/hive/internal/lib/store	6.348s
exit 0
===  go test -race -count=50 -cpu=1,2 ./internal/supervise/ -run TestExitSignalSIGTERM
      1 ok  	github.com/Harrison-Blair/hive/internal/supervise	1.595s
exit 0
=== taskset -c 0 go test -race -count=50 -cpu=1,2 ./internal/supervise/ -run TestExitSignalSIGTERM
      1 ok  	github.com/Harrison-Blair/hive/internal/supervise	1.649s
exit 0
```

### V2 after the fix (1b4a76e5)

```text
===  go test -race -count=50 -cpu=1,2 ./internal/tui/ -run TestRunHomeReportsInputFailure
      1 ok  	github.com/Harrison-Blair/hive/internal/tui	1.172s
exit 0
=== taskset -c 0 go test -race -count=50 -cpu=1,2 ./internal/tui/ -run TestRunHomeReportsInputFailure
      1 ok  	github.com/Harrison-Blair/hive/internal/tui	1.220s
exit 0
===  go test -race -count=50 -cpu=1,2 ./internal/lib/inbox/ -run TestParallelAppendsLoseNoRecord
      1 ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	7.044s
exit 0
=== taskset -c 0 go test -race -count=50 -cpu=1,2 ./internal/lib/inbox/ -run TestParallelAppendsLoseNoRecord
      1 ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	5.236s
exit 0
===  go test -race -count=50 -cpu=1,2 ./internal/supervise/ -run TestWatchFinishedMarkReachesStop
      1 ok  	github.com/Harrison-Blair/hive/internal/supervise	2.732s
exit 0
=== taskset -c 0 go test -race -count=50 -cpu=1,2 ./internal/supervise/ -run TestWatchFinishedMarkReachesStop
      1 ok  	github.com/Harrison-Blair/hive/internal/supervise	2.707s
exit 0
===  go test -race -count=50 -cpu=1,2 ./internal/lib/store/ -run TestGroupedReadSeesEqualVersions
      1 ok  	github.com/Harrison-Blair/hive/internal/lib/store	6.024s
exit 0
=== taskset -c 0 go test -race -count=50 -cpu=1,2 ./internal/lib/store/ -run TestGroupedReadSeesEqualVersions
      1 ok  	github.com/Harrison-Blair/hive/internal/lib/store	6.348s
exit 0
===  go test -race -count=50 -cpu=1,2 ./internal/supervise/ -run TestExitSignalSIGTERM
      1 ok  	github.com/Harrison-Blair/hive/internal/supervise	1.645s
exit 0
=== taskset -c 0 go test -race -count=50 -cpu=1,2 ./internal/supervise/ -run TestExitSignalSIGTERM
      1 ok  	github.com/Harrison-Blair/hive/internal/supervise	1.635s
exit 0
```

### V5

```text
$ gh pr checks 48
gate	pass	58s	https://github.com/Harrison-Blair/hive/actions/runs/39320462539/job/118292867346
$ gh run view 39320462539 --json headSha,conclusion,event
{"conclusion":"success","event":"pull_request","headSha":"1b4a76e5ee9ec31550870de7ac2a2ac103404de8"}
```
