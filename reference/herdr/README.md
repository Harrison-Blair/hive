# herdr API reference

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Procedure: [RESYNC.md](RESYNC.md). Glossary: [STE-GLOSSARY.md](STE-GLOSSARY.md).

herdr is a terminal workspace manager for coding agents. A herdr server keeps workspaces, tabs, panes, and agents for one session. The server accepts JSON requests on a Unix domain socket. It does not use HTTP. The `herdr` CLI is a client of the same socket.

These pages describe the socket API of herdr for any client. A client can be a script, a tool, or an agent in a pane. Each page gives methods, parameters, results, errors, events, and examples.

The schema in `raw/schema.json` is the contract. If a page and the schema do not agree, the schema is correct, and the page has an error.

## Reading order

Read these pages first, in this sequence:

1. [protocol.md](protocol.md) gives the transport, the framing, and the envelopes. It also gives the rule of one request for each connection, the ping handshake, and the version check.
2. [addressing.md](addressing.md) gives the ID scheme (`w1`, `w1:t1`, `w1:p1`) and the rules for targets.
3. [environment.md](environment.md) gives the access model: the socket path, the environment variables, and the session directories. herdr has no token authentication.
4. [data-model.md](data-model.md) gives each data type, such as `WorkspaceInfo`, `PaneInfo`, `AgentInfo`, `AgentStatus`, and `LayoutNode`.
5. [errors.md](errors.md) gives the error envelope and the error codes.
6. [events.md](events.md) gives `events.subscribe`, `events.wait`, the subscription rules, and the event catalog.
7. [cli-mapping.md](cli-mapping.md) gives the CLI command for each method, the CLI-only commands, and the CLI conventions.

Then read the method page for the methods that you use.

## Method reference (102 methods)

| File | Namespace | Methods |
|---|---|---|
| [api/agent.md](api/agent.md) | `agent.*` | 12 |
| [api/pane.md](api/pane.md) | `pane.*` | 35 |
| [api/workspace.md](api/workspace.md) | `workspace.*` | 9 |
| [api/tab.md](api/tab.md) | `tab.*` | 7 |
| [api/worktree.md](api/worktree.md) | `worktree.*` | 4 |
| [api/layout.md](api/layout.md) | `layout.*` (API only, with no CLI group) | 3 |
| [api/plugin.md](api/plugin.md) | `plugin.*` (the `herdr plugin` CLI group is new in 0.9.1) | 11 |
| [api/server.md](api/server.md) | `server.*` | 6 |
| [api/session.md](api/session.md) | `session.snapshot`, `ping` | 2 |
| [api/integration.md](api/integration.md) | `integration.*` | 3 |
| [api/ui.md](api/ui.md) | `notification.show`, `popup.close`, `client.window_title.*`, `client_shell.surface.set`, `command.invoke`, `product_announcement.dismiss`, `release_notes.dismiss` | 8 |
| [events.md](events.md) | `events.*` | 2 |

Each method has a `## <method>` heading on one page only. [RESYNC.md](RESYNC.md) gives the check for this rule.

## Raw artifacts

The `raw/` directory holds the original output of herdr 0.9.5 and of `herdr.dev`. These files have no changes. This folder does not apply its language rules to them.

| File | Source |
|---|---|
| [raw/schema.json](raw/schema.json) | `herdr api schema --json`. This file is the machine-readable contract. |
| [raw/skill.md](raw/skill.md) | `herdr --skill`. This file is the agent guide that the herdr binary contains. |
| [raw/default-config.toml](raw/default-config.toml) | `herdr --default-config`. |
| [raw/agent-guide.md](raw/agent-guide.md) | A copy of `https://herdr.dev/agent-guide.md`. |
| [raw/llms.txt](raw/llms.txt) | A copy of `https://herdr.dev/llms.txt`. |

## Evidence stamps

Parameters, results, and types come from `raw/schema.json` for herdr 0.9.5. Each claim and each example has a stamp, or a note from the last paragraph of this section. A stamp records the evidence for the claim. A stamp has one of these three forms:

| Stamp | Meaning |
|---|---|
| `Checked YYYY-MM-DD against herdr X.Y.Z` | A pass tested the claim on that date. Read-only methods went to a live session. Other methods went to a scratch server. |
| `Derived from schema; not live-checked (YYYY-MM-DD, herdr X.Y.Z)` | The claim comes only from the shapes in `raw/schema.json`. No pass tested it on a server. |
| `Documented by herdr <source> (herdr X.Y.Z); not live-checked: <reason>` | The claim comes from text that herdr publishes. The source is `--skill`, `--default-config`, `agent-guide`, `--help`, or `<cmd> --help`. The `--help` sources are the help output of the installed herdr binary. The reason tells why no pass tested it. |

Pages show the Derived and Documented stamps in inline code. A Checked stamp can show as plain text or in inline code.

A section has a Checked stamp only for the parts that the pass tested. Thus two sections of one page can have different stamps.

Some claims have a note about an earlier herdr version, such as "changed in 0.9.3". Some claims have a note that the pass did not check them, with the reason. These notes are history or limits. They are not stamps.

## Examples

Each example comes from a capture of a herdr server or of the herdr CLI. The examples are anonymous. They use these placeholder values:

| Item | Placeholder |
|---|---|
| Home directory | `/home/user` |
| Project directory | `/home/user/project` |
| Host name | `host` |
| User name | `user` |
| Session name | `demo` or `demo-<purpose>` |
| Labels | Neutral words, such as `api-docs`, `build`, or `reviewer` |
| Request IDs | `1`, `2`, and so on, or `cli:<group>:<cmd>` as herdr writes them |

The examples keep the IDs that herdr makes, such as `w1:p1` and `term_…`. They also keep the agent kinds as herdr shows them.

## Language

The prose of these pages follows ASD-STE100 Simplified Technical English, Strict level. The rules apply to table cells too. They do not apply to fenced code, inline code, or `raw/`.

[STE-GLOSSARY.md](STE-GLOSSARY.md) gives the one meaning of each herdr term in this folder. Use each term only with that meaning.

## Changes in herdr

herdr can update itself with `herdr update`. Thus a new herdr version can make these pages incorrect. These pages describe herdr 0.9.5, protocol 22. [RESYNC.md](RESYNC.md) gives the procedure to check the pages against a new version.
