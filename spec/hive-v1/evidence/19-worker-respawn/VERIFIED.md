# 19-worker-respawn VERIFIED

| Item | Value |
| --- | --- |
| Task | 19-worker-respawn |
| Commit | 2c69b4ad |
| Date | 2026-10-19 |
| Agent | verifier, pi, opencode-go/deepseek-v4.1-flash:xhigh |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | R1, R2, R3, R4, R5 | pass | The new respawn, workspace, name, spawn, and instruction tests fail on base `39a35e6c` by assertion and pass at `2c69b4ad`. The explicit-name collision test passes on both runs. |
| V2 | R4 | pass | Both instruction files carry the optional `--name` text, the free-name text, and the respawn text. The only `hive task assign` text is the prohibition. |
| V3 | R6 | pass | The diff lists the 10 source-table paths and the 4 required evidence files, and no other path. Caveat: the branch copies of V1, V2, and V3 name `39a35e6c` for the green runs, and V4 names `39a35e6c`. The corrected V1-V3 are uncommitted in the worktree. Commit the corrected evidence before the landing. |
| V4 | R1, R2, R3, R4, R5, R6 | pass | `gofmt -l` printed nothing. `go vet`, `golangci-lint`, and `go test -race` exited 0 at `2c69b4ad`. |

## V1 Respawn and name tests

The red run used a scratch worktree at base `39a35e6c` with the task test
files copied in. The new tests do not exist at the base.

```sh
git worktree add --detach /tmp/hive-19-base 39a35e6c
cp internal/agent/register_test.go /tmp/hive-19-base/internal/agent/register_test.go
cp internal/agent/worker_spawn_test.go /tmp/hive-19-base/internal/agent/worker_spawn_test.go
cp internal/agent/spawn_test.go /tmp/hive-19-base/internal/agent/spawn_test.go
cp internal/agent/manager_prompt_test.go /tmp/hive-19-base/internal/agent/manager_prompt_test.go
cd /tmp/hive-19-base
printf 'commit %s\n' "$(git rev-parse --short HEAD)"
go test ./internal/agent -run '^(TestWorkerRespawnRejectsLiveActingWorker|TestWorkerRespawnReopensAndReassigns|TestWorkerSpawnWithoutNamePicksLowestFreeTaskName|TestRegisterSpawnedTaskNameSkipsClaimedName)$' -count=1; echo "exit $?"
```

```text
commit 39a35e6c
--- FAIL: TestRegisterSpawnedTaskNameSkipsClaimedName (0.00s)
    --- FAIL: TestRegisterSpawnedTaskNameSkipsClaimedName/worker (0.00s)
        register_test.go:553: agent name = "worker-hive-1"; want "t1-worker-2"
    --- FAIL: TestRegisterSpawnedTaskNameSkipsClaimedName/verifier (0.00s)
        register_test.go:553: agent name = "verifier-hive-1"; want "t1-verifier-2"
--- FAIL: TestWorkerRespawnRejectsLiveActingWorker (0.01s)
    worker_spawn_test.go:604: envelope = output.Envelope{Outcome:"failed", Phase:"start", Effects:[]output.Effect{output.Effect{Resource:"agent", Path:".hive/state/agents/t16-worker-2.json", Action:"created"}, output.Effect{Resource:"agent", Path:".hive/state/agents/t16-worker-2.json", Action:"exited"}}, Error:"first tab busy: stop the agent or pass --pane", Result:agent.SpawnResult{Agent:records.Agent{Key:"t16-worker-2", Name:"t16-worker-2", Role:"worker", Project:"demo", Task:"16-quiet-no-nudge", Harness:"claude", Model:"m", PaneID:"w5:p1", PaneHistory:[]string(nil), WorkspaceID:"", TabID:"", TerminalID:"", Cwd:"", Status:"exited", AgentStatus:"", Finished:false, QuietAfterMin:10, Generation:0x1, StartedAt:"2026-10-19T18:40:25.323187894Z", UpdatedAt:"2026-10-19T18:40:25.323574112Z", LastError:"first tab busy: stop the agent or pass --pane", Spawned:true, ProcessPID:0, ProcessName:"", ProcessStart:0x0, AgentSessionRef:""}, Claim:(*records.Claim)(nil)}}; want the live acting worker refusal
--- FAIL: TestWorkerRespawnReopensAndReassigns (0.01s)
    worker_spawn_test.go:646: envelope = output.Envelope{Outcome:"failed", Phase:"validate", Effects:[]output.Effect{}, Error:"worktree pane mismatch: pane_id", Result:interface {}(nil)}; want ok
--- FAIL: TestWorkerSpawnWithoutNamePicksLowestFreeTaskName (0.01s)
    worker_spawn_test.go:687: agent name = "worker-demo-1"; want t16-worker-2
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.021s
FAIL
exit 1
```

