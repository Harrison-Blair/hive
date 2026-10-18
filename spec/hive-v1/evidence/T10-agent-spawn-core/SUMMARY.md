# T10-agent-spawn-core SUMMARY

| Item | Value |
| --- | --- |
| Task | T10-agent-spawn-core |
| Pass | t10-fix-01 |
| Date | 2026-10-18 |
| Entry | `440049aecc5fde6b8a7df28f601d4aeaa365f8c4` |
| Tested source | `d55e9de6ec441313884f29e80188f5fd96cb22a4` |
| Source tree | `ef3c910346fd8dbf9b23cbd8297d4d1b71669eb7` |
| Task blob | `5db0e34783fbfa1f73aca232e7e0f8008dff02fd` |
| Verdict | fresh independent verification PENDING |

This fix closes the six open findings of the first independent review.
The prior `VERIFIED.md` stays unchanged with its FAIL verdict.
A fresh independent verifier must replace that verdict.

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | PASS | Unchanged and re-run by the full gate |
| V2 | AC2 | PASS | New red replay and green run |
| V3 | AC3, AC4 | PASS | New red replay and green run |
| V4 | AC2, AC3, AC4 | PASS | Fresh native Claude flow at the tested source |
| V5 | AC5 | PASS | One full gate at the final Go source |

## Findings

| ID | Fix | Evidence |
| --- | --- | --- |
| T10-O-F01 | Inspect writes `exited` only for a proved same-terminal shell | `TestFix01Inspect`, V2 and V3 |
| T10-O-F02 | Spawn keeps the first start session and compares later sessions | `TestFix02SpawnRetainsStartSession`, V2 |
| T10-O-F03 | Stop rechecks the ownership tuple under the final held handle | `TestFix03StopRefusesChangedIdentity`, V3 |
| T10-O-F04 | The herdr package map names `agent.go` and `pane.go` | `internal/lib/herdr/doc.go`, V5 gate |
| T10-P-F01 | Local wire decoders reject malformed success replies | `internal/lib/herdr/fix01_test.go`, V2 and V3 |
| T10-P-F02 | Send refuses an unrepresentable wait timeout before logging or sending | `TestFix02WaitTimeoutBoundary`, V3 |

## Captures

Fix captures: `.hive/tmp/coordinator/reports/t10-fix-01-captures/`.
The fix report is `.hive/tmp/coordinator/reports/t10-fix-01.md`.
The capture tree holds prior bytes and hashes, red replay, green runs,
mutations, gate output, native output, source maps, and cleanup receipts.
The tested binary SHA256 is
`e0f683e54e2374970cc132c08370e9dbe5ee8821852bc232cf1df79a67ea3aa9`.

## Limits

The trusted boundary is the local user and a cooperating writer.
Hostile replacement between `Lstat` and `Open` stays deferred for T01.
The original `/tmp/t10red.lZnMvqUZ` stub snapshot is absent.
The original red logs remain assertion evidence.
This pass does not replay that missing stub source.
The native trust action can leave a Claude trust entry for the removed fixture path.
This pass claims no Root landing, CP2 acceptance, or CP4 acceptance.
