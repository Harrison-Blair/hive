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
