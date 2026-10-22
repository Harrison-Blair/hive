# herdr API: CLI ↔ socket method mapping

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Index: [README.md](README.md). Wire format: [protocol.md](protocol.md).

The `herdr` CLI is a thin client of the socket protocol that this reference describes. Most method-backed CLI commands parse their flags and send one request. Then they print the response envelope as one JSON line on stdout. Some CLI commands print nothing when they succeed (see [CLI conventions](#cli-conventions)).

This page is the full cross-reference. It has a row for each of the 102 socket methods, with its CLI command or the mark `API-only`. It also lists the CLI commands that have no single socket method, and it gives the conventions of the CLI.

Sources: the `--help` text of each CLI command, the usage list of each command group, and probes on scratch servers. The probes include the usage strings that the CLI writes to stderr. If a CLI command does not exist in a known group, herdr prints the usage list of that group, for example for `herdr tab move`. For an unknown top-level group, herdr prints `unknown command: layout`.

## Mapping table (all 102 socket methods)

The table lists the methods in the alphabetical order of the method name. "API-only" means that no CLI command sends the method. A client must write the request to the socket directly. The flag lists are short. For the full set of flags, read the `--help` of each CLI command.

| socket method | CLI equivalent | notes |
| --- | --- | --- |
| `agent.explain` | `herdr agent explain [TARGET] [--file PATH] [--agent LABEL] [--json] [--format text\|json] [-v]` | TARGET is optional. With `--file PATH --agent LABEL`, the CLI explains the detection for a screen file. This form does not use the socket, and it works when no server runs. `--file` without `--agent` is a usage error (exit 2). The default output is text. With `--file`, `--json` prints a bare JSON object with no `id` or `result`. An unknown target gives the error `agent_not_found`. |
| `agent.focus` | `herdr agent focus <target>` | The target is a pane ID or an agent name. |
| `agent.get` | `herdr agent get <target>` | |
| `agent.list` | `herdr agent list` | |
| `agent.prompt` | `herdr agent prompt <target> <text> [--wait] [--until STATE]… [--timeout MS]` | `--wait`, `--until`, and `--timeout` are CLI polling for a settle status, on top of the method response. STATE is one of idle, working, blocked, done, unknown. If the agent is already `blocked`, the server rejects the prompt with `agent_blocked` and sends no input. For the stall and timeout rules, see [api/agent.md](api/agent.md#agentprompt). |
| `agent.read` | `herdr agent read <target> [--source SRC] [--lines N] [--format text\|ansi] [--ansi]` | SRC is one of visible, recent, recent-unwrapped, detection. The default is `recent`. |
| `agent.rename` | `herdr agent rename <target> <name>\|--clear` | Give a positional NAME or `--clear`. While the start of the agent is pending, the server rejects a rename with `agent_launch_pending`. |
| `agent.send_keys` | `herdr agent send-keys <target> <key>…` | `esc` is the canonical name of the Escape key. herdr also accepts `escape`. |
| `agent.start` | `herdr agent start <name> --kind KIND --pane ID [--timeout MS] [-- AGENT_ARG…]` | KIND is a supported agent kind. The `--help` text lists 24 kinds: `pi, claude, codex, gemini, cursor, devin, agy, cline, omp, mastracode, opencode, copilot, kimi, kiro, droid, amp, grok, hermes, kilo, qodercli, qwen, letta, maki, muse`. The pane must show an interactive shell prompt. The `--help` text gives a `--timeout` default of 30000 and a maximum of 300000. Arguments after `--` go to the agent executable. For the line limit on macOS, see [api/agent.md](api/agent.md#line-limit-on-macos). |
| `agent.view.clear` | API-only (no CLI subcommand) | Raw socket only. See [api/agent.md](api/agent.md#agentviewclear). |
| `agent.view.set` | API-only (no CLI subcommand) | Raw socket only. See [api/agent.md](api/agent.md#agentviewset). |
| `agent.wait` | `herdr agent wait <target> [--until STATE]… [--timeout MS]` | Without `--until`, the wait ends at idle, done, or blocked. |
| `client.window_title.clear` | API-only (no CLI subcommand) | |
| `client.window_title.set` | API-only (no CLI subcommand) | Raw socket only. See [api/ui.md](api/ui.md#clientwindow_titleset). |
| `client_shell.surface.set` | API-only (no CLI subcommand) | Sets if the client shell endpoint that sends the request gets and controls the pane presentation. Outside a client shell endpoint, the server returns `connection_local_only`. |
| `command.invoke` | API-only (no CLI subcommand) | Runs an opaque `command_id` from the command projection of the client shell. No CLI command sends it. An unknown ID gives the error `command_not_found`. |
| `events.subscribe` | API-only (no CLI subcommand) | The connection stays open and receives pushed `{"event":…,"data":…}` lines. See [events.md](events.md#eventssubscribe). |
| `events.wait` | API-only (no CLI subcommand) | Raw socket only. See [events.md](events.md). |
| `integration.install` | `herdr integration install <target>` | The target is one of `pi, omp, claude, codex, copilot, devin, droid, kimi, opencode, kilo, hermes, qodercli, qwen, cursor, mastracode, antigravity-cli, grok, letta`. The CLI usage text accepts `letta`. The `IntegrationTarget` enum of the schema and the `integration.list` response do not have `letta`. This drift occurs in 0.9.1, 0.9.3, and 0.9.5. The schema spells `antigravity_cli` with an underscore, but the CLI spells `antigravity-cli`. |
| `integration.list` | API-only (no CLI subcommand) | Returns one `IntegrationInfo`/`IntegrationState` entry for each built-in integration. The response has 17 entries in 0.9.1, 0.9.3, and 0.9.5, and `letta` is not one of them. `herdr integration status` is a different report (see [CLI-only commands](#cli-only-commands)). |
| `integration.uninstall` | `herdr integration uninstall <target>` | The target set is the same as for `integration install`. |
| `layout.apply` | API-only (no CLI subcommand) | `herdr layout …` prints `unknown command: layout`. |
| `layout.export` | API-only (no CLI subcommand) | Raw socket only. See [api/layout.md](api/layout.md#layoutexport). |
| `layout.set_split_ratio` | API-only (no CLI subcommand) | Raw socket only. See [api/layout.md](api/layout.md#layoutset_split_ratio). |
| `notification.show` | `herdr notification show <title> [--body TEXT] [--position POS] [--sound SOUND]` | POS is one of top-left, top-right, bottom-left, bottom-right. SOUND is one of none, done, request. |
| `pane.clear` | API-only (no CLI subcommand) | New in 0.9.3. It clears the screen and the scrollback of a pane. `herdr pane clear` prints the `pane` usage list, with exit 2. In the UI, the `clear_pane` keybinding does the same thing, and it has no default key. |
| `pane.clear_agent_authority` | API-only (no CLI subcommand) | It is a different method from `pane release-agent`. Raw socket only. See [api/pane.md](api/pane.md#paneclear_agent_authority). |
| `pane.close` | `herdr pane close <pane_id>` | For the TUI confirmation of the last pane, see [api/pane.md](api/pane.md#paneclose). |
| `pane.copy_motion` | API-only (no CLI subcommand) | It moves a copy-mode cursor by a `PaneCopyMotion` (word, line, or paragraph). The CLI has no copy mode. |
| `pane.copy_search` | API-only (no CLI subcommand) | It searches the scrollback from a copy-mode cursor in a `PaneCopySearchDirection`. The CLI has no copy mode. |
| `pane.current` | `herdr pane current [--pane ID \| --current]` | It takes no positional pane ID. `pane current w1:p1` prints `unknown option: w1:p1`, with exit 2. |
| `pane.edges` | `herdr pane edges [--pane ID \| --current]` | |
| `pane.edit_scrollback` | API-only (no CLI subcommand) | It takes a bare `PaneTarget`. It opens the scrollback of the pane for edits in an external editor. |
| `pane.focus` | API-only (no CLI subcommand) | It takes a `PaneTarget` and focuses that pane. The CLI `pane focus` is directional and sends `pane.focus_direction`. |
| `pane.focus_direction` | `herdr pane focus --direction DIR [--pane ID \| --current]` | DIR is one of left, right, up, down. The envelope ID is `cli:pane:focus`. A positional pane ID is a usage error (exit 2). |
| `pane.get` | `herdr pane get <pane_id>` | The pane ID is positional, not `--pane`. |
| `pane.input.set` | `herdr pane input --right-click TARGET [PANE_ID] [--pane ID \| --current]` | The CLI command is `pane input`. TARGET is one of herdr, pane, and it sets the routing of a right click. An error has the envelope ID `cli:pane:input:set`. |
| `pane.layout` | `herdr pane layout [--pane ID \| --current]` | |
| `pane.link.activate` | API-only (no CLI subcommand) | It activates a detected link region at a viewport row and column (params `viewport_row` and `col`). When no link is there, the result is `{"type":"pane_link_activated","handled":false}`. |
| `pane.link.resolve` | API-only (no CLI subcommand) | It resolves the detected `PaneLinkRegion`s at a viewport row and column, and it does not activate them. |
| `pane.list` | `herdr pane list [--workspace ID]` | |
| `pane.move` | `herdr pane move <pane_id> [--tab ID\|--new-tab\|--workspace ID\|--new-workspace] [--split right\|down] [--target-pane ID] [--ratio F] [--label TEXT] [--tab-label TEXT] [--focus\|--no-focus]` | The pane ID is positional, and flags select the destination. The usage list gives three forms: `--tab <tab_id> --split right\|down [--target-pane ID] [--ratio FLOAT]`, `--new-tab [--workspace ID] [--label TEXT]`, and `--new-workspace [--label TEXT] [--tab-label TEXT]`. |
| `pane.neighbor` | `herdr pane neighbor --direction DIR [--pane ID \| --current]` | DIR is one of left, right, up, down. |
| `pane.process_info` | `herdr pane process-info [--pane ID \| --current]` | The envelope ID is `cli:pane:process_info`, with an underscore. |
| `pane.read` | `herdr pane read <pane_id> [--source SRC] [--lines N] [--format text\|ansi] [--ansi] [--raw]` | The pane ID is positional. SRC is the same as for `agent.read`. `pane read --help` lists `detection`, but the usage list of the `pane` group does not. An unknown source prints `invalid read source: <SRC>`, with exit 2. |
| `pane.release_agent` | `herdr pane release-agent <pane_id> --source ID --agent LABEL [--seq N]` | Success prints nothing. An error has the envelope ID `cli:request`. |
| `pane.rename` | `herdr pane rename <pane_id> [LABEL…] \| --clear` | |
| `pane.report_agent` | `herdr pane report-agent <pane_id> --source ID --agent LABEL --state STATUS [--message TEXT] [--seq N] [--agent-session-id ID] [--agent-session-path PATH] [-- RESUME_ARG…]` | STATUS is one of idle, working, blocked, unknown. The arguments after `--` become `resume_argv` (new in 0.9.3). The first resume argument must be a plain command name, not a path. Otherwise the server returns `invalid_resume_argv`. Success prints nothing. An error has the envelope ID `cli:request`, with or without resume arguments. The usage list of the `pane` group does not show the `-- RESUME_ARG` part. |
| `pane.report_agent_session` | `herdr pane report-agent-session <pane_id> --source ID --agent LABEL [--seq N] [--agent-session-id ID] [--agent-session-path PATH] [--session-start-source SOURCE] [-- RESUME_ARG…]` | The arguments after `--` become `resume_argv` (new in 0.9.3). Success prints nothing. An error has the envelope ID `cli:request`. The usage list of the `pane` group does not show the `-- RESUME_ARG` part. |
| `pane.report_metadata` | `herdr pane report-metadata <pane_id> --source ID [--agent LABEL] [--applies-to-source ID] [--title TEXT\|--clear-title] [--display-agent TEXT\|--clear-display-agent] [--state-label STATUS=TEXT\|--clear-state-labels] [--token NAME=VALUE\|--clear-token NAME] [--seq N] [--ttl-ms N]` | The metadata changes only what the herdr UI shows in the overlay. Success prints nothing. An error has the envelope ID `cli:request`. |
| `pane.resize` | `herdr pane resize --direction DIR [--amount F] [--pane ID \| --current]` | DIR is one of left, right, up, down. `--amount` takes a float. |
| `pane.scroll` | API-only (no CLI subcommand) | It sets an absolute scrollback `offset_from_bottom`. The CLI has no scroll command. |
| `pane.selection.read` | API-only (no CLI subcommand) | It reads the text between an anchor `PaneTextPoint` and a cursor `PaneTextPoint`. The CLI has no equivalent command. |
| `pane.send_input` | API-only (no CLI subcommand) | Raw socket only. The `send-text` help names `pane run` for text and Enter in one call. Thus `pane run` uses `send-text` and not this method. No pass saw the request that `pane run` sends. |
| `pane.send_keys` | `herdr pane send-keys <pane_id> <key>…` | `esc` is the canonical name of the Escape key. herdr also accepts `escape` and `Escape`. Success prints nothing. An error has the envelope ID `cli:request`. |
| `pane.send_text` | `herdr pane send-text <pane_id> <text>` | It sends literal text with no Enter at the end. Success prints nothing. An error has the envelope ID `cli:request`. |
| `pane.split` | `herdr pane split [PANE_ID] [--pane ID\|--current] --direction right\|down [--ratio F] [--cwd PATH] [--env KEY=VALUE] [--right-click herdr\|pane] [--focus\|--no-focus]` | The usage list shows `--direction` as required. A split without `--direction` is a usage error (exit 2). `--help` lists `--direction` with the options. `--env` sets an environment variable for the process in the new pane. An unknown pane gives the message `pane not found`, with no ID. |
| `pane.swap` | `herdr pane swap [--direction DIR \| --source-pane ID --target-pane ID] [--pane ID\|--current]` | Give a direction, or an explicit source and target pair. An unknown source pane is not an error. The CLI exits with 0, and the result has `"changed":false` and `"reason":"not_found"`. |
| `pane.wait_for_output` | `herdr pane wait-output <pane_id> <--match TEXT\|--regex PATTERN> [--source SRC] [--lines N] [--timeout MS] [--raw]` | SRC is one of visible, recent, recent-unwrapped, and the default is `recent`. The CLI also accepts `--source detection` without a usage error. The regex uses Rust syntax. The CLI searches the selected snapshot at once, with the existing output, and then polls. Without `--timeout`, it waits with no limit. At the timeout, the error is `timeout`. |
| `pane.zoom` | `herdr pane zoom [PANE_ID] [--pane ID\|--current] [--toggle\|--on\|--off]` | |
| `ping` | API-only (no CLI subcommand) | The CLI has no `ping` verb. The `herdr status` output has the version and capability fields of the `ping` result. Thus `herdr status` probably does the ping/pong handshake internally. |
| `plugin.action.invoke` | `herdr plugin action invoke <ACTION_ID> [--plugin <ID>]` | New CLI group in 0.9.1. |
| `plugin.action.list` | `herdr plugin action list [--plugin <ID>]` | |
| `plugin.disable` | `herdr plugin disable <PLUGIN_ID>` | New CLI group in 0.9.1. |
| `plugin.enable` | `herdr plugin enable <PLUGIN_ID>` | New CLI group in 0.9.1. |
| `plugin.link` | `herdr plugin link <PATH> [--enabled \| --disabled]` | New CLI group in 0.9.1. `--help` lists `--enabled` and `--disabled`. The usage list shows only `[--disabled]`. |
| `plugin.list` | `herdr plugin list [--plugin <ID>] [--json]` | The default output without `--json` is text for a person, for example `No plugins installed.`. The sibling commands `plugin action list` and `plugin log list` print the JSON envelope with no flag. |
| `plugin.log.list` | `herdr plugin log list [--plugin <ID>] [--limit <N>]` | New CLI group in 0.9.1. `logs` is an alias of `log`. |
| `plugin.pane.close` | `herdr plugin pane close <PANE_ID>` | New CLI group in 0.9.1. |
| `plugin.pane.focus` | `herdr plugin pane focus <PANE_ID>` | New CLI group in 0.9.1. |
| `plugin.pane.open` | `herdr plugin pane open --plugin <ID> --entrypoint <ID> [--placement <overlay\|split\|tab\|zoomed>] [--workspace <ID>] [--target-pane <PANE>] [--direction <right\|down>] [--cwd <PATH>] [--env <KEY=VALUE>]... [--focus \| --no-focus]` | The CLI help does not list `popup` for `--placement`, but the CLI accepts `--placement popup`. `popup` is also valid on the wire. An unknown placement prints `invalid pane placement: <value>`, with exit 2. |
| `plugin.unlink` | `herdr plugin unlink <PLUGIN_ID>` | New CLI group in 0.9.1. For an unknown plugin, the result has `"removed":false`, and the CLI exits with 0. The call writes the plugin registry files `plugins.json` and `.plugins.lock` also when it removes nothing. |
| `popup.close` | API-only (no CLI subcommand) | Raw socket only. See [api/ui.md](api/ui.md#popupclose). |
| `product_announcement.dismiss` | API-only (no CLI subcommand) | It dismisses an in-app product announcement by `id` and `version`. When the announcement is no longer current, the server returns `stale_announcement`. |
| `release_notes.dismiss` | API-only (no CLI subcommand) | It dismisses the in-app release-notes prompt for a given `version`. A successful call writes `release-notes.json` in the shared herdr config directory, also from the server of a named session. See [api/ui.md](api/ui.md#release_notesdismiss). |
| `server.agent_manifests` | `herdr server agent-manifests [--json]` | A thin 1:1 wrapper. The default output without `--json` is a text table, not the JSON envelope. `--json` prints the envelope with the ID `cli:server:agent-manifests`. |
| `server.live_handoff` | `herdr server live-handoff [--import-exe PATH] [--expected-protocol N] [--expected-version VERSION]` | The usage list of the `server` group shows this verb as `hand off live panes to a new local server`. `herdr server --help` does not list it. `herdr server live-handoff --help` prints the one-line usage, with exit 2. `herdr update --handoff` and attach with `--remote` or `--handoff` also start a handoff. On success, the CLI prints `live handoff complete; server log: <path>` on stderr, with exit 0 and no JSON. For a named session, `<path>` is the log of the default session, but the new server writes to the log of its own session. See [api/server.md](api/server.md#serverlive_handoff). |
| `server.reload_agent_manifests` | `herdr server reload-agent-manifests` | It reloads only the local manifest overrides. |
| `server.reload_config` | `herdr server reload-config` | |
| `server.ssh_agent.register` | API-only (no CLI subcommand) | New in 0.9.3. The schema description says that it registers the absolute `socket_path` of an SSH agent on a remote host while the API connection stays open. `ping` shows the capability `ssh_agent_registration`. `herdr server ssh-agent …` prints the `server` usage list, with exit 2. |
| `server.stop` | `herdr server stop` | Success prints nothing. |
| `session.snapshot` | `herdr api snapshot` | The CLI request ID is `cli:api:snapshot`. |
| `tab.close` | `herdr tab close <tab_id>` | |
| `tab.create` | `herdr tab create [--workspace ID] [--cwd PATH] [--label TEXT] [--env KEY=VALUE] [--focus\|--no-focus]` | |
| `tab.focus` | `herdr tab focus <tab_id>` | |
| `tab.get` | `herdr tab get <tab_id>` | |
| `tab.list` | `herdr tab list [--workspace ID]` | |
| `tab.move` | API-only (no CLI subcommand) | `herdr tab move` prints the `tab` usage list, with exit 2. |
| `tab.rename` | `herdr tab rename <tab_id> <label>…` | LABEL is variadic, and the CLI joins the words. The CLI has no `--clear`, so `--clear` becomes the label text. |
| `workspace.close` | `herdr workspace close <workspace_id> [--group]` | `herdr workspace close --help` does not show `--group`. The usage list of the `workspace` group shows it, and the server accepts it. |
| `workspace.create` | `herdr workspace create [--cwd PATH] [--label TEXT] [--env KEY=VALUE] [--focus\|--no-focus]` | |
| `workspace.focus` | `herdr workspace focus <workspace_id>` | |
| `workspace.get` | `herdr workspace get <workspace_id>` | |
| `workspace.list` | `herdr workspace list` | |
| `workspace.move` | API-only (no CLI subcommand) | `herdr workspace move` prints the `workspace` usage list, with exit 2. |
| `workspace.move_block` | API-only (no CLI subcommand) | Raw socket only. See [api/workspace.md](api/workspace.md#workspacemove_block). |
| `workspace.rename` | `herdr workspace rename <workspace_id> <label>…` | LABEL is variadic. The CLI has no `--clear`, so `--clear` becomes the label text. |
| `workspace.report_metadata` | `herdr workspace report-metadata <workspace_id> --source ID [--token NAME=VALUE] [--clear-token NAME] [--seq N] [--ttl-ms N]` | Success prints nothing. An error has the envelope ID `cli:request`. |
| `worktree.create` | `herdr worktree create [--workspace ID \| --cwd PATH] [--branch NAME] [--base REF] [--path PATH] [--label TEXT] [--focus\|--no-focus] [--trust-repository]` | `--trust-repository` maps to the `trust_repository` param. The herdr skill text says to use it only after the user checked the repository. |
| `worktree.list` | `herdr worktree list [--workspace ID \| --cwd PATH] [--trust-repository]` | Outside a Git work tree, the error is `not_git_worktree`. |
| `worktree.open` | `herdr worktree open [--workspace ID \| --cwd PATH] (--path PATH \| --branch NAME) [--label TEXT] [--focus\|--no-focus] [--trust-repository]` | Give one of `--path` and `--branch`. |
| `worktree.remove` | `herdr worktree remove --workspace ID [--force] [--trust-repository]` | The CLI requires `--workspace`. A bare `worktree remove` prints its usage, with exit 2. `--path` prints `unknown option: --path`, with exit 2, but `create` and `open` accept `--path`. |

Checked 2026-10-17 against herdr 0.9.5:

- The method set is the same as the 102 request methods in `schema.json`.
- A pass read the usage list of each command group and the `--help` of each subcommand again. All flag lists come from that read.
- The pass ran the CLI commands of the table on a scratch server, with these exceptions. It did not run `worktree create`, `worktree open`, `integration install`, and `integration uninstall`. A rename with `agent rename` did not succeed, because the agent stopped first.
- The pass sent `client_shell.surface.set`, `command.invoke`, `pane.clear`, `pane.link.activate`, `ping`, `product_announcement.dismiss`, and `release_notes.dismiss` as raw requests.
- The pass checked that no CLI command sends the 31 API-only methods.
- The `clear_pane` key note comes from the `herdr --default-config` output of 0.9.5.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the behavior notes of the API-only methods that the pass did not send. These methods are `agent.view.*`, `client.window_title.*`, `events.*`, `layout.export`, `layout.set_split_ratio`, `popup.close`, `workspace.move_block`, `pane.clear_agent_authority`, `pane.send_input`, `pane.copy_*`, `pane.edit_scrollback`, `pane.scroll`, `pane.selection.read`, `pane.link.resolve`, and `server.ssh_agent.register`.

`Documented by herdr --help (herdr 0.9.5); not live-checked:` the handoff from `herdr update --handoff` and from an attach with `--remote` or `--handoff`. Each replaces the running server or the binary.

`Documented by herdr --skill (herdr 0.9.5); not live-checked:` the `--trust-repository` advice. It is guidance for a person, and no probe can test it.

### API-only method count

31 of the 102 methods have no CLI subcommand. The other 71 methods have a CLI command. The 31 API-only methods are:

- All 3 `layout.*` methods, and both methods of each of `pane.link.*`, `agent.view.*`, `client.window_title.*`, and `events.*`.
- `pane.clear`, `pane.focus`, `pane.send_input`, `pane.clear_agent_authority`, `pane.copy_motion`, `pane.copy_search`, `pane.edit_scrollback`, `pane.scroll`, and `pane.selection.read`.
- `popup.close`, `ping`, `client_shell.surface.set`, `command.invoke`, `integration.list`, `product_announcement.dismiss`, and `release_notes.dismiss`.
- `server.ssh_agent.register`, `tab.move`, `workspace.move`, and `workspace.move_block`.

herdr 0.9.1 had 33 API-only methods. 0.9.3 removed the three `pane.graphics.*` methods, and it added `pane.clear` and `server.ssh_agent.register`. Also, `server.live_handoff` got the `herdr server live-handoff` verb in 0.9.3. 0.9.5 has the same 31 API-only methods as 0.9.3.

Checked 2026-10-17 against herdr 0.9.5: a pass compared the 31 methods with the usage list of each command group and the top-level `--help`. `herdr pane clear`, `herdr server ssh-agent`, `herdr server ssh-agent register`, and `herdr server register-ssh-agent` each print a usage list and exit with status 2.

## CLI-only commands

These CLI commands do not send one socket method. They use local files or local state, send many socket requests, or run an interactive attach loop. Some do work outside the socket protocol, such as a network fetch or a binary download.

| CLI command | what it does | relation to the socket API |
| --- | --- | --- |
| `herdr status [server\|client] [--json]` | It shows the status of the local client and of the running server. | It is not one method. The output has the version and capability fields of the `ping` result, so it probably uses the ping/pong handshake and the version negotiation. The default output is a text report. `--json` prints a bare JSON object with no envelope. |
| `herdr update [--handoff]` | It downloads and installs the latest binary. | It is a local and network operation, not a socket method. `--handoff` adds `server.live_handoff` after the install. An unknown option prints `usage: herdr update [--handoff]`, with exit 2. In 0.9.3, a run inside a herdr session refused with `` update failed: run `herdr update` outside herdr after detaching from the session ``. That run gave exit 0, and the binary did not change. The pass did not run it again in 0.9.5. |
| `herdr channel show` / `herdr channel set <stable\|preview>` | It reads or writes the update channel in the config. `channel show` prints the channel as text, for example `stable`. | It uses only the local config. It makes no socket call. |
| `herdr config check` | It checks `config.toml` and prints diagnostics, for example `config: ok`. | It reads a local config file. It is different from `server reload-config`, which applies the config in the running server. |
| `herdr config reset-keys` | It makes a backup of `config.toml` and removes the custom keybindings: `[keys]`, `[keys.indexed]`, and `[[keys.command]]`. | It changes a local config file. It prints the backup path and a restore command (see the example below this table). |
| `herdr completion <bash\|elvish\|fish\|powershell\|zsh>` | It prints a shell-completion script. | It is local only and needs no server. An unknown shell prints `unknown shell: <name>` and the usage, with exit 2. |
| `herdr session list [--json]` | It lists the persistent sessions, with the default session as `default`. | It reads the session directories and sockets locally. It is not a socket method. `--json` prints a bare JSON object with no envelope. |
| `herdr session attach <name>` | It attaches the terminal to a session. | It is an interactive attach loop, not a request and response method. |
| `herdr session stop <name> [--json]` | It stops a named session. The name `default` selects the default session. | It goes to the server of that session, and it can start or stop the process. It is not `server.stop` on the socket of the current environment. `--json` prints a bare JSON object with `"stopped":true`. |
| `herdr session delete <name> [--json]` | It deletes a stopped session. | It deletes the session directory locally. The text output is `deleted session <name>`. A second delete of the same session prints the same line and exits with 0. |
| `herdr api schema [--json] [--output PATH]` | It prints the schema that the binary contains, or writes it to a file. | It returns the compiled-in schema and needs no running server. Compare `api snapshot`, which sends `session.snapshot`. The default output is a text summary. `--output PATH` writes the JSON schema and prints `wrote API schema to <path>`. |
| `herdr server agent-manifests [--json]` | It shows the active agent manifests. | It is a thin 1:1 wrapper of `server.agent_manifests`. This table lists it only for a full `server` group. |
| `herdr server update-agent-manifests [--json]` | It fetches agent manifests over the network, and then reloads them. | The network fetch has no socket method. The reload step uses the manifest reload path. This command is a real composite and CLI-only. |
| `herdr server reload-agent-manifests` | It reloads the local manifest overrides. | It is a thin 1:1 wrapper of `server.reload_agent_manifests`. |
| `herdr plugin install [--ref REF] [-y\|--yes] <OWNER/REPO[/SUBDIR]>` | It installs a plugin from GitHub. | It fetches the plugin into the local plugin registry. The method set of `raw/schema.json` has no method for it. |
| `herdr plugin update [<plugin_id\|owner/repo[/subdir...]>...] [-y\|--yes]` | It updates the plugins that came from GitHub. | It fetches over the network into the local plugin registry. The method set of `raw/schema.json` has no method for it. |
| `herdr plugin uninstall <plugin_id\|owner/repo[/subdir...]>` | It uninstalls a plugin. | It changes the local plugin registry and has no socket method. An unknown plugin prints `plugin not installed: <id>` on stderr as plain text, with exit 1. |
| `herdr plugin config-dir <PLUGIN_ID>` | It prints the config directory path of a plugin, `$XDG_CONFIG_HOME/herdr/plugins/config/<id>`. | It only computes a local path, with no socket method. |
| `herdr agent attach <target> [--takeover]` | It attaches the terminal directly to an agent pane. | It is an interactive attach loop, not a request and response method. |
| `herdr pane run <pane_id> <command>…` | It queues the command text and Enter together in one call. | It is a CLI composite of `pane send-text`. The `send-text` help says `herdr pane run … sends text and Enter in one call`. There is no `pane.run` method. See [pane run admission](#pane-run-admission). |
| `herdr integration status [--outdated-only]` | It shows the install status of the built-in integrations, one text line for each target. | No `integration.status` method exists. The CLI computes the status, and the `integration.list` method is different. The command has no `--json` flag. |
| `herdr machine list [--json]` / `status [<LABEL_OR_ID>] [--json]` / `reconnect <LABEL_OR_ID>` / `add [--label LABEL] [--remote-session NAME] <SSH_TARGET>` / `rename --label LABEL <PROFILE_ID>` / `remove\|enable\|disable <PROFILE_ID>` | It manages the saved SSH machine profiles for `--machine` and `--remote`. `status` (new in 0.9.3) checks the saved machines and asks for no authentication. `reconnect` (new in 0.9.3) authenticates a saved machine in the current terminal and checks the connectivity. The herdr skill text says that the user must run `reconnect`, because it needs SSH authentication. | It changes a local config file. `machine add`, `status`, and `reconnect` also connect over SSH. It has no socket method. `machine list --json` prints a bare JSON array. |

`integration status` prints lines in these forms:

```text
claude: current (v11) (/home/user/.claude/hooks/herdr-agent-state.sh)
<target>: not installed (<path>)
```

`config reset-keys` on a config file in `/home/user/project` prints this output. `<unix>` is a Unix timestamp:

```text
Created backup: /home/user/project/config.toml.bak-keybind-v2-<unix>
Removed [keys], [keys.indexed], and [[keys.command]] from /home/user/project/config.toml.
Built-in v2 keybindings will apply after Herdr restarts or reloads config.
If a Herdr server is running, run `herdr server reload-config` to apply this now.
To restore: cp /home/user/project/config.toml.bak-keybind-v2-<unix> /home/user/project/config.toml
```

The top-level launch and attach invocations are client entry points, not socket methods. These invocations are `herdr`, `herdr --session <name>`, `herdr --machine <label-or-id>`, `herdr --remote <target>`, `herdr --remote-keybindings <local\|server>`, `herdr --default-config`, `herdr --skill`, and `herdr --version`. `herdr server` with no subcommand runs the headless server itself.

`herdr --help` does not list the `plugin` group. It lists `machine`, `api`, `worktree`, and the other groups. But `herdr plugin` and all its subcommands work.

Checked 2026-10-17 against herdr 0.9.5:

- A pass read the usage list and the `--help` text of each command in the table.
- The pass ran `status`, `channel show`, `config check`, `config reset-keys` on a scratch config, `completion`, `session list`, `session stop`, and `session delete`.
- The pass also ran `api schema`, `server agent-manifests`, `server reload-agent-manifests`, `plugin uninstall`, `plugin config-dir`, `pane run`, `integration status`, and `machine list`.
- `plugin install`, `plugin update`, `plugin uninstall`, and `plugin config-dir` have no entry in the method set of `schema.json`.

`Documented by herdr --help (herdr 0.9.5); not live-checked:` the descriptions of the commands that the pass did not run. `session attach` and `agent attach` take over the terminal. `machine add`, `machine status`, and `machine reconnect` connect over SSH. `machine rename`, `machine remove`, `machine enable`, `machine disable`, and `channel set` write the shared config. `herdr update` replaces the binary. `server update-agent-manifests`, `plugin install`, and `plugin update` fetch over the network. The `default` name for `session stop` comes from the usage list, and the pass did not use it.

`Documented by herdr --skill (herdr 0.9.5); not live-checked:` the advice that the user must run `machine reconnect`. It needs SSH authentication by a person.

### pane run admission

`herdr pane run <pane_id> <command>…` queues the command text and Enter together. The CLI joins the command words with spaces. Success prints nothing and exits with status 0. Success means that herdr accepted the input for the pane. It does not mean that the shell ran the command, or that the command finished.

Thus a second `pane run` returns at once while the first command runs. The shell gets the second command as typed-ahead input and runs it after the first command. In this capture, `w1:p2` is a pane with an interactive shell:

```text
$ herdr pane run w1:p2 'sleep 4; echo SLEPT'     # no output, exit 0
$ herdr pane run w1:p2 'echo AFTER-RUN'          # no output, exit 0, returns at once
$ herdr pane read w1:p2 --source recent --lines 4
[user@host project]$ sleep 4; echo SLEPT
echo AFTER-RUN
# 5 seconds later:
SLEPT
[user@host project]$ echo AFTER-RUN
AFTER-RUN
```

Checked 2026-10-17 against herdr 0.9.5: the silent success, the joined words, and the admission behavior in the capture above.

On macOS, `pane run` and `agent start` reject a command that is longer than the canonical line-input limit that the terminal has at that time. They reject it before they queue it. Wait for the interactive shell prompt before you try again. The check applies only to this submission. It does not apply to an earlier incomplete line or to a later change of the terminal mode. For `agent start`, see [api/agent.md](api/agent.md#line-limit-on-macos).

In 0.9.3, the herdr skill text said that `pane run` sends the command text and Enter atomically.

`Documented by herdr --skill (herdr 0.9.5); not live-checked:` the line limit applies only on macOS, and the pass ran on Linux.

## CLI conventions

A pass saw these conventions in the help text, in probes on scratch servers, and in the usage strings on stderr.

- **Response envelope on stdout.** A method-backed CLI command prints the full response envelope as one JSON line on stdout, with `id` and `result`. The `id` is usually `cli:<group>:<cmd>`. Examples are `cli:pane:get`, `cli:workspace:create`, `cli:agent:list`, `cli:api:snapshot`, and `cli:pane:wait-output`. These CLI commands use a different `id`:
  - `pane process-info` uses `cli:pane:process_info`, with an underscore.
  - An error of `pane input` uses `cli:pane:input:set`.
  - Each `plugin` subcommand uses the flat ID `cli:plugin`, not `cli:plugin:<cmd>`. Thus a caller cannot use `id` to match concurrent `plugin` responses.
  - The silent CLI commands in the next bullet use `cli:request` for errors.
- **Result type.** Each result object that a pass saw has a `type` discriminator, for example `{"type":"ok"}`, `{"type":"tab_info"}`, `{"type":"workspace_info"}`, or `{"type":"pane_zoom"}`. Most results also put the payload in a named field, for example `{"pane":{…},"type":"pane_info"}` and `{"agents":[…],"type":"agent_list"}`.
- **Silent success.** Some method-backed CLI commands print nothing on success. They exit with status 0. These commands are `pane send-text`, `pane send-keys`, `pane run`, `pane report-agent`, `pane report-agent-session`, `pane release-agent`, `pane report-metadata`, `workspace report-metadata`, and `server stop`. A pass saw the error ID `cli:request` for the `pane` and `workspace` commands in this list. Compare `agent send-keys`, which prints `{"id":"cli:agent:send-keys","result":{"type":"ok"}}`.
- **Server errors: JSON on stderr, exit 1.** The CLI prints a socket error as the error envelope on stderr, with exit status 1. The `code` is snake_case, and the `message` is text for a person. Some messages do not name the ID, for example `pane not found` from `pane zoom` and `pane split`. For the error codes, see [errors.md](errors.md). These are examples:

  ```text
  {"error":{"code":"pane_not_found","message":"pane w1:p99 not found"},"id":"cli:pane:read"}
  {"error":{"code":"workspace_not_found","message":"workspace w99 not found"},"id":"cli:workspace:get"}
  {"error":{"code":"agent_pane_not_found","message":"agent target pane w1:p99 not found"},"id":"cli:agent:start"}
  {"error":{"code":"pane_not_found","message":"pane w1:p99 not found"},"id":"cli:request"}
  ```

  When no server runs at the socket, the CLI makes the error `server_not_running` itself. This envelope has `id` first:

  ```text
  {"id":"cli:agent:list","error":{"code":"server_not_running","message":"no herdr server is running at <socket>; run `herdr` to start or attach it"}}
  ```

- **Syntax errors: usage on stderr, exit 2.** Bad flags, bad arguments, or an unknown subcommand print a usage string on stderr. The usage string is not JSON, and the exit status is 2. Examples are `unknown option: --path` for `worktree remove --path`, and `invalid read source: bogus` for `pane read --source bogus`. `tab move` and `workspace move` print the usage list of the group. An unknown top-level group prints `unknown command: layout` and `run 'herdr --help' for usage`. `herdr server live-handoff --help` also prints its one-line usage, with exit status 2 and not 0.
- **Output that is not the envelope.** Some CLI commands print text for a person by default, not the JSON envelope. These commands are `status`, `channel show`, `config check`, `config reset-keys`, `completion`, `agent explain`, `api schema`, `integration status`, `plugin list`, `plugin config-dir`, `server agent-manifests`, and `session delete`. `server live-handoff` prints text on stderr. `plugin uninstall` prints its errors as text on stderr. Most of these CLI commands add `--json` for a form that a machine can read, but `integration status` has no `--json`. The sibling commands `plugin action list` and `plugin log list` print the envelope with no flag.
- **Bare JSON.** Some `--json` forms print a bare JSON value with no `id` or `result`. These forms are `status --json`, `session list --json`, `session stop --json`, `session delete --json`, `agent explain --file … --json`, and `machine list --json`. `server agent-manifests --json` and `plugin list --json` print the envelope.
- **Positional or flag targets.** Read-by-ID commands take the ID as a positional argument: `pane get`, `pane read`, `tab get`, and `workspace get`. `pane close/rename/move/send-text/send-keys/wait-output/run/report-*/release-agent` also take a positional `<pane_id>`. Directional commands and split-relative commands use flags: `pane resize --direction … [--pane <id>]`, and `pane neighbor/focus/edges/layout/process-info/current`. These commands accept `--pane <id>` or `--current` for the subject pane. `pane split`, `pane zoom`, and `pane input` accept a positional `[PANE_ID]` and also `--pane`/`--current`. `pane current` takes no positional ID. `pane swap` uses `--direction` or an explicit `--source-pane`/`--target-pane` pair.
- **Focus flags.** The create commands accept `--focus` or `--no-focus`, which exclude each other. These commands are `workspace create`, `tab create`, `worktree create`, `worktree open`, `plugin pane open`, `pane split`, and `pane move`. The flags set the focus after the command.
- **Variadic labels.** `tab rename`, `workspace rename`, and `pane rename` take the new label as trailing variadic arguments, and the CLI joins them with spaces. Only `pane rename` and `agent rename` accept `--clear` to remove the label. `tab rename` and `workspace rename` have no `--clear`, so they use `--clear` as the label text:

  ```text
  $ herdr tab rename w1:t2 --clear
  {"id":"cli:tab:rename","result":{"tab":{…,"label":"--clear",…},"type":"tab_info"}}
  $ herdr workspace rename w1 --clear
  {"id":"cli:workspace:rename","result":{"type":"workspace_info","workspace":{…,"label":"--clear",…}}}
  ```

- **Key names.** `send-keys` uses `esc` as the canonical name of the Escape key. herdr also accepts `escape` and `Escape`. An unknown key gives the error `invalid_key`.
- **Pass-through arguments after `--`.** `agent start` gives the arguments after `--` to the agent executable. `pane report-agent` and `pane report-agent-session` send them as `resume_argv` (new in 0.9.3). The first resume argument must be a plain command name, not a path.

Checked 2026-10-17 against herdr 0.9.5: a pass ran each bullet again on a scratch server. The pass checked the positional bullet and the focus bullet against the `--help` of each subcommand.
