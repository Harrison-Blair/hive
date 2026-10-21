# Independent verification: 13-remove-bootstrap

Date: 2026-10-21. Verifier: `t13-verifier-1`.
Verdict: pass.

I ran each command from the task worktree root. I merged local `dev` into
`task/13-remove-bootstrap` before the checks. The merge had no conflict.
Changes from `dev` include the completed task 7 spec copy. Those changes
are not part of this task's diff against `dev`.

| Item | Result | Observation |
| --- | --- | --- |
| V1 Search | pass | The exact search below prints no matches and exits 1. The clause, its pointer, and the developer role are absent from the searched instruction files and glossary. |
| V2 Tracker | pass | The diff adds one dated receipt of 11 lines. It names 2026-10-21, the user decision, the manager's worker capability, and this task. No old receipt changes. |
| V3 Ratchet | pass | The independent red run fails on the brief copy at baseline 0. The restored copy passes. The baseline diff adds only the task brief row at 0. No count rises. |
| V4 Diff scope and gate | pass | The diff has only approved source files and task evidence. Each of the five gate commands exits 0. |

## Commands and results

### V1

```sh
git grep -n -i -w -e bootstrap -e develop -e developer -- AGENTS.md STE-GLOSSARY.md .agents/
```

No output. Exit 1 means no match. I also read the instruction and glossary
diffs. Development applies to a worker. The worker Git rule stays intact.
No skill file needs a change.

### V2

```sh
git diff dev...HEAD -- .hive/spec/hive-v1/05-tracker.md
```

Only `### Dated removal: Bootstrap clause` and its receipt are new.
The diff has no deletion. The v1 design record stays unchanged:

- `.hive/spec/hive-v1/03-architecture.md:107` names the historical clause.
- `.hive/spec/hive-v1/04-risks.md:10` names the historical clause.

These references are history, as the brief requires.

### V3

For the red run, I temporarily restored the original passive wording in the
brief copy: `Expected: only the new receipt is added.` I ran the exact test
command, then restored the committed bytes in a `finally` block.

```text
$ go test ./ -run 'STE|Skill'
--- FAIL: TestSTERatchet
brief.md:98 passive-voice: is added
1 violations, baseline 0
exit 1
```

The green run used the restored committed copy:

```text
$ go test ./ -run 'STE|Skill'
ok github.com/Harrison-Blair/hive 0.688s
exit 0
$ git diff dev...HEAD -- testdata/ste-baseline.json
```

The diff adds only
`".hive/spec/hive-v1/tasks/13-remove-bootstrap/brief.md": 0`.
The glob in `steFiles` already covers this brief.

### V4

```sh
git diff dev...HEAD --stat
```

Before this report, the diff lists nine files: `AGENTS.md`,
`STE-GLOSSARY.md`, `testdata/ste-baseline.json`, the tracker, the brief copy,
and `evidence/V1.md` through `evidence/V4.md`. This report adds one evidence
file. No UI code or board changes belong to this task. No visual check applies.

| Gate command | Output | Exit |
| --- | --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | none | 0 |
| `go mod tidy -diff` | none | 0 |
| `go vet ./...` | none | 0 |
| `golangci-lint run ./...` | `0 issues.` | 0 |
| `go test -race ./...` | Each package passes or reports no test files. | 0 |

The temporary red-run edit is gone. The worktree was clean before this report.
No scratch files or live resources remain from this verification.
