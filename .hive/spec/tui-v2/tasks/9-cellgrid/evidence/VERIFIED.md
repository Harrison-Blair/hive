# Verification: 9-cellgrid

- Verifier: `t9-verifier-2`.
- Date: 2026-10-21.
- Task commit: `dc520867`.
- Base: `dev` at `809bea84aed71476dfd8db092da1dfc0c95fd3de`.
- Verdict: **pass**.

This run checks the task again after the land conflict repair. The verifier read the brief and the spec. `git merge dev` printed `Already up to date.` No source change remains from verification.

| Item | Result | Observation |
| --- | --- | --- |
| V1 | pass | Each of the 16 new tests failed on the base, failed with inert declarations, and passed on the task code. A converter mutation failed the accent tests. The restored converter passed. |
| V2 | pass | The spec diff changes only `Manager decisions`. `go test ./ -run STE` exited 0. |
| V-gate | pass | The diff has only the named source files and evidence. All five gate commands exited 0. The first two printed nothing. |

## V1: independent red and green runs

The verifier made a detached scratch worktree from `dev` and copied only the four new test files into it. For each test below, the verifier ran:

```sh
go test -count=1 ./internal/lib/cellgrid -run '^TEST_NAME$'
```

The base has no cellgrid package implementation. Each base run exited 1 with undefined symbols. To check assertion failures too, the verifier added the inert declarations printed in V1.md to the scratch worktree. These declarations return empty values and no errors. Each test then failed at its assertions. Each test passed on the task code with the same command.

| Test | Base exit | Inert declarations exit | Task exit |
| --- | --- | --- | --- |
| TestCompareReportsEachDifferingCell | 1 | 1 | 0 |
| TestCompareMaskIgnoresRuneOnly | 1 | 1 | 0 |
| TestCompareRejectsBadMask | 1 | 1 | 0 |
| TestCompareRejectsSizeMismatch | 1 | 1 | 0 |
| TestReadMask | 1 | 1 | 0 |
| TestReadMaskErrorsNameTheLine | 1 | 1 | 0 |
| TestConvertBoxWithAccentTitle | 1 | 1 | 0 |
| TestConvertDefaultForegroundBoldOffResetAndPadding | 1 | 1 | 0 |
| TestConvertMapsEachPaletteHex | 1 | 1 | 0 |
| TestConvertFailures | 1 | 1 | 0 |
| TestReadGridFile | 1 | 1 | 0 |
| TestWriteGridFile | 1 | 1 | 0 |
| TestGridRoundTripEachToken | 1 | 1 | 0 |
| TestReadGridErrorsNameTheLine | 1 | 1 | 0 |
| TestWriteGridRejectsWideRune | 1 | 1 | 0 |
| TestPaletteMatchesDecision26 | 1 | 1 | 0 |

The inert runs found missing diffs, missing mask errors, empty masks, zeroed grids, missing converter errors, empty grid output, missing file errors, and empty palette values. These were assertion failures, not build failures.

### Converter mutation

The verifier changed the return in `hexToken` temporarily:

```go
if t == Accent { return Dim, nil }; return t, nil
```

Command:

```sh
go test -count=1 -v ./internal/lib/cellgrid
```

Result: exit 1. `TestConvertBoxWithAccentTitle`, `TestConvertDefaultForegroundBoldOffResetAndPadding`, and `TestConvertMapsEachPaletteHex` failed. The box test reported dim instead of accent at the four title cells. The verifier restored the exact source bytes and ran the same command. Result: exit 0; all 16 tests passed.

The expected box grid is hand-written. Its builder does not call Convert. The palette expectations give the spec hex values directly. Compare expectations give the differing cells directly.

### Mask detection and limits

`Compare` checks each expected cell that a mask touches. A rune in U+2500–U+257F fails as a border. Token `accent` fails as protected text. The tests prove rejection of left and right borders and an accent title. Masked rune changes pass; token and bold changes still produce diffs.

The check finds a title, label, or key hint only through its accent token or a box-drawing rune. It does not infer the meaning of words. A dim label or dim key hint has no semantic protection. A border outside the box-drawing block has no rune protection. These are the limits of the AC3 check specified in this brief; this verdict does not claim semantic mask validation. Future mask use must comply with decision 45 and obtain user approval.

## V2: spec

Commands:

```sh
git diff dev...HEAD -- .hive/spec/tui-v2/README.md
go test ./ -run STE
```

The diff adds 11 lines in `Manager decisions`: package path, grid format, token letters, mask format, and fixture path. No numbered decision changes.

STE output:

```text
ok  	github.com/Harrison-Blair/hive	(cached)
```

Exit: 0.

## V-gate: diff scope

`git merge dev` exited 0 and printed `Already up to date.`

`git diff dev...HEAD --stat` showed these paths:

```text
 .hive/spec/tui-v2/README.md
 .hive/spec/tui-v2/tasks/9-cellgrid/brief.md
 .hive/spec/tui-v2/tasks/9-cellgrid/evidence/V-gate.md
 .hive/spec/tui-v2/tasks/9-cellgrid/evidence/V1.md
 .hive/spec/tui-v2/tasks/9-cellgrid/evidence/V2.md
 .hive/spec/tui-v2/tasks/9-cellgrid/evidence/VERIFIED.md
 internal/lib/cellgrid/compare.go
 internal/lib/cellgrid/compare_test.go
 internal/lib/cellgrid/convert.go
 internal/lib/cellgrid/convert_test.go
 internal/lib/cellgrid/doc.go
 internal/lib/cellgrid/grid.go
 internal/lib/cellgrid/grid_test.go
 internal/lib/cellgrid/palette.go
 internal/lib/cellgrid/palette_test.go
 testdata/ste-baseline.json
```

Each path is in the source table or the evidence folder. This task adds a library and changes no UI renderer. Screen captures and a board comparison do not apply.

## Full gate output

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
Exit: 0

$ go mod tidy -diff
Exit: 0

$ go vet ./...
Exit: 0

$ golangci-lint run ./...
0 issues.
Exit: 0

$ go test -race ./...
ok  	github.com/Harrison-Blair/hive	13.241s
ok  	github.com/Harrison-Blair/hive/cmd	2.977s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.135s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.337s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/spawn	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.115s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.795s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.787s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.284s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.077s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.108s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.017s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	9.331s
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.709s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	(cached)
ok  	github.com/Harrison-Blair/hive/internal/homeinit	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/land	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.251s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/log	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/process	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/records	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/run	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/store	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/logs	(cached)
ok  	github.com/Harrison-Blair/hive/internal/memory	(cached)
ok  	github.com/Harrison-Blair/hive/internal/project	(cached)
ok  	github.com/Harrison-Blair/hive/internal/supervise	(cached)
ok  	github.com/Harrison-Blair/hive/internal/task	(cached)
ok  	github.com/Harrison-Blair/hive/internal/tui	49.069s
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
Exit: 0
```

## Cleanup

The verifier removed the detached scratch worktree and its inert declarations. The converter source was restored before the green run and gate. Temporary verification logs were removed after this report was recorded. No pane, tab, or workspace was created.
