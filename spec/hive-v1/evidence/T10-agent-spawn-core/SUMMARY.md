# T10-agent-spawn-core SUMMARY

| Item | Value |
| --- | --- |
| Task | T10-agent-spawn-core |
| Pass | `t10-fix-03` after `t10-fix-02` and `t10-fix-01` |
| Date | 2026-10-18 |
| Entry | `3a0f006e04edbdcc16f864f4afd6b7a4533b6e94` |
| Tested source | `5d8d3b528b68a2a5ee7652aa4a69129fd437b109` |
| Verdict | fresh independent verification PENDING |

This pass closes the open `T10-O-F02` gaps of `t10-verify-03`.
The other five findings stay CLOSED.
The prior `VERIFIED.md` stays unchanged with its FAIL verdict.
A fresh independent verifier must replace that verdict.

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | PASS | Unchanged and re-run by the full gate |
| V2 | AC2 | PASS | FIX03 red replay, tested green run, and full spawn suite |
| V3 | AC3, AC4 | PASS | Unchanged and re-run by the full gate |
| V4 | AC2, AC3, AC4 | PASS | Fresh native Claude flow at the tested source |
| V5 | AC5 | PASS | Full gate and the uncached race run at the tested source |

## Findings

| ID | Result | Evidence |
| --- | --- | --- |
| T10-O-F01 | CLOSED | `t10-verify-02`, retained independent inspection cases |
| T10-O-F02 | FIXED, verification PENDING | `FIX03.md`, `fix03_test.go`, V2, V4 |
| T10-O-F03 | CLOSED | `t10-verify-02`, retained stop identity matrix |
| T10-O-F04 | CLOSED | `t10-verify-02`, `doc-map-audit.json` |
| T10-P-F01 | CLOSED | `t10-verify-02`, retained wire checklist |
| T10-P-F02 | CLOSED | `t10-verify-02`, retained timeout boundaries |

## F02 dispositions

| Gap | Result at `3a0f006` | Result at `5d8d3b5` |
| --- | --- | --- |
| A. Start session before the read | no durable session | `s` durable before `pane.get` |
| B. Pane topology on a session refusal | topology lost | topology and `s` kept, `unresolved` |
| C. Wait timeout with matching session | false missing session, `unresolved` | `partial`, record `live`, identity kept |

## Captures

Fix captures: `.hive/tmp/coordinator/reports/t10-fix-03-captures/`.
The capture tree holds the red replay, the green run, the V2 suite, the gate,
the uncached gate, the native transcript, the model proof, and the cleanup
receipt. It also holds the earlier V2, V4, V5, and SUMMARY bytes.
The tested binary SHA256 is
`d4ba3d50eed28383f2d2089e842945d74d8a1dbd9353c0221bcfd3b9b59ffb51`.
The tested binary is `/tmp/t10fx03.WIPUhYHp/hive-1792342212`.
The fixture is removed.

## Limits

No existing test encoded the buggy behavior, so this pass corrected no test.
`TestSpawnWaitTimeoutWithOwnershipKeepsLivePartial` stays a required case.
The trusted boundary is the local user and a cooperating writer.
Hostile replacement between `Lstat` and `Open` stays deferred for T01.
The native trust action can leave a Claude trust entry for the removed path.
This pass claims no Root landing, CP2 acceptance, or CP4 acceptance.
No independent acceptance exists for this fix.

END T10 FIX 03
