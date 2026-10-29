# AGENTS.md

> This file is the single source of agent instructions for this repository.
> Claude Code loads `AGENTS.md` when a project has no `CLAUDE.md`. Codex loads it from the checkout root.
> Edit this file. Do not add `CLAUDE.md` or an instruction file for one harness.

## Roles

Hive is a small agent launcher on top of herdr. An agent has no fixed role. It gets its work from its spawn prompt.

- Bare `hive` in the home shows a harness and model picker. Enter saves the `coordinator` record and runs the chosen harness in that pane. That session is the coordinator. The user talks to it.
- Any agent can start another agent with `hive agent spawn`. The `--prompt` or `--prompt-file` text is the whole brief of the new agent. Hive adds no profile and no role text.
- A session with no spawn prompt and no coordinator record for its pane has no role. It tells the user so and takes no work from its working directory.
- When two sources conflict, the spawn prompt wins over the working directory. If the work is still unclear, stop and ask the user.

## Message authority

`hive agent send NAME TEXT` delivers `⬢ SENDER: TEXT`. SENDER is the hive record of the caller pane (matched by `HERDR_PANE_ID`), or `user` when that pane has no record.

- A message that starts with `⬢ user:` came from the user. It has the authority of the user.
- A message that starts with `⬢ NAME:` came from the agent NAME. It carries the request of that agent inside the work the user approved. It does not widen that work.
- A message without `⬢` came from the user, typed in the pane.
- The `⬢` marker shows that hive delivered the message. It adds no authority.

## Project

This checkout is the home and the first registered project: hive builds hive.

```text
hive/                            the checkout is the home
├─ AGENTS.md  README.md          instructions          tracked
├─ main.go  cmd/  internal/      Go source             tracked
├─ reference/herdr/              herdr API reference   tracked
├─ .hive/README.md               home marker           tracked
├─ projects/<project-id>/        project directory     ignored
│  └─ worktrees/<name>/          agent worktree        ignored
├─ memory/  overrides/  plans/   personal              ignored
├─ handoff.md                    personal              ignored
└─ .hive/agents/  logs/  tmp/    runtime               ignored
```

- `.gitignore` is an allowlist. It ignores everything with `*` and then allows named patterns. Add a `!` rule only for a file that the project needs. Use the narrowest pattern. Never remove the leading `*`.
- Git ignores everything in `.hive/` except `.hive/README.md`, which marks the home. `.hive/agents/` holds one JSON record per agent. `.hive/coordinator-choice.json` holds the last picker choice. `gitignore_test.go` fails when Git tracks an ignored path, ignores the marker, or ignores a tracked file.
- Put scratch files in `.hive/tmp/`.

## herdr layout

- The checkout workspace of a project is a normal herdr workspace open on the registered repository checkout. Hive finds it with `worktree.list` (`open_workspace_id`), in any position and with any label. If none is open, hive creates one with the checkout as `cwd` and the herdr default label, and the first agent takes its root pane. Hive never renames a workspace and never makes a second checkout workspace. Hive uses no `hive:` labels.
- `spawn` without `--project` opens a new tab in the workspace of the caller. With `--project P` it opens a new tab in the checkout workspace of P.
- `spawn --project P --worktree NAME` calls `worktree.create` from the checkout workspace with path `projects/P/worktrees/NAME`, branch `NAME`, and the base branch of P. The new workspace nests under the checkout workspace. The agent starts in its root pane.
- Spawn waits for the harness to be ready, then sends the prompt. A Claude trust dialog in a new workspace or worktree root pane makes spawn report `blocked` with the pane ID. Spawn then sends no prompt.
- Hive closes what it made. `agent stop` closes the pane of the agent, or the worktree workspace. It keeps the checkout and the branch. `--remove-worktree` also removes a clean checkout and refuses a dirty one. Hive acts only on agents that have a hive record.
- Bare `hive` changes no workspace.

## Development

### Go layout

