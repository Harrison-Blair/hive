# T05 verification 01

| Item | Value |
| --- | --- |
| Pass | `t05-verify-01`, primary |
| Date | 2026-10-18 |
| Role | fresh independent verifier |
| Route | `pi openai-codex/gpt-6.1-sol:high` |
| Route evidence | `PI_PROVIDER`, `PI_MODEL`, `PI_REASONING_LEVEL` agree |
| Branch | `task/T05-herdr-client-and-doctor` |
| Base | `84da9819b31be83dbe71c36d620eadf3749e1f9f` |
| Frozen worker | `30bd4fc4a5a504ca5ea1a79dc2023202c117ef7e` |
| Worker tree | `46843e65c69df9dd402aa424d89fbfc07475c1de` |
| Initial HEAD | exact frozen worker, clean |
| Verdict | **FAIL** |

This is the primary area report and a provisional full verdict.
The coordinator must combine both area reports before task acceptance.
The second area report remains pending at this report's creation.

## Authority and scope

I read the task, architecture, tracker, instructions, glossaries, V1–V6, and worker report.
The worker report ends with `END WORKER T05-herdr-client-and-doctor t05-impl-01`.
I inspected every changed file in the complete base-to-worker diff.
I checked source imports with `go list` and source search.
No `internal` package imports Cobra or `cmd`.

The herdr authority includes `reference/herdr/README.md`, `protocol.md`, `environment.md`, `errors.md`, `api/session.md`, and protocol 22 schema definitions.
The root D5 clarification at `7ae3fb8a` applies to this historical task.
Complete frames include LF in the 1048576-byte limit.
Doctor can create empty state and lock metadata, but no domain record.
Neither choice is an open finding.

Only the approved local-user and cooperating-writer boundary applies.
Hostile replacement between `Lstat` and `Open` remains deferred for T01.
Known T01 home fixes and the shared store writer fix remain outside T05 ownership.

## Provisional verification items

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | FAIL | Existing suite passes. Independent assertions expose P05-F01 through P05-F05. |
| V2 | AC1 | PASS | Actual pane, live protocol 22, matched ping log IDs. |
| V3 | AC2 | PASS | Required runner suite and assertion-red evidence pass their expected checks. Deeper area checks remain pending. |
| V4 | AC3 | PASS | Required doctor suite and assertion-red evidence pass their expected checks. Deeper area checks remain pending. |
| V5 | AC3, AC4 | PASS | Required transcript, actual JSON statuses, and home-free version checks pass. |
| V6 | AC5 | PASS | Four gates pass, including uncached full race tests. |

P05-F06 is an additional CLI logging failure under AC2 and `AGENTS.md`.
A passing transcript does not remove that failure.

## Findings

All six findings belong to the T05 worker.
The locations below refer to the frozen worker, not an isolated mutation.

### P05-F01 — failed result decoding changes the supplied result

- Authority: AC1, `result` stays unchanged on a failure.
- Location: `internal/lib/herdr/client.go:248–251`, `settleReply`.
- Probe: `TestPrimaryResultPreservedOnDecodeError`.
- Observation: the supplied value starts as `{Keep:before Count:7}`.
- The peer sends `{"keep":"changed","count":"not an integer"}` as the result.
- `Call` returns `bad reply`, but the value becomes `{Keep:changed Count:7}`.
- Minimum fix: decode into a separate value first. Assign the supplied result only after complete success.
- Add preservation assertions for structs, maps, slices, and nested pointers.

### P05-F02 — typed nil parameters reach the wire as null

- Authority: AC1, `params` is never null. Empty parameters use `{}`.
- Location: `internal/lib/herdr/client.go:105–108`.
- Probe: `TestPrimaryParamsNeverNull`.
- Observation: nil maps, nil struct pointers, and `json.RawMessage("null")` send `"params":null`.
- An untyped nil correctly sends `{}`.
- Minimum fix: normalize the encoded null value before frame sizing and before any gate request or write.
- Add wire assertions for typed nil and raw null values.

### P05-F03 — the protocol gate ignores a waiting call's deadline

- Authority: AC1 total deadline and earlier context deadline. The untrusted-input rules require bounded network calls.
- Location: `internal/lib/herdr/protocol.go:46–60`.
- Probe: `TestPrimaryGateWaitDeadline`, including ten independent repetitions.
- Observation: a first call holds the global mutex during its silent ping.
- A second call with a 50 ms deadline waits approximately 500 ms before it returns.
- All ten repetitions fail the independent 200 ms upper bound.
- Minimum fix: keep gate state under a short mutex operation.
- Wait for gate completion with a context-aware select outside that mutex.
- Preserve the required per-path cached success or failure.

### P05-F04 — null fields defeat envelope exclusivity

- Authority: AC1 rejects a response that contains both `result` and `error`.
- Location: `internal/lib/herdr/client.go:228–243`.
- Probe: `TestPrimaryNullFieldExclusivity`.
- Observation: `result:null` with an error returns the server error instead of `bad reply`.
- A result object with `error:null` succeeds.
- Pointer decoding loses the distinction between an absent field and a present null field.
- Minimum fix: check field presence separately from its decoded value.
- Reject both fields before result decoding, including either null combination.

### P05-F05 — failed socket calls lack error log events

- Authority: AC1 socket logging, `AGENTS.md` Logging, and architecture Logging.
- Location: `internal/lib/herdr/client.go:121–149,269–279` and `protocol.go:23–39`.
- Probe: `TestPrimaryErrorEvents` with a real log writer and a socket peer.
- Observation: server errors and malformed replies have no correlated `error` field.
- EOF and deadline failures leave only a request event, with no reply or error event.
- Minimum fix: log each failed call with its ID, method, duration, and error.
- Keep bodies bounded. Include protocol failures and failures before the reply event.

### P05-F06 — doctor omits external process log events

- Authority: AC2 requires process start/end events. `AGENTS.md` requires each external process to be logged.
- Location: `internal/doctor/run.go:18–25`, before the runner calls in `checks.go`.
- Probe: `TestPrimaryDoctorProcessEvents`, plus actual live doctor logs.
- Observation: doctor passes its logger only to herdr, not to the runner context.
- With actual `git` and `sh`, the isolated assertion reports `start=0 end=0, want 7 each`.
- The live doctor log also contains zero process start/end events.
- Minimum fix: give the runner the doctor logger through `run.WithLogger` before the checks execute.
- Keep an integration assertion that checks the resulting file events.

## Runs and captures

Captures remain in the worktree at `.hive/tmp/t05-verify-01/`.
Each capture contains the actual output. The table gives the process status.

| Command or check | Status | Capture |
| --- | --- | --- |
| `go test -race -count=1 ./internal/lib/herdr/ -v` | 0 | `v1.txt` |
| `go test -race -count=1 ./internal/lib/run/ -v` | 0 | `v3.txt` |
| `go test -count=1 ./internal/doctor/ -v` | 0 | `v4.txt` |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | 0, no output | `v6.txt` |
| `go vet ./...` | 0 | `v6.txt` |
| `golangci-lint run ./...` | 0, `0 issues.` | `v6.txt` |
| `go test -race -count=1 ./...` | 0 | `v6.txt` |
| Build and actual V2/V5 commands | 0 for the probe script | `live-checks.txt` |
| All first independent herdr probes | 1, five failing test groups | `primary-probes.txt` |
| Gate/cancellation/transport checks, `-count=10` | 1, gate failure ten times | `timing10.txt` |
| Actual-call IDs and default deadline | 0 | `id-default.txt` |
| Independent passing checks after cap restoration | 0 | `primary-green.txt` |
| Isolated doctor process logging assertion | 1 | `doctor-log-probe.txt` |

The independent probes use `go test -race -count=1` on an isolated archive of the frozen worker.
They mock only the socket peer. They use the actual herdr package, runner, log writer, `git`, and `sh`.
The preserved probe sources are `primary-probe-source.txt` and `doctor-log-probe-source.txt`.

Passing probes cover these behaviors:

- Literal complete-frame sizes 1048575, 1048576, and 1048577 in each direction.
- The oversized request accepts zero connections, including no preliminary gate request.
- UTF-8 request sizing and bounded 10 MiB reply rejection.
- Typed error code/message, wrong IDs, both non-null fields, neither field, and protocol 21 rejection.
- Per-path gate caching and one connection per request.
- EOF, incomplete JSON, an HTTP peer, arrays, slow streams, and total deadlines.
- Earlier caller deadlines and cancellation after the socket accepts the request.
- The actual default deadline, observed at approximately 10 seconds.
- Twenty concurrent callers with one eight-hex prefix and one contiguous counter.
- Two child processes that make actual socket calls with distinct prefixes.
- Request/reply IDs, method, duration, 4096-byte bodies, and truncation flags.

### Assertion-red evidence

I recreated each historical red commit in a separate ignored archive copy.
No worker file changed.

| Item | Isolated commit | Command | Result |
| --- | --- | --- | --- |
| V1 | `1c45fd8c` | `go test -race -count=1 ./internal/lib/herdr/ -v` | exit 1, 20 failing assertions/groups |
| V3 | `0679d1be` | `go test -race -count=1 ./internal/lib/run/ -v` | exit 1, 8 failing assertions/groups |
| V4 | `f558034a` | `go test -race -count=1 ./internal/doctor/ -v` | exit 1, 13 failing assertions/groups |

The captures are `v1-red.txt`, `v3-red.txt`, and `v4-red.txt`.
These runs fail on assertions, not compilation.
I also changed both socket caps to 2097152 in the isolated worker copy.
`TestPrimaryFrameBounds` then fails for both 1048577-byte frames.
The request probe records two accepted connections instead of zero.
The capture is `cap-mutation-red.txt`.
The same independent bounds pass after cap restoration.

### Live transcript checks

The binary was `.hive/tmp/t05-verify-01/hive-primary`, built from the frozen worker.
V2 used an explicit, unique marked fixture home.
The live ping answered with protocol 22.
The request and reply share ID `bd3800b4-1`.
V5 text statuses are 0, 1, 1, 0, as required.
JSON doctor succeeds with status 0 and fails with status 1 when the socket variable is absent.
Doctor prints all 12 check objects with the required fields.
Version prints the five-field envelope and four version fields.
Plain and JSON version both succeed outside a home with invalid home selectors.
Neither version call creates runtime files there.

I printed only presence or absence for named herdr environment variables.
`HERDR_ENV`, `HERDR_PANE_ID`, `HERDR_SOCKET_PATH`, `HERDR_WORKSPACE_ID`, and `HERDR_TAB_ID` were present.
`HERDR_TERMINAL_ID` was absent.
I inspected and printed no credential content.
I started no harness or model agent.

## Preservation, cleanup, and limits

All 168 initial tracked files pass the saved SHA256 checks before the verdict edit.
The isolated worker copy also matches all frozen tracked files after cap restoration.
Only this `VERIFIED.md` is an authorized tracked change.
The delivery checks show only this metadata change from the frozen worker.

I removed the three unique live fixture homes after capture.
I removed the ignored binary and four own archive copies after recording their evidence.
The probe sources and compact captures remain for reproduction.
Existing reports, logs, runtime directories, credentials, and worktrees remain unchanged.
Only subprocess groups that the required runner tests created received signals.
I signalled no PID 0, PID 1, parent, or unrelated process.
No global setting, source, V-file, glossary, tracker, or baseline changed.

The deeper process and doctor probes belong to the second verifier and remain uncompleted here.
This report does not accept the task or the T01/shared-store source changes.

END VERIFIER T05-herdr-client-and-doctor t05-verify-01

---

# T05 verification 02 — 2026-10-18

**FAIL.** V1 still fails on P05-F01 and P05-F05.
V2, V3, V4, V5, and V6 pass separately.
The imported store fix closes S05-D01.
Recommendation: return the two remaining cases to the worker. Do not land T05.

| Item | Value |
| --- | --- |
| Pass | `t05-verify-02` |
| Role | Fresh independent source verifier. No implementation. |
| Actual route | `PI_PROVIDER=openai-codex`, `PI_MODEL=gpt-6.1-sol`, `PI_REASONING_LEVEL=high` |
| Original base | `84da9819b31be83dbe71c36d620eadf3749e1f9f` |
| First worker | `30bd4fc4a5a504ca5ea1a79dc2023202c117ef7e` |
| First verdict | `b361a0c80bf84c65a96bcaf9e975847cd193822d` |
| Fixed source | `ff1ef0f7d98f00b5fd82821e264a173c3eaa5faf` |
| Final worker evidence | `fb19808bf8c8eb051434bba683e5b79551a5342a` |
| Prerequisite | `47da8abb152732d4667707ec98ef2a682e5421a4` |
| Exact tested sync | `5a7e1264741988e3bf72da204c48963c71ad76b5` |
| Initial state | Clean exact local and direct remote task tip. |
| Branch | `task/T05-herdr-client-and-doctor` |
| Separate verdict commit | Evidence-only child of the tested sync. The root report records its full ID after commit. |
| Root report | `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-02.md` |
| Captures | `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-02-captures/` |

I read the complete prior primary and secondary verdicts, tracked FAIL, V1–V6, and both worker reports.
The fixed worker report ends with `END WORKER T05-herdr-client-and-doctor t05-fix-01`.
I read the original/current task, instructions, glossaries, architecture, tracker, API reference, and untrusted-input rules.
I read `later-fix-land-01.md` through `END LATER FIX LAND 01`.
The ordinary sync has exactly the two parents above. Its source receipts match.
The CP1 PASS receipt remains a prerequisite. I did not repeat CP1.

Only the local-user and cooperating-writer boundary applies.
Hostile replacement between `Lstat` and `Open` stays deferred for T01.
No new threat class supplies a finding.
All added assertions use isolated copies. Actual worktree source stays frozen.

## Full-range review

The review covers `84da9819..5a7e1264` and the owned fixes from `b361a0c8..fb19808b`.
The root report lists all 89 full-range changed files.
`full-range-classification.txt` records each path, its owner, and its SHA256.
Every imported path equals the prerequisite object. Every T05 path equals the final worker object.
Imported T03, store, home, and specification changes remain independently reviewed prerequisite content.
The owned source still equals `ff1ef0f7`. Only evidence wording changes after `e98f8331`.

The owned fix changes these four production/test pairs:

- `internal/lib/herdr/client.go`, `internal/lib/herdr/client_test.go`
- `internal/lib/herdr/protocol.go`, `internal/lib/herdr/protocol_test.go`
- `internal/lib/run/run.go`, `internal/lib/run/run_test.go`
- `internal/doctor/run.go`, `internal/doctor/checks_test.go`

The `cmd/root.go` sync preserves both doctor/task imports and registrations.
Actual help lists doctor, task, logs, project, and version. No command is lost.
No internal package imports Cobra or `cmd`.

## Verification items

| Item | Proves | Verdict | Result |
| --- | --- | --- | --- |
| V1 | AC1 | FAIL | Supplied tests pass. Null-preservation and direct-Ping error assertions fail. |
| V2 | AC1 | PASS | Actual pane, protocol 22, matched request/reply ID `43e27be9-1`. |
| V3 | AC2 | PASS | Fresh race suite and complete runner checklist pass. |
| V4 | AC3 | PASS | Fresh doctor suite, complete area checklist, and imported store checks pass. |
| V5 | AC3, AC4 | PASS | Actual text statuses `0,1,1,0`, JSON statuses, version, and home-free version. |
| V6 | AC5 | PASS | Four uncached root gates pass. Format empty, vet 0, lint `0 issues.`, race tests 0. |

### P05-F01 remains OPEN, partially fixed

Location: `internal/lib/herdr/client.go:277–282`, `settleReply`.
Non-null struct, map, slice, and nested-pointer failures now preserve the supplied value.
The new null branch still decodes directly into the caller.

An ordinary caller-owned decoder implements the supported `json.Unmarshaler` interface.
It resets its own count, rejects null as a missing required value, and accepts a valid integer.
It touches no global value, file, credential, or other object.
The Go API calls this method for null. It requires no receiver preservation after an error.
AC1 requires `Call` to preserve the caller on failure.

The direct `json.Unmarshal` control observes the reset and validation error.
The actual socket call then receives a matched `result:null` envelope:

```text
Call result:null: err=bad reply: count is required, value={Value:0 Valid:false}
caller changed on failure: {Value:0 Valid:false}, want {Value:7 Valid:true}
--- FAIL: TestVerify02NullFailurePreservesCaller (0.00s)
```

The same destination preserves its value after a non-null decode error.
Result `9` correctly assigns `{Value:9 Valid:true}`.
Existing plain-struct null preservation and map/slice null-to-nil tests pass.
These controls establish actual behavior and the intended supported API before classification.
The worker's hostile-decoder exclusion does not exclude ordinary local validation.

Recommendation: make null decoding transactional too. Preserve successful null behavior for supported destinations.
I implemented no fix. The assertion remains red at the frozen source.
This is remaining P05-F01, not a new broad callback threat class.

### P05-F05 remains OPEN, partially fixed

Location: `internal/lib/herdr/protocol.go:30–35`, `Ping`.
`Call` now logs its actual failures with bounded errors and correlated metadata.
Direct `Ping` checks the result type/protocol after `Call` succeeds.
Those errors return without an error event. Doctor uses this direct path.

```text
returned=herdr protocol 21, hive needs 22
no correlated actual Ping error: herdr protocol 21, hive needs 22
returned=bad reply: result type "not-pong", want pong
no correlated actual Ping error: bad reply: result type "not-pong", want pong
--- FAIL: TestVerify02DirectPingProtocolErrorLogged (0.00s)
```

Each case logs only the matched request/reply, with no actual error field.
The actual frozen CLI with a protocol-21 fixture exits 1 and prints the protocol diagnostic.
Its complete invocation logs two ping events and zero error fields.
`cli-protocol-red.txt` records the actual JSON output and failed correlated-error assertion.
This is an existing protocol-error checklist case, not a new probe class.

Recommendation: log direct-Ping result errors with the actual ping ID, method, duration, and bounded error.
I implemented no fix. These assertions remain red at the frozen source.

## Every original finding

| Finding | Disposition | Evidence |
| --- | --- | --- |
| P05-F01 | OPEN, partial fix | Ordinary null validation still changes the caller. Other destination cases pass. |
| P05-F02 | CLOSED | Typed nil pointer/map/slice and raw null encode `{}` before sizing/gate/writes. |
| P05-F03 | CLOSED | Waiting callers honor their own deadlines. Ten rounds pass. Per-path cache remains. |
| P05-F04 | CLOSED | Presence includes null. Both/neither refuse before assignment. |
| P05-F05 | OPEN, partial fix | `Call` failures log correctly. Direct-Ping protocol/type errors still lack error events. |
| P05-F06 | CLOSED | Same defect as S05-F03. Actual CLI records seven starts/ends. |
| S05-F01 | CLOSED | Leader-exit cancellation/deadlines stop owned groups. Ten bounded-reap rounds pass without cleanup signals. |
| S05-F02 | CLOSED | Name/dir controls retain one physical error line. |
| S05-F03 | CLOSED | Same defect as P05-F06. Six failed starts also receive actual error end events. |
| S05-F04 | CLOSED | Actual canceled/timeout errors equal logged errors. Ten cancellation rounds pass. |
| S05-D01 | CLOSED, imported | Exact accepted store and eight actual CLI FIFO/socket fixtures pass. |

Store paths equal `ccfe0020a193a91c1493d2d4750e55f2efb81798`.
The complete T22 report ends with `END T22 WALK TYPES VERIFY 01` and closes S05-D01.
Root merge `16bbd00f` and prerequisite `47da8abb` contain that candidate.
Doctor retains its mandated `Store.WalkRecords` call. No store overlay or edit occurs.

## Checklist and meaningful assertions

The root report gives each complete two-area checklist entry and its fresh result.
All old primary and secondary probes run again, including both actual default deadlines.
The existing timing, cancellation, concurrent-runner, and concurrent-store checks pass ten rounds each.

Fixed literal frames include LF and use 1048575, 1048576, and 1048577 bytes.
Oversized requests accept zero connections, including no gate request.
Reply bounds, UTF-8, wrong IDs, slow streams, cached failures, and result-preservation cases pass except the stated null case.
Twenty actual concurrent calls share one eight-hex prefix and contiguous counter.
Two actual child processes have distinct prefixes.
Request/reply logs retain IDs, methods, durations, 4096-byte bodies, and truncation flags.

Actual runner checks cover stdin EOF, argv, cwd, exit code 23, no shell interpolation, separate groups, and stopped groups.
Each stream retains literal bounds and drains 8 MiB.
Actual default runner time is `30.000744889s`, with `process timeout`.
Actual inherited-pipe deadline/cancel checks return near 100 ms and reap owned children within the bound.
Name and directory controls stay on one line. Actual returned/logged canceled and timeout errors agree.

Doctor checks retain literal order, statuses, metadata-only state effects, and every domain byte.
Static linked/special/malformed/oversized/unreadable paths fail safely.
The held lock returns `store lock timeout` after `2.000319531s`.
Actual special collection paths fail with their exact paths. JSON-leaf special paths report `record invalid`.
Safe dependency fixtures cover gh version/auth outcomes and harness discovery without model execution.
Actual CLI checks cover all twelve objects, JSON/text statuses, home-free version, and command preservation.

The record-only signal seam runs in a new session and sends no signal.
The old one-call expectation detects two calls to the same owned group after the drain fix.
The corrected recorder accepts one or two calls only to that same owned negative target.
Cancellation after return records no signal. Impossible supplied-PID cases remain synthetic and syscall-free.
This harness adjustment is not a production finding or acceptance-red evidence.

| Run | Status | Capture |
| --- | --- | --- |
| Supplied V1/V3/V4 suites, uncached | 0 each | `v1.txt`, `v3.txt`, `v4.txt` |
| All old primary probes plus null controls | 1, null-preservation assertion | `primary-fresh.txt` |
| All unchanged old primary/secondary probes | 0 | `primary-restored-green.txt`, `secondary-fresh.txt` |
| Current fix tests on historical `6b063bb` | 1, eleven intended failing groups | `repair-red.txt` |
| Same fix tests on frozen source | 0 | `repair-green.txt` |
| Historical V1/V3/V4 assertion-red replays | 1 each, no compile/setup failures | `v1-historical-red.txt`, `v3-historical-red.txt`, `v4-historical-red.txt` |
| Isolated socket/stream caps changed to 2097152 | 1, literal 1048577 assertions | `frame-mutation-red.txt`, `stream-mutation-red.txt` |
| Isolated doctor order swap | 1, literal order assertions | `order-mutation-red.txt` |
| Exact sources restored, same three checks | 0 | `mutation-restored-green.txt` |
| Direct-Ping protocol/type logging | 1, actual error assertions | `ping-error-probe.txt` |
| Actual CLI V2/V5, process events, store cases | 0 for driver | `cli.txt` |
| Actual CLI protocol logging | 1, missing correlated error | `cli-protocol-red.txt` |

No compile-only or setup failure counts as assertion-red.
P05-F01/P05-F05 do not regain green on frozen source. Their fixes and fresh independent rechecks remain uncompleted.
Other mutations restore to green before cleanup.

## Actual CLI and gate

The binary comes from the frozen sync. It reports commit `5a7e12647419`, with `modified:false`.
Its SHA256 is `e28fc981c66161587e9df85b96e4dcd3bbc59da59a5c99a0f0bb268199e2bd3f`.

```text
ok herdr socket: /home/penguin/.config/herdr/herdr.sock protocol 22
exit 0
fail herdr socket: herdr socket: /tmp/none.sock: no such file
exit 1
fail herdr socket: HERDR_SOCKET_PATH is not set
exit 1
{"outcome":"ok","phase":"complete","effects":[],"error":null,"result":{"version":"dev","commit":"5a7e12647419","modified":false,"go":"go1.27.2-X:nodwarf5"}}
exit 0
process starts 7 ends 7 failed starts 6
```

