# Verification: 12-board-fixtures

Verifier: `t12-verifier-1`. Result: **pass**.

Task commit: `44407e067eef6ccbc6fbe43a11ee9a1412f7ef12`.
Base: `0fe88b95306b47444ba362a3acd47db42835ad9c` (`dev`).

| Item | Result | Observation |
| --- | --- | --- |
| V1 Tests | pass | The new test fails on base for the 19 absent fixtures. Main and TooSmall pass. All 21 fixtures pass on task code. |
| V2 Fixture review | pass | Compared AgentsMany, AgentDetails, WorkGoals, Health, and Components with their board scripts. The data and UI state agree with the board values. See the tables below. |
| V-gate Diff scope and gate | pass | `git merge dev` prints `Already up to date.` The diff contains only the source-table paths and evidence. All five gate commands exit 0. The first two print nothing. |

## V1 method

From the worktree root, made a scratch copy of base with
`git archive dev | tar -x -C .hive/tmp/verify-12/base`.
Copied only `internal/tui/demo_fixtures_test.go` from task code into that copy.
Ran this command in the base copy, then at the worktree root:

```sh
go test ./internal/tui -run '^TestEachBoardFixtureLoadsAtItsSpecSize$' -count=1 -v
```

The spec table supplies the expected dimensions. The test calls the demo
loader. It does not mock the loader. Base exits 1 because the 19 files do
not exist. Task code exits 0. Full output is below.

## V2 method and scope

Read each of the five `.dc.html` scripts and its `.fixture.json` file under
`.hive/spec/tui-v2/design/`. Compared each visible data group, size, clock,
home, selection, focus, view, and open state. Line numbers below refer to
the board scripts. All five use the home `~/source/hive`, UTC clock `14:22`,
working coordinator, live supervisor, and width 120. Health fixes the seconds
at `14:22:05`; Components has height 67; the others have height 40.

A board does not supply each stored record field. Extra IDs, paths, timestamps,
and off-screen records support the fixture. These fields are not claimed as
literal board values. Empty overlays, popups, prompts, and filters agree with
the boards unless a row below names an open state. Scroll positions are zero
on these five fixtures.

No screen renderer changed. The brief defines V2 as a data review and excludes
screen golden checks. No live pane or PNG capture was made. The state column
below records the text or glyph in the board script, without colour.

### AgentsMany

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Frame and size | working, live, needs you 1, 14:22 | matches | Lines 33–34 and 82; task 8 has `needs_user: approve merge`. |
| Project graph | 5 projects; projects 1–4 of 5; example-ml → | matches | Lines 67–77; five projects, graph left 0. Off-screen ML records do not change the visible four columns. |
| hive | g2-tui-v2 · wave A; working; stopped · pass | matches | Line 69; manager 1m, t5 worker 3m, t6 worker 1m, stopped t4 verifier with pass. Ages follow the fixed clock. |
| example-api | [goal] · wave B; working; idle · 6m | matches | Line 70; codex manager 2m, codex worker with no status time, pi verifier idle 6m. |
| example-web | needs you · approve merge; idle · 9m; stopped · done | matches | Line 71; goal wave and task stop text agree. |
| example-docs | [goal] · wave A; idle · 22m; no workers | matches | Line 72; no child agent for docs. |
| Selection and focus | selected t5-worker-1; [graph] | matches | Lines 66 and 74; selection t5-worker-1, focus graph, no popup. |

The two literal `[goal]` IDs are in different projects. The fixture preserves
the board labels. This review does not approve a change to store ID rules.

### AgentDetails

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Background | hive · g1-tui-restart · done; idle; stopped | matches | Lines 43–47; Main records retained. |
| Popup and size | manager-hive · esc closes; [details] | matches | Lines 53, 61–63, 67; popup agent-details, focus details, manager selection. |
| Identity | manager; hive; claude; claude-opus-5-5 | matches | Line 54; selected manager record supplies the values. |
| Status and nudge | ◐ live · idle; 4m ago by supervisor | matches | Line 54; live/idle record and inbox nudge at 14:18 against clock 14:22. |
| Location and lineage | w58:p6; manager-hive; coordinator-g2; manager-hive-g1 | matches | Line 54; pane, agent name, parent, and replaces agree. The board tab label is the agent name; stored tab_id is an opaque ID. |
| Goal and action | g1-tui-restart · done; open task · none | matches | Lines 54 and 58; done goal and manager with no task. |
| Recent events | 15:08 inbox; 15:07, 14:58, 14:44 land | matches | Line 60; four log messages, task IDs, and short commit IDs agree in reverse time order. |

