# T03-task-lifecycle VERIFIED

| Item | Value |
| --- | --- |
| Pass | `t03-verify-01` |
| Date | 2026-10-18 |
| Agent | verifier, pi, `openai-codex/gpt-6.1-sol:high` |
| Tested worker | `638c40e339abb329912d7c09b06d609ddc1c69b3` |
| Base | `84da9819b31be83dbe71c36d620eadf3749e1f9f` |
| Source verdict | **FAIL** |
| Separate SPEC verdict | **PASS**, `7ae3fb8affb7bc577ec1a3aa248d416be2bde340` |

The provider, model, and reasoning level match the `PI_*` values.
This verdict binds the exact worker commit.
The worker report ends with `END WORKER T03-task-lifecycle t03-impl-01`.

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | FAIL | Required tests pass. A computed brief ancestor link permits an external write. |
| V2 | AC2, AC3 | PASS | All 25 pairs pass independently. Last verification replacement passes. |
| V3 | AC1 to AC4 | FAIL | Required transcript passes. Plain show omits a field. One invalid input splits the diagnostic. |
| V4 | AC5 | PASS | Empty format list, vet, lint, and uncached race tests pass. |

The logging defect also blocks acceptance.
The shared store defect remains a separate prerequisite failure.
A full gate PASS does not remove either defect.

## Scope and source inspection
I read the instructions, both glossaries, the STE skill, and the untrusted-input reference.
I read T03, the architecture, the tracker procedure, V1 to V4, and the final worker report.
I inspected the complete base-to-worker diff and every changed source file, import, and test.
The 28 changed paths fit the task scope.
No dependency, baseline, glossary, or shared store source changed.

The Cobra constructors contain flag reads, internal calls, and rendering only.
The root change adds one import and one command registration.
The public record has the listed fields and the approved optional `cancel_reason` string.
`Verification` points to the LAST result. It is not a history list.
The worktree and delivery pointers remain placeholders for T09 and T12.

The approved boundary is a local user and a cooperating writer.
Hostile replacement between `Lstat` and `Open` remains deferred for T01.
No hostile replacement probe ran.

## Runs

Captures are at `.hive/tmp/coordinator/reports/t03-verify-01-captures/` in the root home.
Each text capture has fewer than 200 lines.
### V1 and V2

```sh
go test -race -count=1 ./internal/task/ -run 'TestCreate|TestSlug|TestCycle' -v
go test -race -count=1 ./internal/task/ -run 'TestTransition|TestVerify' -v
```

```text
V1: PASS, exit 0, internal/task 1.117s
V2: PASS, exit 0, internal/task 1.037s
```

`v1.txt` and `v2.txt` hold the exact outputs.
The brief tests use literal sizes 1048575, 1048576, 1048577, and 5242880.
The actual CLI independently accepts the first two sizes and rejects the other two.
Rejected input takes no counter value.

The independent CLI table checks eight valid pairs and 17 invalid pairs.
Each invalid pair leaves the task record byte-identical.
Wrong actor, unknown caller, self-verification, and missing failure reason checks also preserve the record.
A later pass replaces the failed verification object.
Cancel preserves that last object. An empty cancel reason stays absent from JSON.

### Red evidence and test strength

Isolated Git snapshots at `fa55779d` and `8bc39389` compile and fail on assertions.
The original tests detect the empty creation and transition functions.
A separate current snapshot rejects both controlled mutations:

```text
brief cap changed to 2 MiB: exit 1, 1048577-byte rejection assertion fails
created -> complete permitted: exit 1, invalid-transition assertion fails
```

No actual source or test changed.
`red-and-deadline.txt` holds short excerpts.

An added test in the isolated copy uses the exported `Create` dependency.
It passes a 100 millisecond deadline while a pipe supplies a byte every ten milliseconds.
The call returns after about 100 milliseconds, with no counter, task, or effect.
A separate lock acquisition succeeds during the input wait.
Thus the deadline is total, and the input wait holds no store lock.
The production default remains 30 seconds.

### V3 and independent CLI probes

```sh
go build -o "$CAPTURES/hive-t03-verify-01" .
python3 "$CAPTURES/probe.py"
```

