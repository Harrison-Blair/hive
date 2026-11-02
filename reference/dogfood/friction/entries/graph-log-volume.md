---
title: "An open hive agent graph fills the daily log"
description: "The graph refreshes each second and logs each herdr call, each git call, and each list result, so one open graph writes most of the daily log"
status: "Open"
opened: 2026-11-02
decision: ""
---

**Issue:** `hive agent graph` refreshes each second (`refreshInterval` in
`internal/agent/graph/model.go`). Each refresh calls `list.Run` with the
command logger (`internal/agent/graph/run.go`). `list.Run` sends one
`pane.get` per agent record, and the herdr client logs a `herdr request` and a
`herdr reply` or `herdr error` for each call. `list.Run` then logs one
`agent list` event. After each fetch, the graph runs `git` in the start
directory of each agent: `rev-parse --is-inside-work-tree`, then
`symbolic-ref`, and `rev-parse --short HEAD` for a detached checkout
(`internal/agent/graph/checkout.go`). Each `git` call logs one event. So an
open graph writes 1 event plus 4 or 5 events per agent each second, all at
level debug, while nothing changes. Hive never deletes a log file, and
`hive logs` and `hive logs --follow` read the whole daily file.

**Summary:** The reviewer `agent-graph-review1` of the graph pull request
predicted this on 2026-10-31 00:05: "~75x the normal daily log volume … roughly
300 MB in one day's file" for 8 agents and 8 hours. The coordinator sent only
the card finding of that review to the fix. The logs of 2026-11-01 and
2026-11-02 in `~/.hive/logs` show the effect with 1 or 2 agent records:

- 2026-11-01: six graph runs wrote 63358 of the 66197 events (96%), about
  29 MB of the 30.8 MB file. Run `r-1a6dd095` (08:48 to 10:37) wrote 33780
  events. Run `r-4fb03839` (21:49:22 to 22:19:50) wrote 11931 events: 1829
  `agent list`, 2075 `herdr request`, 2075 `herdr reply`, 1805 `herdr error`,
  and 4146 `git` events. Two graph runs were open at the same time from
  08:48 to 08:57.
- 2026-11-02, until 14:22: two graph runs wrote 42594 of the 43493 events,
  about 17 MB.
- Before the graph, the daily files were 0.1 MB to 4.4 MB.

A record whose pane is gone adds one `herdr error` each second (see the
entry `gone-pane-records-stay`). Most of the volume
comes from healthy records: run `r-f2149f0a` on 2026-11-02 wrote 37552 events
in 46 minutes with about 40 `herdr error` events.

**Reproduction steps:**
1. Use a scratch `HIVE_HOME` and spawn one agent.
2. Run `hive agent graph` and leave it open for 60 seconds. Quit with `q`.
3. Read the `request_id` of the `command start` event with command
   `agent graph`. Count the events of that run:
   `jq -c 'select(.request_id=="REQUEST_ID")' "$HIVE_HOME"/logs/*.jsonl | wc -l`.
4. The count is about 5 or 6 events for each second that the graph was open.
