# 85 Logs verification

| Item | Value |
| --- | --- |
| Task | 85-logs-screen |
| Round | After Repair 1 |
| Commit | `1d0abe4a` |
| Base | `a7d0d7ec` (`dev` and `origin/dev`, merged into the task branch in `ae82b982`) |
| Verdict | pass |

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | I copied `cmd/request_id_test.go`, `internal/supervise/request_id_test.go`, and `internal/tui/logs_screen_test.go` to a detached worktree at the old base `8d255eec`. All new tests failed there: 3 in `cmd`, `TestLoopPassLogsOneRequestLifecycle`, and 6 in `internal/tui` (with `TestLogsScreenDayNavigationLoadsSelectedFile` and `TestRunHomeLogsKeysOnStringReader`). Repair test: in a detached worktree at `1d0abe4a` with `internal/tui/tui.go` from `1d0abe4a^` (the old loader), `TestLogsScreenDayNavigationLoadsSelectedFile` failed with `logs_screen_test.go:152: view does not contain "first day"` and the view showed `no events on 2026-04-06`. The test now keeps the clock on the second day and reads the first day with `[`. With the fix it passes. |
| V2 | pass | Temp `HIVE_HOME` and `XDG_CONFIG_HOME`. `hive task list --project x` printed `no tasks`, exit 0. `hive logs --json` showed `command start` and `command end` (`duration_ms 0`, `outcome ok`), both with `request_id r-eaa3e3ad`. The log file has 2 events with this ID. `hive logs --request r-eaa3e3ad` showed only these 2 events, exit 0. `hive version` wrote no event. |
| V3 | pass | Herdr workspace `wNW` with the temp home (header `0 projects`). Fixture: 1 INFO event on 2026-10-19, 1 on 2026-10-21, and on 2026-10-20 (today) a request `r-cccc3333` (start and end, 42ms), a WARN `supervisor` event (project x), and an ERROR `tui` event (project y). `8`: Logs, level info, 3 rows, the request folded to `03:30:00 INFO cli needle cmd 42ms`. `l`: warn (2 rows), error (1), debug (all, with the CLI commands), info (3). `c`: cli, supervisor, tui (1 row each), all. `p`: x (1 row), then all. `/needle`: 1 row. `enter`: JSON side panel with `request_id r-cc…`. `r`: `request: r-cccc3333`, one folded row. `r` again and `r` then `esc`: `request: all`, 3 rows. `]`: `2026-10-21.jsonl`, row `tomorrow event`. `[ [`: `2026-10-19.jsonl`, row `yesterday event`. `]`: back to today, 3 rows. The repair works live: each day loads its file. I quit hive and closed `wNW`. `herdr workspace list` does not show it. |
| V4 | pass | `git diff dev --stat` (`dev` = `origin/dev` = `a7d0d7ec`): 24 files. Each is in the source table (`cmd/common/`, `cmd/logs/`, `internal/lib/log/`, `internal/tui/`, `testdata/ste-baseline.json`, the brief copy), in Scope change 1 (`cmd/supervise/`, `internal/supervise/`), a `cmd/` file that makes its own logger (`cmd/homeinit/homeinit.go`), or the evidence. |
| V5 | pass | `gofmt -l` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` printed `0 issues.`, exit 0. `go test -race -count=1 ./...` exit 0, no failed package. |

## Repair 1 check

1. The default `logLoader` in `internal/tui/tui.go` now reads with a reader whose clock is the selected day. V3 shows the 2026-10-19 and 2026-10-21 files live.
2. The day test keeps the clock on one day and reads the earlier file. It fails on the old loader (V1).
3. `ae82b982` merges `origin/dev`. `screenLoadMsg` in `internal/tui/tick.go` has both `day` and `goalsQueue`.
