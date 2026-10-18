# T15 supervise

| Item | Value |
| --- | --- |
| Task ID | T15-supervise |
| Wave | 4 |
| Tier | strong |
| Depends on | T04, T10, T11, CP2 |
| Blocks | CP3, T20, T21 |
| Branch | `task/T15-supervise` |
| Worktree | `projects/hive/worktrees/T15-supervise/` |
| Evidence | `evidence/T15-supervise/` |

## Overview

`hive supervise` is the one relay process of a home. It holds the supervisor
claim, subscribes to herdr events, and watches the store. It turns events into
inbox records, stops a finished agent when it is idle, and nudges an idle role
that has unread records. No role runs a wait loop. The supervisor never polls
the store and never dispatches work.

The supervisor evaluates its rules on three triggers: its start, a change in
the store, and a herdr event. A finished mark that arrives while the agent is
already idle still leads to a stop. The supervisor identifies each agent by
its pane and its `terminal_id`, never by its herdr name.

## User flows

1. The coordinator runs `hive supervise` in a pane of the hive session. It prints `supervisor live, generation 3` and stays in the foreground.
2. A worker runs `hive task complete`. The supervisor sees the new inbox record and the finished mark. When the worker is idle, the supervisor stops it with the stop function of T10. The manager is idle, so one line arrives in its pane: `hive: 1 new inbox record, run: hive --home /home/user/hive inbox`.
3. A worker shows a permission dialog. herdr reports `blocked`. The supervisor writes a `blocked` record to the manager and nudges it.
4. A worker is idle for ten minutes without a completion. The supervisor writes one `quiet` record. It writes no second record for the same idle period.
5. The supervisor starts after a crash. It compares each live record with its pane and marks a record whose pane holds no agent `exited`, or `unresolved` when the pane is gone.
6. The user moves the pane of a worker to another tab. The supervisor updates the record, keeps the old pane ID in `pane_history`, and subscribes to the new pane.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `cmd/root.go` | Register the `supervise` command. |
| `cmd/supervise/supervise.go` | New. `hive supervise [--json]`. |
| `internal/supervise/doc.go` | New. Package summary and file map. |
| `internal/supervise/supervise.go` | New. The claim, the start, the main loop over the three triggers, the exit. |
| `internal/supervise/subscribe.go` | New. The selector list from the watched panes, the rebuild on a change, and the reconcile across the gap. |
| `internal/supervise/events.go` | New. The event handlers of the table in AC2. |
| `internal/supervise/watch.go` | New. The inotify watch of the inbox and agent directories. |
| `internal/supervise/quiet.go` | New. The quiet timers with an injectable clock. |
| `internal/supervise/nudge.go` | New. The nudge rule and the prompt line. |
| `internal/supervise/reconcile.go` | New. The reconcile by pane and by claim. |
| `internal/lib/herdr/events.go` | New. `Subscribe` with a long connection, a line cap, and a channel of events. |
| `internal/lib/herdr/doc.go` | Add the map line. |
| `internal/lib/records/claim.go` | Add `pid_start`, the start time of the process from `/proc/<pid>/stat`. |
| `internal/lib/records/inbox.go` | Add `nudged_generation`. |
| `internal/lib/process/process.go` | New. `Alive(pid, start)` reads `/proc/<pid>/stat` and compares the start time. `doctor` and `supervise` share it. |
| `internal/doctor/checks.go` | Change the supervisor check: the claim exists and its `pid` with its `pid_start` is a running process. |

## Requirements

### AC1 Singleton

- The start takes the store lock, reads the `supervisor` claim, and checks it with `process.Alive(pid, pid_start)`. A live claim fails the start with `supervisor live: pid <n>`. Otherwise the start registers the role `supervisor` with the own pid and start time in the same mutation, which increments the generation. Two starts at the same time give exactly one supervisor.
- A claim whose pid names a process with another start time is stale. The pid belongs to a new process.
- The process stays in the foreground. On `SIGINT` or `SIGTERM` it closes the subscription and exits with 0. When the subscription connection ends, it exits with 1. Nothing restarts it in a loop.
- With `--json` the process prints one envelope at exit, with the outcome and the exit reason. Without `--json` it prints one line at the start and one line at each record it writes.

