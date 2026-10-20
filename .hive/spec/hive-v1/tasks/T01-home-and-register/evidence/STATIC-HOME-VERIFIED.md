# T01 static home: independent verification

| Item | Value |
| --- | --- |
| Verdict | PASS |
| Pass | `t01-home-and-wave2-verify-01`, check A |
| Date | 2026-10-18 |
| Role | Fresh independent verifier |
| Harness and model | `pi openai-codex/gpt-6.1-sol:high` |
| Branch | `task/T01-static-home` |
| Base | `59bfb8fe100d17aba83c3a657e3c918b56f4f02c` |
| Source commit | `72c7ac486d2c517070f9315ed7e7c26e838637f3` |
| Tested worker HEAD | `e4cc4a811e00d690822eb28dfc959c6e4f738fb0` |
| Proves | T01 AC1, AC6, AC7 for this follow-up |

## Scope and source

The initial worktree was clean at the exact worker HEAD.
The remote task branch matched that HEAD.
I read the task, current evidence, worker report, and original escape report.
The worker report ends with `END T01 STATIC HOME FIX 01`.
I checked the complete base-to-worker diff, including the new evidence file.

The diff changes exactly four paths:

- `internal/lib/home/home.go`
- `internal/lib/home/home_test.go`
- `cmd/project_static_home_test.go`
- `spec/hive-v1/evidence/T01-home-and-register/STATIC-HOME-FOLLOWUP.md`

Production adds only four lines in `hasMarker`.
`Lstat` must find a real `.hive` directory before the resolver accepts its marker.
The existing marker check still requires a regular file.
The resolver performs no write and does not inspect the state root.
`cmd/common` resolves the home before it opens the logger.
No logger, store, common implementation, dependency, or STE baseline changes.
Historical T01 evidence objects equal the base objects.

## Tests and assertion checks

I read all 18 current resolver tests.
The file preserves 15 earlier resolver tests and adds three tests.
All current resolver tests and all existing and new command tests pass independently.

```sh
commit e4cc4a81
go test -race -count=1 ./internal/lib/home ./cmd -run Test -v
```

```text
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.005s
ok  	github.com/Harrison-Blair/hive/cmd	1.010s
focused exit 0
```

The tests preserve these behaviors:

- The flag wins and skips the environment conflict check.
- The environment uses canonical paths and reports the existing conflict message.
- Discovery selects the outermost regular marker.
- Empty values, relative paths, and spaces retain their earlier behavior.
- A linked marker remains invalid.
- A symbolic home root resolves to its absolute canonical path.

I made an isolated source copy from the exact worker HEAD.
I replaced only its `home.go` with the file from base `59bfb8fe`.
I did not change actual source.

```sh
# Run in the isolated copy, first with base home.go, then with current home.go.
go test -race -count=1 ./internal/lib/home ./cmd \
  -run 'TestResolve(RejectsLinkedHiveDirectory|RejectsNonDirectoryHive|AncestorSkipsLinkedHiveMarker)|TestProjectList(RejectsLinkedHiveHome|AcceptsRegularHiveHome|AcceptsSymlinkedHomeRoot)' -v
```

The pre-fix run compiled and failed on the intended assertions:

```text
home_test.go:229: Resolve() = "/tmp/TestResolveRejectsLinkedHiveDirectory1197920438/001/home", <nil>; want no home because .hive is a symlink
project_static_home_test.go:52: status = 0; want 1 (stdout "no projects\n", stderr "")
```

All four linked-home command cases failed with status 0 instead of 1.
The pre-fix run exited 1.
The same selection with current `home.go` passed and exited 0.
The wrong-type and outer-home tests also pass on the old source.
Those two tests preserve behavior rather than demonstrate the original defect.

## Actual CLI and resolver checks

I built a new binary from the actual worktree with `go build -o <unique ignored binary> .`.
The fixture directory was `/tmp/t01-home-wave2-v01-tii_wu36`.
No marker existed above it.
Each named home used an absolute `--home` path.
Environment and cwd discovery checks used no home flag.

The linked fixture held `.hive` as a link to an external fixture directory.
The external fixture held a regular marker and a 256-byte sentinel.
I checked text and JSON for all three home selection methods.

| Check | Exit | Assertion |
| --- | --- | --- |
| `--home` text | 1 | Empty stdout and one `not a hive home` diagnostic. |
| `--home` JSON | 1 | One failed envelope with phase `home` and empty stderr. |
| `HIVE_HOME` text and JSON | 1 | The same named-home refusal. |
| cwd discovery text and JSON | 1 | `not a hive home: no marker above <fixture home>`. |
| Regular file at `.hive` | 1 | Named-home refusal and unchanged fixture. |
| Symbolic home-root alias | 0 | `no projects`, canonical log/state paths, and the alias remains a link. |

After each linked-home command, both fixture snapshots remained equal.
The snapshots include paths, modes, file sizes, SHA256 values, and link targets.
No external log or state path appeared.
Thus the CLI rejects the static link before logging or store creation.

An isolated program also called the exported resolver directly.
It accepted a regular home and a home with a regular file at `.hive/state`.
It refused a missing home.
The complete fixture tree remained unchanged.
The resolver creates no home or state and preserves the bad state root for `doctor`.

## Gate

All four commands ran from the actual worktree root.
Lint ran serially.

```sh
commit e4cc4a81
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run --allow-serial-runners ./...
go test -race -count=1 ./...
```

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
gofmt exit 0 (empty)
$ go vet ./...
vet exit 0
$ golangci-lint run --allow-serial-runners ./...
0 issues.
lint exit 0
$ go test -race -count=1 ./...
ok  	github.com/Harrison-Blair/hive	2.637s
ok  	github.com/Harrison-Blair/hive/cmd	1.020s
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
ok  	github.com/Harrison-Blair/hive/internal/lib/store	3.143s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.009s
ok  	github.com/Harrison-Blair/hive/internal/project	1.018s
ok  	github.com/Harrison-Blair/hive/internal/tui	1.012s
test exit 0
```

Packages without tests also exited 0.
The full short output remains in `a-gate.txt` beside the root report.

## Captures, cleanup, and limits

The root report is `.hive/tmp/coordinator/reports/t01-home-and-wave2-verify-01.md`.
Its unique adjacent directory is `t01-home-and-wave2-verify-01-captures/`.
It holds the exact gate, focused, pre-fix, current, CLI, and resolver outputs.
The programs in that directory state the assertions and command arguments.

Cleanup removed only this pass's temporary fixture, isolated source copy, and unique binaries.
All existing runtime, logs, reports, worktrees, credentials, and historical attribution remain.
No agent started and no global setting changed.
Only this new verdict changes a tracked file.
No required check A step remains uncompleted.

The approved local-user and cooperating-writer boundary applies.
Hostile replacement between `Lstat` and `Open` remains deferred.
Logger parent hardening is outside this pass.
This verdict does not claim that the queued fix is already on `dev`.
Recommendation: the coordinator can accept this source for a separate landing.
