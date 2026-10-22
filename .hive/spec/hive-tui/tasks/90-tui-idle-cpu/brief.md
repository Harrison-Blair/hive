# TUI idle CPU

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, repair after wave C |
| Tier | decent |
| Depends on | 88-tui-terminal-io. Base: `dev` after T1 lands |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The TUI uses 13 to 15 percent CPU when the user does nothing. Find the cause
and fix it. The feature README is `.hive/spec/hive-tui/README.md`. G14 is
approved: each screen loads data in a background command. When the code does
not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `internal/tui/tick.go` sends `tickMsg` every 2 seconds. On each tick,
  `Update` starts a header load and a load of the current screen. The loads
  read agents, the logs of the day, and doctor checks.
- Hypothesis, not proved: these loads, or a redraw on each message, cause the
  CPU use. Measure first.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/` and its tests | R1, R2. |
| A package that a TUI load calls, only when the profile shows that the cost is in it, and its tests | R1. List each file in the evidence. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy, when `hive task create` did not write it. |

Do not edit other files.

## Requirements

### R1 Cause and fix

- Measure the idle CPU of the base binary on each screen in a real terminal
  for 60 seconds. Take a CPU profile (for example with `pprof` from a test
  binary or a temporary build flag that you do not commit). Name the cause in
  the evidence.
- Fix the cause. The idle TUI uses less than 2 percent CPU on each screen.
- The data on each screen stays fresh: a change in state shows within
  5 seconds, as today.

### R2 Tests

Write each test first and see it fail for the correct reason. For example,
when the fix skips a reload with no change, test that a tick with no change
starts no load, and that a tick after a change starts one.

## Not in scope

The Agents content and layout. Other screen content. `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each R2 test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Build the base and the task binaries. Run each in a herdr pane on this home for 60 seconds on each of the 8 screens with no input. Measure CPU with `ps -o %cpu` or `pidstat` | The base shows the reported use. The task binary is below 2 percent on each screen. |
| V3 | With the task binary on the Agents and Logs screens, change state (for example `hive inbox send` or a log event) | The screen shows the change within 5 seconds. |
| V4 | `git diff origin/dev --stat` | Only the files of the source table changed. |
| V5 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other commands exit 0. |

Close each herdr tab and workspace that you open. Do not close a pane that
you did not open.

## Finish

Commit on the task branch with a subject such as `fix(tui): stop the idle
reload cost`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V5 results to the
manager with `hive inbox send`. Read only your own inbox.

## Task names

This brief uses these names: T1 is 88-tui-terminal-io, T2 is 90-tui-idle-cpu,
T3 is 91-agents-live-generation, T4 is 92-agents-board-layout, and T5 is
89-screen-board-audit.

## Scope change: chat safety

A TUI check uses a scratch home (a temp copy of a marked home, set with
`HIVE_HOME`, and a temp `XDG_CONFIG_HOME`), or it does not open the Chat
input. On the real home, never press `i` on the Chat screen and never press
Enter there. A chat message from an agent has the authority of the user.
