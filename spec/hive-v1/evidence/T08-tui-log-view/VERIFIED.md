# T08-tui-log-view VERIFIED

| Item | Value |
| --- | --- |
| Task | T08-tui-log-view |
| Worker | `e26d08bdda4eb3aad5986796c35edb7a7d8c8175` |
| Last source | `d49d77a0c437c61bc1cb88853183308e26f385ac` |
| Base | `eded8f9cb09208bfdc81334870e86479e746c183` |
| Branch | `task/T08-tui-log-view` |
| Date | 2026-10-18 |
| Primary | verifier, external Pi, `openai-codex/gpt-6.1-sol:high`, `t08-verify-01` |
| Secondary | external Pi, `openai-codex/gpt-6.1-sol:high`, `t08-header-verify-01` |
| Verdict | **FAIL** |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1, AC3 | FAIL | Nominal tests pass. H01 shows false supervisor status. H02 shows error feedback on ticks. |
| V2 | AC2, AC3 | FAIL | Nominal tests pass. P01 breaks event order. H02/H03 break error-path behavior. |
| V3 | AC1, AC3 | FAIL | Required q/JSON flows pass. P02/P03 fail lifecycle checks. H04 limits the original runtime evidence. |
| V4 | AC4 | PASS | Empty format list, vet, lint, and full uncached race suite pass from the actual worker root. |

This verdict combines both independent areas. It does not approve a landing or a checkpoint.
No source correction forms part of this evidence.

## Paths and source freeze

The primary report is `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t08-verify-01.md`.
Its captures are `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t08-verify-01-captures/`.
`manifest.json` in that directory gives the full path map.
The secondary report is `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t08-header-verify-01.md`.
Its captures are the adjacent `t08-header-verify-01-captures/` directory.

The secondary report reached `END T08 HEADER VERIFY 01` before this pass started.
Its verifier exited. I read its complete findings, outputs, and limits.
All 245 original tracked files retain their hashes, including V1 to V4.
The worker report also retains its original hash.
Only this new evidence file changes the worker tree.
All probes and controlled mutations ran in owned copies.
No actual worker runtime path served as a disposable fixture.

## Findings and recommendations

### P01 — Two consumption paths can reverse log events

Location: `internal/tui/tui.go:150-180,213-227` and `internal/tui/tick.go:30-39`.
`Init` starts `waitEvent`. Each tick also drains the same channel.
A command can hold event A while a tick appends event B. The command then appends A.
`TestPVOneOrderedConsumptionPath` uses the real follower, real files, and actual `Update` and `View`.
It observes `initial`, `append B`, `append A`, rather than `initial`, `append A`, `append B`.
All ten bounded race runs fail on this assertion. No data race warning is necessary for this ordering defect.
Recommendation: use one ordered consumption path. Keep the required tick visibility.

### P02 — Output failures return success

Location: `internal/tui/tui.go:275-282`, with installed Bubble Tea v2.1.0 renderer behavior.
`TestPVRunWriterErrors` runs the real entry with string-reader q and an actual failing output writer.
The writer receives calls. `RunHome` returns nil for an error and for a zero-byte short write.
Each case fails in all ten bounded race runs.
Recommendation: detect output errors and short writes at the entry boundary. Stop and settle the program on failure.
Do not add a lock around output or a long wait.

### P03 — String-reader EOF leaves the program active

Location: `internal/tui/tui.go:275-282` and its program lifecycle.
`TestPVRunStringReaderEOF` runs the actual entry on a finite empty string reader.
An owned test subprocess remains active until its two-second deadline.
The parent stops and reaps only that subprocess. All ten repetitions fail.
This case uses no arbitrary blocked reader.
Recommendation: make finite input EOF stop the program and settle the follower and commands.

### Secondary H01 to H04

- H01: oversized positive PIDs can show a live supervisor. Reject values outside the platform PID range before lookup.
- H02: unchanged header read errors add log events on every tick. Use an owner-approved quiet read path without deleting errors.
- H03: error logging follows rejected daily links and blocks on a rejected FIFO. A linked log directory also permits target writes.
- H03 recommendation: coordinate static logger-path checks with T01. Preserve explicit errors and the first error event.
- H04: the worker reports deletion of actual worktree `.hive/logs` and `.hive/state`, contrary to the preservation rule.

The original V3 used the worktree home and shows a real CLI log event.
That daily file is absent. No pre-deletion inventory proves all deleted contents or their ownership.
H04 is a procedure failure, not a production deletion path. This pass cannot recreate the missing historic bytes.
The secondary report also discloses an early scratch-path deviation at `/tmp/unused-no-write`.
It moved that owned file into captures. The path is absent. No preexisting file changed.
Its first safe-text probe used a wrong UTF-8 expectation. The corrected probe passes, and the failed run remains available.

## V1 runs

Owned worker copy, exact worker source:

```sh
go test -race -count=1 ./internal/tui/ -run TestHeader -v
```

```text
=== RUN   TestHeaderShowsHomeAndRefreshesStoreCounts
--- PASS: TestHeaderShowsHomeAndRefreshesStoreCounts (0.01s)
=== RUN   TestHeaderInvalidStoreShowsBoundedError
--- PASS: TestHeaderInvalidStoreShowsBoundedError (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/tui	1.013s
exit 0
```

The full output and status are in `worker-green-V1.txt`. The nominal tests do not detect H01 or H02.
The secondary count, role, read-only, lock-policy, privacy, entry, and static-store checks pass.
Its independent matrix remains in the secondary report rather than a duplicate here.

## V2 runs

