# Verification: 10-board-export

- Verifier: `t10-verifier-1` (pi).
- Task commit: `2d32234c`.
- Base: `dev` at `c591a03606705f806708e68d04ce1dc054669398`.
- Verdict: **pass**.

| Item | Result | Observation |
| --- | --- | --- |
| V1 Tests | pass | Each of the 16 new tests fails in an individual red run. All pass on the task code. No test was skipped in the green run. |
| V2 Export again | pass | All 21 exported grids match byte for byte. Main rows 0, 1, 7, 14, and 37 match the text from rows(). |
| V-gate Diff scope and gate | pass | Merge of dev reports already up to date. Only approved source paths and evidence differ. All five commands exit 0. The first two print nothing. |

## Review

The exporter reads cells from T. Its synthetic-board test returns empty rows(),
so that test checks this requirement independently. The checks use hand-written
grids. They cover invalid symbols, tokens, sizes, ragged rows, and board names.
The missing-Node test checks the error and absence of an output grid.
The exporter commit records the required development-tool reason.

An independent comparison with the README board table confirms all 21 sizes.
The review aid contains the exact text rows of all 21 grids, without the t| prefix.
No board source or live screen changed. The brief excludes screen golden checks.
The PNG screen comparison and state-without-colour table do not apply to this
exporter-only task. Human grid review remains due at the end of wave 2.

## Red-run method

A scratch copy of dev received only the three new test files. The initial run
fails to compile because the new API does not exist. To check behavior rather
than compilation, the individual runs then use the API-only stub printed in
V1.md. The stub has no exporter or validation behavior and allocates zero cells
for ToGrid. For TestCheckAcceptsAGoodBoard only, Check returns an error to prove
that rejection of a valid board fails that test. All other stub functions use
the empty bodies in V1.md. No golden files exist in this base copy.
Each top-level test fails independently, as recorded below. The task code is
unchanged throughout verification.

## Commands and output

### Base compile red

```text
# github.com/Harrison-Blair/hive/internal/boardexport [github.com/Harrison-Blair/hive/internal/boardexport.test]
internal/boardexport/check_test.go:13:26: undefined: Raw
internal/boardexport/check_test.go:15:8: undefined: Raw
internal/boardexport/check_test.go:18:15: undefined: RawCell
internal/boardexport/check_test.go:21:26: undefined: RawCell
internal/boardexport/check_test.go:34:36: undefined: Raw
internal/boardexport/check_test.go:44:12: undefined: Check
internal/boardexport/check_test.go:51:9: undefined: Check
internal/boardexport/check_test.go:57:9: undefined: Check
internal/boardexport/check_test.go:64:9: undefined: Check
internal/boardexport/check_test.go:70:9: undefined: Check
internal/boardexport/check_test.go:70:9: too many errors
FAIL	github.com/Harrison-Blair/hive/internal/boardexport [build failed]
FAIL
```

### Individual behavioral red runs

