# Task 135: fourth independent verification — pass

Verifier: `t135-verifier-4`. Date: 2026-10-23.
Worker commit: `0026af33`. Latest fix: `6e2625a7`.
Current `origin/dev`: `3fbbb17f`. Original task base: `25f544e3`.
Approval: coordinator record 554, package `6c466dc6`, chats `c208` and `c209`.

**Verdict: pass. V1 to V5 pass. All three reported defects pass their checks.**
The manager fit rule governs the stopped captions and long-parent hint.
This verification changes no product source, design file, golden grid, fixture in `design/`, or worker report.
All paths below are relative to this evidence folder.
New run output, scripts, and captures are in `verifier4/`.

## Preserved failed reports

| Report | Original commit | Preserved copy |
| --- | --- | --- |
| First, multiline review | `008e9d2d` | `verifier4/failed-1-VERIFIED.md` |
| Second, compact title | `235892c0` | `verifier4/failed-2-VERIFIED.md` |
| Third, missing stopped state | `cef43222` | `verifier4/failed-3-VERIFIED.md` |

Each copy equals its original commit byte for byte.
The earlier verifier folders, worker reports, and `V<n>.md` files remain unchanged.
No existing repository test changed in rework 3.
`verifier4/preservation.txt` records these checks.

## Verification items

| Item | Result | Independent result |
| --- | --- | --- |
| V1 Tests red and green | pass | Base tests detect missing review. Each fix fails its regressions before the fix. Fifteen deliberate breaks cause behavior failures. Focused tests pass with `-count=50` and `-race`. |
| V2 Approved boards and grids | pass | All 50 exports equal committed grids through `internal/lib/cellgrid`. Nine installed grids equal the approval. All approval grids and fixtures remain unchanged. No mask. Board and golden tests pass. |
| V3 Live behavior and visual check | pass | All 16 initial captures equal their grids. Both sizes pass review interactions, multiline scrolling, and stopped-card checks through edits and resizing. |
| V4 End-to-end delivery | pass | Six standard cases and eight stopped-caption cases pass through executable-checked scratch stubs. Review sends nothing. Send delivers the exact final draft to the correct recipient once. |
| V5 Scope, gate, and cleanup | pass | The branch includes current `origin/dev`. Only named files change. All five gate commands pass. Owned scratch processes stop. Scratch paths are absent. |

## V1: red runs, deliberate breaks, and green runs

Exact commands, replacements, and source versions are in `verifier4/scripts/`.
All tests run from the worktree root or an isolated source copy.
Only the copies receive temporary source changes.

```sh
# Base 25f544e3, with task tests and required fixtures copied in:
go test ./internal/tui -count=1 -v -run 'TestAskReview|TestAskCompactReview|TestAskEachCompletionPathOpensTheReview|TestDemoFixtureShowsTheReview'
# Before the multiline fix:
go test ./internal/tui -count=1 -v -run '^TestAskReview(KeepsMultilineTextInsideTheDock|ReachesEachLineOfATallLabel)$'
# Before the compact-title fix:
go test ./internal/tui -count=1 -v -run '^TestAskCompactReviewForAStoppedAskerKeepsTheTitle$'
# Before rework 3:
go test ./internal/tui -count=1 -v -run '^TestAskStopped(Review|Editor)KeepsTheStateInTheCard$'
# Final implementation, from the worktree root:
go test ./internal/tui -count=50 -run 'TestAsk|TestHelp|TestDemoFixtureShowsTheReview|TestEachBoardFixtureLoadsAtItsSpecSize'
go test ./internal/tui -race -count=1 -v -run 'TestAsk|TestHelp|TestDemoFixtureShowsTheReview|TestEachBoardFixtureLoadsAtItsSpecSize'
```

The base run has 17 failing tests and two passing retained guards.
The guards retain the empty-Other check and blocked-caller check.
The focused race run has 42 passing top-level tests.
Logs: `red-base.txt`, `multiline-red.txt`, `compact-title-red.txt`, `stopped-state-red.txt`, `green-count50.txt`, and `green-race.txt` in `verifier4/`.

The pre-rework-3 run reproduces these defects:

- At 80 columns, the card has no stopped-state text when neither route caption fits.
- With both names at 32 characters, the wide caption covers the border gap after the title.
- With a 32-character parent, the compact bottom border loses its hint.

