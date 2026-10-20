# Agent lineage and status time

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, wave A |
| Tier | decent |
| Depends on | 73-hive-tui-spec. Base: `dev` after it lands |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Record who spawned each agent, which generation it replaces, and when its harness status changed (README G4, G5). The feature README is `.hive/spec/hive-tui/README.md`. Read it first. Keep the feature README's approved decisions unchanged. When the code does not agree
with this brief or the README, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/lib/records/agent.go` and its tests | R1. |
| `internal/agent/spawn.go`, `spawn_worker.go`, `register.go` and their tests | R2. |
| `cmd/agent/spawn/` and its tests | R2. Add the `--agent NAME` flag (the caller agent) and pass it to `internal/agent`. |
| `internal/supervise/` (each place that sets `AgentStatus`) and its tests | R3. |
| `internal/agent/show.go`, `list.go`, `render.go` and their tests | R4. |
| `testdata/ste-baseline.json` | Only to lower a count. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### R1 Fields

The agent record has `parent`, `replaces`, and `status_changed_at`, each an
optional string with `omitempty`. An old record without them stays valid.

### R2 Lineage

- On a spawn, `parent` is the key of the live agent record of the caller: the
  `--agent` value when given, else the agent in the caller pane. When hive
  finds no caller agent, `parent` stays empty.
- `replaces` is the key of the newest earlier record with the same name, or
  for a manager or the coordinator the same role, that is not live. Else it
  stays empty.
- A verifier gets its manager as parent in the normal flow, because the
  manager spawns it.

### R3 Status time

Each place that changes `agent_status` sets `status_changed_at` to the time of
the change, only when the value changes. A write that does not change
`agent_status` keeps the old `status_changed_at`.

### R4 Show

`hive agent show` prints `parent`, `replaces`, and `status since` when set.
`hive agent list --json` includes the fields.

## Rules for each task

- Each Go change is test-first. See each new test fail for the correct
  reason. Do not compute an expected value with the code under test.
- When `ste_test.go` fails on the brief copy, change the words of each
  sentence in the copy that causes a violation. Keep its meaning. Record each
  change in the evidence.
- Live checks use a temp home (`HIVE_HOME`) and a temp `XDG_CONFIG_HOME`, and
  clean up each herdr workspace that they make. Do not run `rm -rf` on an
  unquoted variable.
- Other wave A tasks run at the same time. Edit only the files of your source
  table. Expect a small conflict in `cmd/root.go` or `testdata/ste-baseline.json`
  when the task lands.

## Not in scope

The Agents screen (wave B). `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each new test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Live herdr check in a temp home. Register a fake manager in your pane. Spawn one worker with `--agent` set to that manager and a cheap harness. Stop it and spawn it again. Run `hive agent show` for the second worker. | The second worker shows the manager as `parent` and the first worker record as `replaces`. Clean up each workspace. |
| V3 | `git diff dev --stat` | Only the files of the source table and the evidence changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `feat(agent): record the
agent lineage`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V4 results to the manager
with `hive inbox send`.

## Scope change 1

`hive agent spawn` has no `--agent` flag. Add it in `cmd/agent/spawn/`: the
value is the name or key of the caller agent. When the flag is not given, hive
finds the caller agent from the pane of the caller, as R2 says. The source
table now names `cmd/agent/spawn/`. Do not change the other spawn flags.

## Scope change 2

R2 `replaces` changes to this rule. `replaces` is the key of the newest
earlier record that is not live and has the same role and:

- Worker or verifier: same task.
- Manager: same project.
- Coordinator: no other condition.

When no record agrees, `replaces` stays empty. V2 stays as it is: the second
worker of the task replaces the first worker.
