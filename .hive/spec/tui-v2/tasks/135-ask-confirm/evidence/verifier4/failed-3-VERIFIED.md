# Task 135: third independent verification — fail

Verifier: `t135-verifier-3`. Date: 2026-10-23.
Worker commit: `267eef99`. Title fix: `0ae08cf9`. Multiline fix: `cc4c170a`.
Current `origin/dev`: `3fbbb17f`. Original task base: `25f544e3`.
Approval: coordinator record 554, package `6c466dc6`, user chats `c208` and `c209`.

**Verdict: fail. At 80x24, a stopped review can lose its explicit stopped-state signal when neither caption fits.**
The standard compact title fix passes under the manager fit rule.
The multiline fix, approved grids, delivery checks, and full gate pass.
The manager accepted the new finding as V3 fail in the delivered instruction after inbox record 571.

All paths below are relative to this evidence folder.
This verification changes no product source, design source, golden grid, design fixture, or worker report.

## Preserved failed verdicts

- `verifier3/first-failed-VERIFIED.md` equals the report at `008e9d2d`, byte for byte.
- `verifier3/second-failed-VERIFIED.md` equals the report at `235892c0`, byte for byte.
- The earlier `verifier/`, `verifier2/`, `rework/`, `rework2/`, and `V<n>.md` files remain unchanged.

## Verification items

| Item | Result | Independent result |
| --- | --- | --- |
| V1 Tests red and green | pass | All 19 task tests run on the base. Seventeen fail for behavior. Two retained guards pass. Nine deliberate breaks detect failures. Both fixes fail their regressions before the fix. Focused tests pass with `-count=50` and `-race`. |
| V2 Approved boards and grids | pass | All 50 exports equal their committed grids through `internal/lib/cellgrid`. All 9 approval grids and 3 new fixtures equal package `6c466dc6`. No mask. Board and golden checks pass. |
| V3 Live behavior and visual check | fail | All 16 initial captures equal their grids. Required interactions and multiline checks pass. A valid long asker name removes the stopped caption at 80x24. The card then has no explicit stopped-state signal. |
| V4 End-to-end delivery | pass | Six standard cases and two long-name cases pass through executable-checked scratch stubs. Review sends nothing. Records and delivered prompts keep the final answers and unchanged routing. |
| V5 Scope, gate, and cleanup | pass | The branch includes current `origin/dev`. Only named files change. All five gate commands pass. Cleanup checks pass. |

## Blocking finding: neither caption fits

### Accepted input and reproduction

1. Use the normal scratch TUI at 80x24.
2. Create an ask from `asker-scratch-with-long-name-abc`, with parent `parent-scratch`.
3. Set the asker record to `stopped`. Keep the parent live.
4. Complete the question to open review.
5. Let later chat messages move the ask event above the feed viewport.

The asker name has 32 characters. It meets the name pattern in `internal/agent/register.go`.
The ask comes from the checked scratch executable through the actual `hive ask` command.

Expected under the fix brief and manager fit rule:

- The full `? ASKER asks · confirm` title.
- An explicit stopped state inside the card.
- The full parent destination, Send row, and Send hint.
- Unchanged parent delivery.

Actual card at 80x24:

```text
┏━ ? asker-scratch-with-long-name-abc asks · confirm ━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃   Change → Yes                                                               ┃
┃                                                                              ┃
┃ › Send to parent-scratch                                                     ┃
┗━ enter sends to parent-scratch · ↑↓ question to change · x drop ━━━━━━━━━━━━━┛
```

The title and destination remain complete. The explicit stopped state is absent.
When the ask event is above the feed viewport, the screen has no other explicit signal for that asker.
The Send destination and its hint establish routing. They do not identify the asker as stopped.

At 120x40, the same names show the full caption and pass.
Actual parent delivery passes at both widths.

### Evidence and controls

- `verifier3/no-caption/{80,120}-review.{ansi,txt,png}`: normal-mode captures at both sizes.
- `verifier3/no-caption/{80,120}-before.json`: actual accepted records, still open before Send.
- `verifier3/no-caption/{80,120}-after.json`: final replies, destination, and delivery attempts.
- `verifier3/no-caption/{80,120}-prompt.txt`: literal expected prompts, matched against recipient output.
- `verifier3/no-caption/asker-record.json` and `parent-record.json`: exact names, status, and parent relation.
- `verifier3/no-caption/80-feed-event-{visible,hidden}.{ansi,txt,png}`: feed control on the same accepted ask.
- `verifier3/no-caption/feed-control-accepted-record.json` and `feed-control-later-chat.json`: control inputs.
- `verifier3/no-caption/feed-control-result.txt`: the card stays byte-equal when later messages hide the event.
- `verifier3/no-caption-model.txt`: independent card-specific regression, with `-race`.
- `verifier3/scripts/no-caption-test.go.txt`: complete regression source.
- `verifier3/card-status.txt`: checks only the card region. Header and feed words cannot satisfy the stopped-state check.