```text
$ go test -count=1 -run ^TestScriptTakesTheBoardScript$ ./internal/boardexport
--- FAIL: TestScriptTakesTheBoardScript (0.00s)
    export_test.go:46: Script = "", want "var a = 1;"
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/boardexport	0.002s
FAIL
exit: 1
$ go test -count=1 -run ^TestScriptFailsWithoutOneBoardScript$ ./internal/boardexport
--- FAIL: TestScriptFailsWithoutOneBoardScript (0.00s)
    export_test.go:52: Script of a page with no board script: error = nil
    export_test.go:55: Script of a page with two board scripts: error = nil
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/boardexport	0.002s
FAIL
exit: 1
$ go test -count=1 -run ^TestExportDirFailsWithoutNode$ ./internal/boardexport
--- FAIL: TestExportDirFailsWithoutNode (0.00s)
    export_test.go:64: error = nil, want one with ["Node.js"]
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/boardexport	0.002s
FAIL
exit: 1
$ go test -count=1 -run ^TestExportDirWritesTheCellsOfT$ ./internal/boardexport
--- FAIL: TestExportDirWritesTheCellsOfT (0.00s)
    export_test.go:82: open /tmp/TestExportDirWritesTheCellsOfT326806396/002/TooSmall.grid: no such file or directory
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/boardexport	0.002s
FAIL
exit: 1
$ go test -count=1 -run ^TestExportDirFailsABadBoard$ ./internal/boardexport
--- FAIL: TestExportDirFailsABadBoard (0.00s)
    --- FAIL: TestExportDirFailsABadBoard/symbol (0.00s)
        export_test.go:127: error = nil, want one with ["TooSmall" "row 15 col 59" "\"★\""]
    --- FAIL: TestExportDirFailsABadBoard/token (0.00s)
        export_test.go:127: error = nil, want one with ["TooSmall" "row 15 col 59" "\"red\""]
    --- FAIL: TestExportDirFailsABadBoard/size (0.00s)
        export_test.go:127: error = nil, want one with ["TooSmall" "61×16" "60×16"]
    --- FAIL: TestExportDirFailsABadBoard/two_grids (0.00s)
        export_test.go:127: error = nil, want one with ["TooSmall" "2 grids"]
    --- FAIL: TestExportDirFailsABadBoard/script_error (0.00s)
        export_test.go:127: error = nil, want one with ["TooSmall" "boom"]
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/boardexport	0.003s
FAIL
exit: 1
$ go test -count=1 -run ^TestGoldenGridFiles$ ./internal/boardexport
--- FAIL: TestGoldenGridFiles (0.00s)
    golden_test.go:22: found 0 grid files, want 21: []
    --- FAIL: TestGoldenGridFiles/AgentDetails (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/AgentDetails.grid: no such file or directory
    --- FAIL: TestGoldenGridFiles/CompactAgents (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/CompactAgents.grid: no such file or directory
    --- FAIL: TestGoldenGridFiles/WorkGoals (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/WorkGoals.grid: no such file or directory
    --- FAIL: TestGoldenGridFiles/FrictionResolved (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/FrictionResolved.grid: no such file or directory
    --- FAIL: TestGoldenGridFiles/CompactFriction (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/CompactFriction.grid: no such file or directory
    --- FAIL: TestGoldenGridFiles/Logs (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/Logs.grid: no such file or directory
    --- FAIL: TestGoldenGridFiles/CompactLogs (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/CompactLogs.grid: no such file or directory
    --- FAIL: TestGoldenGridFiles/TooSmall (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/TooSmall.grid: no such file or directory
    --- FAIL: TestGoldenGridFiles/AgentsMany (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/AgentsMany.grid: no such file or directory
    --- FAIL: TestGoldenGridFiles/Chat (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/Chat.grid: no such file or directory
    --- FAIL: TestGoldenGridFiles/Work (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/Work.grid: no such file or directory
    --- FAIL: TestGoldenGridFiles/Components (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/Components.grid: no such file or directory
    --- FAIL: TestGoldenGridFiles/Main (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/Main.grid: no such file or directory
    --- FAIL: TestGoldenGridFiles/CompactChat (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/CompactChat.grid: no such file or directory
    --- FAIL: TestGoldenGridFiles/WorkQueue (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/WorkQueue.grid: no such file or directory
    --- FAIL: TestGoldenGridFiles/Compact (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/Compact.grid: no such file or directory
    --- FAIL: TestGoldenGridFiles/HealthAgents (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/HealthAgents.grid: no such file or directory
    --- FAIL: TestGoldenGridFiles/Help (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/Help.grid: no such file or directory
    --- FAIL: TestGoldenGridFiles/Friction (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/Friction.grid: no such file or directory
    --- FAIL: TestGoldenGridFiles/Health (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/Health.grid: no such file or directory
    --- FAIL: TestGoldenGridFiles/CompactHealth (0.00s)
        golden_test.go:28: open ../../.hive/spec/tui-v2/design/CompactHealth.grid: no such file or directory
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/boardexport	0.003s
FAIL
exit: 1
$ go test -count=1 -run ^TestCheckAcceptsAGoodBoard$ ./internal/boardexport
--- FAIL: TestCheckAcceptsAGoodBoard (0.00s)
    check_test.go:45: check stub rejects every grid
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/boardexport	0.002s
FAIL
exit: 1
$ go test -count=1 -run ^TestCheckFailsABadSymbol$ ./internal/boardexport
--- FAIL: TestCheckFailsABadSymbol (0.00s)
    check_test.go:52: error = nil, want one with ["TooSmall" "row 0 col 2" "\"★\""]
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/boardexport	0.002s
FAIL
exit: 1
$ go test -count=1 -run ^TestCheckFailsATokenOutsideThePalette$ ./internal/boardexport
--- FAIL: TestCheckFailsATokenOutsideThePalette (0.00s)
    check_test.go:58: error = nil, want one with ["TooSmall" "row 0 col 2" "\"red\""]
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/boardexport	0.002s
FAIL
exit: 1
$ go test -count=1 -run ^TestCheckFailsACellOfTwoRunes$ ./internal/boardexport
--- FAIL: TestCheckFailsACellOfTwoRunes (0.00s)
    check_test.go:65: error = nil, want one with ["TooSmall" "row 3 col 5" "\"ab\""]
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/boardexport	0.002s
FAIL
exit: 1
$ go test -count=1 -run ^TestCheckFailsAWrongSize$ ./internal/boardexport
--- FAIL: TestCheckFailsAWrongSize (0.00s)
    check_test.go:71: error = nil, want one with ["TooSmall" "2×2" "60×16"]
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/boardexport	0.002s
FAIL
exit: 1
$ go test -count=1 -run ^TestCheckFailsARaggedRow$ ./internal/boardexport
--- FAIL: TestCheckFailsARaggedRow (0.00s)
    check_test.go:78: error = nil, want one with ["TooSmall" "row 7" "59"]
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/boardexport	0.002s
FAIL
exit: 1
$ go test -count=1 -run ^TestCheckFailsABoardOutsideTheTable$ ./internal/boardexport
--- FAIL: TestCheckFailsABoardOutsideTheTable (0.00s)
    check_test.go:84: error = nil, want one with ["Extra" "spec table"]
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/boardexport	0.003s
FAIL
exit: 1
$ go test -count=1 -run ^TestCheckGridFailsABadSymbol$ ./internal/boardexport
--- FAIL: TestCheckGridFailsABadSymbol (0.00s)
    check_test.go:93: error = nil, want one with ["TooSmall" "row 2 col 4" "\"★\""]
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/boardexport	0.002s
FAIL
exit: 1
$ go test -count=1 -run ^TestCheckGridFailsAWrongSize$ ./internal/boardexport
--- FAIL: TestCheckGridFailsAWrongSize (0.00s)
    check_test.go:101: error = nil, want one with ["TooSmall" "2×1" "60×16"]
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/boardexport	0.002s
FAIL
exit: 1
$ go test -count=1 -run ^TestToGridKeepsEachCell$ ./internal/boardexport
--- FAIL: TestToGridKeepsEachCell (0.00s)
    check_test.go:125: cell 0 = {Rune:0 Token:fg Bold:false}, want {Rune:9581 Token:border Bold:false}
    check_test.go:125: cell 1 = {Rune:0 Token:fg Bold:false}, want {Rune:97 Token:accent Bold:true}
    check_test.go:125: cell 2 = {Rune:0 Token:fg Bold:false}, want {Rune:120 Token:fg Bold:false}
    check_test.go:125: cell 3 = {Rune:0 Token:fg Bold:false}, want {Rune:32 Token:dim Bold:false}
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/boardexport	0.002s
FAIL
exit: 1
```

