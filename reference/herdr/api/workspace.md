# herdr API: workspace methods

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Index: [README.md](../README.md). Wire format: [protocol.md](../protocol.md).

The `workspace` namespace manages the top-level containers of herdr. A workspace holds one or more tabs. Each tab holds one or more panes. Each pane runs a terminal, and herdr can detect a coding agent in that terminal. These methods create, read, focus, rename, move, and close workspaces. They also attach display metadata to a workspace.

Workspace IDs are short opaque strings, such as `w1`, `wA`, `w0`, or `w11`. herdr does not use the ID of a closed workspace again. Treat each ID as opaque. Do not expect a decimal shape. For the ID alphabet and the counter sequence, refer to [ID scheme](../addressing.md#id-scheme). For IDs after a close, refer to [Non-reuse of closed IDs](../addressing.md#non-reuse-of-closed-ids). Checked 2026-10-17 against herdr 0.9.5: a scratch server issued `wZ`, then `w0`, then `w11`, `w12`, and it did not issue `w10` after `w0`. It did not issue a closed workspace ID again.

A read (`workspace.get` or `workspace.list`) does not mark the tab of an agent as seen. Only a focus marks it seen. Checked 2026-10-17 against herdr 0.9.5 with a real agent in a background workspace. After the agent finished its task, `workspace.get` and `workspace.list` both showed `done` two times. Then `workspace.focus` returned `idle`, and a later `workspace.get` also showed `idle`.

Most methods that change state push an event to each connection that subscribes with `events.subscribe`. herdr sends events on a fixed tick of approximately 100 ms, after the response for the call. Refer to [events.md](../events.md). Checked 2026-10-17 against herdr 0.9.5: the first event frame came approximately 120 ms after the call. Events in a sequence came at 100 ms intervals. Each **Events** paragraph below gives two names. The first name is the `events.subscribe` subscription type, which has a dot (for example `workspace.closed`). The second name is the event on the wire, which has an underscore (for example `workspace_closed`). The wire name occurs in the `event` field and in the `data.type` field. Checked 2026-10-17 against herdr 0.9.5.

9 methods (`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`):

| method | purpose |
| --- | --- |
| [workspace.close](#workspaceclose) | Close a workspace and its tabs, panes, and terminals. |
| [workspace.create](#workspacecreate) | Create a workspace with a root tab and a root pane. |
| [workspace.focus](#workspacefocus) | Make a workspace the focused workspace in the UI. |
| [workspace.get](#workspaceget) | Get the current information for one workspace. |
| [workspace.list](#workspacelist) | List all workspaces in display order. |
| [workspace.move](#workspacemove) | Move one workspace to an absolute index in the order. |
| [workspace.move_block](#workspacemove_block) | Move a block of workspaces before a target workspace. |
| [workspace.rename](#workspacerename) | Change the label of a workspace. |
| [workspace.report_metadata](#workspacereport_metadata) | Attach display tokens to a workspace. |

[../data-model.md](../data-model.md) defines the shared entities `WorkspaceInfo`, `TabInfo`, and `PaneInfo`. This page names only their top-level fields. The `agent_status` field has one of these values: `idle`, `working`, `blocked`, `done`, or `unknown`. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`. For the meaning of each value, and for the limits of `idle` and `done`, refer to [AgentStatus](../data-model.md#agentstatus).

## workspace.close

Close a workspace. This closes all of its tabs and panes, and it ends their terminals. herdr does not use the closed workspace, tab, and pane IDs again.

Do not close a workspace that you did not create, unless the user tells you explicitly to close it. `Documented by herdr --skill (herdr 0.9.5); not live-checked: this is an instruction for callers, not server behavior`.

A close through the CLI or the API does not get the TUI last-pane confirmation. Refer to [pane.close](pane.md#paneclose). Checked 2026-10-17 against herdr 0.9.5 on a scratch server with no TUI client: a close of each workspace, the last one too, returned `ok`. Then `workspace.list` returned an empty list. The server continued to run, and a later `workspace.create` worked.

**Params** (`WorkspaceCloseParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `workspace_id` | string | yes | — | ID of the workspace to close (for example `w2`). |
| `close_group` | boolean | no | `false` | If true, also close the linked worktree workspaces of the target (its worktree group). |

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the field types, the required flags, and the default.

A worktree group has a root workspace and one or more linked workspaces. The root workspace opens on the primary checkout of a repository. Each linked workspace opens on a linked worktree of that repository. These rules apply to `close_group`:

- herdr refuses a close of a group root when `close_group` is not `true`. Refer to `workspace_group_close_required` below. Thus `close_group` is necessary for a group root.
- From the group root, `close_group` closes the root and each linked workspace.
- From a linked workspace, `close_group` closes only that linked workspace. The root continues to be open. After its last linked workspace closes, the root closes with a plain close.
- On a workspace with no group, herdr accepts `close_group: true`, and the close is a plain close.
- A group close does not remove the linked Git worktree checkouts. `git worktree list` still shows them after the close.

Checked 2026-10-17 against herdr 0.9.5: each rule in the list above.

`workspace close --group` closes the root workspace and its linked worktree workspaces. Do not add `--group` only to prevent `workspace_group_close_required`. `Documented by herdr --skill (herdr 0.9.5); not live-checked: this is an instruction for callers, not server behavior`.

**Result**: `type: "ok"`.

| field | type | meaning |
| --- | --- | --- |
| `type` | string const `"ok"` | Acknowledgment with no payload. |

**Errors**

| code | when |
| --- | --- |
| `workspace_not_found` | `workspace_id` does not identify a live workspace. The message is `workspace w99 not found` for `w99`. |
| `workspace_group_close_required` | The target is the root of a worktree group, and `close_group` is not `true`. |

Other codes can occur. For all error codes, refer to [errors.md](../errors.md). Checked 2026-10-17 against herdr 0.9.5: both codes and the message.

**Events**: one `workspace_closed` event (subscription type `workspace.closed`). The event holds the full final `WorkspaceInfo` and `workspace_id`. A group close sends one `workspace_closed` event for each closed workspace, and the root event comes first. A workspace close sends no `tab_closed` and no `pane_closed` events, also when the connection subscribes to `tab.closed` and `pane.closed`. Checked 2026-10-17 against herdr 0.9.5.

**CLI**: `herdr workspace close <workspace_id> [--group]`. `--group` sets `close_group`. The usage line of `herdr workspace close --help` does not show `--group`. The usage for a bad invocation shows `--group`, and the full `herdr workspace` command list also shows it. The CLI does not refuse an unknown option such as `--bogus`. It sends that text as the workspace ID, and the server returns `workspace --bogus not found`. Checked 2026-10-17 against herdr 0.9.5. For general CLI rules, refer to [CLI conventions](../cli-mapping.md#cli-conventions).

```text
$ herdr workspace close --help
Usage: herdr workspace close <workspace_id>
$ herdr workspace close
usage: herdr workspace close <workspace_id> [--group]
```

**Example**: a plain close. Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"2","method":"workspace.close","params":{"workspace_id":"w3"}}
{"id":"2","result":{"type":"ok"}}
```

A close of a group root without `close_group`, and then with it. Checked 2026-10-17 against herdr 0.9.5 on a real root workspace with a linked worktree workspace.

```json
{"id":"8","method":"workspace.close","params":{"workspace_id":"w1"}}
{"id":"8","error":{"code":"workspace_group_close_required","message":"workspace has linked worktree workspaces; use --group (close_group=true in the API) to close the group"}}
```

```json
{"id":"3","method":"workspace.close","params":{"workspace_id":"w1","close_group":true}}
{"id":"3","result":{"type":"ok"}}
```

## workspace.create

Create a workspace with one root tab and one root pane. The response gives the IDs to use next: `workspace`, `tab`, and `root_pane`. Checked 2026-10-17 against herdr 0.9.5.

Do not create a workspace unless the user explicitly asks for that topology. `Documented by herdr --skill (herdr 0.9.5); not live-checked: this is an instruction for callers, not server behavior`.

herdr selects the working directory of the new pane in this sequence:

1. `cwd`, if the request gives it.
2. The cwd of the focused pane of `source_workspace_id`, if the request gives it.
3. The cwd of the pane of the focused workspace, if a workspace exists.
4. The home directory of the user, if no workspace exists.

herdr adds each `env` entry to the environment of the new process. Checked 2026-10-17 against herdr 0.9.5: each step of the sequence and the `env` entries. In earlier versions of this page, step 3 was not in the sequence, and a request with no `cwd` and no `source_workspace_id` got the home directory. That case was not tested before 0.9.5, so it is not known if the behavior changed.

A `cwd` that does not exist does not cause an error. The pane uses the home directory, also when a different workspace has focus. The only way to find this fallback is to compare the requested `cwd` with the returned `root_pane.cwd`. The returned path is the canonical path. For example, a `cwd` of `…/plain/../plain` returns `…/plain`, and a path through a symbolic link returns the real path. Checked 2026-10-17 against herdr 0.9.5.

The first workspace on a server always has focus. This is true when `focus` is `false` or has no value, because focus must have a location. It is also true for the first workspace after all workspaces close. When a workspace already exists, `focus` has the default `false`. Checked 2026-10-17 against herdr 0.9.5.

**Params** (`WorkspaceCreateParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `cwd` | string \| null | no | null | Working directory for the process of the root pane. Refer to the sequence and the fallback above. |
| `env` | object (string→string) | no | `{}` | Environment variables for the started process. |
| `focus` | boolean | no | `false` | If true, focus the new workspace in the UI. If false, herdr creates the workspace in the background. |
| `label` | string \| null | no | null | Display label. If null, herdr gives a label. Refer to the label rule below. |
| `source_workspace_id` | string \| null | no | null | Workspace whose focused pane gives the fallback cwd. Refer to the sequence above. |

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the field types, the required flags, and the defaults. Checked 2026-10-17 against herdr 0.9.5: the behavior of each field.

Label rule when `label` is null:

- If the effective cwd is in a Git work tree, the label is the basename of the root of the work tree. For example, a cwd of `…/lbl/deep/er` in the repository `lbl` gives `lbl`.
- If the effective cwd is not in a Git work tree, the label is the basename of the cwd.
- For the home directory, the label is `~`.
- The Git check uses the `cwd` as the request gives it. A `cwd` with a `..` part or a symbolic link part gives the basename of the cwd, not the repository name. This occurs although `root_pane.cwd` is the canonical path. For example, `…/lbl/deep/../deep/er` gives `er`. Give a canonical path to get the repository name.

Checked 2026-10-17 against herdr 0.9.5: each rule in the list above. The last rule can be older than 0.9.5.

**Result**: `type: "workspace_created"`.

| field | type | meaning |
| --- | --- | --- |
| `type` | string const `"workspace_created"` | Result discriminant. |
| `workspace` | WorkspaceInfo | The new workspace ([../data-model.md](../data-model.md#workspaceinfo)). |
| `tab` | TabInfo | The root tab of the workspace ([../data-model.md](../data-model.md#tabinfo)). |
| `root_pane` | PaneInfo | The root pane of the root tab ([../data-model.md](../data-model.md#paneinfo)). |

**Errors**:

- `invalid_request`: the input is not correct, for example a field with the wrong JSON type. A rejection at the schema level is always `invalid_request` with a serde message, never `invalid_params`.
- The response contains the `id` of the request when the line is valid JSON with a string `id`. For an integer `id` or JSON that is not complete, the response has `"id":""`. For a valid JSON line with a string `id`, version 0.9.1 returned `"id":""`.
- `workspace_not_found`: `source_workspace_id` does not identify a live workspace.
- Other codes can occur. Refer to [`invalid_request` variants](../errors.md#invalid_request-variants).

Checked 2026-10-17 against herdr 0.9.5:

```json
{"id":"8","method":"workspace.create","params":{"focus":"yes"}}
{"id":"8","error":{"code":"invalid_request","message":"invalid request: invalid type: string \"yes\", expected a boolean at line 1 column 63"}}
```

**Events**: herdr sends all these events in one tick:

- Background create: `workspace_created` (subscription type `workspace.created`), `tab_created` (`tab.created`), `pane_created` (`pane.created`), `pane_updated` (`pane.updated`), `layout_updated` (`layout.updated`).
- Focused create (for example the first workspace on a server): the events of a background create, and also `workspace_focused`, `tab_focused`, and `pane_focused`.

The sequence of the events in the tick follows the sequence of the subscription list of the connection. Two subscribers with different lists got different sequences for the same create. Checked 2026-10-17 against herdr 0.9.5.

The `pane_updated` event has `revision` 1, and `pane_created` has `revision` 0. In a window of 3 s, no other `pane_updated` event came. Checked 2026-10-17 against herdr 0.9.5. For 0.9.3, earlier versions of this page put `pane_updated` after `layout_updated`, when the root shell became stable. It is not known if the behavior changed.

**CLI**: `herdr workspace create [--cwd PATH] [--label TEXT] [--env KEY=VALUE] [--focus] [--no-focus]`. Checked 2026-10-17 against herdr 0.9.5.

**Example**: the first workspace, with focus. Checked 2026-10-17 against herdr 0.9.5. A raw socket response keeps the field sequence of the server, with `workspace_id` first. The CLI prints the keys in alphabetical sequence.

```json
{"id":"1","method":"workspace.create","params":{"cwd":"/home/user/project","label":"api-docs","focus":true}}
{"id":"1","result":{"type":"workspace_created","workspace":{"workspace_id":"w1","number":1,"label":"api-docs","focused":true,"pane_count":1,"tab_count":1,"active_tab_id":"w1:t1","agent_status":"unknown"},"tab":{"tab_id":"w1:t1","workspace_id":"w1","number":1,"label":"1","focused":true,"pane_count":1,"agent_status":"unknown"},"root_pane":{"pane_id":"w1:p1","terminal_id":"term_65e0dc57886751","workspace_id":"w1","tab_id":"w1:t1","focused":true,"cwd":"/home/user/project","foreground_cwd":"/home/user/project","agent_status":"unknown","scroll":{"offset_from_bottom":0,"max_offset_from_bottom":0,"viewport_rows":40},"revision":0}}}
```

`source_workspace_id` gives the cwd of `w1`. Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"2","method":"workspace.create","params":{"label":"build","source_workspace_id":"w1"}}
{"id":"2","result":{"type":"workspace_created","workspace":{"workspace_id":"w2","number":2,"label":"build","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"w2:t1","agent_status":"unknown"},"tab":{"tab_id":"w2:t1","workspace_id":"w2","number":1,"label":"1","focused":false,"pane_count":1,"agent_status":"unknown"},"root_pane":{"pane_id":"w2:p1","terminal_id":"term_65e0dc57889342","workspace_id":"w2","tab_id":"w2:t1","focused":false,"cwd":"/home/user/project","foreground_cwd":"/home/user/project","agent_status":"unknown","scroll":{"offset_from_bottom":0,"max_offset_from_bottom":0,"viewport_rows":40},"revision":0}}}
```

A `cwd` that does not exist gives the home directory. Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"4","method":"workspace.create","params":{"cwd":"/nonexistent/dir","label":"build"}}
{"id":"4","result":{"type":"workspace_created","workspace":{"workspace_id":"w4","number":4,"label":"build","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"w4:t1","agent_status":"unknown"},"tab":{"tab_id":"w4:t1","workspace_id":"w4","number":1,"label":"1","focused":false,"pane_count":1,"agent_status":"unknown"},"root_pane":{"pane_id":"w4:p1","terminal_id":"term_65e0dc5788dbd4","workspace_id":"w4","tab_id":"w4:t1","focused":false,"cwd":"/home/user","foreground_cwd":"/home/user","agent_status":"unknown","scroll":{"offset_from_bottom":0,"max_offset_from_bottom":0,"viewport_rows":40},"revision":0}}}
```

## workspace.focus

Make the target workspace the focused workspace in the herdr UI. A focus of a workspace, or of its tab or pane, marks the tab of the agent as seen. This changes a `done` status to `idle`. A CLI read does not mark a tab as seen. Refer to [AgentStatus](../data-model.md#agentstatus). Checked 2026-10-17 against herdr 0.9.5 with a real agent. The focus result already shows the status after the seen mark (`idle`).

Create workspaces in the background. Do not focus a workspace unless the user asks for a change of context. `Documented by herdr --skill (herdr 0.9.5); not live-checked: this is an instruction for callers, not server behavior`.

**Params** (`WorkspaceTarget`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `workspace_id` | string | yes | — | ID of the workspace to focus. |

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`.

**Result**: `type: "workspace_info"`.

| field | type | meaning |
| --- | --- | --- |
| `type` | string const `"workspace_info"` | Result discriminant. |
| `workspace` | WorkspaceInfo | The workspace that now has focus ([../data-model.md](../data-model.md#workspaceinfo)). |

**Errors**

| code | when |
| --- | --- |
| `workspace_not_found` | `workspace_id` does not identify a live workspace. The CLI exits with status 1. |

Other codes can occur. Checked 2026-10-17 against herdr 0.9.5.

**Events**: one `workspace_focused` event (subscription type `workspace.focused`). The event holds only `{workspace_id}`, not a full `WorkspaceInfo`. The same tick also holds `tab_focused` and `pane_focused` for the active tab and pane of the workspace. A focus of a workspace that already has focus sends no event. Thus a caller that sends a focus and then waits for `workspace_focused` can wait with no end when the target already had focus. Checked 2026-10-17 against herdr 0.9.5.

**CLI**: `herdr workspace focus <workspace_id>`. Checked 2026-10-17 against herdr 0.9.5.

**Example**: a workspace with a shell pane. Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"1","method":"workspace.focus","params":{"workspace_id":"w2"}}
{"id":"1","result":{"type":"workspace_info","workspace":{"workspace_id":"w2","number":4,"label":"build","focused":true,"pane_count":1,"tab_count":1,"active_tab_id":"w2:t1","agent_status":"unknown"}}}
```

A workspace with a real agent. Before the focus, the status was `done`. Checked 2026-10-17 against herdr 0.9.5.

```text
$ herdr workspace focus w1E
{"id":"cli:workspace:focus","result":{"type":"workspace_info","workspace":{"active_tab_id":"w1E:t1","agent_status":"idle","focused":true,"label":"reviewer","number":12,"pane_count":1,"tab_count":1,"workspace_id":"w1E"}}}
```

## workspace.get

Get the current information for one workspace. This is a read. It does not mark the tab of the workspace as seen. Checked 2026-10-17 against herdr 0.9.5.

**Params** (`WorkspaceTarget`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `workspace_id` | string | yes | — | ID of the workspace to get. |

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`.

**Result**: `type: "workspace_info"`.

| field | type | meaning |
| --- | --- | --- |
| `type` | string const `"workspace_info"` | Result discriminant. |
| `workspace` | WorkspaceInfo | The requested workspace ([../data-model.md](../data-model.md#workspaceinfo)). |

**Errors**

| code | when |
| --- | --- |
| `workspace_not_found` | `workspace_id` does not identify a live workspace. The CLI exits with status 1. |

Other codes can occur. Checked 2026-10-17 against herdr 0.9.5.

**CLI**: `herdr workspace get <workspace_id>`. Checked 2026-10-17 against herdr 0.9.5.

**Example**: Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"3","method":"workspace.get","params":{"workspace_id":"w2"}}
{"id":"3","result":{"type":"workspace_info","workspace":{"workspace_id":"w2","number":2,"label":"build","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"w2:t1","agent_status":"unknown"}}}
```

The `workspace_not_found` error. Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"4","method":"workspace.get","params":{"workspace_id":"w99"}}
{"id":"4","error":{"code":"workspace_not_found","message":"workspace w99 not found"}}
```

The CLI prints the same error with the keys in alphabetical sequence. Checked 2026-10-17 against herdr 0.9.5.

```text
$ herdr workspace get w99
{"error":{"code":"workspace_not_found","message":"workspace w99 not found"},"id":"cli:workspace:get"}
```

## workspace.list

List all workspaces in display order. This is a read. It does not mark a tab as seen. Checked 2026-10-17 against herdr 0.9.5.

The list can contain one more workspace than a caller expects. This occurs after a [`worktree.create`](worktree.md#worktreecreate) call with these conditions:

- The `cwd` is the primary checkout of a repository.
- No workspace is open on that checkout.
- The call gives no `workspace_id`.

The call returns the linked worktree workspace. It also opens a second workspace for the primary checkout, with no indication. The response does not show that second workspace. Only a later `workspace.list` or `workspace.get` shows it. Checked 2026-10-17 against herdr 0.9.5: the list grew from 13 to 15 workspaces after one such call.

**Params** (`EmptyParams`): none. Send `params: {}`. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`.

**Result**: `type: "workspace_list"`.

| field | type | meaning |
| --- | --- | --- |
| `type` | string const `"workspace_list"` | Result discriminant. |
| `workspaces` | array of WorkspaceInfo | All workspaces, in display order ([../data-model.md](../data-model.md#workspaceinfo)). |

**Errors**: other codes can occur. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`.

**CLI**: `herdr workspace list`. Checked 2026-10-17 against herdr 0.9.5.

**Example**: the list includes a `tokens` object and two `worktree` objects. `w4` is not in the list because it closed before this call. Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"2","method":"workspace.list","params":{}}
{"id":"2","result":{"type":"workspace_list","workspaces":[{"workspace_id":"w1","number":1,"label":"api-docs","focused":true,"pane_count":1,"tab_count":1,"active_tab_id":"w1:t1","agent_status":"unknown","tokens":{"branch":"main"},"worktree":{"repo_key":"/home/user/project/.git","repo_name":"project","repo_root":"/home/user/project","checkout_path":"/home/user/project","is_linked_worktree":false}},{"workspace_id":"w2","number":2,"label":"build","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"w2:t1","agent_status":"unknown"},{"workspace_id":"w3","number":3,"label":"reviewer","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"w3:t1","agent_status":"unknown"},{"workspace_id":"w5","number":4,"label":"feature","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"w5:t1","agent_status":"unknown","worktree":{"repo_key":"/home/user/project/.git","repo_name":"project","repo_root":"/home/user/project","checkout_path":"/home/user/.herdr/worktrees/project/feature","is_linked_worktree":true}}]}}
```

## workspace.move

Move one workspace to a position in the display order. herdr applies `insert_index` to the order **before** the move, as [`tab.move`](tab.md#tabmove) does. herdr puts the workspace before the item at that index, and then it removes the old slot. Thus a forward move puts the workspace one slot before the index that the request gives. These examples show the effect:

- A move of the first workspace to `insert_index: 2` put it at index 1.
- A move of a workspace to its own index, or to the next index, does not change the order.

The accepted range is `0..=N`. `N` is the number of workspaces, with the moved workspace included.

- `insert_index=N` puts the workspace last.
- `N-1` puts it second from last, when it starts at an earlier position.
- herdr refuses `N+1`. Refer to `workspace_move_failed` below.

The response gives the full list in the new order. The `number` fields show the new positions. Checked 2026-10-17 against herdr 0.9.5 with 13 workspaces.

**Params** (`WorkspaceMoveParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `workspace_id` | string | yes | — | ID of the workspace to move. |
| `insert_index` | integer (uint, ≥ 0) | yes | — | Zero-based target index in the display order **before** the move. The valid range is `0..=N`. |

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the field types and the required flags.

**Result**: `type: "workspace_list"`.

| field | type | meaning |
| --- | --- | --- |
| `type` | string const `"workspace_list"` | Result discriminant. |
| `workspaces` | array of WorkspaceInfo | The full list in the new order ([../data-model.md](../data-model.md#workspaceinfo)). |

**Errors**:

- `workspace_not_found`: `workspace_id` does not identify a live workspace. This error has priority over an index that is out of bounds.
- `workspace_move_failed`: `insert_index` is out of bounds. For example, the message is `"insert_index 14 is out of bounds"` with 13 workspaces.
- `invalid_request`: `insert_index` is negative. The schema layer refuses it, because `insert_index` is an unsigned integer. The message contains `expected usize`.
- Other codes can occur.

Checked 2026-10-17 against herdr 0.9.5: each code in the list above.

**Events**: exactly one `workspace_moved` event (subscription type `workspace.moved`). The event holds `{workspace_id, insert_index, workspaces}`. `workspace.move` never sends `workspace_reordered`, because that event is for `workspace.move_block` (below). A move that does not change the order sends no event.

herdr also sends a `workspace_focused` event in the same tick for the workspace that has focus. It sends this event also when focus did not change, and when the moved workspace does not have focus. Matching `tab_focused` and `pane_focused` events come with it. Checked 2026-10-17 against herdr 0.9.5.

**CLI**: API-only (no CLI command). The `herdr workspace` command group lists only list, create, get, focus, rename, report-metadata, and close. Checked 2026-10-17 against herdr 0.9.5.

**Example**: Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"6","method":"workspace.move","params":{"workspace_id":"w2","insert_index":0}}
{"id":"6","result":{"type":"workspace_list","workspaces":[{"workspace_id":"w2","number":1,"label":"build","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"w2:t1","agent_status":"unknown"},{"workspace_id":"w1","number":2,"label":"api-docs","focused":true,"pane_count":1,"tab_count":1,"active_tab_id":"w1:t1","agent_status":"unknown","tokens":{"branch":"main"},"worktree":{"repo_key":"/home/user/project/.git","repo_name":"project","repo_root":"/home/user/project","checkout_path":"/home/user/project","is_linked_worktree":false}},{"workspace_id":"w3","number":3,"label":"reviewer notes","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"w3:t1","agent_status":"unknown"},{"workspace_id":"w5","number":4,"label":"feature","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"w5:t1","agent_status":"unknown","worktree":{"repo_key":"/home/user/project/.git","repo_name":"project","repo_root":"/home/user/project","checkout_path":"/home/user/.herdr/worktrees/project/feature","is_linked_worktree":true}}]}}
```

## workspace.move_block

Move a block of workspaces to the position immediately before a target workspace. After the move, the block is contiguous, and it keeps the sequence of `workspace_ids`. The IDs do not have to be adjacent before the move. "Contiguous" describes the result, not a precondition. herdr puts IDs that are not adjacent together at the target position. When `before_workspace_id` is null or has no value, herdr moves the block to the end of the display order. The response gives the full list in the new order. Checked 2026-10-17 against herdr 0.9.5.

**Params** (`WorkspaceMoveBlockParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `workspace_ids` | array of string | yes | — | The workspace IDs to move as a block. The block keeps their sequence. |
| `before_workspace_id` | string \| null | no | null | Put the block immediately before this workspace. Null puts the block at the end. |

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the field types, the required flags, and the default.

**Result**: `type: "workspace_list"`.

| field | type | meaning |
| --- | --- | --- |
| `type` | string const `"workspace_list"` | Result discriminant. |
| `workspaces` | array of WorkspaceInfo | The full list in the new order ([../data-model.md](../data-model.md#workspaceinfo)). |

**Errors**:

| code | when | message |
| --- | --- | --- |
| `workspace_not_found` | An ID in `workspace_ids` or in `before_workspace_id` does not identify a live workspace. | `workspace w99 not found` |
| `workspace_move_block_failed` | `workspace_ids` is empty. | `workspace_ids must not be empty` |
| `workspace_move_block_failed` | An ID occurs more than one time in `workspace_ids`. | `workspace w1 appears more than once` |
| `workspace_move_block_failed` | `before_workspace_id` is in `workspace_ids`. | `before_workspace_id must not be part of workspace_ids` |

Other codes can occur. Checked 2026-10-17 against herdr 0.9.5: all five error cases (the first row has two cases).

**Events**: exactly one `workspace_reordered` event (subscription type `workspace.reordered`). The event holds `{workspace_ids, before_workspace_id, workspaces}`. The event has no `before_workspace_id` field when the request gave `null`. `workspace.move_block` never sends `workspace_moved`, because that event is for `workspace.move` (above).

As with `workspace.move`, herdr also sends `workspace_focused`, `tab_focused`, and `pane_focused` in the same tick for the workspace that has focus. It sends them also when focus did not change. Two `workspace.move_block` calls in one tick sent two `workspace_reordered` events and one set of focus events. Checked 2026-10-17 against herdr 0.9.5.

**CLI**: API-only (no CLI command). Checked 2026-10-17 against herdr 0.9.5.

**Example**: Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"7","method":"workspace.move_block","params":{"workspace_ids":["w2"],"before_workspace_id":null}}
{"id":"7","result":{"type":"workspace_list","workspaces":[{"workspace_id":"w1","number":1,"label":"api-docs","focused":true,"pane_count":1,"tab_count":1,"active_tab_id":"w1:t1","agent_status":"unknown","tokens":{"branch":"main"},"worktree":{"repo_key":"/home/user/project/.git","repo_name":"project","repo_root":"/home/user/project","checkout_path":"/home/user/project","is_linked_worktree":false}},{"workspace_id":"w3","number":2,"label":"reviewer notes","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"w3:t1","agent_status":"unknown"},{"workspace_id":"w5","number":3,"label":"feature","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"w5:t1","agent_status":"unknown","worktree":{"repo_key":"/home/user/project/.git","repo_name":"project","repo_root":"/home/user/project","checkout_path":"/home/user/.herdr/worktrees/project/feature","is_linked_worktree":true}},{"workspace_id":"w2","number":4,"label":"build","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"w2:t1","agent_status":"unknown"}]}}
```

## workspace.rename

Change the display label of a workspace. The response gives the updated workspace information. Checked 2026-10-17 against herdr 0.9.5.

**Params** (`WorkspaceRenameParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `workspace_id` | string | yes | — | ID of the workspace to rename. |
| `label` | string | yes | — | The new display label. herdr does not check the length or the content. It accepts an empty string and a string of 500 characters, and it returns each without a change. |

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the field types and the required flags. Checked 2026-10-17 against herdr 0.9.5: the empty label and the 500-character label.

**Result**: `type: "workspace_info"`.

| field | type | meaning |
| --- | --- | --- |
| `type` | string const `"workspace_info"` | Result discriminant. |
| `workspace` | WorkspaceInfo | The renamed workspace ([../data-model.md](../data-model.md#workspaceinfo)). |

**Errors**: `workspace_not_found` (`workspace_id` does not identify a live workspace). Other codes can occur. Checked 2026-10-17 against herdr 0.9.5.

**Events**: one `workspace_renamed` event (subscription type `workspace.renamed`). The event holds only `{workspace_id, label}`, not a full `WorkspaceInfo`. Unlike `workspace.focus`, a rename to the current label still sends the event. Checked 2026-10-17 against herdr 0.9.5.

**CLI**: `herdr workspace rename <workspace_id> <label>`. The CLI accepts the label as one or more trailing arguments: `rename <WORKSPACE_ID> <LABEL>...`. For example, `herdr workspace rename w7 my build label` gives the label `my build label`. The CLI does not parse `--label` here. `herdr workspace rename w7 --label renamed` gives the label `--label renamed` and exits with status 0. Checked 2026-10-17 against herdr 0.9.5.

**Example**: Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"5","method":"workspace.rename","params":{"workspace_id":"w3","label":"reviewer notes"}}
{"id":"5","result":{"type":"workspace_info","workspace":{"workspace_id":"w3","number":3,"label":"reviewer notes","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"w3:t1","agent_status":"unknown"}}}
```

## workspace.report_metadata

Attach display metadata to a workspace. The metadata is a set of short string tokens, each with a name. Tokens are only display state. They occur in `WorkspaceInfo.tokens`, for example `{"branch":"main"}`. They do not change the topology or the agent status.

These rules apply to tokens:

- Each workspace has one shared set of tokens. The name `source` does not make a source the owner of its token values. Any source can set, overwrite, or clear a token that a different source set.
- `source` controls only the `seq` counter (below).
- A token with the value `null` clears that token. A clear of a token that is not set does nothing. It returns `ok` and sends no `workspace_metadata_updated` event.
- When a report clears the last token of a workspace, `WorkspaceInfo` has no `tokens` key. It does not show `{}`. The event `WorkspaceInfo` also has no `tokens` key.
- `seq` sets a sequence for each source. With `seq`, the server ignores reports that come late.
- `ttl_ms` sets the time after which a token expires. The server starts the expiry. It clears the token and sends `workspace_metadata_updated` with no client request.
- The sequence of keys in `tokens` is not stable.

Checked 2026-10-17 against herdr 0.9.5: each rule in the list above. With a `ttl_ms` of 300, the expiry event came approximately 420 ms after the report.

**Params** (`WorkspaceReportMetadataParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `workspace_id` | string | yes | — | ID of the workspace that gets the tokens. |
| `source` | string | yes | — | Name of the reporter. It must not be empty. `""` fails with `invalid_metadata_source`. It controls only the `seq` counter, not the token values. Refer to the rules above. |
| `tokens` | object (string→(string \| null)) | yes | — | Token map. It must not be empty. `{}` fails with `invalid_metadata_token` (`"missing token to set or clear"`). Each key must match `^[A-Za-z0-9_-]{1,32}$`. A null value clears that token. |
| `seq` | integer (uint64, ≥ 0) \| null | no | null | Sequence number for this source. It must be more than the last `seq` of this source. The server ignores a report with an equal or lower `seq`. It still returns `{"type":"ok"}`, with no error and no `workspace_metadata_updated` event. |
| `ttl_ms` | integer (uint64) \| null | no | null | Token lifetime in milliseconds, from 1 to 86400000 (24 h). Null gives no explicit expiry. |

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the field types, the required flags, and the defaults. Checked 2026-10-17 against herdr 0.9.5: each constraint in the meaning column.

Token limits:

- One report can change a maximum of 16 tokens.
- A workspace can hold a maximum of 32 tokens from all reports. The schema does not show this limit. Refer to `metadata_token_limit` below.
- At the limit of 32, a report can overwrite a token that exists.
- At the limit of 32, one report can clear one token and add a new token, for example `{"a0":null,"one_more":"x"}`.

Checked 2026-10-17 against herdr 0.9.5: each limit in the list above.

**Result**: `type: "ok"`.

| field | type | meaning |
| --- | --- | --- |
| `type` | string const `"ok"` | Acknowledgment with no payload. |

**Errors**:

| code | when | message |
| --- | --- | --- |
| `workspace_not_found` | `workspace_id` does not identify a live workspace. | `workspace w99 not found` |
| `invalid_metadata_source` | `source` is empty. | `metadata source must not be empty` |
| `invalid_metadata_token` | A token key does not match `^[A-Za-z0-9_-]{1,32}$`, for example a key with a space or a key of 33 characters. | `invalid metadata token key: <key>` |
| `invalid_metadata_token` | One report changes more than 16 tokens. | `a metadata report may update at most 16 tokens` |
| `invalid_metadata_token` | `tokens` is empty. | `missing token to set or clear` |
| `metadata_token_limit` | A report adds a new token, and the workspace already holds 32 tokens. | `workspace metadata may contain at most 32 tokens` |
| `invalid_metadata_ttl` | `ttl_ms` is less than 1. | `metadata ttl_ms must be at least 1` |
| `invalid_metadata_ttl` | `ttl_ms` is more than 86400000. | `metadata ttl_ms must be 86400000 or less` |

Other codes can occur. No error in this namespace is `invalid_params`. Checked 2026-10-17 against herdr 0.9.5: each row.

**Events**: one `workspace_metadata_updated` event (subscription type `workspace.metadata_updated`). The event holds the full `WorkspaceInfo` after the update, with `tokens`. Most other workspace events hold only an ID. Checked 2026-10-17 against herdr 0.9.5.

**CLI**: `herdr workspace report-metadata <workspace_id> --source ID [--token NAME=VALUE] [--clear-token NAME] [--seq N] [--ttl-ms N]`. `--token` sets a token. `--clear-token` sends a null value for that token. These rules apply to the CLI:

- On success, the CLI prints nothing.
- With no `--token` and no `--clear-token`, the CLI fails before it sends a request. It writes `missing token to set or clear` to stderr, prints no JSON, and exits with status 2.

Checked 2026-10-17 against herdr 0.9.5.

**Example**: Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"1","method":"workspace.report_metadata","params":{"workspace_id":"w1","source":"my-tool","tokens":{"branch":"main"},"ttl_ms":60000}}
{"id":"1","result":{"type":"ok"}}
```

A report that adds a token to a workspace with 32 tokens. Checked 2026-10-17 against herdr 0.9.5.

```json
{"id":"2","method":"workspace.report_metadata","params":{"workspace_id":"w2","source":"my-tool","tokens":{"one_more":"x"}}}
{"id":"2","error":{"code":"metadata_token_limit","message":"workspace metadata may contain at most 32 tokens"}}
```
