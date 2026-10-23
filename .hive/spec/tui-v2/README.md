# Hive TUI v2

## Destination

Bare `hive` opens a TUI that matches the approved design boards cell for cell. Reusable components make up the TUI. Each screen has a fixed frame, and only its regions scroll. The layout fits the pane from 80×24 upward. A live capture of each screen matches its design board.

## Design authority

The canvas artifact "Hive TUI Redesign" is the design authority: https://claude.ai/artifact/VrhcA98g9G2G9cFQao5Z9h. The owner approved it on 2026-10-21. When a board and this spec disagree, stop and ask the user.

Each board is a grid of terminal cells. The script in the board writes each cell with a character, a palette token, and a weight. No browser layout or font metric sets a cell. Wave 1 copies each board source into `.hive/spec/tui-v2/design/`. Wave 2 exports each board once to a golden grid file beside it. Tests read only the golden grid files.

| Board | Screen | Size |
| --- | --- | --- |
| `Main.dc.html` | 1 Agents, one project | 120×40 |
| `AgentsMany.dc.html` | 1 Agents, five projects | 120×40 |
| `AgentDetails.dc.html` | 1 Agents, details popup | 120×40 |
| `CompactAgents.dc.html` | 1 Agents | 80×24 |
| `Chat.dc.html` | 2 Chat | 120×40 |
| `CompactChat.dc.html` | 2 Chat | 80×24 |
| `ChatStatus.dc.html` | 2 Chat, working status and growing input | 120×40 |
| `ChatEvents.dc.html` | 2 Chat, open events | 120×40 |
| `CompactChatStatus.dc.html` | 2 Chat, blocked status and capped input | 80×24 |
| `WorkGoals.dc.html` | 3 Work, Goals view | 120×40 |
| `WorkQueue.dc.html` | 3 Work, Queue view | 120×40 |
| `Work.dc.html` | 3 Work, Tasks view | 120×40 |
| `Compact.dc.html` | 3 Work, details overlay | 80×24 |
| `Friction.dc.html` | 4 Friction, Open view | 120×40 |
| `FrictionResolved.dc.html` | 4 Friction, Resolved view | 120×40 |
| `CompactFriction.dc.html` | 4 Friction, list | 80×24 |
| `Health.dc.html` | 5 Health, Home view | 120×40 |
| `HealthAgents.dc.html` | 5 Health, Agents view, restart prompt | 120×40 |
| `CompactHealth.dc.html` | 5 Health, Agents view | 80×24 |
| `Logs.dc.html` | 6 Logs, JSON panel | 120×40 |
| `CompactLogs.dc.html` | 6 Logs, details overlay | 80×24 |
| `Help.dc.html` | `?` help overlay | 120×40 |
| `TooSmall.dc.html` | too-small notice | 60×16 |
| `Components.dc.html` | component sheet | 120×67 |
| `AgentsHome.dc.html` | 1 Agents, home column | 120×40 |
| `AgentsNested.dc.html` | 1 Agents, nested child of a worker | 120×40 |
| `AgentsHiddenParent.dc.html` | 1 Agents, child moved to home by `h` | 120×40 |
| `AgentsCoordinatorChild.dc.html` | 1 Agents, child of the coordinator | 120×40 |
| `CompactAgentsHome.dc.html` | 1 Agents, home page | 80×24 |
| `AgentDetailsGeneral.dc.html` | 1 Agents, details popup of a general agent | 120×40 |
| `HealthAgentsGeneral.dc.html` | 5 Health, Agents view, general agent | 120×40 |
| `LiveOverview.dc.html` | 1 Agents, pods overview with a lit line | 120×40 |
| `LivePod.dc.html` | 1 Agents, one-pod view with a lit line | 120×40 |
| `CompactLiveOverview.dc.html` | 1 Agents, pods overview with a lit line | 80×24 |
| `CompactLivePod.dc.html` | 1 Agents, one-pod view with a lit line | 80×24 |
| `TreeWide.dc.html` | 1 Agents, one-pod tree wider than the region | 120×40 |
| `TreeNested.dc.html` | 1 Agents, one-pod tree with a nested child | 120×40 |
| `CompactTreeWide.dc.html` | 1 Agents, one-pod tree that scrolls sideways | 80×24 |
| `CompactTreeNested.dc.html` | 1 Agents, one-pod tree that scrolls vertically | 80×24 |
| `ChatNormal.dc.html` | 2 Chat, normal mode | 120×40 |
| `ChatInput.dc.html` | 2 Chat, input mode | 120×40 |
| `CompactChatNormal.dc.html` | 2 Chat, normal mode | 80×24 |

## Out of scope

- New data sources. Each screen shows data that hive already stores.
- TUI write actions other than restart and chat. Decision 14 of `hive-tui` stays in force. Restart stays the one exception.
- A light theme.
- Automatic detection of a scratch home. Decision 34 adds an opt-in switch only.

## Decisions

### Frame

1. Each screen has a fixed header, a tab bar, a content area, and a fixed footer. The header and the footer never scroll.
2. From 120 columns, the header shows the home on the left. On the right it shows the coordinator state, the supervisor state, the needs-you count, and the clock. Below 120 columns, the header uses the short form on the 80×24 boards.
3. The tab bar shows the labels `1 Agents`, `2 Chat`, `3 Work`, `4 Friction`, `5 Health`, and `6 Logs`. Each label has one space on each side. At 120 columns and wider, one more space separates the labels. Below 120 columns, the labels touch. The active tab is accent and bold, with a `━` underline.
4. The footer shows `[region]` in accent, then the key hints of the focused region in dim.
5. Only regions scroll. A region draws its scroll mark `┃` on its right border. A region that scrolls sideways draws `━` on its bottom border. The scroll mark uses `accent` in the focused region and `dim` in the other regions. On Chat, use `dim` for scroll marks in every focus region. An overlay or a popup dims the full screen behind it, including the header and the tab bar.
6. One region has focus. The focused region has an accent border. `tab` moves the focus to the next region.

### Fit to screen

