# Doctor check of the binary version against the home tag

| Item | Value |
| --- | --- |
| Feature | `releases`, wave 5 |
| Tier | decent |
| Depends on | 50-release-version (landed), first release `v0.1.0` (published). Base: `dev` |
| Grant | `grants.md 2026-10-18 hive delivery on dev` |
| Evidence | `.hive/spec/releases/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Decision 19 of `.hive/spec/releases/README.md`: `hive doctor` compares the
version of the binary with the tag of the home checkout. It shows a warning
when the two differ. Decision 18 updates the home and the binary to the same
tag, and this check finds a step that a user did not do.

When the spec and the source files do not agree, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/doctor/checks.go` | AC1. One new check function. |
| `internal/doctor/run.go` | AC1. Add the check to the list. |
| `internal/doctor/` tests | AC1. New tests. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/releases/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### AC1 The check `home version`

- The check is not required. It uses `newCheck` with `required` false, so a
  difference gives `StatusWarn` and the exit status stays 0.
- Read the binary version with `internal/lib/version.Info()`.
- When the binary version is `dev`, the check passes. The detail says that
  a development build has no release version to compare.
- When the binary version is a release version, read the tags that point at
  `HEAD` of the home with Git (`git -C HOME tag --points-at HEAD`).
  - The check passes when that list has the binary version. The detail
    names the version.
  - Otherwise it warns. The detail gives the binary version, the home tags
    at `HEAD` or the short commit when no tag points at it, and the update
    steps of the root `README.md` (`## Update`).
- When Git fails, the check warns with the error.
- Give the binary version and the Git command to the check as parameters,
  so that a test can give each case without a real build. Use a real
  temporary Git repository with tags for the Git part. Do not mock the
  comparison under test.
- Test-first.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The check gives pass or warn for each case, and `hive doctor` shows it | V1, V2 |

## Verification items

### V1 Unit tests

- Proves: AC1.
- Command: `go test ./internal/doctor/ -run HomeVersion -v` on the base and
  on the task branch. Then break the comparison and run the tests again.
- Expected: the new tests fail on the base and pass on the task branch. The
  break fails a test. The tests have these cases:
  - `dev`.
  - A release version with the same tag at `HEAD`.
  - A different tag at `HEAD`.
  - No tag at `HEAD`.
  - Two tags at `HEAD` with one match.
  - A Git failure.

### V2 Real doctor run

- Proves: AC1.
- Command: in `.hive/tmp/`, clone the repository at `v0.1.0` as a scratch
  home and make a scratch config for it. Build a binary with
  `-ldflags "-X github.com/Harrison-Blair/hive/internal/lib/version.Version=v0.1.0"`.
  Run `hive doctor` and `hive --json doctor` for the scratch home. Then check
  out a commit with no tag in the scratch home and run them again. A clone
  inside the real home resolves to the real home, so put the clone outside
  the home, for example in your scratchpad, and remove it at the end.
- Expected: the first run shows `home version` as pass. The second run
  shows it as a warning with the update steps, and the exit status is 0 when
  no required check fails. The real home and `~/.config/hive/config.toml` do
  not change.

### V-gate Gate

- Command: merge `dev` once into the task branch, then run the gate of
  `AGENTS.md` from the worktree root.
- Expected: `gofmt -l` and `go mod tidy -diff` print nothing. The other
  commands exit 0.

## Rules

- Do not run bare `hive` without `--no-startup` or `HIVE_NO_STARTUP=1`.
- Do not push a tag. Do not edit or make a GitHub release.
- Put scratch files in `.hive/tmp/` of the worktree or in your scratchpad.
  Remove them at the end.

## Out of scope

A `hive update` command. A change to the release workflow. `main`.

## Changes

None.
