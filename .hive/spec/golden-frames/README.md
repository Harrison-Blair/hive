# Golden frames

## Destination

Bare `hive` opens one Chat screen. The screen has the style of the Claude Code, Codex, and Pi terminals. The composer always has the focus. A chord opens each other screen at full size. Each other screen shows its state at a glance. The code matches the Golden frames boards cell for cell.

## Design authority

The page "Golden frames" of the canvas "Hive TUI Redesign" is the design authority: https://claude.ai/artifact/VrhcA98g9G2G9cFQao5Z9h. The user approved the plan on 2026-10-23. When a board and this folder disagree, stop and ask the user.

Each board is a grid of terminal cells. The script in the board writes each cell with a character, a palette token, and a weight. The boards use the mechanism of `.hive/spec/tui-v2/README.md`, decisions 39 to 46. A later task copies each board into `design/`, exports its golden grid, and writes its fixture.

## Files

| File | Content |
| --- | --- |
| `README.md` | The frame, the status row, the notice line, the keys, the palette, the boards, and the deferred items. |
| `screens/chat.md` | The Chat screen: the transcript, the composer, the ask card, and the states. |
| `screens/work.md` | The Work screen: the wave board and the states. |
| `screens/friction.md` | The Friction screen: the list, the resolved fold, and the states. |
| `screens/health.md` | The Health screen: the check grid, the agents table, the restart prompt, and the states. |
| `screens/logs.md` | The Logs screen: the density strip, the list, the field panel, and the states. |
| `screens/agents.md` | The Agents screen: the changes to the current rules and the states. |
| `screens/help.md` | The Help overlay. |

Each screen file has a state table, a key table, and numbered rules. A state names the board that shows it, or `no board`. A state with `no board` has its rule in words. A worker reads the screen file of the task and the boards that it names.

## Decisions

### Frame

1. Chat is the home screen. Bare `hive` opens Chat. A chord opens each other screen at full size. `esc` on another screen returns to Chat. The same chord on its own screen also returns to Chat.
2. No screen has a header row or a tab bar. Each screen except Chat starts with a title row on row 0 and a rule on row 1. The title row shows the screen name in `accent` bold, then facts in `dim`, then a right text in `dim`. Chat starts with the transcript on row 0.
3. The last rows of each screen are the status rows of decisions 8 to 13. The row above them is the notice line of decision 14, when a notice exists.
4. The minimum size is 80×24. Below it, the TUI shows the too-small notice of `tui-v2` decision 7. This document gives the 120×40 form of each screen. Decision 27 defers the 80×24 form of each screen.
5. A popup or an overlay dims the full screen behind it. A resize with a popup open keeps the popup and its scroll position. A resize never breaks a border.
6. A screen with two boxes has one focused box. The focused box has the `accent` border. `ctrl+]` moves the focus to the other box, both ways. The focused box takes `↑`, `↓`, `pgup`, `pgdn`, `g`, and `G`. The mouse wheel scrolls the box below the pointer and does not change the focus. Chat has no second box, so `ctrl+]` does nothing on Chat.
7. Each component measures text with `lipgloss.Width`. Components use only the symbols of `tui-v2` decision 13 and these symbols: `▮ ▁ ▂ ▃ ▄ ▅ ▆ ▇`. Each symbol takes one cell.

### Status rows

