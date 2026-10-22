# Verification: 11-demo-mode

Verdict: pass. Verifier: `t11-verifier-2`. Date: 2026-10-21.

Reviewed worker commit `a4e68c1c`. Merged `dev` at `2b1c23f7` before the runs. The merge commit is `8fbfc636`. This report replaces the first failed report. The repaired file readers and the Design section address the first findings.

| Item | Result | Observation |
| --- | --- | --- |
| V1 Tests | pass | Ran the new tests against `dev` in a scratch copy, then against the task code. Ran controlled red tests for the behaviours below. All task tests pass. The full old TUI suite passes with no expected-value change. |
| V2 Live check | pass | In a new herdr tab, a 206×57 pane drew 120×40 from the top-left cell. Two reads one second apart were equal and showed 14:22. A 100×55 pane showed the too-small notice. Only the scratch log file changed. |
| V3 Schema review | pass | Compared `demo.go`, the fixtures, V3.md, the spec board table, and both board scripts. The fields cover all 21 boards, including future UI state and opened file text. |
| V-gate Diff scope and gate | pass | Merged dev first. The diff has only allowed source and evidence paths. All five gate commands pass. Formatting and tidy print nothing. No gate retry was needed. |

## V1: independent test runs

Full output: `verifier-test-runs.txt`.

Commands:

```sh
go test ./cmd -run 'Demo|SizeWithout' -count=1 -v
go test ./internal/tui -count=1 -v
go test ./internal/tui -run 'Demo|ParseDemo' -count=1 -v
```

The base copy used `git archive dev`, with the two new test files and fixtures copied into it. The CLI action test fails with `unknown flag: --demo`. The size and unknown-field CLI tests fail. The internal package cannot compile because the demo API is absent. This compile failure alone does not prove each assertion. The controlled red runs below check each internal test at runtime.

| Controlled break in a scratch task copy | Observed failing test |
| --- | --- |
| Expose the demo flag | `TestDemoFlagsAreHiddenFromHelp` |
| Remove strict JSON decoding | `TestDemoFixtureUnknownFieldNamesTheField` |
| Bypass fixture value checks | `TestDemoFixtureBadValuesFail` |
| Return the wrong parsed width | `TestParseDemoSize` |
| Use the wall clock | `TestDemoMainShowsFixtureDataAtTheFixtureClock` |
| Do not apply the requested size | `TestDemoSizeDrawsAtTheSizeInALargerPane` |
| Do not detect a pane smaller than the requested size | `TestDemoSizeShowsTooSmallInASmallerPane` |
| Move the notice box one cell right | `TestDemoTooSmallMatchesItsBoardGrid` |
| Remove the blocked-action notice | `TestDemoActionKeysShowTheNotice` |
| Omit fixture log events | `TestDemoLogsShowOnlyFixtureEvents` |
| Disable the four action guards | `TestDemoActionsChangeNoFileAndStartNoProcess`; an unwanted `.hive/state/lock` is written |
| Restore the three old file readers from `b55ca666` | All eight cases of `TestDemoOpensFilesOnlyFromTheFixture` |

The help absence test passes on the base because that base has no demo flag. Exposing the flag makes it fail. The final task runs pass after all scratch mutations. No task source file was changed for these runs.

The repaired test uses live files with known text. Each action must show the fixture text or `demo mode: no file`, and must not show the live text. It covers the task brief, evidence, VERIFIED.md, goal design, and friction entry. Review confirms that `readScreenFile` has no live fallback in demo mode and that Goals skips `os.Stat` in demo mode. Refresh snapshot reads are also disabled in demo mode.

Only the two new test files differ from dev. Existing TUI assertions were not changed. The full package run passes, as does the race gate. The known Health test flake did not occur.

A scratch model probe also compares the normal render path (`demo == nil`) with the demo path at 120×40, with equal fixture loader output and UI state. The content is equal. The probe uses a scratch user working directory. Its first diagnostic run from the agent worktree showed the normal-mode chat caller error; this is the existing caller restriction. Changing the probe working directory to the scratch home removes that unrelated error. This probe checks renderer equality, not normal record loading. Existing tests cover normal loading; a second normal-mode live run was not made.

## V2: live check

Built the task binary with `go build -o SCRATCH/hive .`. Each pane ran:

```sh
stty size
HIVE_HOME=SCRATCH/home COLORTERM=truecolor SCRATCH/hive \
  --demo .hive/spec/tui-v2/design/Main.fixture.json --size 120x40
```

The binary is the task build, not the installed hive binary. The home held only a scratch `.hive/README.md` before the run.

- Owned tab: `w7C:t2`, label `verify11-live`.
- Large pane: `w7C:p2`; `stty size` reported `57 206`.
- Small pane: `w7C:p3`; `stty size` reported `55 100`.
- Both captures used `herdr pane read PANE --source visible --format ansi`.
- No truncation was reported. Both pane records showed scroll offset 0.
- Large captures one second apart were byte-equal. The clock remained 14:22. The capture has 40 rows and a maximum row width of 120. The first row starts at column 0. The renderer size tests also pass.
- The small capture shows `this pane is 100×55`, `hive needs at least 120×40`, and `terminal too small`. The footer is on row 54.
- SHA-256 snapshots show the marker unchanged. The only added file is `.hive/logs/2026-10-21.jsonl`. All eight events have `demo: true`. No state record was made.
- The CLI test independently checks all four blocked actions, file changes, fake process executables, and herdr socket calls. It passes. The live check did not use a process tracer.

Captures: `verifier-main.ansi`, `verifier-small.ansi`, their `.txt` files, and `verifier-*-screen.png`. File snapshots and log events: `verifier-live-files.txt`.

