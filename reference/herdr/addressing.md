# herdr API: addressing and IDs

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Index: [README.md](README.md). Wire format: [protocol.md](protocol.md). Access model: [environment.md](environment.md).

herdr gives each workspace, tab, and pane an opaque public ID. herdr also accepts live agent names as targets. A name follows the current occupant of a pane. This page owns the ID scheme. It gives these rules:

- the grammar of an ID
- how stable and opaque an ID is
- why herdr does not use a closed ID again
- how a moved pane gets a new ID
- the difference between a stable ID and the display `number`, which can change
- how `--current` and the injected caller context find a target
- the rules for an agent name as a target

Each claim comes from `raw/skill.md` (section "Use IDs and caller context") and from captures of scratch servers. The `api/*.md` pages give the parameters and results of each method.

## ID scheme

A public ID is an opaque and stable string handle. The ID forms are:

| entity | form | example | notes |
| --- | --- | --- | --- |
| workspace | `w<C>` | `w1`, `wA`, `wZ` | The top-level container. |
| tab | `w<C>:t<C>` | `w1:t1`, `w1:tA` | The ID includes the workspace. A tab belongs to exactly one workspace, and its ID names that workspace. |
| pane | `w<C>:p<C>` | `w1:p1`, `w1:pA` | The ID includes the workspace. The ID of a pane names the workspace that holds the pane now. |

`<C>` is a counter for each workspace or session. It is **not** a decimal integer. herdr writes it in an alphabet that is similar to Crockford Base32:

1. `1` to `9`.
2. `A,B,C,D,E,F,G,H,J,K,M,N,P,Q,R,S,T,V,W,X,Y,Z`. herdr does not use `I`, `L`, `O`, or `U`, because they look like other characters.
3. `0`.
4. Two-character values: `11`, `12`, … `1Z`, `10`, `21`, ….

IDs such as `w1:pA`, `wZ`, `w0`, and `w11` are usual IDs. A caller can parse or check IDs with a pattern for decimal digits only, such as `w\d+` or `w\d+:p\d+`. That caller then refuses valid IDs.

