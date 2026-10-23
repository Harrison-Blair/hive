# V-gate Diff scope and gate

Proves each criterion. Run again on 2026-10-23 by the worker after the
repair of verifier fail F1. `git fetch` and `git merge origin/dev` gave
`Already up to date` (dev at 34a7d85d, an ancestor of HEAD). Result:
**pass**.

## Diff scope

`git diff origin/dev...HEAD --stat`: 268 files, 17978 insertions, 40
deletions (most are evidence, also the verifier output). Outside the evidence directory the diff has only these files:

- Source table files: `cmd/ask/ask.go`, `cmd/root.go`, `internal/ask/ask.go`
  and its test, `internal/tui/{ask,chat,demo,frame,header,mouse,refresh,tui}.go`,
  `internal/tui/ask_test.go`, `internal/tui/component/chat.go` and its test,
  `internal/boardexport/check.go` and its test, the five
  `.agents/skills/hive-*/SKILL.md` files, `.hive/spec/tui-v2/README.md`, the
  five new boards (`.dc.html`, `.fixture.json`, `.grid`) and the `Help`
  board and grid in `.hive/spec/tui-v2/design/`.
- Extra files, each a test or a check file:
  - `cmd/ask_root_test.go`: the root command test of `hive ask` (new test
    beside `cmd/root.go`).
  - `internal/tui/demo_fixtures_test.go`: the five new fixtures added to the
    demo fixture list.
  - `.hive/spec/tui-v2/tasks/131-hive-ask/brief.md`: the updated brief of
    the manager, with STE fixes so that `TestSTERatchet` passes at baseline
    0 (task 129 precedent), and
    `testdata/ste-baseline.json`: the brief added with a baseline of 0.

No `reference/` file lists the hive commands, so no reference file changed.

## Gate

Full output: `V-gate.log`.

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')    (no output)  rc=0
$ go mod tidy -diff                                         (no output)  rc=0
$ go vet ./...                                                           rc=0
$ golangci-lint run ./...                                    0 issues.   rc=0
$ go test -race ./...                         45 packages ok, 0 FAIL  rc=0
```
