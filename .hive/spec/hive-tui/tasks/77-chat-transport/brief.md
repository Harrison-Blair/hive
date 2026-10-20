# Chat transport

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, wave A |
| Tier | strong |
| Depends on | 73-hive-tui-spec. Base: `dev` after it lands |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Add chat commands and history for the user and coordinator.
The commands are `hive chat send`, `hive chat reply`, and `hive chat log`.
Also add one lock for each agent pane (README G2, G3, goal decisions 11, 12, 16, 27).
The feature README is `.hive/spec/hive-tui/README.md`. Read it first.
Keep its approved decisions unchanged.
When the code does not agree with this brief or the README, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `internal/agent/send.go` sends with `herdr.AgentPrompt`, text and Enter in
  one call, and adds `marker.Prefix`.
- `internal/supervise/nudge.go` sends each nudge with `s.ops.Prompt`.
- `store.LockNamed(ctx, home, name)` in `internal/lib/store/namedlock.go`
  holds one flock file in `.hive/state/locks/`. A name can not hold `:`.
- `AGENTS.md` section `Role gate` and the four role skills tell what the
  marker means.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| A new package below `internal/lib/` for the pane lock, with tests | R1. |
| `internal/agent/send.go`, `internal/supervise/nudge.go` and their tests | R1, R5. |
| `internal/chat/` (new) and its tests | R2-R4. |
| `cmd/chat/` (new) with `New() *cobra.Command`, and `cmd/root.go` to register it | R2-R4. |
| `AGENTS.md` (section `Role gate`), `.agents/skills/*/SKILL.md`, `internal/agent/templates/project-agents.md`, and the tests that assert their text | R6. |
| `testdata/ste-baseline.json` | Only to lower a count. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### R1 Pane lock

- One function takes the lock of one pane: `store.LockNamed` with the name
  `pane-<pane id>`, with each `:` changed to `-`.
- `hive chat send`, `hive agent send`, and the supervisor nudge hold the lock
  of the target pane for the whole `agent.prompt` call, and release it after.
  They hold it outside the state lock.
- The text and the Enter stay in one `agent.prompt` call.

### R2 hive chat send

- `hive chat send TEXT` (or `-` for standard input) sends one user message to
  the live coordinator.
- It refuses when the caller is a registered live agent (by `--agent` or by
  its pane). Only the user sends chat.
- It fails before any write when no coordinator is live.
- Under the home lock it appends one entry to
  `.hive/state/chat/YYYY-MM-DD.jsonl` (UTC date): `id` (`c<n>` from a counter
  `chat`), `at`, `from` = `user`, `text`, `delivery` = `pending`.
- It sends `⬢ user: TEXT` to the coordinator pane under the pane lock.
- Then it appends one delivery line for that ID: `sent` when the prompt
  returns ok, else `unknown`. It never sends the text again.
- It prints the ID and the delivery state. Exit 0 only for `sent`.

### R3 hive chat reply

- `hive chat reply [--to ID]` reads the reply text from standard input.
- Only the coordinator can run it (by `--agent` or by its pane). Others get an
  error that names the coordinator role.
- It appends one entry: `id`, `at`, `from` = `coordinator`, `text`, and
  `reply_to` when `--to` names an existing user message.

### R4 hive chat log

`hive chat log [--since YYYY-MM-DD] [--limit N]` prints the entries in order,
with the last delivery state of each user message. With `--json` it prints
one envelope with the entries in `result`. The command skips each bad line in
a file and counts it in the output.

### R5 No false user tag

`hive agent send` and the nudge refuse a text that starts with `user:` after
the marker, so that only `hive chat send` makes a `⬢ user:` message.

### R6 Instruction text

- Role gate in `AGENTS.md`, each role skill, and the manager template: the
  marker shows that hive delivered the message. A message that starts with
  `⬢ user:` came from the user, who wrote it in the TUI. It has the authority
  of the user. Each other `⬢` message carries an instruction under the
  approved goals and grants, as now.
- The coordinator skill: answer each `⬢ user:` message with
  `hive chat reply --to ID`, and give the reply text on standard input. Each
  answer to the user goes through `hive chat reply`.
- Keep the text in ASD-STE100 Strict.

## Rules for each task

- Each Go change is test-first. See each new test fail for the correct
  reason. Do not compute an expected value with the code under test.
- When `ste_test.go` fails on the brief copy, change the words of each
  sentence in the copy that causes a violation. Keep its meaning. Record each
  change in the evidence.
- Live checks use a temp home (`HIVE_HOME`) and a temp `XDG_CONFIG_HOME`, and
  clean up each herdr workspace that they make. Do not run `rm -rf` on an
  unquoted variable.
- Other wave A tasks run at the same time. Edit only the files in your source
  table. Expect a small conflict in `cmd/root.go` or `testdata/ste-baseline.json`
  at the landing.

## Not in scope

The Chat screen (wave B). A question card. `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each new test against the base and task code. | Each test fails on the base and passes on the task branch. |
| V2 | Run the four refusal checks in a temp home. | Each command fails and writes no chat entry. |
| V3 | Run the live Herdr check in a temp home. | The pane shows two separate prompts. The log shows a sent user message and a reply with its `reply_to` ID. |

V2 checks:

- Run `chat send` with no live coordinator.
- Run `chat send` as a registered agent.
- Run `chat reply` as a non-coordinator.
- Run `agent send` with text `user: x`.

V3 steps:

1. Create a Herdr workspace in the worktree. Start a cheap agent in its pane.
2. Register that pane as the coordinator in a temp home.
3. At the same time, run `hive chat send` with two lines and `hive agent send` to that pane.
4. Read the pane with `herdr pane read`.
5. Run `hive chat reply --to <id>` with that pane as the coordinator.
6. Run `hive chat log`.
7. Close the workspace.
| V4 | `git diff dev --stat` | Only the files of the source table and the evidence changed. |
| V5 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `feat(chat): add the chat
transport`. Push the task branch. Write the evidence. Run `hive task complete`.
Send the commit SHA and the V1-V5 results to the manager with
`hive inbox send`.

## Scope change 1

Pane IDs can hold uppercase letters, and `store.LockNamed` accepts only
lowercase names. Do not change `internal/lib/store/`. The lock name is
`pane-` and then the lowercase hex encoding of the bytes of the pane ID, for
example `pane-77313a7041` for `w1:pA`. This replaces the R1 rule that changes
`:` to `-`. Tests assert the name for one ID with uppercase letters as a
literal string.