8. The status rows are the last rows of each screen. The default is one row. The config key `[tui.status] lines` gives one to three rows. Each row is a list of segment names. When the pane has fewer than 30 rows, the TUI shows only the first row.
9. The segments are `state`, `supervisor`, `asks`, `needs_you`, `project`, `screen`, `chords`, `hints`, `clock`, and `command`. The default first row is `state`, `supervisor`, `asks`, `needs_you`, `chords`, `hints`, `clock`.
10. The row joins the segments with ` · ` in `dim`. The `clock` segment sits at the right edge. When the row is wider than the pane, the TUI removes segments from the right, before the clock, until the row fits. The TUI never removes the first segment.
11. `state` shows the coordinator state. The forms are `● working 2m` with `●` in `working`, `○ idle 5m` with `○` in `dim`, and `▲ blocked 1m` with `▲` in `accent`. Two more forms are `● starting` with `●` in `dim` and `○ none` in `dim`. The age is the whole minutes from `status_changed_at` to the TUI clock. `supervisor` shows `● supervisor` with the same glyph rules. `asks` shows `? N asks`, or `? 1 ask`, in `accent` bold when N is more than 0. `needs_you` shows `needs you N`, in `accent` bold when N is more than 0. `project` shows the selected project ID. `screen` shows the screen name in `accent` bold, then the keys of the screen in `dim`. `chords` shows `^g agents  ^t work  ^f friction  ^y health  ^l logs` with two spaces between the chords. On a screen that is not Chat, its own chord reads `esc chat`. `hints` shows `? help`. `clock` shows the TUI clock as `HH:MM`.
12. `command` runs the command of the config key `[tui.status] command`. The TUI writes one JSON object to the standard input of the command. The object has `home`, `screen`, `coordinator`, `supervisor`, `asks`, `needs_you`, `project`, and `clock`. `coordinator` and `supervisor` each have `state` and `changed_at`. The TUI shows the first line of the standard output in the segment. The TUI keeps ANSI colors in that line. The TUI runs the command at most once each 300 ms and never on the input loop. When the command fails, the segment is empty and the TUI logs the error once. When the config has no command, the segment is empty.
13. In demo mode, the status rows read the fixture. The `command` segment runs no process in demo mode.

### Notice line

14. A notice line shows above the composer on Chat and above the status rows on each other screen. The line exists only while a notice exists. The first notice that applies wins. The order is the quit notice `press ctrl+c again to quit`, the demo notice, the restart prompt, the restart notice, the startup notice, the design notice, and then `error: TEXT`. A notice uses `accent`. The restart prompt uses `accent` bold. The prefix `error:` uses `error` bold and the text uses `fg`.
15. The quit notice and the demo notice go away on the next key. The restart prompt goes away on the key that answers it. An error goes away when its cause clears, as the code does today.

### Keys

16. The chords are `ctrl+g` Agents, `ctrl+t` Work, `ctrl+f` Friction, `ctrl+y` Health, and `ctrl+l` Logs. `ctrl+o` opens or folds all chat events. `ctrl+r` moves the focus between the input line and the ask card. `ctrl+]` is the box switch of decision 6. `ctrl+c` twice quits. The TUI takes `ctrl+g`, `ctrl+t`, and `ctrl+f` from the textarea key map. `→` keeps character forward.
17. `?` in an empty composer opens Help. On each other screen, `?` opens Help. `esc` closes a popup or an overlay first. With no popup open, `esc` returns to Chat. On Chat, `esc` on the ask card gives the focus to the input line. `esc` on the input line does nothing.
18. The config table `[tui.keys]` can rebind each chord of decision 16 and `?`. A key name follows the Bubble Tea key names. Hive refuses a config that binds one key to two actions. Hive refuses a config that binds `tab`. Decision 30 keeps `tab` for autocomplete.
19. There are no function keys. There is no `tab` region move. There is no normal mode on Chat. The keys `i`, `j`, `k`, `g`, `G`, and `ctrl+]` have no function on Chat.
20. The mouse rules of `tui-v2` decisions 79 to 81 stay. A click on a transcript event opens or folds that event. A click on the ask card gives the focus to the card. A click on the input line gives the focus to the input line.

### Palette

21. The palette has the eight tokens of `tui-v2` decision 26 and one more token, `error`, hex `#e8705f`. The token `error` colors the `ERROR` log rows, the error cell of the density strip, and the `error:` prefix of the notice line. The grid letter of `error` is `e`. The converter maps `#e8705f` to `error`.
22. The transcript markers are `›` for the user in `user` bold and `●` for an agent in `info` bold. A folded event has `▸` and an open event has `▾`, in `dim`. An ask has `?` in `accent` bold.

