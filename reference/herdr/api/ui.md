# herdr API: ui methods

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Index: [README.md](../README.md). Wire format: [protocol.md](../protocol.md).

The `ui` methods operate on the surfaces of a client. They control these items:

- the window title of the foreground client
- desktop and toast notifications
- the popup overlay of the client
- the surface interest lease of a client-shell connection
- commands that the client-shell projection shows
- the dismissal of product announcements and release notes

These methods operate on the foreground client, or on the connected client-shell endpoint. The target surface can be absent, disabled, or not available. Thus, some methods do not fail in these conditions. They return a boolean and a `reason` enum that tell if the action had an effect.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the method list and the result shapes in this introduction. Each method section gives the checked behavior.

8 methods:

| method | purpose |
| --- | --- |
| [client.window_title.clear](#clientwindow_titleclear) | Set the window title of the foreground client back to its default. |
| [client.window_title.set](#clientwindow_titleset) | Set the window title of the foreground client. |
| [client_shell.surface.set](#client_shellsurfaceset) | Get or release the surface interest lease of a client-shell connection. |
| [command.invoke](#commandinvoke) | Run a command that the client-shell projection shows. |
| [notification.show](#notificationshow) | Show a desktop or toast notification through the foreground client. |
| [popup.close](#popupclose) | Close the popup overlay of the client that is open now. |
| [product_announcement.dismiss](#product_announcementdismiss) | Dismiss a product announcement. |
| [release_notes.dismiss](#release_notesdismiss) | Dismiss the release notes for a version. |

## client.window_title.clear

This method removes a title override from the window of the foreground client. Then it writes the default title of the client. It operates only on the foreground client. When no client is in the foreground, the request succeeds. The result then tells that nothing changed.

Each clear writes an OSC 0 escape sequence with the default title of the client, for example `\x1b]0;host: ~\x07`. Thus, a clear is an active write of the default title. It does not only remove the override. Two clears in sequence write two identical OSC 0 sequences.

**Params**: `EmptyParams`, an empty object. Send `{}`.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| _(none)_ | — | — | — | No parameters. The request must contain `params`. herdr 0.9.5 accepts `{}`. Like 0.9.1 and 0.9.3, it also accepts an empty array `[]` and an object with unknown extra keys. herdr rejects only a missing `params`, `null`, or a scalar that is not an object or an array. |

The rejection is an `invalid_request` error, as in these examples:

- No `params`: `invalid request: missing field `params` at line 1 column 47`.
- `"params":null`: `invalid type: null, expected struct EmptyParams …`.
- `"params":5`: `invalid type: integer `5`, expected struct EmptyParams …`.

**Result**: `type` const `client_window_title`.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| type | string | yes | — | Always `client_window_title`. |
| changed | boolean | yes | — | `true` when a foreground client accepted the clear. This is also true when no caller set a title override before. `false` only when no foreground client exists. The method does not record if an override exists. |
| reason | string (enum) | yes | — | Outcome detail. One of: `set`, `cleared`, `no_foreground_client`. For this method, the effective values are `cleared` (title set back to the default) and `no_foreground_client` (no client for the action). |

**Errors**: a pass saw no error codes, except the `invalid_request` errors for the envelope. Other codes are possible. Refer to [errors.md](../errors.md#observed-error-codes).

**CLI**: API-only (no CLI subcommand).

**Example**:

```json
{"id":"1","method":"client.window_title.clear","params":{}}
{"id":"1","result":{"type":"client_window_title","changed":true,"reason":"cleared"}}
```

`Checked 2026-10-17 against herdr 0.9.5`: this stamp applies to all claims in this section. A pass sent requests to a scratch server with a TUI client attached, and with no client. These results occurred:

- With a TUI client attached, two clears in sequence both returned `changed:true` and `reason:"cleared"`.
- With no client, the result was `changed:false` and `reason:"no_foreground_client"`.
- The `params` values `[]` and `{"x":1}` gave a result. A missing `params`, `null`, and `5` gave `invalid_request`.
- The OSC 0 writes came from the pty of the TUI client.

## client.window_title.set

This method replaces the window title of the foreground client with the string that you supply. It operates only on the foreground client. When no client is in the foreground, the request succeeds. The result then tells that nothing changed.

herdr writes the title to the terminal of the client as an OSC 0 escape sequence (`\x1b]0;<title>\x07`). Since 0.9.3, the client makes the title safe before it writes it:

- The client removes control characters. For example, herdr wrote `"a\u0007b\u001bc\nd 🐧"` as `abcd 🐧`.
- The client cuts the title to 200 characters. For example, herdr wrote only the first 200 characters of an 8192-character title.

The response does not tell about these changes. herdr 0.9.1 wrote the string with no changes.

A timing race occurs when a TUI client starts and when it stops. In these short periods, the server reports `no_foreground_client`, also when a session exists:

- After a TUI client starts, the race continued for about 0.2 to 0.3 s in the 0.9.5 pass. The 0.9.1 pass saw "roughly the first seconds".
- When the client stops, the race starts almost immediately.

**Params**: `ClientWindowTitleSetParams`.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| title | string | yes | — | The window title for the foreground client. It must not be empty (refer to Errors). The wire puts no other limits on it. herdr accepts an 8192-character title, Unicode and emoji, and control characters (BEL, ESC, newline) with `changed:true`. But the client changes the text that gets to the terminal, as the section text above tells. |

**Result**: `type` const `client_window_title`.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| type | string | yes | — | Always `client_window_title`. |
| changed | boolean | yes | — | `true` if herdr applied the title. `false` if herdr could not apply it, for example when no foreground client exists. |
| reason | string (enum) | yes | — | Outcome detail. One of: `set`, `cleared`, `no_foreground_client`. For this method, the effective values are `set` (herdr applied the title) and `no_foreground_client` (no client for the action). |

**Errors**:

| code | when |
| --- | --- |
| invalid_params | `title` is empty (message: `window title is empty`). |

Other codes are possible. Refer to [errors.md](../errors.md#observed-error-codes).

**CLI**: API-only (no CLI subcommand).

**Example**:

```json
{"id":"1","method":"client.window_title.set","params":{"title":"api-docs"}}
{"id":"1","result":{"type":"client_window_title","changed":false,"reason":"no_foreground_client"}}
```

`Checked 2026-10-17 against herdr 0.9.5`: this stamp applies to all claims in this section, except the 0.9.1 behavior. These results occurred:

- The example ran with no foreground client. Thus, `changed` is `false` and `reason` is `no_foreground_client`.
- With a TUI client attached in a pty, the same request returned `changed:true` and `reason:"set"`. A pass read the OSC 0 bytes from that pty.
- With the title sent as the control-character string, the OSC payload was `abcd 🐧`. With the 8192-character title, the payload had 200 characters. Both requests returned `changed:true` and `reason:"set"`.
- Requests up to 0.203 s after the client started returned `no_foreground_client`. Requests at 0.304 s, 0.501 s, and 1.0 s returned `set`.
- After the client got SIGTERM, a request at +0 ms returned `set`. Requests from +50 ms returned `no_foreground_client`.

The 0.9.1 behavior and its timing come from earlier passes. This pass did not check them again.

## client_shell.surface.set

This method sets if the client-shell connection that sends the request holds the surface interest lease. The holder of the lease receives and controls the presentation of panes.

Only a connection from a client-shell endpoint can use this method. herdr rejects other API connections with `connection_local_only`. This method is new on the endpoint protocol. Thus, its result has a `projection_revision`, which sets an activation floor for the client-shell projection.

**Params**: `ClientShellSurfaceSetParams`.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| active | boolean | yes | — | `true` gets the surface interest lease of the client shell (receive and control the presentation of panes). `false` releases it. |

**Result**: `type` const `client_shell_surface_set`.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| type | string | yes | — | Always `client_shell_surface_set`. |
| active | boolean | yes | — | The state of the lease after the request. |
| projection_revision | integer (uint64) | yes | — | Revision floor for the client-shell projection that this acknowledgement sets. |

**Errors**:

| code | when |
| --- | --- |
| connection_local_only | The request did not come from a connection of a client-shell endpoint (message: `client_shell.surface.set is only available through a client shell endpoint`). |

Other codes are possible. Refer to [errors.md](../errors.md#observed-error-codes).

herdr checks the envelope before it checks the kind of connection. Thus, a body that is not correct reports `invalid_request`, not `connection_local_only`. This also occurs over a connection that is not an endpoint connection. These are examples:

- A missing `active`: `invalid request: missing field `active` at line 1 column 59`.
- A string `active`: `invalid type: string "yes", expected a boolean …`.

**CLI**: API-only (no CLI subcommand).

**Example**:

```json
{"id":"2","method":"client_shell.surface.set","params":{"active":true}}
{"id":"2","error":{"code":"connection_local_only","message":"client_shell.surface.set is only available through a client shell endpoint"}}
```

`Checked 2026-10-17 against herdr 0.9.5`: the `connection_local_only` rejection and the two `invalid_request` cases. A pass sent them over a plain API connection, with a TUI client attached and with no client. The kind of connection of the caller controls the result. The state of the session does not.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the success response from the client-shell endpoint, with its `active` and `projection_revision` fields. A pass could not get a response from the `herdr-client.sock` socket of a session:

- With newline JSON, the server reset the connection.
- With a 4-byte big-endian length prefix, the server reset the connection.
- With a 4-byte little-endian length prefix, the read was empty (EOF).

The 0.9.1 pass also saw a reset with a line that started with a greeting.

## command.invoke

This method runs a command that the client-shell projection shows to the client-shell connection that sends the request. `command_id` is an opaque ID that the endpoint gives in that projection. It is not a fixed enum that the caller selects.

herdr does these steps in this sequence:

1. herdr deserializes the request. A `selection` that is not correct, such as `{"bogus":1}` with no required `pane_id`, fails with `invalid_request`.
2. herdr finds the `command_id`. An unknown ID gives `command_not_found`.
3. herdr resolves `pane_id`, `tab_id`, `workspace_id`, and `selection`.

Thus, an unknown `command_id` hides any problem in those fields behind `command_not_found`.

`pane_id`, `tab_id`, and `workspace_id` are optional. They limit the command to one target. `selection` is optional. It holds selection coordinates that the client owns. Its points are `{row, col}`, not `column`. The schema tells that herdr checks them against the content revision of the pane (refer to [pane.selection.read](pane.md#paneselectionread)). No `command_id` from an endpoint was available. Thus, no pass saw herdr do this check.

**Params**: `CommandInvokeParams`.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| command_id | string | yes | — | Opaque command ID from the client-shell projection, which the endpoint gives. |
| pane_id | string \| null | no | null | Pane that limits the command, if any. |
| tab_id | string \| null | no | null | Tab that limits the command, if any. |
| workspace_id | string \| null | no | null | Workspace that limits the command, if any. |
| selection | PaneSelectionReadParams \| null | no | null | Selection coordinates that the client owns. herdr checks them against the content revision of the pane (refer to [pane.selection.read](pane.md#paneselectionread)). |

**Result**: `type` const `ok`.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| type | string | yes | — | Always `ok`. |

**Errors**:

| code | when |
| --- | --- |
| command_not_found | The `command_id` is not valid now. A pass saw this code for each unknown ID that it sent. This occurred also on a server with a valid custom-command manifest that it loaded just before. The message names a stale manifest as the cause, also when the manifest is not stale (message: `custom command manifest is stale; reload configuration`). |

Other codes are possible. Refer to [errors.md](../errors.md#observed-error-codes).

**CLI**: API-only (no CLI subcommand).

**Example**:

```json
{"id":"5","method":"command.invoke","params":{"command_id":"demo.command"}}
{"id":"5","error":{"code":"command_not_found","message":"custom command manifest is stale; reload configuration"}}
```

`Checked 2026-10-17 against herdr 0.9.5`: the `command_not_found` rejection, the masked `pane_id` and `workspace_id`, and the `selection` that is not correct. A pass loaded one `[[keys.command]]` entry and ran `herdr server reload-config`, which returned `status:"applied"`. Then the IDs `custom:0`, `custom.0`, `command:0`, `keys.command.0`, `prefix+alt+g`, `0`, and `true` each gave `command_not_found`.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the `{"type":"ok"}` success response, and the claim that herdr checks `selection` against the content revision. The public API socket gives no `command_id` from an endpoint.

## notification.show

This method shows a desktop or toast notification through the foreground client. herdr does not guarantee delivery. It can suppress a notification in these conditions:

- The client has notifications disabled.
- The rate limit stops the notification.
- The client is busy.
- No foreground client exists.

The `shown` flag and the `reason` enum in the result give the outcome. A suppressed notification is a successful request, not an error.

`shown:true` tells that herdr accepted the notification for delivery. It does not tell that a person saw a notification. A foreground client can have its toast delivery off, for example with `[ui.toast] delivery = "off"`. That client also reports `shown:true` and `reason:"shown"`, but it shows nothing. The caller cannot find this from the response.

**Params**: `NotificationShowParams`.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| title | string | yes | — | Notification title (first line). It must not be empty (refer to Errors). |
| body | string \| null | no | null | Optional body text of the notification. |
| position | ToastHerdrPosition \| null | no | null | Optional corner of the screen for the toast. One of: `top-left`, `top-right`, `bottom-left`, `bottom-right`. `null` uses the default of the client. |
| sound | NotificationShowSound (enum) | no | — | Sound to play. One of: `none`, `done`, `request`. This field does not accept an explicit `null`, but `body` and `position` accept it. herdr rejects a `null` sound with `invalid_request`. To get no sound value, do not include the field. Do not send `null`. |

**Result**: `type` const `notification_show`.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| type | string | yes | — | Always `notification_show`. |
| shown | boolean | yes | — | `true` if herdr accepted the notification for delivery, to a foreground client or through the fallback path with no client. `false` if no delivery path was available. It does not guarantee that a person saw the notification. Refer to `disabled` and the text above. |
| reason | string (enum) | yes | — | Outcome detail. Refer to the reason table below. |

| reason | meaning |
| --- | --- |
| `shown` | herdr accepted the notification for delivery. |
| `disabled` | No foreground client exists, and the config turns off the fallback delivery path with no client. |
| `rate_limited` | herdr delivered a notification in about the last second. |
| `no_foreground_client` | No foreground client exists, and the fallback path is not disabled. |
| `busy` | The client does not accept notifications now. The schema defines this value, but no pass saw it on a server. |

These rules apply to the `reason` values:

- `disabled` occurs only when no foreground client exists. A foreground client with its delivery off reports `shown`.
- The default config has `[ui.toast] delivery = "off"`. Thus, a server on the default config with no client answers `disabled`, not `no_foreground_client`.
- The rate limit permits about one notification each second for each server. A second call in about one second after the first gives `rate_limited`.
- On 0.9.5, a call 0.98 s after a delivered notification gave `rate_limited`. A call at 1.02 s gave `shown` again.
- In a burst of calls with no delay, herdr drops each call after the first.
- The rate limit counts only delivered notifications. Two `no_foreground_client` calls with no delay both returned `no_foreground_client`, not `rate_limited`.
- No pass saw `busy`. A pass tried with no client, an idle client, a client that was starting, and a client with an open popup overlay.

**Errors**:

| code | when |
| --- | --- |
| invalid_params | `title` is empty (message: `notification title is empty`). |

Other codes are possible. Refer to [errors.md](../errors.md#observed-error-codes).

**CLI**: `herdr notification show <TITLE> [--body <TEXT>] [--position <top-left|top-right|bottom-left|bottom-right>] [--sound <none|done|request>]`. The CLI writes the result keys in alphabetical order (`reason`, `shown`, `type`). The wire example below has a different sequence. [cli-mapping.md](../cli-mapping.md#cli-conventions) gives the CLI conventions. The CLI rejects these errors before it connects, with exit status 2. The server gets no request for them:

- A missing `TITLE`.
- A `--position` value that is not in the enum.

An empty `TITLE` (`""`) gets to the server. The CLI then writes the `invalid_params` error and exits with status 1.

**Example**:

```json
{"id":"cli:notification:show","method":"notification.show","params":{"title":"demo"}}
{"id":"cli:notification:show","result":{"type":"notification_show","shown":false,"reason":"disabled"}}
```

`Checked 2026-10-17 against herdr 0.9.5`: this stamp applies to all claims in this section. These results occurred:

- With no client on the default config, the result was `disabled`.
- With a TUI client attached, the result was `shown`. That client had delivery `"off"`.
- With `[ui.toast] delivery = "herdr"` and no client, the result was `no_foreground_client`.
- The rate-limit timings, the `invalid_params` case, and the `invalid_request` case occurred as the text tells.
- With a popup open, the result was `shown`.
- No call gave `busy`.
- The CLI wrote the keys in alphabetical order, and it rejected the errors before it connected, as the CLI text tells.

`disabled` needs no foreground client and delivery `"off"`. The same request with a TUI client attached returns `shown:true` and `reason:"shown"`. This is also true when the CLI sends the request. A CLI request with no TUI client attached returns `disabled`.

## popup.close

This method closes the popup overlay of the client, if a popup is open. If no popup is open, the request fails with `popup_not_open`. It does not succeed with no effect.

A popup overlay does not block the other `ui` methods. While a popup is open, `notification.show`, `client.window_title.set`, and `client.window_title.clear` operate normally.

**Params**: `EmptyParams`, an empty object. Send `{}`.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| _(none)_ | — | — | — | No parameters. The request must contain `params`. herdr 0.9.5 accepts `{}`. Like 0.9.1 and 0.9.3, it also accepts an empty array `[]` and an object with unknown extra keys. herdr rejects only a missing `params`, `null`, or a scalar that is not an object or an array. |

**Result**: `type` const `ok`.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| type | string | yes | — | Always `ok`. It tells that herdr closed the popup. |

**Errors**:

| code | when |
| --- | --- |
| popup_not_open | No popup is open, so herdr has nothing to close (message: `no popup is open`). |

Other codes are possible. Refer to [errors.md](../errors.md#observed-error-codes).

**CLI**: API-only (no CLI subcommand).

**Example**:

```json
{"id":"7","method":"popup.close","params":{}}
{"id":"7","error":{"code":"popup_not_open","message":"no popup is open"}}
```

`Checked 2026-10-17 against herdr 0.9.5`: this stamp applies to these claims:

- The example shows the case with no open popup. The result was the same with a client and with no client, and with the `params` `{}`, `[]`, and `{"x":1}`.
- A pass opened a popup with [plugin.pane.open](plugin.md#pluginpaneopen) and `placement: "popup"`. Then `popup.close` returned `{"type":"ok"}`. A second call immediately after it returned `popup_not_open`.
- While the popup was open, a second `plugin.pane.open` returned `ui_busy` (`a popup pane is already open`).
- While the popup was open, `notification.show` returned `shown`. `client.window_title.set` returned `set`, and `client.window_title.clear` returned `cleared`.

A 0.9.3 pass saw this behavior: a popup that a caller opens with no client attached stays open until `popup.close`. This pass did not check it on 0.9.5, because it opened the popup only with a TUI client attached. Refer to [plugin.pane.open](plugin.md#pluginpaneopen).

## product_announcement.dismiss

This method records that the caller dismissed a product announcement. The `id` and the `version` identify the announcement. The announcement can be different from the current announcement in the server queue. In that condition, the request fails with `stale_announcement`. It does not succeed silently.

**Params**: `ProductAnnouncementDismissParams`.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| id | string | yes | — | ID of the announcement to dismiss. The envelope accepts an empty string, and herdr treats it as not current. herdr does not reject it with `invalid_params`, as it rejects an empty `title` of `notification.show`. |
| version | string | yes | — | Version of the announcement to dismiss. herdr treats an empty string as it treats an empty `id`. |

**Result**: `type` const `ok`.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| type | string | yes | — | Always `ok`. |

**Errors**:

| code | when |
| --- | --- |
| stale_announcement | `id` or `version` does not match the current product announcement (message: `the product announcement is no longer current`). |

Other codes are possible. Refer to [errors.md](../errors.md#observed-error-codes).

**CLI**: API-only (no CLI subcommand).

**Example**:

```json
{"id":"3","method":"product_announcement.dismiss","params":{"id":"demo-announcement","version":"0.9.5"}}
{"id":"3","error":{"code":"stale_announcement","message":"the product announcement is no longer current"}}
```

`Checked 2026-10-17 against herdr 0.9.5`: the `stale_announcement` error, which includes the case with empty strings.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the `{"type":"ok"}` success response. No scratch server had a current product announcement in its queue. A pass found no API or config mechanism that adds one. The isolated config directory had no announcement file.

## release_notes.dismiss

This method records that the caller dismissed the release notes for `version`. The `version` can be different from the release notes that the server thinks are current. In that condition, the request fails with `stale_release_notes`.

These rules apply to the current release notes:

- herdr reads the current version from the `version` field of `release-notes.json` in the config directory. It does not use the version of the binary that runs.
- The same `version` string returns `ok` from a server whose config directory has release notes for that version. It returns `stale_release_notes` from an identical server whose config directory does not have that file.
- The server reads `release-notes.json` when it starts, not for each call. If you write the file while the server runs, the server does not see the change. `herdr server reload-config` does not read the file again.
- Matching is an exact string comparison. herdr does not remove white space, does not accept a `v` prefix, and does not normalize semver.

A successful dismissal writes to the config directory, which all sessions of that config root share. The first successful dismissal writes `"show_on_startup": false` into `release-notes.json`. A second successful dismissal does not write the file again. A stale call does not write. When the server starts, it does not write the file.

**Params**: `ReleaseNotesDismissParams`.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| version | string | yes | — | Version of the release notes to dismiss. herdr compares it as an exact string with `release-notes.json`. The envelope accepts an empty string, and herdr treats it as not current. herdr does not reject it with `invalid_params`. |

**Result**: `type` const `ok`.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| type | string | yes | — | Always `ok`. A second successful dismissal also returns `ok` (idempotent). |

**Errors**:

| code | when |
| --- | --- |
| stale_release_notes | `version` does not match the release notes that the server thinks are current (message: `the release notes are no longer current`). |

Other codes are possible. Refer to [errors.md](../errors.md#observed-error-codes).

**CLI**: API-only (no CLI subcommand).

**Example**:

```json
{"id":"4","method":"release_notes.dismiss","params":{"version":"0.9.5"}}
{"id":"4","result":{"type":"ok"}}
```

`Checked 2026-10-17 against herdr 0.9.5`: this stamp applies to all claims in this section. A pass used a scratch server with an isolated config directory. These results occurred:

- With no `release-notes.json`, the version `"0.9.5"` gave `stale_release_notes`.
- With a file that named `0.9.5`, the version `"0.9.5"` gave `ok`. The versions `" 0.9.5"`, `"0.9.5 "`, `"v0.9.5"`, `"0.9.3"`, and `""` gave `stale_release_notes`.
- Two `"0.9.5"` calls gave `ok` two times.
- A server that started with `"show_on_startup": false` already in the file also answered `ok`. Thus, the dismissed state does not make the notes stale.
- A pass wrote the file into the config directory of a running server and ran `herdr server reload-config`. `"0.9.5"` then still gave `stale_release_notes`. A server that started after the write gave `ok`.
- The first successful dismissal changed the file and its mtime, as this capture shows. The second dismissal did not change the mtime.

```text
before:  {"version": "0.9.5", "body": "demo notes"}
dismiss "0.9.5" -> {"type":"ok"}
after:   {"version": "0.9.5", "body": "demo notes", "show_on_startup": false}
```
