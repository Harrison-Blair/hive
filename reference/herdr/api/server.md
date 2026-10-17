# herdr API: server methods

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Index: [README.md](../README.md). Wire format: [protocol.md](../protocol.md).

The `server` namespace controls the running headless herdr server. It does not control the workspaces, panes, or agents that the server holds. These methods do these operations:

- They stop the server.
- They apply `config.toml` again without a server restart.
- They show and load again the agent manifests, which tell herdr how to detect agent processes.
- They transfer the session to a replacement binary and keep the session state.
- They register a remote-host SSH agent socket for the life of one API connection. This method is new in 0.9.3.

CLI access:

- The `herdr server` CLI command group gives access to five of the six methods.
- `herdr server --help` lists five subcommands: `stop`, `reload-config`, `agent-manifests`, `update-agent-manifests`, and `reload-agent-manifests`. Four of them send a method. `update-agent-manifests` has no method of its own.
- The `herdr server live-handoff` subcommand sends `server.live_handoff`. Only the plain usage list of the group shows it. For example, `herdr server bogus` prints that list.
- `herdr update --handoff` and `--handoff` on remote or app attach also use `server.live_handoff`. Refer to [server.live_handoff](#serverlive_handoff).
- `server.ssh_agent.register` has no CLI subcommand.

Request rules:

- Every method needs a `params` field, also when it has no fields.
- Each request uses its own socket connection.
- When `params` is missing, `null`, or a scalar such as `5`, the request fails with `invalid_request`.
- herdr ignores unknown keys in `params`. herdr also accepts a bare `[]` as `params`.
- When the line is valid JSON and has a string `id`, the `invalid_request` error contains that request `id`. The response `id` is `""` only in two cases. The line is not valid JSON, or its `id` is missing or is not a string. On herdr 0.9.1, every error of this type gave the response `id` `""`. Refer to [protocol.md](../protocol.md).

`server.reload_config`, `server.reload_agent_manifests`, and `server.agent_manifests` send no subscription event. This is true also when the first two methods change the state of the server.

Checked 2026-10-17 against herdr 0.9.5: the subcommand lists, the `params` matrix, and the response `id` cases. The `params` matrix used `params` absent, `null`, `[]`, `{"zz":1}`, and `5` on `server.agent_manifests`, `server.reload_agent_manifests`, and `server.reload_config`. For the event claim, a subscription to all 24 subscription types that have no parameters got no event from the three methods. One `server.reload_config` changed `[ui] sidebar_width` and `[theme] name`. One reload made a local override active. On the same subscription, `workspace.create` gave `workspace_created` and other events.

CLI behavior of the `herdr server` group:

- Each subcommand except `stop` and `live-handoff` first sends a `ping` request on one connection. Then it sends the real request on a second connection.
- `herdr server update-agent-manifests` has no method of its own. First, it does a network fetch in the client. This fetch writes `$XDG_STATE_HOME/herdr/agent-detection/status.toml` again. The default path is `~/.local/state/herdr/agent-detection/status.toml`. Then it sends [server.reload_agent_manifests](#serverreload_agent_manifests) and then [server.agent_manifests](#serveragent_manifests). Each of the two requests has its own `ping` request before it.
- `herdr server update-agent-manifests --json` prints the `server.agent_manifests` envelope, with the keys in alphabetical order. Without `--json`, it prints the same table as `agent-manifests`.
- `reload-agent-manifests` and `reload-config` print the full JSON response envelope on stdout. Neither subcommand has a `--json` flag. `agent-manifests --json` also prints the envelope, not only the result.

Errors of the `herdr server` group do not have one format:

- `stop` prints a plain sentence on failure.
- `reload-config`, `agent-manifests`, `reload-agent-manifests`, and `live-handoff` print a JSON `server_not_running` error envelope.

Checked 2026-10-17 against herdr 0.9.5: the items in the two lists above. A logging proxy in front of a scratch server recorded the wire traffic. `update-agent-manifests` ran with `XDG_STATE_HOME` in a private directory.

A socket path that is too long for `sun_path` causes a different error. All subcommands except `stop` then print a raw Rust debug string, not a sentence or an envelope: `Error: Custom { kind: InvalidInput, error: "local socket name length exceeds capacity of sun_path of sockaddr_un" }`. A pass saw this on herdr 0.9.3 on 2026-10-06. The 2026-10-17 pass did not check it on herdr 0.9.5. The pass tools do not accept a socket path longer than 107 bytes.

6 methods:

| method | purpose |
| --- | --- |
| [server.agent_manifests](#serveragent_manifests) | List the active agent manifests and their remote update status. |
| [server.live_handoff](#serverlive_handoff) | Transfer the running server to a replacement binary, and keep the session state. |
| [server.reload_agent_manifests](#serverreload_agent_manifests) | Load the local agent manifest overrides again from the disk. |
| [server.reload_config](#serverreload_config) | Read `config.toml` again and apply it in the running server. |
| [server.ssh_agent.register](#serverssh_agentregister) | Register a remote-host SSH agent socket while this API connection stays open. |
| [server.stop](#serverstop) | Stop the running server through the socket API. |

Checked 2026-10-17 against herdr 0.9.5: an `unknown variant` error from a scratch server lists exactly these six methods in the `server` namespace.

## server.agent_manifests

This method returns the agent manifests that the server uses now. It returns one entry for each known agent kind. It also returns the time and the result of the last remote update check. This method is read-only. It shows the cached state and does not fetch from the network. Each manifest entry gives these items:

- the source of the manifest: the herdr binary, a remote fetch, or a local override file
- the active version and the cached remote version
- a warning that herdr gave when it selected between the bundled, remote, and local override manifests

A server that you start does an automatic check at startup. Before this check completes, the result has no `last_check_unix` key and no `last_result` key. In that period, each manifest is `"bundled"`. When the network is available, this period is very short, usually less than a few seconds. A client that reads `server.agent_manifests` immediately after it starts a server can get the result from before the check.

A server that cannot get to its catalog URL keeps all manifests `"bundled"`. It did not try again within 31 s. Only the top-level `last_result` shows the failure. Each agent entry still has no `remote_update_error` key.

Checked 2026-10-17 against herdr 0.9.5:

- A new server with no network and an empty state directory gave no `last_check_unix` key and no `last_result` key on the first read. Six seconds later, it gave `"last_result":"failed: failed to fetch https://herdr.dev/agent-detection/index.toml"`.
- A new server with network access gave `last_result: "checked"` and 22 `"remote"` manifests on the first read.
- On the server with no network, two reads 30 s apart gave the same `last_check_unix`. All sources stayed `bundled`, except local overrides.

**Params**: `EmptyParams`, which is `{}`. No fields.

**Result**, with `type: "agent_manifest_status"`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `type` | string const `"agent_manifest_status"` | yes | — | Result discriminator. |
| `manifests` | array of `AgentManifestInfo` | yes | — | One entry for each known agent kind. Refer to the field table below. Each server gave 22 entries, one for each bundled agent kind. |
| `last_check_unix` | integer (uint64) \| null | no | null | Unix time in seconds of the last remote update check. The key is absent until the first check completes. |
| `last_result` | string \| null | no | null | Result label of the last remote check. The value is `"checked"` on success. On failure, it is free text in the form `"failed: <reason>"`, for example `"failed: failed to fetch https://herdr.dev/agent-detection/index.toml"`. The key is absent until the first check completes. |

These are the 22 agent kinds:

```text
pi, claude, codex, gemini, cursor, devin, agy, cline, opencode, copilot, kimi,
kiro, droid, amp, grok, hermes, kilo, qodercli, qwen, letta, maki, muse
```

`AgentManifestInfo` object fields:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `agent` | string | yes | — | Agent kind, for example `pi`, `claude`, or `codex`. |
| `source` | string | yes | — | Source of the active manifest. The value is `"bundled"`, or `"remote:<path>"` for a fetched TOML file. For a local override file, the value is a bare absolute path with no prefix, for example `/home/user/.config/herdr/agent-detection/claude.toml`. Refer to [server.reload_agent_manifests](#serverreload_agent_manifests) for the location of overrides. |
| `source_kind` | string | yes | — | Category of the source: `"bundled"`, `"remote"`, or `"local override"`, with a space. |
| `local_override_shadowing_remote` | boolean | yes | — | True when a local override file exists for this agent kind. It does not tell if the override is active. For a malformed override, this flag is `true`, `source` and `source_kind` stay on the remote or bundled manifest, and `warning` gives the cause. To find the active manifest, read `source` and `source_kind`, not this flag. |
| `active_version` | string \| null | no | null | Version of the active manifest. Each entry had this key, and its value was never null. |
| `cached_remote_version` | string \| null | no | null | Version of the last remote manifest in the cache on the disk. The key is absent, not null, when no usable cached remote manifest exists. An example is a server with all manifests `bundled`. |
| `remote_last_checked_unix` | integer (uint64) \| null | no | null | Unix time in seconds of the last check of the remote manifest of this agent kind. The key is absent, not null, until a remote check completes for this agent kind. |
| `remote_update_result` | string \| null | no | null | Result of the remote check for this agent kind. The value is `"current"` when the manifest had no change, or `"updated"` when herdr fetched a new manifest. The key is absent, not null, until a remote check succeeds for this agent kind. |
| `remote_update_error` | string \| null | no | null | Error text from the last remote update, or null on success. No probe caused this key. When the catalog index is not available, the full fetch fails before any fetch for one agent kind. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`. |
| `warning` | string \| null | no | null | A warning from the selection of this manifest. The key is absent, not null, when no warning exists. Three forms occurred. The first form is for a cached remote version older than the bundled version. The second form is for a remote manifest that herdr could not parse. The third form is for an override file that herdr could not parse. |

A local override that is active has this form. This capture comes after a write of a local override file for `claude` and a call to `server.reload_agent_manifests`:

```json
{"agent":"claude","source":"/home/user/.config/herdr/agent-detection/claude.toml","source_kind":"local override","active_version":"9999.01.01.1","cached_remote_version":"2026.10.11.1","local_override_shadowing_remote":true,"remote_update_result":"current","remote_last_checked_unix":1792263075}
```

A cached remote manifest that herdr cannot parse gives this form:

```json
{"agent":"amp","source":"bundled","source_kind":"bundled","active_version":"2026.07.09.1","local_override_shadowing_remote":false,"remote_update_result":"current","remote_last_checked_unix":1792263075,"warning":"ignored remote manifest /home/user/.local/state/herdr/agent-detection/remote/amp.toml because it could not be loaded: TOML parse error at line 1, column 9\n  |\n1 | garbage [\n  |         ^\nexpected `.`, `=`\n"}
```

**Errors**: none occurred. Other codes are possible.

**CLI**: `herdr server agent-manifests [--json]`. By default, it prints a table for persons. With `--json`, it prints the full response envelope (`{"id":...,"result":{...}}`), not only the result. The keys are in alphabetical order. The table adds a `!` marker and the line `local override shadows cached remote rules` for each entry that has `local_override_shadowing_remote` true. A malformed override that is not active also gets the `!` marker.

```text
last check: 1792262824
result: checked

  pi        remote         active 2026.10.01.1   remote 2026.10.01.1   current
! claude    local override active 9999.01.01.1   remote 2026.10.11.1   current
  local override shadows cached remote rules
! codex     remote         active 2026.10.11.1   remote 2026.10.11.1   current
  local override shadows cached remote rules
```

**Example**

```json
{"id":"1","method":"server.agent_manifests","params":{}}
{"id":"1","result":{"type":"agent_manifest_status","last_check_unix":1792262884,"last_result":"checked","manifests":[{"agent":"pi","source":"remote:/home/user/.local/state/herdr/agent-detection/remote/pi.toml","source_kind":"remote","active_version":"2026.10.01.1","cached_remote_version":"2026.10.01.1","local_override_shadowing_remote":false,"remote_update_result":"current","remote_last_checked_unix":1792262884},{"agent":"grok","source":"bundled","source_kind":"bundled","active_version":"2026.10.05.1","cached_remote_version":"2000.01.01.1","local_override_shadowing_remote":false,"remote_update_result":"current","remote_last_checked_unix":1792263075,"warning":"ignored remote manifest /home/user/.local/state/herdr/agent-detection/remote/grok.toml because cached version 2000.01.01.1 is older than bundled 2026.10.05.1"}, …]}}
```

The `grok` entry comes from a second read of the same server, after a change of its cached remote manifest to an older version. Thus its timestamps are different from the `pi` entry. You can cause this `bundled` form with a `warning` again. Change a cached remote manifest to an older version than the bundled one. On a server with normal updates, each agent kind usually shows `"remote"` and no warning.

Checked 2026-10-17 against herdr 0.9.5, with `XDG_CONFIG_HOME` and `XDG_STATE_HOME` in a private directory: all claims in this section, except `remote_update_error`. These claims include the local override, the malformed override, the three `warning` forms, the `!` marker, and the examples. The examples show real responses from a scratch server, with placeholder paths.

## server.live_handoff

This method transfers the running server to a replacement binary, and it keeps the session state. The current process starts another herdr executable and gives it the live sessions. Attached clients and agents continue after the change. `herdr update --handoff` and `--handoff` on remote attach use this method. The optional guard fields let the caller give the expected protocol or version of the new binary. If the target does not match, herdr refuses the handoff. The caller can also give the executable for the handoff.

A call with empty params (`{}`) does a real handoff with the default or updated binary. It is not a dry run.

After a successful handoff, these changes occur:

- The argv of the process changes from `herdr --session <name> server` to `<abs path>/herdr server --handoff-import <session-dir>/herdr-handoff-<oldpid>.sock <oldpid>-<nonce>`. The default binary is the installed one, for example `/home/user/.local/bin/herdr`.
- The new process has a different parent from the caller. Its `ppid` is 1, or the subreaper of the user, such as `systemd --user`. Thus, a tool that finds servers with `pgrep -f "herdr --session <name>"` does not find them.
- The session state does not change. This state includes workspaces, tabs, panes, focus, and counts. But herdr gives each pane a new `terminal_id`.
- herdr closes each open connection without an event and without a close reason. This includes an `events.subscribe` stream. The subscriber reads EOF. A client must connect again. If the client had a subscription, it must subscribe again.
- `ping` gives `detached_server_daemon: true`. Before the handoff, the server gave `false`. Refer to [protocol.md](../protocol.md#pingpong-handshake).

An `import_exe` that starts but does not use the handoff protocol, such as `/bin/true`, holds the request until a server timeout. After 30 s, the request fails with `handoff_failed` and `timed out waiting for handoff import connection`. During this wait, only `ping` answers. Other requests wait until the timeout occurs, and then they get a response. After the timeout, the original server continues to operate normally.

On herdr 0.9.3, `workspace.list` and `server.reload_config` also answered during the wait. On herdr 0.9.5, `workspace.list`, `pane.list`, `tab.list`, `server.reload_config`, `server.agent_manifests`, and `session.snapshot` all waited.

**Params**, with type `ServerLiveHandoffParams`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `expected_protocol` | integer (uint32) \| null | no | null | If you set it, the target binary must use this protocol version. On a mismatch, herdr refuses the handoff with `handoff_failed`. A negative integer, an integer that is out of range, or a value that is not an integer fails with `invalid_request`. |
| `expected_version` | string \| null | no | null | If you set it, the target binary must give this herdr version. On a mismatch, herdr refuses the handoff with `handoff_failed`. A value that is not a string fails with `invalid_request`. |
| `import_exe` | string \| null | no | null | Path to the replacement executable. When the value is null or absent, herdr uses the default or updated binary. A missing path or a path that is not executable fails with `handoff_failed`, and the message gives the path and the OS error. A value that is not a string fails with `invalid_request`. For a process that does not use the handoff protocol, the request waits 30 s and then fails with `handoff_failed`. Refer to the text above. |

**Result**, with `type: "ok"`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `type` | string const `"ok"` | yes | — | Success acknowledgement. Then the process transfers the session to the replacement binary, and it does not continue to serve the current connection. Refer to the text above. |

**Errors**:

- `handoff_failed` is the only error code that occurred.
- A guard mismatch for the protocol or the version gives `{"code":"handoff_failed","message":"handoff stream closed while reading line"}`. The message does not identify the mismatch.
- A bad `import_exe` gives `handoff_failed` with a message that contains the path and the OS error. An example is `failed to spawn handoff import server at /nonexistent/x-herdr: No such file or directory (os error 2)`. For a path that is not executable, the OS error is `Permission denied (os error 13)`.
- An `import_exe` that does not connect back fails after 30 s with `timed out waiting for handoff import connection`.
- Malformed `params`, such as wrong types or integers out of range, fail with `invalid_request`. herdr does not start a handoff for them.
- Other codes are possible.

**CLI**: `herdr server live-handoff [--import-exe <path>] [--expected-protocol <n>] [--expected-version <version>]`.

- `herdr server --help` does not list this subcommand. The plain usage list of the group shows it as "hand off live panes to a new local server".
- It sends `server.live_handoff` directly with the id `cli:server:live-handoff`. It sends no `ping` request first. It sends only the flags that you give. With no flags, it sends `"params":{}`.
- `--help` and unknown flags print the usage line and exit with status 2.
- On success, it prints `live handoff complete; server log: <path>` and exits with status 0. But `<path>` is always the log of the default server, `$XDG_CONFIG_HOME/herdr/herdr-server.log`. The default path is `~/.config/herdr/herdr-server.log`. This is true also when `HERDR_SOCKET_PATH` points to a `--session` server. That server writes the handoff to the log in its own session directory.
- On an error, it prints the JSON error envelope, with the keys in alphabetical order, and exits with status 1. With no server, it prints the `server_not_running` envelope.

Checked 2026-10-17 against herdr 0.9.5: the CLI items above. A fake socket that answered locally recorded the wire requests and the error output. One CLI handoff ran on a scratch server.

`herdr update --handoff` and `--handoff` on remote or app attach also send this method. `herdr update --help` documents `--handoff` as "Try live handoff after installing". The pass did not run these entry points. `herdr update` installs a new binary into `~/.local/bin`, which is not in a scratch session.

**Example**: matching guards.

```json
{"id":"1","method":"server.live_handoff","params":{"expected_protocol":22,"expected_version":"0.9.5","import_exe":null}}
{"id":"1","result":{"type":"ok"}}
```

A guard mismatch fails. herdr does not do a handoff without the guard:

```json
{"id":"2","method":"server.live_handoff","params":{"expected_protocol":20,"expected_version":"0.8.2"}}
{"id":"2","error":{"code":"handoff_failed","message":"handoff stream closed while reading line"}}
```

Checked 2026-10-17 against herdr 0.9.5, on scratch servers. The checks included these items:

- the malformed `params` cases and the two bad `import_exe` cases
- three guard mismatches: `expected_protocol` only, `expected_version` only, and both
- the `/bin/true` timeout, with seven methods sent during the wait
- real handoffs with a subscription open, with a check of `terminal_id`s, argv, and the parent process
- one `herdr server live-handoff` CLI handoff

The pass did not run the `herdr update --handoff` entry point. It also did not do a handoff to a different herdr version, because the host had only one herdr version.

## server.reload_agent_manifests

This method loads the local agent manifest overrides again from the disk. It returns the active manifest set that results. This method reads the overrides and the cached remote manifests again, and it selects the active manifests again. It does not do a remote fetch. `herdr server update-agent-manifests` does the remote fetch. Thus, the result has no top-level `last_check_unix` or `last_result` field. [server.agent_manifests](#serveragent_manifests) does not read the files again.

A local override manifest is at `$XDG_CONFIG_HOME/herdr/agent-detection/<agent>.toml`. The default path is `~/.config/herdr/agent-detection/<agent>.toml`. After you edit an override, call this method to apply the change to a running server. herdr does not watch the files. Without a reload, the server continues to use the old manifest.

herdr accepts some bad override files without an error:

- For a malformed override file, only a `warning` on the entry of that agent kind shows the problem. `source` and `source_kind` stay on the remote or bundled manifest.
- An override file for an agent kind that herdr does not know adds no entry and no warning.

**Params**: `EmptyParams`, which is `{}`. No fields.

**Result**, with `type: "agent_manifest_reload"`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `type` | string const `"agent_manifest_reload"` | yes | — | Result discriminator. |
| `manifests` | array of `AgentManifestInfo` | yes | — | The active manifest set after the reload. Refer to the `AgentManifestInfo` field table in [server.agent_manifests](#serveragent_manifests). |

A malformed override for `codex` gives this entry:

```json
{"agent":"codex","source":"remote:/home/user/.local/state/herdr/agent-detection/remote/codex.toml","source_kind":"remote","active_version":"2026.10.11.1","cached_remote_version":"2026.10.11.1","local_override_shadowing_remote":true,"remote_update_result":"current","remote_last_checked_unix":1792263075,"warning":"ignored override /home/user/.config/herdr/agent-detection/codex.toml because it could not be loaded: TOML parse error at line 1, column 6\n  |\n1 | this is [ broken\n  |      ^\nexpected `.`, `=`\n"}
```

**Errors**: none occurred. Other codes are possible.

**CLI**: `herdr server reload-agent-manifests`. It loads the local overrides again. It prints the full JSON response envelope on stdout, with the keys in alphabetical order. It has no `--json` flag. `--json` gives a usage error with exit status 2. Compare `herdr server update-agent-manifests`. That subcommand first does a network fetch in the client, and it writes `status.toml` again. Then it sends this method and then `server.agent_manifests`. It has no method of its own.

**Example**

```json
{"id":"1","method":"server.reload_agent_manifests","params":{}}
{"id":"1","result":{"type":"agent_manifest_reload","manifests":[{"agent":"pi","source":"remote:/home/user/.local/state/herdr/agent-detection/remote/pi.toml","source_kind":"remote","active_version":"2026.10.01.1","cached_remote_version":"2026.10.01.1","local_override_shadowing_remote":false,"remote_update_result":"current","remote_last_checked_unix":1792262884}, …]}}
```

`Checked 2026-10-17 against herdr 0.9.5`, with `XDG_CONFIG_HOME` and `XDG_STATE_HOME` in a private directory: all claims in this section. These claims include:

- the result keys, the override location, and the missing file watch
- the malformed override and the override for an unknown agent kind
- the reload of the cached remote manifests, and the CLI output

 The examples show real responses from a scratch server, with placeholder paths.

## server.reload_config

This method reads `config.toml` again from the disk and applies it to the running server. It returns the result of the reload and any diagnostics from the parse or the application of the new configuration. With this method, configuration edits take effect without a server restart.

The global menu of the TUI has a "reload config" action, which does the same operation. `Documented by herdr agent-guide (herdr 0.9.5); not live-checked: no TUI client attached to the scratch servers`.

**Params**: `EmptyParams`, which is `{}`. No fields.

**Result**, with `type: "config_reload"`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `type` | string const `"config_reload"` | yes | — | Result discriminator. |
| `status` | `ConfigReloadStatus` enum | yes | — | Result of the reload. The value is `"applied"` when herdr applied all settings. It is `"partial"` when herdr applied some settings and refused other settings. It is `"failed"` when herdr applied no settings. |
| `diagnostics` | array of string | yes | — | Messages for persons about settings that herdr refused or changed. The array is empty after a clean `applied` reload. |

These configurations gave these results:

| configuration | `status` | `diagnostics` |
| --- | --- | --- |
| valid file, or no file | `applied` | `[]` |
| `[ui] sidebar_width = "wide"` | `partial` | `["invalid ui config: invalid type: string \"wide\", expected u16\nin `sidebar_width`\n; keeping current ui settings"]` |
| `[theme] name = "no-such-theme"` | `partial` | `["unknown theme name theme.name = \"no-such-theme\"; using \"catppuccin\"; valid themes: catppuccin, ... vesper"]` |
| `[ui] bogus = 1` | `partial` | `["unknown config key ui.bogus; ignoring key"]` |
| TOML that herdr cannot parse | `failed` | `["config parse error: TOML parse error at line 1, column 6\n  |\n1 | this is [ not toml\n  |      ^\nexpected `.`, `=`\n; keeping current config"]` |

herdr reports a key at the wrong level only as an unknown key. It does not say that the key has the wrong position. For example, `manifest_check = false` at the top level gives `status: "partial"` with `diagnostics: ["unknown config key manifest_check; ignoring key"]`. This key must be in `[update]`, where it gives `applied`. Unknown keys never cause a failed reload.

Some settings need a server restart, and this method does not apply them. An example is the Windows setting `allow_unelevated_clients`. Refer to [environment.md](../environment.md#no-token-or-api-key-authentication). `Documented by herdr --default-config (herdr 0.9.5); not live-checked: the setting is for Windows only, and the pass ran on Linux`.

**Errors**: none occurred with a valid configuration. A malformed configuration shows in `status` (`partial` or `failed`) and `diagnostics`, not in a protocol error. Other codes are possible.

**CLI**: `herdr server reload-config`.

- With no server, it prints a JSON error envelope to stderr and exits with status 1. An example is `{"id":"cli:server:reload-config","error":{"code":"server_not_running","message":"no herdr server is running at <path>; run \`herdr\` to start or attach it"}}`. A socket file with no listener gives the same envelope. This output is different from the plain sentence of `server stop`. Refer to [server.stop](#serverstop).
- On success, it prints the full response envelope to stdout. It has no `--json` flag. `--json` gives a usage error with exit status 2.

**Example**

```json
{"id":"cli:server:reload-config","method":"server.reload_config","params":{}}
{"id":"cli:server:reload-config","result":{"type":"config_reload","status":"applied","diagnostics":[]}}
```

`Checked 2026-10-17 against herdr 0.9.5`, with `XDG_CONFIG_HOME` in a private directory: all claims in this section, except the TUI menu and the Windows setting. These claims include the five configurations in the table, the key at the wrong level, and the CLI output.

## server.ssh_agent.register

This method is new in 0.9.3. It registers a remote-host SSH agent socket with the server. The registration applies only to the API connection that sent it. The schema says: "Registration lasts until this API connection closes". Thus, a caller sends the request and reads the one-line `{"type":"ok"}` response. Then the caller keeps the connection open while the registration must stay active. The server sends nothing more on that connection. Refer to [protocol.md](../protocol.md#ssh-agent-registration-connections).

`ping` shows support with the capability flag `ssh_agent_registration`. The default is `false`, and herdr 0.9.5 gives `true`. Refer to [protocol.md](../protocol.md#pingpong-handshake).

The effect of a registration on a local scratch server depends on `SSH_AUTH_SOCK` when the server starts.

When the server starts with an `SSH_AUTH_SOCK` from its environment:

- At startup, the server makes the `herdr.sock.agent` symlink in the session directory. The symlink points to the agent socket from the environment.
- A registration does not change the symlink target, while the registration is open and after it closes.
- The registered socket gets no connection.
- The server connected to the socket from the environment two times at startup. It connected one more time immediately after the registration connection closed.

When the server starts without `SSH_AUTH_SOCK`:

- At startup, the server makes no `herdr.sock.agent` symlink.
- A registration makes `herdr.sock.agent`, which points to the registered socket.
- The server connects one time to the registered socket immediately after the registration.
- In a pane, `SSH_AUTH_SOCK` is `<session-dir>/herdr.sock.agent`. A connection from the pane gets to the registered socket.
- With two registrations open at the same time, the first registration stays the target. When the first one closes, the symlink points to the second socket.
- When all registrations close, the symlink points to `<session-dir>/herdr.sock.unavailable`. That path does not exist.

Thus, a registration gives the agent socket of the session when the server has no agent socket from its environment. An agent socket from the environment has priority. The 2026-10-06 pass on herdr 0.9.3 checked only the first case. It is not known if herdr 0.9.3 had the same behavior in the second case. The server log does not show a registration in either case.

The schema calls the socket a "remote-host agent socket". This text suggests that the method is for remote attach (`herdr --remote`). But no herdr caller of this method occurred, because the pass did not use remote attach. The effect of a registration on a remote session is not known.

Evidence from the second case. Each line shows a `readlink` of `herdr.sock.agent` or the output of a command in a pane:

```text
held: /home/user/project/agent.sock
SSHSOCK=/home/user/.config/herdr/sessions/demo/herdr.sock.agent
CON
closed+1s: /home/user/.config/herdr/sessions/demo/herdr.sock.unavailable
```

**Params**, with type `ServerSshAgentRegisterParams`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `socket_path` | string | yes | — | Absolute path to the agent socket. It must be an existing Unix socket that the calling user owns. An empty string, a relative path, a missing path, a regular file, a directory, and a socket of a different owner fail with `invalid_ssh_agent`. herdr accepts a socket file with no listener. A value that is not a string fails with `invalid_request`. |

**Result**, with `type: "ok"`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `type` | string const `"ok"` | yes | — | herdr accepted the registration. The connection stays open after this response. |

**Errors**:

| code | when |
| --- | --- |
| `invalid_ssh_agent` | `socket_path` is not an absolute path to an existing socket that the user owns. The message is `SSH agent must be an absolute, user-owned socket`. A socket that `root` owns gives the same error and message. |
| `invalid_request` | `socket_path` is missing, `null`, or not a string. |

Other codes are possible.

**CLI**: API only, with no CLI subcommand. `herdr server --help` lists no SSH agent subcommand.

**Example**

```json
{"id":"1","method":"server.ssh_agent.register","params":{"socket_path":"/home/user/project/agent.sock"}}
{"id":"1","result":{"type":"ok"}}
{"id":"2","method":"server.ssh_agent.register","params":{"socket_path":"relative/agent.sock"}}
{"id":"2","error":{"code":"invalid_ssh_agent","message":"SSH agent must be an absolute, user-owned socket"}}
```

`Checked 2026-10-17 against herdr 0.9.5`, on scratch servers only, with listener sockets in a private directory. The check included the response, the open connection, the capability flag, and the two startup cases. It also included the parameter errors and the server log. The pass did not use remote attach.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the schema text "remote-host agent socket".

## server.stop

This method stops the running server through the socket API, and it ends all its sessions. A server holds exactly one session, so the method always stops that session.

After the server sends the acknowledgement, it stops in an orderly sequence:

- It closes each connection. This includes an open `events.subscribe` stream. The subscriber reads EOF, and it gets no closing event.
- It removes the socket file.
- For a short time after the `ok` response, the socket can still accept connections and answer. In one test, connections succeeded for 0.091 s, and a `ping` on one of them got a full `pong`.
- After the stop completes, a new connection fails with `FileNotFoundError`. It does not wait.

The session directory stays with `herdr-server.log`. It also keeps `herdr.sock.agent` when that symlink existed.

The CLI prints nothing on success, with an empty stdout and exit status 0. But it does not return when the acknowledgement arrives. It connects to the socket again and again, about 40 empty connections each second, until a connection fails. If the socket still accepts connections after 15 s, it exits with status 1 and this message: `server did not stop within 15000ms; sockets are still reachable at <path>`. In tests, a normal stop completed in less than 0.2 s.

**Params**: `EmptyParams`, which is `{}`. No fields.

**Result**, with `type: "ok"`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `type` | string const `"ok"` | yes | — | Success acknowledgement. Then the server stops. |

**Errors**: none occurred for a request that has the correct form. If no server runs, the CLI writes `server is not running or cannot be reached at <path>: <OS error>` on stderr and exits with status 1. The OS error is different for the two cases:

- For a socket path that does not exist, it is `No such file or directory (os error 2)`.
- For a socket file with no listener, it is `Connection refused (os error 111)`.

**CLI**: `herdr server stop`. Do not stop the herdr server of a live session during agent tests. Use a named session for tests that need a separate server. `Documented by herdr --skill (herdr 0.9.5); not live-checked: this text is a rule for persons and agents, not a server behavior`.

**Example**

```json
{"id":"cli:session:stop","method":"server.stop","params":{}}
{"id":"cli:session:stop","result":{"type":"ok"}}
```

The CLI sends the id `cli:session:stop`, not `cli:server:stop`. On this page, it is the only `herdr server` subcommand with an id that is not in the form `cli:server:<subcommand>`.

Checked 2026-10-17 against herdr 0.9.5:

- A raw socket stop with a subscription open ran on a scratch server.
- A fake socket recorded the CLI request id, the empty stdout, and the 15 s timeout. The timeout run made 601 connections in 15.0 s.
- The no-server messages came from a missing path and from a socket file with no listener.
- CLI stops of two scratch servers gave an empty stdout, an empty stderr, and exit status 0, in 155 ms and 104 ms.

`{"type":"ok"}` is the only result variant in the schema for this method, except errors. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`.
