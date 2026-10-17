# herdr API: layout methods

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Index: [README.md](../README.md). Wire format: [protocol.md](../protocol.md).

The `layout.*` methods read and change the pane tree of a tab.

- `layout.export` writes the current layout of a tab as a `LayoutNode` tree.
- `layout.apply` builds a `LayoutNode` tree into a workspace or a tab. It makes the panes and splits, and it can start a command in each pane.
- `layout.set_split_ratio` changes the divider position of one split in the tab.

`layout.apply` accepts an exported tree and builds the same structure. But the tree is not fully portable. A pane's `env` has an effect on apply. No `layout.apply` or `layout.export` result includes `env`. Thus, when you export a layout and apply it again, every pane loses its environment with no error. Refer to [LayoutNode](#layoutnode).

All three methods are API-only. There is no `herdr layout` CLI group. `herdr layout` writes `unknown command: layout` and exits with status 2. The CLI command `herdr pane layout` is part of the `pane.*` namespace. It only shows layout information, and it does not call these methods.

Each method returns a `LayoutDescription` snapshot of the tab after the operation. One exception applies: when `layout.apply` gets a `tab_id`, the returned tab did not exist before the call. Refer to [layout.apply](#layoutapply).

Checked 2026-10-17 against herdr 0.9.5: the export-to-apply structure, the missing `env`, the CLI error and exit status, and the `pane layout` result type.

3 methods (`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`):

| method | purpose |
|---|---|
| [layout.apply](#layoutapply) | Build a `LayoutNode` tree into a workspace or a tab, and make its panes and splits. |
| [layout.export](#layoutexport) | Write the current layout of a tab, or of the tab that holds a pane, as a `LayoutNode` tree. |
| [layout.set_split_ratio](#layoutset_split_ratio) | Set the divider ratio of one split. A path of child selectors identifies the split. |

## Shared types

### LayoutNode

`LayoutNode` is a recursive union (`oneOf`) for one node of a layout tree. The `type` discriminator selects the variant. Requests use this type for the layout to apply. Results use it for the exported or current layout. The two directions do not have the same field set. `layout.apply` accepts `env`, and `env` has an effect. But no `layout.apply` or `layout.export` result includes `env`. Refer to the `env` row below. [../data-model.md](../data-model.md#layoutnode) also lists this type.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the variants, the field types, and the required flags in the two tables below. The row text gives the checked behavior.

**`type: "pane"`**: a leaf terminal pane.

| field | type | required | default | meaning |
|---|---|---|---|---|
| `type` | `"pane"` (const) | yes | — | Discriminator. |
| `pane_id` | string \| null | no | null | ID of an existing pane. On `layout.apply`, this value is only a hint and does not reserve the ID. Apply gives new pane IDs. In a probe, the requested `w1:p1` became `w1:p2`. On export, this value is the live pane ID. |
| `command` | array of string \| null | no | null | Command line (argv) that the pane runs on apply. Null or no value gives a default shell. Apply refuses an empty array (`[]`) with `invalid_layout` ("pane command must not be empty"). Unlike `env`, the result includes the pane's `command`. The `layout.apply` result and a later `layout.export` both include it. A default-shell pane has no `command` field. |
| `cwd` | string \| null | no | null | Working directory for the pane. A `cwd` that does not exist does not cause an error. herdr uses `$HOME` instead, with no error, and the result shows the `$HOME` path. For a `cwd` that exists, the result shows the canonical path, with symbolic links resolved. |
| `env` | object (string → string) | no | `{}` | Environment variables for the pane. They have an effect on `layout.apply`. No result includes `env`. Refer to the text above. |
| `label` | string \| null | no | null | Label for the pane. The apply and export results include it. |

**`type: "split"`**: an inner node that divides an area between two child nodes.

| field | type | required | default | meaning |
|---|---|---|---|---|
| `type` | `"split"` (const) | yes | — | Discriminator. |
| `direction` | `SplitDirection` enum: `right`, `down` | yes | — | Orientation of the split. `right` puts `second` to the right of `first`, with a vertical divider. `down` puts `second` below `first`, with a horizontal divider. |
| `ratio` | number (float) | yes | — | Part of the area that `first` gets. herdr clamps the value to `0.1`–`0.9` with no error. It accepts values out of range, such as `0.0`, `1.0`, negative values, and values more than `1`. The clamped value has the effect, and the result shows the clamped value. herdr stores the value as `f32`. It does not keep precision past approximately 7 significant digits. For example, `0.123456789` becomes `0.12345679`. |
| `first` | `LayoutNode` | yes | — | The first child subtree, at the left or the top. |
| `second` | `LayoutNode` | yes | — | The second child subtree, at the right or the bottom. |

Checked 2026-10-17 against herdr 0.9.5: the pane-ID hint, the empty `command` error, the `command` and `label` echo, and the `cwd` fallback and canonical path. Also checked: the missing `env`, the split directions, the ratio clamp, and the `f32` rounding.

### LayoutDescription

Every `layout.*` result holds this snapshot in its `layout` field. The schema makes all fields required.

| field | type | meaning |
|---|---|---|
| `workspace_id` | string | Workspace that holds the tab, for example `w1`. |
| `tab_id` | string | Tab that this layout describes, for example `w1:t1`. |
| `zoomed` | boolean | True when one pane fills the tab with zoom. While `zoomed` is `true`, `root` still describes the full split tree, as without zoom. No field in this snapshot identifies the zoomed pane. `focused_pane_id` does not identify it either. `pane.zoom` focuses the pane, but focus can then move to a different pane while the zoom stays. |
| `focused_pane_id` | string | ID of the pane that has focus in the tab now. |
| `root` | `LayoutNode` | Root of the pane tree of the tab. |

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the field list and the required flags. Every observed result had all five fields. Checked 2026-10-17 against herdr 0.9.5: the `zoomed` behavior and the `focused_pane_id` after a focus change.

## layout.apply

`layout.apply` builds a `LayoutNode` tree into a target. It makes the panes and splits of the tree. It can also start the `command` of each pane.

- Apply does not wait for a pane command to exit. If a pane command exits immediately, herdr closes that pane. If that pane was the last pane in the tab, herdr also closes the tab.
- The tree can have a maximum depth of 16 levels and a maximum of 24 panes. A larger tree fails with `invalid_layout`. The messages are `layout depth is 17; maximum is 16` and `layout has more than 24 panes`.
- The depth count includes all nodes, leaves too. Apply accepts a chain of 15 nested splits (depth 16). Apply refuses a chain of 16 nested splits.

Targets:

- With no `workspace_id` and no `tab_id`, apply makes a new tab in the focused workspace. On a server with no workspaces, this form fails with `workspace_not_found` (`no active workspace`).
- With `workspace_id` and no `tab_id`, apply makes a **new tab** in that workspace for the layout.
- Apply refuses a request with both `workspace_id` and `tab_id` (`invalid_target`).
- `tab_id` does **not** change the named tab in place. Apply closes that tab and builds a replacement tab with a new `tab_id`. The replacement goes at the end of the tab order of the workspace.
- The replacement gets the label of the closed tab, unless `tab_label` sets a different label. All pane IDs from the closed tab become invalid.
- Pane IDs in the input `root` are only hints. Apply always gives new pane IDs.

`focus` controls if focus moves to the new layout. The default is `false`. This default agrees with the herdr rule to keep the focus of the caller (refer to [raw/skill.md](../raw/skill.md)). The `focus` row below gives three special cases.

Apply returns the layout of the result tab. For a `tab_id` call, this is the layout of the new replacement tab.

**Params**: `LayoutApplyParams`:

| field | type | required | default | meaning |
|---|---|---|---|---|
| `root` | `LayoutNode` | yes | — | Layout tree to build. The maximum is 16 levels of depth and 24 panes. A deeper or larger tree fails with `invalid_layout`. Refer to [LayoutNode](#layoutnode). |
| `workspace_id` | string \| null | no | null | Workspace for the layout. With no `tab_id`, apply makes a new tab here. With no `workspace_id` and no `tab_id`, apply makes a new tab in the focused workspace. A request with both `workspace_id` and `tab_id` fails with `invalid_target`. |
| `tab_id` | string \| null | no | null | Existing tab to use as the target. Apply does not change this tab in place. Apply closes it and makes a replacement with a different `tab_id` at the end of the tab order. Refer to the text above. |
| `tab_label` | string \| null | no | null | Label for the target tab or the new tab. With `tab_id` and no `tab_label`, the replacement gets the label of the closed tab. |
| `focus` | boolean | no | `false` | True moves focus to the new layout. Usually, `false` keeps focus where it is. Three cases are different. (1) Apply into the only tab of a workspace focuses the replacement, because the previous active tab does not exist now. (2) Apply into the active tab of a workspace without focus makes a sibling tab the active tab of that workspace, not the new tab. (3) Apply into the active tab of the focused workspace makes the replacement the active tab. |

**Result**: `type: "layout_apply"`:

| field | type | meaning |
|---|---|---|
| `type` | `"layout_apply"` (const) | Result discriminator. |
| `layout` | `LayoutDescription` | Snapshot of the tab after apply. Refer to [LayoutDescription](#layoutdescription). |

**Errors** (codes in [../errors.md](../errors.md)):

| code | when |
|---|---|
| `workspace_not_found` | The `workspace_id` is unknown. Or the request gives no target, and the server has no workspace. |
| `tab_not_found` | The `tab_id` is unknown. |
| `invalid_target` | The request gives both `workspace_id` and `tab_id`. |
| `invalid_layout` | The tree has a problem that the schema cannot show. The problems are an empty pane `command` array, a depth of more than 16, or more than 24 panes. |
| `invalid_request` | The `LayoutNode` has an incorrect shape: an unknown `type`, or a `split` without `direction`, `ratio`, `first`, or `second`. |

Other codes are possible.

**Events** (catalog in [../events.md](../events.md)):

- Apply into a new tab sends `tab_created` (subscription type `tab.created`), one `pane_created` for each pane, and `layout_updated`. Subscription types are `pane.created` and `layout.updated`.
- Each pane that starts also sends one `pane_updated` (subscription type `pane.updated`) when its shell becomes stable. A pane whose command exits immediately sends `pane_exited` and no `pane_updated`.
- On one connection, the order of these events follows the order of the subscriptions in the `events.subscribe` request. The server seems to group the events by subscription. The example that follows shows one order only.
- Example: for a three-pane apply, the subscription order `tab.created`, `pane.created`, `pane.updated`, `layout.updated` gave `tab_created`, three `pane_created`, three `pane_updated`, and then `layout_updated`.
- Apply with an existing `tab_id` also sends `tab_closed` (subscription type `tab.closed`) for the closed tab. It sends no `pane_closed` for the panes of that tab.
- The `layout_updated` payload is a `PaneLayoutSnapshot` (area, panes, and splits, with rects). It is not the `LayoutDescription` that this method returns.

**CLI**: API-only (no CLI subcommand).

Checked 2026-10-17 against herdr 0.9.5: a pass ran every claim in this section again on a scratch server. These claims include:

- the targets, the order and label of the replacement tab, the `tab_label` change, and the pane-ID hints
- the 16 and 24 limits, every error row, and the no-workspace error
- the `focus` default and the three special cases
- the `cwd` fallback to `$HOME`, the effect of `env` without an echo, and the close of a pane and a tab when a command exits
- the ratio clamp, the `f32` rounding, the ignored unknown params, and the events above

**Example**: Checked 2026-10-17 against herdr 0.9.5. A pass captured this example as the first apply on a scratch server. The server had one workspace with one pane `w1:p1`, so the result is `w1:t2` and `w1:p2`. The IDs depend on the earlier session state.

```json
{"id":"l3","method":"layout.apply","params":{"workspace_id":"w1","root":{"type":"pane","pane_id":"w1:p1","cwd":"/home/user/project"}}}
{"id":"l3","result":{"type":"layout_apply","layout":{"workspace_id":"w1","tab_id":"w1:t2","zoomed":false,"focused_pane_id":"w1:p2","root":{"type":"pane","pane_id":"w1:p2","cwd":"/home/user/project"}}}}
```

## layout.export

`layout.export` writes the current layout of a tab as a `LayoutNode` tree. `layout.apply` accepts this tree and builds the same structure. But the tree is not fully portable, because it has no `env`. Refer to [LayoutNode](#layoutnode).

- Use `tab_id` to identify the tab. Or use `pane_id` to export the tab that holds that pane.
- Do not use the two fields together. A request with both fails with `layout_not_found`, also when the pane is in that tab.
- With no `tab_id` and no `pane_id`, export uses the active tab of the focused workspace.
- Export only reads. It does not change state.

Export returns a `LayoutDescription` for the exported tab.

**Params**: `LayoutExportParams`:

| field | type | required | default | meaning |
|---|---|---|---|---|
| `tab_id` | string \| null | no | null | Tab to export. |
| `pane_id` | string \| null | no | null | Export the tab that holds this pane. |

> Schema discrepancy: the captured probe requests set `workspace_id`, for example `"w1"`. `LayoutExportParams` does **not** declare this field. The schema declares only `tab_id` and `pane_id`. Use the schema as the reference. Thus, `workspace_id` here is an extra field that herdr ignores without documentation. A request with `"workspace_id":"w99"` also exports the active tab with no error. To select a tab, use `tab_id`. herdr also ignores unknown params to `layout.apply` and `layout.set_split_ratio`. A field name with a typing error has no effect and causes no error.

**Result**: `type: "layout_export"`:

| field | type | meaning |
|---|---|---|
| `type` | `"layout_export"` (const) | Result discriminator. |
| `layout` | `LayoutDescription` | The exported tab layout. Refer to [LayoutDescription](#layoutdescription). |

**Errors** (codes in [../errors.md](../errors.md)):

| code | when |
|---|---|
| `layout_not_found` | herdr cannot find the target. The cause can be an unknown `tab_id`, an unknown `pane_id`, both fields together, or no tab to export. A server with no workspaces is an example of the last cause. This one code covers all these causes and does not identify the cause. |
| `invalid_request` | A field has an incorrect type, for example a `tab_id` that is a number. |

Other codes are possible.

**Events**: none. A subscription to every `layout.*`, `tab.*`, `pane.*`, and `workspace.*` event stayed open during an export call. It received no event.

**CLI**: API-only (no CLI subcommand).

Checked 2026-10-17 against herdr 0.9.5: a pass ran these claims again. They include the targets, every error row, the ignored `workspace_id`, the zoomed snapshot, and the no-events claim. The no-workspaces case ran on a new scratch server.

**Example**: Checked 2026-10-17 against herdr 0.9.5. A pass captured this example on a new scratch server before any apply. The request uses the undocumented `workspace_id` field from the note above. Use `tab_id` instead.

```json
{"id":"r1","method":"layout.export","params":{"workspace_id":"w1"}}
{"id":"r1","result":{"type":"layout_export","layout":{"workspace_id":"w1","tab_id":"w1:t1","zoomed":false,"focused_pane_id":"w1:p1","root":{"type":"pane","pane_id":"w1:p1","cwd":"/home/user/project"}}}}
```

## layout.set_split_ratio

`layout.set_split_ratio` sets the divider `ratio` of one split node in the layout of a tab. A `path` of boolean child selectors identifies the split.

- herdr follows the path from the root of the tab. The empty path selects the top-level split of the tab.
- At each split, one element selects a branch. `false` selects `first`, and `true` selects `second`.
- herdr changes the ratio of the split at the end of the path.
- Use `tab_id` to identify the tab. Or use `pane_id` to identify the tab of that pane.
- Do not use the two fields together. A request with both fails with `layout_not_found`, as with `layout.export`.
- With no `tab_id` and no `pane_id`, herdr uses the active tab.

The method returns the layout of the tab after the change.

If the `path` does not go to an existing split, the method fails with `split_not_found`. A tab with one pane and no split is an example.

**Params**: `LayoutSetSplitRatioParams`:

| field | type | required | default | meaning |
|---|---|---|---|---|
| `path` | array of boolean | yes | — | Child selectors from the root to the target split: `false` selects `first`, and `true` selects `second`. An empty array selects the top-level split of the tab. |
| `ratio` | number (float) | yes | — | New part of the area that `first` gets. herdr clamps the value to `0.1`–`0.9` with no error. Refer to [LayoutNode](#layoutnode) `split.ratio` for the clamp behavior and the `f32` precision. herdr accepts integers (`1` → `0.9`, `0` → `0.1`). |
| `tab_id` | string \| null | no | null | Tab that has the split. |
| `pane_id` | string \| null | no | null | Change the split in the tab that holds this pane. |

**Result**: `type: "layout_split_ratio_set"`:

| field | type | meaning |
|---|---|---|
| `type` | `"layout_split_ratio_set"` (const) | Result discriminator. |
| `layout` | `LayoutDescription` | Snapshot of the tab after the split change. Refer to [LayoutDescription](#layoutdescription). |

**Errors** (codes in [../errors.md](../errors.md)):

| code | when |
|---|---|
| `split_not_found` | The `path` does not go to an existing split in the target tab. In a probe, `path:[false]` on a single-pane tab caused this error. |
| `layout_not_found` | The `tab_id` or the `pane_id` is unknown, or the request gives both. On a server with no workspaces, a request with no target also fails with this code, not with `split_not_found`. |
| `invalid_request` | The `path` has an incorrect shape, for example an element that is not a boolean. Or `ratio` has an incorrect type. Or the request has no `path`. |

Other codes are possible.

**Events** (catalog in [../events.md](../events.md)): the method sends one `layout_updated` event (subscription type `layout.updated`). The payload is the same `PaneLayoutSnapshot` that [layout.apply](#layoutapply) describes. The method sends no `pane_updated`, although the rects of the panes change. Checked 2026-10-17 against herdr 0.9.5 on a scratch server with no TUI client attached.

The behavior with an attached client was not checked. herdr 0.9.3 also sent one `pane_updated` (subscription type `pane.updated`) for each pane that the new ratio resized. The 0.9.1 page gave only the `layout_updated` event.

**CLI**: API-only (no CLI subcommand).

With no `tab_id` and no `pane_id`, herdr uses the active tab. If the active tab of the server has only one pane, `{"path":[],"ratio":0.5}` fails with `split_not_found`.

Checked 2026-10-17 against herdr 0.9.5: a pass ran these claims again on a scratch server. They include the path walk, the two integer ratios, every error row, the ignored unknown param, and the events above.

**Example**: Checked 2026-10-17 against herdr 0.9.5. The capture is the same, byte for byte, as the 0.9.3 capture. This request targets a tab with one pane. The split path does not go to a split, so the server returns an error. A successful call follows.

```json
{"id":"l1","method":"layout.set_split_ratio","params":{"tab_id":"w1:t1","path":[false],"ratio":0.6}}
{"id":"l1","error":{"code":"split_not_found","message":"split path not found"}}
```

A successful call: Checked 2026-10-17 against herdr 0.9.5. The target tab has the root `split(right){A, split(down){B, C}}`. A `layout.apply` with `tab_label` `build` made this tab on the server of the apply example. The path `[]` selects the root split. This call sent only `layout_updated`.

```json
{"id":"l2","method":"layout.set_split_ratio","params":{"tab_id":"w1:t3","path":[],"ratio":0.8}}
{"id":"l2","result":{"type":"layout_split_ratio_set","layout":{"workspace_id":"w1","tab_id":"w1:t3","zoomed":false,"focused_pane_id":"w1:p3","root":{"type":"split","direction":"right","ratio":0.8,"first":{"type":"pane","pane_id":"w1:p3","label":"A","cwd":"/home/user/project"},"second":{"type":"split","direction":"down","ratio":0.25,"first":{"type":"pane","pane_id":"w1:p4","label":"B","cwd":"/home/user/project"},"second":{"type":"pane","pane_id":"w1:p5","label":"C","cwd":"/home/user/project"}}}}}}
```
