# Dogfooding friction log

This file records the friction that agents meet when they use hive in this repository: bugs, missing capabilities, and workarounds found in real work.
The coordinator and a manager can edit this file directly. They can add or extend an entry, set its status to `Resolved YYYY-MM-DD`, and keep the index table in sync.
They edit only `friction.md` in this way. They commit on `dev` with the subject `docs(dogfood): SUMMARY` and push `dev`. They use no worker or verifier for this edit.
For all other files, the coordinator and a manager never change a project source file.

## Entry template

### Title

**Issue:** Short title

**Status:** Open

**Summary:** What happened, and how often it occurred.

**Reproduction steps:**
1. ...
2. ...

| Issue | Status |
| --- | --- |
| [No hive path repairs a verified task](#no-hive-path-repairs-a-verified-task) | Open |
| [Any agent can read and consume the inbox of another role](#any-agent-can-read-and-consume-the-inbox-of-another-role) | Open |
| [A verify fail leaves no workspace and no verifier for a re-verify](#a-verify-fail-leaves-no-workspace-and-no-verifier-for-a-re-verify) | Open |
| [A worker spawn into an assigned task keeps the old acting agent](#a-worker-spawn-into-an-assigned-task-keeps-the-old-acting-agent) | Open |
| [An idle session burns tokens on supervisor nudges](#an-idle-session-burns-tokens-on-supervisor-nudges) | Open |
| [`hive agent send` fails with `agent_not_ready` after a Pi `--pane` spawn](#hive-agent-send-fails-with-agent_not_ready-after-a-pi---pane-spawn) | Open |
| [An agent spawn can skip the first prompt](#an-agent-spawn-can-skip-the-first-prompt) | Open |
| [A verifier spawn with `--pane` requires `--project`](#a-verifier-spawn-with---pane-requires---project) | Open |
| [An exited agent name blocks reuse](#an-exited-agent-name-blocks-reuse) | Open |
| [No `hive task edit --brief` command](#no-hive-task-edit---brief-command) | Open |
| [Full Pi worktree, agent, and delivery bootstrap flow is missing](#full-pi-worktree-agent-and-delivery-bootstrap-flow-is-missing) | Open |
| [Shell-ready race in worker and verifier spawns](#shell-ready-race-in-worker-and-verifier-spawns) | Resolved 2026-10-19 |
| [Manager spawn fails with `agent_pane_busy` right after `workspace.create`](#manager-spawn-fails-with-agent_pane_busy-right-after-workspacecreate) | Resolved 2026-10-19 |

## Open

### No hive path repairs a verified task

**Issue:** No hive path repairs a verified task

**Status:** Open

**Summary:** On 2026-10-19, `hive land --task 12-friction-log --mode merge`
failed with a merge conflict in `testdata/ste-baseline.json`. Task 11 landed
first, and each task added one baseline line. The task had the status
`verified`. `hive agent spawn --role worker` refused it with `task is
verified`. `hive task assign` refused it with `invalid transition: verified ->
assign`. The workaround was a Pi agent that the manager started by hand with
`herdr agent start` in a new tab of the task workspace. That agent merged
`origin/dev` into the task branch, kept both lines, ran the gate, and pushed.
Then `hive land` passed. The repair has no hive agent record and no second
verification record.

**Reproduction steps:**
1. Make two tasks that each add one line at the same place in a file.
2. Verify both tasks. Land the first task.
3. Run `hive land` for the second task. The merge fails with a conflict.
4. Run `hive agent spawn --role worker` for the second task. Hive returns
   `task is verified`.
5. Run `hive task assign` for the second task. Hive returns
   `invalid transition: verified -> assign`.

### Any agent can read and consume the inbox of another role

**Issue:** Any agent can read and consume the inbox of another role

**Status:** Open

**Summary:** On 2026-10-19, verifier `t12-verifier-1` ran
`hive inbox --role manager-hive` at 12:46:30. That command marked record 66
(`verified` for 12-friction-log) as read. The supervisor then did not nudge
the manager, and the landing waited until the coordinator saw the record.
`hive inbox --role` accepts any role from any agent. The workaround is a
sentence in each brief that tells the agent to read only its own inbox.

**Reproduction steps:**
1. Spawn a verifier for a task of project `hive`.
2. In the verifier pane, run `hive inbox --role manager-hive`.
3. Check that the command prints and marks the unread manager records.
4. Check that the supervisor sends no nudge to the manager for them.

### A verify fail leaves no workspace and no verifier for a re-verify

**Issue:** A verify fail leaves no workspace and no verifier for a re-verify

**Status:** Open

**Summary:** On 2026-10-19, `t11-verifier-1` recorded a fail for
11-ci-workflow. Then the verifier stopped, and the supervisor closed the
worker workspace `wCK` when its last agent finished. After the repair,
`hive agent send` to the verifier returned `agent is stopped`, and
`herdr tab create --workspace wCK` returned `workspace_not_found`. The
workaround was a fresh verifier spawn with no `--pane`, which made the new
workspace `wCN` on the same worktree.

**Reproduction steps:**
1. Spawn a worker and a verifier for a task in the same workspace.
2. Let the verifier record a fail and finish.
3. Repair the task and complete it again.
4. Run `hive agent send` to the verifier. Hive returns `agent is stopped`.
5. Run `herdr tab create --workspace ID` for the task workspace. herdr
   returns `workspace_not_found`.

### A worker spawn into an assigned task keeps the old acting agent

**Issue:** A worker spawn into an assigned task keeps the old acting agent

**Status:** Open

**Summary:** On 2026-10-19, the manager spawned `t11-worker-2` with
`hive agent spawn --role worker --pane` to repair 11-ci-workflow after a
verify fail. The task still named `t11-worker-1` as the acting agent.
`hive task complete` from `t11-worker-2` failed with `not the acting agent`.
The workaround was `hive task assign --task 11-ci-workflow --agent
t11-worker-2`, then `hive task complete`.

**Reproduction steps:**
1. Let a worker finish a task. Let the verifier record a fail.
2. Spawn a second worker for the same task with `--pane`.
3. In the second worker, run `hive task complete`.
4. Check that hive returns `not the acting agent`.

### An idle session burns tokens on supervisor nudges

**Issue:** An idle session burns tokens on supervisor nudges

**Status:** Open

**Summary:** On 2026-10-19, with all tasks landed and both managers on hold,
the supervisor kept the coordinator awake. Each idle manager writes a `quiet`
record to the coordinator inbox every ten minutes. The supervisor types a
nudge into the coordinator pane for each unread record. Each nudge costs one
coordinator turn, which costs tokens, and a user who is away pays for it. The
coordinator's own file watch doubled the cost before it was stopped. The
workaround was to stop both managers and the supervisor by hand. Hive has no
hush mode: a state in which the supervisor holds routine kinds (`quiet`,
`completed`, `verified`) and nudges only for `blocked`, `exited`,
`unresolved`, and urgent messages. The firstmate `/quiet` mode is the model.

**Reproduction steps:**
1. Spawn two managers with no goals. Leave the coordinator idle.
2. Wait more than ten minutes.
3. Each manager writes a `quiet` record. The supervisor types
   `hive: 1 new inbox record, run: hive --home ... inbox` into the coordinator
   pane for each one. The coordinator spends a turn on each nudge.

### `hive agent send` fails with `agent_not_ready` after a Pi `--pane` spawn

**Issue:** `hive agent send` fails with `agent_not_ready` after a Pi `--pane` spawn

**Status:** Open

**Summary:** On 2026-10-19, `hive agent send` to a Pi agent spawned with
`--pane` failed with `agent_not_ready` two times. This task records the issue.
It does not fix it.

**Reproduction steps:**
1. Spawn a Pi agent with `hive agent spawn --pane`.
2. Run `hive agent send` for that agent.
3. The command returns `agent_not_ready` (observed two times on 2026-10-19).

### An agent spawn can skip the first prompt

**Issue:** An agent spawn can skip the first prompt

**Status:** Open

**Summary:** A released-shell event could change the new agent record to
`exited` before `agent.start` returned. The agent then ran without its first
prompt. The `starting` status keeps the supervisor from treating this event as
an exit. Fix commit: `c7b2566`.

On 2026-10-19, the manager spawned worker `t11-worker-1` with
`hive agent spawn --role worker` and no `--pane`. The spawn printed
`prompted agent` and `agent t11-worker-1 live`. Pi did not get the first
prompt. The pane stayed idle at context 0k. The workaround was
`herdr pane send-text`, then `herdr pane send-keys Enter`.

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

### A verifier spawn with `--pane` requires `--project`

**Issue:** A verifier spawn with `--pane` requires `--project`

**Status:** Open

**Summary:** The manager skill shows a verifier spawn without `--project`. A
verifier spawn with `--pane` requires `--project`.

**Reproduction steps:**
1. Read the verifier spawn command in `hive-manager/SKILL.md`.
2. Add `--pane PANE`. Do not add `--project`.
3. Run the command.
4. Check that hive requires `--project`.

### An exited agent name blocks reuse

**Issue:** An exited agent name blocks reuse

**Status:** Open

**Summary:** An exited agent name blocks a later spawn with `agent exists`.
Each retry needs a new agent name.

**Reproduction steps:**
1. Spawn an agent with the name `NAME`.
2. Let the agent exit.
3. Spawn another agent with the name `NAME`.
4. Check that hive returns `agent exists`.

### No `hive task edit --brief` command

**Issue:** No `hive task edit --brief` command

**Status:** Open

**Summary:** The manager changed a task brief by hand in
`.hive/state/briefs/` on 2026-10-18. hive has no `hive task edit --brief`
command. The manager must edit the stored brief by hand.

**Reproduction steps:**
1. Run `hive task --help`.
2. Check that the command list has no `edit` command.
3. Edit a task brief in `.hive/state/briefs/` by hand.

### Full Pi worktree, agent, and delivery bootstrap flow is missing

**Issue:** Full Pi worktree, agent, and delivery bootstrap flow is missing

**Status:** Open

**Summary:** The full Pi worktree, agent, and delivery bootstrap flow still
lacks T09, T10, T12, and T18. The coordinator dispatches work by hand with
`herdr worktree` and `herdr agent` commands. This manual dispatch is the current
workaround. No runtime task or agent record simulates a missing capability.

**Reproduction steps:**
1. Run `herdr worktree --help`. It lists `list`, `create`, `open`, and `remove`.
   The hive `worktree` command does not exist.
2. Run `herdr agent --help`. It lists `start`, `prompt`, `wait`, and related
   commands. The hive `agent spawn`, `agent send`, and `agent stop` commands do
   not exist.
3. The safe probe output is in
   `.hive/tmp/coordinator/reports/later-fix-land-01-captures/herdr-help-probes.txt`.

## Resolved

### Shell-ready race in worker and verifier spawns

**Issue:** Shell-ready race in worker and verifier spawns

**Status:** Resolved 2026-10-19

**Summary:** A worker or verifier spawn in a fresh worktree workspace could hit
`agent_pane_busy`. Commit `aa29b8d9` adds the shell-ready wait before the shared
`agent.start` call. This resolves the race for both roles.

**Reproduction steps:**
1. Before commit `aa29b8d9`, create a fresh worktree workspace.
2. Spawn a worker or verifier in its root pane.
3. Check the log for `agent.start` before the shell is ready.
4. Check for `agent_pane_busy`.

### Manager spawn fails with `agent_pane_busy` right after `workspace.create`

**Issue:** Manager spawn fails with `agent_pane_busy` right after `workspace.create`

**Status:** Resolved 2026-10-19

**Summary:** On 2026-10-18 the first `hive agent spawn --role manager --project
hive --harness claude --model claude-opus-5-5` in a home with no manager
workspace failed with `first tab busy: stop the agent or pass --pane`. The log
shows `workspace.create` at 22:34:09.993 and `agent.start` on the new root pane
at 22:34:10.003, ten milliseconds later. herdr answered `agent_pane_busy`:
`agent target pane wBD:p1 is not an available shell`. The shell of the new
pane was not ready. The spawn kept the workspace `wBD` and its project record,
and it marked the agent record `exited`. A second identical spawn one minute
later adopted `wBD` and started the manager. The error text names a wrong cause:
no agent held the pane. It occurred on each first spawn of a project manager:
twice in two attempts on 2026-10-18, for `hive` and for `stenographer`. The
retry succeeded each time. During rc-file startup, the shell runs commands in
its own process group. Thus, a group ID equal to the shell PID is not sufficient.
Commit `aa29b8d9` also requires one foreground process whose PID is the shell
PID.

**Reproduction steps:**
1. Register a project with no manager workspace record.
2. Run `hive agent spawn --role manager --project ID --harness claude --model claude-opus-5-5`.
3. Read `.hive/logs/<date>.jsonl`: `agent.start` follows `workspace.create` within a few milliseconds and herdr returns `agent_pane_busy`.
4. Run the same spawn again: it adopts the workspace and succeeds.
