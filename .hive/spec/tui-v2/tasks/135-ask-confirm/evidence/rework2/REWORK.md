# Task 135 rework 2: compact stopped-review title

Worker: `t135-worker`. Date: 2026-10-23.
Finding: `../VERIFIED.md` (fail by `t135-verifier-2`, commit `235892c0`, manager inbox record 565).
Fix commit: `0ae08cf9`. Merge of `origin/dev` at `3fbbb17f`: `312fa5e1`.
All paths are relative to this `rework2/` folder.
`../VERIFIED.md`, `../verifier/`, `../verifier2/`, and `../rework/` are unchanged.

## Cause and fix

At 80 columns, the review title `? asker-scratch asks · confirm` (30 cells) and
the note `asker stopped · answer goes to parent-scratch` (47 cells with its
spaces) cannot both fit in the 80-cell top border. `askDockView` drew the note
at the right edge without a check, so the note covered `rm` of `confirm`.

The fix is in `askDockView` in `internal/tui/ask.go`. For a stopped asker, the
dock uses the first note that fits after the title:

1. ` asker stopped · answer goes to PARENT ` (decision 95, unchanged)
2. ` asker stopped · to PARENT `

The note can start on the space that follows the title. This is the same
position as `dev`, so each state where the full note fit on `dev` does not
change. If neither note fits, the dock draws no note. The `Send to PARENT` row
and the `enter sends to PARENT` hint still name the route.

Decision 114 in `.hive/spec/tui-v2/README.md` records the shorter text.
The prompt, the route, the reply values, and the record format do not change.
The multiline wrapper fix of `cc4c170a` does not change.

Side effect: at 80 columns, the stopped-asker editor with the long title
`? asker-scratch asks · Scope · Order` had the same overlap on `dev`
(`../verifier2/delivery/80-stopped-01-question.txt`). It now shows the full
title with the shorter note (`live/stopped-80-03-second-editor.txt`).
No approved grid shows that state.

### Decision for the manager: the verifier regression literal

`../verifier2/scripts/additional_test.go.txt` asserts the full title and also
the literal `asker stopped · answer goes to parent-scratch` at 80x24.
Both cannot fit on one 80-cell border row. My fix keeps the full title and
shows `asker stopped · to parent-scratch`. Thus the verifier regression now
fails only on the route literal (`verifier2-regression.txt`).
If the full route text must show at 80 columns, it needs a new row in the
dock body. That is a layout change, so I did not make it.

## R1 Tests red and green

New test `TestAskCompactReviewForAStoppedAskerKeepsTheTitle` in
`internal/tui/ask_review_test.go`. At 80x24, with `asker-scratch` stopped and
`parent-scratch` live, the pick of `1` opens the review. It expects the literal
top border
`┏━ ? asker-scratch asks · confirm ━━━━━━━━━ asker stopped · to parent-scratch ━┓`,
the row `┃ › Send to parent-scratch `, and the hint
`┗━ enter sends to parent-scratch · ↑↓ question to change · x drop ━`.
Review writes no answer. `enter` sends `{Picks:[1]}` to `a1`.

| Run | File | Result |
| --- | --- | --- |
| Red, before the fix | `red.txt` | Fails: the border is `┏━ ? asker-scratch asks · confi asker stopped · answer goes to parent-scratch ━┓` |
| Green, with the fix | `green.txt` | Passes |
| Focused, `-count=50` | `green-count50.txt` | `go test ./internal/tui -count=50 -run TestAsk` passes |
| Focused, `-race` | `green-race.txt` | Passes |
| Verifier regressions | `verifier2-regression.txt` | `TestVerifierMultilineReviewFieldsAndReplies` passes. `TestVerifierCompactStoppedReviewTitle` finds the full title and fails only on the route literal (see above). The file was copied in for the run and removed after. |

The red run is the deliberate break: the pre-fix code always draws the full note.
The existing test `TestAskReviewForAStoppedAsker` (120x40, full note) and
`ask_test.go` stopped-editor checks pass unchanged. No existing test changed.