The R4 check and the changed spawn expectation also fail on the base.

```sh
printf 'commit %s\n' "$(git rev-parse --short HEAD)"
go test ./internal/agent -run '^TestManagerInstructionsRequireRespawnRules$' -count=1; echo "exit $?"
go test ./internal/agent -run '^TestSpawnWorkerWithRealTask$' -count=1; echo "exit $?"
go test ./internal/agent -run '^TestRegisterDuplicateLiveNameAndBusyPane$' -count=1; echo "exit $?"
```

```text
commit 39a35e6c
--- FAIL: TestManagerInstructionsRequireRespawnRules (0.00s)
    manager_prompt_test.go:101: template does not contain "Hive picks a free name when you omit `--name`. Use `--name NAME` only when you need a fixed name."
    manager_prompt_test.go:101: template does not contain "hive agent spawn --role worker --task ID --harness KIND --model MODEL"
    manager_prompt_test.go:101: template does not contain "hive agent spawn --role verifier --task ID --harness KIND --model MODEL"
    manager_prompt_test.go:101: template does not contain "After a verify fail, spawn a worker for the task. Hive sets the acting worker. Do not run `hive task assign`."
    manager_prompt_test.go:101: skill does not contain "Hive picks a free name when you omit `--name`. Use `--name NAME` only when you need a fixed name."
    manager_prompt_test.go:101: skill does not contain "hive agent spawn --role worker --task ID --harness KIND --model MODEL"
    manager_prompt_test.go:101: skill does not contain "hive agent spawn --role verifier --task ID --harness KIND --model MODEL"
    manager_prompt_test.go:101: skill does not contain "After a verify fail, spawn a worker for the task. Hive sets the acting worker. Do not run `hive task assign`."
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.002s
FAIL
exit 1
commit 39a35e6c
--- FAIL: TestSpawnWorkerWithRealTask (0.00s)
    spawn_test.go:465: worker key/name = "worker-hive-1"/"worker-hive-1"
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.003s
FAIL
exit 1
commit 39a35e6c
ok  	github.com/Harrison-Blair/hive/internal/agent	0.002s
exit 0
```

The explicit-name collision check passes on the base, as the regression case
requires.

The green run used a clean scratch worktree at the task commit.

```sh
git worktree add --detach /tmp/hive-19-verify 2c69b4ad
cd /tmp/hive-19-verify
printf 'commit %s\n' "$(git rev-parse --short HEAD)"
go test ./internal/agent -run '^(TestWorkerRespawnRejectsLiveActingWorker|TestWorkerRespawnReopensAndReassigns|TestWorkerSpawnWithoutNamePicksLowestFreeTaskName|TestRegisterSpawnedTaskNameSkipsClaimedName|TestRegisterDuplicateLiveNameAndBusyPane|TestSpawnWorkerWithRealTask|TestManagerInstructionsRequireRespawnRules)$' -count=1; echo "exit $?"
```

```text
commit 2c69b4ad
ok  	github.com/Harrison-Blair/hive/internal/agent	0.022s
exit 0
```

`TestWorkerRespawnReopensAndReassigns` checks the acting agent, the reopened
`workspace_id`, `tab_id`, and `root_pane`, one `worktree.open`, and no
`worktree.create`. `TestWorkerRespawnRejectsLiveActingWorker` checks the error
text, no `agent.start`, an unchanged task record, and only the old agent
record. `TestWorkerSpawnWithoutNamePicksLowestFreeTaskName` checks the picked
name and the acting agent.

