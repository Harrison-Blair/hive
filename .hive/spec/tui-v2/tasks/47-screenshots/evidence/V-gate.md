# V-gate Diff scope and gate

Worker: t47-worker-1. `origin/dev` is `2adaf0ab`; it is already an ancestor of the task branch, so the merge of `dev` had no change (`git merge-base --is-ancestor origin/dev HEAD` exits 0).

## Diff scope

```text
$ git diff dev...HEAD --stat
 .hive/spec/tui-v2/screenshots/AgentDetails.png     | Bin 0 -> 477969 bytes
 .hive/spec/tui-v2/screenshots/AgentsMany.png       | Bin 0 -> 636593 bytes
 .hive/spec/tui-v2/screenshots/Chat.png             | Bin 0 -> 597589 bytes
 .hive/spec/tui-v2/screenshots/Compact.png          | Bin 0 -> 392968 bytes
 .hive/spec/tui-v2/screenshots/CompactAgents.png    | Bin 0 -> 220549 bytes
 .hive/spec/tui-v2/screenshots/CompactChat.png      | Bin 0 -> 294500 bytes
 .hive/spec/tui-v2/screenshots/CompactFriction.png  | Bin 0 -> 427954 bytes
 .hive/spec/tui-v2/screenshots/CompactHealth.png    | Bin 0 -> 279418 bytes
 .hive/spec/tui-v2/screenshots/CompactLogs.png      | Bin 0 -> 287270 bytes
 .hive/spec/tui-v2/screenshots/Friction.png         | Bin 0 -> 1085511 bytes
 .hive/spec/tui-v2/screenshots/FrictionResolved.png | Bin 0 -> 1049884 bytes
 .hive/spec/tui-v2/screenshots/Health.png           | Bin 0 -> 607286 bytes
 .hive/spec/tui-v2/screenshots/HealthAgents.png     | Bin 0 -> 543693 bytes
 .hive/spec/tui-v2/screenshots/Help.png             | Bin 0 -> 562318 bytes
 .hive/spec/tui-v2/screenshots/Logs.png             | Bin 0 -> 1122750 bytes
 .hive/spec/tui-v2/screenshots/Main.png             | Bin 0 -> 339228 bytes
 .hive/spec/tui-v2/screenshots/TooSmall.png         | Bin 0 -> 77162 bytes
 .hive/spec/tui-v2/screenshots/Work.png             | Bin 0 -> 678388 bytes
 .hive/spec/tui-v2/screenshots/WorkGoals.png        | Bin 0 -> 525834 bytes
 .hive/spec/tui-v2/screenshots/WorkQueue.png        | Bin 0 -> 506283 bytes
 .hive/spec/tui-v2/tasks/47-screenshots/brief.md    | 113 +++++++++++++++
 .../tui-v2/tasks/47-screenshots/evidence/V1.md     | 161 +++++++++++++++++++++
 testdata/ste-baseline.json                         |   1 +
 23 files changed, 275 insertions(+)
```

Each path is in the source table of the brief or is evidence: the 20 PNG files, `testdata/ste-baseline.json` (one row for the brief copy, count 0), the brief copy, and `evidence/V1.md`. This file, `V-gate.md`, is added in the next commit. `.gitignore` and `ste_test.go` did not need a change: `!.hive/spec/**` already allows the screenshot folder, and the `steFiles` glob `.hive/spec/*/tasks/*/brief.md` already covers the brief. `go.mod` has no change.

## STE baseline red and green

Red, before the baseline row: `go test -run STE .` failed with `.hive/spec/tui-v2/tasks/47-screenshots/brief.md: no baseline entry (now 0); add it to testdata/ste-baseline.json`. Green, after the row: `ok  github.com/Harrison-Blair/hive 0.839s`.

## Gate

Run from the worktree root on commit `908b8c4f`. Each command exits 0. `gofmt -l` and `go mod tidy -diff` print nothing.

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
ok  	github.com/Harrison-Blair/hive	19.740s
ok  	github.com/Harrison-Blair/hive/cmd	6.680s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.407s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.506s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.547s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.597s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.355s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.329s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.393s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.542s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.337s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.387s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.259s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	14.350s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.674s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.420s
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.890s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.334s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.012s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.307s
ok  	github.com/Harrison-Blair/hive/internal/land	3.628s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.016s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.217s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.016s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	7.803s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.036s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.320s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	4.033s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.182s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.045s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.013s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.015s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.243s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.176s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	27.590s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.011s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.211s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.397s
ok  	github.com/Harrison-Blair/hive/internal/project	1.102s
ok  	github.com/Harrison-Blair/hive/internal/supervise	3.548s
ok  	github.com/Harrison-Blair/hive/internal/task	9.768s
ok  	github.com/Harrison-Blair/hive/internal/tui	37.167s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	17.608s
ok  	github.com/Harrison-Blair/hive/internal/worktree	5.010s
exit 0
```

Result: **pass**.
