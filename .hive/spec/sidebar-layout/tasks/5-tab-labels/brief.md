# Tab labels

| Item | Value |
| --- | --- |
| Wave | sidebar layout |
| Tier | small |
| Depends on | 3-manager-adopts-parent (landed on `dev` as `aa29b8d9`) |
| Plan | `/home/penguin/source/hive/plans/wave-sidebar-layout.md`, section "Task 3" |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

Each tab that holds a hive agent must show the name of the agent in the herdr
sidebar. Task 1 sets the label of a hive tab when it makes the tab. This task
renames the tab in the other cases: after a spawn into a tab that hive did not
make, and on `hive agent register`.

The user made each design decision in the plan. Do not change a decision.
When the code or herdr does not agree with the plan, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Trust boundaries

Read `/home/penguin/.claude/skills/orchestrate/references/untrusted-inputs.md`.
The worst result is a tab rename in a workspace that hive does not own. One
function sends `tab.rename`. Hive renames a tab only. Hive never renames a
workspace or a pane.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/lib/herdr/tab.go` | Add `TabRename`. |
| `internal/agent/spawn.go` | Rename the tab after the agent starts, per R1. |
| `internal/agent/spawn_worker.go` | The same for the worker and verifier path, when it needs a change. |
| `internal/agent/register.go`, or one new file in `internal/agent/` | One wrapper that registers and then renames, per R2. |
| `cmd/agent/register/register.go` | Call the wrapper. |
| `spec/hive-v1/03-architecture.md` | One line in the herdr layout section. |

You can add tests beside each file. Do not edit other files.

## Requirements

### R1 Rename after spawn

- After `agent spawn` starts an agent, hive renames the tab of that pane to
  the agent name when and only when `tab_created_by_hive` is false and
  `--pane` is absent.
- A worker, a verifier, and a manager in the root pane of a workspace that
  hive made each get one rename.
- A manager in a hive tab gets no rename, because task 1 set the label when it
  made the tab.

### R2 Rename on register

- `hive agent register` renames the tab of the registering pane to the role
  name, `coordinator` or `supervisor`. When the role has a project, the label
  is the agent name.
- `Register` in `internal/agent/register.go:69` writes the store only, and
  `Spawn` calls it (`spawn.go:117`). Add one wrapper in `internal/agent` that
  registers and then renames. Make `cmd/agent/register` call the wrapper
  (`cmd/agent/register/register.go:43-58`). Thus a spawn renames one time
  only.
- With `--pane`, get the tab ID from `pane.get`, not from `HERDR_TAB_ID`.

### R3 Failure

One function sends `tab.rename`. A failed rename is an effect in the command
result, and hive logs it. A failed rename never fails the spawn or the
register.

### R4 Spec

Add one line to the herdr layout section of
`spec/hive-v1/03-architecture.md`: hive gives each agent tab the agent name,
or the role name for the coordinator and the supervisor. Write it in
ASD-STE100.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Fake herdr records one `tab.rename` with the agent name for each of three spawns: worker, verifier, and manager. The manager spawn uses the root pane of a workspace that hive made. It records none for a manager spawn into a hive tab and none for a spawn with `--pane`. | V1 |
| AC2 | Fake herdr records one `tab.rename` with `coordinator` for `agent register --role coordinator`, and one with `supervisor` for the supervisor. A spawn gives no second rename from its inner register. | V2 |
| AC3 | A rename error is in `effects` of the `--json` result. The exit status is 0. The agent record is live. | V2 |
| AC4 | `grep -rn "TabRename" internal/ --include=*.go --exclude=*_test.go` lists one function and its callers only. | V3 |
| AC5 | The live check passes. | V4 |
| AC6 | The gate passes. | V5 |

## Verification items

### V1 Spawn rename tests

- Proves: AC1.
- Command: `go test ./internal/agent/ -run '<your test names>' -v`, red and
  green.
- Expected: the five cases of AC1 give the stated `tab.rename` requests.

### V2 Register and failure tests

- Proves: AC2, AC3.
- Command: `go test ./internal/agent/ ./cmd/agent/register/ -run '<your test names>' -v`,
  red and green.
- Expected: AC2 and AC3 as stated.

### V3 Call site search

- Proves: AC4.
- Command: `grep -rn "TabRename" internal/ --include=*.go --exclude=*_test.go`
- Expected: one definition and its callers. Paste the output.

### V4 Live check

- Proves: AC5.
- Command: run from the worktree with `--home "$PWD"`. Build the binary from
  the worktree first, and use that binary.
  1. `herdr tab create --workspace wBC --cwd "$PWD"`. Record the tab and the
     root pane.
  2. Wait until the shell is ready, then run
     `hive --home "$PWD" agent register --role supervisor --pane <that pane>`
     in that pane.
  3. `herdr tab list --workspace wBC`.
  4. Close the scratch tab by its pane ID.
- Expected: step 3 shows the scratch tab with the label `supervisor`. The
  register writes a supervisor claim and an agent record in the worktree home
  only. After step 4, `wBC` has its original tabs. No step sends
  `workspace.close`. Never close a pane that this check did not make.

### V5 Gate

- Proves: AC6.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- A pane label.
- A rename of a workspace.
- A rename of the tabs of the agents that run now.
