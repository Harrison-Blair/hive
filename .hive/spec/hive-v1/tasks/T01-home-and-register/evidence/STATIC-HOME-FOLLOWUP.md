# T01-home-and-register static home follow-up

| Item | Value |
| --- | --- |
| Task | T01-home-and-register |
| Pass | t01-static-home-fix-01 |
| Proves | AC1, AC6, AC7 |
| Commit | 72c7ac48 |
| Base | 59bfb8fe100d17aba83c3a657e3c918b56f4f02c |
| Date | 2026-10-18 |
| Agent | worker, pi, opencode-go/deepseek-v4-flash:medium |
| Branch | `task/T01-static-home` |

## Scope

This pass extends the T01 home correction. `home.Resolve` now checks `.hive`
with `Lstat` as a real directory before it accepts the marker. A linked `.hive`
or a wrong-type `.hive` is not a marker. The resolver still resolves a symbolic
home root, searches for the outermost home, applies the flag and environment
precedence, reports a conflict, and refuses a non-regular marker.

The pass changes these files:

| Path | Change |
| --- | --- |
| `internal/lib/home/home.go` | `hasMarker` checks the `.hive` type with `Lstat` before the marker check. |
| `internal/lib/home/home_test.go` | Adds a linked `.hive`, a wrong-type `.hive`, and an outer-home linked-nested case. |
| `cmd/project_static_home_test.go` | New CLI regression for a linked `.hive`, the external tree, the flag, the environment, cwd discovery, a regular control, and a symbolic home root. |

The pass changes no `cmd/common`, `internal/lib/log`, or `internal/lib/store`
code. The resolver makes no home or state directory. A bad state root stays
visible to `doctor`. The approved local-user and cooperating-writer boundary
applies. A hostile replacement between `Lstat` and `Open` stays deferred.

## Red run

The tree was at base commit `59bfb8fe`. The new tests were present and
`internal/lib/home/home.go` still had the old marker check.

```sh
commit 59bfb8fe
go test ./internal/lib/home/ -run 'TestResolve(RejectsLinkedHiveDirectory|RejectsNonDirectoryHive|AncestorSkipsLinkedHiveMarker)' -v -count=1; echo "exit $?"
```

```text
=== RUN   TestResolveRejectsLinkedHiveDirectory
    home_test.go:229: Resolve() = "/tmp/TestResolveRejectsLinkedHiveDirectory93287544/001/home", <nil>; want no home because .hive is a symlink
--- FAIL: TestResolveRejectsLinkedHiveDirectory (0.00s)
=== RUN   TestResolveRejectsNonDirectoryHive
--- PASS: TestResolveRejectsNonDirectoryHive (0.00s)
=== RUN   TestResolveAncestorSkipsLinkedHiveMarker
--- PASS: TestResolveAncestorSkipsLinkedHiveMarker (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/home	0.001s
FAIL
exit 1
```

```sh
commit 59bfb8fe
go test ./cmd/ -run 'TestProjectList(RejectsLinkedHiveHome|AcceptsRegularHiveHome|AcceptsSymlinkedHomeRoot)' -v -count=1; echo "exit $?"
```

```text
=== RUN   TestProjectListRejectsLinkedHiveHome
=== RUN   TestProjectListRejectsLinkedHiveHome/flag_text
    project_static_home_test.go:52: status = 0; want 1 (stdout "no projects\n", stderr "")
=== RUN   TestProjectListRejectsLinkedHiveHome/flag_json
    project_static_home_test.go:52: status = 0; want 1 (stdout "{\"outcome\":\"ok\",\"phase\":\"complete\",\"effects\":[],\"error\":null,\"result\":[]}\n", stderr "")
=== RUN   TestProjectListRejectsLinkedHiveHome/environment_text
    project_static_home_test.go:52: status = 0; want 1 (stdout "no projects\n", stderr "")
=== RUN   TestProjectListRejectsLinkedHiveHome/cwd_discovery_text
    project_static_home_test.go:52: status = 0; want 1 (stdout "no projects\n", stderr "")
--- FAIL: TestProjectListRejectsLinkedHiveHome (0.00s)
    --- FAIL: TestProjectListRejectsLinkedHiveHome/flag_text (0.00s)
    --- FAIL: TestProjectListRejectsLinkedHiveHome/flag_json (0.00s)
    --- FAIL: TestProjectListRejectsLinkedHiveHome/environment_text (0.00s)
    --- FAIL: TestProjectListRejectsLinkedHiveHome/cwd_discovery_text (0.00s)
=== RUN   TestProjectListAcceptsRegularHiveHome
--- PASS: TestProjectListAcceptsRegularHiveHome (0.00s)
=== RUN   TestProjectListAcceptsSymlinkedHomeRoot
--- PASS: TestProjectListAcceptsSymlinkedHomeRoot (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/cmd	0.002s
FAIL
exit 1
```