`verifier4/mutations.txt` records nine detected breaks:
empty Other, blocked caller, review digits, switched draft, refreshed draft, return from edit, cut rows, frame keys, and tall-row paging.
Six further runs detect removal of the state-only caption, title gap, full caption, short caption, shorter hint, and line-feed handling.
Their logs are `verifier4/mutation-*.txt`.
Each red run fails for behavior, without a build failure.
The scripts restore the source after each break.

### Original independent regressions

The original multiline and missing-state regressions fail before their fixes.
They pass with the final implementation, including under `-race`.
The multiline field test retains literal line feeds in Other text and `note` values.
Its expected reply remains:

```json
[{"picks":[1,3],"other":"other one\nother two","note":"note one\nnote two"}]
```

Logs: `original-multiline-red.txt`, `original-stopped-red.txt`, and `all-regressions-green.txt` in `verifier4/`.

The original `TestVerifierCompactStoppedReviewTitle` runs unchanged before and after the title fix.
Before the fix, it fails because the title loses `rm`.
With the final implementation, it finds the full title but still fails its extra full-caption literal at 80 columns.
The actual line is:

```text
┏━ ? asker-scratch asks · confirm ━━━━━━━━━ asker stopped · to parent-scratch ━┓
```

`verifier4/original-full-caption-result.txt` preserves that failure.
The manager explicitly permits this short caption when the full caption cannot fit.
This old scratch expectation does not govern the final fit rule.
I did not change it or report it as an unchanged passing test.
The repository regression checks the permitted caption and passes.

### Changed existing expectations

`verifier4/changed-expectations.diff.txt` preserves the exact task diff.
I removed, skipped, or weakened no existing test.

| Existing test | Approved cause |
| --- | --- |
| `TestAskBoardsMatch` | Three approved review boards. |
| `TestAskCardHasTheFocusInNormalMode` | Review hints replace immediate-send hints. |
| `TestAskTwoQuestionsWithMarksAndANote` | Review and `note` hints, then one additional Send press. |
| `TestAskDigitPicksAndSendsASingleQuestion` | One additional Send press. |
| `TestAskArrowsAndEnterPickTheCurrentOption` | One additional Send press. |
| `TestAskEnterFromANoteBeforeAPickPicksTheCurrentOption` | One additional Send press. |
| `TestAskNoteTabDropsAndCtrlBracketKeeps` | One additional Send press. |
| `TestAskOtherOpensATextLine` | Keep-answer hint and one additional Send press. |
| `TestAskStoppedAskerShowsTheParent` | Review-for-parent hints. The route check stays. |
| `TestAskMouseFocusesAndScrolls` | A click opens review before Send. |
| `TestAskOtherInAMultiSelectQuestion` | One additional Send press. |
| `TestDemoBlocksTheAnswer` | One additional Send press before the demo effect. |
| `TestHelpListsTheAskKeys` | Approved review Help line. |
| `TestEachBoardFixtureLoadsAtItsSpecSize` | Three new boards increase the count to 50. |
| `TestCheckAcceptsTheAskBoards` | Three approved board registrations. |

Decisions 113 to 119 and package `6c466dc6` authorize these changes.
Rework 3 adds regressions without changing an existing expectation.

## V2: boards and approval

```sh
mkdir -p .hive/tmp/verifier135d/exports
go run ./internal/boardexport/cmd/boardexport -out .hive/tmp/verifier135d/exports .hive/spec/tui-v2/design
go test ./internal/boardexport/... ./internal/lib/cellgrid/... ./internal/tui/... -count=1 -v
```

The exporter exits 0 and prints nothing.
`verifier4/grid-compare.txt` records 59 grid comparisons and three installed fixture comparisons.
All grid comparisons check rune, colour token, and weight through `internal/lib/cellgrid`, with no mask.
`verifier4/all-approval-files.txt` also checks every approval grid, every approval fixture, and all 50 installed fixtures against `6c466dc6`.
All 78 file checks pass.
`verifier4/boards.txt` records the passing board and golden checks.
`verifier4/approval-chat.json` preserves the original plan and user approval.

Changed existing boards: `ChatAsk`, `ChatAskMulti`, `ChatAskPreview`, `ChatAskStopped`, `CompactChatAsk`, and `Help`.
New boards: `ChatAskReview`, `ChatAskReviewMulti`, and `CompactChatAskReview`.
`verifier4/board-scope.txt` lists all other 41 board sets as unchanged from current `origin/dev`.
All design files remain unchanged since `cef43222`.

