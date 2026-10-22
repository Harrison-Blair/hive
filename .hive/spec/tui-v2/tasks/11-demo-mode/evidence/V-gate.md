# V-gate

This record replaces the first record. It is for the repair after the verify fail (t11-verifier-1).

## Merges of dev

- `git merge origin/dev` at `fdb4eedf` gave the merge commit `e9bd66d0`, with no conflict.
- `dev` then got task `21-components-1`. `git merge origin/dev` at `e816f283` gave the merge commit `bfbdc1bf`. One conflict was in `testdata/ste-baseline.json`: this branch has `.hive/spec/tui-v2/tasks/11-demo-mode/brief.md: 0`, and dev added `.hive/spec/tui-v2/tasks/21-components-1/brief.md: 0` on the same line. Both entries were kept, in sorted order.

The V2 live check ran on the build after the first merge. The second merge adds the package `internal/tui/component` and its tests. No demo code calls it. The gate below ran after the second merge, and the demo tests pass in it.

## Diff against dev

```text
$ git diff origin/dev...HEAD --stat
 .hive/spec/tui-v2/README.md                        |  12 +
 .hive/spec/tui-v2/design/Main.fixture.json         |  59 +++
 .hive/spec/tui-v2/design/TooSmall.fixture.json     |  37 ++
 .hive/spec/tui-v2/tasks/11-demo-mode/brief.md      | 167 +++++++
 .../tui-v2/tasks/11-demo-mode/evidence/V-gate.md   | 138 ++++++
 .../spec/tui-v2/tasks/11-demo-mode/evidence/V1.md  | 201 ++++++++
 .../tasks/11-demo-mode/evidence/V2-brief-open.txt  |  40 ++
 .../tasks/11-demo-mode/evidence/V2-main-screen.png | Bin 0 -> 65996 bytes
 .../11-demo-mode/evidence/V2-pane-120x40.ansi      |  40 ++
 .../tasks/11-demo-mode/evidence/V2-pane-120x40.txt |  40 ++
 .../tasks/11-demo-mode/evidence/V2-pane-small.ansi |  55 +++
 .../tasks/11-demo-mode/evidence/V2-pane-small.txt  |  55 +++
 .../11-demo-mode/evidence/V2-too-small-screen.png  | Bin 0 -> 17302 bytes
 .../spec/tui-v2/tasks/11-demo-mode/evidence/V2.md  | 142 ++++++
 .../spec/tui-v2/tasks/11-demo-mode/evidence/V3.md  |  98 ++++
 .../tui-v2/tasks/11-demo-mode/evidence/VERIFIED.md |  31 ++
 cmd/root.go                                        |  42 ++
 cmd/root_demo_test.go                              | 191 +++++++
 internal/tui/agents.go                             |  75 ++-
 internal/tui/chat_screen.go                        |  28 +-
 internal/tui/demo.go                               | 549 +++++++++++++++++++++
 internal/tui/demo_test.go                          | 351 +++++++++++++
 internal/tui/doc.go                                |   1 +
 internal/tui/friction_screen.go                    |  23 +-
 internal/tui/goals_queue.go                        |  16 +-
 internal/tui/header.go                             |  30 +-
 internal/tui/refresh.go                            |   2 +-
 internal/tui/shell.go                              |   2 +
 internal/tui/task_screen.go                        |  40 +-
 internal/tui/tui.go                                |  60 ++-
 testdata/ste-baseline.json                         |   1 +
 31 files changed, 2459 insertions(+), 67 deletions(-)
```

Each file is in the source table of the brief or is task evidence. `VERIFIED.md` is the fail record of the verifier. No other file changed.

## STE fixes

- `.hive/spec/tui-v2/README.md`: the new sentence about `files` had 26 words. It is now two sentences.
- `.hive/spec/tui-v2/tasks/11-demo-mode/brief.md` line 25: the brief copy from the manager had `confirms`, which the glossary does not allow. It now has `checks`. The meaning does not change. The manager was told.

## Full gate

From the worktree root, after the merge commit `bfbdc1bf`:

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
ok  	github.com/Harrison-Blair/hive	12.834s
ok  	github.com/Harrison-Blair/hive/cmd	(cached)
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	(cached)
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	(cached)
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	(cached)
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	(cached)
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	(cached)
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	(cached)
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	(cached)
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	(cached)
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	(cached)
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	(cached)
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	(cached)
ok  	github.com/Harrison-Blair/hive/internal/boardexport	(cached)
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	(cached)
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	(cached)
ok  	github.com/Harrison-Blair/hive/internal/homeinit	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/land	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/log	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/process	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/records	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/run	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/store	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/logs	(cached)
ok  	github.com/Harrison-Blair/hive/internal/memory	(cached)
ok  	github.com/Harrison-Blair/hive/internal/project	(cached)
ok  	github.com/Harrison-Blair/hive/internal/supervise	(cached)
ok  	github.com/Harrison-Blair/hive/internal/task	(cached)
ok  	github.com/Harrison-Blair/hive/internal/tui	48.324s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	5.683s
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
exit 0
```

`gofmt -l` and `go mod tidy -diff` print nothing. Each other command exits 0.

## Known flaky test: TestHealthProgramUsesStringReader

One run of `go test ./internal/tui ./cmd/...` before the gate failed with:

```text
--- FAIL: TestHealthProgramUsesStringReader (0.01s)
    testing.go:1619: TempDir RemoveAll cleanup: unlinkat /tmp/TestHealthProgramUsesStringReader151883470/001/.hive/logs: directory not empty
```

The test is old and not changed by this task. `go test -count=40 -run 'TestHealthProgramUsesStringReader$' ./internal/tui` failed 4 times in 40 on a clean detached worktree of `origin/dev` at `fdb4eedf`, and 3 times in 40 on the task branch. So the flake is in `dev` and this task does not cause it. This is the probable cause of the unnamed `--- FAIL` line in V1. The manager was told.
