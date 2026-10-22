# Health screen and the restart move

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, wave 6 (Screens) |
| Tier | strong |
| Depends on | 29-frame-layout (landed). Base: `dev` |
| Grant | `grants.md 2026-10-21 tui-v2 delivery on dev and PR 47 update` |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Wave 6 of the spec puts each new screen into the registry of task
`29-frame-layout`. This task makes the content of the tab 5 Health. It
replaces the current screen of that content area. Other wave 6 tasks run at
the same time and change other entries of the same registry.

## Design

  - `/home/penguin/source/hive/.hive/spec/tui-v2/design/Health.dc.html` and `Health.grid`
  - `/home/penguin/source/hive/.hive/spec/tui-v2/design/HealthAgents.dc.html` and `HealthAgents.grid`
  - `/home/penguin/source/hive/.hive/spec/tui-v2/design/CompactHealth.dc.html` and `CompactHealth.grid`

The visual check compares the live screen with each board above.

## Manager decisions

- Put the screen in its own file in `internal/tui/`, named for the screen.
  Change only the registry entry of this tab in `frame.go`.
- Use the components of `internal/tui/component/`. A component fix that the
  screen needs is allowed. Record each one.
- Remove the code of the current screen of this tab when no other code uses
  it: `health_screen.go` and its tests. Keep each behavior that a decision of the spec does not
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

### AC1 The screen (decisions 18, 19, 21)

- The views Home and Agents (decision 18). Home lists the checks of
  `hive doctor`. Agents lists each agent, including the coordinator and the
  supervisor.
- The restart move of decision 19: `r` on the coordinator or the supervisor
  row restarts with the behavior of task `3-tui-restart`, and asks the user
  only when the agent is live. On a manager or worker row, `r` shows the
  notice `restart: coordinator and supervisor only` and does nothing. The
  restart prompt footer follows decision 21.
- The restart uses the coordinator restart rule of decision 34 of the
  feature `general-agent`, as it is on `dev`.

### AC2 Golden checks (decisions 42, 46)

- For each board of this task (`Health`, `HealthAgents`, `CompactHealth`), build the model from its fixture
  with the demo loader. Render `View()` at the board size and compare the
  whole grid with the board grid, with `internal/lib/cellgrid`.
- Render the screen at 79×24, 80×24, 119×40, and 120×40 and check the
  properties of decision 46: no line wider than the pane, no broken border,
  and the selected item visible. Also test empty data and text longer than
  its column.
- First failing test: the golden check of `HealthAgents`.

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
| AC1 | The screen follows decisions 18, 19, 21 | V1, V2 |
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