This block selects relevant lines. `cli.txt` retains every actual doctor line and the matched V2 ping events.
The live pane check gives `ok`, with `wA9:p1`.
The private PATH permits only Git/shell for live health checks and only Git for failed-start checks.
No actual `gh auth status` or model executable runs. No credential content is read.

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run ./...
go test -race -count=1 ./...
```

All commands run from the actual frozen task worktree root and exit 0.
Format prints nothing. Lint prints `0 issues.`. Every race package passes.
The complete actual output is in the root report and `gate.txt`.
The source gate does not remove the independent V1 failures.

## Preservation, cleanup, and limits

All 205 frozen tracked paths match their saved SHA256 values before this append.
The old verdict prefix has 11955 bytes and SHA256:
`91e543bdea128a2756aa1efe814f8225938d22aac7ed0a3c18aed76952d5b6af`.
This append preserves each byte. All V-files stay exact.
Only this evidence file enters the separate verdict commit.
The root report records its ID, exact tested parent, and non-force task-branch push receipt.

Cleanup removes only owned source copies, unique fixture homes, and the binary.
Captures, scripts, probe sources, old reports, logs, credentials, worktrees, and other agents remain.
Actual signals target only fixture-owned groups, never PID 0/1, the parent, or unrelated processes.
Historical red sleepers use fixed 30-second durations, which elapsed before cleanup.
No production source, worker evidence, tracker, glossary, STE baseline, global setting, `dev`, or `main` changes.
No worktree removal or other agent control occurs.

Hostile file replacement, hostile PID reuse, and uninterruptible kernel sleep remain outside the approved boundary.
The direct wait after `SIGKILL` cannot prove a bound for uninterruptible kernel sleep.
These limits do not supply either finding.
No store dependency remains open. T05 acceptance and landing remain uncompleted because V1 fails.

`END VERIFIER T05-herdr-client-and-doctor t05-verify-02`

---

# T05 verification 03 — 2026-10-18

**FAIL.** V1 fails on the remaining referenced-storage case of P05-F01.
V2, V3, V4, V5, and V6 pass separately. P05-F05: CLOSED.
Recommendation: return P05-F01 to the worker. Do not land T05.

| Item | Value |
| --- | --- |
| Pass | `t05-verify-03` |
| Role | Fresh independent verifier. No implementation. |
| Actual route | `PI_PROVIDER=openai-codex`, `PI_MODEL=gpt-6.1-sol`, `PI_REASONING_LEVEL=high` |
| Independent session | `01a14e2e-1b4e-72f1-a483-89b51e2f85db` |
| Exact tested worker | `6d034695d3fbe3c92b09bb6f6fda53a2ae4fcd29` |
| Source | `b39b4926f4a916a1894fbddb224b1eed7e52eca2` |
| Second fix initial | `75d0d2effd29a6472db67fe2fc6acd4eb14fc1ef` |
| Original base | `84da9819b31be83dbe71c36d620eadf3749e1f9f` |
| Prerequisite | `47da8abb152732d4667707ec98ef2a682e5421a4` |
| Initial state | Clean. Local and direct remote task tips equal the exact worker. |
| Branch | `task/T05-herdr-client-and-doctor` |
| Separate verdict commit | Evidence-only child of the exact tested worker. The root report records its full ID after commit. |
| Root report | `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-03.md` |
| Captures | `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-03-captures/` |

I read all three complete prior root verdicts, tracked history, V1–V6, current/original task, and worker reports.
The current worker report ends with `END WORKER T05-herdr-client-and-doctor t05-fix-02`.
The provenance report ends with `END LATER FIX LAND 01`.
I read the instructions, glossaries, tracker, API reference, CP1 PASS, and existing checklist/probe sources.
Both original review areas are complete. I did not repeat CP1.

Only the local-user and cooperating-writer boundary applies.
Hostile replacement between `Lstat` and `Open` stays deferred for T01.
Added probes cover only the changed null transaction and direct-Ping error paths.
No malicious global-effect guarantee or future workspace/session type scope supplies a finding.

## Full-range review and prerequisites

I inspected the whole `84da9819..6d034695` diff and both fix ranges.
The fix ranges are `b361a0c8..fb19808b` and `75d0d2ef..6d034695`.
The root report lists all 89 changed paths. The captures retain the complete diffs and per-path hashes.
Each imported object equals prerequisite `47da8abb`. All Go source equals `b39b4926`.
All store paths equal accepted `ccfe0020a193a91c1493d2d4750e55f2efb81798`.
The imported home, store, T03, and docs remain independently checked prerequisites.
No imported dependency remains open.

The second fix changes `client.go`, `protocol.go`, their tests, and V1/V2/V5/V6 only.
The runner and doctor source remain exact from verification 02.
The root command equals the mechanical union of doctor/task imports and registrations.
Actual help retains doctor, task, logs, project, and version. No internal package imports Cobra or `cmd`.
Root `a21f6be1` includes unrelated T04 work and later docs. This frozen task branch needs no unrelated T04 merge.

## Separate verification items

| Item | Proves | Verdict | Actual result |
| --- | --- | --- | --- |
| V1 | AC1 | FAIL | Supplied tests and exact prior probes pass. Null reference preservation still fails. |
| V2 | AC1 | PASS | Actual pane `wA9:p1`, live protocol 22, matched ping ID `d8e2596d-1`. |
| V3 | AC2 | PASS | Fresh race suite and complete independent runner checklist pass. |
| V4 | AC3 | PASS | Fresh suite, both area checklists, seven process starts/ends, and imported store checks pass. |
| V5 | AC3, AC4 | PASS | Actual text statuses `0,1,1,0`, JSON results, version, and home-free version pass. |
| V6 | AC5 | PASS | Full uncached gate from the actual task worktree root. Format empty, vet 0, lint 0, race 0. |

### P05-F01 remains OPEN, partially fixed

Location: `internal/lib/herdr/client.go:321–329`, `decodeResult`.
The exact prior ordinary null decoder now preserves `{Value:7 Valid:true}` on error.
Its non-null error and valid `9` controls pass.
Plain struct null, map/slice null-to-nil, and successful custom null behavior pass.
All standard non-null struct/map/slice/nested-pointer failure and assignment controls pass.

The new null path copies the top-level value but shares pointer-backed fields with the caller.
The added ordinary decoder owns a count pointer and a `Valid` boolean.
It reuses count storage, resets it before validation, and rejects null as a missing count.
It touches no global, file, credential, signal, or other external object.
The direct `encoding/json` control establishes this supported behavior.
The matched socket response is `result:null`. The caller starts with count 7 and `Valid:true`.

```text
=== RUN   TestVerify03NullReferenceFailurePreservesCaller
    verify03_null_reference_test.go:34: null err=bad reply: count is required count=0 valid=true same-pointer=true
    verify03_null_reference_test.go:36: caller changed on failure: count=0 valid=true, want count=7 valid=true and original pointer
--- FAIL: TestVerify03NullReferenceFailurePreservesCaller (0.00s)
```

The same assertion fails on the second-verdict source and this worker, also in ten rounds.
The same destination preserves count 7 after a non-null error and assigns 9 on success.
This is the existing preservation class and the reference handling that this fix changes.
It does not demand rollback of malicious global effects.
Recommendation: isolate caller-owned referenced state during null decoding. Assign only after complete success.
Keep successful null controls. Do not bypass supported user decoders.
I implemented no fix. This assertion remains red. Its fix and restored green run remain uncompleted.

### P05-F05 — CLOSED

The exact direct-Ping protocol-21 and wrong-pong assertions fail on `75d0d2ef` and pass on `6d034695`.
The request, reply, and error events share the actual request ID, method, and duration.
A long direct-Ping error keeps 4096 bytes, valid UTF-8, and `truncated:true` in ten rounds.
The same bounded-error assertion fails on the second-verdict source.
Original `Call` errors and actual-writer body checks remain green.
The actual doctor protocol-21 fixture also passes the correlated-error assertion after a prior-source assertion failure.

```text
returned=herdr protocol 21, hive needs 22
returned=bad reply: result type "not-pong", want pong
actual id=984d9566-1 method=ping duration=1 error-bytes=4096 valid-UTF8=true truncated=true
```

The root report and captures retain the complete actual events and CLI output.
Doctor still exits 1 with `herdr protocol 21, hive needs 22`.

## Every original finding

| ID | Disposition | Fresh evidence |
| --- | --- | --- |
| P05-F01 | OPEN, partial fix | Exact prior scalar case passes. Referenced custom null failure changes the count. |
| P05-F02 | CLOSED | Typed nil pointer/map/slice and raw null encode `{}` before sizing, gate, or writes. |
| P05-F03 | CLOSED | Own-context gate waits pass ten rounds. Per-path pass/failure caching remains correct. |
| P05-F04 | CLOSED | Null counts as field presence. Both/neither reject before assignment. |
| P05-F05 | CLOSED | Exact prior probes, direct-Ping error bounds, and actual CLI correlation pass after assertion-red replays. |
| P05-F06 | CLOSED | Same defect as S05-F03. Actual CLI logs seven starts and seven ends. |
| S05-F01 | CLOSED | Owned leader-exit children stop/reap near 100 ms. Ten rounds pass without cleanup signals. |
| S05-F02 | CLOSED | Process-name and directory controls retain one physical diagnostic line. |
| S05-F03 | CLOSED | Seven actual ends include six failed starts with their errors. |
| S05-F04 | CLOSED | Returned cancellation/timeout errors equal logged errors. Ten cancellation rounds pass. |
| S05-D01 | CLOSED, imported | Exact accepted store. Eight actual CLI FIFO/socket fixtures reject their complete paths. |

## Checklist, actual runs, and test strength

The root report gives every established primary/secondary checklist item and its fresh result.
All old probes rerun, including actual 10-second and 30-second defaults.
Fixed complete frames include LF and use literal 1048575/1048576/1048577 expectations.
Oversized requests accept zero connections. UTF-8, large rejection, wrong IDs, EOF/HTTP/arrays, and slow streams pass.
Protocol gates, cancellation, shared IDs, two child prefixes, and bounded actual-writer bodies remain correct.
The runner drains 8 MiB with literal per-stream caps. Stdin EOF, argv, cwd, and exit code 23 pass.
Owned/stopped groups, inherited pipes, returned/logged errors, and ten concurrent runner/store rounds pass.
Doctor retains literal order, statuses, required flags, metadata-only effects, domain bytes, static path refusals, and the two-second lock bound.
Fake gh alternatives and harness discovery require no real authentication or model execution.
Impossible PID metadata stays recorder-only in a new session. The public runner accepts no supplied PID.

| Run | Status | Capture |
| --- | --- | --- |
| Supplied V1/V3/V4, race and uncached | 0 each | `v1.txt`, `v3.txt`, `v4.txt` |
| Exact prior null/Ping assertions, prior/current | 1 then 0 | `residual-red-rerun.txt`, `residual-green-rerun.txt` |
| All old primary/secondary probes | 0 | `primary-rerun.txt`, `secondary.txt` |
| Old short-deadline checks, ten rounds | 0 | `primary-ten-rerun.txt`, `secondary-ten.txt` |
| Historical V1/V3/V4 assertion-red replays | 1 each | `v1-red.txt`, `v3-red.txt`, `v4-red.txt` |
| First/second fix tests on `6b063bb`/`0007b1cc` | 1, intended assertions | `fix01-red.txt`, `fix02-red.txt` |
| Socket/stream caps changed to 2097152, then restored | 1 then 0 each | `frame-mutation-*.txt`, `stream-mutation-*.txt` |
| First two doctor checks swapped, then restored | 1 then 0 | `order-mutation-*.txt` |
| Null reference failure, prior/current, ten rounds | 1 each | `copies-*-reference-only.txt` |
| Bounded direct-Ping error, prior/current, ten rounds | 1 then 0 | `copies-*-ping-bound-final.txt` |
| Actual protocol-21 CLI, prior/current | 1 then 0 for assertion driver | `cli-protocol-red.txt`, `cli-protocol-correlated-green.txt` |
| Actual V2/V5, process logs, store cases | 0 for driver | `cli.txt` |
| Original secondary CLI matrix | 0 after fixture setup | `secondary-cli-rerun.txt` |
| Recorder with no signal | 0 | `signal-recorder.txt` |

The old probe helper `requiredCount` conflicts with the new worker helper.
Initial compilation failures count as setup failures only, not assertion-red.
Only the isolated probe helper name changes to `verify02RequiredCount`. Logic and assertions stay exact.
The old secondary CLI driver needed its own fixture parent and binary alias before its unchanged assertions could run.
The added Ping probe initially expected writer-specific body truncation from a generic buffer handler.
The final probe checks bounded errors and event correlation. Existing real-writer body assertions remain unchanged.
The root report records these harness corrections and preserves their initial captures.
No compile-only or setup failure supplies acceptance-red evidence.

### Actual V1/V3/V4 commands

```sh
go test -race -count=1 ./internal/lib/herdr -v
go test -race -count=1 ./internal/lib/run -v
go test -race -count=1 ./internal/doctor -v
```

```text
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	6.464s
exit 0
ok  	github.com/Harrison-Blair/hive/internal/lib/run	3.475s
exit 0
ok  	github.com/Harrison-Blair/hive/internal/doctor	3.022s
exit 0
```

These supplied-suite passes do not remove the independent V1 failure.

### Actual V2/V5 transcript and V6 gate

The unique absolute binary reports `6d034695d3fb`, with `modified:false`.
Its SHA256 remains in `binary.sha256`. Each fixture home has an explicit marker and absolute `--home` path.
The complete CLI transcript and matched live events remain in `cli.txt`.
Selected actual lines:

```text
ok herdr socket: /home/penguin/.config/herdr/herdr.sock protocol 22
ok herdr pane: wA9:p1
exit 0
fail herdr socket: herdr socket: /tmp/none.sock: no such file
exit 1
fail herdr socket: HERDR_SOCKET_PATH is not set
exit 1
{"outcome":"ok","phase":"complete","effects":[],"error":null,"result":{"version":"dev","commit":"6d034695d3fb","modified":false,"go":"go1.27.2-X:nodwarf5"}}
exit 0
process starts 7 ends 7 failed starts 6
```

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run ./...
go test -race -count=1 ./...
go test -run TestSTE -count=1 .
git diff --check
```

Each command runs from the actual frozen task worktree root and exits 0.
Format prints nothing. Lint reports `0 issues.`. Every uncached race package passes.
The root report includes the complete final gate output from `final-gate.txt`.
No STE baseline changes or increases.

## Preservation, cleanup, and limits

All 205 tracked paths match the exact worker before this append. Every V file remains exact.
The old verdict prefix has 26692 bytes and SHA256:
`52494e34c1699a947a4f2e3607ad494ad1170aa11f8fdccd628555282b6fd118`.
This append preserves every byte. Only this evidence file enters the separate verdict commit.
The root report records its full ID, exact parent, and non-force task-branch push.

Cleanup removed only owned copies, fixtures, binaries, alias, and outside-home TMPDIR after compact capture.
All copied Go source equals the worker after mutation restoration.
All prior reports, logs, credentials, worktrees, and other agents remain.
Historical fixed-duration sleepers expired before cleanup. Signals target only fixture-owned OS groups.
No actual signal targets PID 0, PID 1, the parent, or an unrelated process.
No real auth, model start, workspace mutation, credential inspection, or global setting change occurs.
No source, worker evidence, tracker, glossary, baseline, attribution, `dev`, or `main` changes.
No worktree removal, merge, co-author, false session trailer, or other agent control occurs.

Hostile file replacement, hostile PID reuse, and uninterruptible kernel sleep remain outside the approved boundary.
The runner's direct wait after SIGKILL cannot prove a bound for uninterruptible kernel sleep.
No imported dependency remains open.
Uncompleted: the remaining P05-F01 fix, its independent green check, task acceptance, and landing.

`END VERIFIER T05-herdr-client-and-doctor t05-verify-03`


---

# T05 verification 04 — 2026-10-18

To: coordinator

**PASS.** V1, V2, V3, V4, V5, and V6 pass separately.
P05-F01: CLOSED. All other primary and secondary findings remain CLOSED.
Recommendation: accept this exact worker for a separate landing.
This pass does not land the task or accept later task source.

## Identity and authority

| Item | Value |
| --- | --- |
| Pass | `t05-verify-04` |
| Role | Fresh independent source verifier. No implementation. |
| Actual route | `PI_PROVIDER=openai-codex`, `PI_MODEL=gpt-6.1-sol`, `PI_REASONING_LEVEL=high` |
| Independent session | `01a14e4e-3dba-7244-97fb-e810075da25c` |
| Branch | `task/T05-herdr-client-and-doctor` |
| Exact tested worker | `46b88c087f6de501cda589f017a4e3a794ab7f81` |
| Source | `51e8b0a81051f11794f997bbba6b65bcf8f2f32f` |
| Third fix initial | `f8ed52c097d7a74be6cf293da5728d3bdab64e1c` |
| Previous tested worker | `6d034695d3fbe3c92b09bb6f6fda53a2ae4fcd29` |
| Second fix initial | `75d0d2effd29a6472db67fe2fc6acd4eb14fc1ef` |
| Original base | `84da9819b31be83dbe71c36d620eadf3749e1f9f` |
| Imported prerequisite | `47da8abb152732d4667707ec98ef2a682e5421a4` |
| Initial state | Clean. Local and direct remote task tips equal the exact worker. |
| Separate verdict | Evidence-only child of the exact worker. The root report records its full ID after commit. |
| Root report | `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04.md` |
| Captures | `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/` |

I read all four complete prior root verdicts, the tracked FAIL history, V1–V6, and all worker reports.
I read the instructions, glossaries, task, tracker, relevant API reference, existing probes, and prerequisite receipts.
The current report ends with `END WORKER T05-herdr-client-and-doctor t05-fix-03`.
The provenance report ends with `END LATER FIX LAND 01`.
Both original review areas are complete. The CP1 PASS remains applicable. I did not repeat CP1.
Only the approved local-user and cooperating-writer boundary applies.

## Whole-range inspection

I inspected the whole base-to-worker diff and all three fix ranges, not only the final commit.
`full-range-path-map.txt` accounts for all 89 paths, their ownership, and their SHA256 values.
The captures retain `full-range.diff` and `repair-01.diff` through `repair-03.diff`.
The fix ranges start at `b361a0c8`, `75d0d2ef`, and `f8ed52c0`, respectively.
Each imported object equals prerequisite `47da8abb`. Every Go path equals source `51e8b0a8`.
Every store path equals accepted `ccfe0020a193a91c1493d2d4750e55f2efb81798`.
Imported home, store, T03, and docs remain prerequisites. No imported dependency remains open.
The root command equals the mechanical doctor/task union. Actual help retains all five required commands.
No internal package imports Cobra or `cmd`. I did not reopen future workspace or session type scopes.
The third fix changes only the null decode guard, its tests, and V1/V5/V6.
Runner, doctor, protocol, schema, types, modules, and the STE baseline remain unchanged by that fix.

## Separate verification items

| Item | Proves | Verdict | Fresh result |
| --- | --- | --- | --- |
| V1 | AC1 | PASS | Supplied race suite, exact null probes, nested references, Ping logs, and complete primary checklist. |
| V2 | AC1 | PASS | Actual pane `wA9:p1`, live protocol 22, matched request/reply events. |
| V3 | AC2 | PASS | Supplied race suite and complete secondary runner checklist. |
| V4 | AC3 | PASS | Supplied suite, doctor checklist, process events, and exact imported store. |
| V5 | AC3, AC4 | PASS | Actual text statuses `0,1,1,0`, JSON results, version, and home-free version. |
| V6 | AC5 | PASS | Full uncached root gate. Format empty, vet 0, lint `0 issues.`, race 0. |

## P05-F01 and P05-F05 decisions

The exact retained `verify03_null_reference_test.go` fails on `f8ed52c0` and passes on the tested worker.
The historical failure changes count 7 to 0. The fixed call keeps count 7, `Valid:true`, and the ORIGINAL pointer.
The direct `encoding/json` control, non-null error, and successful `9` controls pass.
The exact prior scalar decoder also preserves `{Value:7 Valid:true}` and assigns `{Value:9 Valid:true}` on success.
Its null assertion fails on `75d0d2ef` and passes on this worker.
Plain struct null remains a no-op. Map/slice null becomes nil. Supported custom null success remains correct.
All standard non-null struct/map/slice/nested-pointer failures and assignment controls pass.

The worker's new reference test passes. An isolated copy of that test reports every reference assertion instead of stopping at the first.
On `f8ed52c0`, it detects changed pointee, nested child, slice element, map entry, and added map key.
The same assertions pass on the exact worker. This remains the changed null path, not a new probe class.
The code sends no caller state into the custom decoder. It assigns only after success.
Recommendation: retain this small fresh-temporary design. Do not add a general cloning framework.
No prior-state inheritance promise or malicious global-effect guarantee applies.

Direct Ping protocol-21, wrong-pong, and bounded-error assertions fail on `75d0d2ef` and pass on this worker.
The actual request ID, method, duration, and bounded error agree. UTF-8 and `truncated:true` remain correct.
The actual doctor protocol-21 fixture still exits 1 and records the correlated error event.
Its assertion driver fails on the prior binary and passes on the tested worker binary.
Original request/reply bodies and every established `Call` error path remain correct.

## Every finding

| ID | Disposition | Fresh evidence |
| --- | --- | --- |
| P05-F01 | CLOSED | Exact scalar/reference probes and owned nested references pass after assertion-red replays. |
| P05-F02 | CLOSED | Typed nil and raw null encode `{}` before sizing, gate, or writes. |
| P05-F03 | CLOSED | Own-context gate wait passes ten rounds. Per-path pass/failure stays cached. |
| P05-F04 | CLOSED | Null counts as presence. Both/neither reject before caller assignment. |
| P05-F05 | CLOSED | Direct Ping and original `Call` events retain correlation and bounds. |
| P05-F06 | CLOSED | Same defect as S05-F03. Actual doctor logs seven starts and seven ends. |
| S05-F01 | CLOSED | Leader-exit deadline/cancel reaps owned children near 100 ms, also in ten rounds. |
| S05-F02 | CLOSED | Name and directory diagnostics retain one physical line. |
| S05-F03 | CLOSED | Seven process ends include six failed starts with actual errors. |
| S05-F04 | CLOSED | Returned cancellation/timeout errors equal logged errors. Ten rounds pass. |
| S05-D01 | CLOSED, imported | Exact accepted store. Eight CLI special-node fixtures reject their complete paths. |

## Complete established checklist

| Check | Fresh result |
| --- | --- |
| Complete request/reply frames, including LF | PASS at literal 1048575/1048576/1048577. Oversized requests accept zero connections, including no gate. |
| UTF-8, large rejection, wrong IDs, EOF/HTTP/arrays, slow streams | PASS. Total deadlines and caller preservation remain correct. |
| Protocol 22, one request per connection, per-path gate | PASS. Failed gate blocks later methods. Own-context waits pass ten rounds. |
| Shared eight-hex prefix/counter and two child prefixes | PASS with twenty actual concurrent calls and two actual child calls. |
| Default 10-second bound, earlier deadline, cancellation | PASS with actual elapsed-time probes. |
| Request/reply/error metadata, bodies, truncation | PASS. Literal 4096-byte bodies and direct-Ping errors remain bounded. |
| Null and non-null destinations | PASS for all established success and failure controls, including original pointer identity. |
| Default 30-second runner bound | PASS. Actual elapsed `30.000195295s`, error `process timeout`. |
| Stream caps and 8 MiB drain | PASS at literal boundaries. Stdin EOF, argv, cwd, exit code 23, and no interpolation also pass. |
| Owned/stopped groups and inherited pipes | PASS. Ten deadline/cancel/concurrent rounds pass. Known children reap without cleanup signals. |
| Runner errors and logs | PASS. Diagnostics stay on one line. Returned and logged errors agree. |
| Doctor order/status/required flags | PASS. Twelve checks retain literal order and metadata-only state effects. |
| Domain bytes, static bad types/links, record inputs | PASS. All prior primary/secondary assertions and exact imported store checks pass. |
| Store lock and concurrent reads | PASS. Timeout at `2.000595413s`. Ten rounds preserve exact domain bytes. |
| gh alternatives and harness discovery | PASS with fake gh and safe paths. No real authentication or model execution. |
| Actual CLI and version | PASS. Text/JSON statuses and home-free version remain correct. |
| Impossible PID metadata | Recorder-only in a new session. No syscall. Public runner accepts no supplied PID. |
| Full gate, TestSTE, diff check | PASS. No baseline change or increase. |

