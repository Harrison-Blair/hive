# T12 land

| Item | Value |
| --- | --- |
| Task ID | T12-land |
| Wave | 4 |
| Tier | decent |
| Depends on | T03, T09, CP2 |
| Blocks | CP3, T21 |
| Branch | `task/T12-land` |
| Worktree | `projects/hive/worktrees/T12-land/` |
| Evidence | `evidence/T12-land/` |

## Overview

This task gives `hive land`, the delivery of a verified task branch. The mode
`merge` merges the branch into the base branch of the registered repository
without fast-forward. The mode `pr` pushes the branch and opens a pull request
with `gh`. A landing cites a grant. Hive records the grant and enforces nothing.
One landing at a time runs for a repository.

## User flows

1. A manager runs `hive land --task 12-spawn-worker --mode merge --grant "merge into dev, grants.md line 3"`. Hive merges `task/12-spawn-worker` into `dev` in the repository, pushes `dev`, and prints the merge commit.
2. The manager runs `hive land --task 13-x --mode pr --evidence https://example.com/evidence/13-x`. Hive pushes the branch and prints the pull request URL.
3. The merge has a conflict. Hive aborts the merge, leaves the repository clean, and lists the conflicting files.
4. The base branch is not the current branch of the repository. Hive fails with `base branch not checked out: dev` and changes nothing.
5. The push after the merge fails. Hive ends with `partial`, the merge as an effect, and the commit in the record. The manager runs the command again. Hive keeps the commit and repeats only the push.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `cmd/root.go` | Register the `land` command. |
| `cmd/land/land.go` | New. |
| `internal/land/doc.go`, `land.go`, `merge.go`, `pr.go`, `render.go` | New. |
| `internal/lib/gitx/gitx.go` | Add `MergeNoFF`, `MergeAbort`, `MergeInProgress`, `HasRemote`, `Push`, `Tip`. |
| `internal/lib/run/run.go` | Change `Command` to honor the caller context deadline. Keep the 30-second default. |
| `internal/lib/store/namedlock.go` | New. A second `flock` file under `.hive/state/locks/<name>` for one long operation. |
| `internal/lib/store/doc.go` | Add the map line for `namedlock.go`. |
| `internal/lib/records/task.go` | Add the `delivery` fields: `mode`, `target`, `grant`, `commit`, `pushed`, `pr_url`, `evidence`, `at`. |

## Requirements

### AC1 Preconditions

- `hive land --task ID --mode merge|pr [--grant TEXT] [--evidence URL|PATH] [--draft]` needs a task with the status `verified`. Otherwise it fails with `task is <status>: land needs verified`.
- The task needs a `branch` in its worktree fields. The target is the base branch of the project.
- Without `--grant` hive logs a warning `no grant cited` and continues. The record stores an empty grant.
- The landing holds the named lock `land-<project>` from the first check to the record write. A second landing for the same project waits for it. The store lock covers only the record writes.
- A task whose `delivery` holds a `commit` is a retry. A retry keeps `commit` and repeats only the missing effects. These are the push in the mode `merge`, or the push and the pull request in the mode `pr`. A retry with another `--mode` fails with `task delivered in mode <mode>`.

### AC2 Merge

- In the registered repository: `CurrentBranch` must equal the base, else `base branch not checked out: <base>`. `IsDirty` must be false, else `repository is dirty`.
- `git merge --no-ff --no-edit <branch>` runs with a deadline of 60 seconds. On success hive records `mode`, `target`, `grant`, `evidence`, `commit` from `git rev-parse HEAD`, and `at`.
- When the repository has a remote `origin`, hive then runs `git push origin <base>` with a deadline of 120 seconds and sets `pushed` true. A failed push gives `partial` with the merge as an effect and `pushed` false. Without `origin` hive skips the push and says so.
- The 60-second merge bound and the 120-second push bound use `context.WithTimeout` at the call site. `run.Command` uses the caller context deadline and applies 30 seconds only when the context has no deadline. A test in `internal/lib/run/` covers the honored deadline and the earlier caller cancellation.
- A merge with exit status not 0 and `MergeInProgress` true is a conflict. Hive runs `git merge --abort` and fails with `merge conflict: <files>`. The repository is clean after it.
- A merge with another error, or on the deadline, runs `git merge --abort` when `MergeInProgress` is true. The outcome is `failed` with the git output. A failed abort gives `failed` with both messages and the effect `merge in progress`.
- `MergeInProgress` checks `git rev-parse -q --verify MERGE_HEAD`.

### AC3 Pull request

- `git push -u origin <branch>` runs in the repository with a deadline of 120 seconds. On success hive records `mode`, `target`, `grant`, `evidence`, `commit` from `Tip(branch)`, `pushed` true, and `at`.
- Hive writes the body to a file under `.hive/tmp/` and runs `gh pr create --base <base> --head <branch> --title <task title> --body-file <file>`. `--draft` passes through. Hive removes the file after the call, on each outcome.
- The body is the first paragraph of the brief. With `--evidence`, the body ends with the line `Evidence: <value>`.
- Hive records `pr_url` from the output of `gh`. A failure of `gh` gives `partial` with the push as an effect.
- On a retry without `pr_url`, hive first runs `gh pr view <branch> --json url`. A pull request that exists gives its URL without a new one.

