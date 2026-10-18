# T08 TUI log view

| Item | Value |
| --- | --- |
| Task ID | T08-tui-log-view |
| Wave | 3 |
| Tier | decent |
| Depends on | T04, T06 |
| Blocks | CP4 |
| Branch | `task/T08-tui-log-view` |
| Worktree | `projects/hive/worktrees/T08-tui-log-view/` |
| Evidence | `evidence/T08-tui-log-view/` |

## Overview

Bare `hive` opens the TUI. Its first increment is a live view of the log with a
header that shows the state of the home. The TUI is a client of the same files
and records as the CLI. It has no logic of its own. It reuses the reader and
the filter of T06 and the record readers of T04. The root command passes the
resolved home to the TUI, so `--home` selects the home that the TUI shows.

## User flows

1. The user runs `hive` in a pane of the hive session. The header shows `hive /home/user/hive  supervisor: down  agents: 0  inbox: 0 unread`. The body shows the last events and follows new ones.
2. The user presses `/`, types `herdr`, and presses Enter. The body shows only lines that contain `herdr`.
3. The user presses `l` three times and sees the level filter move from debug to info, warn, and error.
4. The user presses `f` to stop the follow, `g` to go to the top, `G` to return to the end, and `q` to quit.
5. The user runs `hive --home /tmp/other-home`. The header shows `/tmp/other-home`, and the body shows the log of that home.
6. The clock passes midnight. The body keeps the lines of the day before and continues with the first event of the new daily file.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `cmd/root.go` | Pass the resolved home and the streams to the TUI entry. |
| `internal/tui/doc.go` | New. Package summary and file map. |
| `internal/tui/tui.go` | Replace the stub model with the log view model. Add an entry that takes the resolved home and the streams. Keep `Run(in, out)` as a wrapper that resolves the home itself. |
| `internal/tui/header.go` | New. The header line from the store: supervisor state, live agents, unread count. |
| `internal/tui/logview.go` | New. The viewport, the text filter, the level filter, the follow flag, the two key modes, and the keys. |
| `internal/tui/tick.go` | New. The 500-millisecond tick that reads new lines through the reader of T06 and refreshes the header. |
| `go.mod`, `go.sum` | Add Bubbles and Lip Gloss. The commit message states the reason: viewport and text input. |

## Requirements

### AC1 Header

- The header shows the home path, the supervisor state, the count of `live` agent records, and the unread inbox count over each role.
- The supervisor state is `live` when a `supervisor` claim exists and its `pid` names a running process. Otherwise it is `down`.
- The header refreshes on each tick. A record change in the store appears in the header after the next tick.
- The home is the value that the root command resolved. The TUI does not resolve the home again.

### AC2 Log body

- The body shows the events of today from the reader of T06, rendered with the same text format, in a viewport.
- Each tick reads the new lines through the reader of T06. With follow on, the viewport stays at the end.
- The reader of T06 handles the change of the daily file at midnight. The lines of the earlier day stay in the viewport. The first event of the new file appears after the next tick.
- When the file of today does not exist, the body shows `no events today`. The tick waits for the file and shows its first lines when it appears.
- The view has two key modes: viewing and editing.
- In viewing mode: `/` enters editing mode. `l` cycles the level filter debug, info, warn, error, and back to debug. `f` toggles follow. `g` goes to the first line. `G` goes to the last line and turns follow on. `q` and Ctrl-C quit.
- In editing mode, a printable key adds its character to the text input. Enter applies the text as a substring filter on the rendered line. Enter then returns to viewing mode.
- In editing mode, Escape clears the input and the applied filter. Escape then returns to viewing mode. Ctrl-C quits. No other key of viewing mode acts in editing mode.
- A status line shows the level filter, the text filter, and the follow state.

### AC3 Model tests

- The tests drive the model with `Update` and read `View`. They use a temporary home with a store and a log file, and a fixed clock for the daily file name.
- An end-to-end test runs the program on a string reader that holds `q`. It checks that the program returns and that the output holds the home path.

### AC4 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The header shows the four values for the resolved home and refreshes them after a change. | V1, V3 |
| AC2 | The body follows the log across the daily file change, and each key acts as the requirements give in its mode. | V2 |
| AC3 | The model tests and the end-to-end test pass. | V1, V2, V3 |
| AC4 | The gate passes. | V4 |

## Verification items

### V1 Header tests

- Proves: AC1, AC3.
- Command: `go test ./internal/tui/ -run TestHeader -v`, red and green.
- Expected: with no claim the header shows `supervisor: down`. With a claim whose `pid` is the test process the header shows `live`. Two live agent records and three unread records give `agents: 2` and `inbox: 3 unread`. After the test adds one live record and one unread record and sends a tick, the header shows `agents: 3` and `inbox: 4 unread`. The header shows the home path that the test passed.

### V2 Key, follow, and daily file tests

- Proves: AC2, AC3.
- Command: `go test -race ./internal/tui/ -run 'TestKeys|TestFollow|TestDailyFile' -v`, red and green.
- Expected: a tick after an appended line shows the line. `/`, the text `gq`, and Enter hide a line without `gq` and do not quit. Escape clears the filter and the input. `l` four times moves the level through info, warn, error, and back to debug. `f` stops the follow, and a new line does not move the viewport. `g` moves the viewport to the first line. `G` moves it to the last line and turns follow on. `q` and Ctrl-C each return a quit command. A test has two daily files and a fixed clock. After the clock passes midnight, the view shows the first line of the new file. The earlier lines stay in the viewport. A test without a file of today shows `no events today`, then the first line after the file appears.

### V3 End-to-end run

- Proves: AC1, AC3.
- Command:

```sh
go test ./internal/tui/ -run TestRunQuits -v
go build .
printf q | ./hive --home "$PWD"
echo "exit $?"
```

- Expected: the test passes and its output holds the home path of the test. The command exits with 0.

### V4 Gate

- Proves: AC4.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- Views of agents, tasks, and attention. They come after v1.

## Changes

None.
