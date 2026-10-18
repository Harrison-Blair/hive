# T01-home-and-register VERIFIED

| Item | Value |
| --- | --- |
| Task | T01-home-and-register |
| Commit | d6a046be |
| Date | 2026-10-17 |
| Agent | verifier, Claude Code, claude-opus-5-5 |
| Verdict | fail |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | pass | Each case of the brief passes. Each home rule of the "untrusted inputs" entry has a test (see Rules). The home probes pass. Remarks: the main red run has the line `commit 45b1bf58`, the branch point, where the test file is not committed, so the run used an uncommitted tree. The green run is at 31540fa, older than d6a046be. The cmd test `TestResolveEmptyHomeFlagHasUsageStatus` matches `TestResolve` but the item command (`./internal/lib/home/`) does not run it. No test checks that a `HIVE_HOME` value is not trimmed; `TestResolveDoesNotTrimHomeValues` checks only `--home`. |
| V2 | AC1, AC6 | pass | The six lines match the expected output. A failing command prints nothing on stdout without `--json` and nothing on stderr with `--json`. Remarks: the worker run uses `2>&1`. The red run and green run of `./cmd -run TestRoot` are extra, and the green run uses a wider `-run` pattern than the red run. |
| V3 | AC2, AC6 | fail | AC2 rule text is missing. Line four `./hive --home "$PWD" project register .hive/tmp/repo-one --id Repo_One` prints only `hive: invalid project id: Repo_One` and `exit 1`. It does not print the rule (the pattern `^[a-z0-9]+(-[a-z0-9]+)*$` and the bound of 18 characters). With `--json` the envelope error is also only `invalid project id: Repo_One`. Each other line matches. Each `--base`, `PATH`, and `project show` rule has a test, and each register and show probe passes. Remarks: `TestRegisterShowRejectsInvalidIDBeforeStoreAccess` and `TestRegisterShowEscapesControlCharactersInInvalidID` pass in the red run. The cmd test `TestRegisterEmptyPathIsUsageError` covers the empty `PATH` rule, but the item command (`./internal/project/`) does not run it. The Result sentence says that the tests prove the ID rule, but the rule text of AC2 is not printed. |
| V4 | AC3, AC6 | pass | The five commands match the expected output. Remarks: the "Invalid-ID red run" has the line `commit 31540fa`, but its output shows `stub` errors that the code at 31540fa cannot give, so the commit line is wrong. The Run block ends with `ls -A .hive/tmp`, not with `echo "exit $?"`. The not-found envelope has `phase` `read`. |
| V5 | AC4 | fail | The rule "The temporary file of an atomic write starts with `.`" has no test in the selection `go test -race ./internal/lib/store/ -run 'TestConcurrent|TestAtomic|TestUpdate|TestRead'`. `TestAtomicWriteReadersSeeCompleteRecord` checks only that no temporary file remains after the writes. No test in the package checks the name of the temporary file. The code does use the prefix (`atomic.go:10`), and the probe with `.tmp-x` passes, but a test that fails when the prefix is removed does not exist. Each other store rule has a test, the grep block meets the rule, and each store probe passes. Remarks: `TestReadAndUpdateRejectUnsafeNames` passes in the red run. `readRecordFile` calls `os.Lstat` and then `os.Open` without `O_NOFOLLOW`, so a symbolic link put in place between the two calls is followed. |
| V6 | AC5 | pass | The tests and the log tail match the expected output. Remarks: the Run block ends with `tail`, not with `echo "exit $?"`. |
| V7 | AC7 | pass | `gofmt -l` prints nothing. `go vet`, `golangci-lint`, and `go test -race` pass, also with `-count=1`. `internal/` does not import Cobra. `cmd/common` does not import `cmd`. Remarks: the worker run is at a24ea1f, older than d6a046be (d6a046be adds only evidence files). The worker test output shows `(cached)` lines. |

Other remarks:

- Scope: each path of `git diff --name-only 45b1bf58 d6a046be` is in the Source files table, a new `_test.go` file beside such a file (`cmd/project_root_test.go` beside `cmd/root.go`), an evidence file, or the brief. The worker appended one line to the "Changes" section of the brief: "2026-10-17, implementation choices. The record reader is `readRecordFile`. The record writer is `writeRecordFile`." The tracker lets a worker edit only the files that its brief names plus new tests, so this line is out of the worker's scope. The verifier did not revert it.
- Red runs: each test item has a red run that fails on assertions of new tests. No red run fails only on a compile error.

