# T04 roles and inbox

| Item | Value |
| --- | --- |
| Task ID | T04-roles-and-inbox |
| Wave | 2 |
| Tier | decent |
| Depends on | T01, T22, CP1 |
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

## Store use

Use these calls:

- `store.WithLock` for each register. The checks and the claim change run under
  one lock.
- `LockedStore.ReadIn`, `WriteIn`, `UpdateIn`, and `ListIn`.
- The collections `["agents"]`, `["claims"]`, and `["inbox", key]`.
- `Next("agent")` and `Next("inbox:<key>")` only through a `LockedStore` handle.

Method matrix: `Store.Read` and `Store.ReadIn` take no lock. A standalone list
(`List`, `ListIn`, `Unread`, and the counts) takes no lock. A related read uses
one `WithLock` callback. Each standalone mutation uses bounded acquisition. A
`LockedStore` method never acquires another lock. Do not read or write a record
file directly.

T04 owns the narrow inbox render-failure path in `cmd/inbox` and
`internal/inbox/render.go`, plus its tests. Keep `cmd/common` and
`internal/lib/output` unchanged. Do not add output work to T22.

## Requirements

### AC1 Agent register

- `hive agent register --role ROLE [--project ID] [--task ID] [--harness KIND] [--model M] [--name NAME] [--pane ID] [--quiet-after MIN] [--pid N]`.
- The roles: `coordinator`, `manager`, `worker`, `verifier`, `supervisor`. A manager needs `--project` and succeeds without a task. A worker or a verifier needs `--project` and `--task`. A coordinator or a supervisor takes neither.
- `--project` must name a registered project. Otherwise the command fails with `project not found: <id>`. This task does not check that `--task` names a task record. T11 adds that check.
- The pane is the `--pane` value, else `HERDR_PANE_ID`. With no pane the command fails with `no pane: run inside herdr or pass --pane`.
- With the pane from `HERDR_PANE_ID`, hive records `HERDR_WORKSPACE_ID` and `HERDR_TAB_ID` when they exist. With an explicit `--pane`, hive leaves `workspace_id` and `tab_id` empty, because the inherited values can name another pane. T10 fills them from the herdr reply.
- The name is the `--name` value, else by role: `coordinator`, `manager-<project>`, `worker-<project>-<n>`, `verifier-<project>-<n>`, `supervisor`. `n` comes from the counter `agent`, one for the home. A name must match `[a-z][a-z0-9_-]{0,31}`, the herdr rule. T01 bounds a project ID to 18 characters, so each default name with `n` below 10000 fits.
- A generated worker or verifier name writes the counter `agent` first and the agent record second. A counter-only failure is `partial`, phase `agent`, with the counter effect known. A gap on retry is safe. An explicit name writes one record and no counter.
- The record key is the name for a worker or a verifier. For a singleton role it is `<name>-g<generation>`, so a stale registration and the live one both exist. The record lives at `agents/<key>.json`.
- The final record key check runs under the same lock as the name and pane checks. A collision with another claim or agent record fails without a write. The singleton exception applies only to replacement of the same singleton claim. An orphan retry needs the same intended identity, claim, and generation. Old generation records stay.
- A name that a record with the status `live` holds fails with `agent exists: <name>`, except for a singleton role.
- A pane that a record with the status `live` or `unresolved` holds fails with `pane busy: <pane> holds <name>`. The one exception: the new record claims the singleton role of that record. Then the old record becomes `stale`, as AC2 gives.
- The record has `key`, `name`, `role`, `project`, `task`, `harness`, `model`, `pane_id`, `workspace_id`, `tab_id`, `terminal_id`, `cwd`, `status`, `finished`, `quiet_after_min`, `generation`, `started_at`, and `updated_at`. The status is `live`. `finished` is false. `quiet_after_min` defaults to 10.
- `harness` is the `--harness` value. `terminal_id` and `cwd` are the stable terminal ID and the canonical working directory of the pane. This command makes no herdr call, so it leaves them empty. T10 fills them from `pane.get`. T20 compares these fields on a resolve and skips an empty one.
- The function in `internal/agent` takes explicit values, not the environment. The command reads the environment and passes the values. The checks and the writes of one registration run under one store lock.

