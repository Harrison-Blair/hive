# Task 135 rework 3: stopped state when no route caption fits

Worker: `t135-worker-4`. Date: 2026-10-23.
Finding: `../VERIFIED.md` (fail by `t135-verifier-3`, commit `cef43222`, manager inbox records 574 and 575).
Fix commit: `6e2625a7`. `origin/dev` at `3fbbb17f` is an ancestor of `HEAD` (`merge.txt`).
All paths are relative to this `rework3/` folder.
`../VERIFIED.md`, `../verifier/`, `../verifier2/`, `../verifier3/`, `../rework/`, `../rework2/`, and the `V<n>.md` files are unchanged.

## Cause and fix

`askDockView` in `internal/tui/ask.go` tried two stopped-asker captions on the
top border. When neither fit, it drew no caption. The 32-character asker
`asker-scratch-with-long-name-abc` gives a 49-cell review title, so at 80
columns the card had no stopped-state word.

The fix, in the same function, under the manager fit rule:

1. ` asker stopped · answer goes to PARENT ` (decision 95, unchanged)
2. ` asker stopped · to PARENT ` (decision 114, unchanged)
3. ` asker stopped ` (new card-local fallback)

The caption now also keeps the ` ━` after the title. Before, the caption could
start on the space after the title, so a caption could touch the title with no
border cell between them. With both names at 32 characters at 120 columns, the
full caption then read `confirm  asker stopped · answer goes to …`. With the
gap rule, that state uses caption 2.

At the name limit, caption 3 always fits beside the review title at 80 columns
(3 + 49 + 2 + 15 + 2 = 71). The state never comes from the feed or the header.

Second defect, found with a 32-character parent at 80 columns: the review hint
` enter sends to PARENT · ↑↓ question to change · x drop ` needs 82 cells.
`overlayCells` drops text that does not fit, so the bottom border lost the hint
and the parent. `askCardHint` now takes the width. When the full hint does not
fit, it gives `enter sends to PARENT · ↑↓ change · x drop`. The parent stays in full.

No key and no layout row is added. Decision 114 in `.hive/spec/tui-v2/README.md`
records caption 3, the gap, and the shorter hint.
The prompt, the route, the reply values, and the record format do not change.

## R1 Tests red and green

New file `internal/tui/ask_stopped_test.go`. It loads the accepted request of
`../verifier3/no-caption/request.json` through `ask.Parse`. The asker is
stopped and the parent is live. Thirty later chat messages move the ask event
above the feed viewport; the test fails if `? asked · Change` is visible. Each
check reads only the card rows, from `┏` to `┗`.

`TestAskStoppedReviewKeepsTheStateInTheCard`, at each start size 120x40 and 80x24:

| Case | Asker / parent | 120 caption | 80 caption | 80 hint |
| --- | --- | --- | --- | --- |
| full caption at both widths | `t1` / `p1` | full | full | full |
| short caption at 80 | `asker-scratch` / `parent-scratch` | full | `asker stopped · to parent-scratch` | full |
| no route caption at 80 | 32-char asker / `parent-scratch` | full | `asker stopped` | full |
| short caption at 120, no route caption at 80 | 32-char asker / 32-char parent | `asker stopped · to PARENT` | `asker stopped` | `↑↓ change` form |

Each case checks the full title `┏━ ? ASKER asks · confirm ━`, the literal caption
ending `━┓`, `┃ › Send to PARENT `, and the literal hint. Then it presses `↑`,
`enter`, `2` (edit, return to review), checks `Change → No`, resizes to the
other width, and checks that width. Review sends nothing. `enter` sends
`{Picks:[2]}` to `a1`.

`TestAskStoppedEditorKeepsTheStateInTheCard`: the editor card at 80x24 with the
32-char asker shows `━ asker stopped ━┓`.

| Run | File | Result |
| --- | --- | --- |
| Red, before the fix | `red.txt` | 5 failures. 80x24: `card lacks "━ asker stopped ━┓"` (no stopped word in the card). 120x40 long/long: `card lacks "…confirm ━"` (caption touches the title). Long parent at 80: the bottom border is empty. The full and short caption cases pass (retained guards). |
| Green, with the fix | `green.txt` | All pass |
| Focused, `-count=50` | `green-count50.txt` | Passes |
| Focused, `-race` | `green-race.txt` | 42 top-level tests pass |
| Verifier 3 regression | `verifier3-regression.txt` | `TestVerifierStoppedReviewWhenNeitherCaptionFits` passes at 80 and 120 with `-race`, unchanged. The file was copied in for the run and removed after. |

