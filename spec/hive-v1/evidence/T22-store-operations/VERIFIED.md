# T22-store-operations VERIFIED

| Item | Value |
| --- | --- |
| Task | T22-store-operations |
| Commit | d9b644f7 |
| Date | 2026-10-18 |
| Agent | verifier, pi, openai-codex/gpt-6.1-sol:high |
| Verdict | fail |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | pass | Fresh race tests and independent handle checks pass. |
| V2 | AC2 | fail | The fresh suite passes, but it also passes with an incorrect 2 MiB limit. See F2. |
| V3 | AC3, AC4 | fail | The fresh suite passes. Added probes show that the walk skips a directory and a linked collection. See F1. |
| V4 | AC2 | pass | One record reader and writer. Lock open and flock stay in lock.go. |
| V5 | AC5 | pass | All four gate commands pass. Tests use count=1. |
| API | AC1, AC2, AC3 | pass | All handle methods, UpdateIn, include selection, cancellation, and related reads pass independent probes. |
| Inputs | AC2 | pass | Static path, record, size, name, permission, and CLI error probes pass, except F1. |
| Mutations | AC1, AC3 | pass | Three broken behaviors fail on assertions. |
| Scope | AC5 | pass | No dependency change. Only owned source, tests, and evidence changed. |

## Findings

### F1: the walk skips a directory and a linked collection (AC2, AC3)

`collectWalkNodes` applies the temporary-name filter before `Lstat`.
A directory named `.probe.tmp-123` hides its `one.json` record.
A symbolic link with that name returns success without its path.
The link does not receive a read, but the walk skips it in silence.
The task requires directory descent and an error for a linked collection.

Expected directory result: `.probe.tmp-123/one.json`.
Actual directory result: no records, exit 0.
Expected link result: `store path invalid: <link path>`, exit 1.
Actual link result: no error, exit 0.

Recommendation: check the entry type before the temporary-file exclusion.
Skip only a regular atomic temporary file.
Descend a real directory.
Reject a linked collection with its path.
No source fix was authorized or applied.

### F2: the size tests use the implementation limit (AC2, test contract)

Changing `maxRecordSize` from 1 MiB to 2 MiB leaves the entire V2 suite green.
`TestReadInSizeBoundary` uses that constant to construct its expected boundary.
The test does not prove the fixed 1 MiB contract.
The added literal-bound test passes on the original source and fails on the mutant.
The current implementation passes the independent 1 MiB boundary probes.

Recommendation: use the literal contract limit in the expected fixtures.
Check one byte under, exactly at, and one byte above 1 MiB.
Do not compute the expected limit from the implementation constant.

## Limits

The approved boundary is a local user and a cooperating writer.
Hostile replacement between `Lstat` and `Open` remains deferred for T01.
No source, V-file, tracker, glossary, dependency, or baseline change forms part of this verification commit.
The separate path verification report was not read.
This pass does not approve landing.
A worker fix and a fresh verification pass remain.
Both independent reports must pass before landing.

## Runs

All item commands ran from the worktree root.
Mutation commands enter only an isolated source copy.
`-count=1` disables the test-result cache.
`--allow-serial-runners` permits another lint process without weakening checks.
The full input transcript and fixture programs are in the unique report:
`/home/penguin/source/hive/.hive/tmp/coordinator/reports/t22-verify-01.md`.


## V1

```sh
go test -count=1 -race ./internal/lib/store/ -run 'TestWithLock' -v
```