## Actual runs and test strength

Supplied suites and the full uncached gate ran from the frozen task worktree root.
Added probes, historical replays, and mutations ran only in isolated copies.
`run-results.json`, `replay-results.json`, and `observed-results.txt` record the actual results.
Historical V1/V3/V4 and all three fix-test commits fail behavior assertions.
The same current selections pass. Socket/stream caps at 2097152 and the doctor order swap fail literal assertions.
Exact restoration makes each selection pass. No compile-only failure counts as red evidence.
One scalar command selected no tests in the actual worktree. It counts only as a setup run.
The corrected isolated scalar replay fails its intended assertion. Its same fixed-worker assertion passes.
The old scalar helper uses the retained name-only adaptation. The exact reference probe retains its original bytes.
Generated run headers name the session entry commit. Archive labels and `setup.py` give each historical source identity.
The final CLI replay also retains exact synthetic log bytes before fixture removal.
Complete commands, outputs, and hashes remain in the captures and the appended evidence.

## Preservation, cleanup, and limits

All 205 tracked paths match the worker before the append. Every worker V file remains exact.
The append preserves all 41012 old verdict bytes, SHA256 `a95ff2f10577577f01a62cb78caada6e42772fea436978630c78a86814d329b9`.
Only `spec/hive-v1/evidence/T05-herdr-client-and-doctor/VERIFIED.md` enters the verdict commit.
Cleanup removes only manifest-listed own copies, fixtures, binaries, alias, and the short outside-home TMPDIR.
Probe sources, synthetic logs, raw outputs, diffs, scripts, manifests, and all prior reports remain.
Actual signals target only owned fixture groups. Impossible metadata uses only the recorder.
No signal targets PID 0, PID 1, the parent, or an unrelated process.
No credential inspection, global setting change, model start, workspace mutation, or other agent control occurs.
No source, worker evidence, tracker, glossary, baseline, attribution, `dev`, or `main` change comes from this pass.
Root `dev` initially equals `4837f1e728a2b958604734a8a3fd9e8edf4c42de` with independent staged T07 work.
At cleanup, root `dev` advanced externally to `eded8f9cb09208bfdc81334870e86479e746c183` and became clean.
This pass preserves that external work. The frozen task branch needs no unrelated root merge.
Hostile file replacement stays deferred for T01. Hostile PID reuse and uninterruptible kernel sleep remain outside this boundary.
No owned acceptance check or prerequisite check remains uncompleted.
Uncompleted: coordinator acceptance, landing, and later integration. These remain the coordinator's decisions.

## Actual run records

Each source identity below names the tested archive or actual worktree.
Raw outputs remain in the capture directory. Long outputs follow the evidence truncation rule.

### V1 supplied suite

Source: `46b88c08`. Capture: `v1.txt`.

```text
commit 46b88c087f6de501cda589f017a4e3a794ab7f81 (archive identity is in run name when different)
$ go test -race -count=1 ./internal/lib/herdr -v
=== RUN   TestCallSuccessEnvelope
--- PASS: TestCallSuccessEnvelope (0.00s)
=== RUN   TestCallErrorEnvelope
--- PASS: TestCallErrorEnvelope (0.00s)
=== RUN   TestCallRejectsBadReplies
=== RUN   TestCallRejectsBadReplies/another_id
=== RUN   TestCallRejectsBadReplies/both_result_and_error
=== RUN   TestCallRejectsBadReplies/neither_result_nor_error
--- PASS: TestCallRejectsBadReplies (0.00s)
    --- PASS: TestCallRejectsBadReplies/another_id (0.00s)
    --- PASS: TestCallRejectsBadReplies/both_result_and_error (0.00s)
    --- PASS: TestCallRejectsBadReplies/neither_result_nor_error (0.00s)
=== RUN   TestCallReplyTooLarge
=== RUN   TestCallReplyTooLarge/1048575
=== RUN   TestCallReplyTooLarge/1048576
=== RUN   TestCallReplyTooLarge/1048577
--- PASS: TestCallReplyTooLarge (0.04s)
    --- PASS: TestCallReplyTooLarge/1048575 (0.02s)
    --- PASS: TestCallReplyTooLarge/1048576 (0.01s)
    --- PASS: TestCallReplyTooLarge/1048577 (0.01s)
=== RUN   TestCallReplyTooLargeTenMegabytes
--- PASS: TestCallReplyTooLargeTenMegabytes (0.23s)
=== RUN   TestCallDeadlineWithSilentServer
--- PASS: TestCallDeadlineWithSilentServer (0.20s)
=== RUN   TestCallSlowReply
--- PASS: TestCallSlowReply (2.50s)
=== RUN   TestCallClosedReply
--- PASS: TestCallClosedReply (0.00s)
=== RUN   TestCallWithoutSocketPath
--- PASS: TestCallWithoutSocketPath (0.00s)
=== RUN   TestCallSocketPathTypes
=== RUN   TestCallSocketPathTypes/missing
=== RUN   TestCallSocketPathTypes/directory
=== RUN   TestCallSocketPathTypes/control_character
--- PASS: TestCallSocketPathTypes (0.00s)
    --- PASS: TestCallSocketPathTypes/missing (0.00s)
    --- PASS: TestCallSocketPathTypes/directory (0.00s)
    --- PASS: TestCallSocketPathTypes/control_character (0.00s)
=== RUN   TestCallRequestSize
=== RUN   TestCallRequestSize/1048575
=== RUN   TestCallRequestSize/1048576
=== RUN   TestCallRequestSize/1048577
--- PASS: TestCallRequestSize (0.32s)
    --- PASS: TestCallRequestSize/1048575 (0.13s)
    --- PASS: TestCallRequestSize/1048576 (0.12s)
    --- PASS: TestCallRequestSize/1048577 (0.07s)
=== RUN   TestRequestIDsShareOnePrefix
--- PASS: TestRequestIDsShareOnePrefix (0.00s)
=== RUN   TestHelperProcessRequestID
--- PASS: TestHelperProcessRequestID (0.00s)
=== RUN   TestRequestIDsDifferBetweenProcesses
--- PASS: TestRequestIDsDifferBetweenProcesses (2.01s)
=== RUN   TestCallLogsRequestAndReply
--- PASS: TestCallLogsRequestAndReply (0.00s)
=== RUN   TestResultUnchangedOnDecodeError
=== RUN   TestResultUnchangedOnDecodeError/struct
=== RUN   TestResultUnchangedOnDecodeError/map
=== RUN   TestResultUnchangedOnDecodeError/slice
=== RUN   TestResultUnchangedOnDecodeError/nested_pointer
--- PASS: TestResultUnchangedOnDecodeError (0.00s)
    --- PASS: TestResultUnchangedOnDecodeError/struct (0.00s)
    --- PASS: TestResultUnchangedOnDecodeError/map (0.00s)
    --- PASS: TestResultUnchangedOnDecodeError/slice (0.00s)
    --- PASS: TestResultUnchangedOnDecodeError/nested_pointer (0.00s)
=== RUN   TestNullResultPreservesCallerOnDecodeError
--- PASS: TestNullResultPreservesCallerOnDecodeError (0.00s)
=== RUN   TestNullResultPreservesReferencedStateOnDecodeError
--- PASS: TestNullResultPreservesReferencedStateOnDecodeError (0.00s)
=== RUN   TestNullResultKeepsExistingSemantics
=== RUN   TestNullResultKeepsExistingSemantics/ordinary_decoder
=== RUN   TestNullResultKeepsExistingSemantics/struct
=== RUN   TestNullResultKeepsExistingSemantics/map
=== RUN   TestNullResultKeepsExistingSemantics/slice
--- PASS: TestNullResultKeepsExistingSemantics (0.00s)
    --- PASS: TestNullResultKeepsExistingSemantics/ordinary_decoder (0.00s)
    --- PASS: TestNullResultKeepsExistingSemantics/struct (0.00s)
    --- PASS: TestNullResultKeepsExistingSemantics/map (0.00s)
    --- PASS: TestNullResultKeepsExistingSemantics/slice (0.00s)
=== RUN   TestReferencedResultAssignedOnSuccess
--- PASS: TestReferencedResultAssignedOnSuccess (0.00s)
=== RUN   TestParamsNeverNull
=== RUN   TestParamsNeverNull/untyped_nil
=== RUN   TestParamsNeverNull/typed_nil_map
=== RUN   TestParamsNeverNull/typed_nil_pointer
=== RUN   TestParamsNeverNull/typed_nil_slice
=== RUN   TestParamsNeverNull/raw_null
--- PASS: TestParamsNeverNull (0.00s)
    --- PASS: TestParamsNeverNull/untyped_nil (0.00s)
    --- PASS: TestParamsNeverNull/typed_nil_map (0.00s)
    --- PASS: TestParamsNeverNull/typed_nil_pointer (0.00s)
    --- PASS: TestParamsNeverNull/typed_nil_slice (0.00s)
    --- PASS: TestParamsNeverNull/raw_null (0.00s)
=== RUN   TestReplyNullFieldPresence
=== RUN   TestReplyNullFieldPresence/result_null_with_error
=== RUN   TestReplyNullFieldPresence/error_null_with_result
=== RUN   TestReplyNullFieldPresence/both_null
=== RUN   TestReplyNullFieldPresence/error_null_only
--- PASS: TestReplyNullFieldPresence (0.00s)
    --- PASS: TestReplyNullFieldPresence/result_null_with_error (0.00s)
    --- PASS: TestReplyNullFieldPresence/error_null_with_result (0.00s)
    --- PASS: TestReplyNullFieldPresence/both_null (0.00s)
    --- PASS: TestReplyNullFieldPresence/error_null_only (0.00s)
=== RUN   TestCallLogsEveryFailedPath
=== RUN   TestCallLogsEveryFailedPath/server_error
=== RUN   TestCallLogsEveryFailedPath/malformed_envelope
=== RUN   TestCallLogsEveryFailedPath/end_of_file
=== RUN   TestCallLogsEveryFailedPath/timeout
=== RUN   TestCallLogsEveryFailedPath/gate_failure
=== RUN   TestCallLogsEveryFailedPath/missing_socket_path
=== RUN   TestCallLogsEveryFailedPath/bounded_error_text
=== RUN   TestCallLogsEveryFailedPath/success_has_no_error_event
--- PASS: TestCallLogsEveryFailedPath (0.11s)
    --- PASS: TestCallLogsEveryFailedPath/server_error (0.00s)
    --- PASS: TestCallLogsEveryFailedPath/malformed_envelope (0.00s)
    --- PASS: TestCallLogsEveryFailedPath/end_of_file (0.00s)
    --- PASS: TestCallLogsEveryFailedPath/timeout (0.10s)
    --- PASS: TestCallLogsEveryFailedPath/gate_failure (0.00s)
    --- PASS: TestCallLogsEveryFailedPath/missing_socket_path (0.00s)
    --- PASS: TestCallLogsEveryFailedPath/bounded_error_text (0.00s)
    --- PASS: TestCallLogsEveryFailedPath/success_has_no_error_event (0.00s)
=== RUN   TestResultAssignedOnSuccessfulDecode
=== RUN   TestResultAssignedOnSuccessfulDecode/map
=== RUN   TestResultAssignedOnSuccessfulDecode/slice
=== RUN   TestResultAssignedOnSuccessfulDecode/nested_pointer
--- PASS: TestResultAssignedOnSuccessfulDecode (0.00s)
    --- PASS: TestResultAssignedOnSuccessfulDecode/map (0.00s)
    --- PASS: TestResultAssignedOnSuccessfulDecode/slice (0.00s)
    --- PASS: TestResultAssignedOnSuccessfulDecode/nested_pointer (0.00s)
=== RUN   TestPingChecksProtocol
--- PASS: TestPingChecksProtocol (0.00s)
=== RUN   TestPingRejectsOtherProtocol
--- PASS: TestPingRejectsOtherProtocol (0.00s)
=== RUN   TestPingLogsResultErrors
=== RUN   TestPingLogsResultErrors/protocol_mismatch
=== RUN   TestPingLogsResultErrors/wrong_type
--- PASS: TestPingLogsResultErrors (0.00s)
    --- PASS: TestPingLogsResultErrors/protocol_mismatch (0.00s)
    --- PASS: TestPingLogsResultErrors/wrong_type (0.00s)
=== RUN   TestProtocolGateRunsPingOnce
--- PASS: TestProtocolGateRunsPingOnce (0.00s)
=== RUN   TestProtocolGateFailureBlocksLaterCalls
--- PASS: TestProtocolGateFailureBlocksLaterCalls (0.00s)
=== RUN   TestProtocolGateIsPerSocketPath
--- PASS: TestProtocolGateIsPerSocketPath (0.00s)
=== RUN   TestCallWithMethodPingDoesNotGate
--- PASS: TestCallWithMethodPingDoesNotGate (0.00s)
=== RUN   TestProtocolGateWaiterHonorsOwnDeadline
--- PASS: TestProtocolGateWaiterHonorsOwnDeadline (0.06s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	6.485s

exit 0
```

### V1 exact reference assertion-red

Source: `f8ed52c0`. Capture: `reference-f8ed52c0-red.txt`.

```text
commit 46b88c087f6de501cda589f017a4e3a794ab7f81 (archive identity is in run name when different)
$ go test -race -count=1 ./internal/lib/herdr -run TestVerify03NullReference|TestVerify03NonNullReference|TestVerify03ReferenceSuccess -v
=== RUN   TestVerify03NullReferenceContract
--- PASS: TestVerify03NullReferenceContract (0.00s)
=== RUN   TestVerify03NullReferenceFailurePreservesCaller
    verify03_null_reference_test.go:34: null err=bad reply: count is required count=0 valid=true same-pointer=true
    verify03_null_reference_test.go:36: caller changed on failure: count=0 valid=true, want count=7 valid=true and original pointer
--- FAIL: TestVerify03NullReferenceFailurePreservesCaller (0.00s)
=== RUN   TestVerify03NonNullReferenceFailurePreservesCaller
--- PASS: TestVerify03NonNullReferenceFailurePreservesCaller (0.00s)
=== RUN   TestVerify03ReferenceSuccess
--- PASS: TestVerify03ReferenceSuccess (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/herdr	0.006s
FAIL

exit 1
```

### V1 exact reference and null controls green

Source: `46b88c08`. Capture: `reference-worker-green.txt`.

```text
commit 46b88c087f6de501cda589f017a4e3a794ab7f81 (archive identity is in run name when different)
$ go test -race -count=1 ./internal/lib/herdr -run TestVerify03|TestNullResult|TestReferencedResult|TestResultUnchanged|TestResultAssigned|TestVerify02Null -v
=== RUN   TestResultUnchangedOnDecodeError
=== RUN   TestResultUnchangedOnDecodeError/struct
=== RUN   TestResultUnchangedOnDecodeError/map
=== RUN   TestResultUnchangedOnDecodeError/slice
=== RUN   TestResultUnchangedOnDecodeError/nested_pointer
--- PASS: TestResultUnchangedOnDecodeError (0.00s)
    --- PASS: TestResultUnchangedOnDecodeError/struct (0.00s)
    --- PASS: TestResultUnchangedOnDecodeError/map (0.00s)
    --- PASS: TestResultUnchangedOnDecodeError/slice (0.00s)
    --- PASS: TestResultUnchangedOnDecodeError/nested_pointer (0.00s)
=== RUN   TestNullResultPreservesCallerOnDecodeError
--- PASS: TestNullResultPreservesCallerOnDecodeError (0.00s)
=== RUN   TestNullResultPreservesReferencedStateOnDecodeError
--- PASS: TestNullResultPreservesReferencedStateOnDecodeError (0.00s)
=== RUN   TestNullResultKeepsExistingSemantics
=== RUN   TestNullResultKeepsExistingSemantics/ordinary_decoder
=== RUN   TestNullResultKeepsExistingSemantics/struct
=== RUN   TestNullResultKeepsExistingSemantics/map
=== RUN   TestNullResultKeepsExistingSemantics/slice
--- PASS: TestNullResultKeepsExistingSemantics (0.00s)
    --- PASS: TestNullResultKeepsExistingSemantics/ordinary_decoder (0.00s)
    --- PASS: TestNullResultKeepsExistingSemantics/struct (0.00s)
    --- PASS: TestNullResultKeepsExistingSemantics/map (0.00s)
    --- PASS: TestNullResultKeepsExistingSemantics/slice (0.00s)
=== RUN   TestReferencedResultAssignedOnSuccess
--- PASS: TestReferencedResultAssignedOnSuccess (0.00s)
=== RUN   TestResultAssignedOnSuccessfulDecode
=== RUN   TestResultAssignedOnSuccessfulDecode/map
=== RUN   TestResultAssignedOnSuccessfulDecode/slice
=== RUN   TestResultAssignedOnSuccessfulDecode/nested_pointer
--- PASS: TestResultAssignedOnSuccessfulDecode (0.00s)
    --- PASS: TestResultAssignedOnSuccessfulDecode/map (0.00s)
    --- PASS: TestResultAssignedOnSuccessfulDecode/slice (0.00s)
    --- PASS: TestResultAssignedOnSuccessfulDecode/nested_pointer (0.00s)
=== RUN   TestVerify02NullOrdinaryDecoderContract
    verify02_null_test.go:28: encoding/json control: err=count is required, value={Value:0 Valid:false}
--- PASS: TestVerify02NullOrdinaryDecoderContract (0.00s)
=== RUN   TestVerify02NullFailurePreservesCaller
    verify02_null_test.go:34: Call result:null: err=bad reply: count is required, value={Value:7 Valid:true}
--- PASS: TestVerify02NullFailurePreservesCaller (0.00s)
=== RUN   TestVerify03NullReferenceContract
--- PASS: TestVerify03NullReferenceContract (0.00s)
=== RUN   TestVerify03NullReferenceFailurePreservesCaller
    verify03_null_reference_test.go:34: null err=bad reply: count is required count=7 valid=true same-pointer=true
--- PASS: TestVerify03NullReferenceFailurePreservesCaller (0.00s)
=== RUN   TestVerify03NonNullReferenceFailurePreservesCaller
--- PASS: TestVerify03NonNullReferenceFailurePreservesCaller (0.00s)
=== RUN   TestVerify03ReferenceSuccess
--- PASS: TestVerify03ReferenceSuccess (0.00s)
=== RUN   TestVerify03PingBoundedError
    verify03_ping_bound_test.go:29: actual id=713aaa05-33 method=ping duration=1 error-bytes=4096 valid-UTF8=true truncated=true
--- PASS: TestVerify03PingBoundedError (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	1.019s

exit 0
```

### V1 exact scalar assertion-red

Source: `75d0d2ef`. Capture: `scalar-75d0d2ef-red.txt`.

```text
=== RUN   TestVerify02NullOrdinaryDecoderContract
    verify02_null_test.go:28: encoding/json control: err=count is required, value={Value:0 Valid:false}
--- PASS: TestVerify02NullOrdinaryDecoderContract (0.00s)
=== RUN   TestVerify02NullFailurePreservesCaller
    verify02_null_test.go:34: Call result:null: err=bad reply: count is required, value={Value:0 Valid:false}
    verify02_null_test.go:36: caller changed on failure: {Value:0 Valid:false}, want {Value:7 Valid:true}
--- FAIL: TestVerify02NullFailurePreservesCaller (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/herdr	0.004s
FAIL
exit 1
```

### V1 all primary and scalar success controls

Source: `46b88c08`. Capture: `primary.txt`.

