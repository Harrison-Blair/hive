# Make the gate test line match CI

| Item | Value |
| --- | --- |
| Feature | `golden-frames`, goal `g12-golden-frames`, wave B repair |
| Tier | small |
| Depends on | `141-gf-ci-tz-term`, landed on `dev` at `cca96aa1`. Base: `dev`. |
| Grant | Goal `g12-golden-frames`. The user approved this change in ask `a25` on 2026-10-24. |
| Evidence | The `evidence/` folder beside this brief. Put `VERIFIED.md` there. |

## Overview

CI runs the tests in UTC with no `TERM`. The local gate runs them in the zone
and the terminal of the computer. Task 137 passed the local gate and then failed
CI. This task changes one line of the gate in `AGENTS.md` so that the local
gate runs the tests as CI does.

When a file does not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `AGENTS.md` | AC1. Change one line of the gate block. |
| `testdata/ste-baseline.json` | One entry of 0 for this brief copy. |

Edit no other file. Do not edit a skill. Goal g13 replaces the skills.

## Requirements

### AC1 The gate line

- In the gate block of `AGENTS.md`, the line `go test -race ./...` becomes
  `env -u TERM TZ=UTC go test -race ./...`.
- No other line of `AGENTS.md` changes.
- `AGENTS.md` stays at its STE baseline or below.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The gate line of `AGENTS.md` matches CI | V1, V2, V3 |

## Verification items

### V1 Diff

- Proves: AC1.
- Command: `git diff dev...HEAD -- AGENTS.md`, after you merge `dev` into the task branch.
- Expected: one removed line and one added line, as AC1 gives.

### V2 The new line runs

- Proves: AC1.
- Command: in a new clone of the task branch, run the gate block of `AGENTS.md` as written.
- Expected: each command exits 0. The first two commands print nothing.

### V3 Gate and ratchet

- Proves: AC1.
- Command: `go test ./ -run STE -count=1`, and the gate of `AGENTS.md` from the worktree root.
- Expected: both pass.

## Out of scope

The skills, the CI workflow, and the profiles of g13. `main`.

## Changes

None.
