# herdr API: wire protocol

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Index: [README.md](README.md). IDs: [addressing.md](addressing.md). Access model: [environment.md](environment.md).

A herdr server accepts requests on a Unix domain socket. The socket uses newline-delimited JSON (NDJSON). A client opens the socket, writes one request line, and reads one response line. Then the server closes the connection.

Two methods keep the connection open after the response:

- `events.subscribe` keeps the connection open for the event lines that the server pushes.
- [`server.ssh_agent.register`](api/server.md#serverssh_agentregister) (new in 0.9.3) keeps the connection open because its registration lasts while the connection stays open.

The socket has no HTTP, no framing header, and no multiplexing. Each connection carries one request and its response, one subscription, or one registration.

This page gives the transport, the framing, the four envelope shapes, and the connection rules. It also gives the `ping` handshake, the check of protocol and version, and the subscription connections. For the params and the results of each method, see the `api/*.md` files.

## Transport

- The server listens on a Unix domain socket of the stream type. Each managed pane gets the socket path in `$HERDR_SOCKET_PATH`. For the socket path of the default session and of a named session, see [environment.md](environment.md#socket-path-and-permissions). That section also gives the `socket:` line of `herdr status server`.
- herdr uses no token, no session cookie, and no API key. For the socket mode and the access model, see [environment.md](environment.md#no-token-or-api-key-authentication). On Windows, the `allow_unelevated_clients` setting also changes the access to an elevated server. The same section gives this setting.
- The `herdr` CLI is a thin client of the same socket. When the server returns an error, the CLI writes the error body to stderr and exits with status 1.
- The CLI serializes the error body again. Thus its output is not byte-identical to the socket response. The key order is different: `error` comes before `id`. The `id` is a token of the CLI, not the `id` of the caller.
- For example, `herdr workspace get w999` exits with status 1 and writes this line:

  ```json
  {"error":{"code":"workspace_not_found","message":"workspace w999 not found"},"id":"cli:workspace:get"}
  ```

  The same request on the socket, with the `id` value `cmp`, returns this line:

  ```json
  {"id":"cmp","error":{"code":"workspace_not_found","message":"workspace w999 not found"}}
  ```

- A CLI syntax error, such as a bad flag or a bad argument, does not go to the server. The CLI then exits with status 2. `herdr workspace get` without an argument writes `usage: herdr workspace get <workspace_id>`. An unknown subcommand writes the command list of the group.
- An unknown `--flag` in an argument position is not always a syntax error. For example, `herdr workspace get --bogus` sends `--bogus` as the ID. The command then fails with `workspace_not_found` and exit status 1. For all CLI conventions, see [cli-mapping.md](cli-mapping.md#cli-conventions).
- The byte encoding is UTF-8. Requests and responses are compact JSON objects. The server does not require pretty-printed JSON. It writes compact JSON on one line.

Stamps for this section:

- `Checked 2026-10-17 against herdr 0.9.5`: the stream socket, the CLI error output, the exit statuses, and the three syntax cases. Also checked: the UTF-8 encoding, with a request `id` that has non-ASCII characters, and the compact single-line responses.
- `Documented by herdr --default-config (herdr 0.9.5); not live-checked:` `allow_unelevated_clients` is a Windows setting, and the host of the pass runs Linux.

## Framing

- Both directions use NDJSON. Each line holds one complete JSON object and ends with one LF (`\n`).
- A request is one line that the client writes to the socket. A response is one line that the client reads back. The protocol has no length prefix and no envelope outside the JSON object.
- On a subscription connection, the client reads more than one line over time. The first line is one acknowledgement. Zero or more event lines follow it. Each event line is one JSON object that ends with LF.
- A client must read a full line, up to and including the LF, before it parses the line.
- The server writes each object and then an LF. At the protocol level, the server does not split an object across writes.
- The server accepts a request line that ends with CRLF (`\r\n`).
- A client can write a request without an LF and then close its write side (`shutdown(SHUT_WR)`). The server then sends zero response bytes.

The server does not answer malformed framing with the `invalid_request` envelope of [Error envelope](#error-envelope). It closes or resets the connection without a response:

- **Long line.** The server answers a request line of 1,048,577 bytes, which includes the LF. It does not answer a line of 1,048,578 bytes. The client gets zero response bytes, and the server resets the connection.
- **Empty first line.** A first line that is empty or has only whitespace (`\n`, `\r\n`, `   \n`) closes the connection with zero response bytes.
- **Blank line before a request.** A blank line and then a valid request on the same connection use the one request of that connection. The server does not answer the valid line. The client gets zero response bytes, and the server resets the connection.
- **Invalid UTF-8.** Invalid UTF-8 bytes at any position in the request line give zero response bytes and an immediate close.

Stamps for this section:

- `Checked 2026-10-17 against herdr 0.9.5`: NDJSON in the two directions, one LF at the end of each response, and the subscription lines. Also checked: the CRLF case, the case without an LF, and the four cases of malformed framing.
- `Checked 2026-10-17 against herdr 0.9.5`, with a limit: a client cannot see the write calls of the server. Each response that the pass read was one complete line.

## Request envelope

Each request is a JSON object with these top-level fields:

| field | type | required | meaning |
| --- | --- | --- | --- |
| `id` | string | yes | A correlation token that the client selects. The server copies it unchanged into the `id` of the response. herdr has one known exception: a `pane_not_found` error from `events.wait` adds the fixed suffix `:sub:0:probe` (see [Subscription connections](#subscription-connections)). The server rejects a request without `id` with `invalid_request`. The value can be any string. The client controls if the values are unique. |
| `method` | string | yes | The method name, such as `ping`, `workspace.list`, `pane.split`, `agent.prompt`, or `events.subscribe`. The name must be one of the known methods. The server rejects an unknown name with `invalid_request`, and the message lists each accepted method. For removed methods, see [Error envelope](#error-envelope). The server rejects a request without `method` with `invalid_request` (`missing field \`method\``). |
| `params` | object | yes | The parameters of the method. A method without arguments also requires `params`. For such a method, send `{}`. The schema types `PingParams` and `EmptyParams` are empty objects. The server rejects a request without `params` with `invalid_request` (`missing field \`params\``). The schema variant of each method sets the shape of `params`. The `api/*.md` files give each shape. The server ignores unknown extra keys at the top level and in `params`. For example, `{"id":"a10","method":"ping","params":{},"extra":1}` and `{"id":"a11","method":"ping","params":{"x":1}}` both return `pong`. |

Schema note: the top-level `Request` schema is an object that requires `id`, with the type string. It also has a `oneOf` with one variant for each method. Each variant requires `method` and `params`. Thus the schema requires all three fields. The server also rejects a request without one of the three fields.

```json
{"id":"r8","method":"agent.list","params":{}}
```

Stamps for this section:

- `Checked 2026-10-17 against herdr 0.9.5`: the three fields, the rejection of a request without each field, and the rejection of an unknown method. Also checked: the extra keys, and the example request, which a scratch server accepted.
- `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the schema note.

## Success envelope

A successful call returns one line:

| field | type | required | meaning |
| --- | --- | --- | --- |
| `id` | string | yes | The `id` of the request, copied unchanged. |
| `result` | object | yes | The result body. It always has a `type` string that names the result variant. The other fields depend on that type. |

The `result` object is the `ResponseResult` `oneOf` in the schema. A `type` const tags each variant. These are some of the tags:

- `pong`, `ok`
- `session_snapshot`
- `workspace_info`, `workspace_list`, `workspace_created`
- `tab_list`
- `pane_info`, `pane_current`, `pane_list`, `pane_move`
- `agent_info`, `agent_list`

A method without a useful payload returns `{"type":"ok"}`. `pane.report_agent` is an example. The `api/*.md` files give the result fields of each method. [data-model.md](data-model.md) defines the shared entities one time, such as `WorkspaceInfo`, `TabInfo`, `PaneInfo`, and `AgentInfo`.

```json
{"id":"r8","result":{"type":"agent_list","agents":[]}}
```

Stamps for this section:

- `Checked 2026-10-17 against herdr 0.9.5`: the two fields, and the `type` field in each success response of the pass. Also checked: the `{"type":"ok"}` result of `pane.report_agent`, and the example, which a new scratch server returned byte for byte.
- `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the list of tags. Each tag is in the 63 variants of `ResponseResult.oneOf`.

## Error envelope

A failed call returns one line. That line has an `error` object in the position of `result`:

| field | type | required | meaning |
| --- | --- | --- | --- |
| `id` | string | yes | The `id` of the request, copied unchanged. The server also copies it for most parse and validation failures, such as an unknown method, a missing `method` or `params`, or bad params. The server returns `"id": ""` (an empty string) only when it cannot get one string `id` from the line. This occurs when the line is not valid JSON. It also occurs when `id` is missing, is not a string, or occurs two times. |
| `error` | object | yes | The error body (`ErrorBody`). |
| `error.code` | string | yes | A stable code in `snake_case` for programs, such as `invalid_request`, `agent_blocked`, `agent_not_ready`, `agent_prompt_stalled`, `timeout`, or `unsupported_event_wait_match`. Use this field for decisions. For the codes, see [errors.md](errors.md). |
| `error.message` | string | yes | Text for persons. It is not stable. Do not parse it. |

All parse and validation failures use the code `invalid_request`. These captures show three examples:

```json
{"request":{"method":"ping","params":{}},"response":{"id":"","error":{"code":"invalid_request","message":"invalid request: missing field `id` at line 1 column 29"}}}
```

```json
{"request":{"id":"r9","method":"bogus.method","params":{}},"response":{"id":"r9","error":{"code":"invalid_request","message":"invalid request: unknown variant `bogus.method`, expected one of `ping`, `server.stop`, `server.live_handoff`, `server.reload_config`, … `plugin.pane.close` at line 1 column 47"}}}
```

```json
{"request":{"id":"r11","method":"workspace.get","params":{"nope":true}},"response":{"id":"r11","error":{"code":"invalid_request","message":"invalid request: missing field `workspace_id` at line 1 column 60"}}}
```

The column in these messages is a byte offset in the request line, as the client sent it. Thus the column changes with the key order and the spaces. Do not parse it. These cases return `"id": ""`:

```json
{"request":"{\"id\":\"a5\",\"method\":\"ping\",\"params\":{}","response":{"id":"","error":{"code":"invalid_request","message":"invalid request: EOF while parsing an object at line 1 column 38"}}}
{"request":{"id":7,"method":"ping","params":{}},"response":{"id":"","error":{"code":"invalid_request","message":"invalid request: invalid type: integer `7`, expected a string at line 1 column 7"}}}
{"request":"{\"id\":\"a12\",\"id\":\"a13\",\"method\":\"ping\",\"params\":{}}","response":{"id":"","error":{"code":"invalid_request","message":"invalid request: duplicate field `id` at line 1 column 16"}}}
```

**Changed from 0.9.1.** On 0.9.1, each parse failure returned `"id": ""`, also for an unknown method or bad params. Since 0.9.3, a client can match a parse error to its request by `id`. This works when the line is valid JSON with exactly one string `id`. The client must still accept `""` for the cases above.

The unknown-variant message lists 102 methods. The first three are `ping`, `server.stop`, and `server.live_handoff`. This list is the same set as the 102 methods in the request `oneOf` of `raw/schema.json`. 0.9.1 also accepted the undocumented method `pane.graphics.stream`, but that method is not in the list now.

A call to a removed graphics method, such as `pane.graphics.set` or `pane.graphics.stream`, does not return `invalid_request`. It returns the separate code `unknown_method`, and the server copies the `id`:

```json
{"id":"g1","error":{"code":"unknown_method","message":"unknown method: pane.graphics.set"}}
```

A name that herdr never had, such as `nope.nope`, still returns `invalid_request` with the unknown-variant message.

Stamps for this section:

- `Checked 2026-10-17 against herdr 0.9.5`: the `id` rules, the three `invalid_request` captures, and the byte offsets. Also checked: the three `"id": ""` captures, the list of 102 methods, the `unknown_method` code for the two graphics methods, and `nope.nope`.
- `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: `ErrorBody` requires `code` and `message`. The codes `agent_blocked`, `agent_not_ready`, `agent_prompt_stalled`, and `timeout` in the `error.code` row come from [errors.md](errors.md). The pass did not cause them on this page.

## One request per connection

The server answers exactly one request on each connection. Then it closes the connection. The two exceptions, `events.subscribe` and [`server.ssh_agent.register`](api/server.md#serverssh_agentregister), follow below. The server does not read or answer a second request on the same connection.

In this capture, a client wrote two requests, one after the other, on one connection:

```json
{
  "first":  [ {"id":"m1","result":{"type":"pong","version":"0.9.5","protocol":22,"capabilities":{"live_handoff":true,"detached_server_daemon":false,"endpoint_protocol_generation":1,"surface_interest":true,"health_check":true,"ssh_agent_registration":true}}} ],
  "second": null
}
```

The server answered the first `ping`. Then it reset the connection and did not answer the second line (`m2`). A client that must send N calls must open N connections. The CLI does this, with one connection for each run.

Two exceptions exist:

- `events.subscribe` accepts no second line (see [Subscription connections](#subscription-connections)). A second line on a subscribe connection resets the connection. The next read of the client fails with a connection reset, and the client loses the subscription.
- [`server.ssh_agent.register`](api/server.md#serverssh_agentregister) returns `{"type":"ok"}` and keeps the connection open. The server sends nothing more on it. The registration lasts until the client closes that connection. Thus a caller keeps the connection open while it needs the registration. See [SSH agent registration connections](#ssh-agent-registration-connections).

Stamps for this section:

- `Checked 2026-10-17 against herdr 0.9.5`: the capture, the reset after the first response, and the reset of a subscribe connection after a second line. Also checked: the registration connection, which stays open.

## Ping/pong handshake

`ping` is the request that checks if the server is live and what it can do. Its params are an empty object. The response is a `pong` result with the version, the protocol number, and the capability flags of the server:

| result field | type | required | meaning |
| --- | --- | --- | --- |
| `type` | const `"pong"` | yes | The result discriminant. |
| `version` | string | yes | The server version, such as `"0.9.5"`. |
| `protocol` | integer (uint32) | yes | The wire protocol number, such as `22`. Compare it with the protocol of your client build. |
| `capabilities` | object \| null | no (default `null`) | A `ServerCapabilities` object. The next table gives its fields. |

| capability field | type | required | meaning |
| --- | --- | --- | --- |
| `live_handoff` | boolean | yes | The server can move its running session to a new process across an update or an attach. |
| `detached_server_daemon` | boolean | no (default `false`) | The server runs as a detached daemon. |
| `health_check` | boolean | no (default `false`) | The server supports health checks of its endpoint. |
| `surface_interest` | boolean | no (default `false`) | The server supports explicit surface interest from the client shell. |
| `endpoint_protocol_generation` | integer (uint32) \| null | no (no default) | The stable endpoint generation that this server supports. The client owns this generation value. |
| `ssh_agent_registration` | boolean | no (default `false`) | New in the 0.9.3 schema. The server supports `server.ssh_agent.register` on the local JSON API. Each registration belongs to one connection. |

This exchange shows a request and its response:

```json
{"id":"m1","method":"ping","params":{}}
{"id":"m1","result":{"type":"pong","version":"0.9.5","protocol":22,"capabilities":{"live_handoff":true,"detached_server_daemon":false,"endpoint_protocol_generation":1,"surface_interest":true,"health_check":true,"ssh_agent_registration":true}}}
```

On the server of a live default session, `herdr status --json` gave the same capability set in `server.capabilities`. The only difference was `detached_server_daemon: true`.

Send `ping` as the first call of a client. The response tells the client that the server answers. The client can also read `protocol`, `version`, and `capabilities` before it uses behavior of one version.

Stamps for this section:

- `Checked 2026-10-17 against herdr 0.9.5`: the fields of the `pong` result and the exchange, on a scratch server. Also checked: the capabilities of the live server, through a read-only `herdr status --json`, not through a `ping`.
- `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the defaults of the capability fields and their meanings.

## Protocol and version negotiation

herdr does not negotiate the protocol on a connection. A client reads the protocol number and the version of the server. Then the client decides if it is compatible.

- The header of the schema document gives `protocol` and `schema_version`. This schema has protocol 22 and schema_version 1.
- Each `pong` result gives `protocol` and `version`. The `snapshot` object of the `session.snapshot` result also gives them. Thus a client can check the compatibility from the two results.
- `herdr status` writes plain text with three blocks: `client:`, `server:`, and `update:`. The `server:` block has the lines `endpoint_compatible` and `private_protocol_compatible`. The `update:` block has the lines `restart_needed` and `server_binary_stale`. This capture comes from a scratch server:

  ```text
  client:
    version: 0.9.5
    channel: stable
    protocol: 22
    endpoint_protocol_generation: 1

  server:
    status: running
    version: 0.9.5
    endpoint_compatible: yes
    private_protocol: 22
    private_protocol_compatible: yes
    socket: /home/user/.config/herdr/sessions/demo/herdr.sock

  update:
    restart_needed: no
    server_binary_stale: no
  ```

- `herdr status --json` gives `client.protocol` and `server.protocol`. It also gives `server.compatible`, `server.restart_needed`, and `server.server_binary_stale`. These fields are in nested objects, not at the top level. `update.restart_needed` and `update.server_binary_stale` repeat the last two values.
- When the client and the server both use protocol 22, the text shows `private_protocol_compatible: yes`. The JSON then has `server.compatible: true`.
- `herdr status --json` has more fields, such as `client.channel`, `client.endpoint_protocol_generation`, and `client.endpoint_capabilities`. Other examples are `server.capabilities`, `server.socket`, `server.session`, and `client.session`. When `HERDR_SOCKET_PATH` pointed to the socket of a named session, `server.session` and `client.session` were `null`.
- The live-handoff request is the one place where a client gives an expected protocol or version. `server.live_handoff` takes `ServerLiveHandoffParams`:

  | field | type | required | meaning |
  | --- | --- | --- | --- |
  | `expected_protocol` | integer (uint32) \| null | no | The protocol number of the caller build. |
  | `expected_version` | string \| null | no | The version string that the caller expects. |
  | `import_exe` | string \| null | no | The path to the executable that gets the running session. |

- All three fields are optional. The server accepts a request without the three fields and does a real handoff:

  ```json
  {"id":"lh-empty","method":"server.live_handoff","params":{}}
  {"id":"lh-empty","result":{"type":"ok"}}
  {"id":"post-lh-ping","method":"ping","params":{}}
  {"id":"post-lh-ping","result":{"type":"pong","version":"0.9.5","protocol":22,"capabilities":{"live_handoff":true,"detached_server_daemon":true,"endpoint_protocol_generation":1,"surface_interest":true,"health_check":true,"ssh_agent_registration":true}}}
  ```

- Without `import_exe`, herdr documents no rejection error for `expected_protocol`. In a matching call, `expected_protocol` is the protocol of the server (22), and the call has no `import_exe`. The call returns `{"type":"ok"}` and does a real handoff to the binary of the server.
- After a real handoff, the first server process exits. A new process, `herdr server --handoff-import <session dir>/herdr-handoff-<old pid>.sock <token>`, then uses the same socket path. After this, `ping` gives `detached_server_daemon: true`.
- The same call with a mismatched `expected_protocol`, such as `1`, fails with `handoff_failed` ("handoff stream closed while reading line"). The first process continues to run, and `detached_server_daemon` stays `false`:

  ```json
  {"id":"lh-mismatch","method":"server.live_handoff","params":{"expected_protocol":1}}
  {"id":"lh-mismatch","error":{"code":"handoff_failed","message":"handoff stream closed while reading line"}}
  {"id":"post-lh-ping","method":"ping","params":{}}
  {"id":"post-lh-ping","result":{"type":"pong","version":"0.9.5","protocol":22,"capabilities":{"live_handoff":true,"detached_server_daemon":false,"endpoint_protocol_generation":1,"surface_interest":true,"health_check":true,"ssh_agent_registration":true}}}
  ```

  ```json
  {"id":"lh-match","method":"server.live_handoff","params":{"expected_protocol":22}}
  {"id":"lh-match","result":{"type":"ok"}}
  {"id":"post-lh-ping","method":"ping","params":{}}
  {"id":"post-lh-ping","result":{"type":"pong","version":"0.9.5","protocol":22,"capabilities":{"live_handoff":true,"detached_server_daemon":true,"endpoint_protocol_generation":1,"surface_interest":true,"health_check":true,"ssh_agent_registration":true}}}
  ```

- A handoff has two results for a caller. First, the PID of the server changes. Second, the command line of the new process does not contain `--session <name>`. It contains the full path of the executable, such as `/home/user/.local/bin/herdr`.
- Thus a cleanup that finds a scratch server with `pgrep -f "session <name>"` does not find the new process. Find it by its socket, with `ss -xlp | grep <session dir>`. A second handoff starts a third process, which names the PID of the second process. Each later handoff continues this chain.
- With `import_exe` set to `/bin/true`, `server.live_handoff` returned `handoff_failed` (`timed out waiting for handoff import connection`). The response came 30.0 s after the request. This occurred with a matching and with a mismatched `expected_protocol`.
- In these two calls, the first server continued to run with `detached_server_daemon: false`. A client with a read timeout of less than 30 s, such as 15 s, gets no response.
- It is not known if a handoff to a different binary does its own protocol check. A complete handoff of this type replaces the server with that binary.
- `expected_protocol` occurs only in `ServerLiveHandoffParams` in the schema. It is the only request field that carries a protocol from the caller.
- Other method calls do not send a protocol, and the envelope has no version gate. A client with a wrong protocol can send a method that the server does not know. It can also omit a field that the server requires. The server then returns `invalid_request`.
- The server ignores a field that it does not know, and it gives no error (see [Request envelope](#request-envelope)). Thus an older server drops a field of a newer protocol without an error.

Stamps for this section:

- `Checked 2026-10-17 against herdr 0.9.5`: the header of `herdr api schema`, the `protocol` and `version` fields in `pong` and in `session.snapshot`, and the `herdr status` text and JSON. Also checked: the handoff without fields, and the matching and mismatched handoffs on new scratch servers. The pass saw the process changes with `pgrep` and `ss -xlp`.
- `Checked 2026-10-17 against herdr 0.9.5`: the two `/bin/true` handoffs and their 30.0 s responses, and the one match of `expected_protocol` in `raw/schema.json`.
- `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the `ServerLiveHandoffParams` table.
- The pass did not check a handoff to a different binary. No herdr binary with a different protocol was available, and a complete handoff replaces the server.

## SSH agent registration connections

[`server.ssh_agent.register`](api/server.md#serverssh_agentregister) (new in 0.9.3) is the second exception to the rule of one request per connection. The server returns `{"type":"ok"}`. Then it keeps the connection open and writes nothing more on it.

The registration belongs to the connection. The schema says that the registration "lasts until this API connection closes". Thus the caller keeps the connection open while it needs the registration. The caller closes the connection to end the registration. Unlike a subscription, this connection carries no pushed lines.

The server does not answer a second request line on a registration connection. It resets the connection, as on a subscription connection. The registration then probably ends, because the connection closes and the schema ties the registration to the connection. The pass did not read the state of the registration separately.

Stamps for this section:

- `Checked 2026-10-17 against herdr 0.9.5`: the `ok` response, the open connection without more lines, and the reset after a second line.
- `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the end of the registration after the reset.

## Subscription connections

`events.subscribe` changes the connection into a channel for pushed lines that stays open. It is one of the two exceptions to the rule of one request per connection. The other exception is [`server.ssh_agent.register`](api/server.md#serverssh_agentregister) (see [SSH agent registration connections](#ssh-agent-registration-connections)).

The client sends a subscribe request. The server returns an acknowledgement result and keeps the connection open. Then it writes one JSON line for each matching event until the client disconnects.

The request `params` object is `EventsSubscribeParams`. Its field `subscriptions` (an array, required) is a list of `Subscription` objects. 24 of the 27 selectors are only `{"type": <event-name>}` objects. The three selectors for pane streams need more fields:

- `pane.agent_status_changed` also requires `pane_id`.
- `pane.scroll_changed` also requires `pane_id`.
- `pane.output_matched` also requires `pane_id`, `source`, and `match`.

The server rejects a selector without a required field with `invalid_request`. For example, `{"type":"pane.output_matched"}` alone fails with `missing field \`pane_id\``. The acknowledgement result type is `subscription_started`. For the selectors, see [events.md](events.md#eventssubscribe).

```json
{
  "subscribe_request": {"id":"e1","method":"events.subscribe","params":{"subscriptions":[{"type":"tab.created"},{"type":"tab.closed"}]}},
  "ack": [ {"id":"e1","result":{"type":"subscription_started"}} ],
  "pushes": [
    {"data":{"tab":{"agent_status":"unknown","focused":true,"label":"1","number":1,"pane_count":1,"tab_id":"w1:t1","workspace_id":"w1"},"type":"tab_created"},"event":"tab_created"},
    {"data":{"tab":{"agent_status":"unknown","focused":false,"label":"2","number":2,"pane_count":1,"tab_id":"w1:t2","workspace_id":"w1"},"type":"tab_created"},"event":"tab_created"}
  ]
}
```

In this capture, the client subscribed on a new scratch server. Then a `workspace.create` call with the label `build` pushed the first line. A `tab.create` call with `{"workspace_id":"w1"}` pushed the second line.

A subscription connection is different from an ordinary connection:

- The server does not close it after the first response. An open stream of event lines follows the `subscription_started` acknowledgement.
- Event lines use the event envelope, not the success envelope. They have `event` and `data`, and they have no `id`. The server starts them, and they do not match a request.
- The client ends the subscription when it closes the socket. The wire protocol has no unsubscribe method.
- After the subscription starts, only the server writes on the connection. A second request line does not add subscriptions or send other calls. The server resets the connection, and the next read gets a `ConnectionResetError`. The client then loses the subscription.

For a single wait instead of a stream, use `events.wait`. It holds its one response until a matching event or a timeout occurs. At this time, it supports only matches on the agent status of a pane. It rejects other matchers:

```json
{"request":{"id":"e3","method":"events.wait","params":{"match_event":{"event":"tab_created"},"timeout_ms":500}},"response":{"id":"e3","error":{"code":"unsupported_event_wait_match","message":"events.wait currently supports pane agent status matches"}}}
```

The `id` copy of `events.wait` is not reliable on all error paths. Its `pane_not_found` error adds the fixed suffix `:sub:0:probe` to the request `id`. In this capture, `w9:p9` is a pane that does not exist:

```json
{"id":"ZZZ","method":"events.wait","params":{"match_event":{"event":"pane_agent_status_changed","pane_id":"w9:p9","agent_status":"unknown"},"timeout_ms":1000}}
{"id":"ZZZ:sub:0:probe","error":{"code":"pane_not_found","message":"pane w9:p9 not found"}}
```

`events.subscribe` does not have this behavior. Its errors for a missing pane copy the request `id` unchanged. For these errors, see [events.md](events.md#eventssubscribe).

Stamps for this section:

- `Checked 2026-10-17 against herdr 0.9.5`: the acknowledgement, the open stream, the three required `pane_id` fields, and the capture. Also checked: the reset after a second line, the `events.wait` captures, and the `id` copy of `events.subscribe` errors.
- `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the 27 selectors, of which 24 require only `type`, and the absence of an unsubscribe method.

## Event envelope

Pushed events from `events.subscribe` are objects with these fields:

| field | type | required | meaning |
| --- | --- | --- | --- |
| `event` | string | yes | The event kind. The `SubscriptionEventKind` enum of the subscription schema lists the streaming pane events: `pane.output_matched`, `pane.agent_status_changed`, and `pane.scroll_changed`. Lifecycle events on the wire use names in snake_case, such as `tab_created`. |
| `data` | object | yes | The event payload. A lifecycle payload has its own `type` discriminant, which matches the event. For example, a `tab_created` payload has a `TabInfo`. The three streaming pane events have no `type` field in `data`. |

For the event catalog, see [events.md](events.md#event-catalog-eventkind--eventdata). The schema gives these fields for the payloads of the three streaming pane events:

| event | schema type | fields |
| --- | --- | --- |
| `pane.output_matched` | `PaneOutputMatchedEvent` | Exactly `pane_id`, `matched_line`, and `read`. |
| `pane.agent_status_changed` | `PaneAgentStatusChangedEvent` | The schema requires `pane_id`, `workspace_id`, and `agent_status`. `agent` is optional and can be null. `display_agent`, `state_labels`, and `title` are optional. |
| `pane.scroll_changed` | `PaneScrollChangedEvent` | Exactly `pane_id`, `workspace_id`, and `scroll`. |

These captures show pushes of the three streaming kinds. No push has a `type` key in `data`:

```json
{"data":{"matched_line":"[user@host project]$ echo HELLO_MARKER","pane_id":"w1:p1","read":{"format":"text","pane_id":"w1:p1","revision":0,"source":"recent_unwrapped","tab_id":"w1:t1","text":"[user@host project]$ echo HELLO_MARKER\nHELLO_MARKER\n[user@host project]$","truncated":false,"workspace_id":"w1"}},"event":"pane.output_matched"}
{"data":{"agent":"reviewer","agent_status":"idle","pane_id":"w1:p2","workspace_id":"w1"},"event":"pane.agent_status_changed"}
{"data":{"agent":"claude","agent_status":"blocked","pane_id":"w1:p4","workspace_id":"w1"},"event":"pane.agent_status_changed"}
{"data":{"pane_id":"w1:p1","scroll":{"max_offset_from_bottom":64,"offset_from_bottom":7,"viewport_rows":40},"workspace_id":"w1"},"event":"pane.scroll_changed"}
```

These calls pushed the lines:

- `pane.output_matched`: the selector had `"source":"recent"` and the substring `HELLO_MARKER`. Then `pane.send_text` sent `echo HELLO_MARKER` and a newline to the pane.
- `pane.agent_status_changed`, first line: `pane.report_agent` with the agent `reviewer` and the state `idle`.
- `pane.agent_status_changed`, second line: a real `claude` process in a scratch pane. Its trust prompt gave the status `blocked`.
- `pane.scroll_changed`: `seq 1 100` in the pane, and then `pane.scroll` with `offset_from_bottom: 7`.

No `pane.agent_status_changed` push in the pass had `display_agent`, `state_labels`, or `title`.

The server writes the keys of a pushed event line in sorted order. Thus `data` comes before `event`, and the keys in `data` are in alphabetical order. A response keeps `id` as its first key.

Event lines never have an `id`. A client can match a lifecycle event by `data.type`. The three streaming pane events have no `data.type`. Match these events only by the top-level `event` field.

Stamps for this section:

- `Checked 2026-10-17 against herdr 0.9.5`: the two fields, the four captures, the absence of `id` and `data.type` in the streaming pushes, and the key order.
- `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the `SubscriptionEventKind` enum and the payload fields in the second table.
