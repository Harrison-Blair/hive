# T10-agent-spawn-core VERIFIED

| Item | Value |
| --- | --- |
| Task | T10-agent-spawn-core |
| Pass | `t10-verify-03` |
| Date | 2026-10-18 |
| Entry | `4227c9583b799111b6402d36b6a550126ad88184` |
| Final Go source | `0eacb8532bc552d650c50aa9dae396bdd11ab43e` |
| Task blob | `5db0e34783fbfa1f73aca232e7e0f8008dff02fd` |
| Agent | verifier, Pi, `openai-codex/gpt-6.1-sol:high` |
| Verdict | **FAIL** |

The current source fails independent verification on `T10-O-F02`.
The other five findings are CLOSED.
The full current finding results and runs are in [FIX02-VERIFIED.md](FIX02-VERIFIED.md).
FIX01-VERIFIED remains exact historical evidence.
No source fix occurred in this verifier pass.

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | PASS | Profile suite and independent literal checks |
| V2 | AC2 | FAIL | Start-session durability, lost pane topology, and false absence after wait timeout |
| V3 | AC3, AC4 | PASS | Production send/stop, real store, identity changes, and CLI checks |
| V4 | AC2, AC3, AC4 native flow | PASS, product smoke only | Fresh final-Go Claude spawn, prompt, and stop |
| V5 | AC5 | PASS | One full actual-worktree gate |

## Current open finding

The exact nine formerly failing endpoint session cases now pass.
The earliest session survives in each final record and result. No close occurs.
The retained v02 matrix and all 16 FIX02 cases pass.

Required independent literal checks still fail:
- A first start reference is not durable before the next pane read.
- Two session-failure paths lose successful pane topology.
- A wait timeout falsely becomes an absent-session observation despite independent matching identity evidence.

Recommendation: persist the first reference before the next external read.
Keep successful pane topology on a session partial path.
Observe session absence only after a successful typed reply.
A source fix and fresh independent verification remain uncompleted.

## Records and limits

Report: `.hive/tmp/coordinator/reports/t10-verify-03.md`.
Captures: `.hive/tmp/coordinator/reports/t10-verify-03-captures/`.
The prior verdict exact bytes remain in `VERIFIED-old.md` in those captures.
Prior verdict SHA256: `9b21c4417946daeea76f2b1abd48f36f95babab5decddb42818b214c58b18900`.
All prior reports and captures remain unchanged.
The complete fix-worker report and supplied completion attestation establish an ended worker, not source acceptance.
No live secondary completion barrier remains.

The exact FIX02 pre-fix replay fails assertions in the nine formerly open cases.
The inherited final tests pass. Two current session mutations fail assertions and pass after exact restoration.
The accepted FIX01 replay and prior mutations received an artifact audit only.
The fresh native child uses installed Claude Code `2.1.302` and `claude-haiku-5-5` for a product smoke only.
The separate pane read shows a distinct `● pong` reply.
Both owned child PIDs and tabs are absent. The source pane retains its original shell.
The compact evidence and exact no-follow inventories remain after owned temporary cleanup.
Manual refusal passes production tests and the retained original native evidence audit.
This pass does not claim a fresh native manual-refusal run.

The original red stub snapshot remains absent. Raw original red logs survive.
Exact original-stub replay remains BLOCKED. No fabricated snapshot replaces it.
The helper import remains exact. Root activation receives an immutable historical audit only.
The local-user/cooperating-writer boundary applies.
Hostile file replacement between `Lstat` and `Open` stays deferred for T01.
No Root landing, main change, worktree removal, CP2 PASS, or CP4 PASS follows from this verdict.

END T10 VERIFY 03
