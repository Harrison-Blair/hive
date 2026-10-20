# Spec layout migration

| Item | Value |
| --- | --- |
| Wave | H2, task C of 3 |
| Tier | decent |
| Depends on | 47-brief-copy. Base: `dev` after it lands |
| Grant | `grants.md 2026-10-19 H2 spec layout wave` |
| Evidence | `.hive/spec/home-spec-planning/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Move the old briefs, evidence, and spec docs of the hive project into
`.hive/spec/<feature>/`. The user approved the migration table
`/home/penguin/source/hive/.hive/tmp/reports/h2-migration-table.md` on
2026-10-19 with decisions D1 to D7 and Q1 as proposed. Read the whole table
first. Do the tracked rows only. The manager does the rows in `.hive/state/`
after the landing. Do not change a row. When the project does not agree with the
table, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- Rows with action `move` and a tracked source: use `git mv` so Git keeps the
  history.
- Rows with action `copy` (D1, D7): copy the file from
  `/home/penguin/source/hive/.hive/state/briefs/<id>.md` to the target in the
  worktree. Do not change the state file.
- Rows with a source in `.hive/state/` and action `stay` or `delete`: do not
  touch them. The manager does them.
- The table has no rows for tasks that started after it: 46 and later of the
  hive project (46-task-feature, 47-brief-copy, and this task). They belong to
  the feature `home-spec-planning`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| Each source and target path of a tracked row in the table | R1. |
| `spec/hive-v1/tasks/46-*.md`, `47-*.md`, `evidence/46-*/`, `evidence/47-*/` | R2. |
| Each live file in section "Q3 Links that a move breaks" of the table | R3. |
| `ste_test.go`, `testdata/ste-baseline.json` | R4. |
| `.hive/spec/home-spec-planning/tasks/<task-id>/brief.md` | R5. New. |

Do not edit the text of a moved brief or a moved evidence file. Do not edit
other files.

## Requirements

### R1 Table rows

- Do each tracked `move` and `copy` row of the table. Target paths are as in
  the table.
- After the moves, `spec/hive-v1/` and `evidence/` at the project root have no
  files.

### R2 Tasks after the table

Move the tracked brief copies and the evidence folders of 46 and 47 to
`.hive/spec/home-spec-planning/tasks/<id>/brief.md` and `.../evidence/`, as
the table does for 43 and 44.

### R3 Live links

Change each reference in the Q3 table to the new path, with the decisions D2
to D4: the tracker moves into `hive-v1`, `AGENTS.md` points at the manager and
worker skills for the general task flow and at the tracker only for v1, the
CP5 checkpoint path becomes `.hive/spec/hive-v1/checkpoints/CP5-<pass>/`, and
the two format files go to `.hive/spec/BRIEF-FORMAT.md` and
`.hive/spec/EVIDENCE-FORMAT.md`. Keep each changed instruction text in
ASD-STE100 Strict. Keep history text (past log entries in the tracker, past
briefs, evidence, the friction log) as it is.

### R4 STE check

Change the globs of `steFiles` to the new paths, for example
`.hive/spec/*.md`, `.hive/spec/hive-v1/05-tracker.md`, and
`.hive/spec/*/tasks/*/brief.md`. Rename each key of
`testdata/ste-baseline.json`. Do not raise a count.

### R5 Brief copy

`hive worktree create` put this brief at
`.hive/spec/home-spec-planning/tasks/<task-id>/brief.md` when the record has a
feature. When it is not there, copy it from
`/home/penguin/source/hive/.hive/state/briefs/<task-id>.md`. Commit it. When the copy causes a violation, change the words of that sentence and keep its meaning. Record each change in the evidence.

## Not in scope

The records and briefs in `.hive/state/` (the manager does them). Stenographer
(Q1: it waits). Feature README files other than the moved
`hive-v1/README.md`. `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | A script that reads each tracked row of the table and checks that the source is absent and the target exists in the worktree | Each row passes. The script and its output are in the evidence. |
| V2 | `git log --follow --oneline -- <target>` for three moved briefs and three moved evidence files | Each shows commits from before this task. |
| V3 | `git ls-files spec/hive-v1 evidence` and `git grep -n -e 'spec/hive-v1' -e '](evidence/' -- ':!.hive/spec/*/tasks' ':!reference/dogfood' ':!.hive/spec/hive-v1/05-tracker.md'` | No files. List each grep line in the evidence with its reason for staying. |
| V4 | `git check-ignore -q .hive/spec/hive-v1/README.md; echo $?` and `git diff dev --stat --find-renames` | Exit 1. The diff shows renames for the moves and only the files of the source table. |
| V5 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `refactor(spec): move specs
into .hive/spec`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V5 results to the manager
with `hive inbox send`.
