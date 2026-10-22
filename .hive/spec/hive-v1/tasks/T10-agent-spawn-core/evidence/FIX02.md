# T10-agent-spawn-core FIX02

| Item | Value |
| --- | --- |
| Task | T10-agent-spawn-core |
| Pass | `t10-fix-02` |
| Date | 2026-10-18 |
| Entry | `153878052acf39592524ce8f7ddbcce798bacb96` |
| Tested source | `0eacb8532bc552d650c50aa9dae396bdd11ab43e` |
| Branch | `task/T10-agent-spawn-core` |
| Agent | worker, Pi, `opencode-go/deepseek-v4-flash:medium`, effective `high` |
| Verdict | independent verification PENDING |

This pass fixes the open first-session finding `T10-O-F02` of `t10-verify-02`.
The other five findings stay CLOSED.
No independent acceptance follows from this report.
No Root landing, CP2 pass, or CP4 pass follows from this report.

## Fix

`internal/agent/spawn.go` reads the session references of start, pane, wait, and get.
It keeps the earliest present reference.
A later reply with no reference gives `unresolved` after the first present reference.
A later reply with a different reference gives `unresolved`.
It writes the earliest reference to the record before the next external read.
It keeps the earliest reference on every partial path.
An absent session before the first present reference stays valid.
A matching later reference stays valid.
No reference at any successful observation stays valid.
`captureIdentity` and `mergeSessions` keep their old signature and behavior.
The start, stop, and send behavior does not change.
The shared wire decoders and `types.go` do not change.
Each durable write uses the existing held-handle mutation.
Each socket wait and process wait stays outside the store lock.
A failed store write adds its error to the envelope. No error claims a durable write.

## Cases

The literal first reference is `s`. The different reference is `other`.
Each case calls production `Spawn`, the real store, and a fake external herdr peer.

| Sequence or failure | Result at entry | Result now |
| --- | --- | --- |
| Start `s`, pane absent, wait and get `s` | `ok/live` | `unresolved`, keeps `s` |
| Start and pane `s`, wait absent, get `s` | `ok/live` | `unresolved`, keeps `s` |
| Start, pane, and wait `s`, get absent | `ok/live` | `unresolved`, keeps `s` |
| First reference at wait, get absent | `ok/live` | `unresolved`, keeps `s` |
| Start `s`, pane `other` | `unresolved`, empty reference | `unresolved`, keeps `s` |
| Pane `s`, wait `other` | `unresolved`, keeps `s` | `unresolved`, keeps `s` |
| First reference at wait, get `other` | `unresolved`, empty reference | `unresolved`, keeps `s` |
| Start `s`, pane read fails | `unresolved`, empty reference | `unresolved`, keeps `s` |
| First reference at wait, process read fails | `unresolved`, empty reference | `unresolved`, keeps `s` |
| First reference at get, process read fails | `unresolved`, empty reference | `unresolved`, keeps `s` |

The controls pass in both runs.
The controls are no session, and a first reference at start, pane, wait, or get.
A failed get reply keeps the first reference and stays `ok`.
Every case keeps the started-agent effect.
Every case sends no close request.

## Test first

`internal/agent/fix02_test.go` is new.
It holds the 16 literal cases above.
The red run used exact entry production in an owned copy.
The new test file was the only added file.

```sh
go test -count=1 ./internal/agent/ -run TestFix02Session -v
```

Exit status 1.
The nine fixed paths each fail on an assertion.
The controls pass.
The failures are behavior assertions, not compile errors.
Captures: `.hive/tmp/coordinator/reports/t10-fix-02-captures/red-fix02.*`.

The green run used tested source `0eacb853`.

```sh
go test -count=1 ./internal/agent/ -run TestFix02Session -v
```

Exit status 0. All 16 cases pass.
Captures: `green-agent.*`.

## Mutation

One mutation drops the missing-observation rule from `sessionTrack.observe`.
The mutant fails four assertions: the four later-absent cases.
The mutant file and the original file have exact retained bytes.
The exact restore gives the old SHA256 and a green run.
Captures: `mutant-fix02.*`, `restore-fix02.*`, and `mut/*`.

## Probe replay

The exact independent probe of `t10-verify-02` ran in the owned copy at `0eacb853`.
The file is `.hive/tmp/coordinator/reports/t10-verify-02-captures/internal_agent_v02_test.go.txt`.

```sh
go test -count=1 ./internal/agent -run TestV02 -v
```

Exit status 0.
All 15 session cases pass.
The 15 stop identity cases and the cosmetic control pass.
Captures: `probe-v02.*`.

## Checklist

| Check | Result |
| --- | --- |
| Previous six-finding independent checklist | PASS |
| Session cases and the literal reference order | PASS |
| Process identity, claim, and effect suites | PASS |
| Stop identity, refusal, and close order | PASS |
| Wire decoders, timeout bounds, and malformed replies | PASS |
| Logging, name-absent log event, and CLI suites | PASS |
| STE baseline and instruction text | PASS |

The commands and their raw output are in `checklist-independent.*`, `checklist-protocol.*`, and `broad-tests.*`.
This pass did not remove, skip, or weaken a test, case, or assertion.

## Native and gate

The fresh native V4 flow, gate V5 run, binary SHA256, build revision, and source map are in V4, V5, and the captures.
The binary SHA256 is `ffd45fe7f8a45e84493d1fdf32b6f2c26969bfdde8edb6ef768bc89659f59679`.
The build revision is `0eacb853` with `vcs.modified=false`.
The full gate ran once in the actual worktree at tested source `0eacb853`.

## Limits

The trusted boundary is the local user and a cooperating writer.
Hostile replacement between `Lstat` and `Open` stays deferred for T01.
The original stub snapshot of the first red run stays absent.
The fix red run is exact and separate from that historical limit.
The native trust action can leave a Claude trust entry for the removed fixture path.
Independent verification of this fix remains pending.

END T10 FIX 02
