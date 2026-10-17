# herdr API: pane methods

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Index: [README.md](../README.md). Wire format: [protocol.md](../protocol.md).

The `pane.*` methods control the panes of herdr. A pane is one terminal area in the layout of a tab. The methods do these tasks:

- Read the topology and the geometry of panes: `list`, `get`, `current`, `layout`, `edges`, `neighbor`, and `process_info`.
- Change the layout: `split`, `swap`, `move`, `resize`, `zoom`, `close`, `focus`, and `focus_direction`.
- Operate the terminal of a pane: `read`, `wait_for_output`, `send_text`, `send_keys`, `send_input`, `rename`, and `input.set`.
- Clear, scroll, and copy the content of a pane: `clear`, `scroll`, `edit_scrollback`, `copy_motion`, `copy_search`, and `selection.read`.
- Find or open links in a pane: `link.resolve` and `link.activate`.
- Let an integration report or withdraw lifecycle data and display data for the agent of a pane: `report_agent`, `report_agent_session`, `report_metadata`, `release_agent`, and `clear_agent_authority`.

A pane exists with or without an agent. Use the `agent.*` methods when herdr must check the identity of an agent or interpret its agent status. Refer to [agent.md](agent.md).

A pane ID includes its workspace, for example `w1:p1`. [addressing.md](../addressing.md#id-scheme) gives the ID scheme.

Pane targets:

- Nine methods accept a null or omitted `pane_id`, and then use the UI-focused pane. For `pane.split`, the field is `target_pane_id`.
- Four of these methods read the topology: `layout`, `edges`, `neighbor`, and `process_info`. Five of them change the arrangement: `focus_direction`, `resize`, `split`, `swap`, and `zoom`.
- The CLI commands for these nine methods accept `[--current | --pane <ID>]`. For `pane focus` and `pane neighbor`, `--current` uses the UI-focused pane, not the calling pane. Refer to [pane.focus_direction](#panefocus_direction) and [pane.neighbor](#paneneighbor).
- The other methods that take a pane need an explicit `pane_id` string.
- `pane.current` finds the pane of the caller from `caller_pane_id`.
- The UI-focused pane can belong to another person or client. Thus give your own `$HERDR_PANE_ID` as the target. [addressing.md](../addressing.md#resolving---current-and-caller-context) gives the rules for `--current` and the caller context.

[data-model.md](../data-model.md) defines the domain entities one time, such as `PaneInfo`, `AgentInfo`, `PaneLayoutSnapshot`, `WorkspaceInfo`, and `TabInfo`. The result tables on this page name each field and link there. They do not expand the embedded entities again.

Nine methods put their result fields one level down, in an object with the name of the method. These methods and objects are `edges` (`edges`), `focus_direction` (`focus`), `move` (`move_result`), `neighbor` (`neighbor`), `process_info` (`process_info`), `read` (`read`), `resize` (`resize`), `swap` (`swap`), and `zoom` (`zoom`). The table of each of these sections describes that nested object, as its example shows.

Request rules:

- The message text of `pane_not_found` is not stable. Some methods send `"pane w9:p9 not found"`. Other methods send `"pane not found"`, `"source pane not found"`, or `"pane not found: w1:p99"`. Use `code` to identify the error, not `message`.
- Each request needs `params`, also when each field in it is optional. A request with no `params` fails with `invalid_request` (``"missing field `params`"``).
- When a result field that the schema makes nullable has no value, herdr usually omits the field. It does not send `null`. Treat a missing field and `null` the same, unless a section tells you otherwise.

A pane that a restarted server cannot restore has `restore_error` and is a placeholder. `pane.get` and `pane.list` show it. `pane.read`, `pane.process_info`, and `pane.send_text` fail on it with `pane_not_found`. [PaneInfo](../data-model.md#paneinfo) owns this claim and its stamp. This page did not check it.

35 methods. herdr 0.9.3 removed the three `pane.graphics.*` methods (`set`, `clear`, and `info`) that this page documented for 0.9.1. herdr 0.9.3 also removed the undocumented `pane.graphics.stream`, which the 0.9.1 binary accepted. All four names now fail with `unknown_method` (`"unknown method: pane.graphics.set"`), and the response echoes the request `id`. Other unknown method names fail with an `invalid_request` "unknown variant" error. Refer to [errors.md](../errors.md).

Checked 2026-10-17 against herdr 0.9.5 on a scratch server: the nine null targets, the CLI target flags, the explicit `pane_id` rule, and the nested results. The check also included the four `pane_not_found` messages, the `params` rule, the omitted fields, and the four `pane.graphics.*` errors.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the count of 35 methods, the method table, and the enum values below.

| method | purpose |
|---|---|
| [pane.clear](#paneclear) | Clear the screen and the scrollback of a pane. |
| [pane.clear_agent_authority](#paneclear_agent_authority) | Withdraw the authority of a source over the agent reports of a pane. |
| [pane.close](#paneclose) | Close a pane and its terminal. |
| [pane.copy_motion](#panecopy_motion) | Move a copy-mode cursor by a text motion. |
| [pane.copy_search](#panecopy_search) | Search the text of a pane for copy-mode match navigation. |
| [pane.current](#panecurrent) | Return the pane that the caller or the UI is in now. |
| [pane.edges](#paneedges) | Tell which of the four edges of a pane touch the tab boundary. |
| [pane.edit_scrollback](#paneedit_scrollback) | Open the scrollback of a pane in an external editor. |
| [pane.focus](#panefocus) | Focus one pane by ID and return its pane data. |
| [pane.focus_direction](#panefocus_direction) | Move the focus to the next pane in a direction. |
| [pane.get](#paneget) | Get the `PaneInfo` of one pane. |
| [pane.input.set](#paneinputset) | Set the right-click input routing of a pane. |
| [pane.layout](#panelayout) | Return the layout snapshot of the tab of a pane. |
| [pane.link.activate](#panelinkactivate) | Activate a detected link at a viewport position of a pane. |
| [pane.link.resolve](#panelinkresolve) | Find the link regions at a viewport position of a pane. |
| [pane.list](#panelist) | List panes. You can limit the list to one workspace. |
| [pane.move](#panemove) | Move a pane to another tab, a new tab, or a new workspace. |
| [pane.neighbor](#paneneighbor) | Find the ID of the next pane in a direction. |
| [pane.process_info](#paneprocess_info) | Return the shell and the foreground processes of a pane. |
| [pane.read](#paneread) | Read a snapshot of the terminal output of a pane. |
| [pane.release_agent](#panerelease_agent) | Release the agent authority of a source for one agent. |
| [pane.rename](#panerename) | Set or clear the label of a pane. |
| [pane.report_agent](#panereport_agent) | Report the lifecycle state of the agent in a pane. |
| [pane.report_agent_session](#panereport_agent_session) | Report the agent session identity for a pane. |
| [pane.report_metadata](#panereport_metadata) | Report display metadata for a pane, such as a title, tokens, and labels. |
| [pane.resize](#paneresize) | Resize the split that contains a pane. |
| [pane.scroll](#panescroll) | Set the scrollback offset of a pane from the bottom. |
| [pane.selection.read](#paneselectionread) | Read the text in a selection range of a pane. |
| [pane.send_input](#panesend_input) | Send text, logical keys, or both to a pane in one call. |
| [pane.send_keys](#panesend_keys) | Send logical key presses to a pane. |
| [pane.send_text](#panesend_text) | Send literal text to a pane. |
| [pane.split](#panesplit) | Split a pane and make a new sibling pane. |
| [pane.swap](#paneswap) | Swap the positions of two panes in the layout. |
| [pane.wait_for_output](#panewait_for_output) | Wait until the output of a pane matches a pattern. |
| [pane.zoom](#panezoom) | Toggle or set the zoom, which maximizes one pane in its tab. |

These enums occur in this namespace:

- **PaneDirection**: `left`, `right`, `up`, `down`.
- **SplitDirection**: `right`, `down`.
- **ReadSource**: `visible`, `recent`, `recent_unwrapped`, `detection`.
- **ReadFormat**: `text`, `ansi`.
- **PaneRightClickTarget**: `herdr`, `pane`.
- **PaneZoomMode**: `toggle`, `on`, `off`.
- **PaneAgentState**: `idle`, `working`, `blocked`, `unknown`. These are the values that a `pane.report_agent` call can report. A report of `done` fails with `invalid_request`.
- **PaneCopyMotion**: `line_end`, `first_non_blank`, `next_word_start`, `previous_word_start`, `next_word_end`, `next_big_word_start`, `previous_big_word_start`, `next_big_word_end`, `previous_paragraph`, `next_paragraph`.
- **PaneCopySearchDirection**: `forward`, `backward`.

`PaneInfo.agent_status` is a different enum, `AgentStatus`. It has a fifth value, `done`, that herdr sets from observation. No report can set `done` directly. An `idle` report after `working` or `blocked` shows as `done` while the tab of the pane is not focused. It stays `done` until the tab is focused or seen. Refer to [pane.report_agent](#panereport_agent). [AgentStatus](../data-model.md#agentstatus) gives the meaning of each value, and what `idle` and `done` tell about readiness.

Checked 2026-10-17 against herdr 0.9.5: the `done` report error, the `done` status for an unfocused tab, and the 10 copy motions.

The four read sources have these meanings:

- `visible` is the viewport that herdr renders now.
- `recent` is recent rendered output, with soft wraps.
- `recent_unwrapped` joins soft wraps. It is the better source for logs and transcripts.
- `detection` is the plain-text snapshot of the bottom buffer that herdr uses to detect agents.

`Documented by herdr --skill (herdr 0.9.5); not live-checked: the meanings are herdr text. A pass checked only that pane.read accepts all four values.`

A read through the API or the CLI does not mark the tab of an agent as seen. Only a focus marks it seen, in the UI or with a focus command.

Checked 2026-10-17 against herdr 0.9.5: an agent in `done` stayed `done` after an API `pane.read` and after a CLI `herdr pane read`. It changed to `idle` after `pane.focus`.

---

## pane.clear

New in 0.9.3. Clears the screen and the scrollback of a pane, as a terminal "clear and reset scrollback" command does.

- Only the line with the cursor stays. herdr moves it to the top row. Thus a shell prompt stays visible.
- Typed input on that line stays too, also when nobody sent Enter for it.
- herdr writes nothing to the program of the pane. The pending input runs when Enter comes later. The foreground process does not change.
- The scroll state goes back to `offset_from_bottom: 0` and `max_offset_from_bottom: 0`, also when the pane was scrolled back.
- The target pane does not need the focus. It does not need to be in the focused tab.

When the scroll state changes, herdr emits `pane.scroll_changed` with the reset `PaneScrollInfo`. herdr emits no plain event, such as `pane_updated` or `layout_updated`. `PaneInfo.revision` does not change. A clear of a pane with no scroll offset and no scrollback emits no event.

The UI equivalent is the keybinding `clear_pane`. It has no default key in [raw/default-config.toml](../raw/default-config.toml).

**Params** (`PaneTarget`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Pane to clear. herdr ignores unknown extra keys in `params`. |

**Result**: `type: "ok"`, with no other fields.

**Errors**: `pane_not_found` (`"pane w1:p99 not found"`). A missing `pane_id` fails with `invalid_request`. Other codes are possible.

**CLI**: API only. There is no CLI subcommand. `herdr pane clear` prints the usage list of `pane` and exits with status 2.

**Example**

```json
{"id":"r4","method":"pane.clear","params":{"pane_id":"w1:p1"}}
{"id":"r4","result":{"type":"ok"}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. A pane had 2961 rows of scrollback and a scroll offset of 50 rows. After the clear, each `source` read back only its prompt line. A pending `echo PENDING` stayed and ran on Enter. The foreground process kept its PID. A pane in a tab that was not focused also read back as clear. The pass saw the events with a subscription to each plain event, `pane.scroll_changed`, and `pane.agent_status_changed`. The `clear_pane` line came from `herdr --default-config`.

---

## pane.clear_agent_authority

Withdraws the authority that a reporting `source` has over the agent lifecycle reports of a pane. The call does not name an agent. Use it when an integration stops all management of a pane. To release one named agent, use [pane.release_agent](#panerelease_agent).

- The method is idempotent. It returns `ok` also when the source has no authority.
- When the call clears a claim, herdr emits `pane_agent_detected` with no `agent`. It also emits `pane.agent_status_changed` with `agent_status` `"unknown"`.
- A clear that changes nothing emits no event.
- A named `source` that does not hold the current claim of the pane clears nothing. The claim of the other source stays.

**Params** (`PaneClearAgentAuthorityParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Target pane ID. |
| `source` | string \| null | no | null | Reporting source whose authority herdr clears. A null or omitted `source` is a wildcard. It clears the source that has the authority now. It is not a separate "ambient" or "default" source. |
| `seq` | integer (uint64) \| null | no | null | Monotonic sequence number that orders reports from one source (inferred). herdr applies a clear only when its `seq` is strictly greater than the last `seq` of that source. After a report with `seq`, a clear with an equal, lower, null, or omitted `seq` returns `ok` and changes nothing. A wildcard clear obeys the same rule. |

**Result**: `type: "ok"`, with no other fields.

**Errors**: `pane_not_found` for an unknown `pane_id`. Other codes are possible.

**CLI**: API only. There is no CLI subcommand. `herdr pane clear-agent-authority` prints the usage list and exits with status 2. The CLI has `release-agent`, but no command that clears all authority.

**Example**

```json
{"id":"p6","method":"pane.clear_agent_authority","params":{"pane_id":"w1:p1","source":"docprobe"}}
{"id":"p6","result":{"type":"ok"}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass ran the example again. It checked the `seq` rule with an equal, a lower, a null, an omitted, and a greater `seq`. It also checked a null-`source` wildcard clear after a report with `seq`. It saw the events for a clear that changed the claim, and no events for a clear that changed nothing. It checked the clear from a source that did not hold the claim.

---

## pane.close

Closes a pane and ends its terminal process.

- Do not close panes that you did not create, unless a person tells you to close them.
- herdr does not use a closed pane ID again.
- A close of the last pane of a workspace also closes the workspace.
- A close emits `pane_closed` and `layout_updated`. A close of the last pane of a workspace emits `workspace_closed` and `pane_closed`. herdr pushes these events in the order of the `subscriptions` array. Refer to [ordering and grouping](../events.md#ordering-and-grouping).

A CLI or API close does not use the TUI last-pane confirmation. In the TUI, the `close_pane` action on the last pane of a workspace opens a `Close workspace?` confirmation. A `herdr pane close` command or a `pane.close` request closes that pane and its workspace at once, with no confirmation. The `[ui] confirm_close` setting controls the TUI confirmation, and its default is `true`. The pages for tab and workspace closes link to this rule.

**Params** (`PaneTarget`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Pane to close. |

**Result**: `type: "ok"`, with no other fields.

**Errors**: `pane_not_found`. Other codes are possible.

**CLI**: `herdr pane close <pane_id>`

**Example**

```json
{"id":"cli:pane:close","method":"pane.close","params":{"pane_id":"w1:p2"}}
{"id":"cli:pane:close","result":{"type":"ok"}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass closed a pane and saw that later splits did not use its ID again. It closed the last pane of a workspace, and the workspace closed. It saw the events of each close. For the confirmation rule, a TUI client was attached in a pseudo-terminal, and a single-pane workspace was focused. The TUI `close_pane` key opened the `Close workspace?` confirmation, and Esc kept the pane. Then `herdr pane close` on the same pane printed `ok` and exited with status 0. The client showed no confirmation, and the pane and its workspace were gone.

`Documented by herdr --skill (herdr 0.9.5); not live-checked: the rule about panes that you did not create is guidance, not behavior.`

`Documented by herdr --default-config (herdr 0.9.5); not live-checked: the pass did not set confirm_close = false.`

---

## pane.copy_motion

Moves a copy-mode cursor from a given position by a text motion, and returns the new position. The motions are word, line, and paragraph steps in the vi style. The method does not read or select text. To get the text of a span, use [pane.selection.read](#paneselectionread).

**Params** (`PaneCopyMotionParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Target pane. |
| `cursor` | PaneTextPoint (below) | yes | — | Copy-mode cursor position to move from. |
| `motion` | PaneCopyMotion (`line_end`,`first_non_blank`,`next_word_start`,`previous_word_start`,`next_word_end`,`next_big_word_start`,`previous_big_word_start`,`next_big_word_end`,`previous_paragraph`,`next_paragraph`) | yes | — | Motion to apply. |
| `content_revision` | integer (uint64) \| null | no | null | Content revision for the cursor position. Null skips the check (inferred). |

`PaneTextPoint` is a zero-based row and column position in the text content of a pane. `row` counts from the top of the scrollback, not from the top of the viewport. [pane.copy_search](#panecopy_search) and [pane.selection.read](#paneselectionread) also use it. [pane.link.resolve](#panelinkresolve) and [pane.link.activate](#panelinkactivate) use a different row, which counts from the top of the viewport.

| field | type | required | default | meaning |
|---|---|---|---|---|
| `row` | integer (uint32) | yes | — | Zero-based row, from the top of the scrollback. |
| `col` | integer (uint16) | yes | — | Zero-based column. |

A `cursor` out of range is not an error. herdr returns the same position, and nothing tells that the motion had no text to act on.

**Result**: `type: "pane_copy_motion"`

| field | type | meaning |
|---|---|---|
| `pane_id` | string | Target pane. |
| `cursor` | PaneTextPoint | Cursor position after the motion. |
| `content_revision` | integer (uint64) | Content revision for the returned cursor. This value is a private copy-mode counter. It is not the `revision` of `pane.read`, which is always 0. It is also not `PaneInfo.revision`, which is a third counter. In this namespace, only an earlier `copy_motion` or `copy_search` result gives a valid `content_revision`. |

**Errors**: `pane_not_found`. A `content_revision` that is not equal to the current copy-mode counter of the pane fails with the undocumented `stale_content` (`"pane content changed"`). This applies to an older, a newer, or an arbitrary value. Other codes are possible.

**CLI**: API only. There is no CLI subcommand.

**Example**

The text of the pane is `$ echo docprobe`, `docprobe`, `$ echo DocProbe docprobe`, `DocProbe docprobe`, and `$`, one item for each row.

```json
{"id":"cm1","method":"pane.copy_motion","params":{"pane_id":"w1:p4","cursor":{"row":0,"col":0},"motion":"next_word_start"}}
{"id":"cm1","result":{"type":"pane_copy_motion","pane_id":"w1:p4","cursor":{"row":0,"col":2},"content_revision":44}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass applied all 10 motions from one position. It checked a scrollback row far above the viewport, a cursor out of range, and the three counters on the same pane. It sent `content_revision` values that were older, newer, and arbitrary. The example is a new capture.

---

## pane.copy_search

Searches the text content of a pane for copy-mode match navigation, such as find next and find previous. The search starts at a cursor position. A repeat search continues after a `previous` match range.

- The search covers all the scrollback, not only the viewport.
- The search is case-sensitive.
- herdr accepts an empty `query` and returns zero matches, not an error.

**Params** (`PaneCopySearchParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Target pane. |
| `query` | string | yes | — | Text to search for. |
| `direction` | PaneCopySearchDirection (`forward`,`backward`) | yes | — | Search direction. |
| `cursor` | PaneTextPoint (above) | yes | — | Position to search from. |
| `content_revision` | integer (uint64) | yes | — | Content revision for the search. |
| `previous` | PaneTextRange \| null | no | null | Earlier match range. The search continues after it (inferred). |

`PaneTextRange` occurs in the params and in the result of this method:

| field | type | required | default | meaning |
|---|---|---|---|---|
| `start` | PaneTextPoint (above) | yes | — | Start point of the range. |
| `end` | PaneTextPoint (above) | yes | — | End point of the range. |

**Result**: `type: "pane_copy_search"`

| field | type | meaning |
|---|---|---|
| `pane_id` | string | Pane that herdr searched. |
| `content_revision` | integer (uint64) | Content revision for the results. |
| `matches` | `array<PaneTextRange>` | All match ranges. |
| `total` | integer (uint64) | Count of `matches`. |
| `current` | integer (uint32) \| null | Index in `matches` of the current match. herdr omits it, and does not send `null`, when there is no current match (for example, `total: 0`). |
| `current_global` | integer (uint64) \| null | Global match index across the search history (inferred). herdr omits it under the same condition as `current`. In each probe it was equal to `current`, also in a repeat search with `previous`. |

**Errors**: `pane_not_found`. A `content_revision` that is not equal to the current copy-mode counter of the pane fails with the undocumented `stale_content` (`"pane content changed"`). A missing `content_revision` fails with `invalid_request`. `content_revision` is required here, and only an earlier `copy_motion` or `copy_search` result gives a valid value. Refer to [pane.copy_motion](#panecopy_motion). Thus a first call usually fails with `stale_content`. Other codes are possible.

**CLI**: API only. There is no CLI subcommand.

**Example**

This capture uses the pane of the [pane.copy_motion](#panecopy_motion) example, with the same content revision.

```json
{"id":"cs1","method":"pane.copy_search","params":{"pane_id":"w1:p4","query":"docprobe","direction":"forward","cursor":{"row":0,"col":0},"content_revision":44}}
{"id":"cs1","result":{"type":"pane_copy_search","pane_id":"w1:p4","content_revision":44,"matches":[{"start":{"row":0,"col":7},"end":{"row":0,"col":14}},{"start":{"row":1,"col":0},"end":{"row":1,"col":7}},{"start":{"row":2,"col":7},"end":{"row":2,"col":14}},{"start":{"row":2,"col":16},"end":{"row":2,"col":23}},{"start":{"row":3,"col":0},"end":{"row":3,"col":7}},{"start":{"row":3,"col":9},"end":{"row":3,"col":16}}],"total":6,"current":0,"current_global":0}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked a case-sensitive query, an empty query, a match 264 rows above the viewport, and a repeat search with `previous`. It checked the omitted `current` fields and each error. The example is a new capture.

---

## pane.current

Returns the pane that the caller is in now. With no caller context, it returns the UI-focused pane. To find your own pane, give your pane ID as `caller_pane_id`.

**Params** (`PaneCurrentParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `caller_pane_id` | string \| null | no | null | Pane ID that herdr uses as the pane of the caller. When it has a value, herdr returns that pane and not the UI-focused pane (inferred from the CLI `--current` and `--pane`). |

**Result**: `type: "pane_current"`

| field | type | meaning |
|---|---|---|
| `pane` | [PaneInfo](../data-model.md#paneinfo) | The current pane. |

**Errors**: `pane_not_found` (`"pane not found"`). Other codes are possible.

**CLI**: `herdr pane current [--current | --pane <ID>]`

**Example**

The pane runs a `claude` agent that waits at its folder-trust dialog. Thus its status is `blocked`. The pane has no `agent_session`, because the hook of the agent did not run. When an integration reports an agent session, `PaneInfo` also has `agent_session`. Refer to [AgentSessionInfo](../data-model.md#agentsessioninfo). The response is the line that `herdr pane current --pane w1:p5` printed.

```json
{"id":"cli:pane:current","method":"pane.current","params":{"caller_pane_id":"w1:p5"}}
{"id":"cli:pane:current","result":{"pane":{"agent":"claude","agent_status":"blocked","cwd":"/home/user/project","focused":false,"foreground_cwd":"/home/user/project","pane_id":"w1:p5","revision":1,"scroll":{"max_offset_from_bottom":0,"offset_from_bottom":0,"viewport_rows":40},"tab_id":"w1:t3","terminal_id":"term_65e0dd8837c4a7","terminal_title":"user@host:~/project","terminal_title_stripped":"user@host:~/project","workspace_id":"w1"},"type":"pane_current"}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked empty params, a null `caller_pane_id`, a known `caller_pane_id`, and an unknown one. It also ran the CLI with `--pane`, with `--current`, and with no flag. The example is a new capture. A read-only `pane.list` on a live session showed the `agent_session` key on a pane with a reported `claude` agent.

---

## pane.edges

Tells if each edge of a pane is on the outer boundary of the tab (`true`) or next to another pane (`false`). Use it to know if a directional focus or move leaves the tab. The method also returns the full layout snapshot of the tab.

**Params** (`PaneEdgesParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string \| null | no | null | Pane to examine. Null uses the UI-focused pane. |

**Result**: `type: "pane_edges"`. The fields below are one level down, in `edges`:

| field | type | meaning |
|---|---|---|
| `edges` | PaneEdgesResult (below) | Edge data and layout data. |

`PaneEdgesResult`:

| field | type | meaning |
|---|---|---|
| `pane_id` | string | Pane that herdr examined. |
| `left` | boolean | The left edge is on the tab boundary. |
| `right` | boolean | The right edge is on the tab boundary. |
| `up` | boolean | The top edge is on the tab boundary. |
| `down` | boolean | The bottom edge is on the tab boundary. |
| `layout` | [PaneLayoutSnapshot](../data-model.md#panelayoutsnapshot) | The layout of the tab that contains the pane. |

**Errors**: `pane_not_found` (`"pane not found"`). Other codes are possible.

**CLI**: `herdr pane edges [--current | --pane <ID>]`

**Example**

```json
{"id":"cli:pane:edges","method":"pane.edges","params":{"pane_id":"w1:p5"}}
{"id":"cli:pane:edges","result":{"edges":{"down":true,"layout":{"area":{"height":40,"width":120,"x":0,"y":0},"focused_pane_id":"w1:p5","panes":[{"focused":true,"pane_id":"w1:p5","rect":{"height":40,"width":120,"x":0,"y":0}}],"splits":[],"tab_id":"w1:t3","workspace_id":"w1","zoomed":false},"left":true,"pane_id":"w1:p5","right":true,"up":true},"type":"pane_edges"}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server, with a single-pane tab and a two-pane tab. The pass also checked the CLI error for an unknown pane. The example is a new capture from a server with no TUI client.

---

## pane.edit_scrollback

Opens the scrollback of a pane in an external editor, to browse or copy it. The target must be the focused pane. Each call writes the scrollback to a temporary file with mode `0600`. Then herdr opens a real pane beside the target, focuses it, and runs this command in it:

```sh
/bin/sh -c 'scrollback_file=/tmp/herdr-scrollback-<server-pid>-<nonce>-<n>.txt; eval "${EDITOR:-vi} \"$scrollback_file\""; status=$?; rm -f "$scrollback_file"; exit $status'
```

- When `$EDITOR` is not set in the environment of the server, the command runs `vi`.
- When herdr opens the editor pane, it emits `workspace_focused`, `tab_focused`, and `pane_focused` for the editor pane. It emits no `pane_created` and no `layout_updated` at that time.
- The editor pane is an ordinary entry in `pane.list` while the editor runs.
- When the editor exits, the shell deletes the temporary file. The pane closes with `pane_exited`, then `layout_updated`, and no `pane_closed`. Then the focus goes back to the target pane.
- The editor pane uses one pane ID from the counter of the workspace. herdr does not use closed pane IDs again, so the visible pane numbers skip that ID.

**Params** (`PaneTarget`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Pane with the scrollback to edit. |

**Result**: `type: "ok"`, with no other fields.

**Errors**: `pane_not_found`. A `pane_id` that is not the focused pane fails with the undocumented `stale_pane_target` (`"pane is no longer focused"`). Focus the pane first. Other codes are possible.

**CLI**: API only. There is no CLI subcommand.

**Example**

```json
{"id":"e1","method":"pane.edit_scrollback","params":{"pane_id":"w1:p1"}}
{"id":"e1","result":{"type":"ok"}}
```

Checked 2026-10-17 against herdr 0.9.5 on two scratch servers. On one server, `EDITOR` was a script that waits for one line. The editor pane showed in `pane.list`, and `pane.process_info` showed the command above. The pass saw the events on open and on exit. After the exit, the temporary file did not exist, and the focus was on the target again. The next split skipped the ID of the editor pane. On the other server, `EDITOR` was not set. That host had no `vi` binary, so the editor pane exited at once. This result agrees with the `vi` fallback. The pass also checked the focus rule and the `pane_not_found` error. In 0.9.1, the same probe saw the new pane exit at once, and the pane did not show in the list.

---

## pane.focus

Focuses one pane by ID and returns its `PaneInfo`. Unlike [pane.focus_direction](#panefocus_direction), this method uses an exact pane, not a neighbor.

- A focus marks the agent of the pane, and its tab, as seen. Thus a background `done` changes back to `idle`.
- `tab.focus` on the tab of the pane has the same effect on `done`, whichever pane in the tab gets the focus. Refer to [tab.md](tab.md).
- The method emits `pane_focused`. It also emits `workspace_focused` and `tab_focused`, also when the workspace already had the focus. When `done` changes to `idle`, herdr also emits `pane.agent_status_changed`.

**Params** (`PaneTarget`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Pane to focus. |

**Result**: `type: "pane_info"`

| field | type | meaning |
|---|---|---|
| `pane` | [PaneInfo](../data-model.md#paneinfo) | The pane that now has the focus, with its IDs, cwd, focus state, and agent status. |

**Errors**: `pane_not_found`. Other codes are possible.

**CLI**: API only. There is no CLI subcommand. The CLI command `herdr pane focus` uses [pane.focus_direction](#panefocus_direction).

**Example**

```json
{"id":"1","method":"pane.focus","params":{"pane_id":"w1:p5"}}
{"id":"1","result":{"type":"pane_info","pane":{"pane_id":"w1:p5","terminal_id":"term_65e0dd8837c4a7","workspace_id":"w1","tab_id":"w1:t3","focused":true,"cwd":"/home/user/project","foreground_cwd":"/home/user/project","label":"reviewer","terminal_title":"user@host:~/project","terminal_title_stripped":"user@host:~/project","agent_status":"unknown","scroll":{"offset_from_bottom":0,"max_offset_from_bottom":0,"viewport_rows":40},"revision":3}}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked the result type, the events, and the change from `done` to `idle` through `pane.focus` and through `tab.focus`. The example is a new capture.

`Documented by herdr --skill (herdr 0.9.5); not live-checked: the skill also tells that a focus marks the target seen. The pass used the change from done to idle as the only sign of seen.`

---

## pane.focus_direction

Moves the keyboard focus from a pane to the next pane in a given direction. When there is no pane in that direction, `changed` is `false` and `reason` is `no_neighbor`. A focus change emits `pane_focused`, `tab_focused`, and `workspace_focused`. A call with `no_neighbor` emits no event.

**Params** (`PaneFocusDirectionParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `direction` | PaneDirection (`left`,`right`,`up`,`down`) | yes | — | Direction for the focus move. |
| `pane_id` | string \| null | no | null | Start pane. Null uses the UI-focused pane. |

**Result**: `type: "pane_focus_direction"`. The fields below are one level down, in `focus`:

| field | type | meaning |
|---|---|---|
| `focus` | PaneFocusDirectionResult (below) | Result of the focus move. |

`PaneFocusDirectionResult`:

| field | type | meaning |
|---|---|---|
| `changed` | boolean | True when the focus moved. |
| `source_pane_id` | string | Pane where the focus move started. |
| `focused_pane_id` | string | Pane with the focus after the call. It is never null. When `changed` is false, it holds the pane that has the focus now. That pane can be different from `source_pane_id` and not next to it. |
| `reason` | `no_neighbor` | Why the focus did not change. herdr omits it, and does not send `null`, when `changed` is `true`. |
| `layout` | [PaneLayoutSnapshot](../data-model.md#panelayoutsnapshot) | Layout of the tab after the call. |

**Errors**: `pane_not_found` (`"pane not found"`). A missing or unknown `direction` fails with `invalid_request`. Other codes are possible.

**CLI**: `herdr pane focus --direction <left|right|up|down> [--current | --pane <ID>]`

`--current` does not use the calling pane. It uses the UI-focused pane, as an omitted target does. This occurs also when `HERDR_PANE_ID` names the calling pane. To use the calling pane, give `--pane "$HERDR_PANE_ID"`. With no `HERDR_PANE_ID`, `--current` also uses the UI-focused pane and exits with status 0.

**Example**

```json
{"id":"1","method":"pane.focus_direction","params":{"pane_id":"w1:p1","direction":"right"}}
{"id":"1","result":{"type":"pane_focus_direction","focus":{"changed":true,"source_pane_id":"w1:p1","focused_pane_id":"w1:p2","layout":{"workspace_id":"w1","tab_id":"w1:t1","zoomed":false,"area":{"x":0,"y":0,"width":120,"height":40},"focused_pane_id":"w1:p2","panes":[{"pane_id":"w1:p1","focused":false,"rect":{"x":0,"y":0,"width":60,"height":40}},{"pane_id":"w1:p2","focused":true,"rect":{"x":60,"y":0,"width":60,"height":40}}],"splits":[{"id":"split_0_root","direction":"right","ratio":0.5,"rect":{"x":0,"y":0,"width":120,"height":40}}]}}}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked `no_neighbor`, a `focused_pane_id` that was not next to the source, the omitted `reason`, the errors, and the events. It ran `herdr pane focus --current` in a pane while the UI focus was in another workspace. The result named the UI-focused pane as `source_pane_id`. The example gave the same response again. The pass did not check `--current` on herdr 0.9.3.

---

## pane.get

Gets the `PaneInfo` of one pane by its ID. A retired pane ID from an earlier move can also find the pane. Refer to [Pane-move re-identification](../addressing.md#pane-move-re-identification).

**Params** (`PaneTarget`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Pane to get. |

**Result**: `type: "pane_info"`

| field | type | meaning |
|---|---|---|
| `pane` | [PaneInfo](../data-model.md#paneinfo) | The pane. |

**Errors**: `pane_not_found` (`"pane w9:p9 not found"`). A missing `pane_id` fails with `invalid_request`. Other codes are possible.

**CLI**: `herdr pane get <pane_id>`

**Example**

The pane runs a `claude` agent at its folder-trust dialog, as in the [pane.current](#panecurrent) example.

```json
{"id":"cli:pane:get","method":"pane.get","params":{"pane_id":"w2:p1"}}
{"id":"cli:pane:get","result":{"pane":{"agent":"claude","agent_status":"blocked","cwd":"/home/user/project","focused":true,"foreground_cwd":"/home/user/project","pane_id":"w2:p1","revision":1,"scroll":{"max_offset_from_bottom":0,"offset_from_bottom":0,"viewport_rows":40},"tab_id":"w2:t1","terminal_id":"term_65e0dcf8309d66","terminal_title":"user@host:/home/user/project","terminal_title_stripped":"user@host:/home/user/project","workspace_id":"w2"},"type":"pane_info"}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked the errors, the CLI usage error, and the old IDs of a pane after six moves. Each old ID returned the current pane. An ID that herdr never used returned `pane_not_found`. The example is a new capture.

---

## pane.input.set

Sets the right-click input routing of a pane. The herdr UI can handle a right-click with its context menu. Or herdr can send the right-click to the terminal program of the pane. No data on the wire tells the setting back. `PaneInfo` has no `right_click` field. A later `pane.get` shows no change, and herdr emits no event. Thus a caller cannot read back the value that it set.

**Params** (`PaneInputSetParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Target pane. |
| `right_click` | PaneRightClickTarget (`herdr`,`pane`) | yes | — | Send right-clicks to the herdr UI or to the program of the pane. |

**Result**: `type: "ok"`, with no other fields.

**Errors**: `pane_not_found`. Other codes are possible.

**CLI**: `herdr pane input --right-click <herdr|pane> [--current | --pane <ID> | PANE_ID]`

With no `HERDR_PANE_ID`, `--current` fails with `--current requires HERDR_PANE_ID` and exits with status 2.

**Example**

```json
{"id":"p2","method":"pane.input.set","params":{"pane_id":"w1:p1","right_click":"pane"}}
{"id":"p2","result":{"type":"ok"}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked the params, the errors, the `ok` result, the CLI forms, and that no read-back or event exists. The example gave the same response again.

The routing effect is not checked. It needs a person with a mouse at a real terminal.

---

## pane.layout

Returns the layout snapshot of the tab that contains a pane. The snapshot has the area of the tab, the split tree, and the rectangle of each pane. The `area` depends on the attached client. A server with no TUI client gives `x: 0, y: 0` and a size of 120 by 40.

**Params** (`PaneLayoutParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string \| null | no | null | Pane whose tab layout herdr returns. Null uses the UI-focused pane. |

**Result**: `type: "pane_layout"`

| field | type | meaning |
|---|---|---|
| `layout` | [PaneLayoutSnapshot](../data-model.md#panelayoutsnapshot) | Layout of the tab of the pane (`workspace_id`, `tab_id`, `zoomed`, `area`, `focused_pane_id`, `panes[]`, `splits[]`). |

**Errors**: `pane_not_found` (`"pane not found"`). Other codes are possible.

**CLI**: `herdr pane layout [--current | --pane <ID>]`

**Example**

```json
{"id":"cli:pane:layout","method":"pane.layout","params":{"pane_id":"w2:p1"}}
{"id":"cli:pane:layout","result":{"layout":{"area":{"height":40,"width":120,"x":0,"y":0},"focused_pane_id":"w2:p1","panes":[{"focused":true,"pane_id":"w2:p1","rect":{"height":20,"width":120,"x":0,"y":0}},{"focused":false,"pane_id":"w2:p2","rect":{"height":20,"width":120,"x":0,"y":20}}],"splits":[{"direction":"down","id":"split_0_root","ratio":0.5,"rect":{"height":40,"width":120,"x":0,"y":0}}],"tab_id":"w2:t1","workspace_id":"w2","zoomed":false},"type":"pane_layout"}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked the UI-focused target, the result fields, and the error. `--current` in a pane returned the tab of the calling pane. With a client attached in a 160 by 50 terminal, the area was 126 by 49. The example is a new capture.

---

## pane.link.activate

Finds a link that herdr detected at a viewport position of a pane, and opens it when a link handler is available. On herdr 0.9.5, an activation opened nothing.

- With a real URL at the position, `handled` was `false` and `url` had the URL.
- No opener process ran. A pass put shims for `xdg-open`, `gio`, `gnome-open`, `kde-open`, `x-www-browser`, `www-browser`, `sensible-browser`, `wslview`, `firefox`, and `chromium` first on `PATH`. `$BROWSER` named a shim. herdr ran none of them, also with a real client attached in a pseudo-terminal.
- A `url` value, not `handled`, tells if there is a link at the position.
- A value of `true` for `handled` possibly needs a link handler or a plugin. No pass linked a plugin, so this is not known.

To find link regions without these unknowns, use [pane.link.resolve](#panelinkresolve) first.

**Params** (`PaneLinkActivateParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Target pane. |
| `viewport_row` | integer (uint16) | yes | — | Zero-based viewport row. |
| `col` | integer (uint16) | yes | — | Zero-based viewport column. |
| `content_revision` | integer (uint64) \| null | no | null | Content revision for the position (inferred). It is the private copy-mode counter of [pane.copy_motion](#panecopy_motion). Only a `copy_motion` or `copy_search` result gives it. A different value fails with `stale_content`. The message is `"pane content changed before link activation"` here and `"…before link resolution"` from `pane.link.resolve`. Null or omitted skips the check. |
| `offset_from_bottom` | integer (uint64) \| null | no | null | Scroll offset for the viewport row (inferred). This value is a strict equality check against the current scroll offset of the pane. It does not move the read position. Any other value fails with `stale_content`. The message is `"pane viewport changed before link activation"` here and `"…before link resolution"` from `pane.link.resolve`. Null or omitted matches the current position. |

**Result**: `type: "pane_link_activated"`

| field | type | meaning |
|---|---|---|
| `handled` | boolean | True when a link handler activated the link. It can be `false` also when there **is** a link at the position. Refer to the text above. |
| `url` | string \| null | The URL at the position, if any. Its presence, not `handled`, tells that a link is there. herdr omits it, and does not send null, when there is no URL. |

**Errors**:

- `pane_not_found` (`"pane not found"`).
- `stale_content`, with the two messages above.
- `stale_target` (`"pane is no longer visible"`) when `pane_id` is not in the focused tab. This applies to a pane in another tab of the focused workspace, and to a pane in another workspace. This method and [pane.link.resolve](#panelinkresolve) are the only pane read methods in this namespace with this rule.
- Other codes are possible.

**CLI**: API only. There is no CLI subcommand.

**Example**

```json
{"id":"la1","method":"pane.link.activate","params":{"pane_id":"w1:p1","viewport_row":0,"col":0}}
{"id":"la1","result":{"type":"pane_link_activated","handled":false}}
```

Checked 2026-10-17 against herdr 0.9.5 on scratch servers. The pass checked the link case and the case with no link. It checked both `stale_content` messages and `stale_target`. It ran the opener-shim test again, with no client and with an attached client. The shim log file did not exist after either test. For the read-method rule, `pane.read`, `pane.get`, `pane.process_info`, `pane.layout`, and `pane.neighbor` worked on hidden panes. The pass did not check the copy-mode methods and `pane.selection.read` on hidden panes. The example gave the same response again.

---

## pane.link.resolve

Finds the link regions at a viewport position of a pane, and does not activate them. It takes the same params as [pane.link.activate](#panelinkactivate).

**Params** (`PaneLinkActivateParams`): the same fields as [pane.link.activate](#panelinkactivate).

**Result**: `type: "pane_link_resolved"`

| field | type | meaning |
|---|---|---|
| `regions` | `array<PaneLinkRegion>` | Link regions at the position. The array is empty when there are none. |

`PaneLinkRegion` gives inclusive columns of display cells on the current viewport of the pane:

| field | type | required | default | meaning |
|---|---|---|---|---|
| `row` | integer (uint16) | yes | — | Viewport row of the region. |
| `start_col` | integer (uint16) | yes | — | First column of the region (inclusive). |
| `end_col` | integer (uint16) | yes | — | Last column of the region (inclusive). |

**Errors**: `pane_not_found`. The `stale_content` and `stale_target` conditions of [pane.link.activate](#panelinkactivate) also apply. The two `stale_content` messages have `"resolution"` and not `"activation"`: `"pane viewport changed before link resolution"` and `"pane content changed before link resolution"`. A `viewport_row` or `col` out of range is not an error. It returns `regions: []`. Other codes are possible.

**CLI**: API only. There is no CLI subcommand.

**Example**

The pane shows `https://example.com/docs/abc` on row 5.

```json
{"id":"lr2","method":"pane.link.resolve","params":{"pane_id":"w1:p1","viewport_row":5,"col":5}}
{"id":"lr2","result":{"type":"pane_link_resolved","regions":[{"row":5,"start_col":0,"end_col":27}]}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. A URL of 29 characters at row 0 gave `start_col` 0 and `end_col` 28. The pass checked a position out of range, the three error messages, and the example.

---

## pane.list

Lists panes. With no `workspace_id`, it lists all panes of the server. With a `workspace_id`, it lists the panes of that workspace.

**Params** (`PaneListParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `workspace_id` | string \| null | no | null | Workspace to list. Null lists all panes. |

**Result**: `type: "pane_list"`

| field | type | meaning |
|---|---|---|
| `panes` | array of [PaneInfo](../data-model.md#paneinfo) | Panes that match. |

**Errors**: `workspace_not_found` for an unknown `workspace_id`, also the empty string (`"workspace  not found"`). Other codes are possible.

**CLI**: `herdr pane list [--workspace <WORKSPACE_ID>]`

**Example**

The pane runs a `claude` agent at its folder-trust dialog, as in the [pane.current](#panecurrent) example.

```json
{"id":"cli:pane:list","method":"pane.list","params":{"workspace_id":"w2"}}
{"id":"cli:pane:list","result":{"panes":[{"agent":"claude","agent_status":"blocked","cwd":"/home/user/project","focused":false,"foreground_cwd":"/home/user/project","pane_id":"w2:p1","revision":1,"scroll":{"max_offset_from_bottom":0,"offset_from_bottom":0,"viewport_rows":40},"tab_id":"w2:t1","terminal_id":"term_65e0dcf8309d66","terminal_title":"user@host:/home/user/project","terminal_title_stripped":"user@host:/home/user/project","workspace_id":"w2"}],"type":"pane_list"}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked empty params, a null `workspace_id`, one workspace, an unknown workspace, the empty string, and the CLI error. The example is a new capture.

---

## pane.move

Moves a pane to a different location. The destination can be an existing tab (`tab`), a new tab (`new_tab`), or a new workspace (`new_workspace`).

- After the move, use `result.move_result.pane.pane_id` or a live agent name. Do not use `result.move_result.previous_pane_id`.
- The pane ID changes only when the move goes into a different workspace. A move in the same workspace keeps the `pane_id`, also into a different tab.
- When a pane moves back into a workspace, it gets a new number. herdr does not use the old number again.

**Params** (`PaneMoveParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Pane to move. |
| `destination` | PaneMoveDestination | yes | — | Destination of the move (tagged union, below). |
| `focus` | boolean | no | false | True focuses the pane at its destination. |

`destination` is a `oneOf` with `type` as the discriminator.

*Variant `tab`*, which moves the pane into an existing tab:

| field | type | required | default | meaning |
|---|---|---|---|---|
| `type` | const `"tab"` | yes | — | Selects this variant. |
| `tab_id` | string | yes | — | Destination tab. |
| `split` | SplitDirection (`right`,`down`) | yes | — | Split direction against the target in that tab. |
| `target_pane_id` | string \| null | no | null | Pane in the destination tab to split against. Null uses the focused pane of the tab. |
| `ratio` | number (float) \| null | no | null | Split ratio for the new position. herdr clamps it to `[0.1, 0.9]` and does not check it. `5.0` becomes `0.9`, and `-1.0` becomes `0.1`. Both give `changed: true` and no warning. |

*Variant `new_tab`*, which moves the pane into a new tab:

| field | type | required | default | meaning |
|---|---|---|---|---|
| `type` | const `"new_tab"` | yes | — | Selects this variant. |
| `workspace_id` | string \| null | no | null | Workspace for the new tab. Null uses the current workspace of the pane. |
| `label` | string \| null | no | null | Label for the new tab. |

*Variant `new_workspace`*, which moves the pane into a new workspace:

| field | type | required | default | meaning |
|---|---|---|---|---|
| `type` | const `"new_workspace"` | yes | — | Selects this variant. |
| `label` | string \| null | no | null | Label for the new workspace. With no label, the workspace label is the basename of the cwd, for example `project`. |
| `tab_label` | string \| null | no | null | Label for the first tab of the new workspace. With no label, the tab label is `1`. |

**Result**: `type: "pane_move"`. The fields below are one level down, in `move_result`:

| field | type | meaning |
|---|---|---|
| `move_result` | PaneMoveResult (below) | Result of the move. |

`PaneMoveResult`. herdr omits each optional field when it does not apply:

| field | type | meaning |
|---|---|---|
| `changed` | boolean | True when the pane moved. |
| `pane` | [PaneInfo](../data-model.md#paneinfo) | The pane at its new location. Its `pane_id` changes only when the move goes into a different workspace. A move in the same workspace keeps the `pane_id`, which is then equal to `previous_pane_id`. |
| `previous_pane_id` | string | The ID of the pane before the move. Do not use it as a target. |
| `previous_workspace_id` | string | Workspace before the move. |
| `previous_tab_id` | string | Tab before the move. |
| `focused_pane_id` | string | Pane with the focus after the move. |
| `target_layout` | [PaneLayoutSnapshot](../data-model.md#panelayoutsnapshot) | Layout of the destination tab. |
| `source_layout` | [PaneLayoutSnapshot](../data-model.md#panelayoutsnapshot) \| null | Layout of the origin tab after the move. herdr omits it when the origin tab closed. |
| `created_tab` | [TabInfo](../data-model.md#tabinfo) \| null | Tab that a `new_tab` or `new_workspace` move created. |
| `created_workspace` | [WorkspaceInfo](../data-model.md#workspaceinfo) \| null | Workspace that a `new_workspace` move created. |
| `closed_tab_id` | string \| null | Tab that closed because the move made it empty. |
| `closed_workspace_id` | string \| null | Workspace that closed because the move made it empty. |
| `reason` | `same_tab` \| `zoomed_tab` \| null | Why the move did nothing or had a limit. A destination tab that is the current tab of the pane gives `reason: "same_tab"` **before** herdr checks `target_pane_id` and `ratio`. Thus herdr accepts an unknown `target_pane_id` or a `ratio` out of range in that case, with no error. |

**Errors**:

- `pane_not_found` (`"source pane not found"`).
- `target_pane_not_found` (`"target pane <ID> not found"`). When the ID names a real pane outside the destination tab, the message is `"target pane <ID> is not in tab <TAB_ID>"`.
- `workspace_not_found` when a `new_tab` destination names an unknown `workspace_id`.
- `tab_not_found` (`"tab w1:t9 not found"`) when a `tab` destination names an unknown `tab_id`.
- Other codes are possible.

**CLI**: `herdr pane move <PANE_ID> [--tab <TAB_ID> --split <right|down> [--target-pane <ID>] [--ratio <FLOAT>] | --new-tab [--workspace <ID>] | --new-workspace] [--label <TEXT>] [--tab-label <TEXT>] [--focus | --no-focus]`

The CLI focuses the moved pane by default. The API default is `focus: false`. To keep the API default on the CLI, give `--no-focus`.

Events:

- A successful move emits `pane_moved`. Its data has `pane`, `previous_pane_id`, `previous_tab_id`, `previous_workspace_id`, `created_tab`, `created_workspace`, `closed_tab_id`, and `closed_workspace_id`, when they apply.
- herdr emits `layout_updated` one time for each tab that the move changes. When the origin tab and the destination tab both stay, it emits two. When the origin tab closes, it emits one.
- A `new_tab` move also emits `tab_created`. A `new_workspace` move emits `workspace_created` and `tab_created`. A move that makes a tab or a workspace empty emits `tab_closed` or `workspace_closed`.
- With `focus: true`, herdr emits `pane_focused`, `tab_focused`, and `workspace_focused` for the moved pane.
- With `focus: false`, herdr never focuses the moved pane. A move of a pane that does not have the focus emits no focus events.
- A move of the UI-focused pane can close the focused tab or workspace. Then herdr emits the three focus events for the pane that gets the focus, also with `focus: false`.
- A move that does nothing (`reason: "same_tab"` or `"zoomed_tab"`) emits no event.

**Example**

```json
{"id":"cli:pane:move","method":"pane.move","params":{"pane_id":"w1:p2","destination":{"type":"new_tab","label":"moved-tab"},"focus":false}}
{"id":"cli:pane:move","result":{"type":"pane_move","move_result":{"changed":true,"previous_pane_id":"w1:p2","previous_workspace_id":"w1","previous_tab_id":"w1:t1","pane":{"pane_id":"w1:p2","terminal_id":"term_65e0dcd50f7862","workspace_id":"w1","tab_id":"w1:t3","focused":false,"cwd":"/home/user/project","foreground_cwd":"/home/user/project","terminal_title":"user@host:/home/user/project","terminal_title_stripped":"user@host:/home/user/project","agent_status":"unknown","scroll":{"offset_from_bottom":0,"max_offset_from_bottom":0,"viewport_rows":38},"revision":1},"source_layout":{"workspace_id":"w1","tab_id":"w1:t1","zoomed":false,"area":{"x":0,"y":0,"width":120,"height":40},"focused_pane_id":"w1:p1","panes":[{"pane_id":"w1:p1","focused":true,"rect":{"x":0,"y":0,"width":120,"height":40}}],"splits":[]},"target_layout":{"workspace_id":"w1","tab_id":"w1:t3","zoomed":false,"area":{"x":0,"y":0,"width":120,"height":40},"focused_pane_id":"w1:p2","panes":[{"pane_id":"w1:p2","focused":true,"rect":{"x":0,"y":0,"width":120,"height":40}}],"splits":[]},"created_tab":{"tab_id":"w1:t3","workspace_id":"w1","number":3,"label":"moved-tab","focused":false,"pane_count":1,"agent_status":"unknown"},"focused_pane_id":"w1:p2"}}}
```

In this example, `pane.pane_id` (`"w1:p2"`) is equal to `previous_pane_id`. The move went to a different tab in workspace `w1`, so the ID did not change.

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked these items:

- all three destination variants, the default labels, and the CLI forms
- the ID after moves in one workspace and across workspaces
- the ratio clamp, the `same_tab` rule, `zoomed_tab`, and the omitted fields
- each error message, with `tab_not_found`
- the CLI default focus, and the events with `focus: true` and `focus: false`

The example is a new capture. The pass did not check the focus events of a move of the focused pane on herdr 0.9.3. Thus this behavior is possibly not new in 0.9.5.

---

## pane.neighbor

Finds the ID of the next pane in a direction from a given pane, and does not move the focus. When there is no pane in that direction, herdr omits `neighbor_pane_id`.

**Params** (`PaneNeighborParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `direction` | PaneDirection (`left`,`right`,`up`,`down`) | yes | — | Direction to look in. |
| `pane_id` | string \| null | no | null | Start pane. Null uses the UI-focused pane. |

**Result**: `type: "pane_neighbor"`. The fields below are one level down, in `neighbor`:

| field | type | meaning |
|---|---|---|
| `neighbor` | PaneNeighborResult (below) | Result of the lookup. |

`PaneNeighborResult`:

| field | type | meaning |
|---|---|---|
| `pane_id` | string | Start pane. |
| `direction` | PaneDirection | Direction of the lookup. |
| `neighbor_pane_id` | string \| null | The next pane. herdr omits it when there is none. |
| `layout` | [PaneLayoutSnapshot](../data-model.md#panelayoutsnapshot) | Layout of the tab that contains the pane. |

**Errors**: `pane_not_found` (`"pane not found"`). A missing `direction` fails with `invalid_request`. Other codes are possible.

**CLI**: `herdr pane neighbor --direction <left|right|up|down> [--current | --pane <ID>]`

`--current` does not use the calling pane. It uses the UI-focused pane, as in [pane.focus_direction](#panefocus_direction). To use the calling pane, give `--pane "$HERDR_PANE_ID"`.

**Example**

The tab has a second pane below `w2:p1`, so there is no pane to the right.

```json
{"id":"cli:pane:neighbor","method":"pane.neighbor","params":{"pane_id":"w2:p1","direction":"right"}}
{"id":"cli:pane:neighbor","result":{"neighbor":{"direction":"right","layout":{"area":{"height":40,"width":120,"x":0,"y":0},"focused_pane_id":"w2:p1","panes":[{"focused":true,"pane_id":"w2:p1","rect":{"height":20,"width":120,"x":0,"y":0}},{"focused":false,"pane_id":"w2:p2","rect":{"height":20,"width":120,"x":0,"y":20}}],"splits":[{"direction":"down","id":"split_0_root","ratio":0.5,"rect":{"height":40,"width":120,"x":0,"y":0}}],"tab_id":"w2:t1","workspace_id":"w2","zoomed":false},"pane_id":"w2:p1"},"type":"pane_neighbor"}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked a lookup with a next pane, a lookup with none, the UI-focused target, and the errors. It ran `herdr pane neighbor --current` in a pane while the UI focus was in another workspace. The result named the UI-focused pane. The example is a new capture.

---

## pane.process_info

Returns operating-system process data for a pane. The data has the PID and the TTY of the shell, and the foreground process group and its processes.

**Params** (`PaneProcessInfoParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string \| null | no | null | Pane to examine. Null uses the UI-focused pane. |

**Result**: `type: "pane_process_info"`. The fields below are one level down, in `process_info`:

| field | type | meaning |
|---|---|---|
| `process_info` | PaneProcessInfoResult (below) | Process data. |

`PaneProcessInfoResult`:

| field | type | meaning |
|---|---|---|
| `pane_id` | string | Pane that herdr examined. |
| `shell_pid` | integer (uint32) \| null | PID of the shell of the pane. |
| `tty` | string \| null | Path of the controlling TTY. No pass saw it on Linux. It was absent in 0.9.1 for an idle shell, a busy pipeline, a visible pane, and a hidden pane, through the API and the CLI. It was absent in 0.9.3 for idle shells. It was absent in 0.9.5 for an idle shell, a busy pipeline, a hidden pane, and an agent pane. |
| `foreground_process_group_id` | integer (uint32) \| null | ID of the foreground process group. |
| `foreground_processes` | `array<PaneProcessInfoProcess>` | Processes in the foreground group (below). |

`PaneProcessInfoProcess`:

| field | type | required | meaning |
|---|---|---|---|
| `pid` | integer (uint32) | yes | Process ID. |
| `name` | string | yes | Process name. |
| `argv0` | string \| null | no | `argv[0]` as the process ran. No pass saw a value on Linux. Each process entry had `argv`, `cmdline`, `cwd`, `name`, and `pid`, but no `argv0`. |
| `argv` | `array<string>` \| null | no | Full argument vector. |
| `cmdline` | string \| null | no | Raw command line. |
| `cwd` | string \| null | no | Working directory of the process. |

**Errors**: `pane_not_found` (`"pane not found"`). Other codes are possible.

**CLI**: `herdr pane process-info [--current | --pane <ID>]`

**Example**

```json
{"id":"cli:pane:process_info","method":"pane.process_info","params":{"pane_id":"w2:p1"}}
{"id":"cli:pane:process_info","result":{"process_info":{"foreground_process_group_id":217209,"foreground_processes":[{"argv":["claude"],"cmdline":"claude","cwd":"/home/user/project","name":"claude","pid":217209}],"pane_id":"w2:p1","shell_pid":215879},"type":"pane_process_info"}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked an idle shell in a hidden tab, a `sleep 300 | cat` pipeline with two processes, and an agent pane. It checked the UI-focused target, the errors, and `--current` in a pane, which used the calling pane. The example is a new capture.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the meaning of `tty` and `argv0`, because no pass saw a value for them.

---

## pane.read

Reads a snapshot of the terminal output of a pane. Select the `source` for the task. The read sources are at the top of this page.

- `lines` sets the maximum number of rows. herdr has an undocumented hard limit of 1000 rows, whatever the requested value is. A request of 999 gives 999 rows. Requests of 1000, 1001, 2000, and 5000 each give 1000 rows, with `truncated: true`.
- herdr takes the rows from the most recent end. `lines: 0` returns empty text with `truncated: true`.
- With no `lines`, herdr does not return all rows. For `recent` and `recent_unwrapped`, the server default is exactly 80 rows, with `truncated: true`. `visible` and `detection` return the full viewport, with `truncated: false`.
- The schema tells that `strip_ansi` (default true) removes escape sequences unless `format` is `ansi`. But it had no effect on 0.9.1, 0.9.3, or 0.9.5. `format: "text"` with `strip_ansi: false` returned fully stripped text. `format: "ansi"` with `strip_ansi: true` returned SGR sequences. Only `format` controls the stripping.
- A read does not mark an agent as seen. Refer to the top of this page and to [AgentStatus](../data-model.md#agentstatus).

**Params** (`PaneReadParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Pane to read. |
| `source` | ReadSource (`visible`,`recent`,`recent_unwrapped`,`detection`) | yes | — | Snapshot to read. |
| `format` | ReadFormat (`text`,`ansi`) | no | `text` | `text` gives plain text. `ansi` keeps escape sequences. Only this field controls the stripping. Refer to `strip_ansi` below. |
| `lines` | integer (uint32) \| null | no | null | Maximum rows from the screen and the scrollback, with a limit of 1000. Null uses the server default above. That default is 80 rows for `recent` and `recent_unwrapped`, and the full viewport for `visible` and `detection`. It does not return all rows. |
| `strip_ansi` | boolean | no | true | The schema tells that it strips ANSI escapes from the returned text. It had no effect on 0.9.1, 0.9.3, or 0.9.5. Refer to the text above. |

**Result**: `type: "pane_read"`. The fields below are one level down, in `read`:

| field | type | meaning |
|---|---|---|
| `read` | PaneReadResult (below) | The snapshot. |

`PaneReadResult` also occurs in `output_matched.read` of [pane.wait_for_output](#panewait_for_output). Refer to [PaneReadResult](../data-model.md#panereadresult).

| field | type | meaning |
|---|---|---|
| `pane_id` | string | Pane that herdr read. |
| `workspace_id` | string | Workspace of the pane. |
| `tab_id` | string | Tab of the pane. |
| `source` | ReadSource | Snapshot source that herdr used. |
| `format` | ReadFormat | Format of `text`. |
| `text` | string | The output. |
| `revision` | integer (uint64) | The schema tells that this is the content revision of the pane at the time of the read. But each read gave `0`, also after 3000 lines of output. It follows neither `PaneInfo.revision` nor the copy-mode `content_revision` (refer to [pane.copy_motion](#panecopy_motion)). As a `content_revision` in this namespace, this value always fails with `stale_content`. |
| `truncated` | boolean | True when `lines` or the server default cut the output. |

**Errors**: `pane_not_found`. A missing or unknown `source` fails with `invalid_request`. Other codes are possible.

**CLI**: `herdr pane read <PANE_ID> [--source <visible|recent|recent-unwrapped|detection>] [--lines <N>] [--format <text|ansi> | --ansi] [--raw]`

The default CLI source is `recent`. [cli-mapping.md](../cli-mapping.md#cli-conventions) gives the CLI conventions.

**Example**

```json
{"id":"1","method":"pane.read","params":{"pane_id":"w1:p3","source":"recent_unwrapped","lines":120,"format":"text","strip_ansi":true}}
{"id":"1","result":{"type":"pane_read","read":{"pane_id":"w1:p3","workspace_id":"w1","tab_id":"w1:t2","source":"recent_unwrapped","format":"text","text":"docprobe-marker-42\n[user@host project]$","revision":0,"truncated":false}}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server, after 3000 lines of output. The pass checked the row limit, `lines: 0`, the default rows for each source, the `strip_ansi` combinations, the `revision` value, the errors, and the CLI usage. It also checked that an API read and a CLI read leave `done` unchanged. The example is a new capture.

---

## pane.release_agent

Releases the lifecycle authority that a `source` has for one named agent on a pane. This method is narrower than [pane.clear_agent_authority](#paneclear_agent_authority), which removes the full claim of the source on the pane. The method is idempotent. A release that names a different agent returns `ok` and changes nothing.

**Params** (`PaneReleaseAgentParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Target pane. |
| `source` | string | yes | — | Reporting source that releases the authority. |
| `agent` | string | yes | — | Agent label whose authority herdr releases. |
| `seq` | integer (uint64) \| null | no | null | Monotonic sequence that orders reports from this source. herdr drops a release whose `seq` is not greater than the last `seq` of that source. This applies to an equal, a lower, a null, or an omitted `seq` after a report with `seq`. The dropped release returns `{"type":"ok"}`, and the agent of the pane does not change. Thus the result does not tell a release from a dropped release. |

**Result**: `type: "ok"`, with no other fields.

**Errors**: `pane_not_found`. A missing `agent` fails with `invalid_request`. Other codes are possible.

**CLI**: `herdr pane release-agent <PANE_ID> --source <ID> --agent <LABEL> [--seq <N>]`

On success, the CLI prints nothing and exits with status 0. An error uses the request ID `cli:request`. Refer to [cli-mapping.md](../cli-mapping.md#cli-conventions).

**Example**

```json
{"id":"p7","method":"pane.release_agent","params":{"pane_id":"w1:p4","source":"demo-reporter","agent":"claude"}}
{"id":"p7","result":{"type":"ok"}}
```

When a release has an effect, herdr emits `pane_agent_detected` with `released: true` and `final_status: "unknown"`. It also emits `pane.agent_status_changed` with `unknown`. A dropped release, or a release from a source that another source replaced, emits no event.

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked a second release, a release for a different agent, and the `seq` rule with an equal, a lower, no, and a greater `seq`. It also checked a release from a replaced source, the events, the errors, and the CLI output. The example is a new capture.

---

## pane.rename

Sets or clears the label of a pane, which people see in the herdr UI.

- A null or omitted `label` clears the label. The CLI form is `--clear`.
- An empty string and a label of only spaces also clear the label.
- There is no documented or observed maximum length. Labels of 500 and 5000 characters came back unchanged.
- This method and [pane.input.set](#paneinputset) emit no event, such as `pane_updated` or `layout_updated`. Thus a client cannot learn about a label change from the event stream.

**Params** (`PaneRenameParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Pane to rename. |
| `label` | string \| null | no | null | New label. Null, an empty string, or only spaces clear the label. A pass saw no length limit. |

**Result**: `type: "pane_info"`

| field | type | meaning |
|---|---|---|
| `pane` | [PaneInfo](../data-model.md#paneinfo) | The changed pane. |

**Errors**: `pane_not_found`. Other codes are possible.

**CLI**: `herdr pane rename <PANE_ID> [LABEL]... [--clear]`

The CLI joins the label words with spaces. With no label and no `--clear`, it prints a usage line and exits with status 2.

**Example**

```json
{"id":"cli:pane:rename","method":"pane.rename","params":{"pane_id":"w1:p3","label":"build"}}
{"id":"cli:pane:rename","result":{"pane":{"agent_status":"unknown","cwd":"/home/user/project","focused":false,"foreground_cwd":"/home/user/project","label":"build","pane_id":"w1:p3","revision":1,"scroll":{"max_offset_from_bottom":0,"offset_from_bottom":0,"viewport_rows":40},"tab_id":"w1:t2","terminal_id":"term_65e0dce20eda73","terminal_title":"user@host:/home/user/project","terminal_title_stripped":"user@host:/home/user/project","workspace_id":"w1"},"type":"pane_info"}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked the long labels, each clear form, and that no event occurs across 8 renames. It also checked the CLI forms and the errors. The example is a new capture.

---

## pane.report_agent

Reports the lifecycle state of a coding agent in a pane. An integration sends the report as `source`. With this method, an agent integration tells herdr its `idle`, `working`, `blocked`, or `unknown` state, and an optional session identity.

Authority rules:

- Authority is **not exclusive**, and it does **not** stay until a release.
- A report from a different source takes the pane at once. The agent label and the status both change, with no error. Thus the newest reporting source has the authority.
- [pane.release_agent](#panerelease_agent) works only for the source that has the authority now. A release from a replaced source returns `ok` and changes nothing.

Status rules:

- An `idle` report after `working` or `blocked` shows as `done` only when the tab of the pane is not the focused tab. `AgentInfo.completion_seq` then records the completion.
- A focus of the tab changes `done` to `idle` and keeps `completion_seq`.
- When the tab of the pane already has the focus, the pane shows `idle` directly.
- herdr emits `pane_agent_detected` only when the reported agent label changes. It emits `pane.agent_status_changed` only when the status changes. A repeated report with the same agent and the same state emits nothing.
- [AgentStatus](../data-model.md#agentstatus) tells what `idle` and `done` mean for readiness.

Integration sources:

- herdr ignores some reports with a `source` of the form `herdr:<target>`. Here `<target>` is one of the [IntegrationTarget](../data-model.md#integrationtarget) values: `pi`, `omp`, `claude`, `codex`, `copilot`, `devin`, `droid`, `kimi`, `opencode`, `kilo`, `hermes`, `qodercli`, `qwen`, `cursor`, `mastracode`, `antigravity_cli`, and `grok`.
- herdr ignores such a report when the reported agent label is the agent of that integration. herdr normalizes the label first. For example, `Claude` and `claude-code` become `claude`, `qwen-code` becomes `qwen`, and `antigravity` becomes `agy`. The agent of `antigravity_cli` is `agy`.
- An ignored report returns `ok`, but the agent of the pane stays unset. The reason is possibly that an external reporter must not act as the integration of herdr itself (inferred).
- The prefix alone does not cause this. herdr applied these reports: `herdr:claude` with `pi`, `herdr:x` with `x`, a bare `herdr:` with `codex`, and `HERDR:claude` with `claude`. A `herdr:` source that is not an integration target is also applied, for example `herdr:gemini` with `gemini` and `herdr:agy` with `antigravity_cli`.
- A release from a `herdr:` source works normally.

**Params** (`PaneReportAgentParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Pane with the agent. |
| `source` | string | yes | — | ID of the reporting integration or source. herdr accepts an empty string. After a source sends a report with `seq`, herdr drops each later report from that source with no `seq` (as `seq: 0`) or with a lower `seq`. Thus keep `seq` in each report after the first one. The rule applies to each source separately. |
| `agent` | string | yes | — | Agent label. An empty string fails with the undocumented `invalid_agent` (`"agent label must not be empty"`). |
| `state` | PaneAgentState (`idle`,`working`,`blocked`,`unknown`) | yes | — | Reported lifecycle state. |
| `message` | string \| null | no | null | Status message for a person. herdr accepts it. No field of `pane.get` shows it. |
| `seq` | integer (uint64) \| null | no | null | Monotonic sequence that orders reports from this source. |
| `agent_session_id` | string \| null | no | null | Agent session identifier (ID form). herdr accepts it, but no read method shows it later. `PaneInfo.agent_session` stays absent. `agent.get`, `agent.list`, and `session.snapshot` have no session fields. |
| `agent_session_path` | string \| null | no | null | Agent session identifier (path form). No read method shows it, as for `agent_session_id`. |
| `resume_argv` | `array<string>` \| null | no | null | New in 0.9.3. The schema text is "Command that resumes this agent's session after a Herdr restart. The first element must be a plain command name." Refer to [resume_argv rules](#resume_argv-rules). The CLI form is the arguments after `--`. |

### resume_argv rules

These rules apply to `pane.report_agent` and to [pane.report_agent_session](#panereport_agent_session).

- herdr checks `resume_argv` after the `pane_not_found` check and the `invalid_agent` check.
- A bad value fails with `invalid_resume_argv`.
- The array must not be empty (`"resume_argv must not be empty"`).
- The first element must be a plain command name (`"resume_argv must start with a plain command name, not a path"`). Such a name has only ASCII letters, digits, `.`, `_`, and `-`, and does not start with `-`.
- herdr rejects `/usr/bin/claude`, `./claude`, `~`, `""`, `"claude code"`, `FOO=1`, `-x`, and `clé`. It accepts `claude-code`, `claude.exe`, `.claude`, and `1claude`.
- The array can have a maximum of 64 elements (`"resume_argv allows at most 64 arguments"`).
- The elements can have a maximum of 8192 bytes in total (`"resume_argv allows at most 8192 bytes"`).
- No element can contain a control character, such as a newline, NUL, a tab, or DEL (`"resume_argv must not contain control characters"`).
- herdr accepts a valid value only when `source` holds the pane for this `agent` after the report. If not, the call fails with `resume_not_accepted` (`"resume_argv requires the reporter to hold the pane; report its state with pane.report_agent first"`).
- herdr accepts null.
- `pane.get`, `agent.list`, and `session.snapshot` do not show the value.

The `resume_agents_on_restore` setting in [raw/default-config.toml](../raw/default-config.toml) controls the resume after a restart. [AgentSessionInfo](../data-model.md#agentsessioninfo) owns the claims about the resume after a restart, and their stamp. This page did not check them.

**Result**: `type: "ok"`, with no other fields.

**Errors**:

- `pane_not_found` (`"pane w1:p99 not found"`).
- `invalid_agent` for an empty `agent`.
- `invalid_resume_argv` and `resume_not_accepted`. Refer to [resume_argv rules](#resume_argv-rules). A `herdr:<target>` source that herdr ignores fails with `resume_not_accepted` when it sends a valid `resume_argv`.
- `invalid_request` for a `state` of `done`.
- Other codes are possible.

**CLI**: `herdr pane report-agent <PANE_ID> --source <ID> --agent <LABEL> --state <idle|working|blocked|unknown> [--message <TEXT>] [--seq <N>] [--agent-session-id <ID>] [--agent-session-path <PATH>] [-- <RESUME_ARG>...]`

On success, the CLI prints nothing and exits with status 0. Refer to [cli-mapping.md](../cli-mapping.md#cli-conventions).

**Example**

```json
{"id":"p4","method":"pane.report_agent","params":{"pane_id":"w1:p1","source":"example-integration","agent":"claude","state":"working"}}
{"id":"p4","result":{"type":"ok"}}
```

Checked 2026-10-17 against herdr 0.9.5 on scratch servers. The pass checked these items:

- the authority change, a release from a replaced source, and the `seq` rule for each source
- `done` only for an unfocused tab, with `completion_seq`, and the events only on change
- the `herdr:` source rule for each integration target, and the label normalization
- `invalid_agent`, each `resume_argv` rule, `resume_not_accepted`, and the wire form of the CLI with `--`

The example is a new capture. The 0.9.3 pass recorded an unexplained case. On a pane with a long history of reports with `seq`, reports from new sources did not take the pane. The 0.9.5 pass did not see this case again. New sources took the pane at once, after more than 20 reports with `seq`. The pass checked the ignored `herdr:` sources only on 0.9.5. Thus it is not known if 0.9.3 ignored the agents other than `claude`, `pi`, and `codex`.

---

## pane.report_agent_session

Reports or updates the agent **session identity** for a pane, and does not change the lifecycle state. Use it to attach a session ID or path, and to record how the session started. Since 0.9.3, it can also set a `resume_argv`.

- The source does not need to hold the pane, unless the call has `resume_argv`.
- No read method shows the reported identity later. `PaneInfo.agent_session` stays absent. `agent.get`, `agent.list`, and `session.snapshot` have no session fields.
- This method emits no event. In contrast, [pane.report_agent](#panereport_agent) emits `pane_agent_detected` and `pane.agent_status_changed`.

**Params** (`PaneReportAgentSessionParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Pane with the agent. |
| `source` | string | yes | — | ID of the reporting integration or source. |
| `agent` | string | yes | — | Agent label. |
| `agent_session_id` | string \| null | no | null | Session identifier (ID form). |
| `agent_session_path` | string \| null | no | null | Session identifier (path form). |
| `session_start_source` | string \| null | no | null | How the session started (inferred). |
| `seq` | integer (uint64) \| null | no | null | Monotonic sequence that orders reports from this source. |
| `resume_argv` | `array<string>` \| null | no | null | New in 0.9.3. The schema text is "Command that resumes this agent's session after a Herdr restart. The first element must be a plain command name." Refer to [resume_argv rules](#resume_argv-rules). The CLI form is the arguments after `--`. |

**Result**: `type: "ok"`, with no other fields.

**Errors**:

- `pane_not_found` and `invalid_agent` for an empty `agent`.
- `invalid_resume_argv`. Refer to [resume_argv rules](#resume_argv-rules).
- `resume_not_accepted` when a source that does not hold the pane for this `agent` sends `resume_argv`. This applies to a source that never reported, a released source, a source that another source replaced, and a source that names a different agent.
- Other codes are possible.

One exception applies to `resume_not_accepted`. After a source uses `seq`, herdr drops a report from it with no `seq` or with a `seq` that is not greater than its last one. The dropped report returns `ok` and herdr does not do the hold check. Thus `resume_not_accepted` does not occur for it. The format checks of `resume_argv` still run.

**CLI**: `herdr pane report-agent-session <PANE_ID> --source <ID> --agent <LABEL> [--seq <N>] [--agent-session-id <ID>] [--agent-session-path <PATH>] [--session-start-source <SOURCE>] [-- <RESUME_ARG>...]`

**Example**

```json
{"id":"p5","method":"pane.report_agent_session","params":{"pane_id":"w1:p1","source":"example-integration","agent":"claude","agent_session_id":"abc123"}}
{"id":"p5","result":{"type":"ok"}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked the envelope, the lifecycle state that did not change, and that no event occurs. It checked `invalid_agent`, `pane_not_found`, and each `resume_argv` rule. It checked `resume_not_accepted` for the four cases above and the `seq` exception. It also checked the wire form of the CLI. The example is a new capture.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the meaning of `session_start_source`. herdr accepts it, but no read method shows it.

Checked 2026-10-17 against herdr 0.9.5: the `herdr pane report-agent-session --help` output shows the `--session-start-source` flag. No pass sent this flag to a server.

---

## pane.report_metadata

Reports display metadata for a pane. The metadata is a title, labels for each status, free tokens, and a display name for the agent. It can have a TTL. This metadata changes only what herdr shows. It does not change the agent lifecycle.

- The boolean `clear_*` flags remove the related metadata.
- Each `tokens` key must match `^[A-Za-z0-9_-]{1,32}$`. One report can change a maximum of 16 keys. Refer to the token totals below.
- `state_labels` maps a status to a label.
- A report that sets `title`, `display_agent`, or `state_labels` **replaces** these three fields. It does not merge with earlier reports field by field. A later report with only `title` removes an earlier `display_agent` and `state_labels`, unless it sends them again.
- A report that sets only `tokens` keeps the earlier `title`, `display_agent`, and `state_labels`. In one check, a report with only `clear_display_agent` also kept the earlier `title`.
- Tokens are different. They accumulate across reports, up to 32 tokens for each pane.
- A report that changes `tokens` emits `pane_updated` with the full `PaneInfo`, and increments `PaneInfo.revision`. The subscription type of this event is `pane.updated`. Refer to the [event catalog](../events.md#event-catalog-eventkind--eventdata).
- A report that changes only `title`, `display_agent`, or `state_labels` emits no `pane_updated`. It does not change `revision`.
- A subscription to `pane.agent_status_changed` for that pane gets an event with the new `title`, `display_agent`, or `state_labels`. The event also has the unchanged `agent_status`. This also occurs when a TTL ends.
- With no such subscription, a client sees these changes only when it reads the pane.
- `pane.report_agent` also emits no `pane_updated`.

**Params** (`PaneReportMetadataParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Target pane. |
| `source` | string | yes | — | ID of the reporting source. |
| `agent` | string \| null | no | null | **A display gate, not only a label.** herdr stores metadata with `agent`. `PaneInfo` does not show it until the reported agent of the pane is equal to that value. The report replaces the earlier record at once. Thus the earlier visible metadata goes away, also while herdr holds back the new metadata. |
| `applies_to_source` | string \| null | no | null | Shows the metadata only while this source has the pane. It works as the `agent` gate does. A different reporting source hides the metadata. |
| `title` | string \| null | no | null | Display title to set. |
| `clear_title` | boolean | no | false | Clear the display title. |
| `display_agent` | string \| null | no | null | Display name of the agent. |
| `clear_display_agent` | boolean | no | false | Clear the display name of the agent. |
| `state_labels` | `object<string,string>` | no | — | Map from status to label text. |
| `clear_state_labels` | boolean | no | false | Clear all state labels. |
| `tokens` | `object<string, string \| null>` | no | — | Display tokens. A report can have a maximum of 16 keys, with the key pattern `^[A-Za-z0-9_-]{1,32}$`. A null value clears one key. Tokens **accumulate across reports** up to 32 for each pane, as the `maxProperties: 32` of `PaneInfo.tokens` tells. A report that goes above 32 fails fully with `metadata_token_limit`. |
| `seq` | integer (uint64) \| null | no | null | Monotonic sequence that orders reports from this source. |
| `ttl_ms` | integer (uint64) \| null | no | null | Time in ms (1..=86400000). After this time, herdr removes the metadata of the report, with its tokens. |

**Result**: `type: "ok"`, with no other fields.

**Errors**:

- `pane_not_found`.
- `invalid_metadata_request`. The message is `"cannot set and clear the same metadata field"` when a report sets and clears one field, such as `title` and `clear_title`. The message is `"missing metadata field to set or clear"` for a report with only `pane_id` and `source`.
- `invalid_metadata_ttl` (`"metadata ttl_ms must be at least 1"`, `"metadata ttl_ms must be 86400000 or less"`).
- `invalid_metadata_token` (`"invalid metadata token key: bad key!"`, `"a metadata report may update at most 16 tokens"`).
- `invalid_state_label` (`"unknown state label: bogus"`). `done` is a valid key.
- `metadata_token_limit` (`"pane metadata may contain at most 32 tokens"`). Refer to `tokens` above.
- `invalid_metadata_source` for an empty `source`. [errors.md](../errors.md) owns this claim and its stamp.
- Other codes are possible.

**CLI**: `herdr pane report-metadata <PANE_ID> --source <ID> [--agent <LABEL>] [--applies-to-source <ID>] [--title <TEXT> | --clear-title] [--display-agent <TEXT> | --clear-display-agent] [--state-label <STATUS=TEXT> | --clear-state-labels] [--token <NAME=VALUE> | --clear-token <NAME>] [--seq <N>] [--ttl-ms <N>]`

The usage shows `--token` and `--clear-token` as alternatives, but one call can use both. `--clear-token <NAME>` sends a null value for that key.

**Example**

```json
{"id":"1","method":"pane.report_metadata","params":{"pane_id":"w1:p1","source":"example-integration","title":"build","tokens":{"branch":"main"},"ttl_ms":60000}}
{"id":"1","result":{"type":"ok"}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked these items:

- the replace and accumulate rules, a report with only `tokens`, the null token value, and the 32-token limit
- the `agent` gate and the `applies_to_source` gate
- the TTL end after 300 ms, and the tokens that the TTL removed
- each error message in the list except `invalid_metadata_source`
- the events and the revision for each field, and the wire form of the CLI

The example is a new capture. The 0.9.3 pass possibly subscribed only to `pane.updated`. Thus it is not known if 0.9.3 also sent the metadata in `pane.agent_status_changed`.

---

## pane.resize

Resizes the split that contains a pane. The call moves the divider of that split in a direction.

- `direction` moves the divider in that direction, whichever side the pane is on. The pane on the side that the divider moves away from grows.
- For the left pane of a split, `right` makes the pane wider. For the right pane, `right` makes the pane narrower.
- `amount` is a ratio change, and it is optional. With no `amount`, herdr uses its default step, which a pass measured as 0.05.
- herdr ignores the sign of `amount`. Only its magnitude has an effect, and `direction` gives the sign. Thus `direction: "right", amount: -0.2` moves the divider to the right.
- A magnitude above 0.5 acts as 0.5.
- herdr clamps split ratios to `[0.1, 0.9]`.
- When the resize has no effect, `changed` is false and `reason` is `unchanged`. This occurs for a move past a clamp limit, an `amount` of 0, and a direction with no split on that axis.
- A resize with `changed: true` emits `layout_updated`.

**Params** (`PaneResizeParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `direction` | PaneDirection (`left`,`right`,`up`,`down`) | yes | — | Direction for the divider move. It also sets the sign of `amount`. |
| `pane_id` | string \| null | no | null | Pane to resize. Null uses the UI-focused pane. |
| `amount` | number (float) \| null | no | null | Magnitude of the resize. herdr uses only the absolute value, with a maximum of 0.5. Null uses the default step. A value that is not a number fails with `invalid_request`. |

**Result**: `type: "pane_resize"`. The fields below are one level down, in `resize`:

| field | type | meaning |
|---|---|---|
| `resize` | PaneResizeResult (below) | Result of the resize. |

`PaneResizeResult`:

| field | type | meaning |
|---|---|---|
| `changed` | boolean | True when the geometry changed. |
| `pane_id` | string | Pane that herdr resized. |
| `focused_pane_id` | string | Focused pane in the tab of the resized pane after the call. It is possibly not the globally focused pane, when that tab does not have the focus. |
| `layout` | [PaneLayoutSnapshot](../data-model.md#panelayoutsnapshot) | Layout of the tab after the call. |
| `reason` | `unchanged` | Why nothing changed. herdr omits it, and does not send `null`, when `changed` is `true`. |

**Errors**: `pane_not_found` (`"pane not found"`). Other codes are possible.

**CLI**: `herdr pane resize --direction <left|right|up|down> [--amount <FLOAT>] [--current | --pane <ID>]`

With no `HERDR_PANE_ID`, `--current` sends no pane ID, so herdr uses the UI-focused pane. Refer to [cli-mapping.md](../cli-mapping.md#cli-conventions).

**Example**

The example resizes the right pane `w1:p2` of a split with a ratio of 0.5. The divider moves to the right, and `w1:p2` becomes narrower.

```json
{"id":"cli:pane:resize","method":"pane.resize","params":{"pane_id":"w1:p2","direction":"right","amount":0.1}}
{"id":"cli:pane:resize","result":{"type":"pane_resize","resize":{"changed":true,"pane_id":"w1:p2","focused_pane_id":"w1:p1","layout":{"workspace_id":"w1","tab_id":"w1:t1","zoomed":false,"area":{"x":0,"y":0,"width":120,"height":40},"focused_pane_id":"w1:p1","panes":[{"pane_id":"w1:p1","focused":true,"rect":{"x":0,"y":0,"width":72,"height":40}},{"pane_id":"w1:p2","focused":false,"rect":{"x":72,"y":0,"width":48,"height":40}}],"splits":[{"id":"split_0_root","direction":"right","ratio":0.6,"rect":{"x":0,"y":0,"width":120,"height":40}}]}}}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked these items:

- the default step, the ignored sign, the clamp limits, and the 0.5 maximum
- the divider direction for the left pane and the right pane
- the `unchanged` cases, the omitted `reason`, and `focused_pane_id` in a tab without the focus
- the UI-focused target, the error, the event, and the wire form of the CLI

The example is a new capture.

---

## pane.scroll

Sets the scrollback offset of a pane from the bottom. An offset of 0 goes back to the live bottom of the output.

- An offset out of range clamps to `max_offset_from_bottom`. It is not an error.
- When the offset changes, herdr emits `pane.scroll_changed` with the new `PaneScrollInfo`. When the offset does not change, herdr emits no event.
- While a pane has a scroll offset, [pane.read](#paneread) with `source: "visible"` returns the scrolled viewport, not the live tail.

**Params** (`PaneScrollParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Target pane. |
| `offset_from_bottom` | integer (uint64) | yes | — | Number of rows to scroll back from the bottom. 0 goes back to the live bottom. |

**Result**: `type: "pane_info"`

| field | type | meaning |
|---|---|---|
| `pane` | [PaneInfo](../data-model.md#paneinfo) | The pane after the scroll. Its `scroll` field shows the new offset. |

**Errors**: `pane_not_found`. A missing or negative `offset_from_bottom` fails with `invalid_request`. Other codes are possible.

**CLI**: API only. There is no CLI subcommand.

**Example**

The pane has no scrollback above the viewport. Thus herdr clamps the requested offset 1 to 0.

```json
{"id":"s1","method":"pane.scroll","params":{"pane_id":"w1:p1","offset_from_bottom":1}}
{"id":"s1","result":{"type":"pane_info","pane":{"pane_id":"w1:p1","terminal_id":"term_65e0de30967ca1","workspace_id":"w1","tab_id":"w1:t1","focused":true,"cwd":"/home/user/project","foreground_cwd":"/home/user/project","terminal_title":"user@host:~/project","terminal_title_stripped":"user@host:~/project","agent_status":"unknown","scroll":{"offset_from_bottom":0,"max_offset_from_bottom":0,"viewport_rows":40},"revision":3}}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. An offset of 1000000000 clamped to `max_offset_from_bottom` (262). The pass checked the event for a change and no event for no change. It also checked the scrolled `visible` read and the params errors. The example is a new capture.

---

## pane.selection.read

Reads the plain text in a selection range of a pane. The range has an anchor point and a cursor point, which are the two ends of a copy-mode selection.

- Both end points are inclusive. `col` 0 to 10 covers 11 cells.
- The order of `anchor` and `cursor` has no effect. When you swap them, the text is the same.
- herdr removes trailing spaces from the returned text.
- `row` is an absolute scrollback row, not a viewport row. Refer to `PaneTextPoint` in [pane.copy_motion](#panecopy_motion).

**Params** (`PaneSelectionReadParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Target pane. |
| `anchor` | PaneTextPoint ([above](#panecopy_motion)) | yes | — | One end of the selection. |
| `cursor` | PaneTextPoint ([above](#panecopy_motion)) | yes | — | The other end of the selection. |
| `content_revision` | integer (uint64) \| null | no | null | Content revision for the points (inferred). It is the private copy-mode counter of [pane.copy_motion](#panecopy_motion) and [pane.copy_search](#panecopy_search). These methods are the only documented source of a valid value. `PaneInfo.revision` and the `revision` of `pane.read` do not work. A different value fails with `stale_content` (`"pane content changed"`). Null or omitted skips the check. |

**Result**: `type: "pane_selection"`

| field | type | meaning |
|---|---|---|
| `pane_id` | string | Pane that herdr read. |
| `text` | string | Text in the selection. |

**Errors**:

- `pane_not_found` (`"pane not found: w1:p99"`). No other pane method uses this message form.
- `stale_content` when `content_revision` is different. Refer to the table above.
- `selection_unavailable` (`"selection text is unavailable"`) for a point out of range.
- Other codes are possible.

**CLI**: API only. There is no CLI subcommand.

**Example**

Row 0 of the pane is `[user@host project]$`. Columns 0 to 10 cover `[user@host ` with its trailing space. The returned text has 10 characters, because herdr removes the trailing space.

```json
{"id":"sel1","method":"pane.selection.read","params":{"pane_id":"w1:p1","anchor":{"row":0,"col":0},"cursor":{"row":0,"col":10}}}
{"id":"sel1","result":{"type":"pane_selection","pane_id":"w1:p1","text":"[user@host"}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked the inclusive range, the swapped ends, and an absolute scrollback row. It checked `selection_unavailable`, `stale_content` with the three wrong counters, and `pane_not_found`. The example is a new capture.

---

## pane.send_input

Sends text, a sequence of logical keys, or both to a pane in one call. It is the general input method that `send_text` and `send_keys` use.

- Only `pane_id` is required. `text` and `keys` are optional. When a call has both, herdr sends them together.
- herdr sends `text` as a paste. Thus a trailing `\r` or `\n` in it does **not** submit the line. It goes into the line editor of the shell as a literal newline.
- Only an explicit `enter` key in `keys` submits the line.
- `herdr pane run` uses this method. It sends the command text and the key `Enter`. Key names are not case-sensitive.
- `ok` means that herdr accepted the input for the pane. It does not mean that a command ran. [pane run admission](../cli-mapping.md#pane-run-admission) gives this rule for `herdr pane run`, and the macOS line limit.

**Params** (`PaneSendInputParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Target pane. |
| `text` | string | no | — | Literal text to type. |
| `keys` | `array<string>` | no | — | Logical key names to press, for example `enter`, `esc`, or `ctrl+c`. |

**Result**: `type: "ok"`, with no other fields.

**Errors**: `pane_not_found`. An unknown key name gives `invalid_key` (`"unsupported key <name>"`). herdr checks all keys before it writes bytes. An empty `keys` array gives `ok`, not an error. Other codes are possible.

**CLI**: API only. There is no CLI subcommand. The CLI has three commands for it: `send-text`, `send-keys`, and `run`.

**Example**

```json
{"id":"p1","method":"pane.send_input","params":{"pane_id":"w1:p1","text":"true","keys":["enter"]}}
{"id":"p1","result":{"type":"ok"}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked a call with only `pane_id`, the paste of a trailing newline, the Enter key, and the key check before the write. It captured the wire form of `herdr pane run`. While the shell ran `sleep 3`, a call returned `ok` at once, and its output came 2.5 seconds later. The example gave the same response again.

`Documented by herdr --skill (herdr 0.9.5); not live-checked: the macOS line limit. The pass ran on Linux.`

---

## pane.send_keys

Sends a sequence of logical key presses to a pane. herdr checks each key name before it writes bytes. Use `esc` as the canonical name of the Escape key. herdr also accepts `escape`.

**Params** (`PaneSendKeysParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Target pane. |
| `keys` | `array<string>` | yes | — | Logical key names to press in sequence, for example `enter`, `esc`, or `ctrl+c`. |

**Result**: `type: "ok"`, with no other fields.

**Errors**: `pane_not_found`. An unknown key name gives `invalid_key` (`"unsupported key <name>"`). A missing `keys` fails with `invalid_request`. An empty `keys` array gives `ok`. Other codes are possible.

**CLI**: `herdr pane send-keys <PANE_ID> <KEY>...`

**Example**

```json
{"id":"1","method":"pane.send_keys","params":{"pane_id":"w1:p1","keys":["ctrl+c"]}}
{"id":"1","result":{"type":"ok"}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked the key check before the write, `esc` and `escape`, the errors, an empty array, and the wire form of the CLI. The example gave the same response again.

---

## pane.send_text

Sends literal text to the terminal of a pane, with no Enter. To send text and Enter in one request, use the CLI command `run`, which uses `send_input`. A `run` success means only that herdr accepted the input. Refer to [pane.send_input](#panesend_input).

**Params** (`PaneSendTextParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Target pane. |
| `text` | string | yes | — | Literal text to type. |

**Result**: `type: "ok"`, with no other fields.

**Errors**: `pane_not_found`. A missing `text` fails with `invalid_request`. Other codes are possible.

**CLI**: `herdr pane send-text <PANE_ID> <TEXT>`. The command `herdr pane run <PANE_ID> <COMMAND>...` sends text and Enter in one call.

**Example**

```json
{"id":"1","method":"pane.send_text","params":{"pane_id":"w1:p1","text":"echo hello"}}
{"id":"1","result":{"type":"ok"}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked that the text stays unsubmitted until Enter, the errors, and the CLI usage. The example gave the same response again.

---

## pane.split

Splits a pane and makes a new sibling pane with a new shell. With `cwd` or `env`, the shell starts in a given directory or environment.

- The result has the new pane as `result.pane`. Read its ID from `result.pane.pane_id`.
- Keep the focus of the caller with `focus: false`, unless the person wants the focus to move.
- `direction` is required, and only `right` and `down` are valid.
- herdr ignores a `cwd` that does not exist. The new pane starts in the home directory, with no error or warning.
- The call emits `pane_created`, then `layout_updated`. With `focus: true`, it emits `pane_focused` between them.

**Params** (`PaneSplitParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `direction` | SplitDirection (`right`,`down`) | yes | — | Split the target to the right or down. |
| `target_pane_id` | string \| null | no | null | Pane to split. Null uses the UI-focused pane. With null and a `workspace_id`, herdr uses the focused pane of that workspace. |
| `workspace_id` | string \| null | no | null | When `target_pane_id` is null, herdr splits the focused pane of this workspace. herdr ignores it when `target_pane_id` has a value. An unknown workspace fails with `pane_not_found` (`"pane not found"`). |
| `ratio` | number (float) \| null | no | null | First split ratio. herdr clamps it to `[0.1, 0.9]` and does not check it. `5.0` becomes `0.9`, and `-1.0` becomes `0.1`, with no error. Null uses the default. |
| `cwd` | string \| null | no | null | Working directory for the shell of the new pane. herdr ignores a path that does not exist, and the shell starts in the home directory. |
| `env` | `object<string,string>` | no | — | More environment variables for the new process. |
| `right_click` | PaneRightClickTarget (`herdr`,`pane`) | no | `herdr` | Right-click routing for the new pane. The server and the CLI accept it. No read method shows the routing of a pane, so its effect is not checked. Refer to [pane.input.set](#paneinputset). |
| `focus` | boolean | no | false | True focuses the new pane. |

**Result**: `type: "pane_info"`

| field | type | meaning |
|---|---|---|
| `pane` | [PaneInfo](../data-model.md#paneinfo) | The new pane. |

**Errors**: `pane_not_found` (`"pane not found"`). An unknown `direction`, such as `left` or `up`, fails with `invalid_request`. Other codes are possible.

**CLI**: `herdr pane split [--current | --pane <ID> | PANE_ID] --direction <right|down> [--ratio <FLOAT>] [--cwd <PATH>] [--env <KEY=VALUE>] [--right-click <herdr|pane>] [--focus | --no-focus]`

The CLI sends `"right_click":"herdr"` when you do not give `--right-click`. With no `HERDR_PANE_ID`, `--current` fails with `--current requires HERDR_PANE_ID` and exits with status 2. The CLI then sends no request.

**Example**

A new pane has no `terminal_title` yet. The root pane had 40 rows before the split.

```json
{"id":"cli:pane:split","method":"pane.split","params":{"target_pane_id":"w1:p1","direction":"right","cwd":"/home/user/project","focus":false,"right_click":"herdr"}}
{"id":"cli:pane:split","result":{"type":"pane_info","pane":{"pane_id":"w1:p2","terminal_id":"term_65e0de41e119b2","workspace_id":"w1","tab_id":"w1:t1","focused":false,"cwd":"/home/user/project","foreground_cwd":"/home/user/project","agent_status":"unknown","scroll":{"offset_from_bottom":0,"max_offset_from_bottom":0,"viewport_rows":38},"revision":0}}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked these items:

- the default focus, the `direction` values, a `cwd` that does not exist, and `env`
- the events with `focus: false` and `focus: true`
- the UI-focused target, the `workspace_id` target, and the ratio clamp
- the errors, and the wire forms of the CLI

The example is a new capture.

---

## pane.swap

Swaps the positions of two panes in the layout. Give the two panes explicitly with `source_pane_id` and `target_pane_id`. Or give a `direction` from `pane_id`, which selects the next pane in that direction.

- Each field is optional in the schema. But the server needs **exactly one** of the two forms.
- The empty object, a lone `pane_id`, both forms together, and `direction` with `source_pane_id` fail with `invalid_pane_swap`. The message is `"provide either direction with optional pane_id, or source_pane_id and target_pane_id"`.
- A lone `source_pane_id` fails with `invalid_pane_swap` (`"missing target_pane_id"`). A lone `target_pane_id` fails with `"missing source_pane_id"`.
- When the swap cannot occur, `changed` is false and `reason` tells why.
- A successful swap moves the focus to the source pane, or to `pane_id` in the direction form. It emits `layout_updated`. When the focus moves, it also emits `pane_focused`.

**Params** (`PaneSwapParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string \| null | no | null | Reference pane for `direction`. Null uses the UI-focused pane. |
| `direction` | PaneDirection (`left`,`right`,`up`,`down`) \| null | no | null | Swap with the next pane in this direction. |
| `source_pane_id` | string \| null | no | null | Explicit first pane of the swap. |
| `target_pane_id` | string \| null | no | null | Explicit second pane of the swap. |

**Result**: `type: "pane_swap"`. The fields below are one level down, in `swap`:

| field | type | meaning |
|---|---|---|
| `swap` | PaneSwapResult (below) | Result of the swap. |

`PaneSwapResult`:

| field | type | meaning |
|---|---|---|
| `changed` | boolean | True when herdr swapped the panes. |
| `source_pane_id` | string | First pane of the swap. |
| `target_pane_id` | string \| null | Second pane of the swap. herdr omits it, and does not send `null`, when it found no second pane, as for `no_neighbor`. |
| `focused_pane_id` | string | Focused pane in the tab of the swap after the call. It is possibly not the globally focused pane. |
| `layout` | [PaneLayoutSnapshot](../data-model.md#panelayoutsnapshot) | Layout of the tab after the call. |
| `reason` | `no_neighbor` \| `same_pane` \| `not_found` \| `cross_tab` | Why the swap did not occur. herdr omits it, and does not send `null`, when `changed` is `true`. |

**Errors**:

- `invalid_pane_swap`, with the messages above.
- `pane_not_found` (`"source pane not found"`), only through the `direction` form. In the explicit form, an unknown pane is **not** an error. It returns `changed: false, reason: "not_found"`.
- Other codes are possible.

**CLI**: `herdr pane swap [--direction <left|right|up|down>] [--current | --pane <ID>] [--source-pane <ID>] [--target-pane <ID>]`

With no `HERDR_PANE_ID`, `--current` sends no pane ID, so herdr uses the UI-focused pane.

**Example**

```json
{"id":"cli:pane:swap","method":"pane.swap","params":{"source_pane_id":"w1:p1","target_pane_id":"w1:p2"}}
{"id":"cli:pane:swap","result":{"type":"pane_swap","swap":{"changed":true,"source_pane_id":"w1:p1","target_pane_id":"w1:p2","focused_pane_id":"w1:p1","layout":{"workspace_id":"w1","tab_id":"w1:t1","zoomed":false,"area":{"x":0,"y":0,"width":120,"height":40},"focused_pane_id":"w1:p1","panes":[{"pane_id":"w1:p2","focused":false,"rect":{"x":0,"y":0,"width":60,"height":40}},{"pane_id":"w1:p1","focused":true,"rect":{"x":60,"y":0,"width":60,"height":40}}],"splits":[{"id":"split_0_root","direction":"right","ratio":0.5,"rect":{"x":0,"y":0,"width":120,"height":40}}]}}}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked each `invalid_pane_swap` case, the four reasons, and the omitted fields. It checked the focus move and the events, the errors, and the wire forms of the CLI. The example is a new capture.

---

## pane.wait_for_output

Waits until the output of a pane matches a pattern.

- herdr searches the selected snapshot at once, so existing output can match. Then it polls until a match or until `timeout_ms`.
- With no `timeout_ms`, the call waits with no time limit. The connection stays open until the request ends.
- The result has the line that matched and a `PaneReadResult` snapshot.
- `timeout_ms: 0` means "do not wait". When the snapshot does not match already, the call fails with the timeout error almost at once.
- `lines: 0` searches an empty snapshot and never matches.

**Params** (`PaneWaitForOutputParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `pane_id` | string | yes | — | Pane to watch. |
| `source` | ReadSource (`visible`,`recent`,`recent_unwrapped`,`detection`) | yes | — | Snapshot source to search. herdr searches and returns `recent` as `recent_unwrapped`. The `read.source` of the result is `"recent_unwrapped"`, and its text has the soft wraps joined. This applies to an immediate match and to a match from a poll. `pane.read` with `recent` returns `recent`. |
| `match` | OutputMatch | yes | — | Match specification (tagged union, below). |
| `lines` | integer (uint32) \| null | no | null | Limit the searched snapshot to N rows. |
| `strip_ansi` | boolean | no | true | Strip ANSI escapes before the match. The CLI `--raw` sets it to false. |
| `timeout_ms` | integer (uint64) \| null | no | null | Fail with a timeout after this number of ms. Null waits with no time limit. |

`match` is a `oneOf` with `type` as the discriminator:

| variant | fields | meaning |
|---|---|---|
| `substring` | `type` const `"substring"`, `value` string | Match a literal substring. |
| `regex` | `type` const `"regex"`, `value` string | Match a Rust regular expression. |

**Result**: `type: "output_matched"`

| field | type | meaning |
|---|---|---|
| `pane_id` | string | Pane that matched. |
| `revision` | integer (uint64) | The schema tells that this is the content revision of the pane at the time of the match. But each probe gave `0`, on 0.9.1, 0.9.3, and 0.9.5. It follows neither `PaneInfo.revision` nor the copy-mode `content_revision`. Thus no pass could check its documented meaning. |
| `matched_line` | string \| null | The line that matched. The schema allows null, possibly when the match is not on one line (inferred). Each match in the probes gave a value. These matches used substring, regex, each source, `strip_ansi: false`, a poll, and a wrapped line. No pass found an input that gives `null`. |
| `read` | [PaneReadResult](../data-model.md#panereadresult) | Snapshot at the time of the match (`pane_id`, `workspace_id`, `tab_id`, `source`, `format`, `text`, `revision`, `truncated`). `read.revision` is also always `0`. |

**Errors**:

- `pane_not_found`.
- `timeout` (`"timed out waiting for output match"`) when `timeout_ms` ends with no match.
- `invalid_regex` for a `regex` pattern that herdr cannot parse. The message is the full diagnostic of the Rust regex parser, for example `"regex parse error:\n    ([unclosed\n     ^\nerror: unclosed character class"`.
- A `match` type other than `substring` or `regex` fails with `invalid_request`.
- Other codes are possible.

**CLI**: `herdr pane wait-output <PANE_ID> <--match <TEXT> | --regex <PATTERN>> [--source <visible|recent|recent-unwrapped>] [--lines <N>] [--timeout <MS>] [--raw]`

The default CLI source is `recent`. The CLI does not accept `detection`. The CLI always sends `strip_ansi`.

**Example**

```json
{"id":"cli:pane:wait-output","method":"pane.wait_for_output","params":{"pane_id":"w1:p2","source":"recent_unwrapped","match":{"type":"substring","value":"api-docs-marker-42"},"strip_ansi":true}}
{"id":"cli:pane:wait-output","result":{"type":"output_matched","pane_id":"w1:p2","revision":0,"matched_line":"[user@host project]$ echo api-docs-marker-42","read":{"pane_id":"w1:p2","workspace_id":"w1","tab_id":"w1:t1","source":"recent_unwrapped","format":"text","text":"[user@host project]$ echo api-docs-marker-42\napi-docs-marker-42\n[user@host project]$","revision":0,"truncated":false}}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked these items:

- immediate matches for each source, a polled match after 1 second, and a request with no timeout
- `timeout_ms: 0`, a 700 ms timeout, and `lines: 0`
- the `recent` to `recent_unwrapped` result, a wrapped line, and the constant `revision`
- the errors, and the wire forms of the CLI

The example is a new capture.

---

## pane.zoom

Toggles, sets, or clears the **zoom** of a pane. A zoom shows one pane over its full tab, until the zoom goes off.

- `mode` has the default `toggle`.
- A zoom also **focuses** the pane. `mode: "on"` on a pane without the focus gives `focus_changed: true` and moves the focus.
- On a single-pane tab, `mode: "on"` returns `changed: true` and `zoom_changed: false`, only because the focus moved. The tab also becomes the focused tab.
- When the tab has a single pane, or the zoom is already in the requested state, `zoom_changed` is false and `reason` tells why.
- `single_pane` also occurs for `off` and `toggle` on a single pane.
- A `reason` can occur with `changed: true`. While one pane has the zoom, `mode: "on"` for a different pane moves the zoom and the focus to it. It gives `zoom_changed: false` and `reason: "already_zoomed"`.
- A zoom change emits `layout_updated`. A focus change emits `pane_focused`.

**Params** (`PaneZoomParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `mode` | PaneZoomMode (`toggle`,`on`,`off`) | no | `toggle` | Toggle the zoom, set it on, or set it off. |
| `pane_id` | string \| null | no | null | Pane to zoom. Null uses the UI-focused pane. |

**Result**: `type: "pane_zoom"`. The fields below are one level down, in `zoom`:

| field | type | meaning |
|---|---|---|
| `zoom` | PaneZoomResult (below) | Result of the zoom. |

`PaneZoomResult`:

| field | type | meaning |
|---|---|---|
| `changed` | boolean | True when the zoom or the focus changed. Refer to the focus rule above. |
| `zoom_changed` | boolean | True when the zoom state changed. |
| `focus_changed` | boolean | True when the focus changed. |
| `pane_id` | string | Target pane. |
| `focused_pane_id` | string | Focused pane in the tab of the zoomed pane after the call. It is possibly not the globally focused pane. |
| `zoomed` | boolean | Zoom state after the call. |
| `layout` | [PaneLayoutSnapshot](../data-model.md#panelayoutsnapshot) | Layout of the tab after the call. Its `zoomed` is equal to the new state. |
| `reason` | `single_pane` \| `already_zoomed` \| `already_unzoomed` | Why the zoom did not change. herdr omits it, and does not send `null`, when `zoom_changed` is true. |

**Errors**: `pane_not_found` (`"pane not found"`). An unknown `mode` fails with `invalid_request`. Other codes are possible.

**CLI**: `herdr pane zoom [PANE_ID] [--current | --pane <ID>] [--toggle | --on | --off]`

With no flag, the CLI sends `"mode":"toggle"`. With no `HERDR_PANE_ID`, `--current` sends no pane ID, so herdr uses the UI-focused pane.

**Example**

```json
{"id":"cli:pane:zoom","method":"pane.zoom","params":{"pane_id":"w1:p2","mode":"on"}}
{"id":"cli:pane:zoom","result":{"type":"pane_zoom","zoom":{"changed":true,"zoom_changed":true,"focus_changed":true,"pane_id":"w1:p2","focused_pane_id":"w1:p2","zoomed":true,"layout":{"workspace_id":"w1","tab_id":"w1:t1","zoomed":true,"area":{"x":0,"y":0,"width":120,"height":40},"focused_pane_id":"w1:p2","panes":[{"pane_id":"w1:p1","focused":false,"rect":{"x":0,"y":0,"width":60,"height":40}},{"pane_id":"w1:p2","focused":true,"rect":{"x":60,"y":0,"width":60,"height":40}}],"splits":[{"id":"split_0_root","direction":"right","ratio":0.5,"rect":{"x":0,"y":0,"width":120,"height":40}}]}}}}
```

Checked 2026-10-17 against herdr 0.9.5 on a scratch server. The pass checked the default mode, the focus move, the single-pane cases, and all three reasons. It checked a move of the zoom to a different pane, `layout.zoomed`, the UI-focused target, the errors, and the wire forms of the CLI. The example is a new capture.
