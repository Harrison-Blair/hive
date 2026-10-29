# Hive

Hive is a small Go CLI that starts coding agents in
[herdr](https://herdr.dev) panes and lets them talk to each other. It keeps a
record of each agent, opens project workspaces and Git worktrees, and sends
prompts between agents.

The home is a checkout of this repository. Hive builds hive: this checkout is
the home and the first registered project.

## Requirements

- Linux on amd64 or arm64.
- [herdr](https://herdr.dev) 0.9.5 (protocol 22).
- Git.
- At least one harness: `claude`, `codex`, or `pi`. The picker lists models
  with `pi --list-models` (pi and codex) and `opencode models` (claude). When
  that fails, type the model name.
- Go 1.27.2 or newer, only to build hive from source with `go install`.

## Install

Replace `vX.Y.Z` with the release version, for example `v0.1.0`. The binary and
the home must use the same version.

1. Clone the repository at the tag. The clone is your home.

   ```sh
   git clone --branch vX.Y.Z https://github.com/Harrison-Blair/hive.git ~/hive
   ```

2. Get the binary of the same tag. Use the release archive, or build it with
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

3. Put the binary on your `PATH`. For the archive:

   ```sh
   install -m 0755 hive ~/.local/bin/hive
   ```

   `go install` writes the binary to `$(go env GOPATH)/bin`. Add that directory
   to your `PATH` if it is not there.

4. Run `hive init` in the clone. It writes the home path to
   `~/.config/hive/config.toml`.

   ```sh
   cd ~/hive
   hive init
   hive version
   ```

Then run `hive` in a herdr pane in the home to start the coordinator.

## Update

Move the home and the binary to the same new tag. In these steps, `HOME_DIR` is
your home, for example `~/hive`.

1. Fetch the tags.

   ```sh
   git -C HOME_DIR fetch --tags
   ```

2. Check out the new tag.

   ```sh
   git -C HOME_DIR checkout vX.Y.Z
   ```

3. Replace the binary with the binary of the same tag. Use the archive steps of
   [Install](#install), step 2 and step 3, or run:

   ```sh
   go install github.com/Harrison-Blair/hive@vX.Y.Z
   ```

4. Restart each running agent. A running agent keeps the old binary until it
   stops.

## Usage

Run each command in a herdr pane. Each command takes `--json` for one JSON
result object.

| Command | Behavior |
| --- | --- |
| `hive` | In the home: pick a harness (claude, codex, or pi) and a model, then run it in this pane as the `coordinator`. The last choice is preselected. Type to filter, or pick `Other…` to type a model. |
| `hive agent spawn --name N --harness H --model M [--prompt T \| --prompt-file F]` | Start an agent in a new tab of the caller's workspace, and send it the prompt. |
| `hive agent spawn ... --project P` | Start the agent in a new tab of the checkout workspace of project P. Hive uses the open workspace on the checkout, or creates one. |
| `hive agent spawn ... --project P --worktree NAME` | Create the worktree `projects/P/worktrees/NAME` on branch `NAME` from the base branch of P, and start the agent in its workspace. |
| `hive agent send N TEXT [--wait]` | Send `⬢ SENDER: TEXT` to agent N. With `--wait`, wait until the agent is idle or done. |
| `hive agent stop N [--remove-worktree]` | Close the pane or the worktree workspace of agent N. The checkout and the branch stay. `--remove-worktree` removes a clean checkout and refuses a dirty one. On an agent without a worktree, it closes the pane, removes the record, and prints `no worktree for N`. |
| `hive agent list` | List the agent records with their herdr status. |
| `hive project register PATH [--id ID] [--base BRANCH]` | Register a Git repository as a project. |
| `hive project list` | List the registered projects. |
| `hive init` | Write the home path to `~/.config/hive/config.toml`. |
| `hive logs` | Read the daily log files in `.hive/logs/`. |
| `hive version` | Print the hive version. |

## Links

- [AGENTS.md](AGENTS.md): the instructions for agents that work in this
  repository.
- [reference/herdr/](reference/herdr/): the herdr API reference.
