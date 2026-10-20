# VERIFIED — 91-agents-live-generation

Verifier: `t91-verifier-1` (Claude Opus 5.5). Task commit: `d310fa18`. Base: `origin/dev` at `ab728cf1`, which is the merge base.

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | The three R3 tests fail on an archive of `origin/dev` with the task test file, and pass on the task branch. |
| V2 | pass | With history off, TREE, GRAPH, and TABLE show one row for each current agent. Stopped agents of landed tasks are absent. With `h`, the footer shows `history: on` and the TABLE reaches all 252 agent names, with each generation of the coordinator (7), the supervisor (16), and manager-hive (9). |
| V3 | pass | Only the brief copy, the evidence files, `internal/tui/agents.go`, `internal/tui/agents_test.go`, `internal/tui/tui.go`, and `testdata/ste-baseline.json` changed. |
| V4 | pass | `gofmt -l` and `go mod tidy -diff` print nothing. `go vet`, `golangci-lint` (`0 issues.`), and `go test -race ./...` exit 0. |

## V1

Red run: `git archive origin/dev` into a scratch folder, copy `internal/tui/agents_test.go` in, then run:

```sh
go test ./internal/tui -run 'TestAgents(DefaultShowsCurrentGenerationsAndActiveTaskGhosts|HistoryToggleShowsAllGenerations|LegacyGenerationFallbackUsesStatusChangeThenRecordTime)$' -count=1
```

It exits FAIL for the correct reasons:

- `TestAgentsDefaultShowsCurrentGenerationsAndActiveTaskGhosts`: each tab has 0 rows for `active-ghost` and 1 row for each hidden generation `coordinator-old`, `supervisor-old`, `manager-old`.
- `TestAgentsHistoryToggleShowsAllGenerations`: the default footer is `h hide stopped`, without `history: off`.
- `TestAgentsLegacyGenerationFallbackUsesStatusChangeThenRecordTime`: all three `legacy-coordinator` records are visible, not only `legacy-c`.

Green run on the task branch: all three tests `PASS`, and `ok github.com/Harrison-Blair/hive/internal/tui`.

### Rules in the evidence against the fixture tests

The rules in `V1.md` agree with the literal names in the fixtures:

- `replaces` wins over time: `coordinator-old` has a newer `status_changed_at` (11:59) than `coordinator-current` (11:58), and only `coordinator-current` shows by default. The same applies to `supervisor-old` and `manager-old`.
- Without lineage, the newest `status_changed_at` wins, then the newest record time: `legacy-b` and `legacy-c` tie at 11:59 and `legacy-c` has the newer record time, so only `legacy-c` shows.
- `active-ghost` (task `11-active`, status `assigned`) shows dim with `×` by default.
- `landed-ghost` (task `12-landed`, commit `abc1234def`, pushed) is hidden by default. With history on, it shows with `✓` and `12 landed · abc1234`. This agrees with the accepted decision: a stopped worker of a landed task is hidden by default and shows with the board's ghost look with `h`.

## V2

Resources that I started, in my own task workspace `wQ0` (before: only tab `wQ0:t1`):

- Tab `wQ0:t2`, pane `wQ0:p2`, label `91-verify-live`.
- Scratch home: a copy of the live home's `.hive/README.md` and `.hive/state` (289 agent records, 252 names) in the session scratchpad. `XDG_CONFIG_HOME` was a temp folder in the scratchpad.
- Command in `wQ0:p2` at `2026-10-20T16:32:15Z`: `env -u HERDR_ENV -u HERDR_PANE_ID -u HERDR_WORKSPACE_ID -u HERDR_SOCKET_PATH HIVE_HOME=<scratch home> XDG_CONFIG_HOME=<scratch xdg> <scratch>/hive-91`, built from `d310fa18` with `go build`.
- PID `1725501`, confirmed with `herdr pane process-info --pane wQ0:p2`.

I pressed `2`, `Tab`, `h`, `j`, `k`, and `ctrl+c` only. I did not open the Chat input. I did not operate in workspace `wBC`.

