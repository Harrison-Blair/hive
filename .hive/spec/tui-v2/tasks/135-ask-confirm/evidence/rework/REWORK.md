# Task 135 rework: line feeds in the review

Worker: `t135-worker-2`. Date: 2026-10-23.
Finding: `VERIFIED.md` (fail by `t135-verifier-1`, commit `008e9d2d`).
Fix commit: `cc4c170a`. All paths are relative to this `rework/` folder.
The first verdict and `../verifier/` stay unchanged.

## Cause and fix

`styledWords` split only at spaces. A line feed stayed inside one word.
The dock counted one logical row, and the panel drew two physical rows.

The fix is in `internal/tui/ask.go`.
`styledWords` splits the text at each line feed first.
The first word after a line feed has `lineFeed` set.
`wrapAskWords` starts a new row before that word.
Each physical row is now one dock row.
The dock cap, the tall-row paging, and the `›` mark use that count.

Every review field goes through this wrapper: the header, the picked labels,
the Other text, and the `note`. The fix applies to each of them.
The reply values keep each line feed. The prompt and record formats do not change.

## R1 Tests red and green

New file `internal/tui/ask_multiline_test.go`:

- `TestAskReviewKeepsMultilineTextInsideTheDock`: the accepted short request
  of the verifier (`ask.Parse` reads it), with a second question that has a
  two-line Other text and a two-line `note`. At 120x40 and 80x24: each line is
  on its own row, each dock row has both borders, the dock has at most
  height/2 rows, and Send delivers `{Picks:[1]}` and
  `{Picks:[2], Other:"other one\nother two", Note:"note one\nnote two"}`.
- `TestAskReviewReachesEachLineOfATallLabel`: the accepted tall request
  (31 lines). At both widths, after each of 12 up and 12 down presses: the
  screen fits, each dock row has both borders, no dock row has `…`, and the
  dock has at most height/2 rows. `ROW01` to `ROW30` and
  `FINAL-LABEL-SENTINEL` each show. Send delivers `{Picks:[1]}`.

| Run | File | Result |
| --- | --- | --- |
| Red, before the fix | `red.txt` | Both tests fail at both widths: `view does not contain "┃   Second line "` (the line feed broke the row), and `dock row has a broken border: "┃   ROW01 …"` |
| Green, with the fix | `green-count50.txt`, `green-race.txt` | Focused tests pass with `-count=50` and with `-race` |
| Verifier regression | `verifier-regression-green.txt` | `../verifier/multiline-regression.go.txt`, copied in for the run and removed after: passes at both widths |
| Deliberate break | `mutation.txt` | Remove the `lineFeed` break in `wrapAskWords`: the short test fails at both widths. The tall test still passes, because the words then join with spaces and wrap. The short test guards that break. |

Focused command: `go test ./internal/tui -run 'TestAsk|TestHelp|TestDemoFixtureShowsTheReview|TestEachBoardFixtureLoadsAtItsSpecSize'`.
No existing test changed in the rework.

## R2 Boards and grids

`boards.txt`: the exporter writes 50 grids. Each is byte-equal to its committed
grid in `design/`. All board, cellgrid, and TUI tests pass.
`git diff 008e9d2d -- .hive/spec/tui-v2/design` is empty.
No board source, golden grid, or fixture changed.

## R3 Live multiline check

Binary: `go build` of the fix commit, in a scratch folder.
Terminal: a private tmux server `-L t135-rework` that I started, at the exact
size (120x40 and 80x24), with the status line off.
Each run: `TERM=xterm-256color COLORTERM=truecolor HIVE_NO_STARTUP=1
HIVE_HOME=SCRATCH/home hive --no-startup`, with no herdr variables.
The scratch home holds one ask record: `../verifier/multiline/short-accepted-record.json`
or `tall-accepted-record.json`, which `hive ask` accepted in the first verification.
The asker is not live in the scratch home, so the card shows the stopped-asker route.
Keys: `1` (pick and review), then 9 `Up`, then 9 `Down`.
Captures: `tmux capture-pane -p -e` to `live/CASE-W-STEP.ansi` and `.txt`.
PNG files: `live/*-02-review.png`, `live/tall-80-up-0.png`, `live/tall-80-down-8.png`.
I did not press Send, and no prompt went to any pane.

| Case | Size | Result |
| --- | --- | --- |
| short | 120x40, 80x24 | `Label → First line` and `Second line` on two rows, each with both borders |
| tall | 120x40 | Review: 16 body rows, 18 dock rows (half the screen: 20), `ROW18` to `FINAL-LABEL-SENTINEL` and Send. Up shows `› Label → ROW01`. Down pages to Send |
| tall | 80x24 | Review: 8 body rows, 10 dock rows (half the screen: 12), `ROW26` to `FINAL-LABEL-SENTINEL` and Send. Up shows `ROW01` to `ROW08`. Down pages through each row to Send |

No dock row of a review capture has `…`. Each dock row has both borders.

| Board element | State without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Question row with line feeds | `HEADER → LINE`, then one row for each next line | matches | AC2 full text and wrapping; decision 114 row form |
| Selected row | glyph `›` on its first visible row | matches | decision 114, tall-row paging |
| Send row | word `Send to parent-scratch` | matches | decision 114, stopped asker |
| Dock height | at most half the screen | matches | retained dock cap |

### Defects outside the review wrapper (not changed)

These predate task 135 and use other code. The brief does not name them.
I report them to the manager.

1. The option editor: `askOptionLine` puts a multiline label into one row.
   The border breaks, and `…` cuts the label (`live/*-01-editor.txt`).
   The same code is on `dev` (`internal/tui/ask.go` line 620 to 622 there).
2. A header with a line feed breaks the chat feed row `? asked · HEADERS`
   and the card title. I saw this in a draft test. `ask.Parse` accepts it.
3. At 80 columns the stopped-asker note covers the end of the review title:
   `asks · confi asker stopped · …` (`live/short-80-02-review.txt`).
   The approved boards do not show a stopped review at 80 columns.

## R4 Delivery

The fix changes only the review layout. The send path, the replies, and the
prompt do not change. R1 shows the delivered replies with their line feeds.
The brief repeats a delivery check only for a behavior change, so I did not repeat V4.

## R5 Scope, gate, cleanup

`merge.txt`: the branch contains `origin/dev` at `981e21de`.
`diff-stat.txt`: `git diff dev...HEAD --stat` outside the task folder. Only
named files change. The rework adds `internal/tui/ask_multiline_test.go`, a
test beside `ask.go`.
`gate.txt`: the five gate commands. The first two print nothing. Each exits 0.
Cleanup: the tmux server `t135-rework` is stopped. I removed the scratch
binary, home, exports, HTML, and Firefox profiles.
