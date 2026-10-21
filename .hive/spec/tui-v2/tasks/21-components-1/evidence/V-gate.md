# 21-components-1 V-gate Diff scope and gate

| Item | Value |
| --- | --- |
| Task | 21-components-1 |
| Item | V-gate |
| Proves | AC1, AC2, AC3, AC4 |
| Commit | c06ea712 |
| Date | 2026-10-21 |
| Agent | worker, Claude Code 2.1.307, claude-opus-5-5 |

## Run

`origin/dev` at 0fe6f78e was merged into the task branch in c06ea712. The local
`dev` is at the same commit.

```text
$ git rev-parse --short refs/heads/dev
0fe6f78e
$ git diff dev...HEAD --stat
 .hive/spec/tui-v2/tasks/21-components-1/brief.md | 144 +++++++++++++++++++++++
 internal/tui/component/doc.go                    |  14 +++
 internal/tui/component/fit_test.go               |  92 +++++++++++++++
 internal/tui/component/golden_test.go            |  70 +++++++++++
 internal/tui/component/header.go                 |  60 ++++++++++
 internal/tui/component/header_test.go            |  53 +++++++++
 internal/tui/component/keyhints.go               |  29 +++++
 internal/tui/component/keyhints_test.go          |  45 +++++++
 internal/tui/component/palette.go                |  23 ++++
 internal/tui/component/panel.go                  |  75 ++++++++++++
 internal/tui/component/panel_test.go             |  26 ++++
 internal/tui/component/scrollmark.go             |  31 +++++
 internal/tui/component/subtabs.go                |  39 ++++++
 internal/tui/component/subtabs_test.go           |  33 ++++++
 internal/tui/component/symbols_test.go           |  21 ++++
 internal/tui/component/tabbar.go                 |  52 ++++++++
 internal/tui/component/tabbar_test.go            |  32 +++++
 internal/tui/component/viewport.go               |  27 +++++
 internal/tui/component/viewport_test.go          |  63 ++++++++++
 testdata/ste-baseline.json                       |   1 +
 20 files changed, 930 insertions(+)
exit 0
```

The changed files are the files of the source table of the brief:
`internal/tui/component/`, `testdata/ste-baseline.json`, and the brief copy.
This evidence directory is added after this run.

```text
commit c06ea712
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
ok  	github.com/Harrison-Blair/hive	14.538s
ok  	github.com/Harrison-Blair/hive/cmd	3.090s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.103s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.218s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.229s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.103s
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
ok  	github.com/Harrison-Blair/hive/cmd/task	1.224s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.056s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.107s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.022s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	10.398s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.355s
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
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.013s
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
ok  	github.com/Harrison-Blair/hive/internal/supervise	2.651s
ok  	github.com/Harrison-Blair/hive/internal/task	(cached)
ok  	github.com/Harrison-Blair/hive/internal/tui	49.982s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	6.866s
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
exit 0
```

## Result

Only the files of the source table changed. `gofmt -l` and `go mod tidy -diff`
print nothing, and `go vet`, `golangci-lint`, and `go test -race` exit 0.
