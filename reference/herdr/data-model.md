# herdr data model

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Index: [README.md](README.md). Wire format: [protocol.md](protocol.md).

This page gives the domain entities that the herdr socket API puts in method results and in pushed events. Each entity comes from the `$defs` of the `success_response` schema in [raw/schema.json](raw/schema.json). The result wrappers in `api/*.md` link to these definitions and do not repeat them. The field tables use the exact JSON names of the schema.

`T | null` marks a schema `type: ["T","null"]`, which is an optional field that can be null. In practice, herdr (0.9.1 to 0.9.5) does not send such a field when it has no value. herdr does not send an explicit `null` for it. Thus callers should treat a missing key and a `null` value as equal. Checked 2026-10-17 against herdr 0.9.5: 105 entity responses (`pane_info`, `agent_info`, `workspace_info`, `tab_info`, `session_snapshot`, `agent_list`, `pane_list`) held no `null`. Results that are not entities can hold an explicit `null`, such as `"matched_rule":null` in an `agent.explain` result.

Unless a row says "(inferred)", the schema gives the meaning, or the field name and the probes make the meaning clear. This folder does not keep the full probe captures.

This page owns the meaning of each [AgentStatus](#agentstatus) value. Other pages link to that section.

IDs are opaque handles: workspace `w1`, tab `w1:t1`, pane `w1:p1`. [addressing.md](addressing.md) owns the ID scheme. That page gives the counter alphabet after `9` and the new ID of a moved pane. It also gives the rule that herdr does not use a closed ID again. Do not parse the number at the end of an ID.

## Contents

| entity | role |
| --- | --- |
| [WorkspaceInfo](#workspaceinfo) | a workspace, which is the top-level container |
| [TabInfo](#tabinfo) | a tab in a workspace |
| [PaneInfo](#paneinfo) | a pane, which is one terminal area and can hold an agent |
| [AgentInfo](#agentinfo) | a detected agent in a pane |
| [AgentStatus](#agentstatus) | the agent status enum and the meaning of each value |
| [AgentSessionInfo](#agentsessioninfo) / [AgentSessionRefKind](#agentsessionrefkind) | the native session handle of an agent |
| [SessionSnapshot](#sessionsnapshot) | a full copy of the session state |
| [LayoutDescription](#layoutdescription) | a layout tree that a caller can export and apply |
| [LayoutNode](#layoutnode) | a recursive node of a layout tree (pane or split) |
| [SplitDirection](#splitdirection) / [PaneDirection](#panedirection) | direction enums |
| [PaneLayoutSnapshot](#panelayoutsnapshot) | the rendered geometry of a tab |
| [PaneLayoutPane](#panelayoutpane) / [PaneLayoutSplit](#panelayoutsplit) / [PaneLayoutRect](#panelayoutrect) | parts of the geometry |
| [PaneScrollInfo](#panescrollinfo) | the position of the scrollback viewport |
| [PaneProcessInfo](#paneprocessinfo) / [PaneProcessInfoProcess](#paneprocessinfoprocess) | the processes of a pane |
| [PaneReadResult](#panereadresult) / [ReadSource](#readsource) / [ReadFormat](#readformat) | terminal read output and its selectors |
| [PaneTextPoint](#panetextpoint) / [PaneTextRange](#panetextrange) | a row and column point and a span, in the copy-motion and copy-search results |
| [PaneLinkRegion](#panelinkregion) | the region of a resolved link on the viewport of a pane |
| [Pane operation result shapes](#pane-operation-result-shapes) | the results of swap, move, zoom, resize, edges, neighbor, and focus |
| [WorktreeInfo](#worktreeinfo) / [WorkspaceWorktreeInfo](#workspaceworktreeinfo) / [WorktreeSourceInfo](#worktreesourceinfo) | Git worktree entities |
| [ServerCapabilities](#servercapabilities) | the feature flags of a server (in `pong`) |
| [AgentManifestInfo](#agentmanifestinfo) | an installed agent manifest |
| [IntegrationInfo](#integrationinfo) / [IntegrationState](#integrationstate) | the install state of an integration |
| [IntegrationTarget](#integrationtarget) | the agent programs that an integration can target |
| [ConfigReloadStatus](#configreloadstatus) | the result enum of a config reload |
| [PopupSize](#popupsize) | the size value of a popup |
| [Plugin entities](#plugin-entities) | installed plugins, actions, panes, logs, and sources |
| [Event entities](#event-entities) | EventKind, EventEnvelope, and EventData |
| [Reason & auxiliary enums](#reason--auxiliary-enums) | the enums that give the reason for an operation result |
| [Appendix: on-disk persistence (session.json, version 3)](#appendix-on-disk-persistence-sessionjson-version-3) | the session file on the disk |

---

## WorkspaceInfo

A workspace is the top-level container that holds tabs. Each tab holds a pane tree. `workspace.get`, `workspace.list`, and `workspace.create` return a `WorkspaceInfo`. `SessionSnapshot` embeds it, and the `workspace_*` events carry it (see [events.md](events.md#event-catalog-eventkind--eventdata)).

| field | type | required | meaning |
| --- | --- | --- | --- |
| `workspace_id` | string | yes | opaque workspace handle, such as `w2` |
| `number` | integer (uint) | yes | 1-based display position in the sidebar. It can change (see below). |
| `label` | string | yes | display name: the custom name, or a derived name (see below) |
| `focused` | boolean | yes | this workspace has the UI focus |
| `pane_count` | integer (uint) | yes | the total number of panes in all its tabs |
| `tab_count` | integer (uint) | yes | the number of tabs |
| `active_tab_id` | string | yes | the ID of the active tab in this workspace |
| `agent_status` | [AgentStatus](#agentstatus) | yes | the rolled-up agent status of all panes in the workspace (see [AgentStatus](#agentstatus) for the order) |
| `tokens` | `object<string,string>` | no | a key-to-value metadata map that callers set. Keys match `^[A-Za-z0-9_-]{1,32}$`. The map holds 32 entries or fewer. |
| `worktree` | [WorkspaceWorktreeInfo](#workspaceworktreeinfo) \| null | no | the Git worktree of this workspace, if any (see below) |

Checked 2026-10-17 against herdr 0.9.5: the eight required keys, and `tokens` and `worktree` only after a caller set them.

The `number` field is a display position, not an ID. [addressing.md](addressing.md#display-numbers-vs-stable-ids-public-pane-numbers) owns the difference between IDs and display numbers. When a workspace closes, herdr at once changes the `number` of all other open workspaces. The numbers stay contiguous from 1, also for workspaces that had no other change. A workspace that a caller creates after the close receives the number (open workspace count + 1). It does not receive the free number. Thus a stored `number` can become incorrect after any workspace close in the session. A `workspace_id` does not change, and herdr does not use it again.

Checked 2026-10-17 against herdr 0.9.5. No other call came between the close and the next read:

```text
before: w1=1 w2=2 w3=3 w4=4 w5=5 w6=6 w7=7 w8=8 w9=9
workspace.close w3
after:  w1=1 w2=2 w4=3 w5=4 w6=5 w7=6 w8=7 w9=8
workspace.create {} -> wA number 9 (not the free number 3)
```

When a workspace has no custom name, herdr derives `label` from the `cwd`:

- When the creation `cwd` is a canonical path in a Git work tree, the label is the basename of the work-tree root. A `cwd` two levels below the root of the repository `project` gave `project`.
- When the creation `cwd` is a path through a symlink, the label is the basename of the `cwd`.
- When the `cwd` is not in a Git work tree, the label is the basename of the `cwd`.
- When the `cwd` is the home directory, the label is `~`.
- A linked worktree is its own work tree. Thus the label is the basename of the worktree directory, such as `wt-feat`.

The derived label also follows the shell. When the shell changes directory, the label changes to the basename of the new directory. herdr does not use the repository root for this later label. This pass did not find which pane sets the later label.

Checked 2026-10-17 against herdr 0.9.5:

```text
workspace.create {"cwd":"/home/user/project/a/b"}            -> label "project"
workspace.create {"cwd":"/home/user/project/a"}              -> label "project"
workspace.create {"cwd":"<path through a symlink>/project/a/b"} -> label "b"
workspace.create {"cwd":"/home/user/plain/leaf"}             -> label "leaf"
workspace.create {"cwd":"/home/user"}                        -> label "~"
linked worktree /home/user/wt-feat                           -> label "wt-feat"
workspace at /home/user/project/a/b, pane.send_text "cd ..\n" -> label "a"
workspace at /home/user/plain/leaf, "cd /home/user/project/a/b" -> label "b"
```

`tokens` is absent until a caller sets metadata. Then `tokens` is an object. The workspace limits are the same as the pane limits (see [PaneInfo](#paneinfo)). Checked 2026-10-17 against herdr 0.9.5:

```text
workspace.get w2 -> {…,"agent_status":"unknown"}                (no tokens key)
workspace.report_metadata {"workspace_id":"w2","source":"s1","tokens":{"build":"ok"}} -> ok
workspace.get w2 -> {…,"tokens":{"build":"ok"}}
tokens {"a b":"v"}            -> invalid_metadata_token "invalid metadata token key: a b"
17 keys in one call           -> invalid_metadata_token "a metadata report may update at most 16 tokens"
more than 32 keys in the map  -> metadata_token_limit "workspace metadata may contain at most 32 tokens"
```

`worktree.create` and `worktree.open` set `worktree`. A workspace that has only a `cwd` in a worktree does not carry `worktree`. Checked 2026-10-17 against herdr 0.9.5:

```text
git worktree add -b feat /home/user/wt-feat              (in /home/user/project)
workspace.create {"cwd":"/home/user/wt-feat"} -> workspace wB, no "worktree" key
worktree.open {"cwd":"/home/user/project","path":"/home/user/wt-feat"} -> workspace wB:
 "worktree":{"repo_key":"/home/user/project/.git","repo_name":"project","repo_root":"/home/user/project","checkout_path":"/home/user/wt-feat","is_linked_worktree":true}
```

## TabInfo

A tab is part of a workspace. A tab holds one pane tree. `tab.get`, `tab.list`, and `tab.create` return a `TabInfo`, and the `tab_*` events carry it.

| field | type | required | meaning |
| --- | --- | --- | --- |
| `tab_id` | string | yes | opaque tab handle, such as `w2:t1` |
| `workspace_id` | string | yes | the workspace that holds the tab |
| `number` | integer (uint) | yes | 1-based position in the workspace |
| `label` | string | yes | display name: the custom name, or else the number |
| `focused` | boolean | yes | this tab is the active tab of the workspace |
| `pane_count` | integer (uint) | yes | the number of panes in this tab |
| `agent_status` | [AgentStatus](#agentstatus) | yes | the rolled-up agent status of the panes in the tab |

Checked 2026-10-17 against herdr 0.9.5. `tab.create {"workspace_id":"wC"}` gave `"label":"3"` for number 3. `tab.create` with `"label":"reviewer"` gave `"label":"reviewer"`. After `tab.focus wC:t2`, `focused_tab_id` in the snapshot was `wC:t2`, and `tab.get wC:t1` gave `focused: false`.

## PaneInfo

A pane is one terminal area in the layout of a tab. A pane can exist with an agent or without one. When herdr detects an agent, herdr fills the agent fields. `pane.get`, `pane.list`, and `pane.current` return a `PaneInfo`, and the `pane_created` and `pane_updated` events carry it.

| field | type | required | meaning |
| --- | --- | --- | --- |
| `pane_id` | string | yes | opaque pane handle, such as `w2:p1` |
| `terminal_id` | string | yes | the stable terminal ID, such as `term_65e0dcf1927ac1`. The length can change. |
| `workspace_id` | string | yes | the workspace that holds the pane |
| `tab_id` | string | yes | the tab that holds the pane |
| `focused` | boolean | yes | this pane has the focus in its tab |
| `agent_status` | [AgentStatus](#agentstatus) | yes | the status of the agent in this pane. The value is `unknown` when the pane has no agent or herdr cannot classify it. |
| `revision` | integer (uint64) | yes | a monotonic output revision. It increases when the pane output changes. |
| `agent` | string \| null | no | the detected agent kind, such as `claude` |
| `display_agent` | string \| null | no | the agent label for people. `pane.report_metadata` sets it. |
| `agent_session` | [AgentSessionInfo](#agentsessioninfo) \| null | no | the native session handle of the agent, if herdr knows it |
| `name` | — | — | `PaneInfo` has no `name` field (see [AgentInfo](#agentinfo)) |
| `label` | string \| null | no | the pane label that a caller set |
| `title` | string \| null | no | the resolved pane title |
| `terminal_title` | string \| null | no | the raw terminal title. It can include status glyphs, such as `◐ …`. |
| `terminal_title_stripped` | string \| null | no | `terminal_title` without the status glyphs |
| `cwd` | string \| null | no | the working directory of the pane shell |
| `foreground_cwd` | string \| null | no | the working directory of the current foreground process |
| `scroll` | [PaneScrollInfo](#panescrollinfo) \| null | no | the position of the scrollback viewport |
| `state_labels` | `object<string,string>` | no | status annotations from the detector or from `pane.report_metadata` (label to value) |
| `tokens` | `object<string,string>` | no | the metadata map of the caller. Keys match `^[A-Za-z0-9_-]{1,32}$`. The map holds 32 entries or fewer. |
| `restore_error` | string \| null | no | new in 0.9.3: the reason that herdr could not restore this pane after a server restart. The pane is then a placeholder (see below). |

Checked 2026-10-17 against herdr 0.9.5: the field set and the required flags, and each optional field on a scratch server. With the live agent `reviewer`, `pane.get` had no `name` key, and `agent.get` had `"name":"reviewer"`. A real `claude` gave `"terminal_title":"✳ Claude Code"` and `"terminal_title_stripped":"Claude Code"`. While it worked, the title started with `◐` or `◑`. `pane.rename` set `label`, and the label stayed after the agent stopped. A real `claude` launch did not set `display_agent`. `pane.report_metadata` set `display_agent`, `title`, `state_labels`, and `tokens`:

```text
pane.report_metadata {"pane_id":"wE:p1","source":"probe","display_agent":"Reviewer Bot","title":"build","state_labels":{"idle":"waiting"},"tokens":{"task":"api-docs"}} -> ok
pane.get -> …"agent":"claude","title":"build","display_agent":"Reviewer Bot","state_labels":{"idle":"waiting"},"tokens":{"task":"api-docs"}…
```

The 32-entry limit applies to the `tokens` map. A different, lower limit applies to one call: a call can change 16 tokens or fewer. A `pane.report_metadata` call with 17 or more tokens fails with `invalid_metadata_token`, also when the map stays below 32 entries. To get from 17 to 32 tokens, use more than one call. A call that makes the map larger than 32 entries fails with a different code, `metadata_token_limit`. Checked 2026-10-17 against herdr 0.9.5:

```text
pane.report_metadata w2:p1 source s1, 17 tokens -> {"code":"invalid_metadata_token","message":"a metadata report may update at most 16 tokens"}
source s1, 16 tokens -> ok ; source s2, 16 tokens -> ok ; the map holds 32 tokens
source s3, 1 token   -> {"code":"metadata_token_limit","message":"pane metadata may contain at most 32 tokens"}
key "a.b" or a 33-character key -> invalid_metadata_token "invalid metadata token key: …" ; a 32-character key -> ok
```

`restore_error` (schema: string \| null, with no description) is absent on a normal pane. For this check, a pane started in a directory, and the directory was deleted while the shell was in it. After a restart of the scratch server, the pane came back with `restore_error`, without `scroll` or `foreground_cwd`, and with `revision` 0. The saved `cwd` had the suffix ` (deleted)`. A directory deleted after the server stop can give the plain path. This pass did not check that case. Checked 2026-10-17 against herdr 0.9.5:

```json
{"pane_id":"wF:p1","terminal_id":"term_65e0de41c759f20","workspace_id":"wF","tab_id":"wF:t1","focused":false,"cwd":"/home/user/project/gone (deleted)","restore_error":"Saved directory is unavailable. Restore the directory and restart this session.","agent_status":"unknown","revision":0}
```

Such a pane is a placeholder without a terminal. `pane.get` and `pane.list` return it. `pane.read`, `pane.process_info`, and `pane.send_text` fail with `pane_not_found`. Checked 2026-10-17 against herdr 0.9.5:

```text
pane.read wF:p1         -> {"code":"pane_not_found","message":"pane wF:p1 not found"}
pane.process_info wF:p1 -> {"code":"pane_not_found","message":"pane not found"}
pane.send_text wF:p1    -> {"code":"pane_not_found","message":"pane wF:p1 not found"}
```

A reported resume command can fail, such as with an unknown command name. This failure does not set `restore_error`, because herdr types the command into the restored shell. The shell then writes the error. Checked 2026-10-17 against herdr 0.9.5:

```text
pane.report_agent_session {"pane_id":"wC:p6","source":"probe","agent":"codex","agent_session_id":"0000-test","resume_argv":["examplebot-missing-cmd","--resume"]} -> ok
after a restart: wC:p6 has no restore_error. The screen holds:
$ examplebot-missing-cmd --resume
bash: examplebot-missing-cmd: command not found
```

This pass did not cause other `restore_error` messages.

## AgentInfo

`AgentInfo` gives the detected agent in a pane. It holds the agent fields of `PaneInfo` and more lifecycle fields. `agent.get`, `agent.list`, `agent.start`, and `agent.prompt` return it, and `SessionSnapshot.agents` embeds it. An agent target is a unique live agent `name` or the `pane_id` of the pane. Names match `[a-z][a-z0-9_-]{0,31}`. A name follows the agent that is in the pane now. For the identity of an agent across Ctrl+Z, see [api/agent.md](api/agent.md#agent-identity-across-ctrlz).

| field | type | required | meaning |
| --- | --- | --- | --- |
| `terminal_id` | string | yes | the terminal ID |
| `agent_status` | [AgentStatus](#agentstatus) | yes | the current agent status |
| `workspace_id` | string | yes | the workspace that holds the pane |
| `tab_id` | string | yes | the tab that holds the pane |
| `pane_id` | string | yes | the pane of the agent |
| `focused` | boolean | yes | the pane has the focus |
| `revision` | integer (uint64) | yes | the output revision |
| `agent` | string \| null | no | the agent kind, such as `claude` |
| `display_agent` | string \| null | no | the agent label for people |
| `agent_session` | [AgentSessionInfo](#agentsessioninfo) \| null | no | the native session handle |
| `name` | string \| null | no | the live agent name that the caller gave. It is a target handle. |
| `cwd` | string \| null | no | the working directory of the shell |
| `foreground_cwd` | string \| null | no | the working directory of the foreground process |
| `interactive_ready` | boolean | no | the agent is ready for interactive input |
| `launch_pending` | boolean | no | a launch is in progress, and herdr has not detected the agent as ready |
| `screen_detection_skipped` | boolean | no | herdr did not use screen detection, such as when an integration gives the status |
| `state_change_seq` | integer (uint64) | no (default 0) | a monotonic counter that increases on each lifecycle change. Use it to find status transitions. |
| `completion_seq` | integer (uint64) \| null | no | new in 0.9.3. The schema says: "The current idle transition completed work, independently of who has viewed it." See below. |
| `state_labels` | `object<string,string>` | no | status annotations |
| `title` | string \| null | no | the resolved title |
| `terminal_title` | string \| null | no | the raw terminal title |
| `terminal_title_stripped` | string \| null | no | the title without the status glyphs |
| `tokens` | `object<string,string>` | no | the metadata map of the caller (the same limits as `PaneInfo`) |

Note: `PaneInfo` has `scroll`, `label`, and `restore_error`. `AgentInfo` has `name`, `interactive_ready`, `launch_pending`, `screen_detection_skipped`, `state_change_seq`, and `completion_seq` instead.

Checked 2026-10-17 against herdr 0.9.5: the field set and the required flags, and the return from `agent.get`, `agent.list`, `agent.start`, and `session.snapshot`. Both a `name` target and a `pane_id` target worked. After the agent process was killed, `agent.get reviewer` gave `{"code":"agent_not_found","message":"agent target reviewer not found"}`, and the pane label `reviewer` stayed. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the return from `agent.prompt`, because this pass typed prompts with `pane.send_text`.

A real `agent.start` of `claude` gave `launch_pending: true` at the start. The key stayed while the folder-trust dialog made the agent `blocked`. The key was absent after `interactive_ready` became `true`. Checked 2026-10-17 against herdr 0.9.5:

```text
agent.start {"name":"reviewer","kind":"claude","pane_id":"wE:p1","timeout_ms":30000}
-> {"type":"agent_started","agent":{"terminal_id":"term_65e0dda9370f21c","name":"reviewer",…,"agent_status":"unknown","workspace_id":"wE","tab_id":"wE:t1","pane_id":"wE:p1","focused":false,"launch_pending":true,"state_change_seq":0,"cwd":"/home/user/project","foreground_cwd":"/home/user/project","revision":1},"argv":["claude"]}
about 8 s later (trust dialog on the screen): agent.get -> "agent_status":"blocked","launch_pending":true,"state_change_seq":13
after an answer to the trust dialog: agent.get ->
{"terminal_id":"term_65e0dda9370f21c","name":"reviewer","agent":"claude","terminal_title":"✳ Claude Code","terminal_title_stripped":"Claude Code","agent_status":"idle","agent_session":{"source":"herdr:claude","agent":"claude","kind":"id","value":"35a01b13-…"},"workspace_id":"wE","tab_id":"wE:t1","pane_id":"wE:p1","focused":false,"interactive_ready":true,"state_change_seq":14,"revision":2}
```

After `pane.report_metadata`, `display_agent`, `title`, `state_labels`, and `tokens` were on `AgentInfo` with the same values as on `PaneInfo`. Checked 2026-10-17 against herdr 0.9.5. `screen_detection_skipped` was not on `AgentInfo` for a reported agent or for a real `claude`. `agent.explain` has its own `screen_detection_skipped` field, which was `false`. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the `AgentInfo.screen_detection_skipped` field.

`completion_seq` records that work completed, also when nobody saw it. After `working` and then `idle`, `completion_seq` was equal to the `state_change_seq` of that idle transition. A focus that changed `done` to `idle` did not change `completion_seq` or `state_change_seq`. This was true for `agent.focus` and for `tab.focus` of the tab of the agent. `completion_seq` was absent while the agent was `working` or `blocked`. A later sequence of `working`, `blocked`, and `idle` set it to the `state_change_seq` of the new idle transition. `done` occurs only while the tab of the agent is not focused. To find "the agent completed a new turn", compare `completion_seq` with a stored value. This test does not depend on the difference between `idle` and `done`.

`state_change_seq` is one counter for the server, not one counter for each pane. A second pane started at 3, not at 1. After a server restart, the counter started again. Checked 2026-10-17 against herdr 0.9.5, with agents that `pane.report_agent` drove and with one real `claude`:

```text
wC:p2 (tab wC:t2 not focused, sleep 3000 in the foreground):
  working -> state_change_seq 1, no completion_seq ; idle -> done, state_change_seq 2, completion_seq 2
wC:p1 (focused tab): working -> 3 (no completion_seq) ; blocked -> 4 (no completion_seq) ; idle -> idle, 5/5
tab.focus wC:t2 -> wC:p2 idle, 2/2 (no change)
wC:p1 working, idle (its tab not focused now) -> done 7/7 ; agent.focus wC:p1 -> idle 7/7
real claude after a typed prompt (workspace not focused): done, state_change_seq 16, completion_seq 16
```

## AgentStatus

Enum. The lifecycle value that herdr gives to an agent in a pane. This section owns the meaning of each value.

| value | meaning |
| --- | --- |
| `idle` | The agent is normally ready for input, **and** a person saw its tab in the focused herdr UI. |
| `working` | The agent is processing a turn. |
| `blocked` | herdr recognized an approval UI or a question UI. The agent waits for a person. |
| `done` | The same idle status as `idle`, which herdr gives after work completed in a tab that nobody saw. |
| `unknown` | An agent is in the pane, but herdr cannot classify it with confidence. It does **not** prove completion. A pane with no agent also gives `unknown`. |

`Documented by herdr --skill (herdr 0.9.5); not live-checked: the table gives herdr's own definitions, which no single probe proves`. This stamp applies to the meaning of each value in this table. The probe rows below give the checked behavior.

Lifecycle meaning (from `herdr --skill`, copy in [raw/skill.md](raw/skill.md)):

- `idle` and `done` normally show that the agent is ready for input. Neither value guarantees that herdr recognized an unfamiliar approval dialog. Neither value guarantees that a retained agent is in the foreground. For a retained agent after Ctrl+Z, see [api/agent.md](api/agent.md#agent-identity-across-ctrlz).
- Antigravity and Grok can be `idle` while background tasks or subagents continue to run. Thus a wait for a settled status does not prove that those jobs are complete.
- The only difference between `idle` and `done` is the seen state of the tab on the server. A focus of the tab, or a focus command on the pane or the agent, marks the target seen. This changes `done` to `idle`. **CLI reads do not mark a tab seen.** Thus a read of the agent output does not change `done`.
- Each TUI client records the completions that it showed independently. Thus the Done badge of a client can be different from the CLI or from the badge of another client.
- `blocked` means that herdr detected an approval dialog or a question dialog. Examine it with `agent get` and `agent read`. Ask the person before you answer it.
- `unknown` does not show completion.

`Documented by herdr --skill (herdr 0.9.5); not live-checked: the checks need an unfamiliar approval dialog, an agent retained after Ctrl+Z, Antigravity or Grok, and an attached TUI client`. This stamp applies to the readiness limits in the first two items and to the Done badge of each TUI client. This pass had none of these four conditions.

`agent.explain` gives the detection rule that set the current status and the evidence text. The CLI form `herdr agent explain` gives these lines for a real `claude` that worked. The title in the evidence is anonymized. Checked 2026-10-17 against herdr 0.9.5:

```text
$ herdr agent explain reviewer
agent: claude
state: working
manifest: remote:/home/user/.local/state/herdr/agent-detection/remote/claude.toml 2026.10.11.1
rule: osc_title_working (region=osc_title priority=1100)
evidence: "◐ api-docs"
```

For the same agent when idle, the rule was `live_prompt_box (region=prompt_box_body priority=950)` and the evidence was `"❯\n"`. The JSON result has more fields, such as `state`, `matched_rule` (an object or `null`), `evaluated_rules[].evidence.region_preview`, `fallback_reason`, and `manifest_source`.

The tab and the workspace roll up the status of their panes in `TabInfo.agent_status` and `WorkspaceInfo.agent_status`. The rolled-up value is the status with the highest priority. The observed order is `blocked`, then `done`, then `working`, then `idle`. The value is `unknown` when no pane has an agent. This pass did not compare `unknown` with the other values. Checked 2026-10-17 against herdr 0.9.5:

```text
tab wC:t1 with wC:p1 done    + wC:p8 blocked -> tab blocked, workspace blocked
           wC:p1 done    + wC:p8 working -> tab done
           wC:p1 idle    + wC:p8 working -> tab working
```

Checked 2026-10-17 against herdr 0.9.5, with agents that `pane.report_agent` drove (reports `working`, `blocked`, and `idle`) and with one real `claude`:

- The pass observed all five values. A real `claude` folder-trust dialog (`Quick safety check … Yes, I trust this folder … Enter to confirm`) gave `blocked` through detection.
- An `idle` report that followed `working`, for an agent in a tab that was not focused, gave `done`. An `idle` report with no earlier work gave `idle`, also in a tab that was not focused. A real `claude` at its first prompt in a workspace that was not focused gave `idle`. After one turn, it gave `done`.
- An `idle` report for an agent in the focused tab gave `idle` at once.
- `agent.focus` and `tab.focus` changed `done` to `idle`.
- `agent.read`, `pane.read`, `agent.explain`, and the CLI `herdr agent read` did not change `done`.

Changed in 0.9.5: some reports for a pane with the plain shell in the foreground stay. On 0.9.3, herdr released such an agent at once after an `idle` report (`pane_agent_detected` with `released: true`), and the pane gave `unknown`. On 0.9.5, the agent kind set the result. herdr kept a report for a recognized kind (`claude`, `codex`) for 11 s or more. For `claude`, the pane kept `"agent":"claude"` with `done`. An `idle` report with no earlier work for `claude` also stayed, with `idle`. herdr dropped a report for an unrecognized kind (`examplebot`, `probe`) within 1 s (0.2 to 0.3 s in this pass). The pane then gave `unknown`. [api/pane.md](api/pane.md#panereport_agent) owns the report and release behavior. Checked 2026-10-17 against herdr 0.9.5:

```text
report working -> event pane_agent_detected {"agent":"claude","pane_id":"wD:p1"} ; status working
report idle    -> status done ; no pane_agent_detected with released:true
10 s later: pane.get wD:p1 -> "agent":"claude","agent_status":"done"
```

## AgentSessionInfo

The native session handle of a detected agent. herdr uses it to link the agent in a pane to the native session of that agent program, for restore and for integrations. A real `claude` (through its integration) gave `{"source":"herdr:claude","agent":"claude","kind":"id","value":"35a01b13-…"}`. Checked 2026-10-17 against herdr 0.9.5.

| field | type | required | meaning |
| --- | --- | --- | --- |
| `source` | string | yes | the provider or namespace of the reference, such as `herdr:claude` |
| `agent` | string | yes | the agent kind, such as `claude` |
| `kind` | [AgentSessionRefKind](#agentsessionrefkind) | yes | `value` is a session ID or a filesystem path |
| `value` | string | yes | the session ID or the path |

A `pane.report_agent_session` with source `herdr:claude`, agent `claude`, and `agent_session_id` `0000-test` set the handle on `pane.get` and on `agent.get`. The same report for an unknown agent kind (`herdr:examplebot`) gave `ok`, but no `agent_session` occurred on the pane or on the agent. This was also true after `pane.report_agent` made `examplebot` the agent of the pane. Thus herdr gives the handle only for agent kinds that it recognizes. Checked 2026-10-17 against herdr 0.9.5:

```text
pane.report_agent_session {"pane_id":"wC:p2","source":"herdr:claude","agent":"claude","agent_session_id":"0000-test"} -> ok
pane.get / agent.get wC:p2 -> "agent_session":{"source":"herdr:claude","agent":"claude","kind":"id","value":"0000-test"}
pane.report_agent_session {"pane_id":"wC:p3","source":"herdr:examplebot","agent":"examplebot","agent_session_id":"0000-test"} -> ok ; no agent_session
```

New in 0.9.3: `pane.report_agent` and `pane.report_agent_session` accept `resume_argv`. Its type is an array of strings or null. The schema says: "Command that resumes this agent's session after a Herdr restart. The first element must be a plain command name." It is a request parameter, not part of `AgentSessionInfo`. No response entity returns it. herdr does these steps with it:

1. herdr checks it. An empty array or a first element that is a path fails with `invalid_resume_argv`.
2. herdr accepts it only from the reporter that holds the pane. Another source fails with `resume_not_accepted`.
3. herdr saves it as `agent_resume` of the pane in `session.json` (see the [appendix](#appendix-on-disk-persistence-sessionjson-version-3)).
4. On a restart, herdr types it into the shell of the restored pane. A pane with `resume_argv: ["sleep","3000"]` came back with `sleep 3000` in the foreground.

Checked 2026-10-17 against herdr 0.9.5:

```text
resume_argv []                               -> invalid_resume_argv "resume_argv must not be empty" (the same for pane.report_agent)
["/bin/sh"], ["./examplebot"], ["my cmd"]    -> invalid_resume_argv "resume_argv must start with a plain command name, not a path"
["sleep","3000"] from source herdr:claude, on a pane that source probe holds
  -> resume_not_accepted "resume_argv requires the reporter to hold the pane; report its state with pane.report_agent first"
the same from source probe -> ok
session.json (version 3): "agent_resume":{"source":"probe","agent":"claude","argv":["sleep","3000"]}
after a restart: pane.process_info wC:p2 -> foreground argv ["sleep","3000"]
```

The [api/pane.md](api/pane.md#panereport_agent_session) page gives the parameters of these methods. [errors.md](errors.md) owns the error codes.

New fact on 0.9.5: herdr can resume a `claude` agent session without `resume_argv`. A pane had the agent `claude` with the `agent_session` `{"source":"herdr:claude","kind":"id","value":"0001-test"}` and no `resume_argv`. After a server stop and a start, herdr ran `claude --resume 0001-test` in the restored shell. The real `claude` process showed its folder-trust dialog, and `agent.list` gave the pane as `claude` with `blocked`. When a pane has `resume_argv`, herdr uses that command instead. Checked 2026-10-17 against herdr 0.9.5.

**Warning for probes**: a test `agent_session` for a recognized agent kind starts that real agent CLI on the next restart. Run such probes only on a scratch server. Stop the agent process after the check.

## AgentSessionRefKind

Enum for `AgentSessionInfo.kind`.

| value | meaning |
| --- | --- |
| `id` | `value` is a native session ID |
| `path` | `value` is a filesystem path to the session |

`id`: Checked 2026-10-17 against herdr 0.9.5 (see [AgentSessionInfo](#agentsessioninfo)). `path`: `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: path-only reports for `claude`, `codex`, and `pi` on new panes gave `ok`, but no `agent_session` occurred. A later report with `agent_session_id` on the same pane gave a handle of kind `id`. A later report with `agent_session_path` did not replace an existing `id` handle. Checked 2026-10-17 against herdr 0.9.5.

## SessionSnapshot

A full copy of the session state: each workspace, tab, pane, layout, and agent, and the current focus. The `session.snapshot` method returns it (`result.type = session_snapshot`) in `result.snapshot`.

| field | type | required | meaning |
| --- | --- | --- | --- |
| `version` | string | yes | the herdr version that made the snapshot, such as `0.9.5` |
| `protocol` | integer (uint32) | yes | the protocol number, such as `22` |
| `workspaces` | array of [WorkspaceInfo](#workspaceinfo) | yes | all workspaces |
| `tabs` | array of [TabInfo](#tabinfo) | yes | all tabs in all workspaces |
| `panes` | array of [PaneInfo](#paneinfo) | yes | all panes |
| `layouts` | array of [PaneLayoutSnapshot](#panelayoutsnapshot) | yes | one rendered layout for each tab |
| `agents` | array of [AgentInfo](#agentinfo) | yes | all detected agents |
| `focused_workspace_id` | string \| null | no | the focused workspace |
| `focused_tab_id` | string \| null | no | the focused tab |
| `focused_pane_id` | string \| null | no | the focused pane |

The five arrays are `[]` when empty. herdr does not omit them. The three `focused_*` fields are absent until an item has the focus. Then all three occur together. Checked 2026-10-17 against herdr 0.9.5, on a live session (read-only) and on a scratch server. The live snapshot had the ten keys of the table, `"version":"0.9.5"`, and `"protocol":22`. On the scratch server, the count of layouts was equal to the count of tabs:

```text
new scratch server: {"type":"session_snapshot","snapshot":{"version":"0.9.5","protocol":22,"workspaces":[],"tabs":[],"panes":[],"layouts":[],"agents":[]}}
after the first workspace.create: focused_workspace_id w1, focused_tab_id w1:t1, focused_pane_id w1:p1
agents entries held completion_seq, such as ("wC:p1","done",completion_seq 8)
```

## LayoutDescription

A portable description of the layout of a tab, as a recursive split tree. `layout.export`, `layout.apply`, and `layout.set_split_ratio` use this shape (`result.type` of `layout_export`, `layout_apply`, and `layout_split_ratio_set`). [PaneLayoutSnapshot](#panelayoutsnapshot) gives cell geometry. This shape is different: it is the logical tree with ratios and pane specs, which a caller can apply again.

| field | type | required | meaning |
| --- | --- | --- | --- |
| `workspace_id` | string | yes | the workspace that holds the tab |
| `tab_id` | string | yes | the tab that the layout describes |
| `zoomed` | boolean | yes | a pane is zoomed now |
| `focused_pane_id` | string | yes | the focused pane in the tree |
| `root` | [LayoutNode](#layoutnode) | yes | the root of the split tree |

Checked 2026-10-17 against herdr 0.9.5. All three result types gave a `layout` object with this shape. An export that a caller applied again gave the same structure:

```text
>> layout.export {"tab_id":"w1:t1"}
<< {"id":"3","result":{"type":"layout_export","layout":{"workspace_id":"w1","tab_id":"w1:t1","zoomed":false,"focused_pane_id":"w1:p1","root":{"type":"split","direction":"right","ratio":0.5,"first":{"type":"split","direction":"down","ratio":0.5,"first":{"type":"pane","pane_id":"w1:p1","cwd":"/home/user/project"},"second":{"type":"pane","pane_id":"w1:p3","cwd":"/home/user/project"}},"second":{"type":"pane","pane_id":"w1:p2","cwd":"/home/user/project"}}}}}
>> layout.set_split_ratio {"tab_id":"w1:t1","path":[],"ratio":0.6}
<< {"id":"6","result":{"type":"layout_split_ratio_set","layout":{"workspace_id":"w1","tab_id":"w1:t1","zoomed":false,"focused_pane_id":"w1:p1","root":{"type":"split","direction":"right","ratio":0.6,…}}}}
```

## LayoutNode

A recursive node of the layout tree. It is a tagged union with the discriminator `type` and two variants: a leaf `pane` or an inner `split`.

Variant `pane` (a leaf terminal):

| field | type | required | meaning |
| --- | --- | --- | --- |
| `type` | const `"pane"` | yes | discriminator |
| `pane_id` | string \| null | no | On export, the live pane ID. On apply, a hint only: apply gives new pane IDs (see [api/layout.md](api/layout.md#layoutapply)). Apply accepts `null`. |
| `label` | string \| null | no | the pane label |
| `cwd` | string \| null | no | the working directory for the launch |
| `command` | array of `string` \| null | no | the command and its arguments for the pane |
| `env` | `object<string,string>` | no | environment overrides. Apply uses them, but no result returns them. |

Variant `split` (an inner division):

| field | type | required | meaning |
| --- | --- | --- | --- |
| `type` | const `"split"` | yes | discriminator |
| `direction` | [SplitDirection](#splitdirection) | yes | `right` (side by side) or `down` (one above the other) |
| `ratio` | number (float) | yes | the fraction of the space that `first` receives (0 to 1) |
| `first` | [LayoutNode](#layoutnode) | yes | the first child subtree |
| `second` | [LayoutNode](#layoutnode) | yes | the second child subtree |

`layout.apply` does not bind an existing pane through `pane_id`. Apply gives new pane IDs. An apply with `tab_id` replaces the tab, so the old panes no longer exist. A result omits `label` and `command` when they have no value. A result never holds `env`. Thus a new apply of an export runs without the earlier environment overrides. [api/layout.md](api/layout.md#layoutapply) owns this behavior.

Checked 2026-10-17 against herdr 0.9.5. The check included both variants, a nested split, labels, a command, `env`, and `pane_id: null`:

```json
{"id":"1","method":"layout.apply","params":{"workspace_id":"w1","tab_label":"applied","root":{"type":"split","direction":"down","ratio":0.7,"first":{"type":"pane","label":"build","cwd":"/home/user/project","command":["sh","-c","echo E=$DMX; exec sleep 600"],"env":{"DMX":"envval"}},"second":{"type":"split","direction":"right","ratio":0.4,"first":{"type":"pane","label":"reviewer","cwd":"/home/user/project"},"second":{"type":"pane","pane_id":null,"cwd":"/home/user/project"}}}}}
{"id":"1","result":{"type":"layout_apply","layout":{"workspace_id":"w1","tab_id":"w1:t2","zoomed":false,"focused_pane_id":"w1:p4","root":{"type":"split","direction":"down","ratio":0.7,"first":{"type":"pane","pane_id":"w1:p4","label":"build","cwd":"/home/user/project","command":["sh","-c","echo E=$DMX; exec sleep 600"]},"second":{"type":"split","direction":"right","ratio":0.4,"first":{"type":"pane","pane_id":"w1:p5","label":"reviewer","cwd":"/home/user/project"},"second":{"type":"pane","pane_id":"w1:p6","cwd":"/home/user/project"}}}}}}
```

The pane `w1:p4` gave `E=envval`. A pane from a new apply of the export gave `E=`. After an apply with `tab_id` `w1:t2`, `pane.get w1:p4` gave `pane_not_found`. The export kept clean ratios (`0.7`, `0.6`). The rendered snapshot in the `layout_updated` event gave float noise, such as `0.699999988079071`.

## SplitDirection

Enum. The direction in which a split divides the space. `LayoutNode`, `PaneLayoutSplit`, and `pane.split` use it.

| value | meaning |
| --- | --- |
| `right` | the children are side by side, with `first` on the left |
| `down` | the children are one above the other, with `first` on top |

Checked 2026-10-17 against herdr 0.9.5. In a 120 x 40 cell area, `pane.split` `right` on `w1:p1` put `w1:p2` at `x` 60. A split `down` put `w1:p3` at `y` 20, below `w1:p1`:

```text
"panes":[{"pane_id":"w1:p1","focused":true,"rect":{"x":0,"y":0,"width":60,"height":20}},{"pane_id":"w1:p3","focused":false,"rect":{"x":0,"y":20,"width":60,"height":20}},{"pane_id":"w1:p2","focused":false,"rect":{"x":60,"y":0,"width":60,"height":40}}],
"splits":[{"id":"split_0_root","direction":"right","ratio":0.5,…},{"id":"split_1_0","direction":"down","ratio":0.5,…}]
```

## PaneDirection

Enum. A direction with four values. The neighbor, edge, focus, resize, and swap methods use it.

| value | meaning |
| --- | --- |
| `left` | toward the neighbor on the left |
| `right` | toward the neighbor on the right |
| `up` | toward the pane above |
| `down` | toward the pane below |

Checked 2026-10-17 against herdr 0.9.5. `pane.neighbor` on the top-left pane of three gave `w1:p2` for `right` and `w1:p3` for `down`. For `left` and `up`, the result had no neighbor key.

## PaneLayoutSnapshot

The rendered geometry of a tab: the cell rectangles of each pane and each split. `pane.layout` returns it (`type: "pane_layout"`). Most pane operation results and the `layout_updated` event embed it.

| field | type | required | meaning |
| --- | --- | --- | --- |
| `workspace_id` | string | yes | the workspace that holds the tab |
| `tab_id` | string | yes | the tab that the geometry describes |
| `zoomed` | boolean | yes | a pane is zoomed |
| `area` | [PaneLayoutRect](#panelayoutrect) | yes | the full area of the tab, in cells |
| `focused_pane_id` | string | yes | the focused pane |
| `panes` | array of [PaneLayoutPane](#panelayoutpane) | yes | the panes and their rectangles |
| `splits` | array of [PaneLayoutSplit](#panelayoutsplit) | yes | the split dividers and their rectangles |

Checked 2026-10-17 against herdr 0.9.5:

```text
>> pane.layout {"pane_id":"w1:p1"}
<< {"id":"4","result":{"type":"pane_layout","layout":{"workspace_id":"w1","tab_id":"w1:t1","zoomed":false,"area":{"x":0,"y":0,"width":120,"height":40},"focused_pane_id":"w1:p1","panes":[…],"splits":[…]}}}
```

## PaneLayoutPane

The position of a pane in a [PaneLayoutSnapshot](#panelayoutsnapshot).

| field | type | required | meaning |
| --- | --- | --- | --- |
| `pane_id` | string | yes | the pane |
| `focused` | boolean | yes | this pane has the focus |
| `rect` | [PaneLayoutRect](#panelayoutrect) | yes | its cell rectangle |

Checked 2026-10-17 against herdr 0.9.5 (see the capture in [SplitDirection](#splitdirection)).

## PaneLayoutSplit

The position of a split divider in a [PaneLayoutSnapshot](#panelayoutsnapshot).

| field | type | required | meaning |
| --- | --- | --- | --- |
| `id` | string | yes | the split ID |
| `direction` | [SplitDirection](#splitdirection) | yes | the direction of the split |
| `ratio` | number (float) | yes | the fraction of the first child |
| `rect` | [PaneLayoutRect](#panelayoutrect) | yes | the cell rectangle of the divider |

This `id` is an ID of the rendered geometry only. `layout.set_split_ratio` does not accept it. The request takes a `path` array of booleans instead. The path goes down the [LayoutNode](#layoutnode) tree to the target split, and `path: []` selects the root split. A request with `split_id` instead fails with `invalid_request`. Checked 2026-10-17 against herdr 0.9.5. The root split had the `id` `split_0_root`, the second split `split_1_0`, and the third split `split_2_00`:

```text
>> layout.set_split_ratio {"tab_id":"w1:t1","split_id":"split_0_root","ratio":0.6}
<< {"id":"5","error":{"code":"invalid_request","message":"invalid request: missing field `path` at line 1 column 120"}}
>> layout.set_split_ratio {"tab_id":"w1:t1","path":[],"ratio":0.6}
<< {"id":"6","result":{"type":"layout_split_ratio_set",…,"root":{"type":"split","direction":"right","ratio":0.6,…}}}}
```

## PaneLayoutRect

A rectangle in terminal cells. All fields are `uint16` (0 to 65535).

| field | type | required | meaning |
| --- | --- | --- | --- |
| `x` | integer (uint16) | yes | the left column, in cells |
| `y` | integer (uint16) | yes | the top row, in cells |
| `width` | integer (uint16) | yes | the width in cells |
| `height` | integer (uint16) | yes | the height in cells |

Checked 2026-10-17 against herdr 0.9.5: the four fields in cells. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the `uint16` range, because this pass did not test the limits.

## PaneScrollInfo

The position of the scrollback viewport of a pane (`PaneInfo.scroll`). All fields are `uint64`.

| field | type | required | meaning |
| --- | --- | --- | --- |
| `offset_from_bottom` | integer (uint64) | yes | the rows above the live bottom. The value 0 means that the viewport is at the bottom. |
| `max_offset_from_bottom` | integer (uint64) | yes | the maximum offset, which is the depth of the history |
| `viewport_rows` | integer (uint64) | yes | the visible rows in the viewport |

Checked 2026-10-17 against herdr 0.9.5: the three fields, after `seq 1 100` in a pane that was 12 rows high. `pane.scroll` limited a large offset to `max_offset_from_bottom`. `viewport_rows` was 2 less than the height of the pane rectangle. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the `uint64` type.

```text
scroll after seq: {"offset_from_bottom":0,"max_offset_from_bottom":94,"viewport_rows":10}
>> pane.scroll {"pane_id":"w1:p9","offset_from_bottom":5}      -> "scroll":{"offset_from_bottom":5,"max_offset_from_bottom":94,"viewport_rows":10}
>> pane.scroll {"pane_id":"w1:p9","offset_from_bottom":100000} -> "scroll":{"offset_from_bottom":94,"max_offset_from_bottom":94,"viewport_rows":10}
```

## PaneProcessInfo

The processes of a pane. `pane.process_info` returns it (`result.type = pane_process_info`) in the key `process_info`.

| field | type | required | meaning |
| --- | --- | --- | --- |
| `pane_id` | string | yes | the pane |
| `shell_pid` | integer (uint32) \| null | no | the PID of the pane shell |
| `tty` | string \| null | no | the path of the controlling tty |
| `foreground_process_group_id` | integer (uint32) \| null | no | the ID of the foreground process group |
| `foreground_processes` | array of [PaneProcessInfoProcess](#paneprocessinfoprocess) | no | the processes in the foreground group |

Checked 2026-10-17 against herdr 0.9.5: `pane_id`, `shell_pid`, `foreground_process_group_id`, and `foreground_processes`. `tty` was absent on Linux, as on 0.9.1 and 0.9.3:

```text
>> pane.process_info {"pane_id":"w1:p3"}
<< {"id":"2","result":{"type":"pane_process_info","process_info":{"pane_id":"w1:p3","shell_pid":215115,"foreground_process_group_id":215115,"foreground_processes":[{"pid":215115,"name":"bash","argv":["/usr/bin/bash"],"cmdline":"/usr/bin/bash","cwd":"/home/user/project"}]}}}
```

## PaneProcessInfoProcess

One process in the foreground group of a pane.

| field | type | required | meaning |
| --- | --- | --- | --- |
| `pid` | integer (uint32) | yes | the process ID |
| `name` | string | yes | the process name, such as `bash` or `claude` |
| `argv0` | string \| null | no | `argv[0]` as the process started |
| `argv` | array of `string` \| null | no | the full argument vector |
| `cmdline` | string \| null | no | the joined command line |
| `cwd` | string \| null | no | the working directory of the process |

Checked 2026-10-17 against herdr 0.9.5: `pid`, `name`, `argv`, `cmdline`, and `cwd` for a plain login shell. `argv0` was absent in two cases with `exec -a`, also when `argv[0]` was different from the executable name. In those cases, `argv[0]` gave the `-a` name, and `name` gave the executable. This pass did not find when `argv0` and `tty` occur.

```text
shell "exec -a demo-bot sleep 3000" (replaces the shell):
<< {"id":"4","result":{"type":"pane_process_info","process_info":{"pane_id":"w1:p3","shell_pid":215115,"foreground_process_group_id":215115,"foreground_processes":[{"pid":215115,"name":"sleep","argv":["demo-bot","3000"],"cmdline":"demo-bot 3000","cwd":"/home/user/project"}]}}}
bash -c 'exec -a demo-bot2 sleep 3000' (a child process group):
<< {"id":"6","result":{"type":"pane_process_info","process_info":{"pane_id":"w1:p8","shell_pid":218622,"foreground_process_group_id":234437,"foreground_processes":[{"pid":234437,"name":"sleep","argv":["demo-bot2","3000"],"cmdline":"demo-bot2 3000","cwd":"/home/user/project"}]}}}
```

## PaneReadResult

The output of a read of the screen or the scrollback of a pane. `pane.read` returns it. The `output_matched` result of `pane.wait_for_output` embeds it in the key `read`. There is no `pane.wait_output` method.

| field | type | required | meaning |
| --- | --- | --- | --- |
| `pane_id` | string | yes | the pane that herdr read |
| `workspace_id` | string | yes | the workspace that holds the pane |
| `tab_id` | string | yes | the tab that holds the pane |
| `source` | [ReadSource](#readsource) | yes | the snapshot that herdr read |
| `format` | [ReadFormat](#readformat) | yes | `text` or `ansi` |
| `text` | string | yes | the content |
| `revision` | integer (uint64) | yes | the output revision of the pane at the time of the read |
| `truncated` | boolean | yes | the output was larger than the requested or allowed size, and herdr cut it |

Checked 2026-10-17 against herdr 0.9.5:

```text
>> pane.read {"pane_id":"w1:p9","source":"visible","lines":5}
<< {"id":"4","result":{"type":"pane_read","read":{"pane_id":"w1:p9","workspace_id":"w1","tab_id":"w1:t3","source":"visible","format":"text","text":"hello world\nsee https://example.com/x\n[user@host repo]$\n","revision":0,"truncated":false}}}
<< {"id":"w","result":{"type":"output_matched","pane_id":"w1:p9","revision":0,"matched_line":"MARKER42","read":{"pane_id":"w1:p9","workspace_id":"w1","tab_id":"w1:t3","source":"recent_unwrapped","format":"text","text":"…\n[user@host repo]$ echo MARKER$((40+2))\nMARKER42\n…","revision":0,"truncated":false}}}
```

With `source` `recent` or `recent_unwrapped`, `truncated` can be `true` when herdr returned all real (non-blank) content. herdr calculates the flag against the full line buffer of the pane, with the blank rows. Thus a small `lines` limit, such as `3` to `12`, on a pane with few non-blank lines can give `truncated: true`. The text is then the same as for a much larger `lines` value. Do not use `truncated: true` alone as proof that herdr cut real content. With `visible` and `detection`, `truncated` was `true` only when herdr cut real content or when `lines` was `0`. Checked 2026-10-17 against herdr 0.9.5.

The values below give `truncated` (T or F), the count of newlines, and the count of characters for each `lines` value. Pane `w1:p9` is 12 rows high with 3 non-blank lines at the top. Pane `w1:p7` is 14 rows high with 1 non-blank line:

```text
w1:p9
visible          1:T/1nl/24c 3:F/3nl/62c 4:F 5:F 12:F 50:F 200:F none:F
recent           1:T/1nl/24c 3:T/3nl/62c 4:T 5:T 12:F 50:F 200:F none:F
recent_unwrapped 1:T/0nl/23c 3:T/2nl/61c 4:T 5:T 12:F 50:F 200:F none:F
detection        1:T/1nl/24c 3:F/3nl/62c 4:F 5:F 12:F 50:F 200:F none:F
w1:p7
visible          1:F/1nl/3c  3:F … none:F
recent           1:T/0nl/0c  3:T/1nl/3c 4:T 5:T 12:T 50:F 200:F none:F
recent_unwrapped 1:T/0nl/0c  3:T/0nl/2c 4:T 5:T 12:T 50:F 200:F none:F
detection        1:F/1nl/3c  3:F … none:F
visible lines:0 on w1:p9 -> {"text":"","truncated":true}
```

`recent` with `lines: 1` on `w1:p7` gave an empty text, because the last row of the buffer was blank.

## ReadSource

Enum. It selects the snapshot that `pane.read` and `pane.wait_for_output` read. The meaning of each value comes from [raw/skill.md](raw/skill.md).

| value | meaning |
| --- | --- |
| `visible` | the viewport that herdr renders now |
| `recent` | the recent rendered output, with soft wraps |
| `recent_unwrapped` | the recent output with joined soft wraps. Use it for logs and transcripts. |
| `detection` | the plain-text snapshot at the bottom of the buffer that herdr uses for agent detection |

Checked 2026-10-17 against herdr 0.9.5. The wire accepted all four values. `recent_unwrapped` joined the soft-wrapped lines and removed the final newline. The wire accepts only the underscore form `recent_unwrapped`. The CLI help gives the hyphen form `recent-unwrapped`. The 0.9.5 CLI also accepts the underscore form. For CLI conventions, see [cli-mapping.md](cli-mapping.md).

```text
>> pane.read {"pane_id":"w1:p9","source":"recent-unwrapped"}
<< {"id":"9","error":{"code":"invalid_request","message":"invalid request: unknown variant `recent-unwrapped`, expected one of `visible`, `recent`, `recent_unwrapped`, `detection` at line 1 column 96"}}
$ herdr pane read --help
      --source <SOURCE>
          Terminal snapshot source (default: recent)
          [possible values: visible, recent, recent-unwrapped, detection]
$ herdr pane read w1:p8 --source recent_unwrapped --lines 2   -> exit 0, output printed
$ herdr pane read w1:p8 --source bogus --lines 2              -> "invalid read source: bogus", exit 2
```

## ReadFormat

Enum. The output encoding of a read.

| value | meaning |
| --- | --- |
| `text` | plain text without styles |
| `ansi` | the text with ANSI color and style escapes. Use it when the styles are evidence. |

Checked 2026-10-17 against herdr 0.9.5. The wire accepted both values, and the result gave the requested `format`. The `ansi` read used `\r\n` line ends and kept the final space after the prompt: `"hello world\r\nsee https://example.com/x\r\n[user@host repo]$ "`. The test text had no colors, so the result had no escape bytes.

## PaneTextPoint

A row and column point in the text content of a pane. The `pane.copy_motion` and `pane.copy_search` results use it (see [api/pane.md](api/pane.md)). These results put their keys directly under `result`, with no wrapper key.

| field | type | required | meaning |
| --- | --- | --- | --- |
| `row` | integer (uint32) | yes | the row index |
| `col` | integer (uint16) | yes | the column index |

Checked 2026-10-17 against herdr 0.9.5: the two fields. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the integer types.

```text
>> pane.copy_motion {"pane_id":"w1:p9","cursor":{"row":0,"col":0},"motion":"next_word_start"}
<< {"id":"11","result":{"type":"pane_copy_motion","pane_id":"w1:p9","cursor":{"row":0,"col":6},"content_revision":20}}
```

## PaneTextRange

A span between two [PaneTextPoint](#panetextpoint) values. The `matches` of the `pane.copy_search` result use it (see [api/pane.md](api/pane.md)). `end` is inclusive.

| field | type | required | meaning |
| --- | --- | --- | --- |
| `start` | [PaneTextPoint](#panetextpoint) | yes | the start of the range |
| `end` | [PaneTextPoint](#panetextpoint) | yes | the end of the range |

The `pane.copy_motion` and `pane.copy_search` results also hold a `content_revision` counter. This counter is not related to [`PaneInfo.revision`](#paneinfo), and it increases much more frequently. On a quiet 0.9.5 pane, `content_revision` was 20 and `PaneInfo.revision` was 1. `pane.copy_search` requires a `content_revision` from the caller that is equal to the current value of the pane. Otherwise the call fails with `stale_content`. Always get a new `content_revision` immediately before a search. The `pane.copy_search` result also holds `matches`, `total`, `current`, and `current_global`. Checked 2026-10-17 against herdr 0.9.5. The word "world" was at columns 6 to 10:

```text
>> pane.copy_search {"pane_id":"w1:p9","query":"world","direction":"forward","cursor":{"row":0,"col":0},"content_revision":20}
<< {"id":"4","result":{"type":"pane_copy_search","pane_id":"w1:p9","content_revision":20,"matches":[{"start":{"row":0,"col":6},"end":{"row":0,"col":10}}],"total":1,"current":0,"current_global":0}}
>> pane.copy_search {…,"content_revision":19}
<< {"id":"5","error":{"code":"stale_content","message":"pane content changed"}}
```

## PaneLinkRegion

The display-cell columns on the current viewport of a pane where a resolved link occurs. The columns are inclusive. `pane.link.resolve` returns it in `regions` (see [api/pane.md](api/pane.md)).

| field | type | required | meaning |
| --- | --- | --- | --- |
| `row` | integer (uint16) | yes | the viewport row |
| `start_col` | integer (uint16) | yes | the first column of the link, inclusive |
| `end_col` | integer (uint16) | yes | the last column of the link, inclusive |

The result field is `row`, but the request parameter for the same value is `viewport_row`. A request with `row` fails with `invalid_request`. Checked 2026-10-17 against herdr 0.9.5. Row 1 held `see https://example.com/x`. The 21-character URL was at columns 4 to 24:

```text
>> pane.link.resolve {"pane_id":"w1:p9","viewport_row":1,"col":8}
<< {"id":"6","result":{"type":"pane_link_resolved","regions":[{"row":1,"start_col":4,"end_col":24}]}}
>> pane.link.resolve {"pane_id":"w1:p9","viewport_row":0,"col":2}
<< {"id":"7","result":{"type":"pane_link_resolved","regions":[]}}
>> pane.link.resolve {"pane_id":"w1:p9","row":1,"col":8}
<< {"id":"8","error":{"code":"invalid_request","message":"invalid request: missing field `viewport_row` at line 1 column 94"}}
```

## Pane operation result shapes

Each `pane.*` method in this section returns one of these result wrappers. Each wrapper embeds a `PaneLayoutSnapshot`, as `layout` or as `source_layout` and `target_layout`. Thus a caller can render the layout again at once. This page collects them because they use the same layout and reason parts. For the method of each result, see [api/pane.md](api/pane.md).

The result is in a key beside `type`. The key name comes from the operation: `edges`, `neighbor`, `focus`, `resize`, `swap`, and `zoom`. `pane.move` uses the key `move_result`, and `pane.process_info` uses `process_info`. Checked 2026-10-17 against herdr 0.9.5.

### PaneEdgesResult

The result of `pane.edges`: for each side of a pane, it tells if the side touches the outer edge of the tab. Required: `pane_id`, `left`, `right`, `up`, `down`, `layout`.

| field | type | meaning |
| --- | --- | --- |
| `pane_id` | string | the pane |
| `left` / `right` / `up` / `down` | boolean | this side touches the outer edge of the tab. Thus there is **no** neighbor on this side. |
| `layout` | [PaneLayoutSnapshot](#panelayoutsnapshot) | the current geometry |

A single pane that fills the tab gives `true` on all four sides. A pane with a neighbor on one side gives `false` on that side and `true` on the other three sides. To find out if a neighbor exists in a direction, use [PaneNeighborResult](#paneneighborresult). The schema gives no description for these fields. Checked 2026-10-17 against herdr 0.9.5:

```text
>> pane.edges {"pane_id":"w3:p1"}   (a single pane)
<< {"id":"1","result":{"type":"pane_edges","edges":{"pane_id":"w3:p1","left":true,"right":true,"up":true,"down":true,"layout":{…}}}}
>> pane.edges {"pane_id":"w1:p2"}   (the right half, with a neighbor on the left)
<< {"id":"2","result":{"type":"pane_edges","edges":{"pane_id":"w1:p2","left":false,"right":true,"up":true,"down":true,"layout":{…}}}}
>> pane.edges {"pane_id":"w1:p1"}   (the top-left pane of three, with neighbors on the right and below)
<< {"id":"3","result":{"type":"pane_edges","edges":{"pane_id":"w1:p1","left":true,"right":false,"up":true,"down":false,"layout":{…}}}}
```

### PaneNeighborResult

The neighbor in a direction (`pane.neighbor`). Required: `pane_id`, `direction`, `layout`.

| field | type | meaning |
| --- | --- | --- |
| `pane_id` | string | the source pane |
| `direction` | [PaneDirection](#panedirection) | the requested direction |
| `neighbor_pane_id` | string \| null | the neighbor pane, or null when there is no neighbor |
| `layout` | [PaneLayoutSnapshot](#panelayoutsnapshot) | the current geometry |

When there is no neighbor, herdr does not send `neighbor_pane_id` as `null`. The key is absent. Checked 2026-10-17 against herdr 0.9.5:

```text
>> pane.neighbor {"pane_id":"w1:p1","direction":"right"}
<< {"id":"4","result":{"type":"pane_neighbor","neighbor":{"pane_id":"w1:p1","direction":"right","neighbor_pane_id":"w1:p2","layout":{…}}}}
>> pane.neighbor {"pane_id":"w1:p1","direction":"left"}
<< {"id":"5","result":{"type":"pane_neighbor","neighbor":{"pane_id":"w1:p1","direction":"left","layout":{…}}}}
```

### PaneFocusDirectionResult

The result of a focus move in a direction (`pane.focus_direction`). Required: `changed`, `source_pane_id`, `layout`.

| field | type | meaning |
| --- | --- | --- |
| `changed` | boolean | the focus moved |
| `source_pane_id` | string | the pane where the focus was before |
| `focused_pane_id` | string \| null | the pane that has the focus now |
| `reason` | [PaneFocusDirectionReason](#reason--auxiliary-enums) \| null | the reason that the focus did not move (`no_neighbor`) |
| `layout` | [PaneLayoutSnapshot](#panelayoutsnapshot) | the current geometry |

As with the other nullable `reason` and `*_id` fields on this page, herdr omits `reason` after a successful move. `focused_pane_id` was present in both cases. Checked 2026-10-17 against herdr 0.9.5:

```text
>> pane.focus_direction {"pane_id":"w1:p1","direction":"left"}
<< {"id":"8","result":{"type":"pane_focus_direction","focus":{"changed":false,"reason":"no_neighbor","source_pane_id":"w1:p1","focused_pane_id":"w1:p1","layout":{…}}}}
>> pane.focus_direction {"pane_id":"w1:p1","direction":"right"}
<< {"id":"9","result":{"type":"pane_focus_direction","focus":{"changed":true,"source_pane_id":"w1:p1","focused_pane_id":"w1:p2","layout":{…}}}}
```

### PaneResizeResult

The result of `pane.resize`. Required: `changed`, `pane_id`, `focused_pane_id`, `layout`.

| field | type | meaning |
| --- | --- | --- |
| `changed` | boolean | the geometry changed |
| `pane_id` | string | the pane that herdr resized |
| `focused_pane_id` | string | the focused pane |
| `reason` | [PaneResizeReason](#reason--auxiliary-enums) \| null | the reason that nothing changed (`unchanged`) |
| `layout` | [PaneLayoutSnapshot](#panelayoutsnapshot) | the current geometry |

Checked 2026-10-17 against herdr 0.9.5:

```text
>> pane.resize {"pane_id":"w2:p1","direction":"right","amount":0.1}    (a single pane)
<< {"id":"2","result":{"type":"pane_resize","resize":{"changed":false,"reason":"unchanged","pane_id":"w2:p1","focused_pane_id":"w2:p1","layout":{…}}}}
>> pane.resize {"pane_id":"w1:p1","direction":"right","amount":0.1}
<< {"id":"11","result":{"type":"pane_resize","resize":{"changed":true,"pane_id":"w1:p1","focused_pane_id":"w1:p1","layout":{…}}}}
```

### PaneSwapResult

The result of a swap of two panes (`pane.swap`). Required: `changed`, `source_pane_id`, `focused_pane_id`, `layout`.

| field | type | meaning |
| --- | --- | --- |
| `changed` | boolean | a swap occurred |
| `source_pane_id` | string | the first pane |
| `target_pane_id` | string \| null | the second pane. herdr gives it also when it did not find the pane (`not_found`). |
| `reason` | [PaneSwapReason](#reason--auxiliary-enums) \| null | the reason that no swap occurred (`no_neighbor`, `same_pane`, `not_found`, or `cross_tab`) |
| `focused_pane_id` | string | the focused pane |
| `layout` | [PaneLayoutSnapshot](#panelayoutsnapshot) | the current geometry |

Checked 2026-10-17 against herdr 0.9.5: all four reasons and a successful swap:

```text
>> pane.swap {"pane_id":"w2:p1","direction":"left"}
<< … "swap":{"changed":false,"reason":"no_neighbor","source_pane_id":"w2:p1","focused_pane_id":"w2:p1",…}
>> pane.swap {"source_pane_id":"w1:p1","target_pane_id":"w1:p1"}
<< … "swap":{"changed":false,"reason":"same_pane","source_pane_id":"w1:p1","target_pane_id":"w1:p1","focused_pane_id":"w1:p3",…}
>> pane.swap {"source_pane_id":"w1:p1","target_pane_id":"w9:p9"}
<< … "swap":{"changed":false,"reason":"not_found","source_pane_id":"w1:p1","target_pane_id":"w9:p9","focused_pane_id":"w1:p3",…}
>> pane.swap {"source_pane_id":"w1:p1","target_pane_id":"w2:p1"}
<< … "swap":{"changed":false,"reason":"cross_tab","source_pane_id":"w1:p1","target_pane_id":"w2:p1","focused_pane_id":"w1:p3",…}
>> pane.swap {"source_pane_id":"w1:p1","target_pane_id":"w1:p2"}
<< … "swap":{"changed":true,"source_pane_id":"w1:p1","target_pane_id":"w1:p2","focused_pane_id":"w1:p1",…}
```

### PaneZoomResult

The result of `pane.zoom`. Required: `changed`, `zoom_changed`, `focus_changed`, `pane_id`, `focused_pane_id`, `zoomed`, `layout`.

| field | type | meaning |
| --- | --- | --- |
| `changed` | boolean | a value changed |
| `zoom_changed` | boolean | the zoom value changed |
| `focus_changed` | boolean | the focus moved as a side effect |
| `pane_id` | string | the target pane |
| `focused_pane_id` | string | the focused pane |
| `zoomed` | boolean | the zoom value after the call |
| `reason` | [PaneZoomReason](#reason--auxiliary-enums) \| null | the reason that the zoom did not change (`single_pane`, `already_zoomed`, or `already_unzoomed`) |
| `layout` | [PaneLayoutSnapshot](#panelayoutsnapshot) | the current geometry |

A zoom of a single pane that did not have the focus gave `single_pane` with `changed: true` and `focus_changed: true`. The zoom moved the focus to the pane, but it could not zoom. Checked 2026-10-17 against herdr 0.9.5: all three reasons:

```text
>> pane.zoom {"pane_id":"w2:p1"}
<< … "zoom":{"changed":true,"zoom_changed":false,"focus_changed":true,"reason":"single_pane","pane_id":"w2:p1","focused_pane_id":"w2:p1","zoomed":false,…}
>> pane.zoom {"pane_id":"w1:p1","mode":"on"}    (second call)
<< … "zoom":{"changed":false,"zoom_changed":false,"focus_changed":false,"reason":"already_zoomed","pane_id":"w1:p1","focused_pane_id":"w1:p1","zoomed":true,…}
>> pane.zoom {"pane_id":"w1:p1","mode":"off"}   (second call)
<< … "zoom":{"changed":false,"zoom_changed":false,"focus_changed":false,"reason":"already_unzoomed","pane_id":"w1:p1","focused_pane_id":"w1:p1","zoomed":false,…}
```

### PaneMoveResult

The result of a move of a pane to a different tab or workspace (`pane.move`). Required: `changed`, `previous_pane_id`, `previous_workspace_id`, `previous_tab_id`, `pane`, `target_layout`, `focused_pane_id`. A move can create or close a tab or a workspace.

| field | type | meaning |
| --- | --- | --- |
| `changed` | boolean | the pane moved |
| `pane` | [PaneInfo](#paneinfo) | the moved pane. Its `pane_id` is new when the move changes the workspace. |
| `previous_pane_id` | string | the pane ID before the move. It continues to resolve for all callers (see [addressing.md](addressing.md#pane-move-re-identification)). |
| `previous_workspace_id` | string | the source workspace |
| `previous_tab_id` | string | the source tab |
| `focused_pane_id` | string | the focused pane after the move |
| `reason` | [PaneMoveReason](#reason--auxiliary-enums) \| null | `same_tab` or `zoomed_tab` |
| `created_tab` | [TabInfo](#tabinfo) \| null | the tab that herdr created for the pane, if any |
| `created_workspace` | [WorkspaceInfo](#workspaceinfo) \| null | the workspace that herdr created for the pane, if any |
| `closed_tab_id` | string \| null | the tab that the move made empty and closed |
| `closed_workspace_id` | string \| null | the workspace that the move made empty and closed |
| `source_layout` | [PaneLayoutSnapshot](#panelayoutsnapshot) \| null | the geometry of the source tab after the move |
| `target_layout` | [PaneLayoutSnapshot](#panelayoutsnapshot) | the geometry of the destination tab |

herdr omits `source_layout` when the move closes the source tab, with or without its workspace. No source tab then remains. A move in one workspace keeps the pane ID. A move to a different workspace gives a new pane ID. A `new_workspace` move also fills `created_tab`. `zoomed_tab` occurred when the source tab was zoomed, for a `new_tab` destination and for a `tab` destination. Checked 2026-10-17 against herdr 0.9.5: both reasons and all created and closed cases.

```text
>> pane.move {"pane_id":"w1:p1","destination":{"type":"tab","tab_id":"w1:t1","split":"right"}}
<< … "move_result":{"changed":false,"reason":"same_tab","previous_pane_id":"w1:p1",…}
>> pane.move {"pane_id":"w1:p3","destination":{"type":"new_tab"}}   (source tab zoomed)
<< … "move_result":{"changed":false,"reason":"zoomed_tab","previous_pane_id":"w1:p3",…}
w1:p3 -> new_tab in w1 : pane.pane_id "w1:p3" (tab w1:t4)
   keys: changed, previous_pane_id, previous_workspace_id, previous_tab_id, pane, source_layout, target_layout, created_tab, focused_pane_id
   "created_tab":{"tab_id":"w1:t4","workspace_id":"w1","number":4,"label":"3","focused":false,"pane_count":1,"agent_status":"unknown"}
w1:p2 -> tab w1:t3     : pane.pane_id "w1:p2"
w2:p1 -> tab w1:t3     : pane.pane_id "w1:pA" (w2 had one pane)
   keys: changed, previous_pane_id, previous_workspace_id, previous_tab_id, pane, target_layout, closed_workspace_id, closed_tab_id, focused_pane_id
   "closed_workspace_id":"w2","closed_tab_id":"w2:t1"
w1:p1 -> new_workspace "reviewer" : pane.pane_id "w3:p1" (w1:t1 had one pane, w1 had other tabs)
   keys: changed, previous_pane_id, previous_workspace_id, previous_tab_id, pane, target_layout, created_workspace, created_tab, closed_tab_id, focused_pane_id
   "created_workspace":{"workspace_id":"w3","number":2,"label":"reviewer",…},"created_tab":{"tab_id":"w3:t1",…},"closed_tab_id":"w1:t1"
>> pane.get {"pane_id":"w1:p1"}   (after the move)
<< {"id":"1","result":{"type":"pane_info","pane":{"pane_id":"w3:p1",…,"workspace_id":"w3","tab_id":"w3:t1",…}}}
```

## WorktreeInfo

A Git worktree that herdr knows. `worktree.list` lists it in `worktrees`. The `worktree.create` and `worktree.open` results return it in `worktree`. The `worktree_created`, `worktree_opened`, and `worktree_removed` events carry it. The `worktree.remove` result does not carry it.

| field | type | required | meaning |
| --- | --- | --- | --- |
| `path` | string | yes | the absolute path of the checkout |
| `label` | string | yes | the display label, which is the repository name |
| `is_bare` | boolean | yes | the repository is bare |
| `is_detached` | boolean | yes | HEAD is detached |
| `is_prunable` | boolean | yes | Git can prune this worktree |
| `is_linked_worktree` | boolean | yes | a linked worktree, not the primary checkout |
| `branch` | string \| null | no | the branch of the checkout. herdr omits it for a detached HEAD or a bare repository. |
| `open_workspace_id` | string \| null | no | the workspace that has this worktree open now, if any |

`label` is the repository name for each entry, also for a linked worktree. It is not the name of the worktree directory. Checked 2026-10-17 against herdr 0.9.5 on these checkouts:

- a primary checkout
- a linked worktree
- a detached-HEAD worktree
- a bare repository with a linked worktree
- a worktree whose directory was deleted (`is_prunable: true`)

After `worktree.open` of `wt-linked`, its entry had `"open_workspace_id":"w5"`. After `workspace.close` of `w5`, the key was absent. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the `worktree.create` result and the `worktree_created` event, because this probe did not run `worktree.create`.

```text
>> worktree.list {"cwd":"/home/user/repo2"}
<< {"id":"1","result":{"type":"worktree_list","source":{"repo_key":"/home/user/repo2/.git","repo_name":"repo2","repo_root":"/home/user/repo2","source_checkout_path":"/home/user/repo2"},"worktrees":[{"path":"/home/user/repo2","branch":"main","is_bare":false,"is_detached":false,"is_prunable":false,"is_linked_worktree":false,"label":"repo2"},{"path":"/home/user/wt-detached","is_bare":false,"is_detached":true,"is_prunable":false,"is_linked_worktree":true,"label":"repo2"},{"path":"/home/user/wt-linked","branch":"feat","is_bare":false,"is_detached":false,"is_prunable":false,"is_linked_worktree":true,"label":"repo2"},{"path":"/home/user/wt-prune","branch":"gone","is_bare":false,"is_detached":false,"is_prunable":true,"is_linked_worktree":true,"label":"repo2"}]}}
>> worktree.list {"cwd":"/home/user/wt-frombare"}
<< {"id":"1","result":{"type":"worktree_list","source":{"repo_key":"/home/user/bare.git","repo_name":"bare.git","repo_root":"/home/user/bare.git","source_checkout_path":"/home/user/bare.git"},"worktrees":[{"path":"/home/user/bare.git","is_bare":true,"is_detached":false,"is_prunable":false,"is_linked_worktree":false,"label":"bare.git"},{"path":"/home/user/wt-frombare","branch":"main","is_bare":false,"is_detached":false,"is_prunable":false,"is_linked_worktree":true,"label":"bare.git"}]}}
worktree_removed event:
{"data":{"forced":false,"type":"worktree_removed","workspace":{…,"workspace_id":"w8","worktree":{…}},"workspace_id":"w8","worktree":{"is_bare":false,"is_detached":true,"is_linked_worktree":true,"is_prunable":false,"label":"repo2","path":"/home/user/wt-ev"}},"event":"worktree_removed"}
worktree.remove result:
{"type":"worktree_removed","workspace_id":"w7","path":"/home/user/wt-detached","forced":false}
```

## WorkspaceWorktreeInfo

The worktree of a workspace (`WorkspaceInfo.worktree`, `PluginInvocationContext.worktree`). It gives the checkout from the view of the workspace.

| field | type | required | meaning |
| --- | --- | --- | --- |
| `repo_key` | string | yes | the canonical identity key of the repository. It is usually the `.git` path. For a bare repository, it is the bare directory. |
| `repo_name` | string | yes | the repository name |
| `repo_root` | string | yes | the root path of the repository |
| `checkout_path` | string | yes | the checkout path of this workspace |
| `is_linked_worktree` | boolean | yes | the checkout is a linked worktree |

Checked 2026-10-17 against herdr 0.9.5. `workspace.get` on a workspace with a worktree gave this shape as `worktree`. In `PluginInvocationContext`, a caller must give all five fields (see [PluginInvocationContext](#plugininvocationcontext)).

```text
>> workspace.get {"workspace_id":"w5"}
<< {"id":"1","result":{"type":"workspace_info","workspace":{"workspace_id":"w5","number":4,"label":"wt-linked","focused":false,"pane_count":1,"tab_count":1,"active_tab_id":"w5:t1","agent_status":"unknown","worktree":{"repo_key":"/home/user/repo2/.git","repo_name":"repo2","repo_root":"/home/user/repo2","checkout_path":"/home/user/wt-linked","is_linked_worktree":true}}}}
```

## WorktreeSourceInfo

The repository that a `worktree.list` result comes from (`worktree_list.source`).

| field | type | required | meaning |
| --- | --- | --- | --- |
| `repo_key` | string | yes | the canonical identity key of the repository |
| `repo_name` | string | yes | the repository name |
| `repo_root` | string | yes | the root path of the repository |
| `source_checkout_path` | string | yes | the primary checkout of the repository. For a bare repository, it is the bare directory. |
| `source_workspace_id` | string \| null | no | the workspace that holds the primary checkout, if any |

`source_checkout_path` is not the checkout that the caller gave. A listing from `wt-linked` and a listing from `wt-detached` both gave `/home/user/repo2`. `source_workspace_id` is absent when no workspace holds the primary checkout. Checked 2026-10-17 against herdr 0.9.5:

```text
>> worktree.list {"cwd":"/home/user/repo2"}   (w4 open on repo2)
<< …"source":{"repo_key":"/home/user/repo2/.git","repo_name":"repo2","repo_root":"/home/user/repo2","source_checkout_path":"/home/user/repo2","source_workspace_id":"w4"},…
>> worktree.list {"cwd":"/home/user/wt-linked"}   (after the close of w4)
<< …"source":{"repo_key":"/home/user/repo2/.git","repo_name":"repo2","repo_root":"/home/user/repo2","source_checkout_path":"/home/user/repo2"},…
```

`worktree.list` from a location outside a Git work tree fails with `not_git_worktree`. The message text depends on the source of the location. For the workspace of the caller, or for a `workspace_id`, the text is "...require a workspace inside a Git work tree". For an explicit `cwd` parameter, the text is "...require a path inside a Git work tree". `worktree.list` has no `path` parameter. Checked 2026-10-17 against herdr 0.9.5:

```text
>> worktree.list {"cwd":"/home/user"}      (not a repository)
<< {"id":"1","error":{"code":"not_git_worktree","message":"Herdr worktree actions require a path inside a Git work tree"}}
>> worktree.list {"workspace_id":"w3"}      (a workspace outside a repository)
<< {"id":"1","error":{"code":"not_git_worktree","message":"Herdr worktree actions require a workspace inside a Git work tree"}}
>> worktree.list {}                         (the active workspace is outside a repository)
<< {"id":"1","error":{"code":"not_git_worktree","message":"Herdr worktree actions require a workspace inside a Git work tree"}}
```

## ServerCapabilities

The feature flags that the server gives in the `pong` result of `ping`. A 0.9.5 scratch server (protocol 22) gave this result. Checked 2026-10-17 against herdr 0.9.5:

```text
>> ping {}
<< {"id":"1","result":{"type":"pong","version":"0.9.5","protocol":22,"capabilities":{"live_handoff":true,"detached_server_daemon":false,"endpoint_protocol_generation":1,"surface_interest":true,"health_check":true,"ssh_agent_registration":true}}}
```

| field | type | required | meaning |
| --- | --- | --- | --- |
| `live_handoff` | boolean | yes | the server can hand off the session live to a different client |
| `detached_server_daemon` | boolean | no (default false) | the server can run as a detached daemon |
| `endpoint_protocol_generation` | integer (uint32) \| null | no | the stable endpoint generation that the client owns and that this server supports |
| `health_check` | boolean | no (default false) | the server supports endpoint health probes |
| `surface_interest` | boolean | no (default false) | the server supports explicit surface interest of the client shell (`client_shell.surface.set`) |
| `ssh_agent_registration` | boolean | no (default false) | new in 0.9.3: the server supports `server.ssh_agent.register` (see [api/server.md](api/server.md)) |

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the required flags, the defaults, and the meaning of each field. `detached_server_daemon` was `true` on the server of the live default session and `false` on the scratch server. A read-only `herdr status --json` gave the capability object of the live server. Checked 2026-10-17 against herdr 0.9.5.

## AgentManifestInfo

An installed agent manifest and its remote update state. `server.agent_manifests` returns it in the `agent_manifest_status` result, and `server.reload_agent_manifests` returns it in the `agent_manifest_reload` result. The `agent_manifest_status` result also holds `last_check_unix` and `last_result` beside `manifests` (see [api/server.md](api/server.md)).

| field | type | required | meaning |
| --- | --- | --- | --- |
| `agent` | string | yes | the agent kind, such as `claude`, `codex`, or `pi` |
| `source` | string | yes | the source path of the manifest, such as `remote:/home/user/.local/state/herdr/agent-detection/remote/claude.toml` |
| `source_kind` | string | yes | the source kind, such as `remote` or `local` |
| `local_override_shadowing_remote` | boolean | yes | a local manifest hides the remote manifest |
| `active_version` | string \| null | no | the version in use now, such as `2026.10.11.1` |
| `cached_remote_version` | string \| null | no | the remote version from the last fetch |
| `remote_update_result` | string \| null | no | the result of the last remote check, such as `current` |
| `remote_update_error` | string \| null | no | the error message of the last failed check |
| `remote_last_checked_unix` | integer (uint64) \| null | no | the Unix time of the last remote check |
| `warning` | string \| null | no | a load warning that is not fatal |

Checked 2026-10-17 against herdr 0.9.5, on a scratch server. There were 22 manifests:

```text
agy amp claude cline codex copilot cursor devin droid gemini grok
hermes kilo kimi kiro letta maki muse opencode pi qodercli qwen
```

Each had `source_kind: "remote"` and `remote_update_result: "current"`. `remote_update_error` and `warning` were absent, not `null`, in this healthy state. This pass did not force them. The 0.9.1 pass found 23 manifests.

```text
>> server.agent_manifests {}
<< result.type "agent_manifest_status", keys: last_check_unix, last_result, manifests
   {"agent":"claude","source":"remote:/home/user/.local/state/herdr/agent-detection/remote/claude.toml","source_kind":"remote","active_version":"2026.10.11.1","cached_remote_version":"2026.10.11.1","local_override_shadowing_remote":false,"remote_update_result":"current","remote_last_checked_unix":1792263341}
```

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: `server.reload_agent_manifests` and its result. This method forces a remote fetch, which writes the shared agent detection cache in the real home directory.

## IntegrationInfo

The integration status of one agent program, from `integration.list` (result `integration_list`, see [api/integration.md](api/integration.md)).

| field | type | required | meaning |
| --- | --- | --- | --- |
| `target` | [IntegrationTarget](#integrationtarget) | yes | the agent program of this entry |
| `label` | string | yes | the display label |
| `command` | string | yes | the command of this agent program that herdr looks for |
| `available` | boolean | yes | the command of the agent program is available on this machine |
| `state` | [IntegrationState](#integrationstate) | yes | the current install state of the integration |

herdr does not always derive `command` from `target` or `label`. The command of `cursor` is `cursor-agent`. The label of `antigravity_cli` is `antigravity-cli`, and its command is `agy`. Checked 2026-10-17 against herdr 0.9.5. All 17 entries had exactly these five fields:

```text
>> integration.list {}
<< {"target":"claude","label":"claude","command":"claude","available":true,"state":"current"}
   {"target":"cursor","label":"cursor","command":"cursor-agent","available":true,"state":"current"}
   {"target":"antigravity_cli","label":"antigravity-cli","command":"agy","available":false,"state":"not_installed"}
```

## IntegrationState

Enum. `IntegrationInfo.state`.

| value | meaning |
| --- | --- |
| `not_installed` | the integration is not installed |
| `current` | the integration is installed and up to date |
| `outdated` | the installed file has an older `HERDR_INTEGRATION_VERSION` marker than the running herdr installs |

Checked 2026-10-17 against herdr 0.9.5: all three values. The check used a scratch server with a private home directory. `integration.install` of `claude` wrote a hook file with the marker `# HERDR_INTEGRATION_VERSION=11`. A change of the marker to `10` made `integration.list` give `outdated`. `integration.uninstall` then gave `not_installed`.

```text
>> integration.install {"target":"claude"}
<< {"id":"1","result":{"type":"integration_install","target":"claude","details":{"messages":["installed claude integration hook to /home/user/.claude/hooks/herdr-agent-state.sh","ensured claude settings at /home/user/.claude/settings.json"]}}}
{"target":"claude","label":"claude","command":"claude","available":true,"state":"current"}
(marker HERDR_INTEGRATION_VERSION=11 changed to 10)
{"target":"claude","label":"claude","command":"claude","available":true,"state":"outdated"}
>> integration.uninstall {"target":"claude"}
<< {"id":"1","result":{"type":"integration_uninstall","target":"claude","details":{"messages":["removed claude hook at /home/user/.claude/hooks/herdr-agent-state.sh","removed herdr claude hook entries from /home/user/.claude/settings.json"]}}}
{"target":"claude","label":"claude","command":"claude","available":true,"state":"not_installed"}
```

## IntegrationTarget

Enum. The agent program that `integration.install` and `integration.uninstall` target. The full value set:

`pi`, `omp`, `claude`, `codex`, `copilot`, `devin`, `droid`, `kimi`, `opencode`, `kilo`, `hermes`, `qodercli`, `qwen`, `cursor`, `mastracode`, `antigravity_cli`, `grok`.

Checked 2026-10-17 against herdr 0.9.5. `integration.list` gave exactly these 17 values in this order.

## ConfigReloadStatus

Enum. The result of `server.reload_config` (result `config_reload`, with a `diagnostics` array).

| value | meaning |
| --- | --- |
| `applied` | herdr applied all of the new config |
| `partial` | herdr applied the config but refused some settings (see `diagnostics`) |
| `failed` | the reload failed, and herdr kept the earlier config |

Checked 2026-10-17 against herdr 0.9.5: all three values. `server.reload_config` reads the shared config file, not a file in the session directory (see [environment.md](environment.md)). The check used a scratch server with a private config directory. `partial` occurred for incorrect values, unknown keys, and unknown theme names:

```text
(no config file)                    -> {"type":"config_reload","status":"applied","diagnostics":[]}
"this is = = not toml"              -> {"type":"config_reload","status":"failed","diagnostics":["config parse error: TOML parse error at line 1, column 6\n  |\n1 | this is = = not toml\n  |      ^\nexpected `.`, `=`\n; keeping current config"]}
[terminal] shell_mode = "bogus"     -> {"status":"partial","diagnostics":["invalid terminal config: unknown variant `bogus`, expected one of `auto`, `login`, `non_login`\nin `shell_mode`\n; keeping current terminal settings"]}
[theme] name = "nonexistent-theme"  -> {"status":"partial","diagnostics":["unknown theme name theme.name = \"nonexistent-theme\"; using \"catppuccin\"; valid themes: catppuccin, …"]}
[keys] prefix = "notakey+++"        -> {"status":"partial","diagnostics":["invalid keybinding: keys.prefix; using fallback ctrl+b; kept current keybinds","invalid keybinding: keys.prefix = \"notakey+++\"; ignoring prefix; kept current keybinds"]}
[ui] bogus_field = 1                -> {"status":"partial","diagnostics":["unknown config key ui.bogus_field; ignoring key"]}
[server] headless_cols = 100        -> {"status":"applied","diagnostics":[]}
```

## PopupSize

A size value for a popup (`PluginManifestPane.width` and `height`). It has two forms:

| form | schema | meaning |
| --- | --- | --- |
| integer | `uint16` (0 to 65535) | the outer size in terminal cells, with the border |
| percentage string | matches `^(100\|[1-9][0-9]?)%$` | the outer size as a percent of the terminal area, such as `"80%"` |

herdr also accepts a leading zero, such as `"05%"`, and changes it to `"5%"`. The schema pattern does not accept `"05%"`. herdr accepts a cell count of `0`. herdr applies the same rules in the manifest and on the wire (`plugin.pane.open` with the placement `popup`). Checked 2026-10-17 against herdr 0.9.5, with a plugin linked on a scratch server with a private config directory:

```text
manifest width -> result after plugin.link
"80%" (height 12) -> "width":"80%","height":12
"100%"  -> "width":"100%"
"05%"   -> "width":"5%"
0       -> "width":0
65535   -> "width":65535
"0%"    -> plugin_manifest_parse_failed "percentage must be between 1% and 100%"
"101%"  -> plugin_manifest_parse_failed "percentage must be between 1% and 100%"
70000   -> plugin_manifest_parse_failed "cell count must be between 0 and 65535"
"50"    -> plugin_manifest_parse_failed "string sizes must be percentages like 80%; use a number for cells"
plugin.pane.open size -> response
"05%" -> {"type":"ok"}     0 -> {"type":"ok"}
"0%"  -> invalid_request "invalid request: percentage must be between 1% and 100% …"
"101%"-> invalid_request "invalid request: percentage must be between 1% and 100% …"
70000 -> invalid_request "invalid request: cell count must fit in u16 …"
"50"  -> invalid_request "invalid request: string sizes must be percentages like 80%; use a number for cells …"
```

## Plugin entities

herdr plugins add actions, panes, event hooks, and link handlers. A manifest that declared all four gave all four in the `plugin_linked` result. Checked 2026-10-17 against herdr 0.9.5. For the plugin methods, see [api/plugin.md](api/plugin.md).

On a new scratch server, `plugin.list`, `plugin.action.list`, and `plugin.log.list` gave empty arrays. `plugin.action.invoke` and `plugin.pane.open` with an unknown `plugin_id` failed with `plugin_not_found`. Checked 2026-10-17 against herdr 0.9.5:

```json
{"id":"1","method":"plugin.list","params":{}}
{"id":"1","result":{"type":"plugin_list","plugins":[]}}
{"id":"2","method":"plugin.action.list","params":{}}
{"id":"2","result":{"type":"plugin_action_list","actions":[]}}
{"id":"3","method":"plugin.log.list","params":{}}
{"id":"3","result":{"type":"plugin_log_list","logs":[]}}
```

`plugin.link`, `plugin.enable`, and `plugin.disable` write the shared plugin config. For the checks on this page, a scratch server ran with a private home directory and private XDG directories. Then `plugin.link` wrote only in that private root. The plugin `$defs` in `raw/schema.json` have no change since 0.9.1. The example plugin `acme.tools` declared a description, platforms, actions, panes, startup and build commands, event hooks, and a link handler. A second plugin `acme.hooks` tested the event names, the platforms of one item, and the source fields.

### InstalledPluginInfo

A registered plugin (`plugin.list`, `plugin.link`, `plugin.enable`, `plugin.disable`). Required: `plugin_id`, `name`, `version`, `manifest_path`, `plugin_root`, `enabled`.

| field | type | meaning |
| --- | --- | --- |
| `plugin_id` | string | the stable plugin ID |
| `name` | string | the plugin name |
| `version` | string | the plugin version |
| `manifest_path` | string | the canonical path of the plugin manifest file |
| `plugin_root` | string | the canonical root directory of the plugin |
| `enabled` | boolean | the plugin is enabled |
| `description` | string \| null | a description for people |
| `min_herdr_version` | string (default "") | the minimum herdr version that the plugin needs |
| `source` | [PluginSourceInfo](#pluginsourceinfo) (default `{kind:"local"}`) | the source of the plugin |
| `platforms` | array of [PluginPlatform](#plugin-enums) \| null | the supported platforms |
| `actions` | array of [PluginManifestAction](#pluginmanifestaction) | the actions of the plugin |
| `panes` | array of `PluginManifestPane` | the panes of the plugin |
| `startup` | array of `PluginManifestStartup` | the startup commands |
| `build` | array of `PluginManifestBuild` | the build commands |
| `events` | array of `PluginManifestEventHook` | the event hooks |
| `link_handlers` | array of `PluginManifestLinkHandler` | the URL and link handlers |
| `warnings` | array of `string` | link and load warnings that are not fatal. herdr keeps them and gives them again in `plugin.list`. |

herdr gives `manifest_path` and `plugin_root` as canonical paths. A `path` through a symlink came back resolved. A collection that the manifest does not declare is absent. herdr sorts the `panes` by `id` and the `events` by `on`. A `warnings` entry occurred for an unknown event, such as `"unknown event 'no_such_event'"`. A manifest without `platforms` gave the warning `"manifest does not declare platforms; platform support unknown"`. Checked 2026-10-17 against herdr 0.9.5: each field.

```json
{"id":"1","method":"plugin.link","params":{"path":"/home/user/project/acme"}}
{"id":"1","result":{"type":"plugin_linked","plugin":{"plugin_id":"acme.tools","name":"Acme Tools","version":"1.0.0","min_herdr_version":"0.9.0","description":"Probe plugin","manifest_path":"/home/user/project/acme/herdr-plugin.toml","plugin_root":"/home/user/project/acme","enabled":true,"platforms":["linux"],"build":[{"command":["sh","-c","touch built.marker"]}],"startup":[{"command":["sh","-c","echo startup"]}],"actions":[{"id":"format-buffer","title":"Format buffer","description":"Formats text","contexts":["selection","pane"],"command":["sh","-c","echo out; echo err >&2; exit 3"]},{"id":"ok","title":"OK","command":["true"]}],"events":[{"on":"no_such_event","command":["true"]},{"on":"pane.focused","command":["sh","-c","echo hook"]}],"panes":[{"id":"dashboard","title":"Dashboard","description":"A pane","placement":"popup","width":"80%","height":20,"command":["sh","-c","echo dash; exec sleep 600"]},{"id":"plain","title":"Plain","placement":"overlay","command":["sh","-c","exec sleep 600"]},{"id":"side","title":"Side","placement":"split","command":["sh","-c","exec sleep 600"]}],"link_handlers":[{"id":"jira","title":"Open ticket","pattern":"https://example.com/(.*)","action":"ok"}],"source":{"kind":"local"},"warnings":["unknown event 'no_such_event'"]}}}
```

A part of the manifest (`herdr-plugin.toml`) for this result:

```toml
id = "acme.tools"
name = "Acme Tools"
version = "1.0.0"
min_herdr_version = "0.9.0"
description = "Probe plugin"
platforms = ["linux"]

[[actions]]
id = "format-buffer"
title = "Format buffer"
description = "Formats text"
contexts = ["selection", "pane"]
command = ["sh", "-c", "echo out; echo err >&2; exit 3"]

[[panes]]
id = "dashboard"
title = "Dashboard"
placement = "popup"
width = "80%"
height = 20
command = ["sh", "-c", "echo dash; exec sleep 600"]

[[startup]]
command = ["sh", "-c", "echo startup"]

[[build]]
command = ["sh", "-c", "touch built.marker"]

[[events]]
on = "pane.focused"
command = ["sh", "-c", "echo hook"]

[[link_handlers]]
id = "jira"
title = "Open ticket"
pattern = "https://example.com/(.*)"
action = "ok"
```

### PluginManifestAction

An action that a manifest declares. Required: `id`, `title`, `command`.

| field | type | meaning |
| --- | --- | --- |
| `id` | string | the action ID |
| `title` | string | the display title |
| `command` | array of `string` | the command and its arguments |
| `contexts` | array of [PluginActionContext](#plugin-enums) | the contexts in which herdr offers the action |
| `description` | string \| null | the description |
| `platforms` | array of [PluginPlatform](#plugin-enums) \| null | the supported platforms |

Checked 2026-10-17 against herdr 0.9.5: each field (see the capture in [InstalledPluginInfo](#installedplugininfo)). The `platforms` of one item override the top-level `platforms` of the plugin. This entry gives only the declared `platforms`. [PluginActionInfo](#pluginactioninfo) gives the effective list.

### PluginActionInfo

A resolved action that a caller can invoke (`plugin.action.list`, `plugin_action_invoked`). Required: `plugin_id`, `action_id`, `title`, `command`.

| field | type | meaning |
| --- | --- | --- |
| `plugin_id` | string | the plugin of the action |
| `action_id` | string | the action ID |
| `title` | string | the display title |
| `command` | array of `string` | the command and its arguments |
| `contexts` | array of [PluginActionContext](#plugin-enums) | the contexts in which herdr offers the action |
| `description` | string \| null | the description |
| `platforms` | array of [PluginPlatform](#plugin-enums) \| null | the effective platforms. When the action declares none, it receives the top-level `platforms` of the plugin. |

Checked 2026-10-17 against herdr 0.9.5:

```json
{"id":"2","method":"plugin.action.list","params":{}}
{"id":"2","result":{"type":"plugin_action_list","actions":[{"plugin_id":"acme.tools","action_id":"format-buffer","title":"Format buffer","description":"Formats text","contexts":["selection","pane"],"command":["sh","-c","echo out; echo err >&2; exit 3"],"platforms":["linux"]},{"plugin_id":"acme.tools","action_id":"ok","title":"OK","command":["true"],"platforms":["linux"]}]}}
```

### PluginManifestBuild / PluginManifestStartup

Each has `command` (array of `string`, required) and an optional `platforms` (array of [PluginPlatform](#plugin-enums) \| null).

The manifest declares build commands for the install or build step. `plugin.link` does not run them. The build commands did not run at link, at enable, or at server boot. `plugin.log.list` gave no build entry. This pass did not check `herdr plugin install`, because it needs the network.

Startup commands run one time for each enabled plugin at server boot. The log gives them with `event: "startup"`. A startup entry whose `platforms` does not include the host does not run, such as `["windows"]` on Linux. Checked 2026-10-17 against herdr 0.9.5, after a restart:

```json
{"id":"1","method":"plugin.log.list","params":{}}
{"id":"1","result":{"type":"plugin_log_list","logs":[{"log_id":"plugin-log-1","plugin_id":"acme.tools","event":"startup","command":["sh","-c","echo startup"],"status":"succeeded","started_unix_ms":1792263615080,"finished_unix_ms":1792263615080,"exit_code":0,"stdout":"startup\n","stderr":""}]}}
```

### PluginManifestEventHook

Runs a command on a herdr event. Required: `on`, `command`.

| field | type | meaning |
| --- | --- | --- |
| `on` | string | the event name. It uses the subscription name, such as `pane.focused` (see [events.md](events.md#subscription-events-subscriptioneventkind--subscriptioneventdata)). An unknown name gives a warning. |
| `command` | array of `string` | the command and its arguments |
| `platforms` | array of [PluginPlatform](#plugin-enums) \| null | the supported platforms |

herdr checks `on` against the subscription names, such as `pane.focused` or `workspace.created`. herdr does not check it against the [EventKind](#eventkind) names. The EventKind name `pane_focused` gave the warning `unknown event 'pane_focused'`. A hook on `pane.focused` ran on each pane focus. Its log entry had `"event":"pane.focused"`. Checked 2026-10-17 against herdr 0.9.5:

```json
{"id":"1","method":"plugin.link","params":{"path":"/home/user/project/hooks","enabled":false}}
{"id":"1","result":{"type":"plugin_linked","plugin":{"plugin_id":"acme.hooks",…,"events":[{"on":"pane_focused","command":["true"]},{"on":"workspace.created","platforms":["linux"],"command":["true"]}],…,"warnings":["unknown event 'pane_focused'","manifest does not declare platforms; platform support unknown"]}}}
```

### PluginManifestLinkHandler

Handles matching URLs and links. Required: `id`, `title`, `pattern`, `action`.

| field | type | meaning |
| --- | --- | --- |
| `id` | string | the handler ID |
| `title` | string | the display title |
| `pattern` | string | the match pattern |
| `action` | string | the action ID that herdr invokes on a match |
| `platforms` | array of [PluginPlatform](#plugin-enums) \| null | the supported platforms |

Checked 2026-10-17 against herdr 0.9.5: `id`, `title`, `pattern`, and `action` as a stored shape (see the capture in [InstalledPluginInfo](#installedplugininfo)). `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: `platforms`, and the start of a handler. A start needs a TUI client with mouse capture and a click on a URL.

### PluginManifestPane

Declares a pane that the plugin gives. Required: `id`, `title`, `command`.

| field | type | meaning |
| --- | --- | --- |
| `id` | string | the pane ID |
| `title` | string | the display title |
| `command` | array of `string` | the command and its arguments |
| `description` | string \| null | the description |
| `placement` | [PluginPanePlacement](#plugin-enums) (default `overlay`) | the position of the pane |
| `width` | [PopupSize](#popupsize) \| null | the pane width. It is allowed only with the placement `popup`. |
| `height` | [PopupSize](#popupsize) \| null | the pane height. It is allowed only with the placement `popup`. |
| `platforms` | array of [PluginPlatform](#plugin-enums) \| null | the supported platforms |

A pane without `placement` came back with `"placement":"overlay"`. With `placement = "popup"`, `"80%"` and `20` came back as `"width":"80%","height":20`. With a different placement, `width` or `height` makes `plugin.link` fail. [errors.md](errors.md) owns the error code. Checked 2026-10-17 against herdr 0.9.5. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: `platforms` on a pane.

```json
{"code":"invalid_plugin_pane_size","message":"pane width and height are only supported when placement is popup"}
```

### PluginPaneInfo

A live plugin pane (`plugin_pane_opened`, `plugin_pane_focused`). Required: `plugin_id`, `entrypoint`, `pane`.

| field | type | meaning |
| --- | --- | --- |
| `plugin_id` | string | the plugin of the pane |
| `entrypoint` | string | the ID of the pane in the manifest |
| `pane` | [PaneInfo](#paneinfo) | the pane |

`pane` is a full `PaneInfo`. Its `label` is the `title` from the manifest. Its `cwd` is the `plugin_root` by default. `plugin.pane.focus` gave the same shape with `"type":"plugin_pane_focused"`. A `popup` placement gave `{"type":"ok"}` with no `PluginPaneInfo`. Checked 2026-10-17 against herdr 0.9.5:

```json
{"id":"5","method":"plugin.pane.open","params":{"plugin_id":"acme.tools","entrypoint":"side","focus":true}}
{"id":"5","result":{"type":"plugin_pane_opened","plugin_pane":{"plugin_id":"acme.tools","entrypoint":"side","pane":{"pane_id":"w1:p2","terminal_id":"term_65e0dd49456182","workspace_id":"w1","tab_id":"w1:t1","focused":true,"cwd":"/home/user/project/acme","foreground_cwd":"/home/user/project/acme","label":"Side","agent_status":"unknown","scroll":{"offset_from_bottom":0,"max_offset_from_bottom":0,"viewport_rows":38},"revision":0}}}}
```

### PluginCommandLogInfo

A record of a plugin command run (`plugin.log.list`, `plugin_action_invoked.log`). Required: `log_id`, `plugin_id`, `command`, `status`, `started_unix_ms`.

| field | type | meaning |
| --- | --- | --- |
| `log_id` | string | the log entry ID, such as `plugin-log-1` |
| `plugin_id` | string | the plugin of the command |
| `command` | array of `string` | the command and its arguments that herdr ran |
| `status` | [PluginCommandStatus](#plugin-enums) | `running`, `succeeded`, or `failed` |
| `started_unix_ms` | integer (uint64) | the start time (Unix ms) |
| `finished_unix_ms` | integer (uint64) \| null | the end time (Unix ms) |
| `action_id` | string \| null | the action that started the command, if any |
| `event` | string \| null | the event that started the command, if any |
| `exit_code` | integer (int32) \| null | the exit code of the process |
| `error` | string \| null | the error message |
| `stdout` | string \| null | the stdout of the command |
| `stderr` | string \| null | the stderr of the command |

A nonzero exit gives `failed` with `exit_code`. A command that cannot start gives `failed` with `error` and no `exit_code`. In `plugin_action_invoked.log`, only the required fields and `action_id` occur, with `status: "running"`. Checked 2026-10-17 against herdr 0.9.5: each field.

```json
{"log_id":"plugin-log-1","plugin_id":"acme.tools","action_id":"format-buffer","command":["sh","-c","echo out; echo err >&2; exit 3"],"status":"failed","started_unix_ms":1792263345953,"finished_unix_ms":1792263345953,"exit_code":3,"stdout":"out\n","stderr":"err\n"}
{"log_id":"plugin-log-3","plugin_id":"acme.tools","event":"pane.focused","command":["sh","-c","echo hook"],"status":"succeeded","started_unix_ms":1792263353738,"finished_unix_ms":1792263353739,"exit_code":0,"stdout":"hook\n","stderr":""}
{"log_id":"plugin-log-6","plugin_id":"acme.tools","action_id":"missing","command":["no-such-cmd-xyz"],"status":"failed","started_unix_ms":1792263364144,"finished_unix_ms":1792263364144,"stdout":"","stderr":"","error":"No such file or directory (os error 2)"}
```

### PluginInvocationContext

The context of a plugin action invocation (`plugin_action_invoked.context`). All fields are optional and can be null. The object records the focus and the selection at the time of the invocation.

| field | type | meaning |
| --- | --- | --- |
| `invocation_source` | string \| null | the source of the invocation |
| `correlation_id` | string \| null | the correlation ID |
| `link_handler_id` | string \| null | the link handler that ran |
| `clicked_url` | string \| null | the URL that a person clicked |
| `selected_text` | string \| null | the selected text |
| `workspace_id` / `workspace_label` / `workspace_cwd` | string \| null | the context of the focused workspace |
| `tab_id` / `tab_label` | string \| null | the context of the focused tab |
| `focused_pane_id` / `focused_pane_cwd` / `focused_pane_agent` | string \| null | the context of the focused pane |
| `focused_pane_status` | [AgentStatus](#agentstatus) \| null | the agent status of the focused pane |
| `worktree` | [WorkspaceWorktreeInfo](#workspaceworktreeinfo) \| null | the worktree of the focused workspace |

herdr gives back the values of the caller. When the caller omits them, herdr sets `invocation_source` to `"api"` and `correlation_id` to the request ID. When a workspace exists, herdr also fills the workspace, tab, and focused pane fields from the active workspace for each field that the caller omits. A value from the caller wins, such as `tab_label`. When the server has no workspace, herdr fills only the source and the correlation ID. `workspace_cwd` gave the `cwd` of the focused pane, not the creation `cwd` of the workspace. A `worktree` value needs all five fields. Otherwise the request fails with `invalid_request: missing field 'repo_key'`. [api/plugin.md](api/plugin.md) owns this behavior. Checked 2026-10-17 against herdr 0.9.5: each field.

```json
{"id":"1","method":"plugin.action.invoke","params":{"action_id":"ok","plugin_id":"acme.tools","context":{"worktree":{"repo_key":"k","repo_name":"project","repo_root":"/home/user/project","checkout_path":"/home/user/project","is_linked_worktree":false},"focused_pane_status":"working","tab_label":"build"}}}
{"id":"1","result":{"type":"plugin_action_invoked","action":{"plugin_id":"acme.tools","action_id":"ok","title":"OK","command":["true"],"platforms":["linux"]},"context":{"workspace_id":"w1","workspace_label":"api-docs","workspace_cwd":"/home/user/project/acme","worktree":{"repo_key":"k","repo_name":"project","repo_root":"/home/user/project","checkout_path":"/home/user/project","is_linked_worktree":false},"tab_id":"w1:t1","tab_label":"build","focused_pane_id":"w1:p2","focused_pane_cwd":"/home/user/project/acme","focused_pane_status":"working","invocation_source":"api","correlation_id":"1"},"log":{"log_id":"plugin-log-6","plugin_id":"acme.tools","action_id":"ok","command":["true"],"status":"running","started_unix_ms":1792263364144}}}
```

### PluginSourceInfo

The source of an installed plugin (`InstalledPluginInfo.source`). No field is required.

| field | type | meaning |
| --- | --- | --- |
| `kind` | [PluginSourceKind](#plugin-enums) (default `local`) | `local` or `github` |
| `owner` / `repo` / `subdir` | string \| null | the GitHub coordinates |
| `requested_ref` | string \| null | the requested Git ref |
| `resolved_commit` | string \| null | the resolved commit |
| `managed_path` | string \| null | the managed install path |
| `installed_unix_ms` | integer (uint64) \| null | the install time (Unix ms) |

A `source` without `kind` gave `"kind":"local"`. A local link gave back the other fields without change. An explicit `"managed_path":null` was absent from the result. Checked 2026-10-17 against herdr 0.9.5:

```json
"source":{"kind":"local","owner":"acme","repo":"r","subdir":"s","requested_ref":"v1","resolved_commit":"abc","installed_unix_ms":5}
```

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: a real GitHub install, in which herdr sets `managed_path`. That install needs a network fetch.

### Plugin enums

- **PluginActionContext**: `global`, `workspace`, `tab`, `pane`, `selection`.
- **PluginPanePlacement**: `overlay`, `popup`, `split`, `tab`, `zoomed`.
- **PluginPlatform**: `linux`, `macos`, `windows`.
- **PluginCommandStatus**: `running`, `succeeded`, `failed`.
- **PluginSourceKind**: `local`, `github`.

Checked 2026-10-17 against herdr 0.9.5. The pass observed all three `PluginCommandStatus` values, and herdr accepted all three platforms. An unknown value failed with the full value list:

```text
contexts = ["bogus"]   -> plugin_manifest_parse_failed "unknown variant `bogus`, expected one of `global`, `workspace`, `tab`, `pane`, `selection`"
placement = "bogus"    -> plugin_manifest_parse_failed "… expected one of `overlay`, `popup`, `split`, `tab`, `zoomed`"
platforms = ["bsd"]    -> plugin_manifest_parse_failed "invalid_plugin_platform: unknown platform 'bsd'"
source {"kind":"gitlab"} -> invalid_request "unknown variant `gitlab`, expected `local` or `github`"
```

## Event entities

The server pushes events to a connection after `events.subscribe`, as `{"event": …, "data": …}` lines (see [protocol.md](protocol.md)). The `wait_matched` result of `events.wait` also embeds `EventEnvelope`. [events.md](events.md#event-catalog-eventkind--eventdata) owns the event catalog. Checked 2026-10-17 against herdr 0.9.5.

### EventEnvelope

| field | type | required | meaning |
| --- | --- | --- | --- |
| `event` | [EventKind](#eventkind) | yes | the event name |
| `data` | EventData | yes | the event payload, a tagged union whose `type` is equal to `event` |

Checked 2026-10-17 against herdr 0.9.5. One connection that subscribed to the 24 plain types received 63 pushes. Each line had exactly the keys `event` and `data`, and `data.type` was equal to `event`. A `pane_agent_status_changed` match gave this result:

```json
{"id":"6","method":"events.wait","params":{"match_event":{"event":"pane_agent_status_changed","pane_id":"w1:p1","agent_status":"working"},"timeout_ms":5000}}
{"id":"6","result":{"type":"wait_matched","event":{"event":"pane_agent_status_changed","data":{"type":"pane_agent_status_changed","pane_id":"w1:p1","workspace_id":"w1","agent_status":"working","agent":"probe"}}}}
```

`events.wait` requires a `match_event` field, also for a wait on any next event. A request with only `timeout_ms` fails with `invalid_request` (`missing field 'match_event'`). On 0.9.5, `events.wait` supports only agent status matches. A different match fails with `unsupported_event_wait_match`, and a wait without a match ends with `timeout`. See [events.md](events.md#eventswait). Checked 2026-10-17 against herdr 0.9.5.

### EventKind

Enum of all event names:

```text
workspace_created, workspace_updated, workspace_metadata_updated, workspace_closed,
workspace_renamed, workspace_moved, workspace_reordered, workspace_focused,
worktree_created, worktree_opened, worktree_removed,
tab_created, tab_closed, tab_renamed, tab_moved, tab_focused,
pane_created, pane_closed, pane_updated, pane_focused, pane_moved, pane_output_changed,
pane_exited, pane_agent_detected, pane_agent_status_changed, layout_updated
```

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the list of 26 names, which has no change since 0.9.1. Checked 2026-10-17 against herdr 0.9.5: the pass observed 25 of the 26 kinds on the wire. Only `pane_output_changed` did not occur. `pane_agent_status_changed` came through `events.wait`, and the other 24 kinds came through a plain subscription. `workspace_updated` came from `worktree.open` of a worktree that was already open. For the notes on each kind, see [events.md](events.md#event-catalog-eventkind--eventdata).

### EventData

A tagged union with the discriminator `type`, equal to the [EventKind](#eventkind). It holds the payload of each event. Payloads embed the domain entities above:

- `workspace_created` holds `workspace` ([WorkspaceInfo](#workspaceinfo)).
- `pane_created` holds `pane` ([PaneInfo](#paneinfo)).
- `pane_agent_status_changed` holds `pane_id`, `workspace_id`, and `agent_status` ([AgentStatus](#agentstatus)), and the optional `agent`, `display_agent`, `title`, and `state_labels`.
- `layout_updated` holds `layout` ([PaneLayoutSnapshot](#panelayoutsnapshot)).

For the payload tables of each event, see [events.md](events.md#event-catalog-eventkind--eventdata).

Checked 2026-10-17 against herdr 0.9.5: the keys of `workspace_created`, `pane_created`, and `layout_updated`, and `pane_id`, `workspace_id`, `agent_status`, and `agent` of `pane_agent_status_changed`. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: `display_agent`, `title`, and `state_labels` in `pane_agent_status_changed`. A real `claude` with a private home directory stopped at its first-run screens and gave only `agent`. A signed-in agent needs the real agent config of a person, which this pass did not use.

## Reason & auxiliary enums

Small enums that give the reason for the result of an operation. Each enum is the nullable `reason` field of the result with the same name above.

| enum | values | used by |
| --- | --- | --- |
| `PaneFocusDirectionReason` | `no_neighbor` | `PaneFocusDirectionResult` |
| `PaneResizeReason` | `unchanged` | `PaneResizeResult` |
| `PaneSwapReason` | `no_neighbor`, `same_pane`, `not_found`, `cross_tab` | `PaneSwapResult` |
| `PaneZoomReason` | `single_pane`, `already_zoomed`, `already_unzoomed` | `PaneZoomResult` |
| `PaneMoveReason` | `same_tab`, `zoomed_tab` | `PaneMoveResult` |
| `NotificationShowReason` | `shown`, `disabled`, `rate_limited`, `no_foreground_client`, `busy` | the `notification.show` result. A scratch server gave `disabled`. |
| `ClientWindowTitleReason` | `set`, `cleared`, `no_foreground_client` | the `client.window_title.set` result. A scratch server gave `no_foreground_client`. |

Checked 2026-10-17 against herdr 0.9.5, on a scratch server with no TUI client. The pass observed `no_neighbor` (focus), `unchanged`, all four `PaneSwapReason` values, all three `PaneZoomReason` values, `same_tab`, `zoomed_tab`, `disabled`, and `no_foreground_client`:

```text
notification.show {title}  -> {"type":"notification_show","shown":false,"reason":"disabled"}
client.window_title.set    -> {"type":"client_window_title","changed":false,"reason":"no_foreground_client"}
client.window_title.clear  -> {"type":"client_window_title","changed":false,"reason":"no_foreground_client"}
```

The 0.9.1 pass recorded `no_foreground_client` for `notification.show`. On the 0.9.3 and 0.9.5 scratch servers, it gave `disabled`. `Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: `shown`, `rate_limited`, and `busy` (`NotificationShowReason`), and `set` and `cleared` (`ClientWindowTitleReason`). They need a real foreground client that is attached to the session.

---

## Appendix: on-disk persistence (session.json, version 3)

herdr saves the layout topology in a `session.json` file, so that it can restore a session. The file of the default session is `~/.config/herdr/session.json`. A named session uses `~/.config/herdr/sessions/<name>/session.json`. For the session directories and the effect of `XDG_CONFIG_HOME`, see [environment.md](environment.md#session-directory-layout).

This shape on the disk is **internal and different from the API entities above**. It uses internal integer pane numbers, not public IDs such as `w1:p1`. Its layout variants `Split` and `Pane` start with a capital letter, but the API uses `split` and `pane`. This page gives it only to relate the two forms. The socket API never returns it.

Top-level object (`version: 3`):

| field | type | meaning |
| --- | --- | --- |
| `version` | integer | the schema version (now `3`) |
| `workspaces` | array | the saved workspaces (see below) |
| `active` | integer | the index of the active workspace |
| `selected` | integer | the index of the selected workspace |
| `sidebar_width` | integer \| null | the width of the UI sidebar |
| `sidebar_section_split` | number \| null | the split ratio of the UI sidebar sections |
| `collapsed_space_keys` | array | the collapsed sidebar groups |

Checked 2026-10-17 against herdr 0.9.5. The file of a scratch session and the live default file both had exactly these seven top-level keys and `version: 3`. `collapsed_space_keys` was `[]`. `sidebar_width` and `sidebar_section_split` can be `null`. They were `null` on a scratch session with no client and on a live session with attached clients.

Each **workspace** entry has these keys:

- `id` (string)
- `custom_name` (string \| null)
- `identity_cwd` (string)
- `public_pane_numbers` (an object that maps internal pane numbers to public pane numbers)
- `next_public_pane_number` (int)
- `public_tab_numbers` (array)
- `next_public_tab_number` (int)
- `tabs` (array)
- `active_tab` (int)

Example values: `"public_pane_numbers":{"5":1}`, `"next_public_pane_number":4`, `"public_tab_numbers":[1]`.

Each **tab** entry has these keys:

- `custom_name` (string \| null)
- `layout` (a recursive tree, see below)
- `panes` (an object with internal pane numbers as keys and `{ cwd, label?, agent_name?, agent_resume?, agent_session?, launch_argv? }` as values)
- `zoomed` (bool)
- `focused` (int pane number)
- `root_pane` (int)

Checked 2026-10-17 against herdr 0.9.5: the workspace keys and the tab keys, on a scratch file and on the live default file. They have no change since 0.9.1.

herdr omits each optional pane key when it does not apply. herdr does not send it as `null`. A usual shell pane has only `{"cwd": "…"}`. On 0.9.5, a pane entry can hold these keys:

- `label`: the custom label of the pane.
- `agent_name`: the unique name of the agent.
- `agent_resume`: `{ source, agent, argv }`, the `resume_argv` from the last `pane.report_agent` or `pane.report_agent_session` that sent one. On a restart, herdr types `argv` into the shell of the restored pane (see [AgentSessionInfo](#agentsessioninfo)).
- `agent_session`: `{ source, agent, kind, value }`, as in [AgentSessionInfo](#agentsessioninfo). A restored pane uses it to connect again to the native session of its agent. On 0.9.5, a reported session was saved only for the source `herdr:claude`.
- `launch_argv`: new on 0.9.5, an array of strings. A pane that a plugin opened saves the command of the pane here. After a restart, the restored pane is a plain shell at the saved `cwd`, with the saved label. herdr does not run the argv again, and the pane is no longer a plugin pane. `plugin.pane.close` on it gave `plugin_pane_not_found`.

Changes reached `session.json` 2 to 5 s after the call. For the agent kind `probe`, herdr saved `agent_name` and `agent_resume` only with a named process in the foreground of the pane. The pass observed this condition only for that kind. On 0.9.5, herdr dropped a `probe` report from a plain-shell pane within 1 s (0.2 to 0.3 s in this pass). herdr kept reports for `claude` and `codex` (see [AgentStatus](#agentstatus) and [api/pane.md](api/pane.md#panereport_agent)). On 0.9.3, one pane entry held all five pane keys of that version. On 0.9.5, the pass found those keys only on two different entries. herdr saved `agent_session` only for the source `herdr:claude`. herdr refused `resume_argv` from `herdr:claude` on a pane that a different source held (`resume_not_accepted`). Checked 2026-10-17 against herdr 0.9.5. After a restart, the first pane gave `$ no-such-cmd-xyz --resume` and `bash: no-such-cmd-xyz: command not found`:

```json
{"cwd":"/home/user/project","label":"reviewer","agent_name":"reviewer","agent_resume":{"source":"herdr:probe","agent":"probe","argv":["no-such-cmd-xyz","--resume"]}}
{"cwd":"/home/user/project","agent_session":{"source":"herdr:claude","agent":"claude","kind":"id","value":"0000-test"}}
{"cwd":"/home/user/project/acme","label":"Side","launch_argv":["sh","-c","exec sleep 600"]}
```

The saved **layout** tree is a recursive union. Its variant keys start with a capital letter, and the direction value is also in PascalCase. A leaf is `{ "Pane": <internal-number> }`. An inner node is `{ "Split": { direction, ratio, first, second } }`. `Horizontal` is a `right` split, and `Vertical` is a `down` split. Checked 2026-10-17 against herdr 0.9.5:

```json
{"Split":{"direction":"Horizontal","ratio":0.5,"first":{"Split":{"direction":"Vertical","ratio":0.5,"first":{"Pane":13},"second":{"Pane":15}}},"second":{"Pane":14}}}
```

A server restart with an existing session directory that is not empty starts a real shell process for each saved pane, at its saved `cwd`. The server does not start clean. Thus a named session directory that a caller uses again brings back its last contents. On 0.9.3 and later, herdr restores a pane whose saved `cwd` no longer exists as a placeholder with no shell, with `restore_error` (see [PaneInfo](#paneinfo)). Checked 2026-10-17 against herdr 0.9.5 for a named session. Each pane came back with a live `bash`, and `pane.process_info` gave a `shell_pid`. This pass did not check the default session, because it is the live session of a person.

```json
{"pane_id":"wA:p2","cwd":"/home/user/project/gone (deleted)","agent_status":"unknown","restore_error":"Saved directory is unavailable. Restore the directory and restart this session."}
```

This JSON gives only some of the `PaneInfo` fields. This page gives the structure of the `session.json` files. It does not give real `cwd` or `identity_cwd` values, because they are personal paths.
