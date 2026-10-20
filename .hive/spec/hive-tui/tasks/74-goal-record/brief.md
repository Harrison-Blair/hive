# Goal record

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, wave A |
| Tier | decent |
| Depends on | 73-hive-tui-spec. Base: `dev` after it lands |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Add the goal record and the `hive goal` command (README G7, and the wave part of G8). The feature README is `.hive/spec/hive-tui/README.md`. Read it first. The README holds the approved decisions. Do not change a decision. When the code does not agree
with this brief or the README, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/lib/records/` (a new `goal.go`) and its tests | R1. |
| `internal/goal/` (new, one leaf for each subcommand or one package with a `doc.go`) and its tests | R2, R3. |
| `cmd/goal/` (new) with `New() *cobra.Command`, and `cmd/root.go` to register it | R2. |
| `.agents/skills/hive-manager/SKILL.md`, `internal/agent/templates/project-agents.md`, and the tests that assert their text | R4. |
| `testdata/ste-baseline.json` | Only to lower a count. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### R1 Record

- A goal is `.hive/state/goals/<id>.json`. The ID is `g<n>-<slug>`, with `<n>`
  from a new counter `goals` and a kebab-case slug.
- Fields: `id`, `project`, `title`, `grant` (the grant line text), `status`
  (`active` or `done`), `waves`, `created_at`, `updated_at`.
- Each wave: `id` (`A`, `B`, ... in order), `title`, `done_when` (one of
  `local`, `dev`, `pr`, `main`, `release`), `tasks` (task IDs), `needs_user`
  (optional text), `status` (`open` or `done`).

### R2 Commands

- `hive goal create --project P --title T --slug S --grant TEXT` makes an
  active goal with no waves.
- `hive goal edit --goal ID` with one of: `--add-wave TITLE --done-when W`,
  `--wave X --add-task TASK`, `--wave X --needs-user TEXT`,
  `--wave X --clear-needs-user`, `--wave X --status open|done`,
  `--status active|done`. A task must exist and belong to the project of the
  goal. One task is in at most one wave of a goal.
- `hive goal show ID` and `hive goal list [--project P]`.
- Only the coordinator and the manager of the project change a goal. The text
  output and the `--json` envelope obey the Command output rules of
  `AGENTS.md`. Log each command.

### R3 Wave progress

`hive goal show` prints, for each wave, the count of tasks that landed, from
the task records: a task landed when its delivery has a commit and `pushed`
is true. It does not set a status.

### R4 Manager text

The manager skill and the manager template list the goal commands in the
command block and tell the manager to add each new task to its wave. Keep the
text in ASD-STE100 Strict.

## Rules for each task

- Each Go change is test-first. See each new test fail for the correct
  reason. Do not compute an expected value with the code under test.
- When `ste_test.go` fails on the brief copy, change the words of each
  sentence in the copy that causes a violation. Keep its meaning. Record each
  change in the evidence.
- Live checks use a temp home (`HIVE_HOME`) and a temp `XDG_CONFIG_HOME`, and
  clean up each herdr workspace that they make. Do not run `rm -rf` on an
  unquoted variable.
- Other wave A tasks run at the same time. Edit only the files of your source
  table. Expect a small conflict in `cmd/root.go` or
  `testdata/ste-baseline.json` when the changes land.

## Not in scope

The TUI screens (waves B and C). Task `needs_user` (another wave A task).
`main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each new test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | In a temp home, create a goal and two waves. Add one task to wave A, then try to add it to wave B. Set and clear `needs_user`. Show and list goals. Try one edit as another agent. | The second add fails. Show prints each wave and task with `0/1 landed`. Hive rejects the other agent's edit. |
| V3 | `git diff dev --stat` | Only the files of the source table and the evidence changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `feat(goal): add the goal
record`. Push the task branch. Write the evidence. Run `hive task complete`.
Send the commit SHA and the V1-V4 results to the manager with
`hive inbox send`.
