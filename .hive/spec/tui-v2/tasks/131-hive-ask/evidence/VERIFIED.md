# Task 131: independent check

- Verifier: `t131-verifier-2`.
- Date: 2026-10-23.
- Branch: `task/131-hive-ask`.
- Checked commit: `5050e4bb`.
- Base for the red run: `e0799d49`.
- Verdict: **pass**.
- New output: `verifier2/`.

This check replaces the failed check at `53d59516`.
Commit `895d2bd4` fixes Enter from a `note` line before a pick.
The updated brief requires demo fixtures for exact V5 comparisons.
This check uses that brief and the board approval in decision 104.
No project source file changed during this check.

## Verification items

| Item | Result | Output |
| --- | --- | --- |
| V1 Boards | pass | All six approved HTML copies equal their staged copies. Twelve new Firefox captures succeed. Each staged/copy PNG pair has identical bytes. No capture shows `{{`. See `verifier2/boards.log`. |
| V2 Grids | pass | The exporter runs on the full design folder. All 47 grids equal the committed grids with `cellgrid.Compare`. The six approved grids also equal `new-grids/`. The comparison uses no mask. See `verifier2/grids.log`. |
| V3 Tests red and green | pass | The new tests fail on the base. Three packages cannot build without the new APIs. Deliberate breaks make all 46 selected tests fail individually. All 46 pass with `-count=50` and with `-race`. The 46 include all 44 new tests and two existing guards. See the test details below. |
| V4 Live boards and input | pass | All six final exact-size captures have zero differing cells. The keys, raw SGR clicks, and wheel events give the expected changes. A further run checks digits 1 to 5 and Enter between questions. See `verifier2/sized-grids.log`, `keys.log`, `key-*`, and `digit-*`. |
| V5 Visual alignment | pass | The real herdr panes have the exact board sizes: 120×40 and 80×24. Each demo capture equals its grid. Each border, title, mark, preview, and count passes. See `verifier2/visual-points.md` and `sized-exact-*`. |
| V6 End to end | pass | A two-question answer reaches the asker. A stopped asker's answer reaches its parent. A dismiss reaches the asker. Enter from a `note` line with no pick also succeeds. Each prompt equals AC5. See `verifier2/verify-asker-prompts.txt`, `verify-parent-prompts.txt`, and `a*-record.json`. |
| V-gate Scope and gate | pass | `git merge dev` reports `Already up to date`. The diff stays within the source table, the approved boards, the brief copy, and the evidence. The five gate commands exit 0. The first two print nothing. See `verifier2/scope.txt` and the full `gate.log`. |

## V1 and V2: boards and grids

The approved stage is `/home/penguin/source/hive/.hive/tmp/g6/approved131/`.
Decision 104 makes these six boards the authority:

- `ChatAsk`.
- `ChatAskMulti`.
- `ChatAskPreview`.
- `ChatAskStopped`.
- `CompactChatAsk`.
- `Help`.

The approval replaces the original five staged ask boards. It also approves
the changed `Help` board. The six committed grids equal `new-grids/`.
No other grid changes in the task diff.

Commands:

```sh
cmp APPROVED/BOARD.dc.html .hive/spec/tui-v2/design/BOARD.dc.html
firefox --headless --no-remote --profile SCRATCH --window-size=1220,780 --screenshot OUT.png file://BOARD.dc.html
go run ./internal/boardexport/cmd/boardexport -out .hive/tmp/verifier2/export .hive/spec/tui-v2/design
```

Each board capture used a new profile in `.hive/tmp/verifier2/`.
A scratch copy had the committed `support.js` beside it. The HTML bytes did
not change. The verifier inspected each PNG again. The capture hashes are
in `verifier2/boards.log`.

The comparison tool uses `cellgrid.ReadGrid`, `cellgrid.Convert`, and
`cellgrid.Compare`. Its source is `verifier2/gridcmp.go.txt`.
A wrong-grid control gives 1526 differing cells. The correct-grid runs give
zero differing cells. See `verifier2/sized-comparison-control.log`.

## V3: red and green tests

The selected test names and exact commands are in `verifier2/green-count50.log`
and `verifier2/green-race.log`.

```sh
go test -count=1 -run 'SELECTED TESTS' ./cmd/ ./internal/ask/ ./internal/boardexport/ ./internal/tui/ ./internal/tui/component/
go test -count=50 -run 'SELECTED TESTS' ./cmd/ ./internal/ask/ ./internal/boardexport/ ./internal/tui/ ./internal/tui/component/
go test -race -count=1 -v -run 'SELECTED TESTS' ./cmd/ ./internal/ask/ ./internal/boardexport/ ./internal/tui/ ./internal/tui/component/
```