The red tests show the defect. The old source accepts the linked `.hive` as a
marker. The CLI exits 0 and writes below the external directory. The next block
is an independent CLI observation at base commit `59bfb8fe`.

```sh
commit 59bfb8fe
go build -o .hive/tmp/t01-static-home-fix-01/red-probe2/hive-red .
cd .hive/tmp/t01-static-home-fix-01/red-probe2
./hive-red --home home project list; echo "exit $?"
find home external | sort
```

```text
no projects
exit 0
/home/penguin/source/hive/projects/hive/worktrees/T01-static-home/.hive/tmp/t01-static-home-fix-01/red-probe2/external
/home/penguin/source/hive/projects/hive/worktrees/T01-static-home/.hive/tmp/t01-static-home-fix-01/red-probe2/external/README.md
/home/penguin/source/hive/projects/hive/worktrees/T01-static-home/.hive/tmp/t01-static-home-fix-01/red-probe2/external/logs
/home/penguin/source/hive/projects/hive/worktrees/T01-static-home/.hive/tmp/t01-static-home-fix-01/red-probe2/external/logs/2026-10-18.jsonl
/home/penguin/source/hive/projects/hive/worktrees/T01-static-home/.hive/tmp/t01-static-home-fix-01/red-probe2/external/state
/home/penguin/source/hive/projects/hive/worktrees/T01-static-home/.hive/tmp/t01-static-home-fix-01/red-probe2/home
/home/penguin/source/hive/projects/hive/worktrees/T01-static-home/.hive/tmp/t01-static-home-fix-01/red-probe2/home/.hive
```

The external directory had one regular `README.md` before the command. It held
`logs/2026-10-18.jsonl` and `state/` after the command. The fixture home held
only the `.hive` link.

## Green run

The tree was at source commit `72c7ac48`. The source had the `.hive` type
check.

```sh
commit 72c7ac48
go test -race -count=1 ./internal/lib/home/ -run TestResolve -v; echo "exit $?"
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
=== RUN   TestResolveRejectsLinkedHiveDirectory
--- PASS: TestResolveRejectsLinkedHiveDirectory (0.00s)
=== RUN   TestResolveRejectsNonDirectoryHive
--- PASS: TestResolveRejectsNonDirectoryHive (0.00s)
=== RUN   TestResolveAncestorSkipsLinkedHiveMarker
--- PASS: TestResolveAncestorSkipsLinkedHiveMarker (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.006s
exit 0
```

```sh
commit 72c7ac48
go test -race -count=1 ./cmd/ -run 'TestProjectList|TestRootMissingHome|TestRootProjectListJSONEnvelope|TestResolveEmptyHomeFlagHasUsageStatus' -v; echo "exit $?"
```

