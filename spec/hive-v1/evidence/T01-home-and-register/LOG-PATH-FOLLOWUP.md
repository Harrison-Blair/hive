# T01-home-and-register log path follow-up

| Item | Value |
| --- | --- |
| Task | T01-home-and-register |
| Pass | t01-log-guard-01 |
| Proves | AC5 and AC7 for the shared log path check |
| Base | ff5e8c44d9b7d1fb510a568e9a248d84cab0b906 |
| Branch | `task/T01-log-path-guards` |
| Test commit | b775c963 |
| Source commit | d5639e11 |
| Evidence commit | The commit that adds this file |
| Reference source | `task/T08-tui-log-view` at e26d08bdda4eb3aad5986796c35edb7a7d8c8175 |
| Date | 2026-10-18 |
| Agent | worker, pi, opencode-go/deepseek-v4-flash, medium requested, high effective |

## Scope

This pass closes the shared logger cause of H03 in the T08 header check.
The base `internal/lib/log/log.go` made `.hive/logs` and opened the daily
file without a static type check. A daily link, FIFO, socket, directory or
non-native directory component could cause an external write or a block.

The pass adds one private check. It uses `Lstat` on `.hive`, `.hive/logs`
and the prospective daily path. It rejects an existing non-native directory.
It rejects an existing nonregular daily file. It allows a missing path and
an existing regular daily file. A whole-home symbolic link stays valid.
`NewWriter` checks the current date. `Handle` checks each new daily path and
rotation. A failed check keeps the current open file. A later call can retry
after the static obstruction is gone. The pass keeps modes 0755 and 0644,
date rotation, handler state, the mutex, the JSON fields, the body cap and
the existing output format.

Changed files:

| Path | Change |
| --- | --- |
| `internal/lib/log/log.go` | Add the private static path check. Open a new daily file before the close of the old file. |
| `internal/lib/log/log_path_guard_test.go` | New assertion-red tests for static path types, rotation, retry and valid controls. |
| `spec/hive-v1/evidence/T01-home-and-register/LOG-PATH-FOLLOWUP.md` | This record. |

No public logging API changed. No dependency changed. No other source file
changed. No code deletes a log file, removes a path or changes a path.

The approved boundary is local user and cooperating writer. A hostile
replacement between `Lstat` and `Open` stays deferred for T01. A hostile or
unbounded writer also stays deferred. Tests use owned finite fixtures under
short `TMPDIR` paths outside the home. No actual source08 file changed.

## Red run

The test commit was present and the base `log.go` was active.
The new tests failed on assertions, not on compilation.

```sh
go test ./internal/lib/log/ -run 'TestStaticLogPath' -count=1 -v
```

```text
--- FAIL: TestStaticLogPathRejectsLinkedHiveDirectory (0.00s)
    NewWriter accepted a linked .hive directory; external entries = [logs]
--- FAIL: TestStaticLogPathRejectsLinkedLogsDirectory (0.00s)
    NewWriter accepted a linked .hive/logs directory; external entries = [2026-01-02.jsonl]
--- FAIL: TestStaticLogPathRejectsNonRegularDailyPaths (0.00s)
    NewWriter accepted a nonregular daily fifo path
--- FAIL: TestStaticLogPathRejectsDailySocket (0.00s)
    NewWriter accepted a socket daily path
--- FAIL: TestStaticLogPathRejectsRotationToFIFOWithoutBlocking (2.00s)
    Handle blocked more than 2 seconds on a static FIFO rotation path
--- FAIL: TestStaticLogPathRejectsRotationToDailyLinkAndRetries (0.00s)
    Handle accepted a symlink rotation path and wrote through it
--- FAIL: TestStaticLogPathRejectsLinkedLogsDirectoryOnRotation (0.00s)
    Handle accepted a linked logs directory on rotation
exit 1
```

The FIFO case used an owned cooperating reader to release the blocked old
call. No process got a signal. The test left no blocked goroutine.

The exact H03 probes also ran in an owned archive of source08. The copy held
the two retained probe files. The original `log.go` was active.

```sh
go test ./internal/tui -run 'TestHVLogging(BadDailyLinkNoWrite|BadDailyFIFOBounded|LinkedLogDirectoryError)' -count=1 -v
```

```text
--- FAIL: TestHVLoggingBadDailyLinkNoWrite (0.00s)
    target before=37 bytes after=241 bytes
    error logger followed rejected link and changed synthetic target
--- FAIL: TestHVLoggingBadDailyFIFOBounded (2.00s)
    New blocked more than 2 seconds on rejected FIFO until an owned reader opened it
--- FAIL: TestHVLoggingLinkedLogDirectoryError (0.00s)
    linked log directory produced a successful view and an external-target write, want explicit error and no write
exit 1
```

## Green run

The same focused tests passed with source commit `d5639e11`.

```sh
go test ./internal/lib/log/ -run 'TestStaticLogPath' -count=1 -v
```

