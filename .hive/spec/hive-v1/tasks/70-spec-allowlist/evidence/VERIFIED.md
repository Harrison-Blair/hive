# 70-spec-allowlist verification

| Item | Value |
| --- | --- |
| Task | 70-spec-allowlist |
| Task commit | bc679c4a (code at fcdcf72f) |
| Base | dev at 8e444460 |
| Date | 2026-10-20 |
| Agent | verifier, Claude Code, claude-opus-5-5 |

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | Red: a detached worktree at `8e444460` with the task `create_test.go` and `worker_spawn_test.go`. `TestCreateSpecIgnoredPathStagesBriefAndReportsAllowlist` failed with `Outcome:failed Phase:validate ... path is ignored`. `TestWorkerSpawnAssignsAndPromptsIdle` failed because the prompt has no allowlist sentence. Both exit 1. Green at `bc679c4a`: the four worktree tests and `TestBriefPrompt` and `TestWorkerSpawnAssignsAndPromptsIdle` pass, exit 0. `TestSTERatchet` passes. The spec case of `TestBriefPrompt` cannot compile on the base because the signature changed. The prompt tests use literal strings. |
| V2 | pass | Temp `HIVE_HOME` and `XDG_CONFIG_HOME`, temp project with `.gitignore` `*` and `!*/`. The spec create of `1-first-spec` gave `outcome ok`, the `allowlist_needed` effect with both rules, and the copied brief. `git status --short` in the worktree printed `A  .hive/spec/demo/tasks/1-first-spec/brief.md`. After a `git add -f` commit of `.hive/spec/demo/README.md` on `main`, the create of the non-spec task `2-second-worker` gave `outcome failed`, `phase validate`, exit 1. No worktree path, no `task/2-second-worker` branch, no workspace. Cleanup: task cancel, `worktree remove --force` closed `wJH`, `herdr workspace close wJG`. No workspace for the temp paths remained. The temp folder was removed. |
| V3 | pass | `git diff dev --stat`: 13 files. `create.go`, `create_test.go`, `brief.go`, `brief_test.go`, `spawn_worker.go` (caller lines only), `worker_spawn_test.go`, the worker skill, `testdata/ste-baseline.json`, the brief copy, and the evidence `V1.md` to `V4.md`. The baseline change adds one entry with count 0 for the new brief copy. `AGENTS.md` requires this entry for a new instruction file. No count rises. |
| V4 | pass | `gofmt -l` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` printed `0 issues.`, exit 0. `go test -race ./...` exit 0. |

All items pass.
