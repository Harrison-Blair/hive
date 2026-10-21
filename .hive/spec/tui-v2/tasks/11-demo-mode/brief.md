# Demo mode and the fixture schema

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, wave 2 (Harness) |
| Tier | strong |
| Depends on | the no-startup task and the cell grid task. Base: `dev` after they land |
| Grant | `grants.md 2026-10-21 tui-v2 delivery on dev and PR 47 update` |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Decisions 35, 36, 37, and 44 add demo mode. `hive --demo FILE` draws the TUI
from a fixture file, with the same renderers as normal mode. The golden check
and the live check read the same fixture. This task defines the fixture
schema, the demo loader, `--demo`, and `--size`, and writes two fixtures. The
next task writes the other fixtures.

## Design

- `/home/penguin/source/hive/.hive/spec/tui-v2/design/Main.dc.html`
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/TooSmall.dc.html`

This task draws the current screens with fixture data. It does not change the
look of a screen to match a board. The visual check checks that demo mode
and normal mode draw the same screen from the same data. A difference from a
board that the current screens also show is not a fail of this task. Record
it.

## Manager decisions

- Fixture file: `<Board>.fixture.json` in `.hive/spec/tui-v2/design/`. The
  worker defines the schema and records it in the evidence and in the
  section `Manager decisions` of the spec.
- The demo loader is in `internal/tui/`. Choose a file or a subpackage and
  record the choice.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `cmd/root.go` and its tests | AC3, AC4. |
| `internal/tui/` and its tests | AC1-AC5. |
| `.hive/spec/tui-v2/design/Main.fixture.json`, `TooSmall.fixture.json` | AC2. New files. |
| `.hive/spec/tui-v2/README.md` | AC1. Only the section `Manager decisions`. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/tui-v2/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### AC1 The fixture schema (decision 44)

- The schema holds the data of each screen: agents, goals, tasks, friction
  entries, checks, chat messages, log events, and the text of each task brief
  and evidence file that a screen can open. It holds the clock with its
  time zone, the home path to show, and the terminal size.
- It holds the UI state: the screen, the view, the focus, the selection, each
  scroll position, the filter, and the open overlay, popup, or prompt.
- Make the UI state cover each board in the spec table, also for screens and
  views that do not exist yet. Record each field and the board that needs it.
- An unknown field fails the load with an error that names the field.

### AC2 Two fixtures

- `Main.fixture.json` holds the data and the UI state of `Main.dc.html`.
- `TooSmall.fixture.json` holds the size and the state of `TooSmall.dc.html`.
- Take each value from the board script. Record the source line of each
  value group in the evidence.

### AC3 Demo mode (decision 35)

- `hive --demo FILE` implies `--no-startup`. It reads all screen data from
  the fixture. It uses the fixture clock and does not advance it.
- It runs no `hive doctor` and no other process. It reads no live record and
  no live file, and writes no record. A task brief or an evidence file comes
  from the fixture text. When the fixture has no text for a file, the screen
  shows the notice `demo mode: no file` and reads nothing. It writes log events as usual, with `demo` set to true.
  The Logs screen shows only the log events of the fixture.
- Chat send, restart, attach, and peek show the notice `demo mode: no action`
  and do nothing.
- Help does not list the flag.
- The demo loader and the normal loader feed the same screen renderers. Only
  the loader differs (decision 37).
- First failing test: a CLI entry test runs `--demo` with `Main.fixture.json`,
  presses the chat send, restart, attach, and peek keys, and checks that no
  file other than the log file changes and that no process starts.

### AC4 Size (decision 36)

- `hive --demo FILE --size WxH` draws the TUI at exactly W×H from the
  top-left cell. When the pane is smaller, the TUI shows the too-small
  notice.
- A bad size value fails with an error that names the value.

### AC5 Normal mode stays the same

- Each existing TUI test passes with no change to its expected value.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The schema holds the data and the UI state of each board | V1, V3 |
| AC2 | Two fixtures match their boards | V3 |
| AC3 | Demo mode reads only the fixture and does no action | V1, V2 |
| AC4 | `--size` draws at the given size or shows the too-small notice | V1, V2 |
| AC5 | Normal mode does not change | V1 |

## Verification items

### V1 Tests

- Proves: AC1, AC3, AC4, AC5.
- Command: run each new test against the base code, then against the task code.
- Expected: each new test fails on the base and passes on the task branch.
  Each old TUI test passes.

### V2 Live check

- Proves: AC3, AC4.
- Command: in a new herdr pane of 130×45 or larger, run
  `COLORTERM=truecolor hive --demo .hive/spec/tui-v2/design/Main.fixture.json --size 120x40`
  with a temp `HIVE_HOME`. Read the pane. Then run it in a pane smaller
  than 120×40.
- Expected: the TUI draws in the top-left 120×40 cells and the clock does
  not change. The small pane shows the too-small notice. No record changes.
  Close the panes.

### V3 Schema review

- Proves: AC1, AC2.
- Command: read the schema record in the evidence against the spec table.
- Expected: each board has the fields that it needs.

### V-gate Diff scope and gate

- Proves: each criterion.
- Command: `git diff dev...HEAD --stat`, after you merge `dev` into the task
  branch. Then the full gate of `AGENTS.md` from the worktree root.
- Expected: only the files of the source table and the evidence changed.
  `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0.
  Put the full gate output in the evidence.

## Rules

- Read `.hive/spec/tui-v2/README.md` first. When the code, a board, or the
  spec does not agree with this brief, stop and report it with
  `hive inbox send --to manager-hive TEXT --agent NAME`.
- Each Go change is test-first. The brief names the first failing test. See
  it fail for the correct reason.
- Do not run bare `hive` in herdr without `--no-startup` or
  `HIVE_NO_STARTUP=1`. Before task `no-startup` lands, do not run bare `hive`
  in herdr. Refer to the friction entry `scratch-hive-starts-real-agents`.
- A live check never changes the real home. Close each workspace, tab, and
  pane that the check makes.

## Out of scope

The other 19 fixtures. The new screens. The golden check of a screen.

## Changes

- 2026-10-21: after a verify fail (t11-verifier-1), the fixture holds the
  text of task brief and evidence files (AC1). Demo mode reads no live file
  (AC3). The Design section is new.
