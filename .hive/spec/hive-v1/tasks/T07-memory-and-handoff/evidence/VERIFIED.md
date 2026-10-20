# T07-memory-and-handoff VERIFIED

| Item | Value |
| --- | --- |
| Task | T07-memory-and-handoff |
| Tested worker | `26b095f0b17b4fa509abe84263f0a22ad4dea515` |
| Base | `c6a4d91d82042f8861a2aa4b722944272117fa23` |
| Date | 2026-10-18 |
| Agent | Fresh independent verifier, `pi openai-codex/gpt-6.1-sol:high` |
| Primary pass | `t07-verify-01` |
| Secondary pass | `t07-atomic-verify-01`, same tested worker |
| Verdict | **FAIL: M-F01, H-F01, H-F02** |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1, AC2 | **FAIL** | M-F01 accepts frontmatter without the required colon-space separator. Other checked memory behavior passes. |
| V2 | AC3 | **PASS** | Secondary handoff and atomic checks pass. Both logging findings remain separate. |
| V3 | AC2, AC3 | **PASS for required transcript** | Exact effects, file bytes, and exits pass. Additional probes reproduce all three findings. |
| V4 | AC4 | **PASS** | Fresh full uncached gate, TestSTE, and diff check pass. |

Recommendation: return all three findings to the worker under a new bounded brief.
This verdict does not accept T07 for landing.
The verifier changed no implementation or worker V file.

## Reports and captures

Primary report:

`/home/penguin/source/hive/.hive/tmp/coordinator/reports/t07-verify-01.md`

Primary captures:

`/home/penguin/source/hive/.hive/tmp/coordinator/reports/t07-verify-01-captures/`

Secondary report:

`/home/penguin/source/hive/.hive/tmp/coordinator/reports/t07-atomic-verify-01.md`

Secondary captures:

`/home/penguin/source/hive/.hive/tmp/coordinator/reports/t07-atomic-verify-01-captures/`

The primary read the complete secondary report through `END T07 ATOMIC VERIFY 01` before this verdict.
The secondary report SHA256 is `e31d814eba7cd279e6b51013dede7a75a7900d7748ecbe3e01a15ae1250998e3`.
Both areas tested the same exact worker.
The full reports give each command, result, scope boundary, and retained capture.

## Open findings

### M-F01: frontmatter separator, medium, T07 memory owner

`internal/memory/frontmatter.go:66-69` accepts a plain colon as a fallback separator.
AC2 specifies `key: value` lines and the first `: ` separator.
The frozen CLI accepts `name:fact` and `description:D.` with exit 0 and both effects.
An existing file with these lines also passes the scan.
A later valid write then changes the old index instead of refusing the operation.

Recommendation: remove the plain-colon fallback.
Keep trimming after the specified separator and unknown keys without effect.
Cover stdin and existing-file refusal before the owner changes the parser.
No dependency or broader parser is necessary.

### H-F01: ACCEPTED, OPEN, high, T07 memory logging owner

The mismatch error inserts the parsed frontmatter name into its error string.
`internal/memory/write.go:193-219` sends that string to the log.
Both independent areas reproduce a synthetic input value in fixture logs.
No real secret or credential was used.

Recommendation: separate required user errors from log-safe categories.
Keep the required mismatch error on command output.
Exclude parsed content from stdin and existing-file error logs.

### H-F02: ACCEPTED, OPEN, medium, T07 memory and handoff logging owner

Project-scoped memory, index, and handoff events omit the applicable `project_id` field.
Both areas reproduce three events with `scope: project:demo` but no `project_id: demo`.

Recommendation: add checked project identity to applicable success and failure events for both commands.
Keep the scope field and add project-scoped assertions.
No new general logging abstraction is necessary.

The final secondary report has no other H-F finding.

## V1 runs

The frozen worktree command passed without cache:

```sh
go test -race -count=1 ./internal/memory/ ./internal/lib/scope/ -v
```

Full output: primary capture `frozen-V1.txt`, exit 0.
The independent positive checks use a separate source copy and fixed specification literals.
They check whole-input and existing-file sizes 1048575, 1048576, 1048577, and 5000000, including frontmatter.
They check all required frontmatter errors, names, scopes, 10000 unknown keys, and existing-file refusal before replacement.
They also check static bad types, unreadable files, and exact scan exclusions.

The final positive command ran from the primary capture `source/` directory:

```sh
go test -race -count=1 ./internal/t07memoryverify -skip TestSeparatorFinding -v
```

Full output: `final-independent-green.txt`, exit 0.
The skipped test is the separately executed open finding, not a hidden gate failure.
The frozen full gate skips no acceptance test.

The independent finding command ran from the same source copy:

```sh
go test -race -count=1 ./internal/t07memoryverify -run TestSeparatorFinding -v
```

```text
=== RUN   TestSeparatorFinding
=== RUN   TestSeparatorFinding/frontmatter:_missing_name
    memory_test.go:99: separator without space accepted: output.Envelope{Outcome:"ok", Phase:"complete", Effects:[]output.Effect{output.Effect{Resource:"memory", Path:"memory/fact.md", Action:"created"}, output.Effect{Resource:"index", Path:"memory/MEMORY.md", Action:"updated"}}, Error:interface {}(nil), Result:interface {}(nil)}, want frontmatter: missing name
    memory_test.go:102: invalid memory was written
=== RUN   TestSeparatorFinding/frontmatter:_missing_description
    memory_test.go:99: separator without space accepted: output.Envelope{Outcome:"ok", Phase:"complete", Effects:[]output.Effect{output.Effect{Resource:"memory", Path:"memory/fact.md", Action:"created"}, output.Effect{Resource:"index", Path:"memory/MEMORY.md", Action:"updated"}}, Error:interface {}(nil), Result:interface {}(nil)}, want frontmatter: missing description
    memory_test.go:102: invalid memory was written
=== RUN   TestSeparatorFinding/existing
    memory_test.go:112: invalid existing separator accepted: output.Envelope{Outcome:"ok", Phase:"complete", Effects:[]output.Effect{output.Effect{Resource:"memory", Path:"memory/fact.md", Action:"created"}, output.Effect{Resource:"index", Path:"memory/MEMORY.md", Action:"updated"}}, Error:interface {}(nil), Result:interface {}(nil)}
    memory_test.go:115: old index changed
--- FAIL: TestSeparatorFinding (0.00s)
    --- FAIL: TestSeparatorFinding/frontmatter:_missing_name (0.00s)
    --- FAIL: TestSeparatorFinding/frontmatter:_missing_description (0.00s)
    --- FAIL: TestSeparatorFinding/existing (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/t07memoryverify	0.004s
FAIL
exit 1
```

Four temporary mutants produced meaningful assertion failures in the isolated copy:

| Mutant | Capture | Result |
| --- | --- | --- |
| Read existing files before the store lock | `mutant-lock-red.txt` | Exit 1. Inotify detects an early open and the scan misses a cooperating change. |
| Increase the existing-file cap to 2 MiB | `mutant-cap-red.txt` | Exit 1. The fixed 1048577-byte case detects the accepted oversized file. |
| Return `ok` for an index failure | `mutant-partial-red.txt` | Exit 1. Both exact partial-envelope assertions fail. |
| Reverse index ordering | `mutant-sort-red.txt` | Exit 1. The two-process exact-index assertion fails. |

The mutants coexisted. The partial run also detected reversed payload ordering.
Every mutant restored to exact worker bytes before the final positive run.
No compilation-only failure counted.
No source fix ran.