The recent events occur after the header clock on the board. The fixture
keeps both board values. No timestamp correction is part of this task.

### WorkGoals

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Frame, size, view, focus | Work; [Goals]; [list]; 14:22 | matches | Lines 33, 48–49, 78; work/goals/list and 120×40. |
| Goal list | g2-tui-v2 planned 0/5; g1-tui-restart done 3/3 | matches | Lines 50–60; five open waves and three done waves. Selection g1-tui-restart. |
| Details | TUI restart and stale coordinator startup; done; hive | matches | Line 62; title, status, project, grant dated 2026-10-18, and manager task records agree. |
| Waves | A Stale coordinator startup; B TUI restart; C Agents graph columns | matches | Lines 66–67; each done_when dev, one delivered task each, status done. |
| Tasks | 2 / 0e079272; 3 / 3ccd93fb; 4 / ede94f3e | matches | Lines 69–70; each task has merge delivery to dev with the board commit prefix. |
| Needs you | nothing | matches | Lines 71–72; no needs_user value. |

The design action is a key hint, not an open design file. No goal design text
is visible on this board. The fixture also supplies brief and verification
text for its task records in the files map.

### Health

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Frame, size, view, focus | Health; [Home]; [list]; 14:22 | matches | Lines 32–33, 47–48, 73; health/home/list, 120×40. |
| Check list | checks · 17 · 1 warn; ✓ and ! | matches | Lines 49–57; 17 checks in board order, only unresolved agents warns. |
| Socket, pane, home, store | protocol 22; w58:p4; ~/source/hive; 76 records | matches | Line 51; full doctor detail strings retain these values. |
| Counts and git | starting 0; unresolved 1; 2.43.0 | matches | Line 51; git detail is git version 2.43.0. |
| Supervisor | live · pid 1620686; current | matches | Line 51; live supervisor data and equal binary modification times. |
| Worktrees and branches | none; none | matches | Line 51; stored doctor name unpushed task branches is the full name of the board label. |
| Harnesses and gh | found for five harnesses; 2.102.0 | matches | Line 51; all five paths are nonempty with ok status; gh version detail retains 2.102.0. |
| Selected details | warn · 1 agent; hive doctor; checked 14:22:05 | matches | Lines 53 and 60; unresolved agents selection and clock seconds agree. |
| Unresolved agent | cp5-worker-01; worker; unresolved; gone | matches | Line 64; unresolved record retains its old pane ID. Gone describes the missing live pane. |

The board has static explanatory text under WHAT IT MEANS and NEXT STEP.
These labels and paragraphs are renderer content, not fixture records.
The doctor pane value w58:p4 is retained independently of the hidden
coordinator record pane w58:p3.

### Components

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Size and header | 120×67; working; live; needs you 0; 14:22 | matches | Lines 36–38 and 84; fixed size, home, clock, and header data agree. |
| List | tasks · 4; landed; C, B, A; ▶ 3-tui-restart | matches | Lines 54–58; four task records, three delivered tasks, selected task sample available by ID. |
| Details | landed; g1-tui-restart · wave B; ✓ AC1 | matches | Lines 60–64; task delivery and wave B agree. Files map contains r restarts, confirms only when live. |
| FilterBar | level:warn agent:manager-hive | matches | Line 67; filter open with exact text. |
| Input | restart the supervisor please | matches | Line 69; chat_input contains exact text. |
| Manager card | ● manager-hive; claude · working | matches | Line 71; live working manager. Root UI selection names the manager sample. |
| Verifier card | ◐ t3-verifier-1; pi · idle · 6m | matches | Line 72; live idle verifier, status time 14:16. |
| Worker card | ○ t4-worker-1; pi · stopped | matches | Line 73; stopped worker. |

This board is a component sheet. It has no single screen focus or selection.
Its root agents state satisfies the existing schema. Later component tests
must select the task sample for List/Details and the manager sample for Card.
The fixture holds both samples. Palette, TabBar, SubTabs, Panel, Viewport,
ScrollMark, Popup, KeyHints, and Notice examples use static component arguments.
The worker records this use in V2.md. This verdict does not claim that the
current agents renderer produces the full component sheet.

