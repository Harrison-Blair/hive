# Verification: 23-components-2

Verdict: **fail**. Verifier: `t23-verifier-1`.

Checked task code at `5d61868d`, after `git merge dev --no-edit`.
Base: `dev` at `77021a80`. No source repair was made.

| Item | Result | Observation |
| --- | --- | --- |
| V1 Tests, AC1–AC5 | fail | Base build is red, task tests are green, and a token mutation fails the List golden check. An independent input fails AC5: the compact Chat heading exceeds its width. The fit test does not cover a long name on an agent or user message. |
| V2 Rectangles | pass | List, FilterBar, and Card rectangles agree with the board script. Each covers the whole component and no other component. |
| V-gate Diff scope and gate | pass | Diff is within the source table and evidence. All five commands exit 0. Formatting and module checks print nothing. |
| Visual check | fail | PNG capture and comparison were not completed. The brief requires a stop and report when code disagrees with the brief. The AC5 failure was reported to manager-hive, inbox 119. No visual pass is claimed. |

## AC5 failure

`internal/tui/component/chat.go` builds the compact heading from the time and
name without a width limit. At panel width 80, its text area has 76 cells.
This input produces an 86-cell heading:

```go
component.ChatLines([]component.ChatMessage{{
    Time: "09:00", Name: strings.Repeat("a", 80),
    Kind: component.ChatAgent, Text: "ok",
}}, 80)
```

The independent probe measured every returned line with `lipgloss.Width`.
Expected: no line exceeds 76 cells. Actual: the first line is 86 cells.
The brief sets no maximum name length. The current fit test supplies a long
name only to an event, whose separate rendering path truncates its heading.
A repair needs a regression check for agent and user headings at compact
widths. No change to a board is needed to establish this width property.

## V1 method

Made a detached scratch worktree at `dev`. Copied all new and changed
component test files into it. Ran:

```sh
go test -gcflags=-e ./internal/tui/component/
```

Exit 1: missing task APIs throughout the new tests. This establishes the
base build failure, not individual behavioral red runs. The worker's earlier
test-first chronology cannot be independently reconstructed by this run.

On task code, ran:

```sh
go test -count=1 -v ./internal/tui/component/
```

Exit 0. In a separate scratch worktree at the task head, changed only the
List header token from `dimStyle` to `accentStyle`, then ran:

```sh
go test -count=1 ./internal/tui/component/ -run '^TestListGolden$'
```

Exit 1: 40 differing cells, with coordinates, expected token and actual token.
Both scratch worktrees were removed. The task source was never mutated.

The dependency commit `d58090b6` gives the required Glamour reason and pins
v2.0.1. Decisions 32 and 33 are present in the spec. Existing hand-grid tests
cover markdown style, and the Chat golden tests include table rows.

## V2 independent rectangle check

Read `Components.dc.html` and the test call sites, without deriving the
expected rectangles from the component code. Coordinates are zero-based.

| Component | Board script | Test rectangle (row, col, width, height) | Result |
| --- | --- | --- | --- |
| List | Lines 54–58: box at x=0, y=33, w=60, h=6; sample rows 34–37 | `{33, 0, 60, 6}` | pass |
| FilterBar | Lines 66–67: box at x=0, y=41, w=120, h=3; sample row 42 | `{41, 0, 120, 3}` | pass |
| Card | Lines 71–73: three boxes at x=0/26/52, y=51, w=24, h=4; body rows 52–53 | `{51, 0/26/52, 24, 4}` | pass |

Notice: V2.md lists an 11×1 Notice hand grid, but the actual test uses 50×1
for the restart notice and 12×1 for truncation. This is an evidence typo.

## State without colour

The following observations come from source and passing cell-grid tests.
They are not a substitute for the incomplete PNG comparison.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| List | `▶` selected marker | matches | List golden check |
| Details | field names and `ACCEPTANCE` | matches | Details golden check |
| FilterBar | `/` and `█` | matches | FilterBar golden check |
| Input | `>` and `█` | matches | Input golden check |
| Card | `● working`, `◐ idle`, `○ stopped` | matches | Three state golden checks |
| Popup | `esc closes`, action names | matches | AgentDetails rectangle; brief overrides the Components sample |
| Notice | notice words | overridden | Brief requires a hand-written dim grid |
| Chat | time/name, event `·` | differs | Compact agent heading can exceed panel width |
| Markdown | table rules, bullets, link text | matches | Hand grids and Chat golden checks |

## Output

Full gate output and independent test output follow. `EXIT:0` is the command
wrapper status marker, not output from gofmt or go mod tidy.

### Diff scope

```sh
git diff dev...HEAD --stat
```

