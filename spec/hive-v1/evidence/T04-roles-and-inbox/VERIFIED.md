# T04-roles-and-inbox VERIFIED

| Item | Value |
| --- | --- |
| Pass | t04-verify-01, primary |
| Source commit | `7a6250f3f20d2a4fb497c9aa4540628651cb79ae` |
| Base | `84da9819b31be83dbe71c36d620eadf3749e1f9f` |
| Date | 2026-10-18 |
| Agent | verifier, pi, openai-codex/gpt-6.1-sol |
| Verdict | FAIL, provisional primary result |

The coordinator must combine both independent reports before task acceptance.
This report does not accept the task.
The secondary inbox review remains pending for this pass.

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1, AC2 | FAIL | Existing tests pass. F01 to F04 contradict the requirements. |
| V2 | AC4 | FAIL | Existing tests pass. F07 defeats the fixed boundary test. |
| V3 | AC1 to AC4 | FAIL | The specified transcript passes. Extended checks find F01, F02, F05, and F06. |
| V4 | AC5 | PASS | Empty gofmt output. Vet, lint, and uncached race tests pass. |

## Findings

### T04-F01: Refused registrations write the counter

- Rule: AC1 final-key collision rule and partial-state step 3 require no write on a failed check.
- Location: `internal/agent/register.go:191-214`, `237-267`.
- Probe: register explicit worker `worker-hive-1`. Then register a generated worker on another pane.
- Actual: exit 1, outcome `partial`, phase `agent`, and a counter effect. The previously absent counter now contains `{"value":1}`.
- Another probe finds counter bytes change from `{"value":1}` to `{"value":2}` on a busy-pane refusal.
- Owner fix: compute the candidate under the held lock. Check all collisions before `Next`. Keep counter, then agent, as the write order.
- The worker test `TestRegisterGeneratedNameCollisionKeepsTheCounterEffect` asserts the wrong contract.

### T04-F02: Singleton and claim names can collide

- Rule: AC1 permits the singleton exception only for the same claim. Cross-claim and record-key collisions require no write.
- Location: `internal/agent/register.go:237-267`, `internal/agent/show.go:30-44`.
- Actual: worker `shared` and coordinator `--name shared` both register live. The coordinator reports `ok`, exit 0.
- A coordinator at generation 2 and a supervisor at generation 1 can both use `shared` when `shared-g1` is absent.
- A worker named `coordinator` can coexist with the claim `coordinator` whose holder is `leader-g1`.
- Then `agent show coordinator` returns the worker, not the claim holder. Both registration orders reproduce this failure.
- Owner fix: check names, record keys, and claim keys together. Limit each exception to the current claim or a matching orphan.

### T04-F03: An unclaimed orphan resolves as a live singleton

- Rule: AC2 and the partial-state rules make the claim authoritative. A live orphan does not hold the role.
- Location: `internal/agent/show.go:38-52`, `63-91`.
- Probe: inject the first claim-write failure after the agent write. Keep the real store and reader.
- Actual: `Show(home, "coordinator")` reports `ok` and returns the live orphan, although no claim exists. Custom name `leader` does the same.
- Owner fix: resolve singleton names through the matching claim under the held lock. Do not use the name fallback as role ownership.
- Keep direct record-key lookup for inspection.

### T04-F04: An orphan retry can replace another pane identity

- Rule: AC1 permits orphan replacement only for the same intended identity, claim, and generation.
- Location: `internal/agent/register.go:253-259`, `294-316`.
- Probe: fail the first claim write for pane `intended-pane`. Retry with only the pane changed to `different-pane`.
- Actual: outcome `ok`, agent effect `updated`, and claim effect `updated`. The original generation record now names the other pane.
- Owner fix: include pane identity in the matching-orphan check. Refuse a different identity before any write.

### T04-F05: Error text contains unescaped input newlines

- Rule: the approved untrusted-input boundary requires one plain error line, with no unescaped multiline input.
- Location: `internal/agent/register.go:101`, `107`, `127`, `internal/agent/show.go:21`.
- Actual: role, name, project, and show input `a\nb` each produce two physical stderr lines with exit 1.
- Raw-byte captures show the embedded byte `0a`. This result does not depend on newline conversion by the test harness.
- Owner fix: quote control characters in these input values before the command renders the error.

