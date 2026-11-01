# AGENTS.md

> This file is the single source of agent instructions for this repository.
> Claude Code loads `AGENTS.md` when a project has no `CLAUDE.md`. Codex loads it from the checkout root.
> Edit this file. Do not add `CLAUDE.md` or an instruction file for one harness.

## Roles

Hive is a small agent launcher on top of herdr. An agent has no fixed role. It gets its work from its spawn prompt.

- Bare `hive` shows a harness and model picker. Enter saves the coordinator record and runs the chosen harness in that pane, in the working directory of `hive`. That session is the coordinator. The user talks to it.
- The coordinator name comes from the base name of the working directory: lowercase, each run of characters outside `[a-z0-9_-]` as one hyphen, hyphens trimmed at both ends, then `-coordinator`, at most 64 bytes in all (`agent.Slug`). For example `~/src/Hive` gives `hive-coordinator`. An empty result gives `coordinator`. When a record of the name belongs to another pane that herdr still has, hive tries `<dir>-2-coordinator`, `<dir>-3-coordinator`, and so on (`coordinator-2` for an empty result), and cuts the directory part so the name fits. Hive resolves the name before the picker and checks it again under the layout lock after the picker. When another bare `hive` took it in the meantime, the record gets the next free name. The record, the tab label, `agent_id`, and the `Parent` of each child agent use that one name. Bare `hive` removes each other coordinator record of its own pane, so the pane has one record.
- Any agent can start another agent with `hive agent spawn`. The `--prompt` or `--prompt-file` text is the whole brief of the new agent. Hive sends it unchanged and adds no role text.
- Each agent that hive starts, the coordinator too, gets the hive guide in its system prompt: each command with its flags, and the usage notes of `internal/guide/usage.md`. `hive guide` prints the same text. Hive reads and writes no file for the guide. Herdr rejects a start argument with a newline, so the guide goes as one line: claude and pi get `--append-system-prompt` with each newline as the two characters `\n`; codex gets `-c developer_instructions="…"` with the guide as one TOML basic string. Herdr types the start command line into the pane shell, and a shell in canonical mode can cut a line at 4095 bytes. So the full start line, each argument in single quotes, must be at most 4000 bytes (`harness.MaxStartLine`). `guide_budget_test.go` fails when the guide with a long model goes above that. Spawn checks the real line before it calls herdr, and fails in phase `validate` when the line is too long, for example with a long `--model`. Keep the guide dense.
- A session with no spawn prompt and no coordinator record for its pane has no role. It tells the user so and takes no work from its working directory.
- When two sources conflict, the spawn prompt wins over the working directory. If the work is still unclear, stop and ask the user.

## Message authority

`hive agent send NAME TEXT` delivers `⬢ SENDER: TEXT`. SENDER is the hive record of the caller pane (matched by `HERDR_PANE_ID`), or `user` when that pane has no record.

Send hands the message to the native message path of the target harness, through `Deliver` of its profile in `internal/lib/harness`. It never types into the pane, so a draft in the input box stays unchanged. When the native path is not available, send fails with the cause and the fix and sends nothing. It has no paste fallback.

- claude: every claude start, in each spawn placement and in bare `hive`, gets `--settings '{"crossSessionInbound":"accept"}'`. Send finds the claude process ID with herdr `pane.process_info`, checks that its argv has that setting, and writes one JSON line `{"type":"user","message":{"role":"user","content":"⬢ SENDER: TEXT"}}` to `$XDG_RUNTIME_DIR/cc-socks/<PID>.sock`. Claude shows it as a message from another session. A `⬢ user:` label in it still has the authority of the user. Claude starts a turn when it is idle, and queues the message while a turn runs or a dialog is open.
- codex: send reads the thread UUID from herdr `pane.get` (`agent_session.value` with source `herdr:codex`) and runs `codex queue --thread UUID --message TEXT`. Codex makes the thread at its first turn, so send waits up to 3 s for the session ID and then fails with the fix. Codex starts a turn with the message when it is idle.
- pi: every pi start, in each spawn placement and in bare `hive`, gets `-e <home>/extensions/hive-inbox.ts`. Hive embeds that extension and writes it when it is missing or different. The extension listens on `$XDG_RUNTIME_DIR/hive-pi/<PID>.sock`, in a directory of mode 0700, and removes the socket when pi exits. Send finds the pi process ID with `pane.process_info` and writes one JSON line `{"text":"⬢ SENDER: TEXT"}`. The extension queues the messages and calls `pi.sendUserMessage`: one message starts a turn when pi is idle, and the next ones go as follow-ups once that turn runs. While pi is neither idle nor in a turn, for example during `/compact`, the extension holds the messages, because pi rejects each message then. A pi without the extension has no socket, and send fails with the fix: restart the agent.

