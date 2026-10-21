# Spec of the general agent feature

| Item | Value |
| --- | --- |
| Feature | `general-agent`, wave 1 |
| Tier | small |
| Depends on | none. Base: `dev` |
| Blocks | each other task of `general-agent` |
| Grant | `grants.md 2026-10-18 hive delivery on dev` |
| Evidence | `.hive/spec/general-agent/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The user approved the general agent spec on 2026-10-21. The spec is in the
personal folder `plans/general-agent/`, which Git ignores. This task puts
the spec in the repository, so that each later task can read it.

When the source files do not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/general-agent/README.md` | AC1. A new file. |
| `ste_test.go` | AC3. One line in `steFiles`. |
| `testdata/ste-baseline.json` | AC3. One entry for the README with 0, and one entry for the brief copy with 0. |
| `.hive/spec/general-agent/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files. Do not edit `plans/general-agent/`.

## Requirements

### AC1 The spec

- Copy `/home/penguin/source/hive/plans/general-agent/README.md` to
  `.hive/spec/general-agent/README.md` in the worktree.
- Keep the text the same, except as AC3 permits.

### AC3 The ratchet

- Add `.hive/spec/general-agent/README.md` to `steFiles` in `ste_test.go`, with a
  baseline of 0 in `testdata/ste-baseline.json`.
- When `ste_test.go` finds a violation in the README, change the words of
  that sentence. Keep its meaning. Do not change a decision number, a key, a
  symbol, a hex value, a size, a path, a version, or a quoted text. Record
  each change, with the old and the new sentence, in the evidence.
- When a change would alter the meaning of a decision, stop and report it.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The spec is in `.hive/spec/general-agent/README.md` | V1 |
| AC3 | The README is in the ratchet with 0 | V3, V5 |

## Verification items

### V1 Spec copy

- Proves: AC1.
- Command: `diff plans/general-agent/README.md` from the home against the copy.
- Expected: no difference, or only the sentences that the evidence records
  for AC3.

### V3 Ratchet

- Proves: AC3.
- Command: `go test ./ -run STE`. Then add a known violation to the README
  copy, run the test again, and remove the violation.
- Expected: the test passes, fails with the violation, and passes again.

### V4 Diff scope

- Proves: AC1, AC3.
- Command: `git diff dev...HEAD --stat`, after you merge `dev` into the task
  branch.
- Expected: only the files of the source table and the evidence changed.

### V5 Gate

- Proves: AC3.
- Command: the gate of `AGENTS.md` from the worktree root.
- Expected: `gofmt -l` and `go mod tidy -diff` print nothing. The other
  commands exit 0.

## Rules

- Do not run bare `hive` in herdr. The `--no-startup` switch does not exist
  until wave 2. Refer to the friction entry `scratch-hive-starts-real-agents`.

## Out of scope

Each other wave of the spec. `main`. Pull request #47.

## Changes

None.
