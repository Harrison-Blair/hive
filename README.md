# Hive

Hive is a Go CLI and TUI that coordinates coding agents in
[herdr](https://herdr.dev) from one home. You talk to one coordinator. A
manager owns each project. Workers do one task each in their own checkout. An
independent agent verifies each task before it lands.

The home is a checkout of this repository. It holds the instructions, the
skills, and the specs that the agents read, so a hive binary needs a home of
the same version. Hive builds hive: this checkout is the home and the first
registered project.

## Requirements

- Linux on amd64 or arm64.
- [herdr](https://herdr.dev) 0.9.5 (protocol 22).
- Git.
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

Then run `hive` in a herdr pane to start the TUI, the coordinator, and the
supervisor.

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

4. Restart the supervisor and the coordinator from **Health › Agents** in the
   TUI. Select each agent and press `r`. They keep running the old binary until
   you restart them.

## Links

- [Hive v1 specification](.hive/spec/hive-v1/README.md): the goals, features,
  architecture, risks, and task plan.
- [AGENTS.md](AGENTS.md): the instructions for agents that work in this
  repository.
- [reference/](reference/): the herdr API reference, design patterns, and the
  dogfood friction log.
