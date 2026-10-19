# 18-task-reopen VERIFIED

| Item | Value |
| --- | --- |
| Task | 18-task-reopen |
| Commit | bbb9d0b8 |
| Date | 2026-10-19 |
| Agent | verifier, pi, opencode-go/deepseek-v4.1-flash:max |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | R1, R2, R3, R4, R5, R6 | pass | Each new test fails by assertion on base `39a35e6c` and passes on `bbb9d0b8`. The repair test also fails on the pre-repair task commit `30c427a4`. |
| V2 | R1, R2, R3 | pass | A registered manager (`manager-demo-g1`, claim key `manager-demo`) reopened a verified task. Show printed `assigned`, `verification: -`, and the old check in the history. Land failed with `task is assigned`. A worker was refused and the task stayed `verified`. |
| V3 | R4 | pass | Both instruction files contain the exact merge-conflict repair rule. |
| V4 | R1, R2, R3, R4, R6 | pass | Every changed path is in the source table, is a test beside a changed file, or is task evidence. |
| V5 | R1, R2, R3, R4, R5, R6 | pass | `gofmt -l` printed nothing. `go vet`, `golangci-lint`, and `go test -race` exited 0. |

## Runs

The worktree is `/home/penguin/source/hive/projects/hive/worktrees/18-task-reopen`.
The checkout was clean before this file was written. Scratch trees were below
`/home/penguin/source/hive/.hive/tmp/` and were removed after each run.

### V1 Red run

The base tree was `git archive 39a35e6c` with the current test files copied
in. `internal/task/reopen.go` was an empty `Reopen` scaffold, so the unit red
failures are assertions. The copied `internal/lib/records/task.go` gives the
tests the history type. The base has no reopen command or land verification
guard.

```sh
set +e
BASE=$(mktemp -d /home/penguin/source/hive/.hive/tmp/18-verify2-base.XXXXXX)
git archive 39a35e6c | tar -x -C "$BASE"
mkdir -p "$BASE/cmd" "$BASE/internal/agent" "$BASE/internal/land" \
  "$BASE/internal/lib/records" "$BASE/internal/task"
cp internal/task/reopen_test.go "$BASE/internal/task/reopen_test.go"
cp internal/lib/records/task.go "$BASE/internal/lib/records/task.go"
printf 'package task\n\nimport "github.com/Harrison-Blair/hive/internal/lib/output"\n\nfunc Reopen(home string, options TransitionOptions) output.Envelope { return output.Envelope{} }\n' \
  > "$BASE/internal/task/reopen.go"
cp internal/land/preconditions_test.go "$BASE/internal/land/preconditions_test.go"
cp internal/land/helpers_test.go "$BASE/internal/land/helpers_test.go"
cp internal/agent/manager_prompt_test.go "$BASE/internal/agent/manager_prompt_test.go"
cp cmd/task_lifecycle_test.go "$BASE/cmd/task_lifecycle_test.go"
cd "$BASE"
echo "base commit 39a35e6c"
go test ./internal/agent -run '^(TestManagerInstructionsRequireFreshVerifierAfterFail|TestManagerInstructionsRequireRespawnRules)$' -count=1; echo "exit $?"
go test ./internal/task -run '^TestReopen' -count=1; echo "exit $?"
go test ./cmd -run '^(TestTaskCLIReopenTextJSONShowHistoryAndLog|TestTaskCLIReopenRegisteredManagerUsesClaimKey)$' -count=1; echo "exit $?"
go test ./internal/land -run '^TestReopenedTaskRequiresFreshPassVerificationBeforeLand$' -count=1; echo "exit $?"
```

