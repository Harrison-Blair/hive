# Task 50 independent verification

- Date: 2026-10-22
- Verifier: `t50-verifier-1` (pi)
- Task commit: `7366e6e9241a5dd18471548a89cd217ba75e7225`
- Verdict: pass

| Item | Result | Observation |
| --- | --- | --- |
| V1 | pass | New tests fail with the base source and pass with the task source. |
| V2 | pass | Text and JSON report `dev`, `v9.9.9`, and `v0.9.8`. The manager approved the module-copy limits below. |
| V-gate | pass | One merge attempt and all five gate commands completed with exit 0. |

## V1

From the worktree root, saved the task version source in `.hive/tmp/t50-verifier/`. Replaced `internal/lib/version/version.go` with the source from `19988d4c`. Kept the new tests. Ran:

```sh
go test ./internal/lib/version/ -run . -v
```

Exit 1. The release and release-candidate tests failed for the expected reason:

```text
info().Version = "dev", want "v1.2.3"
info().Version = "dev", want "v1.2.3-rc.4"
```

Restored the task source and ran the same command. Exit 0. All eight cases passed: ldflags, release, release candidate, pseudo-version, dirty, devel, empty, and other suffix. `git diff --exit-code -- internal/lib/version/version.go` returned 0 after restoration.

## V2

Built the plain and ldflags binaries with these commands:

```sh
go build -o .hive/tmp/t50-verifier/bin/plain/hive .
go build -o .hive/tmp/t50-verifier/bin/ldflags/hive -ldflags '-X github.com/Harrison-Blair/hive/internal/lib/version.Version=v9.9.9' .
```

For the install, made a new module copy with `git archive HEAD`. Used the file proxy method in `V2.md`, with scratch root `.hive/tmp/t50-verifier` and version `v0.9.8`. The copy omitted only these three folders:

- `.hive/spec/tui-v2/tasks/34-screen-chat/evidence/screens`
- `.hive/spec/tui-v2/tasks/35-screen-work-goals/evidence/verifier3`
- `.hive/spec/tui-v2/tasks/35-screen-work-goals/evidence/verifier4`

Created the proxy `.zip`, `.mod`, `.info`, and `list` files. From `/`, ran `go install github.com/Harrison-Blair/hive@v0.9.8` with a scratch `GOMODCACHE`, scratch `GOBIN`, `GOFLAGS=-modcacherw`, `GONOSUMDB=github.com/Harrison-Blair/hive`, and `GOPROXY=file://SCRATCH/proxy,https://proxy.golang.org`. Exit 0.

The manager approved `v0.9.8` because this module path cannot use major version 9. The manager also approved the three folder omissions because the tracked case clashes block a module zip until task 53. This check does not prove that the unchanged repository can be installed as a module.

Ran `version` and `--json version` with each binary. All six commands returned exit 0:

| Build | Text | JSON result.version |
| --- | --- | --- |
| plain | `hive dev (commit 7366e6e9241a, go1.27.2)` | `dev` |
| ldflags | `hive v9.9.9 (commit 7366e6e9241a, go1.27.2)` | `v9.9.9` |
| install | `hive v0.9.8 (commit unknown, go1.27.3)` | `v0.9.8` |

`go version -m` showed module version `v0.0.3-0.20261022045040-7366e6e9241a` in the checkout builds and `v0.9.8` in the installed binary. No tag was pushed. No GitHub release command ran.

## V-gate

At the start of this item, ran `git merge dev` once. Output: `Already up to date.` Exit 0. There was no conflict. During the gate, `dev` advanced to `77deb3ec` with task 52. Per the manager instruction, did not merge again. Gate results apply to task commit `7366e6e9`.

| Command | Exit | Output |
| --- | --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | 0 | empty |
| `go mod tidy -diff` | 0 | empty |
| `go vet ./...` | 0 | empty |
| `golangci-lint run ./...` | 0 | `0 issues.` |
| `go test -race ./...` | 0 | All packages passed or had no tests. No race report. |

The race run included `internal/lib/version`, `internal/task`, `internal/tui`, and `internal/worktree`. No retry was needed.

## Cleanup

Removed the scratch module copy, proxy, module cache, binaries, saved source, and logs from `.hive/tmp/t50-verifier` after this record was written. No source changes remain from the red run.
