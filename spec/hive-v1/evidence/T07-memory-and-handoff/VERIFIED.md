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