The grouped-write checks pass:

- One lock contains existing reads, the fact write, index construction, and index writing.
- Stdin starts and reaches its 100 ms timeout without waiting for a held store lock.
- A separate lock acquisition succeeds while stdin remains blocked.
- Inotify records no existing-memory open while another callback holds the lock.
- A cooperating bad-file change under that lock prevents the waiting command from writing.
- Ten pairs of OS processes produce the exact sorted two-entry index.
- Static directory and link obstructions at `MEMORY.md` refuse without changing the old fact.
- Injected later index failures retain created and updated facts with only the exact memory effect.
- The old regular index remains unchanged, and a later distinct write regenerates all retained facts.

Only explicit external index errors use a fault callback.
The secondary also checks actual CLI index rename failures after successful static checks and fact writes.
The held-lock callback returns nil on partial outcomes.
The CLI wrapper preserves the partial envelope with a nil Go error and exit 1.
No rollback, journal, replay, or static obstruction substitutes for the partial test.

## V2 secondary disposition

The secondary records uncached handoff and scope tests, independent positive checks, and five meaningful mutant-red checks.
Its handoff writer alternates distinct large raw byte sequences across 200 replacements.
Every concurrent read returns one complete sequence.
Empty input, oversized input, total timeout, static bad paths, and pre-rename failures preserve the required state.
No temporary file remains after the checked atomic operations.
The secondary restores every original source byte before its final positive run.
V2 behavior passes. H-F01 and H-F02 remain open across the task.

## V3 actual transcript

The primary built the exact worker into its unique absolute capture binary.
Its SHA256 is `64274a9e887b239c24ed9768e5b04f2e6a6c2ca64762fba9c77abee48a50b711`.
Build information records the exact worker and `vcs.modified=false`.
The secondary production binary has the same hash.
Every command uses an explicit absolute marked fixture home.
`HIVE_TMP` names that fixture only.
`cli_checks.py` records eleven invocations in `cli-commands.json` with separate raw stream captures.

```text
created memory memory/base-is-dev.md
updated index memory/MEMORY.md
- [base-is-dev](base-is-dev.md): The base branch of hive is dev.
hive: frontmatter: missing description
exit 1
updated handoff handoff.md
Open decisions: none.
```

The successful writes exit 0. The invalid write exits 1 and creates no fact or index change.
The primary also reproduces all three findings with this same frozen binary.
V3 fixture cleanup remains uncompleted to preserve the fixture logs and captures.
No real memory or handoff file was read or changed by these commands.

## V4 fresh gate

