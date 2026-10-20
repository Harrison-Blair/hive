# Verify clean branch

| Item | Value |
| --- | --- |
| Wave | wave 4 of 4, task C |
| Tier | decent |
| Depends on | none. Base: `dev` at `5d879894` or later |
| Grant | `grants.md 2026-10-19 verify clean branch task` |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

Today, `hive task verify` records a verdict before the verifier commits or
pushes `VERIFIED.md`. Then `hive land` can merge an old verdict file. After this task,
`hive task verify` refuses a dirty worktree and an unpushed branch. Do not
change a decision in this brief. When the code does not agree with this brief,
stop and report it with `hive inbox send --to manager-hive TEXT --agent NAME`.

The friction entry is
`reference/dogfood/friction/entries/verifier-verdict-without-commit.md`. Do not
edit a file in `reference/dogfood/friction/`. The manager resolves the entry
after the landing.

## Known facts

- `internal/task/verify.go` has `Verify`. It calls the transition of
  `internal/task/transition.go`.
- `internal/lib/gitx/gitx.go` has `IsDirty`, `RefExists`, `Tip`, and
  `HasRemote`.
- The task record has `worktree.path` and `worktree.branch`. The remote is
  `origin`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/task/verify.go` | R1. |
| `internal/lib/gitx/gitx.go` | R1. Only for a new helper that the check needs. |
| Tests beside each changed file | R2. Use a local bare repository as `origin`. |
| `.agents/skills/hive-manager/SKILL.md` | R3. |
| `.agents/skills/hive-verifier/SKILL.md` | R3. |
| `internal/agent/templates/project-agents.md` | R3. |
| `internal/agent/manager_prompt_test.go` | R3 test. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `spec/hive-v1/tasks/<task-id>.md` | R4. New. A copy of this brief. |

Do not edit other files.

## Requirements

### R1 Check

Before it records a pass or a fail, `hive task verify` checks the task
worktree:

- When Git reports uncommitted changes in the worktree, it fails with
  `worktree is dirty: commit and push before verify` and changes nothing.
- It fetches `origin BRANCH`. When the remote branch does not exist, it fails
  with `branch not pushed: BRANCH`. When the local tip differs from the remote
  tip, it fails with `branch not pushed: BRANCH has commits not on origin`.
  Each failure changes nothing.
- When the task has no worktree, it behaves as today.

### R2 Tests

Work test-first. Write each test, run it, and make sure that it fails for the
correct reason before you change the code. The tests cover at least these cases:

- A dirty worktree refuses.
- A branch with no remote branch refuses.
- A branch ahead of its remote refuses.
- A clean, pushed branch passes.

The evidence shows each new test fail against the old code and pass.

### R3 Instructions

In the manager skill, the verifier skill, and the template, add this rule: the
verifier commits and pushes `VERIFIED.md` before it runs `hive task verify`.
Add a check of the template text to `internal/agent/manager_prompt_test.go`.
Run it, and make sure that it fails before you change the template. Write the
text in ASD-STE100 Strict. The counts in `testdata/ste-baseline.json` must not
rise.

### R4 Brief copy

Copy this brief to `spec/hive-v1/tasks/<task-id>.md` with a baseline of 0.
When the copy causes a violation, change the words of that sentence and keep
its meaning. Record each change in the evidence.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each R2 and R3 test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Read the R3 text in the three instruction files | Present in each. |
| V3 | `git diff dev --stat` | Only the files of the source table changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `fix(task): refuse a verify
of a dirty or unpushed branch`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V4 results to the manager
with `hive inbox send`.
