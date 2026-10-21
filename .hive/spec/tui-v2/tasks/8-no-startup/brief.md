# The no-startup switch

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, wave 2 (Harness) |
| Tier | small |
| Depends on | 7-tui-v2-spec-copy. Base: `dev` after they land |
| Grant | `grants.md 2026-10-21 tui-v2 delivery on dev and PR 47 update` |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Decision 34 of the spec adds the switch `--no-startup` and the variable
`HIVE_NO_STARTUP=1`. With the switch, bare `hive` opens the TUI and starts no
coordinator, no supervisor, and no tab, and does not rename the tab. Each
later task and each verifier uses the switch for a live check. This task
lands first in wave 2.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `cmd/root.go` and its tests | AC1. |
| `internal/tui/startup.go` and its tests | AC1. |
| `.agents/skills/hive-worker/SKILL.md`, `.agents/skills/hive-verifier/SKILL.md`, `skills_test.go` | AC2. |
| `reference/dogfood/friction/` | AC3. Only the entry `scratch-hive-starts-real-agents` and its row. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/tui-v2/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### AC1 The switch

- `hive --no-startup` and `HIVE_NO_STARTUP=1 hive` run the TUI. They make no
  tab, start no coordinator and no supervisor, and do not rename the tab.
- Any other value of `HIVE_NO_STARTUP` has no effect.
- Startup logs the skipped steps with the reason `no-startup`.
- First failing test: bare `hive` with the switch, against the fake herdr
  server, makes no herdr call that creates or renames a tab.

### AC2 The skill lines

- The worker skill and the verifier skill tell each agent to run bare `hive`
  only with `--no-startup` or `HIVE_NO_STARTUP=1`. Keep the text in
  ASD-STE100 Strict.

### AC3 The friction entry

- Extend the entry `scratch-hive-starts-real-agents` with the switch. Do not
  resolve it. The spec says that automatic detection stays out of scope.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The switch and the variable skip each startup step | V1, V2 |
| AC2 | The worker and verifier skills require the switch | V3 |
| AC3 | The friction entry names the switch | V3 |

## Verification items

### V1 Tests

- Proves: AC1.
- Command: run each new test against the base code, then against the task code.
- Expected: each new test fails on the base and passes on the task branch.

### V2 Live check

- Proves: AC1.
- Command: in a new herdr workspace with a temp `HIVE_HOME`, run the task
  binary with `--no-startup`, then with `HIVE_NO_STARTUP=1`. List the tabs
  and the agent records each time.
- Expected: no new tab, no new label, no agent record. Close the workspace.

### V3 Instruction text

- Proves: AC2, AC3.
- Command: `go test ./ -run 'STE|Skill'` and read the changed lines.
- Expected: pass. Each skill names the switch.

### V-gate Diff scope and gate

- Proves: each criterion.
- Command: `git diff dev...HEAD --stat`, after you merge `dev` into the task
  branch. Then the full gate of `AGENTS.md` from the worktree root.
- Expected: only the files of the source table and the evidence changed.
  `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0.
  Put the full gate output in the evidence.

## Rules

- Read `.hive/spec/tui-v2/README.md` first. When the code, a board, or the
  spec does not agree with this brief, stop and report it with
  `hive inbox send --to manager-hive TEXT --agent NAME`.
- Each Go change is test-first. The brief names the first failing test. See
  it fail for the correct reason.
- Do not run bare `hive` in herdr without `--no-startup` or
  `HIVE_NO_STARTUP=1`. Before task `no-startup` lands, do not run bare `hive`
  in herdr. Refer to the friction entry `scratch-hive-starts-real-agents`.
- A live check never changes the real home. Close each workspace, tab, and
  pane that the check makes.

## Out of scope

`--demo` and `--size`. Task `demo-mode` adds them.

## Changes

None.
