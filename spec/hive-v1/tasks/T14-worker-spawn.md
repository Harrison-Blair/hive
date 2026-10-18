# T14 worker spawn

| Item | Value |
| --- | --- |
| Task ID | T14-worker-spawn |
| Wave | 4 |
| Tier | decent |
| Depends on | T09, T10, T11, CP2 |
| Blocks | CP3, T21 |
| Branch | `task/T14-worker-spawn` |
| Worktree | `projects/hive/worktrees/T14-worker-spawn/` |
| Evidence | `evidence/T14-worker-spawn/` |

## Overview

This task spawns a worker or a verifier into the worktree of its task without
`--pane`. Hive uses the root pane that `worktree create` recorded, checks that
the pane still belongs to the worktree, starts the harness, assigns the task,
and sends the first prompt. The brief itself stays in the store. The prompt
tells the agent to read the skill of its role and then its brief with
`hive task show --brief`. T21 writes the skills. This task owns the prompt text.

The stop of the worker closes its pane, and with it the workspace of the
worktree. The verifier spawn therefore reopens the worktree with
`worktree.open`, or makes a new tab when the user kept the workspace open.

## User flows

1. A manager runs `hive agent spawn --role worker --task 12-spawn-worker --harness claude --model <decent model>`. The worker starts in the pane of the worktree workspace, and the first prompt tells it to read the skill and its brief.
2. The manager runs the same command for a task without a worktree. Hive fails with `task has no worktree: run hive worktree create --task 12-spawn-worker`.
3. After the worker completes and the supervisor stops it, the manager runs `hive agent spawn --role verifier --task 12-spawn-worker --harness codex`. The workspace of the worktree closed with the worker. Hive reopens the worktree as a new workspace, and the verifier starts in its root pane.
4. The worker record is still `live`. The verifier spawn fails with `worker still live: worker-hive-3`.
5. The worker shows a trust dialog at the start. Hive keeps the task unassigned, ends with `partial`, and prints the two commands that finish the spawn by hand.
6. The user kept a tab open in the workspace of the worktree. The verifier spawn makes a new tab there and starts the verifier in it.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/agent/spawn.go` | Add one case line for the worker role and one for the verifier role in the pane-choice switch. |
| `internal/agent/spawn_worker.go` | New. The worker and verifier path: the task checks, the pane check, the project from the task, the assignment, and the prompt. |
| `internal/agent/brief.go` | New. The two prompt texts and the shell quoting. |
| `internal/agent/doc.go` | Add the map lines. |
| `internal/lib/herdr/worktree.go` | Add `WorktreeOpen`. T09 made the file with `WorktreeCreate`. |
| `internal/lib/herdr/tab_create.go` | New. `TabCreate`. T13 makes `tab.go` at the same time, so this function has its own file. |
| `internal/lib/herdr/doc.go` | Add the map lines. |

## Requirements

### AC1 Worker

- With `--role worker --task ID` and no `--pane`, the task must have the status `created` or `assigned` and a worktree with a `root_pane`. Otherwise the spawn fails with the message of the user flows.
- `--project` can be absent. The project comes from the task. A `--project` that differs from the project of the task fails with `task <id> is in project <p>`.
- Pane check: hive calls `pane.get` on the recorded `root_pane`. It compares the current pane ID with the recorded `root_pane` before the worker start. A difference fails with `worktree pane mismatch: pane_id`.
- The pane must exist, its `workspace_id` must equal the recorded `workspace_id` of the worktree, and its `cwd` must equal the worktree path after symbolic link resolution. It must have no `restore_error`, the status `unknown`, and no `agent`.
- Check the repository identity and checkout path through the real workspace metadata. Read the branch through `gitx.CurrentBranch`, or a typed `WorktreeInfo` result that holds a branch. Do not use a workspace branch field.
- A pane that is gone, moved, or in another directory fails with `worktree pane mismatch: <field>`. A pane with `restore_error` fails with `pane not restored: <error>`.
- A pane with an agent is busy. Hive calls `agent.get` with the pane ID and fails with `pane busy: <name>`. Without a name, it fails with `pane busy: <kind> agent` with the kind from `pane.get`.
- A record with the status `unresolved` and the same pane blocks the spawn with `pane has an unresolved record: <name>`. T20 gives `hive agent resolve` for it.
- The sequence: the spawn core of T10 with the pane, then the assignment, then `agent.prompt` with the worker prompt and no wait. Before the assignment, hive reads the task again under the store lock and checks the status again. A changed status fails with `task is <status>` and leaves the agent live without a task.
- A lost start reply or another uncertain start preserves the topology and the record. Report `partial` and the unresolved record. Do not clean up after an uncertain start.
- The result of the core decides the next steps. `ok` with the status `idle` or `done`: assign, then prompt. `blocked`: no assignment and no prompt. The outcome is `partial` with the error of T10 and the hint `then run: hive task assign --task <id> --agent <name> and hive agent send --agent <name> '<prompt>'`. A timeout: no assignment and no prompt, `partial` as T10 gives. `failed`, or an `unresolved` record: no assignment and no prompt.
- A prompt that fails after the assignment gives `partial` with the effects `agent started` and `task assigned`, and the hint `then run: hive agent send --agent <name> '<prompt>'`.
- The worker prompt is one line: `Read <home>/.agents/skills/hive-worker/SKILL.md first. Then run: hive --home '<home>' task show '<task-id>' --brief. Follow that brief. Your worktree is '<path>' on the branch '<branch>'. The skill hive-worker is /hive-worker in Claude Code and $hive-worker in Codex.`
- Each value inside the single quotes of the prompt is shell-safe. Hive replaces each `'` in the value with `'\''`. A pasted command then works in `sh` with a path that holds a space or a quote.

### AC2 Verifier

- With `--role verifier --task ID`, the task must have the status `completed`. Each record of the task with the role `worker` must have the status `stopped` or `exited`. A `live` record fails with `worker still live: <name>`. An `unresolved` record fails with `pane has an unresolved record: <name>`.
- Hive calls `workspace.get` with the recorded `workspace_id` of the worktree. The error `workspace_not_found` means that the stop of the worker closed the workspace. Hive then sends `worktree.open` with `cwd` equal to the registered repository path, `path` equal to the worktree path, `label` equal to the task ID, and `focus` false. The reply gives the workspace, the tab, and the root pane. Hive writes `workspace_id`, `tab_id`, and `root_pane` to the worktree of the task.
- A found workspace still holds panes of the user. Hive then sends `tab.create` with the `workspace_id`, `cwd` equal to the worktree path, and `focus` false. Hive writes `tab_id` and `root_pane` from the reply.
- Hive tracks whether this invocation made the tab or the workspace. When a start fails before it becomes successful, close only newly made, still-owned topology. Do not close a user workspace to remove one tab. Do not treat `already_open` as ownership.
- Hive permits cleanup only after a known no-start result and a fresh ownership check. A lost start reply or an uncertain start preserves the topology and the record. Report `partial` and the unresolved record.
- The pane rules of AC1 then apply to that root pane.
- The verifier prompt is one line: `Read <home>/.agents/skills/hive-verifier/SKILL.md first. Then run: hive --home '<home>' task show '<task-id>' --brief. You are the verifier. Run each verification item again, write VERIFIED.md, then run: hive --home '<home>' task verify --task '<task-id>', or add --fail --reason. The skill hive-verifier is /hive-verifier in Claude Code and $hive-verifier in Codex.`
- The spawn does not assign the task. The record has the role `verifier` and the task.
- The verifier spawn uses the same core result rules as AC1, without the assignment.

### AC3 Record

- The agent record has the project from the task, the task, the harness, the model, the pane, and the `workspace_id` and `tab_id` of the worktree. The core of T10 fills `terminal_id` and `cwd`. `quiet_after_min` comes from `--quiet-after`, else 10.

### AC4 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The worker spawn checks the task and the pane, assigns only after a settled start, and sends the prompt. | V1, V2, V3 |
| AC2 | The verifier spawn checks the task and the worker records, reopens the worktree or makes a tab, and sends its prompt. | V1, V2, V3 |
| AC3 | The record has the fields from the task and the flags. | V1 |
| AC4 | The gate passes. | V4 |

## Verification items

### V1 Spawn path tests

- Proves: AC1, AC2, AC3.
- Command: `go test ./internal/agent/ -run 'TestWorkerSpawn|TestVerifierSpawn' -v`, red and green.
- Expected: a fake herdr server. Refusal cases: no worktree, another project, a pane in another workspace, and a pane with another `cwd`. Add a moved pane whose current ID differs from `root_pane`. More refusal cases: a pane with `restore_error`, a busy pane with a name, and a busy pane without a name. The last refusal case: an unresolved record on the pane. Each refusal sends no start. Sequence cases: start, wait, assign, prompt in this order for `idle` and for `done`. The blocked case has no assignment, no prompt, and the hint. The timeout case has no assignment. A task that changes status during the start gets no assignment. A failed prompt after the assignment gives `partial` with both effects. The verifier cases: the status rule, a `live` worker record, and no assignment. More verifier cases: `worktree.open` after `workspace_not_found` with the worktree fields updated. Also `tab.create` for an open workspace, a real branch from `gitx.CurrentBranch` or a typed `WorktreeInfo`, and `already_open` with no ownership. Add a topology-write failure, a lost start reply that keeps the topology, and cleanup only after a known no-start result. The record fields with the new workspace.

### V2 Prompt text tests

- Proves: AC1, AC2.
- Command: `go test ./internal/agent/ -run TestBriefPrompt -v`, red and green.
- Expected: each prompt is one line and starts with the `Read` sentence with the skill path of its role. It holds the home, the task ID, both skill invocations, and for the worker the path and the branch. A test writes a script named `hive` on a temporary `PATH` that prints its arguments. It runs the command part of the prompt with `sh -c` for a home path with a space and a `'`. The script receives the home path and the task ID unchanged.