```text
=== RUN   TestRootProjectListJSONEnvelope
--- PASS: TestRootProjectListJSONEnvelope (0.00s)
=== RUN   TestRootMissingHomeUsesOneErrorLine
--- PASS: TestRootMissingHomeUsesOneErrorLine (0.00s)
=== RUN   TestRootMissingHomeJSONEnvelope
--- PASS: TestRootMissingHomeJSONEnvelope (0.00s)
=== RUN   TestResolveEmptyHomeFlagHasUsageStatus
--- PASS: TestResolveEmptyHomeFlagHasUsageStatus (0.00s)
=== RUN   TestProjectListRejectsLinkedHiveHome
=== RUN   TestProjectListRejectsLinkedHiveHome/flag_text
=== RUN   TestProjectListRejectsLinkedHiveHome/flag_json
=== RUN   TestProjectListRejectsLinkedHiveHome/environment_text
=== RUN   TestProjectListRejectsLinkedHiveHome/cwd_discovery_text
--- PASS: TestProjectListRejectsLinkedHiveHome (0.00s)
    --- PASS: TestProjectListRejectsLinkedHiveHome/flag_text (0.00s)
    --- PASS: TestProjectListRejectsLinkedHiveHome/flag_json (0.00s)
    --- PASS: TestProjectListRejectsLinkedHiveHome/environment_text (0.00s)
    --- PASS: TestProjectListRejectsLinkedHiveHome/cwd_discovery_text (0.00s)
=== RUN   TestProjectListAcceptsRegularHiveHome
--- PASS: TestProjectListAcceptsRegularHiveHome (0.00s)
=== RUN   TestProjectListAcceptsSymlinkedHomeRoot
--- PASS: TestProjectListAcceptsSymlinkedHomeRoot (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/cmd	1.008s
exit 0
```

The linked-home CLI cases now fail before `log.Open` and before `store.Open`.
The tests compare the external tree byte for byte. They also compare the fixture
home. The red case gained `logs/` and `state/` below the external directory. The
green case gains no path.

## CLI observation

The binary was built at `72c7ac48`. The external tree held one regular
`README.md`. The fixture home held only the `.hive` link.

```sh
commit 72c7ac48
go build -o green-probe/hive-green .
(cd green-probe && env -u HIVE_HOME ./hive-green --home home project list); echo "exit $?"
(cd green-probe && env -u HIVE_HOME ./hive-green --home home --json project list); echo "exit $?"
```

```text
hive: not a hive home: /home/penguin/source/hive/projects/hive/worktrees/T01-static-home/.hive/tmp/t01-static-home-fix-01/green-probe/home
exit 1
{"outcome":"failed","phase":"home","effects":[],"error":"not a hive home: /home/penguin/source/hive/projects/hive/worktrees/T01-static-home/.hive/tmp/t01-static-home-fix-01/green-probe/home","result":null}
exit 1
```

The external snapshot had the same path, mode, and SHA256 before and after.
The `diff` of the snapshots printed no line and exited 0. The fixture home kept
only the `.hive` link.

The cwd discovery check ran in an isolated fixture below `/tmp`. No real home
was above it. The external tree stayed byte for byte the same.

```sh
commit 72c7ac48
(cd /tmp/t01-static-home-fix-01.XVZNcHRd/home && env -u HIVE_HOME ../hive project list); echo "exit $?"
(cd /tmp/t01-static-home-fix-01.XVZNcHRd/home && env -u HIVE_HOME ../hive --json project list); echo "exit $?"
(cd /tmp/t01-static-home-fix-01.XVZNcHRd/neutral && env HIVE_HOME=/tmp/t01-static-home-fix-01.XVZNcHRd/home ../hive project list); echo "exit $?"
(cd /tmp/t01-static-home-fix-01.XVZNcHRd/neutral && env -u HIVE_HOME ../hive --home /tmp/t01-static-home-fix-01.XVZNcHRd/home project list); echo "exit $?"
```

```text
hive: not a hive home: no marker above /tmp/t01-static-home-fix-01.XVZNcHRd/home
exit 1
{"outcome":"failed","phase":"home","effects":[],"error":"not a hive home: no marker above /tmp/t01-static-home-fix-01.XVZNcHRd/home","result":null}
exit 1
hive: not a hive home: /tmp/t01-static-home-fix-01.XVZNcHRd/home
exit 1
hive: not a hive home: /tmp/t01-static-home-fix-01.XVZNcHRd/home
exit 1
```

The external snapshot had the same path, mode, and SHA256 before and after.
The `diff` of the snapshots printed no line and exited 0.

