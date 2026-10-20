# VERIFIED: 90-tui-idle-cpu

Verifier: t90-verifier-1. Date: 2026-10-20. Task commit: `81f09da2`. Base: `origin/dev` at `5f0d438f` (merge base of the task branch).

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | The four new tests in `idle_cpu_test.go` fail on the base and pass on the task branch. |
| V2 | pass | Task binary below 2 percent on each of the 8 screens. The base binary uses 10 to 20 percent on Chat and Logs. |
| V3 | pass | The Agents and Logs screens show each change within 2.3 seconds. |
| V4 | pass | Only the files of the source table changed. |
| V5 | pass | `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0. |

## V1 Tests

I copied `internal/tui/idle_cpu_test.go` to a detached worktree at `5f0d438f` and ran:

```sh
go test ./internal/tui -run '^(TestIdleTickDoesNotReloadUnchangedData|TestIdleTickReloadsChangedScreenOnce|TestHeaderRefreshReusesUnchangedScreenRender|TestMinuteRefreshUpdatesOnlyTheHeader)$' -count=1
```

Base output:

```text
--- FAIL: TestIdleTickDoesNotReloadUnchangedData: unchanged tick started 1 header loads and 1 screen loads; want none
--- FAIL: TestIdleTickReloadsChangedScreenOnce: changed tick started 1 header loads and 1 screen loads; want 0 and 1
--- FAIL: TestHeaderRefreshReusesUnchangedScreenRender: header refresh rendered screen 2 times; want cached screen
--- FAIL: TestMinuteRefreshUpdatesOnlyTheHeader: minute refresh rendered screen 2 times; want cached screen
```

Each failure is for the correct reason. On the task branch all four tests pass.

Changed existing tests:

- `goals_queue_test.go`: `refreshGoalsQueue` now uses `runIdleTick`. The test now writes a changed goal record and a changed task record before each refresh, so the refresh still reloads the screen data. All prior assertions on the goal selection and the project filter stay.
- `chat_screen_test.go`: the manual tick unpack became `runIdleTick`. The assertions `stalled` after the refresh and one send call stay. A temporary copy of this test with a counting `chatLoader` showed 1 chat load during the tick, so the test still checks that a reload keeps the failed draft. I removed the temporary file.

No assertion got weaker.

## V2 Idle CPU

Setup: base binary from `5f0d438f`, task binary from `81f09da2`. Each ran in its own tab of workspace `wRE` (base: pane `wRE:p2`, pid 2012446; task: pane `wRE:p3`, pid 2012459). Each used a scratch copy of the marked home (`.hive/README.md`, `.hive/state`, `.hive/spec`, the 40 MB `2026-10-20.jsonl` log, `reference/dogfood/friction`) in `HIVE_HOME`, its own temp `XDG_CONFIG_HOME`, and `env -u HERDR_ENV -u HERDR_PANE_ID -u HERDR_WORKSPACE_ID -u HERDR_SOCKET_PATH`. Both ran at the same time.

For each screen I sent the screen digit key, waited 5 seconds, then measured `utime+stime` from `/proc/<pid>/stat` over 60 seconds with no input.

| Screen | Base CPU | Task CPU |
| --- | ---: | ---: |
| Chat | 10.02% | 0.32% |
| Agents | 1.15% | 0.20% |
| Goals | 0.85% | 0.17% |
| Queue | 0.78% | 0.17% |
| Tasks | 0.98% | 0.22% |
| Friction | 0.85% | 0.18% |
| Health | 2.32% | 0.87% |
| Logs | 20.00% | 0.17% |

The high base use is on Chat and Logs. The other base screens are near 1 to 2 percent. This agrees with the worker evidence.

## V3 Live refresh

Task binary, scratch home, pane `wRE:p3`. I polled `herdr pane read` every 100 ms.

- Agents: I wrote the agent record `vprobe-g1.json` in the scratch state. The screen showed it in 1,325 ms. I removed it. The screen dropped it in 2,038 ms.
- Logs: I appended one `INFO` event `vprobe-log-event` to the scratch daily log. The screen showed it in 2,245 ms.

I did not open the Chat input. I pressed no `i` and no Enter. I sent no chat.

## V4 Scope

`git diff origin/dev --stat` lists: the brief, `evidence/V1.md` to `V5.md`, `internal/tui/chat_screen_test.go`, `goals_queue_test.go`, `idle_cpu_test.go`, `refresh.go`, `shell.go`, `tui.go`, and `testdata/ste-baseline.json`. The baseline diff only adds the brief copy with 0.

## V5 Gate

From the worktree root: `gofmt -l` printed nothing. `go mod tidy -diff` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` printed `0 issues.`, exit 0. `go test -race ./...` exit 0, all packages `ok`.

## Cleanup

I stopped pids 2012446 and 2012459. Their tabs `wRE:t2` and `wRE:t3` closed when the processes ended. I removed the temporary base worktree, the scratch homes, and the binaries.
