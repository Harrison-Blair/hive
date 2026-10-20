# Task briefs

Each file in this directory is the brief of one task in the hive v1 plan. A
worker reads its brief before any other file. [../05-tracker.md](../05-tracker.md)
gives the order of the tasks, their status, and the procedures.
[../evidence/README.md](../evidence/README.md) gives the evidence format.

## Rules

- One task has one file with the name `<task-id>.md`. A task ID has the form `T<nn>-<slug>`.
- Write each brief in ASD-STE100 Strict. `ste_test.go` checks each file in this directory against a baseline of 0.
- A brief names each file that the worker edits. The worker edits no other source file. A new test file beside a named file needs no entry.
- Each acceptance criterion has the ID `AC<n>`. Each verification item has the ID `V<n>` and names the criteria that it proves. Each criterion appears in at least one item.
- A test item records a red run and a green run. The last item of each brief is the gate.
- A brief gives the behavior, not the code. The worker chooses each name that the brief leaves open and records the choice in the evidence.
- A brief can change after its task starts. The merger makes the change on `dev`, with the date, in the "Changes" section. The worker merges `dev` into its branch and reads the change.
- Run each CLI check from the worktree root as `HIVE_HOME="$PWD" hive CMD`. The worktree marker makes it a home. The real home stays clean.
- A live check against herdr acts on the herdr session of the user. The check builds the binary first with `go build .`. The worker closes each workspace, tab, and agent that the check made. When herdr does not know a repository, its first `worktree.create` also opens a workspace for that repository. The worker closes that workspace too.
- A live check gives each agent a unique name with `--name t<nn>-check-<suffix>`. herdr names are unique in the session, and several tasks run at the same time.
- A task that adds an instruction file, such as a skill or a template, adds it to the ratchet. It adds one line to `steFiles` in `ste_test.go` and one entry of 0 to `testdata/ste-baseline.json`. The brief names both files in its source table.
- A live check never changes the real home. Run it as `HIVE_HOME="$PWD" hive CMD` from the worktree. Store its fixtures under `.hive/tmp/` in that worktree.

## Template

```markdown
# T<nn> <title>

| Item | Value |
| --- | --- |
| Task ID | T<nn>-<slug> |
| Wave | <n> |
| Tier | strong, decent, or small |
| Depends on | task IDs and checkpoints, or none |
| Blocks | task IDs and checkpoints |
| Branch | `task/T<nn>-<slug>` |
| Worktree | `projects/hive/worktrees/T<nn>-<slug>/` |
| Evidence | `evidence/T<nn>-<slug>/` |

## Overview

What the task builds and why it sits here in the plan.

## User flows

Numbered flows. Each gives who acts, which command runs, and what the actor sees.

## Source files

| Path | Purpose of the edit |
| --- | --- |

## Requirements

### AC<n> <title>

Bulleted requirements that make the criterion precise.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |

## Verification items

### V<n> <title>

- Proves: AC<a>, AC<b>.
- Command: the exact command or a fenced block of commands.
- Expected: what the output shows when the criterion holds.

## Out of scope

What the worker does not build, and which task builds it.

## Changes

None, or dated entries by the merger.
```
