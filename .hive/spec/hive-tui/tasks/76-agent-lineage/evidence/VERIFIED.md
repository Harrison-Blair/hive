# VERIFIED: 76-agent-lineage

Verifier: `t76-verifier-1`. Task head: `4915549c`. Merge base with `dev`: `a1520150`.

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | Red: the five new test files copied onto a detached base worktree at `a1520150`. `TestAgentLineageFieldsRoundTripAndLegacyJSON`, `TestSpawnedRegistrationSelectsNewestMatchingReplacement`, `TestShowAndListExposeLineageFields`, `TestEventsBlockedWritesCounterRecordAndStatus`, `TestAppendAndStatusKeepsTimestampWhenAgentStatusDoesNotChange`, `TestStatusOnlyKeepsTimestampWhenAgentStatusDoesNotChange`, and `TestSpawnRecordsCallerAsParent` failed. The reasons were missing fields, empty `replaces`, missing show lines, empty `status_changed_at`, and `unknown flag: --agent`. Green: `go test -count=1 ./internal/lib/records ./internal/agent ./internal/supervise ./cmd/agent/spawn` passed. `go test . -run TestSTERatchet` passed. |
| V2 | pass | Live in herdr 0.9.5. Temp `HIVE_HOME` and `XDG_CONFIG_HOME`, binary built from `4915549c`. Fake manager `manager-hive` (key `manager-hive-g1`) in workspace `wK4`. Pi worker `t1-worker-1` spawned with `--agent manager-hive` in `wK5`, then stopped. Second spawn `t1-worker-2` in `wK6`. `hive agent show t1-worker-2` printed `parent: manager-hive-g1` and `replaces: t1-worker-1`. `hive agent list --json` had the same two fields. Both workers stopped, `wK4` closed. `herdr workspace list` showed none of `wK4`-`wK6`. Temp files removed. |
| V3 | pass | Local `dev` moved to `ba63f8bc` (task 74 landed) after the worker sync, so `git diff dev --stat` also lists task 74 files as reverse changes. `git diff dev...HEAD --stat` lists 17 files, each in the source table or the evidence: the brief, `V1.md`-`V4.md`, `cmd/agent/spawn/{spawn.go,cli_test.go}`, `internal/agent/{register.go,render.go,spawn.go,lineage_test.go,show_test.go}`, `internal/lib/records/{agent.go,agent_test.go}`, `internal/supervise/{supervise.go,events_test.go}`, `testdata/ste-baseline.json` (one added row for the brief copy at 0). `git merge-tree` with `dev` is clean. |
| V4 | pass | `gofmt -l` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` printed `0 issues.`, exit 0. `go test -race ./...` exit 0. |

## Mutation checks

- `appendAndStatus` with the change-only guard removed: `TestAppendAndStatusKeepsTimestampWhenAgentStatusDoesNotChange` failed. Correct.
- `replacementKey` with the live-record skip removed: `TestSpawnedRegistrationSelectsNewestMatchingReplacement` failed with `replaces = "worker-live"`. Correct.
- `statusOnly` with the change-only guard removed: `TestStatusOnlyKeepsTimestampWhenAgentStatusDoesNotChange` still passed. The fake clock returns a fixed time, so the repeated write gives the same timestamp. The code in `statusOnly` has the correct guard (read in the diff), but no test detects its removal. This is a test gap, not a behavior fault. It does not change the V1 result.
