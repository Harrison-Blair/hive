# herdr API: integration methods

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Index: [README.md](../README.md). Wire format: [protocol.md](../protocol.md).

The `integration.*` methods install and remove the built-in herdr integrations for third-party agent programs.

- Each method operates on one `target`. The target is an enum value that names a supported agent program. herdr knows how to change the configuration of that program.
- `integration.install` writes the integration into the configuration of the target. `integration.uninstall` removes it.
- Both methods change agent configuration files on the disk. Each result has a `messages` array. Each message names one change and the absolute path of the file.
- `integration.list` reads the state of each integration. It changes nothing.

The server process finds the paths from its own home directory, not from the home directory of the caller. A request sends only `{"target":...}`, so the paths come from the server. Thus, a server that runs under one account edits the agent configuration of that account. A pass set `HOME`, `CLAUDE_CONFIG_DIR`, `CODEX_HOME`, and `XDG_CONFIG_HOME` to the same private directory. Thus, the pass did not find which of these variables has priority.

Send a well-formed `integration.install` or `integration.uninstall` request only to a scratch server. The home and configuration variables of that server must point to a private directory. Otherwise, the request edits the real agent configuration of the account.

An integration can give herdr the agent status of an agent, the restore of its agent session, or both. For example, `herdr integration install claude` adds the restore of the agent session. herdr still detects the agent status of Claude from the screen.

`Documented by herdr agent-guide (herdr 0.9.5); not live-checked: no pass ran an agent with an installed integration`: this stamp applies to the paragraph above.

The `herdr integration` CLI commands do not use these methods:

- `herdr integration status`, `install`, and `uninstall` run in the CLI process against `$HOME`. They do not open the herdr socket.
- A fake socket server recorded every byte that it received. Five `herdr integration` calls sent no request line to it. These calls were `status`, `status --outdated-only`, `install claude`, `uninstall claude`, and `install claude` again.
- In the same test, `herdr tab list` and `herdr agent list` each sent `{"id":"api-client:status","method":"ping","params":{}}`.
- When `HERDR_SOCKET_PATH` points to a path that does not exist, `herdr integration status` exits with status `0` and writes its full output. In the same condition, `herdr tab list` fails with `server_not_running`. Neither call made the session directory.
- The CLI `install` and `uninstall` write the same message lines that the `messages` array of the method result has.