7. The minimum size is 80×24. Below it, the TUI shows the too-small notice with the current size. It also shows ` hive` on the first row and `ctrl+c ×2 quit` on the last row. It shows no other content.
8. From 120 columns, a list and its details sit side by side.
9. From 80 to 119 columns, details open as a full-width overlay over the list. `enter` opens the overlay. `esc` closes it.
10. Agents gives each project one column of equal width. The minimum column width is 26 cells. Each card is 24×4 cells. The idle dot is `◐` in `accent` on each screen. When the columns do not fit, the graph scrolls sideways and keeps the selected card visible. With no project, Agents shows the notice "no project agents".
11. A resize never breaks a border. A resize with an overlay open keeps the overlay and its scroll position.
12. Each component measures text with `lipgloss.Width`. No component measures text with `len`.
13. Components use only the symbols in this list: `─ │ ╭ ╮ ╰ ╯ ├ ┤ ┼ ┴ ┬ └ ━ ┃ ┏ ┓ ┗ ┛ ┯ ┷ ┡ ┨ ● ◐ ○ ▶ ✓ • · – — … ↑ ↓ ← → › █ × ▸ ▾ ▲ ┰ ┸ ┺ ┹ ╃ ┻ ◀ ▼ ┟ ┠ ┥ ╅`. Each symbol takes one cell. Wave 2 checks that each exported board uses only these symbols. A test checks that `lipgloss.Width` gives 1 for each symbol. The TUI supports terminals that draw ambiguous-width characters as one cell. A new symbol needs a change to this list.

### Screens

14. Decision 65 replaces this decision. **Agents** shows a full-width graph of project agents only. The coordinator and the supervisor are not on the graph. `enter` opens the details popup. `p` peeks at the pane. `a` attaches. `h` hides stopped agents.
15. **Chat** shows a scrolling feed and a fixed input. User text uses the `user` color. Agent replies render as markdown with Glamour. Events fold to one dim line.
16. **Work** has the views Goals, Queue, and Tasks. Each view is a list with details. `b` opens the brief. `v` opens `VERIFIED.md`. `d` shows the dependencies. `o` opens the goal design. `s` shows stops only.
17. **Friction** has the views Open and Resolved. Each view is a list with details.
18. **Health** has the views Home and Agents. Home lists the checks of `hive doctor`. Agents lists each agent, including the coordinator and the supervisor.
19. Restart of the coordinator and the supervisor moves to Health › Agents. Select the agent and press `r`. The TUI asks the user to accept only when the agent is live. The restart behavior of task `3-tui-restart` does not change. The Agents graph and the details popup have no `r` key. On a manager or worker row in Health › Agents, `r` shows the footer notice "restart: coordinator and supervisor only" and does nothing.
20. **Logs** has a fixed filter bar, the event list, and a JSON details panel that scrolls on its own.

### Keys

21. Each key has one meaning. The Help board is the key map. This table gives the route of each key for each focus state.

| Focus state | Keys that the focus takes | Keys that stay global | Exit |
| --- | --- | --- | --- |
| List, graph, or details | none | all keys | — |
| Chat input | each printable key, `enter` send, `shift+enter` newline, `backspace`, paste | `tab`, `pgup`, `pgdn`, `ctrl+c` | `esc` gives the focus to the feed |
| Filter bar | each printable key, `enter` apply, `backspace`, paste | `ctrl+c` | `esc` closes the bar and keeps the applied filter |
| Overlay or popup | `↑` `↓` `pgup` `pgdn` scroll, the action keys of the overlay | `ctrl+c` | `esc` closes it |
| Help | `↑` `↓` scroll | `ctrl+c` | `?` or `esc` closes it |
| Restart prompt | `y` restart | `ctrl+c` | each other key cancels, as task `3-tui-restart` does |

The key hints in the footer show only the keys of the focus state. The Chat input, an overlay, and a popup show no `?` hint. When the restart prompt is open, the footer shows `[restart] y restart · any other key cancels`, and the list border uses `border`.

22. Global keys: `1`–`6` screen, `tab` region, `[` `]` view, `↑↓` or `j` `k` move, `←→` graph, `pgup` `pgdn` page, `g` `G` top and end, `enter` open, `esc` back, `/` filter, `?` help, `R` refresh, `ctrl+c` twice to quit.
23. Action keys: `a` attach, `p` peek, `r` restart, `h` hide stopped, `t` task, `b` brief, `v` verified, `d` dependencies, `o` design, `s` stops only, `f` follow, `<` `>` day.
24. On the Chat screen, the feed can have the focus. Then a printable key gives the focus to the input and types the key. The keys `1`–`6`, `?`, `/`, `j`, `k`, `g`, `G`, `R`, and `f` stay global. `i` and `enter` also give the focus to the input.
25. The filter bar takes typed fields: `level:`, `comp:`, `proj:`, `agent:`, `task:`, `req:`, and free text.

### Look

26. The palette tokens are these. `fg` and the background are the terminal defaults.

| Token | Hex | Use |
| --- | --- | --- |
| `border` | #3d3a33 | unfocused borders, rules |
| `dim` | #a39d90 | labels, hints, events |
| `accent` | #f5b942 | focus, selection, titles |
| `working` | #6ee69a | working and live dots |
| `done` | #7ccf6a | done, landed, pass |
| `info` | #6fb7ff | agent replies, components |
| `user` | #e9cf9b | user chat text (new) |
| `code` | #9fd0ff | inline code and code blocks (new) |

27. Panels use `lipgloss.RoundedBorder()` with `Padding(0, 1)`. The panel title sits in the top border.

### Markdown

28. Chat renders agent replies with `charm.land/glamour/v2` at v2.0.1. The owner chose Glamour on 2026-10-21. The commit that adds it states this reason: "markdown rendering of agent chat replies, chosen by the owner 2026-10-21".
29. Glamour uses a custom style built from the palette tokens. Headings use `accent` and bold. Strong text uses `fg` and bold. Inline code and code blocks use `code`, with no background and no syntax colors. A code block has no border and no title. List items use `• `. Links show their text in `info` and no URL. The document margin is 0. Wave 4 records the complete style in one Go file.
30. At 120 columns and wider, the feed has a gutter for the time and the name. Message text starts 23 cells from the left edge of the feed panel. The wrap width is the panel width minus 26 cells. Below 120 columns, the feed has no gutter. The time and the name use their own line, and the text starts at the panel padding.
31. The Chat component removes only the padding that Glamour adds: the blank first line and the trailing spaces of each line. It keeps leading spaces, so code indentation stays.
32. Glamour v2.0.1 draws each table, and hive does not change the result. The user decided this on 2026-10-21. Glamour gives each column the width of its widest cell, with one space on each side. One rule cell separates two columns. The table has no outer border. When the table is narrower than the wrap width, the narrowest column gets one more cell. This step repeats until the table is as wide as the wrap width. When two columns have the same width, the left column gets the cell first. When the table is wider than the wrap width, the widest column loses cells first. Then the text of its cells wraps to more lines, with no `…`. When a column has too few cells, its header shows only `…` or nothing, and its values wrap one character to a line. Glamour has no `Header: value` form. The header row uses the style of the other rows: `fg`, not bold. The rules `│`, `┼`, and `─` use `border`, from the style of decision 29. Tests cover uneven content, columns of equal width, and a table with too many columns.
33. The coordinator tested Glamour v2.0.1 on 2026-10-21 in a scratch module:
    - It resolves with `charm.land/lipgloss/v2` v2.0.6 and adds no second Lip Gloss.
    - It adds goldmark, chroma, bluemonday, and their dependencies.
    - It renders the sample reply in about 2 ms.
    - It draws tables with inner rules only. The Chat boards show the tables that Glamour draws, as decision 32 gives.

