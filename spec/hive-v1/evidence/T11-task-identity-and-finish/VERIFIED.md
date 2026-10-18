# T11 independent verification 01

To: coordinator.
Date: 2026-10-18.
Pass: `t11-verify-01`.
Verdict: **FAIL**. No source acceptance or landing approval.
Route: external Pi, `openai-codex/gpt-6.1-sol:high`.
Observed provider/model/level: `openai-codex`, `gpt-6.1-sol`, `high`.
Session: `01a14e7e-31b7-73b9-8d5b-4cf1a3ae132b`.

## Identities and authority

| Item | Exact identity |
| --- | --- |
| Tested worker and required verdict parent | `129b82f2afba5f5c251bc518806193d319831b2c` |
| Last production source | `6836b4b27a7494cf867847d20889e133c374bebe` |
| Last worker test change | `534a153bb495ab6366e3c0da0704b86d2b3a783f` |
| Activated base | `eded8f9cb09208bfdc81334870e86479e746c183` |
| Prepared base | `a21f6be186720dd87ee9f4d4abfdd960fae0d88e` |
| Branch | `task/T11-task-identity-and-finish` |

Initial local, tracking, and direct remote task tips equal the tested worker.
Every production Go file equals the last production source above.
I read the complete worker report through its required end marker.
I read V1–V4, the current task, both base tasks, instructions, glossaries, and untrusted-input rules.
Both preserved activation reports pass at the same activated base.
`ACTIVATION.md` records checks and the authorized fast-forward only. It grants no new permission.
The report stays as the authorized evidence exception.
The available implementation template is stale. It asks for forbidden finish fields and replay.
UNCOMPLETED: no saved concrete checked T11 implementation payload was found in the authorized brief area.
The installed task and this dispatch control the source check. The stale template grants no permission.
Independent root T05 progress requires no merge into this branch.

## Freeze, range, and scope

Captures: `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t11-verify-01-captures/`.
The following capture names refer to that directory.
`freeze.json` records all 247 tracked hashes before probes. `freeze-final.json` matches every hash before this evidence write.
The full activated-base-to-worker diff has 30 paths: 15 source, 10 test, and five evidence paths.
`path-manifest.json` accounts for every path, area, blob, and worker hash.
Its SHA256 is `f449836d92b6137015da9541996e042fa5d65d70ca231c11263e09771affe6cb`.
`diff.patch` retains the full range, not only HEAD.
Its SHA256 is `780b153927c81ac4e6a43b09074973c5626ba49d08120617653c7d1582b05a4f`.
I inspected all changed source, tests, and evidence. The secondary separately checked the inbox and output area.
No common renderer, store, dependency, baseline, supervisor, process, or unrelated source path changed.
The assignment guard uses the shared transition path. `assign.go` needs no edit.
No finish sequence, finish key, journal, rollback, or extra notification field exists.
No unrelated scope expansion was found. The test coverage loss below remains open.

## Combined V1–V4

| Item | Result | Disposition |
| --- | --- | --- |
| V1 / AC1 | FAIL | Selected claim aliases and static path/key mismatches remain accepted. Registration checks otherwise pass. |
| V2 / AC2 | FAIL | Key/name complete, advanced partial-retry guidance, logging, and rework inspection defects remain. |
| V3 / AC2, AC3 | FAIL | The required actual transcript passes. Broader required identity, preservation, logging, and output checks fail. |
| V4 / AC4 | PASS | All four actual-root gate commands pass. TestSTE passes without a baseline rise. |

The secondary result is FAIL, not whole-task acceptance.
I read its complete report through `END T11 INBOX OUTPUT VERIFY 01`, including every finding, passed check, and limit.
The report preceded my source freeze. Its tracked-source hashes match the tested worker.
Default inbox and stdout matrices below rely on that separate result. I did not duplicate its full matrix.

## P-F01: A selected cross-claim alias resolves another role