The commands ran from the exact clean worker worktree.
A unique `TMPDIR` outside every marked home prevents absent-home tests from discovering the real home.
The primary capture `gate-tmpdir.txt` records its absolute path.

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run --allow-serial-runners ./...
go test -race -count=1 ./...
go test -run TestSTE -count=1 .
git diff --check c6a4d91d82042f8861a2aa4b722944272117fa23..HEAD
```

```text
commit 26b095f0b17b4fa509abe84263f0a22ad4dea515
$ gofmt -l $(git ls-files -co --exclude-standard "*.go")
exit 0
$ go vet ./...
exit 0
$ golangci-lint run --allow-serial-runners ./...
0 issues.
exit 0
$ go test -race -count=1 ./...
ok  	github.com/Harrison-Blair/hive	2.767s
ok  	github.com/Harrison-Blair/hive/cmd	1.026s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff/write	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory/write	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/version	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	1.133s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.017s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.051s
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/records	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.134s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	17.256s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.014s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.014s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.167s
ok  	github.com/Harrison-Blair/hive/internal/project	1.023s
ok  	github.com/Harrison-Blair/hive/internal/tui	1.016s
exit 0
$ go test -run TestSTE -count=1 .
ok  	github.com/Harrison-Blair/hive	0.120s
exit 0
$ git diff --check c6a4d91d82042f8861a2aa4b722944272117fa23..HEAD
exit 0
```

This actual uncached run closes the worker report's cached-race evidence limitation.
It does not close the independent acceptance findings.

## Source freeze and limits

The primary inspected the complete base-to-worker diff and actual helper callers.
No Cobra or command import enters an internal package.
No store, baseline, dependency, or unrelated task source changed.
The initial clean worktree and restored copy match all original tracked-file hashes.
`freeze-initial.json` has SHA256 `714cf644c049050987eece27fd465100813382fc244af7dae5b95d5e6fa986a2`.
`freeze-before-evidence.json` records zero differences.
This verdict is the only permitted tracked addition.
All worker V files remain exact.

The approved boundary is a local user and cooperating writers with static bad types and links.
Hostile replacement between `Lstat` and `Open` remains deferred for T01.
The checks close their owned pipes after timeout.
They do not claim cancellation of an arbitrary blocked reader.

The primary retains all owned CLI fixtures, binaries, reports, logs, and captures.
Go tests removed only their own temporary fixtures.
The secondary reports its earlier absent-home fixture error and preserves the resulting appended real log events.
Its corrected outside-home run passed. The primary did not repeat that error.
No credential contents, global harness settings, other worktrees, other agents, or historical attribution changed.
No agent started. No source implementation ran.

Uncompleted: the three owner fixes, a fresh independent check of those fixes, fixture cleanup, and later T07 landing.
The coordinator must not treat this FAIL as a whole-task PASS.

END VERIFIER T07-memory-and-handoff t07-verify-01

## 2026-10-18: t07-verify-02

| Item | Value |
| --- | --- |
| Role | Fresh independent source verifier, both areas |
| Actual route | `pi openai-codex/gpt-6.1-sol:high` |
| Session | `01a14e2c-793d-74da-815d-9ebb29b07988` |
| Tested final worker | `9eba57a39a7d7dadc6af05af654c1bfc008a341c`, clean |
| Last source commit | `41196cc3f7a1b4c3eb4a56f5a30168a43825929e` |
| Fix initial | `07bb763a8bbb685fc3b358bbd16f7aa2ee54716f` |
| Original worker | `26b095f0b17b4fa509abe84263f0a22ad4dea515` |
| Base | `c6a4d91d82042f8861a2aa4b722944272117fa23` |
| Whole-task verdict | **FAIL: H-F01 remains incomplete** |
| Verdict identity | Evidence-only child of the exact final worker. The report records its full commit after push. |

| Item | Proves | Verdict | Disposition |
| --- | --- | --- | --- |
| V1 | AC1, AC2 | **FAIL** | Existing-file mismatch output loses the required reason. The old memory checklist and M-F01 probes pass. |
| V2 | AC3 | **PASS** | Raw bytes, empty refusal, fixed limits, deadlines, atomic replacement, and static paths pass. |
| V3 | AC2, AC3 | **PASS for required transcript** | Extended CLI assertions fail on H-F01 at both scopes. |
| V4 | AC4 | **PASS** | Full uncached gate from the actual task worktree passes. |
| Cleanup | Operational completion | **PASS within the provenance limit below** | Listed old scratch and all new owned scratch are absent. |

Recommendation: return H-F01 to the worker with the exact existing-file output requirement.
Keep M-F01 and H-F02 closed. Do not land T07 from this verdict.
No implementation change ran in this pass.

### Reads, prerequisites, and full-range inspection

I read both complete prior verifier reports, all current V1 through V4, and the old tracked FAIL.
I read the original task and the unchanged current task.
I read both complete worker reports and the cleanup receipt, before/after paths, manifests, and provenance captures.
The fix report ends with `END WORKER T07-memory-and-handoff t07-fix-01`.

CP1 PASS remains accepted. I did not repeat CP1.
The installed-spec report's CAP/HOME LANDING PASS remains accepted within its stated limits.
Its C01 through C07 concern later tasks, not T07.
The cap and home merges are ancestors of the tested worker.
The task tracker records T22 landed. The listed T07 prerequisites pass.

I inspected the complete base-to-final-worker diff: 26 files, 3724 insertions, and one deletion.
The last source commit and the final worker differ only in V4 evidence.
The fix changes four production files and adds or extends five test files.
No store, home, common, output, records, dependency, schema, or lock contract changes.
The owned scope value adds `ProjectID` for checked metadata. No unrelated shared API changes.

Full-range changed files:

```text
cmd/handoff/handoff.go
cmd/handoff/write/write.go
cmd/handoff/write/write_cli_test.go
cmd/memory/memory.go
cmd/memory/write/write.go
cmd/memory/write/write_cli_test.go
cmd/root.go
internal/handoff/write.go
internal/handoff/write_test.go
internal/lib/scope/doc.go
internal/lib/scope/input.go
internal/lib/scope/scope.go
internal/lib/scope/scope_test.go
internal/memory/doc.go
internal/memory/frontmatter.go
internal/memory/frontmatter_test.go
internal/memory/index.go
internal/memory/index_test.go
internal/memory/write.go
internal/memory/write_process_test.go
internal/memory/write_test.go
spec/hive-v1/evidence/T07-memory-and-handoff/V1.md
spec/hive-v1/evidence/T07-memory-and-handoff/V2.md
spec/hive-v1/evidence/T07-memory-and-handoff/V3.md
spec/hive-v1/evidence/T07-memory-and-handoff/V4.md
spec/hive-v1/evidence/T07-memory-and-handoff/VERIFIED.md
```

Both commands use the bounded stdin helper. Handoff uses the same atomic writer as memory and index.
One lock contains all existing reads, the fact write, index construction, and the index write.
The callback returns nil on partial outcomes. Each command wrapper passes a nil Go error to `common.Run`.
No internal package imports Cobra or a command package.
Recommendation: retain these package, helper, and lock choices.

### Finding dispositions

| ID | Prior area | Disposition |
| --- | --- | --- |
| M-F01 | Primary memory | **CLOSED**. Malformed nonempty separators refuse for stdin and existing files without fact or index changes. |
| H-F01 | Secondary logging | **INCOMPLETE**. Both log paths are safe, but existing-file user mismatch text regresses. |
| H-F02 | Secondary logging | **CLOSED**. Applicable checked project events carry `project_id`. Home events add no project identity. |

#### H-F01-R1: existing-file user mismatch text

This regression is in the touched H-F01 error route, not a new probe class.
AC2 specifies the mismatch reason and the existing-file wrapper.
The current brief requires the user mismatch error to remain exact.

`internal/memory/frontmatter.go:15` makes `Error()` return only the safe category.
`internal/memory/index.go:70-72` formats that error with `%v`.
This discards the exact user message before `userErrorText` can select it.

A synthetic `bad.md` has valid frontmatter with a different name.
The actual CLI returns exit 1, phase `scan`, and no effects at home and project scope.
Its reason is `frontmatter: name differs`.
The required reason is `frontmatter: name T07V02_NAME_SENTINEL differs from bad`.
Both have the same required `memory file invalid: <absolute file>: ` prefix.
The actual and expected full strings are in `final-cli/findings.json`.

The old fact and index bytes remain exact.
Parsed name, description, and body sentinels stay absent from both error logs.
The real error events retain category, phase, command, and applicable project metadata.
The stdin user error remains exact.

Recommendation: preserve separate user and log text through the existing-file wrapper.
Keep the safe category and the wrapper prefix.
A bounded local wrapper is sufficient. No shared logging abstraction or lock change is necessary.
A new worker brief must authorize any needed scan-wrapper change.

### V1 and V2 reruns, red checks, and restoration

All capture names below resolve under:

`/home/penguin/source/hive/.hive/tmp/coordinator/reports/t07-verify-02-captures/`

| Capture | Actual command or check | Result |
| --- | --- | --- |
| `V1.txt` | `go test -race -count=1 ./internal/memory/ ./internal/lib/scope/ -v` | Exit 0. All worker cases pass. |
| `V2.txt` | `go test -race -count=1 ./internal/handoff/ -v` | Exit 0. All worker cases pass. |
| `historical-M-red.txt` | Separator probe on exact `07bb763a` source | Exit 1. Stdin and existing-file assertions fail. |
| `historical-H-red.txt` | Old logging probes on exact `07bb763a` source | Exit 1. Content leak and three missing project fields reproduce. |
| `independent-green.txt` | Both retained probe packages on final source | Exit 0. Both established checklists pass. |
| `existing-mismatch-historical.txt` | Added exact-user-and-safe-log probe on old source | Exit 1. User text passes, but the synthetic name reaches the log. |
| `existing-mismatch-source.txt` | Same probe on final source | Exit 1. Log safety passes, but exact user text fails. |
| `rename-fault-green.txt` | Separate external rename-fault build | Exit 0. Old regular memory and handoff bytes remain, without effects or temporary residue. |
| `restored-green.txt` | Old checklists after nine restored mutants | Exit 0. The separately executed known failing probe is excluded explicitly. |
| `final-independent.txt` | Both complete probe packages after restoration, including the added probe | Exit 1. Only the exact existing-file mismatch assertion fails. |

The final complete command was:

```sh
go test -race -count=1 ./internal/t07memoryverify ./internal/t07atomicverify -v
```

The original assertions remain present. The added probe covers only the touched existing-file error route.
Fixed expectations use 1048575, 1048576, 1048577, and 5000000 bytes, not production constants.
Both old checklists pass again, including these cases:

- Scopes, project location, names, frontmatter categories, valid-space trimming, literal quotes, colons, and 10000 unknown keys.
- Whole stdin and existing-file boundaries, active trickle, and total 100 ms deadlines.
- Stdin before acquisition, no lock during input wait, and all existing reads under one lock.
- A cooperating change after acquisition and ten pairs of OS processes with the exact sorted index.
- Static links, wrong types, unreadable files, exclusions, and prewrite index obstructions.
- Actual post-memory index rename faults, truthful created and updated partial effects, retained facts, and the old index.
- Later regeneration by a distinct write without replay. Nil Go errors remain at both callback boundaries.
- Exact raw handoff bytes, empty refusal, fixed limits, deadlines, complete concurrent reads, and no temporary residue.

The final atomic run completed 200 replacements and 540 exact-byte concurrent reads.
Active trickle returned at 100.039473 ms for memory and 100.21018 ms for handoff.

Nine established mutant classes failed on behavior assertions in isolated source:
lock placement, existing cap, partial outcome, sort order, atomic replacement, stdin cap, deadline, paths, and effect events.
Each has a patch and red-output capture. All original source bytes were restored.
No compilation-only failure counts as red.
The added exact-user-and-safe-log probe remains red. No whole-task restored green is claimed.

### V3 actual CLI

The production binary came from the clean actual worktree at the exact final worker.
Its build information records `vcs.modified=false`.
Its SHA256 is `f9c75d9530217129965ae835bf5decfb54263ea6e67b479a099349263a266a55`.
Every invocation used its unique absolute path and an explicit absolute marked fixture `--home`.
The fault binary was separate. It contained only the recorded external rename fault.

Observed required transcript:

```text
created memory memory/base-is-dev.md
updated index memory/MEMORY.md
- [base-is-dev](base-is-dev.md): The base branch of hive is dev.
hive: frontmatter: missing description
exit 1
updated handoff handoff.md
Open decisions: none.
```

Successful writes exit 0. The invalid write exits 1 without fact or index changes.
The final extended driver completed 56 actual invocations and exited 1 for the two H-F01-R1 mismatches.
It inspected text, JSON, file bytes, and synthetic persistent logs separately.
`final-cli/cli-commands.json`, raw stream files, and log captures retain the actual results.
No actual error event or completed-effect event was removed or fabricated.
Project memory, index, handoff, input, frontmatter, scan, and index-failure events have checked identity.
Home events have no invented project field.

The first driver draft incorrectly expected the safe category in existing-file user output.
Its exit-0 `cli-run.txt` is not acceptance evidence for exact user errors.
I corrected the literal expectation without changing source.
The final driver and the independent probe both detect H-F01-R1. Earlier captures remain for attribution.

### V4 actual gate

These commands ran from the exact task worktree root with an external `TMPDIR`:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run ./...
go test -race -count=1 ./...
go test -run TestSTE -count=1 .
git diff --check c6a4d91d..HEAD
```

