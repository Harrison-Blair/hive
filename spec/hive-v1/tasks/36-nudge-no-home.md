# Nudge without the home flag

| Item | Value |
| --- | --- |
| Wave | nudge text |
| Tier | small |
| Depends on | none. Base: `dev` at `a59eb6d5` or later |
| Grant | `grants.md 2026-10-19 nudge without home flag task` |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

The supervisor nudge line names the home with the `--home` flag. The nudge
must not name the home. Each agent runs below the home, and `home.Resolve`
finds the home from the working directory. When the code does not agree with
this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `nudgeLine` in `internal/supervise/nudge.go` returns a home-specific inbox command.
- `internal/supervise/nudge_test.go` compares the prompt text with the output
  of `nudgeLine`. That test computes the expected value with the code under
  test.
- `home.Resolve` in `internal/lib/home/home.go` finds the marker above the
  working directory when the flag and `HIVE_HOME` are not set.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/supervise/nudge.go` | R1. |
| `internal/supervise/nudge_test.go` and each other test that asserts the nudge text | R2. |
| Each doc, skill, or spec file that quotes the old nudge text. Do not edit files below `evidence/`, `spec/hive-v1/evidence/`, or `reference/dogfood/friction/`. | R3. List each file in the evidence. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `spec/hive-v1/tasks/<task-id>.md` | R4. New. A copy of this brief. |

Do not edit other files.

## Requirements

### R1 Nudge text

- The nudge line is `⬢ hive: N new inbox record, run: hive inbox` for one
  record, and `⬢ hive: N new inbox records, run: hive inbox` for more than one.
- The nudge line does not contain `--home` or the home path.
- Remove a parameter or a field that becomes unused.
- Do not remove or change the `--home` flag of the CLI.

### R2 Tests

- Each test asserts the exact nudge text as a literal string, for example
  `"⬢ hive: 1 new inbox record, run: hive inbox"`. Do not compute the expected
  text with `nudgeLine`.
- Assert the singular and the plural form.
- Run each test against the base code first. It must fail there.

### R3 Docs

Find each doc, skill, or spec text outside the excluded folders that quotes the
old nudge text. Change it to the new text.

### R4 Brief copy

Copy this brief to `spec/hive-v1/tasks/<task-id>.md` with a baseline of 0.
When the copy causes a violation, change the words of that sentence and keep
its meaning. Record each change in the evidence.

## Not in scope

The `--home` flag, `home.Resolve`, the inbox command, `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each R2 test against the base code, then against the task code | Each changed test fails on the base and passes on the task branch. |
| V2 | From a directory below the home, for example `projects/hive/worktrees/<task-id>/`, build the task binary and run `env -u HIVE_HOME ./hive task show 36-nudge-no-home` and `env -u HIVE_HOME ./hive task list --project hive` with no `--home` | Each command exits 0 and shows the state of the home at `/home/penguin/source/hive`. Do not run an inbox read. |
| V3 | `git diff dev --stat` and `grep -rn 'hive --home' internal/supervise` | Only the files of the source table changed. The grep prints nothing. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `fix(supervise): drop the
home flag from the nudge`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V4 results to the manager
with `hive inbox send`. Read only your own inbox.