```text
base commit 39a35e6c
--- FAIL: TestManagerInstructionsRequireRespawnRules (0.00s)
    manager_prompt_test.go:102: template does not contain "If `hive land` fails with a merge conflict, reopen the task with `hive task reopen`. Spawn a repair worker. Then spawn a fresh verifier with no `--pane`."
    manager_prompt_test.go:102: skill does not contain "If `hive land` fails with a merge conflict, reopen the task with `hive task reopen`. Spawn a repair worker. Then spawn a fresh verifier with no `--pane`."
FAIL
exit 1
--- FAIL: TestManagerInstructionsRequireFreshVerifierAfterFail (0.00s)
    manager_prompt_test.go:119: template does not contain "If `hive land` fails with a merge conflict, reopen the task with `hive task reopen`. Spawn a repair worker. Then spawn a fresh verifier with no `--pane`."
    manager_prompt_test.go:119: skill does not contain "If `hive land` fails with a merge conflict, reopen the task with `hive task reopen`. Spawn a repair worker. Then spawn a fresh verifier with no `--pane`."
FAIL
exit 1
--- FAIL: TestReopenVerifiedByControllingRole (0.00s)
    reopen_test.go:46: Reopen() = output.Envelope{...}; want ok
--- FAIL: TestReopenByWorkerFails (0.00s)
    reopen_test.go:74: Reopen() = output.Envelope{...}; want controlling-role refusal
--- FAIL: TestReopenNonVerifiedStatusesFail (0.00s)
    --- FAIL: TestReopenNonVerifiedStatusesFail/created
    --- FAIL: TestReopenNonVerifiedStatusesFail/assigned
    --- FAIL: TestReopenNonVerifiedStatusesFail/completed
    --- FAIL: TestReopenNonVerifiedStatusesFail/cancelled
--- FAIL: TestReopenRequiresReason (0.00s)
    reopen_test.go:108: Reopen() = output.Envelope{...}; want required reason error
--- FAIL: TestReopenKeepsVerificationHistoryAcrossReopens (0.00s)
    reopen_test.go:123: first Reopen() = output.Envelope{...}; want ok
FAIL
exit 1
--- FAIL: TestTaskCLIReopenTextJSONShowHistoryAndLog (0.00s)
    task_lifecycle_test.go:239: reopen without reason: status 2, stdout "", stderr "hive: unknown flag: --task\n"; want a reason error
--- FAIL: TestTaskCLIReopenRegisteredManagerUsesClaimKey/agent (0.00s)
    task_lifecycle_test.go:334: reopen status = 2; stdout "", stderr "hive: unknown flag: --task\n"; want success
--- FAIL: TestTaskCLIReopenRegisteredManagerUsesClaimKey/pane (0.00s)
    task_lifecycle_test.go:334: reopen status = 2; stdout "", stderr "hive: unknown flag: --task\n"; want success
FAIL
exit 1
--- FAIL: TestReopenedTaskRequiresFreshPassVerificationBeforeLand/verified_without_a_check
    preconditions_test.go:163: Land() = {Outcome:ok ...}; want a missing verification error
--- FAIL: TestReopenedTaskRequiresFreshPassVerificationBeforeLand/verified_after_a_failed_check
    preconditions_test.go:176: Land() = {Outcome:ok ...}; want a failed verification error
FAIL
exit 1
```

The new R4 test fails on the base only because the sentence is absent. The
augmented existing R4 test also asserts that sentence; on this old base it
also reports earlier instructions that the base predates.

The registered-manager regression has a second, more precise red. The current
test file was copied into a tree of the pre-repair task commit `30c427a4`.
That tree has the reopen command and fails on the old agent-key comparison.

```sh
set +e
PRE=$(mktemp -d /home/penguin/source/hive/.hive/tmp/18-verify2-pre.XXXXXX)
git archive 30c427a4 | tar -x -C "$PRE"
cp cmd/task_lifecycle_test.go "$PRE/cmd/task_lifecycle_test.go"
cd "$PRE"
go test ./cmd -run '^TestTaskCLIReopenRegisteredManagerUsesClaimKey$' -count=1; echo "exit $?"
```

```text
--- FAIL: TestTaskCLIReopenRegisteredManagerUsesClaimKey/agent (0.00s)
    task_lifecycle_test.go:334: reopen status = 1; stdout "", stderr "hive: not the controlling role: manager-repo\n"; want success
--- FAIL: TestTaskCLIReopenRegisteredManagerUsesClaimKey/pane (0.00s)
    task_lifecycle_test.go:334: reopen status = 1; stdout "", stderr "hive: not the controlling role: manager-repo\n"; want success
FAIL
exit 1
```

