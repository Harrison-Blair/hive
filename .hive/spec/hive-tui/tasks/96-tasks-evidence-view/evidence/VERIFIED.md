# 96-tasks-evidence-view VERIFIED

| Item | Value |
| --- | --- |
| Task | 96-tasks-evidence-view |
| Verifier | t96-verifier-1, Claude Code, claude-opus-5-5 |
| Task commit | `dd56c496` |
| Base | `origin/dev` `a51c833f` (code equal to `b4f1d71e`; the newer commit changes one dogfood doc only) |
| Date | 2026-10-20 |
| Verdict | fail (first pass); pass (Repair 1, see below) |

| ID | Result | Note |
| --- | --- | --- |
| V1 | pass | The 3 new tests fail on the base and pass on the task branch. |
| V2 | fail | T1, T2, and T3 show as the board shows at 207 x 57. The `v` and `d` views break the frame: the panels collapse to 8 rows and the frame is 1 line taller than the terminal at 120 x 40 and 200 x 57. The brief requires that each frame stays exactly the terminal size. |
| V3 | pass | The task changes only the files of the source table. |
| V4 | pass | The gate is clean. |

## V1

The red run used a detached scratch worktree at `b4f1d71e` with the task file `internal/tui/task_screen_test.go` copied in.

```sh
go test ./internal/tui -run '^(TestTasksScreenShowsEvidencePanelAndDefinitions|TestTasksScreenVOpensVerifiedEvidence|TestTasksScreenDOpensSelectedTaskDependencyGraph)$' -count=1
```

Base:

```text
--- FAIL: TestTasksScreenShowsEvidencePanelAndDefinitions (0.00s)
    task_screen_test.go:34: Tasks view does not contain "task details":
    task_screen_test.go:34: Tasks view does not contain "┤ evidence ├":
    task_screen_test.go:34: Tasks view does not contain "brief: tasks/79-tui-shell/brief.md":
    task_screen_test.go:34: Tasks view does not contain "landed = delivery commit set and pushed · verified = verification passed":
--- FAIL: TestTasksScreenVOpensVerifiedEvidence (0.00s)
    task_screen_test.go:52: v did not request VERIFIED.md
--- FAIL: TestTasksScreenDOpensSelectedTaskDependencyGraph (0.00s)
    task_screen_test.go:76: d did not log opening the dependency graph
FAIL	github.com/Harrison-Blair/hive/internal/tui	0.003s
```

Task branch (`-v`):

```text
--- PASS: TestTasksScreenShowsEvidencePanelAndDefinitions (0.00s)
--- PASS: TestTasksScreenVOpensVerifiedEvidence (0.00s)
--- PASS: TestTasksScreenDOpensSelectedTaskDependencyGraph (0.00s)
ok  	github.com/Harrison-Blair/hive/internal/tui	0.004s
```

## V2

Scratch home: a marked home in the verifier scratchpad with 3 task records (`79-tui-shell` verified, needs `81-agents-screen`; `81-agents-screen` delivered and pushed; `83-downstream-task` needs `79-tui-shell`), a brief, `V1.md`, and `VERIFIED.md`. `HIVE_HOME` and `XDG_CONFIG_HOME` pointed into it. The binary was built from `dd56c496`.

Before the run: workspace `wQV` had only tab `wQV:t1`. `hive supervise` pid 1288721 already ran; the verifier did not start it.

```sh
herdr tab create --workspace wQV --cwd <scratch>/home/bin --label t96-verifier-v2 --env HIVE_HOME=<scratch>/home --env XDG_CONFIG_HOME=<scratch>/home/config --no-focus
# returned tab wQV:t2, pane wQV:p2; layout 207 x 57
herdr pane run wQV:p2 'env -u HERDR_ENV -u HERDR_PANE_ID -u HERDR_WORKSPACE_ID -u HERDR_SOCKET_PATH -u HERDR_TAB_ID ./hive'
herdr pane send-keys wQV:p2 5
herdr pane read wQV:p2 --source visible --lines 80
```

The TUI printed `not in herdr: no coordinator or supervisor started`. No chat was sent.

### T1 and T3: pass at 207 x 57

The board has a `task details` block and an `┤ evidence ├` block with the brief and the verification history. The capture shows two separate bordered panels to the right of the list, and the note line. The note uses G8 for `landed`, as the brief requires (the board text differs; G8 wins).

