# 85 Logs verification

| Item | Value |
| --- | --- |
| Task | 85-logs-screen |
| Commit | `68f1343f` |
| Base | `8d255eec` (`dev`) |
| Verdict | fail |

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | I copied `cmd/request_id_test.go`, `internal/supervise/request_id_test.go`, and `internal/tui/logs_screen_test.go` to a detached worktree at `8d255eec`. All new tests failed there for the correct reason: an empty request ID, 2 events instead of two loop passes, and no Logs screen behavior. All passed on the task branch. The changed existing tests follow R2 (default level info, day files). They are not weaker. |
| V2 | pass | Temp `HIVE_HOME`. `hive task list --project x` printed `no tasks` and exited 0. `hive logs --json` showed `command start` and `command end` (`duration_ms`, `outcome ok`) with `request_id r-1681c0b0`. `hive logs --request r-1681c0b0` showed only these 2 events. `hive version` wrote no event. |
| V3 | fail | Herdr workspace `wNF` with temp `HIVE_HOME` and `XDG_CONFIG_HOME`. The fixture had events on 2026-10-19, 2026-10-20 (today), and 2026-10-21. `8`, `l` (warn, error, debug, info), `c` (all, cli, supervisor, tui), `/needle`, `enter` (JSON panel), `r` (`request: r-cccc3333`, one folded row `needle cmd 42ms`), `r` again and `esc` (all): each worked. `]` and `[` changed the day label, but with no filter the pane showed `no events on 2026-10-21` and `no events on 2026-10-19`. Each of these files has one INFO event. Cause: the default `logLoader` in `internal/tui/tui.go:120` ignores its day argument and calls `reader.Read` with a zero cutoff, which reads only today's file. `eventsOnDay` then removes all events. `TestLogsScreenDayNavigationLoadsSelectedFile` passes only because it moves the clock to the second day. I closed the workspace. `herdr workspace list` does not show it. |
| V4 | pass | `git diff dev --stat`: 23 files. Each is in the source table, in Scope change 1 (`cmd/supervise/`, `internal/supervise/`), a `cmd/` file that makes its own logger (`cmd/homeinit/homeinit.go`), or the evidence. |
| V5 | pass | `gofmt -l` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` printed `0 issues.`, exit 0. `go test -race ./...` exit 0. |

## Merge check with origin/dev

`origin/dev` is now `a7d0d7ec`. It includes task 82 (Goals and Queue). This is newer than the merge note. In a temp worktree outside `projects/`, `git merge origin/dev` gave a conflict in `internal/tui/tick.go` (`screenLoadMsg` fields: this branch adds `day`, dev adds `goalsQueue`). I kept both fields in the temp worktree only. After that, `go build .` passed and `go test -race ./...` exit 0. I removed the temp worktree.
