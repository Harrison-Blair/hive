# herdr API: worktree methods

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Index: [README.md](../README.md). Wire format: [protocol.md](../protocol.md).

The `worktree.*` methods manage workspaces that a Git worktree backs. They list the
worktrees of the repository that owns a checkout. They create a new linked worktree and
open it as a herdr workspace in one step. They open an existing worktree as a workspace.
They remove a linked worktree checkout.

Each method finds its target repository from a working directory (`cwd`) or from an
existing `workspace_id`. When the request gives neither, the server uses the *focused*
workspace. When no workspace is focused, the method fails with `invalid_request`
(`workspace_id or cwd is required when no workspace is active`).

`worktree.create` and `worktree.open` also require that the source is the parent
(not linked) checkout of the repository. A linked worktree as the source fails with
`linked_worktree_source` (`New and open worktree actions start from the repo parent workspace.`).
This rule applies to a source from `cwd`, from `workspace_id`, or from the focused
workspace. `worktree.list` does not have this rule. It uses the parent checkout of the
repository and succeeds.

Each `cwd` and `path` must be absolute. The server expands a leading `~/` or a bare `~`
against its home directory. A relative path, an empty string, or a `~user/…` form fails
with `invalid_request` (`worktree path must be absolute`).

A worktree becomes a herdr *workspace* when herdr opens it. Thus the create and open
results give the full workspace, tab, and pane topology (see
[data-model.md](../data-model.md)). They also give the `WorktreeInfo`.

`trust_repository` applies to one request only. A trusted call does not make later
untrusted calls succeed. It writes nothing to the global Git config of the user.

Create, open, and remove also send the `worktree_created`, `worktree_opened`, and
`worktree_removed` events to subscriptions. `worktree.open` sends `worktree_opened` also
when the worktree is already open. herdr paces event delivery on a tick of approximately
100 ms. Thus a frame can arrive after the result that it describes. See
[events.md](../events.md).

Checked 2026-10-17 against herdr 0.9.5: a fresh scratch server with no workspace, linked
sources from `cwd`, `workspace_id`, and the focused workspace, each path form, an untrusted
repository, and an events subscription. Frames arrived in bursts after the results. This
pass did not measure the 100 ms tick.

4 methods:

| method | purpose |
|---|---|
| [worktree.create](#worktreecreate) | Create a new linked Git worktree and open it as a workspace. |
| [worktree.list](#worktreelist) | List the worktrees of the repository that owns a checkout. |
| [worktree.open](#worktreeopen) | Open an existing Git worktree as a workspace. |
| [worktree.remove](#worktreeremove) | Remove a worktree checkout by the ID of its open workspace. |

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the method list.
`herdr worktree --help` shows the same four commands.

## worktree.create

Creates a new linked Git worktree for the target repository and opens it as a herdr
workspace in one step. The result gives the new `WorkspaceInfo`, its first `TabInfo`,
its root `PaneInfo`, and the `WorktreeInfo` of the checkout. The server finds the target
repository from `workspace_id` or `cwd`. When the request gives neither, the server uses
the focused workspace. That workspace must be the parent checkout of the repository, as
an explicit source must be.

If the request omits `branch`, the server makes a branch name of the form
`worktree/<adjective>-<noun>-<4 hex>`. Observed names include `worktree/green-river-c8c3`,
`worktree/silver-valley-39ff`, and `worktree/lucky-meadow-a3ac`. `base` sets the ref from
which the server creates the new branch and worktree.

`path` sets a different checkout location. The default location is `<dir>/<repo>/<branch>`
in the managed worktree directory of herdr. `<dir>` is the `[worktrees] directory` value of
the config. Its default is `~/.herdr/worktrees`, per
[raw/default-config.toml](../raw/default-config.toml). In `<branch>`, herdr changes each
`/` to `-`. For example, branch `feature/slash-a` goes to `<dir>/project/feature-slash-a`.

`label` sets the workspace label. When the request omits it, the label is the **basename
of the checkout directory**, not the branch. For example, branch `zebra` at `…/wt/xyzdir`
gives the workspace label `xyzdir`. With the managed default path, the label and the branch
are the same after herdr changes each `/` to `-`. For example, branch `feature/slash-a`
gives the label `feature-slash-a`.

`focus` sets if the new workspace is focused after the server creates it.
`trust_repository` gives Git trust for this one request. Use it only after a person checks
the repository. Do not use it as a routine retry for a failed worktree command.

`Documented by herdr --skill (herdr 0.9.5); not live-checked: the advice about when to use trust_repository is guidance, not behavior`.

The first `worktree.create` in a repository that herdr does not know opens **two**
workspaces. One workspace adopts the parent checkout of the repository. The other
workspace holds the new worktree. When no workspace was focused before the call, herdr
focuses the adopted parent workspace.

On herdr 0.9.5, a subscriber received these events from that first create:

- `workspace_created` for the parent workspace, then for the new worktree workspace.
- `workspace_focused` for the parent workspace, only when no workspace was focused before.
- `worktree_created` for the new workspace.
- `tab_created` for the first tab of each workspace.

herdr groups the frames of one burst by event kind. In the checks, the order of the kinds
followed the order of the `subscriptions` array in `events.subscribe`. Thus do not rely on
the order of these events. See [events.md](../events.md#ordering-and-grouping).
For herdr 0.9.3, this page recorded `worktree_created` before the two `workspace_created`
events. That pass may have used a different order of subscriptions.

**Params** (`WorktreeCreateParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `workspace_id` | string \| null | no | null | Existing workspace whose repository is the source. It must not be the workspace of a linked worktree. |
| `cwd` | string \| null | no | null | Absolute working directory that resolves the source repository when `workspace_id` is absent. It must not be inside a linked worktree. |
| `branch` | string \| null | no | null | Branch name for the new worktree. When it is null, the server makes a name of the form `worktree/<adjective>-<noun>-<4 hex>`. An empty string fails with `invalid_request` (`branch is required`). |
| `base` | string \| null | no | null | Git ref from which the server creates the new branch and worktree. |
| `path` | string \| null | no | null | Explicit absolute checkout path. When it is null, the server uses its managed location. |
| `label` | string \| null | no | null | Workspace label. When it is null, the label is the basename of the checkout directory. The server accepts an empty string and gives an empty label. |
| `focus` | boolean | no | `false` | When it is true, the server focuses the created workspace. |
| `trust_repository` | boolean | no | `false` | Gives Git trust for this request only. It bypasses the ownership check of Git for the repository. |

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the type, required, and
default columns.

**Result** — `type: "worktree_created"`:

| field | type | meaning |
|---|---|---|
| `type` | string const `"worktree_created"` | Result discriminator. |
| `workspace` | WorkspaceInfo | The created workspace. See [data-model.md](../data-model.md#workspaceinfo). |
| `tab` | TabInfo | The first tab of the workspace. See [data-model.md](../data-model.md#tabinfo). |
| `root_pane` | PaneInfo | The root pane of the first tab. See [data-model.md](../data-model.md#paneinfo). |
| `worktree` | WorktreeInfo | The created worktree checkout. See [WorktreeInfo](#worktreeinfo) below. |

**Errors**:

- `worktree_create_failed`: Git refused the request. The branch is checked out in another
  worktree, the target path is occupied, the branch name is not valid, `base` is unknown,
  or the repository is not trusted. herdr gives the message of Git as it is, for example
  `"fatal: 'a..b' is not a valid branch name"`. Some Git versions add `hint:` lines to
  that message. When the path is occupied, Git already created the new branch. That
  branch stays after the failure.
- `worktree_operation_in_progress`: another create or remove for the same target checkout
  path is in progress. The message is
  `worktree operation is already in progress for this checkout`. One of the concurrent
  requests succeeds. A request that starts after the first create ends can fail with
  `worktree_create_failed` instead.
- `linked_worktree_source`: the source checkout is a linked worktree. This includes a
  source from the focused workspace.
- `not_git_worktree`: `cwd` is not inside a Git work tree.
- `workspace_not_found`: `workspace_id` is unknown.
- `invalid_request`: `branch` is empty, or `cwd` or `path` is not absolute.

The result of concurrent creates depends on the target path:

| concurrent creates | result |
|---|---|
| Same branch, managed default path | One success. Each other request fails with `worktree_operation_in_progress`. A request that starts after the first create ends fails with `worktree_create_failed` instead. |
| Different branches, same explicit path | One success. The other request fails with `worktree_operation_in_progress`. A request that starts after the first create ends fails with `worktree_create_failed` instead. |
| Same branch, different paths | One success. The other request fails with `worktree_create_failed`, because Git uses the branch in the first worktree. |
| Different branches, managed paths | All requests succeed. |

A `worktree.list` during a create also succeeds. For herdr 0.9.1, this page recorded
`worktree_create_failed` for a concurrent create of the same branch.

A CLI command with an unknown flag exits with status 2. It writes `unknown option: …`
to stderr. Other codes can occur. See [errors.md](../errors.md).

**CLI**: `herdr worktree create [--workspace <ID> | --cwd <PATH>] [--branch <NAME>] [--base <REF>] [--path <PATH>] [--label <TEXT>] [--focus] [--no-focus] [--trust-repository]`

You cannot use `--workspace` and `--cwd` together. If you give both, the command exits with
status 2 and writes this usage line:

```
usage: herdr worktree create [--workspace ID | --cwd PATH] [--branch NAME] [--base REF] [--path PATH] [--label TEXT] [--focus] [--no-focus] [--trust-repository]
```

herdr accepts `--focus --no-focus` together. The workspace is then not focused.

**Example**

```json
{"id":"wc1","method":"worktree.create","params":{"cwd":"/home/user/project","branch":"api-docs"}}
{"id":"wc1","result":{"type":"worktree_created","workspace":{"workspace_id":"w2","number":2,"label":"api-docs","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"w2:t1","agent_status":"unknown","worktree":{"repo_key":"/home/user/project/.git","repo_name":"project","repo_root":"/home/user/project","checkout_path":"/home/user/.herdr/worktrees/project/api-docs","is_linked_worktree":true}},"tab":{"tab_id":"w2:t1","workspace_id":"w2","number":1,"label":"1","focused":false,"pane_count":1,"agent_status":"unknown"},"root_pane":{"pane_id":"w2:p1","terminal_id":"term_65e0dc041d1992","workspace_id":"w2","tab_id":"w2:t1","focused":false,"cwd":"/home/user/.herdr/worktrees/project/api-docs","foreground_cwd":"/home/user/.herdr/worktrees/project/api-docs","agent_status":"unknown","scroll":{"offset_from_bottom":0,"max_offset_from_bottom":0,"viewport_rows":40},"revision":0},"worktree":{"path":"/home/user/.herdr/worktrees/project/api-docs","branch":"api-docs","is_bare":false,"is_detached":false,"is_prunable":false,"is_linked_worktree":true,"open_workspace_id":"w2","label":"project"}}}
```

The root pane of a new workspace does not include `terminal_title` and
`terminal_title_stripped`. Both fields appear after the pane settles, as in the
[worktree.open](#worktreeopen) example below.

Checked 2026-10-17 against herdr 0.9.5: throwaway `git init` repositories on a scratch
server with its own home directory. herdr wrote to the default `~/.herdr/worktrees`
directory, with no config override. The pass made the repository untrusted with the
`GIT_TEST_ASSUME_DIFFERENT_OWNER=1` variable of Git in the server environment. It did not
use a repository that a different user owns. The pass sent concurrent creates again. The
example comes from this pass.

## worktree.list

Lists all worktrees of the Git repository that owns the resolved checkout. The result
also has a `source` block. That block identifies the repository and the checkout from
which the server resolved the listing. The server finds the repository from
`workspace_id` or `cwd`. When the request gives neither, the server uses the focused
workspace.

Unlike create and open, list accepts a linked worktree as the source. The server uses the
parent checkout of the repository and shows that checkout in `source`. The method is
read-only. Each `WorktreeInfo` in the result shows in `open_workspace_id` if a workspace
holds it open.

`trust_repository` gives Git trust for this one request. Use it only after a person checks
the repository. Do not use it as a routine retry for a failed worktree command.

**Params** (`WorktreeListParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `workspace_id` | string \| null | no | null | Existing workspace whose repository the server lists. It resolves the target repository. |
| `cwd` | string \| null | no | null | Absolute working directory that resolves the repository when `workspace_id` is absent. |
| `trust_repository` | boolean | no | `false` | Gives Git trust for this request only. It bypasses the ownership check of Git for the repository. |

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the type, required, and
default columns.

**Result** — `type: "worktree_list"`:

| field | type | meaning |
|---|---|---|
| `type` | string const `"worktree_list"` | Result discriminator. |
| `source` | WorktreeSourceInfo | The repository and the checkout from which the server resolved the listing. See [WorktreeSourceInfo](#worktreesourceinfo) below. |
| `worktrees` | array of WorktreeInfo | All worktrees of the repository. See [WorktreeInfo](#worktreeinfo) below. |

**Errors**:

- `not_git_worktree`: `cwd` is not inside a Git work tree. This includes a path that does
  not exist and a path with a literal `~` component, such as `/tmp/~`.
- `workspace_not_found`: `workspace_id` is unknown or malformed.
- `invalid_request`: the request has no selector while no workspace is active. Other
  causes are a `cwd` that is not absolute or is empty, a param of the wrong type, and no
  `params` member.
- `worktree_list_failed`: Git refused the request, for example because of dubious
  ownership of the repository.

Other codes can occur. See [errors.md](../errors.md).

**CLI**: `herdr worktree list [--workspace <ID> | --cwd <PATH>] [--trust-repository]`

You cannot use `--workspace` and `--cwd` together. If you give both, the command exits with
status 2 and writes this usage line:

```
usage: herdr worktree list [--workspace ID | --cwd PATH] [--trust-repository]
```

**Example**

```json
{"id":"wl1","method":"worktree.list","params":{"cwd":"/home/user/project"}}
{"id":"wl1","result":{"type":"worktree_list","source":{"repo_key":"/home/user/project/.git","repo_name":"project","repo_root":"/home/user/project","source_checkout_path":"/home/user/project"},"worktrees":[{"path":"/home/user/project","branch":"main","is_bare":false,"is_detached":false,"is_prunable":false,"is_linked_worktree":false,"label":"project"}]}}
```

This capture shows a repository that no workspace opened before. Thus `source` has no
`source_workspace_id`, and the entry has no `open_workspace_id`. When a workspace holds
the primary checkout open, `source_workspace_id` appears. This is also true for a listing
that the server resolved from `cwd`.

Checked 2026-10-17 against herdr 0.9.5: throwaway `git init` repositories on a scratch
server. The pass used a bare clone and worktrees that were detached, prunable, or open. It
also used a linked worktree as the source, a `cwd` with `~/`, and an untrusted repository
with and without `trust_repository`. The example comes from this pass.

## worktree.open

Opens an existing Git worktree as a herdr workspace. Like the `worktree.create` result,
the result gives the `WorkspaceInfo`, `TabInfo`, root `PaneInfo`, and `WorktreeInfo`. It
also gives an `already_open` flag. This flag is true when a workspace already held the
worktree open. In that case, the server gives the existing workspace and does not create
a new one.

The request selects the worktree with **exactly one** of `path` or `branch`. If the
request gives neither or both, it fails with `invalid_request`
(`exactly one of path or branch is required`). This also occurs when the two values
identify the same worktree. The schema types both fields as optional and nullable. Thus
only the server enforces the rule.

The server finds the owning repository from `workspace_id` or `cwd`, or else from the
focused workspace. As with `worktree.create`, that source must be the parent checkout of
the repository. Thus a request with only `path` works only while the parent workspace is
focused.

`label` sets the workspace label, and `focus` sets if the server focuses the workspace.
Both also apply when the worktree is already open. Thus an open call with `label` renames
the live workspace and sends a `workspace_renamed` event. An open call with `focus: true`
focuses the workspace again. The call is not read-only.

`trust_repository` gives Git trust for this one request. Use it only after a person checks
the repository. Do not use it as a routine retry for a failed worktree command.

**Params** (`WorktreeOpenParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `workspace_id` | string \| null | no | null | Existing workspace whose repository owns the worktree. It resolves the target repository. It must not be the workspace of a linked worktree. |
| `cwd` | string \| null | no | null | Absolute working directory that resolves the repository when `workspace_id` is absent. It must not be inside a linked worktree. |
| `path` | string \| null | one of | null | Absolute checkout path of the worktree to open. Give exactly one of `path` or `branch`. |
| `branch` | string \| null | one of | null | Branch name of the worktree to open. Give exactly one of `path` or `branch`. |
| `label` | string \| null | no | null | Workspace label. When it is null, the label is the basename of the checkout directory. It renames the workspace when the workspace is already open. |
| `focus` | boolean | no | `false` | When it is true, the server focuses the workspace after it opens the worktree. |
| `trust_repository` | boolean | no | `false` | Gives Git trust for this request only. It bypasses the ownership check of Git for the repository. |

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the type, required, and
default columns.

**Result** — `type: "worktree_opened"`:

| field | type | meaning |
|---|---|---|
| `type` | string const `"worktree_opened"` | Result discriminator. |
| `workspace` | WorkspaceInfo | The opened workspace, or the workspace that was already open. See [data-model.md](../data-model.md#workspaceinfo). |
| `tab` | TabInfo | The active tab of the workspace. For a workspace that was already open, this is the tab that is active, which is not always `:t1`. See [data-model.md](../data-model.md#tabinfo). |
| `root_pane` | PaneInfo | The root pane of the tab. See [data-model.md](../data-model.md#paneinfo). |
| `worktree` | WorktreeInfo | The opened worktree checkout. See [WorktreeInfo](#worktreeinfo) below. |
| `already_open` | boolean | True if a workspace held the worktree open before this call. |

**Errors**:

- `invalid_request`: the request gives neither or both of `path` and `branch`. A `path`
  that is not absolute also fails with this code (`worktree path must be absolute`).
- `worktree_not_found`: the message is `worktree path not found`, or
  `worktree branch not found`. The second message occurs when no worktree has that branch
  checked out. It also occurs when no branch of that name exists.
- `not_git_worktree`: `cwd` is not inside a Git work tree.
- `linked_worktree_source`: the source checkout is a linked worktree. This includes a
  source from the focused workspace.
- `workspace_not_found`: `workspace_id` is unknown.
- `worktree_list_failed`: Git refused the request. The open method finds the worktree
  through an internal listing. Thus Git failures occur under the list code, not under a
  code for open.

Other codes can occur. See [errors.md](../errors.md).

**CLI**: `herdr worktree open [--workspace <ID> | --cwd <PATH>] (--path <PATH> | --branch <NAME>) [--label <TEXT>] [--focus] [--no-focus] [--trust-repository]`

You cannot use `--workspace` and `--cwd` together. You must give exactly one of `--path`
or `--branch`. If you break one of these rules, the command exits with status 2 and writes
this usage line:

```
usage: herdr worktree open [--workspace ID | --cwd PATH] (--path PATH | --branch NAME) [--label TEXT] [--focus] [--no-focus] [--trust-repository]
```

**Example**

```json
{"id":"wo1","method":"worktree.open","params":{"cwd":"/home/user/project","path":"/home/user/.herdr/worktrees/project/api-docs"}}
{"id":"wo1","result":{"type":"worktree_opened","workspace":{"workspace_id":"w2","number":2,"label":"api-docs","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"w2:t1","agent_status":"unknown","worktree":{"repo_key":"/home/user/project/.git","repo_name":"project","repo_root":"/home/user/project","checkout_path":"/home/user/.herdr/worktrees/project/api-docs","is_linked_worktree":true}},"tab":{"tab_id":"w2:t1","workspace_id":"w2","number":1,"label":"1","focused":false,"pane_count":1,"agent_status":"unknown"},"root_pane":{"pane_id":"w2:p1","terminal_id":"term_65e0dc041d1992","workspace_id":"w2","tab_id":"w2:t1","focused":false,"cwd":"/home/user/.herdr/worktrees/project/api-docs","foreground_cwd":"/home/user/.herdr/worktrees/project/api-docs","terminal_title":"user@host:~/.herdr/worktrees/project/api-docs","terminal_title_stripped":"user@host:~/.herdr/worktrees/project/api-docs","agent_status":"unknown","scroll":{"offset_from_bottom":0,"max_offset_from_bottom":0,"viewport_rows":40},"revision":1},"worktree":{"path":"/home/user/.herdr/worktrees/project/api-docs","branch":"api-docs","is_bare":false,"is_detached":false,"is_prunable":false,"is_linked_worktree":true,"open_workspace_id":"w2","label":"project"},"already_open":true}}
```

`already_open` is `true` because `worktree.create` created the worktree and opened it as
`w2` in the same session just before. This request needs `cwd` unless the parent workspace
of the repository is focused. Right after `worktree.create`, the new linked worktree can be
focused. Then a request with only `path` fails with `linked_worktree_source`.

Checked 2026-10-17 against herdr 0.9.5: the pass opened worktrees by path and by branch.
It used worktrees that were already open, with `label` that renamed the live workspace and
a `tab` that showed an active tab other than `:t1`. It also used a worktree that
`git worktree add` made outside herdr, and a detached worktree. Other cases were a
request with only `path` while the parent or a linked workspace was focused, a branch with
no worktree, and an untrusted repository. The example comes from this pass.

## worktree.remove

Removes a worktree checkout. The request identifies the checkout with the `workspace_id`
of the workspace that holds it open. `workspace_id` is required. Set `force` to remove
the checkout when it has uncommitted or other dirty state that Git usually does not
discard.

`trust_repository` gives Git trust for this one request. Use it only after a person checks
the repository. Do not use it as a routine retry for a failed worktree command.

The result gives the `path` of the removed checkout and the `forced` flag. `forced` shows
the request flag, not if the remove needed force. Thus `force: true` on a clean worktree
also gives `forced: true`.

The remove also **closes the workspace that holds the worktree**. herdr sends
`workspace_closed`, and a later `workspace.get` gives `workspace_not_found`. herdr deletes
the checkout directory but **keeps the branch**. `git branch --list` shows the branch after
the remove.

On herdr 0.9.5, each observed remove sent `workspace_closed` and `worktree_removed`. The
order of the two frames followed the order of the `subscriptions` array in
`events.subscribe`. Thus do not rely on this order. The `worktree_removed` frame includes
the optional `workspace` key. For herdr 0.9.3, this page recorded `worktree_removed` before
`workspace_closed`. That pass may have used a different order of subscriptions.

**Params** (`WorktreeRemoveParams`):

| field | type | required | default | meaning |
|---|---|---|---|---|
| `workspace_id` | string | **yes** | — | Workspace whose worktree checkout the server removes. |
| `force` | boolean | no | `false` | Remove the checkout although it has dirty or uncommitted state. |
| `trust_repository` | boolean | no | `false` | Gives Git trust for this request only. It bypasses the ownership check of Git for the repository. |

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the type, required, and
default columns.

**Result** — `type: "worktree_removed"`:

| field | type | meaning |
|---|---|---|
| `type` | string const `"worktree_removed"` | Result discriminator. |
| `workspace_id` | string | The workspace whose worktree the server removed. |
| `path` | string | File system path of the removed checkout. |
| `forced` | boolean | True if the request asked for force. |

**Errors**:

- `dirty_worktree_requires_force`: the checkout has files with changes or untracked files, and
  `force` is not set.
- `not_linked_worktree`: two cases have different messages.
  `workspace is not a linked worktree checkout` is for the primary checkout of the
  repository. `workspace is not a Herdr-managed worktree checkout` is for a workspace
  with no worktree.
- `workspace_not_found`: the workspace is unknown. This includes a workspace whose
  worktree an earlier call removed.
- `worktree_operation_in_progress`: another create or remove for the same checkout is in
  progress. Two parallel removes of the same workspace gave one success and one failure
  with this code. A parallel open and remove of the same worktree both succeeded.
- `worktree_remove_failed`: Git refused the request, for example because of dubious
  ownership of the repository.
- `invalid_request`: `workspace_id` is missing or null. herdr cannot deserialize these
  requests. On herdr 0.9.5 and 0.9.3, the response echoes the request `id`. herdr 0.9.1
  gave an empty `id`.

A CLI command with an unknown flag exits with status 2. For example, `--path` writes
`unknown option: --path` to stderr. `remove` accepts only `--workspace`, `--force`, and
`--trust-repository`. Other codes can occur. See [errors.md](../errors.md).

**CLI**: `herdr worktree remove --workspace <ID> [--force] [--trust-repository]`

If you do not give `--workspace`, the command exits with status 2 and writes this usage
line:

```
usage: herdr worktree remove --workspace ID [--force] [--trust-repository]
```

**Example**

```json
{"id":"wr1","method":"worktree.remove","params":{"workspace_id":"w2","force":false}}
{"id":"wr1","result":{"type":"worktree_removed","workspace_id":"w2","path":"/home/user/.herdr/worktrees/project/api-docs","forced":false}}
```

Checked 2026-10-17 against herdr 0.9.5: throwaway `git init` repositories on a scratch
server. The targets were clean, dirty, primary-checkout, worktree-less, and untrusted. The
pass also sent concurrent removes. Each remove deleted the checkout and kept the branch.
The example comes from this pass.

## Worktree domain types

These types occur only in `worktree.*` results, and this section expands them.
[data-model.md](../data-model.md) has the shared entities (`WorkspaceInfo`, `TabInfo`,
`PaneInfo`).

herdr **omits** optional fields and does not send them as `null`. Thus a client must test
if a key is present. It must not compare the key to `null`.

### WorktreeInfo

Describes one Git worktree of a repository.

| field | type | required | meaning |
|---|---|---|---|
| `path` | string | yes | Checkout path of the worktree. |
| `is_bare` | boolean | yes | True if this is the entry of the bare repository. |
| `is_detached` | boolean | yes | True if the HEAD of the worktree is detached. |
| `is_prunable` | boolean | yes | True if Git thinks that the worktree is prunable (stale). |
| `is_linked_worktree` | boolean | yes | True if this is a linked worktree, not the primary checkout. |
| `label` | string | yes | Display label for the worktree. It is the repository name, the same on each entry of a listing. It is never the branch or the checkout directory. |
| `branch` | string | no | Checked-out branch. herdr omits it when the worktree is detached or is the bare entry. The schema types it `string \| null`, but herdr never sends a null. |
| `open_workspace_id` | string | no | ID of the workspace that holds this worktree open. herdr omits it when the worktree is not open. The schema types it `string \| null`, but herdr never sends a null. |

### WorktreeSourceInfo

Identifies the repository and the checkout from which the server resolved a
`worktree.list`. For a bare repository, `repo_key` and `repo_root` are both the bare
directory, with no `/.git` suffix. `repo_name` keeps the `.git` suffix, for example
`bare.git`.

| field | type | required | meaning |
|---|---|---|---|
| `repo_key` | string | yes | Stable key for the repository (its `.git` path). |
| `repo_name` | string | yes | Repository name. |
| `repo_root` | string | yes | File system root of the primary checkout of the repository. |
| `source_checkout_path` | string | yes | Checkout path from which the server resolved the listing. It is the parent checkout of the repository, also when the caller named a linked worktree. |
| `source_workspace_id` | string | no | The **parent** workspace of the repository, when one is open. This is true for a listing from `cwd`. It is also true when the caller gave the `workspace_id` of a linked worktree. herdr omits it when no such workspace is open. The schema types it `string \| null`, but herdr never sends a null. |

### WorkspaceWorktreeInfo

The worktree summary in the `WorkspaceInfo.worktree` field. It is on create and open
results. It is also on the `workspace.list` or `workspace.get` entry of each workspace
that a worktree backs. It is different from `WorktreeInfo`.

| field | type | required | meaning |
|---|---|---|---|
| `repo_key` | string | yes | Stable key for the repository (its `.git` path). |
| `repo_name` | string | yes | Repository name. |
| `repo_root` | string | yes | File system root of the primary checkout of the repository. |
| `checkout_path` | string | yes | Checkout path of the worktree of this workspace. |
| `is_linked_worktree` | boolean | yes | True if the checkout of the workspace is a linked worktree. |

Checked 2026-10-17 against herdr 0.9.5: the pass read all three types from
`worktree.list`, `worktree.create`, `worktree.open`, and `workspace.get` responses. These
included bare, detached, and prunable entries. No worktree response had a `null` value.
`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the type and required
columns.
