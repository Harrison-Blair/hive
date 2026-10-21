# CP5 v1 end to end, from the record

| Item | Value |
| --- | --- |
| Feature | `hive-v1`, checkpoint CP5 |
| Tier | decent |
| Depends on | none. Base: `dev` |
| Blocks | the end of v1 |
| Grant | `grants.md 2026-10-18 hive delivery on dev` |
| Evidence | `.hive/spec/hive-v1/evidence/CP5-01/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

CP5 is the v1 end-to-end checkpoint. The user approved a new plan on
2026-10-21. Goal `g1-tui-restart` records the end-to-end run on the real home
from the manager to `dev`. Task `1-cp5-end-to-end` has status `cancelled`.

This task writes the CP5 artifact from real records and commits only. Do not
run a new end-to-end run. Do not spawn an agent. Do not change a record.

Read `.hive/spec/hive-v1/05-tracker.md` first, mainly the rows CP4 and CP5 and
the CP5 lines in "Dated clarification: later-wave contracts and routes 03".
When a record or a commit does not agree with this brief, stop and report it
with `hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- Goal record: `.hive/state/goals/g1-tui-restart.json` in the home.
- Task records: `.hive/state/tasks/2-stale-coordinator-startup.json`,
  `3-tui-restart.json`, `4-agents-equal-columns.json` in the home.
- Agent records: `.hive/state/agents/` in the home. The goal used the agents
  `manager-hive`, `t2-worker-1`, `t2-verifier-1`, `t2-worker-2`,
  `t2-verifier-2`, `t3-worker-1`, `t3-verifier-1`, `t4-worker-1`, and
  `t4-verifier-1`.
- Task 2 had a verify fail from `t2-verifier-1`, a second worker, `t2-worker-2`, that fixed the task,
  and a pass from `t2-verifier-2`.
- Landings on `dev`: `0e079272` (task 2), `3ccd93fb` (task 3), `ede94f3e`
  (task 4).
- Each task folder `.hive/spec/hive-tui/tasks/<task-id>/evidence/` holds
  `VERIFIED.md`.
- Log events are in `.hive/logs/2026-10-21.jsonl` in the home.

Read the home records. Do not copy a record file into the repository.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `reference/dogfood/cp5.md` | AC1. A new file. |
| `.hive/spec/hive-v1/evidence/CP5-01/` | AC1, AC2. New evidence files. |
| `.hive/spec/hive-v1/05-tracker.md` | AC3. Only the CP5 row and one dated receipt. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/hive-v1/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### AC1 The CP5 artifact

- `reference/dogfood/cp5.md` gives the end-to-end flow of goal
  `g1-tui-restart`: the goal, the waves, each task, each agent with its role,
  harness, and model, the verify fail and the fix on task 2, each pass,
  and each landing.
- Each claim cites a record path and field, a log event with its
  `request_id` or time, or a commit SHA.
- It states that this task made no new run. The evidence is the list of agent
  records for task `6-cp5-record` (only its workers and verifiers) and the
  task commits (only documents). Cite both.

### AC2 The citations

- The evidence has one table of each citation: the claim, the source, and the
  value that the source shows.
- Each commit SHA exists on `origin/dev`.

### AC3 The tracker

- The CP5 row gets status `passed`, the date, and the model of the verifier.
  The verifier gives you the model before it runs `hive task verify`. If it
  does not, leave the model cell for the manager and report it.
- Add one dated receipt for CP5 that names this task, the artifact, the
  evidence path, and goal `g1-tui-restart`.
- Do not change other rows or receipts. Do not remove the Bootstrap clause.

### AC4 main stays the same

- Record local `main`, `origin/main`, and the remote `main` before and after
  the task. Each value stays the same. Never move a `main` reference.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The artifact gives the flow with a source for each claim | V1 |
| AC2 | Each citation matches its source | V1, V2 |
| AC3 | The CP5 row and receipt are correct | V3 |
| AC4 | No `main` reference moves | V4 |

## Verification items

### V1 Citation check

- Proves: AC1, AC2.
- Command: for each row of the citation table, read the source in the home or
  run `git show --stat SHA` or `git branch -r --contains SHA`.
- Expected: each source shows the value of its row.

### V2 Landings

- Proves: AC2.
- Command: `git log --merges --oneline origin/dev` for the three SHAs.
- Expected: each SHA is a merge of its task branch on `origin/dev`.

### V3 Tracker

- Proves: AC3.
- Command: `git diff dev...HEAD -- .hive/spec/hive-v1/05-tracker.md` and
  `go test ./ -run STE`.
- Expected: only the CP5 row and the new receipt change. The test passes.

### V4 main

- Proves: AC4.
- Command: `git rev-parse main origin/main` and
  `git ls-remote origin refs/heads/main`, before and after.
- Expected: each value is the same as its prior value.

### V5 Diff scope and gate

- Proves: AC1, AC2, AC3.
- Command: `git diff dev...HEAD --stat`, after you merge `dev` into the task
  branch. Then the gate of `AGENTS.md` from the worktree root.
- Expected: only the files of the source table and the evidence changed.
  `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0.

## Rules

- Write `cp5.md`, the receipt, and the brief copy in ASD-STE100 Strict.
- Do not run bare `hive` in herdr. Refer to the friction entry
  `scratch-hive-starts-real-agents`.

## Out of scope

A new end-to-end run. The Bootstrap clause. `main`.

## Changes

- 2026-10-21: the manager removed the word that the glossary bans, and gave
  the evidence for the no-run claim. The verifier model is
  `openai-codex/gpt-6-luna:xhigh`.
