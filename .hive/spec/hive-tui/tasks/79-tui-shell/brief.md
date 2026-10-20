# TUI shell

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, wave B |
| Tier | strong |
| Depends on | 74-goal-record, 75-needs-user, 76-agent-lineage, 77-chat-transport (all landed). Base: `dev` at `bf39fc51` or later |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Make the frame of the hive TUI that bare `hive` opens: the header, the
sidebar of 8 screens, the look, background data loading, attach, and quit.
Two later wave B tasks fill the Chat and Agents screens. Wave C fills the
other screens. The feature README is `.hive/spec/hive-tui/README.md`. Read it
first. The README approves each decision. Do not change a decision. When the code
does not agree with this brief or the README, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Sources

- The designs in the home: `.hive/tmp/coordinator/designs/hive-tui/Main.dc.html`
  gives the frame (header, sidebar, key hints). The other `*.dc.html` files
  show the screens. When a design and a README decision do not agree, the
  decision wins. G12: the TUI has no action except chat, attach, and quit.
- The current TUI: `internal/tui/` (a log view, about 1900 lines).

## Known facts

- `internal/lib/herdr/` has no call for `pane.focus` or `agent.focus`.
  `reference/herdr/cli-mapping.md` describes both.
- Today `q` and one `ctrl+c` quit the TUI (`internal/tui/logview.go`).
- The colors of the designs: background `#0e0f13`, panel `#12141a`, border
  `#3d3a33`, text `#e8e4da`, dim text `#a39d90`, accent (honey) `#f5b942`,
  working `#6ee69a`, done `#7ccf6a`, info `#6fb7ff`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/` (new files allowed) and its tests | R1-R6. |
| `internal/lib/herdr/` (a new focus call) and its tests | R5. |
| `testdata/ste-baseline.json` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files. Do not change the startup of bare `hive` (wave C).

## Requirements

### R1 Frame

- A header line: `⬢ hive`, the count of projects, the count of live agents
  with the working and idle counts, the supervisor state (fresh or stale
  claim), the count of tasks with `needs_user`, and the clock.
- A sidebar with the 8 screens in this order: 1 Chat, 2 Agents, 3 Goals,
  4 Queue, 5 Tasks, 6 Friction, 7 Health, 8 Logs. The keys `1` to `8` select
  a screen. The current screen has the accent color.
- A footer with the key hints of the current screen.
- The Logs screen shows the current log view, unchanged in behavior except
  for R4. Each other screen shows its name and the text `next wave` until a
  later task fills it.
- Each screen is one Go type behind one small interface (update, view, key
  hints, data request), so that a later task adds a screen in its own file.
  Write a short `doc.go` map, as `AGENTS.md` requires.

### R2 Look

Use Lip Gloss with the colors of the Known facts. Panels have a rounded
border in the border color, and a title in the accent color. The TUI works at
80x24 and at larger sizes. At a width under 100 the sidebar shows only the
screen numbers and names.

### R3 Background loading

- No file read, store read, herdr call, or other I/O runs in `Update` or
  `View`. Each load runs in a `tea.Cmd` and returns a message.
- A refresh every 2 seconds loads the header data and the data of the
  current screen. One load at a time for each source: a slow load does not
  start a second one.
- A load error shows in the footer and does not stop the TUI.

### R4 Quit

- `ctrl+c` one time shows `press ctrl+c again to quit` in the footer. A
  second `ctrl+c` within 3 seconds quits. Each other key clears the notice.
- `q` does not quit. In the Logs screen `q` does nothing.
- Quit stops only the TUI. It does not stop an agent or the supervisor.

### R5 Attach

- Add a herdr call for `agent.focus` (or `pane.focus`), with a test against
  the fake herdr server of the package tests.
- The shell gives a screen one way to ask for attach to a pane ID. The
  shell runs the focus call in a `tea.Cmd` and shows an error in the footer.
  The Agents screen uses it later. No screen in this task uses it, so test it
  at the shell.

### R6 Tests

- Test at the model: call `Update` and `View` with messages.
- A test for each R1 key, the narrow layout, R3 (a slow load does not block a
  key, and does not start twice), each R4 case, and R5.
- A program test on a string reader: the program starts, gets `2`, `8`,
  `ctrl+c`, `ctrl+c`, and ends.
- Do not compute an expected value with the code under test.

## Rules

- Each Go change is test-first. See each new test fail for the correct
  reason.
- When `ste_test.go` fails on the brief copy, change the words of each
  sentence in the copy that causes a violation. Keep its meaning. Record each
  change in the evidence.
- Live checks use a temp `HIVE_HOME` and a temp `XDG_CONFIG_HOME`, and clean
  up each herdr workspace that they make. Do not run `rm -rf` on an unquoted
  variable.

## Not in scope

The content of each screen. The startup change of bare `hive`. Config keys.
`main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each new test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Live, in a new herdr workspace: build the task binary and run it with a temp `HIVE_HOME` that has two agent records. Press `1` to `8`, `q`, `ctrl+c` one time, wait 4 seconds, then `ctrl+c` two times. Read the pane with `herdr pane read` after each step. Resize the pane to under 100 columns and read it again. | Each screen shows in turn. `q` does nothing. The notice shows, then clears. The second pair quits. The narrow layout fits. Clean up the workspace. |
| V3 | `git diff dev --stat` | Only the files of the source table and the evidence changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `feat(tui): add the screen
shell`. Push the task branch. Write the evidence. Run `hive task complete`.
Send the commit SHA and the V1-V4 results to the manager with
`hive inbox send`.

## Scope change 1

`cmd/root_test.go` is now in the source table, only for
`TestBareHiveRunsTUIUntilQuit` and other tests that assert the old TUI header
or quit keys. Change the expected header text to the new header (for example
`⬢ hive`) and the quit input to two `ctrl+c`. Keep each other assertion. The
test must still prove that bare `hive` runs the TUI and that the TUI quits.
List each change in the evidence.
