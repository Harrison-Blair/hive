# Rework of task 134 after verification fail

- Date: 2026-10-23.
- Failed verdict: `t134-verifier-1`, manager inbox record 542, commit `b257f299`.
  `../VERIFIED.md` and `../verifier/` stay unchanged.
- Base of this rework: `b257f299`. `origin/dev` is `aaa2e511`, an ancestor of the task branch.
- This rework changes `internal/tui/agents_screen.go` and `internal/tui/agents_hide_test.go` only.
  No board, golden grid, fixture, or spec file changes.

## Cause and fix

The verifier cause was correct:

- `selectedAgent` returned no agent when no pod was visible.
- `updateAgentsKey` sent overview keys to `updatePodsKey` only when a pod existed.

The fix:

- `inPodView` is false when no pod is visible. With zero pods, the screen is the overview, as the view already drew it.
- `updateAgentsKey` sends each overview key to `updatePodsKey`, also with zero pods.
- `selectedAgent` gives the selected coordinator with zero pods.
- `enter` with zero pods and no coordinator selection does nothing. Before, it could set the pod view with no pod.
- New `coordinatorCard` gives the coordinator only when the screen draws its card.
  With no registered project, the screen is the notice alone (decision 10), so up selects nothing.

## R1 Tests red and green

New tests in `internal/tui/agents_hide_test.go`, each at 120x40 and 80x24:

| Test | Checks |
| --- | --- |
| `TestAgentsCoordinatorSelectableWithAllProjectPodsHidden` | Subtests for up and `k`: heavy coordinator border, enter opens `coordinator · esc closes` and `[details]`, esc closes and keeps the selection, `h` shows the pods and keeps the selection. |
| `TestAgentsCoordinatorSelectionSurvivesTheLastPodHiding` | `h` shows the pods, up selects the coordinator, `h` hides the last pod, the border stays heavy, enter opens the details. |
| `TestAgentsNoRegisteredProjectSelectsNothing` | With no registered project and a live coordinator: notice alone, up and enter open no details. |

Red run on `b257f299`, before the fix: `red-before-fix.log`.
The first two tests failed at both widths, with up and with `k`, for the recorded reason:
no heavy border after up or `k`, and no details popup after enter.
`TestAgentsNoRegisteredProjectSelectsNothing` passed on the base. It is a guard.

Guard mutation: `coordinatorCard` returned the coordinator with no registry check.
The guard failed at both widths with the details popup open over the notice: `guard-no-registry.log`.
The source was restored after the run.

The verifier reproduction `../verifier/coordinator-regression.go.txt` ran as a temporary test file after the fix.
Both subtests passed: `verifier-regression-green.log`. The temporary file was removed.

Focused runs of the 27 g9 and new tests, after the fix:

- `go test ./internal/tui/ -run "$R" -count=50 -v`: exit 0, 1350 PASS, 0 FAIL. `green-c50.log`.
- `go test -race ./internal/tui/ -run "$R" -count=50 -v`: exit 0, 1350 PASS, 0 FAIL. `green-r50.log`.

Each existing test passes unchanged. No expectation changed.

## R2 Grids

- The exporter wrote 47 grids. Each is byte-equal to its committed golden grid: `export-compare.log`.
- `go test ./internal/tui/ ./internal/boardexport/ -run 'Board|Golden|Grid' -count=1`: both packages `ok`.
- `git diff b257f299 -- .hive/spec/tui-v2/design` is empty.

## R3 Live checks

The task binary ran from `.hive/tmp/rework134/hive` in scratch herdr tabs of workspace `wWT`,
with `HIVE_NO_STARTUP=1`, `COLORTERM=truecolor`, and a scratch `HIVE_HOME`.
The runs used the verifier fixtures `../verifier/fixtures/coordinator-only-*.fixture.json`.
The no-registry runs used a copy with no projects and only the coordinator and supervisor agents.
Scripts: `live-keys.py.txt`, `live-grids.py.txt`, `cmpgrid.go.txt`.

### Key checks

`live/live-keys.log` records each step. `live/*.txt` and `live/*.ansi` hold each screen.

| Run | Steps | Result at 120x40 and 80x24 |
| --- | --- | --- |
| `coord-up` | start, up, enter, esc, `h`, `h`, enter, esc | up gives the heavy border. enter opens the details. esc closes and keeps the border. `h` shows the pods with the border kept. `h` hides them. enter opens the details again. |
| `coord-k` | start, `k`, enter | `k` gives the heavy border. enter opens the details. |
| `coord-before` | start, `h`, up, `h`, enter | Coordinator selected with pods visible stays heavy when the last pod hides. enter opens the details. |
| `noreg` | start, up, enter | Notice alone, no coordinator card, no details. |

### Grid comparisons

`grids/grid-compare.log`: 20 live captures, each with 0 differences through `cellgrid.Convert` and `cellgrid.Compare`, no mask.
The captures are the nine changed boards and six scenarios against `../new-grids/`,
and the five unchanged references against `design/`.

### Visual table

PNG captures are in `live/*.png`. The board references are the approved grids above.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| All-hidden notice plus coordinator | `no project agents`, `● coordinator-g8` | matches | Decision 110. Unchanged by this fix. |
| Coordinator not selected at start | Light border `│` | matches | Start selection is the pod set. |
| Coordinator selected with zero pods | Heavy border `┃` | matches | AC2 and decision 65. Fixed by this rework. |
| Coordinator details with zero pods | `coordinator-g8 · esc closes`, `[details]` | matches | Decision 65. Fixed by this rework. |
| Details closed | Popup gone, heavy border kept | matches | Same as with pods visible. |
| `h` cycle with the coordinator selected | Pods return, then hide, border stays heavy | matches | AC2 valid selection. |
| No registered project | `no project agents` alone | matches | Decision 10. |
| Approved boards and scenarios | 20 grids | matches | 0 cell differences. |

## R4 Diff scope and gate

`git diff dev --name-only` outside this task folder names the same 23 files as at `b257f299`.
The rework adds changes only to `internal/tui/agents_screen.go` and `internal/tui/agents_hide_test.go`.

`gate.log` has the full gate output, run from the worktree root:

- `gofmt -l`: no output, exit 0.
- `go mod tidy -diff`: no output, exit 0.
- `go vet ./...`: exit 0.
- `golangci-lint run ./...`: `0 issues.`, exit 0.
- `go test -race ./...`: exit 0.

## Cleanup

- The scratch tabs `rework134-live` and `rework134-grids` closed after each script.
- `.hive/tmp/rework134/` held the binary, scratch home, fixtures, exporter output, and compare tool. It is removed.
- The scratch Firefox profiles and HTML files are removed.
