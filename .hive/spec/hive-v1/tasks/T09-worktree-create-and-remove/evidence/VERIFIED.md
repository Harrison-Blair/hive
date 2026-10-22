# T09-worktree-create-and-remove VERIFIED

| Item | Value |
| --- | --- |
| Task | T09-worktree-create-and-remove |
| Checked worker | `161d637ebf29076aeceea23e66f21db1e5bd9672` |
| Source feature | `917eb66c7aa1f205cb8aba860cb9b99524f05e32` |
| Activated base | `ff5e8c44d9b7d1fb510a568e9a248d84cab0b906` |
| Date | 2026-10-18 |
| Agent | verifier, external Pi, `openai-codex/gpt-6.1-sol:high` |
| Pass | `t09-verify-01` |
| Verdict | FAIL |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | FAIL | Existing assertions pass. Independent create findings F01–F06 remain. |
| V2 | AC2, AC3 | FAIL | Existing assertions pass. Independent removal and Git findings remain. |
| V3 | AC1, AC2 | PASS | Actual native create, refusal, cancellation, clean removal, retained branch, and owned cleanup pass. |
| V4 | AC4 | PASS | Formatting is empty. Vet, lint, and uncached race tests pass. |

The combined verdict is FAIL. Do not accept T09 from the green gate or native transcript alone.
No CP2 or later task receives acceptance.

## Evidence and scope

The complete primary report is:

`.hive/tmp/coordinator/reports/t09-verify-01.md` in the root home.

Its captures are `t09-verify-01-captures/` beside that report.
The secondary report is `t09-create-verify-01.md` in the same directory.
I read that report through `END T09 CREATE VERIFY 01` before probes.
The secondary completed and stopped. Its 277 tracked hashes equal this pass's freeze.

The full activated-base-to-worker diff has 23 paths, 3408 insertions, and 16 deletions.
I read every changed source file, test file, and V file.
The primary report assigns every path and separates secondary reliance.
The old command union and bare TUI remain intact.
The workspace metadata type has five native fields and no branch field.
Only `WorkspaceInfo.Worktree` changes type in `types.go`. The other changes there align formatting.

## Findings

| ID | Disposition |
| --- | --- |
| F01 | A task-write failure loses the observed repository workspace effect. |
| F02 | Unknown create treats an unrelated regular file as a created worktree. |
| F03 | Unknown create does not persist an observed repository workspace without a linked path. |
| F04 | Unknown create omits failed branch inspection and repository ambiguity from its diagnostic. |
| F05 | Loaded record identity bypasses the selected key. Removal can also update a different task. |
| F06 | External logs omit applicable task/project IDs. Some create effects and diagnostic bounds are also absent. |
| F07 | Separate task/project persistence groups remain a design concern. This concern is not necessary for the FAIL verdict. |
| F08 | Removal accepts a registered linked checkout as the primary repository. |
| F09 | Removal accepts an unrelated absolute checkout and non-task branch from a malformed local record. |
| F10 | `RefExists` turns a real Git failure into ordinary ref absence. |

F01–F07 retain the secondary identifiers.
I checked the source for each finding. F05 and F06 also reproduce in this pass.
F08–F10 reproduce with real Git in an owned source copy.
Each socket probe replaces only the external peer.
Recommendations and exact test names are in the primary report.
No source correction occurs in this pass.

## V1 run

The actual-worktree gate runs all unchanged create assertions.
The secondary also records frozen green, assertion-red, and restored green runs.
Its focused mutations expose missing focus and physical-path assertions.
I rely on its create-specific independent probes, not on its area verdict as whole-task acceptance.

```sh
go test -race -count=1 ./...
```

```text
ok  	github.com/Harrison-Blair/hive/internal/worktree	1.294s
exit 0
```

The package line above is an excerpt. `gate-race.txt` retains the complete output.

## V2 runs

```sh
go test -race -count=1 -timeout=60s ./internal/worktree/ ./internal/lib/gitx/ -run 'TestRemove|TestGitx' -v
```

```text
Frozen worker: exit 0
Controlled behavioral stub: exit 1
Exact restored worker: exit 0
Final restored worker: exit 0
```

