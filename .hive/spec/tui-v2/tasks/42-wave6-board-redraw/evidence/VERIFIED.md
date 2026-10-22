# Independent verification: 42-wave6-board-redraw

Verifier: t42-verifier-1. Tested commit: 792f6570.

| Item | Result | Observation |
| --- | --- | --- |
| V1 Boards | pass | All 21 plans boards equal their spec copies. Main, Logs, and Components PNG pairs are byte-identical. No capture shows `{{`. |
| V2 Grid diff | pass | 4,575 cells change across 20 grids. The changed board and grid name sets are equal. TooSmall stays unchanged. The second export exits 0 with no design diff. |
| V3 Shared rules | pass | The three new tests fail with dev implementations and pass with task implementations. Seven deliberate breaks each fail tests. Restored tests pass. Old Chat, CompactChat, and popup sample inputs fail against the new boards; restored inputs pass. |
| V-gate Diff scope and gate | pass | `git merge dev` reports Already up to date. The diff contains only permitted source files, boards, grids, brief, baseline, and evidence. No fixture changes. All five gate commands exit 0. |

## V1 capture and plain-text check

For each plans `*.dc.html`, ran `cmp` against the spec copy: all exit 0.
For Main, Logs, and Components, created a distinct scratch profile folder for
each capture and ran:

```sh
firefox --headless --no-remote --profile SCRATCH --window-size=1220,780 --screenshot OUT.png file://BOARD.dc.html
```

The first attempt had no profile folder and failed. Created the folder and
repeated all six captures successfully. Read all three spec PNGs. Each plans
PNG has identical bytes. Components is taller than the required viewport;
its lower elements are checked in the exported plain-text grid and golden tests.
No screen renderer changes in this task.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Main header | working, live, needs you 0 | matches | Identical PNG pair |
| Main tabs | 1 Agents, underline | matches | Identical PNG pair |
| Main outer panel | project agents · 1 project | matches | Report target |
| Main project pod | hive, g1-tui-restart · done | matches | Report target |
| Main manager card | ◐, idle | matches | State has a glyph and word |
| Main child cards | ○, stopped, done, pass | matches | State has a glyph and word |
| Main tree connectors | ┬, │, branch bus | matches | Report tree form |
| Main stopped note | h hides 2 stopped | matches | Report target |
| Main legend | ● working, ◐ idle, ○ stopped | matches | Words remain without colour |
| Main footer | [graph], named keys | matches | Identical PNG pair |
| Logs header and tabs | working, live, 6 Logs underline | matches | Identical PNG pair |
| Logs filter | / level:debug proj:hive | matches | Identical PNG pair |
| Logs follow and day | follow ●, date, arrows | matches | Glyph remains without colour |
| Logs events panel | events · 312, DEBUG, WARN | matches | Words remain without colour |
| Logs selected event | ▶, verify fail · V4 diff scope | matches | Report row 30 |
| Logs JSON panel | event 13:38:02, WARN, fail | matches | Identical PNG pair |
| Logs scroll marks | ┃ versus │ | matches | Report target |
| Logs footer | [json], named keys | matches | Identical PNG pair |
| Components palette | fg, border, dim, accent, working, done, info, user, code | matches | Labels remain without colour |
| Components header | working, live, needs you 0 | matches | Identical PNG pair |
| Components tabs and subtabs | underline, [Tasks] | matches | Text and decoration remain |
| Components panels | unfocused, focused | matches | Words remain |
| Components viewports | viewport, horizontal, ┃, ━ | matches | Distinct mark glyphs |
| Components list | ▶, landed | matches | Full-column style checked by cell tests |
| Components details | status landed, ACCEPTANCE, ✓ | matches | Words and glyph remain |
| Components filter | / level:warn agent:manager-hive | matches | Identical PNG pair |
| Components lower examples | labels and state words in grid | matches | Plain-text grid and component golden tests |

PNG SHA-256 (plans and spec have the same hash):

- Main: `7c27ebc0e7881199901471248b0822060daf46c001323c66c71291b9e2469e75`
- Logs: `0c4abba8829f58cf32ed3ae9dc28064d9ddfd1df21120ef9bd96120d8221063f`
- Components: `f57528eabd56aaf3e0ba0489cf15cd0bc351dfe28990c63ba2951da44a12b29e`

## V2 target samples

Compared the independently rerun cellgrid diff below with the report targets:

- Main: 238 cells, including the project title and tree layout.
- Components row 36 col 3: accent bold space becomes fg normal;
  col 29 becomes accent bold. Decisions 4 and 7.
- Compact: row 20 becomes blank; row 21 shows `↓ 2 more lines`.
- Work row 27: `pass` ends at col 98, with a dim space at col 94.
- Friction: row 30 shows MITIGATION; row 37 shows `↓ 5 more lines`.
- Logs: row 8 starts with 13:04:36; selected row 30 is 13:38:02;
  event scroll mark is at rows 36–37 col 63.
- WorkGoals/WorkQueue also change unselected marker spaces to fg under the
  shared rule of their approved plans boards. No fixture changes.
- The current brief AC3 supersedes the report's older fixture-change request.

Ran `go run ./internal/boardexport/cmd/boardexport .hive/spec/tui-v2/design`:
exit 0; `git status --short .hive/spec/tui-v2/design` prints nothing.

The scratch diff program uses cellgrid.ReadGrid and cellgrid.Compare without
masks. It uses the worker's documented grouping program for output only.

### Cell-by-cell diff grouped into adjacent runs

