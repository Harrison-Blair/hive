# T10-agent-spawn-core FIX02 verification

| Item | Value |
| --- | --- |
| Pass | `t10-verify-03` |
| Date | 2026-10-18 |
| Entry | `4227c9583b799111b6402d36b6a550126ad88184` |
| Final Go source | `0eacb8532bc552d650c50aa9dae396bdd11ab43e` |
| Agent | verifier, Pi, `openai-codex/gpt-6.1-sol:high` |
| Verdict | **FAIL** |

`T10-O-F02` remains OPEN. The other five findings remain CLOSED.
The nine required endpoint session failures now pass.
The required durability, topology, and session-timeout checks still fail.
The gate and fresh native product smoke pass. They do not replace behavior evidence.
No production source or old test changed in this pass.

## Findings

| Finding | Result | Evidence |
| --- | --- | --- |
| T10-O-F01 | CLOSED | Production Send, real store, six inspection cases, no close |
| T10-O-F02 | OPEN | Start-session durability, lost pane topology, false absence after wait timeout |
| T10-O-F03 | CLOSED | Exact v02 critical identity matrix and cosmetic control |
| T10-O-F04 | CLOSED | Direct old-map preservation and new map entries |
| T10-P-F01 | CLOSED | Required typed fields, valid controls, malformed-success CLI refusal |
| T10-P-F02 | CLOSED | Literal timeout boundaries, immediate peer, zero invalid prompt requests |

## Session matrix

All nine sequences return partial/unresolved and keep earliest `s` in the store and result.
Each keeps the started-agent effect and sends zero close requests.

| Sequence | Endpoint session | Known pane topology |
| --- | --- | --- |
| Start `s`, pane absent, wait/get `s` | PASS | FAIL, lost |
| Start/pane `s`, wait absent, get `s` | PASS | PASS |
| Start/pane/wait `s`, get absent | PASS | PASS |
| First at wait, get absent | PASS | PASS |
| Start `s`, pane `other` | PASS | FAIL, lost |
| First at wait, get `other` | PASS | PASS |
| Start `s`, pane read fails | PASS | Not observed |
| First at wait, process read fails | PASS | PASS |
| First at get, process read fails | PASS | PASS |

Controls pass for no session and matching observations starting at start, pane, wait, or get.
Optional absence before the first present reference stays valid.
The exact retained v02 probe passes all 15 session cases and the stop identity matrix.
All 16 final FIX02 cases pass.
A failed get transport reply remains log-only. The first wait reference is durable before that failed get.
The storage-failure control reports a truthful partial prefix and claims no unwritten session.

### Remaining gaps

1. `spawn.go:124-125` starts `pane.get` before a first start reference reaches the store.
2. `spawn.go:133-145` returns unresolved on pane session rejection before it records successful pane topology.
3. `spawn.go:157` treats the failed wait reply as a successful absent-session observation.

The timeout control supplies matching start, pane, and independent get references.
It supplies a stable foreground identity from one owned child.
The expected result is the existing live readiness-timeout partial result.
The actual result is unresolved with a false missing-session reason.

Recommendations: persist the first start reference before the pane read.
Keep known pane topology on the session partial path.
Observe session absence only after a successful typed reply.
Keep external waits outside the lock and shared helper compatibility unchanged.

## Runs

All independent edits ran in `/tmp/t10v03.tVxKrn2V/p`.
The exact pre-fix replay ran in `/tmp/t10v03.tVxKrn2V/red`.
Captures retain exact commands, output, exit statuses, source, tests, and mutants.

```sh
go test -count=1 ./internal/agent/ -run TestFix02Session -v
```

Exact pre-fix `15387805` production plus final new FIX02 tests exits 1.
Exactly the nine formerly open cases fail assertions. The controls pass.
No compile error substitutes for a red result.

```sh
go test -count=1 ./internal/agent/ -run 'TestFix02Session|TestV02' -v
```

Final source exits 0. Both retained session suites and the stop identity matrix pass.

```sh
go test -count=1 ./internal/agent/ -run TestV03 -v
```

Final source exits 1. Exact output includes:

```text
    v03_test.go:37: successful pane topology lost on session failure: terminal="" workspace="" tab="" cwd=""
    v03_test.go:55: before pane.get: durable reference = "", want earliest literal s
--- FAIL: TestV03SessionTimeoutControl (0.00s)
```

The full timeout assertion records the actual unresolved result and its false missing-session error.
Raw output: `session-final-literals.stdout` in the captures.

The first orchestration checklist, protocol checklist, V1, V2, V3, CLI, and logging suites pass.
A current mutation masks later absence and fails four assertions in each session suite.
Another mutation loses the earliest partial reference and fails three assertions in each suite.
Each exact restore returns the original source hash and green FIX02/v02 tests.
The independent literal failures remain open after restoration.
The accepted FIX01 replay and five prior mutations received an artifact audit only, not a fresh run.
The comment-only map received a direct check.

## Gate and native

One full gate ran in the actual source worktree at final Go:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run --allow-serial-runners ./...
go test -race -count=1 ./...
```

All four exit 0. The format list is empty. Lint reports `0 issues.`.
TestSTE passes. The Go source, dependencies, and baseline remain exact after the gate.
New prose requires zero hard STE findings with both glossaries and no baseline rise.

Binary SHA256: `a5d0a81119f1e0a5155fe3e2af1751b774ef7035ee39fbf679731dd6a98cf945`.
Build revision: `4227c958`, with `vcs.modified=false` and exact Go source `0eacb853`.
The installed catalog supplies `claude-haiku-5-5` for Claude Code `2.1.302`.
This child is a product smoke only, not source implementation or source verification by Claude.

The first owned spawn returns partial at its fixture trust dialog.
Only that dialog receives trust input. Hive stops that child before the fresh spawn.
Fresh pane `wB4:pF` in tab `wB4:tF` reaches ok/complete and native idle.
Send with wait succeeds. A separate pane read shows the distinct reply `● pong`.
The live list shows live. Stop succeeds. The all list shows the final record stopped.
Both owned pane reads parse as pane-not-found. Both owned tabs and child PIDs are absent.
The source pane retains its exact original shell, terminal, cwd, and screen.
Manual refusal uses focused production tests and retained original native evidence, not a claimed fresh native run.

## Scope, records, and limits

Report: `.hive/tmp/coordinator/reports/t10-verify-03.md`.
Captures: `.hive/tmp/coordinator/reports/t10-verify-03-captures/`.
The report gives both complete checklists, the completion audit, and the final cleanup receipt.
The old VERIFIED exact bytes remain in `VERIFIED-old.md` with SHA256
`9b21c4417946daeea76f2b1abd48f36f95babab5decddb42818b214c58b18900`.
The helper import remains exactly four approved paths from `12d25801837c3f0e39261574d504082e990281a5`.
Root activation receives an immutable historical audit only.
The transferred prior tree remains absent. Its archive, binary, inventory, and deletion receipt match.
Own compact evidence and matching no-follow inventories remain after owned temporary cleanup.
Prior reports, logs, state, credentials, worktrees, and other agents stay preserved.

The original red stub snapshot remains absent. Raw original red logs survive.
Exact original-stub replay remains BLOCKED. The exact FIX02 replay is a separate result.
The local-user/cooperating-writer boundary applies. Hostile replacement between Lstat and Open stays deferred for T01.
The trust action can leave a native trust entry for the removed fixture path.
Source acceptance and the broad final review remain uncompleted.
No Root landing, main change, worktree removal, CP2 PASS, or CP4 PASS follows.

END T10 VERIFY 03