Regression command in the isolated task source copy:

```sh
VERIFIER_CAPTURE=ABSOLUTE_OUTPUT go test ./internal/tui -race -count=1 -v -run '^TestVerifierStoppedReviewWhenNeitherCaptionFits$'
```

The 80-column case fails with `review card does not identify the asker as stopped`.
The 120-column case passes. The regression checks the real model and its rendered card.
It replaces only the external answer operation and sends no answer.

The cause is `askDockView` in `internal/tui/ask.go`.
It sets the stopped caption to empty, then tries two captions.
When neither caption fits, it leaves the caption empty and draws no stopped-state replacement inside the card.
`verifier3/no-caption-before-fix.txt` also records the pre-fix overlap for these names.
I made no source change during this verification.

## Standard stopped review and the old regression

The manager permits `asker stopped · to PARENT` only when the full caption would overlap the title.
The user-approved package has no compact stopped-review grid.
Decision 114 records this manager fit rule under the fix brief.
It does not authorize loss of the explicit stopped state.

For `asker-scratch` and `parent-scratch`:

| Size | Full title | Caption | Full destination, Send row/hint, and delivery |
| --- | --- | --- | --- |
| 80x24 | pass | `asker stopped · to parent-scratch` | pass |
| 120x40 | pass | `asker stopped · answer goes to parent-scratch` | pass |

The prior `TestVerifierCompactStoppedReviewTitle` was rerun unchanged.
It finds the full title and **still fails** its extra full-caption literal at 80 columns.
`verifier3/original-compact-regression.txt` preserves this result.
It is not reported as an unchanged passing test.
Its original source remains in `verifier2/scripts/additional_test.go.txt`.
A copy is in `verifier3/scripts/unchanged-additional-test.go.txt`.

## V1: red runs, deliberate breaks, and green runs

Commands:

```sh
# Base 25f544e3, with task tests and their required data copied in:
go test ./internal/tui -count=1 -v -run 'TestAskReview|TestAskCompactReview|TestAskEachCompletionPathOpensTheReview|TestDemoFixtureShowsTheReview'
# Pre-multiline-fix source:
go test ./internal/tui -count=1 -v -run '^TestAskReview(KeepsMultilineTextInsideTheDock|ReachesEachLineOfATallLabel)$'
# Pre-title-fix source:
go test ./internal/tui -count=1 -v -run '^TestAskCompactReviewForAStoppedAskerKeepsTheTitle$'
# Task implementation, from the worktree root:
go test ./internal/tui -count=50 -run 'TestAsk|TestHelp|TestDemoFixtureShowsTheReview|TestEachBoardFixtureLoadsAtItsSpecSize'
go test ./internal/tui -race -count=1 -v -run 'TestAsk|TestHelp|TestDemoFixtureShowsTheReview|TestEachBoardFixtureLoadsAtItsSpecSize'
```

Logs: `verifier3/red-base.txt`, `multiline-red.txt`, `compact-title-red.txt`, `green-count50.txt`, and `green-race.txt`.
The focused race run has 40 passing top-level tests.
The red runs fail for behavior, without build failures.
The two base passes retain the empty-Other and blocked-caller guards.

`verifier3/mutations.txt` records nine detected breaks:
empty Other, blocked caller, review digits, card draft, refreshed draft, edit return,
truncated rows, frame keys, and tall-row paging.
Each break runs in an isolated source copy. I restore the source after each run.

Both earlier multiline regressions also run again:

```sh
go test ./internal/tui -count=1 -v -run '^TestVerifier(ReviewKeepsMultilineLabelInsideDock|MultilineReviewFieldsAndReplies)$'
go test ./internal/tui -race -count=1 -v -run '^TestVerifier(ReviewKeepsMultilineLabelInsideDock|MultilineReviewFieldsAndReplies)$'
```

They fail before the multiline fix and pass after it.
Disabling the line-feed boundary handling makes the field/reply regression fail.
Logs: `original-multiline-red.txt`, `original-multiline-green.txt`, and `linefeed-mutation.txt` in `verifier3/`.
The replies retain the literal multiline Other and `note` values.

### Changed existing expectations

`verifier3/changed-expectations.diff.txt` preserves the exact test diff.
I removed, skipped, or weakened no existing test.