```text
=== RUN   TestWithLockDefaultTimeout
--- PASS: TestWithLockDefaultTimeout (2.00s)
=== RUN   TestWithLockContextDeadline
--- PASS: TestWithLockContextDeadline (0.10s)
=== RUN   TestWithLockCanceledContext
--- PASS: TestWithLockCanceledContext (0.00s)
=== RUN   TestWithLockLongDeadlineWaits
--- PASS: TestWithLockLongDeadlineWaits (2.21s)
=== RUN   TestWithLockFailureChangesNoRecord
--- PASS: TestWithLockFailureChangesNoRecord (0.10s)
=== RUN   TestWithLockCallbackErrorReleases
--- PASS: TestWithLockCallbackErrorReleases (0.00s)
=== RUN   TestWithLockPanicReleases
--- PASS: TestWithLockPanicReleases (0.00s)
=== RUN   TestWithLockHandleAfterRelease
--- PASS: TestWithLockHandleAfterRelease (0.00s)
=== RUN   TestWithLockNestedFails
--- PASS: TestWithLockNestedFails (0.20s)
=== RUN   TestWithLockLegacyMutationTimeout
=== RUN   TestWithLockLegacyMutationTimeout/Write
=== RUN   TestWithLockLegacyMutationTimeout/Update
=== RUN   TestWithLockLegacyMutationTimeout/Next
--- PASS: TestWithLockLegacyMutationTimeout (6.00s)
    --- PASS: TestWithLockLegacyMutationTimeout/Write (2.00s)
    --- PASS: TestWithLockLegacyMutationTimeout/Update (2.00s)
    --- PASS: TestWithLockLegacyMutationTimeout/Next (2.00s)
=== RUN   TestWithLockReadsDoNotTakeTheLock
--- PASS: TestWithLockReadsDoNotTakeTheLock (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/store	11.621s
exit 0
```


## V2

```sh
go test -count=1 -race ./internal/lib/store/ -run 'TestStoreNames|TestReadIn|TestWriteIn|TestListIn' -v
```

```text
=== RUN   TestStoreNamesAcceptValidComponents
--- PASS: TestStoreNamesAcceptValidComponents (0.00s)
=== RUN   TestStoreNamesRejectInvalidComponents
--- PASS: TestStoreNamesRejectInvalidComponents (0.00s)
=== RUN   TestWriteInNestedInbox
--- PASS: TestWriteInNestedInbox (0.00s)
=== RUN   TestListInMissingCollection
--- PASS: TestListInMissingCollection (0.00s)
=== RUN   TestListInIncludeSelectsOne
--- PASS: TestListInIncludeSelectsOne (0.00s)
=== RUN   TestListInSkipsInvalidNames
--- PASS: TestListInSkipsInvalidNames (0.00s)
=== RUN   TestReadInRecordInvalid
--- PASS: TestReadInRecordInvalid (0.00s)
=== RUN   TestReadInSizeBoundary
--- PASS: TestReadInSizeBoundary (0.08s)
=== RUN   TestListInBadSelectedFileReturnsNoRecords
--- PASS: TestListInBadSelectedFileReturnsNoRecords (0.00s)
=== RUN   TestReadInPathTypes
=== RUN   TestReadInPathTypes/state_root_symlink
=== RUN   TestReadInPathTypes/collection_ancestor_symlink
=== RUN   TestReadInPathTypes/record_leaf_symlink
=== RUN   TestReadInPathTypes/dangling_record_leaf_symlink
=== RUN   TestReadInPathTypes/record_leaf_directory
=== RUN   TestReadInPathTypes/lock_symlink
--- PASS: TestReadInPathTypes (0.00s)
    --- PASS: TestReadInPathTypes/state_root_symlink (0.00s)
    --- PASS: TestReadInPathTypes/collection_ancestor_symlink (0.00s)
    --- PASS: TestReadInPathTypes/record_leaf_symlink (0.00s)
    --- PASS: TestReadInPathTypes/dangling_record_leaf_symlink (0.00s)
    --- PASS: TestReadInPathTypes/record_leaf_directory (0.00s)
    --- PASS: TestReadInPathTypes/lock_symlink (0.00s)
=== RUN   TestReadInPathTypesLinux
--- PASS: TestReadInPathTypesLinux (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/store	1.100s
exit 0
```


## V3

```sh
go test -count=1 -race ./internal/lib/store/ -run 'TestGrouped|TestLegacy|TestWalk' -v
```