### Scratch startup and demo mode

34. Bare `hive --no-startup` opens the TUI and starts no coordinator, no supervisor, and no tab. It does not rename the tab. The environment variable `HIVE_NO_STARTUP=1` has the same effect. The worker skill and the verifier skill tell each agent to use the switch when it runs bare `hive`. This change addresses the friction entry `scratch-hive-starts-real-agents`.
35. Bare `hive --demo FILE` opens the TUI in demo mode. Demo mode implies `--no-startup`. It reads all screen data from the fixture file. It uses the clock time in the fixture file and does not advance it. It runs no `hive doctor` and no other process. It reads no live record and writes no record. It writes log events as usual, with the field `demo` set to true, because AGENTS.md requires them. The Logs screen in demo mode shows only the log events of the fixture. Help does not list the flag. Demo mode uses the same renderers as normal mode, but not the same effects. In demo mode, chat send, restart, attach, and peek show the notice "demo mode: no action" and do nothing. A test of the CLI entry point presses these keys. It checks that demo mode writes no file other than the log file. It also checks that demo mode starts no process.
36. Bare `hive --demo FILE --size 120x40` draws the TUI at exactly that size, from the top-left cell of the pane. The pane must be equal to or larger than the size. When the pane is smaller, the TUI shows the too-small notice.

### Visual accuracy

37. Visual accuracy is a requirement. Each screen task passes the golden check and the live check before it can land. Demo mode and normal mode use the same screen renderers. Only the data loader differs.
38. **Normal-mode checks.** Each screen task also keeps or adds tests of normal mode. These tests run in a scratch home with `--no-startup`. They cover the record loaders, refresh, chat delivery, and the restart move. These tests use no live coordinator and no live supervisor.
39. **Golden grid.** A golden grid holds one record for each cell: the character, the foreground token, and the bold attribute. The background is always the terminal default. A cell with no character is a space with the token `fg`.
40. **Converter.** One converter turns ANSI text into a grid. It maps each 24-bit SGR foreground to the palette token with the same hex value. It maps the default foreground to `fg`. It fails on an indexed color, a 16-color code, a background color, or a hex value outside the palette. Tests of the converter use hand-written ANSI input and hand-written expected grids.
41. **Exporter.** The exporter runs the script of each board and writes its grid. Wave 2 runs the exporter once and commits the golden grids. A person reviews the golden grids at the end of wave 2, before wave 3 starts. The gate never runs the exporter.
42. **Golden check, in the gate.** A Go test builds the model from the fixture file of the board with the demo loader. It renders `View()` at the board size with `lipgloss` set to the TrueColor profile. It converts the output with the converter. It compares the result with the golden grid cell by cell. The test fails on each cell that differs outside a mask, and prints the differing cells.
43. **Live check, in verification.** The verifier opens a herdr pane that is equal to or larger than the board size. It runs `COLORTERM=truecolor hive --demo FILE --size WxH` in the pane. It waits until two reads 500 ms apart give the same text. It reads the pane with `pane.read`, `source: visible`, `format: ansi`. It rejects a read with `truncated: true` or a scroll offset other than 0. It crops the top-left W×H cells, converts them, and compares them with the golden grid. It writes the capture and the diff to the task evidence.
44. **Fixture.** Each board has one fixture file in `.hive/spec/tui-v2/design/`. The fixture holds the data that the board shows: agents, goals, tasks, friction entries, checks, chat messages, and log events. It also holds the clock with its time zone, the home path to show, and the terminal size. The fixture also holds the UI state of the board. The UI state is the screen, the view, the focus, the selection, and each scroll position. It also includes the filter and the open overlay, popup, or prompt. The golden check and the live check read the same fixture file.
45. **Masks.** A mask file lists rectangles of cells by row, first column, and width. A mask covers only a field that the fixture cannot fix. The demo clock and fixed fixture data make most masks unnecessary. A mask field has a fixed width. The check still compares the bold attribute and the token in a masked cell. A mask that touches a border, a title, a label, or a key hint fails the check. A change to a mask file needs the approval of the user.
46. **Responsive checks.** The golden check also renders each screen at 79×24, 80×24, 119×40, and 120×40. At 79 columns it expects the too-small notice. At 80 and 119 columns, Work, Friction, Health, and Logs show the list, and show the overlay after `enter`. Agents and Chat have no overlay form. They show the compact boards at 80×24. The tests cover each case at the minimum height and at one row more. Tests also cover empty data, text longer than its column, nine projects on Agents, and a resize with an overlay open. These tests check properties: no line wider than the pane, no broken border, and the selected item visible.
47. **Side-by-side review.** The coordinator renders each live capture with the board renderer and publishes it beside its board in the design artifact. The user signs off each screen. The goal finishes only after the user signs off each screen.

### Repository rules

48. Wave 1 adds `.hive/spec/tui-v2/README.md` to `steFiles` with a baseline of 0. Each brief is in ASD-STE100 Strict.
49. Each task is test-first. Each brief names the failing test to write first and requires the full gate in its evidence.
50. The exporter and the converter are Go code under `internal/`. A command for them, if one is necessary, follows the `cmd/` layout. The exporter can run the board script with Node.js, because the boards are JavaScript. The exporter is a development tool and is not part of the `hive` binary. The commit that adds it states this reason.

### Pull request

51. After the user signs off each screen, the manager updates pull request #47 (`dev` into `main`, https://github.com/Harrison-Blair/hive/pull/47).
52. The manager renders each signed-off live capture to a PNG file with Charm `freeze` v0.2.2. `freeze` is a development tool that `go install` installs. It does not go into `go.mod`. The manager records the exact `freeze` command in the task evidence. The command sets the font JetBrains Mono, the theme, the padding, and the window size of the board. Each PNG has the cell size of its board, with no wrap. The PNG files go in `.hive/spec/tui-v2/screenshots/`, with the board name as the file name. The manager commits them on `dev` and pushes `dev`.
53. The manager runs the `gh-summarize-pr` skill with the argument `approve` on pull request #47. The skill applies the title and the body with no draft review. The body keeps the sections of the skill. It adds a section `## Screenshots` after `## Changes`. The section shows each PNG file with its screen name. Each image link uses the form `https://github.com/Harrison-Blair/hive/raw/SHA/.hive/spec/tui-v2/screenshots/NAME.png`, where SHA is the full commit that adds the screenshots.
54. The manager does not merge pull request #47. A merge to `main` needs the owner.

