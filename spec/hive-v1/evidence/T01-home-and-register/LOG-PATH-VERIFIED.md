# T01 log path guard: independent verification

To: coordinator.

| Item | Value |
| --- | --- |
| Verdict | PASS for this bounded follow-up |
| Pass | `t01-log-guard-verify-01` |
| Date | 2026-10-18 |
| Role | Fresh independent verifier |
| Actual route | External pi, `openai-codex/gpt-6.1-sol:high` |
| Branch | `task/T01-log-path-guards` |
| Exact base | `ff5e8c44d9b7d1fb510a568e9a248d84cab0b906` |
| Test commit | `b775c963a406ef0762553b466bb4c6c3b07937fa` |
| Production commit | `d5639e11031fce412ef261f51afc920e0b7de7a3` |
| Tested worker and verdict parent | `9be9b41e7687212af1f314fd83e0d19599ed0959` |
| Initial local, origin tracking, direct remote | `9be9b41e7687212af1f314fd83e0d19599ed0959` |
| Verdict commit | The sole child that adds `LOG-PATH-VERIFIED.md`. The ignored report records its full SHA after the push. |
| Source08 archive | `e26d08bdda4eb3aad5986796c35edb7a7d8c8175` |
| Report | `.hive/tmp/coordinator/reports/t01-log-guard-verify-01.md` |
| Captures | `.hive/tmp/coordinator/reports/t01-log-guard-verify-01-captures/` |
| Proves | T01 AC5 and AC7 for the static logger guard and retry. Existing AC1 safeguards stay green. |

## Authority and freeze

I read the complete worker report through `END T01 LOG GUARD 01`.
I read all of `LOG-PATH-FOLLOWUP.md` and the original log tests.
I read the complete root header report through `END T08 HEADER VERIFY 01`.
I read the actual T01 task, prior static-home acceptance, current agent rules, and evidence rules.
I read the STE skill, both glossaries, and the untrusted-input reference.

The local worktree was clean before probes.
The named origin tracking reference and direct remote matched the local tip.
The branch had no configured upstream. I checked the named tracking reference without a configuration change.
The actual `PI_PROVIDER`, `PI_MODEL`, and `PI_REASONING_LEVEL` values matched the required route.
I started no agent or model child.

I froze SHA256 values for all 261 tracked files before probes.
All 261 original file hashes stayed exact after probes and before the verdict commit.
The worker report and old evidence stayed unchanged.
Copies held all source probes and temporary mutations.
No actual source, test, baseline, dependency, task, tracker, or handoff changed.

## Complete three-path map

I inspected the complete exact-base-to-worker diff and every changed line.
The range contains exactly three paths, not two:

| Path | Accounting |
| --- | --- |
| `internal/lib/log/log.go` | 52 added lines and 4 removed lines. Production guard and rotation order. |
| `internal/lib/log/log_path_guard_test.go` | New file. 464 added lines. |
| `spec/hive-v1/evidence/T01-home-and-register/LOG-PATH-FOLLOWUP.md` | New file. 230 added lines. |

The worker report's two-code-path sentence is imprecise for the full range.
Two paths contain Go code. The third path contains evidence.
I preserved the original report.
This verdict adds only the new `LOG-PATH-VERIFIED.md` after that range.
It does not reopen all landed T01 source or accept T08.

## Source decisions and criteria

| Criterion | Result and evidence |
| --- | --- |
| Native directory components | PASS. `Lstat` checks `.hive` and `.hive/logs` before creation or daily open. Links, files, FIFOs, and sockets refuse explicitly. |
| Daily types | PASS. Valid links, dangling links, link chains, FIFOs, sockets, empty directories, and nonempty directories refuse. |
| Unchanged refused paths | PASS. Independent no-follow snapshots compare paths, modes, sizes, link targets, and regular-file SHA256 values. |
| No external write | PASS. Target sentinels stay exact. Missing link targets stay absent. Linked directories gain no file. |
| Bounded FIFO refusal | PASS. Fixed calls settle without a reader. Broken calls reach literal time bounds and settle through owned cooperating readers. |
| Missing valid paths | PASS. Existing creation controls retain directory mode 0755 and daily mode 0644. |
| Whole-home alias | PASS. A symbolic home root remains valid. Native `.hive` and `logs` below it remain required. |
| Regular append | PASS. Existing bytes remain a prefix. One accepted event adds one JSON line. |
| Local date and fields | PASS. Original tests and independent literal date/field checks pass. |
| Body cap | PASS. Literal 4096-byte and 4097-byte controls retain the cap and `truncated` behavior. |
| Shared handler state | PASS. `WithAttrs` and `WithGroup` retain one `outputState` pointer and mutex. Concurrent lines remain complete and unique. |
| First open and each rotation | PASS. `Handle` checks the prospective path whenever no file is open or the local date changes. |
| Failed rotation | PASS. Static refusal and native permission failure leave the old file open. Retry writes the rejected event exactly once. |
| Narrow design | PASS. One private guard uses two small private helpers. No public API, framework, dependency, deletion, or unrelated refactor appears. |

The only logger file open is `log.go:77`.
It retains `O_APPEND|O_CREATE|O_WRONLY`, mode 0644, and one write for each encoded event.
The new file opens before the old file closes.
Guard failure and open failure leave both the old file pointer and old date intact.
The mutex covers that state and each write.
The existing body, field, group, attribute, and encoding logic did not change.

The guard checks static paths. It does not check an already-open file on each same-day event.
An already-open regular file remains the write destination until rotation or close.
`NewWriter` still creates directories, not the daily file. The first `Handle` opens that file.
Constructor checks alone cannot prove later rotation safety. Separate probes test those later checks.

## Independent probes and sensitivity

The exact `b775c963` archive contains the old `log.go` and the new worker tests.
It compiles and fails on behavior assertions.
The unchanged frozen worker copy passes the same selection.
The old FIFO rotation blocks for two seconds before its owned reader releases it.
No stuck helper remains after the test process returns.

The independent matrix has 57 cases across three stages:

1. Constructor refusal.
2. First-open refusal after construction.
3. Rotation refusal after a successful prior-day event.

