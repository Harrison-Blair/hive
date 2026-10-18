# T03 task lifecycle

| Item | Value |
| --- | --- |
| Task ID | T03-task-lifecycle |
| Wave | 2 |
| Tier | decent |
| Depends on | T01, T22, CP1 |
| Blocks | T09, T11, T12 |
| Branch | `task/T03-task-lifecycle` |
| Worktree | `projects/hive/worktrees/T03-task-lifecycle/` |
| Evidence | `evidence/T03-task-lifecycle/` |

## Overview

This task gives hive its task records: the ID scheme, the brief store, the
status machine, the dependency graph, and the seven `task` commands. A caller
names itself with `--agent`. T11 later adds the lookup of the caller from its
herdr pane and the inbox records on completion. T09 adds the worktree fields,
and T12 adds the delivery fields.

## User flows

1. A manager runs `hive task create --project hive --title "Spawn worker" --depends-on 3-task-store --brief brief.md`. Hive prints the new ID `12-spawn-worker` and the brief path.
2. The manager runs `hive task list --project hive --ready`. Hive lists each task whose dependencies hold the status `verified`.
3. A worker runs `hive task show 12-spawn-worker --brief` and reads its brief text.
4. The worker runs `hive task complete --task 12-spawn-worker --agent worker-hive-3`. Hive moves the task to `completed`.
5. A verifier runs `hive task verify --task 12-spawn-worker --agent verifier-hive-4 --fail --reason "V3 output differs"`. The task returns to `assigned`.
6. The worker corrects the work and runs `complete` again. The verifier runs `hive task verify --task 12-spawn-worker --agent verifier-hive-4`. The task becomes `verified`, and the new verification replaces the failed one.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `cmd/root.go` | Register the `task` command. |
| `cmd/task/task.go` | New. The `task` parent command and the `--agent` flag. |
| `cmd/task/create/create.go` | New. |
| `cmd/task/assign/assign.go` | New. |
| `cmd/task/list/list.go` | New. |
| `cmd/task/show/show.go` | New. |
| `cmd/task/complete/complete.go` | New. |
| `cmd/task/verify/verify.go` | New. |
| `cmd/task/cancel/cancel.go` | New. |
| `internal/task/doc.go` | New. Package summary and file map. |
| `internal/task/create.go` | New. ID, slug, brief, dependencies, cycle check. |
| `internal/task/transition.go` | New. The status machine and its errors. |
| `internal/task/assign.go`, `complete.go`, `verify.go`, `cancel.go` | New. One file for each command. |
| `internal/task/list.go`, `show.go` | New. Queries and the ready rule. |
| `internal/task/render.go` | New. Text rendering. |
| `internal/lib/records/task.go` | New. The task record type. |
| `internal/lib/records/doc.go` | Add the map line. |

## Store use

Use these calls:

- `store.WithLock` for one create. The reads, the counter, the brief, and the task
  record run under one lock.
- `LockedStore.ReadIn`, `WriteIn`, and `Next`.
- The collection `["tasks"]` for a task list.
- `store.Read` or `LockedStore.ReadIn` for the lock-free record read of `show`.

Method matrix: `Store.Read` and `Store.ReadIn` take no lock. A standalone list
takes no lock. A related read uses one `WithLock` callback. Each standalone
mutation uses bounded acquisition. A `LockedStore` method never acquires another
lock. Do not read or write a task record file directly.

## Requirements

### AC1 Task creation

- `hive task create --project ID --title TEXT [--slug SLUG] [--depends-on ID,ID] [--brief FILE]` needs a registered project.
- The task ID is `<n>-<slug>`. `n` comes from the counter `task`. The home has one task counter, so a task ID is unique across each project of the home. No command needs a project to select a task.
- The slug is the `--slug` value, else the title in lowercase with each run of characters outside `[a-z0-9]` as one `-`, trimmed of `-` at both ends. Both forms must match `^[a-z0-9]+(-[a-z0-9]+)*$` and have at most 40 characters. A 40-character slug is valid, so an actual task ID can be longer than 18 characters. Otherwise the command fails with `invalid slug: <value>` before it takes a counter value.
- The brief is the content of `--brief FILE`, or of stdin when FILE is `-`, or the title. Hive reads at most 1 MiB and fails above it. The read has a total deadline of 30 seconds. The deadline is an explicit parameter of `readBrief`, so a test can shorten it. Check that a brief file is a regular file before the open. A stalled input expires before the lock and changes nothing. Hive writes the brief to `.hive/state/briefs/<task-id>.md` atomically.
- Each dependency must exist in the same project. A dependency path that returns to the new task fails with `dependency cycle: <a> -> <b> -> <a>`.
- The record has `id`, `project`, `title`, `controlling_role`, `acting_agent`, `status`, `depends_on`, `verification`, `worktree`, `delivery`, `created_at`, and `updated_at`. The controlling role is `manager-<project>`. The status is `created`.

