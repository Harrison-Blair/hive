# 47-brief-copy verification

| Item | Value |
| --- | --- |
| Task | 47-brief-copy |
| Commit verified | 89b40b66 (implementation 0e7ababa) |
| Base | d279b7d7 (merge base with `dev`) |
| Date | 2026-10-19 |
| Agent | t47-verifier-1, Claude Code, claude-opus-5-5 |
| Verdict | fail (V3) |

## Results

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | On base d279b7d7 with the new test files copied in: `TestCreateFeatureIgnoredPathRefusesBeforeWorktreeCreation`, `TestCreateCopiesFeatureBriefIntoUntrackedWorktree`, `TestEditCopiesFeatureBriefToActiveWorktree`, `TestWorkerSpawnAssignsAndPromptsIdle`, `TestVerifierSpawnReopensWorktree` fail on the missing behavior (no validate failure, no brief file, prompt without the paths). `TestCreateWithoutFeatureSkipsFeatureIgnoreCheckAndCopy` passes on base; it is a regression guard (accepted by manager guidance). The new `TestBriefPrompt` feature cases cannot compile on base (signature change); the same prompt text is asserted by the two spawn tests that fail on base. On 89b40b66 all of these tests pass. |
| V2 | pass | Temp `HIVE_HOME`, temp `XDG_CONFIG_HOME`, task binary. Allow repo: `worktree create` ok, effect `brief copied`; `git status --short --untracked-files=all` = `?? .hive/spec/f/tasks/1-v2-allow/brief.md`; `check-ignore` = not ignored; `cat` = `first brief`. `task edit --brief` ok with a second `brief copied` effect; `cat` = `second brief text`. Ignore repo: `worktree create` exit 1, `{"outcome":"failed","phase":"validate","effects":[],"error":"path is ignored: .hive/spec/f/tasks/2-v2-ignore/brief.md. The .gitignore of the repo must allow .hive/spec/."}`; no worktree path, no branch, one Git worktree only. |
| V3 | fail | `git diff dev...HEAD --stat` lists `internal/agent/spawn_worker.go` (8 lines: the two callers pass `choice.Task.Feature` to the new prompt signatures). This file is not in the source table, and the brief says "Do not edit other files." All other paths are in the table or are the evidence files under `evidence/47-brief-copy/` (accepted by manager guidance). `worker_spawn_test.go` is counted as a test of the prompts. Note: `git diff dev --stat` (two-dot) also shows `reference/dogfood/friction/` changes; these come from `dev` commit f2a67dce after the merge base, not from the task. |
| V4 | pass | `gofmt -l` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` `0 issues.`, exit 0. `go test -race ./...` exit 0. |

## Cleanup

The V2 run made herdr workspaces wHG (worktree) and wHF (repo). Both are closed; `herdr workspace get` fails for each and no workspace lists the temp path. The first `worktree remove` failed because the worktree held the untracked brief, and the temp files were removed before the retry, so the two workspaces were closed with `herdr workspace close`. The temp base worktree for V1 is removed and pruned.

## Fail reason

V3: `internal/agent/spawn_worker.go` changed and is not in the source table. The change is a mechanical caller update for the new `workerPrompt`/`verifierPrompt` signatures. The manager can accept it by an amendment of the brief, or the worker can move the change.

# 47-brief-copy verification, run 2

| Item | Value |
| --- | --- |
| Task | 47-brief-copy |
| Commit verified | 3c5a9574 (implementation 0e7ababa; 89b40b66..3c5a9574 changes only evidence files) |
| Base | d279b7d7 (merge base with `dev`) |
| Date | 2026-10-19 |
| Agent | t47-verifier-2, Claude Code, claude-opus-5-5 |
| Verdict | pass |

## Results

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | Red: a detached worktree at d279b7d7 with the new `create_test.go`, `edit_test.go`, and `worker_spawn_test.go` copied in. `TestCreateFeatureIgnoredPathRefusesBeforeWorktreeCreation` fails (phase `workspace`, want `validate`), `TestCreateCopiesFeatureBriefIntoUntrackedWorktree` fails (no brief file), `TestEditCopiesFeatureBriefToActiveWorktree` fails (no brief file), `TestWorkerSpawnAssignsAndPromptsIdle` and `TestVerifierSpawnReopensWorktree` fail (prompt without the brief and evidence paths). `TestCreateWithoutFeatureSkipsFeatureIgnoreCheckAndCopy` passes on base; it is a no-feature regression guard (accepted by manager guidance). The new `brief_test.go` does not compile on base (`too many arguments in call to workerPrompt`); the spawn tests assert the same prompt text and fail on base. Green: on 3c5a9574 all of these tests and `TestBriefPrompt*` pass, exit 0. The base worktree is removed and pruned. |
| V2 | pass | Temp `HIVE_HOME` and temp `XDG_CONFIG_HOME` in the scratchpad, task binary. Allow repo: `worktree create` ok with effect `brief copied`; `git status --short --untracked-files=all` = `?? .hive/spec/f/tasks/1-v2a/brief.md`; `check-ignore --no-index` = not ignored; `cat` = `first brief`. `task edit --brief` ok with effects `brief updated` (state) and `brief copied` (worktree); `cat` = `second brief`. Ignore repo: `worktree create` exit 1, `{"outcome":"failed","phase":"validate","effects":[],"error":"path is ignored: .hive/spec/f/tasks/2-v2i/brief.md. The .gitignore of the repo must allow .hive/spec/.","result":null}`; no worktree path, no branch, one Git worktree only. |
| V3 | pass | `git diff dev...HEAD --stat`: 17 files. Each code file is in the amended source table: the two skills, `internal/agent/brief.go` and `brief_test.go`, `internal/agent/spawn_worker.go` (caller of the changed prompt functions, now allowed by the brief), `internal/agent/worker_spawn_test.go` (prompt tests), `internal/task/edit.go` and test, `internal/worktree/create.go` and test, `spec/hive-v1/tasks/47-brief-copy.md`, `testdata/ste-baseline.json`. The other five files are the evidence under `evidence/47-brief-copy/` (accepted by manager guidance). Two-dot `git diff dev` also shows `reference/dogfood/friction/` changes from `dev` after the merge base, not from the task. |
| V4 | pass | `gofmt -l` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` `0 issues.`, exit 0. `go test -race ./...` exit 0. |

## Cleanup

The V2 run made herdr workspaces wHM (worktree) and wHK (repo); both belonged to the temp repo. `hive worktree remove --task 1-v2a` refused (`task is created: complete, verify, or cancel it first`), so `herdr worktree remove --workspace wHM` and `herdr workspace close wHK` closed them. `herdr workspace get` fails for each, and no workspace lists the temp path. The temp directory is removed.
