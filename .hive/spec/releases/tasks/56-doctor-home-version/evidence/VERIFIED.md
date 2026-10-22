# Task 56 verification

- Task: `56-doctor-home-version`
- Verifier: `t56-verifier-1`
- Date: 2026-10-22
- Task commit: `69c8229fd49ec92e143f4ca611ac930de016ed4c`
- Verdict: pass

| Item | Result | Observation |
| --- | --- | --- |
| V1 Unit tests | pass | Base fails to compile the new tests. The task passes all eight tests. A broken comparison fails the two matching-tag tests. The restored code passes. |
| V2 Real doctor run | pass | Text and JSON show `ok` at `v0.1.0` and optional `warn` at untagged commit `499eec5f`. All four commands exit 0. The warning gives the binary version, home commit, README Update reference, and update steps. |
| V-gate Gate | pass | The merge reports `Already up to date.` All five gate commands exit 0. |

## V1

Ran from the worktree root:

```sh
go test ./internal/doctor/ -run HomeVersion -v
```

Used Go file overlays for the base and mutation runs. This kept the tracked
source files unchanged. The base overlay used `checks.go` and `run.go` from
`499eec5f`, with the new tests still present. The command exited 1 with eight
`undefined: checkHomeVersion` errors.

The task run exited 0. These tests passed:

- `TestHomeVersionDevBuildPasses`
- `TestHomeVersionSameTagPasses`
- `TestHomeVersionDifferentTagWarns`
- `TestHomeVersionNoTagWarnsWithCommit`
- `TestHomeVersionTwoTagsOneMatchPasses`
- `TestHomeVersionTwoTagsNoMatchWarns`
- `TestHomeVersionGitFailureWarns`
- `TestHomeVersionInDoctorRun`

The mutation overlay changed `if tag == binary` to `if tag == binary+"x"`.
The same test command exited 1. `TestHomeVersionSameTagPasses` and
`TestHomeVersionTwoTagsOneMatchPasses` received `warn` instead of `ok`.
Without the overlay, all eight tests passed again with exit 0.

The tests use temporary Git repositories for tags and commits. The Git
failure test checks both a non-repository and a missing Git executable.

## V2

Built the task binary:

```sh
go build -ldflags "-X github.com/Harrison-Blair/hive/internal/lib/version.Version=v0.1.0" -o SCRATCH/hive .
```

The binary reported `hive v0.1.0 (commit 69c8229fd49e, go1.27.2)`.
Cloned the repository at `v0.1.0` into `/tmp/hive-verify56-yDC3mS/home`,
outside the real home. Set `XDG_CONFIG_HOME` to a scratch directory with
`hive/config.toml` pointing at that clone. Unset `HIVE_HOME`.

Ran these commands in the scratch home at each revision:

```sh
SCRATCH/hive doctor
SCRATCH/hive --json doctor
```

At tag `v0.1.0`, both commands showed:

```text
ok home version: v0.1.0
```

After `git checkout 499eec5f`, `git tag --points-at HEAD` printed nothing.
Both commands showed this warning:

```text
warn home version: binary v0.1.0, no tag at commit 499eec5f; update (README.md, Update): git -C HOME_DIR fetch --tags; git -C HOME_DIR checkout v0.1.0; go install github.com/Harrison-Blair/hive@v0.1.0; restart the supervisor and the coordinator
```

Compared the warning with root `README.md`, section `## Update`. It names
the reference and gives the fetch, checkout, install, and restart steps.
The JSON check has `required: false`; both JSON envelopes have
`outcome: "ok"`. All required checks pass. All four commands exit 0.
The scratch home also has optional warnings for the absent supervisor and
GH authentication. A separate scratch program constructed a doctor result
with only the optional `home version` warning. It reported
`Failed=false ExitCode=0` and exited 0.

The real home HEAD and Git status matched the snapshots from before the
run. The real `~/.config/hive/config.toml` was absent before and after the
run. No tag was pushed. No release was changed. Removed the scratch clone,
binary, and config. Confirmed that the scratch directory no longer exists.

## V-gate

Ran `git merge dev --no-edit` once before the gate. It reported
`Already up to date.` There was no merge conflict.

| Command | Exit | Output |
| --- | --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | 0 | None |
| `go mod tidy -diff` | 0 | None |
| `go vet ./...` | 0 | None |
| `golangci-lint run ./...` | 0 | `0 issues.` |
| `go test -race ./...` | 0 | All packages pass; no failed tests |

No retry was needed. `git diff --check` also exited 0. Removed the local
scratch overlays, logs, scripts, and program after recording the results.
