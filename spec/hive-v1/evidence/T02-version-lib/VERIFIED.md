# T02-version-lib VERIFIED

| Item | Value |
| --- | --- |
| Task | T02-version-lib |
| Commit | 64540c42 |
| Date | 2026-10-17 |
| Agent | verifier, pi, opencode-go/deepseek-v4.1-flash |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | pass | The Red run in V1.md exists and fails on the version_test.go:46 assertion, not on a compile error. The Green run shows `(cached)`; I also ran it with `-count=1`. |
| V2 | AC1, AC2 | pass | My run matches the expected output. The evidence has Red and Green sections although it is not a test item, the Red command runs `gofmt -w`, the Run block ends with `rm -f hive`, the Cleanup block ends with `find`, and the Green run shows `(cached)`. |
| V3 | AC3 | pass | |

## Runs

### V1

Source lines that the item asks to confirm:

```text
internal/lib/version/version.go:21-25
func Info() BuildInfo {
	return info(debug.ReadBuildInfo)
}

func info(readBuildInfo func() (*debug.BuildInfo, bool)) BuildInfo {

internal/lib/version/version.go:42,45
	result.Commit = result.Commit[:12]
	result.Modified = setting.Value == "true"

internal/lib/version/version_test.go:18,24
	{Key: "vcs.revision", Value: "123456789abcdef0123456789abcdef012345678"}  (40 characters)
	Commit:   "123456789abc"

internal/lib/version/version_test.go:34
	Commit:  "unknown"

internal/lib/version/version_test.go:19,25
	{Key: "vcs.modified", Value: "true"}
	Modified: true

internal/lib/version/version_test.go:42
	got := info(func() (*debug.BuildInfo, bool) {
```

Command and output of my green run:

```sh
go test ./internal/lib/version/ -v; echo "exit $?"
go test ./internal/lib/version/ -v -count=1; echo "exit $?"
```

```text
=== RUN   TestInfoReadsBuildSettings
=== RUN   TestInfoReadsBuildSettings/revision_and_modified
=== RUN   TestInfoReadsBuildSettings/missing_revision
--- PASS: TestInfoReadsBuildSettings (0.00s)
    --- PASS: TestInfoReadsBuildSettings/revision_and_modified (0.00s)
    --- PASS: TestInfoReadsBuildSettings/missing_revision (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
exit 0
=== RUN   TestInfoReadsBuildSettings
=== RUN   TestInfoReadsBuildSettings/revision_and_modified
=== RUN   TestInfoReadsBuildSettings/missing_revision
--- PASS: TestInfoReadsBuildSettings (0.00s)
    --- PASS: TestInfoReadsBuildSettings/revision_and_modified (0.00s)
    --- PASS: TestInfoReadsBuildSettings/missing_revision (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/version	0.001s
exit 0
```

The Red run recorded in V1.md (not rerun), at commit 45b1bf58:

```text
=== RUN   TestInfoReadsBuildSettings/revision_and_modified
    version_test.go:46: info() = {Version: Commit: Modified:false Go:}, want {Version:dev Commit:123456789abc Modified:true Go:go1.27.2-custom}
=== RUN   TestInfoReadsBuildSettings/missing_revision
    version_test.go:46: info() = {Version: Commit: Modified:false Go:}, want {Version:dev Commit:unknown Modified:false Go:go1.27.2-custom}
--- FAIL: TestInfoReadsBuildSettings (0.00s)
    --- FAIL: TestInfoReadsBuildSettings/revision_and_modified (0.00s)
    --- FAIL: TestInfoReadsBuildSettings/missing_revision (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/version	0.001s
FAIL
exit 1
```

### V2

```sh
go build . && ./hive version; echo "exit $?"
go build -ldflags "-X github.com/Harrison-Blair/hive/internal/lib/version.Version=1.0.0" -o .hive/tmp/hive-ldflags . && .hive/tmp/hive-ldflags version; echo "exit $?"
go build -buildvcs=false -o .hive/tmp/hive-novcs . && .hive/tmp/hive-novcs version; echo "exit $?"
rm -f .hive/tmp/hive-ldflags .hive/tmp/hive-novcs; echo "exit $?"
ls internal/version; echo "exit $?"
```

```text
hive dev (commit 64540c42cbe7, go1.27.2-X:nodwarf5)
exit 0
hive 1.0.0 (commit 64540c42cbe7, go1.27.2-X:nodwarf5)
exit 0
hive dev (commit unknown, go1.27.2-X:nodwarf5)
exit 0
exit 0
ls: cannot access 'internal/version': No such file or directory
exit 2
```

The `exit 0` after the third version line is the appended status of the `rm` line. The four substantive outputs match the expected output of the brief.

### V3

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go'); echo "exit $?"
go vet ./...; echo "exit $?"
golangci-lint run ./...; echo "exit $?"
go test -race ./...; echo "exit $?"
```

```text
exit 0
exit 0
0 issues.
exit 0
ok  	github.com/Harrison-Blair/hive	(cached)
ok  	github.com/Harrison-Blair/hive/cmd	(cached)
?   	github.com/Harrison-Blair/hive/cmd/version	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/tui	(cached)
exit 0
```
