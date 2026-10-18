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
