# VERIFIED — 75-needs-user

Verifier re-ran each item at commit `5b073266` on 2026-10-20.

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | Red: in a `git archive dev` tree with the four changed test files copied in, `TestTaskNeedsUserJSON` failed to build (`task.NeedsUser undefined`), `TestShowAndListRenderNeedsUser` failed (ShowText had no needs-user line), `TestNeedsUser*` in `cmd/task/edit` failed (`unknown flag: --needs-user`, no `--clear-needs-user` flag), and `TestManagerInstructionsRequireRespawnRules` failed (skill and template lack the new sentence). Green: each passes on the task branch. `TestSTERatchet` passes. |
| V2 | pass | Built binary, temp `HIVE_HOME` and `XDG_CONFIG_HOME`, verified task `90-decision` controlled by `manager-one`. Manager set: exit 0; show printed `needs user: choose an option`; list line ended `[needs user]`; `--json task list` had `"needs_user":"choose an option"`. Manager clear: exit 0; show and list omitted the text. Worker set: exit 1, `hive: not the controlling role: worker`; field stayed empty. No herdr workspace made. Temp dirs removed. |
| V3 | pass | `git diff dev --stat`: 16 files, each in the source table (`internal/agent/manager_prompt_test.go` is the test that asserts the manager text) or the evidence folder. The `testdata/ste-baseline.json` change adds the brief copy with count 0, which the manager authorized. |
| V4 | pass | `gofmt -l` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` `0 issues.`, exit 0. `go test -race ./...` exit 0. |
