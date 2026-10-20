# Pull request 47 title and body update

| Item | Value |
| --- | --- |
| Tier | small |
| Depends on | 7-pr-dev-main |
| Grant | `grants.md 2026-10-19 pull request 47 body update` |
| Pull request | https://github.com/Harrison-Blair/hive/pull/47 |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

Change the title and the body of pull request #47 to the summary format of the
user. Pull request #47 has base `main` and head `dev`.

## Rules

- Read `/home/penguin/.agents/skills/gh-summarize-pr/SKILL.md` first. Follow
  it in the `approve` mode. Apply the title and the body without a question to
  the user.
- Make no source change. Do not commit a change to a source file.
- Do not merge. Do not push `main`. Do not change the base or the head branch.
- When a step does not agree with this brief, stop and report it with
  `hive inbox send --to manager-hive TEXT --agent NAME`.

## Requirements

### R1 Review

- No usable checkpoint footer exists. Review the full diff from the merge base
  `621a7afd` to the head of `dev`.
- Make the LOC by category table from the file statistics of the pull request.
  Get all pages. Make sure that the sum agrees with the totals of the pull
  request. At the time of opening the totals were +103472 -61252. Use the
  current totals.

### R2 Title and body

- Write the title and the fallback body of the skill: Summary, Changes, LOC by
  category, Verification, and the italic generation footer.
- For Verification, use the four gate commands of `AGENTS.md`. Run them from
  the worktree at the head of `dev`, and give the result of each.
- In the footer, write the exact model identifier that you run on.
- Keep this line from the current body, in the Summary or at the end:
  "Opened under `grants.md 2026-10-19 pull request from dev into main`. The
  owner merges." Add: "Edited under
  `grants.md 2026-10-19 pull request 47 body update`."
- Keep the body file in `.hive/tmp/` of the worktree. Do not commit it.

### R3 Apply

```sh
gh pr edit 47 --title <title> --body-file <your body file>
```

### R4 Report

Report the applied title and the URL with
`hive inbox send --to manager-hive TEXT --agent NAME`. Then run
`hive task complete`.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Pull request #47 has the new title and a body with Summary, Changes, LOC by category, and Verification. The body also has the italic footer with a model identifier and both grant lines. | V1 |
| AC2 | The LOC table sums agree with the additions and deletions of the pull request. | V2 |
| AC3 | Base `main`, head `dev`, state `OPEN`. `main` remains at `928cfa95`. No commit changes a source file. | V3 |

## Verification items

### V1 Body

- Command: `gh pr view 47 --json title,body`
- Expected: AC1.

### V2 LOC

- Command: `gh pr view 47 --json additions,deletions`, and the sum of each
  column of the LOC table.
- Expected: the sums are equal to `additions` and `deletions`.

### V3 No change

- Command: `gh pr view 47 --json baseRefName,headRefName,state`,
  `git ls-remote origin main`, and `git log --oneline dev..task/<task-id>`.
- Expected: AC3. The branch may contain evidence commits only.
