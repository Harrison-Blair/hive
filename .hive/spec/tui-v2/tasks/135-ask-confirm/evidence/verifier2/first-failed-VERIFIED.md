# Task 135: independent verification — fail

Verifier: `t135-verifier-1`. Date: 2026-10-23.
Checked worker commit: `d19709738bd393d4db823446110f2846fefc6e6f`.
Base: `25f544e3`. Current `origin/dev`: `981e21de`.
Approval: coordinator record 554, package `6c466dc6`, user chats `c208` and `c209`.
The approval includes Send after an edit and arrow paging through a tall row.

**Verdict: fail. V3/AC2 fails for accepted multi-line option labels.**
The approved grids, required interaction sequences, delivery checks, and gate pass.
No product source, worker evidence, design file, or golden grid changed during this verification.

All paths below are relative to this evidence folder.
New run output, captures, and scratch script copies are in `verifier/`.

## Verification items

| Item | Result | Independent run and observation |
| --- | --- | --- |
| V1 Tests red and green | pass | All 16 task tests ran against the base. Fourteen fail. Two retained guards pass. Nine deliberate breaks fail their target tests. The focused tests pass with `-count=50` and `-race`. |
| V2 Approved boards and grids | pass | All 50 exports equal their committed grids. All 9 installed grids equal the approval snapshot. The 3 new fixtures equal their approval snapshots. No mask. All board and golden tests pass. |
| V3 Live behavior and visual check | fail | The 16 initial captures equal their grids. Required keys and flows pass at both widths. Accepted multi-line labels break review borders. A tall multi-line label loses text and exceeds the dock height cap. |
| V4 End-to-end delivery | pass | Six normal-mode runs pass: answer, stopped asker, and dismiss at both widths. I checked executable identity and stub output before any test prompt. |
| V5 Scope, gate, cleanup | pass | The branch contains current `origin/dev`. Only named files change. All five gate commands pass. I removed all scratch resources. |

## Blocking finding: multi-line labels break the review

### Accepted input

`verifier/multiline/short-request.json` contains this exact request:

```json
{
  "questions": [
    {
      "question": "Which label?",
      "header": "Label",
      "options": [
        {"label": "First line\nSecond line", "description": "a valid multiline label"},
        {"label": "No", "description": "no"}
      ]
    }
  ]
}
```

The scratch agent ran `hive ask` with this file.
Hive accepted it and wrote open record `a14`.
`verifier/multiline/short-accepted-record.json` preserves that record.
`verifier/multiline/asker-scratch-creation-pane.txt` preserves the command output.
The regression also checks that `ask.Parse` accepts the request.

### Reproduction and difference

1. Create the ask from the scratch agent pane.
2. Open the normal TUI at 120x40 or 80x24.
3. Press `1` to pick the first option and open review.
4. Read the review body borders.

Expected: the full label stays inside the dock, with a left and right border on each body row.
Actual: the first text row has no right border.
The continuation row has no left border.
This failure occurs at both widths.

Evidence:

- `verifier/multiline/short-120-02-review.{ansi,txt,png}`
- `verifier/multiline/short-80-02-review.{ansi,txt,png}`
- `verifier/multiline-regression.go.txt`: the complete independent regression source.
- `verifier/multiline-regression.txt`: both subtests fail on the delivered code.

Exact regression command, from the scratch source copy:

```sh
go test ./internal/tui -count=1 -v -run '^TestVerifierReviewKeepsMultilineLabelInsideDock$'
```

Both failures report `review body has a broken border`.
I removed the source copy after the run.
The durable regression source remains in evidence.

A second accepted request has 30 lines and `FINAL-LABEL-SENTINEL` in its label.
`verifier/multiline/tall-request.json` preserves the exact request.
At both widths, the review draws broken rows and truncates text with `…`.
The sentinel never shows in review or after five up and five down presses.
At 80x24, the Send row and the message panel disappear from the captured screen.
The review body exceeds the half-screen dock cap.
See `verifier/multiline/tall-*-result.txt` and the associated captures.

Source locations: `internal/tui/ask.go:708`, `:777`, and `:787`.
`styledWords` splits only at spaces.
Embedded line feeds remain inside a logical review line.
The dock counts logical lines, while the panel renders the embedded line feeds as physical rows.

This finding concerns V3/AC2 full review text, wrapping, scrolling, and the retained dock cap.
The manager accepted the finding and requested a rendering fix under the existing approval.
The fix needs no new user design decision.
Do not reject this accepted input or change the answer prompt format to address the finding.

## V1 details

Commands:

```sh
# Base source copy, with the final test file and review fixtures copied into it:
go test ./internal/tui -count=1 -v -run 'TestAskReview|TestAskEachCompletionPathOpensTheReview|TestDemoFixtureShowsTheReview'
# Delivered implementation:
go test ./internal/tui -count=50 -run 'TestAsk|TestHelp|TestDemoFixtureShowsTheReview|TestEachBoardFixtureLoadsAtItsSpecSize'
go test ./internal/tui -race -count=1 -v -run 'TestAsk|TestHelp|TestDemoFixtureShowsTheReview|TestEachBoardFixtureLoadsAtItsSpecSize'
```

