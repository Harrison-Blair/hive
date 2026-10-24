---
title: "An Agents state test fails without local design files"
status: "Open"
opened: 2026-10-24
decision: "Task 156 keeps the six state cases and makes their unit input self-contained."
---

**Issue:** Full-fixture verification missed a unit test dependency that fails in CI without local spec files.

**Summary:**
Task 147 landed at `0431585a` after independent verification and screen review.
Its worktree and fresh clone had the matching ignored design fixtures.
The gate passed in that environment.
Push CI `40080999510` and PR CI `40081004556` then failed.
All six `TestGfAgentsCardStates` cases read `GfAgents.fixture.json` directly.
CI has only tracked inputs and cannot open that file.

The complete-fixture gate checked rendering with the accepted inputs.
It did not check that the unit tests work without those inputs.
Both input modes need separate checks. A green run with fixtures supplies
no result for the environment without them.

**Reproduction steps:**
1. Make a fresh clone at `0431585a` without `.hive/spec/`.
2. Run `env -u TERM TZ=UTC go test -race -run '^TestGfAgentsCardStates$' ./internal/tui/`.
3. Observe all six cases fail at `agents_repair_test.go:88` with a missing `GfAgents.fixture.json` file.

**Evidence:** Task 147 `evidence/post-land-ci/` retains both failed CI logs.
Task 156 must retain the absent-file red run, deliberate behavior breaks,
and separate checks with tracked inputs and complete matching spec inputs.

**Decision:** Task `156-gf-agents-state-unit` changes only the unit setup and
its tests, plus the brief baseline entry. It keeps every existing state,
glyph, selection, and freshness assertion. It adds no skip or CI exception.
Production rendering, boards, and spec rules remain unchanged.