### Task green run: go test -count=1 -v ./internal/boardexport/...

```text
=== RUN   TestCheckAcceptsAGoodBoard
--- PASS: TestCheckAcceptsAGoodBoard (0.00s)
=== RUN   TestCheckFailsABadSymbol
--- PASS: TestCheckFailsABadSymbol (0.00s)
=== RUN   TestCheckFailsATokenOutsideThePalette
--- PASS: TestCheckFailsATokenOutsideThePalette (0.00s)
=== RUN   TestCheckFailsACellOfTwoRunes
--- PASS: TestCheckFailsACellOfTwoRunes (0.00s)
=== RUN   TestCheckFailsAWrongSize
--- PASS: TestCheckFailsAWrongSize (0.00s)
=== RUN   TestCheckFailsARaggedRow
--- PASS: TestCheckFailsARaggedRow (0.00s)
=== RUN   TestCheckFailsABoardOutsideTheTable
--- PASS: TestCheckFailsABoardOutsideTheTable (0.00s)
=== RUN   TestCheckGridFailsABadSymbol
--- PASS: TestCheckGridFailsABadSymbol (0.00s)
=== RUN   TestCheckGridFailsAWrongSize
--- PASS: TestCheckGridFailsAWrongSize (0.00s)
=== RUN   TestToGridKeepsEachCell
--- PASS: TestToGridKeepsEachCell (0.00s)
=== RUN   TestScriptTakesTheBoardScript
--- PASS: TestScriptTakesTheBoardScript (0.00s)
=== RUN   TestScriptFailsWithoutOneBoardScript
--- PASS: TestScriptFailsWithoutOneBoardScript (0.00s)
=== RUN   TestExportDirFailsWithoutNode
--- PASS: TestExportDirFailsWithoutNode (0.00s)
=== RUN   TestExportDirWritesTheCellsOfT
--- PASS: TestExportDirWritesTheCellsOfT (0.03s)
=== RUN   TestExportDirFailsABadBoard
=== RUN   TestExportDirFailsABadBoard/symbol
=== RUN   TestExportDirFailsABadBoard/token
=== RUN   TestExportDirFailsABadBoard/size
=== RUN   TestExportDirFailsABadBoard/two_grids
=== RUN   TestExportDirFailsABadBoard/script_error
--- PASS: TestExportDirFailsABadBoard (0.24s)
    --- PASS: TestExportDirFailsABadBoard/symbol (0.04s)
    --- PASS: TestExportDirFailsABadBoard/token (0.04s)
    --- PASS: TestExportDirFailsABadBoard/size (0.05s)
    --- PASS: TestExportDirFailsABadBoard/two_grids (0.06s)
    --- PASS: TestExportDirFailsABadBoard/script_error (0.05s)
=== RUN   TestGoldenGridFiles
=== RUN   TestGoldenGridFiles/FrictionResolved
=== RUN   TestGoldenGridFiles/CompactHealth
=== RUN   TestGoldenGridFiles/Logs
=== RUN   TestGoldenGridFiles/AgentsMany
=== RUN   TestGoldenGridFiles/CompactChat
=== RUN   TestGoldenGridFiles/CompactLogs
=== RUN   TestGoldenGridFiles/AgentDetails
=== RUN   TestGoldenGridFiles/Chat
=== RUN   TestGoldenGridFiles/WorkGoals
=== RUN   TestGoldenGridFiles/Health
=== RUN   TestGoldenGridFiles/HealthAgents
=== RUN   TestGoldenGridFiles/Help
=== RUN   TestGoldenGridFiles/WorkQueue
=== RUN   TestGoldenGridFiles/Compact
=== RUN   TestGoldenGridFiles/Friction
=== RUN   TestGoldenGridFiles/CompactFriction
=== RUN   TestGoldenGridFiles/TooSmall
=== RUN   TestGoldenGridFiles/Components
=== RUN   TestGoldenGridFiles/Main
=== RUN   TestGoldenGridFiles/CompactAgents
=== RUN   TestGoldenGridFiles/Work
--- PASS: TestGoldenGridFiles (0.02s)
    --- PASS: TestGoldenGridFiles/FrictionResolved (0.00s)
    --- PASS: TestGoldenGridFiles/CompactHealth (0.00s)
    --- PASS: TestGoldenGridFiles/Logs (0.00s)
    --- PASS: TestGoldenGridFiles/AgentsMany (0.00s)
    --- PASS: TestGoldenGridFiles/CompactChat (0.00s)
    --- PASS: TestGoldenGridFiles/CompactLogs (0.00s)
    --- PASS: TestGoldenGridFiles/AgentDetails (0.00s)
    --- PASS: TestGoldenGridFiles/Chat (0.00s)
    --- PASS: TestGoldenGridFiles/WorkGoals (0.00s)
    --- PASS: TestGoldenGridFiles/Health (0.00s)
    --- PASS: TestGoldenGridFiles/HealthAgents (0.00s)
    --- PASS: TestGoldenGridFiles/Help (0.00s)
    --- PASS: TestGoldenGridFiles/WorkQueue (0.00s)
    --- PASS: TestGoldenGridFiles/Compact (0.00s)
    --- PASS: TestGoldenGridFiles/Friction (0.00s)
    --- PASS: TestGoldenGridFiles/CompactFriction (0.00s)
    --- PASS: TestGoldenGridFiles/TooSmall (0.00s)
    --- PASS: TestGoldenGridFiles/Components (0.00s)
    --- PASS: TestGoldenGridFiles/Main (0.00s)
    --- PASS: TestGoldenGridFiles/CompactAgents (0.00s)
    --- PASS: TestGoldenGridFiles/Work (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/boardexport	0.298s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
```

