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

## Independent fix verification 02

Date: 2026-10-18.
Pass: `t08-verify-02`.
Verifier: external Pi, `openai-codex/gpt-6.1-sol:high`.
Actual provider, model, and reasoning match that route. No agent or model child started.

**Verdict: FAIL. P03 and the later-path part of H03 remain open.**
This append preserves the entire 11518-byte old verdict prefix.
Its SHA256 remains `71f3ea8e02f475699331020213dcc78562b0b2c964c54e03aaf9b3751a42c843`.
No old worker evidence or report changes. No landing or checkpoint acceptance applies.

| Identity | Exact value |
| --- | --- |
| Frozen worker and sole verdict parent | `e341019bb7d318409dad4fcddae61467d35c9748` |
| Last production | `5a1c0317e205cb88828d3bcc17db5442878148e6` |
| First verdict | `223f3937d5b8c63690e41309da48d7c5f6f3e792` |
| Original worker | `e26d08bdda4eb3aad5986796c35edb7a7d8c8175` |
| Activation base | `eded8f9cb09208bfdc81334870e86479e746c183` |
| Accepted logger | `66d6cd7f1e84fdc964dd757848f0392c78ea1a48` |
| Dependency merge | `f8260e7a9e7dd5e89c23cf51c4389960dad08bf3` |
| Test-first | `a24faa51eb24970e5964bf0064a372ce0d4c4d24` |
| Real-file-output test | `6ef97095793b7ceeaa4cf60bb0df33e6928e4543` |
| Branch | `task/T08-tui-log-view` |

| Item | Result | Basis |
| --- | --- | --- |
| V1 | PASS | Counts, PID bounds, persistent errors, read-only state, safe header, and quit controls pass. |
| V2 | FAIL | Normal keys, follow, order, and constructor refusal pass. Later H03 refusal can remain invisible. |
| V3 | FAIL | q, Ctrl-C, output faults, PTY, and JSON refusal pass. Closed-pipe EOF still hangs. |
| V4 | PASS | Full actual-worktree uncached gate and TestSTE pass. |
| P01 | CLOSED | One production follower consumer. Literal A/B order and visibility pass repeatedly. |
| P02 | CLOSED | Actual errors, nil-error short writes, and real file-output errors return promptly. |
| P03 | OPEN | The adapter exempts nonterminal Fd inputs. Actual closed-pipe EOF does not stop the CLI. |
| H01 | CLOSED | Oversized values refuse before lookup. Valid and dead controls pass. |
| H02 | CLOSED | The first error remains. Unchanged ticks add no repeated error-read events. |
| H03 | PARTIAL, OPEN | Constructor guards pass. Later directory and rotation refusals can lack an error view. |
| H04 | HISTORICAL LIMIT | Fresh preservation passes. It cannot restore the old runtime bytes. |

### Scope and sensitivity

The complete report is `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t08-verify-02.md`.
Its adjacent `t08-verify-02-captures/` directory retains raw outputs, probes, logs, patches, inventories, and hashes.
I read both complete first-review reports and both worker reports through their final markers.
I read the current task, all old evidence, FIX-01, the exact saved brief, T06 contracts, and installed Bubble Tea source.
The saved brief retains SHA256 `44050b7304af32a8fe56f431f99a006cb4394b6e6ef2637e6b875f110eaef2cc`.

The activation range has 16 paths. The full new range has 37 paths: 30 imports and seven manual T08 paths.
The report and `scope-results.json` retain separate full maps.
The mechanical dependency merge and ordered parents match exactly.
All 29 nonshared imported blobs and modes equal `66d6cd7f`.
The shared root keeps original T08 RunE, T05 doctor, and all older registrations.
No source09/11 import or manual logger, T05, tracker, or imported-evidence edit appears.
Dependencies, baseline, T06 behavior, and the fixed-clock constructor stay exact.
I inspected all new assertions and changed old assertions. The worker kept every old assertion without weaker checks or skips.

