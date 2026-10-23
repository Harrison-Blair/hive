# Chat

Chat is the home screen. It shows the transcript, the composer, and the status rows. The composer always has the focus. The rules of `.hive/spec/tui-v2/README.md` decisions 28 to 33, 55 to 57, 86, 87, and 91 to 119 stay, except where a rule below replaces them.

## Layout at 120×40

| Rows | Region |
| --- | --- |
| 0 | `· · · N earlier · · ·` in `dim` from column 52, when N is more than 0. Else the first transcript row. |
| 1 to the row above the notice line | The transcript. Column 119 holds the `┃` thumb in `dim`. |
| one row, only while a notice exists | The notice line. |
| 3 to 19 rows | The composer frame. |
| the last 1 to 3 rows | The status rows. |

## Transcript

1. The transcript takes the full width at every size. There is no gutter. Message text starts at column 1 and wraps at column 118.
2. Each message starts with one line: the marker, the name in bold, and the time in `dim`. The user line is `› you HH:MM`. An agent line is `● NAME HH:MM`. The text of the message starts on the next row. One blank row separates two messages.
3. A folded event is one row: `▸ NAME HH:MM  TEXT` in `dim`, cut with `…` at column 118. An open event starts with `▾` and shows its full text on the next rows, wrapped. `ctrl+o` opens or folds all events. A click on an event opens or folds that event. The TUI folds all events at start. Two folded events have no blank row between them.
4. An agent message renders as markdown with the Glamour style of `tui-v2` decision 29. The wrap width is 118 cells. A table follows `tui-v2` decision 32 at that width.
5. The user line shows `· pending` in `dim` after the time while the delivery is `pending`. It shows `· not delivered` in `accent` when the delivery is `unknown`.
6. An ask line is `? ASKER HH:MM  asked · HEADERS · in the composer, ^r to answer` for the card that has the focus. For another open card, it is `? ASKER HH:MM  asked · HEADERS · waits · I of N`. When the asker stopped, `· asker stopped` follows the headers. When the ask record carries `last_error`, the line ends with `· delivery failed · retrying` in `accent`. After an answer, the line from `you` is `answered ASKER · HEADER → PICKS`. After a dismiss, it is `dismissed ASKER · HEADERS`.
7. The transcript follows its end. The wheel, `pgup`, and `pgdn` scroll it. A wheel step moves 3 rows. The `┃` thumb on column 119 shows the position. When the transcript is shorter than its rows, the thumb is not drawn.
8. While the transcript has no message, no event, and no ask, the welcome block shows from row 1. The block is in `dim`. Its lines are `hive · HOME`, the coordinator state in the form of README decision 11, `type a message to the coordinator.`, the chords of README decision 11, and `? help`. The block goes away with the first message, event, or ask.

## Composer

9. The composer is a rounded frame at the full width above the status rows. The frame has the `accent` border while the input line has the focus. The input line shows `> ` in `user` and the typed text in `user`. A second text row starts with two spaces.
10. The empty input shows a placeholder in `dim` after `> `. The placeholder is `type a message · ? help` while the coordinator is `working` or `idle`. It is `type a message · the coordinator is blocked` while the coordinator is `blocked`. It is `coordinator starting · a message waits here` while the coordinator is `starting`. It is `no coordinator · ^y health, r restarts it` while no live coordinator exists.
11. Text wraps at spaces. The frame grows by one row for each text row, up to one third of the content rows, rounded down. Content rows are the rows above the status rows. Beyond that, the text scrolls inside the frame, the cursor row stays visible, and the `┃` thumb is on the right border.
12. `enter` sends the text to the coordinator with `hive chat send`. `shift+enter` adds a line. When no live coordinator exists, `enter` shows `error: no live coordinator` on the notice line and keeps the text. When the send fails, `enter` shows `error: TEXT` and keeps the text.
13. `?` on an empty input line opens Help. On a line with text, `?` types `?`.

## Ask card

