---
title: "An agent spawn can skip the first prompt"
status: "Resolved 2026-10-19"
opened: 2026-10-19
decision: "Fixed in 28-first-prompt-fix (642fe78d): the first prompt uses the herdr prompt wait until working or blocked, with one retry on a stall. The user accepts the risk of a duplicate prompt."
---

**Summary:**
A released-shell event could change the new agent record to
`exited` before `agent.start` returned. The agent then ran without its first
prompt. The `starting` status keeps the supervisor from treating this event as
an exit. Fix commit: `c7b2566`.

On 2026-10-19, the manager spawned worker `t11-worker-1` with
`hive agent spawn --role worker` and no `--pane`. The spawn printed
`prompted agent` and `agent t11-worker-1 live`. Pi did not get the first
prompt. The pane stayed idle at context 0k. The workaround was
`herdr pane send-text`, then `herdr pane send-keys Enter`.

**Merged text from the `agent_not_ready` entry:**
On 2026-10-19, `hive agent send` to a Pi agent spawned with
`--pane` failed with `agent_not_ready` two times. This task records the issue.
It does not fix it.

**Report 2026-10-19:**
Report: `.hive/tmp/reports/first-prompt-skip.md` in the home (task
21-first-prompt-report). Cause: `promptStarted` in
`internal/agent/spawn_worker.go` accepts `idle` as proof of delivery before it
compares `state_change_seq` and `revision`. For all four wave 4 worker spawns,
the state after `agent.prompt` was `idle` with the same seq and revision as the
baseline, so hive logged `prompted agent` and skipped the retry. herdr
`agent.prompt` success proves only that the text and Enter were written. Why Pi
dropped those prompts stays unknown. A live test showed that an accepted prompt
can also look like an unchanged `idle` at first, so an immediate retry can send
a duplicate. Proposed fixes: use the herdr prompt wait as delivery evidence and
retry after a bounded stall result; or accept `idle` only after a seq or
revision change, with a bounded observation window.

**Probe 2026-10-19:**
Report: `.hive/tmp/reports/prompt-wait-probe.md` in the home (task
26-prompt-wait-probe). Each case ran 3 times with
`herdr agent prompt PANE TEXT --wait --until working --until blocked --timeout 10000`.
A Pi prompt into a new `herdr worktree create` pane returned
`agent_prompt_stalled` after 5.3 s in 3 of 3 runs, with no prompt in the pane.
One fresh Pi in a plain scratch workspace also stalled (1 of 3). Each retry
returned `working` and showed the prompt one time. Fast-finish prompts returned
`working`, then `done`, with no false stall. A Claude trust dialog returned
`agent_blocked` in 0.001 s and typed nothing. Claude never dropped a prompt.
Verdict: option 1 (the herdr prompt wait as delivery evidence, then a retry
after a stall) was safe in all 24 runs. It is not proven for an accepted prompt
that stays idle for more than 5 s; a retry could then send a duplicate.

**Resolution 2026-10-19:**
Fixed in 28-first-prompt-fix, landed at 642fe78d. Each spawn sends its first
prompt with the `agent.prompt` wait until `working` or `blocked` and a bounded
timeout. On `agent_prompt_stalled`, hive retries one time. A second stall fails
the spawn and names the agent and the cause. On `agent_blocked`, hive sends no
more text. The unchanged-`idle` delivery check is gone. The live check spawned
three Pi workers into new worktrees, and each one got its first prompt with no
hand resend.

Accepted risk: when an accepted prompt stays idle longer than the herdr wait
window, hive retries and the agent can get the prompt two times. The user
accepts this risk. Hive has no pane-read guard against it.

**Reproduction steps:**
1. Spawn a verifier with `--pane`.
2. Read the verifier pane.
3. Check that the first prompt is absent.
4. Send the prompt with `hive agent send`.
5. On 2026-10-19, spawn worker `t11-worker-1` with
   `hive agent spawn --role worker` and no `--pane`.
6. Check the output for `prompted agent` and `agent t11-worker-1 live`.
7. Read the Pi pane. It is idle at context 0k and has no first prompt.
8. Run `herdr pane send-text`, then `herdr pane send-keys Enter`.

Later on 2026-10-19, the first prompt was absent again for `t12-worker-1`,
`t11-verifier-1` (Claude, with `--pane`), `t11-worker-2` (with `--pane`), and
`t13-worker-1`. Each spawn printed `agent NAME live`. The first prompt arrived
for `t11-verifier-2` and `t12-verifier-1`, both Pi verifiers with no
`--pane`. The same workaround applied each time.

**Reproduction steps from the `agent_not_ready` entry:**
1. Spawn a Pi agent with `hive agent spawn --pane`.
2. Run `hive agent send` for that agent.
3. The command returns `agent_not_ready` (observed two times on 2026-10-19).
