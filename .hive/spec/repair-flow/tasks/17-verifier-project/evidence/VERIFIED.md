# 17-verifier-project VERIFIED

| Item | Value |
| --- | --- |
| Task | 17-verifier-project |
| Commit | f846ae75 |
| Date | 2026-10-19 |
| Agent | verifier, pi, opencode-go/deepseek-v4.1-flash:xhigh |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | R1, R3 | pass | The four new tests fail by assertion on the base code and pass on the task branch. A live CLI check also takes the project from the task. |
| V2 | R2, R3 | pass | Both instruction files show the worker and verifier spawn commands with no `--project` and no `--pane`. Both files contain the project sentence and the re-verify rule. |
| V3 | R4, R5 | pass | `TestSTE` passes. The baseline adds only `spec/hive-v1/tasks/17-verifier-project.md` at 0, and no count rose. |
| V4 | Source scope | pass | The task branch changes the eight source paths of the source table and the five required `V*.md` evidence paths. `dev` advanced by one friction commit after the branch point, so the current two-dot diff shows that entry in reverse. The three-dot diff and the diff from `1a7a75a6` show only the task paths. This file adds `VERIFIED.md` after the run. |
| V5 | Gate | pass | `gofmt -l` prints nothing. `go vet`, `golangci-lint`, and `go test -race ./...` exit 0. |

## Runs

### V1 Red run

I ran the task test files against the base tree. The tree is `git archive 1a7a75a6`, with the two test files from the task branch copied in.

```sh
BASE=$(mktemp -d /home/penguin/source/hive/.hive/tmp/17-verify-base.XXXXXX)
git archive 1a7a75a6 | tar -x -C "$BASE"
cp internal/agent/spawn_project_test.go "$BASE/internal/agent/spawn_project_test.go"
cp internal/agent/manager_prompt_test.go "$BASE/internal/agent/manager_prompt_test.go"
cd "$BASE"
echo "commit $(git -C /home/penguin/source/hive/projects/hive/worktrees/17-verifier-project rev-parse --short 1a7a75a6)"
for t in TestVerifierSpawnPaneUsesTaskProject TestWorkerSpawnPaneUsesTaskProject TestSpawnPaneRejectsDifferentTaskProject TestManagerInstructionsRequireFreshVerifierAfterFail; do go test ./internal/agent -run "^$t$" -count=1; echo "exit $?"; done
```

```text
commit 1a7a75a6
--- FAIL: TestVerifierSpawnPaneUsesTaskProject (0.01s)
    spawn_project_test.go:40: envelope = output.Envelope{Outcome:"failed", Phase:"validate", Effects:[]output.Effect{}, Error:"verifier requires --project and --task", Result:interface {}(nil)}; want ok
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.008s
FAIL
exit 1
--- FAIL: TestWorkerSpawnPaneUsesTaskProject (0.01s)
    spawn_project_test.go:58: envelope = output.Envelope{Outcome:"failed", Phase:"validate", Effects:[]output.Effect{}, Error:"worker requires --project and --task", Result:interface {}(nil)}; want ok
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.007s
FAIL
exit 1
--- FAIL: TestSpawnPaneRejectsDifferentTaskProject (0.01s)
    spawn_project_test.go:82: envelope = output.Envelope{Outcome:"failed", Phase:"validate", Effects:[]output.Effect{}, Error:"project not found: other", Result:interface {}(nil)}; want failure "task 1-worker-demo is in project demo"
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.007s
FAIL
exit 1
--- FAIL: TestManagerInstructionsRequireFreshVerifierAfterFail (0.00s)
    manager_prompt_test.go:86: manager instructions do not contain "After a verify fail, spawn a fresh verifier with no `--pane`. Do not make placeholder tabs."
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.001s
FAIL
exit 1
```

### V1 Green run

```sh
echo "commit $(git rev-parse --short HEAD)"
for t in TestVerifierSpawnPaneUsesTaskProject TestWorkerSpawnPaneUsesTaskProject TestSpawnPaneRejectsDifferentTaskProject TestManagerInstructionsRequireFreshVerifierAfterFail; do go test ./internal/agent -run "^$t$" -count=1 -v; echo "exit $?"; done
```

```text
commit f846ae75
=== RUN   TestVerifierSpawnPaneUsesTaskProject
--- PASS: TestVerifierSpawnPaneUsesTaskProject (0.01s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/agent	0.008s
exit 0
=== RUN   TestWorkerSpawnPaneUsesTaskProject
--- PASS: TestWorkerSpawnPaneUsesTaskProject (0.01s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/agent	0.008s
exit 0
=== RUN   TestSpawnPaneRejectsDifferentTaskProject
--- PASS: TestSpawnPaneRejectsDifferentTaskProject (0.01s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/agent	0.007s
exit 0
=== RUN   TestManagerInstructionsRequireFreshVerifierAfterFail
--- PASS: TestManagerInstructionsRequireFreshVerifierAfterFail (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/agent	0.001s
exit 0
```

