# Brief copy into the worktree

| Item | Value |
| --- | --- |
| Wave | H2, task B of 3 |
| Tier | decent |
| Depends on | 46-task-feature. Base: `dev` after it lands |
| Grant | `grants.md 2026-10-19 H2 spec layout wave` |
| Evidence | `evidence/<task-id>/` at the root of the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The brief in `.hive/state/briefs/<id>.md` is the canonical copy. Hive copies
it to the tracked path in the worktree. The worker commits it with its work.
Hive never commits. The plan is
`projects/hive/plans/home-spec-planning/MAP.md` in the home, decisions 6, 9,
and 9b. Do not change a decision in this brief. When the code does not agree
with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- 46-task-feature added the `feature` field, the derived paths, and the
  allowlist rule `!.hive/spec/**` in the hive project.
- `internal/worktree/create.go` makes the worktree. `internal/task/edit.go`
  changes the state brief.
- A record without a feature stays valid.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/worktree/create.go` and its tests | R1, R2. |
| `internal/task/edit.go` and its tests | R3. |
| `.agents/skills/hive-worker/SKILL.md`, `.agents/skills/hive-verifier/SKILL.md`, `internal/agent/brief.go` and its tests | R4. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `spec/hive-v1/tasks/<task-id>.md` | R5. New. A copy of this brief. |

Do not edit other files.

## Requirements

### R1 Ignore check

For a task with a feature, `hive worktree create` runs
`git -C <repo> check-ignore -q --no-index .hive/spec/<feature>/tasks/<id>/brief.md`
before it creates the worktree. When Git ignores the path, hive fails in phase
`validate` with a message that names the path and says to allow `.hive/spec/`
in the `.gitignore` of the project. Hive creates nothing. Hive never edits
that `.gitignore`.

### R2 Copy on create

After the worktree exists, `hive worktree create` copies the state brief to
`<worktree>/.hive/spec/<feature>/tasks/<id>/brief.md`. It makes the folders.
It reports the copy as an effect. It does not commit. For a task without a
feature, it copies nothing and checks nothing.

### R3 Copy on edit

`hive task edit --task ID --brief FILE` also writes the new brief to the
worktree path when the task has a feature and a worktree that is not removed.
It reports that write as an effect. It does not commit.

### R4 Worker and verifier text

- The worker skill: read the brief at
  `.hive/spec/<feature>/tasks/<id>/brief.md` in the worktree. Commit it with
  the work. Write each `V<n>.md` to
  `.hive/spec/<feature>/tasks/<id>/evidence/`. For a task without a feature,
  use the evidence path in the brief.
- The verifier skill: write `VERIFIED.md` in that same evidence folder. Never
  write `VERIFIED.md` at the root of the checkout.
- The generated worker and verifier prompts name the brief and evidence paths
  when the task has a feature. Tests assert the exact prompt text as literal
  strings.
- Keep the text in ASD-STE100 Strict.

### R5 Tests and brief copy

- Write each test first and see it fail for the correct reason. Use temporary
  Git repositories.
- Copy this brief to `spec/hive-v1/tasks/<task-id>.md` with a baseline of 0.
  When the copy causes a violation, change the words of that sentence and keep its meaning. Record each change in the evidence.

## Not in scope

The move of old files and the record migration (task C). `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each new test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | In a temporary home with a temporary Git repository that allows `.hive/spec/`, run `worktree create` for a feature task. Run `git status --short` in the worktree. Run `task edit --brief` with new text. Read the worktree brief with `cat`. Repeat with a second temporary Git repository that ignores `.hive/spec/`. Run `worktree create` there. | The brief file exists. Git status shows it as untracked. Git does not ignore it. The edit changes the worktree file. The second create fails and names the path. The second create creates no worktree. |
| V3 | `git diff dev --stat` | Only the files of the source table changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

Live checks with herdr must clean up each workspace that they make.

## Finish

Commit on the task branch with a subject such as `feat(worktree): copy the
brief into the worktree`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V4 results to the manager
with `hive inbox send`.
