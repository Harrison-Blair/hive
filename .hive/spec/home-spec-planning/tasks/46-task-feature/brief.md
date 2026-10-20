# Task feature field

| Item | Value |
| --- | --- |
| Wave | H2, task A of 3 |
| Tier | strong |
| Depends on | none. Base: `dev` at `707e332f` or later |
| Grant | `grants.md 2026-10-19 H2 spec layout wave` |
| Evidence | `evidence/<task-id>/` at the root of the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Each task belongs to a feature. A feature is `.hive/spec/<feature>/` in the
project repository. This task adds the `feature` field to the task record, the
derived paths, the `--feature` flags, and the allowlist rule. Task B copies
the brief into the worktree. Task C moves the old files. The plan is
`projects/hive/plans/home-spec-planning/MAP.md` in the home, decisions 4 to
9a and 11. The user agreed to this plan. Do not change a decision in this brief. When
the code does not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `internal/lib/records/task.go` has the task record. `internal/task/` has
  create, edit, list, and show. `cmd/task/` has the Cobra wiring.
- `slugPattern` in `internal/task/create.go` is the kebab-case pattern.
- The project record names the registered repository path and the base branch.
- `.gitignore` is an allowlist. `.hive/*` ignores everything in `.hive/`
  except `.hive/README.md`. `gitignore_test.go` checks the rules.
- Some records in the home have no feature. They must stay valid.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/lib/records/task.go` and its tests | R1. |
| `internal/task/` (create, edit, list, show, and a new file for the paths) and their tests | R1-R5. |
| `cmd/task/create/`, `cmd/task/edit/`, `cmd/task/list/`, `cmd/task/show/` and their tests | R2-R5. Flags only. |
| `.gitignore`, `gitignore_test.go`, `AGENTS.md` (section `Home layout`), `.hive/README.md` | R6. |
| `.agents/skills/hive-manager/SKILL.md`, `internal/agent/templates/project-agents.md`, and the tests that assert their text | R7. |
| Each existing test that runs `task create` without `--feature`, for example `cmd/task_lifecycle_test.go` | R2. Only to add `--feature` (and `--spec` or a README commit where the test needs it). Change no other assertion. List each file in the evidence. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `spec/hive-v1/tasks/<task-id>.md` | R8. New. A copy of this brief. |

Do not edit other files.

## Requirements

### R1 Record and paths

- The task record has an optional `feature` string (`json:"feature,omitempty"`).
  A record without it stays valid in every command.
- A feature is a kebab-case slug: it matches `slugPattern`, and it is 1 to 40
  characters.
- One function gives the paths of a task with a feature, relative to the project repository
  root: `.hive/spec/<feature>/README.md`,
  `.hive/spec/<feature>/tasks/<id>/brief.md`, and
  `.hive/spec/<feature>/tasks/<id>/evidence/`. A task without a feature has no
  paths.

### R2 Create

- `hive task create` requires `--feature F`. Without it, it fails in phase
  `validate` and writes nothing.
- `--spec` marks the spec task of the feature. Store it in the record as
  `"spec": true`.
- For a task without `--spec`, `.hive/spec/F/README.md` must exist on the
  base branch of the project: `git -C <repo> cat-file -e <base>:.hive/spec/F/README.md`.
  Else fail in phase `validate`, write nothing, and name the missing path.
- For a task with `--spec`, skip that check.

### R3 Edit feature

- `hive task edit --task ID --feature F` sets the feature of a record that has
  none, in any status. Only the controlling role can do it. It does not check
  the README. It refuses a record that already has a feature.
- `--brief` keeps its current rules. One call can give `--brief` or
  `--feature`, not both.

### R4 List

`hive task list --feature F` shows only the tasks of that feature. It can join
`--project` and `--ready`.

### R5 Show

`hive task show ID` prints the feature and the brief and evidence paths when
the task has a feature. It prints a warning line for each linked file that is
missing: the brief and the README. Check in the worktree path when the record
has a worktree that is not removed, else on the base branch with
`git cat-file -e`. A missing file is a warning, not a failure. With `--json`,
put the warnings in `result`.

### R6 Allowlist and layout

- Add `!.hive/spec/` and `!.hive/spec/**` after the `.hive/*` rule.
- `gitignore_test.go`: `.hive/spec/x/README.md` and
  `.hive/spec/x/tasks/1-y/evidence/V1.md` are not ignored. `.hive/state/` and
  `.hive/tmp/` stay ignored.
- Add `.hive/spec/` as tracked to the home layout in `AGENTS.md` and
  `.hive/README.md`. Keep the text in ASD-STE100 Strict.

### R7 Manager instructions

In the manager skill and the manager template, `hive task create` shows
`--feature F`, and one line tells that `--spec` marks the spec task, which
lands `.hive/spec/F/README.md` before any other task of the feature.

### R8 Tests and brief copy

- Write each test first and see it fail for the correct reason. Use temporary Git repositories for the README check. Do not compute an expected value with the code
  under test.
- Copy this brief to `spec/hive-v1/tasks/<task-id>.md` with a baseline of 0.
  When the copy causes a violation, change the words of that sentence and keep its meaning. Record each change in the evidence.

## Not in scope

The brief copy into the worktree, `git check-ignore` (task B). The move of old
files and the record migration (task C). `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each new test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Set `HIVE_HOME` and `XDG_CONFIG_HOME` in a temporary home. Register a temporary Git project.<br>Create without `--feature`.<br>Create with `--feature f` while the base README is absent.<br>Create with `--feature f --spec`.<br>Commit `.hive/spec/f/README.md` on the base branch. Then create with `--feature f`.<br>Run `task list --feature f`.<br>Run `task show` for a task without a linked brief on the base branch.<br>Run `task edit --feature g` on an old record without a feature. Repeat the edit. | Fail, fail naming the path, ok, ok, both tasks listed, warning line for the brief, ok, refused. |
| V3 | `git check-ignore -q .hive/spec/x/README.md; echo $?` in the worktree | Exit 1 (not ignored). `-v` exits 0 on a negation match, so do not use it. |
| V4 | `git diff dev --stat` | Only the files of the source table changed. |
| V5 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `feat(task): add the task
feature`. Push the task branch. Write the evidence. Run `hive task complete`.
Send the commit SHA and the V1-V5 results to the manager with
`hive inbox send`. Read only your own inbox.
