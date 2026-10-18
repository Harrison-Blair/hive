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
