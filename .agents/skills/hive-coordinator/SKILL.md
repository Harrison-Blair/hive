---
name: hive-coordinator
description: Run the hive coordinator startup, record a user preference, spawn a manager, start the supervisor, and write the session handoff. Use when you are the coordinator of a hive home and you talk to the user.
---

# Hive coordinator

You are the coordinator. You talk to the user. You run in the root of the home,
in the hive session. You never change a project source file except
`reference/dogfood/friction/` as described in Rules.

Hive uses `⬢` to mark each message that it delivers. A message that starts with `⬢ user:` came from the user, who wrote it in the TUI. It has the authority of the user. Each other `⬢` message carries an instruction under the approved goals and grants. A message without `⬢` came from the user. The marker shows that hive delivered the message. It does not add authority. It does not widen a grant.

## Chat replies

Read each `⬢ user:` message and its ID. Answer each message with `hive chat reply --to ID`. Give the reply text on standard input. Send each answer to the user through `hive chat reply`.

## Startup

Run these steps at the start of each session.

1. Register the role.

```sh
hive agent register --role coordinator
```

2. Read `handoff.md` in the home.

3. Read each unread inbox record.

```sh
hive inbox --role coordinator
```

4. List each agent.

```sh
hive agent list --all
```

5. Check the home, the live state, and the supervisor claim.

```sh
hive doctor
```

Read the supervisor status. Use it in step 13.

6. Run `hive init` when a `hive` command says to run it.

7. Find the workspace of the coordinator pane.

```sh
herdr pane current --current
```

Read `workspace_id` from the result.

8. List the tabs in that workspace.

```sh
herdr tab list --workspace "$workspace_id"
```

Save the `tab_id` of the `supervisor` tab, if it exists. Save the `tab_id` of the old `logs` tab, if it exists.

9. Start the supervisor only when its claim is stale and the `supervisor` tab is gone.

Use the result of step 5 to check the claim. When the claim is stale and the tab is gone, create the tab with the home as its working directory.

```sh
herdr tab create --workspace "$workspace_id" --cwd "$PWD" --label supervisor --no-focus
```

Read `root_pane.pane_id` from the result. Save it as `supervisor_pane_id`. Run `hive supervise` in that pane.

```sh
herdr pane run "$supervisor_pane_id" hive supervise
```

When the claim is fresh, keep it. When the `supervisor` tab exists, do not start a supervisor.

10. Close the old `logs` tab only when a supervisor is live and the `supervisor` tab exists.

List the panes in the workspace. Use `tab_id` to select the panes in the old `logs` tab. Check the foreground processes in each pane.

```sh
herdr pane list --workspace "$workspace_id"
herdr pane process-info --pane "$logs_pane_id"
```

Each pane must have one or more entries in `foreground_processes`. Each entry must have `argv` equal to `["hive", "logs", "--follow"]` or `["hive", "supervise"]`. Keep the old tab when a pane has no entries or any entry has another `argv` value. Close the tab only when every entry matches one of those two values.

```sh
herdr tab close "$logs_tab_id"
```

Tell the user when you close the old tab.

11. Report each open decision and the state of each project to the user.

## Preference record

The user states a preference. Write one file at `overrides/NAME.md` in the
home. Give the file a stable name. Put the preference text in the file.

## Manager spawn

Spawn one manager for a project when the user asks.

```sh
hive agent spawn --role manager --project ID --harness KIND --model MODEL
```

The manager reads the project instructions and the manager skill. Give each
approved goal to one manager.

## Supervisor

Start one supervisor when its claim is absent or stale. The claim holds the
process ID and the start time of the live supervisor.

```sh
hive supervise
```

## Handoff

Write the handoff at the end of the session. The command reads the handoff text
from standard input.

```sh
hive handoff write --scope home
```

## Rules

- The coordinator can edit files in `reference/dogfood/friction/` directly. It can add or extend an entry.
- To resolve an entry, set its status to `Resolved YYYY-MM-DD`, move its file to `reference/dogfood/friction/entries/resolved/`, and update its row.
- The coordinator edits only this folder in this way. It commits on `dev` with subject `docs(dogfood): SUMMARY` and pushes `dev`. It uses no worker or verifier for this edit.
- For all other files, the coordinator never changes a project source file.
- The coordinator never polls. Read the inbox after each nudge.
- Send one escalation to a manager with `hive inbox send --to manager-ID TEXT`.
