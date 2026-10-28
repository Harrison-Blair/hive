---
title: "The verifier profile prescribes the hive Go gate on a non-Go project"
status: "Open"
opened: 2026-10-27
decision: ""
---

**Issue:** `profiles/verifier.md` gives the five Go gate commands of hive as
the gate item. A verifier on another project runs them and records a fail.

**Status:** Open.

**Summary:**
Task `192-live-meter-data` (project stenographer, Python) failed verify. The
verifier ran the profile Go gate (`go vet`, `go test`, `golangci-lint` found
no packages) beside the project gate. The same harness also set
`NO_COLOR=1`, which fails four existing colour tests of the project. Neither
was a task defect. With `NO_COLOR` unset, the project gate passed. The
manager changed the feature spec gate to `env -u NO_COLOR` and stated that it
replaces the profile Go gate, then respawned a builder to re-complete.

**Reproduction steps:**
1. Register a Python project and create a task with a brief that names the
   project gate.
2. Spawn `hive agent spawn --profile verifier --task ID --harness codex`.
3. Observe the verifier run the Go gate from the profile and fail it.

**Workaround:** name the project gate in the feature spec and state that it
replaces the profile Go gate; pass `env -u NO_COLOR` where tests assert
colour.

**Recurrence 2026-10-28:** task `205-changelog-v0170` (project stenographer).
The verifier harness again exported `NO_COLOR=1`. The full gate gave 6 failed
colour tests in `tests/cli/hub/`. The same tests fail on base `31d955c` with
`NO_COLOR=1` and pass with `env -u NO_COLOR`. The brief of task 205 did not
carry the `env -u NO_COLOR` gate, so the verifier lost time to find the cause.
Put the `env -u NO_COLOR` gate in each stenographer brief, or unset `NO_COLOR`
in the agent environment that hive starts.