## Runs

### Preflight

```sh
git rev-parse --short=8 HEAD; git status --porcelain; echo "status exit $?"
ls -A /home/penguin/source/hive/.hive; echo ---; ls -A /home/penguin/source/hive/projects/hive/worktrees
git diff --name-only 45b1bf58 d6a046be; echo "exit $?"
test "$PWD" = /home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register && rm -rf .hive/state .hive/logs .hive/tmp && go build .; echo "exit $?"
```

```text
d6a046be
status exit 0
README.md
tmp
---
T01-home-and-register
cmd/common/common.go
cmd/project/list/list.go
cmd/project/project.go
cmd/project/register/register.go
cmd/project/show/show.go
cmd/project_root_test.go
cmd/root.go
internal/lib/home/home.go
internal/lib/home/home_test.go
internal/lib/log/log.go
internal/lib/log/log_test.go
internal/lib/output/output.go
internal/lib/records/doc.go
internal/lib/records/project.go
internal/lib/store/atomic.go
internal/lib/store/doc.go
internal/lib/store/store.go
internal/lib/store/store_test.go
internal/project/doc.go
internal/project/list.go
internal/project/list_show_test.go
internal/project/register.go
internal/project/register_test.go
internal/project/render.go
internal/project/show.go
main.go
spec/hive-v1/evidence/T01-home-and-register/V1.md
spec/hive-v1/evidence/T01-home-and-register/V2.md
spec/hive-v1/evidence/T01-home-and-register/V3.md
spec/hive-v1/evidence/T01-home-and-register/V4.md
spec/hive-v1/evidence/T01-home-and-register/V5.md
spec/hive-v1/evidence/T01-home-and-register/V6.md
spec/hive-v1/evidence/T01-home-and-register/V7.md
spec/hive-v1/tasks/T01-home-and-register.md
exit 0
exit 0
```

### V1

The first run was cached, so the verifier ran it again with `-count=1`. The two outputs have the same test lines.

```sh
go test ./internal/lib/home/ -run TestResolve -v; echo "exit $?"
go test ./internal/lib/home/ -run TestResolve -v -count=1; echo "exit $?"
```

```text
=== RUN   TestResolveFlagWins
--- PASS: TestResolveFlagWins (0.00s)
=== RUN   TestResolveEnvWinsOverAncestor
--- PASS: TestResolveEnvWinsOverAncestor (0.00s)
=== RUN   TestResolveOutermostMarkerWins
--- PASS: TestResolveOutermostMarkerWins (0.00s)
=== RUN   TestResolveFlagSkipsConflictCheck
--- PASS: TestResolveFlagSkipsConflictCheck (0.00s)
=== RUN   TestResolveNamedPathWithoutMarker
--- PASS: TestResolveNamedPathWithoutMarker (0.00s)
=== RUN   TestResolveEnvPathWithoutMarker
--- PASS: TestResolveEnvPathWithoutMarker (0.00s)
=== RUN   TestResolveNoMarker
--- PASS: TestResolveNoMarker (0.00s)
=== RUN   TestResolveEmptyHomeFlagIsUsageError
--- PASS: TestResolveEmptyHomeFlagIsUsageError (0.00s)
=== RUN   TestResolveEmptyEnvIsUnset
--- PASS: TestResolveEmptyEnvIsUnset (0.00s)
=== RUN   TestResolveRelativeHomePathsAgainstWorkingDirectory
--- PASS: TestResolveRelativeHomePathsAgainstWorkingDirectory (0.00s)
=== RUN   TestResolveConflictComparesResolvedPaths
--- PASS: TestResolveConflictComparesResolvedPaths (0.00s)
=== RUN   TestResolveDoesNotTrimHomeValues
--- PASS: TestResolveDoesNotTrimHomeValues (0.00s)
=== RUN   TestResolveMarkerMustBeRegularFile
--- PASS: TestResolveMarkerMustBeRegularFile (0.00s)
=== RUN   TestResolveReturnsAbsolutePathWithoutSymlinks
--- PASS: TestResolveReturnsAbsolutePathWithoutSymlinks (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
exit 0
[second run: the same 14 PASS lines]
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/home	0.002s
exit 0
```

