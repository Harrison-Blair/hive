---
title: "The supervisor keeps running an old hive binary after a rebuild"
status: "Resolved 2026-10-21"
opened: 2026-10-19
decision: "Task 3 adds the r restart for the supervisor and an old-binary mark in the TUI header"
---

**Summary:**
On 2026-10-19, the supervisor (pid 1718014) started at 13:22 from the hive
binary of 72a33b5b. The managers rebuilt `~/go/bin/hive` after waves 2, 3,
and 4, last at 15:41. The supervisor kept the old code. Thus the wave 2
change "a quiet record never nudges" was not live: at 16:11 the supervisor
nudged the coordinator for one `quiet` record. `hive doctor` reported the
supervisor as live and did not compare its binary with the installed one. The
workaround was to stop the process and run `hive supervise` again in the logs
tab. No skill tells an agent to restart the supervisor after a rebuild.

**Reproduction steps:**
1. Start `hive supervise`.
2. Land a change to the supervisor and run `go install` or `go build` for hive.
3. Run `hive doctor`. It reports the old supervisor as live.
4. Trigger the changed behavior. The supervisor uses the old code.