`raw/schema.json` puts no `pattern` on `workspace_id`, `tab_id`, or `pane_id`. Thus the schema does not give this alphabet. The alphabet is a runtime fact only. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`.

The workspace, tab, and pane counters are independent. Thus `w2:p1` and `w1:p1` are two different panes that have the same suffix `1`. The tab ID and the pane ID include the workspace. Thus a pane gets a **different** ID when it moves to a different workspace (refer to [Pane-move re-identification](#pane-move-re-identification)). Each entity object holds its IDs in the fields `workspace_id`, `tab_id`, and `pane_id`. Cross-references also use these fields, such as the `tab_id` and `workspace_id` of a pane.

Checked 2026-10-17 against herdr 0.9.5. A scratch server made these IDs in sequence:

```text
WS:  w2 w3 w4 w5 w6 w7 w8 w9 wA wB wC wD wE wF wG wH wJ wK wM wN wP wQ wR wS wT wV wW wX wY wZ w0 w11 w12 … w19 w1A w1B w1C w1D
WS2: w1E w1F w1G w1H w1J w1K w1M w1N w1P w1Q w1R w1S w1T w1V w1W w1X w1Y w1Z w10 w21 w22 w23 w24 w25
TABS: w2:t2 … w2:t9 w2:tA w2:tB w2:tC w2:tD
PANES: w1:p2 … w1:p9 w1:pA … w1:pH w1:pJ w1:pK w1:pM w1:pN
```

The sequence includes `w9` → `wA` → `wB`, `w2:t9` → `w2:tA` → `w2:tB`, and `w1:pH` → `w1:pJ` → `w1:pK`. No ID used `I`, `L`, `O`, or `U`. `w0` came after `wZ`, and `w10` came after `w1Z`, then `w21`. The counters were independent. The pane counter of `w1` came to `w1:pN`, but `w1` had only the tab `w1:t1`. Each workspace started its own `t1` and `p1`.

## Opacity and stability

Use an ID as an opaque handle. Do not calculate with an ID as structured data. The rules are:

- **Do not derive an ID**. Read the ID from a JSON response, such as a create result, a split result, a list, `pane current`, or a snapshot. skill.md: "Parse IDs from JSON responses. Do not derive them from sidebar order or examples."
- **A workspace ID stays the same when the workspaces change order**. A move of a workspace changes only its display `number`. The `workspace_id` does not change. Before this move, the order was `w1` (`demo-a`, number 1), `w2` (`demo-b`, number 2), and `w3` (`demo-c`, number 3). `workspace.move` put `w3` at index 0. After the move, `w3` has `number: 1` and `w1` has `number: 2`. The `workspace_id` values did not change.

  ```json
  {"id":"wm1","method":"workspace.move","params":{"workspace_id":"w3","insert_index":0}}
  {"id":"wm1","result":{"type":"workspace_list","workspaces":[
    {"workspace_id":"w3","number":1,"label":"demo-c",…},
    {"workspace_id":"w1","number":2,"label":"demo-a",…},
    {"workspace_id":"w2","number":3,"label":"demo-b",…}]}}
  ```

  Checked 2026-10-17 against herdr 0.9.5.
- **A tab ID stays the same when the tabs of its workspace change order**. `tab.move` changed the order of the tabs. Each tab kept its `tab_id` (`w1:t1`, `w1:t2`, `w1:t3`). But the `number` of a tab does **not** follow the new order, as the `number` of a workspace does. Refer to [Display numbers vs stable IDs](#display-numbers-vs-stable-ids-public-pane-numbers) for the behavior and an example. Checked 2026-10-17 against herdr 0.9.5.
- **An ID is valid on one server only**. This rule is new in herdr 0.9.1. With saved SSH machines, two machines can each have a `w1:p1` or an agent with the name `reviewer`. Without the global `--machine` prefix, a command continues to use the inherited session and socket context. Inherited local IDs and `--current` never identify remote panes. Discover IDs on the target machine. `herdr --help` lists `--machine <label-or-id>` as "Run an API command on a saved SSH machine". Two scratch servers on one machine each made their own `w1`, `w1:t1`, and `w1:p1`. `Documented by herdr --skill (herdr 0.9.5); not live-checked: needs a second saved SSH machine`.
- **A label is not an ID**. `label` is display text that a rename sets, and two items can have the same label. Never use a label as a target. A rename of two different workspaces to the same label succeeds for each, with no error. `workspace.get`, `tab.get`, and `pane.get` do not accept a label in the ID field. They return a not-found error. No method accepts a label as a selector.

  ```json
  {"id":"r1","method":"workspace.rename","params":{"workspace_id":"w1","label":"api-docs"}}  → result workspace_info
  {"id":"r2","method":"workspace.rename","params":{"workspace_id":"w2","label":"api-docs"}}  → result workspace_info
  {"id":"g1","error":{"code":"workspace_not_found","message":"workspace api-docs not found"}}
  {"id":"g2","error":{"code":"tab_not_found","message":"tab tab-B not found"}}
  {"id":"g3","error":{"code":"pane_not_found","message":"pane api-docs not found"}}
  ```

  Checked 2026-10-17 against herdr 0.9.5.

## Display numbers vs stable IDs (public pane numbers)

There are two different "numbers". Do not mix them:

- The **counter in a pane ID** is the stable public number of the pane, such as `p<C>` in `w1:p3` or `w1:pA`. herdr gives it one time, in the handle. While the pane stays in that workspace, the counter does not change. No other pane gets it again (refer to [Non-reuse of closed IDs](#non-reuse-of-closed-ids)). Use this ID with the `pane` methods. `PaneInfo` has no `number` field. The suffix of `pane_id` is the public number of the pane.
- The **`number` field** is a 1-based **display position**. `WorkspaceInfo` and `TabInfo` do not change in the same way when the order changes:
  - `WorkspaceInfo.number` changes when the workspaces change order (refer to the `workspace.move` example above).
  - `TabInfo.number` does **not** change. It keeps the position from the creation order of the tab for all the life of the tab. A `tab.move` that changes the position of the tab in the tab bar does not change it.

  Example: in a new workspace, a move of the third tab to the front keeps that tab at `number: 3`. The first tab and the second tab keep `number: 1` and `number: 2`. Only the order of the entries in the `tab.move` or `tab.list` response gives the new position.

  The same move also changed the label of the first tab, with no message. Nothing gave this tab a label, so it had the automatic label `"1"`. After the move, its `number` stayed `1`, but its label changed to `"2"`. The new label gives its new *display* position. Thus the `label` and the `number` of this tab give two different positions. The labels `tab-B` and `tab-C` came from `tab.rename`.

  ```json
  {"id":"tm1","method":"tab.move","params":{"tab_id":"w1:t3","insert_index":0}}
  {"id":"tm1","result":{"type":"tab_list","tabs":[
    {"tab_id":"w1:t3","workspace_id":"w1","number":3,"label":"tab-C",…},
    {"tab_id":"w1:t1","workspace_id":"w1","number":1,"label":"2",…},
    {"tab_id":"w1:t2","workspace_id":"w1","number":2,"label":"tab-B",…}]}}
  ```

  A `tab.list` after the move returned the same order and values. A different capture closed the tab `w4:t2`. The next `tab.create` then returned `w4:t3` with `number: 3` and `label: "2"`.

  In application logic, do not use the `number` as the target of a workspace or a tab. Use `workspace_id` or `tab_id` as the target.
- **A numeric selector works, but herdr does not document it**. `workspace.get`, `workspace.rename`, `tab.get`, and `tab.rename` accept the display `number` in the ID field as a different selector. They give no message about it. `pane.get` and `pane.rename` do not accept it.
  - `workspace.get({"workspace_id":"2"})` finds the workspace that has `number: 2` now.
  - The selector accepts a leading `+` and leading zeros (`"+1"`, `"01"`).
  - The selector does not accept whitespace (`" 1"`) or `"1.0"`.
  - `tab.get` and `tab.rename` accept the form `"<workspace_id>:<number>"`, such as `"w1:1"`. This form includes the workspace.
  - herdr finds the tab from its *current* display position, not from its `t<C>` suffix. Thus this selector is the target that changes, which the item above tells you not to use.
  - The selector is not read-only. `workspace.rename({"workspace_id":"3", "label":"x"})` and `tab.rename({"tab_id":"w1:1", "label":"x"})` change the entity that the selector finds at that time.
  - `pane.get` and `pane.rename` have no numeric selector. A bare number, `"p1"`, or a numeric form with a workspace all return `pane_not_found`.
  - herdr compares IDs of all three entity types with case sensitivity. `"W1:P1"` and `"wq:p1"` do not find the real `w1:p1`.

  In this capture, the workspace order was `w3` (number 1), `w1` (number 2), and `w2` (number 3). The tab order was `w1:t3`, `w1:t1`, and `w1:t2`.

  ```text
  workspace.get "1"   → w3 (number 1)      workspace.get "2"   → w1 (number 2)
  workspace.get "+1"  → w3                 workspace.get "01"  → w3
  workspace.get " 1"  → workspace_not_found   "1.0" → workspace_not_found
  workspace.get "W1"  → workspace_not_found   "wq"  → workspace_not_found   "w" → workspace_not_found
  tab.get "w1:1" → w1:t3 (display position 1, number 3)
  tab.get "w1:2" → w1:t1    tab.get "w1:3" → w1:t2    "w1:+1" → w1:t3    "w1:01" → w1:t3
  tab.get "W1:T1" → tab_not_found    "w1:T1" → tab_not_found
  pane.get "1" / "p1" / "w1:1" / "W1:P1" / "wq:p1" / "w1:P1" → pane_not_found
  workspace.rename {"workspace_id":"3","label":"x"} → result: w2 (number 3) label "x"   (changed)
  tab.rename {"tab_id":"w1:1","label":"x"}          → result: w1:t3 label "x"        (changed)
  pane.rename {"pane_id":"1","label":"x"}           → pane_not_found
  ```

Summary:

- The counter in a pane ID is stable.
- For workspaces, the `number` field is a display position that can change.
- For tabs, the `number` field is a fixed value from the creation order.
- The `number` works as a different selector for workspace and tab targets, but herdr does not document it. It never works for panes.

Checked 2026-10-17 against herdr 0.9.5. The check included the `tab.move` result, the label change to `"2"`, each selector form in the list, and the two changes through a numeric selector.

## Non-reuse of closed IDs

herdr does **not** use closed tab IDs and closed pane IDs again. skill.md:

> "Closed tab and pane IDs are not reused."

After a close of `w1:p2`, the next new pane in `w1` gets the next free number (`w1:p3`, …). It never gets `w1:p2` again. The results for a client are:

- A reference to a closed pane does not find a different, new pane. It finds nothing. A request with a closed ID as its target returns a not-found error. It does not act on a different pane.
- You can keep an ID in a cache for the life of the entity. A failed lookup tells you that the entity closed. A number that herdr uses again can never give you a false match.

Workspace IDs work in the same way in one test, but the skill text does not state it. The close of `w25`, the workspace with the highest ID, returned `ok`. The next `workspace.create` returned `w26`, not `w25`.

```json
{"id":"s1","method":"pane.split","params":{"target_pane_id":"w4:p1","direction":"right"}} → pane_id "w4:p2"
{"id":"c1","method":"pane.close","params":{"pane_id":"w4:p2"}} → {"type":"ok"}
{"id":"s2","method":"pane.split","params":{"target_pane_id":"w4:p1","direction":"right"}} → pane_id "w4:p3"
{"id":"g1","error":{"code":"pane_not_found","message":"pane w4:p2 not found"}}
tab.close w4:t2 → ok; tab.create → "w4:t3"; tab.get w4:t2 → {"code":"tab_not_found","message":"tab w4:t2 not found"}
```

Checked 2026-10-17 against herdr 0.9.5. The close of `w4:p2` and a new split made `w4:p3`. After that, `w4:p2` returned `pane_not_found`. For error codes, refer to [errors.md](errors.md).

## Pane-move re-identification

A move of a pane into a different tab or workspace gives the pane a **new identity**. The pane ID includes its container. Thus the pane gets a new ID that includes the workspace, and herdr retires its old ID. skill.md: "A pane moved into another workspace receives a new workspace-qualified pane ID."

`pane.move` returns a `pane_move` result. Its `move_result` (`PaneMoveResult`) gives the new identity and the previous identity:

| field | meaning |
| --- | --- |
| `move_result.pane` | The new `PaneInfo` of the moved pane, with its **new** `pane_id`. Continue with `move_result.pane.pane_id`. |
| `move_result.previous_pane_id` | The ID of the pane **before** the move. herdr retires this ID for general use. |
| `move_result.previous_tab_id`, `previous_workspace_id` | The IDs of the previous containers. |
| `move_result.created_tab` / `created_workspace` | herdr sets these when the move makes a destination tab or workspace (`--new-tab` / `--new-workspace`). |
| `move_result.closed_tab_id` / `closed_workspace_id` | herdr sets these when the move empties the source tab or workspace and closes it. |
| `move_result.focused_pane_id`, `move_result.changed`, `move_result.reason`, `source_layout`, `target_layout` | The focus result, a flag that tells if something changed, an optional reason, and the layouts before and after the move. The `PaneMoveReason` enum has `same_tab` and `zoomed_tab`. A probe on herdr 0.9.5 got each value. |

These are the moves that returned a `reason`, with `changed: false`:

```text
pane.move w8:p2 → {"type":"tab","tab_id":"w8:t1","split":"down"} (its own tab)
  → changed:false, reason:"same_tab"
