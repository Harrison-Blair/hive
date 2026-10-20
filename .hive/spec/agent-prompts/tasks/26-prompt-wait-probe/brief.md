# herdr prompt-wait probe

| Item | Value |
| --- | --- |
| Wave | report only |
| Tier | decent |
| Depends on | none |
| Grant | none. No code change, no commit, no landing. |
| Output | `/home/penguin/source/hive/.hive/tmp/reports/prompt-wait-probe.md` |

## Overview

This task is a probe for the friction entry
`reference/dogfood/friction/entries/agent-spawn-skips-first-prompt.md`. It
tests whether `herdr agent prompt --wait` can prove that a first prompt
reached the agent. Read `/home/penguin/source/hive/.hive/tmp/reports/first-prompt-skip.md`
first. Do not change a source file. Do not commit. Do not push. This brief
replaces the commit, push, gate, and evidence steps of the worker skill. The
report file is the only output. Do not edit a file in
`reference/dogfood/friction/`.

## Scratch resources

- Make a scratch git repository in
  `/home/penguin/source/hive/.hive/tmp/probe-prompt-wait/` with one commit.
- Make a scratch herdr workspace with `--no-focus`. Do all tests there.
- Do not send text to a pane that hive or the user started.
- At the end, close each scratch workspace, remove each scratch worktree, and
  remove `.hive/tmp/probe-prompt-wait/`. Make sure that `herdr workspace list`
  and `git worktree list` do not show them.

## Cases

Run each case 3 times. Use this command form:
`herdr agent prompt PANE TEXT --wait --until working --until blocked --timeout 10000`.
Read `reference/herdr/` for the exact flags first. When a flag differs in
herdr 0.9.5, record the difference and use the documented form.

1. Normal prompt to a fresh Pi agent. Expected: `working` fast.
2. Fast finish: prompt `reply only OK`. Record whether herdr still observes
   `working`, or returns a false `agent_prompt_stalled`.
3. The failing case: `herdr worktree create` on the scratch repository, start
   Pi in the root pane of the new worktree workspace, and prompt at once.
   Expected: `agent_prompt_stalled` in about 5 s when the prompt drops.
4. Retry after case 3: the prompt arrives one time, with no duplicate. Read
   the pane to check.
5. An agent in a blocked dialog. Expected: `agent_blocked`, and nothing typed.
6. Cases 1 to 3 on Claude (`claude`, model `claude-opus-5-5`).

## Report

Write the report to the output path. It contains:

1. Each command, its output (cap at 40 lines), and its timing.
2. A table: case, run, harness, result, time, and expected or not.
3. A verdict: is option 1 of `first-prompt-skip.md` safe alone? Give the
   reason, and each case that would break it.

## Finish

Run `hive task complete`. Send the report path and the verdict to the manager
with `hive inbox send --to manager-hive TEXT --agent NAME`.
