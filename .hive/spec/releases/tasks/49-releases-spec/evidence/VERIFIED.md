# Independent check: 49-releases-spec

- Date: 2026-10-22
- Agent: `t49-verifier-1` (Pi)
- Branch: `task/49-releases-spec`
- Checked commit: `62603efb`
- Verdict: pass

## Verification items

| Item | Result | Observation |
| --- | --- | --- |
| V1 Spec content | pass | Read the README next to the corrected plan. Each approved R1 to R9 option A and each AC1 detail is present. Problem agrees with the repository and read-only GitHub replies. |
| V3 Ratchet | pass | `go test ./ -run STE` passed, failed with a temporary semicolon in the README, and passed after removal. Both new baseline entries are 0. |
| V4 Diff scope | pass | After `git merge --no-edit dev`, `git diff dev...HEAD --stat` showed only the four source-table paths and task evidence. |
| V5 Gate | pass | Each of the five gate commands exited 0. Formatting and module tidiness printed nothing. Lint printed `0 issues.`. The race suite passed. |

## V1: plan and facts

The plan is `/home/penguin/source/hive/.hive/tmp/reports/releases-plan.md`.
The manager supplied the user approval of R1 to R9 option A.

| Plan decision | Spec decisions | Check |
| --- | --- | --- |
| R1 | 1 to 4 | Annotated tag, stable and rc names, prerelease flag, and main ancestry check. |
| R2 | 5 | First version `v0.1.0`. |
| R3 and R4 | 6 to 11 | Linux amd64 and arm64, exact archive names, binary and LICENSE, checksums, build flags, version symbol, and reproducible archives. |
| R5 | 16 and 17 | Tagged home, matching archive or go install, PATH, init, and module-version fallback. |
| R6 | 18 and 19 | Update commands, binary replacement, process restart, and later doctor check. |
| R7 | 20 and 21 | Keep releases and tags, Fledge titles, and predecessor top line. |
| R8 | 22 | GitHub generated release text and install header. |
| R9 | 23 | Root README with install, update, and spec link. |

Decisions 12 to 15 specify the callable CI gate before the build, archive checks on matching runners, and draft/upload/publish order.
Decisions 24 and 25 require separate grants for the Fledge edits and the first tag.
Tasks gives waves 1 to 5, including the PR #47 dependency.
Risks includes the proxy, old module path, home/binary mismatch, and off-main tag.
No decision adds a feature outside the plan and brief.

Independent fact checks:

- Read `.github/workflows/ci.yml`. It is the only current workflow. It runs the gate on pushes and pull requests to main and dev.
- Ran `gh api repos/Harrison-Blair/hive/releases --paginate` with a read-only projection of tag, title, files, and body. All six release bodies link to Fledge pull requests.
- The releases v0.0.4 to v0.0.6 each have `checksums.txt` and two `fledge_v0.0.X_linux_ARCH.tar.gz` archives.
- The releases v0.0.1 and v0.0.2 each have `SHA256SUMS` and two `fledge_0.0.X_linux_ARCH.tar.gz` archives. Release v0.0.3 has no files.
- Ran `git show TAG:go.mod` for each of the six tags. Each gives `github.com/Harrison-Blair/fledge`. The current go.mod gives `github.com/Harrison-Blair/hive`. There is no valid hive module version among these tags.
- `git rev-parse v0.0.6^{} origin/main` gave `928cfa95e0906315fcb4a65a5cdbdef1470a0503` for both. The read-only GitHub main-branch API gave the same SHA.
- The read-only GitHub API for PR #47 gave open, not merged, with dev as head and main as base.
- Inspected commits `e45979f4` and `3dd44730`, plus the historical package script, workflow, release script, and update code. They show the stated addition/removal, manual bump choice, gate dependencies, reproducible package settings, draft/upload/publish sequence, and SHA-256 check before binary replacement.
- Read `internal/lib/version/version.go`. Version defaults to dev. Build information supplies VCS revision and modified status. The code does not read the module version.
- Read `internal/homeinit/homeinit.go` and `internal/lib/home/home.go`. Init requires a marker above its working directory. Agent instructions and skill/brief paths use the checkout. The current root has no README.md.
- Checked `internal/doctor/checks.go`: the supervisor binary check reports an older executable.

These checks support the stated source facts. No install or future release workflow was run: this task only writes the spec.

## V3: negative control

Initial run:

```text
$ go test ./ -run STE
ok github.com/Harrison-Blair/hive 0.842s
```

Added one temporary line after the README title:

```text
The verifier checks the file; the ratchet must fail.
```

The same command exited 1:

```text
--- FAIL: TestSTERatchet
.hive/spec/releases/README.md: 1 violations, baseline 0; fix the new ones
.hive/spec/releases/README.md:3 semicolon: ;
FAIL
```

Removed that exact line and ran the same command again:

```text
ok github.com/Harrison-Blair/hive 0.870s
```

`git diff --exit-code` then exited 0. The temporary mutation did not remain.

## V4: merge and scope

At the start of the gate phase:

```text
$ git merge --no-edit dev
Already up to date.
```

No conflict occurred. `git diff dev...HEAD --stat` gave:

```text
.hive/spec/releases/README.md                      | 121 +
.hive/spec/releases/tasks/49-releases-spec/brief.md | 131 +
.../evidence/V1.md                                 | 154 +
.../evidence/V3.md                                 |  69 +
.../evidence/V4.md                                 |  36 +
.../evidence/V5.md                                 | 120 +
ste_test.go                                       |   1 +
testdata/ste-baseline.json                         |   2 +
8 files changed, 634 insertions(+)
```

The verifier adds only this evidence file.

## V5: gate output

| Exact command | Exit | Output |
| --- | --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | 0 | Empty. |
| `go mod tidy -diff` | 0 | Empty. |
| `go vet ./...` | 0 | Empty. |
| `golangci-lint run ./...` | 0 | `0 issues.` |
| `go test -race ./...` | 0 | Each tested package passed. Packages without tests reported that status. |

Selected race-suite output:

```text
ok github.com/Harrison-Blair/hive 17.711s
ok github.com/Harrison-Blair/hive/cmd 6.062s
ok github.com/Harrison-Blair/hive/internal/agent 12.877s
ok github.com/Harrison-Blair/hive/internal/task 9.487s
ok github.com/Harrison-Blair/hive/internal/tui 37.094s
ok github.com/Harrison-Blair/hive/internal/worktree 4.677s
```

An initial output redirection failed because `.hive/tmp/` did not exist. That attempt did not start the race command. After creation of the scratch folder, the gate command ran and passed. No test retry was necessary.

## Boundaries and cleanup

No UI changed, so no visual check applies. No tag was pushed. No GitHub release was created or edited. No `gh release` command ran. The temporary README violation was removed. The scratch race log was removed after this record captured the result.