These lines summarize four retained outputs, not a literal process transcript.
Files are `v2-frozen-green.txt`, `v2-controlled-stub-red.txt`, `v2-restored-green.txt`, and `final-v2-green.txt`.
The stub fails assertions, including the bounded missing-request assertion. It does not fail compilation or reach a test timeout.
Each of the eight Git helpers also has a focused assertion-red mutation.
Two removal mutations break the common-directory and dirty guards. The old assertions detect both.
All tracked copy hashes match after restoration.

```sh
go test -race -count=1 -timeout=60s ./internal/worktree -run TestAudit -v
```

```text
--- FAIL: TestAuditRemoveRequiresRegisteredPrimary
--- FAIL: TestAuditRemoveRejectsOutsideTaskPathAndBranch
--- FAIL: TestAuditRemoveRecordIDMustMatchSelectedKey
--- FAIL: TestAuditGitxErrorsAreNotAbsence
exit 1
```

This excerpt lists the four independent assertion failures. `independent-behavior.txt` retains the complete output.
Ten fresh race-enabled concurrency runs pass.
Nested-repository refusal, static bad path types, status guards, and typed/unknown close routes pass.
A successful close followed by real Git failure retains the close effect and reports partial, without rollback.

Uncompleted: exact historical test-first stub replay. The retained worker files contain outputs but no exact stub setup.
The controlled red proves assertion sensitivity only. It does not prove the historical development order.
An initial independent probe had one unused variable. The captures retain its compile failure. This failure does not count as red.

## V3 native run

One absolute binary came from the frozen actual worktree with `vcs.modified=false`.

```text
binary: /tmp/t9p.d3ayccb_/hive-frozen
SHA256: b5e183d5ba18f98e81ee36756a9432070004a9fadd6aa8f2a3cfbeda1c97d586
home: /tmp/t9n.aqerjcy8
repository: /tmp/t9n.aqerjcy8/repo-t09-verify
workspace: wAS
tab: wAS:t1
root_pane: wAS:p1
path: /tmp/t9n.aqerjcy8/projects/repo-t09-verify/worktrees/1-live-demo
branch: task/1-live-demo
repository workspace: wAR
```

`native.py` retains the exact commands and bounded socket calls.
The `native/` directory retains each argv, status, stdout, stderr, record stage, request/reply, and log.
The first remove exits 1 with the exact created-status refusal.
Cancellation succeeds. The clean no-force remove then prints this exact output:

```text
workspace closed: wAS
worktree removed: /tmp/t9n.aqerjcy8/projects/repo-t09-verify/worktrees/1-live-demo
branch kept: task/1-live-demo
```

The product leaves `wAR` open. Test cleanup proves its PRIMARY metadata before closing only `wAR` with `close_group=false`.
The close returns `ok`. A parsed list proves that both owned IDs are absent before fixture deletion.
The before and final IDs are `w9X`, `wAD`, `wAE`, and `wAF`.
No unrelated ID changes during this run. No agent or model child starts.

## V4 gate

