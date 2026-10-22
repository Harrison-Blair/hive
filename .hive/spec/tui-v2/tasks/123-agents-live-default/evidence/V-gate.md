# V-gate Diff scope and gate

Proves each criterion. Date 2026-10-22. Result: **pass**.

## Merge

`git merge dev`: already up to date (`dev` = `origin/dev` = `10a7e992`).

## Diff scope

`git diff dev...HEAD --stat` lists only files of the source table and the
evidence:

- `.hive/spec/tui-v2/design/`: the 11 boards and the 8 changed grids (V2). No
  fixture changed.
- `.hive/spec/tui-v2/README.md`: decisions 58 to 61 (AC2), in a new section
  "Agents default and working cards". `ste_test.go` passes with a baseline of 0.
- `.hive/spec/tui-v2/tasks/123-agents-live-default/brief.md`: the brief copy.
  The worker changed "pod note" and "note" on 4 lines to "stopped text",
  because the STE glossary gives `note` the meaning "inbox record" and the
  ratchet failed with 4 violations. The meaning did not change.
- `testdata/ste-baseline.json`: the brief copy with 0.
- `internal/tui/agents.go`, `agents_screen.go`, `frame.go`,
  `component/card.go`, and their tests (`agents_test.go`,
  `agents_screen_test.go`, `component/card_test.go`,
  `component/keyhints_test.go`).
- `internal/tui/tui.go`: one line. Cause: `newModel` builds the model at each
  TUI start, so the default `graph.hideStopped = true` goes there. Demo mode
  then sets the fixture value in `applyDemoUI`, so `demo.go` did not change.
- `internal/tui/component/panel.go`: a new field `Edge`. Cause: `Card` draws
  its border with `Panel`, and `Panel` had only the `border` and `accent`
  colours. `Edge` sets the colour of an unfocused border to `working`.

## Gate

Full output: `V-gate.log`.

```text
$ gofmt -l ...        (no output) exit 0
$ go mod tidy -diff   (no output) exit 0
$ go vet ./...        exit 0
$ golangci-lint run ./...  exit 0
$ go test -race ./... exit 0
```