### V2

```sh
test "$PWD" = /home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register && go build . && rm -rf .hive/state projects .hive/tmp/repo-one .hive/tmp/repo-two; echo "setup exit $?"
./hive --home /tmp/not-a-home project list; echo "exit $?"
./hive --home /tmp/not-a-home --json project list; echo "exit $?"
HIVE_HOME="$PWD" ./hive project list; echo "exit $?"
./hive --home "$PWD" project list; echo "exit $?"
./hive --home "$PWD" --json project list; echo "exit $?"
./hive --home "$PWD" project show; echo "exit $?"
```

```text
setup exit 0
hive: not a hive home: /tmp/not-a-home
exit 1
{"outcome":"failed","phase":"home","effects":[],"error":"not a hive home: /tmp/not-a-home","result":null}
exit 1
hive: home conflict: HIVE_HOME is /home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register, working directory is in /home/penguin/source/hive
exit 1
no projects
exit 0
{"outcome":"ok","phase":"complete","effects":[],"error":null,"result":[]}
exit 0
hive: accepts 1 arg(s), received 0
exit 2
```

Stream check (byte counts of the stream that must be empty):

```sh
./hive --home /tmp/not-a-home project list 2>/dev/null | wc -c
./hive --home /tmp/not-a-home --json project list 2>&1 >/dev/null | wc -c
./hive --home "$PWD" project show 2>/dev/null | wc -c
./hive --home "$PWD" --json project show 2>/dev/null | wc -c
./hive --home "$PWD" --json project show; echo "exit $?"
```

```text
0
0
0
0
hive: accepts 1 arg(s), received 0
exit 2
```

### V3

```sh
go test ./internal/project/ -run TestRegister -v; echo "exit $?"
go test ./internal/project/ -run TestRegister -v -count=1; echo "exit $?"
```

```text
=== RUN   TestRegisterShowRejectsInvalidIDBeforeStoreAccess
--- PASS: TestRegisterShowRejectsInvalidIDBeforeStoreAccess (0.00s)
=== RUN   TestRegisterShowEscapesControlCharactersInInvalidID
--- PASS: TestRegisterShowEscapesControlCharactersInInvalidID (0.00s)
=== RUN   TestRegisterCreatesRecordAndProjectDirectories
--- PASS: TestRegisterCreatesRecordAndProjectDirectories (0.00s)
=== RUN   TestRegisterAcceptsGitFile
--- PASS: TestRegisterAcceptsGitFile (0.00s)
=== RUN   TestRegisterAcceptsThreePartID
--- PASS: TestRegisterAcceptsThreePartID (0.00s)
=== RUN   TestRegisterRejectsInvalidIDs
=== RUN   TestRegisterRejectsInvalidIDs/ab_c
=== RUN   TestRegisterRejectsInvalidIDs/-ab
=== RUN   TestRegisterRejectsInvalidIDs/abcdefghijklmnopqrs
=== RUN   TestRegisterRejectsInvalidIDs/#00
=== RUN   TestRegisterRejectsInvalidIDs/bad_id
=== RUN   TestRegisterRejectsInvalidIDs/bad_id#01
--- PASS: TestRegisterRejectsInvalidIDs (0.00s)
    --- PASS: TestRegisterRejectsInvalidIDs/ab_c (0.00s)
    --- PASS: TestRegisterRejectsInvalidIDs/-ab (0.00s)
    --- PASS: TestRegisterRejectsInvalidIDs/abcdefghijklmnopqrs (0.00s)
    --- PASS: TestRegisterRejectsInvalidIDs/#00 (0.00s)
    --- PASS: TestRegisterRejectsInvalidIDs/bad_id (0.00s)
    --- PASS: TestRegisterRejectsInvalidIDs/bad_id#01 (0.00s)
=== RUN   TestRegisterRejectsInvalidBaseBranches
=== RUN   TestRegisterRejectsInvalidBaseBranches/#00
=== RUN   TestRegisterRejectsInvalidBaseBranches/-main
=== RUN   TestRegisterRejectsInvalidBaseBranches/feature_branch
=== RUN   TestRegisterRejectsInvalidBaseBranches/bad_branch
--- PASS: TestRegisterRejectsInvalidBaseBranches (0.00s)
    --- PASS: TestRegisterRejectsInvalidBaseBranches/#00 (0.00s)
    --- PASS: TestRegisterRejectsInvalidBaseBranches/-main (0.00s)
    --- PASS: TestRegisterRejectsInvalidBaseBranches/feature_branch (0.00s)
    --- PASS: TestRegisterRejectsInvalidBaseBranches/bad_branch (0.00s)
=== RUN   TestRegisterRejectsPathControlCharacterBeforePathAccess
--- PASS: TestRegisterRejectsPathControlCharacterBeforePathAccess (0.00s)
=== RUN   TestRegisterResolvesSymbolicPath
--- PASS: TestRegisterResolvesSymbolicPath (0.00s)
=== RUN   TestRegisterResolvesRelativePathWithoutTrimming
--- PASS: TestRegisterResolvesRelativePathWithoutTrimming (0.00s)
=== RUN   TestRegisterRejectsNonGitRepository
--- PASS: TestRegisterRejectsNonGitRepository (0.00s)
=== RUN   TestRegisterRejectsSymbolicGitMarker
--- PASS: TestRegisterRejectsSymbolicGitMarker (0.00s)
=== RUN   TestRegisterRejectsDuplicateIDAndPath
--- PASS: TestRegisterRejectsDuplicateIDAndPath (0.00s)
=== RUN   TestRegisterReportsPartialDirectoryFailure
--- PASS: TestRegisterReportsPartialDirectoryFailure (0.00s)
=== RUN   TestRegisterExistingDirectoriesRemainAndReportExists
--- PASS: TestRegisterExistingDirectoriesRemainAndReportExists (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/project	0.002s
exit 0
[second run with -count=1: the same PASS lines]
PASS
ok  	github.com/Harrison-Blair/hive/internal/project	0.002s
exit 0
```

