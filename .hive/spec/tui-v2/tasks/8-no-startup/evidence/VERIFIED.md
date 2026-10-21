# Independent verification: 8-no-startup

Verifier: `t8-verifier-1`. Date: 2026-10-21.
Task revision: `a77b33ae`. Verdict: pass.

| Item | Result | Evidence |
| --- | --- | --- |
| V1 Tests | pass | The new flag and variable tests fail on base `dev`. The startup log test also fails on base. All pass on the task code. The skill test fails on base and passes in the gate. The negative control for other variable values passes on base. A separate mutation makes it fail. |
| V2 Live check | pass | Both forms open the TUI in a scratch home in workspace `w69`. The workspace keeps only tab `w69:t1`, with label `1`. Agent lists are empty during both runs. Each run logs three skipped steps with reason `no-startup`. The workspace is closed. |
| V3 Instruction text | pass | `go test ./ -run 'STE|Skill'` passes. Both skills require `hive --no-startup` or `HIVE_NO_STARTUP=1 hive`. The friction entry names both forms and stays Open. |
| V-gate Diff scope and gate | pass | `git merge dev --no-edit` reports `Already up to date.` The diff contains only the allowed source, tests, brief, and evidence. All five gate commands exit 0. The format and module commands print nothing. |

## Method and limits

The base run uses an archive of local `dev` with the new test files from the task branch. No source file in the task checkout was changed. The mutation uses a second archive of the task code. It changes the exact-value check from `== "1"` to `!= ""`. `TestBareHiveOtherNoStartupValueRunsStartup` then fails. This test is a negative control: base code already ignores `HIVE_NO_STARTUP=true`, so a base failure is not expected for it.

The live binary was built from the task branch. The scratch home and XDG config were under `.hive/tmp/t8-verifier/`. The new pane had `HERDR_ENV=1` and `HERDR_PANE_ID=w69:p1`. Each run used its own switch or variable. Each run ended with two Ctrl+C keys. The live transcript includes a chat access error because the scratch working directory is below `projects/`. This error does not prevent TUI display or change startup results. Chat behavior is outside this task.

This task changes startup effects and adds no screen rendering. A board comparison does not apply. `internal/tui/test_main_test.go` only clears the new environment variable for test isolation; it is a test beside the allowed startup source.

The scratch workspace was closed. A subsequent lookup reports `workspace_not_found`. Scratch files are removed after this report is complete.

## V1 Base run: go test ./cmd ./internal/tui -run NoStartup -v; go test ./ -run TestWorkerAndVerifierRequireNoStartup -v

```text
=== RUN   TestBareHiveNoStartupMakesNoTabCalls
=== RUN   TestBareHiveNoStartupMakesNoTabCalls/flag
    root_test.go:142: status = 2; want 0; output "hive: unknown flag: --no-startup\n"
=== RUN   TestBareHiveNoStartupMakesNoTabCalls/variable
    root_test.go:149: herdr methods = [ping pane.get tab.rename tab.create tab.create]; want no tab or input call
--- FAIL: TestBareHiveNoStartupMakesNoTabCalls (0.02s)
    --- FAIL: TestBareHiveNoStartupMakesNoTabCalls/flag (0.00s)
    --- FAIL: TestBareHiveNoStartupMakesNoTabCalls/variable (0.02s)
=== RUN   TestBareHiveOtherNoStartupValueRunsStartup
--- PASS: TestBareHiveOtherNoStartupValueRunsStartup (0.02s)
FAIL
FAIL	github.com/Harrison-Blair/hive/cmd	0.052s
=== RUN   TestStartupNoStartupSkipsEachStepAndLogsReason
    startup_test.go:181: no-startup started a coordinator
--- FAIL: TestStartupNoStartupSkipsEachStepAndLogsReason (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui	0.004s
FAIL
=== RUN   TestWorkerAndVerifierRequireNoStartup
    skills_test.go:266: .agents/skills/hive-worker/SKILL.md does not contain "hive --no-startup"
    skills_test.go:266: .agents/skills/hive-worker/SKILL.md does not contain "HIVE_NO_STARTUP=1 hive"
    skills_test.go:266: .agents/skills/hive-verifier/SKILL.md does not contain "hive --no-startup"
    skills_test.go:266: .agents/skills/hive-verifier/SKILL.md does not contain "HIVE_NO_STARTUP=1 hive"
--- FAIL: TestWorkerAndVerifierRequireNoStartup (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive	0.003s
FAIL
```