```text
AgentDetails.grid: 324 cells
  row 3 cols 18-29: "────────────" [border] -> "· 1 project " [border]
  row 5 cols 36-59: "· g1-tui-restart · done " [border] -> "────────────────────────" [border]
  row 6 cols 30-43: "              " [border] -> "g1-tui-restart" [border]
  row 6 cols 45-45: " " [border] -> "·" [border]
  row 6 cols 47-50: "    " [border] -> "done" [border]
  row 24 cols 34-35: "  " [fg] -> "  " [dim]
  row 26 cols 33-38: "      " [fg] -> "RECENT" [accent+b]
  row 27 cols 33-43: "RECENT     " [accent+b,fg] -> "14:08 inbox" [fg]
  row 27 cols 45-46: "  " [fg] -> "28" [fg]
  row 27 cols 48-51: "    " [fg] -> "sent" [fg]
  row 27 cols 53-54: "  " [fg] -> "to" [fg]
  row 27 cols 56-66: "           " [fg] -> "coordinator" [fg]
  row 28 cols 33-76: "14:08 inbox 28 sent to coordinator          " [fg] -> "14:07 land 4-agents-equal-columns · ede94f3e" [dim]
  row 29 cols 34-34: "4" [dim] -> "3" [dim]
  row 29 cols 36-37: "07" [dim] -> "58" [dim]
  row 29 cols 44-44: "4" [dim] -> "3" [dim]
  row 29 cols 46-86: "agents-equal-columns · ede94f3e          " [dim] -> "tui-restart · 3ccd93fb                   " [dim,fg]
  row 30 cols 36-37: "58" [dim] -> "44" [dim]
  row 30 cols 44-44: "3" [dim] -> "2" [dim]
  row 30 cols 46-54: "tui-resta" [dim] -> "stale-coo" [dim]
  row 30 cols 56-70: "t · 3ccd93fb   " [dim] -> "dinator-startup" [dim]
  row 30 cols 72-72: " " [dim] -> "·" [dim]
  row 30 cols 74-86: "             " [dim] -> "0e079272     " [dim,fg]
  row 31 cols 33-86: "13:44 land 2-stale-coordinator-startup · 0e079272     " [dim] -> "                                                      " [fg]
  row 37 cols 2-3: "  " [border] -> "──" [border]
  row 37 cols 5-9: "     " [border] -> "spawn" [border]
  row 37 cols 13-13: " " [border] -> "●" [border]
  row 37 cols 15-21: "       " [border] -> "working" [border]
  row 37 cols 25-25: " " [border] -> "◐" [border]
  row 37 cols 27-30: "    " [border] -> "idle" [border]
  row 37 cols 34-34: " " [border] -> "○" [border]
  row 37 cols 36-42: "       " [border] -> "stopped" [border]
  row 37 cols 46-48: "   " [border] -> "one" [border]
  row 37 cols 50-55: "      " [border] -> "column" [border]
  row 37 cols 57-59: "   " [border] -> "per" [border]
  row 37 cols 61-67: "       " [border] -> "project" [border]
AgentsMany.grid: 155 cells
  row 5 cols 4-9: " hive " [fg+b] -> " hive " [dim+b]
  row 6 cols 22-27: "      " [dim] -> "      " [fg]
  row 6 cols 56-56: " " [dim] -> " " [fg]
  row 6 cols 109-114: "      " [dim] -> "      " [fg]
  row 9 cols 82-83: "  " [dim] -> "  " [fg]
  row 9 cols 108-112: "     " [dim] -> "     " [fg]
  row 12 cols 91-100: "          " [fg] -> "no workers" [dim]
  row 13 cols 20-26: "       " [accent+b] -> "       " [fg]
  row 13 cols 78-84: "       " [dim] -> "       " [fg]
  row 13 cols 92-101: "no workers" [dim] -> "          " [fg]
  row 14 cols 24-26: "   " [dim] -> "   " [fg]
  row 14 cols 51-55: "     " [dim] -> "     " [fg]
  row 14 cols 84-84: " " [dim] -> " " [fg]
  row 17 cols 62-78: "                 " [fg] -> "h hides 1 stopped" [dim]
  row 19 cols 24-26: "   " [dim] -> "   " [fg]
  row 19 cols 50-55: "      " [dim] -> "      " [fg]
  row 23 cols 22-26: "     " [dim] -> "     " [fg]
  row 24 cols 26-26: " " [dim] -> " " [fg]
  row 27 cols 4-20: "                 " [fg] -> "h hides 1 stopped" [dim]
  row 36 cols 46-57: "columns shar" [dim] -> "one column p" [dim]
  row 36 cols 59-66: " one wid" [dim] -> "r projec" [dim]
  row 36 cols 68-68: "h" [dim] -> " " [fg]
  row 37 cols 101-108: "example-" [dim] -> "     exa" [fg,dim]
  row 37 cols 110-117: "l →     " [dim,fg] -> "ple-ml →" [dim]
Chat.grid: 825 cells
  row 13 cols 23-26: "Bare" [fg] -> "    " [fg]
  row 13 cols 28-31: "hive" [fg] -> "    " [fg]
  row 13 cols 33-37: "makes" [fg] -> "     " [fg]
  row 13 cols 39-41: "the" [fg] -> "   " [fg]
  row 13 cols 43-53: "coordinator" [code] -> "           " [fg]
  row 13 cols 55-58: "tab," [fg] -> "    " [fg]
  row 13 cols 60-62: "but" [fg] -> "   " [fg]
  row 13 cols 64-66: "its" [fg] -> "   " [fg]
  row 13 cols 68-73: "record" [fg] -> "      " [fg]
  row 13 cols 75-77: "has" [fg] -> "   " [fg]
  row 13 cols 79-92: "spawned=false." [code,fg] -> "              " [fg]
  row 13 cols 94-103: "agent stop" [code] -> "          " [fg]
  row 13 cols 105-111: "refuses" [fg] -> "       " [fg]
  row 13 cols 113-115: "it." [fg] -> "   " [fg]
  row 14 cols 23-26: "    " [fg] -> "Bare" [fg]
  row 14 cols 28-31: "    " [fg] -> "hive" [fg]
  row 14 cols 33-37: "     " [fg] -> "makes" [fg]
  row 14 cols 39-41: "   " [fg] -> "the" [fg]
  row 14 cols 43-53: "           " [fg] -> "coordinator" [code]
  row 14 cols 55-58: "    " [fg] -> "tab," [fg]
  row 14 cols 60-62: "   " [fg] -> "but" [fg]
  row 14 cols 64-66: "   " [fg] -> "its" [fg]
  row 14 cols 68-73: "      " [fg] -> "record" [fg]
  row 14 cols 75-77: "   " [fg] -> "has" [fg]
  row 14 cols 79-92: "              " [fg] -> "spawned=false." [code,fg]
  row 14 cols 94-103: "          " [fg] -> "agent stop" [code]
  row 14 cols 105-111: "       " [fg] -> "refuses" [fg]
  row 14 cols 113-115: "   " [fg] -> "it." [fg]
  row 15 cols 24-29: "Choice" [fg] -> "      " [fg]
  row 15 cols 42-42: "│" [border] -> " " [fg]
  row 15 cols 44-49: "Answer" [fg] -> "      " [fg]
  row 15 cols 78-78: "│" [border] -> " " [fg]
  row 15 cols 80-83: "Pros" [fg] -> "    " [fg]
  row 15 cols 97-97: "│" [border] -> " " [fg]
  row 15 cols 99-102: "Cons" [fg] -> "    " [fg]
  row 16 cols 23-116: "───────────────────┼───────────────────────────────────┼──────────────────┼───────────────────" [border] -> " Choice            │ Answer                            │ Pros             │ Cons              " [fg,border]
  row 17 cols 23-116: " A (REC)           │ Record that bare hive started it  │ No tab pile-up   │ Field change      " [fg,border] -> "───────────────────┼───────────────────────────────────┼──────────────────┼───────────────────" [border]
  row 18 cols 24-24: "B" [fg] -> "A" [fg]
  row 18 cols 26-30: "     " [fg] -> "(REC)" [fg]
  row 18 cols 44-44: "L" [fg] -> "R" [fg]
  row 18 cols 46-67: "ave the old tab, open " [fg] -> "cord that bare hive st" [fg]
  row 18 cols 69-70: " n" [fg] -> "rt" [fg]
  row 18 cols 72-72: "w" [fg] -> "d" [fg]
  row 18 cols 74-76: "one" [fg] -> "it " [fg]
  row 18 cols 80-94: "Smallest change" [fg] -> "No tab pile-up " [fg]
  row 18 cols 99-112: "Dead tabs pile" [fg] -> "Field change  " [fg]
  row 18 cols 114-115: "up" [fg] -> "  " [fg]
  row 19 cols 24-24: "C" [fg] -> "B" [fg]
  row 19 cols 44-44: "R" [fg] -> "L" [fg]
  row 19 cols 46-47: "us" [fg] -> "av" [fg]
  row 19 cols 54-66: "same pane    " [fg] -> "old tab, open" [fg]
  row 19 cols 68-68: " " [fg] -> "a" [fg]
  row 19 cols 70-72: "   " [fg] -> "new" [fg]
  row 19 cols 74-76: "   " [fg] -> "one" [fg]
  row 19 cols 80-90: "One tab alw" [fg] -> "Smallest ch" [fg]
  row 19 cols 92-94: "ys " [fg] -> "nge" [fg]
  row 19 cols 99-102: "Must" [fg] -> "Dead" [fg]
  row 19 cols 104-107: "exit" [fg] -> "tabs" [fg]
  row 19 cols 109-115: "harness" [fg] -> "pile up" [fg]
  row 20 cols 24-24: " " [fg] -> "C" [fg]
  row 20 cols 42-42: " " [fg] -> "│" [border]
  row 20 cols 44-48: "     " [fg] -> "Reuse" [fg]
  row 20 cols 50-52: "   " [fg] -> "the" [fg]
  row 20 cols 54-57: "    " [fg] -> "same" [fg]
  row 20 cols 59-62: "    " [fg] -> "pane" [fg]
  row 20 cols 78-78: " " [fg] -> "│" [border]
  row 20 cols 80-82: "   " [fg] -> "One" [fg]
  row 20 cols 84-86: "   " [fg] -> "tab" [fg]
  row 20 cols 88-93: "      " [fg] -> "always" [fg]
  row 20 cols 97-97: " " [fg] -> "│" [border]
  row 20 cols 99-102: "    " [fg] -> "Must" [fg]
  row 20 cols 104-107: "    " [fg] -> "exit" [fg]
  row 20 cols 109-115: "       " [fg] -> "harness" [fg]
  row 20 cols 119-119: "┃" [dim] -> "│" [border]
  row 21 cols 23-36: "• Recommended:" [accent,fg+b] -> "              " [fg]
  row 21 cols 38-39: "A," [fg] -> "  " [fg]
  row 21 cols 41-47: "because" [fg] -> "       " [fg]
  row 21 cols 49-52: "hive" [fg] -> "    " [fg]
  row 21 cols 54-56: "may" [fg] -> "   " [fg]
  row 21 cols 58-62: "close" [fg] -> "     " [fg]
  row 21 cols 64-67: "only" [fg] -> "    " [fg]
  row 21 cols 69-73: "panes" [fg] -> "     " [fg]
  row 21 cols 75-76: "it" [fg] -> "  " [fg]
  row 21 cols 78-85: "started." [fg] -> "        " [fg]
  row 21 cols 119-119: "┃" [dim] -> "│" [border]
  row 22 cols 25-62: "Reply like 1A, 2A, 3A, or simplify N. " [fg,code] -> "Recommended: A, because hive may close" [fg+b,fg]
  row 22 cols 64-67: "    " [fg] -> "only" [fg]
  row 22 cols 69-73: "     " [fg] -> "panes" [fg]
  row 22 cols 75-76: "  " [fg] -> "it" [fg]
  row 22 cols 78-85: "        " [fg] -> "started." [fg]
  row 22 cols 119-119: "┃" [dim] -> "│" [border]
  row 23 cols 23-29: "       " [fg] -> "• Reply" [accent,fg]
  row 23 cols 31-34: "    " [fg] -> "like" [fg]
  row 23 cols 36-46: "           " [fg] -> "1A, 2A, 3A," [code,fg]
  row 23 cols 48-49: "  " [fg] -> "or" [fg]
  row 23 cols 51-61: "           " [fg] -> "simplify N." [code,fg]
  row 23 cols 119-119: "┃" [dim] -> "│" [border]
  row 24 cols 25-52: "hive agent stop manager-hive" [code] -> "                            " [fg]
  row 24 cols 119-119: "┃" [dim] -> "│" [border]
  row 25 cols 37-55: "pawn --role manager" [code] -> "top manager-hive   " [code,fg]
  row 25 cols 119-119: "┃" [dim] -> "│" [border]
  row 26 cols 25-55: "                               " [fg] -> "hive agent spawn --role manager" [code]
  row 26 cols 119-119: "┃" [dim] -> "│" [border]
  row 27 cols 2-6: "14:22" [dim] -> "     " [fg]
  row 27 cols 9-21: "you          " [user+b] -> "             " [fg]
  row 27 cols 23-32: "1A, 2A, 3B" [user] -> "          " [fg]
  row 28 cols 2-6: "     " [fg] -> "14:22" [dim]
  row 28 cols 9-21: "             " [fg] -> "you          " [user+b]
  row 28 cols 23-32: "          " [fg] -> "1A, 2A, 3B" [user]
  row 34 cols 119-119: "│" [border] -> "┃" [dim]
Compact.grid: 37 cells
  row 3 cols 8-9: "  " [dim] -> "  " [fg]
  row 3 cols 17-18: "  " [dim] -> "  " [fg]
  row 5 cols 76-77: "  " [fg] -> "  " [done]
  row 17 cols 79-79: "│" [accent] -> "┃" [accent]
  row 18 cols 79-79: "│" [accent] -> "┃" [accent]
  row 19 cols 79-79: "│" [accent] -> "┃" [accent]
  row 20 cols 2-15: "↓ 3 more lines" [dim] -> "              " [fg]
  row 21 cols 2-15: "              " [fg] -> "↓ 2 more lines" [dim]
CompactAgents.grid: 819 cells
  row 4 cols 2-77: "╭─ hive · g1-tui-restart ──────────────────────────────────────────────────╮" [border,dim+b] -> "                                                                            " [fg]
  row 5 cols 2-77: "│                                                                          │" [border,fg] -> "╭─ hive ───────────────────────────────────────────────────────────────────╮" [border,dim+b]
  row 6 cols 4-24: "                     " [fg] -> "g1-tui-restart · done" [dim]
  row 6 cols 28-51: "╭──────────────────────╮" [accent] -> "                        " [fg]
  row 7 cols 28-51: "│ ◐ manager-hive       │" [accent,fg,accent+b] -> "╭──────────────────────╮" [accent]
  row 8 cols 30-49: "claude · idle · 4m  " [dim] -> "◐ manager-hive      " [accent,fg,accent+b]
  row 9 cols 28-51: "╰──────────────────────╯" [accent] -> "│ claude · idle · 4m   │" [accent,fg,dim]
  row 10 cols 28-51: "            │           " [fg,border] -> "╰───────────┬──────────╯" [accent]
  row 11 cols 16-64: "╭───────────────────────┴───────────────────────╮" [border] -> "                        │                        " [fg,border]
  row 12 cols 4-22: "╭──────────────────" [border] -> "                  ╭" [fg,border]
  row 12 cols 27-52: "╮                        ╭" [border,fg] -> "─────────────┴────────────" [border]
  row 12 cols 58-75: "─────────────────╮" [border] -> "╮                 " [border,fg]
  row 13 cols 4-4: "│" [border] -> " " [fg]
  row 13 cols 6-6: "○" [dim] -> " " [fg]
  row 13 cols 8-18: "t4-worker-1" [fg] -> "           " [fg]
  row 13 cols 22-22: " " [fg] -> "│" [border]
  row 13 cols 27-27: "│" [border] -> " " [fg]
  row 13 cols 52-52: "│" [border] -> " " [fg]
  row 13 cols 54-54: "○" [dim] -> " " [fg]
  row 13 cols 56-68: "t4-verifier-1" [fg] -> "  │          " [fg,border]
  row 13 cols 75-75: "│" [border] -> " " [fg]
  row 14 cols 4-4: "│" [border] -> " " [fg]
  row 14 cols 6-33: "pi · stopped         │      " [dim,fg,border] -> "    ╭───────────┴──────────╮" [fg,border]
  row 14 cols 46-73: "      │ pi · stopped        " [fg,border,dim] -> "╭───────────┴──────────╮    " [border,fg]
  row 14 cols 75-75: "│" [border] -> " " [fg]
  row 15 cols 4-27: "╰──────────────────────╯" [border] -> "      │ ○ t4-worker-1   " [fg,border,dim]
  row 15 cols 33-33: " " [fg] -> "│" [border]
  row 15 cols 46-46: " " [fg] -> "│" [border]
  row 15 cols 48-48: " " [fg] -> "○" [dim]
  row 15 cols 50-75: "  ╰──────────────────────╯" [fg,border] -> "t4-verifier-1      │      " [dim,fg,border]
  row 16 cols 10-10: " " [fg] -> "│" [border]
  row 16 cols 12-30: "                   " [fg] -> "pi · stopped · done" [dim]
  row 16 cols 33-33: " " [fg] -> "│" [border]
  row 16 cols 46-46: " " [fg] -> "│" [border]
  row 16 cols 48-66: "                   " [fg] -> "pi · stopped · pass" [dim]
  row 16 cols 69-69: " " [fg] -> "│" [border]
  row 17 cols 10-33: "                        " [fg] -> "╰──────────────────────╯" [border]
  row 17 cols 46-69: "                        " [fg] -> "╰──────────────────────╯" [border]
  row 18 cols 4-20: "h hides 2 stopped" [dim] -> "                 " [fg]
  row 19 cols 4-20: "                 " [fg] -> "h hides 2 stopped" [dim]
  row 20 cols 2-77: "│                                                                          │" [border,fg] -> "╰──────────────────────────────────────────────────────────────────────────╯" [border]
  row 21 cols 2-77: "╰──────────────────────────────────────────────────────────────────────────╯" [border] -> "                                                                            " [fg]
CompactChat.grid: 60 cells
  row 7 cols 58-71: "              " [fg] -> "spawned=false." [code,fg]
  row 7 cols 73-77: "     " [fg] -> "agent" [code]
  row 8 cols 3-26: "pawned=false. agent stop" [code,fg] -> "top refuses it.         " [code,fg]
  row 8 cols 28-34: "refuses" [fg] -> "       " [fg]
  row 8 cols 36-38: "it." [fg] -> "   " [fg]
  row 8 cols 79-79: "┃" [dim] -> "│" [border]
  row 9 cols 79-79: "┃" [dim] -> "│" [border]
  row 10 cols 79-79: "┃" [dim] -> "│" [border]
  row 11 cols 79-79: "┃" [dim] -> "│" [border]
  row 12 cols 79-79: "┃" [dim] -> "│" [border]
  row 13 cols 79-79: "┃" [dim] -> "│" [border]
  row 18 cols 79-79: "│" [border] -> "┃" [dim]
CompactFriction.grid: 23 cells
  row 6 cols 2-2: " " [accent+b] -> " " [fg]
  row 7 cols 2-2: " " [accent+b] -> " " [fg]
  row 8 cols 2-2: " " [accent+b] -> " " [fg]
  row 9 cols 2-2: " " [accent+b] -> " " [fg]
  row 11 cols 2-2: " " [accent+b] -> " " [fg]
  row 12 cols 2-2: " " [accent+b] -> " " [fg]
  row 13 cols 2-2: " " [accent+b] -> " " [fg]
  row 14 cols 2-2: " " [accent+b] -> " " [fg]
  row 15 cols 2-2: " " [accent+b] -> " " [fg]
  row 16 cols 2-2: " " [accent+b] -> " " [fg]
  row 17 cols 2-2: " " [accent+b] -> " " [fg]
  row 18 cols 2-2: " " [accent+b] -> " " [fg]
  row 19 cols 2-2: " " [accent+b] -> " " [fg]
  row 20 cols 2-2: " " [accent+b] -> " " [fg]
  row 23 cols 45-45: "/" [dim] -> "?" [dim]
  row 23 cols 47-52: "filter" [dim] -> "      " [dim]
  row 23 cols 54-54: "·" [dim] -> " " [dim]
  row 23 cols 56-56: "?" [dim] -> " " [dim]
CompactHealth.grid: 15 cells
  row 7 cols 2-2: " " [accent+b] -> " " [fg]
  row 8 cols 2-2: " " [accent+b] -> " " [fg]
  row 9 cols 2-2: " " [accent+b] -> " " [fg]
  row 10 cols 2-2: " " [accent+b] -> " " [fg]
  row 13 cols 68-78: "it is live." [dim] -> "live.      " [dim,fg]
CompactLogs.grid: 10 cells
  row 7 cols 79-79: "┃" [accent] -> "│" [accent]
  row 8 cols 79-79: "┃" [accent] -> "│" [accent]
  row 9 cols 79-79: "┃" [accent] -> "│" [accent]
  row 10 cols 79-79: "┃" [accent] -> "│" [accent]
  row 11 cols 79-79: "┃" [accent] -> "│" [accent]
  row 12 cols 79-79: "┃" [accent] -> "│" [accent]
  row 13 cols 79-79: "┃" [accent] -> "│" [accent]
  row 14 cols 79-79: "┃" [accent] -> "│" [accent]
  row 15 cols 79-79: "┃" [accent] -> "│" [accent]
  row 16 cols 79-79: "┃" [accent] -> "│" [accent]
Components.grid: 2 cells
  row 36 cols 3-3: " " [accent+b] -> " " [fg]
  row 36 cols 29-29: " " [fg] -> " " [accent+b]
Friction.grid: 353 cells
  row 5 cols 119-119: "│" [border] -> "┃" [dim]
  row 6 cols 2-2: " " [accent+b] -> " " [fg]
  row 7 cols 2-2: " " [accent+b] -> " " [fg]
  row 8 cols 2-2: " " [accent+b] -> " " [fg]
  row 9 cols 2-2: " " [accent+b] -> " " [fg]
  row 11 cols 2-2: " " [accent+b] -> " " [fg]
  row 12 cols 2-2: " " [accent+b] -> " " [fg]
  row 13 cols 2-2: " " [accent+b] -> " " [fg]
  row 14 cols 2-2: " " [accent+b] -> " " [fg]
  row 15 cols 2-2: " " [accent+b] -> " " [fg]
  row 16 cols 2-2: " " [accent+b] -> " " [fg]
  row 17 cols 2-2: " " [accent+b] -> " " [fg]
  row 18 cols 2-2: " " [accent+b] -> " " [fg]
  row 19 cols 2-2: " " [accent+b] -> " " [fg]
  row 20 cols 2-2: " " [accent+b] -> " " [fg]
  row 24 cols 119-119: "│" [border] -> "┃" [dim]
  row 25 cols 119-119: "│" [border] -> "┃" [dim]
  row 26 cols 119-119: "│" [border] -> "┃" [dim]
  row 27 cols 119-119: "│" [border] -> "┃" [dim]
  row 28 cols 119-119: "│" [border] -> "┃" [dim]
  row 29 cols 119-119: "│" [border] -> "┃" [dim]
  row 30 cols 54-68: "↓ 12 more lines" [dim] -> "MITIGATION     " [accent+b,fg]
  row 30 cols 119-119: "│" [border] -> "┃" [dim]
  row 31 cols 54-57: "    " [fg] -> "Task" [fg]
  row 31 cols 59-70: "            " [fg] -> "8-no-startup" [fg]
  row 31 cols 72-73: "  " [fg] -> "of" [fg]
  row 31 cols 75-81: "       " [fg] -> "feature" [fg]
  row 31 cols 83-88: "      " [fg] -> "tui-v2" [fg]
  row 31 cols 90-93: "    " [fg] -> "adds" [fg]
  row 31 cols 95-97: "   " [fg] -> "the" [fg]
  row 31 cols 99-104: "      " [fg] -> "switch" [fg]
  row 31 cols 119-119: "│" [border] -> "┃" [dim]
  row 32 cols 54-57: "    " [fg] -> "hive" [fg]
  row 32 cols 59-70: "            " [fg] -> "--no-startup" [fg]
  row 32 cols 72-74: "   " [fg] -> "and" [fg]
  row 32 cols 76-78: "   " [fg] -> "the" [fg]
  row 32 cols 80-87: "        " [fg] -> "variable" [fg]
  row 32 cols 89-106: "                  " [fg] -> "HIVE_NO_STARTUP=1," [fg]
  row 32 cols 108-111: "    " [fg] -> "from" [fg]
  row 32 cols 119-119: "│" [border] -> "┃" [dim]
  row 33 cols 54-61: "        " [fg] -> "decision" [fg]
  row 33 cols 63-64: "  " [fg] -> "34" [fg]
  row 33 cols 66-67: "  " [fg] -> "of" [fg]
  row 33 cols 69-71: "   " [fg] -> "the" [fg]
  row 33 cols 73-77: "     " [fg] -> "spec." [fg]
  row 33 cols 79-82: "    " [fg] -> "With" [fg]
  row 33 cols 84-89: "      " [fg] -> "either" [fg]
  row 33 cols 91-94: "    " [fg] -> "one," [fg]
  row 33 cols 96-99: "    " [fg] -> "bare" [fg]
  row 33 cols 101-104: "    " [fg] -> "hive" [fg]
  row 33 cols 106-110: "     " [fg] -> "opens" [fg]
  row 33 cols 112-114: "   " [fg] -> "the" [fg]
  row 34 cols 54-57: "    " [fg] -> "TUI." [fg]
  row 34 cols 59-60: "  " [fg] -> "It" [fg]
  row 34 cols 62-66: "     " [fg] -> "makes" [fg]
  row 34 cols 68-69: "  " [fg] -> "no" [fg]
  row 34 cols 71-74: "    " [fg] -> "tab," [fg]
  row 34 cols 76-81: "      " [fg] -> "starts" [fg]
  row 34 cols 83-84: "  " [fg] -> "no" [fg]
  row 34 cols 86-96: "           " [fg] -> "coordinator" [fg]
  row 34 cols 98-100: "   " [fg] -> "and" [fg]
  row 34 cols 102-103: "  " [fg] -> "no" [fg]
  row 34 cols 105-115: "           " [fg] -> "supervisor," [fg]
  row 35 cols 54-56: "   " [fg] -> "and" [fg]
  row 35 cols 58-61: "    " [fg] -> "does" [fg]
  row 35 cols 63-65: "   " [fg] -> "not" [fg]
  row 35 cols 67-72: "      " [fg] -> "rename" [fg]
  row 35 cols 74-76: "   " [fg] -> "the" [fg]
  row 35 cols 78-81: "    " [fg] -> "tab." [fg]
  row 36 cols 54-56: "   " [fg] -> "The" [fg]
  row 36 cols 58-63: "      " [fg] -> "worker" [fg]
  row 36 cols 65-69: "     " [fg] -> "skill" [fg]
  row 36 cols 71-73: "   " [fg] -> "and" [fg]
  row 36 cols 75-77: "   " [fg] -> "the" [fg]
  row 36 cols 79-86: "        " [fg] -> "verifier" [fg]
  row 36 cols 88-92: "     " [fg] -> "skill" [fg]
  row 36 cols 94-97: "    " [fg] -> "tell" [fg]
  row 36 cols 99-102: "    " [fg] -> "each" [fg]
  row 36 cols 104-108: "     " [fg] -> "agent" [fg]
  row 36 cols 110-111: "  " [fg] -> "to" [fg]
  row 36 cols 113-115: "   " [fg] -> "run" [fg]
  row 37 cols 54-67: "              " [fg] -> "↓ 5 more lines" [dim]
  row 39 cols 55-76: "/ filter · g/G top/end" [dim] -> "g/G top/end · ? help  " [dim]
  row 39 cols 78-78: "·" [dim] -> " " [dim]
  row 39 cols 80-80: "?" [dim] -> " " [dim]
  row 39 cols 82-85: "help" [dim] -> "    " [dim]
FrictionResolved.grid: 325 cells
  row 6 cols 2-2: " " [accent+b] -> " " [fg]
  row 6 cols 51-51: "┃" [accent] -> "│" [accent]
  row 7 cols 37-37: " " [fg] -> " " [accent+b]
  row 7 cols 51-51: "┃" [accent] -> "│" [accent]
  row 8 cols 2-2: " " [accent+b] -> " " [fg]
  row 8 cols 51-51: "┃" [accent] -> "│" [accent]
  row 9 cols 2-2: " " [accent+b] -> " " [fg]
  row 9 cols 51-51: "┃" [accent] -> "│" [accent]
  row 10 cols 2-2: " " [accent+b] -> " " [fg]
  row 10 cols 51-51: "┃" [accent] -> "│" [accent]
  row 11 cols 2-2: " " [accent+b] -> " " [fg]
  row 11 cols 51-51: "┃" [accent] -> "│" [accent]
  row 12 cols 2-2: " " [accent+b] -> " " [fg]
  row 12 cols 51-51: "┃" [accent] -> "│" [accent]
  row 13 cols 2-2: " " [accent+b] -> " " [fg]
  row 13 cols 51-51: "┃" [accent] -> "│" [accent]
  row 14 cols 2-2: " " [accent+b] -> " " [fg]
  row 14 cols 51-51: "┃" [accent] -> "│" [accent]
  row 15 cols 2-2: " " [accent+b] -> " " [fg]
  row 15 cols 51-51: "┃" [accent] -> "│" [accent]
  row 16 cols 2-2: " " [accent+b] -> " " [fg]
  row 16 cols 36-37: "… " [fg] -> "s…" [fg]
  row 16 cols 51-51: "┃" [accent] -> "│" [accent]
  row 17 cols 2-2: " " [accent+b] -> " " [fg]
  row 17 cols 51-51: "┃" [accent] -> "│" [accent]
  row 18 cols 2-2: " " [accent+b] -> " " [fg]
  row 18 cols 36-37: "… " [fg] -> "sk" [fg]
  row 18 cols 51-51: "┃" [accent] -> "│" [accent]
  row 19 cols 2-2: " " [accent+b] -> " " [fg]
  row 19 cols 51-51: "┃" [accent] -> "│" [accent]
  row 20 cols 2-2: " " [accent+b] -> " " [fg]
  row 21 cols 2-2: " " [accent+b] -> " " [fg]
  row 21 cols 36-37: "… " [fg] -> "j…" [fg]
  row 22 cols 2-2: " " [accent+b] -> " " [fg]
  row 23 cols 2-2: " " [accent+b] -> " " [fg]
  row 23 cols 54-67: "↓ 9 more lines" [dim] -> "REPRODUCTION  " [accent+b,fg]
  row 24 cols 2-2: " " [accent+b] -> " " [fg]
  row 24 cols 36-37: "… " [fg] -> "k…" [fg]
  row 24 cols 54-55: "  " [fg] -> "1." [fg]
  row 24 cols 57-61: "     " [fg] -> "Start" [fg]
  row 24 cols 63-66: "    " [fg] -> "hive" [fg]
  row 24 cols 68-77: "          " [fg] -> "supervise." [fg]
  row 25 cols 2-2: " " [accent+b] -> " " [fg]
  row 25 cols 36-37: "… " [fg] -> "e…" [fg]
  row 25 cols 54-55: "  " [fg] -> "2." [fg]
  row 25 cols 57-60: "    " [fg] -> "Land" [fg]
  row 25 cols 62-62: " " [fg] -> "a" [fg]
  row 25 cols 64-69: "      " [fg] -> "change" [fg]
  row 25 cols 71-72: "  " [fg] -> "to" [fg]
  row 25 cols 74-76: "   " [fg] -> "the" [fg]
  row 25 cols 78-87: "          " [fg] -> "supervisor" [fg]
  row 25 cols 89-91: "   " [fg] -> "and" [fg]
  row 25 cols 93-95: "   " [fg] -> "run" [fg]
  row 25 cols 97-98: "  " [fg] -> "go" [fg]
  row 25 cols 100-106: "       " [fg] -> "install" [fg]
  row 25 cols 108-109: "  " [fg] -> "or" [fg]
  row 26 cols 57-58: "  " [fg] -> "go" [fg]
  row 26 cols 60-64: "     " [fg] -> "build" [fg]
  row 26 cols 66-68: "   " [fg] -> "for" [fg]
  row 26 cols 70-74: "     " [fg] -> "hive." [fg]
  row 27 cols 54-55: "  " [fg] -> "3." [fg]
  row 27 cols 57-59: "   " [fg] -> "Run" [fg]
  row 27 cols 61-64: "    " [fg] -> "hive" [fg]
  row 27 cols 66-72: "       " [fg] -> "doctor." [fg]
  row 27 cols 74-75: "  " [fg] -> "It" [fg]
  row 27 cols 77-83: "       " [fg] -> "reports" [fg]
  row 27 cols 85-87: "   " [fg] -> "the" [fg]
  row 27 cols 89-91: "   " [fg] -> "old" [fg]
  row 27 cols 93-102: "          " [fg] -> "supervisor" [fg]
  row 27 cols 104-105: "  " [fg] -> "as" [fg]
  row 27 cols 107-111: "     " [fg] -> "live." [fg]
  row 28 cols 54-55: "  " [fg] -> "4." [fg]
  row 28 cols 57-63: "       " [fg] -> "Trigger" [fg]
  row 28 cols 65-67: "   " [fg] -> "the" [fg]
  row 28 cols 69-75: "       " [fg] -> "changed" [fg]
  row 28 cols 77-85: "         " [fg] -> "behavior." [fg]
  row 28 cols 87-89: "   " [fg] -> "The" [fg]
  row 28 cols 91-100: "          " [fg] -> "supervisor" [fg]
  row 28 cols 102-105: "    " [fg] -> "uses" [fg]
  row 28 cols 107-109: "   " [fg] -> "the" [fg]
  row 28 cols 111-113: "   " [fg] -> "old" [fg]
  row 29 cols 57-61: "     " [fg] -> "code." [fg]
  row 30 cols 54-55: "  " [fg] -> "No" [fg]
  row 30 cols 57-61: "     " [fg] -> "skill" [fg]
  row 30 cols 63-66: "    " [fg] -> "told" [fg]
  row 30 cols 68-69: "  " [fg] -> "an" [fg]
  row 30 cols 71-75: "     " [fg] -> "agent" [fg]
  row 30 cols 77-78: "  " [fg] -> "to" [fg]
  row 30 cols 80-86: "       " [fg] -> "restart" [fg]
  row 30 cols 88-90: "   " [fg] -> "the" [fg]
  row 30 cols 92-101: "          " [fg] -> "supervisor" [fg]
  row 30 cols 103-107: "     " [fg] -> "after" [fg]
  row 30 cols 109-109: " " [fg] -> "a" [fg]
  row 31 cols 54-61: "        " [fg] -> "rebuild." [fg]
  row 39 cols 55-76: "/ filter · g/G top/end" [dim] -> "g/G top/end · ? help  " [dim]
  row 39 cols 78-78: "·" [dim] -> " " [dim]
  row 39 cols 80-80: "?" [dim] -> " " [dim]
  row 39 cols 82-85: "help" [dim] -> "    " [dim]
Health.grid: 16 cells
  row 6 cols 2-2: " " [accent+b] -> " " [fg]
  row 7 cols 2-2: " " [accent+b] -> " " [fg]
  row 8 cols 2-2: " " [accent+b] -> " " [fg]
  row 9 cols 2-2: " " [accent+b] -> " " [fg]
  row 10 cols 2-2: " " [accent+b] -> " " [fg]
  row 12 cols 2-2: " " [accent+b] -> " " [fg]
  row 13 cols 2-2: " " [accent+b] -> " " [fg]
  row 14 cols 2-2: " " [accent+b] -> " " [fg]
  row 15 cols 2-2: " " [accent+b] -> " " [fg]
  row 16 cols 2-2: " " [accent+b] -> " " [fg]
  row 17 cols 2-2: " " [accent+b] -> " " [fg]
  row 18 cols 2-2: " " [accent+b] -> " " [fg]
  row 19 cols 2-2: " " [accent+b] -> " " [fg]
  row 20 cols 2-2: " " [accent+b] -> " " [fg]
  row 21 cols 2-2: " " [accent+b] -> " " [fg]
  row 22 cols 2-2: " " [accent+b] -> " " [fg]
HealthAgents.grid: 5 cells
  row 6 cols 2-2: " " [accent+b] -> " " [fg]
  row 8 cols 2-2: " " [accent+b] -> " " [fg]
  row 9 cols 2-2: " " [accent+b] -> " " [fg]
  row 10 cols 2-2: " " [accent+b] -> " " [fg]
  row 10 cols 50-50: " " [dim] -> " " [fg]
Help.grid: 59 cells
  row 3 cols 18-29: "────────────" [border] -> "· 1 project " [border]
  row 37 cols 2-3: "  " [border] -> "──" [border]
  row 37 cols 5-9: "     " [border] -> "spawn" [border]
  row 37 cols 13-13: " " [border] -> "●" [border]
  row 37 cols 15-21: "       " [border] -> "working" [border]
  row 37 cols 25-25: " " [border] -> "◐" [border]
  row 37 cols 27-30: "    " [border] -> "idle" [border]
  row 37 cols 34-34: " " [border] -> "○" [border]
  row 37 cols 36-42: "       " [border] -> "stopped" [border]
  row 37 cols 46-48: "   " [border] -> "one" [border]
  row 37 cols 50-55: "      " [border] -> "column" [border]
  row 37 cols 57-59: "   " [border] -> "per" [border]
  row 37 cols 61-67: "       " [border] -> "project" [border]
Logs.grid: 1256 cells
  row 8 cols 2-2: " " [accent+b] -> " " [fg]
  row 8 cols 7-8: "19" [dim] -> "04" [dim]
  row 8 cols 10-11: "40" [dim] -> "36" [dim]
  row 8 cols 19-23: "inbox" [info] -> "herdr" [info]
  row 8 cols 29-54: "record sent to coordinator" [fg] -> "pane.read w58:p6 · 40ms   " [fg]
  row 8 cols 119-119: "┃" [accent] -> "│" [accent]
  row 9 cols 2-2: " " [accent+b] -> " " [fg]
  row 9 cols 7-8: "19" [dim] -> "05" [dim]
  row 9 cols 10-11: "40" [dim] -> "24" [dim]
  row 9 cols 19-27: "supervise" [info] -> "task     " [info]
  row 9 cols 29-38: "nudge coor" [fg] -> "status rea" [fg]
  row 9 cols 40-61: "inator · inbox 2      " [fg] -> " 4-agents-equal-colum…" [fg]
  row 9 cols 119-119: "┃" [accent] -> "│" [accent]
  row 10 cols 2-2: " " [accent+b] -> " " [fg]
  row 10 cols 7-8: "21" [dim] -> "06" [dim]
  row 10 cols 10-10: "0" [dim] -> "1" [dim]
  row 10 cols 29-41: "spawn worker " [fg] -> "list · 5 agen" [fg]
  row 10 cols 43-52: "2-worker-1" [fg] -> "s         " [fg]
  row 10 cols 119-119: "┃" [accent] -> "│" [accent]
  row 11 cols 2-2: " " [accent+b] -> " " [fg]
  row 11 cols 7-8: "21" [dim] -> "07" [dim]
  row 11 cols 11-11: "9" [dim] -> "0" [dim]
  row 11 cols 19-23: "herdr" [info] -> "inbox" [info]
  row 11 cols 29-39: "agent.start" [fg] -> "record read" [fg]
  row 11 cols 41-55: "w5T:p1 · 812ms " [fg] -> "by manager-hive" [fg]
  row 11 cols 119-119: "┃" [accent] -> "│" [accent]
  row 12 cols 2-2: " " [accent+b] -> " " [fg]
  row 12 cols 7-8: "24" [dim] -> "07" [dim]
  row 12 cols 10-11: "51" [dim] -> "48" [dim]
  row 12 cols 19-27: "task     " [info] -> "supervise" [info]
  row 12 cols 30-30: "t" [fg] -> "c" [fg]
  row 12 cols 32-43: "tus assigned" [fg] -> "n agents · 5" [fg]
  row 12 cols 45-47: "→ w" [fg] -> "rec" [fg]
  row 12 cols 50-53: "king" [fg] -> "ds  " [fg]
  row 12 cols 119-119: "┃" [accent] -> "│" [accent]
  row 13 cols 2-2: " " [accent+b] -> " " [fg]
  row 13 cols 7-8: "33" [dim] -> "08" [dim]
  row 13 cols 10-11: "17" [dim] -> "36" [dim]
  row 13 cols 19-23: "task " [info] -> "herdr" [info]
  row 13 cols 29-44: "status working →" [fg] -> "pane.read w58:p6" [fg]
  row 13 cols 46-54: "completed" [fg] -> "· 40ms   " [fg]
  row 13 cols 119-119: "┃" [accent] -> "│" [accent]
  row 14 cols 2-2: " " [accent+b] -> " " [fg]
  row 14 cols 7-8: "33" [dim] -> "09" [dim]
  row 14 cols 10-11: "18" [dim] -> "24" [dim]
  row 14 cols 19-23: "agent" [info] -> "task " [info]
  row 14 cols 31-53: "op t2-worker-1 · pane c" [fg] -> "atus read 4-agents-equa" [fg]
  row 14 cols 55-61: "osed   " [fg] -> "-colum…" [fg]
  row 14 cols 119-119: "┃" [accent] -> "│" [accent]
  row 15 cols 2-2: " " [accent+b] -> " " [fg]
  row 15 cols 7-8: "33" [dim] -> "10" [dim]
  row 15 cols 10-11: "40" [dim] -> "12" [dim]
  row 15 cols 29-56: "spawn verifier t2-verifier-1" [fg] -> "list · 5 agents             " [fg]
  row 15 cols 119-119: "┃" [accent] -> "│" [accent]
  row 16 cols 2-2: "▶" [accent+b] -> " " [fg]
  row 16 cols 4-11: "13:38:02" [accent] -> "13:11:00" [dim]
  row 16 cols 13-23: "WARN  task " [accent+b,info] -> "DEBUG inbox" [dim,info]
  row 16 cols 29-61: "verify fail · V4 diff scope      " [accent+b] -> "record read by manager-hive      " [fg]
  row 16 cols 119-119: "┃" [accent] -> "│" [accent]
  row 17 cols 2-2: " " [accent+b] -> " " [fg]
  row 17 cols 7-8: "38" [dim] -> "11" [dim]
  row 17 cols 10-11: "30" [dim] -> "48" [dim]
  row 17 cols 19-27: "agent    " [info] -> "supervise" [info]
  row 17 cols 30-30: "p" [fg] -> "c" [fg]
  row 17 cols 32-52: "wn worker t2-worker-2" [fg] -> "n agents · 5 records " [fg]
  row 17 cols 119-119: "┃" [accent] -> "│" [accent]
  row 18 cols 2-2: " " [accent+b] -> " " [fg]
  row 18 cols 7-8: "44" [dim] -> "12" [dim]
  row 18 cols 10-11: "12" [dim] -> "36" [dim]
  row 18 cols 19-23: "task " [info] -> "herdr" [info]
  row 18 cols 29-44: "status completed" [fg] -> "pane.read w58:p6" [fg]
  row 18 cols 46-46: "→" [fg] -> "·" [fg]
  row 18 cols 48-55: "verified" [fg] -> "40ms    " [fg]
  row 18 cols 119-119: "┃" [accent] -> "│" [accent]
  row 19 cols 2-2: " " [accent+b] -> " " [fg]
  row 19 cols 7-8: "44" [dim] -> "13" [dim]
  row 19 cols 11-11: "0" [dim] -> "4" [dim]
  row 19 cols 19-19: "l" [info] -> "t" [info]
  row 19 cols 21-22: "nd" [info] -> "sk" [info]
  row 19 cols 29-61: "merge task/2 into dev · 0e079272 " [fg] -> "status read 4-agents-equal-colum…" [fg]
  row 19 cols 119-119: "┃" [accent] -> "│" [accent]
  row 20 cols 2-2: " " [accent+b] -> " " [fg]
  row 20 cols 7-7: "4" [dim] -> "1" [dim]
  row 20 cols 10-10: "2" [dim] -> "1" [dim]
  row 20 cols 19-23: "land " [info] -> "agent" [info]
  row 20 cols 29-30: "pu" [fg] -> "li" [fg]
  row 20 cols 32-32: "h" [fg] -> "t" [fg]
  row 20 cols 34-36: "dev" [fg] -> "· 5" [fg]
  row 20 cols 38-42: "· 1.9" [fg] -> "agent" [fg]
  row 20 cols 119-119: "┃" [accent] -> "│" [accent]
  row 21 cols 2-2: " " [accent+b] -> " " [fg]
  row 21 cols 7-8: "44" [dim] -> "15" [dim]
  row 21 cols 10-11: "23" [dim] -> "00" [dim]
  row 21 cols 19-26: "worktree" [info] -> "inbox   " [info]
  row 21 cols 31-31: "m" [fg] -> "c" [fg]
  row 21 cols 33-34: "ve" [fg] -> "rd" [fg]
  row 21 cols 36-61: "2-stale-coordinator-start…" [fg] -> "read by manager-hive      " [fg]
  row 21 cols 119-119: "┃" [accent] -> "│" [accent]
  row 22 cols 2-2: " " [accent+b] -> " " [fg]
  row 22 cols 7-8: "44" [dim] -> "19" [dim]
  row 22 cols 10-11: "25" [dim] -> "40" [dim]
  row 22 cols 119-119: "┃" [accent] -> "│" [accent]
  row 23 cols 2-2: " " [accent+b] -> " " [fg]
  row 23 cols 7-8: "45" [dim] -> "19" [dim]
  row 23 cols 10-11: "01" [dim] -> "40" [dim]
  row 23 cols 19-27: "agent    " [info] -> "supervise" [info]
  row 23 cols 29-33: "spawn" [fg] -> "nudge" [fg]
  row 23 cols 35-35: "w" [fg] -> "c" [fg]
  row 23 cols 37-53: "rker t3-worker-1 " [fg] -> "ordinator · inbox" [fg]
  row 23 cols 55-55: " " [fg] -> "2" [fg]
  row 23 cols 119-119: "┃" [accent] -> "│" [accent]
  row 24 cols 4-11: "        " [fg] -> "13:21:02" [dim]
  row 24 cols 13-33: "                     " [fg] -> "DEBUG agent     spawn" [dim,info,fg]
  row 24 cols 35-40: "      " [fg] -> "worker" [fg]
  row 24 cols 42-52: "           " [fg] -> "t2-worker-1" [fg]
  row 24 cols 119-119: "┃" [accent] -> "│" [accent]
  row 25 cols 4-11: "        " [fg] -> "13:21:09" [dim]
  row 25 cols 13-39: "                           " [fg] -> "DEBUG herdr     agent.start" [dim,info,fg]
  row 25 cols 41-46: "      " [fg] -> "w5T:p1" [fg]
  row 25 cols 48-48: " " [fg] -> "·" [fg]
  row 25 cols 50-54: "     " [fg] -> "812ms" [fg]
  row 25 cols 119-119: "┃" [accent] -> "│" [accent]
  row 26 cols 4-11: "        " [fg] -> "13:24:51" [dim]
  row 26 cols 13-34: "                      " [fg] -> "DEBUG task      status" [dim,info,fg]
  row 26 cols 36-43: "        " [fg] -> "assigned" [fg]
  row 26 cols 45-45: " " [fg] -> "→" [fg]
  row 26 cols 47-53: "       " [fg] -> "working" [fg]
  row 26 cols 119-119: "┃" [accent] -> "│" [accent]
  row 27 cols 4-11: "        " [fg] -> "13:33:17" [dim]
  row 27 cols 13-34: "                      " [fg] -> "DEBUG task      status" [dim,info,fg]
  row 27 cols 36-42: "       " [fg] -> "working" [fg]
  row 27 cols 44-44: " " [fg] -> "→" [fg]
  row 27 cols 46-54: "         " [fg] -> "completed" [fg]
  row 27 cols 119-119: "┃" [accent] -> "│" [accent]
  row 28 cols 4-11: "        " [fg] -> "13:33:18" [dim]
  row 28 cols 13-32: "                    " [fg] -> "DEBUG agent     stop" [dim,info,fg]
  row 28 cols 34-44: "           " [fg] -> "t2-worker-1" [fg]
  row 28 cols 46-46: " " [fg] -> "·" [fg]
  row 28 cols 48-51: "    " [fg] -> "pane" [fg]
  row 28 cols 53-58: "      " [fg] -> "closed" [fg]
  row 28 cols 119-119: "┃" [accent] -> "│" [accent]
  row 29 cols 4-11: "        " [fg] -> "13:33:40" [dim]
  row 29 cols 13-33: "                     " [fg] -> "DEBUG agent     spawn" [dim,info,fg]
  row 29 cols 35-42: "        " [fg] -> "verifier" [fg]
  row 29 cols 44-56: "             " [fg] -> "t2-verifier-1" [fg]
  row 30 cols 2-2: " " [fg] -> "▶" [accent+b]
  row 30 cols 4-11: "        " [fg] -> "13:38:02" [accent]
  row 30 cols 13-61: "                                                 " [fg] -> "WARN  task      verify fail · V4 diff scope      " [accent+b,info]
  row 30 cols 63-63: "┃" [dim] -> "│" [border]
  row 31 cols 4-11: "        " [fg] -> "13:38:30" [dim]
  row 31 cols 13-33: "                     " [fg] -> "DEBUG agent     spawn" [dim,info,fg]
  row 31 cols 35-40: "      " [fg] -> "worker" [fg]
  row 31 cols 42-52: "           " [fg] -> "t2-worker-2" [fg]
  row 31 cols 63-63: "┃" [dim] -> "│" [border]
  row 32 cols 4-11: "        " [fg] -> "13:44:12" [dim]
  row 32 cols 13-34: "                      " [fg] -> "DEBUG task      status" [dim,info,fg]
  row 32 cols 36-44: "         " [fg] -> "completed" [fg]
  row 32 cols 46-46: " " [fg] -> "→" [fg]
  row 32 cols 48-55: "        " [fg] -> "verified" [fg]
  row 32 cols 63-63: "┃" [dim] -> "│" [border]
  row 33 cols 4-11: "        " [fg] -> "13:44:20" [dim]
  row 33 cols 13-33: "                     " [fg] -> "DEBUG land      merge" [dim,info,fg]
  row 33 cols 35-40: "      " [fg] -> "task/2" [fg]
  row 33 cols 42-45: "    " [fg] -> "into" [fg]
  row 33 cols 47-49: "   " [fg] -> "dev" [fg]
  row 33 cols 51-51: " " [fg] -> "·" [fg]
  row 33 cols 53-60: "        " [fg] -> "0e079272" [fg]
  row 33 cols 63-63: "┃" [dim] -> "│" [border]
  row 34 cols 4-11: "        " [fg] -> "13:44:22" [dim]
  row 34 cols 13-32: "                    " [fg] -> "DEBUG land      push" [dim,info,fg]
  row 34 cols 34-36: "   " [fg] -> "dev" [fg]
  row 34 cols 38-38: " " [fg] -> "·" [fg]
  row 34 cols 40-43: "    " [fg] -> "1.9s" [fg]
  row 34 cols 63-63: "┃" [dim] -> "│" [border]
  row 35 cols 4-11: "        " [fg] -> "13:44:23" [dim]
  row 35 cols 13-34: "                      " [fg] -> "DEBUG worktree  remove" [dim,info,fg]
  row 35 cols 36-61: "                          " [fg] -> "2-stale-coordinator-start…" [fg]
  row 35 cols 63-63: "┃" [dim] -> "│" [border]
  row 36 cols 4-11: "        " [fg] -> "13:44:25" [dim]
  row 36 cols 13-34: "                      " [fg] -> "DEBUG inbox     record" [dim,info,fg]
  row 36 cols 36-39: "    " [fg] -> "sent" [fg]
  row 36 cols 41-42: "  " [fg] -> "to" [fg]
  row 36 cols 44-54: "           " [fg] -> "coordinator" [fg]
  row 37 cols 4-11: "        " [fg] -> "13:45:01" [dim]
  row 37 cols 13-33: "                     " [fg] -> "DEBUG agent     spawn" [dim,info,fg]
  row 37 cols 35-40: "      " [fg] -> "worker" [fg]
  row 37 cols 42-52: "           " [fg] -> "t3-worker-1" [fg]
  row 37 cols 63-63: "│" [border] -> "┃" [dim]
Main.grid: 238 cells
  row 3 cols 18-29: "────────────" [accent] -> "· 1 project " [accent+b]
  row 5 cols 36-59: "· g1-tui-restart · done " [dim+b] -> "────────────────────────" [border]
  row 6 cols 30-50: "                     " [fg] -> "g1-tui-restart · done" [dim]
  row 8 cols 64-69: "      " [accent+b] -> "      " [fg]
  row 9 cols 66-69: "nud…" [dim] -> "4m  " [dim,fg]
  row 12 cols 45-46: " ╭" [fg,border] -> "╭─" [border]
  row 12 cols 75-76: "─╮" [border] -> "╮ " [border,fg]
  row 13 cols 45-46: " │" [fg,border] -> "│ " [border,fg]
  row 13 cols 75-76: " │" [fg,border] -> "│ " [border,fg]
  row 14 cols 33-34: " ╭" [fg,border] -> "╭─" [border]
  row 14 cols 45-46: "─┴" [border] -> "┴─" [border]
  row 14 cols 56-57: "─╮" [border] -> "╮ " [border,fg]
  row 14 cols 63-64: " ╭" [fg,border] -> "╭─" [border]
  row 14 cols 75-76: "─┴" [border] -> "┴─" [border]
  row 14 cols 86-87: "─╮" [border] -> "╮ " [border,fg]
  row 15 cols 33-57: " │ ○ t4-worker-1        │" [fg,border,dim] -> "│ ○ t4-worker-1        │ " [border,fg,dim]
  row 15 cols 63-87: " │ ○ t4-verifier-1      │" [fg,border,dim] -> "│ ○ t4-verifier-1      │ " [border,fg,dim]
  row 16 cols 33-43: " │ pi · sto" [fg,border,dim] -> "│ pi · stop" [border,fg,dim]
  row 16 cols 45-57: "ped · task… │" [dim,fg,border] -> "ed · done  │ " [dim,fg,border]
  row 16 cols 63-73: " │ pi · sto" [fg,border,dim] -> "│ pi · stop" [border,fg,dim]
  row 16 cols 75-82: "ped · pa" [dim] -> "ed · pas" [dim]
  row 16 cols 84-87: "s  │" [dim,fg,border] -> "  │ " [fg,border]
  row 17 cols 33-34: " ╰" [fg,border] -> "╰─" [border]
  row 17 cols 56-57: "─╯" [border] -> "╯ " [border,fg]
  row 17 cols 63-64: " ╰" [fg,border] -> "╰─" [border]
  row 17 cols 86-87: "─╯" [border] -> "╯ " [border,fg]
  row 19 cols 30-30: "2" [dim] -> "h" [dim]
  row 19 cols 32-38: "stopped" [dim] -> "hides 2" [dim]
  row 19 cols 41-41: "h" [dim] -> "t" [dim]
  row 19 cols 43-59: "wn · h hides them" [dim] -> "pped             " [dim,fg]
  row 37 cols 68-87: ", all the same width" [dim] -> "                    " [fg]
TooSmall.grid: 0 cells
Work.grid: 26 cells
  row 6 cols 2-3: "  " [accent+b] -> "  " [fg]
  row 7 cols 3-3: " " [accent+b] -> " " [fg]
  row 7 cols 29-29: " " [fg] -> " " [accent+b]
  row 8 cols 2-3: "  " [accent+b] -> "  " [fg]
  row 8 cols 28-29: "… " [fg] -> "t…" [fg]
  row 9 cols 2-3: "  " [accent+b] -> "  " [fg]
  row 27 cols 94-96: "pas" [dim] -> " pa" [dim]
  row 27 cols 98-98: " " [fg] -> "s" [dim]
  row 39 cols 87-87: "/" [dim] -> "?" [dim]
  row 39 cols 89-90: "fi" [dim] -> "he" [dim]
  row 39 cols 92-94: "ter" [dim] -> "p  " [dim]
  row 39 cols 96-96: "·" [dim] -> " " [dim]
  row 39 cols 98-98: "?" [dim] -> " " [dim]
  row 39 cols 100-103: "help" [dim] -> "    " [dim]
WorkGoals.grid: 13 cells
  row 6 cols 2-2: " " [accent+b] -> " " [fg]
  row 39 cols 75-75: "/" [dim] -> "?" [dim]
  row 39 cols 77-78: "fi" [dim] -> "he" [dim]
  row 39 cols 80-82: "ter" [dim] -> "p  " [dim]
  row 39 cols 84-84: "·" [dim] -> " " [dim]
  row 39 cols 86-86: "?" [dim] -> " " [dim]
  row 39 cols 88-91: "help" [dim] -> "    " [dim]
WorkQueue.grid: 14 cells
  row 8 cols 2-3: "  " [border] -> "  " [fg]
  row 39 cols 79-79: "/" [dim] -> "?" [dim]
  row 39 cols 81-82: "fi" [dim] -> "he" [dim]
  row 39 cols 84-86: "ter" [dim] -> "p  " [dim]
  row 39 cols 88-88: "·" [dim] -> " " [dim]
  row 39 cols 90-90: "?" [dim] -> " " [dim]
  row 39 cols 92-95: "help" [dim] -> "    " [dim]
total: 4575 cells
```