The base run uses a detached scratch worktree at `e0799d49`. It receives the
new tests, fixtures, grids, and spec. Its CLI tests reject the missing command.
Its board test rejects the missing board registrations. The ask, component,
and model tests cannot build because the base lacks the new APIs.
See `verifier2/red-base.log`.

Go overlays then disable the feature on the task code. Each selected test
runs separately. All 46 runs exit 1. None has a build error.
See `verifier2/individual-red.json` and `test-checks.py.txt`.
The two existing guards are:

- `TestHelpMatchesItsBoard`. A disabled renderer makes it fail.
- `TestEachBoardFixtureLoadsAtItsSpecSize`. A disabled fixture loader makes it fail.

The caller guard also fails when an overlay removes it.
See `verifier2/blocked-caller-guard.jsonl`.

The specific fix test also runs against `internal/tui/ask.go` from `53d59516`.
It fails for the original cause:

```text
calls = [{answer a1 [{[]  ok}]}], want [{answer a1 [{[2]  ok}]}]
```

See `verifier2/repair-red.log`. The same test passes on the checked commit,
including the repeated and race runs. No overlay changes the project source.

An additional live CLI check sends all eleven invalid cases from the asker
pane. Each command exits 1, names the error, and creates no record.
See `verifier2/invalid-cli.json`.

## V4 and V5: live screens

The installed herdr version is 0.9.5, protocol 22.
The verifier starts each demo with:

```sh
COLORTERM=truecolor hive --demo FILE --size WxH
herdr pane read PANE --source visible --format ansi
```

Each final capture follows two equal reads, 500 ms apart.
Each read has `truncated: false`. Each pane has scroll offset 0.
The captures use the top-left board cells and no mask.

The final run uses an isolated herdr configuration and a connected frontend.
`stty size` gives `40 120` and `24 80` in the respective panes.

| Board | Exact pane size | Differing cells |
| --- | --- | ---: |
| ChatAsk | 120×40 | 0 |
| ChatAskMulti | 120×40 | 0 |
| ChatAskPreview | 120×40 | 0 |
| ChatAskStopped | 120×40 | 0 |
| CompactChatAsk | 80×24 | 0 |
| Help | 120×40 | 0 |

The same build also gives zero differences in panes of 121×41 and 81×25.
See `verifier2/sized-grids.log`.

The preliminary headless reads omitted styles on some trailing blank cells.
The final `sized-exact-*` captures replace those preliminary runs.
The final correct-grid diffs are empty. The PNG files show the real captured
terminal text after ANSI SGR conversion to HTML.

The `ChatAskMulti` run checks:

- Up, Down, Space, Tab, Enter, `ctrl+]`, `J`, `K`, `i`, and `x`.
- Digits 1 to 5, including Other, in a scratch fixture on the first question.
- Text in a `note` line and in an Other line.
- A raw SGR click on an option, on the feed, and on the card.
- Raw SGR wheel events over the card.
- Page Up and Page Down, sent as terminal sequences.

Each event has its own capture. The feed scrolls while the card keeps the
focus. A feed click gives the feed the focus. A card click restores the card
focus. Enter moves to the next question. Enter on the last question and `x`
show the demo notice, as decision 101 requires.

The first attempt to use the herdr key name `pgup` fails before it sends input.
The completed run sends `ESC [ 5 ~` and `ESC [ 6 ~` instead.
No product code changed for this command correction.

The preview run moves from option 2 to option 3. Its title changes to
`preview · 3`. See `verifier2/preview-down.*`.

`verifier2/visual-points.md` contains the position table and one row for each
board element. It also checks each state after ANSI colour removal.
Each state has a word or a glyph. No state uses colour as its only signal.

## Approved board rule cases

The complete cases remain in `board-rule-cases.md`.
The table below lists all thirteen cases again. Rows and columns are 0-based.
`Wide` means ChatAsk, ChatAskMulti, ChatAskPreview, and ChatAskStopped.
Decision 104 approves each case and the new Help row.