History off (footer `h history: off`), captured with `herdr pane read wQ0:p2`:

- TREE, GRAPH, and TABLE each show 25 rows: `coordinator`, `supervisor`, `manager-hive`, `manager-qmeter`, `manager-stenographer`, live workers and verifiers, and stopped workers and verifiers of tasks that are not landed or done.
- One row each for `coordinator`, `supervisor`, and `manager-hive` (the live home has 7, 16, and 9 records).
- No row for a stopped agent of a landed task, for example `t51-*` and `t100-*`.
- `herdr pane read --format ansi` shows the dim color `38;2;163;157;144` on the ghost rows, for example `t91-worker-1`.

History on (footer `h history: on`):

- GRAPH shows 7 `coordinator`, 16 `supervisor`, and 9 `manager-hive` rows, with the old rows dim (`38;2;163;157;144`).
- Landed ghosts show with `✓` and a label, for example `✓ t100-worker-1 · pi · 100-no-backgrounds · 100 landed · 2924385`.
- I scrolled the TABLE from the top to the end. The distinct names on the screen are equal to the 252 names in the records. No name is missing.

Stop: two `ctrl+c` in one `herdr pane send-keys` call. `ps -p 1725501` then showed no process. I closed tab `wQ0:t2` with `herdr tab close wQ0:t2`. After, workspace `wQ0` had only `wQ0:t1`. The scratch home chat folder is the same as the live chat folder (`diff -rq` printed nothing).

### Observations for the manager

- Ghosts by default include stopped agents of `cancelled` tasks (`t20`, `t21`, `t25`, `t26`, `t67`) and of `verified` tasks without a pushed delivery (`t10`, `t45`). This agrees with R1 as written ("a task that is not `landed` or `done`") and with the rules in `V1.md`. A cancelled task is possibly not "active" in the sense that the user intends. This is not a failure of this brief.
- The status line shows `error: chat send is for the user only: only the user can send chat; rule 2 (working directory is below projects/)` from the start, before I pressed a key on the Chat screen. The scratch chat folder did not change. The same line is in the worker's run (`V2.md` records only the `not in herdr` line). This is not in the scope of this brief.
- `internal/tui/tui.go` changed by one line: the initial `agentsState` no longer sets `hideStopped: true`. This is the R2 default (history off at start), in the Agents load of `internal/tui/`.

## Incident record

`V2.md` records the incident of the user's TUI: the worker's live check at `2026-10-20T16:04:27.985Z` with Herdr variables set and the real home, SIGINT to the user's PID `1518955` at `16:07:48.995Z`, and the close of `wBC:t11` (pane `wBC:p1A`) at `16:07:54.222Z`. It names the ownership error and states that the V2 run after the manager's rule used only workspace `wPY`.

## V3

```sh
git diff origin/dev --stat
```

```text
 .../tasks/91-agents-live-generation/brief.md       | 100 +++++++++++
 .../tasks/91-agents-live-generation/evidence/V1.md |  36 ++++
 .../tasks/91-agents-live-generation/evidence/V2.md |  23 +++
 .../tasks/91-agents-live-generation/evidence/V3.md |  10 ++
 .../tasks/91-agents-live-generation/evidence/V4.md |  15 ++
 internal/tui/agents.go                             | 181 ++++++++++++++++----
 internal/tui/agents_test.go                        | 187 ++++++++++++++++++++-
 internal/tui/tui.go                                |   2 +-
 testdata/ste-baseline.json                         |   1 +
 9 files changed, 514 insertions(+), 41 deletions(-)
```

`testdata/ste-baseline.json` only adds the brief copy with 0.

## V4

From the worktree root:

```text
== gofmt
rc=0
== tidy
rc=0
== vet
rc=0
== lint
0 issues.
rc=0
== test
ok  	github.com/Harrison-Blair/hive/internal/tui	16.530s
... all packages ok or [no test files]
rc=0
```
