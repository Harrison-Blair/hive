# General agents on the Agents graph

| Item | Value |
| --- | --- |
| Feature | `general-agent`, wave F (TUI) |
| Tier | strong |
| Depends on | 33-screen-agents (landed). Base: `dev` |
| Grant | `grants.md 2026-10-18 hive delivery on dev` |
| Evidence | `.hive/spec/general-agent/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Decision 39 of the feature `general-agent` puts each general agent on the
Agents graph and in Health › Agents. The user gave the rules of the boards on
2026-10-21. This task adds these rules to the Agents renderer of task
`33-screen-agents`. It adds one fixture for each new board. It writes the
target rows of each new board. The coordinator then paints the boards from
the target rows. After the user approval, this task also copies the boards, exports the
grids, and adds the golden checks.

## Design

- `/home/penguin/source/hive/plans/general-agent/home-boards-decisions.md`
  gives the user decisions and the review details.
- The boards `Main`, `AgentsMany`, `CompactAgents`, `AgentDetails`, and
  `HealthAgents` in `.hive/spec/tui-v2/design/` give the current forms.

No new board exists yet. This task makes the target rows for the new boards.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/` and its tests | AC1-AC3. Only the Agents graph, the details popup, and Health › Agents. |
| `.hive/spec/tui-v2/design/*.fixture.json` | AC4. Seven new fixtures. Do not change a current fixture. |
| `.hive/spec/tui-v2/design/*.dc.html`, `canvas.json`, `*.grid` | AC5. The eight approved boards and their grids. |
| `.hive/spec/tui-v2/README.md` | AC5. Only seven new rows in the board table. |
| `internal/boardexport/check.go` and its tests | AC5. Only seven new entries in `Sizes`. |
| `internal/tui/` tests that count the boards | AC5. Only the count of boards. |
| `.hive/spec/general-agent/README.md` | Only the section `Manager decisions`, when it exists, or a new section with that name. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/general-agent/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files. Do not change the logic of the exporter. Copy each
board. Do not
change a board by hand.

## Requirements

### AC1 The home column (decision 39, user decisions 1 and 5)

- Each general agent that has no parent card on the graph is in one last
  column with the title `home`. The column has the same width as a project
  column. It is not shown when it is empty. It is not in the project count.
- With no project, the graph keeps the notice `no project agents`. Home
  agents show only in Health › Agents.
- From 80 to 119 columns, home is the last page of `←` `→`.

### AC2 Children and nesting (user decisions 2, 3, and 4)

- A general agent is under the card of its recorded parent when that card is
  on the graph. This rule wins over the home rule.
- The children of a card are in order of creation time, for each role.
- A child of a child is on a rail below its parent, with 2 cells more indent
  for each level. A column with nesting uses the rail form.
- When `h` hides a stopped parent, its live child moves to home. It moves
  back when `h` shows the parent.
- A general agent uses the card of 24×4 cells: the dot and the name, then
  harness · state · tail. The text `no workers` becomes `no child agents`.
- A child of the coordinator is in home, because the coordinator is not on
  the graph.

### AC3 Popup and Health (user decision 6)

- The details popup of a general agent has the same fields as the popup of
  another agent. A field with no value is not shown. No new field.
- Health › Agents lists each general agent with the role `agent`. `r` on it
  shows the notice `restart: coordinator and supervisor only` and does
  nothing.

### AC4 Fixtures and target rows

- Add one fixture for each new board, in the format of the current fixtures:
  `AgentsHome` (120×40), `AgentsNested` (120×40, a nested child of a worker),
  `AgentsHiddenParent` (120×40, a child moved to home by `h`),
  `AgentsCoordinatorChild` (120×40), `CompactAgentsHome` (80×24, the home
  page), `AgentDetailsGeneral` (120×40, the popup of a general agent), and
  `HealthAgentsGeneral` (120×40, a general agent selected). Each time in a
  fixture is at or before the fixture clock.
- Render each fixture with the demo loader at its size. Write the rows to
  `/home/penguin/source/hive/.hive/tmp/reports/general-agent-home-boards.md`
  in the format of `.hive/tmp/reports/tui-v2-wave-f-redraw/33-screen-agents.md`:
  for each board, each row as plain text with the token and the bold flag of
  each run. Do not commit the report.
- A rule of AC1-AC3 can change a cell of a current board, for example the
  text `no child agents` or the order of children. Then render that current
  board too and add its changed rows to the report, with the cells that
  change. The coordinator repaints it with the new boards. AC5 adds the
  new grids in this task.
- First failing test: a model test that a general agent with no parent card
  is in the `home` column.

### AC5 Boards, grids, and golden checks

- On 2026-10-22 the user approved the seven new boards and the change of
  `AgentsMany` (artifact version 45). Copy the eight boards and `canvas.json`
  from `/home/penguin/source/hive/plans/tui-v2/design/` over the copies in
  the worktree. `cmp` gives no difference.
- Add the seven new boards with their sizes to the board table of the
  `tui-v2` spec, to `Sizes` in `internal/boardexport/check.go`, and to each
  test that counts the boards. Change only data, not logic.
- Run the exporter of task `10-board-export` on the design folder. Commit the
  seven new grids and the changed `AgentsMany.grid`. Each other grid stays
  the same.
- Add a golden check for each of the seven new boards, as for the current
  boards. Each golden check passes.
- The coordinator found one error in the report: section (a) gives no scroll
  mark for `AgentsNested`, but its rows show a cut pod and a `┃` mark on
  column 119, rows 4-34. The rows and the board win. Correct the report and
  any text that gives the old claim.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The home column follows the rules | V1, V2 |
| AC2 | Children and nesting follow the rules | V1, V2 |
| AC3 | The popup and Health follow the rules | V1, V2 |
| AC4 | Seven fixtures load, the report gives their rows and each changed current row, and each golden check passes | V1, V2 |
| AC5 | The board copies equal the approved boards, and each golden check passes | V1, V-gate |

## Verification items

### V1 Tests

- Proves: AC1-AC4.
- Command: run each new test against the base code, then against the task
  code. Then break each rule of AC1-AC3 on purpose and run the tests again.
- Expected: each new test fails on the base and passes on the task branch.
  A new test that cannot compile on the base, because it calls new code, is
  red on the base when a break of that code fails it. Each deliberate break
  fails a test.

### V2 Live rows

- Proves: AC1-AC4.
- Command: for each new fixture, in a scratch pane, run
  `COLORTERM=truecolor hive --demo FILE --size WxH`. Capture the pane as
  decision 43 of `tui-v2` gives. Compare the capture with the rows of the
  report. Press `h`, `←`, `→`, `enter`, and `r` where they apply.
- Expected: no cell differs, under the space cell rule of `tui-v2`. The keys
  act as AC1-AC3 give.

### V-gate Diff scope and gate

- Proves: each criterion.
- Command: `git diff dev...HEAD --stat`, after you merge `dev` into the task
  branch. Then the full gate of `AGENTS.md` from the worktree root.
- Expected: only the files of the source table and the evidence changed.
  `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0.
  Put the full gate output in the evidence.

## Rules

- Read `.hive/spec/general-agent/README.md` and `.hive/spec/tui-v2/README.md` first. When the code, a board, or the
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

The board drawings.

## Changes

- 2026-10-22: AC5 adds the seven boards to the spec table, to `Sizes`, and to
  the board counts.
- 2026-10-22: AC5 adds the approved boards, their grids, and the golden
  checks to this task.
- 2026-10-21: AC4 allows a current board to change by a user rule. The report
  gives its rows, and the task completes after its grid is on `dev`.