## V2 Manager instructions

```sh
printf 'commit %s\n' "$(git rev-parse --short HEAD)"
rg -n 'Hive picks a free name|--name|After a verify fail|hive task assign|hive agent spawn --role' .agents/skills/hive-manager/SKILL.md internal/agent/templates/project-agents.md; echo "exit $?"
```

```text
commit 2c69b4ad
internal/agent/templates/project-agents.md:40:hive agent spawn --role worker --task ID --harness KIND --model MODEL
internal/agent/templates/project-agents.md:41:hive agent spawn --role verifier --task ID --harness KIND --model MODEL
internal/agent/templates/project-agents.md:52:Hive takes the project from the task. Hive picks a free name when you omit `--name`. Use `--name NAME` only when you need a fixed name.
internal/agent/templates/project-agents.md:54:After a verify fail, spawn a worker for the task. Hive sets the acting worker. Do not run `hive task assign`.
internal/agent/templates/project-agents.md:56:After a verify fail, spawn a fresh verifier with no `--pane`. Do not make placeholder tabs.
.agents/skills/hive-manager/SKILL.md:58:Spawn one worker for one ready task. Put the task ID in the preamble. Hive picks a free name when you omit `--name`. Use `--name NAME` only when you need a fixed name.
.agents/skills/hive-manager/SKILL.md:61:hive agent spawn --role worker --task ID --harness KIND --model MODEL
.agents/skills/hive-manager/SKILL.md:64:Hive takes the project from the task. After a verify fail, spawn a worker for the task. Hive sets the acting worker. Do not run `hive task assign`.
.agents/skills/hive-manager/SKILL.md:72:for that task. Hive picks a free name when you omit `--name`. Use `--name NAME`
.agents/skills/hive-manager/SKILL.md:76:hive agent spawn --role verifier --task ID --harness KIND --model MODEL
.agents/skills/hive-manager/SKILL.md:79:After a verify fail, spawn a fresh verifier with no `--pane`. Do not make placeholder tabs.
exit 0
```

Both files show `--name` as optional and a free name when it is absent. Both
files remove the manual `hive task assign` step after a respawn. The only
`hive task assign` text in each file is the instruction not to run it.

## V3 Change scope

```sh
printf 'commit %s\n' "$(git rev-parse --short HEAD)"
git diff dev --stat; echo "exit $?"
BASE=$(git merge-base dev HEAD); printf 'base %s\n' "$BASE"; git diff --stat "$BASE" HEAD; echo "exit $?"
git diff --stat; echo "exit $?"
```

```text
commit 2c69b4ad
 .agents/skills/hive-manager/SKILL.md          |  11 +-
 internal/agent/manager_prompt_test.go         |  30 +++++-
 internal/agent/register.go                    |  30 +++++-
 internal/agent/register_test.go               |  30 ++++++
 internal/agent/spawn_test.go                  |   8 +-
 internal/agent/spawn_worker.go                |  81 +++++++++++----
 internal/agent/templates/project-agents.md    |   8 +-
 internal/agent/worker_spawn_test.go           | 116 +++++++++++++++++++++
 spec/hive-v1/evidence/19-worker-respawn/V1.md | 141 ++++++++++++++++++++++++++
 spec/hive-v1/evidence/19-worker-respawn/V2.md |  35 +++++++
 spec/hive-v1/evidence/19-worker-respawn/V3.md |  61 +++++++++++
 spec/hive-v1/evidence/19-worker-respawn/V4.md |  98 ++++++++++++++++++++++++
 spec/hive-v1/tasks/19-worker-respawn.md       | 140 +++++++++++++++++++++++
 testdata/ste-baseline.json                    |   1 +
 14 files changed, 752 insertions(+), 38 deletions(-)
exit 0
base 39a35e6ce3b124838321676669614d2d7673a0ed
 .agents/skills/hive-manager/SKILL.md          |  11 +-
 internal/agent/manager_prompt_test.go         |  30 +++++-
 internal/agent/register.go                    |  30 +++++-
 internal/agent/register_test.go               |  30 ++++++
 internal/agent/spawn_test.go                  |   8 +-
 internal/agent/spawn_worker.go                |  81 +++++++++++----
 internal/agent/templates/project-agents.md    |   8 +-
 internal/agent/worker_spawn_test.go           | 116 +++++++++++++++++++++
 spec/hive-v1/evidence/19-worker-respawn/V1.md | 141 ++++++++++++++++++++++++++
 spec/hive-v1/evidence/19-worker-respawn/V2.md |  34 +++++++
 spec/hive-v1/evidence/19-worker-respawn/V3.md |  55 ++++++++++
 spec/hive-v1/evidence/19-worker-respawn/V4.md |  98 ++++++++++++++++++++++++
 spec/hive-v1/tasks/19-worker-respawn.md       | 140 +++++++++++++++++++++++
 testdata/ste-baseline.json                    |   1 +
 14 files changed, 745 insertions(+), 38 deletions(-)
exit 0
 spec/hive-v1/evidence/19-worker-respawn/V1.md | 12 ++++----
 spec/hive-v1/evidence/19-worker-respawn/V2.md |  5 ++--
 spec/hive-v1/evidence/19-worker-respawn/V3.md | 42 ++++++++++++++------------
 3 files changed, 33 insertions(+), 26 deletions(-)
exit 0
```

