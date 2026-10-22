# T08-tui-log-view FIX-02

| Item | Value |
| --- | --- |
| Task | T08-tui-log-view |
| Pass | `t08-fix-02` |
| Role | Bounded T08 fix worker |
| Requested route | External Pi, `opencode-go/deepseek-v4-flash:medium` |
| Effective route | `opencode-go` / `deepseek-v4-flash`, `PI_REASONING_LEVEL=high` |
| Substitution | None |
| Branch | `task/T08-tui-log-view` |
| Start tip | `7ec0bba6084cc5f29d97f862f4145ad0b00a0f82` |
| Frozen worker and parent | `e341019bb7d318409dad4fcddae61467d35c9748` |
| Production source | `5a1c0317e205cb88828d3bcc17db5442878148e6` |
| Original activation base | `eded8f9cb09208bfdc81334870e86479e746c183` |
| Accepted logger | `66d6cd7f1e84fdc964dd757848f0392c78ea1a48` |
| Dependency merge | `f8260e7a9e7dd5e89c23cf51c4389960dad08bf3` |
| Test commit | `e694cd82d9037b2de9603eb7cbe85fafcfdb3ff7` |
| Source commit | `ecb9c46ed3d1ce3579ee841ffc64a9613049537b` |
| Test lint commit | `c49d91dc0eb3a4b83967a17ef9e0d7089d49aefd` |
| Date | 2026-10-18 |
| Result | P03 and H03 close in this delivery. A fresh verifier owns source acceptance and the `VERIFIED.md` append. |

## Scope and provenance

This delivery closes two findings of the second verification: P03 and the later-path part of H03.
It keeps P01, P02, H01, and H02 closed.
It changes only `internal/tui/tui.go`, `internal/tui/logview.go`, and the adjacent test file `internal/tui/fix_test.go`.
It adds only this evidence file `spec/hive-v1/evidence/T08-tui-log-view/FIX-02.md`.

The delivery does not change the accepted logger, the T06 reader or follower, `cmd/common`, `internal/lib/output`, the store, or the tracker.
It does not change `go.mod`, `go.sum`, or the STE baseline.
The fixed-clock constructor in `internal/logs/read.go` and its test stay exact.
The original V1 to V4, FIX-01, `VERIFIED.md`, and every old report stay byte-exact.
No command deletes a worktree runtime root or a preexisting entry.
No command sends a nonzero signal to another process or closes another workspace.

The exact received brief is in `.hive/tmp/coordinator/reports/t08-fix-02-captures/received-brief.md`.
Its SHA256 is `107eafd0e229cdbb8dec304129c4e1f5560778d46bcd146acf5ff57b57a07d48`.

The source snapshot at the frozen start is `7ec0bba6`.
The test commit adds the new assertions. It fails on the frozen source.
The source commit adds the two minimal fixes. It passes the new assertions.

## Findings and dispositions

| ID | Result | Basis |
| --- | --- | --- |
| P03 | Closed | A nonterminal file descriptor goes through the finite EOF guard. The actual closed pipe and the empty regular file exit 0 in about 0.50 seconds. |
| H03 | Closed | The checked logging path shows the first handler refusal in a bounded body error. All six static later-path cases show the error and write no external target. |
| P01 | Kept closed | The tick stays the sole consumer of the follower event channel. |
| P02 | Kept closed | Output errors, nil-error short writes, and the real `/dev/full` writer return explicit errors. |
| H01 | Kept closed | The PID bound stays before the process lookup and the zero signal. |
| H02 | Kept closed | The first header error persists and later unchanged ticks add no error event. |
| H04 | Historic limit | The original runtime bytes stay unrecoverable. This pass does not claim a restoration. |

## P03: finite nonterminal input

### Cause

`guardReader` exempted every reader with `Fd`.
The CLI passes `*os.File` for a pipe and for a regular file.
Bubble Tea then used its epoll cancel reader on that file.
A closed pipe gave a clean end of the reader and left the program active.
An empty regular file failed in the epoll interest list and returned status 1.

### Test-first red

The new tests are `TestRunHomeFiniteFDPipeStops` and `TestRunHomeFiniteFDFileStops`.
They assert a nil result within two seconds.
The red run uses the frozen source at `7ec0bba6`.

```sh
TMPDIR=/tmp/t8f02-go go test -race -count=1 -timeout=120s ./internal/tui -run '^(TestRunHomeFiniteFDPipeStops|TestRunHomeFiniteFDFileStops)$' -v
```

