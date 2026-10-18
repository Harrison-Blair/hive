# Machine handoff 01

**Start here:** New agent, read this file, [AGENTS.md](AGENTS.md), [spec/hive-v1/README.md](spec/hive-v1/README.md), and [spec/hive-v1/05-tracker.md](spec/hive-v1/05-tracker.md). Get `origin` without reset. Check the actual `dev` and task references and your local environment. Load the `orchestrate` and `asd-ste100` skills. Then resume the bounded T10 fix.

This portable report records the machine-switch state. It adds no role policy and no skill. The authority stays in [AGENTS.md](AGENTS.md), the spec, the tracker, and the task briefs. This pass writes the handoff only. It makes no source change, checkpoint run, branch merge, archive fix, or resource removal.

## 1. Unfinished objective

The objective is to finish all hive-v1 tasks and all automated checkpoints. The next session is the orchestrate leader and holds the coordinator role. It delegates each content change to a worker and each independent check to a verifier. The existing `AGENTS.md`, spec, tracker, and task briefs are the authority.

The user authorized autonomous completion. No user presentation or demo approval gate applies. The broader code review is later, at the end of v1. Work stopped at this machine switch for this handoff and is not complete.

## 2. Model routes and limits

| Work | Route | Rule |
| --- | --- | --- |
| Implementation | Pi `opencode-go/deepseek-v4-flash:medium` | Effective level is `high`. Record the requested and effective levels. |
| Independent source verification | Pi `openai-codex/gpt-6.1-sol:high` | Use a fresh verifier for each pass. |
| Selected actual task verification | Native Claude Code `claude-opus-5-5 --effort high --dangerously-skip-permissions` | Not Pi. Opus verifies now. The broad review is later. |
| Native product interoperability | Small Claude smoke, currently `claude-haiku-5-5` | Product smoke only. It never implements or verifies Hive source. |

Actual Opus 5.5, Claude Code `2.1.302`, in bypass mode, verified T08 Escape with PASS. Do not add `--bare` because OAuth must stay available. Do not change global harness configuration. At most three external models run beside the coordinator. This limit includes each native product smoke child. Prove the exact installed model before a live run.

## 3. Pass procedure

Use a fresh agent for each pass. Run a fresh no-tools brief preflight before dispatch. Require a complete report and a completion check against the actual final model response. Then stop the agent.

The first large verification splits by area. Each later pass rechecks the fixes and the old checklist. Use only the local-user and cooperating-writer boundary. Check static malformed paths and types. Do not run hostile file-replacement or kernel campaigns. Hostile replacement between `Lstat` and `Open` stays deferred for T01.

## 4. Current Git state

| Item | Value |
| --- | --- |
| Root checkout | `/home/penguin/source/hive` |
| Pre-handoff `dev`, tracking, and direct remote | `ecbd9379efbdea26672726caf4cbf5f9d2edfd8a` |
| Handoff commit | A child of `ecbd9379`. This report cannot contain its own unknown SHA. |
| Root tree | `3a7fbe3718d9ba376f7cb6b589098e64c25e3d42` |
| Root branch | `dev`, clean at entry |
| Local `main` | `3692162633f896b8fee6197469adba91f666a00c` |
| `origin/main` and direct remote `main` | `928cfa95e0906315fcb4a65a5cdbdef1470a0503` |
| T10 branch | `task/T10-agent-spawn-core` at `3a0f006e04edbdcc16f864f4afd6b7a4533b6e94` |
| T01 helper branch | `task/T01-log-body-bound` at `12d25801837c3f0e39261574d504082e990281a5` |
| T08 Escape branch | `task/T08-escape-clear` at `d087123cf82a76361883da7ef0eda9f7d651d1c1` |

The T10, T01 helper, and T08 branch references equal local, tracking, and direct remote. No other Root writer remains. Do not force-push. Do not touch `main`. The two `main` values differ by intent on the old machine. Do not force them into another clone.

## 5. Task and checkpoint status

| State | Tasks |
| --- | --- |
| Landed | T01, T02, T03, T04, T05, T06, T07, T08, T09, T11, T22 |
| In progress, not accepted | T10 |
| Not started | T12 through T21, except T11 |