Each stage checks both directory components and the daily path.
Rotation probes check the old handle with `Stat` and a successful old-day write.
They remove only their owned obstruction and restore their owned native directory.
They retry the same rejected event on the new date.
Literal message lists prove no accepted-event loss, no duplicate retry, and no write through the rejected path.
A separate mode-0500 native directory causes a real `OpenFile` permission failure.
Its old-day handle remains usable, and its retry succeeds after mode restoration.
The concurrent probe writes 100 unique grouped events and one body event.

Five focused `log.go` mutations each compile and produce assertion-red status 1:

| Mutation | Literal failure |
| --- | --- |
| Disable constructor guard | Constructor accepts static obstructions or loses the explicit type error. |
| Disable first-open/rotation guard | Links accept writes. FIFO calls exceed 500 ms in the independent probe and two seconds in the worker probe. |
| Permit daily symbolic links | The logger accepts existing and dangling daily links. |
| Permit linked directory components | Constructor and rotation accept linked directories. |
| Close the old file before checking/opening | `closed stale handle` and `file already closed` assertions fail. |

Each mutation has its exact patch, red output, and restored-green output in captures.
The restored full log suite passes.
Every original file in the fixed copy matches the frozen archive after restoration.
The production `log.go` SHA256 is `cedbe1e03c5ba8c4df2f919e13f750edf2ab0c00baae6187f38a076613ceb90e`.

The first independent fixture used long test names in Unix socket paths.
Three socket setups failed with `invalid argument` before a logger call.
This was a probe setup error, not a source finding.
The corrected probe uses short owned `iv-` roots under the same outside-home `TMPDIR`.
All 57 cases then pass. Both probe versions and the initial output remain in captures.

## Production call-site audit

I searched all production imports and uses of the shared logger.
I read each of the eight production constructor call sites and its error path.

| Call site | Actual constructor-error policy |
| --- | --- |
| `cmd/common/common.go:57` | Render a failed envelope with phase `log` and status 1 before the command body. |
| `cmd/doctor/doctor.go:20` | Return a failed envelope with phase `log`. The outer common path also checks logger construction. |
| `cmd/inbox/inbox.go:92` | Return from the logging helper. Do not propagate the logger error. |
| `internal/agent/register.go:458` | Return from the logging helper. Do not propagate the logger error. |
| `internal/lib/inbox/inbox.go:406` | Return from the logging helper. Do not propagate the logger error. |
| `internal/task/log.go:21` | Return a nil logger and a no-op closer. Every event method guards nil. |
| `cmd/memory/write/write.go:24` | Ignore the constructor error and pass nil. Memory attribute, effect, and error helpers guard nil. |
| `cmd/handoff/write/write.go:24` | Ignore the constructor error and pass nil. Handoff attribute, effect, and error helpers guard nil. |

The worker table's phrase `Returns the error` is inaccurate for the internal logging helpers.
Those helpers return without a log event. They do not return the error to their operation caller.
I preserved that historical table.

A new owned probe passes the actual refused nil logger into memory and handoff operations.
Their success paths and error paths settle without a panic. The external target stays exact.
A static invalid path at CLI start fails in `common.Run` before either command callback.
I found no newly introduced caller regression.
Existing best-effort helpers can still omit an event when logger construction fails.
`slog` also discards later `Handle` errors.
This task adds no general logging-error framework and does not promise caller-level propagation.
No new stdin, stdout, socket, or Git wait appears under a state-store lock.

## Frozen source08 H03 overlay

I used the retained exact `header_verify_test.go` and `logdir_verify_test.go` probes.
Their bytes and SHA256 values remain in captures.
The owned source08 archive starts at the exact frozen commit in the table above.
Its original logger equals the T01 base logger.
Only `internal/lib/log/log.go` changes among its original tracked files.
The two probe files are additional verifier files, not production changes.
No actual T08 worktree or reference changed.

| Exact retained probe | Old logger | Fixed logger |
| --- | --- | --- |
| `TestHVLoggingBadDailyLinkNoWrite` | FAIL. The 37-byte target grows. | PASS. The target stays exactly 37 bytes. |
| `TestHVLoggingBadDailyFIFOBounded` | FAIL. Construction blocks for two seconds until the owned reader attaches. | PASS. Construction settles with an error and no reader. |
| `TestHVLoggingLinkedLogDirectoryError` | FAIL. The view lacks an error and the external directory gains one file. | PASS. The view contains an explicit native-directory error and the external directory stays empty. |

These are constructor-only static H03 probes against one frozen source08 tree.
The observed view text is not acceptance of T08-local error handling or later-path reporting.
This pass accepts only the shared logger correction for those exact cases.
It does not accept all H03, H01, H02, H04, T08, or a checkpoint.
The actual T08 worktree can contain a newer evidence-only verdict. This pass does not inspect it.

## Runs

All actual-root commands use `/tmp/lgv-hwb5yezm` as `TMPDIR` through child environments.
No marked home exists at that path or above it.
The inherited environment remains unchanged. Its `TMPDIR` stays unset.
All race runs use `-count=1`.
Raw commands, outputs, statuses, probes, patches, synthetic logs, and hashes remain in the unique captures.

The following blocks contain exact retained output.

### Exact test-first red copy at b775c963

