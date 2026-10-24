---
title: "The verification passes and CI then fails on the time zone and TERM"
status: "Open"
opened: 2026-10-24
decision: ""
---

**Summary:**
Task 137-gf-frame-chat passed the gate of its worker and two verifiers on the
computer of the user. After the land on `dev` at `e3f107dc`, CI run
`40052654140` failed. Six `internal/tui` tests used the local time zone and
failed with `TZ=UTC`. `cmd` `TestBareHiveRunsTUIUntilQuit` failed with
`TERM=dumb` and in a new clone. No step of the gate or of the verifier skill
runs the tests with the environment of CI before the land. Task
132-tui-clock-tests had a fault of the same class. Repair task:
141-gf-ci-tz-term.

**Reproduction steps:**
1. Check out `dev` at `e3f107dc`.
2. Run `TZ=UTC go test ./internal/tui/ -run TestChatCompactStatusRowKeepsTheClock -count=1`.
3. Run `TERM=dumb go test ./cmd/ -run TestBareHiveRunsTUIUntilQuit -count=1`.
4. Both tests fail. Without the variables, both tests pass.