```text
commit 46b88c087f6de501cda589f017a4e3a794ab7f81 (archive identity is in run name when different)
$ go test -race -count=1 ./internal/lib/herdr -run TestPrimary|TestVerify02|TestVerify03 -v
=== RUN   TestVerify02NullOrdinaryDecoderContract
    verify02_null_test.go:28: encoding/json control: err=count is required, value={Value:0 Valid:false}
--- PASS: TestVerify02NullOrdinaryDecoderContract (0.00s)
=== RUN   TestVerify02NullFailurePreservesCaller
    verify02_null_test.go:34: Call result:null: err=bad reply: count is required, value={Value:7 Valid:true}
--- PASS: TestVerify02NullFailurePreservesCaller (0.00s)
=== RUN   TestVerify02NonNullFailurePreservesCaller
--- PASS: TestVerify02NonNullFailurePreservesCaller (0.00s)
=== RUN   TestVerify02OrdinaryDecoderSuccess
--- PASS: TestVerify02OrdinaryDecoderSuccess (0.00s)
=== RUN   TestVerify02DirectPingProtocolErrorLogged
=== RUN   TestVerify02DirectPingProtocolErrorLogged/{"type":"pong","protocol":21}
    verify02_pinglog_test.go:30: returned=herdr protocol 21, hive needs 22
        {"time":"2026-10-18T05:21:00.621947409-04:00","level":"DEBUG","msg":"herdr request","request_id":"b88fe476-4","method":"ping","duration_ms":0,"body":"{\"id\":\"b88fe476-4\",\"method\":\"ping\",\"params\":{}}"}
        {"time":"2026-10-18T05:21:00.62206071-04:00","level":"DEBUG","msg":"herdr reply","request_id":"b88fe476-4","method":"ping","duration_ms":0,"body":"{\"id\":\"b88fe476-4\",\"result\":{\"type\":\"pong\",\"protocol\":21}}"}
        {"time":"2026-10-18T05:21:00.62213078-04:00","level":"DEBUG","msg":"herdr error","request_id":"b88fe476-4","method":"ping","duration_ms":0,"error":"herdr protocol 21, hive needs 22"}
=== RUN   TestVerify02DirectPingProtocolErrorLogged/{"type":"not-pong","protocol":22}
    verify02_pinglog_test.go:30: returned=bad reply: result type "not-pong", want pong
        {"time":"2026-10-18T05:21:00.622534511-04:00","level":"DEBUG","msg":"herdr request","request_id":"b88fe476-5","method":"ping","duration_ms":0,"body":"{\"id\":\"b88fe476-5\",\"method\":\"ping\",\"params\":{}}"}
        {"time":"2026-10-18T05:21:00.622603309-04:00","level":"DEBUG","msg":"herdr reply","request_id":"b88fe476-5","method":"ping","duration_ms":0,"body":"{\"id\":\"b88fe476-5\",\"result\":{\"type\":\"not-pong\",\"protocol\":22}}"}
        {"time":"2026-10-18T05:21:00.622649074-04:00","level":"DEBUG","msg":"herdr error","request_id":"b88fe476-5","method":"ping","duration_ms":0,"error":"bad reply: result type \"not-pong\", want pong"}
--- PASS: TestVerify02DirectPingProtocolErrorLogged (0.00s)
    --- PASS: TestVerify02DirectPingProtocolErrorLogged/{"type":"pong","protocol":21} (0.00s)
    --- PASS: TestVerify02DirectPingProtocolErrorLogged/{"type":"not-pong","protocol":22} (0.00s)
=== RUN   TestPrimaryResultPreservedOnDecodeError
--- PASS: TestPrimaryResultPreservedOnDecodeError (0.00s)
=== RUN   TestPrimaryParamsNeverNull
=== RUN   TestPrimaryParamsNeverNull/<nil>
=== RUN   TestPrimaryParamsNeverNull/map[string]interface_{}
=== RUN   TestPrimaryParamsNeverNull/*struct_{}
=== RUN   TestPrimaryParamsNeverNull/jsontext.Value
--- PASS: TestPrimaryParamsNeverNull (0.00s)
    --- PASS: TestPrimaryParamsNeverNull/<nil> (0.00s)
    --- PASS: TestPrimaryParamsNeverNull/map[string]interface_{} (0.00s)
    --- PASS: TestPrimaryParamsNeverNull/*struct_{} (0.00s)
    --- PASS: TestPrimaryParamsNeverNull/jsontext.Value (0.00s)
=== RUN   TestPrimaryNullFieldExclusivity
=== RUN   TestPrimaryNullFieldExclusivity/"result":null,"error":{"code":"x","message":"y"}
=== RUN   TestPrimaryNullFieldExclusivity/"result":{"type":"ok"},"error":null
--- PASS: TestPrimaryNullFieldExclusivity (0.00s)
    --- PASS: TestPrimaryNullFieldExclusivity/"result":null,"error":{"code":"x","message":"y"} (0.00s)
    --- PASS: TestPrimaryNullFieldExclusivity/"result":{"type":"ok"},"error":null (0.00s)
=== RUN   TestPrimaryGateWaitDeadline
--- PASS: TestPrimaryGateWaitDeadline (0.60s)
=== RUN   TestPrimaryErrorEvents
=== RUN   TestPrimaryErrorEvents/server_error
=== RUN   TestPrimaryErrorEvents/malformed
=== RUN   TestPrimaryErrorEvents/eof
=== RUN   TestPrimaryErrorEvents/timeout
--- PASS: TestPrimaryErrorEvents (0.10s)
    --- PASS: TestPrimaryErrorEvents/server_error (0.00s)
    --- PASS: TestPrimaryErrorEvents/malformed (0.00s)
    --- PASS: TestPrimaryErrorEvents/eof (0.00s)
    --- PASS: TestPrimaryErrorEvents/timeout (0.10s)
=== RUN   TestPrimaryFrameBounds
=== RUN   TestPrimaryFrameBounds/request1048575
=== RUN   TestPrimaryFrameBounds/reply1048575
=== RUN   TestPrimaryFrameBounds/request1048576
=== RUN   TestPrimaryFrameBounds/reply1048576
=== RUN   TestPrimaryFrameBounds/request1048577
=== RUN   TestPrimaryFrameBounds/reply1048577
--- PASS: TestPrimaryFrameBounds (0.26s)
    --- PASS: TestPrimaryFrameBounds/request1048575 (0.09s)
    --- PASS: TestPrimaryFrameBounds/reply1048575 (0.02s)
    --- PASS: TestPrimaryFrameBounds/request1048576 (0.08s)
    --- PASS: TestPrimaryFrameBounds/reply1048576 (0.02s)
    --- PASS: TestPrimaryFrameBounds/request1048577 (0.05s)
    --- PASS: TestPrimaryFrameBounds/reply1048577 (0.01s)
=== RUN   TestPrimaryTransport
=== RUN   TestPrimaryTransport/empty
=== RUN   TestPrimaryTransport/partial
=== RUN   TestPrimaryTransport/http
=== RUN   TestPrimaryTransport/array
=== RUN   TestPrimaryTransport/slow
--- PASS: TestPrimaryTransport (0.12s)
    --- PASS: TestPrimaryTransport/empty (0.00s)
    --- PASS: TestPrimaryTransport/partial (0.00s)
    --- PASS: TestPrimaryTransport/http (0.00s)
    --- PASS: TestPrimaryTransport/array (0.00s)
    --- PASS: TestPrimaryTransport/slow (0.12s)
=== RUN   TestPrimaryCallerDeadlineAndCancel
=== RUN   TestPrimaryCallerDeadlineAndCancel/false
=== RUN   TestPrimaryCallerDeadlineAndCancel/true
--- PASS: TestPrimaryCallerDeadlineAndCancel (0.40s)
    --- PASS: TestPrimaryCallerDeadlineAndCancel/false (0.20s)
    --- PASS: TestPrimaryCallerDeadlineAndCancel/true (0.20s)
=== RUN   TestPrimaryIDsAcrossCallers
--- PASS: TestPrimaryIDsAcrossCallers (0.00s)
=== RUN   TestPrimaryChildCall
--- PASS: TestPrimaryChildCall (0.00s)
=== RUN   TestPrimaryIDsAcrossProcesses
--- PASS: TestPrimaryIDsAcrossProcesses (2.01s)
=== RUN   TestPrimaryDefaultDeadline
--- PASS: TestPrimaryDefaultDeadline (10.50s)
=== RUN   TestPrimaryLogBodyBound
--- PASS: TestPrimaryLogBodyBound (0.00s)
=== RUN   TestVerify03NullReferenceContract
--- PASS: TestVerify03NullReferenceContract (0.00s)
=== RUN   TestVerify03NullReferenceFailurePreservesCaller
    verify03_null_reference_test.go:34: null err=bad reply: count is required count=7 valid=true same-pointer=true
--- PASS: TestVerify03NullReferenceFailurePreservesCaller (0.00s)
=== RUN   TestVerify03NonNullReferenceFailurePreservesCaller
--- PASS: TestVerify03NonNullReferenceFailurePreservesCaller (0.00s)
=== RUN   TestVerify03ReferenceSuccess
--- PASS: TestVerify03ReferenceSuccess (0.00s)
=== RUN   TestVerify03PingBoundedError
    verify03_ping_bound_test.go:29: actual id=b88fe476-35 method=ping duration=1 error-bytes=4096 valid-UTF8=true truncated=true
--- PASS: TestVerify03PingBoundedError (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	15.017s

exit 0
```

### V1 nested reference assertions red

Source: `f8ed52c0`. Capture: `nested-prior.txt`.

```text
commit f8ed52c0
$ go test -race -count=1 ./internal/lib/herdr -run TestVerify04 -v
=== RUN   TestVerify04OwnedNullReferences
    verify04_owned_null_references_test.go:60: value = 0xc0000183c0 count = 0, want original pointer and 7
    verify04_owned_null_references_test.go:63: child = &{Count:0}, want original pointer and count 3
    verify04_owned_null_references_test.go:66: list = [<nil>], want the original slice and element
    verify04_owned_null_references_test.go:69: lookup = map[new:0xc0000183c0], want the original map and entry
    verify04_owned_null_references_test.go:72: lookup = map[new:0xc0000183c0], want no new entry
--- FAIL: TestVerify04OwnedNullReferences (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/herdr	0.004s
FAIL
exit 1
```

### V1 same nested reference assertions green

Source: `46b88c08`. Capture: `nested-worker.txt`.

```text
commit HEAD
$ go test -race -count=1 ./internal/lib/herdr -run TestVerify04 -v
=== RUN   TestVerify04OwnedNullReferences
--- PASS: TestVerify04OwnedNullReferences (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	1.004s
exit 0
```

### V1 direct Ping assertions red

Source: `75d0d2ef`. Capture: `ping-75d0d2ef-red.txt`.

```text
commit 46b88c087f6de501cda589f017a4e3a794ab7f81 (archive identity is in run name when different)
$ go test -race -count=1 ./internal/lib/herdr -run TestVerify02DirectPing|TestVerify03Ping -v
=== RUN   TestVerify02DirectPingProtocolErrorLogged
=== RUN   TestVerify02DirectPingProtocolErrorLogged/{"type":"pong","protocol":21}
    verify02_pinglog_test.go:30: returned=herdr protocol 21, hive needs 22
        {"time":"2026-10-18T05:20:59.345468046-04:00","level":"DEBUG","msg":"herdr request","request_id":"99b4dbf6-1","method":"ping","duration_ms":0,"body":"{\"id\":\"99b4dbf6-1\",\"method\":\"ping\",\"params\":{}}"}
        {"time":"2026-10-18T05:20:59.345640928-04:00","level":"DEBUG","msg":"herdr reply","request_id":"99b4dbf6-1","method":"ping","duration_ms":0,"body":"{\"id\":\"99b4dbf6-1\",\"result\":{\"type\":\"pong\",\"protocol\":21}}"}
    verify02_pinglog_test.go:31: no correlated actual Ping error: herdr protocol 21, hive needs 22
=== RUN   TestVerify02DirectPingProtocolErrorLogged/{"type":"not-pong","protocol":22}
    verify02_pinglog_test.go:30: returned=bad reply: result type "not-pong", want pong
        {"time":"2026-10-18T05:20:59.346124237-04:00","level":"DEBUG","msg":"herdr request","request_id":"99b4dbf6-2","method":"ping","duration_ms":0,"body":"{\"id\":\"99b4dbf6-2\",\"method\":\"ping\",\"params\":{}}"}
        {"time":"2026-10-18T05:20:59.346187655-04:00","level":"DEBUG","msg":"herdr reply","request_id":"99b4dbf6-2","method":"ping","duration_ms":0,"body":"{\"id\":\"99b4dbf6-2\",\"result\":{\"type\":\"not-pong\",\"protocol\":22}}"}
    verify02_pinglog_test.go:31: no correlated actual Ping error: bad reply: result type "not-pong", want pong
--- FAIL: TestVerify02DirectPingProtocolErrorLogged (0.00s)
    --- FAIL: TestVerify02DirectPingProtocolErrorLogged/{"type":"pong","protocol":21} (0.00s)
    --- FAIL: TestVerify02DirectPingProtocolErrorLogged/{"type":"not-pong","protocol":22} (0.00s)
=== RUN   TestVerify03PingBoundedError
    verify03_ping_bound_test.go:32: events request=1 reply=1 error=0, want 1 each
--- FAIL: TestVerify03PingBoundedError (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/herdr	0.007s
FAIL

exit 1
```

### V1 same direct Ping assertions green

Source: `46b88c08`. Capture: `ping-worker-green.txt`.

```text
commit 46b88c087f6de501cda589f017a4e3a794ab7f81 (archive identity is in run name when different)
$ go test -race -count=1 ./internal/lib/herdr -run TestVerify02DirectPing|TestVerify03Ping -v
=== RUN   TestVerify02DirectPingProtocolErrorLogged
=== RUN   TestVerify02DirectPingProtocolErrorLogged/{"type":"pong","protocol":21}
    verify02_pinglog_test.go:30: returned=herdr protocol 21, hive needs 22
        {"time":"2026-10-18T05:20:59.469090779-04:00","level":"DEBUG","msg":"herdr request","request_id":"e9dbeeaf-1","method":"ping","duration_ms":0,"body":"{\"id\":\"e9dbeeaf-1\",\"method\":\"ping\",\"params\":{}}"}
        {"time":"2026-10-18T05:20:59.469281183-04:00","level":"DEBUG","msg":"herdr reply","request_id":"e9dbeeaf-1","method":"ping","duration_ms":0,"body":"{\"id\":\"e9dbeeaf-1\",\"result\":{\"type\":\"pong\",\"protocol\":21}}"}
        {"time":"2026-10-18T05:20:59.469368465-04:00","level":"DEBUG","msg":"herdr error","request_id":"e9dbeeaf-1","method":"ping","duration_ms":0,"error":"herdr protocol 21, hive needs 22"}
=== RUN   TestVerify02DirectPingProtocolErrorLogged/{"type":"not-pong","protocol":22}
    verify02_pinglog_test.go:30: returned=bad reply: result type "not-pong", want pong
        {"time":"2026-10-18T05:20:59.469767728-04:00","level":"DEBUG","msg":"herdr request","request_id":"e9dbeeaf-2","method":"ping","duration_ms":0,"body":"{\"id\":\"e9dbeeaf-2\",\"method\":\"ping\",\"params\":{}}"}
        {"time":"2026-10-18T05:20:59.469880397-04:00","level":"DEBUG","msg":"herdr reply","request_id":"e9dbeeaf-2","method":"ping","duration_ms":0,"body":"{\"id\":\"e9dbeeaf-2\",\"result\":{\"type\":\"not-pong\",\"protocol\":22}}"}
        {"time":"2026-10-18T05:20:59.4699408-04:00","level":"DEBUG","msg":"herdr error","request_id":"e9dbeeaf-2","method":"ping","duration_ms":0,"error":"bad reply: result type \"not-pong\", want pong"}
--- PASS: TestVerify02DirectPingProtocolErrorLogged (0.00s)
    --- PASS: TestVerify02DirectPingProtocolErrorLogged/{"type":"pong","protocol":21} (0.00s)
    --- PASS: TestVerify02DirectPingProtocolErrorLogged/{"type":"not-pong","protocol":22} (0.00s)
=== RUN   TestVerify03PingBoundedError
    verify03_ping_bound_test.go:29: actual id=e9dbeeaf-3 method=ping duration=1 error-bytes=4096 valid-UTF8=true truncated=true
--- PASS: TestVerify03PingBoundedError (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	1.009s

exit 0
```

### V1 timing and residual checks, ten rounds

Source: `46b88c08`. Capture: `primary-ten.txt`.

```text
commit 46b88c087f6de501cda589f017a4e3a794ab7f81 (archive identity is in run name when different)
$ go test -race -count=10 ./internal/lib/herdr -run TestPrimaryGateWaitDeadline|TestPrimaryCallerDeadlineAndCancel|TestPrimaryTransport|TestVerify02|TestVerify03 -v
=== RUN   TestVerify02NullOrdinaryDecoderContract
    verify02_null_test.go:28: encoding/json control: err=count is required, value={Value:0 Valid:false}
--- PASS: TestVerify02NullOrdinaryDecoderContract (0.00s)
=== RUN   TestVerify02NullFailurePreservesCaller
    verify02_null_test.go:34: Call result:null: err=bad reply: count is required, value={Value:7 Valid:true}
--- PASS: TestVerify02NullFailurePreservesCaller (0.00s)
=== RUN   TestVerify02NonNullFailurePreservesCaller
--- PASS: TestVerify02NonNullFailurePreservesCaller (0.00s)
=== RUN   TestVerify02OrdinaryDecoderSuccess
--- PASS: TestVerify02OrdinaryDecoderSuccess (0.00s)
=== RUN   TestVerify02DirectPingProtocolErrorLogged
=== RUN   TestVerify02DirectPingProtocolErrorLogged/{"type":"pong","protocol":21}
    verify02_pinglog_test.go:30: returned=herdr protocol 21, hive needs 22
        {"time":"2026-10-18T05:21:47.665288217-04:00","level":"DEBUG","msg":"herdr request","request_id":"47d0a4c8-4","method":"ping","duration_ms":0,"body":"{\"id\":\"47d0a4c8-4\",\"method\":\"ping\",\"params\":{}}"}
        {"time":"2026-10-18T05:21:47.665421846-04:00","level":"DEBUG","msg":"herdr reply","request_id":"47d0a4c8-4","method":"ping","duration_ms":0,"body":"{\"id\":\"47d0a4c8-4\",\"result\":{\"type\":\"pong\",\"protocol\":21}}"}
        {"time":"2026-10-18T05:21:47.665490474-04:00","level":"DEBUG","msg":"herdr error","request_id":"47d0a4c8-4","method":"ping","duration_ms":0,"error":"herdr protocol 21, hive needs 22"}
=== RUN   TestVerify02DirectPingProtocolErrorLogged/{"type":"not-pong","protocol":22}
    verify02_pinglog_test.go:30: returned=bad reply: result type "not-pong", want pong
        {"time":"2026-10-18T05:21:47.665912699-04:00","level":"DEBUG","msg":"herdr request","request_id":"47d0a4c8-5","method":"ping","duration_ms":0,"body":"{\"id\":\"47d0a4c8-5\",\"method\":\"ping\",\"params\":{}}"}
        {"time":"2026-10-18T05:21:47.665982599-04:00","level":"DEBUG","msg":"herdr reply","request_id":"47d0a4c8-5","method":"ping","duration_ms":0,"body":"{\"id\":\"47d0a4c8-5\",\"result\":{\"type\":\"not-pong\",\"protocol\":22}}"}
        {"time":"2026-10-18T05:21:47.666066825-04:00","level":"DEBUG","msg":"herdr error","request_id":"47d0a4c8-5","method":"ping","duration_ms":0,"error":"bad reply: result type \"not-pong\", want pong"}
--- PASS: TestVerify02DirectPingProtocolErrorLogged (0.00s)
    --- PASS: TestVerify02DirectPingProtocolErrorLogged/{"type":"pong","protocol":21} (0.00s)
    --- PASS: TestVerify02DirectPingProtocolErrorLogged/{"type":"not-pong","protocol":22} (0.00s)
=== RUN   TestPrimaryGateWaitDeadline
--- PASS: TestPrimaryGateWaitDeadline (0.60s)
=== RUN   TestPrimaryTransport
=== RUN   TestPrimaryTransport/empty
=== RUN   TestPrimaryTransport/partial
=== RUN   TestPrimaryTransport/http
=== RUN   TestPrimaryTransport/array
=== RUN   TestPrimaryTransport/slow
--- PASS: TestPrimaryTransport (0.12s)
    --- PASS: TestPrimaryTransport/empty (0.00s)
    --- PASS: TestPrimaryTransport/partial (0.00s)
    --- PASS: TestPrimaryTransport/http (0.00s)
    --- PASS: TestPrimaryTransport/array (0.00s)
    --- PASS: TestPrimaryTransport/slow (0.12s)
=== RUN   TestPrimaryCallerDeadlineAndCancel
=== RUN   TestPrimaryCallerDeadlineAndCancel/false
=== RUN   TestPrimaryCallerDeadlineAndCancel/true
--- PASS: TestPrimaryCallerDeadlineAndCancel (0.40s)
    --- PASS: TestPrimaryCallerDeadlineAndCancel/false (0.20s)
    --- PASS: TestPrimaryCallerDeadlineAndCancel/true (0.20s)
=== RUN   TestVerify03NullReferenceContract
--- PASS: TestVerify03NullReferenceContract (0.00s)
=== RUN   TestVerify03NullReferenceFailurePreservesCaller
    verify03_null_reference_test.go:34: null err=bad reply: count is required count=7 valid=true same-pointer=true
[416 lines removed. Complete raw output: /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/primary-ten.txt]
=== RUN   TestVerify02DirectPingProtocolErrorLogged/{"type":"not-pong","protocol":22}
    verify02_pinglog_test.go:30: returned=bad reply: result type "not-pong", want pong
        {"time":"2026-10-18T05:21:56.756082877-04:00","level":"DEBUG","msg":"herdr request","request_id":"47d0a4c8-157","method":"ping","duration_ms":0,"body":"{\"id\":\"47d0a4c8-157\",\"method\":\"ping\",\"params\":{}}"}
        {"time":"2026-10-18T05:21:56.756134644-04:00","level":"DEBUG","msg":"herdr reply","request_id":"47d0a4c8-157","method":"ping","duration_ms":0,"body":"{\"id\":\"47d0a4c8-157\",\"result\":{\"type\":\"not-pong\",\"protocol\":22}}"}
        {"time":"2026-10-18T05:21:56.756176031-04:00","level":"DEBUG","msg":"herdr error","request_id":"47d0a4c8-157","method":"ping","duration_ms":0,"error":"bad reply: result type \"not-pong\", want pong"}
--- PASS: TestVerify02DirectPingProtocolErrorLogged (0.00s)
    --- PASS: TestVerify02DirectPingProtocolErrorLogged/{"type":"pong","protocol":21} (0.00s)
    --- PASS: TestVerify02DirectPingProtocolErrorLogged/{"type":"not-pong","protocol":22} (0.00s)
=== RUN   TestPrimaryGateWaitDeadline
--- PASS: TestPrimaryGateWaitDeadline (0.60s)
=== RUN   TestPrimaryTransport
=== RUN   TestPrimaryTransport/empty
=== RUN   TestPrimaryTransport/partial
=== RUN   TestPrimaryTransport/http
=== RUN   TestPrimaryTransport/array
=== RUN   TestPrimaryTransport/slow
--- PASS: TestPrimaryTransport (0.12s)
    --- PASS: TestPrimaryTransport/empty (0.00s)
    --- PASS: TestPrimaryTransport/partial (0.00s)
    --- PASS: TestPrimaryTransport/http (0.00s)
    --- PASS: TestPrimaryTransport/array (0.00s)
    --- PASS: TestPrimaryTransport/slow (0.12s)
=== RUN   TestPrimaryCallerDeadlineAndCancel
=== RUN   TestPrimaryCallerDeadlineAndCancel/false
=== RUN   TestPrimaryCallerDeadlineAndCancel/true
--- PASS: TestPrimaryCallerDeadlineAndCancel (0.40s)
    --- PASS: TestPrimaryCallerDeadlineAndCancel/false (0.20s)
    --- PASS: TestPrimaryCallerDeadlineAndCancel/true (0.20s)
=== RUN   TestVerify03NullReferenceContract
--- PASS: TestVerify03NullReferenceContract (0.00s)
=== RUN   TestVerify03NullReferenceFailurePreservesCaller
    verify03_null_reference_test.go:34: null err=bad reply: count is required count=7 valid=true same-pointer=true
--- PASS: TestVerify03NullReferenceFailurePreservesCaller (0.00s)
=== RUN   TestVerify03NonNullReferenceFailurePreservesCaller
--- PASS: TestVerify03NonNullReferenceFailurePreservesCaller (0.00s)
=== RUN   TestVerify03ReferenceSuccess
--- PASS: TestVerify03ReferenceSuccess (0.00s)
=== RUN   TestVerify03PingBoundedError
    verify03_ping_bound_test.go:29: actual id=47d0a4c8-171 method=ping duration=1 error-bytes=4096 valid-UTF8=true truncated=true
--- PASS: TestVerify03PingBoundedError (0.00s)
=== RUN   TestVerify02NullOrdinaryDecoderContract
    verify02_null_test.go:28: encoding/json control: err=count is required, value={Value:0 Valid:false}
--- PASS: TestVerify02NullOrdinaryDecoderContract (0.00s)
=== RUN   TestVerify02NullFailurePreservesCaller
    verify02_null_test.go:34: Call result:null: err=bad reply: count is required, value={Value:7 Valid:true}
--- PASS: TestVerify02NullFailurePreservesCaller (0.00s)
=== RUN   TestVerify02NonNullFailurePreservesCaller
--- PASS: TestVerify02NonNullFailurePreservesCaller (0.00s)
=== RUN   TestVerify02OrdinaryDecoderSuccess
--- PASS: TestVerify02OrdinaryDecoderSuccess (0.00s)
=== RUN   TestVerify02DirectPingProtocolErrorLogged
=== RUN   TestVerify02DirectPingProtocolErrorLogged/{"type":"pong","protocol":21}
    verify02_pinglog_test.go:30: returned=herdr protocol 21, hive needs 22
        {"time":"2026-10-18T05:21:57.88928452-04:00","level":"DEBUG","msg":"herdr request","request_id":"47d0a4c8-175","method":"ping","duration_ms":0,"body":"{\"id\":\"47d0a4c8-175\",\"method\":\"ping\",\"params\":{}}"}
        {"time":"2026-10-18T05:21:57.8893384-04:00","level":"DEBUG","msg":"herdr reply","request_id":"47d0a4c8-175","method":"ping","duration_ms":0,"body":"{\"id\":\"47d0a4c8-175\",\"result\":{\"type\":\"pong\",\"protocol\":21}}"}
        {"time":"2026-10-18T05:21:57.889384375-04:00","level":"DEBUG","msg":"herdr error","request_id":"47d0a4c8-175","method":"ping","duration_ms":0,"error":"herdr protocol 21, hive needs 22"}
=== RUN   TestVerify02DirectPingProtocolErrorLogged/{"type":"not-pong","protocol":22}
    verify02_pinglog_test.go:30: returned=bad reply: result type "not-pong", want pong
        {"time":"2026-10-18T05:21:57.889766085-04:00","level":"DEBUG","msg":"herdr request","request_id":"47d0a4c8-176","method":"ping","duration_ms":0,"body":"{\"id\":\"47d0a4c8-176\",\"method\":\"ping\",\"params\":{}}"}
        {"time":"2026-10-18T05:21:57.889836065-04:00","level":"DEBUG","msg":"herdr reply","request_id":"47d0a4c8-176","method":"ping","duration_ms":0,"body":"{\"id\":\"47d0a4c8-176\",\"result\":{\"type\":\"not-pong\",\"protocol\":22}}"}
        {"time":"2026-10-18T05:21:57.889880498-04:00","level":"DEBUG","msg":"herdr error","request_id":"47d0a4c8-176","method":"ping","duration_ms":0,"error":"bad reply: result type \"not-pong\", want pong"}
--- PASS: TestVerify02DirectPingProtocolErrorLogged (0.00s)
    --- PASS: TestVerify02DirectPingProtocolErrorLogged/{"type":"pong","protocol":21} (0.00s)
    --- PASS: TestVerify02DirectPingProtocolErrorLogged/{"type":"not-pong","protocol":22} (0.00s)
=== RUN   TestPrimaryGateWaitDeadline
--- PASS: TestPrimaryGateWaitDeadline (0.60s)
=== RUN   TestPrimaryTransport
=== RUN   TestPrimaryTransport/empty
=== RUN   TestPrimaryTransport/partial
=== RUN   TestPrimaryTransport/http
=== RUN   TestPrimaryTransport/array
=== RUN   TestPrimaryTransport/slow
--- PASS: TestPrimaryTransport (0.12s)
    --- PASS: TestPrimaryTransport/empty (0.00s)
    --- PASS: TestPrimaryTransport/partial (0.00s)
    --- PASS: TestPrimaryTransport/http (0.00s)
    --- PASS: TestPrimaryTransport/array (0.00s)
    --- PASS: TestPrimaryTransport/slow (0.12s)
=== RUN   TestPrimaryCallerDeadlineAndCancel
=== RUN   TestPrimaryCallerDeadlineAndCancel/false
=== RUN   TestPrimaryCallerDeadlineAndCancel/true
--- PASS: TestPrimaryCallerDeadlineAndCancel (0.40s)
    --- PASS: TestPrimaryCallerDeadlineAndCancel/false (0.20s)
    --- PASS: TestPrimaryCallerDeadlineAndCancel/true (0.20s)
=== RUN   TestVerify03NullReferenceContract
--- PASS: TestVerify03NullReferenceContract (0.00s)
=== RUN   TestVerify03NullReferenceFailurePreservesCaller
    verify03_null_reference_test.go:34: null err=bad reply: count is required count=7 valid=true same-pointer=true
--- PASS: TestVerify03NullReferenceFailurePreservesCaller (0.00s)
=== RUN   TestVerify03NonNullReferenceFailurePreservesCaller
--- PASS: TestVerify03NonNullReferenceFailurePreservesCaller (0.00s)
=== RUN   TestVerify03ReferenceSuccess
--- PASS: TestVerify03ReferenceSuccess (0.00s)
=== RUN   TestVerify03PingBoundedError
    verify03_ping_bound_test.go:29: actual id=47d0a4c8-190 method=ping duration=1 error-bytes=4096 valid-UTF8=true truncated=true
--- PASS: TestVerify03PingBoundedError (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	12.365s

exit 0
```