pane.move wD:p2 → tab wC:t1 while wC:t1 was zoomed (2 panes, pane.zoom wC:p1 zoomed:true)
  → changed:false, reason:"zoomed_tab"
pane.move wC:p2 (in the zoomed tab) → {"type":"new_tab","workspace_id":"wC"}
  → changed:false, reason:"zoomed_tab"
```

herdr refuses a move into a zoomed tab with `zoomed_tab`. herdr also refuses a move *out of* a zoomed source tab with `zoomed_tab`.

The schema declares `closed_tab_id`, `closed_workspace_id`, `reason`, `created_tab`, and `created_workspace` as nullable, such as `closed_tab_id: ["string", "null"]`. But on the wire, herdr does not send a field that does not apply. herdr does not send it as JSON `null`. The schema gives `source_layout` as `anyOf [PaneLayoutSnapshot, null]`, and the field is not required. herdr also does not send `source_layout` when the move closes the source tab. A caller that checks `"closed_tab_id" in result` and a caller that checks `result.closed_tab_id is None` get different answers. A probe on herdr 0.9.5 got these key sets:

```text
cross-workspace move:   [changed, focused_pane_id, pane, previous_pane_id, previous_tab_id, previous_workspace_id, source_layout, target_layout]
same_tab:               [… , reason, source_layout, target_layout]
source tab emptied:     [changed, closed_tab_id, focused_pane_id, pane, previous_*, target_layout]
source workspace empty: [changed, closed_tab_id, closed_workspace_id, focused_pane_id, pane, previous_*, target_layout]
new_workspace:          [changed, created_tab, created_workspace, focused_pane_id, pane, previous_*, source_layout, target_layout]
```

The skill.md rule is:

```text
After `pane move`, continue with `.result.move_result.pane.pane_id` or the live agent name. The old value is reported as `.result.move_result.previous_pane_id`; only the moved process's inherited caller context keeps resolving that old ID, so do not use it as a general agent target.
```

**For the `pane.*` methods, measured behavior on herdr 0.9.1, 0.9.3, and 0.9.5 does not agree with this rule**. A retired `pane_id` continues to work for *each* caller, not only for the moved process. The caller needs no inherited context:

- `pane.get` succeeds with a retired ID.
- `pane.current` succeeds with a retired ID as `caller_pane_id`.
- `pane.rename` succeeds with a retired ID, and it changes the pane.

Each of these calls acts on the current identity of the pane. This continues across many moves. After two moves of a pane, the first ID and the middle ID both find its latest identity. When the pane closes, each ID in its history returns `pane_not_found`, retired or current.

Agent targets follow the skill.md rule. `agent.get` does not accept a retired pane ID. After a move of the pane of the agent `reviewer` from `wF:p2` to `wG:p2`:

```json
{"id":"g8","method":"agent.get","params":{"target":"wF:p2"}}
{"id":"g8","error":{"code":"agent_not_found","message":"agent target wF:p2 not found"}}
{"id":"g9","method":"agent.get","params":{"target":"wG:p2"}}   → agent_info name "reviewer", pane_id "wG:p2"
{"id":"g10","method":"agent.get","params":{"target":"reviewer"}} → agent_info pane_id "wG:p2"
```

Use `move_result.pane.pane_id` for each ID that you keep. The retired ID works for `pane.*` methods against the documented contract. This behavior can change with no notice.

Example: a different connection finds a pane by an ID that two moves retired. The first move took `w5:p2` to `w6:p2`:

```json
{"id":"pm1","method":"pane.move","params":{"pane_id":"w5:p2","destination":{"type":"tab","tab_id":"w6:t1","split":"right"}}}
{"id":"pm1","result":{"type":"pane_move","move_result":{"changed":true,
  "previous_pane_id":"w5:p2","previous_workspace_id":"w5","previous_tab_id":"w5:t1",
  "pane":{"pane_id":"w6:p2","terminal_id":"term_…","workspace_id":"w6","tab_id":"w6:t1",…},
  "source_layout":{…},"target_layout":{…},"focused_pane_id":"w6:p1"}}}
