# Independent verification: 21-components-1

Verifier: `t21-verifier-1` (pi). Verdict: **fail**.

Task code: `79c8d9b9`. Merge of current `dev` (`fdb4eedf`):
`c9b8c90b10cc555ddb68be095a5cb9b72c312d27`.

## Results

| Item | Result | Observation |
| --- | --- | --- |
| V1 Tests / AC1 review | fail | All 17 tests pass on task code. Sixteen fail with empty implementation stubs on base. The token mutation fails golden checks and prints four differing cells. AC1 requires text width measurement with `lipgloss.Width`, but `tabbar.go:43` uses `len(gap)` to set the width of the separator rule. Also, the literal V1 expectation that each test fails on base is not met: `TestSymbolWidth` passes on base. |
| V2 Rectangles | pass | Independently checked TabBar, Panel, and Viewport against the board script. Each rectangle includes the whole component and no other component. |
| V-gate Diff scope and gate | pass | Merged current dev. Diff has only permitted component files, brief, baseline, and evidence. All five gate commands exit 0. Formatting and tidy print nothing. Full output follows. |
| Visual check | pass | Captured the approved board and rendered component samples again. Checked the changed component sections and compact variants. Geometry, text, token, and bold golden checks pass. |

## Required repair or decision

1. Use `lipgloss.Width(gap)` for the text width in `TabBar.View`.
   The current separator contains only ASCII spaces, so this does not cause
   an observed visual error. It does violate AC1 and decision 12.
2. Clarify V1 for the dependency-only symbol test. `TestSymbolWidth` checks
   existing Lip Gloss behavior. No new component implementation can make
   that test fail on base. Do not weaken this useful test. The independent
   mutation from `×` to `中` fails with width 2, as expected.

No implementation or worker evidence file was changed by this verification.
The two temporary mutations were restored before the final gate.

## V1 independent runs

The base had no component API. A detached scratch worktree at `dev`
(`fdb4eedf`) received the new test files and empty implementation stubs from
V1.md. The stubs only permit compilation; they return empty values. This
checks assertion failures rather than missing-symbol compiler errors.

Commands from the worktree root:

```sh
git worktree add --detach .hive/tmp/verify-21/base dev
# Copy *_test.go and the empty stubs shown in V1.md into the scratch package.
(cd .hive/tmp/verify-21/base && go test -count=1 -v ./internal/tui/component/)
go test -count=1 -v ./internal/tui/component/
# Temporarily replace Foreground(Working) with Foreground(Done) in header.go.
go test -count=1 ./internal/tui/component/
# Restore header.go, then temporarily replace × with 中 in symbols_test.go.
go test -count=1 ./internal/tui/component/ -run TestSymbolWidth
# Restore symbols_test.go.
go test -count=1 -v ./internal/tui/component/
```

Base exit: 1. Task exit: 0. Token mutation exit: 1. Symbol mutation exit: 1.
Restored task exit: 0. The final green run was after the merge of dev.
The base summary, green output, and mutation output follow below.

## V2 independent rectangle review

Coordinates are zero-based. Source is the approved
`/home/penguin/source/hive/.hive/spec/tui-v2/design/Components.dc.html`.
Checked the script and the matching grid rows, including adjacent cells.

| Component | Rectangle (row, col, width, height) | Board script | Result |
| --- | --- | --- | --- |
| TabBar | (10, 0, 120, 2) | Lines 40–42 write the labels at row 10 and the full-width rule at row 11. | Whole component; caption at row 9 and blank row 12 are excluded. |
| Panel | (17, 0, 40, 5); (17, 44, 40, 5) | Lines 46–47 call box with exactly these dimensions and put body text inside. | All four borders, title, padding, and body included; four-cell gap excluded. |
| Viewport | (24, 0, 60, 7); (24, 64, 56, 7) | Lines 49–52 draw both boxes, five content rows, vertical mark at (26–27, 59), and horizontal mark at (30, 65–106). | Both marks and full borders included; gap and caption excluded. |

## Visual review

Captures: `verifier-board.png`, `verifier-samples.png`,
`verifier-CompactAgents.png`, and `verifier-CompactFriction.png`.
The component sample capture renders actual component ANSI through
`cellgrid.Convert` and then HTML spans, with one span per terminal cell.
The scratch capture program is preserved as `verifier-capture.go.txt`.
This is a component-only check. Screen integration is outside this task.

Board capture command (fresh scratch profile for each capture):

```sh
firefox --headless --no-remote --profile SCRATCH --window-size=1220,1280 --screenshot OUT.png file:///home/penguin/source/hive/.hive/spec/tui-v2/design/Components.dc.html
```

Used height 1280 to include all 67 rows and KeyHints. The two compact boards
used `--window-size=1220,780`. Samples used height 1280. Default browser font
and separate sample sections give different pixel placement from the board;
the cell rectangles are compared by the golden tests. No live agent pane
was read and no live home data was changed.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Palette | Token names and hex labels | matches | Source tokens match decision 26; no standalone palette component required. |
| Header, long | `working`, `live`, `needs you 0`, `14:22` | matches | Text and dots agree with board. |
| Header, short | `coord ●`, `sup ●`, `needs you 0` | matches | Compact board form; `◐` idle and `○` stopped are also tested. |
| TabBar, long and compact | Active `━` underline and numbered labels | matches | Six labels, width breakpoint, and rule agree. |
| SubTabs | `[Tasks]` | matches | Brackets identify active view without colour. |
| Panel | `focused` and `unfocused` sample titles | matches | Border, title, padding, and body agree. Caller must retain a plain-text focus signal when screens use panels. |
| Viewport vertical | `┃`, line numbers 41–45 | matches | Two-cell thumb at right border. |
| Viewport horizontal | `━`, `projects 1–4 of 5` | matches | Forty-two-cell thumb at bottom border. |
| KeyHints | `[graph]` and named action keys | matches | Region and action text agree. |

