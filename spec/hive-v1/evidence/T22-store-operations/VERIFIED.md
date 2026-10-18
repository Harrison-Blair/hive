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

## 2026-10-18 — current verdict — t22-verify-02

This section gives the current verdict. The earlier FAIL evidence remains unchanged as history.

| Item | Value |
| --- | --- |
| Role | Fresh independent verifier |
| Agent | pi, `openai-codex/gpt-6.1-sol:high` |
| Branch | `task/T22-store-operations` |
| Base | `840255d3f598ad18239eccd7a2eb2e93547ca1fb` |
| Tested worker HEAD | `a3fea5a4304a29e1c24fd8e93b3da30ccf5aa5ad` |
| Fix source commit | `d8155e5c78a21c0f02e275b8d9c3de1919954ae1` |
| Overall verdict | **PASS** |

The initial worktree was clean on the exact branch and worker HEAD.
The explicit remote branch check returned the same worker HEAD before tests started.
I read the task, V1–V5, current fix report, both earlier summaries/findings, and complete base-to-worker diff.
The fix report marker matched. Actual tests and observable behavior supplied the acceptance evidence.

| Item | Proves | Current verdict | Basis |
| --- | --- | --- | --- |
| V1 | AC1 | PASS | Fresh race tests, all handle methods, cancellation, release, bounded wait, and error identity. |
| V2 | AC2 | PASS | Literal size fixtures, detected cap mutation, name checks, and both static filesystem checklists. |
| V3 | AC3, AC4 | PASS | Related reads, completed prefixes, four-process counters, legacy behavior, and corrected walk behavior. |
| V4 | AC2 | PASS | One bounded record reader, one record writer, one temporary maker, flock in lock.go. |
| V5 | AC5 | PASS | Empty gofmt list, vet, lint, and uncached full race gate. |

### Finding dispositions

| Earlier finding | Current disposition | Evidence |
| --- | --- | --- |
| Primary F1: temporary-name filter | CLOSED | Matching directories descend. Matching links fail with their paths. Only regular atomic temporary files skip. |
| Primary F2: derived size fixtures | CLOSED | Literal 1048575/1048576/1048577 fixtures pass. A 2097152 cap fails generic and legacy assertions. |
| Path F01: linked `.hive` ancestor | CLOSED | Open rejects before external creation. Later legacy/generic access, lock, walk, and counter calls refuse the static ancestor. |
| Path F02: linked state at Open | CLOSED | Open rejects linked and wrong-type roots. Later operations retain the same type rule. |

The original post-Open replaced-root refusal remains in `TestReadInPathTypes/state_root_symlink`.
The test first opens a valid home, replaces its empty root, and requires refusal.
A directory at `.hive/state/lock` refuses acquisition and runs no callback.
A nested directory named `lock` descends through the public walker.

### V1, V2, and V3 runs

```sh
go test -race -count=1 ./internal/lib/store/ -run 'TestWithLock' -v
go test -race -count=1 ./internal/lib/store/ -run 'TestStoreNames|TestReadIn|TestWriteIn|TestListIn' -v
go test -race -count=1 ./internal/lib/store/ -run 'TestGrouped|TestLegacy|TestWalk|TestCollectWalkNodes|TestOpenRejects|TestOperationsRecheck' -v
```

```text
V1: PASS, exit 0, 11.618s
V2: PASS, exit 0, 1.121s
V3 and path fixes: PASS, exit 0, 2.098s
```

Ten separate `-race -count=1` process-counter runs passed.
Each required 200 increments from four OS processes using `LockedStore.Next`.
The independent API executable checked every held, released, and nil handle method.
It also checked UpdateIn errors, include selection, and completed prefixes under a later lock.
Ten deterministic rounds showed unlocked mixtures and equal grouped versions.
Waiting cancellation and standalone lists under a held lock passed.

### Both prior checklists

| Check group | Current result |
| --- | --- |
| Lock/general checklist | PASS: all 23 earlier items, including corrected walk filtering and fixed-size test strength. |
| Filesystem checklist | PASS: safe/unsafe names, depth, missing paths, static types, external preservation, sizes, JSON, lists, walks, legacy behavior, and held-lock access. |
| General input matrix | PASS: 166 commands, zero failures. |
| Lock-type supplement | PASS: 12 commands, record bytes unchanged. |
| Fresh exported-API filesystem tests | PASS: 25 static path/type combinations plus names/depth/missing, JSON/sizes/filtering, walk, and held-lock groups. |

