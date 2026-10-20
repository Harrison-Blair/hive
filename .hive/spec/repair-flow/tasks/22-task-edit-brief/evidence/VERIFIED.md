# 22-task-edit-brief VERIFIED

| Item | Value |
| --- | --- |
| Task | 22-task-edit-brief |
| Commit | c2374641 |
| Base | dev at f49a0135 |
| Date | 2026-10-19 |
| Agent | verifier, Claude Code, claude-opus-5-5 |
| Verdict | pass |

This verification replaces the earlier pass at `3ff6438b`. The task branch now holds the merge repair `10b5b191` (V5.md) and `c2374641`.

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | R1, R2, R3 | pass | In a detached worktree at dev `f49a0135` with the two task test files from `c2374641`, `TestTaskCLIEditBrief` failed (`unknown flag: --task`) and `TestManagerInstructionsRequireRespawnRules` failed (template and skill lack the edit sentence). Both pass at `c2374641`. |
| V2 | R1 | pass | Scratch home: the manager edited the assigned task, and `task show --brief` changed from `Live edit` to `Verifier repair brief`. After `task complete`, the edit failed with `hive: task is completed: edit needs created or assigned`, exit 1. The brief SHA-256 was the same before and after the failed edit. The log event `task brief edit` has `task_id`, `agent_id`, `old_size` 9, `new_size` 22, and no brief text. |
| V3 | R1-R4 | pass | `git diff dev --stat` lists 18 files. Each is in the source table or in `evidence/22-task-edit-brief/`. The merge from dev added no file outside that set. |
| V4 | R1-R4 | pass | `gofmt -l` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` 0 issues, exit 0. `go test -race -count=1 ./...` exit 0, no failed package. |

## Runs

```sh
git worktree add --detach $S/base dev            # f49a0135
git show c2374641:cmd/task_lifecycle_test.go > cmd/task_lifecycle_test.go
git show c2374641:internal/agent/manager_prompt_test.go > internal/agent/manager_prompt_test.go
go test ./cmd -run '^TestTaskCLIEditBrief$' -count=1
go test ./internal/agent -run '^TestManagerInstructionsRequireRespawnRules$' -count=1
# then the same two commands in the task worktree at c2374641
```

```text
--- FAIL: TestTaskCLIEditBrief (0.00s)
    task_lifecycle_test.go:184: created task edit status = 2; stdout "", stderr "hive: unknown flag: --task\n"; want success
exit 1
--- FAIL: TestManagerInstructionsRequireRespawnRules (0.00s)
    manager_prompt_test.go:103: template does not contain "Use `hive task edit --task ID --brief FILE` to change a brief before the work is complete."
    manager_prompt_test.go:103: skill does not contain "Use `hive task edit --task ID --brief FILE` to change a brief before the work is complete."
exit 1
== green c2374641
ok  	github.com/Harrison-Blair/hive/cmd	0.005s
exit 0
ok  	github.com/Harrison-Blair/hive/internal/agent	0.001s
exit 0
```

```text
--- before
Live edit
--- edit assigned as manager
updated brief .hive/state/briefs/1-live-edit.md
exit 0
--- show
Verifier repair brief
id: 1-live-edit
status: completed
a85b6c95a912f6be
--- edit completed
hive: task is completed: edit needs created or assigned
exit 1
a85b6c95a912f6be
--- show
Verifier repair brief
{"agent_id":"manager-demo-g1","command":"task edit","component":"task","level":"DEBUG","msg":"task brief edit","new_size":22,"old_size":9,"project_id":"demo","task_id":"1-live-edit",...}
```

```text
--- gofmt
exit 0
--- vet
exit 0
--- lint
0 issues.
exit 0
--- test
exit 0
```

## Cleanup

The scratch worktree, scratch home, and build binary were in the session scratchpad. They were removed after the runs.
