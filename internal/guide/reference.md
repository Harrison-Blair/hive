# hive reference

The command tree of hive. Each command line gives the path and the short
help. Each flag line below it gives the name, the type, and the usage.

hive: hive
  --demo string: open the TUI in demo mode from a fixture file
  -h, --help bool: help for hive
  --json bool: write JSON output
  --no-startup bool: open the TUI and start no coordinator, no supervisor, and no tab
  --size string: draw the demo TUI at WxH cells

hive agent: manage agent sessions
  -h, --help bool: help for agent
  --json bool: write JSON output

hive agent list: list agent records
  --all bool: include stale, stopped, and exited records
  -h, --help bool: help for list
  --json bool: write JSON output
  --profile string: agent profile
  --project string: project ID
  --role string: agent role

hive agent register: register the current agent session
  --harness string: harness kind
  -h, --help bool: help for register
  --json bool: write JSON output
  --model string: model name
  --name string: agent name
  --pane string: herdr pane ID
  --pid int64: holder process ID
  --project string: project ID
  --quiet-after int: quiet minutes
  --role string: agent role
  --task string: task ID

hive agent report: send the result of a general agent to its parent
  --agent string: reporting agent
  --file string: result file
  -h, --help bool: help for report
  --json bool: write JSON output
  --pane string: herdr pane ID of the caller

hive agent resolve: settle an unresolved agent record
  --agent string: agent name
  --as string: new status: exited or live
  -h, --help bool: help for resolve
  --json bool: write JSON output

hive agent send: send a prompt to an agent
  --agent string: agent name
  -h, --help bool: help for send
  --json bool: write JSON output
  --timeout int: wait timeout in seconds
  --wait bool: wait for a settled status

hive agent show: show one agent record
  -h, --help bool: help for show
  --json bool: write JSON output

hive agent spawn: spawn an agent in a pane
  --adopt string: adopt an existing workspace
  --agent string: caller agent name or key
  --allow-trust-dialog bool: start an agent that can show a trust dialog; keep its first prompt for the supervisor
  --arg stringArray: native agent argument
  --cwd string: working directory of an agent in a new tab
  --harness string: harness kind
  -h, --help bool: help for spawn
  --json bool: write JSON output
  --model string: model name
  --name string: agent name
  --owner bool: make the agent the owner of its project
  --pane string: destination herdr pane ID: hive starts the agent in this pane, and agent stop closes it; herdr then closes its tab or workspace when it holds no other pane
  --profile string: profile name: the default is general
  --project string: project ID
  --prompt string: spawner text of the agent
  --prompt-file string: file with the spawner text of the agent
  --quiet-after int: quiet minutes
  --role string: old name of --profile (wave A)
  --task string: task ID

hive agent stop: stop an agent that hive started
  --agent string: agent name
  -h, --help bool: help for stop
  --json bool: write JSON output

hive ask: ask the user questions in the TUI
  --agent string: calling agent name
  --file string: ask JSON file, or - for stdin
  -h, --help bool: help for ask
  --json bool: write JSON output
  --pane string: herdr pane ID

hive chat: send, reply to, and read chat messages
  --agent string: calling agent name
  -h, --help bool: help for chat
  --json bool: write JSON output
  --pane string: herdr pane ID

hive chat log: show chat history
  --agent string: calling agent name
  -h, --help bool: help for log
  --json bool: write JSON output
  --limit int: show the most recent N messages
  --pane string: herdr pane ID
  --since string: include messages on or after this UTC date

hive chat reply: save a coordinator reply
  --agent string: calling agent name
  -h, --help bool: help for reply
  --json bool: write JSON output
  --pane string: herdr pane ID
  --to string: user message ID

hive completion: Generate the autocompletion script for the specified shell
  -h, --help bool: help for completion
  --json bool: write JSON output

hive completion bash: Generate the autocompletion script for bash
  -h, --help bool: help for bash
  --json bool: write JSON output
  --no-descriptions bool: disable completion descriptions

hive completion fish: Generate the autocompletion script for fish
  -h, --help bool: help for fish
  --json bool: write JSON output
  --no-descriptions bool: disable completion descriptions

hive completion powershell: Generate the autocompletion script for powershell
  -h, --help bool: help for powershell
  --json bool: write JSON output
  --no-descriptions bool: disable completion descriptions

hive completion zsh: Generate the autocompletion script for zsh
  -h, --help bool: help for zsh
  --json bool: write JSON output
  --no-descriptions bool: disable completion descriptions

hive doctor: Check the environment
  -h, --help bool: help for doctor
  --json bool: write JSON output

hive goal: manage project goals
  --agent string: acting agent name
  -h, --help bool: help for goal
  --json bool: write JSON output
  --pane string: herdr pane ID

hive goal create: create a goal
  --agent string: acting agent name
  --grant string: grant line text
  -h, --help bool: help for create
  --json bool: write JSON output
  --pane string: herdr pane ID
  --project string: project ID
  --slug string: goal slug
  --title string: goal title

hive goal edit: edit a goal or wave
  --add-task string: task ID to add
  --add-wave string: new wave title
  --agent string: acting agent name
  --clear-needs-user bool: clear wave stop text
  --done-when string: wave completion point
  --goal string: goal ID
  -h, --help bool: help for edit
  --json bool: write JSON output
  --needs-user string: wave stop text
  --pane string: herdr pane ID
  --status string: goal status, or wave status with --wave
  --wave string: wave ID

hive goal list: list goals
  --agent string: acting agent name
  -h, --help bool: help for list
  --json bool: write JSON output
  --pane string: herdr pane ID
  --project string: project ID

