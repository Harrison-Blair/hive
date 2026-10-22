# Kept first prompt

| Item | Value |
| --- | --- |
| Feature | `general-agent`, wave 5 (Spawn and report) |
| Tier | strong |
| Depends on | 18-general-spawn. Base: `dev` after they land |
| Grant | `grants.md 2026-10-18 hive delivery on dev` |
| Evidence | `.hive/spec/general-agent/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Decisions 15 to 17: when a trust dialog shows, the record keeps the first
prompt. The supervisor sends it one time when the agent first becomes idle.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/supervise/` and its tests | AC1-AC3. |
| `internal/agent/` and its tests | AC1. Only the blocked status and the parent nudge. |
| `internal/lib/records/` and its tests | AC2. Only `first_prompt_attempt_at` and `first_prompt_sent_at`. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/general-agent/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### AC1 Blocked (decision 15)

- With a trust dialog, the record stays live with the agent status blocked.
  The parent gets a blocked nudge. Hive never answers the dialog.

### AC2 Send one time (decision 16)

- Implement the four steps of decision 16 under the pane lock, and its
  failure rules.
- First failing test: a record with a kept first prompt and an idle agent
  status gets the prompt one time across two supervisor passes.

### AC3 Order and tests (decision 17)

- Implement decision 17. Write each test that it names: a supervisor restart
  between steps 2 and 4, an uncertain send, and a failed write.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | A trust dialog gives a live blocked record and a parent nudge | V-red |
| AC2 | The supervisor sends the kept prompt at most one time | V-red, V1 |
| AC3 | Kept prompts go first. The three failure tests exist | V-red |

## Verification items

### V1 Live check

- Proves: AC2.
- Command: in a temp home, spawn a general agent with
  `--allow-trust-dialog` in a new directory for a harness that shows a
  dialog. Answer the dialog by hand. Run the supervisor of the temp home.
- Expected: the prompt arrives one time after the dialog closes. The record
  has both times. Stop the agent and the supervisor. Close the workspace.
  When no installed harness shows a dialog, record this and use the fake
  herdr server test instead.
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
- Exception: a guard test that checks that hive sends nothing can pass on
  the base, because the base sends no kept first prompt. Examples are
  `TestKeptFirstPromptWaitsWhileBlocked` and
  `TestKeptFirstPromptFailedAttemptWriteDoesNotSend`. To prove a guard test,
  remove its guard on purpose. The test must fail.

### V-gate Diff scope and gate

- Proves: each criterion.
- Command: `git diff dev...HEAD --stat`, after you merge `dev` into the task
  branch. Then the full gate of `AGENTS.md` from the worktree root.
- Expected: only the files of the source table and the evidence changed.
  `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0.
  Put the full gate output in the evidence.

## Changes

- 2026-10-21: after a verify fail (t19-verifier-1), V-red has an exception
  for guard tests.
