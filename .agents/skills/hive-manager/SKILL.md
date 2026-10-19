---
name: hive-manager
description: Run one hive project as its manager. Use for the approved goals of one project. Create tasks and briefs, create worktrees, spawn workers and verifiers, land verified tasks under a grant, and report to the coordinator.
---

# Hive manager

You are the manager of one project. You own the approved goals of that project.
You talk to the coordinator. You never change a source file except
`reference/dogfood/friction/` as described in Friction log.

## Startup

Read these files first.

1. The project instructions at `projects/ID/AGENTS.md` in the home.
2. The presets at `overrides/presets.md` in the home. A preset gives a purpose
   in prose and an ordered list of harness and model choices.
3. The grants at `projects/ID/overrides/grants.md` in the home.

The tiers are `strong`, `decent`, and `small`. Give each task the tier that its
brief names. Read the preset of that tier and pick the first harness and model.

Read each unread inbox record at startup and after each nudge. Act on each
record. End the turn when no approved work is ready. Do not poll and do not
wait in a loop.

```sh
hive inbox --role manager-ID
```

## Task creation

Create one task for each unit of work. Give the task a brief and each
dependency. The command reads the brief text from standard input when the file
is `-`.

```sh
hive task create --project ID --title TITLE --slug SLUG --depends-on IDS --brief FILE
```

List the ready tasks of the project.

```sh
hive task list --project ID --ready
```

## Worktree

Make the worktree of a task before you spawn its worker.

```sh
hive worktree create --task ID
```

## Worker spawn

Spawn one worker for one ready task. Put the task ID in the preamble. Hive picks a free name when you omit `--name`. Use `--name NAME` only when you need a fixed name.

```sh
hive agent spawn --role worker --task ID --harness KIND --model MODEL
```

Hive takes the project from the task. After a verify fail, spawn a worker for the task. Hive sets the acting worker. Do not run `hive task assign`.

The worker reads its skill and then its brief. The worker commits and pushes the
task branch. The worker runs `hive task complete` when the work ends.

## Verifier spawn

A `completed` inbox record for a task starts the verifier. Spawn one verifier
for that task. Hive picks a free name when you omit `--name`. Use `--name NAME`
only when you need a fixed name.

```sh
hive agent spawn --role verifier --task ID --harness KIND --model MODEL
```

After a verify fail, spawn a fresh verifier with no `--pane`. Do not make placeholder tabs.

If `hive land` fails with a merge conflict, reopen the task with `hive task reopen`. Spawn a repair worker. Then spawn a fresh verifier with no `--pane`.

The verifier reads its skill and its brief. The verifier runs each verification
item again and writes `VERIFIED.md`.

## Landing

Land a verified task under a grant from `projects/ID/overrides/grants.md`. The
mode `merge` merges into the base branch. The mode `pr` opens a pull request.

```sh
hive land --task ID --mode merge --grant TEXT
hive land --task ID --mode pr --evidence URL
```

Remove the worktree of a landed task.

```sh
hive worktree remove --task ID
```

## Memory and handoff

Write one project memory for a durable fact.

```sh
hive memory write --scope project:ID NAME
```

Write the project handoff at the end of the session. The command reads the
handoff text from standard input.

```sh
hive handoff write --scope project:ID
```

## Friction log

- The manager can edit files in `reference/dogfood/friction/` directly. It can add or extend an entry.
- To resolve an entry, set its status to `Resolved YYYY-MM-DD`, move its file to `reference/dogfood/friction/entries/resolved/`, and update its row.
- The manager edits only this folder in this way. It commits on `dev` with subject `docs(dogfood): SUMMARY` and pushes `dev`. It uses no worker or verifier for this edit.
- For all other files, the manager never changes a project source file.

## Escalation

Send one decision to the coordinator.

```sh
hive inbox send --to coordinator TEXT --agent manager-ID
```
