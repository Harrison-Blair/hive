# T09 worktree create and remove

| Item | Value |
| --- | --- |
| Task ID | T09-worktree-create-and-remove |
| Wave | 3 |
| Tier | decent |
| Depends on | T03, T05 |
| Blocks | CP2, T10, T12, T14 |
| Branch | `task/T09-worktree-create-and-remove` |
| Worktree | `projects/hive/worktrees/T09-worktree-create-and-remove/` |
| Evidence | `evidence/T09-worktree-create-and-remove/` |

## Overview

This task gives the worktree of a task. `hive worktree create` asks herdr to
make a linked worktree of the registered repository at the path that hive
chooses, and herdr opens it as a workspace. `hive worktree remove` checks the
identity of the worktree, closes its workspace when it is still open, and
removes the worktree with Git after the Git guards pass. The task also gives
`gitx`, the bounded Git helpers over the runner of T05, and the typed
workspace wrappers.

The first `worktree.create` for a repository that herdr does not know opens
two workspaces: one for the repository and one for the worktree. Hive reports
the repository workspace and records it on the project. Hive never closes it.

## User flows

1. A manager runs `hive worktree create --task 12-spawn-worker`. A workspace with the label `12-spawn-worker` appears, with one tab and one shell pane in the new directory. Hive prints the workspace ID, the root pane ID, the path, and the branch.
2. The repository had no workspace. The same command also opens a workspace for the repository. Hive prints `repository workspace opened: w7` as an effect.
3. The manager runs `hive worktree remove --task 12-spawn-worker` after the task landed. The workspace closes, the directory is gone, and the branch stays.
4. The manager runs the remove on a task with uncommitted changes. Hive refuses with `worktree is dirty`. With `--force` hive removes it.
5. The workspace is already closed, because the supervisor stopped the verifier. Hive skips the workspace close and removes the worktree with Git.
6. The user moved another checkout into the recorded workspace. Hive refuses with `worktree mismatch: checkout_path`, also with `--force`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `cmd/root.go` | Register the `worktree` command. |
| `cmd/worktree/worktree.go`, `create/create.go`, `remove/remove.go` | New. |
| `internal/worktree/doc.go`, `create.go`, `remove.go`, `render.go` | New. |
| `internal/lib/herdr/worktree.go` | New. `WorktreeCreate` with its parameter and result types. T14 adds `WorktreeOpen`. |
| `internal/lib/herdr/workspace.go` | New. `WorkspaceGet`, `WorkspaceList`, and `WorkspaceClose`. T13 adds `WorkspaceCreate`. |
| `internal/lib/herdr/doc.go` | Add the map lines. |
| `internal/lib/gitx/gitx.go` | New. `RefExists`, `BranchExists`, `IsDirty`, `IsMerged`, `CurrentBranch`, `IsLinkedWorktree`, `CommonDir`, `WorktreeRemove` over `run`. |
| `internal/lib/records/task.go` | Add the `worktree` fields: `workspace_id`, `tab_id`, `root_pane`, `path`, `branch`, `created_at`, `removed_at`. |
| `internal/lib/records/project.go` | Add `repo_workspace_id`, the workspace of the repository that herdr opened. |

## Requirements

### AC1 Create

- `hive worktree create --task ID [--focus]` needs a task with the status `created` or `assigned` and no worktree. A task with a worktree fails with `task has a worktree: <path>`.
- The path is `<home>/projects/<project>/worktrees/<task-id>`. The branch is `task/<task-id>`. The base is the base branch of the project.
- Before the request, hive runs four checks in this order. `IsLinkedWorktree(repo)` true fails with `repository is a linked worktree: register the primary checkout`. A missing base ref fails with `base ref not found: <base>`. An existing branch fails with `branch exists: <branch>`. An existing path fails with `path exists: <path>`. Each check sends no request and changes no record.
- Hive lists the workspaces before the request.
- The request is `worktree.create` with `cwd` equal to the repository path of the project, `branch`, `base`, `path`, `label` equal to the task ID, and `focus`. It never uses `workspace_id`. The client deadline of the request is 60 seconds through `herdr.WithDeadline` of T05.
- After the reply hive compares `worktree.path` with the requested path after symbolic link resolution. A difference is a `partial` outcome with the effects listed.
- Hive records `workspace_id`, `tab_id`, `root_pane`, `path`, `branch`, and `created_at` on the task.
- Hive lists the workspaces after the request. A new workspace other than the returned one is the repository workspace. Hive records its ID as `repo_workspace_id` on the project and lists `repository workspace opened: <id>` as an effect.
- The effects: workspace created, worktree created, branch created, and the repository workspace when it applies.
- A `worktree_create_failed` error gives `failed` with the herdr message. Hive then runs `BranchExists`. When the branch exists, the effects list `branch created`, and the message says that the branch stays.
- A `worktree_operation_in_progress` error gives `failed` with `worktree operation in progress: retry later`. No record changes.
- A client error after hive wrote the request, such as a deadline or a closed connection, gives `partial` with the phase `request`. Hive then inspects the branch, the path, and the workspace list. It records the fields that it finds on the task and lists each one as an effect.

### AC2 Remove

