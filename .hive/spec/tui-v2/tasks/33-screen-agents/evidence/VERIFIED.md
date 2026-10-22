# Verification: 33-screen-agents

Verifier: `t33-verifier-4`. Tested commit: `00763bde`.

## Verdict: pass

This run checks the merge repair. The manager limited this run to the diff
since `4c5ffad3`, the removed-test mappings, five board checks, a deliberate
token break, one live Main check, and V-gate. The earlier independent runs
remain in `verify3/`. The earlier merge failure is repaired.

| Item | Result | Observation |
| --- | --- | --- |
| V1 | pass | The Agents and component test run passes. All five board checks pass. The deliberate idle-token break fails tests. The removed-test mappings keep the required behavior. The manager accepts the earlier base compile failure as red when a deliberate break fails the test. |
| V2 | pass | A new Main live capture has zero differing cells under the space-cell rule. The new Main key checks pass. The manager accepts the earlier five-board live run, heavy-selection check, and key-break proof in `verify3/`. |
| V-gate | pass | One fetch and merge of dev: already up to date at `361b8848`. Diff scope is allowed. All five gate commands exit 0. The first two print nothing. No flake rerun was needed. |

## Merge repair review

Reviewed `4c5ffad3..00763bde`, the combined merge diffs, and the final code.

- `e578f8b1` keeps both shell-test expectations: Agents has `no project agents`;
  Chat has `no chat yet`. The STE baseline keeps both task entries.
- `656bac63` keeps both removals from `legacyDetails`. No legacy details
  case remains. The frame uses each screen's panel hook.
- The Agents `t` action uses the new Tasks state. It closes the file state
  and selects the task. Both the model test and the new live check pass.
- The board tests use dev's `checkBoardGolden`. No duplicate compare helper
  remains in the Agents screen test file.
- Health has its own Agents panel. The removed TREE, GRAPH, and TABLE
  renderers have no caller. The shared loader, peek, and restart remain.
- The README change is in `Manager decisions`. The diff against dev contains
  only the brief's source paths and this task's evidence.

### Removed-test mappings

Reviewed each row of `merge-repair.md`, including the initially failing tests
and the other removed old-view tests. No required kept behavior lacks a
new-screen test. The replacement tests were read and run in this verification.
They assert fixed expected text, state, cells, or log fields.

| Kept behavior | New-screen proof rerun |
| --- | --- |
| Current generation only; stopped current agent remains | `TestAgentsGraphShowsTheCurrentGenerationOnly` |
| Hide and show stopped agents; no children after hide | `TestAgentsHHidesStoppedAgents` |
| Long agent key and missing parent record | `TestAgentsPodHoldsAgentsWithoutAParentRecord` |
| Selection, hide, and attach action logs | `TestAgentsKeysLogTheirActions` |
| Stale agent and stale supervisor show unknown | `TestAgentsCardsShowUnknownWhenTheSupervisorIsStale` |
| Forty workers keep selection visible | `TestAgentsFitEachSize/forty_workers` |
| Vi and arrow movement | `TestAgentsUpDownMoveInThePod`, `TestAgentsArrowsScrollThePods` |
| Refresh keeps selected key | `TestAgentsNormalModeLoadsAndRefreshes` |
| Project record order | `TestAgentsPodsFollowProjectRecordOrder` |
| Unread nudged manager; read, absent, and worker exclusions | `TestAgentsDetailsShowAnUnreadNudge` |
| Equal columns, card size, padding, border, tokens, goal and status text | `TestAgentsBoardsMatchTheirGrids`, component card tests |
| Mixed-weight joints and visible selection without colour | `TestAgentsSelectedCardJoints`, `TestAgentsGraphShowsProjectAgentsOnly` |
| Long text fits; borders stay whole | `TestAgentsFitEachSize/long_text` |
| Popup fields and actions, task action, footer hints | Board checks, `TestAgentsDetailsPopupKeys`, `TestAgentsTOpensTheTask`, `TestAgentsCompactFooter` |

The old topology tabs, history view, active-task ghost restriction, inferred
parent text, nudge edges, old detail strip, task-status and next-task fields,
and old working-count label are removed by the decisions listed in the
worker's mapping table. Their replacements follow decision 14, the approved
boards, and the task 33 Manager decisions. The supervisor PID and age tests
are kept as loader checks. The changed peek and generation-fallback tests
also remain and pass. The full race gate includes restart and Health tests.

## V1 commands and output

`verify4/green.log` contains this independent run, exit 0:

```sh
go test ./internal/tui ./internal/tui/component -run 'TestAgents|TestHelpMatchesItsBoard|TestCard|TestFitWave4' -count=1 -v
```

Deliberate break: in `internal/tui/component/card.go`, change the idle dot
from `accentStyle.Render("◐")` to `dimStyle.Render("◐")`. Run:

```sh
go test ./internal/tui ./internal/tui/component -run 'TestAgentsBoardsMatchTheirGrids|TestAgentsGraphShowsProjectAgentsOnly|TestHelpMatchesItsBoard|TestCard' -count=1 -v
```

Exit 1, as required. The four Agents board cases, the idle-dot assertion,
and the idle card golden fail. Full output: `verify4/token-red.log`.
The source was restored; `git diff --quiet -- internal/tui/component/card.go`
exits 0. After the restore and dev merge, this command exits 0:

```sh
go test ./internal/tui -run 'TestAgentsBoardsMatchTheirGrids|TestHelpMatchesItsBoard' -count=1 -v
```

`verify4/restored-green.log` records Main, AgentsMany, AgentDetails,
CompactAgents, and the full Help check. The base compile red, component
behavioral red, full Help red, and hide-key break of the earlier independent
run remain in `verify3/`. The eight repair-test break results are recorded
in `merge-repair/breaks.txt`; this run reviewed those results and reran their
green tests. It does not claim a new execution of those eight breaks.

## V2 Main live check

Built the tested task with `go build`. Herdr client and server: 0.9.5,
protocol 22. Owned scratch tab `wAQ:t2`, pane `wAQ:p2`, 57 visible rows.
The command used an absolute scratch binary and fixture path:

```sh
HIVE_HOME=SCRATCH_HOME HIVE_NO_STARTUP=1 COLORTERM=truecolor SCRATCH_BINARY --demo WORKTREE/.hive/spec/tui-v2/design/Main.fixture.json --size 120x40
```

The scratch home had its own marker. No real record was used. Two reads
500 ms apart were equal (`Main-first.ansi`, `Main.ansi`). Both used
`herdr pane read wAQ:p2 --source visible --format ansi`. Metadata is in
`Main-pane.json`: scroll offset 0. The CLI gives ANSI text, with no JSON
truncation field; the capture had all required rows and no truncation notice.
The comparison cropped the top-left 120×40 cells. It used the existing
comparison source in `evidence/screens/compare-program.go.txt`, rebuilt for
this run.

`verify4/live-diff.log`:

```text
Main.grid: strict differences 3438, space-cell token or bold only 3438, differences under the live rule 0
```

The new live keys pass (`verify4/keys.log` and key captures): `h` hides two
stopped agents; `h` shows them again; down and enter open the worker popup;
esc returns to graph hints; `t` opens Work Tasks with
`4-agents-equal-columns` selected. Two Ctrl+C keys return to the shell.

### Visual check

New PNGs: `verify4/Main-board.png` and `verify4/Main-live.png`. The board
capture used the absolute approved path and Firefox with a new scratch
profile, `--headless --no-remote --window-size=1220,780 --screenshot`.
The live PNG used an ANSI SGR to HTML pre conversion and a separate new
profile. Both PNGs were opened and compared. Browser font metrics, margins,
and canvas edges differ. The terminal cell comparison has zero differences.
`verify4/Main-plain.txt` checks the states with no colour.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header | coordinator working, supervisor live, needs you 0, clock | matches | Fixed fixture |
| Active tab | `1 Agents`, `━` underline | matches | Full grid check |
| Outer panel | `project agents · 1 project` | matches | Singular title |
| Project pod | `hive`, goal and done status | matches | Centred pod |
| Manager card | `◐`, `idle`, age | matches | Accent idle token also checked |
| Selected card | Heavy `┏━┓┃┗┛` border | matches | Visible in plain text |
| Tree joints and bus | `┯`, `│`, branch and child joints | matches | Fixed tree layout |
| Worker cards | `○`, `stopped`, done or pass | matches | Two stopped agents |
| Hide notice | `h hides 2 stopped` | matches | Live hide and show also pass |
| Legend | Spawn, working, idle, stopped words and glyphs | matches | State does not depend on colour |
| Footer | `[graph]`, movement and action hints | matches | Popup and task routes also checked |

Other boards and focus states retain the independently checked visual rows
and captures in `verify3/`, accepted by the manager for this repair scope.

## V-gate

At the start of V-gate, fetched `origin dev` and ran
`git merge --no-edit origin/dev` once. Result: already up to date.
`origin/dev` and local `dev` were `361b8848`. No conflict or merge-note
exception was needed. `verify4/merge.log` records the output.
`verify4/scope.log` records `git diff dev...HEAD --stat`.

Full output and exit statuses: `verify4/gate.log`.

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go mod tidy -diff
go vet ./...
golangci-lint run ./...
go test -race ./...
```

All exit 0. The first two print nothing. The linter reports `0 issues.`
No known flake occurred.

## Cleanup

The owned live process returned to the shell. The owned scratch tab was
closed and focus returned to `wAQ:t1`. `verify4/cleanup.json` shows only the
verifier tab in this workspace. The scratch binary, comparison program,
home, browser profiles, HTML, and repair diff were removed. Final logs,
ANSI captures, plain text, and PNG evidence remain in `verify4/`.
No task source, worker evidence, or plan tracker was changed by this run.
