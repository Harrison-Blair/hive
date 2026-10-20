# Store list keeps long keys

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, repair after wave C |
| Tier | decent |
| Depends on | none. Base: `dev` |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

`store.List` skips each record whose file name fails `records.ValidID`. That
check is for a project ID and allows at most 18 characters. Many valid task
IDs and agent keys are longer, for example `96-tasks-evidence-view` and
`manager-stenographer-g7`. Each caller that lists with `store.List` loses
those records with no error. The TUI Tasks and Queue screens miss tasks for
this reason. The friction entry is
`reference/dogfood/friction/entries/store-list-skips-long-keys.md`. Do not
edit it. When the code does not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `listUnlocked` in `internal/lib/store/store.go` calls `records.ValidID(id)`
  on each file name and skips a file that fails.
- `validateSegment` in the same file refuses an empty name, `/`, and `..`.
- `records.ValidTaskID` and `records.ValidGoalID` exist for other kinds.
- Callers of the list include `internal/task/`, `internal/supervise/`,
  `internal/tui/`, `internal/doctor/`, `internal/goal/`, and
  `internal/lib/inbox/`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/lib/store/` and its tests | R1, R2. |
| A test in a caller package that proves R3 | R3. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy, when `hive task create` did not write it. |

Do not edit other files.

## Requirements

### R1 Keep each valid key

`store.List` and `ListIn` return each record whose file name is a valid store
segment, with any length. They still skip a file that does not end in
`.json`, and a name that `validateSegment` refuses. Read each file name as
data. Do not let a name select a path outside the kind folder.

### R2 Store test

Write the test first and see it fail on the base. List a kind with records
named `p1`, `96-tasks-evidence-view`, and `manager-stenographer-g7`, and one
file `notes.txt`. Expect the three records as literal names.

### R3 Caller test

Add one test in the TUI or the task package that lists a task with an ID
longer than 18 characters and finds it. Write the test first and see it fail
on the base.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run the R2 and R3 tests against the base code, then against the task code | Each fails on the base and passes on the task branch. |
| V2 | Build the task binary. On a scratch home (`HIVE_HOME` and a temp `XDG_CONFIG_HOME`) with a task `96-tasks-evidence-view`, open the TUI Tasks and Queue screens | Both screens show the task. |
| V3 | `git diff origin/dev --stat` | Only the files of the source table changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0. |

Never open the Chat input on the real home. Close each herdr tab and
workspace that you open.

## Finish

Commit on the task branch with a subject such as `fix(store): keep long keys
in a list`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V4 results to the
manager with `hive inbox send`. Read only your own inbox.