### V2 and V5 actual transcript

Source: `46b88c08`. Capture: `v2-v5-transcript.txt`.

```text
HERDR_ENV present
HERDR_PANE_ID present
HERDR_SOCKET_PATH present
HERDR_WORKSPACE_ID present
HERDR_TAB_ID present
HERDR_TERMINAL_ID absent
## V2 live doctor JSON
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v2-zbv0drvd --json doctor
{"outcome":"ok","phase":"complete","effects":[],"error":null,"result":[{"name":"herdr socket","status":"ok","required":true,"detail":"/home/penguin/.config/herdr/herdr.sock protocol 22"},{"name":"herdr pane","status":"ok","required":false,"detail":"wA9:p1"},{"name":"home","status":"ok","required":true,"detail":"/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v2-zbv0drvd"},{"name":"state store","status":"ok","required":true,"detail":"0 records"},{"name":"git","status":"ok","required":true,"detail":"git version 2.56.0"},{"name":"supervisor","status":"warn","required":false,"detail":"no supervisor claim"},{"name":"harness claude","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness codex","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness opencode","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness pi","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness cursor-agent","status":"warn","required":false,"detail":"not on PATH"},{"name":"gh","status":"warn","required":false,"detail":"not on PATH"}]}
exit 0 elapsed 0.0055
## V2 ping log
{"body": "{\"id\":\"0ef5ca07-1\",\"method\":\"ping\",\"params\":{}}", "component": "doctor", "duration_ms": 0, "level": "DEBUG", "method": "ping", "msg": "herdr request", "request_id": "0ef5ca07-1", "time": "2026-10-18T05:24:47.70488576-04:00"}
{"body": "{\"id\":\"0ef5ca07-1\",\"result\":{\"type\":\"pong\",\"version\":\"0.9.5\",\"protocol\":22,\"capabilities\":{\"live_handoff\":true,\"detached_server_daemon\":true,\"endpoint_protocol_generation\":1,\"surface_interest\":true,\"health_check\":true,\"ssh_agent_registration\":true}}}", "component": "doctor", "duration_ms": 0, "level": "DEBUG", "method": "ping", "msg": "herdr reply", "request_id": "0ef5ca07-1", "time": "2026-10-18T05:24:47.70495589-04:00"}
## V5 live text
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v5-49u5u9ts doctor
ok herdr socket: /home/penguin/.config/herdr/herdr.sock protocol 22
ok herdr pane: wA9:p1
ok home: /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v5-49u5u9ts
ok state store: 0 records
ok git: git version 2.56.0
warn supervisor: no supervisor claim
warn harness claude: not on PATH
warn harness codex: not on PATH
warn harness opencode: not on PATH
warn harness pi: not on PATH
warn harness cursor-agent: not on PATH
warn gh: not on PATH
exit 0 elapsed 0.005
## V5 missing socket
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v5-49u5u9ts doctor
fail herdr socket: herdr socket: /tmp/none.sock: no such file
ok herdr pane: wA9:p1
ok home: /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v5-49u5u9ts
ok state store: 0 records
ok git: git version 2.56.0
warn supervisor: no supervisor claim
warn harness claude: not on PATH
warn harness codex: not on PATH
warn harness opencode: not on PATH
warn harness pi: not on PATH
warn harness cursor-agent: not on PATH
warn gh: not on PATH
exit 1 elapsed 0.0048
## V5 unset socket
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v5-49u5u9ts doctor
fail herdr socket: HERDR_SOCKET_PATH is not set
ok herdr pane: wA9:p1
ok home: /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v5-49u5u9ts
ok state store: 0 records
ok git: git version 2.56.0
warn supervisor: no supervisor claim
warn harness claude: not on PATH
warn harness codex: not on PATH
warn harness opencode: not on PATH
warn harness pi: not on PATH
warn harness cursor-agent: not on PATH
warn gh: not on PATH
exit 1 elapsed 0.0047
## V5 version
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v5-49u5u9ts --json version
{"outcome":"ok","phase":"complete","effects":[],"error":null,"result":{"version":"dev","commit":"46b88c087f6d","modified":false,"go":"go1.27.2-X:nodwarf5"}}
exit 0 elapsed 0.0011
## V5 failed JSON
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v5-49u5u9ts --json doctor
{"outcome":"failed","phase":"complete","effects":[],"error":null,"result":[{"name":"herdr socket","status":"fail","required":true,"detail":"HERDR_SOCKET_PATH is not set"},{"name":"herdr pane","status":"ok","required":false,"detail":"wA9:p1"},{"name":"home","status":"ok","required":true,"detail":"/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v5-49u5u9ts"},{"name":"state store","status":"ok","required":true,"detail":"0 records"},{"name":"git","status":"ok","required":true,"detail":"git version 2.56.0"},{"name":"supervisor","status":"warn","required":false,"detail":"no supervisor claim"},{"name":"harness claude","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness codex","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness opencode","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness pi","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness cursor-agent","status":"warn","required":false,"detail":"not on PATH"},{"name":"gh","status":"warn","required":false,"detail":"not on PATH"}]}
exit 1 elapsed 0.0048
## home-free version
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /t05v04-absent-home version
hive dev (commit 46b88c087f6d, go1.27.2-X:nodwarf5)
exit 0 elapsed 0.0011
## home-free version
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /t05v04-absent-home version --json
{"outcome":"ok","phase":"complete","effects":[],"error":null,"result":{"version":"dev","commit":"46b88c087f6d","modified":false,"go":"go1.27.2-X:nodwarf5"}}
exit 0 elapsed 0.0011
## seven actual CLI process attempts
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-logging-_719fa07 doctor
fail herdr socket: HERDR_SOCKET_PATH is not set
ok herdr pane: wA9:p1
ok home: /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-logging-_719fa07
ok state store: 0 records
ok git: git version 2.56.0
warn supervisor: no supervisor claim
warn harness claude: not on PATH
warn harness codex: not on PATH
warn harness opencode: not on PATH
warn harness pi: not on PATH
warn harness cursor-agent: not on PATH
warn gh: not on PATH
exit 1 elapsed 0.0019
process starts 7 ends 7 failed starts 6
{"command": "git --version", "component": "doctor", "duration_ms": 0, "exit_code": 0, "level": "DEBUG", "msg": "process end", "time": "2026-10-18T05:24:47.733895855-04:00"}
{"command": "sh -c command -v \"$1\" sh claude", "component": "doctor", "duration_ms": 0, "error": "start sh: exec: \"sh\": executable file not found in $PATH", "exit_code": -1, "level": "DEBUG", "msg": "process end", "time": "2026-10-18T05:24:47.733970413-04:00"}
{"command": "sh -c command -v \"$1\" sh codex", "component": "doctor", "duration_ms": 0, "error": "start sh: exec: \"sh\": executable file not found in $PATH", "exit_code": -1, "level": "DEBUG", "msg": "process end", "time": "2026-10-18T05:24:47.733988206-04:00"}
{"command": "sh -c command -v \"$1\" sh opencode", "component": "doctor", "duration_ms": 0, "error": "start sh: exec: \"sh\": executable file not found in $PATH", "exit_code": -1, "level": "DEBUG", "msg": "process end", "time": "2026-10-18T05:24:47.734003465-04:00"}
{"command": "sh -c command -v \"$1\" sh pi", "component": "doctor", "duration_ms": 0, "error": "start sh: exec: \"sh\": executable file not found in $PATH", "exit_code": -1, "level": "DEBUG", "msg": "process end", "time": "2026-10-18T05:24:47.734025826-04:00"}
{"command": "sh -c command -v \"$1\" sh cursor-agent", "component": "doctor", "duration_ms": 0, "error": "start sh: exec: \"sh\": executable file not found in $PATH", "exit_code": -1, "level": "DEBUG", "msg": "process end", "time": "2026-10-18T05:24:47.734040664-04:00"}
{"command": "gh --version", "component": "doctor", "duration_ms": 0, "error": "start gh: exec: \"gh\": executable file not found in $PATH", "exit_code": -1, "level": "DEBUG", "msg": "process end", "time": "2026-10-18T05:24:47.734056133-04:00"}
```

### V3 supplied suite

Source: `46b88c08`. Capture: `v3.txt`.

```text
commit 46b88c087f6de501cda589f017a4e3a794ab7f81 (archive identity is in run name when different)
$ go test -race -count=1 ./internal/lib/run -v
=== RUN   TestCommandTimeout
--- PASS: TestCommandTimeout (1.00s)
=== RUN   TestCommandTimeoutKillsChildThatHoldsPipes
--- PASS: TestCommandTimeoutKillsChildThatHoldsPipes (1.01s)
=== RUN   TestCommandOutputCap
=== RUN   TestCommandOutputCap/1048575
=== RUN   TestCommandOutputCap/1048576
=== RUN   TestCommandOutputCap/1048577
=== RUN   TestCommandOutputCap/2097152
--- PASS: TestCommandOutputCap (0.02s)
    --- PASS: TestCommandOutputCap/1048575 (0.01s)
    --- PASS: TestCommandOutputCap/1048576 (0.01s)
    --- PASS: TestCommandOutputCap/1048577 (0.01s)
    --- PASS: TestCommandOutputCap/2097152 (0.01s)
=== RUN   TestCommandStderrCap
--- PASS: TestCommandStderrCap (0.01s)
=== RUN   TestCommandExitCodeIsNotAnError
--- PASS: TestCommandExitCodeIsNotAnError (0.00s)
=== RUN   TestCommandMissingBinary
--- PASS: TestCommandMissingBinary (0.00s)
=== RUN   TestCommandReadsEndOfFileFromStdin
--- PASS: TestCommandReadsEndOfFileFromStdin (0.00s)
=== RUN   TestCommandUsesDir
--- PASS: TestCommandUsesDir (0.00s)
=== RUN   TestCommandLogsStartAndEnd
--- PASS: TestCommandLogsStartAndEnd (0.00s)
=== RUN   TestCommandDeadlineWithInheritedPipeChild
--- PASS: TestCommandDeadlineWithInheritedPipeChild (0.11s)
=== RUN   TestCommandCancellationWithInheritedPipeChild
--- PASS: TestCommandCancellationWithInheritedPipeChild (0.11s)
=== RUN   TestCommandErrorStaysOneLine
=== RUN   TestCommandErrorStaysOneLine/name_newline
=== RUN   TestCommandErrorStaysOneLine/name_return
=== RUN   TestCommandErrorStaysOneLine/dir_newline
=== RUN   TestCommandErrorStaysOneLine/dir_return
--- PASS: TestCommandErrorStaysOneLine (0.00s)
    --- PASS: TestCommandErrorStaysOneLine/name_newline (0.00s)
    --- PASS: TestCommandErrorStaysOneLine/name_return (0.00s)
    --- PASS: TestCommandErrorStaysOneLine/dir_newline (0.00s)
    --- PASS: TestCommandErrorStaysOneLine/dir_return (0.00s)
=== RUN   TestCommandCancellationLogsReturnedError
--- PASS: TestCommandCancellationLogsReturnedError (0.10s)
=== RUN   TestCommandTimeoutLogsReturnedError
--- PASS: TestCommandTimeoutLogsReturnedError (0.10s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/run	3.472s

exit 0
```

### V4 supplied suite

Source: `46b88c08`. Capture: `v4.txt`.

```text
commit 46b88c087f6de501cda589f017a4e3a794ab7f81 (archive identity is in run name when different)
$ go test -race -count=1 ./internal/doctor -v
=== RUN   TestRunChecksWithFakeHerdr
--- PASS: TestRunChecksWithFakeHerdr (0.00s)
=== RUN   TestRunWithoutHerdrSocket
--- PASS: TestRunWithoutHerdrSocket (0.00s)
=== RUN   TestRunWithMissingSocket
--- PASS: TestRunWithMissingSocket (0.00s)
=== RUN   TestRunPaneWarnWithoutHerdrEnv
--- PASS: TestRunPaneWarnWithoutHerdrEnv (0.00s)
=== RUN   TestRunGHAbsentWarn
--- PASS: TestRunGHAbsentWarn (0.00s)
=== RUN   TestRunHeldLock
--- PASS: TestRunHeldLock (2.00s)
=== RUN   TestRunLinkedRecord
--- PASS: TestRunLinkedRecord (0.00s)
=== RUN   TestRunLinkedCollection
--- PASS: TestRunLinkedCollection (0.00s)
=== RUN   TestRunLinkedStateRoot
--- PASS: TestRunLinkedStateRoot (0.00s)
=== RUN   TestRunLinkedHiveAncestor
--- PASS: TestRunLinkedHiveAncestor (0.00s)
=== RUN   TestRunHomeWithoutMarker
--- PASS: TestRunHomeWithoutMarker (0.00s)
=== RUN   TestRunStateStoreMakesNoRecord
--- PASS: TestRunStateStoreMakesNoRecord (0.00s)
=== RUN   TestRender
--- PASS: TestRender (0.00s)
=== RUN   TestResultJSONFields
--- PASS: TestResultJSONFields (0.00s)
=== RUN   TestRunLogsProcessEvents
--- PASS: TestRunLogsProcessEvents (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/doctor	3.023s

exit 0
```

### V3 and V4 complete secondary probes

Source: `46b88c08`. Capture: `secondary.txt`.

```text
commit 46b88c087f6de501cda589f017a4e3a794ab7f81 (archive identity is in run name when different)
$ go test -race -count=1 ./internal/lib/run ./internal/doctor -run TestSecondary|TestPrimaryDoctor -v
=== RUN   TestSecondaryCancellationErrorLog
--- PASS: TestSecondaryCancellationErrorLog (0.05s)
=== RUN   TestSecondaryConcurrentCancellation
--- PASS: TestSecondaryConcurrentCancellation (0.05s)
=== RUN   TestSecondaryRunnerLiteralCapsDrain
=== RUN   TestSecondaryRunnerLiteralCapsDrain/1048575
=== RUN   TestSecondaryRunnerLiteralCapsDrain/1048576
=== RUN   TestSecondaryRunnerLiteralCapsDrain/1048577
=== RUN   TestSecondaryRunnerLiteralCapsDrain/8388608
--- PASS: TestSecondaryRunnerLiteralCapsDrain (0.06s)
    --- PASS: TestSecondaryRunnerLiteralCapsDrain/1048575 (0.02s)
    --- PASS: TestSecondaryRunnerLiteralCapsDrain/1048576 (0.01s)
    --- PASS: TestSecondaryRunnerLiteralCapsDrain/1048577 (0.01s)
    --- PASS: TestSecondaryRunnerLiteralCapsDrain/8388608 (0.02s)
=== RUN   TestSecondaryRunnerExactCaps
--- PASS: TestSecondaryRunnerExactCaps (0.04s)
=== RUN   TestSecondaryRunnerArgvStdinExit
--- PASS: TestSecondaryRunnerArgvStdinExit (0.01s)
=== RUN   TestSecondaryRunnerStoppedGroup
--- PASS: TestSecondaryRunnerStoppedGroup (0.15s)
=== RUN   TestSecondaryRunnerCancellation
--- PASS: TestSecondaryRunnerCancellation (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesDeadline
    verify02_runner_test.go:47: elapsed=100.367263ms err=process timeout result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesDeadline (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesCancel
    verify02_runner_test.go:51: elapsed=100.24236ms err=context canceled result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesCancel (0.10s)
=== RUN   TestSecondaryRunnerBadStringsOneLine
--- PASS: TestSecondaryRunnerBadStringsOneLine (0.00s)
=== RUN   TestSecondaryRunnerLog
    verify02_runner_test.go:53: four process events with command, end exit/duration, start failure error
--- PASS: TestSecondaryRunnerLog (0.00s)
=== RUN   TestSecondaryRunnerDefaultDeadline
    verify02_runner_test.go:54: default elapsed=30.000195295s err=process timeout
--- PASS: TestSecondaryRunnerDefaultDeadline (30.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/run	31.679s
=== RUN   TestSecondaryDoctorOrderStatusesPreservation
    verify02_doctor_test.go:28: literal order, all statuses, real git, fake dependencies, four unchanged domain records
--- PASS: TestSecondaryDoctorOrderStatusesPreservation (0.01s)
=== RUN   TestSecondaryDoctorGHAndMissingDeps
=== RUN   TestSecondaryDoctorGHAndMissingDeps/absent
=== RUN   TestSecondaryDoctorGHAndMissingDeps/versionfail
=== RUN   TestSecondaryDoctorGHAndMissingDeps/authfail
=== RUN   TestSecondaryDoctorGHAndMissingDeps/success
=== RUN   TestSecondaryDoctorGHAndMissingDeps/missing_git
--- PASS: TestSecondaryDoctorGHAndMissingDeps (0.02s)
    --- PASS: TestSecondaryDoctorGHAndMissingDeps/absent (0.00s)
    --- PASS: TestSecondaryDoctorGHAndMissingDeps/versionfail (0.00s)
    --- PASS: TestSecondaryDoctorGHAndMissingDeps/authfail (0.00s)
    --- PASS: TestSecondaryDoctorGHAndMissingDeps/success (0.00s)
    --- PASS: TestSecondaryDoctorGHAndMissingDeps/missing_git (0.00s)
=== RUN   TestSecondaryDoctorEmptyStoreMetadata
--- PASS: TestSecondaryDoctorEmptyStoreMetadata (0.00s)
=== RUN   TestSecondaryDoctorRecordInputs
--- PASS: TestSecondaryDoctorRecordInputs (0.09s)
=== RUN   TestSecondaryDoctorBadStaticPaths
=== RUN   TestSecondaryDoctorBadStaticPaths/.hivesymlink
=== RUN   TestSecondaryDoctorBadStaticPaths/.hivefifo
=== RUN   TestSecondaryDoctorBadStaticPaths/.hivesocket
=== RUN   TestSecondaryDoctorBadStaticPaths/.hive/statesymlink
=== RUN   TestSecondaryDoctorBadStaticPaths/.hive/statefifo
=== RUN   TestSecondaryDoctorBadStaticPaths/.hive/statesocket
=== RUN   TestSecondaryDoctorBadStaticPaths/.hive/state/locksymlink
=== RUN   TestSecondaryDoctorBadStaticPaths/.hive/state/lockfifo
=== RUN   TestSecondaryDoctorBadStaticPaths/.hive/state/locksocket
=== RUN   TestSecondaryDoctorBadStaticPaths/.hive/state/taskssymlink
=== RUN   TestSecondaryDoctorBadStaticPaths/.hive/state/tasksfifo
=== RUN   TestSecondaryDoctorBadStaticPaths/.hive/state/taskssocket
=== RUN   TestSecondaryDoctorBadStaticPaths/.hive/state/tasks/r.jsonsymlink
=== RUN   TestSecondaryDoctorBadStaticPaths/.hive/state/tasks/r.jsonfifo
=== RUN   TestSecondaryDoctorBadStaticPaths/.hive/state/tasks/r.jsonsocket
--- PASS: TestSecondaryDoctorBadStaticPaths (0.00s)
    --- PASS: TestSecondaryDoctorBadStaticPaths/.hivesymlink (0.00s)
    --- PASS: TestSecondaryDoctorBadStaticPaths/.hivefifo (0.00s)
    --- PASS: TestSecondaryDoctorBadStaticPaths/.hivesocket (0.00s)
    --- PASS: TestSecondaryDoctorBadStaticPaths/.hive/statesymlink (0.00s)
    --- PASS: TestSecondaryDoctorBadStaticPaths/.hive/statefifo (0.00s)
    --- PASS: TestSecondaryDoctorBadStaticPaths/.hive/statesocket (0.00s)
    --- PASS: TestSecondaryDoctorBadStaticPaths/.hive/state/locksymlink (0.00s)
    --- PASS: TestSecondaryDoctorBadStaticPaths/.hive/state/lockfifo (0.00s)
    --- PASS: TestSecondaryDoctorBadStaticPaths/.hive/state/locksocket (0.00s)
    --- PASS: TestSecondaryDoctorBadStaticPaths/.hive/state/taskssymlink (0.00s)
    --- PASS: TestSecondaryDoctorBadStaticPaths/.hive/state/tasksfifo (0.00s)
    --- PASS: TestSecondaryDoctorBadStaticPaths/.hive/state/taskssocket (0.00s)
    --- PASS: TestSecondaryDoctorBadStaticPaths/.hive/state/tasks/r.jsonsymlink (0.00s)
    --- PASS: TestSecondaryDoctorBadStaticPaths/.hive/state/tasks/r.jsonfifo (0.00s)
    --- PASS: TestSecondaryDoctorBadStaticPaths/.hive/state/tasks/r.jsonsocket (0.00s)
=== RUN   TestSecondaryDoctorHeldLockBound
    verify02_doctor_test.go:33: lock elapsed=2.000595413s check={Name:state store Status:fail Required:true Detail:store lock timeout}
--- PASS: TestSecondaryDoctorHeldLockBound (2.00s)
=== RUN   TestSecondaryDoctorBadStrings
--- PASS: TestSecondaryDoctorBadStrings (0.00s)
=== RUN   TestSecondaryDoctorProcessLogging
--- PASS: TestSecondaryDoctorProcessLogging (0.01s)
=== RUN   TestSecondaryDoctorUnreadableRecord
--- PASS: TestSecondaryDoctorUnreadableRecord (0.00s)
=== RUN   TestSecondaryDoctorConcurrentStore
--- PASS: TestSecondaryDoctorConcurrentStore (0.01s)
=== RUN   TestPrimaryDoctorProcessEvents
--- PASS: TestPrimaryDoctorProcessEvents (0.01s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/doctor	3.147s

exit 0
```

