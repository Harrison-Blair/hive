# T03 task lifecycle

| Item | Value |
| --- | --- |
| Task ID | T03-task-lifecycle |
| Wave | 2 |
| Tier | decent |
| Depends on | T01, CP1 |
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

## Requirements

### AC1 Task creation

- `hive task create --project ID --title TEXT [--slug SLUG] [--depends-on ID,ID] [--brief FILE]` needs a registered project.
- The task ID is `<n>-<slug>`. `n` comes from the counter `task`. The home has one task counter, so a task ID is unique across each project of the home. No command needs a project to select a task.
- The slug is the `--slug` value, else the title in lowercase with each run of characters outside `[a-z0-9]` as one `-`, trimmed of `-` at both ends. Both forms must match `^[a-z0-9]+(-[a-z0-9]+)*$` and have at most 40 characters. Otherwise the command fails with `invalid slug: <value>` before it takes a counter value.
- The brief is the content of `--brief FILE`, or of stdin when FILE is `-`, or the title. Hive reads at most 1 MiB and fails above it. It writes the brief to `.hive/state/briefs/<task-id>.md` atomically.
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
| AC1 | `task create` makes the ID, checks the slug, stores the brief, checks the dependencies, and rejects a cycle. | V1, V3 |
| AC2 | Each row of the status table holds, and each other pair fails without a change. | V2, V3 |
| AC3 | A verification by the acting agent fails, and a new verification replaces the old one. | V2, V3 |
| AC4 | `list`, `--ready`, `show`, and `show --brief` print what the requirements give. | V3 |
| AC5 | The gate passes. | V4 |

## Verification items

### V1 Creation tests

- Proves: AC1.
- Command: `go test ./internal/task/ -run 'TestCreate|TestSlug|TestCycle' -v`, red and green.
- Expected: one case each for the ID format, the counter across two projects, the slug rules, and the brief file. The slug cases cover a generated slug, `--slug a-b`, `--slug Bad/Slug`, `--slug -a`, and a 41-character slug. An invalid slug fails and leaves the counter unchanged. One case for the missing dependency error and one for the cycle error.

### V2 Transition tests

- Proves: AC2, AC3.
- Command: `go test ./internal/task/ -run 'TestTransition|TestVerify' -v`, red and green.
- Expected: a table test covers each pair of the five statuses and the five commands `assign`, `complete`, `verify`, `verify --fail`, and `cancel`. The six rows of the status table move the record. Each of the other 19 pairs fails with `invalid transition`, and the record bytes before and after are equal. Three more cases: `complete` by another agent, `complete` without `--agent`, and `verify --fail` without `--reason`. Each fails and leaves the record unchanged. The verify tests cover the self-verification refusal and the replacement after a failed verification and a new completion.

### V3 Command transcript

- Proves: AC1, AC2, AC3, AC4.
- Setup: `go build . && rm -rf .hive/state projects .hive/tmp/repo-one`
- Command:

```sh
git init -q .hive/tmp/repo-one
./hive --home "$PWD" project register .hive/tmp/repo-one
./hive --home "$PWD" task create --project repo-one --title "First step"
./hive --home "$PWD" task create --project repo-one --title "Second step" --depends-on 1-first-step
./hive --home "$PWD" task create --project repo-one --title "Bad" --slug "Bad/Slug"; echo "exit $?"
./hive --home "$PWD" task list --project repo-one --ready
./hive --home "$PWD" task assign --task 1-first-step --agent worker-repo-one-1
./hive --home "$PWD" task complete --task 1-first-step --agent worker-repo-one-1
./hive --home "$PWD" task verify --task 1-first-step --agent worker-repo-one-1; echo "exit $?"
./hive --home "$PWD" task verify --task 1-first-step --agent verifier-repo-one-2 --fail --reason "V3 output differs"
./hive --home "$PWD" task complete --task 1-first-step --agent worker-repo-one-1
./hive --home "$PWD" task verify --task 1-first-step --agent verifier-repo-one-2
./hive --home "$PWD" task verify --task 1-first-step --agent verifier-repo-one-2; echo "exit $?"
./hive --home "$PWD" task list --project repo-one --ready
./hive --home "$PWD" task show 2-second-step --brief
./hive --home "$PWD" --json task show 1-first-step
./hive --home "$PWD" --json task cancel --task 1-first-step; echo "exit $?"
```

- Expected: the IDs `1-first-step` and `2-second-step`. The bad slug fails with `invalid slug: Bad/Slug` and exit 1. The first ready list shows only `1-first-step`. The self-verification fails with exit 1. The failed verification prints the status `assigned`. The second complete prints `completed`, and the next verify prints `verified`. The verify of the verified task fails with `invalid transition: verified -> verify` and exit 1. The second ready list shows only `2-second-step`. The brief prints `Second step`. The JSON show has `verification` with `result` `pass` and the agent `verifier-repo-one-2`. The cancel of a verified task fails with `invalid transition: verified -> cancel` and exit 1.
- Cleanup: `rm -rf .hive/state projects .hive/tmp/repo-one`

### V4 Gate

- Proves: AC5.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- The caller lookup from `HERDR_PANE_ID`, the finished mark, and the inbox records. T11 adds them. T11 also makes `--agent` name a registered agent. In this task `--agent` is a free name.
- The worktree fields. T09 fills them. The delivery fields. T12 fills them.

## Changes

None.
