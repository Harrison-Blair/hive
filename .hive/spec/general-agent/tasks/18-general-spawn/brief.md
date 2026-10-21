# Spawn of a general agent

| Item | Value |
| --- | --- |
| Feature | `general-agent`, wave 5 (Spawn and report) |
| Tier | strong |
| Depends on | 15-harness-trust, 16-identity-inboxes, 17-coordinator-start. Base: `dev` after they land |
| Grant | `grants.md 2026-10-18 hive delivery on dev` |
| Evidence | `.hive/spec/general-agent/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Decisions 1 to 12 and 14: `hive agent spawn --role agent` starts a general
agent with a prompt, in a new tab or a given pane, with a recorded parent.
`hive agent stop` closes only what hive made. The kept first prompt of
decisions 15 to 17 and `hive agent report` are separate tasks after this one.

## Manager decisions

- The general agent spawn code is in `internal/agent/`. `cmd/agent/spawn/`
  keeps the flag wiring.
- The inbox key of a general agent is its name (wave 3).
- The first prompt is exactly these three parts, joined by one blank line:
  1. `⬢ You are the general agent NAME. Your parent is PARENT. This prompt
     sets your role. Do not take a role from your working directory.`
  2. The spawner text.
  3. `Read the skill HOME/.agents/skills/hive-agent/SKILL.md. Send your
     result with hive agent report.`
  NAME, PARENT, and HOME are the values. Record the final text in the
  evidence.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/agent/` and its tests | AC1-AC5. |
| `cmd/agent/` spawn and stop wiring and their tests | AC1, AC3. |
| `internal/lib/records/` and its tests | AC1, AC4. Only the role, the ownership facts, and the kept prompt field. |
| `internal/lib/herdr/` and its tests | AC3, AC4. Only a call that spawn needs. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/general-agent/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### AC1 Role, flags, name, caller (decisions 1 to 4, 7)

- Implement decisions 1 to 4 and 7. Use the caller function of wave 3.
- First failing test: `hive agent spawn --role agent --name x --prompt T`
  from a registered pane creates a record with role `agent` and the caller
  as `parent`.

### AC2 First prompt (decisions 5, 6)

- Build the first prompt as the manager decision states. Check it with the
  rules of `hive agent send` before any tab or process. Send it one time.

### AC3 Pane and working directory (decisions 8 to 10)

- Implement decisions 8 to 10, with the path checks of decision 9.

### AC4 Ownership and cleanup (decisions 11, 12)

- Record the two ownership facts. Change `hive agent stop` as decision 11
  states, also for a pane from `--pane`. Change the help text of `--pane`.
- Implement the failure cleanup of decision 12.

### AC5 Trust (decision 14) and caller tests (decision 25)

- Read the trust result of wave 2 from the harness profile. For "no method",
  fail before any tab, as decision 14 states. Add `--allow-trust-dialog`.
  With it, spawn starts the agent and keeps the first prompt in the record
  for the supervisor task. Do not send it.
- Test spawn with a `--agent` override and with a `--pane` destination that
  differs from the caller pane (decision 25).

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Role, flags, name, and parent follow decisions 1 to 4, 7 | V-red, V1 |
| AC2 | The first prompt follows decisions 5, 6 | V-red, V1 |
| AC3 | Pane and directory follow decisions 8 to 10 | V-red, V1 |
| AC4 | Stop and cleanup close only what hive made | V-red, V1 |
| AC5 | Trust and caller rules hold | V-red |

## Verification items

### V1 Live check

- Proves: AC1-AC4.
- Command: in a temp home in a new herdr workspace, register a scratch
  parent. Spawn a general agent with pi, a cheap model, and a short prompt,
  once with a new tab and once with `--pane` of a split pane. Read the pane
  and the record. Stop each agent.
- Expected: the first prompt shows once with the three parts. The new tab
  closes on stop. The given pane closes on stop, as the help states. The
  parent pane stays. Close the workspace.
## Rules

- Read `.hive/spec/general-agent/README.md` first, then the decisions that
  this brief names. The spec is the authority. This brief does not repeat
  each rule. When the code or the spec does not agree with this brief, stop
  and report it with `hive inbox send --to manager-hive TEXT --agent NAME`.
- Each Go change is test-first. The brief names the first failing test. See
  it fail for the correct reason. Do not mock the behavior under test. Use
  the fake herdr server of the package tests.
- Each new command prints text by default and one JSON object with `--json`,
  and logs its request, its result, and each error (decision 42).
- Run bare `hive` only with `--no-startup` or `HIVE_NO_STARTUP=1`. A live
  check uses a temp `HIVE_HOME` and never changes the real home, its claims,
  or its coordinator. Close each workspace, tab, and pane that the check makes.

## Verification items (common)

### V-red Tests red and green

- Proves: each criterion.
- Command: run each new test against the base code, then against the task
  code. Then break the behavior on purpose once and run the tests again.
- Expected: each new test fails on the base for the correct reason and passes
  on the task branch. The broken behavior fails a test.

### V-gate Diff scope and gate

- Proves: each criterion.
- Command: `git diff dev...HEAD --stat`, after you merge `dev` into the task
  branch. Then the full gate of `AGENTS.md` from the worktree root.
- Expected: only the files of the source table and the evidence changed.
  `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0.
  Put the full gate output in the evidence.

## Changes

None.
