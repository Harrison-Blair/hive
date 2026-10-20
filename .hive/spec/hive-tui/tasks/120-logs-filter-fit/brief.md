# Logs filter line fits the panel

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, fix after wave C |
| Tier | decent |
| Depends on | none. Base: `dev` |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

At 120 x 40 the Logs panel loses its bottom border. The verifier of task 118
found the cause. `internal/tui/shell.go` (about line 71) counts the filter
lines with `wrapFilterLine(m.statusLine(), width-4)`. Then it adds the
`▶ ` prefix to the first line. The first line is now wider than the panel,
and it wraps again. The panel has one line more than the count, and the
bottom border goes out of the frame. The fault is on `dev` before task 118.
When the code does not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/shell.go`, `internal/tui/logview.go`, and their tests | R1, R2. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy, when `hive task create` did not write it. |

Do not edit other files.

## Requirements

### R1 Count the prefix

Count the width of the `▶ ` prefix before the filter line wraps, so that
each filter line fits the panel and the count is correct. Keep the `▶ `
marker and its style.

### R2 Tests

Write the test first and see it fail on the base. At 120 x 40 and at
200 x 57, with a filter line long enough to wrap, the Logs frame has exactly
the height in lines and the width in cells, and the last line of the Logs
panel is its bottom border. Use literal expected text.

## Design

The approved board for this task:

- `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Logs.dc.html`

Follow the visual check rule in the worker skill. Judge only the filter line
and the panel border. Record each other difference as `differs` with the
cause "not in the scope of this task". The user is colourblind: each state
has a glyph or a word, and colour only adds to it.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run the R2 test against the base code, then against the task code | It fails on the base and passes on the task branch. |
| V2 | Build the task binary. In a scratch home, open Logs with a long filter line. Test 120 x 40 and 200 x 57 panes. Set each pane to the exact listed size. You can use tmux. | Each frame has exactly the height and width. The bottom border shows. |
| V3 | `git diff origin/dev...HEAD --stat` | Only the files of the source table and the evidence changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0. |

Set `HIVE_HOME` to a scratch home. Set `XDG_CONFIG_HOME` to a temporary directory.
Run the binary in a tab in your task workspace. Prefix the command with
`env -u HERDR_ENV -u HERDR_PANE_ID -u HERDR_WORKSPACE_ID -u HERDR_SOCKET_PATH`.
Stop and close only the pid and pane that you started. Never work in
workspace `wBC`. When your brief copy has STE violations, reword those
sentences to 0 and keep the meaning of each. Record each change.

## Finish

Commit on the task branch with a subject such as `fix(tui): fit the Logs
filter line`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V4 results to the
manager with `hive inbox send --to manager-hive`.
