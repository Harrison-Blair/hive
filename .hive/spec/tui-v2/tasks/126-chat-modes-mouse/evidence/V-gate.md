# V-gate Diff scope and gate

Proves each criterion. Run on 2026-10-23 by the worker. Result: **pass**.

## Merge and diff scope

`git fetch origin` and `git merge origin/dev` gave `Already up to date`
(`dev` is f7990406). `git diff origin/dev...HEAD --stat` is in
`V-gate-stat.txt`. The changed files are only files of the source table and
the evidence:

- `.hive/spec/tui-v2/README.md` (decisions 79 to 90, the board table).
- `.hive/spec/tui-v2/design/`: the 3 new boards with their fixtures and
  grids, and the approved boards and grids of AC6 (decision 90).
- `internal/boardexport/check.go` and `check_test.go`.
- `internal/tui/*.go` and `internal/tui/component/*.go` with their tests.
- `testdata/ste-baseline.json` (the brief has 12 STE findings, not 0).
- `.hive/spec/tui-v2/tasks/126-chat-modes-mouse/` (the brief and the
  evidence).

## Gate

The full output is in `V-gate.log`. The first run found one lint issue
(QF1001 De Morgan in `mouse.go` line 103). The worker changed the condition
to its equal form, and the second run passed:

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')     (no output)
$ go mod tidy -diff                                           (no output)
$ go vet ./...                                                exit 0
$ golangci-lint run ./...                                     0 issues.
$ go test -race ./...                                         each package ok
```
