# STE glossary

One hive term for each thing that the instruction text names. The "Do not use"
column lists the synonyms that must not replace the term. Use a word from that
column only as a literal name in inline code. Each row adds to the base glossary
of the ASD-STE100 skill and overrides it. `reference/herdr/STE-GLOSSARY.md` is the
glossary of the herdr reference folder. Where a term appears in both, the
meanings agree.

| Term | Meaning | Do not use |
| --- | --- | --- |
| hive | The tool that this repository builds: a Go CLI and TUI that coordinates coding agents in herdr. | |
| herdr | The terminal workspace manager that runs the agents. `reference/herdr/` describes its socket API. | |
| home | The hive checkout: the directory that holds the personal data, the project registry, and the central state. The tracked file `.hive/README.md` marks it. | |
| hive session | The one herdr session that hosts the home, the coordinator, and the supervisor. | |
| project | One registered Git repository with a stable project ID. | repo |
| project directory | The directory `projects/<project-id>/` in the home: the working directory of the planner of that project. | project context directory |
| worktree | A linked Git worktree of a project at `projects/<project-id>/worktrees/<task-id>/`, shown by herdr as a workspace: the working directory of one builder. | task checkout, worktree checkout |
| role | The duty of a session under the role gate. The coordinator is the one built-in agent. | |
| profile | A tracked Markdown file `profiles/NAME.md` in the home. It tells an agent how to do one kind of work. The agent record keeps its name as a free label. | |
| fact | One field of the agent record that hive uses for a mechanic: `project`, `task`, `worktree`, `owns`, `parent`, or `kind`. | |
| owner | The one live agent that owns a project. Mail for the project goes to it. | |
| built-in kind | The `kind` field of a record that hive code starts. Its values are `coordinator` and `supervisor`. | |
| planner | The first profile that plans and lands the work of one project. | |
| builder | The first profile that does one task in a worktree. | |
| coordinator | The one built-in agent that talks to the user. It runs in the root of the home, in the hive session. | orchestrator |
| verifier | The independent agent that re-runs the checks of one task and records the result. | reviewer |
| agent | A coding-agent process in a herdr pane, such as Claude Code or Codex. | bot, assistant |
| harness | The kind of agent program: Claude Code, Codex, Pi, Cursor agent, or OpenCode. | |
| task | One unit of tracked work with a brief and a durable record. | work unit, ticket |
| brief | The Markdown text that gives a task or a profile its scope. | |
| verify | Record a passed independent check of a task by an agent that does not own the task. | |
| check | Run the gate, or test a behavior yourself. | validate, confirm |
| grant | A standing user approval for named delivery actions on one project and named branches. | |
| preset | One entry in the presets document: a purpose in prose and an ordered list of harness and model choices. | |
| tier | One of strong, decent, or small: the class of model that a kind of task gets. | |
| memory | One durable fact in a Markdown file at home scope or at project scope. | |
| handoff | The Markdown summary of one agent session. It refers to the authoritative records. | |
| state store | The central records and their lock in `.hive/state/`. | database |
| claim | The record that gives one agent a singleton owner or built-in kind, with a generation number. | lease |
| inbox | The durable queue of inbox records of one inbox key in the state store. | mailbox |
| inbox record | One durable record in an inbox: completed, verified, blocked, exited, quiet, unresolved, or message. | note |
| nudge | One line that hive types into the pane of an idle agent, which tells it to read its inbox. | poke |
| supervisor | The one `hive supervise` process of a home. It relays herdr events into inbox records and nudges idle roles. | daemon, watchdog |
| log event | One JSON line in a file in `.hive/logs/`. | |
| spawn | Start a new agent. | launch |
| stop | End a live agent. | retire, kill |
| land | Merge a verified branch into its target under a grant. | ship |
| gate | The checks that a Go change passes before it is done. | |
| tracer bullet | A thin end-to-end slice through the CLI, the internal package, the state store, and the log. | spike |
| wave | The set of tasks that can run in parallel after the same merges. | sprint |
| checkpoint | A node in the task graph where the user sees a behavior live and records a critique before dependent tasks start. | demo gate, review gate |
| acceptance criterion | One provable statement in a brief that defines done. | |
| verification item | One step with a command that proves one or more acceptance criteria. | |
| evidence | A Markdown file in `.hive/spec/<feature>/tasks/<task-id>/evidence/` that records one verification item and its output. | proof |
| merger | The person or agent that merges a task branch into `dev` and updates the tracker. | |
