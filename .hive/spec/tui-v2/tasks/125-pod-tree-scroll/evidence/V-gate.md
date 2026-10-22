# V-gate Diff scope and gate

Proves: each criterion. Run on 2026-10-22. Result: **pass**.

## Merge dev

`git fetch origin`: `origin/dev` is `022c7197`, the base of the branch.
`git merge-base --is-ancestor origin/dev HEAD` exits 0, so no merge commit
was necessary.

## Diff scope

`git diff dev...HEAD --stat` is in `V-gate-stat.txt`. Outside the evidence
folder, the diff has only these files of the source table:

```text
 .hive/spec/tui-v2/README.md                        |   19 +-
 .../tui-v2/design/AgentsCoordinatorChild.dc.html   |    4 +-
 .../spec/tui-v2/design/AgentsCoordinatorChild.grid |   44 +-
 .hive/spec/tui-v2/design/AgentsNested.dc.html      |    4 +-
 .hive/spec/tui-v2/design/AgentsNested.grid         |  122 +-
 .hive/spec/tui-v2/design/CompactTreeNested.dc.html |   38 +
 .../tui-v2/design/CompactTreeNested.fixture.json   |   61 +
 .hive/spec/tui-v2/design/CompactTreeNested.grid    |   73 ++
 .hive/spec/tui-v2/design/CompactTreeWide.dc.html   |   38 +
 .../tui-v2/design/CompactTreeWide.fixture.json     |   63 +
 .hive/spec/tui-v2/design/CompactTreeWide.grid      |   73 ++
 .hive/spec/tui-v2/design/TreeNested.dc.html        |   38 +
 .hive/spec/tui-v2/design/TreeNested.fixture.json   |   61 +
 .hive/spec/tui-v2/design/TreeNested.grid           |  121 ++
 .hive/spec/tui-v2/design/TreeWide.dc.html          |   38 +
 .hive/spec/tui-v2/design/TreeWide.fixture.json     |   63 +
 .hive/spec/tui-v2/design/TreeWide.grid             |  121 ++
 .../spec/tui-v2/tasks/125-pod-tree-scroll/brief.md |  239 ++++
 internal/boardexport/check.go                      |    5 +
 internal/boardexport/check_test.go                 |   16 +
 internal/tui/agents_home_test.go                   |   37 +-
 internal/tui/agents_screen.go                      |  475 +++++--
 internal/tui/agents_screen_test.go                 |   81 +-
 internal/tui/agents_tree_test.go                   |  288 +++++
 internal/tui/component/panel.go                    |   11 +-
 internal/tui/component/panel_test.go               |   11 +
 internal/tui/demo_fixtures_test.go                 |    4 +-
 testdata/ste-baseline.json                         |    1 +
```

- `brief.md` is the worktree brief (the worker skill commits it).
- `agents.go`, `frame.go`, `demo.go`, and `tui.go` did not change.
- No grid changed except the 4 new grids and the 2 approved grids of AC6
  (V2).

## Gate

The full output is in `V-gate.log`. Summary:

| Command | Result |
| --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | no output, exit 0 |
| `go mod tidy -diff` | no output, exit 0 |
| `go vet ./...` | exit 0 |
| `golangci-lint run ./...` | `0 issues.`, exit 0 |
| `go test -race ./...` | each package `ok`, exit 0 (run with `-count=1` so that no result is cached) |
