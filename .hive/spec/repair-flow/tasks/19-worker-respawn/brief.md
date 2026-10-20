# Worker respawn

| Item | Value |
| --- | --- |
| Wave | wave 3 of 4, task B |
| Tier | decent |
| Depends on | none. Base: `dev` at `39a35e6c` or later |
| Grant | `grants.md 2026-10-19 worker respawn task` |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

Today a second worker spawn into an `assigned` task keeps the old acting agent,
fails when herdr closed the task workspace, and needs a new explicit name.
After this task, the spawn sets the acting agent, reopens the worktree
workspace, and picks a free name. Do not change a decision in this brief. When
the code does not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

The friction entries are, in `reference/dogfood/friction/entries/`:
`worker-spawn-keeps-old-acting-agent.md`,
`verify-fail-repair-worker-no-workspace.md`, and
`exited-agent-name-blocks-reuse.md`. Read them. Do not edit a file in
`reference/dogfood/friction/`. The manager resolves the entries after the
landing.

## Known facts

- `internal/agent/spawn_worker.go`: `resolveTaskPane` reads the task.
  `checkWorktreePane` fails with `worktree pane mismatch: pane_id` when the pane
  of the task record is gone. `assignTask` assigns the task to the new agent.
  `resolveVerifierTopology` already calls `herdr.WorktreeOpen`
  (`internal/lib/herdr/worktree.go`, herdr `worktree.open`).
- `internal/task/transition.go`: `assigned -> assign` exists and sets
  `ActingAgent`. It refuses a finished agent.
- `internal/agent/register.go` fails with `agent exists: NAME` when a record or
  claim has the name.
- `cmd/agent/spawn/spawn.go` passes `NameSet` when the command has the `--name` flag.
- `reference/herdr/api/worktree.md` documents `worktree.open`. Read it before
  you write a herdr call.
- On 2026-10-19, after a verify fail for 16-quiet-no-nudge, `hive agent stop`
  of the verifier closed the task workspace. The task record kept the old
  pane. `hive worktree create` refused with `task has a worktree`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/agent/spawn_worker.go`, `internal/agent/spawn.go`, `internal/agent/register.go` | R1 to R3. Only the files that the fix needs. |
| `cmd/agent/spawn/spawn.go` | R3. Only when the fix needs it. |
| Tests beside each changed file | R1 to R3 tests. |
| `.agents/skills/hive-manager/SKILL.md` | R4. |
| `internal/agent/templates/project-agents.md` | R4. |
| `internal/agent/manager_prompt_test.go` | R4 test. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `spec/hive-v1/tasks/<task-id>.md` | R6. New. A copy of this brief. |

Do not edit other files.

## Requirements

### R1 Acting agent

- A worker spawn into an `assigned` task sets the acting agent of the task to
  the new worker. The new worker can then run `hive task complete` with no
  manual `hive task assign`.
- The spawn fails while the old acting worker is live. The error names the old
  worker, for example `task 16-quiet-no-nudge has a live worker:
  t16-worker-1`. The spawn makes no agent record and changes no task record.

### R2 Workspace

When the pane of the worktree in the task record is gone, the spawn reopens the
worktree workspace of the task with `worktree.open`, the API that hive already
uses. It updates `workspace_id`, `tab_id`, and `root_pane` of the worktree in
the task record. It does not make a second worktree. When the pane is live, the
spawn behaves as today.

### R3 Free name

When `--name` is absent, hive picks the next free name for a worker or a
verifier: `tNN-ROLE-N`. `NN` is the leading number of the task ID (`16` for
`16-quiet-no-nudge`). `N` is the lowest number from 1 with no agent record and
no claim of that name. An explicit `--name` keeps the current rule.

### R4 Instructions

In both instruction files, make the worker and verifier spawn text match R1 to
R3: `--name` is optional, and a respawn after a verify fail needs no
`hive task assign`. Remove each manual `hive task assign` step after a
respawn. Add a check of this text to `internal/agent/manager_prompt_test.go`.
Run it, and make sure that it fails before you change the template. Write the
text in ASD-STE100 Strict. The counts in `testdata/ste-baseline.json` must not
rise.

### R5 Tests

Work test-first. Use the fake herdr of `internal/agent` for herdr calls. Write
each test, run it, and make sure that it fails for the correct reason before
you change the code. Write one test for each item. One test checks that a
respawn sets the acting agent. One test checks that a live old worker blocks
the respawn and appears in the error. One test checks that a closed task pane
calls `worktree.open` and updates the record. One test checks that hive does
not create another worktree. One test checks that a spawn with no `--name`
picks `t16-worker-2` when `t16-worker-1` exists. One test checks that a
duplicate explicit `--name` returns `agent exists`. Show in the evidence that
each new test fails against the old code, then passes.

### R6 Brief copy

Copy this brief to `spec/hive-v1/tasks/<task-id>.md` with a baseline of 0.
When the copy causes a violation, change the words of that sentence and keep
its meaning. Record each change in the evidence.

## Not in scope

`hive task reopen` (wave 3 task A), the inbox, the friction folder, `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each R5 and R4 test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Read the R4 text in both instruction files | Present in both. No manual `hive task assign` step after a respawn remains. |
| V3 | `git diff dev --stat` | Only the files of the source table changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `fix(agent): set the acting
agent on a worker respawn`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V4 results to the manager
with `hive inbox send`. Read only your own inbox.

## Manager decision

Manager decision, 2026-10-19: Valid task IDs start with a number. A task ID
without a leading number cannot pass validation. Do not add fallback code or a
fallback test. Keep the `tNN-ROLE-N` name, the lowest free `N`, and the current
rule for explicit `--name`.
