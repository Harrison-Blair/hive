# The other board fixtures

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, wave 2 (Harness) |
| Tier | decent |
| Depends on | 11-demo-mode (landed). Base: `dev` |
| Grant | `grants.md 2026-10-21 tui-v2 delivery on dev and PR 47 update` |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Decision 44: each board has one fixture file. Task `demo-mode` defined the
schema and wrote `Main.fixture.json` and `TooSmall.fixture.json`. This task
writes the fixture of each other board in the spec table.

## Design

- Each board of the spec table in
  `/home/penguin/source/hive/.hive/spec/tui-v2/design/`, except `Main` and
  `TooSmall`. Each board loads `./support.js` from the same folder.

This task adds data files and changes no screen. The visual check is V2: the
fixture review against each board.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/tui-v2/design/<Board>.fixture.json` | AC1. 19 new files. |
| `internal/tui/` tests | AC2. Only a test that loads each fixture. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/tui-v2/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files. Do not change the schema. When the schema cannot
hold a value of a board, stop and report it.

## Requirements

### AC1 The fixtures

- Write one fixture for each board of the spec table except `Main` and
  `TooSmall`: 19 files.
- Take each value from the board script: the data, the clock, the home, the
  size, and the UI state. Record the source line of each value group in the
  evidence.
- When a board shows an open task brief, an evidence file, a goal design, or
  a friction entry, put its text in the fixture `files` map of task
  `11-demo-mode`. Demo mode reads no live file.
- The `Components` board is a component sheet. Its fixture holds the sample
  data that the sheet shows. Record how a later task uses it.

### AC2 Load test

- A gate test loads each of the 21 fixtures with the demo loader. Each loads
  with no error, and its size matches the spec table.
- First failing test: the load test, before the 19 files exist.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Each board has a fixture with its data and UI state | V2 |
| AC2 | Each fixture loads and has the size of its board | V1 |

## Verification items

### V1 Tests

- Proves: AC2.
- Command: run the new test against the base code, then against the task code.
- Expected: it fails on the base and passes on the task branch.

### V2 Fixture review

- Proves: AC1.
- Command: for five boards that the verifier picks, compare each value of the
  fixture with the board script.
- Expected: each value matches.

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

The screens and their golden checks.

## Changes

- 2026-10-21: before the start, added the Design section and the rule for
  the `files` map of the fixture.