The exact test-first and file-writer test trees compile and fail on behavior assertions with pre-fix production.
The frozen final tree passes those worker tests.
The exact first-review probe reproduces reverse order and stream failures on old production.
The corrected adaptation keeps literal expected order, starts actual Init, and asserts actual View visibility and order.
Ten repetitions pass. A competing-consumer mutation makes the independent and worker tests fail on assertions.
Four further relevant mutations cover output faults, string EOF, PID bounds, and persistent error reads.
All five produce assertion-red exit 1 and exact-restoration green exit 0.
Two initial driver setup errors remain retained. Neither counts as red evidence.

### P03: finite Fd input remains open

Location: `internal/tui/tui.go:280-335,430-435`.
`guardReader` exempts every Fd input. The actual CLI pipe is an `*os.File`.
Its EOF never closes `inputEnded`. Bubble Tea ends the reader, not the program.
The initial owned CLI case and three repetitions remain active past two seconds on an empty closed pipe.
The parent stops and reaps only each owned child. Cleanup status -9 is not a product exit code.
`pipe-eof.*` and `pipe-repeat*` retain the commands, output, deadlines, statuses, and reap records.

An empty regular file settles with exit 1 and an explicit cancel-reader epoll error on this platform.
It does not hang. It does not prove successful finite EOF completion.
String-reader EOF succeeds in about 500 milliseconds. Input errors remain explicit.
Recommendation: handle finite nonterminal Fd input without breaking terminal identity or cancellation.
Do not add an arbitrary blocked-reader timeout.

### H03: later refusal can remain invisible

Constructor daily-link, FIFO, and linked-log-directory cases pass with unchanged external targets and bounded explicit errors.
The imported logger also passes first-open, rotation, retry, and old-handle safeguards.
The remaining issue concerns the T08-local reporting of later logger refusal.

`late-log.txt` and `probes/late_log_test.go` retain six narrow real-model cases.
A cooperating writer installs the static obstruction after New and before the next action.
No hostile replacement occurs during a check/open interval.
First-open daily links and FIFOs yield explicit follower errors in about 0.52 seconds.
A first-open linked log directory leaves `no events today` visible without an error after 1.2 seconds.
Next-day daily-link, FIFO, and linked-directory cases also lack an explicit error view after 1.2 seconds.
All external sentinels and directory entries stay unchanged. No FIFO blocks.
The rotation test changes the level. Its blank body does not prove loss of old events.

Location: `internal/tui/tui.go:238-251` and the logger/error boundary.
`slog` drops the Handle refusal. The unchanged follower can keep waiting without an error message.
Recommendation: show the later static refusal through a bounded T08-local error view.
Keep T06 and the accepted logger exact. Add no fallback, generic read lock, event suppression, or deletion.

### Combined behavior, actual CLI, and gate

The independently repeated old checklist passes for healthy behavior.
Actual private Offset, incomplete LF, midnight, missing-file start, and 101-event bursts remain correct.
Healthy 500 ms ticks, follow-off ingestion, g/G, four-level filtering, rendered substring matching, and editing keys remain correct.
Safe bodies, control rendering, malformed caps, truncation errors, header counts, unread/nudged behavior, and privacy remain correct.
State hashes stay exact. No consume, nudge, record mutation, or generic read lock appears.
Twenty alternating q/Ctrl-C entries and ten Stop/command checks settle. Goroutines measure 3 before and 3 after.
Both quit keys work in persistent header-error state.

The actual frozen binary SHA256 is `680b7a321fcd546c98f77148e583ef5757049d7fcbf4b1874f13fd785cd1c7cf`.
Each actual domain CLI call gives explicit `--home` and a unique marked fixture home.
q and Ctrl-C exit 0, show that home, and add no envelope. Stderr is empty.
Bare JSON exits 1 with one failed envelope, phase `validate`, error `TUI requires text output`.
Custom output faults and actual `/dev/full` output return errors. The CLI file-output child exits 1 promptly.
The owned PTY enters raw mode. Byte `71` without newline causes exit 0 in about 31 milliseconds.
Its complete terminal attributes restore exactly. The parent reaps the child and closes both PTY descriptors.
This proves the changed adapter behavior, not CP4 or arbitrary terminal-size behavior.