## R2 Boards and grids

`boards.txt`: the exporter writes 50 grids. Each is byte-equal to its committed
grid. Board, cellgrid, and TUI tests pass. `git diff 235892c0 -- .hive/spec/tui-v2/design`
is empty. No board source, golden grid, or fixture changed.

## R3 Live check

Binary: `go build` of the fix, in a scratch folder.
Terminal: a private tmux server `-L t135-rework3` that I started, at 120x40 and
80x24, with the status line off, and no herdr variables.
Each run: `TERM=xterm-256color COLORTERM=truecolor HIVE_NO_STARTUP=1 HIVE_HOME=SCRATCH/home hive --no-startup`.
The scratch home holds two open asks from `asker-scratch` with parent
`parent-scratch`: `a11` (`../verifier2/base/stopped-accepted-record.json`) and
`a12` (the questions of `../verifier2/delivery/request.json`). No agent is live,
so both cards show the stopped-asker route.
I did not press Send. No prompt went to any pane.

| Capture | Keys | Result at 80x24 | Result at 120x40 |
| --- | --- | --- | --- |
| `stopped-W-01-editor` | none | `? asker-scratch asks · Model asker stopped · answer goes to parent-scratch` (same as `dev`) | full title, full note |
| `stopped-W-02-review` | `1` | `? asker-scratch asks · confirm ━━━━━━━━━ asker stopped · to parent-scratch` | full title, full note (same as before) |
| `stopped-W-03-second-editor` | `J` | `? asker-scratch asks · Scope · Order ━━━ asker stopped · to parent-scratch` | full title, full note |
| `scope-W-04-review` | `J 1 2 enter 1` | full title; `Scope → Ask card, Help line`, `Order → Card first`, `› Send to parent-scratch` | same rows |
| `scope-W-05-ignored-keys` | `2 3 space tab a b c` | no change to picks | no change |
| `scope-W-06-highlight-scope` | `↑ ↑` | `› Scope → …`, hint `enter change` | same |
| `scope-W-07-edit-scope` | `enter` | Scope editor, `[x]` on Ask card and Help line | same |
| `scope-W-08-return-review` | `3 enter` | review with `Scope → Ask card, Help line, Header count`, `› Send to parent-scratch` | same |

PNG files: `live/stopped-80-02-review.png`, `live/stopped-120-02-review.png`,
`live/scope-80-08-return-review.png`, `live/stopped-80-03-second-editor.png`.
The footer error `chat send is for the user only` comes from the scratch home
with no coordinator. It is the same in the earlier rework captures.

| Board element | State without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Review title | words `asks · confirm` | matches | decision 114; full at both widths |
| Stopped-asker note | words `asker stopped · to PARENT` (80) or `asker stopped · answer goes to PARENT` (120) | overridden | decision 114 addition: the shorter text when the full note would cover the title |
| Send row | words `Send to parent-scratch`, glyph `›` | matches | decision 114 |
| Bottom hint | words `enter sends to parent-scratch` | matches | decision 117 |

## R4 Delivery

The fix changes only the top-border text. The send path, the route, the replies,
and the prompt do not change. R1 checks the reply of the stopped-asker review.
The brief repeats a delivery check only for a behavior change, so I did not repeat V4.

## R5 Scope, gate, cleanup

`merge.txt`: `origin/dev` at `3fbbb17f` is an ancestor of `HEAD`. The merge adds
only the two friction entries.
`diff-stat.txt`: `git diff dev...HEAD --stat` outside the task folder. Only named
files change. This rework changes `internal/tui/ask.go`,
`internal/tui/ask_review_test.go`, and `.hive/spec/tui-v2/README.md`.
`gate.txt`: the five gate commands. The first two print nothing. Each exits 0.
Cleanup: the tmux server `t135-rework3` is stopped. I removed the scratch
binary, home, exports, HTML, and Firefox profiles.
