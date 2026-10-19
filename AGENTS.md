# AGENTS.md

> This file is the single source of agent instructions for this repository.
> Claude Code loads `AGENTS.md` when a project has no `CLAUDE.md`. Codex loads it from the checkout root.
> Edit this file. Do not add `CLAUDE.md` or an instruction file for one harness.

## Role gate

Read this section first. Your working directory gives your role. A brief can also name your role. The brief wins.

| Working directory | Role | Then read |
| --- | --- | --- |
| The root of this checkout | coordinator, unless the brief names manager | [Coordinator](#coordinator) or the brief |
| `projects/<project-id>/` | manager | the project instructions, then your brief |
| `projects/<project-id>/worktrees/<task-id>/` | worker | your brief, then [Development](#development) |

- A manager or a worker never runs coordinator startup. This rule also applies after a context reset.
- When the role evidence is missing or in conflict, hold and report the conflict to the user. Do not take the coordinator role.
- Claude Code loads this file in every checkout below the root. The role gate applies there too.

### Bootstrap clause

Hive cannot yet register a project or spawn a manager. Until it can, the user can say "develop" to a root session. That session then works as a developer. It changes hive source directly and obeys [Development](#development). Remove this clause when the hive manager spawns its first worker.

During bootstrap, the user spawns workers by hand in worktrees under `projects/hive/worktrees/`. The task file in `spec/hive-v1/tasks/` is the brief of such a worker. [spec/hive-v1/05-tracker.md](spec/hive-v1/05-tracker.md) gives the procedure.

## Project

Hive coordinates coding agents in herdr from one home. The user talks to one coordinator. A manager owns each project. Workers do one task each in their own checkout. Hive builds hive: this checkout is the home and the first registered project.

[spec/hive-v1/README.md](spec/hive-v1/README.md) is the design authority. It gives the goals, the features, the architecture, the risks, and the task plan. The branches `plan/hive-v1` and `integration/hive-v1` are history. Do not import them.

## Home layout

```text
hive/                            the checkout is the home
├─ AGENTS.md  STE-GLOSSARY.md    instructions          tracked
├─ main.go  cmd/  internal/      Go source             tracked
├─ spec/  reference/             design and API docs   tracked
├─ .hive/README.md               home marker           tracked
├─ projects/<project-id>/        project directory     ignored
│  └─ worktrees/<task-id>/       worktree              ignored
├─ memory/  overrides/  plans/   personal              ignored
├─ handoff.md                    coordinator handoff   ignored
└─ .hive/  state/  logs/  tmp/   runtime               ignored
```

- Git ignores `projects/`, `memory/`, `overrides/`, `plans/`, `handoff.md`, and everything in `.hive/` except `.hive/README.md`. That file marks the home. `gitignore_test.go` fails when Git tracks one of the ignored paths, ignores the marker, or ignores a tracked file.
- `.gitignore` is an allowlist. It ignores everything with `*` and then allows named patterns. Add a `!` rule only for a file that the project needs. Use the narrowest pattern. Never remove the leading `*`.
- Put scratch files in `.hive/tmp/`.

## Coordinator

The coordinator runs in a pane of the hive session. At startup it registers itself, reads `handoff.md`, checks the saved and live state, and reports open decisions and project status. It spawns a manager when the user asks. It starts `hive supervise` when the supervisor claim is stale. It records a preference in `overrides/` when the user states one. It never changes project source files. Read `.agents/skills/hive-coordinator/SKILL.md` for the full coordinator procedure.

## herdr layout

- The hive session is the one herdr session that hosts the home. The coordinator and the supervisor run there.
- Each managed project has one workspace with the label `hive:<project>`. Hive makes it when it spawns the manager, with the project directory as the working directory. The manager runs in the root pane of the first tab.
- Hive makes each worktree with `worktree.create`. The source is the registered repository path in `cwd`, never the `hive:<project>` workspace. The request gives an explicit `path` below `projects/<project-id>/worktrees/`, the branch `task/<task-id>`, the base branch of the project, and the task ID as the label. The worker starts in the root pane of the first tab of the new workspace.
- A first tab is never an empty shell. Hive starts the agent in the pane that herdr returns.
- Hive closes what it made. `agent stop` closes the pane in which hive started the agent, and herdr closes the workspace with its last pane. `worktree remove` closes an open worktree workspace, then removes the worktree with Git. Hive never closes the pane of an agent that it did not start.

## Development

These rules apply to a developer and to a worker on the hive project.

### Task flow

A spec task is a brief in `spec/hive-v1/tasks/`. `spec/hive-v1/05-tracker.md` gives the procedure to start, verify, and land a task, and the evidence rules. Read it before you start a task. A worker edits only the files that its brief names, plus new tests beside them.

### Go layout

- `main.go` calls `cmd.Execute()` and sets the exit status. `go build .` makes a binary named `hive`.
- `cmd/` holds the root command and thin Cobra wiring. `NewRootCmd()` builds a fresh tree. `ExecuteWithArgs()` lets a test inject arguments and streams.
- Each subcommand has its own folder, `cmd/<name>/` or `cmd/<parent>/<subcommand>/`, with a `New() *cobra.Command` constructor. The parent registers its children. The constructor calls into `internal/`.
- `internal/<name>/` mirrors the command tree. Each leaf owns its options, orchestration, result types, text rendering, and tests.
- `internal/lib/<capability>/` holds focused code that several packages share. Do not make one flat grab-bag package. Do not extract a utility before two callers exist.
- `internal/tui/` holds the Bubble Tea program that bare `hive` runs.
- `internal/` never imports Cobra or `cmd/`. A child package never imports its parent.
- Use direct functions and explicit dependencies. Do not build the command tree in `init` functions or with shared globals. Keep packages flat until a distinct responsibility needs a narrow API. A package with several non-test files has a `doc.go` with a summary and a one-line map of each file.

### Tests and the gate

- Each Go change is test-first. Write a failing test. Run it. Make sure that it fails for the correct reason. Write the minimal code. Refactor.
- A test counts only when it fails when the behavior breaks and passes when the behavior is correct. Do not compute the expected value with the logic under test. Do not mock the behavior under test. Mock only an external dependency.
- Test the TUI at the model. Call `Update` and `View` with messages. Run the program on a string reader for an end-to-end check.
- Never weaken, skip, or delete a test to get a green run.
- Before you report a change as done, run the gate from the root and report its output:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run ./...
go test -race ./...
```

The first command must print nothing. The gate acts on tracked and new files only. It must not read a checkout below `projects/`.

### Dependencies

Use the standard library first. The TUI can use Bubble Tea, Bubbles, and Lip Gloss. The CLI uses Cobra. Any other dependency needs a stated reason in its commit message.

### Logging

- Each hive process writes log events with `log/slog` as JSON lines to `.hive/logs/YYYY-MM-DD.jsonl` in the home. Level debug is always on.
- Log each socket request and reply, each state transition, each external process, each user action in the TUI, and each error.
- Each event carries the fields that apply: `component`, `command`, `request_id`, `project_id`, `task_id`, `agent_id`, `duration_ms`, and `error`.
- Cap a body at 4 KiB and set `truncated` to true. Never log a secret.
- Never delete a log file in code.

### Command output

A command prints text for people by default. With `--json` it prints one JSON object with `outcome`, `phase`, `effects`, `error`, and `result`. Known effects are separate from errors. The exit status is 0 only on success. `hive logs --json` prints each log event as one JSON line, because it streams. It is the one exception.

### Instruction text

Write instruction text in ASD-STE100 Strict: this file, skills, role instructions, briefs, the task files in `spec/hive-v1/tasks/`, and the tracker. `STE-GLOSSARY.md` gives one meaning for each term. Use each term only with that meaning. `ste_test.go` checks the files in `steFiles` against `testdata/ste-baseline.json`. A count can fall. It cannot rise. When you add an instruction file, add it to `steFiles` with a baseline of 0.

### Git

- Work on `dev` or on a branch from `dev`. Changes reach `main` only through a pull request that the owner approves. Never commit to `main`.
- Write a commit subject as `type(scope): summary`, for example `feat(tui): add the log view`.
- Never add an agent as a co-author.
- An agent commits and pushes its own work. It does not wait for a request from the user. A developer commits on `dev` and pushes `dev`. A worker commits on its task branch and pushes that branch. `spec/hive-v1/05-tracker.md` gives the steps.

### herdr reference

`reference/herdr/` is the API authority for herdr 0.9.5, protocol 22. Read `reference/herdr/README.md` first. Pin the protocol in code. When the installed herdr is newer, run the procedure in `reference/herdr/RESYNC.md` before you trust a changed behavior.

### Dogfooding

When hive offers a capability, use hive for it instead of a harness built-in. Record each bug, missing capability, or workaround in `reference/dogfood/friction.md` with its Issue / Status / Summary / Reproduction steps format.
