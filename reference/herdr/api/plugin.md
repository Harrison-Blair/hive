# herdr API: plugin methods

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Index: [README.md](../README.md). Wire format: [protocol.md](../protocol.md).

The `plugin.*` methods manage the plugin registry of herdr. They do these operations:

- Link a plugin directory into the registry. The registry is not specific to one session. Refer to the text below.
- Enable or disable an installed plugin.
- List the installed plugins and the actions that they declare.
- Invoke an action.
- Read the command logs of plugins.
- Open, focus, and close the terminal panes that a plugin owns.

A plugin manifest describes each plugin. herdr parses the manifest into an `InstalledPluginInfo`. The manifest declares actions, event hooks, link handlers, panes, build commands, startup commands, and target platforms.

Each method on this page has a `herdr plugin` CLI equivalent. Refer to the **CLI** line of each method. Since 0.9.1, these CLI commands exist. `herdr plugin log` also has the alias `herdr plugin logs`. These CLI commands have no `plugin.*` method on the wire:

- `herdr plugin install` installs a plugin from GitHub.
- `herdr plugin uninstall` uninstalls a plugin.
- `herdr plugin update [-y | --yes] [PLUGIN]...` updates the plugins that `herdr plugin install` installed from GitHub. This command is new in 0.9.5.
- `herdr plugin config-dir <PLUGIN_ID>` prints the config directory of a plugin. It also makes that directory, `<config root>/plugins/config/<PLUGIN_ID>`, if the directory does not exist. It does this for each ID, also for an ID that no plugin has. It does not need a running server.

Checked 2026-10-17 against herdr 0.9.5: the CLI command list, the `logs` alias, and the `config-dir` behavior. The `install`, `uninstall`, and `update` commands were not probed, because they fetch plugins from GitHub.

The registry is global to the herdr config root. It is not specific to one session. herdr keeps the registry in `<config root>/plugins.json`, and a `.plugins.lock` file at the config root locks it. All sessions that use the same config root see the same plugins. The registry stays after a server restart. `plugin.link` also makes the directories `<XDG_CONFIG_HOME>/herdr/plugins/config/<id>/` and `<XDG_STATE_HOME>/herdr/plugins/<id>/`. `plugin.unlink` does not remove them. Thus, to isolate `plugin.link`, `plugin.unlink`, `plugin.enable`, and `plugin.disable`, move the full config root to a different directory, for example with `XDG_CONFIG_HOME`. Also move the state directory with `XDG_STATE_HOME`. A scratch session directory is not sufficient. Refer to [../environment.md](../environment.md) for the herdr directories and environment variables.

Checked 2026-10-17 against herdr 0.9.5: the 11 methods on this page, with `XDG_CONFIG_HOME`, `XDG_STATE_HOME`, `XDG_DATA_HOME`, and `XDG_CACHE_HOME` in a private directory. No probe had a TUI client attached. The examples show real responses from a scratch server, with placeholder paths. These items were not probed:

- `herdr plugin install`, `herdr plugin uninstall`, and `herdr plugin update`. They fetch plugins from GitHub and write a managed checkout.
- The activation of a real link handler. It needs a TUI client with mouse capture and a Ctrl+click.
- `integration.install` and `integration.uninstall`. Refer to [integration.md](integration.md).

The stamp of each method gives its exact coverage.

11 methods (`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`):

