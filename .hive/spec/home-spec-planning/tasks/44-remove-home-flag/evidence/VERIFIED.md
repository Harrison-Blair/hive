# VERIFIED — 44-remove-home-flag

Verifier: independent Claude Code session (Opus 5.5), 2026-10-19. Task commit: `648e801b`. Base: `dev` at `b4a35167`.
Each check used a temporary `XDG_CONFIG_HOME` and unset `HIVE_HOME`, except where a check sets `HIVE_HOME`. The user's config was not read or written.

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | Base worktree at `b4a35167` with the task's `internal/agent/brief_test.go` and the two new `cmd` tests added: `TestBriefPrompt` and `TestBriefPromptShellSafe` failed (prompt contained `hive --home '<home>'`); `TestHomeFlagIsUnknown` failed (status 0, version printed); `TestHomeFlagUnknownPrecedesHomeResolution` failed (stderr `hive: --home requires a path`). On the task branch all four pass. |
| V2 | pass | Task binary: `./hive --home /tmp task list --project hive` printed `hive: unknown flag: --home`, exit 2. `./hive task list --project hive` from the worktree listed the home tasks (32 lines, including `44-remove-home-flag`), exit 0. `HIVE_HOME=<temp marked home> ./hive task list --project x` printed `no tasks`, exit 0, no `home conflict`. |
| V3 | pass | One line remains: `spec/hive-v1/tasks/T01-home-and-register.md:56:T01 added the persistent \`--home\` flag.` This is the kept sentence listed in the evidence (V3.md). |
| V4 | pass | Config `$XDG_CONFIG_HOME/hive/config.toml` with `home = "<other temp marked home>"`; `go test -race ./... -count=1` passed for all packages, exit 0. |
| V5 | pass | `git diff dev --stat`: 62 files. Every file is in the source table (callers `cmd/root.go`, `cmd/common/common.go`, `cmd/logs/logs.go`, `internal/lib/home/home.go`, `internal/tui/tui.go`, `internal/agent/brief.go`; Go tests; skills; tracker, architecture, risks, briefs; `testdata/ste-baseline.json` adds only the brief copy at 0) or is task evidence under `evidence/44-remove-home-flag/`. |
| V6 | pass | `gofmt -l` printed nothing; `go vet ./...` exit 0; `golangci-lint run ./...` `0 issues.` exit 0; `go test -race ./...` exit 0. |