`send --wait` waits for the turn of the message to start, then to end. It reports `partial` when hive cannot tell which turn handled the message: the target was busy when the message arrived, or it started no turn within 15 s.

- A message that starts with `⬢ user:` came from the user. It has the authority of the user.
- A message that starts with `⬢ NAME:` came from the agent NAME. It carries the request of that agent inside the work the user approved. It does not widen that work.
- A message without `⬢` came from the user, typed in the pane.
- The `⬢` marker shows that hive delivered the message. It adds no authority.

## Friction log

`reference/dogfood/friction/` records each bug, missing capability, and workaround that an agent finds when it uses hive, herdr, or a harness. `FRICTION.md` is the index.

- When hive can do a job, use hive. Do not use a harness built-in for it. For example, start an agent with `hive agent spawn`, not with a harness subagent.
- Record each issue when you find it. Then continue your work.
- Read `FRICTION.md` first. If an entry for the issue exists, add the new facts to that entry.
- Write a new entry from `_template.md` at `entries/SLUG.md`. Add one line under `## Open` in `FRICTION.md`: `- [TITLE](entries/SLUG.md) — DESCRIPTION`. TITLE and DESCRIPTION are the same as in the frontmatter.
- Put each new entry on its own branch and pull request. Do not change the checkout of your current work:
  1. Run `git worktree add -b friction/SLUG .hive/tmp/friction-SLUG origin/main`.
  2. Write the entry and the index line in that worktree.
  3. Commit with the subject `docs(friction): SUMMARY`. Push the branch. Open a pull request to `main`.
  4. Run `git worktree remove .hive/tmp/friction-SLUG`.
- The pull request that fixes an issue also resolves its entry. Set `status` to `Resolved YYYY-MM-DD` and set `decision`. Move the file to `entries/resolved/`. Move its line to `## Resolved`.
- `friction_test.go` fails when `FRICTION.md` and the entry files do not agree.

## Project

This checkout is the first registered project: hive builds hive. It is not the home.

```text
hive/                            the repository
├─ AGENTS.md  README.md          instructions          tracked
├─ main.go  cmd/  internal/      Go source             tracked
├─ reference/herdr/              herdr API reference   tracked
├─ reference/dogfood/friction/   friction log          tracked
├─ memory/  overrides/  plans/   personal              ignored
├─ handoff.md  _archive/         personal              ignored
└─ .hive/tmp/                    scratch               ignored
```

