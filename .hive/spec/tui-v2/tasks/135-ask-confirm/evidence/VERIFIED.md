# Task 135: second independent verification — fail

Verifier: `t135-verifier-2`. Date: 2026-10-23.
Worker commit: `935bc91b`. Multiline fix: `cc4c170a`.
Current `origin/dev`: `981e21de`. Original task base: `25f544e3`.
Approval: coordinator record 554, package `6c466dc6`, user chats `c208` and `c209`.

**Verdict: fail. V3 loses part of the required review title at 80x24 for a stopped asker.**
The multiline review fix passes. V1, V2, V4, and V5 pass.
The manager accepted the title finding as within V3 and decision 114.

The first failed report is preserved, byte for byte, in `verifier2/first-failed-VERIFIED.md`.
It equals the report at commit `008e9d2d`.
The first `verifier/` folder and the worker evidence remain unchanged.
All paths below are relative to this evidence folder.
This verification changes no product source, design source, golden grid, fixture in `design/`, or worker report.

## Verification items

| Item | Result | Independent result |
| --- | --- | --- |
| V1 Tests red and green | pass | The 18 task tests run on the base. Sixteen fail for missing review behavior. Two retained guards pass. Nine deliberate breaks detect failures. Both rework tests fail before the fix. Focused tests pass with `-count=50` and `-race`. The original verifier regression passes after the fix. |
| V2 Approved boards and grids | pass | All 50 exports equal their committed grids through `internal/lib/cellgrid`. All 9 installed approval grids equal package `6c466dc6`. All 3 new fixtures equal their approval files. No mask. Board and golden tests pass. |
| V3 Live behavior and visual check | fail | All 16 initial captures equal their grids. Required interaction sequences and multiline review checks pass at both sizes. The compact stopped review shows `confi`, not the required `confirm`. |
| V4 End-to-end delivery | pass | Six normal-mode cases pass: answer, stopped asker, and dismiss at both sizes. Before each setup, the recipient executables equal the compiled scratch stub. Records and delivered prompts match literal expected values. |
| V5 Scope, gate, and cleanup | pass | Current `origin/dev` is included. Only named files change. The full gate passes. All scratch processes, panes, homes, binaries, profiles, and source copies are removed. |

## Blocking finding: compact stopped-review title overlap

### Reproduction

1. Use an 80x24 scratch TUI.
2. Create an ask from `asker-scratch`, whose live parent is `parent-scratch`.
3. Set the asker record to `stopped`.
4. Complete the last question to open review.

Expected, from the Design section and decision 114:

```text
? asker-scratch asks · confirm
```

Actual top border:

```text
┏━ ? asker-scratch asks · confi asker stopped · answer goes to parent-scratch ━┓
```

The stopped-asker note overwrites `rm`. Both names are valid agent names.
The route note, Send row, footer, and actual parent delivery remain correct.
The required full state title does not remain visible.

Evidence:

- `verifier2/delivery/80-stopped-02-review.{ansi,txt,png}`: normal TUI, real scratch ask, and live parent.
- `verifier2/base/live/task-stopped-80-review.{ansi,txt,png}`: a separate normal-mode reproduction.
- `verifier2/compact-stopped-title-regression.txt`: the independent model regression fails on the task implementation.
- `verifier2/scripts/additional_test.go.txt`: complete regression source.

Regression command, in the isolated task source copy:

```sh
go test ./internal/tui -count=1 -v -run '^TestVerifierCompactStoppedReviewTitle$'
```

It fails with `view does not contain "? asker-scratch asks · confirm"`.
The 120x40 stopped review keeps the full title.

The cause is the top-border overlay in `internal/tui/ask.go`, in `askDockView`.
The right note is placed without reserving space for the review title.
Current `dev` has the same overlay helper, but has no review state.
Its compact stopped editor fits the complete header `Model`.
The task adds the longer `confirm` state and loses part of that word.
Thus the new review state fails its design requirement.
This finding does not require a change to the existing approved boards.
Preserve the stopped-asker route information when correcting this state.

