# Pull request from dev into main

| Item | Value |
| --- | --- |
| Tier | small |
| Depends on | none |
| Grant | `grants.md 2026-10-19 pull request from dev into main` |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

Open one pull request on GitHub with base `main` and head `dev`. The pull
request replaces the old fledge code on `main` with hive. The user reviews and
merges it. You do not merge it.

Facts:

- `origin/main` is `928cfa95`. It is the last merged pull request from `dev`
  (#16, the old fledge code).
- `dev` is approximately 422 commits ahead of `main` with the hive
  reauthoring. The merge base is `621a7afd`.

## Rules

- Make no source change. Do not commit a change to a source file.
- Do not merge. Do not push `main`. Do not change the base or the head
  branch of the pull request.
- Open one pull request only. First run `gh pr list --base main --head dev`.
  When a pull request is open, do not make a second one. Report it.
- When a step does not agree with this brief, stop and report it with
  `hive inbox send --to manager-hive TEXT --agent NAME`.

## Requirements

### R1 Body

Write the body from the repository. Include:

1. What hive is, in two to four sentences. Use `AGENTS.md` and
   `spec/hive-v1/README.md`.
2. A link to the spec at `spec/hive-v1/README.md`.
3. The landed tasks T01 to T21, one line each. Use
   `spec/hive-v1/05-tracker.md` and `git log`.
4. The sidebar layout wave: tasks 3-manager-adopts-parent,
   4-shell-wait-friction, and 5-tab-labels, from `git log`.
5. The gate status on `dev`: run the four gate commands of `AGENTS.md` from
   the worktree at the head of `dev`, and give the result of each.
6. The line: "Opened under `grants.md 2026-10-19 pull request from dev into
   main`. The owner merges."

### R2 Pull request

Run, from the worktree:

```sh
gh pr create --base main --head dev --title "Hive v1: replace fledge with hive" --body-file <your body file>
```

Keep the body file in `.hive/tmp/` of the worktree. Do not commit it.

### R3 Report

Report the pull request URL with
`hive inbox send --to manager-hive TEXT --agent NAME`. Then run
`hive task complete`.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | One open pull request with base `main` and head `dev` exists, with the title and a body that holds the six items of R1. | V1 |
| AC2 | No commit on any branch changes a source file. `main` remains at `928cfa95`. | V2 |

## Verification items

### V1 Pull request

- Command: `gh pr view --json url,baseRefName,headRefName,title,body,state <number>`
- Expected: base `main`, head `dev`, state `OPEN`, the title, and the six
  items of R1 in the body.

### V2 No change

- Command: `git ls-remote origin main` and
  `git log --oneline dev..task/<task-id>`.
- Expected: `main` is `928cfa95`. The task branch has no commit that changes
  a source file. The branch may contain evidence commits only.