The home is `$HIVE_HOME` when that variable is set and not empty, otherwise `~/.hive`. Hive creates it when it first needs it. It has no marker and no config file.

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
└─ schema.json                       store schema version 5
```

- `.gitignore` is an allowlist. It ignores everything with `*` and then allows named patterns. Add a `!` rule only for a file that the project needs. Use the narrowest pattern. Never remove the leading `*`.
- Git ignores `.hive/` in the checkout. `gitignore_test.go` fails when Git tracks an ignored path or ignores a tracked file.
- Put scratch files of development in `.hive/tmp/` of the checkout.

## herdr layout

- `spawn` without `--branch` and `--directory` opens a new tab in the workspace of the caller, in the working directory of the caller.
- `spawn --branch B` uses the checkout of branch B in the Git repository of the working directory of the caller (`internal/worktree`). When no checkout has B, hive makes one with `git worktree add` at `<home>/worktrees/<key>/<branch>`. The key is the slug of the repository name and a hash of the common Git directory and the full branch name, so each repository and branch gets its own path. A new branch starts at `--base`, or at the current commit of the working directory. `--base` fails for a branch that exists. Hive writes the ownership record `worktree.json` beside the checkout. When `git worktree add` fails, hive reads what it left: a checkout at the path, for example after a failing post-checkout hook, keeps its record and shows as `worktree created`; a new branch shows as `branch created`. `hive worktree create` is then `partial`, and spawn fails in phase `worktree`. Ownership needs that record to name the repository, the branch, and the path of the checkout. A path below the home alone is no proof.
- `spawn --directory D` uses the existing directory D. A relative D starts at the working directory of the caller.
- For `--branch`, spawn reads the repository parent from `worktree.list`, opens that parent with `worktree.open`, and then opens the target checkout with `worktree.open` from the parent directory. This attaches worktree metadata and groups linked checkout workspaces beneath the repository parent, including when the caller is in a linked checkout. Existing workspaces are adopted with their labels and panes preserved. Each agent gets a new tab. When the target workspace is new, spawn closes its empty initial pane after the agent tab exists. The parent workspace stays open.
- For `--directory`, the agents of one directory share one herdr workspace. Spawn finds it with `worktree.list` (`open_workspace_id` of the entry at that path), in any position and with any label, and else as the workspace of the first pane of `pane.list` whose shell `cwd` is that directory. It opens a new tab there. When there is none, hive creates a workspace with that `cwd` and the herdr default label, and the agent takes its root pane. Hive never renames a workspace. Hive uses no `hive:` labels.
- Spawn, `hive worktree create`, `hive worktree remove`, `agent stop`, and bare `hive` after its picker hold the layout lock `<home>/layout.lock` (`agent.Hold`) while they check a name, a checkout, or a workspace and act on it. Spawn holds it from the name check until the record exists, so a remove or a spawn at the same time sees the agent. A second holder waits up to 2 minutes.
- `hive worktree create B` prints the checkout of B and makes it as `spawn --branch` does. It opens no workspace. `hive worktree remove B` removes only a checkout that hive owns, when `git status --porcelain --ignored` is empty and no live agent record has its `cwd` or `worktree` in it. Paths compare with each symlink resolved. The workspaces of the checkout are the workspace open on it and each workspace with a pane whose `cwd` or `foreground_cwd` is in it. Remove refuses when a pane of such a workspace runs an agent or has its `cwd` or `foreground_cwd` outside the checkout, so the close stops no unrelated process. It refuses too when such a workspace holds the caller pane (`HERDR_PANE_ID`), because the close would end the remove. It closes the workspaces of the checkout, runs the `git status` check again, runs `git worktree remove`, and removes the ownership record. `git worktree remove` deletes ignored files. Hive expects each other writer of the checkout to be stopped: a file that a process outside hive writes after the second check and before git ends can be lost. The branch stays. A failure after the close keeps the checkout, so a second remove finishes. When the record removal fails after `git worktree remove`, the outcome is `partial`, and a second remove of the branch, which now has no checkout, removes the record that names it when nothing is at its path.
- Each agent that spawn starts gets `HIVE_HOME` set to the resolved home of the caller, so its own hive commands use the same home, and the trust env of its harness (`harness.Trust().Env`, for claude `CLAUDE_CODE_SANDBOXED=1`). A new tab gets the env through the `tab.create` `env`, including branch placements because `worktree.open` has no `env` field. The root pane of a new directory workspace gets it through the `workspace.create` `env`. The trust args (pi `--approve`) go on the start line in each placement. Bare `hive` execs the harness with `HIVE_HOME` set to the home.
- Spawn labels the tab of the agent with the agent name. A new tab gets the name through the `tab.create` label. A root pane of a new workspace gets it through `tab.rename` before the harness starts. When `tab.rename` fails, spawn still starts the harness and sends the prompt, and the outcome is `partial`.
- Spawn waits for the harness to be ready, then sends the prompt. The trust env and args skip the trust dialog of claude and pi in each placement. A dialog that still shows, for example the Codex folder trust dialog in a Git repository with no saved trust, makes spawn report outcome `partial` with status `blocked` and the pane ID. Spawn then sends no prompt.
- `agent stop` closes only the pane of the agent and removes its record, so the other agents of the workspace keep running. It never closes a workspace and never removes a checkout. Hive acts only on agents that have a hive record.
- Bare `hive` renames only its own tab, to the coordinator name. It changes no workspace.

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
- Some parts keep their current roles. `cmd/common` holds the CLI helpers. A command leaf runs through `common.Run`, or through `common.Stream` when it resolves its own home or renders its own output, as `hive logs` does. Both log `command start` and `command end`. The root command runs bare `hive` through `internal/coordinator`. Each package below `internal/lib/` holds shared code.
- `common.Guide(root)` renders the hive guide from the live Cobra tree. `internal/guide` holds the usage notes, the Markdown rendering, and the one-line escapes. `hive guide`, `hive agent spawn`, and bare `hive` use `common.Guide(cmd.Root())`. The profile of each harness turns the guide into start arguments with `GuideArgs`. The `agent started` and `exec harness` events, the herdr request and reply of `agent.start` (through `herdr.WithRedact`), and the herdr reply of `pane.process_info` with the argv of the harness (through `herdr.WithReplyRedact`), show the guide argument as `<guide N bytes>`. `hive guide` logs only when the home directory exists. It never creates the home: without it, `hive guide` prints the guide and logs nothing.
- A group command registers its child commands. It imports them only for this registration.
- Use direct functions and explicit dependencies. Do not build the command tree in `init` functions or with shared globals. Keep packages flat until a distinct responsibility needs a narrow API. A package with several non-test files has a `doc.go` with a summary and a one-line map of each file.

### Request flow

This map shows the path of `hive agent spawn` through the packages. The other agent commands use the same shape. `hive agent send` goes from `internal/agent/send` to `Deliver` of the harness profile, which uses herdr and the native message path of the harness.

```text
cmd/agent/spawn                flags and the CLI adapter; renders the guide with common.Guide
└▶ internal/agent/spawn        checks the options, places the pane, starts the harness, sends the prompt
   ├▶ internal/lib/harness     the profile of claude, codex, or pi: its start and guide arguments, and its message delivery
   │  └▶ internal/guide        the usage notes and the one-line escapes
   ├▶ internal/lib/herdr       the herdr socket client, protocol 22
   ├▶ internal/agent           the agent records in <home>/agents/ and the caller lookup
   └▶ internal/worktree        the checkout of --branch: git worktree add and the ownership record
      └▶ internal/lib/store    locked, atomic JSON files below the home