```text
│ ▶ 5 Tasks                  ││   83-downstream… Downstream task             hive      assigned   t83-worker-1 · running       │ blocks: 83-downstream-task                                                │ │
...
│                            ││                                                                                                ╭───────────────────────────────────────────────────────────────────────────╮ │
│                            ││                                                                                                │ ┤ evidence ├                                                              │ │
│                            ││                                                                                                │ brief: tasks/79-tui-shell/brief.md                                        │ │
│                            ││                                                                                                │ evidence: VERIFIED.md, V1.md                                              │ │
│                            ││                                                                                                │ V1 fail · V1 failed: missing shell load · 2026-10-20T10:00:00Z            │ │
│                            ││                                                                                                │ current pass · - · 2026-10-20T11:00:00Z                                   │ │
...
│                            ││ landed = delivery commit set and pushed · verified = verification passed                                                                                                     │
```

Full capture:

```text
⬢ hive · 1 project · 0 live agents · working unknown / idle unknown · supervisor stale · needs_user 0 · 12:18
╭────────────────────────────╮╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ SCREENS                    ││ Tasks                                                                                                                                                                        │
│   1 Chat                   ││ [active 2]  queued 0  landed 1  failed 0  all 3                                                ╭───────────────────────────────────────────────────────────────────────────╮ │
│   2 Agents                 ││ project: all                                                                                   │ task details                                                              │ │
│   3 Goals                  ││ ID             title                       project   status     agent · verification           │ 79-tui-shell · TUI shell                                                  │ │
│   4 Queue                  ││ > 79-tui-shell   TUI shell                   hive      verified   t79-worker-2 · pass          │ needs: 81-agents-screen ✓                                                 │ │
│ ▶ 5 Tasks                  ││   83-downstream… Downstream task             hive      assigned   t83-worker-1 · running       │ blocks: 83-downstream-task                                                │ │
│   6 Friction               ││                                                                                                │ branch: task/79-tui-shell                                                 │ │
│   7 Health                 ││                                                                                                │ feature: hive-tui                                                         │ │
│   8 Logs                   ││                                                                                                │ grant: -                                                                  │ │
│                            ││                                                                                                │ needs user: -                                                             │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                ╰───────────────────────────────────────────────────────────────────────────╯ │
│                            ││                                                                                                ╭───────────────────────────────────────────────────────────────────────────╮ │
│                            ││                                                                                                │ ┤ evidence ├                                                              │ │
│                            ││                                                                                                │ brief: tasks/79-tui-shell/brief.md                                        │ │
│                            ││                                                                                                │ evidence: VERIFIED.md, V1.md                                              │ │
│                            ││                                                                                                │ V1 fail · V1 failed: missing shell load · 2026-10-20T10:00:00Z            │ │
│                            ││                                                                                                │ current pass · - · 2026-10-20T11:00:00Z                                   │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                │                                                                           │ │
│                            ││                                                                                                ╰───────────────────────────────────────────────────────────────────────────╯ │
│                            ││ landed = delivery commit set and pushed · verified = verification passed                                                                                                     │
╰────────────────────────────╯╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
not in herdr: no coordinator or supervisor started

tab filter · enter evidence · b brief · v verified · d deps graph · p project · ask in Chat · ↑/↓ move
```

Observation: at 120 x 40 the content width is below 120, so the task shows `┤ task details ├` and `┤ evidence ├` as headings in the list panel, not as separate panels. The V2 item names about 200 x 57, so this is not the fail reason.

### T2: content pass, frame fail

`v` opens `VERIFIED.md` and `d` opens the dependency graph, both read-only. Escape closes each. But the frame breaks: the sidebar closes at row 13, the main panel closes at row 14, and the rest of the screen is blank.

`v` capture (rows 1-15):

```text
⬢ hive · 1 project · 0 live agents · working unknown / idle unknown · supervisor stale · needs_user 0 · 12:18
╭────────────────────────────╮╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ SCREENS                    ││ Tasks                                                                                                                                                                        │
│   1 Chat                   ││ VERIFIED.md                                                                                                                                                                  │
│   2 Agents                 ││ # VERIFIED                                                                                                                                                                   │
│   3 Goals                  ││                                                                                                                                                                              │
│   4 Queue                  ││ | Item | Result |                                                                                                                                                            │
│ ▶ 5 Tasks                  ││ | --- | --- |                                                                                                                                                                │
│   6 Friction               ││ | V1 | pass |                                                                                                                                                                │
│   7 Health                 ││                                                                                                                                                                              │
│   8 Logs                   ││                                                                                                                                                                              │
│                            ││                                                                                                                                                                              │
│                            ││                                                                                                                                                                              │
│                            ││                                                                                                                                                                              │
│                            ││                                                                                                                                                                              │
```

