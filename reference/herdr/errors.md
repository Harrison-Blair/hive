# herdr API: error handling

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Index: [README.md](README.md). Wire format: [protocol.md](protocol.md).

Each herdr request either succeeds with a `result` or fails with one `error` object on the
same connection. This page gives the error envelope and the error output of the CLI. It
also gives each error `code` that a pass got from a herdr 0.9.5 server, unless a row tells
otherwise. Each row gives the trigger and the evidence. This folder does not keep the probe
captures, except the examples on this page. Other pages link to this page for error codes.

## Error envelope

A failed request gives exactly one JSON object with an LF at the end. Then the server closes
the connection. Success responses follow the same rule of one request for each connection.

```json
{"id": "<echoed request id>", "error": {"code": "<snake_case>", "message": "<human text>"}}
```

The `error_response` schema (`ErrorBody`) defines `error.code` and `error.message` as
required strings. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`.

- **`id`** echoes the `id` of the request, also on most `invalid_request` failures. The
  server gives `"id": ""` only when it cannot find one string `id` in the request line.
  This occurs for JSON that it cannot parse. It also occurs for an `id` that is missing, not
  a string, or duplicated. See [protocol.md](protocol.md) and the
  [`invalid_request` variants](#invalid_request-variants) below.
- **`error.code`** is a stable snake_case slug for machine matching. **The schema does not
  list the set of codes.** It gives `code` the type `string` with no fixed values. Thus the
  catalog below comes from probes and [raw/skill.md](raw/skill.md), and it is **not
  complete**. Match on the codes that you handle. Treat an unknown code as a general failure.
- **`error.message`** is text for a person. It can include specific IDs or paths, for
  example `pane w1:p99 not found`. Do not parse it for control flow. Use `code`.
- The server enforces the rule of one request for each connection: it closes the connection
  after the response. On 0.9.5, a second request on the same connection fails at once with a
  broken pipe (`EPIPE`). This occurs when the client writes it immediately or 200 ms after
  the response. A read then gives end-of-file. If the client writes two request lines in one
  write, the server answers the first line and then resets the connection (`ECONNRESET`).
  See [protocol.md](protocol.md). A 2026-09-19 pass against 0.9.1 recorded a reset
  (`ECONNRESET`) for the sequence with the second write. In all these cases the server does
  not answer the second request. Thus treat `EPIPE`, `ECONNRESET`, and end-of-file as the
  same result.

```text
request: workspace.get for w99
response: {"id":"a","error":{"code":"workspace_not_found","message":"workspace w99 not found"}}
second write after 0 ms:   EPIPE [Errno 32] Broken pipe, then a read gives b''
second write after 200 ms: EPIPE [Errno 32] Broken pipe, then a read gives b''
two ping lines in one write: 1 response line, then ECONNRESET
```

Checked 2026-10-17 against herdr 0.9.5.

## CLI surfacing

The CLI is a thin client of the socket. Thus the CLI prints a server error as the **same
JSON object on stderr**, and the process exits with status **1**:

```text
$ herdr pane read w1:p99
{"error":{"code":"pane_not_found","message":"pane w1:p99 not found"},"id":"cli:pane:read"}
```

The CLI usually sets `id` to a `cli:<group>:<command>` string. Some pane commands use
`cli:request` in its place. On 0.9.5 these commands are `pane report-agent` and
`pane report-agent-session`, with or without `--` resume arguments. The other commands are
`pane send-text`, `pane send-keys`, and `pane run`.

A CLI **syntax** error is an unknown flag, a bad argument, or a missing required option. The
CLI finds it before it sends a request to the socket. The output is *not* JSON, and the
process exits with status **2**. This check does not need a running server. Example:

```text
$ herdr worktree remove --path x
unknown option: --path
```

A client-side connection failure also exits with status **1** and prints JSON on stderr. If
no server runs for the session, the code is `server_not_running`. In this object, the `id`
key comes before the `error` key:

```text
$ herdr pane read w1:p99
{"id":"cli:pane:read","error":{"code":"server_not_running","message":"no herdr server is running at /home/user/.config/herdr/sessions/demo/herdr.sock; run `herdr` to start or attach it"}}
```

Thus a caller can tell the cases apart:

- Exit status 2 is a malformed CLI command.
- Exit status 1 is an error with a JSON `code`. The code is one of the codes below, or the
  client-side code `server_not_running`.

For the other CLI conventions, see [cli-mapping.md](cli-mapping.md).

Checked 2026-10-17 against herdr 0.9.5. The pass ran both exit paths and got the examples
above byte for byte.

## Observed error codes

| code | trigger | evidence |
| --- | --- | --- |
| `workspace_not_found` | A method gave a workspace ID that does not exist (`workspace w99 not found`). | Checked 2026-10-17 against herdr 0.9.5 |
| `pane_not_found` | A method gave a pane ID that does not exist (`pane w1:p99 not found`). | Checked 2026-10-17 against herdr 0.9.5 |
| `agent_pane_not_found` | An `agent.*` method targeted a pane that does not exist (`agent target pane w1:p99 not found`). The server gives this code on the agent command path when it cannot find the target pane of the agent. It is different from `pane_not_found`. | Checked 2026-10-17 against herdr 0.9.5 |
| `agent_pane_busy` | `agent.start` targeted a pane that is not an available shell, for example a pane with a foreground program (`agent target pane w1:p2 is not an available shell`). See [api/agent.md](api/agent.md#agentstart). | Checked 2026-10-17 against herdr 0.9.5 |
| `invalid_agent_argument` | `agent.start` got `args` that the server cannot encode safely for the shell of the target pane. An example is an argument with a control character (`agent arguments cannot be encoded safely for the target shell`). | Checked 2026-10-17 against herdr 0.9.5 |
| `invalid_request` | The request envelope or `params` failed to deserialize. This code covers more than one trigger. See the [variants](#invalid_request-variants) below. | Checked 2026-10-17 against herdr 0.9.5 |
| `popup_not_open` | `popup.close` or another popup method ran when no popup is open (`no popup is open`). | Checked 2026-10-17 against herdr 0.9.5 |
| `split_not_found` | `layout.set_split_ratio` got a split `path` that does not point to a split that exists (`split path not found`). The `path` is an array of booleans. An array of numbers fails as `invalid_request`. | Checked 2026-10-17 against herdr 0.9.5 |
| `unsupported_event_wait_match` | `events.wait` got a `match_event` that is not a match on the agent status of a pane. Examples are `tab_created`, `pane_created`, `workspace_focused`, and `pane_agent_detected`. The message is `events.wait currently supports pane agent status matches`. On 0.9.5, as on 0.8.2, 0.9.1, and 0.9.3, the server supports only these agent status matches. The schema allows more match shapes. A `pane_agent_status_changed` match also needs an `agent_status` field. Without it, the request fails as `invalid_request` before the server checks the match shape. | Checked 2026-10-17 against herdr 0.9.5 |
| `stale_content` | A method that takes a `content_revision` got a revision that is not the current content revision of the pane. The field is optional on `pane.selection.read`, `pane.copy_motion`, and `pane.link.activate`. `pane.copy_search` requires the field. The message is `pane content changed`. On `pane.link.activate` the message is `pane content changed before link activation`. | Checked 2026-10-17 against herdr 0.9.5 |
| `connection_local_only` | New in 0.9.1. A plain API connection called `client_shell.surface.set`. Only a connection through a client shell endpoint can use this method (`client_shell.surface.set is only available through a client shell endpoint`). | Checked 2026-10-17 against herdr 0.9.5 |
| `command_not_found` | New in 0.9.1. `command.invoke` got a `command_id` that the command manifest of the client shell does not know. The message is always `custom command manifest is stale; reload configuration`. A server that just started, with no loaded manifest, also gives this message. Thus the "stale" and "reload" text is a general fallback. It is not evidence that the manifest changed. | Checked 2026-10-17 against herdr 0.9.5 |
| `stale_announcement` | New in 0.9.1. `product_announcement.dismiss` got an `id` and `version` pair that is not the current announcement (`the product announcement is no longer current`). | Checked 2026-10-17 against herdr 0.9.5 |
| `stale_release_notes` | New in 0.9.1. `release_notes.dismiss` got a `version` that is not the current version of the release notes (`the release notes are no longer current`). | Checked 2026-10-17 against herdr 0.9.5 |
| `tab_not_found` | A method gave a tab ID that does not exist (`tab w1:t99 not found`). | Checked 2026-10-17 against herdr 0.9.5 |
| `unsupported_agent_kind` | `agent.start` got a `kind` that the server does not know as an interactive agent kind (`unsupported interactive agent kind bogus`). | Checked 2026-10-17 against herdr 0.9.5 |
| `invalid_agent_name` | `agent.start` got a `name` that does not match the required pattern. The name must start with a lowercase letter. It must contain only lowercase letters, digits, `-`, or `_` (1–32 characters). The server gives this code for `Build` and for a name with 33 characters. | Checked 2026-10-17 against herdr 0.9.5 |
| `invalid_agent_timeout` | The `timeout_ms` of `agent.start` was out of the accepted range. It must be greater than 3000 ms and at most 300000 ms. The server gives this code for 3000 and for 300001. | Checked 2026-10-17 against herdr 0.9.5 |
| `agent_not_found` | An `agent.*` method, such as `agent.get`, `agent.prompt`, or `agent.wait`, targeted a `name` with no matching live agent (`agent target nosuchagent not found`). On `agent.get`, a pane ID target with no pane gives the same code (`agent target w1:p99 not found`). `agent.prompt` with that pane target and `agent.rename` with an unknown name also give it. The server gives `agent_pane_not_found` when it cannot find the *pane*. It gives this code when it cannot find the *name*. | Checked 2026-10-17 against herdr 0.9.5 |
| `empty_agent_prompt` | `agent.prompt` got an empty `text` (`agent prompt must not be empty`). The server checks the text before it resolves the target. Thus an unknown target with an empty text also gives this code. | Checked 2026-10-17 against herdr 0.9.5 |
| `agent_name_taken` | New in 0.9.1. `agent.start` got a `name` that another agent uses while its startup is still `launch_pending`. The message lists the `terminal_id`, `pane_id`, `workspace_id`, `tab_id`, `cwd`, and `status` of that agent as candidates. See the [capture](#agent-name-capture) below. | Checked 2026-10-17 against herdr 0.9.5 |
| `agent_launch_pending` | New in 0.9.1. `agent.rename` targeted an agent whose `agent.start` launch is still pending (`agent name cannot change while startup is pending`). The name cannot change until the startup is complete. This code is a third guard on the agent lifecycle, with `agent_not_ready` and `agent_blocked`. | Checked 2026-10-17 against herdr 0.9.5 |
| `layout_not_found` | `layout.set_split_ratio` got both a `tab_id` and a `pane_id` as the target (`layout target not found`). The two fields are mutually exclusive. On 0.9.5, a `tab_id` that does not exist also gives this code. No record tells if 0.9.3 did the same. This code is different from `split_not_found`, which is for a `path` that the server cannot resolve. | Checked 2026-10-17 against herdr 0.9.5 |
| `query_too_large` | The `query` of `pane.copy_search` was longer than the limit of the server (`copy search query is too large`). A query of 4096 characters succeeds. A query of 4097 characters fails. The server checks the size before the `content_revision`. Thus a too-large query with a stale revision also gives this code. | Checked 2026-10-17 against herdr 0.9.5 |
| `invalid_metadata_ttl` | The `ttl_ms` of `pane.report_metadata` was out of range. It must be at least 1 and at most 86400000. The messages are `metadata ttl_ms must be at least 1` and `metadata ttl_ms must be 86400000 or less`. | Checked 2026-10-17 against herdr 0.9.5 |
| `invalid_metadata_request` | `pane.report_metadata` set no field to change or to clear (`missing metadata field to set or clear`). Or it set and cleared the same field (`cannot set and clear the same metadata field`). | Checked 2026-10-17 against herdr 0.9.5 |
| `invalid_metadata_source` | `pane.report_metadata` got an empty `source` (`metadata source must not be empty`). See [api/pane.md](api/pane.md#panereport_metadata). | Checked 2026-10-17 against herdr 0.9.5 |
| `invalid_params` | `notification.show` got an empty `title` (`notification title is empty`). | Checked 2026-10-17 against herdr 0.9.5 |
| `not_linked_worktree` | `worktree.remove` targeted a workspace that herdr does not track as a linked worktree checkout (`workspace is not a Herdr-managed worktree checkout`). The server gives this code also for a workspace whose `cwd` is not in a Git work tree. For that case, `worktree.list` and `worktree.create` give `not_git_worktree`, but `worktree.remove` does not. | Checked 2026-10-17 against herdr 0.9.5 |
| `not_git_worktree` | `worktree.list` or `worktree.create` targeted a workspace whose `cwd` is not in a Git work tree (`Herdr worktree actions require a workspace inside a Git work tree`). Other read and create paths probably give this code too. | Checked 2026-10-17 against herdr 0.9.5 |
| `plugin_not_found` | A `plugin.*` method, such as `plugin.pane.open`, gave a `plugin_id` with no linked or known plugin (`plugin not found`). | Checked 2026-10-17 against herdr 0.9.5 |
| `invalid_pane_swap` | `pane.swap` got neither a `direction` nor a pair of `source_pane_id` and `target_pane_id` (`provide either direction with optional pane_id, or source_pane_id and target_pane_id`). | Checked 2026-10-17 against herdr 0.9.5 |
| `workspace_move_block_failed` | `workspace.move_block` got an empty `workspace_ids` list (`workspace_ids must not be empty`). | Checked 2026-10-17 against herdr 0.9.5 |
| `timeout` | New in 0.9.1. This is a general code for a method that takes a `timeout_ms` and finds no match before the time ends. This pass got it from `events.wait` (`timed out waiting for event match`) and from `pane.wait_for_output` (`timed out waiting for output match`). It also got it from `agent.prompt` with a wait (`timed out waiting for agent status`). See also `timeout` under [Skill-documented codes](#skill-documented-codes). | Checked 2026-10-17 against herdr 0.9.5 |
| `unknown_method` | New in 0.9.3. The request gave a method that this build knows as removed. These methods are `pane.graphics.set`, `pane.graphics.clear`, `pane.graphics.info`, and `pane.graphics.stream` (`unknown method: pane.graphics.set`). Other unknown names, such as `bogus.method`, `pane.foo`, and `pane.graphics.x`, still fail as `invalid_request`. See the variants below. | Checked 2026-10-17 against herdr 0.9.5 |
| `invalid_resume_argv` | New in 0.9.3. `pane.report_agent` or `pane.report_agent_session` got a bad `resume_argv`. The value can be empty. It can start with something that is not a plain command name, such as a path, `~`, a leading `-`, a space, or `=`. It can have more than 64 elements or more than 8192 bytes in total. It can contain a control character. The server checks it after `pane_not_found` and `invalid_agent`. See [api/pane.md](api/pane.md#panereport_agent). | Checked 2026-10-17 against herdr 0.9.5 |
| `resume_not_accepted` | New in 0.9.3. A source sent a valid `resume_argv`, but after that call the source does not hold the pane for that agent. Examples are `pane.report_agent_session` before any `pane.report_agent`, and a call after a release. Other examples are a call from a superseded source and a call from an ignored `herdr:<agent>` source. The message is `resume_argv requires the reporter to hold the pane; report its state with pane.report_agent first`. | Checked 2026-10-17 against herdr 0.9.5 |
| `invalid_ssh_agent` | New in 0.9.3. `server.ssh_agent.register` got a `socket_path` that is not an absolute path to an existing socket with the required owner (`SSH agent must be an absolute, user-owned socket`). See [api/server.md](api/server.md#serverssh_agentregister). | Checked 2026-10-17 against herdr 0.9.5 |
| `invalid_agent` | `pane.report_agent` or `pane.report_agent_session` got an empty `agent` (`agent label must not be empty`). | Checked 2026-10-17 against herdr 0.9.5 |
| `invalid_key` | `pane.send_keys` or `pane.send_input` got a key name that the server does not support (`unsupported key bogus`). The server checks the keys before it writes text. Thus a bad key causes no input at all. One character, such as `X`, is a valid key, and the server types it. | Checked 2026-10-17 against herdr 0.9.5 |
| `target_pane_not_found` | `pane.move` into a tab gave a `target_pane_id` that does not exist (`target pane <ID> not found`). Or the pane is not in the destination tab (`target pane <ID> is not in tab <TAB_ID>`). | Checked 2026-10-17 against herdr 0.9.5 |
| `stale_target` | `pane.link.resolve` or `pane.link.activate` targeted a pane outside the focused tab (`pane is no longer visible`). On `pane.link.resolve`, a pane in the focused tab does not cause this code, also when it is not the focused pane. | Checked 2026-10-17 against herdr 0.9.5 |
| `stale_pane_target` | `pane.edit_scrollback` targeted a pane that is not focused (`pane is no longer focused`). | Checked 2026-10-17 against herdr 0.9.5 |
| `selection_unavailable` | `pane.selection.read` got an anchor or a cursor outside the content of the pane (`selection text is unavailable`). | Checked 2026-10-17 against herdr 0.9.5 |
| `invalid_regex` | `pane.wait_for_output` got a `regex` that the server cannot parse. The message has the full diagnostic of the Rust regex parser. For `(` it ends with `error: unclosed group`. | Checked 2026-10-17 against herdr 0.9.5 |
| `invalid_metadata_token` | `pane.report_metadata` got a token key that does not match `^[A-Za-z0-9_-]{1,32}$` (`invalid metadata token key: bad key!`). Or it got more than 16 tokens in one call (`a metadata report may update at most 16 tokens`). | Checked 2026-10-17 against herdr 0.9.5 |
| `invalid_state_label` | `pane.report_metadata` got a `state_labels` key that is not a known agent status (`unknown state label: bogus`). The server accepts `done`. | Checked 2026-10-17 against herdr 0.9.5 |
| `metadata_token_limit` | After `pane.report_metadata`, the pane would have more than 32 tokens in total (`pane metadata may contain at most 32 tokens`). This pass got it with 16 tokens from each of two sources and 1 more token from a third source. | Checked 2026-10-17 against herdr 0.9.5 |

<a id="agent-name-capture"></a>
Capture of `agent_name_taken`, `agent_launch_pending`, and `agent_not_ready`. The first
`agent.start` for `build` in pane `w1:p1` gave `agent_started` at once with
`launch_pending: true`. A real `claude` agent stopped at its folder-trust dialog in that
pane. Then these requests gave these responses:

```text
{"id":"1","method":"agent.start","params":{"name":"build","kind":"claude","pane_id":"w1:p3","timeout_ms":5000}}
{"id":"1","error":{"code":"agent_name_taken","message":"agent name build is already used; candidates: terminal_id=term_65e0db45540291 pane_id=w1:p1 workspace_id=w1 tab_id=w1:t1 cwd=/home/user/project status=Unknown"}}
{"id":"1","method":"agent.rename","params":{"target":"build","name":"reviewer"}}
{"id":"1","error":{"code":"agent_launch_pending","message":"agent name cannot change while startup is pending"}}
{"id":"1","method":"agent.prompt","params":{"target":"build","text":"hi"}}
{"id":"1","error":{"code":"agent_not_ready","message":"agent build is not an active named agent"}}
```

The message gives `status=Unknown` with a capital letter. Shortly after, `agent.get` gave
`agent_status: "blocked"` for the same agent.

Note: `pane.read` requires a `source` field (`visible`, `recent`, `recent_unwrapped`, or
`detection`). Without it, the request fails as `invalid_request` (`missing field \`source\``).
The server gives this error before it checks if the pane exists. Thus a bad `pane_id` gives
`pane_not_found` only when the request has a `source`.

