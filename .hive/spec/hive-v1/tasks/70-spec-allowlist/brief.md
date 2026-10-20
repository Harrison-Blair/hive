# Spec task allowlist

| Item | Value |
| --- | --- |
| Tier | decent |
| Depends on | none. Base: `dev` at `450515f9` or later |
| Grant | `grants.md 2026-10-19 spec task allowlist fix` |
| Evidence | `.hive/spec/hive-v1/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The first spec task of a feature cannot start in a project whose `.gitignore`
ignores `.hive/spec/`. The friction entry is
`reference/dogfood/friction/entries/spec-task-needs-allowlist-first.md`. Read
it. Do not edit a file in `reference/dogfood/friction/`. The user approved
this fix. Do not change a decision in this brief. When the code does not agree
with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `internal/worktree/create.go` runs
  `git -C <repo> check-ignore -q --no-index .hive/spec/<feature>/tasks/<id>/brief.md`
  before it creates the worktree. When Git ignores the path, the command
  fails in phase `validate`.
- The task record has `"spec": true` for a spec task.
- Hive never commits and never edits the `.gitignore` of a project.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/worktree/create.go` and its tests | R1, R2. |
| `internal/agent/brief.go` and its tests, and each caller of a changed prompt function | R3. Only the call in a caller. |
| `.agents/skills/hive-worker/SKILL.md` | R3. |
| `testdata/ste-baseline.json` | Only to lower a count. |
| `.hive/spec/hive-v1/tasks/<task-id>/brief.md` | The brief copy that hive puts in the worktree. Commit it. |

Do not edit other files.

## Requirements

### R1 Spec task

For a spec task, when Git ignores the brief path:

- `hive worktree create` does not fail. It creates the worktree and copies the
  brief as for each task.
- It runs `git -C <worktree> add -f .hive/spec/<feature>/tasks/<id>/brief.md`.
  It does not commit.
- It reports one effect with the kind `allowlist_needed`. The text tells the
  worker to add `!.hive/spec/` and `!.hive/spec/**` to the `.gitignore` of
  the project.

### R2 Other tasks

- A task that is not a spec task keeps the current failure when Git ignores
  the path.
- When Git does not ignore the path, each task keeps the current behavior. No
  `git add` and no `allowlist_needed` effect.

### R3 Worker text

- The worker skill tells the worker: when the worktree create shows
  `allowlist_needed`, the spec task adds `!.hive/spec/` and `!.hive/spec/**`
  to the `.gitignore` of the project, after the rules that ignore the path.
  It commits that change with the feature README.
- The generated worker prompt for a spec task gives the same instruction.
  Tests assert the exact prompt text as literal strings.
- Keep the text in ASD-STE100 Strict.

### R4 Tests

- Write each test first and see it fail for the correct reason. Use temporary Git
  projects. One project has a `.gitignore` that starts with `*` and does not allow
  `.hive/`. One project allows `.hive/spec/`.
- When `ste_test.go` fails on the brief copy, change the words of each
  sentence that causes a violation in the copy. Keep its meaning. Record each
  change in the evidence.

## Not in scope

The stenographer project (it gets a manual allowlist commit). The friction log.
`main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each new test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | In a temporary home with a temporary project whose `.gitignore` is `*` and `!*/`: create a spec task and its worktree, then run `git status --short` in the worktree. Commit `.hive/spec/<feature>/README.md` on the base with `git add -f`. Then create a task that is not a spec task in the same feature and run its worktree create. | The spec worktree exists. The status shows the brief as `A`. The output shows the `allowlist_needed` effect. The second create fails in phase `validate` and makes no worktree. |
| V3 | `git diff dev --stat` | Only the files of the source table and the evidence changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

Live checks with herdr must clean up each workspace that they make.

## Finish

Commit on the task branch with a subject such as `fix(worktree): stage the
spec brief in an ignored path`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V4 results to the manager
with `hive inbox send`.