The cmd tests whose names match the `-run` pattern. The item command does not run them.

```sh
go test ./cmd/ -run 'TestResolve|TestRegister' -v -count=1; echo "exit $?"
```

```text
=== RUN   TestRegisterEmptyPathIsUsageError
--- PASS: TestRegisterEmptyPathIsUsageError (0.00s)
=== RUN   TestResolveEmptyHomeFlagHasUsageStatus
--- PASS: TestResolveEmptyHomeFlagHasUsageStatus (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/cmd	0.001s
exit 0
```

```sh
git init -q .hive/tmp/repo-one; echo "exit $?"
./hive --home "$PWD" project register .hive/tmp/repo-one --base main; echo "exit $?"
ls projects/repo-one; echo "exit $?"
./hive --home "$PWD" project register .hive/tmp/repo-one; echo "exit $?"
./hive --home "$PWD" project register .hive/tmp/repo-one --id other; echo "exit $?"
./hive --home "$PWD" project register .hive/tmp/repo-one --id Repo_One; echo "exit $?"
git init -q .hive/tmp/repo-two; echo "exit $?"
touch projects/repo-two; echo "exit $?"
./hive --home "$PWD" --json project register .hive/tmp/repo-two; echo "exit $?"
```

```text
exit 0
created record .hive/state/projects/repo-one.json
created directory projects/repo-one
created directory projects/repo-one/worktrees
created directory projects/repo-one/overrides
exit 0
overrides
worktrees
exit 0
hive: project exists: repo-one
exit 1
hive: path registered as repo-one
exit 1
hive: invalid project id: Repo_One
exit 1
exit 0
exit 0
{"outcome":"partial","phase":"dirs","effects":[{"resource":"record","path":".hive/state/projects/repo-two.json","action":"created"}],"error":"create directory projects/repo-two: mkdir /home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register/projects/repo-two: not a directory","result":{"id":"repo-two","path":"/home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register/.hive/tmp/repo-two","base":"main","registered_at":"2026-10-18T02:54:06.778916224Z"}}
exit 1
```

The fail of V3: the invalid-ID output has no rule text on stdout or stderr, and none in the JSON error.

```sh
./hive --home "$PWD" project register .hive/tmp/repo-one --id Repo_One 2>/dev/null | od -c | head; echo "stdout exit ${PIPESTATUS[0]}"
./hive --home "$PWD" --json project register .hive/tmp/repo-one --id Repo_One; echo "exit $?"
```