`d` capture (rows 1-15):

```text
⬢ hive · 1 project · 0 live agents · working unknown / idle unknown · supervisor stale · needs_user 0 · 12:18
╭────────────────────────────╮╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ SCREENS                    ││ Tasks                                                                                                                                                                        │
│   1 Chat                   ││ dependency graph · 79-tui-shell                                                                                                                                              │
│   2 Agents                 ││ 79-tui-shell · TUI shell                                                                                                                                                     │
│   3 Goals                  ││ needs:                                                                                                                                                                       │
│   4 Queue                  ││ 79-tui-shell ← 81-agents-screen (needs)                                                                                                                                      │
│ ▶ 5 Tasks                  ││ blocks:                                                                                                                                                                      │
│   6 Friction               ││ 79-tui-shell → 83-downstream-task (blocks)                                                                                                                                   │
│   7 Health                 ││                                                                                                                                                                              │
│   8 Logs                   ││                                                                                                                                                                              │
│                            ││                                                                                                                                                                              │
│                            ││                                                                                                                                                                              │
│                            ││                                                                                                                                                                              │
│                            ││                                                                                                                                                                              │
```

A verifier model test (scratch only, not committed) checks the full `Model.View()` with `writeTasksFixture` at 120 x 40 and 200 x 57, then sends no key, `v`, `d`, or `b`. It asserts the line count and the display width of each line.

```text
task dd56c496:
--- PASS: TestVerifierTasksFrameWithData/120x40/
--- FAIL: TestVerifierTasksFrameWithData/120x40/v   lines 41 want 40
--- FAIL: TestVerifierTasksFrameWithData/120x40/d   lines 41 want 40
--- FAIL: TestVerifierTasksFrameWithData/120x40/b   lines 41 want 40
--- PASS: TestVerifierTasksFrameWithData/200x57/
--- FAIL: TestVerifierTasksFrameWithData/200x57/v   lines 58 want 57
--- FAIL: TestVerifierTasksFrameWithData/200x57/d   lines 58 want 57
origin/dev a51c833f:
--- FAIL: TestVerifierTasksFrameWithData/120x40/b   lines 41 want 40
--- FAIL: TestVerifierTasksFrameWithData/200x57/b   lines 58 want 57
```

120 x 40 frame after `v` (blank panel rows cut):

```text
⬢ hive · 0 projects · 0 live agents · working unknown / idle unknown · supervisor stale · needs_user 0 · 08:00
╭────────────────────────────╮╭────────────────────────────────────────────────────────────────────────────────────────╮
│ SCREENS                    ││ Tasks                                                                                  │
│   1 Chat                   ││ VERIFIED.md                                                                            │
│   2 Agents                 ││ # VERIFIED                                                                             │
╰────────────────────────────╯│                                                                                        │
                              ╰────────────────────────────────────────────────────────────────────────────────────────╯
(blank rows)
tab filter · enter evidence · b brief · v verified · d deps graph · p project · ask in Chat · ↑/↓ move
```

The `b` brief view had this fault on `origin/dev` before the task. The new `v` and `d` views reuse the same file viewer, so they have the fault too. The viewer code is in `internal/tui/friction_screen.go` (`fileView`, about line 489), which is not in the source table of the brief. The existing `TestFramesMatchTerminalSizesOnEveryScreen` passes because it uses an empty home and opens no file view.

Cleanup: the verifier closed only tab `wQV:t2` (the ID that `tab create` returned). After that, `wQV` has only `wQV:t1`. `pgrep -af 'hive supervise'` shows only pid 1288721, which ran before the check. The scratch worktree was removed.

## V3

```sh
git diff origin/dev --stat
```

The two-dot diff also shows `reference/dogfood/friction/entries/scratch-hive-starts-real-agents.md | 10 -`, because `origin/dev` moved on to `a51c833f` after the last merge into the task branch. The three-dot diff gives the changes of the task:

```text
$ git diff origin/dev...HEAD --stat
 .../hive-tui/tasks/96-tasks-evidence-view/brief.md |  65 +++++
 .../tasks/96-tasks-evidence-view/evidence/V1.md    | 126 +++++++++
 .../tasks/96-tasks-evidence-view/evidence/V2.md    | 287 +++++++++++++++++++++
 .../tasks/96-tasks-evidence-view/evidence/V3.md    |  57 ++++
 .../tasks/96-tasks-evidence-view/evidence/V4.md    | 112 ++++++++
 internal/tui/task_screen.go                        | 140 +++++++++-
 internal/tui/task_screen_test.go                   |  93 +++++++
 testdata/ste-baseline.json                         |   1 +
 8 files changed, 867 insertions(+), 14 deletions(-)
```

The `testdata/ste-baseline.json` change only adds the brief copy with 0.

## V4

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
$ go mod tidy -diff
tidy 0
$ go vet ./...
vet 0
$ golangci-lint run ./...
0 issues.
lint 0
$ go test -race ./...
test 0
```

`gofmt -l` and `go mod tidy -diff` printed nothing. Each `go test` package printed `ok` or `no test files`.

## Repair 1

| Item | Value |
| --- | --- |
| Verifier | t96-verifier-2, Claude Code, claude-opus-5-5 |
| Task commit | `ffcab868` (code fix `32213943`) |
| Base | `origin/dev` `598f2732` |
| Date | 2026-10-20 |
| Verdict | pass |

| ID | Result | Note |
| --- | --- | --- |
| R1 frame test | pass | `TestFileViewsMatchTerminalFrames` fails on `eed52a62` (pre-repair) and passes on the task branch, for Tasks `b`, `v`, `d` and the Friction file view at 120 x 40 and 200 x 57. |
| R1 fix scope | pass | In `friction_screen.go` the only change is the height given to `fileView` (`height-3` to `height-4`). `task_screen.go` has the same change. |
| V1 | pass | The 3 first-pass tests fail on `b4f1d71e` and pass on the task branch. |
| V2 | pass | Each view fills the 207 x 57 pane. The panels reach the bottom border at row 54. |
| V3 | pass | `git diff origin/dev...HEAD --stat` shows only the source table files, the friction file view, and the evidence. |
| V4 | pass | The gate is clean. `git merge-tree --write-tree origin/dev HEAD` exits 0. |

### Frame test

Red run: a detached scratch worktree at `eed52a62` with the task `internal/tui/tasks_friction_test.go` copied in.

```sh
go test ./internal/tui -run '^TestFileViewsMatchTerminalFrames$' -count=1
```

```text
--- FAIL: TestFileViewsMatchTerminalFrames/120x40/Tasks/b
--- FAIL: TestFileViewsMatchTerminalFrames/120x40/Tasks/v
--- FAIL: TestFileViewsMatchTerminalFrames/120x40/Tasks/d
--- FAIL: TestFileViewsMatchTerminalFrames/120x40/Friction/file
--- FAIL: TestFileViewsMatchTerminalFrames/200x57/Tasks/b
--- FAIL: TestFileViewsMatchTerminalFrames/200x57/Tasks/v
--- FAIL: TestFileViewsMatchTerminalFrames/200x57/Tasks/d
--- FAIL: TestFileViewsMatchTerminalFrames/200x57/Friction/file
frame has 41 lines, want 40   (4 times)
frame has 58 lines, want 57   (4 times)
```

Green run on `ffcab868`: all 8 subtests `PASS`, `ok github.com/Harrison-Blair/hive/internal/tui`.

The test checks the line count, the display width of each line, and the main panel bottom border on line `height-3`.

### Fix

```diff
-		return state.fileView(contentWidth, height-3)
+		return state.fileView(contentWidth, height-4)
```

This is the only change in `internal/tui/friction_screen.go` (in `frictionScreen.View`, the file-view branch) and in `internal/tui/task_screen.go` for the repair.

### V1

Red run: a detached scratch worktree at `b4f1d71e` with the task `internal/tui/task_screen_test.go` copied in.

```text
--- FAIL: TestTasksScreenShowsEvidencePanelAndDefinitions (0.00s)
--- FAIL: TestTasksScreenVOpensVerifiedEvidence (0.00s)
    task_screen_test.go:52: v did not request VERIFIED.md
--- FAIL: TestTasksScreenDOpensSelectedTaskDependencyGraph (0.00s)
    task_screen_test.go:76: d did not log opening the dependency graph