### V3 and V4 timing and concurrency, ten rounds

Source: `46b88c08`. Capture: `secondary-ten.txt`.

```text
commit 46b88c087f6de501cda589f017a4e3a794ab7f81 (archive identity is in run name when different)
$ go test -race -count=10 ./internal/lib/run ./internal/doctor -run TestSecondaryRunnerStoppedGroup|TestSecondaryRunnerCancellation|TestSecondaryRunnerInheritedPipes|TestSecondaryCancellationErrorLog|TestSecondaryConcurrentCancellation|TestSecondaryDoctorConcurrentStore|TestSecondaryDoctorUnreadableRecord -v
=== RUN   TestSecondaryCancellationErrorLog
--- PASS: TestSecondaryCancellationErrorLog (0.05s)
=== RUN   TestSecondaryConcurrentCancellation
--- PASS: TestSecondaryConcurrentCancellation (0.05s)
=== RUN   TestSecondaryRunnerStoppedGroup
--- PASS: TestSecondaryRunnerStoppedGroup (0.15s)
=== RUN   TestSecondaryRunnerCancellation
--- PASS: TestSecondaryRunnerCancellation (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesDeadline
    verify02_runner_test.go:47: elapsed=100.243678ms err=process timeout result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesDeadline (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesCancel
    verify02_runner_test.go:51: elapsed=100.285586ms err=context canceled result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesCancel (0.10s)
=== RUN   TestSecondaryCancellationErrorLog
--- PASS: TestSecondaryCancellationErrorLog (0.05s)
=== RUN   TestSecondaryConcurrentCancellation
--- PASS: TestSecondaryConcurrentCancellation (0.05s)
=== RUN   TestSecondaryRunnerStoppedGroup
--- PASS: TestSecondaryRunnerStoppedGroup (0.15s)
=== RUN   TestSecondaryRunnerCancellation
--- PASS: TestSecondaryRunnerCancellation (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesDeadline
    verify02_runner_test.go:47: elapsed=100.285596ms err=process timeout result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesDeadline (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesCancel
    verify02_runner_test.go:51: elapsed=100.269857ms err=context canceled result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesCancel (0.10s)
=== RUN   TestSecondaryCancellationErrorLog
--- PASS: TestSecondaryCancellationErrorLog (0.05s)
=== RUN   TestSecondaryConcurrentCancellation
--- PASS: TestSecondaryConcurrentCancellation (0.05s)
=== RUN   TestSecondaryRunnerStoppedGroup
--- PASS: TestSecondaryRunnerStoppedGroup (0.15s)
=== RUN   TestSecondaryRunnerCancellation
--- PASS: TestSecondaryRunnerCancellation (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesDeadline
    verify02_runner_test.go:47: elapsed=100.198404ms err=process timeout result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesDeadline (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesCancel
    verify02_runner_test.go:51: elapsed=100.254938ms err=context canceled result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesCancel (0.10s)
=== RUN   TestSecondaryCancellationErrorLog
--- PASS: TestSecondaryCancellationErrorLog (0.05s)
=== RUN   TestSecondaryConcurrentCancellation
--- PASS: TestSecondaryConcurrentCancellation (0.05s)
=== RUN   TestSecondaryRunnerStoppedGroup
--- PASS: TestSecondaryRunnerStoppedGroup (0.15s)
=== RUN   TestSecondaryRunnerCancellation
--- PASS: TestSecondaryRunnerCancellation (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesDeadline
    verify02_runner_test.go:47: elapsed=100.243979ms err=process timeout result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesDeadline (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesCancel
    verify02_runner_test.go:51: elapsed=100.257884ms err=context canceled result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesCancel (0.10s)
=== RUN   TestSecondaryCancellationErrorLog
--- PASS: TestSecondaryCancellationErrorLog (0.05s)
=== RUN   TestSecondaryConcurrentCancellation
--- PASS: TestSecondaryConcurrentCancellation (0.05s)
=== RUN   TestSecondaryRunnerStoppedGroup
--- PASS: TestSecondaryRunnerStoppedGroup (0.15s)
=== RUN   TestSecondaryRunnerCancellation
--- PASS: TestSecondaryRunnerCancellation (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesDeadline
    verify02_runner_test.go:47: elapsed=100.257273ms err=process timeout result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesDeadline (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesCancel
    verify02_runner_test.go:51: elapsed=100.311895ms err=context canceled result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesCancel (0.10s)
=== RUN   TestSecondaryCancellationErrorLog
--- PASS: TestSecondaryCancellationErrorLog (0.05s)
=== RUN   TestSecondaryConcurrentCancellation
--- PASS: TestSecondaryConcurrentCancellation (0.05s)
=== RUN   TestSecondaryRunnerStoppedGroup
--- PASS: TestSecondaryRunnerStoppedGroup (0.15s)
=== RUN   TestSecondaryRunnerCancellation
--- PASS: TestSecondaryRunnerCancellation (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesDeadline
    verify02_runner_test.go:47: elapsed=100.263935ms err=process timeout result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesDeadline (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesCancel
    verify02_runner_test.go:51: elapsed=100.27134ms err=context canceled result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesCancel (0.10s)
=== RUN   TestSecondaryCancellationErrorLog
--- PASS: TestSecondaryCancellationErrorLog (0.05s)
=== RUN   TestSecondaryConcurrentCancellation
--- PASS: TestSecondaryConcurrentCancellation (0.05s)
=== RUN   TestSecondaryRunnerStoppedGroup
--- PASS: TestSecondaryRunnerStoppedGroup (0.15s)
=== RUN   TestSecondaryRunnerCancellation
--- PASS: TestSecondaryRunnerCancellation (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesDeadline
    verify02_runner_test.go:47: elapsed=100.243909ms err=process timeout result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesDeadline (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesCancel
    verify02_runner_test.go:51: elapsed=100.259748ms err=context canceled result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesCancel (0.10s)
=== RUN   TestSecondaryCancellationErrorLog
--- PASS: TestSecondaryCancellationErrorLog (0.05s)
=== RUN   TestSecondaryConcurrentCancellation
--- PASS: TestSecondaryConcurrentCancellation (0.05s)
=== RUN   TestSecondaryRunnerStoppedGroup
--- PASS: TestSecondaryRunnerStoppedGroup (0.15s)
=== RUN   TestSecondaryRunnerCancellation
--- PASS: TestSecondaryRunnerCancellation (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesDeadline
    verify02_runner_test.go:47: elapsed=100.275417ms err=process timeout result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesDeadline (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesCancel
    verify02_runner_test.go:51: elapsed=100.272131ms err=context canceled result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesCancel (0.10s)
=== RUN   TestSecondaryCancellationErrorLog
--- PASS: TestSecondaryCancellationErrorLog (0.05s)
=== RUN   TestSecondaryConcurrentCancellation
--- PASS: TestSecondaryConcurrentCancellation (0.05s)
=== RUN   TestSecondaryRunnerStoppedGroup
--- PASS: TestSecondaryRunnerStoppedGroup (0.15s)
=== RUN   TestSecondaryRunnerCancellation
--- PASS: TestSecondaryRunnerCancellation (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesDeadline
    verify02_runner_test.go:47: elapsed=100.241614ms err=process timeout result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesDeadline (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesCancel
    verify02_runner_test.go:51: elapsed=100.22328ms err=context canceled result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesCancel (0.10s)
=== RUN   TestSecondaryCancellationErrorLog
--- PASS: TestSecondaryCancellationErrorLog (0.05s)
=== RUN   TestSecondaryConcurrentCancellation
--- PASS: TestSecondaryConcurrentCancellation (0.05s)
=== RUN   TestSecondaryRunnerStoppedGroup
--- PASS: TestSecondaryRunnerStoppedGroup (0.15s)
=== RUN   TestSecondaryRunnerCancellation
--- PASS: TestSecondaryRunnerCancellation (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesDeadline
    verify02_runner_test.go:47: elapsed=100.264927ms err=process timeout result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesDeadline (0.10s)
=== RUN   TestSecondaryRunnerInheritedPipesCancel
    verify02_runner_test.go:51: elapsed=100.222598ms err=context canceled result=<nil> child reaped without cleanup signal
--- PASS: TestSecondaryRunnerInheritedPipesCancel (0.10s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/run	6.598s
=== RUN   TestSecondaryDoctorUnreadableRecord
--- PASS: TestSecondaryDoctorUnreadableRecord (0.00s)
=== RUN   TestSecondaryDoctorConcurrentStore
--- PASS: TestSecondaryDoctorConcurrentStore (0.01s)
=== RUN   TestSecondaryDoctorUnreadableRecord
--- PASS: TestSecondaryDoctorUnreadableRecord (0.00s)
=== RUN   TestSecondaryDoctorConcurrentStore
--- PASS: TestSecondaryDoctorConcurrentStore (0.01s)
=== RUN   TestSecondaryDoctorUnreadableRecord
--- PASS: TestSecondaryDoctorUnreadableRecord (0.00s)
=== RUN   TestSecondaryDoctorConcurrentStore
--- PASS: TestSecondaryDoctorConcurrentStore (0.01s)
=== RUN   TestSecondaryDoctorUnreadableRecord
--- PASS: TestSecondaryDoctorUnreadableRecord (0.00s)
=== RUN   TestSecondaryDoctorConcurrentStore
--- PASS: TestSecondaryDoctorConcurrentStore (0.01s)
=== RUN   TestSecondaryDoctorUnreadableRecord
--- PASS: TestSecondaryDoctorUnreadableRecord (0.00s)
=== RUN   TestSecondaryDoctorConcurrentStore
--- PASS: TestSecondaryDoctorConcurrentStore (0.01s)
=== RUN   TestSecondaryDoctorUnreadableRecord
--- PASS: TestSecondaryDoctorUnreadableRecord (0.00s)
=== RUN   TestSecondaryDoctorConcurrentStore
--- PASS: TestSecondaryDoctorConcurrentStore (0.01s)
=== RUN   TestSecondaryDoctorUnreadableRecord
--- PASS: TestSecondaryDoctorUnreadableRecord (0.00s)
=== RUN   TestSecondaryDoctorConcurrentStore
--- PASS: TestSecondaryDoctorConcurrentStore (0.01s)
=== RUN   TestSecondaryDoctorUnreadableRecord
--- PASS: TestSecondaryDoctorUnreadableRecord (0.00s)
=== RUN   TestSecondaryDoctorConcurrentStore
--- PASS: TestSecondaryDoctorConcurrentStore (0.01s)
=== RUN   TestSecondaryDoctorUnreadableRecord
--- PASS: TestSecondaryDoctorUnreadableRecord (0.00s)
=== RUN   TestSecondaryDoctorConcurrentStore
--- PASS: TestSecondaryDoctorConcurrentStore (0.01s)
=== RUN   TestSecondaryDoctorUnreadableRecord
--- PASS: TestSecondaryDoctorUnreadableRecord (0.00s)
=== RUN   TestSecondaryDoctorConcurrentStore
--- PASS: TestSecondaryDoctorConcurrentStore (0.01s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/doctor	1.138s

exit 0
```

### V3 signal recorder, no signals

Source: `46b88c08 with recorder seam`. Capture: `signal-recorder.txt`.

```text
$ setsid go test -count=1 ./internal/lib/run -run TestSecondarySignalRecorder -v
=== RUN   TestSecondarySignalRecorderActualCommand
    verify02_signal_test.go:5: actual runner selected the same owned negative group target [-571877 -571877], recorder sent no signal
--- PASS: TestSecondarySignalRecorderActualCommand (0.05s)
=== RUN   TestSecondarySignalRecorderAfterExit
    verify02_signal_test.go:6: exited direct child: cancellation after return records no signal
--- PASS: TestSecondarySignalRecorderAfterExit (0.03s)
=== RUN   TestSecondarySignalRecorderSyntheticMetadata
    verify02_signal_test.go:7: synthetic process metadata pid=0 records target=0, no syscall
    verify02_signal_test.go:7: synthetic process metadata pid=1 records target=-1, no syscall
    verify02_signal_test.go:7: synthetic process metadata pid=-1 records target=1, no syscall
    verify02_signal_test.go:7: synthetic process metadata pid=-123 records target=123, no syscall
    verify02_signal_test.go:7: synthetic process metadata pid=571871 records target=-571871, no syscall
    verify02_signal_test.go:7: synthetic process metadata pid=571731 records target=-571731, no syscall
    verify02_signal_test.go:7: synthetic process metadata pid=2147483647 records target=-2147483647, no syscall
    verify02_signal_test.go:7: synthetic only: Command exposes no input PID. String, null, missing, and array PID inputs cannot enter its typed API.
--- PASS: TestSecondarySignalRecorderSyntheticMetadata (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/run	0.082s

exit 0
```

### V1 historical red

Source: `1c45fd8c`. Capture: `v1-red.txt`.

```text
$ go test -race -count=1 ./internal/lib/herdr -v
=== RUN   TestCallSuccessEnvelope
    client_test.go:131: result type = "", want ok
--- FAIL: TestCallSuccessEnvelope (0.00s)
=== RUN   TestCallErrorEnvelope
    client_test.go:150: Call() error = <nil>, want *herdr.Error
--- FAIL: TestCallErrorEnvelope (0.00s)
=== RUN   TestCallRejectsBadReplies
=== RUN   TestCallRejectsBadReplies/another_id
    client_test.go:196: Call() error = <nil>, want bad reply
=== RUN   TestCallRejectsBadReplies/both_result_and_error
    client_test.go:196: Call() error = <nil>, want bad reply
=== RUN   TestCallRejectsBadReplies/neither_result_nor_error
    client_test.go:196: Call() error = <nil>, want bad reply
--- FAIL: TestCallRejectsBadReplies (0.00s)
    --- FAIL: TestCallRejectsBadReplies/another_id (0.00s)
    --- FAIL: TestCallRejectsBadReplies/both_result_and_error (0.00s)
    --- FAIL: TestCallRejectsBadReplies/neither_result_nor_error (0.00s)
=== RUN   TestCallReplyTooLarge
=== RUN   TestCallReplyTooLarge/1048575
=== RUN   TestCallReplyTooLarge/1048576
=== RUN   TestCallReplyTooLarge/1048577
    client_test.go:224: Call() error = <nil>, want reply too large
--- FAIL: TestCallReplyTooLarge (0.00s)
    --- PASS: TestCallReplyTooLarge/1048575 (0.00s)
    --- PASS: TestCallReplyTooLarge/1048576 (0.00s)
    --- FAIL: TestCallReplyTooLarge/1048577 (0.00s)
=== RUN   TestCallReplyTooLargeTenMegabytes
    client_test.go:246: Call() error = <nil>, want reply too large
--- FAIL: TestCallReplyTooLargeTenMegabytes (0.00s)
=== RUN   TestCallDeadlineWithSilentServer
    client_test.go:261: Call() error = nil, want a deadline error
--- FAIL: TestCallDeadlineWithSilentServer (0.00s)
=== RUN   TestCallSlowReply
    client_test.go:279: Call() error = <nil>, want context.DeadlineExceeded
--- FAIL: TestCallSlowReply (0.00s)
=== RUN   TestCallClosedReply
    client_test.go:303: Call() error = <nil>, want bad reply
--- FAIL: TestCallClosedReply (0.00s)
=== RUN   TestCallWithoutSocketPath
    client_test.go:311: Call() error = <nil>, want the unset socket message
--- FAIL: TestCallWithoutSocketPath (0.00s)
=== RUN   TestCallSocketPathTypes
=== RUN   TestCallSocketPathTypes/missing
    client_test.go:321: Call() error = <nil>, want the path and no such file
=== RUN   TestCallSocketPathTypes/directory
    client_test.go:332: Call() error = <nil>, want not a socket
=== RUN   TestCallSocketPathTypes/control_character
    client_test.go:339: Call() error = %!q(<nil>), want one plain line
--- FAIL: TestCallSocketPathTypes (0.00s)
    --- FAIL: TestCallSocketPathTypes/missing (0.00s)
    --- FAIL: TestCallSocketPathTypes/directory (0.00s)
    --- FAIL: TestCallSocketPathTypes/control_character (0.00s)
=== RUN   TestCallRequestSize
=== RUN   TestCallRequestSize/1048575
=== RUN   TestCallRequestSize/1048576
=== RUN   TestCallRequestSize/1048577
    client_test.go:380: Call() error = <nil>, want request too large
--- FAIL: TestCallRequestSize (0.05s)
    --- PASS: TestCallRequestSize/1048575 (0.02s)
    --- PASS: TestCallRequestSize/1048576 (0.02s)
    --- FAIL: TestCallRequestSize/1048577 (0.02s)
=== RUN   TestRequestIDsShareOnePrefix
    client_test.go:416: call IDs = [], want two
--- FAIL: TestRequestIDsShareOnePrefix (0.00s)
=== RUN   TestHelperProcessRequestID
--- PASS: TestHelperProcessRequestID (0.00s)
=== RUN   TestRequestIDsDifferBetweenProcesses
    client_test.go:459: helper output has no request ID: ID=
        PASS
--- FAIL: TestRequestIDsDifferBetweenProcesses (1.00s)
=== RUN   TestCallLogsRequestAndReply
    client_test.go:495: log files = 0, want one
--- FAIL: TestCallLogsRequestAndReply (0.00s)
=== RUN   TestPingChecksProtocol
    protocol_test.go:19: Ping() protocol = 0, want 22
--- FAIL: TestPingChecksProtocol (0.00s)
=== RUN   TestPingRejectsOtherProtocol
    protocol_test.go:30: Ping() error = <nil>, want the protocol message
--- FAIL: TestPingRejectsOtherProtocol (0.00s)
=== RUN   TestProtocolGateRunsPingOnce
    protocol_test.go:48: methods = [], want [ping workspace.list workspace.list]
--- FAIL: TestProtocolGateRunsPingOnce (0.00s)
=== RUN   TestProtocolGateFailureBlocksLaterCalls
    protocol_test.go:64: Call() error = <nil>, want the protocol error
--- FAIL: TestProtocolGateFailureBlocksLaterCalls (0.00s)
=== RUN   TestProtocolGateIsPerSocketPath
    protocol_test.go:89: server 0 methods = [], want ping then workspace.list
--- FAIL: TestProtocolGateIsPerSocketPath (0.00s)
=== RUN   TestCallWithMethodPingDoesNotGate
    protocol_test.go:104: result type = "", want pong
--- FAIL: TestCallWithMethodPingDoesNotGate (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/herdr	1.064s
FAIL

exit 1
```

### V3 historical red

Source: `0679d1be`. Capture: `v3-red.txt`.

```text
$ go test -race -count=1 ./internal/lib/run -v
=== RUN   TestCommandTimeout
    run_test.go:29: Command() error = <nil>, want ErrTimeout
--- FAIL: TestCommandTimeout (0.00s)
=== RUN   TestCommandTimeoutKillsChildThatHoldsPipes
    run_test.go:47: Command() error = <nil>, want ErrTimeout
--- FAIL: TestCommandTimeoutKillsChildThatHoldsPipes (0.00s)
=== RUN   TestCommandOutputCap
=== RUN   TestCommandOutputCap/1048575
    run_test.go:77: stdout = 0 bytes, want 1048575
=== RUN   TestCommandOutputCap/1048576
    run_test.go:77: stdout = 0 bytes, want 1048576
=== RUN   TestCommandOutputCap/1048577
    run_test.go:77: stdout = 0 bytes, want 1048576
=== RUN   TestCommandOutputCap/2097152
    run_test.go:77: stdout = 0 bytes, want 1048576
--- FAIL: TestCommandOutputCap (0.00s)
    --- FAIL: TestCommandOutputCap/1048575 (0.00s)
    --- FAIL: TestCommandOutputCap/1048576 (0.00s)
    --- FAIL: TestCommandOutputCap/1048577 (0.00s)
    --- FAIL: TestCommandOutputCap/2097152 (0.00s)
=== RUN   TestCommandStderrCap
    run_test.go:92: stderr = 0 bytes, want 1048576
--- FAIL: TestCommandStderrCap (0.00s)
=== RUN   TestCommandExitCodeIsNotAnError
    run_test.go:105: ExitCode = 0, want 1
--- FAIL: TestCommandExitCodeIsNotAnError (0.00s)
=== RUN   TestCommandMissingBinary
    run_test.go:111: Command() error = nil, want a start error
--- FAIL: TestCommandMissingBinary (0.00s)
=== RUN   TestCommandReadsEndOfFileFromStdin
--- PASS: TestCommandReadsEndOfFileFromStdin (0.00s)
=== RUN   TestCommandUsesDir
    run_test.go:138: pwd = "", want "/tmp/h5v4.s92j3yft/TestCommandUsesDir3647695564/001"
--- FAIL: TestCommandUsesDir (0.00s)
=== RUN   TestCommandLogsStartAndEnd
    run_test.go:162: log files = 0, want one
--- FAIL: TestCommandLogsStartAndEnd (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/run	0.003s
FAIL

exit 1
```

### V4 historical red

Source: `f558034a`. Capture: `v4-red.txt`.

```text
$ go test -race -count=1 ./internal/doctor -v
=== RUN   TestRunChecksWithFakeHerdr
    checks_test.go:103: check "herdr socket" is absent from []
--- FAIL: TestRunChecksWithFakeHerdr (0.00s)
=== RUN   TestRunWithoutHerdrSocket
    checks_test.go:146: check "herdr socket" is absent from []
--- FAIL: TestRunWithoutHerdrSocket (0.00s)
=== RUN   TestRunWithMissingSocket
    checks_test.go:164: check "herdr socket" is absent from []
--- FAIL: TestRunWithMissingSocket (0.00s)
=== RUN   TestRunPaneWarnWithoutHerdrEnv
    checks_test.go:180: check "herdr pane" is absent from []
--- FAIL: TestRunPaneWarnWithoutHerdrEnv (0.00s)
=== RUN   TestRunGHAbsentWarn
    checks_test.go:197: check "gh" is absent from []
--- FAIL: TestRunGHAbsentWarn (0.00s)
=== RUN   TestRunHeldLock
    checks_test.go:228: check "state store" is absent from []
--- FAIL: TestRunHeldLock (0.00s)
=== RUN   TestRunLinkedRecord
    checks_test.go:254: check "state store" is absent from []
--- FAIL: TestRunLinkedRecord (0.00s)
=== RUN   TestRunLinkedCollection
    checks_test.go:277: check "state store" is absent from []
--- FAIL: TestRunLinkedCollection (0.00s)
=== RUN   TestRunLinkedStateRoot
    checks_test.go:292: check "state store" is absent from []
--- FAIL: TestRunLinkedStateRoot (0.00s)
=== RUN   TestRunLinkedHiveAncestor
    checks_test.go:311: check "state store" is absent from []
--- FAIL: TestRunLinkedHiveAncestor (0.00s)
=== RUN   TestRunHomeWithoutMarker
    checks_test.go:325: check "home" is absent from []
--- FAIL: TestRunHomeWithoutMarker (0.00s)
=== RUN   TestRunStateStoreMakesNoRecord
    checks_test.go:336: check "state store" is absent from []
--- FAIL: TestRunStateStoreMakesNoRecord (0.00s)
=== RUN   TestRender
    checks_test.go:358: Render() = "", want "ok herdr socket: /tmp/s.sock protocol 22\nwarn gh: not on PATH\n"
--- FAIL: TestRender (0.00s)
=== RUN   TestResultJSONFields
--- PASS: TestResultJSONFields (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/doctor	0.005s
FAIL

exit 1
```

### First fix historical red

Source: `6b063bb`. Capture: `fix01-red.txt`.