## V3: live interactions and prior failures

I built the task binary and used private herdr session `verify135d` under `/tmp/hv135-verify4`.
The server reports herdr 0.9.5 and protocol 22.
Each TUI uses a scratch home, `HIVE_NO_STARTUP=1`, `--no-startup`, and `COLORTERM=truecolor`.
Demo runs use the exact board size in a 140x48 pane.
Normal runs use a private tmux window at 80x24 or 120x40.
No test TUI or test prompt uses the real home.

All 16 initial captures equal their grids: `verifier4/live-grid-compare.txt`.
Each capture is stable across two reads 500 ms apart, with no truncation or scroll offset.
The live comparison follows the spec exception for invisible attributes of space cells.
All visible cells match in rune, colour token, and weight. No mask applies.

I captured all nine changed boards again with fresh Firefox profiles.
I compared each board PNG with its terminal PNG.
`verifier4/visual-pairs-{1,2,3}.png` shows these pairs.
The evidence contains 91 PNG files and 229 ANSI captures.

At both sizes, the live checks cover:

- All five completion paths, with review before Send.
- Question edits and return to review with Send selected.
- Ignored digits, clicks, space, tab, text, and other frame keys.
- Card switches, dismissal, and an ordinary watched-record refresh.
- Multi-picks, full Other text, long `note` values, wrapping, and tall-row paging.

Each of the six multiline scenarios has 11 review captures.
Every physical body row retains both borders and stays within the dock cap.
Arrow navigation reaches `ROW01` through `ROW30` and `FINAL-LABEL-SENTINEL`.
Multiline Other and `note` fields remain readable.
`verifier4/multiline-check.txt` records the independent physical-row checks.
The two pre-existing editor/feed defects remain outside this task and unchanged.

### Stopped-card fit matrix

The accepted request comes from `verifier3/no-caption/request.json`.
Both long names have 32 ASCII characters and meet `namePattern` in `internal/agent/register.go`.
Each check reads only the card region. Header and feed words cannot satisfy it.

| Asker / parent | 80x24 caption | 120x40 caption |
| --- | --- | --- |
| `t1` / `p1` | `asker stopped · answer goes to p1` | Same full caption. |
| `asker-scratch` / `parent-scratch` | `asker stopped · to parent-scratch` | `asker stopped · answer goes to parent-scratch` |
| `asker-scratch-with-long-name-abc` / `parent-scratch` | `asker stopped` | `asker stopped · answer goes to parent-scratch` |
| `asker-scratch-with-long-name-abc` / `parent-scratch-with-long-name-ab` | `asker stopped` | `asker stopped · to parent-scratch-with-long-name-ab` |

Each case starts at both sizes and checks:

1. The full title and the ` ━` gap after it.
2. An explicit stopped word inside the card.
3. The full parent destination in Send and its hint.
4. A question edit from Yes to No, then return to selected Send.
5. A resize to the other size and back.
6. No answer before Send, then the exact parent delivery.

With the long parent at 80 columns, the hint uses `↑↓ change`.
Every other matrix case keeps `↑↓ question to change`.
The destination stays complete in both forms.
Thirty valid later coordinator messages hide the ask event in each case.
A separate visible/hidden control keeps the card byte-equal while the event leaves the viewport.
See `verifier4/stopped/feed-control.txt` and its two captures.
All eight cases pass: `verifier4/stopped-live.txt`.

### Visual table

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header and tabs | status words/glyphs, `? N`, tab names and underline | matches | Approved initial captures equal their grids. |
| Feed | names, times, `? asked`, open/waits words | matches | Approved feed layouts remain unchanged. |
| Input | `message`, status word, typing instruction | matches | Existing input layout remains unchanged. |
| Dock frame and count | heavy border, `open`, `I of N` | matches | Same card and dock cap. |
| Option editor | numbers, `[x]`, `[ ]`, `›` | matches | Edits retain other draft values. |
| Preview | `preview · N`, framed text | matches | Capture equals the approved Preview grid. |
| `note` editor | `note ›`, cursor `█`, `enter keep note` | matches | Approved hint change. |
| Review title | full `confirm` word | matches | Full at both sizes, including both long names. |
| Review answers | headers, `→`, labels, Other text and `note` values | matches | Full text wraps and remains reachable. |
| Selected question | `›`, `enter change` | matches | Arrows select or page the question. |
| Send row | `› Send`, or `› Send to PARENT` | matches | Selected after entry and after edits. |
| Full stopped caption | `asker stopped · answer goes to PARENT` | matches | Retained wherever it fits with the required gap. |
| Short stopped caption | `asker stopped · to PARENT` | overridden | Manager fit rule and decision 114. |
| State-only stopped caption | `asker stopped` | overridden | Manager fit rule when neither route caption fits. |
| Title gap | ` ━` after the full title | matches | Neither stopped caption covers it. |
| Long-parent hint | full parent and `↑↓ change` | overridden | Manager fit rule when the full hint cannot fit. |
| Other review hints | Send/change action, arrows, `x drop` | matches | Approved review keys. |
| Footer | `[ask]`, `[note]`, `[other]`, `[review]` | matches | Each focus state has a word. |
| Help | `ask review:` and its keys | matches | Capture equals the approved Help grid. |
| Tall and multiline review | `›`, scroll mark, line text and sentinels | matches | Every segment is reachable inside intact borders. |