Location: `internal/lib/identity/identity.go:60–65,117–130`.
The selected claim holder passes `validRecord` against the holder's own role claim, not the selected claim.
A literal `manager-one` claim points at a valid `manager-two` holder with equal generation 7.
Expected: `agent not found: manager-one`. Actual: the flag returns the project-two holder.
The worker cross-claim test omits the selected claim alias and a simultaneously valid holder claim.
Actual CLI evidence also points `manager-repo-one` at the registered worker record.
Expected: refusal without effects. Actual: exit 0 and all four complete effects.
This is a static mismatch, not hostile replacement.
Evidence: `identity_probe_test.go`, `independent-final.out`, `cli/finding-cross-claim.json`.
Recommendation: check the selected claim key, holder key, role, project, and generation together.

## P-F02: Complete rejects the acting record key

Location: `internal/task/transition.go:249–267,100–102`.
The first identity check accepts the caller key. The second transition check compares only the caller name.
Fixture: key `worker-key`, name `worker-name`, acting agent `worker-key`.
Expected: complete succeeds through key, name, pane, and history routes.
Actual: all four refuse with `not the acting agent: worker-name`, phase `task`, no effects.
The actual CLI reproduces this with a registered record whose name becomes `custom-worker` before command start.
The task still names its real record key. The pane identifies that record correctly.
Evidence: `task_probe_test.go`, `independent-final.out`, `cli/finding-key-name.json`.
Recommendation: use one resolved acting-agent identity through the FSM check. Preserve the stored task attribution.

## P-F03: Advanced partial retries omit inspection commands

Location: `internal/task/transition.go:245–246,265–267`.
Only a caller with `finished` true receives the advanced-finish inspection commands.
A real counter-write failure leaves the advanced task and an unfinished caller.
Expected: normal refusal without writes, with task show, agent show, and inbox guidance.
Actual complete retry: `invalid transition: completed -> complete` only.
Actual verify retry: `invalid transition: verified -> verify` only.
Actual failed-verify retry: `invalid transition: assigned -> verify` only.
All three independent assertions fail. Real CLI partial-retry captures show the same diagnostics.
Evidence: `TestIndependentPartialRetryInspection`, `independent-final.out`, `cli/*-counter-retry.json`.
Recommendation: add inspection guidance to advanced finish refusals without inferring an old occurrence or replaying writes.

## Secondary findings and primary disposition

| Secondary ID | Disposition | Required correction |
| --- | --- | --- |
| T11-IO-F01 | OPEN, retained | Explicit invalid send opens the store before key checks and creates state. Restore the old check order. |
| T11-IO-F02 | OPEN, retained | Default-read identity/role refusals omit their error event. Log the actual refusal before a role key exists. |
| T11-IO-F03 | OPEN, retained | Finish logs omit actual effects and pane-resolved actor IDs. Owned stdout failures omit the stream error event. |
| T11-IO-F04 | OPEN, independently reproduced | Static claim and holder filenames disagree with embedded keys, yet all four identity routes accept them. |
| T11-IO-F05 | OPEN, retained | Fresh-worker complete with short stdout names the old failed verifier for agent inspection. Pass the actual operation actor. |

F04 is one finding, not a duplicate primary ID.
Primary probes rename either the holder file or the claim file before lookup.
Claim alias, custom name, pane, and history each accept both mismatches. All eight refusal assertions fail.
The secondary actual CLI proves the same mismatch consumes an inbox record.
Recommendation: retain selected filenames and check them against embedded record keys before any effect.
F01–F03 and F05 retain the secondary's exact reproduction, source locations, and recommendation.
The complete report remains in `secondary-report.md`. Its original report and captures remain unchanged.

## T-W01: An old diagnostic test lost its control input

Location: `internal/task/diagnostic_test.go:124–141`.
The changed test replaces `bad\nactor` with plain `worker-2`. Its one-line assertion no longer tests a control character.
An owned mutation removes quoting from the resolved wrong-actor diagnostic.
The changed worker test stays green. The independent literal control-character assertion becomes red.
Restoring source makes the independent assertion green. This is a coverage loss, not an additional current source defect.
The CLI unknown-actor diagnostic still checks a newline. It exercises a different refusal path.
Recommendation: retain a control-bearing fixture or equivalent assertion. Keep the old state-preservation checks.
Other old T03/T04 assertion changes match the intentional identity, finish, registration, and default changes.
Valid live fixtures replace unregistered actors. No skipped test or weaker partial-prefix assertion was found.

## Checks that pass