hive goal show: show a goal
  --agent string: acting agent name
  -h, --help bool: help for show
  --json bool: write JSON output
  --pane string: herdr pane ID

hive guide: print the hive guide
  -h, --help bool: help for guide
  --json bool: write JSON output
  --write bool: write the reference of the live command tree

hive handoff: manage a role handoff
  -h, --help bool: help for handoff
  --json bool: write JSON output

hive handoff write: write the handoff
  -h, --help bool: help for write
  --json bool: write JSON output
  --scope string: handoff scope

hive inbox: read and send inbox records
  --agent string: reading agent
  --all bool: print each record and set nothing
  -h, --help bool: help for inbox
  --inbox string: inbox key or project:ID
  --json bool: write JSON output
  --pane string: herdr pane ID
  --role string: old name of --inbox

hive inbox send: append a message record
  --agent string: sending agent
  -h, --help bool: help for send
  --json bool: write JSON output
  --pane string: herdr pane ID
  --to string: inbox key

hive init: write the global hive home config
  -h, --help bool: help for init
  --json bool: write JSON output

hive land: deliver a verified task branch
  --agent string: caller agent
  --draft bool: open a draft pull request
  --evidence string: evidence URL or path
  --grant string: grant citation
  -h, --help bool: help for land
  --json bool: write JSON output
  --mode string: merge or pr
  --pane string: herdr pane ID
  --task string: task ID

hive logs: read the daily log files
  --agent string: agent ID
  --component string: component
  --follow bool: print new events as they arrive
  -h, --help bool: help for logs
  --json bool: write JSON output
  --level string: minimum level
  --limit int: maximum matches to print
  --request string: request ID
  --since string: duration or RFC 3339 cutoff
  --task string: task ID

hive memory: manage agent memory
  -h, --help bool: help for memory
  --json bool: write JSON output

hive memory write: write one memory
  -h, --help bool: help for write
  --json bool: write JSON output
  --scope string: memory scope

hive project: manage registered projects
  -h, --help bool: help for project
  --json bool: write JSON output

hive project list: list registered projects
  -h, --help bool: help for list
  --json bool: write JSON output

hive project register: register a Git repository
  --base string: base branch
  -h, --help bool: help for register
  --id string: project ID
  --json bool: write JSON output

hive project show: show one registered project
  -h, --help bool: help for show
  --json bool: write JSON output

hive supervise: run the supervisor of this home
  -h, --help bool: help for supervise
  --json bool: write JSON output

hive task: manage task records
  --agent string: acting agent name
  -h, --help bool: help for task
  --json bool: write JSON output
  --pane string: herdr pane ID

hive task assign: assign a task to an agent
  --agent string: acting agent name
  -h, --help bool: help for assign
  --json bool: write JSON output
  --pane string: herdr pane ID
  --task string: task ID

hive task cancel: cancel a task
  --agent string: acting agent name
  -h, --help bool: help for cancel
  --json bool: write JSON output
  --pane string: herdr pane ID
  --reason string: reason of the cancellation
  --task string: task ID

hive task complete: mark a task complete
  --agent string: acting agent name
  -h, --help bool: help for complete
  --json bool: write JSON output
  --pane string: herdr pane ID
  --task string: task ID

hive task create: create a task record and its brief
  --agent string: acting agent name
  --brief string: brief file, or - for standard input
  --depends-on string: comma-separated dependency IDs
  --feature string: feature slug
  -h, --help bool: help for create
  --json bool: write JSON output
  --pane string: herdr pane ID
  --project string: project ID
  --slug string: task slug
  --spec bool: mark the feature spec task
  --title string: task title

hive task edit: edit task metadata
  --agent string: acting agent name
  --brief string: brief file, or - for standard input
  --clear-needs-user bool: clear the user decision stop
  --feature string: feature slug
  -h, --help bool: help for edit
  --json bool: write JSON output
  --needs-user string: user decision needed for this task
  --pane string: herdr pane ID
  --task string: task ID

hive task list: list task records
  --agent string: acting agent name
  --feature string: feature slug
  -h, --help bool: help for list
  --json bool: write JSON output
  --pane string: herdr pane ID
  --project string: project ID
  --ready bool: keep only created tasks with verified dependencies
  --status string: task status

hive task reopen: reopen a verified task for repair
  --agent string: acting agent name
  -h, --help bool: help for reopen
  --json bool: write JSON output
  --pane string: herdr pane ID
  --reason string: reason to reopen the task
  --task string: task ID

hive task show: show one task record
  --agent string: acting agent name
  --brief bool: print the brief text only
  -h, --help bool: help for show
  --json bool: write JSON output
  --pane string: herdr pane ID

hive task verify: record one independent check of a task
  --agent string: acting agent name
  --fail bool: record a failed check
  -h, --help bool: help for verify
  --json bool: write JSON output
  --pane string: herdr pane ID
  --reason string: reason of a failed check
  --task string: task ID

hive version: Print the hive version
  -h, --help bool: help for version
  --json bool: write JSON output

hive worktree: manage task worktrees
  -h, --help bool: help for worktree
  --json bool: write JSON output

hive worktree create: create the worktree of a task
  --focus bool: focus the new workspace
  -h, --help bool: help for create
  --json bool: write JSON output
  --task string: task ID

hive worktree remove: remove the worktree of a task
  --force bool: skip the dirty and merged guards
  -h, --help bool: help for remove
  --json bool: write JSON output
  --task string: task ID

