# V-gate Diff scope and gate

Proves each criterion. Date 2026-10-22. Result: **pass**.

`origin/dev` is merged into the task branch (`git merge-base --is-ancestor
origin/dev HEAD` is true). `git diff origin/dev...HEAD --stat`:
`V-gate-stat.txt`. It names only files of the source table and this task
folder:

- `internal/agent/send.go` and `send_test.go` (AC9).
- `internal/boardexport/check.go` and `check_test.go` (AC1).
- `internal/tui/component/boxjoin.go`, `boxjoin_test.go`, `doc.go`, and
  `symbols_test.go`.
- `internal/tui/agents*.go`, `demo.go`, `frame.go`, and their tests. Also
  `demo_fixtures_test.go` (the board count is now 35) and
  `selection_marker_test.go`.
- `internal/tui/tui.go` and `refresh.go`. The cause is in V3.
- `testdata/ste-baseline.json`: the brief copy with 0.
- `.hive/spec/tui-v2/README.md` and the 15 boards of AC1 and AC11.

Full gate output from the worktree root: `V-gate.log`. `gofmt -l` and
`go mod tidy -diff` print nothing. `go vet` passes. `golangci-lint`
reports 0 issues. `go test -race ./...` passes in each package.

## Repair

The repair after the verification fail is in `REPAIR.md`.
