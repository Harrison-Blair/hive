# Hive

Hive is a small Go CLI that starts coding agents in
[herdr](https://herdr.dev) panes and lets them talk to each other. It keeps a
record of each agent, opens shared workspaces on Git branch checkouts and
directories, and sends prompts between agents.

Hive keeps its state in the home: `$HIVE_HOME` when that variable is set and
not empty, otherwise `~/.hive`. Hive creates the home when it first needs it.
Hive works from any directory and needs no setup step.

```text
<home>/
├─ worktrees/<key>/worktree.json     ownership record of a checkout that hive made
├─ worktrees/<key>/<branch>/         that checkout
├─ agents/<name>.json                agent records
├─ logs/YYYY-MM-DD.jsonl             log events
├─ coordinator-choice.json           last picker choice
├─ extensions/hive-inbox.ts          pi extension of hive agent send
├─ tmp/                              scratch files
├─ lock                              store lock
├─ layout.lock                       lock of spawn, stop, worktree, and coordinator names
└─ schema.json                       home schema version
```

## Requirements

- Linux on amd64 or arm64.
- [herdr](https://herdr.dev) 0.9.5 (protocol 22).
- Git.
- At least one harness: `claude`, `codex`, or `pi`. The picker lists models
  with `pi --list-models` (pi and codex) and `opencode models` (claude). When
  that fails, type the model name.
- Go 1.27.2 or newer, only to build hive from source with `go install`.

## Install

Replace `vX.Y.Z` with the release version, for example `v0.1.0`.

1. Get the binary. Use the release archive, or build it with
   `go install`.

   From the release archive (use `arm64` in place of `amd64` on arm64):

   ```sh
   cd "$(mktemp -d)"
   curl -fsSLO https://github.com/Harrison-Blair/hive/releases/download/vX.Y.Z/hive_vX.Y.Z_linux_amd64.tar.gz
   curl -fsSLO https://github.com/Harrison-Blair/hive/releases/download/vX.Y.Z/checksums.txt
   sha256sum --check --ignore-missing checksums.txt
   tar -xzf hive_vX.Y.Z_linux_amd64.tar.gz hive
   ```

   Or with Go:

   ```sh
   go install github.com/Harrison-Blair/hive@vX.Y.Z
   ```

2. Put the binary on your `PATH`. For the archive:

   ```sh
   install -m 0755 hive ~/.local/bin/hive
   ```

   `go install` writes the binary to `$(go env GOPATH)/bin`. Add that directory
   to your `PATH` if it is not there.

3. Check the binary.

   ```sh
   hive version
   ```

Then run `hive` in a herdr pane to start the coordinator in that directory.

## Update

1. Replace the binary with the binary of the new tag. Use the archive steps of
   [Install](#install), step 1 and step 2, or run:

   ```sh
   go install github.com/Harrison-Blair/hive@vX.Y.Z
   ```

2. Restart each running agent. A running agent keeps the old binary until it
   stops.

### Worktrees of an older hive

An older hive made agent worktrees at `<home>/projects/<id>/worktrees/<name>/`
on branch `<name>`. This hive has no projects and does not migrate them. Those
checkouts have no ownership record, so `hive worktree remove` refuses them.
They keep working as plain Git checkouts: `hive agent spawn --directory PATH`
starts an agent in one. Clean them up by hand with Git:

1. Stop each agent in the checkout with `hive agent stop N`, and close the
   herdr workspace on it.
2. Check the checkout. Commit or move each change, untracked file, and ignored
   file that you want to keep:

   ```sh
   git -C <home>/projects/<id>/worktrees/<name> status --porcelain --ignored
   ```

3. Remove the checkout from the repository. Git deletes the ignored files in
   it. The branch stays; delete it with `git branch -d <name>` when you no
   longer need it:

   ```sh
   git -C <repository> worktree remove <home>/projects/<id>/worktrees/<name>
   ```

   When the directory is already gone, `git -C <repository> worktree prune`
   removes the entry that Git still keeps for it.

4. When no checkout is left below `<home>/projects/`, delete that directory.
   This hive does not read it.

## Usage

Run each command in a herdr pane. Each command takes `--json` for one JSON
result object.

| Command | Behavior |
| --- | --- |
| `hive` | Pick a harness (claude, codex, or pi) and a model, then run it in this pane and in the current directory as the coordinator. Its name comes from the current directory: `hive-coordinator` in `~/src/Hive`, `my-app-coordinator` in `~/src/my.app`. A name that has no letters or digits gives `coordinator`. When a live agent in another pane has the name, hive adds a number: `hive-2-coordinator`, or `coordinator-2`. The last choice is preselected. Type to filter, or pick `Other…` to type a model. |
| `hive agent spawn --name N --harness H --model M [--prompt T \| --prompt-file F]` | Start an agent in a new tab of the caller's workspace, and send it the prompt. |
| `hive agent spawn ... --branch B [--base REF]` | Start the agent in the checkout of branch B of the repository of the current directory. When no checkout has B, hive makes one below `<home>/worktrees/`: a new branch starts at `REF`, or at the current commit. `--base` is an error for a branch that exists. Hive opens the repository parent and checkout with `worktree.open`, so herdr groups linked checkout workspaces beneath their parent. Each agent gets a new tab in the checkout workspace. |
| `hive agent spawn ... --directory D` | Start the agent in the existing directory D. The agents of one directory share a herdr workspace: the workspace open on the checkout, or the workspace of a pane whose shell is in that directory. When there is none, the agent takes the root pane of a new workspace. |
| `hive agent send N TEXT [--wait]` | Send `⬢ SENDER: TEXT` to agent N through the native message path of its harness. It never types into the pane. With `--wait`, wait until the agent is idle or done. |
| `hive agent stop N` | Close the pane of agent N and remove its record. The other agents of its workspace keep running. The checkout and the branch stay. |
| `hive agent list` | List the agent records with their herdr status. |
| `hive agent graph` | Show a live graph of the agents in coordinator and home pods. With `--json`, print one snapshot, the same as `hive agent list --json`. |
| `hive worktree create B [--base REF]` | Print the checkout of branch B of the repository of the current directory. When no checkout has B, make one, as `spawn --branch` does. |
| `hive worktree remove B` | Remove the checkout of branch B that hive made, and close the herdr workspaces on it. Hive refuses a checkout that it did not make, a checkout with changes, untracked files, or ignored files, and a checkout with a live agent. Stop each other process that writes into the checkout first: a file written during the removal can be lost. The branch stays. |
| `hive logs` | Read the daily log files in `<home>/logs/`. |
| `hive guide` | Print the full agent-friendly command reference and usage notes. Agents start with a compact reference to the most-used agent commands and the same notes. |
| `hive version` | Print the hive version. |

## Links

- [AGENTS.md](AGENTS.md): the instructions for agents that work in this
  repository.
- [reference/herdr/](reference/herdr/): the herdr API reference.