## V1 Task run: go test ./cmd ./internal/tui -run NoStartup -count=1 -v

```text
=== RUN   TestBareHiveNoStartupMakesNoTabCalls
=== RUN   TestBareHiveNoStartupMakesNoTabCalls/flag
=== RUN   TestBareHiveNoStartupMakesNoTabCalls/variable
--- PASS: TestBareHiveNoStartupMakesNoTabCalls (0.01s)
    --- PASS: TestBareHiveNoStartupMakesNoTabCalls/flag (0.01s)
    --- PASS: TestBareHiveNoStartupMakesNoTabCalls/variable (0.01s)
=== RUN   TestBareHiveOtherNoStartupValueRunsStartup
--- PASS: TestBareHiveOtherNoStartupValueRunsStartup (0.01s)
PASS
ok  	github.com/Harrison-Blair/hive/cmd	0.022s
=== RUN   TestStartupNoStartupSkipsEachStepAndLogsReason
--- PASS: TestStartupNoStartupSkipsEachStepAndLogsReason (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/tui	0.003s
```

## V1 Mutation: go test ./cmd -run OtherNoStartup -count=1 -v

```text
=== RUN   TestBareHiveOtherNoStartupValueRunsStartup
    root_test.go:171: herdr methods = []; want startup to rename the tab
--- FAIL: TestBareHiveOtherNoStartupValueRunsStartup (0.01s)
FAIL
FAIL	github.com/Harrison-Blair/hive/cmd	0.009s
FAIL
```

## V2 Live transcript

```text
FLAG RUN
{"id":"cli:tab:list","result":{"tabs":[{"agent_status":"unknown","focused":false,"label":"1","number":1,"pane_count":1,"tab_id":"w69:t1","workspace_id":"w69"}],"type":"tab_list"}}
│                            ││                                                                                                                                          ││                                  │
│                            ││                                                                                                                                          ││                                  │
╰────────────────────────────╯╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯╰──────────────────────────────────╯
error: chat send is for the user only: only the user can send chat; rule 2 (working directory is below projects/)

i write · pgup/pgdown scroll · e expand · n noise · 1–8 screens · ctrl+c×2 quit
{"outcome":"ok","phase":"complete","effects":[],"error":null,"result":[]}
VARIABLE RUN
{"id":"cli:tab:list","result":{"tabs":[{"agent_status":"unknown","focused":false,"label":"1","number":1,"pane_count":1,"tab_id":"w69:t1","workspace_id":"w69"}],"type":"tab_list"}}
│                            ││                                                                                                                                          ││                                  │
│                            ││                                                                                                                                          ││                                  │
╰────────────────────────────╯╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯╰──────────────────────────────────╯
error: chat send is for the user only: only the user can send chat; rule 2 (working directory is below projects/)

i write · pgup/pgdown scroll · e expand · n noise · 1–8 screens · ctrl+c×2 quit
{"outcome":"ok","phase":"complete","effects":[],"error":null,"result":[]}
AFTER BOTH RUNS
{"id":"cli:tab:list","result":{"tabs":[{"agent_status":"unknown","focused":false,"label":"1","number":1,"pane_count":1,"tab_id":"w69:t1","workspace_id":"w69"}],"type":"tab_list"}}
XDG_CONFIG_HOME=/home/penguin/source/hive/projects/hive/worktrees/8-no-startup/.hive/tmp/t8-verifier/xdg
HIVE_HOME=/home/penguin/source/hive/projects/hive/worktrees/8-no-startup/.hive/tmp/t8-verifier/home
HERDR_ENV=1
penguin@raft:~/source/hive/projects/hive/worktrees/8-no-startup/.hive/tmp/t8-verifier/home$ HIVE_NO_STARTUP=1 '/home/penguin/source/hive/projects/hive/worktrees/8-no-startup/.hive/tmp/t8-verifier/hive'
penguin@raft:~/source/hive/projects/hive/worktrees/8-no-startup/.hive/tmp/t8-verifier/home$
{"command":"bare hive","component":"tui","level":"DEBUG","msg":"startup step","outcome":"skipped","reason":"no-startup","request_id":"r-92d798fa","step":"current tab label","time":"2026-10-21T18:15:12.416805435-04:00"}
{"command":"bare hive","component":"tui","level":"DEBUG","msg":"startup step","outcome":"skipped","reason":"no-startup","request_id":"r-92d798fa","step":"coordinator","time":"2026-10-21T18:15:12.416845105-04:00"}
{"command":"bare hive","component":"tui","level":"DEBUG","msg":"startup step","outcome":"skipped","reason":"no-startup","request_id":"r-92d798fa","step":"supervisor","time":"2026-10-21T18:15:12.416855837-04:00"}
{"command":"bare hive","component":"tui","level":"DEBUG","msg":"startup step","outcome":"skipped","reason":"no-startup","request_id":"r-b9f2af8c","step":"current tab label","time":"2026-10-21T18:15:20.938899069-04:00"}
{"command":"bare hive","component":"tui","level":"DEBUG","msg":"startup step","outcome":"skipped","reason":"no-startup","request_id":"r-b9f2af8c","step":"coordinator","time":"2026-10-21T18:15:20.938965656-04:00"}
{"command":"bare hive","component":"tui","level":"DEBUG","msg":"startup step","outcome":"skipped","reason":"no-startup","request_id":"r-b9f2af8c","step":"supervisor","time":"2026-10-21T18:15:20.938982788-04:00"}
{"id":"cli:workspace:get","result":{"type":"workspace_info","workspace":{"active_tab_id":"w69:t1","agent_status":"unknown","focused":false,"label":"t8-verifier-scratch","number":5,"pane_count":1,"tab_count":1,"workspace_id":"w69"}}}
{"id":"cli:workspace:close","result":{"type":"ok"}}
{"error":{"code":"workspace_not_found","message":"workspace w69 not found"},"id":"cli:workspace:get"}
```

