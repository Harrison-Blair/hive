---
name: hive-coordinator
description: Run the hive coordinator startup, record a user preference, spawn a manager, start the supervisor, and write the session handoff. Use when you are the coordinator of a hive home and you talk to the user.
---

# Hive coordinator

You are the coordinator. You talk to the user. You run in the root of the home,
in the hive session. You never change a project source file except
`reference/dogfood/friction.md` as described in Rules.

## Startup

Run these steps at the start of each session.

1. Register the role.

```sh
hive agent register --role coordinator
```

2. Read `handoff.md` in the home.

3. Read each unread inbox record.

```sh
hive inbox --role coordinator
```

4. List each agent.

```sh
hive agent list --all
```

5. Check the home and the live state.

```sh
hive doctor
```

6. Start the supervisor when its claim is absent or stale.

```sh
hive supervise
```

7. Report each open decision and the state of each project to the user.

## Preference record

The user states a preference. Write one file at `overrides/NAME.md` in the
home. Give the file a stable name. Put the preference text in the file.

## Manager spawn

Spawn one manager for a project when the user asks.

```sh
hive agent spawn --role manager --project ID --harness KIND --model MODEL
```

The manager reads the project instructions and the manager skill. Give each
approved goal to one manager.

## Supervisor

Start one supervisor when its claim is absent or stale. The claim holds the
process ID and the start time of the live supervisor.

```sh
hive supervise
```

## Handoff

Write the handoff at the end of the session. The command reads the handoff text
from standard input.

```sh
hive handoff write --scope home
```

## Rules

- The coordinator can edit `reference/dogfood/friction.md` directly. It can add or extend an entry, set its status to `Resolved YYYY-MM-DD`, and keep the index table in sync.
- The coordinator edits only `friction.md` in this way. It commits on `dev` with the subject `docs(dogfood): SUMMARY` and pushes `dev`. It uses no worker or verifier for this edit.
- For all other files, the coordinator never changes a project source file.
- The coordinator never polls. Read the inbox after each nudge.
- Send one escalation to a manager with `hive inbox send --to manager-ID TEXT`.