### T04-F06: T04 mutations and errors lack log events

- Rule: `AGENTS.md`, Logging, requires each state transition and error, with each applicable ID.
- Location: T04 registration and inbox mutation paths in `internal/agent/register.go` and `internal/lib/inbox/inbox.go`.
- Actual CLI probe: two registrations, inbox append and consume, and several refusals produce 24 log events.
- All 24 events are `command start` or `command end`. None records a transition, an applicable ID, or an `error` field.
- Owner fix: add structured events for completed T04 effects and errors in owned paths. Preserve the shared output packages.

### T04-F07: The inbox boundary test accepts a broken limit

- Rule: AC4 fixes TEXT at 4096 bytes. `AGENTS.md`, Tests, requires expectations independent of the implementation.
- Location: `internal/lib/inbox/inbox_test.go:232-242`, especially `size <= MaxTextBytes`.
- Mutation: change only `MaxTextBytes` from 4096 to 4097 in an isolated source copy.
- Actual: the complete existing inbox suite still passes. It now expects 4097 bytes to succeed.
- An independent literal-4097 refusal assertion fails on that mutation and passes on the exact worker source.
- Owner fix: use literal acceptance expectations for 4095, 4096, and 4097. Do not derive them from `MaxTextBytes`.

## Runs

Commands ran from the worktree unless the text names the isolated copy.
The binary used a unique ignored path. Each CLI call used an explicit marked fixture home.

```sh
go test -race -count=1 ./internal/agent/ -v
go test -race -count=1 ./internal/lib/inbox/ -v
go test -race -count=1 ./cmd/ -v
```

Exact output excerpts, in command order:

```text
PASS
ok  	github.com/Harrison-Blair/hive/internal/agent	1.198s
exit=0
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	1.173s
exit=0
PASS
ok  	github.com/Harrison-Blair/hive/cmd	1.017s
exit=0
```

