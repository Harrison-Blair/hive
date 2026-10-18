# T08-tui-log-view FIX-01

| Item | Value |
| --- | --- |
| Task | T08-tui-log-view |
| Pass | `t08-fix-01` |
| Role | Bounded T08 fix worker |
| Requested route | External Pi, `opencode-go/deepseek-v4-flash:medium` |
| Effective route | `opencode-go` / `deepseek-v4-flash`, `PI_REASONING_LEVEL=high` |
| Branch | `task/T08-tui-log-view` |
| Start tip | `223f3937d5b8c63690e41309da48d7c5f6f3e792` |
| Original worker | `e26d08bdda4eb3aad5986796c35edb7a7d8c8175` |
| Last T08 source | `d49d77a0c437c61bc1cb88853183308e26f385ac` |
| Activation base | `eded8f9cb09208bfdc81334870e86479e746c183` |
| Imported dependency | `66d6cd7f1e84fdc964dd757848f0392c78ea1a48` |
| Dependency merge | `f8260e7a9e7dd5e89c23cf51c4389960dad08bf3` |
| Test commit | `a24faa51eb24970e5964bf0064a372ce0d4c4d24` |
| Source commit | `4df092fc27fdbe5ab9712fc237078485a77edfdb` |
| Lint test commit | `f43bd731294f713671e522e3839306d12f8a3334` |
| Date | 2026-10-18 |
| Result | New hard findings 0. A fresh verifier owns source acceptance and `VERIFIED.md`. |

## Scope and provenance

This delivery closes the first-review findings P01 to P03 and H01 to H04.
It changes only the T08 table paths in `internal/tui/` and the adjacent tests.
It imports the verified T01 logger commit and its accepted ancestry.
It does not change the imported logger, the T05 source, the tracker, or the old evidence.
It does not change `go.mod`, `go.sum`, `cmd/common`, or `internal/lib/output`.
The reader and follower behavior of T06 stays exact.
The fixed-clock constructor in `internal/logs/read.go` and its test stay exact.

The dependency merge has two parents. The first parent is the start tip.
The second parent is `66d6cd7f`. The merge imports 30 paths. The report records the full map.
The merge adds the T01 static log-path guard and the accepted T05 source and evidence.
The T08 bare entry keeps the resolved home, the text output, and the `--json` refusal.
The T05 `doctor` command and all older command registrations remain.
The merge produced no conflict.

The original worker removed worktree `.hive/logs`, `.hive/state`, and `.hive/tmp`.
That removal violates the preservation rule.
No pre-deletion inventory proves the removed bytes, their hashes, or their ownership.
Source changes cannot restore those historic bytes.
This evidence does not claim that the old cleanup complied.
It does not synthesize old logs.
It keeps the H04 deviation and the provenance limit.
This delivery keeps the full new runtime inventory and all preexisting runtime entries.

## Reading this evidence

The fenced blocks keep the decisive lines and the exact exit status.
The captures keep the exact raw output for every command.
The lint test commit changes only one test variable name.
The product source and the binary are the same at the source commit and the lint test commit.

## Findings and dispositions

| ID | Result | Evidence |
| --- | --- | --- |
| P01 | Closed | The tick is the only consumer of the follower event channel. V2 red shows the reverse order. V2 green shows A then B. |
| P02 | Closed | The entry keeps output errors and short writes. V2 red shows a nil result. V2 green shows an output error. |
| P03 | Closed | A finite input end closes a channel. The next tick returns `tea.Quit`. V2 red stays active. V2 green settles in 0.50 s. |
| H01 | Closed | `processAlive` rejects a PID above `math.MaxInt32` before lookup. V1 red shows false liveness. V1 green shows `down`. |
| H02 | Closed | A header read error becomes a persistent bounded view and stops later reads. V1 red grows 2 to 6 events. V1 green stays at 2. |
| H03 | Closed with the import | The imported logger rejects a static daily link, a FIFO, and a linked log directory. The T08 view shows the bounded error. |
| H04 | Limit kept | The historic bytes stay unrecoverable. The new run keeps its own inventory and the old deviation. |

