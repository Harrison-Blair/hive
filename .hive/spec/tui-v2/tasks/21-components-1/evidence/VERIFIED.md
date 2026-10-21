# Independent verification: 21-components-1

Verifier: `t21-verifier-2` (pi). Verdict: **pass**.

Task commit: `879a1f85`. Base: `dev`, `fdb4eedf`.
`git merge dev` returned `Already up to date.`
This report replaces the first verifier's report after the repair.

## Results

| Item | Result | Observation |
| --- | --- | --- |
| V1 Tests | pass | All 18 tests pass on task code. Each test was run separately on base with empty implementation stubs. All 17 component tests fail. The dependency-only symbol test passes, as the amended brief permits. |
| V1 Mutations | pass | A working-to-done token change fails the Header golden checks and prints four differing cells. Replacing `×` with `中` fails with width 2 and names the symbol. Restoring `len(gap)` fails the new tab-gap regression test. All mutations were restored. |
| V2 Rectangles | pass | Checked Header, SubTabs, and KeyHints against the approved board script and adjacent grid cells. Each rectangle contains the whole sample and excludes other components. |
| V-gate | pass | Diff contains only permitted component files, brief, baseline, and evidence. All five commands exit 0. Formatting and tidy print nothing. Full output follows. |
| Visual check | pass | Captured the board, compact boards, and actual component samples again. Text, geometry, palette tokens, and bold attributes match the golden rectangles. |

## Test method

Commands ran from the worktree root. A detached scratch worktree at `dev`
received each new `*_test.go` and the empty implementation stubs in V1.md,
plus `var tabGap = " "`. These stubs permit assertion failures on the base,
which has no component API. Each test ran with:

```sh
(cd SCRATCH/base && go test -count=1 -v ./internal/tui/component/ -run '^TEST$')
go test -count=1 -v ./internal/tui/component/
```

The combined base run stopped on a panic in the tab-gap test after its empty
stub returned one row. Separate runs then checked all 18 tests, including
Viewport and Thumb. These failures show absent behavior. The independent
`len(gap)` mutation checks the repaired behavior on the task implementation.

Each mutation ran `go test -count=1 ./internal/tui/component/`, with
`-run TestSymbolWidth` or `-run TestTabBarMultiByteGap` for the latter two.
The final full gate ran after all source files were restored.
Source review confirms that the remaining production `len` calls count line
slices. `TabBar.View` now measures the gap with `lipgloss.Width`.

## Independent rectangle review

Source: `/home/penguin/source/hive/.hive/spec/tui-v2/design/Components.dc.html`.
Coordinates are zero-based (row, column, width, height).

| Component | Rectangle | Board script | Result |
| --- | --- | --- | --- |
| Header | (7, 0, 120, 1) | Lines 36–38 write the home and right-aligned status, with coloured dots. | Whole header row; caption row 6 and blank row 8 excluded. |
| SubTabs | (14, 0, 120, 1) | Line 44 writes Goals at column 1, Queue at 10, Tasks at 19, and the hint at 102. | Includes labels, active brackets, and complete hint; excludes captions. |
| KeyHints | (64, 0, 72, 1) | Line 79 writes the region at column 0 and hints at 9. The hint ends at column 71. | Whole sample text. Trailing blank board cells contain no other element. Full-width padding is separately checked against Main and CompactAgents, per the manager decision in V2.md. |

## Visual check

New captures: `verifier-board.png`, `verifier-samples.png`,
`verifier-CompactAgents.png`, and `verifier-CompactFriction.png`.
The saved `verifier-capture.go.txt` rendered actual component ANSI through
`cellgrid.Convert`. A scratch HTML file rendered the resulting cells as spans.
The sample output was also inspected without colour, using the grid characters.

Capture command, with a new, empty profile directory for each capture:

```sh
firefox --headless --no-remote --profile SCRATCH --window-size=1220,1280 --screenshot OUT.png file:///home/penguin/source/hive/.hive/spec/tui-v2/design/Components.dc.html
```

The samples used 1220×1400. Compact boards used 1220×780. The first capture
attempt failed because profile directories did not exist. After creation,
all four captures succeeded. The sample sections have different pixel
positions from the board; the golden tests compare their terminal cells.
No integrated screen or agent pane was used. Screen wiring is out of scope.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Palette | Token names and hex labels | matches | Component colours use decision 26 tokens. |
| Header, long | `working`, `live`, `needs you 0`, `14:22` | matches | Home and status layout agree. |
| Header, compact | `coord ●`, `sup ●`, `needs you 0` | matches | State glyphs remain; idle and stopped use distinct glyphs in tests. |
| TabBar | Numbered labels and active `━` underline | matches | Long and compact spacing agree. |
| SubTabs | `[Tasks]` | matches | Brackets identify the active view. |
| Panel, unfocused | `unfocused` title | matches | Rounded border, title, and padding agree. |
| Panel, focused | `focused`, `takes the keys` | matches | Sample has a text focus signal; screen callers must retain a non-colour signal. |
| Viewport, vertical | `┃`, line numbers 41–45 | matches | Right-border thumb and content agree. |
| Viewport, horizontal | `━`, `projects 1–4 of 5` | matches | Bottom-border thumb and content agree. |
| KeyHints | `[graph]` and action names | matches | Region and hints agree. |