## V3: schema and fixture review

`demoFixture`, `demoData`, and `demoUIState` agree with the schema record in V3.md. The future fields are retained even when the current v1 renderer does not apply them. This is the scope stated in the brief.

| Boards | Required fields confirmed |
| --- | --- |
| Main, AgentsMany, CompactAgents | projects, agents, goals, tasks, inbox, selection, graph scroll, hide_stopped |
| AgentDetails | popup and selected agent |
| Chat, CompactChat | chat, inbox, input, noise, focus, scroll |
| WorkGoals, WorkQueue, Work, Compact | goals, tasks, evidence, files, view, selection, stops_only, details overlay and scroll |
| Friction, FrictionResolved, CompactFriction | entry body and metadata, files, resolved view, selection and scroll |
| Health, HealthAgents, CompactHealth | checks, agents, supervisor, view, stopped visibility, restart prompt and target |
| Logs, CompactLogs | raw events, filter, focus, selection, follow, day, overlay and scroll |
| Help | help overlay and scroll |
| TooSmall | terminal size |
| Components | terminal size; the component sheet has no app state |

All boards also have clock, zone, home, and terminal fields. Unknown typed fields fail with their names. Raw log objects and file-path maps intentionally have arbitrary keys.

Rechecked Main board source groups: size line 59; home line 32; clock and header status line 33; goal line 44; manager line 46; worker line 47; verifier line 48; stopped visibility line 54; selected manager line 60. TooSmall source groups: size line 31; notice and footer lines 32–37. Both fixtures agree with these visible values. Extra IDs, dates, and record metadata support the current loaders and are not visible board values.

## Visual check

The Design section names the approved absolute board paths. Fresh board PNGs were captured from those paths with separate scratch Firefox profiles:

```sh
firefox --headless --no-remote --profile SCRATCH --window-size=1220,780 \
  --screenshot OUT.png file:///home/penguin/source/hive/.hive/spec/tui-v2/design/BOARD.dc.html
```

The first capture attempt had no profile folder and produced no PNG. After creating each folder, all four captures succeeded. The screen PNGs use an ANSI-SGR-to-HTML scratch converter and Firefox at 1220×900. Both screen PNGs include the footer. Compared each fresh screen and board PNG. The Main differences are existing v1 layout differences, which the brief expressly permits. The model probe confirms equal normal and demo rendering with equal data.

### Main

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Home header | `hive` | differs | Current header shows counts and omits the home path. Existing v1 renderer; permitted by Design. |
| Coordinator state | `working`, `●` | differs | State appears in the tree and in aggregate counts, not the board header. Existing v1 renderer. |
| Supervisor state | `●`, tree word `working` | differs | Current header has the live dot but no `live` word; tree shows supervisor working. Existing v1 renderer. |
| Needs-you count | `needs you 0` | matches | Fixture count is zero. |
| Clock | `14:22` | matches | Fixture clock stays fixed. |
| Tab bar | `AGENTS`, named numbered screens | differs | Current sidebar has eight screens. Board has six tabs. Existing v1 renderer. |
| Project panel and done goal | `hive`, worker `landed` | differs | Current tree has no goal box or project-agents panel title. Existing v1 renderer. |
| Selected manager card | `▶`, `idle`, `4m` | differs | Current tree and details panel replace the card; idle glyph is `○`. State and selection have plain-text signals. |
| Worker card | `✓`, `stopped`, `4 landed` | differs | Current tree replaces the card and shows task delivery. Existing v1 renderer. |
| Verifier card | `×`, `stopped` | differs | Current tree omits the board's pass label. Fixture verification is pass. Existing v1 renderer. |
| Stopped count and toggle | two stopped rows, `h hide stopped` | differs | No count line in the current tree. Existing v1 renderer. |
| Spawn and status legend | `working`, `idle`, `stopped` on rows | differs | Current tree has no legend. Existing v1 renderer. |
| Graph key hints | action words in footer | differs | Current footer uses v1 key routes. Existing v1 renderer. |

### TooSmall

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| First-row header | `hive` | matches | First row, bold. |
| Centered notice box | `terminal too small` | matches | Same 42×8 box, centered for the live pane. |
| Current size | `this pane is 100×55` | matches | Live pane dimensions replace board dimensions. |
| Required size | `hive needs at least 120×40` | matches | Requested draw size replaces the board's 80×24 minimum. |
| Resize instruction | `resize or zoom the herdr pane` | matches | Same text. |
| Last-row footer | `ctrl+c ×2 quit` | matches | Present on the final row. |

Removed SGR colour from both screen captures for the state check. Selection, activity, stopped state, and the too-small state retain glyphs or words. No displayed state depends only on colour. The TooSmall fixture also passes the exact board-grid test at 60×16.

## Gate and scope

Full output: `verifier-gate.txt`. Scope output: `verifier-diff-stat.txt`. The README diff changes only Manager decisions. All changed Go files are in cmd/root or internal/tui. The two fixtures, brief, baseline, and evidence are allowed paths. No existing test file changes.

```text
gofmt -l $(git ls-files -co --exclude-standard '*.go'): no output, exit 0
go mod tidy -diff: no output, exit 0
go vet ./...: no output, exit 0
golangci-lint run ./...: 0 issues., exit 0
go test -race ./...: all packages pass, exit 0
```

## Cleanup

Stopped both owned demo processes with ctrl+c twice and checked the shell prompts. Closed `w7C:t2` after checking its label and pane count. The workspace tab list then contained only the verifier tab. The scratch source copies, test probe, binary, home, Firefox profiles, and converter were removed after their results were saved. Final captures and logs remain in this evidence folder. No worker V-file or plan tracker was edited.