## V1: independent red and green runs

Commands:

```sh
# Base 25f544e3, with final task tests, review fixtures, and accepted requests copied in:
go test ./internal/tui -count=1 -v -run 'TestAskReview|TestAskEachCompletionPathOpensTheReview|TestDemoFixtureShowsTheReview'
# Task implementation, from the worktree root:
go test ./internal/tui -count=50 -run 'TestAsk|TestHelp|TestDemoFixtureShowsTheReview|TestEachBoardFixtureLoadsAtItsSpecSize'
go test ./internal/tui -race -count=1 -v -run 'TestAsk|TestHelp|TestDemoFixtureShowsTheReview|TestEachBoardFixtureLoadsAtItsSpecSize'
# Pre-rework source, with final tests:
go test ./internal/tui -count=1 -v -run '^TestAskReview(KeepsMultilineTextInsideTheDock|ReachesEachLineOfATallLabel)$'
# Original independent regression, before and after the fix:
go test ./internal/tui -count=1 -v -run '^TestVerifierReviewKeepsMultilineLabelInsideDock$'
```

Logs: `verifier2/red-base.txt`, `green-count50.txt`, `green-race.txt`, `rework-red.txt`,
`original-and-additional-red.txt`, `original-regression-green.txt`, and `additional-and-original-race.txt`.
The red runs fail for behavior, without a build failure.
The two base passes retain the empty-Other and blocked-caller guards.

`verifier2/mutations.txt` records nine detected breaks:
empty Other, blocked caller, review digits, card draft, refreshed draft, return from edit,
truncated review row, frame keys, and tall-row paging.
The complete replacements are in `verifier2/scripts/mutations.py.txt`.
`verifier2/linefeed-mutation.txt` detects removal of the new line-feed boundary handling.
All deliberate breaks run in an isolated source copy. The source is restored after each run.

### Multiline rework checks

At 120x40 and 80x24, I independently check:

- The original accepted short label `First line\nSecond line`.
- The original accepted 31-line label, including `ROW01` through `ROW30` and `FINAL-LABEL-SENTINEL`.
- Multiline Other text and multiline `note` values.
- A multiline header and a blank line inside a picked label, within the review body.
- Both dock borders, the half-screen cap, full text, arrow access, and unchanged reply values.

The original regression fails before the fix and passes after it.
The added field/reply test also fails before the fix and passes after it, including under `-race`.
Removing the new boundary handling makes the added field/reply test fail.

Literal expected reply in the added test:

```json
[{"picks":[1,3],"other":"other one\nother two","note":"note one\nnote two"}]
```

The original request and the delivered reply keep their line feeds.
The test uses the actual TUI state and answer path. It replaces only the external answer operation.
See `verifier2/additional-multiline-green.txt` and `scripts/additional_test.go.txt`.

### Changed existing expectations

I checked each change against decisions 113 to 119 and approval `6c466dc6`.
`verifier2/changed-expectations.diff.txt` preserves the exact test diff.
No existing test is removed, skipped, or weakened.

| Existing test | Approved cause |
| --- | --- |
| `TestAskBoardsMatch` | Add the three approved review boards. |
| `TestAskCardHasTheFocusInNormalMode` | Change send hints to review hints. |
| `TestAskTwoQuestionsWithMarksAndANote` | Review and keep-note hints, then one additional Send press. |
| `TestAskDigitPicksAndSendsASingleQuestion` | One additional Send press. |
| `TestAskArrowsAndEnterPickTheCurrentOption` | One additional Send press. |
| `TestAskEnterFromANoteBeforeAPickPicksTheCurrentOption` | One additional Send press. |
| `TestAskNoteTabDropsAndCtrlBracketKeeps` | One additional Send press. |
| `TestAskOtherOpensATextLine` | Keep-answer hint and one additional Send press. |
| `TestAskStoppedAskerShowsTheParent` | Review-for-parent hints. The route check stays. |
| `TestAskMouseFocusesAndScrolls` | A click opens review before Send. |
| `TestAskOtherInAMultiSelectQuestion` | One additional Send press. |
| `TestDemoBlocksTheAnswer` | One additional Send press before the demo block. |
| `TestHelpListsTheAskKeys` | Require the approved review help line. |
| `TestEachBoardFixtureLoadsAtItsSpecSize` | Three new boards change the count from 47 to 50. |
| `TestCheckAcceptsTheAskBoards` | Register the three new board sizes. |