## V1 Header fix

The red run uses the new tests on the start source. The green run uses the same tests on the source commit.

```sh
commit f8260e7a
TMPDIR=/tmp/t8f01-go go test -race -count=1 -timeout=120s ./internal/tui -run '^TestHeader' -v
echo "exit $?"
```

```text
=== RUN   TestHeaderRejectsOutOfRangePIDs/plus_2^32
    fix_test.go:214: view does not contain "supervisor: down":
        hive .../TestHeaderRejectsOutOfRangePIDsplus_2^322814881781/001  supervisor: live  agents: 0  inbox: 0 unread
=== RUN   TestHeaderRejectsOutOfRangePIDs/max_int64
    fix_test.go:214: view does not contain "supervisor: down":
        hive .../TestHeaderRejectsOutOfRangePIDsmax_int642242473057/001  supervisor: live  agents: 0  inbox: 0 unread
--- FAIL: TestHeaderRejectsOutOfRangePIDs (0.01s)
    --- FAIL: TestHeaderRejectsOutOfRangePIDs/plus_2^32 (0.00s)
    --- FAIL: TestHeaderRejectsOutOfRangePIDs/max_int64 (0.00s)
=== RUN   TestHeaderPersistentErrorDoesNotLogEachTick/agents
    fix_test.go:262: log events grew from 2 to 6 over unchanged error ticks
=== RUN   TestHeaderPersistentErrorDoesNotLogEachTick/inbox
    fix_test.go:262: log events grew from 2 to 6 over unchanged error ticks
--- FAIL: TestHeaderPersistentErrorDoesNotLogEachTick (0.00s)
--- PASS: TestHeaderShowsHomeAndRefreshesStoreCounts (0.01s)
--- PASS: TestHeaderInvalidStoreShowsBoundedError (0.00s)
FAIL
exit 1
```

The exact red output is in `t08-fix-01-captures/v1/V1-red.txt`.

```sh
commit f43bd731
TMPDIR=/tmp/t8f01-go go test -race -count=1 -timeout=120s ./internal/tui -run '^TestHeader' -v
echo "exit $?"
```

```text
--- PASS: TestHeaderRejectsOutOfRangePIDs (0.01s)
    --- PASS: TestHeaderRejectsOutOfRangePIDs/self (0.00s)
    --- PASS: TestHeaderRejectsOutOfRangePIDs/zero (0.00s)
    --- PASS: TestHeaderRejectsOutOfRangePIDs/negative (0.00s)
    --- PASS: TestHeaderRejectsOutOfRangePIDs/exited (0.00s)
    --- PASS: TestHeaderRejectsOutOfRangePIDs/plus_2^32 (0.00s)
    --- PASS: TestHeaderRejectsOutOfRangePIDs/max_int32_plus_one (0.00s)
    --- PASS: TestHeaderRejectsOutOfRangePIDs/max_int64 (0.00s)
--- PASS: TestHeaderPersistentErrorDoesNotLogEachTick (0.00s)
    --- PASS: TestHeaderPersistentErrorDoesNotLogEachTick/agents (0.00s)
    --- PASS: TestHeaderPersistentErrorDoesNotLogEachTick/inbox (0.00s)
--- PASS: TestHeaderShowsHomeAndRefreshesStoreCounts (0.01s)
--- PASS: TestHeaderInvalidStoreShowsBoundedError (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/tui	1.024s
exit 0
```

The exact green output is in `t08-fix-01-captures/v1/V1-green.txt`.
The header keeps the home path, the supervisor state, the live agent count, and the unread count.
A healthy store still refreshes on the next tick.
Quit still works in the persistent error state.

## V2 Ordered follower and stream lifecycle

The red run uses the new tests on the start source. The green run uses the same tests on the source commit.

```sh
commit f8260e7a
TMPDIR=/tmp/t8f01-go go test -race -count=1 -timeout=180s ./internal/tui -run '^(TestInitDoesNotConsumeFollowerEvents|TestRunHomeReportsOutputFailure|TestRunHomeFiniteEOFStops|TestRunHomeReportsInputFailure|TestNewRejects)' -v
echo "exit $?"
```

