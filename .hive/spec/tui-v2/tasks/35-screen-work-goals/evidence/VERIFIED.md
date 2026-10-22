# Verification: 35-screen-work-goals

Verdict: **fail**. Verifier: `t35-verifier-3`. Date: 2026-10-22.
Code tested: `631e5541`. Evidence below is in `verifier3/`.

The updated brief says that Goals has no `s` key. This verification uses that rule.

| Item | Result | Evidence |
| --- | --- | --- |
| V1 Tests | pass | `base.txt`: all 10 new tests fail against `dev` at `e7cdbd09`. `green.txt` and `restored-green.txt`: all Goals tests pass on the task code. `mutations.txt`: a planned-status token change, removal of `j`, and an added `s` action each fail the corresponding test. |
| V2 Live check | fail | The 120×40 board comparison passes: `live-diff.txt` reports zero differences under the space-cell rule. Live keys and PNG review pass as recorded below. The check stopped at the merge conflict before compact live checks and the remaining keys. No complete V2 pass is claimed. |
| V-gate Diff scope and gate | fail | The five gate commands pass on `631e5541`; full output is in `gate.txt`. Format and tidy print nothing. The required merge of current `dev` at `e7cdbd09` fails in six files. See `merge-conflicts.txt`. No post-merge gate is possible. |

## Blocking finding

`git fetch origin && git merge origin/dev` first produced `631e5541` without a conflict. During verification, task 36 landed on `dev`. The required `git merge dev` then failed in:

- `internal/tui/frame.go`
- `internal/tui/goals_queue.go`
- `internal/tui/goals_queue_test.go`
- `internal/tui/selection_marker_test.go`
- `internal/tui/shell_test.go`
- `testdata/ste-baseline.json`

The verifier ran `git merge --abort`. A repair worker must merge current `dev`, keep both Work screens, and rerun verification. No source repair was made by the verifier.

## Test commands and results

Command on both base and task: `go test ./internal/tui/ -run 'TestWorkGoals|TestGoals' -count=1 -v`.

The detached base worktree had only the new test file and a test-only shim for `goalIDWidth` and `cutCell`. These two helpers are required to compile the tests. Each of the ten new tests fails on behavior, including the first golden test and the no-`s` test. Existing base tests also ran; their passes are not counted as red runs.

Mutation commands used the same package with `-count=1 -v` and these exact `-run` filters:

- `^TestWorkGoalsMatchesItsBoard$`: planned uses dim instead of info; exit 1.
- `^TestGoalsKeysAndHints$`: remove `j` from move-down; exit 1.
- `^TestGoalsHasNoStopsOnlyKey$`: add an `s` action that changes the footer; exit 1.

Each mutation was restored in a `finally` block. The complete Goals suite then passed again. Normal-mode record and refresh tests use scratch homes and start no live coordinator or supervisor.

## Live observations

The verifier built a binary from the tested task code and started only scratch tab `w9N:t2`, pane `w9N:p2`. The scratch home was `.hive/tmp/verify35/home`. The command set `HIVE_HOME` to that folder, `COLORTERM=truecolor`, and `HIVE_NO_STARTUP=1`, then ran `hive --demo ABSOLUTE_WORKGOALS_FIXTURE --size 120x40`.

For each capture, two `pane.read` calls 500 ms apart returned equal ANSI text. Each JSON record shows `truncated: false` and a scroll offset of 0. The verifier cropped the top-left 120×40 cells and used `internal/lib/cellgrid` to compare the capture with the board. There are 2992 strict differences, all token or weight differences in space cells, and zero live-rule differences. `start.ansi`, `start.json`, and `live-diff.txt` hold the results. `capture.py` and `compare.go.txt` preserve the scratch tools.

| Key | Observation |
| --- | --- |
| `s` | Plain text stays identical to the start capture; no stops hint. |
| `up`, `g` | Select `g2-tui-v2`. |
| `j`, `G` | Select `g1-tui-restart`. |
| `tab` | Details gets the focus; footer shows `[details]` and scroll hints. |
| `esc` | List gets the focus; list hints return. |
| `o` | Footer shows `no design for this wave` for the completed goal. |
| `t` on g2 | No change: its open wave has no tasks. The nonempty-task action passes in the model test. |
| `ctrl+c` twice | The scratch program exits. |

## Visual review

The verifier captured the approved absolute-path board again with Firefox and a new scratch profile, and rendered the live ANSI as an HTML preformatted block for a second Firefox PNG. See `board.png` and `start.png`. Font metrics and page padding differ; the cell contents match. The `*.txt` files remove ANSI colour for the state check.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header | `coordinator ● working`, `supervisor ● live`, `needs you 0`, clock | matches | Fixed fixture data. |
| Main tabs | `3 Work`, underline `━` | matches | Active tab has text and an underline. |
| Work views | `[Goals]`, Queue, Tasks | matches | Brackets identify the active view. |
| List panel and header | `goals · 2`, GOAL, STATUS, WAVES | matches | Cell comparison passes. |
| Unselected row | `g2-tui-v2`, `planned`, `0/5` | matches | State uses a word and counts. |
| Selected row | `▶ g1-tui-restart`, `done`, `3/3` | matches | Marker and state word survive colour removal. |
| Details title and fields | `g1-tui-restart`, title, status, project, grant, manager | matches | Labels and values match. |
| Waves table | A, B, C; `1/1`; `✓ done` | matches | Glyph and word show the state. |
| Tasks | task IDs, `merge`, commit IDs | matches | Delivery uses words and IDs. |
| Needs you | `NEEDS YOU`, `nothing` | matches | State uses a word. |
| Action line | `o goal design`, `t open task in Tasks` | matches | Both hints match the board. |
| Footer | `[list]`, move, region, view, open, design, task, help | matches | No `s` hint, as the updated brief requires. |
| Focus after tab | `[details]`, scroll hints | matches | Focus also has a plain-text label. |

## Cleanup and limits

The scratch tab is closed. The detached base worktree, binary, scratch home, browser profiles, generated HTML, and scratch scripts were removed after the evidence tools were copied. `cleanup.txt` records the remaining tabs and worktrees and confirms the scratch folder is absent. Source files have no verifier edits. No coordinator or supervisor was started. Compact live checks and verification after the conflict repair remain required.
