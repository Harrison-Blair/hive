# Remove the home flag

| Item | Value |
| --- | --- |
| Wave | H1, task B of 2 |
| Tier | decent |
| Depends on | 43-home-config. Base: `dev` after 43-home-config lands |
| Grant | `grants.md 2026-10-19 H1 home config wave` |
| Evidence | `evidence/<task-id>/` at the root of the worktree, not `spec/hive-v1/evidence/`. `VERIFIED.md` goes in the same folder. |

## Overview

This task removes the home option. Hive finds the home from
`HIVE_HOME`, the config that `hive init` writes, or the marker above the
working directory. Task 43-home-config added that resolution. The plan is
`projects/hive/plans/home-spec-planning/MAP.md` in the home, decision 3. The
user agreed to it. Do not change a decision in this brief. When the code does
not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `cmd/root.go` declares the persistent home option. `cmd/common/common.go`
  and `cmd/logs/logs.go` read it and pass it to the home resolution.
- `internal/agent/brief.go` adds the home path to generated worker and verifier
  commands.
- Many tests in `cmd/` pass the home option to `ExecuteWithArgs`. Find them
  with `flag=$(printf '%s%s' '--' 'home'); git grep -n -- "$flag" -- '*.go'`.
- `skills_test.go` has fixtures with the home option.
- After 43-home-config, `HIVE_HOME` selects a home and does not fail on a
  different marked home above the working directory.
- `t.Setenv` panics in a test that calls `t.Parallel`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `cmd/root.go`, `cmd/common/common.go`, `cmd/logs/logs.go`, `internal/lib/home/`, each other caller of `home.Resolve` (for example `internal/tui/tui.go`), and their tests | R1. Only the call to the changed resolution in a caller. |
| `internal/agent/brief.go` and its tests | R2. |
| Each Go test that uses the home option | R3. List each file in the evidence. |
| `.agents/skills/*/SKILL.md`, `internal/agent/templates/project-agents.md`, `AGENTS.md`, `skills_test.go` | R4. |
| `spec/hive-v1/05-tracker.md`, `spec/hive-v1/03-architecture.md`, `spec/hive-v1/04-risks.md`, and the briefs in `spec/hive-v1/tasks/` | R4. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `spec/hive-v1/tasks/<task-id>.md` | R5. New. A copy of this brief. |

Do not edit files below `evidence/`, `spec/hive-v1/evidence/`, or
`reference/dogfood/friction/`. Do not edit other files.

## Requirements

### R1 Flag removal

- Remove the home option from the root command. A command with the removed
  option fails as an unknown flag with exit status 2, as other unknown flags do.
- Remove the flag parameters of the home resolution and each code path that
  becomes unused.
- Keep the resolution order of 43-home-config without step 1: `HIVE_HOME`,
  the config, the marker, else the message that says to run `hive init`.

### R2 Prompts

The generated worker and verifier prompts do not contain the home option or the
home path in a hive command. Each test asserts the exact prompt text as a literal
string. Do not compute the expected text with the code under test.

### R3 Tests

- Change each test that uses the home option to set `HIVE_HOME` with `t.Setenv`.
  Remove `t.Parallel` from a test only when `t.Setenv` requires it.
- Each test that resolves a home isolates the config of the user, as in
  43-home-config.
- Add a test that the removed home option is unknown. Run it against the base
  code first. It must fail there.
- Do not weaken, skip, or delete an assertion.

### R4 Docs

- Remove the home option from each skill, the manager template, `AGENTS.md`,
  the tracker checkpoint commands, the architecture and risk docs, and the
  briefs.
- Where a command used `hive` with the home option and a value, write
  `HIVE_HOME=X hive CMD` when the text needs a fixture home. Otherwise, write
  `hive CMD`.
- In a past brief, keep one sentence that only records what that task built.
  For example, say that T01 added the home option. Add no history text. List
  each kept sentence in the evidence.
- Keep each changed instruction text in ASD-STE100 Strict. `ste_test.go` must
  pass.

### R5 Brief copy

Copy this brief to `spec/hive-v1/tasks/<task-id>.md` with a baseline of 0.
When the copy causes a violation, change the words of that sentence and keep
its meaning. Record each change in the evidence.

## Not in scope

The config format, `hive init`, the spec layout (H2), `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run the R2 and R3 new tests against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Build the task binary. With a temp `XDG_CONFIG_HOME` and `HIVE_HOME` unset, run `home_option=$(printf '%s%s' '--' 'home'); ./hive "$home_option" /tmp task list --project hive`. Run `./hive task list --project hive` from the worktree. Run `HIVE_HOME=<temp fixture home> ./hive task list --project x` from the worktree. | The first exits 2 with an unknown flag error. The second exits 0 with the tasks of the home. The third uses the fixture home and does not fail with `home conflict`. |
| V3 | `flag=$(printf '%s%s' '--' 'home'); git grep -n -- "$flag" -- ':!evidence' ':!spec/hive-v1/evidence' ':!reference/dogfood/friction'` | Each line that remains is a kept sentence of R4, listed in the evidence. |
| V4 | Write a config at `$XDG_CONFIG_HOME/hive/config.toml` that names another marked temp home, then run `go test -race ./...` with that `XDG_CONFIG_HOME` | All packages pass. |
| V5 | `git diff dev --stat` | Only the files of the source table changed. |
| V6 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

Never write `~/.config/hive/config.toml` of the user in a check. Use a temp
`XDG_CONFIG_HOME` for each check.

## Finish

Commit on the task branch with a subject such as `feat(cli): remove the home
flag`. Push the task branch. Write the evidence. Run `hive task complete`.
Send the commit SHA and the V1-V6 results to the manager with
`hive inbox send`. Read only your own inbox.
