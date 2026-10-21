# Coordinator start owned by hive

| Item | Value |
| --- | --- |
| Feature | `general-agent`, wave 4 (Coordinator start) |
| Tier | strong |
| Depends on | 16-identity-inboxes. Base: `dev` after they land |
| Grant | `grants.md 2026-10-18 hive delivery on dev` |
| Evidence | `.hive/spec/general-agent/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Decisions 30 to 38: only hive code registers the coordinator, the supervisor,
and each manager. Bare `hive` and the restart take a starting claim before
they start the coordinator harness. A root session with no prompt from hive
has no role.

## Known facts

- Task `13-remove-bootstrap` changed the role gate of `AGENTS.md`. Start from
  the text on `dev`.
- Task `2-stale-coordinator-startup` and task `3-tui-restart` made the
  current start and restart code in `internal/tui/` and `internal/agent/`.
  Task `3-tui-restart` put the restart key on the Agents screen. Goal
  `g2-tui-v2` moves the key to Health › Agents (its decision 19). This task
  changes the restart behavior in the shared code. It does not move the key.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/agent/` and its tests | AC1-AC4. |
| `internal/tui/` start and restart files and their tests | AC2-AC4. |
| `internal/lib/` claim and record code and its tests | AC2, AC3. Only the starting claim. |
| `cmd/` agent wiring and its tests | AC1. |
| `AGENTS.md` (role gate), `.agents/skills/hive-coordinator/SKILL.md`, `internal/agent/templates/` | AC5. |
| `skills_test.go`, `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/general-agent/tasks/<task-id>/brief.md` | The brief copy. Commit it. |
| `.hive/spec/general-agent/README.md` | AC4. Only the text of decision 34. |

Do not edit other files.

## Requirements

### AC1 Refused roles (decisions 30, 36, 37)

- `hive agent register` refuses coordinator, supervisor, and manager.
  `hive agent spawn` refuses coordinator and supervisor. The error tells the
  user to run bare `hive` or the restart. The refusal changes no claim and
  no record.
- `hive supervise` and `hive agent spawn --role manager` keep their current
  start and registration.
- First failing test: `hive agent register --role coordinator` from an empty
  pane fails and leaves each claim and record unchanged.
- Tests cover each refused role from an empty pane, from a registered pane,
  from the coordinator pane, and with a `--pane` override (decision 37).

### AC2 Starting claim (decisions 31, 32)

- Implement the starting claim, the record write, and the prompt order of
  decision 31, and the failure rules of decision 32.

### AC3 Dead launcher (decision 33)

- Implement the settle rule of decision 33.

### AC4 Restart (decision 34) and tests (decision 35)

- Implement decision 34 in the shared restart code, with the change that the
  user decided on 2026-10-21. When no claim exists, the restart takes a new
  starting claim under the store lock. The restart also accepts a coordinator
  record that is stale, stopped, or exited. A claim that names another record
  still makes the restart fail with no change.
- Change the text of decision 34 in the spec to state this rule, in
  ASD-STE100 Strict.
- Write one test for each case: no claim, a stale record, a stopped record,
  an exited record, and a claim that names another record.
- Write each test of decision 35.

### AC5 Instruction text (decision 38)

- `AGENTS.md`: remove the rule that a root session with no brief is the
  coordinator. A prompt from hive that names a role wins over the working
  directory. A root session with no prompt from hive has no role and reports
  this to the user. Change the role gate table to match.
- The hive-coordinator skill: remove the register step. Change the startup
  prompt text in the code so that it does not tell the coordinator to
  register.
- Keep the text in ASD-STE100 Strict.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The CLI refuses each hive-owned role with no change | V-red |
| AC2 | Start takes a starting claim and follows decisions 31, 32 | V-red, V1 |
| AC3 | A dead launcher settles as decision 33 states | V-red |
| AC4 | Restart follows decision 34. Each case of decision 35 has a test | V-red, V1 |
| AC5 | The role gate and the skill match decision 38 | V2 |

## Verification items

### V1 Live check

- Proves: AC2, AC4.
- Command: in a temp home and a temp config with a pi coordinator, in a new
  herdr workspace, run the task binary without `--no-startup`. Read the claim
  and the record. Restart the coordinator. Read them again.
- Expected: one coordinator harness runs after each step. The claim names
  the live record. The old record is stale. Close the workspace.

### V2 Instruction text

- Proves: AC5.
- Command: `go test ./ -run 'STE|Skill'` and read each changed line.
- Expected: pass. No text tells a root session that it is the coordinator.

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

- 2026-10-21: after a verify fail (t17-verifier-1), the user chose option A
  for decision 34. AC4 and the source table changed.
