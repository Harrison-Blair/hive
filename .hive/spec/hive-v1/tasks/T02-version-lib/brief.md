# T02 version lib

| Item | Value |
| --- | --- |
| Task ID | T02-version-lib |
| Wave | 1 |
| Tier | small |
| Depends on | none |
| Blocks | CP1 |
| Branch | `task/T02-version-lib` |
| Worktree | `projects/hive/worktrees/T02-version-lib/` |
| Evidence | `evidence/T02-version-lib/` |

## Overview

The scaffold keeps the version in `internal/version`. The layout rule puts
shared code under `internal/lib/`. This task moves the package, adds build data,
and keeps `hive version` working. It is small and independent, so it runs beside
T01 and tests the two-worktree procedure of the tracker.

## User flows

1. The user runs `hive version` and sees `hive dev (commit 1a2b3c4d5e6f, go1.27.2)`.
2. A build with `-ldflags "-X github.com/Harrison-Blair/hive/internal/lib/version.Version=1.0.0"` prints `1.0.0` as the version.
3. A build with `-buildvcs=false` prints `commit unknown`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/lib/version/version.go` | New. The `Version` variable and `Info()` with version, commit, modified flag, and Go version from `debug.ReadBuildInfo`. |
| `internal/lib/version/version_test.go` | New. |
| `internal/version/version.go` | Delete. |
| `cmd/version/version.go` | Import the new path. Print the full line. |
| `cmd/root_test.go` | Change the assertion of `TestVersionSubcommandPrintsVersion` from `dev\n` to the line format of AC2. |

## Requirements

### AC1 Package move

- The package `internal/lib/version` holds `var Version = "dev"` and `func Info() BuildInfo`. No file remains in `internal/version/`, and no file imports the old path.
- `BuildInfo` has `Version`, `Commit`, `Modified`, and `Go`. `Commit` is the first 12 characters of the `vcs.revision` setting, or `unknown`. `Modified` is true when `vcs.modified` is `true`. `Go` is the Go version of the build.

### AC2 Command output

- `hive version` prints one line: `hive <version> (commit <commit>, <go>)`. When `Modified` is true, the line has `, modified` after the commit.
- The exit status is 0.

### AC3 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The package lives under `internal/lib/`, and `Info()` reads the build data. | V1, V2 |
| AC2 | `hive version` prints the one line and exits with 0. | V2 |
| AC3 | The gate passes. | V3 |

## Verification items

### V1 Version tests

- Proves: AC1.
- Command: `go test ./internal/lib/version/ -v`, red and green.
- Expected: a test gives `Info()` build settings through an injectable reader and checks the 12-character commit, the `unknown` fallback, and the modified flag.

### V2 Command transcript

- Proves: AC1, AC2.
- Command:

```sh
go build . && ./hive version; echo "exit $?"
go build -ldflags "-X github.com/Harrison-Blair/hive/internal/lib/version.Version=1.0.0" -o .hive/tmp/hive-ldflags . && .hive/tmp/hive-ldflags version
go build -buildvcs=false -o .hive/tmp/hive-novcs . && .hive/tmp/hive-novcs version
rm -f .hive/tmp/hive-ldflags .hive/tmp/hive-novcs
ls internal/version; echo "exit $?"
```

- Expected: line one prints one line in the format of AC2 and exit 0. Line two prints `hive 1.0.0 (commit <commit>, <go>)`. Line three prints `commit unknown` in its line. `ls` fails with `No such file or directory` and exit 2.

### V3 Gate

- Proves: AC3.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- `--json` for `hive version`. The envelope comes with T01, and T05 adds the flag.

## Changes

None.