```text
 .hive/spec/tui-v2/README.md                        |   4 +-
 .hive/spec/tui-v2/tasks/23-components-2/brief.md   | 177 +++++++++++++++++++++
 .../tasks/23-components-2/evidence/V-gate.md       | 136 ++++++++++++++++
 .../tui-v2/tasks/23-components-2/evidence/V1.md    | 153 ++++++++++++++++++
 .../tui-v2/tasks/23-components-2/evidence/V2.md    |  46 ++++++
 go.mod                                             |  11 ++
 go.sum                                             |  29 ++++
 internal/tui/component/card.go                     |  52 ++++++
 internal/tui/component/card_test.go                |  33 ++++
 internal/tui/component/chat.go                     | 134 ++++++++++++++++
 internal/tui/component/chat_test.go                | 123 ++++++++++++++
 internal/tui/component/details.go                  |  31 ++++
 internal/tui/component/details_test.go             |  36 +++++
 internal/tui/component/doc.go                      |  10 ++
 internal/tui/component/filterbar.go                |  54 +++++++
 internal/tui/component/filterbar_test.go           |  47 ++++++
 internal/tui/component/fit_test.go                 |  41 +++++
 internal/tui/component/golden_test.go              |  22 +++
 internal/tui/component/input.go                    |  34 ++++
 internal/tui/component/input_test.go               |  43 +++++
 internal/tui/component/list.go                     | 123 ++++++++++++++
 internal/tui/component/list_test.go                |  75 +++++++++
 internal/tui/component/markdown.go                 | 104 ++++++++++++
 internal/tui/component/markdown_style.go           |  55 +++++++
 internal/tui/component/markdown_table_test.go      |  94 +++++++++++
 internal/tui/component/markdown_test.go            | 106 ++++++++++++
 internal/tui/component/notice.go                   |  15 ++
 internal/tui/component/notice_test.go              |  24 +++
 internal/tui/component/popup.go                    |  64 ++++++++
 internal/tui/component/popup_test.go               |  58 +++++++
 testdata/ste-baseline.json                         |   1 +
 31 files changed, 1933 insertions(+), 2 deletions(-)
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
ok  	github.com/Harrison-Blair/hive	19.355s
ok  	github.com/Harrison-Blair/hive/cmd	6.847s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.143s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.378s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.336s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.112s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.976s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.987s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.350s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.070s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.146s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.030s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	13.719s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.518s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	1.467s
ok  	github.com/Harrison-Blair/hive/internal/doctor	5.068s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	2.429s
ok  	github.com/Harrison-Blair/hive/internal/homeinit	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.349s
ok  	github.com/Harrison-Blair/hive/internal/land	4.004s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.260s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.341s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	4.339s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.017s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.287s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.215s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	29.183s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/logs	(cached)
ok  	github.com/Harrison-Blair/hive/internal/memory	2.423s
ok  	github.com/Harrison-Blair/hive/internal/project	1.095s
ok  	github.com/Harrison-Blair/hive/internal/supervise	3.483s
ok  	github.com/Harrison-Blair/hive/internal/task	10.228s
ok  	github.com/Harrison-Blair/hive/internal/tui	67.351s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	(cached)
ok  	github.com/Harrison-Blair/hive/internal/worktree	5.415s
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
internal/tui/component/fit_test.go:123:17: undefined: ChatLines
internal/tui/component/fit_test.go:123:46: undefined: ChatMessage
internal/tui/component/fit_test.go:123:91: undefined: ChatEvent
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
--- PASS: TestCardGolden (0.00s)
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
--- PASS: TestFitFrameRows (0.14s)
=== RUN   TestFitPanels
--- PASS: TestFitPanels (0.61s)
=== RUN   TestFitWave4
--- PASS: TestFitWave4 (1.52s)
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
ok  	github.com/Harrison-Blair/hive/internal/tui/component	2.349s
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
FAIL	github.com/Harrison-Blair/hive/internal/tui/component	0.034s
FAIL
```

### Independent width probe

```sh
go run .hive/tmp/verify-23/probe.go
```

```text
Chat panel width 80, inner width 76, error <nil>
line 0 width 86
line 1 width 2
Chat panel width 119, inner width 115, error <nil>
line 0 width 86
line 1 width 2
Chat panel width 120, inner width 116, error <nil>
line 0 width 23
Markdown wrap width 40, error <nil>
line 0 width 40
line 1 width 40
line 2 width 20
Markdown wrap width 40, error <nil>
line 0 width 0
line 1 width 40
line 2 width 40
line 3 width 20
line 4 width 0
```

## Cleanup

Removed both scratch worktrees and the scratch probe, logs, and folder after recording their output here. No live home, pane, tab, or workspace was made for this check.