### V1 Live CLI check

I built the binary at `f846ae75` and used a temporary home with project `demo` and task `1-demo`. The herdr socket did not exist, so the second spawn stopped at the shell wait after it made the record.

```sh
TMP=/home/penguin/source/hive/.hive/tmp/17-verify-live
HIVE=/home/penguin/source/hive/.hive/tmp/hive17-bin/hive
export HERDR_SOCKET_PATH="$TMP/no-such.sock"
"$HIVE" --home "$TMP/home" --json agent spawn --role verifier --task 1-demo --pane w5:p1 --project other --harness pi --model test --name verifier-demo-1; echo "exit $?"
"$HIVE" --home "$TMP/home" --json agent list | head -c 400; echo; echo "exit $?"
"$HIVE" --home "$TMP/home" --json agent spawn --role verifier --task 1-demo --pane w5:p1 --harness pi --model test --name verifier-demo-2 2>&1 | head -c 1200; echo; echo "exit ${PIPESTATUS[0]}"
```

```text
{"outcome":"failed","phase":"validate","effects":[],"error":"task 1-demo is in project demo","result":null}
exit 1
{"outcome":"ok","phase":"complete","effects":[],"error":null,"result":[]}

exit 0
{"outcome":"failed","phase":"start","effects":[{"resource":"agent","path":".hive/state/agents/verifier-demo-2.json","action":"created"},{"resource":"agent","path":".hive/state/agents/verifier-demo-2.json","action":"exited"}],"error":"timeout: pane w5:p1 did not become an available shell within 10s","result":{"agent":{"key":"verifier-demo-2","name":"verifier-demo-2","role":"verifier","project":"demo","task":"1-demo","harness":"pi","model":"test","pane_id":"w5:p1","workspace_id":"","tab_id":"","terminal_id":"","cwd":"","status":"exited","finished":false,"quiet_after_min":10,"generation":1,"started_at":"2026-10-19T18:09:50.63966273Z","updated_at":"2026-10-19T18:10:00.681515129Z","last_error":"timeout: pane w5:p1 did not become an available shell within 10s","spawned":true}}}
exit 1
```

The mismatched project fails before it makes a record, and the list is empty. The spawn with no `--project` takes project `demo` from the task record. Cleanup:

```sh
rm -rf "$BASE" /home/penguin/source/hive/.hive/tmp/17-verify-live /home/penguin/source/hive/.hive/tmp/hive17-bin /tmp/hive17-brief.txt /tmp/hive17-gate.txt /tmp/hive17-base-dir; echo "exit $?"
```

```text
exit 0
```

### V2 Instruction text

```sh
echo "commit $(git rev-parse --short HEAD)"
grep -nE 'hive agent spawn --role (worker|verifier)|Hive takes the project from the task\.|After a verify fail, spawn a fresh verifier with no `--pane`\.' .agents/skills/hive-manager/SKILL.md internal/agent/templates/project-agents.md; echo "exit $?"
grep -nE '^## (Verifier spawn|Commands)$' .agents/skills/hive-manager/SKILL.md internal/agent/templates/project-agents.md; echo "exit $?"
grep -nE 'agent spawn.*(--project|--pane)' .agents/skills/hive-manager/SKILL.md internal/agent/templates/project-agents.md; echo "exit $?"
```

```text
commit f846ae75
.agents/skills/hive-manager/SKILL.md:61:hive agent spawn --role worker --task ID --name NAME --harness KIND --model MODEL
.agents/skills/hive-manager/SKILL.md:64:Hive takes the project from the task.
.agents/skills/hive-manager/SKILL.md:75:hive agent spawn --role verifier --task ID --name NAME --harness KIND --model MODEL
.agents/skills/hive-manager/SKILL.md:78:After a verify fail, spawn a fresh verifier with no `--pane`. Do not make placeholder tabs.
internal/agent/templates/project-agents.md:40:hive agent spawn --role worker --task ID --name NAME --harness KIND --model MODEL
internal/agent/templates/project-agents.md:41:hive agent spawn --role verifier --task ID --name NAME --harness KIND --model MODEL
internal/agent/templates/project-agents.md:52:Hive takes the project from the task.
internal/agent/templates/project-agents.md:54:After a verify fail, spawn a fresh verifier with no `--pane`. Do not make placeholder tabs.
exit 0
.agents/skills/hive-manager/SKILL.md:69:## Verifier spawn
internal/agent/templates/project-agents.md:33:## Commands
exit 0
exit 1
```