The regular control home wrote `logs/2026-10-18.jsonl` and `state/` below its
own `.hive`. A symbolic link to a real home root resolved to the canonical home.
The alias stayed a symbolic link.

```sh
commit 72c7ac48
(cd /tmp/t01-static-home-fix-01.XVZNcHRd/neutral && env -u HIVE_HOME ../hive --home /tmp/t01-static-home-fix-01.XVZNcHRd/home2 project list); echo "exit $?"
find /tmp/t01-static-home-fix-01.XVZNcHRd/home2/.hive | sort
(cd /tmp/t01-static-home-fix-01.XVZNcHRd/neutral && env -u HIVE_HOME ../hive --home /tmp/t01-static-home-fix-01.XVZNcHRd/alias3 project list); echo "exit $?"
find /tmp/t01-static-home-fix-01.XVZNcHRd/home3/.hive | sort
stat -c '%F %n' /tmp/t01-static-home-fix-01.XVZNcHRd/alias3
```

```text
no projects
exit 0
/tmp/t01-static-home-fix-01.XVZNcHRd/home2/.hive
/tmp/t01-static-home-fix-01.XVZNcHRd/home2/.hive/README.md
/tmp/t01-static-home-fix-01.XVZNcHRd/home2/.hive/logs
/tmp/t01-static-home-fix-01.XVZNcHRd/home2/.hive/logs/2026-10-18.jsonl
/tmp/t01-static-home-fix-01.XVZNcHRd/home2/.hive/state
no projects
exit 0
/tmp/t01-static-home-fix-01.XVZNcHRd/home3/.hive
/tmp/t01-static-home-fix-01.XVZNcHRd/home3/.hive/README.md
/tmp/t01-static-home-fix-01.XVZNcHRd/home3/.hive/logs
/tmp/t01-static-home-fix-01.XVZNcHRd/home3/.hive/logs/2026-10-18.jsonl
/tmp/t01-static-home-fix-01.XVZNcHRd/home3/.hive/state
symbolic link /tmp/t01-static-home-fix-01.XVZNcHRd/alias3
```

## Gate

```sh
commit 72c7ac48
gofmt -l $(git ls-files -co --exclude-standard '*.go'); echo "gofmt exit $?"
go vet ./...; echo "vet exit $?"
golangci-lint run ./...; echo "lint exit $?"
go test -race ./...; echo "test exit $?"
```

```text
gofmt exit 0
vet exit 0
0 issues.
lint exit 0
ok  	github.com/Harrison-Blair/hive	(cached)
ok  	github.com/Harrison-Blair/hive/cmd	(cached)
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
test exit 0
```

The first command printed no Go file. The lint command ran alone. An uncached
repeat passed.

```sh
commit 72c7ac48
go test -race -count=1 ./...; echo "test exit $?"
```

```text
ok  	github.com/Harrison-Blair/hive	2.626s
ok  	github.com/Harrison-Blair/hive/cmd	1.020s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/version	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.013s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.045s
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/records	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/store	3.146s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.009s
ok  	github.com/Harrison-Blair/hive/internal/project	1.018s
ok  	github.com/Harrison-Blair/hive/internal/tui	1.011s
test exit 0
```

## Cleanup

The pass removed only its own fixtures and ignored binary.

```sh
rm -rf .hive/tmp/t01-static-home-fix-01 /tmp/t01-static-home-fix-01.XVZNcHRd
test ! -e .hive/tmp/t01-static-home-fix-01 && test ! -e /tmp/t01-static-home-fix-01.XVZNcHRd && echo 'pass fixtures absent'
git status --short
```

```text
pass fixtures absent
```

## Result

`home.Resolve` rejects a linked `.hive` and a wrong-type `.hive` before the
marker check. A named home with an invalid linked marker fails with no write.
Ancestor search does not accept an invalid linked marker as a home. The CLI
keeps the flag, environment, cwd, conflict, and symbolic home-root behavior.
The external tree stays byte for byte the same. The four gate commands pass.
Independent source verification stays required after this pass.
