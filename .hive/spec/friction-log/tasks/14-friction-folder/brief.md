# Friction folder

| Item | Value |
| --- | --- |
| Wave | friction folder, wave 1 of 4 |
| Tier | decent |
| Depends on | none. Base: `dev` at `72a33b5b` or later |
| Grant | `grants.md 2026-10-19 friction folder task` |
| Evidence | `evidence/<task-id>/` in the worktree |

## Overview

This task moves the friction log from the file `reference/dogfood/friction.md`
to the folder `reference/dogfood/friction/`. Each entry gets its own file. A Go
test keeps the status table and the entry files in sync. The direct-edit rule
moves from the file to the folder. Do not change a decision in this brief. When
the files do not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `reference/dogfood/friction.md` has an intro, an entry template, an index
  table, a section `Open` with 11 entries, and a section `Resolved` with 2
  entries. Each entry has `**Issue:**`, `**Status:**`, `**Summary:**`, and
  `**Reproduction steps:**`.
- `.gitignore` is an allowlist. `gitignore_test.go` checks it.
- `ste_test.go` checks the files in `steFiles` against
  `testdata/ste-baseline.json`. The glob `spec/hive-v1/tasks/*.md` is in
  `steFiles`.
- The root package `main` has the tests `gitignore_test.go`, `skills_test.go`,
  and `ste_test.go`.
- `internal/agent/templates/project-agents.md` is the template of the manager
  instructions. `internal/agent/manager_prompt_test.go` checks the friction
  exception in that text.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `reference/dogfood/friction.md` | R1. Delete. |
| `reference/dogfood/friction/FRICTION.md` | R1. New. |
| `reference/dogfood/friction/_template.md` | R1. New. |
| `reference/dogfood/friction/entries/*.md` | R1, R2. New. |
| `reference/dogfood/friction/entries/resolved/*.md` | R1, R2. New. |
| `friction_test.go` | R3. New, in the root package `main`. |
| `.gitignore` | R1. Only when the new paths need allowlist rules. |
| `AGENTS.md` | R4, in the sections `Coordinator` and `Dogfooding`. |
| `.agents/skills/hive-coordinator/SKILL.md` | R4. |
| `.agents/skills/hive-manager/SKILL.md` | R4. |
| `internal/agent/templates/project-agents.md` | R4. |
| `internal/agent/manager_prompt_test.go` | R4. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `spec/hive-v1/tasks/<task-id>.md` | R6. New. A copy of this brief. |

Do not edit other files. Do not change an old evidence file that names the old
path.

## Requirements

### R1 Layout

1. Make `reference/dogfood/friction/FRICTION.md`. It has a short intro with the
   edit rule of R4. Then it has one status table with one row for each entry.
   The columns are `Issue`, `Status`, and `Decision`. The `Issue` cell is a
   link to the entry file. Put the open rows first, the newest first. Put the
   resolved rows after the open rows.
2. Make `reference/dogfood/friction/_template.md`. It is the template of an
   entry file, with the frontmatter and the body of R1 step 4.
3. Put each open entry in `reference/dogfood/friction/entries/<slug>.md`. Put
   each resolved entry in
   `reference/dogfood/friction/entries/resolved/<slug>.md`. Use a short
   kebab-case slug.
4. Each entry file starts with YAML frontmatter with these fields:
   - `title`: the title of the entry.
   - `status`: `Open`, or `Resolved YYYY-MM-DD`.
   - `opened`: `YYYY-MM-DD`. Use the first date in the entry text. When the
     text has no date, use the date of the commit that added the entry
     (`git log --diff-filter=A` or `git log -S`).
   - `decision`: one line, or empty.

   The body follows with `**Summary:**` and `**Reproduction steps:**`. Keep
   all text of each entry. Do not lose evidence.
5. Delete `reference/dogfood/friction.md`.
6. Add allowlist rules to `.gitignore` only when Git ignores a new path. Use
   the narrowest pattern. `gitignore_test.go` must pass.

### R2 Migration decisions

Set the status and the decision field of each entry as this table gives. The
Keep both entries at Resolved status. Leave both decision fields empty.

