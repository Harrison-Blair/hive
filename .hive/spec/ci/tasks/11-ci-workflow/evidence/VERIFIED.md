# 11-ci-workflow VERIFIED

| Item | Value |
| --- | --- |
| Task | 11-ci-workflow |
| Commit | 0bc15eb7 |
| Date | 2026-10-19 |
| Agent | t11-verifier-2, pi, deepseek-v4.1-flash |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | R4 | pass | No output. Exit status 1. |
| V2 | R4 | pass | Only `.github/workflows/ci.yml`. |
| V3 | R4 | pass | The new case fails without the allow rule and passes with it. |
| V4 | R5 | pass | Exit status 0. No diff. |
| V5 | R2, R3 | pass | `actionlint` is not on PATH. The YAML parse ran with no error. |
| V6 | R1, R2, R3 | pass | Each trigger, each step, the Go version source, and each pin agree. |
| V7 | gate | pass | `gofmt -l` prints nothing. vet, lint, and race tests exit 0. |

## V1

```text
$ git check-ignore -v .github/workflows/ci.yml; echo "exit=$?"
exit=1
```

The command printed no line.

## V2

```text
$ git ls-files .github
.github/workflows/ci.yml
```

## V3 new gitignore_test.go case

The worktree was clean at 0bc15eb7 before this item.

Control: remove the line `!.github/workflows/ci.yml` from `.gitignore` in the
worktree. The file stays tracked.

```text
$ git check-ignore --no-index -v .github/workflows/ci.yml; echo "exit=$?"
.gitignore:2:*	.github/workflows/ci.yml
exit=0
```

Red run, expected fail:

```text
$ go test -count=1 -run '^TestGitHubWorkflowIsNotIgnored$' -v .; echo "exit=$?"
=== RUN   TestGitHubWorkflowIsNotIgnored
    gitignore_test.go:38: git ignores .github/workflows/ci.yml
--- FAIL: TestGitHubWorkflowIsNotIgnored (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive	0.003s
exit=1
```

Restore the allow rule with `git checkout -- .gitignore`. `git status
--porcelain` printed nothing.

Green run, expected pass:

```text
$ go test -count=1 -run '^TestGitHubWorkflowIsNotIgnored$' -v .; echo "exit=$?"
=== RUN   TestGitHubWorkflowIsNotIgnored
--- PASS: TestGitHubWorkflowIsNotIgnored (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive	0.003s
exit=0
```

The related cases also pass with the rule:

```text
$ go test -count=1 -run 'TestGitHubWorkflowIsNotIgnored|TestTrackedFilesAreNotIgnored|TestHomeMarkerIsTracked' -v .; echo "exit=$?"
--- PASS: TestTrackedFilesAreNotIgnored (0.00s)
--- PASS: TestGitHubWorkflowIsNotIgnored (0.00s)
--- PASS: TestHomeMarkerIsTracked (0.00s)
PASS
exit=0
```

## V4

```text
$ go mod tidy && git diff --exit-code -- go.mod go.sum; echo "exit=$?"
exit=0
```

The worktree was clean after the command.

## V5

`actionlint` is not on PATH; `command -v actionlint` exited 1. The YAML parse
with python3 ran instead:

```text
$ python3 -c 'import yaml,sys; yaml.safe_load(open(sys.argv[1]))' .github/workflows/ci.yml; echo "exit=$?"
exit=0
```

## V6

- R1: `pull_request` with branches `main` and `dev`. `push` with branches
  `main` and `dev`.
- R2: `runs-on: ubuntu-latest`. `actions/setup-go` with
  `go-version-file: go.mod`.
- R2 pins: `actions/checkout@v7`, `actions/setup-go@v7`,
  `golangci/golangci-lint-action@v9`. Each is a major tag, not a commit SHA.
  `git ls-remote --tags` shows v7, v7, and v9 as the current major tags of the
  three repositories.
- R3: the steps run in this order: gofmt with a fail on output, the tidy
  check, `go build .`, `go vet ./...`, golangci-lint with `version: v2.15.0`,
  and `go test -race ./...`. Each step fails the job on a failure. The tag
  `v2.15.0` exists in `golangci/golangci-lint`.

## V7

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go'); echo "exit=$?"
exit=0
$ go vet ./...; echo "exit=$?"
exit=0
$ golangci-lint run ./...; echo "exit=$?"
0 issues.
exit=0
$ go test -race ./...; echo "exit=$?"
ok  	github.com/Harrison-Blair/hive	3.797s
... each package ok ...
exit=0
```

`gofmt -l` printed nothing. `golangci-lint --version` prints version 2.15.0.

## Remarks

- `testdata/ste-baseline.json` has the entry
  `spec/hive-v1/tasks/11-ci-workflow.md: 0`. `steFiles` in `ste_test.go` has
  the pattern `spec/hive-v1/tasks/*.md`, so R6 requires this entry.
- The brief copy in `spec/hive-v1/tasks/11-ci-workflow.md` differs from the
  live brief only in the V5 sentence. The manager stated that this change is a
  deliberate STE reword.
- The worktree was clean at the end: `git status --porcelain` printed nothing.
