# T10-agent-spawn-core source summary

| Item | Value |
| --- | --- |
| Task | T10-agent-spawn-core |
| Pass | t10-impl-01 |
| Route | pi, opencode-go/deepseek-v4-flash:high |
| Base | b4ab63ca1943c2c0491ae193392d8f7bdf3ea857 |
| Final source commit | f225bfa9eb4541598edf6f298a743a4f19932736 |
| Branch | task/T10-agent-spawn-core |
| Evidence | V1 to V5 |
| Verdict | PENDING fresh independent verifier |

## Source

The pass adds the harness profile package, the typed herdr agent and pane
methods, the bounded process reader, the spawn, send, stop, and inspection
code, the three command folders, and the narrow record and type fields. The
manual register has no new flag and leaves `spawned` false. The spawn sets
`spawned` true in the first record write.

## Shared logger import

The pass imports the exact verified T01 body-bound tip
`12d25801837c3f0e39261574d504082e990281a5` by an ordinary merge. The merge is
clean. The merge changes only these paths:

- `internal/lib/log/log.go`
- `internal/lib/log/body_bound_test.go`
- `spec/hive-v1/evidence/T01-home-and-register/BODY-BOUND-FOLLOWUP.md`
- `spec/hive-v1/evidence/T01-home-and-register/BODY-BOUND-VERIFIED.md`

The tracker blob stays `bb5ce4fbfac6ab1af65cbf3f0b25d41341f0171f`. The merge
parents are `6747c5a1f96b2aa6e7e9352b7459c083e3b25f03` and
`12d25801837c3f0e39261574d504082e990281a5`.

## Evidence

| Item | Proves | Result |
| --- | --- | --- |
| V1 | AC1 | pass |
| V2 | AC2 | pass |
| V3 | AC3, AC4 | pass |
| V4 | AC2, AC3, AC4 | pass |
| V5 | AC5 | pass |

## Tests

The red runs fail on behavior assertions. The green runs pass. Controlled
mutations cover the harness bounds, the start name, the stop terminal check,
the prompt wait shape, the process start field, and the log body bound. Each
mutation fails a critical test. The exact restore passes.

## Gate

The four gate commands pass at the final Go source. The focused STE check and
the evidence prose check are in the final report and the gate captures.

## Limits

The final source acceptance is pending. A fresh independent verifier must
rerun V1 to V5 and write `VERIFIED.md`. This pass makes no checkpoint or
landing claim. The approved boundary is the local user and a cooperating
writer. Hostile replacement between `Lstat` and `Open` stays deferred for T01.