### AC2 Events to records

| Input | Condition | Action |
| --- | --- | --- |
| Status changed to `blocked` | The pane holds a `live` record with another stored status | Append `blocked` to the controlling role. Store the status. |
| Status changed to `idle` or `done` | The record has `finished` true | Run the stop function of T10. The record becomes `stopped`. |
| Status changed to `idle` or `done` | `finished` false, and no `completed` record for the task since the last `working` | Start the quiet timer of the agent. Store the status. |
| Status changed to `working` | A quiet timer runs | Cancel the timer. Store the status. |
| Quiet timer ends | The stored status is still `idle` or `done` | Append `quiet` to the controlling role. Once for each idle period. |
| `pane_agent_detected` with `released` true and `final_status` `unknown` | The pane holds a `live` record | Set the record `exited`. Append `exited` to the controlling role. |
| `pane_exited`, `pane_closed`, `tab_closed`, or `workspace_closed` | The closed pane, tab, or workspace holds a `live` record | Set the record `unresolved`. Append `unresolved` to the controlling role. |
| `pane_moved` | `previous_pane_id` is the pane of a `live` record | Set `pane_id`, `workspace_id`, and `tab_id` from `pane`. Append the old pane ID to `pane_history`. Rebuild the subscription. |
| A status push with the stored status | Any | Log only. herdr pushes the event also on a metadata change. |

- `done` means idle in a tab that nobody saw yet. Each rule treats it as `idle`.
- The controlling role of a worker or a verifier is `manager-<project>`. The controlling role of a manager is `coordinator`. A coordinator or a supervisor record has no controlling role, so the supervisor logs the event only.
- The quiet period is `quiet_after_min` of the record, else 10 minutes. The supervisor has no flag for it.
- A long `working` status is never an event.
- An agent that exits to its shell gives `pane_agent_detected` with `released` true and no `pane_exited`. Only that evidence gives `exited`. A closed pane or workspace with a live record gives `unresolved`, because the supervisor cannot tell an exit from a loss.
- The supervisor also evaluates the stop rule and the nudge rule at its start and on each change in the store. A `finished` mark on a record with the stored status `idle` or `done` leads to the stop without a new event.
- The subscription is one `events.subscribe` request with one `{"type":"pane.agent_status_changed","pane_id":"<pane>"}` entry for each watched pane, and the plain entries `pane.agent_detected`, `pane.exited`, `pane.closed`, `tab.closed`, `workspace.closed`, and `pane.moved`. The watched panes are the panes of the `live` records. The envelope of a status change is `{"event":"pane.agent_status_changed","pane_id":...,"agent_status":...}`. The other events use the plain envelope with `data.type`.
- herdr has no method that changes a subscription. When the watched set changes, the supervisor opens a new connection with the new list, then closes the old one. A new `live` record, a record that leaves `live`, a `pane_moved`, or a `pane_id` change through the store, such as `hive agent resolve` of T20, changes the set. After each rebuild the supervisor calls `agent.get` on each watched pane and applies the status as a push, so no change in the gap is lost.
- The subscription connection has no deadline. Each other request uses the default deadline of the client of T05. The stop function of T10 sets its own with `herdr.WithDeadline`.
- Each event and each rebuild is a log event.

### AC3 Nudge

- The supervisor learns about a new inbox record from the inotify watch of `.hive/state/inbox/`, including new role directories, within one second. The watch of `.hive/state/agents/` gives the record changes.
- An unread record needs a nudge when its `nudged_at` is empty. It also needs one when its `nudged_generation` differs from the `generation` of the live record of its role. A restarted role gets the nudge again.
- The agent of that role must have the stored status `idle` or `done`. Then the supervisor sends one `agent.prompt` with `target` equal to the pane ID, one line, and no wait. The line is `hive: <n> new inbox record(s), run: hive --home <home> inbox`. It sets `nudged_at` and `nudged_generation` on each record.
- When the agent is `working` or `blocked`, the supervisor waits and nudges at its next transition to `idle` or `done`.
- When the role has no live agent, the records stay unread without a nudge.
- A failed prompt, such as `agent_prompt_stalled`, `agent_blocked`, or `agent_not_ready`, is a log event. The records stay without `nudged_at`. The next transition to `idle` or `done` retries.