CP1 passed. Do not rerun CP1. CP2, CP3, CP4, and CP5 are open. CP4 does not block wave 4.

The user authorizes a T19 skip only, for a Cursor quota limit, with actual evidence. Do not skip another task.

## 6. T10 final source and open finding

| Item | Value |
| --- | --- |
| Initial source | `6747c5a1f96b2aa6e7e9352b7459c083e3b25f03` |
| Helper merge | `dca0df1a` |
| Fix 01 source | `d55e9de6` |
| Fix 02 and final Go source | `0eacb8532bc552d650c50aa9dae396bdd11ab43e` |
| Evidence commit | `4227c9583b799111b6402d36b6a550126ad88184` |
| Final verdict tip | `3a0f006e04edbdcc16f864f4afd6b7a4533b6e94` |
| Sole parent of the verdict tip | `4227c9583b799111b6402d36b6a550126ad88184` |

The coordinator read the complete final source report through `END T10 VERIFY 03`, the committed `VERIFIED.md` and `FIX02-VERIFIED.md`, and the actual complete final model response. Source print session `52969` exited 0 with `PI_EXIT0`. This is the coordinator acknowledgement required by the first contract checker. The original verifier PID `2544427`, start `12975243`, ended. The coordinator checked its absence independently.

The final response records FAIL for F02, a push of `3a0f006e`, a clean branch, and complete cleanup. The verifier closed the other five findings, and the gate and native flow pass. The five closed findings are inspection, stop tuple, document map, malformed wire success, and timeout overflow. F02 stays open as the final report says.

The task branch carries V1 to V5, FIX01, FIX02, and the latest verdict in Git. Ignored raw probes do not travel in Git.

### F02 gaps

| Gap | Actual result | Expected result |
| --- | --- | --- |
| A. First start session `s` | At the next `PaneGet` callback, the durable record still has no session reference. Location: `internal/agent/spawn.go:124-125`. | Save `s` before the next external read. |
| B. Successful typed pane topology | Start `s` with pane session absent, and start `s` with pane session `other`, lose terminal `term_a`, workspace `w1`, tab `w1:t1`, and cwd `/tmp/fixture`. Location: `spawn.go:133-145`. | Keep the earliest `s` and the known topology on the partial or unresolved result. |
| C. Wait timeout | With matching start, pane, and get session `s`, and a valid owned PID and start, the wait timeout becomes a false missing-session Unresolved. Location: `spawn.go:157` and `spawnWaitFailure`. | Return partial/LIVE with independent ownership evidence. Only a successful typed reply supplies a session observation. Keep a missing display name and a failed get transport distinct. |

A wait timeout is an ERROR, not a successful typed reply. The accepted readiness-timeout result stays partial/LIVE. No one dispatched a source fix after verification 03.

Suggested bounded fix: save the earliest start session before `PaneGet`. Save the successful `PaneGet` terminal, workspace, tab, and CWD before an unresolved return on session absence or mismatch. Preserve the earliest reference, the partial effects, and the store-error truth. Keep the shared helpers compatible and keep external waits outside the store lock.

T10 still needs the new fix, a fresh independent verification, Root landing and integration, and then CP2. A passing gate and native flow alone do not accept the source.

Literal behavior for new regression tests:

- Start `s`, pane absent, then wait and get `s` gives unresolved with `s` and known topology.
- Start `s`, pane `other`, gives unresolved with `s` and known topology.
- Matching start, pane, and get `s`, with owned PID and start evidence, and a wait timeout gives partial/LIVE.
- Store failure gives partial with the actual store error and no claim of an unwritten session.

## 7. T01 shared body helper

The accepted helper tip is `12d25801837c3f0e39261574d504082e990281a5` on `task/T01-log-body-bound`. It fixes the decoded UTF-8 body cap of 4096 bytes and truthful truncation. T10 imported exactly four paths from it in `dca0df1a`. Avoid a separate Root landing unless a dependency needs it. The helper has not reached Root. After T10 passes, one Root merge can include both. The old helper worktree remains for archive and removal.