| Existing test | Approved cause |
| --- | --- |
| `TestAskBoardsMatch` | Three approved review boards. |
| `TestAskCardHasTheFocusInNormalMode` | Review hints replace immediate-send hints. |
| `TestAskTwoQuestionsWithMarksAndANote` | Review and keep-`note` hints, with one extra Send press. |
| `TestAskOtherOpensATextLine` | Keep-answer hint and one extra Send press. |
| `TestAskStoppedAskerShowsTheParent` | Review-for-parent hints. The parent-route check stays. |
| `TestAskDigitPicksAndSendsASingleQuestion`, `TestAskArrowsAndEnterPickTheCurrentOption`, `TestAskEnterFromANoteBeforeAPickPicksTheCurrentOption` | One extra Send press. |
| `TestAskNoteTabDropsAndCtrlBracketKeeps`, `TestAskMouseFocusesAndScrolls`, `TestAskOtherInAMultiSelectQuestion`, `TestDemoBlocksTheAnswer` | One extra Send press after the completion path. |
| `TestHelpListsTheAskKeys` | Approved review Help line. |
| `TestEachBoardFixtureLoadsAtItsSpecSize`, `TestCheckAcceptsTheAskBoards` | Three approved board registrations. |

These causes follow decisions 113 to 119 and package `6c466dc6`.
The compact title fix changes no existing test expectation.

## V2: approval comparisons

```sh
mkdir -p .hive/tmp/verifier135c/exports
go run ./internal/boardexport/cmd/boardexport -out .hive/tmp/verifier135c/exports .hive/spec/tui-v2/design
go test ./internal/boardexport/... ./internal/lib/cellgrid/... ./internal/tui/... -count=1 -v
```

The exporter exits 0 and prints nothing. All board and golden checks pass.
`verifier3/grid-compare.txt` records 59 grid comparisons and 3 fixture comparisons.
Every grid comparison checks rune, colour token, and weight through `internal/lib/cellgrid`, without a mask.
The approval files come from `6c466dc6:.hive/spec/tui-v2/tasks/135-ask-confirm/evidence/new-grids/`.

Changed existing boards: `ChatAsk`, `ChatAskMulti`, `ChatAskPreview`, `ChatAskStopped`, `CompactChatAsk`, and `Help`.
New boards: `ChatAskReview`, `ChatAskReviewMulti`, and `CompactChatAskReview`.
The other 41 board sources, grids, and fixtures equal current `dev`.
All design files remain unchanged since `235892c0`.
`verifier3/board-scope.txt` lists each board.
`verifier3/approval-chat.json` preserves chats `c185` to `c189` and the later approval.

## V3: live interactions and visual checks

I built the task binary and used isolated herdr 0.9.5, protocol 22.
The private session is `verify135c`, under `/tmp/hv135-verify3`.
Each TUI uses a scratch home, `HIVE_NO_STARTUP=1`, `--no-startup`, and `COLORTERM=truecolor`.
Demo runs use the exact board size in a 140x48 pane.
Normal runs use a private tmux window at 120x40 or 80x24.
No test TUI or test prompt uses the real home.

All 16 initial captures equal their grids: `verifier3/live-grid-compare.txt`.
Each capture is stable across two reads 500 ms apart, without truncation or a scroll offset.
The live comparison applies the spec exception for invisible attributes of space cells.
All other rune, token, and weight values match. The check uses no mask.

I captured all nine approved boards again with fresh Firefox profiles.
I compared them with the new terminal PNGs.
The PNG pairs are also in `verifier3/visual-pairs-{1,2,3}.png`.
There are 65 new board and terminal PNG captures, plus the three pair sheets.

At both widths, the live checks cover:

- All five completion paths, with review before Send.
- Question edits and return to review with Send selected.
- Ignored digits, clicks, space, tab, text, and other frame keys.
- J/K draft retention, dismissal, and ordinary record refresh.
- Multi-picks, full Other text, long `note` values, wrapping, and tall-row paging.

The six multiline scenarios have 11 review captures each.
All physical body rows retain both borders and stay within the dock cap.
Arrow navigation reaches `ROW01` through `ROW30` and `FINAL-LABEL-SENTINEL`.
Multiline Other and `note` fields remain readable.
`verifier3/multiline-check.txt` records these independent physical-row checks.
The two known pre-existing editor/feed defects are outside this review change and remain unchanged.

Exact commands and socket requests: `verifier3/live-commands.txt`.
Complete scripts: `verifier3/scripts/`.
Interaction logs: `live-run.txt`, `delivery-run.txt`, `multiline-run.txt`, `no-caption-live.txt`, and `feed-control.txt` in `verifier3/`.

