# T08-tui-log-view ESCAPE-VERIFIED

| Item | Value |
| --- | --- |
| Task | T08-tui-log-view |
| Pass | `t08-escape-verify-01` |
| Commit | `25f20261cc25d2f827146b2001643885435137e5` |
| Tree | `d45e6eb56f61dd8fcadebe5da17b8b80991a9d82` |
| Pre-fix source | `b4ab63ca1943c2c0491ae193392d8f7bdf3ea857` |
| Branch | `task/T08-escape-clear` |
| Date | 2026-10-18 |
| Agent | verifier, Claude Code, `claude-opus-5-5` |
| Verdict | **PASS** |

| Item | Result | Remark |
| --- | --- | --- |
| R1 viewing Escape | PASS | The text filter and the input clear. Mode, level, and follow stay. Hidden lines appear. |
| R2 editing keys and quit | PASS | Editing Escape logs `edit cancel`. Editing `g` and `q` are text. `q` and Ctrl-C quit. |
| R3 accepted checklist | PASS | The focused T08 tests and the full race suite pass. |
| R4 checked log path | PASS | Escape logs `filter clear`. A refused write shows the first bounded error. |
| Red and green | PASS | The final test fails on the pre-fix source and passes on the final source. |
| Mutation | PASS | The behavior-only mutation fails an assertion. Exact restoration passes. |
| Native TUI | PASS | The literal key sequence clears the filter. The child exit status is 0. |
| Gate | PASS | Format list empty, vet 0, lint `0 issues.`, race suite 0, TestSTE 0. |

This verdict covers the narrow viewing-Escape fix only. It gives no Root landing and no CP4 pass.
The actual-home CP4 replay needs a verified Root landing first.
The own report is `.hive/tmp/coordinator/reports/t08-escape-verify-01.md` in this worktree.
The raw captures are in the adjacent `t08-escape-verify-01-captures/` directory.

## Entry and scope

The local branch, the tracking reference, and the direct remote branch were each `25f20261` at entry.
The worktree status was clean. The branch has no upstream setting. The three references agree.

The fix commit changes exactly three paths against `b4ab63ca`:

| Path | Blob |
| --- | --- |
| `internal/tui/logview.go` | `02233d57ccd2a362ced6c785493112a8d438b3e2` |
| `internal/tui/escape_clear_test.go` | `93d8d8c1fa0ce34e3d9f622a788c77fa4e7e9b89` |
| `spec/hive-v1/evidence/T08-tui-log-view/ESCAPE-FOLLOWUP.md` | new evidence |

The viewing switch gets one `esc` case. It calls one new helper and then `logAction("filter clear")`.
The editing Escape case calls the same helper and keeps `logAction("edit cancel")`.
The helper body has the same five statements as the old editing case.
`go.mod`, `go.sum`, `ste_test.go`, and `testdata/ste-baseline.json` do not change.

## Red and green

All runs used an exact owned copy of the commit. The copy had 317 files with equal hashes.

```sh
go test -count=1 ./internal/tui/ -run ^TestViewingEscapeClearsAppliedTextFilter$ -v
```

| Run | `logview.go` SHA256 | Result |
| --- | --- | --- |
| Red, pre-fix | `f03da3d2…c493` | `escape_clear_test.go:48: log lines went from 11 to 11` and exit 1 |
| Green, final | `6c457104…1362` | PASS and exit 0 |

The test file SHA256 is `c6eeb82a…0d57` in each run. The red run compiles.
It fails on an assertion because the old viewing Escape does nothing.

## Behavior-only mutation

Mutation M1 removes only `m.clearTextFilter()` from the viewing case. The `filter clear` log call stays.

```text
escape_clear_test.go:63: view does not contain "level: info  text: -  follow: on"
level: info  text: project  follow: on
exit 1
```

Restoration gave SHA256 `6c457104…1362`. That file equals the worktree file. The test then passed with exit 0.

## Independent probes

The probe file has four tests with literal expected strings only. It ran with `-race` in the owned copy.

| Probe | Final source | Pre-fix source |
| --- | --- | --- |
| PV1 CP4 sequence, footer, lines, one `filter clear` event | PASS | FAIL at the Escape footer |
| PV2 warn level and follow off stay after Escape | PASS | FAIL at the Escape footer |
| PV3 editing Escape, editing `g` and `q`, quit keys | PASS | PASS, control |
| PV4 refused daily link on Escape | PASS | FAIL, no `error:` view |

PV4 opens a new daily file through a link to an outside sentinel. The view shows `not a regular file`.
The sentinel bytes and its directory stay exact. A later `l` keeps the first error. `q` quits.
Two early probe revisions had probe defects, not source defects.
The first did not remove style codes from the input line. The second used an event text that contains `project`.
Both defect runs stay in the captures.

## Accepted checklist

The focused selection of the accepted T08 tests passed with `-race` on the exact copy.
It covers keys, follow, daily files, finite EOF and descriptor input, stream errors, log refusal, and quit keys.
Each of the 23 selected tests gave PASS. The full race suite below also passed.

## Native TUI

The binary is `/tmp/t08ev01.p55zdTFf/hive-t08ev01`, built from the clean worktree.

| Fact | Value |
| --- | --- |
| SHA256 | `c8b3f9f4bc2753204f99e1ad1c7ca232ab2beff658b1150c8f4317706f51a914` |
| Size | 10018666 bytes |
| Build | `vcs.revision=25f20261…`, `vcs.modified=false` |
| Source map | 106 package files. Each hash equals the worktree file. |

The run used a new marked fixture home and a new own tab `wBA:t3`, pane `wBA:p3`.
The raw keys were `/`, `project`, Enter, `l`, `f`, `g`, `G`, Escape, `q`.

| Key | Footer |
| --- | --- |
| `/` | `filter:` |
| `project` | `filter: project` |
| Enter | `level: debug  text: project  follow: on` |
| `l` | `level: info  text: project  follow: on` |
| `f` | `level: info  text: project  follow: off` |
| `g` | `level: info  text: project  follow: off` |
| `G` | `level: info  text: project  follow: on` |
| Escape | `level: info  text: -  follow: on` |

After Escape the view shows `nonmatching zulu` and `outside warn yankee` again.
The `info` level still hides `project debug fx`.
The log grew from 13 to 14 lines at Escape. The new line has action `filter clear`.
The wrapper wrote child exit status `0`. The pane then held its shell.
I closed only that tab after an identity check. Parsed lists show `wBA:t3` and `wBA:p3` absent.
Pane `wBA:p1` holds the same shell process as at entry.

## Gate

The full gate ran once in the actual worktree at `25f20261` with `TMPDIR=/tmp/t08ev01.p55zdTFf/tmp`.

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run --allow-serial-runners ./...
go test -race -count=1 ./...
go test -run TestSTE -count=1 -v .
```

The format list was empty. Each command gave exit 0. Lint reported `0 issues.`. Each package passed.

## Worker evidence audit

The worker red, green, and mutation hashes equal the commit bytes.
The worker red run fails at line 48. Its first test revision fails at line 43 on `text: project`.
The worker mutation removed the full viewing case. Its patch is not kept, only its result hash.
The worker binary used a dirty tree at `b4ab63ca`. Its file map is not kept.
This pass closes both gaps with a new mutation patch and a clean binary at `25f20261`.
The worker fixture run shows the same cleared footer and one `filter clear` event.

## Limits

Only the local-user and cooperating-writer boundary applies.
Hostile file replacement between Lstat and Open stays deferred for T01.
The CP4 actual-home replay is not part of this pass. CP4 stays open.
