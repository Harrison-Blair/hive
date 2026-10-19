# 11-ci-workflow VERIFIED

| Item | Value |
| --- | --- |
| Task | 11-ci-workflow |
| Commit | d8fd9bf2 |
| Date | 2026-10-19 |
| Agent | t11-verifier-1, claude, claude-opus-5-5 |
| Verdict | fail |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | R4 | pass | No output. Exit status 1. |
| V2 | R4 | pass | Only `.github/workflows/ci.yml`. |
| V3 | R4 | fail | At the task tip, the new case passes without the rule. `git check-ignore` without `--no-index` does not report a tracked file. |
| V4 | R5 | pass | Exit status 0. No diff. |
| V5 | R2, R3 | pass | `actionlint` is not installed. The YAML parse ran with no error. |
| V6 | R1, R2, R3 | pass | Each trigger, each step, the Go version source, and each pin agree. |
| V7 | gate | pass | `gofmt -l` prints nothing. vet, lint, and race tests exit 0. |

## V1

```text
$ git check-ignore -v .github/workflows/ci.yml; echo "exit=$?"
exit=1
```

## V2

```text
$ git ls-files .github
.github/workflows/ci.yml
```

## V3 new gitignore_test.go case

Control: remove the line `!.github/workflows/ci.yml` from `.gitignore` at d8fd9bf2. The file stays tracked.

Red run, expected fail:

```text
$ go test -count=1 -run 'Ignore|Tracked|GitHub' -v .
=== RUN   TestTrackedFilesAreNotIgnored
    gitignore_test.go:30: .gitignore ignores tracked files; add the narrowest allow rule for each:
        .github/workflows/ci.yml
--- FAIL: TestTrackedFilesAreNotIgnored (0.00s)
=== RUN   TestGitHubWorkflowIsNotIgnored
--- PASS: TestGitHubWorkflowIsNotIgnored (0.00s)
...
FAIL
exit=1
```

The new case `TestGitHubWorkflowIsNotIgnored` passes without the rule. The
existing `TestTrackedFilesAreNotIgnored` fails.

Cause, with the rule removed:

```text
$ git check-ignore -v .github/workflows/ci.yml; echo "plain exit=$?"
plain exit=1
$ git check-ignore -v --no-index .github/workflows/ci.yml; echo "no-index exit=$?"
.gitignore:2:*	.github/workflows/ci.yml
no-index exit=0
```

The worker red run in `V3.md` ran at dadacf88, before Git tracked the file.
After the commit, the new case cannot fail. A possible repair: add
`--no-index` to the `git check-ignore` call in the new case.

Green run, with the rule:

```text
--- PASS: TestGitHubWorkflowIsNotIgnored (0.00s)
ok  	github.com/Harrison-Blair/hive	0.003s
exit=0
```

## V4

```text
$ go mod tidy && git diff --exit-code -- go.mod go.sum; echo "exit=$?"
exit=0
```

## V5

`actionlint` is not installed.

```text
$ python3 -c 'import yaml,sys; yaml.safe_load(open(sys.argv[1]))' .github/workflows/ci.yml; echo "exit=$?"
exit=0
```

## V6

- R1: `pull_request` and `push`, each with branches `main` and `dev`.
- R2: `runs-on: ubuntu-latest`. `actions/setup-go` with `go-version-file: go.mod`.
- R2 pins: `actions/checkout@v7`, `actions/setup-go@v7`, `golangci/golangci-lint-action@v9`. The latest releases on GitHub are `v7.0.1`, `v7.0.0`, and `v9.3.0`. No SHA pin.
- R3: gofmt with a fail on output, tidy check, `go build .`, `go vet ./...`, golangci-lint with `version: v2.15.0` (release exists), `go test -race ./...`. The order agrees.

## V7

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
$ go vet ./...            # exit 0
$ golangci-lint run ./... # 0 issues. exit 0
$ go test -race ./...     # each package ok. exit 0
```

## Remarks

- `testdata/ste-baseline.json` has a new entry for the brief copy at 0. The
  pattern `spec/hive-v1/tasks/*.md` in `steFiles` covers the new file, so R6
  requires this entry.
