# T11-task-identity-and-finish fix 03

| Item | Value |
| --- | --- |
| Task | T11-task-identity-and-finish |
| Pass | t11-fix-03 |
| Date | 2026-10-18 |
| Agent | worker, external Pi, opencode-go/deepseek-v4-flash:medium |
| Requested route | `opencode-go`, `deepseek-v4-flash`, reasoning `medium` |
| Effective route | `opencode-go`, `deepseek-v4-flash`, reasoning `high` |
| Session | `01a14ec7-4c7f-72a9-bb8b-e83e67861a79` |
| Activation base | `eded8f9cb09208bfdc81334870e86479e746c183` |
| Start tip | `065521261194cc5b9c13f62c0bfd1e49c0b4f5cf` |
| Start parent | `59b2b4125562bf075e50ab5d020e34d7310dab04` |
| Test commit | `7b8e009489a4156bf7ddf0641fa2c1833e8b9ec6` |
| Source commit | `395156d1dcaa27c80da7c5f078f18c5f73765bba` |
| Branch | `task/T11-task-identity-and-finish` |
| Result | The fix closes T11-IO-F06. Source acceptance stays open for a fresh verifier. |

The worktree was clean at the start tip. Local, tracking, and direct remote tips were equal. The start tip is the verifier verdict commit. Its sole parent is worker `59b2b4125562bf075e50ab5d020e34d7310dab04`. The activation base stays the source base. No merge from the root or from source08/09 ran.

The received brief is `.hive/tmp/worker/t11-fix-03-captures-1792323124-1339246/brief.txt`. Its SHA256 is `9fa9d22e53806127a0a4387d3add8116945caeae104cc200ef78eddaed9ae696`. The capture root also holds `route.txt`, `start.txt`, `remote-tip.txt`, the raw outputs, the exact probes, and the result files.

## Finding

`common.Run` resolves the home before it calls the command callback. The complete and verify wrappers set `homeUsed` only inside that callback. A failed JSON stdout before the callback still set `tracker.err`. The wrappers then called `LogOutputFailure` with an empty home. The task logger made a relative `.hive/logs` under the current directory. The wrappers also printed a failure report from the zero envelope.

The fix adds one condition to the owned follow-up in each wrapper: `homeUsed != "" && tracker.err != nil`. A resolved home sets `homeUsed`. If the callback never runs, `homeUsed` stays empty and the wrapper returns the original `common.Run` result. The wrapper starts no fallback logger, runs no second home resolution, invents no effect or phase, and prints no empty finish envelope.

| Path | Change |
| --- | --- |
| `cmd/task/complete/complete.go` | Guard the output failure follow-up with `homeUsed`. |
| `cmd/task/verify/verify.go` | Guard the output failure follow-up with `homeUsed`. |
| `cmd/task_finish_root_test.go` | Add the focused home guard tests. |

No identity, inbox, store, logger, common, output, dependency, baseline, task, tracker, root, old evidence, report, or `VERIFIED.md` file changes. `internal/task/transition.go` and the accepted unconditional resolved agent plumbing stay exact. The fix adds no framework, API, persisted field, replay, or rollback.

A search of all production Go files finds two `LogOutputFailure` call sites. Each command call is in the `homeUsed != "" && tracker.err != nil` branch. The callback sets `homeUsed` before each call. The capture `log-output-failure-sites.txt` keeps the search and both source regions.

## Test-first and controls

| Check | Before the source fix | After the source fix |
| --- | --- | --- |
| Invalid-home failing JSON output, 12 subcases | Red. The CWD `.hive` root existed. | Green. |
| Invalid-home working JSON output, six subcases | Green. | Green. |
| Empty explicit home with a failing writer | Green. | Green. |
| Valid-home short JSON output, three operations | Green. | Green. |