The complete V3 transcript passed its 12 expectations.
Raw text, JSON envelopes, exit statuses, records, and actual logs remain in the capture directory.

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run ./...
go test -race -count=1 ./...
```

Exact gate output excerpts:

```text
gofmt_exit=0
vet_exit=0
0 issues.
lint_exit=0
race_exit=0
```

The gofmt command printed no path.
The independent tests used real store operations with deterministic failures before selected writes.
The stale-status mutation fails an existing assertion. The inbox-limit mutation escapes the existing suite.

## Evidence, source freeze, and limits

Root report: `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t04-verify-01.md`.
Captures: `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t04-verify-01-captures/`.
The root report gives the scope, the API recommendations, the cleanup result, and the capture map.

The initial worktree was clean at the exact worker commit.
Every pre-existing tracked path matched its worker Git object before this metadata write.
No product source, V file, specification, baseline, setting, or credential changed.
Only this metadata file belongs in the verification commit.

The check used the local-user and cooperating-writer boundary only.
Hostile swaps remain deferred to T01.
The shared-store oversize-writer gap and the pending T01 home fix are not T04 findings.
No herdr operation, agent start, merge, or global configuration change occurred.
The deeper inbox/output verdict remains pending from the separate verifier.


## 2026-10-18 — t04-verify-02 — current verdict FAIL

| Item | Value |
| --- | --- |
| Role | Fresh independent source verifier |
| Route | pi `openai-codex/gpt-6.1-sol:high` |
| Tested worker HEAD | `34ae036380b81a8e2d945271f71fe745e16fdc21` |
| Source tip | `93aaf8067750abdf5b22aa53548b27f95d3a3490` |
| Original worker | `7a6250f3f20d2a4fb497c9aa4540628651cb79ae` |
| Original base | `84da9819b31be83dbe71c36d620eadf3749e1f9f` |
| Fix initial HEAD | `42a177b43f3d0f5a1b2bcccfd5c2b8d8ec7a004c` |
| Prerequisite | `c6a4d91d82042f8861a2aa4b722944272117fa23` |
| Branch | `task/T04-roles-and-inbox` |
| Overall | **FAIL** |

This pass covers both earlier finding sets. It gives no landing permission.
The earlier verdict keeps its exact bytes and attribution.
The root report records the separate verdict commit and remote tip after delivery.

### V verdicts

| Item | Proves | Verdict | Observed result |
| --- | --- | --- | --- |
| V1 | AC1, AC2 | FAIL | Existing tests pass. P-F02 still permits a cross-claim alias in one order. |
| V2 | AC4 | FAIL | Existing tests pass. I-F03 still redirects a selected-path mark. |
| V3 | AC1 to AC4 | FAIL | All 12 specified steps pass. Extended namespace, identity, and logging checks fail. |
| V4 | AC5 | PASS | Empty format output. Vet, lint, and the full uncached race gate pass. |

### Finding dispositions

`P-Fnn` identifies the earlier primary `T04-Fnn`.
`I-Fnn` identifies the earlier secondary `T04-INBOX-Fnn`.
No new finding class applies.

| ID | Disposition | Evidence |
| --- | --- | --- |
| P-F01 | CLOSED | Generated collision and busy-pane refusal preserve all state bytes and report no effect. The checks precede `Next`. |
| P-F02 | OPEN, partial fix | Worker aliases and live-name collisions refuse. A custom singleton name can still alias a later different claim. |
| P-F03 | CLOSED | Logical and custom singleton names require the matching claim and holder under one lock. Orphans do not resolve. Direct generation inspection remains distinct. |
| P-F04 | CLOSED | A changed orphan pane refuses without a record, claim, or effect change. Matching retries still succeed. |
| P-F05 | CLOSED | Role, name, project, and show control values produce one escaped error line. |
| P-F06 | OPEN, partial fix | Completed mutation logs and registration errors now exist. Owned read errors still lack error events. |
| P-F07 | CLOSED | Fixed literal byte expectations catch both 4097 and 8192 cap mutants. Restored 4096 passes. |
| I-F01 | CLOSED | Short nil-error stdout writes give status 1. Recovery names only actual completed consume IDs. |
| I-F02 | CLOSED | Failed `--all` output claims no consumption and preserves record bytes. |
| I-F03 | OPEN, partial fix | Cross-role, unequal sequence, and scalar records refuse with paths and no mark. A noncanonical numeric filename still redirects a mark. |
| I-F04 | OPEN, partial fix | Append and mark events now exist. Invalid `--all` has no error event. Mark events omit the known sender ID. |
| I-F05 | CLOSED | Text escapes terminal controls in every rendered inbox string. Stored text and JSON text remain intact. |
| I-F06 | CLOSED | Missing role says `pass --role`. Explicit empty and invalid roles retain the invalid-key error. |
| I-F07 | CLOSED | Independent ASCII and UTF-8 byte limits and literal 41/42 key limits pass. Their mutants fail on assertions. |
| I-F08 | CLOSED | Counts reject static links and wrong-type roots or valid role collections with paths. Counts remain lock-free. |

### P-F02: a custom singleton name can alias a later claim

Location: `internal/agent/register.go:276-305`.
Authority: AC1 namespace checks and the same-claim-only exception.
This is the existing claim-alias finding, not a new class.

Actual CLI sequence in an isolated marked home:

```text
$ hive --home <fixture> --json agent register --role coordinator --name supervisor --pane p1
exit 0, outcome ok, agent supervisor-g1, claim coordinator
$ hive --home <fixture> --json agent register --role supervisor --name leader --pane p2
exit 0, outcome ok, agent leader-g1, claim supervisor
$ hive --home <fixture> --json agent show supervisor
exit 0, result key leader-g1, role supervisor
$ hive --home <fixture> --json agent show coordinator
exit 0, result key supervisor-g1, name supervisor, role coordinator
```

Both live registrations succeed and the second changes state.
The new claim hides the first holder's custom name.
The reverse registration order correctly refuses with no write.
The check tests `agentByKey[claimKey]`, but it omits a live singleton's `Name` equal to that claim key.
Recommendation: include live names when a new claim key enters the namespace.
Keep only the exact same-claim and matching-orphan exceptions.

### I-F03: selected numeric filename and mark target differ

Locations: `internal/lib/inbox/inbox.go:175-184`, `216-250`.
Authority: AC4 consume-once and the existing selected-path identity check.

The coordinator's hypothesis reproduces with static regular files.
This result requires no hostile replacement, link, or timing race.
The fixture contains these records:

```text
inbox/coordinator/01.json: seq 1, role coordinator, text SELECTED-01, read_at empty
inbox/coordinator/1.json:  seq 1, role coordinator, text OTHER-1, read_at 2026-01-02T00:00:00Z
```

`ListIn` accepts both filenames. `ParseUint("01", 10, 64)` returns 1.
The stored numeric identity check therefore passes.
The consume selects only `01.json`, but `markRead` reconstructs `1.json` with `FormatUint`.

Two actual CLI consumes each produce this result:

```text
exit 0
outcome ok
result.records[0].text = OTHER-1
effects = [{resource: inbox, path: .hive/state/inbox/coordinator/1.json, action: consumed}]
selected 01.json bytes unchanged = true
```

Each consume changes the already-read `1.json` and prints its text again.
The selected `01.json` remains unread.
This violates the approved static-input contract and keeps I-F03 open.
Recommendation: require the exact canonical sequence filename before any mark, or preserve the exact selected mutation path.
Do not broaden this fix into hostile-store protection.

### P-F06 and I-F04: logging remains incomplete

Locations: `internal/agent/show.go`, `internal/inbox/read.go`, and `internal/lib/inbox/inbox.go:417-439`.
Authority: the existing Logging rule requires each error and applicable IDs.

Actual `agent show unknown` exits 1 with `agent not found: unknown`.
Actual `inbox --role ../x --all` exits 1 with `invalid inbox key: ../x`.
Their home log contains only command start/end events for these failures.
No event contains an `error` field.
The consume path logs role and sequence, but omits the stored sender `worker-x` and its `agent_id`.

The normal registration, stale, claim, append, mark, and refusal events now exist.
Append logs include the sender ID. Registration logs include applicable project and task IDs.
The checked failed writes fabricate no record or claim success.
The fixture message body never appears in a log event.
Recommendation: complete error logs in the owned read paths and retain applicable IDs on mark events.
Do not change the shared output packages.

### Runs and assertion strength

All actual-source commands ran from the task worktree root.
Added Go probes and mutations ran only in isolated Git archives.
The source in the actual worktree stayed frozen.

```sh
go test -race -count=1 ./internal/agent/ -v
go test -race -count=1 ./internal/lib/inbox/ -v
go test -race -count=1 ./cmd/ ./internal/inbox/ -v
```

```text
ok  	github.com/Harrison-Blair/hive/internal/agent	1.210s
exit=0
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	1.191s
exit=0
ok  	github.com/Harrison-Blair/hive/cmd	1.024s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.004s
exit=0
```

The prior primary CLI checklist passes all 193 assertions.
The prior secondary driver completes 49 probes.
Its recorded cross-role and terminal-control observations now show refusal and safe output.
The follow-up driver repeats both alias orders, long names, raw error bytes, and ten simultaneous worker pairs.
The targeted checks above expose the remaining failures.
`residual-assertions-red.txt` records three assertion failures, not build failures.
Those remaining acceptance assertions have no restored green result. Their fixes remain uncompleted.

The historical replay overlays only five old production files from `42a177b4` onto final tests in an isolated copy.
It compiles and fails the intended counter, namespace, orphan, error, log, stdout, identity, count, and render assertions.
Restoring the exact tested source makes the same complete package selection pass.
Both prior independent Go probe sets pass on that restored source.

| Assertion check | Red observation | Restored result |
| --- | --- | --- |
| Historical fixed findings | Exit 1 on behavioral assertions | Exit 0 |
| Cap 4097 | Literal 4097 ASCII and UTF-8 refusal assertions fail | Cap 4096 passes |
| Cap 8192 | Literal 4097 and 8192 refusal assertions fail | Cap 4096 passes |
| Key limit 42 | Literal 42-byte refusal fails | Original limit passes |
| Empty `read_at` mutation | Mark and partial-prefix assertions fail | Exact mark source passes |
| Live instead of stale mutation | Generation status assertion fails | Exact stale source passes |
| Injection-only CLI tests | No injection gives assertion failures | Deterministic mark failures pass in text and JSON |

Independent literal UTF-8 cases cover 4095, 4096, and 4097 bytes.
Both cap mutants fail that added check. The exact source passes.
The corrected generated-collision test strengthens the old incorrect counter expectation.
No unrelated test weakens or skips an assertion.

The repeated checks retain these existing classes:

- Role matrix, environment precedence, explicit pane handling, record fields, and explicit names.
- Counter-first writes after checks, safe gaps, singleton write prefixes, retries, and old generations.
- Ten concurrency repetitions, disjoint grouped consumes, 50 process appends, and numeric order through maximum uint64.
- Lock-free reads and counts, bounded mutation locks, and output after completed marks outside the lock.
- First-mark failure, known partial prefixes, remaining records, nil Go-error routing, and no automatic replay.
- Static invalid record shapes, types, paths, bounds, and unchanged unrelated records.
- Actual failed stdout, short writes, read-only `--all`, safe text, and intact JSON data.

### V3 transcript

The actual binary builds from the frozen task worktree.
All 12 specified transcript steps pass with an explicit marked fixture home.
The list shows generation 1 stale and generation 2 live.
Show returns generation 2 on `w1:p2`.
The first consume prints `hello`. The second prints zero bytes.
The final JSON contains one marked record.
The four specified refusals exit 1 with their exact expected diagnostics.
`v3-transcript.txt` retains the complete user-visible output and each exit status.

### V4 full uncached gate

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run ./...
go test -race -count=1 ./...
```

