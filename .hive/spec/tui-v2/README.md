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
5. Only regions scroll. A region draws its scroll mark `┃` on its right border. A region that scrolls sideways draws `━` on its bottom border. The scroll mark uses `accent` in the focused region and `dim` in the other regions. An overlay or a popup dims the full screen behind it, including the header and the tab bar.
6. One region has focus. The focused region has an accent border. `tab` moves the focus to the next region.

### Fit to screen

7. The minimum size is 80×24. Below it, the TUI shows the too-small notice with the current size. It also shows ` hive` on the first row and `ctrl+c ×2 quit` on the last row. It shows no other content.
8. From 120 columns, a list and its details sit side by side.
9. From 80 to 119 columns, details open as a full-width overlay over the list. `enter` opens the overlay. `esc` closes it.
10. Agents gives each project one column of equal width. The minimum column width is 26 cells. Each card is 24×4 cells. The idle dot is `◐` in `accent` on each screen. When the columns do not fit, the graph scrolls sideways and keeps the selected card visible. With no project, Agents shows the notice "no project agents".
11. A resize never breaks a border. A resize with an overlay open keeps the overlay and its scroll position.
12. Each component measures text with `lipgloss.Width`. No component measures text with `len`.
13. Components use only the symbols in this list: `─ │ ╭ ╮ ╰ ╯ ├ ┤ ┼ ┴ ┬ └ ━ ┃ ● ◐ ○ ▶ ✓ • · – — … ↑ ↓ ← → › █ ×`. Each symbol takes one cell. Wave 2 checks that each exported board uses only these symbols. A test checks that `lipgloss.Width` gives 1 for each symbol. The TUI supports terminals that draw ambiguous-width characters as one cell. A new symbol needs a change to this list.

### Screens

14. **Agents** shows a full-width graph of project agents only. The coordinator and the supervisor are not on the graph. `enter` opens the details popup. `p` peeks at the pane. `a` attaches. `h` hides stopped agents.
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

The manager made this decision in task `29-frame-layout`:

- A live check compares a capture of a pane with a board grid. In a space cell of a live capture, the check does not compare the token or the bold flag. A terminal does not show them. It compares each other cell fully. The golden check at the model compares each cell fully.

The manager made this decision in task `40-screen-logs`:

- A row of the Logs event list shows the `msg` of the event. When the event has the field `reason`, the row also shows ` · ` and the text of the reason before the first `;`. The JSON panel shows the full reason.

The worker made these decisions in task `40-screen-logs`:

- Screen: the file `internal/tui/logs_screen.go`. The frame gives the content area of tab 6 to `logsContent`. From 120 columns, the event list has 8/15 of the width.
- Filter: the filter bar text is the applied filter. The default text is `level:info`. `level:` gives the lowest level. `agent:` matches the agent ID, or the agent name before its generation. Each other typed field must be equal to the field of the event. Each word of free text must be in the row text.
- Selection: when a load gives a new newest event, follow selects it. When the newest event does not change, the selection stays. When no row shows, the selected event stays, and it shows again with the rows. A day change selects the first row.
- Keys: `/` opens the filter bar with the applied text. `esc` in the event list sets the default filter. `f` changes follow. When follow is on, each load selects the newest event. `↑`, `k`, `pgup`, and `g` set follow to off. `G` sets follow to on.
- This task removes the old keys `l`, `c`, `p`, `a`, and `r` of the Logs screen. Decisions 21 to 25 give these keys other meanings or give the typed fields.
- The JSON panel cuts each line to the panel width less 6 cells, with `…`. When the lines do not fit, the last body row shows `↓ N more lines`.
- Component change: `FilterBar` has the fields `Focused` and `Status`. `Status` is at the end of the bar. The Logs screen shows the follow mark and the day there.

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
