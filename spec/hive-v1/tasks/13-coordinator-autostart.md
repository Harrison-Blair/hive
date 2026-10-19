# Automatic coordinator startup and the logs tab

| Item | Value |
| --- | --- |
| Wave | coordinator startup |
| Tier | decent |
| Depends on | 12-friction-log. Base: `dev` after 12-friction-log lands |
| Grant | `grants.md 2026-10-19 coordinator auto startup task` |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

This task changes instruction text only. A session in the root of the checkout
with no brief runs the coordinator startup before its first reply. The
coordinator startup opens a `logs` tab with the supervisor and the log stream.
Do not change a decision in this brief. When the files or herdr do not agree
with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `AGENTS.md` has the section `Role gate` with a table, two rules, and the
  `Bootstrap clause`. The table gives the coordinator role to the root of the
  checkout.
- `.agents/skills/hive-coordinator/SKILL.md` has the section `Startup` with
  seven steps. Step 6 runs `hive supervise` when the supervisor claim is
  absent or stale.
- `hive doctor` prints a line such as `ok supervisor: supervisor live: pid N`.
  The claim file is `.hive/state/claims/supervisor.json` in the home.
- `hive logs --follow` prints new log events as they arrive.
- herdr 0.9.5 has `herdr tab list`, `herdr tab create --workspace ID --cwd PATH
  --label TEXT`, `herdr pane split PANE_ID --direction right --ratio FLOAT`,
  `herdr pane run`, and `herdr pane send-text`. `reference/herdr/` is the API authority. Read it before you write a herdr command.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `AGENTS.md` | R1, in the section `Role gate`. |
| `.agents/skills/hive-coordinator/SKILL.md` | R2. |
| `testdata/ste-baseline.json` | Only to lower a count. |
| `spec/hive-v1/tasks/<task-id>.md` | New. A copy of this brief. |

Do not edit other files. Do not change Go code. Do not add a hive command.

## Requirements

### R1 Role gate

Change the `Role gate` of `AGENTS.md` so that:

- A session in the root of the checkout with no brief runs the coordinator
  startup of `.agents/skills/hive-coordinator/SKILL.md` before its first reply.
- A session that the user tells to "develop" under the `Bootstrap clause` does
  not run the coordinator startup.
- A session whose brief names another role does not run the coordinator
  startup. The brief wins, as today.
- Keep the rule that a manager or a worker never runs the coordinator startup.

### R2 Logs tab in the coordinator startup

Change the `Startup` section of the coordinator skill. Replace step 6 with
steps that do this, with herdr tab and pane commands:

1. Find the hive workspace. It is the workspace of the coordinator pane.
2. Find a tab with the label `logs` in that workspace. When it exists, use it.
   Do not open a second `logs` tab.
3. When no `logs` tab exists, open one with the home as the working directory.
   Split its root pane to the right with the ratio 0.5.
4. The right pane runs `hive logs --follow`. Start it only when the right pane
   does not run it already.
5. Check the supervisor claim with `hive doctor`. When the claim is absent or
   stale, the left pane runs `hive supervise`. When a supervisor is live, do
   not start a second supervisor. A live supervisor stays where it is.

Give each step the exact herdr command. Use only commands and flags that
`reference/herdr/` documents. Keep the other startup steps.

### R3 STE

Write the new text in ASD-STE100 Strict. Use each term only with its meaning
in `STE-GLOSSARY.md`. The counts in `testdata/ste-baseline.json` must not
rise.

### R4 Brief copy

Copy this brief to `spec/hive-v1/tasks/<task-id>.md` with a baseline of 0.
When the copy causes a violation, change the words of that sentence and keep
its meaning. Record each change in the evidence.

## Not in scope

Go code, a new hive command, the manager, worker, and verifier skills, `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Read the `Role gate` of `AGENTS.md` | Each R1 rule is present. The `Bootstrap clause` and the brief exceptions are explicit. |
| V2 | Read the coordinator `Startup` | Each R2 step is present, with an exact herdr command. |
| V3 | Check each herdr command and flag of R2 against `reference/herdr/` and `herdr ... --help` | Each command and flag exists in herdr 0.9.5. |
| V4 | Live dry run: in a scratch herdr workspace, run the R2 commands for the tab, the split, and the log pane. Then run them a second time. Do not start a second supervisor. Use a harmless command in place of `hive supervise` for the left pane. Close the scratch workspace after. | The first run makes one `logs` tab with two side-by-side panes. The second run makes no new tab and no new pane. |
| V5 | `go test -run TestSTE ./...` | Pass. No count in `testdata/ste-baseline.json` rose. |
| V6 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `docs(coordinator): start
the coordinator and the logs tab automatically`. Push the task branch. Write
the evidence. Run `hive task complete`. Send the commit SHA and the V1-V6
results to the manager with `hive inbox send`.
