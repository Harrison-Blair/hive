# Agents: draw each card from the live or newest record

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, goal `g6-g6-hive-ask`, wave A |
| Tier | strong |
| Depends on | task 126 (it lands first). Base: `dev` after task 126 |
| Grant | goal `g6-g6-hive-ask`: the user approved this fix on 2026-10-22. Build and land on `dev`. |
| Evidence | `.hive/spec/tui-v2/tasks/129-agents-stale-card/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The Agents overview draws some cards from old stale records, not from the
live record. Read the friction entry
`reference/dogfood/friction/entries/agents-card-shows-stale-record.md`. It
gives the evidence and the probable cause in `internal/tui/agents.go`:
`agentIdentity` puts one agent name into a "name" group and a "lineage"
group when the old records have no `parent` and no `replaces`. Then
`currentAgentKeys` marks a stale record as current too.

Find the cause with a failing test first. The friction entry gives the
probable cause only.

## Requirements

### AC1 One card for each agent name

- The overview shows one card for each agent name. The card uses the live
  record. When no record is live, the card uses the newest record.
- A stale record that a newer record replaces, or that is older than the live
  record, is not current.
- A stale record with no live successor shows as stale. `h` hides it with the
  stopped agents.
- The one-pod view, the details, and the Health › Agents view follow the same
  rule where they pick the current record.

### AC2 No board change

The fix changes no board. If the worker finds that a board must change, stop
and report to the manager before a golden file changes.

### AC3 Tests

- Write each test first and see it fail for the correct reason on the base.
- Use records with the real shape: copy the fields of
  `.hive/state/agents/manager-qmeter-g1..g4.json` and
  `.hive/state/agents/coordinator-g1..g8.json` in the home into test
  fixtures. Remove secrets and paths that are not needed. The old records
  have no `parent` and no `replaces`. The newest live record has
  `replaces`.
- Expected values are literal: the harness, the status word, and the glyph of
  each card, and which cards `h` hides.

### AC4 Friction entry

When the fix is in, set the friction entry status to `Resolved YYYY-MM-DD`,
move its file to `reference/dogfood/friction/entries/resolved/`, and update
its row in `FRICTION.md`. Do this on the task branch.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/agents.go`, `internal/tui/agents_screen.go`, `internal/tui/health.go`, and their tests | AC1, AC3. |
| `internal/tui/testdata/` | AC3. New fixtures. |
| `reference/dogfood/friction/FRICTION.md` and the entry file | AC4. |
| `testdata/ste-baseline.json` | To add the brief copy with 0, and to lower a count. |

Do not edit other files. Each grid and golden check that exists stays the
same.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | One card for each name, from the live or newest record | V1, V2 |
| AC2 | No board change | V-gate |
| AC3 | The tests fail on the base and pass on the task branch | V1 |
| AC4 | The friction entry has the status `Resolved` | V-gate |

## Verification items

### V1 Tests red and green

- Proves: AC1, AC3.
- Command: run the new tests on the base code, then on the task branch. Run
  each new test with `-count=50` and with `-race`.
- Expected: each fails on the base and passes on the task branch in each run.

### V2 Live check

- Proves: AC1.
- Command: make a scratch home in `.hive/tmp/` of the worktree. Copy the
  record fixtures of AC3 into its `.hive/state/agents/`. Run
  `HIVE_NO_STARTUP=1 COLORTERM=truecolor hive` in that home in a herdr pane
  at 120×40. Open Agents. Read the pane. Press `h` and read the pane again.
- Expected: one card for each name. The `manager-qmeter` card shows the
  harness and status of the live record. The coordinator card shows the live
  record. A stale record with no live successor hides with `h`. Put each
  capture in the evidence.

### V-gate Diff scope and gate

- Proves: each criterion.
- Command: `git diff dev...HEAD --stat`, after you merge `dev` into the task
  branch. Then the full gate of `AGENTS.md` from the worktree root.
- Expected: only the files of the source table and the evidence changed. No
  grid changed. `gofmt -l` and `go mod tidy -diff` print nothing. The other
  commands exit 0. Put the full gate output in the evidence.

## Rules

- Do not run bare `hive` in herdr without `--no-startup` or
  `HIVE_NO_STARTUP=1`. Never run it on the real home for this task.
- Close only the panes and the processes that you started.
- Write a removal of a path in a variable as `rm -f "${VAR:?}"/...`, or use a
  literal path.
- Remove each scratch file that you make in `.hive/tmp/` after the check.

## Finish

Commit on the task branch with a subject such as `fix(tui): draw each agent
card from the live record`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1 and V2 results to the
manager with `hive inbox send --to manager-hive`.