| Entry | Status | Decision |
| --- | --- | --- |
| Full Pi worktree, agent, and delivery bootstrap flow is missing | Resolved 2026-10-19 | Stale: `hive worktree` and `hive agent spawn/send/stop` exist. |
| `hive agent send` fails with `agent_not_ready` after a Pi `--pane` spawn | Merge into the next entry. Remove the separate entry. | - |
| An agent spawn can skip the first prompt | Open | A report task investigates first. The fix decision follows the report. |
| A verify fail leaves no workspace and no verifier for a re-verify | Resolved 2026-10-19 | By design. The manager skill fix arrives in wave 2. |
| A verifier spawn with `--pane` requires `--project` | Open | Hive derives the project from the task. Wave 2 includes the doc fix. |
| Any agent can read and consume the inbox of another role | Open | Only the owner's read marks records read. The coordinator can read a manager inbox and marks nothing. Any other caller gets an error. Workers get no inbox (wave 2). |
| An idle session burns tokens on supervisor nudges | Open | A quiet record never nudges (wave 2). |
| No hive path repairs a verified task | Open | Add `hive task reopen` (verified -> assigned, re-verify required) (wave 3). |
| A worker spawn into an assigned task keeps the old acting agent | Open | The spawn sets the acting agent and refuses while the old worker is live (wave 3). |
| An exited agent name blocks reuse | Open | Hive picks the next free name when `--name` is absent (wave 3). |
| No `hive task edit --brief` command | Open | Add `hive task edit --brief`, allowed only before `completed` (wave 4). |

For the merged entry, keep all summary text and all reproduction steps of both
entries. Mark the part that came from the `agent_not_ready` entry.

### R3 Sync test

Add `friction_test.go` in the root package `main`. It reads `FRICTION.md` and
each entry file. It fails in each of these cases:

1. A table row has no entry file.
2. An entry file has no table row.
3. The title or the status of a row is different from the frontmatter of its
   entry file.
4. An entry with the status `Open` is in `entries/resolved/`.
5. An entry with a `Resolved` status is in `entries/`.

Write the test first. Use the standard library only. Show that the test fails
for each of the five breaks: make the break in the working tree, run the test,
record the output, and then restore the file. Record each failure output in
the evidence.

### R4 Edit rule

Move the direct-edit exception from `friction.md` to the whole folder
`reference/dogfood/friction/`. Change it in `AGENTS.md` (sections
`Coordinator` and `Dogfooding`), `.agents/skills/hive-coordinator/SKILL.md`,
`.agents/skills/hive-manager/SKILL.md`,
`internal/agent/templates/project-agents.md`, and
`internal/agent/manager_prompt_test.go`. The rule also says this: to resolve an
entry, set its status, move its file to `entries/resolved/`, and update its
row. Keep the commit on `dev` with the subject `docs(dogfood): SUMMARY`, and
keep the push of `dev`. Change the test in `manager_prompt_test.go` first, run
it, and make sure that it fails before you change the template.

### R5 STE

Write the new instruction text in ASD-STE100 Strict. Use each term only with
its meaning in `STE-GLOSSARY.md`. The counts in `testdata/ste-baseline.json`
must not rise. The files in `reference/dogfood/friction/` stay out of
`steFiles`.

### R6 Brief copy

Copy this brief to `spec/hive-v1/tasks/<task-id>.md` with a baseline of 0.
When the copy causes a violation, change the words of that sentence and keep
its meaning. Record each change in the evidence.

## Not in scope

The fixes of waves 2 to 4, a new hive command, old evidence files, `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | List `reference/dogfood/friction/` and `git ls-files reference/dogfood` | `FRICTION.md`, `_template.md`, 8 files in `entries/`, 4 files in `entries/resolved/`. Git tracks each file. `friction.md` is gone. |
| V2 | Compare each entry file with its text in `friction.md` at the base commit | No summary text and no reproduction step is lost. The merged entry has the text of both old entries. |
| V3 | Read the frontmatter of each entry and the table of `FRICTION.md` | Each status and decision agrees with R2. Open rows come first, the newest first. |
| V4 | Make each of the five breaks of R3, run `go test -run Friction .`, and restore | Each break fails the test with a message that names the break. The restored tree passes. |
| V5 | `grep -rn 'friction.md' AGENTS.md .agents/skills internal/agent` | No match names the old file. Each file of R4 gives the folder rule and the resolve steps. |
| V6 | `go test ./internal/agent/ -run Manager` and `go test -run 'TestSTE\|Gitignore' .` | Pass. No count in `testdata/ste-baseline.json` rose. |
| V7 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `docs(dogfood): move the
friction log into a folder`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V7 results to the manager
with `hive inbox send`. Read only your own inbox.
