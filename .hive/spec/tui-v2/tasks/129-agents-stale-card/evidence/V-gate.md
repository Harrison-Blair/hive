# V-gate Diff scope and gate

## Merge of dev

`git fetch origin && git merge origin/dev`: `Already up to date.` The base is
`0c12e20e` (dev after task 126).

## Diff scope: `git diff origin/dev...HEAD --stat`

```text
 .../tui-v2/tasks/129-agents-stale-card/brief.md    | 125 +++++++++++++++
 .../tasks/129-agents-stale-card/evidence/V1.md     |  85 ++++++++++
 .../tasks/129-agents-stale-card/evidence/V2.md     | 112 +++++++++++++
 .../129-agents-stale-card/evidence/V2/base-h.txt   |  40 +++++
 .../evidence/V2/base-start.txt                     |  40 +++++
 .../129-agents-stale-card/evidence/V2/herdr-h.txt  |  61 +++++++
 .../evidence/V2/herdr-health.txt                   |  61 +++++++
 .../129-agents-stale-card/evidence/V2/herdr-hh.txt |  61 +++++++
 .../evidence/V2/herdr-start.txt                    |  61 +++++++
 .../129-agents-stale-card/evidence/V2/tmux-h.txt   |  40 +++++
 .../129-agents-stale-card/evidence/V2/tmux-hh.txt  |  40 +++++
 .../evidence/V2/tmux-start.txt                     |  40 +++++
 internal/tui/agents.go                             |  54 ++++++-
 internal/tui/agents_screen.go                      |  10 +-
 internal/tui/agents_screen_test.go                 |   5 +-
 internal/tui/agents_stale_test.go                  | 178 +++++++++++++++++++++
 .../tui/testdata/agents-stale/coordinator-g1.json  |  20 +++
 .../tui/testdata/agents-stale/coordinator-g2.json  |  20 +++
 .../tui/testdata/agents-stale/coordinator-g3.json  |  20 +++
 .../tui/testdata/agents-stale/coordinator-g4.json  |  20 +++
 .../tui/testdata/agents-stale/coordinator-g5.json  |  20 +++
 .../tui/testdata/agents-stale/coordinator-g6.json  |  21 +++
 .../tui/testdata/agents-stale/coordinator-g7.json  |  21 +++
 .../tui/testdata/agents-stale/coordinator-g8.json  |  23 +++
 .../testdata/agents-stale/manager-qmeter-g1.json   |  20 +++
 .../testdata/agents-stale/manager-qmeter-g2.json   |  20 +++
 .../testdata/agents-stale/manager-qmeter-g3.json   |  21 +++
 .../testdata/agents-stale/manager-qmeter-g4.json   |  23 +++
 .../agents-stale/manager-stenographer-g1.json      |  20 +++
 .../agents-stale/manager-stenographer-g2.json      |  20 +++
 .../agents-stale/manager-stenographer-g3.json      |  19 +++
 .../agents-stale/manager-stenographer-g4.json      |  20 +++
 .../agents-stale/manager-stenographer-g5.json      |  20 +++
 .../agents-stale/manager-stenographer-g6.json      |  20 +++
 .../agents-stale/manager-stenographer-g7.json      |  21 +++
 .../agents-stale/manager-stenographer-g8.json      |  23 +++
 reference/dogfood/friction/FRICTION.md             |   2 +-
 .../agents-card-shows-stale-record.md              |   4 +-
 testdata/ste-baseline.json                         |   1 +
 39 files changed, 1417 insertions(+), 15 deletions(-)
```

Each changed file is in the source table of the brief, or is evidence, or is the
brief:

- `internal/tui/agents.go`, `internal/tui/agents_screen.go`, and their tests
  (`agents_screen_test.go`, new `agents_stale_test.go`).
- `internal/tui/testdata/agents-stale/`: new fixtures.
- `reference/dogfood/friction/FRICTION.md` and the entry file (moved to
  `entries/resolved/`, status `Resolved 2026-10-23`).
- `testdata/ste-baseline.json`: the brief with 0. To get 0, the AC4 row of the
  brief changed from "The friction entry is resolved" (passive voice) to "The
  friction entry has the status `Resolved`".

`git diff origin/dev...HEAD --name-only | grep -Ei "grid|golden|design|board"`
gives no file: no grid, golden, or board changed (AC2). Each existing grid and
golden check passes in the test run below.

`internal/tui/header.go` and `health.go` are not changed. They call
`currentAgentKeys`, so they use the new rule: the header and Health › Agents
take the live coordinator record.

## Gate, from the worktree root

```text
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
ok  	github.com/Harrison-Blair/hive	16.962s
ok  	github.com/Harrison-Blair/hive/cmd	(cached)
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	(cached)
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	(cached)
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	(cached)
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	(cached)
ok  	github.com/Harrison-Blair/hive/cmd/chat	(cached)
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	(cached)
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	(cached)
ok  	github.com/Harrison-Blair/hive/cmd/task	(cached)
ok  	github.com/Harrison-Blair/hive/cmd/task/create	(cached)
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	(cached)
ok  	github.com/Harrison-Blair/hive/cmd/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/agent	(cached)
ok  	github.com/Harrison-Blair/hive/internal/boardexport	(cached)
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	(cached)
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
ok  	github.com/Harrison-Blair/hive/internal/tui	(cached)
ok  	github.com/Harrison-Blair/hive/internal/tui/component	(cached)
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
exit=0
```

The TUI packages again, with no cache:

```text
ok  	github.com/Harrison-Blair/hive/internal/tui	26.384s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	8.475s
```

Note: in one earlier run without `-race`, `TestHealthProgramUsesStringReader`
failed once in its `t.TempDir` clean-up (`.hive/logs: directory not empty`).
That test is not changed by this task. It passed in each later run.