```text
gofmt_exit=0
vet_exit=0
0 issues.
lint_exit=0
race_exit=0
```

The format command prints no path. Every package passes the uncached race gate.
The store package takes 17.225s. The full output is in `gate.txt`.
The STE baseline and all source files remain unchanged.

### Scope, prerequisites, and source freeze

Initial HEAD, branch, clean status, and direct remote tip match the required worker.
`PI_PROVIDER`, `PI_MODEL`, and session metadata record the required provider, model, and `high` level.
I read both complete earlier verdicts, both probe sets, all V files, the installed task, and the final worker report.
The worker report ends with `END WORKER T04-roles-and-inbox t04-fix-01`.
I inspected the complete original-base-to-worker diff and the complete fix diff.
The full diff has 47 changed paths and SHA256 `7efafbe09063be66d2e5fd7c5b5b84059113aee5ed4a5a70ff8fbe56406ba476`.
`changed-files.txt` records every path, including inherited prerequisite paths.
The root report lists the full range separately from the 14 fix paths.

The prerequisite is an ancestor of the tested worker.
The CP1 report gives PASS and keeps its final marker. I did not rerun CP1.
The cap and home landing report is final at `c6a4d91d`.
The current cap and home source paths equal their independently verified objects.
The full gate rechecks their installed tests.