```text
=== RUN   TestInitDoesNotConsumeFollowerEvents
    fix_test.go:105: events=2; want 3 before the timeout
--- FAIL: TestInitDoesNotConsumeFollowerEvents (3.72s)
=== RUN   TestRunHomeReportsOutputFailure/short=false
    fix_test.go:143: RunHome returned nil for short=false; want an output error
=== RUN   TestRunHomeReportsOutputFailure/short=true
    fix_test.go:143: RunHome returned nil for short=true; want an output error
--- FAIL: TestRunHomeReportsOutputFailure (0.01s)
=== RUN   TestRunHomeReportsInputFailure
--- PASS: TestRunHomeReportsInputFailure (0.00s)
=== RUN   TestRunHomeFiniteEOFStops
    fix_test.go:185: RunHome did not stop on finite input EOF
--- FAIL: TestRunHomeFiniteEOFStops (2.00s)
=== RUN   TestNewRejectsLinkedDailyLogWithoutWrite
--- PASS: TestNewRejectsLinkedDailyLogWithoutWrite (0.00s)
=== RUN   TestNewRejectsDailyFIFOWithoutBlocking
--- PASS: TestNewRejectsDailyFIFOWithoutBlocking (0.00s)
=== RUN   TestNewRejectsLinkedLogDirectoryWithoutWrite
--- PASS: TestNewRejectsLinkedLogDirectoryWithoutWrite (0.00s)
FAIL
exit 1
```

The exact red output is in `t08-fix-01-captures/v2/V2-red.txt`.
The three H03 constructor tests pass on the red source because the merge already imports the verified logger fix.
Their red run uses the frozen source08 tree with the old logger. The report records that run.

```sh
commit f43bd731
TMPDIR=/tmp/t8f01-go go test -race -count=1 -timeout=180s ./internal/tui -run '^(TestInitDoesNotConsumeFollowerEvents|TestRunHomeReportsOutputFailure|TestRunHomeFiniteEOFStops|TestRunHomeReportsInputFailure|TestNewRejects)' -v
echo "exit $?"
```

```text
--- PASS: TestInitDoesNotConsumeFollowerEvents (0.74s)
--- PASS: TestRunHomeReportsOutputFailure (0.01s)
    --- PASS: TestRunHomeReportsOutputFailure/short=false (0.00s)
    --- PASS: TestRunHomeReportsOutputFailure/short=true (0.00s)
--- PASS: TestRunHomeReportsInputFailure (0.00s)
--- PASS: TestRunHomeFiniteEOFStops (0.50s)
--- PASS: TestNewRejectsLinkedDailyLogWithoutWrite (0.00s)
--- PASS: TestNewRejectsDailyFIFOWithoutBlocking (0.00s)
--- PASS: TestNewRejectsLinkedLogDirectoryWithoutWrite (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/tui	2.265s
exit 0
```

The exact green output is in `t08-fix-01-captures/v2/V2-green.txt`.
The existing key, filter, level, follow, daily-file, and stop tests also pass in the full gate.
The report records the first-review probe replay and the mutation checks.

## H03 imported logger cases

The frozen source08 tree uses the old logger. The fixed tree uses the imported logger.
Each case uses actual synthetic paths and a captured hash or entry count.

```sh
commit e26d08bd
TMPDIR=/tmp/t8f01-go go test -race -count=1 -timeout=120s ./internal/tui -run '^TestHVLogging(BadDailyLinkNoWrite|BadDailyFIFOBounded|LinkedLogDirectoryError)$' -v
echo "exit $?"
```

```text
--- FAIL: TestHVLoggingBadDailyLinkNoWrite (0.00s)
    hv_header_test.go:368: target before=37 bytes after=235 bytes
--- FAIL: TestHVLoggingBadDailyFIFOBounded (2.00s)
    hv_header_test.go:405: New blocked more than 2 seconds on rejected FIFO until an owned reader opened it
--- FAIL: TestHVLoggingLinkedLogDirectoryError (0.00s)
    hv_logdir_test.go:33: synthetic link target files=1
FAIL
exit 1
```