```text
0000000
stdout exit 1
{"outcome":"failed","phase":"validate","effects":[],"error":"invalid project id: Repo_One","result":null}
exit 1
```

### V4

```sh
./hive --home "$PWD" project list; echo "exit $?"
./hive --home "$PWD" --json project list; echo "exit $?"
./hive --home "$PWD" project show repo-one; echo "exit $?"
./hive --home "$PWD" --json project show repo-one; echo "exit $?"
./hive --home "$PWD" --json project show nope; echo "exit $?"
test "$PWD" = /home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register && rm -rf .hive/state projects .hive/tmp/repo-one .hive/tmp/repo-two; echo "exit $?"
```

```text
repo-one	/home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register/.hive/tmp/repo-one	main
repo-two	/home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register/.hive/tmp/repo-two	main
exit 0
{"outcome":"ok","phase":"complete","effects":[],"error":null,"result":[{"id":"repo-one","path":"/home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register/.hive/tmp/repo-one","base":"main","registered_at":"2026-10-18T02:54:06.770921607Z"},{"id":"repo-two","path":"/home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register/.hive/tmp/repo-two","base":"main","registered_at":"2026-10-18T02:54:06.778916224Z"}]}
exit 0
id: repo-one
path: /home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register/.hive/tmp/repo-one
base: main
registered_at: 2026-10-18T02:54:06.770921607Z
exit 0
{"outcome":"ok","phase":"complete","effects":[],"error":null,"result":{"id":"repo-one","path":"/home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register/.hive/tmp/repo-one","base":"main","registered_at":"2026-10-18T02:54:06.770921607Z"}}
exit 0
{"outcome":"failed","phase":"read","effects":[],"error":"project not found: nope","result":null}
exit 1
exit 0
```

### V5

```sh
go test -race ./internal/lib/store/ -run 'TestConcurrent|TestAtomic|TestUpdate|TestRead' -v; echo "exit $?"
```

```text
=== RUN   TestConcurrentCounterProcesses
--- PASS: TestConcurrentCounterProcesses (1.02s)
=== RUN   TestUpdateCounterStartsAtOne
--- PASS: TestUpdateCounterStartsAtOne (0.00s)
=== RUN   TestConcurrentUpdateProcesses
--- PASS: TestConcurrentUpdateProcesses (1.02s)
=== RUN   TestAtomicWriteReadersSeeCompleteRecord
--- PASS: TestAtomicWriteReadersSeeCompleteRecord (0.01s)
=== RUN   TestUpdateCallbackRunsUnderLock
--- PASS: TestUpdateCallbackRunsUnderLock (0.04s)
=== RUN   TestReadRejectsOversizedInvalidAndNonRegularRecords
=== RUN   TestReadRejectsOversizedInvalidAndNonRegularRecords/empty
=== RUN   TestReadRejectsOversizedInvalidAndNonRegularRecords/directory
=== RUN   TestReadRejectsOversizedInvalidAndNonRegularRecords/symlink
=== RUN   TestReadRejectsOversizedInvalidAndNonRegularRecords/large
=== RUN   TestReadRejectsOversizedInvalidAndNonRegularRecords/truncated
=== RUN   TestReadRejectsOversizedInvalidAndNonRegularRecords/array
--- PASS: TestReadRejectsOversizedInvalidAndNonRegularRecords (0.01s)
    --- PASS: TestReadRejectsOversizedInvalidAndNonRegularRecords/empty (0.00s)
    --- PASS: TestReadRejectsOversizedInvalidAndNonRegularRecords/directory (0.00s)
    --- PASS: TestReadRejectsOversizedInvalidAndNonRegularRecords/symlink (0.00s)
    --- PASS: TestReadRejectsOversizedInvalidAndNonRegularRecords/large (0.00s)
    --- PASS: TestReadRejectsOversizedInvalidAndNonRegularRecords/truncated (0.00s)
    --- PASS: TestReadRejectsOversizedInvalidAndNonRegularRecords/array (0.00s)
=== RUN   TestReadAcceptsRecordOfExactlyOneMiB
--- PASS: TestReadAcceptsRecordOfExactlyOneMiB (0.05s)
=== RUN   TestReadListUsesOnlyValidIDFileNames
--- PASS: TestReadListUsesOnlyValidIDFileNames (0.00s)
=== RUN   TestReadListReturnsNoRecordsAfterInvalidFile
--- PASS: TestReadListReturnsNoRecordsAfterInvalidFile (0.00s)
=== RUN   TestReadAndUpdateRejectUnsafeNames
--- PASS: TestReadAndUpdateRejectUnsafeNames (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/store	3.153s
exit 0
```

