# Chat and Logs frame faults

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, Chat and Logs fixes after wave C |
| Tier | decent |
| Depends on | 102-selection-marker (order only). Base: `dev` |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The verifier of task 102 found two faults that are on `dev` before that
task. Chat at 120 x 40 with real chat data gives 46 rows, not 40, and it
wraps a line in the middle of a word. On Logs with a large log in follow
mode, the selected row is off the screen. When the code does not agree with
this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/chat_screen.go` and its tests | R1. |
| `internal/tui/logview.go` and its tests | R2. |
| `internal/tui/shell.go`, `internal/tui/tui.go` | Only when R1 or R2 needs a change there. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy, when `hive task create` did not write it. |

Do not edit other files.

## Requirements

### R1 Chat fits the frame

- At each size, the Chat frame has exactly the height in lines and the width
  in cells, with any number of chat entries and with long entries.
- A long entry wraps at a space. A word that is wider than the line breaks
  at the line edge.
- The newest entries show. Older entries scroll out at the top.

### R2 Logs keeps the selected row on the screen

- In follow mode, the TUI selects and shows the newest event on screen for any number of events.
- When the user moves the selection, keep the selected row on screen.

### R3 Tests

Write each test first and see it fail on the base. Use fixture chat entries
and fixture log events, for example 200 chat entries with some longer than
the width, and 5000 log events. Check the frame at 120 x 40 and 200 x 57,
and check that the selected row is in the frame. Use literal expected text.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run the R3 tests against the base code, then against the task code | Each fails on the base and passes on the task branch. |
| V2 | Build the task binary. Copy the real chat log and real log files to a scratch home. Open Chat and Logs at 120 x 40 and 200 x 57. | Each frame has exactly the height and the width. Chat wraps at spaces. The Logs selected row shows in follow mode and after moves. |
| V3 | `git diff origin/dev --stat` | Only the files of the source table changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0. |

Set `HIVE_HOME` to a scratch home. Set `XDG_CONFIG_HOME` to a temporary directory.
Copy the real chat log and the real log files to the scratch home.
Open Chat and Logs in a tab of your task workspace at 120 x 40 and 200 x 57.
Start the binary with
`env -u HERDR_ENV -u HERDR_PANE_ID -u HERDR_WORKSPACE_ID -u HERDR_SOCKET_PATH`.
Stop only the process ID that you started. Close only the pane that you started. Never work in
workspace `wBC`. Never open the Chat input on the real home. Never send chat.

## Finish

Commit on the task branch with a subject such as `fix(tui): keep Chat and
Logs in the frame`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V4 results to the
manager with `hive inbox send --to manager-hive`.
