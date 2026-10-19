---
title: "An agent spawn can skip the first prompt"
status: "Open"
opened: 2026-10-19
decision: "A report task investigates first. The fix is decided after the report."
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
