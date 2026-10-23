# TUI header clock tests must not depend on the local zone

| Item | Value |
| --- | --- |
| Feature | `tui-v2` |
| Tier | small |
| Depends on | none. Base: `dev` at `7d8f42db` or later |
| Grant | goal `g8-pr-dev-main`: the user chose "Open now" in ask a2 (2026-10-23). The goal needs CI green on `dev`. Land on `dev`. |
| Evidence | `.hive/spec/tui-v2/tasks/132-tui-clock-tests/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The GitHub Actions gate fails on `dev`. CI run 39903251785 at commit
`7d8f42db` fails in the step "Test with the race detector". Two tests fail:

- `TestHeaderShowsTheOpenAskRecords` in `internal/tui/ask_test.go`.
- `TestHeaderShowsTheAskCountOfTheState` in `internal/tui/chat_modes_test.go`.

Both tests expect the clock text `08:00`. The fixed clock instant is
`2026-10-20T12:00:00Z`. The test clock shows the instant in `time.Local`.
The text is 08:00 in the US Eastern zone, where the author wrote the tests. The
CI runner uses UTC, where the same instant shows as `12:00`.

The product shows the header clock in the local zone. That behavior is
correct. This task changes the tests only.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/ask_test.go` | AC1, AC2. `TestHeaderShowsTheOpenAskRecords`. |
| `internal/tui/chat_modes_test.go` | AC1, AC2. `TestHeaderShowsTheAskCountOfTheState`. |
| `internal/tui/tui_test.go` | Only when the chosen fix needs a shared test helper. Record the cause in the evidence. |
| `.hive/spec/tui-v2/tasks/132-tui-clock-tests/brief.md` | A copy of this brief. |
| `.hive/spec/tui-v2/tasks/132-tui-clock-tests/evidence/` | The evidence files. |

Do not edit other files. Do not change product code. Do not change the
header text, the header layout, or the ask count. Do not weaken a test.

For V3 only, you can edit `internal/tui/component/header.go` for a temporary
break. Restore it with `git checkout --` before the next step.

## Requirements

### AC1 Zone independence

- The two named tests pass with `TZ=UTC`.
- The two named tests pass with `TZ=America/New_York`.
- The two named tests pass in the default zone of the worktree.
- The expected clock text does not depend on the zone of the machine that runs the test.

### AC2 Test strength

- Each test still fails when the ask count or the ask word does not agree with the header.
- Each test still fails when the header layout around the count changes.
- A temporary break of the product proves each failure. Restore the product after each break. Do not commit the break.
- Do not replace the suffix check with a check that passes without the clock text.

### AC3 No product change

- `git diff dev --stat` lists only the test files, the brief copy, and the evidence files.
- The header output for a user does not change.

### AC4 Gate

- The gate of `AGENTS.md` passes in the worktree.
- `TZ=UTC go test -race ./...` passes in the worktree. This run has the same zone as the CI runner.

## Verification items

### V1 Reproduce the failure

- Proves: AC1.
- Command: `TZ=UTC go test -race ./internal/tui -run 'TestHeaderShowsTheOpenAskRecords|TestHeaderShowsTheAskCountOfTheState' -count=1` at the base commit, before the test edit.
- Expected: both tests fail with the clock text `12:00` and the expected text `08:00`.

### V2 Zone runs

- Proves: AC1.
- Command: the same test command with `TZ=UTC`, `TZ=America/New_York`, and `TZ=Asia/Tokyo` after the edit.
- Expected: both tests pass in each run.

### V3 Break runs

- Proves: AC2.
- Command: for each break, make a temporary edit to `internal/tui/component/header.go`. The edit changes the ask count, the ask word, or the header layout. Then run the test command in the default zone. Then run `git checkout -- internal/tui/component/header.go`.
- Expected: each test fails at the temporary break. The worktree is clean after the restore.

### V4 Diff

- Proves: AC3.
- Command: `git diff dev --stat` and `git status --short`.
- Expected: only the test files, the brief copy, and the evidence files. No product file. No uncommitted change.

### V5 UTC full suite

- Proves: AC4.
- Command: `TZ=UTC go test -race ./...`.
- Expected: all packages pass.

### V6 Gate

- Proves: AC4.
- Command: the five gate commands of `AGENTS.md` from the worktree root.
- Expected: `gofmt -l` and `go mod tidy -diff` print nothing. `go vet`, `golangci-lint`, and `go test -race` exit 0.

## Finish

Commit on the task branch with a subject such as
`fix(test): make the header clock tests zone-independent`. Push the task
branch. Write the evidence and run `hive task complete`. Send the commit SHA
and the V1-V6 results to the manager with `hive inbox send --to manager-hive`.
