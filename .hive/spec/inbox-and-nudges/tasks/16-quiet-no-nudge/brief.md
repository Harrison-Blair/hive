# Quiet records do not nudge

| Item | Value |
| --- | --- |
| Wave | wave 2 of 4, task B |
| Tier | small |
| Depends on | none. Base: `dev` at `1a7a75a6` or later |
| Grant | `grants.md 2026-10-19 quiet no-nudge task` |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

Today the supervisor nudges an agent for each unread inbox record, also for a
`quiet` record. Each nudge costs one turn of an idle agent. After this task, an
unread `quiet` record never causes a nudge. Do not change a decision in this
brief. When the code does not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

The friction entry is
`reference/dogfood/friction/entries/idle-session-supervisor-nudges.md`. Read
it. Do not edit a file in `reference/dogfood/friction/`. The manager resolves
the entry after the landing.

## Known facts

- `internal/supervise/quiet.go` writes the `quiet` records.
- `internal/supervise/nudge.go`: `nudgeTodos` returns each unread record that
  has no nudge for the current generation. `nudgeRole` types the nudge line.
- `internal/lib/records/inbox.go` has the kind constant `InboxQuiet`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/supervise/nudge.go` | R1. |
| `internal/supervise/*_test.go` | R2. |
| `testdata/ste-baseline.json` | Only to add the brief copy with 0. |
| `spec/hive-v1/tasks/<task-id>.md` | R3. New. A copy of this brief. |

Do not edit other files.

## Requirements

### R1 Behavior

- The supervisor still writes a `quiet` record, as today.
- An unread `quiet` record never causes a nudge.
- When a record of a different kind causes a nudge, the owner reads all unread
  records, the `quiet` records too, as today.
- Each other kind nudges as today.

### R2 Tests

Work test-first. Write each test, run it, and make sure that it fails for the
correct reason before you change the code. The tests cover at least these
cases:

- An inbox with only unread `quiet` records gets no nudge.
- An inbox with a `quiet` record and a `completed` record gets one nudge.
- The `quiet` record is still written.

Show in the evidence that each new test fails against the old code, then
passes.

### R3 Brief copy

Copy this brief to `spec/hive-v1/tasks/<task-id>.md` with a baseline of 0.
When the copy causes a violation, change the words of that sentence and keep
its meaning. Record each change in the evidence.

## Not in scope

A hush mode, other kinds, the inbox command, the friction folder, `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each R2 test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | `git diff dev --stat` | Only the files of the source table changed. |
| V3 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `fix(supervise): do not nudge
for a quiet record`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V3 results to the manager
with `hive inbox send`. Read only your own inbox.