| Case | Board | Rows | Columns | Rule |
| ---: | --- | --- | --- | --- |
| 1 | Wide | 3 | 2–22 | Decisions 26 and 27: the feed title uses `accent` bold. |
| 2 | ChatAsk, ChatAskPreview | 5, 10 | 9–105 | Decisions 15 and 26: user text and the name pad use `user`. |
| 3 | ChatAskMulti, ChatAskStopped | 5, 6 | 9–27 | Decisions 15 and 26: the same user rule as case 2. |
| 4 | Wide | 5, 7, 8, 10, 14 | 9–21 | Decision 26: agent names use `info`. |
| 5 | ChatAsk, ChatAskPreview | 12 | 9–116 | Decisions 15 and 57: a folded event uses one `dim` line. |
| 6 | Wide | 17–27 | 119 | Decision 5: Chat scroll marks use `dim` and `┃`. |
| 7 | CompactChatAsk | 10 | 79 | Decision 5: the same scroll rule as case 6. |
| 8 | Wide, CompactChatAsk | 36, 20 | 2–25 | Decisions 55 and 27: the input title uses `accent` bold. |
| 9 | Wide, CompactChatAsk | 39, 23 | After the hints | Decision 4: footer padding uses `dim`. |
| 10 | CompactChatAsk | 1 | 35–56 | Decision 3: the tab label is `4 Friction`. |
| 11 | CompactChatAsk | 3 | 2–22 | Task 34: the compact feed title is `chat`. |
| 12 | CompactChatAsk | 4–5 | 1–78 | Task 34: the compact feed has no initial blank row. |
| 13 | CompactChatAsk | 5 | 9–17 | Decision 26: the agent name uses `info`. |

The new Help row is row 33, columns 21–95. It lists the ask keys in `dim`.
Its exact-size capture also has zero differing cells.

## V6: scratch delivery

The scratch home is `.hive/tmp/verifier2/home`.
The asker and parent records name only panes in the verifier's scratch session.
The TUI runs with `HIVE_NO_STARTUP=1` and the scratch `HIVE_HOME`.
Its working directory is `/tmp`, outside the chat caller block on `projects/`.

Each external pane runs a prompt sink named `pi`. It records received text
and makes no model request. The hive command, state store, TUI, herdr socket,
and delivery path remain real. This check proves pane delivery, not model
processing of the answer.

The first `hive ask` returns `ask a1 open` in 0.0254 seconds.
The two-question answer reaches the asker exactly as follows:

```text
⬢ user: answer to ask a1 from verify-asker · Scope → Ask card, Help line · note: keep it small; Order → both together
```

The second ask opens while the asker is live. The verifier then stops its
scratch record. The card names the parent. The parent receives exactly:

```text
⬢ user: answer to ask a2 from verify-asker (asker stopped) · Restart → No
```

The verifier makes the scratch asker live again. The third ask closes with `x`.
The asker receives exactly:

```text
⬢ hive: the user dismissed ask a3 from verify-asker · Restart. No answer comes.
```

The fourth ask repeats the previous failure in the real TUI:

1. Press Down with no marked option.
2. Press Tab.
3. Type `ok`.
4. Press Enter.

The asker receives exactly:

```text
⬢ user: answer to ask a4 from verify-asker · Restart → No · note: ok
```

The fourth record is `answered`, with `picks: [2]` and `note: "ok"`.
Each record has one successful delivery. The header changes from `? 0 asks`
to `? 1 ask`, then returns to `? 0 asks` after each action.
The feed keeps the answers and the dismissal.
See `verifier2/e2e-*`, `repair-note-*`, `a*-record.json`, and
`scratch-ask-log.jsonl`.

The failed-delivery test also passes. It checks durable answers, the idle
retry rule, the exact second prompt, and no further delivery after success.
It mocks only the external herdr socket.

## Scope, spec, skills, and gate

`dev` is an ancestor of the checked commit. `git merge dev` changes no file.
The new spec decisions are 91–104. All five role skills give `hive ask`, the
JSON shape, and the AC5 answer prompt. Each skill prohibits its harness's
question tool. The STE check passes without a raised baseline count.

The full output is `verifier2/gate.log`:

```text
gofmt -l $(git ls-files -co --exclude-standard '*.go')  -> no output, exit 0
go mod tidy -diff                                     -> no output, exit 0
go vet ./...                                          -> exit 0
golangci-lint run ./...                                -> 0 issues, exit 0
go test -race ./...                                    -> exit 0
```

The task diff includes the approved Help HTML and grid, the five new boards,
the brief copy, the allowed source files, and task evidence.
The worker's `V<n>.md` files and the plan tracker remain unchanged by this check.

## Cleanup

The verifier stopped each scratch TUI with two Ctrl+C presses.
The verifier stopped and deleted only its two scratch herdr sessions.
The verifier stopped only its two private tmux servers.
The verifier removed the detached base worktree.

The verifier removed `.hive/tmp/verifier2/` and its scratch files.
These files include the home, binaries, overlays, board captures, Firefox
profiles, configuration, and scripts.
The verifier also removed the scratch schema and `/tmp/hv131v2` link.
No scratch session process remains. The main herdr server still runs.
See `verifier2/cleanup.log`.

Final captures, diffs, command output, and script copies remain in the evidence.
