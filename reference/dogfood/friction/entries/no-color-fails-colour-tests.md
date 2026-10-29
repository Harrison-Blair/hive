---
title: "A harness can export NO_COLOR=1 and fail the colour tests of a project"
description: "Hive neither sets nor clears NO_COLOR, so the test run of a spawned agent can fail on colour assertions"
status: "Open"
opened: 2026-10-27
decision: ""
---

**Issue:** The harness environment of a spawned agent had `NO_COLOR=1`. Tests
that assert colour output then fail, and the failure is not a code defect. Hive
neither sets nor clears `NO_COLOR` for the agents it starts.

**Summary:** On 2026-10-27, the verifier of stenographer task
`192-live-meter-data` (harness codex) ran the project gate with `NO_COLOR=1`, and
four colour tests failed. With `NO_COLOR` unset, the gate passed. On 2026-10-28,
the verifier of task `205-changelog-v0170` again had `NO_COLOR=1`, and six colour
tests in `tests/cli/hub/` failed. The same tests fail on base `31d955c` with
`NO_COLOR=1` and pass with `env -u NO_COLOR`. The workaround is to run the gate
with `env -u NO_COLOR`.

**Reproduction steps:**
1. Register a project whose tests assert colour output.
2. Run `hive agent spawn --name N --harness codex --model M --project P --prompt "run the project gate"`.
3. Observe colour test failures with `NO_COLOR=1` in the environment of the agent.
4. Run the gate again with `env -u NO_COLOR` and observe a pass.