```text
$ go test -race -count=1 ./internal/lib/herdr ./internal/lib/run ./internal/doctor -v
=== RUN   TestCallSuccessEnvelope
--- PASS: TestCallSuccessEnvelope (0.00s)
=== RUN   TestCallErrorEnvelope
--- PASS: TestCallErrorEnvelope (0.00s)
=== RUN   TestCallRejectsBadReplies
=== RUN   TestCallRejectsBadReplies/another_id
=== RUN   TestCallRejectsBadReplies/both_result_and_error
=== RUN   TestCallRejectsBadReplies/neither_result_nor_error
--- PASS: TestCallRejectsBadReplies (0.00s)
    --- PASS: TestCallRejectsBadReplies/another_id (0.00s)
    --- PASS: TestCallRejectsBadReplies/both_result_and_error (0.00s)
    --- PASS: TestCallRejectsBadReplies/neither_result_nor_error (0.00s)
=== RUN   TestCallReplyTooLarge
=== RUN   TestCallReplyTooLarge/1048575
=== RUN   TestCallReplyTooLarge/1048576
=== RUN   TestCallReplyTooLarge/1048577
--- PASS: TestCallReplyTooLarge (0.04s)
    --- PASS: TestCallReplyTooLarge/1048575 (0.02s)
    --- PASS: TestCallReplyTooLarge/1048576 (0.01s)
    --- PASS: TestCallReplyTooLarge/1048577 (0.01s)
=== RUN   TestCallReplyTooLargeTenMegabytes
--- PASS: TestCallReplyTooLargeTenMegabytes (0.22s)
=== RUN   TestCallDeadlineWithSilentServer
--- PASS: TestCallDeadlineWithSilentServer (0.20s)
=== RUN   TestCallSlowReply
--- PASS: TestCallSlowReply (2.50s)
=== RUN   TestCallClosedReply
--- PASS: TestCallClosedReply (0.00s)
=== RUN   TestCallWithoutSocketPath
--- PASS: TestCallWithoutSocketPath (0.00s)
=== RUN   TestCallSocketPathTypes
=== RUN   TestCallSocketPathTypes/missing
=== RUN   TestCallSocketPathTypes/directory
=== RUN   TestCallSocketPathTypes/control_character
--- PASS: TestCallSocketPathTypes (0.00s)
    --- PASS: TestCallSocketPathTypes/missing (0.00s)
    --- PASS: TestCallSocketPathTypes/directory (0.00s)
    --- PASS: TestCallSocketPathTypes/control_character (0.00s)
=== RUN   TestCallRequestSize
=== RUN   TestCallRequestSize/1048575
=== RUN   TestCallRequestSize/1048576
=== RUN   TestCallRequestSize/1048577
--- PASS: TestCallRequestSize (0.22s)
    --- PASS: TestCallRequestSize/1048575 (0.10s)
    --- PASS: TestCallRequestSize/1048576 (0.09s)
    --- PASS: TestCallRequestSize/1048577 (0.03s)
=== RUN   TestRequestIDsShareOnePrefix
--- PASS: TestRequestIDsShareOnePrefix (0.00s)
=== RUN   TestHelperProcessRequestID
[88 lines removed. Complete raw output: /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fix01-red.txt]
=== RUN   TestProtocolGateRunsPingOnce
--- PASS: TestProtocolGateRunsPingOnce (0.00s)
=== RUN   TestProtocolGateFailureBlocksLaterCalls
--- PASS: TestProtocolGateFailureBlocksLaterCalls (0.00s)
=== RUN   TestProtocolGateIsPerSocketPath
--- PASS: TestProtocolGateIsPerSocketPath (0.00s)
=== RUN   TestCallWithMethodPingDoesNotGate
--- PASS: TestCallWithMethodPingDoesNotGate (0.00s)
=== RUN   TestProtocolGateWaiterHonorsOwnDeadline
    protocol_test.go:149: second call took 4.998963884s, want at most 200ms
--- FAIL: TestProtocolGateWaiterHonorsOwnDeadline (5.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/herdr	10.326s
=== RUN   TestCommandTimeout
--- PASS: TestCommandTimeout (1.00s)
=== RUN   TestCommandTimeoutKillsChildThatHoldsPipes
--- PASS: TestCommandTimeoutKillsChildThatHoldsPipes (1.00s)
=== RUN   TestCommandOutputCap
=== RUN   TestCommandOutputCap/1048575
=== RUN   TestCommandOutputCap/1048576
=== RUN   TestCommandOutputCap/1048577
=== RUN   TestCommandOutputCap/2097152
--- PASS: TestCommandOutputCap (0.02s)
    --- PASS: TestCommandOutputCap/1048575 (0.01s)
    --- PASS: TestCommandOutputCap/1048576 (0.01s)
    --- PASS: TestCommandOutputCap/1048577 (0.01s)
    --- PASS: TestCommandOutputCap/2097152 (0.01s)
=== RUN   TestCommandStderrCap
--- PASS: TestCommandStderrCap (0.01s)
=== RUN   TestCommandExitCodeIsNotAnError
--- PASS: TestCommandExitCodeIsNotAnError (0.00s)
=== RUN   TestCommandMissingBinary
--- PASS: TestCommandMissingBinary (0.00s)
=== RUN   TestCommandReadsEndOfFileFromStdin
--- PASS: TestCommandReadsEndOfFileFromStdin (0.00s)
=== RUN   TestCommandUsesDir
--- PASS: TestCommandUsesDir (0.00s)
=== RUN   TestCommandLogsStartAndEnd
--- PASS: TestCommandLogsStartAndEnd (0.00s)
=== RUN   TestCommandDeadlineWithInheritedPipeChild
    run_test.go:268: Command() error = <nil>, want ErrTimeout
--- FAIL: TestCommandDeadlineWithInheritedPipeChild (4.00s)
=== RUN   TestCommandCancellationWithInheritedPipeChild
    run_test.go:287: Command() error = <nil>, want context.Canceled
--- FAIL: TestCommandCancellationWithInheritedPipeChild (4.00s)
=== RUN   TestCommandErrorStaysOneLine
=== RUN   TestCommandErrorStaysOneLine/name_newline
    run_test.go:325: error = "start no\nsuch: exec: \"no\\nsuch\": executable file not found in $PATH", want one physical line
=== RUN   TestCommandErrorStaysOneLine/name_return
    run_test.go:325: error = "start no\rsuch: exec: \"no\\rsuch\": executable file not found in $PATH", want one physical line
=== RUN   TestCommandErrorStaysOneLine/dir_newline
=== RUN   TestCommandErrorStaysOneLine/dir_return
--- FAIL: TestCommandErrorStaysOneLine (0.00s)
    --- FAIL: TestCommandErrorStaysOneLine/name_newline (0.00s)
    --- FAIL: TestCommandErrorStaysOneLine/name_return (0.00s)
    --- PASS: TestCommandErrorStaysOneLine/dir_newline (0.00s)
    --- PASS: TestCommandErrorStaysOneLine/dir_return (0.00s)
=== RUN   TestCommandCancellationLogsReturnedError
    run_test.go:363: logged error = "process timeout", want the returned error "context canceled"
--- FAIL: TestCommandCancellationLogsReturnedError (0.10s)
=== RUN   TestCommandTimeoutLogsReturnedError
--- PASS: TestCommandTimeoutLogsReturnedError (0.10s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/run	10.249s
=== RUN   TestRunChecksWithFakeHerdr
--- PASS: TestRunChecksWithFakeHerdr (0.00s)
=== RUN   TestRunWithoutHerdrSocket
--- PASS: TestRunWithoutHerdrSocket (0.00s)
=== RUN   TestRunWithMissingSocket
--- PASS: TestRunWithMissingSocket (0.00s)
=== RUN   TestRunPaneWarnWithoutHerdrEnv
--- PASS: TestRunPaneWarnWithoutHerdrEnv (0.00s)
=== RUN   TestRunGHAbsentWarn
--- PASS: TestRunGHAbsentWarn (0.00s)
=== RUN   TestRunHeldLock
--- PASS: TestRunHeldLock (2.00s)
=== RUN   TestRunLinkedRecord
--- PASS: TestRunLinkedRecord (0.00s)
=== RUN   TestRunLinkedCollection
--- PASS: TestRunLinkedCollection (0.00s)
=== RUN   TestRunLinkedStateRoot
--- PASS: TestRunLinkedStateRoot (0.00s)
=== RUN   TestRunLinkedHiveAncestor
--- PASS: TestRunLinkedHiveAncestor (0.00s)
=== RUN   TestRunHomeWithoutMarker
--- PASS: TestRunHomeWithoutMarker (0.00s)
=== RUN   TestRunStateStoreMakesNoRecord
--- PASS: TestRunStateStoreMakesNoRecord (0.00s)
=== RUN   TestRender
--- PASS: TestRender (0.00s)
=== RUN   TestResultJSONFields
--- PASS: TestResultJSONFields (0.00s)
=== RUN   TestRunLogsProcessEvents
    checks_test.go:446: process events = 0 starts and 0 ends, want 7 and 7
--- FAIL: TestRunLogsProcessEvents (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/doctor	2.023s
FAIL

exit 1
```

### Second fix historical red

Source: `0007b1cc`. Capture: `fix02-red.txt`.

```text
$ go test -race -count=1 ./internal/lib/herdr -run TestNullResultPreservesCallerOnDecodeError|TestPingLogsResultErrors -v
=== RUN   TestNullResultPreservesCallerOnDecodeError
    client_test.go:658: destination = {Value:0 Valid:false}, want {Value:7 Valid:true} unchanged
--- FAIL: TestNullResultPreservesCallerOnDecodeError (0.00s)
=== RUN   TestPingLogsResultErrors
=== RUN   TestPingLogsResultErrors/protocol_mismatch
    protocol_test.go:91: events = [map[body:{"id":"1a7528b9-3","method":"ping","params":{}} component:herdr duration_ms:0 level:DEBUG method:ping msg:herdr request request_id:1a7528b9-3 time:2026-10-18T05:21:55.012445214-04:00] map[body:{"id":"1a7528b9-3","result":{"type":"pong","protocol":21}} component:herdr duration_ms:0 level:DEBUG method:ping msg:herdr reply request_id:1a7528b9-3 time:2026-10-18T05:21:55.01260995-04:00]], want a ping error
=== RUN   TestPingLogsResultErrors/wrong_type
    protocol_test.go:91: events = [map[body:{"id":"1a7528b9-4","method":"ping","params":{}} component:herdr duration_ms:0 level:DEBUG method:ping msg:herdr request request_id:1a7528b9-4 time:2026-10-18T05:21:55.013164712-04:00] map[body:{"id":"1a7528b9-4","result":{"type":"not-pong","protocol":22}} component:herdr duration_ms:0 level:DEBUG method:ping msg:herdr reply request_id:1a7528b9-4 time:2026-10-18T05:21:55.013280717-04:00]], want a ping error
--- FAIL: TestPingLogsResultErrors (0.00s)
    --- FAIL: TestPingLogsResultErrors/protocol_mismatch (0.00s)
    --- FAIL: TestPingLogsResultErrors/wrong_type (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/herdr	0.006s
FAIL

exit 1
```

### Third fix historical red

Source: `daf2fea8`. Capture: `fix03-red.txt`.

```text
$ go test -race -count=1 ./internal/lib/herdr -run TestNullResultPreservesReferencedStateOnDecodeError|TestReferencedResultAssignedOnSuccess -v
=== RUN   TestNullResultPreservesReferencedStateOnDecodeError
    client_test.go:719: value = 0xc00011c980 count = 0, want original pointer and 7
--- FAIL: TestNullResultPreservesReferencedStateOnDecodeError (0.00s)
=== RUN   TestReferencedResultAssignedOnSuccess
--- PASS: TestReferencedResultAssignedOnSuccess (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/herdr	0.005s
FAIL

exit 1
```

### Literal frame mutation red

Source: `46b88c08 with isolated 2097152 caps`. Capture: `frame-mutation-red.txt`.

```text
$ go test -race -count=1 ./internal/lib/herdr -run TestPrimaryFrameBounds -v
=== RUN   TestPrimaryFrameBounds
=== RUN   TestPrimaryFrameBounds/request1048575
=== RUN   TestPrimaryFrameBounds/reply1048575
=== RUN   TestPrimaryFrameBounds/request1048576
=== RUN   TestPrimaryFrameBounds/reply1048576
=== RUN   TestPrimaryFrameBounds/request1048577
    verify02_primary_test.go:217: error=<nil>
    verify02_primary_test.go:220: accepted=2, want zero including gate
=== RUN   TestPrimaryFrameBounds/reply1048577
    verify02_primary_test.go:245: error=<nil>
--- FAIL: TestPrimaryFrameBounds (0.30s)
    --- PASS: TestPrimaryFrameBounds/request1048575 (0.08s)
    --- PASS: TestPrimaryFrameBounds/reply1048575 (0.02s)
    --- PASS: TestPrimaryFrameBounds/request1048576 (0.08s)
    --- PASS: TestPrimaryFrameBounds/reply1048576 (0.02s)
    --- FAIL: TestPrimaryFrameBounds/request1048577 (0.09s)
    --- FAIL: TestPrimaryFrameBounds/reply1048577 (0.02s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/herdr	0.302s
FAIL

exit 1
```

### Literal frame restoration green

Source: `46b88c08`. Capture: `frame-mutation-green.txt`.

```text
$ go test -race -count=1 ./internal/lib/herdr -run TestPrimaryFrameBounds -v
=== RUN   TestPrimaryFrameBounds
=== RUN   TestPrimaryFrameBounds/request1048575
=== RUN   TestPrimaryFrameBounds/reply1048575
=== RUN   TestPrimaryFrameBounds/request1048576
=== RUN   TestPrimaryFrameBounds/reply1048576
=== RUN   TestPrimaryFrameBounds/request1048577
=== RUN   TestPrimaryFrameBounds/reply1048577
--- PASS: TestPrimaryFrameBounds (0.25s)
    --- PASS: TestPrimaryFrameBounds/request1048575 (0.08s)
    --- PASS: TestPrimaryFrameBounds/reply1048575 (0.01s)
    --- PASS: TestPrimaryFrameBounds/request1048576 (0.08s)
    --- PASS: TestPrimaryFrameBounds/reply1048576 (0.02s)
    --- PASS: TestPrimaryFrameBounds/request1048577 (0.05s)
    --- PASS: TestPrimaryFrameBounds/reply1048577 (0.01s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	1.256s

exit 0
```

### Literal stream mutation red

Source: `46b88c08 with isolated 2097152 cap`. Capture: `stream-mutation-red.txt`.

```text
$ go test -race -count=1 ./internal/lib/run -run TestSecondaryRunnerExactCaps -v
=== RUN   TestSecondaryRunnerExactCaps
    verify02_runner_test.go:36: n=1048577 stdout=1048577 stderr=1048577 truncated=false
--- FAIL: TestSecondaryRunnerExactCaps (0.04s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/run	0.044s
FAIL

exit 1
```

### Literal stream restoration green

Source: `46b88c08`. Capture: `stream-mutation-green.txt`.

```text
$ go test -race -count=1 ./internal/lib/run -run TestSecondaryRunnerExactCaps -v
=== RUN   TestSecondaryRunnerExactCaps
--- PASS: TestSecondaryRunnerExactCaps (0.04s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/run	1.045s

exit 0
```

### Literal doctor order mutation red

Source: `46b88c08 with isolated order swap`. Capture: `order-mutation-red.txt`.

```text
$ go test -race -count=1 ./internal/doctor -run TestSecondaryDoctorOrderStatusesPreservation -v
=== RUN   TestSecondaryDoctorOrderStatusesPreservation
    verify02_doctor_test.go:28: check 0={Name:herdr pane Status:ok Required:false Detail:w1:p1???bad}
    verify02_doctor_test.go:28: check 1={Name:herdr socket Status:ok Required:true Detail:/proc/self/fd/6/s protocol 22}
    verify02_doctor_test.go:28: literal order, all statuses, real git, fake dependencies, four unchanged domain records
--- FAIL: TestSecondaryDoctorOrderStatusesPreservation (0.01s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/doctor	0.008s
FAIL

exit 1
```

### Literal doctor order restoration green

Source: `46b88c08`. Capture: `order-mutation-green.txt`.

```text
$ go test -race -count=1 ./internal/doctor -run TestSecondaryDoctorOrderStatusesPreservation -v
=== RUN   TestSecondaryDoctorOrderStatusesPreservation
    verify02_doctor_test.go:28: literal order, all statuses, real git, fake dependencies, four unchanged domain records
--- PASS: TestSecondaryDoctorOrderStatusesPreservation (0.01s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/doctor	1.009s

exit 0
```

### Actual protocol-21 CLI assertion-red

Source: `75d0d2ef`. Capture: `cli-protocol-prior-final.txt`.

```text
{"outcome":"failed","phase":"complete","effects":[],"error":null,"result":[{"name":"herdr socket","status":"fail","required":true,"detail":"herdr protocol 21, hive needs 22"},{"name":"herdr pane","status":"ok","required":false,"detail":"wA9:p1"},{"name":"home","status":"ok","required":true,"detail":"/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-protocol-lancuc7v"},{"name":"state store","status":"ok","required":true,"detail":"0 records"},{"name":"git","status":"ok","required":true,"detail":"git version 2.56.0"},{"name":"supervisor","status":"warn","required":false,"detail":"no supervisor claim"},{"name":"harness claude","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness codex","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness opencode","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness pi","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness cursor-agent","status":"warn","required":false,"detail":"not on PATH"},{"name":"gh","status":"warn","required":false,"detail":"not on PATH"}]}
exit 1
ping events:
{"body": "{\"id\":\"764d9d4d-1\",\"method\":\"ping\",\"params\":{}}", "component": "doctor", "duration_ms": 0, "level": "DEBUG", "method": "ping", "msg": "herdr request", "request_id": "764d9d4d-1", "time": "2026-10-18T05:24:47.906229702-04:00"}
{"body": "{\"id\": \"764d9d4d-1\", \"result\": {\"type\": \"pong\", \"protocol\": 21}}", "component": "doctor", "duration_ms": 0, "level": "DEBUG", "method": "ping", "msg": "herdr reply", "request_id": "764d9d4d-1", "time": "2026-10-18T05:24:47.906316574-04:00"}
actual CLI error fields: 1
removed own fixture /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-protocol-lancuc7v
Traceback (most recent call last):
  File "/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/cli-protocol-correlated.py", line 21, in <module>
    assert len(errs)==1 and errs[0]['request_id']==req[0]['request_id'] and errs[0]['duration_ms']>=0 and len(errs[0]['error'])<=4096 and 'body' not in errs[0], 'missing matched actual CLI error event'
           ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
AssertionError: missing matched actual CLI error event
```

### Actual protocol-21 CLI assertion-green

Source: `46b88c08`. Capture: `cli-protocol-worker-final.txt`.

```text
{"outcome":"failed","phase":"complete","effects":[],"error":null,"result":[{"name":"herdr socket","status":"fail","required":true,"detail":"herdr protocol 21, hive needs 22"},{"name":"herdr pane","status":"ok","required":false,"detail":"wA9:p1"},{"name":"home","status":"ok","required":true,"detail":"/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-protocol-mjkrvbb0"},{"name":"state store","status":"ok","required":true,"detail":"0 records"},{"name":"git","status":"ok","required":true,"detail":"git version 2.56.0"},{"name":"supervisor","status":"warn","required":false,"detail":"no supervisor claim"},{"name":"harness claude","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness codex","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness opencode","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness pi","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness cursor-agent","status":"warn","required":false,"detail":"not on PATH"},{"name":"gh","status":"warn","required":false,"detail":"not on PATH"}]}
exit 1
ping events:
{"body": "{\"id\":\"fa87b2b7-1\",\"method\":\"ping\",\"params\":{}}", "component": "doctor", "duration_ms": 0, "level": "DEBUG", "method": "ping", "msg": "herdr request", "request_id": "fa87b2b7-1", "time": "2026-10-18T05:24:47.87638215-04:00"}
{"body": "{\"id\": \"fa87b2b7-1\", \"result\": {\"type\": \"pong\", \"protocol\": 21}}", "component": "doctor", "duration_ms": 0, "level": "DEBUG", "method": "ping", "msg": "herdr reply", "request_id": "fa87b2b7-1", "time": "2026-10-18T05:24:47.876465636-04:00"}
{"component": "doctor", "duration_ms": 0, "error": "herdr protocol 21, hive needs 22", "level": "DEBUG", "method": "ping", "msg": "herdr error", "request_id": "fa87b2b7-1", "time": "2026-10-18T05:24:47.87651127-04:00"}
actual CLI error fields: 2
removed own fixture /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-protocol-mjkrvbb0
```

### V5 secondary CLI matrix

Source: `46b88c08`. Capture: `secondary-cli-final.txt`.

```text
healthy-required-text exit 0 stdout_lines 12 stderr_lines 0
healthy-required-json exit 0 stdout_lines 1 stderr_lines 0
socket-unset exit 1 stdout_lines 1 stderr_lines 0
bad-socket-27 exit 1 stdout_lines 12 stderr_lines 0
bad-socket-12 exit 1 stdout_lines 12 stderr_lines 0
bad-socket-10000 exit 1 stdout_lines 12 stderr_lines 0
malformed-record-newline-name exit 1 stdout_lines 12 stderr_lines 0
log event counts {'command start': 7, 'command end': 7, 'herdr request': 3, 'herdr reply': 3, 'process start': 56, 'process end': 56}
domain records preserved. CLI assertions PASS. No actual gh auth or harness execution.
own CLI home removed True
```

### V4 and V5 full actual CLI supplement

Source: `46b88c08`. Capture: `cli-final.txt`.

