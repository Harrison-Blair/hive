# Verification: 23-components-2

Verdict: **pass**. Verifier: `t23-verifier-2`.

Checked task head `a872c1e0`. `git merge dev --no-edit` printed
`Already up to date.` Base `dev`: `77021a80`.
This verdict replaces the earlier failed verdict. Git retains that report.

| Item | Result | Observation |
| --- | --- | --- |
| V1 Tests | pass | The base build fails on missing task APIs. All task tests pass. A List token mutation fails with cell coordinates. The repaired Chat fit test fails with the old Chat code and passes with the repair. |
| V2 Rectangles | pass | Details, FilterBar, and Card rectangles cover the complete board components and no adjacent component. |
| V-gate Diff scope and gate | pass | Only permitted source files and evidence changed. All five commands exit 0. Formatting and module checks print nothing. |
| Visual check | pass | Fresh PNG captures agree with the component rectangles. Plain text retains state markers. The brief supplies the Popup and Notice exceptions. |

## V1 method

1. Made a detached scratch worktree at `dev`.
2. Copied each new or changed component test file into that worktree.
3. Ran `go test -gcflags=-e ./internal/tui/component/`. Exit 1.
4. Ran `go test -count=1 -v ./internal/tui/component/` on the task. Exit 0.
5. Made a second detached scratch worktree at the task head.
6. Changed only `dimStyle.Render(l.header())` to `accentStyle.Render(l.header())` in its List component.
7. Ran `go test -count=1 ./internal/tui/component/ -run '^TestListGolden$'`. Exit 1, with 40 differing cells.
8. Restored List. Replaced scratch `chat.go` with its version before repair commit `31894f6c`.
9. Ran `go test -count=1 ./internal/tui/component/ -run '^TestFitWave4$'`. Exit 1, with compact heading width failures.
10. Restored Chat. Ran `go test -count=1 ./internal/tui/component/`. Exit 0.

The base cannot compile the new tests because it lacks the task APIs.
This run proves a build failure for the new test suite. It does not reconstruct individual behavioral red runs before implementation.
The separate mutations prove that the List golden check and repaired Chat fit check detect the respective faults.

An independent probe used agent and user messages with names of 80 ASCII characters or 80 double-width characters.
It checked panel widths 80, 119, 120, and 200 with `lipgloss.Width`.
Every returned line fits the inner width. The original 80-column failure is absent.

The Markdown tests use literal expected text and hand-written grids.
They cover headings, strong text, inline code, code blocks, lists, links, and three table width cases.
The Chat golden checks include the new table rows.
Commit `d58090b6` pins Glamour v2.0.1 and gives the required dependency reason.
Decisions 32 and 33 describe the unchanged Glamour table output.
The report `/home/penguin/source/hive/.hive/tmp/reports/tui-v2-glamour-tables.md` exists and is not tracked.

## V2 independent rectangle check

Read the approved board script and each test call site independently.
Coordinates are zero-based. Format: `{row, col, width, height}`.

| Component | Board script sample | Test rectangle | Result |
| --- | --- | --- | --- |
| Details | Components.dc.html lines 60–64: box at x=64, y=33, width=56, height=6. Fields and section fill rows 34–37. | `{33, 64, 56, 6}` | pass |
| FilterBar | Components.dc.html lines 66–67: box at x=0, y=41, width=120, height=3. Sample text and cursor occupy row 42. | `{41, 0, 120, 3}` | pass |
| Card | Components.dc.html lines 71–73: boxes at x=0/26/52, y=51, width=24, height=4. Names and state text occupy rows 52–53. | `{51, 0/26/52, 24, 4}` | pass |

## Visual check

This task changes reusable components. It does not change a wired screen.
Captured actual component ANSI output during a scratch run of the golden tests.
The scratch helper saved the `got` argument before comparison. It did not change the component code or expected grids.
A small script converted ANSI foreground and bold codes into HTML spans in a `pre` element.
Firefox captured that HTML and each approved HTML board with separate scratch profiles.

Commands used this form:

```sh
firefox --headless --no-remote --profile SCRATCH --window-size=1220,900 --screenshot OUT.png file://BOARD.dc.html
```

The Components board used height 1300 to capture its full 67 rows.
The component capture used height 3100 to include every sample.
The saved PNG files are in `visual/` beside this report:

- `components.png`: actual rendered samples, including hand-grid Markdown and Notice samples.
- `Components-board.png`: approved component sheet.
- `AgentDetails-board.png`: approved Popup.
- `Chat-board.png`: approved wide Chat text.
- `CompactChat-board.png`: approved compact Chat text.
- `Work-board.png`: approved List and Details context.