### V1 Green run

```sh
git rev-parse --short HEAD
go test ./internal/agent -run '^(TestManagerInstructionsRequireFreshVerifierAfterFail|TestManagerInstructionsRequireRespawnRules)$' -count=1; echo "exit $?"
go test ./internal/task -run '^TestReopen' -count=1; echo "exit $?"
go test ./cmd -run '^(TestTaskCLIReopenTextJSONShowHistoryAndLog|TestTaskCLIReopenRegisteredManagerUsesClaimKey)$' -count=1; echo "exit $?"
go test ./internal/land -run '^TestReopenedTaskRequiresFreshPassVerificationBeforeLand$' -count=1; echo "exit $?"
go test . -run '^TestSTERatchet$' -count=1; echo "exit $?"
```

```text
bbb9d0b8
ok  	github.com/Harrison-Blair/hive/internal/agent	0.001s
exit 0
ok  	github.com/Harrison-Blair/hive/internal/task	0.004s
exit 0
ok  	github.com/Harrison-Blair/hive/cmd	0.005s
exit 0
ok  	github.com/Harrison-Blair/hive/internal/land	0.026s
exit 0
ok  	github.com/Harrison-Blair/hive	0.204s
exit 0
```

### R6 brief-copy red and green

The copy was checked at `c3c7e1ee` with the two STE-safe passages replaced by
the original brief wording. `TestSTERatchet` reported six violations at
baseline 0. The committed copy passes at `bbb9d0b8`.

```text
--- R6 red on the original brief wording ---
spec/hive-v1/tasks/18-task-reopen.md: 6 violations, baseline 0; fix the new ones
  spec/hive-v1/tasks/18-task-reopen.md:60 passive-voice: is required
  spec/hive-v1/tasks/18-task-reopen.md:99 semicolon: ;
  spec/hive-v1/tasks/18-task-reopen.md:99 semicolon: ;
  spec/hive-v1/tasks/18-task-reopen.md:100 semicolon: ;
  spec/hive-v1/tasks/18-task-reopen.md:100 semicolon: ;
  spec/hive-v1/tasks/18-task-reopen.md:101 semicolon: ;
FAIL
exit 1

--- R6 green on the committed copy ---
ok  	github.com/Harrison-Blair/hive	0.204s
exit 0
```

The two recorded copy changes are the passive `--reason` sentence and the
semicolon test list. The meaning is unchanged.

### V2 Live check

The scratch home used the real `agent register` command. The manager has agent
record key `manager-demo-g1`, claim key `manager-demo`, and pane `w1:manager`.
The task fixture is `verified` with an old pass by `verifier-old`.

```sh
set +e
RUN=$(mktemp -d /home/penguin/source/hive/.hive/tmp/18-verify2-live.XXXXXX)
HOME_DIR="$RUN/home"
REPO="$RUN/repo"
BIN="$RUN/hive"
mkdir -p "$HOME_DIR/.hive" "$REPO"
printf 'Hive test home\n' > "$HOME_DIR/.hive/README.md"
git init -q "$REPO"
go build -o "$BIN" .; echo "build exit $?"
"$BIN" --home "$HOME_DIR" project register --id demo --base main "$REPO"; echo "project register exit $?"
"$BIN" --home "$HOME_DIR" agent register --role manager --project demo --name manager-demo \
  --pane w1:manager --harness pi --model test; echo "manager register exit $?"
"$BIN" --home "$HOME_DIR" task create --project demo --title 'Reopen live check' --slug reopen-live; echo "task create exit $?"
python3 - "$HOME_DIR" <<'PY'
import json, pathlib, sys
path = pathlib.Path(sys.argv[1]) / ".hive/state/tasks/1-reopen-live.json"
task = json.loads(path.read_text())
task["status"] = "verified"
task["verification"] = {"agent": "verifier-old", "result": "pass", "reason": "", "at": "2026-10-19T00:00:00Z"}
path.write_text(json.dumps(task, separators=(",", ":")) + "\n")
PY
"$BIN" --home "$HOME_DIR" task reopen --task 1-reopen-live --reason 'merge conflict' --agent manager-demo; echo "reopen exit $?"
"$BIN" --home "$HOME_DIR" task show 1-reopen-live; echo "show exit $?"
"$BIN" --home "$HOME_DIR" land --task 1-reopen-live --mode merge --grant 'live check'; echo "land exit $?"
"$BIN" --home "$HOME_DIR" task reopen --task 1-reopen-live --reason 'second' --agent manager-demo; echo "reopen again exit $?"
rm -rf "$RUN"; test ! -e "$RUN"; echo "scratch cleanup exit $?"
```

