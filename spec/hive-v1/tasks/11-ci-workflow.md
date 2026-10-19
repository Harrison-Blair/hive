# Continuous integration with GitHub Actions

| Item | Value |
| --- | --- |
| Wave | continuous integration |
| Tier | decent |
| Depends on | none. Base: `dev` at `dadacf88` or later |
| Grant | `grants.md 2026-10-19 continuous integration task` |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

This task adds one GitHub Actions workflow. The workflow runs the gate of
`AGENTS.md` on a pull request and on a push. The check is advisory. Do not
change a decision in this brief. When the code or GitHub does not agree with
this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `.gitignore` is an allowlist. Today `git check-ignore -v .github/workflows/ci.yml`
  prints `.gitignore:2:*`. Git ignores the workflow file.
- `go.mod` gives `go 1.27.2`.
- `origin/dev` at `dadacf88` is not tidy. `go mod tidy` moves
  `github.com/charmbracelet/x/term v0.2.2` from the indirect block to the
  direct block of `go.mod`. The tidy check of R3 fails until this change is
  on the branch.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.github/workflows/ci.yml` | New. The workflow. |
| `.gitignore` | Add the narrowest allow rule for the workflow file. |
| `gitignore_test.go` | Add a case that fails when Git ignores `.github/workflows/ci.yml`. |
| `go.mod`, `go.sum` | Only the change that `go mod tidy` makes. |
| `spec/hive-v1/tasks/<task-id>.md` | New. A copy of this brief. |

Do not edit other files.

## Requirements

### R1 Triggers

The workflow starts on `pull_request` with the target branch `main` or `dev`.
The workflow starts on `push` to `main` or `dev`.

### R2 Environment

- Use `runs-on: ubuntu-latest`.
- Use `actions/setup-go` with `go-version-file: go.mod`.
- Pin each action to its major tag, for example `actions/checkout@v5`. Do not
  pin an action to a commit SHA. Use the current major tag of each action.

### R3 Steps

The workflow runs each of these steps. Each step fails the job on a failure.

1. `gofmt -l $(git ls-files -co --exclude-standard '*.go')`. The step fails
   when the command prints a line.
2. A tidy check. Run `go mod tidy`. Then run `git diff --exit-code -- go.mod go.sum`.
3. `go build .`
4. `go vet ./...`
5. `golangci-lint run ./...` with golangci-lint at version `v2.15.0`. Use
   `golangci/golangci-lint-action` at its major tag, with `version: v2.15.0`.
6. `go test -race ./...`

### R4 Git ignore

- Add the narrowest `!` rule or rules that let Git track
  `.github/workflows/ci.yml`. Do not remove the leading `*`. Do not allow other
  files below `.github/`.
- Add a case to `gitignore_test.go`. The case fails when Git ignores
  `.github/workflows/ci.yml`. Show that the case fails without the new rule and
  passes with it. Record both outputs in the evidence.

### R5 Tidy

Run `go mod tidy` in the worktree. Commit only the change that it makes to
`go.mod` and `go.sum`.

### R6 Brief copy

Copy this brief to `spec/hive-v1/tasks/<task-id>.md`. If `ste_test.go`
fails on the new file, add the file to `steFiles` with a baseline of 0 only
when `ste_test.go` requires it, and report it.

## Not in scope

Branch protection, a required status check, a change to a GitHub repository
setting, a merge into `main`, or a push to `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | `git check-ignore -v .github/workflows/ci.yml` | No output. Exit status 1. |
| V2 | `git ls-files .github` | Only `.github/workflows/ci.yml`. |
| V3 | The new `gitignore_test.go` case without the new rule, then with the rule | Fail, then pass. |
| V4 | `go mod tidy && git diff --exit-code -- go.mod go.sum` | Exit status 0. |
| V5 | Run `actionlint .github/workflows/ci.yml` when `actionlint` is on PATH. Otherwise, parse the YAML with `python3 -c 'import yaml,sys; yaml.safe_load(open(sys.argv[1]))'`. Record which tool ran. | No error. |
| V6 | Read `ci.yml` against R1, R2, and R3 | Each trigger, each step, the Go version source, and each pin agree. |
| V7 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

The first real Actions run happens after the manager lands the task on `dev`.
That run is not part of this task.

## Finish

Commit on the task branch with a subject such as `ci: add the GitHub Actions
gate workflow`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V7 results to the
manager with `hive inbox send`.
