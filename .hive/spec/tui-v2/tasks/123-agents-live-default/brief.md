# Agents: stopped hidden by default, green working cards

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, goal `g4-agents-live-default`, wave A |
| Tier | strong |
| Depends on | none. Base: `dev` |
| Grant | goal `g4-agents-live-default`: user chat c50 to c69, 2026-10-22. Land on `dev`. |
| Evidence | `.hive/spec/tui-v2/tasks/123-agents-live-default/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The user approved new Agents boards on 2026-10-22 (canvas version 64, id
1792703555-2c9e). This task copies the boards into the spec, exports their
golden grids again, adds the decisions to the tui-v2 spec, and builds the
Agents screen to match the boards:

1. Stopped agents are hidden at each TUI start. `h` shows and hides them.
2. A working card has a border in the token `working`.
3. The footer says `h stopped`. The Help line for `h` says
   `show or hide stopped agents`.

The TUI must match each board cell for cell: the glyph, the colour token, and
the bold flag. When a board and this brief do not agree, stop and report it
with `hive inbox send --to manager-hive TEXT --agent NAME`. Do not choose for
the user.

## Design

The approved boards, copied from the canvas "Hive TUI Redesign"
(https://claude.ai/artifact/VrhcA98g9G2G9cFQao5Z9h, version
1792703555-2c9e), in
`/home/penguin/source/hive/.hive/tmp/g4-agents/boards/`:

- `Main.dc.html` (120×40, after `h`: stopped shown)
- `CompactAgents.dc.html` (80×24, after `h`)
- `AgentsMany.dc.html` (120×40, after `h`)
- `AgentsHome.dc.html` (120×40, after `h`)
- `AgentsNested.dc.html` (120×40, stopped shown: the after-`h` state)
- `AgentsHiddenParent.dc.html` (120×40, the default at start: stopped hidden)
- `AgentsCoordinatorChild.dc.html` (120×40, after `h`)
- `CompactAgentsHome.dc.html` (80×24)
- `AgentDetails.dc.html` (120×40, after `h`)
- `AgentDetailsGeneral.dc.html` (120×40, after `h`)
- `Help.dc.html` (120×40)

Follow the visual check rule in the worker skill. The user is colourblind.
Each state has a glyph and a word. Colour only adds to a state.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/tui-v2/design/<Board>.dc.html` and `<Board>.grid` for the 11 boards | AC1. Copies and grids. |
| `.hive/spec/tui-v2/design/<Board>.fixture.json` for the 11 boards | AC1. Only when a board needs a change of its UI state. |
| `.hive/spec/tui-v2/README.md` | AC2. Decisions 58 to 61. |
| `internal/tui/agents_screen.go`, `internal/tui/agents.go`, `internal/tui/frame.go`, `internal/tui/component/card.go`, `internal/tui/demo.go`, and their tests | AC3 to AC6. |
| `internal/tui/tui.go`, `internal/tui/component/*.go` | Only when AC3 to AC6 need a change there. Record the cause in the evidence. |
| `testdata/ste-baseline.json` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/tui-v2/tasks/123-agents-live-default/brief.md` | The brief copy, when `hive task create` did not write it. |

Do not edit other files. Do not add a mask. Do not change the exporter. Do
not change Health › Agents.

## Requirements

### AC1 Boards and grids

- Copy the 11 boards over the copies in `.hive/spec/tui-v2/design/`. `cmp`
  against the staged copies gives no difference.
- Run the exporter `internal/boardexport/cmd/boardexport` on the design
  folder. Commit only the changed grids. The manager export gave changes in
  these grids only: Main, AgentsMany, AgentsHome, AgentsNested,
  AgentsHiddenParent, AgentsCoordinatorChild, CompactAgentsHome, and Help.
  CompactAgents, AgentDetails, and AgentDetailsGeneral do not change. Each
  other grid stays the same.

### AC2 Spec

Add these decisions to the tui-v2 README. Write them in ASD-STE100 Strict.
Keep the meaning of this brief.

- 58: the default of stopped agents on the Agents screen (AC3).
- 59: Health › Agents does not change. It keeps stopped rows by default and
  its own `h`.
- 60: the border of a working card (AC4).
- 61: the footer `h stopped` and the Help line (AC5).

### AC3 Decision 58: stopped agents hidden at start

- A stopped agent has the status `stopped` or `exited`
  (`internal/tui/agents.go:174`). The Agents graph hides each stopped agent
  at each TUI start: `graph.hideStopped` starts `true`. `h` toggles it.
- The stopped text of the pod stays: `h shows N stopped` while hidden, `h hides N stopped`
  while shown.
- An unresolved agent stays visible.
- Demo mode reads `hide_stopped` from the fixture, as it does now. A fixture
  without the field shows stopped agents. Only normal mode gets the new
  default.

### AC4 Decision 60: working card border

- A card of an agent in the state `working` draws its rounded border in the
  token `working`.
- A selected card keeps the heavy `accent` border, also while it works.
- Each other state (idle, blocked, stopped, and other) keeps the token
  `border`.
- Each joint that sits on the border of a working card takes the card token,
  for example `├`, `┤`, `┴`, `┬`, and `╰`. A rail or a tree line outside the
  card keeps `border`. The boards show each case.

### AC5 Decision 61: footer and Help

- The graph footer hint `h hide stopped` becomes `h stopped`
  (`internal/tui/frame.go:60`).
- `agentsScreen.KeyHints` in `internal/tui/agents.go:372` is not rendered.
  Change only `h hide stopped` to `h stopped` there and in its test.
- The Help line for `h` becomes `show or hide stopped agents`
  (`internal/tui/frame.go:491`).

### AC6 Tests

- Write each test first and see it fail for the correct reason.
- The golden check passes for each of the 11 boards and for each other
  board that passes on `dev`.
- Model tests cover: stopped hidden after a new start in normal mode, `h`
  shows them and `h` hides them again, the stopped text in both states, an
  unresolved agent visible while stopped agents are hidden, Health › Agents
  default unchanged, a working card border, a selected working card, an
  idle and a stopped card border, a joint on a working card, and a rail
  outside a working card. Use literal expected text and tokens.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Board copies equal the staged boards. Only the expected grids change. | V1, V2 |
| AC2 | The spec has decisions 58 to 61 | V-gate |
| AC3 | Stopped agents are hidden at start and `h` toggles them | V3, V4, V5 |
| AC4 | Working cards have the `working` border, joints included | V3, V4, V5 |
| AC5 | Footer and Help text | V3, V4 |
| AC6 | The tests fail on the base and pass on the task branch | V3 |

## Verification items

### V1 Boards

- Proves: AC1.
- Command: `cmp` of each staged board against its copy. Capture both with
  the Firefox command of the verifier skill. Put the profile folder and the
  captures in `.hive/tmp/` of the worktree. Remove them after the check.
- Expected: no difference. No capture shows `{{`.

### V2 Grid diff

- Proves: AC1.
- Command: compare each grid on `dev` with the task grid, cell by cell, with
  `internal/lib/cellgrid`. List the changed rows of each grid.
- Expected: only the grids that AC1 names change. Each change is a footer
  hint, the Help line, or a cell of the border of a working card.

### V3 Tests red and green

- Proves: AC3 to AC6.
- Command: run the new tests on the base code, then on the task branch.
- Expected: each fails on the base and passes on the task branch.

### V4 Live check

- Proves: AC3, AC4, AC5.
- Command: the live check of decision 43 for each of the 11 boards at its
  board size, with `COLORTERM=truecolor hive --demo FILE --size WxH`. Then,
  in the demo of `AgentsHiddenParent`, press `h`, read the pane, press `h`
  again, and read the pane.
- Expected: each capture equals its grid. After the first `h` the stopped
  cards show and the stopped text says `h hides 1 stopped`. After the second `h`
  the capture equals the grid again.

### V5 Visual alignment

- Proves: AC3, AC4.
- Command: run the real TUI in a herdr pane for these states, each at
  120×40 and at 80×24: the default at start (`AgentsHiddenParent`), after
  `h` (`Main`), working cards in a tree and on a rail (`AgentsHome`,
  `AgentsNested`), and the home page (`CompactAgentsHome`). Compare the
  capture at the board size with the board grid, cell by cell. Look at each
  capture and check each point:
  - Each panel border, card border, and corner lines up.
  - Each working card border is in the token `working` all the way round,
    with no gap at a joint.
  - Each selected card keeps the heavy `accent` border.
  - Each rail and tree line outside a card keeps `border` and meets the card
    joint with no gap.
  - The stopped text and the footer show the text of the boards.
- Expected: each capture at the board size equals its grid. Each point
  passes at both sizes. Put each capture, each diff, and a table of the
  points with pass or fail in the evidence. A mismatch is a fail.
- The worker and the verifier each do this item.

### V-gate Diff scope and gate

- Proves: each criterion.
- Command: `git diff dev...HEAD --stat`, after you merge `dev` into the task
  branch. Then the full gate of `AGENTS.md` from the worktree root.
- Expected: only the files of the source table and the evidence changed.
  `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0.
  Put the full gate output in the evidence.

## Rules

- Do not run bare `hive` in herdr without `--no-startup` or
  `HIVE_NO_STARTUP=1`.
- Never open the Chat input on the real home. Never send chat.
- Close only the panes and the processes that you started. A demo quits
  with two `ctrl+c` presses.

## Out of scope

Health › Agents. Each other screen. The canvas.

## Finish

Commit on the task branch with a subject such as `feat(tui): hide stopped
agents at start and mark working cards`. Push the task branch. Write the
evidence. Run `hive task complete`. Send the commit SHA and the V1 to V5
results to the manager with `hive inbox send --to manager-hive`.
