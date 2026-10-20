# T01-home-and-register VERIFIED

| Item | Value |
| --- | --- |
| Task | T01-home-and-register |
| Commit | 68154626db78a23edd6e64f3de9cdaeec815da05 |
| Date | 2026-10-17 |
| Agent | verifier, pi, openai-codex/gpt-6.1-sol |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | pass | Includes both packages and the new `HIVE_HOME` space test. |
| V2 | AC1, AC6 | pass | All six commands give the required output and status. |
| V3 | AC2, AC6 | pass | Includes both packages. The invalid-ID message gives the pattern and limit. |
| V4 | AC3, AC6 | pass | Sorted records, text fields, JSON results, and missing record match the brief. |
| V5 | AC4 | pass | Process mutations, atomic reads, bounds, and the observed temporary basename pass. |
| V6 | AC5 | pass | Log tests and live events pass. Existing logs remain. |
| V7 | AC7 | pass | All four gate commands pass. The uncached race run also passes. |

## Scope and provenance

The submitted commit stayed at `68154626db78a23edd6e64f3de9cdaeec815da05` throughout the checks.
The branch was `task/T01-home-and-register`. The tracked tree was clean before and after the checks.
The only tracked change after the checks is this file.

Read-only herdr calls showed workspace `wA2`, pane `wA2:p1`, and terminal `term_65e152119960ae`.
Both the pane and workspace showed this worktree path:
`/home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register`.
The environment showed `PI_PROVIDER=openai-codex` and `PI_MODEL=gpt-6.1-sol`.
No agent or workspace started or stopped.

I read the task, evidence format, V1–V7, previous verdict, current worker report, and earlier failure report.
I also read the required instructions, glossaries, untrusted-input reference, tests, and relevant source.

The correction tests exist at `cec1f96f`, before source fix `1c0306c4`.
V3 records assertion failures against the old messages in register and show, in text and JSON.
The green runs use the same tests after the source fix.

V1 explicitly records a temporary `strings.TrimSpace` defect at `1c0306c4`.
The space test fails on its path assertion, not compilation.
The raw transcript also contains a failed `git diff` option. That diagnostic is not red-test evidence.
The test output and the declared temporary mutation are the red-test evidence.

V5 explicitly records a temporary missing-dot defect at `1c0306c4`.
Its red output names the actual temporary file `atomic.json.tmp-1998293767`.
The inotify test watches the directory before `Store.Write` and reads actual creation events afterward.
It does not derive the expected basename from the production helper.

Both temporary defects were restored by the worker. This pass introduced no defect.
These independent comparisons gave no diff:

```sh
git diff --exit-code 7545d132 HEAD -- internal/lib/home/home.go internal/lib/store/atomic.go; echo "exit $?"
git diff --exit-code 1c0306c4 HEAD -- cmd internal main.go; echo "exit $?"
```

```text
exit 0
exit 0
```

Older V-file sections retain historical output and agent attribution. They are not current runs.
Their commit labels do not establish clean trees. In particular, earlier V4 stub output required a temporary tree.
The earlier failure report records that limitation. No current verdict depends on those historical labels.
The accepted task append remains once, followed by the correction-wave entry. Its merge-only change is authorized.

## Runs

All commands below ran from the worktree root unless a command states another directory.
Every selected test run used `-count=1`. The exact gate race command ran first, then an uncached repeat.
The V2 setup omitted deletion because all disposable fixture paths were absent at inventory.
V2, V3, and V4 live commands ran in order.

### V1

```sh
go test ./internal/lib/home/ -run TestResolve -v -count=1; echo "exit $?"
go test ./cmd/ -run TestResolve -v -count=1; echo "exit $?"
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
=== RUN   TestResolveDoesNotTrimEnvHomeValue
--- PASS: TestResolveDoesNotTrimEnvHomeValue (0.00s)
=== RUN   TestResolveMarkerMustBeRegularFile
--- PASS: TestResolveMarkerMustBeRegularFile (0.00s)
=== RUN   TestResolveReturnsAbsolutePathWithoutSymlinks
--- PASS: TestResolveReturnsAbsolutePathWithoutSymlinks (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/home	0.002s
exit 0
=== RUN   TestResolveEmptyHomeFlagHasUsageStatus
--- PASS: TestResolveEmptyHomeFlagHasUsageStatus (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/cmd	0.001s
exit 0
```