Other sheet components are outside this task.

## Cleanup

Temporary source mutations were restored. The scratch base worktree, browser
profiles, generated HTML, scripts, grids, and redundant logs were removed after
the results were saved. No workspace, tab, or pane was created. No worker
`V<n>.md` file or plan tracker was changed.

## Run output

### Base tests

```text
TestSymbolWidth: exit 0
TestTabBarGolden: exit 1
TestTabBarSpacing: exit 1
TestTabBarMultiByteGap: exit 1
TestViewportGolden: exit 1
TestViewportXOffset: exit 1
TestThumb: exit 1
TestKeyHintsGolden: exit 1
TestKeyHintsFit: exit 1
TestPanelGolden: exit 1
TestSubTabsGolden: exit 1
TestSubTabsText: exit 1
TestHeaderGolden: exit 1
TestHeaderForms: exit 1
TestHeaderStateDots: exit 1
TestHeaderCutsLongHome: exit 1
TestFitFrameRows: exit 1
TestFitPanels: exit 1
```

### Green tests

```text
=== RUN   TestFitFrameRows
--- PASS: TestFitFrameRows (0.10s)
=== RUN   TestFitPanels
--- PASS: TestFitPanels (0.34s)
=== RUN   TestHeaderGolden
=== RUN   TestHeaderGolden/long
=== RUN   TestHeaderGolden/short
--- PASS: TestHeaderGolden (0.00s)
    --- PASS: TestHeaderGolden/long (0.00s)
    --- PASS: TestHeaderGolden/short (0.00s)
=== RUN   TestHeaderForms
--- PASS: TestHeaderForms (0.00s)
=== RUN   TestHeaderStateDots
--- PASS: TestHeaderStateDots (0.00s)
=== RUN   TestHeaderCutsLongHome
--- PASS: TestHeaderCutsLongHome (0.00s)
=== RUN   TestKeyHintsGolden
=== RUN   TestKeyHintsGolden/sheet
=== RUN   TestKeyHintsGolden/main
=== RUN   TestKeyHintsGolden/compact
--- PASS: TestKeyHintsGolden (0.00s)
    --- PASS: TestKeyHintsGolden/sheet (0.00s)
    --- PASS: TestKeyHintsGolden/main (0.00s)
    --- PASS: TestKeyHintsGolden/compact (0.00s)
=== RUN   TestKeyHintsFit
--- PASS: TestKeyHintsFit (0.00s)
=== RUN   TestPanelGolden
=== RUN   TestPanelGolden/unfocused
=== RUN   TestPanelGolden/focused
--- PASS: TestPanelGolden (0.00s)
    --- PASS: TestPanelGolden/unfocused (0.00s)
    --- PASS: TestPanelGolden/focused (0.00s)
=== RUN   TestSubTabsGolden
=== RUN   TestSubTabsGolden/sheet
=== RUN   TestSubTabsGolden/work
=== RUN   TestSubTabsGolden/compact
--- PASS: TestSubTabsGolden (0.00s)
    --- PASS: TestSubTabsGolden/sheet (0.00s)
    --- PASS: TestSubTabsGolden/work (0.00s)
    --- PASS: TestSubTabsGolden/compact (0.00s)
=== RUN   TestSubTabsText
--- PASS: TestSubTabsText (0.00s)
=== RUN   TestSymbolWidth
--- PASS: TestSymbolWidth (0.00s)
=== RUN   TestTabBarGolden
=== RUN   TestTabBarGolden/sheet
=== RUN   TestTabBarGolden/work
=== RUN   TestTabBarGolden/compact
--- PASS: TestTabBarGolden (0.00s)
    --- PASS: TestTabBarGolden/sheet (0.00s)
    --- PASS: TestTabBarGolden/work (0.00s)
    --- PASS: TestTabBarGolden/compact (0.00s)
=== RUN   TestTabBarSpacing
--- PASS: TestTabBarSpacing (0.00s)
=== RUN   TestTabBarMultiByteGap
--- PASS: TestTabBarMultiByteGap (0.00s)
=== RUN   TestViewportGolden
=== RUN   TestViewportGolden/vertical
=== RUN   TestViewportGolden/horizontal
--- PASS: TestViewportGolden (0.00s)
    --- PASS: TestViewportGolden/vertical (0.00s)
    --- PASS: TestViewportGolden/horizontal (0.00s)
=== RUN   TestViewportXOffset
--- PASS: TestViewportXOffset (0.00s)
=== RUN   TestThumb
--- PASS: TestThumb (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/tui/component	0.450s
```

