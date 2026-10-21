# Trust dialogs of each harness

| Item | Value |
| --- | --- |
| Feature | `general-agent`, wave 2 (Trust) |
| Tier | decent |
| Depends on | 14-general-agent-spec (landed). Base: `dev` after they land |
| Grant | `grants.md 2026-10-18 hive delivery on dev` |
| Evidence | `.hive/spec/general-agent/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Decision 13: for each harness (claude, codex, cursor, opencode, pi), find a
method that starts the harness in a new directory with no trust dialog, or
record the result "no method". The spawn of a general agent (wave 5) uses
this result.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/lib/harness/` and its tests | AC1, AC2. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/general-agent/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### AC1 One result for each harness

- For each harness, test the installed version in a new empty directory
  under `.hive/tmp/` of the worktree. Record the version, the command, and
  what the pane shows.
- When a method exists (an argument, a setting, or a preflight), add it to
  the harness profile in Go. Record the tested version beside it.
- When no method exists, the profile states "no method" in a field that wave
  5 can read. Choose the field name and record it.
- A harness that is not installed gets "no method" with the reason "not
  installed". Do not install a harness.
- A method must not trust a directory for all time outside the spawn, and must
  not change a global config of the user. When the only method does that,
  record it as "no method" and describe the method in the evidence.

### AC2 Tests

- First failing test: the profile of each harness gives its trust result
  (method or "no method").
- For each harness with a method, a test checks the start arguments or the
  preflight. The live check below proves the method.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Each harness has a tested method or "no method", with the version | V1 |
| AC2 | The profiles carry the result and tests prove it | V-red |

## Verification items

### V1 Live check

- Proves: AC1.
- Command: for each harness with a method, start it with the profile command
  in a new empty directory in a new herdr pane. Read the pane.
- Expected: no trust dialog shows. Close each pane. Remove the directories.

## Rules

- Read `.hive/spec/general-agent/README.md` first, then the decisions that
  this brief names. The spec is the authority. This brief does not repeat
  each rule. When the code or the spec does not agree with this brief, stop
  and report it with `hive inbox send --to manager-hive TEXT --agent NAME`.
- Each Go change is test-first. The brief names the first failing test. See
  it fail for the correct reason. Do not mock the behavior under test. Use
  the fake herdr server of the package tests.
- Each new command prints text by default and one JSON object with `--json`,
  and logs its request, its result, and each error (decision 42).
- Run bare `hive` only with `--no-startup` or `HIVE_NO_STARTUP=1`. A live
  check uses a temp `HIVE_HOME` and never changes the real home, its claims,
  or its coordinator. Close each workspace, tab, and pane that the check makes.

## Verification items (common)

### V-red Tests red and green

- Proves: each criterion.
- Command: run each new test against the base code, then against the task
  code. Then break the behavior on purpose once and run the tests again.
- Expected: each new test fails on the base for the correct reason and passes
  on the task branch. The broken behavior fails a test.

### V-gate Diff scope and gate

- Proves: each criterion.
- Command: `git diff dev...HEAD --stat`, after you merge `dev` into the task
  branch. Then the full gate of `AGENTS.md` from the worktree root.
- Expected: only the files of the source table and the evidence changed.
  `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0.
  Put the full gate output in the evidence.

## Changes

None.