```text
=== RUN   TestRunHomeFiniteFDPipeStops
    fix_test.go:311: RunHome did not stop on a finite pipe EOF
--- FAIL: TestRunHomeFiniteFDPipeStops (2.00s)
=== RUN   TestRunHomeFiniteFDFileStops
    fix_test.go:327: RunHome on a finite regular file: tui: bubbletea: could not create cancelable reader: add reader to epoll interest list
--- FAIL: TestRunHomeFiniteFDFileStops (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui	2.017s
FAIL
exit 1
```

The exact raw output is in `t08-fix-02-captures/red/P03-H03-red.txt`.

### Minimal green

`guardReader` now keeps a reader unwrapped only when `xterm.IsTerminal` reports a terminal.
Every other reader gets the existing guard.
Bubble Tea then uses its fallback reader.
The guard turns the finite end of input into a quit on the next tick.
A non-EOF input error still stops the program and returns the input error.
The real terminal keeps its identity, so the raw mode and the cancellable reader stay exact.

```sh
TMPDIR=/tmp/t8f02-go go test -race -count=1 -timeout=120s ./internal/tui -run '^(TestRunHomeFiniteFDPipeStops|TestRunHomeFiniteFDFileStops)$' -v
```

```text
--- PASS: TestRunHomeFiniteFDPipeStops (0.50s)
--- PASS: TestRunHomeFiniteFDFileStops (0.50s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/tui	2.014s
```

The exact raw output is in `t08-fix-02-captures/green/P03-H03-green.txt`.

### Exact verifier replay

The exact verifier probe `primary-adapted-02_test.go` came from the second verification.
This pass bound two execution paths: the capture directory and the fixture root.
The bound file passes `TestPVRunStringReaderEOF` and the P02 writer cases.
The exact output is in `t08-fix-02-captures/replay-primary/green.txt`.

### Actual CLI and PTY

One absolute binary ran the actual CLI cases.
The binary SHA256 is `4e42071a637ac28dea0f74834673487bb1fffa9b80bf24e9d681debec0b44ce1`.
Each case used one unique marked fixture home outside the worktree.

| Case | Status | Time | Result |
| --- | --- | --- | --- |
| closed pipe, empty stdin | 0 | 0.50 s | Program and follower settle. The parent reaps the child. |
| empty regular file stdin | 0 | 0.50 s | Program and follower settle. The parent reaps the child. |
| `q` on a pipe | 0 | 0.003 s | Exit 0. The home is visible. No extra envelope. |
| Ctrl-C byte on a pipe | 0 | 0.002 s | Exit 0. |
| bare `--json` | 1 | 0.002 s | One failed envelope with phase `validate` and error `TUI requires text output`. |
| owned PTY, `q` byte 0x71 without newline | 0 | 0.031 s | Raw mode starts. The complete pre-exit attributes equal the post-exit attributes. The home is visible. |

The exact raw output, statuses, PIDs, and reap records are in `t08-fix-02-captures/cli/`.
The probe script is `t08-fix-02-captures/probe-cli.py`.
It follows the case logic of the verifier `cli.py` with owned paths.

## H03: later log refusal view

### Cause

`logAction` and `logError` called `slog.Logger.Debug`.
`slog` discards the error of `Handler.Handle`.
The accepted writer checks the static log path before it opens a new daily file.
A first-open or a rotation obstruction therefore returned an error that the TUI discarded.
The view stayed healthy, and the follower kept waiting for a regular next-day file.

### Test-first red

The new test is `TestActionLogRefusalShowsError`.
It has six cases: first-open and rotation, each with a daily link, a daily FIFO, and a linked log directory.
It asserts an explicit bounded error and no `no events today` line.
It also checks that the external sentinel bytes and the external directory entry count stay exact.

```sh
TMPDIR=/tmp/t8f02-go go test -race -count=1 -timeout=120s ./internal/tui -run '^TestActionLogRefusalShowsError$' -v
```

```text
--- FAIL: TestActionLogRefusalShowsError (0.01s)
    --- FAIL: TestActionLogRefusalShowsError/first-open/daily-link (0.00s)
    --- FAIL: TestActionLogRefusalShowsError/first-open/daily-fifo (0.00s)
    --- FAIL: TestActionLogRefusalShowsError/first-open/linked-log-dir (0.00s)
    --- FAIL: TestActionLogRefusalShowsError/rotation/daily-link (0.00s)
    --- FAIL: TestActionLogRefusalShowsError/rotation/daily-fifo (0.00s)
    --- FAIL: TestActionLogRefusalShowsError/rotation/linked-log-dir (0.00s)
FAIL
exit 1
```

