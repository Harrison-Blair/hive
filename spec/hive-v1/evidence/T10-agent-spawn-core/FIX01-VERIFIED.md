# T10-agent-spawn-core FIX01 verification

| Item | Value |
| --- | --- |
| Pass | `t10-verify-02` |
| Date | 2026-10-18 |
| Entry | `892d643889d303db064afc21e01df2a018805589` |
| Final Go source | `d55e9de6ec441313884f29e80188f5fd96cb22a4` |
| Agent | verifier, Pi, `openai-codex/gpt-6.1-sol:high` |
| Verdict | **FAIL** |

`T10-O-F02` remains OPEN. The other five findings are CLOSED.
The inherited tests, full gate, and fresh native product smoke pass.
They do not replace the failing independent session assertions.
No production source or old test changed in this pass.

## Finding results

| Finding | Result | Evidence |
| --- | --- | --- |
| T10-O-F01 | CLOSED | Production Send, real store, six inspection cases, no close |
| T10-O-F02 | OPEN | Missing later sessions and lost earliest session on partial paths |
| T10-O-F03 | CLOSED | Final held-handle tuple, changed-current-record preservation, cosmetic control |
| T10-O-F04 | CLOSED | Direct old-map preservation and new-file map check |
| T10-P-F01 | CLOSED | Typed wrappers, valid controls, actual malformed-success CLI refusal |
| T10-P-F02 | CLOSED | Literal timeout boundaries, zero invalid prompt requests, immediate peer |

The F01 cases include the restoration placeholder and a proved same-terminal shell.
The F03 matrix changes each critical identity field during an external call.
Every critical change prevents close. Cosmetic changes alone permit close and retain unrelated values.
External reads remain outside the lock. Probes use only owned child processes.
The wire checks keep the correct native `agent` properties.
The actual malformed-success CLI exits 1 with `partial/prompt`, not `ok/complete`.

## Open first-session behavior

Production spawn accepts each of these sequences as `ok/live`:
- Start session present, pane session absent, later sessions matching.
- Start and pane sessions matching, wait session absent, get session matching.
- Start, pane, and wait sessions matching, get session absent.
- First session at wait, get session absent.

Each requires `unresolved` after the later absent observation.
The current merge skips absent sources and permits another source to mask the absence.

Several partial paths also lose the earliest reference:
- Start session differs from the pane session.
- First session at wait differs from the get session.
- The pane read fails after a present start session.
- The process read fails after a first session at wait or get.

The resulting record is unresolved but has an empty session reference.
The required record retains the earliest present session.
Known started-agent effects remain truthful in the probes.
Controls pass for no session and a first session at each of the four observation points.

Recommendation: retain the first present session before the next external read.
Check each later observation in order, including absence after that first session.
Keep the earliest reference and known effects on partial outcomes.
A source fix and fresh independent verification remain uncompleted.

## Runs

All independent Go edits occurred in `/tmp/t10v02.GitlBsVc/p`.
The real worktree source stayed exact.
The runner retains exact commands, stdout, stderr, durations, and JSON exit statuses.

```sh
go test -count=1 ./internal/agent/ -run TestFix -v
go test -count=1 ./internal/lib/herdr/ -run TestFix -v
```

These pre-fix replay runs exit 1 on assertions for all five behavior findings.
The production bytes equal retained entry `440049ae`.
The new final tests and the two authorized external fixture corrections remain exact.
The final inherited tests pass.

```sh
go test -race -count=1 ./internal/agent/ -run TestSpawn -v
go test -count=1 ./internal/agent/ -run 'TestSend|TestStop|TestFix' -v
go test -count=1 ./internal/agent -run TestV02 -v
```

The first two commands exit 0.
The independent matrix exits 1 with nine session assertion failures.
The same matrix passes all 15 critical stop changes and its cosmetic control.
The raw independent output includes:

```text
    v02_test.go:43: outcome=ok status=live session={"source":"herdr:claude","agent":"claude","kind":"id","value":"s"} phase=complete
    v02_test.go:44: status=live want unresolved
```

Five deliberate current behavior breaks each fail an assertion.
Exact restoration of each source file gives a green affected regression test and the original SHA256.
The comment-only map uses a direct check, not an artificial mutation.
The broader failing session matrix remains part of this verdict.

## Gate and native flow

One full gate ran in the actual source worktree at final Go:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run --allow-serial-runners ./...
go test -race -count=1 ./...
```

Every command exits 0. The format list is empty. Lint reports `0 issues.`.
The root STE tests pass. New prose has zero hard findings with both glossaries.
The baseline, dependencies, and tested Go source stay exact.

Binary SHA256: `e0b0fa78603349819dc7c7a7cd25b7fbc4159e31812e7fd4ba1ce5c9d7ea6494`.
Build revision: `892d643889d303db064afc21e01df2a018805589`, with `vcs.modified=false`.
Its Go source equals `d55e9de6`.
The installed executable catalog supplies `claude-haiku-5-5`.
Claude Code `2.1.302` shows `Haiku 5.5` in the owned pane.
This is a product smoke, not source implementation or source verification by Claude.

The first owned spawn reaches its fixture trust dialog and returns partial.
Only that dialog receives trust input. Hive stops that child before the fresh spawn.
The fresh pane `wB4:pB` in tab `wB4:tB` reaches `ok/complete` and native `idle`.
Send with `--wait --timeout 120` exits 0 with settled status `done`.
The separate pane read shows the prompt and the distinct reply `● pong`.
The live list shows live. Stop exits 0 with pane closed and workspace surviving.
The all list shows the final record stopped.
The later pane get parses as `pane_not_found`. Both owned tabs are absent.
Both native child PIDs are absent. The source pane and its original shell survive.
The trust action can leave a native trust entry for the removed fixture path.
Manual refusal uses focused production tests and retained original native evidence, not a claimed fresh native run.

## Scope, preservation, and limits

Report: `.hive/tmp/coordinator/reports/t10-verify-02.md`.
Captures: `.hive/tmp/coordinator/reports/t10-verify-02-captures/`.
The report includes both complete first-pass checklists and the completion attestation audit.
The helper import remains exactly four approved paths from `12d25801837c3f0e39261574d504082e990281a5`.
Root activation passes an immutable historical audit. No Root write or activation rerun occurs.

The transferred prior temporary root is absent.
Its compact retained archive, binary, inventory, and deletion receipt match.
The current pass archives and checks all 726 owned regular files before deletion.
Two exact no-follow inventories agree for all 905 owned entries.
Only that matching owned tree was removed. Its path is absent. `TMPDIR` is unset.
The captures retain projects, probes, original source, mutants, binary, logs, raw output, and manifests.
Old work, reports, logs, credentials, worktrees, and other agents remain preserved.

The old VERIFIED exact bytes remain in `VERIFIED-old.md` in the captures.
Its SHA256 is `fe4c4cea543bad41def8483502e2e3de1a921545177875a06a46fad59416952b`.
The original red stub snapshot is absent. Raw original red logs survive.
Exact original-stub replay stays BLOCKED. The exact fix replay is a separate result.
The approved boundary remains the local user and a cooperating writer.
Hostile replacement between `Lstat` and `Open` stays deferred for T01.
No Root landing, main change, worktree removal, CP2 PASS, or CP4 PASS follows from this verdict.

END T10 VERIFY 02