```sh
grep -nE 'os\.(Open|OpenFile|ReadFile|WriteFile|Create|CreateTemp)\(' internal/lib/store/*.go; echo "exit $?"
```

```text
internal/lib/store/store.go:168:	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
internal/lib/store/store.go:188:	file, err := os.Open(path)
internal/lib/store/store_test.go:231:	if err := os.WriteFile(large, []byte(`{"padding":"`+strings.Repeat("x", maxRecordSize)+`"}`), 0o644); err != nil {
internal/lib/store/store_test.go:235:	if err := os.WriteFile(truncated, []byte(`{"id":`), 0o644); err != nil {
internal/lib/store/store_test.go:239:	if err := os.WriteFile(empty, nil, 0o644); err != nil {
internal/lib/store/store_test.go:243:	if err := os.WriteFile(array, []byte(`[]`), 0o644); err != nil {
internal/lib/store/store_test.go:250:	if err := os.WriteFile(target, []byte(`{"id":"target"}`), 0o644); err != nil {
internal/lib/store/store_test.go:288:	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
internal/lib/store/store_test.go:314:		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0o644); err != nil {
internal/lib/store/store_test.go:334:	if err := os.WriteFile(filepath.Join(dir, "alpha.json"), []byte(`{"id":"alpha"}`), 0o644); err != nil {
internal/lib/store/store_test.go:338:	if err := os.WriteFile(invalid, []byte(`{"id":`), 0o644); err != nil {
internal/lib/store/atomic.go:10:	file, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-")
exit 0
```

Grep decision: the nine `store_test.go` matches are test fixtures and do not count. `store.go:168` is in `withLock` and opens `.hive/state/lock`, which is correct. `store.go:188` is in `readRecordFile` (lines 180 to 204). `atomic.go:10` is in `atomicWrite`, which only `writeRecordFile` calls. The grep block meets the rule.

Search for a test of the temporary-file prefix (the fail of V5):

```sh
grep -rn 'tmp\|HasPrefix\|"\."' --include=*_test.go internal cmd
```

```text
cmd/project_root_test.go:47:	if got := stderr.String(); !strings.HasPrefix(got, "hive: not a hive home: no marker above ") || strings.Count(got, "\n") != 1 {
cmd/project_root_test.go:76:	if stdout.Len() != 0 || strings.Count(stderr.String(), "\n") != 1 || !strings.HasPrefix(stderr.String(), "hive: ") {
internal/lib/log/log_test.go:33:	if !ok || !strings.HasPrefix(eventTime, "2026-01-02T03:04:05") {
internal/lib/log/log_test.go:136:	if !strings.HasPrefix(string(data), "existing\n") || strings.Count(string(data), "\n") != 2 {
```

### V6

```sh
go test ./internal/lib/log/ -v; echo "exit $?"
./hive --home "$PWD" project list >/dev/null; echo "exit $?"
tail -n 2 .hive/logs/"$(date +%F)".jsonl; echo "exit $?"
```

```text
=== RUN   TestBodyCapsAt4096BytesAndMarksTruncated
--- PASS: TestBodyCapsAt4096BytesAndMarksTruncated (0.00s)
=== RUN   TestBodyAtLimitDoesNotMarkTruncated
--- PASS: TestBodyAtLimitDoesNotMarkTruncated (0.00s)
=== RUN   TestConcurrentLogEventsAreWholeJSONLines
--- PASS: TestConcurrentLogEventsAreWholeJSONLines (0.00s)
=== RUN   TestOpenAppendsToDailyFile
--- PASS: TestOpenAppendsToDailyFile (0.00s)
=== RUN   TestWriterRotatesAtLocalMidnight
--- PASS: TestWriterRotatesAtLocalMidnight (0.00s)
=== RUN   TestLogFileNamesIncludeTheLocalDate
--- PASS: TestLogFileNamesIncludeTheLocalDate (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/log	0.004s
exit 0
exit 0
{"command":"project list","component":"cli","level":"DEBUG","msg":"command start","time":"2026-10-17T22:54:56.930143577-04:00"}
{"command":"project list","component":"cli","duration_ms":0,"level":"DEBUG","msg":"command end","outcome":"ok","time":"2026-10-17T22:54:56.93029355-04:00"}
exit 0
```

The run did not cross midnight (`date` printed `Sat Oct 17 22:54:56 EDT 2026`). The test file asserts a 5000-byte body cut to 4096 bytes with `truncated` true (log_test.go:23-30), a line count equal to the event count of 8 writers with 100 events each (log_test.go:71-108), and one event before and one after midnight in the files 2026-01-02 and 2026-01-03 (log_test.go:149-157).

### V7

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go'); echo "gofmt exit $?"
go vet ./...; echo "vet exit $?"
golangci-lint run ./...; echo "lint exit $?"
go test -race ./...; echo "test exit $?"
go test -race -count=1 ./...; echo "test exit $?"
grep -rln 'spf13/cobra' internal || echo "internal: no cobra import"
grep -n '"github.com/Harrison-Blair/hive/cmd"' cmd/common/*.go || echo "cmd/common: no cmd import"
```

```text
gofmt exit 0
vet exit 0
0 issues.
lint exit 0
ok  	github.com/Harrison-Blair/hive	2.312s
ok  	github.com/Harrison-Blair/hive/cmd	1.007s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/version	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/log	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/records	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/store	(cached)
ok  	github.com/Harrison-Blair/hive/internal/project	(cached)
ok  	github.com/Harrison-Blair/hive/internal/tui	(cached)
?   	github.com/Harrison-Blair/hive/internal/version	[no test files]
test exit 0
ok  	github.com/Harrison-Blair/hive	2.321s
ok  	github.com/Harrison-Blair/hive/cmd	1.015s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/version	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.012s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.044s
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/records	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/store	3.145s
ok  	github.com/Harrison-Blair/hive/internal/project	1.018s
ok  	github.com/Harrison-Blair/hive/internal/tui	1.011s
?   	github.com/Harrison-Blair/hive/internal/version	[no test files]
test exit 0
internal: no cobra import
cmd/common: no cmd import
```

golangci-lint did not print `parallel golangci-lint is running`, so one run was sufficient.

## Rules

| Rule of the "untrusted inputs" entry | Item | Test that covers it |
| --- | --- | --- |
| An empty `HIVE_HOME` counts as unset. | V1 | `TestResolveEmptyEnvIsUnset` |
| An empty `--home` value is a usage error with exit status 2. | V1 | `TestResolveEmptyHomeFlagIsUsageError`; also `cmd` `TestResolveEmptyHomeFlagHasUsageStatus` (not run by the V1 command) |
| An empty `PATH` is a usage error with exit status 2. | V3 | `cmd` `TestRegisterEmptyPathIsUsageError` (not run by the V3 command) |
| A relative `--home` or `HIVE_HOME` resolves against the working directory. | V1 | `TestResolveRelativeHomePathsAgainstWorkingDirectory` |
| A relative `PATH` resolves against the working directory. | V3 | `TestRegisterResolvesRelativePathWithoutTrimming` |
| Hive does not trim spaces around a value. | V1, V3 | `TestResolveDoesNotTrimHomeValues` (`--home` only), `TestRegisterResolvesRelativePathWithoutTrimming` (`PATH`); no test for `HIVE_HOME` (remark) |
| The conflict check compares resolved paths. | V1 | `TestResolveConflictComparesResolvedPaths` |
| A bad `--base` (empty, starts with `-`, space, control character) fails with `invalid base branch: <%q value>`. | V3 | `TestRegisterRejectsInvalidBaseBranches` cases `#00` (empty), `-main`, `feature_branch`, `bad_branch` (newline) |
| A `PATH` with a control character fails with `invalid path: <%q value>` before any file access. | V3 | `TestRegisterRejectsPathControlCharacterBeforePathAccess` |
| `project show` applies the ID rule before it reads a file. | V3 | `TestRegisterShowRejectsInvalidIDBeforeStoreAccess`, `TestRegisterShowEscapesControlCharactersInInvalidID` |
| The store refuses a kind, an ID, or a counter name that is empty or holds `/` or `..`. | V5 | `TestReadAndUpdateRejectUnsafeNames` |
| The temporary file of an atomic write starts with `.`. | V5 | none (fail) |
| `List` reads only `<id>.json` files where `<id>` passes the ID rule. | V5 | `TestReadListUsesOnlyValidIDFileNames` |
| A directory, a symbolic link, an empty file, or a JSON array gives `record invalid: <path>`. | V5 | `TestReadRejectsOversizedInvalidAndNonRegularRecords` cases `directory`, `symlink`, `empty`, `array` |
| The reader does not follow a symbolic link. | V5 | `TestReadRejectsOversizedInvalidAndNonRegularRecords/symlink` (the target is a valid record) |
| A record of exactly 1 MiB is valid. | V5 | `TestReadAcceptsRecordOfExactlyOneMiB` |
| One function reads a record file and one writes one. | V5 | the grep block of V5 (meets the rule) |

## Probes

Before and after each failing probe, `find .hive/state projects 2>&1 | sort` printed the same lines. Only `.hive/logs` changed.

| Probe | Command | Result | Pass or fail |
| --- | --- | --- | --- |
| Empty `--home` | `./hive --home "" project list` | `hive: --home requires a path`, exit 2. With `--json`: the same stderr line, exit 2. | pass |
| Empty `HIVE_HOME` | `ls -A /.hive /tmp/.hive 2>&1`; `env -C /tmp HIVE_HOME= "$PWD/hive" project list` | Both `ls` lines say `No such file or directory`. `hive: not a hive home: no marker above /tmp`, exit 1. `/tmp/.hive` still absent after. | pass |
| Relative `--home` | `./hive --home . project list` | `no projects`, exit 0 | pass |
| `--base -x` | `./hive --home "$PWD" project register .hive/tmp/repo-one --base -x` | `hive: invalid base branch: "-x"`, exit 1, no change | pass |
| `--base "a b"` | `... --base "a b"` | `hive: invalid base branch: "a b"`, exit 1, no change. With `--json`: `{"outcome":"failed","phase":"validate","effects":[],"error":"invalid base branch: \"a b\"","result":null}`, exit 1 | pass |
| `--base ""` | `... --base ""` | `hive: invalid base branch: ""`, exit 1, no change | pass |
| Control character in `PATH` | `./hive --home "$PWD" project register "$(printf 'repo\tx')"` | `hive: invalid path: "repo\tx"`, exit 1, no change | pass |
| Empty `PATH` | `./hive --home "$PWD" project register ""` | `hive: PATH requires a value`, exit 2, no change | pass |
| Show `../x` | `./hive --home "$PWD" project show ../x` | `hive: invalid project id: ../x`, exit 1, no change | pass |
| Show `""` | `./hive --home "$PWD" project show ""` | stdout empty; stderr `hive: invalid project id: ` (with the trailing space); exit 1, no change | pass |
| R is a directory | `rm R && mkdir R`; `project list`, `--json project list`, `project show repo-one` | Each: `record invalid: <abs R>`, exit 1. JSON `outcome` `failed`, `phase` `store`. No change. | pass |
| R is a symbolic link to a valid record | `rm R && ln -s "$PWD/.hive/tmp/probe-record.json" R`; the same three commands | The same as the row above | pass |
| R is empty | `: > R`; the same three commands | The same as the row above | pass |
| R is `[]` | `printf '[]' > R`; the same three commands | The same as the row above | pass |
| `List` filter | `garbage` in `.hive/state/projects/Bad_ID.json` and `.hive/state/projects/.tmp-x`; `./hive --home "$PWD" project list` | Only the `repo-one` line, exit 0 | pass |

R is `.hive/state/projects/repo-one.json`, and `<abs R>` is `/home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register/.hive/state/projects/repo-one.json`. After each R probe, `rm -rf R && cp .hive/tmp/probe-record.json R` (guarded form) printed `restore exit 0`.

### Cleanup

```sh
test "$PWD" = /home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register && rm -rf .hive/state projects .hive/tmp/repo-one .hive/tmp/repo-two .hive/tmp/probe-* hive; echo "exit $?"
ls -A /home/penguin/source/hive/.hive; echo ---; ls -A /home/penguin/source/hive/projects/hive/worktrees
```

```text
exit 0
README.md
tmp
---
T01-home-and-register
```