```sh
commit f43bd731
TMPDIR=/tmp/t8f01-go go test -race -count=1 -timeout=180s ./internal/tui -run '^TestHVLogging(BadDailyLinkNoWrite|BadDailyFIFOBounded|LinkedLogDirectoryError)$' -v
echo "exit $?"
```

```text
--- PASS: TestHVLoggingBadDailyLinkNoWrite (0.00s)
    hv_header_test.go:368: target before=37 bytes after=37 bytes
--- PASS: TestHVLoggingBadDailyFIFOBounded (0.00s)
--- PASS: TestHVLoggingLinkedLogDirectoryError (0.00s)
    synthetic link target files=0
PASS
exit 0
```

The T08 view shows a bounded error and no `no events today` line.
The error is `log path is not a regular file` or `log directory is not a native directory`.
The TUI starts no fallback logger and writes no external target.
The check is static. This evidence does not claim a hostile no-follow-open guarantee.
A hostile replacement between `Lstat` and `Open` stays deferred for T01.

## V3 End-to-end

The test run uses the actual worktree. The CLI run uses one absolute binary and one unique marked fixture home.
The fixture home is outside every marked home. No V3 command uses this worktree as its home.

```sh
commit f43bd731
TMPDIR=/tmp/t8f01-go go test -race -count=1 -timeout=120s ./internal/tui -run '^TestRun(Quits|ResolvesFixtureHome)$' -v
echo "exit $?"
```

```text
--- PASS: TestRunQuits (0.00s)
--- PASS: TestRunResolvesFixtureHome (0.00s)
PASS
exit 0
```

```sh
commit f43bd731
go build -o /tmp/t8f01-cli/hive2 .
sha256sum /tmp/t8f01-cli/hive2
mkdir -p /tmp/t8f01-go/cli-fix01b/.hive/logs
printf 'home\n' > /tmp/t8f01-go/cli-fix01b/.hive/README.md
printf '{"time":"2026-10-18T07:12:18-04:00","level":"INFO","component":"cli","msg":"fix fixture event"}\n' > /tmp/t8f01-go/cli-fix01b/.hive/logs/2026-10-18.jsonl
printf q | /tmp/t8f01-cli/hive2 --home /tmp/t8f01-go/cli-fix01b
echo "exit $?"
/tmp/t8f01-cli/hive2 --json --home /tmp/t8f01-go/cli-fix01b
echo "exit $?"
```

```text
406845208d14bd6d5d7fb5d057f9d240039dda8584d93946841e9461148a8ce6  /tmp/t8f01-cli/hive2
hive /tmp/t8f01-go/cli-fix01b  supervisor: down  agents: 0  inbox: 0 unread
INFO cli fix fixture event
...
level: debug  text: -  follow: on
exit 0
{"outcome":"failed","phase":"validate","effects":[],"error":"TUI requires text output","result":null}
exit 1
```

The q run has status 0. It shows the fixture home and the fixture event. It writes no extra envelope. Its stderr is empty.
The bare `--json` run has status 1. It writes one failed envelope with phase `validate` and the exact error `TUI requires text output`. Its stderr is empty.
The synthetic daily file SHA256 is `2886897baf32e4c244499d517b3a6c8096d7c1f1964b7def50de793d18a9b009`.
The fixture marker SHA256 is `9a3e246041d3c27dc3645f79cb0d1eb41c277965614655d17119ed7498b956ec`.
The raw stdout, stderr, status, binary hash, and fixture hash are in `t08-fix-01-captures/v3/`.
The cleanup removes only the owned binary and fixture after this evidence is compact.

## V4 Gate

The gate runs from the actual worktree root. The Go `TMPDIR` is outside every marked home.
The test run does not use the Go test cache. The test run uses `-count=1`.

```sh
commit f43bd731
TMPDIR=/tmp/t8f01-go gofmt -l $(git ls-files -co --exclude-standard '*.go')
echo "exit $?"
TMPDIR=/tmp/t8f01-go go vet ./...
echo "exit $?"
TMPDIR=/tmp/t8f01-go golangci-lint run --allow-serial-runners ./...
echo "exit $?"
TMPDIR=/tmp/t8f01-go go test -race -count=1 ./...
echo "exit $?"
TMPDIR=/tmp/t8f01-go go test -run TestSTE -count=1 -v .
echo "exit $?"
```

