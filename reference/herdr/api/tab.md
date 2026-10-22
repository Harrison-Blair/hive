# herdr API: tab methods

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Index: [README.md](../README.md). Wire format: [protocol.md](../protocol.md).

The `tab.*` methods manage the tabs of a workspace.

- A tab holds one or more panes. Each tab is in exactly one workspace.
- A tab has a `number`, a `label`, a `focused` flag, a `pane_count`, and an aggregate `agent_status`. The `number` is a creation counter for each workspace, not a position. Refer to [TabInfo](#tabinfo).
- When you create a tab, herdr also creates its `root_pane`. When you close a tab, herdr closes the panes in the tab.
- Most methods operate on one tab, and its `tab_id` identifies it. `tab.list` lists tabs. `tab.move` changes the order of tabs in a workspace.

Tab IDs have the form `w1:t1`. They are opaque, stable, and qualified by the workspace. herdr does not use a tab ID again after the tab closes. The suffix uses the base32 alphabet that [addressing.md](../addressing.md#id-scheme) defines. For example, `w1:tA` comes after `w1:t9`. Thus, a caller that reads the suffix as decimal digits fails on the 10th tab of a workspace. A pass made tab IDs from `t1` to `tY`. herdr did not use `I`, `L`, `O`, or `U`. The `number` of each tab was the base32 value of its suffix.

Request rules:

- All seven methods need an explicit `params` object. This rule applies also when every field is optional, as in `tab.list` and `tab.create`.
- A field with the wrong type gives a hard `invalid_request` error.
- When the line is valid JSON and has a string `id`, the `invalid_request` error contains that request `id`. On herdr 0.9.3 and 0.9.5, the response `id` is `""` only in two cases. The line is not valid JSON, or its `id` is missing or is not a string. On herdr 0.9.1, every deserialization error gave the response `id` `""`.
- herdr ignores unknown extra fields in `params`.

The `herdr tab` CLI writes compact JSON, with the keys in alphabetical order. The examples on this page show the keys in declaration order. On an API error, the CLI writes the error envelope to stderr and exits with status `1`. For an unknown subcommand, it writes a usage message and exits with status `2`. [cli-mapping.md](../cli-mapping.md#cli-conventions) gives these CLI conventions.

`Checked 2026-10-17 against herdr 0.9.5`: this stamp applies to all claims in this introduction. These claims include:

- the TabInfo keys and the tab IDs from `t1` to `tY`
- the IDs that herdr does not use again, and the closed panes
- the `params` rule, the type errors, the response `id` cases, and the extra fields
- the CLI output and the exit status

7 methods:

| method | purpose |
| --- | --- |
| [tab.close](#tabclose) | Close a tab and its panes. |
| [tab.create](#tabcreate) | Create a tab and its root pane in a workspace. |
| [tab.focus](#tabfocus) | Focus a tab in the UI and mark it seen. |
| [tab.get](#tabget) | Get the metadata of one tab by ID. |
| [tab.list](#tablist) | List tabs. You can limit the list to one workspace. |
| [tab.move](#tabmove) | Change the position of a tab in its workspace. |
| [tab.rename](#tabrename) | Change the label of a tab. |

## Shared types

### TabInfo

Every tab result except `tab.close` contains `TabInfo`. `tab.close` returns `ok`. [../data-model.md](../data-model.md#tabinfo) also lists this type.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the field types and the required flags in this table. The row text gives the checked behavior.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `tab_id` | string | yes | — | Opaque tab ID, qualified by the workspace (for example, `w1:t1`). |
| `workspace_id` | string | yes | — | ID of the workspace that contains the tab. |
| `number` | integer (uint, ≥0) | yes | — | Tab creation counter for each workspace, not a position. In practice, the value is 1 or more. It is stable. A move or a close does not change it. |
| `label` | string | yes | — | Tab label that the herdr UI shows. An automatic label (from a null initial label) follows the live position of the tab. It can change with no `tab_renamed` event. herdr stores an explicit label exactly as the caller sent it. |
| `focused` | boolean | yes | — | True when this tab has the focus in the herdr UI. |
| `pane_count` | integer (uint, ≥0) | yes | — | Number of panes in the tab. |
| `agent_status` | enum | yes | — | Aggregate agent status of the panes in the tab. One of `idle`, `working`, `blocked`, `done`, `unknown`. |

The `number` is the decimal value of the base32 suffix of the tab ID. For example, `w1:tA` has the number 10, and `w1:tR` has the number 24. herdr does not calculate the `number` again. When you move `w1:t1` to a later position, its `number` stays `1`. When a tab between other tabs closes, the numbers have a gap. For example, the numbers are `1, 2, 3, 5` after the 4th tab of a workspace closes. The schema gives `number` only the type `uint ≥ 0`, with no field description. Thus, the schema does not say if `number` is a counter or a position.

[../data-model.md](../data-model.md#agentstatus) gives the meaning of each `agent_status` value. In practice, a tab with only a plain shell in its pane also has the status `unknown`. The status does not show if the tab has no agent or has an agent that herdr cannot classify. A tab can show `working` when its pane runs an agent. A tab can show `blocked` when its agent shows an approval UI that herdr recognizes.

`Checked 2026-10-17 against herdr 0.9.5`: this stamp applies to these claims:

- the result types
- the `number` values after moves and closes
- the automatic labels after a move, with no `tab_renamed` event
- one focused tab in each list
- `unknown` for plain-shell tabs, and `blocked` at an approval UI that herdr recognized

A pass saw the value `working` in a read-only list of a live session.

### PaneInfo (embedded in `tab.create` result)

`PaneInfo` is the `root_pane` of a new tab. [../data-model.md](../data-model.md#paneinfo) gives the full entity. When a nullable field with "no" in the required column has no value, herdr does not write the field. It does not write an explicit `null`. A new root pane usually has only these fields: `pane_id`, `terminal_id`, `workspace_id`, `tab_id`, `focused`, `cwd`, `foreground_cwd`, `agent_status`, `scroll`, and `revision`. A later `pane.get` can also show `terminal_title` and `terminal_title_stripped` after the shell sets a title.

`Checked 2026-10-17 against herdr 0.9.5`: the omitted null fields and the field set of a new root pane.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the top-level fields in this table, and the `PaneScrollInfo` and `AgentSessionInfo` shapes.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `pane_id` | string | yes | — | Opaque pane ID, qualified by the workspace (for example, `w1:p2`). |
| `terminal_id` | string | yes | — | Terminal ID of the pane (for example, `term_65e0dc1ce45a127`). |
| `workspace_id` | string | yes | — | ID of the workspace that contains the pane. |
| `tab_id` | string | yes | — | ID of the tab that contains the pane. |
| `focused` | boolean | yes | — | True when the pane has the focus. |
| `agent_status` | enum | yes | — | One of `idle`, `working`, `blocked`, `done`, `unknown`. |
| `revision` | integer (uint64, ≥0) | yes | — | Revision counter of the pane. The value only increases. |
| `agent` | string \| null | no | — | Agent kind that herdr detected, or null. |
| `agent_session` | AgentSessionInfo \| null | no | — | Agent session reference, or null. |
| `cwd` | string \| null | no | — | Working directory of the pane, or null. |
| `foreground_cwd` | string \| null | no | — | Working directory of the foreground process, or null. |
| `display_agent` | string \| null | no | — | Display name of the agent, or null. |
| `label` | string \| null | no | — | Pane label, or null. |
| `title` | string \| null | no | — | Pane title, or null. |
| `terminal_title` | string \| null | no | — | Raw terminal title, or null. |
| `terminal_title_stripped` | string \| null | no | — | Terminal title without control sequences, or null. |
| `scroll` | PaneScrollInfo \| null | no | — | Scrollback position data, or null. |
| `state_labels` | object (string→string) | no | — | Map of state labels with any keys. |
| `tokens` | object (string→string) | no | — | Map of tokens. It has 32 entries or fewer. The keys match `^[A-Za-z0-9_-]{1,32}$`. |

`PaneScrollInfo` is `{ offset_from_bottom, max_offset_from_bottom, viewport_rows }`. Each field is a required integer (uint64, ≥0). `AgentSessionInfo` is `{ source, agent, kind, value }`. Each of the four fields is a string, and the schema requires each field. The `kind` field is the enum `id` | `path`.

## tab.close

Close a tab and the panes in it. herdr sends a `tab_closed` event for the tab. It sends no `pane_closed` event for the panes in the tab. herdr does not use the tab ID again.

When you close the last tab in a workspace, herdr also closes the workspace. A `workspace_closed` event comes in the same delivery tick as `tab_closed`, and before it. After the close, a `tab.list` or `workspace.get` for that workspace returns `workspace_not_found`.

When you close the focused tab, herdr moves the focus to an adjacent tab. It sends no `tab_focused` event for this change. Thus, a client that follows the focus only from events has incorrect focus data.

Do not close tabs that you did not create, unless a person tells you to close them. This rule is herdr safety guidance. A CLI or API close does not use the TUI last-pane confirmation. [pane.close](pane.md#paneclose) gives this rule.

`Checked 2026-10-17 against herdr 0.9.5`: this stamp applies to these claims:

- the closed panes, with no `pane_closed` event
- the last-tab close, the event order, and `workspace_not_found` after the close
- the focus move, with no `tab_focused` event

With no TUI client attached, a close of the last tab returned `ok` with no prompt.

`Documented by herdr --skill (herdr 0.9.5); not live-checked: a pass cannot attach a TUI client to a scratch server safely`. This stamp applies to the safety guidance and to a close with a TUI client attached.

**Params** — `TabTarget`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `tab_id` | string | yes | — | ID of the tab to close. |

**Result** — `type: "ok"`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `type` | const `"ok"` | yes | — | Success value. The result has no other data. |

**Errors**: `tab_not_found` (unknown `tab_id`). Other codes are possible. Refer to [errors.md](../errors.md#observed-error-codes).

**Events**: herdr sends a `tab_closed` event to subscribers (subscription type `tab.closed`). The event data is `{tab_id, workspace_id}`.

**CLI**: `herdr tab close <tab_id>`

**Example** — `Checked 2026-10-17 against herdr 0.9.5.`

```json
{"id":"7","method":"tab.close","params":{"tab_id":"w1:t4"}}
{"id":"7","result":{"type":"ok"}}
```

## tab.create

Create a tab in a workspace, with its root pane. The result contains the new `tab` and its `root_pane`. Use `root_pane.pane_id` and `tab.tab_id` for the next operations. This rule is herdr guidance.

- When `workspace_id` is null, herdr uses the current or target workspace.
- When `focus` is true, herdr focuses the new tab. This action also moves the focus to the workspace of that tab. Refer to [tab.focus](#tabfocus).
- When `cwd` is null, herdr uses its `[terminal] new_cwd` setting. The default value is `"follow"`. With `"follow"`, the new tab got the working directory of the active pane in the target workspace. It did not get the working directory of the caller.
- A `cwd` that does not exist does not cause an error. The root pane process uses the home directory instead. A pass saw this with the default `new_cwd` setting.

`Checked 2026-10-17 against herdr 0.9.5`: this stamp applies to these claims:

- the `tab` and `root_pane` result, the null `workspace_id`, and the focus move
- the `cwd` from the active pane
- the home directory for a `cwd` that does not exist
- `env`, and the automatic and explicit labels

`Documented by herdr --default-config (herdr 0.9.5); not live-checked: the pass used only the default value`. This stamp applies to the `new_cwd` values other than `"follow"`. These values are `"home"` for `$HOME`, `"current"` for the process directory of herdr, or a fixed path.

**Params** — `TabCreateParams`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `workspace_id` | string \| null | no | null | Workspace for the new tab. Null selects the current workspace. |
| `cwd` | string \| null | no | null | Working directory for the process of the root pane. Null uses the `[terminal] new_cwd` setting. |
| `label` | string \| null | no | null | Initial tab label. Null gives an automatic label that follows the live position of the tab in its workspace. The automatic label is not the `number`. It can change with no `tab_renamed` event. herdr does not check an explicit label. It stores any length or content, with control characters, and returns the label exactly as the caller sent it. |
| `env` | object (string→string) | no | — | Environment variables for the process that herdr starts. |
| `focus` | boolean | no | `false` | True focuses the new tab when herdr creates it. |

**Result** — `type: "tab_created"`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `type` | const `"tab_created"` | yes | — | Result discriminator. |
| `tab` | TabInfo | yes | — | The new tab. Refer to [TabInfo](#tabinfo). |
| `root_pane` | PaneInfo | yes | — | The root pane of the tab. Refer to [PaneInfo](#paneinfo-embedded-in-tabcreate-result). |

**Errors**: `workspace_not_found` (unknown `workspace_id`). When the session has no workspace, the message is "no active workspace". It does not name the missing ID. Do not compare the message text. Other codes are possible. Refer to [errors.md](../errors.md#observed-error-codes).

**Events**: herdr sends a `tab_created` event to subscribers (subscription type `tab.created`). The event `data` contains the same `tab` as the result. The new root pane also gives a `pane_created` event after `tab_created`. Then it gives `layout_updated` and, after the shell is ready, `pane_updated`. These two events can come in either order.

**CLI**: `herdr tab create [--workspace <workspace_id>] [--cwd PATH] [--label TEXT] [--env KEY=VALUE]... [--focus] [--no-focus]`

**Example** — `Checked 2026-10-17 against herdr 0.9.5.` No tab method gave a value to the optional `agent`, `agent_session`, `display_agent`, `label`, `title`, or state fields of `root_pane`. These fields come from the schema.

```json
{"id":"1","method":"tab.create","params":{"workspace_id":"w1","label":"build","cwd":"/home/user/project"}}
{"id":"1","result":{"type":"tab_created","tab":{"tab_id":"w1:t2","workspace_id":"w1","number":2,"label":"build","focused":false,"pane_count":1,"agent_status":"unknown"},"root_pane":{"pane_id":"w1:p2","terminal_id":"term_65e0dc1ce45a127","workspace_id":"w1","tab_id":"w1:t2","focused":false,"cwd":"/home/user/project","foreground_cwd":"/home/user/project","agent_status":"unknown","scroll":{"offset_from_bottom":0,"max_offset_from_bottom":0,"viewport_rows":40},"revision":0}}}
```

## tab.focus

Focus a tab in the herdr UI and mark it seen. The result contains the changed tab metadata.

- The focus marks the agent status of the tab seen. An unseen `done` tab or an unseen idle tab with background work becomes `idle`. CLI reads do not mark a tab seen.
- The focus on a tab also focuses its workspace. Only one tab in the session has the focus at one time.
- Thus, the `active_tab_id` of a workspace can identify a tab with `focused` set to `false`. This occurs when that workspace does not have the focus.

`Checked 2026-10-17 against herdr 0.9.5`: a tab with a real agent showed `done` after two `tab.get` calls and a `tab.list` call. `tab.focus` changed it to `idle`. A pass also checked the workspace focus and the `active_tab_id` behavior.

**Params** — `TabTarget`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `tab_id` | string | yes | — | ID of the tab to focus. |

**Result** — `type: "tab_info"`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `type` | const `"tab_info"` | yes | — | Result discriminator. |
| `tab` | TabInfo | yes | — | The focused tab, with `focused: true`. Refer to [TabInfo](#tabinfo). |

**Errors**: `tab_not_found` (unknown `tab_id`). Other codes are possible. Refer to [errors.md](../errors.md#observed-error-codes).

**Events**: herdr sends a `tab_focused` event to subscribers (subscription type `tab.focused`). The event data is `{tab_id, workspace_id}`. On herdr 0.9.3 and 0.9.5, the same tick also contains `workspace_focused` (`{workspace_id}`) and `pane_focused` (`{pane_id, workspace_id}`). The order is `workspace_focused`, `tab_focused`, `pane_focused`. herdr sends `workspace_focused` also when the workspace has the focus before the call. A focus on the tab that has the focus changes nothing and sends no event. [events.md](../events.md) gives the event catalog.

`Checked 2026-10-17 against herdr 0.9.5`: the three events, their order, and no event for a second focus on the same tab.

**CLI**: `herdr tab focus <tab_id>`

**Example** — `Checked 2026-10-17 against herdr 0.9.5.` No pass tested a race between this call and a person who focuses a different tab in a second UI client.

```json
{"id":"6","method":"tab.focus","params":{"tab_id":"w1:t2"}}
{"id":"6","result":{"type":"tab_info","tab":{"tab_id":"w1:t2","workspace_id":"w1","number":2,"label":"reviewer","focused":true,"pane_count":1,"agent_status":"unknown"}}}
```

## tab.get

Get the metadata of one tab by ID. This method is a read. It does not mark the tab seen.

**Params** — `TabTarget`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `tab_id` | string | yes | — | ID of the tab to get. |

**Result** — `type: "tab_info"`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `type` | const `"tab_info"` | yes | — | Result discriminator. |
| `tab` | TabInfo | yes | — | The tab that the request identifies. Refer to [TabInfo](#tabinfo). |

**Errors**: `tab_not_found` (unknown `tab_id`). Other codes are possible. Refer to [errors.md](../errors.md#observed-error-codes).

**CLI**: `herdr tab get <tab_id>`

**Example** — `Checked 2026-10-17 against herdr 0.9.5, with tab_not_found and no change to a done status.`

```json
{"id":"3","method":"tab.get","params":{"tab_id":"w1:t1"}}
{"id":"3","result":{"type":"tab_info","tab":{"tab_id":"w1:t1","workspace_id":"w1","number":1,"label":"1","focused":false,"pane_count":1,"agent_status":"unknown"}}}
```

## tab.list

List tabs. When `workspace_id` is null, the result contains all tabs. When `workspace_id` identifies a workspace, the result contains only the tabs of that workspace. This method is a read. It does not mark tabs seen.

**Params** — `TabListParams`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `workspace_id` | string \| null | no | null | Workspace that limits the list. Null lists the tabs of all workspaces. |

**Result** — `type: "tab_list"`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `type` | const `"tab_list"` | yes | — | Result discriminator. |
| `tabs` | array of TabInfo | yes | — | Tabs in workspace order. Refer to [TabInfo](#tabinfo). |

**Errors**: `workspace_not_found` (unknown `workspace_id`). Other codes are possible. Refer to [errors.md](../errors.md#observed-error-codes).

**CLI**: `herdr tab list [--workspace <workspace_id>]`

**Example** — `Checked 2026-10-17 against herdr 0.9.5, with workspace_not_found, the list of all workspaces, and the order after a move.`

```json
{"id":"2","method":"tab.list","params":{"workspace_id":"w1"}}
{"id":"2","result":{"type":"tab_list","tabs":[{"tab_id":"w1:t1","workspace_id":"w1","number":1,"label":"1","focused":false,"pane_count":1,"agent_status":"unknown"},{"tab_id":"w1:t2","workspace_id":"w1","number":2,"label":"build","focused":false,"pane_count":1,"agent_status":"unknown"},{"tab_id":"w1:t3","workspace_id":"w1","number":3,"label":"3","focused":false,"pane_count":1,"agent_status":"unknown"},{"tab_id":"w1:t4","workspace_id":"w1","number":4,"label":"4","focused":false,"pane_count":1,"agent_status":"unknown"}]}}
```

## tab.move

Move a tab to a target index in its workspace. The result contains the full tab list of the workspace in the new order. The result type is `tab_list`, not `ok`. A move does not change `number` values. Refer to [TabInfo](#tabinfo).

herdr applies `insert_index` to the tab order **before** the move.

1. herdr puts the tab before the tab that has that index now.
2. Then herdr removes the tab from its old position.

Thus, a forward move puts the tab one position earlier than a remove-then-insert operation. In a 28-tab workspace, a move of the first tab (index 0) to `insert_index: 8` put it at index 7. A move of that tab to `insert_index: 1` did not change the order.

The range of accepted values is `0..=tab_count`, with the two limits included. A larger value does not change the order. It returns the error code `tab_move_failed`. Refer to Errors below. A move that does not change the order sends no `tab_moved` event.

An automatic `label` follows the live position of the tab. Thus, a move can change the automatic labels of the moved tab and of other tabs, with no rename. In the example below, `w1:t1` keeps `number: 1`. Its label becomes `"3"` because three tabs are now before it.

**Params** — `TabMoveParams`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `tab_id` | string | yes | — | ID of the tab to move. |
| `insert_index` | integer (uint, ≥0) | yes | — | Target position, counted from 0. herdr applies it to the tab order of the workspace **before** the move. Valid range: `0..=tab_count`. |

**Result** — `type: "tab_list"`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `type` | const `"tab_list"` | yes | — | Result discriminator. |
| `tabs` | array of TabInfo | yes | — | The tabs of the workspace in their new order. Refer to [TabInfo](#tabinfo). |

**Errors**:

- `tab_not_found` (unknown `tab_id`).
- `tab_move_failed` (`insert_index` is more than `tab_count`). The schema has no list of error codes. Thus, the schema does not show this code or any other code on this page.
- Other codes are possible. Refer to [errors.md](../errors.md#observed-error-codes).

**Events**: herdr sends a `tab_moved` event to subscribers (subscription type `tab.moved`). The event data is `{tab_id, workspace_id, insert_index, tabs}`. The `tabs` field is the same list as the RPC result. A move that does not change the order sends no event.

**CLI**: API only. No CLI subcommand calls this method. The `herdr tab` command has only `list`, `create`, `get`, `focus`, `rename`, and `close`. `herdr tab move` writes the usage list and exits with status `2`.

**Example** — `Checked 2026-10-17 against herdr 0.9.5 (insert_index 8 from index 0 put the tab at index 7. insert_index 1 from index 0 did not change the order and sent no event. herdr accepted tab_count. tab_count + 1 gave tab_move_failed "insert_index 29 is out of bounds".)`

```json
{"id":"4","method":"tab.move","params":{"tab_id":"w1:t1","insert_index":3}}
{"id":"4","result":{"type":"tab_list","tabs":[{"tab_id":"w1:t2","workspace_id":"w1","number":2,"label":"build","focused":false,"pane_count":1,"agent_status":"unknown"},{"tab_id":"w1:t3","workspace_id":"w1","number":3,"label":"2","focused":false,"pane_count":1,"agent_status":"unknown"},{"tab_id":"w1:t1","workspace_id":"w1","number":1,"label":"3","focused":false,"pane_count":1,"agent_status":"unknown"},{"tab_id":"w1:t4","workspace_id":"w1","number":4,"label":"4","focused":false,"pane_count":1,"agent_status":"unknown"}]}}
```

## tab.rename

Change the label of a tab. The result contains the changed tab metadata.

**Params** — `TabRenameParams`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `tab_id` | string | yes | — | ID of the tab to rename. |
| `label` | string | yes | — | New label for the tab. herdr does not check it. It stores any length or content, with control characters, and returns the label exactly as the caller sent it. |

**Result** — `type: "tab_info"`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `type` | const `"tab_info"` | yes | — | Result discriminator. |
| `tab` | TabInfo | yes | — | The renamed tab, with its new `label`. Refer to [TabInfo](#tabinfo). |

**Errors**: `tab_not_found` (unknown `tab_id`). Other codes are possible. Refer to [errors.md](../errors.md#observed-error-codes).

**Events**: herdr sends a `tab_renamed` event to subscribers (subscription type `tab.renamed`). The event data is `{tab_id, workspace_id, label}`. `tab.focus` and `tab.move` send no event when nothing changes. But a rename to the same label sends the event.

**CLI**: `herdr tab rename <tab_id> <label>...`. The CLI joins all `LABEL` words into one label string. Words with the form of a flag, and `--`, also become part of the label.

**Example** — `Checked 2026-10-17 against herdr 0.9.5 (herdr stored control characters and a 303-character label exactly as sent. A rename to the same label sent tab_renamed.)`

```json
{"id":"5","method":"tab.rename","params":{"tab_id":"w1:t2","label":"reviewer"}}
{"id":"5","result":{"type":"tab_info","tab":{"tab_id":"w1:t2","workspace_id":"w1","number":2,"label":"reviewer","focused":false,"pane_count":1,"agent_status":"unknown"}}}
```
