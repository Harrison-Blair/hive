# Frame and layout

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, wave 5 (Frame and layout) |
| Tier | strong |
| Depends on | 23-components-2, 28-board-data-fix (landed). Base: `dev` |
| Grant | `grants.md 2026-10-21 tui-v2 delivery on dev and PR 47 update` |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Wave 5 of the spec makes the frame of the new TUI: the header, the tab bar,
the content area, and the footer of each screen. It adds the focus, the key
routes, the breakpoints, the overlay below 120 columns, the too-small notice,
and the help overlay. This task owns the root-model wiring and the screen
registry. Wave 6 puts one screen in the registry for each task.

## Design

- Each board in `/home/penguin/source/hive/.hive/spec/tui-v2/design/`. The
  header rows, the tab bar rows, and the footer row of each screen board are
  the expected frame.
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/Help.dc.html` and
  `Help.grid`, the key map and the help overlay.
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/TooSmall.dc.html` and
  `TooSmall.grid`.

The visual check compares the frame rows of the live TUI with each board, and
the help overlay and the too-small notice with their boards. The content area
is not part of this task.

## Manager decisions

- The registry gives each of the six screens a number, a label, its views,
  and its key hints for each focus state. A screen gives the content of the
  content area for a size and a focus. Choose the exact API and record it in
  the spec section `Manager decisions`.
- Until wave 6 replaces a screen, the content area shows the current screen
  of that tab. This task removes no current screen. Record which
  current screen each tab shows.
- The frame uses the components of `internal/tui/component/`.
- A live check compares a capture of a pane with a board grid. In a space
  cell of a live capture, the check does not compare the token or the bold
  flag, because a terminal does not show them. It compares each other cell
  fully. The golden check at the model compares each cell fully. Add this
  rule to the spec section `Manager decisions`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/` and its tests | AC1-AC6. |
| `internal/tui/component/` and its tests | Only a fix that the frame needs. Record each one. |
| `cmd/root_test.go`, `cmd/root_demo_test.go` | AC6. Only the expected header text and the demo key sequence that the new frame changes. |
| `.hive/spec/tui-v2/README.md` | Only the section `Manager decisions`. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/tui-v2/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### AC1 The frame (decisions 1 to 4)

- Each screen has the fixed header, the tab bar, the content area, and the
  fixed footer. The header and the footer never scroll.
- The header has the long form from 120 columns and the short form below 120.
  The tab bar follows decision 3. The footer shows `[region]` and the key
  hints of the focus state (decision 21).

### AC2 Focus and keys (decisions 6, 21 to 24)

- One region has the focus, with an `accent` border. `tab` moves the focus.
- Each key follows the route table of decision 21 for each focus state, and
  decisions 22 to 24. Each key has one meaning.
- `ctrl+c` two times quits, in each focus state.
- While an overlay or a popup is open, a screen key such as `1` does not
  change the screen (decision 21).
- The restart prompt keeps the behavior of task `3-tui-restart`. Its footer
  follows decision 21.

### AC3 Breakpoints (decisions 7 to 9, 11)

- Below 80×24 the TUI shows the too-small notice of decision 7.
- From 120 columns a list and its details sit side by side. From 80 to 119
  columns `enter` opens the details as a full-width overlay, and `esc` closes
  it.
- A resize never breaks a border. A resize with an overlay open keeps the
  overlay and its scroll position.

### AC4 Help overlay

- `?` opens the help overlay of the Help board. `?` or `esc` closes it. `↑`
  and `↓` scroll it. The screen behind it uses `border` (decision 5). Help
  does not list `--demo`.

### AC5 Golden checks

- With the fixture of each screen board, render the TUI in demo mode at the
  board size. Compare the header rows, the tab bar rows, and the footer row
  with the same rows of the board grid, with `internal/lib/cellgrid`.
- Compare the whole grid for `TooSmall`. For `Help`, compare the header, the
  tab bar, the footer, and the help box (rows 6-34, columns 18-101) cell for
  cell. In each other cell of the content area, compare only the token,
  because the Agents screen behind it is wave 6 work. The wave 6 Agents task
  turns on the full `Help` check.
- Render each screen at 79×24, 80×24, 119×40, and 120×40, and check the
  properties of decision 46 that the frame owns: the too-small notice at 79
  columns, no line wider than the pane, and no broken border.
- First failing test: the frame rows of `Main` against `Main.grid`.

### AC6 Normal mode

- Normal mode and demo mode use the same frame and the same screens. Only the
  loader differs (decision 37). Each old TUI test passes, or the evidence
  names the decision that changed its expected result.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The frame follows decisions 1 to 4 | V1, V2 |
| AC2 | Focus and keys follow decisions 6 and 21 to 24 | V1, V2 |
| AC3 | Breakpoints follow decisions 7 to 9 and 11 | V1, V2 |
| AC4 | The help overlay matches its board | V1, V2 |
| AC5 | The frame rows match each board | V1 |
| AC6 | Normal mode uses the same frame | V1 |

## Verification items

### V1 Tests

- Proves: AC1-AC6.
- Command: run each new test against the base code, then against the task
  code. Then change one key route and one frame token on purpose and run the
  tests again.
- Expected: each new test fails on the base and passes on the task branch.
  Each deliberate change fails a test.

### V2 Live check

- Proves: AC1-AC4.
- Command: in a scratch pane, run `hive --demo FILE --size WxH` with the
  `Main`, `Help`, `CompactAgents`, and `TooSmall` fixtures. Press `tab`, `?`,
  `esc`, `enter` at 80 columns, and `ctrl+c` two times. Capture the pane with
  the command of the verifier skill.
- Expected: the frame rows agree with each board. The keys act as AC2 to AC4
  give.

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
  `HIVE_NO_STARTUP=1`. Refer to the friction entry
  `scratch-hive-starts-real-agents`.
- A live check never changes the real home. Close each workspace, tab, and
  pane that the check makes.

## Out of scope

The content of each screen. Wave 6 does it.

## Changes

- 2026-10-21: AC5 compares only the token behind the Help box, because the
  Agents screen is wave 6 work.
- 2026-10-21: the source table allows two `cmd/` tests, for the header text
  and the demo keys only.
- 2026-10-21: after a verify fail (t29-verifier-1), AC2 states the overlay
  rule for screen keys, and a manager decision sets the live compare of space
  cells.
