# Chat transport

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, wave A |
| Tier | strong |
| Depends on | 73-hive-tui-spec. Base: `dev` after it lands |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Add the chat between the user and the coordinator: `hive chat send`, `hive chat reply`, `hive chat log`, the chat history, and one lock for each agent pane (README G2, G3, goal decisions 11, 12, 16, 27). The feature README is `.hive/spec/hive-tui/README.md`. Read it first. Its
decisions are approved. Do not change a decision. When the code does not agree
with this brief or the README, stop and report it with
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
one envelope with the entries in `result`. A bad line in a file is skipped
and counted in the output.

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
- Other wave A tasks run at the same time. Edit only the files of your source
  table. A conflict in `cmd/root.go` or `testdata/ste-baseline.json` at the
  landing is expected and small.

## Not in scope

The Chat screen (wave B). A question card. `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each new test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | In a temp home: `chat send` with no live coordinator; `chat send` as a registered agent; `chat reply` as a non-coordinator; `agent send` with text `user: x` | Each one fails and writes no chat entry. |
| V3 | Live, in herdr, from a temp home with a coordinator record in a pane that runs a cheap agent: run `hive chat send` with a two-line text and, at the same time, `hive agent send` to the same pane. Then `herdr pane read` the pane, run `hive chat reply --to <id>` as that coordinator, and run `hive chat log` | The pane shows the two prompts as two separate submissions, not mixed. The log shows the user message with `sent` and the reply with its `reply_to`. Clean up each workspace. |
| V4 | `git diff dev --stat` | Only the files of the source table and the evidence changed. |
| V5 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `feat(chat): add the chat
transport`. Push the task branch. Write the evidence. Run `hive task complete`.
Send the commit SHA and the V1-V5 results to the manager with
`hive inbox send`.