One held handle contains identity, role/project/task, FSM, notification checks, and the durable group.
The claim exceptions above prevent full acceptance of those identity checks.
Independent custom-name success cases retain literal notification actors, text, effect paths, and finished marks.
Key/name self-verification aliases refuse. Worker/verifier registrations with missing or other-project tasks refuse before all domain writes.
The registration probes preserve a preexisting counter value of 7 and compare every JSON record byte.
Missing, stale, generation, holder-value, role, and project refusals pass through explicit, pane, and history routes.
Sorted ambiguity, unknown callers, and no library environment read pass the existing literal assertions and source check.
All 12 real storage failures pass: three operations times task, counter, inbox record, and finished mark.
First failure is failed/no effects. Later failures report partial, the actual phase, and the exact actual prefix.
Task failures use an unwritable task directory. Other failures independently deny the corresponding directory write.
Command JSON retains the partial envelope. Source callbacks return nil Go errors and do not replace it.
No rollback or notification replay occurs. Retained counter gaps remain explicit.
Full notification checks run before the task write. Literal ASCII and UTF-8 limits pass the secondary's real CLI checks.
Same-reason and changed-reason failed-verify retries preserve all JSON bytes and refuse without effects.
Finished assignment gives the exact required message. Unknown assignment and fresh registration/assignment/complete still work.
Cancellation after retained complete and failed-verify partials remains safe. A later finish refuses without writes.
Ten independent race repetitions cover competing complete, verify pass, and verify fail attempts.
Each pair produces exactly one successful group, one no-effect refusal, and one notification.
A controlled pause after the task write blocks a cooperating combined task/agent reader until the group finishes.
The observed storage-call order is tasks, counter, inbox, agents. The reader then sees completed plus finished.
Show keeps unknown actors displayable. Brief checks remain bounded and outside the held group.
No stdin, process, Git, network, or stdout wait occurs under the finish lock.
The secondary's lock-acquiring output writers pass. Held inbox API, old explicit routes, and partial consume safeguards pass except F01–F04.

## Red, green, CLI, and gate evidence

Exact worker commits replay in owned copies: V1 `4de9975a` red, `6836b4b2` green. V2 `0e5f6690` red, `534a153b` green.
Each red run fails on behavior assertions, not compilation. The raw outputs and exact tree hashes remain.
Owned identity, finished-mark, and diagnostic mutations each produce assertion-red and restored green.
Exact patches and `red-*.out` / `restored-green-*.out` remain. Actual defects stay red. No source fix ran.
The required V3 text transcript uses one marked fixture home, unique repository, and absolute actual-root binary.
Binary SHA256: `7cc079d09b9a6c32bb4bf62b982d064067de126ea2b0668f635ee7da116dae6c`.
`cli/v3-transcript.txt` records sender defaults, worker inbox refusal/status 1, pane complete/verify, manager records, and both finished marks.
Every CLI subprocess has an eight-second deadline. No native agent process starts.
Actual-root format is empty. Vet exits 0. Lint reports `0 issues.`. Full `go test -race -count=1 ./...` exits 0.
The gate uses `/tmp/t11v-_qp3fxe4`, outside every marked home. `gate-*.out` retains commands and results.
TestSTE exits 0. Baseline, ratchet source, and dependencies remain byte-exact to base.
Full range diff check exits 2 for 12 trailing spaces in literal V3 output. All other range paths pass.
This evidence write does not rewrite that output or claim an empty full-range check.

## Cleanup, preservation, and limits

`copy-restored.json` proves all original copy bytes match. Actual source remains frozen through the verdict write.
`cleanup-manifest.json` names only five owned obsolete roots. `cleanup-entries.json` records their 1072 entries.
`cleanup-receipt.json` records all five absent: copy, replay, fixtures, binary, and external TMPDIR.
Exact probes, outputs, patches, hashes, transcripts, and compact synthetic records/logs remain.
No preexisting report, log, credential, worktree, state, attribution, main reference, or global harness setting changed.
No source/V evidence rewrite, root edit, tracker edit, handoff edit, agent child, or model child occurred.
LIMIT: initial process checks showed other root Pi processes without captured session bindings. Later checks found only this verifier.
The secondary report was complete before probes. Its precise exit time before primary startup lacks independent process-binding evidence.
UNCOMPLETED: concrete checked implementation-payload discovery, all source/test corrections, and fresh acceptance after those corrections.
Hostile Lstat/Open swaps, malicious custom callbacks, and unrelated process control remain outside this check.
Recommendation: return every open finding for a bounded fix and fresh independent checks. Do not land this source.
This verdict passes no future checkpoint and claims no native model interoperability result.

