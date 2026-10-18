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
