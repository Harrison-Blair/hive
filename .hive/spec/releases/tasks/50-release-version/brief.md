# Version from the Go build information

| Item | Value |
| --- | --- |
| Feature | `releases`, wave 2 |
| Tier | decent |
| Depends on | 49-releases-spec (landed). Base: `dev` |
| Grant | `grants.md 2026-10-18 hive delivery on dev` |
| Evidence | `.hive/spec/releases/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Decision 17 of the spec: `hive version` reports the module version of the Go
build information when `-ldflags -X` does not set `Version`. Thus a
`go install github.com/Harrison-Blair/hive@vX.Y.Z` build reports `vX.Y.Z`.

The spec is `.hive/spec/releases/README.md`. Read it first. When the spec
and the source files do not agree, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/lib/version/version.go` | AC1. |
| `internal/lib/version/version_test.go` | AC1. New tests. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/releases/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### AC1 The version rule

- When `-ldflags -X` sets `Version`, `hive version` reports that value. This
  rule does not change.
- When `Version` is `dev`, use `Main.Version` of the build information when
  it is a release version: `vX.Y.Z` or `vX.Y.Z-rc.N`, with no other suffix.
- Go 1.24 and later stamps a pseudo-version, for example
  `v0.0.3-0.20261022041405-6f14e3bd53b9`, on a build in a checkout. It can
  also add `+dirty`. Manager decision: a pseudo-version, a `+dirty` version,
  `(devel)`, and an empty value keep `dev`. The commit and the modified
  fields already identify such a build.
- Test-first. Give the build information to the function under test as
  data. Do not compute an expected value with the code under test.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The version rule for each case | V1, V2 |

## Verification items

### V1 Unit tests

- Proves: AC1.
- Command: `go test ./internal/lib/version/ -run . -v` on the base and on the
  task branch.
- Expected: the new tests fail on the base and pass on the task branch. Each
  case has a test: ldflags value, `vX.Y.Z`, `vX.Y.Z-rc.N`, pseudo-version,
  `+dirty`, `(devel)`, and empty.

### V2 Real builds

- Proves: AC1.
- Command: build three binaries into `.hive/tmp/`: `go build .` in the
  worktree, `go build -ldflags "-X github.com/Harrison-Blair/hive/internal/lib/version.Version=v9.9.9" .`,
  and a `go install` of a local module version. For the last case, make a
  temporary module copy with a `v9.9.8` tag in `.hive/tmp/`, and use a
  `GOPROXY` folder or `GOFLAGS=-mod=mod` with a `replace`. When no practical
  method gives a real module version, record that and rely on V1 for it.
  Run `hive version` and `hive --json version` with each binary.
- Expected: `dev`, `v9.9.9`, and `v9.9.8`.

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
