# First-prompt fix

| Item | Value |
| --- | --- |
| Wave | first-prompt fix |
| Tier | strong |
| Depends on | none. Base: `dev` at `586cfe4f` or later |
| Grant | `grants.md 2026-10-19 first prompt fix task` |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

A spawn prints `agent NAME live` while the first prompt never arrives. This
task makes each spawn check delivery of its first prompt with the herdr
prompt wait. Do not change a decision in this brief. When the code does not
agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

The friction entry is
`reference/dogfood/friction/entries/agent-spawn-skips-first-prompt.md`. Do not
edit a file in `reference/dogfood/friction/`. The manager resolves the entry
after the landing.

## Known facts

- Read `/home/penguin/source/hive/.hive/tmp/reports/first-prompt-skip.md` and
  `/home/penguin/source/hive/.hive/tmp/reports/prompt-wait-probe.md` first.
- Cause: `promptStarted` in `internal/agent/spawn_worker.go` accepts an
  unchanged `idle` as delivery, so `promptTaskAgent` skips its retry.
- Probe, 24 runs: `agent.prompt` with a wait until `working` or `blocked`
  returns `agent_prompt_stalled` after about 5.3 s for a lost Pi prompt. A
  retry delivers one time. A fast finish gives `working`, then `done`. A dialog
  gives `agent_blocked` and types nothing.
- `internal/lib/herdr/agent.go` has `AgentPromptWaitOptions` and the field
  `Wait` of `AgentPromptParams`. `internal/lib/herdr/errors.go` has
  `CodeAgentPromptStalled` and `CodeAgentBlocked`.
- First-prompt call sites: `promptTaskAgent` in `internal/agent/spawn_worker.go`
  (worker and verifier) and the `herdr.AgentPrompt` call near line 426 of
  `internal/agent/spawn.go` (check which spawn it serves). `internal/agent/send.go`
  and `internal/supervise/supervise.go` are not first prompts. Do not change
  them.
- The fake herdr of `internal/agent` can fail a method with a code
  (`fake.failure("agent.prompt", "agent_prompt_stalled", ...)`).

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/agent/spawn_worker.go`, `internal/agent/spawn.go` | R1, R2. Only the files that the fix needs. |
| `internal/lib/herdr/agent.go` | Only when the fix needs a helper. |
| Tests beside each changed file | R3. |
| `testdata/ste-baseline.json` | Only to add the brief copy with 0. |
| `spec/hive-v1/tasks/<task-id>.md` | R5. New. A copy of this brief. |

Do not edit other files.

## Requirements

### R1 Prompt wait

- The first prompt of each spawn uses the `agent.prompt` wait until `working`
  or `blocked`, with a bounded timeout. Use 10 s, and name it as a constant.
- On `agent_prompt_stalled`, hive retries the prompt one time with the same
  wait.
- When the retry also stalls, the spawn fails. The output names the agent and
  the cause. Hive never prints `live` for an agent without a checked prompt.
- On `agent_blocked`, the spawn reports the agent as blocked and sends no more
  text.
- Keep the current handling of `agent_not_ready`.

### R2 Remove the old check

Remove the unchanged-`idle` check as evidence of delivery (`promptStarted` and
`promptStatusStarted`, when nothing else uses them). The user accepts the risk
of a duplicate prompt when an accepted prompt stays idle longer than the herdr
window. Do not add a pane-read guard.

### R3 Tests

Work test-first. Use the fake herdr. Write each test, run it, and make sure
that it fails for the correct reason before you change the code. The tests
cover at least:

1. The fake stalls first and then accepts: one retry, then `live`.
2. The fake stalls two times: the spawn fails, names the agent and the cause,
   and does not print `live`.
3. The fake returns a blocked state: the spawn reports blocked, and the fake gets one
   prompt only.
4. Each prompt request carries the wait until `working` and `blocked` with the
   bounded timeout.

Show in the evidence that each new test fails against the old code, then
passes.

### R4 Live check

Build hive from your branch to `.hive/tmp/<task-id>/hive` in your worktree. Use
a scratch hive home and a scratch git repository below `.hive/tmp/<task-id>/`.
Register the scratch repository as a project. Then three times: create a
scratch task, run `hive worktree create`, and spawn a Pi worker
(`openai-codex/gpt-6-luna:xhigh`) with no `--pane`. Use a harmless brief that
tells the agent to reply `OK` and stop. Each spawn must get its first prompt
with no hand resend. Read each pane to check. Record each command and output.
Then stop each scratch agent, remove each scratch worktree, close each scratch
workspace, and delete `.hive/tmp/<task-id>/`. Make sure that
`herdr workspace list` does not show them.

### R5 Brief copy

Copy this brief to `spec/hive-v1/tasks/<task-id>.md` with a baseline of 0.
When the copy causes a violation, change the words of that sentence and keep
its meaning. Record each change in the evidence.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each R3 test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | The R4 live check | Three of three spawns get the first prompt with no hand resend. The scratch resources are gone. |
| V3 | `grep -n 'promptStarted\|promptStatusStarted' internal/agent/*.go` | No match in non-test code, or each match is not a delivery check. |
| V4 | `git diff dev --stat` | Only the files of the source table changed. |
| V5 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `fix(agent): check the
first prompt with the herdr prompt wait`. Push the task branch. Write the
evidence. Run `hive task complete`. Send the commit SHA and the V1-V5 results
to the manager with `hive inbox send`.
