# Agents screen

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, wave 6 (Screens) |
| Tier | strong |
| Depends on | 29-frame-layout (landed). Base: `dev` |
| Grant | `grants.md 2026-10-21 tui-v2 delivery on dev and PR 47 update` |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Wave 6 of the spec puts each new screen into the registry of task
`29-frame-layout`. This task makes the content of the tab 1 Agents. It
replaces the current screen of that content area. Other wave 6 tasks run at
the same time and change other entries of the same registry.

## Design

  - `/home/penguin/source/hive/.hive/spec/tui-v2/design/Main.dc.html` and `Main.grid`
  - `/home/penguin/source/hive/.hive/spec/tui-v2/design/AgentsMany.dc.html` and `AgentsMany.grid`
  - `/home/penguin/source/hive/.hive/spec/tui-v2/design/AgentDetails.dc.html` and `AgentDetails.grid`
  - `/home/penguin/source/hive/.hive/spec/tui-v2/design/CompactAgents.dc.html` and `CompactAgents.grid`
  - `/home/penguin/source/hive/.hive/spec/tui-v2/design/Help.dc.html` and `Help.grid`

The visual check compares the live screen with each board above.

## Manager decisions

- Put the screen in its own file in `internal/tui/`, named for the screen.
  Change only the registry entry of this tab in `frame.go`.
- Use the components of `internal/tui/component/`. You can make a component
  fix that the screen needs. Record each one.
- Remove the code of the current screen of this tab when no other code uses
  it: `agents.go` and its tests. Keep each behavior that a decision of the spec does not
  change.
- A live check does not compare the token or the bold flag of a space cell.
  Refer to the spec section `Manager decisions`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/` and its tests | AC1-AC4. Only this screen, its registry entry, and the removal of its current screen. |
| `internal/tui/component/` and its tests | Only a fix that the screen needs. Record each one. |
| `.hive/spec/tui-v2/README.md` | Only the section `Manager decisions`. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/tui-v2/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### AC1 The screen (decisions 10, 14, 19)

- A full-width graph of project agents only, one column of equal width for
  each project, cards of 24×4 cells, the idle dot `◐` in `accent`
  (decisions 10, 14). The coordinator and the supervisor are not on the graph.
  With no project, the notice `no project agents`.
- When the columns do not fit, the graph scrolls sideways and keeps the
  selected card visible.
- `enter` opens the details popup. `p` peeks, `a` attaches, `h` hides stopped
  agents. The graph and the popup have no `r` key (decision 19).
- Turn on the full `Help` grid check that task `29-frame-layout` limited to the
  help box, because the Agents screen is now behind it.

### AC2 Golden checks (decisions 42, 46)

- For each board of this task (`Main`, `AgentsMany`, `AgentDetails`, `CompactAgents`, `Help`), build the model from its fixture
  with the demo loader. Render `View()` at the board size and compare the
  whole grid with the board grid, with `internal/lib/cellgrid`.
- Render the screen at 79×24, 80×24, 119×40, and 120×40 and check the
  properties of decision 46: no line wider than the pane, no broken border,
  and the selected item visible. Also test empty data and text longer than
  its column.
- First failing test: the golden check of `Main`.

### AC3 Normal mode (decisions 37, 38)

- Normal mode and demo mode use the same renderer. Only the loader differs.
- Keep or add tests of normal mode in a scratch home with `--no-startup`:
  the record loaders and refresh of this screen. Use no live coordinator and
  no live supervisor.

### AC4 Keys

- Each key of this screen follows decisions 21 to 23. The footer shows the
  key hints of each focus state of this screen.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The screen follows decisions 10, 14, 19 | V1, V2 |
| AC2 | Each board grid matches, and the screen fits each size | V1, V2 |
| AC3 | Normal mode uses the same renderer and keeps its tests | V1 |
| AC4 | The keys follow decisions 21 to 23 | V1, V2 |

## Verification items

### V1 Tests

- Proves: AC1-AC4.
- Command: run each new test against the base code, then against the task
  code. Then change one token and one key of this screen on purpose and run
  the tests again.
- Expected: each new test fails on the base and passes on the task branch.
  Each deliberate change fails a test.

### V2 Live check (decision 43)

- Proves: AC1, AC2, AC4.
- Command: for each board of this task, in a scratch pane, run
  `COLORTERM=truecolor hive --demo FILE --size WxH` with its fixture. Capture
  the pane as decision 43 gives, and compare it with the board grid. Press the
  keys of this screen.
- Expected: no cell differs, under the space cell rule. The keys act as AC1
  and AC4 give.

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
- In a cleanup, do not use `rm -rf` on a shell variable without `"${VAR:?}"`.
  Push in a command of its own.

## Out of scope

The other screens.

## Changes

None.
