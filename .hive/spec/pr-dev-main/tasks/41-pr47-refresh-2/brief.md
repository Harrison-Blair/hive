# Pull request 47 body refresh 2

| Item | Value |
| --- | --- |
| Tier | small |
| Depends on | none. Base: `dev` at `68ae1dc9` |
| Edit grant | `grants.md 2026-10-19 pull request 47 body update` |
| Landing grant | `grants.md 2026-10-19 pull request 47 body refresh 2` |
| Pull request | https://github.com/Harrison-Blair/hive/pull/47 |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

Write the body of pull request #47 again. Pull request #47 has base `main` and
head `dev`. The current body is stale. Its footer reviewed commit is
`dadacf88`. It says "No GitHub checks", but the CI workflow now runs on `dev`.
The CI run `39339786673` (push, commit `68ae1dc9`) passed.

## Rules

- Read `/home/penguin/.agents/skills/gh-summarize-pr/SKILL.md` first. Follow
  it in the `approve` mode. Apply the body without a question to the user.
- Make no source change. Do not merge. Do not change the base or the head
  branch. Push nothing to `dev` or `main`.
- Commit the evidence only on the task branch and push the task branch.
- When a step does not agree with this brief, stop and report it with
  `hive inbox send --to manager-hive TEXT --agent NAME`.

## Requirements

### R1 Review

- Review the incremental diff from `dadacf88` to the head of `origin/dev`.
- Add a Changes bullet for each material change. Cover at least: the friction
  folder, inbox ownership, quiet records do not nudge, `hive task reopen`,
  `hive task edit --brief`, worker respawn, `hive task verify` refuses a dirty
  worktree or an unpushed branch, the CI flake fix and the Ubuntu-only CI, the
  first prompt fix, the hive message marker and the harness bypass flag, and
  the nudge without `--home`.
- Keep the existing bullets that are still correct. Change a bullet that the
  increment makes wrong.

### R2 LOC

Make the full LOC by category table again from the file statistics of the
pull request. Get all pages. Make sure that the sums agree with the current
totals of the pull request.

### R3 Body

- In Verification, replace "No GitHub checks" with the CI result for the
  reviewed commit. Cite the run ID and its URL from `gh run list --branch dev`.
- Run the four gate commands of `AGENTS.md` from the worktree at the head of
  `origin/dev`. Give the result of each in Verification.
- Replace the footer. The reviewed commit is the head of `origin/dev` at
  review time. Write the base `928cfa95e0906315fcb4a65a5cdbdef1470a0503` and
  the exact model identifier that you run on.
- Keep both grant lines of the current body. Keep the title, unless the skill
  requires a change.
- Keep the body file in `.hive/tmp/` of the worktree.

### R4 Apply

```sh
gh pr edit 47 --body-file <your body file>
```

Then run `gh pr view 47 --json headRefOid`. It must equal the reviewed commit
in the footer. When it does not, `dev` moved: do the review again from R1.

### R5 Evidence and report

- Write `evidence/<task-id>/V1.md` to `V3.md` with each command and its output.
  Copy the applied body to `evidence/<task-id>/body.md`.
- Commit with subject `docs(evidence): record <task-id> body refresh`. Push
  the task branch.
- Run `hive task complete`. Send the commit SHA, the reviewed commit,
  `headRefOid`, the LOC totals and the PR totals, and the result of each gate
  command to the manager with `hive inbox send`. Read only your own inbox.

The landing of the evidence moves `dev` after the review. The footer then
names the reviewed commit, which is correct. Do not refresh the body again for
the evidence commit.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The footer reviewed commit equals `headRefOid` of pull request #47 and the head of `origin/dev` at review time. | V1 |
| AC2 | The LOC table sums agree with the additions, deletions, and changed files of the pull request. | V2 |
| AC3 | The body has both grant lines, the CI run result, the gate results at the reviewed commit, and a bullet for each R1 item. | V1 |
| AC4 | Base `main`, head `dev`, state `OPEN`. `main` remains at `928cfa95`. The task branch has only evidence files. | V3 |

## Verification items

The verifier checks the live pull request. It writes
`evidence/<task-id>/VERIFIED.md`, never `VERIFIED.md` at the checkout root. It
commits and pushes the file on the task branch before it runs
`hive task verify`.

### V1 Body and head

- Command: `gh pr view 47 --json title,body,headRefOid`,
  `git ls-remote origin dev`, and `gh run view <run ID>`.
- Expected: AC1 and AC3.

### V2 LOC

- Command: `gh pr view 47 --json additions,deletions,changedFiles`, and the
  sum of each column of the LOC table.
- Expected: AC2.

### V3 No change

- Command: `gh pr view 47 --json baseRefName,headRefName,state`,
  `git ls-remote origin main`, and `git diff --stat origin/dev...task/<task-id>`.
- Expected: AC4.