Checked 2026-10-17 against herdr 0.9.5 on a scratch server for each row and the note, with
the messages in the rows.

### Removed with `pane.graphics.*` in 0.9.3

herdr 0.9.3 removed the `pane.graphics.*` methods (see [api/pane.md](api/pane.md)), and
these three codes went with them. A pass found no method on 0.9.3 or 0.9.5 that gives them.
A search of the 0.9.3 and 0.9.5 binaries found none of the three code strings. The same
search found no `unknown_method` string in the 0.9.5 binary, but the server gives that
code. Thus a string search does not prove that a code is absent. This page keeps the three
codes as history for 0.9.1 servers only.

| code | trigger (0.9.1) | evidence |
| --- | --- | --- |
| `feature_disabled` | A method needs an experimental or optional feature that is off. In 0.8.2, `pane.graphics.info` needed `experimental.kitty_graphics`. In 0.9.1, `kitty_graphics` is `true` by default, and it moved from `[experimental]` to `[terminal]`. Thus that trigger does not occur on a default config. A 0.9.1 probe set `terminal.kitty_graphics` to `false` on a scratch server. Then `pane.graphics.info`, `.set`, and `.clear` all gave this code (`pane graphics are disabled by terminal.kitty_graphics`). | History: probes on 0.8.2, and on 0.9.1 with a non-default config. The methods do not exist on 0.9.5. |
| `cell_size_unavailable` | New in 0.9.1. `pane.graphics.info` ran on a headless server, or with an outer terminal that did not send its cell size. A default config can cause it, because `kitty_graphics` is on by default. | History: probe on 0.9.1. The methods do not exist on 0.9.5. |
| `invalid_image` | `pane.graphics.set` got image dimensions that are not both greater than zero. | History: probe on 0.9.1. The methods do not exist on 0.9.5. |