Compared the fresh PNG files visually. Golden tests separately compare each character, token, and bold flag.
The HTML font and page layout differ from the board wrapper. These capture-only differences do not change terminal cells.
Removed ANSI codes from the rendered sample text to check each state marker.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| List | `▶`, task IDs, `landed` | matches | Full List rectangle agrees with Components. |
| List scroll mark | `┃` | matches | Plain-text scroll test and fit tests cover the overflowing state. |
| Details | `status`, `goal`, `ACCEPTANCE`, `✓` | matches | Full Details rectangle agrees with Components. |
| FilterBar | `/`, typed field names, `█` | matches | Edit sample agrees with Components. Applied state has no cursor. |
| Input | `>`, message words, `█` | matches | Components and Chat input rectangles agree. |
| Card working | `●`, `working` | matches | Selected card has the complete 24×4 border. |
| Card idle | `◐`, `idle` | matches | Idle dot uses accent. |
| Card stopped | `○`, `stopped` | matches | Stopped sample agrees with Components. |
| Popup content | `esc closes`, `ACTIONS`, `RECENT`, `◐ live · idle` | matches | AgentDetails rows 7–33, columns 30–89 agree. |
| Popup sample choice | `esc closes` | overridden | Brief requires the AgentDetails rectangle instead of the small Components sample. |
| Popup background | header and tab words remain readable | matches | Brief requires border token behind the box. Hand-grid test checks every background cell and centering. |
| Notice | notice words, truncation `…` | overridden | Brief requires a hand-written dim grid because the board has no sample. |
| Wide Chat text | time, `you`, agent name, event `·` | matches | Gutter, wrapping, user text, and folded event agree with Chat. |
| Compact Chat text | time/name on a separate line | matches | CompactChat sample agrees. Long-name probe also fits. |
| Markdown | heading words, strong words, code text, `•`, link text | matches | Hand grid and Chat sample retain structure without colour. |
| Markdown tables | headers, `│`, `┼`, `─` | matches | Updated Chat boards show Glamour v2.0.1 output. |
| Work context | task IDs and field labels | overridden | AC3 requires Components rectangles. Work screen layout and its padding cells remain for wave 6. |

No state in the samples depends only on colour.
No live coordinator, supervisor, workspace, tab, or pane was created.

## Output

The following sections contain the full command output.
`EXIT:0` is a wrapper status marker, not output from gofmt or go mod tidy.
No test failed in the race gate. No flake rerun was necessary.

### Diff scope

```sh
git diff dev...HEAD --stat
```

```text
 .hive/spec/tui-v2/README.md                        |   4 +-
 .hive/spec/tui-v2/tasks/23-components-2/brief.md   | 177 ++++++
 .../tasks/23-components-2/evidence/V-gate.md       | 269 +++++++++
 .../tui-v2/tasks/23-components-2/evidence/V1.md    | 201 +++++++
 .../tui-v2/tasks/23-components-2/evidence/V2.md    |  46 ++
 .../tasks/23-components-2/evidence/VERIFIED.md     | 604 +++++++++++++++++++++
 go.mod                                             |  11 +
 go.sum                                             |  29 +
 internal/tui/component/card.go                     |  52 ++
 internal/tui/component/card_test.go                |  33 ++
 internal/tui/component/chat.go                     | 135 +++++
 internal/tui/component/chat_test.go                | 123 +++++
 internal/tui/component/details.go                  |  31 ++
 internal/tui/component/details_test.go             |  36 ++
 internal/tui/component/doc.go                      |  10 +
 internal/tui/component/filterbar.go                |  54 ++
 internal/tui/component/filterbar_test.go           |  47 ++
 internal/tui/component/fit_test.go                 |  46 ++
 internal/tui/component/golden_test.go              |  22 +
 internal/tui/component/input.go                    |  34 ++
 internal/tui/component/input_test.go               |  43 ++
 internal/tui/component/list.go                     | 123 +++++
 internal/tui/component/list_test.go                |  75 +++
 internal/tui/component/markdown.go                 | 104 ++++
 internal/tui/component/markdown_style.go           |  55 ++
 internal/tui/component/markdown_table_test.go      |  94 ++++
 internal/tui/component/markdown_test.go            | 106 ++++
 internal/tui/component/notice.go                   |  15 +
 internal/tui/component/notice_test.go              |  24 +
 internal/tui/component/popup.go                    |  64 +++
 internal/tui/component/popup_test.go               |  58 ++
 testdata/ste-baseline.json                         |   1 +
 32 files changed, 2724 insertions(+), 2 deletions(-)
```