```text
$ go test -race -count=1 -v ./internal/lib/log -run TestStaticLogPath
=== RUN   TestStaticLogPathRejectsLinkedHiveDirectory
    log_path_guard_test.go:54: NewWriter accepted a linked .hive directory; external entries = [logs]
--- FAIL: TestStaticLogPathRejectsLinkedHiveDirectory (0.00s)
=== RUN   TestStaticLogPathRejectsLinkedLogsDirectory
    log_path_guard_test.go:82: NewWriter accepted a linked .hive/logs directory; external entries = [2026-01-02.jsonl]
--- FAIL: TestStaticLogPathRejectsLinkedLogsDirectory (0.00s)
=== RUN   TestStaticLogPathRejectsNonRegularDailyPaths
=== RUN   TestStaticLogPathRejectsNonRegularDailyPaths/valid_symlink
    log_path_guard_test.go:166: NewWriter accepted a nonregular daily valid symlink path
=== RUN   TestStaticLogPathRejectsNonRegularDailyPaths/dangling_symlink
    log_path_guard_test.go:166: NewWriter accepted a nonregular daily dangling symlink path
=== RUN   TestStaticLogPathRejectsNonRegularDailyPaths/fifo
    log_path_guard_test.go:166: NewWriter accepted a nonregular daily fifo path
=== RUN   TestStaticLogPathRejectsNonRegularDailyPaths/directory
    log_path_guard_test.go:166: NewWriter accepted a nonregular daily directory path
--- FAIL: TestStaticLogPathRejectsNonRegularDailyPaths (0.00s)
    --- FAIL: TestStaticLogPathRejectsNonRegularDailyPaths/valid_symlink (0.00s)
    --- FAIL: TestStaticLogPathRejectsNonRegularDailyPaths/dangling_symlink (0.00s)
    --- FAIL: TestStaticLogPathRejectsNonRegularDailyPaths/fifo (0.00s)
    --- FAIL: TestStaticLogPathRejectsNonRegularDailyPaths/directory (0.00s)
=== RUN   TestStaticLogPathRejectsDailySocket
    log_path_guard_test.go:207: NewWriter accepted a socket daily path
--- FAIL: TestStaticLogPathRejectsDailySocket (0.00s)
=== RUN   TestStaticLogPathRejectsDailyFIFOWithoutBlocking
    log_path_guard_test.go:230: NewWriter accepted a FIFO daily path without an error
--- FAIL: TestStaticLogPathRejectsDailyFIFOWithoutBlocking (0.00s)
=== RUN   TestStaticLogPathRejectsRotationToFIFOWithoutBlocking
    log_path_guard_test.go:282: Handle blocked more than 2 seconds on a static FIFO rotation path
--- FAIL: TestStaticLogPathRejectsRotationToFIFOWithoutBlocking (2.00s)
=== RUN   TestStaticLogPathRejectsRotationToDailyLinkAndRetries
    log_path_guard_test.go:327: Handle accepted a symlink rotation path and wrote through it
--- FAIL: TestStaticLogPathRejectsRotationToDailyLinkAndRetries (0.00s)
=== RUN   TestStaticLogPathRejectsLinkedLogsDirectoryOnRotation
    log_path_guard_test.go:388: Handle accepted a linked logs directory on rotation
--- FAIL: TestStaticLogPathRejectsLinkedLogsDirectoryOnRotation (0.00s)
=== RUN   TestStaticLogPathAllowsValidCreationAppendAndWholeHomeAlias
--- PASS: TestStaticLogPathAllowsValidCreationAppendAndWholeHomeAlias (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/log	2.007s
FAIL
exit 1
```

### Frozen worker green copy at 9be9b41e

```text
$ go test -race -count=1 -v ./internal/lib/log -run TestStaticLogPath
=== RUN   TestStaticLogPathRejectsLinkedHiveDirectory
--- PASS: TestStaticLogPathRejectsLinkedHiveDirectory (0.00s)
=== RUN   TestStaticLogPathRejectsLinkedLogsDirectory
--- PASS: TestStaticLogPathRejectsLinkedLogsDirectory (0.00s)
=== RUN   TestStaticLogPathRejectsNonRegularDailyPaths
=== RUN   TestStaticLogPathRejectsNonRegularDailyPaths/valid_symlink
=== RUN   TestStaticLogPathRejectsNonRegularDailyPaths/dangling_symlink
=== RUN   TestStaticLogPathRejectsNonRegularDailyPaths/fifo
=== RUN   TestStaticLogPathRejectsNonRegularDailyPaths/directory
--- PASS: TestStaticLogPathRejectsNonRegularDailyPaths (0.00s)
    --- PASS: TestStaticLogPathRejectsNonRegularDailyPaths/valid_symlink (0.00s)
    --- PASS: TestStaticLogPathRejectsNonRegularDailyPaths/dangling_symlink (0.00s)
    --- PASS: TestStaticLogPathRejectsNonRegularDailyPaths/fifo (0.00s)
    --- PASS: TestStaticLogPathRejectsNonRegularDailyPaths/directory (0.00s)
=== RUN   TestStaticLogPathRejectsDailySocket
--- PASS: TestStaticLogPathRejectsDailySocket (0.00s)
=== RUN   TestStaticLogPathRejectsDailyFIFOWithoutBlocking
--- PASS: TestStaticLogPathRejectsDailyFIFOWithoutBlocking (0.00s)
=== RUN   TestStaticLogPathRejectsRotationToFIFOWithoutBlocking
--- PASS: TestStaticLogPathRejectsRotationToFIFOWithoutBlocking (0.00s)
=== RUN   TestStaticLogPathRejectsRotationToDailyLinkAndRetries
--- PASS: TestStaticLogPathRejectsRotationToDailyLinkAndRetries (0.00s)
=== RUN   TestStaticLogPathRejectsLinkedLogsDirectoryOnRotation
--- PASS: TestStaticLogPathRejectsLinkedLogsDirectoryOnRotation (0.00s)
=== RUN   TestStaticLogPathAllowsValidCreationAppendAndWholeHomeAlias
--- PASS: TestStaticLogPathAllowsValidCreationAppendAndWholeHomeAlias (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.006s
exit 0
```

### Retained H03 old logger overlay

```text
$ go test -race -count=1 -v ./internal/tui -run 'TestHVLogging(BadDailyLinkNoWrite|BadDailyFIFOBounded|LinkedLogDirectoryError)'
=== RUN   TestHVLoggingBadDailyLinkNoWrite
    header_verify_test.go:368: target before=37 bytes after=239 bytes
    header_verify_test.go:370: error logger followed rejected link and changed synthetic target: synthetic target must stay unchanged
        {"component":"tui","context":"view","error":"log is not a regular file: /tmp/lgv-hwb5yezm/h-3536837433/.hive/logs/2042-06-07.jsonl","level":"DEBUG","msg":"tui error","time":"2042-06-07T12:00:00-04:00"}
--- FAIL: TestHVLoggingBadDailyLinkNoWrite (0.00s)
=== RUN   TestHVLoggingBadDailyFIFOBounded
    header_verify_test.go:405: New blocked more than 2 seconds on rejected FIFO until an owned reader opened it
--- FAIL: TestHVLoggingBadDailyFIFOBounded (2.00s)
=== RUN   TestHVLoggingLinkedLogDirectoryError
    logdir_verify_test.go:33: View="\x1b[1mhive /tmp/lgv-hwb5yezm/h-3074483389  supervisor: down  agents: 0  inbox: 0 unread\x1b[m\nno events today                                                                 \n                                                                                \n                                                                                \n                                                                                \n                                                                                \n                                                                                \n                                                                                \n                                                                                \n                                                                                \n                                                                                \n                                                                                \n                                                                                \n                                                                                \n                                                                                \n                                                                                \n                                                                                \n                                                                                \n                                                                                \n                                                                                \n                                                                                \n                                                                                \n                                                                                \nlevel: debug  text: -  follow: on"; synthetic link target files=1
    logdir_verify_test.go:35: linked log directory produced a successful view and an external-target write, want explicit error and no write
--- FAIL: TestHVLoggingLinkedLogDirectoryError (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui	2.009s
FAIL
exit 1
```

