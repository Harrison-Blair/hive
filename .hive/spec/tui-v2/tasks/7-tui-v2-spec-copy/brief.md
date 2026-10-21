# Follow-up copy of the TUI v2 spec and boards

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, wave 1 follow-up |
| Tier | small |
| Depends on | 5-tui-v2-spec (landed). Base: `dev` |
| Blocks | each other task of `tui-v2` |
| Grant | `grants.md 2026-10-21 tui-v2 delivery on dev and PR 47 update` |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Task `5-tui-v2-spec` put the spec and the boards in `.hive/spec/tui-v2/`. A
manager audit then found conflicts between the boards and the spec. On
2026-10-21 the user accepted each fix. The coordinator changed the spec and
20 boards in `plans/tui-v2/`. The design artifact is now version 24. This
task copies the changed files into the repository again.

When the source files do not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/tui-v2/README.md` | AC1. Replace it. |
| `.hive/spec/tui-v2/design/` | AC2. Replace the changed files. |
| `ste_test.go` | Only to add the brief copy, if the test needs it. |
| `testdata/ste-baseline.json` | One entry for the brief copy with 0. |
| `.hive/spec/tui-v2/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files. Do not edit `plans/tui-v2/`.

## Requirements

### AC1 The spec

- Copy `/home/penguin/source/hive/plans/tui-v2/README.md` over
  `.hive/spec/tui-v2/README.md` in the worktree.
- Do not change the text. `cmp` gives no difference.

### AC2 The board sources

- Copy each file of `/home/penguin/source/hive/plans/tui-v2/design/` over
  its copy in `.hive/spec/tui-v2/design/`. There are 21 `.dc.html` files and
  `canvas.json`. About 20 files change.
- Do not change a board file. `cmp` gives no difference for each file.

### AC3 The ratchet

- The README stays in the ratchet with a baseline of 0. The coordinator
  reports 0 violations. Do not reword the README.
- When `ste_test.go` finds a violation in the README, stop and report it.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The spec is in `.hive/spec/tui-v2/README.md` | V1 |
| AC2 | Each board source is in `.hive/spec/tui-v2/design/`, with no change | V2 |
| AC3 | The README is in the ratchet with 0 | V3, V5 |

## Verification items

### V1 Spec copy

- Proves: AC1.
- Command: `cmp` of `plans/tui-v2/README.md` in the home against the copy.
- Expected: no difference.

### V2 Board copies

- Proves: AC2.
- Command: `cmp` for each file of the source folder against its copy, and a
  count of the files.
- Expected: 22 files. No difference. List the changed files in the evidence.

### V3 Ratchet

- Proves: AC3.
- Command: `go test ./ -run STE`.
- Expected: the test passes.

### V4 Diff scope

- Proves: AC1, AC2, AC3.
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
