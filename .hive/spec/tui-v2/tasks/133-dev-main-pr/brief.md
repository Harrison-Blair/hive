# Open the dev to main pull request

| Item | Value |
| --- | --- |
| Feature | `tui-v2` (the pull request carries the TUI work) |
| Tier | small |
| Depends on | 132-tui-clock-tests. Base: `dev` after 132 lands |
| Grant | PR: `grants.md 2026-10-19 pull request from dev into main`. Evidence landing: goal `g8-pr-dev-main`. |
| Evidence | `.hive/spec/tui-v2/tasks/133-dev-main-pr/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

PR #47 merged `dev` into `main` on 2026-10-22 (merge `9569ba99`, dev head
`499eec5f`). It carried g1-hive-tui waves A-D. Since that merge, task 56 and
the TUI goals g3, g4, g5, and g6 landed on `dev`, and task 132 fixed the CI
tests. Open one new pull request with base `main` and head `dev`. The user
reviews and merges it. You do not merge it.

## Rules

- Make no source change. Do not edit a product file, a test, or a spec file.
- Do not merge. Do not push `main`. Do not change the base or the head branch.
- Open one pull request only. First run `gh pr list --base main --head dev --state open`. The old PR #47 merged on 2026-10-22. When a pull request is open, use it and do not open a second one. Report it.
- When a step does not agree with this brief, stop and report it with `hive inbox send --to manager-hive TEXT --agent NAME`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/tui-v2/tasks/133-dev-main-pr/brief.md` | A copy of this brief. |
| `.hive/spec/tui-v2/tasks/133-dev-main-pr/evidence/` | The evidence files, and `VERIFIED.md` from the verifier. |
| `testdata/ste-baseline.json` | AC6. To add the brief copy with a count of 0 in the STE ratchet. |

Do not edit other files.

## Requirements

### R1 Body

Write the body from the repository. Include:

1. What hive is, in two to four sentences. Use `AGENTS.md` and `.hive/spec/hive-v1/README.md`.
2. A link to `.hive/spec/hive-v1/README.md` on `dev`.
3. A section "What changed since #47". Give one line for each task that landed on `dev` after PR 47 merged. The tasks are task 56-doctor-home-version, g3-chat-status (122), g4-agents-live-default (123), g5-live-agent-activity (124, 125), g6-hive-ask (126, 129, 131), task 132, and the docs commits. Name each other task or docs commit in the range. Use `git log origin/main..dev` and the task records to get the task IDs, titles, and commit subjects.
3a. Add one short context line that g1-hive-tui waves A-D (tasks 73 to 120) are already on `main` through PR 47. Do not list them as new work.
4. The gate status on `dev` at the head commit of the pull request. Run the five gate commands of `AGENTS.md` in the worktree after `git fetch origin`. Give the result of each.
5. The CI status of the head commit: the run ID and the conclusion from `gh run list --branch dev --limit 1`.
6. The line: "Opened under `grants.md 2026-10-19 pull request from dev into main`. The owner merges."

### R2 Pull request

Run, from the worktree:

```sh
gh pr create --base main --head dev --title "Hive TUI dashboard, agent screens, and hive ask" --body-file <your body file>
```

Keep the body file in `.hive/tmp/` of the worktree. Do not commit it. Print
the URL and the number.

### R3 CI

Run `gh pr checks <number> --watch --interval 10`. Record the check name, the
run ID, and the result. When a check fails, stop and report. Do not fix
source in this task.

### R4 Evidence

- Write `V1.md` to `V5.md` in the evidence folder.
- Record the pull request URL and number in each item that uses it.
- Commit and push the brief copy and the evidence on the task branch. The verifier writes `VERIFIED.md` in the same folder, commits it, and pushes it.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | One pull request has base `main`, head `dev`, state `OPEN`, and the title of R2. | V1 |
| AC2 | The body has the six items of R1. | V1 |
| AC3 | The CI check on the pull request passes. | V2 |
| AC4 | The task branch changes only the brief copy and the evidence files. No source file changes. | V3 |
| AC5 | `main` does not move. The pull request is not merged. The base and the head do not change. | V4 |
| AC6 | The gate of `AGENTS.md` passes in the worktree. | V5 |

## Verification items

### V1 Pull request

- Proves: AC1, AC2.
- Command: `gh pr list --base main --head dev --state open --json number,url,title,headRefName,baseRefName,state` and `gh pr view <number> --json body --jq .body`.
- Expected: one open pull request, base `main`, head `dev`, and the body has each item of R1.

### V2 CI

- Proves: AC3.
- Command: `gh pr checks <number> --json name,state,link` and `gh run view <run-id>`.
- Expected: each check has state `SUCCESS`. The run ID is the run on the head of the pull request.

### V3 Diff

- Proves: AC4.
- Command: `git diff dev..HEAD --stat` and `git status --short`.
- Expected: only `.hive/spec/tui-v2/tasks/133-dev-main-pr/` files. The worktree is clean.

### V4 Main and pull request state

- Proves: AC5.
- Command: `git ls-remote origin main` and `gh pr view <number> --json state,mergedAt,baseRefName,headRefName`.
- Expected: `origin/main` is `9569ba99281aec1105393ba75685268fa7e0aaa9` or the value at the start. State `OPEN`, `mergedAt` null, base `main`, head `dev`.

### V5 Gate

- Proves: AC6.
- Command: the five gate commands of `AGENTS.md` from the worktree root.
- Expected: `gofmt -l` and `go mod tidy -diff` print nothing. `go vet`, `golangci-lint`, and `go test -race` exit 0.

## Finish

Commit on the task branch with a subject such as
`docs(pr): open the dev to main pull request`. Push the task branch. Write
the evidence and run `hive task complete`. Send the commit SHA, the pull
request URL, and the V1-V5 results to the manager with
`hive inbox send --to manager-hive`.