```text
=== RUN   TestGroupedReadSeesEqualVersions
--- PASS: TestGroupedReadSeesEqualVersions (0.06s)
=== RUN   TestGroupedPartialWriteStays
--- PASS: TestGroupedPartialWriteStays (0.00s)
=== RUN   TestGroupedCounterProcesses
--- PASS: TestGroupedCounterProcesses (1.05s)
=== RUN   TestLegacyListHidesLongID
--- PASS: TestLegacyListHidesLongID (0.00s)
=== RUN   TestLegacyLongTaskIDAndColonCounter
--- PASS: TestLegacyLongTaskIDAndColonCounter (0.00s)
=== RUN   TestLegacyUnsafeNamesStillRejected
--- PASS: TestLegacyUnsafeNamesStillRejected (0.00s)
=== RUN   TestWalkRecordsFindsNestedAndSkipsOnlyTemporary
--- PASS: TestWalkRecordsFindsNestedAndSkipsOnlyTemporary (0.00s)
=== RUN   TestWalkStopsAtCallbackError
--- PASS: TestWalkStopsAtCallbackError (0.00s)
=== RUN   TestWalkRejectsLinkedCollection
--- PASS: TestWalkRejectsLinkedCollection (0.00s)
=== RUN   TestWalkRejectsLinkedRecord
--- PASS: TestWalkRejectsLinkedRecord (0.00s)
=== RUN   TestWalkBadNestedRecord
--- PASS: TestWalkBadNestedRecord (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/store	2.116s
exit 0
```


## V4

```sh
grep -nE 'os\.(Open|OpenFile|ReadFile|WriteFile|Create|CreateTemp)\(' internal/lib/store/*.go
```

