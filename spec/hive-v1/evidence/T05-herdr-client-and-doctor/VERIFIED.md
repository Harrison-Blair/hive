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