## V2: approval comparisons

```sh
go run ./internal/boardexport/cmd/boardexport -out .hive/tmp/verifier135b/exports .hive/spec/tui-v2/design
go test ./internal/boardexport/... ./internal/lib/cellgrid/... ./internal/tui/... -count=1 -v
```

`verifier2/export.txt` is empty. The exporter exits 0 and writes 50 grids.
`verifier2/grid-compare.txt` records 59 cell-equal grid comparisons and 3 equal fixture comparisons.
Each grid comparison checks rune, colour token, and weight through `internal/lib/cellgrid`, without a mask.
Approval lookups use the full Git paths:

```text
6c466dc6:.hive/spec/tui-v2/tasks/135-ask-confirm/evidence/new-grids/BOARD.grid
6c466dc6:.hive/spec/tui-v2/tasks/135-ask-confirm/evidence/new-grids/fixtures/BOARD.fixture.json
```

Changed existing boards: `ChatAsk`, `ChatAskMulti`, `ChatAskPreview`, `ChatAskStopped`, `CompactChatAsk`, and `Help`.
New boards: `ChatAskReview`, `ChatAskReviewMulti`, and `CompactChatAskReview`.
The other 41 grids, their sources, and their fixtures equal current `dev`.
No design file changed after the first failed verdict.
`verifier2/board-scope.txt` names each unchanged grid.
`verifier2/boards.txt` records the passing board and golden checks.
`verifier2/approval-chat.json` preserves the user approval and the original plan.

## V3: live and visual checks

I built the task binary and used isolated herdr 0.9.5, protocol 22.
The private session is `reverify135`, under `/tmp/hv135-reverify`.
Each TUI uses a scratch home, `HIVE_NO_STARTUP=1`, `--no-startup`, and `COLORTERM=truecolor`.
Demo runs use `--size 120x40` or `--size 80x24` in a 140x48 pane.
Normal runs use a private tmux window at the exact required size.
No test TUI or test prompt uses the real home.

Each initial read is stable across two reads 500 ms apart, without truncation or a scroll offset.
All 16 initial captures have zero differing cells: `verifier2/live-grid-compare.txt`.
The comparison checks all visible rune, token, and weight values, without a mask.
It applies the spec exception for invisible attributes of space cells.
All nine changed boards are captured again as PNG files with fresh Firefox profiles.
Both board PNGs and terminal PNGs remain in `verifier2/boards/` and `verifier2/live/`.
I compared each changed board with its terminal capture.

Both sizes pass the five completion paths, question edits, return to review with Send selected,
ignored digits/clicks/space/tab/text, card switches, dismissal, and ordinary refreshed data.
Long single-line notes retain every word and both sentinels during arrow paging.
The normal-mode record refresh retains the highlighted question and its draft.
The model tests also check the unchanged global quit behavior.

Live multiline captures pass their independent physical-row checks.
Each short, tall, and Other/note case has 11 review captures at each size.
Each review body row has both borders, has no cut, and stays inside the dock cap.
Every tall label line and the final sentinel shows during arrow navigation.
See `verifier2/multiline-check.txt` and `capture-checks.txt`.
The title failure above remains the V3 blocker.

