---
title: "A brief tells a worker to read its inbox, but a worker has no inbox"
status: "Open"
opened: 2026-10-20
decision: ""
---

**Summary:**
The Finish section of each wave D brief says "Read only your own inbox." The
worker skill does not name an inbox command. A worker has no inbox: on
2026-10-20 t91-worker-1 ran `hive inbox --json --agent t91-worker-1 --all`
and got `role has no inbox: worker`. The worker then sent its completion
report to the coordinator, not to the manager. The manager gets messages to
a worker with `hive agent send`, so the brief line is wrong. Hive has no rule
that stops a worker from sending to the coordinator.

**Reproduction steps:**
1. As a worker, run `hive inbox --agent NAME --all`: `role has no inbox:
   worker`.
2. Run `hive inbox send --to coordinator TEXT --agent NAME`: the record is
   created, though the brief names the manager as the receiver.
