# Task 129 independent verification

- Task: `129-agents-stale-card`
- Verifier: `t129-verifier-1`
- Date: 2026-10-23
- Checked commit: `8ebb18c7`
- Base: `dev` and `origin/dev`, both `0c12e20e`
- Verdict: **pass**

## Verification items

| Item | Result | Observation |
| --- | --- | --- |
| V1 Tests red and green | pass | Each of the seven new tests failed 50 times on the base implementation. Each passed on the task branch with `-race -count=50`. |
| V2 Live check | pass | The scratch TUI showed one card for each name, from the live or newest record. `h` hid the stale worker and kept the live cards. |
| V-gate Diff scope and gate | pass | The source changes stay in the named files. No board, grid, or golden file changed. All five gate commands exited 0. |
| Manager's extra stale-supervisor check | pass | Live current workers and verifiers remain visible when the supervisor claim is stale and stopped agents are hidden. |

## V1 Tests

Command on each implementation:

```sh
go test -race -count=50 -run '^(TestCurrentAgentKeysGiveOneRecordForEachName|TestCurrentAgentKeysPreferTheLiveRecord|TestAgentsOverviewDrawsOneCardForEachName|TestAgentsPodViewDrawsTheLiveManager|TestAgentsHHidesAStaleWorkerWithNoSuccessor|TestAgentsCoordinatorIsTheLiveRecord|TestHealthAgentsListOneRowForEachName)$' ./internal/tui/
```

I made a scratch source copy with `git archive HEAD`. In that copy, I replaced
`agents.go` and `agents_screen.go` with their `dev` versions for the red run.
These are the only changed production source files. The new tests and fixtures
stayed in the copy. The task source in the worktree did not change.

| Test | Base result | Task result |
| --- | --- | --- |
| `TestCurrentAgentKeysGiveOneRecordForEachName` | 50 failures: six current records instead of three | pass, 50 runs with race check |
| `TestCurrentAgentKeysPreferTheLiveRecord` | 50 failures: newer stale record selected | pass, 50 runs with race check |
| `TestAgentsOverviewDrawsOneCardForEachName` | 50 failures: stale coordinator card | pass, 50 runs with race check |
| `TestAgentsPodViewDrawsTheLiveManager` | 50 failures: stale manager card | pass, 50 runs with race check |
| `TestAgentsHHidesAStaleWorkerWithNoSuccessor` | 50 failures: stale worker not hidden | pass, 50 runs with race check |
| `TestAgentsCoordinatorIsTheLiveRecord` | 50 failures: coordinator g6 selected instead of g8 | pass, 50 runs with race check |
| `TestHealthAgentsListOneRowForEachName` | 50 failures: six rows instead of three | pass, 50 runs with race check |

The red command exited 1. The green command exited 0.
[Red failure counts and messages](verifier/tests-red-summary.txt) and
[green output](verifier/tests-green.txt) record these runs.

I compared all 20 sanitized fixtures with their home records.
The retained fields match, except for subsequent activity updates in the two
live records. Those differences affect `status_changed_at`, `agent_status`,
and `updated_at`. The fixtures retain the recorded activity at the worker's
capture time. Old records retain their missing `parent` and `replaces` fields.
[Fixture comparison](verifier/fixture-comparison.txt).

## V2 Live check

I built the checked task commit as `hive` in a scratch `bin` folder. I put that
folder first in `PATH` in my own herdr pane, `wWC:p2`.

The scratch home was
`.hive/tmp/t129-verifier-live/home` in this worktree. It held:

- All 20 agent fixtures.
- Project records for `qmeter` and `stenographer`.
- A fresh supervisor claim on an owned `sleep` process, with its process start time.
- One additional stale worker with no successor. This record copies
  `manager-qmeter-g1` and changes its key, name, role, and task.

From that home, with `HIVE_HOME` set to that home, I ran:

```sh
HIVE_NO_STARTUP=1 COLORTERM=truecolor hive
```

No coordinator or supervisor process started. I did not run the TUI on the real home.

The outer herdr layout changed to 207×57 during this check. I set only the
owned TUI's PTY to 120×40 with `stty -F /dev/pts/7 rows 40 cols 120`.
`stty` reported `40 120`. Every saved screen has 40 rows and a maximum width
of 120 columns. Thus the checked TUI viewport was 120×40 inside the larger
herdr layout. [Dimensions](verifier/live-dimensions.txt).

I opened Agents with `1`, pressed `h`, and pressed `h` again.

