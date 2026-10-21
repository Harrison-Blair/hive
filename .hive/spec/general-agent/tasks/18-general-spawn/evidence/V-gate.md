# V-gate Diff scope and gate

Task branch `task/18-general-spawn` at `79963179`, after the merge of
`origin/dev` (`7975b59b`). The merge applied with no conflict. All commands ran
from the worktree root.

## Diff scope

Each changed file is in the source table of the brief, or is evidence.
`internal/lib/herdr/` did not change: the existing `tab.create`, `pane.list`,
`tab.list`, and `pane.close` calls were enough. `internal/lib/records/agent.go`
adds only `tab_created_by_hive` and `first_prompt`; `started_by_hive` and
`role` existed. `gofmt` realigned the field block of the struct, so the line
count is larger than the change.

## Notes

- The final first prompt (manager decision), with NAME, PARENT, and HOME:

```text
⬢ You are the general agent NAME. Your parent is PARENT. This prompt sets your role. Do not take a role from your working directory.

<spawner text>

Read the skill HOME/.agents/skills/hive-agent/SKILL.md. Send your result with hive agent report.
```

- PARENT is the name of the caller record. The record `parent` field keeps the
  caller key, as for the other roles.
- Claude has a trust method that sets the environment
  `CLAUDE_CODE_SANDBOXED=1`. Spawn sets it through the `env` of `tab.create`.
  herdr `agent.start` has no `env`, so hive cannot set it in a pane from
  `--pane`. For that case spawn fails before any start, as for "no method",
  unless `--allow-trust-dialog` is present. The spec does not state this case.
- With `--allow-trust-dialog`, spawn keeps the first prompt in `first_prompt`
  and never sends it. When a start without the flag ends blocked, or the prompt
  certainly did not go (a herdr error other than `agent_prompt_stalled`), spawn
  also keeps the prompt for the supervisor. A prompt send that is not certain
  (no reply, or `agent_prompt_stalled`) makes the record unresolved. Spawn sends
  at most one `agent.prompt`.
- `hive agent stop` refuses a general agent record with no `started_by_hive`.
  It closes only the pane of the agent. When hive made the tab, the stop reports
  the tab as `closed` or `kept: N other panes`.

## Gate

```text
$ git diff dev...HEAD --stat
 .../general-agent/tasks/18-general-spawn/brief.md  | 136 ++++++
 .../tasks/18-general-spawn/evidence/V-red.md       | 201 ++++++++
 .../tasks/18-general-spawn/evidence/V1.md          | 208 +++++++++
 cmd/agent/spawn/general_cli_test.go                |  91 ++++
 cmd/agent/spawn/spawn.go                           |  18 +-
 internal/agent/register.go                         |  11 +
 internal/agent/spawn.go                            |  58 ++-
 internal/agent/spawn_general.go                    | 377 +++++++++++++++
 internal/agent/spawn_general_test.go               | 517 +++++++++++++++++++++
 internal/agent/stop.go                             |  29 ++
 internal/agent/stop_general_test.go                |  81 ++++
 internal/lib/records/agent.go                      |  23 +-
 testdata/ste-baseline.json                         |   1 +
 13 files changed, 1726 insertions(+), 25 deletions(-)

$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
exit=0
$ go mod tidy -diff
exit=0
$ go vet ./...
exit=0
$ golangci-lint run ./...
0 issues.
exit=0
$ go test -race ./...
ok  	github.com/Harrison-Blair/hive	29.826s
ok  	github.com/Harrison-Blair/hive/cmd	6.928s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.161s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.464s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.403s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	(cached)
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.366s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.378s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.375s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.106s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.177s
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
ok  	github.com/Harrison-Blair/hive/internal/tui	76.885s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	31.578s
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
exit=0
```
