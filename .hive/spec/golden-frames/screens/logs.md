# Logs

Logs shows the log events of one day with a filter, a density strip, the event list, and a field panel. The rules of `tui-v2` decisions 20 and 25 stay.

## Layout at 120×40

| Rows | Region |
| --- | --- |
| 0 | title row: `logs` in `accent` bold, `· N events · W warn · E error · PATH` in `dim`, right text `← DATE → day` |
| 2 | the filter line: `/` in `accent` bold, the filter text, `follow ●`, and `f toggles follow` |
| 4 to 6 | the density strip, its base rule, and the marker text |
| 7 to 37 | the event list on the left, 66 cells wide, and the field panel on the right |
| the last rows | the status rows |

## Rules

1. The density strip has one cell for each bucket of the day range. The range starts at the first event and ends at the last event that the filter keeps. The bucket width is the range divided by 104 cells. A cell is `▁` to `▇` by the event count of its bucket, relative to the largest bucket. A cell with a `WARN` event is `accent`. A cell with an `ERROR` event is `error`. Each other cell is `dim`. The start time and the end time sit at the ends of the strip.
2. The base rule under the strip shows `┴` under the cell of the selected event. The row under it shows `HH:MM LEVEL · MSG` of the selected event, in the color of its level.
3. The event list has the columns `TIME`, `LEVEL`, `COMP`, and `MESSAGE`. `DEBUG` rows are `dim`. `INFO` rows are `fg`. `WARN` rows have the level in `accent` bold. `ERROR` rows have the level in `error` bold. The selected row starts with `▶` and is `accent` bold.
4. The field panel shows the selected event as label and value rows. Nested objects indent by two cells. The value `fail` and the value `WARN` are `accent`. The value `ERROR` is `error`. When the event has `truncated` true, the panel ends with `body cut at 4 KiB` in `accent`.
5. `j` shows the raw JSON of the event in the field panel, with the title `event HH:MM:SS · raw`. `j` again shows the fields.
6. The list and the field panel are the two boxes of README decision 6. `ctrl+]` moves the focus.
7. `/` opens the filter line. The typed fields of `tui-v2` decision 25 apply. `enter` applies the filter. A field with a bad name shows `error: unknown field NAME` on the notice line and keeps the line open. With no match, the list shows `no event matches TEXT` in `dim`, and the strip is empty.
8. `f` toggles follow. While follow is on, the list keeps the newest event selected and the strip extends. `←` and `→` change the day. A day with no log file shows `no log file for DATE` in `dim` in the list, an empty strip, and an empty field panel.
9. `r` applies the filter `req:` with the `request_id` of the selected event. `g` and `G` select the first and the last event.

## States

| State | Trigger | What shows | Board |
| --- | --- | --- | --- |
| list | a day with events | rules 1 to 4 | `GfLogs` |
| no file | a day with no log file | rule 8 | `GfLogsEmpty` |
| no match | a filter with no match | rule 7 | no board |
| bad field | a filter with a bad field name | the error of rule 7 | no board |
| raw JSON | `j` | rule 5 | `GfLogsRaw` |
| error rows | the day has `ERROR` events, one truncated | rules 1, 3, and 4 | `GfLogsError` |
| follow off | `f` | `follow ○` on the filter line | no board |
| panel focus | `ctrl+]` | the field panel has the `accent` border | no board |

## Keys

| Focus | Key | Action |
| --- | --- | --- |
| list | `↑`, `↓`, `g`, `G` | select |
| list, panel | `ctrl+]` | move the focus to the other box |
| panel | `↑`, `↓`, `pgup`, `pgdn` | scroll |
| any | `/` | filter |
| any | `f` | follow |
| any | `←`, `→` | day |
| any | `j` | raw JSON, or fields |
| any | `r` | filter by the request of the selected event |
| any | `esc` | Chat, when no filter line is open |
