# Goals and Queue navigation

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, fixes after wave C |
| Tier | decent |
| Depends on | 88-tui-terminal-io, 92-agents-board-layout (both change the header). Base: `dev` after both land |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Task 89-screen-board-audit compared each screen with the board that the user
accepted. This task fixes the gaps below. Read the gap list in
`.hive/spec/hive-tui/tasks/89-screen-board-audit/evidence/GAPS.md` (on
`dev` after 89 lands, else on the branch `origin/task/89-screen-board-audit`)
and the board `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Goals.dc.html` and `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Queue.dc.html`. The feature README is
`.hive/spec/hive-tui/README.md`. Its decisions and G1-G14 win over the
board. G12: the TUI has no action except chat, attach, peek, and quit, and
read-only views. When the code does not agree with this brief, stop and report
it with `hive inbox send --to manager-hive TEXT --agent NAME`.

## Gaps to fix

- G1: On Goals, `t` opens the tasks of the selected wave and `o` opens its design, as read-only views.
- Q1: The header on Queue shows the counts of tasks that run, that stop, and that wait behind a stop, as on the board.
- Q2: Queue shows one panel for each registered project, also a project with no open task. Remove the `p` project cycle (decision 9).
- Q3: On Queue, the arrows select a task and `enter` opens it on the Tasks screen.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/goals_queue.go`, `internal/tui/header.go` and their tests | The gaps. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy, when `hive task create` did not write it. |

Do not edit other files. When a gap needs a change in another file, stop and
report it.

## Requirements

- Write each test first and see it fail for the correct reason. Test at the
  model with `Update` and `View`. Write each expected text as a literal.
- Each frame stays exactly the size of the terminal at 120 x 40 and 200 x 57.
  Use the frame size test of 88-tui-terminal-io.
- Keep each test that passes today. Do not weaken an assertion.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each new test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Build the task binary. Make a scratch home with fixture data. Set `HIVE_HOME` to it and set a temp `XDG_CONFIG_HOME`. Run the binary in a herdr tab of about 200 x 57. Capture the screen with `herdr pane read` for each gap | The screen agrees with the board for each gap. Put the captures in the evidence. |
| V3 | `git diff origin/dev --stat` | Only the files of the source table changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0. |

Never open the Chat input on the real home. Close each herdr tab and
workspace that you open. Do not close a pane that you did not open.

## Finish

Commit on the task branch with a subject such as `fix(tui): match the Goals and Queue boards`. Push the task
branch. Write the evidence. Run `hive task complete`. Send the commit SHA and
the V1-V4 results to the manager with `hive inbox send --to manager-hive`.

## Scope change: no startup in a scratch home

Bare `hive` starts a coordinator tab and a supervisor process when it runs in
herdr. In a live check on a scratch home, start the binary with the herdr
variables removed, so that it runs only the TUI:
`env -u HERDR_ENV -u HERDR_PANE_ID -u HERDR_WORKSPACE_ID -u HERDR_SOCKET_PATH ./hive`.
After each check, make sure that you started no `hive supervise` process and
no coordinator tab. Stop each process and close each tab that you started.

## Scope change: ownership

Never send a signal to a process and never close a tab, pane, or workspace
that you did not start. Record the pid and the pane ID of each thing that you
start. Before you stop or close it, compare its pid or ID with the record.
Never work in workspace `wBC`: it holds the user's TUI, the coordinator, and
the supervisor.

## Design

The approved boards for this task:

- `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Goals.dc.html`
- `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Queue.dc.html`

Visual check rule (the user approved it on 2026-10-20):

- Worker: capture each changed screen and each board above as PNG. Compare
  them. Fix each gap before you finish.
- Board to PNG: `firefox --headless --no-remote --profile SCRATCH
  --window-size=1220,780 --screenshot OUT.png file://BOARD`, with a new
  scratch profile folder below `.hive/tmp/`.
- Screen to PNG: start the TUI on a scratch home in a pane that you started.
  Run `herdr pane read PANE --source visible --format ansi > F.ansi`. Change
  the ANSI SGR codes to an HTML `<pre>` with a short script in `.hive/tmp/`.
  Then use the firefox command on that HTML file.
- Evidence: commit the screen PNGs in the task evidence. Do not commit a
  board PNG. Keep it below `.hive/tmp/` and give its path.
- Report: in the evidence, one table with one row for each board element:
  `matches`, `differs` (with the cause), or `overridden` (with the
  decision that overrides it).
- Capture only your own panes or a scratch run.

## Scope change: colour is never the only signal

The user is colourblind (`overrides/colourblind-ui.md` in the home). In each
screen that this task changes, a glyph or a word shows each state. Colour can
add to a state. Colour must not be the only signal of a state.

Examples:
- A live agent has `●`.
- A stopped agent has `○`.
- A failed check has `fail` or `✗`, not red alone.
- A selected row has `▶`, not honey alone.

Write a test that removes SGR codes, reads plain text, and finds one glyph or
word for each state. Add a "state without colour" column to the Design element
table. Each row must show its glyph or word.