The unique fixture homes have explicit markers and absolute `--home` arguments.
`cli.txt` holds the required transcript and each exit status.
Its IDs, ready lists, failed verification, second completion, and final pass match T03.
Raw `show --brief` returns exactly `Second step`, without an added newline.
That behavior meets the contract.

Additional passing checks:

- A 40-character slug gives a 42-character ID through create, show, assign, and dependency selection.
- Invalid slugs, missing dependencies, duplicate dependencies, and cross-project dependencies fail before the counter changes.
- A seeded cycle gives `dependency cycle: 1-a -> 2-b -> 1-a` without a write.
- Missing, unreadable, directory, link, FIFO, and socket brief inputs fail before the counter changes.
- Static links include valid, missing, directory, device, and chained targets.
- Empty files and opaque brief bytes remain valid input.
- Both partial prefixes report the actual phase and each completed effect.
- Counter gaps and orphan briefs remain. A retry gets a new ID.
- List ignores an orphan brief. Show of its missing task reports `task not found`.
- Plain list and show succeed while another writer holds the lock.
- A ready query waits for the held lock and then succeeds.
- Ten pairs of concurrent creates give distinct IDs and a final counter of 20.

Source inspection checks that ready-list dependency reads use the same locked handle.
An unreadable or invalid stored dependency excludes the task as not verified.
This choice does not admit unsafe work. It is not a separate selection defect.
It does not excuse the missing error log events.

