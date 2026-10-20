# Home spec planning review

| Item | Value |
| --- | --- |
| Wave | report only, beside wave 4 |
| Tier | strong |
| Depends on | none |
| Grant | none. No code change, no commit, no landing. |
| Input | `/home/penguin/source/hive/projects/hive/plans/home-spec-planning/MAP.md` |
| Output | `/home/penguin/source/hive/.hive/tmp/reports/home-spec-planning-review.md` |

## Overview

This task is a spec review of a draft plan. Do not change a source file. Do
not commit. Do not push. This brief replaces the commit, push, gate, and
evidence steps of the worker skill. The report file is the only output.

## Scope

Read `MAP.md` and each code file and doc that it names. Find:

- gaps,
- edge cases,
- contradictions inside the plan,
- conflicts with the current code on `dev`, `AGENTS.md`, the skills in
  `.agents/skills/`, `spec/hive-v1/05-tracker.md`, and the wave 4 tasks
  (22-task-edit-brief, 23-verify-clean-branch, and 24-19-evidence-fix). Tasks
  23 and 24 are on `dev`.

Do not choose a design. When a finding has more than one fix, list the
options.

## Report

Write the report to the output path. For each finding give:

1. What: one sentence.
2. Why it matters.
3. Evidence: `file:line` for each claim.
4. Recommendation, or the options.

Number the findings. Put the count at the top.

## Finish

Run `hive task complete`. Send the report path and the finding count to the
manager with `hive inbox send --to manager-hive TEXT --agent NAME`.
