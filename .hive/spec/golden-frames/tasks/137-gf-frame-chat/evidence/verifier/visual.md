# Visual check

Date: 2026-10-24. Agent: `t137-verifier-1`. Code: `70de1640`.

## Captures

The verifier captured the 14 Chat states again in herdr 0.9.5, protocol 22.
The scratch session was `t137-check-verifier`.
The pane name was `t137-check-verifier-live`.
The pane had more than 120 columns and 40 rows.
`live/session.txt` gives its layout.

Each demo used the task binary, `TERM=vt100`, `COLORTERM=truecolor`, and `--size 120x40`.
A scratch home held the marker and the demo logs.
Each capture came from `pane.read`, with `source: visible` and `format: ansi`.
Two reads 500 ms apart had to agree.
The driver refused truncation and a scroll offset other than zero.

`live/` holds the ANSI captures, the text without colour, and the converted grids.
`live/boards.diff.txt` gives zero differing cells for each board, with no mask.
A changed letter in a scratch capture made the comparison fail at row 1, column 1.
`live-negative.txt` records that check.

The PNG folders are:

- `png/pinned/`: the 15 pinned boards, before the recorded changes.
- `png/revised/`: the 15 board copies at `70de1640`.
- `png/screen/`: the 14 live Chat captures.
- `png/keys/`: the captures after the keys.

The board files need `support.js`, which the pinned folder does not contain.
The verifier exported their scripts and rendered the cells as HTML for Firefox.
This adapter keeps each character, palette token, and bold flag.
The screen adapter uses the converted live cells.
Each revised Chat board PNG equals its live screen PNG byte for byte.
The PNGs use JetBrains Mono at 13 px, a 17 px line height, and 24 px padding.
`screenshots.txt` records each Firefox command and its new scratch profile.

## Board results

| Board | Live cells that differ | PNG |
| --- | --- | --- |
| GfChat | 0 | `png/screen/GfChat-120x40.png` |
| GfChatEmpty | 0 | `png/screen/GfChatEmpty-120x40.png` |
| GfChatNoCoordinator | 0 | `png/screen/GfChatNoCoordinator-120x40.png` |
| GfChatEvents | 0 | `png/screen/GfChatEvents-120x40.png` |
| GfChatTall | 0 | `png/screen/GfChatTall-120x40.png` |
| GfChatMarkdown | 0 | `png/screen/GfChatMarkdown-120x40.png` |
| GfChatNotice | 0 | `png/screen/GfChatNotice-120x40.png` |
| GfChatAsk | 0 | `png/screen/GfChatAsk-120x40.png` |
| GfChatAskInput | 0 | `png/screen/GfChatAskInput-120x40.png` |
| GfChatAskMulti | 0 | `png/screen/GfChatAskMulti-120x40.png` |
| GfChatAskPreview | 0 | `png/screen/GfChatAskPreview-120x40.png` |
| GfChatAskStopped | 0 | `png/screen/GfChatAskStopped-120x40.png` |
| GfChatAskReview | 0 | `png/screen/GfChatAskReview-120x40.png` |
| GfChatAskMany | 0 | `png/screen/GfChatAskMany-120x40.png` |

## Elements

The verifier inspected the PNGs and the text without colour.
`colour-check.txt` records a plain-text state signal for each Chat board.
No checked state depends on colour alone.

`matches` below refers to the revised board.
`overridden` names a recorded change from the pinned board.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| User message | `› you`, time, and text | matches | Chat rule 2 |
| Agent message | `● coordinator`, time, and text | matches | Chat rules 2 and 4 |
| Pending delivery | `· pending` | matches | Chat rule 5 |
| Unknown delivery | `· not delivered` | overridden | C16 gives the accent token |
| Folded event | `▸`, name, and time | matches | Chat rule 3 |
| Open event | `▾`, name, time, and full text below | matches | Chat rule 3 |
| Transcript start | The first message starts on row 0 | overridden | C11 removes the two empty rows |
| Earlier count and transcript thumb | No earlier count or thumb for content that fits | overridden | C10 removes the false scroll signals |
| Markdown headings and lists | Heading words and `•` items | matches | Chat rule 4 |
| Markdown table | Header words, `│`, and `─┼─` | overridden | C12 gives the Glamour table at 118 cells |
| Markdown code and margins | Code text and indentation | overridden | C5 and C13, then C12, give the wrap and tokens |
| Empty welcome | `hive`, home, state, and help words | overridden | C3 gives `● working 2m` |
| Composer | `╭`, `│`, `╰`, `>`, and typed text | matches | Chat rules 9 to 11 |
| Tall composer | Ten text rows with no scroll thumb | overridden | C2 removes the false thumb |
| Error notice | `error: no live coordinator` | overridden | C4 gives the error text of rule 12 |
| Quit notice | `press ctrl+c again to quit` | matches | Decision 14 |
| Card title | `? ASKER asks`, headers, and `open · I of N` | matches | Chat rule 14 |
| Current option | `›` | matches | Chat rules 15 and 17 |
| Picked option | `[x]` | matches | Chat rule 15 |
| Empty `note` line | `note › none · n adds one` | matches | Chat rule 15 |
| `note` editor | `note ›`, text, and cursor | matches | Chat rule 15 and decision 33 |
| `note` editor keys | `note`, `type the note`, and `enter keep note` | overridden | C14 removes `n close note` |
| Input focus with an open ask | Cursor and `^r answer the ask` | matches | Decision 33 |
| Preview | `preview · 1` and preview text | overridden | C9 removes the two extra card rows |
| Stopped asker | `asker stopped · answer goes to manager-hive` | matches | Chat rule 14 |
| Review | `confirm`, answer text, and `› Send` | matches | Chat rule 15 |
| Several asks | `waits · I of N`, `J next`, and `? 3 asks` | matches | Chat rules 6 and 14 |
| Status state | `● working`, `○ idle`, or `○ none` | matches | Decision 11 |
| Status segments | Words, ` · ` joins, and right clock | matches | Decisions 8 to 11 |
| Status focus | `ask`, `note`, or `review`, with its keys | matches | Decision 33 |
| Components status examples | `? 1 ask`, `○ idle 5m`, and `▲ blocked 1m` | overridden | C7 corrects the forms and removal from the right |
| Components title and composer | `work`, rule, `>`, and card words | matches | Frame and Chat component regions |
| Components event and answer | `▾`, `› you`, and `answered manager-hive` | overridden | C7 and C15 correct the event and answer rows |
| Other screen title rows | `agents`, `work`, `friction`, `health`, or `logs` | matches | Interim frame of this task, no body restyle |
| Help overlay | `keys`, key names, and action words | matches | Existing layout with the new key text |
| Long ask after a resize | The top border loses `╮` | differs | Finding F1 in `../VERIFIED.md` |

The wave C component rows, 47 to 65 of GfComponents, remain out of scope.
This task does not claim their implementation.

## Keys and cleanup

`keys-check.txt` records the live key checks.
Each chord opened its screen.
The repeated Agents and Logs chords returned to Chat.
`esc` returned from Work, Friction, and Health.
`ctrl+]` did nothing on Chat.
`ctrl+o` opened and folded the events.
`ctrl+r` moved between the card and input.
`?` on empty input opened Help.
`esc` closed Help and returned the input focus from the card.
The first `ctrl+c` showed the notice.
The second `ctrl+c` ended the app.

The verifier stopped and deleted the scratch herdr session.
No scratch pane remains.
The long-card PNGs in `probe-captures/` come from the model, without colour.
They show the resize failure separately from the 14 passing live fixtures.