## Manager decisions

- The package layout of the components, the converter, and the exporter.
- The format of the golden grid files, the fixture files, and the mask files.

The manager made these decisions in task `9-cellgrid`:

- Package: `internal/lib/cellgrid/`. The golden check and the exporter call it.
- Grid file: `<Board>.grid` beside each board in `.hive/spec/tui-v2/design/`. It is UTF-8 text. Line 1 is `hive-grid 1 W H`. Then each row has three lines:
  - `t|` and the W characters of the row.
  - `k|` and W token letters.
  - `b|` and W digits `0` or `1` for bold.
- Token letters: `f` fg, `r` border, `d` dim, `a` accent, `w` working, `o` done, `i` info, `u` user, `c` code.
- Mask file: `<Board>.mask` beside the grid. Each line is `ROW COL WIDTH`, zero-based, and an optional `# reason`. A blank line or a `#` line is a comment. A board with no mask has no mask file.
- Fixture file: `<Board>.fixture.json` beside the grid. Task `demo-mode` defines its schema.

The worker made these decisions in task `11-demo-mode`:

- Demo loader: the file `internal/tui/demo.go` in the package `tui`. The loaders give the data types of this package, so the demo loader stays in it.
- Fixture schema: one JSON object. An unknown field fails the load, and the error names the field. The object has these fields:
  - `board`: the board file name.
  - `terminal`: the `width` and the `height` of the board.
  - `clock`: the `time` in RFC 3339 and the `zone` as an IANA time zone. Demo mode does not advance the clock.
  - `home`: the home path that the header shows.
  - `data`: the records in the format that hive stores. The fields are `projects`, `agents`, `goals`, `tasks`, `inbox` by recipient, `evidence` by task, `friction`, `checks`, `chat`, `logs`, and `files`. The field `supervisor` replaces the supervisor claim and its process check. The field `files` holds the text of each file that a screen can open. These are a task brief, an evidence file, a feature README, and a friction entry. Its key is the path relative to the home or to a project directory. Demo mode reads no live file. When `files` has no text for a file, the screen shows `demo mode: no file`.
  - `ui`: the fields `screen`, `view`, `focus`, `selection`, `scroll` by region, `filter`, `overlay`, `popup`, `prompt`, `hide_stopped`, `stops_only`, `follow`, `log_day`, `chat_noise`, and `chat_input`.
- The evidence of task `11-demo-mode` gives each field and each board that uses it.

The worker made these decisions in task `29-frame-layout`:

- Registry: `screenRegistry` in `internal/tui/frame.go` has one `frameScreen` for each tab. A `frameScreen` has the `label`, the `views`, the `regions`, the flag `listDetails`, and the key hints for each focus state in `hints` and `compact`. A hint key is `view/focus` or `focus`. The `compact` hints apply below 120 columns.
- Content: `contentView` gives the content area for the size and the focus. For a tab with `listDetails`, the list and the details sit side by side from 120 columns. The list has 13/30 of the width. Below 120 columns, `enter` gives the focus to the details, and the details fill the content area. Wave 6 puts a screen in the registry and replaces `legacyPanel` and `legacyDetails` for that tab.
- Current screens until wave 6:

| Tab | View | Current screen |
| --- | --- | --- |
| 1 Agents | — | Agents. `[` and `]` change its TREE, GRAPH, and TABLE views. |
| 2 Chat | — | Chat |
| 3 Work | Goals, Queue, Tasks | Goals, Queue, Tasks |
| 4 Friction | Open, Resolved | Friction, with its open and resolved tabs |
| 5 Health | Home, Agents | Health, Agents. The restart key `r` works only in Health › Agents. |
| 6 Logs | — | Logs. `<` and `>` change the day. |

- A key of a current screen that the frame does not route goes to that screen. The frame takes `tab`, so the status tabs of the current Tasks screen have no key until wave 6.
- Below 120 columns, the details region of Work, Friction, Health, and Logs is an overlay. The overlay takes `esc`, `↑` `↓` `j` `k` `pgup` `pgdn`, and the action keys of decision 23. Each other key does nothing, as decision 21 gives. From 120 columns, the details region is a region of the screen, and each key stays global.

The worker made these decisions in task `33-screen-agents`. They obey the user decision of 2026-10-21. That decision gives a tree when the children fit side by side, and a rail in the other cases. From 120 columns, one column has a maximum width of 64 cells.

- Screen: the file `internal/tui/agents_screen.go`. Task `39-screen-health` gave Health › Agents its own panel. Thus no code draws the old Agents screen. Task 33 removed its views and their tests. `agents.go` keeps only the shared parts: the record loader, the peek, and the restart.
- Outer title: `project agents · N projects`. For one project, the title is `project agents · 1 project`.
- Columns: a pod has a minimum width of 28 cells. One cell separates two pods. The number of pods on the screen is `min(N, (B+1)/29)`, where B is the body width of the panel. The width of each pod is `(B−(V−1))/V`, where V is the number of pods on the screen. From 120 columns, a pod has a maximum width of 64 cells. The pods are at the center of the body. They start on body row 1. Each pod on the screen has the height of the highest pod.
- Pod: a `border` box. The title ` PROJECT ` is dim and bold. Row 1 is the goal line in dim. For an active goal, the goal line is `GOAL · wave W` for the first wave that is not done. When that wave has `needs_user`, the line is `needs you · TEXT` in accent. When no goal is active, the line is `GOAL · STATUS` for the last goal. When the project has no goal, the row is empty.
- Tree: the pod uses the tree when `content / children` is 26 cells or more. The manager card is at the center. Its bottom border has a `┬`. Below it are a `│` row, the bus `╭─┴─╮` from the first child to the last child, and one more `│` row. Then the child cards come, each with a `┴` on its top border. Each child is at the center of a slot of `content / children` cells. A joint on a card border is accent when the user selects the card. In the other cases, it is `border`.
- Rail: the manager card is on content column 0 with `├` on its bottom-left corner. Each child uses 5 rows: one `│` row, then the card on content column 1. The second row of the card has `├┤`, or `╰┤` for the last child.
- A pod with no child shows `no workers` in dim on the second row below the manager card.
- Card sub line: `HARNESS · STATE · TAIL` at each width. For a live agent, TAIL is the age of `status_changed_at`. When the age is not known, the line has no TAIL. For a stopped agent, TAIL is the verification result of the agent, or else its `agent_status`.
- Stopped text: `h hides N stopped` in dim, on the second row below the last child. When `h` hides the stopped agents, the text is `h shows N stopped`.
- Legend: `── spawn   ● working   ◐ idle   ○ stopped   one column per project` in dim, on the last body row. The legend shows only from 120 columns. When the graph scrolls sideways, the legend moves up one row.
- Scroll line: when the pods do not fit, the last body row shows `←→ at the edge scrolls · projects A–B of N` on the left, and `← PREVIOUS` and `NEXT →` on the right. The bottom border shows the `━` mark of `component.Thumb`. When the pods are higher than the region, the graph scrolls vertically and keeps the selected card on the screen. The right border then shows the `┃` mark.
- Keys: `←` `→` move between the children of a tree, and at the edge of a pod to the next pod. The index in the pod stays the same, and it is cut to the pod. `↑` `↓` `j` `k` move between the manager and the children. The details popup is 60×27 cells. The popup shows each field that has a value, then `ACTIONS`, then `RECENT` with the last four log events of the agent. One blank line comes before each section.
- Selected card: the user decision of 2026-10-21, option A. The selected agent card has a heavy border `┏━┓┃┗┛` in accent. This border is the plain-text signal of selection. The layout and the sizes do not change. The `Heavy` field of `component.Panel` draws this border. `component.Card` sets it for the selected card. A light connector meets the heavy border at a joint of two weights. The joints are `┯` below a tree manager and `┷` above a tree child. On the rail, the joints are `┡` on the corner of the manager and `┨` beside a child.