Formatting printed nothing. Vet exited 0. Lint reported `0 issues.` and exited 0.
Every race-test package passed without cache. TestSTE passed. The diff check printed nothing.
The full actual outputs and exit statuses are in `gate-*.txt`.
Memory passed in 2.177 s, handoff in 1.144 s, and store in 17.286 s.
`TMPDIR` was `/tmp/t07-verify-02-gate.h4q2dw1z`, outside every marked home.
No absent-home test discovered the real home in this pass.

### Freeze, cleanup, preservation, and limits

Initial tracked-file freeze SHA256:

`1cc95840788d724625ef4d241195141251ea64407582201015cee45801506f55`

All source and V hashes matched before this appendix.
The old 290-line verdict remains an exact Git-object prefix.
All V files retain original worker bytes as prefixes.
This appendix is the only tracked change.

The worker receipt lists 20 obsolete scratch paths. Every exact listed path is absent.
Both preserved cleanup manifests pass. Old reports, outputs, patches, scripts, freezes, and log evidence remain unchanged during this pass.
No broad deletion or new removal of prior scratch ran here.
The unlisted empty secondary `tmp/` directory remains.

Retained probe SHA256 values:

```text
fa5bcf92b4049a2e2da126f58d135acd189d886199325bfe68f9e3a67fe6df11  probe-t07memoryverify_test.go
a9934aa68f557ff209823f263152f915790872ab4677d0d83c5e09746dc9d507  probe-t07atomicverify_test.go
6ab57e5aafda632e5af812ad3efee5654840022eafc006ae761e498789c03752  probe-mismatch_test.go
```

The first two match the worker receipt and retained source-copy provenance captures.
Replay copies matched these definitions byte for byte.
Prior tracked-source freezes do not include the added independent probes.
No separate pre-cleanup immutable hash receipt for those files was supplied.
Historical identity therefore relies on the retained cleanup provenance and captured assertions.
I do not claim a direct comparison with the removed old source copies.
Ownership of the 20 listed paths agrees with both old reports and the before-state receipt.

I removed both new source copies, both fixture sets, the binaries, and the external test container after compact captures.
Before removal, all original source bytes matched their exact Git objects.
Probes, synthetic logs, raw streams, patches, scripts, hashes, and binary metadata remain.
`cleanup.txt` records exact present and absent paths.
Real root and task log prefixes remain exact by hash-only comparison.
The earlier real-root setup append remains preserved. No rollback ran.

Two setup checks needed correction before probes ran: the receipt has 20 paths, and its manifests resolve from their parent directory.
Neither setup failure is a source finding or a meaningful red test.
No old capture changed.

Only the approved local-user and cooperating-writer boundary applies.
Hostile `Lstat`/`Open` replacement remains deferred for T01.
Timeout probes close their owned pipes. They do not prove arbitrary blocked-reader cancellation.
No credential content, global configuration, worktree, other agent, or historical attribution changed.
No agent started. No source fix, tracker edit, handoff edit, force push, or `main` change ran.
No co-author or false session trailer was added. No STE baseline increased.

Uncompleted: H-F01-R1, its fresh whole-task green check, and later T07 landing.
No other owned acceptance check remains unrun.

Full report and separate verdict commit receipt:

`/home/penguin/source/hive/.hive/tmp/coordinator/reports/t07-verify-02.md`

END VERIFIER T07-memory-and-handoff t07-verify-02

## 2026-10-18: t07-verify-03

| Item | Value |
| --- | --- |
| Role | Fresh independent source verifier, both areas |
| Actual route | `pi openai-codex/gpt-6.1-sol:high` |
| Actual provider/model/level | `openai-codex` / `gpt-6.1-sol` / `high` |
| Session | `01a14e40-f207-767a-83a1-88025b81e5fc` |
| Exact tested final worker | `7a1049489c1efc94a46ca69620d00f762f2f8eb7`, clean |
| Last source commit | `b47a5c39a2da081a9bd41a8ddf9cf1d7d99faa2f` |
| Second fix initial | `6bbc2a8fbe8063629a51698e12bba0e120c7eed4` |
| First fix initial | `07bb763a8bbb685fc3b358bbd16f7aa2ee54716f` |
| Original worker | `26b095f0b17b4fa509abe84263f0a22ad4dea515` |
| Base | `c6a4d91d82042f8861a2aa4b722944272117fa23` |
| Whole-task verdict | **PASS** |
| Verdict identity | Evidence-only child of the exact final worker. The separate report records its full commit after push. |

| Item | Proves | Verdict | Disposition |
| --- | --- | --- | --- |
| V1 | AC1, AC2 | **PASS** | Worker tests and both established memory checklists pass, including exact existing-file user errors. |
| V2 | AC3 | **PASS** | Worker tests and the established handoff checklist pass. |
| V3 | AC2, AC3 | **PASS** | Required transcript and all 60 extended actual CLI invocations pass. |
| V4 | AC4 | **PASS** | Full uncached gate passes from the actual task worktree root. |
| Cleanup | Operational completion | **PASS within the accepted provenance limit** | All old listed scratch stays absent. I removed all new manifest-listed scratch. |

Recommendation: accept source at the exact tested worker for the separate landing procedure.
No source implementation or task landing ran in this pass.

### Reads, prerequisites, and full-range inspection

I read all three complete prior verdict reports, all old tracked verdict bytes, and all current V files.
I read the original and current task, every worker report, and the exact cleanup receipt and captures.
The latest worker report ends with `END WORKER T07-memory-and-handoff t07-fix-02`.
The current task equals the original task byte for byte.
CP1 PASS remains accepted. I read its report and did not repeat CP1.
The installed-spec CAP/HOME LANDING PASS stays accepted within its stated limits.
Its later-contract C01 to C07 do not change T07.
Both cap and home merge commits are ancestors. The task tracker records T22 landed.
All listed T07 prerequisites pass.

I inspected the full base-to-final-worker diff: 26 files, 4304 insertions, and one deletion.
The final source change touches only `internal/memory/index.go` and the adjacent tests in `internal/memory/write_test.go`.
The exact final worker then appends only V1, V3, and V4 evidence.
Every old assertion stays unchanged.
The second fix changes no store, scope, input, handoff, atomic helper, command, schema, dependency, shared API, or lock contract.
Full-range changed files:

```text
cmd/handoff/handoff.go
cmd/handoff/write/write.go
cmd/handoff/write/write_cli_test.go
cmd/memory/memory.go
cmd/memory/write/write.go
cmd/memory/write/write_cli_test.go
cmd/root.go
internal/handoff/write.go
internal/handoff/write_test.go
internal/lib/scope/doc.go
internal/lib/scope/input.go
internal/lib/scope/scope.go
internal/lib/scope/scope_test.go
internal/memory/doc.go
internal/memory/frontmatter.go
internal/memory/frontmatter_test.go
internal/memory/index.go
internal/memory/index_test.go
internal/memory/write.go
internal/memory/write_process_test.go
internal/memory/write_test.go
spec/hive-v1/evidence/T07-memory-and-handoff/V1.md
spec/hive-v1/evidence/T07-memory-and-handoff/V2.md
spec/hive-v1/evidence/T07-memory-and-handoff/V3.md
spec/hive-v1/evidence/T07-memory-and-handoff/V4.md
spec/hive-v1/evidence/T07-memory-and-handoff/VERIFIED.md
```

The local wrapper keeps the cause instead of flattening it with `%v`.
Its `UserMessage` preserves exact user text. Its `LogCategory` and `Error` preserve safe log text.
Both existing-file scan error routes use the wrapper.
Recommendation: retain this bounded wrapper and the existing package structure.
One lock still contains all existing reads, the fact write, index construction, and the index write.
Both commands use the bounded stdin helper. Handoff uses the same atomic writer as memory and index.
The callback and CLI wrapper return nil Go errors for partial outcomes.
No internal package imports Cobra or a command package.

### Finding dispositions

| ID | Prior area | Disposition |
| --- | --- | --- |
| M-F01 | Primary memory | **CLOSED**. Invalid nonempty separators refuse from stdin and existing files without changing old facts or index. |
| H-F01 | Secondary memory logging | **CLOSED**. Both routes preserve exact user errors and exclude parsed values from real logs. |
| H-F01-R1 | H-F01 fix regression | **CLOSED**. Exact existing-file mismatch text passes at home and project scope, in text and JSON. |
| H-F02 | Secondary memory and handoff logging | **CLOSED**. Applicable checked project events carry identity. Home events add no project field. |

Required existing-file user text is `memory file invalid: <absolute bad.md>: frontmatter: name <sentinel> differs from bad`.
The actual logs instead contain `memory file invalid: <absolute bad.md>: frontmatter: name differs`.
The stdin user error remains exact. Its log keeps `frontmatter: name differs`.
No parsed name, description, body, or handoff sentinel reaches the checked logs.
Actual error and effect events retain category, phase, command, scope, and applicable project metadata.
Every event reflects an actual effect or error.

### Capture authority and complete run register

All capture names below resolve under:

`/home/penguin/source/hive/.hive/tmp/coordinator/reports/t07-verify-03-captures/`

`base-worker.diff`, `changed-files.txt`, and `call-sites.txt` retain the full-range inspection map.
The worktree stayed frozen. All probes, mutants, and fault hooks ran only in isolated copies.
Every Go run below uses `-race -count=1`, except the separately named TestSTE run.

| Capture | Actual command or check | Status |
| --- | --- | --- |
| `V1.txt` | `go test -race -count=1 ./internal/memory/ ./internal/lib/scope/ -v` from the actual worktree | Exit 0 |
| `V2.txt` | `go test -race -count=1 ./internal/handoff/ -v` from the actual worktree | Exit 0 |
| `historical-mismatch-red.txt` | Exact retained mismatch probe on `6bbc2a8f` | Exit 1, exact user assertion |
| `historical-M-red.txt` | Retained separator probe on `07bb763a` | Exit 1, stdin and existing-file assertions |
| `historical-H-red.txt` | Retained content-log and project-identity probes on `07bb763a` | Exit 1, content leak and three missing project fields |
| `independent-green.txt` | Both complete retained probe packages on `7a104948` | Exit 0, exact mismatch included |
| `rename-fault-green.txt` | `TestRenameFaultOldBytes`, external fault enabled in a separate build | Exit 0 |
| `mutant-lock-red.txt` | Existing-file scan before acquisition, `TestExistingReadsUnderLock` | Exit 1, early open and missed cooperating change |
| `mutant-existing-cap-red.txt` | 2 MiB existing cap, fixed 1048577-byte case | Exit 1, accepted oversized file |
| `mutant-partial-red.txt` | Index failure reports `ok`, `TestPartialAndHeldLock` | Exit 1, exact created and updated partial envelopes |
| `mutant-sort-red.txt` | Reverse sort, `TestTwoProcessesExactIndex` | Exit 1, exact sorted index |
| `mutant-atomic-red.txt` | Direct write, `TestRawAtomic200` | Exit 1, torn bytes |
| `mutant-stdin-cap-red.txt` | Read only the limit, `TestInputEntireStreamBoundaries` | Exit 1, accepted oversized input |
| `mutant-deadline-red.txt` | Add one second, `TestTotalTrickleDeadlineAndLock` | Exit 1, total deadline |
| `mutant-path-red.txt` | Bypass checks, `TestStaticWrongTypesAndLinks` | Exit 1, static path refusals |
| `mutant-log-red.txt` | Change effect event name, `TestLoggingEffectsAndActualErrors` | Exit 1, actual event identity |
| `restored-green.txt` | Both complete packages after restoration of every original source byte | Exit 0, no known finding excluded |
| `final-independent.txt` | Both complete packages in a final uncached run | Exit 0 |
| `cli-run.txt` | `python3 cli_checks.py`, 60 actual invocations | Exit 0 |
| `gate-format.txt` | Worktree formatting command | Exit 0, empty output |
| `gate-vet.txt` | `go vet ./...` | Exit 0 |
| `gate-lint.txt` | `golangci-lint run ./...` | Exit 0, `0 issues.` |
| `gate-race.txt` | `go test -race -count=1 ./...` | Exit 0, every package uncached |
| `gate-ste.txt` | `go test -run TestSTE -count=1 .` | Exit 0 |
| `gate-diff.txt` | Full base-to-worker diff check | Exit 0, empty output |

Every red above fails on behavior assertions, not compilation alone.
All nine established mutation patches and full outputs remain.
No new probe class ran. Original probe definitions and assertions stay byte-exact.
Both complete checklists pass again, including these cases:

- Scope and name rules, all frontmatter categories, valid-space trimming, literal quotes, colons, and 10000 unknown keys.
- Whole stdin and existing files at literal 1048575, 1048576, 1048577, and 5000000 bytes.
- Total 100 ms deadlines, active trickle, stdin before acquisition, and no lock during input wait.
- All existing reads under one held lock, a cooperating change, and ten OS-process pairs with the exact sorted index.
- Static links, wrong types, unreadable files, scan exclusions, and prewrite index obstructions.
- Actual post-memory index rename faults, truthful created and updated partial effects, retained facts, and the old index.
- Later regeneration by a distinct write without replay. Nil Go errors remain at both callback boundaries.
- Exact raw handoff bytes, empty refusal, complete concurrent reads, and no temporary residue.

The final handoff run completed 200 replacements and 533 concurrent exact-byte reads.
Active trickle returned at 100.029671 ms for memory and 100.173979 ms for handoff.
The complete package command ran three times: initial green, restored green, and final green.
Each ran `go test -race -count=1 ./internal/t07memoryverify ./internal/t07atomicverify -v`.
The explicit rename-fault test runs separately instead of relying on its normal no-hook skip.
The child-only skips are process entry points. Their parent tests actually run the OS processes.