END VERIFIER T11-task-identity-and-finish t11-verify-01

# T11 independent fix verification 02

Date: 2026-10-18.
Pass: `t11-verify-02`.
Verdict: **FAIL**. T11-IO-F03 remains open. No landing or future checkpoint acceptance occurs.
Route: external Pi, `openai-codex/gpt-6.1-sol:high`.
Observed provider, model, and reasoning: `openai-codex`, `gpt-6.1-sol`, `high`.
Session: `01a14e9e-9e50-75a9-8ca7-af81d1a42680`.

## Identities and scope

| Item | Exact SHA |
| --- | --- |
| Tested worker and sole verdict parent | `0124ab636ffa62dfcb41256801154970528fdb01` |
| Fix production | `44235d664f553c94f3039f8080ec0a534a34fe20` |
| Fix test-first | `6dce8b53131bc96c3b0a419589284cce65288ceb` |
| First failed verdict | `4d052b28fb71e486bf2c816260bf3cb3cd347b41` |
| Original worker | `129b82f2afba5f5c251bc518806193d319831b2c` |
| Activation base | `eded8f9cb09208bfdc81334870e86479e746c183` |

Branch: `task/T11-task-identity-and-finish`.
Initial local, tracking, and direct remote task tips equal the tested worker. The worktree is clean.
The saved fix brief has SHA256 `4ac7d09e16e85d575dc0861cfdc0051f443ee05359749f0974baf8fa3c0c0470`.
Its worktree path is `.hive/tmp/worker/t11-fix-01-captures-1792319830-988476/brief.txt`.
That concrete brief controls this check. The stale template grants no finish fields or replay.
I read both complete first-review reports and both complete worker reports through their final markers.
I read the current task, T03/T04 requirements, V1 to V4, ACTIVATION, FIX-01, and the old VERIFIED text.
The exact retained probes, reproduction captures, and new red/green/CLI/gate captures remain unchanged.

Full report: `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t11-verify-02.md`.
Captures: `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t11-verify-02-captures/`.
Capture names below refer to that directory.
The report gives the exact verdict SHA and final direct remote SHA after delivery.

I inspected the full activation range and the complete first-verdict-to-worker diff.
The activation range has 32 paths: 15 source, 10 test, and seven evidence paths.
The fix range has 14 paths: seven source, six test, and FIX-01.
No common, output, store, logger, dependency, baseline, task brief, tracker, source05, or root source change occurs.
ACTIVATION remains the bounded historical evidence exception. The old V3 literal whitespace exception stays historical.
`ResolvedActor` is in-memory plumbing. No new persisted record or JSON envelope field occurs.
No finish sequence, finish key, journal, rollback, or replay mechanism occurs.

## V1 to V4 and fixed findings

| Item | Result | Basis |
| --- | --- | --- |
| V1 / AC1 | PASS | Selected claims, static paths, all identity routes, and task registration pass. |
| V2 / AC2 | FAIL | Direct library finish logs still lose the resolved actor. |
| V3 / AC2, AC3 command scope | PASS | The required transcript and retained real CLI matrices pass. The library failure remains. |
| V4 / AC4 | PASS | The full actual-worktree gate and TestSTE pass. |

| ID | Disposition | Independent result |
| --- | --- | --- |
| P-F01 | CLOSED | A selected cross-claim alias refuses. The legitimate holder still resolves. |
| T11-IO-F04 | CLOSED | Renamed holder/claim files refuse on all routes. Default consume preserves the unread record. |
| P-F02 | CLOSED | Key, name, pane, and history complete succeed. Acting attribution stays exact. Self-verification aliases refuse. |
| P-F03 | CLOSED | All three actual counter-failure retries refuse without writes and include all three inspection commands. |
| T11-IO-F01 | CLOSED | Explicit invalid send creates no state directory. Resolved sender checks stay under the held handle. |
| T11-IO-F02 | CLOSED | Default refusals log the real error without an invented role or message body. |
| T11-IO-F03 | OPEN | Command logs pass. Direct library pane/history calls still omit the resolved actor ID. |
| T11-IO-F05 | CLOSED | Fresh-worker rework reports the current worker on short/error stdout. |
| T-W01 | CLOSED | Removing diagnostic quoting makes the restored control-bearing assertion fail. Restored source passes. |

