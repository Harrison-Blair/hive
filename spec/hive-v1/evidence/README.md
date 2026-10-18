# Evidence

Each task has one directory here with the name of the task ID. The directory
holds one file for each verification item of the brief, `V<n>.md`, and one file
`VERIFIED.md` from the verifier. A task reaches the status "evidence" when each
`V<n>.md` exists. It reaches the status "verified" when `VERIFIED.md` records a
pass for each item.

## Rules

- The worker writes each `V<n>.md` in its worktree, commits it on the task branch, and pushes the branch.
- The command in an evidence file is the exact command that ran, from the root of the worktree. The output is the exact output, in a fenced block.
- When an output has more than 200 lines, keep the first 50 lines and the last 100 lines. Write the number of removed lines between them.
- A test item has two runs. The red run shows the new test fail before the change. The green run shows the same test pass after the change.
- The red run fails on an assertion. A compile failure alone does not count. When the change adds a symbol that the test needs, the worker first adds the symbol with an empty body. Then it records the red run.
- Each run ends with the line `echo "exit $?"` and its output, so the file shows the exit status.
- The red run and the green run each start with the line `commit <short hash>` of the worktree at that run.
- An observation outside the command output goes into its own fenced block with the command that made it. Examples are a `herdr pane read` capture and a workspace list.
- A live check records its cleanup commands and their output as the last block of the item.
- The gate item records the four gate commands and their output. A `gofmt -l` run counts as a pass only when it prints nothing.
- An evidence file claims no result that its output does not show.
- The verifier does not edit a `V<n>.md`. It records its own runs in `VERIFIED.md`.
- When the verifier records a fail, the worker continues in the same worktree. It corrects the behavior, replaces the affected `V<n>.md`, commits, and pushes. The verifier then runs again and replaces `VERIFIED.md`.

## Format of `V<n>.md`

```markdown
# <task-id> V<n> <title>

| Item | Value |
| --- | --- |
| Task | <task-id> |
| Item | V<n> |
| Proves | AC<a>, AC<b> |
| Commit | <short hash of the worktree at the green run> |
| Date | YYYY-MM-DD |
| Agent | worker, <harness>, <model> |

## Red run

A fenced sh block with the command. A fenced text block with the output.
Test items only.

## Green run

A fenced sh block with the command. A fenced text block with the output.
For an item without a test, name this section "Run".

## Result

One sentence that states what the output shows.
```

## Format of `VERIFIED.md`

```markdown
# <task-id> VERIFIED

| Item | Value |
| --- | --- |
| Task | <task-id> |
| Commit | <short hash at the verifier run> |
| Date | YYYY-MM-DD |
| Agent | verifier, <harness>, <model> |
| Verdict | pass or fail |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | pass | |
| V2 | AC2, AC3 | fail | What differed from the expected output. |

## Runs

One section for each item, with a fenced sh block and a fenced text block
from the run of the verifier.
```
