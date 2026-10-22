# Independent verification: 45-selected-card-boards

Verdict: pass.

Verified after merge commit `5301832929008f8369d51686a7534633ce79aed5`.
Merged dev at `86f48a68b932bc2d15005c4cef151ed5acba2549`. The merge kept
both sides in README.md and ste-baseline.json. A final `git merge dev`
reported `Already up to date`.

| Item | Result | Observation |
| --- | --- | --- |
| V1 Board | pass | All three cmp commands exit 0. All three PNG pairs are identical. No capture shows `{{`. |
| V2 Grid diff | pass | cellgrid.ReadGrid and cellgrid.Compare find exactly 52 changes for each target board. Each coordinate, old rune, and new rune agrees with the independent report. Each changed cell stays accent and not bold. All other grids have zero changes. |
| V-gate Diff scope and gate | pass | Only allowed files and evidence differ from dev. The five gate commands exit 0. The first two print nothing. |
| AC3 red and green | pass | Removal of new symbols makes TestCheckAcceptsTheHeavyBorder fail at ┏. Restoration makes the check and width tests pass. All three lists agree. |

## V1 capture and plain-text check

For each of Main, AgentsMany, and CompactAgents, ran:

```sh
cmp /home/penguin/source/hive/plans/tui-v2/design/BOARD.dc.html .hive/spec/tui-v2/design/BOARD.dc.html
firefox --headless --no-remote --profile SCRATCH --window-size=1220,780 --screenshot OUT.png file://BOARD.dc.html
```

Used separate new profiles for the six captures under `.hive/tmp/verify45/`.
The first attempt failed because the profile folder did not exist. Created
those folders and repeated the command successfully. Read each copy capture
and compared it byte for byte with the approved board capture.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Main selected manager | `◐ manager-hive`, `┏━┓┃┗┛` | matches | Identical board and PNG |
| Main connector | `┯` | matches | Identical board and PNG |
| AgentsMany selected worker | `● t5-worker-1`, `┏━┓┃┗┛` | matches | Identical board and PNG |
| AgentsMany connector | `┨` | matches | Identical board and PNG |
| CompactAgents selected manager | `◐ manager-hive`, `┏━┓┃┗┛` | matches | Identical board and PNG |
| CompactAgents connector | `┯` | matches | Identical board and PNG |
| All boards: header, tabs, project groups, other cards, footer | Text labels, `● working`, `◐ idle`, `○ stopped` | matches | Full PNG pairs identical; no screen changes in this task |

Removed token and bold information by reading only the text rows of each grid.
The plain text keeps each selected border and state glyph:

```text
Main: ◐ manager-hive; heavy border ┏━┓┃┗┛; joint ┯; no placeholder
AgentsMany: ● t5-worker-1; heavy border ┏━┓┃┗┛; joint ┨; no placeholder
CompactAgents: ◐ manager-hive; heavy border ┏━┓┃┗┛; joint ┯; no placeholder
```

## AC3 source check

Read Card and Panel on `task/33-screen-agents` without merging that branch.
Card sets Heavy when selected. Panel uses lipgloss.ThickBorder. The screen
uses the mixed joints `┯ ┷ ┡ ┨` at lines 295, 325, 333, and 347.
The three symbol lists agree, including all four heavy corners and four joints:

```text
─│╭╮╰╯├┤┼┴┬└━┃┏┓┗┛┯┷┡┨●◐○▶✓•·–—…↑↓←→›█×
```

Temporarily removed `┏┓┗┛┯┷┡┨` from check.go. Ran the existing new check
without changing its expectation. Restored check.go byte for byte before
the green run and gate. No source modification remains from this check.

## Red: go test ./internal/boardexport/ -run Heavy -v (exit 1)

```text
=== RUN   TestCheckAcceptsTheHeavyBorder
    check_test.go:52: board TooSmall: row 0 col 0: symbol "┏" is not in the list of decision 13
--- FAIL: TestCheckAcceptsTheHeavyBorder (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/boardexport	0.002s
FAIL
```

## Green: go test ./internal/boardexport/ ./internal/tui/component/ -v -run Heavy\|SymbolWidth (exit 0)

```text
=== RUN   TestCheckAcceptsTheHeavyBorder
--- PASS: TestCheckAcceptsTheHeavyBorder (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/boardexport	(cached)
=== RUN   TestSymbolWidth
--- PASS: TestSymbolWidth (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/tui/component	0.020s
```

## V2 cell-by-cell output (exit 0)

