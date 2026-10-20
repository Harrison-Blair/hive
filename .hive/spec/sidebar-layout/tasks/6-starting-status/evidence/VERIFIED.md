# 6-starting-status VERIFIED

| Item | Value |
| --- | --- |
| Task | 6-starting-status |
| Commit | 94771372 |
| Date | 2026-10-19 |
| Agent | verifier, claude, claude-opus-5-5 |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | pass | Green: both spawn tests pass. Red: with `initialAgentStatus` set to `live`, both tests fail on the `starting` assertions. |
| V2 | AC2, AC3 | pass | Green: four tests pass. Red: three separate controls fail the matching test. `TestEventsAgentDetectedReleasedExited` passes in each control. |
| V3 | AC4 | pass | Green: the list test passes. Red: with `List` set to skip `starting` records, the test fails. |
| V4 | AC5 | pass | Five `--pane` spawns pass on the first attempt. Each record shows `starting`, `live`, `stopped`. `wBC` has its original tabs at the end. |
| V5 | AC6 | pass | All four gate commands pass. Tests use `-count=1`. |

## V1 spawn test

```sh
go test ./internal/agent/ -run 'TestManagerSpawnWaitsForShellSecondRead|TestWorkerSpawnWaitsForShellSecondRead' -v -count=1
```

Green, at commit 94771372:

```text
--- PASS: TestManagerSpawnWaitsForShellSecondRead (0.20s)
--- PASS: TestWorkerSpawnWaitsForShellSecondRead (0.21s)
ok  	github.com/Harrison-Blair/hive/internal/agent	0.414s
```

Red, in a scratch worktree at 94771372. The control changes `initialAgentStatus` in `internal/agent/register.go` to return `records.AgentLive` for a spawn.

```text
    shell_ready_test.go:75: statuses during shell reads = [live live]; want [starting starting]
--- FAIL: TestManagerSpawnWaitsForShellSecondRead (0.20s)
    shell_ready_test.go:165: status at agent.start/final = "live"/"live"; want starting/live
--- FAIL: TestWorkerSpawnWaitsForShellSecondRead (0.21s)
exit 1
```

## V2 supervisor tests

```sh
go test ./internal/supervise/ -run 'TestEventsAgentDetectedReleasedStartingIgnored|TestEventsAgentDetectedReleasedExited|TestEventsStartingTopologyClosedUnresolved|TestReconcileStartingWithMissingPaneUnresolved' -v -count=1
```

Green, at commit 94771372:

```text
--- PASS: TestEventsAgentDetectedReleasedStartingIgnored (0.00s)
--- PASS: TestEventsAgentDetectedReleasedExited (0.00s)
--- PASS: TestEventsStartingTopologyClosedUnresolved (0.00s)
    --- PASS: TestEventsStartingTopologyClosedUnresolved/pane (0.00s)
    --- PASS: TestEventsStartingTopologyClosedUnresolved/tab (0.00s)
    --- PASS: TestEventsStartingTopologyClosedUnresolved/workspace (0.00s)
--- PASS: TestReconcileStartingWithMissingPaneUnresolved (0.00s)
ok  	github.com/Harrison-Blair/hive/internal/supervise	0.003s
```

Red, in a scratch worktree. Each control was applied alone and then removed.

Control a: remove the `starting` check in `onAgentDetected`.

```text
    events_test.go:197: status = "exited"; want starting
--- FAIL: TestEventsAgentDetectedReleasedStartingIgnored (0.00s)
--- PASS: TestEventsAgentDetectedReleasedExited (0.00s)
exit 1
```

Control b: `activeAllByField` matches `live` records only.

```text
    events_test.go:248: status = "starting"; want unresolved   (pane, tab, workspace)
--- FAIL: TestEventsStartingTopologyClosedUnresolved (0.00s)
exit 1
```

Control c: the reconcile scan matches `live` records only.

```text
    reconcile_test.go:102: status = "starting"; want unresolved
--- FAIL: TestReconcileStartingWithMissingPaneUnresolved (0.00s)
exit 1
```

## V3 list test

```sh
go test ./internal/agent/ -run 'TestListAllShowsStartingStatus' -v -count=1
```

