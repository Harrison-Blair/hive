# Independent verification: 41-workgoals-board-fix

Verifier: `t41-verifier-1`. Reviewed commit: `766b45e5`.
Verdict: **pass**.

| Item | Result | Observation |
| --- | --- | --- |
| V1 Board | pass | Source `cmp` and capture `cmp` exit 0. Both PNG files have 95836 bytes. No visible `{{` or removed note. |
| V2 Grid diff | pass | `cellgrid.ReadGrid` and `cellgrid.Compare`, without masks, find exactly 25 changes: row 9, columns 2–26. Each new cell is a space, token `fg`, normal weight. Each other grid and the fixture have no difference from `dev`. |
| V-gate Diff scope and gate | pass | `git merge dev`: Already up to date. Only the permitted source files and evidence differ. All five gate commands exit 0. Formatting and tidy print nothing. |

## V1 captures and visual check

Ran from the worktree root:

```sh
cmp /home/penguin/source/hive/plans/tui-v2/design/WorkGoals.dc.html .hive/spec/tui-v2/design/WorkGoals.dc.html
mkdir -p .hive/tmp/verify41/profile-approved .hive/tmp/verify41/profile-copy
firefox --headless --no-remote --profile "$PWD/.hive/tmp/verify41/profile-approved" --window-size=1220,780 --screenshot "$PWD/.hive/tmp/verify41/approved.png" file:///home/penguin/source/hive/plans/tui-v2/design/WorkGoals.dc.html
firefox --headless --no-remote --profile "$PWD/.hive/tmp/verify41/profile-copy" --window-size=1220,780 --screenshot "$PWD/.hive/tmp/verify41/copy.png" "file://$PWD/.hive/spec/tui-v2/design/WorkGoals.dc.html"
cmp .hive/tmp/verify41/approved.png .hive/tmp/verify41/copy.png
```

Each command exits 0. The first capture attempt failed because the new profile directory did not exist. Created both directories, then captured both boards successfully. Inspected the copy PNG; the two PNG files are byte-identical.

This task changes the board copy, not a running screen. Checked its elements against the approved board. Also printed every grid rune without token or bold data to check states without colour.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header | coordinator working; supervisor live; needs you 0 | matches | Same text in both captures. |
| Main navigation | 3 Work and underline | matches | Same active tab. |
| Work navigation | [Goals]; Queue; Tasks; [ ] switch view | matches | Same selected view and labels. |
| Goals panel | goals · 2; planned 0/5; done 3/3 | matches | Both goals remain. |
| Selected goal | ▶ g1-tui-restart | matches | Selection has a glyph. |
| Removed note | blank row 9, columns 2–26 | matches | User decision in the brief removes the note. |
| Goal detail | g1-tui-restart; status done; project hive; grant; manager-hive | matches | Same labels and values. |
| Waves | A, B, C; 1/1; ✓ done | matches | Each wave has a completion glyph and word. |
| Tasks | A, B, C; task names; merge hashes | matches | Same three rows. |
| Needs you | nothing | matches | State has a word. |
| Detail actions | o goal design; t open task in Tasks | matches | Same key labels. |
| Footer | [list]; move; tab region; view; enter open; design; task; filter; help | matches | Focus region and controls have text. |

## V2 independent cell comparison

A scratch Go command loads `git show dev:.hive/spec/tui-v2/design/WorkGoals.grid` and the task grid with `cellgrid.ReadGrid`. It calls `cellgrid.Compare` with no masks and prints `cellgrid.FormatDiffs`. It checks that exactly 25 cells differ and that every changed cell is in row 9, columns 2–26, with a space, `cellgrid.FG`, and `Bold == false`.

The check exits 0 for the task grid. As a negative control, the same check against the old grid exits 1: `0 changed cells`, `panic: expected 25 changed cells`. No repository source was changed for this control.

`go run ./internal/boardexport/cmd/boardexport .hive/spec/tui-v2/design` exits 0 and leaves the worktree clean. `git diff --name-only dev -- '.hive/spec/tui-v2/design/*.grid'` lists only `WorkGoals.grid`. The fixture diff exits 0. `go test -count=1 ./internal/boardexport/ ./internal/lib/cellgrid/` passes both packages and loads the fixture.

### Changed cells and colour-free board text