### V3 Live transcript

- Proves: AC1, AC2.
- Use one unique fixture home with the marker `.hive/README.md`, one unique
  fixture repository, and one absolute tested binary. Put fixture skills only in
  `$FIX/.agents/skills`. Do not replace a real home skill. Record the binary
  SHA256 and the actual path that the worker reads.
- Command, from a pane of the hive session:

```sh
FIX=$(mktemp -d)
mkdir -p "$FIX/.hive" "$FIX/.agents/skills/hive-worker" "$FIX/.agents/skills/hive-verifier"
: > "$FIX/.hive/README.md"
printf '%s\n' '# hive-worker fixture' 'Follow the brief.' > "$FIX/.agents/skills/hive-worker/SKILL.md"
printf '%s\n' '# hive-verifier fixture' 'Follow the brief.' > "$FIX/.agents/skills/hive-verifier/SKILL.md"
BIN="$FIX/hive-$(date +%s)"
go build -o "$BIN" .
sha256sum "$BIN"
REPO="$FIX/repo-w"
git init -q "$REPO" && git -C "$REPO" commit -q --allow-empty -m init
"$BIN" --home "$FIX" project register "$REPO" --base "$(git -C "$REPO" branch --show-current)"
"$BIN" --home "$FIX" task create --project repo-w --title "worker demo" --brief - <<B
Make the file hello.txt in your worktree with the one line hello. Do not commit.
Verification item V1: cat hello.txt prints hello.
When V1 passes, run: hive --home '$FIX' task complete --task 1-worker-demo
B
"$BIN" --home "$FIX" worktree create --task 1-worker-demo
W="t14-check-w-$(date +%s)"
"$BIN" --home "$FIX" agent spawn --role worker --task 1-worker-demo --name "$W" --harness claude --model {{T14_SMOKE_MODEL}}
echo "exit $?"
"$BIN" --home "$FIX" task show 1-worker-demo
# wait until the task shows completed, then:
cat "$FIX/projects/repo-w/worktrees/1-worker-demo/hello.txt"
"$BIN" --home "$FIX" agent stop --agent "$W"
herdr workspace list
V="t14-check-v-$(date +%s)"
"$BIN" --home "$FIX" agent spawn --role verifier --task 1-worker-demo --name "$V" --harness claude --model {{T14_SMOKE_MODEL}}
echo "exit $?"
herdr workspace list
# wait until the task shows verified, then:
"$BIN" --home "$FIX" task show 1-worker-demo
"$BIN" --home "$FIX" agent stop --agent "$V"
"$BIN" --home "$FIX" worktree remove --task 1-worker-demo --force
herdr workspace close <repository workspace ID from the create output>
rm -rf "$FIX"
```

- Expected: both spawns exit with 0. The first task show gives the status `assigned` and the acting agent `$W`. The file holds `hello`. The first workspace list has no workspace `1-worker-demo`, and the second has one again. The second task show gives the status `verified` and a `completed` and a `verified` inbox record for `manager-repo-w`. The cleanup leaves no workspace of this check. The fixture cleanup removes only `$FIX`.

### V4 Gate

- Proves: AC4.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- The stop of a finished worker. T15 does it.
- The choice of harness and model from the presets. The manager reads the presets and passes the flags.
- The text of the skills `hive-worker` and `hive-verifier`. T21 writes them.

## Changes

2026-10-18. Add the current pane check, the real branch source, the invocation
ownership rule, the unknown-start preservation, and the cleanup rule. Replace
the live transcript with a unique fixture home, fixture skills, and an absolute
tested binary.
