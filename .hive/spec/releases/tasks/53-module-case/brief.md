# Module zip: remove the letter-case path clashes

| Item | Value |
| --- | --- |
| Feature | `releases`, wave 2 |
| Tier | decent |
| Depends on | 49-releases-spec (landed). Base: `dev` |
| Grant | `grants.md 2026-10-18 hive delivery on dev` |
| Evidence | `.hive/spec/releases/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Decision 16 of `.hive/spec/releases/README.md` installs hive with
`go install github.com/Harrison-Blair/hive@vX.Y.Z`. The worker of task
`50-release-version` found that Go refuses the module zip of this
repository. Eight tracked files have paths that differ from other tracked
paths only in letter case:

- `.hive/spec/tui-v2/tasks/34-screen-chat/evidence/screens/chat-G.ansi` and `chat-G.png`
- `.hive/spec/tui-v2/tasks/35-screen-work-goals/evidence/verifier3/G.ansi`, `G.json`, and `G.txt`
- `.hive/spec/tui-v2/tasks/35-screen-work-goals/evidence/verifier4/G.ansi`, `G.json`, and `G.txt`

Each one clashes with the file of the same name in lower case. This task
renames the eight files and adds a test that stops a new clash.

When the source files do not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| The eight files above | AC1. Rename with `git mv`. |
| Markdown files in the two evidence folders that name a renamed file | AC1. Change the name only. |
| `module_paths_test.go` | AC2. New Go test at the root. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/releases/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files. Do not change the contents of a renamed file.

## Requirements

### AC1 Rename

- Give each upper-case `G` file a name that tells its key: `chat-shift-g.*`
  and `shift-g.*`. Read the evidence first to make sure that `G` is the
  shift-g key. When it is not, stop and report.
- Change the file name in each Markdown file that names it.
- Do not change a `.txt` or `.json` file that records command output, for
  example a diff stat. That text is a record of a past run.

### AC2 Clash test

- `module_paths_test.go` reads the tracked and new files with
  `git ls-files -co --exclude-standard`, and fails when two paths are equal
  after `strings.ToLower`. The message names both paths.
- The test does not read a checkout below `projects/`.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | No tracked path clashes in letter case | V1, V2 |
| AC2 | The test fails on a clash | V1 |

## Verification items

### V1 Clash test

- Proves: AC1, AC2.
- Command: `go test -run TestNoLetterCaseClash -v .` on the base and on the
  task branch.
- Expected: it fails on the base with the eight clashes, and passes on the
  task branch.

### V2 Module install

- Proves: AC1.
- Command: in `.hive/tmp/`, make a file proxy for a commit of the task
  branch at version `v0.9.7`, for example with a bare clone, a local tag, and
  `GOPROXY=file://...` or `GOFLAGS=-modcacherw` with `go mod download`. Then
  run `GOBIN=.hive/tmp/bin go install github.com/Harrison-Blair/hive@v0.9.7`.
  Push no tag to `origin`.
- Expected: the install succeeds. Record each command and its output. When
  Go reports another module zip error, stop and report it.

### V-gate Gate

- Command: merge `dev` once into the task branch, then run the gate of
  `AGENTS.md` from the worktree root.
- Expected: `gofmt -l` and `go mod tidy -diff` print nothing. The other
  commands exit 0.

## Rules

- Do not run bare `hive` without `--no-startup` or `HIVE_NO_STARTUP=1`.
- Do not push a tag. Do not edit or make a GitHub release. Do not run
  `gh release`.
- Put scratch files in `.hive/tmp/` of the worktree. Remove them at the end.

## Out of scope

The other tasks of wave 2. `main`. Pull request #47.

## Changes

None.