The red run is the deliberate break for each new guard: the caption 3 fallback,
the ` ━` gap, and the shorter hint each have one failing case before the fix.
No existing test changed.

## R2 Boards and grids

`boards.txt`: the exporter writes 50 grids. Each is byte-equal to its committed
grid. `git diff cef43222 -- .hive/spec/tui-v2/design` is empty. Board, cellgrid,
and TUI tests pass. No board source, golden grid, or fixture changed.
Approval reference: coordinator record 554, package `6c466dc6`.

## R3 Live check

Binary: `go build` of `6e2625a7` in a scratch folder.
Terminal: a private tmux server `-L t135-rework3-live` that I started, status line off, no herdr variables.
Each run: `TERM=xterm-256color COLORTERM=truecolor HIVE_NO_STARTUP=1 HIVE_HOME=SCRATCH/home-CASE hive --no-startup`.
Each scratch home holds `../verifier3/no-caption/asker-record.json` (stopped),
the parent record, `feed-control-accepted-record.json` (ask `a12`), and the 30
messages of `feed-control-later-chat.json`. For `long-both`, the parent name is
`parent-scratch-with-long-name-ab` in the three records.
Keys: none, `1`, `↑`, `enter`, `2`, then a resize to the other width.
I did not press Send. After each run the ask record is still open, with no
replies and no attempts. No prompt went to any pane.

| Capture (`live/`) | Top border | Send row and hint |
| --- | --- | --- |
| `long-asker-80-01-editor` | `asks · Change ━━━ asker stopped ━┓` | `enter review · tab note` |
| `long-asker-80-02-review` | `asks · confirm ━━━ asker stopped ━┓` | `Send to parent-scratch`, full hint |
| `long-asker-80-05-return-review` | same, `Change → No` | same |
| `long-asker-80-06-resized-120` | `asker stopped · answer goes to parent-scratch` | same |
| `long-asker-120-06-resized-80` | `asker stopped` | same |
| `long-both-80-02-review` | `asker stopped` | `Send to parent-scratch-with-long-name-ab`, `enter sends to parent-scratch-with-long-name-ab · ↑↓ change · x drop` |
| `long-both-120-02-review` | `asker stopped · to parent-scratch-with-long-name-ab` | full hint |
| `long-both-80-06-resized-120` / `long-both-120-06-resized-80` | as at the new width | as at the new width |

The ask event is not in the feed viewport in any capture (`↑ 28 earlier`).
PNG files: `live/long-asker-80-01-editor.png`, `long-asker-80-02-review.png`,
`long-asker-80-05-return-review.png`, `long-asker-80-06-resized-120.png`,
`long-both-80-02-review.png`, `long-both-120-02-review.png`.
The footer error `chat send is for the user only` comes from the scratch home
with no coordinator, as in the earlier captures.

| Board element | State without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Review title | words `asks · confirm` | matches | decision 114; full at both widths and both name lengths |
| Stopped-asker caption | words `asker stopped`, with `· to PARENT` or `· answer goes to PARENT` where it fits | overridden | decision 114 addition and the manager fit rule |
| Send row | words `Send to PARENT`, glyph `›` | matches | decision 114 |
| Bottom hint | words `enter sends to PARENT` | overridden | decision 114 addition: `↑↓ change` form only when the full hint does not fit |

## R4 Delivery

The fix changes only the top-border and bottom-border text. The send path,
the route, the replies, and the prompt do not change. R1 checks the reply of the
stopped-asker review after an edit and a resize. The brief repeats a delivery
check only for a behavior change, so I did not repeat V4.

## R5 Scope, gate, cleanup

`diff-stat.txt`: only named files change. This rework changes
`internal/tui/ask.go`, the new `internal/tui/ask_stopped_test.go`,
`.hive/spec/tui-v2/README.md`, and this task folder.
`gate.txt`: the five gate commands. The first two print nothing. Each exits 0.
Cleanup: the tmux server `t135-rework3-live` is stopped. I removed the scratch
binary, homes, exports, HTML, and Firefox profiles.

Not fixed (out of scope): the compact editor hint `enter review · tab note` at
80 columns does not name the parent. This is the approved `CompactChatAsk` hint.
The pre-existing editor/feed defects are also not changed.
