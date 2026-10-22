# Spec of the releases feature

| Item | Value |
| --- | --- |
| Feature | `releases`, wave 1 |
| Tier | strong |
| Depends on | none. Base: `dev` |
| Blocks | each other task of `releases` |
| Grant | `grants.md 2026-10-18 hive delivery on dev` |
| Evidence | `.hive/spec/releases/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

On 2026-10-22 the user approved the decisions R1 to R9 of the release plan,
each with option A. The plan is in
`/home/penguin/source/hive/.hive/tmp/reports/releases-plan.md`, which Git
ignores. This task writes the spec of the feature from that plan, so that
each later task can read it.

When the plan and the source files do not agree, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/releases/README.md` | AC1. A new file. |
| `ste_test.go` | AC3. One line in `steFiles`. |
| `testdata/ste-baseline.json` | AC3. One entry for the README with 0, and one entry for the brief copy with 0. |
| `.hive/spec/releases/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files. Do not edit the plan.

## Requirements

### AC1 The spec

- Write `.hive/spec/releases/README.md` with these sections, in this order:
  Destination, Problem, Out of scope, Decisions, Tasks, Risks.
- Use the facts of the plan for Problem. Check each fact against the
  repository before you write it.
- Write one numbered decision for each approved option A of R1 to R9. Give
  each decision the details that a task needs:
  - R1: the tag pattern `vX.Y.Z`, the `-rc.N` prerelease form, the annotated
    tag, and the check that the tag commit is on `main`.
  - R2: the first version `v0.1.0`.
  - R3 and R4: the archive names `hive_vX.Y.Z_linux_amd64.tar.gz` and
    `hive_vX.Y.Z_linux_arm64.tar.gz` with the binary and `LICENSE`, and
    `checksums.txt`. The build uses `CGO_ENABLED=0`, `-trimpath`, and sets
    `internal/lib/version.Version` with `-ldflags -X`. The archives are
    reproducible.
  - R5: the install steps, and the change of `hive version` that reports the
    module version of a `go install` build.
  - R6: the update steps now, and a later `hive doctor` check that compares
    the binary version with the home tag.
  - R7: the new title and the top line of each Fledge release.
  - R8: the release text that GitHub generates, with an install header.
  - R9: the root `README.md`.
- Add these decisions about the gate and the publish:
  - The release workflow runs the gate of `ci.yml` before it builds.
    `ci.yml` accepts `workflow_call` for this.
  - The workflow makes a draft release, uploads each file, and then
    publishes. A failed step leaves no published release.
  - The workflow checks that the binary of each archive prints the tag with
    `hive version` on a matching runner.
  - No task pushes a tag or edits a GitHub release. The user approves the
    edit of the Fledge releases and the first tag, each one separately.
- Tasks: the waves of the plan. Wave 1 is this spec. Wave 2: the version
  change, the package script and the workflow, and the docs. Wave 3: the
  edit of the Fledge releases. Wave 4: the first release after pull
  request #47 merges. Wave 5: the doctor check of R6.
- Risks: at least the Go module proxy and the Fledge tags with another
  module path, a binary that runs with an older home, and a tag on a commit
  that is not on `main`.

### AC3 The ratchet

- Add `.hive/spec/releases/README.md` to `steFiles` in `ste_test.go`, with a
  baseline of 0 in `testdata/ste-baseline.json`.
- Write the README in ASD-STE100 Strict with the terms of `STE-GLOSSARY.md`.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The spec gives each approved decision and the waves | V1 |
| AC3 | The README is in the ratchet with 0 | V3, V5 |

## Verification items

### V1 Spec content

- Proves: AC1.
- Command: read the README next to the plan.
- Expected: each of R1 to R9 is present with option A and the details of
  AC1. Each fact in Problem agrees with the repository. No decision adds
  scope that the plan does not give.

### V3 Ratchet

- Proves: AC3.
- Command: `go test ./ -run STE`. Then add a known violation to the README,
  run the test again, and remove the violation.
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

- Do not run bare `hive` without `--no-startup` or `HIVE_NO_STARTUP=1`.
- Do not push a tag. Do not edit a GitHub release. Do not run `gh release`.

## Out of scope

Each other wave of the spec. `main`. Pull request #47.

## Changes

None.
