# Ask cards: review every answer before send

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, goal `g10-ask-confirm`, wave A |
| Tier | strong |
| Depends on | `134-agents-hide-managers`, landed at `25f544e3`. Base: `dev` after that landing. |
| Grant | Goal `g10-ask-confirm`. The user chose review on every card in chat `c186` and approved the plan in `c188`. |
| Evidence | The `evidence/` folder beside this brief. Put `VERIFIED.md` there. |

## User approval: 2026-10-23

Coordinator inbox record 554 relays the user approval: use package `6c466dc6` as the golden files.
The approval uses the same rule as tasks 126, 131, and 134.
It includes Send selected after an edit returns to review.
It includes arrow paging through a highlighted row taller than the dock.

Install the approved review boards and golden grids.
Keep every other grid unchanged.
Record coordinator record 554 and package `6c466dc6` in the evidence.
The manager clears task and wave `needs_user`.
No approval hold remains for these grids.
Report any rendering that disagrees with the approved package before a design change.

V4 passed at commit `259aa9d7` before this approval.
Repeat a worker delivery check only if a later behavior change affects that check.
The independent verifier must rerun delivery checks with the intended scratch stub executable.
Check that executable before any test delivery.
Never run a test TUI or send a test prompt on the real home.

## Current rework: stopped state when route captions do not fit

Verifier `t135-verifier-3` recorded fail in manager inbox records 574 and 575.
Its evidence commit `cef43222` is on this branch and its remote.
Read `evidence/VERIFIED.md` and `evidence/verifier3/` beside this brief.

At 80x24, neither route caption fits for the accepted asker `asker-scratch-with-long-name-abc`.
The parent is `parent-scratch`.
The full review title, Send destination, hint, and parent delivery stay correct.
The card loses every explicit stopped-state word.
Later valid chat messages move the ask event above the feed viewport.
The card must identify the stopped state independently of that feed event.

Agent registration permits ASCII names of at most 32 characters.
Check `namePattern` in `internal/agent/register.go`.
Cover valid names at this limit, including a long parent.

1. Read `evidence/verifier3/scripts/no-caption-test.go.txt` and its failure output.
2. Reuse the accepted records and visible/hidden feed controls in `evidence/verifier3/no-caption/`.
3. Add permanent regressions before the source fix.
4. Check the failure for missing stopped-state text inside the card at 80x24.
5. Cover full-caption, short-caption, and neither-route-caption cases at 80x24 and 120x40.
6. Keep the full review title and an explicit `asker stopped` signal inside the card.
7. Keep the full parent destination in Send and its existing hint before confirmation.
8. Keep this information independent of the scrolling feed and unrelated header words.
9. Check question edits, return to Send, and resizing between the two widths.
10. Keep the dock cap, wrapping, paging, original reply values, and actual parent route.
11. Run focused and race tests, normal live checks, approved grid comparisons, and the full root gate.
12. Record this rework separately from all earlier evidence.

Manager fit rule: keep the full route caption wherever it fits.
Use `asker stopped · to PARENT` when the full sentence would overlap the title.
When neither route caption fits, a card-local `asker stopped` fallback can supply the state.
The full parent still appears in Send and its hint.
Add no key. Add no layout row unless the existing card cannot retain the required information.
Do not leave the state empty when both route captions fail.

Keep every approved board, golden grid, and design fixture unchanged.
Keep the original multiline fix and all three failed verdicts with their evidence.
Do not fix the pre-existing editor/feed defects in this task.
The user cancelled goal g11. Keep .hive/spec tracked and retain the current task/evidence/test workflow.
The manager starts another fresh verifier after this fix.

## Earlier rework: compact stopped-review title

Verifier `t135-verifier-2` recorded fail in manager inbox record 565.
Its committed evidence is `235892c0`, on this branch and its remote.
Read `evidence/VERIFIED.md` and `evidence/verifier2/` beside this brief.