A 2026-09-19 pass against herdr 0.9.1 recorded the triggers in this table. Checked
2026-10-17 against herdr 0.9.5: each `pane.graphics.*` name gives `unknown_method`. The
method list in the unknown-variant message has no `pane.graphics.*` name, and the search of the
binary found none of the three code strings.

### `invalid_request` variants

`invalid_request` is one code for all failures to deserialize the envelope or `params`. The
`message` tells the cause. Since 0.9.3 the server echoes the `id` of the request when the
request line is valid JSON with exactly one string `id`. For JSON that it cannot parse, and
for a missing, non-string, or duplicate `id`, the server gives `"id": ""`.
[protocol.md](protocol.md) also gives these cases. A 2026-09-19 pass recorded `"id": ""`
for each variant in its list on 0.9.1.

The column in each message is a position in the request line as sent. Thus the column
changes with the request text. These requests give the columns in the table:

```text
{"id":"7","method":"bogus.method","params":{}}
-> {"id":"7","error":{"code":"invalid_request","message":"invalid request: unknown variant `bogus.method`, expected one of `ping`, `server.stop`, ..."}}
{"id":"7","method":"workspace.get","params":{}}
-> {"id":"7","error":{"code":"invalid_request","message":"invalid request: missing field `workspace_id` at line 1 column 47"}}
{"id":"7","method":"workspace.get"}
-> {"id":"7","error":{"code":"invalid_request","message":"invalid request: missing field `params` at line 1 column 35"}}
{"method":"ping","params":{}}
-> {"id":"","error":{"code":"invalid_request","message":"invalid request: missing field `id` at line 1 column 29"}}
{1:2}
-> {"id":"","error":{"code":"invalid_request","message":"invalid request: key must be a string at line 1 column 2"}}
{"id":"7","method":"layout.set_split_ratio","params":{"tab_id":"w1:t1","path":[],"ratio":1e400}}
-> {"id":"7","error":{"code":"invalid_request","message":"invalid request: number out of range at line 1 column 94"}}
{"id":"7","method":"workspace.get","params":{"workspace_id":5}}
-> {"id":"7","error":{"code":"invalid_request","message":"invalid request: invalid type: integer `5`, expected a string at line 1 column 63"}}
{"id":7,"method":"ping","params":{}}
-> {"id":"","error":{"code":"invalid_request","message":"invalid request: invalid type: integer `7`, expected a string at line 1 column 7"}}
{"id":"7","id":"8","method":"ping","params":{}}
-> {"id":"","error":{"code":"invalid_request","message":"invalid request: duplicate field `id` at line 1 column 14"}}
```