### V2

```sh
go build .; echo "exit $?"
./hive --home /tmp/not-a-home project list; echo "exit $?"
./hive --home /tmp/not-a-home --json project list; echo "exit $?"
env HIVE_HOME="$PWD" ./hive project list; echo "exit $?"
./hive --home "$PWD" project list; echo "exit $?"
./hive --home "$PWD" --json project list; echo "exit $?"
./hive --home "$PWD" project show; echo "exit $?"
```

```text
exit 0
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

### V3 tests

```sh
go test ./internal/project/ -run TestRegister -v -count=1; echo "exit $?"
go test ./cmd/ -run TestRegister -v -count=1; echo "exit $?"
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
=== RUN   TestRegisterInvalidIDReportsRuleAndStatus
=== RUN   TestRegisterInvalidIDReportsRuleAndStatus/register_text
=== RUN   TestRegisterInvalidIDReportsRuleAndStatus/register_json
=== RUN   TestRegisterInvalidIDReportsRuleAndStatus/show_text
=== RUN   TestRegisterInvalidIDReportsRuleAndStatus/show_json
--- PASS: TestRegisterInvalidIDReportsRuleAndStatus (0.00s)
    --- PASS: TestRegisterInvalidIDReportsRuleAndStatus/register_text (0.00s)
    --- PASS: TestRegisterInvalidIDReportsRuleAndStatus/register_json (0.00s)
    --- PASS: TestRegisterInvalidIDReportsRuleAndStatus/show_text (0.00s)
    --- PASS: TestRegisterInvalidIDReportsRuleAndStatus/show_json (0.00s)