At 80x24, the stopped-asker route text overwrites the review title.
The normal scratch TUI shows `asks · confi asker stopped`.
Decision 114 requires the full title `? asker-scratch asks · confirm`.
The route and actual parent delivery still work.
The existing design approval covers this rendering fix.

1. Read the regression in `evidence/verifier2/scripts/additional_test.go.txt`.
2. Add a permanent compact stopped-review regression before the fix.
3. Run it against the current code and check the recorded title failure.
4. Fit the full review title and explicit stopped-asker route information without overlap.
5. Check the compact stopped state in the normal scratch TUI at 80x24 and 120x40.
6. Keep the multiline wrapper fix and every approved grid unchanged.
7. Keep prompt format, routing, and reply values unchanged.
8. Run focused and race tests, live key checks, approval comparisons, and the full root gate.
9. Record this rework separately from the earlier reports and evidence.

Do not change board sources, golden grids, or design fixtures for this fix.
Keep both failed verdicts and their evidence intact.
Do not fix the two pre-existing editor/feed defects in this task.
The manager records those defects in friction separately.
The manager starts another fresh verifier after this fix.

## Earlier rework: multiline review

Verifier `t135-verifier-1` recorded a fail in manager inbox records 560 and 561.
Evidence commit `008e9d2d` is on this task branch and its remote.
Read `evidence/VERIFIED.md` and `evidence/verifier/` beside this brief.

The accepted label `First line\nSecond line` breaks the review side borders at both widths.
A tall accepted multiline label also loses text and exceeds the dock height cap.
Its final sentinel stays hidden during review navigation.
`ask.Parse` accepts these inputs.
The review wrapper splits at spaces but leaves line feeds inside logical rows.

This fix falls within AC2 full text, wrapping, scrolling, and the retained height cap.
The existing design approval covers this rendering fix.

1. Read `evidence/verifier/multiline-regression.go.txt` and its failed output.
2. Read the short and tall accepted requests in `evidence/verifier/multiline/`.
3. Add regressions beside the TUI tests before the source fix.
4. Run them against the current implementation at 120x40 and 80x24.
5. Check failures for broken borders, dock overflow, and unreachable sentinel text.
6. Fix the shared review wrapper to handle accepted line breaks as layout rows.
7. Keep each physical body row inside its left and right borders.
8. Keep wrapping, the dock cap, and arrow access to every text segment.
9. Check other accepted multiline fields that use this same review wrapper.
10. Keep the original input and delivered reply values unchanged.
11. Repeat focused and race tests, live multiline checks, approved grid comparisons, and the full root gate.
12. Record the rework evidence separately from the first verifier evidence.

Do not reject or flatten accepted multiline input to avoid the defect.
Do not change the answer prompt format, routing, or record format.
Do not change any board source, golden grid, or design fixture for this fix.
Keep the first failed verdict and its evidence intact.
The manager starts a fresh verifier after the worker completes the fix.

The normal test-first rule applies to task 135.
The user cancelled goal g11. Keep existing repository storage and test rules.

## Authority and plan

Coordinator inbox record 537 assigns this goal.
Read user chats `c185` to `c189` through `hive chat log`.
The approved plan is in `c187`. Its review view is in `c185`.
Task 134 landed, so this task can start.

1. Write failing tests for review before delivery.
2. Add the smallest card state and layout change.
3. Commit code grids and live captures for user approval.
4. Install the grids approved in coordinator record 554.
5. Apply the approved grids and complete each verification item.

## Design

Read `/home/penguin/source/hive/.hive/spec/tui-v2/README.md`.
The approved existing layout references are:

- `/home/penguin/source/hive/.hive/spec/tui-v2/design/ChatAsk.dc.html`
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/ChatAskMulti.dc.html`
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/ChatAskPreview.dc.html`
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/ChatAskStopped.dc.html`
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/CompactChatAsk.dc.html`
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/Help.dc.html`

The user approved this review structure in `c185` and `c187`:

```text
? ASKER asks · confirm
HEADER → FULL PICK
HEADER → FULL PICK
HEADER → FULL PICK

