# Move the golden frames spec into the spec tree

| Item | Value |
| --- | --- |
| Feature | `golden-frames`, spec task |
| Wave | 0 |
| Tier | small |
| Depends on | none. Base: `dev` |
| Blocks | each other task of `golden-frames` |
| Grant | goal `g12-golden-frames` |
| Evidence | `.hive/spec/golden-frames/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Commit `0effc449` added the behavior spec of the restyled TUI in
`redesign-spec/`. The user signed off all 40 boards on the page "Golden
frames" of the design canvas, version `1792795470-106f`. This task moves the
spec to `.hive/spec/golden-frames/`, the folder of the feature. It changes
no text of the spec. Later tasks copy the boards, export the golden grids,
and write the fixtures.

When a file does not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `redesign-spec/README.md` | AC1. Move it with `git mv` to `.hive/spec/golden-frames/README.md`. |
| `redesign-spec/screens/*.md` | AC1. Move the seven files with `git mv` to `.hive/spec/golden-frames/screens/`. |
| `ste_test.go` | AC2. Change the two `redesign-spec/` lines of `steFiles` to the new paths. |
| `testdata/ste-baseline.json` | AC2. Change the eight `redesign-spec/` keys to the new paths. Keep each value at 0. |
| `.hive/spec/golden-frames/tasks/<task-id>/brief.md` | The brief copy, if hive did not write it. Commit it. |

Do not edit other files. Do not change the text of a moved file.

## Requirements

### AC1 The move

- After the task, `redesign-spec/` does not exist in the worktree.
- `.hive/spec/golden-frames/README.md` and the seven files of
  `.hive/spec/golden-frames/screens/` have the same bytes as the files of
  `redesign-spec/` at commit `0effc449`.
- Git records each file as a rename. `git diff --stat -M dev...HEAD` shows
  each move at 100% similarity.
- `git check-ignore` does not ignore a moved file. The `.gitignore` allowlist
  already allows `.hive/spec/`. Do not edit `.gitignore`. When Git ignores a
  moved file, stop and report it.

### AC2 The ratchet

- `steFiles` in `ste_test.go` names `.hive/spec/golden-frames/README.md` and
  `.hive/spec/golden-frames/screens/*.md` in place of the two `redesign-spec/`
  lines. Keep the order and the format of the list.
- `testdata/ste-baseline.json` has the eight new keys with the value 0 and no
  `redesign-spec/` key. Keep the order and the format of the file.
- When the ratchet reports a violation in a moved file, stop and report it.
  Do not reword the file.

### AC3 No other reference

- `git grep -n redesign-spec` in the worktree finds no match outside the
  evidence and the brief copy.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The spec is in `.hive/spec/golden-frames/`, byte for byte, as renames | V1, V4 |
| AC2 | The ratchet names the new paths with 0 | V2, V5 |
| AC3 | No reference to `redesign-spec/` remains | V3 |

## Verification items

### V1 Move

- Proves: AC1.
- Command: for each of the eight files, `git show 0effc449:redesign-spec/PATH | cmp - .hive/spec/golden-frames/PATH`. Then `test ! -e redesign-spec` and `git check-ignore -v .hive/spec/golden-frames/README.md .hive/spec/golden-frames/screens/*.md`.
- Expected: no difference for each file. The folder is absent. `git check-ignore` prints nothing and exits 1.

### V2 Ratchet

- Proves: AC2.
- Command: `go test ./ -run 'STE|Gitignore' -count=1 -v`, and `grep -c golden-frames testdata/ste-baseline.json`.
- Expected: the tests pass. The count is 8.

### V3 References

- Proves: AC3.
- Command: `git grep -n redesign-spec -- ':!.hive/spec/golden-frames/tasks/'`.
- Expected: no output.

### V4 Diff scope

- Proves: AC1.
- Command: `git diff --stat -M dev...HEAD`, after you merge `dev` into the task branch.
- Expected: eight renames with no line change, the two named files, the brief copy, and the evidence. No other file.

### V5 Gate

- Proves: AC2.
- Command: the gate of `AGENTS.md` from the worktree root.
- Expected: `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0.

## Out of scope

The boards, the golden grids, the fixtures, and each code change. Wave 1 and
wave 2 tasks do them. `main`.

## Changes

None.