List, Details, FilterBar, Input, Card, Popup, and Notice on the sheet are
unchanged and outside this task.

## Diff scope

Command: `git diff dev...HEAD --stat` after merge. Only the permitted source
paths and evidence appear. Full output follows.

## Run output

### Base test summary

```text
=== RUN   TestFitFrameRows
--- FAIL: TestFitFrameRows (0.00s)
=== RUN   TestFitPanels
--- FAIL: TestFitPanels (0.02s)
=== RUN   TestHeaderGolden
=== RUN   TestHeaderGolden/long
=== RUN   TestHeaderGolden/short
--- FAIL: TestHeaderGolden (0.00s)
=== RUN   TestHeaderForms
--- FAIL: TestHeaderForms (0.00s)
=== RUN   TestHeaderStateDots
--- FAIL: TestHeaderStateDots (0.00s)
=== RUN   TestHeaderCutsLongHome
--- FAIL: TestHeaderCutsLongHome (0.00s)
=== RUN   TestKeyHintsGolden
=== RUN   TestKeyHintsGolden/sheet
=== RUN   TestKeyHintsGolden/main
=== RUN   TestKeyHintsGolden/compact
--- FAIL: TestKeyHintsGolden (0.00s)
=== RUN   TestKeyHintsFit
--- FAIL: TestKeyHintsFit (0.00s)
=== RUN   TestPanelGolden
=== RUN   TestPanelGolden/unfocused
=== RUN   TestPanelGolden/focused
--- FAIL: TestPanelGolden (0.00s)
=== RUN   TestSubTabsGolden
=== RUN   TestSubTabsGolden/sheet
=== RUN   TestSubTabsGolden/work
=== RUN   TestSubTabsGolden/compact
--- FAIL: TestSubTabsGolden (0.00s)
=== RUN   TestSubTabsText
--- FAIL: TestSubTabsText (0.00s)
=== RUN   TestSymbolWidth
--- PASS: TestSymbolWidth (0.00s)
=== RUN   TestTabBarGolden
=== RUN   TestTabBarGolden/sheet
=== RUN   TestTabBarGolden/work
=== RUN   TestTabBarGolden/compact
--- FAIL: TestTabBarGolden (0.00s)
=== RUN   TestTabBarSpacing
--- FAIL: TestTabBarSpacing (0.00s)
=== RUN   TestViewportGolden
=== RUN   TestViewportGolden/vertical
=== RUN   TestViewportGolden/horizontal
--- FAIL: TestViewportGolden (0.00s)
=== RUN   TestViewportXOffset
--- FAIL: TestViewportXOffset (0.00s)
=== RUN   TestThumb
--- FAIL: TestThumb (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui/component	0.038s
FAIL
```

### Final green test run

```text
=== RUN   TestFitFrameRows
--- PASS: TestFitFrameRows (0.08s)
=== RUN   TestFitPanels
--- PASS: TestFitPanels (0.28s)
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
ok  	github.com/Harrison-Blair/hive/internal/tui/component	0.370s
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
FAIL	github.com/Harrison-Blair/hive/internal/tui/component	0.400s
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

### Full gate output

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
ok  	github.com/Harrison-Blair/hive	13.616s
ok  	github.com/Harrison-Blair/hive/cmd	3.643s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.126s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.279s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.266s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.112s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.704s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.680s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.292s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.066s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.128s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.029s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	9.678s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	(cached)
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.827s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	(cached)
ok  	github.com/Harrison-Blair/hive/internal/homeinit	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.319s
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
ok  	github.com/Harrison-Blair/hive/internal/supervise	3.053s
ok  	github.com/Harrison-Blair/hive/internal/task	(cached)
ok  	github.com/Harrison-Blair/hive/internal/tui	49.269s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	(cached)
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
exit 0
```

### Diff scope output

```text
 .hive/spec/tui-v2/tasks/21-components-1/brief.md   | 144 ++++++++++
 .../tasks/21-components-1/evidence/V-gate.md       | 143 ++++++++++
 .../tui-v2/tasks/21-components-1/evidence/V1.md    | 316 +++++++++++++++++++++
 .../tui-v2/tasks/21-components-1/evidence/V2.md    | 163 +++++++++++
 internal/tui/component/doc.go                      |  14 +
 internal/tui/component/fit_test.go                 |  92 ++++++
 internal/tui/component/golden_test.go              |  70 +++++
 internal/tui/component/header.go                   |  60 ++++
 internal/tui/component/header_test.go              |  53 ++++
 internal/tui/component/keyhints.go                 |  29 ++
 internal/tui/component/keyhints_test.go            |  45 +++
 internal/tui/component/palette.go                  |  23 ++
 internal/tui/component/panel.go                    |  75 +++++
 internal/tui/component/panel_test.go               |  26 ++
 internal/tui/component/scrollmark.go               |  31 ++
 internal/tui/component/subtabs.go                  |  39 +++
 internal/tui/component/subtabs_test.go             |  33 +++
 internal/tui/component/symbols_test.go             |  21 ++
 internal/tui/component/tabbar.go                   |  52 ++++
 internal/tui/component/tabbar_test.go              |  32 +++
 internal/tui/component/viewport.go                 |  27 ++
 internal/tui/component/viewport_test.go            |  63 ++++
 testdata/ste-baseline.json                         |   1 +
 23 files changed, 1552 insertions(+)
```

## Cleanup

The scratch base worktree was removed. Temporary source mutations were restored.
Scratch browser profiles, scripts, and redundant logs were removed after the
captures and results were saved here. No workspace, tab, or pane was created.