## V3 rerun output

Commands: `go test ./internal/tui/component/ ./internal/tui/ -run 'TestListRowCells|TestListGolden|TestDetailsShowsMoreLinesOnOverflow|TestListHintsHaveNoFilterOnWorkAndFriction|TestFrameRowsMatchEachScreenBoard' -count=1`.

Deliberate breaks: marker space style; selected name final space style; unselected marker spaces; suppressed overflow hint; incorrect Thumb visible count; Friction filter restored; Logs filter removed. Each file was restored immediately after its run.

```text

### dev implementations, task tests and grids
--- FAIL: TestDetailsShowsMoreLinesOnOverflow (0.00s)
    details_test.go:49: top:
        ╭─ x ──────────────╮
        │ l0               ┃
        │ l1               │
        │ l2               │
        ╰──────────────────╯
        want:
        ╭─ x ──────────────╮
        │ l0               ┃
        │ l1               │
        │ ↓ 8 more lines   │
        ╰──────────────────╯
    details_test.go:54: offset 3:
        ╭─ x ──────────────╮
        │ l3               ┃
        │ l4               │
        │ l5               │
        ╰──────────────────╯
        want:
        ╭─ x ──────────────╮
        │ l3               ┃
        │ l4               │
        │ ↓ 5 more lines   │
        ╰──────────────────╯
    details_test.go:60: one line over (Thumb(3, 4, 2, 0) is one cell):
        ╭─ x ──────────────╮
        │ l0               ┃
        │ l1               ┃
        │ l2               │
        ╰──────────────────╯
        want:
        ╭─ x ──────────────╮
        │ l0               ┃
        │ l1               │
        │ ↓ 2 more lines   │
        ╰──────────────────╯
--- FAIL: TestListGolden (0.00s)
    list_test.go:36: 2 cells differ from Components.grid (board row and column):
        row 36 col 3: want ' ' fg normal, got ' ' accent bold
        row 36 col 29: want ' ' accent bold, got ' ' fg normal
        
        rendered:
        ╭─ tasks · 4 ──────────────────────────────────────────────╮
        │   ID                        STATUS    WAVE               │
        │   4-agents-equal-columns    landed    C                  │
        │ ▶ 3-tui-restart             landed    B                  │
        │   2-stale-coordinator-star… landed    A                  │
        ╰──────────────────────────────────────────────────────────╯
--- FAIL: TestListRowCells (0.00s)
    list_test.go:107: row 0 col 3 = ' ' accent bold true; want ' ' fg bold false
    list_test.go:107: row 0 col 9 = ' ' fg bold false; want ' ' accent bold true
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui/component	0.017s
--- FAIL: TestFrameRowsMatchEachScreenBoard (0.05s)
    --- FAIL: TestFrameRowsMatchEachScreenBoard/CompactFriction (0.00s)
        frame_test.go:116: 9 frame cells differ from CompactFriction.grid:
            row 23 col 45: want '?' dim normal, got '/' dim normal
            row 23 col 47: want ' ' dim normal, got 'f' dim normal
            row 23 col 48: want ' ' dim normal, got 'i' dim normal
            row 23 col 49: want ' ' dim normal, got 'l' dim normal
            row 23 col 50: want ' ' dim normal, got 't' dim normal
            row 23 col 51: want ' ' dim normal, got 'e' dim normal
            row 23 col 52: want ' ' dim normal, got 'r' dim normal
            row 23 col 54: want ' ' dim normal, got '·' dim normal
            row 23 col 56: want ' ' dim normal, got '?' dim normal
    --- FAIL: TestFrameRowsMatchEachScreenBoard/Friction (0.00s)
        frame_test.go:116: 28 frame cells differ from Friction.grid:
            row 39 col 55: want 'g' dim normal, got '/' dim normal
            row 39 col 56: want '/' dim normal, got ' ' dim normal
            row 39 col 57: want 'G' dim normal, got 'f' dim normal
            row 39 col 58: want ' ' dim normal, got 'i' dim normal
            row 39 col 59: want 't' dim normal, got 'l' dim normal
            row 39 col 60: want 'o' dim normal, got 't' dim normal
            row 39 col 61: want 'p' dim normal, got 'e' dim normal
            row 39 col 62: want '/' dim normal, got 'r' dim normal
            row 39 col 63: want 'e' dim normal, got ' ' dim normal
            row 39 col 64: want 'n' dim normal, got '·' dim normal
            row 39 col 65: want 'd' dim normal, got ' ' dim normal
            row 39 col 66: want ' ' dim normal, got 'g' dim normal
            row 39 col 67: want '·' dim normal, got '/' dim normal
            row 39 col 68: want ' ' dim normal, got 'G' dim normal
            row 39 col 69: want '?' dim normal, got ' ' dim normal
            row 39 col 70: want ' ' dim normal, got 't' dim normal
            row 39 col 71: want 'h' dim normal, got 'o' dim normal
            row 39 col 72: want 'e' dim normal, got 'p' dim normal
            row 39 col 73: want 'l' dim normal, got '/' dim normal
            row 39 col 74: want 'p' dim normal, got 'e' dim normal
            row 39 col 75: want ' ' dim normal, got 'n' dim normal
            row 39 col 76: want ' ' dim normal, got 'd' dim normal
            row 39 col 78: want ' ' dim normal, got '·' dim normal
            row 39 col 80: want ' ' dim normal, got '?' dim normal
            row 39 col 82: want ' ' dim normal, got 'h' dim normal
            row 39 col 83: want ' ' dim normal, got 'e' dim normal
            row 39 col 84: want ' ' dim normal, got 'l' dim normal
            row 39 col 85: want ' ' dim normal, got 'p' dim normal
    --- FAIL: TestFrameRowsMatchEachScreenBoard/FrictionResolved (0.00s)
        frame_test.go:116: 28 frame cells differ from FrictionResolved.grid:
            row 39 col 55: want 'g' dim normal, got '/' dim normal
            row 39 col 56: want '/' dim normal, got ' ' dim normal
            row 39 col 57: want 'G' dim normal, got 'f' dim normal
            row 39 col 58: want ' ' dim normal, got 'i' dim normal
            row 39 col 59: want 't' dim normal, got 'l' dim normal
            row 39 col 60: want 'o' dim normal, got 't' dim normal
            row 39 col 61: want 'p' dim normal, got 'e' dim normal
            row 39 col 62: want '/' dim normal, got 'r' dim normal
            row 39 col 63: want 'e' dim normal, got ' ' dim normal
            row 39 col 64: want 'n' dim normal, got '·' dim normal
            row 39 col 65: want 'd' dim normal, got ' ' dim normal
            row 39 col 66: want ' ' dim normal, got 'g' dim normal
            row 39 col 67: want '·' dim normal, got '/' dim normal
            row 39 col 68: want ' ' dim normal, got 'G' dim normal
            row 39 col 69: want '?' dim normal, got ' ' dim normal
            row 39 col 70: want ' ' dim normal, got 't' dim normal
            row 39 col 71: want 'h' dim normal, got 'o' dim normal
            row 39 col 72: want 'e' dim normal, got 'p' dim normal
            row 39 col 73: want 'l' dim normal, got '/' dim normal
            row 39 col 74: want 'p' dim normal, got 'e' dim normal
            row 39 col 75: want ' ' dim normal, got 'n' dim normal
            row 39 col 76: want ' ' dim normal, got 'd' dim normal
            row 39 col 78: want ' ' dim normal, got '·' dim normal
            row 39 col 80: want ' ' dim normal, got '?' dim normal
            row 39 col 82: want ' ' dim normal, got 'h' dim normal
            row 39 col 83: want ' ' dim normal, got 'e' dim normal
            row 39 col 84: want ' ' dim normal, got 'l' dim normal
            row 39 col 85: want ' ' dim normal, got 'p' dim normal
    --- FAIL: TestFrameRowsMatchEachScreenBoard/Work (0.00s)
        frame_test.go:116: 12 frame cells differ from Work.grid:
            row 39 col 87: want '?' dim normal, got '/' dim normal
            row 39 col 89: want 'h' dim normal, got 'f' dim normal
            row 39 col 90: want 'e' dim normal, got 'i' dim normal
            row 39 col 92: want 'p' dim normal, got 't' dim normal
            row 39 col 93: want ' ' dim normal, got 'e' dim normal
            row 39 col 94: want ' ' dim normal, got 'r' dim normal
            row 39 col 96: want ' ' dim normal, got '·' dim normal
            row 39 col 98: want ' ' dim normal, got '?' dim normal
            row 39 col 100: want ' ' dim normal, got 'h' dim normal
            row 39 col 101: want ' ' dim normal, got 'e' dim normal
            row 39 col 102: want ' ' dim normal, got 'l' dim normal
            row 39 col 103: want ' ' dim normal, got 'p' dim normal
    --- FAIL: TestFrameRowsMatchEachScreenBoard/WorkGoals (0.00s)
        frame_test.go:116: 12 frame cells differ from WorkGoals.grid:
            row 39 col 75: want '?' dim normal, got '/' dim normal
            row 39 col 77: want 'h' dim normal, got 'f' dim normal
            row 39 col 78: want 'e' dim normal, got 'i' dim normal
            row 39 col 80: want 'p' dim normal, got 't' dim normal
            row 39 col 81: want ' ' dim normal, got 'e' dim normal
            row 39 col 82: want ' ' dim normal, got 'r' dim normal
            row 39 col 84: want ' ' dim normal, got '·' dim normal
            row 39 col 86: want ' ' dim normal, got '?' dim normal
            row 39 col 88: want ' ' dim normal, got 'h' dim normal
            row 39 col 89: want ' ' dim normal, got 'e' dim normal
            row 39 col 90: want ' ' dim normal, got 'l' dim normal
            row 39 col 91: want ' ' dim normal, got 'p' dim normal
    --- FAIL: TestFrameRowsMatchEachScreenBoard/WorkQueue (0.00s)
        frame_test.go:116: 12 frame cells differ from WorkQueue.grid:
            row 39 col 79: want '?' dim normal, got '/' dim normal
            row 39 col 81: want 'h' dim normal, got 'f' dim normal
            row 39 col 82: want 'e' dim normal, got 'i' dim normal
            row 39 col 84: want 'p' dim normal, got 't' dim normal
            row 39 col 85: want ' ' dim normal, got 'e' dim normal
            row 39 col 86: want ' ' dim normal, got 'r' dim normal
            row 39 col 88: want ' ' dim normal, got '·' dim normal
            row 39 col 90: want ' ' dim normal, got '?' dim normal
            row 39 col 92: want ' ' dim normal, got 'h' dim normal
            row 39 col 93: want ' ' dim normal, got 'e' dim normal
            row 39 col 94: want ' ' dim normal, got 'l' dim normal
            row 39 col 95: want ' ' dim normal, got 'p' dim normal
--- FAIL: TestListHintsHaveNoFilterOnWorkAndFriction (0.00s)
    frame_test.go:440: Work "" list, wide false: hints ["↑↓ move" "enter details" "[ ] view" "/ filter" "?"] have / filter
    frame_test.go:440: Work "Goals" list, wide true: hints ["↑↓ move" "tab region" "[ ] view" "enter open" "o design" "t task" "/ filter" "? help"] have / filter
    frame_test.go:440: Work "Goals" list, wide false: hints ["↑↓ move" "enter details" "[ ] view" "/ filter" "?"] have / filter
    frame_test.go:440: Work "Queue" list, wide true: hints ["↑↓ move" "tab region" "[ ] view" "enter open" "s stops only" "t task" "/ filter" "? help"] have / filter
    frame_test.go:440: Work "Queue" list, wide false: hints ["↑↓ move" "enter details" "[ ] view" "/ filter" "?"] have / filter
    frame_test.go:440: Work "Tasks" list, wide true: hints ["↑↓ move" "tab region" "[ ] view" "enter open" "b brief" "v verified" "d deps" "/ filter" "? help"] have / filter
    frame_test.go:440: Work "Tasks" list, wide false: hints ["↑↓ move" "enter details" "[ ] view" "/ filter" "?"] have / filter
    frame_test.go:440: Friction "" list, wide true: hints ["↑↓ move" "tab region" "[ ] view" "enter open" "/ filter" "g/G top/end" "? help"] have / filter
    frame_test.go:440: Friction "" list, wide false: hints ["↑↓ move" "enter details" "[ ] view" "/ filter" "?"] have / filter
    frame_test.go:440: Friction "Open" list, wide true: hints ["↑↓ move" "tab region" "[ ] view" "enter open" "/ filter" "g/G top/end" "? help"] have / filter
    frame_test.go:440: Friction "Open" list, wide false: hints ["↑↓ move" "enter details" "[ ] view" "/ filter" "?"] have / filter
    frame_test.go:440: Friction "Resolved" list, wide true: hints ["↑↓ move" "tab region" "[ ] view" "enter open" "/ filter" "g/G top/end" "? help"] have / filter
    frame_test.go:440: Friction "Resolved" list, wide false: hints ["↑↓ move" "enter details" "[ ] view" "/ filter" "?"] have / filter
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui	0.067s
FAIL
exit 1

### task green
ok  	github.com/Harrison-Blair/hive/internal/tui/component	0.018s
ok  	github.com/Harrison-Blair/hive/internal/tui	0.066s
exit 0

### deliberate break 1
--- FAIL: TestListGolden (0.00s)
    list_test.go:36: 1 cells differ from Components.grid (board row and column):
        row 36 col 3: want ' ' fg normal, got ' ' accent bold
        
        rendered:
        ╭─ tasks · 4 ──────────────────────────────────────────────╮
        │   ID                        STATUS    WAVE               │
        │   4-agents-equal-columns    landed    C                  │
        │ ▶ 3-tui-restart             landed    B                  │
        │   2-stale-coordinator-star… landed    A                  │
        ╰──────────────────────────────────────────────────────────╯
--- FAIL: TestListRowCells (0.00s)
    list_test.go:107: row 0 col 3 = ' ' accent bold true; want ' ' fg bold false
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui/component	0.017s
ok  	github.com/Harrison-Blair/hive/internal/tui	0.064s
FAIL
exit 1

### deliberate break 2
--- FAIL: TestListGolden (0.00s)
    list_test.go:36: 1 cells differ from Components.grid (board row and column):
        row 36 col 29: want ' ' accent bold, got ' ' fg normal
        
        rendered:
        ╭─ tasks · 4 ──────────────────────────────────────────────╮
        │   ID                        STATUS    WAVE               │
        │   4-agents-equal-columns    landed    C                  │
        │ ▶ 3-tui-restart             landed    B                  │
        │   2-stale-coordinator-star… landed    A                  │
        ╰──────────────────────────────────────────────────────────╯
--- FAIL: TestListRowCells (0.00s)
    list_test.go:107: row 0 col 9 = ' ' fg bold false; want ' ' accent bold true
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui/component	0.017s
ok  	github.com/Harrison-Blair/hive/internal/tui	0.066s
FAIL
exit 1

### deliberate break 3
--- FAIL: TestListGolden (0.00s)
    list_test.go:36: 4 cells differ from Components.grid (board row and column):
        row 35 col 2: want ' ' fg normal, got ' ' accent bold
        row 35 col 3: want ' ' fg normal, got ' ' accent bold
        row 37 col 2: want ' ' fg normal, got ' ' accent bold
        row 37 col 3: want ' ' fg normal, got ' ' accent bold
        
        rendered:
        ╭─ tasks · 4 ──────────────────────────────────────────────╮
        │   ID                        STATUS    WAVE               │
        │   4-agents-equal-columns    landed    C                  │
        │ ▶ 3-tui-restart             landed    B                  │
        │   2-stale-coordinator-star… landed    A                  │
        ╰──────────────────────────────────────────────────────────╯
--- FAIL: TestListRowCells (0.00s)
    list_test.go:107: row 1 col 2 = ' ' accent bold true; want ' ' fg bold false
    list_test.go:107: row 1 col 3 = ' ' accent bold true; want ' ' fg bold false
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui/component	0.021s
ok  	github.com/Harrison-Blair/hive/internal/tui	0.074s
FAIL
exit 1

### deliberate break 4
--- FAIL: TestDetailsShowsMoreLinesOnOverflow (0.00s)
    details_test.go:49: top:
        ╭─ x ──────────────╮
        │ l0               ┃
        │ l1               │
        │ l2               │
        ╰──────────────────╯
        want:
        ╭─ x ──────────────╮
        │ l0               ┃
        │ l1               │
        │ ↓ 8 more lines   │
        ╰──────────────────╯
    details_test.go:54: offset 3:
        ╭─ x ──────────────╮
        │ l3               ┃
        │ l4               │
        │ l5               │
        ╰──────────────────╯
        want:
        ╭─ x ──────────────╮
        │ l3               ┃
        │ l4               │
        │ ↓ 5 more lines   │
        ╰──────────────────╯
    details_test.go:60: one line over (Thumb(3, 4, 2, 0) is one cell):
        ╭─ x ──────────────╮
        │ l0               ┃
        │ l1               ┃
        │ l2               │
        ╰──────────────────╯
        want:
        ╭─ x ──────────────╮
        │ l0               ┃
        │ l1               │
        │ ↓ 2 more lines   │
        ╰──────────────────╯
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui/component	0.022s
ok  	github.com/Harrison-Blair/hive/internal/tui	0.072s
FAIL
exit 1

### deliberate break 5
--- FAIL: TestDetailsShowsMoreLinesOnOverflow (0.00s)
    details_test.go:60: one line over (Thumb(3, 4, 2, 0) is one cell):
        ╭─ x ──────────────╮
        │ l0               ┃
        │ l1               ┃
        │ ↓ 2 more lines   │
        ╰──────────────────╯
        want:
        ╭─ x ──────────────╮
        │ l0               ┃
        │ l1               │
        │ ↓ 2 more lines   │
        ╰──────────────────╯
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui/component	0.018s
ok  	github.com/Harrison-Blair/hive/internal/tui	0.066s
FAIL
exit 1

### deliberate break 6
ok  	github.com/Harrison-Blair/hive/internal/tui/component	0.018s
--- FAIL: TestFrameRowsMatchEachScreenBoard (0.05s)
    --- FAIL: TestFrameRowsMatchEachScreenBoard/Friction (0.00s)
        frame_test.go:116: 28 frame cells differ from Friction.grid:
            row 39 col 55: want 'g' dim normal, got '/' dim normal
            row 39 col 56: want '/' dim normal, got ' ' dim normal
            row 39 col 57: want 'G' dim normal, got 'f' dim normal
            row 39 col 58: want ' ' dim normal, got 'i' dim normal
            row 39 col 59: want 't' dim normal, got 'l' dim normal
            row 39 col 60: want 'o' dim normal, got 't' dim normal
            row 39 col 61: want 'p' dim normal, got 'e' dim normal
            row 39 col 62: want '/' dim normal, got 'r' dim normal
            row 39 col 63: want 'e' dim normal, got ' ' dim normal
            row 39 col 64: want 'n' dim normal, got '·' dim normal
            row 39 col 65: want 'd' dim normal, got ' ' dim normal
            row 39 col 66: want ' ' dim normal, got 'g' dim normal
            row 39 col 67: want '·' dim normal, got '/' dim normal
            row 39 col 68: want ' ' dim normal, got 'G' dim normal
            row 39 col 69: want '?' dim normal, got ' ' dim normal
            row 39 col 70: want ' ' dim normal, got 't' dim normal
            row 39 col 71: want 'h' dim normal, got 'o' dim normal
            row 39 col 72: want 'e' dim normal, got 'p' dim normal
            row 39 col 73: want 'l' dim normal, got '/' dim normal
            row 39 col 74: want 'p' dim normal, got 'e' dim normal
            row 39 col 75: want ' ' dim normal, got 'n' dim normal
            row 39 col 76: want ' ' dim normal, got 'd' dim normal
            row 39 col 78: want ' ' dim normal, got '·' dim normal
            row 39 col 80: want ' ' dim normal, got '?' dim normal
            row 39 col 82: want ' ' dim normal, got 'h' dim normal
            row 39 col 83: want ' ' dim normal, got 'e' dim normal
            row 39 col 84: want ' ' dim normal, got 'l' dim normal
            row 39 col 85: want ' ' dim normal, got 'p' dim normal
    --- FAIL: TestFrameRowsMatchEachScreenBoard/FrictionResolved (0.00s)
        frame_test.go:116: 28 frame cells differ from FrictionResolved.grid:
            row 39 col 55: want 'g' dim normal, got '/' dim normal
            row 39 col 56: want '/' dim normal, got ' ' dim normal
            row 39 col 57: want 'G' dim normal, got 'f' dim normal
            row 39 col 58: want ' ' dim normal, got 'i' dim normal
            row 39 col 59: want 't' dim normal, got 'l' dim normal
            row 39 col 60: want 'o' dim normal, got 't' dim normal
            row 39 col 61: want 'p' dim normal, got 'e' dim normal
            row 39 col 62: want '/' dim normal, got 'r' dim normal
            row 39 col 63: want 'e' dim normal, got ' ' dim normal
            row 39 col 64: want 'n' dim normal, got '·' dim normal
            row 39 col 65: want 'd' dim normal, got ' ' dim normal
            row 39 col 66: want ' ' dim normal, got 'g' dim normal
            row 39 col 67: want '·' dim normal, got '/' dim normal
            row 39 col 68: want ' ' dim normal, got 'G' dim normal
            row 39 col 69: want '?' dim normal, got ' ' dim normal
            row 39 col 70: want ' ' dim normal, got 't' dim normal
            row 39 col 71: want 'h' dim normal, got 'o' dim normal
            row 39 col 72: want 'e' dim normal, got 'p' dim normal
            row 39 col 73: want 'l' dim normal, got '/' dim normal
            row 39 col 74: want 'p' dim normal, got 'e' dim normal
            row 39 col 75: want ' ' dim normal, got 'n' dim normal
            row 39 col 76: want ' ' dim normal, got 'd' dim normal
            row 39 col 78: want ' ' dim normal, got '·' dim normal
            row 39 col 80: want ' ' dim normal, got '?' dim normal
            row 39 col 82: want ' ' dim normal, got 'h' dim normal
            row 39 col 83: want ' ' dim normal, got 'e' dim normal
            row 39 col 84: want ' ' dim normal, got 'l' dim normal
            row 39 col 85: want ' ' dim normal, got 'p' dim normal
--- FAIL: TestListHintsHaveNoFilterOnWorkAndFriction (0.00s)
    frame_test.go:440: Friction "" list, wide true: hints ["↑↓ move" "tab region" "[ ] view" "enter open" "/ filter" "g/G top/end" "? help"] have / filter
    frame_test.go:440: Friction "Open" list, wide true: hints ["↑↓ move" "tab region" "[ ] view" "enter open" "/ filter" "g/G top/end" "? help"] have / filter
    frame_test.go:440: Friction "Resolved" list, wide true: hints ["↑↓ move" "tab region" "[ ] view" "enter open" "/ filter" "g/G top/end" "? help"] have / filter
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui	0.067s
FAIL
exit 1

### deliberate break 7
ok  	github.com/Harrison-Blair/hive/internal/tui/component	0.018s
--- FAIL: TestListHintsHaveNoFilterOnWorkAndFriction (0.00s)
    frame_test.go:447: Logs events, wide false: hints ["↑↓ move" "enter details" "f follow" "?"] have no / filter
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui	0.066s
FAIL
exit 1

### restored green
ok  	github.com/Harrison-Blair/hive/internal/tui/component	0.018s
ok  	github.com/Harrison-Blair/hive/internal/tui	0.068s
exit 0

### dev sample inputs against task grids
--- FAIL: TestChatGolden (0.00s)
    chat_test.go:45: 789 cells differ from Chat.grid (board row and column):
        row 13 col 23: want ' ' fg normal, got 'B' fg normal
        row 13 col 24: want ' ' fg normal, got 'a' fg normal
        row 13 col 25: want ' ' fg normal, got 'r' fg normal
        row 13 col 26: want ' ' fg normal, got 'e' fg normal
        row 13 col 28: want ' ' fg normal, got 'h' fg normal
        row 13 col 29: want ' ' fg normal, got 'i' fg normal
        row 13 col 30: want ' ' fg normal, got 'v' fg normal
        row 13 col 31: want ' ' fg normal, got 'e' fg normal
        row 13 col 33: want ' ' fg normal, got 'm' fg normal
        row 13 col 34: want ' ' fg normal, got 'a' fg normal
        row 13 col 35: want ' ' fg normal, got 'k' fg normal
        row 13 col 36: want ' ' fg normal, got 'e' fg normal
        row 13 col 37: want ' ' fg normal, got 's' fg normal
        row 13 col 39: want ' ' fg normal, got 't' fg normal
        row 13 col 40: want ' ' fg normal, got 'h' fg normal
        row 13 col 41: want ' ' fg normal, got 'e' fg normal
        row 13 col 43: want ' ' fg normal, got 'c' code normal
        row 13 col 44: want ' ' fg normal, got 'o' code normal
        row 13 col 45: want ' ' fg normal, got 'o' code normal
        row 13 col 46: want ' ' fg normal, got 'r' code normal
        row 13 col 47: want ' ' fg normal, got 'd' code normal
        row 13 col 48: want ' ' fg normal, got 'i' code normal
        row 13 col 49: want ' ' fg normal, got 'n' code normal
        row 13 col 50: want ' ' fg normal, got 'a' code normal
        row 13 col 51: want ' ' fg normal, got 't' code normal
        row 13 col 52: want ' ' fg normal, got 'o' code normal
        row 13 col 53: want ' ' fg normal, got 'r' code normal
        row 13 col 55: want ' ' fg normal, got 't' fg normal
        row 13 col 56: want ' ' fg normal, got 'a' fg normal
        row 13 col 57: want ' ' fg normal, got 'b' fg normal
        row 13 col 58: want ' ' fg normal, got ',' fg normal
        row 13 col 60: want ' ' fg normal, got 'b' fg normal
        row 13 col 61: want ' ' fg normal, got 'u' fg normal
        row 13 col 62: want ' ' fg normal, got 't' fg normal
        row 13 col 64: want ' ' fg normal, got 'i' fg normal
        row 13 col 65: want ' ' fg normal, got 't' fg normal
        row 13 col 66: want ' ' fg normal, got 's' fg normal
        row 13 col 68: want ' ' fg normal, got 'r' fg normal
        row 13 col 69: want ' ' fg normal, got 'e' fg normal
        row 13 col 70: want ' ' fg normal, got 'c' fg normal
        row 13 col 71: want ' ' fg normal, got 'o' fg normal
        row 13 col 72: want ' ' fg normal, got 'r' fg normal
        row 13 col 73: want ' ' fg normal, got 'd' fg normal
        row 13 col 75: want ' ' fg normal, got 'h' fg normal
        row 13 col 76: want ' ' fg normal, got 'a' fg normal
        row 13 col 77: want ' ' fg normal, got 's' fg normal
        row 13 col 79: want ' ' fg normal, got 's' code normal
        row 13 col 80: want ' ' fg normal, got 'p' code normal
        row 13 col 81: want ' ' fg normal, got 'a' code normal
        row 13 col 82: want ' ' fg normal, got 'w' code normal
        row 13 col 83: want ' ' fg normal, got 'n' code normal
        row 13 col 84: want ' ' fg normal, got 'e' code normal
        row 13 col 85: want ' ' fg normal, got 'd' code normal
        row 13 col 86: want ' ' fg normal, got '=' code normal
        row 13 col 87: want ' ' fg normal, got 'f' code normal
        row 13 col 88: want ' ' fg normal, got 'a' code normal
        row 13 col 89: want ' ' fg normal, got 'l' code normal
        row 13 col 90: want ' ' fg normal, got 's' code normal
        row 13 col 91: want ' ' fg normal, got 'e' code normal
        row 13 col 92: want ' ' fg normal, got '.' fg normal
        row 13 col 94: want ' ' fg normal, got 'a' code normal
        row 13 col 95: want ' ' fg normal, got 'g' code normal
        row 13 col 96: want ' ' fg normal, got 'e' code normal
        row 13 col 97: want ' ' fg normal, got 'n' code normal
        row 13 col 98: want ' ' fg normal, got 't' code normal
        row 13 col 99: want ' ' fg normal, got ' ' code normal
        row 13 col 100: want ' ' fg normal, got 's' code normal
        row 13 col 101: want ' ' fg normal, got 't' code normal
        row 13 col 102: want ' ' fg normal, got 'o' code normal
        row 13 col 103: want ' ' fg normal, got 'p' code normal
        row 13 col 105: want ' ' fg normal, got 'r' fg normal
        row 13 col 106: want ' ' fg normal, got 'e' fg normal
        row 13 col 107: want ' ' fg normal, got 'f' fg normal
        row 13 col 108: want ' ' fg normal, got 'u' fg normal
        row 13 col 109: want ' ' fg normal, got 's' fg normal
        row 13 col 110: want ' ' fg normal, got 'e' fg normal
        row 13 col 111: want ' ' fg normal, got 's' fg normal
        row 13 col 113: want ' ' fg normal, got 'i' fg normal
        row 13 col 114: want ' ' fg normal, got 't' fg normal
        row 13 col 115: want ' ' fg normal, got '.' fg normal
        row 14 col 23: want 'B' fg normal, got ' ' fg normal
        row 14 col 24: want 'a' fg normal, got ' ' fg normal
        row 14 col 25: want 'r' fg normal, got ' ' fg normal
        row 14 col 26: want 'e' fg normal, got ' ' fg normal
        row 14 col 28: want 'h' fg normal, got ' ' fg normal
        row 14 col 29: want 'i' fg normal, got ' ' fg normal
        row 14 col 30: want 'v' fg normal, got ' ' fg normal
        row 14 col 31: want 'e' fg normal, got ' ' fg normal
        row 14 col 33: want 'm' fg normal, got ' ' fg normal
        row 14 col 34: want 'a' fg normal, got ' ' fg normal
        row 14 col 35: want 'k' fg normal, got ' ' fg normal
        row 14 col 36: want 'e' fg normal, got ' ' fg normal
        row 14 col 37: want 's' fg normal, got ' ' fg normal
        row 14 col 39: want 't' fg normal, got ' ' fg normal
        row 14 col 40: want 'h' fg normal, got ' ' fg normal
        row 14 col 41: want 'e' fg normal, got ' ' fg normal
        row 14 col 43: want 'c' code normal, got ' ' fg normal
        row 14 col 44: want 'o' code normal, got ' ' fg normal
        row 14 col 45: want 'o' code normal, got ' ' fg normal
        row 14 col 46: want 'r' code normal, got ' ' fg normal
        row 14 col 47: want 'd' code normal, got ' ' fg normal
        row 14 col 48: want 'i' code normal, got ' ' fg normal
        row 14 col 49: want 'n' code normal, got ' ' fg normal
        row 14 col 50: want 'a' code normal, got ' ' fg normal
        row 14 col 51: want 't' code normal, got ' ' fg normal
        row 14 col 52: want 'o' code normal, got ' ' fg normal
        row 14 col 53: want 'r' code normal, got ' ' fg normal
        row 14 col 55: want 't' fg normal, got ' ' fg normal
        row 14 col 56: want 'a' fg normal, got ' ' fg normal
        row 14 col 57: want 'b' fg normal, got ' ' fg normal
        row 14 col 58: want ',' fg normal, got ' ' fg normal
        row 14 col 60: want 'b' fg normal, got ' ' fg normal
        row 14 col 61: want 'u' fg normal, got ' ' fg normal
        row 14 col 62: want 't' fg normal, got ' ' fg normal
        row 14 col 64: want 'i' fg normal, got ' ' fg normal
        row 14 col 65: want 't' fg normal, got ' ' fg normal
        row 14 col 66: want 's' fg normal, got ' ' fg normal
        row 14 col 68: want 'r' fg normal, got ' ' fg normal
        row 14 col 69: want 'e' fg normal, got ' ' fg normal
        row 14 col 70: want 'c' fg normal, got ' ' fg normal
        row 14 col 71: want 'o' fg normal, got ' ' fg normal
        row 14 col 72: want 'r' fg normal, got ' ' fg normal
        row 14 col 73: want 'd' fg normal, got ' ' fg normal
        row 14 col 75: want 'h' fg normal, got ' ' fg normal
        row 14 col 76: want 'a' fg normal, got ' ' fg normal
        row 14 col 77: want 's' fg normal, got ' ' fg normal
        row 14 col 79: want 's' code normal, got ' ' fg normal
        row 14 col 80: want 'p' code normal, got ' ' fg normal
        row 14 col 81: want 'a' code normal, got ' ' fg normal
        row 14 col 82: want 'w' code normal, got ' ' fg normal
        row 14 col 83: want 'n' code normal, got ' ' fg normal
        row 14 col 84: want 'e' code normal, got ' ' fg normal
        row 14 col 85: want 'd' code normal, got ' ' fg normal
        row 14 col 86: want '=' code normal, got ' ' fg normal
        row 14 col 87: want 'f' code normal, got ' ' fg normal
        row 14 col 88: want 'a' code normal, got ' ' fg normal
        row 14 col 89: want 'l' code normal, got ' ' fg normal
        row 14 col 90: want 's' code normal, got ' ' fg normal
        row 14 col 91: want 'e' code normal, got ' ' fg normal
        row 14 col 92: want '.' fg normal, got ' ' fg normal
        row 14 col 94: want 'a' code normal, got ' ' fg normal
        row 14 col 95: want 'g' code normal, got ' ' fg normal
        row 14 col 96: want 'e' code normal, got ' ' fg normal
        row 14 col 97: want 'n' code normal, got ' ' fg normal
        row 14 col 98: want 't' code normal, got ' ' fg normal
        row 14 col 99: want ' ' code normal, got ' ' fg normal
        row 14 col 100: want 's' code normal, got ' ' fg normal
        row 14 col 101: want 't' code normal, got ' ' fg normal
        row 14 col 102: want 'o' code normal, got ' ' fg normal
        row 14 col 103: want 'p' code normal, got ' ' fg normal
        row 14 col 105: want 'r' fg normal, got ' ' fg normal
        row 14 col 106: want 'e' fg normal, got ' ' fg normal
        row 14 col 107: want 'f' fg normal, got ' ' fg normal
        row 14 col 108: want 'u' fg normal, got ' ' fg normal
        row 14 col 109: want 's' fg normal, got ' ' fg normal
        row 14 col 110: want 'e' fg normal, got ' ' fg normal
        row 14 col 111: want 's' fg normal, got ' ' fg normal
        row 14 col 113: want 'i' fg normal, got ' ' fg normal
        row 14 col 114: want 't' fg normal, got ' ' fg normal
        row 14 col 115: want '.' fg normal, got ' ' fg normal
        row 15 col 24: want ' ' fg normal, got 'C' fg normal
        row 15 col 25: want ' ' fg normal, got 'h' fg normal
        row 15 col 26: want ' ' fg normal, got 'o' fg normal
        row 15 col 27: want ' ' fg normal, got 'i' fg normal
        row 15 col 28: want ' ' fg normal, got 'c' fg normal
        row 15 col 29: want ' ' fg normal, got 'e' fg normal
        row 15 col 42: want ' ' fg normal, got '│' border normal
        row 15 col 44: want ' ' fg normal, got 'A' fg normal
        row 15 col 45: want ' ' fg normal, got 'n' fg normal
        row 15 col 46: want ' ' fg normal, got 's' fg normal
        row 15 col 47: want ' ' fg normal, got 'w' fg normal
        row 15 col 48: want ' ' fg normal, got 'e' fg normal
        row 15 col 49: want ' ' fg normal, got 'r' fg normal
        row 15 col 78: want ' ' fg normal, got '│' border normal
        row 15 col 80: want ' ' fg normal, got 'P' fg normal
        row 15 col 81: want ' ' fg normal, got 'r' fg normal
        row 15 col 82: want ' ' fg normal, got 'o' fg normal
        row 15 col 83: want ' ' fg normal, got 's' fg normal
        row 15 col 97: want ' ' fg normal, got '│' border normal
        row 15 col 99: want ' ' fg normal, got 'C' fg normal
        row 15 col 100: want ' ' fg normal, got 'o' fg normal
        row 15 col 101: want ' ' fg normal, got 'n' fg normal
        row 15 col 102: want ' ' fg normal, got 's' fg normal
        row 16 col 23: want ' ' fg normal, got '─' border normal
        row 16 col 24: want 'C' fg normal, got '─' border normal
        row 16 col 25: want 'h' fg normal, got '─' border normal
        row 16 col 26: want 'o' fg normal, got '─' border normal
        row 16 col 27: want 'i' fg normal, got '─' border normal
        row 16 col 28: want 'c' fg normal, got '─' border normal
        row 16 col 29: want 'e' fg normal, got '─' border normal
        row 16 col 30: want ' ' fg normal, got '─' border normal
        row 16 col 31: want ' ' fg normal, got '─' border normal
        row 16 col 32: want ' ' fg normal, got '─' border normal
        row 16 col 33: want ' ' fg normal, got '─' border normal
        row 16 col 34: want ' ' fg normal, got '─' border normal
        row 16 col 35: want ' ' fg normal, got '─' border normal
        row 16 col 36: want ' ' fg normal, got '─' border normal
        row 16 col 37: want ' ' fg normal, got '─' border normal
        row 16 col 38: want ' ' fg normal, got '─' border normal
        row 16 col 39: want ' ' fg normal, got '─' border normal
        row 16 col 40: want ' ' fg normal, got '─' border normal
        row 16 col 41: want ' ' fg normal, got '─' border normal
        row 16 col 42: want '│' border normal, got '┼' border normal
        row 16 col 43: want ' ' fg normal, got '─' border normal
        row 16 col 44: want 'A' fg normal, got '─' border normal
        row 16 col 45: want 'n' fg normal, got '─' border normal
        row 16 col 46: want 's' fg normal, got '─' border normal
        row 16 col 47: want 'w' fg normal, got '─' border normal
        row 16 col 48: want 'e' fg normal, got '─' border normal
        row 16 col 49: want 'r' fg normal, got '─' border normal
        row 16 col 50: want ' ' fg normal, got '─' border normal
        row 16 col 51: want ' ' fg normal, got '─' border normal
        row 16 col 52: want ' ' fg normal, got '─' border normal
        row 16 col 53: want ' ' fg normal, got '─' border normal
        row 16 col 54: want ' ' fg normal, got '─' border normal
        row 16 col 55: want ' ' fg normal, got '─' border normal
        row 16 col 56: want ' ' fg normal, got '─' border normal
        row 16 col 57: want ' ' fg normal, got '─' border normal
        row 16 col 58: want ' ' fg normal, got '─' border normal
        row 16 col 59: want ' ' fg normal, got '─' border normal
        row 16 col 60: want ' ' fg normal, got '─' border normal
        row 16 col 61: want ' ' fg normal, got '─' border normal
        row 16 col 62: want ' ' fg normal, got '─' border normal
        row 16 col 63: want ' ' fg normal, got '─' border normal
        row 16 col 64: want ' ' fg normal, got '─' border normal
        row 16 col 65: want ' ' fg normal, got '─' border normal
        row 16 col 66: want ' ' fg normal, got '─' border normal
        row 16 col 67: want ' ' fg normal, got '─' border normal
        row 16 col 68: want ' ' fg normal, got '─' border normal
        row 16 col 69: want ' ' fg normal, got '─' border normal
        row 16 col 70: want ' ' fg normal, got '─' border normal
        row 16 col 71: want ' ' fg normal, got '─' border normal
        row 16 col 72: want ' ' fg normal, got '─' border normal
        row 16 col 73: want ' ' fg normal, got '─' border normal
        row 16 col 74: want ' ' fg normal, got '─' border normal
        row 16 col 75: want ' ' fg normal, got '─' border normal
        row 16 col 76: want ' ' fg normal, got '─' border normal
        row 16 col 77: want ' ' fg normal, got '─' border normal
        row 16 col 78: want '│' border normal, got '┼' border normal
        row 16 col 79: want ' ' fg normal, got '─' border normal
        row 16 col 80: want 'P' fg normal, got '─' border normal
        row 16 col 81: want 'r' fg normal, got '─' border normal
        row 16 col 82: want 'o' fg normal, got '─' border normal
        row 16 col 83: want 's' fg normal, got '─' border normal
        row 16 col 84: want ' ' fg normal, got '─' border normal
        row 16 col 85: want ' ' fg normal, got '─' border normal
        row 16 col 86: want ' ' fg normal, got '─' border normal
        row 16 col 87: want ' ' fg normal, got '─' border normal
        row 16 col 88: want ' ' fg normal, got '─' border normal
        row 16 col 89: want ' ' fg normal, got '─' border normal
        row 16 col 90: want ' ' fg normal, got '─' border normal
        row 16 col 91: want ' ' fg normal, got '─' border normal
        row 16 col 92: want ' ' fg normal, got '─' border normal
        row 16 col 93: want ' ' fg normal, got '─' border normal
        row 16 col 94: want ' ' fg normal, got '─' border normal
        row 16 col 95: want ' ' fg normal, got '─' border normal
        row 16 col 96: want ' ' fg normal, got '─' border normal
        row 16 col 97: want '│' border normal, got '┼' border normal
        row 16 col 98: want ' ' fg normal, got '─' border normal
        row 16 col 99: want 'C' fg normal, got '─' border normal
        row 16 col 100: want 'o' fg normal, got '─' border normal
        row 16 col 101: want 'n' fg normal, got '─' border normal
        row 16 col 102: want 's' fg normal, got '─' border normal
        row 16 col 103: want ' ' fg normal, got '─' border normal
        row 16 col 104: want ' ' fg normal, got '─' border normal
        row 16 col 105: want ' ' fg normal, got '─' border normal
        row 16 col 106: want ' ' fg normal, got '─' border normal
        row 16 col 107: want ' ' fg normal, got '─' border normal
        row 16 col 108: want ' ' fg normal, got '─' border normal
        row 16 col 109: want ' ' fg normal, got '─' border normal
        row 16 col 110: want ' ' fg normal, got '─' border normal
        row 16 col 111: want ' ' fg normal, got '─' border normal
        row 16 col 112: want ' ' fg normal, got '─' border normal
        row 16 col 113: want ' ' fg normal, got '─' border normal
        row 16 col 114: want ' ' fg normal, got '─' border normal
        row 16 col 115: want ' ' fg normal, got '─' border normal
        row 16 col 116: want ' ' fg normal, got '─' border normal
        row 17 col 23: want '─' border normal, got ' ' fg normal
        row 17 col 24: want '─' border normal, got 'A' fg normal
        row 17 col 25: want '─' border normal, got ' ' fg normal
        row 17 col 26: want '─' border normal, got '(' fg normal
        row 17 col 27: want '─' border normal, got 'R' fg normal
        row 17 col 28: want '─' border normal, got 'E' fg normal
        row 17 col 29: want '─' border normal, got 'C' fg normal
        row 17 col 30: want '─' border normal, got ')' fg normal
        row 17 col 31: want '─' border normal, got ' ' fg normal
        row 17 col 32: want '─' border normal, got ' ' fg normal
        row 17 col 33: want '─' border normal, got ' ' fg normal
        row 17 col 34: want '─' border normal, got ' ' fg normal
        row 17 col 35: want '─' border normal, got ' ' fg normal
        row 17 col 36: want '─' border normal, got ' ' fg normal
        row 17 col 37: want '─' border normal, got ' ' fg normal
        row 17 col 38: want '─' border normal, got ' ' fg normal
        row 17 col 39: want '─' border normal, got ' ' fg normal
        row 17 col 40: want '─' border normal, got ' ' fg normal
        row 17 col 41: want '─' border normal, got ' ' fg normal
        row 17 col 42: want '┼' border normal, got '│' border normal
        row 17 col 43: want '─' border normal, got ' ' fg normal
        row 17 col 44: want '─' border normal, got 'R' fg normal
        row 17 col 45: want '─' border normal, got 'e' fg normal
        row 17 col 46: want '─' border normal, got 'c' fg normal
        row 17 col 47: want '─' border normal, got 'o' fg normal
        row 17 col 48: want '─' border normal, got 'r' fg normal
        row 17 col 49: want '─' border normal, got 'd' fg normal
        row 17 col 50: want '─' border normal, got ' ' fg normal
        row 17 col 51: want '─' border normal, got 't' fg normal
        row 17 col 52: want '─' border normal, got 'h' fg normal
        row 17 col 53: want '─' border normal, got 'a' fg normal
        row 17 col 54: want '─' border normal, got 't' fg normal
        row 17 col 55: want '─' border normal, got ' ' fg normal
        row 17 col 56: want '─' border normal, got 'b' fg normal
        row 17 col 57: want '─' border normal, got 'a' fg normal
        row 17 col 58: want '─' border normal, got 'r' fg normal
        row 17 col 59: want '─' border normal, got 'e' fg normal
        row 17 col 60: want '─' border normal, got ' ' fg normal
        row 17 col 61: want '─' border normal, got 'h' fg normal
        row 17 col 62: want '─' border normal, got 'i' fg normal
        row 17 col 63: want '─' border normal, got 'v' fg normal
        row 17 col 64: want '─' border normal, got 'e' fg normal
        row 17 col 65: want '─' border normal, got ' ' fg normal
        row 17 col 66: want '─' border normal, got 's' fg normal
        row 17 col 67: want '─' border normal, got 't' fg normal
        row 17 col 68: want '─' border normal, got 'a' fg normal
        row 17 col 69: want '─' border normal, got 'r' fg normal
        row 17 col 70: want '─' border normal, got 't' fg normal
        row 17 col 71: want '─' border normal, got 'e' fg normal
        row 17 col 72: want '─' border normal, got 'd' fg normal
        row 17 col 73: want '─' border normal, got ' ' fg normal
        row 17 col 74: want '─' border normal, got 'i' fg normal
        row 17 col 75: want '─' border normal, got 't' fg normal
        row 17 col 76: want '─' border normal, got ' ' fg normal
        row 17 col 77: want '─' border normal, got ' ' fg normal
        row 17 col 78: want '┼' border normal, got '│' border normal
        row 17 col 79: want '─' border normal, got ' ' fg normal
        row 17 col 80: want '─' border normal, got 'N' fg normal
        row 17 col 81: want '─' border normal, got 'o' fg normal
        row 17 col 82: want '─' border normal, got ' ' fg normal
        row 17 col 83: want '─' border normal, got 't' fg normal
        row 17 col 84: want '─' border normal, got 'a' fg normal
        row 17 col 85: want '─' border normal, got 'b' fg normal
        row 17 col 86: want '─' border normal, got ' ' fg normal
        row 17 col 87: want '─' border normal, got 'p' fg normal
        row 17 col 88: want '─' border normal, got 'i' fg normal
        row 17 col 89: want '─' border normal, got 'l' fg normal
        row 17 col 90: want '─' border normal, got 'e' fg normal
        row 17 col 91: want '─' border normal, got '-' fg normal
        row 17 col 92: want '─' border normal, got 'u' fg normal
        row 17 col 93: want '─' border normal, got 'p' fg normal
        row 17 col 94: want '─' border normal, got ' ' fg normal
        row 17 col 95: want '─' border normal, got ' ' fg normal
        row 17 col 96: want '─' border normal, got ' ' fg normal
        row 17 col 97: want '┼' border normal, got '│' border normal
        row 17 col 98: want '─' border normal, got ' ' fg normal
        row 17 col 99: want '─' border normal, got 'F' fg normal
        row 17 col 100: want '─' border normal, got 'i' fg normal
        row 17 col 101: want '─' border normal, got 'e' fg normal
        row 17 col 102: want '─' border normal, got 'l' fg normal
        row 17 col 103: want '─' border normal, got 'd' fg normal
        row 17 col 104: want '─' border normal, got ' ' fg normal
        row 17 col 105: want '─' border normal, got 'c' fg normal
        row 17 col 106: want '─' border normal, got 'h' fg normal
        row 17 col 107: want '─' border normal, got 'a' fg normal
        row 17 col 108: want '─' border normal, got 'n' fg normal
        row 17 col 109: want '─' border normal, got 'g' fg normal
        row 17 col 110: want '─' border normal, got 'e' fg normal
        row 17 col 111: want '─' border normal, got ' ' fg normal
        row 17 col 112: want '─' border normal, got ' ' fg normal
        row 17 col 113: want '─' border normal, got ' ' fg normal
        row 17 col 114: want '─' border normal, got ' ' fg normal
        row 17 col 115: want '─' border normal, got ' ' fg normal
        row 17 col 116: want '─' border normal, got ' ' fg normal
        row 18 col 24: want 'A' fg normal, got 'B' fg normal
        row 18 col 26: want '(' fg normal, got ' ' fg normal
        row 18 col 27: want 'R' fg normal, got ' ' fg normal
        row 18 col 28: want 'E' fg normal, got ' ' fg normal
        row 18 col 29: want 'C' fg normal, got ' ' fg normal
        row 18 col 30: want ')' fg normal, got ' ' fg normal
        row 18 col 44: want 'R' fg normal, got 'L' fg normal
        row 18 col 46: want 'c' fg normal, got 'a' fg normal
        row 18 col 47: want 'o' fg normal, got 'v' fg normal
        row 18 col 48: want 'r' fg normal, got 'e' fg normal
        row 18 col 49: want 'd' fg normal, got ' ' fg normal
        row 18 col 50: want ' ' fg normal, got 't' fg normal
        row 18 col 51: want 't' fg normal, got 'h' fg normal
        row 18 col 52: want 'h' fg normal, got 'e' fg normal
        row 18 col 53: want 'a' fg normal, got ' ' fg normal
        row 18 col 54: want 't' fg normal, got 'o' fg normal
        row 18 col 55: want ' ' fg normal, got 'l' fg normal
        row 18 col 56: want 'b' fg normal, got 'd' fg normal
        row 18 col 57: want 'a' fg normal, got ' ' fg normal
        row 18 col 58: want 'r' fg normal, got 't' fg normal
        row 18 col 59: want 'e' fg normal, got 'a' fg normal
        row 18 col 60: want ' ' fg normal, got 'b' fg normal
        row 18 col 61: want 'h' fg normal, got ',' fg normal
        row 18 col 62: want 'i' fg normal, got ' ' fg normal
        row 18 col 63: want 'v' fg normal, got 'o' fg normal
        row 18 col 64: want 'e' fg normal, got 'p' fg normal
        row 18 col 65: want ' ' fg normal, got 'e' fg normal
        row 18 col 66: want 's' fg normal, got 'n' fg normal
        row 18 col 67: want 't' fg normal, got ' ' fg normal
        row 18 col 69: want 'r' fg normal, got ' ' fg normal
        row 18 col 70: want 't' fg normal, got 'n' fg normal
        row 18 col 72: want 'd' fg normal, got 'w' fg normal
        row 18 col 74: want 'i' fg normal, got 'o' fg normal
        row 18 col 75: want 't' fg normal, got 'n' fg normal
        row 18 col 76: want ' ' fg normal, got 'e' fg normal
        row 18 col 80: want 'N' fg normal, got 'S' fg normal
        row 18 col 81: want 'o' fg normal, got 'm' fg normal
        row 18 col 82: want ' ' fg normal, got 'a' fg normal
        row 18 col 83: want 't' fg normal, got 'l' fg normal
        row 18 col 84: want 'a' fg normal, got 'l' fg normal
        row 18 col 85: want 'b' fg normal, got 'e' fg normal
        row 18 col 86: want ' ' fg normal, got 's' fg normal
        row 18 col 87: want 'p' fg normal, got 't' fg normal
        row 18 col 88: want 'i' fg normal, got ' ' fg normal
        row 18 col 89: want 'l' fg normal, got 'c' fg normal
        row 18 col 90: want 'e' fg normal, got 'h' fg normal
        row 18 col 91: want '-' fg normal, got 'a' fg normal
        row 18 col 92: want 'u' fg normal, got 'n' fg normal
        row 18 col 93: want 'p' fg normal, got 'g' fg normal
        row 18 col 94: want ' ' fg normal, got 'e' fg normal
        row 18 col 99: want 'F' fg normal, got 'D' fg normal
        row 18 col 100: want 'i' fg normal, got 'e' fg normal
        row 18 col 101: want 'e' fg normal, got 'a' fg normal
        row 18 col 102: want 'l' fg normal, got 'd' fg normal
        row 18 col 103: want 'd' fg normal, got ' ' fg normal
        row 18 col 104: want ' ' fg normal, got 't' fg normal
        row 18 col 105: want 'c' fg normal, got 'a' fg normal
        row 18 col 106: want 'h' fg normal, got 'b' fg normal
        row 18 col 107: want 'a' fg normal, got 's' fg normal
        row 18 col 108: want 'n' fg normal, got ' ' fg normal
        row 18 col 109: want 'g' fg normal, got 'p' fg normal
        row 18 col 110: want 'e' fg normal, got 'i' fg normal
        row 18 col 111: want ' ' fg normal, got 'l' fg normal
        row 18 col 112: want ' ' fg normal, got 'e' fg normal
        row 18 col 114: want ' ' fg normal, got 'u' fg normal
        row 18 col 115: want ' ' fg normal, got 'p' fg normal
        row 19 col 24: want 'B' fg normal, got 'C' fg normal
        row 19 col 44: want 'L' fg normal, got 'R' fg normal
        row 19 col 46: want 'a' fg normal, got 'u' fg normal
        row 19 col 47: want 'v' fg normal, got 's' fg normal
        row 19 col 54: want 'o' fg normal, got 's' fg normal
        row 19 col 55: want 'l' fg normal, got 'a' fg normal
        row 19 col 56: want 'd' fg normal, got 'm' fg normal
        row 19 col 57: want ' ' fg normal, got 'e' fg normal
        row 19 col 58: want 't' fg normal, got ' ' fg normal
        row 19 col 59: want 'a' fg normal, got 'p' fg normal
        row 19 col 60: want 'b' fg normal, got 'a' fg normal
        row 19 col 61: want ',' fg normal, got 'n' fg normal
        row 19 col 62: want ' ' fg normal, got 'e' fg normal
        row 19 col 63: want 'o' fg normal, got ' ' fg normal
        row 19 col 64: want 'p' fg normal, got ' ' fg normal
        row 19 col 65: want 'e' fg normal, got ' ' fg normal
        row 19 col 66: want 'n' fg normal, got ' ' fg normal
        row 19 col 68: want 'a' fg normal, got ' ' fg normal
        row 19 col 70: want 'n' fg normal, got ' ' fg normal
        row 19 col 71: want 'e' fg normal, got ' ' fg normal
        row 19 col 72: want 'w' fg normal, got ' ' fg normal
        row 19 col 74: want 'o' fg normal, got ' ' fg normal
        row 19 col 75: want 'n' fg normal, got ' ' fg normal
        row 19 col 76: want 'e' fg normal, got ' ' fg normal
        row 19 col 80: want 'S' fg normal, got 'O' fg normal
        row 19 col 81: want 'm' fg normal, got 'n' fg normal
        row 19 col 82: want 'a' fg normal, got 'e' fg normal
        row 19 col 83: want 'l' fg normal, got ' ' fg normal
        row 19 col 84: want 'l' fg normal, got 't' fg normal
        row 19 col 85: want 'e' fg normal, got 'a' fg normal
        row 19 col 86: want 's' fg normal, got 'b' fg normal
        row 19 col 87: want 't' fg normal, got ' ' fg normal
        row 19 col 88: want ' ' fg normal, got 'a' fg normal
        row 19 col 89: want 'c' fg normal, got 'l' fg normal
        row 19 col 90: want 'h' fg normal, got 'w' fg normal
        row 19 col 92: want 'n' fg normal, got 'y' fg normal
        row 19 col 93: want 'g' fg normal, got 's' fg normal
        row 19 col 94: want 'e' fg normal, got ' ' fg normal
        row 19 col 99: want 'D' fg normal, got 'M' fg normal
        row 19 col 100: want 'e' fg normal, got 'u' fg normal
        row 19 col 101: want 'a' fg normal, got 's' fg normal
        row 19 col 102: want 'd' fg normal, got 't' fg normal
        row 19 col 104: want 't' fg normal, got 'e' fg normal
        row 19 col 105: want 'a' fg normal, got 'x' fg normal
        row 19 col 106: want 'b' fg normal, got 'i' fg normal
        row 19 col 107: want 's' fg normal, got 't' fg normal
        row 19 col 109: want 'p' fg normal, got 'h' fg normal
        row 19 col 110: want 'i' fg normal, got 'a' fg normal
        row 19 col 111: want 'l' fg normal, got 'r' fg normal
        row 19 col 112: want 'e' fg normal, got 'n' fg normal
        row 19 col 113: want ' ' fg normal, got 'e' fg normal
        row 19 col 114: want 'u' fg normal, got 's' fg normal
        row 19 col 115: want 'p' fg normal, got 's' fg normal
        row 20 col 24: want 'C' fg normal, got ' ' fg normal
        row 20 col 42: want '│' border normal, got ' ' fg normal
        row 20 col 44: want 'R' fg normal, got ' ' fg normal
        row 20 col 45: want 'e' fg normal, got ' ' fg normal
        row 20 col 46: want 'u' fg normal, got ' ' fg normal
        row 20 col 47: want 's' fg normal, got ' ' fg normal
        row 20 col 48: want 'e' fg normal, got ' ' fg normal
        row 20 col 50: want 't' fg normal, got ' ' fg normal
        row 20 col 51: want 'h' fg normal, got ' ' fg normal
        row 20 col 52: want 'e' fg normal, got ' ' fg normal
        row 20 col 54: want 's' fg normal, got ' ' fg normal
        row 20 col 55: want 'a' fg normal, got ' ' fg normal
        row 20 col 56: want 'm' fg normal, got ' ' fg normal
        row 20 col 57: want 'e' fg normal, got ' ' fg normal
        row 20 col 59: want 'p' fg normal, got ' ' fg normal
        row 20 col 60: want 'a' fg normal, got ' ' fg normal
        row 20 col 61: want 'n' fg normal, got ' ' fg normal
        row 20 col 62: want 'e' fg normal, got ' ' fg normal
        row 20 col 78: want '│' border normal, got ' ' fg normal
        row 20 col 80: want 'O' fg normal, got ' ' fg normal
        row 20 col 81: want 'n' fg normal, got ' ' fg normal
        row 20 col 82: want 'e' fg normal, got ' ' fg normal
        row 20 col 84: want 't' fg normal, got ' ' fg normal
        row 20 col 85: want 'a' fg normal, got ' ' fg normal
        row 20 col 86: want 'b' fg normal, got ' ' fg normal
        row 20 col 88: want 'a' fg normal, got ' ' fg normal
        row 20 col 89: want 'l' fg normal, got ' ' fg normal
        row 20 col 90: want 'w' fg normal, got ' ' fg normal
        row 20 col 91: want 'a' fg normal, got ' ' fg normal
        row 20 col 92: want 'y' fg normal, got ' ' fg normal
        row 20 col 93: want 's' fg normal, got ' ' fg normal
        row 20 col 97: want '│' border normal, got ' ' fg normal
        row 20 col 99: want 'M' fg normal, got ' ' fg normal
        row 20 col 100: want 'u' fg normal, got ' ' fg normal
        row 20 col 101: want 's' fg normal, got ' ' fg normal
        row 20 col 102: want 't' fg normal, got ' ' fg normal
        row 20 col 104: want 'e' fg normal, got ' ' fg normal
        row 20 col 105: want 'x' fg normal, got ' ' fg normal
        row 20 col 106: want 'i' fg normal, got ' ' fg normal
        row 20 col 107: want 't' fg normal, got ' ' fg normal
        row 20 col 109: want 'h' fg normal, got ' ' fg normal
        row 20 col 110: want 'a' fg normal, got ' ' fg normal
        row 20 col 111: want 'r' fg normal, got ' ' fg normal
        row 20 col 112: want 'n' fg normal, got ' ' fg normal
        row 20 col 113: want 'e' fg normal, got ' ' fg normal
        row 20 col 114: want 's' fg normal, got ' ' fg normal
        row 20 col 115: want 's' fg normal, got ' ' fg normal
        row 21 col 23: want ' ' fg normal, got '•' accent normal
        row 21 col 24: want ' ' fg normal, got ' ' accent normal
        row 21 col 25: want ' ' fg normal, got 'R' fg bold
        row 21 col 26: want ' ' fg normal, got 'e' fg bold
        row 21 col 27: want ' ' fg normal, got 'c' fg bold
        row 21 col 28: want ' ' fg normal, got 'o' fg bold
        row 21 col 29: want ' ' fg normal, got 'm' fg bold
        row 21 col 30: want ' ' fg normal, got 'm' fg bold
        row 21 col 31: want ' ' fg normal, got 'e' fg bold
        row 21 col 32: want ' ' fg normal, got 'n' fg bold
        row 21 col 33: want ' ' fg normal, got 'd' fg bold
        row 21 col 34: want ' ' fg normal, got 'e' fg bold
        row 21 col 35: want ' ' fg normal, got 'd' fg bold
        row 21 col 36: want ' ' fg normal, got ':' fg bold
        row 21 col 38: want ' ' fg normal, got 'A' fg normal
        row 21 col 39: want ' ' fg normal, got ',' fg normal
        row 21 col 41: want ' ' fg normal, got 'b' fg normal
        row 21 col 42: want ' ' fg normal, got 'e' fg normal
        row 21 col 43: want ' ' fg normal, got 'c' fg normal
        row 21 col 44: want ' ' fg normal, got 'a' fg normal
        row 21 col 45: want ' ' fg normal, got 'u' fg normal
        row 21 col 46: want ' ' fg normal, got 's' fg normal
        row 21 col 47: want ' ' fg normal, got 'e' fg normal
        row 21 col 49: want ' ' fg normal, got 'h' fg normal
        row 21 col 50: want ' ' fg normal, got 'i' fg normal
        row 21 col 51: want ' ' fg normal, got 'v' fg normal
        row 21 col 52: want ' ' fg normal, got 'e' fg normal
        row 21 col 54: want ' ' fg normal, got 'm' fg normal
        row 21 col 55: want ' ' fg normal, got 'a' fg normal
        row 21 col 56: want ' ' fg normal, got 'y' fg normal
        row 21 col 58: want ' ' fg normal, got 'c' fg normal
        row 21 col 59: want ' ' fg normal, got 'l' fg normal
        row 21 col 60: want ' ' fg normal, got 'o' fg normal
        row 21 col 61: want ' ' fg normal, got 's' fg normal
        row 21 col 62: want ' ' fg normal, got 'e' fg normal
        row 21 col 64: want ' ' fg normal, got 'o' fg normal
        row 21 col 65: want ' ' fg normal, got 'n' fg normal
        row 21 col 66: want ' ' fg normal, got 'l' fg normal
        row 21 col 67: want ' ' fg normal, got 'y' fg normal
        row 21 col 69: want ' ' fg normal, got 'p' fg normal
        row 21 col 70: want ' ' fg normal, got 'a' fg normal
        row 21 col 71: want ' ' fg normal, got 'n' fg normal
        row 21 col 72: want ' ' fg normal, got 'e' fg normal
        row 21 col 73: want ' ' fg normal, got 's' fg normal
        row 21 col 75: want ' ' fg normal, got 'i' fg normal
        row 21 col 76: want ' ' fg normal, got 't' fg normal
        row 21 col 78: want ' ' fg normal, got 's' fg normal
        row 21 col 79: want ' ' fg normal, got 't' fg normal
        row 21 col 80: want ' ' fg normal, got 'a' fg normal
        row 21 col 81: want ' ' fg normal, got 'r' fg normal
        row 21 col 82: want ' ' fg normal, got 't' fg normal
        row 21 col 83: want ' ' fg normal, got 'e' fg normal
        row 21 col 84: want ' ' fg normal, got 'd' fg normal
        row 21 col 85: want ' ' fg normal, got '.' fg normal
        row 22 col 25: want 'R' fg bold, got 'R' fg normal
        row 22 col 26: want 'e' fg bold, got 'e' fg normal
        row 22 col 27: want 'c' fg bold, got 'p' fg normal
        row 22 col 28: want 'o' fg bold, got 'l' fg normal
        row 22 col 29: want 'm' fg bold, got 'y' fg normal
        row 22 col 30: want 'm' fg bold, got ' ' fg normal
        row 22 col 31: want 'e' fg bold, got 'l' fg normal
        row 22 col 32: want 'n' fg bold, got 'i' fg normal
        row 22 col 33: want 'd' fg bold, got 'k' fg normal
        row 22 col 34: want 'e' fg bold, got 'e' fg normal
        row 22 col 35: want 'd' fg bold, got ' ' fg normal
        row 22 col 36: want ':' fg bold, got '1' code normal
        row 22 col 37: want ' ' fg normal, got 'A' code normal
        row 22 col 38: want 'A' fg normal, got ',' code normal
        row 22 col 39: want ',' fg normal, got ' ' code normal
        row 22 col 40: want ' ' fg normal, got '2' code normal
        row 22 col 41: want 'b' fg normal, got 'A' code normal
        row 22 col 42: want 'e' fg normal, got ',' code normal
        row 22 col 43: want 'c' fg normal, got ' ' code normal
        row 22 col 44: want 'a' fg normal, got '3' code normal
        row 22 col 45: want 'u' fg normal, got 'A' code normal
        row 22 col 46: want 's' fg normal, got ',' fg normal
        row 22 col 47: want 'e' fg normal, got ' ' fg normal
        row 22 col 48: want ' ' fg normal, got 'o' fg normal
        row 22 col 49: want 'h' fg normal, got 'r' fg normal
        row 22 col 50: want 'i' fg normal, got ' ' fg normal
        row 22 col 51: want 'v' fg normal, got 's' code normal
        row 22 col 52: want 'e' fg normal, got 'i' code normal
        row 22 col 53: want ' ' fg normal, got 'm' code normal
        row 22 col 54: want 'm' fg normal, got 'p' code normal
        row 22 col 55: want 'a' fg normal, got 'l' code normal
        row 22 col 56: want 'y' fg normal, got 'i' code normal
        row 22 col 57: want ' ' fg normal, got 'f' code normal
        row 22 col 58: want 'c' fg normal, got 'y' code normal
        row 22 col 59: want 'l' fg normal, got ' ' code normal
        row 22 col 60: want 'o' fg normal, got 'N' code normal
        row 22 col 61: want 's' fg normal, got '.' fg normal
        row 22 col 62: want 'e' fg normal, got ' ' fg normal
        row 22 col 64: want 'o' fg normal, got ' ' fg normal
        row 22 col 65: want 'n' fg normal, got ' ' fg normal
        row 22 col 66: want 'l' fg normal, got ' ' fg normal
        row 22 col 67: want 'y' fg normal, got ' ' fg normal
        row 22 col 69: want 'p' fg normal, got ' ' fg normal
        row 22 col 70: want 'a' fg normal, got ' ' fg normal
        row 22 col 71: want 'n' fg normal, got ' ' fg normal
        row 22 col 72: want 'e' fg normal, got ' ' fg normal
        row 22 col 73: want 's' fg normal, got ' ' fg normal
        row 22 col 75: want 'i' fg normal, got ' ' fg normal
        row 22 col 76: want 't' fg normal, got ' ' fg normal
        row 22 col 78: want 's' fg normal, got ' ' fg normal
        row 22 col 79: want 't' fg normal, got ' ' fg normal
        row 22 col 80: want 'a' fg normal, got ' ' fg normal
        row 22 col 81: want 'r' fg normal, got ' ' fg normal
        row 22 col 82: want 't' fg normal, got ' ' fg normal
        row 22 col 83: want 'e' fg normal, got ' ' fg normal
        row 22 col 84: want 'd' fg normal, got ' ' fg normal
        row 22 col 85: want '.' fg normal, got ' ' fg normal
        row 23 col 23: want '•' accent normal, got ' ' fg normal
        row 23 col 24: want ' ' accent normal, got ' ' fg normal
        row 23 col 25: want 'R' fg normal, got ' ' fg normal
        row 23 col 26: want 'e' fg normal, got ' ' fg normal
        row 23 col 27: want 'p' fg normal, got ' ' fg normal
        row 23 col 28: want 'l' fg normal, got ' ' fg normal
        row 23 col 29: want 'y' fg normal, got ' ' fg normal
        row 23 col 31: want 'l' fg normal, got ' ' fg normal
        row 23 col 32: want 'i' fg normal, got ' ' fg normal
        row 23 col 33: want 'k' fg normal, got ' ' fg normal
        row 23 col 34: want 'e' fg normal, got ' ' fg normal
        row 23 col 36: want '1' code normal, got ' ' fg normal
        row 23 col 37: want 'A' code normal, got ' ' fg normal
        row 23 col 38: want ',' code normal, got ' ' fg normal
        row 23 col 39: want ' ' code normal, got ' ' fg normal
        row 23 col 40: want '2' code normal, got ' ' fg normal
        row 23 col 41: want 'A' code normal, got ' ' fg normal
        row 23 col 42: want ',' code normal, got ' ' fg normal
        row 23 col 43: want ' ' code normal, got ' ' fg normal
        row 23 col 44: want '3' code normal, got ' ' fg normal
        row 23 col 45: want 'A' code normal, got ' ' fg normal
        row 23 col 46: want ',' fg normal, got ' ' fg normal
        row 23 col 48: want 'o' fg normal, got ' ' fg normal
        row 23 col 49: want 'r' fg normal, got ' ' fg normal
        row 23 col 51: want 's' code normal, got ' ' fg normal
        row 23 col 52: want 'i' code normal, got ' ' fg normal
        row 23 col 53: want 'm' code normal, got ' ' fg normal
        row 23 col 54: want 'p' code normal, got ' ' fg normal
        row 23 col 55: want 'l' code normal, got ' ' fg normal
        row 23 col 56: want 'i' code normal, got ' ' fg normal
        row 23 col 57: want 'f' code normal, got ' ' fg normal
        row 23 col 58: want 'y' code normal, got ' ' fg normal
        row 23 col 59: want ' ' code normal, got ' ' fg normal
        row 23 col 60: want 'N' code normal, got ' ' fg normal
        row 23 col 61: want '.' fg normal, got ' ' fg normal
        row 24 col 25: want ' ' fg normal, got 'h' code normal
        row 24 col 26: want ' ' fg normal, got 'i' code normal
        row 24 col 27: want ' ' fg normal, got 'v' code normal
        row 24 col 28: want ' ' fg normal, got 'e' code normal
        row 24 col 29: want ' ' fg normal, got ' ' code normal
        row 24 col 30: want ' ' fg normal, got 'a' code normal
        row 24 col 31: want ' ' fg normal, got 'g' code normal
        row 24 col 32: want ' ' fg normal, got 'e' code normal
        row 24 col 33: want ' ' fg normal, got 'n' code normal
        row 24 col 34: want ' ' fg normal, got 't' code normal
        row 24 col 35: want ' ' fg normal, got ' ' code normal
        row 24 col 36: want ' ' fg normal, got 's' code normal
        row 24 col 37: want ' ' fg normal, got 't' code normal
        row 24 col 38: want ' ' fg normal, got 'o' code normal
        row 24 col 39: want ' ' fg normal, got 'p' code normal
        row 24 col 40: want ' ' fg normal, got ' ' code normal
        row 24 col 41: want ' ' fg normal, got 'm' code normal
        row 24 col 42: want ' ' fg normal, got 'a' code normal
        row 24 col 43: want ' ' fg normal, got 'n' code normal
        row 24 col 44: want ' ' fg normal, got 'a' code normal
        row 24 col 45: want ' ' fg normal, got 'g' code normal
        row 24 col 46: want ' ' fg normal, got 'e' code normal
        row 24 col 47: want ' ' fg normal, got 'r' code normal
        row 24 col 48: want ' ' fg normal, got '-' code normal
        row 24 col 49: want ' ' fg normal, got 'h' code normal
        row 24 col 50: want ' ' fg normal, got 'i' code normal
        row 24 col 51: want ' ' fg normal, got 'v' code normal
        row 24 col 52: want ' ' fg normal, got 'e' code normal
        row 25 col 37: want 't' code normal, got 'p' code normal
        row 25 col 38: want 'o' code normal, got 'a' code normal
        row 25 col 39: want 'p' code normal, got 'w' code normal
        row 25 col 40: want ' ' code normal, got 'n' code normal
        row 25 col 41: want 'm' code normal, got ' ' code normal
        row 25 col 42: want 'a' code normal, got '-' code normal
        row 25 col 43: want 'n' code normal, got '-' code normal
        row 25 col 44: want 'a' code normal, got 'r' code normal
        row 25 col 45: want 'g' code normal, got 'o' code normal
        row 25 col 46: want 'e' code normal, got 'l' code normal
        row 25 col 47: want 'r' code normal, got 'e' code normal
        row 25 col 48: want '-' code normal, got ' ' code normal
        row 25 col 49: want 'h' code normal, got 'm' code normal
        row 25 col 50: want 'i' code normal, got 'a' code normal
        row 25 col 51: want 'v' code normal, got 'n' code normal
        row 25 col 52: want 'e' code normal, got 'a' code normal
        row 25 col 53: want ' ' fg normal, got 'g' code normal
        row 25 col 54: want ' ' fg normal, got 'e' code normal
        row 25 col 55: want ' ' fg normal, got 'r' code normal
        row 26 col 25: want 'h' code normal, got ' ' fg normal
        row 26 col 26: want 'i' code normal, got ' ' fg normal
        row 26 col 27: want 'v' code normal, got ' ' fg normal
        row 26 col 28: want 'e' code normal, got ' ' fg normal
        row 26 col 29: want ' ' code normal, got ' ' fg normal
        row 26 col 30: want 'a' code normal, got ' ' fg normal
        row 26 col 31: want 'g' code normal, got ' ' fg normal
        row 26 col 32: want 'e' code normal, got ' ' fg normal
        row 26 col 33: want 'n' code normal, got ' ' fg normal
        row 26 col 34: want 't' code normal, got ' ' fg normal
        row 26 col 35: want ' ' code normal, got ' ' fg normal
        row 26 col 36: want 's' code normal, got ' ' fg normal
        row 26 col 37: want 'p' code normal, got ' ' fg normal
        row 26 col 38: want 'a' code normal, got ' ' fg normal
        row 26 col 39: want 'w' code normal, got ' ' fg normal
        row 26 col 40: want 'n' code normal, got ' ' fg normal
        row 26 col 41: want ' ' code normal, got ' ' fg normal
        row 26 col 42: want '-' code normal, got ' ' fg normal
        row 26 col 43: want '-' code normal, got ' ' fg normal
        row 26 col 44: want 'r' code normal, got ' ' fg normal
        row 26 col 45: want 'o' code normal, got ' ' fg normal
        row 26 col 46: want 'l' code normal, got ' ' fg normal
        row 26 col 47: want 'e' code normal, got ' ' fg normal
        row 26 col 48: want ' ' code normal, got ' ' fg normal
        row 26 col 49: want 'm' code normal, got ' ' fg normal
        row 26 col 50: want 'a' code normal, got ' ' fg normal
        row 26 col 51: want 'n' code normal, got ' ' fg normal
        row 26 col 52: want 'a' code normal, got ' ' fg normal
        row 26 col 53: want 'g' code normal, got ' ' fg normal
        row 26 col 54: want 'e' code normal, got ' ' fg normal
        row 26 col 55: want 'r' code normal, got ' ' fg normal
        row 27 col 2: want ' ' fg normal, got '1' dim normal
        row 27 col 3: want ' ' fg normal, got '4' dim normal
        row 27 col 4: want ' ' fg normal, got ':' dim normal
        row 27 col 5: want ' ' fg normal, got '2' dim normal
        row 27 col 6: want ' ' fg normal, got '2' dim normal
        row 27 col 9: want ' ' fg normal, got 'y' user bold
        row 27 col 10: want ' ' fg normal, got 'o' user bold
        row 27 col 11: want ' ' fg normal, got 'u' user bold
        row 27 col 12: want ' ' fg normal, got ' ' user bold
        row 27 col 13: want ' ' fg normal, got ' ' user bold
        row 27 col 14: want ' ' fg normal, got ' ' user bold
        row 27 col 15: want ' ' fg normal, got ' ' user bold
        row 27 col 16: want ' ' fg normal, got ' ' user bold
        row 27 col 17: want ' ' fg normal, got ' ' user bold
        row 27 col 18: want ' ' fg normal, got ' ' user bold
        row 27 col 19: want ' ' fg normal, got ' ' user bold
        row 27 col 20: want ' ' fg normal, got ' ' user bold
        row 27 col 21: want ' ' fg normal, got ' ' user bold
        row 27 col 23: want ' ' fg normal, got '1' user normal
        row 27 col 24: want ' ' fg normal, got 'A' user normal
        row 27 col 25: want ' ' fg normal, got ',' user normal
        row 27 col 26: want ' ' fg normal, got ' ' user normal
        row 27 col 27: want ' ' fg normal, got '2' user normal
        row 27 col 28: want ' ' fg normal, got 'A' user normal
        row 27 col 29: want ' ' fg normal, got ',' user normal
        row 27 col 30: want ' ' fg normal, got ' ' user normal
        row 27 col 31: want ' ' fg normal, got '3' user normal
        row 27 col 32: want ' ' fg normal, got 'B' user normal
        
        rendered:
        14:02  you           I want to give the user an ability to restart the coordinator or supervisor within the hive    
                             UI.                                                                                            
                                                                                                                            
        14:19  · manager-hi… goal g1-tui-restart created · wave A started · worker t2-worker-1 live                         
                                                                                                                            
        14:21  coordinator   Interrogate, round 2                                                                           
                                                                                                                            
                             1. What happens to the old tab on restart?                                                     
                             Bare hive makes the coordinator tab, but its record has spawned=false. agent stop refuses it.  
                                                                                                                            
                              Choice            │ Answer                            │ Pros             │ Cons               
                             ───────────────────┼───────────────────────────────────┼──────────────────┼─────────────────── 
                              A (REC)           │ Record that bare hive started it  │ No tab pile-up   │ Field change       
                              B                 │ Leave the old tab, open a new one │ Smallest change  │ Dead tabs pile up  
                              C                 │ Reuse the same pane               │ One tab always   │ Must exit harness  
                                                                                                                            
                             • Recommended: A, because hive may close only panes it started.                                
                             • Reply like 1A, 2A, 3A, or simplify N.                                                        
                                                                                                                            
                               hive agent stop manager-hive                                                                 
                               hive agent spawn --role manager                                                              
                                                                                                                            
        14:22  you           1A, 2A, 3B                                                                                     
--- FAIL: TestCompactChatGolden (0.00s)
    chat_test.go:73: 53 cells differ from CompactChat.grid (board row and column):
        row 7 col 58: want 's' code normal, got ' ' fg normal
        row 7 col 59: want 'p' code normal, got ' ' fg normal
        row 7 col 60: want 'a' code normal, got ' ' fg normal
        row 7 col 61: want 'w' code normal, got ' ' fg normal
        row 7 col 62: want 'n' code normal, got ' ' fg normal
        row 7 col 63: want 'e' code normal, got ' ' fg normal
        row 7 col 64: want 'd' code normal, got ' ' fg normal
        row 7 col 65: want '=' code normal, got ' ' fg normal
        row 7 col 66: want 'f' code normal, got ' ' fg normal
        row 7 col 67: want 'a' code normal, got ' ' fg normal
        row 7 col 68: want 'l' code normal, got ' ' fg normal
        row 7 col 69: want 's' code normal, got ' ' fg normal
        row 7 col 70: want 'e' code normal, got ' ' fg normal
        row 7 col 71: want '.' fg normal, got ' ' fg normal
        row 7 col 73: want 'a' code normal, got ' ' fg normal
        row 7 col 74: want 'g' code normal, got ' ' fg normal
        row 7 col 75: want 'e' code normal, got ' ' fg normal
        row 7 col 76: want 'n' code normal, got ' ' fg normal
        row 7 col 77: want 't' code normal, got ' ' fg normal
        row 8 col 3: want 't' code normal, got 'p' code normal
        row 8 col 4: want 'o' code normal, got 'a' code normal
        row 8 col 5: want 'p' code normal, got 'w' code normal
        row 8 col 6: want ' ' fg normal, got 'n' code normal
        row 8 col 7: want 'r' fg normal, got 'e' code normal
        row 8 col 8: want 'e' fg normal, got 'd' code normal
        row 8 col 9: want 'f' fg normal, got '=' code normal
        row 8 col 10: want 'u' fg normal, got 'f' code normal
        row 8 col 11: want 's' fg normal, got 'a' code normal
        row 8 col 12: want 'e' fg normal, got 'l' code normal
        row 8 col 13: want 's' fg normal, got 's' code normal
        row 8 col 14: want ' ' fg normal, got 'e' code normal
        row 8 col 15: want 'i' fg normal, got '.' fg normal
        row 8 col 16: want 't' fg normal, got ' ' fg normal
        row 8 col 17: want '.' fg normal, got 'a' code normal
        row 8 col 18: want ' ' fg normal, got 'g' code normal
        row 8 col 19: want ' ' fg normal, got 'e' code normal
        row 8 col 20: want ' ' fg normal, got 'n' code normal
        row 8 col 21: want ' ' fg normal, got 't' code normal
        row 8 col 22: want ' ' fg normal, got ' ' code normal
        row 8 col 23: want ' ' fg normal, got 's' code normal
        row 8 col 24: want ' ' fg normal, got 't' code normal
        row 8 col 25: want ' ' fg normal, got 'o' code normal
        row 8 col 26: want ' ' fg normal, got 'p' code normal
        row 8 col 28: want ' ' fg normal, got 'r' fg normal
        row 8 col 29: want ' ' fg normal, got 'e' fg normal
        row 8 col 30: want ' ' fg normal, got 'f' fg normal
        row 8 col 31: want ' ' fg normal, got 'u' fg normal
        row 8 col 32: want ' ' fg normal, got 's' fg normal
        row 8 col 33: want ' ' fg normal, got 'e' fg normal
        row 8 col 34: want ' ' fg normal, got 's' fg normal
        row 8 col 36: want ' ' fg normal, got 'i' fg normal
        row 8 col 37: want ' ' fg normal, got 't' fg normal
        row 8 col 38: want ' ' fg normal, got '.' fg normal
        
        rendered:
        14:21 coordinator                                                           
        Interrogate, round 2                                                        
                                                                                    
        Bare hive makes the coordinator tab, but its record has                     
        spawned=false. agent stop refuses it.                                       
                                                                                    
         Choice  │ Answer                           │ Pros            │ Cons        
        ─────────┼──────────────────────────────────┼─────────────────┼─────────────
         A (REC) │ Record that bare hive started it │ No tab pile-up  │ Field change
         B       │ Leave the old tab                │ Smallest change │ Tabs pile up
                                                                                    
        • Recommended: A, hive closes only panes it started.                        
                                                                                    
        14:22 you                                                                   
        1A, 2A, 3B                                                                  
--- FAIL: TestPopupGolden (0.00s)
    popup_test.go:31: 222 cells differ from AgentDetails.grid (board row and column):
        row 24 col 34: want ' ' dim normal, got ' ' fg normal
        row 24 col 35: want ' ' dim normal, got ' ' fg normal
        row 26 col 33: want 'R' accent bold, got ' ' fg normal
        row 26 col 34: want 'E' accent bold, got ' ' fg normal
        row 26 col 35: want 'C' accent bold, got ' ' fg normal
        row 26 col 36: want 'E' accent bold, got ' ' fg normal
        row 26 col 37: want 'N' accent bold, got ' ' fg normal
        row 26 col 38: want 'T' accent bold, got ' ' fg normal
        row 27 col 33: want '1' fg normal, got 'R' accent bold
        row 27 col 34: want '4' fg normal, got 'E' accent bold
        row 27 col 35: want ':' fg normal, got 'C' accent bold
        row 27 col 36: want '0' fg normal, got 'E' accent bold
        row 27 col 37: want '8' fg normal, got 'N' accent bold
        row 27 col 38: want ' ' fg normal, got 'T' accent bold
        row 27 col 39: want 'i' fg normal, got ' ' fg normal
        row 27 col 40: want 'n' fg normal, got ' ' fg normal
        row 27 col 41: want 'b' fg normal, got ' ' fg normal
        row 27 col 42: want 'o' fg normal, got ' ' fg normal
        row 27 col 43: want 'x' fg normal, got ' ' fg normal
        row 27 col 45: want '2' fg normal, got ' ' fg normal
        row 27 col 46: want '8' fg normal, got ' ' fg normal
        row 27 col 48: want 's' fg normal, got ' ' fg normal
        row 27 col 49: want 'e' fg normal, got ' ' fg normal
        row 27 col 50: want 'n' fg normal, got ' ' fg normal
        row 27 col 51: want 't' fg normal, got ' ' fg normal
        row 27 col 53: want 't' fg normal, got ' ' fg normal
        row 27 col 54: want 'o' fg normal, got ' ' fg normal
        row 27 col 56: want 'c' fg normal, got ' ' fg normal
        row 27 col 57: want 'o' fg normal, got ' ' fg normal
        row 27 col 58: want 'o' fg normal, got ' ' fg normal
        row 27 col 59: want 'r' fg normal, got ' ' fg normal
        row 27 col 60: want 'd' fg normal, got ' ' fg normal
        row 27 col 61: want 'i' fg normal, got ' ' fg normal
        row 27 col 62: want 'n' fg normal, got ' ' fg normal
        row 27 col 63: want 'a' fg normal, got ' ' fg normal
        row 27 col 64: want 't' fg normal, got ' ' fg normal
        row 27 col 65: want 'o' fg normal, got ' ' fg normal
        row 27 col 66: want 'r' fg normal, got ' ' fg normal
        row 28 col 33: want '1' dim normal, got '1' fg normal
        row 28 col 34: want '4' dim normal, got '4' fg normal
        row 28 col 35: want ':' dim normal, got ':' fg normal
        row 28 col 36: want '0' dim normal, got '0' fg normal
        row 28 col 37: want '7' dim normal, got '8' fg normal
        row 28 col 38: want ' ' dim normal, got ' ' fg normal
        row 28 col 39: want 'l' dim normal, got 'i' fg normal
        row 28 col 40: want 'a' dim normal, got 'n' fg normal
        row 28 col 41: want 'n' dim normal, got 'b' fg normal
        row 28 col 42: want 'd' dim normal, got 'o' fg normal
        row 28 col 43: want ' ' dim normal, got 'x' fg normal
        row 28 col 44: want '4' dim normal, got ' ' fg normal
        row 28 col 45: want '-' dim normal, got '2' fg normal
        row 28 col 46: want 'a' dim normal, got '8' fg normal
        row 28 col 47: want 'g' dim normal, got ' ' fg normal
        row 28 col 48: want 'e' dim normal, got 's' fg normal
        row 28 col 49: want 'n' dim normal, got 'e' fg normal
        row 28 col 50: want 't' dim normal, got 'n' fg normal
        row 28 col 51: want 's' dim normal, got 't' fg normal
        row 28 col 52: want '-' dim normal, got ' ' fg normal
        row 28 col 53: want 'e' dim normal, got 't' fg normal
        row 28 col 54: want 'q' dim normal, got 'o' fg normal
        row 28 col 55: want 'u' dim normal, got ' ' fg normal
        row 28 col 56: want 'a' dim normal, got 'c' fg normal
        row 28 col 57: want 'l' dim normal, got 'o' fg normal
        row 28 col 58: want '-' dim normal, got 'o' fg normal
        row 28 col 59: want 'c' dim normal, got 'r' fg normal
        row 28 col 60: want 'o' dim normal, got 'd' fg normal
        row 28 col 61: want 'l' dim normal, got 'i' fg normal
        row 28 col 62: want 'u' dim normal, got 'n' fg normal
        row 28 col 63: want 'm' dim normal, got 'a' fg normal
        row 28 col 64: want 'n' dim normal, got 't' fg normal
        row 28 col 65: want 's' dim normal, got 'o' fg normal
        row 28 col 66: want ' ' dim normal, got 'r' fg normal
        row 28 col 67: want '·' dim normal, got ' ' fg normal
        row 28 col 68: want ' ' dim normal, got ' ' fg normal
        row 28 col 69: want 'e' dim normal, got ' ' fg normal
        row 28 col 70: want 'd' dim normal, got ' ' fg normal
        row 28 col 71: want 'e' dim normal, got ' ' fg normal
        row 28 col 72: want '9' dim normal, got ' ' fg normal
        row 28 col 73: want '4' dim normal, got ' ' fg normal
        row 28 col 74: want 'f' dim normal, got ' ' fg normal
        row 28 col 75: want '3' dim normal, got ' ' fg normal
        row 28 col 76: want 'e' dim normal, got ' ' fg normal
        row 29 col 34: want '3' dim normal, got '4' dim normal
        row 29 col 36: want '5' dim normal, got '0' dim normal
        row 29 col 37: want '8' dim normal, got '7' dim normal
        row 29 col 44: want '3' dim normal, got '4' dim normal
        row 29 col 46: want 't' dim normal, got 'a' dim normal
        row 29 col 47: want 'u' dim normal, got 'g' dim normal
        row 29 col 48: want 'i' dim normal, got 'e' dim normal
        row 29 col 49: want '-' dim normal, got 'n' dim normal
        row 29 col 50: want 'r' dim normal, got 't' dim normal
        row 29 col 51: want 'e' dim normal, got 's' dim normal
        row 29 col 52: want 's' dim normal, got '-' dim normal
        row 29 col 53: want 't' dim normal, got 'e' dim normal
        row 29 col 54: want 'a' dim normal, got 'q' dim normal
        row 29 col 55: want 'r' dim normal, got 'u' dim normal
        row 29 col 56: want 't' dim normal, got 'a' dim normal
        row 29 col 57: want ' ' dim normal, got 'l' dim normal
        row 29 col 58: want '·' dim normal, got '-' dim normal
        row 29 col 59: want ' ' dim normal, got 'c' dim normal
        row 29 col 60: want '3' dim normal, got 'o' dim normal
        row 29 col 61: want 'c' dim normal, got 'l' dim normal
        row 29 col 62: want 'c' dim normal, got 'u' dim normal
        row 29 col 63: want 'd' dim normal, got 'm' dim normal
        row 29 col 64: want '9' dim normal, got 'n' dim normal
        row 29 col 65: want '3' dim normal, got 's' dim normal
        row 29 col 66: want 'f' dim normal, got ' ' dim normal
        row 29 col 67: want 'b' dim normal, got '·' dim normal
        row 29 col 68: want ' ' fg normal, got ' ' dim normal
        row 29 col 69: want ' ' fg normal, got 'e' dim normal
        row 29 col 70: want ' ' fg normal, got 'd' dim normal
        row 29 col 71: want ' ' fg normal, got 'e' dim normal
        row 29 col 72: want ' ' fg normal, got '9' dim normal
        row 29 col 73: want ' ' fg normal, got '4' dim normal
        row 29 col 74: want ' ' fg normal, got 'f' dim normal
        row 29 col 75: want ' ' fg normal, got '3' dim normal
        row 29 col 76: want ' ' fg normal, got 'e' dim normal
        row 29 col 77: want ' ' fg normal, got ' ' dim normal
        row 29 col 78: want ' ' fg normal, got ' ' dim normal
        row 29 col 79: want ' ' fg normal, got ' ' dim normal
        row 29 col 80: want ' ' fg normal, got ' ' dim normal
        row 29 col 81: want ' ' fg normal, got ' ' dim normal
        row 29 col 82: want ' ' fg normal, got ' ' dim normal
        row 29 col 83: want ' ' fg normal, got ' ' dim normal
        row 29 col 84: want ' ' fg normal, got ' ' dim normal
        row 29 col 85: want ' ' fg normal, got ' ' dim normal
        row 29 col 86: want ' ' fg normal, got ' ' dim normal
        row 30 col 36: want '4' dim normal, got '5' dim normal
        row 30 col 37: want '4' dim normal, got '8' dim normal
        row 30 col 44: want '2' dim normal, got '3' dim normal
        row 30 col 46: want 's' dim normal, got 't' dim normal
        row 30 col 47: want 't' dim normal, got 'u' dim normal
        row 30 col 48: want 'a' dim normal, got 'i' dim normal
        row 30 col 49: want 'l' dim normal, got '-' dim normal
        row 30 col 50: want 'e' dim normal, got 'r' dim normal
        row 30 col 51: want '-' dim normal, got 'e' dim normal
        row 30 col 52: want 'c' dim normal, got 's' dim normal
        row 30 col 53: want 'o' dim normal, got 't' dim normal
        row 30 col 54: want 'o' dim normal, got 'a' dim normal
        row 30 col 56: want 'd' dim normal, got 't' dim normal
        row 30 col 57: want 'i' dim normal, got ' ' dim normal
        row 30 col 58: want 'n' dim normal, got '·' dim normal
        row 30 col 59: want 'a' dim normal, got ' ' dim normal
        row 30 col 60: want 't' dim normal, got '3' dim normal
        row 30 col 61: want 'o' dim normal, got 'c' dim normal
        row 30 col 62: want 'r' dim normal, got 'c' dim normal
        row 30 col 63: want '-' dim normal, got 'd' dim normal
        row 30 col 64: want 's' dim normal, got '9' dim normal
        row 30 col 65: want 't' dim normal, got '3' dim normal
        row 30 col 66: want 'a' dim normal, got 'f' dim normal
        row 30 col 67: want 'r' dim normal, got 'b' dim normal
        row 30 col 68: want 't' dim normal, got ' ' dim normal
        row 30 col 69: want 'u' dim normal, got ' ' dim normal
        row 30 col 70: want 'p' dim normal, got ' ' dim normal
        row 30 col 72: want '·' dim normal, got ' ' dim normal
        row 30 col 74: want '0' dim normal, got ' ' dim normal
        row 30 col 75: want 'e' dim normal, got ' ' dim normal
        row 30 col 76: want '0' dim normal, got ' ' dim normal
        row 30 col 77: want '7' dim normal, got ' ' dim normal
        row 30 col 78: want '9' dim normal, got ' ' dim normal
        row 30 col 79: want '2' dim normal, got ' ' dim normal
        row 30 col 80: want '7' dim normal, got ' ' dim normal
        row 30 col 81: want '2' dim normal, got ' ' dim normal
        row 30 col 82: want ' ' fg normal, got ' ' dim normal
        row 30 col 83: want ' ' fg normal, got ' ' dim normal
        row 30 col 84: want ' ' fg normal, got ' ' dim normal
        row 30 col 85: want ' ' fg normal, got ' ' dim normal
        row 30 col 86: want ' ' fg normal, got ' ' dim normal
        row 31 col 33: want ' ' fg normal, got '1' dim normal
        row 31 col 34: want ' ' fg normal, got '3' dim normal
        row 31 col 35: want ' ' fg normal, got ':' dim normal
        row 31 col 36: want ' ' fg normal, got '4' dim normal
        row 31 col 37: want ' ' fg normal, got '4' dim normal
        row 31 col 38: want ' ' fg normal, got ' ' dim normal
        row 31 col 39: want ' ' fg normal, got 'l' dim normal
        row 31 col 40: want ' ' fg normal, got 'a' dim normal
        row 31 col 41: want ' ' fg normal, got 'n' dim normal
        row 31 col 42: want ' ' fg normal, got 'd' dim normal
        row 31 col 43: want ' ' fg normal, got ' ' dim normal
        row 31 col 44: want ' ' fg normal, got '2' dim normal
        row 31 col 45: want ' ' fg normal, got '-' dim normal
        row 31 col 46: want ' ' fg normal, got 's' dim normal
        row 31 col 47: want ' ' fg normal, got 't' dim normal
        row 31 col 48: want ' ' fg normal, got 'a' dim normal
        row 31 col 49: want ' ' fg normal, got 'l' dim normal
        row 31 col 50: want ' ' fg normal, got 'e' dim normal
        row 31 col 51: want ' ' fg normal, got '-' dim normal
        row 31 col 52: want ' ' fg normal, got 'c' dim normal
        row 31 col 53: want ' ' fg normal, got 'o' dim normal
        row 31 col 54: want ' ' fg normal, got 'o' dim normal
        row 31 col 55: want ' ' fg normal, got 'r' dim normal
        row 31 col 56: want ' ' fg normal, got 'd' dim normal
        row 31 col 57: want ' ' fg normal, got 'i' dim normal
        row 31 col 58: want ' ' fg normal, got 'n' dim normal
        row 31 col 59: want ' ' fg normal, got 'a' dim normal
        row 31 col 60: want ' ' fg normal, got 't' dim normal
        row 31 col 61: want ' ' fg normal, got 'o' dim normal
        row 31 col 62: want ' ' fg normal, got 'r' dim normal
        row 31 col 63: want ' ' fg normal, got '-' dim normal
        row 31 col 64: want ' ' fg normal, got 's' dim normal
        row 31 col 65: want ' ' fg normal, got 't' dim normal
        row 31 col 66: want ' ' fg normal, got 'a' dim normal
        row 31 col 67: want ' ' fg normal, got 'r' dim normal
        row 31 col 68: want ' ' fg normal, got 't' dim normal
        row 31 col 69: want ' ' fg normal, got 'u' dim normal
        row 31 col 70: want ' ' fg normal, got 'p' dim normal
        row 31 col 71: want ' ' fg normal, got ' ' dim normal
        row 31 col 72: want ' ' fg normal, got '·' dim normal
        row 31 col 73: want ' ' fg normal, got ' ' dim normal
        row 31 col 74: want ' ' fg normal, got '0' dim normal
        row 31 col 75: want ' ' fg normal, got 'e' dim normal
        row 31 col 76: want ' ' fg normal, got '0' dim normal
        row 31 col 77: want ' ' fg normal, got '7' dim normal
        row 31 col 78: want ' ' fg normal, got '9' dim normal
        row 31 col 79: want ' ' fg normal, got '2' dim normal
        row 31 col 80: want ' ' fg normal, got '7' dim normal
        row 31 col 81: want ' ' fg normal, got '2' dim normal
        row 31 col 82: want ' ' fg normal, got ' ' dim normal
        row 31 col 83: want ' ' fg normal, got ' ' dim normal
        row 31 col 84: want ' ' fg normal, got ' ' dim normal
        row 31 col 85: want ' ' fg normal, got ' ' dim normal
        row 31 col 86: want ' ' fg normal, got ' ' dim normal
        
        rendered:
        ╭─ manager-hive · esc closes ──────────────────────────────╮
        │                                                          │
        │  role      manager                                       │
        │  project   hive                                          │
        │  status    ◐ live · idle                                 │
        │  nudged    4m ago by supervisor                          │
        │  harness   claude                                        │
        │  model     claude-opus-5-5                               │
        │  pane      w58:p6                                        │
        │  tab       manager-hive                                  │
        │  goal      g1-tui-restart · done                         │
        │  parent    coordinator-g2                                │
        │  replaces  manager-hive-g1                               │
        │                                                          │
        │  ACTIONS                                                 │
        │  a  attach to pane                                       │
        │  p  peek pane                                            │
        │  t  open task · none                                     │
        │                                                          │
        │                                                          │
        │  RECENT                                                  │
        │  14:08 inbox 28 sent to coordinator                      │
        │  14:07 land 4-agents-equal-columns · ede94f3e            │
        │  13:58 land 3-tui-restart · 3ccd93fb                     │
        │  13:44 land 2-stale-coordinator-startup · 0e079272       │
        │                                                          │
        ╰──────────────────────────────────────────────────────────╯
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui/component	0.023s
FAIL
exit 1

### restored sample inputs
ok  	github.com/Harrison-Blair/hive/internal/tui/component	0.024s
exit 0
```

