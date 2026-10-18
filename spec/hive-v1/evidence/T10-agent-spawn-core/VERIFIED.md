# T10-agent-spawn-core VERIFIED

| Item | Value |
| --- | --- |
| Task | T10-agent-spawn-core |
| Pass | t10-verify-01 |
| Date | 2026-10-18 |
| Entry | `31989721c575fd81860a77664ed87b066ebfe3be` |
| Final Go source | `f225bfa9eb4541598edf6f298a743a4f19932736` |
| Task blob | `5db0e34783fbfa1f73aca232e7e0f8008dff02fd` |
| Agent | verifier, Pi, `openai-codex/gpt-6.1-sol:high` |
| Verdict | **FAIL** |

The source does not pass independent verification.
The native flow and the full gate pass. Six source findings remain open.
No source fix or test change occurs in this verdict.
The approved boundary is the local user and a cooperating writer.
Hostile replacement between `Lstat` and `Open` remains deferred for T01.

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | PASS | Profile tests and independent literal checks |
| V2 | AC2 | FAIL | T10-O-F01, T10-O-F02, T10-P-F01 |
| V3 | AC3, AC4 | FAIL | T10-O-F01, T10-O-F03, T10-P-F01, T10-P-F02 |
| V4 | AC2, AC3, AC4 | PASS for the native flow | Fresh exact-source Claude spawn, prompt, and stop |
| V5 | AC5 | PASS | One full actual-worktree gate |

## Reports and artifacts

Primary report: `.hive/tmp/coordinator/reports/t10-verify-01.md`.
Primary captures: `.hive/tmp/coordinator/reports/t10-verify-01-captures/`.
Secondary report: `.hive/tmp/coordinator/reports/t10-protocol-verify-01.md`.
Secondary captures: `.hive/tmp/coordinator/reports/t10-protocol-verify-01-captures/`.
The secondary report ends with `END T10 PROTOCOL VERIFY 01` and records area FAIL.
Its final preservation and cleanup checks pass.
The coordinator read the complete report and actual complete final response in this pass.
The coordinator attests external Pi session `85387` exit `PI_EXIT0` and complete preservation and cleanup.
Coordinator attestation closes the secondary completion barrier. Both source areas FAIL. This verdict claims no overall PASS.

The primary actual route comes from `PI_PROVIDER`, `PI_MODEL`, and `PI_REASONING_LEVEL`.
The route is `openai-codex/gpt-6.1-sol:high`.
The primary session is `01a14f51-8365-7545-8c86-3bff7bc9a379`.
All independent Go edits occur only in owned temporary project copies.
Raw runs retain their commands, working directories, durations, stdout, stderr, and exit statuses.
The binary, probes, red mutants, restored source, logs, and cleanup receipts remain retained.

## Open findings

### T10-O-F01: uncertain inspection marks the record exited

Criterion: AC3, AC4 inspection, and architecture Recovery.
Location: `internal/agent/inspect.go:38-77`.
Production `Send` receives `agent_not_found` from the external herdr boundary.
Inspection marks the record `exited` for each of these conditions:
- The pane is missing.
- The foreground process list is empty.
- An unrelated process holds the foreground.
- A shell holds a different terminal.

Each condition requires `unresolved`, not positive evidence of agent exit.
The correct same-terminal shell control passes.
The restoration placeholder control also passes through the existing process-info error path.
Reproduction: `TestIndependentInspect` in `independent_primary_test.go`.
Raw result: `independent-probes-2.stdout`, exit 1.
Recommendation: require positive shell evidence and the recorded terminal match before writing `exited`.
Keep absent or uncertain ownership unresolved. No shared type extension is necessary.

### T10-O-F02: spawn ignores the start-reply session

Criterion: AC2 requires a present session reference to be recorded and compared.
Location: `internal/agent/spawn.go:120-145,185-186`.
A typed start reply contains `start-session`.
The later wait and get replies contain `different-session`.
Production spawn returns `ok` and writes `live` with the different session.
The required result is `unresolved` after that identity change.
Reproduction: `TestIndependentStartSessionIsRecorded` in `independent_primary_test.go`.
Raw result: `independent-probes-2.stdout`, exit 1.
Recommendation: retain the first present session and compare later session evidence against it.
Keep optional absence alone valid. This probe makes no native-observation claim.

