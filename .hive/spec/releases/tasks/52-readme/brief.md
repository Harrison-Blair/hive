# Root README with install and update

| Item | Value |
| --- | --- |
| Feature | `releases`, wave 2 |
| Tier | decent |
| Depends on | 49-releases-spec (landed). Base: `dev` |
| Grant | `grants.md 2026-10-18 hive delivery on dev` |
| Evidence | `.hive/spec/releases/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Decisions 16, 18, and 23 of the spec: the repository gets a root
`README.md`. It tells what hive is, how to install hive, and how to update
hive.

The spec is `.hive/spec/releases/README.md`. Read it first. When the spec
and the source files do not agree, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `README.md` | AC1. New. |
| `.gitignore` | Only when Git ignores `README.md`. Use the narrowest rule. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/releases/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### AC1 The README

- Sections: what hive is (from the Destination of `.hive/spec/hive-v1/README.md`
  and the Project section of `AGENTS.md`), Requirements (herdr 0.9.5, Go
  from `go.mod` for a source build, Linux amd64 or arm64), Install (the four
  steps of decision 16), Update (the four steps of decision 18), and Links
  (the hive v1 spec, `AGENTS.md`, `reference/`).
- The install section has an anchor that the release header can link to:
  the heading `## Install`.
- Give each command in a code block. Use `vX.Y.Z` as the version. Show the
  checksum check of the archive with `sha256sum --check --ignore-missing
  checksums.txt`.
- The README is for people. It does not need ASD-STE100 and does not go in
  `steFiles`.
- Check each command that does not need a release: the clone, `go install`
  from a local checkout path, and `hive init` in a scratch home under
  `.hive/tmp/` with a scratch config. Do not change the real home or
  `~/.config/hive/config.toml`.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The README gives the sections, and each local command works | V1 |

## Verification items

### V1 README

- Proves: AC1.
- Command: read the README next to decisions 16, 18, and 23. Run each local
  command of AC1 in `.hive/tmp/` with a scratch home and config.
- Expected: each section is present. Each command works as the README says.
  The real home and config do not change.

### V-gate Gate

- Command: merge `dev` once into the task branch, then run the gate of
  `AGENTS.md` from the worktree root.
- Expected: `gofmt -l` and `go mod tidy -diff` print nothing. The other
  commands exit 0.

## Rules

- Do not run bare `hive` without `--no-startup` or `HIVE_NO_STARTUP=1`.
- Do not push a tag. Do not edit or make a GitHub release. Do not run
  `gh release`. Do not start a workflow on GitHub.
- Put scratch files in `.hive/tmp/` of the worktree. Remove them at the end.

## Out of scope

The other tasks of wave 2. Waves 3 to 5. `main`. Pull request #47.

## Changes

None.