### AC4 Reconcile

- At the start and after each rebuild, the supervisor reconciles each record with the status `live`, except the own supervisor record.
- For a record with a pane, it calls `agent.get` with the pane ID. A reply with the `terminal_id` of the record gives the stored status. A record without `terminal_id`, such as a registration by hand, takes the `terminal_id` of the reply. A reply with another `terminal_id` makes the record `unresolved`, because another agent holds the pane.
- On `agent_not_found` it calls `pane.get`. A pane with the status `unknown`, no `agent`, and no `restore_error` makes the record `exited`. A pane that does not exist, or a pane with `restore_error`, makes the record `unresolved`.
- A record of the role `supervisor` with a `generation` other than the own claim becomes `stopped`. The supervisor makes no herdr call for it, because the claim was free.
- Each `exited` or `unresolved` result appends its kind to the controlling role. A live herdr agent without a record is a log event.
- The supervisor then starts the quiet timers for the settled agents and evaluates the stop rule and the nudge rule.

### AC5 Doctor

- The supervisor check of `hive doctor` prints `ok` when the claim exists and `process.Alive(pid, pid_start)` is true, else `warn` with `supervisor down`.

### AC6 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Exactly one supervisor runs for a home, and the process exits as the requirements give. | V1, V6 |
| AC2 | Each row of the event table produces its record, its stop, or its rebuild, and only once for each period. | V2, V5, V6 |
| AC3 | A nudge goes to a settled role once for each generation, waits for a working role, and retries after a failure. | V3, V5, V6 |
| AC4 | The reconcile marks each record by the state of its pane and the own claim. | V4 |
| AC5 | `doctor` shows the live state of the supervisor. | V6 |
| AC6 | The gate passes. | V7 |

## Verification items

### V1 Singleton tests

- Proves: AC1.
- Command: `go test -race ./internal/supervise/ -run 'TestSingleton|TestExit|TestJSON' -v`, red and green.
- Expected: a claim with the pid and the start time of the test process blocks a second start. A claim with a dead pid, or with the own pid and another start time, lets the start take generation plus one. Twenty starts in goroutines on one home give one claim and nineteen `supervisor live` errors. A `SIGINT` and a `SIGTERM` to the own process each make the run return without an error after the subscription closed. A closed subscription makes the run return an error. With `--json` the output is one envelope.

### V2 Event tests

- Proves: AC2.
- Command: `go test -race ./internal/supervise/ -run 'TestEvents|TestSubscription' -v`, red and green.
- Expected: a fake herdr server pushes the literal frames of `reference/herdr/events.md` over a subscription. A fake clock drives the quiet timer. Each row of the table has a case, and the once-only rules have a case with a repeated event. Cases: `done` as idle, a repeated push with the stored status, and `pane_agent_detected` with `released` true. More cases: `workspace_closed` over a live record, and `pane_moved` with the record update and `pane_history`. The subscribe request has one `pane.agent_status_changed` entry for each live pane and the six plain entries. A new live record, a `pane_moved`, and a `pane_id` change in a record file each open a new connection. Each of these closes the old connection and calls `agent.get` for each watched pane. A `finished` mark written to a record with the stored status `idle` leads to the stop without an event.

### V3 Nudge tests

- Proves: AC3.
- Command: `go test -race ./internal/supervise/ -run TestNudge -v`, red and green.
- Expected: an appended record and an idle manager give one prompt with the line and the pane as the target. A `done` manager gets the prompt. A working manager and a blocked manager give no prompt until the idle event. A new generation of the manager gets the prompt again for the same record. A scripted `agent_prompt_stalled` and a scripted `agent_blocked` leave `nudged_at` empty, and the next idle event retries. A role without a live agent gets no prompt.