```text
25 changed cells
row 9 col 2: want 'g' dim normal, got ' ' fg normal
row 9 col 3: want '2' dim normal, got ' ' fg normal
row 9 col 4: want '-' dim normal, got ' ' fg normal
row 9 col 5: want 't' dim normal, got ' ' fg normal
row 9 col 6: want 'u' dim normal, got ' ' fg normal
row 9 col 7: want 'i' dim normal, got ' ' fg normal
row 9 col 8: want '-' dim normal, got ' ' fg normal
row 9 col 9: want 'v' dim normal, got ' ' fg normal
row 9 col 10: want '2' dim normal, got ' ' fg normal
row 9 col 11: want ' ' dim normal, got ' ' fg normal
row 9 col 12: want 'i' dim normal, got ' ' fg normal
row 9 col 13: want 's' dim normal, got ' ' fg normal
row 9 col 14: want ' ' dim normal, got ' ' fg normal
row 9 col 15: want 'e' dim normal, got ' ' fg normal
row 9 col 16: want 'x' dim normal, got ' ' fg normal
row 9 col 17: want 'a' dim normal, got ' ' fg normal
row 9 col 18: want 'm' dim normal, got ' ' fg normal
row 9 col 19: want 'p' dim normal, got ' ' fg normal
row 9 col 20: want 'l' dim normal, got ' ' fg normal
row 9 col 21: want 'e' dim normal, got ' ' fg normal
row 9 col 22: want ' ' dim normal, got ' ' fg normal
row 9 col 23: want 'd' dim normal, got ' ' fg normal
row 9 col 24: want 'a' dim normal, got ' ' fg normal
row 9 col 25: want 't' dim normal, got ' ' fg normal
row 9 col 26: want 'a' dim normal, got ' ' fg normal
 hive  ~/source/hive                                    coordinator ● working   supervisor ● live   needs you 0   14:22 
  1 Agents   2 Chat   3 Work   4 Friction   5 Health   6 Logs                                                           
─────────────────────━━━━━━━━───────────────────────────────────────────────────────────────────────────────────────────
 [Goals]   Queue    Tasks                                                                             [ ] switch view   
╭─ goals · 2 ──────────────────────────────────────╮╭─ g1-tui-restart ─────────────────────────────────────────────────╮
│   GOAL                    STATUS    WAVES        ││                                                                  │
│   g2-tui-v2               planned   0/5          ││ title     TUI restart and stale coordinator startup              │
│ ▶ g1-tui-restart          done      3/3          ││ status    done                                                   │
│                                                  ││ project   hive                                                   │
│                                                  ││ grant     grants.md 2026-10-18 hive delivery on dev              │
│                                                  ││ manager   manager-hive                                           │
│                                                  ││                                                                  │
│                                                  ││ WAVES                                                            │
│                                                  ││ WAVE TITLE                      DONE WHEN  LANDED  STATUS        │
│                                                  ││ A    Stale coordinator startup  dev        1/1     ✓ done        │
│                                                  ││ B    TUI restart                dev        1/1     ✓ done        │
│                                                  ││ C    Agents graph columns       dev        1/1     ✓ done        │
│                                                  ││                                                                  │
│                                                  ││ TASKS                                                            │
│                                                  ││ A   2-stale-coordinator-startup   merge 0e079272                 │
│                                                  ││ B   3-tui-restart                 merge 3ccd93fb                 │
│                                                  ││ C   4-agents-equal-columns        merge ede94f3e                 │
│                                                  ││                                                                  │
│                                                  ││ NEEDS YOU                                                        │
│                                                  ││ nothing                                                          │
│                                                  ││                                                                  │
│                                                  ││ o goal design   t open task in Tasks                             │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
╰──────────────────────────────────────────────────╯╰──────────────────────────────────────────────────────────────────╯
 [list] ↑↓ move · tab region · [ ] view · enter open · o design · t task · / filter · ? help                            
```

## Full gate output

```text
$ git diff dev...HEAD --stat
 .hive/spec/tui-v2/design/WorkGoals.dc.html         |   1 -
 .hive/spec/tui-v2/design/WorkGoals.grid            |   4 +-
 .../tui-v2/tasks/41-workgoals-board-fix/brief.md   | 101 +++++++++++++++++
 .../41-workgoals-board-fix/evidence/V-gate.md      | 114 +++++++++++++++++++
 .../tasks/41-workgoals-board-fix/evidence/V1.md    |  64 +++++++++++
 .../tasks/41-workgoals-board-fix/evidence/V2.md    | 123 +++++++++++++++++++++
 testdata/ste-baseline.json                         |   1 +
 7 files changed, 405 insertions(+), 3 deletions(-)
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
exit 0
$ go mod tidy -diff
exit 0
$ go vet ./...
exit 0
$ golangci-lint run ./...
0 issues.
exit 0
$ go test -race ./...
ok  	github.com/Harrison-Blair/hive	18.957s
ok  	github.com/Harrison-Blair/hive/cmd	6.918s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.361s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.498s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.609s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.615s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.334s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.226s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.231s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.531s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.266s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.291s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.169s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	13.727s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.469s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	5.043s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	(cached)
ok  	github.com/Harrison-Blair/hive/internal/homeinit	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/land	3.477s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.186s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.291s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/log	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.015s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.272s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/store	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/logs	(cached)
ok  	github.com/Harrison-Blair/hive/internal/memory	(cached)
ok  	github.com/Harrison-Blair/hive/internal/project	(cached)
ok  	github.com/Harrison-Blair/hive/internal/supervise	(cached)
ok  	github.com/Harrison-Blair/hive/internal/task	10.178s
ok  	github.com/Harrison-Blair/hive/internal/tui	56.184s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	(cached)
ok  	github.com/Harrison-Blair/hive/internal/worktree	5.445s
exit 0
```

## Cleanup

Removed `.hive/tmp/verify41/`, including the two PNG files, both Firefox profiles, the scratch Go command, the old grid, and the scratch logs. Confirmed that this folder no longer exists. Only this report is added by the verifier.
