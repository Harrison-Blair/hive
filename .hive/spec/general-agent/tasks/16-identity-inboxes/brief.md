# Identity and inboxes

| Item | Value |
| --- | --- |
| Feature | `general-agent`, wave 3 (Identity and inboxes) |
| Tier | strong |
| Depends on | 14-general-agent-spec (landed). Base: `dev` after they land |
| Grant | `grants.md 2026-10-18 hive delivery on dev` |
| Evidence | `.hive/spec/general-agent/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Decisions 24 to 29: the caller of a hive command in a herdr pane is the agent
of that pane. Each agent has its own inbox. The supervisor sends the nudges of
a spawned agent to its parent. This wave does not add the role `agent`. Wave
5 adds it and the tests of `hive agent report` and `hive agent spawn` from
decision 25.

## Manager decisions

- The inbox key of a worker, a verifier, and a general agent is the agent
  name. The keys `coordinator` and `manager-<project>` stay (decision 26).

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/inbox/`, `cmd/inbox/` and their tests | AC1-AC3. |
| `internal/agent/` and its tests | AC1. Only the caller rule, in a shared function. |
| `internal/supervise/` and its tests | AC4. |
| `internal/doctor/` and its tests | AC4. Only the unread-record check. |
| `cmd/` other files and their tests | AC1. Only the flag wiring of the caller rule. |
| `.agents/skills/` and `internal/agent/templates/` | AC3. Only the inbox command lines that change. |
| `skills_test.go`, `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/general-agent/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### AC1 Caller (decision 24)

- In a herdr pane, the agent of the pane is the caller. `--agent` that names
  another agent fails. For `hive inbox` and `hive inbox send`, a `--pane`
  that names another pane fails. Outside herdr, `--agent` keeps its current
  behavior.
- Put the rule in one function that wave 5 can call for `hive agent report`
  and `hive agent spawn`. Record its name.
- First failing test: `hive inbox send --agent OTHER` from a registered pane
  fails and writes no record.

### AC2 Inboxes (decisions 26, 27)

- Each agent has an inbox keyed as in the manager decision above.
- `hive inbox` with no flag reads the inbox of the caller. The coordinator
  can read each inbox with `--role KEY`, and that read consumes no record.
  Each other agent reads only its own inbox.
- `hive inbox send --to KEY` accepts each inbox key.

### AC3 Instruction text

- Update each skill or template line that tells an agent to use
  `--role` or `--agent` in a way that AC1 or AC2 now refuses. Keep the text in
  ASD-STE100 Strict. Record each change.

### AC4 Routing (decisions 28, 29)

- The supervisor sends each nudge to the inbox of the recorded `parent`. When
  the parent record is not live or does not exist, the nudge goes to
  `coordinator`. The quiet rule does not change.
- A record stays in an inbox when its owner stops. `hive doctor` lists each
  unread record in an inbox that has no live owner.
- A test proves for a worker, a verifier, and a manager with the usual
  parents that the result is the same as the current rule. A test proves
  that the nudges of a worker that the coordinator spawned go to the
  coordinator.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The pane agent is the caller; overrides fail in herdr | V-red, V1 |
| AC2 | Each agent has its inbox and reads only it, except the coordinator | V-red, V1 |
| AC3 | Instruction text matches the new rules | V2 |
| AC4 | Nudges go to the parent, else to the coordinator | V-red |

## Verification items

### V1 Live check

- Proves: AC1, AC2.
- Command: in a temp home in a new herdr workspace, register two scratch
  agents in two panes. From pane A, run `hive inbox send --agent B` and
  `hive inbox --pane PANE-B`. Then send a record from A to B and read it in B.
- Expected: both overrides fail and change nothing. B reads the record.

### V2 Instruction text

- Proves: AC3.
- Command: `go test ./ -run 'STE|Skill'` and read each changed line.
- Expected: pass. No skill line uses a refused form.

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
