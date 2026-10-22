# The agent report command and skill

| Item | Value |
| --- | --- |
| Feature | `general-agent`, wave 5 (Spawn and report) |
| Tier | decent |
| Depends on | 18-general-spawn. Base: `dev` after they land |
| Grant | `grants.md 2026-10-18 hive delivery on dev` |
| Evidence | `.hive/spec/general-agent/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Decisions 18 to 23: a general agent sends its result to its parent with
`hive agent report`. The skill `hive-agent` tells it how.

## Manager decisions

- The report code is in `internal/agent/`. The command is
  `cmd/agent/report/`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/agent/` and its tests | AC1-AC3. |
| `cmd/agent/report/`, `cmd/agent/` and their tests | AC1. |
| `internal/inbox/` and its tests | AC2. Only the kind `report` and the path field. |
| `internal/lib/records/` and its tests | AC2, AC3. Only `reported_at`, and in `inbox.go` the kind `report` and the path field. |
| `internal/lib/inbox/` and its tests | AC2. Only the kind `report` and the path field. |
| `.agents/skills/hive-agent/SKILL.md` | AC4. A new file. |
| `skills_test.go`, `testdata/ste-baseline.json`, `ste_test.go` | AC4, and to add the brief copy with 0. |
| `.hive/spec/general-agent/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### AC1 The command (decisions 18, 19, 20)

- Implement decisions 18 to 20. Only a general agent can run it.
- First failing test: a general agent reports text from standard input, and
  the parent inbox gets one record of kind `report` with that text.

### AC2 Caller and routing (decisions 21, 28, 25)

- The caller is the agent of the pane (the caller function of wave 3). The
  inbox is chosen with the rule of decision 28.
- Test the command with a `--agent` override and a `--pane` override
  (decision 25).

### AC3 Record (decisions 21, 22)

- Write the inbox record, then set `reported_at`. Implement the failure rule
  of decision 21. A report does not set `finished`.

### AC4 The skill (decision 23)

- Write `.agents/skills/hive-agent/SKILL.md` in ASD-STE100 Strict. Add it to
  `steFiles` with 0 and to the skill test, as the other skills.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The command reads text or a file within the limits | V-red |
| AC2 | The caller is the pane agent. The report goes to the parent. | V-red |
| AC3 | The report sets `reported_at` and not `finished` | V-red |
| AC4 | The skill exists and passes the ratchet at 0 | V1 |

## Verification items

### V1 Skill

- Proves: AC4.
- Command: `go test ./ -run 'STE|Skill'` and read the skill.
- Expected: pass. The skill states the work, the report, and that the agent
  has no task.
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

- 2026-10-21: the source table allows the inbox kind and the path field in
  `internal/lib/records/inbox.go` and `internal/lib/inbox/`.
