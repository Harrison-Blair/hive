# Restart of the coordinator and the supervisor from the TUI

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, restart goal, wave 2 |
| Tier | strong |
| Depends on | the stale coordinator task of wave 1. Base: `dev` after it lands |
| Blocks | none |
| Grant | `grants.md 2026-10-18 hive delivery on dev` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The user can restart the coordinator or the supervisor from the Agents screen
of the TUI. Old records keep their files and get the correct status. The
header marks a supervisor that runs an older binary than the installed
`hive`.

Read these files first:

- `reference/dogfood/friction/entries/stale-coordinator-blocks-startup.md`,
  section `Plan`. The user approved it on 2026-10-21. This task does steps 3
  to 7.
- `reference/dogfood/friction/entries/supervisor-runs-stale-binary.md`. This
  task removes its manual workaround.
- `.hive/spec/hive-tui/README.md`, mainly decisions 1, 14, 26 and G9, G10,
  G12, G14. Decision 14 and G12 say that the TUI has no restart action. The
  user approved the restart on 2026-10-21. Add a new numbered decision that
  records this approval. The decision makes the restart of the coordinator
  and the supervisor the one exception to decision 14 and G12. The user approved only this restart action. Change G12 to refer to the new decision. Do not
  change other decisions.
- The startup mark from the wave 1 task, in its evidence.

When the code does not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Design

There is no approved board for this task. Use the existing Agents screen,
header, and footer styles. Each state has a word in plain text: `old`,
`restart coordinator? y/n`, `restarting`, and the result. Record a screenshot
at 120x40 and at 200x57 in the evidence.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/` and its tests | AC1-AC5. |
| `internal/agent/` and its tests | AC2-AC4. Only the restart and record code. |
| `internal/lib/herdr/` and its tests | AC2, AC3. Only a call that restart needs. |
| `internal/agent/start_coordinator.go` and its tests | AC6. Only the readiness wait. |
| `internal/lib/herdr/` readiness fields and their tests | AC6. Only the fields that the wait reads. |
| `internal/lib/records/` and its tests | AC4. Only a status that AC4 needs. |
| `.hive/spec/hive-tui/README.md` | The new decision and the G12 change. |
| `.agents/skills/hive-coordinator/SKILL.md`, `skills_test.go` | Only to replace a manual supervisor restart step with the TUI action, if the skill has one. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### AC1 The restart key

- On the Agents screen, the user selects the coordinator node or the
  supervisor node and presses `r`.
- When the target is live, the footer asks for a confirmation. `y` restarts.
  Each other key cancels.
- When the target is dead or stale, hive restarts it at once.
- On each other node, `r` does nothing and the footer shows that restart
  applies only to the coordinator and the supervisor.
- Log each user action and each result.

### AC2 Coordinator restart

- Hive closes the old coordinator pane only when its record carries the
  startup mark. Else hive does not close the pane, and tells the user.
- Hive opens a new `coordinator` tab and starts the configured harness and
  model with the startup prompt, with the same code as bare `hive`.

### AC3 Supervisor restart

- Hive closes the supervisor pane through herdr when its record carries the
  startup mark. This task accepts the abrupt SIGHUP exit.
- Hive opens a new `supervisor` tab and runs `hive supervise` in it.

### AC4 Old records

- After a restart, the old record has status `stale` or `stopped`. Record
  the rule in the evidence.
- Hive keeps each record file. Hive deletes no record. `h` hides the old
  records, as for other stopped agents.

### AC5 Old supervisor mark

- The header marks the supervisor `old` when it runs an older binary than the
  installed `hive`, with the hint `r to restart`.
- Use the same binary check as the Health screen.

### AC6 Coordinator start waits for readiness

- `StartCoordinator` sends no first prompt while Herdr reports
  `launch_pending: true`. It waits for `interactive_ready: true` and
  `launch_pending: false`, with a bounded timeout.
- The same code serves bare `hive` and the restart. When a shared wait
  covers the restart, remove the restart-only shell-prompt wait.
- On a timeout, the footer names the failed step. Hive does not send the
  prompt a second time. Log the wait with its duration.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | `r` restarts a selected coordinator or supervisor, with a confirmation only when it is live | V1, V3 |
| AC2 | A coordinator restart closes only a pane that hive started, then starts a new coordinator | V1, V3 |
| AC3 | A supervisor restart closes its pane and runs `hive supervise` in a new tab | V1, V3 |
| AC4 | Old records get `stale` or `stopped`, stay on disk, and `h` hides them | V1, V3 |
| AC5 | The header shows `old` and `r to restart` for an old supervisor binary | V1, V3 |
| AC6 | Coordinator start waits for herdr readiness before the first prompt | V1, V3 |

## Verification items

### V1 Tests

- Proves: AC1, AC2, AC3, AC4, AC5.
- Command: run each new test against the base code, then against the task
  code. Test the TUI at the model with `Update` and `View`. Use the fake herdr
  server for the restart calls.
- Expected: each new test fails on the base for the correct reason and passes
  on the task branch.

### V2 Brief copy and ratchet

- Proves: AC1.
- Command: `go test ./ -run 'STE|Skill'`
- Expected: pass.

### V3 Live check

- Proves: AC1, AC2, AC3, AC4, AC5.
- Command: in a new herdr workspace with a temp `HIVE_HOME` and a temp
  `XDG_CONFIG_HOME` with a cheap coordinator model, run the task binary.
  Restart the live supervisor with `r` and `y`. Restart the live coordinator.
  Close the coordinator pane by hand and restart it with `r`. Make the
  installed binary newer than the supervisor binary and look at the header.
- Expected: each restart opens a new tab and closes the old pane. The old
  records keep their files with the correct status. The header shows `old`
  and `r to restart`. Record screenshots. Stop the agents and close the
  workspace.

### V4 Diff scope

- Proves: AC1, AC2, AC3, AC4, AC5.
- Command: `git diff dev...HEAD --stat`, after you merge `dev` into the task
  branch.
- Expected: only the files of the source table and the evidence changed.

### V5 Gate

- Proves: AC1, AC2, AC3, AC4, AC5.
- Command: the gate of `AGENTS.md` from the worktree root.
- Expected: `gofmt -l` and `go mod tidy -diff` print nothing. The other
  commands exit 0.

## Rules

- Each Go change is test-first. See each new test fail for the correct reason.
- Live checks never restart the coordinator or the supervisor of the real
  home. Do not run `rm -rf` on an unquoted variable.

## Out of scope

The resolution of the friction entries. The coordinator does it after this
task lands. `main`.

## Changes

- 2026-10-21: the user chose to keep the `r` restart action as the one
  exception to decision 14 and G12.
- 2026-10-21: V4 uses `git diff dev...HEAD`, because `dev` gets friction
  commits after the branch starts.
- 2026-10-21: the manager added AC6 and its files. The worker found that
  `StartCoordinator` sends the prompt while herdr reports `launch_pending`.