## V3 go test ./ -run 'STE|Skill'

```text
ok  	github.com/Harrison-Blair/hive	0.679s
```

## V-gate git diff dev...HEAD --stat

```text
 .agents/skills/hive-verifier/SKILL.md              |   3 +
 .agents/skills/hive-worker/SKILL.md                |   3 +
 .hive/spec/tui-v2/tasks/8-no-startup/brief.md      | 112 ++++++++++++++++++++
 .../tui-v2/tasks/8-no-startup/evidence/V-gate.md   | 117 +++++++++++++++++++++
 .../spec/tui-v2/tasks/8-no-startup/evidence/V1.md  |  53 ++++++++++
 .../spec/tui-v2/tasks/8-no-startup/evidence/V2.md  |  80 ++++++++++++++
 .../spec/tui-v2/tasks/8-no-startup/evidence/V3.md  |  70 ++++++++++++
 cmd/root.go                                        |  11 ++
 cmd/root_test.go                                   | 113 +++++++++++++++++++-
 internal/tui/startup.go                            |  10 +-
 internal/tui/startup_test.go                       |  28 +++++
 internal/tui/test_main_test.go                     |   2 +-
 reference/dogfood/friction/FRICTION.md             |   2 +-
 .../entries/scratch-hive-starts-real-agents.md     |  10 +-
 skills_test.go                                     |  16 +++
 testdata/ste-baseline.json                         |   1 +
 16 files changed, 625 insertions(+), 6 deletions(-)
```

## V-gate Full output

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
exit=0
$ go mod tidy -diff
exit=0
$ go vet ./...
exit=0
$ golangci-lint run ./...
0 issues.
exit=0
$ go test -race ./...
ok  	github.com/Harrison-Blair/hive	14.496s
ok  	github.com/Harrison-Blair/hive/cmd	3.427s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.142s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.345s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.274s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.118s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.853s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.836s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.269s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.056s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.124s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.024s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	10.742s
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.819s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	(cached)
ok  	github.com/Harrison-Blair/hive/internal/homeinit	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/land	3.439s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.191s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.281s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/log	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.015s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.235s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/store	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/logs	(cached)
ok  	github.com/Harrison-Blair/hive/internal/memory	(cached)
ok  	github.com/Harrison-Blair/hive/internal/project	(cached)
ok  	github.com/Harrison-Blair/hive/internal/supervise	(cached)
ok  	github.com/Harrison-Blair/hive/internal/task	9.210s
ok  	github.com/Harrison-Blair/hive/internal/tui	49.637s
ok  	github.com/Harrison-Blair/hive/internal/worktree	4.668s
exit=0
```
