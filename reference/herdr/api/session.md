# herdr API: session methods

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Index: [README.md](../README.md). Wire format: [protocol.md](../protocol.md).

The `session` namespace has two read-only methods that show the whole session. `ping` is the handshake for liveness and version. It returns the server version, the protocol number, and the capabilities of the server. `session.snapshot` returns the full model of the running session at one point in time. The model has every workspace, tab, pane, pane layout, and detected agent. The two methods do not change state, and they take empty params. Neither method emits events. As with all herdr methods, the server closes the connection after one response (see [protocol.md](../protocol.md)). Checked 2026-10-17 against herdr 0.9.5.

On this page, a field with the type `X | null` quotes the JSON schema. On the wire, herdr never emits `null` for these fields. The server omits an optional field from the JSON object when the field has no value. On herdr 0.9.5, no `null` value occurred in a scratch snapshot or in a live snapshot. The scratch snapshots were empty, populated with agents, populated with 26 workspaces, and restored after a restart with `restore_error`. Read each optional field with a default value. Do not expect a `null`. Checked 2026-10-17 against herdr 0.9.5.

> Note: earlier passes got the same result. About fifty 0.9.1 snapshots had no `null` value. These snapshots came from an empty session, a populated session, and a session with agents in launch. No 0.9.3 snapshot of the 2026-10-06 pass had a `null` value. That pass scanned empty snapshots, populated snapshots with five agents, and snapshots after a restart.

2 methods:

| method | purpose |
| --- | --- |
| [`ping`](#ping) | Handshake for liveness and version. Returns the server version, the protocol, and the capabilities. |
| [`session.snapshot`](#sessionsnapshot) | Returns the full live model of the session: workspaces, tabs, panes, layouts, and agents. |

## ping

`ping` is the handshake for liveness and version. It returns the semantic version of the running server and its protocol number. It can also return a `ServerCapabilities` object that shows the optional features of the server. The server returns this object when it supports the capabilities field. Call `ping` first. Make sure that the socket is a live herdr server. Make sure that its `protocol` agrees with the value that your client expects. Then send other methods. `ping` does not change state, and it emits no events.

**Params**: `PingParams` (an empty object. Send `{}`.)

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| _(none)_ | — | — | — | No parameters. Send `"params": {}`. |

The `params` key is mandatory. A request without it fails with `invalid_request` and the message ``invalid request: missing field `params` at line 1 column 27``. Other parts of the request have fewer rules. The server ignores unknown keys in `params` and unknown keys in the request envelope. It accepts `"params": []` and gives a normal pong. The server rejects `"params": null` and `"params": "x"` with `invalid_request` and the message `invalid type: …, expected struct PingParams`. Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"m2","method":"ping"}
{"id":"m2","error":{"code":"invalid_request","message":"invalid request: missing field `params` at line 1 column 27"}}
```

**Result**: `result.type` = `"pong"`

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `type` | string const `"pong"` | yes | — | Result discriminator. |
| `version` | string | yes | — | Semantic version of the server (for example `"0.9.5"`). |
| `protocol` | integer (uint32) | yes | — | Protocol number of the server (for example `22`). |
| `capabilities` | `ServerCapabilities` \| null | no | `null` | Optional feature flags of the server. See the next table. herdr 0.9.1, 0.9.3, and 0.9.5 always give the object. Thus no pass got the `null` case of the schema. |

The types and the default values in the tables of this section come from the schema. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`. The presence of each field: Checked 2026-10-17 against herdr 0.9.5.

`ServerCapabilities`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `live_handoff` | boolean | yes | — | The server supports the live handoff of an attached client from one process to a different process. |
| `detached_server_daemon` | boolean | no | `false` | The server runs as a detached daemon in the background. |
| `endpoint_protocol_generation` | integer (uint32) \| null | no | — | The stable endpoint generation that the client owns and that this server supports. |
| `health_check` | boolean | no | `false` | Shows if this server supports health probes of the endpoint. |
| `surface_interest` | boolean | no | `false` | Shows if this server supports explicit surface interest from the client shell. |
| `ssh_agent_registration` | boolean | no | `false` | The server supports `server.ssh_agent.register` for one connection on the local JSON API. New in 0.9.3. |

herdr 0.9.5 emits all six fields on each `ping`. herdr 0.9.3 also emitted all six, and 0.9.1 emitted the first five. Thus no pass used a documented default value. The values of `live_handoff`, `endpoint_protocol_generation`, `health_check`, `surface_interest`, and `ssh_agent_registration` were `true` or `1` on each server that a pass probed. Only `detached_server_daemon` had two values on the same binary. A foreground `herdr --session <name> server` gives `false`. A daemonized session gives `true`. A client can use this flag to find if a server is a temporary scratch server or a persistent server. Checked 2026-10-17 against herdr 0.9.5 (a foreground scratch server and a daemonized scratch session).

**Errors**: No error code occurred for `ping`. A malformed envelope fails at the protocol layer (see [protocol.md](../protocol.md)). Other codes are possible. For error codes, see [errors.md](../errors.md).

An `invalid_request` error at the envelope level echoes the `id` of the request when the line is valid JSON with a string `id`. This rule applies to a missing `params`, a `null` `params`, and a `params` of the wrong type. It also applies to a missing `method` and to an unknown method. The position of `id` in the object has no effect. The response has `"id":""` only in three conditions. The line is not valid JSON, or `id` is missing, or `id` is not a string. Checked 2026-10-17 against herdr 0.9.5. On 0.9.1, each error at the envelope level gave `"id":""`.

```text
{"method":"ping","id":"m7"}             -> {"id":"m7","error":{... missing field `params` ...}}
{"id":"m8","params":{}}                 -> {"id":"m8","error":{... missing field `method` at line 1 column 23"}}
{"id":"m9","method":"nope","params":{}} -> {"id":"m9","error":{"code":"invalid_request","message":"invalid request: unknown variant `nope`, expected one of `ping`, ... at line 1 column 39"}}
{"params":{},"method":"ping"}           -> {"id":"","error":{... missing field `id` at line 1 column 29"}}
{"id":5,"method":"ping"}                -> {"id":"","error":{... invalid type: integer `5`, expected a string at line 1 column 7"}}
{"id":null,"method":"ping"}             -> {"id":"","error":{... invalid type: null, expected a string at line 1 column 10"}}
notjson                                 -> {"id":"","error":{... expected ident at line 1 column 2"}}
```

**CLI**: API only. No CLI command sends `ping`. `herdr api ping` shows a usage message and exits with status 2. `herdr api` has only `snapshot` and `schema`. Checked 2026-10-17 against herdr 0.9.5.

```text
$ herdr api ping
herdr api commands:
  herdr api snapshot
  herdr api schema [--json | --output PATH]
$ echo $?
2
```

**Example**: a foreground scratch server. Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"m1","method":"ping","params":{}}
{"id":"m1","result":{"type":"pong","version":"0.9.5","protocol":22,"capabilities":{"live_handoff":true,"detached_server_daemon":false,"endpoint_protocol_generation":1,"surface_interest":true,"health_check":true,"ssh_agent_registration":true}}}
```

A daemonized scratch session gave the same response with `"detached_server_daemon":true`. Checked 2026-10-17 against herdr 0.9.5.

## session.snapshot

`session.snapshot` returns the full model of the running session at one point in time. The model has the focus pointers and the full lists of workspaces, tabs, panes, pane layouts, and detected agents. A client that must list or reconcile the session state reads this method first. `session.snapshot` does not change state, and it emits no events. The returned `snapshot` also has its own `version` and `protocol` fields, with the same values as `ping`. Thus one `session.snapshot` call also checks the version. Checked 2026-10-17 against herdr 0.9.5.

The call is fast, so you can poll it. On herdr 0.9.5, a pass sent ten calls of each method in two rounds on scratch servers:

| scratch session | `session.snapshot` | `ping` |
| --- | --- | --- |
| 26 workspaces, 54 tabs, 55 panes, 3 agents | 0.92–1.22 ms | 0.04–0.10 ms |
| 1 workspace, 4 tabs, 5 panes, 3 agents | 0.13–0.17 ms | 0.03–0.06 ms |

Checked 2026-10-17 against herdr 0.9.5. The 0.9.5 pass did not measure a live session. No method detects changes at a lower cost, because `PaneInfo.revision` does not follow most changes (see the [`revision` note](#paneinfo)).

> Note: earlier passes measured slower values. On herdr 0.9.3, a scratch session with 26 workspaces, 56 tabs, 58 panes, and 5 agents gave 2.1–2.7 ms over ten calls. A live session with 7 workspaces, 16 panes, and 16 agents gave 0.85–3.0 ms. `ping` gave 0.07–0.19 ms. On 0.9.1, a session with 2 workspaces, 3 tabs, 5 panes, and 3 agents gave 0.144–0.200 ms.

**Params**: `EmptyParams` (an empty object. Send `{}`.)

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| _(none)_ | — | — | — | No parameters. Send `"params": {}`. |

The `params` key is mandatory, as for `ping`. The server also accepts `[]` and `{"x":1}`. It rejects `null` with `expected struct EmptyParams`. Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"s2","method":"session.snapshot"}
{"id":"s2","error":{"code":"invalid_request","message":"invalid request: missing field `params` at line 1 column 39"}}
{"id":"s5","method":"session.snapshot","params":null}
{"id":"s5","error":{"code":"invalid_request","message":"invalid request: invalid type: null, expected struct EmptyParams at line 1 column 53"}}
```

**Result**: `result.type` = `"session_snapshot"`