Static checks cover `.hive`, state root, collection ancestor, lock, and record leaf.
They cover links, dangling links, wrong files/directories, FIFO, and socket where each type is invalid.
External bytes stay unchanged. Dangling targets remain absent. No external state appears.
The size probes accept 1048575 and 1048576 bytes. They reject 1048577 and 5000000 bytes.
Oversize updates run no callback and preserve bytes.
Other probes cover malformed/non-object JSON, 10000 extra keys, selected files, and 500-record lists.
Unprivileged permission checks cover unreadable files, unwritable parents, and untraversable directories.
Each matrix subprocess has a five-second bound. Rejected cases print one plain error without a stack trace.
Static snapshots check unchanged files. The unreadable-file snapshot uses metadata. The untraversable case checks its error only.

One initial ignored verifier assertion incorrectly required rejection of a real directory with a `.json` suffix.
The prior checklist permits that directory to descend.
I corrected only the ignored assertion. The final exported-API race run passed, exit 0.
The artifact retains the initial assertion error. No worker source changed.

### Isolated regression and mutation checks

Current tests with pre-fix `store.go` and `lock.go` from `d9b644f7` compiled and failed intended assertions.
Exit 1. Exact selected evidence:

```text
open_path_test.go:19: Open() = &{/tmp/TestOpenRejectsLinkedHiveAncestor1350681167/001}, <nil>; want store path invalid for /tmp/TestOpenRejectsLinkedHiveAncestor1350681167/001/.hive
walk_test.go:166: WalkRecords() paths = []; want [".probe.tmp-123/one.json"]
walk_test.go:185: WalkRecords() error = <nil>; want store path invalid for /tmp/TestWalkTemporaryPatternChecksTypeFirstlinked_collection_fails2565940013/001/.hive/state/.probe.tmp-123
```

The same run failed linked-state Open, wrong-type Open, and later linked-ancestor access assertions.
The preserved post-Open root refusal test passed on pre-fix source.
The historical red source `3b6c1323` also compiled and failed behavior assertions, exit 1.

The cap mutation changed 1048576 to 2097152 only in an isolated source copy.
Command within that copy:

```sh
go test -race -count=1 ./internal/lib/store -run 'TestReadInSizeBoundary|TestReadAcceptsRecordOfExactlyOneMiB|TestReadRejectsOversizedInvalidAndNonRegularRecords' -v
```

```text
generic_test.go:223: ReadIn("above") error = <nil>; want record too large
store_test.go:276: Read() error = <nil>; want "record too large: /tmp/TestReadRejectsOversizedInvalidAndNonRegularRecords3973075267/001/.hive/state/projects/large.json"
FAIL, exit 1
```

Three existing mutation classes also failed assertions in copies:

- Zero default wait: `returned after 20.698µs; want about two seconds`.
- Shared flock: `counter = 56; want 200`.
- Broad dot-file exclusion: the walk loses expected `counters/.other.json`.

No mutation changed actual task source.

### V4 source and architecture check

```sh
grep -nE 'os\.(Open|OpenFile|ReadFile|WriteFile|Create|CreateTemp)\(' internal/lib/store/*.go
grep -n 'syscall.Flock' internal/lib/store/*.go
grep -n 'io.LimitReader' internal/lib/store/*.go
```

The combined equivalent search passed. I also traced each API through the source.
`readRecordFile` is the sole record reader and reads at most 1048577 bytes.
`writeRecordFile` uses unchanged `atomicWrite`, the sole temporary maker.
The temporary name starts with `.` and contains `.tmp-`.
The lock open and production flock calls stay in `lock.go`.
All changed source paths and imports obey task scope and architecture.
Only `doc.go`, `store.go`, `lock.go`, and adjacent store tests change.
No dependency, baseline, tracker, task, home, or logger source change appears.

### Observable CLI and V5 gate

A fresh ignored binary used only unique marked fixture homes with absolute `--home` paths.
`project list` printed `no projects` with exit 0 for an empty registry.
An array project record produced one `hive: record invalid: <fixture path>` line with exit 1.
Both checks preserved state bytes. Only fixture logs could change.
The API executable emitted `.probe.tmp-123/one.json` for the matching directory, exit 0.
It refused the matching link with its path, exit 1.

| Gate command | Result |
| --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | Empty output, exit 0. Empty output separately required. |
| `go vet ./...` | Exit 0. |
| `golangci-lint run --allow-serial-runners ./...` | `0 issues.`, exit 0. |
| `go test -race -count=1 ./...` | PASS, exit 0. Store: 15.080s. |

### Limits and delivery

No required in-scope check remains uncompleted.
The approved boundary remains a local user and a cooperating writer.
Hostile replacement between Lstat and Open remains deferred for T01.
Home.Resolve/logger behavior remains a separate source follow-up, outside T22 acceptance.
No unrelated probe class, credential inspection, global configuration change, or historical attribution change occurred.
Only this current verdict section forms the verification commit.
The source check before commit found 46 source/module files equal to the tested worker HEAD.
Cleanup removed only this pass's disposable fixtures, binaries, and source copies.
The coordinator owns the next delivery decision. This pass does not merge or update the tracker.