| method | purpose |
| --- | --- |
| [plugin.action.invoke](#pluginactioninvoke) | Run a declared plugin action with an optional invocation context. |
| [plugin.action.list](#pluginactionlist) | List the actions that installed plugins declare. |
| [plugin.disable](#plugindisable) | Disable an installed plugin by ID. |
| [plugin.enable](#pluginenable) | Enable an installed plugin by ID. |
| [plugin.link](#pluginlink) | Link a plugin directory into the registry from a filesystem path. |
| [plugin.list](#pluginlist) | List the installed plugins and their manifests. |
| [plugin.log.list](#pluginloglist) | List the recent command logs of plugins. |
| [plugin.pane.close](#pluginpaneclose) | Close a plugin pane by ID. |
| [plugin.pane.focus](#pluginpanefocus) | Focus a plugin pane by ID. |
| [plugin.pane.open](#pluginpaneopen) | Open a new terminal pane that a plugin owns. |
| [plugin.unlink](#pluginunlink) | Remove a plugin from the registry by ID. |

[Shared plugin types](#shared-plugin-types) at the end of this file defines the composite response types one time. These types are `InstalledPluginInfo`, `PluginActionInfo`, `PluginCommandLogInfo`, `PluginPaneInfo`, and their nested manifest and source types. Each method links to them. `PaneInfo` is a domain type for more than one namespace. [../data-model.md](../data-model.md#paneinfo) defines it.

---

## plugin.action.invoke

This method runs a declared plugin action. `action_id` identifies the action. `plugin_id` can limit the search to one plugin. An optional `PluginInvocationContext` describes the workspace, tab, pane, selection, or clicked link that caused the invocation. The command of the action runs as a plugin subprocess. herdr records the run as a `PluginCommandLogInfo`. Refer to [plugin.log.list](#pluginloglist).

The invoke does not block. herdr sends the response when it starts the process. In the response, `log.status` is always `"running"`. The response has no `finished_unix_ms`, `exit_code`, `stdout`, or `stderr`. To find how the command finished, poll [plugin.log.list](#pluginloglist) for the `log_id` in the response. No method waits for the command to complete. If the process cannot start, the response also has `status: "running"`. The log entry then has `status: "failed"` and an `error`, such as `"No such file or directory (os error 2)"`. That entry has empty `stdout` and `stderr` and no `exit_code`.

This method changes state, because it starts a process. Checked 2026-10-17 against herdr 0.9.5.

**Params** (`PluginActionInvokeParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `action_id` | string | yes | — | ID of the action to invoke. It matches `PluginActionInfo.action_id`. A qualified ID, the plugin ID and the action ID joined with `.` (for example `acme.tools.ok`), also finds the action without `plugin_id`. |
| `plugin_id` | string \| null | no | null | Limits the search to the actions of this plugin. When this value is null, herdr searches the actions of all installed plugins. If more than one installed plugin declares the same `action_id`, the call without `plugin_id` fails with `ambiguous_plugin_action`. This count includes the actions of a disabled plugin. Refer to **Errors**. |
| `context` | [PluginInvocationContext](#plugininvocationcontext-request) \| null | no | null | The invocation context for the action, with workspace, tab, pane, selection, or link details. herdr accepts no value, an explicit `null`, `{}`, or a full object. |

**Result** — `type: "plugin_action_invoked"`:

| field | type | meaning |
| --- | --- | --- |
| `type` | const `"plugin_action_invoked"` | Result discriminator. |
| `action` | [PluginActionInfo](#pluginactioninfo) | The action that herdr invoked. |
| `context` | [PluginInvocationContext](#plugininvocationcontext-response) | The context that the action ran with. The server can add values to it. Refer to [PluginInvocationContext (response)](#plugininvocationcontext-response). If the caller gives no `invocation_source`, the server sets it to `"api"`. If the caller gives no `correlation_id`, the server sets it to the `id` of the request envelope. The server keeps a value that the caller gives. |
| `log` | [PluginCommandLogInfo](#plugincommandloginfo) | The log entry for the command. herdr makes it when the process starts, with `status: "running"`. This is never the finished entry. Refer to the text above. |

**Errors**:

- `plugin_not_found`: `plugin_id` is unknown. Message: "plugin not found".
- `plugin_action_not_found`: `action_id` is unknown. Message: "plugin action not found".
- `ambiguous_plugin_action`: `action_id` matches the actions of more than one installed plugin, and the call has no `plugin_id`. Message: `plugin action id matches more than one action; include plugin_id`.
- `plugin_disabled`: the plugin of the action is disabled. Message: "plugin \<id\> is disabled".
- `platform_unsupported`: the action does not support the current platform. Message: "action '\<plugin\>.\<action\>' does not support the current platform (linux)". Refer to [PluginPlatform](#pluginplatform).

Other codes can occur. Refer to [errors.md](../errors.md).

**Events**: this method sends no event to `events.subscribe` subscribers.

**CLI**: `herdr plugin action invoke <ACTION_ID> [--plugin <ID>]`. The CLI sets `invocation_source` to `"cli"` and `correlation_id` to `"cli:plugin"`.

**Example** — Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"1","method":"plugin.action.invoke","params":{"action_id":"format-buffer","plugin_id":"acme.tools","context":{"selected_text":"…"}}}
{"id":"1","result":{"type":"plugin_action_invoked","action":{"plugin_id":"acme.tools","action_id":"format-buffer","title":"Format buffer","contexts":["selection"],"command":["fmt","-w","72"]},"context":{"workspace_id":"w1","workspace_label":"project","workspace_cwd":"/home/user/project","tab_id":"w1:t1","tab_label":"1","focused_pane_id":"w1:p1","focused_pane_cwd":"/home/user/project","focused_pane_status":"unknown","selected_text":"…","invocation_source":"api","correlation_id":"1"},"log":{"log_id":"plugin-log-1","plugin_id":"acme.tools","action_id":"format-buffer","command":["fmt","-w","72"],"status":"running","started_unix_ms":1792263132049}}}
```

The request gives only `selected_text`. The server added the workspace, tab, and pane fields from the active workspace. The response above is the full result that an invoke shows. At this point, `log` never has `finished_unix_ms`, `exit_code`, `stdout`, or `stderr`. This is true also when the command finishes very quickly.

---

## plugin.action.list

This method lists all actions that installed plugins declare. `plugin_id` can limit the list to one plugin. The list does **not** filter on `enabled`. It also includes the actions of a disabled plugin. But an invoke of such an action fails with `plugin_disabled`. Refer to [plugin.action.invoke](#pluginactioninvoke). To know which listed actions you can invoke, compare them with the `enabled` field from `plugin.list`.

This method is read-only. Checked 2026-10-17 against herdr 0.9.5, with an empty registry and with a registry that has a disabled plugin.

**Params** (`PluginActionListParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `plugin_id` | string \| null | no | null | Limits the list to the actions of this plugin. When this value is null, herdr returns the actions of all installed plugins. |

**Result** — `type: "plugin_action_list"`:

| field | type | meaning |
| --- | --- | --- |
| `type` | const `"plugin_action_list"` | Result discriminator. |
| `actions` | array of [PluginActionInfo](#pluginactioninfo) | The declared actions, including the actions of disabled plugins. The array is empty when no plugin or action matches. |

**Errors**: an unknown `plugin_id` gives no error. The response then has an empty `actions` array. A `plugin_id` that is not a string, such as an integer, fails with `invalid_request` ("invalid type: integer `5`, expected a string ..."). Other codes can occur. Refer to [errors.md](../errors.md).

**CLI**: `herdr plugin action list [--plugin <ID>]`. The CLI writes the JSON envelope also without `--json`.

**Example** — Checked 2026-10-17 against herdr 0.9.5, with an empty registry.

```json
{"id":"r4","method":"plugin.action.list","params":{}}
{"id":"r4","result":{"type":"plugin_action_list","actions":[]}}
```

---

## plugin.disable

This method disables an installed plugin by ID. After this call, an invoke of its actions and an open of its panes fail with `plugin_disabled`. Its event hooks do not run. herdr keeps the registry entry. In contrast, [plugin.unlink](#pluginunlink) removes the entry. A disable of a plugin that is already disabled succeeds. An enable of a plugin that is already enabled also succeeds.

This method changes state. Checked 2026-10-17 against herdr 0.9.5.

**Params** (`PluginSetEnabledParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `plugin_id` | string | yes | — | ID of the plugin to disable. |

**Result** — `type: "plugin_disabled"`:

| field | type | meaning |
| --- | --- | --- |
| `type` | const `"plugin_disabled"` | Result discriminator. |
| `plugin` | [InstalledPluginInfo](#installedplugininfo) | The changed plugin entry, with `enabled: false`. |

**Errors**: `plugin_not_found`: `plugin_id` is unknown. Message: "plugin not found". Other codes can occur. Refer to [errors.md](../errors.md).

**Events**: this method sends no event to `events.subscribe` subscribers.

**CLI**: `herdr plugin disable <PLUGIN_ID>`

**Example** — Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"1","method":"plugin.disable","params":{"plugin_id":"acme.tools"}}
{"id":"1","result":{"type":"plugin_disabled","plugin":{"plugin_id":"acme.tools","name":"Acme Tools","version":"1.0.0","min_herdr_version":"0.9.0","manifest_path":"/home/user/project/herdr-plugin.toml","plugin_root":"/home/user/project","enabled":false,"actions":[{"id":"format-buffer","title":"Format buffer","contexts":["selection"],"command":["fmt","-w","72"]}],"panes":[{"id":"dashboard","title":"Dashboard","placement":"overlay","command":["sh","dashboard.sh"]}],"source":{"kind":"local"},"warnings":["manifest does not declare platforms; platform support unknown"]}}}
```

---

## plugin.enable

This method enables an installed plugin by ID. Its actions, event hooks, link handlers, and panes become active again.

**Startup commands do not run on enable.** An enable made no new log entry for the `[[startup]]` commands of the plugin. A server restart with the plugin enabled in the registry made such an entry, with `event: "startup"`. Refer to [plugin.log.list](#pluginloglist). Startup commands run one time for each enabled plugin when the server starts. They do not run on this call. A second server that used the same registry also ran the startup command when it started.

This method changes state. Checked 2026-10-17 against herdr 0.9.5.

**Params** (`PluginSetEnabledParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `plugin_id` | string | yes | — | ID of the plugin to enable. |

**Result** — `type: "plugin_enabled"`:

| field | type | meaning |
| --- | --- | --- |
| `type` | const `"plugin_enabled"` | Result discriminator. |
| `plugin` | [InstalledPluginInfo](#installedplugininfo) | The changed plugin entry, with `enabled: true`. |

**Errors**: `plugin_not_found`: `plugin_id` is unknown. Message: "plugin not found". Other codes can occur. Refer to [errors.md](../errors.md).

**Events**: this method sends no event to `events.subscribe` subscribers.

**CLI**: `herdr plugin enable <PLUGIN_ID>`

**Example** — Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"1","method":"plugin.enable","params":{"plugin_id":"acme.tools"}}
{"id":"1","result":{"type":"plugin_enabled","plugin":{"plugin_id":"acme.tools","name":"Acme Tools","version":"1.0.0","min_herdr_version":"0.9.0","manifest_path":"/home/user/project/herdr-plugin.toml","plugin_root":"/home/user/project","enabled":true,"actions":[{"id":"format-buffer","title":"Format buffer","contexts":["selection"],"command":["fmt","-w","72"]}],"panes":[{"id":"dashboard","title":"Dashboard","placement":"overlay","command":["sh","dashboard.sh"]}],"source":{"kind":"local"},"warnings":["manifest does not declare platforms; platform support unknown"]}}}
```

---

## plugin.link

This method links a plugin directory into the registry from a filesystem `path`. herdr parses the manifest into an `InstalledPluginInfo`. The plugin is enabled by default. The response gives canonical paths, with symbolic links resolved.

An optional `source` records where the plugin comes from: local or GitHub, with the repository, owner, ref, and commit. A GitHub source also **requires** `managed_path`. Refer to [PluginSourceInfo](#pluginsourceinfo).

The manifest must have `min_herdr_version`. Refer to [InstalledPluginInfo](#installedplugininfo). If the value is newer than the running herdr, herdr refuses the link with `plugin_requires_newer_herdr`. It does not record a warning for this.

Some manifest problems do not stop the link. Examples are a manifest without `platforms`, and an event hook with an unknown event name. herdr puts these problems into the `warnings` of the entry, and the link succeeds. An explicit empty `platforms` array fails with `invalid_plugin_platform`.

A link of a `plugin_id` that is already linked replaces the existing entry. herdr gives no error for this. The `source` of the last link stays in the entry. The link does not run the `[[build]]` commands of the plugin.

This method changes the registry. Checked 2026-10-17 against herdr 0.9.5.

**Params** (`PluginLinkParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `path` | string | yes | — | Filesystem path to the plugin directory or to the manifest. |
| `enabled` | boolean | no | `true` | Whether herdr enables the plugin immediately. |
| `source` | [PluginSourceInfo](#pluginsourceinfo) \| null | no | null | Data about where the plugin comes from. When the object is present without `kind`, `kind` is local. herdr refuses a `"github"` `kind` without `managed_path`. |

**Result** — `type: "plugin_linked"`:

| field | type | meaning |
| --- | --- | --- |
| `type` | const `"plugin_linked"` | Result discriminator. |
| `plugin` | [InstalledPluginInfo](#installedplugininfo) | The new plugin entry. If a collection field is empty, the JSON does not include it. herdr does not send `[]`. These collection fields are `actions`, `build`, `startup`, `events`, `link_handlers`, `panes`, and `warnings`. |

**Errors**:

- `plugin_manifest_not_found`: `path` is not correct. The message gives the I/O error, for example "No such file or directory (os error 2)".
- `invalid_plugin_min_herdr_version`: the manifest has no `min_herdr_version` ("plugin min_herdr_version is required"). A value that is not a semantic version gives "plugin min_herdr_version must be a semantic version like 0.9.5".
- `plugin_requires_newer_herdr`: the `min_herdr_version` of the manifest is newer than the running herdr.
- `invalid_plugin_platform`: `platforms` is an explicit empty array.
- `invalid_plugin_source`: the GitHub `source` has a problem. Refer to [PluginSourceInfo](#pluginsourceinfo) for the two messages.

Other codes can occur for other manifest errors. Refer to [errors.md](../errors.md).

**Events**: this method sends no event to `events.subscribe` subscribers.

**CLI**: `herdr plugin link <PATH> [--enabled | --disabled]`. Put the flag after `PATH`. `herdr plugin link --disabled <PATH>` fails with "unknown option: \<PATH\>". herdr resolves a relative `PATH` against the working directory of the caller.

**Example** — Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"1","method":"plugin.link","params":{"path":"/home/user/project/herdr-plugin.toml"}}
{"id":"1","result":{"type":"plugin_linked","plugin":{"plugin_id":"acme.tools","name":"Acme Tools","version":"1.0.0","min_herdr_version":"0.9.0","manifest_path":"/home/user/project/herdr-plugin.toml","plugin_root":"/home/user/project","enabled":true,"actions":[{"id":"format-buffer","title":"Format buffer","contexts":["selection"],"command":["fmt","-w","72"]}],"panes":[{"id":"dashboard","title":"Dashboard","placement":"overlay","command":["sh","dashboard.sh"]}],"source":{"kind":"local"},"warnings":["manifest does not declare platforms; platform support unknown"]}}}
```

This manifest has no `platforms`, so the entry has the standing warning. The manifest declares no `build`, `startup`, `events`, or `link_handlers`. Thus the response does not include these fields, and it does not send them as `[]`. A GitHub source needs `managed_path`:

```json
{"id":"2","method":"plugin.link","params":{"path":"/home/user/project","source":{"kind":"github","owner":"acme","repo":"herdr-acme","requested_ref":"v1.0.0"}}}
{"id":"2","error":{"code":"invalid_plugin_source","message":"GitHub plugin source requires managed_path"}}
```

---

## plugin.list

This method lists the installed plugins and their parsed manifests. `plugin_id` can limit the list to one plugin.

This method is read-only. Checked 2026-10-17 against herdr 0.9.5. The check used an empty registry, which gives an empty `plugins` array. It also used a registry with a linked plugin, a disabled plugin, a plugin with warnings, and a plugin with `platforms = ["macos"]`. A second session on the same config root got the same entries.

**Params** (`PluginListParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `plugin_id` | string \| null | no | null | Limits the list to this plugin. When this value is null, herdr returns all installed plugins. |

**Result** — `type: "plugin_list"`:

| field | type | meaning |
| --- | --- | --- |
| `type` | const `"plugin_list"` | Result discriminator. |
| `plugins` | array of [InstalledPluginInfo](#installedplugininfo) | The installed plugin entries. The array is empty when no plugin is installed. |

**Errors**: a `plugin_id` that is not a string, such as an integer, fails with `invalid_request`. An unknown `plugin_id` gives no error. The response then has an empty `plugins` array. Other codes can occur. Refer to [errors.md](../errors.md).

**CLI**: `herdr plugin list [--plugin <ID>] [--json]`. Without `--json`, the CLI writes text for a person, such as "2 plugins installed:", with `config:` and `warning:` lines. With `--json`, it writes the envelope `{"id":"cli:plugin","result":{"plugins":[...],"type":"plugin_list"}}`.

**Example** — Checked 2026-10-17 against herdr 0.9.5, with an empty registry.

```json
{"id":"r3","method":"plugin.list","params":{}}
{"id":"r3","result":{"type":"plugin_list","plugins":[]}}
```

---

## plugin.log.list

This method lists recent log entries for plugin commands, **oldest first**. `plugin_id` can filter the entries, and `limit` can set their number.

herdr logs these commands:

- Action invocations, when they run.
- Event hook runs, when they run.
- Startup commands, one time for each enabled plugin when the server starts. Refer to [plugin.enable](#pluginenable).

No log entry for a build command occurred after a link, an enable, or a server restart. Refer to the `build` row of [InstalledPluginInfo](#installedplugininfo).

The log store is in memory and holds a maximum of 200 entries. After that, herdr removes the oldest entries. After 309 invocations, the store held `plugin-log-110` to `plugin-log-309`. A call without `limit` returns only the newest 50 entries. On 0.9.1, the store held a maximum of 50 entries. After a server restart, `log_id` starts again at `plugin-log-1`. You cannot get a removed entry or an entry from an earlier server run.

This method is read-only. Checked 2026-10-17 against herdr 0.9.5.

**Params** (`PluginLogListParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `plugin_id` | string \| null | no | null | Limits the list to the logs of this plugin. When this value is null, herdr returns the logs of all plugins. An unknown ID gives an empty `logs` array and no error. |
| `limit` | integer (uint, ≥ 0) \| null | no | null | Selects the last N entries, still oldest first. The maximum is the 200 entries in the store. `limit: 120` returned 120 entries, and `limit: 500` returned 200. When this value is null or not present, herdr returns the last 50. `limit: 0` returns **one** entry, not zero. The count is `max(limit, 1)`. A negative value fails with `invalid_request` ("expected usize"). herdr accepts values up to `u64::MAX`. It refuses a larger value as a float that is not an integer. It also refuses `1.5`. |

**Result** — `type: "plugin_log_list"`:

| field | type | meaning |
| --- | --- | --- |
| `type` | const `"plugin_log_list"` | Result discriminator. |
| `logs` | array of [PluginCommandLogInfo](#plugincommandloginfo) | The log entries for plugin commands, oldest first. |

**Errors**: `invalid_request` for a negative `limit` or a `limit` that is not an integer. Other codes can occur. Refer to [errors.md](../errors.md).

**CLI**: `herdr plugin log list [--plugin <ID>] [--limit <N>]`. `herdr plugin logs list` also works.

**Example** — Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"1","method":"plugin.log.list","params":{"plugin_id":"acme.tools","limit":2}}
{"id":"1","result":{"type":"plugin_log_list","logs":[{"log_id":"plugin-log-1","plugin_id":"acme.tools","action_id":"format-buffer","command":["fmt","-w","72"],"status":"succeeded","started_unix_ms":1792263132049,"finished_unix_ms":1792263132053,"exit_code":0,"stdout":"","stderr":""},{"log_id":"plugin-log-2","plugin_id":"acme.tools","action_id":"format-buffer","command":["fmt","-w","72"],"status":"succeeded","started_unix_ms":1792263132049,"finished_unix_ms":1792263132053,"exit_code":0,"stdout":"","stderr":""}]}}
```

`limit: 2` returns the last two entries, and the older entry comes first. When the command writes no output, `stdout` and `stderr` are empty strings, not `null`. A command that exits with a status other than zero gives `status: "failed"` with its `exit_code`, for example `"exit_code":3,"stdout":"out\n","stderr":"err\n"`. The `error` field is for a process that cannot start. The `event` field gives the event hook name and replaces `action_id`. When these two fields do not apply, the entry does not include them, and herdr does not send `null`. All optional fields on this page follow this rule. Refer to [Shared plugin types](#shared-plugin-types).

---

## plugin.pane.close

This method closes a plugin pane by ID. It is only for panes that this plugin API opened. These panes all fail in the same way:

- A real pane that a plugin does not own.
- An unknown ID.
- A plugin pane that is already closed.
- A plugin pane after a server restart. After the restart, herdr restores the pane with the same `pane_id` and label, but it is a normal pane.

Refer to **Errors**. To close a normal pane, use [pane.close](pane.md#paneclose).

If the plugin pane is the last pane of its workspace, this method closes the workspace. It asks for no confirmation. Refer to [pane.close](pane.md#paneclose) for the TUI confirmation that CLI and API closes do not use.

This method changes state, because it closes the pane. Checked 2026-10-17 against herdr 0.9.5.

**Params** (`PluginPaneCloseParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `pane_id` | string | yes | — | ID of the plugin pane to close. It matches `PluginPaneInfo.pane.pane_id`. |

**Result** — `type: "plugin_pane_closed"`:

| field | type | meaning |
| --- | --- | --- |
| `type` | const `"plugin_pane_closed"` | Result discriminator. |
| `pane_id` | string | ID of the closed pane. |

**Errors**: `plugin_pane_not_found` (message: "plugin pane not found"). This error occurs for each case in the list above. Other codes can occur. Refer to [errors.md](../errors.md).

**Events**: the events depend on the placement of the pane.

| closed pane | events, in order |
| --- | --- |
| A `split` plugin pane, or another plugin pane in a tab | `pane_closed`, `layout_updated` |
| A `tab` plugin pane | `pane_closed` only. herdr also removes the tab, but it sends no `tab_closed` and no `layout_updated`. |
| The last pane of a workspace | `workspace_closed`, `pane_closed`. herdr sends no `tab_closed` and no `layout_updated`. |

**CLI**: `herdr plugin pane close <PANE_ID>`

**Example** — Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"1","method":"plugin.pane.close","params":{"pane_id":"w1:p2"}}
{"id":"1","result":{"type":"plugin_pane_closed","pane_id":"w1:p2"}}
```

---

## plugin.pane.focus

This method focuses a plugin pane by ID and brings it to the front. Like [plugin.pane.close](#pluginpaneclose), it is only for plugin panes. A real pane that a plugin does not own fails in the same way as an unknown ID. A plugin pane after a server restart also fails in this way.

This method changes state, because it changes the focus. Checked 2026-10-17 against herdr 0.9.5, with no TUI client attached.

**Params** (`PluginPaneFocusParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `pane_id` | string | yes | — | ID of the plugin pane to focus. |

**Result** — `type: "plugin_pane_focused"`:

| field | type | meaning |
| --- | --- | --- |
| `type` | const `"plugin_pane_focused"` | Result discriminator. |
| `plugin_pane` | [PluginPaneInfo](#pluginpaneinfo) | The focused plugin pane. |

**Errors**: `plugin_pane_not_found` (message: "plugin pane not found"). This error occurs for an unknown `pane_id` and for a real pane that a plugin does not own. Other codes can occur. Refer to [errors.md](../errors.md).

**Events**: this method sends `workspace_focused`, `tab_focused`, and `pane_focused` to subscribers. A focus also runs each enabled `pane.focused` event hook.

**CLI**: `herdr plugin pane focus <PANE_ID>`

**Example** — Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"1","method":"plugin.pane.focus","params":{"pane_id":"w1:p2"}}
{"id":"1","result":{"type":"plugin_pane_focused","plugin_pane":{"plugin_id":"acme.tools","entrypoint":"dashboard","pane":{"pane_id":"w1:p2","terminal_id":"term_65e0dc75c6c0b2","workspace_id":"w1","tab_id":"w1:t1","focused":true,"cwd":"/home/user/project","foreground_cwd":"/home/user/project","label":"Dashboard","agent_status":"unknown","scroll":{"offset_from_bottom":0,"max_offset_from_bottom":0,"viewport_rows":18},"revision":0}}}}
```

---

## plugin.pane.open

This method opens a new terminal pane that a plugin owns. The pane runs the command of the plugin `entrypoint`. `placement` sets where the pane goes: `overlay`, `popup`, `split`, `tab`, or `zoomed`. The working directory of the pane **defaults to the `plugin_root` of the plugin**. It is not the working directory of the caller or of the server. The `label` of the pane comes from the `title` of the manifest pane. `placement: "tab"` opens the pane in a **new tab**, not in the current tab. The method can focus the new pane.

The method works on a server with no TUI client attached, if the server has a workspace. With no client, `overlay`, `split`, `tab`, and `zoomed` panes opened, and a `popup` open returned `ok`. Only a server with no workspace fails. Refer to **Errors**. This is true since 0.9.3. On 0.9.1, each placement failed when no foreground client was attached.

This method changes state, because it starts a process and makes a pane. Checked 2026-10-17 against herdr 0.9.5. The check could not read the pane body of a popup, because a popup is never in `pane.list`. Thus the check used only the result shape and the errors of a popup.

**Params** (`PluginPaneOpenParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `plugin_id` | string | yes | — | ID of the plugin that owns the pane. |
| `entrypoint` | string | yes | — | The plugin pane entrypoint. It identifies the manifest pane and its command. An unknown entrypoint fails with `plugin_pane_not_found` ("plugin pane entrypoint '\<id\>' not found"). It does not give a validation error. |
| `placement` | [PluginPanePlacement](#pluginpaneplacement) \| null | no | null | Placement of the pane. Enum: `overlay`, `popup`, `split`, `tab`, `zoomed`. When this value is not set, herdr uses the manifest default, `overlay`. An unknown value fails with `invalid_request` ("unknown variant ..."). |
| `direction` | [SplitDirection](#splitdirection) \| null | no | null | Split direction. Enum: `right`, `down`. herdr accepts it only when `placement` is `split` or `zoomed`. For `tab`, it fails with `invalid_params` ("tab plugin panes support workspace_id but not target_pane_id or direction"). For `overlay` and `popup`, it fails with `invalid_params` ("overlay and popup plugin panes target the active pane"). |
| `width` | [PopupSize](#popupsize) \| null | no | null | Outer width, in terminal cells (integer 0–65535) or as a percentage string such as `"80%"`. **herdr accepts it only when `placement` is `popup`.** Each other placement fails with `invalid_params` ("width and height are only supported when placement is popup"). The default placement also fails in this way. `"0%"` fails with `invalid_request` ("percentage must be between 1% and 100%"). `70000` fails with `invalid_request` ("cell count must fit in u16"). |
| `height` | [PopupSize](#popupsize) \| null | no | null | Outer height, in terminal cells or as a percentage string. The popup-only rule of `width` also applies. |
| `cwd` | string \| null | no | null | Working directory for the process of the pane. When this value is not set, herdr uses the `plugin_root` of the plugin. It does not use the working directory of the caller or of the server. |
| `env` | object (string → string) | no | `{}` | Environment variables for the process of the pane. The CLI accepts `--env FOO=bar`. A probe did not read the value inside the pane. |
| `target_pane_id` | string \| null | no | null | The existing pane that the placement uses as its anchor. **The placement limits it.** herdr accepts it only for `split` and `zoomed`. For `overlay` and `popup`, it fails with `invalid_params` ("overlay and popup plugin panes target the active pane"). For `tab`, it fails with the same message as `direction` above. |
| `workspace_id` | string \| null | no | null | The workspace for a new `tab` pane. When this value is null, herdr uses the active workspace. **The placement limits it.** herdr accepts it only for `tab`. An unknown ID fails with `workspace_not_found` ("workspace not found"). For `split` and `zoomed`, it fails with `invalid_params` (`split and zoomed plugin panes target an existing pane; use target_pane_id`). For `overlay` and `popup`, it fails with the "target the active pane" message above. |
| `focus` | boolean | no | `false` | Whether herdr focuses the new pane after it opens. A `zoomed` or `overlay` pane gets the focus also when this value is `false`. Refer to **Events**. |

**Result**: for the placements `overlay`, `split`, `tab`, and `zoomed`, the result is `type: "plugin_pane_opened"` with a `plugin_pane`. For **`popup`, the result is only `{"type":"ok"}`**, with no `plugin_pane`. The popup is never in `pane.list`. Thus a caller has no `pane_id` for `plugin.pane.focus` or `plugin.pane.close`. If a popup shows, a second `popup` open fails with `ui_busy` ("a popup pane is already open"). Only [popup.close](ui.md#popupclose) closes the popup. `plugin.pane.close` does not close it. After `popup.close`, a new popup can open.

| field | type | meaning |
| --- | --- | --- |
| `type` | const `"plugin_pane_opened"` | Result discriminator, only for placements that are not `popup`. A popup returns `"ok"`. Refer to the text above. |
| `plugin_pane` | [PluginPaneInfo](#pluginpaneinfo) | The new plugin pane. A popup result does not include it. |

**Errors**:

- `plugin_not_found`: `plugin_id` is unknown.
- `plugin_disabled`: the plugin is disabled. A disabled plugin with an unknown entrypoint also fails with this code.
- `plugin_pane_not_found`: `entrypoint` is unknown. Message: "plugin pane entrypoint '\<id\>' not found".
- `platform_unsupported`: the pane does not support the current platform. Message: "plugin pane does not support the current platform (linux)". Refer to [PluginPlatform](#pluginplatform).
- `pane_not_found`: `target_pane_id` is unknown, for example "pane w9:p9 not found".
- `workspace_not_found`: `workspace_id` is unknown.
- `invalid_params`: `width`, `height`, `direction`, `target_pane_id`, or `workspace_id` is not correct for the placement. Refer to the params above.
- `ui_busy`: a popup is already open.

On a server with no workspace, each placement fails with one of three codes. On 0.9.1, these codes also occurred when no foreground client was attached.

| placement | code |
| --- | --- |
| `overlay`, `popup`, or no placement | `plugin_pane_open_failed` ("no active workspace") |
| `tab` | `no_active_workspace` |
| `split` or `zoomed` | `no_active_pane` |

Other codes can occur. Refer to [errors.md](../errors.md).

**Events**: the events depend on the placement and on `focus`. A `popup` open sends no event. These sequences occurred with no TUI client attached. The sequences with a client attached were not checked in this pass.

| placement | `focus` | events, in order |
| --- | --- | --- |
| `split` | `false` | `pane_created`, `layout_updated` |
| `split` | `true` | `workspace_focused`, `tab_focused`, `pane_created`, `pane_focused`, `layout_updated` |
| `tab` | `false` | `tab_created`, `pane_created`, `layout_updated` |
| `tab` | `true` | `workspace_focused`, `tab_created`, `tab_focused`, `pane_created`, `pane_focused`, `layout_updated` |
| `zoomed` or `overlay` | `false` or `true` | `workspace_focused`, `tab_focused`, `pane_created`, `pane_focused`, `layout_updated` |
| `popup` | `false` or `true` | none |

**CLI**: `herdr plugin pane open --plugin <ID> --entrypoint <ID> [--placement <overlay|split|tab|zoomed>] [--workspace <ID>] [--target-pane <PANE>] [--direction <right|down>] [--cwd <PATH>] [--env <KEY=VALUE>]... [--focus | --no-focus]`. The `--help` text does not list `popup` for `--placement`. But the CLI accepts `--placement popup` and returns `{"type":"ok"}`. An unknown value fails with "invalid pane placement: \<value\>" and exit status 2.

**Example** — Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"1","method":"plugin.pane.open","params":{"plugin_id":"acme.tools","entrypoint":"dashboard","placement":"split","direction":"down","target_pane_id":"w1:p1","focus":true}}
{"id":"1","result":{"type":"plugin_pane_opened","plugin_pane":{"plugin_id":"acme.tools","entrypoint":"dashboard","pane":{"pane_id":"w1:p2","terminal_id":"term_65e0dc75c6c0b2","workspace_id":"w1","tab_id":"w1:t1","focused":true,"cwd":"/home/user/project","foreground_cwd":"/home/user/project","label":"Dashboard","agent_status":"unknown","scroll":{"offset_from_bottom":0,"max_offset_from_bottom":0,"viewport_rows":18},"revision":0}}}}
```

The request above has no `cwd`. Thus `cwd` and `foreground_cwd` are the `plugin_root` of the plugin, `/home/user/project`. In this capture, the workspace has the same directory. In a different probe, the workspace directory and the `plugin_root` were different, and the pane used the `plugin_root`. A `popup` open gives this response:

```json
{"id":"2","method":"plugin.pane.open","params":{"plugin_id":"acme.tools","entrypoint":"dashboard","placement":"popup","width":"40%","height":12}}
{"id":"2","result":{"type":"ok"}}
```

---

## plugin.unlink

This method removes a plugin from the registry by ID. It is the opposite of [plugin.link](#pluginlink). The `removed` flag of the result tells if an entry was present and herdr removed it. The method does not remove the config and state directories of the plugin.

This method changes state. Checked 2026-10-17 against herdr 0.9.5.

**Params** (`PluginUnlinkParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `plugin_id` | string | yes | — | ID of the plugin to unlink. |

**Result** — `type: "plugin_unlinked"`:

| field | type | meaning |
| --- | --- | --- |
| `type` | const `"plugin_unlinked"` | Result discriminator. |
| `plugin_id` | string | The ID in the request. |
| `removed` | boolean | Whether a matching entry was present and herdr removed it. |

**Errors**: none occurred. A plugin that is not in the registry gives `removed: false`, not an error. Two unlinks of the same `plugin_id` give `true` and then `false`. Other codes can occur. Refer to [errors.md](../errors.md).

**Events**: this method sends no event to `events.subscribe` subscribers.

**CLI**: `herdr plugin unlink <PLUGIN_ID>`

**Example** — Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"1","method":"plugin.unlink","params":{"plugin_id":"acme.tools"}}
{"id":"1","result":{"type":"plugin_unlinked","plugin_id":"acme.tools","removed":true}}
```

---

## Shared plugin types

The methods above use these composite types. [../data-model.md](../data-model.md#paneinfo) defines `PaneInfo`. This section defines the other types, which only the plugin methods use.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the fields, types, required flags, and defaults in the tables below. The text and the row notes give the behavior that a pass checked.

Checked 2026-10-17 against herdr 0.9.5: if an optional field has no value or a null value, the JSON object does not include it. Examples are `description`, `warnings`, and each declaration array below. herdr never sends `null` or `[]` for these fields. Read a missing key as its default. One exception applies: herdr fills in the `placement` of a manifest pane, so a response always includes it.

### InstalledPluginInfo

A linked plugin entry with its parsed manifest.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `plugin_id` | string | yes | — | Stable plugin identifier. |
| `name` | string | yes | — | Plugin name for a person. |
| `version` | string | yes | — | Plugin version string. |
| `manifest_path` | string | yes | — | Path to the plugin manifest file. |
| `plugin_root` | string | yes | — | Root directory of the plugin. |
| `enabled` | boolean | yes | — | Whether the plugin is enabled now. |
| `description` | string \| null | no | null | Plugin description. |
| `min_herdr_version` | string | no | `""` | The minimum herdr version that the plugin needs. The schema makes this field optional, with an empty-string default. But the server **requires it in the manifest**. If the manifest does not have it, the link fails with `invalid_plugin_min_herdr_version` ("plugin min_herdr_version is required"). No probe got the `""` default of the schema. The value must be a semantic version. If it is newer than the running herdr, the link fails with `plugin_requires_newer_herdr`. An example message is `plugin requires Herdr 99.0.0 or newer; current Herdr is 0.9.5`. Checked 2026-10-17 against herdr 0.9.5. |
| `source` | [PluginSourceInfo](#pluginsourceinfo) | no | `{"kind":"local"}` | Data about where the plugin comes from. |
| `platforms` | array of [PluginPlatform](#pluginplatform) \| null | no | null | The platforms that the plugin supports. Null or no value means no limit. But a manifest without `platforms` adds a standing warning to `warnings` (`manifest does not declare platforms; platform support unknown`). This warning occurs on each link and each `plugin.list`, also after a restart. An explicit empty array fails with `invalid_plugin_platform` (`platforms must not be an empty array; omit the field to leave platforms undeclared`). Checked 2026-10-17 against herdr 0.9.5. |
| `actions` | array of [PluginManifestAction](#pluginmanifestaction) | no | `[]` | Declared actions. The manifest default is `[]`. If the list is empty, the response does not include it, and herdr does not send `[]`. |
| `build` | array of [PluginManifestBuild](#pluginmanifestbuild) | no | `[]` | Build commands. The empty-list rule of `actions` also applies. `plugin.link`, `plugin.enable`, and a server start did not run them. Possibly only `herdr plugin install` runs `[[build]]` commands. This pass did not probe `herdr plugin install`. Checked 2026-10-17 against herdr 0.9.5 for link, enable, and server start. |
| `startup` | array of [PluginManifestStartup](#pluginmanifeststartup) | no | `[]` | Startup commands. The empty-list rule of `actions` also applies. They run one time for each enabled plugin **when the server starts**, with `event: "startup"` in the log. They do not run on `plugin.link` or `plugin.enable`. Refer to [plugin.enable](#pluginenable). Checked 2026-10-17 against herdr 0.9.5. |
| `events` | array of [PluginManifestEventHook](#pluginmanifesteventhook) | no | `[]` | Event hooks. The empty-list rule of `actions` also applies. |
| `link_handlers` | array of [PluginManifestLinkHandler](#pluginmanifestlinkhandler) | no | `[]` | Link handlers. The empty-list rule of `actions` also applies. |
| `panes` | array of [PluginManifestPane](#pluginmanifestpane) | no | `[]` | Declared panes. The empty-list rule of `actions` also applies. |
| `warnings` | array of string | no | `[]` | Warnings that do not stop the link. herdr collects them when it links the plugin or loads the registry. Examples are an unknown event name and a manifest without `platforms`. herdr keeps the entry, and `plugin.list` shows the warnings. If there are no warnings, the response does not include this field, and herdr does not send `[]`. Checked 2026-10-17 against herdr 0.9.5. |

### PluginSourceInfo

Data about where a plugin comes from. A request can use it, on `plugin.link`, and a response can use it.

As a request field, it has more rules than the field list below shows. When `kind` is `"github"`, it also **requires** `managed_path`. Without `managed_path`, `plugin.link` fails with `invalid_plugin_source` ("GitHub plugin source requires managed_path").

`managed_path` alone is not sufficient:

- If `managed_path` is an existing directory, the link fails with `invalid_plugin_source` ("GitHub plugin managed_path does not match the plugin id"). This occurred for a directory with the name of the plugin ID, and for the plugin root.
- If `managed_path` does not exist, the link fails with "No such file or directory (os error 2)".

On 0.9.3, the existing-directory case gave "No such file or directory (os error 2)". The server expects a specific layout below the managed path. The probes did not find that layout.

Checked 2026-10-17 against herdr 0.9.5: the case without `managed_path`, the two existing-directory cases, and the missing-path case.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `kind` | [PluginSourceKind](#pluginsourcekind) | no | `"local"` | Source kind. Enum: `local`, `github`. |
| `owner` | string \| null | no | null | Repository owner, for GitHub sources. |
| `repo` | string \| null | no | null | Repository name, for GitHub sources. |
| `subdir` | string \| null | no | null | Subdirectory in the repository. |
| `requested_ref` | string \| null | no | null | Requested Git ref, such as a branch or a tag. |
| `resolved_commit` | string \| null | no | null | Resolved commit SHA. |
| `managed_path` | string \| null | no | null | Managed path on the disk for fetched sources. **Required** when `kind` is `"github"`. |
| `installed_unix_ms` | integer (uint64, ≥ 0) \| null | no | null | Install time in Unix milliseconds. |

### PluginSourceKind

Enum: `local`, `github`.

### PluginPlatform

Enum: `linux`, `macos`, `windows`.

The `platforms` list of the plugin applies to each item. An item-level `platforms` list replaces the top-level list for that item. A plugin with `platforms = ["macos"]` links on Linux with no error and no warning. Then `plugin.action.invoke` and `plugin.pane.open` fail with `platform_unsupported`. An action with an item-level `platforms = ["linux"]` in the same plugin runs on Linux. Checked 2026-10-17 against herdr 0.9.5.

### PluginActionContext

Enum: `global`, `workspace`, `tab`, `pane`, `selection`. Where herdr can show or invoke an action.

### PluginPanePlacement

Enum: `overlay`, `popup`, `split`, `tab`, `zoomed`. Where herdr puts a plugin pane.

### SplitDirection

Enum: `right`, `down`. The split direction when the placement of a plugin pane makes a split.

### PopupSize

A `oneOf` type. The value is an integer or a percentage string:

- An integer is a number of terminal cells, 0–65535, and includes the border.
- A string matches `^(100|[1-9][0-9]?)%$`, such as `"80%"`, and is a percentage of the terminal area.

Checked 2026-10-17 against herdr 0.9.5: herdr refuses `"0%"` and `70000`.

### PluginManifestAction

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `id` | string | yes | — | Action ID. |
| `title` | string | yes | — | Display title. |
| `command` | array of string | yes | — | Command line (argv) to run. |
| `contexts` | array of [PluginActionContext](#pluginactioncontext) | no | `[]` | Contexts that the action applies to. |
| `description` | string \| null | no | null | Action description. |
| `platforms` | array of [PluginPlatform](#pluginplatform) \| null | no | null | Platforms that this action supports. |

### PluginManifestBuild

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `command` | array of string | yes | — | Build command line (argv). |
| `platforms` | array of [PluginPlatform](#pluginplatform) \| null | no | null | Platforms that this build applies to. |

### PluginManifestStartup

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `command` | array of string | yes | — | Startup command line (argv). |
| `platforms` | array of [PluginPlatform](#pluginplatform) \| null | no | null | Platforms that this startup command applies to. |

### PluginManifestEventHook

`on` must use the **dotted event name** that [events.subscribe](../protocol.md#subscription-connections) uses, for example `"pane.focused"`. Refer to [../events.md](../events.md) for the event catalog. There are no `plugin.*` event kinds. The manifest hook names a herdr event. The form with an underscore (`"pane_focused"`) links with a warning ("unknown event 'pane_focused'"). After that, it never runs. [plugin.log.list](#pluginloglist) logs each hook run with `event` and without `action_id`. Hooks do not run while the plugin is disabled. Checked 2026-10-17 against herdr 0.9.5.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `on` | string | yes | — | The event name that makes the hook run. It must be a dotted event name to run. Refer to the text above. |
| `command` | array of string | yes | — | Command line (argv) to run on the event. |
| `platforms` | array of [PluginPlatform](#pluginplatform) \| null | no | null | Platforms that this hook applies to. |

### PluginManifestLinkHandler

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `id` | string | yes | — | Link handler ID. |
| `title` | string | yes | — | Display title. |
| `pattern` | string | yes | — | The URL or link pattern that the handler matches. |
| `action` | string | yes | — | The action ID that herdr invokes when a person activates a matching link. |
| `platforms` | array of [PluginPlatform](#pluginplatform) \| null | no | null | Platforms that this handler applies to. |

### PluginManifestPane

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `id` | string | yes | — | Pane ID. |
| `title` | string | yes | — | Display title. |
| `command` | array of string | yes | — | Command line (argv) to run in the pane. |
| `description` | string \| null | no | null | Pane description. |
| `placement` | [PluginPanePlacement](#pluginpaneplacement) | no | `"overlay"` | Default placement. A response always includes it. When the manifest does not set it, the response gives `"overlay"`. |
| `width` | [PopupSize](#popupsize) \| null | no | null | Default width. |
| `height` | [PopupSize](#popupsize) \| null | no | null | Default height. |
| `platforms` | array of [PluginPlatform](#pluginplatform) \| null | no | null | Platforms that this pane applies to. |

### PluginActionInfo

A resolved action that `plugin.action.list` and `plugin.action.invoke` return.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `plugin_id` | string | yes | — | ID of the plugin that owns the action. |
| `action_id` | string | yes | — | Action ID. |
| `title` | string | yes | — | Display title. |
| `command` | array of string | yes | — | Command line (argv). |
| `contexts` | array of [PluginActionContext](#pluginactioncontext) | no | `[]` | Contexts that the action applies to. |
| `description` | string \| null | no | null | Action description. |
| `platforms` | array of [PluginPlatform](#pluginplatform) \| null | no | null | Platforms that this action supports. This is the effective list. If the action has no list, it gets the top-level `platforms` of the plugin. |

### PluginCommandLogInfo

A record of one run of a plugin command.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `log_id` | string | yes | — | Log entry ID, such as `plugin-log-1`. |
| `plugin_id` | string | yes | — | ID of the plugin that owns the command. |
| `command` | array of string | yes | — | Command line (argv) that herdr ran. |
| `status` | [PluginCommandStatus](#plugincommandstatus) | yes | — | Run status. Enum: `running`, `succeeded`, `failed`. |
| `started_unix_ms` | integer (uint64, ≥ 0) | yes | — | Start time in Unix milliseconds. |
| `action_id` | string \| null | no | null | Action ID, when the command was an action invocation. |
| `event` | string \| null | no | null | Event name, when the command was an event hook run or a startup command. |
| `finished_unix_ms` | integer (uint64, ≥ 0) \| null | no | null | End time in Unix milliseconds. It is null while the command runs. |
| `exit_code` | integer (int32) \| null | no | null | Exit code of the process. It is null while the command runs and when the process cannot start. |
| `stdout` | string \| null | no | null | Captured standard output. |
| `stderr` | string \| null | no | null | Captured standard error. |
| `error` | string \| null | no | null | Error message when the command could not start. |

### PluginCommandStatus

Enum: `running`, `succeeded`, `failed`.

### PluginPaneInfo

A pane that a plugin owns.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `plugin_id` | string | yes | — | ID of the plugin that owns the pane. |
| `entrypoint` | string | yes | — | The pane entrypoint that herdr started. |
| `pane` | [PaneInfo](../data-model.md#paneinfo) | yes | — | The pane itself, a domain type for more than one namespace. |

### PluginInvocationContext (request)

An optional context on `plugin.action.invoke`. Each field is optional and can be null.

| field | type | meaning |
| --- | --- | --- |
| `workspace_id` | string \| null | ID of the workspace that caused the invocation. |
| `workspace_label` | string \| null | Label of that workspace. |
| `workspace_cwd` | string \| null | Working directory of that workspace. |
| `worktree` | [WorkspaceWorktreeInfo](#workspaceworktreeinfo) \| null | Worktree data for the workspace. |
| `tab_id` | string \| null | ID of the tab that caused the invocation. |
| `tab_label` | string \| null | Label of that tab. |
| `focused_pane_id` | string \| null | ID of the focused pane at the invocation. |
| `focused_pane_cwd` | string \| null | Working directory of the focused pane. |
| `focused_pane_agent` | string \| null | The agent of the focused pane. |
| `focused_pane_status` | [AgentStatus](#agentstatus) \| null | Agent status of the focused pane. Enum: `idle`, `working`, `blocked`, `done`, `unknown`. |
| `selected_text` | string \| null | Selected text at the invocation. |
| `clicked_url` | string \| null | The clicked URL that started a link handler. |
| `link_handler_id` | string \| null | ID of the link handler that matched. |
| `invocation_source` | string \| null | Where the invocation came from. |
| `correlation_id` | string \| null | Correlation ID that the caller gives. |

### PluginInvocationContext (response)

`plugin.action.invoke` returns this type. It has the same structure as the request form above. Each field is optional and can be null.

The server can add values for fields that the caller did not give. The server fills `workspace_id`, `workspace_label`, `workspace_cwd`, `tab_id`, `tab_label`, `focused_pane_id`, `focused_pane_cwd`, and `focused_pane_status` from the active workspace. The server keeps an ID that the caller gives, and it does not check that ID. It still fills the other fields from the active workspace. Thus a context can mix values from different workspaces. For example, a caller `"workspace_id":"ws-1"` can occur with `"workspace_label":"home"` from `w1`. For `invocation_source` and `correlation_id`, refer to [plugin.action.invoke](#pluginactioninvoke). Checked 2026-10-17 against herdr 0.9.5.

### WorkspaceWorktreeInfo

Git worktree data for a workspace, in `PluginInvocationContext`. The probes used no worktree workspace. For a plain workspace, the context does not include the `worktree` key.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `repo_key` | string | yes | — | Stable key that identifies the repository. |
| `repo_name` | string | yes | — | Repository name. |
| `repo_root` | string | yes | — | Path of the repository root. |
| `checkout_path` | string | yes | — | Path of the checkout of this worktree. |
| `is_linked_worktree` | boolean | yes | — | Whether this is a linked Git worktree, not the primary checkout. |

### AgentStatus

Enum: `idle`, `working`, `blocked`, `done`, `unknown`. [../data-model.md](../data-model.md#agentstatus) gives the meaning of each value and the readiness limits of `idle` and `done`.