### V1 actual worktree run

```sh
# commit 7a1049489c1efc94a46ca69620d00f762f2f8eb7
go test -race -count=1 ./internal/memory/ ./internal/lib/scope/ -v
echo "exit $?"
```

```text
=== RUN   TestParseFrontmatterValid
--- PASS: TestParseFrontmatterValid (0.00s)
=== RUN   TestParseFrontmatterKeepsColonInValue
--- PASS: TestParseFrontmatterKeepsColonInValue (0.00s)
=== RUN   TestParseFrontmatterRequiresColonSpace
=== RUN   TestParseFrontmatterRequiresColonSpace/missing_name
=== RUN   TestParseFrontmatterRequiresColonSpace/missing_description
=== RUN   TestParseFrontmatterRequiresColonSpace/both
--- PASS: TestParseFrontmatterRequiresColonSpace (0.00s)
    --- PASS: TestParseFrontmatterRequiresColonSpace/missing_name (0.00s)
    --- PASS: TestParseFrontmatterRequiresColonSpace/missing_description (0.00s)
    --- PASS: TestParseFrontmatterRequiresColonSpace/both (0.00s)
=== RUN   TestParseFrontmatterTrimsAfterColonSpace
--- PASS: TestParseFrontmatterTrimsAfterColonSpace (0.00s)
=== RUN   TestParseFrontmatterErrors
=== RUN   TestParseFrontmatterErrors/no_block
=== RUN   TestParseFrontmatterErrors/no_close
=== RUN   TestParseFrontmatterErrors/missing_name
=== RUN   TestParseFrontmatterErrors/missing_description
=== RUN   TestParseFrontmatterErrors/empty_name
=== RUN   TestParseFrontmatterErrors/empty_description
=== RUN   TestParseFrontmatterErrors/duplicate_name
=== RUN   TestParseFrontmatterErrors/duplicate_description
--- PASS: TestParseFrontmatterErrors (0.00s)
    --- PASS: TestParseFrontmatterErrors/no_block (0.00s)
    --- PASS: TestParseFrontmatterErrors/no_close (0.00s)
    --- PASS: TestParseFrontmatterErrors/missing_name (0.00s)
    --- PASS: TestParseFrontmatterErrors/missing_description (0.00s)
    --- PASS: TestParseFrontmatterErrors/empty_name (0.00s)
    --- PASS: TestParseFrontmatterErrors/empty_description (0.00s)
    --- PASS: TestParseFrontmatterErrors/duplicate_name (0.00s)
    --- PASS: TestParseFrontmatterErrors/duplicate_description (0.00s)
=== RUN   TestParseMemoryNameDiffers
--- PASS: TestParseMemoryNameDiffers (0.00s)
=== RUN   TestParseMemoryNameDifferenceStaysOnOneLine
--- PASS: TestParseMemoryNameDifferenceStaysOnOneLine (0.00s)
=== RUN   TestReadMemoryFileBoundary
--- PASS: TestReadMemoryFileBoundary (0.01s)
=== RUN   TestMemoryWriteChildProcess
    write_process_test.go:15: child helper for the two-process test
--- SKIP: TestMemoryWriteChildProcess (0.00s)
=== RUN   TestTwoProcessesWriteMemories
--- PASS: TestTwoProcessesWriteMemories (1.02s)
=== RUN   TestNewMemoryWritesFileAndIndex
--- PASS: TestNewMemoryWritesFileAndIndex (0.00s)
=== RUN   TestReplacementReportsUpdated
--- PASS: TestReplacementReportsUpdated (0.00s)
=== RUN   TestTwoMemoriesIndexNameOrder
--- PASS: TestTwoMemoriesIndexNameOrder (0.00s)
=== RUN   TestExistingInvalidFileStopsWrite
--- PASS: TestExistingInvalidFileStopsWrite (0.00s)
=== RUN   TestExistingOversizedFileStopsWrite
--- PASS: TestExistingOversizedFileStopsWrite (0.00s)
=== RUN   TestMemoryNameValidation
=== RUN   TestMemoryNameValidation/#00
=== RUN   TestMemoryNameValidation/Base
=== RUN   TestMemoryNameValidation/base_is
=== RUN   TestMemoryNameValidation/_base
=== RUN   TestMemoryNameValidation/base_
=== RUN   TestMemoryNameValidation/base/dev
=== RUN   TestMemoryNameValidation/..
=== RUN   TestMemoryNameValidation/%2e%2e%2f
=== RUN   TestMemoryNameValidation/quote'
=== RUN   TestMemoryNameValidation/#
=== RUN   TestMemoryNameValidation/?
=== RUN   TestMemoryNameValidation/%
=== RUN   TestMemoryNameValidation/&
=== RUN   TestMemoryNameValidation/;
=== RUN   TestMemoryNameValidation/+
=== RUN   TestMemoryNameValidation/lé
=== RUN   TestMemoryNameValidation/bad_name
--- PASS: TestMemoryNameValidation (0.00s)
    --- PASS: TestMemoryNameValidation/#00 (0.00s)
    --- PASS: TestMemoryNameValidation/Base (0.00s)
    --- PASS: TestMemoryNameValidation/base_is (0.00s)
    --- PASS: TestMemoryNameValidation/_base (0.00s)
    --- PASS: TestMemoryNameValidation/base_ (0.00s)
    --- PASS: TestMemoryNameValidation/base/dev (0.00s)
    --- PASS: TestMemoryNameValidation/.. (0.00s)
    --- PASS: TestMemoryNameValidation/%2e%2e%2f (0.00s)
    --- PASS: TestMemoryNameValidation/quote' (0.00s)
    --- PASS: TestMemoryNameValidation/# (0.00s)
    --- PASS: TestMemoryNameValidation/? (0.00s)
    --- PASS: TestMemoryNameValidation/% (0.00s)
    --- PASS: TestMemoryNameValidation/& (0.00s)
    --- PASS: TestMemoryNameValidation/; (0.00s)
    --- PASS: TestMemoryNameValidation/+ (0.00s)
    --- PASS: TestMemoryNameValidation/lé (0.00s)
    --- PASS: TestMemoryNameValidation/bad_name (0.00s)
=== RUN   TestVeryLongMemoryNameFailsWithoutPanic
--- PASS: TestVeryLongMemoryNameFailsWithoutPanic (0.00s)
=== RUN   TestStaticIndexDirectoryRefusedBeforeWrite
--- PASS: TestStaticIndexDirectoryRefusedBeforeWrite (0.00s)
=== RUN   TestStaticIndexLinkRefusedBeforeWrite
--- PASS: TestStaticIndexLinkRefusedBeforeWrite (0.00s)
=== RUN   TestStaticMemoryDirectoryLinkRefused
--- PASS: TestStaticMemoryDirectoryLinkRefused (0.00s)
=== RUN   TestStaticLeafLinkRefused
--- PASS: TestStaticLeafLinkRefused (0.00s)
=== RUN   TestPartialIndexForNewMemory
--- PASS: TestPartialIndexForNewMemory (0.00s)
=== RUN   TestPartialIndexForReplacement
--- PASS: TestPartialIndexForReplacement (0.00s)
=== RUN   TestMemoryOversizedInputWritesNothing
--- PASS: TestMemoryOversizedInputWritesNothing (0.00s)
=== RUN   TestMemoryTimeoutWritesNothing
--- PASS: TestMemoryTimeoutWritesNothing (0.10s)
=== RUN   TestWriteLogsEffectsWithoutContent
--- PASS: TestWriteLogsEffectsWithoutContent (0.00s)
=== RUN   TestInvalidSeparatorStdinStopsWrite
=== RUN   TestInvalidSeparatorStdinStopsWrite/missing_name
=== RUN   TestInvalidSeparatorStdinStopsWrite/missing_description
--- PASS: TestInvalidSeparatorStdinStopsWrite (0.00s)
    --- PASS: TestInvalidSeparatorStdinStopsWrite/missing_name (0.00s)
    --- PASS: TestInvalidSeparatorStdinStopsWrite/missing_description (0.00s)
=== RUN   TestExistingInvalidSeparatorStopsWrite
--- PASS: TestExistingInvalidSeparatorStopsWrite (0.00s)
=== RUN   TestWriteLogsSafeFrontmatterCategory
--- PASS: TestWriteLogsSafeFrontmatterCategory (0.00s)
=== RUN   TestExistingInvalidFileLogsNoParsedValues
--- PASS: TestExistingInvalidFileLogsNoParsedValues (0.00s)
=== RUN   TestExistingMismatchKeepsExactUserTextAndSafeLog
--- PASS: TestExistingMismatchKeepsExactUserTextAndSafeLog (0.00s)
=== RUN   TestProjectExistingMismatchKeepsExactUserTextAndSafeLog
--- PASS: TestProjectExistingMismatchKeepsExactUserTextAndSafeLog (0.00s)
=== RUN   TestWriteLogsProjectID
--- PASS: TestWriteLogsProjectID (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/memory	2.154s
=== RUN   TestResolveHomeScope
--- PASS: TestResolveHomeScope (0.00s)
=== RUN   TestResolveProjectScope
--- PASS: TestResolveProjectScope (0.00s)
=== RUN   TestResolveRejectsBadScopes
=== RUN   TestResolveRejectsBadScopes/#00
=== RUN   TestResolveRejectsBadScopes/project:
=== RUN   TestResolveRejectsBadScopes/project:HIVE
=== RUN   TestResolveRejectsBadScopes/project:a/b
=== RUN   TestResolveRejectsBadScopes/project:..
=== RUN   TestResolveRejectsBadScopes/project:%2e%2e%2f
=== RUN   TestResolveRejectsBadScopes/other
=== RUN   TestResolveRejectsBadScopes/_home
=== RUN   TestResolveRejectsBadScopes/HOME
=== RUN   TestResolveRejectsBadScopes/project:_hive
=== RUN   TestResolveRejectsBadScopes/project:hive_
--- PASS: TestResolveRejectsBadScopes (0.00s)
    --- PASS: TestResolveRejectsBadScopes/#00 (0.00s)
    --- PASS: TestResolveRejectsBadScopes/project: (0.00s)
    --- PASS: TestResolveRejectsBadScopes/project:HIVE (0.00s)
    --- PASS: TestResolveRejectsBadScopes/project:a/b (0.00s)
    --- PASS: TestResolveRejectsBadScopes/project:.. (0.00s)
    --- PASS: TestResolveRejectsBadScopes/project:%2e%2e%2f (0.00s)
    --- PASS: TestResolveRejectsBadScopes/other (0.00s)
    --- PASS: TestResolveRejectsBadScopes/_home (0.00s)
    --- PASS: TestResolveRejectsBadScopes/HOME (0.00s)
    --- PASS: TestResolveRejectsBadScopes/project:_hive (0.00s)
    --- PASS: TestResolveRejectsBadScopes/project:hive_ (0.00s)
=== RUN   TestResolveReportsMissingProject
--- PASS: TestResolveReportsMissingProject (0.00s)
=== RUN   TestResolveRejectsLinkedProjectDirectory
--- PASS: TestResolveRejectsLinkedProjectDirectory (0.00s)
=== RUN   TestCheckFileRejectsStaticDirectoryAndLink
--- PASS: TestCheckFileRejectsStaticDirectoryAndLink (0.00s)
=== RUN   TestCheckFileRejectsLinkedAncestor
--- PASS: TestCheckFileRejectsLinkedAncestor (0.00s)
=== RUN   TestCheckFileRejectsOutsideHome
--- PASS: TestCheckFileRejectsOutsideHome (0.00s)
=== RUN   TestCheckFileAllowsMissingPath
--- PASS: TestCheckFileAllowsMissingPath (0.00s)
=== RUN   TestReadInputAtLimit
--- PASS: TestReadInputAtLimit (0.01s)
=== RUN   TestReadInputAboveLimit
--- PASS: TestReadInputAboveLimit (0.00s)
=== RUN   TestReadInputTimeout
--- PASS: TestReadInputTimeout (0.10s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.119s
exit 0
```

