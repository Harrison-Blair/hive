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

## Repair after the verify fail (merge of task 36)

Task `36-screen-work-queue` landed on `dev` at `e7cdbd09`. The verifier
evidence of the first verify stays in this folder (`VERIFIED.md`,
`verifier/`).

### Merge

`git merge origin/dev` (`e7cdbd09`) gave two conflicts. The merge commit is
`34b7757f`.

- `.hive/spec/tui-v2/README.md`, section `Manager decisions`: both sides
  added a block of worker decisions. The merge keeps both blocks, first
  `34-screen-chat`, then `36-screen-work-queue`, with a blank line between
  them.
- `testdata/ste-baseline.json`: both sides added a brief entry with 0. The
  merge keeps both entries in sorted order.

The merge commit also changes one line in `internal/tui/doc.go`: the file map
named `chat_screen.go`, which this task removed. It now names `chat.go`.

### Diff scope

Command: `git diff origin/dev...HEAD --stat` at `34b7757f` plus this
evidence. Output: `diff-stat.txt`. Outside this evidence folder, only these
files changed:

```text
.hive/spec/tui-v2/README.md            (section Manager decisions only)
.hive/spec/tui-v2/tasks/34-screen-chat/brief.md
internal/tui/chat.go
internal/tui/chat_screen.go            (removed)
internal/tui/chat_screen_test.go       (removed)
internal/tui/chat_test.go
internal/tui/demo.go
internal/tui/doc.go                    (file map line of chat.go)
internal/tui/frame.go
internal/tui/selection_marker_test.go
internal/tui/shell_test.go
internal/tui/tui.go
testdata/ste-baseline.json             (one entry)
```

Each file is in the source table of the brief.

### Golden checks

Output: `golden-rerun.txt`. `TestChatMatchesItsBoard`,
`TestCompactChatMatchesItsBoard`, and `TestChatFitsEachSize` pass. Each test
of `internal/tui` and `internal/tui/component` that matches `Chat`, `Board`,
or `FitsEachSize` passes.

### Live check

`V2.md`, section `Repair rerun`: no cell differs under the space cell rule
for `Chat` and `CompactChat`.

### Gate run 3: pass

Full output: `gate-run-3.txt`, at `34b7757f`.

```text
$ gofmt -l
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

`gofmt -l` and `go mod tidy -diff` printed nothing. Each package printed `ok`
or `no test files`.

## Repair after the land conflict (merge of task 40)

The verify of `72d44ff7` passed. Then `hive land` stopped on a conflict with
`dev` at `fbca94d5` (task `40-screen-logs`).

### Merge

`git merge origin/dev` at `fbca94d5` gave a conflict in two files. The merge
commit is `a0c5bf60`.

- `.hive/spec/tui-v2/README.md`: both blocks stay. The decisions of task
  `34-screen-chat` come first, then the manager and worker decisions of task
  `40-screen-logs`.
- `internal/tui/tui.go`: the `Model` fields are the Chat fields of this task
  (`chatFeed`, `chatNoise`, and the send fields) and the Logs fields of task 40
  (`textFilter`, `filterDraft`, `logOffset`, `logPendingTime`, and the others).
  Each side removed one viewport: task 40 removed `viewport` (old Logs) and
  this task removed `chatViewport` (old Chat). The merge removes both. Then
  `layout()` had no body, so the merge removes it and its calls in `tui.go`,
  `demo.go`, `agents_test.go`, and `shell_test.go`. Each removed test line was
  only a call of `m.layout()` before an assertion. No assertion changed. The
  `viewport` import of `tui.go` went away.

### Diff scope

Output: `diff-stat-2.txt` (`git diff origin/dev...HEAD --stat`, evidence
lines removed).

```text
 .hive/spec/tui-v2/README.md                        |  10 +
 .hive/spec/tui-v2/tasks/34-screen-chat/brief.md    | 144 +++++
 internal/tui/agents_test.go                        |   1 -
 internal/tui/chat.go                               | 401 ++++++++++++++
 internal/tui/chat_screen.go                        | 497 -----------------
 internal/tui/chat_screen_test.go                   | 581 --------------------
 internal/tui/chat_test.go                          | 598 +++++++++++++++++++++
 internal/tui/demo.go                               |  15 +-
 internal/tui/doc.go                                |   2 +-
 internal/tui/frame.go                              |  11 -
 internal/tui/selection_marker_test.go              |   7 -
 internal/tui/shell_test.go                         |  23 +-
 internal/tui/tui.go                                | 100 +---
 testdata/ste-baseline.json                         |   1 +
```

Each file is in the source table of the brief.

### Golden checks

Output: `merge2-golden.txt`, from
`go test -count=1 -v -run 'TestChat|TestCompactChat|TestDemoChat' ./internal/tui/`.
Each of the 14 tests passes, with `TestChatMatchesItsBoard`,
`TestCompactChatMatchesItsBoard`, and `TestChatFitsEachSize`.

### Live check

`V2.md`, section `Repair rerun 2`: no cell differs under the space cell rule
for `Chat` and `CompactChat`.

### Gate run 4: pass

Full output: `gate-run-4.txt`, on the tree of `a0c5bf60`.

```text
$ gofmt -l
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

`gofmt -l` and `go mod tidy -diff` printed nothing. Each package printed `ok`
or `no test files`.

## Second merge before land (merge of task 38)

After the commit `c57973b2`, `dev` moved to `86f48a68` (task
`38-screen-friction`). The merge commit is `1d850724`.

- `.hive/spec/tui-v2/README.md`: both blocks stay. The decisions of task
  `34-screen-chat` come first, then the decisions of task `38-screen-friction`.
- `testdata/ste-baseline.json`: both brief entries stay with 0.
- `internal/tui/chat_test.go`: `friction_test.go` of task 38 has a
  `checkBoard` helper with the same check as the helper of `chat_test.go`.
  The two declarations did not compile. The merge removes the helper of
  `chat_test.go` and its `cellgrid` import. The Chat golden tests now use the
  helper of `friction_test.go`. Its failure text does not include the
  rendered view.
- A deliberate change proves that the golden check still fails: in
  `chat.go` `borderNote`, `dimStyle` changed to `titleStyle`. Output:
  `merge3-break-token.txt`:

```text
--- FAIL: TestChatMatchesItsBoard (0.01s)
    chat_test.go:24: 14 cells differ from Chat.grid:
        row 3 col 60: want ' ' dim normal, got ' ' accent bold
        row 3 col 61: want '↑' dim normal, got '↑' accent bold
```

The change was then removed. `git diff HEAD -- internal/tui/chat.go` is empty.

### Diff scope

Output: `diff-stat-3.txt` (`git diff origin/dev...HEAD --stat` at
`86f48a68`, evidence lines removed). The files are the files of
`diff-stat-2.txt`. Each file is in the source table of the brief.

### Golden checks

Output: `merge3-golden.txt`. Each of the 14 Chat tests passes. Each Friction
test passes.

### Live check

`V2.md`, section `Repair rerun 3`: no cell differs under the space cell rule.

### Gate run 5: pass

Full output: `gate-run-5.txt`, on the tree of `1d850724`.

```text
$ gofmt -l
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