### Retained H03 fixed logger overlay

```text
$ go test -race -count=1 -v ./internal/tui -run 'TestHVLogging(BadDailyLinkNoWrite|BadDailyFIFOBounded|LinkedLogDirectoryError)'
=== RUN   TestHVLoggingBadDailyLinkNoWrite
    header_verify_test.go:368: target before=37 bytes after=37 bytes
--- PASS: TestHVLoggingBadDailyLinkNoWrite (0.00s)
=== RUN   TestHVLoggingBadDailyFIFOBounded
--- PASS: TestHVLoggingBadDailyFIFOBounded (0.00s)
=== RUN   TestHVLoggingLinkedLogDirectoryError
    logdir_verify_test.go:33: View="\x1b[1mhive /tmp/lgv-hwb5yezm/h-2638116175  supervisor: down  agents: 0  inbox: 0 unread\x1b[m\nerror: log: log directory is not a native directory: /tmp/lgv-hwb5yezm/h-2638116175/.hive/logs\nlevel: debug  text: -  follow: on"; synthetic link target files=0
--- PASS: TestHVLoggingLinkedLogDirectoryError (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/tui	1.007s
exit 0
```

### Actual-root original log safeguards

```text
$ go test -race -count=1 -v ./internal/lib/log
=== RUN   TestStaticLogPathRejectsLinkedHiveDirectory
--- PASS: TestStaticLogPathRejectsLinkedHiveDirectory (0.00s)
=== RUN   TestStaticLogPathRejectsLinkedLogsDirectory
--- PASS: TestStaticLogPathRejectsLinkedLogsDirectory (0.00s)
=== RUN   TestStaticLogPathRejectsNonRegularDailyPaths
=== RUN   TestStaticLogPathRejectsNonRegularDailyPaths/valid_symlink
=== RUN   TestStaticLogPathRejectsNonRegularDailyPaths/dangling_symlink
=== RUN   TestStaticLogPathRejectsNonRegularDailyPaths/fifo
=== RUN   TestStaticLogPathRejectsNonRegularDailyPaths/directory
--- PASS: TestStaticLogPathRejectsNonRegularDailyPaths (0.00s)
    --- PASS: TestStaticLogPathRejectsNonRegularDailyPaths/valid_symlink (0.00s)
    --- PASS: TestStaticLogPathRejectsNonRegularDailyPaths/dangling_symlink (0.00s)
    --- PASS: TestStaticLogPathRejectsNonRegularDailyPaths/fifo (0.00s)
    --- PASS: TestStaticLogPathRejectsNonRegularDailyPaths/directory (0.00s)
=== RUN   TestStaticLogPathRejectsDailySocket
--- PASS: TestStaticLogPathRejectsDailySocket (0.00s)
=== RUN   TestStaticLogPathRejectsDailyFIFOWithoutBlocking
--- PASS: TestStaticLogPathRejectsDailyFIFOWithoutBlocking (0.00s)
=== RUN   TestStaticLogPathRejectsRotationToFIFOWithoutBlocking
--- PASS: TestStaticLogPathRejectsRotationToFIFOWithoutBlocking (0.00s)
=== RUN   TestStaticLogPathRejectsRotationToDailyLinkAndRetries
--- PASS: TestStaticLogPathRejectsRotationToDailyLinkAndRetries (0.00s)
=== RUN   TestStaticLogPathRejectsLinkedLogsDirectoryOnRotation
--- PASS: TestStaticLogPathRejectsLinkedLogsDirectoryOnRotation (0.00s)
=== RUN   TestStaticLogPathAllowsValidCreationAppendAndWholeHomeAlias
--- PASS: TestStaticLogPathAllowsValidCreationAppendAndWholeHomeAlias (0.00s)
=== RUN   TestBodyCapsAt4096BytesAndMarksTruncated
--- PASS: TestBodyCapsAt4096BytesAndMarksTruncated (0.00s)
=== RUN   TestBodyAtLimitDoesNotMarkTruncated
--- PASS: TestBodyAtLimitDoesNotMarkTruncated (0.00s)
=== RUN   TestConcurrentLogEventsAreWholeJSONLines
--- PASS: TestConcurrentLogEventsAreWholeJSONLines (0.03s)
=== RUN   TestOpenAppendsToDailyFile
--- PASS: TestOpenAppendsToDailyFile (0.00s)
=== RUN   TestWriterRotatesAtLocalMidnight
--- PASS: TestWriterRotatesAtLocalMidnight (0.00s)
=== RUN   TestLogFileNamesIncludeTheLocalDate
--- PASS: TestLogFileNamesIncludeTheLocalDate (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.036s
exit 0
```

### Actual-root resolver safeguards

