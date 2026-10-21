# Stale coordinator check at startup

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, restart goal, wave 1 |
| Tier | strong |
| Depends on | none. Base: `dev` |
| Blocks | the restart task of wave 2 |
| Grant | `grants.md 2026-10-18 hive delivery on dev` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Bare `hive` does not start a new coordinator when an old coordinator record
has status `live` but its herdr pane is gone. This task makes startup check
the recorded pane. It also records that bare `hive` started the coordinator
and the supervisor, so that a later restart can close their panes.

Read these files first:

- `reference/dogfood/friction/entries/stale-coordinator-blocks-startup.md`,
  section `Plan`. The user approved it on 2026-10-21. This task does steps 1
  and 2.
- `.hive/spec/hive-tui/README.md`, mainly G4, G9, and G10.

When the code does not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `hasLiveCoordinator` in `internal/tui/startup.go` checks only
  `record.Status == records.AgentLive`.
- `hasFreshSupervisorClaim` in the same file checks the claim PID.
- Register in `internal/agent/register.go` marks the holder of the current
  claim `stale` when a new holder registers (the claim rule).
- `AGENTS.md`, section `herdr layout`: hive closes what it made, and never
  closes the pane of an agent that it did not start.

## Design

There is no new board. The only UI change is the footer text of a failed
pane check. Keep the existing footer style. The state has plain words: the
footer names the failed check and the pane ID. Record a screenshot of the
footer at 120x40 and at 200x57 in the evidence.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/startup.go` and its tests | R1, R2. |
| `internal/agent/` and its tests | R1, R2. Only the register and record code that R1 and R2 need. |
| `internal/lib/records/` and its tests | R2. Only a new record field. |
| `internal/lib/herdr/` and its tests | R1. Only a pane lookup that startup needs. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### AC1 A gone pane does not block startup

- When bare `hive` runs in herdr, it counts a coordinator record with status
  `live` as dead when herdr reports that the recorded pane does not exist.
- Then startup starts a new coordinator in a new `coordinator` tab, as for a
  fresh start.
- When herdr reports the pane, startup keeps the coordinator, as now.
- When the herdr lookup fails for a different reason, startup does not start
  a second coordinator. It shows the failure in the footer and logs it.
- Log the check with its outcome and the pane ID.

### AC2 The old record becomes stale

- When the new coordinator registers, hive marks the old coordinator record
  `stale` with the existing claim rule. Hive keeps the record file. Hive
  deletes no record.
- If the old record does not hold the claim, startup or register still marks
  it `stale`. Record the rule that you use in the evidence.

### AC3 Startup ownership

- The coordinator record and the supervisor record show that bare `hive`
  started them. Choose the field name and record it in the evidence.
- An agent that hive did not start in this way does not show the mark.
- Old records without the field load with no error.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | A `live` coordinator record with a gone pane does not block startup | V1, V3 |
| AC2 | The old coordinator record becomes `stale` and stays on disk | V1, V3 |
| AC3 | Records started by bare `hive` carry the startup mark | V1, V3 |

## Verification items

### V1 Tests

- Proves: AC1, AC2, AC3.
- Command: run each new test against the base code, then against the task
  code. Use the fake herdr server of the package tests.
- Expected: each new test fails on the base for the correct reason and passes
  on the task branch. Cases: a gone pane (new coordinator), a present pane
  (no new tab), a lookup error (no new coordinator), the old record marked
  `stale`, the startup mark on both records, and an old record with no mark.

### V2 Brief copy and ratchet

- Proves: AC1.
- Command: `go test ./ -run STE`
- Expected: pass, with the brief copy in the ratchet at 0.

### V3 Live check

- Proves: AC1, AC2, AC3.
- Command: in a new herdr workspace with a temp `HIVE_HOME` and a temp
  `XDG_CONFIG_HOME` with a cheap coordinator model, run the task binary. Quit
  the TUI. Close the `coordinator` pane by hand. Run the task binary again.
  List the tabs and the agent records.
- Expected: the second run makes a new `coordinator` tab. The old record is
  `stale` and its file exists. Both new records carry the startup mark. Stop
  the agents and close the workspace.

### V4 Diff scope

- Proves: AC1, AC2, AC3.
- Command: `git diff dev...HEAD --stat`, after you merge `dev` into the task
  branch.
- Expected: only the files of the source table and the evidence changed.

### V5 Gate

- Proves: AC1, AC2, AC3.
- Command: the gate of `AGENTS.md` from the worktree root.
- Expected: `gofmt -l` and `go mod tidy -diff` print nothing. The other
  commands exit 0.

## Rules

- Each Go change is test-first. See each new test fail for the correct reason.
- Live checks never start a second coordinator or supervisor for the real
  home. Do not run `rm -rf` on an unquoted variable.

## Out of scope

The TUI restart action, the `old` supervisor mark, and the `h` filter. The
wave 2 task builds them. `main`.

## Changes

- 2026-10-21: added the `Design` section for the footer text. V4 uses
  `git diff dev...HEAD`, because `dev` gets friction commits after the
  branch starts.