### Formatting

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
```

```text
EXIT:0
```

### Modules

```sh
go mod tidy -diff
```

```text
EXIT:0
```

### Vet

```sh
go vet ./...
```

```text
EXIT:0
```

### Lint

```sh
golangci-lint run ./...
```

```text
0 issues.
EXIT:0
```

### Race tests

```sh
go test -race ./...
```

```text
ok  	github.com/Harrison-Blair/hive	31.461s
ok  	github.com/Harrison-Blair/hive/cmd	7.204s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.198s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.542s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.376s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.177s
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
ok  	github.com/Harrison-Blair/hive/cmd/task	1.366s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.097s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.181s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.035s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	16.685s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	(cached)
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	6.054s
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
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.394s
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
ok  	github.com/Harrison-Blair/hive/internal/tui	77.909s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	33.124s
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
EXIT:0
```

### Base red

```sh
go test -gcflags=-e ./internal/tui/component/
```

```text
# github.com/Harrison-Blair/hive/internal/tui/component [github.com/Harrison-Blair/hive/internal/tui/component.test]
internal/tui/component/chat_test.go:23:20: undefined: ChatMessage
internal/tui/component/chat_test.go:24:37: undefined: ChatUser
internal/tui/component/chat_test.go:25:46: undefined: ChatEvent
internal/tui/component/chat_test.go:26:45: undefined: ChatAgent
internal/tui/component/chat_test.go:27:37: undefined: ChatUser
internal/tui/component/chat_test.go:34:16: undefined: ChatLines
internal/tui/component/chat_test.go:59:16: undefined: ChatLines
internal/tui/component/chat_test.go:59:28: undefined: ChatMessage
internal/tui/component/chat_test.go:60:46: undefined: ChatAgent
internal/tui/component/chat_test.go:61:38: undefined: ChatUser
internal/tui/component/chat_test.go:80:12: undefined: ChatMessage
internal/tui/component/chat_test.go:81:38: undefined: ChatUser
internal/tui/component/chat_test.go:82:47: undefined: ChatEvent
internal/tui/component/chat_test.go:83:46: undefined: ChatAgent
internal/tui/component/chat_test.go:85:16: undefined: ChatLines
internal/tui/component/chat_test.go:106:17: undefined: ChatLines
internal/tui/component/chat_test.go:106:29: undefined: ChatMessage
internal/tui/component/chat_test.go:106:76: undefined: ChatUser
internal/tui/component/list_test.go:14:19: undefined: List
internal/tui/component/list_test.go:17:41: undefined: Cell
internal/tui/component/list_test.go:18:12: undefined: Cell
internal/tui/component/list_test.go:20:9: undefined: List
internal/tui/component/list_test.go:23:14: undefined: Column
internal/tui/component/list_test.go:24:13: undefined: Cell
internal/tui/component/list_test.go:41:7: undefined: List
internal/tui/component/list_test.go:41:23: undefined: Column
internal/tui/component/list_test.go:43:29: undefined: Cell
internal/tui/component/list_test.go:62:7: undefined: List
internal/tui/component/list_test.go:62:23: undefined: Column
internal/tui/component/list_test.go:64:29: undefined: Cell
internal/tui/component/card_test.go:13:8: undefined: Card
internal/tui/component/card_test.go:16:4: undefined: Card
internal/tui/component/card_test.go:17:4: undefined: Card
internal/tui/component/card_test.go:18:4: undefined: Card
internal/tui/component/card_test.go:28:7: undefined: Card
internal/tui/component/details_test.go:15:7: undefined: Details
internal/tui/component/details_test.go:16:3: undefined: DetailField
internal/tui/component/details_test.go:17:3: undefined: DetailField
internal/tui/component/details_test.go:18:3: undefined: DetailSection
internal/tui/component/details_test.go:27:7: undefined: Details
internal/tui/component/filterbar_test.go:14:7: undefined: FilterBar
internal/tui/component/filterbar_test.go:22:7: undefined: FilterBar
internal/tui/component/filterbar_test.go:42:7: undefined: FilterBar
internal/tui/component/fit_test.go:101:10: undefined: List
internal/tui/component/fit_test.go:101:54: undefined: Column
internal/tui/component/fit_test.go:102:14: undefined: Details
internal/tui/component/fit_test.go:104:35: undefined: Cell
internal/tui/component/fit_test.go:105:41: undefined: DetailField
internal/tui/component/fit_test.go:113:31: undefined: FilterBar
internal/tui/component/fit_test.go:114:27: undefined: Input
internal/tui/component/fit_test.go:115:26: undefined: Notice
internal/tui/component/fit_test.go:118:26: undefined: Popup
internal/tui/component/fit_test.go:121:25: undefined: Card
internal/tui/component/fit_test.go:121:92: undefined: CardWidth
internal/tui/component/fit_test.go:121:103: undefined: CardHeight
internal/tui/component/fit_test.go:122:17: undefined: ChatMessage
internal/tui/component/fit_test.go:123:37: undefined: ChatEvent
internal/tui/component/fit_test.go:124:37: undefined: ChatAgent
internal/tui/component/fit_test.go:125:37: undefined: ChatUser
internal/tui/component/fit_test.go:128:17: undefined: ChatLines
internal/tui/component/input_test.go:14:18: undefined: Input
internal/tui/component/input_test.go:17:18: undefined: Input
internal/tui/component/input_test.go:24:9: undefined: Input
internal/tui/component/input_test.go:38:35: undefined: Input
internal/tui/component/markdown_table_test.go:18:16: undefined: RenderMarkdown
internal/tui/component/markdown_table_test.go:89:16: undefined: RenderMarkdown
internal/tui/component/markdown_test.go:51:16: undefined: RenderMarkdown
internal/tui/component/markdown_test.go:70:16: undefined: RenderMarkdown
internal/tui/component/markdown_test.go:81:17: undefined: RenderMarkdown
internal/tui/component/markdown_test.go:95:16: undefined: RenderMarkdown
internal/tui/component/notice_test.go:13:19: undefined: Notice
internal/tui/component/notice_test.go:23:19: undefined: Notice
internal/tui/component/popup_test.go:18:25: undefined: DetailField
internal/tui/component/popup_test.go:20:28: undefined: DetailSection
internal/tui/component/popup_test.go:24:11: undefined: DetailSection
internal/tui/component/popup_test.go:25:3: undefined: fitCell
internal/tui/component/popup_test.go:27:36: undefined: fitCell
internal/tui/component/popup_test.go:29:7: undefined: Popup
internal/tui/component/popup_test.go:29:22: undefined: Details
internal/tui/component/popup_test.go:39:7: undefined: Popup
internal/tui/component/popup_test.go:39:22: undefined: Details
internal/tui/component/popup_test.go:57:19: undefined: DimScreen
FAIL	github.com/Harrison-Blair/hive/internal/tui/component [build failed]
FAIL
```

### Task green

```sh
go test -count=1 -v ./internal/tui/component/
```

```text
=== RUN   TestCardGolden
=== RUN   TestCardGolden/working
=== RUN   TestCardGolden/idle
=== RUN   TestCardGolden/stopped
--- PASS: TestCardGolden (0.01s)
    --- PASS: TestCardGolden/working (0.00s)
    --- PASS: TestCardGolden/idle (0.00s)
    --- PASS: TestCardGolden/stopped (0.00s)
