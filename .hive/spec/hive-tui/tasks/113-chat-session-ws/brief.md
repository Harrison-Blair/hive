# User chat from the hive session workspace

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, repair after wave C |
| Tier | decent |
| Depends on | none. Base: `dev` |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Task 94 added three rules that stop an agent from sending user chat. Rule 3
refuses a caller pane in the workspace of a live manager, worker, or
verifier. The manager of the hive project runs in a tab of the hive session
workspace. The TUI of the user runs in the same workspace. Thus rule 3
refuses the user. The TUI shows this error: `chat send is for the user only: only the user can send chat; rule 3 (caller pane shares a live manager workspace)`.
The user cannot send chat. This is a high priority fault. When the code does
not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `checkCallerContext` in `internal/chat/chat.go` applies rule 2 and rule 3.
  Rule 3 compares `pane.WorkspaceID` with the `WorkspaceID` of each live
  agent record with the role `manager`, `worker`, or `verifier`.
- Rule 1 (`checkAgentCaller`) refuses `--agent` and a pane that belongs to a
  live agent.
- The hive session workspace hosts the live coordinator. The coordinator
  record has its `WorkspaceID`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/chat/chat.go` and its tests | R1, R2. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy, when `hive task create` did not write it. |

Do not edit other files.

## Requirements

### R1 Rule 3 skips the hive session workspace

Rule 3 does not apply when the caller pane is in the workspace of the live
coordinator. In that workspace, rule 1 and rule 2 still apply: a pane of a
live agent and a working directory below `projects/` are still refused.
When no live coordinator exists, rule 3 applies as before.

### R2 Tests

Write each test first and see it fail on the base. Use fixture agent records
and a fake herdr pane lookup. Cover:

- A pane with no agent in the workspace of the live coordinator, where a live
  manager is also in that workspace: the send succeeds.
- The pane of that live manager: refused by rule 1.
- A pane with no agent in the workspace of a live worker, which is not the
  coordinator workspace: refused by rule 3.
- A pane in the coordinator workspace with a working directory below
  `projects/`: refused by rule 2.

Use literal expected error text.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run the R2 tests against the base code, then against the task code | Each new test fails on the base where the behavior changes, and passes on the task branch. |
| V2 | Read the code path for each rule and compare it with R1 | The code skips rule 3 only for the workspace of the live coordinator. |
| V3 | `git diff origin/dev --stat` | Only the files of the source table changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0. |

Never send chat on the real home. Never work in workspace `wBC`. Stop and
close only the pid and pane that you started.

## Finish

Commit on the task branch with a subject such as `fix(chat): let the user
send from the hive session workspace`. Push the task branch. Write the
evidence. Run `hive task complete`. Send the commit SHA and the V1-V4 results
to the manager with `hive inbox send --to manager-hive`.
