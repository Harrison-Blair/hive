# T06 logs command

| Item | Value |
| --- | --- |
| Task ID | T06-logs-command |
| Wave | 2 |
| Tier | decent |
| Depends on | T01, CP1 |
| Blocks | T08 |
| Branch | `task/T06-logs-command` |
| Worktree | `projects/hive/worktrees/T06-logs-command/` |
| Evidence | `evidence/T06-logs-command/` |

## Overview

This task gives `hive logs`, the reader of the daily log files. It filters by
time, level, component, task, agent, and request, and it follows the current
file. T08 reuses the reader and the filter for the TUI. `logs` is a stream, so
its `--json` output is JSON lines, the one exception to the envelope rule.

## User flows

1. The user runs `hive logs --since 30m --level warn` and sees each warning and error of the last 30 minutes, one line each.
2. A manager runs `hive logs --task 12-spawn-worker` and sees each event of that task.
3. The user runs `hive logs --follow` in a pane and watches new events arrive within one second. Ctrl-C ends the command with exit status 0.
4. A script runs `hive logs --request 3f9a1c2e-17 --json` and parses the two JSON lines of one herdr request.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `cmd/root.go` | Register the `logs` command. |
| `cmd/logs/logs.go` | New. The flags and the signal handling. |
| `internal/logs/doc.go` | New. Package summary and file map. |
| `internal/logs/read.go` | New. Find the files of a date range, read the events in order with a line cap, and return the end offset. |
| `internal/logs/filter.go` | New. The filter type and its match function. T08 imports it. |
| `internal/logs/follow.go` | New. Follow the current file from an offset and the day rollover. |
| `internal/logs/render.go` | New. The text line for one event. |

## Public API for T08

T06 exposes the reader and the filter of `internal/logs` to T08. Record these
signatures in `V1.md`. T08 imports these names.

```go
type Event struct {
    Raw    []byte
    Fields map[string]any
}

type Filter struct {
    Level     string
    Component string
    TaskID    string
    AgentID   string
    RequestID string
}

func (f Filter) Match(e Event) bool

type Offset struct {
    Path string
    Byte int64
}

type Reader struct { /* unexported fields */ }

func NewReader(home string) *Reader
func (r *Reader) Read(ctx context.Context, since time.Time) ([]Event, Offset, error)
func (r *Reader) Follow(ctx context.Context, from Offset) (<-chan Event, <-chan error)
```

## Requirements

### AC1 Read and filter

- `hive logs [--since VALUE] [--level L] [--component C] [--task ID] [--agent NAME] [--request ID] [--limit N] [--follow]`.
- `--since` takes a duration such as `30m` or `2h`, or an RFC 3339 time. The cutoff is now minus the duration, or the given time. The reader keeps each event whose `time` is at or after the cutoff. Without `--since` the command reads the file of today and keeps each event.
- The reader opens each daily file from the date of the cutoff to today in date order and skips the dates without a file.
- A filter matches the fields `level`, `component`, `task_id`, `agent_id`, and `request_id`. `--level` keeps the named level and the levels above it.
- `--limit N` prints the last N matches. The default is 200. With `--follow` the limit applies to the first print only.
- A line has at most 1 MiB. The reader drains a longer line without a copy of it and treats it as malformed.
- In text mode a malformed line prints as `?` with the first 200 bytes of the raw line. In JSON mode the command skips the line and writes `hive logs: malformed line <file>:<n>` to stderr. The command continues in both modes.

### AC2 Render

- A text line: the local time with milliseconds, the level in four letters, the component, and the message. Each other field follows as `key=value` in key order.
- With `--json` the command prints each matching event as its raw JSON line. Each stdout line parses as JSON. `logs --json` streams raw event lines and does not print a success envelope.

### AC3 Follow

- `--follow` prints the matches, then continues from the offset after the last byte that the reader consumed. The follower keeps each partial line between the initial snapshot and the follow start. A line that arrives between the first print and the first poll prints too.
- The follower checks the file size every 500 milliseconds and prints each new line that matches. It keeps an incomplete line until its newline arrives.
- At midnight the follower reads the rest of the old file. Then it moves to the new daily file within one second after its first line.
- `SIGINT` ends the command with exit status 0. The transcript measures the child status, not the wrapper status. GNU `timeout` returns 124 by default, so use `timeout --preserve-status --kill-after=2s -s INT 2`, or inspect the child wait status directly.

### AC4 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The reader orders files by date, applies the cutoff and the line cap, and the filters keep exactly the matching events. | V1, V3 |
| AC2 | Text lines and JSON lines have the format of the requirements. | V3 |
| AC3 | The follower prints a new line within one second, loses no line, moves to the next daily file, and ends with 0 on `SIGINT`. | V2, V3 |
| AC4 | The gate passes. | V4 |

## Verification items

### V1 Reader and filter tests

- Proves: AC1.
- Command: `go test ./internal/logs/ -run 'TestRead|TestFilter|TestSince' -v`, red and green.
- Expected: the reader reads three daily files in a temporary directory in date order and prints a malformed line as `?`. A line of 1 MiB plus one byte counts as malformed, and the valid line after it prints. The cutoff cases: an event one second before the cutoff is gone, an event at the cutoff stays, and an event after it stays. A cutoff on the day before today reads two files. Each filter field has a match case and a miss case.

### V2 Follow test

- Proves: AC3.
- Command: `go test -race ./internal/logs/ -run TestFollow -v`, red and green.
- Expected: a writer goroutine appends a line after the follow starts, and the line reaches the output within one second. A line that the writer appends between the first read and the follow start reaches the output once. A line that arrives in two writes prints once, after the newline. A case writes a file with the next date and sees the follower finish the old file and move. A cancelled context ends the follower with a nil error.

### V3 Command transcript

- Proves: AC1, AC2, AC3.
- Setup: `go build .` and make a unique temporary home. Create its marker `.hive/README.md` explicitly. Export `HIVE_TMP=<absolute fixture path>`. Do not append malformed data to a real daily log. Make a unique malformed fixture in `$HIVE_TMP`.
- Command:

```sh
./hive --home "$HIVE_TMP" project list >/dev/null
./hive --home "$HIVE_TMP" logs --limit 2
./hive --home "$HIVE_TMP" logs --component cli --json | tail -n 1 | python3 -m json.tool
./hive --home "$HIVE_TMP" logs --level error
printf 'not json\n' >> "$HIVE_TMP"/.hive/logs/"$(date +%F)".jsonl
./hive --home "$HIVE_TMP" logs --limit 1
./hive --home "$HIVE_TMP" logs --limit 1 --json
timeout --preserve-status --kill-after=2s -s INT 2 ./hive --home "$HIVE_TMP" logs --follow --limit 1; echo "exit $?"
```

- Expected: two text lines in the format of AC2, one parsed JSON event, and no line for the error level. Then one `?` line with `not json`. Then one JSON line on stdout and one `malformed line` diagnostic on stderr. The follow prints one line, waits, and ends with exit 0 on the signal. The measured status is the child status, not GNU timeout status 124.
- Cleanup: remove only `$HIVE_TMP`. Do not remove the worktree `.hive/state`, `.hive/logs`, `.hive/tmp`, a report, or a log.

### V4 Gate

- Proves: AC4.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- The TUI view. T08 builds it on this reader.

## Changes

None.
