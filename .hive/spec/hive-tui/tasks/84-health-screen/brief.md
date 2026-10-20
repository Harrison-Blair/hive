# Health screen and new checks

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, wave C |
| Tier | decent |
| Depends on | 80-chat-screen, 81-agents-screen. Base: `dev` after both land |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Fill screen 7, Health, and add two checks to `hive doctor`. The feature README is `.hive/spec/hive-tui/README.md`. Read it first,
mainly goal decision 6 and G5, G12, G14. The manager approved its decisions. Do not change a decision. When the
code does not agree with this brief or the README, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Sources

- The design in the home: `.hive/tmp/coordinator/designs/hive-tui/Health.dc.html`. The
  decisions win over the design. G12: the TUI has no action except chat,
  attach, peek, and quit. Where a design shows another action, show the hint
  `ask in Chat` in the footer, or leave the action out.
- The shell: `internal/tui/shell.go` (the `screen` interface, the data
  requests, attach). The Chat and Agents screens are examples.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/` (new files for the screens, and the shell wiring) and its tests | The R items. |
| `internal/doctor/` and its tests | R1. |
| `testdata/ste-baseline.json` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### R1 New checks

Add to `hive doctor` (text and `--json`), each with a status `ok` or `warn`
and a detail:

- `supervisor binary`: warn when the executable of the live supervisor
  process is older than the installed `hive` binary (the `hive` on `PATH`),
  or when its executable is gone (`(deleted)` in `/proc/<pid>/exe`). The
  detail gives both times.
- `stale worktrees`: warn for each worktree below `projects/*/worktrees/`
  whose task landed (README rule) or has status `cancelled`. The detail names each.
- `unpushed task branches`: warn for each task worktree whose branch has
  commits that its upstream does not have.

### R2 Health screen

- Panels `home` and `agents` with one line for each check of `hive doctor`
  and the new checks: `✓`, `!`, or `✕`, and a short text. The screen calls the
  doctor code in a `tea.Cmd`. It does not run the `hive` binary.
- The selected line shows its detail below the panels.
- A panel `capacity` says `later: capacity from qmeter. Not in this goal.`
- `r` checks again now. The footer shows `ask in Chat` for a restart or a
  cleanup. No `R`, `w`, or `c` action.

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

Restart of the supervisor. Worktree cleanup. Capacity. `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each new test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | In a temp home, add a fake supervisor claim for a process with an older executable than the newer `hive` on `PATH`. Add a worktree for a landed task. Run `hive doctor`. | Both new checks warn and name the cause. |
| V3 | Live, in a new herdr workspace: run the task binary with `HIVE_HOME` set to the real home. Press `7`, `↓`, `r`. Read the pane after each step. | Each check shows. Clean up the workspace. |
| V4 | `git diff dev --stat` | Only the files of the source table and the evidence changed. |
| V5 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `feat(doctor): add binary and worktree checks`. Push the task
branch. Write the evidence. Run `hive task complete`. Send the commit SHA
and the V1-V5 results to the manager with `hive inbox send`.
