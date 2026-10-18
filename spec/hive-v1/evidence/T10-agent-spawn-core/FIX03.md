# T10-agent-spawn-core FIX03

| Item | Value |
| --- | --- |
| Task | T10-agent-spawn-core |
| Pass | `t10-fix-03` |
| Date | 2026-10-18 |
| Entry | `3a0f006e04edbdcc16f864f4afd6b7a4533b6e94` |
| Tested source | `5d8d3b528b68a2a5ee7652aa4a69129fd437b109` |
| Branch | `task/T10-agent-spawn-core` |
| Agent | worker, Pi, `opencode-go/deepseek-v4.1-flash`, requested `medium`, effective `high` |
| Verdict | independent verification PENDING |

This pass closes the open `T10-O-F02` gaps of `t10-verify-03`.
It changes `internal/agent/spawn.go` and adds `internal/agent/fix03_test.go`.
No other source changes.
No independent acceptance follows from this report.
No Root landing, CP2 pass, or CP4 pass follows from this report.

## Fix

`internal/agent/spawn.go` now writes the earliest start session before the next
external read. The pre-read write uses the existing held-handle mutation.
A failed write returns `partial` with the actual store error. The result keeps
the pre-write record, so it claims no unwritten session.
The pane read follows the successful write.
The pane topology write now precedes the pane session check. A pane reply whose
session is absent or different keeps `terminal_id`, `workspace_id`, `tab_id`,
`cwd`, and the earliest session. The result stays `unresolved`.
The wait observation now applies only to a successful typed reply. A wait
timeout is an error, so it supplies no session observation. A matching start,
pane, and `agent.get` session with a proved process identity keeps `partial`
with the record `live`.
`captureIdentity`, `mergeSessions`, and `sessionTrack` keep their old behavior.
The start, send, and stop behavior does not change.
Each socket wait and process wait stays outside the store lock.

## Cases

The literal first reference is `s`. The different reference is `other`.
Each case calls production `Spawn`, the real store, and a fake external peer.
No existing test was corrected. No existing test was weakened or removed.

| Case | Entry `3a0f006` | Tested source `5d8d3b5` |
| --- | --- | --- |
| Start `s`, record read at the `pane.get` request | no durable session | `s` is durable before the read |
| Start `s`, pane session absent | unresolved, topology lost | unresolved, `s` and topology kept |
| Start `s`, pane session `other` | unresolved, topology lost | unresolved, `s` and topology kept |
| Start, pane, and get `s`, owned PID and start, wait timeout | unresolved, false missing session | partial, record `live`, identity kept |
| Start `s`, session write fails at the store | read happened first | partial, store error, no read, no session claim |

The `TestSpawnWaitTimeoutWithOwnershipKeepsLivePartial` control stays a
required case. It passes at the tested source.
The no-session controls, the claim-replacement control, and the lock control
pass at the tested source.

## Test first

`internal/agent/fix03_test.go` is new. It holds five test functions:
`TestSpawnKeepsStartSessionBeforePaneRead`,
`TestSpawnKeepsPaneTopologyOnSessionRefusal` with the two subtests `absent` and
`differs`, `TestSpawnWaitTimeoutWithMatchingSessionKeepsLive`, and
`TestSpawnSessionStoreFailureKeepsTruth`.
The red run used the exact entry production `3a0f006`. The new test file was
the only added file. No new symbol was needed, so no empty-body step applied.

```sh
commit 3a0f006
go test -count=1 ./internal/agent/ -run 'TestSpawnKeepsStartSessionBeforePaneRead|TestSpawnKeepsPaneTopologyOnSessionRefusal|TestSpawnWaitTimeoutWithMatchingSessionKeepsLive|TestSpawnSessionStoreFailureKeepsTruth' -v
echo "exit $?"
```

