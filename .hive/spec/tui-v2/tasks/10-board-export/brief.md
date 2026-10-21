# The board exporter and the golden grids

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, wave 2 (Harness) |
| Tier | decent |
| Depends on | the cell grid task. Base: `dev` after they land |
| Grant | `grants.md 2026-10-21 tui-v2 delivery on dev and PR 47 update` |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Decision 41: the exporter runs the script of each board and writes its grid.
This task builds the exporter, runs it once, and commits the 21 golden grids.
A person reviews the grids at the end of wave 2. The gate never runs the
exporter.

## Known facts

- Each board in `.hive/spec/tui-v2/design/` has one
  `<script type="text/x-dc" data-dc-script>`. The script builds a grid with
  `T(w, h)`: each cell is `{c, k, b}`, with a rune, a token name, and a bold
  flag. `class Component extends DCLogic` returns the rows in `renderVals()`.
- `support.js` is not in the folder. The exporter does not need it.
- Node.js is on this machine. The gate does not need Node.js.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/boardexport/` and its tests | AC1, AC2. A new package. |
| `internal/boardexport/cmd/boardexport/main.go` | AC1. The entry for `go run`. It is not part of the `hive` binary. |
| `.hive/spec/tui-v2/design/*.grid` | AC3. New files. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/tui-v2/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files. Do not edit a board.

## Requirements

### AC1 The exporter

- `go run ./internal/boardexport/cmd/boardexport DIR` reads each `.dc.html`
  in DIR, runs its script once with Node.js, takes the cell grid of `T`, and
  writes `<Board>.grid` with `internal/lib/cellgrid`.
- The exporter uses the cells of `T`, not the merged runs of `rows()`.
- The commit that adds the exporter states this reason: "board exporter for
  the tui-v2 golden grids. It is a development tool and is not part of the
  hive binary."
- When Node.js is missing, the exporter fails with a clear error.

### AC2 Board checks (decisions 13 and 26)

- The exporter fails a board that uses a symbol outside the list of decision
  13, a token outside decision 26, or a size other than the spec table. The
  error names the board, the cell, and the value.
- A Go test of the checks uses small hand-written grids. It does not need
  Node.js.

### AC3 The golden grids

- Run the exporter once. Commit the 21 grid files.
- A gate test reads each grid file, checks its size against the spec table,
  and checks each symbol. It runs no Node.js.
- First failing test: the gate test of the grid files, before the files
  exist.
- Write a review aid in the evidence: for each board, the `t|` text of its
  grid. The person reviews this at the end of wave 2.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The exporter writes one grid for each board from the cells of `T` | V1, V2 |
| AC2 | The exporter fails a board with a bad symbol, token, or size | V1 |
| AC3 | The gate checks the 21 committed golden grids | V1, V2 |

## Verification items

### V1 Tests

- Proves: AC1, AC2, AC3.
- Command: run each new test against the base code, then against the task code.
- Expected: each new test fails on the base and passes on the task branch.

### V2 Export again

- Proves: AC1, AC3.
- Command: run the exporter into a temp folder. Compare each new grid with
  the committed grid. Compare the `t|` lines of `Main.grid` with the text of
  the `rows()` output of the `Main.dc.html` script, row by row, for five rows.
- Expected: no difference in the grids. The five rows match.

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

The golden check of a screen. Each wave 6 task adds it.

## Changes

- 2026-10-21, the manager: the commit reason of AC1 is now "board exporter
  for the tui-v2 golden grids. It is a development tool and is not part of
  the hive binary." The old text had a semicolon. The STE check counts a
  semicolon.
- 2026-10-21, the manager: the AC3 row of the acceptance criteria is now in
  the active voice. The meaning does not change.