### V4 Reconcile tests

- Proves: AC4.
- Command: `go test ./internal/supervise/ -run TestReconcile -v`, red and green.
- Expected: a live record with the `terminal_id` of the fake reply keeps `live` with the status stored. A record with another `terminal_id` becomes `unresolved`. A record with `agent_not_found` and a shell pane becomes `exited`. A record with `agent_not_found` and a missing pane becomes `unresolved`. A coordinator record from a registration by hand without `terminal_id` takes the `terminal_id` of the reply. An old supervisor record becomes `stopped` without a herdr call. Each `exited` and `unresolved` result has an inbox record. An agent of the fake list without a record gives a log event.

### V5 Watch tests

- Proves: AC2, AC3.
- Command: `go test -race ./internal/supervise/ -run TestWatch -v`, red and green.
- Expected: the test uses the real inotify watch on a temporary home. A record written into a new role directory under `.hive/state/inbox/` reaches the nudge rule within one second. A change of `finished` in an agent record reaches the stop rule within one second. The test has no sleep longer than the bound.

### V6 Live transcript

- Proves: AC1, AC2, AC3, AC5.
- Command, from a pane of the hive session, with two new tabs in the own workspace of the worker:

```sh
go build .
rm -rf .hive/state .hive/tmp/repo-sup projects/repo-sup .hive/tmp/supervise.out
git init -q .hive/tmp/repo-sup && git -C .hive/tmp/repo-sup commit -q --allow-empty -m init
./hive --home "$PWD" project register .hive/tmp/repo-sup
./hive --home "$PWD" task create --project repo-sup --title "supervise demo"
herdr tab create --workspace "$HERDR_WORKSPACE_ID"
M_PANE=<root pane ID from the output>
herdr tab create --workspace "$HERDR_WORKSPACE_ID"
W_PANE=<root pane ID from the output>
M="t15-check-m-$(date +%s)"
W="t15-check-w-$(date +%s)"
./hive --home "$PWD" agent spawn --pane "$M_PANE" --role manager --project repo-sup --name "$M" --harness claude --model <small model>
./hive --home "$PWD" agent spawn --pane "$W_PANE" --role worker --project repo-sup --task 1-supervise-demo --name "$W" --harness claude --model <small model>
./hive --home "$PWD" task assign --task 1-supervise-demo --agent "$W"
./hive --home "$PWD" supervise > .hive/tmp/supervise.out 2>&1 &
SUP=$!
sleep 2
./hive --home "$PWD" doctor
./hive --home "$PWD" agent send --agent "$W" --wait --timeout 120 "Run this command and then stop: hive --home '$PWD' task complete --task 1-supervise-demo"
sleep 5
./hive --home "$PWD" agent list --all
./hive --home "$PWD" inbox --role manager-repo-sup
herdr pane read "$M_PANE" --lines 20
./hive --home "$PWD" agent stop --agent "$M"
kill -INT "$SUP"
wait "$SUP"
echo "exit $?"
cat .hive/tmp/supervise.out
./hive --home "$PWD" doctor
herdr tab close <the first new tab>
herdr tab close <the second new tab>
rm -rf .hive/tmp/repo-sup projects/repo-sup .hive/tmp/supervise.out
```

- Expected: the first `doctor` shows `ok supervisor`. The agent list shows the worker `stopped` and the manager `live`. The inbox shows one `completed` record with `nudged_at` set. The pane read of the manager shows the nudge line. The supervisor exits with 0 after the signal, and its output shows the stop and the nudge. The second `doctor` shows `warn supervisor down`. Both tabs are gone at the end.

### V7 Gate

- Proves: AC6.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- `hive agent resolve` and the guards on an unresolved record. T20 adds them.
- A restart loop. The coordinator starts the supervisor again at its next startup.
- A nudge to a role without an agent in herdr, such as a coordinator in a plain shell.

## Changes

None.