F04 is one finding. No finding closes from a worker claim alone.

## T11-IO-F03 retained failure

Location: `internal/task/transition.go:226-230,294-296`.
`runFinish` gets the resolved actor only through the caller-supplied optional `ResolvedActor` pointer.
The existing direct library routes do not require that output pointer.

Literal fixture: task `1-literal`, project `one`, acting agent `worker-name`.
The live record has key `worker-key`, name `worker-name`, and pane `worker-pane`.
Call `Complete(home, TransitionOptions{Task: "1-literal", PaneID: "worker-pane"})`.
The durable result is ok with task, counter, inbox, and finished effects.
The actual default log has one transition and four effects. Every event omits `agent_id`.
Expected: each applicable event carries resolved actor `worker-key`.

The same omission affects verify pass/fail and actual task/counter/record/finished failures.
The name route records the supplied name instead of the resolved key.
The explicit-key route and optional-pointer control pass. Known effects and phases remain truthful.
This is not a durable group failure.

Evidence: `direct-library-logging.out`, `direct-default-logger.out`, and both retained logging probes.
`library-logs/default-logger-complete-pane.jsonl` retains the actual default log without logger injection.
`library-logs/complete-counter-pane.jsonl` retains the actual partial log.
The matrix has 45 failed identity subcases. No failure relies on compilation.
Recommendation: retain an unconditional local resolved actor in `runFinish` for its own logs.
Copy that actor to the optional output when supplied. Keep the lock, result, effects, and schema unchanged.
This check stays within the changed logging area and the existing F03 criterion.

## Old checks, red sensitivity, and CLI

All 56 old test functions in changed test files remain. Those files now contain 101 test functions.
The 25-pair status table keeps eight allowed pairs and 17 refusals with unchanged task bytes.
Only advanced finish diagnostics add inspection guidance. The fresh-worker and live-registration fixtures match the installed task.
The diagnostic test retains its task-byte comparison and checks the literal control-bearing resolved name.
This pass removes, skips, or weakens no old test.

One current handle checks identity, role, project, task, claims, FSM, and full notification before all writes.
The order stays task, counter, inbox, then finished. Held inbox calls retain public wrappers and take no nested lock.
All twelve actual storage failures retain failed/no-effects first, then exact partial prefixes and actual phases.
Command callbacks retain nil partial Go errors. Counter gaps remain explicit. No rollback or replay occurs.
Same/changed-reason retries, finished assignment, fresh-worker rework, and cancellation after partials remain safe.
4090 ASCII bytes and 2045 two-byte characters fit exactly. The next tested byte counts refuse before all writes.
No truncation or general reason cap occurs.
Ten race repetitions pass for cooperating readers and competing complete/verify pass/fail attempts.
The controlled related reader waits for the whole group. Output writers that acquire the store lock pass.
No output, Git, socket, process, or long wait occurs under the finish lock.
Default coordinator/manager key routes, no-inbox roles, explicit inbox semantics, and unknown/live/finished show checks pass.

The exact `6dce8b53` snapshot with old production gives assertion-red. The frozen worker gives green.
Eleven focused mutations give assertion-red and restored green for the fixed identity/path/actor/logging/diagnostic behavior.
Every original tracked copy hash matches after restore. The actual source stays frozen.
The unresolved direct-library finding stays red. I apply no source fix.

Exact first-review library probes replay with provenance. The primary real CLI transcript and all twelve actual prefix cases pass.
The secondary command matrix passes twelve text/JSON short/error stdout cases and six partial stdout cases.
The inbox checks retain actual consumed IDs 1 and 3. All-mode failures claim no consumption.
Each JSON failure attempts one envelope, not a second object.
Thirty CLI log checks retain actual effects, IDs, roles, and phases. Stream errors stay separate from durable success.
No message or reason body appears in the checked logs.
The historical secondary expectation that invalid send creates state now fails as expected.
Its corrected replay reverses only that expectation. All binding and fixed-expectation changes retain exact patches.