The last `grep` finds no spawn line with `--project` or `--pane` (exit 1). The section headings exist.

### V3 STE

```sh
echo "commit $(git rev-parse --short HEAD)"
go test -run TestSTE . -count=1; echo "exit $?"
python3 - <<'PY'
import json, subprocess
now=json.load(open('testdata/ste-baseline.json'))
base=json.loads(subprocess.check_output(['git','show','dev:testdata/ste-baseline.json']))
raised=[]
for k,v in now.items():
    if base.get(k,0) < v:
        raised.append((k, base.get(k,'absent'), v))
print('raised:', raised if raised else 'none')
print('added:', sorted(set(now)-set(base)))
print('removed:', sorted(set(base)-set(now)))
print('changed:', [(k,base[k],now[k]) for k in now if k in base and base[k]!=now[k]])
PY
echo "exit $?"
```

```text
commit f846ae75
ok  	github.com/Harrison-Blair/hive	0.194s
exit 0
raised: none
added: ['spec/hive-v1/tasks/17-verifier-project.md']
removed: []
changed: []
exit 0
```

I also compared the brief copy with the task record. The only differences are the three sentences that `V3.md` records as STE changes, and each change keeps its meaning:

```sh
hive --home /home/penguin/source/hive task show 17-verifier-project --brief > /tmp/hive17-brief.txt
diff -u /tmp/hive17-brief.txt spec/hive-v1/tasks/17-verifier-project.md; echo "diff exit $?"
```

```text
--- /tmp/hive17-brief.txt	2026-10-19 14:08:31.750537475 -0400
+++ spec/hive-v1/tasks/17-verifier-project.md	2026-10-19 14:04:23.324591323 -0400
@@ -26,7 +26,7 @@
 
 ## Known facts
 
-- `internal/agent/spawn.go`: `choosePane` returns at once when `--pane` is set.
+- `internal/agent/spawn.go`: `choosePane` returns at once when a user passes `--pane`.
   It does not read the task. Without `--pane`, it calls `resolveTaskPane` in
   `internal/agent/spawn_worker.go`.
 - `resolveTaskPane` reads the task and fails with `task ID is in project P`
@@ -61,15 +61,16 @@
 - A worker or verifier spawn with `--task` and no `--project` takes the
   project from the task record. This applies with `--pane` and without
   `--pane`.
-- When `--project` is given and differs from the task project, the spawn fails
-  with `task ID is in project P` and makes no agent record.
+- When a user gives a `--project` value that differs from the task project, the
+  spawn fails. It returns `task ID is in project P` and makes no agent record.
 - When the task does not exist, the spawn fails as today.
 
 Work test-first. Write each test, run it, and make sure that it fails for the
-correct reason before you change the code. The tests cover at least: a
-verifier spawn with `--pane` and no `--project`; a worker spawn with `--pane`
-and no `--project`; a spawn with `--pane` and a different `--project`. Show in
-the evidence that each new test fails against the old code, then passes.
+correct reason before you change the code. The tests cover these cases:
+1. A verifier spawn with `--pane` and no `--project`.
+2. A worker spawn with `--pane` and no `--project`.
+3. A spawn with `--pane` and a different `--project`.
+Show in the evidence that each new test fails against the old code, then passes.
 
 ### R2 Spawn command in the docs
 
diff exit 1
```

The zero-baseline linter also passes on the copy:

```sh
python3 /home/penguin/.agents/skills/asd-ste100/scripts/ste-lint.py --glossary /home/penguin/.agents/skills/asd-ste100/references/base-glossary.md --glossary STE-GLOSSARY.md spec/hive-v1/tasks/17-verifier-project.md; echo "exit $?"
```

```text
0 violations (0 hard, baseline 0), 664 words, 0.0 per 100 words
Hedges/modality (may, might, could) are never flagged: confidence is content.
exit 0
```

### V4 Changed files

The first run of `git diff dev --stat` was at `dev` `1a7a75a6` and showed only the eight source paths and the five evidence paths. `dev` then advanced to `0fa287c6` with one friction entry. I ran the item again and isolated the task change with the merge base.

```sh
echo "commit $(git rev-parse --short HEAD)"
git diff dev --stat; echo "exit $?"
git log --oneline 1a7a75a6..dev
git diff 1a7a75a6 --stat; echo "exit $?"
git diff dev...HEAD --stat; echo "exit $?"
```

