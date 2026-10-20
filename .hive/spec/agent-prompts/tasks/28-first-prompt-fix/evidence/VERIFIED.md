# 28-first-prompt-fix VERIFIED

| Item | Value |
| --- | --- |
| Task | 28-first-prompt-fix |
| Commit | 20b75e72 |
| Base | dev ed33d8e0 |
| Date | 2026-10-19 |
| Agent | verifier, claude-code, claude-opus-5-5 |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | R1, R3 | pass | On a detached base worktree at `ed33d8e0` with the three task test files copied in, all six first-prompt tests fail for the expected reason (no retry, one prompt only, blocked not reported, old lost-prompt path). On the task branch all six pass. A mutation of the task code (`Until` = `[working]`, then timeout 5 s) makes all four wait assertions fail, so R3 item 4 detects a wrong wait. |
| V2 | R4 | pass | Branch binary built at `.hive/tmp/28-first-prompt-fix/hive`. Three Pi spawns (`openai-codex/gpt-6-luna:xhigh`, no `--pane`) each printed `agent v28-check-N live` with exit 0. Each pane shows the first prompt one time, the brief, and the reply `OK`, with no hand resend. Agents stopped, tasks cancelled, worktrees removed with `--force`, repository workspace `wEW` closed, `.hive/tmp/28-first-prompt-fix/` deleted. `herdr workspace list` shows no scratch workspace. |
| V3 | R2 | pass | `grep -n 'promptStarted\|promptStatusStarted' internal/agent/*.go` prints no match (exit 1). |
| V4 | scope | pass | `git diff dev --stat` shows only `internal/agent/spawn.go`, `internal/agent/spawn_worker.go`, tests beside them, `testdata/ste-baseline.json` (one new entry with 0), the brief copy, and the evidence folder. |
| V5 | gate | pass | `gofmt -l` prints nothing. `go vet ./...`, `golangci-lint run ./...` (0 issues), and `go test -race ./...` exit 0. |

## Runs

### V1 red (base ed33d8e0, task tests copied in)

```text
--- FAIL: TestWorkerFirstPromptRetriesOneStall (0.01s)
    first_prompt_test.go:68: outcome=partial error=agent_prompt_stalled: the prompt did not start; [hint trimmed]
--- FAIL: TestWorkerFirstPromptFailsAfterTwoStalls (0.01s)
    first_prompt_test.go:96: agent.prompt requests = 1; want exactly two
--- FAIL: TestWorkerFirstPromptBlockedDoesNotRetry (0.01s)
    first_prompt_test.go:122: error = agent_blocked: trust dialog is open; [hint trimmed]
--- FAIL: TestManagerPromptSendsGeneratedInstructionsAfterReady (0.00s)
    manager_prompt_test.go:40: outcome=partial error=manager prompt: agent_prompt_stalled: first manager prompt stalled; then run hive agent send --agent manager-demo; want ok
--- FAIL: TestWorkerSpawnPromptFailureIsPartial (0.01s)
    worker_spawn_test.go:549: agent.prompt requests = 1; want two
--- FAIL: TestWorkerSpawnResendsLostPrompt (0.01s)
    fakeherdr_test.go:147: unexpected method "agent.get"
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.040s
exit 1
```

### V1 green (task branch 20b75e72)

```text
--- PASS: TestWorkerFirstPromptRetriesOneStall (0.01s)
--- PASS: TestWorkerFirstPromptFailsAfterTwoStalls (0.01s)
--- PASS: TestWorkerFirstPromptBlockedDoesNotRetry (0.01s)
--- PASS: TestManagerPromptSendsGeneratedInstructionsAfterReady (0.00s)
--- PASS: TestWorkerSpawnPromptFailureIsPartial (0.01s)
--- PASS: TestWorkerSpawnResendsLostPrompt (0.01s)
ok  	github.com/Harrison-Blair/hive/internal/agent	0.039s
exit 0
```

### V1 wait mutation (task code, scratch copy)

```text
until = [working]:   first_prompt_test.go:74/98/128, manager_prompt_test.go:46: prompt 1 wait.until = [working]; want [working blocked]
timeout = 5 s:       first_prompt_test.go:74/98/128, manager_prompt_test.go:46: prompt 1 wait.timeout_ms = 5000; want 10000
```

### V2 live check

```text
commit 20b75e72
sha256 0eb5b788c8949b154109f3413d4ff4cceffaaac44e489a4ed9d82cfa63f05469  .hive/tmp/28-first-prompt-fix/hive
project register ... --id prompt-check --base dev   -> created record .hive/state/projects/prompt-check.json
run 1: task 1-vp-1, workspace wEX (repository workspace wEW opened), spawn v28-check-1 -> "agent v28-check-1 live", exit 0
run 2: task 2-vp-2, workspace wEY, spawn v28-check-2 -> "agent v28-check-2 live", exit 0
run 3: task 3-vp-3, workspace wEZ, spawn v28-check-3 -> "agent v28-check-3 live", exit 0
herdr agent wait <pane> --until done; herdr agent read <pane>:
  wEX:p1 status done; prompt "Read .../hive-worker/SKILL.md first. ..." count 1; "# Live prompt check"; reply "OK"
  wEY:p1 status done; prompt count 1; "# Live prompt check"; reply "OK"
  wEZ:p1 status done; prompt count 1; "# Live prompt check"; reply "OK"
cleanup:
  agent stop v28-check-1..3 -> closed pane wEX:p1 / wEY:p1 / wEZ:p1, workspace closed
  worktree remove (assigned) -> "task is assigned: complete, verify, or cancel it first"
  task cancel 1-vp-1..3-vp-3 -> status: cancelled
  worktree remove --force -> worktree removed (x3)
  herdr workspace close wEW -> ok
  rm -rf .hive/tmp/28-first-prompt-fix
herdr workspace list labels: agents, hive, hive:qmeter, hive:stenographer, 27-ci-flake-fix, 28-first-prompt-fix
```

### V3

```text
$ grep -n 'promptStarted\|promptStatusStarted' internal/agent/*.go
exit 1
```

### V4

```text
 internal/agent/first_prompt_test.go             | 129 +++++++
 internal/agent/manager_prompt_test.go           |  22 +-
 internal/agent/spawn.go                         |  38 +-
 internal/agent/spawn_worker.go                  | 105 +++---
 internal/agent/worker_spawn_test.go             |  17 +-
 spec/hive-v1/evidence/28-first-prompt-fix/V1.md |  57 +++
 spec/hive-v1/evidence/28-first-prompt-fix/V2.md | 476 ++++++++++++++++++++++++
 spec/hive-v1/evidence/28-first-prompt-fix/V3.md |  27 ++
 spec/hive-v1/evidence/28-first-prompt-fix/V4.md |  45 +++
 spec/hive-v1/evidence/28-first-prompt-fix/V5.md | 109 ++++++
 spec/hive-v1/tasks/28-first-prompt-fix.md       | 129 +++++++
 testdata/ste-baseline.json                      |   3 +-
 12 files changed, 1078 insertions(+), 79 deletions(-)
```

### V5

```text
gofmt exit 0 (no output)
vet exit 0
0 issues.
lint exit 0
go test -race ./... : every package ok; test exit 0
```
