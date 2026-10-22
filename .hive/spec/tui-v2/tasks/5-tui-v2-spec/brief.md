# Spec of the TUI v2 feature

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, wave 1 |
| Tier | small |
| Depends on | none. Base: `dev` |
| Blocks | each other task of `tui-v2` |
| Grant | `grants.md 2026-10-21 tui-v2 delivery on dev and PR 47 update` |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The user approved the TUI v2 spec and its design boards on 2026-10-21. The
spec is in the personal folder `plans/tui-v2/`, which Git ignores. This task
puts the spec and the board sources in the repository, so that each later
task can read them.

When the source files do not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/tui-v2/README.md` | AC1. A new file. |
| `.hive/spec/tui-v2/design/` | AC2. New files. |
| `ste_test.go` | AC3. One line in `steFiles`. |
| `testdata/ste-baseline.json` | AC3. One entry for the README with 0, and one entry for the brief copy with 0. |
| `.hive/spec/tui-v2/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files. Do not edit `plans/tui-v2/`.

## Requirements

### AC1 The spec

- Copy `/home/penguin/source/hive/plans/tui-v2/README.md` to
  `.hive/spec/tui-v2/README.md` in the worktree.
- Keep the text the same, except as AC3 permits.

### AC2 The board sources

- Copy each file of `/home/penguin/source/hive/plans/tui-v2/design/` to
  `.hive/spec/tui-v2/design/` in the worktree. There are 21 `.dc.html` files
  and `canvas.json`.
- Do not change a board file. `cmp` gives no difference for each file.

### AC3 The ratchet

- Add `.hive/spec/tui-v2/README.md` to `steFiles` in `ste_test.go`, with a
  baseline of 0 in `testdata/ste-baseline.json`.
- When `ste_test.go` finds a violation in the README, change the words of
  that sentence. Keep its meaning. Do not change a decision number, a key, a
  symbol, a hex value, a size, a path, a version, or a quoted text. Record
  each change, with the old and the new sentence, in the evidence.
- When a change would alter the meaning of a decision, stop and report it.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The spec is in `.hive/spec/tui-v2/README.md` | V1 |
| AC2 | Each board source is in `.hive/spec/tui-v2/design/`, with no change | V2 |
| AC3 | The README is in the ratchet with 0 | V3, V5 |

## Verification items

### V1 Spec copy

- Proves: AC1.
- Command: `diff plans/tui-v2/README.md` from the home against the copy.
- Expected: no difference, or only the sentences that the evidence records
  for AC3.

### V2 Board copies

- Proves: AC2.
- Command: `cmp` for each file of the source folder against its copy, and a
  count of the files.
- Expected: 22 files. No difference.

### V3 Ratchet

- Proves: AC3.
- Command: `go test ./ -run STE`. Then add a known violation to the README
  copy, run the test again, and remove the violation.
- Expected: the test passes, fails with the violation, and passes again.

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