Root report: `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t22-verify-02.md`.
Compact captures and probe sources: the sibling `t22-verify-02-artifacts/` directory.
The report records the verification commit, push, source check, and completed cleanup.

## 2026-10-18 — current verdict — t22-write-cap-verify-01

The current T22 verdict is **PASS** at worker HEAD `a05e617ebb7d431b56bb9f500f3042cb531ddf22`.
Every verdict below applies to that exact worker HEAD.
Earlier verdicts keep their original bytes and attribution.

| Item | Value |
| --- | --- |
| Pass | `t22-write-cap-verify-01` |
| Role | Fresh independent verifier |
| Route | pi, `openai-codex/gpt-6.1-sol:high`, from `PI_*` |
| Branch | `task/T22-write-cap` |
| Base | `7ae3fb8affb7bc577ec1a3aa248d416be2bde340` |
| Tested worker HEAD | `a05e617ebb7d431b56bb9f500f3042cb531ddf22` |
| Source fix | `f049a7b0949520018ad45001c81df689d2fb2fe5` |
| Overall | **PASS** |

The initial worktree was clean at the required HEAD.
The remote task branch matched that HEAD.
The provider, model, and reasoning level matched the brief.
The independent SPEC PASS in `t03-verify-01.md` and `spec-verdict.json` checks the exact base.
That SPEC PASS accepts contract text only.
This pass supplies independent source evidence.

### Scope and report location

I read the complete base-to-worker diff and each changed path.
The diff adds 586 lines across four files.
The only production change adds three lines to `writeRecordFile`.
The guard checks encoded bytes before `atomicWrite`.
The other paths are `write_limit_test.go`, `V2.md`, and `V5.md`.
The test imports use only the standard library.
No production import, dependency, public API, reader, lock, baseline, or title limit changes.
The V2 and V5 historical prefixes remain equal to their base versions.

The actual worker report is `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t22-write-cap-01.md`.
Its final marker is `END WORKER T22-write-cap t22-write-cap-01`.
The worker did not use the planned worktree report path.
This is an output-location deviation, not a source defect.

| Item | Proves | Verdict | Evidence at the tested worker HEAD |
| --- | --- | --- | --- |
| V1 | AC1 | PASS | All existing lock tests pass under the race detector. |
| V2 | AC2 | PASS | Literal byte boundaries, all writer paths, static paths, names, lists, escapes, and UTF-8 pass. |
| V3 | AC3, AC4 | PASS | Grouped reads, retained prefixes, process counters, legacy behavior, and walk tests pass. |
| V4 | AC2 | PASS | Shared bounded reader and guarded writer. Atomic and lock helpers remain unchanged. |
| V5 | AC5 | PASS | Empty format output, vet, lint, and full uncached race gate pass. |
| G1 | Shared writer size | CLOSED | Actual T03 CLI rejects oversized verify and preserves readable prior state. |
| Test strength | Writer size | PASS | Removing only the guard makes new tests fail on assertions. |

### Store tests and independent byte matrix

Command from the actual worktree:

```sh
go test -race -count=1 ./internal/lib/store/ -v
```

```text
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/store	17.237s
exit 0
```

This run includes every existing store test and all five new writer tests.
The expected sizes are fixed literals: `1048575`, `1048576`, `1048577`, and `2097152`.
The tests do not get expectations from `maxRecordSize`.
The literal JSON frames include the final LF.
The escape case counts U+0001 as six encoded bytes and `é` as two UTF-8 bytes.
Exactly 1048576 bytes remain readable.
One added byte fails without a replacement.

An independent matrix adds 48 cases in an ignored source copy.
It checks eight APIs, three literal sizes, and create/replacement modes.
The eight APIs are legacy and generic write/update methods, both standalone and held.
Generic cases use the nested collection `inbox/coordinator`.
The matrix compares exact encoded bytes, including LF.
It checks old-byte equality, readable old values, absent rejected creates, and absent temporary artifacts.
Command inside the isolated candidate overlay:

```sh
go test -race -count=1 ./internal/lib/store/ -run TestCapVerify -v
```

```text
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/store	6.284s
exit 0
```

The worker group test retains the earlier write after a later size rejection.
The existing grouped tests retain completed prefixes after callback errors.
No rollback claim applies.

### Isolated red evidence

The red copy contains the exact worker tree with only the three guard lines removed.
A file comparison checks all 147 tracked paths and finds only that difference.
Actual worktree source never changes.
Command inside the red copy:

```sh
go test -race -count=1 ./internal/lib/store/ -run TestWriteLimit -v
```

