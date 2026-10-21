# 21-components-1 V-gate Diff scope and gate

| Item | Value |
| --- | --- |
| Task | 21-components-1 |
| Item | V-gate |
| Proves | AC1, AC2, AC3, AC4 |
| Commit | 6ea87ed8 |
| Date | 2026-10-21 |
| Agent | worker, Claude Code 2.1.307, claude-opus-5-5 |

## Run

This run is after the repair of the verify fail of t21-verifier-1. The task
branch has `origin/dev` at fdb4eedf from merge c9b8c90b. A new merge of
`origin/dev` gave "Already up to date". The local `dev` is at the same commit.

```text
$ git rev-parse --short dev
fdb4eedf
$ git diff dev...HEAD --stat
 .hive/spec/tui-v2/tasks/21-components-1/brief.md   | 148 ++++++++
 .../tasks/21-components-1/evidence/V-gate.md       | 143 ++++++++
 .../tui-v2/tasks/21-components-1/evidence/V1.md    | 316 ++++++++++++++++
 .../tui-v2/tasks/21-components-1/evidence/V2.md    | 163 +++++++++
 .../tasks/21-components-1/evidence/VERIFIED.md     | 402 +++++++++++++++++++++
 .../evidence/verifier-CompactAgents.png            | Bin 0 -> 44561 bytes
 .../evidence/verifier-CompactFriction.png          | Bin 0 -> 90896 bytes
 .../21-components-1/evidence/verifier-board.png    | Bin 0 -> 225318 bytes
 .../evidence/verifier-capture.go.txt               |  27 ++
 .../21-components-1/evidence/verifier-samples.png  | Bin 0 -> 78302 bytes
 internal/tui/component/doc.go                      |  14 +
 internal/tui/component/fit_test.go                 |  92 +++++
 internal/tui/component/golden_test.go              |  70 ++++
 internal/tui/component/header.go                   |  60 +++
 internal/tui/component/header_test.go              |  53 +++
 internal/tui/component/keyhints.go                 |  29 ++
 internal/tui/component/keyhints_test.go            |  45 +++
 internal/tui/component/palette.go                  |  23 ++
 internal/tui/component/panel.go                    |  75 ++++
 internal/tui/component/panel_test.go               |  26 ++
 internal/tui/component/scrollmark.go               |  31 ++
 internal/tui/component/subtabs.go                  |  39 ++
 internal/tui/component/subtabs_test.go             |  33 ++
 internal/tui/component/symbols_test.go             |  21 ++
 internal/tui/component/tabbar.go                   |  55 +++
 internal/tui/component/tabbar_test.go              |  48 +++
 internal/tui/component/viewport.go                 |  27 ++
 internal/tui/component/viewport_test.go            |  63 ++++
 testdata/ste-baseline.json                         |   1 +
 29 files changed, 2004 insertions(+)
exit 0
```

The changed files are the files of the source table of the brief:
`internal/tui/component/`, `testdata/ste-baseline.json`, and the brief copy.
The other files are the evidence of this task, with `VERIFIED.md` and the
captures of the verifier. This file is changed after this run.

```text
commit 6ea87ed8
$ gofmt -l $(git ls-files -co --exclude-standard "*.go")
exit 0

$ go mod tidy -diff
exit 0

$ go vet ./...
exit 0

$ golangci-lint run ./...
0 issues.
exit 0

$ go test -race ./...
ok  	github.com/Harrison-Blair/hive	14.396s
ok  	github.com/Harrison-Blair/hive/cmd	3.275s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.132s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.342s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.248s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.116s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.836s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.802s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.290s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.064s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.098s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.023s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	9.638s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.393s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.676s
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
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.232s
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
ok  	github.com/Harrison-Blair/hive/internal/tui	51.408s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	6.110s
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
exit 0
```

Each command exits 0. `gofmt -l` and `go mod tidy -diff` print nothing.