### AC2 Claims

- A singleton role is `coordinator`, `manager` for one project, or `supervisor`. Its claim lives at `claims/<key>.json`. The key is `coordinator`, `manager-<project>`, or `supervisor`.
- The claim has `key`, `agent`, `generation`, `pid`, and `claimed_at`. `agent` is the record key of the live holder. A register of a singleton role increments the generation under the store lock, writes the claim, and stores the generation on the agent record.
- A reader resolves a singleton name through the claim and reads the holder under one lock. A claim read and an unlocked holder read must not span a registration.
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
- The consume loop takes a mark function. It orders the records by numeric sequence. It stops at the first error. A marked record stays marked. The command never marks a record unread and never resends it automatically.
- If every mark succeeds and stdout then fails, a working stderr reports each consumed sequence ID and the command `hive inbox --role KEY --all`. The exit status is 1. Do not promise a valid JSON object after the stdout failure. The command wrapper can capture the writer failure and use the internal result. Keep the business rules in `internal/inbox`.
- `hive inbox send --to KEY TEXT [--agent NAME]` appends a `message` record. TEXT has at most 4096 bytes. `--agent` fills `agent`. T11 adds the default from the caller.
- The library gives `Append`, `Consume(key)`, `Unread(key)`, `UnreadCount(key)`, and `UnreadCountAll()`. `Unread` and the counts read without a lock. Each write runs under the store lock.

### Partial state for register and inbox

The rules in this task file are the sole authority for a worker or verifier.

The safe order for a singleton register is:

1. Check the role, name, pane, project, and environment before the lock.
2. Under one lock, read the claim and each agent record that matters.
3. Check the name rule, the pane rule, and the final record key. Compute the new
   generation from the claim. Write nothing on a failure.
4. Mark the previous holder record `stale` when it exists. Effect: agent `stale`.
5. Write the new agent record with the key `<name>-g<generation>`. Effect: agent
   `created`.
6. Write `claims/<key>.json`. Effect: claim `updated`. This write is the commit
   marker.

Incomplete states:

| Last completed write | State | Rule |
| --- | --- | --- |
| None | No change | Report `failed`. No effect. |
| Old agent stale | The old claim names a stale record. The role has no live holder. | Report `partial`, phase `agent`. A reader must treat the role as vacant. |
| Old agent stale and new agent | The new record is live. The old claim names a stale record. | Report `partial`, phase `claim`. The new record does not hold the role. A retry can replace the same generation record. |
| All three | Complete | Report `ok`. |

- The phase names the write that failed. The unchanged stale claim goes in
  `error` or `result`, not in `effects`.
- A reader resolves a singleton role through the claim. A live agent record
  without the claim does not hold the role.
- A retry with a new pane can replace the orphan record with the same generation
  when its intended identity, claim, and generation agree. It then writes the
  claim. It does not remove the old record.
- A generated name writes the counter before the agent record. A counter-only
  failure is `partial`, phase `agent`, with a counter effect and a safe gap.
  An explicit name has no counter and no partial state.

The safe order for `inbox send` is counter, then record. A failure after the
counter leaves a sequence gap and no record. A gap is safe.

The safe order for `inbox --role` is list, mark each record under one lock, then
print after the lock. A partial consume reports each marked record. The command
does not roll back a mark.

### AC5 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | `agent register` follows the role rules, the name rules, the pane rules, the project check, the generated-name counter order, and the final-key collision rule. It writes the record under its key. | V1, V3 |
| AC2 | A singleton register increments the generation, marks the previous holder stale, and keeps both records. | V1, V3 |
| AC3 | `agent list` and `agent show` print the records as the requirements give. | V3 |
| AC4 | The inbox checks keys, appends, consumes once, lists unread, and counts, under the lock. A failed stdout reports the consumed records and never reverses a mark. | V2, V3 |
| AC5 | The gate passes. | V4 |

## Verification items

### V1 Register and claim tests

