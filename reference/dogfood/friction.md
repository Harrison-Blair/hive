# Dogfooding friction log

This file records the friction that agents meet when they use hive itself in
this repository: bugs, missing capabilities, and workarounds found in real work.
Append each new entry at the end with the template below. An entry has the
status `Open` until a fix lands, then `Resolved YYYY-MM-DD` with the fix in its
summary. A resolved entry keeps its full text as a record.

## Entry template

**Issue:** Short title

**Status:** Open

**Summary:** What happened, and how often it occurred.

**Reproduction steps:**
1. ...
2. ...

## Entries

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

**Issue:** No `hive task edit --brief` command

**Status:** Open

**Summary:** The manager changed a task brief by hand in
`.hive/state/briefs/` on 2026-10-18. hive has no `hive task edit --brief`
command. The manager must edit the stored brief by hand.

**Reproduction steps:**
1. Run `hive task --help`.
2. Check that the command list has no `edit` command.
3. Edit a task brief in `.hive/state/briefs/` by hand.

**Issue:** An exited agent name blocks reuse

**Status:** Open

**Summary:** An exited agent name blocks a later spawn with `agent exists`.
Each retry needs a new agent name.

**Reproduction steps:**
1. Spawn an agent with the name `NAME`.
2. Let the agent exit.
3. Spawn another agent with the name `NAME`.
4. Check that hive returns `agent exists`.

**Issue:** A verifier spawn with `--pane` requires `--project`

**Status:** Open

**Summary:** The manager skill shows a verifier spawn without `--project`. A
verifier spawn with `--pane` requires `--project`.

**Reproduction steps:**
1. Read the verifier spawn command in `hive-manager/SKILL.md`.
2. Add `--pane PANE`. Do not add `--project`.
3. Run the command.
4. Check that hive requires `--project`.

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

**Issue:** A verifier spawn with `--pane` skips the first prompt

**Status:** Resolved 2026-10-19

**Summary:** A released-shell event could change the new agent record to
`exited` before `agent.start` returned. The agent then ran without its first
prompt. The `starting` status keeps the supervisor from treating this event as
an exit. Fix commit: `c7b2566`.

**Reproduction steps:**
1. Spawn a verifier with `--pane`.
2. Read the verifier pane.
3. Check that the first prompt is absent.
4. Send the prompt with `hive agent send`.

**Issue:** `hive agent send` fails with `agent_not_ready` after a Pi `--pane` spawn

**Status:** Open

**Summary:** On 2026-10-19, `hive agent send` to a Pi agent spawned with
`--pane` failed with `agent_not_ready` two times. This task records the issue.
It does not fix it.

**Reproduction steps:**
1. Spawn a Pi agent with `hive agent spawn --pane`.
2. Run `hive agent send` for that agent.
3. The command returns `agent_not_ready` (observed two times on 2026-10-19).