### Visual table

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header and tabs | status glyphs/words, `? N`, tab labels and underline | matches | All approved initial captures equal their grids. |
| Feed and message input | asker names, open/waits words, `message`, input instruction | matches | Approved layouts remain unchanged. |
| Dock frame and card count | heavy border, `open`, `I of N` | matches | Same dock and height cap. |
| Option editor | numbers, `[x]`, `[ ]`, `›` | matches | Draft picks remain after edits. |
| Preview | `preview · N`, framed preview text | matches | Capture equals the approved Preview grid. |
| `note` editor | `note ›`, `█`, `enter keep note` | matches | Approved hint change. |
| Review title | full `confirm` word | matches | Present in approved grids and both standard stopped cases. |
| Review answers | headers, `→`, full labels, Other text and `note` values | matches | Wrapping and scrolling retain the text. |
| Highlighted question | `›`, `enter change` | matches | Arrow selection and edits follow the approved rules. |
| Send row | `› Send`, or `› Send to PARENT` | matches | The card selects Send after entry and after edits. |
| Full stopped caption | `asker stopped · answer goes to PARENT` | matches | Retained wherever it fits. |
| Compact stopped caption | `asker stopped · to PARENT` | overridden | Manager fit decision under the fix brief and decision 114. |
| No-fit stopped state at 80x24 | no explicit stopped word inside the card | differs | Both caption attempts fail. The card draws no stopped-state replacement. |
| Parent destination and Send hint | full `parent-scratch`, `enter sends to parent-scratch` | matches | Retained even in the failing no-fit case. Actual routing also passes. |
| Footer and review hints | `[ask]`, `[note]`, `[other]`, `[review]`, action words | matches | Colour is supplemental. |
| Help | `ask review:` and its keys | matches | Capture equals the approved Help grid. |
| Tall and multiline review | `›`, scroll mark, line text and sentinels | matches | Every segment is reachable. Both body borders remain intact. |

Plain-text captures retain all tested state signals except the reported no-fit stopped state.
The plain-text capture has the same failure.

## V4: real TUI delivery through checked scratch executables

Before any test delivery, the recipient panes show the explicit compiled scratch executable path.
Before each setup, `/proc/PID/exe` equals `/tmp/hv135-verify3/bin/pi`.
No installed Pi process runs for these checks.
`verifier3/stub-executables.txt` records the executable checks and hash.
Process, startup, and recipient output remain in `verifier3/delivery/` and `no-caption/`.

| Size | Edited answer | Stopped asker | Dismiss | Long-name stopped asker |
| --- | --- | --- | --- | --- |
| 120x40 | `a1`, pass | `a2`, pass | `a3`, pass | `a9`, pass |
| 80x24 | `a4`, pass | `a5`, pass | `a6`, pass | `a8`, pass |

Before Send, each record remains open without replies or delivery attempts.
Neither recipient gets a prompt, including after edits and the watched-record refresh.
Send delivers the final draft once. Literal expected edited replies:

```json
[{"picks":[2,3]},{"picks":[3],"other":"Next release","note":"keep this"}]
```

Literal expected prompt for `a1`:

```text
⬢ user: answer to ask a1 from asker-scratch · Scope → Help line, Header count; Order → Next release · note: keep this
```

Stopped cases reach only the live parent and retain `(asker stopped)`.
Dismiss reaches only the asker and retains `No answer comes.`
Each of the eight final records has `attempts: 1` and the expected `delivered_to`.
The feed-control ask stays open and sends no answer.

## V5: scope, gate, and cleanup

Both base checks report `Already up to date.`
`origin/dev` at `3fbbb17f` is an ancestor of the checked worker commit.
`verifier3/diff-stat.txt` records `git diff dev...HEAD --stat`.
All product changes are within the named files or tests beside them.
This verification adds only this report and `verifier3/` evidence.

`verifier3/gate.txt` preserves the full root gate output:

| Command | Result |
| --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | No output, exit 0. |
| `go mod tidy -diff` | No output, exit 0. |
| `go vet ./...` | No output, exit 0. |
| `golangci-lint run ./...` | `0 issues.`, exit 0. |
| `go test -race ./...` | All packages pass, exit 0. |

I corrected two setup errors before their counted runs: a missing export directory and missing imports in the new scratch regression.
An initial feed probe used invalid chat delivery data, so the loader correctly ignored those messages.
The final probe uses accepted coordinator entries and checks the visible feed.
Those setup logs remain separate. They are not counted as behavioral red runs.

Cleanup is complete before this verdict commit.
`verifier3/cleanup.txt` records the checks:

- I stopped the private herdr server, its panes, and all three compiled recipient stubs.
- The private tmux server is absent.
- `/tmp/hv135-verify3` and `.hive/tmp/verifier135c` are absent.
- The worktree `.hive/tmp/` folder is empty.
- I removed scratch source copies, homes, binaries, exports, HTML, and Firefox profiles.
- Final evidence, scripts, accepted inputs, and captures remain in this task folder.

The manager will assign the no-fit stopped-state fix and start a fresh verifier.