14. An open ask draws its card inside the composer frame, above the input line. The frame title is `? ASKER asks · HEADERS`. The right side of the top border shows `open · I of N`, and `· J next` when a next card exists. When the asker stopped, the right side shows `asker stopped · answer goes to PARENT` in `accent` bold. A rule row `├──┤` separates the card from the input line.
15. The card body follows `tui-v2` decisions 96 and 113 to 116. It has the question, the option rows, the `[x]` marks, the `›` row marker, the `note ›` line, the preview, and the review. The `note ›` line is below the options. It shows `none · n adds one` in `dim` while the `note` text is empty.
16. The frame is at most half of the content rows, rounded down. The input line keeps its grown rows, up to one third. The card gets the rest. When the card has more rows, it scrolls and keeps the current row in view. The `┃` thumb is on the frame border beside the card rows. The preview box shows only when the card has 8 rows or more. Else the preview text replaces the description of the current option.
17. `ctrl+r` moves the focus between the input line and the card, both ways. Typed text stays. When the card has the focus, the `›` row marker is `accent` bold and the input line has no cursor. When the input line has the focus, the `›` row marker is `dim` and the cursor is on the input line.
18. On the card, `↑` and `↓` move, and `1` to `9` pick. `space` marks an option of a multi-select question. `enter` opens the review, or sends from the review. `n` opens the `note` line. `x` dismisses. `J` and `K` change the card. `esc` gives the focus to the input line. `tab` does nothing on the card.
19. The oldest open card has the focus first, as `tui-v2` decision 97 gives. When the last open card closes, the frame returns to the composer alone and the input line keeps its text.

## States

| State | Trigger | What shows | Board |
| --- | --- | --- | --- |
| working | the coordinator is `working` | `● working Nm` on the status row, placeholder `type a message · ? help` | `GfChat` |
| idle | the coordinator is `idle` | `○ idle Nm` | no board |
| blocked | the coordinator is `blocked` | `▲ blocked Nm`, placeholder of rule 10 | no board |
| starting | bare `hive` starts the coordinator | `● starting`, placeholder of rule 10, the startup notice on the notice line | no board |
| no coordinator | no live coordinator record | `○ none`, placeholder of rule 10, `error: no live coordinator` after `enter` | `GfChatNoCoordinator` |
| empty | no message, event, or ask | the welcome block of rule 8 | `GfChatEmpty` |
| pending | the last user message has delivery `pending` | `· pending` on the user line | `GfChat` |
| not delivered | the last user message has delivery `unknown` | `· not delivered` on the user line | no board |
| events open | `ctrl+o` | each event open with `▾` | `GfChatEvents` |
| tall input | the text has more rows than one third | the frame at the cap, the `┃` thumb | `GfChatTall` |
| markdown | an agent reply has a table and a code block | the Glamour table and code block at 118 cells | `GfChatMarkdown` |
| notice | a notice exists | the notice line above the composer | `GfChatNotice` |
| ask, card focus | one open ask, the card has the focus | rules 14 to 18 | `GfChatAsk` |
| ask, input focus | one open ask, the input line has the focus | rule 17, text on the input line | `GfChatAskInput` |
| ask, two questions | an ask with two questions, multi-select, and a `note` | `tui-v2` decision 96 inside the frame | `GfChatAskMulti` |
| ask, preview | an ask with a `preview` | the preview box of rule 16 | `GfChatAskPreview` |
| ask, asker stopped | the asker is not live | the right text of rule 14, `Send to PARENT` in the review | `GfChatAskStopped` |
| ask, review | `enter` on the last question | `tui-v2` decision 114 inside the frame | `GfChatAskReview` |
| ask, several | three open asks | `open · 1 of 3 · J next`, two `waits` lines in the transcript | `GfChatAskMany` |
| ask, delivery failed | the ask record carries `last_error` | the line of rule 6 | no board |
| ask, answered | the user sent an answer | the `answered` line of rule 6, the frame without the card | no board |

## Keys

| Focus | Key | Action |
| --- | --- | --- |
| input line | printable | types |
| input line | `enter` | send |
| input line | `shift+enter` | new line |
| input line | `ctrl+o` | open or fold all events |
| input line | `ctrl+r` | focus the ask card, when a card is open |
| input line | `?` | Help, when the line is empty |
| input line | `pgup`, `pgdn`, wheel | scroll the transcript |
| input line | `esc` | nothing |
| ask card | `↑`, `↓` | move |
| ask card | `1` to `9` | pick |
| ask card | `space` | mark |
| ask card | `enter` | review, then send |
| ask card | `n` | the `note` line |
| ask card | `x` | dismiss |
| ask card | `J`, `K` | next or previous card |
| ask card | `ctrl+r`, `esc` | focus the input line |
| any | chords of README decision 16 | open a screen |
