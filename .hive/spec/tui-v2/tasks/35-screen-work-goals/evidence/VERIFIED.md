# Verification: 35-screen-work-goals

## Land repair verification

Verdict: **pass**. Verifier: `t35-verifier-5`. Date: 2026-10-22.
Tested code: `0b2cff085da589b7309edf9ce59d3f19ac3e7cdf`.
Merged current dev `5a4b29e68930c21359d8db3119b6a765f04514c6`.
A second fetch and merge before the final gate reported `Already up to date.`
Evidence for this run is in `verifier5/`.

The manager approved reuse of V1 and V2 from `d286dcfe` if the deleted
`checkBoard` was a duplicate of the shared helper and a fresh token mutation
failed the WorkGoals golden test. Both conditions pass. The prior report
below remains as evidence of V1, V2, and the visual check. This repair run
did not repeat the live capture or the base and key mutation runs.

| Item | Result | Evidence |
| --- | --- | --- |
| V1 Tests | pass | Prior independent V1 results carry forward under the manager decision. `helper.txt` proves the helpers are identical and WorkGoals calls the shared helper. `token-mutation.txt` fails on ten cells. `green.txt` passes after restoration. |
| V2 Live check | pass | Prior independent V2 and visual results carry forward under the manager decision. No Goals renderer or key change remains in this repair. |
| WorkGoals golden check | pass | Changed the planned status token from info to dim. The exact golden command failed with ten differing cells and exit 1. Restored the token. The Goals suite, including the golden test, passed with exit 0. |
| V-gate Diff scope and gate | pass | `scope.txt` lists only permitted source, brief, baseline, and task evidence. `gate.txt` records all five exact gate commands and their full output. All exit 0. Format and tidy print nothing. No retry was needed. |

Commit `69c8811f` merges dev and resolves the STE baseline. Commit `31764f81`
also deletes the duplicate helper from `work_goals_test.go` and adds gate
evidence. The deleted helper is byte-for-byte identical to the helper in
`friction_test.go` that dev supplied. `TestWorkGoalsMatchesItsBoard` calls
that shared helper. The additional dev merge brings the selected-card board
changes; it does not change the WorkGoals board or renderer.

Fresh commands:

```sh
go test ./internal/tui/ -run '^TestWorkGoalsMatchesItsBoard$' -count=1 -v
go test ./internal/tui/ -run 'TestWorkGoals|TestGoals' -count=1 -v
```

The first command ran with the deliberate token change. The second ran
after restoration. No source mutation remains. This run created no scratch
pane, home, profile, or binary. All new files are verification evidence.
User sign-off remains a later step.

## Prior independent verification

Verdict: **pass**. Verifier: `t35-verifier-4`. Date: 2026-10-22.
Tested code: `d639457bcec3d828f34bb3e09b6965c3725daf6c`.
Evidence paths below are relative to `verifier4/`.

The current brief gives Goals no `s` key. Queue keeps its `s` key.
This report replaces the previous failed verdict after the merge repair.

| Item | Result | Evidence |
| --- | --- | --- |
| V1 Tests | pass | `base.txt`: each of the ten new tests fails on base `13210c3f`. `green.txt`, `restored-green.txt`, and `green-after-merge.txt`: each passes on the task code. `token-mutation.txt` and `key-mutation.txt`: each deliberate change fails its test. |
| V2 Live check | pass | `live-diff.txt`: zero live-rule differences at 120×40. Fresh captures at 119×40, 80×24, and 79×24 check the compact layout, overlay, scrolling, and too-small notice. `keys.txt`, ANSI, plain text, JSON, and PNG files record the checks. |
| V-gate Diff scope and gate | pass | `merge-latest.txt`: clean merge of dev at `fbca94d5`. `scope-after-merge.txt`: only permitted source files, brief, baseline, and evidence. `gate-after-merge.txt`: all five commands exit 0. Format and tidy print nothing. |

## V1: independent runs

Command on base and task:

```sh
go test ./internal/tui/ -run 'TestWorkGoals|TestGoals' -count=1 -v
```

The detached base worktree had the new test file and a test-only shim for
`goalIDWidth` and `cutCell`. These names are necessary to compile the tests.
All ten new tests fail on behavior. Existing passing base tests do not count
as red runs. The golden test fails on the old screen layout.

The token mutation changes `planned` from `infoStyle` to `dimStyle`.
`go test ./internal/tui/ -run '^TestWorkGoalsMatchesItsBoard$' -count=1 -v`
fails with ten differing cells. The key mutation removes `j` from move down.
`go test ./internal/tui/ -run '^TestGoalsKeysAndHints$' -count=1 -v` fails:
`key "j" selected "g2-two"; want "g3-three"`.
Both mutations were restored. The full Goals suite passes again. No source
mutation remains.

The normal-mode tests use scratch homes. They check record loading, refresh,
selection, design opening, task opening, and the program on a string reader.
The responsive tests check empty data and long text at 79×24, 80×24, 80×25,
119×40, 120×40, 120×24, and 120×25. They check line widths, whole borders,
and visible selection before and after `enter`.

## V2: live run