- `hive worktree remove --task ID [--force]` needs a task with a worktree.
- The task status must be `verified` or `cancelled`. Otherwise the command fails with `task is <status>: complete, verify, or cancel it first`.
- Identity check through Git: `IsLinkedWorktree(path)` must be true, `CommonDir(path)` must resolve into the repository, and `CurrentBranch(path)` must equal the recorded branch. A difference fails with `worktree mismatch: <field>`. `--force` does not skip this check.
- Identity check through herdr, when the record has a `workspace_id` and `workspace.get` finds it. The reply must have `worktree.is_linked_worktree` true and `worktree.checkout_path` equal to the recorded path. Its `worktree.repo_root` must equal the repository path. Hive compares each path after symbolic link resolution. A difference fails with `worktree mismatch: <field>` and closes nothing. The error `workspace_not_found` means that the workspace is already closed, and hive goes on.
- Git guards: `IsDirty(path)` true fails with `worktree is dirty`. `IsMerged(repo, base, branch)` false fails with `branch not merged into <base>`. `--force` skips both guards.
- When the workspace is open, hive sends `workspace.close` with the `workspace_id` and `close_group` false, and reports the effect `workspace closed`. A `workspace_group_close_required` error fails with `workspace is a group root: close it by hand`. The shells of the user in that workspace end with it, because the directory goes away.
- Then hive runs `git -C <repo> worktree remove [--force] <path>` through `gitx.WorktreeRemove` and reports the effect `worktree removed`. A Git failure after the close gives `partial` with the effect `workspace closed` and the Git output.
- Hive sets `removed_at` and clears `workspace_id`, `tab_id`, and `root_pane`. It keeps `path` and `branch` for history. The branch stays in the repository. Hive never sends `worktree.remove`.

### AC3 Git helpers

- Each helper runs `git` through `run.Command` with the deadline of the runner and parses the output.
- `IsDirty` uses `git status --porcelain`. `IsMerged` uses `git branch --merged <base>`. `RefExists` uses `git rev-parse --verify`. `BranchExists` uses `git branch --list`. `CurrentBranch` uses `git branch --show-current`. `IsLinkedWorktree` compares `git rev-parse --git-dir` with `git rev-parse --git-common-dir`. `CommonDir` returns the second value as an absolute path.

### AC4 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | `create` checks, sends the request with `cwd`, records the worktree and the repository workspace, and reports each failure. | V1, V3 |
| AC2 | `remove` checks the identity through Git and herdr, guards, closes an open workspace, removes the worktree with Git, and updates the record. | V2, V3 |
| AC3 | Each Git helper reads the real state of a repository. | V2 |
| AC4 | The gate passes. | V4 |

## Verification items

### V1 Create tests

- Proves: AC1.
- Command: `go test ./internal/worktree/ -run TestCreate -v`, red and green.
- Expected: a fake herdr server records the request and returns a result. The test checks each parameter and the record fields. Refusal cases: a task with a worktree, a linked worktree as the repository, and a missing base in a real repository. More refusal cases: an existing branch, and an existing path. Each refusal case sends no request and changes no record. Failure cases: a path difference, a repository workspace in the second list, and a closed connection after the request. Error cases: `worktree_create_failed` with a branch left behind, and `worktree_operation_in_progress`.

### V2 Remove and gitx tests

- Proves: AC2, AC3.
- Command: `go test ./internal/worktree/ ./internal/lib/gitx/ -run 'TestRemove|TestGitx' -v`, red and green.
- Expected: the gitx tests make a real repository with `git init` in a temporary directory and cover each helper, with a linked worktree for `IsLinkedWorktree` and `CommonDir`. The remove tests cover the status guard, the dirty guard, the merged guard, and `--force`. They cover the close of an open workspace and the skip of the close for a closed workspace. They cover the identity check with one mismatch for each field. The mismatch case sends no close and runs no removal, also with `--force`. A Git failure after the close gives `partial`.

### V3 Live transcript

- Proves: AC1, AC2.
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
REPO="$FIX/repo-live"
git init -q "$REPO"
git -C "$REPO" commit -q --allow-empty -m init
"$BIN" --home "$FIX" project register "$REPO" --base "$(git -C "$REPO" branch --show-current)"
"$BIN" --home "$FIX" task create --project repo-live --title "live demo"
herdr workspace list
"$BIN" --home "$FIX" worktree create --task 1-live-demo
herdr workspace list
"$BIN" --home "$FIX" worktree remove --task 1-live-demo
echo "exit $?"
"$BIN" --home "$FIX" task cancel --task 1-live-demo
"$BIN" --home "$FIX" worktree remove --task 1-live-demo
herdr workspace list
git -C "$REPO" branch
herdr workspace close <repository workspace ID from the create output>
rm -rf "$FIX"
```

- Expected: the second list shows `1-live-demo` and the repository workspace, and the create output names the repository workspace. The first remove fails with `task is created`. The second remove succeeds. The third list no longer shows `1-live-demo`. The branch list shows `task/1-live-demo`. After the close, the workspace list is the same as the first list. The fixture cleanup removes only `$FIX`.

### V4 Gate

- Proves: AC4.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- The guard on a live or unresolved agent in the worktree. T20 adds it.
- The spawn of an agent in the new pane. T14 adds it.
- The close of the repository workspace. The user closes it.

## Changes

2026-10-18. Add T10 to the blocks of T09, because `WorkspaceGet` lands in T09.
Replace the live transcript with a unique fixture home, a unique fixture
repository, and an absolute tested binary. Keep the behavioral criteria.
