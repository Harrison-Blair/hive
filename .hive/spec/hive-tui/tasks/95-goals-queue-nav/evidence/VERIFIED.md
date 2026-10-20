# 95-goals-queue-nav verification

| Item | Value |
| --- | --- |
| Task | `95-goals-queue-nav` |
| Commit | `a49c148f` (task code `5ed02380`), base `origin/dev` `bb731d3e` |
| Date | 2026-10-20 |
| Agent | `t95-verifier-1`, Claude Opus 5.5 |
| Verdict | **fail** |

## Results

| ID | Check | Result | Observation |
| --- | --- | --- | --- |
| V1 | Red run on the base, green run on the task code | pass | All 8 new or changed tests fail on `bb731d3e` for the correct reason. All 8 pass on the task code. |
| V2 | Live screen agrees with the board for each gap | **fail** | Q1, Q2, Q3 pass. G1 `o` passes. G1 `t` fails: when the Tasks project filter is not `all`, `t` does not show the wave task. See [G1 `t` defect](#g1-t-defect). |
| V3 | `git diff origin/dev --stat` | pass | Only the files of the source table and the task evidence changed. |
| V4 | Gate | pass | `gofmt -l` and `go mod tidy -diff` print nothing. `go vet`, `golangci-lint` (0 issues), and `go test -race ./...` exit 0. |
| Frame | Exact frame at 120 x 40 and 200 x 57 | pass | Each capture is exactly W x H. |
| Colour | Each state has a glyph or a word in plain text | pass | `TestGoalsAndQueueStatesHavePlainTextSignals` uses `ansi.Strip`, fails on the base, and passes on the task code. The live plain text shows each state. |
| Design | Worker element table | pass, with notes | See [Design check](#design-check). |

## V1

Red run: a detached scratch worktree at `origin/dev` `bb731d3e` with the task test file copied in. I removed the worktree after the run.

```text
go test ./internal/tui -run '^(TestGoalsOpenTasksForSelectedWave|TestGoalsOpenWaveDesign|TestGoalsAndQueueStatesHavePlainTextSignals|TestQueueHeaderShowsRunStopAndWaitCounts|TestQueueShowsEveryRegisteredProjectWithoutProjectCycle|TestQueueArrowsSelectTasksAndEnterOpensSelectedTask|TestQueueScreenShowsTransitiveStopsAndIgnoresProjectCycle|TestGoalsRefreshKeepsGoalSelectionAndRefreshesQueueProjects)$' -count=1
--- FAIL: TestQueueScreenShowsTransitiveStopsAndIgnoresProjectCycle   p changed Queue project panels
--- FAIL: TestGoalsOpenTasksForSelectedWave   screen=2 tab=0 task=""
--- FAIL: TestGoalsOpenWaveDesign (one_feature, first_feature_wins: no file opened; no_feature, missing_file: no "no design for this wave")
--- FAIL: TestGoalsAndQueueStatesHavePlainTextSignals   omitted "▶ demo · Goal one · active · done when: landed on dev"
--- FAIL: TestQueueHeaderShowsRunStopAndWaitCounts   omitted "⬢ hive · runs without you 3 · stops 1 · waits behind a stop 1"
--- FAIL: TestQueueShowsEveryRegisteredProjectWithoutProjectCycle   omitted "empty · no active goal · grant: —"
--- FAIL: TestQueueArrowsSelectTasksAndEnterOpensSelectedTask   screen=3 tab=0 task=""
--- FAIL: TestGoalsRefreshKeepsGoalSelectionAndRefreshesQueueProjects   did not retain all projects
FAIL	github.com/Harrison-Blair/hive/internal/tui	0.011s
```

Green run on the task code: the same command prints `--- PASS` for each of the 8 tests and `ok github.com/Harrison-Blair/hive/internal/tui 0.016s`.

## V4

```text
== gofmt          (no output) exit 0
== go mod tidy -diff (no output) exit 0
== go vet ./...   exit 0
== golangci-lint run ./...   0 issues. exit 0
== go test -race ./...   every package ok or no test files, exit 0
```

## V2 live run

- Binary: `go build -o .hive/tmp/v95/hive-v95 .` at `a49c148f`.
- Scratch home `.hive/tmp/v95/home`: copies of the `goals`, `projects`, and `tasks` records of the real home, and `.hive/spec/hive-tui/`. To get a stop and waits, I set `needs_user` on the scratch copy of `102-selection-marker`. Temp `XDG_CONFIG_HOME` `.hive/tmp/v95/xdg`.
- My tab `wTG:t2` (label `v95-live`), pane `wTG:p2`, in my task workspace `wTG`. I did not use workspace `wBC`.
- Exact frame: in pane `wTG:p2`, a tmux server on a private socket (`tmux -L v95`) held one window of fixed size (`-x 200 -y 57`, then `-x 120 -y 40`, status off, `window-size manual`). The TUI ran in that window with `env -u HERDR_ENV -u HERDR_PANE_ID -u HERDR_WORKSPACE_ID -u HERDR_SOCKET_PATH HIVE_HOME=… XDG_CONFIG_HOME=…`. `tmux list-panes` showed `200x57` and `120x40`. Each capture used `herdr pane read wTG:p2 --source visible` and `tmux capture-pane`.
- TUI pids: `3268725` (200 x 57), `3276224` (120 x 40). The footer read `not in herdr: no coordinator or supervisor started`. I did not open the Chat input.

| Check | 200 x 57 | 120 x 40 |
| --- | --- | --- |
| Goals frame | rows=57, maxcols=200 | rows=40, maxcols=120 |
| Goals `t` | Tasks, `> 87-tidy-gate` (first task of wave D) | Tasks, frame exact |
| Goals `o` | Tasks file viewer, `README.md`, `# Hive TUI` | frame exact |
| Goals `o`, no README | — | `no design for this wave` on the Goals screen, rows=40 |
| Queue header (Q1) | `⬢ hive · runs without you 2 · stops 1 · waits behind a stop 2` | same, time at the right edge |
| Queue panels (Q2) | `hive`, `qmeter · no active goal · grant: —` with `(no open tasks)`, `stenographer` | same |
| Queue `p` (Q2) | frame unchanged after `p` | — |
| Queue `s` | stops only; `stenographer` shows `(no tasks needing you)` | — |
| Queue `↓` (Q3) | `▶ 1  102-selection-marker …  ✋ you` | `▶` moves |
| Queue `enter` (Q3) | Tasks, `> 102-selection…`, details `102-selection-marker` | Tasks, `> 102-selection…` |

The counts agree with the records: run = `95-goals-queue-nav`, `111-hub-config-board`; stop = `102-selection-marker`; waits = `105-verified-close`, `114-visual-check-rule`.

### G1 `t` defect

Steps at 120 x 40: press `5` (Tasks), press `p` twice (`project: qmeter`), press `3` (Goals), press `t`.

```text
│ SCREENS                    ││ Tasks                                                                                  │
│   1 Chat                   ││ active 0  queued 0  landed 8  failed 0  [all 9]                                        │
│   2 Agents                 ││ project: qmeter                                                                        │
│   3 Goals                  ││ ID             title                       project   status     agent · verification   │
│   4 Queue                  ││ > 49-simplify-p… Shared A5 bar, card and R1… qmeter    landed     t49-worker-1 · pass… │
```

Expected: Tasks with the wave D task `87-tidy-gate` selected and visible. Observed: the project filter stays `qmeter`, no wave task is visible, and `> 49-simplify-p…` is selected. Cause: the `t` case in `goalsScreen.Update` (`internal/tui/goals_queue.go`) sets `tab`, `selectedID`, and `scrollOffset`, but it does not set `s.model.tasks.projectIndex = 0`. The Queue `enter` case sets it. `TestGoalsOpenTasksForSelectedWave` starts with the default filter, so it does not find this defect.

Note for the manager: with the default filter, `t` opens the full Tasks list with only the first wave task selected. The other wave tasks are not grouped or filtered. The board says only `t tasks of wave`. I record this as a note. It is not a cause of the fail.

### G1 `o` and the manager decision

The decision: open `.hive/spec/<feature>/README.md` of the first task of the wave; `no design for this wave` when that task has no feature or no file; no new record field. The code agrees for each case that the tests and the live run cover. One difference: the code takes the first task of the wave **that has a feature**. When the first task has no feature and a later task has one, the code opens the README of the later task, and the decision gives `no design for this wave`. No test covers this case. No record field was added.

## Design check

Boards rendered with `firefox --headless --no-remote --profile .hive/tmp/v95/ffprof --window-size=1220,780 --screenshot`. Screens rendered from the `herdr pane read --format ansi` captures (cropped to W x H) through a short SGR-to-HTML script, then the same firefox command. Not committed:

- Boards: `.hive/tmp/v95/png/Goals-board.png`, `.hive/tmp/v95/png/Queue-board.png`
- Screens: `.hive/tmp/v95/png/goals-200.png`, `goals-120.png`, `queue-200.png`, `queue-sel-200.png`, `queue-120.png` in the same folder

The browser font draws some glyphs (`✓`, `▶`, `◎`) wider, so the right borders in the screen PNGs do not align. The plain-text captures show each border at the correct cell.

Judgement of the worker element table:

| Board element | Worker result | My result |
| --- | --- | --- |
| Goals header, screen list | differs (fixture) / matches | agree |
| Goals wave note, goal heading, progress path, task summary, unassigned panel | differs | agree. These are layout from task 82 and are outside G1. Not a fail. |
| Goals footer `t` / `o` | overridden by G12 | agree. Board `enter decisions` is absent, which G12 removes. |
| Queue header counts | matches | agree. The words and order match the board. |
| Queue column headings | matches | partly. The words match. The board has aligned columns; the screen joins the fields with two spaces. Layout from task 82, outside Q1–Q3. Not a fail. |
| Queue project headings | differs (no done-when point) | agree. Outside Q1–Q3. |
| Queue state words | differs (`assigned` for `running`) | agree. Record status. |
| Registered project with no open tasks | matches | agree |
| Queue selected task `▶` | overridden (colourblind rule) | agree |
| Queue footer, no `p` | overridden by decision 9 | agree |

The "state without colour" column of the worker table agrees with the live plain text: `active`/`done`/`open` words, `✓ ▶ ✋ · ◎`, `▶ verify`, `▶ assigned`, `✋ you`, `◆ auto`, `┄ waits`, `(no open tasks)`, `▶` on the selected goal, screen, and Queue row.

## Cleanup

- I stopped only TUI pids `3268725` and `3276224` with `ctrl+c ctrl+c`. After each stop, `ps -p` showed no process and `tmux -L v95 ls` showed no server.
- I closed only my split pane `wTG:p3` (made and closed before the tmux run) and my tab `wTG:t2`. `herdr tab list --workspace wTG` then showed only `wTG:t1`.
- I started no `hive supervise` process. The one `hive supervise` process (pid `2107907`, workspace `wBC`, started 13:38) is the user's. I did not touch it.
- I removed the scratch home, binary, captures, scripts, and browser profile below `.hive/tmp/v95/`. I kept `.hive/tmp/v95/png/`.