The verifier built the task binary again after the latest dev merge.
Only scratch tab `wA2:t2`, pane `wA2:p2`, was used. The pane had 57 rows and
was wider than 120 columns. The command used a scratch `HIVE_HOME`,
`HIVE_NO_STARTUP=1`, and `COLORTERM=truecolor`:

```sh
hive --demo ABSOLUTE_WORKGOALS_FIXTURE --size 120x40
```

The same command ran at 119×40, 80×24, and 79×24. `live.py` records the
commands and keys. `capture.py` records the capture method. Each capture
uses two equal reads 500 ms apart, `source: visible`, and `format: ansi`.
Each JSON record has `truncated: false` and scroll offset 0. Each ANSI file
contains the top-left W×H cells. `compare.go.txt` records the converter call.

The 120×40 comparison has 2992 strict differences, all token or weight
changes on space cells. It has **zero differences under the space-cell rule**.
The model golden test compares space cells fully and passes.

| Keys or size | Observation |
| --- | --- |
| `up`, `k`, `g` | Select `g2-tui-v2`. |
| `down`, `j`, `G` | Select `g1-tui-restart`. |
| `s` | No view change at 120, 119, or 80 columns. No Goals stops hint. |
| `tab`, `esc` | Details gains focus; `esc` returns to the list. Footer names the focus and its keys. |
| `o` | Footer says `no design for this wave`. The fixture has no available design. |
| `t` | No task opens for the fixture goal with no open wave. The normal-mode test checks opening a task from a nonempty wave. |
| `]`, `[` | Queue opens with its stops hint; Goals returns with its design and task hints. |
| `?`, `esc`, `R` | Help opens and closes; refresh keeps the Goals view. |
| 119×40 | Full-width list; `enter` opens full-width details; `esc` returns. Selected goal stays visible. |
| 80×24 | Same overlay flow. Details shows `↓ 6 more lines`; `down` changes it to `↓ 5 more lines`. Borders stay whole. |
| 79×24 | Too-small notice names 79×24 and the minimum 80×24. Header and quit hint remain. |
| `ctrl+c` twice | Each scratch run exits to its shell. |

The live driver could not send page keys: herdr rejects `pgdn`, `pagedown`,
`PageDown`, and `page_down`. No live page-key result is claimed. The live
scroll check uses `down`.

## Visual check

The verifier captured the approved board again from
`/home/penguin/source/hive/.hive/spec/tui-v2/design/WorkGoals.dc.html` with
Firefox, a new scratch profile, and a 1220×780 window. The approved HTML and
grid agree byte for byte with the worktree copies. `board.png` is that
capture. Each live PNG renders the captured ANSI as an HTML `pre`.
The board and terminal PNGs use different font metrics and padding.
Cell positions and contents agree. Plain-text files remove colour.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header | `coordinator ● working`, `supervisor ● live`, needs-you count, clock | matches | Fixed fixture data. |
| Main tabs | `3 Work`, `━` underline | matches | Text and underline identify the active tab. |
| Work views | `[Goals]`, Queue, Tasks | matches | Brackets identify the view. |
| List title and columns | `goals · 2`, GOAL, STATUS, WAVES | matches | Cell comparison passes. |
| Unselected goal | `g2-tui-v2`, `planned`, `0/5` | matches | Status is a word. |
| Selected goal | `▶ g1-tui-restart`, `done`, `3/3` | matches | Marker, word, and count remain without colour. |
| Details title | `g1-tui-restart` | matches | Selected goal ID. |
| Details fields | title, status, project, grant, manager | matches | Labels and values agree. |
| Waves table | A, B, C; `1/1`; `✓ done` | matches | Glyph, word, and count show each state. |
| Tasks | task IDs, `merge`, commit IDs | matches | Delivery uses words and IDs. |
| Needs you | `NEEDS YOU`, `nothing` | matches | State has a word. |
| Action line | `o goal design`, `t open task in Tasks` | matches | Hints agree. |
| Footer | `[list]`, movement, region, view, open, design, task, help | matches | No `s` hint, as the current brief requires. |
| Details focus | `[details]`, scroll and exit hints | matches | Focus has a text label. |
| Compact list and overlay | `▶`, selected ID, `esc back`, `↓ N more lines` | matches | Responsive form follows decisions 9 and 46. |
| Too-small view | `terminal too small`, dimensions, quit hint | matches | Decision 7. |

## Merge, gate, and cleanup

The initial `git merge dev` was already up to date. The later merge of
`origin/dev` at `fbca94d5` made `d639457b` without a conflict. Both Goals and
Queue registry entries remain. The merged Logs screen remains. A final
fetch and merge before the verdict was already up to date.

Full gate output is in `gate-after-merge.txt`. All five commands pass:
`gofmt -l`, `go mod tidy -diff`, `go vet ./...`, `golangci-lint run ./...`,
and `go test -race ./...`. No retry was necessary.

The scratch program exited. The owned tab was closed. The detached base
worktree, binary, scratch home, browser profiles, generated HTML, and scratch
tools were removed after their evidence was saved. `cleanup.txt` records
the remaining tabs and worktrees. `.hive/tmp/verify35v4` is absent. No live
coordinator or supervisor was started. No verifier source edit remains.
User sign-off under decision 47 remains a later step.
