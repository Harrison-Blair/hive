# Verification of task 134

Verdict: **fail**. V3 fails the coordinator selection check when no pod is visible.

- Date: 2026-10-23.
- Verifier: `t134-verifier-1`.
- Task commit: `0bbf3a079bb3e20054fb045dbad3ec8531d32605`.
- Base for the red run: `6a586f1a`.
- Current `dev` and `origin/dev`: `aaa2e51164723662e632c6c5bceb61a58e28362b`.
- `git merge origin/dev` printed `Already up to date.` before the gate.
- A second `git fetch origin dev` before the verdict gave the same `origin/dev`. It is an ancestor of the task commit.
- Approval: ask `a5` and coordinator record 536. Package: `219eddba`. Final approved grids: `21495e2e`.
- This run changed no product source, worker evidence file, board, golden grid, or tracker.

## Verification items

| Item | Command or check | Result | Evidence |
| --- | --- | --- | --- |
| V1 | Base red run, guard mutations, focused tests with `-count=50` and `-race -count=50` | pass | `verifier/red-base.log`, `guard-*.log`, `green-count50.log`, `green-race50.log` |
| V2 | Export each board, compare with `cellgrid.Compare`, run each golden check, check the approved diff | pass | `verifier/export.log`, `grid-compare.log`, `golden.log`, `diff-stat.log` |
| V3 | Build the task binary, capture both views at both widths, check keys and visual states | fail | `verifier/live/`, `live-summary.log`, `live-unchanged.log`, `coordinator-selection.log`, `coordinator-model-red.log` |
| V4 | Merge current `dev`, check scope, run the five gate commands | pass | `verifier/diff-stat.log`, `gate.log` |

All paths in the Evidence column are relative to this folder.

## V1: independent red and green runs

The run used each test in `agents_hide_test.go` and `TestAgentsPodsFollowProjectRecordOrder`.
The full command and test names are in the logs.

An isolated source copy held the new tests with the base `agents_screen.go`.
Nine tests failed on the base for the expected manager, pod, hint, or selection difference.
The base still drew the hidden manager and lacked the new title hint.
Three guards passed on the base.

Each base-passing guard failed after a separate deliberate change:

| Guard | Deliberate change | Observed failure |
| --- | --- | --- |
| `NoHintWithNothingHidden` | Add the hint when the count is zero | Unexpected `h shows 0 stopped` |
| `GeneralChildOfAHiddenWorkerMovesToHome` | Keep the hidden worker as the parent anchor | Missing `home` pod |
| `PodsFollowProjectRecordOrder` | Reverse the visible pods | Wrong project order |

An additional mutation moved the empty-state notice down one row.
`NoticeAloneWhenNothingIsVisible` failed on its literal row check.
Each mutation used only the isolated copy. The verifier restored the source after each run.

Both green commands ran from the task worktree:

- `go test ./internal/tui/ -run "$R" -count=50 -v`: exit 0, 600 PASS.
- `go test -race ./internal/tui/ -run "$R" -count=50 -v`: exit 0, 600 PASS.

## V2: grids and approval

Commands:

```sh
go run ./internal/boardexport/cmd/boardexport -out .hive/tmp/verify134/export .hive/spec/tui-v2/design
go test ./internal/tui/ ./internal/boardexport/ -run 'Board|Golden|Grid' -count=1 -v
```

The exporter wrote 47 grids. Each grid matched its committed golden grid.
Each of the nine changed grids also matched its approved grid in `new-grids/`.
The 56 comparisons used `cellgrid.ReadGrid` and `cellgrid.Compare` with no mask.
They compared every rune, token, and bold flag. Each comparison gave zero differences.
The test command exited 0, with 201 PASS lines and no FAIL line.

Only the nine approved board/grid pairs changed in `design/`.
Their changed cells are on title row 3. No board fixture or mask changed.

## V3: live captures

The verifier built `.hive/tmp/verify134/hive` with `go build`.
The runs used the verifier's scratch pane `wWS:p2`, in tab `wWS:t2`.
The pane had at least 120 columns and 40 rows.
Each run used a fixture and a scratch home, with these environment values:

```text
HIVE_HOME=<task-worktree>/.hive/tmp/verify134/home
HIVE_NO_STARTUP=1
TERM=vt100
COLORTERM=truecolor
```

Each read used `source=visible` and `format=ansi`.
Two equal reads, 500 ms apart, established a stable screen.
Each metadata record shows `truncated=false` and scroll offset 0.
The check cropped the top-left board size and used `cellgrid.Convert` and `cellgrid.Compare` with no mask.

Each of these 20 initial captures matched its approved grid, including colour and weight:

- Nine changed boards: `AgentsHiddenParent`, `LiveOverview`, `LivePod`, `CompactLiveOverview`, `CompactLivePod`, `TreeNested`, `TreeWide`, `CompactTreeNested`, `CompactTreeWide`.
- Six scenarios: `HideManagerOverview`, `CompactHideManagerOverview`, `HideManagerPod`, `CompactHideManagerPod`, `HideAllPods`, `CompactHideAllPods`.
- Five unchanged references: `Main`, `AgentsMany`, `CompactAgents`, `AgentsHome`, `AgentsCoordinatorChild`.

The two overview captures after the second `h` also matched their approved grids.
Thus all 22 recorded grid comparisons gave zero differences.
`verifier/live/` holds the independent ANSI, plain text, PNG, metadata, and diff files.

The key runs checked these behaviors at both widths:

- A direct general child of the hidden manager stays in its project pod.
- The one-pod line reaches each top child. Right selects the general child.
- `h` shows the stopped manager. The second `h` restores the approved screen.
- With all project pods hidden, the coordinator, home cards, notice, and title hint stay visible.
- `enter` opens home. `esc` returns to the overview.
- In a separate run, `h`, right, `h`, `h` restores the selected project pod.
- With no registered project, the body keeps the notice alone, even with live coordinator/home records.
- With registered projects but no visible card, the notice is alone at row 5, column 3.
- With only the coordinator visible, the notice and coordinator render. Selection fails as described below.

### Finding: coordinator actions fail when the last pod hides

This is an AC2 selection failure. Decision 65 keeps coordinator selection and its details popup.
The manager accepted this scope in the message after inbox record 541.

Reproduction fixtures are durable:

- `verifier/fixtures/coordinator-only-120.fixture.json`
- `verifier/fixtures/coordinator-only-80.fixture.json`

Run the task binary with either fixture, `HIVE_NO_STARTUP=1`, `COLORTERM=truecolor`, and the corresponding `--size`.
Use a scratch home. Each fixture contains projects, hidden stopped project agents, a live coordinator, and no home card.

| Step | Expected | Actual at 120x40 and 80x24 |
| --- | --- | --- |
| Start, then up | Select the visible coordinator with a heavy border | No screen change |
| Press enter | Open the coordinator details popup | No screen change |
| Press `h`, up, enter | Open coordinator details with project pods visible | Works: heavy border and details popup |
| Press esc, then `h` | Hide the project pods and keep coordinator selection | The coordinator keeps its heavy border |
| Press enter again | Open the selected coordinator details | No screen change |

`coordinator-selection.log` records both widths and the `h` control.
The `coordinator-only-*-keys-*.txt` and `.ansi` files record each step.
The `keys-enter.png` and `keys-selected-hidden.png` files show the failure at each width.

A separate model reproduction also failed at both widths.
The durable source is `verifier/coordinator-regression.go.txt`.
Copy it beside the TUI tests as a `.go` test file, then run:

```sh
go test ./internal/tui/ -run '^TestVerifierCoordinatorSelectableWithAllProjectPodsHidden$' -count=1 -v
```

Expected: the view contains `coordinator · esc closes` and `[details]` after up and enter.
Actual: both subtests fail because no popup opens. See `coordinator-model-red.log`.

Cause:

- `selectedAgent` returns no agent when `len(pods) == 0` (`agents_screen.go:372`).
- `updateAgentsKey` routes overview keys only when a pod exists (`agents_screen.go:1258`).
- The zero-pod case therefore cannot select or open the coordinator that the new layout keeps visible.

### Visual table

The verifier captured 14 approved boards again with Firefox and fresh scratch profiles.
The nine changed board PNGs were byte-equal to the worker's board PNGs.
The grid comparisons establish cell equality independently of browser margins and fonts.
The six scenario grids remain the approved references from decision 112.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header and tabs | `coordinator`, `supervisor`, `●`, active underline | matches | Same frame as the approved boards |
| Outer title and persistent hint | `project agents`, `h shows N stopped` | matches | Decision 109, including 80 columns |
| Hint with nothing hidden | No `h shows` text | matches | First `h` restores every stopped card |
| Hidden-manager pod border, title, goal | Box glyphs, project name, goal words | matches | Decision 107 |
| Overview top children | Complete cards without rail joints | matches | Same layout as home |
| One-pod top-child connectors | `┼`, `│`, bus, `┷`, `┴` | matches | Rule A and decision 107 |
| Direct general child | `◐ api-review-1` in `example-api` | matches | Ask `a5`, decision 107 |
| Nested cards and scroll marks | `┬`, `┴`, `┃`, `━`, arrow words | matches | Existing tree boards remain equal |
| Pod counts | `h shows 2 stopped`, `h hides 2 stopped` | matches | Hidden and shown managers count |
| Card states | `● working`, `◐ idle`, `○ stopped` | matches | State words or glyphs remain without colour |
| Lit message path | Heavy line and arrow, legend words | matches | Existing live boards remain equal |
| Selected pod or card | Heavy border | matches | Approved scenario and key captures |
| All-hidden project notice plus home | `no project agents`, coordinator card, `home` | matches | Decision 110 at both widths |
| No registry, or no visible card | `no project agents` alone in the body | matches | Decision 10 and decision 110 |
| Legend and key hints | Plain text key names | matches | Same as the approved grids |
| Coordinator selection with zero pods | No heavy border after up | differs | AC2 and decision 65 require valid selection |
| Coordinator details with zero pods | No popup after enter | differs | The selected coordinator also loses its action |

## V4: diff scope and gate

`git diff dev...HEAD --stat` names only the allowed source, spec, design, baseline, and task evidence files.
The verifier adds only evidence under this task.
`verifier/gate.log` contains the full output of these commands:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go mod tidy -diff
go vet ./...
golangci-lint run ./...
go test -race ./...
```

Each command exited 0. The first two commands printed nothing.
The linter printed `0 issues.`. All race tests passed.
These green checks do not cover the new coordinator-only reproduction.

## Cleanup

The verifier closed only its scratch pane `wWS:p2` and tab `wWS:t2`.
The task worktree and the verifier agent pane remain for the failed task.
The manager must stop the verifier agent after the verdict.

The verifier removed these scratch folders after recording the results:

- `<task-worktree>/.hive/tmp/verify134/`: source copy, scripts, binaries, exports, profiles, scratch home, and test data.
- `/home/penguin/source/hive/.hive/tmp/verify134-boards/`: independent board PNGs.
- `/home/penguin/source/hive/.hive/tmp/134-boards/`: worker board PNGs, as the manager requested.

The final captures, logs, reproduction source, and reproduction fixtures remain in this evidence folder.
No other task scratch resource remains. The task worktree stays available for the next worker.