| Record | Expected card | Observed result |
| --- | --- | --- |
| `manager-qmeter-g4` | `● manager-qmeter`, `pi · working` | One card, correct harness and state, before and after `h` |
| `coordinator-g8` | `◐ coordinator`, `claude · idle` | One card, from the live record with `agent_status: done` |
| `manager-stenographer-g8` | `○ manager-stenograp…`, `claude · stopped` | One card from the newest exited record |
| `worker-qmeter-g1` | `○ worker-qmeter`, `claude · unknown` | Hidden at start, shown after `h`, hidden after the second `h` |

The stale worker counts in `h shows 1 stopped` and `h hides 1 stopped`.
The stopped manager stays visible, as the existing manager-card design requires.
The manager's decision in worker evidence V1 preserves that rule.

I also opened the qmeter pod, both details popups, and Health › Agents.
The manager details show harness `pi`, parent `coordinator-g8`, and
`replaces manager-qmeter-g3`. The coordinator details show harness `claude`
and `replaces coordinator-g7`. Health lists exactly four current rows:
coordinator g8, manager-qmeter, manager-stenographer, and the stale worker.

Each capture has a plain-text file, an ANSI file, and a PNG in `verifier/`:

- `live-start`, `live-h`, `live-hh`
- `live-pod`
- `live-manager-details`, `live-coordinator-details`
- `live-health`
- `live-supervisor-stale`

I checked the expected card counts and `h` behavior against the saved text.
All checks passed.

## Manager's extra check

Inbox record 1 asks whether a stale supervisor claim hides live current agents.
I added a test only in the scratch source copy. It opens Agents with a dead
supervisor claim and leaves `hideStopped` true. Its independent expectations
require `worker-1` and `legacy-verifier` to stay visible with `unknown` status.
It also requires their record status to stay `live` and their current flag to stay true.

The test passed with `-race -count=50`. I then changed the scratch implementation
to hide all nodes with shown status `unknown`. The test failed for the missing
live worker. After I restored the task implementation, it passed another 50 runs.

- [Test source](verifier/supervisor-test.txt)
- [Deliberately broken run](verifier/supervisor-mutation.txt)
- [Restored green run](verifier/supervisor-green.txt)

For the live check, I added a live worker with no replacement to the scratch home.
I changed only the scratch supervisor claim to PID 0. The live worker stayed
visible as `○ live-worker`, `pi · unknown`. The actual stale worker stayed
hidden, and the pod still reported `h shows 1 stopped`.
See [plain text](verifier/live-supervisor-stale.txt) and
[PNG](verifier/live-supervisor-stale.png).

## Visual check

### Design authority flag

The task brief has no `Design` section. I reported this flag to the manager.
Manager inbox record 2 supplies the existing committed Agents and Health boards
and their grids as the authority. AC2 permits no board change.

I captured these approved boards again with Firefox and separate new profiles:

- `/home/penguin/source/hive/projects/hive/worktrees/129-agents-stale-card/.hive/spec/tui-v2/design/Main.dc.html`
- `/home/penguin/source/hive/projects/hive/worktrees/129-agents-stale-card/.hive/spec/tui-v2/design/LiveOverview.dc.html`
- `/home/penguin/source/hive/projects/hive/worktrees/129-agents-stale-card/.hive/spec/tui-v2/design/LivePod.dc.html`
- `/home/penguin/source/hive/projects/hive/worktrees/129-agents-stale-card/.hive/spec/tui-v2/design/AgentDetails.dc.html`
- `/home/penguin/source/hive/projects/hive/worktrees/129-agents-stale-card/.hive/spec/tui-v2/design/HealthAgents.dc.html`

Their PNG files are `verifier/board-NAME.png`.
I captured each changed screen from the owned pane with
`herdr pane read wWC:p2 --source visible --format ansi`.
A short scratch script converted SGR codes to an HTML terminal grid.
Firefox captured each HTML file with `--window-size=1220,780`.
[Conversion script](verifier/ansi-to-html-script.txt).

I compared all screen PNGs with these board PNGs. These captures use live
fixtures, rather than board fixtures. The board HTML adds one leading blank
line that the ANSI renderer does not add. This creates a 17-pixel vertical
offset between PNG renderers. The unchanged golden checks compare terminal
cells and all passed. I do not claim pixel equality for these different data sets.

