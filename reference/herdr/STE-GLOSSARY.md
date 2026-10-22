# herdr STE glossary

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Index: [README.md](README.md).

This glossary gives one meaning to each herdr term in this folder. It adds to the base STE glossary and overrides it. Use each term only with the meaning in this table. Do not use a word in the third column, except as a literal name in inline code.

| Term | Meaning | Do not use |
| --- | --- | --- |
| herdr | The terminal workspace manager for coding agents that this folder describes. | |
| server | The herdr process that owns the state of one session and answers requests on the socket of that session. | |
| client | A program that connects to a herdr server, such as the TUI, the `herdr` CLI, or a script that writes to the socket. | |
| TUI | The full-screen terminal interface of herdr, which is a client that attaches to a server and shows its session. | |
| foreground client | The attached TUI client that receives UI methods, such as `client.window_title.set`. | |
| caller | The process or connection that sends a request, and the calling pane when that process runs in a pane. | |
| machine | A computer that herdr can connect to: the local computer, or a remote computer that `herdr machine add` registers. | |
| session | A persistent herdr namespace with one server, one socket, and one state directory. | |
| default session | The session without a name, with its socket `herdr.sock` directly in the herdr config directory. | |
| named session | A session that `--session <name>` selects, with its files in `sessions/<name>/` in the herdr config directory. | |
| live session | A session that holds the real work of a person, so that only read-only methods can go to its socket. | primary session, main session |
| scratch server | A herdr server that a pass starts under its own session name for tests that change state, so that no live session changes. | test server |
| workspace | The top-level herdr container for tabs and panes, usually one for each repository or task. | |
| tab | A part of a workspace that holds one layout of panes. | |
| pane | One terminal area in the layout of a tab, which can hold an agent. | |
| cell | One character position on a terminal screen, the unit that herdr uses for pane and popup sizes. | |
| terminal | The terminal emulator that herdr runs for one pane, with its own terminal ID, shell process, screen, and scrollback. | |
| shell | The command interpreter process that the terminal of a pane runs, such as a login shell. | |
| layout | The tree of splits that sets the position and the size of each pane in a tab. | |
| split | A layout node that divides an area between two children, as `pane.split` makes. | |
| popup | A pane that herdr shows as an overlay above the layout. | |
| ID | An opaque string handle for a workspace, tab, or pane, such as `w1`, `w1:t1`, or `w1:p1`, that you read from a response. | |
| terminal ID | The `term_…` ID of the terminal of a pane, which agent targets do not accept. | |
| label | Text that a rename puts on a workspace, tab, or pane, which is not an ID and can occur on two items. | |
| display | A modifier for the visible form of a value in the herdr UI, such as a display name or number, not for the action. | |
| agent | A coding-agent process, such as Claude Code or Codex, that herdr detects in a pane or that reports itself in a pane. | bot, assistant |
| harness | The base-glossary word for a program that runs a model, which this folder calls an agent when it runs in a pane. | |
| agent kind | The canonical name of an agent program, such as `claude` or `codex`, in the `agent` field. | agent type |
| agent name | A unique live name that a caller gives to an agent in a pane, used as a target for agent methods. | |
| agent status | The lifecycle value that herdr gives to an agent: `idle`, `working`, `blocked`, `done`, or `unknown`. | agent state |
| agent session | The native session of an agent program, such as a conversation ID, that herdr links to a pane for restore. | |
| agent manifest | A set of rules that tells herdr how to detect one agent kind and its agent status. | |
| detect | Find that a pane runs an agent of a known agent kind. | |
| report | Give the server data about a pane, a workspace, or its agent with a `*.report_*` method. | |
| release | End the link between a pane and its agent, as `pane.release_agent` does, so that the pane has no agent. | |
| seen | The server state that records if a person saw a tab in the focused herdr UI, which makes `done` different from `idle`. | |
| focus | Make a workspace, tab, or pane the active one in the herdr UI, which also marks the target seen. | |
| attach | Connect a TUI client to a running server, so that the client shows the session of that server. | |
| detach | Disconnect a TUI client from a server, while the server and its panes continue to run. | |
| confirmation | A herdr TUI dialog that asks a person to accept an action before herdr does it, such as a close. | |
| notification | A desktop or toast message that herdr shows to a person, such as with `notification.show`. | |
| worktree | A Git worktree that herdr lists, creates, opens, or removes with the `worktree.*` methods. | |
| checkout | The directory of a Git working copy: the primary checkout of a repository, or the directory of a linked worktree. | |
| integration | The herdr configuration for a supported agent program that `integration.install` writes into the configuration of that program. | |
| plugin | An extension that herdr loads from a plugin manifest, which can add actions, panes, event hooks, and link handlers. | |
| socket | The Unix domain socket file through which clients send requests to the server of one session. | |
| connection | One client connection to the socket, which carries one request and its response, or one subscription. | |
| method | A named operation of the socket API, such as `pane.list`, that a request gives in its `method` field. | |
| CLI command | A `herdr` subcommand that a person or a script runs in a shell, such as `herdr pane list`. | |
| request | One JSON line that a client writes to the socket, with the fields `id`, `method`, and `params`. | |
| response | One JSON line that the server writes back for a request, in the success envelope or the error envelope. | reply |
| envelope | The outer shape of a JSON line on the socket: the request, success, error, or event envelope. | |
| event | A JSON line that the server pushes on a subscription connection, in the event envelope and with no request `id`. | |
| event kind | The value that names the type of an event in the event catalog. | |
| subscription | A connection that stays open after `events.subscribe` and receives the events that match its filters. | |
| start | Make a process, server, session, or agent begin, such as with `herdr server` or `agent.start`. | initiate |
| launch | Start a program from a terminal command line, such as the TUI or an agent, which is pending until herdr detects the agent. | |
| stop | Make a herdr server or another running process end in an orderly sequence, such as with `server.stop` or `herdr session stop`. | halt, terminate |
| kill | End a process with an operating-system signal, without an orderly stop. | |
| close | Remove a workspace, tab, pane, popup, or connection, which for a pane also ends its terminal process. | |
| remove | Make an item no longer exist in herdr state or in a list, such as a worktree with `worktree.remove`. | erase |
| delete | Remove files or directories from the disk, such as the directory of a stopped session with `herdr session delete`. | |
| get | Read one item or one value from a source, such as one pane with `pane.get`. | retrieve, obtain |
| fetch | Download data over the network, such as a plugin, agent manifests, or a page from `herdr.dev`. | |
| show | Make information visible to a person, in the herdr UI or as CLI output, such as with `notification.show`. | |
| pass | One run of the `RESYNC.md` procedure that checks the pages of this folder against one herdr version. | |
| probe | A request or a CLI command that a pass sends to a herdr server to check a claim. | |
| capture | A stored copy of real output from a herdr server or the herdr CLI. | |
| stamp | A line that records the evidence for a claim: a Checked, Derived, or Documented stamp. | |
| Checked (stamp) | The stamp `Checked YYYY-MM-DD against herdr X.Y.Z`: a pass tested the claim with read-only methods on a live session or other methods on a scratch server. | |
| Derived (stamp) | The stamp `Derived from schema; not live-checked (YYYY-MM-DD, herdr X.Y.Z)`: the claim comes only from the shapes in `raw/schema.json`, and no pass tested it on a server. | constructed |
| Documented (stamp) | The stamp `Documented by herdr <source> (herdr X.Y.Z); not live-checked: <reason>`: the claim comes from text that herdr publishes, and no pass tested it on a server. | |