#### Exact retained mismatch probe: historical red, whole-restored green

The probe SHA256 is `6ab57e5aafda632e5af812ad3efee5654840022eafc006ae761e498789c03752`.
Its source stays byte-exact. It uses a synthetic sentinel, not a credential.
Historical red, from the isolated exact `6bbc2a8f` copy:

```sh
go test -race -count=1 ./internal/t07atomicverify -run TestExistingMismatchExactUserAndSafeLog -v
echo "exit $?"
```

```text
=== RUN   TestExistingMismatchExactUserAndSafeLog
    mismatch_test.go:25: user error: got output.Envelope{Outcome:"failed", Phase:"scan", Effects:[]output.Effect{}, Error:"memory file invalid: /tmp/t07-verify-03-gate.lln8by69/TestExistingMismatchExactUserAndSafeLog87852488/001/memory/bad.md: frontmatter: name differs", Result:interface {}(nil)}; want exact "memory file invalid: /tmp/t07-verify-03-gate.lln8by69/TestExistingMismatchExactUserAndSafeLog87852488/001/memory/bad.md: frontmatter: name T07V02_EXISTING_NAME_NOT_A_CREDENTIAL differs from bad"
--- FAIL: TestExistingMismatchExactUserAndSafeLog (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/t07atomicverify	0.003s
FAIL
exit 1
```

The same probe passes in each complete green run on the exact final source.
`final-independent.txt` retains the full final output, including the mismatch PASS and both package PASS lines.
`restored-green.txt` records the whole-restored green after all nine mutants.
No source fix ran in any copy.
Observed final output selection, from the capture directory:

```sh
rg 'TestExistingMismatch|^ok|replacements|trickle timeout' final-independent.txt; echo "exit $?"
```

```text
ok  	github.com/Harrison-Blair/hive/internal/t07memoryverify	12.005s
    atomic_test.go:111: 200 replacements, 533 concurrent exact-byte reads
    atomic_test.go:187: active trickle timeout 100.029671ms
    atomic_test.go:187: active trickle timeout 100.173979ms
=== RUN   TestExistingMismatchExactUserAndSafeLog
--- PASS: TestExistingMismatchExactUserAndSafeLog (0.00s)
ok  	github.com/Harrison-Blair/hive/internal/t07atomicverify	2.122s
exit 0
```

### V2 actual worktree run

```sh
# commit 7a1049489c1efc94a46ca69620d00f762f2f8eb7
go test -race -count=1 ./internal/handoff/ -v
echo "exit $?"
```