```text
$ go test -race -count=1 -v ./internal/lib/home -run TestResolve
=== RUN   TestResolveFlagWins
--- PASS: TestResolveFlagWins (0.00s)
=== RUN   TestResolveEnvWinsOverAncestor
--- PASS: TestResolveEnvWinsOverAncestor (0.00s)
=== RUN   TestResolveOutermostMarkerWins
--- PASS: TestResolveOutermostMarkerWins (0.00s)
=== RUN   TestResolveFlagSkipsConflictCheck
--- PASS: TestResolveFlagSkipsConflictCheck (0.00s)
=== RUN   TestResolveNamedPathWithoutMarker
--- PASS: TestResolveNamedPathWithoutMarker (0.00s)
=== RUN   TestResolveEnvPathWithoutMarker
--- PASS: TestResolveEnvPathWithoutMarker (0.00s)
=== RUN   TestResolveNoMarker
--- PASS: TestResolveNoMarker (0.00s)
=== RUN   TestResolveEmptyHomeFlagIsUsageError
--- PASS: TestResolveEmptyHomeFlagIsUsageError (0.00s)
=== RUN   TestResolveEmptyEnvIsUnset
--- PASS: TestResolveEmptyEnvIsUnset (0.00s)
=== RUN   TestResolveRelativeHomePathsAgainstWorkingDirectory
--- PASS: TestResolveRelativeHomePathsAgainstWorkingDirectory (0.00s)
=== RUN   TestResolveConflictComparesResolvedPaths
--- PASS: TestResolveConflictComparesResolvedPaths (0.00s)
=== RUN   TestResolveDoesNotTrimHomeValues
--- PASS: TestResolveDoesNotTrimHomeValues (0.00s)
=== RUN   TestResolveDoesNotTrimEnvHomeValue
--- PASS: TestResolveDoesNotTrimEnvHomeValue (0.00s)
=== RUN   TestResolveMarkerMustBeRegularFile
--- PASS: TestResolveMarkerMustBeRegularFile (0.00s)
=== RUN   TestResolveReturnsAbsolutePathWithoutSymlinks
--- PASS: TestResolveReturnsAbsolutePathWithoutSymlinks (0.00s)
=== RUN   TestResolveRejectsLinkedHiveDirectory
--- PASS: TestResolveRejectsLinkedHiveDirectory (0.00s)
=== RUN   TestResolveRejectsNonDirectoryHive
--- PASS: TestResolveRejectsNonDirectoryHive (0.00s)
=== RUN   TestResolveAncestorSkipsLinkedHiveMarker
--- PASS: TestResolveAncestorSkipsLinkedHiveMarker (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.006s
exit 0
```

### Actual-root static-home command safeguards

```text
$ go test -race -count=1 -v ./cmd -run TestProjectList|TestRootMissingHome|TestRootProjectListJSONEnvelope|TestResolveEmptyHomeFlagHasUsageStatus
=== RUN   TestRootProjectListJSONEnvelope
--- PASS: TestRootProjectListJSONEnvelope (0.00s)
=== RUN   TestRootMissingHomeUsesOneErrorLine
--- PASS: TestRootMissingHomeUsesOneErrorLine (0.00s)
=== RUN   TestRootMissingHomeJSONEnvelope
--- PASS: TestRootMissingHomeJSONEnvelope (0.00s)
=== RUN   TestResolveEmptyHomeFlagHasUsageStatus
--- PASS: TestResolveEmptyHomeFlagHasUsageStatus (0.00s)
=== RUN   TestProjectListRejectsLinkedHiveHome
=== RUN   TestProjectListRejectsLinkedHiveHome/flag_text
=== RUN   TestProjectListRejectsLinkedHiveHome/flag_json
=== RUN   TestProjectListRejectsLinkedHiveHome/environment_text
=== RUN   TestProjectListRejectsLinkedHiveHome/cwd_discovery_text
--- PASS: TestProjectListRejectsLinkedHiveHome (0.00s)
    --- PASS: TestProjectListRejectsLinkedHiveHome/flag_text (0.00s)
    --- PASS: TestProjectListRejectsLinkedHiveHome/flag_json (0.00s)
    --- PASS: TestProjectListRejectsLinkedHiveHome/environment_text (0.00s)
    --- PASS: TestProjectListRejectsLinkedHiveHome/cwd_discovery_text (0.00s)
=== RUN   TestProjectListAcceptsRegularHiveHome
--- PASS: TestProjectListAcceptsRegularHiveHome (0.00s)
=== RUN   TestProjectListAcceptsSymlinkedHomeRoot
--- PASS: TestProjectListAcceptsSymlinkedHomeRoot (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/cmd	1.012s
exit 0
```

### Independent matrix and literal controls

