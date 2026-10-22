# Spawn project from the task

| Item | Value |
| --- | --- |
| Wave | wave 2 of 4, task C |
| Tier | decent |
| Depends on | none. Base: `dev` at `1a7a75a6` or later |
| Grant | `grants.md 2026-10-19 verifier project task` |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

Today a worker or verifier spawn with `--pane` requires `--project`, but the
manager skill shows a verifier spawn without `--project`. After this task, a
spawn with `--task` takes the project from the task record. The task also adds
the re-verify rule to the manager instructions. Do not change a decision in
this brief. When the code does not agree with this brief, stop and report it
with `hive inbox send --to manager-hive TEXT --agent NAME`.

The friction entries are
`reference/dogfood/friction/entries/verifier-spawn-pane-requires-project.md`
and
`reference/dogfood/friction/entries/resolved/verify-fail-no-reverify-workspace.md`.
Read them. Do not edit a file in `reference/dogfood/friction/`. The manager
updates the entries after the landing.

## Known facts

- `internal/agent/spawn.go`: `choosePane` returns at once when a user passes `--pane`.
  It does not read the task. Without `--pane`, it calls `resolveTaskPane` in
  `internal/agent/spawn_worker.go`.
- `resolveTaskPane` reads the task and fails with `task ID is in project P`
  when `--project` differs from the task project.
- `internal/agent/register.go`: `validateRegistration` fails with
  `ROLE requires --project and --task` for a worker or verifier with no
  project.
- `.agents/skills/hive-manager/SKILL.md` (section `Verifier spawn`) and
  `internal/agent/templates/project-agents.md` (section `Commands`) show the
  verifier spawn command.
- `internal/agent/manager_prompt_test.go` and `skills_test.go` check text of
  these files.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/agent/spawn.go`, `internal/agent/spawn_worker.go`, `internal/agent/register.go` | R1. Only the files that the fix needs. |
| `internal/agent/*_test.go`, `cmd/agent/spawn/*_test.go` | R1 tests. |
| `.agents/skills/hive-manager/SKILL.md` | R2, R3. |
| `internal/agent/templates/project-agents.md` | R2, R3. |
| `internal/agent/manager_prompt_test.go` | R3 test. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `spec/hive-v1/tasks/<task-id>.md` | R5. New. A copy of this brief. |

Do not edit other files.

## Requirements

### R1 Project from the task

- A worker or verifier spawn with `--task` and no `--project` takes the
  project from the task record. This applies with `--pane` and without
  `--pane`.
- When a user gives a `--project` value that differs from the task project, the
  spawn fails. It returns `task ID is in project P` and makes no agent record.
- When the task does not exist, the spawn fails as today.

Work test-first. Write each test, run it, and make sure that it fails for the
correct reason before you change the code. The tests cover these cases:
1. A verifier spawn with `--pane` and no `--project`.
2. A worker spawn with `--pane` and no `--project`.
3. A spawn with `--pane` and a different `--project`.
Show in the evidence that each new test fails against the old code, then passes.

### R2 Spawn command in the docs

In both instruction files, the worker and the verifier spawn commands show no
`--project` and no `--pane`. Add one sentence: hive takes the project from the
task.

### R3 Re-verify rule

In both instruction files, add this rule: after a verify fail, spawn a fresh
verifier with no `--pane`. Do not make placeholder tabs. Add a check of this
text to `internal/agent/manager_prompt_test.go`. Run it, and make sure that it
fails before you change the template.

### R4 STE

Write the new instruction text in ASD-STE100 Strict. Use each term only with
its meaning in `STE-GLOSSARY.md`. The counts in `testdata/ste-baseline.json`
must not rise.

### R5 Brief copy

Copy this brief to `spec/hive-v1/tasks/<task-id>.md` with a baseline of 0.
When the copy causes a violation, change the words of that sentence and keep
its meaning. Record each change in the evidence.

## Not in scope

The inbox, the supervisor, the friction folder, a new hive command, `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each R1 and R3 test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Read the `Verifier spawn` section of the manager skill and the `Commands` section of the template | R2 and R3 text is present in both. |
| V3 | `go test -run TestSTE .` | Pass. No count in `testdata/ste-baseline.json` rose. |
| V4 | `git diff dev --stat` | Only the files of the source table changed. |
| V5 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `fix(agent): take the spawn
project from the task`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V5 results to the manager
with `hive inbox send`. Read only your own inbox.