```

A second move took `w6:p2` to `w7:p2`. Then:

```json
{"id":"pg1","method":"pane.get","params":{"pane_id":"w5:p2"}}
{"id":"pg1","result":{"type":"pane_info","pane":{"pane_id":"w7:p2","workspace_id":"w7","tab_id":"w7:t1",…}}}
```

The `pane.get` call used the *retired* ID `w5:p2` from a new connection with no caller context. It found the current identity of the pane at `w7:p2`, with no error. A `pane.rename` through the same retired ID also succeeds.

Example: a move into a new tab. The result has the new tab and the previous and new IDs:

```json
{"id":"pm3","method":"pane.move","params":{"pane_id":"w5:p3","destination":{"type":"new_tab","workspace_id":"w5","label":"moved-tab"}}}
{"id":"pm3","result":{"type":"pane_move","move_result":{
  "changed":true,
  "previous_pane_id":"w5:p3","previous_workspace_id":"w5","previous_tab_id":"w5:t1",
  "pane":{"pane_id":"w5:p3","workspace_id":"w5","tab_id":"w5:t2","label":"build",…},
  "source_layout":{…},"target_layout":{…},
  "created_tab":{"tab_id":"w5:t2","workspace_id":"w5","number":2,"label":"moved-tab","pane_count":1,"focused":false,"agent_status":"unknown"},
  "focused_pane_id":"w5:p3"}}}