The types in the tables of this section come from the schema. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`. A pass checked the presence of each key and the statements in the meaning column, unless a row or a note tells otherwise. Checked 2026-10-17 against herdr 0.9.5.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `type` | string const `"session_snapshot"` | yes | — | Result discriminator. |
| `snapshot` | `SessionSnapshot` | yes | — | The full session model. See the next section. |

### `SessionSnapshot`

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `version` | string | yes | — | Semantic version of the server. |
| `protocol` | integer (uint32) | yes | — | Protocol number. |
| `workspaces` | `WorkspaceInfo[]` | yes | — | All workspaces in the session. |
| `tabs` | `TabInfo[]` | yes | — | All tabs in all workspaces. |
| `panes` | `PaneInfo[]` | yes | — | All panes in all tabs. |
| `layouts` | `PaneLayoutSnapshot[]` | yes | — | The pane geometry and layout of each tab. |
| `agents` | `AgentInfo[]` | yes | — | Detected agents, which are the panes that hold an agent. The array is empty when no agent occurs. |
| `focused_workspace_id` | string \| null | no | — | ID of the focused workspace. The key occurs only when an item has focus. Otherwise the key is absent. The server never emits it as `null`. |
| `focused_tab_id` | string \| null | no | — | ID of the focused tab. The key occurs only when an item has focus. Otherwise the key is absent. |
| `focused_pane_id` | string \| null | no | — | ID of the focused pane. The key occurs only when an item has focus. Otherwise the key is absent. |

A headless `herdr server` without an attached client starts empty. The `workspaces`, `tabs`, `panes`, `layouts`, and `agents` arrays are all `[]`. The three focus pointers are absent. Only `version` and `protocol` have values. When you close each workspace, the snapshot goes back to exactly this shape. Some clients expect one or more workspaces, or read `snapshot["focused_pane_id"]` without a default. These clients fail on a server that just started. Checked 2026-10-17 against herdr 0.9.5 (on a new server, and after a pass closed all 27 workspaces).

```json
{"id":"s1","result":{"type":"session_snapshot","snapshot":{"version":"0.9.5","protocol":22,"workspaces":[],"tabs":[],"panes":[],"layouts":[],"agents":[]}}}
```

### `WorkspaceInfo`

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `workspace_id` | string | yes | — | Workspace ID (for example `"w1"`). See [addressing.md](../addressing.md). |
| `number` | integer (uint) | yes | — | The position of the workspace, which starts at 1. |
| `label` | string | yes | — | Display label. |
| `focused` | boolean | yes | — | Shows if this workspace has focus. |
| `pane_count` | integer (uint) | yes | — | Number of panes in the workspace. |
| `tab_count` | integer (uint) | yes | — | Number of tabs in the workspace. |
| `active_tab_id` | string | yes | — | ID of the active tab of the workspace. |
| `agent_status` | `AgentStatus` | yes | — | Combined agent status of the workspace. |
| `tokens` | map<string,string> | no | — | Key and value tokens. The map has 32 or fewer entries from all sources. Keys match `^[A-Za-z0-9_-]{1,32}$` (for example `{"branch":"main"}`). The key is absent until a client reports metadata. One `workspace.report_metadata` call can set 16 tokens at maximum. |
| `worktree` | `WorkspaceWorktreeInfo` \| null | no | — | The Git worktree of the workspace. The key is absent until an explicit `worktree.open` or `worktree.create`. A workspace that you create with a `cwd` in a Git repository does not get this key. |

The token limits give these errors. Checked 2026-10-17 against herdr 0.9.5.

```text
17 tokens in one call                 -> {"code":"invalid_metadata_token","message":"a metadata report may update at most 16 tokens"}
source a 16 + source b 16 (1 existing) -> {"code":"metadata_token_limit","message":"workspace metadata may contain at most 32 tokens"}
key "bad key" or a 33-character key    -> {"code":"invalid_metadata_token","message":"invalid metadata token key: bad key"}
```

### `WorkspaceWorktreeInfo`

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `repo_key` | string | yes | — | Stable key for the repository (usually its `.git` path). |
| `repo_name` | string | yes | — | Name of the repository. |
| `repo_root` | string | yes | — | Absolute path to the root of the repository. |
| `checkout_path` | string | yes | — | Absolute path of the checkout of this workspace. |
| `is_linked_worktree` | boolean | yes | — | True when `checkout_path` is a linked Git worktree. False when it is the primary checkout. |

### `TabInfo`

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `tab_id` | string | yes | — | Tab ID (for example `"w1:t1"`). |
| `workspace_id` | string | yes | — | ID of the workspace that holds the tab. |
| `number` | integer (uint) | yes | — | The position of the tab in the workspace, which starts at 1. |
| `label` | string | yes | — | Display label. |
| `focused` | boolean | yes | — | Shows if this tab has focus. Focus applies to the full session. One tab or no tab in the session gives `true`. To find the active tab of one workspace, read the `active_tab_id` of that workspace. |
| `pane_count` | integer (uint) | yes | — | Number of panes in the tab. |
| `agent_status` | `AgentStatus` | yes | — | Combined agent status of the tab. |

### `PaneInfo`

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `pane_id` | string | yes | — | Pane ID (for example `"w1:p1"`). |
| `terminal_id` | string | yes | — | Terminal ID of the pane (for example `"term_65970bc8958f71"`). |
| `workspace_id` | string | yes | — | ID of the workspace that holds the pane. |
| `tab_id` | string | yes | — | ID of the tab that holds the pane. |
| `focused` | boolean | yes | — | Shows if this pane has focus. |
| `agent_status` | `AgentStatus` | yes | — | Agent status of the pane. |
| `revision` | integer (uint64) | yes | — | Monotonic revision counter. It does not count all changes. On 0.9.5, a change of the terminal title or an agent session increments it. See the note after this table. |
| `agent` | string \| null | no | — | Agent kind of the detected agent (for example `"claude"`). The key is absent when it has no value. |
| `display_agent` | string \| null | no | — | Agent label for a person to read. The key is absent when it has no value. |
| `agent_session` | `AgentSessionInfo` \| null | no | — | The agent session that herdr links to the pane. The key is absent when it has no value. |
| `cwd` | string \| null | no | — | Working directory of the pane shell. The key is absent when it has no value. |
| `foreground_cwd` | string \| null | no | — | Working directory of the foreground process. The key is absent when it has no value. |
| `label` | string \| null | no | — | Pane label that a user gives with `pane.rename`. The key is absent when it has no value. |
| `title` | string \| null | no | — | Pane title. The key is absent when it has no value. |
| `terminal_title` | string \| null | no | — | Raw terminal title. The key is absent when it has no value. |
| `terminal_title_stripped` | string \| null | no | — | Terminal title without control characters and decoration. The key is absent when it has no value. |
| `scroll` | `PaneScrollInfo` \| null | no | — | Position of the scrollback and the viewport. It occurred on each live pane that a pass examined. It is absent on a pane with `restore_error`. |
| `state_labels` | map<string,string> | no | — | State labels that the server recognizes. The keys come from an allowlist that the API does not show. The write path rejects other keys with `invalid_state_label` and the message `unknown state label: <key>`. |
| `tokens` | map<string,string> | no | — | Key and value tokens. The map has 32 or fewer entries. Keys match `^[A-Za-z0-9_-]{1,32}$`. The key is absent until a client reports metadata. One report can set 16 tokens at maximum. |
| `restore_error` | string \| null | no | — | New in 0.9.3. The reason why a restarted server could not restore this pane. Otherwise the key is absent. See the note after this table. |

> Note: `revision` does not follow most pane activity. On 0.9.5, a pane without an agent also increments `revision`. A new pane shows `revision` 0 in the `tab.create` result. It shows 1 after about 0.2 s, when its shell starts. Each change of the terminal title (OSC 0) increments it. This includes a shell prompt that writes `user@host:cwd` into the title after `cd`. A change of the working directory alone does not increment it. A pass did not examine this behavior on 0.9.3. Thus this page does not tell that the behavior is new in 0.9.5. Checked 2026-10-17 against herdr 0.9.5.
>
> On 0.9.5, these operations did not change `revision`:
>
> - 500 lines of output
> - a scroll to offset 37 (maximum 462), and a scroll from 100000 that the server clamped to 462
> - `pane.rename`, and a title report with `pane.report_metadata`
> - `pane.zoom` on and off, and `layout.set_split_ratio` to 0.3
>
> During that test, the value stayed at 3 from earlier title changes. On agent panes, a herdr-launched claude pane went from 1 to 2 when it got its `agent_session`. That pane went from 2 to 3 during its first turn, at the time when its terminal title changed. An agent session that a client reported did not increment it. `pane.report_agent_session` with source `herdr:claude` set `agent_session`, and `revision` stayed at 1. Checked 2026-10-17 against herdr 0.9.5.
>
> Do not use `revision` to find if a pane changed. Read the snapshot again.

```text
created w1:p4 revision 0; 0.2 s -> 1
seq 1 500                                         -> 1 (no change)
cd /tmp (prompt writes the title)                 -> 2
printf '\033]0;hello\007' (title set, prompt resets it) -> 4
bash --norc (no title prompt), cd /               -> 4 (cwd change alone: no change)
printf OSC 0 "hello"                              -> 5
seq 1 300, pane.rename, pane.zoom, title report   -> 5 (no change)
```

> Note: earlier versions gave different results. On 0.9.1, `revision` stayed at 1 during 500 lines of terminal output. It stayed at 1 during a scroll to offset 37 and a clamp to 462. It also stayed at 1 during `pane.rename`, `pane.report_metadata` (title), and `pane.report_agent`. The same applied to `pane.report_agent_session` with `agent_session_path`, `pane.zoom` on and off, and a split ratio change. On 0.9.1, the only increments (1 to 2) occurred on panes that got an agent session. On 0.9.3, it stayed at 1 during 500 lines of output and a scroll to offset 37 (maximum 462). It also stayed at 1 during `pane.rename`, a title report, and `pane.zoom` on and off. All 0.9.3 increments occurred on agent panes. Each herdr-launched claude pane went to 2 when it had an agent session. A pane went to 3 when its agent stopped. A pane went to 7 after some turns from prompts. A turn did not always increment it. One completed turn moved a pane from 2 to 3, and the next turn left it at 3.

> Note: `restore_error` occurs after a server restart on a pane that the server could not restore. In a test on 0.9.5, a pass stopped a scratch server and deleted the saved `cwd` of a pane. After the restart, the pane came back with this entry. It had no `foreground_cwd`, no `scroll`, and no terminal titles. Its `revision` was 0. The pass did not test other causes. Checked 2026-10-17 against herdr 0.9.5.

```json
{"pane_id":"wV:p1","terminal_id":"term_65e0dd2f028ae37","workspace_id":"wV","tab_id":"wV:t1","focused":false,"cwd":"/home/user/project/gone","restore_error":"Saved directory is unavailable. Restore the directory and restart this session.","agent_status":"unknown","revision":0}
```

> Note: `state_labels` was absent from each `PaneInfo` and `AgentInfo` that a pass examined, on scratch sessions and on a live session. On 0.9.5, the server rejected each of 16 possible keys with `invalid_state_label`. The keys were `plan`, `mode`, `status`, `model`, `state`, `context`, `cost`, `tokens`, `branch`, `queue`, `permission`, `thinking`, `output_style`, `phase`, `lane`, and `detail`. No key list occurs in `--default-config`, in `--skill`, in the schema, or in the `server.agent_manifests` output. Thus a default configuration accepts none of these keys. The API does not show the allowlist. Checked 2026-10-17 against herdr 0.9.5.

```json
{"code": "invalid_state_label", "message": "unknown state label: plan"}
```

### `PaneScrollInfo`

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `offset_from_bottom` | integer (uint64) | yes | — | Scroll offset from the bottom. The value 0 shows the newest output. |
| `max_offset_from_bottom` | integer (uint64) | yes | — | Maximum scroll offset from the bottom. |
| `viewport_rows` | integer (uint64) | yes | — | Number of rows in the visible viewport. |

After 500 lines of output on a headless tab of 120 by 40 cells, `max_offset_from_bottom` was 462 and `viewport_rows` was 40. Checked 2026-10-17 against herdr 0.9.5.

### `AgentInfo`

The `agents` array has one entry for each pane that holds a detected agent. It omits panes without an agent. For example, a test session with 5 panes had 3 agents. The shape is almost the same as `PaneInfo`. `AgentInfo` adds `name`, `interactive_ready`, `launch_pending`, `screen_detection_skipped`, `state_change_seq`, and `completion_seq` (new in 0.9.3). It has no `label` and no `scroll`. Checked 2026-10-17 against herdr 0.9.5. For the method that reads one agent, see [agent.md](agent.md#agentinfo).

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `terminal_id` | string | yes | — | Terminal ID of the pane. |
| `agent_status` | `AgentStatus` | yes | — | Agent status. |
| `workspace_id` | string | yes | — | ID of the workspace that holds the pane. |
| `tab_id` | string | yes | — | ID of the tab that holds the pane. |
| `pane_id` | string | yes | — | ID of the pane that holds the agent. |
| `focused` | boolean | yes | — | Shows if the pane of the agent has focus. |
| `revision` | integer (uint64) | yes | — | Monotonic revision counter. See the [`revision` note](#paneinfo). |
| `agent` | string \| null | no | — | Agent kind of the detected agent (for example `"claude"`). The key is absent until detection runs. See the note on the launch race. |
| `display_agent` | string \| null | no | — | Agent label for a person to read. The key is absent when it has no value. |
| `agent_session` | `AgentSessionInfo` \| null | no | — | The agent session that herdr links to the pane. The key is absent when it has no value. |
| `cwd` | string \| null | no | — | Working directory of the shell. The key is absent when it has no value. |
| `foreground_cwd` | string \| null | no | — | Working directory of the foreground process. The key is absent when it has no value. |
| `name` | string \| null | no | — | Agent name. Only herdr-launched agents (`agent.start --name`) have it. Agents that herdr knows from `pane.report_agent` do not have it. Thus not all agents have this key. For agent identity across Ctrl+Z, see [agent.md](agent.md#agent-identity-across-ctrlz). |
| `title` | string \| null | no | — | Title of the pane or the agent. The key is absent when it has no value. |
| `terminal_title` | string \| null | no | — | Raw terminal title. The key is absent when it has no value. |
| `terminal_title_stripped` | string \| null | no | — | Terminal title without decoration. The key is absent when it has no value. |
| `interactive_ready` | boolean | no | `false` | True when the agent is ready for interactive input. The server emits the key only when the value is true. Otherwise the key is absent. For the limits of readiness, see [data-model.md](../data-model.md#agentstatus). |
| `launch_pending` | boolean | no | `false` | True while the launch of an agent continues. The server emits the key only when the value is true. The key goes away when the launch is complete. It does not change to `false`. |
| `screen_detection_skipped` | boolean | no | `false` | True when herdr did not use detection from the screen for this pane. The server emits the key only when the value is true. No pass saw it on 0.9.1, 0.9.3, or 0.9.5. |
| `state_change_seq` | integer (uint64) | no | `0` | Sequence number that increments on each change of agent status. The server always emits it. A new agent shows an explicit `0`. |
| `completion_seq` | integer (uint64) \| null | no | — | New in 0.9.3. The schema text is `The current idle transition completed work, independently of who has viewed it.` The key occurs only while the agent is `idle` or `done` after it completed a turn. Its value is the `state_change_seq` of that transition. The key is absent on a new `idle` agent and while the agent is `working`. It stays when `done` changes to `idle` because a person sees the pane. See [agent.md](agent.md#agentinfo). |
| `state_labels` | map<string,string> | no | — | State labels that the server recognizes. The allowed keys are the same as for `PaneInfo.state_labels`. |
| `tokens` | map<string,string> | no | — | Key and value tokens. The map has 32 or fewer entries. Keys match `^[A-Za-z0-9_-]{1,32}$`. |

> Note: the required fields of the schema are `terminal_id`, `agent_status`, `workspace_id`, `tab_id`, `pane_id`, `focused`, and `revision`. They occurred on each agent entry that a pass examined. The schema gives `interactive_ready`, `launch_pending`, and `screen_detection_skipped` as booleans that cannot be `null`. They are not in its `required` set. herdr 0.9.1, 0.9.3, and 0.9.5 emit each of them only when it is true. No pass saw a `false` value of the three. A settled agent has `interactive_ready: true` and no `launch_pending` key. Agents from `pane.report_agent` have none of the three keys. No pass saw `screen_detection_skipped` in about fifty 0.9.1 snapshots or in any 0.9.3 or 0.9.5 snapshot. Read all three keys with a `false` default. Most agents do not have these keys, so a direct read fails. Checked 2026-10-17 against herdr 0.9.5.

> Note: the sequence numbers follow the agent status. In a test, a pass sent one turn to a real claude agent with `agent.prompt`. After `agent.start`, `state_change_seq` was an explicit `0`. It was 7 when claude showed its trust prompt at about 3.4 s. Checked 2026-10-17 against herdr 0.9.5.

```text
0.30s idle    state_change_seq 8  completion_seq absent
0.60s working state_change_seq 9  completion_seq absent
3.62s done    state_change_seq 10 completion_seq 10
pane.focus -> idle, completion_seq 10, state_change_seq 10
```

> Note: a detection race occurs after `agent.start`. On 0.9.5, the call returned in 0.66 ms. The `agents[]` entry occurs immediately with `name` and `launch_pending: true`. It has no `agent` key, and its `agent_status` is `"unknown"`. A pass started claude and read the snapshot every 100 ms. The `agent` key occurred at about 0.1 s. `agent_status` changed to `blocked` at about 3.4 s, when claude showed its first-run trust prompt. `launch_pending` stayed true while claude waited at that prompt. After a pass answered the prompt, `agent_status` changed to `idle` with `interactive_ready` in 0.1 s. `launch_pending` went away at the same time. `agent_session` occurred in 0.4 s. Thus `launch_pending` can stay true much longer than the usual time for a launch. A client that polls `session.snapshot` for a new agent must use `pane_id` or `name` as the key. Do not use `agent` as the key. Checked 2026-10-17 against herdr 0.9.5.
>
> Earlier versions had slower timings. On 0.9.3, the call returned in 0.5–4.9 ms. The `agent` key occurred at about 0.5 s, and `agent_status` changed to `blocked` at about 3.7 s. On 0.9.1, `interactive_ready`, `agent_session`, and `agent_status: "idle"` occurred about two seconds after a pass answered the trust prompt.
>
> The default answer of the claude trust prompt is "No, exit". Thus Enter alone stops claude. The pass sent Down and then Enter.

```text
agent.start 0.66 ms
 0.00s agent absent, unknown, launch_pending true, name reviewer, state_change_seq 0
 0.10s agent "claude", unknown, launch_pending true
 3.42s blocked (first-run trust prompt), launch_pending true, state_change_seq 7
