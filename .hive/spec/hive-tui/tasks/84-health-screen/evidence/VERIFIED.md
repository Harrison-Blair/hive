# VERIFIED — 84-health-screen

Verifier re-ran each item on 2026-10-20 from the worktree root at task commit `3823d98d`.

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | Base `833ee784` (merge base with `dev`) with the new tests copied in: `internal/doctor` failed (4 new tests: check `supervisor binary` / `stale worktrees` / `unpushed task branches` absent; `TestRunChecksWithFakeHerdr` 14 checks, want 17). `internal/tui` failed to build (`undefined: healthDataRequest`, `unknown field health in struct literal of type screenLoadMsg`). Task branch: all 5 doctor tests and `TestHealth*` (3) plus `TestScreenKeysSelectAllEightScreens` pass. |
| V2 | pass | Temp home with a live claim for a copied `sleep` (mtime 3 h old), a newer task `hive` first on `PATH`, temp `XDG_CONFIG_HOME`, and a task record `1-landed` (delivery commit set, `pushed: true`) with its worktree directory. `hive doctor`: `warn supervisor binary: supervisor …/oldsup modified 2026-10-20T04:05:14Z; hive …/bin/hive modified 2026-10-20T07:05:14Z; supervisor binary is older` and `warn stale worktrees: demo/1-landed`. `--json`: outcome `ok`, both checks `warn` with the same detail. |
| V3 | pass | New herdr workspace `wN5` (pane `wN5:p1`), temp `XDG_CONFIG_HOME`, `HIVE_HOME=/home/penguin/source/hive`, task binary. After `7`: panels `home` (8 checks) and `agents` (9 checks), all 17 doctor checks with `✓`/`!`, `DETAIL herdr socket: …`, capacity line, footer with `r re-check · ask in Chat`. After `↓`: `DETAIL herdr pane: wN5:p1`. After `r`: checks reloaded, same selection kept; log shows `health selection` then `health refresh`. Quit with two Ctrl+C, closed `wN5` (absent from `workspace list`). |
| V4 | pass | `git diff dev --stat`: only the brief copy, `evidence/V1-V5.md`, `internal/doctor/` (checks.go, run.go, checks_test.go, health_test.go), `internal/tui/` (health_screen.go, health_screen_test.go, shell.go, shell_test.go, tick.go, tui.go, doc.go), and `testdata/ste-baseline.json` (brief copy added with 0). |
| V5 | pass | `gofmt -l` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` 0 issues, exit 0. `go test -race ./...` exit 0. |

Notes: the stale-worktree rule matches README G8 ("landed" = delivery commit and `pushed` true) plus status `cancelled`. The V3 run against the real home only read state; the TUI wrote its normal log events.