```text
build exit 0
project register exit 0
created agent .hive/state/agents/manager-demo-g1.json
updated claim .hive/state/claims/manager-demo.json
manager register exit 0
task create exit 0
id: 1-reopen-live
status: assigned
reopen exit 0
id: 1-reopen-live
project: demo
title: Reopen live check
controlling_role: manager-demo
acting_agent: 
acting_agent_status: no record
status: assigned
depends_on: 
verification: -
verification_history: [{"verification":{"agent":"verifier-old","result":"pass","reason":"","at":"2026-10-19T00:00:00Z"},"reason":"merge conflict","caller":"manager-demo","at":"2026-10-19T19:10:53.810847295Z"}]
worktree: -
delivery: -
cancel_reason: 
created_at: 2026-10-19T19:10:53.793228777Z
updated_at: 2026-10-19T19:10:53.810847816Z
brief: .hive/state/briefs/1-reopen-live.md
show exit 0
hive: task is assigned: land needs verified
land exit 1
hive: invalid transition: assigned -> reopen
reopen again exit 1
scratch cleanup exit 0
```

The same home also checked the other R1 refusals and the pane lookup:

```text
--- empty reason on a fresh verified task ---
hive: --reason is required
exit 1

--- worker caller on a fresh verified task ---
hive: not the controlling role: worker-demo
exit 1
status: verified
verification: pass by verifier-old at 2026-10-19T00:00:00Z
verification_history: []

--- manager by pane ---
id: 4-pane-reopen
status: assigned
exit 0
```

The worker's record has no task claim key. Hive resolved it, then refused it by
role. The task record did not change.

### V3 Instruction text

```sh
rg -n -F 'If `hive land` fails with a merge conflict, reopen the task with `hive task reopen`. Spawn a repair worker. Then spawn a fresh verifier with no `--pane`.' \
  .agents/skills/hive-manager/SKILL.md internal/agent/templates/project-agents.md
echo "exit $?"
```

```text
.agents/skills/hive-manager/SKILL.md:81:If `hive land` fails with a merge conflict, reopen the task with `hive task reopen`. Spawn a repair worker. Then spawn a fresh verifier with no `--pane`.
internal/agent/templates/project-agents.md:58:If `hive land` fails with a merge conflict, reopen the task with `hive task reopen`. Spawn a repair worker. Then spawn a fresh verifier with no `--pane`.
exit 0
```

### V4 Change scope

```sh
git rev-parse --short HEAD
git diff dev --stat
git diff dev --name-status
git status --short
```

