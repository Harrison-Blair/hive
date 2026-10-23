# Task 132 independent verification

- Task: `132-tui-clock-tests`
- Verifier: `t132-verifier-1` (pi, `gpt-6.1-sol`)
- Date: 2026-10-23
- Task revision: `143d5bb23163252ce6cc175b6eb9dc434f2300cf`
- Base and local `dev`: `7d8f42db123e1da455f8dd0762b4e80d407441d6`
- Default zone: EDT, UTC-04:00 (`TZ` is unset).
- Go: `go1.27.2-X:nodwarf5 linux/amd64`
- Linter: `golangci-lint 2.15.0`
- Verdict: **pass**

All commands ran from the task worktree root. Each result below comes from a
new verifier run. No worker evidence file was changed.

## Results

| Item | Result | Observation |
| --- | --- | --- |
| V1 Reproduce the failure | pass | At the base commit, both named tests fail in UTC. All ten cases show `12:00` and expect `08:00`. Test exit: 1. |
| V2 Zone runs | pass | Both tests pass with UTC, America/New_York, Asia/Tokyo, and the default zone. Each command exits 0. |
| V3 Break runs | pass | Both tests fail for each of the count, word, and layout breaks. Each test command exits 1. The header file is restored after each break. The final green run exits 0. |
| V4 Diff | pass | The diff has only the two test files, the brief, evidence, and the explicitly permitted STE baseline entry. No product Go file changes. The worktree is clean before this report. |
| V5 UTC full suite | pass | `TZ=UTC go test -race ./...` exits 0. All packages pass or have no tests. |
| V6 Gate | pass | All five commands exit 0. Formatting and module checks print nothing. The linter reports `0 issues.`. |

## V1: base red run

The worktree was clean before the run. The verifier checked out the full base
commit with `git switch --detach 7d8f42db`, then ran the exact command:

```sh
TZ=UTC go test -race ./internal/tui -run 'TestHeaderShowsTheOpenAskRecords|TestHeaderShowsTheAskCountOfTheState' -count=1
```

Both tests failed. Representative output:

```text
--- FAIL: TestHeaderShowsTheOpenAskRecords (0.04s)
    ask_test.go:511: 0 open at 120: header = " hive  /tmp/TestHeaderShowsTheOpenAskRec… coordinator ○ stopped   supervisor ○ stopped   ? 0 asks   needs you 0   12:00 ", want suffix "   ? 0 asks   needs you 0   08:00 "
--- FAIL: TestHeaderShowsTheAskCountOfTheState (0.02s)
    chat_modes_test.go:349: asks 0 at 120: header = " hive  /tmp/TestHeaderShowsTheAskCountOf… coordinator ○ stopped   supervisor ○ stopped   ? 0 asks   needs you 0   12:00 ", want suffix "   ? 0 asks   needs you 0   08:00 "
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui	0.114s
FAIL
```

The test command exited 1. The verifier returned to
`task/132-tui-clock-tests` before V2. No base file was left in the task branch.

## V2: zone green runs

```sh
TZ=UTC go test -race ./internal/tui -run 'TestHeaderShowsTheOpenAskRecords|TestHeaderShowsTheAskCountOfTheState' -count=1
TZ=America/New_York go test -race ./internal/tui -run 'TestHeaderShowsTheOpenAskRecords|TestHeaderShowsTheAskCountOfTheState' -count=1
TZ=Asia/Tokyo go test -race ./internal/tui -run 'TestHeaderShowsTheOpenAskRecords|TestHeaderShowsTheAskCountOfTheState' -count=1
go test -race ./internal/tui -run 'TestHeaderShowsTheOpenAskRecords|TestHeaderShowsTheAskCountOfTheState' -count=1
```

| Zone | Output for `github.com/Harrison-Blair/hive/internal/tui` | Exit |
| --- | --- | --- |
| UTC | `ok`, 1.128s | 0 |
| America/New_York | `ok`, 1.126s | 0 |
| Asia/Tokyo | `ok`, 1.127s | 0 |
| Default, EDT | `ok`, 1.131s | 0 |

Each test injects 08:00 on 2026-10-20 in `time.Local`. The expected suffix is
still a literal. It includes the ask count, the ask word where present,
separators, spacing, and `08:00`. No shared helper changed.

## V3: mutation red runs and restored green run

Each break changed only `internal/tui/component/header.go`. After each edit,
the verifier ran this command without a pipe and recorded its own exit code:

```sh
go test -race ./internal/tui -run 'TestHeaderShowsTheOpenAskRecords|TestHeaderShowsTheAskCountOfTheState' -count=1
```

| Break | Temporary change | Observed failure | Exit |
| --- | --- | --- | --- |
| Count | Replace `h.Asks` with `h.Asks+1` in both count format calls. | Both tests fail at 120 and 80 columns. The zero-ask case shows `? 1` and expects `? 0`. | 1 |
| Word | Swap the long-form ` ask` and ` asks` literals. | Both tests fail at 120 columns. The zero-ask case shows `? 0 ask` and expects `? 0 asks`; the one-ask case shows `? 1 asks` and expects `? 1 ask`. | 1 |
| Layout | Use two spaces before long-form `needs you`; replace the short-form separator before `needs you` with ` | `. | Both tests fail at 120 and 80 columns. The clock remains `08:00`; the spacing and separator differ from the expected suffix. | 1 |

After each break:

```sh
git checkout -- internal/tui/component/header.go
git status --short
```

Each restore exited 0. Each status command printed nothing. After the last
restore, the same test command passed with `ok`, 1.111s, exit 0.

## V4: scope and product check

```sh
git diff dev --stat
git status --short
git diff dev --name-only
git diff dev -- ':(glob)**/*.go' ':(exclude,glob)**/*_test.go'
git diff dev -- internal/tui/component/header.go
git diff dev --check
```

The initial stat reports 10 files and 710 insertions:

- `internal/tui/ask_test.go`: one clock injection.
- `internal/tui/chat_modes_test.go`: one clock injection and the `time` import.
- `.hive/spec/tui-v2/tasks/132-tui-clock-tests/brief.md`.
- `.hive/spec/tui-v2/tasks/132-tui-clock-tests/evidence/V1.md` through `V6.md`.
- `testdata/ste-baseline.json`: the brief path with a count of 0.

The source-file table in the brief explicitly permits the baseline entry for
AC4. This permission takes priority over the shorter path list in AC3 and V4.
The entry changes no existing count.

The product Go diff, header diff, whitespace check, and status print nothing.
There is no change to the header output for a user. This report adds one
permitted evidence file. No screen changes, so no visual board check applies.

## V5: UTC full suite

```sh
TZ=UTC go test -race ./...
```

Exit: 0. All packages report `ok` or `[no test files]`. Some unchanged packages
use the Go test cache. Selected output:

```text
ok  github.com/Harrison-Blair/hive                         19.032s
ok  github.com/Harrison-Blair/hive/internal/tui             27.648s
ok  github.com/Harrison-Blair/hive/internal/tui/component   9.504s
```

## V6: complete gate

| Exact command | Output | Exit |
| --- | --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | None. | 0 |
| `go mod tidy -diff` | None. | 0 |
| `go vet ./...` | None. | 0 |
| `golangci-lint run ./...` | `0 issues.` | 0 |
| `go test -race ./...` | All packages report `ok` or `[no test files]`; some unchanged packages use the Go test cache. Root: 18.509s. TUI: 27.362s. | 0 |

## Cleanup

All temporary product edits are restored. No scratch files, panes, or
workspaces were created. Before this report, `git status --short` printed
nothing. Only this report is added by the verifier.
