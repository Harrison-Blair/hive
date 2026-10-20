# VERIFIED — 97-friction-detail-panel

Verifier: independent agent. Task commit `ce68ce2f` on `task/97-friction-detail-panel`. Base `origin/dev` at `ef2dde2e` (the merge base).

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | `TestFrictionScreenShowsSelectedDetailsInRightPanel` fails on the base at 120x40 and 200x57 ("list column 32, detail columns [47 32]": details are below the list). It passes on the task branch with `TestFrictionScreenLoadsEntriesAndOpensSelectedEntry` and `TestFramesMatchTerminalSizesOnEveryScreen` (all sizes, Friction included). |
| V2 | pass | Live in herdr at 206x57 on a scratch home: the details of the selected entry show in a bordered panel on the right of the list, as on the board. The panel follows the selection. The `decision` column stays (F1 is skipped by decision). Frames are exactly w x h at 120x40, 200x57, and 206x57. |
| V3 | pass | `git diff origin/dev --stat`: only `internal/tui/friction_screen.go`, `internal/tui/tasks_friction_test.go` (the Friction tests), `testdata/ste-baseline.json` (+1 line for the brief copy at 0), the brief copy, and V1-V4 evidence. |
| V4 | pass | `gofmt -l` and `go mod tidy -diff` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` "0 issues." exit 0. `go test -race ./...` exit 0. |

## V1

Red run: a scratch `git worktree` at `origin/dev` with the task test file copied in.

```text
--- FAIL: TestFrictionScreenShowsSelectedDetailsInRightPanel (0.00s)
    --- FAIL: .../120x40: selected entry details are not in a right-hand panel (list column 32, detail columns [47 32])
    --- FAIL: .../200x57: selected entry details are not in a right-hand panel (list column 32, detail columns [47 32])
FAIL	github.com/Harrison-Blair/hive/internal/tui
```

Green run on the task branch:

```text
--- PASS: TestFrictionScreenShowsSelectedDetailsInRightPanel (0.00s)   (120x40, 200x57)
--- PASS: TestFrictionScreenLoadsEntriesAndOpensSelectedEntry (0.00s)
--- PASS: TestFramesMatchTerminalSizesOnEveryScreen (0.07s)   (Friction 80x24 120x40 200x57 207x57 103x57 57x20 300x80)
ok  	github.com/Harrison-Blair/hive/internal/tui	0.076s
```

Extra check (scratch worktree at HEAD, temporary test, not committed): with a loaded entry with a long title and decision, each line of `View` has `ansi.StringWidth` equal to w and the row count equals h at 120x40, 200x57, and 206x57.

## V2

Setup: `go build` of the task code. Scratch home with `.hive/README.md` and two open friction entries, set with `HIVE_HOME`. A temp `XDG_CONFIG_HOME`. `HERDR_ENV=0`. herdr workspace `verify-97` (`wQ9`), pane 206x57. Key `6` selects Friction. I did not press `i` on Chat or Enter on Chat. Quit with ctrl+c twice.

Board (`Friction.dc.html`): list on the left (opened, issue, goal), details of the selected entry in a panel on the right (title, status, opened, file, summary, reproduction). The task matches this layout. The list keeps `decision` in place of `goal` (F1 skipped).

`herdr pane read wQ9:p1 --source visible --lines 57`, first 16 rows, columns 31 and after:

```text
ve agents · working unknown / idle unknown · supervisor stale · needs_user 0 · 11:53
───────────────────╮╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
 ││ Friction                                                                                                                                                                     │
 ││ [open 2]  resolved 0                                                                                                   ╭──────────────────────────────────────────────────╮  │
 ││ opened       title                                      decision                                                       │                                                  │  │
 ││ > 2026-10-20   Verifier panel entry                       Keep the decision column.                                    │ Verifier panel entry                             │  │
 ││   2026-10-19   Second open entry                          -                                                            │ status: Open                                     │  │
 ││                                                                                                                        │ opened: 2026-10-20                               │  │
   ││                                                                                                                        │ file: verifier-panel.md                          │  │
 ││                                                                                                                        │ Summary                                          │  │
 ││                                                                                                                        │ Details of the selected entry show in the right  │  │
 ││                                                                                                                        │ panel.                                           │  │
 ││                                                                                                                        │ Reproduction steps                               │  │
 ││                                                                                                                        │ 1. Open Friction.                                │  │
 ││                                                                                                                        │ 2. Select the entry.                             │  │
 ││                                                                                                                        │                                                  │  │
```

After Down, the panel shows the second entry:

```text
 ││ Friction                                                                                                                                                                     │
 ││ [open 2]  resolved 0                                                                                                   ╭──────────────────────────────────────────────────╮  │
 ││ opened       title                                      decision                                                       │                                                  │  │
 ││   2026-10-20   Verifier panel entry                       Keep the decision column.                                    │ Second open entry                                │  │
 ││ > 2026-10-19   Second open entry                          -                                                            │ status: Open                                     │  │
 ││                                                                                                                        │ opened: 2026-10-19                               │  │
   ││                                                                                                                        │ file: second.md                                  │  │