Actual-root binary SHA256: `dd29bacecbf3a2248947ac2ae7cce3da6c593a3951df982c94f5f1b9d0f05ec2`.
Output-driver SHA256: `37e77325aa0acff320841e54aaaab6b0242b8806bba80da64245149ebecbb85d`.
The exact driver changes only the external output writer. It mocks no internal behavior.
CLI calls use absolute hashed binaries, unique marked fixture homes, and eight-second deadlines.
No native agent or model child starts.

## Gate, freeze, and cleanup

The full actual-worktree gate uses short `TMPDIR=/tmp/t11v2-lhpmauqk`, outside every marked home.
Tracked/new Go formatting prints nothing. Vet passes. Lint with `--allow-serial-runners` reports `0 issues.`.
Full race tests use count 1 and pass. TestSTE passes. The extra relevant package race run passes.
Baseline, ratchet source, and dependencies equal the activation base exactly.
New report/evidence text has zero hard findings with both glossaries. The fix-range diff check passes.
The full historical range retains only the old V3 trailing spaces.

All 249 tracked hashes match before this append.
The old VERIFIED prefix has 14375 bytes and SHA256 `8e720d072b1a227453bd87f7a98198b8d14dcf07e66c7535ca41d466e6816f70`.
This append preserves that exact prefix.
The no-follow cleanup manifest records 4985 entries in two owned obsolete roots.
The receipt records both roots absent: the copy/fixture/binary root and external TMPDIR.
Exact probes, patches, raw outputs, transcripts, hashes, and compact synthetic records/logs remain.
No preexisting report, log, credential, state, worktree, attribution, or harness configuration changes.
This pass changes only this evidence file among tracked paths.

The boundary stays local users, cooperating writers, and static file types.
Hostile Lstat/Open replacement, malicious callbacks, and unrelated process control remain deferred.
UNCOMPLETED: F03 direct-library identity logging and source acceptance.
Recommendation: return F03 for a bounded correction. Preserve all closed findings and passing checks.

END VERIFIER T11-task-identity-and-finish t11-verify-02

# T11 independent fix verification 03

Date: 2026-10-18.
Pass: `t11-verify-03`.
Verdict: **FAIL**. The direct-library F03 defect closes. The home-before-log check fails.
Route: external Pi, `openai-codex/gpt-6.1-sol:high`.
Observed provider, model, and reasoning: `openai-codex`, `gpt-6.1-sol`, `high`.
Session: `01a14eb6-327d-70ac-920f-f34c6554a931`.

## Identities and scope

| Item | Full SHA |
| --- | --- |
| Frozen worker and sole verdict parent | `59b2b4125562bf075e50ab5d020e34d7310dab04` |
| Actual final production | `94ebe8dba0e814d4779df4aa751ee88d32bf5119` |
| Exact test-first snapshot | `9d4ecb234381f21d0ad66c850578ee738088f027` |
| Prior verdict | `18afc0edd1b016e14e1e8819d68f9e12dc6e2304` |
| Prior worker | `0124ab636ffa62dfcb41256801154970528fdb01` |
| Prior production | `44235d664f553c94f3039f8080ec0a534a34fe20` |
| Activation base | `eded8f9cb09208bfdc81334870e86479e746c183` |

Branch: `task/T11-task-identity-and-finish`.
Initial local, tracking, and direct remote tips equal the frozen worker. The worktree is clean before this append.
The concrete fix02 brief SHA256 is `aaa7a8a4ebce50a82201d607b5f3e5ae8f1468231007333a9552ef5584e16265`.
Its path is `.hive/tmp/worker/t11-fix-02-captures-1792321366-1238026554/brief.txt` in the worktree.
I read the complete prior verdicts, the fix02 report through its marker, the concrete brief, requirements, and old evidence.
The worker report's `44235d66` production statement describes the start. Actual final production is `94ebe8db`. Report bytes stay unchanged.
The source diff has 11 added and two removed lines, or 13 changed lines, not 13 added lines.
I inspected every line and all 368 new test lines. The exact required source SHA matches.
The fix02 range has only `transition.go`, new `finish_log_identity_test.go`, and new `FIX-02.md`.
The activation range has 34 paths: 15 source, 11 test, and eight evidence paths.
Unchanged identity, inbox, command, and other paths match the prior verdict blobs. No root/source08/09 merge occurs.
No new persisted field, JSON schema, replay, rollback, generic logger framework, or external wait under the store lock occurs.