All 179 tracked paths match the exact worker objects before this append.
The source tip and tested worker contain identical Go source, module files, and STE baseline.
The isolated source copies restore every original path before removal.
The old verdict prefix has SHA256 `e7d70367f2f5bcd74f130af4f1153de4513b095e32f0ccf3d9412dc5f8dce41e`.
Only this append enters the verdict commit.

### Captures, cleanup, and limits

Root report: `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t04-verify-02.md`.
Captures: `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t04-verify-02-captures/`.
The captures retain scripts, actual outputs, red/green runs, path bytes, logs, source comparisons, and cleanup receipts.

Cleanup removes all own marked fixture homes, both unique binaries, and isolated source copies.
The primary scratch contains 65 marked homes before removal.
The secondary and targeted drivers remove their own fixture directories.
All earlier reports, logs, credentials, worktrees, and attribution remain unchanged.
No credential content inspection, global configuration change, agent start, herdr call, merge, or worktree removal occurs.

The approved boundary remains local user and cooperating writer only.
Hostile replacement between `Lstat` and `Open` remains deferred for T01.
The public held-handle inbox API belongs to T11 and is not a finding.
S05-D01 remains a separate T22 store-walker follow-up. It changes no T04 acceptance criterion.
The open P-F02, P-F06, I-F03, and I-F04 fixes remain uncompleted.
Later integration and landing remain uncompleted.
Recommendation: return these bounded findings to the worker. Do not accept T04 yet.