```text
AgentDetails: 0 changed cells
AgentsMany: 52 changed cells
row 12 col 5: want '╭' accent normal, got '┏' accent normal
row 12 col 6: want '─' accent normal, got '━' accent normal
row 12 col 7: want '─' accent normal, got '━' accent normal
row 12 col 8: want '─' accent normal, got '━' accent normal
row 12 col 9: want '─' accent normal, got '━' accent normal
row 12 col 10: want '─' accent normal, got '━' accent normal
row 12 col 11: want '─' accent normal, got '━' accent normal
row 12 col 12: want '─' accent normal, got '━' accent normal
row 12 col 13: want '─' accent normal, got '━' accent normal
row 12 col 14: want '─' accent normal, got '━' accent normal
row 12 col 15: want '─' accent normal, got '━' accent normal
row 12 col 16: want '─' accent normal, got '━' accent normal
row 12 col 17: want '─' accent normal, got '━' accent normal
row 12 col 18: want '─' accent normal, got '━' accent normal
row 12 col 19: want '─' accent normal, got '━' accent normal
row 12 col 20: want '─' accent normal, got '━' accent normal
row 12 col 21: want '─' accent normal, got '━' accent normal
row 12 col 22: want '─' accent normal, got '━' accent normal
row 12 col 23: want '─' accent normal, got '━' accent normal
row 12 col 24: want '─' accent normal, got '━' accent normal
row 12 col 25: want '─' accent normal, got '━' accent normal
row 12 col 26: want '─' accent normal, got '━' accent normal
row 12 col 27: want '─' accent normal, got '━' accent normal
row 12 col 28: want '╮' accent normal, got '┓' accent normal
row 13 col 5: want '┤' accent normal, got '┨' accent normal
row 13 col 28: want '│' accent normal, got '┃' accent normal
row 14 col 5: want '│' accent normal, got '┃' accent normal
row 14 col 28: want '│' accent normal, got '┃' accent normal
row 15 col 5: want '╰' accent normal, got '┗' accent normal
row 15 col 6: want '─' accent normal, got '━' accent normal
row 15 col 7: want '─' accent normal, got '━' accent normal
row 15 col 8: want '─' accent normal, got '━' accent normal
row 15 col 9: want '─' accent normal, got '━' accent normal
row 15 col 10: want '─' accent normal, got '━' accent normal
row 15 col 11: want '─' accent normal, got '━' accent normal
row 15 col 12: want '─' accent normal, got '━' accent normal
row 15 col 13: want '─' accent normal, got '━' accent normal
row 15 col 14: want '─' accent normal, got '━' accent normal
row 15 col 15: want '─' accent normal, got '━' accent normal
row 15 col 16: want '─' accent normal, got '━' accent normal
row 15 col 17: want '─' accent normal, got '━' accent normal
row 15 col 18: want '─' accent normal, got '━' accent normal
row 15 col 19: want '─' accent normal, got '━' accent normal
row 15 col 20: want '─' accent normal, got '━' accent normal
row 15 col 21: want '─' accent normal, got '━' accent normal
row 15 col 22: want '─' accent normal, got '━' accent normal
row 15 col 23: want '─' accent normal, got '━' accent normal
row 15 col 24: want '─' accent normal, got '━' accent normal
row 15 col 25: want '─' accent normal, got '━' accent normal
row 15 col 26: want '─' accent normal, got '━' accent normal
row 15 col 27: want '─' accent normal, got '━' accent normal
row 15 col 28: want '╯' accent normal, got '┛' accent normal
Chat: 0 changed cells
Compact: 0 changed cells
CompactAgents: 52 changed cells
row 7 col 28: want '╭' accent normal, got '┏' accent normal
row 7 col 29: want '─' accent normal, got '━' accent normal
row 7 col 30: want '─' accent normal, got '━' accent normal
row 7 col 31: want '─' accent normal, got '━' accent normal
row 7 col 32: want '─' accent normal, got '━' accent normal
row 7 col 33: want '─' accent normal, got '━' accent normal
row 7 col 34: want '─' accent normal, got '━' accent normal
row 7 col 35: want '─' accent normal, got '━' accent normal
row 7 col 36: want '─' accent normal, got '━' accent normal
row 7 col 37: want '─' accent normal, got '━' accent normal
row 7 col 38: want '─' accent normal, got '━' accent normal
row 7 col 39: want '─' accent normal, got '━' accent normal
row 7 col 40: want '─' accent normal, got '━' accent normal
row 7 col 41: want '─' accent normal, got '━' accent normal
row 7 col 42: want '─' accent normal, got '━' accent normal
row 7 col 43: want '─' accent normal, got '━' accent normal
row 7 col 44: want '─' accent normal, got '━' accent normal
row 7 col 45: want '─' accent normal, got '━' accent normal
row 7 col 46: want '─' accent normal, got '━' accent normal
row 7 col 47: want '─' accent normal, got '━' accent normal
row 7 col 48: want '─' accent normal, got '━' accent normal
row 7 col 49: want '─' accent normal, got '━' accent normal
row 7 col 50: want '─' accent normal, got '━' accent normal
row 7 col 51: want '╮' accent normal, got '┓' accent normal
row 8 col 28: want '│' accent normal, got '┃' accent normal
row 8 col 51: want '│' accent normal, got '┃' accent normal
row 9 col 28: want '│' accent normal, got '┃' accent normal
row 9 col 51: want '│' accent normal, got '┃' accent normal
row 10 col 28: want '╰' accent normal, got '┗' accent normal
row 10 col 29: want '─' accent normal, got '━' accent normal
row 10 col 30: want '─' accent normal, got '━' accent normal
row 10 col 31: want '─' accent normal, got '━' accent normal
row 10 col 32: want '─' accent normal, got '━' accent normal
row 10 col 33: want '─' accent normal, got '━' accent normal
row 10 col 34: want '─' accent normal, got '━' accent normal
row 10 col 35: want '─' accent normal, got '━' accent normal
row 10 col 36: want '─' accent normal, got '━' accent normal
row 10 col 37: want '─' accent normal, got '━' accent normal
row 10 col 38: want '─' accent normal, got '━' accent normal
row 10 col 39: want '─' accent normal, got '━' accent normal
row 10 col 40: want '┬' accent normal, got '┯' accent normal
row 10 col 41: want '─' accent normal, got '━' accent normal
row 10 col 42: want '─' accent normal, got '━' accent normal
row 10 col 43: want '─' accent normal, got '━' accent normal
row 10 col 44: want '─' accent normal, got '━' accent normal
row 10 col 45: want '─' accent normal, got '━' accent normal
row 10 col 46: want '─' accent normal, got '━' accent normal
row 10 col 47: want '─' accent normal, got '━' accent normal
row 10 col 48: want '─' accent normal, got '━' accent normal
row 10 col 49: want '─' accent normal, got '━' accent normal
row 10 col 50: want '─' accent normal, got '━' accent normal
row 10 col 51: want '╯' accent normal, got '┛' accent normal
CompactChat: 0 changed cells
CompactFriction: 0 changed cells
CompactHealth: 0 changed cells
CompactLogs: 0 changed cells
Components: 0 changed cells
Friction: 0 changed cells
FrictionResolved: 0 changed cells
Health: 0 changed cells
HealthAgents: 0 changed cells
Help: 0 changed cells
Logs: 0 changed cells
Main: 52 changed cells
row 7 col 48: want '╭' accent normal, got '┏' accent normal
row 7 col 49: want '─' accent normal, got '━' accent normal
row 7 col 50: want '─' accent normal, got '━' accent normal
row 7 col 51: want '─' accent normal, got '━' accent normal
row 7 col 52: want '─' accent normal, got '━' accent normal
row 7 col 53: want '─' accent normal, got '━' accent normal
row 7 col 54: want '─' accent normal, got '━' accent normal
row 7 col 55: want '─' accent normal, got '━' accent normal
row 7 col 56: want '─' accent normal, got '━' accent normal
row 7 col 57: want '─' accent normal, got '━' accent normal
row 7 col 58: want '─' accent normal, got '━' accent normal
row 7 col 59: want '─' accent normal, got '━' accent normal
row 7 col 60: want '─' accent normal, got '━' accent normal
row 7 col 61: want '─' accent normal, got '━' accent normal
row 7 col 62: want '─' accent normal, got '━' accent normal
row 7 col 63: want '─' accent normal, got '━' accent normal
row 7 col 64: want '─' accent normal, got '━' accent normal
row 7 col 65: want '─' accent normal, got '━' accent normal
row 7 col 66: want '─' accent normal, got '━' accent normal
row 7 col 67: want '─' accent normal, got '━' accent normal
row 7 col 68: want '─' accent normal, got '━' accent normal
row 7 col 69: want '─' accent normal, got '━' accent normal
row 7 col 70: want '─' accent normal, got '━' accent normal
row 7 col 71: want '╮' accent normal, got '┓' accent normal
row 8 col 48: want '│' accent normal, got '┃' accent normal
row 8 col 71: want '│' accent normal, got '┃' accent normal
row 9 col 48: want '│' accent normal, got '┃' accent normal
row 9 col 71: want '│' accent normal, got '┃' accent normal
row 10 col 48: want '╰' accent normal, got '┗' accent normal
row 10 col 49: want '─' accent normal, got '━' accent normal
row 10 col 50: want '─' accent normal, got '━' accent normal
row 10 col 51: want '─' accent normal, got '━' accent normal
row 10 col 52: want '─' accent normal, got '━' accent normal
row 10 col 53: want '─' accent normal, got '━' accent normal
row 10 col 54: want '─' accent normal, got '━' accent normal
row 10 col 55: want '─' accent normal, got '━' accent normal
row 10 col 56: want '─' accent normal, got '━' accent normal
row 10 col 57: want '─' accent normal, got '━' accent normal
row 10 col 58: want '─' accent normal, got '━' accent normal
row 10 col 59: want '─' accent normal, got '━' accent normal
row 10 col 60: want '┬' accent normal, got '┯' accent normal
row 10 col 61: want '─' accent normal, got '━' accent normal
row 10 col 62: want '─' accent normal, got '━' accent normal
row 10 col 63: want '─' accent normal, got '━' accent normal
row 10 col 64: want '─' accent normal, got '━' accent normal
row 10 col 65: want '─' accent normal, got '━' accent normal
row 10 col 66: want '─' accent normal, got '━' accent normal
row 10 col 67: want '─' accent normal, got '━' accent normal
row 10 col 68: want '─' accent normal, got '━' accent normal
row 10 col 69: want '─' accent normal, got '━' accent normal
row 10 col 70: want '─' accent normal, got '━' accent normal
row 10 col 71: want '╯' accent normal, got '┛' accent normal
TooSmall: 0 changed cells
Work: 0 changed cells
WorkGoals: 0 changed cells
WorkQueue: 0 changed cells
```