## 8. T08 Escape and archive R01

| Item | Value |
| --- | --- |
| Accepted source | `25f20261` |
| Opus verdict | `d087123c` |
| Root merge | `1dea4c56` |
| Docs receipt | `2c79f10e` |

The pass archived worktree `wBA` and removed it without force. Root integration passed. The original archive R01 had two defects. Its payload included `.git`. Its report and state record gave a wrong hash.

Corrected archive `r01` has 596 payload entries and 318 tracked matches. Its pointer metadata gives the actual SHA `00e72cf14448b73d1c0f71e0e05d05abe2bfbfc3c4bf875efff307878ff4b8ab`. The 597-entry original stays.

The correction commit `ecbd9379` exists. Its independent acceptance is PENDING.

## 9. CP4 and B01

The CP4 source Escape fix exists. The earlier actual-home run cleared the footer and matched the header, event, and exit 0. Its actual info view was empty because all 71 daily events were DEBUG.

B01 data is ready. It used 81 real failed read-only `task show` calls. Each call exited 1 with `failed`/`read` and empty effects. No task record exists for these IDs: `0-cp4-project-r01-01` through `0-cp4-project-r01-80`, and `0-cp4-other-r01-81`. The calls produced 81 ERROR events. The daily log went from 71 to 314 lines, plus handoff metadata.

A fresh independent R01 and CP4 verification is still pending. Do not claim CP4 passed. New machine logs can differ. Do not append raw synthetic JSONL or copy false production evidence.

The literal CP4 keys are `/`, `project`, Enter, `l`, `f`, `g`, `G`, Escape, and `q`. It needs an actual-home header, a real project-list event, a visible `g` and `G` result, and a returned nonproject event. It also needs a separate child status of 0 and owned-resource absence.

## 10. Recommended next order

1. Fix T10 and run a fresh independent verification. Run the independent R01 and CP4 check in parallel with slot reservation.
2. Land T10 and the helper, then verify Root integration.
3. Run CP2 and get its independent acceptance.
4. Start T12 and T18 first. Start T13 before T14 to avoid shared `spawn.go` edits.
5. CP3 needs T16. T15 is large. Start T17 and T19 before CP5.
6. Run CP3 after T12, T13, T14, T15, T16, and T18.
7. Run T20 and T21 after CP3. Run CP5 last.
8. Use the T19 quota skip only with actual evidence.

## 11. Ownership gaps before dispatch

These gaps came from a Root read. They are not installed and not implemented. Each needs a checked brief or spec ownership clarification before a worker starts. Do not dispatch speculative code.

| Task | Gap | Decision |
| --- | --- | --- |
| T12 | `run.Command` always caps a run at 30 seconds. A parent context of 60 or 120 seconds does not meet the merge 60-second and push 120-second bounds. | Add a narrow explicit-timeout runner support API. Keep the existing `Command` behavior at 30 seconds. Add the `namedlock.go` map line to `internal/lib/store/doc.go`. |
| T13 | `PaneInfo` has no `RestoreError` field. The AC needs `restore_error`, and the task source table omits `types.go`. | Add optional `RestoreError` as `string` or `null`, with local wire forwarding and type checks. The existing T10 placeholder is already unresolved. The task needs no old source correction. |
| T15 | Public `agent.Register` locks internally and `registerLocked` is private. The singleton check and the registration must use the same handle. | Add a narrow held registration API or an internal `PIDStart` option. Avoid nested locks and duplicated generation rules. Write `pid_start` in the first claim write. Preserve existing registration semantics. |
| T16 to T19 | The V2 text says that `send` prints a reply, but `send --wait` reports a settled status. | Clarify the wording. Prove an actual pong with a separate native agent or pane read. Do not add a T10 read API. |

## 12. Dogfood transition

After T09, T10, T12, and T18 land, use the available Hive task, worktree, spawn `--pane`, send, stop, land, and remove commands immediately. The old friction workaround was about missing capabilities. It does not permit waiting until CP3. Record an honest mapping from each runtime task ID to its spec ID.