```

Line widths of the 57-row capture: 53 frame rows at 206, the status and help rows shorter (trailing spaces trimmed).

Exact sizes: herdr clamps a split at ratio 0.9, and `stty` in a pane does not change the herdr grid, so it gave false offsets. For 120x40 and 200x57 I used a private detached tmux server (`tmux -L verify97 new-session -x W -y H`) with the same binary, home, and config. `capture-pane` gave 40 rows with widest row 120, and 57 rows with widest row 200. Each border row ends at column w. (Row 2 contains a TAB that the renderer uses to move the cursor; the cell at column w is the border.) The 120x40 capture:

```text
╭────────────────────────────╮╭────────────────────────────────────────────────────────────────────────────────────────╮
│ SCREENS                    ││ Friction                                                                                  │
│   1 Chat                   ││ [open 2]  resolved 0                                       ╭────────────────────────╮  │
│   2 Agents                 ││ opened       title                        decision         │                        │  │
│   3 Goals                  ││ > 2026-10-20   Verifier panel entry         Keep the decis…│ Verifier panel entry   │  │
│   4 Queue                  ││   2026-10-19   Second open entry            -              │ status: Open           │  │
│   5 Tasks                  ││                                                            │ opened: 2026-10-20     │  │
```

Note, not a task defect: the tmux sessions started with the worktree as working directory, so the startup Chat guard showed `chat send is for the user only: ... rule 2 (working directory is below projects/)` in the status area. The scratch home log has no chat send and no Chat input action.

Cleanup: closed herdr pane `wQ9:p2` (split test) and workspace `wQ9`. The tmux server `verify97` exited. Removed the two scratch worktrees.

## V3

```text
 .../tasks/97-friction-detail-panel/brief.md        | 64 ++++++++++++++++++++++
 .../tasks/97-friction-detail-panel/evidence/V1.md  | 18 ++++++
 .../tasks/97-friction-detail-panel/evidence/V2.md  | 59 ++++++++++++++++++++
 .../tasks/97-friction-detail-panel/evidence/V3.md  |  5 ++
 .../tasks/97-friction-detail-panel/evidence/V4.md  | 16 ++++++
 internal/tui/friction_screen.go                    | 53 +++++++++++++++++-
 internal/tui/tasks_friction_test.go                | 55 +++++++++++++++++++
 testdata/ste-baseline.json                         |  1 +
 8 files changed, 270 insertions(+), 1 deletion(-)
```

## V4

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')     (no output, exit 0)
$ go mod tidy -diff                                          (no output, exit 0)
$ go vet ./...                                               exit 0
$ golangci-lint run ./...                                    0 issues. exit 0
$ go test -race ./...                                        all ok, exit 0
```

## Repair 1

Verifier: claude-opus-5-5, 2026-10-20. Repair commit `01741b56` (merge of
`c003eabc` and dev `2924385e`). The TUI was not run.

| Item | Result | Observation |
| --- | --- | --- |
| Baseline rows of both sides | pass | `git diff 2924385e HEAD -- testdata/ste-baseline.json` adds only the `97-friction-detail-panel` row. `git diff c003eabc HEAD` adds only the dev rows `93-ci-test-fixes`, `98-health-panels`, and `100-no-backgrounds`. No row is removed. Order and format are kept. |
| `git merge-tree --write-tree origin/dev HEAD` | pass | Prints `6cbf602e0c9f5fe4e073120ab49cfbea0e2f139b`, exit 0 (origin/dev `7bf0387b`). |
| Source files change only by the merge | pass | `git diff --stat 2924385e HEAD -- ':!.hive'` shows only `internal/tui/friction_screen.go`, `internal/tui/tasks_friction_test.go`, and `testdata/ste-baseline.json` (+1 row), which are the task changes. Each other non-spec file that differs from `c003eabc` comes from dev. |
| Brief copy STE rewording | pass | The V2 row splits into short sentences. The steps (build, temp marked home with fixture data, `HIVE_HOME`, temp `XDG_CONFIG_HOME`, herdr tab at about 200 x 57, `herdr pane read` per gap) and the expected result are kept. The Repair 1 section is added. |
| Gate | pass | `gofmt -l` no output; `go mod tidy -diff` no output; `go vet ./...` exit 0; `golangci-lint run ./...` 0 issues; `go test -race ./...` all ok, exit 0. |

Note: `git diff origin/dev --stat` also shows `reference/dogfood/friction/` changes. These come from dev commits `1320b605` and `7bf0387b` after the merge base. The task branch does not change these files (`git diff --stat $(git merge-base HEAD origin/dev) HEAD` has no file in that folder).
