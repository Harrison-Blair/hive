# VERIFIED: 46-task-feature

Verifier: independent Claude Code session (claude-opus-5-5). Task commit `9ab94f2b`. Base `dev` at `707e332f`.

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | Detached worktree at `707e332f` with only the six changed test files copied in. All five new tests failed for the correct reason: Git ignored both spec paths; task JSON had no `feature`/`spec`; template and skill lacked the `--feature F` and `--spec` lines; `unknown flag: --feature`; create without `--feature` succeeded. On the task branch all five passed, with `TestSTERatchet`. |
| V2 | pass | Live run with the built binary, temp `HIVE_HOME` and temp `XDG_CONFIG_HOME`, temp repo registered as `one` (base `main`). (1) create without `--feature`: exit 1, `validate`, `feature requires a value`, no record. (2) `--feature f`, no README: exit 1, `validate`, `feature README missing on base: .hive/spec/f/README.md`. (3) `--feature f --spec`: ok, `1-spec`, record has `"spec": true`. (4) README committed on `main`, `--feature f`: ok, `2-normal`. (5) `task list --feature f`: `1-spec` and `2-normal`. (6) `task show 2-normal`: feature, brief and evidence paths, `warning: linked file missing: .hive/spec/f/tasks/2-normal/brief.md`; `--json` has it in `result.warnings`. (7) Hand-written old record `3-old` in status `verified` with no feature: list and show work; worker edit refused (`not the controlling role`); manager edit `--feature g` ok, record has `"feature": "g"`, status unchanged. (8) Edit again: exit 1, `task already has a feature: g`. Also: `--feature` with `--brief` refused (`not both`). |
| V3 | pass | `git check-ignore -q .hive/spec/x/README.md; echo $?` printed `1`. `.hive/state/x` still ignored (exit 0). |
| V4 | pass | `git diff dev --stat`: 29 files. All are in the source table or evidence, except the new test `cmd/task/feature_test.go`, which is a new test beside the `cmd/task/*` sources (allowed by the task flow). `cmd/task_lifecycle_test.go` changes add only `--feature f --spec`. `testdata/ste-baseline.json` only adds the brief copy with 0. |
| V5 | pass | `gofmt -l` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` `0 issues.`, exit 0. `go test -race ./...` exit 0. |

Note: no direct unit test exists for `internal/task/paths.go`; the CLI test covers the paths through `task show`.
