# V-gate Diff scope and gate

Proves: each criterion. `origin/dev` was merged into the task branch first (merge `662cac8e`).

## Diff scope

```text
$ git diff origin/dev...HEAD --stat
 .agents/skills/hive-verifier/SKILL.md              |   3 +
 .agents/skills/hive-worker/SKILL.md                |   3 +
 .hive/spec/tui-v2/tasks/8-no-startup/brief.md      | 112 ++++++++++++++++++++
 .../spec/tui-v2/tasks/8-no-startup/evidence/V1.md  |  53 ++++++++++
 .../spec/tui-v2/tasks/8-no-startup/evidence/V2.md  |  80 +++++++++++++++
 .../spec/tui-v2/tasks/8-no-startup/evidence/V3.md  |  70 +++++++++++++
 cmd/root.go                                        |  11 ++
 cmd/root_test.go                                   | 113 ++++++++++++++++++++-
 internal/tui/startup.go                            |  10 +-
 internal/tui/startup_test.go                       |  28 +++++
 internal/tui/test_main_test.go                     |   2 +-
 reference/dogfood/friction/FRICTION.md             |   2 +-
 .../entries/scratch-hive-starts-real-agents.md     |  10 +-
 skills_test.go                                     |  16 +++
 testdata/ste-baseline.json                         |   1 +
 15 files changed, 508 insertions(+), 6 deletions(-)
```

Each file is in the source table or the evidence, with one note: `internal/tui/test_main_test.go` is a test of `startup.go`. Its `TestMain` now also unsets `HIVE_NO_STARTUP`, as `cmd/root_test.go` does. Without this, an agent that exports `HIVE_NO_STARTUP=1` (as the skills now say) gets false failures in the startup tests.

Design note: `cmd/root.go` cannot pass the flag through `internal/tui/tui.go`, because that file is out of the source table. The flag sets `HIVE_NO_STARTUP=1` in the process, and `internal/tui/startup.go` reads the variable. Only the value `1` has an effect. With the switch, startup shows no footer notice.

## Gate

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
ok  	github.com/Harrison-Blair/hive	14.105s
ok  	github.com/Harrison-Blair/hive/cmd	3.289s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.122s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.288s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.248s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.115s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.673s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.692s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.274s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.078s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.132s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.029s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	11.072s
ok  	github.com/Harrison-Blair/hive/internal/chat	1.397s
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.757s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.341s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.011s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.285s
ok  	github.com/Harrison-Blair/hive/internal/land	3.336s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.169s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.013s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	7.418s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.036s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.288s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	3.894s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.157s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.042s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.013s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.013s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.234s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.161s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	31.225s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.010s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.645s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.371s
ok  	github.com/Harrison-Blair/hive/internal/project	1.085s
ok  	github.com/Harrison-Blair/hive/internal/supervise	2.782s
ok  	github.com/Harrison-Blair/hive/internal/task	9.033s
ok  	github.com/Harrison-Blair/hive/internal/tui	54.998s
ok  	github.com/Harrison-Blair/hive/internal/worktree	4.581s
exit 0
```
