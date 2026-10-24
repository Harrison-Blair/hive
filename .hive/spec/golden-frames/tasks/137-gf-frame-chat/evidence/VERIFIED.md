# Task 137 verification

- Task: `137-gf-frame-chat`
- Verifier: `t137-verifier-2`
- Date: 2026-10-24
- Code checked: `a4d99be810f1cc543b708d39e792658485e93f74`
- Verdict: **pass**

Both findings of the first verification are fixed.
The verifier ran each verification item again.
The fresh evidence is in `verifier-2/`.
The first failed report remains in Git at `c4badeaa`.

## Verification items

| Item | Result | Command and observation | Evidence |
| --- | --- | --- | --- |
| V1 Board copies and export | pass | `sha256sum -c -` accepts all 16 initial copies at `6ce218d0`. The current unchanged `canvas.json` also passes. The exporter reproduces all 15 current grids. `git status --short .hive/spec/golden-frames/design/` prints nothing. AC10 accounts for the changed boards. | `verifier-2/board-copy-sha.txt`, `verifier-2/export.txt` |
| V2 Behavior, red and green | pass | `go test ./internal/... -count=1 -v -run NAMES` passes for 385 test names from changed test files. Thirteen deliberate breaks fail their named tests, then pass without the break. Both fixed regressions fail on pre-fix production files and pass on current files. Five additional verifier checks pass. | `verifier-2/behavior-all-green.txt`, `verifier-2/mutations/`, `verifier-2/repair-*-red.txt`, `verifier-2/repair-*-green.txt` |
| V3 Golden checks | pass | The golden commands below pass for all 14 screen boards and the in-scope component regions, without masks. The retirement test accounts for 48 stopped grids and 2 kept grids. The property tests pass at all four sizes. | `verifier-2/golden.txt`, `verifier-2/breakpoints.txt` |
| V4 Live check | pass | The task binary ran all 14 fixtures in a 120×40 herdr pane with `COLORTERM=truecolor` and `--size 120x40`. Each capture has zero differing cells. The required key checks pass. Fresh board and screen PNGs agree. Plain-text state signals remain without colour. The scratch session is deleted. | `verifier-2/live/`, `verifier-2/live-compare.txt`, `verifier-2/keys-check.txt`, `verifier-2/visual.md`, `verifier-2/png/` |
| V6 Board changes | pass | `git diff --no-index` compares each current board with its pinned source. The 15 changed-row lists and old/new hashes match the recorded cases. `go test ./ -run STE -count=1` passes. | `verifier-2/board-diffs/`, `verifier-2/board-changes-check.txt`, `verifier-2/ste.txt` |
| V5 Gate | pass | All five gate commands exit 0. `gofmt -l` and `go mod tidy -diff` print nothing. The linter reports `0 issues.` The race tests pass. | `verifier-2/gate.txt` |

The V3 commands were:

```sh
go test ./internal/boardexport/ ./internal/tui/ ./internal/tui/component/ -count=1 -v -run 'Golden|Gf|Stopped'
go test ./internal/tui/ -count=1 -v -run 'Retired|EachScreenKeepsTheFrame'
```

## Earlier findings

### F1: the ask border after a resize — fixed

The valid long ask from the first verification now keeps its top-right `╮` corner.
The top row has exactly 80 cells after the resize from 120×40 to 80×24.
Both the worker regression test and the independent verifier check pass.
An overlay of the pre-fix `composer.go` makes both tests fail again for the missing corner.
`verifier-2/probe-captures/` shows the current model before and after the resize.
No board or golden grid changed for this fix.

### F2: the valid Bubble Tea key `alt+up` — fixed

The real config loader now accepts `alt+up`.
The worker test also checks seven other named or modified keys and four invalid keys.
The independent model check loads `help = "alt+up"`, sends that Bubble Tea key, and observes Help.
`esc` returns to Chat.
An overlay of pre-fix `config.go` makes the config and model checks fail again.

## Deliberate rule breaks

Each break uses a Go overlay. No production file changes in the worktree.
Each red run exits 1 for the named assertion. Each green run exits 0.
The first 13 rows have patches and outputs in `verifier-2/mutations/`.
The final two rows restore production files from `c4badeaa`.

| Break | AC | Rule | Test |
| --- | --- | --- | --- |
| Error token hex | AC2 | `error` is `#e8705f` | `TestConvertErrorNoticePrefix` |
| Agents chord routes to Chat | AC3, AC6 | Each chord opens its screen | `TestFrameChordsOpenScreens` |
| Heavy title rule | AC3 | The title row has the approved rule | `TestTitleRowGolden` |
| Extra rows below 30 rows | AC4 | Only the first status row shows | `TestVerifierStatusRowsHeight` |
| Join with ` / ` | AC4 | Segments join with ` · ` | `TestStatusRowGolden` |
| No command throttle | AC4 | At most one start per 300 ms | `TestStatusCommandRunsAtMostOnceEach300ms` |
| Run the command in demo | AC4 | Demo runs no status process | `TestStatusCommandRunsOnlyWhenItApplies` |
| Omit the startup notice | AC5 | Startup notice precedes an error | `TestFooterShowsStartupNoticeBeforeError` |
| Ignore the box chord | AC6 | The box chord moves focus | `TestFrameBoxFocusKeys` |
| Wrong blocked placeholder | AC7 | Chat rule 10 | `TestVerifierChatStatePlaceholders` |
| Send with no coordinator | AC7 | Chat rule 12 keeps the text | `TestChatSendWithNoCoordinatorKeepsTheText` |
| Earlier count at column 50 | AC7 | Earlier count starts at column 52 | `TestChatTallTranscriptFollowsItsEnd` |
| Omit the blank message row | AC7 | Chat rule 2 | `TestChatBlankRows` |
| Restore the old ask title truncation | AC3 | A resize keeps each border | `TestAskResizeKeepsTheTopRightCorner`, verifier check |
| Restore the old key validator | AC6 | Bubble Tea named keys are valid | `TestParseConfigTUIAcceptsNamedKeys`, verifier config and model checks |

The whole pre-feature production overlay at `55df4ace` fails to compile against the current tests.
Its missing APIs include `Error`, `Composer`, and `Settings.TUI`.
That run does not prove each behavior.
The deliberate breaks above supply behavioral red runs for the named rules.
This report does not claim a separate semantic red run for every assertion in all 385 tests.

## Limits and cleanup

This verdict covers the task scope. It excludes the wave C bodies and component regions.
The user must review the changed grids before landing, as AC10 requires.
`board-review.html` remains the side-by-side review file.
This verification does not record user approval of that review.

The verifier removed its binary, overlays, exports, browser profiles, and scratch home after recording the evidence.
The verifier stopped and deleted `t137-check-verifier-2`.
`verifier-2/cleanup.txt` records the checks.
No production file, worker `V<n>.md` file, or plan tracker changed during verification.