### Token mutation

```text
--- FAIL: TestHeaderGolden (0.00s)
    --- FAIL: TestHeaderGolden/long (0.00s)
        header_test.go:15: 2 cells differ from Components.grid (board row and column):
            row 7 col 68: want '●' working bold, got '●' done bold
            row 7 col 91: want '●' working bold, got '●' done bold
            
            rendered:
            [1m hive  ~/source/hive[m                                    [1mcoordinator [m[1;38;2;124;207;106m●[m[1m working   supervisor [m[1;38;2;124;207;106m●[m[1m live   needs you 0   14:22 [m
    --- FAIL: TestHeaderGolden/short (0.00s)
        header_test.go:16: 2 cells differ from CompactAgents.grid (board row and column):
            row 0 col 48: want '●' working bold, got '●' done bold
            row 0 col 56: want '●' working bold, got '●' done bold
            
            rendered:
            [1m hive[m                                     [1mcoord [m[1;38;2;124;207;106m●[m[1m · sup [m[1;38;2;124;207;106m●[m[1m · needs you 0 · 14:22 [m
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui/component	0.399s
FAIL
```

### Symbol mutation

```text
--- FAIL: TestSymbolWidth (0.00s)
    symbols_test.go:18: lipgloss.Width("中") = 2, want 1
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui/component	0.002s
FAIL
```

### Tab-gap regression mutation

```text
--- FAIL: TestTabBarMultiByteGap (0.00s)
    tabbar_test.go:46: underline starts at cell 59, want cell 54 (labels "  1 Agents \u00a0 2 Chat \u00a0 3 Work \u00a0 4 Friction \u00a0 5 Health \u00a0 6 Logs \u00a0                                                          ", rule "───────────────────────────────────────────────────────────━━━━━━━━──────────────────────────────────────────────────────")
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui/component	0.002s
FAIL
```

### Full gate

```text

$ gofmt -l $(git ls-files -co --exclude-standard "*.go")
exit 0

$ go mod tidy -diff
exit 0

$ go vet ./...
exit 0

$ golangci-lint run ./...
0 issues.
exit 0

$ go test -race ./...
ok  	github.com/Harrison-Blair/hive	16.762s
ok  	github.com/Harrison-Blair/hive/cmd	3.376s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.114s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.288s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.239s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.096s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	(cached)
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	(cached)
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.244s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.070s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.121s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.027s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	11.324s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.366s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	5.583s
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
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.281s
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
ok  	github.com/Harrison-Blair/hive/internal/tui	50.949s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	(cached)
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
exit 0
```

### Diff scope

```text
 .hive/spec/tui-v2/tasks/21-components-1/brief.md   | 148 +++++++
 .../tasks/21-components-1/evidence/V-gate.md       | 155 +++++++
 .../tui-v2/tasks/21-components-1/evidence/V1.md    | 466 +++++++++++++++++++++
 .../tui-v2/tasks/21-components-1/evidence/V2.md    | 163 +++++++
 .../tasks/21-components-1/evidence/VERIFIED.md     | 402 ++++++++++++++++++
 .../evidence/verifier-CompactAgents.png            | Bin 0 -> 44561 bytes
 .../evidence/verifier-CompactFriction.png          | Bin 0 -> 90896 bytes
 .../21-components-1/evidence/verifier-board.png    | Bin 0 -> 225318 bytes
 .../evidence/verifier-capture.go.txt               |  27 ++
 .../21-components-1/evidence/verifier-samples.png  | Bin 0 -> 78302 bytes
 internal/tui/component/doc.go                      |  14 +
 internal/tui/component/fit_test.go                 |  92 ++++
 internal/tui/component/golden_test.go              |  70 ++++
 internal/tui/component/header.go                   |  60 +++
 internal/tui/component/header_test.go              |  53 +++
 internal/tui/component/keyhints.go                 |  29 ++
 internal/tui/component/keyhints_test.go            |  45 ++
 internal/tui/component/palette.go                  |  23 +
 internal/tui/component/panel.go                    |  75 ++++
 internal/tui/component/panel_test.go               |  26 ++
 internal/tui/component/scrollmark.go               |  31 ++
 internal/tui/component/subtabs.go                  |  39 ++
 internal/tui/component/subtabs_test.go             |  33 ++
 internal/tui/component/symbols_test.go             |  21 +
 internal/tui/component/tabbar.go                   |  55 +++
 internal/tui/component/tabbar_test.go              |  48 +++
 internal/tui/component/viewport.go                 |  27 ++
 internal/tui/component/viewport_test.go            |  63 +++
 testdata/ste-baseline.json                         |   1 +
 29 files changed, 2166 insertions(+)
```
