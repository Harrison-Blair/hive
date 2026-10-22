# herdr API: events methods

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Index: [README.md](README.md). Wire format: [protocol.md](protocol.md).

The `events` namespace changes the request and response socket into a push channel.

- `events.subscribe` registers a set of subscriptions on a connection. The connection then stays open. The server pushes one line for each matching state change.
- `events.wait` keeps one connection open until one event matches a predicate, or until a timeout ends. The method then returns that event as a usual response.

Two envelope families use this channel:

- Plain events (`EventKind` / `EventData`) have snake_case names. The [event catalog](#event-catalog-eventkind--eventdata) lists these 26 lifecycle events.
- Subscription events (`SubscriptionEventKind` / `SubscriptionEventData`) use dotted names. herdr computes these three events from the parameters of a subscription. Refer to [subscription events](#subscription-events-subscriptioneventkind--subscriptioneventdata).

The `type` of a subscription selects the family that you receive.

`Checked 2026-10-17 against herdr 0.9.5`: the stream on the open connection, and the one response of `events.wait`.
`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the count of 26 event kinds and the count of three subscription event kinds.

2 methods:

| method | purpose |
| --- | --- |
| [events.subscribe](#eventssubscribe) | Register subscriptions on the connection. Push matching events until the connection closes. |
| [events.wait](#eventswait) | Wait for one event that matches a predicate. Then return that event. |

Both methods use these reference sections:

| section | contents |
| --- | --- |
| [Event catalog](#event-catalog-eventkind--eventdata) | Each `EventKind` value and its `EventData` payload. |
| [Subscription events](#subscription-events-subscriptioneventkind--subscriptioneventdata) | The payloads that `pane.output_matched`, `pane.agent_status_changed`, and `pane.scroll_changed` push. |
| [Subscription vs EventKind diffs](#subscription-vs-eventkind-name-and-coverage-diffs) | The exact differences in names and coverage between the subscription types and the event kinds. |

## events.subscribe

`events.subscribe` registers one or more subscriptions on the current connection and starts the stream.

1. The server sends one response with `{"type":"subscription_started"}` (the ack).
2. The connection stays open. This is different from the usual rule of one request for each connection. Refer to [protocol.md](protocol.md#one-request-per-connection).
3. The server writes one JSON line for each matching event while the caller continues to read.

No method removes a subscription. Close the connection to stop the stream.

Each pushed line is an envelope with the keys `event` and `data`. The server writes the keys of a pushed line in alphabetical order. Thus, `data` comes before `event`. Responses, such as the ack, keep the key order of the schema.

- For the 24 plain subscription types, the envelope is an [EventEnvelope](#event-catalog-eventkind--eventdata). Its `event` is snake_case, and its `data` follows the catalog.
- For the three parameterized types, the envelope is a [SubscriptionEventEnvelope](#subscription-events-subscriptioneventkind--subscriptioneventdata). Its `event` uses a dotted name.

An empty `subscriptions` array is valid in the schema. The server sends the ack and then pushes no event.

`Checked 2026-10-17 against herdr 0.9.5`: the ack, the open connection, and the key order. The check also covered the two envelope families and the empty array. The schema has no method that removes a subscription.

**Params**: `EventsSubscribeParams`.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| subscriptions | array of `Subscription` | yes | — | The subscriptions to register on this connection. Each item is one variant below. You can mix the variants. |

### Subscription variants

`Subscription` is a `oneOf` with the discriminator `type`. It has 27 variants.

- 24 variants have no field other than `type`. Each of these variants pushes the plain event with the same name. Change the dot in the `type` to an underscore to get the `EventKind`. Refer to the [diffs section](#subscription-vs-eventkind-name-and-coverage-diffs).
- Three variants have parameters. Each of these variants pushes a subscription event.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the 27 variants and the name mapping in this table.
`Checked 2026-10-17 against herdr 0.9.5`: an unknown `type` gives an error message that lists the same 27 names in the same order.

| `type` const | fields (other than `type`) | delivered envelope | delivered `event` value |
| --- | --- | --- | --- |
| `workspace.created` | — | EventEnvelope | `workspace_created` |
| `workspace.updated` | — | EventEnvelope | `workspace_updated` |
| `workspace.metadata_updated` | — | EventEnvelope | `workspace_metadata_updated` |
| `workspace.renamed` | — | EventEnvelope | `workspace_renamed` |
| `workspace.moved` | — | EventEnvelope | `workspace_moved` |
| `workspace.reordered` | — | EventEnvelope | `workspace_reordered` |
| `workspace.closed` | — | EventEnvelope | `workspace_closed` |
| `workspace.focused` | — | EventEnvelope | `workspace_focused` |
| `worktree.created` | — | EventEnvelope | `worktree_created` |
| `worktree.opened` | — | EventEnvelope | `worktree_opened` |
| `worktree.removed` | — | EventEnvelope | `worktree_removed` |
| `tab.created` | — | EventEnvelope | `tab_created` |
| `tab.closed` | — | EventEnvelope | `tab_closed` |
| `tab.focused` | — | EventEnvelope | `tab_focused` |
| `tab.renamed` | — | EventEnvelope | `tab_renamed` |
| `tab.moved` | — | EventEnvelope | `tab_moved` |
| `pane.created` | — | EventEnvelope | `pane_created` |
| `pane.closed` | — | EventEnvelope | `pane_closed` |
| `pane.updated` | — | EventEnvelope | `pane_updated` |
| `pane.focused` | — | EventEnvelope | `pane_focused` |
| `pane.moved` | — | EventEnvelope | `pane_moved` |
| `pane.exited` | — | EventEnvelope | `pane_exited` |
| `pane.agent_detected` | — | EventEnvelope | `pane_agent_detected` |
| `pane.output_matched` | `pane_id`, `source`, `match`, `lines`, `strip_ansi` | SubscriptionEventEnvelope | `pane.output_matched` |
| `pane.agent_status_changed` | `pane_id`, `agent_status` | SubscriptionEventEnvelope | `pane.agent_status_changed` |
| `pane.scroll_changed` | `pane_id` | SubscriptionEventEnvelope | `pane.scroll_changed` |
| `layout.updated` | — | EventEnvelope | `layout_updated` |

Note: no `pane.output_changed` subscription exists. A subscription with this `type` gives `invalid_request` (`unknown variant`). To watch output, use the parameterized `pane.output_matched` subscription. It filters the output and reads the pane for you. herdr does not push the raw `pane_output_changed` event. Refer to the [diffs section](#subscription-vs-eventkind-name-and-coverage-diffs). `Checked 2026-10-17 against herdr 0.9.5.`

**`pane.output_matched` fields**: watch the output of a pane, and push only when a match occurs.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the field types, the required flags, and the defaults in this table. The notes below the table give the checked behavior.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| type | string | yes | — | Const `pane.output_matched`. |
| pane_id | string | yes | — | The pane to watch. |
| source | `ReadSource` enum | yes | — | The text to search: `visible`, `recent`, `recent_unwrapped`, or `detection`. |
| match | `OutputMatch` | yes | — | The match predicate. Refer to `OutputMatch` below. |
| lines | integer (uint32) \| null | no | null (default window) | The size of the read window, in rows from the bottom of the selected `source`. herdr searches for the match only in this window. The window is also the pushed `read.text`. Refer to the window notes below. |
| strip_ansi | boolean | no | `true` | Remove ANSI escapes before the match, and before herdr returns `matched_line` and `read.text`. Refer to the `strip_ansi` note below. |

Window notes for `lines`:

- A match outside the window never pushes. The subscription sends the usual ack and then stays silent.
- `lines: 0` never pushes.
- A small value can also miss a real match when the rows below the match fill the window. In one check, a shell prompt used 2 rows. With `lines: 2`, the window held only the prompt, and the subscription did not push for the output line above the prompt. With `lines: 3`, it pushed.
- The window counts screen rows, also for `recent_unwrapped`. With a prompt on 3 rows, `lines: 2` and `lines: 3` did not push, and `lines: 4` pushed. The result was the same for `visible`, `recent`, and `recent_unwrapped`.

`Checked 2026-10-17 against herdr 0.9.5`: all window notes.

**Delivery occurs one time, not continuously.** The schema does not document this behavior. A `pane.output_matched` subscription pushes exactly one event on its connection:

- If the watched text contains a match at subscribe time, the subscription pushes that match at once. This match is level-triggered and uses the existing scrollback.
- If not, the subscription pushes the first match in output that comes after the subscribe call.

After that push, the subscription never pushes again, although later output continues to match the pattern. To get continuous notifications, close the connection and subscribe again after each push.

The `matched_line` is the first matching row in the window. That row can be the echoed command line, such as `$ echo MARKA`, and not the output row.

`Checked 2026-10-17 against herdr 0.9.5`: one push for each subscription with the sources `recent`, `visible`, `detection`, and `recent_unwrapped`, the immediate push for an existing match, and the `matched_line` row.

Two more behaviors that the schema does not document:

- With `source: "recent"`, the pushed `read.source` is `recent_unwrapped`, never `recent`. A plain `pane.read` call with `source: "recent"` correctly returns `recent`. The other three sources come back unchanged in the push.
- `source: "visible"` depends on the current scroll offset of the pane. Sometimes the viewport is not at the bottom when new output arrives. Then a match on that new output does not push, because the output is outside the visible window. In one check, the pane was 50 rows above the bottom. A `visible` subscription stayed silent, and a `recent` subscription on the same output pushed.

`Checked 2026-10-17 against herdr 0.9.5`: both behaviors.

`OutputMatch` is a `oneOf`:

| `type` | fields | meaning |
| --- | --- | --- |
| `substring` | `value`: string | Match when `value` occurs as a substring. |
| `regex` | `value`: string | Match when the regex `value` matches. |

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: this table. `Checked 2026-10-17 against herdr 0.9.5`: a substring and a regex each matched as the table gives.

Note: `strip_ansi: false` had no effect that a check could see.

- A `printf` of a red `MARKE` put real escape bytes in the scrollback. A `pane.read` with `format: "ansi"` returned them. Its text contained `\x1b[38;5;1mMARKE\x1b[0m`.
- Subscriptions with `strip_ansi: true` and with `strip_ansi: false` both pushed `matched_line: "MARKE"` with `read.format: "text"`.
- With `strip_ansi: false`, a pattern with the escape never matched. The check used the substrings `\x1b[38;5;1mMARKF`, `38;5;1mMARKF`, and `[31mMARKF`, and the regex `\x1b\[3`.

Thus, expect the match to use plain text always. `Checked 2026-10-17 against herdr 0.9.5.`

**`pane.agent_status_changed` fields**: push when the agent status of a pane changes.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| type | string | yes | — | Const `pane.agent_status_changed`. |
| pane_id | string | yes | — | The pane to watch. |
| agent_status | `AgentStatus` enum \| null | no | null | If you set it, herdr pushes only the changes **into** this status. If it is null or not given, herdr pushes on each change. Values: `idle`, `working`, `blocked`, `done`, `unknown`. |

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the field types and defaults. `Checked 2026-10-17 against herdr 0.9.5`: a subscription without a filter pushed `working`, then `done`. A filter of `"done"` pushed only the change into `done`.

**`pane.scroll_changed` fields**: push when the scroll position of a pane changes.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| type | string | yes | — | Const `pane.scroll_changed`. |
| pane_id | string | yes | — | The pane to watch. |

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: this table.

**Result** (ack): `type` const `subscription_started`. The server sends it one time, immediately, before any pushed event.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| type | string | yes | — | Always `subscription_started`. |

**Pushed lines**: after the ack, the server pushes zero or more lines. These lines are not responses. They have no `id` and no `result`. Each line has only the keys `data` and `event`, in this shape: `{"data":<EventData|SubscriptionEventData>,"event":<EventKind|SubscriptionEventKind>}`. Refer to the [event catalog](#event-catalog-eventkind--eventdata) and to [subscription events](#subscription-events-subscriptioneventkind--subscriptioneventdata).

`Checked 2026-10-17 against herdr 0.9.5`: the ack and the keys of each pushed line.

**Errors**: [errors.md](errors.md) gives the error codes. A well-formed subscribe gave no error in the checks.

- Malformed `params` give `invalid_request`. For example, `{"subs":[]}` gives `missing field \`subscriptions\``.
- An unknown `type` variant gives `invalid_request` (`unknown variant`).
- An invalid `agent_status` filter value gives `invalid_request` (`unknown variant`).
- An invalid regex in an `OutputMatch` gives a different code, `invalid_regex`, and not `invalid_request`. The message contains the parse error from the regex crate. For `"("`, the message is `regex parse error: … error: unclosed group`.
- Subscription variant objects accept unknown extra fields. `{"type":"tab.created","extra":"junk"}` gives the usual ack and no error.
- An empty `subscriptions` array gives the usual ack.
- A `pane_id` that names no pane gives `pane_not_found`. The error contains the `id` of the request, at each index of the failing item.
- After each error, the server closes the connection.
- Other codes are possible.

```json
{"id":"e4","method":"events.subscribe","params":{"subscriptions":[{"type":"tab.created"},{"type":"pane.scroll_changed","pane_id":"w99:p99"}]}}
{"id":"e4","error":{"code":"pane_not_found","message":"pane w99:p99 not found"}}
```

**Changed from 0.9.1**: on 0.9.1, a failing item after the first item returned the `id` `<request id>:sub:<index>:probe`, and not the request `id`. For this request, the `id` was `e4:sub:1:probe`. `events.wait` still returns this changed `id`. Refer to [events.wait](#eventswait).

`Checked 2026-10-17 against herdr 0.9.5`: each error in this list, `pane_not_found` at the item indexes 0, 1, and 2, and the close of the connection.

**CLI**: API-only. No `herdr events` CLI group exists. Two CLI commands give related one-shot waits: `herdr pane wait-output` and `herdr agent wait`. `Checked 2026-10-17 against herdr 0.9.5.`

**Example**:

```json
{"id":"e1","method":"events.subscribe","params":{"subscriptions":[{"type":"tab.created"},{"type":"tab.closed"}]}}
{"id":"e1","result":{"type":"subscription_started"}}
{"data":{"tab":{"agent_status":"unknown","focused":true,"label":"1","number":1,"pane_count":1,"tab_id":"w1:t1","workspace_id":"w1"},"type":"tab_created"},"event":"tab_created"}
{"data":{"tab":{"agent_status":"unknown","focused":false,"label":"2","number":2,"pane_count":1,"tab_id":"w1:t2","workspace_id":"w1"},"type":"tab_created"},"event":"tab_created"}
```

The first line is the request, and the second line is the ack. The server pushed the next lines on the same open connection when herdr created tabs.

`Checked 2026-10-17 against herdr 0.9.5`: the two `tab_created` pushes came from `workspace.create`, then from `tab.create` with `{"workspace_id":"w1"}`, on a new scratch server.

## events.wait

`events.wait` keeps the connection open until herdr sees one event that matches the `match_event` predicate. The method then returns that event as a usual response (`wait_matched`). It closes the connection, as other one-shot requests do.

- With `timeout_ms` set, the wait returns an error if the time ends before a match.
- With `timeout_ms` null or not given, the wait has no time limit.
- Different from `events.subscribe`, this method returns exactly one event, and the connection does not stay open.

`Checked 2026-10-17 against herdr 0.9.5`: the `wait_matched` path and the timeout path. The connection closed after each response.
`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the wait with no time limit. The schema gives `timeout_ms` as `["integer","null"]`. The pass did not run this wait, because the probe would not end.

**Implementation gap**: the `EventMatch` schema accepts 19 predicate variants (below). But herdr 0.9.5 rejects each match other than a pane agent-status match. The error code is `unsupported_event_wait_match`, with the message `events.wait currently supports pane agent status matches`.

- A pass sent these matches, and each match got this error: `tab_created`, `tab_closed`, `workspace_created`, `pane_created`, `pane_exited`, `pane_output_changed`, `pane_agent_detected`, `workspace_focused`, `pane_focused`, and `tab_renamed`.
- This gap is the same as in 0.8.2, 0.9.1, and 0.9.3.

Use the schema as the surface for later versions. For now, send only a `pane_agent_status_changed` match. For the other kinds, use `events.subscribe` and read the first push.

`Checked 2026-10-17 against herdr 0.9.5`: the rejection of each match in the list.

**Params**: `EventsWaitParams`.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| match_event | `EventMatch` | yes | — | The predicate that selects the event to wait for. Refer to the variants below. |
| timeout_ms | integer (uint64) \| null | no | null | The maximum time to wait, in milliseconds. Null or no value means no time limit. |

### EventMatch variants

`EventMatch` is a `oneOf` with the discriminator `event`. The `event` value is snake_case and matches `EventKind`. `EventMatch` has 19 variants.

- On 0.9.5, herdr accepts only `pane_agent_status_changed`. The other variants are valid in the schema, but they return `unsupported_event_wait_match`.
- The "required" column lists the required fields of each variant. Each other field is an optional filter.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: this table.

| `event` const | filter fields | required |
| --- | --- | --- |
| `workspace_created` | `workspace_id`: string \| null | `event` |
| `workspace_updated` | `workspace_id`: string | `event`, `workspace_id` |
| `workspace_closed` | `workspace_id`: string | `event`, `workspace_id` |
| `workspace_renamed` | `workspace_id`: string, `label`: string \| null | `event`, `workspace_id` |
| `workspace_moved` | `workspace_id`: string | `event`, `workspace_id` |
| `workspace_focused` | `workspace_id`: string | `event`, `workspace_id` |
| `tab_created` | `tab_id`: string \| null, `workspace_id`: string \| null | `event` |
| `tab_closed` | `tab_id`: string | `event`, `tab_id` |
| `tab_renamed` | `tab_id`: string, `label`: string \| null | `event`, `tab_id` |
| `tab_moved` | `tab_id`: string | `event`, `tab_id` |
| `tab_focused` | `tab_id`: string | `event`, `tab_id` |
| `pane_created` | `pane_id`: string \| null, `workspace_id`: string \| null | `event` |
| `pane_closed` | `pane_id`: string | `event`, `pane_id` |
| `pane_focused` | `pane_id`: string | `event`, `pane_id` |
| `pane_moved` | `pane_id`: string | `event`, `pane_id` |
| `pane_output_changed` | `pane_id`: string, `min_revision`: integer (uint64) \| null | `event`, `pane_id` |
| `pane_exited` | `pane_id`: string | `event`, `pane_id` |
| `pane_agent_detected` | `pane_id`: string, `agent`: string \| null | `event`, `pane_id` |
| `pane_agent_status_changed` | `pane_id`: string, `agent_status`: `AgentStatus` (`idle`\|`working`\|`blocked`\|`done`\|`unknown`) | `event`, `pane_id`, `agent_status` |

The `EventMatch` surface is smaller than the full `EventKind` list. It does not include these 7 kinds: `workspace_metadata_updated`, `workspace_reordered`, `worktree_created`, `worktree_opened`, `worktree_removed`, `pane_updated`, and `layout_updated`. You cannot wait for these kinds, also in the schema.

`Checked 2026-10-17 against herdr 0.9.5`:

- A `layout_updated` match gives `invalid_request` (`unknown variant`). The message lists the 19 variants.
- A `pane_agent_status_changed` match without `agent_status` gives `invalid_request` (`missing field \`agent_status\``).

**Result**: `type` const `wait_matched`.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| type | string | yes | — | Always `wait_matched`. |
| event | `EventEnvelope` | yes | — | The matched event, with the keys `event` (an `EventKind`) and `data` (an `EventData`). In this response, `event` comes before `data`, as in the schema. Refer to the [event catalog](#event-catalog-eventkind--eventdata). |

`Checked 2026-10-17 against herdr 0.9.5`: the result shape and its key order.

**Errors**: [errors.md](errors.md) gives the error codes.

| code | when |
| --- | --- |
| `unsupported_event_wait_match` | On herdr 0.9.5, the `match_event` variant is not `pane_agent_status_changed`. Message: `events.wait currently supports pane agent status matches`. |
| `timeout` | The `timeout_ms` time ended before a matching event arrived. Message: `timed out waiting for event match`. |
| `pane_not_found` | The `match_event.pane_id` names a pane that does not exist. |

- `pane_not_found` changes the response `id` to `<request id>:sub:0:probe`, and not the request `id`. For example, a request with `id: "ZZZ"` gets `"id":"ZZZ:sub:0:probe"`.
- On 0.9.1, `events.subscribe` also changed the `id`. On 0.9.3 and 0.9.5, only `events.wait` changes it.
- herdr checks for an unsupported match first. Thus, a `pane_exited` match on a missing pane returns `unsupported_event_wait_match` with the request `id` unchanged.
- A request without `match_event`, with only `timeout_ms`, gives `invalid_request` (`missing field \`match_event\``). This error contains the request `id`.
- Other codes are possible.

`Checked 2026-10-17 against herdr 0.9.5`: each error in the table and in this list. With `timeout_ms: 500`, the `timeout` error arrived after 503 ms.

**CLI**: API-only. No `herdr events` CLI group exists. `herdr agent wait <TARGET>` gives the supported agent-status wait. `herdr pane wait-output` gives the output wait. `Checked 2026-10-17 against herdr 0.9.5.`

**Example**: herdr rejects a match that is not an agent-status match. A supported match can end with a timeout.

```json
{"id":"e1","method":"events.wait","params":{"match_event":{"event":"tab_created"},"timeout_ms":500}}
{"id":"e1","error":{"code":"unsupported_event_wait_match","message":"events.wait currently supports pane agent status matches"}}
```

```json
{"id":"e3","method":"events.wait","params":{"match_event":{"event":"pane_agent_status_changed","pane_id":"w1:p1","agent_status":"idle"},"timeout_ms":500}}
{"id":"e3","error":{"code":"timeout","message":"timed out waiting for event match"}}
```

`Checked 2026-10-17 against herdr 0.9.5`: herdr returned these two responses byte for byte.

A match on the current status of the pane returns `wait_matched` at once and does not block. The match is level-triggered, not edge-triggered. A check gave these results:

1. An agent in a tab without focus got the report `working`, then the report `idle`. herdr shows this agent as `done`.
2. A wait for `agent_status: "done"` returned at once, with no other state change. The round trip took about 4 ms, and this time includes the start of the probe script.
3. A wait for `idle` on the same pane got a timeout.
4. After `agent.focus` changed the status to `idle`, a wait for `idle` returned at once.

A pane with no agent gets a level match as `unknown` in the same way. The result contains `{"agent_status":"unknown"}` and no `agent` key.

[AgentStatus](data-model.md#agentstatus) gives the meaning of each status. On herdr 0.9.5, that section also tells if `idle` and `done` show that an agent is ready.

```json
{"id":"e4","method":"events.wait","params":{"match_event":{"event":"pane_agent_status_changed","pane_id":"w1:p2","agent_status":"done"},"timeout_ms":2000}}
{"id":"e4","result":{"type":"wait_matched","event":{"event":"pane_agent_status_changed","data":{"type":"pane_agent_status_changed","pane_id":"w1:p2","workspace_id":"w1","agent_status":"done","agent":"testbot"}}}}
```

`Checked 2026-10-17 against herdr 0.9.5`: the four steps, the `unknown` level match, and this example.

To get this result with `pane.report_agent`, first run a real foreground process in the pane. The check used `exec -a testbot sleep 3000`. Sometimes the foreground process of the pane is the plain shell. Then an `idle` report releases the agent at once. herdr pushes `pane_agent_detected` with `released: true` and `final_status: "unknown"`. The pane then shows `unknown`, and not `idle` or `done`:

```json
{"data":{"agent":"testbot","final_status":"unknown","pane_id":"w1:p2","released":true,"type":"pane_agent_detected","workspace_id":"w1"},"event":"pane_agent_detected"}
```

`Checked 2026-10-17 against herdr 0.9.5`: the release and this push.

## Event catalog (EventKind / EventData)

This section lists the 26 `EventKind` values and their `EventData` payloads. The `EventKind` enum and the `EventData` variants in `raw/schema.json` are the same as in 0.9.1 and 0.9.3.

- Each pushed plain event and each `wait_matched.event` is an `EventEnvelope`. It has the keys `event` (an `EventKind`) and `data` (an `EventData`). The `data.type` repeats the kind.
- `EventData` is a `oneOf` with the discriminator `data.type`. The tables below list the fields of each variant.
- [data-model.md](data-model.md) defines the domain types `WorkspaceInfo`, `TabInfo`, `PaneInfo`, `WorktreeInfo`, `PaneLayoutSnapshot`, and `AgentStatus`. This page gives only the field names.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the tables in this section.
`Checked 2026-10-17 against herdr 0.9.5`: the key set of 25 of the 26 kinds matched its row. The key set of `pane_output_changed` was not checked. Refer to the sweep below.

`EventEnvelope`:

| field | type | required | meaning |
| --- | --- | --- | --- |
| event | `EventKind` | yes | The kind, in snake_case. It is one of the 26 kinds below. |
| data | `EventData` | yes | The payload. Its `type` is equal to `event`. |

**Workspace kinds**

| `type` | payload fields (required unless noted) |
| --- | --- |
| `workspace_created` | `workspace`: WorkspaceInfo |
| `workspace_updated` | `workspace`: WorkspaceInfo |
| `workspace_metadata_updated` | `workspace`: WorkspaceInfo |
| `workspace_closed` | `workspace_id`: string, `workspace`: WorkspaceInfo \| null (optional) |
| `workspace_renamed` | `workspace_id`: string, `label`: string |
| `workspace_moved` | `workspace_id`: string, `insert_index`: integer (uint), `workspaces`: array of WorkspaceInfo (the new order) |
| `workspace_reordered` | `workspace_ids`: array of string (the new order), `workspaces`: array of WorkspaceInfo, `before_workspace_id`: string \| null (optional) |
| `workspace_focused` | `workspace_id`: string |

herdr pushed `workspace_updated` in two cases on herdr 0.9.5. Earlier passes did not see this event. It is not known if 0.9.3 pushed it in the same cases.

- The first `worktree.create` from a workspace that has no `worktree` data. herdr pushes `workspace_updated` for the source workspace, which now contains its `worktree` field. This push comes after `workspace_created` (for the new workspace) and before `worktree_created`. A second `worktree.create` from the same workspace does not push `workspace_updated`.
- A `worktree.open` without a `label`, for a worktree that is already open. herdr pushes `workspace_updated` for the opened workspace before `worktree_opened` (`already_open: true`). With a new `label`, herdr pushes `workspace_renamed` in its place.

Other operations that change `WorkspaceInfo` pushed a more specific event or no event. These operations include rename, move, reorder, metadata reports, and focus. The specific events include `workspace_renamed`, `workspace_moved`, `workspace_metadata_updated`, and `workspace_focused`.

The first capture comes from `worktree.create` with `{"workspace_id":"w7","branch":"feat-x","label":"build"}`. Before the call, `workspace.create` made `w7` in the repository, and `w7` had no `worktree` key:

```json
{"data":{"type":"workspace_updated","workspace":{"active_tab_id":"w7:t1","agent_status":"unknown","focused":false,"label":"repo2","number":5,"pane_count":1,"tab_count":1,"workspace_id":"w7","worktree":{"checkout_path":"/home/user/project/repo2","is_linked_worktree":false,"repo_key":"/home/user/project/repo2/.git","repo_name":"repo2","repo_root":"/home/user/project/repo2"}}},"event":"workspace_updated"}
```

The second capture comes from `worktree.open` with `{"workspace_id":"w1","branch":"feat-b"}`, while `feat-b` was open in `w4`:

```json
{"data":{"type":"workspace_updated","workspace":{"active_tab_id":"w4:t1","agent_status":"unknown","focused":false,"label":"reviewer","number":2,"pane_count":1,"tab_count":1,"workspace_id":"w4","worktree":{"checkout_path":"/home/user/project/wt/repo/feat-b","is_linked_worktree":true,"repo_key":"/home/user/project/repo/.git","repo_name":"repo","repo_root":"/home/user/project/repo"}}},"event":"workspace_updated"}
```

`Checked 2026-10-17 against herdr 0.9.5`: the two cases, the second `worktree.create`, the `label` case, and the two captures.

**Worktree kinds**

| `type` | payload fields (required unless noted) |
| --- | --- |
| `worktree_created` | `workspace`: WorkspaceInfo, `worktree`: WorktreeInfo |
| `worktree_opened` | `workspace`: WorkspaceInfo, `worktree`: WorktreeInfo, `already_open`: boolean |
| `worktree_removed` | `workspace_id`: string, `worktree`: WorktreeInfo, `forced`: boolean, `workspace`: WorkspaceInfo \| null (optional) |

The pushed `WorktreeInfo` contained the keys `branch`, `is_bare`, `is_detached`, `is_linked_worktree`, `is_prunable`, `label`, `open_workspace_id`, and `path`. `Checked 2026-10-17 against herdr 0.9.5.`

**Tab kinds**

| `type` | payload fields (required unless noted) |
| --- | --- |
| `tab_created` | `tab`: TabInfo |
| `tab_closed` | `tab_id`: string, `workspace_id`: string |
| `tab_renamed` | `tab_id`: string, `workspace_id`: string, `label`: string |
| `tab_moved` | `tab_id`: string, `workspace_id`: string, `insert_index`: integer (uint), `tabs`: array of TabInfo (the new order) |
| `tab_focused` | `tab_id`: string, `workspace_id`: string |

**Pane kinds**

| `type` | payload fields (required unless noted) |
| --- | --- |
| `pane_created` | `pane`: PaneInfo |
| `pane_closed` | `pane_id`: string, `workspace_id`: string |
| `pane_updated` | `pane`: PaneInfo |
| `pane_focused` | `pane_id`: string, `workspace_id`: string |
| `pane_moved` | `previous_pane_id`: string, `previous_workspace_id`: string, `previous_tab_id`: string, `pane`: PaneInfo. Optional: `closed_tab_id`: string \| null, `closed_workspace_id`: string \| null, `created_tab`: TabInfo \| null, `created_workspace`: WorkspaceInfo \| null |
| `pane_output_changed` | `pane_id`: string, `workspace_id`: string, `revision`: integer (uint64) |
| `pane_exited` | `pane_id`: string, `workspace_id`: string |
| `pane_agent_detected` | `pane_id`: string, `workspace_id`: string. Optional: `agent`: string \| null, `final_status`: AgentStatus \| null, `released`: boolean |
| `pane_agent_status_changed` | `pane_id`: string, `workspace_id`: string, `agent_status`: AgentStatus. Optional: `agent`: string \| null, `display_agent`: string \| null, `title`: string \| null, `state_labels`: map<string,string> |

herdr pushes `pane_updated` in these cases:

- herdr creates a pane.
- The current directory of the foreground process changes. One `cd /tmp` gave one push.
- A `pane.report_metadata` call changes the `tokens` of the pane. The push contains the new `tokens`.

These calls push nothing: a `pane.report_metadata` call that does not change `tokens`, and `pane.rename`. **Changed from 0.9.1**: on 0.9.1, `pane.report_metadata` did not push `pane_updated`.

Earlier passes also saw `pane_updated` for pane changes near layout changes. On a server with no TUI client attached, `layout.set_split_ratio` pushed no `pane_updated`. Refer to the [ordering notes](#ordering-and-grouping) below.

`Checked 2026-10-17 against herdr 0.9.5`: the three cases, the two calls that push nothing, and the `layout.set_split_ratio` result.

The `pane_output_changed` payload shape comes only from the schema. No plain subscription exists for this kind. `events.wait` rejects it with `unsupported_event_wait_match`. No method that a client can call triggers or returns this kind directly. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5).`

A recognized agent keeps its identity when Ctrl+Z stops its process. In this case, herdr pushes no `pane_agent_detected` release. An agent that only `pane.report_agent` gave, with a label that herdr does not recognize, got a release at once on Ctrl+Z (`released: true`, `final_status: "unknown"`). [api/agent.md](api/agent.md#agent-identity-across-ctrlz) owns this fact. `Checked 2026-10-17 against herdr 0.9.5.`

**Layout kind**

| `type` | payload fields (required unless noted) |
| --- | --- |
| `layout_updated` | `layout`: PaneLayoutSnapshot |

`AgentStatus` enum (used above): `idle`, `working`, `blocked`, `done`, `unknown`. [AgentStatus](data-model.md#agentstatus) gives the meaning of each value.

Cross-check: the checked `tab.created` capture pushed `{"data":{"tab":{…TabInfo…},"type":"tab_created"},"event":"tab_created"}`. This line matches the `tab_created` row. It has one `tab: TabInfo` field, and `data.type` repeats `event`. `Checked 2026-10-17 against herdr 0.9.5.`

Scratch-server sweep: one connection subscribed to all 24 plain types on a scratch server. The server then ran these operations:

- workspace: create, rename, `report_metadata`, move, `move_block`, focus, close
- tab: create, rename, move, focus, close
- pane: split, focus, move to a new tab, close, exit, `report_metadata`, rename
- `layout.set_split_ratio` and `agent.focus`

The connection received 19 kinds:

- `workspace_created`, `workspace_renamed`, `workspace_metadata_updated`, `workspace_moved`, `workspace_reordered` (with and without `before_workspace_id`), `workspace_focused`, `workspace_closed` (with `workspace`)
- `tab_created`, `tab_renamed`, `tab_moved`, `tab_focused`, `tab_closed`
- `pane_created`, `pane_updated`, `pane_focused`, `pane_moved`, `pane_closed`, `pane_exited`
- `layout_updated`

Other checks saw more kinds:

- A second scratch server with worktree operations pushed `workspace_updated`, `worktree_created`, `worktree_opened`, and `worktree_removed`.
- Agent checks saw `pane_agent_detected`, in the detection form and in the `released: true` form with `final_status`.
- `events.wait` results contained `pane_agent_status_changed`.

The key set of each payload matched its row above. Optional keys were absent when they did not apply. A `pane_moved` to a new tab contained `created_tab`. In total, the checks saw 25 of the 26 kinds. They did not see `pane_output_changed`.

`Checked 2026-10-17 against herdr 0.9.5`: the sweep and the other checks.

### Ordering and grouping

The same checks saw this order and grouping of events. For one state change, herdr pushes the events in the order of the `subscriptions` array. The checks listed the 24 plain types in schema order. A different array order gives a different push order. Differences from 0.9.3 in this list can come from a different subscription order.

- `workspace.create` pushed `workspace_created`, `tab_created`, `pane_created`, then `pane_updated` and `layout_updated`. `tab.create` pushed `tab_created`, `pane_created`, then `pane_updated` and `layout_updated`. The order of the last two events is not fixed. Sometimes `pane_updated` comes in a later push, after `layout_updated`. On 0.9.3, a pass saw no `pane_created` from these calls.
- Sometimes the new workspace gets the focus. Examples are the first workspace on a server, and `worktree.create` with `focus: true`. Then `workspace_focused`, `tab_focused`, and `pane_focused` come after the matching `..._created` event.
- `pane.split` pushed `pane_created`, then `pane_updated` and `layout_updated`. The order of the last two events is not fixed. In 6 runs, `pane_updated` came first 5 times, and `layout_updated` came first 1 time.
- `worktree.create` pushed `workspace_created`, `workspace_updated` (only in the case above), `worktree_created`, `tab_created`, `pane_created`, `pane_updated`, and `layout_updated`. With `focus: true`, the three focus events occur between these events.
- `worktree.open` of a closed checkout pushed `workspace_created`, `worktree_opened` (`already_open: false`), `tab_created`, `pane_created`, `pane_updated`, and `layout_updated`. For a worktree that is already open, it pushed `worktree_opened` with `already_open: true`.
- `worktree.remove` pushed `workspace_closed` (with `workspace`), then `worktree_removed` with `forced: false`.
- `tab.focus` to a tab of another workspace pushed `workspace_focused`, `tab_focused`, and `pane_focused` together. `tab.focus` to another tab in the same workspace also pushed all three. It pushed `workspace_focused` for the workspace that already had the focus.
- `workspace.move_block` pushed `workspace_reordered` first. It then pushed `workspace_focused`, `tab_focused`, and `pane_focused` for the focused workspace, although the focus did not move. `workspace.move` pushed `workspace_moved`, then the same three focus events. On 0.9.3, a pass saw the three focus events before `workspace_reordered`, for a `workspace.move_block` that also moved the focus.
- A `workspace.move_block` that changes no position pushed nothing.
- `agent.focus` on the agent that already had the focus pushed nothing. `workspace.focus` on the focused workspace and `pane.focus` on the focused pane also pushed nothing.
- A close of the last tab of a workspace pushed `workspace_closed` before `tab_closed`. A close of another tab pushed only `tab_closed`.
- `workspace.close` pushed only `workspace_closed` (with `workspace`). It pushed no `tab_closed` and no `pane_closed`. It also pushed no focus events for the workspace that gets the focus.
- `pane.close` of the last pane in the last tab of a workspace pushed `workspace_closed`, then `pane_closed`, with no `tab_closed`. After the call, the workspace does not exist. The API call returned `ok` with no confirmation. Refer to [pane.close](api/pane.md#paneclose) for the TUI confirmation.
- `pane.close` of the only pane of a tab that is not the last tab pushed only `pane_closed`. After the call, the tab does not exist, and herdr pushed no `tab_closed`.
- A pane whose process exits, such as with a shell `exit`, pushes `pane_exited`. It also pushes `layout_updated` when other panes stay in the tab. It never pushes `pane_closed`. When the pane was the only pane of a tab or of a workspace, herdr pushed only `pane_exited`. The tab or the workspace then did not exist, and herdr pushed no `tab_closed` or `workspace_closed`.
- `workspace_reordered` does not contain `before_workspace_id` when the request sent `null`. It contains `before_workspace_id` when the request named a workspace.

`Checked 2026-10-17 against herdr 0.9.5`: each item in this list. The checks used a scratch server with no TUI client attached.

Other pages own these event claims:

- `layout.set_split_ratio` pushed only `layout_updated`, and no `pane_updated`. The server had no TUI client attached. The layout rects of the panes changed. Their terminal sizes (`viewport_rows`) did not change. On 0.9.3, a pass saw one `pane_updated` for each pane that changed size. Refer to [api/layout.md](api/layout.md). `Checked 2026-10-17 against herdr 0.9.5.`
- `layout.apply` with a `tab_id` pushes `tab_closed` for the replaced tab. It pushes no `pane_closed` for the panes of that tab. Refer to [api/layout.md](api/layout.md#layoutapply). `Checked 2026-10-17 against herdr 0.9.5.`
- `plugin.pane.close` of a plugin pane with `tab` placement removes its tab, but it pushes only `pane_closed`. Refer to [api/plugin.md](api/plugin.md). `Checked 2026-10-17 against herdr 0.9.5.`
- The editor pane of `pane.edit_scrollback` closes with `pane_exited`, then `layout_updated`, and no `pane_closed`. [api/pane.md](api/pane.md#paneedit_scrollback) owns this claim and its stamp. This page did not check it.

## Subscription events (SubscriptionEventKind / SubscriptionEventData)

The three parameterized subscription types push a different envelope from the plain event catalog. This envelope is the `SubscriptionEventEnvelope`. It has a dotted `event` (`SubscriptionEventKind`) and a `SubscriptionEventData` payload that herdr computes for the subscription. Only `events.subscribe` makes this envelope. `events.wait` never returns it, because its result contains the plain `EventEnvelope`.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the tables in this section.
`Checked 2026-10-17 against herdr 0.9.5`: each `events.wait` result contained the snake_case `pane_agent_status_changed` with a `data.type` key.

`SubscriptionEventKind` enum: `pane.output_matched`, `pane.agent_status_changed`, `pane.scroll_changed`.

`SubscriptionEventEnvelope`:

| field | type | required | meaning |
| --- | --- | --- | --- |
| event | `SubscriptionEventKind` | yes | The dotted kind. It is one of the three kinds above. |
| data | `SubscriptionEventData` | yes | The payload, one of the three variants below. The payload is an `anyOf`. Use `event` to find the variant, because the payloads have no `type` discriminator. |

**`pane.output_matched` → `PaneOutputMatchedEvent`**

| field | type | required | meaning |
| --- | --- | --- | --- |
| pane_id | string | yes | The pane that matched. |
| matched_line | string | yes | The line that satisfied the `OutputMatch`. herdr removes ANSI escapes if `strip_ansi` is `true`. |
| read | `PaneReadResult` | yes | A snapshot read of the pane around the match. |

`PaneReadResult`:

| field | type | required | meaning |
| --- | --- | --- | --- |
| pane_id | string | yes | The pane that herdr read. |
| workspace_id | string | yes | The workspace that contains the pane. |
| tab_id | string | yes | The tab that contains the pane. |
| source | `ReadSource` enum | yes | `visible`, `recent`, `recent_unwrapped`, or `detection`. |
| format | `ReadFormat` enum | yes | `text` or `ansi`. |
| text | string | yes | The read text. |
| revision | integer (uint64) | yes | The output revision that this read shows. |
| truncated | boolean | yes | True if herdr cut `text` to a shorter length. |

`Checked 2026-10-17 against herdr 0.9.5`: a push contained the keys `matched_line`, `pane_id`, and `read`. Its `read` contained the keys `format`, `pane_id`, `revision`, `source`, `tab_id`, `text`, `truncated`, and `workspace_id`.

**`pane.agent_status_changed` → `PaneAgentStatusChangedEvent`**

| field | type | required | meaning |
| --- | --- | --- | --- |
| pane_id | string | yes | The pane whose agent status changed. |
| workspace_id | string | yes | The workspace that contains the pane. |
| agent_status | `AgentStatus` enum | yes | The new status: `idle`, `working`, `blocked`, `done`, or `unknown`. |
| agent | string \| null | no | The agent identifier, if known. |
| display_agent | string \| null | no | The agent name for a person to read, if known. |
| title | string \| null | no | The title of the pane or of the agent, if known. |
| state_labels | map<string,string> | no | State labels that the agent gives. |

Note: this payload has the same fields as the plain `pane_agent_status_changed` `EventData`. But a `pane.agent_status_changed` subscription delivers it in the dotted `SubscriptionEventEnvelope`. A subscription with no `agent_status` filter pushed each change of an agent that `pane.report_agent` controlled. It pushed `working`, then `done` in a tab without focus, and then `idle` after `agent.focus`. Each push had no optional keys:

```json
{"data":{"agent":"testbot","agent_status":"working","pane_id":"w1:p2","workspace_id":"w1"},"event":"pane.agent_status_changed"}
```

`Checked 2026-10-17 against herdr 0.9.5`: the three pushes and this capture.

**`pane.scroll_changed` → `PaneScrollChangedEvent`**

| field | type | required | meaning |
| --- | --- | --- | --- |
| pane_id | string | yes | The pane whose scroll position changed. |
| workspace_id | string | yes | The workspace that contains the pane. |
| scroll | `PaneScrollInfo` | yes | The scroll snapshot. Refer to `PaneScrollInfo` below. |

`PaneScrollInfo`:

| field | type | required | meaning |
| --- | --- | --- | --- |
| offset_from_bottom | integer (uint64) | yes | The current scrollback offset from the bottom. The value 0 means that the view is at the bottom. |
| max_offset_from_bottom | integer (uint64) | yes | The maximum possible offset. |
| viewport_rows | integer (uint64) | yes | The height of the visible viewport, in rows. |

A `pane.scroll` to `offset_from_bottom: 50` pushed this line:

```json
{"data":{"pane_id":"w1:p1","scroll":{"max_offset_from_bottom":261,"offset_from_bottom":50,"viewport_rows":40},"workspace_id":"w1"},"event":"pane.scroll_changed"}
```

New output also pushes:

- While the pane was above the bottom, 4 new rows changed the pane to `offset_from_bottom: 54` and `max_offset_from_bottom: 265`. The viewport stays on the same content.
- After a `pane.scroll` back to 0, new output at the bottom pushed a change to `max_offset_from_bottom` only: `{"max_offset_from_bottom":267,"offset_from_bottom":0,"viewport_rows":40}`.

The numbers depend on the capture. The pane had 40 viewport rows on a server with no TUI client attached. `Checked 2026-10-17 against herdr 0.9.5.`

`ReadFormat` enum: `text`, `ansi`. `ReadSource` enum: `visible`, `recent`, `recent_unwrapped`, `detection`.

Each `PaneReadResult` from `pane.output_matched` had `format: "text"` in the checks. This includes a subscription with `strip_ansi: false` on a line with real escape bytes. `Checked 2026-10-17 against herdr 0.9.5.`

No parameter of this subscription selects `format: "ansi"`. Thus, it is not known if or when `ansi` can occur here. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5).`

## Subscription vs EventKind: name and coverage diffs

Three surfaces use different names for the same domain. A client must map the names explicitly.

- `Subscription.type` and `SubscriptionEventKind` use dotted names.
- `EventKind`, `EventData.type`, and `EventMatch.event` use snake_case names.

For the 24 plain subscription types, change each dot to an underscore (`tab.created` → `tab_created`). A direct comparison of the enum lists shows these exceptions.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the enum comparisons in this section.
`Checked 2026-10-17 against herdr 0.9.5`: the rejection of `pane.output_changed` and the `events.wait` rejections.

**Subscription types with no `EventKind` of the same name**:

| `Subscription.type` | the delivered data |
| --- | --- |
| `pane.output_matched` | A `SubscriptionEventEnvelope` with the `event` `pane.output_matched` (`PaneOutputMatchedEvent`). No `pane_output_matched` `EventKind` exists. The related raw catalog event is `pane_output_changed`. You **cannot** subscribe to that event alone. Only this filtered subscription with parameters gives access to output. |
| `pane.scroll_changed` | A `SubscriptionEventEnvelope` with the `event` `pane.scroll_changed` (`PaneScrollChangedEvent`). No `pane_scroll_changed` `EventKind` exists. Scroll changes occur only as a subscription event, and not in the plain catalog. |

**`EventKind` values with no matching `Subscription.type`**:

| `EventKind` | note |
| --- | --- |
| `pane_output_changed` | No `pane.output_changed` subscription exists. Use `pane.output_matched`. This kind is still in `EventData`, and `EventMatch` can match it. |

All other names map one to one. Each of the 24 plain subscription types maps to a different catalog `EventKind`. The `pane.agent_status_changed` subscription maps to two kinds:

- the plain `pane_agent_status_changed` `EventKind`, through `events.wait` and other sources
- the dotted `pane.agent_status_changed` `SubscriptionEventKind`, through this subscription

**`EventMatch` coverage in `events.wait`**: this surface covers fewer kinds. Only 19 of the 26 `EventKind` values have a match variant. The schema has no match for `workspace_metadata_updated`, `workspace_reordered`, `worktree_created`, `worktree_opened`, `worktree_removed`, `pane_updated`, or `layout_updated`. Also, herdr 0.9.5 accepts only `pane_agent_status_changed` matches. All other matches give `unsupported_event_wait_match`.
