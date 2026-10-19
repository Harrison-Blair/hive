# Inbox ownership

| Item | Value |
| --- | --- |
| Wave | wave 2 of 4, task A |
| Tier | decent |
| Depends on | none. Base: `dev` at `1a7a75a6` or later |
| Grant | `grants.md 2026-10-19 inbox ownership task` |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

Today any caller can read the inbox of any role with `hive inbox --role KEY`,
and that read marks the records read. After this task, only the owner of an
inbox marks its records read. Do not change a decision in this brief. When the
code does not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

The friction entry is
`reference/dogfood/friction/entries/any-agent-read-another-role-inbox.md`.
Read it. Do not edit a file in `reference/dogfood/friction/`. The manager
resolves the entry after the landing.

## Known facts

- `cmd/inbox/inbox.go` has the flags `--role`, `--all`, and `--pane` (default
  `HERDR_PANE_ID`). It has no `--agent` flag for a read.
- `internal/inbox/read.go`: `Read` calls `readKey` when the caller sets `--role`.
  `readKey` consumes the records with no caller check. Without `--role`, `Read`
  resolves the caller with `identity.ResolveHeld(locked, options.Agent,
  options.PaneID)` and uses `roleKey`. `roleKey` gives `coordinator` and
  `manager-<project>`. Other roles have no inbox.
- `ReadOptions` has the field `Agent`. The command does not set it.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `cmd/inbox/inbox.go` | R1. Add `--agent` to the read. |
| `internal/inbox/read.go` | R1 to R4. |
| `internal/inbox/*_test.go`, `cmd/inbox_root_test.go` | Tests. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `spec/hive-v1/tasks/<task-id>.md` | R6. New. A copy of this brief. |

Do not edit other files.

## Requirements

### R1 Caller

Each read resolves the caller from `--agent` or from `--pane`, including when
the caller sets `--role`. When hive cannot resolve the caller, the read fails with
`caller unknown: pass --agent or run inside herdr`, and marks nothing.

### R2 Owner read

The owner of an inbox is the agent whose `roleKey` is that key. A read by the
owner, with or without `--role`, consumes the unread records as today.

### R3 Coordinator read of a manager inbox

The coordinator can name a `manager-<project>` inbox with `--role`. That read
prints the unread records and marks nothing. With `--all` it prints all
records, as today.

### R4 Other callers

Each other caller that names an inbox that is not its own gets an error, and
the read marks nothing. The error names the owner, for example
`inbox manager-hive belongs to the manager of project hive`. A worker, a
verifier, and the supervisor have no inbox: a read with no `--role` fails as
today with `role has no inbox: ROLE`. `--all` stays read-only and obeys the
same access rule.

### R5 Tests

Work test-first. Write each test, run it, and make sure that it fails for the
correct reason before you change the code. Cover these cases:

- An owner read marks records read.
- A coordinator read of a manager inbox prints the records and leaves them unread.
- A verifier read of `--role manager-hive` fails, names the owner, and leaves the records unread.
- A read with `--all` by a caller that is not allowed fails.
- An unknown caller fails. Show in the evidence
that each new test fails against the old code, then passes.

### R6 Brief copy

Copy this brief to `spec/hive-v1/tasks/<task-id>.md` with a baseline of 0.
When the copy causes a violation, change the words of that sentence and keep
its meaning. Record each change in the evidence.

## Not in scope

`hive inbox send`, the supervisor, the friction folder, the skills, `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each R5 test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Live check in a scratch home (`hive --home DIR`). Register a coordinator, a manager, and a verifier with fake panes. Send records to the manager inbox. Read as each caller with `--role manager-ID` and with `--all`. | The owner read marks the records read. The coordinator read prints them and marks nothing. The verifier read fails, names the owner, and marks nothing. |
| V3 | `git diff dev --stat` | Only the files of the source table changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `feat(inbox): let only the
owner mark records read`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V4 results to the manager
with `hive inbox send`. Read only your own inbox.
