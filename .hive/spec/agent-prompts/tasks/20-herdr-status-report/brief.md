# herdr status control report

| Item | Value |
| --- | --- |
| Wave | report only |
| Tier | decent |
| Depends on | none |
| Grant | none. No code change, no commit, no landing. |
| Output | `/home/penguin/source/hive/.hive/tmp/reports/herdr-status-control.md` |

## Overview

This task is an investigation. It answers one question: can hive change or
clear agent statuses in herdr from a program? Do not change a source file. Do
not commit. Do not push. When the facts do not agree with this brief, report
it with `hive inbox send --to manager-hive TEXT --agent NAME`.

This brief replaces the commit, push, gate, and evidence steps of the worker
skill. The report file is the only output.

## Scope

- The statuses that herdr shows for a pane, a tab, or a workspace: the agent
  status (for example `idle`, `working`, `done`, `blocked`), the unread or
  attention markers, and each status text or label.
- For each status, record who sets it (herdr detection, the harness, or an API call).
  Record whether a CLI command or a socket API method can set or clear it.
  Record whether the change persists or herdr overwrites it.

## Sources

- `reference/herdr/`, the API authority for herdr 0.9.5, protocol 22. Read
  `reference/herdr/README.md` first.
- `herdr --help` and `herdr SUBCOMMAND --help` for each subcommand.
- Live tests.

## Rules

- Do live tests only in a scratch workspace that you create with
  `herdr workspace create --label status-scratch --no-focus`. Do not change
  the status of a pane, tab, or workspace that hive or the user started.
- Close the scratch workspace after the investigation ends. Make sure that
  `herdr workspace list` does not show it.
- Read only your own inbox.

## Report

Write the report to the output path. It contains:

1. Each command that you ran and its output. Cap each output at 40 lines.
2. A table with one row for each status: the status, who sets it, `can set`,
   `can clear`, and `persists` (yes, no, or unknown), and the command or API
   method.
3. What hive could use, for example a status clear after the agent reads a
   nudge.

## Finish

Run `hive task complete`. Send the report path and a three-line summary to the
manager with `hive inbox send --to manager-hive TEXT --agent NAME`.