The manager made this decision in task `29-frame-layout`:

- A live check compares a capture of a pane with a board grid. In a space cell of a live capture, the check does not compare the token or the bold flag. A terminal does not show them. It compares each other cell fully. The golden check at the model compares each cell fully.

The worker made these decisions in task `37-screen-work-tasks`:

- Hook: the Tasks view uses the `panels` hook of task `36-screen-work-queue`. The field `back` of `screenPanels` closes a part of the details on `esc`, before `esc` gives the focus to the list.
- Tasks view: the file `internal/tui/work_tasks.go`. It replaces `task_screen.go`. The list shows all tasks with the columns ID, STATUS, and WAVE. It has no status tabs and no project filter, because decision 23 gives `p` to peek.
- Details from 120 columns: a blank line, then the fields title, status, goal, worker, verifier, branch, merge, and depends. A value that is not in the records shows `—`. Below 120 columns, the details show the fields status, goal, verifier, and merge, as the Compact board shows.
- ACCEPTANCE lists the AC rows of the table in `brief.md`. The mark is `✓` in `done` when the last verification is a pass, and `○` in `dim` for each other result. EVIDENCE shows the key line and then the V rows of the table in `VERIFIED.md`. The command column of a V row is 38 cells.
- `b`, `v`, and `d` show the brief, `VERIFIED.md`, or the dependencies in the details region. Below 120 columns they open the overlay. A new selection closes the file. The task loader reads `brief.md` and `VERIFIED.md`. Demo mode reads them from `files` in the fixture.
- Details: the view uses `component.Details`, with the overflow rule of task `42-wave6-board-redraw`.
- While the details overlay is open below 120 columns, the view row shows the hint `list › details`. It ends 2 cells before the right edge.
- Component changes: `List.Fill` gives the padding of each cell the style of the cell. `SubTabs.Hint` replaces the view hint.
- Refresh of the Tasks view watches the agent records and the task folder of each task. The details show the worker status and the files of the task.

The worker made these decisions in task `34-screen-chat`:

- Screen: `internal/tui/chat.go` replaces `chat_screen.go` and its tests. The registry entry of the Chat tab does not change. `scrollChatFeed` moves from `frame.go` to `chat.go`.
- Feed: `component.ChatLines` gives the lines of each message. A blank line separates two messages. From 120 columns, the first body row of the feed is blank, and the title is `chat · noise LEVEL`. The top border also shows ` ↑ N earlier ` in `dim` from the middle column. N is the count of messages above the view. Below 120 columns, the title is `chat`.
- Scroll: `ui.scroll.feed.top` gives the index of the message at the top of the feed. The feed then does not follow its end. In normal mode, the feed follows its end. `pgup` and `pgdn` page the feed. In the feed, `↑` `↓` `j` `k` move it by one line, `g` goes to the top, and `G` goes to the end and follows it. `component.Thumb` gives the scroll mark. Its total is the count of feed lines.
- Noise: the levels are `quiet`, `normal`, and `all`, from `ui.chat_noise`. The default is `normal`. They show the inbox kinds of the earlier levels "decisions only", "landings and decisions", and "all". No key changes the level, because decision 24 gives each printable key in the feed to the input. The screen has no keys `n` and `e`, which the earlier screen had. Each event folds to one line, as decision 15 gives.
- Name: when a user message is not sent, its delivery shows after the name, for example `you · pending`. The time uses the form `15:04` in the zone of the clock.
- Input: `component.Input` draws the input and its cursor cell `█`. While the input has the focus, the terminal cursor is on that cell. The input shows no placeholder text, as the boards give.
- Removed: the coordinator line and the column of hive events of the earlier screen. The header shows the coordinator state.

The worker made these decisions in task `38-screen-friction`:

- Screen: the file `internal/tui/friction.go`. It replaces `friction_screen.go` and keeps its loader.
- Registry: the Friction entry has a `panels` entry for the views Open and Resolved. Task `36-screen-work-queue` added the field `panels`.
- List: the column ENTRY shows the file name without `.md`. From 120 columns, the date column shows `MM-DD`. Below 120 columns, the list shows the full date, the column SEEN, and the hint `enter opens details`.
- Order: the Open view shows the newest opened date first. The Resolved view shows the newest date of the status `Resolved YYYY-MM-DD` first. Entries with the same date show in the order of their names.
- Seen: the count of lines in the section Occurrences, and the last date of these lines. An entry with no Occurrences section shows 1 and its opened date.
- Details: the fields title, status, opened, seen or decision, and file. The Open view shows seen. The Resolved view shows decision. Then each section shows with its title in capital letters. Occurrences lines are dim. In Resolved by, the first field has 16 cells, and the remaining text is dim.
- Keys: from 120 columns, `enter` opens the entry file in the details region, and `esc` closes it. The file scrolls with the offset of the details region. After `esc`, the details of the entry show from the top. In the details region, `g` and `G` go to the top and the end of the details.
- Component fix: `Details` shows `↓ N more lines` in the last body row when lines stay below the view. The scroll mark uses the lines that show. This follows the user decision of 2026-10-21.
- Component fix: `List` has the field `Hint`. The hint shows in dim at the right of the last body row.
- Component fix: `List` draws the selected first cell in accent and bold over its full column. The space after `▶` is fg. This follows the user decisions 4 and 7 of 2026-10-21.

The manager made this decision in task `40-screen-logs`:

- A row of the Logs event list shows the `msg` of the event. When the event has the field `reason`, the row also shows ` · ` and the text of the reason before the first `;`. The JSON panel shows the full reason.

The worker made these decisions in task `40-screen-logs`:

- Screen: the file `internal/tui/logs_screen.go`. The frame gives the content area of tab 6 to `logsContent`. From 120 columns, the event list has 8/15 of the width.
- Filter: the filter bar text is the applied filter. The default text is `level:info`. `level:` gives the lowest level. `agent:` matches the agent ID, or the agent name before its generation. Each other typed field must be equal to the field of the event. Each word of free text must be in the row text.
- Selection: when a load gives a new newest event, follow selects it. When the newest event does not change, the selection stays. When no row shows, the selected event stays, and it shows again with the rows. A day change selects the first row.
- Keys: `/` opens the filter bar with the applied text. `esc` in the event list sets the default filter. `f` changes follow. `↑`, `k`, `pgup`, and `g` set follow to off. `G` sets follow to on.
- This task removes the old keys `l`, `c`, `p`, `a`, and `r` of the Logs screen. Decisions 21 to 25 give these keys other meanings or give the typed fields.
- The JSON panel cuts each line to the panel width less 6 cells, with `…`. When the lines do not fit, the last body row shows `↓ N more lines`.
- Component change: `FilterBar` has the fields `Focused` and `Status`. `Status` is at the end of the bar. The Logs screen shows the follow mark and the day there.

The worker made these decisions in task `36-screen-work-queue`:

- Registry: a `frameScreen` has the field `panels`. Its key is a view name. A `screenPanels` gives the list region and the details of that view. The frame uses `legacyPanel` and `legacyDetails` only for a view with no `panels` entry.
- Queue list: the list shows each task of each project, also a landed task. The title gives the project when there is one project, the task count, and the stop count.
- Queue groups: the tasks of a project are in groups by grant. The grant of a task is the grant of its goal, then the grant of its delivery. The group `NO GRANT` is last. With more than one project, each group heading starts with the project. A project with no tasks shows `PROJECT · no tasks`.
- Queue tree: in a group, a task is below the first task of the group that it depends on. A task in a dependency cycle starts a tree.
- Queue stop: an open task stops when it has a needs-user text, or when its acting agent has the status `unresolved`. A stop row `! stop: REASON` follows the task row. `s` shows only the stopped tasks, each at depth 0.
- Queue keys: `t` opens the selected task in Work › Tasks. From 120 columns, `enter` gives the focus to the details.
- The board replaces the Queue columns `#`, `starts when`, and the states `◆ auto` and `┄ waits`. The header counts of the Queue stay.

The worker made these decisions in task `39-screen-health`:

- Registry: Health uses the field `panels` of task `36-screen-work-queue`. The views Home and Agents have entries with the same functions.
- Health: the file `internal/tui/health.go` holds the screen. Each view has its own selection. One load gives the checks of `hive doctor` and the agent records, so Home and Agents show the same data. The details show the time of the last load.
- Restart: `r` in Health › Agents calls `requestRestart`, with the behavior of task `3-tui-restart`. The Agents screen has no `r` key. The restart prompt box is in the details of the selected agent, as the HealthAgents board shows.
- Rows: the Health lists draw their rows in a `Panel`. The columns and the colours of the space cells come from the Health boards. The selected row follows the rule of `List`.
- Component fixes: none.

### Chat status and events

55. The input border shows the coordinator status after ` message `. Show `· ` in dim, then the status glyph, the status word, and the whole-minute age. Use `●` in `working` for working. Use `○` in dim for idle. Use `▲` in accent for blocked. For another `agent_status` value, use `○` in dim and show the value as the word. When no live coordinator record exists, show `○ none`. Omit the age when `status_changed_at` has no value. Read `agent_status` and `status_changed_at` from the live coordinator record. Calculate age from `status_changed_at` to the TUI clock. Round down to whole minutes. When the latest user chat entry has delivery `pending`, add `· queued` in dim after the status. Below 120 columns, the header shows `▲` in accent for a blocked coordinator. At 120 columns and wider, the header shows the glyph and status word.

56. Wrap input text at spaces. Break a word that is wider than the text width at the line edge. Grow the input panel by one row for each text row. Limit the panel, including its borders, to one third of the content rows. Round down. Content rows are between the tab rule and the footer. The limit is 12 rows at 120×40 and 6 rows at 80×24. When the text needs more rows than the limit, scroll inside the panel. Keep the cursor row visible. Draw the scroll mark on the right border. Keep the input height when focus moves to the feed. Shrink the feed by the same number of rows. The empty input has one text row.

57. A folded event has one dim line. Use `▸` instead of `·` before its name. `ctrl+o` opens or folds all events from the feed or the input. An open event starts with `▾`. Show its full text in dim. Wrap the text to the feed width. Start each source line on a new row. Start continuation text in the message text column. Fold all events when the TUI starts. Do not keep the toggle state. Do not add a noise key. Show `ctrl+o` in the key hints of the feed and input. The Help board shows `ctrl+o opens or folds all chat events`.

### Agents default and working cards

58. At each TUI start, the Agents screen hides each stopped agent. A stopped agent has the status `stopped` or `exited`. `h` shows the stopped agents. `h` hides them again. While `h` hides them, the stopped text of the pod is `h shows N stopped`. While `h` shows them, the stopped text is `h hides N stopped`. An unresolved agent stays on the graph. Demo mode reads `hide_stopped` from the fixture. A fixture without this field shows the stopped agents.

59. Health › Agents does not change. It shows the stopped agents at start. Its `h` key hides them and shows them again.

60. When the agent of a card has the state `working`, the card draws its rounded border in the token `working`. A selected card keeps the heavy `accent` border, also when its agent has the state `working`. A card of each other state keeps the token `border`. Each joint on the border of a working card, for example `├`, `┤`, `┴`, `┬`, and `╰`, has the token `working`. A rail or a tree line outside a card keeps the token `border`.

61. The key hint of `h` in the graph footer is `h stopped`. The Help line for `h` is `show or hide stopped agents`.

### Live agent activity

The boards `LiveOverview`, `LivePod`, `CompactLiveOverview`, and `CompactLivePod` show these decisions.

62. The Agents screen has two levels of selection. The overview shows all pods with the tree, the rail, and the pod width rules of decision 10. In the overview, you select a full pod. The home pod is a pod as a project is. A selected pod has a heavy `accent` border (`┏ ━ ┓ ┃ ┗ ┛`) and an `accent` bold title. The title of a pod that is not selected stays `dim` bold. In the overview, a card is not selectable and has no heavy border. `←` and `→` select the pod on the left and on the right. At the edge of a page, `←` and `→` scroll the pages. `enter` on a pod opens the one-pod view of that pod. In the one-pod view, you select one agent with the keys of decision 14. The arrows stay in the pod. `esc` goes back to the overview with the same pod selected.

