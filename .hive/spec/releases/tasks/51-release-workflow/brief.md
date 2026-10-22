# Release package script and workflow

| Item | Value |
| --- | --- |
| Feature | `releases`, wave 2 |
| Tier | strong |
| Depends on | 49-releases-spec (landed). Base: `dev` |
| Grant | `grants.md 2026-10-18 hive delivery on dev` |
| Evidence | `.hive/spec/releases/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Decisions 1 to 15 and 22 of the spec: the release workflow
`.github/workflows/release.yml`, the package script
`.github/scripts/package.sh`, and `workflow_call` in `ci.yml`. The Fledge
versions of the script and the workflow are in commit `3dd44730^`. Use them
as a start.

The spec is `.hive/spec/releases/README.md`. Read it first. When the spec
and the source files do not agree, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.github/scripts/package.sh` | AC2. New. |
| `.github/scripts/check-tag.sh` | AC1. New. The tag checks of decisions 2 to 4. |
| `.github/workflows/release.yml` | AC3. New. |
| `.github/workflows/ci.yml` | AC3. Add `workflow_call` only. |
| `release_scripts_test.go` | AC1, AC2. New Go test at the root, so that the gate runs it. |
| `.gitignore` | Only when Git ignores a new file. Use the narrowest rule. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/releases/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files. Do not add a Go dependency.

## Requirements

### AC1 Tag checks

- `check-tag.sh TAG` exits 0 for a valid tag and prints the kind: `release`
  or `prerelease`. It exits 1 with a reason for each refusal of decisions 2
  to 4: a bad name, a leading zero, a lightweight tag, and a tag commit that
  is not on `origin/main`.
- `release_scripts_test.go` runs the script in a temporary Git repository
  with a bare remote, for each case. The test skips only when `bash` or
  `git` is absent.

### AC2 Package

- `package.sh TAG DIST` makes the three files of decision 7 with decisions 8
  to 10. It sets `internal/lib/version.Version` to the tag. On a host with a
  matching architecture, it checks that `hive version` prints `hive TAG`.
- `release_scripts_test.go` runs `package.sh` two times for one test tag
  into two folders. It checks the archive names, the archive members, the
  checksum lines, the version text of the host binary, and that the two runs
  give the same bytes. When this test takes more than 60 seconds, record the
  time and ask the manager.

### AC3 Workflow

- `release.yml` starts on the push of a tag that matches `v*`. It does the
  steps of decision 12 in order. It uses `check-tag.sh`, the gate of
  `ci.yml` through `workflow_call`, and `package.sh`.
- It checks the binary of each archive on a runner with the matching
  architecture (decision 14), for example `ubuntu-latest` and
  `ubuntu-24.04-arm`.
- It makes a draft release with the GitHub generated text and an install
  header that links to the install steps of the root `README.md` (decision
  22). It uploads each file. It publishes only after each upload (decision
  15). A `-rc.N` tag makes a prerelease.
- It has `permissions: contents: read` at the top. Only the publish job has
  `contents: write`. It has a `concurrency` group.
- `ci.yml` keeps its current triggers and steps.
- Install `actionlint` as a development tool with `go install`. Do not add
  it to `go.mod`.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The tag checks of decisions 2 to 4 | V1 |
| AC2 | Reproducible archives with the correct names, members, checksums, and version | V2 |
| AC3 | The workflow follows decisions 12 to 15 and 22, and `actionlint` passes | V3 |

## Verification items

### V1 Tag checks

- Proves: AC1.
- Command: `go test -run TestCheckTag -v .`. Then break one refusal in the
  script and run the test again.
- Expected: each case passes. The break fails its case.

### V2 Package

- Proves: AC2.
- Command: `go test -run TestPackage -v .`. Then run
  `.github/scripts/package.sh v0.1.0-rc.1 .hive/tmp/dist` and list the
  archives with `tar -tvzf`. Run `sha256sum --check checksums.txt`. Run the
  host binary with `hive version`.
- Expected: the tests pass. Each archive holds `hive` and `LICENSE` with
  owner 0 and the commit time. The checksums match. The binary prints
  `hive v0.1.0-rc.1`.

### V3 Workflow

- Proves: AC3.
- Command: `actionlint .github/workflows/*.yml`. Read `release.yml` next to
  decisions 12 to 15 and 22.
- Expected: `actionlint` prints nothing. Each step is present in order. The
  workflow pushes no tag and runs no step outside a tag push.

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