```text
bbb9d0b8
 .agents/skills/hive-manager/SKILL.md       |   2 +
 cmd/task/reopen/reopen.go                  |  35 ++
 cmd/task/task.go                           |   2 +
 cmd/task_lifecycle_test.go                 | 154 +++++++++
 evidence/18-task-reopen/V1.md              | 302 +++++++++++++++++
 evidence/18-task-reopen/V2.md              | 100 ++++++
 evidence/18-task-reopen/V3.md              |  30 ++
 evidence/18-task-reopen/V4.md              |  54 +++
 evidence/18-task-reopen/V5.md              |  98 ++++++
 evidence/18-task-reopen/VERIFIED.md        | 498 +++++++++++++++++++++++++++++
 internal/agent/manager_prompt_test.go      |  14 +
 internal/agent/templates/project-agents.md |   2 +
 internal/land/helpers_test.go              |   1 +
 internal/land/land.go                      |   3 +
 internal/land/preconditions_test.go        |  56 ++++
 internal/lib/records/task.go               |  33 +-
 internal/task/create.go                    |  19 +-
 internal/task/doc.go                       |   6 +-
 internal/task/log.go                       |   3 +-
 internal/task/render.go                    |   9 +
 internal/task/reopen.go                    |  12 +
 internal/task/reopen_test.go               | 157 +++++++++
 internal/task/transition.go                |  67 +++-
 spec/hive-v1/tasks/18-task-reopen.md       | 134 ++++++++
 testdata/ste-baseline.json                 |   3 +-
25 files changed
```

The `A` paths are `cmd/task/reopen/reopen.go`, the five `V<n>.md` files,
`VERIFIED.md`, `internal/task/reopen.go`, `internal/task/reopen_test.go`, and
`spec/hive-v1/tasks/18-task-reopen.md`. Each other path is modified. There are
no deletions.

The paths map to the source table:

- `cmd/task/reopen/`, `cmd/task/task.go`, and `internal/task/` are the R1 and
  R2 rows. The `internal/task/` files support the reopen command, the history
  render, and the reason log.
- `internal/lib/records/task.go` is the R2 history field.
- `internal/land/land.go` is the R3 verification guard.
- The tests beside those files are `cmd/task_lifecycle_test.go`,
  `internal/task/reopen_test.go`, `internal/land/preconditions_test.go`, and
  `internal/land/helpers_test.go`.
- `.agents/skills/hive-manager/SKILL.md`,
  `internal/agent/templates/project-agents.md`, and
  `internal/agent/manager_prompt_test.go` are R4.
- `testdata/ste-baseline.json` adds only the new brief copy with count 0.
- `spec/hive-v1/tasks/18-task-reopen.md` is the R6 copy.
- `evidence/18-task-reopen/` is the evidence directory.

The worktree was clean after the temporary trees were removed. This file is
the verifier output and is not committed by this verifier.

### V5 Gate

```sh
set +e
git rev-parse --short HEAD
gofmt -l $(git ls-files -co --exclude-standard '*.go'); echo "exit $?"
go vet ./...; echo "exit $?"
golangci-lint run ./...; echo "exit $?"
go test -race ./...; echo "exit $?"
exit 0
```

```text
bbb9d0b8
exit 0
exit 0
0 issues.
exit 0
ok  	github.com/Harrison-Blair/hive	4.071s
ok  	github.com/Harrison-Blair/hive/cmd	1.532s
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.021s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.321s
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.020s
ok  	github.com/Harrison-Blair/hive/cmd/version	1.009s
ok  	github.com/Harrison-Blair/hive/internal/agent	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	3.028s
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
ok  	github.com/Harrison-Blair/hive/internal/lib/process	(cached)
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

The changed packages were also run without the test cache:

```sh
go test -race -count=1 ./internal/task ./internal/land ./internal/agent ./cmd; echo "exit $?"
go test -race -count=1 . -run '^TestSTERatchet$'; echo "exit $?"
```

```text
ok  	github.com/Harrison-Blair/hive/internal/task	1.580s
ok  	github.com/Harrison-Blair/hive/internal/land	1.953s
ok  	github.com/Harrison-Blair/hive/internal/agent	2.880s
ok  	github.com/Harrison-Blair/hive/cmd	1.522s
exit 0
ok  	github.com/Harrison-Blair/hive	4.054s
exit 0
```

## Verdict

V1 through V5 pass. The repair to the caller check accepts a manager with the
real registration shape (`manager-demo-g1` record, `manager-demo` claim key).
The reopened task is `assigned`, its current verification is null, the old
check is in `verification_history`, and land refuses it until a new pass.
