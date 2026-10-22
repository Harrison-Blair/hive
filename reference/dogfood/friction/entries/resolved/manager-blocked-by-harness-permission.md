---
title: "A manager stops on a harness permission block and the coordinator does not know"
status: "Resolved 2026-10-19"
opened: 2026-10-19
decision: "Fixed in 29-message-marker (5c8f6846): Claude Code and Codex spawns launch with the bypass flag, each hive message starts with the marker, and a manager sends a blocked record before it stops."
---

**Summary:**
On 2026-10-19, the coordinator sent `manager-stenographer` (Claude Code, auto
mode) two inbox records. Record 3 relayed a user statement: the real-machine
acceptance of dev is verified. Record 4 asked for a read-only architecture
report. The auto-mode permission check blocked the memory write for record 3
as instruction poisoning, because the user statement came only through the
inbox. It also blocked a read-only `git ls-files` for record 4. The manager
asked the user two questions in its own pane and ended its turn. It sent no
inbox record to the coordinator. herdr showed the pane as `done`, and the
supervisor sends no nudge for a `quiet` record. Thus the coordinator saw no
stop. The user found the stop by hand in the manager tab.

Two causes:
1. The harness permission check treats an inbox record as untrusted input. A
   user statement or an approved goal that the coordinator relays does not
   authorize an action for the harness.
2. The manager instructions do not tell a manager to send a record to the
   coordinator when it stops for a user decision.

**Resolution 2026-10-19:**
Fixed in 29-message-marker, landed at 5c8f6846. Hive starts each spawned Claude
Code agent with `--dangerously-skip-permissions` and each Codex agent with
`--dangerously-bypass-approvals-and-sandbox`. The user approved these flags
directly. Each message that hive types into an agent pane (first prompt,
`hive agent send`, supervisor nudge) starts with `⬢ `. The role instructions
say that a `⬢` message came from hive, that a message without it came from the
user, and that the marker does not widen a grant. The manager skill and the
project template tell a manager to send the coordinator one record before it
stops for a user decision or on a harness block.

The fix was used before it landed: the auto-mode check of `manager-hive`
blocked the spawn of this task's worker with `[Create Unsafe Agents]`. The
manager sent the coordinator a blocked record and stopped, and the user
approved in the manager pane.

**Reproduction steps:**
1. Spawn a Claude Code manager in auto mode.
2. Send it `hive inbox send --to manager-ID TEXT` with a user statement that
   it must record, or with a report goal.
3. Read the manager pane. The permission check blocks the action, and the
   manager asks the user in its pane.
4. Run `hive inbox --role coordinator`. No record from the manager is present.