### Export and grid comparisons

```text
$ go run ./internal/boardexport/cmd/boardexport -out SCRATCH/out .hive/spec/tui-v2/design
exit: 0
same AgentDetails.grid
same AgentsMany.grid
same Chat.grid
same CompactAgents.grid
same CompactChat.grid
same CompactFriction.grid
same Compact.grid
same CompactHealth.grid
same CompactLogs.grid
same Components.grid
same Friction.grid
same FrictionResolved.grid
same HealthAgents.grid
same Health.grid
same Help.grid
same Logs.grid
same Main.grid
same TooSmall.grid
same WorkGoals.grid
same Work.grid
same WorkQueue.grid
row comparison exit: 0
```

### Five rows from rows(); diff against grid text exits 0

```text
0	 hive  ~/source/hive                                    coordinator ● working   supervisor ● live   needs you 0   14:22 
1	  1 Agents   2 Chat   3 Work   4 Friction   5 Health   6 Logs                                                           
7	│                           │                   ╭──────────────────────╮                   │                           │
14	│                           │     ╭───────────┴──────────╮      ╭───────────┴──────────╮   │                           │
37	│ ── spawn   ● working   ◐ idle   ○ stopped   one column per project, all the same width                               │
```

### git diff dev...HEAD --stat

```text
 .hive/spec/tui-v2/design/AgentDetails.grid         | 121 +++
 .hive/spec/tui-v2/design/AgentsMany.grid           | 121 +++
 .hive/spec/tui-v2/design/Chat.grid                 | 121 +++
 .hive/spec/tui-v2/design/Compact.grid              |  73 ++
 .hive/spec/tui-v2/design/CompactAgents.grid        |  73 ++
 .hive/spec/tui-v2/design/CompactChat.grid          |  73 ++
 .hive/spec/tui-v2/design/CompactFriction.grid      |  73 ++
 .hive/spec/tui-v2/design/CompactHealth.grid        |  73 ++
 .hive/spec/tui-v2/design/CompactLogs.grid          |  73 ++
 .hive/spec/tui-v2/design/Components.grid           | 202 +++++
 .hive/spec/tui-v2/design/Friction.grid             | 121 +++
 .hive/spec/tui-v2/design/FrictionResolved.grid     | 121 +++
 .hive/spec/tui-v2/design/Health.grid               | 121 +++
 .hive/spec/tui-v2/design/HealthAgents.grid         | 121 +++
 .hive/spec/tui-v2/design/Help.grid                 | 121 +++
 .hive/spec/tui-v2/design/Logs.grid                 | 121 +++
 .hive/spec/tui-v2/design/Main.grid                 | 121 +++
 .hive/spec/tui-v2/design/TooSmall.grid             |  49 ++
 .hive/spec/tui-v2/design/Work.grid                 | 121 +++
 .hive/spec/tui-v2/design/WorkGoals.grid            | 121 +++
 .hive/spec/tui-v2/design/WorkQueue.grid            | 121 +++
 .hive/spec/tui-v2/tasks/10-board-export/brief.md   | 127 +++
 .../tasks/10-board-export/evidence/V-gate.md       | 153 ++++
 .../tui-v2/tasks/10-board-export/evidence/V1.md    | 303 +++++++
 .../tui-v2/tasks/10-board-export/evidence/V2.md    |  79 ++
 .../tasks/10-board-export/evidence/grids-review.md | 897 +++++++++++++++++++++
 internal/boardexport/check.go                      | 141 ++++
 internal/boardexport/check_test.go                 | 140 ++++
 internal/boardexport/cmd/boardexport/main.go       |  30 +
 internal/boardexport/doc.go                        |   7 +
 internal/boardexport/export.go                     | 117 +++
 internal/boardexport/export_test.go                | 149 ++++
 internal/boardexport/golden_test.go                |  43 +
 testdata/ste-baseline.json                         |   1 +
 34 files changed, 4449 insertions(+)
```

