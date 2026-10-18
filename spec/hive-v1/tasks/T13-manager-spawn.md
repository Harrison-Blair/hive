# T13 manager spawn

| Item | Value |
| --- | --- |
| Task ID | T13-manager-spawn |
| Wave | 4 |
| Tier | decent |
| Depends on | T10, CP2 |
| Blocks | T21, CP3, CP5 |
| Branch | `task/T13-manager-spawn` |
| Worktree | `projects/hive/worktrees/T13-manager-spawn/` |
| Evidence | `evidence/T13-manager-spawn/` |

## Overview

This task gives the manager its workspace. `hive agent spawn --role manager`
without `--pane` finds or makes the workspace `hive:<project>`, writes the
project instructions, and starts the harness in the root pane of the first tab.
`hive agent stop` closes the pane of the manager. herdr closes the workspace
when that pane was its last pane, and the tabs of the user stay otherwise. A
live manager claim is what "active project" means.

Hive reuses only the workspace whose ID the project record stores. A workspace
with the same label that hive does not know has no owner. The user closes it or
passes `--adopt` with its ID.

## User flows

1. The coordinator runs `hive agent spawn --role manager --project hive --harness claude --model <strong model>`. A workspace `hive:hive` appears with the project directory as its working directory, and the manager starts in its first tab.
2. The coordinator runs the same command a day later. The project record names the workspace, and it is still open with the label and the directory. Hive reuses it and starts the manager in the root pane of its first tab, if that pane holds a shell.
3. The user made a workspace `hive:hive` by hand. The spawn fails with `workspace w5 has the label hive:hive but hive did not make it: close it or pass --adopt w5`. The coordinator runs the spawn again with `--adopt w5`, and hive records the workspace as not made by hive.
4. The coordinator runs `hive agent stop --agent manager-hive`. Hive closes the manager pane. It was the only pane, so the workspace is gone and the project record forgets it.
5. The user opened a second tab in `hive:hive`. The stop closes the manager pane. The workspace stays with the tab of the user, and the project record keeps the workspace.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/agent/spawn.go` | Add one case line for the manager role in the pane-choice switch. |
| `internal/agent/workspace.go` | New. Find, check, adopt, or make the workspace. Choose the root pane of the first tab. |
| `internal/agent/stop.go` | Update `workspace` on the project record from the result of the stop function. |
| `internal/agent/instructions.go` | New. Write `projects/<id>/AGENTS.md` from the embedded template. |
| `internal/agent/templates/project-agents.md` | New. The template, in ASD-STE100 Strict. T21 refines it. |
| `internal/agent/doc.go` | Add the map lines. |
| `cmd/agent/spawn/spawn.go` | Add `--adopt WORKSPACE`. |
| `internal/lib/herdr/workspace.go` | Add `WorkspaceCreate`. T09 made the file with `WorkspaceGet`, `WorkspaceList`, and `WorkspaceClose`. |
| `internal/lib/herdr/tab.go` | New. `TabList`. |
| `internal/lib/herdr/types.go` | Add the optional `RestoreError` field to `PaneInfo` as a `*string`. Keep each other field. |
| `internal/lib/herdr/pane.go` | Add `PaneList` and add the optional `restore_error` field to `paneInfoWire`. Decode it into `PaneInfo.RestoreError`. An absent value or null gives nil. A value of another type fails the decode. Keep each other field. |
| `internal/lib/herdr/doc.go` | Add the map lines. |
| `internal/lib/records/project.go` | Add `workspace` with `id`, `root_pane`, `created_by_hive`, and `created_at`. |
| `ste_test.go` | Add the template to `steFiles`. |
| `testdata/ste-baseline.json` | Add the entry of the template with 0. |

## Requirements

### AC1 Workspace

- With `--role manager` and no `--pane`, hive reads `workspace` from the project record. With an ID, hive calls `workspace.get` for the workspace and `pane.get` for the recorded root pane. The workspace must exist with the label `hive:<project>`. The `cwd` of the recorded root pane must equal the project directory after symbolic link resolution. Then hive reuses it.
- The recorded workspace can be gone, or its label or its directory can differ. Then hive clears `workspace` on the project record and continues as without an ID. It logs the reason.
- A workspace that exists with the label and directory but has no recorded root keeps its metadata. Hive refuses with `workspace <id> has no recorded root pane: pass --pane <shell pane>`. A moved root that returns a different current pane ID gets the same refusal. Hive does not select a shell by itself and does not close a user tab.
- Without an ID, hive lists the workspaces and keeps those with the label `hive:<project>`. One match or more has no owner. Hive fails with `workspace <id> has the label hive:<project> but hive did not make it: close it or pass --adopt <id>`. The message names the first match. No match means hive makes one.
- `--adopt WORKSPACE` takes a workspace without an owner. The workspace must exist with the label, and its first tab must hold exactly one pane whose `cwd` is the project directory. Otherwise the spawn fails with `first tab has <n> panes: pass --pane` or `workspace cwd mismatch`. Hive records the workspace with `created_by_hive` false and the pane as `root_pane`.
- Hive makes the directory `<home>/projects/<project>` when it does not exist. The request is `workspace.create` with `cwd` equal to that directory, `label` equal to `hive:<project>`, and `focus` false.
- herdr falls back to the home directory of the user when `cwd` does not exist. Hive compares `root_pane.cwd` of the reply with the requested path after symbolic link resolution. On a difference hive closes the new workspace and fails with `workspace cwd mismatch`.
- Hive records `workspace` on the project with the ID, the root pane, `created_by_hive` true, and `created_at` for a new workspace. `created_by_hive` belongs to that ID. A workspace with another ID never inherits it.

### AC2 First tab

- The pane is the `root_pane` of a new workspace. For a reused or adopted workspace, hive calls `tab.list` with the workspace ID. The first tab is the first element of the list, because `number` is a creation counter and not a position. The pane is the recorded `root_pane`, which must be in that tab.
- The pane must be an available shell: `pane.get` shows no `restore_error`, the status `unknown`, and no `agent`. Otherwise the spawn fails with `first tab busy: stop the agent or pass --pane`. The error `agent_pane_busy` from `agent.start` gives the same message.
- Hive never makes a tab for the manager. The first tab is never an empty shell after the spawn.
- When a check or the record write fails after `workspace.create` and before `agent.start`, hive closes the new workspace and clears `workspace` on the project record.

### AC3 Instructions

- Before the start, hive writes `projects/<project>/AGENTS.md` from the embedded template, with the project ID, the home path, and the repository path filled in. It replaces the file on each spawn. The first line says that hive generates the file.
- Before the write, hive checks the static type of the project directory, each ancestor, and the instruction destination. It rejects a symbolic link and a non-regular destination.
- The template gives the manager role gate, the commands of the manager, and the rule to delegate each source change to a worker. T21 refines it.

### AC4 Stop

- `hive agent stop --agent NAME` runs the stop function of T10. For a manager record, hive then updates the project record from the post-close result.
- A `closed` result clears `workspace` on the project record and adds the effect `workspace closed`.
- A `surviving` result keeps `workspace` on the project record and adds the effect `workspace kept: <n> other panes`. The next manager spawn reuses the workspace when its first tab holds a shell.
- An `unknown` result keeps `workspace` on the project record and reports the known `pane closed` effect. Do not infer closure from an old pane count.
- A project-record write failure after the close is `partial` with the known effects.
- Hive sends no `workspace.close` on a stop, and it never passes `close_group` true.
- The stop function checks the process identity. A pane and terminal match is not an ownership check.

### AC5 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The spawn reuses only the recorded workspace, refuses one without an owner unless `--adopt`, and makes one with the project directory. | V1, V4 |
| AC2 | The manager starts in the root pane of the first tab, or the spawn fails when that pane is not a shell. | V1, V4 |
| AC3 | The instructions file comes from the template with the values filled in. | V2, V4 |
| AC4 | The stop closes the manager pane, and the project record forgets the workspace only when that pane was its last one. | V3, V4 |
| AC5 | The gate passes. | V5 |

## Verification items

### V1 Workspace and pane tests

- Proves: AC1, AC2.
- Command: `go test ./internal/agent/ -run 'TestManagerWorkspace|TestFirstTab' -v`, red and green.
- Expected: a fake herdr server scripts each case, and the test checks each request parameter and each error. Reuse cases: a recorded workspace with the label and the directory, a recorded workspace that is gone, and a recorded workspace with another directory. More reuse cases: a workspace with no recorded root, a moved root with a new pane ID, and a user tab that stays. Cases without an owner: one match and two matches without `--adopt`, `--adopt` with one pane, and `--adopt` with two panes. Create cases: an empty list, a cwd mismatch with a close, and a record failure after the create with a close. Pane cases: a reordered tab list, a root pane with `restore_error`, a root pane with an agent, and `agent_pane_busy`. A `closed` stop clears the workspace. A `surviving` stop with a user tab keeps it. An `unknown` stop keeps the metadata. A project-record failure after the close is `partial`.

### V2 Instructions tests

- Proves: AC3.
- Command: `go test ./internal/agent/ -run TestInstructions -v`, red and green.
- Expected: the file exists after the spawn, holds the project ID and the paths, and starts with the generated line.

### V3 Stop tests

- Proves: AC4.
- Command: `go test ./internal/agent/ -run TestStopWorkspace -v`, red and green.
- Expected: a fake herdr server. A stop with a post-close `closed` result clears `workspace` on the project record and has the effect `workspace closed`. A stop with a post-close `surviving` result keeps `workspace` and has the effect `workspace kept: 1 other panes`. Neither case sends `workspace.close`. An `unknown` result keeps `workspace` and reports the `pane closed` effect. A project-record failure after the close is `partial`.

### V4 Live transcript

- Proves: AC1, AC2, AC3, AC4.
- Use one unique fixture home with the marker `.hive/README.md`, one unique
  fixture repository, and one absolute tested binary. Record the binary SHA256.
- Command, from a pane of the hive session:

```sh
FIX=$(mktemp -d)
mkdir -p "$FIX/.hive"
: > "$FIX/.hive/README.md"
BIN="$FIX/hive-$(date +%s)"
go build -o "$BIN" .
sha256sum "$BIN"
REPO="$FIX/repo-mgr"
git init -q "$REPO"
"$BIN" --home "$FIX" project register "$REPO"
herdr workspace list
"$BIN" --home "$FIX" agent spawn --role manager --project repo-mgr --name "t13-check-$(date +%s)" --harness claude --model {{T13_SMOKE_MODEL}}
echo "exit $?"
herdr workspace list
herdr tab list --workspace <workspace ID from the spawn output>
head -n 3 "$FIX/projects/repo-mgr/AGENTS.md"
"$BIN" --home "$FIX" agent stop --agent <name from the spawn output>
echo "exit $?"
herdr workspace list
rm -rf "$FIX"
```

- Expected: both commands exit with 0. The second list shows `hive:repo-mgr`, and the tab list shows one tab with the agent status of the manager. The file starts with the generated line. The last list is the same as the first list. The fixture cleanup removes only `$FIX`.

### V5 Gate

- Proves: AC5.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- The final text of the manager instructions. T21 writes it.
- A manager in a workspace that is not `hive:<project>`. The user passes `--pane`.

## Changes

2026-10-18. Keep the workspace metadata when the recorded root is missing or
moved. Add the explicit `--pane` refusal. Add the static path checks. Replace the
pane-count stop result with the T10 post-close result. Replace the live
transcript with a unique fixture home and an absolute tested binary.
2026-10-18. Add the D8 restore field. `PaneInfo` gains the optional
`RestoreError` as a `*string`. `paneInfoWire` gains `restore_error`, and the
decode gives nil for an absent value or null and fails on another type. Keep
each other field and the AC2 refusal.