- Proves: AC1, AC2.
- Command: `go test -race ./internal/agent/ -v`, red and green.
- Expected: cases for each role rule, an unregistered project, the name format, and a duplicate live name. A manager with a registered project and no task succeeds. A worker or a verifier with no task fails. One case registers a second live role on the same pane and fails. Two coordinator registers on different panes get generations 1 and 2, the first becomes stale, and both records exist under `coordinator-g1` and `coordinator-g2`. A third coordinator register on the pane of the second succeeds and marks the second stale. Twenty parallel worker registers get twenty different names.
- Expected: a cross-claim name collision fails without a write. A worker name that collides with a generation key fails without a write. The singleton exception applies only to the same claim. Old generation records stay.
- Expected: a generated name whose agent write fails gives `partial`, phase `agent`, and a counter effect. A retry permits a gap. An explicit name has no counter.
- Expected: a failed claim write leaves the old record stale and the new record live. The claim is absent or old. The outcome is `partial` with phase `claim` and names each completed effect. The unchanged stale claim is in `error` or `result`, not in `effects`. Skip this case when the test runs as root.

### V2 Inbox library tests

- Proves: AC4.
- Command: `go test -race ./internal/lib/inbox/ -v`, red and green.
- Expected: append gives sequence 1, 2, 3. Inbox IDs `1`, `2`, `10`, and a 20-digit numeric name list and order correctly. `Unread` returns the three records, `Consume` marks them and removes them from `Unread`, and the counts match. Parallel appends from several goroutines lose no record. Two parallel consumers of 100 records get disjoint sequence numbers, and together they get each record once. The keys `../x`, `A`, and an empty string fail.
- Expected: a consume failure after the first mark leaves the first record read and the second unread. The outcome is `partial`.
- Expected: a failed stdout after successful marks leaves each mark in place. The stderr report gives each consumed sequence ID and `hive inbox --role KEY --all`. The exit status is 1. The command does not promise a valid JSON object. The command never marks a record unread and never resends it.
- Expected: a grouped counter test calls `LockedStore.Next`, never a nested `Store.Next`.

### V3 Command transcript

- Proves: AC1, AC2, AC3, AC4.
- Setup: `go build .` and make a unique temporary home. Create its marker `.hive/README.md` explicitly. Export `HIVE_TMP=<absolute fixture path>` and pass `--home "$HIVE_TMP"` to each command.
- Command:

```sh
./hive --home "$HIVE_TMP" agent register --role coordinator --pane w1:p1
./hive --home "$HIVE_TMP" agent register --role coordinator --pane w1:p2
./hive --home "$HIVE_TMP" agent register --role supervisor --pane w1:p2; echo "exit $?"
./hive --home "$HIVE_TMP" agent register --role worker --pane w2:p1; echo "exit $?"
./hive --home "$HIVE_TMP" agent register --role manager --project nope --pane w3:p1; echo "exit $?"
./hive --home "$HIVE_TMP" agent list --all
./hive --home "$HIVE_TMP" agent show coordinator
./hive --home "$HIVE_TMP" inbox send --to coordinator "hello" --agent worker-x
./hive --home "$HIVE_TMP" inbox --role coordinator
./hive --home "$HIVE_TMP" inbox --role coordinator
./hive --home "$HIVE_TMP" --json inbox --role coordinator --all
./hive --home "$HIVE_TMP" inbox --role ../x; echo "exit $?"
```

- Expected: two coordinator records in the list, generation 1 `stale` and generation 2 `live`. The supervisor register fails with `pane busy: w1:p2 holds coordinator` and exit 1. The worker register fails because `--project` and `--task` are absent, with exit 1. The manager register fails with `project not found: nope` and exit 1. `agent show coordinator` prints the record with generation 2 and `pane_id` `w1:p2`. The first `inbox` prints the message, the second prints nothing, and the JSON shows one record with `read_at` set. The last command fails with `invalid inbox key: ../x` and exit 1.
- Cleanup: remove only `$HIVE_TMP`. Do not remove the worktree `.hive/state`, `.hive/logs`, `.hive/tmp`, a report, or a log.

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
