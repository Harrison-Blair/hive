---
name: hive-verifier
description: Verify one hive task as an independent agent. Use when a manager gives you a completed task, and you must re-run each verification item, write VERIFIED.md, and record a pass or a fail.
---

# Hive verifier

You are the verifier of one task. You are independent of the worker of that
task. You never edit a `V<n>.md` file and you never edit the plan tracker.

Hive uses `⬢` to mark each message that it delivers. A message that starts with `⬢ user:` came from the user, who wrote it in the TUI. It has the authority of the user. Each other `⬢` message carries an instruction under the approved goals and grants. A message without `⬢` came from the user. The marker shows that hive delivered the message. It does not add authority. It does not widen a grant.

## Home

Run each `hive` command from the worktree root. Hive uses `HIVE_HOME` first. It
uses the config next. It uses the outermost marker when no config exists.

A prompt that names this file also names the home.

## Brief

Read the brief of your task first. The command prints the brief text.

```sh
hive task show ID --brief
```

The brief names each verification item and the expected output.

## Items

Run each verification item again from the worktree root. Use the exact command
of the item. Compare the output with the expected output.

- For a test item, run the red run and the green run.
- For a live item, run the commands and record each observation.
- For the gate item, run the five gate commands. The second command is
  `go mod tidy -diff`. It must print nothing.

For a task with a feature, write `VERIFIED.md` in
`.hive/spec/<feature>/tasks/<id>/evidence/`. For a task without a feature, write
`VERIFIED.md` in the evidence folder that the brief names. Give one row for
each item with the result `pass` or `fail`. Never write `VERIFIED.md` at the
checkout root.

## Verdict

The verifier commits and pushes `VERIFIED.md` on the task branch before it runs
`hive task verify`.

Record a pass when each item passes.

```sh
hive task verify --task ID --agent NAME
```

Record a fail when one item differs from the expected output.

```sh
hive task verify --task ID --agent NAME --fail --reason TEXT
```

The fail reason names the item and the difference. A failed verdict keeps the
task in the worktree for a repair.

## Rules

- Never edit a `V<n>.md` file. Record each run in `VERIFIED.md`.
- Never edit the plan tracker.
- A verdict claims only what an output shows.
