# T10-agent-spawn-core VERIFIED

| Item | Value |
| --- | --- |
| Task | T10-agent-spawn-core |
| Pass | `t10-verify-02` |
| Date | 2026-10-18 |
| Entry | `892d643889d303db064afc21e01df2a018805589` |
| Final Go source | `d55e9de6ec441313884f29e80188f5fd96cb22a4` |
| Task blob | `5db0e34783fbfa1f73aca232e7e0f8008dff02fd` |
| Agent | verifier, Pi, `openai-codex/gpt-6.1-sol:high` |
| Verdict | **FAIL** |

The current source fails independent verification on `T10-O-F02`.
The other five findings are CLOSED.
The full current finding results and runs are in [FIX01-VERIFIED.md](FIX01-VERIFIED.md).
No source fix occurred in this verifier pass.

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | PASS | Profile suite and independent literal checks |
| V2 | AC2 | FAIL | Later missing sessions pass, and partial paths lose the earliest session |
| V3 | AC3, AC4 | PASS | Production send/stop, real store, identity changes, and CLI checks |
| V4 | AC2, AC3, AC4 native flow | PASS, product smoke only | Fresh final-Go Claude spawn, prompt, and stop |
| V5 | AC5 | PASS | One full actual-worktree gate |

## Current open finding

`T10-O-F02` requires ordered first-session persistence and comparison.
Production spawn returns `ok/live` when a later absent session has another matching source.
Some unresolved partial paths also lose the earliest observed session reference.
The independent matrix records nine assertion failures with the known started-agent effect intact.
Controls pass for optional absence before the first present session.

Recommendation: persist the first present reference before the next external read.
Check each later observation, including absence after that first reference.
Keep the earliest reference and known effects on partial outcomes.
A source fix and fresh independent verification remain uncompleted.

## Records and limits

Report: `.hive/tmp/coordinator/reports/t10-verify-02.md`.
Captures: `.hive/tmp/coordinator/reports/t10-verify-02-captures/`.
The prior verdict exact bytes remain in `VERIFIED-old.md` in those captures.
Prior verdict SHA256: `fe4c4cea543bad41def8483502e2e3de1a921545177875a06a46fad59416952b`.
The original four reports and all old captures remain unchanged.
The complete fix-worker report and supplied completion attestation establish an ended worker, not source acceptance.
No live secondary completion barrier remains.

The exact fix pre-fix replay fails assertions for the five behavior findings.
The inherited final tests pass. Five current mutations fail assertions and pass after exact restoration.
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

END T10 VERIFY 02
