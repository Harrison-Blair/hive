# T10-agent-spawn-core FIX01

| Item | Value |
| --- | --- |
| Task | T10-agent-spawn-core |
| Pass | t10-fix-01 |
| Date | 2026-10-18 |
| Entry | `440049aecc5fde6b8a7df28f601d4aeaa365f8c4` |
| Tested source | `d55e9de6ec441313884f29e80188f5fd96cb22a4` |
| Source tree | `ef3c910346fd8dbf9b23cbd8297d4d1b71669eb7` |
| First Go source | `f225bfa9eb4541598edf6f298a743a4f19932736` |
| Base | `b4ab63ca1943c2c0491ae193392d8f7bdf3ea857` |
| Task blob | `5db0e34783fbfa1f73aca232e7e0f8008dff02fd` |
| Agent | worker, Pi, `opencode-go/deepseek-v4-flash:medium`, effective `high` |

The requested level is `medium`. The effective level is `high`.
The first independent source verdict is FAIL. No Root landing or CP2 acceptance exists.
The prior `VERIFIED.md` stays unchanged with its FAIL verdict.

## Fixes

### T10-O-F01 inspection evidence

`internal/agent/inspect.go` now writes `exited` only when the pane terminal
matches the record, the foreground set has one entry, that entry is the pane
shell PID, and the shell is alive.
A missing pane, an empty set, an unrelated process, or a shell in another
terminal becomes `unresolved`.
The existing restoration placeholder path stays `unresolved`.
The shared `PaneInfo` type did not change.
The native process reads stay outside the store lock.

Red: `TestFix01Inspect` fails with `error = agent exited`. The expected value is `agent unresolved`.
Green: the same test passes.
Mutation: restore `pane_not_found` to `exited`. The test fails on the assertion.
The exact restore passes.

### T10-O-F02 first session

`internal/agent/spawn.go` keeps the first present session from the successful
start reply. A later different session becomes `unresolved`.
A later missing reference also becomes `unresolved`.
An absent session with no recorded reference still passes.
The partial durable writes and the started-agent effect stay truthful.

Red: `TestFix02SpawnRetainsStartSession` gets `ok` and a different session.
Green: the same test passes.
Mutation: drop the start session. The test fails on the assertion.

### T10-O-F03 final ownership check

`internal/agent/stop.go` makes the stopped transition under one held store
handle. It compares the key, name, role, project, task, pane, generation,
status, spawned flag, claim, harness, terminal, workspace, tab, process PID,
process name, process start, and session reference.
A changed tuple refuses the close and keeps the current record.
A timestamp or another unrelated field can change.
All external reads and the close stay outside the lock.

Red: `TestFix03StopRefusesChangedIdentity` sends a close after the change.
Green: the same test passes with zero close requests.
Mutation: use the old status write. The test fails on the close assertion.

### T10-O-F04 package map

`internal/lib/herdr/doc.go` adds the `agent.go` and `pane.go` responsibilities.
The existing map lines stay. The change is comment only.

### T10-P-F01 local wire checks

`internal/lib/herdr/agent.go` and `pane.go` decode required presence, null
values, enums, arrays, and array items before a successful typed result.
The checks cover `focused`, `revision`, `agent_status`, `agents`, `argv`,
process PID, process name, and process argv.
Valid `false`, zero, empty arrays, extra fields, optional null, and optional
session absence still pass.
The shared `types.go` and the global request helper did not change.

Red: `TestFix01AgentWireValidation` returns nil errors for malformed replies.
Green: the same test and the valid-reply control pass.
Mutation: default a missing `focused`. The test fails on the assertion.

### T10-P-F02 timeout representability

`internal/agent/send.go` checks the requested milliseconds and the duration
plus five seconds before multiplication or logging.
The largest accepted value is 9223372031 seconds and sends 9223372031000 ms.
9223372032 and `MaxInt64` fail with zero prompt requests.
The default 600 seconds and the no-wait path do not change.

Red: `TestFix02WaitTimeoutBoundary` gets `ok` for invalid values.
Green: the same test passes.
Mutation: use the old nonpositive check. The test fails on the assertion.

## Fixture corrections

`TestSpawnWaitNotRunningInspectsExited` and
`TestSendAgentNotFoundInspectsExited` now script a same-terminal shell with the
real shell PID and foreground evidence for the intended `exited` case.
Their expected status and their full case sets did not change.
This pass removed no assertion, skip, or test.
New cases pin missing, empty, unrelated, and other-terminal evidence to
`unresolved`.

## Red, green, and sensitivity

The red replay ran the new tests against exact entry `440049ae` in an owned copy.
The raw replay is in `red-replay/`. Each failure is an assertion.
The green runs are in `green/`.
Five controlled mutations fail the matching assertion.
Each exact restore passed a byte comparison and the focused tests.
The F04 fix is a comment and needs no behavior mutation.
Captures: `.hive/tmp/coordinator/reports/t10-fix-01-captures/mutations/`.

## Gate and native

The full gate ran once at `d55e9de6` with a short `TMPDIR`.
`gofmt` printed nothing. Vet exited 0. Lint reported `0 issues.`.
`go test -race -count=1 ./...` passed.
The focused STE check passed with both glossaries and zero hard findings.

The native V4 flow used the exact binary SHA256
`e0f683e54e2374970cc132c08370e9dbe5ee8821852bc232cf1df79a67ea3aa9`.
The binary records revision `d55e9de6` and `vcs.modified=false`.
The blocks, the trust acceptance, the retry, the `pong` reply, the stop, the
post-close reads, and the preserved source workspace are in V4 and `native/`.

## Preservation and limits

Prior V2, V3, V4, V5, and SUMMARY bytes and hashes are in `prior/`.
No old report, capture, log, credential, worktree, or other agent changed.
The original stub snapshot `/tmp/t10red.lZnMvqUZ` is absent.
The original red logs remain assertion evidence. This pass does not replay the
missing stub source and does not call a new mutation the original red.
The native trust action can leave a Claude trust entry for the removed fixture.
Fresh independent source verification is PENDING.
