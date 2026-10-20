# Module tidy check in the gate

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, repair after wave C |
| Tier | decent |
| Depends on | none. Base: `dev` at `bac988e9` or later |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The CI job `gate` fails on `dev` from merge 7204d699 to 7def7814. The step
"Check module tidiness" fails. `internal/tui` imports
`github.com/charmbracelet/x/ansi` directly, but `go.mod` lists it as
indirect, and `go.sum` needs new lines. The gate in `AGENTS.md` and in the
skills has no tidy check, so each local gate passed. This task makes the
module tidy and adds the tidy check to the gate. The friction entry is
`reference/dogfood/friction/entries/local-gate-misses-ci-tidy.md`. Do not edit
it. When the code does not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `.github/workflows/ci.yml` runs `go mod tidy` and then
  `git diff --exit-code -- go.mod go.sum`.
- `go mod tidy -diff` prints the diff and exits 1 when the module is not
  tidy. It does not change a file. It prints nothing and exits 0 when the
  module is tidy.
- The gate has four commands in `AGENTS.md` (section "Tests and the gate")
  and in `.agents/skills/hive-worker/SKILL.md`.
  `.agents/skills/hive-verifier/SKILL.md` and `.hive/spec/EVIDENCE-FORMAT.md`
  say "the four gate commands".

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `go.mod`, `go.sum` | R1. Only the output of `go mod tidy`. |
| `.github/workflows/ci.yml` | R2. Only the step "Check module tidiness". |
| `AGENTS.md`, `.agents/skills/hive-worker/SKILL.md`, `.agents/skills/hive-verifier/SKILL.md`, `.hive/spec/EVIDENCE-FORMAT.md` | R2. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy, when `hive task create` did not write it. |

Do not edit other files. Do not edit Go source.

## Requirements

### R1 Tidy module

Run `go mod tidy`. Commit the changes to `go.mod` and `go.sum`. After this,
`go mod tidy -diff` prints nothing and exits 0.

### R2 Same tidy check in CI and in the gate

- Add `go mod tidy -diff` to the gate as the second command, after
  `gofmt -l`, in `AGENTS.md` and in the worker skill. The gate then has five
  commands.
- Write the rule for the new command: it must print nothing. Use the style of
  the `gofmt -l` rule.
- Change "the four gate commands" to "the five gate commands" in the verifier
  skill and in `.hive/spec/EVIDENCE-FORMAT.md`. Where a file names the
  `gofmt -l` rule, add the same rule for `go mod tidy -diff`.
- In `.github/workflows/ci.yml`, change the run of the step "Check module
  tidiness" to `go mod tidy -diff`. Do not change other steps.
- Keep each changed instruction text in ASD-STE100 Strict. `ste_test.go` must
  pass.

## Not in scope

Go source changes. Other CI steps. A CI check in the manager flow. `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | On the base commit, run `go mod tidy -diff`. On the task branch, run it again | The base prints a diff of `go.mod` and `go.sum` and exits 1. The task branch prints nothing and exits 0. |
| V2 | On the task branch, add `// indirect` to the `github.com/charmbracelet/x/ansi` line of `go.mod` in the working tree only. Run `go mod tidy -diff`. Then run `git checkout -- go.mod` | The command exits 1 and prints the diff. The restore leaves the tree clean. |
| V3 | `git grep -n -e 'four gate' -e 'go mod tidy' -- AGENTS.md .agents/skills .hive/spec/EVIDENCE-FORMAT.md .github/workflows/ci.yml` | No line says "four gate". The gate in `AGENTS.md` and the worker skill, the verifier skill, `EVIDENCE-FORMAT.md`, and `ci.yml` name `go mod tidy -diff`. |
| V4 | `git diff origin/dev --stat` | Only the files of the source table changed. |
| V5 | The five gate commands from the worktree root | `gofmt -l` and `go mod tidy -diff` print nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `fix(build): tidy the module
and add the tidy check to the gate`. Push the task branch. Write the evidence.
Run `hive task complete`. Send the commit SHA and the V1-V5 results to the
manager with `hive inbox send`. Read only your own inbox.

## Scope change 1

V3 in the table above is too wide. The brief copy and the friction entry
contain the words of the check. Use this V3:

| ID | Check | Expected result |
| --- | --- | --- |
| V3 | `git grep -n -e 'four gate' -e 'go mod tidy' -- AGENTS.md .agents/skills .hive/spec/EVIDENCE-FORMAT.md .github/workflows/ci.yml` | No line says "four gate". The gate in `AGENTS.md` and the worker skill, the verifier skill, `EVIDENCE-FORMAT.md`, and `ci.yml` name `go mod tidy -diff`. |

Do not edit other files for V3. Continue with V4 and V5.
