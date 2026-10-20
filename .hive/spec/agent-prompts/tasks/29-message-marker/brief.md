# Hive message marker

| Item | Value |
| --- | --- |
| Wave | message marker |
| Tier | strong |
| Depends on | 28-first-prompt-fix. Base: `dev` at `3717fb22` or later |
| Grant | `grants.md 2026-10-19 hive message marker task` |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

This task does three things. Hive starts each Claude Code and Codex agent that
it spawns with the harness bypass flag. Each message that hive types into an
agent pane starts with a visible marker. The instructions tell each role what
the marker means, and tell a manager to send a record before it stops. The user
approved R1 directly on 2026-10-19. Do not change a decision in this brief.
When the code does not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

The friction entry is
`reference/dogfood/friction/entries/manager-blocked-by-harness-permission.md`.
Read it. Do not edit a file in `reference/dogfood/friction/`. The manager
resolves the entry after the landing.

## Known facts

- Harness profiles: `internal/lib/harness/claude.go` and `codex.go`. `Args`
  builds the native arguments.
- `codex --help` lists `--dangerously-bypass-approvals-and-sandbox`. Claude
  Code has `--dangerously-skip-permissions`. Check each with `--help` before
  you use it.
- Hive types the worker and verifier first prompts with `workerPrompt` and
  `verifierPrompt` in `internal/agent/brief.go`. Hive types the manager first
  prompt in `internal/agent/spawn.go`. The `hive agent send` message is in
  `internal/agent/send.go`. The supervisor nudge line `nudgeLine` is in
  `internal/supervise/nudge.go`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/lib/harness/claude.go`, `codex.go`, and their tests | R1. |
| One package below `internal/lib/` for the marker constant, with a `doc.go` when it needs one | R2. |
| `internal/agent/brief.go`, `spawn.go`, `send.go`, `internal/supervise/nudge.go`, and their tests | R2. |
| `AGENTS.md` (section `Role gate`) | R4. |
| `.agents/skills/hive-coordinator/SKILL.md`, `hive-manager/SKILL.md`, `hive-worker/SKILL.md`, `hive-verifier/SKILL.md` | R4, R5. |
| `internal/agent/templates/project-agents.md`, `internal/agent/manager_prompt_test.go` | R4, R5. |
| `cmd/agent/spawn/cli_test.go` and each other existing test that asserts the exact Claude Code or Codex spawn arguments | R1. Only to add the new flag to the expected arguments. Change no other assertion. List each file in the evidence. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `spec/hive-v1/tasks/<task-id>.md` | R7. New. A copy of this brief. |

Do not edit other files.

## Requirements

### R1 Harness start

- Hive starts each Claude Code agent that it spawns (manager, worker,
  verifier) with `--dangerously-skip-permissions`.
- Hive starts each Codex agent that it spawns with
  `--dangerously-bypass-approvals-and-sandbox`, or the exact flag that
  `codex --help` gives.
- Pi and the other harnesses do not change.
- Put the flags in the harness profiles. Add tests that fail without the
  flags.

### R2 Marker

- Each message that hive types into an agent pane starts with the character
  `⬢` (U+2B22) and one space. This covers the first spawn prompt of each role,
  `hive agent send`, and each supervisor nudge.
- Define the marker one time in Go and use that definition in each place.
- Add tests for each place. Each test must fail without the marker.

### R3 Live check

Type a marker message into a Claude Code pane and into a Pi pane with herdr.
Use scratch agents in a scratch workspace that you create and close. Show that
each harness receives `⬢` intact: read each pane and record the output. Then
stop each scratch agent and close the workspace.

### R4 Marker rule

`AGENTS.md` (`Role gate`), the coordinator, manager, worker, and verifier
skills, and `internal/agent/templates/project-agents.md` say this: a message
that starts with `⬢` came from hive. It carries an instruction from the
coordinator or the manager under the approved goals and grants. A message
without `⬢` came from the user. The marker shows the source. It is not
evidence of extra authority. It does not widen a grant.

### R5 Blocked record

The manager skill and the template say this: before a manager stops for a user
decision, or when a harness blocks an action, it sends the coordinator one
record with `hive inbox send --to coordinator TEXT --agent NAME`. The record
names the decision or the block. Then the manager ends its turn. Add checks of
the R4 and R5 template text to `internal/agent/manager_prompt_test.go`. Run
them, and make sure that they fail before you change the template.

### R6 STE

Write the instruction text in ASD-STE100 Strict. Use each term only with its
meaning in `STE-GLOSSARY.md`. The counts in `testdata/ste-baseline.json` must
not rise.

### R7 Brief copy

Copy this brief to `spec/hive-v1/tasks/<task-id>.md` with a baseline of 0.
When the copy causes a violation, change the words of that sentence and keep
its meaning. Record each change in the evidence.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each R1, R2, and R5 test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | The R3 live check | Both panes show `⬢` intact. The scratch resources are gone. |
| V3 | `grep -rn '⬢' --include=*.go internal` | Go defines the marker character one time in non-test code. |
| V4 | Read the R4 and R5 text in each named file | Present in each. |
| V5 | `go test -run TestSTE .` and `git diff dev --stat` | Pass, no count rose. Only the files of the source table changed. |
| V6 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `feat(agent): mark hive
messages and bypass harness prompts`. Push the task branch. Write the evidence.
Run `hive task complete`. Send the commit SHA and the V1-V6 results to the
manager with `hive inbox send`.
