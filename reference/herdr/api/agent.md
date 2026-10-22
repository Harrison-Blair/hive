# herdr API: agent methods

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Index: [README.md](../README.md). Wire format: [protocol.md](../protocol.md).

The `agent.*` methods inspect and control the agent in a pane.

- An agent is not a process that is separate from the pane. It is the herdr classification of the program that runs in the terminal of the pane.
- An agent has an agent status: `idle`, `working`, `blocked`, `done`, or `unknown`. [../data-model.md](../data-model.md#agentstatus) gives the meaning of each value.
- An agent can have an agent name. The name follows the agent that is in the pane now.

Targets:

- Every method except `agent.list`, `agent.start`, `agent.view.set`, and `agent.view.clear` takes a `target`.
- A `target` is **either** a unique live agent name (pattern `[a-z][a-z0-9_-]{0,31}`) **or** the pane ID of the pane that holds the agent (for example, `w2:p1`).
- herdr does not accept a terminal ID or a bare agent kind as a target. Such a target gives `agent_not_found`. A pane ID of a pane with no agent also gives `agent_not_found`.
- `agent.start` has no `target` field. It uses `pane_id` to identify the pane, and it gives the `name` to the new agent.
- `agent.start` needs a shell pane that exists and is at its interactive prompt. It does not create, split, or move panes.
- The `agent.view.*` methods set a saved filter and sort order on the agent list for a client. They do not target one agent.

Agent names:

- herdr removes the name when the agent exits. herdr also removes the name when a release occurs, as with `pane.release_agent`.
- When a different agent kind takes over the pane, herdr removes the name. For example, a `claude` that started in the shell of a pane with a suspended `opencode` agent had no name.
- A new process of the same agent kind can keep the name. A second `opencode` that started in the same shell kept the name of the suspended `opencode` agent.
- In that case, herdr removed the old `agent_session`, and `state_change_seq` increased. Thus, a name does not prove that the same process is in the pane.
- [Agent identity across Ctrl+Z](#agent-identity-across-ctrlz) gives the rules for a suspended agent.

`Checked 2026-10-17 against herdr 0.9.5`: this stamp applies to all claims in this introduction. These claims include:

- the accepted and the rejected targets
- the `agent_pane_busy` error and the unchanged layout after `agent.start`
- the name after an exit, after a different agent kind, and after a new process of the same kind

12 methods:

| method | purpose |
| --- | --- |
| [agent.explain](#agentexplain) | Show how herdr detected and classified an agent. |
| [agent.focus](#agentfocus) | Focus the pane of an agent and mark it seen. |
| [agent.get](#agentget) | Get the full information of one agent. |
| [agent.list](#agentlist) | List all live agents. |
| [agent.prompt](#agentprompt) | Send prompt text to an agent. The call can wait for a status. |
| [agent.read](#agentread) | Read the terminal output of the pane of an agent. |
| [agent.rename](#agentrename) | Set or remove the name of an agent. |
| [agent.send_keys](#agentsend_keys) | Send logical key presses to an agent. |
| [agent.start](#agentstart) | Start a supported interactive agent in a pane. |
| [agent.view.clear](#agentviewclear) | Make the saved agent view of a client not active. |
| [agent.view.set](#agentviewset) | Set the saved agent view of a client and make it active. |
| [agent.wait](#agentwait) | Wait until an agent has a requested status. |

## Agent identity across Ctrl+Z

herdr 0.9.5 gives these rules for an agent that a person suspends with Ctrl+Z:

- herdr keeps the identity of the agent while that exact process continues to run. The identity stays when a command that is not an agent runs in the shell.
- herdr does not keep the original identity when a different agent that herdr recognizes takes over the pane.
- A kept identity does not mean that the agent is in the foreground. Read the pane before you send input.

`Documented by herdr --skill (herdr 0.9.5); not live-checked: a pass can test only examples of these general rules`. A pass checked the behavior below.

While the agent is suspended, the methods of this page operate as follows:

- `agent.get` and `agent.list` show the old name, agent kind, agent status, and `agent_session`. The status does not change while other commands run in the shell.
- `pane.process_info` shows the shell as the foreground process, not the agent.
- `agent.wait` returns immediately with the old status.
- `agent.read` returns the screen of the shell, not the screen of the agent.
- `agent.prompt` returns `agent_not_ready` (`"agent oc is no longer the pane foreground process"`).
- `agent.send_keys` returns `agent_not_ready` (`"agent oc is not an active named agent"`).
- `agent.start` on the pane returns `agent_pane_busy`. The shell is at its prompt, but the suspended agent still holds the pane.
- After `fg` in the shell, the agent is in the foreground again with the same name. Then `agent.send_keys` operates again.

```text
agent.send_keys {"target":"oc","keys":["ctrl+z"]} -> ok      # shell prints "[1]+  Stopped  opencode"
pane.process_info w1:p1 -> foreground_processes: [bash]
agent.get w1:p1 -> name "oc", agent "opencode", agent_status "idle", agent_session kept
pane.send_text "sleep 2; echo nonagent-done\r" -> during and after: still ('oc','opencode','idle')
agent.wait {"target":"oc","timeout_ms":1000} -> returns at once, agent_status "idle"  (0.2 ms)
agent.read {"target":"oc","source":"visible"} -> ok, returns the bash screen ("nonagent-done")
agent.prompt {"target":"oc","text":"..."} -> agent_not_ready "agent oc is no longer the pane foreground process"
agent.send_keys {"target":"oc","keys":["ctrl+l"]} -> agent_not_ready "agent oc is not an active named agent"
agent.start {"name":"second","kind":"opencode","pane_id":"w1:p1"} -> agent_pane_busy "agent target pane w1:p1 is not an available shell"
pane.send_text "fg\r" -> agent_status "idle", name "oc"; agent.send_keys oc ["ctrl+l"] -> ok
```

`Checked 2026-10-17 against herdr 0.9.5`: the behavior while suspended, with an `opencode` agent.

## Shared types

### AgentInfo

`agent.get`, `agent.focus`, `agent.rename`, `agent.start`, `agent.wait`, and `agent.prompt` return `AgentInfo` as `agent`. `agent.list` returns it as array elements. [../data-model.md](../data-model.md#agentinfo) also lists this type.

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the field types, the required flags, and the defaults in this table. The notes after the table give the checked behavior.

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `terminal_id` | string | yes | — | Stable internal handle of the terminal that holds the agent. |
| `agent_status` | enum | yes | — | Agent status: `idle`, `working`, `blocked`, `done`, or `unknown`. Refer to [AgentStatus](../data-model.md#agentstatus). |
| `workspace_id` | string | yes | — | Workspace of the pane that holds the agent (for example, `w2`). |
| `tab_id` | string | yes | — | Tab of the pane that holds the agent (for example, `w2:t1`). |
| `pane_id` | string | yes | — | Public ID of the pane that holds the agent (for example, `w2:p1`). |
| `focused` | boolean | yes | — | True when the pane that holds the agent has the focus in the herdr UI. |
| `revision` | uint64 | yes | — | Monotonic revision of the output and state of the pane. It increases on each change. |
| `agent` | string \| null | no | — | Detected agent kind (for example, `claude` or `codex`). herdr does not write the field until the classification is complete. |
| `display_agent` | string \| null | no | — | Agent label for persons (inferred). |
| `name` | string \| null | no | — | Unique agent name, or no value when the agent has no name. |
| `title` | string \| null | no | — | Title that the agent reports (inferred). |
| `cwd` | string \| null | no | — | Working directory of the shell of the pane. |
| `foreground_cwd` | string \| null | no | — | Working directory of the foreground process. |
| `agent_session` | AgentSessionInfo \| null | no | — | Reference to the agent session. Refer to [AgentSessionInfo](#agentsessioninfo). |
| `interactive_ready` | boolean | no | — | True when the agent is ready for interactive input (inferred). Refer to the notes below. |
| `launch_pending` | boolean | no | — | True while a launch from `agent.start` continues (inferred). |
| `screen_detection_skipped` | boolean | no | — | True when herdr does not use screen detection for the agent. Refer to the notes below. |
| `state_change_seq` | uint64 | no | `0` | Monotonic counter. It increases on each change of agent status. Refer to the notes below. |
| `completion_seq` | uint64 \| null | no | — | New in 0.9.3. The `state_change_seq` of the transition that completed a turn. Refer to the notes below. |
| `state_labels` | object<string,string> | no | — | State labels from detection, as a map from label key to value. |
| `terminal_title` | string \| null | no | — | Raw terminal title (OSC). |
| `terminal_title_stripped` | string \| null | no | — | Terminal title without control characters and markup. |
| `tokens` | object<string,string> | no | — | Custom tokens. The maximum is 32 keys, with the key pattern `^[A-Za-z0-9_-]{1,32}$`. An `agent.view` filter or sort can use them as fields. |

#### Null values

The schema types only some optional fields as nullable (`X \| null`, or `anyOf` with a null variant for `agent_session`):

- `agent`, `display_agent`, `name`, `cwd`, `foreground_cwd`, `agent_session`, `terminal_title`, `terminal_title_stripped`, `title`
- `completion_seq` (`uint64 \| null`)

The schema types the other optional fields as follows:

- plain boolean, not nullable: `interactive_ready`, `launch_pending`, `screen_detection_skipped`
- plain `object`: `state_labels`, `tokens`
- plain `integer` (`uint64`, default `0`): `state_change_seq`

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`.

The type does not change the output. herdr 0.9.1, 0.9.3, and 0.9.5 never write `null` for these fields. herdr does not write a field that has no value. For example, `agent` and `interactive_ready` are absent while a launch from `agent.start` continues. Read each optional field with a default value. Do not expect a literal `null`.

`Checked 2026-10-17 against herdr 0.9.5`: no `null` value in an `agent.list` of five agents, in a pending `agent.start` result, or in the event log of the session. The free-form `explain` object of `agent.explain` can contain `null`, for example `fallback_reason: null`. That object is not `AgentInfo`.

#### interactive_ready

`interactive_ready` was present only on an agent that `agent.start` started and that still had its name. The field disappeared when `agent.rename` removed the name. It did not come back when a later `agent.rename` set a name again. An agent that started in another way did not show the field. Examples are an agent that a person typed in the shell, and an agent that lost its name after a short timeout.

`Checked 2026-10-17 against herdr 0.9.5`.

#### state_change_seq

`state_change_seq` is a sequence for the full **session**, not a counter for one agent. All agents in the session share it. Thus, two agents never show the same value. The value of one agent can increase by more than one between two reads, because other agents changed their status.

`Checked 2026-10-17 against herdr 0.9.5`: the values increased across five agents, and no two agents had the same value.

#### completion_seq

`completion_seq` (new in 0.9.3) tells a caller that the current settled status of the agent came after completed work. The schema says that the current idle transition completed work, independent of who saw the pane.

- A new `idle` agent from a launch does not have `completion_seq`.
- When a prompted turn ends while the pane is not seen, the agent status is `done`. `completion_seq` is equal to the `state_change_seq` of that transition (for example, both `10`).
- `agent.focus` then changes the status to `idle`, and `completion_seq` stays at `10`.
- When a turn ends while the pane has the focus, the status changes directly to `idle`, with `completion_seq` set (for example, `state_change_seq` `12` and `completion_seq` `12`).
- The key is absent while the agent is `working`.
- Unlike `done`, the value does not depend on whether a person saw the pane.
- A synthetic agent from `pane.report_agent` also shows it. A report of `working` and then `idle` on a pane in a workspace without the focus gave `done`. The `state_change_seq` was `20` and the `completion_seq` was `20`. The same reports on a pane without the focus in the focused tab gave `idle`, with `completion_seq` set. herdr 0.9.3 did not show the key for a synthetic agent.

`Checked 2026-10-17 against herdr 0.9.5`.

#### tokens and state_labels

`tokens` collects keys from all sources, up to the limit of 32 keys on `AgentInfo`. A 33rd key gives `metadata_token_limit: "pane metadata may contain at most 32 tokens"`. One `pane.report_metadata` call can set a maximum of 16 keys. A call with more keys gives `invalid_metadata_token: "a metadata report may update at most 16 tokens"`. Send more than 16 new tokens in two or more calls.

`state_labels` accepts only a closed vocabulary. No published source lists this vocabulary. herdr rejected each key that a pass tried (`phase`, `lane`, `status`, `detail`) with `invalid_state_label` (`"unknown state label: <key>"`). The `--skill` text, the agent guide, and the schema give no vocabulary.

`Checked 2026-10-17 against herdr 0.9.5`.

#### screen_detection_skipped

On herdr 0.9.5, `AgentInfo` shows `"screen_detection_skipped": true` for an agent with a hook integration that reports the full lifecycle. In this case, herdr does not use screen detection.

- An `opencode` agent with its hook integration installed showed the field after its first prompt. The field then stayed set.
- `agent.explain` for the same agent gives the reason: `screen_detection_skip_reason: "full_lifecycle_hook_authority"`.
- A `claude` agent that herdr detected from the screen did not show the field. A synthetic agent did not show it.
- On herdr 0.9.1 and 0.9.3, a pass never saw the field on `AgentInfo`.

`Checked 2026-10-17 against herdr 0.9.5`.

### AgentSessionInfo

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the types in this table and the `path` value of `kind`.

| field | type | required | meaning |
| --- | --- | --- | --- |
| `source` | string | yes | Origin of the session reference (for example, `herdr:claude`). |
| `agent` | string | yes | Agent kind of the agent session. |
| `kind` | enum | yes | `id` or `path`. It tells how to read `value`. |
| `value` | string | yes | The session identifier or path. |

An `opencode` agent gave `{"source":"herdr:opencode","agent":"opencode","kind":"id","value":"ses_…"}`. The field came after the first prompt, not at the launch. `Checked 2026-10-17 against herdr 0.9.5`.

## agent.explain

Shows how herdr classified the agent in the target pane, or why the classification failed. The result gives the detected agent kind, the current status, the manifest and rule that matched, and the matched evidence. The method is read-only. It does not mark the pane seen.

**Params** (`AgentTarget`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `target` | string | yes | — | Live agent name or the ID of the pane that holds the agent. |

**Result** — `type: "agent_explain"`:

| field | type | required | meaning |
| --- | --- | --- | --- |
| `type` | const `"agent_explain"` | yes | Result discriminator. |
| `explain` | any | yes | Free-form detection explanation: agent kind, status, manifest, rule, and evidence. The schema accepts any JSON value. |

The `explain` object has 18 keys for an agent that herdr detects from the screen:

- `agent`, `state`, `matched_rule`, `evaluated_rules` (each rule has its own `evidence`)
- `manifest_source`, `manifest_version`, `cached_remote_version`, `local_override_shadowing_remote`
- `remote_update_status`, `remote_update_error`, `fallback_reason`
- `screen_detection_skipped`, `skip_state_update`, `skipped_update_reason`
- `visible_idle`, `visible_blocker`, `visible_working`, `warning`

On herdr 0.9.5, an agent with a hook integration that reports the full lifecycle has a 19th key: `screen_detection_skip_reason`. For that agent, the value was `"full_lifecycle_hook_authority"`. The other values were `screen_detection_skipped: true`, `manifest_source: null`, `manifest_version: null`, `matched_rule: null`, and `evaluated_rules: []`. herdr 0.9.1 and 0.9.3 gave the 18-key object.

`Checked 2026-10-17 against herdr 0.9.5`: the 18 keys for a blocked `claude` agent, the 19 keys for an `opencode` agent, and no events after the calls. A synthetic `done` agent stayed `done` after a call.

**Errors**: `agent_not_found` (target not valid or not unique). Other codes are possible. Refer to [errors.md](../errors.md#observed-error-codes).

**CLI**: `herdr agent explain [TARGET] [--file <PATH>] [--agent <LABEL>] [--json] [--format text|json] [-v]`

- The CLI shows `explain` as text.
- `--json` and `--format json` print the bare `explain` object, not the response envelope.
- `-v` adds the lines `visible:`, `cached_remote_version:`, `local_override_shadowing_remote:`, `remote_update_status:`, and `evaluated_rules:`. For the `opencode` agent with hook authority, `-v` added only `visible:`, `cached_remote_version: none`, and `local_override_shadowing_remote: false`.
- The `--file`/`--agent` form uses no live pane. It prints `rule: none` and a `fallback_reason` line (for example, `fallback_reason: default_known_agent_idle_fallback`).

**Example** (a `claude` agent on a scratch server, blocked at its workspace-trust dialog):

```text
agent: claude
state: blocked
manifest: remote:/home/user/.local/state/herdr/agent-detection/remote/claude.toml 2026.10.11.1
rule: live_blocked_form (region=after_last_horizontal_rule priority=980)
evidence: " Accessing workspace:\n\n /home/user/project\n\n Quick safety check: Is this a project you created or one you trust? (Like your own code, a we..."
```

**Example** (an idle `opencode` agent with hook authority):

```text
agent: opencode
state: idle
manifest: none unknown
rule: none
screen_detection_skip_reason: full_lifecycle_hook_authority
```

**Example** (the `--file`/`--agent` form, with `sample.txt` that contains `❯ ` and a newline):

```text
$ herdr agent explain --file sample.txt --agent claude
agent: claude
state: idle
manifest: remote:/home/user/.local/state/herdr/agent-detection/remote/claude.toml 2026.10.11.1
rule: none
fallback_reason: default_known_agent_idle_fallback
```

`Checked 2026-10-17 against herdr 0.9.5`: the CLI options, the `-v` lines, the `--json` output, and the three examples. A pass anonymized the paths in the examples. The evidence text is shorter than the real text.

## agent.focus

Focuses the pane that holds the target agent and marks it seen. When herdr marks an unseen `done` agent seen, the agent status changes to `idle`. The herdr `--skill` text says that explicit focus commands mark the target seen. Reads do not mark it seen.

**Params** (`AgentTarget`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `target` | string | yes | — | Live agent name or the ID of the pane that holds the agent. |

**Result** — `type: "agent_info"`: the [AgentInfo](#agentinfo) of the focused agent under `agent`, with `focused: true`.

| field | type | required | meaning |
| --- | --- | --- | --- |
| `type` | const `"agent_info"` | yes | Result discriminator. |
| `agent` | [AgentInfo](#agentinfo) | yes | The focused agent. |

**Errors**: `agent_not_found`. Other codes are possible.

**Events**: when the call moves the focus to the target pane, herdr sends `pane_focused`, `tab_focused`, and `workspace_focused` together. herdr sends all three also when the tab and the workspace of the target pane were already active. When the target pane already has the focus, herdr sends **no** events. On herdr 0.9.1, each call sent all three events.

**CLI**: `herdr agent focus <target>`

**Example** (before this call, the agent `reviewer` was `done`):

```json
{"id":"1","method":"agent.focus","params":{"target":"reviewer"}}
{"id":"1","result":{"type":"agent_info","agent":{"terminal_id":"term_65e0dbacd49af3","name":"reviewer","agent":"opencode","terminal_title":"OC | Empty conversation","terminal_title_stripped":"OC | Empty conversation","agent_status":"idle","screen_detection_skipped":true,"agent_session":{"source":"herdr:opencode","agent":"opencode","kind":"id","value":"ses_…"},"workspace_id":"w2","tab_id":"w2:t1","pane_id":"w2:p1","focused":true,"interactive_ready":true,"state_change_seq":10,"completion_seq":10,"cwd":"/home/user/project","foreground_cwd":"/home/user/project","revision":3}}}
```

`Checked 2026-10-17 against herdr 0.9.5`: the change from `done` to `idle`, the reads that did not mark the agent seen, `focused: true`, the events, and this example.

## agent.get

Returns the full [AgentInfo](#agentinfo) of one agent. The method is read-only. It does not mark the pane seen.

**Params** (`AgentTarget`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `target` | string | yes | — | Live agent name or the ID of the pane that holds the agent. |

**Result** — `type: "agent_info"`:

| field | type | required | meaning |
| --- | --- | --- | --- |
| `type` | const `"agent_info"` | yes | Result discriminator. |
| `agent` | [AgentInfo](#agentinfo) | yes | The requested agent. |

**Errors**: `agent_not_found` (for example, `"agent target nobody not found"`). Other codes are possible.

**CLI**: `herdr agent get <target>`

herdr rejects a terminal ID and a bare agent kind. On herdr 0.9.5, `"target":"term_65e0db481a22c1"` and `"target":"opencode"` both gave `agent_not_found` (`"agent target opencode not found"`). Two `opencode` agents were live, and neither had the name `opencode`.

**Example**:

```json
{"id":"cli:agent:get","method":"agent.get","params":{"target":"w1:p1"}}
{"id":"cli:agent:get","result":{"type":"agent_info","agent":{"terminal_id":"term_65e0db481a22c1","name":"oc","agent":"opencode","terminal_title":"OpenCode","terminal_title_stripped":"OpenCode","agent_status":"idle","workspace_id":"w1","tab_id":"w1:t1","pane_id":"w1:p1","focused":true,"interactive_ready":true,"state_change_seq":7,"cwd":"/home/user/project","foreground_cwd":"/home/user/project","revision":6}}}
```

`Checked 2026-10-17 against herdr 0.9.5`: the read-only behavior, the rejected targets, `agent_not_found`, and this example.

## agent.list

Lists every live agent that herdr recognizes in all workspaces. The method takes no parameters. It is read-only.

**Params** (`EmptyParams`): none. Send `"params": {}`.

**Result** — `type: "agent_list"`:

| field | type | required | meaning |
| --- | --- | --- | --- |
| `type` | const `"agent_list"` | yes | Result discriminator. |
| `agents` | array<[AgentInfo](#agentinfo)> | yes | All live agents. The array is empty when there are no agents. |

**Errors**: a pass saw no error for a correct request. The wire layer gives `invalid_request` in two cases:

- `params` is absent: `` "invalid request: missing field `params` at line 1 column 35" ``.
- `params` is `null`: `"invalid request: invalid type: null, expected struct EmptyParams at line 1 column 46"`.

**CLI**: `herdr agent list`

**Example**:

```json
{"id":"1","method":"agent.list","params":{}}
{"id":"1","result":{"type":"agent_list","agents":[]}}
```

A list on a scratch server gave five agents:

- a blocked `claude` at `w1:p1`
- three `opencode` agents at `w1:p2`, `w1:p3`, and `w2:p1`
- a synthetic `claude` at `w2:p3`

`Checked 2026-10-17 against herdr 0.9.5`: the empty list (this example), the list of five agents, and the two `params` errors.

## agent.prompt

Sends prompt text to the agent. The method uses the live bracketed-paste mode of the pane. It sends the text, and then an encoded Enter after a short delay.

- The herdr `--skill` text says that herdr sends the text and the Enter as one ordered submission. herdr reports a successful submission only after it writes both. That success alone does not prove that the agent started a turn.
- When the agent is at an approval or question dialog, herdr rejects the prompt with `agent_blocked` **before it sends input**. This occurs only for a dialog that the detection manifest recognizes as blocking.
- herdr can report `idle` for a modal that the manifest does not recognize. Then herdr accepts the prompt and types it into the modal. The call fails after about 5.3 s with `agent_prompt_stalled`. Examples of such a modal are the Rewind menu of Claude Code and the command palette of OpenCode.

The `wait` option:

- When you give `wait` and the agent is not `working`, herdr first needs an observed `working` or `blocked` status within 5000 ms. A change to `idle` or `done` does not satisfy this check. Otherwise, herdr returns `agent_prompt_stalled`. A shorter `timeout_ms` returns `timeout`.
- Then herdr matches the first settled `idle`, `done`, or `blocked` status, or any exact status in `until`.
- The wait follows the agent status, not one turn. When the agent is already `working`, the completion of the active turn can satisfy the wait.
- Without `timeout_ms`, the settled wait continues with no limit.
- `idle` and `done` do not always mean that the agent is ready for input. [../data-model.md](../data-model.md#agentstatus) gives this caveat. A suspended agent also shows its old status (refer to [Agent identity across Ctrl+Z](#agent-identity-across-ctrlz)).

**Params** (`AgentPromptParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `target` | string | yes | — | Live agent name or the ID of the pane that holds the agent. |
| `text` | string | yes | — | Prompt text to send. |
| `wait` | AgentPromptWaitOptions \| null | no | null | Wait behavior after the submission. With `null` or no field, herdr sends the prompt and does not wait. |

`AgentPromptWaitOptions`:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `until` | array<AgentStatus> | no | `[]` | Statuses to match after the submission. An empty array means the default settled set (`idle`, `done`, `blocked`). Values: `idle`, `working`, `blocked`, `done`, `unknown`. |
| `timeout_ms` | uint64 \| null | no | null | The call fails after this number of milliseconds. With `null`, the wait has no limit. |

**Result** — `type: "agent_prompted"`:

| field | type | required | meaning |
| --- | --- | --- | --- |
| `type` | const `"agent_prompted"` | yes | Result discriminator. |
| `agent` | [AgentInfo](#agentinfo) | yes | The agent after the submission, and after the wait if the request has one. |

**Errors**:

- `empty_agent_prompt`: `text` is `""` (`"agent prompt must not be empty"`).
- `agent_blocked`: the agent is at an approval or question UI (`"agent blk is blocked and requires interactive input"`). A pending launch that reached `blocked` also gives `agent_blocked`, not `agent_not_ready`.
- `agent_prompt_stalled`: no observed `working` or `blocked` status within 5000 ms from a status that is not `working`.
- `agent_not_ready`: the target cannot take a prompt now. The causes that a pass saw include these:
  - The `agent.start` launch of the target continues and the status is still `unknown` (`"agent blk is not an active named agent"`).
  - A synthetic agent from `pane.report_agent` is not the foreground process of the pane (`"agent synth is no longer the pane foreground process"`).
  - The agent is suspended with Ctrl+Z (`"agent oc is no longer the pane foreground process"`).
- `timeout`: the wait took more time than a shorter `timeout_ms`.
- `agent_not_found`.
- Other codes are possible.

**CLI**: `herdr agent prompt <TARGET> <TEXT> [--wait] [--until <STATUS>]... [--timeout <MS>]`

**Example** (the pane is in a workspace without the focus, so the result is `done`. The call returned after 2302 ms):

```json
{"id":"1","method":"agent.prompt","params":{"target":"reviewer","text":"Reply with the single word OK. Do not use tools.","wait":{"until":[],"timeout_ms":120000}}}
{"id":"1","result":{"type":"agent_prompted","agent":{"terminal_id":"term_65e0dbacd49af3","name":"reviewer","agent":"opencode","terminal_title":"OpenCode","terminal_title_stripped":"OpenCode","agent_status":"done","screen_detection_skipped":true,"agent_session":{"source":"herdr:opencode","agent":"opencode","kind":"id","value":"ses_…"},"workspace_id":"w2","tab_id":"w2:t1","pane_id":"w2:p1","focused":false,"interactive_ready":true,"state_change_seq":10,"completion_seq":10,"cwd":"/home/user/project","foreground_cwd":"/home/user/project","revision":2}}}
```

`Checked 2026-10-17 against herdr 0.9.5`: these claims and measurements:

- A prompt without `wait` returned in 301 ms.
- `empty_agent_prompt`, `agent_blocked` at the trust dialog of `claude`, the three `agent_not_ready` causes, `timeout`, and `agent_not_found`.
- A prompt into the command palette of OpenCode (the agent was still `idle`) failed after 5304 ms with `agent_prompt_stalled`: `"agent prompt produced no observed working or blocked state within 5000 ms; current status is idle"`. The palette search box then showed the text two times.
- The same prompt with `timeout_ms: 2000` failed after 2002 ms with `timeout`.
- `wait: {}` and `until: []` used the defaults.
- This example.

A pass on herdr 0.9.3 saw the same `agent_prompt_stalled` behavior with the Rewind menu of Claude Code (opened with `esc esc`). On herdr 0.9.3, a prompt without `wait` returned in about 300 ms. The stalled prompt failed after 5307 ms, and the prompt with `timeout_ms: 2000` failed after 2003 ms.

`Documented by herdr --skill (herdr 0.9.5); not live-checked: a pass cannot see a turn from the socket, and a wait with no limit blocks the pass`: the ordered submission, the completion of an active turn, and the wait with no limit.

`Documented by herdr --skill (herdr 0.9.5); not live-checked: the pass did not cause a change to only idle or done within 5000 ms`: the rule that a change to `idle` or `done` does not satisfy the 5000 ms check.

## agent.read

Reads a snapshot of the terminal output of the pane of the agent. The method returns the shared `pane_read` result (`PaneReadResult`). Refer to [pane.md](pane.md) and [../data-model.md](../data-model.md#panereadresult).

- A CLI read does **not** mark the pane seen.
- `lines` requests more rows from the screen and the host scrollback of the pane.
- Rows of an alternate screen do not go into the ordinary host scrollback.
- For a supported idle agent, herdr can collect the history that the application keeps, and then restore the viewport. herdr cannot recover every application or response in this way.

**Params** (`AgentReadParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `target` | string | yes | — | Live agent name or the ID of the pane that holds the agent. |
| `source` | enum | yes | — | Snapshot source. `visible` is the rendered viewport. `recent` is the recent output with soft wraps. `recent_unwrapped` joins soft wraps, and is best for logs and transcripts. `detection` is the plain-text snapshot of the bottom buffer that agent detection uses. The CLI default is `recent`. |
| `format` | enum | no | `text` | `text` or `ansi`. Use `ansi` when colors and styles are evidence. `source: "detection"` ignores this field. |
| `lines` | uint32 \| null | no | null | Number of rows to request. With `null`, herdr uses the default extent. A value smaller than the default keeps only the bottom rows (`truncated: true`). `lines: 0` returns an empty string with `truncated: true`, not an error. |
| `strip_ansi` | boolean | no | `true` | The name of the field says that it removes ANSI escapes from the text. On herdr 0.9.1, 0.9.3, and 0.9.5, it has no observable effect in either direction. Only `format` (`text` or `ansi`) sets whether escapes occur. |

`source: "detection"` with `format: "ansi"` returned plain text with no escapes, the same 835 bytes as `format: "text"`. `read.format` showed `ansi`.

**Result** — `type: "pane_read"`, with `read` (`PaneReadResult`):

| field | type | required | meaning |
| --- | --- | --- | --- |
| `type` | const `"pane_read"` | yes | Result discriminator. |
| `read.pane_id` | string | yes | Pane of the text. |
| `read.workspace_id` | string | yes | Workspace of the pane. |
| `read.tab_id` | string | yes | Tab of the pane. |
| `read.source` | enum | yes | The source of the request (`visible`, `recent`, `recent_unwrapped`, or `detection`). |
| `read.format` | enum | yes | The format of the request (`text` or `ansi`). |
| `read.text` | string | yes | The terminal text. |
| `read.revision` | uint64 | yes | Output revision of the pane at the time of the read. |
| `read.truncated` | boolean | yes | True when herdr removed rows from the snapshot. |

**Errors**: `agent_not_found` (`"agent target nobody not found"`). Other codes are possible. A request without `source` gives `invalid_request` (`` "invalid request: missing field `source` at line 1 column 63" ``).

**CLI**: `herdr agent read <TARGET> [--source visible|recent|recent-unwrapped|detection] [--lines <N>] [--format text|ansi] [--ansi]`

- The CLI accepts the spelling `recent-unwrapped` and changes it to the API value `recent_unwrapped`.
- herdr rejects `recent-unwrapped` on the wire with `invalid_request` (`` "invalid request: unknown variant `recent-unwrapped`, expected one of `visible`, `recent`, `recent_unwrapped`, `detection` at line 1 column 93" ``).
- `--ansi` and `--format ansi` give the same output with escapes.

**Example** (the CLI prints `read.text` directly. This is the bottom of an `opencode` screen):

```text
                      ┃
                      ┃  Ask anything… "Fix broken tests"
                      ┃
                      ╹▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀
                                                     tab agents  ctrl+p commands
  /home/user/project                                                       1.18.34
```

Read times on herdr 0.9.5, with three runs each on a screen of 39 rows:

- A `format: "text"` read of `recent` or `recent_unwrapped` with no `lines`, or a large `lines`, took 374–376 ms.
- `format: "ansi"`, `source: "detection"`, `source: "visible"`, and a small `lines` value took less than 1 ms.
- On herdr 0.9.5, `recent_unwrapped` was not slower than `recent`. On herdr 0.9.3, `recent` took 395–410 ms and `recent_unwrapped` took 463–485 ms.
- On herdr 0.9.1, `recent` took about 374 ms. On herdr 0.9.3, the fast reads took 1–7 ms, and one `detection` read took 101 ms.

A caller that calls `agent.read` in a loop has this cost on each call.

`Checked 2026-10-17 against herdr 0.9.5`: these claims:

- the four sources, the `lines` truncation, `lines: 0`, the `strip_ansi` result, and the `detection` format
- the wire spelling, the absent `source`, `agent_not_found`, and the CLI default
- the times, and this example

The example omits one line with the model and the provider.

`Documented by herdr --skill (herdr 0.9.5); not live-checked: the alternate screen of the agent did not let a pass test these rows`: the alternate-screen rows and the history collection.

## agent.rename

Sets or removes the unique name of an agent.

- To set the name, give `name`. The name must match `[a-z][a-z0-9_-]{0,31}` and must be unique among live agents.
- To remove the name, give `null`.
- The name follows the agent in the pane until the agent exits or a release occurs. Refer to the introduction for the rules.

**Params** (`AgentRenameParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `target` | string | yes | — | Live agent name or the ID of the pane that holds the agent. |
| `name` | string \| null | no | — | New name. `null` **or no `name` key** removes the current name. The call has no form that keeps the name. |

**Result** — `type: "agent_info"`: the changed [AgentInfo](#agentinfo) under `agent`, with the new `name`.

| field | type | required | meaning |
| --- | --- | --- | --- |
| `type` | const `"agent_info"` | yes | Result discriminator. |
| `agent` | [AgentInfo](#agentinfo) | yes | The renamed agent. |

**Errors**:

- `agent_not_found`.
- `invalid_agent_name`: `name` does not match the pattern (`"agent name must start with a lowercase letter and contain only lowercase letters, digits, '-' or '_' (1-32 characters)"`).
- `agent_name_taken`: another live agent uses the name. The message gives the `terminal_id`, `pane_id`, `workspace_id`, `tab_id`, `cwd`, and `status` of that agent. The status has a capital letter, for example `status=Idle`. A rename of an agent to its current name is not a conflict, and it succeeds.
- `agent_launch_pending` (`"agent name cannot change while startup is pending"`): the `agent.start` launch of the target continues. herdr gives this error with `name` and without `name`.
- Other codes are possible.

**CLI**: `herdr agent rename <TARGET> <NAME>|--clear`

- `--clear` sends `name: null`.
- With no `NAME` and no `--clear`, the CLI refuses the command before it sends a request: `usage: herdr agent rename <target> <name>|--clear`. Thus, only the API has the default that removes the name.

**Example**:

```json
{"id":"1","method":"agent.rename","params":{"target":"w1:p1","name":"api-docs"}}
{"id":"1","result":{"type":"agent_info","agent":{"terminal_id":"term_65e0db481a22c1","name":"api-docs","agent":"opencode","terminal_title":"OpenCode","terminal_title_stripped":"OpenCode","agent_status":"idle","workspace_id":"w1","tab_id":"w1:t1","pane_id":"w1:p1","focused":true,"interactive_ready":true,"state_change_seq":7,"cwd":"/home/user/project","foreground_cwd":"/home/user/project","revision":6}}}
```

**Undocumented default that removes the name**: `agent.rename {"target":"tmpname"}` with no `name` key returns a normal `agent_info` result without the name. The result is exactly the same as for `name: null`. A later `agent.get` with the old name returns `agent_not_found`.

- Always send `name` explicitly.
- An explicit name does not prevent `agent_launch_pending`. Wait until the launch is complete before the rename (`agent.wait`).
- A removal of the name also removes `interactive_ready` from `AgentInfo` (refer to [interactive_ready](#interactive_ready)).

`Checked 2026-10-17 against herdr 0.9.5`: these claims:

- the set, the removal with `null` and without the key, and the rename to the current name
- `agent_name_taken`, `invalid_agent_name`, and `agent_launch_pending` with and without `name`
- the CLI refusal, the removed `interactive_ready`, and this example

## agent.send_keys

Sends logical key presses to the terminal of the agent.

- herdr checks all keys before it writes any bytes. One key that is not valid rejects the full request.
- Use `esc` as the canonical name of the Escape key. herdr also accepts `escape`.
- herdr accepts chords, for example `ctrl+c`.

**Params** (`AgentSendKeysParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `target` | string | yes | — | Live agent name or the ID of the pane that holds the agent. |
| `keys` | array<string> | yes | — | Ordered logical key names to send (for example, `["esc"]` or `["ctrl+c"]`). |

**Result** — `type: "ok"`:

| field | type | required | meaning |
| --- | --- | --- | --- |
| `type` | const `"ok"` | yes | herdr wrote the keys. |

**Errors**:

- `invalid_key`: a key name that herdr does not support rejects the full request (for example, `"unsupported key bogus"`).
- `agent_not_ready`: the target is not an active named agent (`"agent X is not an active named agent"`). A pass saw these causes:
  - a synthetic agent from `pane.report_agent`
  - an `agent.start` launch that continues while the status is still `unknown`
  - an agent that is suspended with Ctrl+Z
- `agent_not_found`.
- Other codes are possible.

herdr accepts an empty `keys` array and returns `ok` for an active agent. For a target that is not ready, an empty array also gives `agent_not_ready`.

A pending launch that reached `blocked` accepts keys. The status was `blocked` and `launch_pending` was still `true`. Then `agent.send_keys` with `["esc"]` returned `ok` and cancelled the trust dialog of `claude`. On herdr 0.9.3, `agent_not_ready` occurred until the launch settled (`idle` or `blocked`).

**CLI**: `herdr agent send-keys <TARGET> <KEY>...`

**Example**:

```json
{"id":"1","method":"agent.send_keys","params":{"target":"reviewer","keys":["esc"]}}
{"id":"1","result":{"type":"ok"}}
```

`Checked 2026-10-17 against herdr 0.9.5`: `esc`, `escape`, `ctrl+c`, `ctrl+z`, `invalid_key`, the three `agent_not_ready` causes, the empty array, the keys to a blocked launch, `agent_not_found`, and this example.

## agent.start

Starts a supported interactive agent of `kind` in an existing pane, and gives it `name`. `pane_id` identifies the pane.

- The pane must be at its interactive shell prompt, with no foreground command.
- `agent.start` does not create, split, or move panes.

### API timing

The API call returns almost immediately after the launch starts. It does not wait until the agent is ready for interactive input.

- herdr 0.9.5 returned in 0.2 ms for each of five launches. herdr 0.9.1 took 0.2–0.3 ms, and herdr 0.9.3 took 0.5–4.9 ms.
- The returned [AgentInfo](#agentinfo) has `agent_status: "unknown"` and `launch_pending: true`. It has no `agent` key and no `interactive_ready` key.
- `state_change_seq` is `0` only on a pane that never held an agent. On a pane that held an agent before, it has the last value of that pane (for example, `3` and then `6`).
- herdr detected the agent kind after about 300 ms for `opencode` and about 700 ms for `claude`. On herdr 0.9.3, the detection came after about 300–540 ms.
- The launch settled after 3.4–4.0 s. On herdr 0.9.3, it settled after 3.4–3.8 s, and on herdr 0.9.1 after about 3.6 s. Then the target has one of two states:
  - `idle`, with `interactive_ready: true` and no `launch_pending`.
  - `blocked`, with `launch_pending` still `true`, when the agent stops at its own startup UI (for example, a workspace-trust dialog).
- A pass never saw the raw API call return `agent_not_ready` in these cases.

During the launch:

- `agent.read` with the name operates all the time.
- `agent.send_keys` and `agent.prompt` with the name return `agent_not_ready` (`"agent X is not an active named agent"`) while the status is `unknown`. On herdr 0.9.1, `agent.send_keys` operated during the launch.
- After the status is `blocked`, `agent.send_keys` operates, and `agent.prompt` returns `agent_blocked`.

A caller that needs a ready agent before `agent.prompt` must wait for the launch:

- Call `agent.wait`. Right after `agent.start`, it returned `idle` after 3502 ms for `opencode`, and `blocked` after 3402 ms for `claude`. On herdr 0.9.3, it returned after 3.4–3.6 s. On herdr 0.9.1, it returned after about 3.6 s, and then `agent.prompt` succeeded on the first attempt.
- Or call `agent.get` until `agent_status` is not `unknown`.
- `agent.wait` matches `agent_status`, not `launch_pending`. Examine the returned status.
- `blocked` means that the agent is at a startup dialog, and `launch_pending` is still `true`. Then `agent.prompt` returns `agent_blocked`.

The default startup timeout is 30 seconds. The raw API shows no effect of this default, because `agent.start` returns before the timeout. On the raw API, the only observable effect of a short timeout is the dropped name below.

### CLI timing

The CLI `herdr agent start` does **not** return immediately. It returns after herdr detects the expected agent in the same terminal, and the agent is ready for input.

- A launch of `opencode` returned exit status `0`, with `agent_status: "idle"` and `interactive_ready: true`.
- A launch of `claude` that stopped at the trust dialog returned exit status `1` after 3917 ms, with `agent_not_ready`. The name stayed, and `agent.get` showed `blocked` with `launch_pending: true`.
- A launch with `--timeout 3001` returned exit status `1` after 3013 ms, with `timeout`. herdr removed the name, and the pane still held an `opencode` agent with no name.

```text
herdr agent start blk --kind claude --pane w1:p3 -- --model claude-haiku-4-5-20251001
{"error":{"code":"agent_not_ready","message":"agent blk is blocked during startup and is not ready for prompts"},"id":"cli:agent:start"}

herdr agent start shorty --kind opencode --pane w1:p3 --timeout 3001
{"error":{"code":"timeout","message":"timed out waiting for agent startup"},"id":"cli:agent:start"}
```

The request `id` is `cli:agent:start`, and the schema has no parameter for this wait. A pass did not find how the CLI waits.

`Checked 2026-10-17 against herdr 0.9.5`: the API timing, the CLI timing, the launch states, and the behavior of the other methods during the launch. The pass started `claude` and `opencode`. It did not start other agent kinds.

### Line limit on macOS

On macOS, `agent start` rejects a command that is longer than the current canonical line-input limit of the terminal. herdr rejects the command before it queues it.

- Wait for the interactive shell prompt before you try again.
- The check applies to this submission only. It does not apply to an earlier incomplete line or to a later change of the terminal mode.
- `herdr pane run` has the same check. [cli-mapping.md](../cli-mapping.md#pane-run-admission) gives the `pane run` part.

`Documented by herdr --skill (herdr 0.9.5); not live-checked: the pass ran on Linux`. The `--skill` text and the schema do not give the error code for this rejection.

### Silent name drop

**Undocumented silent data loss**: a `timeout_ms` that is shorter than the real launch makes the raw API drop the name with no error.

- `agent.start` returns a normal `agent_started` result with the requested `name`.
- When the timeout ends, herdr removes the name. herdr sends no error.
- The agent process continues to run. `agent.get` and `pane.process_info` with the pane ID still operate.
- `agent.get`, `agent.prompt`, and `agent.send_keys` with the removed name return `agent_not_found` from then on.
- A pass saw no `agent_not_ready` or `timeout` error from the raw API in this case. On herdr 0.9.5, the CLI gives `timeout` (refer to [CLI timing](#cli-timing)).

In a check with `timeout_ms: 3001` and an `opencode` launch, the name was gone at 3025 ms. At 3780 ms the agent was `idle`, still with no name and no `interactive_ready`. On herdr 0.9.3, the name was gone about 3.0 s after the call.

`Checked 2026-10-17 against herdr 0.9.5`.

**Params** (`AgentStartParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `name` | string | yes | — | Unique name for the agent (`[a-z][a-z0-9_-]{0,31}`). |
| `kind` | string | yes | — | Agent kind. The CLI supports these values: `pi, claude, codex, gemini, cursor, devin, agy, cline, omp, mastracode, opencode, copilot, kimi, kiro, droid, amp, grok, hermes, kilo, qodercli, qwen, letta, maki, muse`. |
| `pane_id` | string | yes | — | Existing pane at an interactive shell prompt. |
| `args` | array<string> | no | `[]` | Native arguments of the agent. The CLI takes them after `--`. |
| `timeout_ms` | uint64 \| null | no | null (server default 30000) | Startup timeout in milliseconds. The value must be more than 3000 and a maximum of 300000. |

**Result** — `type: "agent_started"`:

| field | type | required | meaning |
| --- | --- | --- | --- |
| `type` | const `"agent_started"` | yes | Result discriminator. |
| `agent` | [AgentInfo](#agentinfo) | yes | The started agent. |
| `argv` | array<string> | yes | The full argument vector that herdr started (for example, `["claude","--model","claude-haiku-4-5-20251001"]`). |

**Errors**:

- `unsupported_agent_kind`: `kind` is not in the supported list (`"unsupported interactive agent kind nope"`).
- `invalid_agent_name`: refer to the name pattern of [AgentInfo](#agentinfo).
- `agent_name_taken`: a live agent uses `name`. The message has the same shape as for [agent.rename](#agentrename).
- `agent_pane_not_found`: the target pane does not exist.
- `agent_pane_busy` (`"agent target pane X is not an available shell"`): the pane is not an available shell. A foreground command runs in the pane, or the pane holds an agent. A suspended agent also holds the pane.
- `invalid_agent_timeout`: `timeout_ms` is not in `(3000, 300000]` (`"agent start timeout must be greater than 3000ms and at most 300000ms"`).
- `agent_not_ready` (agent blocked during startup) and `timeout` (startup timeout): the CLI gives these errors. A pass never saw the raw API give them on herdr 0.9.1, 0.9.3, or 0.9.5. Refer to [CLI timing](#cli-timing) and [Silent name drop](#silent-name-drop).
- Other codes are possible.

herdr checks the request in this sequence: name, kind, pane exists, pane available, timeout.

- When `kind` and `name` are both not valid, herdr returns `invalid_agent_name`, never `unsupported_agent_kind`. An earlier pass saw this result 3 times.
- An unknown kind with an unknown pane gives `unsupported_agent_kind`.
- An unknown pane with `timeout_ms: 5` gives `agent_pane_not_found`.
- A busy pane with `timeout_ms: 5` gives `agent_pane_busy`.

`Checked 2026-10-17 against herdr 0.9.5`: these claims:

- the CLI kind list, the timeout range (3000 and 300001 rejected, 3001 accepted), `args`, and the result
- the check sequence, and each error above except `agent_name_taken`

This pass did not send `agent.start` with a name that another agent used. It checked the same message on [agent.rename](#agentrename).

**Events**: a successful start sends `pane_agent_detected` (with underscores). Each later change of agent status sends `pane.agent_status_changed`, with a **dot**. Subscribers must match the spelling with the dot, not `pane_agent_status_changed`. [events.md](../events.md#event-catalog-eventkind--eventdata) gives the event catalog.

```json
{"data":{"agent":"opencode","pane_id":"w2:p1","type":"pane_agent_detected","workspace_id":"w2"},"event":"pane_agent_detected"}
{"data":{"agent":"opencode","agent_status":"idle","pane_id":"w2:p1","workspace_id":"w2"},"event":"pane.agent_status_changed"}
```

`Checked 2026-10-17 against herdr 0.9.5`: the two events.

**CLI**: `herdr agent start <NAME> --kind <KIND> --pane <ID> [--timeout <MS>] [-- <AGENT_ARG>...]`

The CLI help says: `Success means the expected agent was detected in the same terminal and is ready for input.` The `--timeout` help gives `(default: 30000; max: 300000)`.

**Example** (an error):

```json
{"id":"cli:agent:start","method":"agent.start","params":{"name":"reviewer","kind":"codex","pane_id":"w1:p99"}}
{"id":"cli:agent:start","error":{"code":"agent_pane_not_found","message":"agent target pane w1:p99 not found"}}
```

**Example** (the silent name drop):

```json
{"id":"1","method":"agent.start","params":{"name":"shorty","kind":"opencode","pane_id":"w1:p2","timeout_ms":3001}}
{"id":"1","result":{"type":"agent_started","agent":{"terminal_id":"term_65e0dbacd25682","name":"shorty","terminal_title":"user@host:/home/user/project","terminal_title_stripped":"user@host:/home/user/project","agent_status":"unknown","workspace_id":"w1","tab_id":"w1:t1","pane_id":"w1:p2","focused":false,"launch_pending":true,"state_change_seq":0,"cwd":"/home/user/project","foreground_cwd":"/home/user/project","revision":1},"argv":["opencode"]}}
```

3.0 seconds later, with no other request, `agent.get` for `"shorty"` returns `agent_not_found`. `agent.get` for `"w1:p2"` returns the `opencode` agent that still runs, with no `name` field.

`Checked 2026-10-17 against herdr 0.9.5`: the two examples and the CLI help text.

## agent.view.clear

The method name says that it makes the saved agent view for `source` not active. `source` is an identifier of a client or a view. On herdr 0.9.1, 0.9.3, and 0.9.5, the method **does not do this**:

- The server has one live agent-view slot, not one slot for each `source`.
- `agent.view.clear` operates only when its `source` is the name of the active view.
- With another `source`, the call does nothing. This includes a `source` that was never set, and an earlier `source` that a later `set` replaced.
- Then the call returns the `source` and `label` of the other active view, with `active: true`. It does not return the requested `source` or `active: false`.
- Only a clear of the last `source` that `agent.view.set` set makes the view not active.
- When no view is active, a clear of any `source` returns `active: false`.

Refer to the bug note below and to [agent.view.set](#agentviewset).

**Params** (`AgentViewClearParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `source` | string \| null | no | — | The view or client identifier to clear. With `null` or no field, the call targets the default view of the caller (inferred). herdr checks it with the same rules as the `source` of `agent.view.set`. A `source` that is not the active view has no effect (refer to the bug note). |

**Result** — `type: "agent_view"`:

| field | type | required | meaning |
| --- | --- | --- | --- |
| `type` | const `"agent_view"` | yes | Result discriminator. |
| `active` | boolean | yes | `false` only when the `source` of the request was the active view. Otherwise `true`, for the other view that is still active (refer to the bug note). |
| `source` | string \| null | no | When the clear has no effect, the `source` of the view that is still active. This is not always the requested `source`. |
| `label` | string \| null | no | The `label` of the view that is still active, in the same condition. |

**Errors**: `invalid_agent_view` for a `source` that fails the checks of `agent.view.set`. An example is a `source` with 121 characters: `"agent view source must be non-empty, at most 120 characters, and contain only ASCII letters, digits, colon, dot, underscore, or hyphen"`. A `source` that was never set, or that is not the active view, gives no error. On the wire, the result looks like a successful clear of a different view (refer to the bug note). Other codes are possible.

**Undocumented bug with two sources**: `agent.view.clear` does not clear by `source`. One slot holds the view for the full server. The call reports `active: false` only when the named `source` holds that slot.

1. `agent.view.set` with `source: bugA`.
2. `agent.view.set` with `source: bugB`.
3. `agent.view.clear` with `source: bugA` returns `{"type":"agent_view","active":true,"source":"bugB","label":"B"}`. This is the `bugB` view that is still active, not `active: false` for `bugA`.
4. A next `agent.view.clear` with `source: bugB` returns `{"type":"agent_view","active":false}`.

An earlier pass saw this sequence 3 times.

The same result occurred with `viewX` and `viewY`, and with a `source` that was never set (`onlyOne` and `differentNeverSet`).

**CLI**: API-only (no CLI subcommand). `herdr agent view --help` prints the usage list of `herdr agent`.

**Example** (a clear of the active view):

```json
{"id":"2","method":"agent.view.clear","params":{"source":"api-docs"}}
{"id":"2","result":{"type":"agent_view","active":false}}
```

**Example** (the bug: a set of `bugA` comes first, then a set of `bugB`. A clear of `bugA` returns the active `bugB` view. It does not clear `bugA` or report on `bugA`):

```json
{"id":"3","method":"agent.view.clear","params":{"source":"bugA"}}
{"id":"3","result":{"type":"agent_view","active":true,"source":"bugB","label":"B"}}
```

`Checked 2026-10-17 against herdr 0.9.5`: these claims:

- the `bugA`/`bugB` sequence, the `viewX`/`viewY` sequence, and the sequence with a `source` that was never set
- `source: null` and no `source` (both `active: false`), and a clear with no active view
- the 121-character `source`, the CLI, and the two examples

## agent.view.set

Sets a saved agent view for `source` and makes it active. `source` is an identifier of a client or a view. The view has an optional `filter` and an optional `sort` order on the live agent list, and an optional display `label`.

- The view controls how a client shows and orders agents. It does not change the agents.
- A view that one connection sets is visible to a `set` or `clear` from all other connections to the same server. Thus, herdr does not key views by connection.
- herdr also does not keep a separate view for each `source`. The bug note of [agent.view.clear](#agentviewclear) shows one live view slot for the full server.
- The last `set` call, for any `source`, replaces the view in that slot.
- The result of `agent.view.set` only returns the `source` and `label` of the request. This does not prove that views for different sources exist together.
- No `agent.list` or other agent result shows the active view. Thus, a caller cannot read the filter, sort, label, or source of the live view. Only a `clear` that targets that view shows it.

**Params** (`AgentViewSetParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `source` | string | yes | — | The view or client identifier of this view. It must not be empty. It must have a maximum of 120 characters, and contain only ASCII letters, digits, `:`, `.`, `_`, or `-`. Otherwise herdr gives `invalid_agent_view`. |
| `filter` | AgentViewFilter \| null | no | null | Condition that selects the agents in the view. With `null`, the view shows all agents. |
| `sort` | array<AgentViewSort> | no | `[]` | Ordered sort keys for the filtered agents. |
| `label` | string \| null | no | null | Label of the view for persons. |

**`AgentViewFilter`** — one of these shapes. `op` identifies the shape:

| `op` | shape | meaning |
| --- | --- | --- |
| `all` | `{op, filters: AgentViewFilter[]}` | Logical AND of the subfilters. |
| `any` | `{op, filters: AgentViewFilter[]}` | Logical OR of the subfilters. |
| `not` | `{op, filter: AgentViewFilter}` | Negation of a subfilter. |
| `eq` | `{op, field: AgentViewField, value: AgentViewValue}` | The field is equal to the value. |
| `in` | `{op, field: AgentViewField, values: AgentViewValue[]}` | The field is one of the values. |
| `exists` | `{op, field: AgentViewField}` | The field is present. |

**`AgentViewField`** — a builtin field name (enum: `status`, `workspace_id`, `tab_id`, `pane_id`, `agent`, `seen`, `state_change_seq`), or a reference object `{"token": "<name>"}` for a custom token. The name must be a key of the [AgentInfo](#agentinfo) `tokens`.

**`AgentViewValue`** — one of `string`, `boolean`, a `uint64` integer, or a context object `{"context": "current_workspace_id" | "current_tab_id"}`. herdr replaces the context object with the current workspace or tab of the caller when it evaluates the filter.

**`AgentViewSort`**:

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `field` | AgentViewSortField | yes | — | Sort key. |
| `order` | enum | no | `asc` | `asc` or `desc`. |

**`AgentViewSortField`** — a builtin sort field (enum: `workspace_order`, `tab_order`, `pane_order`, `attention`, `status`, `agent`, `seen`, `state_change_seq`), or a reference object `{"token": "<name>"}` for a custom token.

**Result** — `type: "agent_view"`:

| field | type | required | meaning |
| --- | --- | --- | --- |
| `type` | const `"agent_view"` | yes | Result discriminator. |
| `active` | boolean | yes | True when the view is active. It is `true` after a set. |
| `source` | string \| null | no | The `source` of the request. |
| `label` | string \| null | no | The `label` of the request. |

**Errors**:

- `invalid_request` for a filter or sort with a bad shape:
  - an unknown `op`: `` "invalid request: unknown variant `xor`, expected one of `all`, `any`, `not`, `eq`, `in`, `exists` at line 1 column 109" ``
  - an unknown `AgentViewField`: `"invalid request: data did not match any variant of untagged enum AgentViewField at line 1 column 115"`
  - a bad sort `order`: `` "invalid request: unknown variant `sideways`, expected `asc` or `desc` at line 1 column 121" ``
- `invalid_agent_view` for a `source` that fails the checks above.
- Other codes are possible.

**CLI**: API-only (no CLI subcommand).

**Example**:

```json
{"id":"1","method":"agent.view.set","params":{"source":"api-docs","filter":null,"label":"build"}}
{"id":"1","result":{"type":"agent_view","active":true,"source":"api-docs","label":"build"}}
```

`Checked 2026-10-17 against herdr 0.9.5`: these claims:

- this example, and the view across connections
- an empty `source`, a `source` with 121 characters, and `"bad source"` (`invalid_agent_view`), and a `source` with 120 characters (accepted)
- a filter with each `op`, the builtin fields `status`, `agent`, `workspace_id`, and `seen`, `state_change_seq`, a token field, and a context value
- the sorts `attention` with `desc`, a token, and `workspace_order`
- the three `invalid_request` errors

`Derived from schema; not live-checked (2026-10-17, herdr 0.9.5)`: the other enum values of the fields and sort fields.

## agent.wait

Waits until the target agent has one of the requested statuses.

- Without `until`, the call matches the first settled `idle`, `done`, or `blocked` status. This is the same default set as the wait of `agent.prompt`.
- Give `until` to match specific statuses. Give `unknown` explicitly when you need it.
- Without `timeout_ms`, the call waits with no limit.
- The wait is level-triggered. When the agent already has a matching status, the call returns immediately. On an agent that was already `idle`, herdr 0.9.5 returned in 0.10–0.14 ms. herdr 0.9.1 returned in 0.2 ms, and herdr 0.9.3 returned in 1.5 ms.
- A suspended agent keeps its old status. Thus, `agent.wait` returns immediately for it (refer to [Agent identity across Ctrl+Z](#agent-identity-across-ctrlz)).
- `idle` and `done` do not always mean that the agent is ready for input. [../data-model.md](../data-model.md#agentstatus) gives this caveat.

The result has a full [AgentInfo](#agentinfo), with the same shape as the result of `agent.get`. It is not an event envelope. The schema gives the result type `wait_matched` with an `event` field. The response on herdr 0.9.1, 0.9.3, and 0.9.5 is `type: "agent_info"` with an `agent` field. The 0.9.5 schema still gives `wait_matched`.

**Params** (`AgentWaitParams`):

| field | type | required | default | meaning |
| --- | --- | --- | --- | --- |
| `target` | string | yes | — | Live agent name or the ID of the pane that holds the agent. |
| `until` | array<AgentStatus> | no | `[]` | Statuses to match. An empty array means the default settled set (`idle`, `done`, `blocked`). Values: `idle`, `working`, `blocked`, `done`, `unknown`. |
| `timeout_ms` | uint64 \| null | no | null | The call fails after this number of milliseconds. With `null`, the wait has no limit. |

**Result** — `type: "agent_info"`. A pass saw this response. The schema gives `wait_matched` with an `event: EventEnvelope` field (refer to the note above):

| field | type | required | meaning |
| --- | --- | --- | --- |
| `type` | const `"agent_info"` | yes | Result discriminator. |
| `agent` | [AgentInfo](#agentinfo) | yes | The agent with its matched status. |

**Errors**:

- `timeout`: no matching status within `timeout_ms` (`"timed out waiting for agent status"`). For `timeout_ms: 500`, the error came after 500.5 ms on herdr 0.9.5, and after 509 ms on herdr 0.9.3.
- `agent_not_found`.
- `agent_not_running` (`"agent is no longer running in the target pane"`): the agent process exits while the call waits. On herdr 0.9.5, the error came about 301 ms after the agent exited. On herdr 0.9.1, it came 2502 ms into a wait of 30 s. On herdr 0.9.3, it came about 0.4 s after the agent exited.
- Other codes are possible.

**CLI**: `herdr agent wait <TARGET> [--until <STATUS>]... [--timeout <MS>]`

**Example** (the call came right after `agent.start` and returned after 3502 ms):

```json
{"id":"1","method":"agent.wait","params":{"target":"reviewer","timeout_ms":15000}}
{"id":"1","result":{"type":"agent_info","agent":{"terminal_id":"term_65e0dbacd49af3","name":"reviewer","agent":"opencode","terminal_title":"OpenCode","terminal_title_stripped":"OpenCode","agent_status":"idle","workspace_id":"w2","tab_id":"w2:t1","pane_id":"w2:p1","focused":false,"interactive_ready":true,"state_change_seq":8,"cwd":"/home/user/project","foreground_cwd":"/home/user/project","revision":2}}}
```

`Checked 2026-10-17 against herdr 0.9.5`: these claims:

- the success right after `agent.start`, and the immediate return for a matching status
- `until: ["working"]`, and `timeout` after 500.5 ms for `timeout_ms: 500`
- `blocked` on a launch at the trust dialog, `agent_not_running`, and `agent_not_found`
- the immediate return for a suspended agent, and this example

A pass did not check the wait without `timeout_ms`, because a wait with no limit blocks the pass. The `herdr agent wait --help` text gives this behavior.