These commands run in the actual worktree with a short outside-home `TMPDIR`:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run --allow-serial-runners ./...
go test -race -count=1 ./...
go test -run TestSTE -count=1 .
```

```text
format: empty output, exit 0
vet: exit 0
lint: 0 issues., exit 0
race: every package passes, exit 0
TestSTE: PASS, exit 0
```

`gate-*.txt` retains literal output. These lines summarize the results.
Both-glossary lint adds zero hard findings in the new report and verdict.
The STE baseline stays exact. The T09 range diff check passes.
Historical evidence retains eight literal trailing spaces in T04 V3 and three in T22 VERIFIED.
Those lines precede the activated base and remain unchanged.

## Preservation and limits

All 277 original tracked hashes remain exact, including V1–V4 and the baseline.
Only this evidence file enters the new commit. The worker commit is its sole parent.
No source, dependency, task, tracker, AGENTS, state, handoff, or attribution changes.
No root merge, force push, main write, credential read, or global configuration change occurs.

The no-follow cleanup manifest lists 418 entries in four owned obsolete roots.
Compact records, exact probes, patches, raw outputs, synthetic logs, and hashes remain in the captures.
The copy, native fixture home, binary, and two outside-home temporary roots are absent after cleanup.
All owned subprocesses returned. Each driver reaped its subprocesses. No unowned resource closed.

Hostile file replacement, malicious peers beyond T05, and the pre-T20 live-agent guard remain deferred.
Source fixes and fresh acceptance remain uncompleted.

END VERIFIER T09-worktree-create-and-remove t09-verify-01


## Repeat independent verification: t09-verify-02

| Item | Value |
| --- | --- |
| Checked worker and sole verdict parent | `2d9c96d8116140bc1a151dea9b8c9d2874a39ecc` |
| Fix source/tests | `1653bf488e61d4f9f910f0a0fe28efdf2c25d066` |
| Fix evidence | `d2afc78618384433c6a8741b37db24c15b35f19b` |
| First verdict | `ef396c34809e9b5da157928afe92bde7568004e6` |
| Original worker | `161d637ebf29076aeceea23e66f21db1e5bd9672` |
| Activated base | `ff5e8c44d9b7d1fb510a568e9a248d84cab0b906` |
| Date | 2026-10-18 |
| Agent | verifier, external Pi, `openai-codex/gpt-6.1-sol:high` |
| Actual route | `openai-codex`, `gpt-6.1-sol`, reasoning `high` |
| Verdict | FAIL |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | FAIL | Retained probes pass. Current metadata conflicts and the multibyte log bound fail. |
| V2 | AC2, AC3 | FAIL | Removal overwrites conflicting current topology. |
| V3 | AC1, AC2 | PASS | Fresh native CLI, branch/history, and owned cleanup pass. |
| V4 | AC4 | PASS | One actual-worktree uncached gate passes. |

The complete report is root `.hive/tmp/coordinator/reports/t09-verify-02.md`.
Its initially absent capture directory is `t09-verify-02-captures/` beside that report.
I read both complete first-review reports and both complete worker reports through their end markers.
All three source agents completed and stopped before this repeat. I started no agent or model child.
The current task and current verifier brief control this pass. Historical payloads remain audit data.

### Finding dispositions

| ID | Disposition |
| --- | --- |
| F01 | CLOSED. Task write failure keeps all known external effects and result `repo_workspace_id: w7`. |
| F02 | CLOSED. A plain file alone grants no created-worktree or unobserved-branch metadata. |
| F03 | CLOSED. Repository-only observation persists the unique PRIMARY project ID independently. |
| F04 | CLOSED. Original call error, failed observations, list failure, and ambiguity remain explicit. |
| F05 | CLOSED. Both commands check selected task/project identities before external work. |
| F06 | PARTLY CLOSED. IDs and effects pass. A multibyte copied diagnostic still exceeds 4096 bytes. |
| F07 | Group shape PASS. It remains a design observation. R01 checks the required current resource ownership. |
| F08 | CLOSED. Linked registered source refuses at the permitted target path, with and without force. |
| F09 | CLOSED. Outside paths and precious branches refuse. Physical aliases still work. |
| F10 | CLOSED. Valid ref, absent ref, and fatal non-repository controls pass. |

R01 concerns the changed latest-record helpers, not a new broad probe class.
Source: `internal/worktree/create.go:435-490` and `internal/worktree/remove.go:284-324`.
A cooperating writer stores current repository workspace `w99` before create persistence.
Create overwrites it with `w7` and returns `ok` instead of partial/project.
A changed physical repository owner also receives old repository workspace `w7` with `ok`.
During removal, a cooperating writer records topology `w3`, `w3:t2`, and `w3:p4` before the close reply.
Removal closes `w2`, removes its worktree, then clears the different current topology and returns `ok`.
The expected result keeps conflicting current metadata and reports partial/task with the actual external effects.
Recommendation: check current resource ownership and owned metadata before replacement or clearing.
Keep the short groups and completed prefix. Add no retry, journal, transaction marker, or rollback.

F06 concerns `boundLogText` at `internal/worktree/create.go:609-614`, which both command error paths use.
A 9000-byte euro-sign diagnostic becomes a decoded logged error of 4098 bytes, with `truncated: true`.
The byte cut splits UTF-8. JSON encoding replaces the incomplete character.
Both commands fail the literal 4096-byte assertion. ASCII controls pass.
Recommendation: keep valid UTF-8 within the byte limit in the T09 helper.
Keep the accepted shared logger, runner, and socket helper unchanged.

### V1/V2 and retained probes

These commands ran in an owned exact source copy with a short outside-home `TMPDIR`:

```sh
go test -race -count=1 -timeout=120s ./internal/worktree ./internal/lib/gitx -run TestCreate -v
go test -race -count=1 -timeout=120s ./internal/worktree ./internal/lib/gitx -run 'TestRemove|TestGitx' -v
go test -race -count=1 -timeout=180s ./internal/worktree -run TestAudit -v
python3 cli-probe.py
```

The exact retained create and remove probes ran separately because both define the private name `auditFixture`.
No assertion or fixture adaptation was necessary. F08 already binds the permitted path.
`v1-green.txt`, `v2-green.txt`, `create-replay.txt`, `remove-replay.txt`, and `cli-replay.txt` retain complete output.
Every command exits 0. The CLI driver records child statuses `0, 0, 1, 1, 0` for its five cases.
The task-fault and unknown cases retain the actual worktree and branch without rollback.
The task-fault envelope keeps the four known effects and result `repo_workspace_id: w7`.
The concurrent CLI cancellation completes during the held socket request and survives creation.

The new fix tests run assertion-red on exact base `ef396c34` plus the two new test files.
The same tests pass on the frozen source. Raw outputs are `fix-base-red.txt` and `fix-frozen-green.txt`.
Fifteen focused mutations each fail assertions. The changed target-linked fixture also detects its broken guard.
Every mutation restores exact source hashes. The restored V1/V2/fix suite passes.

```sh
go test -race -count=2 -timeout=120s ./internal/worktree -run 'TestRepeat(CreateLatest|RemoveLatest|Unicode)' -v
```

```text
--- FAIL: TestRepeatCreateLatestRepositoryWorkspaceConflict
--- FAIL: TestRepeatCreateLatestRepositoryOwnerConflict
--- FAIL: TestRepeatRemoveLatestTopologyConflict
--- FAIL: TestRepeatUnicodeDiagnosticBound
exit 1
```

This excerpt lists the repeated failing assertions, not the complete transcript.
`repeat-failures-count2.txt` retains both runs and the actual envelopes.
Changed task ID/project/worktree conflict controls, physical aliases, and no-external-work identity checks pass.
Ten fresh real store/socket coordination runs pass for each command. External waits release the lock.
No compile failure or timeout counts as red.
One driver attempt selected its own path receipt as test output. It ran no test and counts as no red.
The exact historical worker stub setup remains unavailable. I fabricated no development-order evidence.

### V3 native run

```text
binary: /tmp/t92p.ofqpx6hs/hive-frozen
SHA256: f7d477747befcaab1ab1ed6dd6b7155d7619a39efe2cf73db23deafbf2442bad
home: /tmp/t92n.k8u342md
project: repo-t09-verify
task: 1-live-demo
PRIMARY workspace: wAW
linked workspace: wAX
root pane: wAX:p1
create exit 0
remove-created exit 1
cancel exit 0
remove-cancelled exit 0
branch-kept exit 0
```

The binary comes from the checked actual worktree with `vcs.modified=false`.
The native driver starts native workspaces and default shells only.
It captures actual CLI output, records, logs, requests, replies, effects, and before/after topology.
The first remove returns the exact created-status refusal. Cancellation then permits clean no-force removal.
The physical path is absent. The branch, creation history, cancellation reason, and removal time stay.
All 42 actual external log events carry the applicable task/project IDs.
The product leaves PRIMARY `wAW` open.
Cleanup proves its owned PRIMARY metadata, closes only `wAW` with `close_group=false`, and receives `ok`.
Parsed lists prove both owned IDs absent before fixture deletion.
The before/final IDs are `w9X`, `wAD`, `wAE`, and `wAF`. No unowned resource changes in this native run.
The driver bounds calls and reaps its own subprocesses. It sends no user pane input or focus request.

### V4 actual-worktree gate

The full gate ran once after final Go checks, with a short outside-home `TMPDIR`:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run --allow-serial-runners ./...
go test -race -count=1 ./...
go test -run TestSTE -count=1 .
```

