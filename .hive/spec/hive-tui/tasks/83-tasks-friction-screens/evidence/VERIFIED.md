# VERIFIED — 83-tasks-friction-screens

Verifier: Claude Code, Opus 5.5. Task commit: `f5e7e062` (code at `580ca11b`).
Base: `a7d0d7ec` (merge-base with `dev`). Date: 2026-10-20.

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | The six new tests in `internal/tui/tasks_friction_test.go`, copied onto a base worktree at `a7d0d7ec`, fail: five with `Tasks/Friction screen did not request data`, the program test with `program output does not contain "tab filter"` (exit 1). On the task branch all six pass (exit 0). `TestTasksScreenRendersLandedTaskAndVerificationHistory` asserts the history line `V1 fail · V1 failed: missing shell load · 2026-10-20T10:00:00Z` from a fixture (Scope change 1). |
| V2 | pass | herdr workspace `wNZ`, task binary, `HIVE_HOME=/home/penguin/source/hive`, temp `XDG_CONFIG_HOME`. `5` then `tab` x2: `active 5  queued 7  [landed 45]  failed 1  all 57`. Row `79-tui-shell  TUI shell  hive  landed  t79-worker-2 · pass 7204d69`; detail `blocks: 80-chat-screen, 81-agents-screen`, `feature: hive-tui`, `grant: grants.md 2026-10-20 hive TUI goal`, `evidence: VERIFIED.md, V1.md, V2.md, V3.md, V4.md`, `current pass · - · 2026-10-20T05:31:26.78188019Z`. `enter` opened `# VERIFIED — 79-tui-shell`; `esc` returned; `b` opened `brief.md` `# TUI shell`; `esc` returned. `6`: `[open 10]  resolved 16`, 10 open rows, matching 10 files in `entries/` and 16 in `entries/resolved/`. `tab`: `open 10  [resolved 16]`; `enter` showed the full file of `agent-spawn-skips-first-prompt.md` with front matter. Workspace closed; `herdr workspace list` shows none left. |
| V3 | pass | `git diff a7d0d7ec --stat`: 14 files, only the source table files (`internal/tui/*`, `testdata/ste-baseline.json`, the brief copy) and V1-V4 evidence. Note: `git diff dev` now also lists task 86 files, because `dev` moved to `b12cce04` during this check; the branch does not touch them. |
| V4 | pass | `gofmt -l` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` `0 issues.` exit 0. `go test -race ./...` exit 0. |

## Merge check with dev

Temp worktree in the scratchpad (outside `projects/`), `git merge origin/dev` at `b12cce04` (tasks 84, 82, and 86): auto-merged `shell.go`, `tui.go`, `ste-baseline.json` with no conflict. `go build .` ok. `go test -race ./internal/tui/` ok (exit 0). Worktree removed with `git worktree remove`.

## Observation (not a fail of an item)

On the Friction screen the list height is `height - 5 - len(detailLines)` with an unbounded detail (`friction_screen.go:317`). For a long entry the table shrinks to one row: on the resolved tab only the selected row showed, and the detail ran past the panel border. Each row is still reachable with `↓`. Suggest a cap on the detail height in a follow-up.