```text
commit 3a0f006
=== RUN   TestSpawnKeepsStartSessionBeforePaneRead
    fix03_test.go:53: durable session before pane.get = ""; want "{\"source\":\"herdr:claude\",\"agent\":\"claude\",\"kind\":\"id\",\"value\":\"s\"}"
--- FAIL: TestSpawnKeepsStartSessionBeforePaneRead (0.04s)
=== RUN   TestSpawnKeepsPaneTopologyOnSessionRefusal
=== RUN   TestSpawnKeepsPaneTopologyOnSessionRefusal/absent
    fix03_test.go:85: record topology = terminal "" workspace "" tab "" cwd ""; want term_a w1 w1:t1 /tmp/fixture
    fix03_test.go:90: record topology = terminal "" workspace "" tab "" cwd ""; want term_a w1 w1:t1 /tmp/fixture
=== RUN   TestSpawnKeepsPaneTopologyOnSessionRefusal/differs
    fix03_test.go:85: record topology = terminal "" workspace "" tab "" cwd ""; want term_a w1 w1:t1 /tmp/fixture
    fix03_test.go:90: record topology = terminal "" workspace "" tab "" cwd ""; want term_a w1 w1:t1 /tmp/fixture
--- FAIL: TestSpawnKeepsPaneTopologyOnSessionRefusal (0.04s)
    --- FAIL: TestSpawnKeepsPaneTopologyOnSessionRefusal/absent (0.02s)
    --- FAIL: TestSpawnKeepsPaneTopologyOnSessionRefusal/differs (0.02s)
=== RUN   TestSpawnWaitTimeoutWithMatchingSessionKeepsLive
    fix03_test.go:132: error = agent process identity unresolved: the agent session is missing; want the readiness timeout
--- FAIL: TestSpawnWaitTimeoutWithMatchingSessionKeepsLive (0.02s)
=== RUN   TestSpawnSessionStoreFailureKeepsTruth
    fix03_test.go:193: pane.get ran 1 times; the session write must fail before the read
--- FAIL: TestSpawnSessionStoreFailureKeepsTruth (0.01s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.128s
FAIL
exit 1
```

Each failure is an assertion. No failure is a compile error or a panic.
Capture: `red-fix03.txt`.

The `TestSpawnSessionStoreFailureKeepsTruth` case fails at the entry because the
entry code sends `pane.get` before any start-session write. At the tested
source the failed session write ends the spawn before the read. The case also
asserts the store error, an empty result session, the unchanged record bytes,
the started effect, and no close request.

## Green run

```sh
commit 5d8d3b5
go test -count=1 ./internal/agent/ -run 'TestSpawnKeepsStartSessionBeforePaneRead|TestSpawnKeepsPaneTopologyOnSessionRefusal|TestSpawnWaitTimeoutWithMatchingSessionKeepsLive|TestSpawnSessionStoreFailureKeepsTruth' -v
echo "exit $?"
```

```text
commit 5d8d3b5
=== RUN   TestSpawnKeepsStartSessionBeforePaneRead
--- PASS: TestSpawnKeepsStartSessionBeforePaneRead (0.05s)
=== RUN   TestSpawnKeepsPaneTopologyOnSessionRefusal
=== RUN   TestSpawnKeepsPaneTopologyOnSessionRefusal/absent
=== RUN   TestSpawnKeepsPaneTopologyOnSessionRefusal/differs
--- PASS: TestSpawnKeepsPaneTopologyOnSessionRefusal (0.06s)
    --- PASS: TestSpawnKeepsPaneTopologyOnSessionRefusal/absent (0.03s)
    --- PASS: TestSpawnKeepsPaneTopologyOnSessionRefusal/differs (0.03s)
=== RUN   TestSpawnWaitTimeoutWithMatchingSessionKeepsLive
--- PASS: TestSpawnWaitTimeoutWithMatchingSessionKeepsLive (0.03s)
=== RUN   TestSpawnSessionStoreFailureKeepsTruth
--- PASS: TestSpawnSessionStoreFailureKeepsTruth (0.02s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/agent	0.178s
exit 0
```

Capture: `green-fix03.txt`.

## Regression suites

The existing spawn suite, the FIX01 suite, the FIX02 session suite, and the
full `internal/agent` package pass at the tested source.

```sh
commit 5d8d3b5
go test -race ./internal/agent/ -run TestSpawn -v
echo "exit $?"
```