## V-gate scope

```text
 .hive/spec/tui-v2/design/AgentDetails.dc.html      |   3 +
 .hive/spec/tui-v2/design/AgentDetails.grid         |  22 +-
 .hive/spec/tui-v2/design/AgentsMany.dc.html        |  38 +--
 .hive/spec/tui-v2/design/AgentsMany.grid           |  40 +--
 .hive/spec/tui-v2/design/Chat.dc.html              |   4 +-
 .hive/spec/tui-v2/design/Chat.grid                 |  30 +--
 .hive/spec/tui-v2/design/Compact.dc.html           |   8 +-
 .hive/spec/tui-v2/design/Compact.grid              |  16 +-
 .hive/spec/tui-v2/design/CompactAgents.dc.html     |  12 +-
 .hive/spec/tui-v2/design/CompactAgents.grid        |  54 ++--
 .hive/spec/tui-v2/design/CompactChat.dc.html       |   6 +-
 .hive/spec/tui-v2/design/CompactChat.grid          |  32 +--
 .hive/spec/tui-v2/design/CompactFriction.dc.html   |   4 +-
 .hive/spec/tui-v2/design/CompactFriction.grid      |  58 ++--
 .hive/spec/tui-v2/design/CompactHealth.dc.html     |   4 +-
 .hive/spec/tui-v2/design/CompactHealth.grid        |  20 +-
 .hive/spec/tui-v2/design/CompactLogs.dc.html       |   1 -
 .hive/spec/tui-v2/design/CompactLogs.grid          |  20 +-
 .hive/spec/tui-v2/design/Components.dc.html        |   2 +-
 .hive/spec/tui-v2/design/Components.grid           |   4 +-
 .hive/spec/tui-v2/design/Friction.dc.html          |  11 +-
 .hive/spec/tui-v2/design/Friction.grid             | 112 ++++----
 .hive/spec/tui-v2/design/FrictionResolved.dc.html  |  11 +-
 .hive/spec/tui-v2/design/FrictionResolved.grid     | 130 ++++-----
 .hive/spec/tui-v2/design/Health.dc.html            |   2 +-
 .hive/spec/tui-v2/design/Health.grid               |  64 ++---
 .hive/spec/tui-v2/design/HealthAgents.dc.html      |   4 +-
 .hive/spec/tui-v2/design/HealthAgents.grid         |  16 +-
 .hive/spec/tui-v2/design/Help.dc.html              |   3 +
 .hive/spec/tui-v2/design/Help.grid                 |   4 +-
 .hive/spec/tui-v2/design/Logs.dc.html              |  14 +-
 .hive/spec/tui-v2/design/Logs.grid                 | 154 +++++------
 .hive/spec/tui-v2/design/Main.dc.html              |  20 +-
 .hive/spec/tui-v2/design/Main.grid                 |  54 ++--
 .hive/spec/tui-v2/design/Work.dc.html              |   8 +-
 .hive/spec/tui-v2/design/Work.grid                 |  24 +-
 .hive/spec/tui-v2/design/WorkGoals.dc.html         |   4 +-
 .hive/spec/tui-v2/design/WorkGoals.grid            |   6 +-
 .hive/spec/tui-v2/design/WorkQueue.dc.html         |   4 +-
 .hive/spec/tui-v2/design/WorkQueue.grid            |   4 +-
 .../tui-v2/tasks/42-wave6-board-redraw/brief.md    | 155 +++++++++++
 .../tasks/42-wave6-board-redraw/evidence/V-gate.md | 189 +++++++++++++
 .../tasks/42-wave6-board-redraw/evidence/V1.md     |  96 +++++++
 .../tasks/42-wave6-board-redraw/evidence/V2.md     | 291 +++++++++++++++++++++
 .../tasks/42-wave6-board-redraw/evidence/V3.md     | 209 +++++++++++++++
 internal/tui/component/chat_test.go                |  14 +-
 internal/tui/component/details.go                  |  18 +-
 internal/tui/component/details_test.go             |  31 +++
 internal/tui/component/list.go                     |  13 +-
 internal/tui/component/list_test.go                |  35 +++
 internal/tui/component/popup_test.go               |   9 +-
 internal/tui/frame.go                              |  12 +-
 internal/tui/frame_test.go                         |  21 ++
 testdata/ste-baseline.json                         |   1 +
 54 files changed, 1564 insertions(+), 557 deletions(-)
```

