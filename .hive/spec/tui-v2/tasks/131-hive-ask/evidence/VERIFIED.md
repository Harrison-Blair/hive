# Task 131: independent check

- Verifier: `t131-verifier-1`.
- Date: 2026-10-23.
- Branch: `task/131-hive-ask`.
- Checked commit: `b327f6b5`.
- Base for the red run: `e0799d49`.
- Verdict: **fail**. F1 breaks AC5. V5 also lacks an equal real-TUI capture.
- All new output is in `verifier/`. No project source file changed during this check.

## Verification items

| Item | Result | Output and difference |
| --- | --- | --- |
| V1 Boards | pass | All six approved HTML copies equal the staged copies. Twelve Firefox captures exit 0. Each staged/copy PNG pair has identical bytes. No image shows `{{`. See `boards.log`, `board-captures.log`, and `board-capture-hashes.log`. |
| V2 Grids | pass | The exporter runs on the full design folder. `cellgrid.Compare`, with no mask, finds zero differences for every committed grid. All six approved grids also equal `new-grids/`. See `grids.log` and `approved-grids.log`. |
| V3 Tests red and green | pass | The tests fail on the base. Missing new APIs prevent three packages from building there. Independent feature breaks make each of the 45 selected tests fail individually, without a build error. The task code passes `-count=50` and `-race`. See the test details below. |
| V4 Live boards and input | fail | All six demo captures equal their grids. The recorded keys and raw SGR mouse events give the expected changes. An additional live `note` check fails: Enter submits no option. See F1, `demo-grids.log`, and `keys-summary.log`. |
| V5 Visual alignment | fail | Real TUI captures are exactly 120×40 and 80×24. Their borders and measured columns pass. They differ from the board grids in 1264 and 637 cells. The scratch records differ from the board data. These captures cannot establish the required zero-difference result. See the V5 table below. |
| V6 End to end | pass | A two-question answer reaches the asker. An answer for a stopped asker reaches its parent. A dismissal reaches the asker. Each prompt equals AC5. The header count rises and falls. See the V6 details below. |
| V-gate Scope and gate | pass | `dev` is an ancestor of the checked commit. `git merge dev` reports no merge work. The five gate commands exit 0. Formatting and module checks print nothing. See `scope.txt` and the full `gate.log`. |

## F1: Enter from a `note` line does not pick the current option

Location: `internal/tui/ask.go`, `updateAskKey`, the editing branch for Enter.

Reproduction in the real scratch TUI:

1. Create an ask with one question and two options: `Yes` and `No`.
2. Press Down. The current option becomes `No`.
3. Press Tab before an option has a mark.
4. Type `ok`.
5. Press Enter.

Expected: the asker gets this prompt, and the card closes:

```text
⬢ user: answer to ask a4 from verify-asker · Restart → No · note: ok
```

Actual: the footer shows `question 1: pick an option`. The record stays `open`.
The asker gets no prompt. The `note` text remains on the card.

The editing branch calls `askEnter` without marking the current option.
The normal Enter branch marks that option first. The store then rejects the empty picks.

Independent model output:

```text
note enter: got [{answer a1 [{[]  ok}]}], want [{answer a1 [{[2]  ok}]}]
```

Evidence:

- `verifier/note-enter_test.go.txt`: the independent test. It uses the existing model helpers.
- `verifier/note-enter.log`: the failing test output.
- `verifier/note-before-enter.ansi` and `.png`: the open editing line.
- `verifier/note-after-enter.ansi` and `.png`: the error after Enter.
- `verifier/a4-record.json`: the record remains open.
- `verifier/note-asker.txt`: no fourth answer reaches the asker.

This check uses a Go overlay. It adds no test to the project source.

## V1 and V2: commands and approval

The approved stage is `/home/penguin/source/hive/.hive/tmp/g6/approved131/`.
Decision 104 replaces the original staged boards with these six approved boards.
The six names are `ChatAsk`, `ChatAskMulti`, `ChatAskPreview`, `ChatAskStopped`, `CompactChatAsk`, and `Help`.

Commands:

```sh
cmp /home/penguin/source/hive/.hive/tmp/g6/approved131/BOARD.dc.html .hive/spec/tui-v2/design/BOARD.dc.html
firefox --headless --no-remote --profile SCRATCH --window-size=1220,780 --screenshot OUT.png file://BOARD.dc.html
go run ./internal/boardexport/cmd/boardexport -out .hive/tmp/verifier131/export .hive/spec/tui-v2/design
```

The exporter output folder must exist. The first attempt omitted this folder and failed.
After folder creation, the same export command exits 0. `cellgrid.Compare` compares each output with its committed grid.

Each staged HTML capture uses a scratch copy with the committed `support.js` beside it.
The HTML bytes remain unchanged. Board captures and Firefox profiles were removed after the check.

The changed `Help.dc.html` is part of the approval in decision 104.
The remaining diff stays within the brief's source table, the brief copy, and task evidence.

## V3: test details

Commands use the 45 test names recorded in `verifier/individual-red.json`:

```sh
go test -count=1 -run 'SELECTED TESTS' ./cmd/ ./internal/ask/ ./internal/boardexport/ ./internal/tui/ ./internal/tui/component/
go test -count=50 -run 'SELECTED TESTS' ./cmd/ ./internal/ask/ ./internal/boardexport/ ./internal/tui/ ./internal/tui/component/
go test -race -count=1 -v -run 'SELECTED TESTS' ./cmd/ ./internal/ask/ ./internal/boardexport/ ./internal/tui/ ./internal/tui/component/
go test -count=50 -race ./internal/tui -run '^TestEachBoardFixtureLoadsAtItsSpecSize$'
```

The base run uses a separate worktree with the new tests, new fixtures, grids, and spec.
Its CLI tests fail for missing command behavior. Its board registration test rejects the new boards.
The ask, model, and component tests cannot build because the base lacks the new APIs.
See `verifier/red-base.log`.

To check behavioral failure independently, Go overlays disable the new behavior on the task code.
Each selected test runs separately. Every run exits 1. None has a build error.
Some rendering checks panic after the disabled renderer returns no rows. Their output is recorded without alteration.

The overlays disable ask creation, prompt text, routing, rendering, card actions, header counting, refresh paths, and delivery ticks.
They also remove the new board registration and styled-line behavior.
See `verifier/mutation-checks.py`, `mutations.json`, and `individual-red.json`.

Existing regression guards also fail under deliberate breaks:

- `TestHelpMatchesItsBoard`: the disabled renderer removes the Help content.
- `TestEachBoardFixtureLoadsAtItsSpecSize`: the disabled fixture loader refuses the fixtures.
- `TestABlockedCallerCannotAnswer`: removal of the caller guard permits an answer and a dismissal.

The task code remains unchanged. Green output is in `green-count50.log`, `green-race.log`, and `green-fixtures.log`.
The new independent F1 test remains red on the task code.

## V4: live demo checks

The verifier created herdr session `t131-verifier`, workspace `w1`, and pane `w1:p1`.
The installed herdr version is 0.9.5, protocol 22.
No input went to a pre-existing pane.

```sh
COLORTERM=truecolor hive --demo FILE --size WxH
herdr --session t131-verifier pane read w1:p1 --source visible --format ansi
```

Each capture follows two equal reads, 500 ms apart. Each read has `truncated: false`.
The pane has scroll offset 0. Comparison crops the top-left board cells and uses no mask.

| Board | Size | Differing cells |
| --- | --- | ---: |
| ChatAsk | 120×40 | 0 |
| ChatAskMulti | 120×40 | 0 |
| ChatAskPreview | 120×40 | 0 |
| ChatAskStopped | 120×40 | 0 |
| CompactChatAsk | 80×24 | 0 |
| Help | 120×40 | 0 |

The `ChatAskMulti` run checks arrows, Space, Enter, Tab, `ctrl+]`, `J`, `K`, `i`, `x`, and the Other digit.
It checks text entry in both editing lines. It sends raw SGR clicks to an option, the feed, and the card.
It sends raw wheel, Page Up, and Page Down sequences. Each event has its own capture in `key-*.ansi`.
The summary is `keys-summary.log`. Send and dismiss show the demo notice, as decision 101 requires.

PNG files contain a rendering of each terminal capture. A scratch script converts ANSI SGR to HTML before Firefox captures it.

## Visual elements and states without colour

This table compares the live demo captures with the approved boards.
The PNG files were inspected again. Each state remains visible after ANSI colour removal.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| ChatAsk header | `? 1 ask` | matches | Zero differing cells. |
| ChatAsk feed | `? asked`, `open below` | matches | Zero differing cells. |
| ChatAsk dock and title | heavy border, `? manager-hive asks`, `open`, `1 of 1` | matches | Dock sits between feed and input. |
| ChatAsk options | `›`, `[ ]`, `(REC)`, `Other` | matches | Marks and cursor have text glyphs. |
| ChatAsk input and footer | `message`, `[ask]` | matches | Focus has an explicit tag. |
| ChatAskMulti header and feed | `? 2 asks`, `waits`, `2 of 2` | matches | Both open asks have text states. |
| ChatAskMulti dock title | `Scope`, `Order`, `1 of 2`, `J next` | matches | Zero differing cells. |
| ChatAskMulti options | `[x]`, `[ ]`, `›` | matches | Selection remains visible without colour. |
| ChatAskMulti editing line | `note ›`, `█`, `[note]` | matches | Text identifies the editing state. |
| ChatAskMulti input | `message` | matches | Dock stays above input. |
| ChatAskPreview header and feed | `? 1 ask`, `open below` | matches | Zero differing cells. |
| ChatAskPreview dock and options | `Layout`, `›`, `[ ]` | matches | Zero differing cells. |
| ChatAskPreview preview box | box border, `preview · 2` | matches | Preview starts at column 44. |
| ChatAskPreview input and footer | `message`, `[ask]` | matches | Explicit focus tag. |
| ChatAskStopped header and history | `? 1 ask`, `answered`, `→` | matches | Answer history remains in the feed. |
| ChatAskStopped routing title | `asker stopped`, `answer goes to manager-hive` | matches | Routing has plain text. |
| ChatAskStopped options and footer | `[ ]`, `›`, `send to parent`, `[ask]` | matches | Zero differing cells. |
| CompactChatAsk header | `? 1` | matches | Approved compact form. |
| CompactChatAsk feed | `? asked`, `below` | matches | Approved compact form. |
| CompactChatAsk dock | heavy border, `Model`, `1 of 1` | matches | Full-width dock. |
| CompactChatAsk options | `[ ]`, `›`, `(REC)`, `Other` | matches | Zero differing cells. |
| CompactChatAsk input and footer | `message`, `[ask]` | matches | Explicit focus tag. |
| Help overlay and title | `keys`, `? or esc closes`, `[help]` | matches | Zero differing cells. |
| Help ask-key line | `ask card`, arrows, `space mark`, `tab note`, `x dismiss`, `J/K card` | matches | Approved Help row. |
| Original board palette and rule slips | text and state glyphs remain | overridden | Decision 104 approves all 13 cases in `board-rule-cases.md`. |