```text
commit f846ae75
 .agents/skills/hive-manager/SKILL.md               |   6 +-
 evidence/17-verifier-project/V1.md                 |  72 +++++++++++++
 evidence/17-verifier-project/V2.md                 |  31 ++++++
 evidence/17-verifier-project/V3.md                 |  56 ++++++++++
 evidence/17-verifier-project/V4.md                 |  32 ++++++++
 evidence/17-verifier-project/V5.md                 |  88 +++++++++++++++
 internal/agent/manager_prompt_test.go              |   8 ++
 internal/agent/spawn.go                            |  11 ++
 internal/agent/spawn_project_test.go               |  90 ++++++++++++++++
 internal/agent/spawn_worker.go                     |  22 ++++
 internal/agent/templates/project-agents.md         |   6 +-
 reference/dogfood/friction/FRICTION.md             |   1 -
 .../verify-fail-repair-worker-no-workspace.md      |  29 -----
 spec/hive-v1/tasks/17-verifier-project.md          | 119 +++++++++++++++++++++
 testdata/ste-baseline.json                         |   3 +-
 15 files changed, 541 insertions(+), 33 deletions(-)
exit 0
0fa287c6 docs(dogfood): record the missing repair worker workspace
 .agents/skills/hive-manager/SKILL.md       |   6 +-
 evidence/17-verifier-project/V1.md         |  72 +++++++++++++++++
 evidence/17-verifier-project/V2.md         |  31 ++++++++
 evidence/17-verifier-project/V3.md         |  56 ++++++++++++++
 evidence/17-verifier-project/V4.md         |  32 ++++++++
 evidence/17-verifier-project/V5.md         |  88 +++++++++++++++++++
 internal/agent/manager_prompt_test.go      |   8 ++
 internal/agent/spawn.go                    |  11 +++
 internal/agent/spawn_project_test.go       |  90 ++++++++++++++++++++++
 internal/agent/spawn_worker.go             |  22 ++++++
 internal/agent/templates/project-agents.md |   6 +-
 spec/hive-v1/tasks/17-verifier-project.md  | 119 +++++++++++++++++++++++++++++
 testdata/ste-baseline.json                 |   3 +-
 13 files changed, 541 insertions(+), 3 deletions(-)
exit 0
 .agents/skills/hive-manager/SKILL.md       |   6 +-
 evidence/17-verifier-project/V1.md         |  72 +++++++++++++++++
 evidence/17-verifier-project/V2.md         |  31 ++++++++
 evidence/17-verifier-project/V3.md         |  56 ++++++++++++++
 evidence/17-verifier-project/V4.md         |  32 ++++++++
 evidence/17-verifier-project/V5.md         |  88 +++++++++++++++++++
 internal/agent/manager_prompt_test.go      |   8 ++
 internal/agent/spawn.go                    |  11 +++
 internal/agent/spawn_project_test.go       |  90 ++++++++++++++++++++++
 internal/agent/spawn_worker.go             |  22 ++++++
 internal/agent/templates/project-agents.md |   6 +-
 spec/hive-v1/tasks/17-verifier-project.md  | 119 +++++++++++++++++++++++++++++
 testdata/ste-baseline.json                 |   3 +-
 13 files changed, 541 insertions(+), 3 deletions(-)
exit 0
```

The eight non-evidence paths are the source table paths (the test glob covers `spawn_project_test.go`). The five `V*.md` paths are the evidence directory that the brief and the tracker require. The friction paths in the two-dot diff come from `dev` commit `0fa287c6`; the task branch does not touch the friction folder. The worktree was clean before this file was written.

### V5 Gate

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go'); echo "exit $?"
go vet ./...; echo "exit $?"
golangci-lint run ./...; echo "exit $?"
go test -race ./...; echo "exit $?"
```

```text
exit 0
exit 0
0 issues.
exit 0
ok  	github.com/Harrison-Blair/hive	3.913s
ok  	github.com/Harrison-Blair/hive/cmd	1.413s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.022s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.361s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.021s
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
ok  	github.com/Harrison-Blair/hive/internal/agent	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	3.029s
ok  	github.com/Harrison-Blair/hive/internal/handoff	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/land	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.027s
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
exit 0
```

I also forced the two changed packages without the cache:

```sh
go test -race ./internal/agent ./cmd/agent/spawn -count=1; echo "exit $?"
```

```text
ok  	github.com/Harrison-Blair/hive/internal/agent	2.874s
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.015s
exit 0
```

## Result

Each item V1 to V5 passes against commit `f846ae75`. The new tests fail by assertion on the base tree. The two instruction files carry the R2 and R3 text. The STE ratchet and the gate pass.