```text
format: empty output, exit 0
vet: exit 0
lint: 0 issues., exit 0
race: every package passes, exit 0
TestSTE: PASS, exit 0
```

These lines summarize `gate-*.txt` and `gate-results.json`.
`gate-tested-hashes.json` records exact tested Go, source, dependency, and baseline bytes.
The evidence-only comparison retains those bytes afterward. Both-glossary new prose has zero hard findings.
No baseline increase or dependency change occurs. The activation and fix diff checks pass.

### Freeze, preservation, and limits

All 281 tracked files received a SHA256 freeze. All source-copy hashes match after restoration.
The original 8697-byte VERIFIED prefix stays exact.
Its SHA256 is `26077a27494e7a7c1cb6f75f1d6c5c3ef4647c692cf5755bf5438f81102a27a3`.
All other original tracked bytes and all 1008 historical artifact hashes stay exact.
The seven-path fix scope matches. Every other first-verdict blob and accepted-base blob stays exact.
The fix worker captures pass SHA256SUMS. Its eleven removed roots remain absent. I removed no old leftover.
The no-follow manifest records 1156 entries in five owned obsolete roots. Cleanup removes only those roots.
Exact scripts, patches, raw outputs, synthetic logs, records, and hashes remain in the new capture directory.
Worktree runtime roots, state, logs, old reports, credentials, and other worktrees stay.
I read or printed no credential content and changed no global configuration.
The sole new commit appends only this file to the checked worker. The push changes only the named task branch.
No force push, main write, root source/state/handoff edit, co-author, or false session trailer occurs.
Hostile Lstat/Open replacement, malicious peers beyond T05, and the pre-T20 live-agent guard remain deferred.
The nonconforming nil-error short writer stays outside T09 acceptance.
Uncompleted: R01 and complete F06 correction, fresh acceptance, landing, CP2, and later checkpoints.