```text
format: no output
exit 0
vet: no output
exit 0
0 issues.
exit 0
ok  	github.com/Harrison-Blair/hive	3.261s
ok  	github.com/Harrison-Blair/hive/cmd	1.075s
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.251s
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.225s
ok  	github.com/Harrison-Blair/hive/cmd/version	1.014s
ok  	github.com/Harrison-Blair/hive/internal/agent	1.208s
ok  	github.com/Harrison-Blair/hive/internal/doctor	3.039s
ok  	github.com/Harrison-Blair/hive/internal/handoff	1.135s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.010s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	6.497s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.023s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	1.224s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.062s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	3.481s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.142s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	17.332s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.003s
ok  	github.com/Harrison-Blair/hive/internal/logs	5.517s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.146s
ok  	github.com/Harrison-Blair/hive/internal/project	1.011s
ok  	github.com/Harrison-Blair/hive/internal/task	1.161s
ok  	github.com/Harrison-Blair/hive/internal/tui	6.664s
exit 0
--- PASS: TestSTERatchet (0.15s)
PASS
ok  	github.com/Harrison-Blair/hive	0.150s
exit 0
```

The full raw gate output is in `t08-fix-01-captures/v4/`.
The STE baseline stays exact. New prose passes both glossaries with hard count 0.
The evidence commit changes one Markdown file only. The Go tree equals the gated tree.
The four gate commands pass. The format list is empty.

## Sensitivity and probe replay

Each finding has an assertion-red run and a restored-green run. No compile error counts as red.

| Finding | Broken behavior | Red | Green |
| --- | --- | --- | --- |
| P01 | Add a competing follower consumer to `Init` | `events=2; want 3` | Pass |
| P02 | Ignore the write fault in the guard | `RunHome returned nil` | Pass |
| P03 | Ignore the input-end channel in the tick | `did not stop on finite input EOF` | Pass |
| H01 | Remove the PID upper bound | `plus_2^32` and `max_int64` show `live` | Pass |
| H02 | Remove the persistent error stop | events grow 2 to 10 | Pass |
| H03 | Use the old logger | target grows 37 to 235 bytes, FIFO blocks | Pass |

The exact patches, red output, restored-green output, and restored hashes are in `t08-fix-01-captures/mutations/`.

The exact first-review probes are in `t08-fix-01-captures/probe-primary/` and `probe-header/`.
The exact P01 to P03 probe fails on `f8260e7a` with order `initial, append B, append A`, a nil output result, and a two-second input-end timeout.
The fixed design has one consumer. The green replay changes only the probe lines that use the removed competing consumer and its stop check.
It keeps the real follower, the actual `Update` and `View`, and the same assertion.
The green replay shows order `initial, append A, append B`.
The 20-run q and Ctrl-C lifecycle check starts at 2 goroutines and ends at 3.
No owned follower or command goroutine remains.
The H01 to H03 exact header probes fail on the old TUI or old logger source and pass on the fixed source.
The frozen source08 tree shows the old logger H03 failures.

## Preservation, inventory, and limits

The pre-change worktree runtime inventory records `.hive/README.md` and the original `t08-impl-01.md`.
The post-change inventory keeps every preexisting entry and adds only this pass report and captures.
No command deletes worktree `.hive/logs`, `.hive/state`, or `.hive/tmp`.
No command sends a nonzero signal to another process or closes another workspace.
This pass reads no credential content. It changes no global harness configuration.
It adds no co-author and no false session trailer.

Uncompleted: source acceptance and the `VERIFIED.md` update. A fresh verifier owns both.
The historic H04 bytes stay absent. This delivery cannot restore them and does not claim to.
Hostile `Lstat` and `Open` replacement and hostile writers stay deferred.
Physical terminal clipping stays a limit. The model checks do not prove every terminal size.

End of T08-tui-log-view FIX-01 evidence.
