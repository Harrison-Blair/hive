# First-prompt skip report

| Item | Value |
| --- | --- |
| Wave | wave 4 of 4, task A, report only |
| Tier | decent |
| Depends on | none |
| Grant | none. No code change, no commit, no landing. |
| Output | `/home/penguin/source/hive/.hive/tmp/reports/first-prompt-skip.md` |

## Overview

This task is an investigation. A spawn prints `prompted agent` and
`agent NAME live`, but the agent never gets its first prompt. Find why. Do not
change a source file. Do not commit. Do not push. This brief replaces the
commit, push, gate, and evidence steps of the worker skill. The report file is
the only output. Do not edit a file in `reference/dogfood/friction/`. The
manager adds the report path and the cause to the friction entry.

## Known facts

- The friction entry is
  `reference/dogfood/friction/entries/agent-spawn-skips-first-prompt.md`.
- It occurred for Pi and Claude, with and without `--pane`.
- Observations of the manager on 2026-10-19, harness Pi with
  `openai-codex/gpt-6-luna:xhigh`: each first worker spawn into a new task
  worktree (`hive worktree create`, then `hive agent spawn --role worker` with
  no `--pane`) lost the prompt. That is tasks 14 to 20: the pane stayed at
  `ctx 0k`. Each verifier spawn on Pi with
  `opencode-go/deepseek-v4.1-flash` got the prompt. A worker respawn for
  18-task-reopen at about 18:52, which reopened the workspace with
  `worktree.open`, got the prompt. A worker spawn with `--pane` into a
  workspace from `herdr workspace create` (t16-worker-2, about 18:10) lost it.
- Logs: `.hive/logs/2026-10-19.jsonl` in the home. Spawn code:
  `internal/agent/spawn.go`, `internal/agent/spawn_worker.go`
  (`promptTaskAgent`, `promptStarted`, `waitAgentInteractive`),
  `internal/agent/shell_ready.go`. herdr API: `reference/herdr/`.

## Rules

- Do live spawns only in a scratch hive home (`hive --home DIR`) or a scratch
  herdr workspace that you create and close. Do not send text to a pane that
  hive or the user started.
- Close each scratch workspace and stop each scratch agent after the checks.
  Make sure that `herdr workspace list` does not show them.

## Report

Write the report to the output path. It contains:

1. The cause, with evidence: log lines, code lines, and live-test output. Or
   state what stays unknown.
2. Each command that you ran and its output. Cap each output at 40 lines.
3. One or more fixes, each with its cost and risk.

## Finish

Run `hive task complete`. Send the report path and a three-line summary to the
manager with `hive inbox send --to manager-hive TEXT --agent NAME`.