=== RUN   TestCardCutsLongText
--- PASS: TestCardCutsLongText (0.00s)
=== RUN   TestChatGolden
--- PASS: TestChatGolden (0.01s)
=== RUN   TestCompactChatGolden
--- PASS: TestCompactChatGolden (0.00s)
=== RUN   TestChatNarrow
--- PASS: TestChatNarrow (0.00s)
=== RUN   TestChatWrapWidth
--- PASS: TestChatWrapWidth (0.00s)
=== RUN   TestDetailsGolden
--- PASS: TestDetailsGolden (0.00s)
=== RUN   TestDetailsScrolls
--- PASS: TestDetailsScrolls (0.00s)
=== RUN   TestFilterBarGolden
--- PASS: TestFilterBarGolden (0.00s)
=== RUN   TestFilterBarFields
--- PASS: TestFilterBarFields (0.00s)
=== RUN   TestFilterBarLongTextKeepsCursor
--- PASS: TestFilterBarLongTextKeepsCursor (0.00s)
=== RUN   TestFitFrameRows
--- PASS: TestFitFrameRows (0.27s)
=== RUN   TestFitPanels
--- PASS: TestFitPanels (0.70s)
=== RUN   TestFitWave4
--- PASS: TestFitWave4 (2.03s)
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
=== RUN   TestInputGolden
=== RUN   TestInputGolden/components
=== RUN   TestInputGolden/chat
--- PASS: TestInputGolden (0.00s)
    --- PASS: TestInputGolden/components (0.00s)
    --- PASS: TestInputGolden/chat (0.00s)
=== RUN   TestInputUnfocused
--- PASS: TestInputUnfocused (0.00s)
=== RUN   TestInputShowsEndOfText
--- PASS: TestInputShowsEndOfText (0.00s)
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
=== RUN   TestListGolden
--- PASS: TestListGolden (0.00s)
=== RUN   TestListKeepsSelectionVisible
--- PASS: TestListKeepsSelectionVisible (0.00s)
=== RUN   TestListScrollMark
--- PASS: TestListScrollMark (0.00s)
=== RUN   TestMarkdownTableUneven
--- PASS: TestMarkdownTableUneven (0.00s)
=== RUN   TestMarkdownTableEqualColumns
--- PASS: TestMarkdownTableEqualColumns (0.00s)
=== RUN   TestMarkdownTableTooManyColumns
--- PASS: TestMarkdownTableTooManyColumns (0.00s)
=== RUN   TestMarkdownTableTokens
--- PASS: TestMarkdownTableTokens (0.00s)
=== RUN   TestMarkdownGrid
--- PASS: TestMarkdownGrid (0.00s)
=== RUN   TestMarkdownAutoLinks
--- PASS: TestMarkdownAutoLinks (0.00s)
=== RUN   TestMarkdownRemovesBlankFirstLine
--- PASS: TestMarkdownRemovesBlankFirstLine (0.00s)
=== RUN   TestMarkdownListFitsWidth
--- PASS: TestMarkdownListFitsWidth (0.00s)
=== RUN   TestNoticeGrid
--- PASS: TestNoticeGrid (0.00s)
=== RUN   TestNoticeCutsLongText
--- PASS: TestNoticeCutsLongText (0.00s)
=== RUN   TestPanelGolden
=== RUN   TestPanelGolden/unfocused
=== RUN   TestPanelGolden/focused
--- PASS: TestPanelGolden (0.00s)
    --- PASS: TestPanelGolden/unfocused (0.00s)
    --- PASS: TestPanelGolden/focused (0.00s)
