# Logs screen and request IDs

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, wave C |
| Tier | decent |
| Depends on | 80-chat-screen, 81-agents-screen. Base: `dev` after both land |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Fill screen 8, Logs, with filters and request groups, and give each CLI command one `request_id`. The feature README is `.hive/spec/hive-tui/README.md`. Read it first,
mainly goal decision 5 and G14. The manager approved the README decisions. Do not change a decision. When the
code does not agree with this brief or the README, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Sources

- The design in the home: `.hive/tmp/coordinator/designs/hive-tui/Logs.dc.html`. The
  decisions win over the design. G12: the TUI has no action except chat,
  attach, peek, and quit. Where a design shows another action, show the hint
  `ask in Chat` in the footer, or leave the action out.
- The shell: `internal/tui/shell.go` (the `screen` interface, the data
  requests, attach). The Chat and Agents screens are examples.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/` (new files for the screens, and the shell wiring) and its tests | The R items. |
| `cmd/common/`, `cmd/logs/`, `internal/lib/log/`, `internal/logs/` and their tests | R1. |
| Each other `cmd/` file that makes its own logger, and its tests | R1. Only the `request_id` field. |
| `testdata/ste-baseline.json` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### R1 Request ID

- Each CLI command makes one `request_id` (`r-` and 8 lowercase hex
  characters, random) at start, and each log event of that command carries
  it. A long command (`hive supervise`) makes one for each loop pass.
- One `command start` event and one `command end` event with `duration_ms`
  and the outcome, for each CLI command.
- `hive logs --request ID` shows only the events of that request.

### R2 Logs screen

- One line for each event: time, level, component, the event text, and the
  duration when it has one. A command start and its end fold into one line
  with the duration.
- Filters: `l` cycles the level through debug, info, warn, and error. Info is
  the default level. `c` cycles the component, `p` cycles the project, and `/`
  opens text search. The filter line shows each filter. `f` toggles follow.
- `enter` shows the JSON of the selected event in a side panel.
- `r` shows only the events of the `request_id` of the selected event. `r`
  again or `esc` shows all.
- `[` and `]` change the day file.
- Keep the behavior of the current log view where it does not conflict.

## Rules

- Each Go change is test-first. See each new test fail for the correct
  reason. Test at the model: call `Update` and `View` with messages. Assert
  the rendered lines of one small fixture as literal strings. Add one program
  test on a string reader.
- All reads run in a `tea.Cmd` through the shell data request. A screen keeps
  its selection across a refresh.
- When `ste_test.go` fails on the brief copy, change the words of each
  sentence in the copy that causes a violation. Keep its meaning. Record each
  change in the evidence. Add the brief copy to `testdata/ste-baseline.json`
  with 0.
- Live checks use a temp `XDG_CONFIG_HOME`, and clean up each herdr workspace
  that they make. A live check against the real home only reads. Do not run
  `rm -rf` on an unquoted variable.
- Other wave C tasks run at the same time and edit the shell wiring in
  `internal/tui/`. Keep the change to shared files small.

## Not in scope

The startup change (another task). `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each new test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | With a temp `HIVE_HOME`: run `hive task list --project x`, then `hive logs --json` and `hive logs --request <id of that command>` | Each event of that command has the same `request_id`. The start and end events exist. The request filter shows only them. |
| V3 | Live, in a new herdr workspace with the temp home: press `8`, `l`, `c`, `/`, a text, `enter`, `r`, `]`. Read the pane after each step. | Each filter works. The request view shows one command. Clean up the workspace. |
| V4 | `git diff dev --stat` | Only the files of the source table and the evidence changed. |
| V5 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `feat(logs): add request IDs and the Logs screen`. Push the task
branch. Write the evidence. Run `hive task complete`. Send the commit SHA
and the V1-V5 results to the manager with `hive inbox send`.

## Scope change 1

- The source table now also allows `cmd/supervise/` and
  `internal/supervise/` (only the loop pass: one `request_id` and the start
  and end events for each pass), with their tests.
- `hive version` is an exclusion. It writes no log events and gets no
  `request_id`.

## Repair 1

The verifier failed V3. Read
`.hive/spec/hive-tui/tasks/85-logs-screen/evidence/VERIFIED.md`. Do only this
repair:

1. `[` and `]` change only the day label. The default log loader
   (`internal/tui/tui.go`, near line 120) reads only the file of today, so
   each other day shows `no events on <day>`. Make the loader read the file
   of the selected day.
2. `TestLogsScreenDayNavigationLoadsSelectedFile` passes only because it
   moves the clock. Change it so that the clock stays on one day and the test
   reads events from the file of an earlier day. See it fail on the current
   code first.
3. Merge `origin/dev` into the task branch. It conflicts in
   `internal/tui/tick.go` (`screenLoadMsg`: the `day` field of this task and
   the `goalsQueue` field of task 82). Keep both fields.
4. Run V3 again. Run the gate. Commit, push, run `hive task complete`, and
   send the SHA to the manager with `hive inbox send`.

Do not change other behavior.

## Repair 2

Task 86 landed on `dev` after the verification. The merge of this branch into
`dev` conflicts only in `testdata/ste-baseline.json`. Do only this repair:

1. Merge `origin/dev` into the task branch. In `testdata/ste-baseline.json`,
   keep each entry of both sides. When both sides have the same key, keep the
   lower count.
2. Run the gate. Commit, push, run `hive task complete`, and send the SHA to
   the manager with `hive inbox send`.

Do not change other files.
