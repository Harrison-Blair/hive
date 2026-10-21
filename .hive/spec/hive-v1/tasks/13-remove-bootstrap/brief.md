# Removal of the Bootstrap clause

| Item | Value |
| --- | --- |
| Feature | `hive-v1`, end of bootstrap |
| Tier | small |
| Depends on | none. Base: `dev` |
| Blocks | none |
| Grant | `grants.md 2026-10-18 hive delivery on dev` |
| Evidence | `.hive/spec/hive-v1/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The Bootstrap clause of `AGENTS.md` let a root session develop hive directly
until the hive manager could spawn a worker. The manager now spawns workers.
On 2026-10-21 the user decided to remove the clause now. This task removes
the clause and each reference to it in the instruction text.

When a file does not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `AGENTS.md` line 20: the role-gate bullet about "develop".
- `AGENTS.md` section `### Bootstrap clause` (two paragraphs).
- `AGENTS.md` section `Development`: "These rules apply to a developer and to
  a worker". Section `Git`: "A developer commits on `dev` and pushes `dev`."
- `STE-GLOSSARY.md`: the rows `developer` and `bootstrap clause`.
- `.hive/spec/hive-v1/05-tracker.md` line 254: "Bootstrap clause removal
  needs actual manager-first-worker evidence." Line 2402: the decision at a
  checkpoint.
- `.hive/spec/hive-v1/03-architecture.md` line 107 and `04-risks.md` line 10
  describe the clause as part of the v1 design.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `AGENTS.md` | AC1. |
| `STE-GLOSSARY.md` | AC2. |
| `.hive/spec/hive-v1/05-tracker.md` | AC3. Only one new dated receipt. |
| `.agents/skills/` | AC1. Only a line that points to the clause, if one exists. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/hive-v1/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### AC1 AGENTS.md

- Remove the section `Bootstrap clause` and the role-gate bullet about
  "develop".
- In `Development`, the rules apply to a worker on the hive project. In
  `Git`, remove the sentence about a developer. Keep each other rule.
- Search `.agents/skills/` for "bootstrap" and "develop". Remove each line
  that points to the clause. Record the search and its result.
- Keep the text in ASD-STE100 Strict.

### AC2 Glossary

- Remove the rows `developer` and `bootstrap clause` when no instruction file
  uses the term after AC1. Record the search.

### AC3 Tracker

- Add one dated receipt: on 2026-10-21 the user removed the Bootstrap clause,
  because the hive manager spawns workers. Name this task.
- Do not change the old receipts. They are history.
- Do not change `03-architecture.md` or `04-risks.md`. They are the v1 design
  record. List them in the evidence as history that names the clause.

### AC4 Ratchet

- Each STE baseline count stays at or below its current value.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | No instruction file has the clause or a pointer to it | V1 |
| AC2 | The glossary has no unused term of the clause | V1 |
| AC3 | The tracker records the removal | V2 |
| AC4 | No STE count rises | V3 |

## Verification items

### V1 Search

- Proves: AC1, AC2.
- Command: `git grep -n -i -w -e bootstrap -e develop -e developer -- AGENTS.md STE-GLOSSARY.md .agents/`
- Expected: no line refers to the clause or to the developer role.

### V2 Tracker

- Proves: AC3.
- Command: `git diff dev...HEAD -- .hive/spec/hive-v1/05-tracker.md`
- Expected: the diff adds only the new receipt.

### V3 Ratchet

- Proves: AC4.
- Command: `go test ./ -run 'STE|Skill'` and
  `git diff dev...HEAD -- testdata/ste-baseline.json`
- Expected: pass. No count rises. Only the brief copy row is new.

### V4 Diff scope and gate

- Proves: AC1, AC2, AC3, AC4.
- Command: `git diff dev...HEAD --stat`, after you merge `dev` into the task
  branch. Then the full gate of `AGENTS.md` from the worktree root.
- Expected: only the files of the source table and the evidence changed.
  `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0.

## Rules

- Do not run bare `hive` in herdr. Refer to the friction entry
  `scratch-hive-starts-real-agents`.

## Out of scope

The v1 design files. CP5. `main`.

## Changes

None.
