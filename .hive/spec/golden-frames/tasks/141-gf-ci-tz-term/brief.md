# Make the task 137 tests pass in CI

| Item | Value |
| --- | --- |
| Feature | `golden-frames`, goal `g12-golden-frames`, wave B repair |
| Tier | strong |
| Depends on | `137-gf-frame-chat`, landed on `dev` at `e3f107dc`. Base: `dev`. |
| Grant | Goal `g12-golden-frames`. This task repairs the landing of task 137. |
| Evidence | The `evidence/` folder beside this brief. Put `VERIFIED.md` there. |

## Overview

CI fails on `dev` since the merge of task 137 (run `40052654140`). The tests
pass on the computer of the user. They fail in CI for two causes:

1. Six tests of `internal/tui` depend on the local time zone. With `TZ=UTC`
   they fail: `TestAskReviewOneQuestion`, `TestAskCardSitsInTheComposerFrame`,
   `TestAskJAndKMoveBetweenCards`, `TestAskHistoryShowsTheAnswerAndTheDismissal`,
   `TestChatCompactStatusRowKeepsTheClock`, and `TestChatCompactAskStatusRowCutsKeys`.
2. `cmd` `TestBareHiveRunsTUIUntilQuit` fails with `TERM=dumb` and in a new
   clone. The output has no hide-cursor sequence. The test also expects the
   old header text ` hive`.

Task `132-tui-clock-tests` fixed the same class of fault for the old header
clock. Read its brief for the method.

When a cause is not one of these two, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| The six named tests in `internal/tui/` | AC1. Fix the time zone of each test. |
| `internal/tui/` non-test files | AC1. Only when a test shows that the code uses the local zone where the spec gives the zone of the fixture or of the TUI clock. |
| `cmd/root_test.go` | AC2. Check the TUI of the golden-frames frame in a form that does not depend on `TERM`. |

Edit no other file. Do not weaken a test. Each test still checks the text
that it checks today, at a fixed time zone.

## Requirements

### AC1 No local time zone

- Each of the six tests passes with `TZ=UTC`, `TZ=America/New_York`, and
  `TZ=Asia/Kolkata`.
- When the code is correct and the test uses `time.Local`, fix the test.
  When the code uses the local zone against the spec, fix the code.

### AC2 The bare hive test

- `TestBareHiveRunsTUIUntilQuit` passes with `TERM=dumb`, with
  `TERM=xterm-256color`, and with no `TERM`.
- It still proves that the TUI ran and quit with status 0, and that it wrote
  no JSON envelope. It checks a text of the new frame, not the old header.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The six tests pass in each time zone | V1, V3 |
| AC2 | The bare hive test passes with each `TERM` | V2, V3 |

## Verification items

### V1 Time zones, red and green

- Proves: AC1.
- Command: for each zone of AC1, `TZ=ZONE go test ./internal/tui/ -count=1 -run 'NAMES'` on `e3f107dc` and on the task branch.
- Expected: the base fails in at least one zone. The task branch passes in each zone.

### V2 TERM, red and green

- Proves: AC2.
- Command: for each `TERM` of AC2, `go test ./cmd/ -count=1 -run TestBareHiveRunsTUIUntilQuit` on `e3f107dc` and on the task branch. Then break the quit on purpose and show that the test fails.
- Expected: the base fails with `TERM=dumb`. The task branch passes for each value. The break fails.

### V3 CI

- Proves: AC1, AC2.
- Command: in a new clone of the task branch, `TZ=UTC TERM=dumb go test -race ./...`. Do not open a pull request. No grant covers it.
- Expected: the run passes. The manager checks the CI run on `dev` after the landing.

### V4 Gate

- Proves: AC1, AC2.
- Command: the gate of `AGENTS.md` from the worktree root.
- Expected: `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0.

## Out of scope

Each other test. Task 140. `main`.

## Changes

None.
