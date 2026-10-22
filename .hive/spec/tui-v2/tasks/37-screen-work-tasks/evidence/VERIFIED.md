# VERIFIED 37-screen-work-tasks

Result: fail

Verifier: `t37-verifier-2`. Date: 2026-10-22.
Reviewed task commit: `17367658e593e3925e97c70fe9db785616822e7c`.

## Verdict

The duplicate-helper repair passes its tests. V-gate fails: the required merge of current dev (`f1847c7bf455dc4ec9a1841e5973a66e07402df6`) has conflicts in `internal/tui/frame.go`, `internal/tui/goals_queue.go`, and `testdata/ste-baseline.json`. A worker must repair this integration. I saved the conflicts and gate output, then aborted the merge. I did not resolve source conflicts. There was one merge attempt, at the start of V-gate.

| Item | Result | Evidence |
| --- | --- | --- |
| V1 | pass | All 12 WorkTasks tests fail on base and pass on task. Goals project-list and both component tests also fail on base and pass on task. Token and key mutations fail tests. See `verify2/`. |
| V2 | pass | New Work and Compact captures each have 0 differing cells under the live space-cell rule. Work keys open the fixture files, dependencies, and details. Both board and screen PNG files were captured and reviewed again. |
| V-gate | fail | Required dev merge conflicts in three files. Full gate run on the conflicted tree fails on conflict markers. `go mod tidy -diff` prints nothing and exits 0. See `verify2/merge-conflicts.txt` and `verify2/gate.log`. |

## Repair review

I reviewed the changes after `c42ed6d7`. Merge `838ef5e5` imports dev's selected-card board changes. Repair `ef06d9fe` removes duplicate `checkBoard` and `requireFits` helpers and uses the shared helpers from `friction_test.go`. The shared fit helper checks size and whole borders and returns the uncoloured view. Its callers preserve their content assertions. The repair also removes duplicate `clipCell` in favour of dev's helper. Commit `17367658` adds worker evidence. These changes are acceptable within the manager's repair scope.

## V1 runs

Base: `866bde2a`, extracted into a scratch folder. I copied the pre-repair task test file with its local helpers into the base, plus the updated normal-mode tests. This avoids a dependency on helpers introduced by dev's Friction screen.

```sh
go test ./internal/tui -run TestWorkTasks -v
go test ./internal/tui -run TestGoalsOpenTasksListsEachProject -v
go test ./internal/tui/component -run 'TestListFillGolden|TestSubTabsHint' -v
```

Each of the 12 WorkTasks test functions fails on the base. The Goals test was extracted into its own test file; it fails because `p` hides a project's tasks. Component tests first fail compilation because the base lacks `Fill` and `Hint`. I added only inert fields in the scratch base; both tests then fail their assertions. Logs: `base.log`, `base-goals.log`, `base-component.log`, `base-component-behavior.log`.

Task green command, run before and after the mutations:

```sh
go test ./internal/tui ./internal/tui/component -run 'TestWorkTasks|TestGoalsOpenTasksListsEachProject|TestListFillGolden|TestSubTabsHint|TestQueueKeys|TestScreenSelectionsUseHoneyMarker' -v
```

Both runs exit 0 (`green.log`, `restored-green.log`). This includes the whole Work and Compact grid comparisons, responsive and empty/long-data tests, normal-mode scratch-home loaders and refresh, and key routes.

Token break: changed the landed status from `doneStyle` to `infoStyle`. `go test ./internal/tui -run TestWorkTasksMatchesCompactBoard -v` fails with 66 differing cells (`token.log`).

Key break: changed `case "d"` to `case "D"` in `workTasksScreen.Update`. `go test ./internal/tui -run 'TestWorkTasks(FileKeys|OverlayKeys)' -v` fails because dependencies do not open (`key.log`). Both mutations were restored. The source diff was empty before the merge attempt.

## V2 runs

I built a fresh binary with `go build -o .hive/tmp/verify37r-hive .`. I created scratch tab `wAD:t2`, pane `wAD:p2`, and a scratch home with a home marker. Each run used:

```sh
HIVE_HOME=SCRATCH_HOME COLORTERM=truecolor SCRATCH_BINARY --demo .hive/spec/tui-v2/design/Work.fixture.json --size 120x40
HIVE_HOME=SCRATCH_HOME COLORTERM=truecolor SCRATCH_BINARY --demo .hive/spec/tui-v2/design/Compact.fixture.json --size 80x24
herdr pane read wAD:p2 --source visible --format ansi
```

For each board, two reads 500 ms apart were identical. The pane records show scroll offset 0. The CLI capture does not expose a `truncated` field, so that metadata was not independently checked in this run. Captures contain the complete board areas. The comparator from the first verifier crops the top-left board area, converts ANSI with `cellgrid`, and compares every non-space cell fully. Each diff file reports `0 cells differ`. Captures, pane records, and diff files are in `verify2/`.

Work key observations are saved in `key-*.txt`: `j` selects task 2; `k b` opens task 3's restart brief; `esc v` opens `VERIFIED.md` with `Result: pass`; `d` opens its dependency graph; `esc enter` returns to task details with the details footer. Compact keys were covered by the repeated model tests and the prior verifier's live checks; I did not repeat Compact live keys in this repair run.

Approved boards captured again with Firefox, new scratch profiles, `--headless --no-remote --window-size=1220,780 --screenshot`:

- `/home/penguin/source/hive/.hive/spec/tui-v2/design/Work.dc.html`
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/Compact.dc.html`

I converted each new ANSI capture to HTML and captured that with Firefox. PNG margins and font size differ from the boards; the grid comparisons prove the cell positions and styles. I reviewed all four PNG files and the plain text without colour.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Work header | working, live, needs you 0, 14:22 | matches | Fixed fixture. |
| Work tab bar | 3 Work, underline | matches | Active tab. |
| Work view row | [Tasks], switch view | matches | Active view. |
| Work list frame and title | tasks · 4, borders | matches | Focused list. |
| Work columns | ID, STATUS, WAVE | matches | Board labels. |
| Work selection | ▶ 3-tui-restart | matches | Visible selection glyph. |
| Work task rows | landed, assigned, C, B, A, — | matches | Words identify states. |
| Work details frame | 3-tui-restart, borders | matches | Side-by-side details. |
| Work detail fields | title, status, goal, worker, verifier, branch, merge, depends | matches | Fixture values. |
| Work acceptance | ✓ AC1 through ✓ AC6 | matches | Pass glyphs. |
| Work evidence | EVIDENCE, b, v, d, V1 pass, V3 pass | matches | Words identify results. |
| Work footer | [list], action hints | matches | List focus. |
| Compact header | coord ●, sup ●, needs you 0, 14:22 | matches | State glyphs. |
| Compact tab bar | 3 Work, underline | matches | Active tab. |
| Compact view row | [Tasks], list › details | matches | Overlay focus. |
| Compact overlay frame | 3-tui-restart · esc back, borders | matches | Full-width details. |
| Compact fields | status landed, goal, verifier pass, merge | matches | Words identify states. |
| Compact acceptance | ✓ AC1 through ✓ AC6 | matches | Pass glyphs. |
| Compact evidence | b brief.md, v VERIFIED.md, d deps graph | matches | Action labels. |
| Compact overflow | ↓ 2 more lines, ┃ | matches | Visible overflow glyphs. |
| Compact footer | [details], scroll and action hints | matches | No Help hint. |

## Gate and cleanup

`verify2/gate.log` contains all five exact gate commands and their exit statuses on the conflicted merge result. The failures are syntax errors from merge markers, not the known runtime flakes. No successful merged-tree gate is claimed. `verify2/scope.txt` is the task diff against dev after abort; it is not a successful merged-result scope check.

I closed the owned scratch tab and pane. The tab list then showed only `wAD:t1`. I removed the extracted base, scratch home, binary, browser profiles, and temporary capture scripts after recording this evidence. No task source mutation remains. Worker evidence and the plan tracker were not edited.