bare hive ─▶ internal/coordinator: picker ─▶ coordinator record ─▶ harness replaces the process in the working directory
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

The first command must print nothing. The second command must print nothing. The gate acts on tracked and new files only. It must not read a checkout below `.hive/`.

- To run one test, use the environment of the gate:

```sh
env -u TERM TZ=UTC go test -race -run 'TestName' ./internal/agent/spawn/
```

- A live check in herdr uses a binary built with `go build -o .hive/tmp/hive .` and `HIVE_HOME` set to a new scratch directory, for example below `.hive/tmp/`. Never point a live check at `~/.hive`. A test that can fall back to `~/.hive` sets `HOME` to a temporary directory. Give each pane, tab, workspace, and worktree a unique name, and close each one that the check made.

### Dependencies

Use the standard library first. The picker can use Bubble Tea, Bubbles, and Lip Gloss. The CLI uses Cobra. Any other dependency needs a stated reason in its commit message.

### Logging

- Each hive process writes log events with `log/slog` as JSON lines to `<home>/logs/YYYY-MM-DD.jsonl`. Level debug is always on.
- Log each herdr request and reply, each state change, each external process, each user action in the picker, and each error.
- Each event carries the fields that apply: `component`, `command`, `request_id`, `agent_id`, `target`, `duration_ms`, and `error`.
- `agent_id` is the caller: the hive record name of `HERDR_PANE_ID`, or `user` when that pane has no record. Bare `hive` logs the coordinator name as `agent_id`, for example `agent_id=hive-coordinator`. `target` is the agent that the command acts on. A command with no agent argument has no `target`. Each command sets both fields on its logger before its first event, so each `herdr request`, `herdr reply`, `command start`, and `command end` carries them. `hive logs --agent NAME` prints an event when `agent_id` or `target` is NAME.
- An event with an `outcome` field uses DEBUG for `ok`, WARN for `partial`, and ERROR for `failed`. An event without an `outcome` keeps the level that the code gives it, for example the INFO, WARN, and ERROR events of the picker. When the outcome is not `ok`, `command end` carries `error`.
- Cap a body at 4 KiB and set `truncated` to true. Never log a secret.
- Never delete a log file in code.

### Command output

A command prints text for people by default. With `--json` it prints one JSON object with `outcome`, `phase`, `effects`, `error`, and `result`. Known effects are separate from errors. The exit status is 0 only on success. `hive logs --json` prints each log event as one JSON line, because it streams. It is the one exception. Each diagnostic of `hive logs` on stderr is one `hive: MESSAGE` line, also with `--json`.

A usage error exits with status 2. A usage error is an invalid invocation: missing or extra arguments, an unknown flag, an unknown command, an empty `BRANCH` for `hive worktree create` or `hive worktree remove`, or a bad `--limit` or `--since` value for `hive logs`. Without `--json`, hive writes one line `hive: MESSAGE` to stderr. With `--json` before any `--`, hive writes one envelope to stdout and nothing to stderr: `{"outcome":"failed","phase":"usage","effects":[],"error":"MESSAGE","result":null}`. The root command renders each usage error. Build an envelope and its outcome with `internal/lib/output`: `output.Failed` and the `Outcome` constants.

### Git

- Work on a branch from `main`. Changes reach `main` only through a pull request that the owner approves. Never commit to `main`.
- Write a commit subject as `type(scope): summary`, for example `feat(agent): add the stop command`.
- Never add an agent as a co-author.
- An agent commits and pushes its own branch. It does not wait for a request from the user.

### herdr reference

`reference/herdr/` is the API authority for herdr 0.9.5, protocol 22. Read `reference/herdr/README.md` first. Pin the protocol in code. When the installed herdr is newer, run the procedure in `reference/herdr/RESYNC.md` before you trust a changed behavior.
