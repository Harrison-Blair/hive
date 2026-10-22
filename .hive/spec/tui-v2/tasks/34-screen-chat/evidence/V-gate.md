# V-gate Diff scope and gate

Proves each criterion.

## Merge

`git merge origin/dev` at `8679a063` merged the grids and the shared code
of task `42-wave6-board-redraw` (`dev` = `origin/dev` = `5db10420`). The one
conflict was in `testdata/ste-baseline.json`: both sides added brief
entries. The merge keeps all three entries in sorted order. No line of
`component.List`, `component.Details`, or the frame hints had a conflict.

## Diff scope

Command: `git diff origin/dev...HEAD --stat` at `dd45eb01`. Output:
`diff-stat.txt`. Outside this evidence folder, only these files changed:

```text
.hive/spec/tui-v2/README.md            (section Manager decisions only)
.hive/spec/tui-v2/tasks/34-screen-chat/brief.md
internal/tui/chat.go                   (new screen)
internal/tui/chat_screen.go            (removed old screen)
internal/tui/chat_screen_test.go       (removed old screen tests)
internal/tui/chat_test.go
internal/tui/demo.go
internal/tui/frame.go                  (old scrollChatFeed moved to chat.go)
internal/tui/selection_marker_test.go  (old Chat case removed, see V1)
internal/tui/shell_test.go
internal/tui/tui.go
testdata/ste-baseline.json
```

Each file is in the source table of the brief. The task made no change in
`internal/tui/component/`.

## Gate

Full output: `gate-run-1.txt` and `gate-run-2.txt`.

### Run 1: fail in a package that the task does not change

`gofmt -l` and `go mod tidy -diff` printed nothing. `go vet` exit 0.
`golangci-lint` `0 issues.` `go test -race ./...` exit 1:

```text
--- FAIL: TestConcurrentUpdateProcesses (3.29s)
    store_test.go:62: helper failed: exit status 1: --- FAIL: TestStoreProcessHelper (2.00s)
            store_test.go:107: store lock timeout
--- FAIL: TestUpdateCallbackRunsUnderLock (2.66s)
    store_test.go:218: store lock timeout
FAIL	github.com/Harrison-Blair/hive/internal/lib/store	46.194s
```

The diff has no file in `internal/lib/store/`. The package passed 3 times
alone with `go test -race -count=1 ./internal/lib/store/` (`store-alone.txt`).
The 2-second store lock timed out under the load of the full parallel run.
Each other package passed, `internal/tui` too.

### Run 2: pass

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
exit 0
$ go mod tidy -diff
exit 0
$ go vet ./...
exit 0
$ golangci-lint run ./...
0 issues.
exit 0
$ go test -race ./...
exit 0
```

Each package printed `ok` or `no test files`. See `gate-run-2.txt`.