END VERIFIER T09-worktree-create-and-remove t09-verify-02


## Repeat independent verification: t09-verify-03

| Item | Value |
| --- | --- |
| Checked worker, evidence, and sole verdict parent | `79360a6460c369af2eb4d0aa92be0ad83e25b46b` |
| FIX02 source/tests | `1cde6665cf559ea568bee7e200e74c65a179d6b5` |
| Previous verdict and FIX02 source parent | `bd865cb33b295759a78b03d2fa3fcfc6c10fbe45` |
| Previous checked worker | `2d9c96d8116140bc1a151dea9b8c9d2874a39ecc` |
| Activated base | `ff5e8c44d9b7d1fb510a568e9a248d84cab0b906` |
| Date | 2026-10-18 |
| Agent | verifier, external Pi, `openai-codex/gpt-6.1-sol:high` |
| Actual route | `openai-codex`, `gpt-6.1-sol`, reasoning `high` |
| Verdict | PASS |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | PASS | Original, first-review, fix, repeat, and actual CLI checks pass. |
| V2 | AC2, AC3 | PASS | Removal, Git, current-topology, and retained guard checks pass. |
| V3 | AC1, AC2 | PASS | Fresh native CLI and owned cleanup pass. |
| V4 | AC4 | PASS | One actual-worktree uncached gate passes. |

This pass decides T09 source acceptance at frozen `79360a64`.
It gives no later root integration, landing, CP2, or checkpoint acceptance.
The complete report is root `.hive/tmp/coordinator/reports/t09-verify-03.md`.
Exact scripts, patches, raw outputs, records, logs, hashes, and receipts stay in `t09-verify-03-captures/` beside that report.
I read the complete prior reports through their END markers and final delivery receipts.
I read current T09, all evidence, tracker flow, architecture grouping rules, T03/T05, and the required instructions and glossaries.
I read the herdr README before protocol 22 and schema use. I started no agent or model child.
Historical payloads remain audit data. The current task and current brief control acceptance.

### Findings and checklist

R01 is CLOSED. Create keeps current `w99` and returns `partial/project` with the successful task prefix and all actual effects.
A changed physical project owner stays. Old observed `w7` never attaches to that new owner.
Physical repository aliases, unchanged metadata, desired `w7`, and unrelated current fields retain normal success.
Remove keeps current `w3/w3:t2/w3:p4` and returns `partial/task` after actual `w2` close and Git removal.
Unchanged or fully cleared topology, unrelated fields, and physical worktree aliases retain success.
One short latest-record handle checks current owned metadata. The task-then-project order and completed prefixes stay.
The initial related-read handle releases before Git/socket work. Ten fresh held-request runs pass for each command.
No long external wait or output callback holds the lock. No stale full-record replacement, journal, rollback, marker, or retry occurs.