### Elements and states without colour

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header and counts | `done`, `old`, `? 0 asks`, `needs you 0` | differs | Scratch path, clock, and records differ from the board data. No supervisor lifecycle record exists in the scratch home. |
| Screen tabs | `1 Agents`, `5 Health` | matches | The selected tab uses bold text and its underline. |
| Agents region | `project agents`, project count or pod name | matches | Fixed frame, title, and graph region remain present. |
| Coordinator card | `◐ coordinator`, `claude · idle` | differs | The live fixture is idle. The live boards show a working coordinator. |
| Coordinator branch | Line glyphs from the coordinator to each pod | matches | Two project pods match the scratch project records. No message exists to light a line. |
| Project pod titles | `qmeter`, `stenographer` | differs | Scratch project names and project count differ from the board fixtures. |
| Selected pod | Heavy border glyphs `┏`, `┃`, `┗` | matches | Selection has a shape signal without colour. |
| Manager card | `● manager-qmeter`, `pi · working` | differs | The live manager fixture supplies the name, harness, and state. Card shape matches. |
| Stopped manager card | `○`, `claude · stopped` | differs | This scratch fixture has an exited manager. The existing design keeps manager cards visible. |
| Stale worker card | `○ worker-qmeter`, `claude · unknown` | overridden | AC1 adds stale records to the child-card `h` rule. Existing state rendering uses `unknown`. |
| Hidden-agent count | `h shows 1 stopped`, `h hides 1 stopped` | matches | The existing count now includes the stale worker. |
| Empty pod text | `no child agents` | matches | The stenographer pod has no child records. |
| One-pod layout | Pod title, coordinator branch, heavy selected card | matches | The fixed one-pod frame remains. Fewer cards follow from the scratch records. |
| Graph legend | `● working`, `◐ idle`, `○ stopped`, line glyphs | matches | Each state has a glyph and a word. |
| Graph key footer | `[graph]`, `enter`, `h stopped`, `esc pods` | matches | The footer changes with overview and one-pod modes as the boards require. |
| Details popup frame | Agent name and `esc closes` | matches | The popup retains its border, fields, and dim background. |
| Manager details fields | `● live · working`, `harness pi`, parent and replaces keys | differs | Fields come from live manager g4, rather than the board manager. |
| Coordinator details fields | `◐ live · idle`, `harness claude`, `replaces coordinator-g7` | differs | Fields come from live coordinator g8. The role has no project or parent field. |
| Details actions and recent events | `ACTIONS`, `a`, `p`, `t`, `RECENT` | matches | Action labels stay present. Event contents follow the scratch run. |
| Health view selector | `Home [Agents]` | matches | The selected view has brackets and bold text. |
| Health list columns and rows | `AGENT`, `ROLE`, `STATE`, `◐ idle`, `● working`, `○ stopped`, `stale` | differs | Four current scratch records replace the board's five demo records. The stale word remains visible without colour. |
| Health selection | `▶ coordinator-g8` | matches | The selection arrow remains visible without colour. |
| Health details and actions | `role`, `state`, `pane`, `ACTIONS`, restart text | differs | The selected scratch record is the coordinator. The board selects the supervisor. |
| Health restart prompt | No prompt in the scratch run | differs | I did not request restart. The board shows that optional prompt. Existing prompt tests passed. |
| Health key footer | `[list]`, movement keys, `r restart`, `R refresh` | differs | The live capture shows list mode. The board shows restart-confirmation mode. |
| Live agent with a stale supervisor | `○ live-worker`, `pi · unknown` | matches | The live card stays visible. Only an actual stale or stopped child record can hide. |

I checked each changed screen after colour removal in its plain-text capture.
Every displayed agent state has a glyph or a word. No checked state uses colour
as its only signal.

## V-gate

After `git fetch origin`, both base refs still named `0c12e20e`.
`git merge dev --no-edit` returned `Already up to date.`
[Diff stat](verifier/diff-stat.txt) records `git diff dev...HEAD --stat`.

The changed production files are `internal/tui/agents.go` and
`internal/tui/agents_screen.go`. The tests, fixtures, friction row and moved
entry, and STE baseline are in the source table. The registered brief copy
and task evidence are also present. No board, grid, or golden file changed.
The existing stale-supervisor test retains its assertions after the new `h` action.

The friction entry has status `Resolved 2026-10-23` in
`reference/dogfood/friction/entries/resolved/`. Its index row links to that file.
The old entry path no longer exists.

Full gate output: [verifier/gate.txt](verifier/gate.txt).

| Command | Result |
| --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | exit 0, no output |
| `go mod tidy -diff` | exit 0, no output |
| `go vet ./...` | exit 0, no output |
| `golangci-lint run ./...` | exit 0, `0 issues.` |
| `go test -race ./...` | exit 0, all packages passed |

## Clean-up

I stopped the owned TUI and the owned claim process. I closed only my scratch
tab `wWC:t2`. A subsequent read of `wWC:p2` returned `pane_not_found`.
I removed the scratch home, binary, source copy, browser profiles, HTML files,
and temporary logs. A search of `.hive/tmp/` found no remaining
`t129-verifier` paths. Only the final evidence remains.