The command compiles and exits 1 on assertions.
Below-limit and exact-limit cases still pass.
Create rejection fails at line 174 because each writer returns nil.
Replacement rejection fails at line 275 for all eight writer APIs.
The group assertion fails at line 312.
The escaped-content assertion fails at line 367.
`red-proof.txt` retains the complete output and exact three-line diff.

### Prior T22 checklist

I rechecked all 23 earlier checklist items against current tests, unchanged source, and prior captures.
The prior source is `a3fea5a4304a29e1c24fd8e93b3da30ccf5aa5ad`.
Its production store code differs only by the current guard.

| Checklist items | Current basis |
| --- | --- |
| 1–7: lock and handles | Fresh `TestWithLock*` tests. Prior `api.txt` covers every nil/released method and standalone list. |
| 8–14: names, paths, reader | Fresh generic, Open, Linux type, and literal reader tests. Prior filesystem and 166-command captures cover the broader static matrix. |
| 15–17: groups and processes | Fresh grouped tests, retained-prefix tests, and four-process counter test. Prior API capture checks a later locked read. |
| 18–20: legacy and walk | Fresh legacy, linked-path, temporary-pattern type, nested-lock, and walker tests. |
| 21–23: helpers and scope | Fresh V4 searches and complete diff/import inspection. No dependency or unrelated tracked path changes. |

Prior captures include `t22-verify-02-artifacts/api.txt`, `filesystem.txt`, `lock-types.txt`, and `matrix-summary.txt`.
They cover static `.hive`, state, collection, lock, and record paths.
They include permissions, external-byte preservation, JSON shapes, selected lists, and fixed reader sizes.
These are prior captures, not new executions of that broader matrix.
The current full store run rechecks the existing regression tests.
The earlier F1/F2 and path F01/F02 closures remain valid.
No unrelated probe class applies.

### G1 closure through the actual CLI

The CLI source copy starts at T03 `638c40e339abb329912d7c09b06d609ddc1c69b3`.
Only the verified candidate `store.go` overlays that source before the CLI build.
A comparison checks all 172 tracked paths and finds only this overlay.
The copied source is a probe, not accepted T03 source.
The fixture has an explicit home marker and a separate Git project.

The commands repeat the actual escaped-title and escaped-reason case from `t03-verify-01-captures/probe.py`.
The title contains 120000 U+0001 bytes.
The verification reason contains 120000 U+0002 bytes.
Create, assign, and complete each exit 0.
Create writes 720332 bytes.
The completed prior task has 720290 bytes.
Verify with `--fail` exits 1 with `record too large: <fixture path>`.
Its JSON envelope reports `failed`, phase `transition`, and no effects.
The prior task bytes remain equal.
Every state file remains equal, and no rejected record appears.
Plain and JSON show/list each exit 0 and retain status `completed`.
`g1.py` and `g1.txt` retain the reproducible procedure and measured results.
G1 is CLOSED for the verified store candidate.
T03 F1–F4 remain outside this fix and uncompleted.

### V4 and full gate

Every legacy, generic, and held write/update reaches `writePath`, then `writeRecordFile`.
`marshalRecord` includes LF before the guard sees the data.
Update callbacks supply their encoded bytes to the same guard.
`writeRecordFile` alone calls `atomicWrite`.
`atomicWrite` alone creates record temporary files.
The sole reader uses `Lstat`, then `io.LimitReader` with 1048577 bytes.
The lock open and production flock calls remain in `lock.go`.

| Command from actual worktree | Result |
| --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | Exit 0. Output has zero bytes. |
| `go vet ./...` | Exit 0. |
| `golangci-lint run --allow-serial-runners ./...` | Exit 0. `0 issues.` |
| `go test -race -count=1 ./...` | Exit 0. Every package passes. Store takes 17.378s. |

### Source freeze, cleanup, and limits

Before commit, every tracked path except this verdict equals the exact worker HEAD.
The complete historical verdict prefix remains byte-for-byte equal.
The source, module files, and STE baseline remain unchanged.
Only this dated section enters the verdict commit.
The root report and unique capture folder are ignored outputs.
The report is `/home/penguin/source/hive/.hive/tmp/coordinator/reports/t22-write-cap-verify-01.md`.
The capture folder is the sibling `t22-write-cap-verify-01-captures/`.

Cleanup removes only this pass's marked fixtures, unique binary, and two source copies.
Captures retain commands, assertions, hashes, source-copy checks, and cleanup records.
Existing state, logs, reports, archives, credentials, worktrees, and agents remain unchanged.
No credential content inspection, global configuration change, merge, tracker change, or agent start occurs.
The approved boundary remains a local user and a cooperating writer.
Hostile replacement between `Lstat` and `Open` remains deferred for T01.
No required in-scope check remains uncompleted.
T03 F1–F4, later integration, and landing remain uncompleted.
Recommendation: accept this bounded T22 fix only.