The exact raw output is in `t08-fix-02-captures/red/H03-late-path-red.txt`.

### Minimal green

One TUI-local path `recordLog` builds the record and calls the handler.
On a refusal it keeps the first error in a bounded view field.
The body shows that error in place of the viewport.
The path never logs the refusal, so it cannot recurse.
The first error stays. A later refusal does not replace it.
`logAction` and `logError` are the only callers. No fallback logger starts.

```sh
TMPDIR=/tmp/t8f02-go go test -race -count=1 -timeout=120s ./internal/tui -run '^TestActionLogRefusalShowsError$' -v
```

```text
--- PASS: TestActionLogRefusalShowsError (0.00s)
    --- PASS: TestActionLogRefusalShowsError/first-open/daily-link (0.00s)
    --- PASS: TestActionLogRefusalShowsError/first-open/daily-fifo (0.00s)
    --- PASS: TestActionLogRefusalShowsError/first-open/linked-log-dir (0.00s)
    --- PASS: TestActionLogRefusalShowsError/rotation/daily-link (0.00s)
    --- PASS: TestActionLogRefusalShowsError/rotation/daily-fifo (0.00s)
    --- PASS: TestActionLogRefusalShowsError/rotation/linked-log-dir (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/tui	1.009s
```

The exact raw output is in `t08-fix-02-captures/green/P03-H03-green.txt`.

### Exact verifier replay

The exact verifier probe `late_log_test.go` came from the second verification.
Its SHA256 is `28cd230a0847d9dec617be583f9760555e76813a8ec9607b8d3112f02b101123`.
The frozen source replay fails the same four cases as the second verification.
The fixed source replay passes all six cases. Each view shows the error in about 0.00004 seconds or less.
The exact files are in `t08-fix-02-captures/red/H03-late-path-red.txt` and `t08-fix-02-captures/green/H03-late-path-green.txt`.

The fixed replay shows these errors:

```text
first-open/daily-link       error: log path is not a regular file: .../2044-02-03.jsonl
first-open/daily-fifo       error: log is not a regular file: .../2044-02-03.jsonl
first-open/linked-log-dir   error: log directory is not a native directory: .../.hive/logs
rotation/daily-link         error: log path is not a regular file: .../2044-02-04.jsonl
rotation/daily-fifo         error: log path is not a regular file: .../2044-02-04.jsonl
rotation/linked-log-dir     error: log directory is not a native directory: .../.hive/logs
```

### Healthy action event

A healthy action writes one event through the same checked path.
`TestHealthyActionLogIsWrittenOnce` asserts one new line with message `tui action`, action `level`, and component `tui`.
The existing H02 privacy and no-feedback checks stay green.
The complete probe replay is in `t08-fix-02-captures/replay-primary/combined-green.txt`.

## V1 to V4

The new V1 to V4 evidence uses the fixed source commit `c49d91dc`.

### V1 Header

```sh
TMPDIR=/tmp/t8f02-go go test -race -count=1 -timeout=120s ./internal/tui -run '^TestHeader' -v
```

```text
--- PASS: TestHeaderRejectsOutOfRangePIDs (0.01s)
--- PASS: TestHeaderPersistentErrorDoesNotLogEachTick (0.00s)
--- PASS: TestHeaderShowsHomeAndRefreshesStoreCounts (0.01s)
--- PASS: TestHeaderInvalidStoreShowsBoundedError (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/tui	1.025s
```

The raw output is in `t08-fix-02-captures/v1/V1-green.txt`.

### V2 Keys, follow, and daily file

```sh
TMPDIR=/tmp/t8f02-go go test -race -count=1 -timeout=120s ./internal/tui -run 'TestKeys|TestFollow|TestDailyFile' -v
```

```text
--- PASS: TestFollowReadsAppendedLineOnTick (0.70s)
--- PASS: TestFollowControlsViewport (0.53s)
--- PASS: TestDailyFileRolloverKeepsEarlierLines (0.53s)
--- PASS: TestFollowKeepsIncompleteLineUntilNewline (1.03s)
--- PASS: TestDailyFileWaitsForFirstFile (0.53s)
--- PASS: TestDailyFileInvalidTypeShowsError (0.00s)
--- PASS: TestKeysFilterLevelFollowAndQuit (1.07s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/tui
```