```

In this capture, the move stayed in `w5`. The numeric suffix of the pane did not change by chance: the ID was `w5:p3` before and after, and only the tab changed. The general rule stays the same. Always read the returned `pane` and continue with its `pane_id`. Do not continue with the ID before the move, although the old ID also works for `pane.*` methods.

Checked 2026-10-17 against herdr 0.9.5. The check included these steps:

- A pane moved `w5:p2` → `w6:p2` → `w7:p2`, and its `terminal_id` did not change.
- `pane.get` with each retired ID found `w7:p2`.
- `pane.rename` through `w5:p2` succeeded.
- `pane.current` with `caller_pane_id` set to `w5:p2` returned `w7:p2`.
- After `pane.close`, all three IDs returned `pane_not_found`.
- A move into a new tab returned `created_tab`. A move into a new workspace returned `created_workspace` and `created_tab`.
- `agent.get` with a retired pane ID returned `agent_not_found`.

## Resolving --current and caller context

herdr injects the identity of the calling pane into each managed pane as environment variables. Thus pane commands can use the pane that they run in as the target, with no ID:

- The injected context is `$HERDR_WORKSPACE_ID`, `$HERDR_TAB_ID`, and `$HERDR_PANE_ID`. Refer to [environment.md](environment.md#environment-variables).
- **Use `--current` on a pane command to make the calling pane the target**. skill.md: "An omitted `pane split` target uses the calling pane when `HERDR_PANE_ID` is available, otherwise the focused pane. Other commands may use the UI-focused pane, which can belong to the user or another client."
  - Thus an omitted target is *not* a safe default for most pane commands. It can act on the focused pane of a different person or client.
  - `pane split` is the one documented exception. It uses the calling pane from `$HERDR_PANE_ID` first, and the focused pane only after that.
  - A `$HERDR_PANE_ID` that finds no pane is not the same as no value. `pane current --current` with a stale value, such as the ID of a closed pane, returns `pane_not_found` (exit 1). It does not use the focused pane.
- The same explicit forms are `--pane "$HERDR_PANE_ID"` or a known ID from an earlier response. These pane subcommands accept `--pane <ID>` and `--current` for the same purpose: `pane current`, `pane layout`, `pane split`, `pane neighbor`, `pane edges`, `pane focus`, `pane resize`, `pane zoom`, `pane input`, `pane swap`, and `pane process-info`. In herdr 0.9.5, no other `pane`, `tab`, `workspace`, or `agent` subcommand has `--current`. For CLI rules, refer to [cli-mapping.md](cli-mapping.md#cli-conventions).
- `pane current --current` finds the calling pane and returns its full `PaneInfo`. It is the reliable way to learn your `pane_id`, `tab_id`, `workspace_id`, and current agent occupant. `pane current` without `--current` also uses the calling pane from `$HERDR_PANE_ID`. With a stale `$HERDR_PANE_ID`, it also returns `pane_not_found` (exit 1).

In this capture, the caller ran in `wE:p2` with `HERDR_WORKSPACE_ID=wE`, `HERDR_TAB_ID=wE:t1`, and `HERDR_PANE_ID=wE:p2`. The UI-focused pane was `wE:p1`.

```text
$ herdr pane split --direction down --no-focus                 → wE:p4, placed under wE:p2 (rect x=60,y=20): calling pane
$ env -u HERDR_PANE_ID herdr pane split --direction down --no-focus → wE:p5, placed under wE:p1 (rect x=0,y=20): focused pane
$ herdr pane zoom   (no target, from wE:p2)                    → zoom.pane_id "wE:p1": the UI-focused pane, not the caller
$ env HERDR_PANE_ID=wE:p3 herdr pane current --current      (wE:p3 closed)
{"error":{"code":"pane_not_found","message":"pane not found"},"id":"cli:pane:current"}
exit=1
```

```json
{"id":"cli:pane:current","result":{"pane":{"agent_status":"unknown","cwd":"/home/user/project","focused":false,"foreground_cwd":"/home/user/project","pane_id":"wE:p2","revision":1,…,"tab_id":"wE:t1",…,"workspace_id":"wE"},"type":"pane_current"}}
```

These discovery calls find live state from the caller context:

```text
herdr workspace list
herdr tab list --workspace "$HERDR_WORKSPACE_ID"
herdr pane current --current
herdr pane list --workspace "$HERDR_WORKSPACE_ID"
herdr agent list
```

Checked 2026-10-17 against herdr 0.9.5. The check included the target that `pane split` uses with and without `$HERDR_PANE_ID`, and the UI-focused target of `pane zoom`. It also included the stale `$HERDR_PANE_ID` case, for a closed pane and for a pane that never existed. All five discovery commands returned exit 0 in a scratch pane. The `--current` list came from the `--help` output of each `pane`, `tab`, `workspace`, and `agent` subcommand.

## Agent-name targeting

Agent commands (`herdr agent …` and the `agent.*` methods) accept one of two targets:

- a live agent name
- the ID of the pane that holds the agent now

They do **not** accept terminal IDs or bare agent-kind labels. The rules come from skill.md, section "Understand layout, panes, and agents":

- **Name grammar**. A name matches `[a-z][a-z0-9_-]{0,31}`. The first character is a lowercase letter. The other characters are lowercase letters, digits, `_`, or `-`. A name has a maximum of 32 characters.
  - Names outside the grammar fail with `invalid_agent_name`, such as `Bad`, `1abc`, `a.b`, `""`, and a name of 33 characters.
  - herdr accepts `ok-name_1`, `a-b`, and a name of 32 characters.
  - The herdr 0.9.5 error message is `"agent name must start with a lowercase letter and contain only lowercase letters, digits, '-' or '_' (1-32 characters)"`.
- **Uniqueness**. A name must be unique among live agents. Only *live* agents count. Thus you can use a name again after its holder goes away. The `agent_name_taken` error also names the live agent that has the name. The name is in the `message` text, not in separate JSON fields. Example: `agent name ok-name_1 is already used; candidates: terminal_id=… pane_id=w1:p2 workspace_id=w1 tab_id=w1:t2 cwd=… status=Idle`. This text helps a person decide the next step. The message is not a stable contract, so do not parse it.

  ```json
  {"id":"rn2","method":"agent.rename","params":{"target":"wF:p2","name":"reviewer"}}
  {"id":"rn2","error":{"code":"agent_name_taken","message":"agent name reviewer is already used; candidates: terminal_id=term_… pane_id=wF:p1 workspace_id=wF tab_id=wF:t1 cwd=/home/user/project status=Idle"}}
  ```

- **Lifetime and opacity**. skill.md:

  > "A name follows the current pane occupant and is cleared when that agent exits, is released, or is replaced."

  The name is a handle to the agent that is in the pane now. It is not a durable identity of a process. herdr 0.9.5 keeps the agent identity, and thus the name, through a Ctrl+Z suspend while the same process stays alive. Refer to [api/agent.md](api/agent.md#agent-identity-across-ctrlz) for that rule.
- **herdr does not accept kind labels and terminal IDs as targets**. `agent.get` with the agent kind (`demobot`) or with the `terminal_id` of the pane returns `agent_not_found`. Names are case-sensitive.

  ```text
  agent.get {"target":"reviewer"} → agent_info pane_id "wF:p1"
  agent.get {"target":"wF:p2"}    → agent_info (pane-ID target)
  agent.get {"target":"demobot"}  → {"code":"agent_not_found","message":"agent target demobot not found"}   (kind label)
  agent.get {"target":"term_…"}   → agent_not_found   (terminal_id)
  agent.get {"target":"REVIEWER"} → agent_not_found   (case-sensitive)
  ```

- **Pane ID as target**. You can also use the ID of the pane that holds the agent, such as `w2:p1`. After a `pane move`, use the new `pane_id` or the agent name as the target. The retired pane ID returns `agent_not_found` for `agent.get`. Refer to [Pane-move re-identification](#pane-move-re-identification).
- **Name assignment**. `agent start <name> --kind <kind> --pane <id>` gives a name to the agent that it starts. `agent rename <target> <name>|--clear` changes the name or clears it. For the CLI syntax, refer to [cli-mapping.md](cli-mapping.md).
  - `agent rename` refuses to set a name while the launch of the agent is pending (`launch_pending: true`). It also refuses to clear the name (`name: null`) at that time.
  - The refusal is the error `agent_launch_pending` ("agent name cannot change while startup is pending").
  - The `--skill` text gives `agent rename` with no condition, so it does not document this refusal.
  - After the launch settles, `agent rename` succeeds. `AgentInfo` then does not include `launch_pending`. It does not send `false`.

  ```json
  {"id":"as1","result":{"type":"agent_started","agent":{…,"name":"build","agent_status":"unknown",…,"launch_pending":true,…},"argv":["claude"]}}
  {"id":"rn1","method":"agent.rename","params":{"target":"wH:p2","name":"reviewer"}}
  {"id":"rn1","error":{"code":"agent_launch_pending","message":"agent name cannot change while startup is pending"}}
  {"id":"rn2","method":"agent.rename","params":{"target":"wH:p2","name":null}}
  {"id":"rn2","error":{"code":"agent_launch_pending","message":"agent name cannot change while startup is pending"}}
  ```

  - **Settle time**. On herdr 0.9.1, a launch usually settled in about 3.9 s. On herdr 0.9.1, the refusal can continue much longer when the started process waits for interactive input. An example is the first-run trust prompt of an agent program. On herdr 0.9.5, one launch of `claude` waited on its first-run theme screen. `launch_pending` cleared after 3.51 s, with `agent_status: "idle"` and `interactive_ready: true`. A different launch in the same setup settled in about 5 s. Thus a wait for input on that screen did not keep the launch pending on herdr 0.9.5. The trust-prompt case is not probed on herdr 0.9.5, because it needs the real configuration of the agent program.
- **Identity fields in `AgentInfo`** (refer to [data-model.md](data-model.md#agentinfo)):
  - `name`: the unique name that a caller can give. It can be null.
  - `agent`: the detected kind label, such as `claude` or `codex`. It can be null.
  - `agent_session`: an `AgentSessionInfo` with `source`, `agent`, `kind` ∈ `{id,path}`, and `value`. It holds the native session reference.

  Use `name` or the pane ID as a target. Read `agent` and `agent_session` as attributes, not as targets. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`. A probe on herdr 0.9.5 got `agent: "claude"` on a real launch and `agent: "demobot"` on a reported agent. The probe did not get an `agent_session` value, because the isolated agent did not start a session.

Checked 2026-10-17 against herdr 0.9.5. The check included these items:

- the name grammar and uniqueness
- the name clear after a release, and after a kill of the agent process
- the name free for a different agent after each of these two cases
- pane-ID targets, and the refusal of kind labels, terminal IDs, and retired pane IDs
- the `agent_launch_pending` refusal for set and clear, with a real `agent.start` launch
- `agent.start` with a `name`, which returned `agent_started` with that name

The case of a replaced agent is not probed separately.
