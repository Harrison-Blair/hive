# T04 roles and inbox

| Item | Value |
| --- | --- |
| Task ID | T04-roles-and-inbox |
| Wave | 2 |
| Tier | decent |
| Depends on | T01, CP1 |
| Blocks | T08, T10, T11, T15 |
| Branch | `task/T04-roles-and-inbox` |
| Worktree | `projects/hive/worktrees/T04-roles-and-inbox/` |
| Evidence | `evidence/T04-roles-and-inbox/` |

## Overview

This task gives hive its agent records, the singleton claims with a generation
number, and the durable inbox of each role. It makes no herdr call. An agent
registers itself from inside its pane with the herdr environment variables.
A registration is a hive record only. herdr knows a name only for an agent
that `agent.start` started, so T10 and T15 find a registered agent through its
recorded pane, never through its name. T10 later registers an agent that hive
spawns through the same library call. T15 writes most inbox records. This task
writes the `message` kind.

## User flows

1. The user starts a Claude Code session in the hive session and runs `hive agent register --role coordinator`. Hive records the pane and claims the coordinator role with generation 1.
2. The user opens a second session later and registers again. The claim moves to generation 2. The first record gets the status `stale`, and both records stay.
3. A manager runs `hive inbox send --to coordinator "Task 12 needs a decision on the base branch" --agent manager-hive`. The coordinator later runs `hive inbox --role coordinator` and reads the message once.
4. The TUI asks the library for the unread count of each inbox.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `cmd/root.go` | Register the `agent` and `inbox` commands. |
| `cmd/agent/agent.go` | New. The `agent` parent command. |
| `cmd/agent/register/register.go`, `list/list.go`, `show/show.go` | New. |
| `cmd/inbox/inbox.go` | New. `hive inbox [--role KEY] [--all]`. |
| `cmd/inbox/send/send.go` | New. `hive inbox send --to KEY TEXT [--agent NAME]`. |
| `internal/agent/doc.go` | New. Package summary and file map. |
| `internal/agent/register.go` | New. Role rules, names, record keys, claims. T10 calls this function. |
| `internal/agent/list.go`, `show.go`, `render.go` | New. |
| `internal/inbox/doc.go`, `read.go`, `send.go`, `render.go` | New. The command logic. |
| `internal/lib/inbox/inbox.go` | New. Key check, append, consume, list unread, count. T11 and T15 call it. |
| `internal/lib/records/agent.go` | New. The agent record type. |
| `internal/lib/records/claim.go` | New. The claim record type. |
| `internal/lib/records/inbox.go` | New. The inbox record type. |
| `internal/lib/records/doc.go` | Add the map lines. |

## Requirements

### AC1 Agent register

- `hive agent register --role ROLE [--project ID] [--task ID] [--harness KIND] [--model M] [--name NAME] [--pane ID] [--quiet-after MIN] [--pid N]`.
- The roles: `coordinator`, `manager`, `worker`, `verifier`, `supervisor`. A manager needs `--project`. A worker or a verifier needs `--project` and `--task`. A coordinator or a supervisor takes neither.
- `--project` must name a registered project. Otherwise the command fails with `project not found: <id>`. This task does not check that `--task` names a task record. T11 adds that check.
- The pane is the `--pane` value, else `HERDR_PANE_ID`. With no pane the command fails with `no pane: run inside herdr or pass --pane`.
- With the pane from `HERDR_PANE_ID`, hive records `HERDR_WORKSPACE_ID` and `HERDR_TAB_ID` when they exist. With an explicit `--pane`, hive leaves `workspace_id` and `tab_id` empty, because the inherited values can name another pane. T10 fills them from the herdr reply.
- The name is the `--name` value, else by role: `coordinator`, `manager-<project>`, `worker-<project>-<n>`, `verifier-<project>-<n>`, `supervisor`. `n` comes from the counter `agent`, one for the home. A name must match `[a-z][a-z0-9_-]{0,31}`, the herdr rule. T01 bounds a project ID to 18 characters, so each default name with `n` below 10000 fits.
- The record key is the name for a worker or a verifier. For a singleton role it is `<name>-g<generation>`, so a stale registration and the live one both exist. The record lives at `agents/<key>.json`.
- A name that a record with the status `live` holds fails with `agent exists: <name>`, except for a singleton role.
- A pane that a record with the status `live` or `unresolved` holds fails with `pane busy: <pane> holds <name>`. The one exception: the new record claims the singleton role of that record. Then the old record becomes `stale`, as AC2 gives.
- The record has `key`, `name`, `role`, `project`, `task`, `harness`, `model`, `pane_id`, `workspace_id`, `tab_id`, `terminal_id`, `cwd`, `status`, `finished`, `quiet_after_min`, `generation`, `started_at`, and `updated_at`. The status is `live`. `finished` is false. `quiet_after_min` defaults to 10.
- `harness` is the `--harness` value. `terminal_id` and `cwd` are the stable terminal ID and the canonical working directory of the pane. This command makes no herdr call, so it leaves them empty. T10 fills them from `pane.get`. T20 compares these fields on a resolve and skips an empty one.
- The function in `internal/agent` takes explicit values, not the environment. The command reads the environment and passes the values. The checks and the writes of one registration run under one store lock.

### AC2 Claims