CP5 needs an actual record that the manager starts its first worker before the coordinator removes the bootstrap clause. A CP3 runner dispatch alone is not that record.

## 13. Existing decisions

- The shared state store has a cancellable default lock wait of 2 seconds. An explicit context deadline can lengthen it.
- Grouped related reads lock. A single `Read` or `List` keeps its behavior.
- The store keeps partial progress with no journal and no rollback. Each operation handles an incomplete prefix.
- External waits stay outside the store lock.
- The native wire field is `agent_info.agent`. The old `agent_info{agent_info}` form is wrong.
- A manual registration is not spawned. A stop refuses it.
- The small Claude smoke uses `claude-haiku-5-5` for product interoperability only. Prove the exact installed model.

## 14. Test and cleanup discipline

Write the assertion first and see it fail before implementation. Deliberately break a behavior and restore the exact source. Run the full `AGENTS.md` gate once at final Go. Use a short `TMPDIR` outside marked homes. Later docs-only passes use equality checks and focused checks. Never weaken, skip, or delete a test.

Archive each worktree payload, mode, and SHA, and each accepted Git blob, before a no-force removal. Keep the branches. New-machine resources need a fresh identity. Old IDs are historical only.

## 15. Local inventory and transfer limits

| Local item | Value | Rule |
| --- | --- | --- |
| Root | Old path `/home/penguin/source/hive` | Historical. A new clone may differ. |
| T10 worktree | `wB4:p1`, terminal `term_65e1cd29b3f1a30` | Preserve. |
| Helper worktree | `wB8:p1` | Preserve. |
| User workspace | `w9X:p4` and `w9X:p5` | Protect. Send no input. |
| User Codex session | Preserve | It is not a worker. Do not stop it. |

All models the coordinator started ended before the switch, if checked. Ignore `projects/`, the `.hive` runtime, reports, and archives, and the personal `handoff.md`. They do not travel through Git. Never assume that they exist on the new machine. Commit no secret and no runtime data.

## 16. Old-machine reports and state

| Item | Path | Transfer |
| --- | --- | --- |
| T10 verifier report | `projects/hive/worktrees/T10-agent-spawn-core/.hive/tmp/coordinator/reports/t10-verify-03.md` | Local only |
| T10 fix report | `projects/hive/worktrees/T10-agent-spawn-core/.hive/tmp/coordinator/reports/t10-fix-02.md` | Local only |
| T08 Root integration | `.hive/tmp/coordinator/reports/t08-escape-integration-01.md` | Local only |
| T08 records fix | `.hive/tmp/coordinator/reports/t08-records-fix-01.md` | Local only |
| Accepted source evidence | `spec/hive-v1/evidence/` | Tracked |
| Current state | `.hive/tmp/coordinator/state/wave-status.json` | Local only |
| Historical INPROGRESS text | Section 5 of `handoff.md` | Local only |

These paths are relative to the Root checkout.

## 17. Historical limits

The original T10 red stub snapshot is absent. Raw logs survive. Fresh fix snapshots and mutations are separate evidence. The old V4 provenance mismatch, `9ab35c26` against claimed `f225`, gave way to clean independent native runs. T08 H04 missing historical runtime data cannot be restored by later archives. Do not fabricate a replay.

One exact old automatic harness log may remain at `/tmp/pi-bash-cfefc4e403523a9d.log` from the logger body reproduction. It was not in later cleanup custody. A future checked archive worker may keep and hash that exact known file, then remove only that file. Never glob other Pi logs.

Do not make a main change, force-push, co-author, or false session trailer. Preserve historical attribution.

## 18. Portability and herdr

`plan/hive-v1` and `integration/hive-v1` are historical branches. They are not imports. Ignored old-machine archive verification cannot be claimed from a fresh clone. Missing historical captures are a stated limit. A new agent can recreate the source regressions from the literal outcomes and the tracked task-branch tests.

Inspect the actual installed herdr before a live run. `reference/herdr/` is authoritative for version 0.9.5 and protocol 22. If the installed behavior is newer, run the procedure in [reference/herdr/RESYNC.md](reference/herdr/RESYNC.md) before you trust it.