## Full command output

The following output was captured again by this verifier.

### Base red run

```text
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Main
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/AgentsMany
    demo_fixtures_test.go:37: fixture: open ../../.hive/spec/tui-v2/design/AgentsMany.fixture.json: no such file or directory
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/AgentDetails
    demo_fixtures_test.go:37: fixture: open ../../.hive/spec/tui-v2/design/AgentDetails.fixture.json: no such file or directory
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/CompactAgents
    demo_fixtures_test.go:37: fixture: open ../../.hive/spec/tui-v2/design/CompactAgents.fixture.json: no such file or directory
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Chat
    demo_fixtures_test.go:37: fixture: open ../../.hive/spec/tui-v2/design/Chat.fixture.json: no such file or directory
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/CompactChat
    demo_fixtures_test.go:37: fixture: open ../../.hive/spec/tui-v2/design/CompactChat.fixture.json: no such file or directory
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/WorkGoals
    demo_fixtures_test.go:37: fixture: open ../../.hive/spec/tui-v2/design/WorkGoals.fixture.json: no such file or directory
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/WorkQueue
    demo_fixtures_test.go:37: fixture: open ../../.hive/spec/tui-v2/design/WorkQueue.fixture.json: no such file or directory
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Work
    demo_fixtures_test.go:37: fixture: open ../../.hive/spec/tui-v2/design/Work.fixture.json: no such file or directory
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Compact
    demo_fixtures_test.go:37: fixture: open ../../.hive/spec/tui-v2/design/Compact.fixture.json: no such file or directory
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Friction
    demo_fixtures_test.go:37: fixture: open ../../.hive/spec/tui-v2/design/Friction.fixture.json: no such file or directory
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/FrictionResolved
    demo_fixtures_test.go:37: fixture: open ../../.hive/spec/tui-v2/design/FrictionResolved.fixture.json: no such file or directory
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/CompactFriction
    demo_fixtures_test.go:37: fixture: open ../../.hive/spec/tui-v2/design/CompactFriction.fixture.json: no such file or directory
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Health
    demo_fixtures_test.go:37: fixture: open ../../.hive/spec/tui-v2/design/Health.fixture.json: no such file or directory
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/HealthAgents
    demo_fixtures_test.go:37: fixture: open ../../.hive/spec/tui-v2/design/HealthAgents.fixture.json: no such file or directory
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/CompactHealth
    demo_fixtures_test.go:37: fixture: open ../../.hive/spec/tui-v2/design/CompactHealth.fixture.json: no such file or directory
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Logs
    demo_fixtures_test.go:37: fixture: open ../../.hive/spec/tui-v2/design/Logs.fixture.json: no such file or directory
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/CompactLogs
    demo_fixtures_test.go:37: fixture: open ../../.hive/spec/tui-v2/design/CompactLogs.fixture.json: no such file or directory
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Help
    demo_fixtures_test.go:37: fixture: open ../../.hive/spec/tui-v2/design/Help.fixture.json: no such file or directory
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/TooSmall
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Components
    demo_fixtures_test.go:37: fixture: open ../../.hive/spec/tui-v2/design/Components.fixture.json: no such file or directory
--- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Main (0.00s)
    --- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize/AgentsMany (0.00s)
    --- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize/AgentDetails (0.00s)
    --- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize/CompactAgents (0.00s)
    --- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize/Chat (0.00s)
    --- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize/CompactChat (0.00s)
    --- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize/WorkGoals (0.00s)
    --- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize/WorkQueue (0.00s)
    --- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize/Work (0.00s)
    --- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize/Compact (0.00s)
    --- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize/Friction (0.00s)
    --- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize/FrictionResolved (0.00s)
    --- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize/CompactFriction (0.00s)
    --- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize/Health (0.00s)
    --- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize/HealthAgents (0.00s)
    --- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize/CompactHealth (0.00s)
    --- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize/Logs (0.00s)
    --- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize/CompactLogs (0.00s)
    --- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize/Help (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/TooSmall (0.00s)
    --- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize/Components (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui	0.005s
FAIL

exit=1
```

