# Pull request 47 body refresh

| Item | Value |
| --- | --- |
| Tier | small |
| Depends on | 8-pr47-body |
| Grant | `grants.md 2026-10-19 pull request 47 body update` |
| Pull request | https://github.com/Harrison-Blair/hive/pull/47 |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

`dev` is now at `8eaa2d8f`. Task 6 (the agent status `starting`) landed, and
the evidence of tasks 7 and 8 landed. Refresh the body of pull request #47 one
more time. Pull request #47 has base `main` and head `dev`.

## Rules

- Read `/home/penguin/.agents/skills/gh-summarize-pr/SKILL.md` first. Follow
  it in the `approve` mode. Apply the body without a question to the user.
- Make no source change. Do not commit a change to a source file.
- Do not merge. Do not push `main`. Do not change the base or the head branch.
- When a step does not agree with this brief, stop and report it with
  `hive inbox send --to manager-hive TEXT --agent NAME`.

## Requirements

### R1 Review

- The current body has a valid footer: reviewed commit `46925f2f`, base
  `928cfa95`. Review the incremental diff `46925f2f..` the head of `dev`.
- Change the Changes bullets when task 6 adds a material change. The agent
  status `starting` is one.
- Make the full LOC by category table again from the file statistics of the
  pull request. Get all pages. Make sure that the sums agree with the current
  totals of the pull request.

### R2 Body

- Run the four gate commands of `AGENTS.md` from the worktree at the head of
  `dev`. Give the result of each in Verification.
- Replace the footer. Write the reviewed commit, the base, and the exact model
  identifier that you run on.
- Keep both grant lines of the current body.
- Keep the title, unless the skill requires a change.
- Keep the body file in `.hive/tmp/` of the worktree. Do not commit it.

### R3 Apply

```sh
gh pr edit 47 --body-file <your body file>
```

Add `--title <title>` only when the title changes.

### R4 Report

Report the applied title and the URL with
`hive inbox send --to manager-hive TEXT --agent NAME`. Then run
`hive task complete`.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The body of pull request #47 names the `starting` status in Changes and gives the gate results at the new head. It has both grant lines. Its footer gives the reviewed commit. That commit equals the head of `dev` at review time. | V1 |
| AC2 | The LOC table sums agree with the additions, deletions, and changed files of the pull request. | V2 |
| AC3 | Base `main`, head `dev`, state `OPEN`. `main` remains at `928cfa95`. No commit changes a source file. | V3 |

## Verification items

### V1 Body

- Command: `gh pr view 47 --json title,body,headRefOid`
- Expected: AC1.

### V2 LOC

- Command: `gh pr view 47 --json additions,deletions,changedFiles`, and the
  sum of each column of the LOC table.
- Expected: the sums are equal to the pull request values. When `dev` moved
  after the edit, say so and compare at the reviewed commit.

### V3 No change

- Command: `gh pr view 47 --json baseRefName,headRefName,state`,
  `git ls-remote origin main`, and `git log --oneline dev..task/<task-id>`.
- Expected: AC3. The branch may contain evidence commits only.