Plain-text captures retain the state words and glyphs.
No tested state requires colour as its only signal.

## V4: real TUI delivery through checked scratch stubs

Before any test prompt, recipient panes show the compiled executable path in `STUB_READY`.
Before each delivery setup, `/proc/PID/exe` equals `/tmp/hv135-verify4/bin/pi`.
The stopped cases also check the parent executable immediately before Send.
`verifier4/stub-executables.txt` records the paths, PIDs, and binary hash.
No installed Pi executable runs for these checks.

The scratch agent creates each ask from its own pane through the actual `hive ask` command.
Review leaves the record open, without replies or attempts.
Neither recipient receives a prompt before Send, including after edits, refresh, and resizing.

| Size | Edited answer | Standard stopped asker | Dismiss | Stopped fit matrix |
| --- | --- | --- | --- | --- |
| 120x40 | pass | pass | pass | Four cases pass. |
| 80x24 | pass | pass | pass | Four cases pass. |

Literal final replies in the standard edit sequence:

```json
[{"picks":[2,3]},{"picks":[3],"other":"Next release","note":"keep this"}]
```

Literal expected prompt for `a1`:

```text
⬢ user: answer to ask a1 from asker-scratch · Scope → Help line, Header count; Order → Next release · note: keep this
```

Stopped cases reach only the live parent and retain `(asker stopped)`.
Each stopped matrix case delivers `Change → No` after its edit and resizes.
Dismiss reaches only the asker and retains `No answer comes.`
All 14 completed records have `attempts: 1` and the expected `delivered_to`.
Before/after records, literal prompts, and recipient output remain in `verifier4/delivery/` and `verifier4/stopped/`.
Exact commands and socket requests are in `verifier4/live-commands.txt`.

## V5: scope, gate, and cleanup

The merge of current `origin/dev` reports `Already up to date.`
`origin/dev` at `3fbbb17f` is an ancestor of the checked worker commit.
`verifier4/merge.txt` records the check before the gate.
`verifier4/diff-stat.txt` records `git diff dev...HEAD --stat`.
`verifier4/source-paths.txt` lists changed files outside this task folder.
All product changes are within the brief's named files or tests beside them.
This verification adds only this report and `verifier4/` evidence.

`verifier4/gate.txt` preserves the full root gate output:

| Command | Result |
| --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | No output, exit 0. |
| `go mod tidy -diff` | No output, exit 0. |
| `go vet ./...` | No output, exit 0. |
| `golangci-lint run ./...` | `0 issues.`, exit 0. |
| `go test -race ./...` | All packages pass, exit 0. |

One scratch mutation used an obsolete source pattern on its first setup run.
I corrected the pattern before the counted run.
`verifier4/regression-setup-error.txt` preserves the setup error separately.
`verifier4/regression-runs.txt` records the complete successful red/green sequence.

Cleanup is complete before this verdict commit.
`verifier4/cleanup.txt` records the checks:

- The private herdr server, its panes, and all six compiled recipient stubs are absent.
- The private tmux server and socket are absent.
- `/tmp/hv135-verify4` and `.hive/tmp/verifier135d` are absent.
- I removed scratch homes, binaries, source copies, exports, HTML, and Firefox profiles.
- The worktree `.hive/tmp/` folder is empty.
- Final evidence, exact scripts, accepted inputs, and captures remain in this task folder.

Cleanup uses literal guarded paths and checks process identity before any signal.
No real home pane, session, process, or data changes during the test runs.