```text
HERDR_ENV present
HERDR_PANE_ID present
HERDR_SOCKET_PATH present
HERDR_WORKSPACE_ID present
HERDR_TAB_ID present
HERDR_TERMINAL_ID absent
## V2 live doctor JSON
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v2-zbv0drvd --json doctor
{"outcome":"ok","phase":"complete","effects":[],"error":null,"result":[{"name":"herdr socket","status":"ok","required":true,"detail":"/home/penguin/.config/herdr/herdr.sock protocol 22"},{"name":"herdr pane","status":"ok","required":false,"detail":"wA9:p1"},{"name":"home","status":"ok","required":true,"detail":"/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v2-zbv0drvd"},{"name":"state store","status":"ok","required":true,"detail":"0 records"},{"name":"git","status":"ok","required":true,"detail":"git version 2.56.0"},{"name":"supervisor","status":"warn","required":false,"detail":"no supervisor claim"},{"name":"harness claude","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness codex","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness opencode","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness pi","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness cursor-agent","status":"warn","required":false,"detail":"not on PATH"},{"name":"gh","status":"warn","required":false,"detail":"not on PATH"}]}
exit 0 elapsed 0.0055
## V2 ping log
{"body": "{\"id\":\"0ef5ca07-1\",\"method\":\"ping\",\"params\":{}}", "component": "doctor", "duration_ms": 0, "level": "DEBUG", "method": "ping", "msg": "herdr request", "request_id": "0ef5ca07-1", "time": "2026-10-18T05:24:47.70488576-04:00"}
{"body": "{\"id\":\"0ef5ca07-1\",\"result\":{\"type\":\"pong\",\"version\":\"0.9.5\",\"protocol\":22,\"capabilities\":{\"live_handoff\":true,\"detached_server_daemon\":true,\"endpoint_protocol_generation\":1,\"surface_interest\":true,\"health_check\":true,\"ssh_agent_registration\":true}}}", "component": "doctor", "duration_ms": 0, "level": "DEBUG", "method": "ping", "msg": "herdr reply", "request_id": "0ef5ca07-1", "time": "2026-10-18T05:24:47.70495589-04:00"}
## V5 live text
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v5-49u5u9ts doctor
ok herdr socket: /home/penguin/.config/herdr/herdr.sock protocol 22
ok herdr pane: wA9:p1
ok home: /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v5-49u5u9ts
ok state store: 0 records
ok git: git version 2.56.0
warn supervisor: no supervisor claim
warn harness claude: not on PATH
warn harness codex: not on PATH
warn harness opencode: not on PATH
warn harness pi: not on PATH
warn harness cursor-agent: not on PATH
warn gh: not on PATH
exit 0 elapsed 0.005
## V5 missing socket
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v5-49u5u9ts doctor
fail herdr socket: herdr socket: /tmp/none.sock: no such file
ok herdr pane: wA9:p1
ok home: /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v5-49u5u9ts
ok state store: 0 records
ok git: git version 2.56.0
warn supervisor: no supervisor claim
warn harness claude: not on PATH
warn harness codex: not on PATH
warn harness opencode: not on PATH
warn harness pi: not on PATH
warn harness cursor-agent: not on PATH
warn gh: not on PATH
exit 1 elapsed 0.0048
## V5 unset socket
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v5-49u5u9ts doctor
fail herdr socket: HERDR_SOCKET_PATH is not set
ok herdr pane: wA9:p1
ok home: /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v5-49u5u9ts
ok state store: 0 records
ok git: git version 2.56.0
warn supervisor: no supervisor claim
warn harness claude: not on PATH
warn harness codex: not on PATH
warn harness opencode: not on PATH
warn harness pi: not on PATH
warn harness cursor-agent: not on PATH
warn gh: not on PATH
exit 1 elapsed 0.0047
## V5 version
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v5-49u5u9ts --json version
{"outcome":"ok","phase":"complete","effects":[],"error":null,"result":{"version":"dev","commit":"46b88c087f6d","modified":false,"go":"go1.27.2-X:nodwarf5"}}
exit 0 elapsed 0.0011
## V5 failed JSON
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v5-49u5u9ts --json doctor
{"outcome":"failed","phase":"complete","effects":[],"error":null,"result":[{"name":"herdr socket","status":"fail","required":true,"detail":"HERDR_SOCKET_PATH is not set"},{"name":"herdr pane","status":"ok","required":false,"detail":"wA9:p1"},{"name":"home","status":"ok","required":true,"detail":"/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v5-49u5u9ts"},{"name":"state store","status":"ok","required":true,"detail":"0 records"},{"name":"git","status":"ok","required":true,"detail":"git version 2.56.0"},{"name":"supervisor","status":"warn","required":false,"detail":"no supervisor claim"},{"name":"harness claude","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness codex","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness opencode","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness pi","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness cursor-agent","status":"warn","required":false,"detail":"not on PATH"},{"name":"gh","status":"warn","required":false,"detail":"not on PATH"}]}
exit 1 elapsed 0.0048
## home-free version
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /t05v04-absent-home version
hive dev (commit 46b88c087f6d, go1.27.2-X:nodwarf5)
exit 0 elapsed 0.0011
## home-free version
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /t05v04-absent-home version --json
{"outcome":"ok","phase":"complete","effects":[],"error":null,"result":{"version":"dev","commit":"46b88c087f6d","modified":false,"go":"go1.27.2-X:nodwarf5"}}
exit 0 elapsed 0.0011
## seven actual CLI process attempts
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-logging-_719fa07 doctor
fail herdr socket: HERDR_SOCKET_PATH is not set
ok herdr pane: wA9:p1
ok home: /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-logging-_719fa07
ok state store: 0 records
ok git: git version 2.56.0
warn supervisor: no supervisor claim
warn harness claude: not on PATH
warn harness codex: not on PATH
warn harness opencode: not on PATH
warn harness pi: not on PATH
warn harness cursor-agent: not on PATH
warn gh: not on PATH
exit 1 elapsed 0.0019
process starts 7 ends 7 failed starts 6
{"command": "git --version", "component": "doctor", "duration_ms": 0, "exit_code": 0, "level": "DEBUG", "msg": "process end", "time": "2026-10-18T05:24:47.733895855-04:00"}
{"command": "sh -c command -v \"$1\" sh claude", "component": "doctor", "duration_ms": 0, "error": "start sh: exec: \"sh\": executable file not found in $PATH", "exit_code": -1, "level": "DEBUG", "msg": "process end", "time": "2026-10-18T05:24:47.733970413-04:00"}
{"command": "sh -c command -v \"$1\" sh codex", "component": "doctor", "duration_ms": 0, "error": "start sh: exec: \"sh\": executable file not found in $PATH", "exit_code": -1, "level": "DEBUG", "msg": "process end", "time": "2026-10-18T05:24:47.733988206-04:00"}
{"command": "sh -c command -v \"$1\" sh opencode", "component": "doctor", "duration_ms": 0, "error": "start sh: exec: \"sh\": executable file not found in $PATH", "exit_code": -1, "level": "DEBUG", "msg": "process end", "time": "2026-10-18T05:24:47.734003465-04:00"}
{"command": "sh -c command -v \"$1\" sh pi", "component": "doctor", "duration_ms": 0, "error": "start sh: exec: \"sh\": executable file not found in $PATH", "exit_code": -1, "level": "DEBUG", "msg": "process end", "time": "2026-10-18T05:24:47.734025826-04:00"}
{"command": "sh -c command -v \"$1\" sh cursor-agent", "component": "doctor", "duration_ms": 0, "error": "start sh: exec: \"sh\": executable file not found in $PATH", "exit_code": -1, "level": "DEBUG", "msg": "process end", "time": "2026-10-18T05:24:47.734040664-04:00"}
{"command": "gh --version", "component": "doctor", "duration_ms": 0, "error": "start gh: exec: \"gh\": executable file not found in $PATH", "exit_code": -1, "level": "DEBUG", "msg": "process end", "time": "2026-10-18T05:24:47.734056133-04:00"}
## S05-D01 tasks fifo
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-och2vinu --json doctor
{"outcome":"failed","phase":"complete","effects":[],"error":null,"result":[{"name":"herdr socket","status":"fail","required":true,"detail":"HERDR_SOCKET_PATH is not set"},{"name":"herdr pane","status":"ok","required":false,"detail":"wA9:p1"},{"name":"home","status":"ok","required":true,"detail":"/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-och2vinu"},{"name":"state store","status":"fail","required":true,"detail":"store path invalid: /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-och2vinu/.hive/state/tasks"},{"name":"git","status":"ok","required":true,"detail":"git version 2.56.0"},{"name":"supervisor","status":"warn","required":false,"detail":"no supervisor claim"},{"name":"harness claude","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness codex","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness opencode","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness pi","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness cursor-agent","status":"warn","required":false,"detail":"not on PATH"},{"name":"gh","status":"warn","required":false,"detail":"not on PATH"}]}
exit 1 elapsed 0.0049
## S05-D01 tasks socket
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-2u11c9fu --json doctor
{"outcome":"failed","phase":"complete","effects":[],"error":null,"result":[{"name":"herdr socket","status":"fail","required":true,"detail":"HERDR_SOCKET_PATH is not set"},{"name":"herdr pane","status":"ok","required":false,"detail":"wA9:p1"},{"name":"home","status":"ok","required":true,"detail":"/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-2u11c9fu"},{"name":"state store","status":"fail","required":true,"detail":"store path invalid: /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-2u11c9fu/.hive/state/tasks"},{"name":"git","status":"ok","required":true,"detail":"git version 2.56.0"},{"name":"supervisor","status":"warn","required":false,"detail":"no supervisor claim"},{"name":"harness claude","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness codex","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness opencode","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness pi","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness cursor-agent","status":"warn","required":false,"detail":"not on PATH"},{"name":"gh","status":"warn","required":false,"detail":"not on PATH"}]}
exit 1 elapsed 0.0048
## S05-D01 inbox/coordinator/tasks fifo
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-sp3v7jmz --json doctor
{"outcome":"failed","phase":"complete","effects":[],"error":null,"result":[{"name":"herdr socket","status":"fail","required":true,"detail":"HERDR_SOCKET_PATH is not set"},{"name":"herdr pane","status":"ok","required":false,"detail":"wA9:p1"},{"name":"home","status":"ok","required":true,"detail":"/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-sp3v7jmz"},{"name":"state store","status":"fail","required":true,"detail":"store path invalid: /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-sp3v7jmz/.hive/state/inbox/coordinator/tasks"},{"name":"git","status":"ok","required":true,"detail":"git version 2.56.0"},{"name":"supervisor","status":"warn","required":false,"detail":"no supervisor claim"},{"name":"harness claude","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness codex","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness opencode","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness pi","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness cursor-agent","status":"warn","required":false,"detail":"not on PATH"},{"name":"gh","status":"warn","required":false,"detail":"not on PATH"}]}
exit 1 elapsed 0.0048
## S05-D01 inbox/coordinator/tasks socket
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-khxuagd_ --json doctor
{"outcome":"failed","phase":"complete","effects":[],"error":null,"result":[{"name":"herdr socket","status":"fail","required":true,"detail":"HERDR_SOCKET_PATH is not set"},{"name":"herdr pane","status":"ok","required":false,"detail":"wA9:p1"},{"name":"home","status":"ok","required":true,"detail":"/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-khxuagd_"},{"name":"state store","status":"fail","required":true,"detail":"store path invalid: /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-khxuagd_/.hive/state/inbox/coordinator/tasks"},{"name":"git","status":"ok","required":true,"detail":"git version 2.56.0"},{"name":"supervisor","status":"warn","required":false,"detail":"no supervisor claim"},{"name":"harness claude","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness codex","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness opencode","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness pi","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness cursor-agent","status":"warn","required":false,"detail":"not on PATH"},{"name":"gh","status":"warn","required":false,"detail":"not on PATH"}]}
exit 1 elapsed 0.0046
## S05-D01 .tasks.json.tmp-123 fifo
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-jqu7h9wb --json doctor
{"outcome":"failed","phase":"complete","effects":[],"error":null,"result":[{"name":"herdr socket","status":"fail","required":true,"detail":"HERDR_SOCKET_PATH is not set"},{"name":"herdr pane","status":"ok","required":false,"detail":"wA9:p1"},{"name":"home","status":"ok","required":true,"detail":"/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-jqu7h9wb"},{"name":"state store","status":"fail","required":true,"detail":"store path invalid: /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-jqu7h9wb/.hive/state/.tasks.json.tmp-123"},{"name":"git","status":"ok","required":true,"detail":"git version 2.56.0"},{"name":"supervisor","status":"warn","required":false,"detail":"no supervisor claim"},{"name":"harness claude","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness codex","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness opencode","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness pi","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness cursor-agent","status":"warn","required":false,"detail":"not on PATH"},{"name":"gh","status":"warn","required":false,"detail":"not on PATH"}]}
exit 1 elapsed 0.005
## S05-D01 .tasks.json.tmp-123 socket
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-ezrcpcls --json doctor
{"outcome":"failed","phase":"complete","effects":[],"error":null,"result":[{"name":"herdr socket","status":"fail","required":true,"detail":"HERDR_SOCKET_PATH is not set"},{"name":"herdr pane","status":"ok","required":false,"detail":"wA9:p1"},{"name":"home","status":"ok","required":true,"detail":"/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-ezrcpcls"},{"name":"state store","status":"fail","required":true,"detail":"store path invalid: /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-ezrcpcls/.hive/state/.tasks.json.tmp-123"},{"name":"git","status":"ok","required":true,"detail":"git version 2.56.0"},{"name":"supervisor","status":"warn","required":false,"detail":"no supervisor claim"},{"name":"harness claude","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness codex","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness opencode","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness pi","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness cursor-agent","status":"warn","required":false,"detail":"not on PATH"},{"name":"gh","status":"warn","required":false,"detail":"not on PATH"}]}
exit 1 elapsed 0.0046
## S05-D01 tasks/one.json fifo
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-pzfqatqr --json doctor
{"outcome":"failed","phase":"complete","effects":[],"error":null,"result":[{"name":"herdr socket","status":"fail","required":true,"detail":"HERDR_SOCKET_PATH is not set"},{"name":"herdr pane","status":"ok","required":false,"detail":"wA9:p1"},{"name":"home","status":"ok","required":true,"detail":"/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-pzfqatqr"},{"name":"state store","status":"fail","required":true,"detail":"record invalid: /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-pzfqatqr/.hive/state/tasks/one.json"},{"name":"git","status":"ok","required":true,"detail":"git version 2.56.0"},{"name":"supervisor","status":"warn","required":false,"detail":"no supervisor claim"},{"name":"harness claude","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness codex","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness opencode","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness pi","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness cursor-agent","status":"warn","required":false,"detail":"not on PATH"},{"name":"gh","status":"warn","required":false,"detail":"not on PATH"}]}
exit 1 elapsed 0.0046
## S05-D01 tasks/one.json socket
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-2l09ju7n --json doctor
{"outcome":"failed","phase":"complete","effects":[],"error":null,"result":[{"name":"herdr socket","status":"fail","required":true,"detail":"HERDR_SOCKET_PATH is not set"},{"name":"herdr pane","status":"ok","required":false,"detail":"wA9:p1"},{"name":"home","status":"ok","required":true,"detail":"/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-2l09ju7n"},{"name":"state store","status":"fail","required":true,"detail":"record invalid: /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-2l09ju7n/.hive/state/tasks/one.json"},{"name":"git","status":"ok","required":true,"detail":"git version 2.56.0"},{"name":"supervisor","status":"warn","required":false,"detail":"no supervisor claim"},{"name":"harness claude","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness codex","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness opencode","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness pi","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness cursor-agent","status":"warn","required":false,"detail":"not on PATH"},{"name":"gh","status":"warn","required":false,"detail":"not on PATH"}]}
exit 1 elapsed 0.0047
## actual CLI record preservation
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-preserve-qd9n0x92 --json doctor
{"outcome":"ok","phase":"complete","effects":[],"error":null,"result":[{"name":"herdr socket","status":"ok","required":true,"detail":"/home/penguin/.config/herdr/herdr.sock protocol 22"},{"name":"herdr pane","status":"ok","required":false,"detail":"wA9:p1"},{"name":"home","status":"ok","required":true,"detail":"/home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-preserve-qd9n0x92"},{"name":"state store","status":"ok","required":true,"detail":"3 records"},{"name":"git","status":"ok","required":true,"detail":"git version 2.56.0"},{"name":"supervisor","status":"warn","required":false,"detail":"no supervisor claim"},{"name":"harness claude","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness codex","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness opencode","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness pi","status":"warn","required":false,"detail":"not on PATH"},{"name":"harness cursor-agent","status":"warn","required":false,"detail":"not on PATH"},{"name":"gh","status":"warn","required":false,"detail":"not on PATH"}]}
exit 0 elapsed 0.0048
## root help no command loss
$ /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive --home /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-preserve-qd9n0x92 --help
hive

Usage:
  hive [flags]
  hive [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  doctor      Check the environment
  help        Help about any command
  logs        read the daily log files
  project     manage registered projects
  task        manage task records
  version     Print the hive version

Flags:
  -h, --help          help for hive
      --home string   hive home
      --json          write JSON output

Use "hive [command] --help" for more information about a command.
exit 0 elapsed 0.001
CLI assertions PASS. No harness or real gh execution. No credential access.
removed own fixture /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-deps-wboffv8o
removed own fixture /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v2-zbv0drvd
removed own fixture /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-v5-49u5u9ts
removed own fixture /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-version-free-sdeglu0k
removed own fixture /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-logging-_719fa07
removed own fixture /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-och2vinu
removed own fixture /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-2u11c9fu
removed own fixture /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-sp3v7jmz
removed own fixture /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-khxuagd_
removed own fixture /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-jqu7h9wb
removed own fixture /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-ezrcpcls
removed own fixture /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-pzfqatqr
removed own fixture /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-special-2l09ju7n
removed own fixture /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures/t05v04-preserve-qd9n0x92
```

### V6 formatting

Source: `46b88c08`. Capture: `format.txt`.

```text
commit 46b88c087f6de501cda589f017a4e3a794ab7f81 (archive identity is in run name when different)
$ bash -c gofmt -l $(git ls-files -co --exclude-standard '*.go')

exit 0
```

### V6 vet

Source: `46b88c08`. Capture: `vet.txt`.

```text
commit 46b88c087f6de501cda589f017a4e3a794ab7f81 (archive identity is in run name when different)
$ go vet ./...

exit 0
```

### V6 lint

Source: `46b88c08`. Capture: `lint.txt`.

```text
commit 46b88c087f6de501cda589f017a4e3a794ab7f81 (archive identity is in run name when different)
$ golangci-lint run ./...
0 issues.

exit 0
```

### V6 full uncached race gate

Source: `46b88c08`. Capture: `race-gate.txt`.

```text
commit 46b88c087f6de501cda589f017a4e3a794ab7f81 (archive identity is in run name when different)
$ go test -race -count=1 ./...
ok  	github.com/Harrison-Blair/hive	3.041s
ok  	github.com/Harrison-Blair/hive/cmd	1.052s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
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
ok  	github.com/Harrison-Blair/hive/cmd/version	1.020s
ok  	github.com/Harrison-Blair/hive/internal/doctor	3.041s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	6.489s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.023s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.056s
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/records	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/run	3.473s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	17.308s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.018s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.033s
ok  	github.com/Harrison-Blair/hive/internal/project	1.027s
ok  	github.com/Harrison-Blair/hive/internal/task	1.182s
ok  	github.com/Harrison-Blair/hive/internal/tui	1.021s

exit 0
```

### STE baseline check

Source: `46b88c08`. Capture: `ste.txt`.

```text
commit 46b88c087f6de501cda589f017a4e3a794ab7f81 (archive identity is in run name when different)
$ go test -run TestSTE -count=1 .
ok  	github.com/Harrison-Blair/hive	0.136s

exit 0
```

### Diff check

Source: `46b88c08`. Capture: `diff-check.txt`.

```text
commit 46b88c087f6de501cda589f017a4e3a794ab7f81 (archive identity is in run name when different)
$ git diff --check

exit 0
```

### Whole-range source and prerequisite checks

Source: `46b88c08`. Capture: `scope-check.txt`.

```text
PASS: 89 full-range paths. Imported objects exact. Store exact ccfe0020. Mechanical root union exact. All Go source exact 51e8b0a8. Old verdict exact.
Root observed 4837f1e728a2b958604734a8a3fd9e8edf4c42de

Root status A  cmd/handoff/handoff.go
A  cmd/handoff/write/write.go
A  cmd/handoff/write/write_cli_test.go
A  cmd/memory/memory.go
A  cmd/memory/write/write.go
A  cmd/memory/write/write_cli_test.go
M  cmd/root.go
A  internal/handoff/write.go
A  internal/handoff/write_test.go
A  internal/lib/scope/doc.go
A  internal/lib/scope/input.go
A  internal/lib/scope/scope.go
A  internal/lib/scope/scope_test.go
A  internal/memory/doc.go
A  internal/memory/frontmatter.go
A  internal/memory/frontmatter_test.go
A  internal/memory/index.go
A  internal/memory/index_test.go
A  internal/memory/write.go
A  internal/memory/write_process_test.go
A  internal/memory/write_test.go
A  spec/hive-v1/evidence/T07-memory-and-handoff/V1.md
A  spec/hive-v1/evidence/T07-memory-and-handoff/V2.md
A  spec/hive-v1/evidence/T07-memory-and-handoff/V3.md
A  spec/hive-v1/evidence/T07-memory-and-handoff/V4.md
A  spec/hive-v1/evidence/T07-memory-and-handoff/VERIFIED.md

// Package cmd holds the hive root command and registers its subcommands.
package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Harrison-Blair/hive/cmd/agent"
	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/cmd/inbox"
	"github.com/Harrison-Blair/hive/cmd/logs"
	"github.com/Harrison-Blair/hive/cmd/project"
	"github.com/Harrison-Blair/hive/cmd/task"
	"github.com/Harrison-Blair/hive/cmd/version"
	"github.com/Harrison-Blair/hive/internal/tui"
	"github.com/spf13/cobra"
)

// NewRootCmd builds a fresh hive command tree.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "hive",
		Short:         "hive",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return tui.Run(cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
	root.PersistentFlags().String("home", "", "hive home")
	root.PersistentFlags().Bool("json", false, "write JSON output")
	root.AddCommand(version.New(), project.New(), logs.New(), task.New(), agent.New(), inbox.New())
	return root
}

// Execute runs hive with the process arguments and standard streams.
func Execute() int {
	return executeWithStreams(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
}

// ExecuteWithArgs runs hive with the given arguments and streams.
func ExecuteWithArgs(args []string, in io.Reader, out io.Writer) error {
	if status := executeWithStreams(args, in, out, out); status != 0 {
		return fmt.Errorf("hive exited with status %d", status)
	}
	return nil
}

func executeWithStreams(args []string, in io.Reader, out, errOut io.Writer) int {
	root := NewRootCmd()
	root.SetArgs(args)
	root.SetIn(in)
	root.SetOut(out)
	root.SetErr(errOut)
	if err := root.Execute(); err != nil {
		var statusErr *common.StatusError
		if errors.As(err, &statusErr) {
			return statusErr.Status
		}
		var usage *common.UsageError
		if errors.As(err, &usage) {
			_, _ = fmt.Fprintf(errOut, "hive: %s\n", usage.Error())
			return 2
		}
		_, _ = fmt.Fprintf(errOut, "hive: %s\n", err)
		return 2
	}
	return 0
}

PASS: no internal Cobra or cmd imports.
full-range 15bef382b22379c18c0af2918a17f611716d67141501d35a86abf27ed51752fa
repair-01 c8493009241d43ac8f4454dd79cef6ee4628c7260b6434eea4aa1517c1934823
repair-02 e849d0f1f5631812cc32359cbaabd8077e9458494e2d736fa990e22cc9baeb39
repair-03 da24638823cc17913c42aef5802c33fa0c22d70dbc1c73bffeb3f144b628d667
```

### Cleanup and preservation

Source: `46b88c08`. Capture: `cleanup.txt`.

```text
PASS: 205 tracked worker paths exact before append. All worker evidence unchanged.
Old verdict bytes 41012 SHA256 a95ff2f10577577f01a62cb78caada6e42772fea436978630c78a86814d329b9
PASS: restored current all tracked paths exact.
PASS: restored mutation all tracked paths exact.
PASS: restored nested-worker all tracked paths exact.
PASS: no process cwd is below an owned scratch path. Historical fixed-duration sleepers elapsed.
removed own directory /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/copy-current
removed own directory /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/copy-prior
removed own directory /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/copy-pingprior
removed own directory /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/fixtures
removed own file /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive
removed own file /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive-secondary
removed own file /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive-pingprior
removed own directory /tmp/h5v4.s92j3yft
removed own directory /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/copy-v1-red
removed own directory /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/copy-v3-red
removed own directory /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/copy-v4-red
removed own directory /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/copy-fix01-red
removed own directory /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/copy-fix02-red
removed own directory /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/copy-fix03-red
removed own directory /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/copy-mutation
removed own directory /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/copy-recorder
removed own directory /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/copy-nested-prior
removed own directory /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/copy-nested-worker
PASS: only manifest-listed own copies, fixtures, binaries, alias, and short TMPDIR removed. Prior TMPDIR environment was not changed in the parent.
PASS: probe sources, exact synthetic logs, raw outputs, diffs, and manifests remain.
Root current eded8f9cb09208bfdc81334870e86479e746c183
Root status preserved without a write:

Local main and origin/main: 3692162633f896b8fee6197469adba91f666a00c
928cfa95e0906315fcb4a65a5cdbdef1470a0503

Direct remote main: 928cfa95e0906315fcb4a65a5cdbdef1470a0503	refs/heads/main

```

### Binary hashes before cleanup

Source: `46b88c08 and historical 75d0d2ef`. Capture: `binary.sha256`.

```text
4f7e5685c0bfaa4742a38a9190246d84feac48cb901078f8c68aadb76140d202  /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive
86d575251153908f0f2be4fca42201a652fe8460a2a57a721504c2541932f740  /home/penguin/source/hive/.hive/tmp/coordinator/reports/t05-verify-04-captures/hive-pingprior
```

`END VERIFIER T05-herdr-client-and-doctor t05-verify-04`