### Boards

23. The page "Golden frames" has these boards. Each board is 120×40 cells unless the table says otherwise. The `Gf` prefix keeps each stem unique on the canvas.

| Board | Screen and state |
| --- | --- |
| `GfChat` | Chat, coordinator working, typed text in the composer |
| `GfChatAsk` | Chat, one open ask inside the composer frame, the card has the focus |
| `GfWork` | Work, the wave board of one goal |
| `GfFriction` | Friction, the open list and the details box |
| `GfHealth` | Health, one warn expanded, the agents table |
| `GfLogs` | Logs, the density strip, the list, the field panel |
| `GfAgents` | Agents, the overview with wave progress in the pod headers |
| `GfHelp` | Help, the overlay over Chat |
| `GfComponents` | The component sheet, 120×67 |

The state boards are on the same page, in rows by screen. Each screen file names the state boards of its screen in its state table.

| Row | Boards |
| --- | --- |
| Chat states | `GfChatEmpty`, `GfChatNoCoordinator`, `GfChatEvents`, `GfChatTall`, `GfChatMarkdown`, `GfChatNotice`, `GfChatAskInput`, `GfChatAskMulti`, `GfChatAskPreview`, `GfChatAskStopped`, `GfChatAskReview`, `GfChatAskMany` |
| Work states | `GfWorkEmpty`, `GfWorkNoGoal`, `GfWorkScroll`, `GfWorkStops`, `GfWorkCancelled`, `GfWorkTask`, `GfWorkDeps` |
| Friction and Health states | `GfFrictionResolved`, `GfFrictionEmpty`, `GfFrictionFilter`, `GfHealthGreen`, `GfHealthRestart`, `GfHealthChecking` |
| Logs and Agents states | `GfLogsEmpty`, `GfLogsRaw`, `GfLogsError`, `GfAgentsPod`, `GfAgentsDetails`, `GfAgentsNone` |

### Out of scope

24. New data sources. Each screen shows data that hive stores today.
25. TUI write actions other than restart and chat. There is no key that opens a file in an editor.
26. A light theme.

### Deferred

27. The 80×24 form of each screen.
28. Logs grouped by request.
29. A queue for a chat message while no coordinator is live.
30. Autocomplete on `tab`.
31. Message history on `↑` in an empty composer.
32. The spec tasks, the briefs, the fixtures, and the golden grids. They come after the user signs off the boards.

### Changes after sign-off

33. On 2026-10-23 the user answered ask a18: the boards give the status row of an open ask. While the card has the focus, the row starts with `state`, then the focus name in `accent` bold, then the keys of that focus in `dim`. Then the row has the `asks` segment when more than one ask is open, and `clock`. When the row is wider than the pane, the TUI removes keys from the right, before the `asks` segment. The focus names and their keys are:
    - `ask`: `↑↓ move`, `1..N pick`, `enter review`, `n note`, `x dismiss`, `J/K card`, `^r input`, `esc input`. N is the number of option rows. When the asker stopped, `enter review` is `enter review for parent`.
    - `note`: `type the note`, `enter keep note`, `↑↓ option`, `space mark`, `^r input`. `space mark` shows only for a multi-select question.
    - `review`: `↑↓ question`, `enter send`, `x dismiss`, `J/K card`, `^r input`.

    While an ask is open and the input line has the focus, the row is `state`, `supervisor`, `asks`, `needs_you`, `^r answer the ask` in `accent`, and `clock`. The segment `^r answer the ask` is in place of `chords` and `hints`. The boards `GfChatAsk`, `GfChatAskInput`, `GfChatAskMulti`, `GfChatAskStopped`, `GfChatAskReview`, and `GfChatAskMany` show these rows.