```text
$ go test -race -count=1 -v ./internal/lib/log -run TestIV
=== RUN   TestIVStaticMatrix
=== RUN   TestIVStaticMatrix/constructor/.hive/file
=== RUN   TestIVStaticMatrix/constructor/.hive/link
=== RUN   TestIVStaticMatrix/constructor/.hive/dangling
=== RUN   TestIVStaticMatrix/constructor/.hive/chain
=== RUN   TestIVStaticMatrix/constructor/.hive/fifo
=== RUN   TestIVStaticMatrix/constructor/.hive/socket
=== RUN   TestIVStaticMatrix/constructor/.hive-logs/file
=== RUN   TestIVStaticMatrix/constructor/.hive-logs/link
=== RUN   TestIVStaticMatrix/constructor/.hive-logs/dangling
=== RUN   TestIVStaticMatrix/constructor/.hive-logs/chain
=== RUN   TestIVStaticMatrix/constructor/.hive-logs/fifo
=== RUN   TestIVStaticMatrix/constructor/.hive-logs/socket
=== RUN   TestIVStaticMatrix/constructor/.hive-logs-2043-02-03.jsonl/directory
=== RUN   TestIVStaticMatrix/constructor/.hive-logs-2043-02-03.jsonl/nonempty-directory
=== RUN   TestIVStaticMatrix/constructor/.hive-logs-2043-02-03.jsonl/link
=== RUN   TestIVStaticMatrix/constructor/.hive-logs-2043-02-03.jsonl/dangling
=== RUN   TestIVStaticMatrix/constructor/.hive-logs-2043-02-03.jsonl/chain
=== RUN   TestIVStaticMatrix/constructor/.hive-logs-2043-02-03.jsonl/fifo
=== RUN   TestIVStaticMatrix/constructor/.hive-logs-2043-02-03.jsonl/socket
=== RUN   TestIVStaticMatrix/first-open/.hive/file
=== RUN   TestIVStaticMatrix/first-open/.hive/link
=== RUN   TestIVStaticMatrix/first-open/.hive/dangling
=== RUN   TestIVStaticMatrix/first-open/.hive/chain
=== RUN   TestIVStaticMatrix/first-open/.hive/fifo
=== RUN   TestIVStaticMatrix/first-open/.hive/socket
=== RUN   TestIVStaticMatrix/first-open/.hive-logs/file
=== RUN   TestIVStaticMatrix/first-open/.hive-logs/link
=== RUN   TestIVStaticMatrix/first-open/.hive-logs/dangling
=== RUN   TestIVStaticMatrix/first-open/.hive-logs/chain
=== RUN   TestIVStaticMatrix/first-open/.hive-logs/fifo
=== RUN   TestIVStaticMatrix/first-open/.hive-logs/socket
=== RUN   TestIVStaticMatrix/first-open/.hive-logs-2043-02-03.jsonl/directory
=== RUN   TestIVStaticMatrix/first-open/.hive-logs-2043-02-03.jsonl/nonempty-directory
=== RUN   TestIVStaticMatrix/first-open/.hive-logs-2043-02-03.jsonl/link
=== RUN   TestIVStaticMatrix/first-open/.hive-logs-2043-02-03.jsonl/dangling
=== RUN   TestIVStaticMatrix/first-open/.hive-logs-2043-02-03.jsonl/chain
=== RUN   TestIVStaticMatrix/first-open/.hive-logs-2043-02-03.jsonl/fifo
=== RUN   TestIVStaticMatrix/first-open/.hive-logs-2043-02-03.jsonl/socket
=== RUN   TestIVStaticMatrix/rotation/.hive/file
=== RUN   TestIVStaticMatrix/rotation/.hive/link
=== RUN   TestIVStaticMatrix/rotation/.hive/dangling
=== RUN   TestIVStaticMatrix/rotation/.hive/chain
=== RUN   TestIVStaticMatrix/rotation/.hive/fifo
=== RUN   TestIVStaticMatrix/rotation/.hive/socket
=== RUN   TestIVStaticMatrix/rotation/.hive-logs/file
=== RUN   TestIVStaticMatrix/rotation/.hive-logs/link
=== RUN   TestIVStaticMatrix/rotation/.hive-logs/dangling
=== RUN   TestIVStaticMatrix/rotation/.hive-logs/chain
=== RUN   TestIVStaticMatrix/rotation/.hive-logs/fifo
=== RUN   TestIVStaticMatrix/rotation/.hive-logs/socket
=== RUN   TestIVStaticMatrix/rotation/.hive-logs-2043-02-03.jsonl/directory
=== RUN   TestIVStaticMatrix/rotation/.hive-logs-2043-02-03.jsonl/nonempty-directory
=== RUN   TestIVStaticMatrix/rotation/.hive-logs-2043-02-03.jsonl/link
=== RUN   TestIVStaticMatrix/rotation/.hive-logs-2043-02-03.jsonl/dangling
=== RUN   TestIVStaticMatrix/rotation/.hive-logs-2043-02-03.jsonl/chain
=== RUN   TestIVStaticMatrix/rotation/.hive-logs-2043-02-03.jsonl/fifo
=== RUN   TestIVStaticMatrix/rotation/.hive-logs-2043-02-03.jsonl/socket
--- PASS: TestIVStaticMatrix (0.05s)
    --- PASS: TestIVStaticMatrix/constructor/.hive/file (0.00s)
    --- PASS: TestIVStaticMatrix/constructor/.hive/link (0.00s)
    --- PASS: TestIVStaticMatrix/constructor/.hive/dangling (0.00s)
    --- PASS: TestIVStaticMatrix/constructor/.hive/chain (0.00s)
    --- PASS: TestIVStaticMatrix/constructor/.hive/fifo (0.00s)
    --- PASS: TestIVStaticMatrix/constructor/.hive/socket (0.00s)
    --- PASS: TestIVStaticMatrix/constructor/.hive-logs/file (0.00s)
    --- PASS: TestIVStaticMatrix/constructor/.hive-logs/link (0.00s)
    --- PASS: TestIVStaticMatrix/constructor/.hive-logs/dangling (0.00s)
    --- PASS: TestIVStaticMatrix/constructor/.hive-logs/chain (0.00s)
    --- PASS: TestIVStaticMatrix/constructor/.hive-logs/fifo (0.00s)
    --- PASS: TestIVStaticMatrix/constructor/.hive-logs/socket (0.00s)
    --- PASS: TestIVStaticMatrix/constructor/.hive-logs-2043-02-03.jsonl/directory (0.00s)
    --- PASS: TestIVStaticMatrix/constructor/.hive-logs-2043-02-03.jsonl/nonempty-directory (0.00s)
    --- PASS: TestIVStaticMatrix/constructor/.hive-logs-2043-02-03.jsonl/link (0.00s)
    --- PASS: TestIVStaticMatrix/constructor/.hive-logs-2043-02-03.jsonl/dangling (0.00s)
    --- PASS: TestIVStaticMatrix/constructor/.hive-logs-2043-02-03.jsonl/chain (0.00s)
    --- PASS: TestIVStaticMatrix/constructor/.hive-logs-2043-02-03.jsonl/fifo (0.00s)
    --- PASS: TestIVStaticMatrix/constructor/.hive-logs-2043-02-03.jsonl/socket (0.00s)
    --- PASS: TestIVStaticMatrix/first-open/.hive/file (0.00s)
    --- PASS: TestIVStaticMatrix/first-open/.hive/link (0.00s)
    --- PASS: TestIVStaticMatrix/first-open/.hive/dangling (0.00s)
    --- PASS: TestIVStaticMatrix/first-open/.hive/chain (0.00s)
    --- PASS: TestIVStaticMatrix/first-open/.hive/fifo (0.00s)
    --- PASS: TestIVStaticMatrix/first-open/.hive/socket (0.00s)
    --- PASS: TestIVStaticMatrix/first-open/.hive-logs/file (0.00s)
    --- PASS: TestIVStaticMatrix/first-open/.hive-logs/link (0.00s)
    --- PASS: TestIVStaticMatrix/first-open/.hive-logs/dangling (0.00s)
    --- PASS: TestIVStaticMatrix/first-open/.hive-logs/chain (0.00s)
    --- PASS: TestIVStaticMatrix/first-open/.hive-logs/fifo (0.00s)
    --- PASS: TestIVStaticMatrix/first-open/.hive-logs/socket (0.00s)
    --- PASS: TestIVStaticMatrix/first-open/.hive-logs-2043-02-03.jsonl/directory (0.00s)
    --- PASS: TestIVStaticMatrix/first-open/.hive-logs-2043-02-03.jsonl/nonempty-directory (0.00s)
    --- PASS: TestIVStaticMatrix/first-open/.hive-logs-2043-02-03.jsonl/link (0.00s)
    --- PASS: TestIVStaticMatrix/first-open/.hive-logs-2043-02-03.jsonl/dangling (0.00s)
    --- PASS: TestIVStaticMatrix/first-open/.hive-logs-2043-02-03.jsonl/chain (0.00s)
    --- PASS: TestIVStaticMatrix/first-open/.hive-logs-2043-02-03.jsonl/fifo (0.00s)
    --- PASS: TestIVStaticMatrix/first-open/.hive-logs-2043-02-03.jsonl/socket (0.00s)
    --- PASS: TestIVStaticMatrix/rotation/.hive/file (0.00s)
    --- PASS: TestIVStaticMatrix/rotation/.hive/link (0.00s)
    --- PASS: TestIVStaticMatrix/rotation/.hive/dangling (0.00s)
    --- PASS: TestIVStaticMatrix/rotation/.hive/chain (0.00s)
    --- PASS: TestIVStaticMatrix/rotation/.hive/fifo (0.00s)
    --- PASS: TestIVStaticMatrix/rotation/.hive/socket (0.00s)
    --- PASS: TestIVStaticMatrix/rotation/.hive-logs/file (0.00s)
    --- PASS: TestIVStaticMatrix/rotation/.hive-logs/link (0.00s)
    --- PASS: TestIVStaticMatrix/rotation/.hive-logs/dangling (0.00s)
    --- PASS: TestIVStaticMatrix/rotation/.hive-logs/chain (0.00s)
    --- PASS: TestIVStaticMatrix/rotation/.hive-logs/fifo (0.00s)
    --- PASS: TestIVStaticMatrix/rotation/.hive-logs/socket (0.00s)
    --- PASS: TestIVStaticMatrix/rotation/.hive-logs-2043-02-03.jsonl/directory (0.00s)
    --- PASS: TestIVStaticMatrix/rotation/.hive-logs-2043-02-03.jsonl/nonempty-directory (0.00s)
    --- PASS: TestIVStaticMatrix/rotation/.hive-logs-2043-02-03.jsonl/link (0.00s)
    --- PASS: TestIVStaticMatrix/rotation/.hive-logs-2043-02-03.jsonl/dangling (0.00s)
    --- PASS: TestIVStaticMatrix/rotation/.hive-logs-2043-02-03.jsonl/chain (0.00s)
    --- PASS: TestIVStaticMatrix/rotation/.hive-logs-2043-02-03.jsonl/fifo (0.00s)
    --- PASS: TestIVStaticMatrix/rotation/.hive-logs-2043-02-03.jsonl/socket (0.00s)
=== RUN   TestIVOpenFailureRetry
--- PASS: TestIVOpenFailureRetry (0.00s)
=== RUN   TestIVSharedStateAndLiteralFields
--- PASS: TestIVSharedStateAndLiteralFields (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.055s
exit 0
```