```text
internal/lib/store/atomic.go:10:	file, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-")
internal/lib/store/generic_test.go:68:	data, err := os.ReadFile(path)
internal/lib/store/generic_test.go:132:		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0o644); err != nil {
internal/lib/store/generic_test.go:152:	if err := os.WriteFile(empty, nil, 0o644); err != nil {
internal/lib/store/generic_test.go:156:	if err := os.WriteFile(array, []byte(`[]`), 0o644); err != nil {
internal/lib/store/generic_test.go:164:	if err := os.WriteFile(target, []byte(`{"id":"target"}`), 0o644); err != nil {
internal/lib/store/generic_test.go:195:	if err := os.WriteFile(filepath.Join(dir, "exact.json"), []byte(exact), 0o644); err != nil {
internal/lib/store/generic_test.go:203:	if err := os.WriteFile(oversize, []byte(`{"padding":"`+strings.Repeat("x", maxRecordSize)+`"}`), 0o644); err != nil {
internal/lib/store/generic_test.go:217:	if err := os.WriteFile(filepath.Join(dir, "alpha.json"), []byte(`{"id":"alpha"}`), 0o644); err != nil {
internal/lib/store/generic_test.go:221:	if err := os.WriteFile(invalid, []byte(`{"id":`), 0o644); err != nil {
internal/lib/store/generic_test.go:272:		if err := os.WriteFile(target, []byte(`{"id":"target"}`), 0o644); err != nil {
internal/lib/store/generic_test.go:283:		data, readErr := os.ReadFile(target)
internal/lib/store/generic_test.go:327:		if err := os.WriteFile(target, []byte("keep"), 0o600); err != nil {
internal/lib/store/generic_test.go:337:		data, readErr := os.ReadFile(target)
internal/lib/store/lock.go:67:	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
internal/lib/store/lock_test.go:33:	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
internal/lib/store/store.go:546:	file, err := os.Open(path)
internal/lib/store/store_test.go:240:	if err := os.WriteFile(large, []byte(`{"padding":"`+strings.Repeat("x", maxRecordSize)+`"}`), 0o644); err != nil {
internal/lib/store/store_test.go:244:	if err := os.WriteFile(truncated, []byte(`{"id":`), 0o644); err != nil {
internal/lib/store/store_test.go:248:	if err := os.WriteFile(empty, nil, 0o644); err != nil {
internal/lib/store/store_test.go:252:	if err := os.WriteFile(array, []byte(`[]`), 0o644); err != nil {
internal/lib/store/store_test.go:259:	if err := os.WriteFile(target, []byte(`{"id":"target"}`), 0o644); err != nil {
internal/lib/store/store_test.go:297:	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
internal/lib/store/store_test.go:323:		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0o644); err != nil {
internal/lib/store/store_test.go:343:	if err := os.WriteFile(filepath.Join(dir, "alpha.json"), []byte(`{"id":"alpha"}`), 0o644); err != nil {
internal/lib/store/store_test.go:347:	if err := os.WriteFile(invalid, []byte(`{"id":`), 0o644); err != nil {
internal/lib/store/walk_test.go:27:	if err := os.WriteFile(filepath.Join(stateDir, "counters", ".hidden.json.tmp-123"), []byte(`{"value":0}`), 0o644); err != nil {
internal/lib/store/walk_test.go:30:	if err := os.WriteFile(filepath.Join(stateDir, "counters", ".other.json"), []byte(`{"value":0}`), 0o644); err != nil {
internal/lib/store/walk_test.go:33:	if err := os.WriteFile(filepath.Join(stateDir, "tasks", "notes.txt"), []byte("not a record"), 0o644); err != nil {
internal/lib/store/walk_test.go:89:	if err := os.WriteFile(filepath.Join(target, "inside.json"), []byte(`{"id":"inside"}`), 0o644); err != nil {
internal/lib/store/walk_test.go:112:	if err := os.WriteFile(target, []byte(`{"id":"target"}`), 0o644); err != nil {
internal/lib/store/walk_test.go:135:	if err := os.WriteFile(bad, []byte(`{"id":`), 0o644); err != nil {
exit 0
```

```sh
grep -n 'syscall.Flock' internal/lib/store/*.go
```

```text
internal/lib/store/lock.go:57:		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
internal/lib/store/lock.go:77:		err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
internal/lib/store/lock_test.go:37:	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
internal/lib/store/lock_test.go:44:			_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
exit 0
```

```sh
grep -n 'io.LimitReader' internal/lib/store/*.go
```

```text
internal/lib/store/store.go:551:	data, err := io.ReadAll(io.LimitReader(file, maxRecordSize+1))
exit 0
```


## V5

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
```

```text
exit 0
```

```sh
go vet ./...
```

```text
exit 0
```

```sh
golangci-lint run --allow-serial-runners ./...
```

```text
0 issues.
exit 0
```

```sh
go test -count=1 -race ./...
```

```text
ok  	github.com/Harrison-Blair/hive	2.573s
ok  	github.com/Harrison-Blair/hive/cmd	1.014s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/version	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.011s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.042s
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/records	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/store	15.035s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.009s
ok  	github.com/Harrison-Blair/hive/internal/project	1.016s
ok  	github.com/Harrison-Blair/hive/internal/tui	1.011s
exit 0
```


## api

```sh
/home/penguin/source/hive/projects/hive/worktrees/T22-store-operations/.hive/tmp/t22-verify-01.MbwxtI/store-probe --home /home/penguin/source/hive/projects/hive/worktrees/T22-store-operations/.hive/tmp/t22-verify-01.MbwxtI/home semantics
```

```text
PASS long task ID, UpdateIn, include selection, all held methods, all released/nil methods, error identity
PASS ten deterministic related-read rounds, unlocked mixtures, unlocked lists, waiting cancellation
PASS completed prefix under a later lock
exit 0
```


## repeated-processes

```sh
go test -count=10 -race ./internal/lib/store/ -run '^TestGroupedCounterProcesses$' -v
```

```text
=== RUN   TestGroupedCounterProcesses
--- PASS: TestGroupedCounterProcesses (1.04s)
=== RUN   TestGroupedCounterProcesses
--- PASS: TestGroupedCounterProcesses (1.04s)
=== RUN   TestGroupedCounterProcesses
--- PASS: TestGroupedCounterProcesses (1.04s)
=== RUN   TestGroupedCounterProcesses
--- PASS: TestGroupedCounterProcesses (1.04s)
=== RUN   TestGroupedCounterProcesses
--- PASS: TestGroupedCounterProcesses (1.04s)
=== RUN   TestGroupedCounterProcesses
--- PASS: TestGroupedCounterProcesses (1.04s)
=== RUN   TestGroupedCounterProcesses
--- PASS: TestGroupedCounterProcesses (1.04s)
=== RUN   TestGroupedCounterProcesses
--- PASS: TestGroupedCounterProcesses (1.04s)
=== RUN   TestGroupedCounterProcesses
--- PASS: TestGroupedCounterProcesses (1.04s)
=== RUN   TestGroupedCounterProcesses
--- PASS: TestGroupedCounterProcesses (1.04s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/store	11.426s
exit 0
```


## mutation-lock-wait

```sh
diff -u .hive/tmp/t22-verify-01.MbwxtI/original/internal/lib/store/lock.go .hive/tmp/t22-verify-01.MbwxtI/mutation-lock-wait/internal/lib/store/lock.go
```

```text
--- .hive/tmp/t22-verify-01.MbwxtI/original/internal/lib/store/lock.go	2026-10-18 01:03:41.000000000 -0400
+++ .hive/tmp/t22-verify-01.MbwxtI/mutation-lock-wait/internal/lib/store/lock.go	2026-10-18 01:17:47.482889974 -0400
@@ -18,7 +18,7 @@
 )
 
 const (
-	defaultLockWait = 2 * time.Second
+	defaultLockWait = 0 * time.Second
 	lockRetry       = 10 * time.Millisecond
 )
 
exit 1
```

```sh
cd .hive/tmp/t22-verify-01.MbwxtI/mutation-lock-wait && go test -count=1 -race ./internal/lib/store/ -run '^TestWithLockDefaultTimeout$' -v
```

```text
=== RUN   TestWithLockDefaultTimeout
    lock_test.go:69: WithLock() returned after 31.008µs; want about two seconds
--- FAIL: TestWithLockDefaultTimeout (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/store	0.005s
FAIL
exit 1
```


## mutation-record-size

```sh
diff -u .hive/tmp/t22-verify-01.MbwxtI/original/internal/lib/store/store.go .hive/tmp/t22-verify-01.MbwxtI/mutation-record-size/internal/lib/store/store.go
```

```text
--- .hive/tmp/t22-verify-01.MbwxtI/original/internal/lib/store/store.go	2026-10-18 01:03:41.000000000 -0400
+++ .hive/tmp/t22-verify-01.MbwxtI/mutation-record-size/internal/lib/store/store.go	2026-10-18 01:17:47.482889974 -0400
@@ -17,7 +17,7 @@
 )
 
 const (
-	maxRecordSize    = 1 << 20
+	maxRecordSize    = 1 << 21
 	maxComponents    = 8
 	maxComponentSize = 200
 )
exit 1
```

```sh
cd .hive/tmp/t22-verify-01.MbwxtI/mutation-record-size && go test -count=1 -race ./internal/lib/store/ -run '^TestReadInSizeBoundary$' -v
```

```text
=== RUN   TestReadInSizeBoundary
--- PASS: TestReadInSizeBoundary (0.12s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/store	1.126s
exit 0
```


## mutation-size-suite

```sh
cd .hive/tmp/t22-verify-01.MbwxtI/mutation-record-size && go test -count=1 -race ./internal/lib/store/ -run 'TestStoreNames|TestReadIn|TestWriteIn|TestListIn' -v
```

```text
=== RUN   TestStoreNamesAcceptValidComponents
--- PASS: TestStoreNamesAcceptValidComponents (0.00s)
=== RUN   TestStoreNamesRejectInvalidComponents
--- PASS: TestStoreNamesRejectInvalidComponents (0.00s)
=== RUN   TestWriteInNestedInbox
--- PASS: TestWriteInNestedInbox (0.00s)
=== RUN   TestListInMissingCollection
--- PASS: TestListInMissingCollection (0.00s)
=== RUN   TestListInIncludeSelectsOne
--- PASS: TestListInIncludeSelectsOne (0.00s)
=== RUN   TestListInSkipsInvalidNames
--- PASS: TestListInSkipsInvalidNames (0.00s)
=== RUN   TestReadInRecordInvalid
--- PASS: TestReadInRecordInvalid (0.00s)
=== RUN   TestReadInSizeBoundary
--- PASS: TestReadInSizeBoundary (0.12s)
=== RUN   TestListInBadSelectedFileReturnsNoRecords
--- PASS: TestListInBadSelectedFileReturnsNoRecords (0.00s)
=== RUN   TestReadInPathTypes
=== RUN   TestReadInPathTypes/state_root_symlink
=== RUN   TestReadInPathTypes/collection_ancestor_symlink
=== RUN   TestReadInPathTypes/record_leaf_symlink
=== RUN   TestReadInPathTypes/dangling_record_leaf_symlink
=== RUN   TestReadInPathTypes/record_leaf_directory
=== RUN   TestReadInPathTypes/lock_symlink
--- PASS: TestReadInPathTypes (0.00s)
    --- PASS: TestReadInPathTypes/state_root_symlink (0.00s)
    --- PASS: TestReadInPathTypes/collection_ancestor_symlink (0.00s)
    --- PASS: TestReadInPathTypes/record_leaf_symlink (0.00s)
    --- PASS: TestReadInPathTypes/dangling_record_leaf_symlink (0.00s)
    --- PASS: TestReadInPathTypes/record_leaf_directory (0.00s)
    --- PASS: TestReadInPathTypes/lock_symlink (0.00s)
=== RUN   TestReadInPathTypesLinux
--- PASS: TestReadInPathTypesLinux (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/store	1.132s
exit 0
```


## mutation-walk-skip

```sh
diff -u .hive/tmp/t22-verify-01.MbwxtI/original/internal/lib/store/store.go .hive/tmp/t22-verify-01.MbwxtI/mutation-walk-skip/internal/lib/store/store.go
```

```text
--- .hive/tmp/t22-verify-01.MbwxtI/original/internal/lib/store/store.go	2026-10-18 01:03:41.000000000 -0400
+++ .hive/tmp/t22-verify-01.MbwxtI/mutation-walk-skip/internal/lib/store/store.go	2026-10-18 01:17:47.482889974 -0400
@@ -321,7 +321,7 @@
 		if root && name == "lock" {
 			continue
 		}
-		if strings.HasPrefix(name, ".") && strings.Contains(name, ".tmp-") {
+		if strings.HasPrefix(name, ".") {
 			continue
 		}
 		path := filepath.Join(dir, name)
exit 1
```

```sh
cd .hive/tmp/t22-verify-01.MbwxtI/mutation-walk-skip && go test -count=1 -race ./internal/lib/store/ -run '^TestWalkRecordsFindsNestedAndSkipsOnlyTemporary$' -v
```

```text
=== RUN   TestWalkRecordsFindsNestedAndSkipsOnlyTemporary
    walk_test.go:57: WalkRecords() paths = ["counters/inbox:coordinator.json" "inbox/coordinator/1.json" "projects/hive.json" "tasks/12-spawn-worker.json"]; want ["counters/.other.json" "counters/inbox:coordinator.json" "inbox/coordinator/1.json" "projects/hive.json" "tasks/12-spawn-worker.json"]
--- FAIL: TestWalkRecordsFindsNestedAndSkipsOnlyTemporary (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/store	0.006s
FAIL
exit 1
```


## mutation-shared-lock

```sh
diff -u .hive/tmp/t22-verify-01.MbwxtI/original/internal/lib/store/lock.go .hive/tmp/t22-verify-01.MbwxtI/mutation-shared-lock/internal/lib/store/lock.go
```

```text
--- .hive/tmp/t22-verify-01.MbwxtI/original/internal/lib/store/lock.go	2026-10-18 01:03:41.000000000 -0400
+++ .hive/tmp/t22-verify-01.MbwxtI/mutation-shared-lock/internal/lib/store/lock.go	2026-10-18 01:17:47.483889967 -0400
@@ -74,7 +74,7 @@
 		waitUntil, _ = ctx.Deadline()
 	}
 	for {
-		err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
+		err := syscall.Flock(int(file.Fd()), syscall.LOCK_SH|syscall.LOCK_NB)
 		if err == nil {
 			return file, nil
 		}
exit 1
```

```sh
cd .hive/tmp/t22-verify-01.MbwxtI/mutation-shared-lock && go test -count=1 -race ./internal/lib/store/ -run '^TestGroupedCounterProcesses$' -v
```

```text
=== RUN   TestGroupedCounterProcesses
    grouped_test.go:132: counter = 54; want 200
--- FAIL: TestGroupedCounterProcesses (1.01s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/store	1.018s
FAIL
exit 1
```


## regression

```sh
cd .hive/tmp/t22-verify-01.MbwxtI/original && go test -count=1 -race ./internal/lib/store/ -run '^TestVerifier' -v
```

```text
=== RUN   TestVerifierLiteralSizeLimit
--- PASS: TestVerifierLiteralSizeLimit (0.01s)
=== RUN   TestVerifierWalkTemporaryTypes
=== RUN   TestVerifierWalkTemporaryTypes/directory
    verifier_test.go:54: walk records = []; want .collection.tmp-123/one.json
=== RUN   TestVerifierWalkTemporaryTypes/symlink
    verifier_test.go:50: linked collection error = <nil>; want store path invalid: /tmp/TestVerifierWalkTemporaryTypessymlink2962120624/001/.hive/state/.collection.tmp-123
--- FAIL: TestVerifierWalkTemporaryTypes (0.00s)
    --- FAIL: TestVerifierWalkTemporaryTypes/directory (0.00s)
    --- FAIL: TestVerifierWalkTemporaryTypes/symlink (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/store	0.008s
FAIL
exit 1
```

```sh
cd .hive/tmp/t22-verify-01.MbwxtI/mutation-record-size && go test -count=1 -race ./internal/lib/store/ -run '^TestVerifierLiteralSizeLimit$' -v
```

```text
=== RUN   TestVerifierLiteralSizeLimit
    verifier_test.go:30: 1 MiB + 1 read error = <nil>; want record too large
--- FAIL: TestVerifierLiteralSizeLimit (0.03s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/store	0.038s
FAIL
exit 1
```

```sh
git status --porcelain
```

```text
exit 0
```


## scope

```sh
git diff 840255d3f598ad18239eccd7a2eb2e93547ca1fb..d9b644f7d7cddd3612a92c08d0e1e141a199e87e -- go.mod go.sum
```

```text
exit 0
```

```sh
git diff --name-only 840255d3f598ad18239eccd7a2eb2e93547ca1fb..d9b644f7d7cddd3612a92c08d0e1e141a199e87e
```

```text
internal/lib/store/doc.go
internal/lib/store/generic_test.go
internal/lib/store/grouped_test.go
internal/lib/store/legacy_test.go
internal/lib/store/lock.go
internal/lib/store/lock_test.go
internal/lib/store/path_linux_test.go
internal/lib/store/store.go
internal/lib/store/store_test.go
internal/lib/store/walk_test.go
spec/hive-v1/evidence/T22-store-operations/V1.md
spec/hive-v1/evidence/T22-store-operations/V2.md
spec/hive-v1/evidence/T22-store-operations/V3.md
spec/hive-v1/evidence/T22-store-operations/V4.md
spec/hive-v1/evidence/T22-store-operations/V5.md
exit 0
```

```sh
git log --format=full 840255d3..d9b644f7
```

```text
commit d9b644f7d7cddd3612a92c08d0e1e141a199e87e
Author: Harrison-Blair <hjblair@me.com>
Commit: Harrison-Blair <hjblair@me.com>

    docs(store): add T22 store operation evidence

commit 8c3b24d3816cd0ed317680f848974db4ade89ed5
Author: Harrison-Blair <hjblair@me.com>
Commit: Harrison-Blair <hjblair@me.com>

    feat(store): add grouped and nested store operations

commit 3b6c1323492694812c26c2c2f492183e244bf894
Author: Harrison-Blair <hjblair@me.com>
Commit: Harrison-Blair <hjblair@me.com>

    test(store): add store operation tests and empty symbols
exit 0
```


### temp-regular

```sh
/home/penguin/source/hive/projects/hive/worktrees/T22-store-operations/.hive/tmp/t22-verify-01.MbwxtI/store-probe-fast --home /home/penguin/source/hive/projects/hive/worktrees/T22-store-operations/.hive/tmp/t22-verify-01.MbwxtI/fixtures-r2/temp-regular walk
```

```text
exit 0
PASS temp-regular: expected exit 0, no write
```


### temp-directory

```sh
/home/penguin/source/hive/projects/hive/worktrees/T22-store-operations/.hive/tmp/t22-verify-01.MbwxtI/store-probe-fast --home /home/penguin/source/hive/projects/hive/worktrees/T22-store-operations/.hive/tmp/t22-verify-01.MbwxtI/fixtures-r2/temp-directory walk
```

```text
exit 0
FAIL temp-directory: expected exit 0, no write
```


### temp-link

```sh
/home/penguin/source/hive/projects/hive/worktrees/T22-store-operations/.hive/tmp/t22-verify-01.MbwxtI/store-probe-fast --home /home/penguin/source/hive/projects/hive/worktrees/T22-store-operations/.hive/tmp/t22-verify-01.MbwxtI/fixtures-r2/temp-link walk
```

```text
exit 0
FAIL temp-link: expected exit 1, no write, one plain error
```