The raw output is in `t08-fix-02-captures/v2/V2-green.txt`.

### V3 End-to-end run

```sh
TMPDIR=/tmp/t8f02-go go test -race -count=1 -timeout=120s ./internal/tui -run '^TestRun(Quits|ResolvesFixtureHome)$' -v
TMPDIR=/tmp/t8f02-go go build -o /tmp/t8f02-cli/hive-t08 .
printf q | /tmp/t8f02-cli/hive-t08 --home <owned-fixture-home>
/tmp/t8f02-cli/hive-t08 --json --home <owned-fixture-home>
```

The test passes. The default command exits 0, shows the home, and writes no envelope.
The `--json` command exits 1 with one failed envelope. The phase is `validate`. The error is `TUI requires text output`.
The actual CLI and PTY cases are in the P03 section.
The raw output is in `t08-fix-02-captures/cli/`.

### V4 Gate

The gate runs from the worktree root at `c49d91dc`.
The Go `TMPDIR` is `/tmp/t8f02-go`, outside every marked home.

```sh
TMPDIR=/tmp/t8f02-go gofmt -l $(git ls-files -co --exclude-standard '*.go')
TMPDIR=/tmp/t8f02-go go vet ./...
TMPDIR=/tmp/t8f02-go golangci-lint run --allow-serial-runners ./...
TMPDIR=/tmp/t8f02-go go test -race -count=1 -timeout=300s ./...
TMPDIR=/tmp/t8f02-go go test -run TestSTE -count=1 -v .
```

```text
format: no output, exit 0
vet: no output, exit 0
0 issues., exit 0
22 packages pass, exit 0
TestSTERatchet pass, exit 0
```

The full raw gate output is in `t08-fix-02-captures/v4/`.
The STE baseline SHA256 is `d4c5882428b5e9ae7e829fc60f395c501f5eeb82b1d5ec2d74654e9e1197cfc9` and stays exact.
The `go.mod` SHA256 is `441604e3567e8afd6445a1def1ac68670072dfd891e9271500e5fc69586e0b79`.
The `go.sum` SHA256 is `a4ebc7753f3933c21ffdad3c152412aaa263f0c868db0d636c0fc857dcb7a71c`.
Both stay exact.

## Sensitivity mutations

Each mutation compiles. Each red run fails on a new assertion. Each restored run passes.
The restored `internal/tui/tui.go` SHA256 is `52062dd4c9e0954552f1a30f79087d53f6a52059bbb000046c5827eec723adfe`.

| Finding | Broken behavior | Red | Restored |
| --- | --- | --- | --- |
| P03 | Restore the blanket Fd exemption | `did not stop on a finite pipe EOF` and the epoll error | Pass in 0.50 s |
| H03 | Discard the handler refusal | Six `view does not contain "error:"` failures | Pass |

The patches, red output, and restored output are in `t08-fix-02-captures/mutations/`.

## Freeze, preservation, and limits

The before and after no-follow worktree inventories are in `t08-fix-02-captures/wt-runtime-inventory-before.tsv` and `wt-runtime-inventory-final.tsv`.
The comparison holds every preexisting path.
Hashes of the stable preexisting entries are separate from this pass's own capture files.
The pass adds only the new report and the new captures.
No command deletes worktree `.hive/logs`, `.hive/state`, or `.hive/tmp`.
The own fixtures, binaries, copies, and `TMPDIR` roots are in the cleanup manifest.
All owned CLI, EOF, PTY, and test children exited or reached their finite deadline. The parent reaped each child.

The historic H04 bytes stay absent. This pass does not synthesize old logs.
It does not claim that the old cleanup complied. It does not rewrite the old report or the old evidence.
The trust boundary is the local user, the cooperating writer, and the static types.
A hostile replacement between `Lstat` and `Open` stays deferred for T01.
An arbitrary forever-blocked reader or writer stays out of scope.
Physical terminal clipping stays a limit.

## Uncompleted and recommendation

Uncompleted: the fresh verifier check, the source acceptance, and the `VERIFIED.md` append.
The coordinator owns those steps and the later checkpoint decision.
No CP4 or future checkpoint pass applies.

Recommendation: send this branch to a fresh verifier.
The verifier should replay the exact P03 and H03 probes, run the full gate at the final commit, and keep the H04 provenance limit.

End of T08-tui-log-view FIX-02 evidence.
