# Tasks and Friction screens

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, wave C |
| Tier | decent |
| Depends on | 80-chat-screen, 81-agents-screen. Base: `dev` after both land |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Fill screen 5, Tasks, and screen 6, Friction. The feature README is `.hive/spec/hive-tui/README.md`. Read it first,
mainly goal decision 3 and G8, G12. The README approves each decision. Do not change a decision. When the
code does not agree with this brief or the README, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Sources

- The design in the home: `.hive/tmp/coordinator/designs/hive-tui/Tasks.dc.html` and `Friction.dc.html`. The
  decisions win over the design. G12: the TUI has no action except chat,
  attach, peek, and quit. Where a design shows another action, show the hint
  `ask in Chat` in the footer, or leave the action out.
- The shell: `internal/tui/shell.go` (the `screen` interface, the data
  requests, attach). The Chat and Agents screens are examples.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/` (new files for the screens, and the shell wiring) and its tests | The R items. |
| `testdata/ste-baseline.json` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### R1 Tasks

- A tab bar with counts: `active` (assigned, completed, verified and not
  landed), `queued` (created), `landed`, `failed` (the last verification
  failed and the task is not verified again), and `all`. `tab` cycles it.
- A table: ID, title, project, status (`landed` per the README rule), the
  acting agent, and the verification state (`pass` and the commit, `V<n>
  fail`, or `running`).
- The selected task shows a detail panel: the title, `needs` (dependencies
  with marks), `blocks` (tasks that depend on it), branch, feature, grant of
  its goal, `needs user` text, and an evidence list: the files in
  `.hive/spec/<feature>/tasks/<id>/evidence/` of the home checkout, with
  `VERIFIED.md` first, and the verification history with reason and time.
- `enter` shows the selected evidence file in a scroll view. `esc` closes it.
  `b` shows the brief the same way. `p` cycles the project filter.

### R2 Friction

- Read `reference/dogfood/friction/FRICTION.md` and the entry files in
  `entries/` and `entries/resolved/` of the home.
- A tab bar: `open` and `resolved` with counts. A table: opened date, title,
  and the decision text. Resolved rows are dim, newest first.
- The selected entry shows: title, status, opened date, file name, and the
  sections of the entry body (summary, reproduction steps, and the others in
  order). `enter` shows the full entry in a scroll view.
- A file without valid front matter shows as one row with the mark `?` and
  the file name. It does not stop the screen.

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

`d deps graph`, `v open`, `n new entry`, `e edit`, `g ask coordinator to fix`
(show `ask in Chat`). `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each new test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Live, in a new herdr workspace: run the task binary with `HIVE_HOME` set to the real home. Press `5`, `tab` to `landed`, select `79-tui-shell`, `enter`, `esc`, `b`, `esc`, `6`, `tab`, `enter`. Read the pane after each step. | Task 79 shows `landed`, its V1 fail in the history, and its `VERIFIED.md`. Friction shows the open and resolved entries. Clean up the workspace. |
| V3 | `git diff dev --stat` | Only the files of the source table and the evidence changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `feat(tui): add the Tasks and Friction screens`. Push the task
branch. Write the evidence. Run `hive task complete`. Send the commit SHA
and the V1-V4 results to the manager with `hive inbox send`.

## Scope change 1

Hive does not keep a failed verification in `verification_history` when the
next verification passes. Task 79 therefore has an empty history. Change V2:
task 79 shows `landed`, its `pass` verification, and its `VERIFIED.md`. A test fixture with history entries tests the history display. The live home
does not.
