# Friction log structure and direct edits

| Item | Value |
| --- | --- |
| Wave | friction log |
| Tier | decent |
| Depends on | none. Base: `dev` at `dadacf88` or later |
| Grant | `grants.md 2026-10-19 friction log task` |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

This task changes the structure of `reference/dogfood/friction.md`, adds one
friction case, and lets the coordinator and a manager edit that file directly.
This task runs at the same time as `11-ci-workflow`. The files do not overlap.
Do not change a decision in this brief. When the files do not agree with this
brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `reference/dogfood/friction.md` has an intro, the section `## Entry template`,
  and the section `## Entries`. Each entry starts with a `**Issue:**` line. The
  file adds new entries at the end. Thus the oldest entry is first.
- Each entry has a `**Status:**` line with `Open` or `Resolved YYYY-MM-DD`.
- The entry `A verifier spawn with --pane skips the first prompt` has the
  status `Resolved 2026-10-19`.
- `ste_test.go` checks the files in `steFiles` against
  `testdata/ste-baseline.json`. `AGENTS.md` and the two skill files are in
  `steFiles`. `reference/dogfood/friction.md` is not in `steFiles`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `reference/dogfood/friction.md` | R1, R2, R3, R4. |
| `AGENTS.md` | R4, in the sections `Coordinator` and `Dogfooding`. |
| `.agents/skills/hive-coordinator/SKILL.md` | R4. |
| `.agents/skills/hive-manager/SKILL.md` | R4. |
| `internal/agent/templates/project-agents.md` | R6, the manager friction edit exception. |
| `internal/agent/manager_prompt_test.go` | R6, test the rendered home path. |
| `testdata/ste-baseline.json` | Add a baseline of 0 or lower a count. |
| `spec/hive-v1/tasks/<task-id>.md` | New. A copy of this brief. |

Do not edit other files.

## Requirements

### R1 Order of the file

Put the parts of `friction.md` in this order:

1. The intro.
2. The entry template.
3. An index table with the columns `Issue` and `Status`. Each `Issue` cell is
   a link to the heading of its entry. Put the `Open` rows first, then the
   `Resolved` rows. In each group, use the same order as the sections of R2.
4. The section `## Open`.
5. The section `## Resolved`.

### R2 Entries

- Give each entry its own `### Title` heading. The title is the text of its
  `**Issue:**` line. Make sure that each index link goes to its heading.
- Put each `Open` entry in `## Open`. Put each `Resolved` entry in
  `## Resolved`.
- In each section, put the newest entry first. The newest entry is the entry
  that is nearest to the end of the file today.
- Keep the full text of each entry. Do not change the text of an entry,
  except for R3.
- Change the entry template so that it shows the `### Title` heading.

### R3 New friction case

Extend the entry `A verifier spawn with --pane skips the first prompt` with
this case. On 2026-10-19, the manager spawned the worker `t11-worker-1` with
`hive agent spawn --role worker` and with no `--pane`. The spawn printed
`prompted agent` and `agent t11-worker-1 live`. Pi did not get the first
prompt. The pane stayed idle at context 0k. The workaround was
`herdr pane send-text` and then `herdr pane send-keys Enter`.

The problem occurs again. Thus set the status of this entry to `Open`, move
the entry to `## Open`, and keep the earlier fix text. Change the title so
that it does not name only the verifier and `--pane`. Update the index table.

### R4 Direct edit rule

Write this rule in the `friction.md` intro, in `AGENTS.md` (the sections
`Coordinator` and `Dogfooding`), and in the two skill files:

- The coordinator and a manager can edit `reference/dogfood/friction.md`
  directly. They can add an entry, extend an entry, set the status of an entry
  to `Resolved YYYY-MM-DD`, and keep the index table in sync.
- They edit only `friction.md` in this way.
- They commit on `dev` with a subject `docs(dogfood): SUMMARY` and push `dev`.
  They use no worker and no verifier for this edit.
- For all other files, keep the rule that the coordinator and a manager never
  change a project source file.

Write the new text in ASD-STE100 Strict. Use each term only with its meaning
in `STE-GLOSSARY.md`. The counts in `testdata/ste-baseline.json` must not
rise.

### R5 Brief copy

Copy this brief to `spec/hive-v1/tasks/<task-id>.md` with a baseline of 0 in
`ste_test.go`. When the copy causes a violation, change the words of that
sentence and keep its meaning. Record each change in the evidence.

### R6 Manager prompt exception

The user approved this scope through the coordinator on 2026-10-19.
The manager template uses the absolute home path for the hive friction file.

- Add the exception to `internal/agent/templates/project-agents.md`.
- Use the `{{HOME}}` value to name `reference/dogfood/friction.md` in the home.
- Keep the baseline count for the template at 0.
- Add a test in `internal/agent/manager_prompt_test.go`. The test checks the rendered home path.
- Run the test before and after the template change. The test fails before the change and passes after it.

## Not in scope

`ste_test.go`, code and tests that this brief does not name, and `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Count the entries before and after | The same count. Each old entry has the same text, except R3. Record a diff check of the entry bodies. |
| V2 | Read the order of `friction.md` | R1 order. `## Open` before `## Resolved`. Newest entry first in each section. |
| V3 | Each link in the index table | It goes to an existing `###` heading. The status in each row is the same as in its entry. |
| V4 | The extended entry | It has the R3 case, the status `Open`, and it is in `## Open`. |
| V5 | `grep` the R4 rule in five files | The rule is in each file. The source file rule stays for all other files. |
| V6 | `go test -run TestSTE ./...` | Pass. No count in `testdata/ste-baseline.json` rose. |
| V7 | The manager prompt test and gate | The test fails before the template edit and passes after it. `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `docs(dogfood): restructure
the friction log`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V7 results to the
manager with `hive inbox send`.