Thus, no first-party CLI command sends these methods. The **CLI** line in each section below names the nearest CLI command. It does not name a request that the CLI sends. [cli-mapping.md](../cli-mapping.md#cli-conventions) gives the general CLI conventions.

`Checked 2026-10-17 against herdr 0.9.5`: this stamp applies to all other claims in this introduction. A scratch server ran with `HOME`, every `XDG_*` variable, `CLAUDE_CONFIG_DIR`, and `CODEX_HOME` set to a private directory.

3 methods:

| method | purpose |
| --- | --- |
| [integration.install](#integrationinstall) | Install a built-in agent integration into a target program. |
| [integration.list](#integrationlist) | List all known integration targets and their installation state. |
| [integration.uninstall](#integrationuninstall) | Remove an installed integration from a target program. |

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the list of three `integration.*` methods.

## integration.install

Installs the built-in herdr integration into the named target agent program. This method changes the configuration of the target on the disk. The `details.messages` list of the result gives the changes that herdr made.

**Params** (`IntegrationInstallParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `target` | `IntegrationTarget` (enum) | yes | — | The agent program that gets the integration. Refer to [IntegrationTarget](#integrationtarget-enum). |

**Result** — `type: "integration_install"`:

| field | type | required | meaning |
| --- | --- | --- | --- |
| `type` | `"integration_install"` (const) | yes | Result discriminator. |
| `target` | `IntegrationTarget` (enum) | yes | The same target as the request. Refer to [IntegrationTarget](#integrationtarget-enum). |
| `details` | `IntegrationInstallResult` object | yes | Data about the installation. Refer to the table below. |

`IntegrationInstallResult`:

| field | type | required | meaning |
| --- | --- | --- | --- |
| `messages` | `string[]` | yes | Text for a person. Each message gives one file that herdr wrote or made sure of, with its absolute path. |

Every success had exactly the fields `type`, `target`, and `details.messages`. The number of messages changes with the target. These are the messages for each target that succeeded:

```
claude:     ["installed claude integration hook to /home/user/.claude/hooks/herdr-agent-state.sh","ensured claude settings at /home/user/.claude/settings.json"]
codex:      ["installed codex integration hook to /home/user/.codex/herdr-agent-state.sh","ensured codex hooks at /home/user/.codex/hooks.json","ensured codex config at /home/user/.codex/config.toml"]
pi:         ["installed pi integration to /home/user/.pi/agent/extensions/herdr-agent-state.ts"]
opencode:   ["installed opencode integration plugin to /home/user/.config/opencode/plugins/herdr-agent-state.js","installed opencode tui integration plugin to /home/user/.config/opencode/herdr-tui-session.js","ensured opencode tui plugin config at /home/user/.config/opencode/tui.jsonc"]
cursor:     ["installed cursor integration hook to /home/user/.cursor/herdr-agent-state.sh","updated cursor hooks at /home/user/.cursor/hooks.json"]
mastracode: ["installed mastracode integration hook to /home/user/.mastracode/hooks/herdr-agent-state.sh","ensured mastracode hooks at /home/user/.mastracode/hooks.json"]
```

You can send the install again. A second install of `claude` succeeded with the same two messages. It also succeeded over an `outdated` hook, and it set the state back to `current`.

**Errors**:

- Request validation gives `invalid_request`. This validation fails before the request gets to the handler, so herdr writes nothing.
  - `params: {}` gives ``invalid request: missing field `target` at line 1 column 54``.
  - A value that is not in the enum gives ``invalid request: unknown variant `zzz_not_a_real_target`, expected one of `pi`, `omp`, … `grok` ``.
  - `{"target":5}` gives ``invalid request: invalid type: integer `5`, expected string or map``.
  - The column number depends on the bytes of the request. The value `54` is for the request id `i1`.
  - These errors contain the request id. herdr 0.9.1 gave `"id":""`.
- The handler gives `integration_install_failed`. The message comes from the step that failed. A pass found two causes:
  - The configuration directory of the target does not exist. In an empty home directory, every target except `mastracode` failed for this cause. This failure includes `claude` and `codex`. Only `mastracode` makes its own directory.
  - An existing configuration file is not valid. A `.claude/settings.json` file with the text `{not json` gave this error.
- A failed install is not atomic. With the bad `settings.json` and no hook file, the failed install still wrote `.claude/hooks/herdr-agent-state.sh`.
- [errors.md](../errors.md) gives the error envelope and the error code catalog.

```
{"id":"in-claude","error":{"code":"integration_install_failed","message":"claude directory not found at /home/user/.claude. install claude code first"}}
{"id":"in-codex","error":{"code":"integration_install_failed","message":"codex config directory not found at /home/user/.codex. install codex first"}}
{"id":"in-pi","error":{"code":"integration_install_failed","message":"pi extension directory not found at /home/user/.pi/agent/extensions. install pi first"}}
{"id":"in-omp","error":{"code":"integration_install_failed","message":"omp extension directory not found at /home/user/.omp/agent/extensions. install omp first"}}
{"id":"in-copilot","error":{"code":"integration_install_failed","message":"copilot config directory not found at /home/user/.copilot. install github copilot cli first"}}
{"id":"in-devin","error":{"code":"integration_install_failed","message":"devin config directory not found at /home/user/.config/devin. install devin cli first"}}
{"id":"in-droid","error":{"code":"integration_install_failed","message":"droid config directory not found at /home/user/.factory. install droid first"}}
{"id":"in-kimi","error":{"code":"integration_install_failed","message":"kimi code config directory not found at /home/user/.kimi-code. install kimi code first"}}
{"id":"in-opencode","error":{"code":"integration_install_failed","message":"opencode config directory not found at /home/user/.config/opencode. install opencode first"}}
{"id":"in-kilo","error":{"code":"integration_install_failed","message":"kilo config directory not found at /home/user/.config/kilo. install kilo first"}}
{"id":"in-hermes","error":{"code":"integration_install_failed","message":"hermes config directory not found at /home/user/.hermes. install hermes agent first"}}
{"id":"in-qodercli","error":{"code":"integration_install_failed","message":"qodercli config directory not found at /home/user/.qoder. install qodercli first"}}
{"id":"in-qwen","error":{"code":"integration_install_failed","message":"qwen code config directory not found at /home/user/.qwen. install qwen code first"}}
{"id":"in-cursor","error":{"code":"integration_install_failed","message":"cursor config directory not found at /home/user/.cursor. install cursor agent cli first"}}
{"id":"in-antigravity_cli","error":{"code":"integration_install_failed","message":"antigravity cli config directory not found at /home/user/.gemini/config. install antigravity cli first"}}
{"id":"in-grok","error":{"code":"integration_install_failed","message":"grok config directory not found at /home/user/.grok. install grok cli first"}}
{"id":"i","error":{"code":"integration_install_failed","message":"failed to parse /home/user/.claude/settings.json: key must be a string at line 1 column 2"}}
```

**Events**: none. The `Subscription` enum of the schema has no `integration.*` variant. A subscriber held all 24 subscription types that need only a `type` field. It got only the `subscription_started` acknowledgement in 5 seconds around a successful install. It got the same result around an install over an `outdated` hook.

**CLI**: `herdr integration install <TARGET>`. This command does the install in the CLI process and does not send this method.

- `herdr integration install --help` lists 18 possible values, one more than the wire enum: `pi, omp, claude, codex, copilot, devin, droid, kimi, opencode, kilo, hermes, qodercli, qwen, cursor, mastracode, antigravity-cli, grok, letta`.
- `letta` has no wire equivalent. Refer to [IntegrationTarget](#integrationtarget-enum).
- The help spells the antigravity target `antigravity-cli`, with a hyphen. The wire enum value is `antigravity_cli`, with an underscore. The wire rejects the hyphen spelling as an unknown variant.
- The CLI accepts the two spellings. `herdr integration install antigravity_cli` and `herdr integration install antigravity-cli` gave the same result.
- The herdr parser rejects an unknown CLI target with exit status `2`. The `currently supported` list in this message puts `letta` after `qwen`. The `--help` list puts it last.
- A failed install exits with status `1` and writes the handler message. A successful install exits with status `0`.

```
$ herdr integration install zzz
unknown integration target: zzz
currently supported: pi, omp, claude, codex, copilot, devin, droid, kimi, opencode, kilo, hermes, qodercli, qwen, letta, cursor, mastracode, antigravity-cli, grok
```

**Example**

```json
{"id":"1","method":"integration.install","params":{"target":"claude"}}
{"id":"1","result":{"type":"integration_install","target":"claude","details":{"messages":["installed claude integration hook to /home/user/.claude/hooks/herdr-agent-state.sh","ensured claude settings at /home/user/.claude/settings.json"]}}}
```

`Checked 2026-10-17 against herdr 0.9.5`: this stamp applies to all claims and captures in this section. The scratch server had a private home directory. The schema gives `IntegrationInstallResult` as an object with one required `messages` string array.

## integration.list

Lists every integration target that herdr knows on the wire. Each entry gives these values:

- the target and its display label
- the CLI command that herdr looks for
- if the integration is available
- the installation state of the integration

This method is read-only. It changes nothing. The CLI knows one more target than this method returns. Refer to [IntegrationTarget](#integrationtarget-enum).

**Params**: `EmptyParams`. Send `{}`. The `params` key must be in the request, but the server does not do a strict check of its contents.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| _(none)_ | — | — | — | herdr reads no parameters. `params` must be in the request. It must be a JSON object with any keys, or an empty array. herdr ignores all keys. |

`EmptyParams` accepts unknown fields:

- `{}`, `{"outdated_only":true}`, and `{"a":1,"b":[1,2],"c":{"d":null}}` each gave the full list of 17 entries. `[]` gave the same list.
- Thus, a caller who types a wrong param name gets a success, not an error.
- herdr also ignores unknown top-level request keys, such as `extra` and `jsonrpc`.
- The schema gives `EmptyParams` only as an object with no properties.

**Result** — `type: "integration_list"`:

| field | type | required | meaning |
| --- | --- | --- | --- |
| `type` | `"integration_list"` (const) | yes | Result discriminator. |
| `integrations` | `IntegrationInfo[]` | yes | One entry for each known integration target. Refer to the table below. |

`IntegrationInfo`:

| field | type | required | meaning |
| --- | --- | --- | --- |
| `target` | `IntegrationTarget` (enum) | yes | The integration target of this entry. Refer to [IntegrationTarget](#integrationtarget-enum). |
| `label` | `string` | yes | Display name of the target. |
| `command` | `string` | yes | The name of the CLI command that herdr looks for, for this target. |
| `available` | `boolean` | yes | `true` when herdr finds `command` on the `PATH` of the server. It is also `true` when the target has an installed integration, with `state` `current` or `outdated`. |
| `state` | `IntegrationState` (enum) | yes | Installation state of the integration. One of `not_installed`, `current`, `outdated`. |

Field behavior:

- All 17 entries had exactly these five keys, with no null values.
- herdr looks for `command` on the `PATH` of the server process at each call. It does not keep the result. When a pass deleted a fake `grok` executable, `grok` changed to `available: false`.
- On a server with a `PATH` that did not have `claude`, `pi`, `codex`, `opencode`, or `cursor-agent`, these five installed targets showed `available: true`.
- After `uninstall pi`, `pi` changed to `available: false` and `not_installed`. The extensions directory stayed, so herdr does not use a directory check.
- On a server with the normal `PATH` and no installed integration, `available` was `true` only for the targets whose `command` `command -v` found.
- `cursor` (`cursor-agent`) and `antigravity_cli` (`agy`) are the only targets with a `command` that is not the same as the target name.
- A pass saw all three `state` values. The installed hook has a version line, such as `# HERDR_INTEGRATION_VERSION=11` in the claude hook. A change of this value to `10` gave `outdated`. A new install set the state back to `current`.

```
{"target":"claude","label":"claude","command":"claude","available":true,"state":"outdated"}
```

**Errors**:

- Request validation gives `invalid_request`. The column numbers depend on the bytes of the request.
  - When `params` is not in the request, the error is ``invalid request: missing field `params` at line 1 column 38``.
  - `params` with the value `null`, `""`, `true`, or `5` gives `invalid request: invalid type: <value>, expected struct EmptyParams`.
  - `params: [1,2]` gives `invalid request: invalid length 2, expected 0 elements in sequence`.
- Since herdr 0.9.3, these errors contain the request id. For example, a request with the id `e` gets `"id":"e"`. herdr 0.9.1 gave `"id":""`.
- A pass found no handler error. With a bad `.claude/settings.json`, install and uninstall failed, but `integration.list` succeeded with 17 entries. Other codes are possible. Refer to [errors.md](../errors.md).

```
"params":null  -> "invalid request: invalid type: null, expected struct EmptyParams at line 1 column 52"
"params":""    -> "invalid request: invalid type: string \"\", expected struct EmptyParams at line 1 column 50"
"params":true  -> "invalid request: invalid type: boolean `true`, expected struct EmptyParams at line 1 column 52"
"params":5     -> "invalid request: invalid type: integer `5`, expected struct EmptyParams at line 1 column 49"
"params":[1,2] -> "invalid request: invalid length 2, expected 0 elements in sequence at line 1 column 53"
```

**Events**: none. A subscriber held all 24 subscription types that need no `pane_id`. It got only the `subscription_started` acknowledgement in 4 seconds across three `integration.list` calls. The `Subscription` enum of the schema has no `integration.*` variant.

**CLI**: `herdr integration status [--outdated-only]` is the nearest CLI command. It does not send this method.

- It opens no socket. It calculates all values from `$HOME` in the CLI process.
- Its output was identical against a real server, against a fake socket server, and against a socket path that does not exist.
- Each line gives the target, the state, the installed hook version, and the path of the hook file.
- It does not show `available` or `command`. It adds an 18th line for `letta`.
- `integration.list` does not give the hook version or the hook path.
- The status line uses the CLI spelling `antigravity-cli`.

```
$ herdr integration status
pi: current (v9) (/home/user/.pi/agent/extensions/herdr-agent-state.ts)
omp: not installed (/home/user/.omp/agent/extensions/herdr-omp-agent-state.ts)
claude: current (v11) (/home/user/.claude/hooks/herdr-agent-state.sh)
codex: current (v10) (/home/user/.codex/herdr-agent-state.sh)
copilot: not installed (/home/user/.copilot/hooks/herdr-agent-state.sh)
devin: not installed (/home/user/.config/devin/herdr-agent-state.sh)
droid: not installed (/home/user/.factory/hooks/herdr-agent-state.sh)
kimi: not installed (/home/user/.kimi-code/hooks/herdr-agent-state.sh)
opencode: current (v13) (/home/user/.config/opencode/plugins/herdr-agent-state.js)
kilo: not installed (/home/user/.config/kilo/plugin/herdr-agent-state.js)
hermes: not installed (/home/user/.hermes/plugins/herdr-agent-state/__init__.py)
qodercli: not installed (/home/user/.qoder/hooks/herdr-agent-state.sh)
qwen: not installed (/home/user/.qwen/hooks/herdr-agent-session.sh)
cursor: current (v2) (/home/user/.cursor/herdr-agent-state.sh)
mastracode: not installed (/home/user/.mastracode/hooks/herdr-agent-state.sh)
antigravity-cli: not installed (/home/user/.gemini/config/hooks/herdr-agent-state.sh)
grok: not installed (/home/user/.grok/hooks/herdr-agent-state.sh)
letta (experimental): not installed (/home/user/.letta/hooks/herdr-agent-session.sh)
```

`--outdated-only` is a local option, not a request param. It does not filter the status lines:

- When every installed integration is current, it writes nothing and exits with status `0`.
- Otherwise, it writes one line. The line names the `herdr integration install` commands to run.
- `herdr integration status --help` shows only the `--outdated-only` option.

```
$ herdr integration status --outdated-only
installed herdr integrations need updating; run herdr integration install claude.
```

**Example**

The scratch server had the integrations for `pi`, `claude`, `codex`, `opencode`, and `cursor`. Their five commands were on the `PATH` of the server.

```json
{"id":"2","method":"integration.list","params":{}}
{"id":"2","result":{"type":"integration_list","integrations":[{"target":"pi","label":"pi","command":"pi","available":true,"state":"current"},{"target":"omp","label":"omp","command":"omp","available":false,"state":"not_installed"},{"target":"claude","label":"claude","command":"claude","available":true,"state":"current"},{"target":"codex","label":"codex","command":"codex","available":true,"state":"current"},{"target":"copilot","label":"copilot","command":"copilot","available":false,"state":"not_installed"},{"target":"devin","label":"devin","command":"devin","available":false,"state":"not_installed"},{"target":"droid","label":"droid","command":"droid","available":false,"state":"not_installed"},{"target":"kimi","label":"kimi","command":"kimi","available":false,"state":"not_installed"},{"target":"opencode","label":"opencode","command":"opencode","available":true,"state":"current"},{"target":"kilo","label":"kilo","command":"kilo","available":false,"state":"not_installed"},{"target":"hermes","label":"hermes","command":"hermes","available":false,"state":"not_installed"},{"target":"qodercli","label":"qodercli","command":"qodercli","available":false,"state":"not_installed"},{"target":"qwen","label":"qwen","command":"qwen","available":false,"state":"not_installed"},{"target":"cursor","label":"cursor","command":"cursor-agent","available":true,"state":"current"},{"target":"mastracode","label":"mastracode","command":"mastracode","available":false,"state":"not_installed"},{"target":"antigravity_cli","label":"antigravity-cli","command":"agy","available":false,"state":"not_installed"},{"target":"grok","label":"grok","command":"grok","available":false,"state":"not_installed"}]}}
```

The `available` and `state` values in the example come from the installed integrations and from the `PATH` of the server. They are not a fixed fixture.

`Checked 2026-10-17 against herdr 0.9.5`: this stamp applies to all claims and captures in this section. A pass sent 50 calls in sequence. Without the id, all 50 responses were byte-identical. The md5 of each of the 13 installed integration files did not change. The target order was the same in each response, and it was the order of the schema enum.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the required `IntegrationInfo` fields and the `IntegrationState` enum. A pass also saw these fields and all three states on a server.

## integration.uninstall

Removes an installed herdr integration from the named target agent program. This method changes the configuration of the target on the disk. The `details.messages` list of the result gives the changes that herdr made.

**Params** (`IntegrationUninstallParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `target` | `IntegrationTarget` (enum) | yes | — | The agent program that loses the integration. Refer to [IntegrationTarget](#integrationtarget-enum). |

**Result** — `type: "integration_uninstall"`:

| field | type | required | meaning |
| --- | --- | --- | --- |
| `type` | `"integration_uninstall"` (const) | yes | Result discriminator. |
| `target` | `IntegrationTarget` (enum) | yes | The same target as the request. Refer to [IntegrationTarget](#integrationtarget-enum). |
| `details` | `IntegrationUninstallResult` object | yes | Data about the removal. Refer to the table below. |

`IntegrationUninstallResult`:

| field | type | required | meaning |
| --- | --- | --- | --- |
| `messages` | `string[]` | yes | Text for a person. Each message gives one removed file or entry, or one item that herdr did not find, with its absolute path. |

Every success had exactly the fields `type`, `target`, and `details.messages`.

- `uninstall claude` removed the hook file and the hook entries. herdr kept `settings.json`, and its contents changed to `{"hooks":{}}`.
- An uninstall of a target that is not installed succeeds with no error. The messages tell that herdr found nothing. This also applies to a target with no configuration directory, such as `omp`.

```
uninstall mastracode -> ["removed mastracode hook at /home/user/.mastracode/hooks/herdr-agent-state.sh","removed herdr mastracode hook entries from /home/user/.mastracode/hooks.json"]
uninstall pi -> ["removed pi integration extension at /home/user/.pi/agent/extensions/herdr-agent-state.ts"]
```

```json
{"id":"2","method":"integration.uninstall","params":{"target":"claude"}}
{"id":"2","result":{"type":"integration_uninstall","target":"claude","details":{"messages":["no claude hook found at /home/user/.claude/hooks/herdr-agent-state.sh","no herdr claude hook entries found in /home/user/.claude/settings.json"]}}}
{"id":"3","method":"integration.uninstall","params":{"target":"omp"}}
{"id":"3","result":{"type":"integration_uninstall","target":"omp","details":{"messages":["no omp integration extension found at /home/user/.omp/agent/extensions/herdr-omp-agent-state.ts"]}}}
```

**Errors**:

- Request validation gives `invalid_request`, as for `integration.install`. This validation fails before the handler, so herdr removes nothing.
  - `params: {}` gives ``invalid request: missing field `target` at line 1 column 56``. The value `56` is for the request id `u1`.
  - A value that is not in the enum gives ``invalid request: unknown variant `zzz_not_a_real_target`, expected one of `pi`, `omp`, … `grok` ``. The list has the same 17 values as for `install`.
  - `{"target":"letta"}` gives ``invalid request: unknown variant `letta` ``, with the same 17 values.
  - These errors contain the request id. herdr 0.9.1 gave `"id":""`.
- The handler gives `integration_uninstall_failed`. A `.claude/settings.json` file with the text `{not json` gave this error. The failed uninstall did not remove the hook file.
- [errors.md](../errors.md) gives the error envelope and the error code catalog.

```
{"id":"u","error":{"code":"integration_uninstall_failed","message":"failed to parse /home/user/.claude/settings.json: key must be a string at line 1 column 2"}}
```

**Events**: none. The `Subscription` enum of the schema has no `integration.*` variant. A subscriber held all 24 subscription types that need only a `type` field. It got only the `subscription_started` acknowledgement in 5 seconds around a successful `integration.uninstall` of `claude`.

**CLI**: `herdr integration uninstall <TARGET>`. This command does the removal in the CLI process and does not send this method.

- `herdr integration uninstall --help` lists the same 18 possible values as `install`. The list includes `letta`, which has no wire equivalent.
- As for `install`, the help spells the antigravity target `antigravity-cli`, with a hyphen. The wire enum value is `antigravity_cli`, with an underscore.
- The CLI also accepts `antigravity_cli`. `herdr integration uninstall antigravity_cli` exited with status `0`. Its messages were `no antigravity-cli hook found at /home/user/.gemini/config/hooks/herdr-agent-state.sh` and `no herdr antigravity-cli hook entries found in /home/user/.gemini/config/hooks.json`.
- `herdr integration uninstall letta` exited with status `0`. Its messages were `no letta hook found at /home/user/.letta/hooks/herdr-agent-session.sh` and `no herdr letta hook entry found in /home/user/.letta/settings.json`.

**Example**

A pass installed `claude` before this request.

```json
{"id":"1","method":"integration.uninstall","params":{"target":"claude"}}
{"id":"1","result":{"type":"integration_uninstall","target":"claude","details":{"messages":["removed claude hook at /home/user/.claude/hooks/herdr-agent-state.sh","removed herdr claude hook entries from /home/user/.claude/settings.json"]}}}
```

`Checked 2026-10-17 against herdr 0.9.5`: this stamp applies to all claims and captures in this section. The scratch server had a private home directory.

## IntegrationTarget (enum)

This shared enum names the supported integration targets. herdr uses it in these places:

- the `target` param of `install` and `uninstall`
- the `target` field of each of their results
- the `target` field of each `IntegrationInfo` entry from `integration.list`

The wire and schema values are in the table below. The server does not return the tool names in the second column. These names are an inference from `label`, `command`, and the handler messages, such as `install github copilot cli first`.

| value | tool (inferred) |
| --- | --- |
| `pi` | pi agent |
| `omp` | omp agent |
| `claude` | Claude Code |
| `codex` | Codex |
| `copilot` | GitHub Copilot CLI |
| `devin` | Devin CLI |
| `droid` | Droid |
| `kimi` | Kimi Code |
| `opencode` | OpenCode |
| `kilo` | Kilo |
| `hermes` | Hermes Agent |
| `qodercli` | Qoder CLI |
| `qwen` | Qwen Code |
| `cursor` | Cursor Agent CLI |
| `mastracode` | Mastra Code |
| `antigravity_cli` | Antigravity CLI (CLI argument: `antigravity-cli`) |
| `grok` | Grok CLI |

The enum is identical in the request definition (`request/$defs/IntegrationTarget`) and in the response definition (`success_response/$defs/IntegrationTarget`). The 17 values have the same order in the schema, in `integration.list`, and in the unknown-variant error.

On the wire, the value must be the exact enum string. The server rejects `{"target":"antigravity-cli"}` with ``invalid request: unknown variant `antigravity-cli` ``. The CLI accepts `antigravity-cli` and `antigravity_cli`. The CLI never sends a target on the wire. Thus, the two spellings are a fact about the CLI arguments only. No translation between them is visible on the wire.

The CLI knows one target that the wire does not know:

- `herdr integration install --help` and `herdr integration uninstall --help` list an 18th possible value, `letta`.
- `herdr integration status` writes `letta (experimental): not installed (/home/user/.letta/hooks/herdr-agent-session.sh)` as an 18th line.
- `letta` is not in the schema enum. It is not in the 17 entries of `integration.list`. It is not in the unknown-variant error list of the server.
- The wire rejects `{"target":"letta"}` as an unknown variant for install and uninstall.
- Thus, a caller that uses only the wire API cannot get to the letta integration.

`Checked 2026-10-17 against herdr 0.9.5`: this stamp applies to the 17 values, their order, the two wire rejections, and the `letta` claims. A pass did not check the `tool (inferred)` column against more than the `label`, the `command`, and the handler messages.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the identical request and response definitions of the enum.
