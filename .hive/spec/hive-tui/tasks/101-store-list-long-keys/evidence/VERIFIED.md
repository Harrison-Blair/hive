# VERIFIED

Verifier: `t101-verifier-1`. Task head: `64cf3f52`. Base: `origin/dev` at
`1320b605`.

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | The R2 and R3 tests fail on the base code and pass on the task code. |
| V2 | pass | The task binary shows `96-tasks-evidence-view` on the Tasks screen (queued tab) and on the Queue screen. |
| V3 | pass | `git diff origin/dev...HEAD` changes only the files of the source table, the brief copy, and the evidence. |
| V4 | pass | `gofmt -l` and `go mod tidy -diff` print nothing. `go vet`, `golangci-lint` (`0 issues.`), and `go test -race ./...` exit 0. |

## V1

Base run: a detached worktree at `origin/dev` with the four task test files
copied in.

```text
$ go test ./internal/lib/store -run 'TestListKeepsLongStoreSegments$|TestListInKeepsLongStoreSegment$' -count=1
--- FAIL: TestListInKeepsLongStoreSegment (0.00s)
    generic_test.go:43: ListIn() records = []; want the long store segment
--- FAIL: TestListKeepsLongStoreSegments (0.00s)
    store_test.go:333: List() IDs = ["p1"]; want ["96-tasks-evidence-view" "manager-stenographer-g7" "p1"]
FAIL
$ go test ./internal/tui -run 'TestLoadTasksFindsLongTaskID$' -count=1
--- FAIL: TestLoadTasksFindsLongTaskID (0.00s)
    tasks_friction_test.go:29: loadTasks() omitted task "96-tasks-evidence-view"
FAIL
```

Task run, from the worktree root:

```text
--- PASS: TestListInKeepsLongStoreSegment
--- PASS: TestListInSkipsInvalidStoreSegments
--- PASS: TestListKeepsLongID
--- PASS: TestListKeepsLongStoreSegments
--- PASS: TestListSkipsInvalidStoreSegments
ok  	github.com/Harrison-Blair/hive/internal/lib/store
--- PASS: TestLoadTasksFindsLongTaskID
ok  	github.com/Harrison-Blair/hive/internal/tui
```

## V2

Scratch home with `HIVE_HOME` and a temporary `XDG_CONFIG_HOME`. One task
record `96-tasks-evidence-view`, status `created`. The binary ran in a private
tmux server with `env -u HERDR_ENV -u HERDR_PANE_ID`, so the TUI skipped the
herdr setup. The Chat input was not opened. Keys: `5`, `Tab`, `4`, then
`ctrl+c` twice.

| Binary | Tasks, queued tab | Queue |
| --- | --- | --- |
| task | `queued 1`, row `96-tasks-evid… Tasks evidence view hive created` | `1  96-tasks-evidence-view  now  ◆ auto` |
| base | `queued 0  all 0`, `no tasks` | `1  96-tasks-evidence-view  now  ◆ auto` |

The base Queue screen already showed the task, because it lists with `ListIn`,
which allowed 200 characters. The Tasks screen is the screen that the fix
changes. After the runs, `pgrep -af 'hive supervise'` showed no new process,
and `herdr tab list --workspace wQ8` showed no new tab.

## V3

`git diff origin/dev --stat` (two dots) also lists the files of
`100-no-backgrounds` in reverse, because `origin/dev` advanced after the merge
`12dab4a7`. The task changes are the three-dot diff:

```text
.hive/spec/hive-tui/tasks/101-store-list-long-keys/brief.md
.hive/spec/hive-tui/tasks/101-store-list-long-keys/evidence/V1.md .. V4.md
internal/lib/store/generic_test.go
internal/lib/store/legacy_test.go
internal/lib/store/store.go
internal/lib/store/store_test.go
internal/tui/tasks_friction_test.go
testdata/ste-baseline.json   (+1 line, the brief copy with 0)
11 files changed, 245 insertions(+), 20 deletions(-)
```

A trial merge of the task branch into `origin/dev` at `1320b605` has no
conflict. After the trial merge, `go test` of `./internal/lib/store`,
`./internal/tui`, and `.` passes.

## V4

```text
== gofmt   (no output) rc=0
== tidy    (no output) rc=0
== vet     rc=0
== lint    0 issues.  rc=0
== test    rc=0, no FAIL lines
```

## Manager questions

### 1. `TestLegacyListHidesLongID`

Commit `3b6c1323` (T22, 2026-10-18) added the test as a characterization test.
The T22 brief (`.hive/spec/hive-v1/tasks/T22-store-operations/brief.md`)
says: "`List` keeps the `records.ValidID` filter" and, in AC4, "`List("projects")`
still hides a file whose ID has more than 18 characters". It puts "A change to
the `records.ValidID` rule" out of scope. The filter came from `31540fa2`
(T01, project registry), where `List` served only `projects`.

The rule pinned the T01 behavior so that T22 did not change it. No document
gives a caller that needs the hidden IDs. The production callers of `List` are
`internal/project/list.go`, `internal/project/register.go` (duplicate check),
`internal/doctor/checks.go` (`tasks`), and `internal/tui/agents.go`.
`project register` refuses an ID that `records.ValidID` refuses, so a long
project ID can come only from a file that a person wrote by hand. With the fix,
`project list` shows that file and the duplicate check sees it. Neither is a
regression. The inversion is correct for the brief R1.

### 2. `ListIn` filter `validateComponent` to `validateSegment`

Path safety: each name comes from `os.ReadDir`, so it has no `/`. The path is
`filepath.Join(dir, name)` with the full name and its `.json` suffix. A scratch
probe test (removed after the run) showed:

- `..json` (ID `.`) reads the literal file inside the folder.
- `...json` (ID `..`) is refused.
- `.json` (empty ID) is refused.
- A symlink `esc.json` to a file outside the folder is refused by
  `readRecordFile` (`Lstat`, regular file only). That is the base behavior too.

`atomicWrite` temp files end in `.tmp-<digits>`, not `.json`, so the suffix
check skips them.

Callers: `ListIn` with an include function is used by `internal/lib/inbox`
(name list beside the records), `internal/lib/identity` (`collectNames`), and
`internal/agent/spawn_worker.go`. The include function runs after the name
check, so the names stay aligned with the records. No caller depends on the
stricter pattern. The full suite passes.

One behavior change to note: a file under a name that the old pattern refused
(for example `Bad.json`) is now read. When that file is not a regular file or
is not a JSON object, `List` or `ListIn` now returns an error for the whole
kind. Before, the name check skipped it. A file with a valid name already gave
that error on the base.