### Full gate

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
exit: 0
$ go mod tidy -diff
exit: 0
$ go vet ./...
exit: 0
$ golangci-lint run ./...
0 issues.
exit: 0
$ go test -race ./...
ok  	github.com/Harrison-Blair/hive	14.802s
ok  	github.com/Harrison-Blair/hive/cmd	3.294s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.137s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.386s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.275s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.138s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.904s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.918s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.301s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.079s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.101s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.029s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	10.795s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.402s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.793s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	(cached)
ok  	github.com/Harrison-Blair/hive/internal/homeinit	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/land	3.441s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.207s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.285s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/log	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.017s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.221s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/store	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/logs	(cached)
ok  	github.com/Harrison-Blair/hive/internal/memory	(cached)
ok  	github.com/Harrison-Blair/hive/internal/project	(cached)
ok  	github.com/Harrison-Blair/hive/internal/supervise	(cached)
ok  	github.com/Harrison-Blair/hive/internal/task	9.279s
ok  	github.com/Harrison-Blair/hive/internal/tui	49.303s
ok  	github.com/Harrison-Blair/hive/internal/worktree	4.710s
exit: 0
```

## Cleanup

The scratch base copy, exported grids, scripts, and logs were removed after
their results were copied into this report. No live home or pane was changed.
