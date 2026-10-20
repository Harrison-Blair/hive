# Shell ready wait: friction record and live check

| Item | Value |
| --- | --- |
| Wave | sidebar layout |
| Tier | decent |
| Depends on | 3-manager-adopts-parent (landed on `dev` as `aa29b8d9`) |
| Plan | `/home/penguin/source/hive/plans/wave-sidebar-layout.md`, sections "Change 2026-10-18", "Change 2026-10-19", "Task 2", and "Backlog from this wave" |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

Task `3-manager-adopts-parent` added the shell ready wait before each
`agent.start` (`internal/agent/shell_ready.go`). The user moved that code into
task 1. This task has two remaining parts: a live check of three manager
spawns in a row, and the friction record.

This task changes no Go source. When the live check fails, do not change the
code. Record the failure in the evidence and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Trust boundaries

Read `/home/penguin/.claude/skills/orchestrate/references/untrusted-inputs.md`.
The live check makes and closes herdr workspaces. Close only a workspace that
this check made. Never close `wBC` or a pane that this check did not make.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `reference/dogfood/friction.md` | Mark the entry "Manager spawn fails with `agent_pane_busy` right after `workspace.create`" resolved. Add one entry for each backlog item. |

Do not edit other files, except the evidence files.

## Requirements

### R1 Resolved entry

- Set **Status** of the entry "Manager spawn fails with `agent_pane_busy`
  right after `workspace.create`" to resolved, with the date and the commit
  `aa29b8d9`.
- Add one sentence about the cause: during rc-file startup, the shell runs
  commands in its own process group. Thus the group ID equal to the shell PID
  is not sufficient. The landed rule also requires one foreground process,
  the shell.

### R2 New entries

Add one entry for each item of the plan section "Backlog from this wave",
except the item about the ready rule, which R1 covers. Use the format
Issue / Status / Summary / Reproduction steps of the file. Each new entry has
the status Open, except the race on worker and verifier spawns: task 1 also
fixes that, so mark it resolved with `aa29b8d9`. Write each entry in
ASD-STE100 and use the terms of `STE-GLOSSARY.md`.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Three manager spawns in a row into a new scratch project each succeed on the first attempt and make the workspace at the scratch path. | V1 |
| AC2 | The friction file has the resolved entry of R1 and the new entries of R2. | V2 |
| AC3 | The gate passes. | V3 |

## Verification items

### V1 Live check

- Proves: AC1.
- Command: run from the worktree with `--home "$PWD"`. Build the binary from
  the worktree first, and use that binary.
  1. Make a scratch Git repository under `.hive/tmp/` of the worktree, with
     one commit.
  2. Register it as a project in the worktree home.
  3. Spawn a manager with the harness `pi` and the model
     `opencode-go/deepseek-v4.1-flash`. Use one new agent name for each
     spawn.
  4. Record the result, the workspace ID, and the workspace cwd.
  5. Stop the manager with `hive --home "$PWD" agent stop`.
  6. Do steps 3 to 5 two more times.
  7. Close the scratch workspace when it still exists. Remove the scratch
     repository, the project record, and the files that the spawns made.
- Expected: each spawn gives the outcome `ok` on the first attempt. Each log
  shows the wait before `agent.start`, and no `agent_pane_busy`. The
  workspace cwd is the scratch path. After step 7, no scratch workspace and no
  scratch file stay.

### V2 Friction review

- Proves: AC2.
- Command: `git diff origin/dev -- reference/dogfood/friction.md`.
- Expected: the diff has the R1 change and one new entry for each R2 item,
  and no other change.

### V3 Gate

- Proves: AC3.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- A change to Go source.
- The tab labels task.
