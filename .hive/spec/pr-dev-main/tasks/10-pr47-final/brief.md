# Pull request 47 final body refresh

| Item | Value |
| --- | --- |
| Tier | small |
| Depends on | 9-pr47-refresh |
| Grant | `grants.md 2026-10-19 pull request 47 body update` |
| Pull request | https://github.com/Harrison-Blair/hive/pull/47 |
| Evidence | in the inbox only, see R5 |

## Overview

Refresh the body of pull request #47 one more time. Pull request #47 has base
`main` and head `dev`. This task lands no evidence on `dev`. Thus the head of
`dev` does not move after the refresh, and the footer stays correct.

## Rules

- Read `/home/penguin/.agents/skills/gh-summarize-pr/SKILL.md` first. Follow
  it in the `approve` mode. Apply the body without a question to the user.
- Make no commit on any branch. Push nothing. This rule applies to the worker
  and to the verifier. It overrides each skill step that commits or pushes
  evidence. Keep the evidence files uncommitted in the worktree.
- Make no source change. Do not merge. Do not change the base or the head
  branch.
- When a step does not agree with this brief, stop and report it with
  `hive inbox send --to manager-hive TEXT --agent NAME`.

## Requirements

### R1 Review

- Read the reviewed commit from the footer of the current body. Review the
  incremental diff from that commit to the head of `dev`.
- Change the Changes bullets only when the increment adds a material change.

### R2 LOC

Make the full LOC by category table again from the file statistics of the
pull request. Get all pages. Make sure that the sums agree with the current
totals of the pull request.

### R3 Body

- Run the four gate commands of `AGENTS.md` from the worktree at the head of
  `dev`. Give the result of each in Verification.
- Replace the footer. The reviewed commit is the head of `dev` at review time.
  Write the base and the exact model identifier that you run on.
- Keep both grant lines of the current body. Keep the title, unless the skill
  requires a change.
- Keep the body file in `.hive/tmp/` of the worktree.

### R4 Apply

```sh
gh pr edit 47 --body-file <your body file>
```

Then run `gh pr view 47 --json headRefOid`. It must equal the reviewed commit
in the footer. When it does not, `dev` moved: do the review again from R1.

### R5 Report

Send the evidence to the manager in the inbox with
`hive inbox send --to manager-hive TEXT --agent NAME`: the title, the URL,
`headRefOid`, the reviewed commit, the LOC totals and the PR totals, and the
result of each gate command. Then run `hive task complete`.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The footer reviewed commit equals `headRefOid` of pull request #47 and the head of `origin/dev`. | V1 |
| AC2 | The LOC table sums agree with the additions, deletions, and changed files of the pull request. | V2 |
| AC3 | The body has both grant lines and the gate results at the head. | V1 |
| AC4 | Base `main`, head `dev`, state `OPEN`. `main` remains at `928cfa95`. The task branch has no commit. | V3 |

## Verification items

The verifier checks the live pull request. It writes `VERIFIED.md` and does
not commit it.

### V1 Body and head

- Command: `gh pr view 47 --json title,body,headRefOid` and
  `git ls-remote origin dev`.
- Expected: AC1 and AC3.

### V2 LOC

- Command: `gh pr view 47 --json additions,deletions,changedFiles`, and the
  sum of each column of the LOC table.
- Expected: AC2.

### V3 No change

- Command: `gh pr view 47 --json baseRefName,headRefName,state`,
  `git ls-remote origin main`, `git ls-remote origin task/<task-id>`, and
  `git log --oneline dev..task/<task-id>`.
- Expected: AC4. The task branch has no commit and is not on `origin`.