### V4

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run --allow-serial-runners ./...
go test -race -count=1 ./...
```

```text
gofmt: empty output, exit 0
vet: exit 0
lint: 0 issues, exit 0
race: all packages pass, exit 0
```

`gate.txt` holds the exact output. The source gate ran at the exact worker HEAD.

## Owned T03 findings
### F1: computed brief ancestor link permits an external write

Owner: T03. Priority: high.
`internal/task/create.go:296-329` checks only the brief leaf after `MkdirAll`.
A static `home/.hive/state/briefs` link names a directory outside the fixture home.
Actual `task create` returns `ok` and writes `1-seed.md` in that external directory.
The counter and task record also commit.
This failure needs no hostile replacement race.

A wrong-type brief ancestor and linked, directory, FIFO, or socket leaves fail as expected.
Recommendation: check each computed brief ancestor before creating or reading below it.
Keep the existing atomic write and partial-effect order.
Add a static external-ancestor regression for creation and stored brief reads.

### F2: task transitions and errors have no required log events

Owner: T03. Priority: high.
The actual transcript produces 34 task events. Every message is command start or command end.
None carries `task_id`, `project_id`, `agent_id`, or `error`.
Successful assign, complete, verify, and cancel have no state-transition event.
Failed self-verification and invalid transitions have no error event.
`cmd/common/common.go:62-74` logs only command boundaries and outcome.
The owned task packages add no operation log events.

Recommendation: log each committed task transition and each task error with applicable identifiers.
Keep business logic outside Cobra. Add actual CLI log assertions.
Do not log full brief bodies or unbounded user fields.
`cli-events.jsonl` contains the observed fixture events.

### F3: plain show omits the cancellation reason

Owner: T03. Priority: medium.
After `cancel --reason "stop reason"`, JSON show preserves `cancel_reason`.
Plain show omits the field and its value.
`internal/task/render.go:38-59` never renders `CancelReason`.
AC4 requires each field. D1 supplies the approved field contract.
Recommendation: render the optional field with single-line text and add a command test.

### F4: an invalid task ID splits the plain error line

Owner: T03. Priority: medium.
`internal/task/show.go:26` and `transition.go:117` append the raw invalid ID.
Actual show with an ID containing LF exits 1 with:

```text
hive: task not found: bad
second-line
```

The untrusted-input contract requires one plain diagnostic line.
Recommendation: quote control characters at each owned diagnostic boundary.
Apply the same rule to the wrong-actor error at `transition.go:85`.
Add text-mode assertions for LF and other control characters.

## Shared prerequisite G1: successful writes can exceed the read cap

Owner: T22 follow-up, not a T03 title-limit change. Priority: high.
Actual CLI creation uses an explicit slug and a 120000-byte control-character title.
The title encodes to about 720000 bytes. Creation succeeds.
A later failed verification adds a separate 120000-byte control-character reason.
It returns `ok` and leaves a 1440371-byte task record.
Both arguments fit the Linux per-argument limit.
Show and list then fail with `record too large`.

`store.writeRecordFile` checks JSON shape but not encoded size.
D2 assigns the shared check to T22 before `atomicWrite`.
Recommendation: reject oversized shared writes while preserving old bytes and earlier completed effects.
Do not add a T03 title limit. Recheck T03 against the landed store follow-up.

## Separate urgent SPEC check
SPEC PASS binds Git commit `7ae3fb8affb7bc577ec1a3aa248d416be2bde340`, parent `56663839`.
The final report ends with `END WAVE2 BOUNDARY SPEC 01`.
The check read Git objects, not concurrent root edits.
Exactly the six approved spec paths changed. No source or baseline changed.

D1 records the optional reason field and preserves verification.
D2 defines the shared encoded-write cap, final LF, fixed boundary tests, and old-byte preservation.
D3 changes code-search instructions only. Collision policy remains open for source review.
D4 records the prior no-tools model argument as `:medium`, not observed runtime thinking.
D5 permits doctor metadata effects only and counts complete request/reply frames, including LF.
No T05 source implementation check ran here.

The three archived worker reports match their frozen sources, byte counts, and hashes.
T03/T04/T05 evidence files exist at the named commits.
T22 reopens for its follow-up. T07 stays blocked. Other history and attribution remain.
The four saved dispatch briefs match their exact payload bytes and hash receipts.
R01 matches the 12580-byte historical handoff tail and its recorded hash.
R02 explanatory state keeps the missing-file observation and adds the approved medium-argument correction.
The old reports keep their bytes.
B01 now separates local `main` `36921626` from remote `main` `928cfa95`.
Direct reference checks match those distinct values. No reference moved.

The pinned spec passes uncached `TestSTE` and `git diff --check`.
New prose passes the external linter with zero hard findings.
Unchanged status-table labels retain their old linter findings. The STE baseline did not rise.
This SPEC PASS accepts contract text and record corrections only. It accepts no source or user approval.

## Cleanup and uncompleted work
I removed all own fixture homes, the unique binary, and all isolated source copies.
Existing runtime, reports, logs, credentials, branches, and worktrees remain.
I inspected no credential content. I changed no global harness setting.
Only this verdict is an authorized tracked edit.

Uncompleted: F1 to F4, the T22 write-cap fix, later source verification, and landing.
No source correction or merge occurred in this pass.

## 2026-10-18: current source verdict, t03-verify-02

| Item | Value |
| --- | --- |
| Pass | `t03-verify-02` |
| Role | Fresh independent source verifier |
| Route | pi, `openai-codex/gpt-6.1-sol:high` |
| Tested worker HEAD | `ba91e8a750fa6639781a9717550161923e02398f` |
| Source commit | `e8600ab15f7d5d46d27d56fddf4d65a0f2846c13` |
| Original base | `84da9819b31be83dbe71c36d620eadf3749e1f9f` |
| Fix initial HEAD | `d2664fc8e987aaa6c9c0760953da30b981b9c6ef` |
| Landed prerequisite | `c6a4d91d82042f8861a2aa4b722944272117fa23` |
| Branch | `task/T03-task-lifecycle` |
| Overall source verdict | **PASS** |
| Separate verdict commit | Evidence-only child of the tested worker. The root report records its exact ID. |

The initial worktree was clean at the exact tested worker HEAD.
The remote task branch matched that HEAD.
The Go source at that HEAD equals source commit `e8600ab1`.
`PI_PROVIDER` and `PI_MODEL` give `openai-codex` and `gpt-6.1-sol`.
The session event records thinking level `high`.
`route.json` records the route and the independent session ID.

I read the complete prior FAIL verdict and root `t03-verify-01.md`.
I read the original installed T03 task, current T03 task, all V evidence, and the complete fix report.
The fix report ends with `END WORKER T03-task-lifecycle t03-fix-01`.
The prerequisite report ends with `END CAP HOME LAND 01`.
Both required markers are present.
The historical 11832-byte FAIL verdict remains an exact prefix.
Its SHA256 is `99c41008a326088f63595eff700c2b4aa99797267945c5764dad91bd736d1321`.
The original V1 to V4 bytes also remain exact prefixes.

### Separate verification items

| Item | Proves | Verdict | Current evidence |
| --- | --- | --- | --- |
| V1 | AC1 | PASS | Creation suite, literal bounds, static paths, dependencies, partial prefixes, and concurrent creates pass. |
| V2 | AC2, AC3 | PASS | All 25 pairs, caller errors, self-verification, and last verification replacement pass. |
| V3 | AC1 to AC4 | PASS | Actual CLI transcript, safe cancellation text, diagnostics, queries, and JSON pass. |
| V4 | AC5 | PASS | Empty format output, vet, lint, and the full uncached race gate pass. |

### Finding dispositions

| ID | Owner | Disposition | Independent result |
| --- | --- | --- | --- |
| F1 | T03 | CLOSED | Computed brief ancestors reject static links and wrong types. Stored brief reads reject linked ancestors and invalid leaves. |
| F2 | T03 | CLOSED | Actual CLI events include effects, committed transitions, and errors with applicable IDs. |
| F3 | T03 | CLOSED | Plain show prints safe `cancel_reason`. Empty JSON omission and unchanged verification pass. |
| F4 | T03 | CLOSED | Control-bearing IDs, actors, slugs, dependencies, and brief paths retain one physical diagnostic line. |
| G1 | T22 prerequisite | CLOSED in T03 integration | Oversized encoded verification fails and preserves every prior state byte. Show and list remain readable. |

These are the original primary finding IDs.
No separate secondary T03 finding set was supplied.
No new finding or unrelated probe class applies.
The prior separate SPEC PASS remains historical. This pass does not repeat that separate spec check.

### Full-range inspection and prerequisite identity

I inspected the complete `84da9819..ba91e8a7` diff, not only the final commit.
I also inspected the complete `d2664fc8..ba91e8a7` fix diff.
The full range changes 49 paths. The fix range changes 13 paths.
The full-range paths are:

```text
cmd/project_static_home_test.go
cmd/root.go
cmd/task/assign/assign.go
cmd/task/cancel/cancel.go
cmd/task/complete/complete.go
cmd/task/create/create.go
cmd/task/list/list.go
cmd/task/show/show.go
cmd/task/task.go
cmd/task/verify/verify.go
cmd/task_lifecycle_test.go
internal/lib/home/home.go
internal/lib/home/home_test.go
internal/lib/records/doc.go
internal/lib/records/task.go
internal/lib/store/store.go
internal/lib/store/write_limit_test.go
internal/task/assign.go
internal/task/brief_path_test.go
internal/task/cancel.go
internal/task/complete.go
internal/task/create.go
internal/task/create_test.go
internal/task/diagnostic_test.go
internal/task/doc.go
internal/task/list.go
internal/task/list_show_test.go
internal/task/log.go
internal/task/render.go
internal/task/show.go
internal/task/transition.go
internal/task/transition_test.go
internal/task/verify.go
spec/hive-v1/03-architecture.md
spec/hive-v1/05-tracker.md
spec/hive-v1/evidence/T01-home-and-register/STATIC-HOME-FOLLOWUP.md
spec/hive-v1/evidence/T01-home-and-register/STATIC-HOME-VERIFIED.md
spec/hive-v1/evidence/T03-task-lifecycle/V1.md
spec/hive-v1/evidence/T03-task-lifecycle/V2.md
spec/hive-v1/evidence/T03-task-lifecycle/V3.md
spec/hive-v1/evidence/T03-task-lifecycle/V4.md
spec/hive-v1/evidence/T03-task-lifecycle/VERIFIED.md
spec/hive-v1/evidence/T22-store-operations/V2.md
spec/hive-v1/evidence/T22-store-operations/V5.md
spec/hive-v1/evidence/T22-store-operations/VERIFIED.md
spec/hive-v1/tasks/T03-task-lifecycle.md
spec/hive-v1/tasks/T04-roles-and-inbox.md
spec/hive-v1/tasks/T05-herdr-client-and-doctor.md
spec/hive-v1/tasks/T22-store-operations.md
```

The imported home, store, and spec paths are not T03-owned changes.
The five cap paths equal verified tip `30ef7e2d` and landed `dev` prerequisite `c6a4d91d`.
The five home paths equal verified tip `938d8737` and the same landed prerequisite.
The equality checks include their source, tests, and evidence.
`c6a4d91d` is an ancestor of the tested worker.
The fix changes no shared home, store, common, output, module, glossary, or STE baseline file.
Cobra wrappers remain thin. The record keeps one verification pointer, not a history list.
Worktree and delivery remain placeholders for T09 and T12.

### V1 and V2 runs

Commands from the task worktree root:

```sh
go test -race -count=1 ./internal/task/ -run 'TestCreate|TestSlug|TestCycle' -v
go test -race -count=1 ./internal/task/ -run 'TestTransition|TestVerify' -v
```

```text
V1: exit 0
ok  	github.com/Harrison-Blair/hive/internal/task	1.142s
V2: exit 0
ok  	github.com/Harrison-Blair/hive/internal/task	1.035s
```

`v1.txt` and `v2.txt` hold the complete outputs.
The actual CLI independently checks eight valid pairs and 17 invalid pairs.
Each invalid pair leaves the task bytes unchanged.
Wrong actor, unknown caller, self-verification, and missing failure reason also preserve the task bytes.
The failed verification returns to `assigned`. A later pass replaces it.
Cancel keeps the complete prior verification object.

### Red assertions and restored green runs

All source copies came from Git archives in the unique capture directory.
The actual task source stayed frozen.
The historical empty-symbol snapshots `fa55779d` and `8bc39389` compile and fail on assertions.
The fix test commit `25032f83` also compiles and fails on the F1 to F4 assertions.
The same current test selections pass.

The independent `repairs.py` CLI assertions fail separately for F1, F2, F3, and F4 against `25032f83`.
The same script passes all four groups against the current binary.
The G1 preservation assertion fails against original worker `638c40e3` and passes against the current binary.
These are semantic failures, not compile-only failures.

Controlled mutations in an isolated current copy also fail:

| Mutation | Observed assertion |
| --- | --- |
| Increase brief cap to 2 MiB | 1048577-byte rejection fails. |
| Permit created-to-complete | The invalid-transition assertion fails. |
| Remove `omitempty` | Empty `cancel_reason` appears in JSON. |
| Delay the total deadline tenfold | The 100 millisecond deadline assertion fails. |
| Remove verification during cancel | The unchanged-pointer assertion fails. |

Each mutation was restored. Each corresponding restored selection passes.
`red-and-deadline.txt` records these results.
The independent trickle check returns after about 100 milliseconds and acquires the lock during the input wait.
It consumes no counter and creates no task or effect.
The production default stays 30 seconds.

### V3 actual CLI and complete prior checklist

I built the binary from the actual task worktree root:

```sh
go build -o /home/penguin/source/hive/.hive/tmp/coordinator/reports/t03-verify-02-captures/hive-t03-verify-02 .
python3 /home/penguin/source/hive/.hive/tmp/coordinator/reports/t03-verify-02-captures/probe.py
python3 /home/penguin/source/hive/.hive/tmp/coordinator/reports/t03-verify-02-captures/repairs.py
python3 /home/penguin/source/hive/.hive/tmp/coordinator/reports/t03-verify-02-captures/g1.py
```

All four commands exit 0.
Each fixture has an explicit marker, a unique Git project, and an absolute `--home` argument.
`cli.txt` records every required transcript command, stdout, stderr, and exit status.
The two IDs, both ready lists, failed verification, second completion, and final verification match the task.
Bad slug, self-verification, repeated verification, and cancel of verified return exit 1 with the required errors.
Raw brief output is exactly `Second step`, with no added newline.
JSON show holds the last passing verification by `verifier-repo-one-2`.

The complete source checklist from the prior verdict passes again:

- The home counter spans two projects. A 40-character slug gives a usable 42-character task ID.
- Invalid slugs and missing, duplicate, invalid, or cross-project dependencies consume no counter.
- A seeded cycle gives exactly `dependency cycle: 1-a -> 2-b -> 1-a` without a write.
- Literal brief sizes 1048575 and 1048576 pass. Sizes 1048577 and 5242880 fail before counter use.
- Missing, unreadable, directory, linked, FIFO, and socket input briefs fail before counter use.
- Static input links cover valid, missing, directory, device, and chained targets.
- Empty and opaque briefs remain valid. The read does not parse the brief body.
- Counter-only and counter-plus-brief prefixes report the actual phase and each completed effect.
- Counter gaps and orphan briefs remain. List ignores an orphan brief. Retry gets a new ID.
- Show of the missing task fails with `task not found`. Raw brief output stays exact.
- Plain show and list remain lock-free. Ready waits for the held lock and then succeeds.
- Ready reads its task list and dependencies through one locked handle.
- Missing or invalid stored dependencies remain not ready. The approved behavior stays unchanged.
- Query filters and project/number order pass the existing tests in the uncached gate.
- Ten concurrent create pairs produce 20 distinct task IDs and final counter 20.

F1 adds 12 ancestor type checks across `.hive`, state, and briefs.
Static links, regular files, FIFOs, and sockets fail before counter or task effects.
Six stored-brief ancestor/leaf cases fail without returning external text or changing the external fixture.
User-selected briefs below a linked directory remain supported. A legitimate home-root alias also works.
Creation with a wrong-type leaf retains the required V1 counter-only partial outcome.
No outside brief write or task commit occurs. This preserves the original partial-state policy.

F2 checks six successful transitions with their applicable task, project, and actor IDs.
Self-verification emits one error and no transition. Both partial prefixes log only completed effects.
The required transcript now contains 51 task-command events, including domain events and command boundaries.
The separate domain capture contains no body, title, brief, or reason fields.
Synthetic brief, title, verification-reason, and cancel-reason sentinels do not appear in those events.
No real credential was read or used.

F3 prints `cancel_reason: stop reason tab here` for the LF/CR/TAB reason.
The empty reason stays absent in JSON. Cancel preserves the prior failed verification object.
F4 checks LF, CR, TAB, ESC, DEL, and U+0085 through owned interpolation sites.
Each failure retains one physical diagnostic line. The prior assigned task stays unchanged.
One actual output is:

```text
hive: task not found: "bad\nsecond-line"
```

G1 uses a 120000-byte U+0001 title and a 120000-byte U+0002 verification reason.
Create, assign, and complete succeed. Verify fails with exit 1:

```text
outcome=failed phase=transition effects=[] error=record too large: <fixture task path>
prior task bytes=720290 state bytes unchanged=True
```

The whole prior state remains byte-identical. No temporary file remains.
Plain and JSON show/list each succeed with status `completed` and no verification.
T03 adds no title cap. The shared store owns this prerequisite.

### V4 full uncached gate

All commands ran from the actual task worktree root at `ba91e8a7`:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run --allow-serial-runners ./...
go test -race -count=1 ./...
```