Exact commands and socket requests: `verifier2/live-commands.txt`.
Complete reusable scripts: `verifier2/scripts/`.
Interaction logs: `live-run.txt`, `delivery-run.txt`, `multiline-run.txt`, and `base/live-run.txt` in `verifier2/`.

### Visual table

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header and tabs | status words, `? N`, screen names, active underline | matches | All approved initial captures equal their grids. |
| Chat feed | names, times, `? asked`, open/waits words | matches | Ordinary approved fixtures retain the layout. |
| Message input | `message`, status words, input instruction | matches | Existing input layout stays. |
| Dock frame and count | heavy border, `open`, `I of N` | matches | Approved fixtures retain the dock. |
| Option editor | numbers, `[x]`, `[ ]`, `›` | matches | Picks remain after an edit in approved scenarios. |
| Preview | `preview · N`, framed text | matches | The Preview capture equals its board. |
| Note editor | `note ›`, `█` | matches | The hint says `enter keep note`. |
| Ordinary review title | `confirm` | matches | Approved one-question, multiple-question, and compact grids. |
| Review answers | headers, `→`, full labels and note text | matches | Full picks, Other text, and notes wrap. |
| Selected question | `›` | matches | Arrows select and page a question. |
| Send row | `Send` | matches | It is selected after entry and after an edit. |
| Stopped route | `asker stopped`, parent name, `Send to PARENT` | matches | Route information and actual delivery remain correct. |
| Review hints | `enter send` or `enter change`, arrows, `x drop` | matches | Required review keys. |
| Footer | `[ask]`, `[note]`, `[other]`, `[review]` | matches | Each focus state has a word. |
| Help | `ask review:` and its keys | matches | Approved Help grid. |
| Tall review | `›`, scroll mark `┃`, sentinels | matches | Arrow paging reaches every text segment. |
| Multiline review body | separate text rows, both borders | matches | The wrapper fix retains line breaks and the dock cap. |
| Compact stopped-review title | `confi`, `[review]` in footer | differs | The route note overwrites `rm`. Decision 114 requires full `confirm`. |
| Existing multiline option editor | first and second label lines | differs | Same broken physical rows on current dev. Outside this task's review change. |
| Existing newline-header feed/title | `Head`, `Tail` | differs | Same broken physical rows on current dev. Outside this task's review change. |

Plain-text captures retain the state words and glyphs. No tested state uses colour as its only signal.

## Confirmed pre-existing issues for friction

The manager will record these two issues in friction. This verification changes no source for them.
I checked actual model output and normal-mode live output on current `dev` and the task branch.

Source and build commands:

```sh
mkdir -p .hive/tmp/verifier135b/dev
git archive origin/dev | tar -x -C .hive/tmp/verifier135b/dev
(cd .hive/tmp/verifier135b/dev && go build -o /tmp/hv135-reverify/bin/hive-dev .)
go build -o /tmp/hv135-reverify/bin/hive .
python3 .hive/tmp/verifier135b/base_live.py
```

`verifier2/base/commits.txt` records both source commits.
`verifier2/scripts/base_live.py.txt` preserves the exact accepted requests, setup, keys, binary selection, and capture calls.
`verifier2/live-commands.txt` records each full command and the private socket path.
Each request comes from the executable-checked scratch asker with `/ask FILE`, which calls `hive ask --agent asker-scratch --file FILE`.
Both versions run `--no-startup` against the scratch home. No answer is sent for these comparisons.

1. **Multiline option label in the editor.** The accepted label is `First line\nSecond line`.
   Both versions draw a first physical row without a right border and a continuation without a left border.
   This occurs at 120x40 and 80x24 before review.
   Captures: `verifier2/base/live/{dev,task}-label-{120,80}-editor.{ansi,txt,png}`.
   Input and accepted record: `verifier2/base/label-request.json` and `label-accepted-record.json`.
   The task's subsequent review has intact rows. AC2's review wrapper fix passes.