Green: `--- PASS: TestListAllShowsStartingStatus (0.00s)`.

Red, with `List` set to skip `starting` records:

```text
    list_order_test.go:23: ListText() = "no agents\n"; want starting status
--- FAIL: TestListAllShowsStartingStatus (0.00s)
exit 1
```

## V4 live check

The binary was built from the worktree at 94771372 into the verifier scratch directory. The home was the worktree (`--home "$PWD"`). A scratch project `hive` and a scratch task `1-t6-verify-check` were made in that home. herdr 0.9.5.

Original `wBC` tabs: `wBC:t1`, `wBC:t2`, `wBC:tC`.

1. `herdr tab create --workspace wBC --cwd "$PWD"` gave tab `wBC:tK`, pane `wBC:pM`. `hive --home "$PWD" supervise` in `wBC:pM` printed `supervisor live, generation 1` (pid 636992).
2. Five times: a new tab, then `agent spawn --role verifier --project hive --task 1-t6-verify-check --harness pi --model opencode-go/deepseek-v4.1-flash --arg=--no-approve --name t6-check-<n> --pane <pane> --json`, then `agent stop`. A loop polled `agent list --all --json` every 20 ms during the spawn.

```text
created wBC:tM wBC:pN
t6-check-1 spawn exit 0; observed: starting live
{"outcome":"ok","phase":"complete","error":null,"status":"live","pane":"wBC:pN"}
{"outcome":"ok","phase":"complete","error":null}
t6-check-1 after stop: stopped
created wBC:tN wBC:pP
t6-check-2 spawn exit 0; observed: starting live
{"outcome":"ok","phase":"complete","error":null,"status":"live","pane":"wBC:pP"}
{"outcome":"ok","phase":"complete","error":null}
t6-check-2 after stop: stopped
created wBC:tP wBC:pQ
t6-check-3 spawn exit 0; observed: starting live
{"outcome":"ok","phase":"complete","error":null,"status":"live","pane":"wBC:pQ"}
{"outcome":"ok","phase":"complete","error":null}
t6-check-3 after stop: stopped
created wBC:tQ wBC:pR
t6-check-4 spawn exit 0; observed: starting live
{"outcome":"ok","phase":"complete","error":null,"status":"live","pane":"wBC:pR"}
{"outcome":"ok","phase":"complete","error":null}
t6-check-4 after stop: stopped
created wBC:tR wBC:pS
t6-check-5 spawn exit 0; observed: starting live
{"outcome":"ok","phase":"complete","error":null,"status":"live","pane":"wBC:pS"}
{"outcome":"ok","phase":"complete","error":null}
t6-check-5 after stop: stopped
```

A search of the five spawn outputs for `agent record changed` found nothing.

3. SIGINT stopped the supervisor (pid 636992). `herdr pane close wBC:pM` returned `ok`. `agent stop` had closed the five agent panes: `herdr pane get` returns `pane_not_found` for `wBC:pN`, `wBC:pP`, `wBC:pQ`, `wBC:pR`, and `wBC:pS`.

```text
herdr tab list --workspace wBC  ->  ["wBC:t1","wBC:t2","wBC:tC"]
```

No step sent `workspace.close`. The check closed only the pane that it made.

## V5 gate

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run ./...
go test -race -count=1 ./...
```

```text
gofmt exit 0          (no output)
vet exit 0
0 issues.
lint exit 0
ok  	github.com/Harrison-Blair/hive	3.933s
ok  	github.com/Harrison-Blair/hive/internal/agent	2.977s
ok  	github.com/Harrison-Blair/hive/internal/doctor	3.053s
ok  	github.com/Harrison-Blair/hive/internal/supervise	1.426s
ok  	github.com/Harrison-Blair/hive/internal/tui	7.682s
... each other package ok or no test files
test exit 0
```

## Remarks

- The worker evidence for V1–V3 names commit `18111ae1` for the red runs. The tests use `records.AgentStarting`, so they do not compile at that commit. The text says that the red runs used negative controls. This verification repeats the red runs with its own controls at 94771372.
- R5: `03-architecture.md` names `starting` in the lifecycle section and in the supervision table. `friction.md` marks the race resolved with commit `c7b2566` and adds the `agent_not_ready` entry.