=== RUN   TestPopupGolden
--- PASS: TestPopupGolden (0.00s)
=== RUN   TestPopupDimsScreen
--- PASS: TestPopupDimsScreen (0.00s)
=== RUN   TestDimScreen
--- PASS: TestDimScreen (0.00s)
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
ok  	github.com/Harrison-Blair/hive/internal/tui/component	3.075s
```

### Token mutation red

```sh
go test -count=1 ./internal/tui/component/ -run '^TestListGolden$'
```

```text
--- FAIL: TestListGolden (0.00s)
    list_test.go:35: 40 cells differ from Components.grid (board row and column):
        row 34 col 4: want 'I' dim normal, got 'I' accent normal
        row 34 col 5: want 'D' dim normal, got 'D' accent normal
        row 34 col 6: want ' ' dim normal, got ' ' accent normal
        row 34 col 7: want ' ' dim normal, got ' ' accent normal
        row 34 col 8: want ' ' dim normal, got ' ' accent normal
        row 34 col 9: want ' ' dim normal, got ' ' accent normal
        row 34 col 10: want ' ' dim normal, got ' ' accent normal
        row 34 col 11: want ' ' dim normal, got ' ' accent normal
        row 34 col 12: want ' ' dim normal, got ' ' accent normal
        row 34 col 13: want ' ' dim normal, got ' ' accent normal
        row 34 col 14: want ' ' dim normal, got ' ' accent normal
        row 34 col 15: want ' ' dim normal, got ' ' accent normal
        row 34 col 16: want ' ' dim normal, got ' ' accent normal
        row 34 col 17: want ' ' dim normal, got ' ' accent normal
        row 34 col 18: want ' ' dim normal, got ' ' accent normal
        row 34 col 19: want ' ' dim normal, got ' ' accent normal
        row 34 col 20: want ' ' dim normal, got ' ' accent normal
        row 34 col 21: want ' ' dim normal, got ' ' accent normal
        row 34 col 22: want ' ' dim normal, got ' ' accent normal
        row 34 col 23: want ' ' dim normal, got ' ' accent normal
        row 34 col 24: want ' ' dim normal, got ' ' accent normal
        row 34 col 25: want ' ' dim normal, got ' ' accent normal
        row 34 col 26: want ' ' dim normal, got ' ' accent normal
        row 34 col 27: want ' ' dim normal, got ' ' accent normal
        row 34 col 28: want ' ' dim normal, got ' ' accent normal
        row 34 col 29: want ' ' dim normal, got ' ' accent normal
        row 34 col 30: want 'S' dim normal, got 'S' accent normal
        row 34 col 31: want 'T' dim normal, got 'T' accent normal
        row 34 col 32: want 'A' dim normal, got 'A' accent normal
        row 34 col 33: want 'T' dim normal, got 'T' accent normal
        row 34 col 34: want 'U' dim normal, got 'U' accent normal
        row 34 col 35: want 'S' dim normal, got 'S' accent normal
        row 34 col 36: want ' ' dim normal, got ' ' accent normal
        row 34 col 37: want ' ' dim normal, got ' ' accent normal
        row 34 col 38: want ' ' dim normal, got ' ' accent normal
        row 34 col 39: want ' ' dim normal, got ' ' accent normal
        row 34 col 40: want 'W' dim normal, got 'W' accent normal
        row 34 col 41: want 'A' dim normal, got 'A' accent normal
        row 34 col 42: want 'V' dim normal, got 'V' accent normal
        row 34 col 43: want 'E' dim normal, got 'E' accent normal
        
        rendered:
        [38;2;245;185;66m╭[m[38;2;245;185;66m─[m[1;38;2;245;185;66m tasks · 4 [m[38;2;245;185;66m──────────────────────────────────────────────[m[38;2;245;185;66m╮[m
        [38;2;245;185;66m│[m   [38;2;245;185;66mID                        STATUS    WAVE[m               [38;2;245;185;66m│[m
        [38;2;245;185;66m│[m   4-agents-equal-columns    [38;2;124;207;106mlanded[m    [38;2;163;157;144mC[m                  [38;2;245;185;66m│[m
        [38;2;245;185;66m│[m [1;38;2;245;185;66m▶ [m[1;38;2;245;185;66m3-tui-restart            [m [38;2;124;207;106mlanded[m    [38;2;163;157;144mB[m                  [38;2;245;185;66m│[m
        [38;2;245;185;66m│[m   2-stale-coordinator-star… [38;2;124;207;106mlanded[m    [38;2;163;157;144mA[m                  [38;2;245;185;66m│[m
        [38;2;245;185;66m╰[m[38;2;245;185;66m[m[38;2;245;185;66m──────────────────────────────────────────────────────────╯[m
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui/component	0.025s
FAIL
```

### Repair regression red

```sh
go test -count=1 ./internal/tui/component/ -run '^TestFitWave4$'
```

```text
--- FAIL: TestFitWave4 (1.86s)
    fit_test.go:134: Chat at 30: line 54 is 306 cells, inner width 26
    fit_test.go:134: Chat at 30: line 57 is 306 cells, inner width 26
    fit_test.go:134: Chat at 31: line 54 is 306 cells, inner width 27
    fit_test.go:134: Chat at 31: line 57 is 306 cells, inner width 27
    fit_test.go:134: Chat at 32: line 54 is 306 cells, inner width 28
    fit_test.go:134: Chat at 32: line 57 is 306 cells, inner width 28
    fit_test.go:134: Chat at 33: line 54 is 306 cells, inner width 29
    fit_test.go:134: Chat at 33: line 57 is 306 cells, inner width 29
    fit_test.go:134: Chat at 34: line 53 is 306 cells, inner width 30
    fit_test.go:134: Chat at 34: line 56 is 306 cells, inner width 30
    fit_test.go:134: Chat at 35: line 53 is 306 cells, inner width 31
    fit_test.go:134: Chat at 35: line 56 is 306 cells, inner width 31
    fit_test.go:134: Chat at 36: line 52 is 306 cells, inner width 32
    fit_test.go:134: Chat at 36: line 55 is 306 cells, inner width 32
    fit_test.go:134: Chat at 37: line 51 is 306 cells, inner width 33
    fit_test.go:134: Chat at 37: line 54 is 306 cells, inner width 33
    fit_test.go:134: Chat at 38: line 50 is 306 cells, inner width 34
    fit_test.go:134: Chat at 38: line 53 is 306 cells, inner width 34
    fit_test.go:134: Chat at 39: line 50 is 306 cells, inner width 35
    fit_test.go:134: Chat at 39: line 53 is 306 cells, inner width 35
    fit_test.go:134: Chat at 40: line 49 is 306 cells, inner width 36
    fit_test.go:134: Chat at 40: line 52 is 306 cells, inner width 36
    fit_test.go:134: Chat at 41: line 49 is 306 cells, inner width 37
    fit_test.go:134: Chat at 41: line 52 is 306 cells, inner width 37
    fit_test.go:134: Chat at 42: line 47 is 306 cells, inner width 38
    fit_test.go:134: Chat at 42: line 50 is 306 cells, inner width 38
    fit_test.go:134: Chat at 43: line 46 is 306 cells, inner width 39
    fit_test.go:134: Chat at 43: line 49 is 306 cells, inner width 39
    fit_test.go:134: Chat at 44: line 46 is 306 cells, inner width 40
    fit_test.go:134: Chat at 44: line 49 is 306 cells, inner width 40
    fit_test.go:134: Chat at 45: line 46 is 306 cells, inner width 41
    fit_test.go:134: Chat at 45: line 49 is 306 cells, inner width 41
    fit_test.go:134: Chat at 46: line 44 is 306 cells, inner width 42
    fit_test.go:134: Chat at 46: line 47 is 306 cells, inner width 42
    fit_test.go:134: Chat at 47: line 44 is 306 cells, inner width 43
    fit_test.go:134: Chat at 47: line 47 is 306 cells, inner width 43
    fit_test.go:134: Chat at 48: line 44 is 306 cells, inner width 44
    fit_test.go:134: Chat at 48: line 47 is 306 cells, inner width 44
    fit_test.go:134: Chat at 49: line 41 is 306 cells, inner width 45
    fit_test.go:134: Chat at 49: line 44 is 306 cells, inner width 45
    fit_test.go:134: Chat at 50: line 41 is 306 cells, inner width 46
    fit_test.go:134: Chat at 50: line 44 is 306 cells, inner width 46
    fit_test.go:134: Chat at 51: line 41 is 306 cells, inner width 47
    fit_test.go:134: Chat at 51: line 44 is 306 cells, inner width 47
    fit_test.go:134: Chat at 52: line 39 is 306 cells, inner width 48
    fit_test.go:134: Chat at 52: line 42 is 306 cells, inner width 48
    fit_test.go:134: Chat at 53: line 38 is 306 cells, inner width 49
    fit_test.go:134: Chat at 53: line 41 is 306 cells, inner width 49
    fit_test.go:134: Chat at 54: line 38 is 306 cells, inner width 50
    fit_test.go:134: Chat at 54: line 41 is 306 cells, inner width 50
    fit_test.go:134: Chat at 55: line 37 is 306 cells, inner width 51
    fit_test.go:134: Chat at 55: line 40 is 306 cells, inner width 51
    fit_test.go:134: Chat at 56: line 37 is 306 cells, inner width 52
    fit_test.go:134: Chat at 56: line 40 is 306 cells, inner width 52
    fit_test.go:134: Chat at 57: line 37 is 306 cells, inner width 53
    fit_test.go:134: Chat at 57: line 40 is 306 cells, inner width 53
    fit_test.go:134: Chat at 58: line 37 is 306 cells, inner width 54
    fit_test.go:134: Chat at 58: line 40 is 306 cells, inner width 54
    fit_test.go:134: Chat at 59: line 37 is 306 cells, inner width 55
    fit_test.go:134: Chat at 59: line 40 is 306 cells, inner width 55
    fit_test.go:134: Chat at 60: line 37 is 306 cells, inner width 56
    fit_test.go:134: Chat at 60: line 40 is 306 cells, inner width 56
    fit_test.go:134: Chat at 61: line 36 is 306 cells, inner width 57
    fit_test.go:134: Chat at 61: line 39 is 306 cells, inner width 57
    fit_test.go:134: Chat at 62: line 36 is 306 cells, inner width 58
    fit_test.go:134: Chat at 62: line 39 is 306 cells, inner width 58
    fit_test.go:134: Chat at 63: line 36 is 306 cells, inner width 59
    fit_test.go:134: Chat at 63: line 39 is 306 cells, inner width 59
    fit_test.go:134: Chat at 64: line 36 is 306 cells, inner width 60
    fit_test.go:134: Chat at 64: line 39 is 306 cells, inner width 60
    fit_test.go:134: Chat at 65: line 36 is 306 cells, inner width 61
    fit_test.go:134: Chat at 65: line 39 is 306 cells, inner width 61
    fit_test.go:134: Chat at 66: line 36 is 306 cells, inner width 62
    fit_test.go:134: Chat at 66: line 39 is 306 cells, inner width 62
    fit_test.go:134: Chat at 67: line 35 is 306 cells, inner width 63
    fit_test.go:134: Chat at 67: line 38 is 306 cells, inner width 63
    fit_test.go:134: Chat at 68: line 35 is 306 cells, inner width 64
    fit_test.go:134: Chat at 68: line 38 is 306 cells, inner width 64
    fit_test.go:134: Chat at 69: line 34 is 306 cells, inner width 65
    fit_test.go:134: Chat at 69: line 37 is 306 cells, inner width 65
    fit_test.go:134: Chat at 70: line 34 is 306 cells, inner width 66
    fit_test.go:134: Chat at 70: line 37 is 306 cells, inner width 66
    fit_test.go:134: Chat at 71: line 34 is 306 cells, inner width 67
    fit_test.go:134: Chat at 71: line 37 is 306 cells, inner width 67
    fit_test.go:134: Chat at 72: line 33 is 306 cells, inner width 68
    fit_test.go:134: Chat at 72: line 36 is 306 cells, inner width 68
    fit_test.go:134: Chat at 73: line 32 is 306 cells, inner width 69
    fit_test.go:134: Chat at 73: line 35 is 306 cells, inner width 69
    fit_test.go:134: Chat at 74: line 32 is 306 cells, inner width 70
    fit_test.go:134: Chat at 74: line 35 is 306 cells, inner width 70
    fit_test.go:134: Chat at 75: line 32 is 306 cells, inner width 71
    fit_test.go:134: Chat at 75: line 35 is 306 cells, inner width 71
    fit_test.go:134: Chat at 76: line 32 is 306 cells, inner width 72
    fit_test.go:134: Chat at 76: line 35 is 306 cells, inner width 72
    fit_test.go:134: Chat at 77: line 32 is 306 cells, inner width 73
    fit_test.go:134: Chat at 77: line 35 is 306 cells, inner width 73
    fit_test.go:134: Chat at 78: line 32 is 306 cells, inner width 74
    fit_test.go:134: Chat at 78: line 35 is 306 cells, inner width 74
    fit_test.go:134: Chat at 79: line 32 is 306 cells, inner width 75
    fit_test.go:134: Chat at 79: line 35 is 306 cells, inner width 75
    fit_test.go:134: Chat at 80: line 32 is 306 cells, inner width 76
    fit_test.go:134: Chat at 80: line 35 is 306 cells, inner width 76
    fit_test.go:134: Chat at 81: line 32 is 306 cells, inner width 77
    fit_test.go:134: Chat at 81: line 35 is 306 cells, inner width 77
    fit_test.go:134: Chat at 82: line 32 is 306 cells, inner width 78
    fit_test.go:134: Chat at 82: line 35 is 306 cells, inner width 78
    fit_test.go:134: Chat at 83: line 32 is 306 cells, inner width 79
    fit_test.go:134: Chat at 83: line 35 is 306 cells, inner width 79
    fit_test.go:134: Chat at 84: line 30 is 306 cells, inner width 80
    fit_test.go:134: Chat at 84: line 33 is 306 cells, inner width 80
    fit_test.go:134: Chat at 85: line 30 is 306 cells, inner width 81
    fit_test.go:134: Chat at 85: line 33 is 306 cells, inner width 81
    fit_test.go:134: Chat at 86: line 30 is 306 cells, inner width 82
    fit_test.go:134: Chat at 86: line 33 is 306 cells, inner width 82
    fit_test.go:134: Chat at 87: line 30 is 306 cells, inner width 83
    fit_test.go:134: Chat at 87: line 33 is 306 cells, inner width 83
    fit_test.go:134: Chat at 88: line 30 is 306 cells, inner width 84
    fit_test.go:134: Chat at 88: line 33 is 306 cells, inner width 84
    fit_test.go:134: Chat at 89: line 30 is 306 cells, inner width 85
    fit_test.go:134: Chat at 89: line 33 is 306 cells, inner width 85
    fit_test.go:134: Chat at 90: line 30 is 306 cells, inner width 86
    fit_test.go:134: Chat at 90: line 33 is 306 cells, inner width 86
    fit_test.go:134: Chat at 91: line 30 is 306 cells, inner width 87
    fit_test.go:134: Chat at 91: line 33 is 306 cells, inner width 87
    fit_test.go:134: Chat at 92: line 30 is 306 cells, inner width 88
    fit_test.go:134: Chat at 92: line 33 is 306 cells, inner width 88
    fit_test.go:134: Chat at 93: line 30 is 306 cells, inner width 89
    fit_test.go:134: Chat at 93: line 33 is 306 cells, inner width 89
    fit_test.go:134: Chat at 94: line 30 is 306 cells, inner width 90
    fit_test.go:134: Chat at 94: line 33 is 306 cells, inner width 90
    fit_test.go:134: Chat at 95: line 30 is 306 cells, inner width 91
    fit_test.go:134: Chat at 95: line 33 is 306 cells, inner width 91
    fit_test.go:134: Chat at 96: line 30 is 306 cells, inner width 92
    fit_test.go:134: Chat at 96: line 33 is 306 cells, inner width 92
    fit_test.go:134: Chat at 97: line 29 is 306 cells, inner width 93
    fit_test.go:134: Chat at 97: line 32 is 306 cells, inner width 93
    fit_test.go:134: Chat at 98: line 29 is 306 cells, inner width 94
    fit_test.go:134: Chat at 98: line 32 is 306 cells, inner width 94
    fit_test.go:134: Chat at 99: line 28 is 306 cells, inner width 95
    fit_test.go:134: Chat at 99: line 31 is 306 cells, inner width 95
    fit_test.go:134: Chat at 100: line 28 is 306 cells, inner width 96
    fit_test.go:134: Chat at 100: line 31 is 306 cells, inner width 96
    fit_test.go:134: Chat at 101: line 28 is 306 cells, inner width 97
    fit_test.go:134: Chat at 101: line 31 is 306 cells, inner width 97
    fit_test.go:134: Chat at 102: line 28 is 306 cells, inner width 98
    fit_test.go:134: Chat at 102: line 31 is 306 cells, inner width 98
    fit_test.go:134: Chat at 103: line 28 is 306 cells, inner width 99
    fit_test.go:134: Chat at 103: line 31 is 306 cells, inner width 99
    fit_test.go:134: Chat at 104: line 28 is 306 cells, inner width 100
    fit_test.go:134: Chat at 104: line 31 is 306 cells, inner width 100
    fit_test.go:134: Chat at 105: line 28 is 306 cells, inner width 101
    fit_test.go:134: Chat at 105: line 31 is 306 cells, inner width 101
    fit_test.go:134: Chat at 106: line 28 is 306 cells, inner width 102
    fit_test.go:134: Chat at 106: line 31 is 306 cells, inner width 102
    fit_test.go:134: Chat at 107: line 28 is 306 cells, inner width 103
    fit_test.go:134: Chat at 107: line 31 is 306 cells, inner width 103
    fit_test.go:134: Chat at 108: line 28 is 306 cells, inner width 104
    fit_test.go:134: Chat at 108: line 31 is 306 cells, inner width 104
    fit_test.go:134: Chat at 109: line 28 is 306 cells, inner width 105
    fit_test.go:134: Chat at 109: line 31 is 306 cells, inner width 105
    fit_test.go:134: Chat at 110: line 28 is 306 cells, inner width 106
    fit_test.go:134: Chat at 110: line 31 is 306 cells, inner width 106
    fit_test.go:134: Chat at 111: line 28 is 306 cells, inner width 107
    fit_test.go:134: Chat at 111: line 31 is 306 cells, inner width 107
    fit_test.go:134: Chat at 112: line 28 is 306 cells, inner width 108
    fit_test.go:134: Chat at 112: line 31 is 306 cells, inner width 108
    fit_test.go:134: Chat at 113: line 28 is 306 cells, inner width 109
    fit_test.go:134: Chat at 113: line 31 is 306 cells, inner width 109
    fit_test.go:134: Chat at 114: line 28 is 306 cells, inner width 110
    fit_test.go:134: Chat at 114: line 31 is 306 cells, inner width 110
    fit_test.go:134: Chat at 115: line 28 is 306 cells, inner width 111
    fit_test.go:134: Chat at 115: line 31 is 306 cells, inner width 111
    fit_test.go:134: Chat at 116: line 28 is 306 cells, inner width 112
    fit_test.go:134: Chat at 116: line 31 is 306 cells, inner width 112
    fit_test.go:134: Chat at 117: line 28 is 306 cells, inner width 113
    fit_test.go:134: Chat at 117: line 31 is 306 cells, inner width 113
    fit_test.go:134: Chat at 118: line 28 is 306 cells, inner width 114
    fit_test.go:134: Chat at 118: line 31 is 306 cells, inner width 114
    fit_test.go:134: Chat at 119: line 28 is 306 cells, inner width 115
    fit_test.go:134: Chat at 119: line 31 is 306 cells, inner width 115
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui/component	1.888s
FAIL
```

### Restored task green

```sh
go test -count=1 ./internal/tui/component/
```

```text
ok  	github.com/Harrison-Blair/hive/internal/tui/component	2.629s
```

### Independent width probe

```sh
go run .hive/tmp/verify23/probe.go
```

```text
name cells 80, kind 1, panel 80, inner 76, maximum line 76
name cells 80, kind 1, panel 119, inner 115, maximum line 86
name cells 80, kind 1, panel 120, inner 116, maximum line 23
name cells 80, kind 1, panel 200, inner 196, maximum line 23
name cells 80, kind 0, panel 80, inner 76, maximum line 76
name cells 80, kind 0, panel 119, inner 115, maximum line 86
name cells 80, kind 0, panel 120, inner 116, maximum line 23
name cells 80, kind 0, panel 200, inner 196, maximum line 23
name cells 160, kind 1, panel 80, inner 76, maximum line 75
name cells 160, kind 1, panel 119, inner 115, maximum line 115
name cells 160, kind 1, panel 120, inner 116, maximum line 23
name cells 160, kind 1, panel 200, inner 196, maximum line 23
name cells 160, kind 0, panel 80, inner 76, maximum line 75
name cells 160, kind 0, panel 119, inner 115, maximum line 115
name cells 160, kind 0, panel 120, inner 116, maximum line 23
name cells 160, kind 0, panel 200, inner 196, maximum line 23
```

## Cleanup

Removed both scratch worktrees, the capture helper, the probe, logs, and Firefox profiles after recording the results.
Kept the final PNG files and plain-text samples in `visual/`. No task source file changed during verification.
