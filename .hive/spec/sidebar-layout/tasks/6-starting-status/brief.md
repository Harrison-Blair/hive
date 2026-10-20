# Starting status for a spawned agent

| Item | Value |
| --- | --- |
| Wave | sidebar layout |
| Tier | decent |
| Depends on | 5-tab-labels (landed on `dev` as `18111ae1`) |
| Plan | `/home/penguin/source/hive/plans/wave-sidebar-layout.md`, section "Task 6" |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

The spawn writes a live agent record. Then it runs the shell ready wait for
approximately 300 ms. Then it calls `agent.start`. In that period the
supervisor function `onAgentDetected` (`internal/supervise/events.go:121`)
sees the pane as a released shell and marks the record `exited`. The spawn
then fails with "agent record changed", and the agent runs as an orphan with
no first prompt. The manager saw this on two verifier spawns on 2026-10-19.

This task adds the agent status `starting`. The user selected this approach.
Do not change a decision. When the code or herdr does not agree with the plan,
stop and report it with `hive inbox send --to manager-hive TEXT --agent NAME`.

## Trust boundaries

Read `/home/penguin/.claude/skills/orchestrate/references/untrusted-inputs.md`.
The inputs that hive does not control are the herdr socket replies and the
supervisor events. The worst result is a record that stays `starting` for all
time, or an agent that the supervisor does not see. A `starting` record must
always go to `live`, `exited`, or `unresolved`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/lib/records/agent.go` | Add the status value `starting`. |
| `internal/agent/spawn.go`, `spawn_worker.go` | Write `starting` before the shell ready wait. Write `live` after the agent is ready. Write `exited` on a start failure, as today. |
| `internal/agent/stop.go` | Wait for a `starting` record to settle, per R4. |
| `internal/agent/list.go`, `show.go`, `render.go`, `inspect.go`, `resolve.go`, `send.go`, `register.go` | Only where a `live` check must also see `starting`, or where a status shows. |
| `internal/supervise/*.go` (no new package) | Per R3. |
| `internal/doctor/*.go` | Show `starting` as its own status. |
| `internal/tui/*.go` | Show `starting` as its own status. |
| `spec/hive-v1/03-architecture.md` | Name the status in the agent lifecycle section and the supervision table. |
| `reference/dogfood/friction.md` | Per R5. |

You can add tests beside each file. Do not edit other files.

## Requirements

### R1 Status value

The records package gains the agent status `starting`. Examine each use of
`records.AgentLive` (`grep -rn AgentLive internal/ --include=*.go`). For each
use, decide if a `starting` record must also match. Record each decision in
the evidence with one line for each call site.

### R2 Spawn

- The spawn writes the agent record with the status `starting` before the
  shell ready wait.
- After `agent.start` returns and the agent is ready, the spawn sets the
  record to `live`.
- On a start failure the record goes to `exited`, as today.
- This applies to each role: manager, worker, and verifier, with and without
  `--pane`.

### R3 Supervisor

- The supervisor ignores a `released` or `unknown` pane report for a record in
  `starting`. It writes no inbox record and does not change the agent record.
  It logs the ignored event.
- A closed pane, tab, or workspace for a `starting` record gives `unresolved`,
  as for a `live` record today.
- Reconcile at supervisor start treats a `starting` record with no pane as
  today: `unresolved`.

### R4 Commands

- `agent list`, `agent show`, `doctor`, and the TUI show `starting` as its own
  status.
- `agent stop` on a `starting` record waits for the start to settle, up to the
  start timeout. Then it stops as today.

### R5 Spec and friction

- Name the status `starting` in the agent lifecycle section and in the
  supervision table of `spec/hive-v1/03-architecture.md`. Write it in
  ASD-STE100.
- In `reference/dogfood/friction.md`, mark the spawn-versus-supervisor race
  resolved with the date and the commit, when an entry exists. Else add one
  in the Issue / Status / Summary / Reproduction steps format.
- Add one entry: `hive agent send` to a pi agent spawned with `--pane` failed
  with `agent_not_ready` two times on 2026-10-19.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Fake herdr: a spawn whose shell ready wait takes two reads writes `starting` before the first read and `live` after `agent.start`. A test reads the record between the two reads. | V1 |
| AC2 | Fake supervisor input: a `released` report for a `starting` record changes nothing and writes no inbox record. The same report for a `live` record marks it `exited`, as today. | V2 |
| AC3 | A closed pane for a `starting` record marks it `unresolved`. | V2 |
| AC4 | `agent list --all` prints `starting` for such a record. | V3 |
| AC5 | The live check passes. | V4 |
| AC6 | The gate passes. | V5 |

## Verification items

### V1 Spawn test

- Proves: AC1.
- Command: `go test ./internal/agent/ -run '<your test names>' -v`, red and
  green.
- Expected: the record status is `starting` between the two shell reads and
  `live` after `agent.start`.

### V2 Supervisor tests

- Proves: AC2, AC3.
- Command: `go test ./internal/supervise/ -run '<your test names>' -v`, red
  and green.
- Expected: AC2 and AC3 as stated.

### V3 List test

- Proves: AC4.
- Command: `go test ./internal/agent/ -run '<your test names>' -v`, red and
  green.
- Expected: the text output of `agent list --all` holds `starting` for that
  record.

### V4 Live check

- Proves: AC5.
- Command: run from the worktree with `--home "$PWD"`. Build the binary from
  the worktree first, and use that binary.
  1. Open a scratch tab in the hive session with
     `herdr tab create --workspace wBC --cwd "$PWD"`. Start
     `hive --home "$PWD" supervise` in its pane.
  2. Five times: open a fresh scratch tab with `herdr tab create --workspace
     wBC --cwd "$PWD"`, then run a verifier-style spawn into its root pane
     with `--pane`, the harness `pi`, the model
     `opencode-go/deepseek-v4.1-flash`, `--arg --no-approve`, and a new name
     `t6-check-<n>`. Stop each agent at once with `hive agent stop`.
  3. Stop the supervisor and close each scratch tab by its pane ID.
- Expected: each of the five spawns succeeds on the first attempt. No output
  holds "agent record changed". Each record goes `starting`, then `live`,
  then `stopped`. After step 3, `wBC` has its original tabs. No step sends
  `workspace.close`. Never close a pane that this check did not make. If a
  verifier-style spawn needs a task, create a scratch task in the worktree
  home.

### V5 Gate

- Proves: AC6.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- A change to the shell ready rule.
- A migration of the records of the agents that run now.
- A fix for the `hive agent send` failure. Record it only.
