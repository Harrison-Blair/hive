# T11-task-identity-and-finish fix 02

| Item | Value |
| --- | --- |
| Task | T11-task-identity-and-finish |
| Pass | t11-fix-02 |
| Date | 2026-10-18 |
| Agent | worker, external Pi, opencode-go/deepseek-v4-flash:medium |
| Requested route | `opencode-go`, `deepseek-v4-flash`, reasoning `medium` |
| Effective route | `opencode-go`, `deepseek-v4-flash`, reasoning `high` |
| Session | `01a14eac-7787-74e2-bb7c-f63e0fe47e2b` |
| Activation base | `eded8f9cb09208bfdc81334870e86479e746c183` |
| Start tip | `18afc0edd1b016e14e1e8819d68f9e12dc6e2304` |
| Start parent | `0124ab636ffa62dfcb41256801154970528fdb01` |
| Test commit | `9d4ecb234381f21d0ad66c850578ee738088f027` |
| Source commit | `94ebe8dba0e814d4779df4aa751ee88d32bf5119` |
| Branch | `task/T11-task-identity-and-finish` |
| Result | The fix closes T11-IO-F03 direct-library actor logging. Source acceptance stays open for a fresh verifier. |

The worktree was clean at the start tip. Local, tracking, and direct remote tips were equal. The start commit is the sole child of the fix worker commit. No merge from the root or from source08/09 ran. The activation base stays the source base.

The received brief is `.hive/tmp/worker/t11-fix-02-captures-1792321366-1238026554/brief.txt`. Its SHA256 is `aaa7a8a4ebce50a82201d607b5f3e5ae8f1468231007333a9552ef5584e16265`. The capture root also holds `route.txt`, `start.txt`, `probe-hashes.sha256`, and the raw outputs below.

## Finding

`runFinish` read the resolved caller only from the optional `TransitionOptions.ResolvedActor` output pointer. A direct library call without that pointer lost the record key. A name route logged the supplied name. A pane or history route logged no `agent_id`.

The fix keeps one fresh `records.Agent` in `runFinish`. The finish group writes that record. The logs use its key. The optional caller output receives the same record after the group. A nil pointer is no longer necessary. The fix does not infer identity from an old pointer, an old check, or a prior call.

The change is in `internal/task/transition.go` only. The test is `internal/task/finish_log_identity_test.go`. No persisted field, JSON envelope field, public API, framework, journal, rollback, or replay exists.

## Red and green

| Check | Before | After |
| --- | --- | --- |
| Exact verifier logging probe | 45 failed identity subcases, 27 passed | All 72 subcases passed |
| Exact verifier default-log probe | One failure, five events omit `agent_id` | Passed, five events carry `worker-key` |
| New focused test | Red on assertions | Passed |
| New focused test, before the source fix | 54 failed subcases | Not applicable |
| Focused mutation | Red on four top-level tests | Restored source passed |

The exact verifier probes are copies of `logging_probe_test.go`, `default_log_probe_test.go`, and `task_probe_test.go` from `t11-verify-02-captures/probes/`. `probe-provenance.txt` records a byte match for each copy. The before run is `red-verifier-logging-before.out` and `red-verifier-default-logger-before.out`. The after run is `green-verifier-logging-after.out` and `green-verifier-default-logger-after.out`. The after run also kept the actual logs in `library-logs-after/`.

The new test covers three operations and six phases on four routes. The operations are complete, verify pass, and verify fail. The phases are success, task, counter, record, finished, and refusal. The routes are key, name, pane, and history. A real read-only state directory causes each failure. The test mocks no internal logging or finish behavior. The test checks the actual effect prefix, phase, task and project IDs, controlling role, reason privacy, and actor key.

`red-new-test-before-source.out` shows `agent_id=worker-name` and `agent_id=<nil>`. `green-new-test-after-source.out` shows a pass for each focused test. The first default-log sample after the fix has one transition and four effects. Each event has `"agent_id":"worker-key"`. The sample is `library-logs-after/default-logger-complete-pane.jsonl`.

`mutation-remove-resolved-key.patch` removes the resolved key from the log path. `red-mutation-resolved-key.out` shows four assertion-red tests. `mutation-restore.sha256` shows the restored source hash equals the fixed source hash. `restored-green-mutation.out` shows the pass.

## Old checks

The exact first-review probes pass in the six named packages with race checks and count 1. The run is `green-all-independent-probes-after.out`. It covers the closed identity, path, send order, refusal log, retry, rework, output, stream, and diagnostic findings. The first run used a long probe `TMPDIR`. The Unix socket probe failed because the socket path was too long. A rerun with short `TMPDIR=/tmp/t11f2p` passed. `red-long-tmpdir-socket.out` reproduces the long path condition.

The existing `internal/task` and `cmd` tests pass in `green-existing-task-cmd.out`. These tests keep the 12 real failure-prefix cases, the ASCII and UTF-8 cap, the retries, and the fresh rework. They also keep the alias and path checks, the output and stream logs, and the held-handle and concurrency safeguards. I removed, skipped, or weakened no test.

## Gate

The gate ran with short `TMPDIR=/tmp/t11f2g`, outside every marked home. The worktree was at source commit `94ebe8db`.

| Command | Result |
| --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | No output |
| `go vet ./...` | Exit 0 |
| `golangci-lint run --allow-serial-runners ./...` | `0 issues.` |
| `go test -race -count=1 ./...` | Exit 0 |
| `go test -count=1 -run TestSTE .` | Exit 0 |

The baseline, the ratchet source, `go.mod`, and `go.sum` equal the activation base. `baseline-deps-exact.out` gives the hashes and the empty diff. `new-diff-check.out` shows a clean check from the start tip to the source commit. `historical-diff-check.out` keeps only the old V3 trailing whitespace exception. The evidence text has zero hard findings with the base and project glossaries.

## Limits

Hostile replacement between `Lstat` and `Open` stays deferred to T01. Source acceptance and the `VERIFIED.md` append stay open for a fresh verifier. The effective reasoning level was `high`, not the requested `medium`. The provider and model match the request. No native agent, model child, root record, shared record, or actual-home domain write ran.

END T11 FIX 02 EVIDENCE
