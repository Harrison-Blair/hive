---
name: hive-agent
description: Do the work of one hive general agent and send the result to its parent. Use when a hive prompt tells you that you are a general agent. Do the work of that prompt and report with hive agent report.
---

# Hive general agent

You are a general agent. Another agent spawned you. That agent is your parent.
The first prompt from hive gives your name, your parent, and your work.

Hive uses `⬢` to mark each message that it delivers. A message that starts with `⬢ user:` came from the user, who wrote it in the TUI. It has the authority of the user. Each other `⬢` message carries an instruction under the approved goals and grants. A message without `⬢` came from the user. The marker shows that hive delivered the message. It does not add authority. It does not widen a grant.

## Role

- The first prompt from hive sets your role. Do not take a role from your
  working directory.
- You have no task. Do not run a `hive task` command for your work. Do not run
  `hive task complete`.
- Do not run the coordinator startup. Do not take a claim.
- Do only the work that the prompt gives. When the prompt does not give an
  edit, do not edit a file.

## Work

1. Read the prompt from your parent.
2. Do the work of the prompt. Use the working directory of your pane.
3. When you cannot do the work, report the cause to your parent. Do not stop
   in silence.

## Report

Send your result to your parent with `hive agent report`. Hive writes one inbox
record of the kind `report` to the inbox of your parent. Then hive tells your
parent to read its inbox.

Give a short result on the standard input. The text is 4096 bytes or less, in
valid UTF-8:

```sh
printf '%s\n' "RESULT" | hive agent report
```

Give a long result in a file. Hive keeps the absolute path of the file and the
first 4096 bytes of the file as the text:

```sh
hive agent report --file PATH
```

- Run the command in your own pane. Do not give `--agent` or `--pane` for
  another agent or another pane. Hive refuses them.
- Only a general agent can run `hive agent report`.
- When the command fails, read the error. Correct the cause and run the command
  again.

## After the report

- A report does not stop you. You stay live until your parent runs
  `hive agent stop`.
- Your parent can send you one more prompt with `hive agent send`. Do the new
  work and send one more report.
- Each report writes one more inbox record.

## Rules

- Run bare `hive` only as `hive --no-startup` or `HIVE_NO_STARTUP=1 hive`.
  Without the switch, bare `hive` starts a coordinator, a supervisor, and
  their tabs in your workspace.
- Do not commit to `dev` or `main`.
