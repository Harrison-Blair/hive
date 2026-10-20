# VERIFIED: 36-nudge-no-home

| Item | Value |
| --- | --- |
| Task | 36-nudge-no-home |
| Commit verified | 76fb4337 |
| Date | 2026-10-19 |
| Verifier | Claude Opus 5.5 (claude-opus-5-5), Claude Code |

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | Task `nudge_test.go` copied onto a detached worktree at base `a59eb6d5`: `TestNudgePluralInboxPrompt` and `TestNudgeCompletedWithQuietPromptsOnce` FAIL with `run: hive --home /tmp/... inbox`. On the task branch both PASS. Both tests assert literal strings (singular and plural). |
| V2 | pass | `go build -o ./hive .` in the worktree; `env -u HIVE_HOME ./hive task show 36-nudge-no-home` exit 0, shows the task as `completed` with worktree path under `/home/penguin/source/hive`. `env -u HIVE_HOME ./hive task list --project hive` exit 0, lists home tasks through `36-nudge-no-home completed`. No inbox read. Binary removed. |
| V3 | pass | `git diff dev --stat`: `internal/supervise/nudge.go`, `internal/supervise/nudge_test.go`, `spec/hive-v1/tasks/36-nudge-no-home.md`, `spec/hive-v1/tasks/T15-supervise.md`, `testdata/ste-baseline.json` (adds brief with 0), and `evidence/36-nudge-no-home/V1-V4.md`. All in the source table or the evidence folder. `grep -rn 'hive --home' internal/supervise` prints nothing (exit 1). No other doc, skill, or spec outside the excluded folders quotes the old nudge. The brief copy rewords one Known-facts sentence ("returns a home-specific inbox command"); recorded in V3 evidence. |
| V4 | pass | `gofmt -l` prints nothing; `go vet ./...` exit 0; `golangci-lint run ./...` 0 issues, exit 0; `go test -race ./...` exit 0, all packages ok. |

## Run 2

| Item | Value |
| --- | --- |
| Commit verified | 5f3b131c |
| Date | 2026-10-19 |
| Verifier | t36-verifier-2, Claude Opus 5.5 (claude-opus-5-5), Claude Code |

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | Task `nudge_test.go` copied onto a detached worktree at base `a59eb6d5`: `TestNudgePluralInboxPrompt` FAIL with `run: hive --home /tmp/... inbox` (plural), `TestNudgeCompletedWithQuietPromptsOnce` FAIL with the same text (singular). On the task branch both PASS. Both tests assert literal strings. Scratch worktree removed. |
| V2 | pass | `go build -o ./hive .` in the worktree. `env -u HIVE_HOME ./hive task show 36-nudge-no-home` exit 0, shows `status: completed` and the worktree path below `/home/penguin/source/hive`. `env -u HIVE_HOME ./hive task list --project hive` exit 0, lists the home tasks through `36-nudge-no-home completed`. No inbox read. Binary removed. |
| V3 | pass | `git diff dev --stat` also shows `reference/dogfood/friction/` files, because `dev` moved to `89af1aeb` after the base. `git diff $(git merge-base HEAD dev) --stat` shows only `internal/supervise/nudge.go`, `internal/supervise/nudge_test.go`, `spec/hive-v1/tasks/36-nudge-no-home.md`, `spec/hive-v1/tasks/T15-supervise.md`, `testdata/ste-baseline.json`, and `evidence/36-nudge-no-home/`. `grep -rn 'hive --home' internal/supervise` prints nothing (exit 1). No doc outside the excluded folders quotes the old nudge text. |
| V4 | pass | `gofmt -l` prints nothing. `go vet ./...` exit 0. `golangci-lint run ./...` 0 issues, exit 0. `go test -race ./...` exit 0. |
| Repair 1 | pass | `git ls-tree --name-only HEAD` shows no `VERIFIED.md` at the checkout root. The file is at `evidence/36-nudge-no-home/VERIFIED.md`. |