### AC2 Status machine

| From | Command | To |
| --- | --- | --- |
| created | assign | assigned |
| assigned | assign | assigned, with the new acting agent |
| assigned | complete | completed |
| completed | verify | verified |
| completed | verify --fail | assigned |
| created, assigned, or completed | cancel | cancelled |

- Any other pair fails with `invalid transition: <status> -> <command>`. The record does not change. This includes `verify` of a task with the status `verified`.
- `assign --task ID --agent NAME` sets `acting_agent`. This task does not check that an agent record exists.
- `complete` needs `--agent` equal to `acting_agent`. Another name fails with `not the acting agent: <name>`.
- `cancel [--reason TEXT]` stores the reason.
- Without `--agent`, `complete` and `verify` fail with `caller unknown: pass --agent`. T11 adds the lookup from the pane.

### AC3 Verify rules

- `verify --task ID --agent NAME [--fail --reason TEXT]` refuses `NAME` equal to `acting_agent` with `verifier must not be the acting agent`.
- `verification` holds one object with `agent`, `result`, `reason`, and `at`. A new verification, after a new completion, replaces it.
- `--fail` needs `--reason`. `--fail` without it fails with `--fail needs --reason`, and the record does not change.

### Partial state for task create

The rules in this task file are the sole authority for a worker or verifier. The
safe order for `task create` is:

1. Check the slug, read the brief with the 1 MiB bound and the 30 second
   deadline, and parse the dependency list. Check that a brief file is a regular
   file before the open. No lock is necessary.
2. Under one lock, read the project record and each task record.
3. Check each dependency. Check the project. Check the dependency graph for a
   cycle. Write nothing on a failure.
4. Get the next value of the counter `task`. Effect: counter `updated`.
5. Write `.hive/state/briefs/<task-id>.md` with an atomic write. Effect: brief
   `created`.
6. Write `.hive/state/tasks/<task-id>.json`. Effect: task `created`. This write
   is the commit marker.

Incomplete states:

| Last completed write | State | Rule |
| --- | --- | --- |
| None | No change | Report `failed`. No effect. |
| Counter | The ID is gone. No brief and no task exist. | Report `partial`, phase `brief`. The next create gets the next number. |
| Counter and brief | An orphan brief exists. No task exists. | Report `partial`, phase `task`. A retry gets a new ID and replaces no task. |
| Counter, brief, and task | Complete | Report `ok`. |

- A command with a partial outcome returns a nil Go error. It builds an
  `output.Envelope` with outcome `partial`, the actual failed `phase`, each
  completed write in `effects`, and one plain error string. `cmd/common.Run`
  replaces a non-nil error with outcome `failed`.
- The command does not remove the orphan brief. It does not claim a rollback.
- `task list` reads only task records. It ignores an orphan brief. `task show` of
  a missing ID fails with `task not found`.

### AC4 Queries

- `task list [--project ID] [--status S] [--ready]` prints one line for each task: ID, status, acting agent, title. The order is by project, then by number.
- `--ready` keeps each task with the status `created` whose dependencies all hold the status `verified`. A task without dependencies is ready.
- `task show ID` prints each field and the brief path. `task show ID --brief` prints the brief text only. With `--json`, `result` holds the record and a `brief` field with the text.
- An unknown task fails with `task not found: <id>`.

### AC5 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | `task create` makes the ID, checks the slug, and bounds the brief read and its deadline. It stores the brief, checks the dependencies, rejects a cycle, and reports each partial step. | V1, V3 |
| AC2 | Each row of the status table holds, and each other pair fails without a change. | V2, V3 |
| AC3 | A verification by the acting agent fails, and a new verification replaces the old one. | V2, V3 |
| AC4 | `list`, `--ready`, `show`, and `show --brief` print what the requirements give. | V3 |
| AC5 | The gate passes. | V4 |

## Verification items

### V1 Creation tests