trust prompt answered (Down, Enter):
 0.10s idle, launch_pending gone, interactive_ready true, state_change_seq 8
 0.40s agent_session {"source":"herdr:claude","kind":"id",...}, revision 2
```

This entry is a settled herdr-launched agent before its first turn, on a scratch server. Checked 2026-10-17 against herdr 0.9.5.

```json
{"terminal_id":"term_65e0dccc114ba5","name":"reviewer","agent":"claude","terminal_title":"✳ Claude Code","terminal_title_stripped":"Claude Code","agent_status":"idle","agent_session":{"source":"herdr:claude","agent":"claude","kind":"id","value":"39e104e8-f853-4c6d-9a37-cc385ba68a49"},"workspace_id":"w1","tab_id":"w1:t4","pane_id":"w1:p5","focused":false,"interactive_ready":true,"state_change_seq":8,"cwd":"/home/user/project","foreground_cwd":"/home/user/project","revision":2}
```

After one completed turn, the same entry also has `completion_seq`: `"agent_status":"done","state_change_seq":10,"completion_seq":10`. The `revision` changed to 3, and the terminal title changed to `✳ Single word reply`. Checked 2026-10-17 against herdr 0.9.5.

### `AgentSessionInfo`

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `source` | string | yes | — | Source of the agent session (for example `"herdr:claude"`). |
| `agent` | string | yes | — | Agent kind of the agent session. |
| `kind` | `AgentSessionRefKind` | yes | — | Kind of reference: `id` or `path`. |
| `value` | string | yes | — | The session ID or the path, as `kind` tells. |

`AgentSessionRefKind` enum: `id`, `path`. A snapshot showed only `id`. A herdr-detected claude agent session gave `kind` `"id"`. A client can report an agent session with `pane.report_agent_session` and `agent_session_id`. The snapshot showed this agent session only when its `source` was `herdr:claude`. It then occurred on `PaneInfo` and on `AgentInfo`. With a custom `source`, the snapshot showed no `agent_session_id` and no `agent_session_path`. This applied to `pane.report_agent_session` and to `pane.report_agent`, and the result was `ok`. A report of `agent_session_path` with source `herdr:claude` also did not occur in the snapshot. Thus no pass saw the `path` variant, and it comes only from the schema. Checked 2026-10-17 against herdr 0.9.5.

```json
{"source":"herdr:claude","agent":"claude","kind":"id","value":"0000-test"}
```

### `PaneLayoutSnapshot`

The pane geometry of one tab. The array has one entry for each tab.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `workspace_id` | string | yes | — | ID of the workspace that holds the tab. |
| `tab_id` | string | yes | — | The tab of this layout. |
| `zoomed` | boolean | yes | — | True when a zoomed pane fills the tab. |
| `area` | `PaneLayoutRect` | yes | — | The full layout area of the tab. |
| `focused_pane_id` | string | yes | — | ID of the focused pane in this tab. |
| `panes` | `PaneLayoutPane[]` | yes | — | The panes and their rectangles. |
| `splits` | `PaneLayoutSplit[]` | yes | — | The split nodes that divide the area. The array is empty for a tab with one pane. |

A session with 54 tabs gave 54 layouts. A tab with one pane gave `"splits":[]`. Checked 2026-10-17 against herdr 0.9.5.

> Note: when `zoomed` is true, the pane rectangles do **not** change. Each pane in the tab keeps its `rect` from before the zoom. The `area` also does not change. A client that draws from `layouts[].panes[].rect` draws the tiles from before the zoom. To prevent this, the client must examine `zoomed` and make the `focused_pane_id` pane fill `area`. Also, `focused_pane_id` here applies to one tab. It is the focused pane of that tab. `SessionSnapshot.focused_pane_id` applies to the full session. Checked 2026-10-17 against herdr 0.9.5.

```text
zoomed:false panes w1:p1 {x0,y0,w60,h40}, w1:p3 {x60,y0,w60,h40}, area {0,0,120,40}
zoomed:true  identical rects and area
per tab: w1:t2 -> w1:p2, w1:t3 -> w1:p4, while the session focused_pane_id was w1:p5
```

### `PaneLayoutPane`

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `pane_id` | string | yes | — | Pane ID. |
| `focused` | boolean | yes | — | Shows if this pane has focus. |
| `rect` | `PaneLayoutRect` | yes | — | Rectangle of the pane in the layout area. |

### `PaneLayoutSplit`

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `id` | string | yes | — | ID of the split node. |
| `direction` | `SplitDirection` | yes | — | Direction of the split: `right` or `down`. |
| `ratio` | number (float) | yes | — | Split ratio between the two children. |
| `rect` | `PaneLayoutRect` | yes | — | Rectangle that this split node covers. |

`SplitDirection` enum: `right`, `down`.

A split occurred as `{"id":"split_0_root","direction":"right","ratio":0.5,"rect":{...}}`. `layout.set_split_ratio` changed `ratio` to 0.3. Checked 2026-10-17 against herdr 0.9.5.

### `PaneLayoutRect`

All fields are integers (uint16, range 0–65535). The unit is the terminal cell. The range comes only from the schema. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `x` | integer (uint16) | yes | — | Left column. |
| `y` | integer (uint16) | yes | — | Top row. |
| `width` | integer (uint16) | yes | — | Width in cells. |
| `height` | integer (uint16) | yes | — | Height in cells. |

### `AgentStatus`

`WorkspaceInfo`, `TabInfo`, `PaneInfo`, and `AgentInfo` use this enum: `idle`, `working`, `blocked`, `done`, `unknown`. For the meaning of each value, see [data-model.md](../data-model.md#agentstatus).

Clients cannot write `done`. `pane.report_agent` accepts the enum `PaneAgentState`, which has only four values: `idle`, `working`, `blocked`, and `unknown`. It rejects `"done"` with ``unknown variant `done` ``. Thus an external client can never set `done`. Only herdr detection sets `done`. Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"1","error":{"code":"invalid_request","message":"invalid request: unknown variant `done`, expected one of `idle`, `working`, `blocked`, `unknown` at line 1 column 129"}}
```

