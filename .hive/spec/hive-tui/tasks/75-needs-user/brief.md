# Task needs_user

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, wave A |
| Tier | small |
| Depends on | 73-hive-tui-spec. Base: `dev` after it lands |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Add the `needs_user` field of a task (README G8, task part). The feature README is `.hive/spec/hive-tui/README.md`. Read it first. The README
gives approved decisions. Do not change a decision. When the code does not agree
with this brief or the README, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/lib/records/task.go` and its tests | R1. |
| `internal/task/edit.go`, `show.go`, `list.go`, `render.go` and their tests | R2, R3. |
| `cmd/task/edit/` and its tests | R2. Flags only. |
| `.agents/skills/hive-manager/SKILL.md`, `internal/agent/templates/project-agents.md`, and the tests that assert their text | R4. |
| `testdata/ste-baseline.json` | Only to lower a count. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### R1 Field

The task record has an optional `needs_user` string
(`json:"needs_user,omitempty"`). It is at most 200 characters, one line.

### R2 Edit

`hive task edit --task ID --needs-user TEXT` sets it.
`hive task edit --task ID --clear-needs-user` clears it. Both work in any
status. Only the controlling role can use them. One call gives one of
`--brief`, `--feature`, `--needs-user`, or `--clear-needs-user`.

### R3 Show and list

`hive task show` prints `needs user: TEXT` when the task has that value. `hive task list`
marks such a task with `[needs user]` at the end of its line. With `--json`
the field is in the record.

### R4 Manager text

The manager skill and the manager template tell the manager to set
`needs_user` before it stops to ask the user for a decision. The manager clears
the field when the user gives a decision. Keep the text in ASD-STE100 Strict.

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
  table. A small conflict in `cmd/root.go` or `testdata/ste-baseline.json` can
  occur at the landing.

## Not in scope

The goal record. The TUI. `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each new test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | In a temp home, set `needs_user` on a verified task. Show and list it. Clear the field, then show it again. Try to set the field as an agent without control. | Show and list display the text before the clear. They omit it after the clear. Hive refuses the other agent's edit. |
| V3 | `git diff dev --stat` | Only the files of the source table and the evidence changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `feat(task): add needs_user`.
Push the task branch. Write the evidence. Run `hive task complete`. Send the
commit SHA and the V1-V4 results to the manager with `hive inbox send`.
