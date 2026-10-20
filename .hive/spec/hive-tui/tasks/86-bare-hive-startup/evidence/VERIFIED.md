# VERIFIED — 86-bare-hive-startup

Verifier: `t86-verifier-1`. Task commit under check: `bb52bdaa` (base `dev` = `origin/dev` = `a7d0d7ec`).

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | In a temporary worktree at `dev`, I added the task test files. Red: `internal/lib/home`, `internal/lib/herdr`, `internal/agent`, `internal/tui` fail to build (`undefined: ReadSettings`, `PaneSendInput`, `StartCoordinator`, `setupSessionWith`). `TestInitializeKeepsCoordinatorConfigAndComments` fails with `expected one home key`. `TestBareHiveWithoutHerdrShowsStartupNotice` and `TestCoordinatorStartupUsesBareHiveTabs` fail. Green: each package passes on the task branch in `go test -race ./...`. `internal/tui/startup_test.go` covers a fresh start, a live coordinator with a fresh claim (no new tab), a stale claim, no herdr environment, a missing pane ID, and a failed step in the footer. |
| V2 | pass | Task binary, temp `XDG_CONFIG_HOME`. `bogus = "x"` gives `home config <path>: line 2: unknown key "bogus"`, exit 1. `[other]` gives `line 2: unknown table "other"`, exit 1. `hive init` on a config with a comment and `[coordinator]` (`claude`, `claude-haiku-4-5`) exits 0 and leaves the file byte-identical. The live V3 run reads the table (`--model claude-haiku-4-5`). |
| V3 | pass | Temp `HIVE_HOME`, temp `XDG_CONFIG_HOME`, coordinator model `claude-haiku-4-5`, new workspace `wNV`. First run: tabs `hive`, `coordinator`, `supervisor`. `hive agent list`: `coordinator` and `supervisor` live. The coordinator pane got the `⬢ You are the coordinator...` prompt and ran startup. After two `ctrl+c`, the root pane returned to the shell and both agents stayed live. Second run: the same three tab IDs, the same two agents, log steps `coordinator kept` and `supervisor kept`. I closed `wNV`; both temp processes exited. The real home supervisor (PID 1893563) was not touched. |
| V4 | pass | New workspace `wNY`, binary run with `HERDR_ENV`, `HERDR_PANE_ID`, `HERDR_SOCKET_PATH` unset. Footer: `not in herdr: no coordinator or supervisor started`. One tab, label unchanged (`1`). `hive agent list`: no agents. Workspace closed. |
| V5 | pass | `git diff dev --stat`: 27 files. Each path is in the source table (`internal/tui/`, `internal/agent/`, `internal/lib/home/`, `internal/lib/herdr/`, `internal/homeinit/`, `cmd/root_test.go`, coordinator skill, `AGENTS.md` section `Coordinator`, `skills_test.go`, `testdata/ste-baseline.json` with the brief copy at 0) or in the task folder. |
| V6 | pass | `gofmt -l` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` `0 issues.`, exit 0. `go test -race -count=1 ./...` exit 0. |

## Merge check

In a temporary worktree outside `projects/` at `bb52bdaa`, `git merge origin/dev` reported `Already up to date`. The branch already contains `origin/dev` (`a7d0d7ec`, with tasks 82 and 84). `go build` passed. `go test -race -count=1 ./...` exit 0. I removed the worktree with `git worktree remove`.

## Notes

- The first live attempt (workspace `wNS`) stopped at the Claude folder-trust prompt for the new temp directory. The TUI showed `error: coordinator: coordinator harness is blocked during startup` in the footer and kept running, as R3 requires. I accepted the trust prompt, closed `wNS`, reset the temp state, and did the V3 run above.
- The login shell of a new tab resets `PATH`, so the `supervisor` tab ran the installed `hive` with the temp `HIVE_HOME`. The TUI process was the task binary.
