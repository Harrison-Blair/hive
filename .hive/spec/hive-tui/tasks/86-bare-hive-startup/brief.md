# Startup of bare hive

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, wave C |
| Tier | strong |
| Depends on | 80-chat-screen, 81-agents-screen. Base: `dev` after both land |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Bare `hive` sets up the hive session: the TUI in a tab named `hive`, the
coordinator in its own tab, and the supervisor in its own tab. The `logs` tab
goes away. The config gets the coordinator harness and model. The feature
README is `.hive/spec/hive-tui/README.md`. Read it first, mainly goal
decisions 18, 20, 21, 26 and G9, G10, G11. The manager approved these decisions. Do not change them. When the code does not agree with this brief or the
README, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `internal/lib/home/config.go` reads and writes only `home = "PATH"`.
- `hive agent spawn` does not start a coordinator. The coordinator registers
  itself in startup step 1 of `.agents/skills/hive-coordinator/SKILL.md`.
- Steps 7 to 13 of that skill make the `logs` tab, run `hive logs --follow`,
  and start `hive supervise` in it.
- `StartSupervisor` in `internal/agent/supervisor.go` refuses a live claim.
- `internal/lib/harness/` knows how to start each harness with a model.
- herdr sets `HERDR_ENV=1`, `HERDR_PANE_ID`, and `HERDR_SOCKET_PATH` in a
  pane.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/lib/home/` and its tests | R1. |
| `cmd/homeinit/`, `internal/homeinit/` and their tests | R1. Only to keep the other config lines. |
| `cmd/root.go`, `internal/tui/` (startup files) and their tests | R2-R4. |
| `internal/agent/` and its tests | R3. Only to start a coordinator pane with the harness start code. |
| `internal/lib/herdr/` and its tests | R3. Only a call that the startup needs. |
| `.agents/skills/hive-coordinator/SKILL.md`, `AGENTS.md` (section `Coordinator`), `skills_test.go` | R5. |
| `testdata/ste-baseline.json` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### R1 Config (G11)

- The config accepts `home = "PATH"` and a `[coordinator]` table with
  `harness = "KIND"` and `model = "MODEL"`. Defaults: `claude` and
  `claude-opus-5-5`. You can add comments and blank lines.
- An unknown key or table fails with an error that names the key and the
  config path.
- `hive init` changes only the `home` line and keeps each other line.
- No new dependency.

### R2 Without herdr (G9)

When `HERDR_ENV` is not `1` or `HERDR_PANE_ID` is empty, bare `hive` runs only
the TUI, with the footer notice `not in herdr: no coordinator or supervisor
started`.

### R3 With herdr (G9, G10)

Each step checks first and is safe to run again:

1. Label the tab of the current pane `hive`.
2. Coordinator: when a live coordinator record exists, keep it. Else create
   a tab `coordinator` in the workspace of the current pane, with the home as
   its working directory, and start the harness and model of the config in
   its root pane. The first prompt tells the agent that it is the
   coordinator and to run the coordinator startup.
3. Supervisor: when a fresh supervisor claim exists, keep it. Else create a
   tab `supervisor` in the same workspace and run `hive supervise` in its
   root pane.
4. Run the TUI in the current pane.

- Each step runs before the TUI starts, or in a `tea.Cmd`. A failed step shows
  in the footer. It does not stop the TUI.
- Startup never closes or relabels a tab or pane that it did not make, except
  the label of the current tab. It never stops an old `logs` tab.
- When the TUI quits (two `ctrl+c`), the coordinator and the supervisor keep
  running.
- Log each step with its outcome.

### R4 Tests

Test each step against the fake herdr server of the package tests: a fresh
start, a second start with a live coordinator and a fresh claim (no new tab),
a stale claim, and no herdr environment.

### R5 Instruction text

- The coordinator skill: remove steps 7 to 12 (the `logs` tab). Change the
  supervisor step: bare `hive` starts the supervisor, so the coordinator
  starts it only when the claim is stale and the `supervisor` tab is gone.
  Add one step: when an old `logs` tab exists and the `supervisor` tab is
  live, close the old `logs` tab by hand, after a check that its panes run
  only `hive logs --follow` or an old `hive supervise`, and tell the user.
- `AGENTS.md` section `Coordinator`: say that bare `hive` starts the
  coordinator and the supervisor in their tabs.
- Keep the text in ASD-STE100 Strict.

## Rules

- Each Go change is test-first. See each new test fail for the correct
  reason.
- When `ste_test.go` fails on the brief copy, change the words of each
  sentence in the copy that causes a violation. Keep its meaning. Record each
  change in the evidence.
- Live checks use a temp `HIVE_HOME`, a temp `XDG_CONFIG_HOME`, a cheap
  coordinator model in the temp config, and a new herdr workspace. Clean up
  each workspace. Never start a second coordinator or supervisor for the real
  home. Do not run `rm -rf` on an unquoted variable.
- Other wave C tasks run at the same time and edit the shell wiring in
  `internal/tui/`. Keep the change to shared files small.

## Not in scope

A restart of a live coordinator or supervisor. `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each new test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Temp config tests for an unknown key, a `[coordinator]` table with both keys, and `hive init` on a config with a `[coordinator]` table and a comment | The unknown key fails and names the key. The parser reads the table. Init keeps the other lines. |
| V3 | Live, in a new herdr workspace with the temp home and config: run the task binary. List the tabs. Press `ctrl+c` two times. Run it again. List the tabs and agents again. | First run: tabs `hive`, `coordinator`, `supervisor`, and a live coordinator and supervisor. After quit both keep running. Second run: no new tab. Stop the agents and clean up the workspace. |
| V4 | Run the task binary in a plain terminal pane with `HERDR_ENV` unset | Only the TUI starts, with the notice. |
| V5 | `git diff dev --stat` | Only the files of the source table and the evidence changed. |
| V6 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `feat(cli): start the hive
session from bare hive`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V6 results to the
manager with `hive inbox send`.