```text
=== RUN   TestHandoffWriteCreatesAndReplaces
--- PASS: TestHandoffWriteCreatesAndReplaces (0.00s)
=== RUN   TestHandoffEmptyInputWritesNothing
--- PASS: TestHandoffEmptyInputWritesNothing (0.00s)
=== RUN   TestHandoffOversizedInputKeepsOld
--- PASS: TestHandoffOversizedInputKeepsOld (0.00s)
=== RUN   TestHandoffTimeoutKeepsOld
--- PASS: TestHandoffTimeoutKeepsOld (0.10s)
=== RUN   TestHandoffStaticDirectoryRefused
--- PASS: TestHandoffStaticDirectoryRefused (0.00s)
=== RUN   TestHandoffStaticLinkRefused
--- PASS: TestHandoffStaticLinkRefused (0.00s)
=== RUN   TestHandoffConcurrentReplacement
--- PASS: TestHandoffConcurrentReplacement (0.01s)
=== RUN   TestHandoffLogsProjectID
--- PASS: TestHandoffLogsProjectID (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/handoff	1.118s
exit 0
```

### V3 actual CLI

The production binary came from the exact clean final worker with `vcs.modified=false`.
Its SHA256 is `cb01fd0e93e9e05ab7e4ca1e1949fc9756b0ebabba40cf44acad0cf4536214db`.
Its unique absolute path was `t07-verify-03-captures/bin/hive-t07-verify-03` under the absolute capture directory above.
The separate fault binary contains only the recorded external rename fault.
Its SHA256 is `aa25147de83a1d4ae1886d0a78923aec75eb0bf3067ccde2ee9c5c12cc6e3e11`.
The isolated build inherited the enclosing root's Git stamp, not the task revision.
Exact archived source equality and the retained fault patch establish its source identity. The production stamp is exact.
Every invocation uses an explicit absolute marked fixture home and an absolute binary.
`cli/cli-commands.json` records each command, status, stdout, and stderr.
Separate raw streams and synthetic logs remain.
The exact driver command from the actual task worktree root was:

```sh
python3 /home/penguin/source/hive/.hive/tmp/coordinator/reports/t07-verify-03-captures/cli_checks.py
echo "exit $?"
```

```text
Completed: 60 CLI invocations. V3, text, JSON, registration, raw bytes, errors, external rename faults, partial/recovery and logs checked.
PASS: all literal expectations
exit 0
```

Observed required transcript in `cli/V3-actual-transcript.txt`:

```text
created memory memory/base-is-dev.md
updated index memory/MEMORY.md
- [base-is-dev](base-is-dev.md): The base branch of hive is dev.
hive: frontmatter: missing description
exit 1
updated handoff handoff.md
Open decisions: none.
```

Successful writes exit 0. The invalid memory write exits 1 without a fact or index change.
The extended driver checks exact existing-file and stdin mismatch errors at both scopes, in text and JSON.
It checks safe log categories separately from user text, JSON envelopes, and file bytes.
Home logs add no project field. Checked project memory/index/handoff success and failure logs carry `project_id: demo`.
Actual index rename failures occur after successful static checks and fact writes.
Created and updated partial effects remain truthful. No false index success occurs.
The old index remains. A distinct later write regenerates all retained facts without replay.
The cleanup block below records removal of only this pass's fixture roots after compact logs and outputs.

### V4 actual uncached gate

The following commands ran from the actual task worktree root at exact `7a104948`.
`TMPDIR=/tmp/t07-verify-03-gate.lln8by69` was outside every marked home.
No absent-home test discovered the real home.

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run ./...
go test -race -count=1 ./...
go test -run TestSTE -count=1 .
git diff --check c6a4d91d82042f8861a2aa4b722944272117fa23..HEAD
```

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
exit 0
$ go vet ./...
exit 0
$ golangci-lint run ./...
0 issues.
exit 0
$ go test -race -count=1 ./...
ok  	github.com/Harrison-Blair/hive	2.769s
ok  	github.com/Harrison-Blair/hive/cmd	1.031s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.180s
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.183s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/version	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	1.139s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.022s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.061s
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/records	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.141s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	17.312s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.017s
ok  	github.com/Harrison-Blair/hive/internal/logs	5.531s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.176s
ok  	github.com/Harrison-Blair/hive/internal/project	1.030s
ok  	github.com/Harrison-Blair/hive/internal/tui	1.022s
exit 0
$ go test -run TestSTE -count=1 .
ok  	github.com/Harrison-Blair/hive	0.120s
exit 0
$ git diff --check c6a4d91d82042f8861a2aa4b722944272117fa23..HEAD
exit 0
```

Formatting prints nothing. Vet, lint, all uncached race packages, TestSTE, and the full-range diff check pass.
Full original per-command outputs remain in `gate-*.txt`.

### Freeze, cleanup, preservation, and limits

Initial tracked freeze SHA256: `367de307b69ea277a1ed74aa68609952aff29528bf4753c887baf7b8f6c1142a`.
Every initially tracked byte matched before this appendix. Every old verdict byte stays an exact prefix.
All V files stay exact. This appendix is the only tracked change.
After restoration, the current and historical copies each match all 175 original files at their exact Git commits.
The first-verdict copy matches all 173 original files at `07bb763a`.
All probe copies match these retained definitions byte for byte:

```text
fa5bcf92b4049a2e2da126f58d135acd189d886199325bfe68f9e3a67fe6df11  probe-t07memoryverify_test.go
a9934aa68f557ff209823f263152f915790872ab4677d0d83c5e09746dc9d507  probe-t07atomicverify_test.go
6ab57e5aafda632e5af812ad3efee5654840022eafc006ae761e498789c03752  probe-mismatch_test.go
```

The prior cleanup audit remains PASS within its accepted provenance limit.
All 20 exact old listed scratch paths remain absent. Both retained hash manifests pass.
Historical probe identity relies on retained provenance and captures, not a new comparison with removed historical copies.
The unlisted empty secondary `tmp/` stays. I removed no prior scratch.
All 541 old report/capture files match this pass's initial hash manifest.
The real-root log stays exactly 3424 bytes. Its old prefix and whole-file hash stay exact.
No task log existed at the initial freeze. The earlier real-root setup append stays preserved without rollback.

Own cleanup command from the actual task worktree root:

```sh
python3 /home/penguin/source/hive/.hive/tmp/coordinator/reports/t07-verify-03-captures/cleanup.py
```

`cleanup.txt` records source/probe equality, old evidence preservation, and every exact present/removed/absent path.
Only these six roots appear in `scratch-manifest.json`:

```text
t07-verify-03-captures/source
t07-verify-03-captures/historical
t07-verify-03-captures/original
t07-verify-03-captures/bin
t07-verify-03-captures/cli/fixtures
/tmp/t07-verify-03-gate.lln8by69
```

The first five are below the absolute capture parent above. Every listed root is absent after removal.
Compact probes, outputs, synthetic logs, scripts, patches, binary metadata, and before/after manifests remain.
No broad deletion ran. No real memory, handoff, registry, or log received a CLI write from this pass.

Only the approved local-user and cooperating-writer boundary applies.
Hostile replacement between `Lstat` and `Open` stays deferred for T01.
Timeout probes close their owned pipes. They do not prove cancellation of an arbitrary blocked reader.
No credential content, global harness configuration, historical attribution, worktree, or other agent changed.
No agent started. No tracker edit, handoff edit, force push, main change, co-author, or false session trailer occurred.
No STE baseline increased.
Uncompleted: the separate T07 landing. No owned acceptance check remains unrun.

Full report and separate verdict commit receipt:

`/home/penguin/source/hive/.hive/tmp/coordinator/reports/t07-verify-03.md`

END VERIFIER T07-memory-and-handoff t07-verify-03