=== RUN   TestRegisterEmptyPathIsUsageError
--- PASS: TestRegisterEmptyPathIsUsageError (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/cmd	0.002s
exit 0
```

### V3 live commands

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
hive: invalid project id: Repo_One: an ID matches ^[a-z0-9]+(-[a-z0-9]+)*$ and has at most 18 characters
exit 1
exit 0
exit 0
{"outcome":"partial","phase":"dirs","effects":[{"resource":"record","path":".hive/state/projects/repo-two.json","action":"created"}],"error":"create directory projects/repo-two: mkdir /home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register/projects/repo-two: not a directory","result":{"id":"repo-two","path":"/home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register/.hive/tmp/repo-two","base":"main","registered_at":"2026-10-18T03:46:53.364322173Z"}}
exit 1
```

### V4

```sh
./hive --home "$PWD" project list; echo "exit $?"
./hive --home "$PWD" --json project list; echo "exit $?"
./hive --home "$PWD" project show repo-one; echo "exit $?"
./hive --home "$PWD" --json project show repo-one; echo "exit $?"
./hive --home "$PWD" --json project show nope; echo "exit $?"
test "$PWD" = /home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register && rm -rf .hive/state projects .hive/tmp/repo-one .hive/tmp/repo-two; echo "cleanup exit $?"
```

```text
repo-one	/home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register/.hive/tmp/repo-one	main
repo-two	/home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register/.hive/tmp/repo-two	main
exit 0
{"outcome":"ok","phase":"complete","effects":[],"error":null,"result":[{"id":"repo-one","path":"/home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register/.hive/tmp/repo-one","base":"main","registered_at":"2026-10-18T03:46:53.356217609Z"},{"id":"repo-two","path":"/home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register/.hive/tmp/repo-two","base":"main","registered_at":"2026-10-18T03:46:53.364322173Z"}]}
exit 0
id: repo-one
path: /home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register/.hive/tmp/repo-one
base: main
registered_at: 2026-10-18T03:46:53.356217609Z
exit 0
{"outcome":"ok","phase":"complete","effects":[],"error":null,"result":{"id":"repo-one","path":"/home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register/.hive/tmp/repo-one","base":"main","registered_at":"2026-10-18T03:46:53.356217609Z"}}
exit 0
{"outcome":"failed","phase":"read","effects":[],"error":"project not found: nope","result":null}
exit 1
cleanup exit 0
```

### V5

```sh
go test -race ./internal/lib/store/ -run 'TestConcurrent|TestAtomic|TestUpdate|TestRead' -v -count=1; echo "exit $?"
grep -nE 'os\.(Open|OpenFile|ReadFile|WriteFile|Create|CreateTemp)\(' internal/lib/store/*.go; echo "exit $?"
grep -RIn 'readRecordFile\|writeRecordFile\|atomicWrite' internal/lib/store; echo "exit $?"
```

```text
=== RUN   TestAtomicWriteCreatesDotPrefixedTemporaryFile
--- PASS: TestAtomicWriteCreatesDotPrefixedTemporaryFile (0.00s)
=== RUN   TestConcurrentCounterProcesses
--- PASS: TestConcurrentCounterProcesses (1.02s)
=== RUN   TestUpdateCounterStartsAtOne
--- PASS: TestUpdateCounterStartsAtOne (0.00s)
=== RUN   TestConcurrentUpdateProcesses
--- PASS: TestConcurrentUpdateProcesses (1.03s)
=== RUN   TestAtomicWriteReadersSeeCompleteRecord
--- PASS: TestAtomicWriteReadersSeeCompleteRecord (0.01s)
=== RUN   TestUpdateCallbackRunsUnderLock
--- PASS: TestUpdateCallbackRunsUnderLock (0.03s)
=== RUN   TestReadRejectsOversizedInvalidAndNonRegularRecords
=== RUN   TestReadRejectsOversizedInvalidAndNonRegularRecords/array
=== RUN   TestReadRejectsOversizedInvalidAndNonRegularRecords/empty
=== RUN   TestReadRejectsOversizedInvalidAndNonRegularRecords/directory
=== RUN   TestReadRejectsOversizedInvalidAndNonRegularRecords/symlink
=== RUN   TestReadRejectsOversizedInvalidAndNonRegularRecords/large
=== RUN   TestReadRejectsOversizedInvalidAndNonRegularRecords/truncated
--- PASS: TestReadRejectsOversizedInvalidAndNonRegularRecords (0.01s)
    --- PASS: TestReadRejectsOversizedInvalidAndNonRegularRecords/array (0.00s)
    --- PASS: TestReadRejectsOversizedInvalidAndNonRegularRecords/empty (0.00s)
    --- PASS: TestReadRejectsOversizedInvalidAndNonRegularRecords/directory (0.00s)
    --- PASS: TestReadRejectsOversizedInvalidAndNonRegularRecords/symlink (0.00s)
    --- PASS: TestReadRejectsOversizedInvalidAndNonRegularRecords/large (0.00s)
    --- PASS: TestReadRejectsOversizedInvalidAndNonRegularRecords/truncated (0.00s)
=== RUN   TestReadAcceptsRecordOfExactlyOneMiB
--- PASS: TestReadAcceptsRecordOfExactlyOneMiB (0.06s)
=== RUN   TestReadListUsesOnlyValidIDFileNames
--- PASS: TestReadListUsesOnlyValidIDFileNames (0.00s)
=== RUN   TestReadListReturnsNoRecordsAfterInvalidFile
--- PASS: TestReadListReturnsNoRecordsAfterInvalidFile (0.00s)
=== RUN   TestReadAndUpdateRejectUnsafeNames
--- PASS: TestReadAndUpdateRejectUnsafeNames (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/store	3.162s
exit 0
internal/lib/store/atomic.go:10:	file, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-")
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
exit 0
internal/lib/store/store_test.go:20:	data, err := readRecordFile(filepath.Join(home, ".hive", "state", "counters", "shared.json"))
internal/lib/store/atomic.go:8:func atomicWrite(path string, data []byte) error {
internal/lib/store/store.go:38:	data, err := readRecordFile(path)
internal/lib/store/store.go:67:		current, err := readRecordFile(path)
internal/lib/store/store.go:108:		data, err := readRecordFile(filepath.Join(dir, name))
internal/lib/store/store.go:153:	return writeRecordFile(path, data)
internal/lib/store/store.go:180:func readRecordFile(path string) ([]byte, error) {
internal/lib/store/store.go:206:func writeRecordFile(path string, data []byte) error {
internal/lib/store/store.go:210:	return atomicWrite(path, data)
exit 0
```

Only `readRecordFile` opens records. Only `writeRecordFile` calls `atomicWrite`.
The other production open is the lock file. Test matches create fixtures.

### V6

```sh
go test ./internal/lib/log/ -v -count=1; echo "exit $?"
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
{"command":"project list","component":"cli","level":"DEBUG","msg":"command start","time":"2026-10-17T23:46:56.717662636-04:00"}
{"command":"project list","component":"cli","duration_ms":0,"level":"DEBUG","msg":"command end","outcome":"ok","time":"2026-10-17T23:46:56.717797875-04:00"}
exit 0
```

### V7

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go'); echo "gofmt exit $?"
go vet ./...; echo "exit $?"
golangci-lint run ./...; echo "exit $?"
go test -race ./...; echo "exit $?"
go test -race -count=1 ./...; echo "exit $?"
grep -Rn 'spf13/cobra' internal; echo "exit $?"
grep -n '"github.com/Harrison-Blair/hive/cmd"' cmd/common/*.go; echo "exit $?"
```

```text
gofmt exit 0
exit 0
0 issues.
exit 0
ok  	github.com/Harrison-Blair/hive	(cached)
ok  	github.com/Harrison-Blair/hive/cmd	1.008s
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
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/project	(cached)
ok  	github.com/Harrison-Blair/hive/internal/tui	(cached)
exit 0
ok  	github.com/Harrison-Blair/hive	2.327s
ok  	github.com/Harrison-Blair/hive/cmd	1.018s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/version	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.013s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.047s
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/records	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/store	3.147s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.009s
ok  	github.com/Harrison-Blair/hive/internal/project	1.019s
ok  	github.com/Harrison-Blair/hive/internal/tui	1.012s
exit 0
exit 1
exit 1
```

The two grep statuses mean no forbidden import. `gofmt` printed no filename.
The linter reported no lock conflict. The STE baseline did not increase.

## Rule and test mapping

| Rule | Item | Test |
| --- | --- | --- |
| Home precedence, markers, conflict, canonical paths | V1 | `TestResolveFlagWins`, `TestResolveEnvWinsOverAncestor`, `TestResolveOutermostMarkerWins`, `TestResolveFlagSkipsConflictCheck`, marker and symbolic-path cases |
| Empty environment is unset | V1 | `TestResolveEmptyEnvIsUnset` |
| Empty home flag gives status 2 | V1 | `TestResolveEmptyHomeFlagIsUsageError`, cmd `TestResolveEmptyHomeFlagHasUsageStatus` |
| Relative homes and resolved conflict comparison | V1 | `TestResolveRelativeHomePathsAgainstWorkingDirectory`, `TestResolveConflictComparesResolvedPaths` |
| Home values retain spaces | V1 | `TestResolveDoesNotTrimHomeValues`, `TestResolveDoesNotTrimEnvHomeValue` |
| Relative project path retains spaces | V3 | `TestRegisterResolvesRelativePathWithoutTrimming` |
| Empty project path gives status 2 | V3 | cmd `TestRegisterEmptyPathIsUsageError` |
| Invalid base and control path refuse before state access | V3 | `TestRegisterRejectsInvalidBaseBranches`, `TestRegisterRejectsPathControlCharacterBeforePathAccess` |
| ID pattern, 18-character bound, and quoting | V3 | `TestRegisterAcceptsThreePartID`, `TestRegisterRejectsInvalidIDs`, both `TestRegisterShow` tests |
| Register/show text and JSON rule, streams, status, no state change | V3 | cmd `TestRegisterInvalidIDReportsRuleAndStatus` |
| Record, directories, duplicates, existing directories, partial effects | V3 | `TestRegisterCreatesRecordAndProjectDirectories`, duplicate, existing-directory, and partial-directory cases |
| Sorted list, empty array, show fields, missing ID | V4 | `TestListSortsProjectsByID`, `TestListEmptyProjectsIsAnEmptyArray`, `TestShowReturnsOneProject`, `TestShowUnknownProject`, live commands |
| Process counters and updates | V5 | `TestConcurrentCounterProcesses`, `TestConcurrentUpdateProcesses`, `TestUpdateCounterStartsAtOne` |
| Atomic records and locked callback | V5 | `TestAtomicWriteReadersSeeCompleteRecord`, `TestUpdateCallbackRunsUnderLock` |
| Actual temporary basename starts with dot | V5 | `TestAtomicWriteCreatesDotPrefixedTemporaryFile` |
| Oversized, truncated, directory, symlink, empty, array refusal | V5 | `TestReadRejectsOversizedInvalidAndNonRegularRecords` |
| Exactly 1 MiB accepted | V5 | `TestReadAcceptsRecordOfExactlyOneMiB` |
| List filename filter and no partial result on invalid record | V5 | `TestReadListUsesOnlyValidIDFileNames`, `TestReadListReturnsNoRecordsAfterInvalidFile` |
| Unsafe kind, ID, and counter name refusal | V5 | `TestReadAndUpdateRejectUnsafeNames` |
| Single record reader and writer | V5 | Source inspection and access scan above |
| Debug JSON fields, mode, cap, append, concurrency, local rotation | V6 | All six log tests and the live tail |
| Five envelope fields and usage status | V2–V4 | Live commands and `TestRootProjectListJSONEnvelope` in the full gate |
| Formatting, vet, lint, race, dependency direction | V7 | Gate and import searches above |

## Existing probe classes

The probe runner uses a 10-second timeout for each command.
It captures stdout and stderr separately. Each refusal gives one line or one JSON object, without a crash.
Before and after every call, it compares state and project paths, types, modes, symlink targets, and file hashes.
Each comparison passed. Logs are outside that comparison because commands append log events.

The runner and all exact commands, outputs, and statuses remain under the ignored reports directory:

- `.hive/tmp/coordinator/reports/t01-verify-01-probes.py`
- `.hive/tmp/coordinator/reports/t01-verify-01-probes.txt`

Here `H` is the exact worktree path above. `B` is `./hive --home "$PWD"`.
`P` is `$PWD/.hive/tmp/t01-verify-01-fixtures/repo`.
`R` is `$PWD/.hive/state/projects/repo-one.json`.
The probe setup made `P/.git` and one valid project record. All probe results passed.

| Exact command or argument set | Observed output or result | Status |
| --- | --- | --- |
| `./hive --home '' project list` | `hive: --home requires a path` on stderr, empty stdout | 2 |
| `./hive --home '' --json project list` | Same usage line and streams | 2 |
| `./hive --home . project list` | The single `repo-one` row | 0 |
| From `/tmp`: `env HIVE_HOME= "$H/hive" project list` | `hive: not a hive home: no marker above /tmp` | 1 |
| `env HIVE_HOME=. ./hive project list` | `hive: home conflict: HIVE_HOME is /home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register, working directory is in /home/penguin/source/hive` | 1 |
| `B project register "$P" --base ''` | `hive: invalid base branch: ""` | 1 |
| `B project register "$P" --base -x` | `hive: invalid base branch: "-x"` | 1 |
| `B project register "$P" --base 'a b'` | `hive: invalid base branch: "a b"` | 1 |
| `B project register "$P" --base $'bad\nbranch'` | `hive: invalid base branch: "bad\nbranch"` | 1 |
| `B --json project register "$P" --base 'a b'` | `{"outcome":"failed","phase":"validate","effects":[],"error":"invalid base branch: \"a b\"","result":null}` | 1 |
| `B project register $'repo\tx'` | `hive: invalid path: "repo\tx"` | 1 |
| `B project register ''` | `hive: PATH requires a value` | 2 |
| `B project show ../x` and its JSON form | ID rule with value `../x`, no state change | 1 |
| `B project show ''` and its JSON form | ID rule with empty value, no state change | 1 |
| `B project show $'bad\nid'` and its JSON form | ID rule with value `"bad\nid"`, one line | 1 |
| `B project show Repo_One` and its JSON form | Approved complete ID rule | 1 |
| `B project register "$P" --id Repo_One` and its JSON form | Approved complete ID rule | 1 |
| `B project register "$P" --id $'bad\nid'` and its JSON form | Quoted control value plus complete ID rule | 1 |
| `B project list`, `B --json project list`, `B project show repo-one` for each wrong-type record | Directory, symlink to valid record, empty file, and `[]` each give `record invalid: <absolute R>` | 1 |
| `B project list` with garbage in `Bad_ID.json` and `.tmp-x` | Only the valid `repo-one` row | 0 |
| `B --json project show repo-one` with an exactly 1048576-byte valid object | Complete original record returned | 0 |
| `B project show repo-one` with 1048577 bytes | `record too large: <absolute R>` | 1 |

Each invalid-ID text error includes `hive: `, the escaped value, and the exact approved rule suffix.
Each JSON error has that same message without `hive: `. Every refusal has empty effects.
The no-home probe checked that `/.hive` and `/tmp/.hive` were absent before execution.
`/tmp/.hive` remained absent afterward.
No new probe class was added.

## Safety and cleanup

The initial inventory found no binary, state, projects, or disposable repository fixtures.
It found the old report, worker report, worker transcripts, and the existing daily log.
This pass preserved all reports and transcripts. It did not remove `.hive/tmp` or `.hive/logs`.

The earlier `.hive/tmp/verify-report.md` retained this SHA-256:
`e457add95a2f3ec65032e428081c3022600851211f747d7f8fba6a037a3c5843`.
The original 15097-byte prefix of `.hive/logs/2026-10-17.jsonl` retained its hash.
The log grew to 29050 bytes through append operations.

The V4 command removed only fixtures from this pass, with the exact worktree guard.
The probe runner checked the exact worktree before each fixture removal.
It removed its state and unique fixture directory after the checks.
The final command removed only this pass's binary:

```sh
test "$PWD" = /home/penguin/source/hive/projects/hive/worktrees/T01-home-and-register && rm -f hive; echo "binary cleanup exit $?"
find /home/penguin/source/hive/.hive -maxdepth 1 -printf '%y %f\n'; echo "exit $?"
ls -A /home/penguin/source/hive/projects/hive/worktrees; echo "exit $?"
test ! -e .hive/state && test ! -e projects && test ! -e hive && test ! -e .hive/tmp/repo-one && test ! -e .hive/tmp/repo-two && test ! -e .hive/tmp/t01-verify-01-fixtures; echo "fixture absence exit $?"
```

```text
binary cleanup exit 0
d .hive
d tmp
f README.md
exit 0
T01-home-and-register
exit 0
fixture absence exit 0
```

The real home has no product state or logs before or after this pass.
No credential contents or global harness settings were read or changed.
No source, test, V file, task, tracker, handoff, or other worktree changed.
All unique pass transcripts remain under `.hive/tmp/coordinator/reports/t01-verify-01-*`.

## Remarks and recommendation

No unmet acceptance criterion or ordinary reachable safety defect remains in this pass.
The agreed boundary is the local user with cooperating writers.
Hostile replacement between `Lstat` and `Open` remains deferred, not verified.
Broad command error classification and missing log error attributes remain deferred remarks.
These remarks do not change the current verdict.

Recommendation: accept T01 under the agreed boundary. The coordinator owns landing and CP1.
This pass did not land the branch, update the tracker, or run CP1.

END VERIFY T01 verify-01