### Actual-root full uncached gate and STE

```text
commit 9be9b41e
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
exit 0

$ go vet ./...
exit 0

$ golangci-lint run --allow-serial-runners ./...
0 issues.
exit 0

$ go test -race -count=1 ./...
ok  	github.com/Harrison-Blair/hive	3.288s
ok  	github.com/Harrison-Blair/hive/cmd	1.063s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.227s
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.213s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.021s
ok  	github.com/Harrison-Blair/hive/internal/agent	1.269s
ok  	github.com/Harrison-Blair/hive/internal/doctor	3.040s
ok  	github.com/Harrison-Blair/hive/internal/handoff	1.137s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.018s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	6.506s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.020s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	1.227s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.060s
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/records	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/run	3.479s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.139s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	17.291s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.003s
ok  	github.com/Harrison-Blair/hive/internal/logs	5.511s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.154s
ok  	github.com/Harrison-Blair/hive/internal/project	1.011s
ok  	github.com/Harrison-Blair/hive/internal/task	1.160s
ok  	github.com/Harrison-Blair/hive/internal/tui	1.005s
exit 0

$ go test -run TestSTE -count=1 -v .
=== RUN   TestSTESemicolon
=== RUN   TestSTESemicolon/pass
=== RUN   TestSTESemicolon/fail
--- PASS: TestSTESemicolon (0.00s)
    --- PASS: TestSTESemicolon/pass (0.00s)
    --- PASS: TestSTESemicolon/fail (0.00s)
=== RUN   TestSTEPassiveVoice
=== RUN   TestSTEPassiveVoice/pass
=== RUN   TestSTEPassiveVoice/fail
--- PASS: TestSTEPassiveVoice (0.00s)
    --- PASS: TestSTEPassiveVoice/pass (0.00s)
    --- PASS: TestSTEPassiveVoice/fail (0.00s)
=== RUN   TestSTEPresentPerfect
=== RUN   TestSTEPresentPerfect/pass
=== RUN   TestSTEPresentPerfect/fail
=== RUN   TestSTEPresentPerfect/modal_perfect_stays_a_hedge
--- PASS: TestSTEPresentPerfect (0.00s)
    --- PASS: TestSTEPresentPerfect/pass (0.00s)
    --- PASS: TestSTEPresentPerfect/fail (0.00s)
    --- PASS: TestSTEPresentPerfect/modal_perfect_stays_a_hedge (0.00s)
=== RUN   TestSTELongSentence
=== RUN   TestSTELongSentence/pass
=== RUN   TestSTELongSentence/fail
--- PASS: TestSTELongSentence (0.00s)
    --- PASS: TestSTELongSentence/pass (0.00s)
    --- PASS: TestSTELongSentence/fail (0.00s)
=== RUN   TestSTEPhrasalVerb
=== RUN   TestSTEPhrasalVerb/pass
=== RUN   TestSTEPhrasalVerb/fail
--- PASS: TestSTEPhrasalVerb (0.00s)
    --- PASS: TestSTEPhrasalVerb/pass (0.00s)
    --- PASS: TestSTEPhrasalVerb/fail (0.00s)
=== RUN   TestSTEGlossaryTerm
=== RUN   TestSTEGlossaryTerm/pass
=== RUN   TestSTEGlossaryTerm/fail
=== RUN   TestSTEGlossaryTerm/inflections_count
=== RUN   TestSTEGlossaryTerm/silent_e_drops_before_-ing
=== RUN   TestSTEGlossaryTerm/silent_e_drops_before_-ing,_second_word
=== RUN   TestSTEGlossaryTerm/consonant_y_becomes_-ies
=== RUN   TestSTEGlossaryTerm/consonant_y_becomes_-ied
=== RUN   TestSTEGlossaryTerm/regular_-ing_survives_the_y_rule
=== RUN   TestSTEGlossaryTerm/multi-word_term
=== RUN   TestSTEGlossaryTerm/word_boundary
--- PASS: TestSTEGlossaryTerm (0.00s)
    --- PASS: TestSTEGlossaryTerm/pass (0.00s)
    --- PASS: TestSTEGlossaryTerm/fail (0.00s)
    --- PASS: TestSTEGlossaryTerm/inflections_count (0.00s)
    --- PASS: TestSTEGlossaryTerm/silent_e_drops_before_-ing (0.00s)
    --- PASS: TestSTEGlossaryTerm/silent_e_drops_before_-ing,_second_word (0.00s)
    --- PASS: TestSTEGlossaryTerm/consonant_y_becomes_-ies (0.00s)
    --- PASS: TestSTEGlossaryTerm/consonant_y_becomes_-ied (0.00s)
    --- PASS: TestSTEGlossaryTerm/regular_-ing_survives_the_y_rule (0.00s)
    --- PASS: TestSTEGlossaryTerm/multi-word_term (0.00s)
    --- PASS: TestSTEGlossaryTerm/word_boundary (0.00s)
=== RUN   TestSTECodeSpanIsolation
=== RUN   TestSTECodeSpanIsolation/live_rules
=== RUN   TestSTECodeSpanIsolation/code_span_holding_a_pipe_and_a_semicolon
=== RUN   TestSTECodeSpanIsolation/code_span_does_not_join_the_prose_around_it
--- PASS: TestSTECodeSpanIsolation (0.00s)
    --- PASS: TestSTECodeSpanIsolation/live_rules (0.00s)
    --- PASS: TestSTECodeSpanIsolation/code_span_holding_a_pipe_and_a_semicolon (0.00s)
    --- PASS: TestSTECodeSpanIsolation/code_span_does_not_join_the_prose_around_it (0.00s)
=== RUN   TestSTESegmentsSkipCode
=== RUN   TestSTESegmentsSkipCode/fenced_code
=== RUN   TestSTESegmentsSkipCode/tilde_fence
=== RUN   TestSTESegmentsSkipCode/inline_code_span_masked_to_its_own_width
=== RUN   TestSTESegmentsSkipCode/table_row_of_only_code
=== RUN   TestSTESegmentsSkipCode/table_cells_split
=== RUN   TestSTESegmentsSkipCode/pipe_inside_a_code_span_does_not_split_a_cell
--- PASS: TestSTESegmentsSkipCode (0.00s)
    --- PASS: TestSTESegmentsSkipCode/fenced_code (0.00s)
    --- PASS: TestSTESegmentsSkipCode/tilde_fence (0.00s)
    --- PASS: TestSTESegmentsSkipCode/inline_code_span_masked_to_its_own_width (0.00s)
    --- PASS: TestSTESegmentsSkipCode/table_row_of_only_code (0.00s)
    --- PASS: TestSTESegmentsSkipCode/table_cells_split (0.00s)
    --- PASS: TestSTESegmentsSkipCode/pipe_inside_a_code_span_does_not_split_a_cell (0.00s)
=== RUN   TestSTECompare
=== RUN   TestSTECompare/equal
=== RUN   TestSTECompare/above_baseline
=== RUN   TestSTECompare/below_baseline
=== RUN   TestSTECompare/missing_entry
=== RUN   TestSTECompare/stale_entry
--- PASS: TestSTECompare (0.00s)
    --- PASS: TestSTECompare/equal (0.00s)
    --- PASS: TestSTECompare/above_baseline (0.00s)
    --- PASS: TestSTECompare/below_baseline (0.00s)
    --- PASS: TestSTECompare/missing_entry (0.00s)
    --- PASS: TestSTECompare/stale_entry (0.00s)
=== RUN   TestSTERatchet
--- PASS: TestSTERatchet (0.15s)
PASS
ok  	github.com/Harrison-Blair/hive	0.155s
exit 0
```