## 2026-10-18 — t04-verify-03 — current verdict PASS

| Item | Value |
| --- | --- |
| Role | Fresh independent source verifier |
| Route | pi `openai-codex/gpt-6.1-sol:high` |
| Tested worker HEAD | `c33f58965302f0934294b2586e0055cd528746fa` |
| Source tip | `0b896f66c3868a75fc5bff1dc168ffa297dd054d` |
| Original base | `84da9819b31be83dbe71c36d620eadf3749e1f9f` |
| Original worker | `7a6250f3f20d2a4fb497c9aa4540628651cb79ae` |
| First primary verdict | `3769d5651f8b475c65fd395bf267398f83efe49d` |
| First fix initial HEAD | `42a177b43f3d0f5a1b2bcccfd5c2b8d8ec7a004c` |
| Second verdict and second fix initial HEAD | `64fdb038bd7d35fbecc4d89046893cb9a2dddd2d` |
| Prerequisite | `c6a4d91d82042f8861a2aa4b722944272117fa23` |
| Branch | `task/T04-roles-and-inbox` |
| Overall | **PASS** |

This verdict covers both earlier finding sets and both fix passes.
Both earlier verdicts retain their exact bytes and attribution.
The root report records the separate verdict commit and remote tip after delivery.
This verdict does not perform integration or landing.

### V verdicts

| Item | Proves | Verdict | Actual result |
| --- | --- | --- | --- |
| V1 | AC1, AC2 | PASS | All register and claim tests pass. Independent failure-prefix, identity, namespace, and concurrency checks pass. |
| V2 | AC4 | PASS | All inbox tests pass. Independent byte bounds, identity, counts, partial marks, and output checks pass. |
| V3 | AC1 to AC4 | PASS | All 12 specified steps pass. Both prior CLI checklists and all four residual findings pass. |
| V4 | AC5 | PASS | Empty format output. Vet, lint, and the full uncached race gate pass from the task worktree root. |

### Every finding disposition

`P-Fnn` identifies primary `T04-Fnn`.
`I-Fnn` identifies secondary `T04-INBOX-Fnn`.
No new finding class applies.

| ID | Disposition | Current evidence |
| --- | --- | --- |
| P-F01 | CLOSED | Generated collision and busy-pane refusal preserve counter and record bytes with no effect. Checks precede `Next`. |
| P-F02 | CLOSED | Live names, record keys, and claim aliases refuse in both orders. The custom-name-before-claim and record-key-before-name cases refuse without writes. Same-claim replacements and matching orphans still succeed. |
| P-F03 | CLOSED | Logical and custom singleton lookups require matching claim and holder under one lock. An unclaimed orphan does not hold the role. Direct generation-key inspection remains separate. |
| P-F04 | CLOSED | A changed orphan pane refuses with unchanged bytes and no effect. A matching retry updates the intended generation. |
| P-F05 | CLOSED | Role, name, project, and show controls produce one escaped physical error line. |
| P-F06 | CLOSED | Registration and claim events carry applicable IDs. Show and List errors now log. Checked failed writes fabricate no success. |
| P-F07 | CLOSED | Literal 4095/4096/4097 byte assertions detect both cap mutants. Restored 4096 passes. |
| I-F01 | CLOSED | Short nil-error stdout gives status 1 and actual consumed IDs. Marks stay. Output occurs outside the lock. |
| I-F02 | CLOSED | Failed `--all` output claims no consumption and preserves record bytes. |
| I-F03 | CLOSED | Cross-role, sequence mismatch, scalar, and noncanonical-name records refuse with paths before any mark. Both `01.json` and `1.json` retain exact bytes. No wrong record prints. |
| I-F04 | CLOSED | Append and mark events carry the sender. Actual mark `agent` and `agent_id` equal `worker-x`. All, Unread, and count errors log. No message body appears. |
| I-F05 | CLOSED | Every rendered inbox string escapes terminal controls. Raw stored and JSON text stay intact. |
| I-F06 | CLOSED | Absent role says `pass --role`. Explicit empty and invalid keys retain the invalid-key error. |
| I-F07 | CLOSED | Independent ASCII and UTF-8 4095/4096/4097 byte cases and 41/42 key cases pass. Their mutants fail. |
| I-F08 | CLOSED | Lock-free counts reject static linked and wrong-type roots or valid role collections with paths and no partial map. |