## V5: real TUI at both exact sizes

The real TUI uses the scratch home, with `HIVE_NO_STARTUP=1` and no demo loader.
Its working directory is `/tmp`. A working directory below the real home's `projects/` correctly blocks user input.
The scratch agent records route only to panes of the scratch session.

A private tmux server hosts the herdr terminal program.
Terminal sizes 154×41 and 114×25 give the herdr pane exactly 120×40 and 80×24.
Captures contain 40 and 24 rows. Their maximum row widths are 120 and 80 cells.

Columns below are zero-based. Rows are one-based.

| Size | Point | Result | Observation |
| --- | --- | --- | --- |
| 120×40 | borders and corners | pass | Left column 0 and right column 119 align. |
| 120×40 | dock between feed and input | pass | Dock rows 23–36. Input starts at row 37. |
| 120×40 | title, option mark, cursor | pass | Title column 3, mark column 6, cursor column 2. |
| 120×40 | header count | pass | `? 1 ask` starts at column 90. |
| 120×40 | preview | fail | The real two-question ask has no preview. Only the demo checks this point. |
| 120×40 | exact board comparison | fail | 1264 cells differ from `ChatAskMulti.grid`. Live data and card content differ. |
| 80×24 | borders and corners | pass | Left column 0 and right column 79 align. |
| 80×24 | dock between feed and input | pass | Dock rows 11–20. Input starts at row 21. |
| 80×24 | title, option mark, cursor | pass | Title column 3, mark column 6, cursor column 2. |
| 80×24 | header count | pass | `? 1` starts at column 54. |
| 80×24 | preview | fail | The real ask has no preview. |
| 80×24 | exact board comparison | fail | 637 cells differ from `CompactChatAsk.grid`. Live data and card content differ. |

Captures, PNG files, and full diffs are `verifier/real-120x40-open.*` and `real-80x24-open.*`.
These differences do not establish a new rendering defect. They prevent a pass under the brief's exact V5 requirement.
A demo match alone does not satisfy that requirement.

## V6: scratch delivery

The scratch home is `.hive/tmp/verifier131/home`.
The asker pane is `w1:p2`. The parent pane is `w1:p3`.
Each pane runs an external prompt sink named `pi`. It prints each received line and makes no model request.
Herdr detects each sink as idle. The hive command, state store, TUI, socket delivery, and parent routing remain real.

The first ask runs from the asker shell. It returns `ask a1 open` in 0.012 seconds.
Later asks also run from that shell. An initial attempt from the verifier pane correctly fails the identity check.

The two-question answer selects all three Scope options. It includes a `note` and an Other answer for Order.
The asker gets exactly:

```text
⬢ user: answer to ask a1 from verify-asker · Scope → Ask card, Header count, Help line · note: no help yet; Order → both together
```

After the asker record stops, the parent gets exactly:

```text
⬢ user: answer to ask a2 from verify-asker (asker stopped) · Restart → No
```

After the asker record becomes live again, `x` sends exactly:

```text
⬢ hive: the user dismissed ask a3 from verify-asker · Restart. No answer comes.
```

Each record has one successful delivery. The header returns from `? 1 ask` to `? 0 asks` after each action.
See `a1-asker.txt`, `a2-parent.txt`, `a3-asker.txt`, `a1-record.json` through `a3-record.json`, and `e2e-*.ansi`.
The state-transition output is in `scratch-ask-log.jsonl`.

## Cleanup

- The verifier stopped the real scratch TUI with two Ctrl+C presses.
- The verifier stopped and deleted only session `t131-verifier`.
- The verifier removed only its private tmux server, `t131verify`.
- The verifier removed the base worktree and `.hive/tmp/verifier131/`.
- The verifier removed `.hive/tmp/verifier131-schema.json`.
- No matching scratch process remains. The main herdr server still runs.
- All scratch board PNG files and Firefox profiles are removed. Final screen captures and test output remain in evidence.
- The worker's `V<n>.md` files and the plan tracker remain unchanged.