- Proves: AC1.
- Command: `go test ./internal/task/ -run 'TestCreate|TestSlug|TestCycle' -v`, red and green.
- Expected: one case each for the ID format, the counter across two projects, the slug rules, and the brief file. The slug cases cover a generated slug, `--slug a-b`, `--slug Bad/Slug`, `--slug -a`, and a 41-character slug. An invalid slug fails and leaves the counter unchanged. An actual task ID with a 40-character slug works. One case for the missing dependency error and one for the cycle error.
- Expected: a brief pipe that sends no end of file fails at the 30 second deadline, or at the shorter test deadline. The counter and the files do not change. A brief path that is a static symbol link or a directory fails before the open.
- Expected: two partial cases. A directory at the brief path makes the counter increment and the brief write fail with phase `brief`. A directory at the task path makes the brief write succeed and the task write fail with phase `task`. Each run shows a `partial` envelope and each completed effect. The counter gap and the orphan brief stay. No task record exists.

### V2 Transition tests

- Proves: AC2, AC3.
- Command: `go test ./internal/task/ -run 'TestTransition|TestVerify' -v`, red and green.
- Expected: a table test covers each pair of the five statuses and the five commands `assign`, `complete`, `verify`, `verify --fail`, and `cancel`. The six rows of the status table move the record. Each of the other 19 pairs fails with `invalid transition`, and the record bytes before and after are equal. Three more cases: `complete` by another agent, `complete` without `--agent`, and `verify --fail` without `--reason`. Each fails and leaves the record unchanged. The verify tests cover the self-verification refusal and the replacement after a failed verification and a new completion.

### V3 Command transcript

- Proves: AC1, AC2, AC3, AC4.
- Setup: `go build .` and make a unique temporary home. Create its marker `.hive/README.md` explicitly. Make a unique temporary project in it. Export `HIVE_TMP=<absolute fixture path>` and pass `--home "$HIVE_TMP"` to each command.
- Command:

```sh
git init -q "$HIVE_TMP/repo-one"
./hive --home "$HIVE_TMP" project register "$HIVE_TMP/repo-one"
./hive --home "$HIVE_TMP" task create --project repo-one --title "First step"
./hive --home "$HIVE_TMP" task create --project repo-one --title "Second step" --depends-on 1-first-step
./hive --home "$HIVE_TMP" task create --project repo-one --title "Bad" --slug "Bad/Slug"; echo "exit $?"
./hive --home "$HIVE_TMP" task list --project repo-one --ready
./hive --home "$HIVE_TMP" task assign --task 1-first-step --agent worker-repo-one-1
./hive --home "$HIVE_TMP" task complete --task 1-first-step --agent worker-repo-one-1
./hive --home "$HIVE_TMP" task verify --task 1-first-step --agent worker-repo-one-1; echo "exit $?"
./hive --home "$HIVE_TMP" task verify --task 1-first-step --agent verifier-repo-one-2 --fail --reason "V3 output differs"
./hive --home "$HIVE_TMP" task complete --task 1-first-step --agent worker-repo-one-1
./hive --home "$HIVE_TMP" task verify --task 1-first-step --agent verifier-repo-one-2
./hive --home "$HIVE_TMP" task verify --task 1-first-step --agent verifier-repo-one-2; echo "exit $?"
./hive --home "$HIVE_TMP" task list --project repo-one --ready
./hive --home "$HIVE_TMP" task show 2-second-step --brief
./hive --home "$HIVE_TMP" --json task show 1-first-step
./hive --home "$HIVE_TMP" --json task cancel --task 1-first-step; echo "exit $?"
```

- Expected: the IDs `1-first-step` and `2-second-step`. The bad slug fails with `invalid slug: Bad/Slug` and exit 1. The first ready list shows only `1-first-step`. The self-verification fails with exit 1. The failed verification prints the status `assigned`. The second complete prints `completed`, and the next verify prints `verified`. The verify of the verified task fails with `invalid transition: verified -> verify` and exit 1. The second ready list shows only `2-second-step`. The brief prints `Second step`. The JSON show has `verification` with `result` `pass` and the agent `verifier-repo-one-2`. The cancel of a verified task fails with `invalid transition: verified -> cancel` and exit 1.
- Cleanup: remove only `$HIVE_TMP`. Do not remove the worktree `.hive/state`, `.hive/logs`, `.hive/tmp`, a report, or a log.

### V4 Gate

- Proves: AC5.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- The caller lookup from `HERDR_PANE_ID`, the finished mark, and the inbox records. T11 adds them. T11 also makes `--agent` name a registered agent. In this task `--agent` is a free name.
- The worktree fields. T09 fills them. The delivery fields. T12 fills them.

## Changes

None.
