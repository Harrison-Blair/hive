# Home config and hive init

| Item | Value |
| --- | --- |
| Wave | H1, task A of 2 |
| Tier | strong |
| Depends on | none. Base: `dev` at `68ae1dc9` or later |
| Grant | `grants.md 2026-10-19 H1 home config wave` |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

Hive finds the home from a global config file. `hive init` writes that file.
This task adds the config, `hive init`, and the new resolution order. It keeps
the `--home` flag. Task B of H1 removes the flag later. The plan is
`projects/hive/plans/home-spec-planning/MAP.md` in the home, decisions 1 and
2. The user agreed to it. Do not change a decision in this brief. When the code
does not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `home.Resolve(homeFlag, homeFlagSet, envHome, cwd)` in
  `internal/lib/home/home.go` resolves the home: the flag, then `HIVE_HOME`,
  then the outermost `.hive/README.md` marker above the working directory.
  Today `HIVE_HOME` fails with `home conflict` when a different marked home is
  above the working directory.
- The callers are `cmd/common/common.go` (`Run`), `cmd/logs/logs.go`, and
  `internal/tui/tui.go` (`Run`).
- A worktree of the hive repository has a tracked `.hive/README.md`. The outermost
  marker above a worktree below `projects/` is the home.
- `go test` runs each package with the package folder as the working
  directory. That folder is below a marked home.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/lib/home/` (new files allowed, for example `config.go`) and its tests | R1, R2. |
| A new command folder below `cmd/`, for example `cmd/homeinit/`, with `New() *cobra.Command`, and `cmd/root.go` to register it. Go does not accept `init` as a package name. | R3. |
| A new leaf package below `internal/` that mirrors the command, for example `internal/homeinit/`, with a `doc.go` when it needs one | R3. |
| `cmd/common/common.go`, `cmd/logs/logs.go`, `internal/tui/tui.go` | R2. Only the call to the new resolution. |
| Each existing test that must isolate the config to stay correct (R4) | R4. List each file in the evidence. |
| `.agents/skills/hive-coordinator/SKILL.md` | R5. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `spec/hive-v1/tasks/<task-id>.md` | R6. New. A copy of this brief. |

Do not edit other files.

## Requirements

### R1 Config file

- The config path is `filepath.Join(os.UserConfigDir(), "hive", "config.toml")`.
  On Linux this is `$XDG_CONFIG_HOME/hive/config.toml`, else
  `~/.config/hive/config.toml`.
- The file has one key: `home = "PATH"`, with an absolute path. Hive writes
  the file in this form. The CLI uses no new dependency: read and write this
  one key with the standard library. A file that hive cannot read, or with no
  valid `home` key, is an error that names the config path.

### R2 Resolution order

Resolve the home in this order:

1. The `--home` flag has first priority. Task B removes it later.
2. `HIVE_HOME`, when it has a value. It must name a marked home. It selects that
   home. It does not fail on a different marked home above the working
   directory. This removes the current `HIVE_HOME` conflict. Fixture tests in
   task B need it.
3. The config, when the file exists. The home in it must have the marker.
   When a marked home is above the working directory and it differs from the
   config home, fail with an error that starts with `home conflict:` and names
   both paths.
4. The outermost marker above the working directory.
5. Else fail with an error that says to run `hive init` in the home.

### R3 hive init

- `hive init` takes no arguments. It finds the outermost marked home above
  the working directory. It does not read the config or `HIVE_HOME` to find
  it. With no marker, it fails and writes nothing.
- When no config exists, it makes the config folder and writes the config with
  that home. When the config names the same home, it succeeds and changes
  nothing. When the config names a different home, it fails, writes nothing,
  and names both paths.
- It obeys the Command output rules of `AGENTS.md`: text by default, one JSON
  envelope with `--json`, exit 0 only on success. Report the write as an
  effect. Log the command as other commands do.

### R4 Tests

- Write each test first and see it fail for the correct reason.
- Cover each step of R2 and each case of R3 with a fixture config in a temp
  folder. Set `XDG_CONFIG_HOME` with `t.Setenv` in the test.
- No test reads or writes the real config of the user. Make each test that
  resolves a home isolate the config, for example with `XDG_CONFIG_HOME` set
  to a temp folder in the test or in a `TestMain`.
- Update the existing `HIVE_HOME` conflict test to the R2 behavior.

### R5 Coordinator skill

In `.agents/skills/hive-coordinator/SKILL.md`, add one startup step: when
`hive` fails with the message that says to run `hive init`, run `hive init` in
the home. Keep the text in ASD-STE100 Strict.

### R6 Brief copy

Copy this brief to `spec/hive-v1/tasks/<task-id>.md` with a baseline of 0.
When the copy causes a violation, change the words of that sentence and keep
its meaning. Record each change in the evidence.

## Not in scope

The removal of the `--home` flag and of `--home` in docs, skills, prompts, and
tests (task B). The spec layout (H2). `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each R4 test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Build the task binary. Set `XDG_CONFIG_HOME` to a new temp folder and unset `HIVE_HOME`. Run `hive init` from the worktree twice. Run `hive task list --project hive` from `/tmp`. Write config for another marked temp home. Run task list and init from the worktree with that config. Run task list from `/tmp` with no config. | First init writes config for `/home/penguin/source/hive` and exits 0. Second init exits 0 and does not change the config. The list from `/tmp` exits 0 and shows the home tasks. The other-home config returns an error that starts with `home conflict:`. Init returns an error and leaves the other-home config unchanged. With no config, the list from `/tmp` fails and says to run `hive init`. |
| V3 | Write a config at `$XDG_CONFIG_HOME/hive/config.toml` that names another marked temp home, then run `go test -race ./...` with that `XDG_CONFIG_HOME` | All packages pass. No test reads that config. |
| V4 | `git diff dev --stat` | Only the files of the source table changed. |
| V5 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

Never write `~/.config/hive/config.toml` of the user in a check. Use a temp
`XDG_CONFIG_HOME` for each check.

## Finish

Commit on the task branch with a subject such as `feat(home): add the home
config and hive init`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V5 results to the manager
with `hive inbox send`. Read only your own inbox.
