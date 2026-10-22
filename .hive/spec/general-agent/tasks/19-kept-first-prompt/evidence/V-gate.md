# 19-kept-first-prompt V-gate Diff scope and gate

| Item | Value |
| --- | --- |
| Task | 19-kept-first-prompt |
| Item | V-gate |
| Proves | AC1, AC2, AC3 |
| Commit | 1e791a74 (merge of `origin/dev` at `4ddb1413`) |
| Date | 2026-10-21 |
| Agent | worker, claude, claude-opus-5-5 |

## Merge

`git fetch origin` and `git merge origin/dev` merged `4ddb1413` (task
20-agent-report). Two files had conflicts. Both conflicts were additions on
the two sides. The merge keeps each addition.

- `internal/lib/records/agent.go`: the fields `FirstPromptAttemptAt` and
  `FirstPromptSentAt` of this task, and the field `ReportedAt` of `dev`.
- `testdata/ste-baseline.json`: the entries for the 19 and 20 briefs, each 0.

The merge changes no code of this task.

## Diff scope

Each changed file is in the source table of the brief, or is the brief copy
or the evidence. `VERIFIED.md` and `verifier-runs.md` are from the verifier.

```text
$ git diff origin/dev...HEAD --stat
 .../tasks/19-kept-first-prompt/brief.md            | 109 ++++++++
 .../tasks/19-kept-first-prompt/evidence/V-gate.md  | 133 ++++++++++
 .../tasks/19-kept-first-prompt/evidence/V-red.md   | 195 ++++++++++++++
 .../tasks/19-kept-first-prompt/evidence/V1.md      | 161 ++++++++++++
 .../19-kept-first-prompt/evidence/VERIFIED.md      |  88 +++++++
 .../19-kept-first-prompt/evidence/verifier-runs.md | 288 +++++++++++++++++++++
 internal/agent/spawn.go                            |  10 +-
 internal/agent/spawn_general.go                    |  47 ++++
 internal/agent/spawn_general_test.go               |  42 +++
 internal/lib/records/agent.go                      |  14 +-
 internal/lib/records/agent_test.go                 |  38 +++
 internal/supervise/doc.go                          |   5 +-
 internal/supervise/firstprompt.go                  | 120 +++++++++
 internal/supervise/firstprompt_test.go             | 245 ++++++++++++++++++
 internal/supervise/nudge.go                        |   3 +
 internal/supervise/reconcile.go                    |   7 +-
 testdata/ste-baseline.json                         |   1 +
 17 files changed, 1495 insertions(+), 11 deletions(-)
```

## Run

From the worktree root at `1e791a74`. Each `$` line is the exact command.

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
ok  	github.com/Harrison-Blair/hive	28.478s
ok  	github.com/Harrison-Blair/hive/cmd	7.741s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.210s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.364s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.524s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.448s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.176s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.305s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.394s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.435s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.144s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.241s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.055s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	17.693s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.633s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.501s
ok  	github.com/Harrison-Blair/hive/internal/doctor	5.824s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	(cached)
ok  	github.com/Harrison-Blair/hive/internal/homeinit	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.374s
ok  	github.com/Harrison-Blair/hive/internal/land	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.386s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	4.505s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/process	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.037s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/store	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/logs	(cached)
ok  	github.com/Harrison-Blair/hive/internal/memory	(cached)
ok  	github.com/Harrison-Blair/hive/internal/project	(cached)
ok  	github.com/Harrison-Blair/hive/internal/supervise	3.980s
ok  	github.com/Harrison-Blair/hive/internal/task	12.149s
ok  	github.com/Harrison-Blair/hive/internal/tui	78.733s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	(cached)
ok  	github.com/Harrison-Blair/hive/internal/worktree	8.232s
exit 0
```

## Result

`gofmt -l` and `go mod tidy -diff` print nothing. `go vet`, `golangci-lint`,
and `go test -race` exit 0. Only the files of the source table, the brief copy,
and the evidence changed.
