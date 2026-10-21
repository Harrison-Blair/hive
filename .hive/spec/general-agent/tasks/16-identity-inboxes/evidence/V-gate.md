# V-gate Diff scope and gate

Commit: `cba5223e` on `task/16-identity-inboxes`, after a merge of `origin/dev` (cf5d94ef).

## Diff scope

```text
$ git diff origin/dev...HEAD --stat
 .hive/spec/general-agent/README.md                 |   2 +-
 .../tasks/16-identity-inboxes/brief.md             | 148 +++++++++++++++++++++
 cmd/caller_root_test.go                            |  94 +++++++++++++
 cmd/chat/chat.go                                   |   6 +
 cmd/common/common.go                               |  28 ++++
 cmd/goal/goal.go                                   |   6 +
 cmd/inbox/inbox.go                                 |   4 +-
 cmd/inbox_root_test.go                             | 136 +++++++++++++++++++
 cmd/task/complete/complete.go                      |   3 +
 cmd/task/edit/edit.go                              |   3 +
 cmd/task/edit/edit_test.go                         |   2 +
 cmd/task/feature_test.go                           |   1 +
 cmd/task/reopen/reopen.go                          |   3 +
 cmd/task/verify/verify.go                          |   3 +
 internal/agent/caller.go                           |  44 ++++++
 internal/agent/doc.go                              |   1 +
 internal/agent/route.go                            |  51 +++++++
 internal/agent/templates/project-agents.md         |   1 -
 internal/doctor/checks.go                          |  50 +++++++
 internal/doctor/checks_test.go                     |   7 +-
 internal/doctor/orphan_inbox_test.go               |  43 ++++++
 internal/doctor/run.go                             |   1 +
 internal/inbox/read.go                             |  32 ++---
 internal/inbox/read_send_test.go                   |  28 ++--
 internal/inbox/send.go                             |  10 +-
 internal/supervise/events.go                       |   8 +-
 internal/supervise/events_test.go                  |  20 +--
 internal/supervise/fix_test.go                     |   8 +-
 internal/supervise/reconcile.go                    |   2 +-
 internal/supervise/reconcile_test.go               |  10 +-
 internal/supervise/routing_test.go                 | 141 ++++++++++++++++++++
 internal/supervise/supervise.go                    |  21 +--
 testdata/ste-baseline.json                         |   1 +
 33 files changed, 841 insertions(+), 77 deletions(-)
```

Each path is in the source table of the brief. The evidence files are in the next commit.

- `internal/agent/route.go` holds `ParentInbox` and `InboxKey`. The brief limits `internal/agent/` to the caller rule, but the changed AC4 asks for one shared function that `hive agent report` (wave 5, `internal/agent`) reuses. The supervisor imports `internal/agent` already.
- `cmd/task/edit/edit_test.go` and `cmd/task/feature_test.go` set `HERDR_PANE_ID` to empty. These tests use `--agent` for a caller with no pane, which is the case outside herdr. Without this line they read the pane of the shell that runs the tests. The `cmd` package `TestMain` does the same.

## Recorded names and scope

- Caller rule (AC1, decision 24): `agent.Caller` in `internal/agent/caller.go`, with `agent.CallerOptions{Agent, Pane, HerdrPane}`. Wave 5 calls it for `hive agent report` and `hive agent spawn`. `cmd/common.CheckCaller` wraps it for the commands that have only an `--agent` caller flag.
- Inbox key (AC2, decision 26): `agent.InboxKey` in `internal/agent/route.go`. `coordinator`, `manager-PROJECT`, else the agent name.
- Parent route (AC4, decision 28 as changed on 2026-10-21): `agent.ParentInbox` in `internal/agent/route.go`. It follows the `replaces` chain from the parent to its live successor, else `coordinator`. Wave 5 calls it for `hive agent report`. The supervisor calls it through `parentInbox`. The coordinator and the supervisor get no record about themselves, as before.
- `hive doctor`: the new optional check `ownerless inbox records` lists each unread record as `KEY/SEQ`.

Commands that now apply the caller rule:

| Command | `--agent` | `--pane` |
| --- | --- | --- |
| `hive inbox` | refused when it is not the pane agent | refused when it is not the caller pane |
| `hive inbox send` | refused when it is not the pane agent | refused when it is not the caller pane |
| `hive task complete`, `verify`, `reopen`, `edit` | refused when it is not the pane agent | unchanged |
| `hive goal create`, `goal edit` | refused when it is not the pane agent | unchanged |
| `hive chat send`, `chat reply` | refused when it is not the pane agent | unchanged |

Commands left unchanged:

- Target `--agent` flags: `hive task assign`, `hive agent send`, `hive agent stop`, `hive agent resolve`. `hive agent show` takes the name as an argument. `hive logs --agent` is a filter.
- `hive agent spawn --agent` names the caller. Wave 5 adds its rule and tests (decision 25).
- `hive agent register` uses `--pane` with no `--agent`. Wave 4 or 5 adds its rule (decision 24).
- `hive task create`, `list`, `show`, `cancel`, `goal show`, `goal list`, and `chat log` accept the persistent `--agent` flag, but they do not read it.

## Gate

`gofmt -l` and `go mod tidy -diff` print nothing. Each command exits 0.

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
ok  	github.com/Harrison-Blair/hive	14.520s
ok  	github.com/Harrison-Blair/hive/cmd	3.610s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.108s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.282s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.241s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.102s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.735s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.750s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.258s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.067s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.122s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.024s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	11.068s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.360s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.384s
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.774s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.367s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	1.010s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.306s
ok  	github.com/Harrison-Blair/hive/internal/land	3.381s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	1.014s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.166s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	1.014s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	7.434s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.033s
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.291s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	3.925s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.164s
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	1.043s
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.015s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	1.014s
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.227s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.174s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	27.181s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.009s
ok  	github.com/Harrison-Blair/hive/internal/logs	5.644s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.382s
ok  	github.com/Harrison-Blair/hive/internal/project	1.082s
ok  	github.com/Harrison-Blair/hive/internal/supervise	3.112s
ok  	github.com/Harrison-Blair/hive/internal/task	9.180s
ok  	github.com/Harrison-Blair/hive/internal/tui	50.147s
ok  	github.com/Harrison-Blair/hive/internal/worktree	4.533s
exit 0
```
