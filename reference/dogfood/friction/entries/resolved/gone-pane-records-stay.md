---
title: "The record of a closed pane stays until a stop"
description: "When a pane closes without hive agent stop, its agent record stays, and each agent list and each graph refresh queries the gone pane again"
status: "Resolved 2026-11-02"
opened: 2026-11-02
decision: "hive agent list, and so each graph refresh, deletes the record of a gone pane, and of a pane that runs no process of the record harness when the record is at least 60 s old, under the layout lock and only when the file still names the same pane"
---

**Issue:** When the pane of an agent closes without `hive agent stop`, for
example when the user closes the tab or herdr restarts, the record in
`<home>/agents/` stays. `hive agent list` shows it with status `gone`, and the
graph shows it as a dimmed card. `list.Run` sends `pane.get` for it on each
call, and herdr answers `pane_not_found`, which the herdr client logs as a
`herdr error` (`internal/agent/list/list.go`). The graph then also runs `git`
in the start directory of the record. Only `hive agent stop NAME`, bare `hive`
for another coordinator record of its own pane
(`removeOldCoordinators` in `internal/coordinator/name.go`), and a spawn or a
bare `hive` that takes the same name remove or replace such a record. No
command removes the records of gone panes in bulk, and nothing removes them
on its own.

**Summary:** In `~/.hive/logs`, 2613 of the 2636 `herdr error` events from
2026-10-28 to 2026-11-02 are `pane_not_found`, nearly all from open graphs
that queried the same few gone panes each second:

- Run `r-4fb03839` on 2026-11-01: 1804 errors for pane `w15E:p4`, the pane of
  the record `ror-desktop-coordinator`, from 21:49:46 to 22:19:50.
- Runs `r-a2ef0d16` and `r-1a6dd095` on 2026-11-01: 694 errors for pane
  `w16A:p1` and 35 for `w16B:p1`.
- Run `r-f2149f0a` on 2026-11-02: 20 errors each for `w176:p2` and `w15E:p4`.

Each gone record adds one `herdr error` and 2 `git` events each second while
a graph is open (see the entry `graph-log-volume`). A record that stays also
holds its name, so a later bare `hive` or spawn sees it, though both replace a
record whose pane is gone.

**Reproduction steps:**
1. Use a scratch `HIVE_HOME` and spawn one agent.
2. Close its pane in herdr, not with `hive agent stop`.
3. Run `hive agent list`. The record shows with status `gone`.
4. Run `hive agent list` again, or open `hive agent graph`. Each run logs a
   `herdr error` with `pane_not_found` for the same pane, and the record stays.