63. The one-pod view shows one pod at the center of the body. The pod width rule of decision 10 for one pod applies. The pod is at most 64 cells wide from 120 columns. Below 120 columns, it has the full body width. The pod uses the tree or the rail by the rule of decision 10. The box title is `project agents · POD`, for example `project agents · hive`. Decision 74 replaces the rail in the one-pod view.

64. With one project and no home pod, the Agents screen opens the one-pod view at once. `esc` does nothing in this case. The footer does not show `esc pods`. It shows the graph hints of decision 14.

65. This decision replaces decision 14. The coordinator card is at the top center of the body, in the overview and in the one-pod view. It is a card as the other cards are, with the colour rules of decision 60. From 120 columns, one blank row is above the coordinator card. Below 120 columns, no blank row is above it, and no blank row is above the footer lines of the body. The coordinator joint is at the card x+12. In the overview, one bus row is below the coordinator card. A line goes from the coordinator joint to the bus row. From the bus row, a line goes to the top border of each pod on the page. Each pod drop is at the center of its pod, with `┴` on the pod top border. When the pod title is wider than the left half of the pod, the drop is 1 cell after the title. A drop 1 column or less from the coordinator joint moves to the column of the joint. Then the two lines merge (`┼`, `╃`). The bus goes only to the pods on the visible page. In the one-pod view, the line goes down from the coordinator joint through the pod top border (`┼`). It stops at the top joint of the manager card. When the manager card is not below the joint, the line turns on the bus row to the center of the manager card. When the pod has no manager, the line stops on the pod top border. In the one-pod view, the goal line ends before the line. The line has the token `border` when it is not lit. In the overview, `↑` selects the coordinator card, and `↓` selects the last selected pod again. `esc` does nothing there. `enter` on the coordinator card opens its details popup. In the one-pod view, the coordinator card is not selectable. The box title of the overview stays `project agents · N projects`. The supervisor has no card. `enter` opens the details popup of a card. `p` peeks at the pane. `a` attaches. `h` hides stopped agents.

66. A message lights the line from its sender to its receiver. Each inbox record is a message: the field `agent` is the sender, the field `role` is the receiver, and the field `created_at` is the time. Each `agent prompt` log event of `hive agent send` is a message: the field `from` of decision 72 is the sender, and the field `agent_id` is the receiver. A message with a sender or a receiver that has no card on the graph lights nothing. The user, the supervisor, and an unknown name have no card. A lit line uses heavy glyphs in the token `info`, for example `━ ┃ ┏ ┓ ┰ ┸ ┺ ┹ ╃ ┻`. On a joint of a card border, `info` wins over the colour of decision 60. A joint with arms that are not lit uses the mixed glyph with heavy arms only on the lit side. On a rail, a lit joint uses the mixed glyph with the lit arms heavy, for example `┟ ┠ ┡ ┥`. The line has one arrow in `info` (`◀ ▶ ▲ ▼`) that points to the receiver. The arrow is on the straight cell nearest to the receiver. On a horizontal run, one `━` stays between the corner and the arrow. When two cards have no direct line, the full path through their nearest common parent lights. In the overview, a line from or to a card in a pod lights up to the pod border.

67. A line stays lit for 5 seconds after the time of the message. A new message on the same line starts the 5 seconds again. When two messages on one line are in the 5 seconds, the newest message gives the arrow. Lines light in the overview and in the one-pod view. In normal mode, the TUI draws the screen again when a lit line goes dark. A change to the log file of the day loads the Agents data again.

68. The fixture of each Live board gives the message that lights its line. Demo mode uses the fixture clock, so a lit line in demo mode stays lit.

69. The legend from 120 columns is `── spawn   ● working   ◐ idle   ○ stopped   ━ ▶ message, lit 5s   one column per project`. `━ ▶` has the token `info`. The other text is `dim`.

70. The footer of the overview is `←→ select pod · ↑ coordinator · enter open pod · h stopped · ? help` from 120 columns and `←→ select pod · ↑ coordinator · enter open · ?` below 120 columns. The footer of the one-pod view is `←↑↓→ move · enter details · a attach · p peek · t task · esc pods · h stopped · ? help` from 120 columns and `←↑↓→ move · enter details · esc pods · ?` below 120 columns.

71. The Help overlay does not change. Its keys agree with decisions 62 to 70.

72. The `agent prompt` log event of `hive agent send` has the field `from` with the name of the acting agent. The acting agent is the agent record with a `pane_id` equal to the `HERDR_PANE_ID` of the caller. When no record agrees, the event has no `from` field, and the message lights nothing.

73. The user approved new boards for `Main`, `AgentsMany`, `AgentDetails`, `CompactAgents`, `AgentsHome`, `AgentsNested`, `AgentsHiddenParent`, `AgentsCoordinatorChild`, `CompactAgentsHome`, `AgentDetailsGeneral`, and `Help` on 2026-10-22. Their new grids replace the old grids. The approved boards show two cases of decision 65. The first case is the drop 1 cell after a long pod title. The second case is the line that turns on the bus row in a rail pod of the one-pod view.

### One-pod tree

The boards `TreeWide`, `TreeNested`, `CompactTreeWide`, and `CompactTreeNested` show these decisions.

74. This decision replaces the rail of decision 63 in the one-pod view. The one-pod view always draws the pod as a tree. The overview keeps the tree and the rail of decision 10. A nested child is one more tree level below its parent. The parent card has the joint `┬` on its bottom border. One `│` row is below the joint. The child card has the joint `┴` on its top border, or `┷` on a selected card. When a parent has more than one child, its children use a bus, as at level 1. The bus has one `│` row, the bus row `╭─┴─╮`, and one more `│` row. The children of the manager always use the bus. Each card is at the center of its span. A span has one slot for each card at the bottom of its subtree. A slot is the content width divided by the number of slots of the tree, with a minimum of 26 cells. When the tree is wider than the content width of decision 63, the pod is as wide as the tree. The coordinator card and its line stay above the manager card (decision 65). They move with the tree when the view scrolls sideways. In a pod with no manager, the top cards are on one row. In a project pod with no manager, no line goes to them. In home, the coordinator line goes down through the pod top border, with the joint `┼`, to each top card. This is rule A. It is the same as the line to the manager card of a project pod. With more than one top card, the line uses a bus below the pod top border, as at a tree level.

