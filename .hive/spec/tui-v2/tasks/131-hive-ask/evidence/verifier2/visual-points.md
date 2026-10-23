# Visual points

The verifier captured each board and each screen again. The verifier inspected
all twelve board PNG files and the six exact-size screen PNG files.
No board capture shows `{{`.

The final screen files are `sized-exact-BOARD.ansi`, `.txt`, `.png`,
`.meta.json`, and `.diff.txt`. Each diff is empty. The comparison checks the
rune, the colour token, and the bold flag. It uses no mask.

Rows below are 1-based. Columns are 0-based. The expected positions come
from the approved boards of decision 104.

| Screen | Point | Expected and observed | Result |
| --- | --- | --- | --- |
| ChatAsk 120×40 | Dock and input | Dock rows 29–36. Input starts at row 37. | pass |
| ChatAsk 120×40 | Title, option mark, current row | Columns 3, 6, and 2. | pass |
| ChatAsk 120×40 | Header count | `? 1 ask` starts at column 90. | pass |
| ChatAskMulti 120×40 | Dock and input | Dock rows 21–36. Input starts at row 37. | pass |
| ChatAskMulti 120×40 | Height cap | The dock has 16 rows. The cap is 20 rows. | pass |
| ChatAskMulti 120×40 | Title, option mark, current row | Columns 3, 6, and 2. | pass |
| ChatAskMulti 120×40 | Header count | `? 2 asks` starts at column 89. | pass |
| ChatAskMulti 120×40 | Editing line | `note ›` and the cursor are on row 35. The footer shows `[note]`. | pass |
| ChatAskPreview 120×40 | Dock and input | Dock rows 26–36. Input starts at row 37. | pass |
| ChatAskPreview 120×40 | Title, option mark, current row | Columns 3, 6, and 2. | pass |
| ChatAskPreview 120×40 | Preview box | The top-left corner is at row 29, column 44. | pass |
| ChatAskPreview 120×40 | Header count | `? 1 ask` starts at column 90. | pass |
| ChatAskStopped 120×40 | Dock and input | Dock rows 30–36. Input starts at row 37. | pass |
| ChatAskStopped 120×40 | Title, option mark, current row | Columns 3, 6, and 2. | pass |
| ChatAskStopped 120×40 | Stopped asker | The title names `manager-hive` as the answer target. The feed keeps the earlier answer. | pass |
| ChatAskStopped 120×40 | Header count | `? 1 ask` starts at column 90. | pass |
| CompactChatAsk 80×24 | Dock and input | Dock rows 13–20. Input starts at row 21. | pass |
| CompactChatAsk 80×24 | Height cap | The dock has 8 rows. The cap is 12 rows. | pass |
| CompactChatAsk 80×24 | Title, option mark, current row | Columns 3, 6, and 2. | pass |
| CompactChatAsk 80×24 | Header count | `? 1` starts at column 54. | pass |
| Help 120×40 | Added ask keys | The ask-key line starts at row 34, column 21. | pass |
| Each ask screen | Borders and corners | Borders end at columns 0 and W−1. Each dock sits between the feed and the input. | pass |

## States without colour

The `.txt` captures remove ANSI codes. Each state below remains visible.
Each row names one board element.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| ChatAsk header | `? 1 ask` | matches | Zero differing cells. |
| ChatAsk feed | `? asked`, `open below` | matches | The open state has words. |
| ChatAsk dock title | `? manager-hive asks`, `open`, `1 of 1` | matches | The title names the asker. |
| ChatAsk options | `›`, `[ ]`, `(REC)`, `Other` | matches | The current row has a glyph. |
| ChatAsk input and footer | `message`, `[ask]` | matches | The footer names the focus. |
| ChatAskMulti header | `? 2 asks` | matches | The count has a number and a word. |
| ChatAskMulti feed | `open below`, `waits`, `2 of 2` | matches | Each open card has a text state. |
| ChatAskMulti dock title | `Scope`, `Order`, `1 of 2`, `J next` | matches | The title names both questions. |
| ChatAskMulti options | `[x]`, `[ ]`, `›` | matches | Marks show each selection. |
| ChatAskMulti editing line | `note ›`, `█`, `[note]` | matches | The line and footer name the editing state. |
| ChatAskMulti input | `message` | matches | The input stays below the dock. |
| ChatAskPreview header and feed | `? 1 ask`, `open below` | matches | Zero differing cells. |
| ChatAskPreview dock and options | `Layout`, `›`, `[ ]` | matches | Zero differing cells. |
| ChatAskPreview preview box | Border, `preview · 2` | matches | The box names the current preview. |
| ChatAskPreview input and footer | `message`, `[ask]` | matches | The footer names the focus. |
| ChatAskStopped header and history | `? 1 ask`, `answered`, `→` | matches | The previous answer stays in the feed. |
| ChatAskStopped routing title | `asker stopped`, `answer goes to manager-hive` | matches | Words name the state and the target. |
| ChatAskStopped options and footer | `[ ]`, `›`, `send to parent`, `[ask]` | matches | The footer names the answer route. |
| CompactChatAsk header | `? 1` | matches | Decision 104 approves this compact form. |
| CompactChatAsk feed | `? asked`, `below` | matches | The open state has words. |
| CompactChatAsk dock | Heavy border, `Model`, `1 of 1` | matches | The card uses the full width. |
| CompactChatAsk options | `[ ]`, `›`, `(REC)`, `Other` | matches | Marks do not depend on colour. |
| CompactChatAsk input and footer | `message`, `[ask]` | matches | The footer names the focus. |
| Help title and footer | `keys`, `? or esc closes`, `[help]` | matches | Words name the overlay. |
| Help ask-key line | `ask card`, `space mark`, `tab note`, `x dismiss`, `J/K card` | matches | Words describe each key. |
| Original board rule slips | The state words and glyphs remain. | overridden | Decision 104 approves the 13 cases in `../board-rule-cases.md`. |

## Capture conditions

The final run used an isolated herdr configuration and a connected frontend.
`stty size` gave `40 120` and `24 80` for the exact-size captures.
The same build also passed in panes of 121×41 and 81×25, with top-left crops.
See `sized-grids.log`.

The first run used a headless pane. Some reads omitted the styles of trailing
blank cells. Those preliminary captures and diffs remain in this folder.
The final `sized-exact-*` and `sized-larger-*` captures replace those runs.
No final correct-grid diff has a differing cell.

The wrong-grid control has 1526 differing cells. Its file is
`sized-wrong-grid.diff.txt`. That control does not compare a screen with its
own board.