### Four residual cases

The actual CLI now refuses custom-name-before-claim in both orders with `agent exists: supervisor`.
The prospective record-key alias refuses with `agent exists: leader-g1`.
Each refusal exits 1, reports no effect, and preserves all state bytes.
Normal and same-claim generation replacements remain successful.

The static `01.json` and already-read `1.json` fixture now refuses two consumes and `--all`.
The error names `01.json` and says `the sequence name is not canonical`.
Both files retain their full bytes. No `OTHER-1` record prints and no consumed effect appears.
This closes the reproduced ParseUint/FormatUint defect, not a broader hostile-store guarantee.

Actual `agent show unknown` logs `agent show failed` with its name and error.
Actual invalid `inbox --all` logs `inbox operation failed` with phase `read`.
The actual mark event retains role `coordinator`, sequence 1, and sender `worker-x` in both sender fields.
No event contains `BODY-MUST-NOT-LOG`.
Added All, Unread, UnreadCount, and UnreadCountAll checks cover scalar records and wrong-type collections.
All 12 preserve bytes, retain paths, and log errors.
Existing Show and List error checks retain their applicable name, role, and project fields.

### Actual runs

```sh
go test -race -count=1 ./internal/agent/ -v
go test -race -count=1 ./internal/lib/inbox/ -v
go test -race -count=1 ./cmd/ ./internal/inbox/ -v
```

```text
ok  	github.com/Harrison-Blair/hive/internal/agent	1.230s
exit=0
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	1.198s
exit=0
ok  	github.com/Harrison-Blair/hive/cmd	1.027s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.004s
exit=0
```

The primary CLI checklist passes all 193 assertions.
The secondary driver completes all 49 probes.
The worker residual driver passes all 47 checks on the independently built binary.
All eight independent residual assertions pass.
Both prior independent Go probe sets pass. Ten uncached concurrency repetitions pass.
The follow-up driver retains long names, reverse aliases, raw controls, and ten simultaneous worker pairs.
The checklist retains numeric order through maximum uint64, disjoint consumes, and 50 process appends.
It retains all singleton failure prefixes, safe gaps, matching retries, and old generations.
It retains lock-free reads, bounded mutation locks, known partial-mark prefixes, nil Go-error routing, and no replay.
Output follows completed marks outside the lock. Static invalid inputs produce no unintended mark.

### Assertion strength

All added Go probes, mutations, and historical replays run only in isolated Git archives.
The actual worktree source stays frozen.

| Check | Observed red | Restored green |
| --- | --- | --- |
| First fix replay | Six production files from `42a177b4` fail current behavioral assertions | Exact worker passes the same package selection |
| Second fix replay | Four production files from `64fdb038` fail namespace, selected-path, reader-log, and sender assertions | Exact worker passes the same selection |
| Second fix CLI replay | Six independent residual assertions fail on the historical binary | All eight pass on the frozen worker binary |
| Added reader error logs | All 12 assertions fail on second-verdict source | All 12 pass on exact worker source |
| Cap 4097 and cap 8192 | Literal ASCII and UTF-8 refusal assertions fail | Cap 4096 passes |
| Key cap 42 | Literal 42-character refusal fails | Original cap passes |
| Empty mark | `read_at` and partial-prefix assertions fail | Exact mark source passes |
| Live instead of stale | Generation status assertion fails | Exact stale source passes |
| Injection-only CLI checks | Without injection, error and prefix assertions fail | Deterministic mark failures at positions 1, 2, 3 pass in text and JSON |