Logs: `verifier/red-base.txt`, `green-count50.txt`, and `green-race.txt`.
Base failures show immediate delivery instead of review, missing review rows, or unsupported fixture focus.
The two base passes retain the empty-Other check and the blocked-caller check.

`verifier/mutations.txt` records nine detected breaks:
empty Other, blocked caller, review digit, card draft, refresh draft, edit return,
truncated row, frame keys, and tall-row paging.
Each break ran in a scratch source copy.
Each target failed for behavior, with no build failure.
I restored the source after each break.
`verifier/scripts/mutations.py.txt` preserves the exact replacements.

The additional multi-line regression fails under V3. It exposes a gap in the task tests.

### Changed existing expectations

I checked the test diff and the approval against each change.
The diff removes or skips no existing test.

| Test | Approved cause |
| --- | --- |
| `TestAskBoardsMatch` | Add the three approved review boards. |
| `TestAskCardHasTheFocusInNormalMode` | Replace send hints with review hints. |
| `TestAskTwoQuestionsWithMarksAndANote` | Review hint, `enter keep note` hint, and one additional Send press. |
| `TestAskOtherOpensATextLine` | Keep-answer hint and one additional Send press. |
| `TestAskStoppedAskerShowsTheParent` | Review-for-parent hints and one additional Send press. |
| `TestAskArrowsAndEnterPickTheCurrentOption` | One additional Send press. |
| `TestAskDigitPicksAndSendsASingleQuestion` | One additional Send press. |
| `TestAskEnterFromANoteBeforeAPickPicksTheCurrentOption` | One additional Send press. |
| `TestAskMouseFocusesAndScrolls` | A click opens review before Send. |
| `TestAskNoteTabDropsAndCtrlBracketKeeps` | One additional Send press. |
| `TestAskOtherInAMultiSelectQuestion` | One additional Send press. |
| `TestDemoBlocksTheAnswer` | One additional Send press before the demo block. |
| `TestHelpListsTheAskKeys` | Require the approved review help line. |
| `TestEachBoardFixtureLoadsAtItsSpecSize` | The three new boards increase the count from 47 to 50. |
| `TestCheckAcceptsTheAskBoards` | Register the three new board sizes. |

The tests retain the returned model where the new review step requires it.
Decisions 113 to 119 and package `6c466dc6` authorize these changes.

## V2 details

```sh
mkdir -p .hive/tmp/verifier135/exports
go run ./internal/boardexport/cmd/boardexport -out .hive/tmp/verifier135/exports .hive/spec/tui-v2/design
go test ./internal/boardexport/... ./internal/lib/cellgrid/... ./internal/tui/... -count=1 -v
```

The exporter exits 0 with no output. It writes 50 grids.
`verifier/grid-compare.txt` records 59 equal comparisons through `internal/lib/cellgrid`.
Each comparison includes rune, colour token, and weight, with no mask.
`verifier/fixture-approval.txt` records three equal fixture comparisons.
`verifier/boards.txt` records all board and golden checks.

The approval snapshot uses the full path:

```text
6c466dc6:.hive/spec/tui-v2/tasks/135-ask-confirm/evidence/new-grids/BOARD.grid
```

Changed existing boards: `ChatAsk`, `ChatAskMulti`, `ChatAskPreview`,
`ChatAskStopped`, `CompactChatAsk`, and `Help`.
New boards: `ChatAskReview`, `ChatAskReviewMulti`, and `CompactChatAskReview`.
All other board sources, grids, and fixtures stay unchanged.
The diff adds or changes no golden mask.

## V3 details and visual table

The task binary ran on isolated herdr 0.9.5, protocol 22.
The virtual terminal had 140 columns and 48 rows.
Demo runs used `--size 120x40` or `--size 80x24`.
Normal runs used a private tmux window resized to the exact required size.
Each run used `HIVE_NO_STARTUP=1`, a scratch home, and `COLORTERM=truecolor`.
No test TUI ran on the real home.

Each initial read was stable across two reads 500 ms apart.
Each read reported no truncation and a scroll offset of zero.
The capture cropped the top-left board rectangle.
All 16 initial captures have zero differing cells: `verifier/live-grid-compare.txt`.
The live comparison follows the spec exception for invisible space-cell attributes.
It compares each visible rune, colour token, and weight. It uses no mask.

I captured the nine changed boards as PNG files with fresh Firefox profiles.
I compared each board with its terminal capture.
Terminal PNG and ANSI captures remain in `verifier/live/`.
I removed the temporary board PNG files after comparison.

Both widths pass these required interaction checks:

- Five completion paths: enter, digit, click, `note`, and Other text.
- Question, review, highlighted question, edit, review with Send selected, then Send.
- Digits, clicks, space, tab, text, backspace, and other frame keys leave review unchanged.
- `J` then `K` restores the reviewed draft and selection.
- `x` retains the dismiss action.
- Long single-line `note` values page through every word, including middle and final sentinels.