## V-gate: git diff dev...HEAD --stat

```text
 .hive/spec/tui-v2/README.md                        |   2 +-
 .hive/spec/tui-v2/design/AgentsMany.dc.html        |   2 +-
 .hive/spec/tui-v2/design/AgentsMany.grid           |   8 +-
 .hive/spec/tui-v2/design/CompactAgents.dc.html     |   2 +-
 .hive/spec/tui-v2/design/CompactAgents.grid        |   8 +-
 .hive/spec/tui-v2/design/Main.dc.html              |   2 +-
 .hive/spec/tui-v2/design/Main.grid                 |   8 +-
 .../tui-v2/tasks/45-selected-card-boards/brief.md  | 123 +++++++++++++
 .../45-selected-card-boards/evidence/V-gate.md     | 165 +++++++++++++++++
 .../tasks/45-selected-card-boards/evidence/V1.md   |  47 +++++
 .../tasks/45-selected-card-boards/evidence/V2.md   | 200 +++++++++++++++++++++
 internal/boardexport/check.go                      |   2 +-
 internal/boardexport/check_test.go                 |   7 +
 internal/tui/component/symbols_test.go             |   6 +-
 testdata/ste-baseline.json                         |   1 +
 15 files changed, 563 insertions(+), 20 deletions(-)
```

## Full gate output

```text
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
ok  	github.com/Harrison-Blair/hive	18.365s
ok  	github.com/Harrison-Blair/hive/cmd	6.251s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.371s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.465s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.494s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.534s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.312s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.005s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.037s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.478s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.252s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.362s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.240s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	13.254s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.576s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.986s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	(cached)
ok  	github.com/Harrison-Blair/hive/internal/homeinit	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/land	3.470s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.157s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.274s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/log	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.016s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.220s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/store	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/logs	(cached)
ok  	github.com/Harrison-Blair/hive/internal/memory	(cached)
ok  	github.com/Harrison-Blair/hive/internal/project	(cached)
ok  	github.com/Harrison-Blair/hive/internal/supervise	(cached)
ok  	github.com/Harrison-Blair/hive/internal/task	9.544s
ok  	github.com/Harrison-Blair/hive/internal/tui	47.736s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	17.969s
ok  	github.com/Harrison-Blair/hive/internal/worktree	4.855s
exit 0
```

The exporter also ran with `go run ./internal/boardexport/cmd/boardexport .hive/spec/tui-v2/design` (exit 0). No grid changed after the export. Scratch scripts, profiles, and captures were removed after the result was recorded.