```text
commit ba91e8a750fa6639781a9717550161923e02398f
gofmt exit 0
vet exit 0
0 issues.
lint exit 0
ok  	github.com/Harrison-Blair/hive	2.766s
ok  	github.com/Harrison-Blair/hive/cmd	1.043s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.016s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.046s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	17.296s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.013s
ok  	github.com/Harrison-Blair/hive/internal/logs	5.519s
ok  	github.com/Harrison-Blair/hive/internal/project	1.021s
ok  	github.com/Harrison-Blair/hive/internal/task	1.172s
ok  	github.com/Harrison-Blair/hive/internal/tui	1.016s
test exit 0
```

The format list is empty. Lint ran serially. All packages pass.
`gate.txt` also records the packages without tests.

### Capture paths, cleanup, and limits

Root report: `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t03-verify-02.md`.
Captures: `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t03-verify-02-captures/`.
The captures contain exact outputs, scripts, diffs, route evidence, and worker file hashes.

Cleanup removes only this pass's fixture homes, binary, and isolated source copies.
All existing source, worker evidence, reports, logs, credentials, worktrees, and agents remain unchanged.
Only this dated section changes a tracked file.
The STE baseline does not rise. No global harness setting changes.
No agent starts. No merge, tracker change, handoff change, or worktree removal occurs.

The approved boundary remains a local user and a cooperating writer.
Hostile file replacement between `Lstat` and `Open` remains deferred for T01.
The later T09, T11, and T12 features remain outside T03.
No owned acceptance check or prerequisite check remains uncompleted.
Uncompleted: coordinator acceptance, landing, and the later tracker update.
Recommendation: accept this exact T03 worker for a separate landing.
