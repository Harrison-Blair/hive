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
