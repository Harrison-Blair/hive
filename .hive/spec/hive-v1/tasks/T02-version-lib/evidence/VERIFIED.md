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

## Current Sol audit 2026-10-18

| Item | Value |
| --- | --- |
| Audit report | `.hive/tmp/coordinator/reports/t02-sol-audit-01.md` |
| Tested HEAD | `840255d3f598ad18239eccd7a2eb2e93547ca1fb` |
| Branch | `dev` |
| Route | `pi openai-codex/gpt-6.1-sol:high` |
| V1 | PASS |
| V2 | PASS |
| V3 | PASS |

This section records a current independent audit. It does not replace the
historical verdict or the historical evidence above. The original landing date
stays `2026-10-17`. The original verifier attribution stays in the first table.

Source provenance:

| Commit | Purpose |
| --- | --- |
| `a462848470419f1eb27b5c266324aef3a9b06d3c` | Add the shared package and injectable-reader tests. |
| `1b1501cd0e2c0f55e54c1e6c8a4074b2f4bba139` | Change command output and the root assertion. Remove the old package. |
| `64540c42cbe7eefb1136cd9fd23b0e4628c40d56` | Add worker evidence. |
| `388072ee4a1dee7847b300687146f391affef75f` | Add the historical independent verdict. |
| `d8d6fc8b77d509b79816181a82f56b64f609a7cd` | Land T02. |

The audit rebuilt three binaries in a unique ignored directory. The normal run
printed `hive dev (commit 840255d3f598, go1.27.2-X:nodwarf5)`. The linker
override printed `hive 1.0.0 (commit 840255d3f598, go1.27.2-X:nodwarf5)`. The
`-buildvcs=false` run printed `hive dev (commit unknown,
go1.27.2-X:nodwarf5)`. The audit also passed the four gate commands.

Limits:

- The audit did not repeat the historical red run.
- The audit did not build a new modified tree. The injected-reader test covers
`vcs.modified` true.
- The route is trusted dispatch metadata. The audit did not inspect the model
backend.
- The audit did not test reader failure, a short revision, an empty revision
setting, or explicit `vcs.modified=false`.
- The approved boundary is a local user and a cooperating writer.
- Hostile file replacement between `Lstat` and `Open` stays deferred for T01.

The audit scratch directory stays as necessary evidence. It holds
`results.json`, `provenance.json`, `format-checks.json`, `final-snapshot.json`,
the capture runner, and the text captures. The report embeds the command
transcript. The JSON files hold the per-command snapshots and status records that
the report summarizes. Remove no file from that directory in this pass.
