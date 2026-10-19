# 22-task-edit-brief VERIFIED

| Item | Value |
| --- | --- |
| Task | 22-task-edit-brief |
| Commit | 3ff6438b |
| Base | dev at 457a85a4 |
| Date | 2026-10-19 |
| Agent | verifier, Claude Code, claude-opus-5-5 |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | R1, R2, R3 | pass | In a detached worktree at dev `457a85a4` with the two task test files from `3ff6438b`, `TestTaskCLIEditBrief` failed (`unknown flag: --task`) and `TestManagerInstructionsRequireRespawnRules` failed (template and skill lack the edit sentence). Both pass at `3ff6438b`. Extra mutation runs at `3ff6438b`: with the status check removed, the test fails at line 201 (completed edit succeeds); with the role check removed, it fails at line 211 (worker edit succeeds); with the log sizes swapped, it fails at line 267. |
| V2 | R1 | pass | Scratch home: the manager edited the assigned task, and `task show --brief` changed from `Live edit` to `Verifier live brief`. After `task complete`, the edit failed with `hive: task is completed: edit needs created or assigned`, exit 1. The brief SHA-256 was the same before and after the failed edit. The log event `task brief edit` has `task_id`, `agent_id`, `old_size` 9, `new_size` 20, and no brief text. |
| V3 | R1-R4 | pass | `git diff dev --stat` lists 16 files. Each is in the source table (`cmd/task/edit/edit.go`, `cmd/task/task.go`, `cmd/task_lifecycle_test.go`, `internal/task/{doc,edit,log,transition}.go`, both instruction files, `manager_prompt_test.go`, `testdata/ste-baseline.json` with one added line of value 0, the brief copy) or in `evidence/22-task-edit-brief/`. The brief copy differs from `task show --brief` only in the two STE rewordings that V1.md records. |
| V4 | R1-R4 | pass | `gofmt -l` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` 0 issues, exit 0. `go test -race ./...` exit 0, no failed package. |

## Runs

```sh
git worktree add --detach $S/base dev            # 457a85a4
git show 3ff6438b:cmd/task_lifecycle_test.go > cmd/task_lifecycle_test.go
git show 3ff6438b:internal/agent/manager_prompt_test.go > internal/agent/manager_prompt_test.go
go test ./cmd -run '^TestTaskCLIEditBrief$' -count=1
go test ./internal/agent -run '^TestManagerInstructionsRequireRespawnRules$' -count=1
# then the same two commands in the task worktree at 3ff6438b
# mutations at 3ff6438b in the scratch worktree, one at a time, each restored with git checkout
```

```text
--- FAIL: TestTaskCLIEditBrief (0.00s)
    task_lifecycle_test.go:184: created task edit status = 2; stdout "", stderr "hive: unknown flag: --task\n"; want success
--- FAIL: TestManagerInstructionsRequireRespawnRules (0.00s)
    manager_prompt_test.go:103: template does not contain "Use `hive task edit --task ID --brief FILE` to change a brief before the work is complete."
    manager_prompt_test.go:103: skill does not contain "Use `hive task edit --task ID --brief FILE` to change a brief before the work is complete."
ok  	github.com/Harrison-Blair/hive/cmd	0.005s
ok  	github.com/Harrison-Blair/hive/internal/agent	0.001s
== mutant: no status check
    task_lifecycle_test.go:201: completed task edit status = 0; stdout "updated brief .hive/state/briefs/2-edit-in-completed.md\n", stderr ""; want status refusal
== mutant: no role check
    task_lifecycle_test.go:211: worker edit status = 0; stdout "updated brief .hive/state/briefs/3-edit-by-worker.md\n", stderr ""; want role refusal
== mutant: sizes swapped
    task_lifecycle_test.go:267: brief edit event = ... "new_size":15, "old_size":45 ...; want task, caller, and byte sizes
```

```text
--- before
Live edit
--- edit assigned as manager
updated brief .hive/state/briefs/1-live-edit.md
--- show
Verifier live brief
id: 1-live-edit
status: completed
b4e2b6c8...0a7d  .../live/.hive/state/briefs/1-live-edit.md
--- edit completed
hive: task is completed: edit needs created or assigned
exit 1
b4e2b6c8...0a7d  .../live/.hive/state/briefs/1-live-edit.md
--- show
Verifier live brief
{"agent_id":"manager-demo-g1","command":"task edit","component":"task","level":"DEBUG","msg":"task brief edit","new_size":20,"old_size":9,"project_id":"demo","task_id":"1-live-edit",...}
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
