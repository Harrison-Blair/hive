# 14-friction-folder VERIFIED

| Item | Value |
| --- | --- |
| Task | 14-friction-folder |
| Change commit | 077ba655 |
| Branch head | ac69e1b7 |
| Date | 2026-10-19 |
| Agent | t14-verifier-1, pi, opencode-go/deepseek-v4.1-flash |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | R1 | pass | The folder has `FRICTION.md`, `_template.md`, 8 files in `entries/`, and 4 files in `entries/resolved/`. Git tracks each file. `reference/dogfood/friction.md` is absent and untracked. |
| V2 | R1, R2 | pass | Each of the 11 moved entries matches its base text exactly (whitespace-normalized). The merged entry has the full text and reproduction steps of both source entries, with the `agent_not_ready` part marked. One source entry became two entries; no text is lost. |
| V3 | R1, R2 | pass | All 12 rows agree with their frontmatter on title, status, and decision, and each agrees with R2. The open rows come first and sort by `opened` newest first, and the resolved rows follow. |
| V4 | R3 | pass | Each of the five breaks fails `go test -run Friction .` with a message that names the break. The restored tree passes and `git status` is clean. |
| V5 | R4 | pass | `grep -rn 'friction.md' AGENTS.md .agents/skills internal/agent` finds no match. All six R4 files name the folder and give the resolve steps (set the status, move the file to `entries/resolved/`, update the row). |
| V6 | R4, R5, R6 | pass | The manager prompt test passes and is red against the old template. All STE and Gitignore tests pass. The new task copy has baseline 0, and no baseline count rose. The literal brief filter `TestSTE\|Gitignore` matches no test under Go regex syntax, so I also ran the explicit test names. |
| V7 | Gate | pass | `gofmt -l` prints nothing. `go vet`, `golangci-lint`, and `go test -race` exit 0. |

## V1 Folder contents

```sh
find reference/dogfood/friction -type f -print | sort
git ls-files reference/dogfood
```

Output: `FRICTION.md`, `_template.md`, `entries/` with 8 files
(`agent-spawn-skips-first-prompt`, `any-agent-read-another-role-inbox`,
`exited-agent-name-blocks-reuse`, `idle-session-supervisor-nudges`,
`no-hive-path-repairs-verified-task`, `task-edit-brief-command`,
`verifier-spawn-pane-requires-project`, `worker-spawn-keeps-old-acting-agent`),
and `entries/resolved/` with 4 files (`full-pi-bootstrap-flow`,
`manager-spawn-shell-ready-race`, `shell-ready-spawn-race`,
`verify-fail-no-reverify-workspace`). The two lists match and `git ls-files`
lists each file. `test -e reference/dogfood/friction.md` reports absent, and
`git ls-files --error-unmatch reference/dogfood/friction.md` fails with
`did not match any file(s) known to git`.

## V2 Entry text

I parsed the 13 entries of `git show 72a33b5b:reference/dogfood/friction.md`
and each of the 12 entry files. All 11 entries that were not merged match
their base summary and steps exactly after whitespace normalization. For the
merge, I split the entry at the two marker lines:

```text
**Merged text from the `agent_not_ready` entry:**
**Reproduction steps from the `agent_not_ready` entry:**
```

Each part equals the matching base text. 12 files cover the 13 base entries:
one merge and no loss.

## V3 Frontmatter and order

A script compared each row with the frontmatter of the linked file. All 12
rows match on title, status, and decision, and each status and decision equals
the R2 table. The order is 8 `Open` rows, then 4 `Resolved 2026-10-19` rows.
The open `opened` values are 2026-10-19 (seven) and 2026-10-18 (one), in
descending order.

Observation: the entry `full-pi-bootstrap-flow.md` has `opened: 2026-10-19`.
`git log -S` on the old file gives the first addition as a01167cd on
2026-10-18, while `git log --diff-filter=A` on the new file gives the
migration commit 077ba655 on 2026-10-19. The brief lists both commands, so
this does not fail V3. The manager can decide if one date is preferred.

## V4 Five breaks

I ran the baseline (`ok`), made each break in the working tree, ran
`go test -run Friction .`, and restored the tree. The outputs:

```text
baseline: ok
break 1 (row with no file):  table row has no entry file: entries/missing-entry.md
break 2 (file with no row):  entry file has no table row: entries/extra-entry.md
break 3a (title mismatch):   title mismatch for entries/task-edit-brief-command.md: table has "No `hive task edit --brief` command", frontmatter has "A changed title"
break 3b (status mismatch):  status mismatch for entries/task-edit-brief-command.md: table has "Open", frontmatter has "Resolved 2026-10-19"
break 4 (open in resolved):  open entry is in entries/resolved/: entries/resolved/task-edit-brief-command.md
break 5 (resolved in open):  resolved entry is in entries/: entries/shell-ready-spawn-race.md
restored: ok
```

`git status --porcelain` was empty after each restore. Breaks 4 and 5 also
updated the table link so the case under test was the only failure.

## V5 Direct-edit rule

```sh
grep -rn 'friction.md' AGENTS.md .agents/skills internal/agent
```

No match, exit 1. The folder rule and the resolve steps are present in both
`AGENTS.md` sections (`Coordinator`, `Dogfooding`), both skills
(`hive-coordinator`, `hive-manager`), the manager template
`internal/agent/templates/project-agents.md`, and are asserted in
`internal/agent/manager_prompt_test.go`.

## V6 Tests and STE

```sh
go test ./internal/agent/ -run Manager       -> ok
go test -run 'TestSTE\|Gitignore' .          -> ok, [no tests to run]
go test -run 'TestSTE|TestTrackedFilesAreNotIgnored|TestGitHubWorkflowIsNotIgnored'
go test -run 'TestHomeMarkerIsTracked|TestPersonalPathsAreIgnored|TestSourcePathsUnderIgnoredNames' .
```

The literal brief filter uses `\|`, which Go reads as a literal `|`, so it
selects no test. I ran the STE and Gitignore tests by their explicit names:
all 15 tests pass. The manager prompt test goes red against the old template
(`manager instructions do not name the home friction folder exception`) and
green with the new template. `testdata/ste-baseline.json` adds only
`spec/hive-v1/tasks/14-friction-folder.md: 0`; no count rose.

The brief copy differs from the task brief only at the three sentences that
the evidence records as STE changes, and each keeps its meaning.

## V7 Gate

```text
gofmt -l $(git ls-files -co --exclude-standard '*.go')   no output, rc 0
go vet ./...                                             rc 0
golangci-lint run ./...                                  0 issues, rc 0
go test -race ./...                                      ok for each package, rc 0
```

The sequential run passed on the first attempt; I saw no flake.

## Result

All seven items pass. The folder layout, the entry text, the R2 decisions,
the sync test's five red cases, the direct-edit rule, the STE ratchet, and the
gate agree with the brief.