| variant | message shape | evidence |
| --- | --- | --- |
| unknown method | `invalid request: unknown variant \`bogus.method\`, expected one of \`ping\`, \`server.stop\`, …`. The message lists each valid method name. On 0.9.5 the list has the 102 schema methods and no `pane.graphics.*` name. The server echoes the `id`. The removed `pane.graphics.*` names give `unknown_method`. | Checked 2026-10-17 against herdr 0.9.5 |
| missing required param field | `invalid request: missing field \`workspace_id\` at line 1 column 47`. The server echoes the `id`. | Checked 2026-10-17 against herdr 0.9.5 |
| missing `params` | `invalid request: missing field \`params\` at line 1 column 35`. The server echoes the `id`. | Checked 2026-10-17 against herdr 0.9.5 |
| missing envelope field (`id`) | `invalid request: missing field \`id\` at line 1 column 29`. The server gives `"id": ""`. | Checked 2026-10-17 against herdr 0.9.5 |
| unparseable JSON | `invalid request: key must be a string at line 1 column 2`, or another serde parse message. The server gives `"id": ""`. | Checked 2026-10-17 against herdr 0.9.5 |
| non-finite number | `invalid request: number out of range at line 1 column 94`. An example is `ratio: 1e400` in `layout.set_split_ratio`, which overflows to infinity in JSON. The server rejects it at parse time, before any range check on the value. The server echoes the `id`. | Checked 2026-10-17 against herdr 0.9.5 |