```text
commit 5d8d3b5
=== RUN   TestSpawnKeepsStartSessionBeforePaneRead
--- PASS: TestSpawnKeepsStartSessionBeforePaneRead (0.06s)
=== RUN   TestSpawnKeepsPaneTopologyOnSessionRefusal
=== RUN   TestSpawnKeepsPaneTopologyOnSessionRefusal/absent
=== RUN   TestSpawnKeepsPaneTopologyOnSessionRefusal/differs
--- PASS: TestSpawnKeepsPaneTopologyOnSessionRefusal (0.08s)
    --- PASS: TestSpawnKeepsPaneTopologyOnSessionRefusal/absent (0.04s)
    --- PASS: TestSpawnKeepsPaneTopologyOnSessionRefusal/differs (0.04s)
=== RUN   TestSpawnWaitTimeoutWithMatchingSessionKeepsLive
--- PASS: TestSpawnWaitTimeoutWithMatchingSessionKeepsLive (0.04s)
=== RUN   TestSpawnSessionStoreFailureKeepsTruth
--- PASS: TestSpawnSessionStoreFailureKeepsTruth (0.02s)
=== RUN   TestSpawnIdleRecordsIdentity
--- PASS: TestSpawnIdleRecordsIdentity (0.06s)
=== RUN   TestSpawnRecordExistsBeforeStart
--- PASS: TestSpawnRecordExistsBeforeStart (0.06s)
=== RUN   TestSpawnDone
--- PASS: TestSpawnDone (0.06s)
=== RUN   TestSpawnBlockedKeepsLivePartial
--- PASS: TestSpawnBlockedKeepsLivePartial (0.07s)
=== RUN   TestSpawnWaitTimeoutWithOwnershipKeepsLivePartial
--- PASS: TestSpawnWaitTimeoutWithOwnershipKeepsLivePartial (0.05s)
=== RUN   TestSpawnWaitTimeoutWithoutOwnershipIsUnresolved
--- PASS: TestSpawnWaitTimeoutWithoutOwnershipIsUnresolved (0.05s)
=== RUN   TestSpawnStartServerErrorExitsRecord
--- PASS: TestSpawnStartServerErrorExitsRecord (0.03s)
=== RUN   TestSpawnStartTransportErrorIsUnresolved
--- PASS: TestSpawnStartTransportErrorIsUnresolved (0.03s)
=== RUN   TestSpawnWaitNotRunningInspectsExited
--- PASS: TestSpawnWaitNotRunningInspectsExited (0.06s)
=== RUN   TestSpawnAgentGetMissingNameIsOk
--- PASS: TestSpawnAgentGetMissingNameIsOk (0.06s)
=== RUN   TestSpawnProcessIdentityCases
=== RUN   TestSpawnProcessIdentityCases/absent
=== RUN   TestSpawnProcessIdentityCases/multiple
=== RUN   TestSpawnProcessIdentityCases/shell
=== RUN   TestSpawnProcessIdentityCases/unknown_name
=== RUN   TestSpawnProcessIdentityCases/wrapper_only
--- PASS: TestSpawnProcessIdentityCases (0.23s)
    --- PASS: TestSpawnProcessIdentityCases/absent (0.05s)
    --- PASS: TestSpawnProcessIdentityCases/multiple (0.05s)
    --- PASS: TestSpawnProcessIdentityCases/shell (0.04s)
    --- PASS: TestSpawnProcessIdentityCases/unknown_name (0.04s)
    --- PASS: TestSpawnProcessIdentityCases/wrapper_only (0.05s)
=== RUN   TestSpawnUnreadableProcessIsUnresolved
--- PASS: TestSpawnUnreadableProcessIsUnresolved (0.05s)
=== RUN   TestSpawnSessionRules
=== RUN   TestSpawnSessionRules/absent_passes
=== RUN   TestSpawnSessionRules/present_records
=== RUN   TestSpawnSessionRules/different_sources_fail
--- PASS: TestSpawnSessionRules (0.18s)
    --- PASS: TestSpawnSessionRules/absent_passes (0.06s)
    --- PASS: TestSpawnSessionRules/present_records (0.07s)
    --- PASS: TestSpawnSessionRules/different_sources_fail (0.06s)
=== RUN   TestSpawnClaimReplacementDuringWaitPreventsUpdate
--- PASS: TestSpawnClaimReplacementDuringWaitPreventsUpdate (0.06s)
=== RUN   TestSpawnDoesNotHoldLockAcrossWait
--- PASS: TestSpawnDoesNotHoldLockAcrossWait (0.06s)
=== RUN   TestSpawnWorkerWithRealTask
--- PASS: TestSpawnWorkerWithRealTask (0.07s)
=== RUN   TestSpawnWithoutPaneFails
--- PASS: TestSpawnWithoutPaneFails (0.00s)
=== RUN   TestSpawnUnknownHarness
--- PASS: TestSpawnUnknownHarness (0.00s)
=== RUN   TestSpawnStoresNoRecordBeforeValidation
--- PASS: TestSpawnStoresNoRecordBeforeValidation (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/agent	(cached)
exit 0
```