```sh
go test -race -count=1 ./internal/tui/ -run 'TestKeys|TestFollow|TestDailyFile' -v
```

All eight selected worker tests pass. `worker-green-V2.txt` retains the exact output and exit 0.
The primary positive suite also passes with real files and a fixed clock:

- Initial incomplete JSON waits for LF, appears once, and retains its actual line number.
- Midnight retains both old lines and the first new-day line.
- A missing file shows `no events today`, then its new events.
- A tick-only burst of 101 events stays ordered and unique.
- Debug is the default. Four level changes enforce the expected levels and return to debug.
- Editing `gq` does not quit or move the view. Enter applies the filter.
- A rendered-only `project_id=match-gq` substring matches. Escape clears the input and applied filter.
- Follow off keeps ingestion without viewport movement. `g` shows the first line. `G` shows the end and enables follow.
- Both quit keys work, including Ctrl-C during editing.
- Normal and malformed bodies use T06 rendering without added raw control leakage.
- Malformed sizes below, at, and above 1 MiB, plus 5 MiB, settle and retain bounded rendered prefixes.
- Healthy ticks take approximately 500 milliseconds. Action ingestion produces no new log event.
- A truncated follower file gives an explicit error and settles.

`primary-positive-final.txt` records exit 0. `primary-defects-repeat.txt` records P01 to P03 and exit 1.
The primary defect command was:

```sh
go test -race -count=10 -timeout=75s ./internal/tui/ -run '^TestPV(OneOrderedConsumptionPath|RunWriterErrors|RunStringReaderEOF)$' -v
```

## V3 runs

```sh
go test -race -count=1 ./internal/tui/ -run TestRunQuits -v
```

The test passes with exit 0. `worker-green-V3.txt` retains the full output.
Ten lifecycle batches also pass. Each batch runs ten q entries and ten Ctrl-C entries.
Followers and waiting commands settle. Input errors return explicitly. No growing follower or command leak appears.
The final batch starts with five goroutines and settles at three.
`lifecycle-repeat-final.txt` retains the exact runs and exit 0.
P02 and P03 remain separate failures.

The binary came from the actual worker root, with `TMPDIR=/tmp/t08v01-go`:

```sh
go build -o /home/penguin/source/hive/.hive/tmp/coordinator/reports/t08-verify-01-captures/bin/hive .
printf q | /home/penguin/source/hive/.hive/tmp/coordinator/reports/t08-verify-01-captures/bin/hive --home /tmp/t08v01-go/cli-3p2wiry4
/home/penguin/source/hive/.hive/tmp/coordinator/reports/t08-verify-01-captures/bin/hive --json --home /tmp/t08v01-go/cli-3p2wiry4
```

Build and q exit 0. JSON refusal exits 1. Both CLI stderr streams are empty.
The q output holds the unique marked fixture home and no extra envelope.
The JSON stdout is exactly:

```json
{"outcome":"failed","phase":"validate","effects":[],"error":"TUI requires text output","result":null}
```

Binary SHA256: `fd351d2ab966b1107dafc2bb8de7ad6100c9d2db948ff8cc0e8486921b858d71`.
`cli-shell-status.txt`, stdout, stderr, metadata, and synthetic logs retain the actual run.

## V4 runs

All commands ran from `/home/penguin/source/hive/projects/hive/worktrees/T08-tui-log-view` at the exact worker.
The unique Go `TMPDIR` was outside every marked home.

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run --allow-serial-runners ./...
go test -race -count=1 ./...
go test -count=1 -run TestSTE -v .
```

Format stdout is empty. Every command exits 0. Lint reports `0 issues.`
The full race suite passes every package. The STE baseline does not change.
`gate-summary.txt` and the five `gate-*.txt` files retain exact output and status.

## Sensitivity, scope, cleanup, and limits

I reproduced assertion-red header, key/follow/daily, q-run, and JSON-refusal checks with the retained `6418b304` test-first scaffold.
The worker's red labels name the base, but that base lacks the new tests and symbols.
The fixed-clock red uses an explicit isolated inert-constructor mutation. Its historical uncommitted setup remains a provenance limit.
No compile-only red counts as evidence.
Seven primary mutations each fail on an assertion and pass after exact source restoration.
They cover both worker mutations, private Offset state, rendered filtering, the level cycle, follow tracking, and body rendering.
The secondary retains six more assertion-red/restored-green mutations.
Exact probes, patches, outputs, source hashes, and version maps remain in captures.

The full 16-file base-to-worker diff received an accounting check.
T06 adds only `NewReaderWithClock`. Original `Read`, `Follow`, rendering, and filter behavior remain unchanged.
Installed Bubbles v2.2.1 and LipGloss v2.0.6 fit Bubble Tea v2.1.0. No old Bubble Tea v1 API assumption applies.
Existing test behaviors remain represented. Fixture changes include unique homes and sufficient viewport height for the retained assertions.
`git diff --check` exits 2 only for literal terminal whitespace in V1/V2. Those frozen files remain unchanged.
The absent later accepted T05 root changes are not a source finding.

Cleanup removes only manifest-listed owned obsolete copies, binary, fixture, and Go temporary files after compact evidence retention.
No worker runtime, existing report, state, log, worktree, credential, or other agent resource is removed by this pass.
Physical terminal clipping remains a limit. Model checks do not prove every terminal size or a visible full-width header.
Hostile `Lstat`/`Open` replacement and hostile or unbounded writers remain deferred.
Uncompleted: all source corrections and any new independent verification after correction.

END VERIFIER T08-tui-log-view t08-verify-01
