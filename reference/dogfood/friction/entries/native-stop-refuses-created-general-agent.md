---
title: "Native stop refuses a general agent that Hive created"
status: "Open"
opened: 2026-10-26
decision: "Preserve the refusal. Check actual pane ownership and current identity before the authorized manual closure. No product fix."
---

**Issue:** `hive agent stop` refuses an observer that the same manager created with native `hive agent spawn`.

**Status:** Open.

**Summary:**
The manager created `ci185-observer-1` with the general profile to read the
post-landing CI results for task 185. The native spawn returned created,
started, prompted, and live effects. The observer reported both exact-head
CI runs successful and stayed idle for its parent to stop it.

The native stop command failed with exit 1:

```text
hive: agent ci185-observer-1: hive did not start it: exit it by hand
```

The observed record had `spawned: true`, pane `w0J:p8`, terminal
`term_65eb7fa1f9366297`, and Pi PID `2342261` with process start `31455706`.
This entry does not infer the cause or an actor that changed a record.

Coordinator record 223 authorized the manual closure that the error suggests.
The manager used the `herdr-use` skill and Herdr 0.9.5, protocol 22.
The context gate passed. Explicit pane and process reads matched the recorded
pane, terminal, working directory, Pi PID, process start, and completed state.
The caller remained the manager in `w0J:p5`.

Only after those checks, the manager closed its own observer pane with
`herdr pane close w0J:p8`. A fresh pane list showed that target absent and
the manager pane still present. The manager changed no ownership flag or
state JSON. Normal supervisor reconciliation owns the later record update.
The manager used no force stop, source fix, server restart, or new user decision.

**Reproduction steps:**
1. Use native Hive spawn to create a general observer with no task fact.
2. Record the returned agent, pane, terminal, process, and creation evidence.
3. Let the observer report its read-only result and become idle.
4. Run native `hive agent stop` for that observer.
5. Observe the refusal above. Preserve the actual record and error.

**Workaround and limit:**
Close only a pane that the task actually created and that still contains the
same process. Stop if its identity changed. Do not close another occupant or
change flags to satisfy the guard. This entry records one observed failure.
It requests no new product feature or guard.

Task 185 evidence contains `manager-ci-observer-stop-refusal.json`,
`manager-ci-green.json`, and `observer-manual-cleanup/` with all raw identity,
version, close, and post-close observations. The general agent read only
GitHub results and wrote no source, PR, runtime, or settings change.
