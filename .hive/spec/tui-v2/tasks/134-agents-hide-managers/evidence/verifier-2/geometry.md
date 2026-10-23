# Scratch terminal geometry

Cause: **unknown**. Repeatability outside this verification run: **unknown**.

The first scratch tabs could not hold a 120x40 demo. No task defect follows from this observation.
The manager recorded the observation in friction commit `b34ee38d`.

## Commands and observations

Run from the task worktree:

```sh
herdr tab create --workspace "$HERDR_WORKSPACE_ID" --cwd "$PWD" --label verify134-second-diagnostic --no-focus
# New pane: wWV:p5. New tab: wWV:t5.
herdr pane run wWV:p5 'stty size; printf "COLUMNS=%s LINES=%s TERM=%s\n" "$COLUMNS" "$LINES" "$TERM"'
herdr pane read wWV:p5 --source visible --format text
herdr pane layout --pane wWV:p5
```

Observed terminal output:

```text
40 119
COLUMNS=119 LINES=40 TERM=xterm-256color
```

The layout response reported both the area and the pane rectangle as 120x40:

```json
{"area":{"height":40,"width":120,"x":0,"y":0},"panes":[{"pane_id":"wWV:p5","rect":{"height":40,"width":120,"x":0,"y":0}}]}
```

The preceding demo in scratch pane `wWV:p4` showed:

```text
this pane is 119×40
hive needs at least 120×40
```

A diagnostic `stty cols 120 rows 40` did not give a valid capture.
The terminal wrapped the screen. The comparison differed in 1464 cells.
The verifier closed only the diagnostic tab that it created:

```sh
herdr pane send-keys wWV:p5 ctrl+c
herdr pane send-keys wWV:p5 ctrl+c
herdr tab close wWV:t5
```

The failed setup attempts also closed their own tabs `wWV:t2`, `wWV:t3`, and `wWV:t4`.

## Isolated scratch server

The verifier checked that this session directory did not exist before this run:
`/home/penguin/.config/herdr/sessions/verify134-second/`.

Configuration file: `<worktree>/.hive/tmp/verify134-second/herdr.toml`.

```toml
[server]
headless_cols = 140
headless_rows = 48
[ui]
pane_scrollbars = false
```

Commands:

```sh
env -u HERDR_SOCKET_PATH -u HERDR_WORKSPACE_ID -u HERDR_TAB_ID -u HERDR_PANE_ID \
  HERDR_CONFIG_PATH="$PWD/.hive/tmp/verify134-second/herdr.toml" \
  nohup herdr --session verify134-second server \
  > .hive/tmp/verify134-second/server.out 2>&1 &
# Server PID: 3874404.
HERDR_SOCKET_PATH=/home/penguin/.config/herdr/sessions/verify134-second/herdr.sock \
  herdr workspace create --cwd "$PWD" --label verify134-second
# New workspace: w1. New root pane: w1:p1.
HERDR_SOCKET_PATH=/home/penguin/.config/herdr/sessions/verify134-second/herdr.sock \
  herdr pane run w1:p1 'stty size'
```

`scratch-stty.log` records `48 140` from that pane.
The independent live script used this socket and workspace, never the real home:

```sh
HERDR_SOCKET_PATH=/home/penguin/.config/herdr/sessions/verify134-second/herdr.sock \
HERDR_WORKSPACE_ID=w1 python3 .hive/tmp/verify134-second/live.py
```

Each demo used a scratch home marker, the task binary, and these environment values:
`HIVE_NO_STARTUP=1`, `TERM=vt100`, and `COLORTERM=truecolor`.
The demo sizes were 120x40 and 80x24. The script compared only those top-left cells.

The larger scratch server gave valid captures. This does not establish the cause of the original size difference.
`cleanup.log` records removal of this server, its session directory, and all scratch files.