The new test `TestTaskFinishOutputFailureBeforeHomeResolutionCreatesNothing` covers complete, verify pass, and verify fail. It uses an actual error writer and an actual short writer. It covers a missing home and an invalid explicit home. Every case asserts exit 1, one JSON envelope attempt, no `.hive` root in the current directory or the invalid home, and no finish report. The red run failed on the `.hive` assertion, not on a compile error. The raw output is `red-focused-before-source.out`. The green output is `green-focused-after-source.out`.

`TestTaskFinishHomeFailureWithWorkingOutputKeepsFailedEnvelope` keeps the ordinary failed/home JSON envelope with an empty effect list. `TestTaskFinishEmptyHomeFlagWithFailingStdoutKeepsUsageStatus` keeps status 2 and one usage line. `TestTaskFinishValidHomeOutputFailureLogsResolvedActor` keeps one stream-error event with the actual resolved agent for complete, verify pass, and verify fail.

A focused mutation removed the guard from `cmd/task/complete/complete.go` only. The complete subcases failed on the `.hive` assertion. The verify subcases stayed green. The restored source hash `c54a34daa046621944c51629cf45c8f28fbd057807c84e2074d6b901d4b81cfd` equals the source before the mutation. The outputs are `red-mutation-broken-guard.out` and `green-restored-mutation.out`. The exact patch is `mutation-broken-guard.patch`.

## Actual CLI reproduction

The fix script `home_guard_fix.py` uses 36 cases. It covers the actual binary, the exact real Cobra output driver with an error writer, and the same driver with a short writer. It covers complete, verify pass, and verify fail. It covers a missing home and an invalid explicit home. It covers a failing stdout and a working stdout.

Every case exits 1. No case creates a `.hive` root, state directory, log file, or marker in the current directory or the invalid home. The failing writer cases use real `/dev/full`, the actual writer error, or the actual short writer. No failing case prints a finish report. Each JSON writer gets one envelope attempt. The working-output controls keep one failed/home envelope with an empty effect list.

The fixtures are below `/tmp/t11f3c-ww0vdom9/outside`. They are outside every marked home. `HIVE_HOME` and `HERDR_PANE_ID` are unset in each child. Each binary has an absolute path and a recorded hash. The result file is `home-guard-fix-results.json`. `home-guard-fix-provenance.json` records the source repro, the source result, the driver hash, the script hash, and each expectation change. The driver copy matches the verifier driver byte-exact. No probe ran against the actual root or another existing home.

## Old checks

The exact retained verifier probes are byte copies from the verifier capture. `probe-copies.sha256` records the eight hashes. `green-independent-probes.out` shows every probe passes with race checks and count 1. The logging matrix passes all 72 cases. The run covers the six direct logging phases and four routes. It also covers the default file logger, static types, registration order, held APIs, and output outside the lock. The run removes the copies.

The full test suite keeps the 12 real storage prefixes, the ASCII and UTF-8 caps, and the same-reason and changed-reason retries. It keeps the finished assignment refusal, the fresh worker rework, and the cancel behavior. It keeps the held-handle checks and the no-lock output checks. No pass removes or skips a test. No pass weakens an assertion.

## Gate

The gate uses short `TMPDIR=/tmp/t11f3g`, outside every marked home. The worktree is at source commit `395156d1`.

| Command | Result |
| --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | No output |
| `go vet ./...` | Exit 0 |
| `golangci-lint run --allow-serial-runners ./...` | `0 issues.` |
| `go test -race -count=1 ./...` | Exit 0 |
| `go test -count=1 -run TestSTE .` | Exit 0 |

The baseline, the ratchet source, `go.mod`, and `go.sum` equal the activation base. `new-diff-check.out` gives a clean check from the start tip to the source commit. `historical-diff-check.out` keeps only the old V3 literal trailing spaces. The new evidence text has zero hard findings with the base and project glossaries.

## Limits

Hostile replacement between `Lstat` and `Open` stays deferred to T01. Source acceptance stays open for a fresh verifier. The effective reasoning level was `high`, not the requested `medium`. The provider and model match the request. No native agent, model child, root record, shared record, or actual-home domain write ran.

END T11 FIX 03 EVIDENCE