Every red result is an assertion failure, not a compile failure.
The corrected generated-name collision test strengthens its incorrect historical expectation.
No unrelated test weakens or skips an assertion.

### V3 transcript and V4 gate

All 12 installed transcript steps pass with an explicit absolute marked fixture home.
The actual binary builds from the frozen task worktree root.
The list shows generation 1 stale and generation 2 live. Show returns generation 2 on `w1:p2`.
The first consume prints `hello`. The second prints zero bytes.
JSON `--all` contains one record with `read_at` set.
The four specified refusals exit 1 with their required diagnostics.
`v3-transcript.txt` retains every command, output, and status.

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run ./...
go test -race -count=1 ./...
```

```text
gofmt_exit=0
vet_exit=0
0 issues.
lint_exit=0
race_exit=0
```

The gate runs serially from the actual task worktree root.
The format command prints no path. Every package passes. Store takes 17.293s.
The STE baseline stays unchanged.

### Scope, source freeze, and delivery

Initial HEAD, branch, clean status, and direct remote match the required worker.
Route metadata records `openai-codex`, `gpt-6.1-sol`, and `high` at `2026-10-18T08:07:07.895Z`.
The session ID is `01a14e0c-db61-76a2-b039-0e21e2f4281b`.
I selected only route metadata from the session file.
I read the required instructions, glossaries, installed task, all V files, complete prior verdicts, and both probe sets.
The complete worker report ends with `END WORKER T04-roles-and-inbox t04-fix-02`.
I read its captures and inspected the whole task range and both fix ranges.
The root report lists every changed path for each range.

| Range to exact tested worker | Changed paths | Diff SHA256 |
| --- | --- | --- |
| `84da9819` | 47 | `5e9660cdb1855936aaa5466d67a7f618166cfb524ff30681a6997b019e30617a` |
| `42a177b4` | 17 | `5a8a4e934549b4feee89363266b2c1f9ba04fbd05862b4e5ac6cc59cfd37b6ef` |
| `64fdb038` | 13 | `ee5018ac8d46dad1f9582791e3975cf50b20a6299abc1dd824daf019fd891847` |

The prerequisite is an ancestor. The CP1 PASS and cap/home landing reports retain their final markers.
I did not rerun CP1. All two home paths and 14 store paths equal `c6a4d91d`.
The full gate rechecks their installed tests.
Later root T03/T22 merges and T05 synchronization require no merge into this frozen source.
No T04 production path calls WalkRecords.

All 179 tracked paths match exact worker objects before this append.
Every other tracked path remains exact at commit time.
The old 22034-byte prefix keeps SHA256 `717f4aef7db83274d529a214fa790bec1cb6fed844bd72fab315ac437b81b81a`.
Source, module files, and the STE baseline remain unchanged.
Each isolated source copy restores every original tracked path before removal.
Only this append enters the evidence-only verdict commit.

Cleanup removes the primary scratch with 65 marked homes, both worker-source binaries, and both isolated source copies.
Other drivers remove their own marked fixtures. Cleanup also removes the historical binary.
Necessary captures and probe definitions remain.
A copied binary initially lacked execute mode. I corrected only that mode and reran both affected drivers.
The separate setup-error captures count as no product result.
All earlier work, reports, logs, credentials, worktrees, and attribution remain unchanged.
No credential inspection, global configuration change, herdr call, agent start, merge, or worktree removal occurs.
The push changes only the named task branch without force. No main reference changes.
No co-author or session trailer applies.

Root report: `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t04-verify-03.md`.
Captures: `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t04-verify-03-captures/`.
Only the approved local-user and cooperating-writer boundary applies.
Hostile replacement between Lstat and Open remains deferred for T01.
The public held-handle inbox API remains assigned to T11 and is not a T04 finding.
S05-D01 remains a separate T22 follow-up. No bypass or store edit occurs.
No owned acceptance check remains uncompleted. Later integration and landing remain uncompleted.
Recommendation: accept this exact bounded T04 source for the coordinator's next delivery step.
