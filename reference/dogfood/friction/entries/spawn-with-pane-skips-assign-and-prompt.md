---
title: "A worker spawn with --pane starts the agent but assigns nothing and sends no first prompt"
status: "Open"
opened: 2026-10-23
decision: ""
---

**Summary:**
On 2026-10-23, the manager of qmeter created the spec task `127-dash-spec`
for the goal `g7-dash-clock-cache`. `hive worktree create` made the workspace
`wV0` with root pane `wV0:p1`. The manager then ran

```sh
hive agent spawn --role worker --task 127-dash-spec --harness claude \
  --model claude-opus-5-5 --pane wV0:p1
```

The command printed `agent t127-worker-1 live` and nothing else. The agent
record was live, but the Claude pane stayed at an empty prompt for more than
three minutes, the task record kept `status: created` and an empty
`acting_agent`, and no first prompt arrived. The manager sent the worker
prompt by hand with `hive agent send`, and the worker did the task and pushed
`eb62ada`. Its `hive task complete` then failed with
`not the acting agent: t127-worker-1`. `hive task assign --task ... --agent
t127-worker-1` set the acting agent, and the worker completed.

The two later worker spawns of the same session that omitted `--pane` (tasks
`128-dash-cache` and `130-dash-clock`) each printed `assigned task` and
`prompted agent` and needed no repair. The verifier spawns without `--pane`
also prompted at once. So `--pane` is the branch that skipped both steps,
and the output did not say so.

**Reproduction steps:**
1. `hive worktree create --task ID`. Note the root pane of the new workspace.
2. `hive agent spawn --role worker --task ID --harness claude --model MODEL
   --pane PANE` with that root pane.
3. The output ends with `agent NAME live`. The agent stays idle at an empty
   prompt, `hive task show ID` keeps `status: created` and an empty
   `acting_agent`, and `hive task complete` fails with `not the acting agent`.
4. Workaround: `hive task assign --task ID --agent NAME`, then
   `hive agent send --agent NAME "<worker prompt>"`. Or omit `--pane`, which
   assigns and prompts in one step.