On herdr 0.9.5, `idle` and `done` do not guarantee that the agent is ready for input. For this limit, see [data-model.md](../data-model.md#agentstatus). `Documented by herdr --skill (herdr 0.9.5); not live-checked: no unrecognized approval dialog and no agent kind with background tasks was available`.

**Errors**: No error code occurred for `session.snapshot`. Other codes are possible. The only error that occurred was the `invalid_request` error at the envelope level. For error codes, see [errors.md](../errors.md). Checked 2026-10-17 against herdr 0.9.5.

**CLI**: `herdr api snapshot` writes the *full response envelope* as one line of JSON: `{"id":"cli:api:snapshot","result":{"type":"session_snapshot","snapshot":{...}}}`. It does not write only the `snapshot` object. The snapshot is at `.result.snapshot`. Thus `jq '.workspaces'` on the direct output gives `null`. The CLI writes the keys in alphabetical order. It exits with status 0. Checked 2026-10-17 against herdr 0.9.5.

```text
{"id":"cli:api:snapshot","result":{"snapshot":{"agents":[{"agent":"claude","agent_status":"idle","cwd":"/home/user/project","focused":false,"fore…
top keys: id, result
result keys: snapshot, type
snapshot keys: agents, focused_pane_id, focused_tab_id, focused_workspace_id, layouts, panes, protocol, tabs, version, workspaces
```

**Example**: a headless scratch server with one worktree workspace that has a `branch` token, one split, and one more tab. The capture is complete. herdr generated the IDs. The server made 31 workspaces before this capture. Thus the workspace ID is `wW`. The `area` of a headless server is `{0,0,120,40}`. An attached client can change it, for example with a sidebar. The `viewport_rows` of `wW:p2` was 38 in a rectangle of 40 rows, 1.5 s after the split. This page does not explain that value. Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"r2","method":"session.snapshot","params":{}}
{"id":"r2","result":{"type":"session_snapshot","snapshot":{"version":"0.9.5","protocol":22,"focused_workspace_id":"wW","focused_tab_id":"wW:t1","focused_pane_id":"wW:p1","workspaces":[{"workspace_id":"wW","number":1,"label":"api-docs","focused":true,"pane_count":3,"tab_count":2,"active_tab_id":"wW:t1","agent_status":"unknown","tokens":{"branch":"main"},"worktree":{"repo_key":"/home/user/project/.git","repo_name":"project","repo_root":"/home/user/project","checkout_path":"/home/user/project","is_linked_worktree":false}}],"tabs":[{"tab_id":"wW:t1","workspace_id":"wW","number":1,"label":"1","focused":true,"pane_count":2,"agent_status":"unknown"},{"tab_id":"wW:t2","workspace_id":"wW","number":2,"label":"build","focused":false,"pane_count":1,"agent_status":"unknown"}],"panes":[{"pane_id":"wW:p1","terminal_id":"term_65e0dd551016d38","workspace_id":"wW","tab_id":"wW:t1","focused":true,"cwd":"/home/user/project","foreground_cwd":"/home/user/project","terminal_title":"user@host:~/project","terminal_title_stripped":"user@host:~/project","agent_status":"unknown","scroll":{"offset_from_bottom":0,"max_offset_from_bottom":0,"viewport_rows":40},"revision":1},{"pane_id":"wW:p2","terminal_id":"term_65e0dd551085f39","workspace_id":"wW","tab_id":"wW:t1","focused":false,"cwd":"/home/user/project","foreground_cwd":"/home/user/project","terminal_title":"user@host:~/project","terminal_title_stripped":"user@host:~/project","agent_status":"unknown","scroll":{"offset_from_bottom":0,"max_offset_from_bottom":0,"viewport_rows":38},"revision":1},{"pane_id":"wW:p3","terminal_id":"term_65e0dd5510c3b3a","workspace_id":"wW","tab_id":"wW:t2","focused":false,"cwd":"/home/user/project","foreground_cwd":"/home/user/project","terminal_title":"user@host:~/project","terminal_title_stripped":"user@host:~/project","agent_status":"unknown","scroll":{"offset_from_bottom":0,"max_offset_from_bottom":0,"viewport_rows":40},"revision":1}],"layouts":[{"workspace_id":"wW","tab_id":"wW:t1","zoomed":false,"area":{"x":0,"y":0,"width":120,"height":40},"focused_pane_id":"wW:p1","panes":[{"pane_id":"wW:p1","focused":true,"rect":{"x":0,"y":0,"width":60,"height":40}},{"pane_id":"wW:p2","focused":false,"rect":{"x":60,"y":0,"width":60,"height":40}}],"splits":[{"id":"split_0_root","direction":"right","ratio":0.5,"rect":{"x":0,"y":0,"width":120,"height":40}}]},{"workspace_id":"wW","tab_id":"wW:t2","zoomed":false,"area":{"x":0,"y":0,"width":120,"height":40},"focused_pane_id":"wW:p3","panes":[{"pane_id":"wW:p3","focused":true,"rect":{"x":0,"y":0,"width":120,"height":40}}],"splits":[]}],"agents":[]}}}
```