Capture: `v2-spawn.txt`. The full package passes in the same gate:

```sh
commit 5d8d3b5
go test -race -count=1 ./internal/agent/
echo "exit $?"
```

```text
ok  	github.com/Harrison-Blair/hive/internal/agent	5.771s
exit 0
```

## Gate

One full gate ran at tested source `5d8d3b5` from the worktree root.

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run ./...
go test -race ./...
```

```text
commit 5d8d3b5
--- gofmt ---
gofmt exit 0
--- go vet ---
vet exit 0
--- golangci-lint ---
0 issues.
lint exit 0
--- go test -race ---
ok  	github.com/Harrison-Blair/hive/internal/agent	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/store	(cached)
ok  	github.com/Harrison-Blair/hive/internal/tui	(cached)
test exit 0
```

Capture: `gate.txt`. The format list is empty. Lint reports `0 issues.`.
The uncached `go test -race -count=1 ./...` also passed at the tested tree.

## Live transcript

V4.md holds the fresh native transcript. The tested binary is
`/tmp/t10fx03.WIPUhYHp/hive-1792342212`. Its SHA256 is
`d4ba3d50eed28383f2d2089e842945d74d8a1dbd9353c0221bcfd3b9b59ffb51`.
The build records `vcs.revision=5d8d3b5` and `vcs.modified=false`.
The installed model catalog proves `claude-haiku-5-5` for Claude Code
`2.1.302`. The owned pane showed `Haiku 5.5: medium`.
The first owned spawn returned the truthful trust-dialog partial.
Trust was accepted only in the owned fixture pane.
The fresh spawn returned `ok` and `live`. The send returned `done`.
A separate pane read showed `● pong`. The stop returned exit 0 with
`pane closed`. The post-close reads answered `pane_not_found`.
The last list showed the record `stopped`. Both owned tabs and child PIDs are
absent. The fixture is removed.

## Files changed

| Path | Change |
| --- | --- |
| `internal/agent/spawn.go` | Persist the start session before the pane read. Keep the pane topology on a session refusal. Observe a wait session only on a successful typed reply. |
| `internal/agent/fix03_test.go` | New. Five tests for the four literal cases. |
| `spec/hive-v1/evidence/T10-agent-spawn-core/V2.md` | Add the FIX03 red and green runs. |
| `spec/hive-v1/evidence/T10-agent-spawn-core/V4.md` | Replace with the FIX03 native transcript. |
| `spec/hive-v1/evidence/T10-agent-spawn-core/V5.md` | Replace with the FIX03 gate. |
| `spec/hive-v1/evidence/T10-agent-spawn-core/SUMMARY.md` | Replace with the FIX03 result. |
| `spec/hive-v1/evidence/T10-agent-spawn-core/FIX03.md` | New. This report. |

## Limits

No existing test encoded the buggy behavior, so this pass corrected no test.
`TestSpawnWaitTimeoutWithOwnershipKeepsLivePartial` stays a required case.
The fixture trust action can leave a Claude trust entry for the removed path.
This pass claims no Root landing, CP2 acceptance, or CP4 acceptance.
No independent acceptance exists for this fix.

END T10 FIX 03