The 14 paths are the 10 source-table paths and the 4 required evidence files.
The branch range `39a35e6c..2c69b4ad` has the same 14 paths. The last block is
the uncommitted V1-V3 correction in the worktree. The branch copies of V1, V2,
and V3 name `39a35e6c` for the green runs; the worktree copies name
`2c69b4ad`. V4 names `39a35e6c` in the table and its run has no commit line.
Commit and push the corrected V1-V3 and update V4 before the landing. Until
then, the branch carries stale evidence. The copied brief removes the R3
sentence about a task ID with no leading number and adds a `Manager decision`.
`records.ValidTaskID` rejects such an ID before the spawn path, so the omitted
fallback is unreachable through `hive agent spawn`.

## V4 Gate

The four gate commands ran from the worktree root.

```sh
printf 'commit %s\n' "$(git rev-parse --short HEAD)"
printf '%s\n' '--- gofmt ---'
gofmt -l $(git ls-files -co --exclude-standard '*.go'); echo "gofmt exit $?"
printf '%s\n' '--- go vet ---'
go vet ./...; echo "vet exit $?"
golangci-lint run ./...; echo "lint exit $?"
go test -race ./...; echo "test exit $?"
```

```text
commit 2c69b4ad
--- gofmt ---
gofmt exit 0
--- go vet ---
vet exit 0
0 issues.
lint exit 0
ok  	github.com/Harrison-Blair/hive	4.048s
ok  	github.com/Harrison-Blair/hive/cmd	1.418s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.020s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.323s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.020s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	(cached)
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	(cached)
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.009s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	2.944s
ok  	github.com/Harrison-Blair/hive/internal/doctor	3.028s
ok  	github.com/Harrison-Blair/hive/internal/handoff	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/land	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.024s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/log	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/process	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/records	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/run	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/store	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/logs	(cached)
ok  	github.com/Harrison-Blair/hive/internal/memory	(cached)
ok  	github.com/Harrison-Blair/hive/internal/project	(cached)
ok  	github.com/Harrison-Blair/hive/internal/supervise	(cached)
ok  	github.com/Harrison-Blair/hive/internal/task	(cached)
ok  	github.com/Harrison-Blair/hive/internal/tui	(cached)
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
test exit 0
```

`gofmt -l` printed nothing. Vet, lint, and the race tests exited 0.

## Scratch cleanup

```sh
git worktree remove --force /tmp/hive-19-base; echo "remove base exit $?"
git worktree remove --force /tmp/hive-19-verify; echo "remove verify exit $?"
if test ! -e /tmp/hive-19-base && test ! -e /tmp/hive-19-verify; then printf '%s\n' 'scratch worktrees removed'; else printf '%s\n' 'scratch worktree remains'; fi
git worktree list | rg 'hive-19' || true
```

```text
remove base exit 0
remove verify exit 0
scratch worktrees removed
```