The full actual-worktree gate passes with short outside-home `TMPDIR=/tmp/t8v2.au1ato6x`:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run --allow-serial-runners ./...
go test -race -count=1 -timeout=120s ./...
go test -run TestSTE -count=1 -v .
```

Format output is empty. Every command exits 0. Lint reports `0 issues.`.
Fresh actual-worktree V1 to V3 selections and bare-command tests also pass for their tested inputs.
The independent finite-pipe and late-error probes remain separate failures.
The baseline stays exact. New prose adds zero hard STE findings.
The verdict child changes evidence only. Focused docs checks and Go/source/dependency/baseline equality accompany delivery.

### H04 and preservation

The original cleanup names `.hive/logs`, `.hive/state`, and subdirectory `.hive/tmp/t08-impl-01`.
It does not prove deletion of every `.hive/tmp` entry. New broad wording does not enlarge that historic loss claim.
Original daily bytes remain absent without a pre-deletion content inventory. No recovery claim applies.
The old report, V1 to V4, literal terminal whitespace, and FAILED verdict prefix stay exact.

The fresh worker inventory keeps every preexisting path. The marker and original report retain separate hashes.
That inventory records type, mode, and size, not full content hashes for its own changing scratch files.
Fresh compliance does not restore historic H04 bytes.
My own no-follow runtime inventories match exactly. All 274 tracked hashes stayed exact before this append.
All 330 retained worker artifacts, both worker reports, and 4243 frozen completed root records stay exact.
I excluded concurrent T09/T11 records from that completed-record freeze.

Compact evidence existed before cleanup. A no-follow manifest lists 1130 owned temporary entries.
The obsolete root, copies, fixtures, binaries, and Go temporary data are absent.
All owned children exited or reached their finite deadline. The parent reaped each child. No helper remains.
I removed no preexisting worktree runtime root, report, log, state, archive, credential, branch, workspace, or other agent resource.
No global harness configuration, credential content, or historical attribution changed.
The separate root report gives logger source integration PASS and records FAIL for a historical phase label.

Uncompleted: P03 correction, later H03 error reporting, and fresh independent source acceptance.
Only the local-user and cooperating-writer boundary applies.
Hostile Lstat/Open replacement, malicious callbacks, and unbounded input remain deferred.
No CP4 or future checkpoint pass applies.

END VERIFIER T08-tui-log-view t08-verify-02

## Independent source verification 03

Date: 2026-10-18.
Pass: `t08-verify-03`.
Verifier: external Pi, `openai-codex/gpt-6.1-sol:high`.
The actual provider, model, and reasoning match that route. No agent or model child started.

**Verdict: PASS for the bounded T08 source delivery.**
P03 and the later-path part of H03 close. P01, P02, H01, and H02 stay closed.
H04 remains an irreversible historical provenance limit.
This result gives no root landing, CP4, or future checkpoint acceptance.

| Identity | Exact value |
| --- | --- |
| Frozen worker and sole verdict parent | `3904de4e74585cbd5992b7091c28eaea8a192a1d` |
| Previous verdict and fix02 start | `7ec0bba6084cc5f29d97f862f4145ad0b00a0f82` |
| Previous worker | `e341019bb7d318409dad4fcddae61467d35c9748` |
| Fix02 tests | `e694cd82d9037b2de9603eb7cbe85fafcfdb3ff7` |
| Fix02 production | `ecb9c46ed3d1ce3579ee841ffc64a9613049537b` |
| Final Go source | `c49d91dc0eb3a4b83967a17ef9e0d7089d49aefd` |
| Original base | `eded8f9cb09208bfdc81334870e86479e746c183` |
| Accepted logger | `66d6cd7f1e84fdc964dd757848f0392c78ea1a48` |
| Branch | `task/T08-tui-log-view` |

| Item | Proves | Result | Basis |
| --- | --- | --- | --- |
| V1 | AC1, AC3 | PASS | Counts, PID bounds, persistent errors, read-only state, sanitization, and quit controls. |
| V2 | AC2, AC3 | PASS | One follower consumer, literal visible A/B order, key modes, daily files, and explicit log errors. |
| V3 | AC1, AC3 | PASS | Actual CLI, finite EOF, stream errors, raw PTY, full terminal restoration, and JSON refusal. |
| V4 | AC4 | PASS | One full actual-worktree uncached gate at final Go source. |

The complete report is `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t08-verify-03.md`.
Adjacent `t08-verify-03-captures/` retains originals, hashes, patches, commands, raw outputs, synthetic logs, and inventories.
I read all complete prior worker and verifier reports, current task, old evidence, T06 contracts, and accepted logger contracts.
I inspected the installed Bubble Tea v2.1.0, Ultraviolet, and Linux cancelreader lifecycle path.
The saved fix02 brief retains SHA256 `107eafd0e229cdbb8dec304129c4e1f5560778d46bcd146acf5ff57b57a07d48`.

### Scope and exact history

The current fix changes exactly `internal/tui/tui.go`, `internal/tui/logview.go`, `internal/tui/fix_test.go`, and new `FIX-02.md`.
I inspected every changed line. The fix preserves every existing assertion and test without weaker checks or skips.
The complete original/import/fix01/fix02 maps contain 16, 30, seven, and four paths, respectively.
All 29 nonshared imported blobs and modes remain exact to `66d6cd7f` and the prior verified map.
The shared root preserves T08 bare entry, T05 doctor, and every old registration.
Every other original blob and mode stays exact to the previous verdict.
T06, logger, T05, store, common, output, dependencies, baseline, V1 to V4, and FIX-01 remain exact.
No source09/11 or advancing root merge appears.

Historical merge `f8260e7a9e7dd5e89c23cf51c4389960dad08bf3` has ordered parents `223f3937d5b8c63690e41309da48d7c5f6f3e792` and `66d6cd7f1e84fdc964dd757848f0392c78ea1a48`.
The current fix start is `7ec0bba6`, not that old merge parent.
The worker report's frozen-start sentence is imprecise. This append states the exact facts and preserves that report.

### V1 and V2 runs

From the actual worktree at the frozen worker:

```sh
go test -race -count=1 -timeout=45s ./internal/tui -run '^TestHeader' -v
go test -race -count=1 -timeout=45s ./internal/tui -run 'TestKeys|TestFollow|TestDailyFile' -v
```

Both selections exit 0. Full output remains in `V1.txt` and `V2.txt`.
The independent combined probes also exit 0 in `combined-probes.txt`.
Literal initial/A/B order and rendered visibility pass ten repetitions with actual Init commands active.
Real Offset, complete/incomplete LF, midnight, missing-file start, and 101-event bursts stay ordered and unique.
Follow-off ingestion, scrolling, four levels, rendered filters, editing, Enter/Escape, and safe bounded bodies pass.
T06 level matching remains case-insensitive. Rendered substring matching keeps its existing literal semantics.
Healthy ticks measure about 500 ms. Header counts refresh from 2/3 to 3/4 without domain writes.
PID bounds precede lookup and signal 0. Self/live, zero, negative, dead, and oversized controls pass.
The first header error persists without unchanged error-read feedback. Both quit keys work.
No consume, nudge, state mutation, generic read lock, or nonzero process signal appears.

The exact six later-path probes use the real logger, follower, Update, and View.
First-open log directories begin empty. Rotation opens the old day before the obstruction.
Daily links, FIFOs, and linked log directories give explicit errors at the action in both stages.
No probe blocks or changes its outside sentinel or directory count. Constructor controls stay green.
One checked `recordLog` path preserves the first bounded refusal and never logs that refusal recursively.
Repeated different refusals and ticks preserve the first error. q and Ctrl-C still quit.
A healthy level action adds exactly one event with component `tui`, action `level`, DEBUG level, and the fixed time.
Privacy, fields, safe rendering, and the 300-byte error bound remain intact.

### V3 and lifecycle runs

```sh
go test -race -count=1 -timeout=45s ./internal/tui -run '^TestRun(Quits|ResolvesFixtureHome)$' -v
```

The selection exits 0. The actual bare-command and real Cobra fault-driver probes also pass.
The absolute frozen binary is `/tmp/t8v3.35zuhuaf/hive-t08`.
Its SHA256 is `85fca839737d2fd7f83ed207d6d8dbe2aa73707e8e7e55bc42fdb0cd83dd3dd1`.
Every actual CLI case uses a unique marked fixture home, not the actual worktree or root home.
The exact previous closed-pipe probe and three repetitions exit 0 in about 0.503 seconds.
Empty regular-file EOF also exits 0 in about 0.503 seconds. Finite string EOF still passes.
Only real terminals bypass the guard. Actual non-EOF input errors remain explicit.
q and Ctrl-C exit 0, show the home, and add no envelope. Stderr is empty.
Bare JSON exits 1 with empty stderr and exactly one envelope:

```json
{"outcome":"failed","phase":"validate","effects":[],"error":"TUI requires text output","result":null}
```

The actual PTY enters raw mode. Byte `0x71` without newline causes exit 0 about 31 ms later.
The home is visible. All pre/post terminal attributes compare equal. Both PTY descriptors close after the parent reaps the child.
Custom error and nil-error short writers return errors promptly. Actual `/dev/full` stdout exits 1 with one bounded diagnostic.
Twenty finite OS-pipe q/Ctrl-C runs leave descriptor counts 7/7 and goroutine counts 2/3.
No follower, waiting command, or stream reader remains after settlement.
The retained twenty string q/Ctrl-C and ten Stop/follower-command controls pass, with goroutine counts 3/3.
All owned children exit and the parent reaps them. Arbitrary forever-blocked readers and writers stay deferred.

### Test-first replay, sensitivity, and V4

The exact `e694cd82` archive compiles with old production.
Both finite-FD tests fail assertions, including the literal two-second pipe deadline and actual regular-file epoll error.
All six checked-handler cases fail explicit error-view assertions. The frozen source passes those eight tests.
The exact later-path probe reproduces the old four failures and two passing controls, then passes all six on frozen source.
Blanket-Fd-exemption and ignored-handler-refusal mutations each compile and fail assertions.
Exact restoration passes. No compile failure or Go panic timeout counts as red. No unrelated mutation class runs.
Original probes, execution bindings, patches, red/green output, and restoration hashes remain in captures.

The full gate ran once from the actual worktree with short outside-home `TMPDIR=/tmp/t8v3.35zuhuaf`:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run --allow-serial-runners ./...
go test -race -count=1 -timeout=180s ./...
go test -run TestSTE -count=1 -v .
```

