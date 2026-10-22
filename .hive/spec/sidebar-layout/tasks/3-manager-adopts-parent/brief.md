# Manager workspace adopts the parent checkout

| Item | Value |
| --- | --- |
| Wave | sidebar layout |
| Tier | strong |
| Depends on | none |
| Blocks | the friction and live check task, the tab labels task |
| Plan | `/home/penguin/source/hive/plans/wave-sidebar-layout.md`, section "Task 1" |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

Today hive makes the manager workspace at `projects/<id>/` with the label
`hive:<project>` (`internal/agent/workspace.go:160`). herdr does not see that
directory as a checkout. Thus no task worktree nests under the manager in the
herdr sidebar.

herdr nests a worktree workspace under the workspace that holds the primary
checkout of the repository. This task makes the manager use that workspace.
When it exists, hive adopts it and opens one new hive tab in it. When it does
not exist, hive makes one at the repository path. The manager learns its role
from its first prompt, which is the generated manager instructions.

Read the plan file first. Its section "Facts the user and the coordinator
established" gives each herdr behavior that this task uses. The user made each
design decision in that section. Do not change a decision. When the code or
herdr does not agree with a fact, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Trust boundaries

Read `/home/penguin/.claude/skills/orchestrate/references/untrusted-inputs.md`.
Hive does not control these inputs: the herdr socket replies, the workspace and
pane IDs from the state store, and the repository path from the project record.
The worst result is an agent start, a tab create, or a tab rename in a
workspace that hive does not own, or a close of a workspace that hive did not
create. A hostile herdr server is out of scope. One function resolves the
parent workspace.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/agent/workspace.go` | Resolve the manager pane in the order (a), (b), (c). Remove the label refusal. Extend `cleanup` for a created tab. |
| `internal/agent/spawn.go` | Call the resolver. Send the first prompt. |
| `internal/agent/stop.go` | Keep the workspace ID after a tab close. Treat `confirmation_required` as "pane kept". |
| `internal/agent/templates/project-agents.md` | Start with one line that names the role: manager of project ID. |
| `internal/lib/records/project.go` | Add `TabID` and `TabCreatedByHive`. |
| `internal/lib/herdr/worktree.go` | Add `WorktreeList`. |
| `AGENTS.md` | Role gate: a manager can run at the root of a checkout, and its brief names the role. |
| `spec/hive-v1/03-architecture.md` | herdr layout: the new manager placement. |
| One new file in `internal/agent/` | The shell ready wait (R10). Added 2026-10-18. |

You can add tests beside each file. Do not edit other files.

## Requirements

### R1 Pane resolution

`hive agent spawn --role manager --project ID` resolves the manager pane in
this order.

- (a) The project record names a workspace and a tab. The workspace exists.
  The tab exists in it. The tab holds one pane. That pane is a free shell with
  the repository path as its cwd. Then hive reuses that pane.
  - When the tab is gone or busy, drop the tab fields from the record. When
    the workspace still holds the checkout, continue with (b) on the same
    workspace, and keep the recorded `created_by_hive` value.
  - When the recorded workspace is gone or does not hold the checkout, clear
    the record and run `worktree.list`.
- (b) `worktree.list` with `cwd` set to the repository path gives a
  `source.source_workspace_id`. Then hive adopts that workspace. Hive sends
  `tab.create` with `workspace_id`, `cwd` set to the repository path, `label`
  set to the agent name, and `focus` false. Hive uses the root pane of the new
  tab. The record stores the workspace ID with `created_by_hive` false, the tab
  ID, and `tab_created_by_hive` true. Always give `cwd` explicitly, because a
  tab with no `cwd` follows the active pane.
- (c) No workspace holds the checkout. Then hive sends `workspace.create` at
  the repository path with the label `hive:<project>` and `focus` false. The
  manager pane is its root pane. The record stores `created_by_hive` true and
  the tab of the root pane with `tab_created_by_hive` false.

### R2 Old records and labels

- A record from the old layout has a root pane cwd of `projects/<id>/`. It
  fails (a) on the cwd. Hive clears the record. Hive never closes that old
  workspace.
- Remove the refusal of a label that is not `hive:<id>`
  (`internal/agent/workspace.go:47-51` and `:69`). An adopted workspace keeps
  its label. Hive never renames an adopted workspace.

### R3 Flags

- `--adopt ID` names the workspace to adopt in place of (b). It no longer
  needs the label `hive:<id>` (`workspace.go:115`). It needs a workspace that
  holds the parent checkout of the repository. Else the spawn fails.
- `--pane ID` bypasses the resolution as today (`spawn.go:299`). The first
  prompt still applies. Hive renames nothing.

### R4 The hive project

For the project `hive`, the parent is the hive session. Its first tab holds the
coordinator. The manager tab opens in that session. This is the intended
result.

### R5 Stop

- `agent stop` of a manager closes the pane as today.
- In an adopted workspace, the pane close closes the hive tab. The workspace
  survives. The topology is "surviving" (`internal/agent/stop.go:153-157`).
  When `tab_created_by_hive` is true, `recordStoppedWorkspace`
  (`stop.go:182-189`) drops `TabID` and `TabCreatedByHive` and keeps the
  workspace ID. On "closed", it clears the record as today.
- A workspace that hive made at the repository path becomes the root of the
  worktree group after a worker opens a worktree. Then herdr refuses
  `pane.close` of its last pane with `confirmation_required`. Stop treats that
  error as "pane kept": the stop marks the agent record as stopped, the project
  record keeps the workspace, and the result returns one effect that names the
  refusal.
- Hive never sends a confirmation flag. Hive never sends `workspace.close`
  with `close_group`.

### R6 Start failure cleanup

When `agent.start` fails after hive made a tab in an adopted workspace, hive
closes only the root pane of that tab with `pane.close` and drops the tab
fields. Hive never sends `workspace.close` when `created_by_hive` is false.
Extend `cleanup` (`workspace.go:187-191`), which today acts only on a created
workspace.

### R7 Not a repository

`worktree.list` fails with `not_git_worktree` on a path that is not a
repository. The spawn reports that error and stops. A listing from a linked
worktree gives the parent checkout (`reference/herdr/api/worktree.md:205`).

### R8 First prompt

- Hive still writes `projects/<id>/AGENTS.md` from
  `internal/agent/templates/project-agents.md`. The template starts with one
  line that names the role: manager of project ID.
- After the agent is ready, hive sends the file content as the first prompt
  through `agent.prompt`. Use the retry and turn-start logic of
  `promptTaskAgent` (`internal/agent/spawn_worker.go:392`).

### R9 Documents

- `spec/hive-v1/03-architecture.md`, herdr layout: the manager runs in a hive
  tab of the workspace that holds the parent checkout. Hive makes a workspace
  only when none exists. The manager learns its role from its first prompt.
- `AGENTS.md`, role gate: a manager can run at the root of a checkout, and its
  brief names the role.
- Write both in ASD-STE100. Use the terms of `STE-GLOSSARY.md`.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Fake herdr gives a parent workspace. The spawn sends one `tab.create` with the repository cwd and the agent name. It starts the agent in the returned root pane. The record holds the workspace ID, `created_by_hive` false, the tab ID, and `tab_created_by_hive` true. | V1 |
| AC2 | Fake herdr gives no parent workspace. The spawn sends one `workspace.create` at the repository path with the label `hive:<project>`, and no `tab.create`. The record holds `created_by_hive` true. | V1 |
| AC3 | A record that names a live hive tab with a free shell gets no `tab.create` and no `workspace.create`. A record whose tab is gone gets one `tab.create`. Then `TabID` is the tab ID that the fake gave, `TabCreatedByHive` is true, and `created_by_hive` keeps its old value. | V1 |
| AC4 | The spawn clears an old-layout record with cwd `projects/<id>/`. It sends no `workspace.close` request for that record. The spawn continues with (b) or (c). | V1 |
| AC5 | The spawn sends the generated instructions through `agent.prompt` after the agent is ready. With an instructions file of more than 4 KiB, the logged request body is 4 KiB or less and `truncated` is true. | V2 |
| AC6 | Three fake stop cases. For (b), the adopted topology is `surviving`. The record keeps the workspace ID and drops both tab fields. For (c), with no worktrees, the topology is `closed` and the record clears. For (c), with a worktree group, `pane.close` returns `confirmation_required`. The stop marks the agent record as stopped. The project record does not change. The result has an effect that names the refusal. Hive sends no `workspace.close` call. | V3 |
| AC6b | A fake `agent.start` failure after `tab.create` in an adopted workspace gives one `pane.close` call on the new tab's pane. It gives zero `workspace.close` calls and leaves a record with no tab fields. | V3 |
| AC7 | The code search below must list the definitions in `internal/lib/herdr`. It must list one `WorkspaceCreate(` call in the manager workspace resolver. It must list one `TabCreate(` call there. It must list the existing `TabCreate(` call in `spawn_worker.go`. | V4 |
| AC8 | The live check passes. | V5 |
| AC9 | `ste_test.go` covers `AGENTS.md` and the template through `steFiles`. The counts for both do not rise. The spec is not in `steFiles`. Write it in the same style. The gate passes. | V6 |
| AC10 | Fake herdr gives the shell ready on the second read. The spawn sends one `agent.start` and succeeds. The test covers a manager spawn and a worker spawn. | V7 |
| AC11 | Fake herdr never gives a shell. The spawn fails after the deadline with an error that names the pane. No `agent.start` goes. The test injects a short deadline. | V7 |

## Verification items

### V1 Resolution tests

- Proves: AC1, AC2, AC3, AC4.
- Command: `go test ./internal/agent/ -run '<your test names>' -v`, red and
  green. Write the test names in the evidence.
- Expected: the red run fails on each criterion for the correct reason. The
  green run passes. Each test records the herdr requests of the fake and
  compares them with the expected list.

### V2 First prompt test

- Proves: AC5.
- Command: `go test ./internal/agent/ -run '<your test names>' -v`, red and
  green.
- Expected: one `agent.prompt` after the agent is ready, with the file
  content. The log event for the large file has a body of 4 KiB or less and
  `truncated` true.

### V3 Stop and cleanup tests

- Proves: AC6, AC6b.
- Command: `go test ./internal/agent/ -run '<your test names>' -v`, red and
  green.
- Expected: the four cases of AC6 and AC6b give the stated requests, records,
  topology, and effects.

### V4 Call site search

- Proves: AC7.
- Command: `grep -rn "WorkspaceCreate(\|TabCreate(" internal/ --include=*.go --exclude=*_test.go`
- Expected: the list of AC7 and no other line. Paste the output in the
  evidence. Also search for other call sites of the parent resolver and of
  `WorktreeList`, and paste that output.

### V5 Live check

- Proves: AC8.
- Command: run from the worktree with `--home "$PWD"`.
  1. Register the worktree path as a project in that home.
  2. Spawn a manager with `--name t3-check-manager`, the harness `pi`, and the
     model `opencode-go/deepseek-v4.1-flash`. This model is cheap, so a stray
     action costs little.
  3. Run `herdr worktree list` with the worktree path as cwd. Record the
     parent workspace.
  4. Run `herdr tab list --workspace wBC`. Record the new tab.
  5. Run `herdr pane read` on the manager pane.
  6. Stop the manager at once with `hive --home "$PWD" agent stop`.
  7. Run `herdr tab list --workspace wBC` again.
- Expected:
  - Step 3 names the hive session `wBC` as the parent.
  - Step 4 shows a new tab with the label `t3-check-manager`. The pane of the
    agent record is the pane of that tab. It is not `wBC:p1`. It is not in the
    tab of the supervisor.
  - Step 5 shows that the first prompt arrived as one message.
  - Step 7 shows that the tab is gone and that `wBC` survives with its other
    tabs.
- Rules: close an orphan tab only by its pane ID. No step sends
  `workspace.close` for `wBC`. Never close a pane that this check did not
  make.

### V6 Gate

- Proves: AC9.
- Command: the four gate commands of `AGENTS.md`, and
  `go test . -run TestSTE -v`.
- Expected: `gofmt -l` prints nothing. The other commands pass. The counts for
  `AGENTS.md` and the template do not rise.

### R10 Shell ready wait

Added 2026-10-18. Source: the plan section "Change 2026-10-18" and the
behavior of its task 2.

- Before each `agent.start`, hive waits until the target pane is an available
  shell. Ready means: `pane.get` gives `agent_status` `unknown` and no
  `agent`, and `pane.process_info` gives a `shell_pid` that is not null and a
  `foreground_process_group_id` that is equal to `shell_pid`, and
  `foreground_processes` holds exactly one entry, with `pid` equal to
  `shell_pid`. Added 2026-10-19: shell startup runs in the group of the shell
  before job control starts.
- Fallback, only when V5 fails with the rule above: when `agent.start` gives
  `agent_pane_busy`, send `agent.start` again until the same deadline. After
  the deadline, keep the busy message of today.
- A `pane.process_info` failure counts as not ready. An example is a restore
  placeholder (`reference/herdr/api/pane.md:38`).
- The deadline is 10 seconds. The interval is 200 milliseconds. One function
  does this wait. `PaneProcessInfoGet` exists in
  `internal/lib/herdr/pane.go:120`.
- The wait goes before the one `AgentStart` call site in `spawn.go`. The
  worker and verifier path gets to it through `spawnTaskAgent`. Thus each role
  gets the wait.
- A pane that never becomes a shell fails with an error that names the pane
  and the deadline. Then no `agent.start` goes. A pane that herdr gives as busy
  because an agent holds it keeps the message of today.

### V7 Shell ready wait tests

- Proves: AC10, AC11.
- Command: `go test ./internal/agent/ -run '<your test names>' -v`, red and
  green.
- Expected: AC10 and AC11 as stated. A fake reply with the shell PID as the
  group ID and six foreground processes (five `bash`, one `tail -1`) is not
  ready. With the fallback, a fake that gives `agent_pane_busy` once and then
  success gives two `agent.start` and a spawn that succeeds. A fake that always
  gives `agent_pane_busy` fails after the deadline with the busy message. Also search for each call site of the
  wait function and of `AgentStart(`, and paste the output.

## Out of scope

- Migration code. The coordinator moves the two running managers by hand after
  the landing.
- The friction entry in `reference/dogfood/friction.md`. Another task of this wave marks it resolved.
- Tab rename after spawn and on register. Another task of this wave does it.
- A pane label.

## Changes

2026-10-18. The user decided that this task takes the shell ready wait. Each
manager spawn into the hive session opens a new tab, and `agent.start` gave
`agent_pane_busy` at once. Thus V5 could not pass. Added R10, AC10, AC11, V7,
and one new file in `internal/agent/`. Re-run V5 after V7 is green, until V5
passes. The friction entry stays out of scope.

2026-10-19. The V5 retry failed: the wait accepted a pane in shell startup.
R10 now also requires one foreground process, the shell. When V5 still fails
with that rule, add the `agent.start` retry of R10 and re-run V5.