```

Green run on `ffcab868`: the 3 tests `PASS`.

### V2

Scratch home in the verifier scratchpad: the `.hive/README.md` marker, project `hive` with its path set to the scratch home and no workspace, the task records 88, 89, 96, 97, 98, and 99 copied from the home, a copy of `.hive/spec/hive-tui` (with the task-branch 96 folder), and a copy of `reference/dogfood/friction`. A temp `XDG_CONFIG_HOME`. The binary was built from `ffcab868`. herdr 0.9.5.

Before the run, workspace `wR2` had only tab `wR2:t1`. The verifier did not use workspace `wBC`.

```sh
herdr tab create --workspace wR2 --cwd <scratch>/v2/bin --label t96-verifier-r1 --env HIVE_HOME=<scratch>/v2/home --env XDG_CONFIG_HOME=<scratch>/v2/config --no-focus
# returned tab wR2:t2, pane wR2:p2; 57 rows
herdr pane run wR2:p2 'env -u HERDR_ENV -u HERDR_PANE_ID -u HERDR_WORKSPACE_ID -u HERDR_SOCKET_PATH -u HERDR_TAB_ID ./hive'
# started pid 1755211; /proc/1755211/environ has the scratch HIVE_HOME and XDG_CONFIG_HOME, and no HERDR_ENV
herdr pane send-keys wR2:p2 5          # Tasks; then b, v, d, each closed with Escape
herdr pane send-keys wR2:p2 6 ; herdr pane send-keys wR2:p2 Enter   # Friction file view
herdr pane read wR2:p2 --source visible --lines 80
```

| Capture | Lines | Main panel bottom border | Content |
| --- | --- | --- | --- |
| Tasks | 57 | row 54 | `task details` panel and `┤ evidence ├` panel, both to row 54; note `landed = delivery commit set and pushed · verified = verification passed` |
| `b` | 57 | row 54 | `brief.md` of 96 |
| `v` | 57 | row 54 | `VERIFIED.md` of 96 |
| `d` | 57 | row 54 | `dependency graph · 96-tasks-evidence-view`, `96-tasks-evidence-view ← 88-tui-terminal-io (needs)`, `blocks: -` |
| Friction file | 57 | row 54 | `agent-can-send-user-chat.md` front matter and summary |

Rows 55-57 of each capture are the status line, a blank row, and the key help. The TUI printed `not in herdr: no coordinator or supervisor started`. No Chat input was opened.

`v` capture (rows 1-6, cut at 150 columns):

```text
⬢ hive · 1 project · 0 live agents · working unknown / idle unknown · supervisor stale · needs_user 0 · 12:36
╭────────────────────────────╮╭───────────────────
│ SCREENS                    ││ Tasks
│   1 Chat                   ││ VERIFIED.md
│   2 Agents                 ││ # 96-tasks-evidence-view VERIFIED
│   3 Goals                  ││
```

Cleanup: two Ctrl+C presses stopped pid 1755211. The verifier closed only tab `wR2:t2` (pane `wR2:p2`). After that, workspace `wR2` had only `wR2:t1`.

### V3

`origin/dev` moved to `598f2732` (task 91) after the task merge. The two-dot diff now also shows the dev change in reverse, so this check uses the merge base.

```text
$ git diff origin/dev...HEAD --stat
 .hive/spec/hive-tui/tasks/96-tasks-evidence-view/brief.md, evidence/ (V1-V4, V2-*.txt, VERIFIED.md)
 internal/tui/friction_screen.go     |   2 +-
 internal/tui/task_screen.go         | 142 +++++++++--
 internal/tui/task_screen_test.go    |  93 ++++++++
 internal/tui/tasks_friction_test.go |  68 ++++++
 testdata/ste-baseline.json          |   1 +
 16 files changed, 1056 insertions(+), 16 deletions(-)
```

`testdata/ste-baseline.json` adds only the 96 brief copy with 0.

### V4

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
$ go mod tidy -diff
$ go vet ./...          exit 0
$ golangci-lint run ./...
0 issues.
$ go test -race ./...   exit 0, 40 packages ok
$ git merge-tree --write-tree origin/dev HEAD   exit 0
```

The verifier also ran `go vet ./...` and `go test -race ./...` on a scratch merge of `origin/dev` `598f2732` and `ffcab868`. Both exit 0.