## Limits and recommendation

The approved boundary is local user and cooperating writer.
Hostile replacement between `Lstat` and `Open` remains deferred for T01.
Hostile or unbounded writers also remain deferred.
No device-node fixture ran. The regular-file predicate rejects those static nonregular modes by inspection.
These results do not claim atomic no-follow opens against a hostile path replacement.

Recommendation: accept this bounded three-path T01 source follow-up for a separate landing.
The coordinator owns that decision and any root merge or record write.
T08-local handling and the other T08 findings remain outside this acceptance.
No source correction is necessary from this pass.

## Preservation and cleanup

Cleanup is complete. Compact evidence existed before the cleanup.
`cleanup-manifest.json` records every entry without a link traversal before removal.
I removed only the owned `red`, `fixed`, and `source08` copies and `/tmp/lgv-hwb5yezm`.
`cleanup-output.txt` and `cleanup-receipt.json` record the removal and absence checks.
Exact probes, patches, raw outputs, synthetic logs, hashes, and frozen source snapshots remain in captures.
No product binary or product workspace was necessary.
I removed no actual worktree `.hive/logs`, `.hive/state`, or `.hive/tmp` root.
No required T01 verification step remains uncompleted.
The separate landing and T08 work remain uncompleted here.

The STE baseline remains exact. New report and evidence prose pass hard count 0 with both glossaries.
No global harness configuration, credential, historical attribution, or other agent resource changed.
No co-author or false session trailer appears.

END T01 LOG GUARD VERIFY 01