A field of the wrong type also gives `invalid_request` with a serde type-mismatch message.
Examples are a string where the schema expects an integer, or an integer where it expects a
string. For `"workspace_id":5` the message is `invalid request: invalid type: integer \`5\`,
expected a string at line 1 column 63`, and the server echoes the `id`. Treat each
`invalid_request` as a client defect to fix. Do not retry it as a runtime condition.

Checked 2026-10-17 against herdr 0.9.5.

## Skill-documented codes

[raw/skill.md](raw/skill.md) documents these error codes. On 0.9.5 a pass got each code from
a server, except two cases. These cases are `agent_not_ready` from `agent.start` and
`workspace_group_close_required`. Each row gives its own evidence.

| code | trigger | evidence |
| --- | --- | --- |
| `agent_not_ready` | Per the skill text, `agent.start` gives this code at once when the agent is blocked during startup. The name stays available for `agent read` and `agent send-keys`. Wait until the agent is idle before you prompt it. | `Documented by herdr --skill (herdr 0.9.5); not live-checked`: on 0.9.1 and 0.9.5 no trial got this code from `agent.start`. On 0.9.5, a real `claude` agent stopped at its folder-trust dialog. `agent.start` then gave `agent_started` with `launch_pending: true` at once. |
| `agent_not_ready` (from `agent.prompt`) | `agent.prompt` also gives this code in two cases. In the first case, the `agent.start` launch of the target is still pending (`launch_pending: true`). The message is `agent <name> is not an active named agent`. In the second case, the agent is not the foreground process of the pane. The message is `agent <target> is no longer the pane foreground process`. | Checked 2026-10-17 against herdr 0.9.5. See the [capture](#agent-name-capture) for the first case. |
| `agent_not_ready` (foreground cause) | The second message occurs when the server knows the agent only from `pane.report_agent` and the shell is in the foreground. It also occurs after Ctrl+Z stops a detected agent. herdr keeps the identity of that agent, but the agent is not in the foreground. See [api/agent.md](api/agent.md#agentprompt). | Checked 2026-10-17 against herdr 0.9.5 |
| `agent_blocked` | `agent.prompt` rejects an agent that waits at an approval or question dialog, before herdr sends any input (`agent w1:p2 is blocked and requires interactive input`). Inspect the blocked UI and ask the user before you answer it. The server does this check before the foreground check. | Checked 2026-10-17 against herdr 0.9.5. A `pane.read` after the call gave no prompt text. |
| `agent_prompt_stalled` | A prompt from a state that is not `working` gave no observed lifecycle change in five seconds. Thus herdr gives this code and does not wait without a limit. The message is `agent prompt produced no observed working or blocked state within 5000 ms; current status is idle`. | Checked 2026-10-17 against herdr 0.9.5. `agent.prompt` with `wait` gave it after 5.3 s for an idle agent that ignored input. |
| `timeout` | New in 0.9.1. Per the skill text, `agent.prompt --wait` with a caller `--timeout` gives this code in one case. This case is a caller timeout that ends before the five-second activity check of `agent_prompt_stalled`. The message is `timed out waiting for agent status`. See also `timeout` in the [catalog](#observed-error-codes). | Checked 2026-10-17 against herdr 0.9.5. `agent.prompt` with `wait.timeout_ms` 1000 gave it after 1.0 s. |
| `workspace_group_close_required` | A close of a workspace that has linked worktree workspaces did not set `close_group` (`--group` in the CLI). `workspace close --group` closes the primary workspace and its linked worktree workspaces. Do not add `--group` only to avoid this code. See [api/workspace.md](api/workspace.md#workspaceclose). | `Documented by herdr --skill (herdr 0.9.5); not live-checked`: the trigger needs a linked worktree workspace, and this pass did not make one. |

## Notes for implementers

- The schema gives `error.code` as a free string. The lists on this page come from probes and
  the skill text, and they are **not complete**. Other methods can give codes that this page
  does not list. Handle an unknown code as a general failure. Treat `unknown_method` like
  `invalid_request`: the server does not support the method that you called.
- Match on `code`. Do not match on `message`. One code can have more than one message, for
  example `agent_not_ready` and `stale_content`. The column in an `invalid_request` message
  changes with the request text.
- Use exit status 2 only for CLI syntax errors. Each server error exits with status 1 and
  prints a JSON body on stderr. The client-side `server_not_running` error also exits with
  status 1 and prints JSON.
- After each response, success or error, the server closes the connection. Open a new
  connection for each request. Two types of connection stay open: an `events.subscribe`
  connection and a [`server.ssh_agent.register`](api/server.md#serverssh_agentregister)
  connection.

Checked 2026-10-17 against herdr 0.9.5.

On macOS, `pane run` and `agent start` reject a command that is longer than the canonical
line-input limit of the terminal. The rejection occurs before herdr queues the command. The
error code of this rejection is not known. See [cli-mapping.md](cli-mapping.md) and
[api/agent.md](api/agent.md#agentstart). `Documented by herdr --skill (herdr 0.9.5); not live-checked`:
the test host runs Linux, and the skill text does not name the error code.
