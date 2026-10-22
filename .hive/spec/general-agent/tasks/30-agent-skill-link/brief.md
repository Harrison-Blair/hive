# Link of the hive-agent skill

| Item | Value |
| --- | --- |
| Feature | `general-agent`, wave 5 (Spawn and report) |
| Tier | small |
| Depends on | 20-agent-report. Base: `dev` after it lands |
| Grant | `grants.md 2026-10-18 hive delivery on dev` |
| Evidence | `.hive/spec/general-agent/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Each hive skill has a link in `.claude/skills/` and a `!` rule in
`.gitignore`, so that Claude Code finds it. Task `20-agent-report` added
the skill `.agents/skills/hive-agent/` without its link. This task adds the
link.

When the code does not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.claude/skills/hive-agent` | AC1. A new relative link. |
| `.gitignore` | AC1. One `!` rule, beside the rules of the other skills. |
| `gitignore_test.go`, `skills_test.go` | AC2. Only to add the new path, if a test lists the skill links. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/general-agent/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### AC1 The link

- `.claude/skills/hive-agent` is a relative link to
  `../../.agents/skills/hive-agent`, as the links of the other hive skills
  are. Git tracks it.

### AC2 Tests

- When a test lists the tracked skill links or the allowed paths, it includes
  the new link. The test fails when the link is absent.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The link exists, Git tracks it, and it resolves to the skill | V1 |
| AC2 | A test fails without the link | V1 |

## Verification items

### V1 Link

- Proves: AC1, AC2.
- Command: `git ls-files .claude/skills`, `readlink`, and `cat` of
  `.claude/skills/hive-agent/SKILL.md`. Remove the link on purpose and run
  the tests.
- Expected: Git tracks the link and the link resolves. Without the link a test fails,
  or the evidence states that no test lists the links.

### V-gate Diff scope and gate

- Proves: each criterion.
- Command: `git diff dev...HEAD --stat`, after you merge `dev` into the task
  branch. Then the full gate of `AGENTS.md` from the worktree root.
- Expected: only the files of the source table and the evidence changed.
  `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0.
  Put the full gate output in the evidence.
## Rules

- Do not run bare `hive` in herdr without `--no-startup` or
  `HIVE_NO_STARTUP=1`.

## Out of scope

The skill text.

## Changes

None.