```text
--- PASS: TestStaticLogPathRejectsLinkedHiveDirectory (0.00s)
--- PASS: TestStaticLogPathRejectsLinkedLogsDirectory (0.00s)
--- PASS: TestStaticLogPathRejectsNonRegularDailyPaths (0.00s)
--- PASS: TestStaticLogPathRejectsDailySocket (0.00s)
--- PASS: TestStaticLogPathRejectsDailyFIFOWithoutBlocking (0.00s)
--- PASS: TestStaticLogPathRejectsRotationToFIFOWithoutBlocking (0.00s)
--- PASS: TestStaticLogPathRejectsRotationToDailyLinkAndRetries (0.00s)
--- PASS: TestStaticLogPathRejectsLinkedLogsDirectoryOnRotation (0.00s)
--- PASS: TestStaticLogPathAllowsValidCreationAppendAndWholeHomeAlias (0.00s)
PASS
```

The same H03 probes passed in the owned source08 archive. Only `log.go`
changed in that archive.

```text
--- PASS: TestHVLoggingBadDailyLinkNoWrite (0.00s)
    target before=37 bytes after=37 bytes
--- PASS: TestHVLoggingBadDailyFIFOBounded (0.00s)
--- PASS: TestHVLoggingLinkedLogDirectoryError (0.00s)
    View="... error: log: log directory is not a native directory: .../.hive/logs ..."
    synthetic link target files=0
PASS
```

The old log tests and the home and static-home safeguards passed without a
change.

```sh
go test -race -count=1 ./internal/lib/log/ -v
go test -race -count=1 ./internal/lib/home/ -run TestResolve -v
go test -race -count=1 ./cmd/ -run 'TestProjectList|TestRootMissingHome|TestRootProjectListJSONEnvelope|TestResolveEmptyHomeFlagHasUsageStatus' -v
```

```text
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.037s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.006s
ok  	github.com/Harrison-Blair/hive/cmd	1.011s
```

## H03 shared cases

| Static case | Base logger | New logger | T08-local remainder |
| --- | --- | --- | --- |
| Daily link | The error logger followed the link. The target grew from 37 to 241 bytes. | `NewWriter` returns a regular-file error. The target stays at 37 bytes. | T08 owns the reader path error and the bounded error view. |
| Daily FIFO | `New` blocked for more than 2 seconds. | `NewWriter` returns a regular-file error at once. No open or block occurs. | T08 owns the view text for the constructor error. |
| Linked `.hive/logs` | The view showed no error. The quit action made one file in the target. | `NewWriter` returns a native-directory error. The target has no file. | T08 owns the reader path error and the bounded error view. |

These results close the shared logger cause for the three exact H03 probes.
They do not close all of H03 and they do not accept T08. A static bad path
that appears after `NewWriter` can still make a later `Handle` return an
error. The `slog` package drops that error. T08 must decide how to show a
later path failure. T08 also owns H01, H02 and H04. A fresh verifier owns
`LOG-PATH-VERIFIED.md`.

## Call sites

The logger open remains only in `internal/lib/log/log.go`.
The production call sites in this worktree are:

| Call site | Error handling |
| --- | --- |
| `cmd/common/common.go` | Returns the error. |
| `cmd/doctor/doctor.go` | Returns a failed envelope with phase `log`. |
| `cmd/inbox/inbox.go` | Returns the error. |
| `cmd/handoff/write/write.go` | Ignores the error and writes no log event. |
| `cmd/memory/write/write.go` | Ignores the error and writes no log event. |
| `internal/agent/register.go` | Returns the error. |
| `internal/lib/inbox/inbox.go` | Returns the error. |
| `internal/task/log.go` | Returns the error. |

The two handoff and memory call sites keep their old error policy. On a
rejected path they write no log event and no external target. The T08 TUI
call site is not in this worktree.

The exact search was:

```sh
grep -rnE 'hivelog\.(Open|NewWriter)|[^a-zA-Z]log\.(Open|NewWriter)\(' --include='*.go' --exclude-dir=.hive .
grep -nE 'os\.(Open|OpenFile|ReadFile|WriteFile|Create|CreateTemp)\(' internal/lib/log/log.go
```

The only direct open in the logger source is the guarded daily `OpenFile`.

## Gate

The gate ran from the worktree with a short `TMPDIR` outside the home.

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run ./...
go test -race -count=1 ./...
go test -run 'TestSTE' -count=1 -v .
```

```text
gofmt exit 0
vet exit 0
0 issues.
lint exit 0
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.048s
ok  	github.com/Harrison-Blair/hive/cmd	1.020s
ok  	github.com/Harrison-Blair/hive/internal/tui	1.005s
test exit 0
--- PASS: TestSTERatchet (0.15s)
ste exit 0
```

The full test output passed for each package. Raw outputs stay in the worker
report captures at `.hive/tmp/coordinator/reports/t01-log-guard-01-captures/`.

## Cleanup

The pass removed its owned source08 archive copy and its short `TMPDIR`.
It retained the red and green outputs, the probe outputs, the patches and the
hashes in the worker report captures. It removed no worktree `.hive/logs`,
`.hive/state` or `.hive/tmp` path. It changed no actual source08 file.

## Limits

A hostile replacement between `Lstat` and `Open` stays deferred. The pass
uses static checks only. It does not prove TUI error presentation, a later
rotation failure, H01, H02 or H04. Independent verification stays required.