- `main.go` calls `cmd.Execute()` and sets the exit status. `go build .` makes a binary named `hive`.
- `cmd/` holds the root command and thin Cobra wiring. `NewRootCmd()` builds a fresh tree. `ExecuteWithArgs()` lets a test inject arguments and streams.
- Each subcommand has its own folder, `cmd/<name>/` or `cmd/<parent>/<subcommand>/`, with a `New() *cobra.Command` constructor. The parent registers its children. The constructor calls into `internal/`.
- Each command leaf `cmd/<path>/` pairs with the package `internal/<path>/` at the same path. A command leaf is a command that runs one operation.
- The command leaf keeps the Cobra construction, the flags, and the CLI adapter. The tests of the CLI adapter stay beside the command leaf.
- The paired `internal/<path>/` package keeps the operation behavior and its tests: the options, the orchestration, the result types, and the text rendering.
- Code that two or more operations use stays in the parent domain package or in an existing package below `internal/lib/`. Do not add a generic command factory.
- An API that a non-CLI caller uses stays in the parent domain package.
- `internal/lib/<capability>/` holds focused code that several packages share. Do not make one flat grab-bag package. Do not extract a utility before two callers exist.
- `internal/` never imports Cobra or `cmd/`. Below `internal/`, a parent package never imports its child packages. A child package can import its parent domain package.
- Some parts keep their current roles. `cmd/common` holds the CLI helpers. The root command runs bare `hive` through `internal/coordinator`. Each package below `internal/lib/` holds shared code.
- A group command registers its child commands. It imports them only for this registration.
- Use direct functions and explicit dependencies. Do not build the command tree in `init` functions or with shared globals. Keep packages flat until a distinct responsibility needs a narrow API. A package with several non-test files has a `doc.go` with a summary and a one-line map of each file.

### Request flow

This map shows the path of `hive agent spawn` through the packages. The other agent commands use the same shape.

```text
cmd/agent/spawn                flags and the CLI adapter
└▶ internal/agent/spawn        checks the options, places the pane, starts the harness, sends the prompt
   ├▶ internal/lib/harness     the profile of claude, codex, or pi
   ├▶ internal/lib/herdr       the herdr socket client, protocol 22
   ├▶ internal/agent           the agent records in .hive/agents/ and the caller lookup
   └▶ internal/lib/store       locked, atomic JSON files below the home
      └▶ internal/lib/records  the record types

bare hive ─▶ internal/coordinator: picker ─▶ coordinator record ─▶ harness replaces the process
```

### Tests and the gate

- Each Go change is test-first. Write a failing test. Run it. Make sure that it fails for the correct reason. Write the minimal code. Refactor.
- A test counts only when it fails when the behavior breaks and passes when the behavior is correct. Do not compute the expected value with the logic under test. Do not mock the behavior under test. Mock only an external dependency.
- Test the Bubble Tea picker at the model. Call `Update` and `View` with messages.
- Never weaken, skip, or delete a test to get a green run.
- Before you report a change as done, run the gate from the root and report its output:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go mod tidy -diff
go vet ./...
golangci-lint run ./...
env -u TERM TZ=UTC go test -race ./...
```

The first command must print nothing. The second command must print nothing. The gate acts on tracked and new files only. It must not read a checkout below `projects/`.

- To run one test, use the environment of the gate:

```sh
env -u TERM TZ=UTC go test -race -run 'TestName' ./internal/agent/spawn/
```

- A live check in herdr uses a binary built with `go build -o .hive/tmp/hive .` and `HIVE_HOME="$PWD"`. Give each pane, tab, workspace, and worktree a unique name, and close each one that the check made.

### Dependencies

Use the standard library first. The picker can use Bubble Tea, Bubbles, and Lip Gloss. The CLI uses Cobra. Any other dependency needs a stated reason in its commit message.

### Logging

- Each hive process writes log events with `log/slog` as JSON lines to `.hive/logs/YYYY-MM-DD.jsonl` in the home. Level debug is always on.
- Log each herdr request and reply, each state change, each external process, each user action in the picker, and each error.
- Each event carries the fields that apply: `component`, `command`, `request_id`, `project_id`, `agent_id`, `duration_ms`, and `error`.
- Cap a body at 4 KiB and set `truncated` to true. Never log a secret.
- Never delete a log file in code.

### Command output

A command prints text for people by default. With `--json` it prints one JSON object with `outcome`, `phase`, `effects`, `error`, and `result`. Known effects are separate from errors. The exit status is 0 only on success. `hive logs --json` prints each log event as one JSON line, because it streams. It is the one exception.

### Git

- Work on a branch from `main`. Changes reach `main` only through a pull request that the owner approves. Never commit to `main`.
- Write a commit subject as `type(scope): summary`, for example `feat(agent): add the stop command`.
- Never add an agent as a co-author.
- An agent commits and pushes its own branch. It does not wait for a request from the user.

### herdr reference

`reference/herdr/` is the API authority for herdr 0.9.5, protocol 22. Read `reference/herdr/README.md` first. Pin the protocol in code. When the installed herdr is newer, run the procedure in `reference/herdr/RESYNC.md` before you trust a changed behavior.