Full report: `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t11-verify-03.md`.
Captures: `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t11-verify-03-captures/`.
The report records the final verdict and remote SHAs. Capture names below refer to that directory.

## Results

| Item | Result | Basis |
| --- | --- | --- |
| V1 / AC1 | PASS | Claims, paths, all identity routes, and task registration pass. |
| V2 / AC2 | PASS | Direct logs, durable prefixes, retries, caps, and held checks pass. |
| V3 / command scope | FAIL | The required transcript passes. Pre-callback stdout failure creates an outside-home log. |
| V4 / AC4 | PASS | Full actual-worktree uncached gate and TestSTE pass. |

| ID | Disposition | Independent result |
| --- | --- | --- |
| P-F01 | CLOSED | Selected cross-claim aliases refuse. Legitimate holders still resolve. |
| T11-IO-F04 | CLOSED | Holder/claim path mismatches refuse on all routes and preserve unread records. |
| P-F02 | CLOSED | Key/name/pane/history complete retain stored attribution. Self-verification aliases refuse. |
| P-F03 | CLOSED | All three actual counter-failure retries refuse without writes and name inspection commands. |
| T11-IO-F01 | CLOSED | Explicit invalid send creates no state directory. |
| T11-IO-F02 | CLOSED | Default refusals log the actual error without an invented role. |
| T11-IO-F03 | CLOSED | Both logger routes record resolved keys and actual prefixes, IDs, roles, phases, and stream errors. |
| T11-IO-F05 | CLOSED | Fresh-worker rework reports the current worker on short/error stdout. |
| T-W01 | CLOSED | Control-bearing wrong-actor assertions and old task-byte checks remain and pass. |
| T11-IO-F06 | OPEN | The changed stream-error path starts a relative logger before successful home resolution. |

The exact retained logging probes replay with provenance and no byte changes.
All 72 injected cases pass: complete/verify pass/verify fail, six phases, and key/name/pane/history routes.
The new worker test also checks all four refusal routes. The retained probe repeats its key refusal across route labels.
An additional actual default-file matrix passes all 72 cases without an injected logger or output pointer.
`literal-actor-matrix.json`, `literal-default-events.json`, and `default-log-audit.json` record literal expectations and results.
Success has one transition and four effects. Every applicable actor is the resolved record key, not its name.
Actual failure prefixes, project/task IDs, controlling roles, phases, reason privacy, optional output, and per-call isolation pass.
The default complete/pane sample has five events with `agent_id` equal to `worker-key`.
The exact `9d4ecb23` snapshot with old production gives 54 assertion-red cases. Fixed source gives 72 green cases.
Exact retained probes give 45 red/27 green cases before the fix and 72 green afterward. The default probe also changes red to green.
One owned mutation uses the resolved name instead of key. It gives assertion-red and restored green, with exact source restoration.
No red result relies on compilation. No unrelated mutation or actual-worktree source edit occurs.
The first default-log comparison included unique fixture names in errors. The final audit normalizes only owned names and excludes time.
Raw logs stay exact. This comparison correction changes no product expectation.

## T11-IO-F06: home-before-log failure

Criterion: inherited T01 AC1 home guard and no-creation failure rule, plus this dispatch's explicit pre-callback check.
Locations: `cmd/task/complete/complete.go:28,33-34,40-49` and `cmd/task/verify/verify.go:33,38-39,47-56`.
Home failure prevents the callback that sets `homeUsed`. Failed JSON stdout still calls `LogOutputFailure` with an empty home.
`internal/task/transition.go:279-280` opens the logger. `internal/task/log.go:21` passes the empty home to `NewWriter`.
`internal/lib/log/log.go:45-46` then makes relative `.hive/logs`.
The owned CWDs are below `/tmp/tv3-njozmq16/outside/`, outside every marked home. Home/pane environment values are unset.
The absolute actual binary uses real `/dev/full`. The absolute real Cobra driver also uses actual short/error writers.
Each of complete, verify pass, and verify fail exits 1 but creates CWD `.hive/logs/YYYY-MM-DD.jsonl`.
All nine failed-output cases create one output-error log. No task, state directory, or marker appears.
All nine working-output controls exit 1 with failed/home and empty effects. They create no `.hive`.
Expected: no `.hive` before successful home resolution. Actual: the nine failing-writer cases violate that guard.
`home-guard-results.json`, `home_guard.py`, and `home-guard-logs/` retain exact commands, hashes, outputs, statuses, and events.
Recommendation: guard owned stream-error logging until home resolution succeeds. Preserve valid-home logs and the original pre-callback failure.
Add a focused outside-home regression test. Keep the shared logger framework unchanged. No source fix ran in this pass.

