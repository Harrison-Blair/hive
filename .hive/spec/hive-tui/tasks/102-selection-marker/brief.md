# Selection marker and default text color

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, fix after wave C |
| Tier | decent |
| Depends on | 92-agents-board-layout, 95-goals-queue-nav, 96-tasks-evidence-view, 97-friction-detail-panel, 90-tui-idle-cpu. Base: `dev` after each lands |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The user approved these style rules for the TUI. Task 100-no-backgrounds
removed each painted background and added a test for it. This task applies
the other two rules on each screen. It runs after the other screen tasks, so
that it changes each screen once. When the code does not agree with this
brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Rules

1. No painted background anywhere. Task 100 did this. Keep its test.
2. A selected item has honey bold text (`#f5b942`) and a `▶` marker before
   it. This applies to each selectable thing on each screen: the sidebar
   item, each selected row, each selected tab or filter where the user moves
   a selection, and the selected graph node on Agents. No fill block.
3. Normal text uses the default foreground of the terminal. Remove
   `Foreground(colorText)` from the base styles (`headerStyle`,
   `panelStyle`, and any other base style). Remove `colorText` when it
   becomes unused. Keep the accent, the status colors, the dim color, and the
   border color.

## Known facts

- `internal/tui/shell.go` defines the styles. `selectedStyle` is honey bold
  text.
- On `dev` at the time of this brief, some selections have no `▶`, for example
  the graph node in `internal/tui/agents.go`, the Health check name in
  `internal/tui/health_screen.go`, and the `[text]` selection in
  `internal/tui/task_screen.go`. Find each place that renders a selection
  after the other tasks land.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/` and its tests | R1, R2. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy, when `hive task create` did not write it. |

Do not edit other files.

## Requirements

### R1 Styles

Apply rules 2 and 3 on each of the 8 screens.

### R2 Tests

- Write each test first and see it fail on the base for the correct reason.
- For each screen, render the model with a selection at 120 x 40 and check the
  raw output: the selected text is in a honey bold SGR span
  (`38;2;245;185;66` and `1`) and starts with `▶`. Parse each SGR sequence
  in the test. Do not use the style code under test to make the expected
  codes.
- Check that no SGR sequence sets the foreground `38;2;232;228;218`
  (`#e8e4da`).
- Keep the background test of task 100.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each R2 test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Build the task binary. Run it on a scratch home (`HIVE_HOME` and a temp `XDG_CONFIG_HOME`) in a herdr tab. Capture each screen with ANSI codes. Move the selection once on each screen | Each selection has the marker and the honey bold text. Normal text has no foreground code. No background code. |
| V3 | `git diff origin/dev --stat` | Only the files of the source table changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0. |

Never open the Chat input on the real home. Close each herdr tab and
workspace that you open. Do not close a pane that you did not open.

## Finish

Commit on the task branch with a subject such as `fix(tui): mark each
selection and use the terminal text color`. Push the task branch. Write the
evidence. Run `hive task complete`. Send the commit SHA and the V1-V4 results
to the manager with `hive inbox send --to manager-hive`.

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

- `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Main.dc.html`
- `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Chat.dc.html`
- `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Goals.dc.html`
- `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Queue.dc.html`
- `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Tasks.dc.html`
- `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Friction.dc.html`
- `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Health.dc.html`
- `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Logs.dc.html`

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

## Scope change: gaps from task 92

The user decided to land task 92 when its structure and colors pass, and to
move each other visual gap of the Agents screen into this task. The verifier
of task 92 listed these gaps against the board `Main.dc.html`. Do each row
with the action "fix". Keep each row with the action "keep": a user rule
overrides it. The source table of this task also includes
`internal/tui/agents*.go` and its tests for these rows.

| Element | Board | Output of task 92 | Action |
| --- | --- | --- | --- |
| Header working count | `6 working` in green | default text | fix |
| Header needs-you count | `0` in honey | default text | fix |
| Tabs | `[ GRAPH ]` honey, `TREE` and `TABLE` dim | the whole tab line honey bold | fix |
| Legend | right-aligned, dim | left, after the tabs, honey bold | fix |
| Explorer status dots | live `●` green | default text | fix |
| Explorer working count | `· 1 working` dim | default text | fix |
| Explorer selected row | fill block behind the row | honey bold with `▶`, no fill | keep (user rule: no background) |
| Pod title text | full goal title, for example `≡ hive · H2 spec layout` | `≡ hive · Hi…` | fix: fit the full title when the frame has space. Use `…` only when the title does not fit. |
| Selected node border | honey | default | fix |
| Detail strip border | honey | muted line color | fix |
| Detail strip name | honey bold | default text | fix |
| Detail strip labels and values | labels dim, values bright | all default text | fix |
| Key bar keys | key chips with a light fill | plain text | keep (user rule: no background) |

For each "fix" row, write a test first that reads the text or the SGR codes,
and see it fail. In the Design element table of the evidence, give one row
for each row above. The verifier of this task checks each row.

## Scope change: colour is never the only signal

The user is colourblind (`overrides/colourblind-ui.md` in the home). In each
screen that this task changes, a glyph or a word shows each state. Colour can
add to a state. Colour must not be the only signal of a state. Examples: a
live agent has `●` and a stopped agent has `○`. A failed check has `fail` or
`✗`, not only red. The selected row has `▶`, not only honey. Write a test
that reads the plain text (with the SGR codes removed) and finds the glyph or
word of each state. In the Design element table, add one column "state
without colour" with the glyph or word of each row.
