# 85 Logs verification

| Item | Value |
| --- | --- |
| Task | 85-logs-screen |
| Round | Full check after Repair 3 |
| Verifier | t85-verifier-4 |
| Task branch | `origin/task/85-logs-screen` = `b3218a7b` |
| Dev | `origin/dev` = `f4fc2ddf` (contains `765d61d3`, the task 83 merge) |
| Verdict | pass |

| ID | Result | Observation |
| --- | --- | --- |
| Merge | pass | `git fetch origin`, then `git merge-tree --write-tree origin/dev HEAD` exited 0. `origin/dev` is an ancestor of the task branch. In `internal/tui/tui.go`, the `screenLoadMsg` switch has the `tasksDataRequest` and `frictionDataRequest` cases and the `logsDataRequest` case with the day check and `acceptLogEvents`. |
| V1 | pass | Red: I made a detached worktree at `origin/dev` (`f4fc2ddf`) and copied in `cmd/request_id_test.go`, `cmd/logs_root_test.go`, `internal/supervise/request_id_test.go`, `internal/tui/logs_screen_test.go`, `internal/tui/tui_test.go`, and `internal/tui/escape_clear_test.go`. `go test ./cmd -run 'TestCommandRequestID\|TestLogs'` failed: `TestLogsJSONPrintsRawEventLines` (no logs start event), `TestCommandRequestIDAndLogsRequestFilter`, `TestCommandRequestIDReachesOperationLogs`, and the `memory` and `handoff` subtests of `TestCommandRequestIDReachesCommandOwnedLoggers` (request ID `""`). `go test ./internal/supervise -run TestLoopPassLogsOneRequestLifecycle` failed: 2 events, want two passes. `go test ./internal/tui -run 'TestLogsScreen\|TestRunHomeLogs\|TestViewingEscape'` failed in all 7 tests, including `TestLogsScreenDayNavigationLoadsSelectedFile`. Green: the same commands on the task branch passed each test. `TestLogsScreenDayNavigationLoadsSelectedFile` keeps the clock on 2026-04-07 and reads `first day` from the 2026-04-06 file. I removed the red worktree. |
| V2 | pass | Temp `HIVE_HOME` and `XDG_CONFIG_HOME`. `hive task list --project x` printed `no tasks`, exit 0. `hive logs --json` showed `command start` and `command end` (`duration_ms` 0, `outcome` ok) for `task list`, both with `request_id` `r-91fc7664`. `hive logs --request r-91fc7664` (text and `--json`) printed only those two events. `hive version` wrote no log event. |
| V3 | pass | New herdr workspace `wP5` (label `t85-verify-live`), cwd the temp home, with temp `HIVE_HOME` and `XDG_CONFIG_HOME`. I ran the task binary with `HERDR_ENV` unset, so startup did not spawn a coordinator or a supervisor. Seed: a project `demo`, a task `1-demo-task`, a copy of `reference/dogfood/friction/`, marker events in `2026-10-19.jsonl` and `2026-10-21.jsonl`, and WARN/ERROR events in `2026-10-20.jsonl` (one start/end pair with `request_id` `r-0badf00d`, `duration_ms` 42). Results are below. I closed `wP5`. `herdr workspace list` no longer shows the label. |
| V4 | pass | `git diff dev --stat`: 24 files. Each file is in the source table: the brief copy and the evidence, `cmd/common/`, `cmd/logs/`, `cmd/homeinit/homeinit.go` (only the request ID), `cmd/supervise/`, `cmd/*_test.go`, `internal/lib/log/`, `internal/supervise/`, `internal/tui/`, and `testdata/ste-baseline.json` (one added entry: the brief copy with 0). |
| V5 | pass | `gofmt -l` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` printed `0 issues.`, exit 0. `go test -race ./...` exit 0, each package `ok`. |

## V3 pane reads

| Input | Pane result |
| --- | --- |
| `8` | Logs, `logs · 2026-10-20.jsonl`, filter line `level: info component: all project: all ... day: 2026-10-20 request: all`. The start and end of `r-0badf00d` show as one row `WARN task needle probe 42ms`. DEBUG events are hidden. |
| `l` | `level: warn`. |
| `c`, `c` | `component: cli` shows only `ERROR cli needle cli error`. Then `component: task` shows only task rows. |
| `p` | `project: demo`. |
| `/`, `needle`, `enter` | The input showed `filter: needle`, then the filter line showed `text: needle` and one row `needle probe 42ms`. |
| `enter` | A side panel `event · 03:50:00` showed the JSON of the event, with `"request_id": "r-0badf00d"`. |
| `esc`, `r` | `request: r-0badf00d`, one row `needle probe 42ms`. |
| `r`, `esc`, `r`, `esc` | With the text filter clear, the view had 4 task rows. `r` showed only `needle probe 42ms`. `esc` showed all 4 rows again with `request: all`. |
| `]` | `logs · 2026-10-21.jsonl`, `day: 2026-10-21`, row `INFO cli next day marker`. |
| `[` | `logs · 2026-10-20.jsonl`, today's rows. |
| `[` | `logs · 2026-10-19.jsonl`, `day: 2026-10-19`, row `INFO cli previous day marker`. |
| `5`, `tab` | Tasks loaded: `active 0 queued 1 landed 0 failed 0 all 1`. The queued tab showed `1-demo-task  Verifier demo task  demo  created`. |
| `6` | Friction loaded: `open 11  resolved 16`, with the entry list. |