### T10-O-F03: stop accepts a changed recorded identity

Criterion: AC4 and the required fresh ownership check before final mutation and close.
Locations: `internal/agent/stop.go:105` and `internal/agent/spawn.go:476-480`.
A cooperating store write changes the terminal and process start during the external workspace read.
The key, pane, holder, and generation stay unchanged.
Production stop writes this changed record `stopped` and sends one close request.
It returns `ok`. The required result is refusal before the close.
The final check omits the terminal, process, and session tuple that the earlier reads checked.
Reproduction: `TestIndependentStopFreshIdentity` in `independent_primary_test.go`.
Raw result: `independent-probes-2.stdout`, exit 1.
Recommendation: compare the complete checked ownership tuple under the final held handle.
Refuse a changed identity. Keep socket waits outside the store lock.
This fake-boundary probe stops no real pane or unknown process.

### T10-O-F04: the herdr package map omits the new files

Criterion: the task source table requires new `internal/lib/herdr/doc.go` map lines.
Location: `internal/lib/herdr/doc.go:3-7`.
The actual committed map omits `agent.go` and `pane.go`.
The base-to-entry diff is empty. Both blobs equal `a6952fb2a38eaf647799c0f4d9e567a94a1f57c3`.
I checked this existing-scope obligation after the coordinator named the omission.
Recommendation: add the two required map lines. Preserve the existing map.
No source edit or new behavior probe occurred for this finding.

### T10-P-F01: typed wrappers accept malformed success replies

Criterion: protocol 22 typed reply presence, null, enum, and collection checks.
Locations: `internal/lib/herdr/agent.go:76,133-173` and `internal/lib/herdr/pane.go:65-102`.
Missing or null required `focused` and `revision` fields pass as Go zero values.
An invalid agent status passes. Missing or null agent lists pass as empty lists.
A null argv item passes as an empty string. Required process fields also lose presence information.
The actual CLI receives `agent_prompted` with `focused:null` and exits 0 with `ok/complete`.
The secondary report and raw CLI capture record the observed result.
Reproductions: `TestPVRequiredTypes`, `TestPVRequiredCollections`, and `TestPVSendRejectsMalformedSuccess`.
Recommendation: check the local wire shape in the new wrapper files before reporting success.
Keep unknown extra fields and legitimate optional absence or null values valid.
Do not widen the authorized shared `types.go` edit.

### T10-P-F02: a large prompt timeout wraps

Criterion: AC3 requires the requested server timeout and a local deadline five seconds longer.
Location: `internal/agent/send.go:47-52,76-81`.
The actual CLI accepts `--wait --timeout 9223372036854775807` and sends the prompt.
It sends `timeout_ms:18446744073709550616` after integer overflow.
The local deadline becomes four seconds. A pending request returns partial after about four seconds.
A peer that replies at once permits the same invalid timeout to return `ok`.
Reproduction: secondary `overflow-deadline.txt` and `cli-probe-results.json`.
Recommendation: check both conversions before logging or sending the prompt.
Refuse an unrepresentable value without an arbitrary policy limit.

## Runs

### V1

The secondary ran the profile, wrapper, process, and inherited agent suites with race detection.
Its `existing-race-short.txt` records exit 0.
Independent profile and literal process checks also pass.
The primary actual-worktree race gate independently includes the same tracked profile tests.

### V2 and V3

The primary ran these commands in its exact owned project copy:

```sh
go test -race -count=1 ./internal/agent/ -run TestSpawn -v
go test -count=1 ./internal/agent/ -run 'TestSend|TestStop' -v
go test -count=1 ./internal/agent/ -run TestIndependent -v
go test -count=1 ./cmd/agent/spawn/ -v
```

The inherited V2 and V3 commands exit 0.
The independent production-behavior assertions exit 1 with T10-O-F01 to T10-O-F03.
The CLI suite exits 0. Its partial-claim case retains the created-record effect and phase.
The CLI checks also cover one JSON object, repeated arguments, output failure, and early home failure.
Complete output is in `v2-replay.stdout`, `v3-replay.stdout`, `independent-probes-2.stdout`, and `cli-replay.stdout`.
The first independent run hit a Unix socket path limit. The corrected fixture retains the source failures.
No worktree Go source or inherited test changed.