- A singleton role is `coordinator`, `manager` for one project, or `supervisor`. Its claim lives at `claims/<key>.json`. The key is `coordinator`, `manager-<project>`, or `supervisor`.
- The claim has `key`, `agent`, `generation`, `pid`, and `claimed_at`. `agent` is the record key of the live holder. A register of a singleton role increments the generation under the store lock, writes the claim, and stores the generation on the agent record.
- The previous holder of the claim gets the status `stale`. Its record stays under its own key.
- `--pid N` stores the process ID of the holder. The supervisor passes it in T15. Without it the value is 0.

### AC3 Agent list and show

- `hive agent list [--project ID] [--role ROLE] [--all]` prints one line for each record: name, generation, role, status, project, task, pane. Without `--all` it hides `stopped`, `exited`, and `stale` records.
- The status values of a record: `live`, `stale`, `stopped`, `exited`, `unresolved`. Later tasks set the last three.
- `hive agent show NAME` prints each field. A singleton name resolves to the record that its claim names. A record key such as `coordinator-g1` also resolves. An unknown name fails with `agent not found: <name>`.

### AC4 Inbox

- An inbox key must match `^[a-z][a-z0-9_-]{0,40}$`. `--role KEY` and `--to KEY` with another value fail with `invalid inbox key: <key>` before any file operation.
- An inbox record lives at `inbox/<key>/<seq>.json`. `seq` comes from the counter `inbox:<key>`. The record has `seq`, `role`, `kind`, `task`, `agent`, `text`, `created_at`, `read_at`, and `nudged_at`.
- The kinds: `completed`, `verified`, `blocked`, `exited`, `quiet`, `unresolved`, `message`. This task writes only `message`.
- `hive inbox --role KEY` consumes the unread records: it selects them and sets `read_at` on each under one store lock, then prints them oldest first. Two parallel consumers never print the same record. `--all` prints each record and sets nothing. Without `--role` the command fails with `pass --role`. T11 adds the default from the caller.
- `hive inbox send --to KEY TEXT [--agent NAME]` appends a `message` record. TEXT has at most 4096 bytes. `--agent` fills `agent`. T11 adds the default from the caller.
- The library gives `Append`, `Consume(key)`, `Unread(key)`, `UnreadCount(key)`, and `UnreadCountAll()`. `Unread` and the counts read without a lock. Each write runs under the store lock.

### AC5 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | `agent register` follows the role rules, the name rules, the pane rules, and the project check, and writes the record under its key. | V1, V3 |
| AC2 | A singleton register increments the generation, marks the previous holder stale, and keeps both records. | V1, V3 |
| AC3 | `agent list` and `agent show` print the records as the requirements give. | V3 |
| AC4 | The inbox checks keys, appends, consumes once, lists unread, and counts, under the lock. | V2, V3 |
| AC5 | The gate passes. | V4 |

## Verification items

### V1 Register and claim tests

- Proves: AC1, AC2.
- Command: `go test -race ./internal/agent/ -v`, red and green.
- Expected: cases for each role rule, an unregistered project, the name format, and a duplicate live name. One case registers a second live role on the same pane and fails. Two coordinator registers on different panes get generations 1 and 2, the first becomes stale, and both records exist under `coordinator-g1` and `coordinator-g2`. A third coordinator register on the pane of the second succeeds and marks the second stale. Twenty parallel worker registers get twenty different names.

### V2 Inbox library tests

- Proves: AC4.
- Command: `go test -race ./internal/lib/inbox/ -v`, red and green.
- Expected: append gives sequence 1, 2, 3. `Unread` returns the three records, `Consume` marks them and removes them from `Unread`, and the counts match. Parallel appends from several goroutines lose no record. Two parallel consumers of 100 records get disjoint sequence numbers, and together they get each record once. The keys `../x`, `A`, and an empty string fail.

### V3 Command transcript

- Proves: AC1, AC2, AC3, AC4.
- Setup: `go build . && rm -rf .hive/state projects`
- Command:

```sh
./hive --home "$PWD" agent register --role coordinator --pane w1:p1
./hive --home "$PWD" agent register --role coordinator --pane w1:p2
./hive --home "$PWD" agent register --role supervisor --pane w1:p2; echo "exit $?"
./hive --home "$PWD" agent register --role worker --pane w2:p1; echo "exit $?"
./hive --home "$PWD" agent register --role manager --project nope --pane w3:p1; echo "exit $?"
./hive --home "$PWD" agent list --all
./hive --home "$PWD" agent show coordinator
./hive --home "$PWD" inbox send --to coordinator "hello" --agent worker-x
./hive --home "$PWD" inbox --role coordinator
./hive --home "$PWD" inbox --role coordinator
./hive --home "$PWD" --json inbox --role coordinator --all
./hive --home "$PWD" inbox --role ../x; echo "exit $?"
```

- Expected: two coordinator records in the list, generation 1 `stale` and generation 2 `live`. The supervisor register fails with `pane busy: w1:p2 holds coordinator` and exit 1. The worker register fails because `--project` and `--task` are absent, with exit 1. The manager register fails with `project not found: nope` and exit 1. `agent show coordinator` prints the record with generation 2 and `pane_id` `w1:p2`. The first `inbox` prints the message, the second prints nothing, and the JSON shows one record with `read_at` set. The last command fails with `invalid inbox key: ../x` and exit 1.
- Cleanup: `rm -rf .hive/state projects`

### V4 Gate

- Proves: AC5.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- Any herdr call. The caller lookup from the pane and the inbox default role. T11 adds them.
- The check that `--task` names a task record of the project. T11 adds it.
- The records of the kinds `completed`, `verified`, `blocked`, `exited`, `quiet`, and `unresolved`. T11 and T15 write them.

## Changes

None.