The normal delivery runs also force a watched-record refresh during highlighted review.
The refreshed review keeps the selection, picks, and `note` at both widths.
Model tests cover the unchanged global quit behavior.
Logs and captures: `verifier/live-run.txt`, `flow-*`, `completion-*`, `tall-*`, and `delivery/`.
Exact commands and socket requests: `verifier/live-commands.txt`.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header and tabs | status words, `? N`, screen names, active underline | matches | All approved initial captures equal their grids. |
| Chat feed | names, times, `? asked`, open/waits words | matches | Existing feed layout stays unchanged. |
| Input panel | `message`, status words, input instruction | matches | Existing input layout stays unchanged. |
| Dock frame and count | heavy border, `open`, `I of N` | matches | Approved fixtures retain the existing dock. |
| Option editor | option number, `[x]`, `[ ]`, `›` | matches | Existing picks stay visible after an edit. |
| Preview box | `preview · N`, framed text | matches | `ChatAskPreview` matches its board. |
| `note` editor | `note ›`, cursor `█` | matches | Updated hint says `enter keep note`. |
| Review title | `confirm` | matches | Decision 114 in the approved fixtures. |
| Review answers | header, `→`, full labels and `note` text | matches | Approved one-question and multi-select fixtures. |
| Other answer | full typed text | matches | Review replaces the Other label with its text. |
| Selected review row | `›` | matches | Arrows highlight a question or Send. |
| Send row | `Send` or `Send to PARENT` | matches | The review selects Send after entry and after an edit. |
| Stopped route | `asker stopped`, parent name | matches | Parent routing remains explicit in plain text. |
| Review border hints | `enter send` or `enter change`, arrows, `x drop` | matches | Approved review actions. |
| Footer | `[ask]`, `[note]`, `[other]`, or `[review]` | matches | Each focus state has a word. |
| Help | `ask review:` and its keys | matches | Approved Help change. |
| Tall single-line answer | `›`, scroll mark `┃`, both sentinels | matches | Arrow paging reaches every word. |
| Accepted multi-line answer | first and second label lines | differs | Embedded line feeds break the left and right review borders. |
| Tall multi-line answer | `ROW01` to `ROW30`, final sentinel | differs | The review truncates text, exceeds the dock cap, and hides the final sentinel. |

The plain-text captures preserve all state words and glyphs.
No tested state depends on colour alone.

## V4 details

Read the worker incident and `scratch-harness-stub-runs-real-pi` before setup.
The executable stub was a compiled Go program at `/tmp/hv135-verify/bin/pi`.
Both panes printed `STUB_READY` with the executable path, name, and PID.
`pane.process_info` identified each foreground PID.
`/proc/PID/exe` equalled the explicit scratch stub path before any test prompt.
`verifier/stub-executables.txt` records the executable checks and hash.
No installed Pi process ran in these checks.

The scratch agent created each ask from its own pane.
Before Send, the ask record stays open, has no replies or attempts, and neither target receives a new prompt.
An edit returns to review with Send selected.
A second edit sets Other text and keeps the `note`.
Send stores the final replies and delivers the exact expected prompt once.

| Width | Answer | Stopped asker | Dismiss |
| --- | --- | --- | --- |
| 120x40 | `a8`, pass | `a9`, pass | `a10`, pass |
| 80x24 | `a11`, pass | `a12`, pass | `a13`, pass |

Exact reply example:

```json
[{"picks":[2,3]},{"picks":[3],"other":"Next release","note":"keep this"}]
```

Exact prompt example:

```text
⬢ user: answer to ask a8 from asker-scratch · Scope → Help line, Header count; Order → Next release · note: keep this
```

Stopped delivery reaches only `parent-scratch` and includes `(asker stopped)`.
Dismiss reaches only the asker and retains `No answer comes.`
Every completed record has `attempts: 1` and the expected `delivered_to`.
See `verifier/delivery-run.txt` and `verifier/delivery/` for before/after records, prompts, and pane output.

## V5 details and cleanup

`git fetch origin` and `git merge --no-edit origin/dev` report `Already up to date.`
`verifier/merge.txt` records the exact commits before the final gate.
`verifier/diff-stat.txt` records `git diff dev...HEAD --stat`.
Outside this task folder, all changed files are in the brief.
This verifier adds only this report and the `verifier/` evidence folder.

`verifier/gate.txt` preserves the full gate output:

| Command | Result |
| --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | No output, exit 0. |
| `go mod tidy -diff` | No output, exit 0. |
| `go vet ./...` | No output, exit 0. |
| `golangci-lint run ./...` | `0 issues.`, exit 0. |
| `go test -race ./...` | All packages pass, exit 0. |

Cleanup is complete. `verifier/cleanup.txt` records the checks.

- The isolated herdr server and its panes stopped.
- The private tmux server stopped.
- `/tmp/hv135-verify` is absent.
- `.hive/tmp/verifier135` and the two scratch identity/history files are absent.
- The worktree `.hive/tmp/` folder is empty.
- The worker board PNG folder `/home/penguin/source/hive/.hive/tmp/135-boards/` is absent, as requested.
- No scratch hive or stub process runs.

Final evidence and terminal captures remain in this task folder.
The manager must start a fix worker, then a fresh independent verifier.