### AC4 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Each precondition fails with its message and changes nothing, landings run one at a time, and a retry repeats only the missing effects. | V1, V2 |
| AC2 | A merge lands without fast-forward, pushes the base, records the delivery, and leaves the repository clean after each failure. | V2, V4 |
| AC3 | The pull request mode pushes, calls `gh` with the arguments and the body, and records the URL. | V3 |
| AC4 | The gate passes. | V5 |

## Verification items

### V1 Precondition tests

- Proves: AC1.
- Command: `go test ./internal/land/ -run TestPreconditions -v`, red and green.
- Expected: each case fails with its message. The cases: a task that is not verified, a task without a branch, and a base that is not the current branch. More cases: a dirty repository, and a retry with another mode. Two landings in two goroutines on one repository run one after the other, and each task records its own commit.

### V2 Merge tests

- Proves: AC1, AC2.
- Command: `go test ./internal/land/ -run TestMerge -v` and `go test ./internal/lib/run/ -run TestCommandHonorsCallerDeadline -v`, red and green.
- Expected: a real repository in a temporary directory with a base branch, a task branch, and a bare repository as `origin`. The merge gives a commit with two parents, and `origin` has the same base commit afterwards. A repository without `origin` lands without a push. A conflicting branch gives the error, and `git status --porcelain` is empty afterwards. A remote path that does not exist gives `partial` with `pushed` false and the commit in the record. A retry after it keeps the commit, pushes, and sets `pushed` true. A pre-merge hook that exits with 1 gives `failed` with the hook output and a clean repository. The 60-second merge bound and the 120-second push bound use `context.WithTimeout` at the call site. `run.Command` uses the caller context deadline and applies 30 seconds only when the context has no deadline. The runner test covers the honored deadline and the earlier caller cancellation.

### V3 Pull request tests

- Proves: AC3.
- Command: `go test ./internal/land/ -run TestPR -v`, red and green.
- Expected: a shell script named `gh` on a temporary `PATH` records its arguments and the content of the body file, and prints a URL. A bare repository as `origin` receives the push. The record has the URL, `pushed` true, and the other delivery fields. The body holds the first paragraph of the brief and the `Evidence:` line. `--draft` appears in the arguments. A `gh` script that exits with 1 gives `partial` with the push as an effect. A retry with a `gh pr view` that prints a URL records that URL and runs no `gh pr create`. The body file is gone after each case.

### V4 Live merge

- Proves: AC2.
- Use one unique fixture home with the marker `.hive/README.md`, one unique
  fixture repository, and one absolute tested binary. Record the binary SHA256.
- Command:

```sh
FIX=$(mktemp -d)
mkdir -p "$FIX/.hive"
: > "$FIX/.hive/README.md"
BIN="$FIX/hive-$(date +%s)"
go build -o "$BIN" .
sha256sum "$BIN"
REPO="$FIX/repo-land"
git init -q "$REPO"
git -C "$REPO" commit -q --allow-empty -m init
BASE=$(git -C "$REPO" branch --show-current)
HIVE_HOME="$FIX" "$BIN" project register "$REPO" --base "$BASE"
HIVE_HOME="$FIX" "$BIN" task create --project repo-land --title "land demo"
HIVE_HOME="$FIX" "$BIN" worktree create --task 1-land-demo
git -C "$FIX/projects/repo-land/worktrees/1-land-demo" commit -q --allow-empty -m work
HIVE_HOME="$FIX" "$BIN" agent register --role worker --project repo-land --task 1-land-demo --pane w0:p1
HIVE_HOME="$FIX" "$BIN" agent register --role verifier --project repo-land --task 1-land-demo --pane w0:p2
HIVE_HOME="$FIX" "$BIN" task assign --task 1-land-demo --agent worker-repo-land-1
HIVE_HOME="$FIX" "$BIN" task complete --task 1-land-demo --agent worker-repo-land-1
HIVE_HOME="$FIX" "$BIN" task verify --task 1-land-demo --agent verifier-repo-land-2
HIVE_HOME="$FIX" "$BIN" land --task 1-land-demo --mode merge --grant "demo"
echo "exit $?"
git -C "$REPO" log --oneline -n 3
HIVE_HOME="$FIX" "$BIN" task show 1-land-demo
HIVE_HOME="$FIX" "$BIN" worktree remove --task 1-land-demo
herdr workspace close <repository workspace ID from the create output>
rm -rf "$FIX"
```

- Expected: the land prints the merge commit and says that the repository has no `origin`. The exit status is 0. The log shows the merge commit, `work`, and `init`. The task shows the delivery with the commit. The remove succeeds because the merge guard passes. The fixture cleanup removes only `$FIX`.

### V5 Gate

- Proves: AC4.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- Enforcement of grants. The manager reads `projects/<id>/overrides/grants.md`.
- A merge of `dev` into `main`. The owner does that through a pull request.

## Changes

2026-10-18. Add CP3 as a dependent. Replace the live transcript with a unique
fixture home, a unique fixture repository, and an absolute tested binary.
2026-10-18. Add the D7 deadline contract. The 60-second merge bound and the
120-second push bound use `context.WithTimeout` at the call site. `run.Command`
honors the caller context deadline and applies 30 seconds only when the context
has no deadline. Add the runner test. Plan D7 supersedes the handoff constraint
to keep `Command` at 30 seconds.
