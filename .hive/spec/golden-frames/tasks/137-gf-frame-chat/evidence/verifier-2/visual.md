# Visual check, second verifier

Date: 2026-10-24. Agent: `t137-verifier-2`. Code: `a4d99be8`.

## Captures

The verifier captured all 14 Chat fixtures again in herdr 0.9.5, protocol 22.
The scratch session was `t137-check-verifier-2`.
The pane name was `t137-check-verifier-2-live`.
The pane was 120×40, as `live/session.txt` records.

Each demo used the task binary with `COLORTERM=truecolor`, `TERM=vt100`, and `--size 120x40`.
A scratch home held its marker and logs.
Each capture used `pane.read` with `source: visible` and `format: ansi`.
Two reads 500 ms apart had to agree.
The driver refused truncation and a scroll offset other than zero.
The driver stopped and deleted its session after the checks.

`live/` holds the ANSI captures, text without colour, converted grids, and cell differences.
All 14 grids have zero differing cells with no mask.
A deliberate change to one captured letter made the comparison fail.
`live-negative.txt` records that failure.
The restored capture passes in `live-compare.txt`.

The verifier captured 68 PNGs with fresh Firefox profiles:

- `png/pinned/`: 15 pinned boards.
- `png/revised/`: 15 current board copies.
- `png/screen/`: 14 live Chat states.
- `png/keys/`: 24 key results.

The pinned files refer to `support.js`. That file is absent.
The board adapter exports the board scripts, then renders their cells as HTML for Firefox.
It keeps the character, palette token, and bold flag of each cell.
The screen adapter renders the converted live cells in the same way.
Each live Chat PNG equals its revised board PNG byte for byte.
`png-compare.txt` records this comparison.
`screenshots.txt` records the Firefox commands and fresh profiles.

`probe-captures/` has two more PNGs from model output.
They show the long ask before and after the resize to 80×24.
The top-right corner now stays visible.
These two PNGs are model captures, separate from the live fixture captures.

## Elements

The verifier inspected the screen PNGs, board differences, and text without colour.
`colour-check.txt` records the plain-text state signals of each fixture.
No checked state depends on colour alone.

`matches` refers to the revised board.
`overridden` names a recorded decision that changes the pinned board.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| User message | `› you`, time, and text | matches | Chat rule 2 |
| Agent message | `● coordinator`, time, and text | matches | Chat rules 2 and 4 |
| Pending delivery | `· pending` | matches | Chat rule 5 |
| Unknown delivery | `· not delivered` | overridden | C16 gives the accent token |
| Folded event | `▸`, name, and time | matches | Chat rule 3 |
| Open event | `▾`, name, time, and text below | matches | Chat rule 3 |
| Transcript start | The first message starts on row 0 | overridden | C11 removes the two empty rows |
| Earlier count and transcript thumb | Neither shows when the content fits | overridden | C10 removes the false scroll signals |
| Markdown headings and lists | Heading words and `•` items | matches | Chat rule 4 |
| Markdown table | Header words, `│`, and `─┼─` | overridden | C12 gives the Glamour table at 118 cells |
| Markdown code and margins | Code text and indentation | overridden | C5, C12, and C13 give the wrap and tokens |
| Empty welcome | `hive`, home, state, and help words | overridden | C3 gives `● working 2m` |
| Composer | `╭`, `│`, `╰`, `>`, and typed text | matches | Chat rules 9 to 11 |
| Tall composer | Ten text rows with no thumb | overridden | C2 removes the false thumb |
| Error notice | `error: no live coordinator` | overridden | C4 gives the text of rule 12 |
| Quit notice | `press ctrl+c again to quit` | matches | Decision 14 |
| Card title | `? ASKER asks`, headers, and `open · I of N` | matches | Chat rule 14 |
| Current option | `›` | matches | Chat rules 15 and 17 |
| Picked option | `[x]` | matches | Chat rule 15 |
| Empty note | `note › none · n adds one` | matches | Chat rule 15 |
| Note editor | `note ›`, text, and cursor | matches | Chat rule 15 and decision 33 |
| Note editor keys | `note`, `type the note`, and `enter keep note` | overridden | C14 removes `n close note` |
| Input focus with an ask | Cursor and `^r answer the ask` | matches | Decision 33 |
| Preview | `preview · 1` and preview text | overridden | C9 removes the two extra card rows |
| Stopped asker | `asker stopped · answer goes to manager-hive` | matches | Chat rule 14 |
| Review | `confirm`, answer text, and `› Send` | matches | Chat rule 15 |
| Several asks | `waits · I of N`, `J next`, and `? 3 asks` | matches | Chat rules 6 and 14 |
| Status state | `● working`, `○ idle`, or `○ none` | matches | Decision 11 |
| Status segments | Words, ` · ` joins, and right clock | matches | Decisions 8 to 11 |
| Status focus | `ask`, `note`, or `review`, with its keys | matches | Decision 33 |
| Component status examples | `? 1 ask`, `○ idle 5m`, and `▲ blocked 1m` | overridden | C7 corrects the forms and segment removal |
| Component title and composer | `work`, rule, `>`, and card words | matches | In-scope frame and Chat regions |
| Component event and answer | `▾`, `› you`, and `answered manager-hive` | overridden | C7 and C15 correct these rows |
| Other screen titles | `agents`, `work`, `friction`, `health`, or `logs` | matches | The interim frame keeps each body |
| Help overlay | `keys`, key names, and action words | matches | Existing layout with the new keys and dim background |
| Long ask after resize | A complete border with `╮` | matches | F1 is fixed under decision 5 |

The wave C regions of GfComponents, rows 47 to 65, remain out of scope.
The user must review the changed grids before landing, as AC10 requires.
The existing `../board-review.html` gives the side-by-side view.

## Live keys

`keys-check.txt` records the assertions against fresh captures.

- Each chord opened its screen.
- Repeated Agents and Logs chords returned to Chat.
- `esc` returned from Work, Friction, and Health.
- `ctrl+]` did nothing on Chat.
- `ctrl+o` opened and folded the events.
- `ctrl+r` moved between the card and input.
- `esc` returned the card focus to the input.
- `?` on empty input opened Help. `esc` closed Help.
- The first `ctrl+c` showed the quit notice. The second ended the app.