2. **Newline header in the existing feed and editor title.** The accepted header is `Head\nTail`.
   Both versions split the feed line and card title into unframed physical rows at both sizes.
   Captures: `verifier2/base/live/{dev,task}-header-{120,80}-editor.{ansi,txt,png}`.
   Input and accepted record: `verifier2/base/header-request.json` and `header-accepted-record.json`.
   The new review body handles the multiline header. The pre-existing feed defect remains.

Independent model captures are in `verifier2/base/{dev,task}/`.
Their test commands are:

```sh
VERIFIER_CAPTURE=ABSOLUTE_OUTPUT go test ./internal/tui -count=1 -v -run '^TestVerifierExistingMultilineEditorStates$'
```

The command runs in each isolated source copy. The complete test is in `scripts/additional_test.go.txt`.
`verifier2/base/dev-tests.txt` and `task-tests.txt` preserve the output.
The compact stopped-title comparison is separate from these two pre-existing issues.
Current dev has no review state, so it cannot establish that the new state's overlap is outside scope.

## V4: real TUI delivery through checked scratch stubs

Before the first prompt, both panes show `STUB_READY` with the explicit compiled executable path.
Before each of the six delivery setups, `/proc/PID/exe` equals `/tmp/hv135-reverify/bin/pi`.
`verifier2/stub-executables.txt` records each check and the executable hash.
`verifier2/delivery/*-process.json` and `*-ready.txt` preserve the process and pane evidence.
No installed Pi process runs for these checks.

| Size | Edited answer | Stopped asker | Dismiss |
| --- | --- | --- | --- |
| 120x40 | `a1`, pass | `a2`, pass | `a3`, pass |
| 80x24 | `a4`, pass | `a5`, pass | `a6`, pass |

Before Send, the record remains open with no replies or delivery attempts.
Neither recipient gets a prompt, including after the question edits and the watched-record refresh.
Send delivers the final edited draft once. Literal expected replies:

```json
[{"picks":[2,3]},{"picks":[3],"other":"Next release","note":"keep this"}]
```

Literal expected prompt for `a1`:

```text
⬢ user: answer to ask a1 from asker-scratch · Scope → Help line, Header count; Order → Next release · note: keep this
```

The stopped case reaches only `parent-scratch` and retains `(asker stopped)`.
Dismiss reaches only the asker and retains `No answer comes.`
All six final records have `attempts: 1` and the expected `delivered_to`.
Before/after records, expected prompts, pane output, and TUI captures remain in `verifier2/delivery/`.
`verifier2/scripts/v4.py.txt` gives the exact commands, keys, refresh, and literal assertions.
The compact title failure does not change the passing delivery result.

## V5: scope, gate, and cleanup

`git fetch origin` followed by `git merge --no-edit origin/dev` reports `Already up to date.`
`verifier2/merge.txt` records the commits.
`verifier2/diff-stat.txt` records `git diff dev...HEAD --stat`.
All changes outside the task folder are named in the brief, including tests beside the named source files.
This verification adds only the new report and `verifier2/` evidence.

`verifier2/gate.txt` preserves the full output from the root gate:

| Command | Result |
| --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | No output, exit 0. |
| `go mod tidy -diff` | No output, exit 0. |
| `go vet ./...` | No output, exit 0. |
| `golangci-lint run ./...` | `0 issues.`, exit 0. |
| `go test -race ./...` | All packages pass, exit 0. |

Cleanup is complete before the verdict commit.
`verifier2/cleanup.txt` records the process checks and path checks.

- The private herdr server and all its panes are stopped.
- Both compiled recipient stubs are stopped.
- The private tmux server is absent.
- `/tmp/hv135-reverify` and `.hive/tmp/verifier135b` are absent.
- The worktree `.hive/tmp/` folder is empty.
- Scratch source copies, binaries, exported grids, HTML, homes, and Firefox profiles are removed.
- Final evidence, board captures, and terminal captures remain in this task folder.

The manager must assign the compact review rendering fix, then start a fresh verifier.
