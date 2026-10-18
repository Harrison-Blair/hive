# T11-task-identity-and-finish fix 01

| Item | Value |
| --- | --- |
| Task | T11-task-identity-and-finish |
| Pass | t11-fix-01 |
| Date | 2026-10-18 |
| Agent | worker, external Pi, opencode-go/deepseek-v4-flash:medium |
| Requested route | `opencode-go`, `deepseek-v4-flash`, reasoning `medium` |
| Effective route | `opencode-go`, `deepseek-v4-flash`, reasoning `high` |
| Session | `01a14e94-4033-7271-881c-263602db9d6f` |
| Activation base | `eded8f9cb09208bfdc81334870e86479e746c183` |
| First-review tip | `4d052b28fb71e486bf2c816260bf3cb3cd347b41` |
| Test commit | `6dce8b53` |
| Source commit | `44235d66` |
| Branch | `task/T11-task-identity-and-finish` |
| Result | This pass closes all open fix findings. Source acceptance stays open for a fresh verifier. |

The worktree was clean at `4d052b28`. Local, tracking, and direct remote tips were equal. The verdict commit is the sole evidence child of worker `129b82f2`. No merge from the current root ran. The activation base above stays the source base.

## Findings and fixes

| ID | Correction |
| --- | --- |
| P-F01 | The selected claim must agree with the holder record key, role, project, and generation. A valid claim of another role does not make the selected alias valid. |
| T11-IO-F04 | The agent and claim loaders keep each selected file name. A record whose file name differs from its embedded key does not resolve and cannot consume. The store still checks the static file types. |
| P-F02 | One resolved caller record drives the identity check, the transition check, the notification name, and the finished mark. Stored acting attribution does not change. |
| P-F03 | A finish refusal on an advanced task adds the truthful `task show`, `agent show`, and `inbox` commands. It writes nothing and does not use the status or reason as a key. |
| T11-IO-F01 | `inbox send` checks the explicit append values before `store.Open`. It checks the resolved sender again under the held handle before the append. |
| T11-IO-F02 | A default inbox refusal writes one error event with no role key. It writes no message body and no fabricated role. |
| T11-IO-F03 | A finish logs each completed effect with the resolved actor ID. A partial log holds the exact retained prefix. The command logs a stream error after a durable group and a failed stdout. Message and reason text stay out of the log. |
| T11-IO-F05 | The failure report uses the current operation actor from the resolved result. It does not use an old verification record. |
| T-W01 | The diagnostic test again checks a control-bearing record name on the resolved wrong-actor path. The task-byte preservation assertion stays. |

T11-IO-F04 is one finding. Its identity and consume effects are in the same correction.

## File map

| Path | Change |
| --- | --- |
| `internal/lib/identity/identity.go` | Selected-claim agreement, selected file-name identity, and loader changes. |
| `internal/task/transition.go` | One resolved actor, advanced refusal inspection, finish effect logs, stream error log, and the resolved actor result field. |
| `cmd/task/complete/complete.go`, `cmd/task/verify/verify.go` | Pass the resolved actor to the report and log a stream error. |
| `internal/inbox/send.go` | Explicit prevalidation before store access and resolved-sender validation under the handle. |
| `internal/inbox/read.go`, `internal/lib/inbox/inbox.go` | Log the default refusal and omit an absent role key. |
| Tests | `internal/lib/identity/identity_test.go`, `internal/inbox/read_send_test.go`, `internal/task/finish_test.go`, `internal/task/transition_test.go`, `internal/task/diagnostic_test.go`, `cmd/task_finish_root_test.go`. |

No store, common renderer, output package, logger, dependency, baseline, tracker, task brief, ACTIVATION file, or VERIFIED file changed. No finish sequence, finish key, journal, rollback, or replay mechanism exists.

## Red and green evidence

Captures are in `.hive/tmp/worker/t11-fix-01-captures-1792319830-988476/`. These paths are relative to that directory.

| Finding | Red | Green |
| --- | --- | --- |
| P-F01, T11-IO-F04 identity | `red-identity.out` | `green-identity.out`, `green-home-alias.out` |
| T11-IO-F01, F02 inbox | `red-inbox.out` | `green-inbox.out` |
| P-F02, P-F03 finish | `red-task-alias-and-retry.out` | `green-alias-and-retry.out` |
| T11-IO-F03, F05 command | `red-cmd-log-f05.out` | `green-cmd-log-f05-2.out` |
| T-W01 quoting | `red-diag-mutation.out` and `mutation-diagnostic.patch` | `restored-green-diag.out` |

Each red run failed on a behavior assertion. No red run failed on a compile error. The diagnostic mutation removed the quote of the resolved actor name and changed one assertion only. I restored the source before the gate.

The first-review probes are exact copies with provenance in `first-review-probes/`. `SHA256SUMS` gives their hashes. `replay-green.out` shows every `TestIndependent` probe passes. I removed the copies in the package directories after the run. `replay-copies.sha256` records them.

The exact first-review CLI probe is in `exact-cli/cli.py`. It uses a unique marked home and an absolute binary. `exact-cli/cli-run.out` reports these results. `required-v3` passes. All twelve prefix cases pass. `normal-retry-guard-fresh` passes. `cross-claim-expected-refusal` is true. `assigned-key-expected-success` is true. `exact-cli/binary.sha256` gives the binary hash.

The owned CLI probe is in `fix-cli-4/cli_fix_extra.py`. It uses a unique marked home and `/dev/full` for real stdout failures. `fix-cli-4/run.out` reports pass for both path mismatches and no-consume. It reports pass for invalid send and the unknown inbox log. It reports pass for finish effect logs, reason privacy, and the partial log prefix. It reports pass for the rework actor and the stream error. It reports pass for text output, JSON output, all three advanced retries, and the four literal reason bounds.

## Old checks

`gate/race.out` records `go test -race -count=1 ./...`. It exits 0. This run covers the real prefix failure matrix and the same-reason and changed-reason failed-verify retries. It covers the cancel behavior, the T03 transition table, and the T04 inbox safeguards. It covers the concurrent finish attempts and the cooperating reader checks.
The reason cases keep 4090 ASCII bytes accepted, 4091 ASCII bytes refused, 2045 two-byte characters accepted, and 2046 refused.
The advanced transition table keeps the exact text for the non-advanced `created` case and requires the three inspection commands for each advanced refusal.

## Gate

The gate ran with `TMPDIR=/tmp/t11fix-gate-oDemYe`, outside every marked home.

| Command | Result |
| --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | No output |
| `go vet ./...` | Exit 0 |
| `golangci-lint run --allow-serial-runners ./...` | `0 issues.` |
| `go test -race -count=1 ./...` | Exit 0 |
| `go test -count=1 -run TestSTE .` | Exit 0, baseline unchanged |
| Base and project glossary lint | `0 violations (0 hard, baseline 0)` |

`gate/diff-check` uses `git diff --check` on the fix commits. It reports no whitespace error. The literal V3 trailing-space exception remains historical and is not in this diff.

## Limits

- Hostile replacement between `Lstat` and `Open` stays deferred to T01.
- The source acceptance and the appended `VERIFIED.md` stay open for a fresh verifier.
- The effective reasoning level was `high`, not the requested `medium`. The provider and model were the requested ones.
- No native agent, model child, root record, or workspace change ran.

END T11 FIX 01 EVIDENCE
