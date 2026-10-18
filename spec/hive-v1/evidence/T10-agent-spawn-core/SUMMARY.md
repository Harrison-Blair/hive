# T10-agent-spawn-core SUMMARY

| Item | Value |
| --- | --- |
| Task | T10-agent-spawn-core |
| Pass | t10-fix-02 |
| Date | 2026-10-18 |
| Entry | `153878052acf39592524ce8f7ddbcce798bacb96` |
| Tested source | `0eacb8532bc552d650c50aa9dae396bdd11ab43e` |
| Source tree | `36f3f4b3194ff58136edf4f3e46dc8bee8e89247` |
| Task blob | `5db0e34783fbfa1f73aca232e7e0f8008dff02fd` |
| Verdict | fresh independent verification PENDING |

This pass fixes the open first-session finding `T10-O-F02` of `t10-verify-02`.
The other five findings stay CLOSED.
The prior `VERIFIED.md` stays unchanged with its FAIL verdict.
A fresh independent verifier must replace that verdict.

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | PASS | Unchanged and re-run by the full gate |
| V2 | AC2 | PASS | Entry red replay, tested green run, and probe replay |
| V3 | AC3, AC4 | PASS | Unchanged and re-run by the full gate |
| V4 | AC2, AC3, AC4 | PASS | Fresh native Claude flow at the tested source |
| V5 | AC5 | PASS | One full gate at the tested source |

## Findings

| ID | Result | Evidence |
| --- | --- | --- |
| T10-O-F01 | CLOSED | `t10-verify-02`, retained independent inspection cases |
| T10-O-F02 | FIXED, verification PENDING | `TestFix02Session`, V2, and the probe replay |
| T10-O-F03 | CLOSED | `t10-verify-02`, retained stop identity matrix |
| T10-O-F04 | CLOSED | `t10-verify-02`, `doc-map-audit.json` |
| T10-P-F01 | CLOSED | `t10-verify-02`, retained wire checklist |
| T10-P-F02 | CLOSED | `t10-verify-02`, retained timeout boundaries |

## Captures

Fix captures: `.hive/tmp/coordinator/reports/t10-fix-02-captures/`.
The fix report is `.hive/tmp/coordinator/reports/t10-fix-02.md`.
The capture tree holds old bytes and hashes, the entry red replay, and the tested green run.
It also holds the mutation, the probe replay, gate output, native output, the source map, and the cleanup receipt.
The tested binary SHA256 is
`ffd45fe7f8a45e84493d1fdf32b6f2c26969bfdde8edb6ef768bc89659f59679`.

## Limits

The trusted boundary is the local user and a cooperating writer.
Hostile replacement between `Lstat` and `Open` stays deferred for T01.
The original `/tmp/t10red.lZnMvqUZ` stub snapshot is absent.
The original red logs remain assertion evidence.
This pass does not replay that missing stub source.
The native trust action can leave a Claude trust entry for the removed fixture path.
This pass claims no Root landing, CP2 acceptance, or CP4 acceptance.
No independent acceptance exists for this fix.

END T10 FIX 02