## Full gate output

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
exit 0

$ go mod tidy -diff
exit 0

$ go vet ./...
exit 0

$ golangci-lint run ./...
0 issues.
exit 0

$ go test -race ./...
ok  	github.com/Harrison-Blair/hive	18.914s
ok  	github.com/Harrison-Blair/hive/cmd	6.168s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.344s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.481s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.591s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.557s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.321s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.222s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.268s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.418s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.216s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.276s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.187s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	13.391s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.413s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.837s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	(cached)
ok  	github.com/Harrison-Blair/hive/internal/homeinit	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/land	3.395s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.168s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.273s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/log	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.016s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.246s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/store	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/logs	(cached)
ok  	github.com/Harrison-Blair/hive/internal/memory	(cached)
ok  	github.com/Harrison-Blair/hive/internal/project	(cached)
ok  	github.com/Harrison-Blair/hive/internal/supervise	(cached)
ok  	github.com/Harrison-Blair/hive/internal/task	9.643s
ok  	github.com/Harrison-Blair/hive/internal/tui	52.335s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	18.622s
ok  	github.com/Harrison-Blair/hive/internal/worktree	4.824s
exit 0

```

## Cleanup

All temporary mutations were restored. Scratch profiles, PNGs, programs, old grids, and logs in `.hive/tmp/verify42/` were removed after this record was written. No source changes remain.