Format output is empty. All commands exit 0. Lint reports `0 issues.`.
All packages pass. Earlier T03/T04/T05, log, home, store, and T06 guards remain green.
The baseline stays exact. New instruction prose has zero hard findings with both glossaries.
After this evidence-only append, source/dependency/baseline equality and focused docs checks apply.
No changed Go source requires a repeated full gate.

### Preservation and limits

The exact old VERIFIED prefix is 22412 bytes with SHA256 `626b8eda52320c775e58f388b9e320c1a647d08278d7479657f1f5ce83fef705`.
This append preserves every prefix byte, all old V/FIX files, old terminal whitespace, and every worker report.
All 275 tracked hashes remain exact before the append. All 924 preexisting runtime entries and stable contents remain exact.
Fix02 inventories retain all 660 before entries among 920 after entries. All 333 stable hashes match.
All six fix02 obsolete cleanup targets are absent on read-only inspection. No unowned cleanup occurs.
Compact evidence preceded cleanup. The no-follow manifest lists 764 owned entries below `/tmp/t8v3.35zuhuaf`.
I removed only that listed root, including the obsolete copies, fixtures, binary, and TMPDIR. That root is absent.
Every worktree runtime root, old report, log, state, and preexisting entry remains unchanged.
Root and other tasks can advance independently. This pass makes no root freeze claim.

H04 proves old removal of `.hive/logs`, `.hive/state`, and subdirectory `.hive/tmp/t08-impl-01`.
Whole `.hive/tmp` deletion is not proved. Missing historical bytes remain an irreversible limit.
This pass synthesizes no old logs and makes no old-cleanup compliance claim.
Only the local-user, cooperating-writer, static-type boundary applies.
Hostile Lstat/Open replacement, malicious callbacks, and arbitrary forever-blocked input/output remain deferred.
Physical terminal clipping remains a limit. No landing or checkpoint acceptance applies.
Recommendation: accept this bounded source delivery and keep the H04 provenance limit explicit.

END VERIFIER T08-tui-log-view t08-verify-03