enter send · ↑↓ question to change · x drop
```

Use the same card and dock. Add no popup or separate screen.
The existing palette, frame, dock height cap, and scrolling rules apply.
Use Send as the initial review selection, so another `enter` sends.
Arrows select a question for an edit.
Complete the edited question by returning to review with its updated draft.
Include that navigation in the approval package.
Follow the visual check procedure in the worker skill.
Each state has a glyph or word. Colour only adds to the state.

## Requirements

### AC1 Review before every answer

- Every card gets the review step, including a card with one question.
- `enter` on the last question opens review instead of sending.
- Apply the same rule to each path that completes the last answer.
- This includes a final digit, a click, a `note`, and Other text.
- Entering review writes no answer record and sends no answer prompt.
- In review, `enter` on Send sends the answer.
- The default review selection is Send, so a second `enter` sends.
- Retain the existing validation for an empty Other answer or missing pick.
- Keep the answer prompt format, routing, retries, and record format unchanged.

### AC2 Full review and editing

- Show every question with its full picked labels and `note`.
- Show every marked option of a multi-select question.
- Show the full Other text.
- Wrap long text to the card width without truncation.
- Use the existing dock scroll when the content exceeds its height.
- Keep the highlighted question and its answer reachable at 120x40 and 80x24.
- Up and down highlight a question to change.
- `enter` on that question returns to its option and `note` editor.
- Preserve all other picks and `note` values during the edit.
- Complete the edited question by returning to review before send.
- Show the updated draft there.
- Keep each card draft and review state when `J` or `K` changes cards.
- Keep the draft across ordinary data refreshes.

### AC3 Review keys and mouse

- Only up, down, enter, x, J, and K act as card keys in review.
- Digits 1 to 5 and a click do not change a pick in review.
- Space, tab, and text entry do not change a reviewed answer.
- `x` keeps its existing dismiss behavior.
- `J` and `K` keep their existing card-switch behavior.
- Keep these meanings outside review unchanged.
- Keep the existing global quit behavior.
- Update card hints and Help where this behavior requires them.

### AC4 Grids and user approval

- Render each changed existing board at its board size from the code.
- Add review grids at 120x40 and 80x24.
- Include one question, multiple questions, multi-select, Other, `note` values, and long text.
- Include a highlighted question and the return from editing in the live captures.
- Put the grids, plain text, fixtures, and a changed-row report in `evidence/new-grids/`.
- Put terminal PNG and ANSI captures in the evidence.
- List each existing board that changes and each that stays unchanged.
- Commit and push the approval package.
- Send its commit and path to `manager-hive` through `hive inbox send`.
- Do not change any existing board source, golden grid, or board fixture before approval.
- Do not install new design boards or golden grids before approval.
- The manager sets task and wave `needs_user` if the worker holds for approval.
- Do independent work while approval is pending.
- After approval, install only the approved board and grid changes.
- Add `ChatAskReview`, `ChatAskReviewMulti`, and `CompactChatAskReview` as permanent review references.
- Keep additional long-content scenarios as evidence fixtures unless approval requires another permanent board.
- Make each board source export its approved code grid.
- Add no golden mask.

### AC5 Spec and tests

- Add the approved decisions to the TUI spec after decision 112.
- Write each new behavior test before the implementation.
- Use literal expected selections, `note` text, reply values, and prompt text.
- Prove that entering review does not close the card or deliver an answer.
- Prove that confirmation sends the final edited draft.
- Cover one and several questions, multi-select, Other, `note` values, and long content.
- Cover each completion path and review input restriction.
- Cover card switches and a data refresh during review.
- Cover a stopped asker and the unchanged parent route.
- Cover dismiss from review and the unchanged dismiss prompt.
- Cover both widths, overflow scrolling, and the highlighted question after an edit.
- Preserve the existing checks. Change an expectation only for the approved review step.
- Record the reason for each changed existing expectation.

## Source files

| Path | Purpose |
| --- | --- |
| `internal/tui/ask.go` and tests beside it | Card state, review layout, edits, confirmation, and draft tests. |
| `internal/tui/frame.go`, `mouse.go`, `demo.go`, `chat.go`, and tests beside them | Review focus, hints, mouse guards, demo state, and dock integration only. |
| `.hive/spec/tui-v2/README.md` | The approved review decisions. |
| `.hive/spec/tui-v2/design/*.dc.html`, `*.grid`, `*.fixture.json` | Only the approved review boards and affected existing boards. |
| `internal/boardexport/check.go` and its tests | Register the three approved review board sizes. |
| `testdata/ste-baseline.json` | Add this brief at baseline 0. Counts can fall. |
| This task folder | The brief, fixtures for evidence, and evidence. |

Do not change other files.
Do not change the ask transport or the answer prompt format.
Do not change the exporter except the board-size registration.
Report a required scope change before you edit another source file.

## Verification items

### V1 Tests red and green

Run each new test against the base implementation before the source change.
Check that it fails for the correct reason.
Run it with the task implementation.
Run the focused tests with `-count=50` and with `-race`.
For a guard that passes on the base, break its guarded behavior deliberately.
Record the failure, then restore the source.
Expected: tests detect send without review, input changes during review, or lost draft data.
Expected: the final implementation passes each test.
Record each changed existing expectation and its approved cause.

### V2 Approved boards and grids

After approval, run the exporter on the design folder.
Compare each exported grid with its committed grid through `internal/lib/cellgrid`.
Compare each changed grid with its approved code grid.
Run all existing board and golden checks.
Expected: cell equality for rune, colour token, and weight, with no mask.
Expected: only approved board, grid, and fixture changes.
Record the approval reference and commit.

### V3 Live behavior and visual check

Build the task binary.
Use scratch herdr panes or an isolated scratch server that you start.
Use a scratch home or demo fixtures, `HIVE_NO_STARTUP=1`, and `COLORTERM=truecolor`.
Capture each changed screen and its approved reference at the board size.
Run the full question to review to edit to review to send sequence at both widths.
Test digits, clicks, space, tab, and text while review has focus.
Check that they do not change picks or `note` values.
Run dismiss and card-switch sequences.
Check long `note` values, full Other text, multi-picks, wrapping, and scroll access.
Compare each initial reference capture with its approved grid cell for cell.
Include the visual table required by the worker skill.
Expected: each approved grid matches and each interaction follows AC1 to AC3.
The worker and verifier each run this item.

### V4 End to end delivery

Use a scratch home and an isolated scratch herdr session.
Create an ask from a scratch agent pane.
Complete its questions in the real TUI and read the ask record and target pane.
Expected before review confirmation: the ask remains open and no answer prompt arrives.
Edit a question through review, return to review, and press `enter` on Send.
Expected: the record and delivered prompt contain the final picks and `note` values in the unchanged format.
Repeat for a stopped asker with a live parent and for dismiss from review.
Expected: the parent route and dismiss prompt stay unchanged.
Record exact commands and captures.
Do not send any test message on the real home.

### V5 Diff scope, gate, and cleanup

Merge current `origin/dev` before the final gate.
Run `git diff dev...HEAD --stat`.
Run the full gate from the worktree root:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go mod tidy -diff
go vet ./...
golangci-lint run ./...
go test -race ./...
```

Expected: only the named files change.
Expected: the first two commands print nothing.
Expected: each other command exits 0.
Record the full output.
Close only the panes, sessions, and processes that you started.
Remove task scratch after you record the result.
Keep final evidence and captures.
Check cleanup before completion.

## Finish

Commit and push the task branch without an agent co-author.
Run `hive task complete` only after approval and each verification item passes.
Send the commit and results to `manager-hive`.
The manager starts an independent verifier before landing under goal g10.