## Previous checklist, CLI, gate, and preservation

All 56 old test functions remain. The changed test files now contain 105 test functions, without new skips or weaker old assertions.
The 25-pair status table keeps eight allowed pairs and 17 refusals with unchanged task bytes.
One handle checks identity/role/project/task/claims/FSM/full notification before task, counter, inbox record, and finished writes.
Held Append/Consume/All and ordinary wrappers keep their semantics and no nested lock. Output and external work stay outside the finish lock.
All 12 real storage prefixes pass. Partial callbacks keep nil Go errors. Counter gaps remain explicit, without replay or rollback.
ASCII 4090/4091 and UTF-8 2045/2046 two-byte cases retain exact accept/refuse behavior before writes.
Same/changed-reason retries, finished assignment, fresh rework, cancel, static types, numeric ordering, defaults, and show checks pass.
Ten race repetitions cover competing finishes, cooperating readers, parallel inbox calls, and held APIs.
Lock-acquiring output writers pass for complete, verify pass/fail, consume, and all.
The retained V3 transcript, 12 CLI prefixes, 12 stdout-failure successes, six partial stdout failures, and four consume/all pairs pass.
Consumed IDs remain 1 and 3. All-mode failures preserve unread marks. Each JSON writer attempts only one envelope.
Thirty command-log audits pass actual effects, phases, IDs, roles, privacy, and stream errors.
Only the previously identified expectation corrections apply. Script provenance and execution-binding patches remain.
The extra script's historical final sentence is stale. Its actual corrected assertions pass.
Actual binary SHA256: `06ad4defca52a08eb00e25d60de76083e168ebc837a3b192c2e7d6f09985b6e2`.
Real Cobra output-driver SHA256: `a7168e3fc5cb75c67b48fdbdec1f33e9eaec9d2624459730e35352ace1ad36b5`.
Every real CLI uses an absolute hashed binary, unique marked fixtures, and eight-second subprocess bounds.

The full actual-worktree gate uses short `TMPDIR=/tmp/tv3-njozmq16`, outside every marked home.
Tracked/new Go format output is empty. Vet passes. Lint with `--allow-serial-runners` reports `0 issues.`.
Full race count1 and separate count1 TestSTE pass. Baseline, ratchet source, and dependencies equal activation exactly.
New report/evidence text has zero hard findings with both glossaries. The fix02 diff check is clean.
The historical range retains only the old V3 literal trailing spaces.
All 251 tracked hashes match before this append. All original tracked copy hashes match after the mutation restore.
The old VERIFIED prefix has 25100 bytes and SHA256 `4594632dfa31e6875af68f8e8bbb51f9745cc9eb05d7efa30a2ba8eae89a1a48`.
This append preserves those exact bytes. The old reports, source, evidence, and runtime remain unchanged.
The no-follow cleanup manifest records 3735 entries in two owned obsolete roots. Both roots are absent.
Only owned copies, fixtures, binaries, and external TMPDIR were removed after compact synthetic evidence and hashes were saved.
No worktree runtime root, preexisting data, credential, attribution, shared record, or harness setting changed.
The boundary remains local users, cooperating writers, and static types. Hostile Lstat/Open replacement remains deferred to T01.
Malicious callbacks and unbounded input remain deferred. No agent/model child, source acceptance, landing, or checkpoint pass occurs.
UNCOMPLETED: F06 correction and fresh independent source acceptance.
Recommendation: return only F06 for a bounded fix. Preserve every closed ID and the accepted finish contract.

END VERIFIER T11-task-identity-and-finish t11-verify-03