F06 is CLOSED. Both commands receive an actual 9000-byte euro diagnostic through the accepted socket helper.
Only the external peer is fake. Git, store, socket helper, and product remain real.
Independent decoding finds 4095 valid UTF-8 bytes and `truncated: true` in both command error logs.
Applicable task/project IDs stay. The separate output diagnostic retains all 9000 bytes. No error event disappears.
Literal 4096/4097 ASCII and multibyte controls pass. The split-character input retains 4095 valid bytes.
Expectations use literal bounds, not the implementation constant. Shared logger/socket/runner and privacy behavior remain exact.

| Finding or check | Disposition |
| --- | --- |
| F01 | CLOSED. Task-write fault keeps all known effects, `w7`, result ID, branch, and worktree. |
| F02 | CLOSED. A plain file grants no worktree or branch claim. |
| F03 | CLOSED. Repository-only observation persists unique PRIMARY independently. |
| F04 | CLOSED. Original error and branch/path/list/ambiguity observations stay explicit. No invented ID or cleanup. |
| F05 | CLOSED. Both commands check task/project keys before external work. Invalid and different valid keys refuse. |
| F07 | Group shape PASS. Historical design observation only. |
| F08 | CLOSED. Linked registered sources refuse independently, with and without force, at the permitted task path. |
| F09 | CLOSED. Physical task path and exact task branch bind authority. Outside paths and precious branches refuse, even with force. |
| F10 | CLOSED. Real present/missing refs and fatal non-repository errors remain distinct. |
| Create parameters and errors | PASS. Exact cwd/path/branch/base/label/root pane and both focus values. No workspace selector. |
| Typed and unknown create effects | PASS. Known observations only. No invented candidate, retry, rollback, or unowned close. |
| Latest unrelated fields and identity conflicts | PASS. Status/title/actor/verification/dependencies/reason fields survive. Conflicting current records stay. |
| Removal guards | PASS. Status/cancel/verify, linked target, common directory, nested repository, dirty, and merged checks stay. |
| Force | PASS. It skips only dirty and merged guards. |
| Close routes | PASS. Open/closed/not-found/group-root and explicit `close_group=false`. No product `worktree.remove` or PRIMARY close. |
| Unknown close and Git failure | PASS. Unknown stops Git. Known close plus Git failure keeps partial/worktree, actual close effect, and Git output. |
| Normal removal | PASS. Physical path absent. Live topology clears. Branch/path/history stay with `removed_at`. |
| Git and regressions | PASS. Direct argv, deadlines, nonzero ExitCode, root command union, T03/T05, and store. |

### Red, green, mutations, and CLI

The independent red copy is exact `bd865cb3` plus only the saved new `fix02_test.go`.
Test SHA256: `93eb1724d00f7b8589e7e3f67b3e8c31febfd3618ba00340be098ddc9b60c362`.
Old create SHA256: `2d26277c825c1c1a31e63b798c85db459b088945ed091913f03fa882c143a25e`.
Old remove SHA256: `5cdf536735a72de798b631f6a84e817bcb8acce7dbfd0ba1d960774041bd3aec`.

```sh
go test -race -count=1 -timeout=180s ./internal/worktree ./internal/lib/gitx -run TestFix02 -v
```

```text
base bd865cb3 plus exact new test: exit 1
frozen 79360a64: exit 0
```

These lines summarize `fix02-base-red.txt` and `fix02-frozen-green.txt`.
All three R01 variants and both Unicode commands fail assertions on old production. The boundary test detects invalid UTF-8.
No compile failure, panic, or Go timeout counts as red.
Five focused mutations detect workspace conflict, owner conflict, physical alias refusal, topology conflict, and UTF-8 bound defects.
Each mutation restores exact hashes. Restored green passes. No unrelated mutation class repeats.
The worker's first UTF-8 mutation is an INVALID compile failure. Its corrected assertion-red counts and stays visible.
The exact original historical worker stub setup remains unavailable. This pass invents no development chronology.