Four current mutations break the start key, wait omission, terminal guard, and UTF-8 body cap.
`mutation-red.stdout` records four assertion failures, exit 1.
The exact source restore gives `mutation-restored-green.stdout`, exit 0.
The original red logs show assertions, but the original stub source snapshot is absent.
Exact original-stub replay remains BLOCKED. No reconstructed chronology replaces that missing evidence.

### V4

The tested binary SHA256 is
`111cb4e9981c5bbb13b78a071149030597bf09882abb30ea8ebc9849d5545616`.
Its build information records entry revision `31989721` and `vcs.modified=false`.
Its Go source equals `f225bfa9` exactly.
The unique fixture home was `/tmp/t10primv01.BhtPZaDb/native` with its own marker.
The installed harness was Claude Code `2.1.302`, model `claude-haiku-5-5`.

The first owned pane `wB4:p5` reached the trust dialog and gave a truthful partial result.
I accepted that fixture trust dialog. I stopped that Claude before the retry.
The second owned pane was `wB4:p6` in new tab `wB4:t6`.
These are exact output lines from the successful attempt:

```text
agent t10-ver-02 live
agent t10-ver-02 done
● pong
closed pane wB4:p6
workspace surviving
```

Spawn, send with `--wait --timeout 120`, and stop each exit 0.
My own pane read shows the separate Claude text reply, not only the echoed prompt.
The live list shows `live`. The all list after stop shows `stopped`.
The later pane get gives `pane_not_found`. The owned tab is absent.
The parent `wB4:p1`, its terminal, and workspace survive.
Both owned Claude processes ended before the verdict. Only one ran at a time.
The worker's actual manual stop refusal also passes transcript audit.
The independent tests prove its zero-request refusal. I did not repeat that native manual case.

### V5

These commands ran once in the actual task worktree after the behavior probes:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run --allow-serial-runners ./...
go test -race -count=1 ./...
```

The format output is empty. Each command exits 0. Lint reports `0 issues.`.
Full output and statuses are in `gate-gofmt`, `gate-vet`, `gate-lint`, and `gate-race` capture files.
The focused STE tests pass. New prose has zero hard findings with both glossaries.
The baseline and dependencies remain exact. The full gate does not close the independent findings.

## Import, activation, preservation, and limits

The helper import is exactly `12d25801837c3f0e39261574d504082e990281a5`.
It is the sole child of `e696d9ca`. Merge `dca0df1a` has parents `6747c5a1` and `12d25801`.
Only the four approved helper paths enter the merge. Their modes and blobs match exactly.
Every nonincoming entry equals the first parent. T10 consumer logging regressions pass.
Real daily JSONL covers successful and failed external replies, argv, prompts, state events, and errors.
Decoded bodies stay within 4096 bytes with complete UTF-8 and truthful truncation flags.
A synthetic environment secret marker does not appear. I read no credential content.
The native fixture trust action can leave a trust entry. I changed no global harness configuration.

The immutable Root activation audit passes against read-only Root `b4ab63ca`.
The actual Hive handoff envelope is `ok/complete` with one effect.
The prospective input, current handoff, old sections, historical tail, and prior pointer match the retained evidence.
Requested worker level `medium` and effective level `high` remain separate historical facts.
No later Root write occurs in this pass. This audit accepts no Root source or checkpoint.

The freeze checks 625 preexisting entries with no change.
The old reports, worker captures, tracked source, controls, and named Root state keep their bytes.
I removed only the owned fixture and empty Go temporary directory after exact no-follow manifest checks.
Their copies, binary, source probes, red output, restored source, and cleanup receipts remain retained.
`TMPDIR` is unset. I removed no source runtime data or other agent resource.
The local and remote `main` references stay separate and unchanged.
Source fixes and fresh independent verification remain uncompleted.
No Root landing, CP2 acceptance, or CP4 acceptance follows from this verdict.

END T10 VERIFY 01