### Task green run (exit 0)

```text
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Main
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/AgentsMany
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/AgentDetails
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/CompactAgents
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Chat
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/CompactChat
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/WorkGoals
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/WorkQueue
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Work
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Compact
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Friction
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/FrictionResolved
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/CompactFriction
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Health
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/HealthAgents
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/CompactHealth
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Logs
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/CompactLogs
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Help
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/TooSmall
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Components
--- PASS: TestEachBoardFixtureLoadsAtItsSpecSize (0.01s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Main (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/AgentsMany (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/AgentDetails (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/CompactAgents (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Chat (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/CompactChat (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/WorkGoals (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/WorkQueue (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Work (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Compact (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Friction (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/FrictionResolved (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/CompactFriction (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Health (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/HealthAgents (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/CompactHealth (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Logs (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/CompactLogs (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Help (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/TooSmall (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Components (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/tui	0.014s
```

### Diff scope after merge of dev

```text
 .hive/spec/tui-v2/design/AgentDetails.fixture.json |  58 ++++
 .hive/spec/tui-v2/design/AgentsMany.fixture.json   |  71 ++++
 .hive/spec/tui-v2/design/Chat.fixture.json         |  93 ++++++
 .hive/spec/tui-v2/design/Compact.fixture.json      |  60 ++++
 .../spec/tui-v2/design/CompactAgents.fixture.json  |  53 +++
 .hive/spec/tui-v2/design/CompactChat.fixture.json  |  93 ++++++
 .../tui-v2/design/CompactFriction.fixture.json     |  82 +++++
 .../spec/tui-v2/design/CompactHealth.fixture.json  |  66 ++++
 .hive/spec/tui-v2/design/CompactLogs.fixture.json  | 356 +++++++++++++++++++++
 .hive/spec/tui-v2/design/Components.fixture.json   |  57 ++++
 .hive/spec/tui-v2/design/Friction.fixture.json     |  82 +++++
 .../tui-v2/design/FrictionResolved.fixture.json    |  82 +++++
 .hive/spec/tui-v2/design/Health.fixture.json       |  66 ++++
 .hive/spec/tui-v2/design/HealthAgents.fixture.json |  66 ++++
 .hive/spec/tui-v2/design/Help.fixture.json         |  45 +++
 .hive/spec/tui-v2/design/Logs.fixture.json         | 356 +++++++++++++++++++++
 .hive/spec/tui-v2/design/Work.fixture.json         |  60 ++++
 .hive/spec/tui-v2/design/WorkGoals.fixture.json    |  61 ++++
 .hive/spec/tui-v2/design/WorkQueue.fixture.json    |  60 ++++
 .hive/spec/tui-v2/tasks/12-board-fixtures/brief.md | 110 +++++++
 .../tasks/12-board-fixtures/evidence/V-gate.md     | 128 ++++++++
 .../tui-v2/tasks/12-board-fixtures/evidence/V1.md  | 103 ++++++
 .../tui-v2/tasks/12-board-fixtures/evidence/V2.md  | 250 +++++++++++++++
 internal/tui/demo_fixtures_test.go                 |  50 +++
 testdata/ste-baseline.json                         |   1 +
 25 files changed, 2509 insertions(+)
```

### Full gate

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
ok  	github.com/Harrison-Blair/hive	17.646s
ok  	github.com/Harrison-Blair/hive/cmd	5.795s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.136s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.392s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.309s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.134s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.007s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.977s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.332s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.066s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.120s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.024s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	11.363s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.425s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	6.550s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	(cached)
ok  	github.com/Harrison-Blair/hive/internal/homeinit	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/land	3.692s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.198s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.254s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/log	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.015s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.244s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/store	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/logs	(cached)
ok  	github.com/Harrison-Blair/hive/internal/memory	(cached)
ok  	github.com/Harrison-Blair/hive/internal/project	(cached)
ok  	github.com/Harrison-Blair/hive/internal/supervise	(cached)
ok  	github.com/Harrison-Blair/hive/internal/task	9.561s
ok  	github.com/Harrison-Blair/hive/internal/tui	53.756s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	(cached)
ok  	github.com/Harrison-Blair/hive/internal/worktree	5.322s
exit=0

```