75. When the tree is wider than the region, the view scrolls sideways and keeps the selected card on the screen. The region includes the padding cell on each side of the body, so a cut card border or pod border meets the region border. From 120 columns, the view moves only as far as necessary. Below 120 columns, the selected card is at the center of the region when possible. The top cards are the children of the manager, or the cards with no parent. The last body row shows `←→ at the edge scrolls · children A–B of N` in `dim` on the left. A to B are the top cards that are fully on the screen. N is the number of top cards. A card that is only partly on the screen counts as hidden. On the right, the row shows the name of the next hidden top card on the right and `→`. When no top card is hidden on the right, the right text is `←` and the name of the nearest hidden top card on the left. From 120 columns, the legend moves up one row. The bottom border shows the `━` thumb of `component.Thumb`. The track is the inner cells of the bottom border. The total is the pod width and one cell.

76. When the tree is higher than the region, the view scrolls vertically and keeps the selected card on the screen. The view moves only as far as necessary. The right border shows the `┃` thumb.

77. The view scrolls only to keep the selection on the screen. There is no free scroll of the view. `pgup`, `pgdn`, and the wheel do not move it. The user chose this. In the tree, `↑` selects the parent card, `↓` selects the first child, and `←` and `→` select the card before or after on the same level.

78. The boards `AgentsNested` and `AgentsCoordinatorChild` show the one-pod view with the rail. Decision 74 draws them as trees. The boards `TreeWide`, `TreeNested`, and `CompactTreeWide` break 3 rules that exist. The evidence of task `125-pod-tree-scroll` lists the 3 cases in `board-rule-cases.md`. The code follows the rules. For these 5 boards, the grids of the code replace the old grids after the user approves the new boards. The user approved the new boards `AgentsNested`, `TreeWide`, `TreeNested`, and `CompactTreeWide` on 2026-10-22. The user approved the new board `AgentsCoordinatorChild`, which shows rule A, on 2026-10-22. The approved code grids of the 3 boards with rule cases and the 2 boards of AC6 replaced the old grids.

### Mouse and chat modes

The boards `ChatNormal`, `ChatInput`, and `CompactChatNormal` show these decisions. The user approved them on 2026-10-22 with the plan `plans/hive-ask.md`.

79. The TUI turns on mouse reports in cell motion mode. This mode gives clicks and wheel steps. To select text for a copy, hold shift and drag. The terminal then selects the text.

80. On each tab, the wheel scrolls the region below the pointer. One wheel step moves the selection of a list by one row. One wheel step scrolls a text region by 3 lines. The text regions are the Chat feed, a details region, the JSON panel, the help overlay, and the details popup. A wheel step over a region that cannot scroll does nothing. The wheel does not move the Agents graph, as decision 77 gives. The wheel does not change the focus.

81. A click on a tab name opens that tab. A click on a region gives the focus to that region. A click on a list row selects that row. A click below the last row selects the last row. The Agents graph has one region, so a click on the graph does nothing. The TUI writes a log event for each mouse action that changes the screen. The event has the cell of the pointer and the focus.

82. This decision replaces decision 24 and the Chat input row of decision 21. Chat has two modes: normal mode and input mode. Normal mode is the default. In normal mode, the feed has the `accent` border, and the input has the `border` border. When the input has no text, it shows `i or click here to type` in `dim`. The footer starts with `[normal]`.

83. In normal mode, `i` starts input mode. A click on the input also starts input mode. `↑`, `↓`, `j`, `k`, and the wheel scroll the feed. `pgup` and `pgdn` page the feed. `g` and `G` go to the top and the end. `ctrl+o` and `?` keep their functions. Each other letter keeps its global function and does not go into the input. `enter`, `esc`, and `tab` do not start input mode.

84. In input mode, the input has the `accent` border and a bold title. The feed has the `border` border. The input shows `> ` and the cursor. The footer starts with `[input]`. Each printable key goes into the input. `enter` sends the text and keeps input mode. `shift+enter` adds a line. `ctrl+]` stops input mode. A click on the feed also stops input mode. The wheel, `pgup`, and `pgdn` scroll the feed. `esc` and `tab` do not change the mode. When input mode stops, the text stays in the input.

85. From 120 columns, the footer hints of normal mode are `i type · ↑↓ j/k scroll · wheel scroll · pgup/pgdn page · g/G top/end · ctrl+o events · ? help`. The footer hints of input mode are `enter send · shift+enter newline · ctrl+] leave input · click feed leave · wheel or pgup/pgdn scroll`. Below 120 columns, the hints of normal mode are `i type · ↑↓ wheel scroll · pgup/pgdn · ctrl+o events · ? help`. The hints of input mode are `enter send · ctrl+] leave · wheel or pgup/pgdn scroll`. The CHAT line of the Help board is `i or a click on the input types · ctrl+] or a click on the feed leaves`.

86. The top border of the feed shows ` ↑ N earlier ` at each width, from the middle column. N is the count of messages above the view.

87. The header shows `? N asks` before `needs you` on each tab, also when N is 0. For one ask, the text is `? 1 ask`. Below 120 columns, the header shows `? N`. The count is `accent` and bold when N is more than 0. In the other cases, it is `fg` and bold. The header state holds the count. Task 2 of the goal `g6-hive-ask` sets it.

88. The three boards break 14 rules that exist. The evidence of task `126-chat-modes-mouse` lists the 14 cases in `board-rule-cases.md`. The code follows the rules.

## Waves

1. **Spec.** Land this README and the board copies. Add the README to `steFiles`.
2. **Harness.** The converter, the exporter, the golden grids, the fixture files with their UI state, the demo loader, `--no-startup`, `--demo`, and `--size`. The worker and verifier skill lines for `--no-startup`. A person reviews the golden grids at the end of this wave.
3. **Components, part 1.** Header, TabBar, SubTabs, Panel, Viewport with ScrollMark, and KeyHints. The symbol width test.
4. **Components, part 2.** List, Details, FilterBar, Input, Card, Popup, Notice, and Markdown with the Glamour style.
5. **Frame and layout.** The frame, focus, the key routes, the breakpoints, the overlay below 120 columns, the too-small notice, and the help overlay. This task owns the root-model wiring and the screen registry.
6. **Screens.** One task for each of Agents, Chat, Friction, Health with the restart move, and Logs. Work has one task for each of Goals, Queue, and Tasks. These tasks can run in parallel. Each task passes the golden check and the live check before it lands.
7. **Sign-off.** The side-by-side review of each screen and the user sign-off.
8. **Pull request.** The screenshots and the update of pull request #47. Refer to decisions 51 to 54.

## Fog of war

- The exporter needs Node.js on the machine that runs wave 2. The gate does not need it.
- `freeze` needs a font that has each symbol in decision 13. Wave 8 checks each PNG for a missing symbol before the manager commits it.
- Herdr can give a pane that is smaller than the board. Decision 36 then shows the too-small notice, and the verifier needs a larger pane.
