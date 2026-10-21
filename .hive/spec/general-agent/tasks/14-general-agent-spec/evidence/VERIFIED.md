# Independent verification: 14-general-agent-spec

| Item | Value |
| --- | --- |
| Date | 2026-10-21 |
| Agent | `t14-verifier-1` |
| Harness and model | pi, `openai-codex/gpt-6.1-sol` |
| Checked commit | `83239d36` |
| Verdict | pass |

## Items

| Item | Result | Observation |
| --- | --- | --- |
| V1 Spec copy | pass | The diff shows only the sentence change that V1.md records for AC3. The meaning stays the same. |
| V3 Ratchet | pass | The test passes, fails with the added violation, and passes after removal. The README and brief each have baseline 0. |
| V4 Diff scope | pass | `git merge dev` reports `Already up to date.` The diff contains only the source table files and evidence. |
| V5 Gate | pass | All five commands exit 0. Formatting and tidy print nothing. |

## V1

Command:

```sh
diff /home/penguin/source/hive/plans/general-agent/README.md .hive/spec/general-agent/README.md
```

Output, exit 1:

```text
50c50
<     1. It reads the record again and stops when `first_prompt_attempt_at` is set.
---
>     1. It reads the record again and stops when `first_prompt_attempt_at` has a value.
```

This is the sole difference. It changes no decision number, key, symbol, hex value, size, path, version, or quoted text.

## V3

I ran `go test ./ -run STE` three times:

1. Original README: exit 0, `ok github.com/Harrison-Blair/hive 0.693s`.
2. README with the sentence `The pane is closed; spawn stops.` at its end: exit 1.
3. README after removal: exit 0, `ok github.com/Harrison-Blair/hive 2.758s`.

The red run reports:

```text
--- FAIL: TestSTERatchet (1.07s)
    ste_test.go:571: .hive/spec/general-agent/README.md: 2 violations, baseline 0; fix the new ones
        .hive/spec/general-agent/README.md:128 semicolon: ;
        .hive/spec/general-agent/README.md:128 passive-voice: is closed
FAIL
FAIL github.com/Harrison-Blair/hive 1.071s
```

`git diff --exit-code -- .hive/spec/general-agent/README.md` then exits 0 with no output.

## V4

I ran `git merge dev`, then `git diff dev...HEAD --stat`.
The merge reports `Already up to date.` The diff lists these eight files:

```text
.hive/spec/general-agent/README.md
.hive/spec/general-agent/tasks/14-general-agent-spec/brief.md
.hive/spec/general-agent/tasks/14-general-agent-spec/evidence/V1.md
.hive/spec/general-agent/tasks/14-general-agent-spec/evidence/V3.md
.hive/spec/general-agent/tasks/14-general-agent-spec/evidence/V4.md
.hive/spec/general-agent/tasks/14-general-agent-spec/evidence/V5.md
ste_test.go
testdata/ste-baseline.json
```

The diff reports 490 insertions. The ratchet code change adds only the README path to `steFiles`.
The baseline change adds only the README and brief entries, each with 0.
This report adds one evidence file.

## V5

| Command | Exit | Output |
| --- | --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | 0 | none |
| `go mod tidy -diff` | 0 | none |
| `go vet ./...` | 0 | none |
| `golangci-lint run ./...` | 0 | `0 issues.` |
| `go test -race ./...` | 0 | Each package reports `ok` or `[no test files]`. No failure or race report. |

The race run includes `ok github.com/Harrison-Blair/hive 19.458s` and
`ok github.com/Harrison-Blair/hive/internal/tui 53.561s`.
Some unchanged packages report cached results.

## Scope and cleanup

This task changes no screen. A visual check does not apply.
The temporary violation is absent. No scratch file remains.
Before this report, `git status --short` prints nothing.
I did not change worker evidence or the plan tracker.