Exact first-review create/remove probes run separately because both define `auditFixture`. Their assertions and fixtures stay exact.
The exact repeat probe passes. Original V1/V2 and FIX01 regressions pass.
The unchanged CLI probe has SHA256 `a78e75a6a6f98272b04e996a8788009db276d13c73ed457ced0b1543387437e3`.
Its text, JSON, task-fault, unknown, and concurrent statuses are `0, 0, 1, 1, 0`.
Task-fault retains all four effects and result `w7`. Actual CLI cancellation completes during the held request and survives create.
The real Git branch and worktree remain after fault and unknown outcomes. No rollback occurs.

### Native V3 and full gate

```text
binary: /tmp/t93t.dz5o0uk_/hive-frozen
SHA256: 7fde808d74c6228d78640eb4b72964ef1c83f195e8c7e13965bf987fd86f8277
build: vcs.modified=false, revision 79360a6460c369af2eb4d0aa92be0ad83e25b46b
home: /tmp/t93n.cq78iodu
PRIMARY: wA0
linked/root pane: wB1 / wB1:p1
create exit 0
remove-created exit 1
cancel exit 0
remove-cancelled exit 0
```

The driver uses one unique marked home/repository and the absolute frozen binary. Native herdr answers `0.9.5`, protocol `22`.
Only native workspaces and default shells start. The path is absent after no-force removal. Branch and task history stay.
The product leaves PRIMARY `wA0` open. Cleanup proves ownership and closes only `wA0` with `close_group=false`.
The close succeeds. Parsed lists prove both owned IDs absent before fixture deletion.
Before and final sets contain `w9X`, `wAE`, and `wAF`. User panes `w9X:p4/p5` stay.
All 42 native external events carry applicable task/project IDs. Drivers bound calls and reap subprocesses.

The full gate runs once in the actual worktree at final Go source, with short outside-home `TMPDIR=/tmp/t93g.npbm8_3s`.

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run --allow-serial-runners ./...
go test -race -count=1 ./...
go test -run TestSTE -count=1 .
```

```text
format: empty output, exit 0
vet: exit 0
lint: 0 issues., exit 0
race: every package passes, exit 0
TestSTE: PASS, exit 0
```

These lines summarize `gate-*.txt` and `gate-results.json`. The captures retain complete command output.
Evidence-only checks compare tested Go/source, dependencies, AGENTS, and baseline bytes afterward.
Both-glossary new prose has zero hard findings. No baseline or dependency change occurs.

### Scope, preservation, and limits

The 29-path activation map and exact four-path FIX02 map remain separate in captures.
I read every changed line in `create.go`, `remove.go`, new adjacent `fix02_test.go`, and new `FIX-02.md`.
All other 279 prior-verdict blobs/modes and all 254 unchanged accepted-base blobs/modes stay exact.
All 283 tracked files received a freeze before probes. All 2116 historical artifact hashes remain exact.
The old 18820-byte VERIFIED prefix stays exact, with SHA256 `dea14ee34975198dab978774c2d8ea5da3804ddda28299d8c4f02a80ee8aed1a`.
The fix02 preservation check passes for 1630 files and 493 capture hashes. Its 13 removed roots remain absent.
The dispatched fix02 brief retains its final newline and SHA256 `77b9fa6b0d7debcb55905d0a60913254f07f0a2adcce465fa70c80f2cdedc12f`.
The stripped variant has SHA256 `8137685eea5036fadd601010a6d30f793f2dc48b1951ddd89c5bd20096eecde9`.
The original worker report stays. Correction: this T09 pass did not edit root, but the independent T11 landing did change root.
No root freeze claim applies. Other tasks and root can advance independently.
The no-follow cleanup manifest lists 1520 entries in six owned obsolete roots. Only those roots were removed.
Worktree runtime roots, logs, state, old reports, worktrees, credentials, and other agents stay.
No credential content, global configuration change, user pane input, unowned close, or unowned signal occurs.
The sole new commit appends only this file to `79360a64`. Final delivery receipts stay in captures.
Hostile Lstat/Open, malicious peers beyond T05, and the pre-T20 live/unresolved-agent removal guard remain deferred.
The nonconforming nil-error short writer remains outside T09 acceptance.
Uncompleted: exact original historical stub chronology, later root integration, landing, CP2, and later checkpoints.

END VERIFIER T09-worktree-create-and-remove t09-verify-03
